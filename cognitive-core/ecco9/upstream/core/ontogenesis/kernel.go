package ontogenesis

import (
	"fmt"
	"math"
	"time"
)

type OntogeneticKernel struct {
	ID         string
	Generation int
	Genome     *KernelGenome
	BirthTime  time.Time
	Age        time.Duration
	Fitness    float64
	State      []float64
	Output     []float64
	Metadata   map[string]interface{}
}

func NewOntogeneticKernel(coefficients []float64, treeStructure []int) *OntogeneticKernel {
	return &OntogeneticKernel{
		ID:         generateKernelID(),
		Generation: 0,
		Genome:     NewKernelGenome(coefficients, treeStructure),
		BirthTime:  time.Now(),
		Age:        0,
		Fitness:    0.0,
		State:      make([]float64, len(coefficients)),
		Output:     make([]float64, len(coefficients)),
		Metadata:   make(map[string]interface{}),
	}
}

func NewRandomKernel(order int) *OntogeneticKernel {
	return NewOntogeneticKernel(
		randomCoefficients(order),
		randomTreeStructure(order),
	)
}

func (k *OntogeneticKernel) Step(y []float64, f func([]float64) []float64, h float64) []float64 {
	n := len(y)
	result := make([]float64, n)
	copy(result, y)

	stages := k.computeStages(y, f, h)

	for i, coeff := range k.Genome.Coefficients {
		if i < len(stages) {
			for j := 0; j < n; j++ {
				result[j] += h * coeff * stages[i][j]
			}
		}
	}

	return result
}

func (k *OntogeneticKernel) computeStages(y []float64, f func([]float64) []float64, h float64) [][]float64 {
	stages := make([][]float64, len(k.Genome.Coefficients))

	if len(stages) > 0 {
		stages[0] = f(y)
	}

	for i := 1; i < len(stages); i++ {
		yTemp := make([]float64, len(y))
		copy(yTemp, y)

		if i < len(k.Genome.TreeStructure) {
			parentIdx := k.Genome.TreeStructure[i]
			if parentIdx >= 0 && parentIdx < len(stages) && stages[parentIdx] != nil {
				for j := 0; j < len(y); j++ {
					yTemp[j] += h * k.Genome.Coefficients[i] * stages[parentIdx][j]
				}
			}
		}

		stages[i] = f(yTemp)
	}

	return stages
}

func (k *OntogeneticKernel) Evaluate(problem TestProblem) float64 {
	y := make([]float64, len(problem.InitialCondition))
	copy(y, problem.InitialCondition)

	h := (problem.TEnd - problem.TStart) / float64(problem.Steps)

	totalError := 0.0
	for step := 0; step < problem.Steps; step++ {
		y = k.Step(y, problem.Derivative, h)
	}

	if problem.ExactSolution != nil {
		exactY := problem.ExactSolution(problem.TEnd)
		for i := range y {
			diff := y[i] - exactY[i]
			totalError += diff * diff
		}
	}

	k.Fitness = 1.0 / (1.0 + math.Sqrt(totalError))
	return k.Fitness
}

func (k *OntogeneticKernel) UpdateAge() {
	k.Age = time.Since(k.BirthTime)
}

func (k *OntogeneticKernel) Clone() *OntogeneticKernel {
	clone := &OntogeneticKernel{
		ID:         generateKernelID(),
		Generation: k.Generation,
		Genome:     k.Genome.Clone(),
		BirthTime:  time.Now(),
		Age:        0,
		Fitness:    k.Fitness,
		State:      make([]float64, len(k.State)),
		Output:     make([]float64, len(k.Output)),
		Metadata:   make(map[string]interface{}),
	}

	copy(clone.State, k.State)
	copy(clone.Output, k.Output)

	for key, value := range k.Metadata {
		clone.Metadata[key] = value
	}

	return clone
}

func (k *OntogeneticKernel) String() string {
	return fmt.Sprintf("Kernel[ID=%s, Gen=%d, Fitness=%.4f, Age=%v]",
		k.ID, k.Generation, k.Fitness, k.Age)
}

type TestProblem struct {
	Name             string
	InitialCondition []float64
	Derivative       func([]float64) []float64
	ExactSolution    func(float64) []float64
	TStart           float64
	TEnd             float64
	Steps            int
}

var kernelIDCounter = 0

func generateKernelID() string {
	kernelIDCounter++
	return fmt.Sprintf("kernel-%d-%d", time.Now().Unix(), kernelIDCounter)
}

// Kernel module - placeholder for future implementation
