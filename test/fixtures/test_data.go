package fixtures

import (
	"time"
)

// TestMemoryNode represents a test memory node
type TestMemoryNode struct {
	ID        string
	Type      string
	Content   string
	Embedding []float64
	Metadata  map[string]interface{}
	CreatedAt time.Time
}

// TestMemoryEdge represents a test memory edge
type TestMemoryEdge struct {
	ID       string
	Type     string
	Source   string
	Target   string
	Weight   float64
	Metadata map[string]interface{}
}

// SampleMemoryNodes returns sample memory nodes for testing
func SampleMemoryNodes() []*TestMemoryNode {
	return []*TestMemoryNode{
		{
			ID:        "node-episodic-1",
			Type:      "episodic",
			Content:   "I learned about Go concurrency patterns today",
			Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5},
			Metadata: map[string]interface{}{
				"source":    "learning",
				"timestamp": time.Now().Add(-time.Hour * 24),
			},
			CreatedAt: time.Now().Add(-time.Hour * 24),
		},
		{
			ID:        "node-episodic-2",
			Type:      "episodic",
			Content:   "Successfully implemented the actor model",
			Embedding: []float64{0.2, 0.3, 0.4, 0.5, 0.6},
			Metadata: map[string]interface{}{
				"source":    "achievement",
				"timestamp": time.Now().Add(-time.Hour * 12),
			},
			CreatedAt: time.Now().Add(-time.Hour * 12),
		},
		{
			ID:        "node-semantic-1",
			Type:      "semantic",
			Content:   "Actors communicate through message passing",
			Embedding: []float64{0.3, 0.4, 0.5, 0.6, 0.7},
			Metadata: map[string]interface{}{
				"domain":   "computer_science",
				"category": "concurrency",
			},
			CreatedAt: time.Now().Add(-time.Hour * 48),
		},
		{
			ID:        "node-procedural-1",
			Type:      "procedural",
			Content:   "To create an actor: 1. Define message types 2. Implement Receive method 3. Spawn actor",
			Embedding: []float64{0.4, 0.5, 0.6, 0.7, 0.8},
			Metadata: map[string]interface{}{
				"skill":      "actor_creation",
				"difficulty": "intermediate",
			},
			CreatedAt: time.Now().Add(-time.Hour * 36),
		},
		{
			ID:        "node-working-1",
			Type:      "working",
			Content:   "Currently processing cognitive loop step 5",
			Embedding: []float64{0.5, 0.6, 0.7, 0.8, 0.9},
			Metadata: map[string]interface{}{
				"active":    true,
				"step":      5,
				"timestamp": time.Now(),
			},
			CreatedAt: time.Now(),
		},
	}
}

// SampleMemoryEdges returns sample memory edges for testing
func SampleMemoryEdges() []*TestMemoryEdge {
	return []*TestMemoryEdge{
		{
			ID:     "edge-1",
			Type:   "association",
			Source: "node-episodic-1",
			Target: "node-semantic-1",
			Weight: 0.8,
			Metadata: map[string]interface{}{
				"strength": "strong",
			},
		},
		{
			ID:     "edge-2",
			Type:   "causal",
			Source: "node-episodic-1",
			Target: "node-episodic-2",
			Weight: 0.9,
			Metadata: map[string]interface{}{
				"relationship": "led_to",
			},
		},
		{
			ID:     "edge-3",
			Type:   "similarity",
			Source: "node-semantic-1",
			Target: "node-procedural-1",
			Weight: 0.7,
			Metadata: map[string]interface{}{
				"similarity_type": "conceptual",
			},
		},
	}
}

// TestCognitiveStep represents a test cognitive step
type TestCognitiveStep struct {
	StepNumber int
	Phase      string
	Engine     string
	Input      interface{}
	Expected   interface{}
}

// SampleCognitiveSteps returns sample cognitive steps for testing
func SampleCognitiveSteps() []*TestCognitiveStep {
	return []*TestCognitiveStep{
		{StepNumber: 0, Phase: "relevance", Engine: "relevance", Input: "initial_focus", Expected: "relevance_realized"},
		{StepNumber: 1, Phase: "affordance", Engine: "affordance", Input: "past_context_1", Expected: "affordance_1"},
		{StepNumber: 2, Phase: "affordance", Engine: "affordance", Input: "past_context_2", Expected: "affordance_2"},
		{StepNumber: 3, Phase: "affordance", Engine: "affordance", Input: "past_context_3", Expected: "affordance_3"},
		{StepNumber: 4, Phase: "affordance", Engine: "affordance", Input: "past_context_4", Expected: "affordance_4"},
		{StepNumber: 5, Phase: "affordance", Engine: "affordance", Input: "past_context_5", Expected: "affordance_5"},
		{StepNumber: 6, Phase: "relevance", Engine: "relevance", Input: "reorient_focus", Expected: "relevance_realized"},
		{StepNumber: 7, Phase: "salience", Engine: "salience", Input: "future_option_1", Expected: "scenario_1"},
		{StepNumber: 8, Phase: "salience", Engine: "salience", Input: "future_option_2", Expected: "scenario_2"},
		{StepNumber: 9, Phase: "salience", Engine: "salience", Input: "future_option_3", Expected: "scenario_3"},
		{StepNumber: 10, Phase: "salience", Engine: "salience", Input: "future_option_4", Expected: "scenario_4"},
		{StepNumber: 11, Phase: "salience", Engine: "salience", Input: "future_option_5", Expected: "scenario_5"},
	}
}

// TestLLMPrompt represents a test LLM prompt
type TestLLMPrompt struct {
	Prompt      string
	Mode        string
	MaxTokens   int
	Temperature float64
	Expected    string
}

// SampleLLMPrompts returns sample LLM prompts for testing
func SampleLLMPrompts() []*TestLLMPrompt {
	return []*TestLLMPrompt{
		{
			Prompt:      "What is consciousness?",
			Mode:        "thought",
			MaxTokens:   100,
			Temperature: 0.7,
			Expected:    "Consciousness is",
		},
		{
			Prompt:      "Learn about quantum computing",
			Mode:        "goal",
			MaxTokens:   50,
			Temperature: 0.5,
			Expected:    "Goal created",
		},
		{
			Prompt:      "Hello, how are you?",
			Mode:        "conversation",
			MaxTokens:   50,
			Temperature: 0.8,
			Expected:    "Hello",
		},
	}
}

// TestTriad represents a test triad configuration
type TestTriad struct {
	Name  string
	Steps []int
}

// SampleTriads returns sample triad configurations for testing
func SampleTriads() []*TestTriad {
	return []*TestTriad{
		{Name: "pivotal_relevance_1", Steps: []int{1, 5, 9}},
		{Name: "affordance_action", Steps: []int{2, 6, 10}},
		{Name: "salience_simulation", Steps: []int{3, 7, 11}},
		{Name: "meta_reflection", Steps: []int{4, 8, 12}},
	}
}

// TestConfig represents a test configuration
type TestConfig struct {
	ServerURL       string
	Timeout         time.Duration
	RetryCount      int
	RetryDelay      time.Duration
	EnableTelemetry bool
}

// DefaultTestConfig returns default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		ServerURL:       "http://localhost:8081",
		Timeout:         time.Second * 30,
		RetryCount:      3,
		RetryDelay:      time.Second,
		EnableTelemetry: false,
	}
}
