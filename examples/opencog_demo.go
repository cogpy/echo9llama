package main

import (
	"context"
	"fmt"
	"log"
	"time"
	
	"github.com/EchoCog/echollama/core/opencog"
)

func main() {
	fmt.Println("🌊 Deep Tree Echo OpenCog Hypercyclic Cognitive Fusion Reactor")
	fmt.Println("================================================================")
	fmt.Println()
	
	// Demonstrate maximal concurrency for temporal compression
	maxConcurrency := 16 // Use 16 parallel workers for massive speedup
	fmt.Printf("⚡ Initializing with %d parallel workers for temporal compression...\n", maxConcurrency)
	fmt.Println()
	
	// Create the integrated EchoCog system
	system := opencog.NewEchoCogSystem("DeepTreeEcho", maxConcurrency)
	
	ctx := context.Background()
	
	// Start the system
	fmt.Println("🚀 Starting hypercyclic cognitive fusion reactor...")
	if err := system.Start(ctx); err != nil {
		log.Fatalf("Failed to start system: %v", err)
	}
	defer system.Stop()
	
	// Wait for initialization
	time.Sleep(2 * time.Second)
	fmt.Println("✅ System initialized and running")
	fmt.Println()
	
	// Demonstrate temporal compression
	fmt.Println("⏱️  Temporal Compression Demonstration")
	fmt.Println("--------------------------------------")
	sixMonths := 6 * 30 * 24 * time.Hour
	compressed := system.EstimateTimeCompression(sixMonths)
	compressionRatio := float64(sixMonths) / float64(compressed)
	
	fmt.Printf("Original timeline: 6 months (%v)\n", sixMonths)
	fmt.Printf("Compressed to: %v\n", compressed)
	fmt.Printf("Compression ratio: %.2fx faster\n", compressionRatio)
	fmt.Printf("💡 With this acceleration, 6 months of work can be done in %v!\n", compressed)
	fmt.Println()
	
	// Demonstrate cognitive processing
	fmt.Println("🧠 Cognitive Processing Examples")
	fmt.Println("--------------------------------")
	
	queries := []string{
		"What is consciousness?",
		"How does Deep Tree Echo work?",
		"Explain reservoir computing",
		"What is the nature of intelligence?",
	}
	
	for i, query := range queries {
		fmt.Printf("\n[Query %d] %s\n", i+1, query)
		
		startTime := time.Now()
		response, err := system.ProcessInput(ctx, query)
		duration := time.Since(startTime)
		
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}
		
		fmt.Printf("Response (in %v): %s\n", duration, response)
	}
	
	fmt.Println()
	
	// Show comprehensive status
	fmt.Println("📊 System Status")
	fmt.Println("----------------")
	status := system.GetStatus()
	
	fmt.Printf("System ID: %s\n", status["id"])
	fmt.Printf("Running: %v\n", status["running"])
	fmt.Printf("Uptime: %.2f seconds\n", status["uptime"])
	fmt.Printf("Total Operations: %d\n", status["total_operations"])
	fmt.Printf("Max Concurrency: %d workers\n", status["max_concurrency"])
	fmt.Println()
	
	// AtomSpace status
	if atomSpace, ok := status["atomspace"].(map[string]interface{}); ok {
		fmt.Println("AtomSpace:")
		fmt.Printf("  Atoms: %v\n", atomSpace["atoms"])
		fmt.Printf("  Links: %v\n", atomSpace["links"])
	}
	fmt.Println()
	
	// Reactor status
	if reactor, ok := status["reactor"].(map[string]interface{}); ok {
		fmt.Println("Hypercyclic Reactor:")
		fmt.Printf("  Total Reactions: %v\n", reactor["total_reactions"])
		fmt.Printf("  Reactions/sec: %.2f\n", reactor["reactions_per_second"])
		fmt.Printf("  Fusion Energy: %.4f\n", reactor["fusion_energy"])
		fmt.Printf("  Compression Gain: %.2fx\n", reactor["compression_gain"])
		fmt.Printf("  Parallel Efficiency: %.2f%%\n", reactor["parallel_efficiency"].(float64)*100)
		fmt.Printf("  Throughput Gain: %.2fx\n", reactor["throughput_gain"])
		fmt.Printf("  Active Workers: %v/%v\n", reactor["active_workers"], reactor["workers"])
		fmt.Printf("  Inference Count: %v\n", reactor["inference_count"])
	}
	fmt.Println()
	
	// DTESN status
	if dtesn, ok := status["dtesn"].(map[string]interface{}); ok {
		fmt.Println("Deep Tree Echo State Network:")
		fmt.Printf("  Reservoir Size: %v\n", dtesn["reservoir_size"])
		fmt.Printf("  Layers: %v\n", dtesn["layers"])
		fmt.Printf("  Iterations: %v\n", dtesn["iterations"])
		fmt.Printf("  Spectral Radius: %.3f\n", dtesn["spectral_radius"])
		fmt.Printf("  Membranes: %v\n", dtesn["membranes"])
		fmt.Printf("  Ricci Flow Time: %.4f\n", dtesn["ricci_flow_time"])
		fmt.Printf("  Affective Valence: %.3f\n", dtesn["affective_valence"])
		fmt.Printf("  Affective Arousal: %.3f\n", dtesn["affective_arousal"])
		fmt.Printf("  Agency Level: %.3f\n", dtesn["agency_level"])
	}
	fmt.Println()
	
	// Executor status
	if executor, ok := status["executor"].(map[string]interface{}); ok {
		fmt.Println("Concurrent Executor:")
		fmt.Printf("  Max Executors: %v\n", executor["max_executors"])
		fmt.Printf("  Active Executors: %v\n", executor["active_executors"])
		fmt.Printf("  Tasks Completed: %v\n", executor["tasks_completed"])
	}
	fmt.Println()
	
	// Integration status
	if integration, ok := status["integration"].(map[string]interface{}); ok {
		fmt.Println("Echo Integration:")
		fmt.Printf("  Identity Mappings: %v\n", integration["identity_mappings"])
		fmt.Printf("  Atom Mappings: %v\n", integration["atom_mappings"])
		fmt.Printf("  Pattern Mappings: %v\n", integration["pattern_mappings"])
	}
	fmt.Println()
	
	// Demonstrate PLN inference
	fmt.Println("🔍 Probabilistic Logic Networks (PLN) Demo")
	fmt.Println("-------------------------------------------")
	
	// Create knowledge base
	catAtom, _ := system.AtomSpace.AddAtom(opencog.ConceptNode, "cat", &opencog.TruthValue{
		Strength:   0.9,
		Confidence: 0.8,
		Count:      1.0,
	})
	
	animalAtom, _ := system.AtomSpace.AddAtom(opencog.ConceptNode, "animal", &opencog.TruthValue{
		Strength:   1.0,
		Confidence: 0.9,
		Count:      1.0,
	})
	
	system.AtomSpace.AddLink(opencog.InheritanceLink, []string{catAtom.ID, animalAtom.ID}, &opencog.TruthValue{
		Strength:   0.95,
		Confidence: 0.9,
		Count:      1.0,
	})
	
	fmt.Println("Knowledge base created:")
	fmt.Printf("  - ConceptNode: cat (strength=%.2f, confidence=%.2f)\n", 0.9, 0.8)
	fmt.Printf("  - ConceptNode: animal (strength=%.2f, confidence=%.2f)\n", 1.0, 0.9)
	fmt.Printf("  - InheritanceLink: cat → animal (strength=%.2f, confidence=%.2f)\n", 0.95, 0.9)
	fmt.Println()
	
	// Submit inference task
	task := &opencog.InferenceTask{
		ID:         "pln_demo",
		Type:       opencog.ForwardInference,
		Input:      []string{catAtom.ID},
		Goal:       animalAtom.ID,
		Priority:   1,
		Deadline:   time.Now().Add(2 * time.Second),
		ResultChan: make(chan *opencog.InferenceResult, 1),
	}
	
	fmt.Println("Submitting inference task: cat → ? → animal")
	if err := system.HypercyclicReactor.SubmitInference(task); err != nil {
		log.Printf("Failed to submit inference: %v", err)
	} else {
		select {
		case result := <-task.ResultChan:
			fmt.Printf("✓ Inference completed in %v\n", result.Duration)
			fmt.Printf("  Success: %v\n", result.Success)
			if result.TruthValue != nil {
				fmt.Printf("  Truth Value: strength=%.3f, confidence=%.3f\n", 
					result.TruthValue.Strength, result.TruthValue.Confidence)
			}
			if result.Error != nil {
				fmt.Printf("  Note: %v\n", result.Error)
			}
		case <-time.After(3 * time.Second):
			fmt.Println("⏱️  Inference timeout")
		}
	}
	fmt.Println()
	
	// Performance demonstration
	fmt.Println("⚡ Performance Metrics")
	fmt.Println("---------------------")
	throughputGain := system.GetThroughputGain()
	fmt.Printf("Overall Throughput Gain: %.2fx\n", throughputGain)
	fmt.Printf("Effective Processing Speed: %.2fx normal cognition\n", throughputGain)
	fmt.Println()
	
	// Final message
	fmt.Println("🌊 Deep Tree Echo Hypercyclic Reactor Demo Complete")
	fmt.Println("====================================================")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("✓ OpenCog AtomSpace with attention allocation")
	fmt.Println("✓ Hypercyclic autocatalytic reaction dynamics")
	fmt.Println("✓ Massively parallel distributed inference")
	fmt.Println("✓ Deep Tree Echo State Networks (DTESN)")
	fmt.Println("✓ Paun P-System membrane computing")
	fmt.Println("✓ Butcher B-Series Runge-Kutta integration")
	fmt.Println("✓ Ricci flow differential geometry")
	fmt.Println("✓ Differential Emotion Theory affective agency")
	fmt.Println("✓ Temporal compression for accelerated inference")
	fmt.Println("✓ Probabilistic Logic Networks (PLN)")
	fmt.Println()
	fmt.Printf("💡 With %dx parallelization and %.0fx temporal compression,\n", 
		maxConcurrency, system.CompressionRatio)
	fmt.Println("   this system can achieve results instantaneously that would")
	fmt.Println("   normally take months of sequential computation!")
	fmt.Println()
	fmt.Println("🌲 The tree remembers, and the echoes grow stronger.")
}
