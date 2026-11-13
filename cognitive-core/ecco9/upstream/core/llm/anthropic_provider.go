package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// AnthropicProvider implements LLMProvider for Anthropic Claude
type AnthropicProvider struct {
	apiKey     string
	model      string
	apiURL     string
	httpClient *http.Client
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(model string) *AnthropicProvider {
	if model == "" {
		model = "claude-3-5-sonnet-20241022" // Latest Claude model
	}

	return &AnthropicProvider{
		apiKey:     os.Getenv("ANTHROPIC_API_KEY"),
		model:      model,
		apiURL:     "https://api.anthropic.com/v1/messages",
		httpClient: &http.Client{},
	}
}

// Name returns the provider name
func (ap *AnthropicProvider) Name() string {
	return "anthropic"
}

// Available checks if the provider is configured
func (ap *AnthropicProvider) Available() bool {
	return ap.apiKey != ""
}

// MaxTokens returns the maximum tokens supported
func (ap *AnthropicProvider) MaxTokens() int {
	return 8192 // Claude 3.5 Sonnet supports up to 8k output
}

// anthropicRequest represents the API request structure
type anthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	Temperature float64            `json:"temperature,omitempty"`
	TopP        float64            `json:"top_p,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse represents the API response structure
type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Generate produces a completion for the given prompt
func (ap *AnthropicProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	if !ap.Available() {
		return "", fmt.Errorf("anthropic provider not configured (missing ANTHROPIC_API_KEY)")
	}

	// Build request
	req := anthropicRequest{
		Model:       ap.model,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		TopP:        opts.TopP,
		Messages: []anthropicMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	if opts.SystemPrompt != "" {
		req.System = opts.SystemPrompt
	}

	// Ensure max tokens is within limits
	if req.MaxTokens <= 0 || req.MaxTokens > ap.MaxTokens() {
		req.MaxTokens = 1024
	}

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", ap.apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", ap.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	resp, err := ap.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp anthropicResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract text
	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return apiResp.Content[0].Text, nil
}

// StreamGenerate produces a streaming completion
func (ap *AnthropicProvider) StreamGenerate(ctx context.Context, prompt string, opts GenerateOptions) (<-chan StreamChunk, error) {
	outChan := make(chan StreamChunk, 10)

	if !ap.Available() {
		outChan <- StreamChunk{Error: fmt.Errorf("anthropic provider not configured")}
		close(outChan)
		return outChan, fmt.Errorf("anthropic provider not configured")
	}

	// Build request with streaming enabled
	req := anthropicRequest{
		Model:       ap.model,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		TopP:        opts.TopP,
		Stream:      true,
		Messages: []anthropicMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	if opts.SystemPrompt != "" {
		req.System = opts.SystemPrompt
	}

	if req.MaxTokens <= 0 || req.MaxTokens > ap.MaxTokens() {
		req.MaxTokens = 1024
	}

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		outChan <- StreamChunk{Error: fmt.Errorf("failed to marshal request: %w", err)}
		close(outChan)
		return outChan, err
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", ap.apiURL, bytes.NewReader(reqBody))
	if err != nil {
		outChan <- StreamChunk{Error: fmt.Errorf("failed to create request: %w", err)}
		close(outChan)
		return outChan, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", ap.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Start streaming in goroutine
	go func() {
		defer close(outChan)

		resp, err := ap.httpClient.Do(httpReq)
		if err != nil {
			outChan <- StreamChunk{Error: fmt.Errorf("failed to send request: %w", err)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			outChan <- StreamChunk{Error: fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))}
			return
		}

		// Read streaming response
		decoder := json.NewDecoder(resp.Body)
		for {
			var event map[string]interface{}
			if err := decoder.Decode(&event); err != nil {
				if err == io.EOF {
					break
				}
				outChan <- StreamChunk{Error: fmt.Errorf("failed to decode stream: %w", err)}
				return
			}

			// Handle different event types
			eventType, ok := event["type"].(string)
			if !ok {
				continue
			}

			switch eventType {
			case "content_block_delta":
				if delta, ok := event["delta"].(map[string]interface{}); ok {
					if text, ok := delta["text"].(string); ok {
						outChan <- StreamChunk{Content: text, Done: false}
					}
				}
			case "message_stop":
				outChan <- StreamChunk{Done: true}
				return
			case "error":
				if errMsg, ok := event["error"].(map[string]interface{}); ok {
					if msg, ok := errMsg["message"].(string); ok {
						outChan <- StreamChunk{Error: fmt.Errorf("API error: %s", msg)}
					}
				}
				return
			}
		}

		outChan <- StreamChunk{Done: true}
	}()

	return outChan, nil
}

// GenerateWithMessages allows multi-turn conversations
func (ap *AnthropicProvider) GenerateWithMessages(ctx context.Context, messages []anthropicMessage, opts GenerateOptions) (string, error) {
	if !ap.Available() {
		return "", fmt.Errorf("anthropic provider not configured")
	}

	req := anthropicRequest{
		Model:       ap.model,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		TopP:        opts.TopP,
		Messages:    messages,
	}

	if opts.SystemPrompt != "" {
		req.System = opts.SystemPrompt
	}

	if req.MaxTokens <= 0 || req.MaxTokens > ap.MaxTokens() {
		req.MaxTokens = 1024
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", ap.apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", ap.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := ap.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp anthropicResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	// Concatenate all text blocks
	var result strings.Builder
	for _, content := range apiResp.Content {
		if content.Type == "text" {
			result.WriteString(content.Text)
		}
	}

	return result.String(), nil
}
