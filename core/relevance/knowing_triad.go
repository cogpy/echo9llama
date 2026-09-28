package relevance

import (
	"fmt"
	"math"
	"sync"
)

// KnowingTriad implements Triad I: The Ways of Knowing (Epistemological)
// Integrates four ways of knowing: Propositional, Procedural, Perspectival, Participatory
type KnowingTriad struct {
	mu sync.RWMutex

	// The four ways of knowing
	Propositional float64 // Knowing-that (facts, beliefs)
	Procedural    float64 // Knowing-how (skills, abilities)
	Perspectival  float64 // Knowing-as (framing, salience)
	Participatory float64 // Knowing-by-being (transformation)

	// Integration state
	GnosticIntegration float64 // Transformative knowing integration

	// Learning rates
	PropositionalLearningRate float64
	ProceduralLearningRate    float64
	PerspectivalLearningRate  float64
	ParticipatoryLearningRate float64
}

// NewKnowingTriad creates a new Ways of Knowing triad
func NewKnowingTriad() *KnowingTriad {
	return &KnowingTriad{
		Propositional: 0.5,
		Procedural:    0.5,
		Perspectival:  0.5,
		Participatory: 0.5,

		GnosticIntegration: 0.5,

		PropositionalLearningRate: 0.1,
		ProceduralLearningRate:    0.08,
		PerspectivalLearningRate:  0.12,
		ParticipatoryLearningRate: 0.05, // Slowest - transformation takes time
	}
}

// Balance balances the four ways of knowing
func (kt *KnowingTriad) Balance() {
	kt.mu.Lock()
	defer kt.mu.Unlock()

	// Calculate target balance (not equal, but optimal proportions)
	// Perspectival should be slightly higher (determines relevance)
	// Participatory develops slower

	total := kt.Propositional + kt.Procedural + kt.Perspectival + kt.Participatory
	if total == 0 {
		return
	}

	// Optimal proportions based on Vervaeke's framework
	optimalProportions := map[string]float64{
		"propositional": 0.25,
		"procedural":    0.25,
		"perspectival":  0.30, // Slightly higher - determines salience
		"participatory": 0.20, // Deeper but develops slower
	}

	// Gently nudge toward optimal proportions
	nudgeFactor := 0.05

	kt.Propositional += nudgeFactor * (optimalProportions["propositional"]*total - kt.Propositional)
	kt.Procedural += nudgeFactor * (optimalProportions["procedural"]*total - kt.Procedural)
	kt.Perspectival += nudgeFactor * (optimalProportions["perspectival"]*total - kt.Perspectival)
	kt.Participatory += nudgeFactor * (optimalProportions["participatory"]*total - kt.Participatory)

	// Update gnostic integration
	kt.updateGnosticIntegration()
}

// updateGnosticIntegration updates transformative knowing integration
func (kt *KnowingTriad) updateGnosticIntegration() {
	// Gnosis = integration of all four ways of knowing
	// Weighted toward participatory and perspectival

	kt.GnosticIntegration = (0.2*kt.Propositional +
		0.2*kt.Procedural +
		0.3*kt.Perspectival +
		0.3*kt.Participatory)
}

// Analyze analyzes input through all four ways of knowing
func (kt *KnowingTriad) Analyze(input interface{}) *KnowingAnalysis {
	kt.mu.RLock()
	defer kt.mu.RUnlock()

	analysis := &KnowingAnalysis{
		Input: input,
	}

	// Analyze propositionally (what facts/beliefs apply)
	analysis.PropositionalScore = kt.analyzePropositional(input)

	// Analyze procedurally (what skills/procedures apply)
	analysis.ProceduralScore = kt.analyzeProcedural(input)

	// Analyze perspectivally (how to frame/attend to it)
	analysis.PerspectivalScore = kt.analyzePerspectival(input)

	// Analyze participatorily (how it transforms us)
	analysis.ParticipatoryScore = kt.analyzeParticipatory(input)

	// Overall knowing score
	analysis.OverallScore = (analysis.PropositionalScore*0.25 +
		analysis.ProceduralScore*0.25 +
		analysis.PerspectivalScore*0.3 +
		analysis.ParticipatoryScore*0.2)

	return analysis
}

// analyzePropositional analyzes input propositionally
func (kt *KnowingTriad) analyzePropositional(input interface{}) float64 {
	// How well do we know facts/beliefs about this?
	// Scaled by current propositional knowledge level
	return kt.Propositional * 0.8 // Some uncertainty
}

// analyzeProcedural analyzes input procedurally
func (kt *KnowingTriad) analyzeProcedural(input interface{}) float64 {
	// How skilled are we at dealing with this?
	// Scaled by current procedural knowledge level
	return kt.Procedural * 0.85 // Slightly more certain
}

// analyzePerspectival analyzes input perspectivally
func (kt *KnowingTriad) analyzePerspectival(input interface{}) float64 {
	// How well can we frame and attend to this?
	// This is key for relevance realization
	return kt.Perspectival * 0.9 // High certainty in framing
}

// analyzeParticipatory analyzes input participatorily
func (kt *KnowingTriad) analyzeParticipatory(input interface{}) float64 {
	// How much does this transform who we are?
	// Deepest but hardest to articulate
	return kt.Participatory * 0.7 // Lower certainty but deeper
}

// UpdateFromExperience updates knowing from experience
func (kt *KnowingTriad) UpdateFromExperience(exp *Experience) {
	kt.mu.Lock()
	defer kt.mu.Unlock()

	// Update each way of knowing based on feedback
	feedback := math.Max(-1.0, math.Min(1.0, exp.Feedback))

	// Propositional: learn facts
	kt.Propositional += kt.PropositionalLearningRate * feedback
	kt.Propositional = math.Max(0, math.Min(1, kt.Propositional))

	// Procedural: improve skills through practice
	kt.Procedural += kt.ProceduralLearningRate * feedback
	kt.Procedural = math.Max(0, math.Min(1, kt.Procedural))

	// Perspectival: refine framing
	kt.Perspectival += kt.PerspectivalLearningRate * feedback
	kt.Perspectival = math.Max(0, math.Min(1, kt.Perspectival))

	// Participatory: transform through significant experiences
	if math.Abs(feedback) > 0.5 { // Only significant feedback transforms
		kt.Participatory += kt.ParticipatoryLearningRate * feedback
		kt.Participatory = math.Max(0, math.Min(1, kt.Participatory))
	}

	kt.updateGnosticIntegration()
}

// UpdateFromWisdom updates knowing from wisdom development
func (kt *KnowingTriad) UpdateFromWisdom(wt *WisdomTriad) {
	kt.mu.Lock()
	defer kt.mu.Unlock()

	// Wisdom transforms knowing
	// Morality enhances participatory knowing
	kt.Participatory = 0.95*kt.Participatory + 0.05*wt.Morality

	// Meaning enhances perspectival knowing
	kt.Perspectival = 0.95*kt.Perspectival + 0.05*wt.Meaning

	// Mastery enhances procedural knowing
	kt.Procedural = 0.95*kt.Procedural + 0.05*wt.Mastery

	kt.updateGnosticIntegration()
}

// GetState returns the current state
func (kt *KnowingTriad) GetState() map[string]float64 {
	kt.mu.RLock()
	defer kt.mu.RUnlock()

	return map[string]float64{
		"propositional":       kt.Propositional,
		"procedural":          kt.Procedural,
		"perspectival":        kt.Perspectival,
		"participatory":       kt.Participatory,
		"gnostic_integration": kt.GnosticIntegration,
	}
}

// KnowingAnalysis represents analysis through the ways of knowing
type KnowingAnalysis struct {
	Input              interface{}
	PropositionalScore float64
	ProceduralScore    float64
	PerspectivalScore  float64
	ParticipatoryScore float64
	OverallScore       float64
}

func (ka *KnowingAnalysis) String() string {
	return fmt.Sprintf("KnowingAnalysis(prop: %.2f, proc: %.2f, persp: %.2f, part: %.2f, overall: %.2f)",
		ka.PropositionalScore, ka.ProceduralScore, ka.PerspectivalScore,
		ka.ParticipatoryScore, ka.OverallScore)
}
