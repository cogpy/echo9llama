package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// EdgeCompletionProvider is a lightweight adapter for a local 0.5B-1B edge
// model. Every completion comes from a real model: either a llama.cpp/Kobold/
// OpenAI-like HTTP completion server (ECHO_EDGE_COMPLETION_URL) or a GGUF file
// loaded in-process through the native llama.cpp bindings
// (ECHO_EDGE_MODEL_PATH). With neither mounted, Generate returns ErrNoEdgeModel
// instead of answering in Echo's voice without a model.
type EdgeCompletionProvider struct {
	endpoint   string
	modelPath  string
	modelName  string
	client     *http.Client
	local      LLMProvider
	lastError  string
	lastSource string

	// gate is shared by every priority view of this provider: it allows one
	// in-flight completion, so a busy single-slot edge model never queues
	// background work in front of a foreground request.
	gate chan struct{}
	// background views skip the model when the gate is taken and return
	// ErrEdgeBusy for that call instead of queueing; the caller simply has no
	// thought on that tick.
	background bool
}

// ErrNoEdgeModel reports that no model is mounted: neither an edge completion
// endpoint nor a local GGUF file is configured.
var ErrNoEdgeModel = errors.New("no edge model mounted: set ECHO_EDGE_COMPLETION_URL or ECHO_EDGE_MODEL_PATH")

// ErrEdgeBusy reports that a background call found the edge model in use.
var ErrEdgeBusy = errors.New("edge model busy with another request")

// Background returns a view of the provider that shares its model and gate
// but never waits for it: when another completion is in flight it returns
// ErrEdgeBusy. Use it for autonomous loops so interactive requests (chat) are
// not stuck behind them.
func (p *EdgeCompletionProvider) Background() *EdgeCompletionProvider {
	bg := *p
	bg.background = true
	return &bg
}

type EdgeCompletionStatus struct {
	Name       string `json:"name"`
	Endpoint   string `json:"endpoint,omitempty"`
	ModelPath  string `json:"model_path,omitempty"`
	ModelName  string `json:"model_name"`
	Available  bool   `json:"available"`
	LastSource string `json:"last_source"`
	LastError  string `json:"last_error,omitempty"`
	Target     string `json:"target"`
}

// NewEdgeCompletionProviderFromEnv creates a provider from ECHO_EDGE_* vars.
// ECHO_EDGE_COMPLETION_URL may point to llama.cpp /completion, KoboldCpp
// /api/v1/generate, or an OpenAI-compatible /v1/completions endpoint. Without
// a URL, ECHO_EDGE_MODEL_PATH names a GGUF file served in-process.
func NewEdgeCompletionProviderFromEnv() *EdgeCompletionProvider {
	modelName := strings.TrimSpace(os.Getenv("ECHO_EDGE_MODEL_NAME"))
	if modelName == "" {
		modelName = "edge-0.5b-1b-local"
	}
	p := &EdgeCompletionProvider{
		endpoint:   strings.TrimSpace(os.Getenv("ECHO_EDGE_COMPLETION_URL")),
		modelPath:  strings.TrimSpace(os.Getenv("ECHO_EDGE_MODEL_PATH")),
		modelName:  modelName,
		lastSource: "none",
		gate:       make(chan struct{}, 1),
		client: &http.Client{
			Timeout: edgeTimeoutFromEnv(),
		},
	}
	if p.endpoint == "" && p.modelPath != "" {
		p.local = NewLocalGGUFProvider(p.modelPath)
	}
	return p
}

// defaultEdgeTimeout bounds one edge completion. A small CPU model with a
// single slot can need minutes for a long answer; callers that want less
// pass a context deadline.
const defaultEdgeTimeout = 5 * time.Minute

// edgeTimeoutFromEnv reads ECHO_EDGE_TIMEOUT (a Go duration, e.g. "90s").
func edgeTimeoutFromEnv() time.Duration {
	if d, err := time.ParseDuration(strings.TrimSpace(os.Getenv("ECHO_EDGE_TIMEOUT"))); err == nil && d > 0 {
		return d
	}
	return defaultEdgeTimeout
}

func (p *EdgeCompletionProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return "", fmt.Errorf("prompt is required")
	}

	var (
		completion string
		err        error
	)
	switch {
	case p.endpoint != "":
		completion, err = p.withGate(ctx, func() (string, error) { return p.generateRemote(ctx, prompt, opts) })
		p.lastSource = "edge-model"
	case p.local != nil:
		completion, err = p.withGate(ctx, func() (string, error) { return p.generateLocal(ctx, prompt, opts) })
		p.lastSource = "local-gguf"
	default:
		err = ErrNoEdgeModel
	}
	if err == nil && strings.TrimSpace(completion) == "" {
		err = errors.New("edge model returned an empty completion")
	}
	if err != nil {
		p.lastError = err.Error()
		return "", err
	}
	p.lastError = ""
	return completion, nil
}

func (p *EdgeCompletionProvider) StreamGenerate(ctx context.Context, prompt string, opts GenerateOptions) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 1)
	go func() {
		defer close(ch)
		response, err := p.Generate(ctx, prompt, opts)
		if err != nil {
			ch <- StreamChunk{Error: err, Done: true}
			return
		}
		ch <- StreamChunk{Content: response, Done: true}
	}()
	return ch, nil
}

func (p *EdgeCompletionProvider) Name() string { return "EdgeCompletion" }

// Available reports whether a model is mounted.
func (p *EdgeCompletionProvider) Available() bool {
	if p.endpoint != "" {
		return true
	}
	return p.local != nil && p.local.Available()
}

func (p *EdgeCompletionProvider) MaxTokens() int { return 4096 }

func (p *EdgeCompletionProvider) Status() EdgeCompletionStatus {
	return EdgeCompletionStatus{
		Name:       p.Name(),
		Endpoint:   p.endpoint,
		ModelPath:  p.modelPath,
		ModelName:  p.modelName,
		Available:  p.Available(),
		LastSource: p.lastSource,
		LastError:  p.lastError,
		Target:     "real local 0.5B-1B GGUF edge cognition with tool-aware sentence completion",
	}
}

// withGate runs one completion while holding the shared slot. Foreground
// calls wait for it; background calls return ErrEdgeBusy when it is taken.
func (p *EdgeCompletionProvider) withGate(ctx context.Context, run func() (string, error)) (string, error) {
	if p.gate != nil {
		if p.background {
			select {
			case p.gate <- struct{}{}:
			default:
				return "", ErrEdgeBusy
			}
		} else {
			select {
			case p.gate <- struct{}{}:
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
		defer func() { <-p.gate }()
	}
	return run()
}

func (p *EdgeCompletionProvider) generateLocal(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	local := opts
	local.SystemPrompt = ""
	local.MaxTokens = normalizeMaxTokens(opts.MaxTokens)
	local.Stop = edgeStops(opts.Stop)
	return p.local.Generate(ctx, buildEdgePrompt(prompt, opts.SystemPrompt), local)
}

func (p *EdgeCompletionProvider) generateRemote(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	payload := map[string]any{
		"prompt":      buildEdgePrompt(prompt, opts.SystemPrompt),
		"model":       p.modelName,
		"n_predict":   normalizeMaxTokens(opts.MaxTokens),
		"max_tokens":  normalizeMaxTokens(opts.MaxTokens),
		"temperature": opts.Temperature,
		"top_p":       opts.TopP,
		"stop":        edgeStops(opts.Stop),
		"stream":      false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("edge completion endpoint returned %s", resp.Status)
	}

	var decoded map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", err
	}
	return extractCompletion(decoded), nil
}

// defaultEdgeStops end a completion at the next conversational turn marker, so
// small models do not invent further "User:" turns after answering.
// The bare markers catch small models that start the next turn mid-line
// ("...time? Echo: A great day...").
var defaultEdgeStops = []string{"\nUser:", "\nuser:", "\nEcho:", "User:", "Echo:"}

// EdgeTurnStops returns a copy of the default turn-marker stops, for callers
// that add their own stops and still want completions to end at a new turn.
func EdgeTurnStops() []string {
	return append([]string(nil), defaultEdgeStops...)
}

func edgeStops(stop []string) []string {
	if len(stop) > 0 {
		return stop
	}
	return defaultEdgeStops
}

func buildEdgePrompt(prompt, system string) string {
	system = strings.TrimSpace(system)
	if isTranscript(prompt) {
		// Already a "User:"/"Echo:" transcript (chat): don't wrap it in another
		// User turn; just cue Echo's reply.
		if system == "" {
			return prompt + "\nEcho:"
		}
		return system + "\n\n" + prompt + "\nEcho:"
	}
	if system == "" {
		return prompt
	}
	return system + "\n\nUser: " + prompt + "\nEcho:"
}

// isTranscript reports whether prompt is already a turn-labelled transcript.
func isTranscript(prompt string) bool {
	return strings.HasPrefix(prompt, "User: ") || strings.HasPrefix(prompt, "Echo: ")
}

func normalizeMaxTokens(maxTokens int) int {
	if maxTokens <= 0 {
		return 256
	}
	if maxTokens > 1024 {
		return 1024
	}
	return maxTokens
}

func extractCompletion(decoded map[string]any) string {
	for _, key := range []string{"content", "response", "completion", "text"} {
		if value, ok := decoded[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	if choices, ok := decoded["choices"].([]any); ok && len(choices) > 0 {
		if first, ok := choices[0].(map[string]any); ok {
			if text, ok := first["text"].(string); ok && strings.TrimSpace(text) != "" {
				return strings.TrimSpace(text)
			}
			if message, ok := first["message"].(map[string]any); ok {
				if content, ok := message["content"].(string); ok && strings.TrimSpace(content) != "" {
					return strings.TrimSpace(content)
				}
			}
		}
	}
	return ""
}
