package mocks

import (
	"context"
	"sync"
	"time"
)

// MockLLMProvider provides a mock implementation of LLM providers for testing
type MockLLMProvider struct {
	mu           sync.RWMutex
	name         string
	responses    map[string]string
	defaultResp  string
	callCount    int
	latency      time.Duration
	shouldFail   bool
	failureError error
}

// NewMockLLMProvider creates a new mock LLM provider
func NewMockLLMProvider(name string) *MockLLMProvider {
	return &MockLLMProvider{
		name:        name,
		responses:   make(map[string]string),
		defaultResp: "Mock response",
		latency:     time.Millisecond * 10,
	}
}

// Name returns the provider name
func (m *MockLLMProvider) Name() string {
	return m.name
}

// SetResponse sets a specific response for a prompt
func (m *MockLLMProvider) SetResponse(prompt, response string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[prompt] = response
}

// SetDefaultResponse sets the default response
func (m *MockLLMProvider) SetDefaultResponse(response string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultResp = response
}

// SetLatency sets the simulated latency
func (m *MockLLMProvider) SetLatency(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.latency = latency
}

// SetFailure configures the mock to fail
func (m *MockLLMProvider) SetFailure(shouldFail bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
	m.failureError = err
}

// Generate generates a mock response
func (m *MockLLMProvider) Generate(ctx context.Context, prompt string, opts interface{}) (string, error) {
	m.mu.Lock()
	m.callCount++
	latency := m.latency
	shouldFail := m.shouldFail
	failureError := m.failureError
	m.mu.Unlock()

	// Simulate latency
	select {
	case <-time.After(latency):
	case <-ctx.Done():
		return "", ctx.Err()
	}

	if shouldFail {
		return "", failureError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if response, ok := m.responses[prompt]; ok {
		return response, nil
	}
	return m.defaultResp, nil
}

// GetCallCount returns the number of calls made
func (m *MockLLMProvider) GetCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCount
}

// Reset resets the mock state
func (m *MockLLMProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount = 0
	m.responses = make(map[string]string)
	m.shouldFail = false
	m.failureError = nil
}

// MockMemoryStore provides a mock implementation of memory storage
type MockMemoryStore struct {
	mu    sync.RWMutex
	nodes map[string]interface{}
	edges map[string]interface{}
}

// NewMockMemoryStore creates a new mock memory store
func NewMockMemoryStore() *MockMemoryStore {
	return &MockMemoryStore{
		nodes: make(map[string]interface{}),
		edges: make(map[string]interface{}),
	}
}

// AddNode adds a node to the mock store
func (m *MockMemoryStore) AddNode(id string, data interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nodes[id] = data
	return nil
}

// GetNode retrieves a node from the mock store
func (m *MockMemoryStore) GetNode(id string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.nodes[id]
	return data, ok
}

// AddEdge adds an edge to the mock store
func (m *MockMemoryStore) AddEdge(id string, data interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.edges[id] = data
	return nil
}

// GetEdge retrieves an edge from the mock store
func (m *MockMemoryStore) GetEdge(id string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.edges[id]
	return data, ok
}

// NodeCount returns the number of nodes
func (m *MockMemoryStore) NodeCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.nodes)
}

// EdgeCount returns the number of edges
func (m *MockMemoryStore) EdgeCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.edges)
}

// Clear clears all data
func (m *MockMemoryStore) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nodes = make(map[string]interface{})
	m.edges = make(map[string]interface{})
}

// MockCognitiveEngine provides a mock cognitive engine for testing
type MockCognitiveEngine struct {
	mu          sync.RWMutex
	name        string
	stepResults map[int]interface{}
	coherence   float64
}

// NewMockCognitiveEngine creates a new mock cognitive engine
func NewMockCognitiveEngine(name string) *MockCognitiveEngine {
	return &MockCognitiveEngine{
		name:        name,
		stepResults: make(map[int]interface{}),
		coherence:   1.0,
	}
}

// Name returns the engine name
func (m *MockCognitiveEngine) Name() string {
	return m.name
}

// ProcessStep processes a cognitive step
func (m *MockCognitiveEngine) ProcessStep(step int) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := map[string]interface{}{
		"step":      step,
		"engine":    m.name,
		"timestamp": time.Now(),
		"success":   true,
	}

	m.stepResults[step] = result
	return result, nil
}

// GetStepResult retrieves a step result
func (m *MockCognitiveEngine) GetStepResult(step int) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result, ok := m.stepResults[step]
	return result, ok
}

// SetCoherence sets the coherence score
func (m *MockCognitiveEngine) SetCoherence(coherence float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.coherence = coherence
}

// GetCoherence returns the coherence score
func (m *MockCognitiveEngine) GetCoherence() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.coherence
}

// Reset resets the mock state
func (m *MockCognitiveEngine) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stepResults = make(map[int]interface{})
	m.coherence = 1.0
}
