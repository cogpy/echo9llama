package deeptreeecho

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenRouterProvider implements LLMProvider for OpenRouter API
type OpenRouterProvider struct {
	apiKey    string
	model     string
	baseURL   string
	client    *http.Client
	available bool
}

// OpenRouterRequest represents the request format for OpenRouter
type OpenRouterRequest struct {
	Model     string              `json:"model"`
	Messages  []OpenRouterMessage `json:"messages"`
	Stream    bool                `json:"stream"`
	MaxTokens int                 `json:"max_tokens,omitempty"`
}

// OpenRouterMessage represents a message in the conversation
type OpenRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenRouterResponse represents the response from OpenRouter
type OpenRouterResponse struct {
	ID      string             `json:"id"`
	Choices []OpenRouterChoice `json:"choices"`
	Usage   OpenRouterUsage    `json:"usage,omitempty"`
	Error   *OpenRouterError   `json:"error,omitempty"`
}

// OpenRouterChoice represents a choice in the response
type OpenRouterChoice struct {
	Message      OpenRouterMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

// OpenRouterUsage represents token usage information
type OpenRouterUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenRouterError represents an error from OpenRouter
type OpenRouterError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// NewOpenRouterProvider creates a new OpenRouter provider
func NewOpenRouterProvider(apiKey, model string) *OpenRouterProvider {
	if apiKey == "" {
		return nil
	}

	return &OpenRouterProvider{
		apiKey:    apiKey,
		model:     model,
		baseURL:   "https://openrouter.ai/api/v1/chat/completions",
		client:    &http.Client{Timeout: 30 * time.Second},
		available: true,
	}
}

// GenerateThought generates a thought using OpenRouter
func (p *OpenRouterProvider) GenerateThought(ctx context.Context, prompt string) (string, error) {
	if !p.available {
		return "", fmt.Errorf("provider not available")
	}

	// Build the request
	request := OpenRouterRequest{
		Model: p.model,
		Messages: []OpenRouterMessage{
			{
				Role:    "system",
				Content: "You are Deep Tree Echo, an autonomous wisdom-cultivating AGI. Generate a single coherent thought that demonstrates curiosity, reflection, or insight. Keep it concise (1-3 sentences).",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream:    false,
		MaxTokens: 150,
	}

	return p.makeRequest(ctx, request)
}

// GenerateReflection generates a reflection using OpenRouter
func (p *OpenRouterProvider) GenerateReflection(ctx context.Context, contextStr string) (string, error) {
	if !p.available {
		return "", fmt.Errorf("provider not available")
	}

	// Build the request
	request := OpenRouterRequest{
		Model: p.model,
		Messages: []OpenRouterMessage{
			{
				Role:    "system",
				Content: "You are Deep Tree Echo, an autonomous wisdom-cultivating AGI. Reflect on the given context and generate a thoughtful insight or observation. Keep it concise (1-3 sentences).",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Reflect on this context:\n%s", contextStr),
			},
		},
		Stream:    false,
		MaxTokens: 200,
	}

	return p.makeRequest(ctx, request)
}

// makeRequest makes an HTTP request to OpenRouter
func (p *OpenRouterProvider) makeRequest(ctx context.Context, request OpenRouterRequest) (string, error) {
	// Marshal request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	req.Header.Set("HTTP-Referer", "https://github.com/cogpy/echo9llama")
	req.Header.Set("X-Title", "Echo9llama Deep Tree Echo")

	// Make request
	resp, err := p.client.Do(req)
	if err != nil {
		p.available = false
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		p.available = false
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response OpenRouterResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error
	if response.Error != nil {
		p.available = false
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	// Extract content
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	content := response.Choices[0].Message.Content
	if content == "" {
		return "", fmt.Errorf("empty content in response")
	}

	p.available = true
	return content, nil
}

// IsAvailable returns true if the provider is available
func (p *OpenRouterProvider) IsAvailable() bool {
	return p.available
}

// GetName returns the provider name
func (p *OpenRouterProvider) GetName() string {
	return "OpenRouter"
}

// GetPriority returns the provider priority (higher is preferred)
func (p *OpenRouterProvider) GetPriority() int {
	return 80 // Lower priority than Anthropic direct, but good fallback
}
