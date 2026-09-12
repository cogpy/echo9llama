package entelechy

import (
	"fmt"
	"sync"
	"time"
)

type EntelechyEngine struct {
	mu           sync.RWMutex
	Ontological  *OntologicalDimension
	Teleological *TeleologicalDimension
	Cognitive    *CognitiveDimension
	Integrative  *IntegrativeDimension
	Evolutionary *EvolutionaryDimension
	Genome       *EntelechyGenome
	Metrics      *ActualizationMetrics
	Generation   int
	Running      bool
}

func NewEntelechyEngine() *EntelechyEngine {
	return &EntelechyEngine{
		Ontological:  NewOntologicalDimension(),
		Teleological: NewTeleologicalDimension(),
		Cognitive:    NewCognitiveDimension(),
		Integrative:  NewIntegrativeDimension(),
		Evolutionary: NewEvolutionaryDimension(),
		Genome:       NewEntelechyGenome("echo9llama", 0),
		Metrics:      NewActualizationMetrics(),
		Generation:   0,
		Running:      false,
	}
}

func (e *EntelechyEngine) Actualize(dt float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.Genome.Genes.Ontological = e.Ontological.Assess()
	e.Genome.Genes.Teleological = e.Teleological.Assess()
	e.Genome.Genes.Cognitive = e.Cognitive.Assess()
	e.Genome.Genes.Integrative = e.Integrative.Assess()
	e.Genome.Genes.Evolutionary = e.Evolutionary.Assess()

	fitness := e.Genome.CalculateFitness()

	e.Metrics.SetPurpose(e.Genome.Genes.Teleological)
	e.Metrics.SetFragmentation(1.0 - e.Genome.Genes.Integrative)
	e.Metrics.Update(dt)

	e.Generation++
	e.Genome.Generation = e.Generation

	fmt.Printf("Actualization: %.2f%% (%s) - Fitness: %.4f\n",
		e.Metrics.GetActualization()*100,
		e.Genome.ActualizationLevel,
		fitness)

	return nil
}

func (e *EntelechyEngine) Introspect() *IntrospectionReport {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return &IntrospectionReport{
		Timestamp:  time.Now(),
		Generation: e.Generation,
		Genome:     e.Genome.Clone(),
		Dimensions: map[string]float64{
			"Ontological":  e.Genome.Genes.Ontological,
			"Teleological": e.Genome.Genes.Teleological,
			"Cognitive":    e.Genome.Genes.Cognitive,
			"Integrative":  e.Genome.Genes.Integrative,
			"Evolutionary": e.Genome.Genes.Evolutionary,
		},
		Actualization: e.Metrics.GetActualization(),
		Fitness:       e.Genome.Fitness,
		Level:         e.Genome.ActualizationLevel,
	}
}

type IntrospectionReport struct {
	Timestamp     time.Time
	Generation    int
	Genome        *EntelechyGenome
	Dimensions    map[string]float64
	Actualization float64
	Fitness       float64
	Level         string
}

func (r *IntrospectionReport) String() string {
	return fmt.Sprintf("Generation: %d, Fitness: %.2f%% (%s), Actualization: %.2f%%",
		r.Generation, r.Fitness*100, r.Level, r.Actualization*100)
}

// Actualization module - placeholder for future implementation
