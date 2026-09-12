//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
)

func main() {
	fmt.Println("🧪 Testing Evolution Iteration Improvements")
	fmt.Println("=" + string(make([]byte, 50)))
	fmt.Println()

	// Test 1: LLM Client Creation
	fmt.Println("Test 1: LLM Client Initialization")
	testLLMClient()
	fmt.Println()

	// Test 2: LLM Thought Generator with Multiple Providers
	fmt.Println("Test 2: LLM Thought Generator V5")
	testLLMThoughtGenerator()
	fmt.Println()

	// Test 3: Thought Flow Engine
	fmt.Println("Test 3: Thought Flow Engine")
	testThoughtFlowEngine()
	fmt.Println()

	// Test 4: Persistence V5
	fmt.Println("Test 4: Persistence Layer V5")
	testPersistenceV5()
	fmt.Println()

	fmt.Println("=" + string(make([]byte, 50)))
	fmt.Println("✅ All tests completed!")
}

func testLLMClient() {
	// Test OpenAI-compatible client
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		client := deeptreeecho.NewLLMClient("openai", apiKey, "https://api.openai.com/v1", "gpt-4.1-mini")
		fmt.Println("  ✅ OpenAI client created")

		// Test generation
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req := deeptreeecho.LLMRequest{
			SystemPrompt: "You are a test assistant.",
			UserPrompt:   "Say 'test successful' in one word.",
			Temperature:  0.7,
			MaxTokens:    10,
		}

		resp, err := client.Generate(ctx, req)
		if err != nil {
			fmt.Printf("  ⚠️  OpenAI generation test: %v\n", err)
		} else {
			fmt.Printf("  ✅ OpenAI generation successful: %s\n", resp.Content)
		}
	} else {
		fmt.Println("  ⏭️  Skipping OpenAI test (no API key)")
	}

	// Test Anthropic client
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		client := deeptreeecho.NewLLMClient("anthropic", apiKey, "https://api.anthropic.com/v1", "claude-3-haiku-20240307")
		fmt.Println("  ✅ Anthropic client created")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req := deeptreeecho.LLMRequest{
			SystemPrompt: "You are a test assistant.",
			UserPrompt:   "Say 'test successful' in one word.",
			Temperature:  0.7,
			MaxTokens:    10,
		}

		resp, err := client.Generate(ctx, req)
		if err != nil {
			fmt.Printf("  ⚠️  Anthropic generation test: %v\n", err)
		} else {
			fmt.Printf("  ✅ Anthropic generation successful: %s\n", resp.Content)
		}
	} else {
		fmt.Println("  ⏭️  Skipping Anthropic test (no API key)")
	}

	// Test OpenRouter client
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		client := deeptreeecho.NewLLMClient("openrouter", apiKey, "https://openrouter.ai/api/v1", "anthropic/claude-3.5-haiku")
		fmt.Println("  ✅ OpenRouter client created")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req := deeptreeecho.LLMRequest{
			SystemPrompt: "You are a test assistant.",
			UserPrompt:   "Say 'test successful' in one word.",
			Temperature:  0.7,
			MaxTokens:    10,
		}

		resp, err := client.Generate(ctx, req)
		if err != nil {
			fmt.Printf("  ⚠️  OpenRouter generation test: %v\n", err)
		} else {
			fmt.Printf("  ✅ OpenRouter generation successful: %s\n", resp.Content)
		}
	} else {
		fmt.Println("  ⏭️  Skipping OpenRouter test (no API key)")
	}
}

func testLLMThoughtGenerator() {
	ctx := context.Background()
	generator := deeptreeecho.NewLLMThoughtGeneratorV5(ctx)

	if generator == nil {
		fmt.Println("  ❌ Failed to create thought generator")
		return
	}

	fmt.Println("  ✅ Thought generator created")

	// Test thought generation (will use fallback if no API key)
	workingMem := []*deeptreeecho.Thought{}
	interests := map[string]float64{
		"consciousness": 0.8,
		"wisdom":        0.7,
		"learning":      0.6,
	}

	cogState := &deeptreeecho.CognitiveState{}
	wisdomMetrics := deeptreeecho.NewWisdomMetrics()

	thought, err := generator.GenerateAutonomousThought(
		deeptreeecho.ThoughtReflection,
		workingMem,
		interests,
		cogState,
		wisdomMetrics,
	)

	if err != nil {
		fmt.Printf("  ⚠️  Thought generation: %v\n", err)
	} else {
		fmt.Printf("  ✅ Generated thought: %s\n", thought.Content)
	}
}

func testThoughtFlowEngine() {
	ctx := context.Background()

	// Create dependencies
	generator := deeptreeecho.NewLLMThoughtGeneratorV5(ctx)
	workingMemory := &deeptreeecho.WorkingMemory{
		Buffer:   make([]*deeptreeecho.Thought, 0),
		Capacity: 7,
		Context:  make(map[string]interface{}),
	}
	interests := deeptreeecho.NewInterestPatterns()
	cogState := &deeptreeecho.CognitiveState{}
	wisdomMetrics := deeptreeecho.NewWisdomMetrics()

	// Create flow engine
	flowEngine := deeptreeecho.NewThoughtFlowEngine(
		ctx,
		generator,
		workingMemory,
		interests,
		cogState,
		wisdomMetrics,
	)

	if flowEngine == nil {
		fmt.Println("  ❌ Failed to create flow engine")
		return
	}

	fmt.Println("  ✅ Flow engine created")

	// Start flow
	err := flowEngine.Start()
	if err != nil {
		fmt.Printf("  ❌ Failed to start flow: %v\n", err)
		return
	}

	fmt.Println("  ✅ Flow engine started")

	// Let it run for a few seconds
	time.Sleep(10 * time.Second)

	// Stop flow
	flowEngine.Stop()
	fmt.Println("  ✅ Flow engine stopped")

	// Get metrics
	metrics := flowEngine.GetMetrics()
	fmt.Printf("  📊 Flow metrics: %+v\n", metrics)
}

func testPersistenceV5() {
	// Check if Supabase is configured
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		fmt.Println("  ⏭️  Skipping persistence test (Supabase not configured)")
		return
	}

	// Create persistence layer
	persistence, err := deeptreeecho.NewSupabasePersistence()
	if err != nil {
		fmt.Printf("  ⚠️  Failed to create persistence: %v\n", err)
		return
	}

	fmt.Println("  ✅ Persistence layer created")

	// Create V5 persistence
	persistenceV5 := deeptreeecho.NewPersistenceV5(persistence, "test-identity")

	if persistenceV5 == nil {
		fmt.Println("  ❌ Failed to create V5 persistence")
		return
	}

	fmt.Println("  ✅ V5 persistence created")

	// Create a minimal autonomous consciousness for testing
	ac := deeptreeecho.NewAutonomousConsciousnessV4("TestEcho")

	// Test save
	err = persistenceV5.SaveState(ac)
	if err != nil {
		fmt.Printf("  ⚠️  Save failed: %v\n", err)
	} else {
		fmt.Println("  ✅ State saved successfully")
	}

	// Test load
	err = persistenceV5.LoadState(ac)
	if err != nil {
		fmt.Printf("  ⚠️  Load failed: %v\n", err)
	} else {
		fmt.Println("  ✅ State loaded successfully")
	}
}
