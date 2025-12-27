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
	"time"
)

// APIProvider implements LLMProvider using external API services
// Supports both OpenAI-compatible APIs and Anthropic Claude
type APIProvider struct {
	apiKey      string
	baseURL     string
	model       string
	provider    string // "openai" or "anthropic" or "openrouter"
	httpClient  *http.Client
	temperature float64
	maxTokens   int
}

// NewAPIProvider creates a new API-based LLM provider
func NewAPIProvider(provider, model string) (*APIProvider, error) {
	var apiKey, baseURL string
	
	switch provider {
	case "anthropic":
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		baseURL = "https://api.anthropic.com/v1"
		if model == "" {
			model = "claude-3-5-sonnet-20241022"
		}
	case "openrouter":
		apiKey = os.Getenv("OPENROUTER_API_KEY")
		baseURL = "https://openrouter.ai/api/v1"
		if model == "" {
			model = "anthropic/claude-3.5-sonnet"
		}
	case "openai":
		apiKey = os.Getenv("OPENAI_API_KEY")
		baseURL = "https://api.openai.com/v1"
		if model == "" {
			model = "gpt-4o-mini"
		}
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
	
	if apiKey == "" {
		return nil, fmt.Errorf("%s API key not found in environment", strings.ToUpper(provider))
	}
	
	return &APIProvider{
		apiKey:      apiKey,
		baseURL:     baseURL,
		model:       model,
		provider:    provider,
		httpClient:  &http.Client{Timeout: 120 * time.Second},
		temperature: 0.7,
		maxTokens:   4096,
	}, nil
}

// Name returns the provider name
func (p *APIProvider) Name() string {
	return p.provider
}

// Available checks if the provider is configured and available
func (p *APIProvider) Available() bool {
	return p.apiKey != ""
}

// MaxTokens returns the maximum tokens supported
func (p *APIProvider) MaxTokens() int {
	return p.maxTokens
}

// Generate implements LLMProvider.Generate
func (p *APIProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	// Apply options
	temp := p.temperature
	if opts.Temperature > 0 {
		temp = opts.Temperature
	}
	
	maxTok := p.maxTokens
	if opts.MaxTokens > 0 {
		maxTok = opts.MaxTokens
	}
	
	if p.provider == "anthropic" {
		return p.generateAnthropic(ctx, opts.SystemPrompt, prompt, temp, maxTok)
	}
	return p.generateOpenAI(ctx, opts.SystemPrompt, prompt, temp, maxTok)
}

// StreamGenerate implements LLMProvider.StreamGenerate
func (p *APIProvider) StreamGenerate(ctx context.Context, prompt string, opts GenerateOptions) (<-chan StreamChunk, error) {
	// For now, implement as non-streaming with single chunk
	// TODO: Implement true streaming
	outChan := make(chan StreamChunk, 1)
	
	go func() {
		defer close(outChan)
		
		result, err := p.Generate(ctx, prompt, opts)
		if err != nil {
			outChan <- StreamChunk{Error: err, Done: true}
			return
		}
		
		outChan <- StreamChunk{Content: result, Done: true}
	}()
	
	return outChan, nil
}

// generateOpenAI uses OpenAI-compatible API
func (p *APIProvider) generateOpenAI(ctx context.Context, system, prompt string, temperature float64, maxTokens int) (string, error) {
	messages := []map[string]string{}
	
	if system != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": system,
		})
	}
	
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": prompt,
	})
	
	requestBody := map[string]interface{}{
		"model":       p.model,
		"messages":    messages,
		"temperature": temperature,
		"max_tokens":  maxTokens,
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}
	
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}
	
	return result.Choices[0].Message.Content, nil
}

// generateAnthropic uses Anthropic Claude API
func (p *APIProvider) generateAnthropic(ctx context.Context, system, prompt string, temperature float64, maxTokens int) (string, error) {
	requestBody := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens": maxTokens,
	}
	
	if system != "" {
		requestBody["system"] = system
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}
	
	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	if len(result.Content) == 0 {
		return "", fmt.Errorf("no response from API")
	}
	
	return result.Content[0].Text, nil
}

// SetTemperature sets the temperature for generation
func (p *APIProvider) SetTemperature(temp float64) {
	p.temperature = temp
}

// SetMaxTokens sets the maximum tokens for generation
func (p *APIProvider) SetMaxTokens(tokens int) {
	p.maxTokens = tokens
}

// GetModel returns the current model name
func (p *APIProvider) GetModel() string {
	return p.model
}

// GetProvider returns the provider name
func (p *APIProvider) GetProvider() string {
	return p.provider
}
