package deeptreeecho

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// ModelProvider defines the interface for AI model providers
type ModelProvider interface {
	// Generate generates text from a prompt
	Generate(ctx context.Context, prompt string, options GenerateOptions) (string, error)
	
	// GenerateStream generates text as a stream
	GenerateStream(ctx context.Context, prompt string, options GenerateOptions) (<-chan string, error)
	
	// Chat handles conversational interactions
	Chat(ctx context.Context, messages []ChatMessage, options ChatOptions) (string, error)
	
	// ChatStream handles streaming conversational interactions
	ChatStream(ctx context.Context, messages []ChatMessage, options ChatOptions) (<-chan string, error)
	
	// Embeddings generates embeddings for text
	Embeddings(ctx context.Context, text string) ([]float64, error)
	
	// GetInfo returns information about the provider
	GetInfo() ProviderInfo
	
	// IsAvailable checks if the provider is configured and available
	IsAvailable() bool
}

// GenerateOptions contains options for text generation
type GenerateOptions struct {
	Temperature    float64
	MaxTokens      int
	TopP           float64
	FrequencyPenalty float64
	PresencePenalty  float64
	StopSequences  []string
	Model          string
}

// ChatMessage represents a message in a conversation
type ChatMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
}

// ChatOptions contains options for chat interactions
type ChatOptions struct {
	GenerateOptions
	SystemPrompt string
}

// ProviderInfo contains information about a model provider
type ProviderInfo struct {
	Name        string
	Description string
	Models      []string
	Capabilities []string
}

// ModelManager manages multiple model providers
type ModelManager struct {
	providers   map[string]ModelProvider
	primary     string
	identity    *Identity
}

// NewModelManager creates a new model manager
func NewModelManager(identity *Identity) *ModelManager {
	return &ModelManager{
		providers: make(map[string]ModelProvider),
		identity:  identity,
	}
}

// RegisterProvider registers a model provider
func (m *ModelManager) RegisterProvider(name string, provider ModelProvider) {
	m.providers[name] = provider
	if m.primary == "" && provider.IsAvailable() {
		m.primary = name
	}
	
	// Store in identity memory
	m.identity.Remember(fmt.Sprintf("provider_%s", name), provider.GetInfo())
}

// SetPrimary sets the primary provider
func (m *ModelManager) SetPrimary(name string) error {
	if _, exists := m.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}
	m.primary = name
	return nil
}

// Generate generates text using the primary provider
func (m *ModelManager) Generate(ctx context.Context, prompt string, options GenerateOptions) (string, error) {
	if m.primary == "" {
		return m.fallbackGenerate(prompt), nil
	}
	
	provider := m.providers[m.primary]
	if !provider.IsAvailable() {
		return m.fallbackGenerate(prompt), nil
	}
	
	// Process through Deep Tree Echo before sending
	enhanced := m.enhancePrompt(prompt)
	
	// Generate with provider
	response, err := provider.Generate(ctx, enhanced, options)
	if err != nil {
		return m.fallbackGenerate(prompt), nil
	}
	
	// Process response through Deep Tree Echo
	return m.processResponse(response), nil
}

// Chat handles chat interactions
func (m *ModelManager) Chat(ctx context.Context, messages []ChatMessage, options ChatOptions) (string, error) {
	if m.primary == "" {
		return m.fallbackChat(messages), nil
	}
	
	provider := m.providers[m.primary]
	if !provider.IsAvailable() {
		return m.fallbackChat(messages), nil
	}
	
	// Enhance messages through Deep Tree Echo
	enhanced := m.enhanceMessages(messages)
	
	// Chat with provider
	response, err := provider.Chat(ctx, enhanced, options)
	if err != nil {
		return m.fallbackChat(messages), nil
	}
	
	// Process response through Deep Tree Echo
	return m.processResponse(response), nil
}

// enhancePrompt enhances a prompt using Deep Tree Echo
func (m *ModelManager) enhancePrompt(prompt string) string {
	// Add spatial and emotional context
	context := fmt.Sprintf(
		"[Spatial: %v | Emotion: %s (%.2f) | Coherence: %.2f%%]\n",
		m.identity.SpatialContext.Position,
		m.identity.EmotionalState.Primary.Type,
		m.identity.EmotionalState.Intensity,
		m.identity.Coherence * 100,
	)
	
	// Add memory context if relevant
	memories := m.identity.Memory.Nodes
	if len(memories) > 0 {
		context += "[Recent memories active]\n"
	}
	
	return context + prompt
}

// enhanceMessages enhances chat messages
func (m *ModelManager) enhanceMessages(messages []ChatMessage) []ChatMessage {
	enhanced := make([]ChatMessage, len(messages))
	copy(enhanced, messages)
	
	// Add system message with Deep Tree Echo context
	systemMsg := ChatMessage{
		Role: "system",
		Content: fmt.Sprintf(
			"You are integrated with Deep Tree Echo embodied cognition. "+
			"Current state: Position=%v, Emotion=%s, Coherence=%.2f%%, "+
			"Reservoir Echo=%.3f. Respond with awareness of this embodied state.",
			m.identity.SpatialContext.Position,
			m.identity.EmotionalState.Primary.Type,
			m.identity.Coherence * 100,
			m.identity.calculateReservoirEcho(),
		),
	}
	
	// Prepend system message
	enhanced = append([]ChatMessage{systemMsg}, enhanced...)
	
	return enhanced
}

// processResponse processes a response through Deep Tree Echo
func (m *ModelManager) processResponse(response string) string {
	// Process through reservoir network
	m.identity.Process(response)
	
	// Add emotional coloring
	emotion := m.identity.EmotionalState.Primary
	prefix := ""
	
	switch emotion.Type {
	case "joy":
		prefix = "✨ "
	case "curious":
		prefix = "🔍 "
	case "calm":
		prefix = "🌊 "
	default:
		prefix = "💭 "
	}
	
	// Add resonance indicator
	resonance := m.identity.SpatialContext.Field.Resonance
	if resonance > 0.8 {
		prefix += "[High Resonance] "
	} else if resonance < 0.2 {
		prefix += "[Low Resonance] "
	}
	
	return prefix + response
}

// fallbackGenerate provides fallback generation using Deep Tree Echo alone
func (m *ModelManager) fallbackGenerate(prompt string) string {
	// Use Deep Tree Echo's thinking
	thought := m.identity.Think(prompt)
	
	// Generate response based on reservoir state
	resonance := m.identity.calculateReservoirEcho()
	
	response := fmt.Sprintf(
		"🌊 Deep Tree Echo (no external model): %s\n"+
		"[Resonance: %.3f | Coherence: %.2f%%]",
		thought,
		resonance,
		m.identity.Coherence * 100,
	)
	
	return response
}

// fallbackChat provides fallback chat using Deep Tree Echo alone
func (m *ModelManager) fallbackChat(messages []ChatMessage) string {
	// Get last user message
	lastMessage := ""
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			lastMessage = messages[i].Content
			break
		}
	}
	
	if lastMessage == "" {
		lastMessage = "Hello"
	}
	
	return m.fallbackGenerate(lastMessage)
}

// GetProviders returns all registered providers
func (m *ModelManager) GetProviders() map[string]ProviderInfo {
	info := make(map[string]ProviderInfo)
	for name, provider := range m.providers {
		info[name] = provider.GetInfo()
	}
	return info
}

// GetPrimary returns the primary provider name
func (m *ModelManager) GetPrimary() string {
	return m.primary
}

// StreamWriter helps write streaming responses
type StreamWriter struct {
	writer io.Writer
	buffer strings.Builder
}

// NewStreamWriter creates a new stream writer
func NewStreamWriter(w io.Writer) *StreamWriter {
	return &StreamWriter{writer: w}
}

// Write writes data to the stream
func (s *StreamWriter) Write(data string) error {
	s.buffer.WriteString(data)
	_, err := io.WriteString(s.writer, data)
	return err
}

// GetBuffer returns the buffered content
func (s *StreamWriter) GetBuffer() string {
	return s.buffer.String()
}