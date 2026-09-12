package deeptreeecho

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// LLMProvider interface for different LLM backends
type LLMProvider interface {
	GenerateThought(ctx context.Context, prompt string) (string, error)
	GenerateReflection(ctx context.Context, context string) (string, error)
	IsAvailable() bool
	GetName() string
	GetPriority() int
}

// MultiProviderLLM manages multiple LLM providers with fallback
type MultiProviderLLM struct {
	mu        sync.RWMutex
	providers []LLMProvider
	current   int
	stats     map[string]*ProviderStats
}

// ProviderStats tracks usage statistics for a provider
type ProviderStats struct {
	Requests     int64
	Successes    int64
	Failures     int64
	TotalLatency time.Duration
	LastUsed     time.Time
	LastError    error
}

// NewMultiProviderLLM creates a new multi-provider LLM manager
func NewMultiProviderLLM() *MultiProviderLLM {
	m := &MultiProviderLLM{
		providers: make([]LLMProvider, 0),
		stats:     make(map[string]*ProviderStats),
	}

	// Initialize providers based on available API keys
	m.initializeProviders()

	return m
}

// initializeProviders sets up all available LLM providers
func (m *MultiProviderLLM) initializeProviders() {
	// Try Anthropic (Claude)
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		provider := NewAnthropicProvider(apiKey, "claude-3-5-sonnet-20241022")
		if provider != nil {
			m.AddProvider(provider)
			fmt.Println("✅ Initialized Anthropic Claude provider")
		}
	}

	// Try OpenRouter (multi-model access)
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		provider := NewOpenRouterProvider(apiKey, "anthropic/claude-3.5-sonnet")
		if provider != nil {
			m.AddProvider(provider)
			fmt.Println("✅ Initialized OpenRouter provider")
		}
	}

	// Try OpenAI (GPT models)
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		provider := NewOpenAIProvider(apiKey, "gpt-4.1-mini")
		if provider != nil {
			m.AddProvider(provider)
			fmt.Println("✅ Initialized OpenAI provider")
		}
	}

	if len(m.providers) == 0 {
		fmt.Println("⚠️  No LLM providers available - check API keys")
	}
}

// AddProvider adds a new provider to the manager
func (m *MultiProviderLLM) AddProvider(provider LLMProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.providers = append(m.providers, provider)
	m.stats[provider.GetName()] = &ProviderStats{}

	// Sort providers by priority
	m.sortProvidersByPriority()
}

// sortProvidersByPriority sorts providers by their priority (higher first)
func (m *MultiProviderLLM) sortProvidersByPriority() {
	// Simple bubble sort by priority
	for i := 0; i < len(m.providers)-1; i++ {
		for j := 0; j < len(m.providers)-i-1; j++ {
			if m.providers[j].GetPriority() < m.providers[j+1].GetPriority() {
				m.providers[j], m.providers[j+1] = m.providers[j+1], m.providers[j]
			}
		}
	}
}

// GenerateThought generates a thought using available providers with fallback
func (m *MultiProviderLLM) GenerateThought(ctx context.Context, prompt string) (string, error) {
	m.mu.RLock()
	if len(m.providers) == 0 {
		m.mu.RUnlock()
		return "", fmt.Errorf("no LLM providers available")
	}
	current := m.current
	m.mu.RUnlock()

	// Try current provider first
	if current < len(m.providers) {
		provider := m.providers[current]
		if provider.IsAvailable() {
			result, err := m.tryProvider(ctx, provider, prompt, "thought")
			if err == nil {
				return result, nil
			}
		}
	}

	// Fallback to other providers
	m.mu.RLock()
	providers := make([]LLMProvider, len(m.providers))
	copy(providers, m.providers)
	m.mu.RUnlock()

	for i, provider := range providers {
		if i == current {
			continue // Already tried
		}

		if provider.IsAvailable() {
			result, err := m.tryProvider(ctx, provider, prompt, "thought")
			if err == nil {
				// Switch to working provider
				m.mu.Lock()
				m.current = i
				m.mu.Unlock()
				return result, nil
			}
		}
	}

	return "", fmt.Errorf("all LLM providers failed or unavailable")
}

// GenerateReflection generates a reflection using available providers with fallback
func (m *MultiProviderLLM) GenerateReflection(ctx context.Context, contextStr string) (string, error) {
	m.mu.RLock()
	if len(m.providers) == 0 {
		m.mu.RUnlock()
		return "", fmt.Errorf("no LLM providers available")
	}
	current := m.current
	m.mu.RUnlock()

	// Try current provider first
	if current < len(m.providers) {
		provider := m.providers[current]
		if provider.IsAvailable() {
			result, err := m.tryProviderReflection(ctx, provider, contextStr)
			if err == nil {
				return result, nil
			}
		}
	}

	// Fallback to other providers
	m.mu.RLock()
	providers := make([]LLMProvider, len(m.providers))
	copy(providers, m.providers)
	m.mu.RUnlock()

	for i, provider := range providers {
		if i == current {
			continue
		}

		if provider.IsAvailable() {
			result, err := m.tryProviderReflection(ctx, provider, contextStr)
			if err == nil {
				m.mu.Lock()
				m.current = i
				m.mu.Unlock()
				return result, nil
			}
		}
	}

	return "", fmt.Errorf("all LLM providers failed or unavailable")
}

// tryProvider attempts to use a provider and updates statistics
func (m *MultiProviderLLM) tryProvider(ctx context.Context, provider LLMProvider, prompt string, thoughtType string) (string, error) {
	start := time.Now()

	m.mu.Lock()
	stats := m.stats[provider.GetName()]
	stats.Requests++
	m.mu.Unlock()

	result, err := provider.GenerateThought(ctx, prompt)

	latency := time.Since(start)

	m.mu.Lock()
	stats.TotalLatency += latency
	stats.LastUsed = time.Now()
	if err != nil {
		stats.Failures++
		stats.LastError = err
	} else {
		stats.Successes++
	}
	m.mu.Unlock()

	return result, err
}

// tryProviderReflection attempts to use a provider for reflection
func (m *MultiProviderLLM) tryProviderReflection(ctx context.Context, provider LLMProvider, contextStr string) (string, error) {
	start := time.Now()

	m.mu.Lock()
	stats := m.stats[provider.GetName()]
	stats.Requests++
	m.mu.Unlock()

	result, err := provider.GenerateReflection(ctx, contextStr)

	latency := time.Since(start)

	m.mu.Lock()
	stats.TotalLatency += latency
	stats.LastUsed = time.Now()
	if err != nil {
		stats.Failures++
		stats.LastError = err
	} else {
		stats.Successes++
	}
	m.mu.Unlock()

	return result, err
}

// GetStats returns statistics for all providers
func (m *MultiProviderLLM) GetStats() map[string]*ProviderStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy
	stats := make(map[string]*ProviderStats)
	for name, s := range m.stats {
		statsCopy := *s
		stats[name] = &statsCopy
	}
	return stats
}

// GetCurrentProvider returns the name of the current provider
func (m *MultiProviderLLM) GetCurrentProvider() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.current < len(m.providers) {
		return m.providers[m.current].GetName()
	}
	return "none"
}

// GetAvailableProviders returns a list of available provider names
func (m *MultiProviderLLM) GetAvailableProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	available := make([]string, 0)
	for _, provider := range m.providers {
		if provider.IsAvailable() {
			available = append(available, provider.GetName())
		}
	}
	return available
}

// IsAvailable returns true if at least one provider is available
func (m *MultiProviderLLM) IsAvailable() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, provider := range m.providers {
		if provider.IsAvailable() {
			return true
		}
	}
	return false
}
