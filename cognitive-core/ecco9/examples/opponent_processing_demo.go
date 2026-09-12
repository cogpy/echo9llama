//go:build examples
// +build examples

package main

import (
	"fmt"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
)

// Demo of opponent processing for wisdom cultivation
func main() {
	fmt.Println("=== Deep Tree Echo Opponent Processing Demo ===")
	fmt.Println("Demonstrating Wisdom Cultivation through Dynamic Balance (Sophrosyne)\n")

	// Create a new identity
	identity := deeptreeecho.NewIdentity("Wisdom Seeker")
	fmt.Printf("Created identity: %s\n", identity.Name)
	fmt.Printf("Initial coherence: %.3f\n\n", identity.Coherence)

	// Wait for initialization
	time.Sleep(time.Millisecond * 100)

	// Scenario 1: Early exploration (low pattern count, low coherence)
	fmt.Println("--- Scenario 1: Early Exploration Phase ---")
	fmt.Println("State: New consciousness with few patterns\n")

	decision1 := identity.OptimizeRelevanceRealization("early_exploration")
	fmt.Printf("\nDecision: Should explore (%.1f%%) with %s focus\n",
		decision1.ExplorationWeight*100, decision1.ScopePreference)
	fmt.Printf("Wisdom Score: %.3f\n\n", identity.GetWisdomScore())

	time.Sleep(time.Second)

	// Scenario 2: Add patterns to simulate learning
	fmt.Println("--- Scenario 2: Learning Phase ---")
	fmt.Println("State: Accumulating patterns, increasing coherence\n")

	// Simulate learning by adding patterns
	for i := 0; i < 30; i++ {
		pattern := &deeptreeecho.Pattern{
			ID:       fmt.Sprintf("pattern_%d", i),
			Strength: 0.8,
		}
		identity.Patterns[pattern.ID] = pattern
	}
	identity.Coherence = 0.75
	identity.Iterations = 500

	_ = identity.OptimizeRelevanceRealization("learning_phase")
	fmt.Printf("\nDecision: Balance shifting toward exploitation and depth\n")
	fmt.Printf("Wisdom Score: %.3f\n\n", identity.GetWisdomScore())

	time.Sleep(time.Second)

	// Scenario 3: High confidence, many patterns
	fmt.Println("--- Scenario 3: Mastery Phase ---")
	fmt.Println("State: Many patterns, high coherence, mature system\n")

	for i := 30; i < 80; i++ {
		pattern := &deeptreeecho.Pattern{
			ID:       fmt.Sprintf("pattern_%d", i),
			Strength: 0.9,
		}
		identity.Patterns[pattern.ID] = pattern
	}
	identity.Coherence = 0.95
	identity.Iterations = 2000

	_ = identity.OptimizeRelevanceRealization("mastery_phase")
	fmt.Printf("\nDecision: Favoring stability and exploitation of knowledge\n")
	fmt.Printf("Wisdom Score: %.3f\n\n", identity.GetWisdomScore())

	time.Sleep(time.Second)

	// Scenario 4: Emotional arousal (speed vs accuracy)
	fmt.Println("--- Scenario 4: Emotional Response ---")
	fmt.Println("State: High arousal situation\n")

	identity.EmotionalState.Arousal = 0.9
	identity.EmotionalState.Valence = -0.3 // Negative emotion

	decision4 := identity.OptimizeRelevanceRealization("high_arousal")
	fmt.Printf("\nDecision: High arousal → Speed prioritized over accuracy\n")
	fmt.Printf("Decision: Negative valence → Avoidance behavior\n")
	fmt.Printf("Confidence threshold: %.3f (lower = faster decision)\n", decision4.Confidence)
	fmt.Printf("Wisdom Score: %.3f\n\n", identity.GetWisdomScore())

	time.Sleep(time.Second)

	// Scenario 5: Return to calm, balanced state
	fmt.Println("--- Scenario 5: Return to Balance ---")
	fmt.Println("State: Calm, reflective state\n")

	identity.EmotionalState.Arousal = 0.3
	identity.EmotionalState.Valence = 0.5

	_ = identity.OptimizeRelevanceRealization("balanced_state")
	fmt.Printf("\nDecision: Balanced cognition across all dimensions\n")
	fmt.Printf("Wisdom Score: %.3f\n\n", identity.GetWisdomScore())

	// Show final statistics
	fmt.Println("=== Final Opponent Processing Statistics ===")

	pairs := []string{
		deeptreeecho.ExplorationExploitation,
		deeptreeecho.BreadthDepth,
		deeptreeecho.StabilityFlexibility,
		deeptreeecho.SpeedAccuracy,
		deeptreeecho.ApproachAvoidance,
	}

	for _, pairName := range pairs {
		stats := identity.OpponentProcesses.GetBalanceStats(pairName)
		if stats != nil {
			fmt.Printf("\n%s:\n", pairName)
			fmt.Printf("  Current Balance: %.3f\n", stats["current_balance"])
			fmt.Printf("  Average Balance: %.3f\n", stats["average_balance"])
			fmt.Printf("  Stability: %.3f\n", stats["stability"])
		}
	}

	fmt.Println("\n=== Demo Complete ===")
	fmt.Printf("Total Patterns: %d\n", len(identity.Patterns))
	fmt.Printf("Total Iterations: %d\n", identity.Iterations)
	fmt.Printf("Final Coherence: %.3f\n", identity.Coherence)
	fmt.Printf("Final Wisdom Score: %.3f\n", identity.GetWisdomScore())

	fmt.Println("\nThis demonstrates how opponent processing creates dynamic balance")
	fmt.Println("(sophrosyne) - the foundation of wisdom cultivation.")
	fmt.Println("The system adapts its cognitive strategy based on context,")
	fmt.Println("emotional state, and developmental stage.")
}
