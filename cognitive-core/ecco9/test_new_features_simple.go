package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
	"github.com/EchoCog/echollama/core/echoself"
)

func main() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("🧪 Echo9llama November 21 Iteration - New Features Test")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Test 1: Multi-Provider LLM
	fmt.Println("Test 1: Multi-Provider LLM System")
	fmt.Println(strings.Repeat("-", 41))
	testMultiProviderLLM()
	fmt.Println()

	// Test 2: Repository Introspection
	fmt.Println("Test 2: Repository Self-Introspection")
	fmt.Println(strings.Repeat("-", 41))
	testRepositoryIntrospection()
	fmt.Println()

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("✅ All tests completed")
	fmt.Println(strings.Repeat("=", 80))
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
		fmt.Println("   Set ANTHROPIC_API_KEY, OPENROUTER_API_KEY, or OPENAI_API_KEY")
		return
	}

	// Test thought generation (only if providers available)
	fmt.Println("\nAttempting thought generation...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	thought, err := multiLLM.GenerateThought(ctx, "What is the nature of autonomous cognition?")
	if err != nil {
		fmt.Printf("⚠️  Thought generation failed: %v\n", err)
	} else {
		fmt.Printf("✅ Generated thought:\n")
		fmt.Printf("   %s\n", thought)
	}

	// Show statistics
	fmt.Println("\nProvider statistics:")
	stats := multiLLM.GetStats()
	for provider, stat := range stats {
		if stat.Requests > 0 {
			fmt.Printf("  %s:\n", provider)
			fmt.Printf("    Requests: %d\n", stat.Requests)
			fmt.Printf("    Successes: %d\n", stat.Successes)
			fmt.Printf("    Failures: %d\n", stat.Failures)
			if stat.Successes > 0 {
				avgLatency := stat.TotalLatency / time.Duration(stat.Successes)
				fmt.Printf("    Avg Latency: %v\n", avgLatency)
			}
		}
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
	fmt.Println("Attention threshold: 0.6")
	fmt.Println("This may take a moment...")

	// Scan repository
	startTime := time.Now()
	err := introspector.Scan()
	scanDuration := time.Since(startTime)

	if err != nil {
		fmt.Printf("❌ Scan failed: %v\n", err)
		return
	}

	// Show statistics
	stats := introspector.GetStats()
	fmt.Printf("\n✅ Scan complete in %v!\n", scanDuration)
	fmt.Printf("Total files examined: %v\n", stats["total_files"])
	fmt.Printf("High-salience files: %v\n", stats["scanned_files"])
	fmt.Printf("Attention threshold: %v\n", stats["attention_threshold"])

	// Show high-salience files
	fmt.Println("\nTop 15 high-salience files:")
	files := introspector.GetHighSalienceFiles()
	count := 0
	for _, file := range files {
		if count >= 15 {
			break
		}
		relPath := strings.TrimPrefix(file.Path, repoRoot+"/")
		if len(relPath) > 65 {
			relPath = "..." + relPath[len(relPath)-62:]
		}
		fmt.Printf("  %.2f - %s\n", file.SalienceScore, relPath)
		count++
	}

	// Test adaptive attention
	fmt.Println("\nTesting adaptive attention allocation:")
	scenarios := []struct {
		load     float64
		activity float64
	}{
		{0.3, 0.8}, // Low load, high activity
		{0.7, 0.3}, // High load, low activity
		{0.5, 0.5}, // Balanced
	}

	for _, scenario := range scenarios {
		newThreshold := introspector.AdaptiveAttentionAllocation(scenario.load, scenario.activity)
		fmt.Printf("  Load=%.1f, Activity=%.1f → Threshold=%.2f\n",
			scenario.load, scenario.activity, newThreshold)
	}

	// Generate summary
	fmt.Println("\nGenerating hypergraph summary...")
	summary := introspector.GenerateHypergraphSummary()
	lines := strings.Split(summary, "\n")
	for i, line := range lines {
		if i > 20 {
			fmt.Println("  ... (truncated)")
			break
		}
		fmt.Println(line)
	}
}
