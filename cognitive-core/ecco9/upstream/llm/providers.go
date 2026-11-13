package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Provider interface for LLM providers
type Provider interface {
	Generate(ctx context.Context, prompt string, options GenerateOptions) (string, error)
	GetName() string
}

// GenerateOptions configures generation parameters
type GenerateOptions struct {
	Temperature  float64
	MaxTokens    int
	SystemPrompt string
}

// AnthropicProvider implements Provider for Anthropic Claude API
type AnthropicProvider struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

// OpenRouterProvider implements Provider for OpenRouter API
type OpenRouterProvider struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey string) (*AnthropicProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}

	return &AnthropicProvider{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		model: "claude-3-5-sonnet-20241022",
	}, nil
}

// NewOpenRouterProvider creates a new OpenRouter provider
func NewOpenRouterProvider(apiKey string) (*OpenRouterProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenRouter API key is required")
	}

	return &OpenRouterProvider{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		model: "anthropic/claude-3.5-sonnet",
	}, nil
}

// Generate implements Provider for AnthropicProvider
func (ap *AnthropicProvider) Generate(ctx context.Context, prompt string, options GenerateOptions) (string, error) {
	// Prepare request
	requestBody := map[string]interface{}{
		"model":       ap.model,
		"max_tokens":  options.MaxTokens,
		"temperature": options.Temperature,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	if options.SystemPrompt != "" {
		requestBody["system"] = options.SystemPrompt
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", ap.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	resp, err := ap.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return response.Content[0].Text, nil
}

// GetName implements Provider for AnthropicProvider
func (ap *AnthropicProvider) GetName() string {
	return "Anthropic Claude"
}

// Generate implements Provider for OpenRouterProvider
func (orp *OpenRouterProvider) Generate(ctx context.Context, prompt string, options GenerateOptions) (string, error) {
	// Prepare request
	messages := []map[string]string{
		{
			"role":    "user",
			"content": prompt,
		},
	}

	if options.SystemPrompt != "" {
		messages = append([]map[string]string{
			{
				"role":    "system",
				"content": options.SystemPrompt,
			},
		}, messages...)
	}

	requestBody := map[string]interface{}{
		"model":       orp.model,
		"max_tokens":  options.MaxTokens,
		"temperature": options.Temperature,
		"messages":    messages,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+orp.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/cogpy/echo9llama")
	req.Header.Set("X-Title", "Echo9llama Autonomous Agent")

	// Send request
	resp, err := orp.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

// GetName implements Provider for OpenRouterProvider
func (orp *OpenRouterProvider) GetName() string {
	return "OpenRouter"
}

// DefaultGenerateOptions returns sensible defaults
func DefaultGenerateOptions() GenerateOptions {
	return GenerateOptions{
		Temperature:  0.7,
		MaxTokens:    1024,
		SystemPrompt: "",
	}
}
