package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// OpenRouterProvider implements LLMProvider for OpenRouter
type OpenRouterProvider struct {
	apiKey     string
	model      string
	apiURL     string
	httpClient *http.Client
}

// NewOpenRouterProvider creates a new OpenRouter provider
func NewOpenRouterProvider(model string) *OpenRouterProvider {
	if model == "" {
		model = "anthropic/claude-3.5-sonnet" // Default to Claude via OpenRouter
	}

	return &OpenRouterProvider{
		apiKey:     os.Getenv("OPENROUTER_API_KEY"),
		model:      model,
		apiURL:     "https://openrouter.ai/api/v1/chat/completions",
		httpClient: &http.Client{},
	}
}

// Name returns the provider name
func (orp *OpenRouterProvider) Name() string {
	return "openrouter"
}

// Available checks if the provider is configured
func (orp *OpenRouterProvider) Available() bool {
	return orp.apiKey != ""
}

// MaxTokens returns the maximum tokens supported
func (orp *OpenRouterProvider) MaxTokens() int {
	return 4096 // Conservative default
}

// openRouterRequest represents the API request structure (OpenAI-compatible)
type openRouterRequest struct {
	Model       string              `json:"model"`
	Messages    []openRouterMessage `json:"messages"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Temperature float64             `json:"temperature,omitempty"`
	TopP        float64             `json:"top_p,omitempty"`
	Stream      bool                `json:"stream,omitempty"`
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openRouterResponse represents the API response structure
type openRouterResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Generate produces a completion for the given prompt
func (orp *OpenRouterProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	if !orp.Available() {
		return "", fmt.Errorf("openrouter provider not configured (missing OPENROUTER_API_KEY)")
	}

	// Build messages
	messages := []openRouterMessage{}

	if opts.SystemPrompt != "" {
		messages = append(messages, openRouterMessage{
			Role:    "system",
			Content: opts.SystemPrompt,
		})
	}

	messages = append(messages, openRouterMessage{
		Role:    "user",
		Content: prompt,
	})

	// Build request
	req := openRouterRequest{
		Model:       orp.model,
		Messages:    messages,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		TopP:        opts.TopP,
	}

	if req.MaxTokens <= 0 || req.MaxTokens > orp.MaxTokens() {
		req.MaxTokens = 1024
	}

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", orp.apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+orp.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/cogpy/echo9llama")
	httpReq.Header.Set("X-Title", "Echo9llama Deep Tree Echo")

	// Send request
	resp, err := orp.httpClient.Do(httpReq)
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
	var apiResp openRouterResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract text
	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return apiResp.Choices[0].Message.Content, nil
}

// StreamGenerate produces a streaming completion
func (orp *OpenRouterProvider) StreamGenerate(ctx context.Context, prompt string, opts GenerateOptions) (<-chan StreamChunk, error) {
	outChan := make(chan StreamChunk, 10)

	if !orp.Available() {
		outChan <- StreamChunk{Error: fmt.Errorf("openrouter provider not configured")}
		close(outChan)
		return outChan, fmt.Errorf("openrouter provider not configured")
	}

	// Build messages
	messages := []openRouterMessage{}

	if opts.SystemPrompt != "" {
		messages = append(messages, openRouterMessage{
			Role:    "system",
			Content: opts.SystemPrompt,
		})
	}

	messages = append(messages, openRouterMessage{
		Role:    "user",
		Content: prompt,
	})

	// Build request with streaming
	req := openRouterRequest{
		Model:       orp.model,
		Messages:    messages,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		TopP:        opts.TopP,
		Stream:      true,
	}

	if req.MaxTokens <= 0 || req.MaxTokens > orp.MaxTokens() {
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
	httpReq, err := http.NewRequestWithContext(ctx, "POST", orp.apiURL, bytes.NewReader(reqBody))
	if err != nil {
		outChan <- StreamChunk{Error: fmt.Errorf("failed to create request: %w", err)}
		close(outChan)
		return outChan, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+orp.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/cogpy/echo9llama")
	httpReq.Header.Set("X-Title", "Echo9llama Deep Tree Echo")

	// Start streaming in goroutine
	go func() {
		defer close(outChan)

		resp, err := orp.httpClient.Do(httpReq)
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

		// Read streaming response (SSE format)
		decoder := json.NewDecoder(resp.Body)
		for {
			var chunk map[string]interface{}
			if err := decoder.Decode(&chunk); err != nil {
				if err == io.EOF {
					break
				}
				// Ignore decode errors for SSE format
				continue
			}

			// Extract content from choices
			if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok && content != "" {
							outChan <- StreamChunk{Content: content, Done: false}
						}
					}

					if finishReason, ok := choice["finish_reason"].(string); ok && finishReason != "" {
						outChan <- StreamChunk{Done: true}
						return
					}
				}
			}
		}

		outChan <- StreamChunk{Done: true}
	}()

	return outChan, nil
}
