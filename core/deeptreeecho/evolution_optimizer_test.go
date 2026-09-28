package deeptreeecho

import (
	"context"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// MockLLMProvider for testing
type MockLLMProvider struct{}

func (m *MockLLMProvider) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	return "Mock response for testing", nil
}

func (m *MockLLMProvider) StreamGenerate(ctx context.Context, prompt string, opts llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: "Mock stream response", Done: true}
	close(ch)
	return ch, nil
}

func (m *MockLLMProvider) Name() string {
	return "MockProvider"
}

func (m *MockLLMProvider) Available() bool {
	return true
}

func (m *MockLLMProvider) MaxTokens() int {
	return 4096
}

func TestNewEvolutionOptimizer(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultEvolutionConfig()

	eo := NewEvolutionOptimizer(provider, config)

	if eo == nil {
		t.Fatal("Expected non-nil EvolutionOptimizer")
	}

	if eo.consciousness == nil {
		t.Error("Expected non-nil consciousness")
	}

	if eo.scheduler == nil {
		t.Error("Expected non-nil scheduler")
	}

	if eo.dreamIntegration == nil {
		t.Error("Expected non-nil dreamIntegration")
	}

	if len(eo.geneticTraits) != 5 {
		t.Errorf("Expected 5 genetic traits, got %d", len(eo.geneticTraits))
	}

	if eo.fitnessLandscape == nil {
		t.Error("Expected non-nil fitness landscape")
	}
}

func TestGeneticTraits(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultEvolutionConfig()

	eo := NewEvolutionOptimizer(provider, config)

	traits := eo.GetGeneticTraits()

	expectedTraits := []string{
		"curiosity_drive",
		"wisdom_accumulation",
		"pattern_recognition",
		"coherence_maintenance",
		"adaptive_learning",
	}

	for _, name := range expectedTraits {
		trait, ok := traits[name]
		if !ok {
			t.Errorf("Missing expected trait: %s", name)
			continue
		}

		if trait.Value < 0 || trait.Value > 1 {
			t.Errorf("Trait %s has invalid value: %f", name, trait.Value)
		}

		if trait.Expression < 0 || trait.Expression > 1 {
			t.Errorf("Trait %s has invalid expression: %f", name, trait.Expression)
		}
	}
}

func TestEvolutionStatus(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultEvolutionConfig()

	eo := NewEvolutionOptimizer(provider, config)

	status := eo.GetEvolutionStatus()

	if status["generation"].(uint64) != 0 {
		t.Errorf("Expected generation 0, got %d", status["generation"])
	}

	if status["fitness_score"].(float64) != 0.5 {
		t.Errorf("Expected initial fitness 0.5, got %f", status["fitness_score"])
	}

	if status["running"].(bool) {
		t.Error("Expected running to be false initially")
	}
}

func TestEvolutionHistory(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultEvolutionConfig()

	eo := NewEvolutionOptimizer(provider, config)

	history := eo.GetEvolutionHistory()

	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d entries", len(history))
	}
}

func TestEchobeatsSchedulerGoals(t *testing.T) {
	provider := &MockLLMProvider{}
	scheduler := NewEchobeatsScheduler(provider)

	// Test adding goals
	goalID := scheduler.AddGoal("Test goal 1", 0.8)
	if goalID == "" {
		t.Error("Expected non-empty goal ID")
	}

	scheduler.AddGoal("Test goal 2", 0.5)
	scheduler.AddGoal("Test goal 3", 0.9)

	queue := scheduler.GetGoalQueue()
	if len(queue) != 3 {
		t.Errorf("Expected 3 goals, got %d", len(queue))
	}

	// Goals should be sorted by priority (highest first)
	if queue[0].Priority != 0.9 {
		t.Errorf("Expected first goal priority 0.9, got %f", queue[0].Priority)
	}

	// Test active goal retrieval
	activeGoal := scheduler.GetActiveGoal()
	if activeGoal == nil {
		t.Fatal("Expected non-nil active goal")
	}

	if activeGoal.Priority != 0.9 {
		t.Errorf("Expected active goal priority 0.9, got %f", activeGoal.Priority)
	}
}

func TestEchobeatsSchedulerTriads(t *testing.T) {
	provider := &MockLLMProvider{}
	scheduler := NewEchobeatsScheduler(provider)

	triads := scheduler.GetTriadStates()

	if len(triads) != 4 {
		t.Errorf("Expected 4 triads, got %d", len(triads))
	}

	expectedNames := []string{
		"RelevanceRealization",
		"AffordanceInteraction",
		"SalienceSimulation",
		"MetaCognitiveReflection",
	}

	for i, name := range expectedNames {
		if triads[i].Name != name {
			t.Errorf("Expected triad %d name %s, got %s", i, name, triads[i].Name)
		}
	}

	// Test triad update
	scheduler.UpdateTriadState(0, 0.8, 0.9)
	updatedTriads := scheduler.GetTriadStates()

	if updatedTriads[0].Activation != 0.8 {
		t.Errorf("Expected triad 0 activation 0.8, got %f", updatedTriads[0].Activation)
	}

	if updatedTriads[0].Coherence != 0.9 {
		t.Errorf("Expected triad 0 coherence 0.9, got %f", updatedTriads[0].Coherence)
	}
}

func TestEchobeatsPhaseRotation(t *testing.T) {
	provider := &MockLLMProvider{}
	scheduler := NewEchobeatsScheduler(provider)

	initialTriads := scheduler.GetTriadStates()
	initialActivation := initialTriads[0].Activation

	scheduler.RotatePhase()

	updatedTriads := scheduler.GetTriadStates()
	newActivation := updatedTriads[0].Activation

	// Activation should change after rotation
	if newActivation == initialActivation {
		t.Error("Expected activation to change after phase rotation")
	}
}

func TestEchoDreamKnowledgeIntegration(t *testing.T) {
	provider := &MockLLMProvider{}
	edi := NewEchoDreamKnowledgeIntegration(provider)

	if edi == nil {
		t.Fatal("Expected non-nil EchoDreamKnowledgeIntegration")
	}

	// Test adding memory
	memID := edi.AddMemory("Test memory content", 0.8, []string{"test", "memory"})
	if memID == "" {
		t.Error("Expected non-empty memory ID")
	}

	metrics := edi.GetMetrics()
	if metrics["total_memories"].(int) != 1 {
		t.Errorf("Expected 1 memory, got %d", metrics["total_memories"])
	}

	// Test wisdom depth
	depth := edi.GetWisdomDepth()
	if depth != 0.0 {
		t.Errorf("Expected initial wisdom depth 0.0, got %f", depth)
	}

	// Test dream phase
	phase := edi.GetDreamPhase()
	if phase != PhaseWaking {
		t.Errorf("Expected initial phase Waking, got %s", phase.String())
	}
}

func TestEchodreamSemanticNetwork(t *testing.T) {
	provider := &MockLLMProvider{}
	edi := NewEchoDreamKnowledgeIntegration(provider)

	network := edi.GetSemanticNetwork()
	if len(network) != 0 {
		t.Errorf("Expected empty semantic network, got %d nodes", len(network))
	}

	concepts := edi.GetEmergentConcepts()
	if len(concepts) != 0 {
		t.Errorf("Expected no emergent concepts, got %d", len(concepts))
	}
}

func TestStreamOfConsciousnessMetrics(t *testing.T) {
	provider := &MockLLMProvider{}
	soc := NewStreamOfConsciousness(provider)

	metrics := soc.GetMetrics()

	if metrics["total_thoughts"].(uint64) != 0 {
		t.Errorf("Expected 0 total thoughts, got %d", metrics["total_thoughts"])
	}

	if metrics["current_focus"].(string) != "exploring existence" {
		t.Errorf("Expected initial focus 'exploring existence', got %s", metrics["current_focus"])
	}

	if metrics["awake"].(bool) != true {
		t.Error("Expected consciousness to be awake initially")
	}
}

func TestStreamOfConsciousnessGoals(t *testing.T) {
	provider := &MockLLMProvider{}
	soc := NewStreamOfConsciousness(provider)

	soc.AddGoal("Learn about evolution")
	soc.AddGoal("Understand patterns")

	metrics := soc.GetMetrics()
	if metrics["active_goals"].(int) != 2 {
		t.Errorf("Expected 2 active goals, got %d", metrics["active_goals"])
	}
}

func TestStreamOfConsciousnessKnowledgeGaps(t *testing.T) {
	provider := &MockLLMProvider{}
	soc := NewStreamOfConsciousness(provider)

	soc.AddKnowledgeGap("quantum computing", 0.9)
	soc.AddKnowledgeGap("neural networks", 0.7)

	metrics := soc.GetMetrics()
	if metrics["knowledge_gaps"].(int) != 2 {
		t.Errorf("Expected 2 knowledge gaps, got %d", metrics["knowledge_gaps"])
	}
}

func TestDefaultEvolutionConfig(t *testing.T) {
	config := DefaultEvolutionConfig()

	if config.InitialAdaptationRate != 0.1 {
		t.Errorf("Expected InitialAdaptationRate 0.1, got %f", config.InitialAdaptationRate)
	}

	if config.MutationProbability != 0.15 {
		t.Errorf("Expected MutationProbability 0.15, got %f", config.MutationProbability)
	}

	if config.SelectionPressure != 0.7 {
		t.Errorf("Expected SelectionPressure 0.7, got %f", config.SelectionPressure)
	}

	if config.EvolutionCycle != 30*time.Second {
		t.Errorf("Expected EvolutionCycle 30s, got %v", config.EvolutionCycle)
	}
}

func TestCycleMetrics(t *testing.T) {
	provider := &MockLLMProvider{}
	scheduler := NewEchobeatsScheduler(provider)

	metrics := scheduler.GetCompleteCycleMetrics()

	if metrics.CycleNumber != 0 {
		t.Errorf("Expected cycle number 0, got %d", metrics.CycleNumber)
	}

	if metrics.StepsCompleted != 12 {
		t.Errorf("Expected 12 steps completed, got %d", metrics.StepsCompleted)
	}

	if len(metrics.EnginePerformance) != 3 {
		t.Errorf("Expected 3 engine performance values, got %d", len(metrics.EnginePerformance))
	}

	if len(metrics.TriadCoherence) != 4 {
		t.Errorf("Expected 4 triad coherence values, got %d", len(metrics.TriadCoherence))
	}
}

func TestInjectExternalStimulus(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultEvolutionConfig()
	eo := NewEvolutionOptimizer(provider, config)

	// Test high importance stimulus
	eo.InjectExternalStimulus("Critical discovery about consciousness", 0.9)

	metrics := eo.GetConsciousness().GetMetrics()

	// Should be added as knowledge gap, interest, and goal
	if metrics["knowledge_gaps"].(int) != 1 {
		t.Errorf("Expected 1 knowledge gap, got %d", metrics["knowledge_gaps"])
	}

	if metrics["interests"].(int) != 1 {
		t.Errorf("Expected 1 interest, got %d", metrics["interests"])
	}

	if metrics["active_goals"].(int) != 1 {
		t.Errorf("Expected 1 active goal, got %d", metrics["active_goals"])
	}
}
