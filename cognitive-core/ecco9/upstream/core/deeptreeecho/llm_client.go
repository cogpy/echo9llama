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

// LLMClient provides unified interface for multiple LLM providers
type LLMClient struct {
	provider   string // "openai", "anthropic", "openrouter"
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
	maxRetries int
	timeout    time.Duration
}

// LLMRequest represents a unified request structure
type LLMRequest struct {
	SystemPrompt string
	UserPrompt   string
	Temperature  float64
	MaxTokens    int
	Context      []Message
}

// Message represents a conversation message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMResponse represents a unified response structure
type LLMResponse struct {
	Content      string
	FinishReason string
	TokensUsed   int
	Model        string
}

// NewLLMClient creates a new LLM client with the specified provider
func NewLLMClient(provider, apiKey, baseURL, model string) *LLMClient {
	return &LLMClient{
		provider: provider,
		apiKey:   apiKey,
		baseURL:  baseURL,
		model:    model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries: 3,
		timeout:    30 * time.Second,
	}
}

// Generate makes a completion request to the LLM provider
func (c *LLMClient) Generate(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	switch c.provider {
	case "openai", "openrouter":
		return c.generateOpenAI(ctx, req)
	case "anthropic":
		return c.generateAnthropic(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", c.provider)
	}
}

// generateOpenAI handles OpenAI-compatible API calls
func (c *LLMClient) generateOpenAI(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	// Build messages array
	messages := []Message{
		{Role: "system", Content: req.SystemPrompt},
	}

	// Add context messages if provided
	messages = append(messages, req.Context...)

	// Add user prompt
	messages = append(messages, Message{
		Role:    "user",
		Content: req.UserPrompt,
	})

	// Build request body
	requestBody := map[string]interface{}{
		"model":       c.model,
		"messages":    messages,
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
	}

	// Make API call with retries
	var lastErr error
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		response, err := c.makeOpenAIRequest(ctx, requestBody)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Don't retry on certain errors
		if isNonRetryableError(err) {
			break
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries, lastErr)
}

// makeOpenAIRequest makes the actual HTTP request for OpenAI format
func (c *LLMClient) makeOpenAIRequest(ctx context.Context, requestBody map[string]interface{}) (*LLMResponse, error) {
	// Marshal request
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// For OpenRouter, add additional headers
	if c.provider == "openrouter" {
		httpReq.Header.Set("HTTP-Referer", "https://github.com/cogpy/echo9llama")
		httpReq.Header.Set("X-Title", "Deep Tree Echo")
	}

	// Make request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &LLMResponse{
		Content:      apiResp.Choices[0].Message.Content,
		FinishReason: apiResp.Choices[0].FinishReason,
		TokensUsed:   apiResp.Usage.TotalTokens,
		Model:        apiResp.Model,
	}, nil
}

// generateAnthropic handles Anthropic Messages API calls
func (c *LLMClient) generateAnthropic(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	// Build messages array (Anthropic doesn't use system in messages)
	messages := []Message{}

	// Add context messages if provided
	messages = append(messages, req.Context...)

	// Add user prompt
	messages = append(messages, Message{
		Role:    "user",
		Content: req.UserPrompt,
	})

	// Build request body (Anthropic format)
	requestBody := map[string]interface{}{
		"model":       c.model,
		"messages":    messages,
		"system":      req.SystemPrompt, // System prompt is separate in Anthropic
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
	}

	// Make API call with retries
	var lastErr error
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		response, err := c.makeAnthropicRequest(ctx, requestBody)
		if err == nil {
			return response, nil
		}

		lastErr = err

		if isNonRetryableError(err) {
			break
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries, lastErr)
}

// makeAnthropicRequest makes the actual HTTP request for Anthropic format
func (c *LLMClient) makeAnthropicRequest(ctx context.Context, requestBody map[string]interface{}) (*LLMResponse, error) {
	// Marshal request
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.baseURL + "/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers (Anthropic-specific)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Make request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response (Anthropic format)
	var apiResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// Extract text content
	var content string
	for _, c := range apiResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &LLMResponse{
		Content:      content,
		FinishReason: apiResp.StopReason,
		TokensUsed:   apiResp.Usage.InputTokens + apiResp.Usage.OutputTokens,
		Model:        apiResp.Model,
	}, nil
}

// isNonRetryableError determines if an error should not be retried
func isNonRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Don't retry on authentication errors
	if containsString(errStr, "401") || containsString(errStr, "403") {
		return true
	}

	// Don't retry on invalid request errors
	if containsString(errStr, "400") || containsString(errStr, "422") {
		return true
	}

	return false
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			bytes.Contains([]byte(s), []byte(substr))))
}
