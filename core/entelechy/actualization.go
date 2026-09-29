package entelechy

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Actualization represents the process of realizing potential into actuality
// Core concept from Aristotelian philosophy adapted for AGI development
type Actualization struct {
	mu                sync.RWMutex
	potentials        map[string]*Potential
	actualizations    []*ActualizedCapability
	developmentStage  DevelopmentStage
	actualizationRate float64
	lastUpdate        time.Time
}

// Potential represents an unrealized capability or capacity
type Potential struct {
	ID           string
	Name         string
	Description  string
	Readiness    float64 // 0.0 to 1.0, how ready this is to be actualized
	Priority     float64
	Dependencies []string
	CreatedAt    time.Time
}

// ActualizedCapability represents a potential that has been realized
type ActualizedCapability struct {
	PotentialID  string
	Name         string
	ActualizedAt time.Time
	Proficiency  float64 // 0.0 to 1.0, how well this capability is developed
	UsageCount   int
	LastUsed     time.Time
}

// DevelopmentStage represents the current stage of ontogenetic development
type DevelopmentStage string

const (
	StageEmergent   DevelopmentStage = "emergent"
	StageDeveloping DevelopmentStage = "developing"
	StageMaturing   DevelopmentStage = "maturing"
	StageMature     DevelopmentStage = "mature"
)

// NewActualization creates a new actualization system
func NewActualization() *Actualization {
	return &Actualization{
		potentials:        make(map[string]*Potential),
		actualizations:    make([]*ActualizedCapability, 0),
		developmentStage:  StageEmergent,
		actualizationRate: 0.1,
		lastUpdate:        time.Now(),
	}
}

// AddPotential registers a new potential capability
func (a *Actualization) AddPotential(p *Potential) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if p.ID == "" {
		return fmt.Errorf("potential ID cannot be empty")
	}

	a.potentials[p.ID] = p
	return nil
}

// GetPotential retrieves a potential by ID
func (a *Actualization) GetPotential(id string) (*Potential, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	p, exists := a.potentials[id]
	return p, exists
}

// GetReadyPotentials returns potentials that are ready for actualization
func (a *Actualization) GetReadyPotentials(threshold float64) []*Potential {
	a.mu.RLock()
	defer a.mu.RUnlock()

	ready := make([]*Potential, 0)
	for _, p := range a.potentials {
		if p.Readiness >= threshold && a.dependenciesMet(p) {
			ready = append(ready, p)
		}
	}
	return ready
}

// dependenciesMet checks if all dependencies for a potential are actualized
func (a *Actualization) dependenciesMet(p *Potential) bool {
	for _, depID := range p.Dependencies {
		found := false
		for _, ac := range a.actualizations {
			if ac.PotentialID == depID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Actualize converts a potential into an actualized capability
func (a *Actualization) Actualize(ctx context.Context, potentialID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	p, exists := a.potentials[potentialID]
	if !exists {
		return fmt.Errorf("potential %s not found", potentialID)
	}

	if !a.dependenciesMet(p) {
		return fmt.Errorf("dependencies not met for potential %s", potentialID)
	}

	// Create actualized capability
	ac := &ActualizedCapability{
		PotentialID:  potentialID,
		Name:         p.Name,
		ActualizedAt: time.Now(),
		Proficiency:  0.1, // Start with low proficiency
		UsageCount:   0,
	}

	a.actualizations = append(a.actualizations, ac)
	delete(a.potentials, potentialID)

	return nil
}

// GetActualizations returns all actualized capabilities
func (a *Actualization) GetActualizations() []*ActualizedCapability {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]*ActualizedCapability, len(a.actualizations))
	copy(result, a.actualizations)
	return result
}

// UpdateProficiency increases proficiency through use
func (a *Actualization) UpdateProficiency(potentialID string, improvement float64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, ac := range a.actualizations {
		if ac.PotentialID == potentialID {
			ac.Proficiency = min(1.0, ac.Proficiency+improvement)
			ac.UsageCount++
			ac.LastUsed = time.Now()
			return nil
		}
	}

	return fmt.Errorf("actualized capability %s not found", potentialID)
}

// GetDevelopmentStage returns the current development stage
func (a *Actualization) GetDevelopmentStage() DevelopmentStage {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.developmentStage
}

// UpdateDevelopmentStage advances the development stage based on actualization progress
func (a *Actualization) UpdateDevelopmentStage() {
	a.mu.Lock()
	defer a.mu.Unlock()

	totalActualized := len(a.actualizations)
	avgProficiency := 0.0
	if totalActualized > 0 {
		for _, ac := range a.actualizations {
			avgProficiency += ac.Proficiency
		}
		avgProficiency /= float64(totalActualized)
	}

	// Stage progression logic
	if totalActualized >= 20 && avgProficiency >= 0.8 {
		a.developmentStage = StageMature
	} else if totalActualized >= 10 && avgProficiency >= 0.5 {
		a.developmentStage = StageMaturing
	} else if totalActualized >= 3 {
		a.developmentStage = StageDeveloping
	} else {
		a.developmentStage = StageEmergent
	}
}

// GetActualizationMetrics returns metrics about the actualization process
func (a *Actualization) GetActualizationMetrics() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return map[string]interface{}{
		"total_potentials":     len(a.potentials),
		"total_actualizations": len(a.actualizations),
		"development_stage":    string(a.developmentStage),
		"actualization_rate":   a.actualizationRate,
		"last_update":          a.lastUpdate,
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
