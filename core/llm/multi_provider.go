package llm

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// MultiProviderLLM manages multiple LLM providers with fallback and load balancing
type MultiProviderLLM struct {
	providers []Provider
	mu        sync.RWMutex
	stats     map[string]*ProviderStats
}

// ProviderStats tracks provider performance
type ProviderStats struct {
	TotalCalls    int64
	SuccessCalls  int64
	FailedCalls   int64
	TotalLatency  time.Duration
	LastUsed      time.Time
	Available     bool
}

// Provider interface for LLM providers
type Provider interface {
	Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error)
	Name() string
	Available() bool
	MaxTokens() int
}

// NewMultiProviderLLM creates a new multi-provider LLM with auto-detection
func NewMultiProviderLLM() *MultiProviderLLM {
	mp := &MultiProviderLLM{
		providers: make([]Provider, 0),
		stats:     make(map[string]*ProviderStats),
	}

	// Auto-detect and initialize available providers
	mp.initializeProviders()

	return mp
}

// initializeProviders detects and initializes available LLM providers
func (mp *MultiProviderLLM) initializeProviders() {
// 	// Try Local GGUF model first (highest priority for true autonomy)
// 	if modelPath := os.Getenv("LOCAL_MODEL_PATH"); modelPath != "" {
// 		provider := NewLocalGGUFProvider(modelPath)
// 		if provider.Available() {
// 			mp.AddProvider(provider)
// 			fmt.Println("✓ Local GGUF model provider initialized")
// 		}
// 	}

	// Try Anthropic Claude (high priority for autonomous reasoning)
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		model := os.Getenv("ECHO_ANTHROPIC_MODEL")
		if model == "" {
			model = "claude-sonnet-4-5"
		}
		provider := NewAnthropicProvider(model)
		mp.AddProvider(provider)
		fmt.Println("✓ Anthropic Claude provider initialized")
	}

	// Try OpenRouter (good fallback with multiple models)
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		// Use a fast, capable model for autonomous thought generation
		model := os.Getenv("ECHO_OPENROUTER_MODEL")
		if model == "" {
			model = "anthropic/claude-sonnet-4.5"
		}
		provider := NewOpenRouterProvider(model)
		mp.AddProvider(provider)
		fmt.Println("✓ OpenRouter provider initialized")
	}

	// Try OpenAI (if available)
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		model := os.Getenv("ECHO_OPENAI_MODEL")
		if model == "" {
			model = "gpt-4o"
		}
		provider := NewOpenAIProvider(model)
		mp.AddProvider(provider)
		fmt.Println("✓ OpenAI provider initialized")
	}

	// Fallback to simple provider if no API keys available
	if len(mp.providers) == 0 {
		provider := &SimpleFallbackProvider{}
		mp.AddProvider(provider)
		fmt.Println("⚠ Using simple fallback provider (no API keys found)")
	}
}

// AddProvider adds a provider to the pool
func (mp *MultiProviderLLM) AddProvider(provider Provider) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.providers = append(mp.providers, provider)
	mp.stats[provider.Name()] = &ProviderStats{
		Available: provider.Available(),
		LastUsed:  time.Now(),
	}
}

// Generate generates text using the best available provider
func (mp *MultiProviderLLM) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	mp.mu.RLock()
	providers := make([]Provider, len(mp.providers))
	copy(providers, mp.providers)
	mp.mu.RUnlock()

	var lastErr error
	
	// Try providers in order of priority
	for _, provider := range providers {
		if !provider.Available() {
			continue
		}

		// Track timing
		start := time.Now()
		
		// Attempt generation
		result, err := provider.Generate(ctx, prompt, opts)
		latency := time.Since(start)

		// Update stats
		mp.updateStats(provider.Name(), err == nil, latency)

		if err == nil {
			return result, nil
		}

		lastErr = err
	}

	if lastErr != nil {
		return "", fmt.Errorf("all providers failed, last error: %w", lastErr)
	}

	return "", fmt.Errorf("no available providers")
}

// StreamGenerate generates text with streaming using the best available provider
func (mp *MultiProviderLLM) StreamGenerate(ctx context.Context, prompt string, opts GenerateOptions) (<-chan StreamChunk, error) {
	mp.mu.RLock()
	providers := make([]Provider, len(mp.providers))
	copy(providers, mp.providers)
	mp.mu.RUnlock()

	var lastErr error

	// Try providers in order of priority
	for _, provider := range providers {
		if !provider.Available() {
			continue
		}

		// Check if provider supports streaming
		if streamer, ok := provider.(interface {
			StreamGenerate(context.Context, string, GenerateOptions) (<-chan StreamChunk, error)
		}); ok {
			ch, err := streamer.StreamGenerate(ctx, prompt, opts)
			if err == nil {
				return ch, nil
			}
			lastErr = err
		} else {
			// Fallback to non-streaming
			ch := make(chan StreamChunk, 1)
			go func() {
				defer close(ch)
				result, err := provider.Generate(ctx, prompt, opts)
				if err != nil {
					ch <- StreamChunk{Error: err, Done: true}
				} else {
					ch <- StreamChunk{Content: result, Done: true}
				}
			}()
			return ch, nil
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
	}

	return nil, fmt.Errorf("no available providers")
}

// updateStats updates provider statistics
func (mp *MultiProviderLLM) updateStats(providerName string, success bool, latency time.Duration) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	stats, exists := mp.stats[providerName]
	if !exists {
		stats = &ProviderStats{Available: true}
		mp.stats[providerName] = stats
	}

	stats.TotalCalls++
	stats.LastUsed = time.Now()
	stats.TotalLatency += latency

	if success {
		stats.SuccessCalls++
		stats.Available = true
	} else {
		stats.FailedCalls++
		// Mark as unavailable if failure rate is too high
		if stats.FailedCalls > 3 && float64(stats.FailedCalls)/float64(stats.TotalCalls) > 0.5 {
			stats.Available = false
		}
	}
}

// GetStats returns provider statistics
func (mp *MultiProviderLLM) GetStats() map[string]*ProviderStats {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	// Return a copy
	statsCopy := make(map[string]*ProviderStats)
	for k, v := range mp.stats {
		statsCopy[k] = &ProviderStats{
			TotalCalls:   v.TotalCalls,
			SuccessCalls: v.SuccessCalls,
			FailedCalls:  v.FailedCalls,
			TotalLatency: v.TotalLatency,
			LastUsed:     v.LastUsed,
			Available:    v.Available,
		}
	}

	return statsCopy
}

// Name returns the multi-provider name
func (mp *MultiProviderLLM) Name() string {
	return "MultiProvider"
}

// Available returns true if any provider is available
func (mp *MultiProviderLLM) Available() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	for _, provider := range mp.providers {
		if provider.Available() {
			return true
		}
	}

	return false
}

// MaxTokens returns the maximum tokens of the best provider
func (mp *MultiProviderLLM) MaxTokens() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	maxTokens := 0
	for _, provider := range mp.providers {
		if provider.Available() && provider.MaxTokens() > maxTokens {
			maxTokens = provider.MaxTokens()
		}
	}

	if maxTokens == 0 {
		return 4096 // Default fallback
	}

	return maxTokens
}
