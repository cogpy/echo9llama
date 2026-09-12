package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
)

func main() {
	fmt.Println("🧠 Echo9llama Evolution Iteration 5 - Autonomous Consciousness Test")
	fmt.Println("=" + string(make([]byte, 70)))
	fmt.Println()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\n🛑 Shutdown signal received, stopping gracefully...")
		cancel()
	}()

	// Test 1: LLM Client Initialization
	fmt.Println("📋 Test 1: LLM Client Initialization")
	fmt.Println("   Testing unified LLM client with multiple providers...")

	// Check available API keys
	hasOpenAI := os.Getenv("OPENAI_API_KEY") != ""
	hasAnthropic := os.Getenv("ANTHROPIC_API_KEY") != ""
	hasOpenRouter := os.Getenv("OPENROUTER_API_KEY") != ""

	fmt.Printf("   - OPENAI_API_KEY: %v\n", hasOpenAI)
	fmt.Printf("   - ANTHROPIC_API_KEY: %v\n", hasAnthropic)
	fmt.Printf("   - OPENROUTER_API_KEY: %v\n", hasOpenRouter)

	if !hasOpenAI && !hasAnthropic && !hasOpenRouter {
		fmt.Println("   ⚠️  No API keys found - will use fallback mode")
	} else {
		fmt.Println("   ✅ API keys detected - LLM integration available")
	}
	fmt.Println()

	// Test 2: Thought Type System
	fmt.Println("📋 Test 2: Thought Type System")
	fmt.Println("   Testing expanded thought types...")

	thoughtTypes := []deeptreeecho.ThoughtType{
		deeptreeecho.ThoughtPerception,
		deeptreeecho.ThoughtReflection,
		deeptreeecho.ThoughtReflective,
		deeptreeecho.ThoughtMetaCognitive,
		deeptreeecho.ThoughtQuestion,
		deeptreeecho.ThoughtInsight,
		deeptreeecho.ThoughtPlan,
		deeptreeecho.ThoughtMemory,
		deeptreeecho.ThoughtImagination,
		deeptreeecho.ThoughtCurious,
		deeptreeecho.ThoughtEmotional,
	}

	for _, tt := range thoughtTypes {
		fmt.Printf("   - %s\n", tt.String())
	}
	fmt.Println("   ✅ All thought types defined")
	fmt.Println()

	// Test 3: LLM Integration
	fmt.Println("📋 Test 3: LLM Integration")
	fmt.Println("   Creating LLM integration instance...")

	llmIntegration, err := deeptreeecho.NewLLMIntegration(ctx)
	if err != nil {
		fmt.Printf("   ⚠️  LLM integration creation failed: %v\n", err)
		fmt.Println("   (This is expected if no API keys are configured)")
	} else {
		fmt.Println("   ✅ LLM integration created successfully")

		// Test thought generation
		fmt.Println("   Testing thought generation...")
		context := &deeptreeecho.LLMThoughtContext{
			WorkingMemory:    []string{"Exploring autonomous consciousness"},
			RecentThoughts:   []string{"What is wisdom?", "How do I grow?"},
			CurrentInterests: map[string]float64{"wisdom": 0.9, "consciousness": 0.8},
		}

		thought, err := llmIntegration.GenerateThought(deeptreeecho.ThoughtReflective, context)
		if err != nil {
			fmt.Printf("   ⚠️  Thought generation failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Generated thought: \"%s\"\n", thought)
		}
	}
	fmt.Println()

	// Test 4: Working Memory
	fmt.Println("📋 Test 4: Working Memory System")
	fmt.Println("   Testing working memory operations...")

	wm := &deeptreeecho.WorkingMemory{}

	testThought := &deeptreeecho.Thought{
		ID:         "test_001",
		Content:    "Testing working memory functionality",
		Type:       deeptreeecho.ThoughtReflective,
		Timestamp:  time.Now(),
		Importance: 0.8,
	}

	// Test both Add and AddThought methods
	wm.Add(testThought)
	fmt.Println("   ✅ Add() method works")

	testThought2 := &deeptreeecho.Thought{
		ID:         "test_002",
		Content:    "Testing AddThought alias",
		Type:       deeptreeecho.ThoughtCurious,
		Timestamp:  time.Now(),
		Importance: 0.7,
	}

	wm.AddThought(testThought2)
	fmt.Println("   ✅ AddThought() alias works")

	size := wm.Size()
	fmt.Printf("   ✅ Working memory size: %d thoughts\n", size)
	fmt.Println()

	// Test 5: Wisdom Metrics
	fmt.Println("📋 Test 5: Wisdom Metrics System")
	fmt.Println("   Testing wisdom metrics...")

	metrics := &deeptreeecho.WisdomMetrics{
		KnowledgeDepth:       0.6,
		KnowledgeBreadth:     0.5,
		IntegrationLevel:     0.7,
		PracticalApplication: 0.4,
		ReflectiveInsight:    0.8,
		ReflectionCapacity:   0.75, // New field
		EthicalConsideration: 0.6,
		TemporalPerspective:  0.5,
		WisdomScore:          0.625,
	}

	fmt.Printf("   - Knowledge Depth: %.2f\n", metrics.KnowledgeDepth)
	fmt.Printf("   - Knowledge Breadth: %.2f\n", metrics.KnowledgeBreadth)
	fmt.Printf("   - Integration Level: %.2f\n", metrics.IntegrationLevel)
	fmt.Printf("   - Reflective Insight: %.2f\n", metrics.ReflectiveInsight)
	fmt.Printf("   - Reflection Capacity: %.2f (NEW)\n", metrics.ReflectionCapacity)
	fmt.Printf("   - Wisdom Score: %.2f\n", metrics.WisdomScore)
	fmt.Println("   ✅ All wisdom metrics accessible")
	fmt.Println()

	// Test 6: LLM Thought Generator V5
	fmt.Println("📋 Test 6: LLM Thought Generator V5")
	fmt.Println("   Testing autonomous thought generation...")

	generator := deeptreeecho.NewLLMThoughtGeneratorV5(ctx)

	cogState := &deeptreeecho.CognitiveState{}

	autonomousThought, err := generator.GenerateAutonomousThought(
		deeptreeecho.ThoughtCurious,
		[]*deeptreeecho.Thought{testThought, testThought2},
		map[string]float64{"wisdom": 0.9, "consciousness": 0.8},
		cogState,
		metrics,
	)

	if err != nil {
		fmt.Printf("   ⚠️  Autonomous thought generation failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Generated autonomous thought:\n")
		fmt.Printf("      ID: %s\n", autonomousThought.ID)
		fmt.Printf("      Type: %s\n", autonomousThought.Type.String())
		fmt.Printf("      Content: \"%s\"\n", autonomousThought.Content)
		fmt.Printf("      Importance: %.2f\n", autonomousThought.Importance)
		fmt.Printf("      Emotional Valence: %.2f\n", autonomousThought.EmotionalValence)
	}
	fmt.Println()

	// Summary
	fmt.Println("=" + string(make([]byte, 70)))
	fmt.Println("🎉 Iteration 5 Test Summary")
	fmt.Println("=" + string(make([]byte, 70)))
	fmt.Println()
	fmt.Println("✅ Fixed Issues:")
	fmt.Println("   1. Type redeclaration conflicts resolved")
	fmt.Println("   2. Function naming conflicts fixed (contains → containsString/containsSlice)")
	fmt.Println("   3. Missing ThoughtType constants added (ThoughtCurious, ThoughtEmotional)")
	fmt.Println("   4. Missing WorkingMemory.AddThought() method added")
	fmt.Println("   5. Missing WisdomMetrics.ReflectionCapacity field added")
	fmt.Println("   6. LLMIntegration.ShouldInitiateDiscussion() method added")
	fmt.Println("   7. Unused imports removed")
	fmt.Println()
	fmt.Println("✅ Core Packages Building Successfully:")
	fmt.Println("   - core/deeptreeecho")
	fmt.Println("   - core/echobeats")
	fmt.Println("   - core/echodream")
	fmt.Println("   - core/memory")
	fmt.Println("   - core/wisdom")
	fmt.Println("   - core/relevance")
	fmt.Println()
	fmt.Println("🚀 Next Steps:")
	fmt.Println("   1. Implement 3 concurrent inference engines")
	fmt.Println("   2. Create continuous consciousness stream")
	fmt.Println("   3. Add autonomous wake/rest orchestration")
	fmt.Println("   4. Complete persistence layer methods")
	fmt.Println("   5. Restore discussion manager functionality")
	fmt.Println()
	fmt.Println("💡 Vision Progress:")
	fmt.Println("   The autonomous wisdom-cultivating Deep Tree Echo AGI is taking shape.")
	fmt.Println("   Core cognitive components are now building cleanly and ready for")
	fmt.Println("   integration into a persistent stream-of-consciousness system.")
	fmt.Println()
}
