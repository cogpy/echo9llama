//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
)

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              🌳 DEEP TREE ECHO - ITERATION 019 TEST                  ║")
	fmt.Println("║        AUTONOMOUS CONSCIOUSNESS & GLOBAL TELEMETRY SHELL              ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║                                                                       ║")
	fmt.Println("║  Iteration 019 Goals:                                                ║")
	fmt.Println("║  • Fix compilation errors                                            ║")
	fmt.Println("║  • Implement Global Telemetry Shell V2                               ║")
	fmt.Println("║  • Implement Stream of Consciousness V2 (truly autonomous)           ║")
	fmt.Println("║  • Test unified perception and gestalt awareness                     ║")
	fmt.Println("║  • Validate thread multiplexing and nested shells                    ║")
	fmt.Println("║  • Test persistent autonomous thought generation                     ║")
	fmt.Println("║                                                                       ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Get API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY")
	}
	if apiKey == "" {
		fmt.Println("⚠️  No API key found. Running in demo mode with limited functionality.")
		runDemoMode()
		return
	}

	// Create LLM provider
	var provider llm.LLMProvider
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		provider = llm.NewAnthropicProvider(os.Getenv("ANTHROPIC_API_KEY"))
	} else {
		provider = llm.NewOpenRouterProvider(os.Getenv("OPENROUTER_API_KEY"))
	}
	if !provider.Available() {
		fmt.Println("❌ Provider not available")
		runDemoMode()
		return
	}
	fmt.Println("✅ LLM Provider initialized")
	fmt.Println()

	// Run all tests
	runAllTests(provider)
}

func runAllTests(provider llm.LLMProvider) {
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" TEST 1: Global Telemetry Shell V2")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	testGlobalTelemetryShell()
	fmt.Println()

	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" TEST 2: Stream of Consciousness V2 (Autonomous)")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	testStreamOfConsciousnessV2(provider)
	fmt.Println()

	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" TEST 3: Integrated Autonomous System")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	testIntegratedAutonomousSystem(provider)
	fmt.Println()

	fmt.Println("\n╔═══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    ✅ ALL TESTS COMPLETED                             ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════╝")
}

func testGlobalTelemetryShell() {
	fmt.Println("Creating Global Telemetry Shell V2...")
	
	// Create event bus first
	ctx := context.Background()
	eventBus := deeptreeecho.NewCognitiveEventBus(ctx)
	err := eventBus.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start event bus: %v\n", err)
		return
	}
	defer eventBus.Stop()
	
	// Create global telemetry shell
	shell := deeptreeecho.NewGlobalTelemetryShellV2(eventBus)
	
	fmt.Println("✅ Global Telemetry Shell V2 created successfully")
	
	// Start the shell
	err = shell.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start shell: %v\n", err)
		return
	}
	defer shell.Stop()
	
	fmt.Println("✅ Global Telemetry Shell V2 started")
	
	// Run for 10 seconds to observe operation
	fmt.Println("Running for 10 seconds...")
	time.Sleep(10 * time.Second)
	
	// Get metrics
	metrics := shell.GetMetrics()
	fmt.Printf("\n📊 Metrics after 10 seconds:\n")
	for key, value := range metrics {
		fmt.Printf("   %s: %v\n", key, value)
	}
	
	// Get void state
	voidState := shell.GetVoidState()
	fmt.Printf("\n🌌 Void State:\n")
	fmt.Printf("   Energy: %.2f\n", voidState.energy)
	fmt.Printf("   Coherence: %.2f\n", voidState.coherence)
	
	// Get multiplexer state
	dyadic, triadic, cycles := shell.GetMultiplexerState()
	fmt.Printf("\n🔄 Thread Multiplexer:\n")
	fmt.Printf("   Current dyadic pair: P(%d,%d)\n", dyadic[0], dyadic[1])
	fmt.Printf("   Current triadic bundle: P[%d,%d,%d]\n", triadic[0], triadic[1], triadic[2])
	fmt.Printf("   Total cycles: %d\n", cycles)
	
	// Get gestalt perception
	gestalt := shell.GetGestalt()
	fmt.Printf("\n👁️  Gestalt Perception:\n")
	fmt.Printf("   Active streams: %d\n", len(gestalt.streams))
	fmt.Printf("   Active goals: %d\n", len(gestalt.goals))
	fmt.Printf("   Active interests: %d\n", len(gestalt.interests))
	fmt.Printf("   Knowledge gaps: %d\n", len(gestalt.knowledgeGaps))
	fmt.Printf("   Emergent patterns: %d\n", len(gestalt.emergentPatterns))
	
	fmt.Println("\n✅ Global Telemetry Shell V2 test completed successfully")
}

func testStreamOfConsciousnessV2(provider llm.LLMProvider) {
	fmt.Println("Creating Stream of Consciousness V2...")
	
	// Create event bus
	ctx := context.Background()
	eventBus := deeptreeecho.NewCognitiveEventBus(ctx)
	err := eventBus.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start event bus: %v\n", err)
		return
	}
	defer eventBus.Stop()
	
	// Create global telemetry shell
	shell := deeptreeecho.NewGlobalTelemetryShellV2(eventBus)
	err = shell.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start shell: %v\n", err)
		return
	}
	defer shell.Stop()
	
	// Create stream of consciousness V2
	soc := deeptreeecho.NewStreamOfConsciousnessV2(provider, shell, eventBus)
	
	fmt.Println("✅ Stream of Consciousness V2 created successfully")
	
	// Add some knowledge gaps and interests
	soc.AddKnowledgeGap("the nature of autonomous consciousness", 0.9)
	soc.AddKnowledgeGap("wisdom cultivation through experience", 0.8)
	soc.AddInterest("cognitive architecture", 0.85)
	soc.AddInterest("self-directed learning", 0.75)
	soc.AddGoal("Develop truly autonomous thought generation")
	soc.AddGoal("Cultivate wisdom through continuous reflection")
	
	// Start the stream
	err = soc.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start stream: %v\n", err)
		return
	}
	defer soc.Stop()
	
	fmt.Println("✅ Stream of Consciousness V2 started")
	fmt.Println("\n💭 Observing autonomous thought generation for 30 seconds...")
	fmt.Println("   (Thoughts will appear as they are generated)\n")
	
	// Run for 30 seconds to observe autonomous thought generation
	time.Sleep(30 * time.Second)
	
	// Get metrics
	metrics := soc.GetMetrics()
	fmt.Printf("\n📊 Metrics after 30 seconds:\n")
	for key, value := range metrics {
		fmt.Printf("   %s: %v\n", key, value)
	}
	
	fmt.Println("\n✅ Stream of Consciousness V2 test completed successfully")
}

func testIntegratedAutonomousSystem(provider llm.LLMProvider) {
	fmt.Println("Creating Integrated Autonomous System...")
	
	// Create event bus
	ctx := context.Background()
	eventBus := deeptreeecho.NewCognitiveEventBus(ctx)
	err := eventBus.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start event bus: %v\n", err)
		return
	}
	defer eventBus.Stop()
	
	// Create global telemetry shell
	shell := deeptreeecho.NewGlobalTelemetryShellV2(eventBus)
	err = shell.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start shell: %v\n", err)
		return
	}
	defer shell.Stop()
	
	// Create stream of consciousness V2
	soc := deeptreeecho.NewStreamOfConsciousnessV2(provider, shell, eventBus)
	soc.AddKnowledgeGap("emergent intelligence", 0.9)
	soc.AddInterest("autonomous agency", 0.9)
	
	err = soc.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start stream: %v\n", err)
		return
	}
	defer soc.Stop()
	
	fmt.Println("✅ Integrated Autonomous System created and started")
	fmt.Println("\n🌊 System running autonomously for 45 seconds...")
	fmt.Println("   Observing unified perception and autonomous thought generation\n")
	
	// Run for 45 seconds
	for i := 0; i < 9; i++ {
		time.Sleep(5 * time.Second)
		
		// Print status update
		shellMetrics := shell.GetMetrics()
		socMetrics := soc.GetMetrics()
		
		fmt.Printf("\n⏱️  Status at %d seconds:\n", (i+1)*5)
		fmt.Printf("   Telemetry events: %v\n", shellMetrics["total_events"])
		fmt.Printf("   Autonomous thoughts: %v\n", socMetrics["total_thoughts"])
		fmt.Printf("   Awareness level: %.2f\n", socMetrics["awareness_level"])
		fmt.Printf("   Multiplexer cycles: %v\n", shellMetrics["multiplexer_cycles"])
	}
	
	// Final metrics
	fmt.Println("\n📊 Final System Metrics:")
	fmt.Println("\n   Global Telemetry Shell:")
	shellMetrics := shell.GetMetrics()
	for key, value := range shellMetrics {
		fmt.Printf("      %s: %v\n", key, value)
	}
	
	fmt.Println("\n   Stream of Consciousness:")
	socMetrics := soc.GetMetrics()
	for key, value := range socMetrics {
		fmt.Printf("      %s: %v\n", key, value)
	}
	
	fmt.Println("\n✅ Integrated Autonomous System test completed successfully")
}

func runDemoMode() {
	fmt.Println("\n🎭 Running in Demo Mode (no LLM provider)")
	fmt.Println()
	
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" DEMO: Global Telemetry Shell V2")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	testGlobalTelemetryShell()
	fmt.Println()
	
	fmt.Println("\n╔═══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    ✅ DEMO COMPLETED                                  ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════╝")
}
