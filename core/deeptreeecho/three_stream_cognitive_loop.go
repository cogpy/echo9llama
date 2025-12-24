package deeptreeecho

import (
	"context"
	"fmt"
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
	mu                  sync.RWMutex
	ctx                 context.Context
	cancel              context.CancelFunc
	
	// Three concurrent consciousness streams
	stream1             *ConsciousnessStream
	stream2             *ConsciousnessStream
	stream3             *ConsciousnessStream
	
	// 12-step cycle state
	currentStep         int
	currentPhase        LoopPhase
	cycleCount          uint64
	
	// Global telemetry shell
	telemetryShell      *GlobalTelemetryShell
	
	// LLM provider
	llmProvider         llm.LLMProvider
	
	// Nested shells structure (OEIS A000081: 1→1, 2→2, 3→4, 4→9)
	nestedShells        *NestedShellsStructure
	
	// Mode tracking (7 expressive + 5 reflective)
	expressiveSteps     []int // Steps 1,2,3,5,6,7,9
	reflectiveSteps     []int // Steps 4,8,10,11,12
	
	// Triad groupings (every 4 steps)
	triad1              []int // {1,5,9}
	triad2              []int // {2,6,10}
	triad3              []int // {3,7,11}
	triad4              []int // {4,8,12}
	
	// Metrics
	totalSteps          uint64
	totalCycles         uint64
	
	// Running state
	running             bool
	startTime           time.Time
}

// ConsciousnessStream represents one of the three concurrent consciousness threads
type ConsciousnessStream struct {
	mu                  sync.RWMutex
	ID                  int
	
	// Current state
	currentStep         int
	currentPhase        LoopPhase
	
	// Cognitive state
	perception          *PerceptionState
	action              *ActionState
	simulation          *SimulationState
	
	// Stream-specific LLM context
	llmContext          []string
	
	// Awareness of other streams (concurrent perception)
	stream1State        interface{}
	stream2State        interface{}
	stream3State        interface{}
	
	// Metrics
	stepCount           uint64
	thoughtsGenerated   uint64
	actionsPerformed    uint64
	simulationsRun      uint64
}

// LoopPhase represents the current phase of the 12-step loop
type LoopPhase int

const (
	PhasePerceptionOrientation LoopPhase = iota // Steps 1-4
	PhaseActionInteraction                      // Steps 5-8
	PhaseReflectionAnticipation                 // Steps 9-12
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
	SalienceLandscape   map[string]float64
	Affordances         []string
	RelevanceRealized   float64
	PresentCommitment   string
}

// ActionState represents the action state of a stream
type ActionState struct {
	CurrentAction       string
	PastPerformance     []string
	AffordanceSelected  string
	InteractionResult   string
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
	mu                  sync.RWMutex
	
	// Four nesting levels
	nest1               *NestLevel // 1 term
	nest2               *NestLevel // 2 terms
	nest3               *NestLevel // 4 terms
	nest4               *NestLevel // 9 terms
	
	// Steps between nestings: 1, 2, 3, 4 steps apart
	nestingSteps        []int
	
	// Relationship to 3 streams and 9 terms
	streamToTerms       map[int][]int
}

// NestLevel represents one level of nesting
type NestLevel struct {
	Level               int
	TermCount           int
	Terms               []interface{}
	StepOffset          int
}

// NewThreeStreamCognitiveLoop creates a new 3-stream cognitive loop
func NewThreeStreamCognitiveLoop(llmProvider llm.LLMProvider, telemetryShell *GlobalTelemetryShell) *ThreeStreamCognitiveLoop {
	ctx, cancel := context.WithCancel(context.Background())
	
	loop := &ThreeStreamCognitiveLoop{
		ctx:                ctx,
		cancel:             cancel,
		llmProvider:        llmProvider,
		telemetryShell:     telemetryShell,
		currentStep:        1,
		currentPhase:       PhasePerceptionOrientation,
		
		// Define expressive steps (7 total)
		expressiveSteps:    []int{1, 2, 3, 5, 6, 7, 9},
		
		// Define reflective steps (5 total)
		reflectiveSteps:    []int{4, 8, 10, 11, 12},
		
		// Define triad groupings (every 4 steps)
		triad1:             []int{1, 5, 9},
		triad2:             []int{2, 6, 10},
		triad3:             []int{3, 7, 11},
		triad4:             []int{4, 8, 12},
	}
	
	// Create three consciousness streams phased 120 degrees apart
	loop.stream1 = newConsciousnessStream(1, 1)  // Starts at step 1
	loop.stream2 = newConsciousnessStream(2, 2)  // Starts at step 2 (4 steps behind)
	loop.stream3 = newConsciousnessStream(3, 3)  // Starts at step 3 (4 steps behind)
	
	// Initialize nested shells structure
	loop.nestedShells = newNestedShellsStructure()
	
	return loop
}

// newConsciousnessStream creates a new consciousness stream
func newConsciousnessStream(id, startStep int) *ConsciousnessStream {
	return &ConsciousnessStream{
		ID:                id,
		currentStep:       startStep,
		currentPhase:      PhasePerceptionOrientation,
		perception:        &PerceptionState{SalienceLandscape: make(map[string]float64), Affordances: make([]string, 0)},
		action:            &ActionState{PastPerformance: make([]string, 0)},
		simulation:        &SimulationState{VirtualScenarios: make([]string, 0), FuturePotential: make([]string, 0), AnticipatedOutcomes: make(map[string]float64)},
		llmContext:        make([]string, 0),
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

// processPerception processes perception for a stream
func (loop *ThreeStreamCognitiveLoop) processPerception(stream *ConsciousnessStream) {
	// Update salience landscape
	stream.perception.SalienceLandscape["current_focus"] = 0.8
	stream.perception.Affordances = append(stream.perception.Affordances, fmt.Sprintf("affordance_%d", stream.stepCount))
	
	stream.thoughtsGenerated++
}

// processAction processes action for a stream
func (loop *ThreeStreamCognitiveLoop) processAction(stream *ConsciousnessStream) {
	// Perform action
	stream.action.CurrentAction = fmt.Sprintf("action_%d", stream.stepCount)
	stream.action.PastPerformance = append(stream.action.PastPerformance, stream.action.CurrentAction)
	
	// Keep history bounded
	if len(stream.action.PastPerformance) > 100 {
		stream.action.PastPerformance = stream.action.PastPerformance[1:]
	}
	
	stream.actionsPerformed++
}

// processReflection processes reflection for a stream
func (loop *ThreeStreamCognitiveLoop) processReflection(stream *ConsciousnessStream) {
	// Reflect on past actions
	stream.thoughtsGenerated++
}

// processRelevanceRealization processes relevance realization (pivotal step)
func (loop *ThreeStreamCognitiveLoop) processRelevanceRealization(stream *ConsciousnessStream) {
	// Orient present commitment
	stream.perception.RelevanceRealized = 0.9
	stream.perception.PresentCommitment = fmt.Sprintf("commitment_%d", stream.stepCount)
}

// processSimulation processes virtual simulation for a stream
func (loop *ThreeStreamCognitiveLoop) processSimulation(stream *ConsciousnessStream) {
	// Run virtual simulation
	scenario := fmt.Sprintf("scenario_%d", stream.stepCount)
	stream.simulation.VirtualScenarios = append(stream.simulation.VirtualScenarios, scenario)
	stream.simulation.AnticipatedOutcomes[scenario] = 0.7
	
	// Keep scenarios bounded
	if len(stream.simulation.VirtualScenarios) > 100 {
		stream.simulation.VirtualScenarios = stream.simulation.VirtualScenarios[1:]
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
