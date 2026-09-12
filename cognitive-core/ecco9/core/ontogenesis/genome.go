package ontogenesis

import (
	"fmt"
	"math"
	"math/rand"
)

// KernelGenome represents the genetic information of an ontogenetic kernel
// It encodes B-series coefficients as mutable genes
type KernelGenome struct {
	// B-series coefficients (mutable genes)
	Coefficients []float64

	// Tree structure (defines stage dependencies)
	TreeStructure []int

	// Mutation parameters
	MutationRate     float64
	MutationStrength float64

	// Metadata
	Generation int
	ParentIDs  []string
}

// NewKernelGenome creates a new kernel genome
func NewKernelGenome(coefficients []float64, treeStructure []int) *KernelGenome {
	return &KernelGenome{
		Coefficients:     coefficients,
		TreeStructure:    treeStructure,
		MutationRate:     0.1,
		MutationStrength: 0.05,
		Generation:       0,
		ParentIDs:        []string{},
	}
}

// Clone creates a deep copy of the genome
func (g *KernelGenome) Clone() *KernelGenome {
	coeffs := make([]float64, len(g.Coefficients))
	copy(coeffs, g.Coefficients)

	tree := make([]int, len(g.TreeStructure))
	copy(tree, g.TreeStructure)

	parents := make([]string, len(g.ParentIDs))
	copy(parents, g.ParentIDs)

	return &KernelGenome{
		Coefficients:     coeffs,
		TreeStructure:    tree,
		MutationRate:     g.MutationRate,
		MutationStrength: g.MutationStrength,
		Generation:       g.Generation,
		ParentIDs:        parents,
	}
}

// Mutate applies random mutations to the genome
func (g *KernelGenome) Mutate() {
	for i := range g.Coefficients {
		if rand.Float64() < g.MutationRate {
			// Gaussian mutation
			g.Coefficients[i] += rand.NormFloat64() * g.MutationStrength
		}
	}
}

// Crossover performs genetic crossover with another genome
func (g *KernelGenome) Crossover(other *KernelGenome) *KernelGenome {
	minLen := len(g.Coefficients)
	if len(other.Coefficients) < minLen {
		minLen = len(other.Coefficients)
	}

	// Single-point crossover
	crossoverPoint := rand.Intn(minLen)

	childCoeffs := make([]float64, minLen)
	copy(childCoeffs[:crossoverPoint], g.Coefficients[:crossoverPoint])
	copy(childCoeffs[crossoverPoint:], other.Coefficients[crossoverPoint:])

	// Inherit tree structure from parent with better fitness (decided by caller)
	childTree := make([]int, len(g.TreeStructure))
	copy(childTree, g.TreeStructure)

	child := &KernelGenome{
		Coefficients:     childCoeffs,
		TreeStructure:    childTree,
		MutationRate:     (g.MutationRate + other.MutationRate) / 2.0,
		MutationStrength: (g.MutationStrength + other.MutationStrength) / 2.0,
		Generation:       int(math.Max(float64(g.Generation), float64(other.Generation))) + 1,
		ParentIDs:        []string{}, // Set by caller
	}

	return child
}

// Distance calculates the genetic distance to another genome
func (g *KernelGenome) Distance(other *KernelGenome) float64 {
	if len(g.Coefficients) != len(other.Coefficients) {
		return math.MaxFloat64
	}

	sumSquares := 0.0
	for i := range g.Coefficients {
		diff := g.Coefficients[i] - other.Coefficients[i]
		sumSquares += diff * diff
	}

	return math.Sqrt(sumSquares)
}

// String returns a string representation of the genome
func (g *KernelGenome) String() string {
	return fmt.Sprintf("Genome[Gen=%d, Coeffs=%d, MutRate=%.3f]",
		g.Generation, len(g.Coefficients), g.MutationRate)
}

// Helper functions for random generation

func randomCoefficients(order int) []float64 {
	coeffs := make([]float64, order)
	for i := range coeffs {
		coeffs[i] = rand.Float64()*2.0 - 1.0 // Range [-1, 1]
	}
	return coeffs
}

func randomTreeStructure(order int) []int {
	tree := make([]int, order)
	tree[0] = -1 // Root has no parent
	for i := 1; i < order; i++ {
		tree[i] = rand.Intn(i) // Parent is any previous stage
	}
	return tree
}

// Genome module - placeholder for future implementation
