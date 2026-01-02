// Package lampstack provides a unified integration stack for Deep Tree Echo
// LAMPSTACK = LangChain + API + Memory + Persistence + STACK
//
// This package serves as the foundation layer for all external integrations,
// providing a cohesive API that combines cognitive capabilities with
// standard LLM integration patterns.
package lampstack

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Version of the LAMPSTACK integration layer
const Version = "1.0.0"

// StackName identifies this integration stack
const StackName = "LAMPSTACK Echo"

// Stack represents the complete LAMPSTACK integration
// It provides unified access to all cognitive subsystems
type Stack struct {
	mu sync.RWMutex

	// Configuration
	config *Config

	// Core components
	echo        EchoEngine
	api         APIProvider
	memory      MemoryStore
	persistence PersistenceLayer

	// State tracking
	initialized bool
	startTime   time.Time
	metrics     *StackMetrics
}

// Config holds the configuration for the LAMPSTACK
type Config struct {
	// Echo configuration
	EnableEchobeats      bool          `json:"enable_echobeats"`
	EnableConsciousness  bool          `json:"enable_consciousness"`
	EnableDreaming       bool          `json:"enable_dreaming"`
	CycleDuration        time.Duration `json:"cycle_duration"`

	// API configuration
	PrimaryProvider   string            `json:"primary_provider"`
	FallbackProviders []string          `json:"fallback_providers"`
	APIKeys           map[string]string `json:"api_keys"`
	DefaultModel      string            `json:"default_model"`
	MaxRetries        int               `json:"max_retries"`
	Timeout           time.Duration     `json:"timeout"`

	// Memory configuration
	EnableHypergraph    bool   `json:"enable_hypergraph"`
	EnableEpisodicMemory bool  `json:"enable_episodic_memory"`
	EnableSemanticMemory bool  `json:"enable_semantic_memory"`
	MemoryCapacity      int    `json:"memory_capacity"`
	VectorDimension     int    `json:"vector_dimension"`

	// Persistence configuration
	PersistenceBackend string `json:"persistence_backend"`
	DataDirectory      string `json:"data_directory"`
	AutoSave           bool   `json:"auto_save"`
	SaveInterval       time.Duration `json:"save_interval"`

	// Integration configuration
	EnableLangChain bool `json:"enable_langchain"`
	EnableMetrics   bool `json:"enable_metrics"`
	EnableTracing   bool `json:"enable_tracing"`
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *Config {
	return &Config{
		// Echo defaults
		EnableEchobeats:     true,
		EnableConsciousness: true,
		EnableDreaming:      true,
		CycleDuration:       30 * time.Second,

		// API defaults
		PrimaryProvider:   "anthropic",
		FallbackProviders: []string{"openrouter", "openai"},
		APIKeys:           make(map[string]string),
		DefaultModel:      "claude-sonnet-4-20250514",
		MaxRetries:        3,
		Timeout:           60 * time.Second,

		// Memory defaults
		EnableHypergraph:     true,
		EnableEpisodicMemory: true,
		EnableSemanticMemory: true,
		MemoryCapacity:       10000,
		VectorDimension:      1536,

		// Persistence defaults
		PersistenceBackend: "json",
		DataDirectory:      "data/state",
		AutoSave:           true,
		SaveInterval:       30 * time.Second,

		// Integration defaults
		EnableLangChain: true,
		EnableMetrics:   true,
		EnableTracing:   false,
	}
}

// StackMetrics tracks usage and performance metrics
type StackMetrics struct {
	mu sync.RWMutex

	// Timing
	StartTime    time.Time     `json:"start_time"`
	Uptime       time.Duration `json:"uptime"`

	// Request counts
	TotalRequests      int64 `json:"total_requests"`
	SuccessfulRequests int64 `json:"successful_requests"`
	FailedRequests     int64 `json:"failed_requests"`

	// Cognitive metrics
	ThoughtsGenerated    int64 `json:"thoughts_generated"`
	ReflectionsGenerated int64 `json:"reflections_generated"`
	ConsolidationsCycles int64 `json:"consolidation_cycles"`

	// Memory metrics
	MemoriesStored   int64 `json:"memories_stored"`
	MemoriesRetrieved int64 `json:"memories_retrieved"`

	// Performance
	AverageLatencyMs float64 `json:"average_latency_ms"`
	TotalTokensUsed  int64   `json:"total_tokens_used"`
}

// NewStack creates a new LAMPSTACK instance with the given configuration
func NewStack(config *Config) (*Stack, error) {
	if config == nil {
		config = DefaultConfig()
	}

	s := &Stack{
		config:    config,
		startTime: time.Now(),
		metrics:   &StackMetrics{StartTime: time.Now()},
	}

	return s, nil
}

// Initialize sets up all stack components
func (s *Stack) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.initialized {
		return fmt.Errorf("stack already initialized")
	}

	fmt.Println("🔷 Initializing LAMPSTACK Echo...")

	// Initialize API provider
	if err := s.initializeAPI(ctx); err != nil {
		return fmt.Errorf("failed to initialize API: %w", err)
	}
	fmt.Println("   ✓ API provider initialized")

	// Initialize memory store
	if err := s.initializeMemory(ctx); err != nil {
		return fmt.Errorf("failed to initialize memory: %w", err)
	}
	fmt.Println("   ✓ Memory store initialized")

	// Initialize persistence layer
	if err := s.initializePersistence(ctx); err != nil {
		return fmt.Errorf("failed to initialize persistence: %w", err)
	}
	fmt.Println("   ✓ Persistence layer initialized")

	// Initialize Echo engine (depends on above components)
	if err := s.initializeEcho(ctx); err != nil {
		return fmt.Errorf("failed to initialize Echo: %w", err)
	}
	fmt.Println("   ✓ Echo cognitive engine initialized")

	s.initialized = true
	fmt.Println("🔷 LAMPSTACK Echo ready")

	return nil
}

// initializeAPI sets up the API provider
func (s *Stack) initializeAPI(ctx context.Context) error {
	api := NewMultiAPIProvider(s.config)
	s.api = api
	return nil
}

// initializeMemory sets up the memory store
func (s *Stack) initializeMemory(ctx context.Context) error {
	memory := NewUnifiedMemoryStore(s.config)
	s.memory = memory
	return nil
}

// initializePersistence sets up the persistence layer
func (s *Stack) initializePersistence(ctx context.Context) error {
	persistence := NewJSONPersistence(s.config)
	s.persistence = persistence
	return nil
}

// initializeEcho sets up the Echo cognitive engine
func (s *Stack) initializeEcho(ctx context.Context) error {
	echo := NewEchoIntegration(s.config, s.api, s.memory, s.persistence)
	s.echo = echo
	return nil
}

// Start begins stack operation
func (s *Stack) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		if err := s.Initialize(ctx); err != nil {
			return err
		}
	}

	fmt.Println("🔷 Starting LAMPSTACK Echo...")

	// Start Echo engine
	if s.echo != nil {
		if err := s.echo.Start(ctx); err != nil {
			return fmt.Errorf("failed to start Echo: %w", err)
		}
	}

	fmt.Println("🔷 LAMPSTACK Echo running")
	return nil
}

// Stop gracefully shuts down the stack
func (s *Stack) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Println("🔷 Stopping LAMPSTACK Echo...")

	// Stop Echo engine
	if s.echo != nil {
		if err := s.echo.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop Echo: %w", err)
		}
	}

	// Persist final state
	if s.persistence != nil {
		if err := s.persistence.SaveAll(ctx); err != nil {
			fmt.Printf("   ⚠️  Failed to save final state: %v\n", err)
		}
	}

	fmt.Println("🔷 LAMPSTACK Echo stopped")
	return nil
}

// Generate performs a cognitive generation with the full stack
func (s *Stack) Generate(ctx context.Context, request *GenerationRequest) (*GenerationResponse, error) {
	s.mu.RLock()
	if !s.initialized {
		s.mu.RUnlock()
		return nil, fmt.Errorf("stack not initialized")
	}
	s.mu.RUnlock()

	start := time.Now()
	s.metrics.incrementRequests()

	// Use Echo for cognitive generation if enabled
	if s.echo != nil && request.UseCognition {
		response, err := s.echo.Generate(ctx, request)
		if err != nil {
			s.metrics.incrementFailed()
			return nil, err
		}
		s.metrics.recordSuccess(time.Since(start))
		return response, nil
	}

	// Fall back to direct API generation
	response, err := s.api.Generate(ctx, request)
	if err != nil {
		s.metrics.incrementFailed()
		return nil, err
	}

	s.metrics.recordSuccess(time.Since(start))
	return response, nil
}

// Chat performs a conversational interaction
func (s *Stack) Chat(ctx context.Context, messages []Message) (*ChatResponse, error) {
	s.mu.RLock()
	if !s.initialized {
		s.mu.RUnlock()
		return nil, fmt.Errorf("stack not initialized")
	}
	s.mu.RUnlock()

	start := time.Now()
	s.metrics.incrementRequests()

	// Use Echo for cognitive chat
	if s.echo != nil {
		response, err := s.echo.Chat(ctx, messages)
		if err != nil {
			s.metrics.incrementFailed()
			return nil, err
		}
		s.metrics.recordSuccess(time.Since(start))
		return response, nil
	}

	// Fall back to direct API chat
	response, err := s.api.Chat(ctx, messages)
	if err != nil {
		s.metrics.incrementFailed()
		return nil, err
	}

	s.metrics.recordSuccess(time.Since(start))
	return response, nil
}

// Think generates a cognitive thought using the Echo engine
func (s *Stack) Think(ctx context.Context, prompt string) (*ThoughtResponse, error) {
	s.mu.RLock()
	if !s.initialized || s.echo == nil {
		s.mu.RUnlock()
		return nil, fmt.Errorf("Echo engine not available")
	}
	s.mu.RUnlock()

	start := time.Now()
	s.metrics.incrementRequests()

	response, err := s.echo.Think(ctx, prompt)
	if err != nil {
		s.metrics.incrementFailed()
		return nil, err
	}

	s.metrics.recordSuccess(time.Since(start))
	s.metrics.incrementThoughts()
	return response, nil
}

// Reflect generates a meta-cognitive reflection
func (s *Stack) Reflect(ctx context.Context, contextStr string) (*ReflectionResponse, error) {
	s.mu.RLock()
	if !s.initialized || s.echo == nil {
		s.mu.RUnlock()
		return nil, fmt.Errorf("Echo engine not available")
	}
	s.mu.RUnlock()

	start := time.Now()
	s.metrics.incrementRequests()

	response, err := s.echo.Reflect(ctx, contextStr)
	if err != nil {
		s.metrics.incrementFailed()
		return nil, err
	}

	s.metrics.recordSuccess(time.Since(start))
	s.metrics.incrementReflections()
	return response, nil
}

// StoreMemory stores information in the memory system
func (s *Stack) StoreMemory(ctx context.Context, memory *MemoryEntry) error {
	s.mu.RLock()
	if s.memory == nil {
		s.mu.RUnlock()
		return fmt.Errorf("memory store not available")
	}
	s.mu.RUnlock()

	if err := s.memory.Store(ctx, memory); err != nil {
		return err
	}

	s.metrics.incrementMemoriesStored()
	return nil
}

// RetrieveMemory retrieves relevant memories
func (s *Stack) RetrieveMemory(ctx context.Context, query string, limit int) ([]*MemoryEntry, error) {
	s.mu.RLock()
	if s.memory == nil {
		s.mu.RUnlock()
		return nil, fmt.Errorf("memory store not available")
	}
	s.mu.RUnlock()

	memories, err := s.memory.Retrieve(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	s.metrics.incrementMemoriesRetrieved()
	return memories, nil
}

// GetMetrics returns current stack metrics
func (s *Stack) GetMetrics() *StackMetrics {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	// Update uptime
	s.metrics.Uptime = time.Since(s.metrics.StartTime)

	// Create a copy
	return &StackMetrics{
		StartTime:            s.metrics.StartTime,
		Uptime:               s.metrics.Uptime,
		TotalRequests:        s.metrics.TotalRequests,
		SuccessfulRequests:   s.metrics.SuccessfulRequests,
		FailedRequests:       s.metrics.FailedRequests,
		ThoughtsGenerated:    s.metrics.ThoughtsGenerated,
		ReflectionsGenerated: s.metrics.ReflectionsGenerated,
		ConsolidationsCycles: s.metrics.ConsolidationsCycles,
		MemoriesStored:       s.metrics.MemoriesStored,
		MemoriesRetrieved:    s.metrics.MemoriesRetrieved,
		AverageLatencyMs:     s.metrics.AverageLatencyMs,
		TotalTokensUsed:      s.metrics.TotalTokensUsed,
	}
}

// GetCognitiveState returns the current cognitive state
func (s *Stack) GetCognitiveState() (*CognitiveState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.echo == nil {
		return nil, fmt.Errorf("Echo engine not available")
	}

	return s.echo.GetState()
}

// API returns the API provider for direct access
func (s *Stack) API() APIProvider {
	return s.api
}

// Memory returns the memory store for direct access
func (s *Stack) Memory() MemoryStore {
	return s.memory
}

// Persistence returns the persistence layer for direct access
func (s *Stack) Persistence() PersistenceLayer {
	return s.persistence
}

// Echo returns the Echo engine for direct access
func (s *Stack) Echo() EchoEngine {
	return s.echo
}

// Metrics helper methods
func (m *StackMetrics) incrementRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalRequests++
}

func (m *StackMetrics) incrementFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FailedRequests++
}

func (m *StackMetrics) recordSuccess(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SuccessfulRequests++

	// Rolling average
	if m.SuccessfulRequests == 1 {
		m.AverageLatencyMs = float64(latency.Milliseconds())
	} else {
		m.AverageLatencyMs = (m.AverageLatencyMs*float64(m.SuccessfulRequests-1) + float64(latency.Milliseconds())) / float64(m.SuccessfulRequests)
	}
}

func (m *StackMetrics) incrementThoughts() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ThoughtsGenerated++
}

func (m *StackMetrics) incrementReflections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ReflectionsGenerated++
}

func (m *StackMetrics) incrementMemoriesStored() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MemoriesStored++
}

func (m *StackMetrics) incrementMemoriesRetrieved() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MemoriesRetrieved++
}
