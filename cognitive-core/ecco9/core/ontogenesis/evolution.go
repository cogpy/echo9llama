package ontogenesis

import (
	"fmt"
	"math/rand"
	"sort"
)

// Population represents a population of ontogenetic kernels
type Population struct {
	Kernels        []*OntogeneticKernel
	Generation     int
	BestFitness    float64
	BestKernel     *OntogeneticKernel
	PopulationSize int
}

// NewPopulation creates a new population with random kernels
func NewPopulation(size, kernelOrder int) *Population {
	pop := &Population{
		Kernels:        make([]*OntogeneticKernel, size),
		Generation:     0,
		BestFitness:    0.0,
		PopulationSize: size,
	}

	for i := 0; i < size; i++ {
		pop.Kernels[i] = NewRandomKernel(kernelOrder)
	}

	return pop
}

// Evolve evolves the population for one generation using tournament selection
func (p *Population) Evolve(problem TestProblem, tournamentSize int) {
	for _, kernel := range p.Kernels {
		kernel.Evaluate(problem)
	}

	p.updateBest()

	newKernels := make([]*OntogeneticKernel, p.PopulationSize)
	newKernels[0] = p.BestKernel.Clone()

	for i := 1; i < p.PopulationSize; i++ {
		parent1 := p.tournamentSelect(tournamentSize)
		parent2 := p.tournamentSelect(tournamentSize)
		offspring := SelfReproduce(parent1, parent2)
		newKernels[i] = offspring
	}

	p.Kernels = newKernels
	p.Generation++
}

func (p *Population) tournamentSelect(tournamentSize int) *OntogeneticKernel {
	var best *OntogeneticKernel
	bestFitness := -1.0

	for i := 0; i < tournamentSize; i++ {
		idx := rand.Intn(len(p.Kernels))
		kernel := p.Kernels[idx]

		if kernel.Fitness > bestFitness {
			best = kernel
			bestFitness = kernel.Fitness
		}
	}

	return best
}

func (p *Population) updateBest() {
	for _, kernel := range p.Kernels {
		if kernel.Fitness > p.BestFitness {
			p.BestFitness = kernel.Fitness
			p.BestKernel = kernel.Clone()
		}
	}
}

// GetStatistics returns population statistics
func (p *Population) GetStatistics() PopulationStats {
	fitnesses := make([]float64, len(p.Kernels))
	for i, k := range p.Kernels {
		fitnesses[i] = k.Fitness
	}

	sort.Float64s(fitnesses)

	sum := 0.0
	for _, f := range fitnesses {
		sum += f
	}
	mean := sum / float64(len(fitnesses))
	median := fitnesses[len(fitnesses)/2]

	return PopulationStats{
		Generation:    p.Generation,
		PopSize:       len(p.Kernels),
		BestFitness:   p.BestFitness,
		MeanFitness:   mean,
		MedianFitness: median,
		MinFitness:    fitnesses[0],
		MaxFitness:    fitnesses[len(fitnesses)-1],
	}
}

func (p *Population) String() string {
	stats := p.GetStatistics()
	return fmt.Sprintf("Pop[Gen=%d, Size=%d, Best=%.4f, Mean=%.4f]",
		stats.Generation, stats.PopSize, stats.BestFitness, stats.MeanFitness)
}

type PopulationStats struct {
	Generation    int
	PopSize       int
	BestFitness   float64
	MeanFitness   float64
	MedianFitness float64
	MinFitness    float64
	MaxFitness    float64
}

// Evolution module - placeholder for future implementation
