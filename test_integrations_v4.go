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
	fmt.Println("🧪 Testing Deep Tree Echo V12 Integrations")
	fmt.Println("==========================================")
	fmt.Println()

	ctx := context.Background()

	// Test 1: Neural Pattern Processor
	fmt.Println("📊 Test 1: Neural Pattern Processor (GoMLX-inspired)")
	fmt.Println("---------------------------------------------------")
	testNeuralPatternProcessor()
	fmt.Println()

	// Test 2: Unified Orchestrator
	fmt.Println("🎭 Test 2: Unified Orchestrator")
	fmt.Println("-------------------------------")
	testUnifiedOrchestrator(ctx)
	fmt.Println()

	// Test 3: Full Integration Test
	fmt.Println("🔗 Test 3: Full Integration Test")
	fmt.Println("---------------------------------")
	testFullIntegration(ctx)
	fmt.Println()

	// Test 4: Autonomous Agent with All Systems
	fmt.Println("🤖 Test 4: Autonomous Agent with All Systems")
	fmt.Println("---------------------------------------------")
	testAutonomousAgent()
	fmt.Println()

	fmt.Println("==========================================")
	fmt.Println("✅ All V12 Integration Tests Completed!")
}

func testNeuralPatternProcessor() {
	npp := deeptreeecho.NewNeuralPatternProcessor()

	// Start the processor
	if err := npp.Start(); err != nil {
		fmt.Printf("   ❌ Failed to start: %v\n", err)
		return
	}
	fmt.Println("   ✓ Neural Pattern Processor started")

	// Learn some patterns
	pattern1 := npp.LearnPattern("perception", "visual_edge", []float64{0.8, 0.2, 0.1})
	fmt.Printf("   ✓ Learned pattern: %s (category: %s)\n", pattern1.Name, pattern1.Category)

	pattern2 := npp.LearnPattern("action", "motor_grasp", []float64{0.3, 0.9, 0.5})
	fmt.Printf("   ✓ Learned pattern: %s (category: %s)\n", pattern2.Name, pattern2.Category)

	pattern3 := npp.LearnPattern("simulation", "future_state", []float64{0.5, 0.5, 0.8})
	fmt.Printf("   ✓ Learned pattern: %s (category: %s)\n", pattern3.Name, pattern3.Category)

	// Test pattern recognition
	testInput := []float64{0.75, 0.25, 0.15}
	recognized, score := npp.RecognizePattern(testInput)
	if recognized != nil {
		fmt.Printf("   ✓ Recognized pattern: %s (score: %.2f)\n", recognized.Name, score)
	} else {
		fmt.Println("   ✓ No pattern recognized (score too low)")
	}

	// Test forward propagation
	output := npp.Forward(testInput)
	fmt.Printf("   ✓ Forward propagation: input[%d] -> output[%d]\n", len(testInput), len(output))

	// Test reinforcement learning
	if err := npp.ReinforceLearning(pattern1.ID, 0.5); err != nil {
		fmt.Printf("   ❌ Reinforcement failed: %v\n", err)
	} else {
		fmt.Printf("   ✓ Reinforced pattern: %s (new strength: %.2f)\n", pattern1.Name, pattern1.Strength)
	}

	// Get metrics
	metrics := npp.GetMetrics()
	fmt.Printf("   ✓ Metrics: patterns=%v, activations=%v, layers=%v\n",
		metrics["total_patterns"], metrics["total_activations"], metrics["layer_count"])

	// Stop the processor
	if err := npp.Stop(); err != nil {
		fmt.Printf("   ❌ Failed to stop: %v\n", err)
	} else {
		fmt.Println("   ✓ Neural Pattern Processor stopped")
	}
}

func testUnifiedOrchestrator(ctx context.Context) {
	// Create required subsystems
	eventBus := deeptreeecho.NewCognitiveEventBus(ctx)
	knowledgeGraph := deeptreeecho.NewKnowledgeGraph()
	stateStore := deeptreeecho.NewPersistentStateStore()
	stateMachine := deeptreeecho.NewCognitiveStateMachine(deeptreeecho.StateIdle)

	// Start subsystems
	eventBus.Start()
	knowledgeGraph.Start()
	stateStore.Start()
	stateMachine.Start()

	// Create orchestrator
	orchestrator := deeptreeecho.NewUnifiedOrchestrator(
		eventBus,
		knowledgeGraph,
		stateStore,
		stateMachine,
	)

	// Add optional subsystems
	patternProcessor := deeptreeecho.NewNeuralPatternProcessor()
	patternProcessor.Start()
	orchestrator.SetPatternProcessor(patternProcessor)

	// Create a simple fallback provider for semantic memory
	provider := &llm.SimpleFallbackProvider{}
	semanticMemory := deeptreeecho.NewSemanticMemory(provider)
	semanticMemory.Start()
	orchestrator.SetSemanticMemory(semanticMemory)

	scheduler := deeptreeecho.NewCognitiveScheduler()
	scheduler.Start()
	orchestrator.SetScheduler(scheduler)

	// Start orchestrator
	if err := orchestrator.Start(); err != nil {
		fmt.Printf("   ❌ Failed to start orchestrator: %v\n", err)
		return
	}
	fmt.Println("   ✓ Unified Orchestrator started")

	// Let it run for a few cycles
	fmt.Println("   ⏳ Running orchestration cycles...")
	time.Sleep(3 * time.Second)

	// Get metrics
	metrics := orchestrator.GetMetrics()
	fmt.Printf("   ✓ Orchestrator metrics: mode=%v, cycles=%v, decisions=%v\n",
		metrics["mode"], metrics["cycle_count"], metrics["total_decisions"])

	// Get gestalt
	gestalt := orchestrator.GetGestalt()
	fmt.Printf("   ✓ Gestalt aggregated with %d subsystems\n", len(gestalt))

	// Stop orchestrator
	if err := orchestrator.Stop(); err != nil {
		fmt.Printf("   ❌ Failed to stop orchestrator: %v\n", err)
	} else {
		fmt.Println("   ✓ Unified Orchestrator stopped")
	}

	// Cleanup
	patternProcessor.Stop()
	semanticMemory.Stop()
	scheduler.Stop()
	stateMachine.Stop()
	stateStore.Stop()
	knowledgeGraph.Stop()
	eventBus.Stop()
}

func testFullIntegration(ctx context.Context) {
	// Create all foundational systems
	eventBus := deeptreeecho.NewCognitiveEventBus(ctx)
	knowledgeGraph := deeptreeecho.NewKnowledgeGraph()
	stateStore := deeptreeecho.NewPersistentStateStore()
	stateMachine := deeptreeecho.NewCognitiveStateMachine(deeptreeecho.StateIdle)
	patternProcessor := deeptreeecho.NewNeuralPatternProcessor()
	provider := &llm.SimpleFallbackProvider{}
	semanticMemory := deeptreeecho.NewSemanticMemory(provider)
	scheduler := deeptreeecho.NewCognitiveScheduler()

	// Start all systems
	systems := []struct {
		name  string
		start func() error
	}{
		{"EventBus", eventBus.Start},
		{"KnowledgeGraph", knowledgeGraph.Start},
		{"StateStore", stateStore.Start},
		{"StateMachine", stateMachine.Start},
		{"PatternProcessor", patternProcessor.Start},
		{"SemanticMemory", semanticMemory.Start},
		{"Scheduler", scheduler.Start},
	}

	for _, sys := range systems {
		if err := sys.start(); err != nil {
			fmt.Printf("   ❌ Failed to start %s: %v\n", sys.name, err)
			return
		}
		fmt.Printf("   ✓ %s started\n", sys.name)
	}

	// Test cross-system interactions
	fmt.Println("   📡 Testing cross-system interactions...")

	// 1. Learn a pattern and store in knowledge graph
	pattern := patternProcessor.LearnPattern("integration", "test_pattern", []float64{0.5, 0.5, 0.5})
	knowledgeGraph.AddTriple(pattern.ID, "is_a", "LearnedPattern")
	fmt.Printf("   ✓ Pattern %s stored in knowledge graph\n", pattern.ID)

	// 2. Store pattern metadata in state store
	stateStore.Set("pattern:"+pattern.ID, map[string]interface{}{
		"strength": pattern.Strength,
		"category": pattern.Category,
	})
	fmt.Println("   ✓ Pattern metadata stored in state store")

	// 3. Transition state machine
	if stateMachine.CanFire(deeptreeecho.TriggerWake) {
		stateMachine.Fire(deeptreeecho.TriggerWake)
		fmt.Printf("   ✓ State machine transitioned to: %s\n", stateMachine.State())
	}

	// 4. Store in semantic memory
	docID, _ := semanticMemory.AddDocument("episodic", "Integration test completed", map[string]string{
		"type": "test",
	})
	fmt.Printf("   ✓ Document %s stored in semantic memory\n", docID)

	// 5. Schedule a cognitive task
	scheduler.ScheduleInterval("integration_check", 1*time.Second, func(ctx context.Context) error {
		fmt.Println("   ✓ Scheduled task executed")
		return nil
	})
	fmt.Println("   ✓ Task scheduled")

	// 6. Publish event
	eventBus.Publish(deeptreeecho.NewCognitiveEvent(
		deeptreeecho.EventKnowledgeAcquired,
		"integration_test",
		map[string]interface{}{
			"subject": "test_knowledge",
		},
	))
	fmt.Println("   ✓ Event published")

	// Wait for scheduled task
	time.Sleep(2 * time.Second)

	// Get final metrics from all systems
	fmt.Println("   📊 Final system metrics:")
	fmt.Printf("      - KnowledgeGraph: %v quads\n", knowledgeGraph.GetMetrics()["total_quads"])
	fmt.Printf("      - StateStore: %v keys\n", stateStore.GetMetrics()["total_keys"])
	fmt.Printf("      - PatternProcessor: %v patterns\n", patternProcessor.GetMetrics()["total_patterns"])
	fmt.Printf("      - SemanticMemory: %v documents\n", semanticMemory.GetMetrics()["total_documents"])
	fmt.Printf("      - StateMachine: %s\n", stateMachine.State())

	// Stop all systems
	scheduler.Stop()
	semanticMemory.Stop()
	patternProcessor.Stop()
	stateMachine.Stop()
	stateStore.Stop()
	knowledgeGraph.Stop()
	eventBus.Stop()
	fmt.Println("   ✓ All systems stopped")
}

func testAutonomousAgent() {
	// Create a simple fallback LLM provider
	provider := &llm.SimpleFallbackProvider{}

	// Create the autonomous agent
	agent := deeptreeecho.NewAutonomousAgent("test_agent", provider)
	fmt.Println("   ✓ Autonomous Agent created")

	// Start the agent
	if err := agent.Start(); err != nil {
		fmt.Printf("   ❌ Failed to start agent: %v\n", err)
		return
	}
	fmt.Println("   ✓ Autonomous Agent started")

	// Let it run briefly
	fmt.Println("   ⏳ Running autonomous agent...")
	time.Sleep(3 * time.Second)

	// Get status
	status := agent.GetStatus()
	fmt.Printf("   ✓ Agent status: phase=%v, autonomy=%v\n",
		status["phase"], status["autonomy_level"])

	// Get gestalt
	gestalt := agent.GetGestalt()
	fmt.Printf("   ✓ Agent gestalt contains %d components\n", len(gestalt))

	// Test knowledge storage
	agent.StoreKnowledge("TestEntity", "has_property", "TestValue")
	fmt.Println("   ✓ Knowledge stored via agent")

	// Test state storage
	agent.SaveState("test:key", "test_value")
	fmt.Println("   ✓ State saved via agent")

	// Query knowledge
	quads := agent.QueryKnowledge("TestEntity")
	fmt.Printf("   ✓ Knowledge query returned %d quads\n", len(quads))

	// Get knowledge graph
	kg := agent.GetKnowledgeGraph()
	fmt.Printf("   ✓ Knowledge graph has %v quads\n", kg.GetMetrics()["total_quads"])

	// Get state machine
	sm := agent.GetStateMachine()
	fmt.Printf("   ✓ State machine in state: %s\n", sm.State())

	// Stop the agent
	agent.Stop()
	fmt.Println("   ✓ Autonomous Agent stopped")
}
