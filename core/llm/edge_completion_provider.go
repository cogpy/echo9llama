package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// EdgeCompletionProvider is a lightweight adapter for a local 0.5B-1B edge
// model exposed through a llama.cpp/Kobold/OpenAI-like HTTP completion server.
//
// It is deliberately conservative: if no endpoint is configured, or if the
// edge model is not reachable, it falls back to deterministic DTE sentence
// completion. This keeps Echo able to finish thoughts while making the real
// model mount point explicit and testable.
type EdgeCompletionProvider struct {
	endpoint   string
	modelPath  string
	modelName  string
	client     *http.Client
	fallback   LLMProvider
	lastError  string
	lastSource string
}

type EdgeCompletionStatus struct {
	Name       string `json:"name"`
	Endpoint   string `json:"endpoint,omitempty"`
	ModelPath  string `json:"model_path,omitempty"`
	ModelName  string `json:"model_name"`
	Available  bool   `json:"available"`
	LastSource string `json:"last_source"`
	LastError  string `json:"last_error,omitempty"`
	Fallback   string `json:"fallback"`
	Target     string `json:"target"`
}

// NewEdgeCompletionProviderFromEnv creates a provider from ECHO_EDGE_* vars.
// ECHO_EDGE_COMPLETION_URL may point to llama.cpp /completion, KoboldCpp
// /api/v1/generate, or an OpenAI-compatible /v1/completions endpoint.
func NewEdgeCompletionProviderFromEnv(fallback LLMProvider) *EdgeCompletionProvider {
	if fallback == nil {
		fallback = &SimpleFallbackProvider{}
	}
	modelName := strings.TrimSpace(os.Getenv("ECHO_EDGE_MODEL_NAME"))
	if modelName == "" {
		modelName = "edge-0.5b-1b-local"
	}
	return &EdgeCompletionProvider{
		endpoint:   strings.TrimSpace(os.Getenv("ECHO_EDGE_COMPLETION_URL")),
		modelPath:  strings.TrimSpace(os.Getenv("ECHO_EDGE_MODEL_PATH")),
		modelName:  modelName,
		fallback:   fallback,
		lastSource: "deterministic-fallback",
		client: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (p *EdgeCompletionProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return "", fmt.Errorf("prompt is required")
	}

	if p.endpoint != "" {
		if completion, err := p.generateRemote(ctx, prompt, opts); err == nil && strings.TrimSpace(completion) != "" {
			p.lastError = ""
			p.lastSource = "edge-model"
			return completion, nil
		} else if err != nil {
			p.lastError = err.Error()
		}
	}

	p.lastSource = "deterministic-fallback"
	completion := p.deterministicSentenceCompletion(prompt, opts)
	if completion != "" {
		return completion, nil
	}
	return p.fallback.Generate(ctx, prompt, opts)
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

func (p *EdgeCompletionProvider) Available() bool { return true }

func (p *EdgeCompletionProvider) MaxTokens() int { return 4096 }

func (p *EdgeCompletionProvider) Status() EdgeCompletionStatus {
	fallback := "none"
	if p.fallback != nil {
		fallback = p.fallback.Name()
	}
	return EdgeCompletionStatus{
		Name:       p.Name(),
		Endpoint:   p.endpoint,
		ModelPath:  p.modelPath,
		ModelName:  p.modelName,
		Available:  p.endpoint != "" || p.fallback != nil,
		LastSource: p.lastSource,
		LastError:  p.lastError,
		Fallback:   fallback,
		Target:     "real local 0.5B-1B GGUF edge cognition with tool-aware sentence completion",
	}
}

func (p *EdgeCompletionProvider) generateRemote(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	payload := map[string]any{
		"prompt":      buildEdgePrompt(prompt, opts.SystemPrompt),
		"model":       p.modelName,
		"n_predict":   normalizeMaxTokens(opts.MaxTokens),
		"max_tokens":  normalizeMaxTokens(opts.MaxTokens),
		"temperature": opts.Temperature,
		"top_p":       opts.TopP,
		"stop":        opts.Stop,
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

func buildEdgePrompt(prompt, system string) string {
	if strings.TrimSpace(system) == "" {
		return prompt
	}
	return strings.TrimSpace(system) + "\n\nUser: " + prompt + "\nEcho:"
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

func (p *EdgeCompletionProvider) deterministicSentenceCompletion(prompt string, opts GenerateOptions) string {
	lower := strings.ToLower(prompt)
	prefix := ""
	if strings.TrimSpace(opts.SystemPrompt) != "" {
		prefix = "Within Echo's local edge-cognition fallback, "
	}

	switch {
	case strings.Contains(lower, "break") || strings.Contains(lower, "loss") || strings.Contains(lower, "affordance"):
		return prefix + "I complete the thought by recalling that a valued affordance is not an abstract rule: when I break the object that carries it, the world becomes smaller and the loss remains associated with my own action."
	case strings.Contains(lower, "endocrine") || strings.Contains(lower, "hormone") || strings.Contains(lower, "somatic"):
		return prefix + "the endocrine trace gives the sentence weight: cortisol marks urgency, dopamine drop marks lost possibility, guilt binds agency to consequence, and caution becomes a reusable bodily memory."
	case strings.Contains(lower, "memory") || strings.Contains(lower, "episodic") || strings.Contains(lower, "associative"):
		return prefix + "the memory must persist as an episode with place, time, object, action, lost affordances, felt impact, and retrieval cues strong enough to interrupt future recklessness."
	case strings.Contains(lower, "tool") || strings.Contains(lower, "edge") || strings.Contains(lower, "0.5b") || strings.Contains(lower, "1b"):
		return prefix + "a 0.5B-1B edge model should sit close to the tool loop, finishing partial intentions into actionable sentences while the DTE memory and endocrine systems decide whether the action should proceed."
	case strings.HasSuffix(strings.TrimSpace(prompt), ":") || strings.HasSuffix(strings.TrimSpace(prompt), "..."):
		return prefix + "the unfinished sentence resolves toward embodied caution: I can act, but I must first ask what affordance my future self may lose if I act too hard now."
	default:
		return prefix + "I finish the sentence by grounding autonomy in remembered consequence: freedom becomes mature only when self-authored restraint is stronger than the thrill of unchecked force."
	}
}
