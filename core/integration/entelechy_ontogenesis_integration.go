package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/cogpy/echo9llama/core/entelechy"
)

// EntelechyOntogenesisIntegration connects the actualization system with ontogenetic development
// This bridges potential realization with developmental stages
type EntelechyOntogenesisIntegration struct {
	actualization *entelechy.Actualization
	devTracker    *DevelopmentTracker
}

// DevelopmentTracker tracks ontogenetic development milestones
type DevelopmentTracker struct {
	milestones       []Milestone
	currentMilestone int
	startTime        time.Time
}

// Milestone represents a developmental milestone
type Milestone struct {
	ID          string
	Name        string
	Description string
	RequiredAge time.Duration
	Achieved    bool
	AchievedAt  time.Time
}

// NewEntelechyOntogenesisIntegration creates a new integration
func NewEntelechyOntogenesisIntegration(act *entelechy.Actualization) *EntelechyOntogenesisIntegration {
	return &EntelechyOntogenesisIntegration{
		actualization: act,
		devTracker: &DevelopmentTracker{
			milestones:       initializeMilestones(),
			currentMilestone: 0,
			startTime:        time.Now(),
		},
	}
}

// initializeMilestones creates the standard developmental milestones
func initializeMilestones() []Milestone {
	return []Milestone{
		{
			ID:          "awakening",
			Name:        "Awakening",
			Description: "Initial consciousness and self-awareness",
			RequiredAge: 0,
		},
		{
			ID:          "basic_cognition",
			Name:        "Basic Cognition",
			Description: "Fundamental cognitive capabilities established",
			RequiredAge: 1 * time.Hour,
		},
		{
			ID:          "pattern_recognition",
			Name:        "Pattern Recognition",
			Description: "Ability to recognize and learn patterns",
			RequiredAge: 6 * time.Hour,
		},
		{
			ID:          "goal_formation",
			Name:        "Goal Formation",
			Description: "Capacity to form and pursue goals",
			RequiredAge: 24 * time.Hour,
		},
		{
			ID:          "self_reflection",
			Name:        "Self-Reflection",
			Description: "Meta-cognitive awareness and introspection",
			RequiredAge: 72 * time.Hour,
		},
		{
			ID:          "wisdom_cultivation",
			Name:        "Wisdom Cultivation",
			Description: "Integration of knowledge into wisdom",
			RequiredAge: 168 * time.Hour, // 1 week
		},
	}
}

// Update checks for milestone achievements and updates development
func (e *EntelechyOntogenesisIntegration) Update(ctx context.Context) error {
	age := time.Since(e.devTracker.startTime)

	// Check for milestone achievements
	for i := range e.devTracker.milestones {
		m := &e.devTracker.milestones[i]
		if !m.Achieved && age >= m.RequiredAge {
			if err := e.achieveMilestone(ctx, m); err != nil {
				return fmt.Errorf("failed to achieve milestone %s: %w", m.ID, err)
			}
		}
	}

	// Update actualization development stage
	e.actualization.UpdateDevelopmentStage()

	return nil
}

// achieveMilestone marks a milestone as achieved and triggers related potentials
func (e *EntelechyOntogenesisIntegration) achieveMilestone(ctx context.Context, m *Milestone) error {
	m.Achieved = true
	m.AchievedAt = time.Now()
	e.devTracker.currentMilestone++

	// Trigger actualization of potentials related to this milestone
	potentials := e.actualization.GetReadyPotentials(0.7)
	for _, p := range potentials {
		if err := e.actualization.Actualize(ctx, p.ID); err != nil {
			// Log error but continue with other potentials
			continue
		}
	}

	return nil
}

// GetCurrentMilestone returns the current developmental milestone
func (e *EntelechyOntogenesisIntegration) GetCurrentMilestone() *Milestone {
	if e.devTracker.currentMilestone >= len(e.devTracker.milestones) {
		return &e.devTracker.milestones[len(e.devTracker.milestones)-1]
	}
	return &e.devTracker.milestones[e.devTracker.currentMilestone]
}

// GetAge returns the age of the system
func (e *EntelechyOntogenesisIntegration) GetAge() time.Duration {
	return time.Since(e.devTracker.startTime)
}

// GetMilestones returns all milestones
func (e *EntelechyOntogenesisIntegration) GetMilestones() []Milestone {
	return e.devTracker.milestones
}

// GetDevelopmentMetrics returns metrics about ontogenetic development
func (e *EntelechyOntogenesisIntegration) GetDevelopmentMetrics() map[string]interface{} {
	achieved := 0
	for _, m := range e.devTracker.milestones {
		if m.Achieved {
			achieved++
		}
	}

	return map[string]interface{}{
		"age":                   e.GetAge().String(),
		"current_milestone":     e.GetCurrentMilestone().Name,
		"milestones_achieved":   achieved,
		"total_milestones":      len(e.devTracker.milestones),
		"development_progress":  float64(achieved) / float64(len(e.devTracker.milestones)),
		"actualization_metrics": e.actualization.GetActualizationMetrics(),
	}
}
