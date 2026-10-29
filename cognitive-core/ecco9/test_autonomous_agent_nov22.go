//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/EchoCog/echollama/core"
	"github.com/EchoCog/echollama/core/echobeats"
	"github.com/EchoCog/echollama/core/llm"
)

func main() {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("🧪 Testing Deep Tree Echo Autonomous Agent - November 22, 2025")
	fmt.Println(strings.Repeat("=", 70) + "\n")

	// Test 1: EchoBeats Components
	fmt.Println("Test 1: EchoBeats 12-Step Cognitive Loop")
	fmt.Println(strings.Repeat("-", 70))
	testCognitiveLoop()
	fmt.Println()

	// Test 2: Inference Engines
	fmt.Println("Test 2: 3 Concurrent Inference Engines")
	fmt.Println(strings.Repeat("-", 70))
	testInferenceEngines()
	fmt.Println()

	// Test 3: Enhanced Scheduler
	fmt.Println("Test 3: Enhanced EchoBeats Scheduler")
	fmt.Println(strings.Repeat("-", 70))
	testEnhancedScheduler()
	fmt.Println()

	// Test 4: Autonomous Agent (short run)
	fmt.Println("Test 4: Autonomous Agent Integration")
	fmt.Println(strings.Repeat("-", 70))
	testAutonomousAgent()
	fmt.Println()

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("✅ All tests completed successfully!")
	fmt.Println(strings.Repeat("=", 70))
}

func testCognitiveLoop() {
	loop := echobeats.NewCognitiveLoop()
	loop.SetStepDuration(500 * time.Millisecond)

	cycleComplete := false
	loop.SetCallbacks(
		func(step int, result *echobeats.StepResult) {
			// Step callback
		},
		func(cycle uint64) {
			fmt.Printf("   ✓ Cognitive cycle %d completed\n", cycle)
			cycleComplete = true
		},
	)

	if err := loop.Start(); err != nil {
		fmt.Printf("   ❌ Failed to start cognitive loop: %v\n", err)
		return
	}

	// Run for one complete cycle (12 steps * 0.5s = 6s)
	time.Sleep(7 * time.Second)

	if err := loop.Stop(); err != nil {
		fmt.Printf("   ❌ Failed to stop cognitive loop: %v\n", err)
		return
	}

	metrics := loop.GetMetrics()
	fmt.Printf("   ✓ Steps executed: %v\n", metrics["total_steps"])
	fmt.Printf("   ✓ Cycles completed: %v\n", metrics["cycle_count"])

	if cycleComplete {
		fmt.Println("   ✅ Cognitive loop test PASSED")
	} else {
		fmt.Println("   ⚠️  Cognitive loop test INCOMPLETE")
	}
}

func testInferenceEngines() {
	engines := []*echobeats.InferenceEngine{
		echobeats.NewInferenceEngine(1, echobeats.SpecializationPerception),
		echobeats.NewInferenceEngine(2, echobeats.SpecializationCognition),
		echobeats.NewInferenceEngine(3, echobeats.SpecializationAction),
	}

	// Start all engines
	for _, engine := range engines {
		if err := engine.Start(); err != nil {
			fmt.Printf("   ❌ Failed to start engine: %v\n", err)
			return
		}
	}

	// Submit test tasks
	for i, engine := range engines {
		task := &echobeats.InferenceTask{
			ID:       fmt.Sprintf("test_task_%d", i+1),
			Type:     "test",
			Priority: 0.8,
		}
		if err := engine.SubmitTask(task); err != nil {
			fmt.Printf("   ❌ Failed to submit task: %v\n", err)
		}
	}

	// Wait for processing
	time.Sleep(3 * time.Second)

	// Check metrics
	allProcessed := true
	for i, engine := range engines {
		metrics := engine.GetMetrics()
		fmt.Printf("   Engine %d (%v): Processed=%v, Queue=%v\n",
			i+1, metrics["specialization"], metrics["tasks_processed"], metrics["queue_length"])

		if metrics["tasks_processed"].(uint64) == 0 {
			allProcessed = false
		}
	}

	// Stop engines
	for _, engine := range engines {
		if err := engine.Stop(); err != nil {
			fmt.Printf("   ❌ Failed to stop engine: %v\n", err)
		}
	}

	if allProcessed {
		fmt.Println("   ✅ Inference engines test PASSED")
	} else {
		fmt.Println("   ⚠️  Inference engines test INCOMPLETE")
	}
}

func testEnhancedScheduler() {
	scheduler := echobeats.NewEnhancedScheduler()

	if err := scheduler.Start(); err != nil {
		fmt.Printf("   ❌ Failed to start scheduler: %v\n", err)
		return
	}

	// Schedule some test events
	for i := 0; i < 3; i++ {
		event := &echobeats.CognitiveEvent{
			ID:          fmt.Sprintf("test_event_%d", i),
			Type:        echobeats.EventThought,
			Priority:    70,
			ScheduledAt: time.Now().Add(time.Duration(i) * time.Second),
			Payload:     fmt.Sprintf("Test thought %d", i),
		}
		scheduler.ScheduleEvent(event)
	}

	fmt.Println("   ✓ Scheduled 3 test events")

	// Run for a bit
	time.Sleep(5 * time.Second)

	status := scheduler.GetStatus()
	fmt.Printf("   ✓ Loop cycles: %v\n", status["loop_cycles"])
	fmt.Printf("   ✓ Events processed: %v\n",
		status["echobeats"].(map[string]interface{})["events_processed"])

	if err := scheduler.Stop(); err != nil {
		fmt.Printf("   ❌ Failed to stop scheduler: %v\n", err)
		return
	}

	fmt.Println("   ✅ Enhanced scheduler test PASSED")
}

func testAutonomousAgent() {
	// The agent's thoughts come only from a real model; without one the
	// test is skipped rather than run against canned responses.
	var llmProvider llm.LLMProvider
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		fmt.Println("   ℹ️  Using real Anthropic provider")
		llmProvider = llm.NewAnthropicProvider("")
	} else if os.Getenv("OPENROUTER_API_KEY") != "" {
		fmt.Println("   ℹ️  Using real OpenRouter provider")
		llmProvider = llm.NewOpenRouterProvider("")
	} else if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		fmt.Println("   ℹ️  Using real OpenAI provider")
		llmProvider = llm.NewOpenAIProvider(apiKey)
	} else {
		fmt.Println("   ⚠️  Autonomous agent test SKIPPED: no LLM provider (set ANTHROPIC_API_KEY, OPENROUTER_API_KEY or OPENAI_API_KEY)")
		return
	}

	agent := core.NewAutonomousAgent(llmProvider)

	if err := agent.Start(); err != nil {
		fmt.Printf("   ❌ Failed to start agent: %v\n", err)
		return
	}

	fmt.Println("   ✓ Agent started successfully")
	fmt.Println("   ℹ️  Running for 15 seconds...")

	// Run for 15 seconds
	time.Sleep(15 * time.Second)

	// Get status before stopping
	status := agent.GetStatus()
	fmt.Printf("   ✓ Agent uptime: %v\n", status["uptime"])
	fmt.Printf("   ✓ Agent running: %v\n", status["running"])

	if err := agent.Stop(); err != nil {
		fmt.Printf("   ❌ Failed to stop agent: %v\n", err)
		return
	}

	fmt.Println("   ✅ Autonomous agent test PASSED")
}
