
package meta

import (
	"sync"
	"time"
)

// MetaLearner implements learning about learning strategies
type MetaLearner struct {
	mu                sync.RWMutex
	strategies        map[string]LearningStrategy
	performance       map[string]PerformanceMetrics
	strategyEvaluator StrategyEvaluator
	adaptationCycles  []AdaptationCycle
	currentStrategy   string
	explorationRate   float64
}

// LearningStrategy defines a learning approach
type LearningStrategy struct {
	ID          string
	Name        string
	Parameters  map[string]interface{}
	Approach    string // "gradient", "evolutionary", "reinforcement", etc.
	Adaptivity  float64
	Complexity  float64
	Performance func(context LearningContext) float64
}

// PerformanceMetrics tracks strategy effectiveness
type PerformanceMetrics struct {
	Accuracy        float64
	LearningRate    float64
	Convergence     time.Duration
	Generalization  float64
	Efficiency      float64
	Robustness      float64
	LastUpdated     time.Time
}

// StrategyEvaluator assesses learning strategy performance
type StrategyEvaluator interface {
	EvaluateStrategy(strategy LearningStrategy, context LearningContext) PerformanceMetrics
	CompareStrategies(strategies []LearningStrategy, context LearningContext) []StrategyRanking
	SuggestImprovements(strategy LearningStrategy, metrics PerformanceMetrics) []Improvement
}

// LearningContext provides environmental information for strategy evaluation
type LearningContext struct {
	TaskType        string
	DataCharacteristics map[string]interface{}
	PerformanceTargets  map[string]float64
	Constraints     map[string]interface{}
	TimeHorizon     time.Duration
}

// AdaptationCycle records meta-learning adaptation events
type AdaptationCycle struct {
	Timestamp       time.Time
	TriggerContext  LearningContext
	OldStrategy     string
	NewStrategy     string
	ExpectedGain    float64
	ActualGain      float64
	ConfidenceLevel float64
}

// StrategyRanking ranks strategies by performance
type StrategyRanking struct {
	StrategyID string
	Score      float64
	Confidence float64
	Rationale  string
}

// NewMetaLearner creates a new meta-learning system
func NewMetaLearner(evaluator StrategyEvaluator) *MetaLearner {
	return &MetaLearner{
		strategies:        make(map[string]LearningStrategy),
		performance:       make(map[string]PerformanceMetrics),
		strategyEvaluator: evaluator,
		adaptationCycles:  make([]AdaptationCycle, 0),
		explorationRate:   0.1,
	}
}

// AdaptLearningStrategy selects optimal learning strategy for context
func (ml *MetaLearner) AdaptLearningStrategy(context LearningContext) error {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	
	// Evaluate all available strategies
	var strategies []LearningStrategy
	for _, strategy := range ml.strategies {
		strategies = append(strategies, strategy)
	}
	
	rankings := ml.strategyEvaluator.CompareStrategies(strategies, context)
	
	if len(rankings) == 0 {
		return nil
	}
	
	// Select best strategy with exploration
	var selectedStrategy string
	if ml.shouldExplore() {
		// Exploration: select randomly from top strategies
		selectedStrategy = ml.exploreStrategies(rankings)
	} else {
		// Exploitation: select best strategy
		selectedStrategy = rankings[0].StrategyID
	}
	
	// Record adaptation cycle
	cycle := AdaptationCycle{
		Timestamp:       time.Now(),
		TriggerContext:  context,
		OldStrategy:     ml.currentStrategy,
		NewStrategy:     selectedStrategy,
		ExpectedGain:    rankings[0].Score,
		ConfidenceLevel: rankings[0].Confidence,
	}
	
	ml.currentStrategy = selectedStrategy
	ml.adaptationCycles = append(ml.adaptationCycles, cycle)
	
	return nil
}

// shouldExplore determines if exploration should occur
func (ml *MetaLearner) shouldExplore() bool {
	return len(ml.adaptationCycles)%10 < int(ml.explorationRate*10)
}

// exploreStrategies selects strategy for exploration
func (ml *MetaLearner) exploreStrategies(rankings []StrategyRanking) string {
	if len(rankings) > 1 {
		return rankings[1].StrategyID // Select second-best for exploration
	}
	return rankings[0].StrategyID
}

// UpdatePerformance records actual performance of strategy
func (ml *MetaLearner) UpdatePerformance(strategyID string, metrics PerformanceMetrics) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	
	ml.performance[strategyID] = metrics
	
	// Update last adaptation cycle with actual performance
	if len(ml.adaptationCycles) > 0 {
		lastCycle := &ml.adaptationCycles[len(ml.adaptationCycles)-1]
		if lastCycle.NewStrategy == strategyID {
			lastCycle.ActualGain = metrics.Accuracy
		}
	}
}

// GetCurrentStrategy returns the active learning strategy
func (ml *MetaLearner) GetCurrentStrategy() (LearningStrategy, bool) {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	
	strategy, exists := ml.strategies[ml.currentStrategy]
	return strategy, exists
}

// GetAdaptationHistory returns meta-learning adaptation history
func (ml *MetaLearner) GetAdaptationHistory() []AdaptationCycle {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	return ml.adaptationCycles
}

// RegisterStrategy adds a new learning strategy to the meta-learner
func (ml *MetaLearner) RegisterStrategy(strategy LearningStrategy) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.strategies[strategy.ID] = strategy
}

// Improvement represents a suggested strategy enhancement
type Improvement struct {
	Parameter    string
	CurrentValue interface{}
	SuggestedValue interface{}
	ExpectedGain float64
	Confidence   float64
}
