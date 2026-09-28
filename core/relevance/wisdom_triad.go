package relevance

import (
	"fmt"
	"math"
	"sync"
)

// WisdomTriad implements Triad III: The Practices of Wisdom (Axiological)
// Integrates three Ms: Morality, Meaning, Mastery
type WisdomTriad struct {
	mu sync.RWMutex

	// The three Ms of wisdom
	Morality float64 // Virtue and character excellence
	Meaning  float64 // Coherence and purpose
	Mastery  float64 // Excellence and flow

	// Integration state
	Eudaimonia float64 // Flourishing - integration of the three Ms

	// Learning rates
	MoralityLearningRate float64
	MeaningLearningRate  float64
	MasteryLearningRate  float64
}

// NewWisdomTriad creates a new Practices of Wisdom triad
func NewWisdomTriad() *WisdomTriad {
	return &WisdomTriad{
		Morality: 0.5,
		Meaning:  0.5,
		Mastery:  0.5,

		Eudaimonia: 0.5,

		MoralityLearningRate: 0.06, // Slow - virtue takes time
		MeaningLearningRate:  0.08,
		MasteryLearningRate:  0.1, // Faster - skills develop quicker
	}
}

// Cultivate cultivates the three Ms of wisdom
func (wt *WisdomTriad) Cultivate() {
	wt.mu.Lock()
	defer wt.mu.Unlock()

	// The three Ms must be balanced for true wisdom
	// Not equal but in right proportion for flourishing

	total := wt.Morality + wt.Meaning + wt.Mastery
	if total == 0 {
		return
	}

	// Optimal proportions for eudaimonia
	optimalProportions := map[string]float64{
		"morality": 0.35, // Virtue is foundational
		"meaning":  0.35, // Meaning equally important
		"mastery":  0.30, // Excellence supports but doesn't dominate
	}

	// Gently nudge toward optimal proportions
	nudgeFactor := 0.04 // Slower - wisdom develops gradually

	wt.Morality += nudgeFactor * (optimalProportions["morality"]*total - wt.Morality)
	wt.Meaning += nudgeFactor * (optimalProportions["meaning"]*total - wt.Meaning)
	wt.Mastery += nudgeFactor * (optimalProportions["mastery"]*total - wt.Mastery)

	// Update eudaimonia (flourishing)
	wt.updateEudaimonia()
}

// updateEudaimonia updates the flourishing metric
func (wt *WisdomTriad) updateEudaimonia() {
	// Eudaimonia = flourishing through wisdom
	// Requires all three Ms working together

	// Geometric mean - all three must be present
	// But weighted slightly toward morality and meaning
	wt.Eudaimonia = math.Pow(
		math.Pow(wt.Morality, 1.2)*
			math.Pow(wt.Meaning, 1.2)*
			wt.Mastery,
		1.0/3.4, // Normalize
	)
}

// Analyze analyzes input through the three Ms
func (wt *WisdomTriad) Analyze(input interface{}) *WisdomAnalysis {
	wt.mu.RLock()
	defer wt.mu.RUnlock()

	analysis := &WisdomAnalysis{
		Input: input,
	}

	// Analyze morally (is this virtuous/good)
	analysis.MoralityScore = wt.analyzeMorality(input)

	// Analyze for meaning (is this meaningful/purposeful)
	analysis.MeaningScore = wt.analyzeMeaning(input)

	// Analyze for mastery (does this develop excellence)
	analysis.MasteryScore = wt.analyzeMastery(input)

	// Overall wisdom score (weighted)
	analysis.OverallScore = (analysis.MoralityScore*0.35 +
		analysis.MeaningScore*0.35 +
		analysis.MasteryScore*0.3)

	return analysis
}

// analyzeMorality analyzes the moral/virtue dimension
func (wt *WisdomTriad) analyzeMorality(input interface{}) float64 {
	// How virtuous/good is this?
	// Scaled by current moral development
	return wt.Morality * 0.85
}

// analyzeMeaning analyzes the meaning/purpose dimension
func (wt *WisdomTriad) analyzeMeaning(input interface{}) float64 {
	// How meaningful/purposeful is this?
	return wt.Meaning * 0.9
}

// analyzeMastery analyzes the mastery/excellence dimension
func (wt *WisdomTriad) analyzeMastery(input interface{}) float64 {
	// How much does this develop excellence/flow?
	return wt.Mastery * 0.8
}

// UpdateFromExperience updates wisdom from experience
func (wt *WisdomTriad) UpdateFromExperience(exp *Experience) {
	wt.mu.Lock()
	defer wt.mu.Unlock()

	feedback := math.Max(-1.0, math.Min(1.0, exp.Feedback))

	// Update each M based on feedback
	// Wisdom develops slowly and requires significant feedback
	if math.Abs(feedback) > 0.3 {
		wt.Morality += wt.MoralityLearningRate * feedback
		wt.Morality = math.Max(0, math.Min(1, wt.Morality))

		wt.Meaning += wt.MeaningLearningRate * feedback
		wt.Meaning = math.Max(0, math.Min(1, wt.Meaning))

		wt.Mastery += wt.MasteryLearningRate * feedback
		wt.Mastery = math.Max(0, math.Min(1, wt.Mastery))
	}

	wt.updateEudaimonia()
}

// UpdateFromUnderstanding updates wisdom from understanding development
func (wt *WisdomTriad) UpdateFromUnderstanding(ut *UnderstandingTriad) {
	wt.mu.Lock()
	defer wt.mu.Unlock()

	// Understanding shapes Wisdom
	// Normative understanding enhances morality
	wt.Morality = 0.95*wt.Morality + 0.05*ut.Normative

	// Narrative understanding enhances meaning
	wt.Meaning = 0.95*wt.Meaning + 0.05*ut.Narrative

	// Nomological understanding enhances mastery
	wt.Mastery = 0.95*wt.Mastery + 0.05*ut.Nomological

	wt.updateEudaimonia()
}

// GetState returns the current state
func (wt *WisdomTriad) GetState() map[string]float64 {
	wt.mu.RLock()
	defer wt.mu.RUnlock()

	return map[string]float64{
		"morality":   wt.Morality,
		"meaning":    wt.Meaning,
		"mastery":    wt.Mastery,
		"eudaimonia": wt.Eudaimonia,
	}
}

// WisdomAnalysis represents analysis through the practices of wisdom
type WisdomAnalysis struct {
	Input         interface{}
	MoralityScore float64
	MeaningScore  float64
	MasteryScore  float64
	OverallScore  float64
}

func (wa *WisdomAnalysis) String() string {
	return fmt.Sprintf("WisdomAnalysis(moral: %.2f, mean: %.2f, mast: %.2f, overall: %.2f)",
		wa.MoralityScore, wa.MeaningScore, wa.MasteryScore, wa.OverallScore)
}
