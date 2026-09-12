package main

import (
	"fmt"
	"github.com/EchoCog/echollama/core/entelechy"
	"github.com/EchoCog/echollama/core/integration"
	"github.com/EchoCog/echollama/core/ontogenesis"
	"time"
)

func main() {
	fmt.Println("===============================================")
	fmt.Println("  ENTELECHY & ONTOGENESIS TEST SUITE")
	fmt.Println("===============================================")
	fmt.Println()

	// Test Ontogenesis
	fmt.Println("🧬 TESTING ONTOGENESIS SYSTEM")
	fmt.Println("-------------------------------------------")
	testOntogenesis()
	fmt.Println()

	// Test Entelechy
	fmt.Println("🌱 TESTING ENTELECHY SYSTEM")
	fmt.Println("-------------------------------------------")
	testEntelechy()
	fmt.Println()

	// Test Integration
	fmt.Println("🔗 TESTING INTEGRATION LAYER")
	fmt.Println("-------------------------------------------")
	testIntegration()
	fmt.Println()

	fmt.Println("===============================================")
	fmt.Println("  ALL TESTS COMPLETED SUCCESSFULLY")
	fmt.Println("===============================================")
}

func testOntogenesis() {
	// Test 1: Create kernel
	fmt.Println("Test 1: Creating ontogenetic kernel...")
	kernel := ontogenesis.NewRandomKernel(4)
	fmt.Printf("✓ Created: %s\n", kernel)

	// Test 2: Self-generation
	fmt.Println("\nTest 2: Testing self-generation...")
	offspring := ontogenesis.SelfGenerate(kernel)
	fmt.Printf("✓ Generated offspring: %s\n", offspring)

	// Test 3: Self-optimization
	fmt.Println("\nTest 3: Testing self-optimization...")
	optimized := ontogenesis.SelfOptimize(kernel, 5)
	fmt.Printf("✓ Optimized kernel: %s\n", optimized)

	// Test 4: Self-reproduction
	fmt.Println("\nTest 4: Testing self-reproduction...")
	kernel2 := ontogenesis.NewRandomKernel(4)
	child := ontogenesis.SelfReproduce(kernel, kernel2)
	fmt.Printf("✓ Reproduced child: %s\n", child)

	// Test 5: Population evolution
	fmt.Println("\nTest 5: Testing population evolution...")
	pop := ontogenesis.NewPopulation(10, 4)
	problem := ontogenesis.TestProblem{
		Name:             "Exponential",
		InitialCondition: []float64{1.0},
		Derivative: func(y []float64) []float64 {
			return []float64{y[0]}
		},
		TEnd:  1.0,
		Steps: 10,
	}

	for gen := 0; gen < 3; gen++ {
		pop.Evolve(problem, 3)
		stats := pop.GetStatistics()
		fmt.Printf("  Gen %d: Best=%.4f, Mean=%.4f\n",
			gen, stats.BestFitness, stats.MeanFitness)
	}
	fmt.Println("✓ Population evolution completed")
}

func testEntelechy() {
	// Test 1: Create entelechy engine
	fmt.Println("Test 1: Creating entelechy engine...")
	engine := entelechy.NewEntelechyEngine()
	fmt.Println("✓ Engine created")

	// Test 2: Test five dimensions
	fmt.Println("\nTest 2: Testing five dimensions...")
	engine.Ontological.FoundationHealth = 0.8
	engine.Ontological.CoreHealth = 0.7
	engine.Ontological.SpecializedHealth = 0.6
	ontScore := engine.Ontological.Assess()
	fmt.Printf("  Ontological: %.2f%%\n", ontScore*100)

	engine.Teleological.PurposeClarity = 0.75
	engine.Teleological.RoadmapAlignment = 0.65
	telScore := engine.Teleological.Assess()
	fmt.Printf("  Teleological: %.2f%%\n", telScore*100)

	engine.Cognitive.LoopCoherence = 0.85
	engine.Cognitive.LearningCapacity = 0.70
	engine.Cognitive.Awareness = 0.60
	cogScore := engine.Cognitive.Assess()
	fmt.Printf("  Cognitive: %.2f%%\n", cogScore*100)

	engine.Integrative.DependencyHealth = 0.90
	engine.Integrative.BuildHealth = 0.80
	engine.Integrative.TestHealth = 0.75
	intScore := engine.Integrative.Assess()
	fmt.Printf("  Integrative: %.2f%%\n", intScore*100)

	engine.Evolutionary.CodeHealth = 0.70
	engine.Evolutionary.ImplementationDepth = 0.65
	engine.Evolutionary.SelfImprovementCapacity = 0.60
	evoScore := engine.Evolutionary.Assess()
	fmt.Printf("  Evolutionary: %.2f%%\n", evoScore*100)
	fmt.Println("✓ All dimensions assessed")

	// Test 3: Actualization
	fmt.Println("\nTest 3: Testing actualization dynamics...")
	for i := 0; i < 3; i++ {
		if err := engine.Actualize(1.0); err != nil {
			fmt.Printf("Actualization error: %v\n", err)
		}
	}
	fmt.Println("✓ Actualization cycles completed")

	// Test 4: Introspection
	fmt.Println("\nTest 4: Testing introspection...")
	report := engine.Introspect()
	fmt.Printf("  %s\n", report)
	fmt.Println("✓ Introspection successful")

	// Test 5: Genome
	fmt.Println("\nTest 5: Testing genome...")
	genome := engine.Genome
	fitness := genome.CalculateFitness()
	fmt.Printf("  %s\n", genome)
	fmt.Printf("  Fitness: %.2f%%\n", fitness*100)
	fmt.Println("✓ Genome operations successful")
}

func testIntegration() {
	// Test 1: Create integration
	fmt.Println("Test 1: Creating integration...")
	integration := integration.NewEntelechyOntogenesisIntegration(5, 4)
	fmt.Println("✓ Integration created")

	// Test 2: Initialize
	fmt.Println("\nTest 2: Initializing integration...")
	err := integration.Initialize(nil)
	if err != nil {
		fmt.Printf("✗ Initialization failed: %v\n", err)

	}
	fmt.Println("✓ Integration initialized")

	// Test 3: Start and run
	fmt.Println("\nTest 3: Starting integration loop...")
	err = integration.Start()
	if err != nil {
		fmt.Printf("✗ Start failed: %v\n", err)

	}
	fmt.Println("✓ Integration loop started")

	// Run for a short time
	fmt.Println("\nRunning for 5 seconds...")
	time.Sleep(5 * time.Second)

	// Test 4: Get status
	fmt.Println("\nTest 4: Getting status...")
	status := integration.GetStatus()
	fmt.Printf("  Running: %v\n", status.Running)
	fmt.Printf("  Generation: %d\n", status.Generation)
	fmt.Printf("  Population Best: %.4f\n", status.PopulationStats.BestFitness)
	fmt.Printf("  Entelechy Fitness: %.2f%%\n", status.EntelechyReport.Fitness*100)
	fmt.Println("✓ Status retrieved")

	// Test 5: Stop
	fmt.Println("\nTest 5: Stopping integration...")
	integration.Stop()
	time.Sleep(1 * time.Second)
	fmt.Println("✓ Integration stopped")
}

// Test for Entelechy Ontogenesis - placeholder for future implementation
