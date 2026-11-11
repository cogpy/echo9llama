package deeptreeecho

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

// OntogeneticTracker manages development stages of cognitive primitives
// Based on ontogenesis: self-generating, evolving kernels
type OntogeneticTracker struct {
	mu         sync.RWMutex
	primitives map[string]*CognitivePrimitive

	// Development parameters
	embryonicDuration time.Duration
	juvenileDuration  time.Duration
	matureDuration    time.Duration
	maturityThreshold float64

	// Evolution parameters
	mutationRate  float64
	crossoverRate float64

	// Metrics
	totalGenerations int
	totalEvolutions  int
}

// CognitivePrimitive represents an evolving cognitive operation
type CognitivePrimitive struct {
	ID          string
	Name        string
	Generation  int
	Lineage     []string
	Stage       DevelopmentStage
	Fitness     float64
	Age         time.Duration
	CreatedAt   time.Time
	LastUpdated time.Time
	Genome      *PrimitiveGenome

	// Performance tracking
	SuccessCount int
	FailureCount int
	TotalUses    int
}

// PrimitiveGenome contains genetic information
type PrimitiveGenome struct {
	// Coefficient genes (mutable)
	CoefficientGenes []float64

	// Operator genes (mutable)
	OperatorGenes map[string]float64

	// Symmetry genes (immutable)
	SymmetryGenes []string

	// Preservation genes (immutable)
	PreservationGenes []string
}

// DevelopmentStage represents ontogenetic stage
type DevelopmentStage string

const (
	StageEmbryonic DevelopmentStage = "embryonic"
	StageJuvenile  DevelopmentStage = "juvenile"
	StageMature    DevelopmentStage = "mature"
	StageSenescent DevelopmentStage = "senescent"
)

// NewOntogeneticTracker creates development tracker
func NewOntogeneticTracker() *OntogeneticTracker {
	return &OntogeneticTracker{
		primitives:        make(map[string]*CognitivePrimitive),
		embryonicDuration: 1 * time.Hour,
		juvenileDuration:  24 * time.Hour,
		matureDuration:    7 * 24 * time.Hour,
		maturityThreshold: 0.7,
		mutationRate:      0.1,
		crossoverRate:     0.7,
	}
}

// RegisterPrimitive adds a new cognitive primitive
func (ot *OntogeneticTracker) RegisterPrimitive(name string, genome *PrimitiveGenome) string {
	ot.mu.Lock()
	defer ot.mu.Unlock()

	id := uuid.New().String()

	primitive := &CognitivePrimitive{
		ID:          id,
		Name:        name,
		Generation:  0,
		Lineage:     []string{},
		Stage:       StageEmbryonic,
		Fitness:     0.5,
		CreatedAt:   time.Now(),
		LastUpdated: time.Now(),
		Genome:      genome,
	}

	ot.primitives[id] = primitive

	return id
}

// UpdateStages progresses primitives through development stages
func (ot *OntogeneticTracker) UpdateStages() {
	ot.mu.Lock()
	defer ot.mu.Unlock()

	for _, primitive := range ot.primitives {
		age := time.Since(primitive.CreatedAt)
		primitive.Age = age

		// Determine stage based on age and fitness
		newStage := ot.determineStage(age, primitive.Fitness)

		if newStage != primitive.Stage {
			fmt.Printf("🧬 Primitive %s: %s → %s (fitness: %.2f)\n",
				primitive.Name,
				primitive.Stage,
				newStage,
				primitive.Fitness,
			)
			primitive.Stage = newStage
		}

		primitive.LastUpdated = time.Now()
	}
}

// determineStage calculates appropriate development stage
func (ot *OntogeneticTracker) determineStage(
	age time.Duration,
	fitness float64,
) DevelopmentStage {
	switch {
	case age < ot.embryonicDuration:
		return StageEmbryonic

	case age < ot.juvenileDuration && fitness < ot.maturityThreshold:
		return StageJuvenile

	case fitness >= ot.maturityThreshold && age < ot.matureDuration:
		return StageMature

	case age >= ot.matureDuration:
		return StageSenescent

	default:
		return StageJuvenile
	}
}

// UpdateFitness updates primitive fitness based on performance
func (ot *OntogeneticTracker) UpdateFitness(id string, success bool) {
	ot.mu.Lock()
	defer ot.mu.Unlock()

	primitive, exists := ot.primitives[id]
	if !exists {
		return
	}

	primitive.TotalUses++

	if success {
		primitive.SuccessCount++
	} else {
		primitive.FailureCount++
	}

	// Calculate fitness as success rate with smoothing
	if primitive.TotalUses > 0 {
		successRate := float64(primitive.SuccessCount) / float64(primitive.TotalUses)

		// Exponential moving average
		primitive.Fitness = (primitive.Fitness * 0.7) + (successRate * 0.3)
	}

	primitive.LastUpdated = time.Now()
}

// SelfGenerate creates offspring through recursive self-composition
// Implements: offspring = f(parent)
func (ot *OntogeneticTracker) SelfGenerate(parentID string) (string, error) {
	ot.mu.Lock()
	defer ot.mu.Unlock()

	parent, exists := ot.primitives[parentID]
	if !exists {
		return "", fmt.Errorf("parent primitive not found: %s", parentID)
	}

	// Create offspring genome through mutation
	offspringGenome := ot.mutateGenome(parent.Genome)

	offspring := &CognitivePrimitive{
		ID:          uuid.New().String(),
		Name:        fmt.Sprintf("%s_gen%d", parent.Name, parent.Generation+1),
		Generation:  parent.Generation + 1,
		Lineage:     append(parent.Lineage, parent.ID),
		Stage:       StageEmbryonic,
		Fitness:     parent.Fitness * 0.9, // Slight regression
		CreatedAt:   time.Now(),
		LastUpdated: time.Now(),
		Genome:      offspringGenome,
	}

	ot.primitives[offspring.ID] = offspring
	ot.totalGenerations++

	fmt.Printf("🧬 Self-generated: %s (gen %d) from %s\n",
		offspring.Name,
		offspring.Generation,
		parent.Name,
	)

	return offspring.ID, nil
}

// SelfReproduce combines two primitives to create offspring
// Implements crossover and mutation
func (ot *OntogeneticTracker) SelfReproduce(parent1ID, parent2ID string) (string, error) {
	ot.mu.Lock()
	defer ot.mu.Unlock()

	parent1, exists1 := ot.primitives[parent1ID]
	parent2, exists2 := ot.primitives[parent2ID]

	if !exists1 || !exists2 {
		return "", fmt.Errorf("one or both parents not found")
	}

	// Crossover genomes
	offspringGenome := ot.crossoverGenomes(parent1.Genome, parent2.Genome)

	// Apply mutation
	if math.Round(0.5) < ot.mutationRate { // Simplified random
		offspringGenome = ot.mutateGenome(offspringGenome)
	}

	maxGen := parent1.Generation
	if parent2.Generation > maxGen {
		maxGen = parent2.Generation
	}

	offspring := &CognitivePrimitive{
		ID:          uuid.New().String(),
		Name:        fmt.Sprintf("hybrid_%s_%s", parent1.Name, parent2.Name),
		Generation:  maxGen + 1,
		Lineage:     []string{parent1.ID, parent2.ID},
		Stage:       StageEmbryonic,
		Fitness:     (parent1.Fitness + parent2.Fitness) / 2.0,
		CreatedAt:   time.Now(),
		LastUpdated: time.Now(),
		Genome:      offspringGenome,
	}

	ot.primitives[offspring.ID] = offspring
	ot.totalEvolutions++

	fmt.Printf("🧬 Reproduced: %s from %s + %s\n",
		offspring.Name,
		parent1.Name,
		parent2.Name,
	)

	return offspring.ID, nil
}

// mutateGenome applies random mutations to genome
func (ot *OntogeneticTracker) mutateGenome(genome *PrimitiveGenome) *PrimitiveGenome {
	newGenome := &PrimitiveGenome{
		CoefficientGenes:  make([]float64, len(genome.CoefficientGenes)),
		OperatorGenes:     make(map[string]float64),
		SymmetryGenes:     genome.SymmetryGenes,     // Immutable
		PreservationGenes: genome.PreservationGenes, // Immutable
	}

	// Mutate coefficient genes
	for i, coeff := range genome.CoefficientGenes {
		mutation := (0.5 - 0.5) * 0.2 // ±10% mutation (simplified random)
		newGenome.CoefficientGenes[i] = coeff + mutation
	}

	// Mutate operator genes
	for key, value := range genome.OperatorGenes {
		mutation := (0.5 - 0.5) * 0.2
		newGenome.OperatorGenes[key] = value + mutation
	}

	return newGenome
}

// crossoverGenomes performs single-point crossover
func (ot *OntogeneticTracker) crossoverGenomes(
	genome1, genome2 *PrimitiveGenome,
) *PrimitiveGenome {
	newGenome := &PrimitiveGenome{
		CoefficientGenes:  make([]float64, len(genome1.CoefficientGenes)),
		OperatorGenes:     make(map[string]float64),
		SymmetryGenes:     genome1.SymmetryGenes,
		PreservationGenes: genome1.PreservationGenes,
	}

	// Single-point crossover on coefficients
	if len(genome1.CoefficientGenes) > 0 {
		point := len(genome1.CoefficientGenes) / 2

		for i := 0; i < len(genome1.CoefficientGenes); i++ {
			if i < point {
				newGenome.CoefficientGenes[i] = genome1.CoefficientGenes[i]
			} else if i < len(genome2.CoefficientGenes) {
				newGenome.CoefficientGenes[i] = genome2.CoefficientGenes[i]
			} else {
				newGenome.CoefficientGenes[i] = genome1.CoefficientGenes[i]
			}
		}
	}

	// Combine operator genes
	for key, value := range genome1.OperatorGenes {
		newGenome.OperatorGenes[key] = value
	}
	for key, value := range genome2.OperatorGenes {
		if _, exists := newGenome.OperatorGenes[key]; !exists {
			newGenome.OperatorGenes[key] = value
		}
	}

	return newGenome
}

// PruneSenescent removes senescent primitives
func (ot *OntogeneticTracker) PruneSenescent() int {
	ot.mu.Lock()
	defer ot.mu.Unlock()

	pruned := 0

	for id, primitive := range ot.primitives {
		if primitive.Stage == StageSenescent && primitive.Fitness < 0.3 {
			delete(ot.primitives, id)
			pruned++
			fmt.Printf("🗑️  Pruned senescent primitive: %s (fitness: %.2f)\n",
				primitive.Name,
				primitive.Fitness,
			)
		}
	}

	return pruned
}

// GetMetrics returns ontogenetic tracker metrics
func (ot *OntogeneticTracker) GetMetrics() map[string]interface{} {
	ot.mu.RLock()
	defer ot.mu.RUnlock()

	stageCount := make(map[DevelopmentStage]int)
	totalFitness := 0.0

	for _, primitive := range ot.primitives {
		stageCount[primitive.Stage]++
		totalFitness += primitive.Fitness
	}

	avgFitness := 0.0
	if len(ot.primitives) > 0 {
		avgFitness = totalFitness / float64(len(ot.primitives))
	}

	return map[string]interface{}{
		"total_primitives":  len(ot.primitives),
		"embryonic_count":   stageCount[StageEmbryonic],
		"juvenile_count":    stageCount[StageJuvenile],
		"mature_count":      stageCount[StageMature],
		"senescent_count":   stageCount[StageSenescent],
		"average_fitness":   avgFitness,
		"total_generations": ot.totalGenerations,
		"total_evolutions":  ot.totalEvolutions,
	}
}

// GetPrimitive retrieves a primitive by ID
func (ot *OntogeneticTracker) GetPrimitive(id string) (*CognitivePrimitive, bool) {
	ot.mu.RLock()
	defer ot.mu.RUnlock()

	primitive, exists := ot.primitives[id]
	return primitive, exists
}

// GetMaturePrimitives returns all mature primitives
func (ot *OntogeneticTracker) GetMaturePrimitives() []*CognitivePrimitive {
	ot.mu.RLock()
	defer ot.mu.RUnlock()

	mature := make([]*CognitivePrimitive, 0)

	for _, primitive := range ot.primitives {
		if primitive.Stage == StageMature {
			mature = append(mature, primitive)
		}
	}

	return mature
}
