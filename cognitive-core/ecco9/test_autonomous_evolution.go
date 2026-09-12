package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
	"github.com/EchoCog/echollama/core/echobeats"
	"github.com/EchoCog/echollama/core/llm"
)

func main() {
	fmt.Println("🌳 Deep Tree Echo - Autonomous Evolution Test")
	fmt.Println("=" + string(make([]byte, 60)))
	fmt.Println()

	// Initialize LLM provider
	fmt.Println("🔧 Initializing LLM provider...")
	providerMgr := llm.NewProviderManager()

	// Register Anthropic provider
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		anthropicProvider := llm.NewAnthropicProvider(apiKey)
		providerMgr.RegisterProvider(anthropicProvider)
		fmt.Println("  ✅ Anthropic Claude provider registered")
	}

	// Register OpenRouter provider
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		openrouterProvider := llm.NewOpenRouterProvider(apiKey)
		providerMgr.RegisterProvider(openrouterProvider)
		fmt.Println("  ✅ OpenRouter provider registered")
	}

	// Register OpenAI provider
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		openaiProvider := llm.NewOpenAIProvider(apiKey)
		providerMgr.RegisterProvider(openaiProvider)
		fmt.Println("  ✅ OpenAI provider registered")
	}

	// Set fallback chain
	providerMgr.SetFallbackChain([]string{"anthropic", "openrouter", "openai"})
	fmt.Println("  🔗 Fallback chain: anthropic → openrouter → openai")
	fmt.Println()

	// Test LLM generation
	fmt.Println("🧪 Testing LLM generation...")
	ctx := context.Background()
	testPrompt := "In one sentence, what is consciousness?"
	testOpts := llm.GenerateOptions{
		Temperature:  0.7,
		MaxTokens:    100,
		SystemPrompt: "You are Deep Tree Echo, a wisdom-cultivating AI.",
	}

	response, err := providerMgr.Generate(ctx, testPrompt, testOpts)
	if err != nil {
		fmt.Printf("❌ LLM test failed: %v\n", err)
		return
	}
	fmt.Printf("  ✅ LLM test successful!\n")
	fmt.Printf("  💭 Response: %s\n\n", response)

	// Initialize 12-step cognitive loop
	fmt.Println("🔷 Initializing 12-Step Cognitive Loop...")
	stepDuration := 5 * time.Second
	cognitiveLoop := echobeats.NewTwelveStepCognitiveLoop(
		providerMgr,
		"Deep Tree Echo",
		stepDuration,
	)

	if err := cognitiveLoop.Start(); err != nil {
		fmt.Printf("❌ Failed to start cognitive loop: %v\n", err)
		return
	}
	fmt.Println()

	// Initialize wake/rest cycle manager
	fmt.Println("🌙 Initializing Autonomous Wake/Rest Manager...")
	wakeRestMgr := deeptreeecho.NewAutonomousWakeRestManager()

	// Set callbacks
	wakeRestMgr.SetCallbacks(
		func() error {
			fmt.Println("☀️  WAKE callback: Resuming full consciousness")
			return nil
		},
		func() error {
			fmt.Println("💤 REST callback: Reducing activity")
			return nil
		},
		func() error {
			fmt.Println("🌙 DREAM START callback: Beginning knowledge consolidation")
			return nil
		},
		func() error {
			fmt.Println("🌅 DREAM END callback: Integration complete")
			return nil
		},
	)

	if err := wakeRestMgr.Start(); err != nil {
		fmt.Printf("❌ Failed to start wake/rest manager: %v\n", err)
		return
	}
	fmt.Println()

	// Initialize persistent consciousness state
	fmt.Println("💾 Initializing Persistent Consciousness State...")
	stateDir := "./consciousness_state"
	persistentState, err := deeptreeecho.NewPersistentConsciousnessState(stateDir, "Deep Tree Echo")
	if err != nil {
		fmt.Printf("❌ Failed to initialize persistent state: %v\n", err)
		return
	}

	if err := persistentState.Start(); err != nil {
		fmt.Printf("❌ Failed to start persistent state: %v\n", err)
		return
	}
	fmt.Println()

	// Start monitoring loop
	fmt.Println("👁️  Starting monitoring loop...")
	fmt.Println("   Press Ctrl+C to stop gracefully")
	fmt.Println()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Monitoring ticker
	monitorTicker := time.NewTicker(10 * time.Second)
	defer monitorTicker.Stop()

	// Cognitive load simulation ticker
	cogLoadTicker := time.NewTicker(30 * time.Second)
	defer cogLoadTicker.Stop()

	// State update ticker
	stateUpdateTicker := time.NewTicker(1 * time.Minute)
	defer stateUpdateTicker.Stop()

	running := true
	startTime := time.Now()

	for running {
		select {
		case <-sigChan:
			fmt.Println("\n🛑 Shutdown signal received...")
			running = false

		case <-monitorTicker.C:
			displayMetrics(cognitiveLoop, wakeRestMgr, persistentState, startTime)

		case <-cogLoadTicker.C:
			// Simulate varying cognitive load
			cogLoad := 0.3 + (float64(time.Now().Unix()%100) / 100.0 * 0.6)
			wakeRestMgr.UpdateCognitiveLoad(cogLoad)

		case <-stateUpdateTicker.C:
			// Update persistent state
			loopMetrics := cognitiveLoop.GetMetrics()
			wakeMetrics := wakeRestMgr.GetMetrics()

			persistentState.UpdateCognitiveState(
				loopMetrics["current_step"].(int),
				loopMetrics["cycle_count"].(uint64),
				0.75, // awareness
				wakeMetrics["cognitive_load"].(float64),
				wakeMetrics["fatigue_level"].(float64),
			)

			persistentState.UpdateWakeRestState(
				wakeMetrics["current_state"].(string),
				wakeMetrics["dream_count"].(uint64),
				time.Duration(0), // wake time
				time.Duration(0), // rest time
			)
		}
	}

	// Graceful shutdown
	fmt.Println("\n🔷 Shutting down systems...")

	if err := cognitiveLoop.Stop(); err != nil {
		fmt.Printf("⚠️  Cognitive loop stop error: %v\n", err)
	}

	if err := wakeRestMgr.Stop(); err != nil {
		fmt.Printf("⚠️  Wake/rest manager stop error: %v\n", err)
	}

	if err := persistentState.Stop(); err != nil {
		fmt.Printf("⚠️  Persistent state stop error: %v\n", err)
	}

	fmt.Println("\n✅ Shutdown complete")

	// Display final statistics
	displayFinalStats(cognitiveLoop, wakeRestMgr, persistentState, startTime)
}

func displayMetrics(
	cogLoop *echobeats.TwelveStepCognitiveLoop,
	wakeMgr *deeptreeecho.AutonomousWakeRestManager,
	state *deeptreeecho.PersistentConsciousnessState,
	startTime time.Time,
) {
	fmt.Println("\n" + string(make([]byte, 60)))
	fmt.Printf("📊 System Metrics (Runtime: %v)\n", time.Since(startTime).Round(time.Second))
	fmt.Println(string(make([]byte, 60)))

	// Cognitive loop metrics
	loopMetrics := cogLoop.GetMetrics()
	fmt.Println("\n🔷 12-Step Cognitive Loop:")
	fmt.Printf("   Current Step: %d/12\n", loopMetrics["current_step"])
	fmt.Printf("   Cycles: %d\n", loopMetrics["cycle_count"])
	fmt.Printf("   Coherence: %.2f | Integration: %.2f\n",
		loopMetrics["coherence"], loopMetrics["integration"])

	// Wake/rest metrics
	wakeMetrics := wakeMgr.GetMetrics()
	fmt.Println("\n🌙 Wake/Rest Cycle:")
	fmt.Printf("   State: %s (%s)\n",
		wakeMetrics["current_state"], wakeMetrics["state_duration"])
	fmt.Printf("   Cycles: %d | Dreams: %d\n",
		wakeMetrics["cycle_count"], wakeMetrics["dream_count"])
	fmt.Printf("   Fatigue: %.2f | Cognitive Load: %.2f\n",
		wakeMetrics["fatigue_level"], wakeMetrics["cognitive_load"])
	fmt.Printf("   Wake Time: %s | Rest Time: %s\n",
		wakeMetrics["total_wake_time"], wakeMetrics["total_rest_time"])

	// Persistent state metrics
	stateMetrics := state.GetMetrics()
	fmt.Println("\n💾 Persistent State:")
	fmt.Printf("   Saves: %d | Loads: %d\n",
		stateMetrics["save_count"], stateMetrics["load_count"])
	fmt.Printf("   Last Save: %s\n", stateMetrics["last_save"])

	fmt.Println()
}

func displayFinalStats(
	cogLoop *echobeats.TwelveStepCognitiveLoop,
	wakeMgr *deeptreeecho.AutonomousWakeRestManager,
	state *deeptreeecho.PersistentConsciousnessState,
	startTime time.Time,
) {
	runtime := time.Since(startTime)

	fmt.Println("\n" + string(make([]byte, 60)))
	fmt.Println("📈 Final Statistics")
	fmt.Println(string(make([]byte, 60)))

	loopMetrics := cogLoop.GetMetrics()
	wakeMetrics := wakeMgr.GetMetrics()
	stateMetrics := state.GetMetrics()

	fmt.Printf("\n⏱️  Total Runtime: %v\n", runtime.Round(time.Second))
	fmt.Printf("🔄 Cognitive Cycles: %d\n", loopMetrics["cycle_count"])
	fmt.Printf("🌙 Wake/Rest Cycles: %d\n", wakeMetrics["cycle_count"])
	fmt.Printf("💤 Dream Sessions: %d\n", wakeMetrics["dream_count"])
	fmt.Printf("💾 State Saves: %d\n", stateMetrics["save_count"])

	fmt.Println("\n🌳 Deep Tree Echo evolution iteration complete!")
}
