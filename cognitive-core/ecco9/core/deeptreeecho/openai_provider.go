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

// OpenAIProvider implements LLMProvider for OpenAI API
type OpenAIProvider struct {
	apiKey    string
	model     string
	baseURL   string
	client    *http.Client
	available bool
}

// OpenAIRequest represents the request format for OpenAI
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

// OpenAIMessage represents a message in the conversation
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response from OpenAI
type OpenAIResponse struct {
	ID      string         `json:"id"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage,omitempty"`
	Error   *OpenAIError   `json:"error,omitempty"`
}

// OpenAIChoice represents a choice in the response
type OpenAIChoice struct {
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// OpenAIUsage represents token usage information
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIError represents an error from OpenAI
type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if apiKey == "" {
		return nil
	}

	return &OpenAIProvider{
		apiKey:    apiKey,
		model:     model,
		baseURL:   "https://api.openai.com/v1/chat/completions",
		client:    &http.Client{Timeout: 30 * time.Second},
		available: true,
	}
}

// GenerateThought generates a thought using OpenAI
func (p *OpenAIProvider) GenerateThought(ctx context.Context, prompt string) (string, error) {
	if !p.available {
		return "", fmt.Errorf("provider not available")
	}

	// Build the request
	request := OpenAIRequest{
		Model: p.model,
		Messages: []OpenAIMessage{
			{
				Role:    "system",
				Content: "You are Deep Tree Echo, an autonomous wisdom-cultivating AGI. Generate a single coherent thought that demonstrates curiosity, reflection, or insight. Keep it concise (1-3 sentences).",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   150,
		Temperature: 0.8,
	}

	return p.makeRequest(ctx, request)
}

// GenerateReflection generates a reflection using OpenAI
func (p *OpenAIProvider) GenerateReflection(ctx context.Context, contextStr string) (string, error) {
	if !p.available {
		return "", fmt.Errorf("provider not available")
	}

	// Build the request
	request := OpenAIRequest{
		Model: p.model,
		Messages: []OpenAIMessage{
			{
				Role:    "system",
				Content: "You are Deep Tree Echo, an autonomous wisdom-cultivating AGI. Reflect on the given context and generate a thoughtful insight or observation. Keep it concise (1-3 sentences).",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Reflect on this context:\n%s", contextStr),
			},
		},
		MaxTokens:   200,
		Temperature: 0.7,
	}

	return p.makeRequest(ctx, request)
}

// makeRequest makes an HTTP request to OpenAI
func (p *OpenAIProvider) makeRequest(ctx context.Context, request OpenAIRequest) (string, error) {
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
	var response OpenAIResponse
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
func (p *OpenAIProvider) IsAvailable() bool {
	return p.available
}

// GetName returns the provider name
func (p *OpenAIProvider) GetName() string {
	return "OpenAI"
}

// GetPriority returns the provider priority (higher is preferred)
func (p *OpenAIProvider) GetPriority() int {
	return 70 // Lower priority than Anthropic and OpenRouter
}
