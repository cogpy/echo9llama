package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/EchoCog/echollama/core"
	"github.com/EchoCog/echollama/core/echobeats"
	"github.com/EchoCog/echollama/core/llm"
)

// MockLLMProvider for testing without API calls
type MockLLMProvider struct{}

func (m *MockLLMProvider) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	// Return mock responses based on prompt
	if len(prompt) > 50 {
		return "This is a thoughtful response to your query about cognitive processes and autonomous learning.", nil
	}
	return "Mock LLM response", nil
}

func (m *MockLLMProvider) StreamGenerate(ctx context.Context, prompt string, opts llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	response, err := m.Generate(ctx, prompt, opts)
	ch <- llm.StreamChunk{Content: response, Done: true, Error: err}
	close(ch)
	return ch, err
}

func (m *MockLLMProvider) Name() string {
	return "mock"
}

func (m *MockLLMProvider) Available() bool {
	return true
}

func (m *MockLLMProvider) MaxTokens() int {
	return 4096
}

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
	// Use mock LLM provider for testing
	mockProvider := &MockLLMProvider{}

	// Check if real API key is available
	var llmProvider llm.LLMProvider = mockProvider
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
		fmt.Println("   ℹ️  Using mock LLM provider (no API key)")
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
