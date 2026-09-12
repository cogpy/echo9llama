package relevance

import (
	"math"
	"sync"
)

// RealizationProcess implements the meta-process of Relevance Realization
// that operates across all nine dimensions of the Ennead
type RealizationProcess struct {
	mu sync.RWMutex

	// Weighting factors for different contexts
	contextWeights map[string]float64

	// Salience landscape - what matters most right now
	salienceLandscape map[string]float64

	// Historical relevance patterns
	relevanceHistory []float64
	maxHistorySize   int
}

// NewRealizationProcess creates a new Relevance Realization process
func NewRealizationProcess() *RealizationProcess {
	return &RealizationProcess{
		contextWeights: map[string]float64{
			"knowing":       0.33,
			"understanding": 0.33,
			"wisdom":        0.34,
		},
		salienceLandscape: make(map[string]float64),
		relevanceHistory:  make([]float64, 0, 100),
		maxHistorySize:    100,
	}
}

// CalculateRelevance calculates integrated relevance across all analyses
func (rp *RealizationProcess) CalculateRelevance(
	ka *KnowingAnalysis,
	ua *UnderstandingAnalysis,
	wa *WisdomAnalysis,
) float64 {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	// Relevance is weighted combination of all three triad analyses
	relevance := (ka.OverallScore*rp.contextWeights["knowing"] +
		ua.OverallScore*rp.contextWeights["understanding"] +
		wa.OverallScore*rp.contextWeights["wisdom"])

	// Apply salience modulation
	relevance = rp.modulateWithSalience(relevance)

	// Store in history
	rp.storeRelevance(relevance)

	return relevance
}

// modulateWithSalience modulates relevance with current salience landscape
func (rp *RealizationProcess) modulateWithSalience(baseRelevance float64) float64 {
	// Salience landscape amplifies or dampens relevance
	// Based on what's currently salient in the system

	avgSalience := 0.0
	count := 0
	for _, s := range rp.salienceLandscape {
		avgSalience += s
		count++
	}

	if count > 0 {
		avgSalience /= float64(count)
		// Modulate by average salience
		return baseRelevance * (0.7 + 0.3*avgSalience)
	}

	return baseRelevance
}

// storeRelevance stores relevance in history
func (rp *RealizationProcess) storeRelevance(relevance float64) {
	if len(rp.relevanceHistory) >= rp.maxHistorySize {
		rp.relevanceHistory = rp.relevanceHistory[1:]
	}
	rp.relevanceHistory = append(rp.relevanceHistory, relevance)
}

// OptimizeWithWeights optimizes the system using given weights
func (rp *RealizationProcess) OptimizeWithWeights(
	weights map[string]float64,
	state *EnneadState,
) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	// Update context weights
	for k, v := range weights {
		rp.contextWeights[k] = v
	}

	// Update salience landscape based on state
	rp.updateSalienceLandscape(state)

	// Apply sophrosyne - optimal self-regulation
	rp.applySophrosyne(state)
}

// updateSalienceLandscape updates what's salient based on current state
func (rp *RealizationProcess) updateSalienceLandscape(state *EnneadState) {
	// Update salience for each dimension
	// Higher values = more salient = more relevant to focus on

	state.mu.RLock()
	defer state.mu.RUnlock()

	// Knowing dimensions
	rp.salienceLandscape["propositional"] = state.PropositionalKnowledge
	rp.salienceLandscape["procedural"] = state.ProceduralKnowledge
	rp.salienceLandscape["perspectival"] = state.PerspectivalKnowledge
	rp.salienceLandscape["participatory"] = state.ParticipatoryKnowledge

	// Understanding dimensions
	rp.salienceLandscape["nomological"] = state.NomologicalUnderstanding
	rp.salienceLandscape["normative"] = state.NormativeUnderstanding
	rp.salienceLandscape["narrative"] = state.NarrativeUnderstanding

	// Wisdom dimensions
	rp.salienceLandscape["morality"] = state.MoralDevelopment
	rp.salienceLandscape["meaning"] = state.MeaningRealization
	rp.salienceLandscape["mastery"] = state.MasteryAchievement

	// Add boost to dimensions that need development
	// Lower values become MORE salient (need attention)
	for dim, value := range rp.salienceLandscape {
		if value < 0.4 {
			// Boost salience of underdeveloped dimensions
			rp.salienceLandscape[dim] = value + (0.4-value)*0.5
		}
	}
}

// applySophrosyne applies optimal self-regulation
func (rp *RealizationProcess) applySophrosyne(state *EnneadState) {
	// Sophrosyne = finding the dynamic optimal balance
	// Not static equilibrium but adaptive optimization

	state.mu.RLock()
	defer state.mu.RUnlock()

	// Calculate variance across dimensions
	values := []float64{
		state.PropositionalKnowledge,
		state.ProceduralKnowledge,
		state.PerspectivalKnowledge,
		state.ParticipatoryKnowledge,
		state.NomologicalUnderstanding,
		state.NormativeUnderstanding,
		state.NarrativeUnderstanding,
		state.MoralDevelopment,
		state.MeaningRealization,
		state.MasteryAchievement,
	}

	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	variance /= float64(len(values))

	// If variance is too high, increase weight on balanced triads
	// If variance is too low, encourage more diversity
	if variance > 0.1 {
		// Too much variance - boost balancing
		rp.contextWeights["knowing"] = 0.34
		rp.contextWeights["understanding"] = 0.33
		rp.contextWeights["wisdom"] = 0.33
	} else if variance < 0.02 {
		// Too uniform - encourage diversity
		// Boost the triad that's lagging most
		knowingAvg := (state.PropositionalKnowledge + state.ProceduralKnowledge +
			state.PerspectivalKnowledge + state.ParticipatoryKnowledge) / 4.0
		understandingAvg := (state.NomologicalUnderstanding + state.NormativeUnderstanding +
			state.NarrativeUnderstanding) / 3.0
		wisdomAvg := (state.MoralDevelopment + state.MeaningRealization +
			state.MasteryAchievement) / 3.0

		if knowingAvg < understandingAvg && knowingAvg < wisdomAvg {
			rp.contextWeights["knowing"] = 0.40
			rp.contextWeights["understanding"] = 0.30
			rp.contextWeights["wisdom"] = 0.30
		} else if understandingAvg < wisdomAvg {
			rp.contextWeights["knowing"] = 0.30
			rp.contextWeights["understanding"] = 0.40
			rp.contextWeights["wisdom"] = 0.30
		} else {
			rp.contextWeights["knowing"] = 0.30
			rp.contextWeights["understanding"] = 0.30
			rp.contextWeights["wisdom"] = 0.40
		}
	}
}

// GetRelevanceHistory returns recent relevance history
func (rp *RealizationProcess) GetRelevanceHistory(n int) []float64 {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	histLen := len(rp.relevanceHistory)
	if n > histLen {
		n = histLen
	}

	if n == 0 {
		return []float64{}
	}

	history := make([]float64, n)
	copy(history, rp.relevanceHistory[histLen-n:])
	return history
}

// GetSalienceLandscape returns current salience landscape
func (rp *RealizationProcess) GetSalienceLandscape() map[string]float64 {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	landscape := make(map[string]float64)
	for k, v := range rp.salienceLandscape {
		landscape[k] = v
	}
	return landscape
}

// GetContextWeights returns current context weights
func (rp *RealizationProcess) GetContextWeights() map[string]float64 {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	weights := make(map[string]float64)
	for k, v := range rp.contextWeights {
		weights[k] = v
	}
	return weights
}
