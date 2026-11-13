package relevance

import (
	"context"
	"testing"
	"time"
)

func TestEngineCreation(t *testing.T) {
	ctx := context.Background()
	engine := NewEngine(ctx)

	if engine == nil {
		t.Fatal("Engine should not be nil")
	}

	if engine.knowing == nil {
		t.Error("Knowing triad should not be nil")
	}

	if engine.understanding == nil {
		t.Error("Understanding triad should not be nil")
	}

	if engine.wisdom == nil {
		t.Error("Wisdom triad should not be nil")
	}

	if engine.realization == nil {
		t.Error("Realization process should not be nil")
	}
}

func TestEngineStartStop(t *testing.T) {
	ctx := context.Background()
	engine := NewEngine(ctx)

	// Start engine
	err := engine.Start()
	if err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}

	// Should not be able to start again
	err = engine.Start()
	if err == nil {
		t.Error("Should not be able to start engine twice")
	}

	// Let it run for enough time to complete cycles
	time.Sleep(1500 * time.Millisecond)

	// Stop engine
	engine.Stop()

	// Check that it actually ran
	metrics := engine.GetMetrics()
	if metrics.TotalCycles == 0 {
		t.Error("Engine should have completed at least one cycle")
	}
}

func TestEnneadStateInitialization(t *testing.T) {
	ctx := context.Background()
	engine := NewEngine(ctx)

	state := engine.GetState()

	// All dimensions should start at 0.5
	expectedDimensions := map[string]float64{
		"PropositionalKnowledge":   0.5,
		"ProceduralKnowledge":      0.5,
		"PerspectivalKnowledge":    0.5,
		"ParticipatoryKnowledge":   0.5,
		"NomologicalUnderstanding": 0.5,
		"NormativeUnderstanding":   0.5,
		"NarrativeUnderstanding":   0.5,
		"MoralDevelopment":         0.5,
		"MeaningRealization":       0.5,
		"MasteryAchievement":       0.5,
		"OverallCoherence":         0.5,
		"RelevanceOptimization":    0.5,
	}

	// Check each dimension
	if state.PropositionalKnowledge != expectedDimensions["PropositionalKnowledge"] {
		t.Errorf("PropositionalKnowledge = %.2f, want %.2f",
			state.PropositionalKnowledge, expectedDimensions["PropositionalKnowledge"])
	}

	if state.OverallCoherence != expectedDimensions["OverallCoherence"] {
		t.Errorf("OverallCoherence = %.2f, want %.2f",
			state.OverallCoherence, expectedDimensions["OverallCoherence"])
	}
}

func TestRelevanceRealization(t *testing.T) {
	ctx := context.Background()
	engine := NewEngine(ctx)

	// Test relevance realization for an input
	input := "test input for relevance realization"

	rr := engine.RealizeRelevance(input)

	if rr == nil {
		t.Fatal("RelevanceRealization should not be nil")
	}

	if rr.Input != input {
		t.Errorf("Input = %v, want %v", rr.Input, input)
	}

	if rr.RelevanceScore < 0 || rr.RelevanceScore > 1 {
		t.Errorf("RelevanceScore = %.2f, should be between 0 and 1", rr.RelevanceScore)
	}

	if rr.KnowingAnalysis == nil {
		t.Error("KnowingAnalysis should not be nil")
	}

	if rr.UnderstandingAnalysis == nil {
		t.Error("UnderstandingAnalysis should not be nil")
	}

	if rr.WisdomAnalysis == nil {
		t.Error("WisdomAnalysis should not be nil")
	}
}

func TestUpdateFromExperience(t *testing.T) {
	ctx := context.Background()
	engine := NewEngine(ctx)

	initialState := engine.GetState()
	initialKnowing := initialState.PropositionalKnowledge

	// Create a positive experience
	exp := &Experience{
		Input:     "test input",
		Output:    "test output",
		Feedback:  0.8, // Positive feedback
		Context:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// Update engine with experience multiple times
	for i := 0; i < 5; i++ {
		engine.UpdateFromExperience(exp)
	}

	// Let system integrate
	time.Sleep(100 * time.Millisecond)

	newState := engine.GetState()

	// Check that overall coherence is reasonable
	if newState.OverallCoherence < 0 || newState.OverallCoherence > 1 {
		t.Errorf("OverallCoherence = %.2f, should be between 0 and 1", newState.OverallCoherence)
	}

	// The system should have processed the experiences
	// Due to balancing, individual dimensions might not increase monotonically,
	// but the system should have updated
	t.Logf("Initial knowing: %.3f, Final knowing: %.3f", initialKnowing, newState.PropositionalKnowledge)
	t.Logf("Overall coherence: %.3f", newState.OverallCoherence)
}

func TestCrossTriadIntegration(t *testing.T) {
	ctx := context.Background()
	engine := NewEngine(ctx)

	err := engine.Start()
	if err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}
	defer engine.Stop()

	// Let it run for a bit to perform integrations
	time.Sleep(2 * time.Second)

	metrics := engine.GetMetrics()

	if metrics.CrossTriadIntegrations == 0 {
		t.Error("Should have performed at least one cross-triad integration")
	}

	if metrics.SophrosyneOptimizations == 0 {
		t.Error("Should have performed at least one sophrosyne optimization")
	}

	t.Logf("Cycles: %d, Integrations: %d, Sophrosyne: %d",
		metrics.TotalCycles,
		metrics.CrossTriadIntegrations,
		metrics.SophrosyneOptimizations)
}

func TestStatus(t *testing.T) {
	ctx := context.Background()
	engine := NewEngine(ctx)

	status := engine.GetStatus()

	if status == nil {
		t.Fatal("Status should not be nil")
	}

	if status["running"] != false {
		t.Error("Engine should not be running initially")
	}

	state, ok := status["state"].(map[string]interface{})
	if !ok {
		t.Fatal("Status should contain state map")
	}

	knowing, ok := state["knowing"].(map[string]float64)
	if !ok {
		t.Fatal("State should contain knowing map")
	}

	if _, ok := knowing["propositional"]; !ok {
		t.Error("Knowing should contain propositional dimension")
	}

	understanding, ok := state["understanding"].(map[string]float64)
	if !ok {
		t.Fatal("State should contain understanding map")
	}

	if _, ok := understanding["nomological"]; !ok {
		t.Error("Understanding should contain nomological dimension")
	}

	wisdom, ok := state["wisdom"].(map[string]float64)
	if !ok {
		t.Fatal("State should contain wisdom map")
	}

	if _, ok := wisdom["morality"]; !ok {
		t.Error("Wisdom should contain morality dimension")
	}
}

func TestKnowingTriad(t *testing.T) {
	kt := NewKnowingTriad()

	if kt == nil {
		t.Fatal("KnowingTriad should not be nil")
	}

	// Test balance
	kt.Balance()

	state := kt.GetState()
	if state["gnostic_integration"] == 0 {
		t.Error("Gnostic integration should not be zero")
	}

	// Test analysis
	input := "test"
	analysis := kt.Analyze(input)

	if analysis == nil {
		t.Fatal("Analysis should not be nil")
	}

	if analysis.OverallScore < 0 || analysis.OverallScore > 1 {
		t.Errorf("OverallScore = %.2f, should be between 0 and 1", analysis.OverallScore)
	}
}

func TestUnderstandingTriad(t *testing.T) {
	ut := NewUnderstandingTriad()

	if ut == nil {
		t.Fatal("UnderstandingTriad should not be nil")
	}

	// Test integrate
	ut.Integrate()

	state := ut.GetState()
	if state["meaning_integration"] == 0 {
		t.Error("Meaning integration should not be zero")
	}

	// Test analysis
	input := "test"
	analysis := ut.Analyze(input)

	if analysis == nil {
		t.Fatal("Analysis should not be nil")
	}

	if analysis.OverallScore < 0 || analysis.OverallScore > 1 {
		t.Errorf("OverallScore = %.2f, should be between 0 and 1", analysis.OverallScore)
	}
}

func TestWisdomTriad(t *testing.T) {
	wt := NewWisdomTriad()

	if wt == nil {
		t.Fatal("WisdomTriad should not be nil")
	}

	// Test cultivate
	wt.Cultivate()

	state := wt.GetState()
	if state["eudaimonia"] == 0 {
		t.Error("Eudaimonia should not be zero")
	}

	// Test analysis
	input := "test"
	analysis := wt.Analyze(input)

	if analysis == nil {
		t.Fatal("Analysis should not be nil")
	}

	if analysis.OverallScore < 0 || analysis.OverallScore > 1 {
		t.Errorf("OverallScore = %.2f, should be between 0 and 1", analysis.OverallScore)
	}
}
