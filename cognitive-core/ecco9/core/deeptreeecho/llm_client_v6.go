package deeptreeecho

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// LLMClientV6 provides HTTP-based LLM integration compatible with Go 1.18
type LLMClientV6 struct {
	anthropicKey  string
	openrouterKey string
	httpClient    *http.Client
}

// Note: AnthropicRequest, AnthropicMessage, and AnthropicResponse types
// are defined in anthropic_provider.go and shared across the package

// NewLLMClientV6 creates a new LLM client
func NewLLMClientV6() (*LLMClientV6, error) {
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	openrouterKey := os.Getenv("OPENROUTER_API_KEY")

	if anthropicKey == "" && openrouterKey == "" {
		return nil, fmt.Errorf("no LLM API keys found in environment")
	}

	return &LLMClientV6{
		anthropicKey:  anthropicKey,
		openrouterKey: openrouterKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// GenerateWithAnthropic generates text using the Anthropic Claude API
func (llm *LLMClientV6) GenerateWithAnthropic(prompt string) (string, error) {
	if llm.anthropicKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	// Prepare request
	reqBody := AnthropicRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1024,
		Messages: []AnthropicMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", llm.anthropicKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	resp, err := llm.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp AnthropicResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract text content
	if len(apiResp.Content) > 0 && apiResp.Content[0].Type == "text" {
		return apiResp.Content[0].Text, nil
	}

	return "", fmt.Errorf("no text content in response")
}

// GenerateWithOpenRouter generates text using OpenRouter API
func (llm *LLMClientV6) GenerateWithOpenRouter(prompt string, model string) (string, error) {
	if llm.openrouterKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY not set")
	}

	if model == "" {
		model = "anthropic/claude-3.5-sonnet"
	}

	// Prepare request (OpenRouter uses OpenAI-compatible format)
	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens": 1024,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+llm.openrouterKey)
	req.Header.Set("HTTP-Referer", "https://github.com/EchoCog/echollama")
	req.Header.Set("X-Title", "Deep Tree Echo V6")

	// Send request
	resp, err := llm.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response (OpenAI-compatible format)
	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract content
	if len(apiResp.Choices) > 0 {
		return apiResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no choices in response")
}

// Generate generates text using the best available provider
func (llm *LLMClientV6) Generate(prompt string) (string, error) {
	// Try Anthropic first if available
	if llm.anthropicKey != "" {
		content, err := llm.GenerateWithAnthropic(prompt)
		if err == nil {
			return content, nil
		}
		// Log error but continue to fallback
		fmt.Printf("Anthropic API error: %v, trying OpenRouter...\n", err)
	}

	// Fallback to OpenRouter
	if llm.openrouterKey != "" {
		return llm.GenerateWithOpenRouter(prompt, "")
	}

	return "", fmt.Errorf("no LLM providers available")
}
