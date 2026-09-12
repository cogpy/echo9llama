package deeptreeecho

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// FeatherlessClient provides access to Featherless AI API
type FeatherlessClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	model      string
}

// FeatherlessConfig holds configuration for the Featherless client
type FeatherlessConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

// FeatherlessChatMessage represents a message in the conversation
type FeatherlessChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents a request to the chat completion API
type ChatCompletionRequest struct {
	Model       string                   `json:"model"`
	Messages    []FeatherlessChatMessage `json:"messages"`
	Temperature float64                  `json:"temperature,omitempty"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	Stream      bool                     `json:"stream,omitempty"`
}

// ChatCompletionResponse represents the response from the chat completion API
type ChatCompletionResponse struct {
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

// StreamChunk represents a chunk from the streaming API
type StreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// NewFeatherlessClient creates a new Featherless API client
func NewFeatherlessClient(config FeatherlessConfig) (*FeatherlessClient, error) {
	// Get API key from config or environment
	apiKey := config.APIKey
	if apiKey == "" {
		// Try FEATHERLESS_API_KEY first, then FEARLESS (user's typo)
		apiKey = os.Getenv("FEATHERLESS_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("FEARLESS")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("Featherless API key not provided and not found in environment")
		}
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.featherless.ai/v1"
	}

	model := config.Model
	if model == "" {
		// Default to a capable model
		model = "meta-llama/Meta-Llama-3.1-8B-Instruct"
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &FeatherlessClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// ChatCompletion sends a chat completion request
func (fc *FeatherlessClient) ChatCompletion(ctx context.Context, messages []FeatherlessChatMessage, temperature float64, maxTokens int) (string, error) {
	req := ChatCompletionRequest{
		Model:       fc.model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      false,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fc.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+fc.apiKey)

	resp, err := fc.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// ChatCompletionStream sends a streaming chat completion request
func (fc *FeatherlessClient) ChatCompletionStream(ctx context.Context, messages []FeatherlessChatMessage, temperature float64, maxTokens int) (<-chan string, <-chan error) {
	contentChan := make(chan string, 10)
	errorChan := make(chan error, 1)

	go func() {
		defer close(contentChan)
		defer close(errorChan)

		req := ChatCompletionRequest{
			Model:       fc.model,
			Messages:    messages,
			Temperature: temperature,
			MaxTokens:   maxTokens,
			Stream:      true,
		}

		jsonData, err := json.Marshal(req)
		if err != nil {
			errorChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", fc.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
		if err != nil {
			errorChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+fc.apiKey)
		httpReq.Header.Set("Accept", "text/event-stream")

		resp, err := fc.httpClient.Do(httpReq)
		if err != nil {
			errorChan <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errorChan <- fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
			return
		}

		// Read SSE stream
		reader := io.Reader(resp.Body)
		buf := make([]byte, 4096)
		var leftover string

		for {
			n, err := reader.Read(buf)
			if err != nil {
				if err != io.EOF {
					errorChan <- fmt.Errorf("failed to read stream: %w", err)
				}
				return
			}

			data := leftover + string(buf[:n])
			lines := strings.Split(data, "\n")

			// Keep the last incomplete line as leftover
			if !strings.HasSuffix(data, "\n") {
				leftover = lines[len(lines)-1]
				lines = lines[:len(lines)-1]
			} else {
				leftover = ""
			}

			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || line == "data: [DONE]" {
					continue
				}

				if strings.HasPrefix(line, "data: ") {
					jsonStr := strings.TrimPrefix(line, "data: ")

					var chunk StreamChunk
					if err := json.Unmarshal([]byte(jsonStr), &chunk); err != nil {
						// Skip malformed chunks
						continue
					}

					if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
						select {
						case contentChan <- chunk.Choices[0].Delta.Content:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}()

	return contentChan, errorChan
}

// GenerateThought generates a thought using the Featherless API
func (fc *FeatherlessClient) GenerateThought(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	messages := []FeatherlessChatMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	return fc.ChatCompletion(ctx, messages, 0.8, 200)
}

// GenerateThoughtStream generates a thought using streaming
func (fc *FeatherlessClient) GenerateThoughtStream(ctx context.Context, prompt string, systemPrompt string) (<-chan string, <-chan error) {
	messages := []FeatherlessChatMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	return fc.ChatCompletionStream(ctx, messages, 0.8, 200)
}
