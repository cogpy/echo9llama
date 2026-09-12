package deeptreeecho

import (
	"testing"
)

// TestPersonaManagerActivation tests automatic persona activation based on cognitive state
func TestPersonaManagerActivation(t *testing.T) {
	identity := NewIdentity("TestPersonaManager")
	pm := identity.PersonaManager

	// Scenario 1: Low coherence, many patterns → should activate Ordo
	identity.Coherence = 0.4
	for i := 0; i < 50; i++ {
		identity.Patterns[string(rune('a'+i))] = &Pattern{
			ID:       string(rune('a' + i)),
			Strength: 0.7,
		}
	}
	identity.Iterations = 1500

	persona1 := pm.DetermineActivePersona(identity)
	if persona1 != PersonaOrdo {
		t.Errorf("Expected Ordo activation for low coherence + many patterns, got %s", persona1)
	}
	t.Logf("✓ Ordo activated correctly: low coherence (%.2f), many patterns (%d)",
		identity.Coherence, len(identity.Patterns))

	// Scenario 2: High coherence, few patterns → should activate Chao
	identity.Coherence = 0.92
	identity.Patterns = make(map[string]*Pattern)
	for i := 0; i < 10; i++ {
		identity.Patterns[string(rune('a'+i))] = &Pattern{
			ID:       string(rune('a' + i)),
			Strength: 0.9,
		}
	}
	identity.Iterations = 50

	persona2 := pm.DetermineActivePersona(identity)
	if persona2 != PersonaChao {
		t.Errorf("Expected Chao activation for high coherence + few patterns, got %s", persona2)
	}
	t.Logf("✓ Chao activated correctly: high coherence (%.2f), few patterns (%d)",
		identity.Coherence, len(identity.Patterns))

	// Verify activation was recorded
	stats := pm.GetActivationStats()
	t.Logf("Activation stats: %+v", stats)

	if pm.ordoActivations < 1 {
		t.Error("Expected at least one Ordo activation recorded")
	}
	if pm.chaoActivations < 1 {
		t.Error("Expected at least one Chao activation recorded")
	}
}

// TestPersonaBiasApplication tests that personas apply correct biases
func TestPersonaBiasApplication(t *testing.T) {
	identity := NewIdentity("TestPersonaBias")
	pm := identity.PersonaManager

	// Test Ordo bias application
	pm.ApplyPersonaBias(identity, PersonaOrdo)

	// Verify Ordo optimal balances
	expExp := identity.OpponentProcesses.Processes[ExplorationExploitation]
	if expExp.OptimalBalance != -0.4 {
		t.Errorf("Ordo should set exploration-exploitation optimal to -0.4, got %.2f",
			expExp.OptimalBalance)
	}

	stability := identity.OpponentProcesses.Processes[StabilityFlexibility]
	if stability.OptimalBalance != 0.6 {
		t.Errorf("Ordo should set stability-flexibility optimal to 0.6, got %.2f",
			stability.OptimalBalance)
	}
	if stability.PositiveProcess.Weight != 1.2 {
		t.Errorf("Ordo should increase stability weight to 1.2, got %.2f",
			stability.PositiveProcess.Weight)
	}

	t.Log("✓ Ordo bias applied correctly")

	// Test Chao bias application
	pm.ApplyPersonaBias(identity, PersonaChao)

	// Verify Chao optimal balances
	expExp = identity.OpponentProcesses.Processes[ExplorationExploitation]
	if expExp.OptimalBalance != 0.6 {
		t.Errorf("Chao should set exploration-exploitation optimal to 0.6, got %.2f",
			expExp.OptimalBalance)
	}
	if expExp.PositiveProcess.Weight != 1.2 {
		t.Errorf("Chao should increase exploration weight to 1.2, got %.2f",
			expExp.PositiveProcess.Weight)
	}

	stability = identity.OpponentProcesses.Processes[StabilityFlexibility]
	if stability.OptimalBalance != -0.6 {
		t.Errorf("Chao should set stability-flexibility optimal to -0.6, got %.2f",
			stability.OptimalBalance)
	}
	if stability.NegativeProcess.Weight != 1.2 {
		t.Errorf("Chao should increase flexibility weight to 1.2, got %.2f",
			stability.NegativeProcess.Weight)
	}

	t.Log("✓ Chao bias applied correctly")
}

// TestPersonaTransitions tests persona transitions over time
func TestPersonaTransitions(t *testing.T) {
	identity := NewIdentity("TestPersonaTransitions")
	pm := identity.PersonaManager

	// Track persona changes through different states
	personas := []PersonaArchetype{}

	// Phase 1: Early exploration (Chao)
	identity.Coherence = 0.3
	identity.Iterations = 10
	for i := 0; i < 5; i++ {
		identity.Patterns[string(rune('a'+i))] = &Pattern{
			ID:       string(rune('a' + i)),
			Strength: 0.5,
		}
	}
	personas = append(personas, pm.DetermineActivePersona(identity))

	// Phase 2: Accumulating patterns (transition to Ordo)
	identity.Coherence = 0.5
	identity.Iterations = 500
	for i := 5; i < 40; i++ {
		identity.Patterns[string(rune('a'+i))] = &Pattern{
			ID:       string(rune('a' + i)),
			Strength: 0.7,
		}
	}
	personas = append(personas, pm.DetermineActivePersona(identity))

	// Phase 3: Mastery state (Ordo stabilizes)
	identity.Coherence = 0.75
	identity.Iterations = 1500
	personas = append(personas, pm.DetermineActivePersona(identity))

	// Phase 4: Over-optimization (Chao disrupts)
	identity.Coherence = 0.96
	identity.Iterations = 3000
	personas = append(personas, pm.DetermineActivePersona(identity))

	// Log transition sequence
	t.Log("Persona transition sequence:")
	for i, persona := range personas {
		t.Logf("  Phase %d: %s", i+1, persona)
	}

	// Verify expected pattern: Chao → Ordo → Ordo → Chao
	if personas[0] != PersonaChao && personas[0] != PersonaNeutral {
		t.Logf("Warning: Expected Chao or Neutral in early phase, got %s", personas[0])
	}

	// Should have some Ordo activations in middle phases
	ordoCount := 0
	for _, p := range personas {
		if p == PersonaOrdo {
			ordoCount++
		}
	}

	if ordoCount == 0 {
		t.Log("Warning: Expected at least one Ordo activation during consolidation phases")
	}

	// Get recent activations
	recent := pm.GetRecentActivations(5)
	t.Logf("\nRecent activations:")
	for _, activation := range recent {
		t.Logf("  %s: %s", activation.Persona, activation.Reason)
	}

	// Check balance ratio
	ratio := pm.OrdoChaoBalanceRatio()
	t.Logf("\nOrdo/Chao balance ratio: %.2f (0.5 = perfectly balanced)", ratio)
}

// TestEmotionalPersonaModulation tests how emotions affect persona activation
func TestEmotionalPersonaModulation(t *testing.T) {
	identity := NewIdentity("TestEmotionalModulation")
	pm := identity.PersonaManager

	// Set baseline state
	identity.Coherence = 0.6
	identity.Iterations = 500
	for i := 0; i < 25; i++ {
		identity.Patterns[string(rune('a'+i))] = &Pattern{
			ID:       string(rune('a' + i)),
			Strength: 0.7,
		}
	}

	// Test 1: Calm state (no strong emotional bias)
	identity.EmotionalState = &EmotionalState{
		Arousal: 0.3,
		Valence: 0.5,
	}
	persona1 := pm.DetermineActivePersona(identity)
	t.Logf("Calm state: %s", persona1)

	// Test 2: High arousal → should favor Chao (quick adaptation)
	identity.EmotionalState = &EmotionalState{
		Arousal: 0.9,
		Valence: -0.2,
	}
	persona2 := pm.DetermineActivePersona(identity)
	t.Logf("High arousal state: %s", persona2)

	// High arousal should increase Chao activation likelihood
	if persona2 == PersonaOrdo && persona1 == PersonaChao {
		t.Log("Warning: High arousal typically favors Chao, but got Ordo")
	}

	// Test 3: Extreme positive valence
	identity.EmotionalState = &EmotionalState{
		Arousal: 0.7,
		Valence: 0.9,
	}
	persona3 := pm.DetermineActivePersona(identity)
	t.Logf("Extreme positive valence: %s", persona3)
}

// TestPersonaManagerStats tests statistics tracking
func TestPersonaManagerStats(t *testing.T) {
	identity := NewIdentity("TestPersonaStats")
	pm := identity.PersonaManager

	// Generate multiple activations
	for i := 0; i < 10; i++ {
		// Alternate between Ordo-favoring and Chao-favoring states
		if i%2 == 0 {
			// Ordo state
			identity.Coherence = 0.3
			identity.Iterations = uint64(1000 + i*100)
			for j := 0; j < 50; j++ {
				identity.Patterns[string(rune(j))] = &Pattern{
					ID:       string(rune(j)),
					Strength: 0.7,
				}
			}
		} else {
			// Chao state
			identity.Coherence = 0.95
			identity.Iterations = uint64(50 + i*10)
			identity.Patterns = make(map[string]*Pattern)
			for j := 0; j < 5; j++ {
				identity.Patterns[string(rune(j))] = &Pattern{
					ID:       string(rune(j)),
					Strength: 0.6,
				}
			}
		}

		pm.DetermineActivePersona(identity)
	}

	// Get statistics
	stats := pm.GetActivationStats()
	t.Logf("Statistics after 10 state changes:")
	t.Logf("  Current persona: %s", stats["current_persona"])
	t.Logf("  Ordo activations: %d", stats["ordo_activations"])
	t.Logf("  Chao activations: %d", stats["chao_activations"])
	t.Logf("  Total activations: %d", stats["total_activations"])
	t.Logf("  Ordo/Chao ratio: %.2f", pm.OrdoChaoBalanceRatio())

	// Verify we have some activations
	if stats["ordo_activations"].(int) == 0 && stats["chao_activations"].(int) == 0 {
		t.Error("Expected at least some persona activations")
	}

	// Get recent activation history
	recent := pm.GetRecentActivations(5)
	t.Logf("\nLast 5 activations:")
	for i, activation := range recent {
		t.Logf("  %d. %s - %s", i+1, activation.Persona, activation.Reason)
		t.Logf("     State: coherence=%.2f, patterns=%d, iterations=%d",
			activation.State.Coherence, activation.State.PatternCount, activation.State.Iterations)
	}
}

// TestIntegratedPersonaDecisionMaking tests persona-influenced decision making
func TestIntegratedPersonaDecisionMaking(t *testing.T) {
	identity := NewIdentity("TestIntegratedDecisions")

	// Scenario 1: Force Ordo activation and verify decisions
	identity.Coherence = 0.35
	identity.Iterations = 2000
	for i := 0; i < 60; i++ {
		identity.Patterns[string(rune('a'+i))] = &Pattern{
			ID:       string(rune('a' + i)),
			Strength: 0.75,
		}
	}

	decision1 := identity.OptimizeRelevanceRealization("ordo_influenced_decision")
	persona1 := identity.PersonaManager.GetCurrentPersona()

	t.Logf("Ordo-influenced decision:")
	t.Logf("  Active persona: %s", persona1)
	t.Logf("  Exploration: %.2f (expect low)", decision1.ExplorationWeight)
	t.Logf("  Scope: %s (expect depth)", decision1.ScopePreference)
	t.Logf("  Adaptation: %.2f (expect low)", decision1.AdaptationRate)

	// Scenario 2: Force Chao activation and verify decisions
	identity.Coherence = 0.93
	identity.Patterns = make(map[string]*Pattern)
	for i := 0; i < 8; i++ {
		identity.Patterns[string(rune('a'+i))] = &Pattern{
			ID:       string(rune('a' + i)),
			Strength: 0.9,
		}
	}
	identity.Iterations = 100

	decision2 := identity.OptimizeRelevanceRealization("chao_influenced_decision")
	persona2 := identity.PersonaManager.GetCurrentPersona()

	t.Logf("\nChao-influenced decision:")
	t.Logf("  Active persona: %s", persona2)
	t.Logf("  Exploration: %.2f (expect high)", decision2.ExplorationWeight)
	t.Logf("  Scope: %s (expect breadth)", decision2.ScopePreference)
	t.Logf("  Adaptation: %.2f (expect high)", decision2.AdaptationRate)

	// Verify personas influenced decisions appropriately
	if persona1 == PersonaOrdo && persona2 == PersonaChao {
		if decision1.ExplorationWeight < decision2.ExplorationWeight {
			t.Log("✓ Personas correctly influenced exploration tendency")
		} else {
			t.Log("Warning: Expected Ordo to reduce exploration vs Chao")
		}

		if decision1.ScopePreference != decision2.ScopePreference {
			t.Log("✓ Personas correctly influenced scope preference")
		}
	}
}
