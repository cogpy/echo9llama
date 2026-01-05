package lampstack

import (
	"context"
	"time"
)

// Message represents a chat message
type Message struct {
	Role      string                 `json:"role"`       // "user", "assistant", "system"
	Content   string                 `json:"content"`
	Name      string                 `json:"name,omitempty"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// GenerationRequest represents a request for text generation
type GenerationRequest struct {
	Prompt       string                 `json:"prompt"`
	SystemPrompt string                 `json:"system_prompt,omitempty"`
	MaxTokens    int                    `json:"max_tokens,omitempty"`
	Temperature  float64                `json:"temperature,omitempty"`
	TopP         float64                `json:"top_p,omitempty"`
	StopSequences []string              `json:"stop_sequences,omitempty"`
	UseCognition bool                   `json:"use_cognition"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// GenerationResponse represents a generation response
type GenerationResponse struct {
	Content       string                 `json:"content"`
	Model         string                 `json:"model"`
	TokensUsed    int                    `json:"tokens_used"`
	FinishReason  string                 `json:"finish_reason"`
	CognitiveInfo *CognitiveInfo         `json:"cognitive_info,omitempty"`
	Latency       time.Duration          `json:"latency"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Message       Message                `json:"message"`
	Model         string                 `json:"model"`
	TokensUsed    int                    `json:"tokens_used"`
	FinishReason  string                 `json:"finish_reason"`
	CognitiveInfo *CognitiveInfo         `json:"cognitive_info,omitempty"`
	Latency       time.Duration          `json:"latency"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ThoughtResponse represents a cognitive thought
type ThoughtResponse struct {
	Thought       string          `json:"thought"`
	ThoughtType   ThoughtType     `json:"thought_type"`
	Confidence    float64         `json:"confidence"`
	Associations  []string        `json:"associations,omitempty"`
	CognitiveInfo *CognitiveInfo  `json:"cognitive_info,omitempty"`
	Timestamp     time.Time       `json:"timestamp"`
}

// ThoughtType categorizes different types of thoughts
type ThoughtType int

const (
	ThoughtTypePerception ThoughtType = iota
	ThoughtTypeReflection
	ThoughtTypeSimulation
	ThoughtTypePlanning
	ThoughtTypeCreative
	ThoughtTypeAnalytical
)

func (t ThoughtType) String() string {
	return [...]string{
		"Perception",
		"Reflection",
		"Simulation",
		"Planning",
		"Creative",
		"Analytical",
	}[t]
}

// ReflectionResponse represents a meta-cognitive reflection
type ReflectionResponse struct {
	Reflection    string         `json:"reflection"`
	Insights      []string       `json:"insights,omitempty"`
	WisdomGains   map[string]float64 `json:"wisdom_gains,omitempty"`
	CognitiveInfo *CognitiveInfo `json:"cognitive_info,omitempty"`
	Timestamp     time.Time      `json:"timestamp"`
}

// CognitiveInfo provides information about cognitive processing
type CognitiveInfo struct {
	ProcessingMode    string             `json:"processing_mode"`
	CurrentPhase      int                `json:"current_phase"`
	ActiveEngines     []string           `json:"active_engines"`
	ResonanceLevel    float64            `json:"resonance_level"`
	AwarenessLevel    float64            `json:"awareness_level"`
	EmotionalState    *EmotionalState    `json:"emotional_state,omitempty"`
	WisdomDimensions  map[string]float64 `json:"wisdom_dimensions,omitempty"`
}

// EmotionalState represents the current emotional state
type EmotionalState struct {
	Valence    float64 `json:"valence"`    // -1 to 1 (negative to positive)
	Arousal    float64 `json:"arousal"`    // 0 to 1 (calm to excited)
	Dominance  float64 `json:"dominance"`  // 0 to 1 (submissive to dominant)
	Curiosity  float64 `json:"curiosity"`  // 0 to 1
	Confidence float64 `json:"confidence"` // 0 to 1
}

// CognitiveState represents the overall cognitive state
type CognitiveState struct {
	Phase            string         `json:"phase"`
	Mode             string         `json:"mode"`
	Awareness        float64        `json:"awareness"`
	CognitiveLoad    float64        `json:"cognitive_load"`
	Coherence        float64        `json:"coherence"`
	EnergyLevel      float64        `json:"energy_level"`
	EmotionalState   *EmotionalState `json:"emotional_state"`
	ActiveProcesses  []string       `json:"active_processes"`
	CurrentGoals     []string       `json:"current_goals"`
	RecentThoughts   []string       `json:"recent_thoughts"`
	Timestamp        time.Time      `json:"timestamp"`
}

// MemoryEntry represents a memory stored in the system
type MemoryEntry struct {
	ID          string                 `json:"id"`
	Type        MemoryType             `json:"type"`
	Content     string                 `json:"content"`
	Embedding   []float64              `json:"embedding,omitempty"`
	Importance  float64                `json:"importance"`
	Associations []string              `json:"associations,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	AccessedAt  time.Time              `json:"accessed_at"`
	AccessCount int                    `json:"access_count"`
}

// MemoryType categorizes different types of memories
type MemoryType int

const (
	MemoryTypeEpisodic MemoryType = iota  // Specific experiences
	MemoryTypeSemantic                     // General knowledge
	MemoryTypeProcedural                   // Skills and how-to
	MemoryTypeWorking                      // Temporary/active
)

func (m MemoryType) String() string {
	return [...]string{
		"Episodic",
		"Semantic",
		"Procedural",
		"Working",
	}[m]
}

// ProviderInfo contains information about an LLM provider
type ProviderInfo struct {
	Name        string   `json:"name"`
	Available   bool     `json:"available"`
	Models      []string `json:"models"`
	Priority    int      `json:"priority"`
	RateLimited bool     `json:"rate_limited"`
}

// EchoEngine defines the interface for the Echo cognitive engine
type EchoEngine interface {
	// Lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// Generation
	Generate(ctx context.Context, request *GenerationRequest) (*GenerationResponse, error)
	Chat(ctx context.Context, messages []Message) (*ChatResponse, error)

	// Cognitive operations
	Think(ctx context.Context, prompt string) (*ThoughtResponse, error)
	Reflect(ctx context.Context, contextStr string) (*ReflectionResponse, error)

	// State
	GetState() (*CognitiveState, error)
	GetMetrics() map[string]interface{}
}

// APIProvider defines the interface for LLM API providers
type APIProvider interface {
	// Generation
	Generate(ctx context.Context, request *GenerationRequest) (*GenerationResponse, error)
	Chat(ctx context.Context, messages []Message) (*ChatResponse, error)

	// Embeddings
	Embed(ctx context.Context, text string) ([]float64, error)

	// Provider management
	GetProviders() []ProviderInfo
	SetPrimaryProvider(name string) error
	IsAvailable() bool
}

// MemoryStore defines the interface for memory storage
type MemoryStore interface {
	// CRUD operations
	Store(ctx context.Context, entry *MemoryEntry) error
	Retrieve(ctx context.Context, query string, limit int) ([]*MemoryEntry, error)
	Get(ctx context.Context, id string) (*MemoryEntry, error)
	Update(ctx context.Context, entry *MemoryEntry) error
	Delete(ctx context.Context, id string) error

	// Specialized retrieval
	RetrieveByType(ctx context.Context, memType MemoryType, limit int) ([]*MemoryEntry, error)
	RetrieveByTags(ctx context.Context, tags []string, limit int) ([]*MemoryEntry, error)
	RetrieveRecent(ctx context.Context, limit int) ([]*MemoryEntry, error)

	// Memory management
	Consolidate(ctx context.Context) error
	GetStats() map[string]interface{}
}

// PersistenceLayer defines the interface for state persistence
type PersistenceLayer interface {
	// Save operations
	Save(ctx context.Context, key string, value interface{}) error
	SaveAll(ctx context.Context) error

	// Load operations
	Load(ctx context.Context, key string, value interface{}) error
	LoadAll(ctx context.Context) error

	// Key management
	Keys(ctx context.Context) ([]string, error)
	Exists(ctx context.Context, key string) (bool, error)
	Delete(ctx context.Context, key string) error

	// Lifecycle
	Close() error
}

// StreamHandler handles streaming responses
type StreamHandler func(chunk string) error

// StreamingGeneration represents a streaming generation interface
type StreamingGeneration interface {
	Generate(ctx context.Context, request *GenerationRequest, handler StreamHandler) error
	Chat(ctx context.Context, messages []Message, handler StreamHandler) error
}

// LangChainIntegration provides LangChain-compatible interfaces
type LangChainIntegration interface {
	// LLM interface
	Call(ctx context.Context, prompt string) (string, error)
	Generate(ctx context.Context, prompts []string) ([]string, error)

	// Chat interface
	ChatComplete(ctx context.Context, messages []Message) (Message, error)

	// Embeddings interface
	EmbedDocuments(ctx context.Context, texts []string) ([][]float64, error)
	EmbedQuery(ctx context.Context, text string) ([]float64, error)

	// Tool interface
	RunTool(ctx context.Context, name string, input string) (string, error)
	GetTools() []ToolDefinition
}

// ToolDefinition defines a tool that can be used by LangChain
type ToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema string `json:"input_schema"`
}

// EventType defines cognitive event types
type EventType int

const (
	EventTypeThought EventType = iota
	EventTypeReflection
	EventTypeStateChange
	EventTypeMemoryStore
	EventTypeMemoryRetrieve
	EventTypeGoalSet
	EventTypeGoalComplete
	EventTypeEmergence
)

func (e EventType) String() string {
	return [...]string{
		"Thought",
		"Reflection",
		"StateChange",
		"MemoryStore",
		"MemoryRetrieve",
		"GoalSet",
		"GoalComplete",
		"Emergence",
	}[e]
}

// CognitiveEvent represents an event in the cognitive system
type CognitiveEvent struct {
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// EventHandler handles cognitive events
type EventHandler func(event CognitiveEvent)

// EventBus defines the interface for event distribution
type EventBus interface {
	Publish(event CognitiveEvent)
	Subscribe(eventType EventType, handler EventHandler) string
	Unsubscribe(subscriptionID string)
}
