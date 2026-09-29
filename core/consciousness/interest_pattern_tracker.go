package consciousness

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// InterestPatternTracker tracks and scores topics, domains, and concepts of interest
type InterestPatternTracker struct {
	mu sync.RWMutex

	// Interest maps
	topics   map[string]*InterestScore
	domains  map[string]*InterestScore
	concepts map[string]*InterestScore
	skills   map[string]*InterestScore

	// Scoring parameters
	recencyWeight   float64
	frequencyWeight float64
	depthWeight     float64
	noveltyWeight   float64

	// Decay parameters
	decayRate float64
	lastDecay time.Time

	// Metrics
	totalInteractions uint64
	topicsTracked     int
}

// InterestScore represents the interest level in a topic
type InterestScore struct {
	Name     string
	Category string // "topic", "domain", "concept", "skill"
	Score    float64

	// Scoring components
	Recency   float64
	Frequency int
	Depth     float64
	Novelty   float64

	// Temporal tracking
	FirstSeen    time.Time
	LastSeen     time.Time
	Interactions []time.Time

	// Context
	RelatedTopics []string
	Tags          []string
}

// NewInterestPatternTracker creates a new interest tracker
func NewInterestPatternTracker() *InterestPatternTracker {
	return &InterestPatternTracker{
		topics:          make(map[string]*InterestScore),
		domains:         make(map[string]*InterestScore),
		concepts:        make(map[string]*InterestScore),
		skills:          make(map[string]*InterestScore),
		recencyWeight:   0.4,
		frequencyWeight: 0.3,
		depthWeight:     0.2,
		noveltyWeight:   0.1,
		decayRate:       0.95, // 5% decay per day
		lastDecay:       time.Now(),
	}
}

// RecordInterest records an interaction with a topic
func (ipt *InterestPatternTracker) RecordInterest(category, name string, depth float64, tags []string) {
	ipt.mu.Lock()
	defer ipt.mu.Unlock()

	var interestMap map[string]*InterestScore
	switch category {
	case "topic":
		interestMap = ipt.topics
	case "domain":
		interestMap = ipt.domains
	case "concept":
		interestMap = ipt.concepts
	case "skill":
		interestMap = ipt.skills
	default:
		interestMap = ipt.topics
	}

	now := time.Now()

	if score, exists := interestMap[name]; exists {
		// Update existing interest
		score.Frequency++
		score.LastSeen = now
		score.Interactions = append(score.Interactions, now)
		score.Depth = (score.Depth + depth) / 2.0 // Running average

		// Update tags
		for _, tag := range tags {
			if !containsString(score.Tags, tag) {
				score.Tags = append(score.Tags, tag)
			}
		}

		// Recalculate score
		score.Score = ipt.calculateScore(score)
	} else {
		// Create new interest
		score := &InterestScore{
			Name:          name,
			Category:      category,
			Frequency:     1,
			Depth:         depth,
			Novelty:       1.0, // New topics are novel
			FirstSeen:     now,
			LastSeen:      now,
			Interactions:  []time.Time{now},
			RelatedTopics: make([]string, 0),
			Tags:          tags,
		}

		score.Score = ipt.calculateScore(score)
		interestMap[name] = score
		ipt.topicsTracked++
	}

	ipt.totalInteractions++
}

// calculateScore computes the overall interest score
func (ipt *InterestPatternTracker) calculateScore(score *InterestScore) float64 {
	// Recency: How recently was this topic engaged?
	timeSinceLastSeen := time.Since(score.LastSeen).Hours()
	recency := math.Exp(-timeSinceLastSeen / 24.0) // Exponential decay over days
	score.Recency = recency

	// Frequency: How often has this topic been engaged?
	// Normalize by log to prevent dominance
	frequency := math.Log(float64(score.Frequency)+1.0) / 5.0
	if frequency > 1.0 {
		frequency = 1.0
	}

	// Depth: How deeply has this topic been explored?
	depth := score.Depth
	if depth > 1.0 {
		depth = 1.0
	}

	// Novelty: How new is this topic?
	daysSinceFirst := time.Since(score.FirstSeen).Hours() / 24.0
	novelty := math.Exp(-daysSinceFirst / 7.0) // Decays over a week
	score.Novelty = novelty

	// Weighted combination
	totalScore := ipt.recencyWeight*recency +
		ipt.frequencyWeight*frequency +
		ipt.depthWeight*depth +
		ipt.noveltyWeight*novelty

	return totalScore
}

// GetInterestScore returns the interest score for a topic
func (ipt *InterestPatternTracker) GetInterestScore(category, name string) float64 {
	ipt.mu.RLock()
	defer ipt.mu.RUnlock()

	var interestMap map[string]*InterestScore
	switch category {
	case "topic":
		interestMap = ipt.topics
	case "domain":
		interestMap = ipt.domains
	case "concept":
		interestMap = ipt.concepts
	case "skill":
		interestMap = ipt.skills
	default:
		return 0.0
	}

	if score, exists := interestMap[name]; exists {
		return score.Score
	}

	return 0.0
}

// GetTopInterests returns the top N interests by score
func (ipt *InterestPatternTracker) GetTopInterests(category string, n int) []*InterestScore {
	ipt.mu.RLock()
	defer ipt.mu.RUnlock()

	var interestMap map[string]*InterestScore
	switch category {
	case "topic":
		interestMap = ipt.topics
	case "domain":
		interestMap = ipt.domains
	case "concept":
		interestMap = ipt.concepts
	case "skill":
		interestMap = ipt.skills
	case "all":
		// Combine all categories
		combined := make(map[string]*InterestScore)
		for k, v := range ipt.topics {
			combined[k] = v
		}
		for k, v := range ipt.domains {
			combined[k] = v
		}
		for k, v := range ipt.concepts {
			combined[k] = v
		}
		for k, v := range ipt.skills {
			combined[k] = v
		}
		interestMap = combined
	default:
		return []*InterestScore{}
	}

	// Convert to slice
	scores := make([]*InterestScore, 0, len(interestMap))
	for _, score := range interestMap {
		scores = append(scores, score)
	}

	// Sort by score (descending)
	for i := range len(scores) - 1 {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].Score > scores[i].Score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// Return top N
	if n > len(scores) {
		n = len(scores)
	}

	return scores[:n]
}

// IsInterested checks if interest score exceeds threshold
func (ipt *InterestPatternTracker) IsInterested(category, name string, threshold float64) bool {
	score := ipt.GetInterestScore(category, name)
	return score >= threshold
}

// ApplyDecay applies time-based decay to all interest scores
func (ipt *InterestPatternTracker) ApplyDecay() {
	ipt.mu.Lock()
	defer ipt.mu.Unlock()

	now := time.Now()
	daysSinceLastDecay := now.Sub(ipt.lastDecay).Hours() / 24.0

	if daysSinceLastDecay < 1.0 {
		return // Don't decay more than once per day
	}

	decayFactor := math.Pow(ipt.decayRate, daysSinceLastDecay)

	// Apply decay to all categories
	ipt.applyDecayToMap(ipt.topics, decayFactor)
	ipt.applyDecayToMap(ipt.domains, decayFactor)
	ipt.applyDecayToMap(ipt.concepts, decayFactor)
	ipt.applyDecayToMap(ipt.skills, decayFactor)

	ipt.lastDecay = now

	fmt.Printf("🔄 Applied interest decay (factor: %.3f)\n", decayFactor)
}

// applyDecayToMap applies decay to a specific interest map
func (ipt *InterestPatternTracker) applyDecayToMap(interestMap map[string]*InterestScore, decayFactor float64) {
	for name, score := range interestMap {
		score.Score *= decayFactor

		// Remove very low interest items
		if score.Score < 0.01 {
			delete(interestMap, name)
			ipt.topicsTracked--
		}
	}
}

// LinkTopics creates a relationship between two topics
func (ipt *InterestPatternTracker) LinkTopics(topic1, topic2 string) {
	ipt.mu.Lock()
	defer ipt.mu.Unlock()

	if score1, exists := ipt.topics[topic1]; exists {
		if !containsString(score1.RelatedTopics, topic2) {
			score1.RelatedTopics = append(score1.RelatedTopics, topic2)
		}
	}

	if score2, exists := ipt.topics[topic2]; exists {
		if !containsString(score2.RelatedTopics, topic1) {
			score2.RelatedTopics = append(score2.RelatedTopics, topic1)
		}
	}
}

// GetRelatedTopics returns topics related to a given topic
func (ipt *InterestPatternTracker) GetRelatedTopics(topic string) []string {
	ipt.mu.RLock()
	defer ipt.mu.RUnlock()

	if score, exists := ipt.topics[topic]; exists {
		return score.RelatedTopics
	}

	return []string{}
}

// GetMetrics returns interest tracking metrics
func (ipt *InterestPatternTracker) GetMetrics() map[string]interface{} {
	ipt.mu.RLock()
	defer ipt.mu.RUnlock()

	return map[string]interface{}{
		"total_interactions": ipt.totalInteractions,
		"topics_tracked":     ipt.topicsTracked,
		"topics_count":       len(ipt.topics),
		"domains_count":      len(ipt.domains),
		"concepts_count":     len(ipt.concepts),
		"skills_count":       len(ipt.skills),
	}
}

// ExportInterests exports all interests for persistence
func (ipt *InterestPatternTracker) ExportInterests() map[string]map[string]*InterestScore {
	ipt.mu.RLock()
	defer ipt.mu.RUnlock()

	return map[string]map[string]*InterestScore{
		"topics":   ipt.topics,
		"domains":  ipt.domains,
		"concepts": ipt.concepts,
		"skills":   ipt.skills,
	}
}

// ImportInterests imports interests from persistence
func (ipt *InterestPatternTracker) ImportInterests(data map[string]map[string]*InterestScore) {
	ipt.mu.Lock()
	defer ipt.mu.Unlock()

	if topics, ok := data["topics"]; ok {
		ipt.topics = topics
	}
	if domains, ok := data["domains"]; ok {
		ipt.domains = domains
	}
	if concepts, ok := data["concepts"]; ok {
		ipt.concepts = concepts
	}
	if skills, ok := data["skills"]; ok {
		ipt.skills = skills
	}

	// Recalculate total topics tracked
	ipt.topicsTracked = len(ipt.topics) + len(ipt.domains) + len(ipt.concepts) + len(ipt.skills)
}

// Helper function to check if slice contains string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
