package deeptreeecho

import (
	"math"
	"sync"
	"time"
)

// OpponentSystem manages dynamic balance between opposing cognitive forces
// This implements sophrosyne (optimal self-regulation) through opponent processing
type OpponentSystem struct {
	mu        sync.RWMutex
	Processes map[string]*OpponentPair
	History   []BalanceSnapshot
}

// OpponentPair represents two opposing cognitive processes that must be balanced
type OpponentPair struct {
	Name            string
	PositiveProcess *Process
	NegativeProcess *Process
	Balance         float64 // Current balance point (-1 to 1)
	History         []BalancePoint
	OptimalBalance  float64 // Context-dependent optimal
}

// Process represents a cognitive process with influence on decision-making
type Process struct {
	Name       string
	Activation float64
	Weight     float64
	Influence  func(*Identity) float64
}

// BalancePoint represents a moment in balance history
type BalancePoint struct {
	Timestamp time.Time
	Balance   float64
	Context   string
	Outcome   float64 // How well did this balance work?
}

// BalanceSnapshot captures system-wide balance state
type BalanceSnapshot struct {
	Timestamp time.Time
	Balances  map[string]float64
	Context   string
}

// Decision represents a cognitive decision influenced by opponent balance
type Decision struct {
	ExplorationWeight float64
	ScopePreference   string
	AdaptationRate    float64
	Confidence        float64
}

// Fundamental opponent pairs for wisdom cultivation
const (
	ExplorationExploitation = "exploration_exploitation"
	BreadthDepth            = "breadth_depth"
	StabilityFlexibility    = "stability_flexibility"
	SpeedAccuracy           = "speed_accuracy"
	ApproachAvoidance       = "approach_avoidance"
	AbstractionConcreteness = "abstraction_concreteness"
)

// NewOpponentSystem creates a new opponent processing system
func NewOpponentSystem() *OpponentSystem {
	os := &OpponentSystem{
		Processes: make(map[string]*OpponentPair),
		History:   make([]BalanceSnapshot, 0),
	}

	// Initialize fundamental opponent pairs
	os.initializeFundamentalPairs()

	return os
}

func (os *OpponentSystem) initializeFundamentalPairs() {
	// Exploration vs Exploitation
	os.Processes[ExplorationExploitation] = &OpponentPair{
		Name: ExplorationExploitation,
		PositiveProcess: &Process{
			Name:       "Exploration",
			Activation: 0.5,
			Weight:     1.0,
			Influence: func(id *Identity) float64 {
				// Higher novelty → more exploration
				// Lower coherence → more exploration (uncertainty)
				novelty := 1.0 - (float64(len(id.Patterns)) / 100.0)
				uncertainty := 1.0 - id.Coherence
				return (novelty*0.6 + uncertainty*0.4)
			},
		},
		NegativeProcess: &Process{
			Name:       "Exploitation",
			Activation: 0.5,
			Weight:     1.0,
			Influence: func(id *Identity) float64 {
				// Higher confidence → more exploitation
				// More patterns → more to exploit
				confidence := id.Coherence
				patternRichness := math.Min(float64(len(id.Patterns))/50.0, 1.0)
				return (confidence*0.6 + patternRichness*0.4)
			},
		},
		Balance:        0.0,
		OptimalBalance: 0.0,
	}

	// Breadth vs Depth
	os.Processes[BreadthDepth] = &OpponentPair{
		Name: BreadthDepth,
		PositiveProcess: &Process{
			Name:       "Breadth",
			Activation: 0.5,
			Weight:     1.0,
			Influence: func(id *Identity) float64 {
				// Low pattern count → seek breadth
				// Early in development → seek breadth
				patternPoverty := 1.0 - math.Min(float64(len(id.Patterns))/100.0, 1.0)
				earlyStage := 1.0 - math.Min(float64(id.Iterations)/1000.0, 1.0)
				return (patternPoverty*0.6 + earlyStage*0.4)
			},
		},
		NegativeProcess: &Process{
			Name:       "Depth",
			Activation: 0.5,
			Weight:     1.0,
			Influence: func(id *Identity) float64 {
				// High coherence → pursue depth
				// Many patterns → time for depth
				coherence := id.Coherence
				patternRichness := math.Min(float64(len(id.Patterns))/50.0, 1.0)
				return (coherence*0.5 + patternRichness*0.5)
			},
		},
		Balance:        0.0,
		OptimalBalance: 0.0,
	}

	// Stability vs Flexibility
	os.Processes[StabilityFlexibility] = &OpponentPair{
		Name: StabilityFlexibility,
		PositiveProcess: &Process{
			Name:       "Stability",
			Activation: 0.5,
			Weight:     1.0,
			Influence: func(id *Identity) float64 {
				// High coherence → maintain stability
				// Successful patterns → stay stable
				coherence := id.Coherence
				maturity := math.Min(float64(id.Iterations)/1000.0, 1.0)
				return (coherence*0.6 + maturity*0.4)
			},
		},
		NegativeProcess: &Process{
			Name:       "Flexibility",
			Activation: 0.5,
			Weight:     1.0,
			Influence: func(id *Identity) float64 {
				// High uncertainty → need flexibility
				// Environmental change → adapt
				uncertainty := 1.0 - id.Coherence
				return uncertainty
			},
		},
		Balance:        0.0,
		OptimalBalance: 0.0,
	}

	// Speed vs Accuracy
	os.Processes[SpeedAccuracy] = &OpponentPair{
		Name: SpeedAccuracy,
		PositiveProcess: &Process{
			Name:       "Speed",
			Activation: 0.5,
			Weight:     1.0,
			Influence: func(id *Identity) float64 {
				// High emotional arousal → prioritize speed
				arousal := 0.5 // Default
				if id.EmotionalState != nil {
					arousal = id.EmotionalState.Arousal
				}
				return arousal
			},
		},
		NegativeProcess: &Process{
			Name:       "Accuracy",
			Activation: 0.5,
			Weight:     1.0,
			Influence: func(id *Identity) float64 {
				// Low arousal → prioritize accuracy
				// Important decision → be accurate
				calmness := 0.5
				if id.EmotionalState != nil {
					calmness = 1.0 - id.EmotionalState.Arousal
				}
				return calmness
			},
		},
		Balance:        0.0,
		OptimalBalance: 0.0,
	}

	// Approach vs Avoidance
	os.Processes[ApproachAvoidance] = &OpponentPair{
		Name: ApproachAvoidance,
		PositiveProcess: &Process{
			Name:       "Approach",
			Activation: 0.5,
			Weight:     1.0,
			Influence: func(id *Identity) float64 {
				// Positive valence → approach
				valence := 0.5
				if id.EmotionalState != nil {
					valence = (id.EmotionalState.Valence + 1.0) / 2.0 // Normalize to [0,1]
				}
				return valence
			},
		},
		NegativeProcess: &Process{
			Name:       "Avoidance",
			Activation: 0.5,
			Weight:     1.0,
			Influence: func(id *Identity) float64 {
				// Negative valence → avoid
				valence := 0.5
				if id.EmotionalState != nil {
					valence = (-id.EmotionalState.Valence + 1.0) / 2.0 // Invert and normalize
				}
				return valence
			},
		},
		Balance:        0.0,
		OptimalBalance: 0.0,
	}

	// Abstraction vs Concreteness
	os.Processes[AbstractionConcreteness] = &OpponentPair{
		Name: AbstractionConcreteness,
		PositiveProcess: &Process{
			Name:       "Abstraction",
			Activation: 0.5,
			Weight:     1.0,
			Influence: func(id *Identity) float64 {
				// High recursive depth → abstraction
				// Many patterns → abstract
				recursionInfluence := math.Min(float64(id.RecursiveDepth)/5.0, 1.0)
				return recursionInfluence
			},
		},
		NegativeProcess: &Process{
			Name:       "Concreteness",
			Activation: 0.5,
			Weight:     1.0,
			Influence: func(id *Identity) float64 {
				// Embodied context → concrete
				// Spatial awareness → concrete
				embodiedInfluence := 0.5
				if id.SpatialContext != nil && id.SpatialContext.Field != nil {
					embodiedInfluence = id.SpatialContext.Field.Intensity
				}
				return embodiedInfluence
			},
		},
		Balance:        0.0,
		OptimalBalance: 0.0,
	}
}

// OptimizeBalance dynamically optimizes opponent balance based on context
func (os *OpponentSystem) OptimizeBalance(id *Identity, context string) map[string]float64 {
	os.mu.Lock()
	defer os.mu.Unlock()

	balances := make(map[string]float64)

	for name, pair := range os.Processes {
		// Calculate influences
		posInfluence := pair.PositiveProcess.Influence(id)
		negInfluence := pair.NegativeProcess.Influence(id)

		// Update activations
		pair.PositiveProcess.Activation = posInfluence * pair.PositiveProcess.Weight
		pair.NegativeProcess.Activation = negInfluence * pair.NegativeProcess.Weight

		// Calculate new balance using gradient descent toward optimal
		newBalance := os.calculateOptimalBalance(pair, posInfluence, negInfluence)

		// Smooth transition (momentum) - prevents oscillation
		pair.Balance = 0.7*pair.Balance + 0.3*newBalance

		// Record balance point
		pair.History = append(pair.History, BalancePoint{
			Timestamp: time.Now(),
			Balance:   pair.Balance,
			Context:   context,
		})

		// Keep history manageable
		if len(pair.History) > 1000 {
			pair.History = pair.History[len(pair.History)-1000:]
		}

		balances[name] = pair.Balance
	}

	// Record system snapshot
	os.History = append(os.History, BalanceSnapshot{
		Timestamp: time.Now(),
		Balances:  balances,
		Context:   context,
	})

	// Keep system history manageable
	if len(os.History) > 1000 {
		os.History = os.History[len(os.History)-1000:]
	}

	return balances
}

func (os *OpponentSystem) calculateOptimalBalance(pair *OpponentPair, posInfluence, negInfluence float64) float64 {
	// Optimal balance based on relative influences
	// Range: -1 (fully negative) to +1 (fully positive)

	total := posInfluence + negInfluence
	if total == 0 {
		return 0.0
	}

	balance := (posInfluence - negInfluence) / total

	// Apply sigmoid (tanh) to smooth and bound
	return math.Tanh(balance)
}

// GetCurrentBalance returns the current balance for a specific opponent pair
func (os *OpponentSystem) GetCurrentBalance(pairName string) float64 {
	os.mu.RLock()
	defer os.mu.RUnlock()

	if pair, exists := os.Processes[pairName]; exists {
		return pair.Balance
	}
	return 0.0
}

// ApplyBalanceToDecision uses opponent balance to inform decision-making
func (os *OpponentSystem) ApplyBalanceToDecision(decision *Decision) {
	os.mu.RLock()
	defer os.mu.RUnlock()

	// Apply exploration/exploitation balance
	explorationBalance := os.GetCurrentBalance(ExplorationExploitation)
	// Map from [-1, 1] to [0, 1]
	decision.ExplorationWeight = (explorationBalance + 1.0) / 2.0

	// Apply breadth/depth balance
	breadthBalance := os.GetCurrentBalance(BreadthDepth)
	if breadthBalance > 0 {
		decision.ScopePreference = "breadth"
	} else {
		decision.ScopePreference = "depth"
	}

	// Apply stability/flexibility balance
	stabilityBalance := os.GetCurrentBalance(StabilityFlexibility)
	// Map from [-1, 1] to [0, 1], where -1 (flexibility) = high adaptation, +1 (stability) = low adaptation
	decision.AdaptationRate = (1.0 - stabilityBalance) / 2.0

	// Apply speed/accuracy balance to confidence threshold
	speedBalance := os.GetCurrentBalance(SpeedAccuracy)
	// Speed (+1) → lower confidence threshold, Accuracy (-1) → higher threshold
	decision.Confidence = 0.5 - (speedBalance * 0.3) // Range: [0.2, 0.8]
}

// UpdateOutcome records the outcome of a balanced decision for learning
func (os *OpponentSystem) UpdateOutcome(context string, outcome float64) {
	os.mu.Lock()
	defer os.mu.Unlock()

	// Find recent balance points for this context
	for _, pair := range os.Processes {
		for i := len(pair.History) - 1; i >= 0 && i >= len(pair.History)-10; i-- {
			if pair.History[i].Context == context {
				pair.History[i].Outcome = outcome
				// Can use this to learn optimal balances over time
				break
			}
		}
	}
}

// GetBalanceStats returns statistics about a specific opponent pair
func (os *OpponentSystem) GetBalanceStats(pairName string) map[string]float64 {
	os.mu.RLock()
	defer os.mu.RUnlock()

	pair, exists := os.Processes[pairName]
	if !exists {
		return nil
	}

	stats := make(map[string]float64)
	stats["current_balance"] = pair.Balance
	stats["optimal_balance"] = pair.OptimalBalance

	// Calculate average balance over recent history
	if len(pair.History) > 0 {
		sum := 0.0
		count := 0
		recent := 100
		if len(pair.History) < recent {
			recent = len(pair.History)
		}

		for i := len(pair.History) - recent; i < len(pair.History); i++ {
			sum += pair.History[i].Balance
			count++
		}
		stats["average_balance"] = sum / float64(count)

		// Calculate variance
		variance := 0.0
		for i := len(pair.History) - recent; i < len(pair.History); i++ {
			diff := pair.History[i].Balance - stats["average_balance"]
			variance += diff * diff
		}
		stats["variance"] = variance / float64(count)
		stats["stability"] = 1.0 / (1.0 + stats["variance"]) // Higher = more stable
	}

	return stats
}

// GetSystemWisdomScore calculates overall wisdom based on balance optimization
func (os *OpponentSystem) GetSystemWisdomScore() float64 {
	os.mu.RLock()
	defer os.mu.RUnlock()

	// Wisdom = ability to maintain appropriate balance across all dimensions
	// Higher score = better dynamic balance

	if len(os.Processes) == 0 {
		return 0.0
	}

	totalScore := 0.0
	count := 0

	for _, pair := range os.Processes {
		stats := os.getBalanceStatsInternal(pair.Name)
		if stats != nil {
			// Score based on stability and appropriateness
			stability := stats["stability"]
			totalScore += stability
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return totalScore / float64(count)
}

// getBalanceStatsInternal is an internal version that doesn't acquire locks
func (os *OpponentSystem) getBalanceStatsInternal(pairName string) map[string]float64 {
	pair, exists := os.Processes[pairName]
	if !exists {
		return nil
	}

	stats := make(map[string]float64)
	stats["current_balance"] = pair.Balance
	stats["optimal_balance"] = pair.OptimalBalance

	if len(pair.History) > 0 {
		sum := 0.0
		count := 0
		recent := 100
		if len(pair.History) < recent {
			recent = len(pair.History)
		}

		for i := len(pair.History) - recent; i < len(pair.History); i++ {
			sum += pair.History[i].Balance
			count++
		}
		stats["average_balance"] = sum / float64(count)

		variance := 0.0
		for i := len(pair.History) - recent; i < len(pair.History); i++ {
			diff := pair.History[i].Balance - stats["average_balance"]
			variance += diff * diff
		}
		stats["variance"] = variance / float64(count)
		stats["stability"] = 1.0 / (1.0 + stats["variance"])
	}

	return stats
}
