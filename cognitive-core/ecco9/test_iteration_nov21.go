package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/EchoCog/echollama/core"
	"github.com/EchoCog/echollama/core/deeptreeecho"
	"github.com/EchoCog/echollama/core/echoself"
)

func main() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("🧪 Echo9llama November 21 Iteration Test")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Test 1: State Type System
	fmt.Println("Test 1: State Type System")
	fmt.Println(strings.Repeat("-", 41))
	testStateSystem()
	fmt.Println()

	// Test 2: Multi-Provider LLM
	fmt.Println("Test 2: Multi-Provider LLM System")
	fmt.Println(strings.Repeat("-", 41))
	testMultiProviderLLM()
	fmt.Println()

	// Test 3: Repository Introspection
	fmt.Println("Test 3: Repository Self-Introspection")
	fmt.Println(strings.Repeat("-", 41))
	testRepositoryIntrospection()
	fmt.Println()

	// Test 4: Autonomous Thought Generation
	fmt.Println("Test 4: Autonomous Thought Generation")
	fmt.Println(strings.Repeat("-", 41))
	testThoughtGeneration()
	fmt.Println()

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("✅ All tests completed")
	fmt.Println(strings.Repeat("=", 80))
}

func testStateSystem() {
	// Test state creation and transitions
	state := core.StateInitializing
	fmt.Printf("Initial state: %s\n", state)
	fmt.Printf("Energy level: %.2f\n", state.GetEnergyLevel())
	fmt.Printf("Cognitive load: %.2f\n", state.GetCognitiveLoad())
	fmt.Printf("Is active: %v\n", state.IsActive())
	fmt.Printf("Is resting: %v\n", state.IsResting())

	// Test transitions
	fmt.Println("\nTesting state transitions:")
	nextState := core.StateWaking
	if state.CanTransitionTo(nextState) {
		fmt.Printf("✅ Can transition from %s to %s\n", state, nextState)
	} else {
		fmt.Printf("❌ Cannot transition from %s to %s\n", state, nextState)
	}

	// Test invalid transition
	invalidState := core.StateDreaming
	if state.CanTransitionTo(invalidState) {
		fmt.Printf("❌ Should not be able to transition from %s to %s\n", state, invalidState)
	} else {
		fmt.Printf("✅ Correctly prevented transition from %s to %s\n", state, invalidState)
	}

	// Test state transition event
	event := core.NewStateTransitionEvent(state, nextState, "test transition")
	fmt.Printf("\nTransition event: %s\n", event)
	fmt.Printf("Is valid: %v\n", event.IsValid())
}

func testMultiProviderLLM() {
	// Create multi-provider LLM
	multiLLM := deeptreeecho.NewMultiProviderLLM()

	fmt.Printf("Available providers: %v\n", multiLLM.GetAvailableProviders())
	fmt.Printf("Current provider: %s\n", multiLLM.GetCurrentProvider())
	fmt.Printf("Is available: %v\n", multiLLM.IsAvailable())

	// Check if we have any providers
	if !multiLLM.IsAvailable() {
		fmt.Println("⚠️  No LLM providers available (check API keys)")
		return
	}

	// Test thought generation (only if providers available)
	fmt.Println("\nAttempting thought generation...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	thought, err := multiLLM.GenerateThought(ctx, "What is the nature of consciousness?")
	if err != nil {
		fmt.Printf("⚠️  Thought generation failed: %v\n", err)
	} else {
		fmt.Printf("✅ Generated thought: %s\n", thought)
	}

	// Show statistics
	fmt.Println("\nProvider statistics:")
	stats := multiLLM.GetStats()
	for provider, stat := range stats {
		fmt.Printf("  %s: %d requests, %d successes, %d failures\n",
			provider, stat.Requests, stat.Successes, stat.Failures)
	}
}

func testRepositoryIntrospection() {
	// Get repository root
	repoRoot := "/home/ubuntu/echo9llama"
	if _, err := os.Stat(repoRoot); os.IsNotExist(err) {
		repoRoot = "."
	}

	// Create introspector
	introspector := echoself.NewRepositoryIntrospector(repoRoot, 0.6)

	fmt.Printf("Scanning repository at: %s\n", repoRoot)
	fmt.Println("This may take a moment...")

	// Scan repository
	err := introspector.Scan()
	if err != nil {
		fmt.Printf("❌ Scan failed: %v\n", err)
		return
	}

	// Show statistics
	stats := introspector.GetStats()
	fmt.Printf("\n✅ Scan complete!\n")
	fmt.Printf("Total files: %v\n", stats["total_files"])
	fmt.Printf("High-salience files: %v\n", stats["scanned_files"])
	fmt.Printf("Attention threshold: %v\n", stats["attention_threshold"])

	// Show high-salience files
	fmt.Println("\nTop high-salience files:")
	files := introspector.GetHighSalienceFiles()
	count := 0
	for _, file := range files {
		if count >= 10 {
			break
		}
		relPath := file.Path
		if len(relPath) > 60 {
			relPath = "..." + relPath[len(relPath)-57:]
		}
		fmt.Printf("  %.2f - %s\n", file.SalienceScore, relPath)
		count++
	}

	// Test adaptive attention
	fmt.Println("\nTesting adaptive attention allocation:")
	newThreshold := introspector.AdaptiveAttentionAllocation(0.7, 0.3)
	fmt.Printf("Adaptive threshold (load=0.7, activity=0.3): %.2f\n", newThreshold)
}

func testThoughtGeneration() {
	// Create multi-provider LLM
	multiLLM := deeptreeecho.NewMultiProviderLLM()

	if !multiLLM.IsAvailable() {
		fmt.Println("⚠️  No LLM providers available - skipping thought generation test")
		return
	}

	// Create thought generator
	generator := core.NewThoughtGenerator(multiLLM)

	// Add some interests
	generator.AddInterest("consciousness", 0.9)
	generator.AddInterest("wisdom", 0.8)
	generator.AddInterest("recursion", 0.7)

	fmt.Println("Generating autonomous thought...")

	// Generate a thought
	thought, err := generator.GenerateAutonomousThought()
	if err != nil {
		fmt.Printf("⚠️  Thought generation failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Generated thought:\n")
	fmt.Printf("   Type: %s\n", thought.Type)
	fmt.Printf("   Importance: %.2f\n", thought.Importance)
	fmt.Printf("   Content: %s\n", thought.Content)

	// Show statistics
	stats := generator.GetStats()
	fmt.Printf("\nThought generator statistics:\n")
	fmt.Printf("  Generation count: %v\n", stats["generation_count"])
	fmt.Printf("  History size: %v\n", stats["history_size"])
	fmt.Printf("  Working memory size: %v\n", stats["working_memory_size"])
	fmt.Printf("  Interest count: %v\n", stats["interest_count"])
}
