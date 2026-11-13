package integration

import (
	"fmt"
	"sync"
	"time"

	"github.com/EchoCog/echollama/core/entelechy"
	"github.com/EchoCog/echollama/core/ontogenesis"
)

type EntelechyOntogenesisIntegration struct {
	mu                    sync.RWMutex
	Entelechy             *entelechy.EntelechyEngine
	Population            *ontogenesis.Population
	EvolutionInterval     time.Duration
	ActualizationInterval time.Duration
	Running               bool
	Generation            int
	OnEvolution           func(*ontogenesis.PopulationStats)
	OnActualization       func(*entelechy.IntrospectionReport)
}

func NewEntelechyOntogenesisIntegration(popSize, kernelOrder int) *EntelechyOntogenesisIntegration {
	return &EntelechyOntogenesisIntegration{
		Entelechy:             entelechy.NewEntelechyEngine(),
		Population:            ontogenesis.NewPopulation(popSize, kernelOrder),
		EvolutionInterval:     30 * time.Second,
		ActualizationInterval: 10 * time.Second,
		Running:               false,
		Generation:            0,
	}
}

func (i *EntelechyOntogenesisIntegration) Initialize(seedKernels []*ontogenesis.OntogeneticKernel) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if len(seedKernels) > 0 {
		for idx, kernel := range seedKernels {
			if idx < len(i.Population.Kernels) {
				i.Population.Kernels[idx] = kernel
			}
		}
	}
	fmt.Println("Entelechy-Ontogenesis Integration initialized")
	return nil
}

func (i *EntelechyOntogenesisIntegration) Start() error {
	i.mu.Lock()
	if i.Running {
		i.mu.Unlock()
		return fmt.Errorf("already running")
	}
	i.Running = true
	i.mu.Unlock()

	fmt.Println("Starting Entelechy-Ontogenesis continuous loop")
	go i.evolutionLoop()
	go i.actualizationLoop()
	return nil
}

func (i *EntelechyOntogenesisIntegration) Stop() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Running = false
	fmt.Println("Stopped Entelechy-Ontogenesis integration")
}

func (i *EntelechyOntogenesisIntegration) evolutionLoop() {
	ticker := time.NewTicker(i.EvolutionInterval)
	defer ticker.Stop()

	problem := ontogenesis.TestProblem{
		Name:             "Integration",
		InitialCondition: []float64{1.0},
		Derivative: func(y []float64) []float64 {
			return []float64{y[0]}
		},
		TEnd:  1.0,
		Steps: 10,
	}

	for {
		select {
		case <-ticker.C:
			i.mu.RLock()
			running := i.Running
			i.mu.RUnlock()
			if !running {
				return
			}

			i.Population.Evolve(problem, 3)
			stats := i.Population.GetStatistics()
			fmt.Printf("Evolution Gen %d: Best=%.4f, Mean=%.4f\n",
				i.Generation, stats.BestFitness, stats.MeanFitness)
			if i.OnEvolution != nil {
				i.OnEvolution(&stats)
			}
			i.Generation++
		}
	}
}

func (i *EntelechyOntogenesisIntegration) actualizationLoop() {
	ticker := time.NewTicker(i.ActualizationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			i.mu.RLock()
			running := i.Running
			i.mu.RUnlock()
			if !running {
				return
			}

			dt := i.ActualizationInterval.Seconds()
			if err := i.Entelechy.Actualize(dt); err != nil {
				fmt.Printf("Actualization error: %v\n", err)
				continue
			}
			report := i.Entelechy.Introspect()
			if i.OnActualization != nil {
				i.OnActualization(report)
			}
		}
	}
}

func (i *EntelechyOntogenesisIntegration) GetStatus() IntegrationStatus {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return IntegrationStatus{
		Running:         i.Running,
		Generation:      i.Generation,
		PopulationStats: i.Population.GetStatistics(),
		EntelechyReport: i.Entelechy.Introspect(),
	}
}

type IntegrationStatus struct {
	Running         bool
	Generation      int
	PopulationStats ontogenesis.PopulationStats
	EntelechyReport *entelechy.IntrospectionReport
}
