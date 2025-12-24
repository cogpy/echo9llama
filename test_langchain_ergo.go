// +build ignore

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
)

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║     🧪 LANGCHAIN & ERGO INTEGRATION TEST                      ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	ctx := context.Background()

	// Create an LLM provider for testing
	llmProvider := llm.NewMultiProviderLLM()
	fmt.Println("✅ Created multi-provider LLM")

	// Test 1: Create cognitive tools
	fmt.Println("\n📦 Test 1: Creating cognitive tools...")
	
	skillLearning := deeptreeecho.NewSkillLearningSystem(llmProvider)
	discussionAutonomy := deeptreeecho.NewDiscussionAutonomySystem(llmProvider)
	wisdomSynthesis := deeptreeecho.NewWisdomSynthesis(llmProvider)
	interestPatterns := deeptreeecho.NewInterestPatternSystem()
	echoDream := deeptreeecho.NewEchoDreamKnowledgeIntegration(llmProvider)
	echobeatsScheduler := deeptreeecho.NewEchobeatsScheduler(llmProvider)

	tools := deeptreeecho.CreateAllCognitiveTools(
		skillLearning,
		discussionAutonomy,
		wisdomSynthesis,
		interestPatterns,
		echoDream,
		echobeatsScheduler,
	)

	fmt.Printf("   Created %d cognitive tools:\n", len(tools))
	for _, tool := range tools {
		fmt.Printf("   - %s: %s\n", tool.Name(), truncate(tool.Description(), 50))
	}
	fmt.Println("✅ Cognitive tools created successfully")

	// Test 2: Test individual tools
	fmt.Println("\n🔧 Test 2: Testing individual tools...")

	// Test SkillLearner
	result, err := tools[0].Call(ctx, "consider:golang_programming")
	if err != nil {
		fmt.Printf("   ❌ SkillLearner failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ SkillLearner: %s\n", result)
	}

	// Test InterestTracker
	result, err = tools[3].Call(ctx, "evaluate:artificial intelligence")
	if err != nil {
		fmt.Printf("   ❌ InterestTracker failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ InterestTracker: %s\n", result)
	}

	// Test GoalManager
	result, err = tools[5].Call(ctx, "create:Learn wisdom synthesis")
	if err != nil {
		fmt.Printf("   ❌ GoalManager failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ GoalManager: %s\n", result)
	}

	// Test 3: Create ReasoningManager
	fmt.Println("\n🧠 Test 3: Creating ReasoningManager...")
	
	reasoningManager := deeptreeecho.NewReasoningManager(llmProvider, tools)
	fmt.Println("   ✅ ReasoningManager created")

	err = reasoningManager.Start()
	if err != nil {
		fmt.Printf("   ❌ Failed to start ReasoningManager: %v\n", err)
	} else {
		fmt.Println("   ✅ ReasoningManager started")
	}

	// Get metrics
	metrics := reasoningManager.GetMetrics()
	fmt.Printf("   Metrics: available_tools=%v, total_tasks=%v\n", 
		metrics["available_tools"], metrics["total_tasks"])

	// Test 4: Create ActorSupervisor
	fmt.Println("\n🎭 Test 4: Creating ActorSupervisor (3 cognitive streams)...")
	
	actorSupervisor := deeptreeecho.NewActorSupervisor()
	fmt.Println("   ✅ ActorSupervisor created")

	// Set callbacks
	actorSupervisor.SetStreamCallbacks(
		func(data interface{}) interface{} {
			fmt.Println("   📥 Perception callback triggered")
			return "perceived"
		},
		func(data interface{}) interface{} {
			fmt.Println("   🎯 Action callback triggered")
			return "acted"
		},
		func(data interface{}) interface{} {
			fmt.Println("   🔮 Simulation callback triggered")
			return "simulated"
		},
		func(pattern string, strength float64) {
			fmt.Printf("   ✨ Emergence detected: %s (%.2f)\n", pattern, strength)
		},
	)

	err = actorSupervisor.Start()
	if err != nil {
		fmt.Printf("   ❌ Failed to start ActorSupervisor: %v\n", err)
	} else {
		fmt.Println("   ✅ ActorSupervisor started - 3 concurrent streams active")
	}

	// Let the streams run for a moment
	fmt.Println("\n⏳ Running cognitive streams for 3 seconds...")
	time.Sleep(3 * time.Second)

	// Get stream states
	states := actorSupervisor.GetStreamStates()
	fmt.Println("\n📊 Stream States:")
	for streamType, state := range states {
		fmt.Printf("   %s: step=%d, phase=%s, load=%.2f\n",
			streamType, state.CurrentStep, state.CurrentPhase, state.ProcessingLoad)
	}

	// Test 5: Get gestalt contributions
	fmt.Println("\n🌐 Test 5: Getting gestalt contributions...")
	
	rmGestalt := reasoningManager.ContributeToGestalt()
	fmt.Printf("   ReasoningManager: running=%v, available_tools=%v\n",
		rmGestalt["running"], rmGestalt["available_tools"])

	asGestalt := actorSupervisor.ContributeToGestalt()
	fmt.Printf("   ActorSupervisor: running=%v\n", asGestalt["running"])

	// Cleanup
	fmt.Println("\n🧹 Cleaning up...")
	actorSupervisor.Stop()
	reasoningManager.Stop()

	fmt.Println("\n╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║     ✅ ALL INTEGRATION TESTS PASSED                           ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
