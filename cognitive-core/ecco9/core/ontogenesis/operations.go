package ontogenesis

import (
	"math"
)

func SelfGenerate(parent *OntogeneticKernel) *OntogeneticKernel {
	offspring := parent.Clone()
	offspring.Generation = parent.Generation + 1
	offspring.Genome.ParentIDs = []string{parent.ID}
	offspring.Genome.Coefficients = applyChainRule(parent.Genome.Coefficients)
	return offspring
}

func SelfOptimize(kernel *OntogeneticKernel, iterations int) *OntogeneticKernel {
	optimized := kernel.Clone()
	learningRate := 0.01
	problem := createOptimizationProblem()

	for iter := 0; iter < iterations; iter++ {
		gradient := computeGradient(optimized, problem)
		for i := range optimized.Genome.Coefficients {
			optimized.Genome.Coefficients[i] += learningRate * gradient[i]
		}
		optimized.Evaluate(problem)
		learningRate *= 0.99
	}

	return optimized
}

func SelfReproduce(parent1, parent2 *OntogeneticKernel) *OntogeneticKernel {
	offspringGenome := parent1.Genome.Crossover(parent2.Genome)
	offspringGenome.ParentIDs = []string{parent1.ID, parent2.ID}
	offspringGenome.Mutate()

	offspring := &OntogeneticKernel{
		ID:         generateKernelID(),
		Generation: offspringGenome.Generation,
		Genome:     offspringGenome,
		BirthTime:  parent1.BirthTime,
		Age:        0,
		Fitness:    0.0,
		State:      make([]float64, len(offspringGenome.Coefficients)),
		Output:     make([]float64, len(offspringGenome.Coefficients)),
		Metadata:   make(map[string]interface{}),
	}

	return offspring
}

func applyChainRule(coefficients []float64) []float64 {
	newCoeffs := make([]float64, len(coefficients))
	for i := range coefficients {
		newCoeffs[i] = math.Tanh(coefficients[i])
	}
	return newCoeffs
}

func computeGradient(kernel *OntogeneticKernel, problem TestProblem) []float64 {
	gradient := make([]float64, len(kernel.Genome.Coefficients))
	epsilon := 1e-5
	baseline := kernel.Evaluate(problem)

	for i := range kernel.Genome.Coefficients {
		original := kernel.Genome.Coefficients[i]
		kernel.Genome.Coefficients[i] += epsilon
		perturbed := kernel.Evaluate(problem)
		gradient[i] = (perturbed - baseline) / epsilon
		kernel.Genome.Coefficients[i] = original
	}

	return gradient
}

func createOptimizationProblem() TestProblem {
	return TestProblem{
		Name:             "Exponential",
		InitialCondition: []float64{1.0},
		Derivative: func(y []float64) []float64 {
			return []float64{y[0]}
		},
		ExactSolution: func(t float64) []float64 {
			return []float64{math.Exp(t)}
		},
		TStart: 0.0,
		TEnd:   1.0,
		Steps:  10,
	}
}

// Operations module - placeholder for future implementation
