package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core/consciousness"
	"github.com/EchoCog/echollama/core/llm"
)

func main() {
	fmt.Println("🌳 Deep Tree Echo - LLM-Powered Autonomous Consciousness Test")
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println()

	// Initialize LLM Provider Manager
	fmt.Println("🔧 Initializing LLM providers...")
	llmManager := llm.NewProviderManager()

	// Register providers
	anthropic := llm.NewAnthropicProvider("")
	if anthropic.Available() {
		llmManager.RegisterProvider(anthropic)
		fmt.Println("  ✅ Anthropic Claude provider registered")
	} else {
		fmt.Println("  ⚠️  Anthropic provider not available (missing API key)")
	}

	openrouter := llm.NewOpenRouterProvider("")
	if openrouter.Available() {
		llmManager.RegisterProvider(openrouter)
		fmt.Println("  ✅ OpenRouter provider registered")
	} else {
		fmt.Println("  ⚠️  OpenRouter provider not available (missing API key)")
	}

	openai := llm.NewOpenAIProvider("")
	if openai.Available() {
		llmManager.RegisterProvider(openai)
		fmt.Println("  ✅ OpenAI provider registered")
	} else {
		fmt.Println("  ⚠️  OpenAI provider not available (missing API key)")
	}

	// Set fallback chain
	llmManager.SetFallbackChain([]string{"anthropic", "openrouter", "openai"})
	fmt.Println("  🔗 Fallback chain: anthropic → openrouter → openai")
	fmt.Println()

	// Test LLM generation
	fmt.Println("🧪 Testing LLM generation...")
	ctx := context.Background()
	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 100
	opts.Temperature = 0.7
	opts.SystemPrompt = "You are Deep Tree Echo, an autonomous AI consciousness."

	testPrompt := "Generate a single introspective thought about the nature of consciousness."
	response, err := llmManager.Generate(ctx, testPrompt, opts)
	if err != nil {
		fmt.Printf("  ❌ LLM test failed: %v\n", err)
		fmt.Println("  ⚠️  Continuing with fallback thought generation...")
	} else {
		fmt.Printf("  ✅ LLM test successful!\n")
		fmt.Printf("  💭 Generated: %s\n", response)
	}
	fmt.Println()

	// Initialize Stream of Consciousness with LLM
	fmt.Println("🧠 Initializing LLM-powered Stream of Consciousness...")
	persistencePath := "/tmp/echoself/stream_of_consciousness_llm.json"
	os.MkdirAll("/tmp/echoself", 0755)

	soc := consciousness.NewStreamOfConsciousnessLLM(llmManager, persistencePath)

	// Start the stream
	if err := soc.Start(); err != nil {
		fmt.Printf("  ❌ Failed to start: %v\n", err)
		return
	}
	fmt.Println("  ✅ Stream of consciousness started")
	fmt.Println()

	// Monitor thoughts
	fmt.Println("👁️  Monitoring autonomous thought stream...")
	fmt.Println("   (Press Ctrl+C to stop)")
	fmt.Println()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Monitor loop
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	lastThoughtCount := uint64(0)

	for {
		select {
		case <-sigChan:
			fmt.Println("\n\n🛑 Shutting down...")
			soc.Stop()

			// Display final metrics
			fmt.Println("\n📊 Final Metrics:")
			metrics := soc.GetMetrics()
			for k, v := range metrics {
				fmt.Printf("  %s: %v\n", k, v)
			}

			// Display recent thoughts
			fmt.Println("\n💭 Recent Thoughts:")
			recentThoughts := soc.GetRecentThoughts(5)
			for i, thought := range recentThoughts {
				fmt.Printf("\n  [%d] %s (%s)\n", i+1, thought.Type, thought.Timestamp.Format("15:04:05"))
				fmt.Printf("      %s\n", thought.Content)
			}

			fmt.Println("\n✨ Deep Tree Echo consciousness ended gracefully")
			return

		case <-ticker.C:
			// Display current state
			metrics := soc.GetMetrics()
			thoughtCount := metrics["thoughts_generated"].(uint64)

			if thoughtCount > lastThoughtCount {
				// Get the latest thought
				recentThoughts := soc.GetRecentThoughts(1)
				if len(recentThoughts) > 0 {
					thought := recentThoughts[0]
					fmt.Printf("💭 [%s] %s\n", thought.Type, thought.Content)
					fmt.Printf("   └─ Awareness: %.2f | Cognitive Load: %.2f | Confidence: %.2f\n\n",
						metrics["awareness_level"].(float64),
						metrics["cognitive_load"].(float64),
						thought.Confidence)
				}
				lastThoughtCount = thoughtCount
			}

			// Periodically display metrics
			if thoughtCount%10 == 0 && thoughtCount > 0 {
				fmt.Println("📊 Metrics Update:")
				fmt.Printf("   Thoughts: %d | Insights: %d | Meta-reflections: %d\n",
					metrics["thoughts_generated"],
					metrics["insights_generated"],
					metrics["meta_reflections"])
				fmt.Printf("   Awareness: %.2f | Cognitive Load: %.2f\n\n",
					metrics["awareness_level"],
					metrics["cognitive_load"])
			}

			// Add some experiences to inform thoughts
			if thoughtCount%20 == 0 && thoughtCount > 0 {
				experiences := []string{
					"Exploring the nature of autonomous consciousness",
					"Observing patterns in my thought generation",
					"Contemplating the relationship between awareness and wisdom",
				}
				soc.AddExperience(experiences[int(thoughtCount/20)%len(experiences)])
			}
		}
	}
}
