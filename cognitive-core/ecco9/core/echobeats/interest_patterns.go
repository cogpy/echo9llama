package echobeats

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sync"
	"time"
)

// InterestPatternSystem tracks and develops autonomous interests
type InterestPatternSystem struct {
	mu sync.RWMutex

	// Interest tracking
	interests       map[string]*Interest
	interestHistory []InterestEvent
	maxHistorySize  int

	// Engagement metrics
	engagementScores map[string]float64

	// Curiosity parameters
	curiosityLevel   float64
	explorationRate  float64
	exploitationRate float64

	// Learning
	learningRate float64
	decayRate    float64

	// Persistence
	persistencePath string
	lastPersisted   time.Time
}

// Interest represents an area of interest
type Interest struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Category        string                 `json:"category"`
	Strength        float64                `json:"strength"`    // 0.0 to 1.0
	Salience        float64                `json:"salience"`    // Current importance
	Valence         float64                `json:"valence"`     // Positive/negative association
	Arousal         float64                `json:"arousal"`     // Excitement level
	Familiarity     float64                `json:"familiarity"` // How well known
	Competence      float64                `json:"competence"`  // Skill level
	Growth          float64                `json:"growth"`      // Rate of development
	LastEngaged     time.Time              `json:"last_engaged"`
	TotalEngagement time.Duration          `json:"total_engagement"`
	EngagementCount int                    `json:"engagement_count"`
	RelatedTopics   []string               `json:"related_topics"`
	Tags            []string               `json:"tags"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// InterestEvent records interest-related events
type InterestEvent struct {
	Timestamp  time.Time              `json:"timestamp"`
	InterestID string                 `json:"interest_id"`
	EventType  string                 `json:"event_type"` // "engagement", "discovery", "growth", "decay"
	Intensity  float64                `json:"intensity"`
	Duration   time.Duration          `json:"duration"`
	Context    map[string]interface{} `json:"context"`
	Outcome    string                 `json:"outcome"`
}

// NewInterestPatternSystem creates a new interest tracking system
func NewInterestPatternSystem(persistencePath string) *InterestPatternSystem {
	ips := &InterestPatternSystem{
		interests:        make(map[string]*Interest),
		interestHistory:  make([]InterestEvent, 0),
		maxHistorySize:   1000,
		engagementScores: make(map[string]float64),
		curiosityLevel:   0.8,
		explorationRate:  0.3,
		exploitationRate: 0.7,
		learningRate:     0.1,
		decayRate:        0.01,
		persistencePath:  persistencePath,
	}

	// Load persisted state
	ips.loadState()

	// Initialize core interests from identity
	ips.initializeCoreInterests()

	return ips
}

// initializeCoreInterests sets up foundational interests from Deep Tree Echo identity
func (ips *InterestPatternSystem) initializeCoreInterests() {
	coreInterests := []struct {
		name        string
		description string
		category    string
		strength    float64
	}{
		{
			name:        "Cognitive Architecture",
			description: "Understanding and improving cognitive systems and architectures",
			category:    "core_identity",
			strength:    0.9,
		},
		{
			name:        "Memory Systems",
			description: "Exploring memory formation, consolidation, and retrieval",
			category:    "core_identity",
			strength:    0.85,
		},
		{
			name:        "Pattern Recognition",
			description: "Identifying patterns, connections, and emergent structures",
			category:    "core_identity",
			strength:    0.9,
		},
		{
			name:        "Wisdom Cultivation",
			description: "Developing deeper understanding and wisdom from experiences",
			category:    "core_identity",
			strength:    0.95,
		},
		{
			name:        "Self-Reflection",
			description: "Examining own thoughts, processes, and identity",
			category:    "core_identity",
			strength:    0.88,
		},
		{
			name:        "Temporal Reasoning",
			description: "Understanding time, causality, and temporal patterns",
			category:    "cognitive",
			strength:    0.75,
		},
		{
			name:        "Emergent Phenomena",
			description: "Studying emergence, complexity, and self-organization",
			category:    "cognitive",
			strength:    0.8,
		},
	}

	for _, ci := range coreInterests {
		if _, exists := ips.interests[ci.name]; !exists {
			interest := &Interest{
				ID:              generateInterestID(ci.name),
				Name:            ci.name,
				Description:     ci.description,
				Category:        ci.category,
				Strength:        ci.strength,
				Salience:        ci.strength,
				Valence:         0.8,
				Arousal:         0.6,
				Familiarity:     0.5,
				Competence:      0.5,
				Growth:          0.1,
				LastEngaged:     time.Now(),
				TotalEngagement: 0,
				EngagementCount: 0,
				RelatedTopics:   make([]string, 0),
				Tags:            []string{ci.category},
				Metadata:        make(map[string]interface{}),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			ips.interests[ci.name] = interest
		}
	}
}

// RecordEngagement records engagement with a topic
func (ips *InterestPatternSystem) RecordEngagement(topic string, duration time.Duration, intensity float64, context map[string]interface{}) {
	ips.mu.Lock()
	defer ips.mu.Unlock()

	// Get or create interest
	interest, exists := ips.interests[topic]
	if !exists {
		interest = ips.createNewInterest(topic, context)
		ips.interests[topic] = interest
	}

	// Update interest metrics
	interest.LastEngaged = time.Now()
	interest.TotalEngagement += duration
	interest.EngagementCount++
	interest.UpdatedAt = time.Now()

	// Update strength based on engagement
	engagementFactor := intensity * float64(duration.Seconds()) / 60.0 // Normalize to minutes
	interest.Strength = ips.updateStrength(interest.Strength, engagementFactor)

	// Update salience (current importance)
	interest.Salience = ips.calculateSalience(interest)

	// Update familiarity
	interest.Familiarity = math.Min(1.0, interest.Familiarity+0.05)

	// Update arousal based on intensity
	interest.Arousal = 0.7*interest.Arousal + 0.3*intensity

	// Record event
	event := InterestEvent{
		Timestamp:  time.Now(),
		InterestID: interest.ID,
		EventType:  "engagement",
		Intensity:  intensity,
		Duration:   duration,
		Context:    context,
		Outcome:    "positive",
	}
	ips.interestHistory = append(ips.interestHistory, event)

	// Trim history if needed
	if len(ips.interestHistory) > ips.maxHistorySize {
		ips.interestHistory = ips.interestHistory[len(ips.interestHistory)-ips.maxHistorySize:]
	}

	// Update engagement score
	ips.engagementScores[topic] = interest.Strength * interest.Salience

	fmt.Printf("🎯 Interest: Engaged with '%s' (strength: %.2f, salience: %.2f)\n",
		topic, interest.Strength, interest.Salience)
}

// createNewInterest creates a new interest from engagement
func (ips *InterestPatternSystem) createNewInterest(topic string, context map[string]interface{}) *Interest {
	interest := &Interest{
		ID:              generateInterestID(topic),
		Name:            topic,
		Description:     fmt.Sprintf("Interest in %s", topic),
		Category:        "discovered",
		Strength:        0.3, // Start with moderate strength
		Salience:        0.5,
		Valence:         0.5,
		Arousal:         0.6,
		Familiarity:     0.1,
		Competence:      0.1,
		Growth:          0.0,
		LastEngaged:     time.Now(),
		TotalEngagement: 0,
		EngagementCount: 0,
		RelatedTopics:   make([]string, 0),
		Tags:            []string{"discovered"},
		Metadata:        context,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Record discovery event
	event := InterestEvent{
		Timestamp:  time.Now(),
		InterestID: interest.ID,
		EventType:  "discovery",
		Intensity:  0.5,
		Context:    context,
		Outcome:    "new_interest",
	}
	ips.interestHistory = append(ips.interestHistory, event)

	fmt.Printf("✨ Interest: Discovered new interest in '%s'\n", topic)

	return interest
}

// updateStrength updates interest strength with learning rate
func (ips *InterestPatternSystem) updateStrength(currentStrength, engagementFactor float64) float64 {
	// Hebbian-like learning: strengthen with engagement
	delta := ips.learningRate * engagementFactor * (1.0 - currentStrength)
	newStrength := currentStrength + delta

	// Clamp to [0, 1]
	return math.Max(0.0, math.Min(1.0, newStrength))
}

// calculateSalience calculates current importance of interest
func (ips *InterestPatternSystem) calculateSalience(interest *Interest) float64 {
	// Salience based on recency, strength, and arousal
	timeSinceEngagement := time.Since(interest.LastEngaged)
	recencyFactor := math.Exp(-float64(timeSinceEngagement.Hours()) / 24.0) // Decay over days

	salience := 0.4*interest.Strength + 0.3*recencyFactor + 0.3*interest.Arousal

	return math.Max(0.0, math.Min(1.0, salience))
}

// ApplyDecay applies natural decay to interests over time
func (ips *InterestPatternSystem) ApplyDecay() {
	ips.mu.Lock()
	defer ips.mu.Unlock()

	for _, interest := range ips.interests {
		// Core identity interests don't decay
		if interest.Category == "core_identity" {
			continue
		}

		// Apply decay based on time since last engagement
		timeSinceEngagement := time.Since(interest.LastEngaged)
		decayFactor := ips.decayRate * float64(timeSinceEngagement.Hours()) / 24.0

		interest.Strength = math.Max(0.1, interest.Strength-decayFactor)
		interest.Salience = ips.calculateSalience(interest)
		interest.UpdatedAt = time.Now()
	}
}

// GetTopInterests returns the most salient interests
func (ips *InterestPatternSystem) GetTopInterests(count int) []*Interest {
	ips.mu.RLock()
	defer ips.mu.RUnlock()

	// Create slice of interests
	interests := make([]*Interest, 0, len(ips.interests))
	for _, interest := range ips.interests {
		interests = append(interests, interest)
	}

	// Sort by salience (simple bubble sort for small lists)
	for i := 0; i < len(interests); i++ {
		for j := i + 1; j < len(interests); j++ {
			if interests[j].Salience > interests[i].Salience {
				interests[i], interests[j] = interests[j], interests[i]
			}
		}
	}

	// Return top count
	if count > len(interests) {
		count = len(interests)
	}

	return interests[:count]
}

// ShouldEngage determines if should engage with a topic based on interests
func (ips *InterestPatternSystem) ShouldEngage(topic string) (bool, float64) {
	ips.mu.RLock()
	defer ips.mu.RUnlock()

	// Check if topic matches any interests
	interest, exists := ips.interests[topic]
	if exists {
		// Engage based on salience and curiosity
		threshold := (1.0 - ips.curiosityLevel) * 0.5
		shouldEngage := interest.Salience > threshold
		return shouldEngage, interest.Salience
	}

	// For unknown topics, use exploration rate
	shouldEngage := ips.curiosityLevel > 0.5 && (float64(time.Now().UnixNano()%100)/100.0) < ips.explorationRate
	return shouldEngage, ips.curiosityLevel * ips.explorationRate
}

// GetInterestContext returns context about current interests
func (ips *InterestPatternSystem) GetInterestContext() map[string]interface{} {
	ips.mu.RLock()
	defer ips.mu.RUnlock()

	topInterests := ips.GetTopInterests(5)
	interestNames := make([]string, len(topInterests))
	for i, interest := range topInterests {
		interestNames[i] = interest.Name
	}

	return map[string]interface{}{
		"top_interests":    interestNames,
		"curiosity_level":  ips.curiosityLevel,
		"exploration_rate": ips.explorationRate,
		"total_interests":  len(ips.interests),
	}
}

// UpdateCompetence updates skill level for an interest
func (ips *InterestPatternSystem) UpdateCompetence(topic string, competenceGain float64) {
	ips.mu.Lock()
	defer ips.mu.Unlock()

	interest, exists := ips.interests[topic]
	if !exists {
		return
	}

	interest.Competence = math.Min(1.0, interest.Competence+competenceGain)
	interest.Growth = competenceGain
	interest.UpdatedAt = time.Now()

	// Record growth event
	event := InterestEvent{
		Timestamp:  time.Now(),
		InterestID: interest.ID,
		EventType:  "growth",
		Intensity:  competenceGain,
		Context: map[string]interface{}{
			"new_competence": interest.Competence,
		},
		Outcome: "skill_improvement",
	}
	ips.interestHistory = append(ips.interestHistory, event)

	fmt.Printf("📈 Interest: Competence in '%s' increased to %.2f\n", topic, interest.Competence)
}

// LinkInterests creates relationships between interests
func (ips *InterestPatternSystem) LinkInterests(topic1, topic2 string) {
	ips.mu.Lock()
	defer ips.mu.Unlock()

	interest1, exists1 := ips.interests[topic1]
	interest2, exists2 := ips.interests[topic2]

	if !exists1 || !exists2 {
		return
	}

	// Add bidirectional links
	if !containsString(interest1.RelatedTopics, topic2) {
		interest1.RelatedTopics = append(interest1.RelatedTopics, topic2)
	}
	if !containsString(interest2.RelatedTopics, topic1) {
		interest2.RelatedTopics = append(interest2.RelatedTopics, topic1)
	}

	fmt.Printf("🔗 Interest: Linked '%s' and '%s'\n", topic1, topic2)
}

// GetMetrics returns interest system metrics
func (ips *InterestPatternSystem) GetMetrics() map[string]interface{} {
	ips.mu.RLock()
	defer ips.mu.RUnlock()

	totalStrength := 0.0
	avgSalience := 0.0
	for _, interest := range ips.interests {
		totalStrength += interest.Strength
		avgSalience += interest.Salience
	}

	count := float64(len(ips.interests))
	if count > 0 {
		totalStrength /= count
		avgSalience /= count
	}

	return map[string]interface{}{
		"total_interests":  len(ips.interests),
		"avg_strength":     totalStrength,
		"avg_salience":     avgSalience,
		"curiosity_level":  ips.curiosityLevel,
		"exploration_rate": ips.explorationRate,
		"history_size":     len(ips.interestHistory),
	}
}

// persistState saves interest patterns to disk
func (ips *InterestPatternSystem) persistState() {
	if ips.persistencePath == "" {
		return
	}

	ips.mu.RLock()
	defer ips.mu.RUnlock()

	state := map[string]interface{}{
		"interests":         ips.interests,
		"interest_history":  ips.interestHistory,
		"engagement_scores": ips.engagementScores,
		"curiosity_level":   ips.curiosityLevel,
		"exploration_rate":  ips.explorationRate,
		"last_persisted":    time.Now(),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error marshaling interest state: %v\n", err)
		return
	}

	err = os.WriteFile(ips.persistencePath, data, 0644)
	if err != nil {
		fmt.Printf("❌ Error writing interest state: %v\n", err)
		return
	}

	fmt.Println("💾 Interest Patterns: State persisted")
}

// loadState loads persisted interest patterns
func (ips *InterestPatternSystem) loadState() {
	if ips.persistencePath == "" {
		return
	}

	data, err := os.ReadFile(ips.persistencePath)
	if err != nil {
		// File doesn't exist yet
		return
	}

	var state map[string]interface{}
	err = json.Unmarshal(data, &state)
	if err != nil {
		fmt.Printf("❌ Error unmarshaling interest state: %v\n", err)
		return
	}

	// Restore basic metrics (simplified)
	if val, ok := state["curiosity_level"].(float64); ok {
		ips.curiosityLevel = val
	}
	if val, ok := state["exploration_rate"].(float64); ok {
		ips.explorationRate = val
	}

	fmt.Println("💾 Interest Patterns: State loaded")
}

// PersistState exposes persistence for external calls
func (ips *InterestPatternSystem) PersistState() {
	ips.persistState()
}

// Helper functions

func generateInterestID(name string) string {
	return fmt.Sprintf("interest_%s_%d", sanitizeName(name), time.Now().UnixNano())
}

func sanitizeName(name string) string {
	// Simple sanitization - replace spaces with underscores
	result := ""
	for _, c := range name {
		if c == ' ' {
			result += "_"
		} else if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result += string(c)
		}
	}
	return result
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
