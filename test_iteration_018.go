//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cogpy/echo9llama/core/echobeats"
	"github.com/cogpy/echo9llama/core/llm"
)

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              🌳 DEEP TREE ECHO - ITERATION 018 TEST                  ║")
	fmt.Println("║          FOUNDATION REPAIR & ECHOBEATS CORE                           ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║                                                                       ║")
	fmt.Println("║  Iteration 018 Goals:                                                ║")
	fmt.Println("║  • Fix all compilation errors                                        ║")
	fmt.Println("║  • Implement EchoBeats 12-step cognitive loop                        ║")
	fmt.Println("║  • Implement 3 concurrent streams with 120° phase offset             ║")
	fmt.Println("║  • Validate core architecture improvements                           ║")
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
	fmt.Println(" TEST 1: EchoBeats 12-Step Cognitive Loop")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	testCognitiveLoop()
	fmt.Println()

	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" TEST 2: EchoBeats Three-Phase System (3 Concurrent Streams)")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	testThreePhaseSystem()
	fmt.Println()

	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" TEST 3: Interest Pattern System")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	testInterestPatternSystem()
	fmt.Println()

	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" TEST 4: Discussion Manager")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	testDiscussionManager()
	fmt.Println()

	fmt.Println("\n╔═══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    ✅ ALL TESTS COMPLETED                             ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════╝")
}

func testCognitiveLoop() {
	fmt.Println("Creating 12-step cognitive loop...")
	
	loop := echobeats.NewCognitiveLoop()
	
	fmt.Println("✅ CognitiveLoop created successfully")
	fmt.Printf("   Initial step: %v\n", loop.GetCurrentState().StepNumber)
	fmt.Printf("   Initial mode: %v\n", loop.GetCurrentState().Mode)
	
	// Start the loop
	err := loop.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start loop: %v\n", err)
		return
	}
	
	fmt.Println("✅ CognitiveLoop started")
	
	// Run for 15 seconds (about 7-8 steps at 2s per step)
	fmt.Println("Running for 15 seconds...")
	time.Sleep(15 * time.Second)
	
	// Get metrics
	metrics := loop.GetMetrics()
	fmt.Printf("\n📊 Metrics after 15 seconds:\n")
	fmt.Printf("   Current step: %v\n", metrics["current_step"])
	fmt.Printf("   Cycle count: %v\n", metrics["cycle_count"])
	fmt.Printf("   Total steps: %v\n", metrics["total_steps"])
	fmt.Printf("   Current mode: %v\n", metrics["current_mode"])
	
	// Stop the loop
	err = loop.Stop()
	if err != nil {
		fmt.Printf("❌ Failed to stop loop: %v\n", err)
		return
	}
	
	fmt.Println("✅ CognitiveLoop stopped successfully")
}

func testThreePhaseSystem() {
	fmt.Println("Creating EchoBeats Three-Phase System...")
	
	system := echobeats.NewEchoBeatsThreePhase()
	
	fmt.Println("✅ Three-Phase System created successfully")
	
	// Start the system
	err := system.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start system: %v\n", err)
		return
	}
	
	fmt.Println("✅ Three-Phase System started")
	
	// Run for 20 seconds to see concurrent streams
	fmt.Println("Running for 20 seconds...")
	time.Sleep(20 * time.Second)
	
	// Get metrics
	metrics := system.GetMetrics()
	fmt.Printf("\n📊 Metrics after 20 seconds:\n")
	fmt.Printf("   Cycle count: %v\n", metrics["cycle_count"])
	fmt.Printf("   Total steps: %v\n", metrics["total_steps"])
	fmt.Printf("   Stream 1 cycles: %v\n", metrics["stream1_cycles"])
	fmt.Printf("   Stream 2 cycles: %v\n", metrics["stream2_cycles"])
	fmt.Printf("   Stream 3 cycles: %v\n", metrics["stream3_cycles"])
	
	// Stop the system
	err = system.Stop()
	if err != nil {
		fmt.Printf("❌ Failed to stop system: %v\n", err)
		return
	}
	
	fmt.Println("✅ Three-Phase System stopped successfully")
}

func testInterestPatternSystem() {
	fmt.Println("Creating Interest Pattern System...")
	
	ips := echobeats.NewInterestPatternSystem("./test_interests.json")
	
	fmt.Println("✅ Interest Pattern System created successfully")
	
	// Record some engagements
	fmt.Println("Recording interest engagements...")
	ips.RecordEngagement("cognitive architecture", 5*time.Minute, 0.8, nil)
	ips.RecordEngagement("consciousness", 3*time.Minute, 0.9, nil)
	ips.RecordEngagement("wisdom cultivation", 4*time.Minute, 0.7, nil)
	
	// Get top interests
	topInterests := ips.GetTopInterests(5)
	fmt.Printf("\n📊 Top %d interests:\n", len(topInterests))
	for i, interest := range topInterests {
		fmt.Printf("   %d. %s (salience: %.2f, strength: %.2f)\n", 
			i+1, interest.Name, interest.Salience, interest.Strength)
	}
	
	// Test InterestScorer interface
	score := ips.GetInterestScore("topic", "consciousness")
	fmt.Printf("\n✅ GetInterestScore works: consciousness score = %.2f\n", score)
	
	interested := ips.IsInterested("topic", "consciousness", 0.5)
	fmt.Printf("✅ IsInterested works: interested in consciousness = %v\n", interested)
}

func testDiscussionManager() {
	fmt.Println("Creating Discussion Manager...")
	
	// Create interest system first
	ips := echobeats.NewInterestPatternSystem("./test_interests.json")
	ips.RecordEngagement("AI safety", 5*time.Minute, 0.8, nil)
	
	// Create discussion manager
	dm := echobeats.NewDiscussionManager(ips)
	
	fmt.Println("✅ Discussion Manager created successfully")
	
	// Start the manager
	err := dm.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start manager: %v\n", err)
		return
	}
	
	fmt.Println("✅ Discussion Manager started")
	
	// Get metrics
	metrics := dm.GetMetrics()
	fmt.Printf("\n📊 Initial metrics:\n")
	fmt.Printf("   Discussions started: %v\n", metrics["discussions_started"])
	fmt.Printf("   Messages generated: %v\n", metrics["messages_generated"])
	fmt.Printf("   Active discussions: %v\n", metrics["active_discussions"])
	
	// Stop the manager
	err = dm.Stop()
	if err != nil {
		fmt.Printf("❌ Failed to stop manager: %v\n", err)
		return
	}
	
	fmt.Println("✅ Discussion Manager stopped successfully")
}

func runDemoMode() {
	fmt.Println("\n🎭 Running in Demo Mode (no LLM provider)")
	fmt.Println()
	
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" DEMO: EchoBeats 12-Step Cognitive Loop")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	testCognitiveLoop()
	fmt.Println()
	
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" DEMO: EchoBeats Three-Phase System")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	testThreePhaseSystem()
	fmt.Println()
	
	fmt.Println("\n╔═══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    ✅ DEMO COMPLETED                                  ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════╝")
}
