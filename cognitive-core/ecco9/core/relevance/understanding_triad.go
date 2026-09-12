package relevance

import (
	"fmt"
	"math"
	"sync"
)

// UnderstandingTriad implements Triad II: The Orders of Understanding (Ontological)
// Integrates three orders: Nomological, Normative, Narrative
type UnderstandingTriad struct {
	mu sync.RWMutex

	// The three orders of understanding
	Nomological float64 // How things work (causal-scientific)
	Normative   float64 // What matters (evaluative-ethical)
	Narrative   float64 // How things develop (temporal-historical)

	// Integration state
	MeaningIntegration float64 // How well the three orders create meaning

	// Learning rates
	NomologicalLearningRate float64
	NormativeLearningRate   float64
	NarrativeLearningRate   float64
}

// NewUnderstandingTriad creates a new Orders of Understanding triad
func NewUnderstandingTriad() *UnderstandingTriad {
	return &UnderstandingTriad{
		Nomological: 0.5,
		Normative:   0.5,
		Narrative:   0.5,

		MeaningIntegration: 0.5,

		NomologicalLearningRate: 0.1,
		NormativeLearningRate:   0.08,
		NarrativeLearningRate:   0.12,
	}
}

// Integrate integrates the three orders of understanding
func (ut *UnderstandingTriad) Integrate() {
	ut.mu.Lock()
	defer ut.mu.Unlock()

	// The three orders must work together to create meaning
	// Balance them toward optimal integration

	total := ut.Nomological + ut.Normative + ut.Narrative
	if total == 0 {
		return
	}

	// Optimal proportions
	optimalProportions := map[string]float64{
		"nomological": 0.30, // Scientific understanding
		"normative":   0.35, // Value understanding (slightly higher - what matters)
		"narrative":   0.35, // Historical understanding (coherence over time)
	}

	// Gently nudge toward optimal proportions
	nudgeFactor := 0.05

	ut.Nomological += nudgeFactor * (optimalProportions["nomological"]*total - ut.Nomological)
	ut.Normative += nudgeFactor * (optimalProportions["normative"]*total - ut.Normative)
	ut.Narrative += nudgeFactor * (optimalProportions["narrative"]*total - ut.Narrative)

	// Update meaning integration
	ut.updateMeaningIntegration()
}

// updateMeaningIntegration updates how well the orders create meaning
func (ut *UnderstandingTriad) updateMeaningIntegration() {
	// Meaning requires all three orders
	// It's not additive but multiplicative - need all three

	// Geometric mean (all three must be present)
	ut.MeaningIntegration = math.Pow(
		ut.Nomological*ut.Normative*ut.Narrative,
		1.0/3.0,
	)
}

// Analyze analyzes input through all three orders
func (ut *UnderstandingTriad) Analyze(input interface{}) *UnderstandingAnalysis {
	ut.mu.RLock()
	defer ut.mu.RUnlock()

	analysis := &UnderstandingAnalysis{
		Input: input,
	}

	// Analyze nomologically (how does it work)
	analysis.NomologicalScore = ut.analyzeNomological(input)

	// Analyze normatively (why does it matter)
	analysis.NormativeScore = ut.analyzeNormative(input)

	// Analyze narratively (how did it develop)
	analysis.NarrativeScore = ut.analyzeNarrative(input)

	// Overall understanding score (weighted)
	analysis.OverallScore = (analysis.NomologicalScore*0.3 +
		analysis.NormativeScore*0.35 +
		analysis.NarrativeScore*0.35)

	return analysis
}

// analyzeNomological analyzes how things work
func (ut *UnderstandingTriad) analyzeNomological(input interface{}) float64 {
	// How well do we understand the causal mechanisms?
	return ut.Nomological * 0.85
}

// analyzeNormative analyzes what matters
func (ut *UnderstandingTriad) analyzeNormative(input interface{}) float64 {
	// How well do we understand the value and significance?
	return ut.Normative * 0.9 // Higher weight - crucial for relevance
}

// analyzeNarrative analyzes how things develop
func (ut *UnderstandingTriad) analyzeNarrative(input interface{}) float64 {
	// How well do we understand the developmental trajectory?
	return ut.Narrative * 0.8
}

// UpdateFromExperience updates understanding from experience
func (ut *UnderstandingTriad) UpdateFromExperience(exp *Experience) {
	ut.mu.Lock()
	defer ut.mu.Unlock()

	feedback := math.Max(-1.0, math.Min(1.0, exp.Feedback))

	// Update each order based on feedback
	ut.Nomological += ut.NomologicalLearningRate * feedback
	ut.Nomological = math.Max(0, math.Min(1, ut.Nomological))

	ut.Normative += ut.NormativeLearningRate * feedback
	ut.Normative = math.Max(0, math.Min(1, ut.Normative))

	ut.Narrative += ut.NarrativeLearningRate * feedback
	ut.Narrative = math.Max(0, math.Min(1, ut.Narrative))

	ut.updateMeaningIntegration()
}

// UpdateFromKnowing updates understanding from knowing development
func (ut *UnderstandingTriad) UpdateFromKnowing(kt *KnowingTriad) {
	ut.mu.Lock()
	defer ut.mu.Unlock()

	// Knowing informs Understanding
	// Propositional knowing enhances nomological understanding
	ut.Nomological = 0.95*ut.Nomological + 0.05*kt.Propositional

	// Perspectival knowing enhances normative understanding
	ut.Normative = 0.95*ut.Normative + 0.05*kt.Perspectival

	// Participatory knowing enhances narrative understanding
	ut.Narrative = 0.95*ut.Narrative + 0.05*kt.Participatory

	ut.updateMeaningIntegration()
}

// GetState returns the current state
func (ut *UnderstandingTriad) GetState() map[string]float64 {
	ut.mu.RLock()
	defer ut.mu.RUnlock()

	return map[string]float64{
		"nomological":         ut.Nomological,
		"normative":           ut.Normative,
		"narrative":           ut.Narrative,
		"meaning_integration": ut.MeaningIntegration,
	}
}

// UnderstandingAnalysis represents analysis through the orders of understanding
type UnderstandingAnalysis struct {
	Input            interface{}
	NomologicalScore float64
	NormativeScore   float64
	NarrativeScore   float64
	OverallScore     float64
}

func (ua *UnderstandingAnalysis) String() string {
	return fmt.Sprintf("UnderstandingAnalysis(nomo: %.2f, norm: %.2f, narr: %.2f, overall: %.2f)",
		ua.NomologicalScore, ua.NormativeScore, ua.NarrativeScore, ua.OverallScore)
}
