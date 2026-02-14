package live2d

import (
	"testing"
	"time"
)

func TestEcho9AvatarOrchestrator_StateAggregation(t *testing.T) {
	orchestrator := NewEcho9AvatarOrchestrator()

	// Set up test states
	reservoirState := ReservoirVisualizationState{
		SpectralRadius:  0.95,
		InputScaling:    0.3,
		LeakRate:        0.2,
		Stability:       0.9,
		Fatigue:         0.1,
		Persona:         "contemplative-scholar",
		ActivationLevel: 0.7,
	}

	echoBeatsPhase := EchoBeatPhase{
		Step:           7,
		Phase:          "reorientation",
		PhaseProgress:  0.5,
		CycleStartTime: time.Now(),
		TimeInStep:     500 * time.Millisecond,
	}

	wisdomMetrics := WisdomMetrics{
		KnowledgeDepth:       0.7,
		KnowledgeBreadth:     0.6,
		IntegrationLevel:     0.8,
		PracticalApplication: 0.65,
		ReflectiveInsight:    0.75,
		EthicalConsideration: 0.7,
		TemporalPerspective:  0.6,
		Overall:              0.7,
	}

	aarState := AARState{
		Awareness:    0.8,
		Attention:    0.7,
		Reflection:   0.75,
		Coherence:    0.85,
		Relevance:    0.8,
		Optimization: 0.75,
	}

	emotionalDynamics := EmotionalDynamics{
		Joy:          0.6,
		Curiosity:    0.7,
		Confidence:   0.65,
		Anticipation: 0.5,
	}

	// Update orchestrator
	orchestrator.UpdateReservoirState(reservoirState)
	orchestrator.UpdateEchoBeatsPhase(echoBeatsPhase)
	orchestrator.UpdateWisdomMetrics(wisdomMetrics)
	orchestrator.UpdateAARState(aarState)
	orchestrator.UpdateEmotionalDynamics(emotionalDynamics)

	// Trigger aggregation
	orchestrator.aggregateAndUpdate()

	// Get unified state
	unified := orchestrator.GetCurrentState()

	// Verify state aggregation
	if unified.EchoBeatPosition.Phase != "reorientation" {
		t.Errorf("Expected phase 'reorientation', got '%s'", unified.EchoBeatPosition.Phase)
	}

	if unified.Cognitive.ProcessingMode != "contemplative" {
		t.Errorf("Expected mode 'contemplative', got '%s'", unified.Cognitive.ProcessingMode)
	}

	if unified.Emotional.Valence <= 0 {
		t.Error("Expected positive valence from joy")
	}

	if unified.Confidence <= 0 {
		t.Error("Expected positive confidence score")
	}
}

func TestEcho9AvatarOrchestrator_ArchetypeDetermination(t *testing.T) {
	orchestrator := NewEcho9AvatarOrchestrator()

	tests := []struct {
		name              string
		arousal           float64
		coherence         float64
		stability         float64
		expectedArchetype CognitiveArchetype
	}{
		{
			name:              "High chaos",
			arousal:           0.9,
			coherence:         0.4,
			stability:         0.3,
			expectedArchetype: ArchetypeChaos,
		},
		{
			name:              "High order",
			arousal:           0.2,
			coherence:         0.9,
			stability:         0.95,
			expectedArchetype: ArchetypeOrder,
		},
		{
			name:              "Balanced",
			arousal:           0.5,
			coherence:         0.7,
			stability:         0.7,
			expectedArchetype: ArchetypeBalance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := UnifiedAvatarState{
				Emotional: EmotionalState{Arousal: tt.arousal},
				Cognitive: CognitiveState{Coherence: tt.coherence},
				ReservoirDynamics: ReservoirVisualizationState{
					Stability: tt.stability,
				},
			}

			archetype := orchestrator.determineArchetype(state)

			if archetype != tt.expectedArchetype {
				t.Errorf("Expected archetype %v, got %v", tt.expectedArchetype, archetype)
			}
		})
	}
}

func TestOntogeneticProfile_Evolution(t *testing.T) {
	profile := NewOntogeneticProfile()

	// Initial stage should be embryonic
	if profile.Stage != StageEmbryonic {
		t.Errorf("Expected initial stage embryonic, got %v", profile.Stage)
	}

	wisdom := WisdomMetrics{
		Overall: 0.6,
	}

	// Evolve with low interactions - should stay embryonic
	for i := 0; i < 50; i++ {
		profile.Update(wisdom)
	}

	if profile.Stage != StageEmbryonic {
		t.Error("Should still be embryonic with low interactions")
	}

	// Evolve to juvenile
	for i := 0; i < 100; i++ {
		profile.Update(wisdom)
	}

	if profile.Stage != StageJuvenile {
		t.Errorf("Expected juvenile stage, got %v", profile.Stage)
	}

	// Check baseline evolution
	if profile.BaselineEmotion.Confidence <= 0.5 {
		t.Error("Baseline confidence should have increased with wisdom")
	}
}

func TestAvatarLearningMemory_ExpressionRecording(t *testing.T) {
	memory := NewAvatarLearningMemory()

	expr := EmotionalState{
		Valence:    0.7,
		Arousal:    0.6,
		Confidence: 0.8,
	}

	context := "conversation"
	duration := 5 * time.Second
	success := 0.85

	// Record expression
	memory.RecordExpression(expr, context, duration, success)

	// Verify record exists
	key := expr.Hash()
	record := memory.ExpressionHistory[key]

	if record == nil {
		t.Fatal("Expression record not created")
	}

	if record.TimesUsed != 1 {
		t.Errorf("Expected 1 use, got %d", record.TimesUsed)
	}

	if record.SuccessScore != success {
		t.Errorf("Expected success score %.2f, got %.2f", success, record.SuccessScore)
	}

	// Suggest expression for same context
	suggested, found := memory.SuggestExpression(context)

	if !found {
		t.Error("Should find suggestion for recorded context")
	}

	if suggested.Confidence != expr.Confidence {
		t.Error("Suggested expression should match recorded expression")
	}
}

func TestEcho9OptimizedMapper_ParameterCalculation(t *testing.T) {
	mapper := NewEcho9OptimizedMapper()

	state := UnifiedAvatarState{
		Emotional: EmotionalState{
			Valence:    0.6,
			Arousal:    0.5,
			Dominance:  0.6,
			Curiosity:  0.7,
			Confidence: 0.65,
		},
		Cognitive: CognitiveState{
			Awareness:      0.7,
			Attention:      0.8,
			CognitiveLoad:  0.4,
			Coherence:      0.8,
			EnergyLevel:    0.75,
			ProcessingMode: "contemplative",
		},
		EchoBeatPosition: EchoBeatPhase{
			Step:  5,
			Phase: "affordance",
		},
		ActiveArchetype: ArchetypeBalance,
	}

	// Calculate parameters
	params := mapper.MapCombinedState(state)

	if len(params) == 0 {
		t.Fatal("No parameters calculated")
	}

	// Verify eye parameters exist
	hasEyeOpen := false
	hasEyeSmile := false
	for _, p := range params {
		if p.ID == StandardParameterNames.EyeOpenLeft {
			hasEyeOpen = true
			// Verify value is reasonable
			if p.Value < 0.0 || p.Value > 1.0 {
				t.Errorf("Eye open value out of range: %.2f", p.Value)
			}
		}
		if p.ID == StandardParameterNames.EyeSmileLeft {
			hasEyeSmile = true
		}
	}

	if !hasEyeOpen {
		t.Error("Eye open parameter not calculated")
	}
	if !hasEyeSmile {
		t.Error("Eye smile parameter not calculated")
	}
}

func TestEcho9OptimizedMapper_DifferentialUpdate(t *testing.T) {
	mapper := NewEcho9OptimizedMapper()

	// Initial state
	state1 := UnifiedAvatarState{
		Emotional: EmotionalState{Valence: 0.5},
		Cognitive: CognitiveState{ProcessingMode: "contemplative"},
	}

	params1 := mapper.MapCombinedState(state1)
	initialCount := len(params1)

	// Slightly modified state
	state2 := state1
	state2.Emotional.Valence = 0.505 // Very small change

	params2 := mapper.MapDifferential(state2)

	// Should return fewer parameters (only significantly changed ones)
	if len(params2) >= initialCount {
		t.Errorf("Differential update should return fewer params: got %d, initial %d",
			len(params2), initialCount)
	}
}

func TestEcho9OptimizedMapper_Caching(t *testing.T) {
	mapper := NewEcho9OptimizedMapper()

	state := UnifiedAvatarState{
		Emotional: EmotionalState{Valence: 0.5, Arousal: 0.5},
		Cognitive: CognitiveState{ProcessingMode: "contemplative"},
	}

	// First calculation
	start1 := time.Now()
	params1 := mapper.MapCombinedState(state)
	duration1 := time.Since(start1)

	// Second calculation with same state (should hit cache)
	start2 := time.Now()
	params2 := mapper.MapCombinedState(state)
	duration2 := time.Since(start2)

	// Verify same results
	if len(params1) != len(params2) {
		t.Error("Cached result should have same length")
	}

	// Cached call should be faster (note: may not always be true in tests)
	t.Logf("First call: %v, Second call (cached): %v", duration1, duration2)
}

func TestCircularBuffer_Operations(t *testing.T) {
	buffer := NewCircularBuffer(5)

	// Add items
	for i := 0; i < 10; i++ {
		buffer.Add(i)
	}

	// Size should be max
	if buffer.Len() != 5 {
		t.Errorf("Expected size 5, got %d", buffer.Len())
	}

	// Get last 3
	last3 := buffer.GetLast(3)
	if len(last3) != 3 {
		t.Errorf("Expected 3 items, got %d", len(last3))
	}

	// Should be [7, 8, 9] (last 3 of 0-9)
	expected := []int{7, 8, 9}
	for i, item := range last3 {
		if item.(int) != expected[i] {
			t.Errorf("Expected %d at position %d, got %d", expected[i], i, item.(int))
		}
	}
}

func TestStateDiffCalculator_ChangeDetection(t *testing.T) {
	calc := NewStateDiffCalculator()

	state1 := UnifiedAvatarState{
		Emotional: EmotionalState{Valence: 0.5, Arousal: 0.5},
		Cognitive: CognitiveState{Awareness: 0.7, ProcessingMode: "contemplative"},
	}

	state2 := state1
	state2.Emotional.Valence = 0.8 // Significant change

	diff := calc.Calculate(state1, state2)

	// Should detect valence change
	hasValenceChange := false
	for _, dim := range diff.ChangedDimensions {
		if dim == "emotional.valence" {
			hasValenceChange = true
			break
		}
	}

	if !hasValenceChange {
		t.Error("Should detect valence change")
	}

	if diff.EmotionalChange <= 0 {
		t.Error("Emotional change magnitude should be positive")
	}
}

func TestArchetypalModulator_ChaosModulation(t *testing.T) {
	modulator := NewArchetypalModulator()

	baseParams := []ModelParameter{
		{ID: "Test1", Value: 0.5, Min: 0.0, Max: 1.0},
		{ID: "Test2", Value: 0.5, Min: 0.0, Max: 1.0},
	}

	// Modulate with chaos
	chaosParams := modulator.Modulate(baseParams, ArchetypeChaos)

	// Values should be different (due to added noise)
	allSame := true
	for i, p := range chaosParams {
		if p.Value != baseParams[i].Value {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Chaos modulation should change values")
	}

	// Values should still be in range
	for _, p := range chaosParams {
		if p.Value < p.Min || p.Value > p.Max {
			t.Errorf("Chaos modulation produced out-of-range value: %.2f", p.Value)
		}
	}
}

func TestArchetypalModulator_OrderModulation(t *testing.T) {
	modulator := NewArchetypalModulator()

	baseParams := []ModelParameter{
		{ID: "Test1", Value: 0.8, Min: 0.0, Max: 1.0},
		{ID: "Test2", Value: 0.2, Min: 0.0, Max: 1.0},
	}

	// Modulate with order (should smooth toward neutral)
	orderParams := modulator.Modulate(baseParams, ArchetypeOrder)

	// Values should move toward 0.5 (neutral)
	if orderParams[0].Value >= baseParams[0].Value {
		t.Error("Order modulation should reduce extreme high values")
	}
	if orderParams[1].Value <= baseParams[1].Value {
		t.Error("Order modulation should increase extreme low values")
	}
}

func BenchmarkEcho9OptimizedMapper_MapCombinedState(b *testing.B) {
	mapper := NewEcho9OptimizedMapper()

	state := UnifiedAvatarState{
		Emotional: EmotionalState{
			Valence:    0.6,
			Arousal:    0.5,
			Confidence: 0.65,
		},
		Cognitive: CognitiveState{
			Awareness:      0.7,
			Attention:      0.8,
			ProcessingMode: "contemplative",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mapper.MapCombinedState(state)
	}
}

func BenchmarkEcho9AvatarOrchestrator_AggregateState(b *testing.B) {
	orchestrator := NewEcho9AvatarOrchestrator()

	// Set up test states
	orchestrator.UpdateReservoirState(ReservoirVisualizationState{
		SpectralRadius: 0.95,
		Stability:      0.9,
	})
	orchestrator.UpdateWisdomMetrics(WisdomMetrics{Overall: 0.7})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orchestrator.aggregateAndUpdate()
	}
}

func BenchmarkStateDiffCalculator_Calculate(b *testing.B) {
	calc := NewStateDiffCalculator()

	state1 := UnifiedAvatarState{
		Emotional: EmotionalState{Valence: 0.5},
		Cognitive: CognitiveState{Awareness: 0.7},
	}

	state2 := UnifiedAvatarState{
		Emotional: EmotionalState{Valence: 0.6},
		Cognitive: CognitiveState{Awareness: 0.8},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calc.Calculate(state1, state2)
	}
}
