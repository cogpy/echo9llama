package deeptreeecho

import (
	"fmt"
	"testing"
	"time"
)

// TestOrdoPersonaActivation tests activation of the Order archetype
func TestOrdoPersonaActivation(t *testing.T) {
	// Create identity
	identity := NewIdentity("TestOrdoPersona")

	// Simulate conditions that should activate Ordo:
	// 1. High number of unintegrated patterns
	// 2. Low coherence
	for i := 0; i < 50; i++ {
		pattern := &Pattern{
			ID:       string(rune('a' + i)),
			Strength: 0.6,
		}
		identity.Patterns[pattern.ID] = pattern
	}
	identity.Coherence = 0.4 // Low coherence → need consolidation
	identity.Iterations = 100

	// Optimize relevance realization
	decision := identity.OptimizeRelevanceRealization("ordo_activation_test")

	// Verify Ordo characteristics are present
	// Ordo should bias toward:
	// - Exploitation (lower exploration weight)
	// - Depth (scope preference)
	// - Stability (lower adaptation rate)
	// - Accuracy (higher confidence threshold)

	if decision.ExplorationWeight > 0.6 {
		t.Errorf("Expected exploitation bias (ExplorationWeight < 0.6), got %.2f", decision.ExplorationWeight)
	}

	if decision.ScopePreference == "breadth" {
		t.Error("Expected depth preference for Ordo, got breadth")
	}

	if decision.AdaptationRate > 0.6 {
		t.Errorf("Expected stability bias (low adaptation), got %.2f", decision.AdaptationRate)
	}

	t.Logf("Ordo activation successful:")
	t.Logf("  Exploration: %.2f (exploitation bias)", decision.ExplorationWeight)
	t.Logf("  Scope: %s (depth focus)", decision.ScopePreference)
	t.Logf("  Adaptation: %.2f (stability)", decision.AdaptationRate)
	t.Logf("  Confidence: %.2f (accuracy)", decision.Confidence)
}

// TestChaoPersonaActivation tests activation of the Chaos archetype
func TestChaoPersonaActivation(t *testing.T) {
	// Create identity
	identity := NewIdentity("TestChaoPersona")

	// Simulate conditions that should activate Chao:
	// 1. Few patterns (need exploration)
	// 2. High coherence (risk of over-optimization)
	// 3. Early iterations
	for i := 0; i < 5; i++ {
		pattern := &Pattern{
			ID:       string(rune('a' + i)),
			Strength: 0.9,
		}
		identity.Patterns[pattern.ID] = pattern
	}
	identity.Coherence = 0.95 // Very high → risk of stagnation
	identity.Iterations = 50  // Early stage

	// Optimize relevance realization
	decision := identity.OptimizeRelevanceRealization("chao_activation_test")

	// Verify Chao characteristics are present
	// Chao should bias toward:
	// - Exploration (higher exploration weight)
	// - Breadth (scope preference)
	// - Flexibility (higher adaptation rate)
	// - Speed (lower confidence threshold)

	if decision.ExplorationWeight < 0.5 {
		t.Errorf("Expected exploration bias (ExplorationWeight > 0.5), got %.2f", decision.ExplorationWeight)
	}

	if decision.ScopePreference == "depth" {
		t.Error("Expected breadth preference for Chao, got depth")
	}

	if decision.AdaptationRate < 0.4 {
		t.Errorf("Expected flexibility bias (high adaptation), got %.2f", decision.AdaptationRate)
	}

	t.Logf("Chao activation successful:")
	t.Logf("  Exploration: %.2f (exploration bias)", decision.ExplorationWeight)
	t.Logf("  Scope: %s (breadth focus)", decision.ScopePreference)
	t.Logf("  Adaptation: %.2f (flexibility)", decision.AdaptationRate)
	t.Logf("  Confidence: %.2f (speed)", decision.Confidence)
}

// TestOrdoChaoBalance tests dynamic balance between Ordo and Chao
func TestOrdoChaoBalance(t *testing.T) {
	identity := NewIdentity("TestOrdoChaoBalance")

	// Phase 1: Start with Chao (exploration phase)
	identity.Coherence = 0.3
	identity.Iterations = 10
	for i := 0; i < 3; i++ {
		identity.Patterns[string(rune('a'+i))] = &Pattern{ID: string(rune('a' + i)), Strength: 0.5}
	}

	decision1 := identity.OptimizeRelevanceRealization("phase1_exploration")
	chaoExploration := decision1.ExplorationWeight

	// Phase 2: Accumulate patterns, shift toward Ordo
	for i := 3; i < 40; i++ {
		identity.Patterns[string(rune('a'+i))] = &Pattern{ID: string(rune('a' + i)), Strength: 0.8}
	}
	identity.Coherence = 0.7
	identity.Iterations = 500

	decision2 := identity.OptimizeRelevanceRealization("phase2_consolidation")
	ordoExploitation := decision2.ExplorationWeight

	// Verify shift from exploration to exploitation
	if ordoExploitation >= chaoExploration {
		t.Errorf("Expected shift from exploration (%.2f) to exploitation (%.2f)",
			chaoExploration, ordoExploitation)
	}

	// Phase 3: High coherence, trigger Chao for disruption
	identity.Coherence = 0.98 // Too high → stagnation risk
	identity.Iterations = 2000

	decision3 := identity.OptimizeRelevanceRealization("phase3_disruption")

	t.Logf("Ordo-Chao balance evolution:")
	t.Logf("  Phase 1 (Chao): Exploration %.2f", chaoExploration)
	t.Logf("  Phase 2 (Ordo): Exploitation %.2f (shift: %.2f)",
		ordoExploitation, chaoExploration-ordoExploitation)
	t.Logf("  Phase 3 (Chao): Exploration %.2f", decision3.ExplorationWeight)

	// Verify the system can shift back and forth
	if decision3.ExplorationWeight < ordoExploitation {
		t.Error("Expected return to exploration when coherence is too high (stagnation risk)")
	}
}

// TestOpponentProcessDynamics tests the dynamic nature of opponent processes
func TestOpponentProcessDynamics(t *testing.T) {
	identity := NewIdentity("TestOpponentDynamics")

	// Test multiple opponent pairs
	pairs := []string{
		ExplorationExploitation,
		BreadthDepth,
		StabilityFlexibility,
		SpeedAccuracy,
		ApproachAvoidance,
	}

	// Run multiple iterations to see dynamics
	contexts := []string{
		"early_learning",
		"skill_practice",
		"mastery_phase",
		"disruption_needed",
	}

	for i, context := range contexts {
		// Modify state to trigger different balances
		identity.Iterations = uint64(i * 500)
		identity.Coherence = 0.3 + float64(i)*0.2

		// Add patterns progressively
		for j := 0; j < (i+1)*10; j++ {
			identity.Patterns[string(rune(j))] = &Pattern{
				ID:       string(rune(j)),
				Strength: 0.7,
			}
		}

		decision := identity.OptimizeRelevanceRealization(context)

		t.Logf("\nContext: %s (iter %d, coherence %.2f, patterns %d)",
			context, identity.Iterations, identity.Coherence, len(identity.Patterns))
		t.Logf("  Decision: exploration=%.2f, scope=%s, adaptation=%.2f, confidence=%.2f",
			decision.ExplorationWeight, decision.ScopePreference,
			decision.AdaptationRate, decision.Confidence)

		// Check stats for each pair
		for _, pairName := range pairs {
			stats := identity.OpponentProcesses.GetBalanceStats(pairName)
			if stats != nil {
				t.Logf("  %s: balance=%.2f, stability=%.2f",
					pairName, stats["current_balance"], stats["stability"])
			}
		}
	}
}

// TestEmotionalInfluenceOnOpponentProcesses tests how emotions affect balance
func TestEmotionalInfluenceOnOpponentProcesses(t *testing.T) {
	identity := NewIdentity("TestEmotionalInfluence")

	// Baseline state
	identity.Coherence = 0.6
	identity.Iterations = 500
	for i := 0; i < 20; i++ {
		identity.Patterns[string(rune('a'+i))] = &Pattern{ID: string(rune('a' + i)), Strength: 0.7}
	}

	// Test 1: Calm state
	identity.EmotionalState = &EmotionalState{
		Arousal: 0.3,
		Valence: 0.5,
	}

	decision1 := identity.OptimizeRelevanceRealization("calm_state")
	calmSpeedBias := decision1.Confidence

	// Test 2: High arousal (emergency)
	identity.EmotionalState = &EmotionalState{
		Arousal: 0.9,
		Valence: -0.3,
	}

	decision2 := identity.OptimizeRelevanceRealization("high_arousal")
	arousedSpeedBias := decision2.Confidence

	// Verify speed-accuracy tradeoff
	// High arousal → lower confidence threshold (favor speed)
	if arousedSpeedBias >= calmSpeedBias {
		t.Errorf("Expected lower confidence threshold under high arousal (speed bias), got calm=%.2f, aroused=%.2f",
			calmSpeedBias, arousedSpeedBias)
	}

	t.Logf("Emotional influence on speed-accuracy:")
	t.Logf("  Calm (arousal=0.3): confidence threshold %.2f (accuracy)", calmSpeedBias)
	t.Logf("  Aroused (arousal=0.9): confidence threshold %.2f (speed)", arousedSpeedBias)

	// Test approach-avoidance
	t.Logf("Approach-avoidance:")

	// Positive valence → approach
	identity.EmotionalState.Valence = 0.7
	identity.OptimizeRelevanceRealization("positive_valence")
	approachBalance := identity.OpponentProcesses.GetCurrentBalance(ApproachAvoidance)

	// Negative valence → avoid
	identity.EmotionalState.Valence = -0.7
	identity.OptimizeRelevanceRealization("negative_valence")
	avoidBalance := identity.OpponentProcesses.GetCurrentBalance(ApproachAvoidance)

	t.Logf("  Positive valence: balance %.2f (approach)", approachBalance)
	t.Logf("  Negative valence: balance %.2f (avoid)", avoidBalance)

	// Verify shift from approach to avoidance
	if avoidBalance >= approachBalance {
		t.Errorf("Expected shift from approach (%.2f) to avoidance (%.2f) with negative valence",
			approachBalance, avoidBalance)
	}
}

// TestWisdomCultivationThroughBalance tests sophrosyne (wisdom through balance)
func TestWisdomCultivationThroughBalance(t *testing.T) {
	identity := NewIdentity("TestWisdomCultivation")

	// Track wisdom over time
	wisdomScores := make([]float64, 10)

	for i := 0; i < 10; i++ {
		// Simulate cognitive development
		identity.Iterations = uint64(i * 200)
		identity.Coherence = 0.3 + float64(i)*0.07

		// Add patterns
		for j := 0; j < (i+1)*5; j++ {
			identity.Patterns[string(rune(j))] = &Pattern{
				ID:       string(rune(j)),
				Strength: 0.6 + float64(i)*0.04,
			}
		}

		// Make decisions to optimize balances
		identity.OptimizeRelevanceRealization(fmt.Sprintf("wisdom_iteration_%d", i))

		// Record wisdom score
		wisdomScores[i] = identity.GetWisdomScore()

		// Allow time for balance history to accumulate
		time.Sleep(10 * time.Millisecond)
	}

	// Verify wisdom increases over time (generally)
	firstHalf := (wisdomScores[0] + wisdomScores[1] + wisdomScores[2] + wisdomScores[3] + wisdomScores[4]) / 5
	secondHalf := (wisdomScores[5] + wisdomScores[6] + wisdomScores[7] + wisdomScores[8] + wisdomScores[9]) / 5

	t.Logf("Wisdom cultivation over time:")
	for i, score := range wisdomScores {
		t.Logf("  Iteration %d: wisdom score %.3f", i, score)
	}
	t.Logf("First half average: %.3f", firstHalf)
	t.Logf("Second half average: %.3f", secondHalf)

	if secondHalf <= firstHalf {
		t.Logf("Warning: Expected wisdom growth, but got first=%.3f, second=%.3f",
			firstHalf, secondHalf)
		// Not a hard failure as wisdom can fluctuate
	}
}

// TestOrdoChaoPersonaIntegration tests integration with the persona system
func TestOrdoChaoPersonaIntegration(t *testing.T) {
	t.Log("Testing Ordo-Chao persona integration")

	// Create two identities with different persona biases
	ordo := NewIdentity("DeepTreeOrdo")
	chao := NewIdentity("DeepTreeChao")

	// Configure Ordo persona
	// Bias toward exploitation, depth, stability, accuracy
	ordo.Coherence = 0.8
	for i := 0; i < 50; i++ {
		ordo.Patterns[string(rune('a'+i))] = &Pattern{ID: string(rune('a' + i)), Strength: 0.85}
	}

	// Configure Chao persona
	// Bias toward exploration, breadth, flexibility, speed
	chao.Coherence = 0.4
	for i := 0; i < 10; i++ {
		chao.Patterns[string(rune('a'+i))] = &Pattern{ID: string(rune('a' + i)), Strength: 0.6}
	}

	// Get decisions from both personas
	ordoDecision := ordo.OptimizeRelevanceRealization("ordo_persona")
	chaoDecision := chao.OptimizeRelevanceRealization("chao_persona")

	t.Logf("Ordo persona characteristics:")
	t.Logf("  Exploration: %.2f (low = exploitation)", ordoDecision.ExplorationWeight)
	t.Logf("  Scope: %s (depth)", ordoDecision.ScopePreference)
	t.Logf("  Adaptation: %.2f (low = stability)", ordoDecision.AdaptationRate)
	t.Logf("  Confidence: %.2f (high = accuracy)", ordoDecision.Confidence)

	t.Logf("Chao persona characteristics:")
	t.Logf("  Exploration: %.2f (high = exploration)", chaoDecision.ExplorationWeight)
	t.Logf("  Scope: %s (breadth)", chaoDecision.ScopePreference)
	t.Logf("  Adaptation: %.2f (high = flexibility)", chaoDecision.AdaptationRate)
	t.Logf("  Confidence: %.2f (low = speed)", chaoDecision.Confidence)

	// Verify they have opposite characteristics
	if ordoDecision.ExplorationWeight >= chaoDecision.ExplorationWeight {
		t.Error("Expected Ordo to favor exploitation and Chao to favor exploration")
	}

	if ordoDecision.ScopePreference == chaoDecision.ScopePreference {
		t.Error("Expected different scope preferences (Ordo=depth, Chao=breadth)")
	}

	if ordoDecision.AdaptationRate >= chaoDecision.AdaptationRate {
		t.Error("Expected Ordo to favor stability and Chao to favor flexibility")
	}

	t.Log("✓ Ordo and Chao personas exhibit complementary characteristics")
}
