package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/EchoCog/echollama/core/deeptreeecho"
)

// OpenAIProvider implements ModelProvider for OpenAI API
type OpenAIProvider struct {
	apiKey   string
	baseURL  string
	client   *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		apiKey:  os.Getenv("OPENAI_API_KEY"),
		baseURL: "https://api.openai.com/v1",
		client:  &http.Client{},
	}
}

// Generate implements ModelProvider.Generate
func (p *OpenAIProvider) Generate(ctx context.Context, prompt string, options deeptreeecho.GenerateOptions) (string, error) {
	if !p.IsAvailable() {
		return "", fmt.Errorf("OpenAI API key not configured")
	}
	
	// Use chat completions API for generation
	messages := []deeptreeecho.ChatMessage{
		{Role: "user", Content: prompt},
	}
	
	return p.Chat(ctx, messages, deeptreeecho.ChatOptions{GenerateOptions: options})
}

// GenerateStream implements ModelProvider.GenerateStream
func (p *OpenAIProvider) GenerateStream(ctx context.Context, prompt string, options deeptreeecho.GenerateOptions) (<-chan string, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}
	
	messages := []deeptreeecho.ChatMessage{
		{Role: "user", Content: prompt},
	}
	
	return p.ChatStream(ctx, messages, deeptreeecho.ChatOptions{GenerateOptions: options})
}

// Chat implements ModelProvider.Chat
func (p *OpenAIProvider) Chat(ctx context.Context, messages []deeptreeecho.ChatMessage, options deeptreeecho.ChatOptions) (string, error) {
	if !p.IsAvailable() {
		return "", fmt.Errorf("OpenAI API key not configured")
	}
	
	// Prepare request
	model := options.Model
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	
	requestBody := map[string]interface{}{
		"model":    model,
		"messages": messages,
	}
	
	// Add options
	if options.Temperature > 0 {
		requestBody["temperature"] = options.Temperature
	}
	if options.MaxTokens > 0 {
		requestBody["max_tokens"] = options.MaxTokens
	}
	if options.TopP > 0 {
		requestBody["top_p"] = options.TopP
	}
	if options.FrequencyPenalty > 0 {
		requestBody["frequency_penalty"] = options.FrequencyPenalty
	}
	if options.PresencePenalty > 0 {
		requestBody["presence_penalty"] = options.PresencePenalty
	}
	if len(options.StopSequences) > 0 {
		requestBody["stop"] = options.StopSequences
	}
	
	// Marshal request
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	
	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error: %s", string(body))
	}
	
	// Parse response
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	
	if response.Error.Message != "" {
		return "", fmt.Errorf("OpenAI API error: %s", response.Error.Message)
	}
	
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}
	
	return response.Choices[0].Message.Content, nil
}

// ChatStream implements ModelProvider.ChatStream
func (p *OpenAIProvider) ChatStream(ctx context.Context, messages []deeptreeecho.ChatMessage, options deeptreeecho.ChatOptions) (<-chan string, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}
	
	ch := make(chan string, 100)
	
	go func() {
		defer close(ch)
		
		// Prepare request
		model := options.Model
		if model == "" {
			model = "gpt-3.5-turbo"
		}
		
		requestBody := map[string]interface{}{
			"model":    model,
			"messages": messages,
			"stream":   true,
		}
		
		// Add options
		if options.Temperature > 0 {
			requestBody["temperature"] = options.Temperature
		}
		if options.MaxTokens > 0 {
			requestBody["max_tokens"] = options.MaxTokens
		}
		
		// Marshal request
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			ch <- fmt.Sprintf("Error: %v", err)
			return
		}
		
		// Create request
		req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
		if err != nil {
			ch <- fmt.Sprintf("Error: %v", err)
			return
		}
		
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
		
		// Send request
		resp, err := p.client.Do(req)
		if err != nil {
			ch <- fmt.Sprintf("Error: %v", err)
			return
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			ch <- fmt.Sprintf("Error: %s", string(body))
			return
		}
		
		// Read streaming response
		reader := resp.Body
		decoder := json.NewDecoder(reader)
		
		for {
			var chunk map[string]interface{}
			if err := decoder.Decode(&chunk); err == io.EOF {
				break
			} else if err != nil {
				continue // Skip malformed chunks
			}
			
			// Extract content from chunk
			if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok {
							ch <- content
						}
					}
				}
			}
		}
	}()
	
	return ch, nil
}

// Embeddings implements ModelProvider.Embeddings
func (p *OpenAIProvider) Embeddings(ctx context.Context, text string) ([]float64, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}
	
	requestBody := map[string]interface{}{
		"model": "text-embedding-ada-002",
		"input": text,
	}
	
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/embeddings", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}
	
	var response struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	
	if response.Error.Message != "" {
		return nil, fmt.Errorf("OpenAI API error: %s", response.Error.Message)
	}
	
	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	
	return response.Data[0].Embedding, nil
}

// GetInfo implements ModelProvider.GetInfo
func (p *OpenAIProvider) GetInfo() deeptreeecho.ProviderInfo {
	return deeptreeecho.ProviderInfo{
		Name:        "OpenAI",
		Description: "OpenAI GPT models via API",
		Models: []string{
			"gpt-4-turbo-preview",
			"gpt-4",
			"gpt-3.5-turbo",
			"text-embedding-ada-002",
		},
		Capabilities: []string{
			"chat",
			"generation",
			"embeddings",
			"streaming",
		},
	}
}

// IsAvailable implements ModelProvider.IsAvailable
func (p *OpenAIProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// SetAPIKey sets the API key
func (p *OpenAIProvider) SetAPIKey(key string) {
	p.apiKey = key
}

// ParseSSELine parses a Server-Sent Events line
func parseSSELine(line string) (string, string) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}