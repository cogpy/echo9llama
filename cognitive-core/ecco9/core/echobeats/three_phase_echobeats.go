package echobeats

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EchoCog/echollama/llm"
)

// EchoBeatsThreePhase implements the 12-step 3-phase cognitive loop
// with 3 concurrent inference engines as specified in the architecture
type EchoBeatsThreePhase struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Three concurrent simple inference engines
	engine1 *SimpleInferenceEngine
	engine2 *SimpleInferenceEngine
	engine3 *SimpleInferenceEngine

	// LLM providers for actual inference
	llmProvider1 llm.Provider
	llmProvider2 llm.Provider
	llmProvider3 llm.Provider

	// 12-step cognitive loop state
	currentStep  int
	currentPhase CognitivePhaseEnum
	stepHistory  []StepExecution

	// Phase-specific processors
	expressiveProcessor *ExpressiveProcessor
	reflectiveProcessor *ReflectiveProcessor

	// Relevance realization
	relevanceRealizer *RelevanceRealizer

	// Affordance and salience
	affordanceTracker *AffordanceTracker
	salienceSimulator *SalienceSimulator

	// Cognitive state
	presentCommitment string
	pastPerformance   []PerformanceRecord
	futurePotential   []PotentialScenario

	// Callbacks
	onThoughtGenerated func(thought string)
	onStepComplete     func(step int, phase CognitivePhaseEnum)

	// Metrics
	cyclesCompleted uint64
	stepsExecuted   uint64

	// Running state
	running bool
}

// CognitivePhaseEnum represents the three phases of the loop
// Renamed from CognitivePhase to avoid conflict with struct in threephase.go
type CognitivePhaseEnum int

const (
	PhaseExpressive CognitivePhaseEnum = iota // Steps 1-7
	PhaseReflective                           // Steps 8-12
	PhaseTransition                           // Between phases
)

func (cp CognitivePhaseEnum) String() string {
	return [...]string{"Expressive", "Reflective", "Transition"}[cp]
}

// StepExecution and StepType are now defined in shared_types.go to avoid redeclaration

// SimpleInferenceEngine is a simplified inference engine for three-phase processing
// This is different from the full InferenceEngine in inference_engine.go
type SimpleInferenceEngine struct {
	mu             sync.RWMutex
	id             int
	state          EngineState
	currentTask    *SimpleInferenceTask
	completedTasks uint64
	processingTime time.Duration
}

// EngineState represents engine state
type EngineState int

const (
	EngineIdle EngineState = iota
	EngineProcessing
	EngineWaiting
)

// SimpleInferenceTask represents a simple task for the three-phase engine
// This is different from the full InferenceTask in inference_engine.go
type SimpleInferenceTask struct {
	ID        string
	Type      StepType
	Input     interface{}
	Output    interface{}
	StartTime time.Time
	EndTime   time.Time
}

// ExpressiveProcessor handles expressive mode steps (1-7)
type ExpressiveProcessor struct {
	mu             sync.RWMutex
	activeActions  []string
	performanceLog []PerformanceRecord
}

// ReflectiveProcessor handles reflective mode steps (8-12)
type ReflectiveProcessor struct {
	mu          sync.RWMutex
	simulations []PotentialScenario
	evaluations []ScenarioEvaluation
}

// RelevanceRealizer handles pivotal relevance realization (steps 1, 7)
type RelevanceRealizer struct {
	mu             sync.RWMutex
	currentFocus   string
	relevanceScore float64
	commitments    []Commitment
}

// AffordanceTracker tracks actual affordance interactions (steps 2-6)
type AffordanceTracker struct {
	mu           sync.RWMutex
	affordances  []ThreePhaseAffordance
	interactions []Interaction
}

// SalienceSimulator handles virtual salience simulation (steps 8-12)
type SalienceSimulator struct {
	mu          sync.RWMutex
	scenarios   []PotentialScenario
	salienceMap map[string]float64
}

// Supporting types

type PerformanceRecord struct {
	Action    string
	Timestamp time.Time
	Outcome   string
	Quality   float64
}

type PotentialScenario struct {
	ID           string
	Description  string
	Probability  float64
	Desirability float64
	Timestamp    time.Time
}

type ScenarioEvaluation struct {
	ScenarioID string
	Score      float64
	Reasoning  string
}

type Commitment struct {
	Focus     string
	Strength  float64
	Timestamp time.Time
}

type ThreePhaseAffordance struct {
	ID        string
	Type      string
	Available bool
	Quality   float64
}

type Interaction struct {
	AffordanceID string
	Timestamp    time.Time
	Result       string
	Performance  float64
}

// NewEchoBeatsThreePhase creates a new 3-phase EchoBeats system
func NewEchoBeatsThreePhase() *EchoBeatsThreePhase {
	return NewEchoBeatsThreePhaseWithProviders(nil, nil, nil)
}

// NewEchoBeatsThreePhaseWithProviders creates a new 3-phase EchoBeats system with LLM providers
func NewEchoBeatsThreePhaseWithProviders(provider1, provider2, provider3 llm.Provider) *EchoBeatsThreePhase {
	ctx, cancel := context.WithCancel(context.Background())

	eb := &EchoBeatsThreePhase{
		ctx:                 ctx,
		cancel:              cancel,
		engine1:             newSimpleInferenceEngine(1),
		engine2:             newSimpleInferenceEngine(2),
		engine3:             newSimpleInferenceEngine(3),
		llmProvider1:        provider1,
		llmProvider2:        provider2,
		llmProvider3:        provider3,
		currentStep:         1,
		currentPhase:        PhaseExpressive,
		stepHistory:         make([]StepExecution, 0),
		expressiveProcessor: newExpressiveProcessor(),
		reflectiveProcessor: newReflectiveProcessor(),
		relevanceRealizer:   newRelevanceRealizer(),
		affordanceTracker:   newAffordanceTracker(),
		salienceSimulator:   newSalienceSimulator(),
		pastPerformance:     make([]PerformanceRecord, 0),
		futurePotential:     make([]PotentialScenario, 0),
	}

	return eb
}

func newSimpleInferenceEngine(id int) *SimpleInferenceEngine {
	return &SimpleInferenceEngine{
		id:    id,
		state: EngineIdle,
	}
}

func newExpressiveProcessor() *ExpressiveProcessor {
	return &ExpressiveProcessor{
		activeActions:  make([]string, 0),
		performanceLog: make([]PerformanceRecord, 0),
	}
}

func newReflectiveProcessor() *ReflectiveProcessor {
	return &ReflectiveProcessor{
		simulations: make([]PotentialScenario, 0),
		evaluations: make([]ScenarioEvaluation, 0),
	}
}

func newRelevanceRealizer() *RelevanceRealizer {
	return &RelevanceRealizer{
		commitments: make([]Commitment, 0),
	}
}

func newAffordanceTracker() *AffordanceTracker {
	return &AffordanceTracker{
		affordances:  make([]ThreePhaseAffordance, 0),
		interactions: make([]Interaction, 0),
	}
}

func newSalienceSimulator() *SalienceSimulator {
	return &SalienceSimulator{
		scenarios:   make([]PotentialScenario, 0),
		salienceMap: make(map[string]float64),
	}
}

// Start begins the 12-step cognitive loop
func (eb *EchoBeatsThreePhase) Start() error {
	eb.mu.Lock()
	if eb.running {
		eb.mu.Unlock()
		return fmt.Errorf("already running")
	}
	eb.running = true
	eb.mu.Unlock()

	fmt.Println("🎵 ═══════════════════════════════════════════════════════")
	fmt.Println("🎵 EchoBeats Three-Phase: 12-Step Cognitive Loop Starting")
	fmt.Println("🎵 ═══════════════════════════════════════════════════════")
	fmt.Println("🎵 Architecture:")
	fmt.Println("🎵   - 3 Concurrent Inference Engines")
	fmt.Println("🎵   - 12-Step Loop (7 Expressive + 5 Reflective)")
	fmt.Println("🎵   - Phase 1: Steps 1-7 (Expressive Mode)")
	fmt.Println("🎵   - Phase 2: Steps 8-12 (Reflective Mode)")
	fmt.Println("🎵 ═══════════════════════════════════════════════════════")
	fmt.Println()

	// Start the three concurrent engines
	go eb.runEngine(eb.engine1)
	go eb.runEngine(eb.engine2)
	go eb.runEngine(eb.engine3)

	// Start the main cognitive loop
	go eb.cognitiveLoop()

	return nil
}

// Stop gracefully stops the system
func (eb *EchoBeatsThreePhase) Stop() error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if !eb.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("\n🎵 Stopping EchoBeats Three-Phase...")
	eb.running = false
	eb.cancel()

	return nil
}

// cognitiveLoop executes the 12-step loop
func (eb *EchoBeatsThreePhase) cognitiveLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-eb.ctx.Done():
			return
		case <-ticker.C:
			eb.executeNextStep()
		}
	}
}

// executeNextStep executes the next step in the 12-step loop
func (eb *EchoBeatsThreePhase) executeNextStep() {
	eb.mu.Lock()
	step := eb.currentStep
	eb.mu.Unlock()

	startTime := time.Now()

	// Execute step and capture output
	var output interface{}

	switch step {
	case 1:
		// Step 1: Pivotal Relevance Realization (Orienting Present Commitment)
		output = eb.executeRelevanceRealization("present_commitment_initial")
		fmt.Printf("🎵 Step %d: Relevance Realization - Orienting Present Commitment\n", step)

	case 2, 3, 4, 5, 6:
		// Steps 2-6: Actual Affordance Interaction (Conditioning Past Performance)
		output = eb.executeAffordanceInteraction(step)
		fmt.Printf("🎵 Step %d: Affordance Interaction - Conditioning Past Performance\n", step)

	case 7:
		// Step 7: Pivotal Relevance Realization (Orienting Present Commitment)
		output = eb.executeRelevanceRealization("present_commitment_refined")
		fmt.Printf("🎵 Step %d: Relevance Realization - Orienting Present Commitment (Refined)\n", step)

	case 8, 9, 10, 11, 12:
		// Steps 8-12: Virtual Salience Simulation (Anticipating Future Potential)
		output = eb.executeSalienceSimulation(step)
		fmt.Printf("🎵 Step %d: Salience Simulation - Anticipating Future Potential\n", step)
	}

	duration := time.Since(startTime)

	// Record step execution
	execution := StepExecution{
		StepNumber: step,
		// PhaseType omitted - using local CognitivePhase tracking instead
		Timestamp: startTime,
		StartTime: startTime,
		Duration:  duration,
		Output:    output,
		Success:   true, // Assume success if no error
	}

	eb.mu.Lock()
	eb.stepHistory = append(eb.stepHistory, execution)
	eb.stepsExecuted++

	// Advance to next step
	eb.currentStep++
	if eb.currentStep > 12 {
		eb.currentStep = 1
		eb.cyclesCompleted++
		fmt.Printf("\n🎵 ═══ Cycle %d Complete ═══\n\n", eb.cyclesCompleted)
	}

	// Update phase
	eb.currentPhase = eb.determinePhase(eb.currentStep)
	eb.mu.Unlock()

	// Callback
	if eb.onStepComplete != nil {
		eb.onStepComplete(step, eb.currentPhase)
	}
}

// executeRelevanceRealization executes relevance realization step
func (eb *EchoBeatsThreePhase) executeRelevanceRealization(context string) string {
	eb.relevanceRealizer.mu.Lock()
	defer eb.relevanceRealizer.mu.Unlock()

	// Determine current focus and commitment
	focus := fmt.Sprintf("Focus_%s_%d", context, time.Now().Unix())
	eb.relevanceRealizer.currentFocus = focus
	eb.relevanceRealizer.relevanceScore = 0.8

	commitment := Commitment{
		Focus:     focus,
		Strength:  0.8,
		Timestamp: time.Now(),
	}
	eb.relevanceRealizer.commitments = append(eb.relevanceRealizer.commitments, commitment)

	eb.mu.Lock()
	eb.presentCommitment = focus
	eb.mu.Unlock()

	return focus
}

// executeAffordanceInteraction executes affordance interaction step
func (eb *EchoBeatsThreePhase) executeAffordanceInteraction(step int) string {
	eb.affordanceTracker.mu.Lock()
	defer eb.affordanceTracker.mu.Unlock()

	// Simulate affordance interaction
	affordanceID := fmt.Sprintf("affordance_%d_%d", step, time.Now().Unix())

	interaction := Interaction{
		AffordanceID: affordanceID,
		Timestamp:    time.Now(),
		Result:       "success",
		Performance:  0.7 + float64(step)*0.05,
	}
	eb.affordanceTracker.interactions = append(eb.affordanceTracker.interactions, interaction)

	// Record performance
	performance := PerformanceRecord{
		Action:    fmt.Sprintf("Action_Step_%d", step),
		Timestamp: time.Now(),
		Outcome:   "completed",
		Quality:   interaction.Performance,
	}

	eb.mu.Lock()
	eb.pastPerformance = append(eb.pastPerformance, performance)
	eb.mu.Unlock()

	return affordanceID
}

// executeSalienceSimulation executes salience simulation step
func (eb *EchoBeatsThreePhase) executeSalienceSimulation(step int) string {
	eb.salienceSimulator.mu.Lock()
	defer eb.salienceSimulator.mu.Unlock()

	// Simulate future scenario
	scenarioID := fmt.Sprintf("scenario_%d_%d", step, time.Now().Unix())

	scenario := PotentialScenario{
		ID:           scenarioID,
		Description:  fmt.Sprintf("Future scenario from step %d", step),
		Probability:  0.6 + float64(step-8)*0.08,
		Desirability: 0.7 + float64(step-8)*0.05,
		Timestamp:    time.Now(),
	}

	eb.salienceSimulator.scenarios = append(eb.salienceSimulator.scenarios, scenario)
	eb.salienceSimulator.salienceMap[scenarioID] = scenario.Probability * scenario.Desirability

	eb.mu.Lock()
	eb.futurePotential = append(eb.futurePotential, scenario)
	eb.mu.Unlock()

	return scenarioID
}

// runEngine runs a simple inference engine
func (eb *EchoBeatsThreePhase) runEngine(engine *SimpleInferenceEngine) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-eb.ctx.Done():
			return
		case <-ticker.C:
			eb.processEngineTask(engine)
		}
	}
}

// processEngineTask processes tasks for a simple engine
func (eb *EchoBeatsThreePhase) processEngineTask(engine *SimpleInferenceEngine) {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	if engine.state == EngineIdle {
		// Assign new task based on current step
		eb.mu.RLock()
		step := eb.currentStep
		eb.mu.RUnlock()

		task := &SimpleInferenceTask{
			ID:        fmt.Sprintf("task_e%d_%d", engine.id, time.Now().UnixNano()),
			Type:      eb.getStepType(step),
			StartTime: time.Now(),
		}

		engine.currentTask = task
		engine.state = EngineProcessing
	} else if engine.state == EngineProcessing {
		// Complete current task
		if engine.currentTask != nil {
			engine.currentTask.EndTime = time.Now()
			engine.processingTime += engine.currentTask.EndTime.Sub(engine.currentTask.StartTime)
			engine.completedTasks++
		}

		engine.currentTask = nil
		engine.state = EngineIdle
	}
}

// getStepType returns the step type for a given step number
func (eb *EchoBeatsThreePhase) getStepType(step int) StepType {
	switch step {
	case 1, 7:
		return StepRelevanceRealization
	case 2, 3, 4, 5, 6:
		return StepAffordanceInteraction
	case 8, 9, 10, 11, 12:
		return StepSalienceSimulation
	default:
		return StepRelevanceRealization
	}
}

// determinePhase determines the phase for a given step
func (eb *EchoBeatsThreePhase) determinePhase(step int) CognitivePhaseEnum {
	if step >= 1 && step <= 7 {
		return PhaseExpressive
	} else if step >= 8 && step <= 12 {
		return PhaseReflective
	}
	return PhaseTransition
}

// SetThoughtCallback sets the thought generation callback
func (eb *EchoBeatsThreePhase) SetThoughtCallback(callback func(thought string)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.onThoughtGenerated = callback
}

// SetStepCompleteCallback sets the step complete callback
func (eb *EchoBeatsThreePhase) SetStepCompleteCallback(callback func(step int, phase CognitivePhaseEnum)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.onStepComplete = callback
}

// GetMetrics returns current metrics
func (eb *EchoBeatsThreePhase) GetMetrics() map[string]interface{} {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	return map[string]interface{}{
		"current_step":     eb.currentStep,
		"current_phase":    eb.currentPhase.String(),
		"cycles_completed": eb.cyclesCompleted,
		"steps_executed":   eb.stepsExecuted,
		"engine1_tasks":    eb.engine1.completedTasks,
		"engine2_tasks":    eb.engine2.completedTasks,
		"engine3_tasks":    eb.engine3.completedTasks,
		"past_performance": len(eb.pastPerformance),
		"future_scenarios": len(eb.futurePotential),
	}
}

// GetCurrentState returns the current cognitive state
func (eb *EchoBeatsThreePhase) GetCurrentState() map[string]interface{} {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	return map[string]interface{}{
		"step":               eb.currentStep,
		"phase":              eb.currentPhase.String(),
		"present_commitment": eb.presentCommitment,
		"past_performance":   len(eb.pastPerformance),
		"future_potential":   len(eb.futurePotential),
	}
}
