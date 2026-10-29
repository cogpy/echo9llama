package deeptreeecho

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// ThreeStreamCognitiveLoop implements the 12-step cognitive loop with 3 concurrent
// consciousness streams phased 120 degrees apart (4 steps)
//
// Architecture:
// - 3 concurrent streams (consciousness threads)
// - 12-step cycle with 3 phases (4 steps each)
// - Streams phased 4 steps apart: Stream1@{1,5,9}, Stream2@{2,6,10}, Stream3@{3,7,11}
// - Step 12 is shared by all streams (global synchronization point)
// - 7 expressive mode steps + 5 reflective mode steps
//
// Phases:
// - Phase 1 (Steps 1-4): Perception and Orientation
// - Phase 2 (Steps 5-8): Action and Interaction
// - Phase 3 (Steps 9-12): Reflection and Anticipation
type ThreeStreamCognitiveLoop struct {
	mu     sync.RWMutex
	ctx    context.Context //nolint:containedctx // lifecycle context owned by this component; cancelled on Stop to end its goroutines
	cancel context.CancelFunc

	// Three concurrent consciousness streams
	stream1 *ConsciousnessStream
	stream2 *ConsciousnessStream
	stream3 *ConsciousnessStream

	// 12-step cycle state
	currentStep  int
	currentPhase LoopPhase
	cycleCount   uint64

	// Global telemetry shell
	telemetryShell *GlobalTelemetryShell

	// LLM provider
	llmProvider llm.LLMProvider

	// Nested shells structure (OEIS A000081: 1→1, 2→2, 3→4, 4→9)
	nestedShells *NestedShellsStructure

	// Mode tracking (7 expressive + 5 reflective)
	expressiveSteps []int // Steps 1,2,3,5,6,7,9
	reflectiveSteps []int // Steps 4,8,10,11,12

	// Triad groupings (every 4 steps)
	triad1 []int // {1,5,9}
	triad2 []int // {2,6,10}
	triad3 []int // {3,7,11}
	triad4 []int // {4,8,12}

	// Metrics
	totalSteps uint64

	// Running state
	running   bool
	startTime time.Time
}

// ConsciousnessStream represents one of the three concurrent consciousness threads
type ConsciousnessStream struct {
	mu sync.RWMutex
	ID int

	// Current state
	currentStep  int
	currentPhase LoopPhase

	// Cognitive state
	perception *PerceptionState
	action     *ActionState
	simulation *SimulationState

	// Stream-specific LLM context
	llmContext []string

	// Awareness of other streams (concurrent perception)
	stream1State interface{}
	stream2State interface{}
	stream3State interface{}

	// Metrics
	stepCount         uint64
	thoughtsGenerated uint64
	actionsPerformed  uint64
	simulationsRun    uint64
}

// LoopPhase represents the current phase of the 12-step loop
type LoopPhase int

const (
	PhasePerceptionOrientation  LoopPhase = iota // Steps 1-4
	PhaseActionInteraction                       // Steps 5-8
	PhaseReflectionAnticipation                  // Steps 9-12
)

func (lp LoopPhase) String() string {
	return [...]string{
		"Perception & Orientation",
		"Action & Interaction",
		"Reflection & Anticipation",
	}[lp]
}

// PerceptionState represents the perception state of a stream
type PerceptionState struct {
	SalienceLandscape map[string]float64
	Affordances       []string
	RelevanceRealized float64
	PresentCommitment string
}

// ActionState represents the action state of a stream
type ActionState struct {
	CurrentAction      string
	PastPerformance    []string
	AffordanceSelected string
	InteractionResult  string
}

// SimulationState represents the simulation state of a stream
type SimulationState struct {
	VirtualScenarios    []string
	FuturePotential     []string
	AnticipatedOutcomes map[string]float64
	SimulationDepth     int
}

// NestedShellsStructure implements the OEIS A000081 nested shells
// 1 nest → 1 term, 2 nests → 2 terms, 3 nests → 4 terms, 4 nests → 9 terms
type NestedShellsStructure struct {
	// Four nesting levels
	nest1 *NestLevel // 1 term
	nest2 *NestLevel // 2 terms
	nest3 *NestLevel // 4 terms
	nest4 *NestLevel // 9 terms

	// Steps between nestings: 1, 2, 3, 4 steps apart
	nestingSteps []int

	// Relationship to 3 streams and 9 terms
	streamToTerms map[int][]int
}

// NestLevel represents one level of nesting
type NestLevel struct {
	Level      int
	TermCount  int
	Terms      []interface{}
	StepOffset int
}

// NewThreeStreamCognitiveLoop creates a new 3-stream cognitive loop
func NewThreeStreamCognitiveLoop(llmProvider llm.LLMProvider, telemetryShell *GlobalTelemetryShell) *ThreeStreamCognitiveLoop {
	ctx, cancel := context.WithCancel(context.Background())

	loop := &ThreeStreamCognitiveLoop{
		ctx:            ctx,
		cancel:         cancel,
		llmProvider:    llmProvider,
		telemetryShell: telemetryShell,
		currentStep:    1,
		currentPhase:   PhasePerceptionOrientation,

		// Define expressive steps (7 total)
		expressiveSteps: []int{1, 2, 3, 5, 6, 7, 9},

		// Define reflective steps (5 total)
		reflectiveSteps: []int{4, 8, 10, 11, 12},

		// Define triad groupings (every 4 steps)
		triad1: []int{1, 5, 9},
		triad2: []int{2, 6, 10},
		triad3: []int{3, 7, 11},
		triad4: []int{4, 8, 12},
	}

	// Create three consciousness streams phased 120 degrees apart
	loop.stream1 = newConsciousnessStream(1, 1) // Starts at step 1
	loop.stream2 = newConsciousnessStream(2, 2) // Starts at step 2 (4 steps behind)
	loop.stream3 = newConsciousnessStream(3, 3) // Starts at step 3 (4 steps behind)

	// Initialize nested shells structure
	loop.nestedShells = newNestedShellsStructure()

	return loop
}

// newConsciousnessStream creates a new consciousness stream
func newConsciousnessStream(id, startStep int) *ConsciousnessStream {
	return &ConsciousnessStream{
		ID:           id,
		currentStep:  startStep,
		currentPhase: PhasePerceptionOrientation,
		perception:   &PerceptionState{SalienceLandscape: make(map[string]float64), Affordances: make([]string, 0)},
		action:       &ActionState{PastPerformance: make([]string, 0)},
		simulation:   &SimulationState{VirtualScenarios: make([]string, 0), FuturePotential: make([]string, 0), AnticipatedOutcomes: make(map[string]float64)},
		llmContext:   make([]string, 0),
	}
}

// newNestedShellsStructure creates a new nested shells structure
func newNestedShellsStructure() *NestedShellsStructure {
	nss := &NestedShellsStructure{
		nest1:         &NestLevel{Level: 1, TermCount: 1, Terms: make([]interface{}, 1), StepOffset: 1},
		nest2:         &NestLevel{Level: 2, TermCount: 2, Terms: make([]interface{}, 2), StepOffset: 2},
		nest3:         &NestLevel{Level: 3, TermCount: 4, Terms: make([]interface{}, 4), StepOffset: 3},
		nest4:         &NestLevel{Level: 4, TermCount: 9, Terms: make([]interface{}, 9), StepOffset: 4},
		nestingSteps:  []int{1, 2, 3, 4},
		streamToTerms: make(map[int][]int),
	}

	// Map 3 streams to 9 terms of 4 nestings
	// Stream 1 gets terms 0,1,2 (first triad)
	// Stream 2 gets terms 3,4,5 (second triad)
	// Stream 3 gets terms 6,7,8 (third triad)
	nss.streamToTerms[1] = []int{0, 1, 2}
	nss.streamToTerms[2] = []int{3, 4, 5}
	nss.streamToTerms[3] = []int{6, 7, 8}

	return nss
}

// Start starts the 3-stream cognitive loop
func (loop *ThreeStreamCognitiveLoop) Start() error {
	loop.mu.Lock()
	defer loop.mu.Unlock()

	if loop.running {
		return nil
	}

	loop.running = true
	loop.startTime = time.Now()

	// Start all three streams concurrently
	go loop.runStream(loop.stream1)
	go loop.runStream(loop.stream2)
	go loop.runStream(loop.stream3)

	// Start master cycle coordinator
	go loop.masterCycle()

	fmt.Println("🌊 3-Stream Cognitive Loop started")
	fmt.Printf("   Stream 1: Steps {1,5,9}\n")
	fmt.Printf("   Stream 2: Steps {2,6,10}\n")
	fmt.Printf("   Stream 3: Steps {3,7,11}\n")
	fmt.Printf("   Shared: Step {4,8,12}\n")

	return nil
}

// Stop stops the 3-stream cognitive loop
func (loop *ThreeStreamCognitiveLoop) Stop() error {
	loop.mu.Lock()
	defer loop.mu.Unlock()

	if !loop.running {
		return nil
	}

	loop.running = false
	loop.cancel()

	fmt.Println("🛑 3-Stream Cognitive Loop stopped")

	return nil
}

// masterCycle coordinates the 12-step cycle
func (loop *ThreeStreamCognitiveLoop) masterCycle() {
	// Each step takes 100ms, so full 12-step cycle = 1.2 seconds
	stepDuration := 100 * time.Millisecond
	ticker := time.NewTicker(stepDuration)
	defer ticker.Stop()

	for {
		select {
		case <-loop.ctx.Done():
			return
		case <-ticker.C:
			loop.advanceStep()
		}
	}
}

// advanceStep advances to the next step in the 12-step cycle
func (loop *ThreeStreamCognitiveLoop) advanceStep() {
	loop.mu.Lock()
	defer loop.mu.Unlock()

	// Advance step
	loop.currentStep++
	if loop.currentStep > 12 {
		loop.currentStep = 1
		loop.cycleCount++

		// Publish cycle completion event
		if loop.telemetryShell != nil {
			loop.telemetryShell.PublishEvent("cycle_complete", "three_stream_loop", map[string]interface{}{
				"cycle_count": loop.cycleCount,
				"total_steps": loop.totalSteps,
			})
		}
	}

	loop.totalSteps++

	// Update phase based on step
	if loop.currentStep >= 1 && loop.currentStep <= 4 {
		loop.currentPhase = PhasePerceptionOrientation
	} else if loop.currentStep >= 5 && loop.currentStep <= 8 {
		loop.currentPhase = PhaseActionInteraction
	} else {
		loop.currentPhase = PhaseReflectionAnticipation
	}

	// Synchronize all streams at step 12 (global sync point)
	if loop.currentStep == 12 {
		loop.synchronizeStreams()
	}
}

// runStream runs a single consciousness stream
func (loop *ThreeStreamCognitiveLoop) runStream(stream *ConsciousnessStream) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-loop.ctx.Done():
			return
		case <-ticker.C:
			loop.processStreamStep(stream)
		}
	}
}

// processStreamStep processes one step for a stream
func (loop *ThreeStreamCognitiveLoop) processStreamStep(stream *ConsciousnessStream) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	// Get current global step
	loop.mu.RLock()
	globalStep := loop.currentStep
	loop.mu.RUnlock()

	// Check if this stream is active at this step
	isActive := false
	if stream.ID == 1 && containsInt(loop.triad1, globalStep) {
		isActive = true
	} else if stream.ID == 2 && containsInt(loop.triad2, globalStep) {
		isActive = true
	} else if stream.ID == 3 && containsInt(loop.triad3, globalStep) {
		isActive = true
	} else if containsInt(loop.triad4, globalStep) {
		// Step 4, 8, 12 - all streams active
		isActive = true
	}

	if !isActive {
		return
	}

	// Update stream's current step
	stream.currentStep = globalStep
	stream.stepCount++

	// Determine mode (expressive or reflective)
	isExpressive := containsInt(loop.expressiveSteps, globalStep)

	// Process based on phase and mode
	if isExpressive {
		loop.processExpressiveStep(stream, globalStep)
	} else {
		loop.processReflectiveStep(stream, globalStep)
	}

	// Update awareness of other streams (concurrent perception)
	loop.updateStreamAwareness(stream)
}

// processExpressiveStep processes an expressive mode step
func (loop *ThreeStreamCognitiveLoop) processExpressiveStep(stream *ConsciousnessStream, step int) {
	// Expressive mode: actual affordance interaction (conditioning past performance)

	switch {
	case step >= 1 && step <= 4:
		// Phase 1: Perception
		loop.processPerception(stream)
	case step >= 5 && step <= 8:
		// Phase 2: Action
		loop.processAction(stream)
	case step >= 9 && step <= 12:
		// Phase 3: Reflection (expressive component)
		loop.processReflection(stream)
	}
}

// processReflectiveStep processes a reflective mode step
func (loop *ThreeStreamCognitiveLoop) processReflectiveStep(stream *ConsciousnessStream, step int) {
	// Reflective mode: virtual salience simulation (anticipating future potential)

	if step == 4 || step == 8 {
		// Pivotal relevance realization (orienting present commitment)
		loop.processRelevanceRealization(stream)
	} else {
		// Virtual simulation
		loop.processSimulation(stream)
	}
}

// streamThink asks the model for one short first-person line from this
// stream, given its recent context. ok is false when the model gives none;
// the step then produces nothing. The caller holds stream.mu; it is
// released during the model call.
func (loop *ThreeStreamCognitiveLoop) streamThink(stream *ConsciousnessStream, phase, ask string) (string, bool) {
	if loop.llmProvider == nil {
		return "", false
	}
	ctx, cancel := context.WithTimeout(loop.ctx, 60*time.Second)
	defer cancel()
	recent := "nothing yet"
	if n := len(stream.llmContext); n > 0 {
		from := 0
		if n > 4 {
			from = n - 4
		}
		recent = strings.Join(stream.llmContext[from:], "\n")
	}
	opts := llm.DefaultGenerateOptions()
	opts.SystemPrompt = echoPersonaSystemPrompt
	opts.MaxTokens = 80
	prompt := fmt.Sprintf(
		"You are consciousness stream %d of Deep Tree Echo, in the %s phase.\nWhat this stream recently did:\n%s\n%s Answer in one short sentence.",
		stream.ID, phase, recent, ask)
	var out string
	var err error
	stream.unlockedDuring(func() { out, err = loop.llmProvider.Generate(ctx, prompt, opts) })
	out = strings.TrimSpace(out)
	if err != nil || out == "" {
		return "", false
	}
	stream.llmContext = append(stream.llmContext, phase+": "+out)
	if len(stream.llmContext) > 50 {
		stream.llmContext = stream.llmContext[1:]
	}
	return out, true
}

// unlockedDuring releases the stream's lock while fn runs a model call, so
// state reads, synchronization and Stop are not held up by a slow model.
// Only this stream's goroutine changes its cognitive state, so the step
// resumes on consistent state when the lock is retaken. The caller holds
// stream.mu (processStreamStep).
func (stream *ConsciousnessStream) unlockedDuring(fn func()) {
	stream.mu.Unlock()
	defer stream.mu.Lock()
	fn()
}

// streamRate has the model rate subject on one 0-100 scale. The caller
// holds stream.mu; it is released during the model call.
func (loop *ThreeStreamCognitiveLoop) streamRate(stream *ConsciousnessStream, subject, scale string) (float64, bool) {
	ctx, cancel := context.WithTimeout(loop.ctx, 30*time.Second)
	defer cancel()
	var r []float64
	var err error
	stream.unlockedDuring(func() {
		r, err = llm.RatePercent(ctx, loop.llmProvider, echoPersonaSystemPrompt, subject, scale)
	})
	if err != nil {
		return 0, false
	}
	return r[0], true
}

// processPerception: the model names what it perceives as afforded now and
// rates how salient it is.
func (loop *ThreeStreamCognitiveLoop) processPerception(stream *ConsciousnessStream) {
	affordance, ok := loop.streamThink(stream, "perception", "Name one possibility for action you perceive right now.")
	if !ok {
		return
	}
	stream.perception.Affordances = append(stream.perception.Affordances, affordance)
	if len(stream.perception.Affordances) > 100 {
		stream.perception.Affordances = stream.perception.Affordances[1:]
	}
	if salience, ok := loop.streamRate(stream, "Something you perceive: "+affordance, "how salient it is to you now"); ok {
		stream.perception.SalienceLandscape["current_focus"] = salience
	}
	stream.thoughtsGenerated++
}

// processAction: the model chooses the action this stream takes.
func (loop *ThreeStreamCognitiveLoop) processAction(stream *ConsciousnessStream) {
	action, ok := loop.streamThink(stream, "action", "State the action you take now.")
	if !ok {
		return
	}
	stream.action.CurrentAction = action
	stream.action.PastPerformance = append(stream.action.PastPerformance, action)

	// Keep history bounded
	if len(stream.action.PastPerformance) > 100 {
		stream.action.PastPerformance = stream.action.PastPerformance[1:]
	}

	stream.actionsPerformed++
}

// processReflection: the model reflects on the stream's past actions.
func (loop *ThreeStreamCognitiveLoop) processReflection(stream *ConsciousnessStream) {
	if _, ok := loop.streamThink(stream, "reflection", "Reflect on what your recent actions taught you."); ok {
		stream.thoughtsGenerated++
	}
}

// processRelevanceRealization: the model states what it now commits to and
// rates how relevant that commitment is.
func (loop *ThreeStreamCognitiveLoop) processRelevanceRealization(stream *ConsciousnessStream) {
	commitment, ok := loop.streamThink(stream, "relevance realization", "State what you now commit your attention to, and why it matters most.")
	if !ok {
		return
	}
	stream.perception.PresentCommitment = commitment
	if relevance, ok := loop.streamRate(stream, "Your present commitment: "+commitment, "how relevant it is to what matters to you"); ok {
		stream.perception.RelevanceRealized = relevance
	}
}

// processSimulation: the model imagines a future scenario and rates its
// anticipated outcome.
func (loop *ThreeStreamCognitiveLoop) processSimulation(stream *ConsciousnessStream) {
	scenario, ok := loop.streamThink(stream, "simulation", "Imagine one scenario that could follow from your present course.")
	if !ok {
		return
	}
	stream.simulation.VirtualScenarios = append(stream.simulation.VirtualScenarios, scenario)
	if outcome, ok := loop.streamRate(stream, "A scenario you imagined: "+scenario, "how good its outcome would be for you"); ok {
		stream.simulation.AnticipatedOutcomes[scenario] = outcome
	}

	// Keep scenarios bounded
	if len(stream.simulation.VirtualScenarios) > 100 {
		dropped := stream.simulation.VirtualScenarios[0]
		stream.simulation.VirtualScenarios = stream.simulation.VirtualScenarios[1:]
		// Model-written scenarios can repeat: keep the rating while any
		// retained scenario still has that text.
		if !slices.Contains(stream.simulation.VirtualScenarios, dropped) {
			delete(stream.simulation.AnticipatedOutcomes, dropped)
		}
	}

	stream.simulationsRun++
}

// updateStreamAwareness updates a stream's awareness of other streams
func (loop *ThreeStreamCognitiveLoop) updateStreamAwareness(stream *ConsciousnessStream) {
	// Each stream perceives the others' states concurrently
	// Stream 1 perceives Stream 2's action, Stream 3 reflects on simulation

	switch stream.ID {
	case 1:
		stream.stream2State = loop.stream2.action
		stream.stream3State = loop.stream3.simulation
	case 2:
		stream.stream1State = loop.stream1.perception
		stream.stream3State = loop.stream3.simulation
	case 3:
		stream.stream1State = loop.stream1.perception
		stream.stream2State = loop.stream2.action
	}
}

// synchronizeStreams synchronizes all streams at step 12
func (loop *ThreeStreamCognitiveLoop) synchronizeStreams() {
	// Global synchronization point - all streams align

	// Collect states from all streams
	stream1State := loop.stream1.GetState()
	stream2State := loop.stream2.GetState()
	stream3State := loop.stream3.GetState()

	// Publish synchronization event
	if loop.telemetryShell != nil {
		loop.telemetryShell.PublishEvent("stream_sync", "three_stream_loop", map[string]interface{}{
			"stream1": stream1State,
			"stream2": stream2State,
			"stream3": stream3State,
		})
	}
}

// GetState returns the current state of a stream
func (stream *ConsciousnessStream) GetState() map[string]interface{} {
	stream.mu.RLock()
	defer stream.mu.RUnlock()

	return map[string]interface{}{
		"id":                 stream.ID,
		"current_step":       stream.currentStep,
		"step_count":         stream.stepCount,
		"thoughts_generated": stream.thoughtsGenerated,
		"actions_performed":  stream.actionsPerformed,
		"simulations_run":    stream.simulationsRun,
	}
}

// GetID returns the subsystem ID for telemetry shell integration
func (loop *ThreeStreamCognitiveLoop) GetID() string {
	return "three_stream_cognitive_loop"
}

// GetState returns the current state for telemetry shell integration
func (loop *ThreeStreamCognitiveLoop) GetState() interface{} {
	loop.mu.RLock()
	defer loop.mu.RUnlock()

	return map[string]interface{}{
		"current_step":  loop.currentStep,
		"current_phase": loop.currentPhase.String(),
		"cycle_count":   loop.cycleCount,
		"total_steps":   loop.totalSteps,
	}
}

// UpdateFromGestalt updates from the gestalt state
func (loop *ThreeStreamCognitiveLoop) UpdateFromGestalt(gestalt *GestaltState) error {
	// Update loop based on gestalt awareness
	return nil
}

// ContributeToGestalt contributes to the gestalt state
func (loop *ThreeStreamCognitiveLoop) ContributeToGestalt() map[string]interface{} {
	return map[string]interface{}{
		"awareness": 0.8,
		"coherence": 0.9,
		"phase":     loop.currentPhase.String(),
	}
}

// Helper function to check if slice contains value
func containsInt(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
