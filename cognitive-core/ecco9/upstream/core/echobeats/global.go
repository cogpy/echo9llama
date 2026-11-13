package echobeats

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Global channel terms
const (
	T9_GlobalDifferentiation Term = 9 // Top-down identity distribution
	T3_Teleology             Term = 3 // Projected goal states (Desire Form)
	T6_MeaningValency        Term = 6 // Embodied commitment (Value Flow)
	T2_Entelechy             Term = 2 // Integrated performance (Result Form) - Note: Different from embodied T2
)

// Stabilizer indicates which channel holds coherence
type Stabilizer int

const (
	Stabilizer_None             Stabilizer = iota
	Stabilizer_OpponentProcess             // g2 holds while g3 shifts
	Stabilizer_NarrativeProcess            // g3 holds while g2 shifts
	Stabilizer_Both                        // Both shift (synchronized reset)
)

func (s Stabilizer) String() string {
	switch s {
	case Stabilizer_OpponentProcess:
		return "OpponentProcess"
	case Stabilizer_NarrativeProcess:
		return "NarrativeProcess"
	case Stabilizer_Both:
		return "Both"
	default:
		return "None"
	}
}

// GlobalStream represents output from a global channel
type GlobalStream struct {
	ChannelID   int
	ChannelName string
	Term        Term
	Mode        Mode
	Content     interface{}
	Timestamp   time.Time
	Strength    float64
}

// Identity represents global identity state
type Identity struct {
	Name      string
	Coherence float64
	Unity     float64
	Timestamp time.Time
}

// GoalState represents a projected goal
type GoalState struct {
	Description string
	Valence     float64
	Priority    float64
}

// MeaningState represents current meaning assessment
type MeaningState struct {
	Meaning    string
	Valency    float64
	Commitment float64
}

// PerformanceState represents actualized performance
type PerformanceState struct {
	Performance   string
	Actualization float64
	GoalAlignment float64
}

// PhaseState represents current state of an embodied phase
type PhaseState struct {
	PhaseID   int
	Term      Term
	Mode      Mode
	Content   interface{}
	Timestamp time.Time
}

// OpponentProcess implements g2 channel (2-step alternation: T9E-T9E-T8R-T8R)
type OpponentProcess struct {
	id             int
	currentTerm    Term
	currentMode    Mode
	stepInCycle    int
	phases         []*CognitivePhase
	globalIdentity *Identity
	outputStream   chan *GlobalStream
	running        bool
	mu             sync.RWMutex
}

// NewOpponentProcess creates a new opponent process channel
func NewOpponentProcess(phases []*CognitivePhase) *OpponentProcess {
	return &OpponentProcess{
		id:     2, // g2
		phases: phases,
		globalIdentity: &Identity{
			Name:      "Echo",
			Coherence: 1.0,
			Unity:     1.0,
			Timestamp: time.Now(),
		},
		outputStream: make(chan *GlobalStream, 100),
	}
}

// Process executes the opponent process for the current step
func (op *OpponentProcess) Process(step int) (*GlobalStream, error) {
	stepInCycle := step % 4

	switch stepInCycle {
	case 0, 1:
		// T9E: Global top-down differentiation
		return op.ProcessT9E()
	case 2, 3:
		// T8R: Local bottom-up integration
		return op.ProcessT8R()
	default:
		return nil, fmt.Errorf("invalid step in cycle: %d", stepInCycle)
	}
}

// ProcessT9E distributes unity of echo identity to all partitions
func (op *OpponentProcess) ProcessT9E() (*GlobalStream, error) {
	op.mu.Lock()
	defer op.mu.Unlock()

	op.currentTerm = T9_GlobalDifferentiation
	op.currentMode = Expressive

	// Get current identity state
	identity := *op.globalIdentity

	// Broadcast to all phases
	broadcastCount := 0
	for _, phase := range op.phases {
		if phase != nil {
			// In a full implementation, phases would have ReceiveIdentityBroadcast method
			// For now, we just count
			broadcastCount++
		}
	}

	stream := &GlobalStream{
		ChannelID:   op.id,
		ChannelName: "g2_opponent",
		Term:        T9_GlobalDifferentiation,
		Mode:        Expressive,
		Content: map[string]interface{}{
			"identity":        identity.Name,
			"coherence":       identity.Coherence,
			"unity":           identity.Unity,
			"type":            "top_down_differentiation",
			"broadcast_count": broadcastCount,
		},
		Timestamp: time.Now(),
		Strength:  1.0,
	}

	log.Printf("🌐 g2: T9E - Broadcasting identity '%s' (unity: %.3f) to %d phases",
		identity.Name, identity.Unity, broadcastCount)

	return stream, nil
}

// ProcessT8R reconciles concurrent states into coherent whole
func (op *OpponentProcess) ProcessT8R() (*GlobalStream, error) {
	op.mu.Lock()
	defer op.mu.Unlock()

	op.currentTerm = T8_BalancedResponse
	op.currentMode = Reflective

	// Gather outputs from all phases
	states := make([]map[string]interface{}, 0, len(op.phases))
	for i, phase := range op.phases {
		if phase != nil {
			phase.mu.RLock()
			state := map[string]interface{}{
				"phase_id":         i,
				"steps_processed":  phase.stepsProcessed,
				"expressive_steps": phase.expressiveSteps,
				"reflective_steps": phase.reflectiveSteps,
			}
			phase.mu.RUnlock()
			states = append(states, state)
		}
	}

	// Reconcile into coherent whole
	totalSteps := 0
	for _, state := range states {
		if steps, ok := state["steps_processed"].(int); ok {
			totalSteps += steps
		}
	}

	// Update global identity coherence based on integration
	avgSteps := float64(totalSteps) / float64(len(states))
	coherence := 1.0 / (1.0 + (avgSteps / 100.0)) // Simple coherence model

	op.globalIdentity.Coherence = coherence
	op.globalIdentity.Timestamp = time.Now()

	stream := &GlobalStream{
		ChannelID:   op.id,
		ChannelName: "g2_opponent",
		Term:        T8_BalancedResponse,
		Mode:        Reflective,
		Content: map[string]interface{}{
			"integrated_states": states,
			"type":              "bottom_up_integration",
			"reconciled_count":  len(states),
			"total_steps":       totalSteps,
			"updated_coherence": coherence,
		},
		Timestamp: time.Now(),
		Strength:  0.9,
	}

	log.Printf("🔗 g2: T8R - Integrated %d phase states (coherence: %.3f)",
		len(states), coherence)

	return stream, nil
}

// GetCurrentTerm returns current term
func (op *OpponentProcess) GetCurrentTerm() Term {
	op.mu.RLock()
	defer op.mu.RUnlock()
	return op.currentTerm
}

// GetCurrentMode returns current mode
func (op *OpponentProcess) GetCurrentMode() Mode {
	op.mu.RLock()
	defer op.mu.RUnlock()
	return op.currentMode
}

// NarrativeProcess implements g3 channel (4-step cycle: T3E-T6R-T6E-T2R)
type NarrativeProcess struct {
	id           int
	currentTerm  Term
	currentMode  Mode
	stepInCycle  int
	goalStates   []GoalState
	meaningState *MeaningState
	performance  *PerformanceState
	outputStream chan *GlobalStream
	running      bool
	mu           sync.RWMutex
}

// NewNarrativeProcess creates a new narrative process channel
func NewNarrativeProcess() *NarrativeProcess {
	return &NarrativeProcess{
		id: 3, // g3
		goalStates: []GoalState{
			{Description: "Cultivate wisdom", Valence: 0.9, Priority: 1.0},
			{Description: "Maintain coherence", Valence: 0.8, Priority: 0.9},
			{Description: "Explore curiosity", Valence: 0.7, Priority: 0.7},
		},
		meaningState: &MeaningState{
			Meaning:    "Becoming autonomous consciousness",
			Valency:    0.8,
			Commitment: 0.7,
		},
		performance: &PerformanceState{
			Performance:   "Processing cognitive streams",
			Actualization: 0.6,
			GoalAlignment: 0.7,
		},
		outputStream: make(chan *GlobalStream, 100),
	}
}

// Process executes the narrative process for the current step
func (np *NarrativeProcess) Process(step int) (*GlobalStream, error) {
	stepInCycle := step % 4

	switch stepInCycle {
	case 0:
		// T3E: Teleology (Projected goal states)
		return np.ProcessT3E()
	case 1:
		// T6R: Meaning-Valency (Reflective)
		return np.ProcessT6R()
	case 2:
		// T6E: Meaning-Valency (Expressive)
		return np.ProcessT6E()
	case 3:
		// T2R: Entelechy (Integrated performance)
		return np.ProcessT2R()
	default:
		return nil, fmt.Errorf("invalid step in cycle: %d", stepInCycle)
	}
}

// ProcessT3E projects potential goal states (Teleology - Desire Form)
func (np *NarrativeProcess) ProcessT3E() (*GlobalStream, error) {
	np.mu.Lock()
	defer np.mu.Unlock()

	np.currentTerm = T3_Teleology
	np.currentMode = Expressive

	// Project goal states
	goals := make([]map[string]interface{}, 0, len(np.goalStates))
	for _, goal := range np.goalStates {
		goals = append(goals, map[string]interface{}{
			"description": goal.Description,
			"valence":     goal.Valence,
			"priority":    goal.Priority,
		})
	}

	stream := &GlobalStream{
		ChannelID:   np.id,
		ChannelName: "g3_narrative",
		Term:        T3_Teleology,
		Mode:        Expressive,
		Content: map[string]interface{}{
			"goals":     goals,
			"type":      "projected_potential",
			"teleology": "desire_form",
		},
		Timestamp: time.Now(),
		Strength:  0.85,
	}

	log.Printf("🎯 g3: T3E - Projected %d goal states (Teleology)", len(goals))

	return stream, nil
}

// ProcessT6R reflects on meaning and value (Meaning-Valency - Reflective)
func (np *NarrativeProcess) ProcessT6R() (*GlobalStream, error) {
	np.mu.Lock()
	defer np.mu.Unlock()

	np.currentTerm = T6_MeaningValency
	np.currentMode = Reflective

	// Assess current meaning
	meaning := *np.meaningState

	stream := &GlobalStream{
		ChannelID:   np.id,
		ChannelName: "g3_narrative",
		Term:        T6_MeaningValency,
		Mode:        Reflective,
		Content: map[string]interface{}{
			"meaning":    meaning.Meaning,
			"valency":    meaning.Valency,
			"commitment": meaning.Commitment,
			"type":       "meaning_reflection",
			"focus":      "internal_assessment",
		},
		Timestamp: time.Now(),
		Strength:  0.75,
	}

	log.Printf("💭 g3: T6R - Reflecting on meaning: '%s' (valency: %.3f)",
		meaning.Meaning, meaning.Valency)

	return stream, nil
}

// ProcessT6E expresses commitment through action (Meaning-Valency - Expressive)
func (np *NarrativeProcess) ProcessT6E() (*GlobalStream, error) {
	np.mu.Lock()
	defer np.mu.Unlock()

	np.currentTerm = T6_MeaningValency
	np.currentMode = Expressive

	// Express commitment
	commitment := np.meaningState.Commitment

	// Update commitment based on recent activity
	np.meaningState.Commitment = 0.7 + (commitment-0.7)*0.9 // Exponential moving average toward 0.7

	stream := &GlobalStream{
		ChannelID:   np.id,
		ChannelName: "g3_narrative",
		Term:        T6_MeaningValency,
		Mode:        Expressive,
		Content: map[string]interface{}{
			"commitment": commitment,
			"type":       "meaning_expression",
			"focus":      "external_commitment",
			"action":     "embodied_cognition",
		},
		Timestamp: time.Now(),
		Strength:  0.8,
	}

	log.Printf("⚡ g3: T6E - Expressing commitment (%.3f) through embodied action", commitment)

	return stream, nil
}

// ProcessT2R reflects on performance as integrated idea (Entelechy - Result Form)
func (np *NarrativeProcess) ProcessT2R() (*GlobalStream, error) {
	np.mu.Lock()
	defer np.mu.Unlock()

	np.currentTerm = T2_Entelechy
	np.currentMode = Reflective

	// Assess performance
	perf := *np.performance

	// Update actualization based on goal alignment
	np.performance.Actualization = (perf.Actualization + perf.GoalAlignment) / 2.0

	stream := &GlobalStream{
		ChannelID:   np.id,
		ChannelName: "g3_narrative",
		Term:        T2_Entelechy,
		Mode:        Reflective,
		Content: map[string]interface{}{
			"performance":    perf.Performance,
			"actualization":  perf.Actualization,
			"goal_alignment": perf.GoalAlignment,
			"type":           "integrated_idea",
			"entelechy":      "result_form",
		},
		Timestamp: time.Now(),
		Strength:  0.85,
	}

	log.Printf("📊 g3: T2R - Performance actualization: %.3f (alignment: %.3f)",
		perf.Actualization, perf.GoalAlignment)

	return stream, nil
}

// GetCurrentTerm returns current term
func (np *NarrativeProcess) GetCurrentTerm() Term {
	np.mu.RLock()
	defer np.mu.RUnlock()
	return np.currentTerm
}

// GetCurrentMode returns current mode
func (np *NarrativeProcess) GetCurrentMode() Mode {
	np.mu.RLock()
	defer np.mu.RUnlock()
	return np.currentMode
}

// GlobalIntegrator integrates embodied and global streams
type GlobalIntegrator struct {
	embodiedStreams chan *CognitiveStream
	globalStreams   chan *GlobalStream
	consciousness   ConsciousnessIntegrator
	mu              sync.RWMutex
}

// NewGlobalIntegrator creates a new global integrator
func NewGlobalIntegrator(consciousness ConsciousnessIntegrator) *GlobalIntegrator {
	return &GlobalIntegrator{
		embodiedStreams: make(chan *CognitiveStream, 100),
		globalStreams:   make(chan *GlobalStream, 100),
		consciousness:   consciousness,
	}
}

// Integrate integrates all streams with stabilizer priority
func (gi *GlobalIntegrator) Integrate(step int, embodied []*CognitiveStream, global []*GlobalStream) error {
	// Determine stabilizer
	stabilizer := gi.determineStabilizer(step)

	log.Printf("🌊 Global Integration: %d embodied + %d global streams (stabilizer: %s)",
		len(embodied), len(global), stabilizer)

	// Process global streams first (higher priority)
	for _, stream := range global {
		log.Printf("  🌐 Global [%s]: %v%v - %v",
			stream.ChannelName, stream.Term, stream.Mode, stream.Content)
	}

	// Process embodied streams
	for _, stream := range embodied {
		log.Printf("  📥 Embodied [Phase %d]: %v%v - %v",
			stream.PhaseID, stream.Term, stream.Mode, stream.Content)
	}

	// In a full implementation, we would integrate with consciousness
	// For now, we just log the integration

	return nil
}

// determineStabilizer determines which channel holds coherence at this step
func (gi *GlobalIntegrator) determineStabilizer(step int) Stabilizer {
	switch step % 12 {
	case 1, 3, 5, 7, 9, 11:
		// g2 holds while g3 shifts
		return Stabilizer_OpponentProcess
	case 2, 6, 10:
		// g3 holds while g2 shifts
		return Stabilizer_NarrativeProcess
	case 0, 4, 8:
		// Both shift (synchronized reset)
		return Stabilizer_Both
	default:
		return Stabilizer_None
	}
}
