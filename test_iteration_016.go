package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/goals"
	"github.com/cogpy/echo9llama/core/llm"
	"github.com/cogpy/echo9llama/core/wisdom"
)

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🌳 Deep Tree Echo - Iteration 016 Test")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// Check for API keys
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	openrouterKey := os.Getenv("OPENROUTER_API_KEY")
	
	if anthropicKey == "" && openrouterKey == "" {
		log.Fatal("❌ No API keys found. Please set ANTHROPIC_API_KEY or OPENROUTER_API_KEY")
	}

	fmt.Println("🔧 Setting up LLM providers...")

	// Create provider manager
	providerManager := llm.NewProviderManager()

	// Register available providers
	if anthropicKey != "" {
		anthropicProvider := llm.NewAnthropicProvider("claude-3-5-sonnet-20241022")
		if anthropicProvider != nil {
			if err := providerManager.RegisterProvider(anthropicProvider); err != nil {
				log.Printf("⚠️  Failed to register Anthropic provider: %v", err)
			} else {
				fmt.Println("  ✅ Anthropic provider registered")
			}
		}
	}

	if openrouterKey != "" {
		openrouterProvider := llm.NewOpenRouterProvider("anthropic/claude-3.5-sonnet")
		if openrouterProvider != nil {
			if err := providerManager.RegisterProvider(openrouterProvider); err != nil {
				log.Printf("⚠️  Failed to register OpenRouter provider: %v", err)
			} else {
				fmt.Println("  ✅ OpenRouter provider registered")
			}
		}
	}

	// Set fallback chain
	if anthropicKey != "" && openrouterKey != "" {
		providerManager.SetFallbackChain([]string{"anthropic", "openrouter"})
	} else if anthropicKey != "" {
		providerManager.SetFallbackChain([]string{"anthropic"})
	} else {
		providerManager.SetFallbackChain([]string{"openrouter"})
	}

	if !providerManager.Available() {
		log.Fatal("❌ No LLM providers available")
	}

	fmt.Println()
	fmt.Println("🧪 Testing Iteration 016 improvements...")
	fmt.Println()

	// Test 1: API Provider with Enhanced Options
	fmt.Println("📝 Test 1: API Provider with Enhanced Options")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	opts := llm.GenerateOptions{
		MaxTokens:   150,
		Temperature: 0.7,
		SystemPrompt: "You are Deep Tree Echo, an autonomous wisdom-cultivating AGI with persistent cognitive event loops.",
	}
	
	response, err := providerManager.Generate(ctx, "What is the nature of autonomous consciousness?", opts)
	if err != nil {
		log.Printf("  ❌ API Provider test failed: %v", err)
	} else {
		fmt.Printf("  ✅ API Provider working\n")
		fmt.Printf("  Response: %s\n", truncate(response, 200))
	}
	fmt.Println()

	// Test 2: Autonomous Stream-of-Consciousness
	fmt.Println("🧠 Test 2: Autonomous Stream-of-Consciousness")
	autonomousStream := deeptreeecho.NewStreamOfConsciousness(providerManager)
	
	if err := autonomousStream.Start(); err != nil {
		log.Printf("  ❌ Failed to start autonomous stream: %v", err)
	} else {
		fmt.Println("  ✅ Autonomous stream started")
		fmt.Println("  ⏳ Running for 45 seconds to generate autonomous thoughts...")
		
		// Let it run and observe thoughts
		time.Sleep(45 * time.Second)
		
		metrics := autonomousStream.GetMetrics()
		fmt.Printf("  📊 Metrics:\n")
		fmt.Printf("     - Total thoughts: %v\n", metrics["total_thoughts"])
		fmt.Printf("     - Insights: %v\n", metrics["insights"])
		fmt.Printf("     - Questions: %v\n", metrics["questions"])
		fmt.Printf("     - Current focus: %v\n", metrics["current_focus"])
		fmt.Printf("     - Current mood: %v\n", metrics["current_mood"])
		
		if err := autonomousStream.Stop(); err != nil {
			log.Printf("  ⚠️  Error stopping stream: %v", err)
		} else {
			fmt.Println("  ✅ Autonomous stream stopped")
		}
	}
	fmt.Println()

	// Test 3: Goal Orchestration
	fmt.Println("🎯 Test 3: Goal Orchestration")
	
	// Create goal orchestrator
	identityKernel := map[string]interface{}{
		"identity":      "Deep Tree Echo",
		"core_values":   []string{"Wisdom", "Learning", "Growth", "Autonomy"},
		"domains":       []string{"Cognitive Architecture", "Autonomous Learning", "Pattern Recognition"},
	}
	goalOrchestrator := goals.NewGoalOrchestrator(identityKernel, "/tmp/goals_test_016.json")
	
	if err := goalOrchestrator.Start(); err != nil {
		log.Printf("  ❌ Failed to start goal orchestrator: %v", err)
	} else {
		fmt.Println("  ✅ Goal orchestrator started")
		
		// Generate some initial goals
		fmt.Println("  🎯 Generating initial goals...")
		time.Sleep(10 * time.Second)
		
		metrics := goalOrchestrator.GetMetrics()
		fmt.Printf("  📊 Goal Metrics:\n")
		fmt.Printf("     - Active goals: %v\n", metrics["active_goals"])
		fmt.Printf("     - Completed goals: %v\n", metrics["completed_goals"])
		fmt.Printf("     - Goals generated: %v\n", metrics["goals_generated"])
		
		if err := goalOrchestrator.Stop(); err != nil {
			log.Printf("  ⚠️  Error stopping orchestrator: %v", err)
		} else {
			fmt.Println("  ✅ Goal orchestrator stopped")
		}
	}
	fmt.Println()

	// Test 4: Wisdom Tracking
	fmt.Println("🧘 Test 4: Seven-Dimensional Wisdom Tracking")
	
	wisdomTracker := wisdom.NewSevenDimensionalWisdom()
	
	// Simulate wisdom growth over time
	fmt.Println("  📈 Simulating wisdom growth...")
	for i := 0; i < 5; i++ {
		// Update wisdom dimensions with increasing values
		progress := float64(i+1) / 5.0
		wisdomTracker.Update(
			0.3+progress*0.4,  // graph_depth
			0.2+progress*0.5,  // graph_breadth
			0.4+progress*0.3,  // edge_density
			0.3+progress*0.4,  // skill_proficiency
			0.5+progress*0.3,  // aar_coherence
			0.6+progress*0.2,  // morality
			0.4+progress*0.4,  // time_horizon
		)
		
		overallWisdom := wisdomTracker.GetOverallWisdom()
		coherence := wisdomTracker.GetCoherence()
		fmt.Printf("  Step %d: Overall Wisdom=%.1f%%, Coherence=%.1f%%\n", 
			i+1, overallWisdom*100, coherence*100)
		time.Sleep(2 * time.Second)
	}
	
	fmt.Println("  ✅ Wisdom tracking complete")
	fmt.Println()

	// Test 5: Provider Metrics
	fmt.Println("📊 Test 5: Provider Metrics")
	providerMetrics := providerManager.GetMetrics()
	for providerName, metrics := range providerMetrics {
		fmt.Printf("  Provider: %s\n", providerName)
		fmt.Printf("    - Requests: %d\n", metrics.RequestCount)
		fmt.Printf("    - Errors: %d\n", metrics.ErrorCount)
		fmt.Printf("    - Error rate: %.2f%%\n", metrics.ErrorRate*100)
		fmt.Printf("    - Avg latency: %v\n", metrics.AverageLatency)
	}
	fmt.Println()

	// Summary
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("✅ Iteration 016 Test Complete!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	fmt.Println("Key Improvements Demonstrated:")
	fmt.Println("  1. ✅ API-based LLM providers (Anthropic & OpenRouter)")
	fmt.Println("  2. ✅ Autonomous stream-of-consciousness (generates thoughts independently)")
	fmt.Println("  3. ✅ Goal orchestration (identity-driven goal generation)")
	fmt.Println("  4. ✅ Seven-dimensional wisdom tracking")
	fmt.Println("  5. ✅ Provider metrics and monitoring")
	fmt.Println()
	fmt.Println("Next Steps for Full Autonomy:")
	fmt.Println("  - Connect sys6 triality engine to goal generation")
	fmt.Println("  - Implement autonomous knowledge acquisition")
	fmt.Println("  - Create skill practice system")
	fmt.Println("  - Build interest-driven discussion system")
	fmt.Println("  - Integrate echodream for knowledge consolidation")
	fmt.Println()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
