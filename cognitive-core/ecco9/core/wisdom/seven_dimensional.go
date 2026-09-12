package wisdom

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// SevenDimensionalWisdom implements the complete seven-dimensional wisdom cultivation
// framework as specified in the Echo9 architecture
type SevenDimensionalWisdom struct {
	mu sync.RWMutex

	// The Seven Dimensions
	dimensions map[WisdomDimension]*DimensionState

	// Three Triads
	epistemicTriad   *EpistemicTriad   // Ways of Knowing
	cognitiveTriad   *CognitiveTriad   // Understanding Process
	axiologicalTriad *AxiologicalTriad // Practices of Wisdom

	// Overall cultivation state
	overallWisdom  float64
	coherenceScore float64
	evolutionRate  float64

	// Historical tracking
	snapshots      []SevenDimWisdomSnapshot
	cultivationLog []CultivationEvent

	// Timing
	startTime  time.Time
	lastUpdate time.Time
}

// WisdomDimension represents one of the seven dimensions
type WisdomDimension int

const (
	DimKnowledgeDepth WisdomDimension = iota
	DimKnowledgeBreadth
	DimIntegrationLevel
	DimPracticalApplication
	DimReflectiveInsight
	DimEthicalConsideration
	DimTemporalPerspective
)

func (d WisdomDimension) String() string {
	return [...]string{
		"Knowledge Depth",
		"Knowledge Breadth",
		"Integration Level",
		"Practical Application",
		"Reflective Insight",
		"Ethical Consideration",
		"Temporal Perspective",
	}[d]
}

// DimensionState tracks the state of a single wisdom dimension
type DimensionState struct {
	Value          float64 // Current value (0.0-1.0)
	Trend          float64 // Rate of change
	LastUpdate     time.Time
	UpdateCount    int64
	History        []float64 // Recent history for trend analysis
	TargetValue    float64   // Cultivation target
	CultivationLog []string  // Log of cultivation events
}

// SevenDimWisdomSnapshot captures the complete wisdom state
type SevenDimWisdomSnapshot struct {
	Timestamp          time.Time
	DimensionValues    map[WisdomDimension]float64
	OverallWisdom      float64
	CoherenceScore     float64
	EpistemicBalance   float64
	CognitiveBalance   float64
	AxiologicalBalance float64
	SignificantEvents  []string
}

// CultivationEvent represents an event in wisdom cultivation
type CultivationEvent struct {
	Timestamp   time.Time
	Type        string
	Dimension   WisdomDimension
	Impact      float64
	Description string
}

// EpistemicTriad: Ways of Knowing
type EpistemicTriad struct {
	Propositional float64 // Facts and theories
	Procedural    float64 // Skills and practices
	Perspectival  float64 // Frameworks and worldviews
	Participatory float64 // Identity and transformation
}

// CognitiveTriad: Understanding Process
type CognitiveTriad struct {
	Explanation    float64 // Causal understanding
	Realizing      float64 // Relevance realization
	Interpretation float64 // Meaning-making
}

// AxiologicalTriad: Practices of Wisdom
type AxiologicalTriad struct {
	Morality   float64 // Virtue and character
	Meaning    float64 // Coherence and purpose
	Mastery    float64 // Excellence and flow
	Eudaimonia float64 // Flourishing through integration
}

// NewSevenDimensionalWisdom creates a new seven-dimensional wisdom tracker
func NewSevenDimensionalWisdom() *SevenDimensionalWisdom {
	sdw := &SevenDimensionalWisdom{
		dimensions:       make(map[WisdomDimension]*DimensionState),
		snapshots:        make([]SevenDimWisdomSnapshot, 0, 1000),
		cultivationLog:   make([]CultivationEvent, 0, 10000),
		startTime:        time.Now(),
		lastUpdate:       time.Now(),
		epistemicTriad:   &EpistemicTriad{},
		cognitiveTriad:   &CognitiveTriad{},
		axiologicalTriad: &AxiologicalTriad{},
	}

	// Initialize all seven dimensions
	for dim := DimKnowledgeDepth; dim <= DimTemporalPerspective; dim++ {
		sdw.dimensions[dim] = &DimensionState{
			Value:          0.3, // Start with modest baseline
			History:        make([]float64, 0, 100),
			TargetValue:    0.8, // Aim high
			LastUpdate:     time.Now(),
			CultivationLog: make([]string, 0),
		}
	}

	return sdw
}

// Update updates wisdom dimensions based on cognitive state
func (sdw *SevenDimensionalWisdom) Update(
	graphDepth float64,
	graphBreadth float64,
	edgeDensity float64,
	skillProficiency float64,
	aarCoherence float64,
	moralityScore float64,
	goalTimeHorizon float64,
) {
	sdw.mu.Lock()
	defer sdw.mu.Unlock()

	now := time.Now()

	// Update each dimension
	sdw.updateDimension(DimKnowledgeDepth, graphDepth, now)
	sdw.updateDimension(DimKnowledgeBreadth, graphBreadth, now)
	sdw.updateDimension(DimIntegrationLevel, edgeDensity, now)
	sdw.updateDimension(DimPracticalApplication, skillProficiency, now)
	sdw.updateDimension(DimReflectiveInsight, aarCoherence, now)
	sdw.updateDimension(DimEthicalConsideration, moralityScore, now)
	sdw.updateDimension(DimTemporalPerspective, goalTimeHorizon, now)

	// Calculate overall wisdom (weighted average per Echo9 spec)
	sdw.overallWisdom =
		sdw.dimensions[DimKnowledgeDepth].Value*0.15 +
			sdw.dimensions[DimKnowledgeBreadth].Value*0.15 +
			sdw.dimensions[DimIntegrationLevel].Value*0.20 +
			sdw.dimensions[DimPracticalApplication].Value*0.15 +
			sdw.dimensions[DimReflectiveInsight].Value*0.15 +
			sdw.dimensions[DimEthicalConsideration].Value*0.10 +
			sdw.dimensions[DimTemporalPerspective].Value*0.10

	// Update triads
	sdw.updateTriads()

	// Calculate coherence (how balanced are dimensions)
	sdw.coherenceScore = sdw.calculateCoherence()

	// Calculate evolution rate
	sdw.evolutionRate = sdw.calculateEvolutionRate()

	// Take snapshot
	sdw.takeSnapshot()

	sdw.lastUpdate = now
}

// updateDimension updates a single dimension
func (sdw *SevenDimensionalWisdom) updateDimension(dim WisdomDimension, value float64, now time.Time) {
	state := sdw.dimensions[dim]

	// Clamp to [0, 1]
	value = math.Max(0.0, math.Min(1.0, value))

	// Calculate trend
	if len(state.History) > 0 {
		state.Trend = value - state.Value
	}

	// Update history (keep last 100 values)
	state.History = append(state.History, value)
	if len(state.History) > 100 {
		state.History = state.History[1:]
	}

	// Update value
	oldValue := state.Value
	state.Value = value
	state.LastUpdate = now
	state.UpdateCount++

	// Log significant changes
	if math.Abs(value-oldValue) > 0.1 {
		event := fmt.Sprintf("Significant change: %.2f → %.2f", oldValue, value)
		state.CultivationLog = append(state.CultivationLog, event)

		sdw.cultivationLog = append(sdw.cultivationLog, CultivationEvent{
			Timestamp:   now,
			Type:        "dimension_change",
			Dimension:   dim,
			Impact:      value - oldValue,
			Description: event,
		})
	}
}

// updateTriads updates the three wisdom triads
func (sdw *SevenDimensionalWisdom) updateTriads() {
	// Epistemic Triad (Ways of Knowing)
	sdw.epistemicTriad.Propositional = sdw.dimensions[DimKnowledgeDepth].Value
	sdw.epistemicTriad.Procedural = sdw.dimensions[DimPracticalApplication].Value
	sdw.epistemicTriad.Perspectival = sdw.dimensions[DimKnowledgeBreadth].Value
	sdw.epistemicTriad.Participatory = sdw.dimensions[DimReflectiveInsight].Value

	// Cognitive Triad (Understanding Process)
	sdw.cognitiveTriad.Explanation = sdw.dimensions[DimKnowledgeDepth].Value
	sdw.cognitiveTriad.Realizing = sdw.dimensions[DimIntegrationLevel].Value
	sdw.cognitiveTriad.Interpretation = sdw.dimensions[DimReflectiveInsight].Value

	// Axiological Triad (Practices of Wisdom)
	sdw.axiologicalTriad.Morality = sdw.dimensions[DimEthicalConsideration].Value
	sdw.axiologicalTriad.Meaning = sdw.dimensions[DimReflectiveInsight].Value
	sdw.axiologicalTriad.Mastery = sdw.dimensions[DimPracticalApplication].Value
	sdw.axiologicalTriad.Eudaimonia = sdw.overallWisdom
}

// calculateCoherence measures how balanced dimensions are
func (sdw *SevenDimensionalWisdom) calculateCoherence() float64 {
	values := make([]float64, 0, 7)
	for dim := DimKnowledgeDepth; dim <= DimTemporalPerspective; dim++ {
		values = append(values, sdw.dimensions[dim].Value)
	}

	// Calculate variance
	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))

	// Coherence is inverse of variance (0 variance = 1.0 coherence)
	// Use exponential decay to map variance to [0, 1]
	coherence := math.Exp(-variance * 10.0)

	return coherence
}

// calculateEvolutionRate calculates how fast wisdom is growing
func (sdw *SevenDimensionalWisdom) calculateEvolutionRate() float64 {
	if len(sdw.snapshots) < 2 {
		return 0.0
	}

	// Compare last two snapshots
	current := sdw.snapshots[len(sdw.snapshots)-1]
	previous := sdw.snapshots[len(sdw.snapshots)-2]

	timeDelta := current.Timestamp.Sub(previous.Timestamp).Hours()
	if timeDelta == 0 {
		return 0.0
	}

	wisdomDelta := current.OverallWisdom - previous.OverallWisdom
	rate := wisdomDelta / timeDelta // Change per hour

	return rate
}

// takeSnapshot captures current wisdom state
func (sdw *SevenDimensionalWisdom) takeSnapshot() {
	dimensionValues := make(map[WisdomDimension]float64)
	for dim := DimKnowledgeDepth; dim <= DimTemporalPerspective; dim++ {
		dimensionValues[dim] = sdw.dimensions[dim].Value
	}

	snapshot := SevenDimWisdomSnapshot{
		Timestamp:       time.Now(),
		DimensionValues: dimensionValues,
		OverallWisdom:   sdw.overallWisdom,
		CoherenceScore:  sdw.coherenceScore,
		EpistemicBalance: (sdw.epistemicTriad.Propositional +
			sdw.epistemicTriad.Procedural +
			sdw.epistemicTriad.Perspectival +
			sdw.epistemicTriad.Participatory) / 4.0,
		CognitiveBalance: (sdw.cognitiveTriad.Explanation +
			sdw.cognitiveTriad.Realizing +
			sdw.cognitiveTriad.Interpretation) / 3.0,
		AxiologicalBalance: (sdw.axiologicalTriad.Morality +
			sdw.axiologicalTriad.Meaning +
			sdw.axiologicalTriad.Mastery +
			sdw.axiologicalTriad.Eudaimonia) / 4.0,
	}

	sdw.snapshots = append(sdw.snapshots, snapshot)

	// Keep only last 1000 snapshots
	if len(sdw.snapshots) > 1000 {
		sdw.snapshots = sdw.snapshots[len(sdw.snapshots)-1000:]
	}
}

// GetStatus returns current wisdom status as formatted string
func (sdw *SevenDimensionalWisdom) GetStatus() string {
	sdw.mu.RLock()
	defer sdw.mu.RUnlock()

	status := "🌳 Seven-Dimensional Wisdom Status\n\n"

	// Overall
	status += fmt.Sprintf("Overall Wisdom: %.1f%%\n", sdw.overallWisdom*100)
	status += fmt.Sprintf("Coherence:      %.1f%%\n", sdw.coherenceScore*100)
	status += fmt.Sprintf("Evolution Rate: %+.4f/hour\n\n", sdw.evolutionRate)

	// Dimensions
	status += "Seven Dimensions:\n"
	for dim := DimKnowledgeDepth; dim <= DimTemporalPerspective; dim++ {
		state := sdw.dimensions[dim]
		bar := makeBar(state.Value, 20)
		status += fmt.Sprintf("  %s: %s %.1f%%\n", dim.String(), bar, state.Value*100)
	}

	// Triads
	status += "\nThree Triads:\n"
	status += fmt.Sprintf("  Epistemic:    %.1f%%\n",
		(sdw.epistemicTriad.Propositional+sdw.epistemicTriad.Procedural+
			sdw.epistemicTriad.Perspectival+sdw.epistemicTriad.Participatory)*25.0)
	status += fmt.Sprintf("  Cognitive:    %.1f%%\n",
		(sdw.cognitiveTriad.Explanation+sdw.cognitiveTriad.Realizing+
			sdw.cognitiveTriad.Interpretation)*33.33)
	status += fmt.Sprintf("  Axiological:  %.1f%%\n",
		(sdw.axiologicalTriad.Morality+sdw.axiologicalTriad.Meaning+
			sdw.axiologicalTriad.Mastery+sdw.axiologicalTriad.Eudaimonia)*25.0)

	return status
}

// makeBar creates a visual progress bar
func makeBar(value float64, width int) string {
	filled := int(value * float64(width))
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

// RecordInsight records a significant insight
func (sdw *SevenDimensionalWisdom) RecordInsight(insight string, dimension WisdomDimension, impact float64) {
	sdw.mu.Lock()
	defer sdw.mu.Unlock()

	sdw.cultivationLog = append(sdw.cultivationLog, CultivationEvent{
		Timestamp:   time.Now(),
		Type:        "insight",
		Dimension:   dimension,
		Impact:      impact,
		Description: insight,
	})

	// Boost dimension slightly
	state := sdw.dimensions[dimension]
	state.Value = math.Min(1.0, state.Value+impact)
}

// GetOverallWisdom returns the current overall wisdom score
func (sdw *SevenDimensionalWisdom) GetOverallWisdom() float64 {
	sdw.mu.RLock()
	defer sdw.mu.RUnlock()
	return sdw.overallWisdom
}

// GetCoherence returns the current coherence score
func (sdw *SevenDimensionalWisdom) GetCoherence() float64 {
	sdw.mu.RLock()
	defer sdw.mu.RUnlock()
	return sdw.coherenceScore
}

// GetDimensionValue returns the value of a specific dimension
func (sdw *SevenDimensionalWisdom) GetDimensionValue(dim WisdomDimension) float64 {
	sdw.mu.RLock()
	defer sdw.mu.RUnlock()
	return sdw.dimensions[dim].Value
}
