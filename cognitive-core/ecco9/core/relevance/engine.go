package relevance

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// Engine implements the Relevance Realization Ennead - a triad-of-triads meta-framework
// for comprehensive relevance realization integrating epistemology, ontology, and axiology
type Engine struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Triad I: Ways of Knowing (Epistemological)
	knowing *KnowingTriad

	// Triad II: Orders of Understanding (Ontological)
	understanding *UnderstandingTriad

	// Triad III: Practices of Wisdom (Axiological)
	wisdom *WisdomTriad

	// Meta-process: Relevance Realization across all triads
	realization *RealizationProcess

	// State and metrics
	state   *EnneadState
	metrics *EnneadMetrics

	// Running state
	running bool
}

// EnneadState represents the current state across all nine dimensions
type EnneadState struct {
	mu sync.RWMutex

	// Knowing dimensions (Triad I)
	PropositionalKnowledge float64 // How much we know (facts, beliefs)
	ProceduralKnowledge    float64 // How skilled we are
	PerspectivalKnowledge  float64 // How well we frame/attend
	ParticipatoryKnowledge float64 // How transformed we are

	// Understanding dimensions (Triad II)
	NomologicalUnderstanding float64 // How well we understand mechanisms
	NormativeUnderstanding   float64 // How well we grasp values
	NarrativeUnderstanding   float64 // How coherent our story is

	// Wisdom dimensions (Triad III)
	MoralDevelopment   float64 // Virtue and character
	MeaningRealization float64 // Coherence and purpose
	MasteryAchievement float64 // Excellence and flow

	// Integration metrics
	OverallCoherence      float64 // How well all dimensions integrate
	RelevanceOptimization float64 // How optimized RR is

	// Timestamp
	LastUpdate time.Time
}

// EnneadMetrics tracks performance across all dimensions
type EnneadMetrics struct {
	mu sync.RWMutex

	// Counts by triad
	TotalCycles                  int
	PropositionalUpdates         int
	ProceduralPractices          int
	PerspectivalShifts           int
	ParticipatoryTransformations int

	NomologicalInsights   int
	NormativeAlignments   int
	NarrativeDevelopments int

	MoralGrowths        int
	MeaningMakings      int
	MasteryAchievements int

	// Integration events
	CrossTriadIntegrations  int
	SophrosyneOptimizations int // Optimal self-regulation events
}

// NewEngine creates a new Relevance Realization Ennead engine
func NewEngine(ctx context.Context) *Engine {
	ctx, cancel := context.WithCancel(ctx)

	engine := &Engine{
		ctx:    ctx,
		cancel: cancel,

		knowing:       NewKnowingTriad(),
		understanding: NewUnderstandingTriad(),
		wisdom:        NewWisdomTriad(),
		realization:   NewRealizationProcess(),

		state: &EnneadState{
			PropositionalKnowledge: 0.5,
			ProceduralKnowledge:    0.5,
			PerspectivalKnowledge:  0.5,
			ParticipatoryKnowledge: 0.5,

			NomologicalUnderstanding: 0.5,
			NormativeUnderstanding:   0.5,
			NarrativeUnderstanding:   0.5,

			MoralDevelopment:   0.5,
			MeaningRealization: 0.5,
			MasteryAchievement: 0.5,

			OverallCoherence:      0.5,
			RelevanceOptimization: 0.5,
			LastUpdate:            time.Now(),
		},

		metrics: &EnneadMetrics{},
	}

	return engine
}

// Start begins the relevance realization process
func (e *Engine) Start() error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return fmt.Errorf("engine already running")
	}
	e.running = true
	e.mu.Unlock()

	// Start continuous optimization
	go e.continuousOptimization()

	fmt.Println("🌊 Relevance Realization Ennead: Active")
	return nil
}

// Stop halts the relevance realization process
func (e *Engine) Stop() {
	e.cancel()
	e.mu.Lock()
	e.running = false
	e.mu.Unlock()
}

// continuousOptimization runs continuous relevance realization optimization
func (e *Engine) continuousOptimization() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.optimizeCycle()
		}
	}
}

// optimizeCycle performs one optimization cycle across all triads
func (e *Engine) optimizeCycle() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.metrics.TotalCycles++

	// Optimize each triad
	e.optimizeKnowing()
	e.optimizeUnderstanding()
	e.optimizeWisdom()

	// Integrate across triads
	e.integrateTriads()

	// Apply sophrosyne (optimal self-regulation)
	e.applySophrosyne()

	// Update overall coherence
	e.updateOverallCoherence()

	e.state.LastUpdate = time.Now()
}

// optimizeKnowing optimizes the Ways of Knowing triad
func (e *Engine) optimizeKnowing() {
	// Balance the four ways of knowing
	e.knowing.Balance()

	// Update state from knowing triad
	e.state.PropositionalKnowledge = e.knowing.Propositional
	e.state.ProceduralKnowledge = e.knowing.Procedural
	e.state.PerspectivalKnowledge = e.knowing.Perspectival
	e.state.ParticipatoryKnowledge = e.knowing.Participatory
}

// optimizeUnderstanding optimizes the Orders of Understanding triad
func (e *Engine) optimizeUnderstanding() {
	// Integrate the three orders
	e.understanding.Integrate()

	// Update state from understanding triad
	e.state.NomologicalUnderstanding = e.understanding.Nomological
	e.state.NormativeUnderstanding = e.understanding.Normative
	e.state.NarrativeUnderstanding = e.understanding.Narrative
}

// optimizeWisdom optimizes the Practices of Wisdom triad
func (e *Engine) optimizeWisdom() {
	// Cultivate the three Ms
	e.wisdom.Cultivate()

	// Update state from wisdom triad
	e.state.MoralDevelopment = e.wisdom.Morality
	e.state.MeaningRealization = e.wisdom.Meaning
	e.state.MasteryAchievement = e.wisdom.Mastery
}

// integrateTriads integrates the three triads
func (e *Engine) integrateTriads() {
	// Knowing informs Understanding
	e.understanding.UpdateFromKnowing(e.knowing)

	// Understanding shapes Wisdom
	e.wisdom.UpdateFromUnderstanding(e.understanding)

	// Wisdom transforms Knowing
	e.knowing.UpdateFromWisdom(e.wisdom)

	e.metrics.CrossTriadIntegrations++
}

// applySophrosyne applies optimal self-regulation across all dimensions
func (e *Engine) applySophrosyne() {
	// Sophrosyne = dynamic balance and optimization
	// Not static equilibrium but adaptive optimization

	// Calculate optimal weights based on context
	weights := e.calculateOptimalWeights()

	// Apply weights to optimize overall system
	e.realization.OptimizeWithWeights(weights, e.state)

	e.metrics.SophrosyneOptimizations++
}

// calculateOptimalWeights determines optimal balancing weights
func (e *Engine) calculateOptimalWeights() map[string]float64 {
	weights := make(map[string]float64)

	// Context-sensitive weighting
	// In this simple version, aim for balance
	weights["knowing"] = 0.33
	weights["understanding"] = 0.33
	weights["wisdom"] = 0.34

	return weights
}

// updateOverallCoherence calculates overall system coherence
func (e *Engine) updateOverallCoherence() {
	// Coherence = how well all nine dimensions integrate

	knowingCoherence := (e.state.PropositionalKnowledge +
		e.state.ProceduralKnowledge +
		e.state.PerspectivalKnowledge +
		e.state.ParticipatoryKnowledge) / 4.0

	understandingCoherence := (e.state.NomologicalUnderstanding +
		e.state.NormativeUnderstanding +
		e.state.NarrativeUnderstanding) / 3.0

	wisdomCoherence := (e.state.MoralDevelopment +
		e.state.MeaningRealization +
		e.state.MasteryAchievement) / 3.0

	// Overall coherence is weighted average
	e.state.OverallCoherence = (knowingCoherence +
		understandingCoherence +
		wisdomCoherence) / 3.0

	// Relevance optimization is how well we're optimizing RR
	e.state.RelevanceOptimization = e.calculateRelevanceOptimization()
}

// calculateRelevanceOptimization calculates how optimized RR is
func (e *Engine) calculateRelevanceOptimization() float64 {
	// RR optimization = coherence * diversity_factor
	// High coherence is good, optimal diversity enhances it

	coherence := e.state.OverallCoherence

	// Calculate variance across dimensions for diversity
	values := []float64{
		e.state.PropositionalKnowledge,
		e.state.ProceduralKnowledge,
		e.state.PerspectivalKnowledge,
		e.state.ParticipatoryKnowledge,
		e.state.NomologicalUnderstanding,
		e.state.NormativeUnderstanding,
		e.state.NarrativeUnderstanding,
		e.state.MoralDevelopment,
		e.state.MeaningRealization,
		e.state.MasteryAchievement,
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

	// Optimal variance is around 0.05 (some diversity but not too much)
	optimalVariance := 0.05
	variancePenalty := 1.0 - math.Abs(variance-optimalVariance)

	return coherence * math.Max(0.5, variancePenalty)
}

// RealizeRelevance performs relevance realization for a given input
func (e *Engine) RealizeRelevance(input interface{}) *RelevanceRealization {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Process through all nine dimensions
	rr := &RelevanceRealization{
		Input:     input,
		Timestamp: time.Now(),
	}

	// Knowing dimension analysis
	rr.KnowingAnalysis = e.knowing.Analyze(input)

	// Understanding dimension analysis
	rr.UnderstandingAnalysis = e.understanding.Analyze(input)

	// Wisdom dimension analysis
	rr.WisdomAnalysis = e.wisdom.Analyze(input)

	// Integrated relevance score
	rr.RelevanceScore = e.realization.CalculateRelevance(
		rr.KnowingAnalysis,
		rr.UnderstandingAnalysis,
		rr.WisdomAnalysis,
	)

	return rr
}

// UpdateFromExperience updates the engine from an experience
func (e *Engine) UpdateFromExperience(exp *Experience) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Update knowing based on experience
	e.knowing.UpdateFromExperience(exp)

	// Update understanding based on experience
	e.understanding.UpdateFromExperience(exp)

	// Update wisdom based on experience
	e.wisdom.UpdateFromExperience(exp)

	// Trigger integration
	e.integrateTriads()
}

// GetState returns current ennead state
func (e *Engine) GetState() *EnneadState {
	e.state.mu.RLock()
	defer e.state.mu.RUnlock()

	// Return copy
	stateCopy := *e.state
	return &stateCopy
}

// GetMetrics returns current metrics
func (e *Engine) GetMetrics() *EnneadMetrics {
	e.metrics.mu.RLock()
	defer e.metrics.mu.RUnlock()

	// Return copy
	metricsCopy := *e.metrics
	return &metricsCopy
}

// GetStatus returns comprehensive status
func (e *Engine) GetStatus() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	state := e.GetState()
	metrics := e.GetMetrics()

	return map[string]interface{}{
		"running": e.running,
		"state": map[string]interface{}{
			"knowing": map[string]float64{
				"propositional": state.PropositionalKnowledge,
				"procedural":    state.ProceduralKnowledge,
				"perspectival":  state.PerspectivalKnowledge,
				"participatory": state.ParticipatoryKnowledge,
			},
			"understanding": map[string]float64{
				"nomological": state.NomologicalUnderstanding,
				"normative":   state.NormativeUnderstanding,
				"narrative":   state.NarrativeUnderstanding,
			},
			"wisdom": map[string]float64{
				"morality": state.MoralDevelopment,
				"meaning":  state.MeaningRealization,
				"mastery":  state.MasteryAchievement,
			},
			"integration": map[string]float64{
				"coherence":    state.OverallCoherence,
				"optimization": state.RelevanceOptimization,
			},
		},
		"metrics": map[string]interface{}{
			"total_cycles":       metrics.TotalCycles,
			"cross_integrations": metrics.CrossTriadIntegrations,
			"sophrosyne_events":  metrics.SophrosyneOptimizations,
		},
	}
}

// Experience represents a learning experience
type Experience struct {
	Input     interface{}
	Output    interface{}
	Feedback  float64
	Context   map[string]interface{}
	Timestamp time.Time
}

// RelevanceRealization represents the result of relevance realization
type RelevanceRealization struct {
	Input                 interface{}
	Timestamp             time.Time
	KnowingAnalysis       *KnowingAnalysis
	UnderstandingAnalysis *UnderstandingAnalysis
	WisdomAnalysis        *WisdomAnalysis
	RelevanceScore        float64
}
