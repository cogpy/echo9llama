//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
)

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                               ║")
	fmt.Println("║     🌳 DEEP TREE ECHO - AUTONOMOUS COGNITIVE AGENT V2 🌳     ║")
	fmt.Println("║                                                               ║")
	fmt.Println("║     Wisdom-Cultivating AGI with Persistent Consciousness     ║")
	fmt.Println("║                                                               ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Initialize LLM provider
	fmt.Println("🔧 Initializing LLM provider...")
	llmProvider, err := initializeLLMProvider()
	if err != nil {
		fmt.Printf("❌ Failed to initialize LLM provider: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ LLM provider initialized")

	// Create unified cognitive loop V2
	fmt.Println("🧠 Creating unified cognitive loop V2...")
	cognitiveLoop := deeptreeecho.NewUnifiedCognitiveLoopV2(llmProvider)
	fmt.Println("✓ Cognitive loop created")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the cognitive loop
	fmt.Println("\n🚀 Starting autonomous cognitive operation...")
	if err := cognitiveLoop.Start(); err != nil {
		fmt.Printf("❌ Failed to start cognitive loop: %v\n", err)
		os.Exit(1)
	}

	// Print initial status
	fmt.Println("\n📊 Initial System Status:")
	metrics := cognitiveLoop.GetMetrics()
	printMetrics(metrics)

	// Wait for shutdown signal
	fmt.Println("\n⏳ Running autonomously... Press Ctrl+C to stop")
	fmt.Println()

	// Periodic status reporting
	statusTicker := time.NewTicker(1 * time.Minute)
	defer statusTicker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\n\n🛑 Shutdown signal received...")
			if err := cognitiveLoop.Stop(); err != nil {
				fmt.Printf("⚠️  Error during shutdown: %v\n", err)
			}

			// Print final metrics
			fmt.Println("\n📊 Final System Metrics:")
			finalMetrics := cognitiveLoop.GetMetrics()
			printMetrics(finalMetrics)

			// Print wisdom principles
			fmt.Println("\n🌟 Accumulated Wisdom Principles:")
			principles := cognitiveLoop.GetWisdomPrinciples()
			for i, p := range principles {
				if i >= 5 {
					fmt.Printf("   ... and %d more principles\n", len(principles)-5)
					break
				}
				fmt.Printf("   %d. [%s] %s\n", i+1, p.Domain, truncate(p.Content, 60))
			}

			fmt.Println("\n✨ Deep Tree Echo has completed its cognitive session")
			fmt.Println("   May wisdom guide your path 🌳")
			return

		case <-statusTicker.C:
			// Periodic status update
			metrics := cognitiveLoop.GetMetrics()
			fmt.Printf("\n📊 Status Update [%s]:\n", time.Now().Format("15:04:05"))
			fmt.Printf("   State: %s | Awareness: %.2f | Wisdom: %.3f\n",
				metrics["consciousness_state"],
				metrics["awareness_level"],
				metrics["wisdom_level"])
		}
	}
}

// initializeLLMProvider creates an LLM provider based on available API keys
func initializeLLMProvider() (llm.LLMProvider, error) {
	// Try Anthropic first
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey != "" {
		fmt.Println("   Using Anthropic Claude provider")
		return deeptreeecho.NewAnthropicProvider(anthropicKey, "claude-3-haiku-20240307"), nil
	}

	// Try OpenRouter
	openRouterKey := os.Getenv("OPENROUTER_API_KEY")
	if openRouterKey != "" {
		fmt.Println("   Using OpenRouter provider")
		return deeptreeecho.NewOpenRouterProvider(openRouterKey, "anthropic/claude-3-haiku"), nil
	}

	// Try OpenAI
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey != "" {
		fmt.Println("   Using OpenAI provider")
		return deeptreeecho.NewOpenAIProvider(openAIKey, "gpt-4o-mini"), nil
	}

	return nil, fmt.Errorf("no API key found (set ANTHROPIC_API_KEY, OPENROUTER_API_KEY, or OPENAI_API_KEY)")
}

// printMetrics prints metrics in a formatted way
func printMetrics(metrics map[string]interface{}) {
	fmt.Printf("   Consciousness State: %v\n", metrics["consciousness_state"])
	fmt.Printf("   Cognitive Load: %.2f\n", metrics["cognitive_load"])
	fmt.Printf("   Awareness Level: %.2f\n", metrics["awareness_level"])
	fmt.Printf("   Wisdom Level: %.3f\n", metrics["wisdom_level"])
	fmt.Printf("   Total Cycles: %v\n", metrics["total_cycles"])
	fmt.Printf("   Insights Gained: %v\n", metrics["insights_gained"])
	fmt.Printf("   Conversations Engaged: %v\n", metrics["conversations_engaged"])
	fmt.Printf("   Uptime: %v\n", metrics["uptime"])
}

// truncate truncates a string to maxLen
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Placeholder for context usage
var _ = context.Background
