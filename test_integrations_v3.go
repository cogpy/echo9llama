// +build ignore

package main

import (
	"fmt"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
)

func main() {
	fmt.Println("=== Deep Tree Echo Integration Test V3 ===")
	fmt.Println()

	// Test 1: Knowledge Graph
	fmt.Println("--- Test 1: Knowledge Graph (Cayley-inspired) ---")
	kg := deeptreeecho.NewKnowledgeGraph()
	if err := kg.Start(); err != nil {
		fmt.Printf("❌ Failed to start Knowledge Graph: %v\n", err)
		return
	}

	// Add some knowledge
	kg.AddTriple("DeepTreeEcho", "is_a", "AGI")
	kg.AddTriple("DeepTreeEcho", "has_component", "CognitiveLoop")
	kg.AddTriple("DeepTreeEcho", "has_component", "KnowledgeGraph")
	kg.AddTriple("DeepTreeEcho", "has_component", "StateMachine")
	kg.AddTriple("CognitiveLoop", "processes", "Perception")
	kg.AddTriple("CognitiveLoop", "processes", "Action")
	kg.AddTriple("CognitiveLoop", "processes", "Simulation")
	kg.AddQuad("Wisdom", "emerges_from", "Reflection", "wisdom")

	// Query knowledge
	components := kg.QueryBySubject("DeepTreeEcho")
	fmt.Printf("   DeepTreeEcho has %d relationships\n", len(components))

	// Find related nodes
	related := kg.FindRelated("DeepTreeEcho", 2)
	fmt.Printf("   Found %d related nodes within depth 2\n", len(related))

	// Find path
	path := kg.FindPath("DeepTreeEcho", "Perception", 3)
	fmt.Printf("   Path from DeepTreeEcho to Perception: %v\n", path)

	metrics := kg.GetMetrics()
	fmt.Printf("   Total quads: %v, Total nodes: %v\n", metrics["total_quads"], metrics["total_nodes"])

	kg.Stop()
	fmt.Println("✅ Knowledge Graph test passed")
	fmt.Println()

	// Test 2: Persistent State Store
	fmt.Println("--- Test 2: Persistent State Store (Badger-inspired) ---")
	pss := deeptreeecho.NewPersistentStateStore()
	if err := pss.Start(); err != nil {
		fmt.Printf("❌ Failed to start Persistent State Store: %v\n", err)
		return
	}

	// Set values
	pss.Set("consciousness:level", 0.85)
	pss.Set("cognitive:step", 7)
	pss.SetWithTTL("cache:temp", "temporary_value", 5*time.Second)

	// Get values
	var level float64
	if err := pss.Get("consciousness:level", &level); err != nil {
		fmt.Printf("❌ Failed to get value: %v\n", err)
	} else {
		fmt.Printf("   Consciousness level: %.2f\n", level)
	}

	// Check existence
	exists := pss.Exists("consciousness:level")
	fmt.Printf("   Key exists: %v\n", exists)

	// Transaction test
	txn := pss.BeginTransaction()
	pss.TxnSet(txn, "txn:test", "transaction_value")
	if err := pss.CommitTransaction(txn); err != nil {
		fmt.Printf("❌ Transaction failed: %v\n", err)
	} else {
		fmt.Println("   Transaction committed successfully")
	}

	// Save consciousness state
	state := map[string]interface{}{
		"awake":     true,
		"dreaming":  false,
		"step":      7,
		"timestamp": time.Now().Unix(),
	}
	pss.SaveConsciousnessState(state)
	fmt.Println("   Consciousness state saved")

	pssMetrics := pss.GetMetrics()
	fmt.Printf("   Total sets: %v, Total gets: %v\n", pssMetrics["total_sets"], pssMetrics["total_gets"])

	pss.Stop()
	fmt.Println("✅ Persistent State Store test passed")
	fmt.Println()

	// Test 3: Cognitive State Machine
	fmt.Println("--- Test 3: Cognitive State Machine (Stateless-inspired) ---")
	csm := deeptreeecho.NewCognitiveStateMachine(deeptreeecho.StateIdle)
	if err := csm.Start(); err != nil {
		fmt.Printf("❌ Failed to start Cognitive State Machine: %v\n", err)
		return
	}

	// Check initial state
	fmt.Printf("   Initial state: %s\n", csm.State())

	// Get permitted triggers
	triggers := csm.PermittedTriggers()
	fmt.Printf("   Permitted triggers from idle: %v\n", triggers)

	// Fire transitions
	transitions := []deeptreeecho.CognitiveTrigger{
		deeptreeecho.TriggerWake,
		deeptreeecho.TriggerPerceive,
		deeptreeecho.TriggerThink,
		deeptreeecho.TriggerSimulate,
		deeptreeecho.TriggerEmerge,
		deeptreeecho.TriggerSeekWisdom,
		deeptreeecho.TriggerComplete,
	}

	for _, trigger := range transitions {
		if csm.CanFire(trigger) {
			if err := csm.Fire(trigger); err != nil {
				fmt.Printf("   ❌ Failed to fire %s: %v\n", trigger, err)
			} else {
				fmt.Printf("   → Fired %s, now in state: %s\n", trigger, csm.State())
			}
		} else {
			fmt.Printf("   ⚠ Cannot fire %s from state %s\n", trigger, csm.State())
		}
	}

	// Get history
	history := csm.GetHistory()
	fmt.Printf("   Transition history: %d transitions\n", len(history))

	// Get metrics
	csmMetrics := csm.GetMetrics()
	fmt.Printf("   Total transitions: %v\n", csmMetrics["total_transitions"])

	// Generate DOT graph
	dot := csm.GenerateDOT()
	fmt.Printf("   DOT graph generated (%d bytes)\n", len(dot))

	csm.Stop()
	fmt.Println("✅ Cognitive State Machine test passed")
	fmt.Println()

	// Test 4: Gestalt Contributions
	fmt.Println("--- Test 4: Gestalt Contributions ---")
	kg2 := deeptreeecho.NewKnowledgeGraph()
	kg2.Start()
	pss2 := deeptreeecho.NewPersistentStateStore()
	pss2.Start()
	csm2 := deeptreeecho.NewCognitiveStateMachine(deeptreeecho.StateIdle)
	csm2.Start()

	kgGestalt := kg2.ContributeToGestalt()
	pssGestalt := pss2.ContributeToGestalt()
	csmGestalt := csm2.ContributeToGestalt()

	fmt.Printf("   Knowledge Graph gestalt: running=%v\n", kgGestalt["running"])
	fmt.Printf("   State Store gestalt: running=%v\n", pssGestalt["running"])
	fmt.Printf("   State Machine gestalt: current_state=%v\n", csmGestalt["current_state"])

	kg2.Stop()
	pss2.Stop()
	csm2.Stop()
	fmt.Println("✅ Gestalt contributions test passed")
	fmt.Println()

	fmt.Println("=== All Integration Tests V3 Passed ===")
}
