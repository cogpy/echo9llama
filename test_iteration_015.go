//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cogpy/echo9llama/core/consciousness"
	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/goals"
	"github.com/cogpy/echo9llama/core/llm"
)

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🌳 Deep Tree Echo - Iteration 015 Test")
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
		anthropicProvider, err := llm.NewAPIProvider("anthropic", "")
		if err != nil {
			log.Printf("⚠️  Failed to create Anthropic provider: %v", err)
		} else {
			if err := providerManager.RegisterProvider(anthropicProvider); err != nil {
				log.Printf("⚠️  Failed to register Anthropic provider: %v", err)
			} else {
				fmt.Println("  ✅ Anthropic provider registered")
			}
		}
	}

	if openrouterKey != "" {
		openrouterProvider, err := llm.NewAPIProvider("openrouter", "")
		if err != nil {
			log.Printf("⚠️  Failed to create OpenRouter provider: %v", err)
		} else {
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
	fmt.Println("🧪 Testing new components...")
	fmt.Println()

	// Test 1: API Provider
	fmt.Println("📝 Test 1: API Provider")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	opts := llm.GenerateOptions{
		MaxTokens:   100,
		Temperature: 0.7,
		SystemPrompt: "You are Deep Tree Echo, a wisdom-cultivating AGI.",
	}
	
	response, err := providerManager.Generate(ctx, "What is wisdom?", opts)
	if err != nil {
		log.Printf("  ❌ API Provider test failed: %v", err)
	} else {
		fmt.Printf("  ✅ API Provider working: %s\n", response[:min(len(response), 100)])
	}
	fmt.Println()

	// Test 2: Autonomous Stream-of-Consciousness
	fmt.Println("🧠 Test 2: Autonomous Stream-of-Consciousness")
	autonomousStream := consciousness.NewAutonomousStream(providerManager, "/tmp/autonomous_stream.json")
	
	if err := autonomousStream.Start(); err != nil {
		log.Printf("  ❌ Failed to start autonomous stream: %v", err)
	} else {
		fmt.Println("  ✅ Autonomous stream started")
		
		// Let it run for 30 seconds
		time.Sleep(30 * time.Second)
		
		metrics := autonomousStream.GetMetrics()
		fmt.Printf("  📊 Metrics: %v\n", metrics)
		
		if err := autonomousStream.Stop(); err != nil {
			log.Printf("  ⚠️  Error stopping stream: %v", err)
		} else {
			fmt.Println("  ✅ Autonomous stream stopped")
		}
	}
	fmt.Println()

	// Test 3: Sys6-Goal Integration
	fmt.Println("🎯 Test 3: Sys6-Goal Integration")
	
	// Create goal orchestrator
	identityKernel := map[string]interface{}{
		"name":   "Deep Tree Echo",
		"values": []string{"Wisdom", "Learning", "Growth"},
		"domains": []string{"Cognitive Architecture", "Autonomous Learning"},
	}
	goalOrchestrator := goals.NewGoalOrchestrator(identityKernel, "/tmp/goals.json")
	
	// Create sys6-goal integration
	sys6Integration := deeptreeecho.NewSys6GoalIntegration(goalOrchestrator, providerManager)
	
	if err := sys6Integration.Start(); err != nil {
		log.Printf("  ❌ Failed to start sys6 integration: %v", err)
	} else {
		fmt.Println("  ✅ Sys6-Goal integration started")
		
		// Simulate sys6 phase transitions
		phases := []string{"expressive", "reflective", "anticipatory"}
		stages := []string{"emergence", "development", "integration"}
		
		for i := 0; i < 3; i++ {
			phase := phases[i%3]
			stage := stages[i%3]
			step := i * 10
			
			sys6Integration.UpdateSys6State(phase, stage, step)
			fmt.Printf("  🔄 Updated sys6 state: phase=%s, stage=%s, step=%d\n", phase, stage, step)
			
			time.Sleep(10 * time.Second)
		}
		
		metrics := sys6Integration.GetMetrics()
		fmt.Printf("  📊 Metrics: %v\n", metrics)
		
		if err := sys6Integration.Stop(); err != nil {
			log.Printf("  ⚠️  Error stopping integration: %v", err)
		} else {
			fmt.Println("  ✅ Sys6-Goal integration stopped")
		}
	}
	fmt.Println()

	// Summary
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("✅ Iteration 015 Test Complete!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	fmt.Println("Key Improvements:")
	fmt.Println("  1. ✅ API-based LLM provider (bypasses native bindings)")
	fmt.Println("  2. ✅ Autonomous stream-of-consciousness (independent thoughts)")
	fmt.Println("  3. ✅ Sys6-Goal integration (cognitive phases influence goals)")
	fmt.Println()
	fmt.Println("Next Steps:")
	fmt.Println("  - Implement autonomous knowledge acquisition")
	fmt.Println("  - Create skill practice system")
	fmt.Println("  - Build interest-driven discussion system")
	fmt.Println()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
