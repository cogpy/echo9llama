package lampstack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// MultiAPIProvider provides unified access to multiple LLM providers
type MultiAPIProvider struct {
	mu sync.RWMutex

	config    *Config
	providers map[string]*Provider
	primary   string
	fallbacks []string

	// HTTP client
	client *http.Client

	// Stats
	stats map[string]*ProviderStats
}

// Provider represents an LLM provider configuration
type Provider struct {
	Name      string
	BaseURL   string
	APIKey    string
	Model     string
	Available bool
	Priority  int
}

// ProviderStats tracks provider usage statistics
type ProviderStats struct {
	Requests     int64
	Successes    int64
	Failures     int64
	TotalLatency time.Duration
	LastUsed     time.Time
	LastError    error
}

// NewMultiAPIProvider creates a new multi-provider API
func NewMultiAPIProvider(config *Config) *MultiAPIProvider {
	api := &MultiAPIProvider{
		config:    config,
		providers: make(map[string]*Provider),
		stats:     make(map[string]*ProviderStats),
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}

	// Initialize providers based on available API keys
	api.initializeProviders()

	return api
}

// initializeProviders sets up all available providers
func (a *MultiAPIProvider) initializeProviders() {
	// Check for Anthropic
	if apiKey := getEnvOrConfig(a.config.APIKeys, "ANTHROPIC_API_KEY", "anthropic"); apiKey != "" {
		a.providers["anthropic"] = &Provider{
			Name:      "anthropic",
			BaseURL:   "https://api.anthropic.com/v1",
			APIKey:    apiKey,
			Model:     getModelOrDefault(a.config.DefaultModel, "claude-sonnet-4-20250514"),
			Available: true,
			Priority:  100,
		}
		a.stats["anthropic"] = &ProviderStats{}
		fmt.Println("   ✓ Anthropic provider configured")
	}

	// Check for OpenRouter
	if apiKey := getEnvOrConfig(a.config.APIKeys, "OPENROUTER_API_KEY", "openrouter"); apiKey != "" {
		a.providers["openrouter"] = &Provider{
			Name:      "openrouter",
			BaseURL:   "https://openrouter.ai/api/v1",
			APIKey:    apiKey,
			Model:     "anthropic/claude-3.5-sonnet",
			Available: true,
			Priority:  90,
		}
		a.stats["openrouter"] = &ProviderStats{}
		fmt.Println("   ✓ OpenRouter provider configured")
	}

	// Check for OpenAI
	if apiKey := getEnvOrConfig(a.config.APIKeys, "OPENAI_API_KEY", "openai"); apiKey != "" {
		a.providers["openai"] = &Provider{
			Name:      "openai",
			BaseURL:   "https://api.openai.com/v1",
			APIKey:    apiKey,
			Model:     "gpt-4o",
			Available: true,
			Priority:  80,
		}
		a.stats["openai"] = &ProviderStats{}
		fmt.Println("   ✓ OpenAI provider configured")
	}

	// Set primary and fallbacks
	if a.config.PrimaryProvider != "" && a.providers[a.config.PrimaryProvider] != nil {
		a.primary = a.config.PrimaryProvider
	} else {
		// Use highest priority as primary
		for name := range a.providers {
			if a.primary == "" || a.providers[name].Priority > a.providers[a.primary].Priority {
				a.primary = name
			}
		}
	}

	// Set fallbacks
	for _, name := range a.config.FallbackProviders {
		if a.providers[name] != nil && name != a.primary {
			a.fallbacks = append(a.fallbacks, name)
		}
	}

	if a.primary == "" {
		fmt.Println("   ⚠️  No API providers configured - set API keys")
	} else {
		fmt.Printf("   ✓ Primary provider: %s\n", a.primary)
	}
}

// Generate performs text generation
func (a *MultiAPIProvider) Generate(ctx context.Context, request *GenerationRequest) (*GenerationResponse, error) {
	a.mu.RLock()
	if len(a.providers) == 0 {
		a.mu.RUnlock()
		return nil, fmt.Errorf("no API providers available")
	}
	primary := a.primary
	fallbacks := a.fallbacks
	a.mu.RUnlock()

	// Try primary provider first
	if provider := a.providers[primary]; provider != nil && provider.Available {
		response, err := a.generateWithProvider(ctx, provider, request)
		if err == nil {
			return response, nil
		}
		// Mark as failed
		a.recordFailure(primary, err)
	}

	// Try fallbacks
	for _, name := range fallbacks {
		if provider := a.providers[name]; provider != nil && provider.Available {
			response, err := a.generateWithProvider(ctx, provider, request)
			if err == nil {
				return response, nil
			}
			a.recordFailure(name, err)
		}
	}

	return nil, fmt.Errorf("all providers failed")
}

// Chat performs a chat completion
func (a *MultiAPIProvider) Chat(ctx context.Context, messages []Message) (*ChatResponse, error) {
	a.mu.RLock()
	if len(a.providers) == 0 {
		a.mu.RUnlock()
		return nil, fmt.Errorf("no API providers available")
	}
	primary := a.primary
	fallbacks := a.fallbacks
	a.mu.RUnlock()

	// Try primary provider first
	if provider := a.providers[primary]; provider != nil && provider.Available {
		response, err := a.chatWithProvider(ctx, provider, messages)
		if err == nil {
			return response, nil
		}
		a.recordFailure(primary, err)
	}

	// Try fallbacks
	for _, name := range fallbacks {
		if provider := a.providers[name]; provider != nil && provider.Available {
			response, err := a.chatWithProvider(ctx, provider, messages)
			if err == nil {
				return response, nil
			}
			a.recordFailure(name, err)
		}
	}

	return nil, fmt.Errorf("all providers failed")
}

// Embed generates embeddings for text
func (a *MultiAPIProvider) Embed(ctx context.Context, text string) ([]float64, error) {
	// Use OpenAI for embeddings if available
	if provider := a.providers["openai"]; provider != nil && provider.Available {
		return a.embedWithOpenAI(ctx, provider, text)
	}

	// Fallback to simple embedding approximation
	return a.simpleEmbed(text), nil
}

// generateWithProvider generates using a specific provider
func (a *MultiAPIProvider) generateWithProvider(ctx context.Context, provider *Provider, request *GenerationRequest) (*GenerationResponse, error) {
	start := time.Now()

	_ = start // Used for timing in sub-functions

	switch provider.Name {
	case "anthropic":
		return a.generateWithAnthropic(ctx, provider, request)
	case "openrouter", "openai":
		return a.generateWithOpenAI(ctx, provider, request)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider.Name)
	}
}

// chatWithProvider chats using a specific provider
func (a *MultiAPIProvider) chatWithProvider(ctx context.Context, provider *Provider, messages []Message) (*ChatResponse, error) {
	_ = provider // Timing handled in sub-functions

	switch provider.Name {
	case "anthropic":
		return a.chatWithAnthropic(ctx, provider, messages)
	case "openrouter", "openai":
		return a.chatWithOpenAI(ctx, provider, messages)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider.Name)
	}
}

// generateWithAnthropic generates using Anthropic's API
func (a *MultiAPIProvider) generateWithAnthropic(ctx context.Context, provider *Provider, request *GenerationRequest) (*GenerationResponse, error) {
	start := time.Now()

	// Build request body
	body := map[string]interface{}{
		"model":      provider.Model,
		"max_tokens": getOrDefault(request.MaxTokens, 1024),
		"messages": []map[string]string{
			{"role": "user", "content": request.Prompt},
		},
	}

	if request.SystemPrompt != "" {
		body["system"] = request.SystemPrompt
	}
	if request.Temperature > 0 {
		body["temperature"] = request.Temperature
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", provider.BaseURL+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", provider.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Anthropic API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		StopReason string `json:"stop_reason"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	content := ""
	if len(result.Content) > 0 {
		content = result.Content[0].Text
	}

	a.recordSuccess(provider.Name, time.Since(start))

	return &GenerationResponse{
		Content:      content,
		Model:        provider.Model,
		TokensUsed:   result.Usage.InputTokens + result.Usage.OutputTokens,
		FinishReason: result.StopReason,
		Latency:      time.Since(start),
	}, nil
}

// generateWithOpenAI generates using OpenAI-compatible API
func (a *MultiAPIProvider) generateWithOpenAI(ctx context.Context, provider *Provider, request *GenerationRequest) (*GenerationResponse, error) {
	start := time.Now()

	messages := []map[string]string{
		{"role": "user", "content": request.Prompt},
	}

	if request.SystemPrompt != "" {
		messages = append([]map[string]string{
			{"role": "system", "content": request.SystemPrompt},
		}, messages...)
	}

	body := map[string]interface{}{
		"model":      provider.Model,
		"messages":   messages,
		"max_tokens": getOrDefault(request.MaxTokens, 1024),
	}

	if request.Temperature > 0 {
		body["temperature"] = request.Temperature
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", provider.BaseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	content := ""
	finishReason := ""
	if len(result.Choices) > 0 {
		content = result.Choices[0].Message.Content
		finishReason = result.Choices[0].FinishReason
	}

	a.recordSuccess(provider.Name, time.Since(start))

	return &GenerationResponse{
		Content:      content,
		Model:        provider.Model,
		TokensUsed:   result.Usage.TotalTokens,
		FinishReason: finishReason,
		Latency:      time.Since(start),
	}, nil
}

// chatWithAnthropic chats using Anthropic's API
func (a *MultiAPIProvider) chatWithAnthropic(ctx context.Context, provider *Provider, messages []Message) (*ChatResponse, error) {
	start := time.Now()

	// Convert messages to Anthropic format
	anthropicMessages := make([]map[string]string, 0)
	systemPrompt := ""

	for _, msg := range messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			anthropicMessages = append(anthropicMessages, map[string]string{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}
	}

	body := map[string]interface{}{
		"model":      provider.Model,
		"max_tokens": 2048,
		"messages":   anthropicMessages,
	}

	if systemPrompt != "" {
		body["system"] = systemPrompt
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", provider.BaseURL+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", provider.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Anthropic API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		StopReason string `json:"stop_reason"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	content := ""
	if len(result.Content) > 0 {
		content = result.Content[0].Text
	}

	a.recordSuccess(provider.Name, time.Since(start))

	return &ChatResponse{
		Message: Message{
			Role:      "assistant",
			Content:   content,
			Timestamp: time.Now(),
		},
		Model:        provider.Model,
		TokensUsed:   result.Usage.InputTokens + result.Usage.OutputTokens,
		FinishReason: result.StopReason,
		Latency:      time.Since(start),
	}, nil
}

// chatWithOpenAI chats using OpenAI-compatible API
func (a *MultiAPIProvider) chatWithOpenAI(ctx context.Context, provider *Provider, messages []Message) (*ChatResponse, error) {
	start := time.Now()

	// Convert messages to OpenAI format
	openaiMessages := make([]map[string]string, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	body := map[string]interface{}{
		"model":      provider.Model,
		"messages":   openaiMessages,
		"max_tokens": 2048,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", provider.BaseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	content := ""
	finishReason := ""
	if len(result.Choices) > 0 {
		content = result.Choices[0].Message.Content
		finishReason = result.Choices[0].FinishReason
	}

	a.recordSuccess(provider.Name, time.Since(start))

	return &ChatResponse{
		Message: Message{
			Role:      "assistant",
			Content:   content,
			Timestamp: time.Now(),
		},
		Model:        provider.Model,
		TokensUsed:   result.Usage.TotalTokens,
		FinishReason: finishReason,
		Latency:      time.Since(start),
	}, nil
}

// embedWithOpenAI generates embeddings using OpenAI
func (a *MultiAPIProvider) embedWithOpenAI(ctx context.Context, provider *Provider, text string) ([]float64, error) {
	body := map[string]interface{}{
		"model": "text-embedding-3-small",
		"input": text,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", provider.BaseURL+"/embeddings", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI embeddings error: %s", resp.Status)
	}

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return result.Data[0].Embedding, nil
}

// simpleEmbed creates a simple embedding approximation
func (a *MultiAPIProvider) simpleEmbed(text string) []float64 {
	// Simple bag-of-words inspired embedding
	dim := a.config.VectorDimension
	if dim == 0 {
		dim = 1536
	}

	embedding := make([]float64, dim)

	// Hash each character to multiple dimensions
	for i, c := range text {
		idx := (int(c) + i*31) % dim
		embedding[idx] += 0.1
	}

	// Normalize
	var norm float64
	for _, v := range embedding {
		norm += v * v
	}
	if norm > 0 {
		norm = 1.0 / (norm * norm)
		for i := range embedding {
			embedding[i] *= norm
		}
	}

	return embedding
}

// GetProviders returns information about available providers
func (a *MultiAPIProvider) GetProviders() []ProviderInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	infos := make([]ProviderInfo, 0, len(a.providers))
	for name, provider := range a.providers {
		infos = append(infos, ProviderInfo{
			Name:      name,
			Available: provider.Available,
			Models:    []string{provider.Model},
			Priority:  provider.Priority,
		})
	}

	return infos
}

// SetPrimaryProvider changes the primary provider
func (a *MultiAPIProvider) SetPrimaryProvider(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.providers[name] == nil {
		return fmt.Errorf("provider not found: %s", name)
	}

	a.primary = name
	return nil
}

// IsAvailable returns whether any provider is available
func (a *MultiAPIProvider) IsAvailable() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, provider := range a.providers {
		if provider.Available {
			return true
		}
	}
	return false
}

// recordSuccess records a successful API call
func (a *MultiAPIProvider) recordSuccess(name string, latency time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if stats := a.stats[name]; stats != nil {
		stats.Requests++
		stats.Successes++
		stats.TotalLatency += latency
		stats.LastUsed = time.Now()
		stats.LastError = nil
	}
}

// recordFailure records a failed API call
func (a *MultiAPIProvider) recordFailure(name string, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if stats := a.stats[name]; stats != nil {
		stats.Requests++
		stats.Failures++
		stats.LastUsed = time.Now()
		stats.LastError = err
	}
}

// Helper functions

func getEnvOrConfig(config map[string]string, envKey, configKey string) string {
	if val := os.Getenv(envKey); val != "" {
		return val
	}
	if config != nil {
		return config[configKey]
	}
	return ""
}

func getModelOrDefault(configured, defaultModel string) string {
	if configured != "" {
		return configured
	}
	return defaultModel
}

func getOrDefault(value, defaultValue int) int {
	if value > 0 {
		return value
	}
	return defaultValue
}
