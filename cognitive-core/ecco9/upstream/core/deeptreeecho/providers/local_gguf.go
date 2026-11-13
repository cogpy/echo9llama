// Package providers contains the real model providers for Deep Tree Echo.
//
// LocalGGUFProvider serves a GGUF model through a locally running
// OpenAI-compatible server — Ollama (http://localhost:11434/v1) or
// llama.cpp's `llama-server` (http://localhost:8080/v1). It never fabricates
// text: if the server is not reachable the provider is unavailable and every
// call returns an error, so the ModelManager moves on to another real model
// or reports that none is available.
package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
)

// LocalGGUFProvider talks to a local OpenAI-compatible inference server.
type LocalGGUFProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewLocalGGUFProvider reads LOCAL_LLM_BASE_URL (default Ollama) and
// LOCAL_LLM_MODEL (default llama3.2).
func NewLocalGGUFProvider() *LocalGGUFProvider {
	base := strings.TrimRight(os.Getenv("LOCAL_LLM_BASE_URL"), "/")
	if base == "" {
		base = "http://localhost:11434/v1"
	}
	model := os.Getenv("LOCAL_LLM_MODEL")
	if model == "" {
		model = "llama3.2"
	}
	return &LocalGGUFProvider{
		baseURL: base,
		model:   model,
		client:  &http.Client{Timeout: requestTimeout()},
	}
}

// LoadModel selects which served model subsequent calls use. The server must
// already have it; nothing is downloaded or faked here.
func (p *LocalGGUFProvider) LoadModel(modelName string) error {
	models := p.ListAvailableModels()
	for _, m := range models {
		if m == modelName {
			p.model = modelName
			return nil
		}
	}
	return fmt.Errorf("model %q is not served by %s (available: %v)", modelName, p.baseURL, models)
}

// GetLoadedModel returns the model name used for generation.
func (p *LocalGGUFProvider) GetLoadedModel() string { return p.model }

// GetModelInfo returns server/model metadata.
func (p *LocalGGUFProvider) GetModelInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":     p.model,
		"base_url": p.baseURL,
		"served":   p.IsAvailable(),
	}
}

// ListAvailableModels asks the server what it serves. Empty if unreachable.
func (p *LocalGGUFProvider) ListAvailableModels() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/models", nil)
	if err != nil {
		return nil
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	var body struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil
	}
	out := make([]string, 0, len(body.Data))
	for _, m := range body.Data {
		out = append(out, m.ID)
	}
	return out
}

// IsAvailable is true only when the local server answers.
func (p *LocalGGUFProvider) IsAvailable() bool {
	return len(p.ListAvailableModels()) > 0
}

// defaultMaxTokens bounds a reply when the caller did not specify a length.
const defaultMaxTokens = 512

// requestTimeout reads LOCAL_LLM_TIMEOUT_SECONDS (default 180) so slow CPU-only
// hosts can be given more time instead of their replies being dropped.
func requestTimeout() time.Duration {
	if raw := os.Getenv("LOCAL_LLM_TIMEOUT_SECONDS"); raw != "" {
		if secs, err := strconv.Atoi(raw); err == nil && secs > 0 {
			return time.Duration(secs) * time.Second
		}
	}
	return 180 * time.Second
}

type chatRequest struct {
	Model       string                     `json:"model"`
	Messages    []deeptreeecho.ChatMessage `json:"messages"`
	Temperature float64                    `json:"temperature"`
	MaxTokens   int                        `json:"max_tokens,omitempty"`
	TopP        float64                    `json:"top_p,omitempty"`
	Stop        []string                   `json:"stop,omitempty"`
	Stream      bool                       `json:"stream"`
}

func (p *LocalGGUFProvider) buildRequest(messages []deeptreeecho.ChatMessage, o deeptreeecho.GenerateOptions, stream bool) chatRequest {
	model := p.model
	if o.Model != "" {
		model = o.Model
	}
	maxTokens := o.MaxTokens
	if maxTokens <= 0 {
		// A locally served model on CPU must never be asked for an unbounded
		// reply: without a cap a rambling completion outlives the request.
		maxTokens = defaultMaxTokens
	}
	return chatRequest{
		Model:       model,
		Messages:    messages,
		Temperature: o.Temperature,
		MaxTokens:   maxTokens,
		TopP:        o.TopP,
		Stop:        o.StopSequences,
		Stream:      stream,
	}
}

func (p *LocalGGUFProvider) post(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+path, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("local model server %s unreachable: %w", p.baseURL, err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("local model server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return resp, nil
}

// Chat implements ModelProvider.Chat with a real completion.
func (p *LocalGGUFProvider) Chat(ctx context.Context, messages []deeptreeecho.ChatMessage, options deeptreeecho.ChatOptions) (string, error) {
	if options.SystemPrompt != "" {
		messages = append([]deeptreeecho.ChatMessage{{Role: "system", Content: options.SystemPrompt}}, messages...)
	}
	resp, err := p.post(ctx, "/chat/completions", p.buildRequest(messages, options.GenerateOptions, false))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode local model response: %w", err)
	}
	if len(out.Choices) == 0 || strings.TrimSpace(out.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("local model returned no content")
	}
	return out.Choices[0].Message.Content, nil
}

// Generate implements ModelProvider.Generate as a single-turn chat.
func (p *LocalGGUFProvider) Generate(ctx context.Context, prompt string, options deeptreeecho.GenerateOptions) (string, error) {
	return p.Chat(ctx, []deeptreeecho.ChatMessage{{Role: "user", Content: prompt}}, deeptreeecho.ChatOptions{GenerateOptions: options})
}

// ChatStream streams real tokens from the server (SSE). On error the channel
// is closed without emitting text; the error is returned up front when the
// request itself fails.
func (p *LocalGGUFProvider) ChatStream(ctx context.Context, messages []deeptreeecho.ChatMessage, options deeptreeecho.ChatOptions) (<-chan string, error) {
	if options.SystemPrompt != "" {
		messages = append([]deeptreeecho.ChatMessage{{Role: "system", Content: options.SystemPrompt}}, messages...)
	}
	resp, err := p.post(ctx, "/chat/completions", p.buildRequest(messages, options.GenerateOptions, true))
	if err != nil {
		return nil, err
	}
	ch := make(chan string, 64)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				return
			}
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if json.Unmarshal([]byte(data), &chunk) != nil || len(chunk.Choices) == 0 {
				continue
			}
			if c := chunk.Choices[0].Delta.Content; c != "" {
				select {
				case ch <- c:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return ch, nil
}

// GenerateStream implements ModelProvider.GenerateStream.
func (p *LocalGGUFProvider) GenerateStream(ctx context.Context, prompt string, options deeptreeecho.GenerateOptions) (<-chan string, error) {
	return p.ChatStream(ctx, []deeptreeecho.ChatMessage{{Role: "user", Content: prompt}}, deeptreeecho.ChatOptions{GenerateOptions: options})
}

// Embeddings calls the server's embeddings endpoint; errors if unsupported.
func (p *LocalGGUFProvider) Embeddings(ctx context.Context, text string) ([]float64, error) {
	resp, err := p.post(ctx, "/embeddings", map[string]interface{}{"model": p.model, "input": text})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode embeddings: %w", err)
	}
	if len(out.Data) == 0 {
		return nil, fmt.Errorf("local model server returned no embedding")
	}
	return out.Data[0].Embedding, nil
}

// GetInfo implements ModelProvider.GetInfo.
func (p *LocalGGUFProvider) GetInfo() deeptreeecho.ProviderInfo {
	return deeptreeecho.ProviderInfo{
		Name:         "local_gguf",
		Description:  "Local GGUF model via OpenAI-compatible server (" + p.baseURL + ")",
		Models:       p.ListAvailableModels(),
		Capabilities: []string{"chat", "generate", "stream", "embeddings"},
	}
}
