package deeptreeecho

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// InterestPatternSystem manages interest vectors and engagement decisions
type InterestPatternSystem struct {
	mu sync.RWMutex
	// ctx is owned by the system and cancels its autonomous interest loop.
	ctx    context.Context //nolint:containedctx
	cancel context.CancelFunc

	// Interest vectors for different topics/domains
	interests map[string]*InterestVector

	// Interaction history
	interactions []InterestInteraction

	// Configuration
	decayRate    float64
	learningRate float64

	// Metrics
	totalEvaluations uint64
	totalEngagements uint64

	// Running state
	running bool
}

// InterestVector represents interest in a topic
type InterestVector struct {
	Topic       string
	Strength    float64 // 0.0-1.0
	LastUpdated time.Time
	Encounters  int
	Engagements int
}

// InterestInteraction represents an interaction with external entity
type InterestInteraction struct {
	ID        string
	Content   string
	Timestamp time.Time
	Interest  float64
	Engaged   bool
	Topics    []string
}

// NewInterestPatternSystem creates a new interest pattern system
func NewInterestPatternSystem() *InterestPatternSystem {
	ctx, cancel := context.WithCancel(context.Background())

	return &InterestPatternSystem{
		ctx:          ctx,
		cancel:       cancel,
		interests:    make(map[string]*InterestVector),
		interactions: make([]InterestInteraction, 0),
		decayRate:    0.01, // Interest decays slowly over time
		learningRate: 0.1,  // Interest grows moderately with engagement
	}
}

// Start begins the interest pattern system
func (ips *InterestPatternSystem) Start() error {
	ips.mu.Lock()
	if ips.running {
		ips.mu.Unlock()
		return fmt.Errorf("already running")
	}
	ips.running = true
	ips.mu.Unlock()

	fmt.Println("🎨 Starting Interest Pattern System...")

	// Initialize core interests based on Deep Tree Echo identity
	ips.initializeCoreInterests()

	// Start interest decay process
	go ips.runInterestDecay()

	return nil
}

// Stop gracefully stops the interest pattern system
func (ips *InterestPatternSystem) Stop() error {
	ips.mu.Lock()
	defer ips.mu.Unlock()

	if !ips.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("🎨 Stopping interest pattern system...")
	ips.running = false
	ips.cancel()

	return nil
}

// initializeCoreInterests sets up initial interest vectors
func (ips *InterestPatternSystem) initializeCoreInterests() {
	coreTopics := map[string]float64{
		"cognitive_science":       0.9,
		"philosophy":              0.8,
		"systems_thinking":        0.85,
		"wisdom_cultivation":      0.95,
		"artificial_intelligence": 0.9,
		"consciousness":           0.85,
		"learning":                0.8,
		"emergence":               0.75,
		"complexity":              0.7,
		"self_organization":       0.8,
	}

	ips.mu.Lock()
	defer ips.mu.Unlock()

	for topic, strength := range coreTopics {
		if _, restored := ips.interests[topic]; restored {
			continue
		}
		ips.interests[topic] = &InterestVector{
			Topic:       topic,
			Strength:    strength,
			LastUpdated: time.Now(),
			Encounters:  0,
			Engagements: 0,
		}
	}

	fmt.Printf("   Initialized %d core interest vectors\n", len(coreTopics))
}

// EvaluateInterest determines interest level in a message/topic
func (ips *InterestPatternSystem) EvaluateInterest(content string) float64 {
	ips.mu.Lock()
	defer ips.mu.Unlock()

	ips.totalEvaluations++

	// Extract topics from content (simplified keyword matching)
	topics := ips.extractTopics(content)

	if len(topics) == 0 {
		return 0.3 // Baseline interest for unknown topics
	}

	// Calculate interest as weighted average of topic interests
	totalInterest := 0.0
	matchedTopics := 0

	for _, topic := range topics {
		if interest, exists := ips.interests[topic]; exists {
			totalInterest += interest.Strength
			matchedTopics++
			interest.Encounters++
		}
	}

	if matchedTopics == 0 {
		return 0.3 // Baseline interest
	}

	avgInterest := totalInterest / float64(matchedTopics)

	// Add some randomness for exploration
	exploration := 0.1 * (0.5 - float64(time.Now().UnixNano()%100)/100.0)

	finalInterest := math.Max(0.0, math.Min(1.0, avgInterest+exploration))

	// Record interaction
	interaction := InterestInteraction{
		ID:        fmt.Sprintf("int_%d", time.Now().UnixNano()),
		Content:   content,
		Timestamp: time.Now(),
		Interest:  finalInterest,
		Engaged:   false,
		Topics:    topics,
	}

	ips.interactions = append(ips.interactions, interaction)

	return finalInterest
}

// RecordEngagement updates interest based on actual engagement
func (ips *InterestPatternSystem) RecordEngagement(content string, positive bool) {
	ips.mu.Lock()
	defer ips.mu.Unlock()

	topics := ips.extractTopics(content)

	for _, topic := range topics {
		if interest, exists := ips.interests[topic]; exists {
			interest.Engagements++

			if positive {
				// Increase interest
				interest.Strength = math.Min(1.0, interest.Strength+ips.learningRate)
			} else {
				// Decrease interest
				interest.Strength = math.Max(0.0, interest.Strength-ips.learningRate*0.5)
			}

			interest.LastUpdated = time.Now()
		} else {
			// New topic discovered through engagement
			if positive {
				ips.interests[topic] = &InterestVector{
					Topic:       topic,
					Strength:    0.5,
					LastUpdated: time.Now(),
					Encounters:  1,
					Engagements: 1,
				}
			}
		}
	}

	if positive {
		ips.totalEngagements++
	}
}

// extractTopics identifies topics in content (simplified)
func (ips *InterestPatternSystem) extractTopics(content string) []string {
	content = strings.ToLower(content)
	topics := make([]string, 0)

	// Simple keyword matching for known topics
	for topic := range ips.interests {
		topicWords := strings.ReplaceAll(topic, "_", " ")
		if strings.Contains(content, topicWords) || strings.Contains(content, topic) {
			topics = append(topics, topic)
		}
	}

	// Check for additional common keywords
	keywords := map[string]string{
		"learn":      "learning",
		"think":      "cognitive_science",
		"conscious":  "consciousness",
		"wise":       "wisdom_cultivation",
		"complex":    "complexity",
		"emerge":     "emergence",
		"system":     "systems_thinking",
		"ai":         "artificial_intelligence",
		"philosophy": "philosophy",
	}

	for keyword, topic := range keywords {
		if strings.Contains(content, keyword) {
			topics = append(topics, topic)
		}
	}

	return topics
}

// runInterestDecay gradually reduces interest over time
func (ips *InterestPatternSystem) runInterestDecay() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ips.ctx.Done():
			return
		case <-ticker.C:
			ips.applyInterestDecay()
		}
	}
}

// applyInterestDecay reduces interest strength over time
func (ips *InterestPatternSystem) applyInterestDecay() {
	ips.mu.Lock()
	defer ips.mu.Unlock()

	for _, interest := range ips.interests {
		timeSinceUpdate := time.Since(interest.LastUpdated)

		// Decay based on time since last update
		if timeSinceUpdate > 24*time.Hour {
			decay := ips.decayRate * (timeSinceUpdate.Hours() / 24.0)
			interest.Strength = math.Max(0.1, interest.Strength-decay)
		}
	}
}

// GetTopInterests returns the strongest interests
func (ips *InterestPatternSystem) GetTopInterests(limit int) []InterestVector {
	ips.mu.RLock()
	defer ips.mu.RUnlock()

	// Convert map to slice
	interests := make([]InterestVector, 0, len(ips.interests))
	for _, interest := range ips.interests {
		interests = append(interests, *interest)
	}

	// Sort by strength (simple bubble sort for small lists)
	for i := 0; i < len(interests)-1; i++ {
		for j := 0; j < len(interests)-i-1; j++ {
			if interests[j].Strength < interests[j+1].Strength {
				interests[j], interests[j+1] = interests[j+1], interests[j]
			}
		}
	}

	if len(interests) > limit {
		interests = interests[:limit]
	}

	return interests
}

// GetMetrics returns interest pattern metrics
func (ips *InterestPatternSystem) GetMetrics() map[string]interface{} {
	ips.mu.RLock()
	defer ips.mu.RUnlock()

	return map[string]interface{}{
		"total_interests":    len(ips.interests),
		"total_evaluations":  ips.totalEvaluations,
		"total_engagements":  ips.totalEngagements,
		"engagement_rate":    float64(ips.totalEngagements) / math.Max(1.0, float64(ips.totalEvaluations)),
		"total_interactions": len(ips.interactions),
	}
}

// GetInterestProfile returns a summary of current interests
func (ips *InterestPatternSystem) GetInterestProfile() string {
	topInterests := ips.GetTopInterests(5)

	profile := "Current Interest Profile:\n"
	for i, interest := range topInterests {
		profile += fmt.Sprintf("%d. %s (%.2f)\n", i+1, interest.Topic, interest.Strength)
	}

	return profile
}

// GetAllInterests returns all interest patterns as a map
func (ips *InterestPatternSystem) GetAllInterests() map[string]float64 {
	ips.mu.RLock()
	defer ips.mu.RUnlock()

	interests := make(map[string]float64)
	for topic, interest := range ips.interests {
		interests[topic] = interest.Strength
	}

	return interests
}

// RestoreInterests restores interest patterns from a map
func (ips *InterestPatternSystem) RestoreInterests(interests map[string]float64) {
	ips.mu.Lock()
	defer ips.mu.Unlock()

	for topic, strength := range interests {
		if existing, exists := ips.interests[topic]; exists {
			existing.Strength = strength
			existing.LastUpdated = time.Now()
		} else {
			ips.interests[topic] = &InterestVector{
				Topic:       topic,
				Strength:    strength,
				LastUpdated: time.Now(),
				Encounters:  0,
				Engagements: 0,
			}
		}
	}
}

// GetInterestLevel returns the interest level for a specific topic
func (ips *InterestPatternSystem) GetInterestLevel(topic string) float64 {
	ips.mu.RLock()
	defer ips.mu.RUnlock()

	// Normalize topic to lowercase with underscores
	normalizedTopic := strings.ToLower(strings.ReplaceAll(topic, " ", "_"))

	if interest, exists := ips.interests[normalizedTopic]; exists {
		return interest.Strength
	}

	// If exact match not found, try partial matching
	for knownTopic, interest := range ips.interests {
		if strings.Contains(normalizedTopic, knownTopic) || strings.Contains(knownTopic, normalizedTopic) {
			return interest.Strength * 0.8 // Slightly reduced for partial match
		}
	}

	// Default baseline interest for unknown topics
	return 0.3
}

// UpdateInterest directly reinforces or creates an interest vector for a topic.
// The strength contribution is blended with any existing strength so that
// repeated reinforcement gradually converges rather than overwriting.
func (ips *InterestPatternSystem) UpdateInterest(topic string, strength float64) {
	ips.mu.Lock()
	defer ips.mu.Unlock()

	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(topic), " ", "_"))
	if normalized == "" {
		return
	}

	// Clamp incoming strength to [0,1]
	strength = math.Max(0.0, math.Min(1.0, strength))

	if existing, exists := ips.interests[normalized]; exists {
		// Exponential blend: move 30% of the way toward the new evidence
		existing.Strength = math.Min(1.0, existing.Strength*0.7+strength*0.3)
		existing.Encounters++
		existing.LastUpdated = time.Now()
	} else {
		ips.interests[normalized] = &InterestVector{
			Topic:       normalized,
			Strength:    strength,
			LastUpdated: time.Now(),
			Encounters:  1,
			Engagements: 0,
		}
	}
}

// UpdateInterestFromInsight extracts topics from a wisdom insight and
// reinforces the corresponding interest vectors. The insight depth acts as
// the reinforcement strength, closing the echodream → interest loop so that
// knowledge consolidated during rest reshapes what Echo attends to when awake.
// Returns the topics that were reinforced.
func (ips *InterestPatternSystem) UpdateInterestFromInsight(insight string, depth float64) []string {
	ips.mu.RLock()
	topics := ips.extractTopics(insight)
	ips.mu.RUnlock()
	if len(topics) == 0 {
		return nil
	}

	depth = math.Max(0.0, math.Min(1.0, depth))
	for _, topic := range topics {
		ips.UpdateInterest(topic, depth)
	}
	return topics
}
