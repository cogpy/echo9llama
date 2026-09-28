// Package jan provides an LLM provider implementation for janecho-server.
// This client communicates with a janecho-server instance via its OpenAI-compatible API.
package jan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// JanClient communicates with a janecho-server instance.
// It implements the llm.LLMProvider interface.
type JanClient struct {
	config Config
	client *http.Client
}

// Config holds the configuration for the JanClient.
type Config struct {
	// Name is the identifier for this provider instance
	Name string `json:"name"`

	// URL is the base URL of the janecho-server (e.g., "http://localhost:8080")
	URL string `json:"url"`

	// APIKey is the authentication key (optional, depends on janecho-server config)
	APIKey string `json:"api_key,omitempty"`

	// Timeout is the HTTP request timeout
	Timeout time.Duration `json:"timeout"`

	// Model is the default model to use (janecho-server handles routing)
	Model string `json:"model"`
}

// NewJanClient creates a new client for the janecho-server.
func NewJanClient(config Config) (*JanClient, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("janecho-server URL cannot be empty")
	}
	if config.Name == "" {
		config.Name = "janecho"
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.Model == "" {
		config.Model = "default"
	}

	return &JanClient{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}, nil
}

// Name returns the provider's name.
func (c *JanClient) Name() string {
	return c.config.Name
}

// Available checks if the janecho-server is reachable.
func (c *JanClient) Available() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to reach the health endpoint or models endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", c.config.URL+"/v1/models", nil)
	if err != nil {
		return false
	}

	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// MaxTokens returns the maximum tokens supported.
// This is a default value; can be enhanced to fetch from model info.
func (c *JanClient) MaxTokens() int {
	return 4096
}

// Generate sends a request to the janecho-server chat completions endpoint.
func (c *JanClient) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	// Build the request payload
	messages := []ChatMessage{{Role: "user", Content: prompt}}

	// Add system prompt if provided
	if opts.SystemPrompt != "" {
		messages = append([]ChatMessage{{Role: "system", Content: opts.SystemPrompt}}, messages...)
	}

	payload := ChatCompletionRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		TopP:        opts.TopP,
		Stop:        opts.Stop,
		Stream:      false,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.URL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("janecho-server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract content
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

// StreamGenerate generates streaming text from janecho-server.
func (c *JanClient) StreamGenerate(ctx context.Context, prompt string, opts llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	outChan := make(chan llm.StreamChunk, 10)

	// Build the request payload with streaming enabled
	messages := []ChatMessage{{Role: "user", Content: prompt}}

	if opts.SystemPrompt != "" {
		messages = append([]ChatMessage{{Role: "system", Content: opts.SystemPrompt}}, messages...)
	}

	payload := ChatCompletionRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		TopP:        opts.TopP,
		Stop:        opts.Stop,
		Stream:      true,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		outChan <- llm.StreamChunk{Error: fmt.Errorf("failed to marshal request: %w", err)}
		close(outChan)
		return outChan, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.URL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		outChan <- llm.StreamChunk{Error: fmt.Errorf("failed to create request: %w", err)}
		close(outChan)
		return outChan, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	// Start streaming in a goroutine
	go func() {
		defer close(outChan)

		resp, err := c.client.Do(req)
		if err != nil {
			outChan <- llm.StreamChunk{Error: fmt.Errorf("failed to send request: %w", err)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			outChan <- llm.StreamChunk{Error: fmt.Errorf("janecho-server returned status %d: %s", resp.StatusCode, string(body))}
			return
		}

		// Read SSE stream
		decoder := json.NewDecoder(resp.Body)
		for {
			var streamResp ChatCompletionStreamResponse
			if err := decoder.Decode(&streamResp); err != nil {
				if err == io.EOF {
					break
				}
				outChan <- llm.StreamChunk{Error: fmt.Errorf("failed to decode stream: %w", err)}
				return
			}

			if len(streamResp.Choices) > 0 {
				delta := streamResp.Choices[0].Delta
				if delta.Content != "" {
					outChan <- llm.StreamChunk{Content: delta.Content, Done: false}
				}
				if streamResp.Choices[0].FinishReason != "" {
					outChan <- llm.StreamChunk{Done: true}
					return
				}
			}
		}
	}()

	return outChan, nil
}
