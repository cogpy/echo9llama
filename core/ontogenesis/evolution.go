package ontogenesis

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Evolution manages the evolutionary development of cognitive capabilities
// Combines ontogenetic (individual development) with phylogenetic (species-level) patterns
type Evolution struct {
	mu                    sync.RWMutex
	generationCount       int
	developmentalCurve    *DevelopmentalCurve
	adaptations           []*Adaptation
	evolutionaryPressures []EvolutionaryPressure
	fitnessScore          float64
	lastEvolution         time.Time
}

// DevelopmentalCurve tracks growth patterns over time
type DevelopmentalCurve struct {
	DataPoints []DevelopmentPoint
	GrowthRate float64
	Plateau    float64
}

// DevelopmentPoint represents a measurement in development
type DevelopmentPoint struct {
	Timestamp  time.Time
	Capability string
	Level      float64
	Context    string
}

// Adaptation represents an evolved capability or trait
type Adaptation struct {
	ID          string
	Name        string
	Description string
	Fitness     float64
	Generation  int
	Emerged     time.Time
	Stable      bool
}

// EvolutionaryPressure represents environmental factors driving evolution
type EvolutionaryPressure struct {
	Name      string
	Intensity float64
	Direction string // "increase", "decrease", "maintain"
}

// NewEvolution creates a new evolution system
func NewEvolution() *Evolution {
	return &Evolution{
		generationCount: 0,
		developmentalCurve: &DevelopmentalCurve{
			DataPoints: make([]DevelopmentPoint, 0),
			GrowthRate: 0.1,
			Plateau:    0.9,
		},
		adaptations:           make([]*Adaptation, 0),
		evolutionaryPressures: initializeBasePressures(),
		fitnessScore:          0.5,
		lastEvolution:         time.Now(),
	}
}

// initializeBasePressures sets up fundamental evolutionary pressures
func initializeBasePressures() []EvolutionaryPressure {
	return []EvolutionaryPressure{
		{Name: "complexity_management", Intensity: 0.7, Direction: "increase"},
		{Name: "efficiency", Intensity: 0.8, Direction: "increase"},
		{Name: "adaptability", Intensity: 0.9, Direction: "increase"},
		{Name: "stability", Intensity: 0.6, Direction: "maintain"},
		{Name: "wisdom_cultivation", Intensity: 0.5, Direction: "increase"},
	}
}

// RecordDevelopment adds a development data point
func (e *Evolution) RecordDevelopment(capability string, level float64, context string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	dp := DevelopmentPoint{
		Timestamp:  time.Now(),
		Capability: capability,
		Level:      level,
		Context:    context,
	}

	e.developmentalCurve.DataPoints = append(e.developmentalCurve.DataPoints, dp)

	// Update growth rate based on recent development
	if len(e.developmentalCurve.DataPoints) > 10 {
		e.updateGrowthRate()
	}
}

// updateGrowthRate calculates growth rate from recent data points
func (e *Evolution) updateGrowthRate() {
	points := e.developmentalCurve.DataPoints
	n := len(points)
	if n < 2 {
		return
	}

	// Calculate average growth over last 10 points
	start := n - 10
	if start < 0 {
		start = 0
	}

	totalGrowth := 0.0
	count := 0
	for i := start + 1; i < n; i++ {
		if points[i].Capability == points[i-1].Capability {
			growth := points[i].Level - points[i-1].Level
			totalGrowth += growth
			count++
		}
	}

	if count > 0 {
		e.developmentalCurve.GrowthRate = totalGrowth / float64(count)
	}
}

// EvolveAdaptation creates a new adaptation based on pressures
func (e *Evolution) EvolveAdaptation(ctx context.Context, name, description string) (*Adaptation, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Calculate fitness based on evolutionary pressures
	fitness := e.calculateFitness(name)

	adaptation := &Adaptation{
		ID:          fmt.Sprintf("adapt_%d_%d", e.generationCount, time.Now().Unix()),
		Name:        name,
		Description: description,
		Fitness:     fitness,
		Generation:  e.generationCount,
		Emerged:     time.Now(),
		Stable:      false,
	}

	e.adaptations = append(e.adaptations, adaptation)
	e.generationCount++
	e.lastEvolution = time.Now()

	return adaptation, nil
}

// calculateFitness determines fitness score for an adaptation
func (e *Evolution) calculateFitness(adaptationName string) float64 {
	// Base fitness
	fitness := 0.5

	// Adjust based on evolutionary pressures
	for _, pressure := range e.evolutionaryPressures {
		// Simple heuristic: increase fitness if adaptation aligns with pressure
		if pressure.Direction == "increase" {
			fitness += pressure.Intensity * 0.1
		}
	}

	// Normalize to 0-1 range
	if fitness > 1.0 {
		fitness = 1.0
	}
	if fitness < 0.0 {
		fitness = 0.0
	}

	return fitness
}

// StabilizeAdaptation marks an adaptation as stable after testing
func (e *Evolution) StabilizeAdaptation(adaptationID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, a := range e.adaptations {
		if a.ID == adaptationID {
			a.Stable = true
			return nil
		}
	}

	return fmt.Errorf("adaptation %s not found", adaptationID)
}

// GetAdaptations returns all adaptations
func (e *Evolution) GetAdaptations() []*Adaptation {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]*Adaptation, len(e.adaptations))
	copy(result, e.adaptations)
	return result
}

// GetStableAdaptations returns only stable adaptations
func (e *Evolution) GetStableAdaptations() []*Adaptation {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stable := make([]*Adaptation, 0)
	for _, a := range e.adaptations {
		if a.Stable {
			stable = append(stable, a)
		}
	}
	return stable
}

// UpdateFitnessScore recalculates overall fitness
func (e *Evolution) UpdateFitnessScore() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.adaptations) == 0 {
		e.fitnessScore = 0.5
		return
	}

	// Calculate average fitness of stable adaptations
	totalFitness := 0.0
	stableCount := 0
	for _, a := range e.adaptations {
		if a.Stable {
			totalFitness += a.Fitness
			stableCount++
		}
	}

	if stableCount > 0 {
		e.fitnessScore = totalFitness / float64(stableCount)
	}
}

// GetEvolutionMetrics returns metrics about evolutionary progress
func (e *Evolution) GetEvolutionMetrics() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stableCount := 0
	for _, a := range e.adaptations {
		if a.Stable {
			stableCount++
		}
	}

	return map[string]interface{}{
		"generation_count":     e.generationCount,
		"total_adaptations":    len(e.adaptations),
		"stable_adaptations":   stableCount,
		"fitness_score":        e.fitnessScore,
		"growth_rate":          e.developmentalCurve.GrowthRate,
		"development_points":   len(e.developmentalCurve.DataPoints),
		"last_evolution":       e.lastEvolution,
		"time_since_evolution": time.Since(e.lastEvolution).String(),
	}
}

// GetDevelopmentalCurve returns the developmental curve data
func (e *Evolution) GetDevelopmentalCurve() *DevelopmentalCurve {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.developmentalCurve
}
