package echobeats

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// InterestEvolutionEngine handles adaptive learning and evolution of interest patterns
type InterestEvolutionEngine struct {
	mu sync.RWMutex

	// Evolution parameters
	reinforcementRate float64
	extinctionRate    float64
	mutationRate      float64
	crossoverRate     float64

	// Learning history
	engagementOutcomes map[string]*EngagementOutcome
	rewardHistory      []RewardSignal

	// Pattern discovery
	discoveredPatterns map[string]*InterestCluster
	emergentInterests  []*Interest

	// Metrics
	totalEvolutions uint64
	totalMutations  uint64
	totalCrossovers uint64
}

// EngagementOutcome tracks the result of engaging with an interest
type EngagementOutcome struct {
	InterestID   string
	Timestamp    time.Time
	Duration     time.Duration
	Reward       float64 // Positive or negative reinforcement
	LearningGain float64 // How much was learned
	Satisfaction float64 // Emotional satisfaction
	Competence   float64 // Skill improvement
	NoveltyValue float64 // How novel the experience was
	Feedback     string
}

// RewardSignal represents reinforcement learning feedback
type RewardSignal struct {
	Timestamp  time.Time
	InterestID string
	Reward     float64
	Source     string // "intrinsic", "extrinsic", "social"
	Context    map[string]interface{}
}

// InterestCluster represents related interests discovered through pattern analysis
type InterestCluster struct {
	ID            string
	CoreTopic     string
	RelatedTopics []string
	Strength      float64
	Coherence     float64
	Members       []*Interest
	Centroid      []float64 // Feature vector centroid
	CreatedAt     time.Time
	LastUpdated   time.Time
}

// NewInterestEvolutionEngine creates a new interest evolution engine
func NewInterestEvolutionEngine() *InterestEvolutionEngine {
	return &InterestEvolutionEngine{
		reinforcementRate:  0.1,
		extinctionRate:     0.05,
		mutationRate:       0.15,
		crossoverRate:      0.25,
		engagementOutcomes: make(map[string]*EngagementOutcome),
		rewardHistory:      make([]RewardSignal, 0),
		discoveredPatterns: make(map[string]*InterestCluster),
		emergentInterests:  make([]*Interest, 0),
	}
}

// ApplyReinforcement adjusts interest strength based on engagement outcomes
func (iee *InterestEvolutionEngine) ApplyReinforcement(interest *Interest, outcome *EngagementOutcome) {
	iee.mu.Lock()
	defer iee.mu.Unlock()

	// Store outcome
	iee.engagementOutcomes[outcome.InterestID] = outcome

	// Calculate total reward
	totalReward := outcome.Reward +
		(outcome.LearningGain * 0.3) +
		(outcome.Satisfaction * 0.2) +
		(outcome.NoveltyValue * 0.15)

	// Apply reinforcement learning update
	// Q-learning inspired: new_value = old_value + learning_rate * (reward + discount * max_future_value - old_value)
	learningRate := iee.reinforcementRate

	// Positive reinforcement
	if totalReward > 0 {
		interest.Strength = interest.Strength + learningRate*totalReward*(1.0-interest.Strength)
		interest.Salience += learningRate * totalReward * 0.5
		interest.Arousal += learningRate * outcome.NoveltyValue * 0.3
	} else {
		// Negative reinforcement (extinction)
		extinctionFactor := iee.extinctionRate * math.Abs(totalReward)
		interest.Strength *= (1.0 - extinctionFactor)
		interest.Salience *= (1.0 - extinctionFactor*0.5)
	}

	// Update competence based on learning gain
	interest.Competence += outcome.Competence * learningRate

	// Update familiarity
	interest.Familiarity = math.Min(1.0, interest.Familiarity+0.05)

	// Adjust growth rate
	interest.Growth = calculateGrowthRate(interest, outcome)

	// Ensure bounds [0, 1]
	interest.Strength = math.Max(0.0, math.Min(1.0, interest.Strength))
	interest.Salience = math.Max(0.0, math.Min(1.0, interest.Salience))
	interest.Arousal = math.Max(0.0, math.Min(1.0, interest.Arousal))
	interest.Competence = math.Max(0.0, math.Min(1.0, interest.Competence))

	interest.UpdatedAt = time.Now()

	iee.totalEvolutions++
}

// DiscoverClusters finds related interest patterns through clustering
func (iee *InterestEvolutionEngine) DiscoverClusters(interests []*Interest) []*InterestCluster {
	iee.mu.Lock()
	defer iee.mu.Unlock()

	if len(interests) < 2 {
		return nil
	}

	// Simple clustering based on shared topics and engagement patterns
	clusters := make([]*InterestCluster, 0)
	visited := make(map[string]bool)

	for _, interest := range interests {
		if visited[interest.ID] {
			continue
		}

		cluster := &InterestCluster{
			ID:            generateClusterID(),
			CoreTopic:     interest.Name,
			RelatedTopics: make([]string, 0),
			Members:       []*Interest{interest},
			CreatedAt:     time.Now(),
			LastUpdated:   time.Now(),
		}

		visited[interest.ID] = true

		// Find related interests
		for _, other := range interests {
			if visited[other.ID] {
				continue
			}

			similarity := calculateInterestSimilarity(interest, other)
			if similarity > 0.6 {
				cluster.Members = append(cluster.Members, other)
				cluster.RelatedTopics = append(cluster.RelatedTopics, other.Name)
				visited[other.ID] = true
			}
		}

		// Only keep clusters with multiple members
		if len(cluster.Members) > 1 {
			cluster.Strength = calculateClusterStrength(cluster)
			cluster.Coherence = calculateClusterCoherence(cluster)
			clusters = append(clusters, cluster)
			iee.discoveredPatterns[cluster.ID] = cluster
		}
	}

	return clusters
}

// GenerateEmergentInterest creates new interests through combination and mutation
func (iee *InterestEvolutionEngine) GenerateEmergentInterest(parent1, parent2 *Interest) *Interest {
	iee.mu.Lock()
	defer iee.mu.Unlock()

	// Crossover: combine features from both parents
	emergent := &Interest{
		ID:            generateInterestID("emergent_" + parent1.Name + "_" + parent2.Name),
		Name:          synthesizeInterestName(parent1.Name, parent2.Name),
		Description:   "Emergent interest from " + parent1.Name + " and " + parent2.Name,
		Category:      "emergent",
		Strength:      (parent1.Strength + parent2.Strength) / 2.0,
		Salience:      (parent1.Salience + parent2.Salience) / 2.0,
		Valence:       (parent1.Valence + parent2.Valence) / 2.0,
		Arousal:       math.Max(parent1.Arousal, parent2.Arousal),
		Familiarity:   0.0,                                             // New interest starts unfamiliar
		Competence:    (parent1.Competence + parent2.Competence) / 3.0, // Lower competence for new area
		Growth:        0.5,                                             // Moderate initial growth
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		LastEngaged:   time.Now(),
		RelatedTopics: append(parent1.RelatedTopics, parent2.RelatedTopics...),
		Tags:          append(parent1.Tags, parent2.Tags...),
		Metadata:      make(map[string]interface{}),
	}

	// Apply mutation
	if shouldMutate(iee.mutationRate) {
		applyMutation(emergent)
		iee.totalMutations++
	}

	iee.emergentInterests = append(iee.emergentInterests, emergent)
	iee.totalCrossovers++

	return emergent
}

// ConsolidateDuringRest performs interest consolidation during rest/dream cycles
func (iee *InterestEvolutionEngine) ConsolidateDuringRest(interests []*Interest) {
	iee.mu.Lock()
	defer iee.mu.Unlock()

	// Strengthen important patterns
	for _, interest := range interests {
		if interest.Strength > 0.7 && interest.EngagementCount > 5 {
			// Memory consolidation strengthens important interests
			interest.Strength = math.Min(1.0, interest.Strength*1.05)
		}
	}

	// Discover new clusters
	clusters := iee.DiscoverClusters(interests)

	// Generate emergent interests from strong clusters
	for _, cluster := range clusters {
		if cluster.Strength > 0.8 && len(cluster.Members) >= 2 {
			// Create emergent interest from cluster
			if len(cluster.Members) >= 2 {
				iee.GenerateEmergentInterest(cluster.Members[0], cluster.Members[1])
			}
		}
	}

	// Prune weak interests
	pruneWeakInterests(interests)
}

// Helper functions

func calculateGrowthRate(interest *Interest, outcome *EngagementOutcome) float64 {
	// Growth rate based on learning gain and competence
	baseGrowth := outcome.LearningGain * 0.5

	// Novelty boosts growth
	noveltyBonus := outcome.NoveltyValue * 0.3

	// Competence affects growth (less growth at high competence)
	competenceFactor := 1.0 - (interest.Competence * 0.5)

	return math.Min(1.0, (baseGrowth+noveltyBonus)*competenceFactor)
}

func calculateInterestSimilarity(i1, i2 *Interest) float64 {
	// Simple similarity based on shared topics and categories
	similarity := 0.0

	// Category match
	if i1.Category == i2.Category {
		similarity += 0.3
	}

	// Shared related topics
	sharedTopics := 0
	for _, t1 := range i1.RelatedTopics {
		for _, t2 := range i2.RelatedTopics {
			if t1 == t2 {
				sharedTopics++
			}
		}
	}

	if len(i1.RelatedTopics) > 0 && len(i2.RelatedTopics) > 0 {
		topicSim := float64(sharedTopics) / float64(len(i1.RelatedTopics)+len(i2.RelatedTopics))
		similarity += topicSim * 0.4
	}

	// Strength similarity
	strengthSim := 1.0 - math.Abs(i1.Strength-i2.Strength)
	similarity += strengthSim * 0.3

	return similarity
}

func calculateClusterStrength(cluster *InterestCluster) float64 {
	if len(cluster.Members) == 0 {
		return 0.0
	}

	totalStrength := 0.0
	for _, member := range cluster.Members {
		totalStrength += member.Strength
	}

	return totalStrength / float64(len(cluster.Members))
}

func calculateClusterCoherence(cluster *InterestCluster) float64 {
	if len(cluster.Members) < 2 {
		return 1.0
	}

	// Calculate average pairwise similarity
	totalSimilarity := 0.0
	pairs := 0

	for i := 0; i < len(cluster.Members); i++ {
		for j := i + 1; j < len(cluster.Members); j++ {
			totalSimilarity += calculateInterestSimilarity(cluster.Members[i], cluster.Members[j])
			pairs++
		}
	}

	if pairs == 0 {
		return 1.0
	}

	return totalSimilarity / float64(pairs)
}

func synthesizeInterestName(name1, name2 string) string {
	// Simple synthesis - in practice, could use LLM to generate creative names
	return name1 + " & " + name2
}

func shouldMutate(rate float64) bool {
	return math.Floor(math.Mod(float64(time.Now().UnixNano()), 100)) < (rate * 100)
}

func applyMutation(interest *Interest) {
	// Random mutation to parameters
	mutationStrength := 0.1

	// Mutate arousal
	interest.Arousal += (math.Mod(float64(time.Now().UnixNano()), 2.0) - 1.0) * mutationStrength
	interest.Arousal = math.Max(0.0, math.Min(1.0, interest.Arousal))

	// Mutate growth
	interest.Growth += (math.Mod(float64(time.Now().UnixNano()), 2.0) - 1.0) * mutationStrength
	interest.Growth = math.Max(0.0, math.Min(1.0, interest.Growth))
}

func pruneWeakInterests(interests []*Interest) {
	// Mark interests below threshold for removal
	threshold := 0.1

	for _, interest := range interests {
		if interest.Strength < threshold && interest.Category != "core_identity" {
			interest.Strength = 0.0 // Mark for removal
		}
	}
}

func generateClusterID() string {
	return fmt.Sprintf("cluster_%d", time.Now().UnixNano())
}
