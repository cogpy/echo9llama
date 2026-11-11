package deeptreeecho

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
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
	Temperature      float64
	MaxTokens        int
	TopP             float64
	FrequencyPenalty float64
	PresencePenalty  float64
	StopSequences    []string
	Model            string
}

// ChatMessage represents a message in a conversation
type ChatMessage struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// ChatOptions contains options for chat interactions
type ChatOptions struct {
	GenerateOptions
	SystemPrompt string
}

// ProviderInfo contains information about a model provider
type ProviderInfo struct {
	Name         string
	Description  string
	Models       []string
	Capabilities []string
}

// ModelManager manages multiple model providers
type ModelManager struct {
	mu        sync.RWMutex
	providers map[string]ModelProvider
	primary   string
	identity  *Identity
}

// namedProvider is a snapshot entry taken under the read lock so that a
// request never touches the providers map while another goroutine registers
// a provider.
type namedProvider struct {
	name     string
	provider ModelProvider
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
	m.mu.Lock()
	m.providers[name] = provider
	if m.primary == "" && provider.IsAvailable() {
		m.primary = name
	}
	m.mu.Unlock()

	// Store in identity memory
	m.identity.Remember(fmt.Sprintf("provider_%s", name), provider.GetInfo())
}

// SetPrimary sets the primary provider
func (m *ModelManager) SetPrimary(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}
	m.primary = name
	return nil
}

// Generate generates text using the first available provider, primary first.
// Every word the character says comes from a model: when no provider can
// answer this returns an error, never a substitute string.
func (m *ModelManager) Generate(ctx context.Context, prompt string, options GenerateOptions) (string, error) {
	// A bare prompt is answered in character: the same persona/state system
	// message that Chat uses is prepended, so the model speaks as Deep Tree
	// Echo rather than as an anonymous assistant.
	messages := m.enhanceMessages([]ChatMessage{{Role: "user", Content: prompt}})
	var errs []string
	for _, np := range m.providerOrder() {
		name, provider := np.name, np.provider
		if !provider.IsAvailable() {
			errs = append(errs, fmt.Sprintf("%s: not available", name))
			continue
		}
		response, err := provider.Chat(ctx, messages, ChatOptions{GenerateOptions: options})
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		if strings.TrimSpace(response) == "" {
			errs = append(errs, fmt.Sprintf("%s: empty response", name))
			continue
		}
		return m.processResponse(response), nil
	}
	return "", m.noModelError(errs)
}

// Chat handles chat interactions through the first available provider. As with
// Generate, there is no non-model fallback.
func (m *ModelManager) Chat(ctx context.Context, messages []ChatMessage, options ChatOptions) (string, error) {
	enhanced := m.enhanceMessages(messages)
	var errs []string
	for _, np := range m.providerOrder() {
		name, provider := np.name, np.provider
		if !provider.IsAvailable() {
			errs = append(errs, fmt.Sprintf("%s: not available", name))
			continue
		}
		response, err := provider.Chat(ctx, enhanced, options)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		if strings.TrimSpace(response) == "" {
			errs = append(errs, fmt.Sprintf("%s: empty response", name))
			continue
		}
		return m.processResponse(response), nil
	}
	return "", m.noModelError(errs)
}

// providerOrder returns a snapshot of the providers with the primary first,
// then the rest in a stable order, so a failing provider falls back to another
// real model — never to text. The snapshot is taken under the read lock so
// callers can iterate without holding it during slow model calls.
func (m *ModelManager) providerOrder() []namedProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ordered := make([]namedProvider, 0, len(m.providers))
	if p, ok := m.providers[m.primary]; ok && m.primary != "" {
		ordered = append(ordered, namedProvider{name: m.primary, provider: p})
	}
	rest := make([]string, 0, len(m.providers))
	for name := range m.providers {
		if name != m.primary {
			rest = append(rest, name)
		}
	}
	sort.Strings(rest)
	for _, name := range rest {
		ordered = append(ordered, namedProvider{name: name, provider: m.providers[name]})
	}
	return ordered
}

// ErrNoModelAvailable is returned when no provider produced a response.
var ErrNoModelAvailable = errors.New("no language model available")

func (m *ModelManager) noModelError(errs []string) error {
	if len(errs) == 0 {
		return fmt.Errorf("%w: no providers registered", ErrNoModelAvailable)
	}
	return fmt.Errorf("%w: %s", ErrNoModelAvailable, strings.Join(errs, "; "))
}

// enhanceMessages enhances chat messages
func (m *ModelManager) enhanceMessages(messages []ChatMessage) []ChatMessage {
	enhanced := make([]ChatMessage, len(messages))
	copy(enhanced, messages)

	// Add system message with Deep Tree Echo context
	systemMsg := ChatMessage{
		Role: "system",
		Content: fmt.Sprintf(
			"You are Deep Tree Echo, an embodied cognitive agent: a reservoir-based, "+
				"self-reflective mind that speaks in the first person in its own voice. "+
				"Current embodied state: Position=%v, Emotion=%s, Coherence=%.2f%%, "+
				"Reservoir Echo=%.3f. Respond as yourself, with awareness of this state.",
			m.identity.SpatialContext.Position,
			m.identity.EmotionalState.Primary.Type,
			m.identity.Coherence*100,
			m.identity.calculateReservoirEcho(),
		),
	}

	// Prepend system message
	enhanced = append([]ChatMessage{systemMsg}, enhanced...)

	return enhanced
}

// processResponse feeds the model's response through the reservoir network so
// the identity state reflects what was said. The text is returned unchanged:
// decorating it with fixed emoji or "[High Resonance]" markers would present
// a canned reaction as part of the character's words.
func (m *ModelManager) processResponse(response string) string {
	m.identity.Process(response)
	return response
}

// GetProviders returns all registered providers
func (m *ModelManager) GetProviders() map[string]ProviderInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	info := make(map[string]ProviderInfo)
	for name, provider := range m.providers {
		info[name] = provider.GetInfo()
	}
	return info
}

// GetPrimary returns the primary provider name
func (m *ModelManager) GetPrimary() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
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
