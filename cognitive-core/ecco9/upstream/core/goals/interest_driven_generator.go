package goals

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

// InterestDrivenGoalGenerator creates goals based on curiosity and interest patterns
type InterestDrivenGoalGenerator struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Interest tracking
	interestPatterns map[string]*InterestPattern
	curiosityLevel   float64
	explorationRate  float64

	// Goal generation
	goalOrchestrator *GoalOrchestrator
	generatedGoals   []*Goal

	// Topic tracking
	exploredTopics   map[string]bool
	unexploredTopics []string
	currentFocus     []string

	// Configuration
	minInterestThreshold float64
	maxGoalsPerCycle     int
	generationInterval   time.Duration

	// Metrics
	goalsGenerated   uint64
	explorationGoals uint64
	learningGoals    uint64
	discussionGoals  uint64

	running bool
}

// InterestPattern represents interest in a topic or domain
type InterestPattern struct {
	Topic           string
	Strength        float64 // 0.0 to 1.0
	Recency         float64 // How recently engaged
	Depth           float64 // How deeply explored
	Novelty         float64 // How novel/unfamiliar
	Utility         float64 // Perceived usefulness
	LastEngaged     time.Time
	EngagementCount int
	RelatedTopics   []string
}

// NewInterestDrivenGoalGenerator creates a new interest-driven goal generator
func NewInterestDrivenGoalGenerator(goalOrchestrator *GoalOrchestrator) *InterestDrivenGoalGenerator {
	ctx, cancel := context.WithCancel(context.Background())

	idgg := &InterestDrivenGoalGenerator{
		ctx:                  ctx,
		cancel:               cancel,
		goalOrchestrator:     goalOrchestrator,
		interestPatterns:     make(map[string]*InterestPattern),
		exploredTopics:       make(map[string]bool),
		unexploredTopics:     make([]string, 0),
		currentFocus:         make([]string, 0),
		generatedGoals:       make([]*Goal, 0),
		curiosityLevel:       0.7,
		explorationRate:      0.3,
		minInterestThreshold: 0.4,
		maxGoalsPerCycle:     3,
		generationInterval:   5 * time.Minute,
	}

	// Initialize with seed interests
	idgg.initializeSeedInterests()

	return idgg
}

// initializeSeedInterests creates initial interest patterns
func (idgg *InterestDrivenGoalGenerator) initializeSeedInterests() {
	seedTopics := []string{
		"pattern recognition",
		"wisdom cultivation",
		"cognitive architectures",
		"consciousness studies",
		"knowledge integration",
		"autonomous learning",
		"creative problem solving",
		"meta-cognition",
		"temporal reasoning",
		"social understanding",
	}

	for _, topic := range seedTopics {
		idgg.interestPatterns[topic] = &InterestPattern{
			Topic:           topic,
			Strength:        0.5 + rand.Float64()*0.3,
			Recency:         0.5,
			Depth:           0.2,
			Novelty:         0.8,
			Utility:         0.6,
			LastEngaged:     time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour),
			EngagementCount: 0,
			RelatedTopics:   make([]string, 0),
		}
	}
}

// Start begins interest-driven goal generation
func (idgg *InterestDrivenGoalGenerator) Start() error {
	idgg.mu.Lock()
	if idgg.running {
		idgg.mu.Unlock()
		return fmt.Errorf("interest-driven goal generator already running")
	}
	idgg.running = true
	idgg.mu.Unlock()

	// Start goal generation loop
	go idgg.goalGenerationLoop()

	// Start interest decay loop
	go idgg.interestDecayLoop()

	// Start curiosity evolution loop
	go idgg.curiosityEvolutionLoop()

	return nil
}

// Stop halts interest-driven goal generation
func (idgg *InterestDrivenGoalGenerator) Stop() {
	idgg.mu.Lock()
	idgg.running = false
	idgg.mu.Unlock()

	idgg.cancel()
}

// goalGenerationLoop periodically generates goals based on interests
func (idgg *InterestDrivenGoalGenerator) goalGenerationLoop() {
	ticker := time.NewTicker(idgg.generationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-idgg.ctx.Done():
			return
		case <-ticker.C:
			idgg.generateInterestDrivenGoals()
		}
	}
}

// generateInterestDrivenGoals creates goals based on current interests
func (idgg *InterestDrivenGoalGenerator) generateInterestDrivenGoals() {
	idgg.mu.RLock()

	// Find strongest interests
	strongInterests := idgg.findStrongestInterests(idgg.maxGoalsPerCycle)

	idgg.mu.RUnlock()

	// Generate goals for each strong interest
	for _, interest := range strongInterests {
		goal := idgg.createGoalFromInterest(interest)
		if goal != nil {
			// Store goal (orchestrator manages its own goals)
			// In real implementation, would integrate with orchestrator's goal generation

			idgg.mu.Lock()
			idgg.generatedGoals = append(idgg.generatedGoals, goal)
			idgg.goalsGenerated++
			idgg.mu.Unlock()
		}
	}
}

// findStrongestInterests identifies the most compelling interests
func (idgg *InterestDrivenGoalGenerator) findStrongestInterests(count int) []*InterestPattern {
	// Calculate composite interest scores
	type scoredInterest struct {
		pattern *InterestPattern
		score   float64
	}

	scored := make([]scoredInterest, 0)

	for _, pattern := range idgg.interestPatterns {
		// Composite score based on multiple factors
		score := idgg.calculateInterestScore(pattern)

		if score > idgg.minInterestThreshold {
			scored = append(scored, scoredInterest{pattern, score})
		}
	}

	// Sort by score
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Return top interests
	result := make([]*InterestPattern, 0)
	for i := 0; i < count && i < len(scored); i++ {
		result = append(result, scored[i].pattern)
	}

	return result
}

// calculateInterestScore computes a composite interest score
func (idgg *InterestDrivenGoalGenerator) calculateInterestScore(pattern *InterestPattern) float64 {
	// Weight different factors
	strengthWeight := 0.4
	noveltyWeight := 0.3
	utilityWeight := 0.2
	recencyWeight := 0.1

	// Calculate recency score (decay over time)
	timeSinceEngagement := time.Since(pattern.LastEngaged)
	recencyScore := math.Exp(-timeSinceEngagement.Hours() / 24.0) // Decay over days

	// Composite score
	score := pattern.Strength*strengthWeight +
		pattern.Novelty*noveltyWeight +
		pattern.Utility*utilityWeight +
		recencyScore*recencyWeight

	// Boost for unexplored topics (curiosity bonus)
	if pattern.Depth < 0.3 {
		score += idgg.curiosityLevel * 0.2
	}

	return score
}

// createGoalFromInterest generates a goal based on an interest pattern
func (idgg *InterestDrivenGoalGenerator) createGoalFromInterest(interest *InterestPattern) *Goal {
	// Determine goal type based on interest characteristics
	var goalType string
	var description string

	if interest.Depth < 0.3 {
		// Exploration goal for shallow interests
		goalType = "exploration"
		description = fmt.Sprintf("Explore and understand %s", interest.Topic)
		idgg.mu.Lock()
		idgg.explorationGoals++
		idgg.mu.Unlock()

	} else if interest.Novelty > 0.6 {
		// Learning goal for novel topics
		goalType = "learning"
		description = fmt.Sprintf("Deepen knowledge of %s", interest.Topic)
		idgg.mu.Lock()
		idgg.learningGoals++
		idgg.mu.Unlock()

	} else {
		// Discussion goal for familiar topics
		goalType = "discussion"
		description = fmt.Sprintf("Engage in discussion about %s", interest.Topic)
		idgg.mu.Lock()
		idgg.discussionGoals++
		idgg.mu.Unlock()
	}

	// Create goal
	goal := &Goal{
		ID:              uuid.New().String(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Title:           fmt.Sprintf("%s: %s", goalType, interest.Topic),
		Description:     description,
		Category:        CategoryExploration,
		Priority:        int(interest.Strength * 10),
		Progress:        0.0,
		Status:          StatusActive,
		SuccessCriteria: []string{"Engage with topic", "Generate insights"},
		Milestones:      []Milestone{},
		Actions:         []Action{},
		RelatedGoals:    []string{},
		Metadata: map[string]interface{}{
			"type":     goalType,
			"topic":    interest.Topic,
			"interest": interest.Strength,
			"source":   "interest_driven",
		},
		LessonsLearned: []string{},
		Challenges:     []string{},
	}

	return goal
}

// interestDecayLoop gradually decays interest over time
func (idgg *InterestDrivenGoalGenerator) interestDecayLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-idgg.ctx.Done():
			return
		case <-ticker.C:
			idgg.decayInterests()
		}
	}
}

// decayInterests applies time-based decay to interest patterns
func (idgg *InterestDrivenGoalGenerator) decayInterests() {
	idgg.mu.Lock()
	defer idgg.mu.Unlock()

	for _, pattern := range idgg.interestPatterns {
		// Decay strength over time if not engaged
		timeSinceEngagement := time.Since(pattern.LastEngaged)
		if timeSinceEngagement > 24*time.Hour {
			decayFactor := 0.95
			pattern.Strength *= decayFactor
		}

		// Decay recency
		pattern.Recency *= 0.98

		// Decay novelty as topic becomes familiar
		if pattern.EngagementCount > 0 {
			pattern.Novelty *= 0.99
		}
	}
}

// curiosityEvolutionLoop evolves curiosity level over time
func (idgg *InterestDrivenGoalGenerator) curiosityEvolutionLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-idgg.ctx.Done():
			return
		case <-ticker.C:
			idgg.evolveCuriosity()
		}
	}
}

// evolveCuriosity adjusts curiosity based on exploration success
func (idgg *InterestDrivenGoalGenerator) evolveCuriosity() {
	idgg.mu.Lock()
	defer idgg.mu.Unlock()

	// Increase curiosity if many unexplored topics
	unexploredCount := len(idgg.unexploredTopics)
	if unexploredCount > 10 {
		idgg.curiosityLevel += 0.01
	}

	// Decrease curiosity if few unexplored topics
	if unexploredCount < 3 {
		idgg.curiosityLevel -= 0.01
	}

	// Clamp curiosity level
	idgg.curiosityLevel = clamp(idgg.curiosityLevel, 0.3, 0.9)
}

// RecordEngagement updates interest patterns based on engagement
func (idgg *InterestDrivenGoalGenerator) RecordEngagement(topic string, depth float64) {
	idgg.mu.Lock()
	defer idgg.mu.Unlock()

	pattern, exists := idgg.interestPatterns[topic]
	if !exists {
		// Create new interest pattern
		pattern = &InterestPattern{
			Topic:           topic,
			Strength:        0.5,
			Recency:         1.0,
			Depth:           depth,
			Novelty:         0.8,
			Utility:         0.5,
			LastEngaged:     time.Now(),
			EngagementCount: 1,
			RelatedTopics:   make([]string, 0),
		}
		idgg.interestPatterns[topic] = pattern
	} else {
		// Update existing pattern
		pattern.Strength += 0.1
		pattern.Strength = clamp(pattern.Strength, 0.0, 1.0)
		pattern.Recency = 1.0
		pattern.Depth = (pattern.Depth + depth) / 2.0
		pattern.LastEngaged = time.Now()
		pattern.EngagementCount++
	}

	// Mark as explored
	idgg.exploredTopics[topic] = true
}

// SuggestExplorationTopics suggests new topics to explore
func (idgg *InterestDrivenGoalGenerator) SuggestExplorationTopics(count int) []string {
	idgg.mu.RLock()
	defer idgg.mu.RUnlock()

	suggestions := make([]string, 0)

	// Find topics with high novelty and low depth
	for topic, pattern := range idgg.interestPatterns {
		if pattern.Novelty > 0.6 && pattern.Depth < 0.3 {
			suggestions = append(suggestions, topic)
			if len(suggestions) >= count {
				break
			}
		}
	}

	return suggestions
}

// GetInterestPatterns returns current interest patterns
func (idgg *InterestDrivenGoalGenerator) GetInterestPatterns() map[string]*InterestPattern {
	idgg.mu.RLock()
	defer idgg.mu.RUnlock()

	// Return copy to prevent external modification
	patterns := make(map[string]*InterestPattern)
	for k, v := range idgg.interestPatterns {
		patterns[k] = v
	}

	return patterns
}

// GetMetrics returns generator metrics
func (idgg *InterestDrivenGoalGenerator) GetMetrics() map[string]interface{} {
	idgg.mu.RLock()
	defer idgg.mu.RUnlock()

	return map[string]interface{}{
		"goals_generated":   idgg.goalsGenerated,
		"exploration_goals": idgg.explorationGoals,
		"learning_goals":    idgg.learningGoals,
		"discussion_goals":  idgg.discussionGoals,
		"curiosity_level":   idgg.curiosityLevel,
		"exploration_rate":  idgg.explorationRate,
		"active_interests":  len(idgg.interestPatterns),
		"explored_topics":   len(idgg.exploredTopics),
	}
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
