package drivers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/EchoCog/echollama/core/llm"
)

// TextGenerator is the language-model interface the ecco9 LLM and NPU devices
// drive. *llm.ProviderManager and every llm.LLMProvider satisfy it.
type TextGenerator interface {
	Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error)
}

// ErrNoModelBackend is returned when a device is asked to produce text but no
// language model is configured. No substitute text is ever produced.
var ErrNoModelBackend = errors.New("ecco9: no language model backend configured")

// Environment variables selecting a model backend for ModelBackendFromEnv.
const (
	// EnvLLMEndpoint is the base URL of an OpenAI-compatible chat completions
	// API (for example a llama.cpp server at http://127.0.0.1:8080/v1).
	EnvLLMEndpoint = "ECCO9_LLM_ENDPOINT"
	// EnvLLMModel names the model requested from EnvLLMEndpoint.
	EnvLLMModel = "ECCO9_LLM_MODEL"
	// EnvLLMAPIKey is an optional bearer token for EnvLLMEndpoint.
	EnvLLMAPIKey = "ECCO9_LLM_API_KEY"
)

// ModelBackendFromEnv builds a real model backend from the environment:
// ECCO9_LLM_ENDPOINT (OpenAI-compatible / llama.cpp server) takes precedence;
// otherwise any of OPENAI_API_KEY, ANTHROPIC_API_KEY or OPENROUTER_API_KEY
// registers the corresponding core/llm provider. It returns ErrNoModelBackend
// when none is configured.
func ModelBackendFromEnv() (TextGenerator, error) {
	if endpoint := strings.TrimSpace(os.Getenv(EnvLLMEndpoint)); endpoint != "" {
		return NewOpenAICompatibleGenerator(endpoint, os.Getenv(EnvLLMModel), os.Getenv(EnvLLMAPIKey)), nil
	}

	manager := llm.NewProviderManager()
	var providers []llm.LLMProvider
	if key := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); key != "" {
		providers = append(providers, llm.NewOpenAIProvider(key))
	}
	if strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")) != "" {
		providers = append(providers, llm.NewAnthropicProvider(""))
	}
	if strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY")) != "" {
		providers = append(providers, llm.NewOpenRouterProvider(""))
	}
	if len(providers) == 0 {
		return nil, ErrNoModelBackend
	}
	chain := make([]string, 0, len(providers))
	for _, p := range providers {
		if err := manager.RegisterProvider(p); err != nil {
			return nil, err
		}
		chain = append(chain, p.Name())
	}
	if err := manager.SetFallbackChain(chain); err != nil {
		return nil, err
	}
	return manager, nil
}

// OpenAICompatibleGenerator calls an OpenAI-compatible /chat/completions
// endpoint, such as a llama.cpp or vLLM server.
type OpenAICompatibleGenerator struct {
	BaseURL    string
	Model      string
	APIKey     string
	HTTPClient *http.Client
}

// NewOpenAICompatibleGenerator creates a generator for baseURL (e.g.
// "http://127.0.0.1:8080/v1").
func NewOpenAICompatibleGenerator(baseURL, model, apiKey string) *OpenAICompatibleGenerator {
	return &OpenAICompatibleGenerator{
		BaseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		Model:      strings.TrimSpace(model),
		APIKey:     strings.TrimSpace(apiKey),
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model,omitempty"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	Stop        []string      `json:"stop,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Generate implements TextGenerator by querying the configured endpoint.
func (g *OpenAICompatibleGenerator) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	if g == nil || g.BaseURL == "" {
		return "", ErrNoModelBackend
	}
	messages := make([]chatMessage, 0, 2)
	if strings.TrimSpace(opts.SystemPrompt) != "" {
		messages = append(messages, chatMessage{Role: "system", Content: opts.SystemPrompt})
	}
	messages = append(messages, chatMessage{Role: "user", Content: prompt})

	body, err := json.Marshal(chatRequest{
		Model:       g.Model,
		Messages:    messages,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		TopP:        opts.TopP,
		Stop:        opts.Stop,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if g.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+g.APIKey)
	}
	client := g.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ecco9: model request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return "", fmt.Errorf("ecco9: reading model response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ecco9: model endpoint returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("ecco9: decoding model response: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", fmt.Errorf("ecco9: model error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("ecco9: model returned no choices")
	}
	return parsed.Choices[0].Message.Content, nil
}

// generateText runs prompt through gen and rejects empty output, so callers
// never surface anything that did not come from the model.
func generateText(ctx context.Context, gen TextGenerator, prompt string, opts llm.GenerateOptions) (string, error) {
	if gen == nil {
		return "", ErrNoModelBackend
	}
	out, err := gen.Generate(ctx, prompt, opts)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(out) == "" {
		return "", errors.New("ecco9: model returned an empty completion")
	}
	return out, nil
}
