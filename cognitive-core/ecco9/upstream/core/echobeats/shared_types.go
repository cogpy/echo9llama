package echobeats

import (
	"sync"
	"time"
)

// =============================================================================
// SHARED TYPES FOR ALL ECHOBEATS SCHEDULER VARIANTS
// This file consolidates type definitions to avoid redeclaration conflicts
// =============================================================================

// CognitiveMode represents the processing mode
// Note: This is defined in cognitive_loop.go as the canonical version
// type CognitiveMode string
// const (
//   ModeExpressive           CognitiveMode = "expressive"
//   ModeReflective           CognitiveMode = "reflective"
//   ModeRelevanceRealization CognitiveMode = "relevance_realization"
//   ModeMetaCognitive        CognitiveMode = "metacognitive"
// )

// CognitivePhaseType represents the three phases of cognitive processing
type CognitivePhaseType int

const (
	PhaseAffordance CognitivePhaseType = iota // Past performance (conditioning)
	PhaseRelevance                            // Present commitment (orienting)
	PhaseSalience                             // Future potential (anticipating)
)

func (p CognitivePhaseType) String() string {
	return [...]string{"Affordance", "Relevance", "Salience"}[p]
}

// StepContext provides context for step execution in 12-step loop
type StepContext struct {
	StepNumber      int
	Phase           int
	Mode            CognitiveMode
	PreviousOutputs map[int]interface{}
	SharedState     map[string]interface{}
	Timestamp       time.Time
}

// StepHandler is a function that handles a specific step in the cognitive loop
type StepHandler func(context *StepContext) error

// CognitiveLoopMetrics tracks performance metrics for cognitive loops
type CognitiveLoopMetrics struct {
	mu                  sync.RWMutex
	CyclesCompleted     uint64
	StepsProcessed      uint64
	AverageStepDuration time.Duration
	PhaseTransitions    map[CognitivePhaseType]uint64
	ModeDistribution    map[CognitiveMode]uint64
}

// NewCognitiveLoopMetrics creates a new metrics tracker
func NewCognitiveLoopMetrics() *CognitiveLoopMetrics {
	return &CognitiveLoopMetrics{
		PhaseTransitions: make(map[CognitivePhaseType]uint64),
		ModeDistribution: make(map[CognitiveMode]uint64),
	}
}

// RecordCycle increments the cycle counter
func (m *CognitiveLoopMetrics) RecordCycle() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CyclesCompleted++
}

// RecordStep increments the step counter and updates average duration
func (m *CognitiveLoopMetrics) RecordStep(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StepsProcessed++

	// Update running average
	if m.AverageStepDuration == 0 {
		m.AverageStepDuration = duration
	} else {
		// Exponential moving average
		alpha := 0.1
		m.AverageStepDuration = time.Duration(
			float64(m.AverageStepDuration)*(1-alpha) + float64(duration)*alpha,
		)
	}
}

// RecordPhaseTransition records a transition to a new phase
func (m *CognitiveLoopMetrics) RecordPhaseTransition(phase CognitivePhaseType) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PhaseTransitions[phase]++
}

// RecordMode records execution in a specific mode
func (m *CognitiveLoopMetrics) RecordMode(mode CognitiveMode) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ModeDistribution[mode]++
}

// GetMetrics returns a snapshot of current metrics
func (m *CognitiveLoopMetrics) GetMetrics() (uint64, uint64, time.Duration) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.CyclesCompleted, m.StepsProcessed, m.AverageStepDuration
}

// StepExecution records execution of a cognitive step
// Unified definition to avoid redeclaration conflicts
type StepExecution struct {
	StepNumber   int                    `json:"step_number"`
	PhaseType    CognitivePhaseType     `json:"phase_type"`
	Mode         CognitiveMode          `json:"mode"`
	Timestamp    time.Time              `json:"timestamp"`
	StartTime    time.Time              `json:"start_time"`
	Duration     time.Duration          `json:"duration"`
	Success      bool                   `json:"success"`
	Output       interface{}            `json:"output"`
	Error        error                  `json:"error,omitempty"`
	EngineID     int                    `json:"engine_id,omitempty"`
	StateUpdates map[string]interface{} `json:"state_updates,omitempty"`
}

// StepType categorizes the 12 steps
type StepType int

const (
	StepRelevanceRealization  StepType = iota // Steps 1, 7
	StepAffordanceInteraction                 // Steps 2-6
	StepSalienceSimulation                    // Steps 8-12
)

func (s StepType) String() string {
	return [...]string{"RelevanceRealization", "AffordanceInteraction", "SalienceSimulation"}[s]
}
