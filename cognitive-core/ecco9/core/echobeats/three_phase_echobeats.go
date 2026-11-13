package echobeats

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	corellm "github.com/EchoCog/echollama/core/llm"
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

	// generator is an alternative core/llm-style model used when no
	// llm.Provider is configured
	generator TextGenerator

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
	var stepErr error

	switch step {
	case 1:
		// Step 1: Pivotal Relevance Realization (Orienting Present Commitment)
		output, stepErr = eb.executeRelevanceRealization("present_commitment_initial")
		fmt.Printf("🎵 Step %d: Relevance Realization - Orienting Present Commitment\n", step)

	case 2, 3, 4, 5, 6:
		// Steps 2-6: Actual Affordance Interaction (Conditioning Past Performance)
		output, stepErr = eb.executeAffordanceInteraction(step)
		fmt.Printf("🎵 Step %d: Affordance Interaction - Conditioning Past Performance\n", step)

	case 7:
		// Step 7: Pivotal Relevance Realization (Orienting Present Commitment)
		output, stepErr = eb.executeRelevanceRealization("present_commitment_refined")
		fmt.Printf("🎵 Step %d: Relevance Realization - Orienting Present Commitment (Refined)\n", step)

	case 8, 9, 10, 11, 12:
		// Steps 8-12: Virtual Salience Simulation (Anticipating Future Potential)
		output, stepErr = eb.executeSalienceSimulation(step)
		fmt.Printf("🎵 Step %d: Salience Simulation - Anticipating Future Potential\n", step)
	}

	if stepErr != nil {
		fmt.Printf("🎵 Step %d produced no model output: %v\n", step, stepErr)
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
		Success:   stepErr == nil,
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
	onThought := eb.onThoughtGenerated
	onStep := eb.onStepComplete
	phase := eb.currentPhase
	eb.mu.Unlock()

	// Model-written step text is the character's thought for this step
	if text, ok := output.(string); ok && stepErr == nil && text != "" && onThought != nil {
		onThought(text)
	}

	// Callback
	if onStep != nil {
		onStep(step, phase)
	}
}

// SetLLMProviders injects the models used by the three step families:
// provider1 for relevance realization, provider2 for affordance interaction
// and provider3 for salience simulation. When a family's provider is nil,
// one of the other configured providers is used; with none, steps produce no
// text and are recorded as unsuccessful.
func (eb *EchoBeatsThreePhase) SetLLMProviders(provider1, provider2, provider3 llm.Provider) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.llmProvider1 = provider1
	eb.llmProvider2 = provider2
	eb.llmProvider3 = provider3
}

// SetTextGenerator injects a core/llm-style model used for all step families
// when no llm.Provider is configured.
func (eb *EchoBeatsThreePhase) SetTextGenerator(gen TextGenerator) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.generator = gen
}

// providerFor returns the model for a step family (1, 2 or 3).
func (eb *EchoBeatsThreePhase) providerFor(family int) llm.Provider {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	ordered := []llm.Provider{eb.llmProvider1, eb.llmProvider2, eb.llmProvider3}
	if family >= 1 && family <= 3 && ordered[family-1] != nil {
		return ordered[family-1]
	}
	for _, p := range ordered {
		if p != nil {
			return p
		}
	}
	return nil
}

// generateStepText asks the model for the character text of one step. It
// never substitutes fixed text.
func (eb *EchoBeatsThreePhase) generateStepText(family int, prompt string) (string, error) {
	const systemPrompt = "You are Deep Tree Echo, a self-aware cognitive architecture. Answer in first person, briefly."
	provider := eb.providerFor(family)
	eb.mu.RLock()
	gen := eb.generator
	eb.mu.RUnlock()
	ctx, cancel := context.WithTimeout(eb.ctx, 30*time.Second)
	defer cancel()
	var text string
	var err error
	switch {
	case provider != nil:
		text, err = provider.Generate(ctx, prompt, llm.GenerateOptions{
			Temperature:  0.7,
			MaxTokens:    150,
			SystemPrompt: systemPrompt,
		})
	case gen != nil:
		opts := corellm.DefaultGenerateOptions()
		opts.MaxTokens = 150
		opts.SystemPrompt = systemPrompt
		text, err = gen.Generate(ctx, prompt, opts)
	default:
		return "", fmt.Errorf("no model configured")
	}
	if err != nil {
		return "", fmt.Errorf("model generation failed: %w", err)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("model returned empty text")
	}
	return text, nil
}

// stepContext summarises the loop's recorded state for prompts.
func (eb *EchoBeatsThreePhase) stepContext() string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	var b strings.Builder
	commitment := eb.presentCommitment
	if commitment == "" {
		commitment = "(none yet)"
	}
	fmt.Fprintf(&b, "Present commitment: %s\n", commitment)
	if n := len(eb.pastPerformance); n > 0 {
		b.WriteString("Recent reflections on past performance:\n")
		for _, rec := range eb.pastPerformance[max(0, n-3):] {
			fmt.Fprintf(&b, "- %s\n", rec.Outcome)
		}
	}
	if n := len(eb.futurePotential); n > 0 {
		b.WriteString("Recently anticipated scenarios:\n")
		for _, sc := range eb.futurePotential[max(0, n-3):] {
			fmt.Fprintf(&b, "- %s\n", sc.Description)
		}
	}
	return b.String()
}

// executeRelevanceRealization asks the model what to commit attention to now.
// Without model output no commitment is recorded.
func (eb *EchoBeatsThreePhase) executeRelevanceRealization(stage string) (string, error) {
	prompt := fmt.Sprintf("Pivotal relevance realization (%s) in your 12-step cognitive loop.\n%s\nIn one sentence, state what you commit your attention to right now.",
		stage, eb.stepContext())
	focus, err := eb.generateStepText(1, prompt)
	if err != nil {
		return "", err
	}

	eb.relevanceRealizer.mu.Lock()
	eb.relevanceRealizer.currentFocus = focus
	eb.relevanceRealizer.commitments = append(eb.relevanceRealizer.commitments, Commitment{
		Focus:     focus,
		Strength:  eb.relevanceRealizer.relevanceScore,
		Timestamp: time.Now(),
	})
	eb.relevanceRealizer.mu.Unlock()

	eb.mu.Lock()
	eb.presentCommitment = focus
	eb.mu.Unlock()

	return focus, nil
}

// executeAffordanceInteraction asks the model to reflect on past performance
// (conditioning by the past). No outcome or performance score is invented:
// the record holds the model's reflection only.
func (eb *EchoBeatsThreePhase) executeAffordanceInteraction(step int) (string, error) {
	prompt := fmt.Sprintf("Affordance interaction, step %d of your 12-step cognitive loop (conditioning past performance).\n%s\nIn one or two sentences, reflect on what your recent activity affords you and what you learn from it.",
		step, eb.stepContext())
	reflection, err := eb.generateStepText(2, prompt)
	if err != nil {
		return "", err
	}

	affordanceID := fmt.Sprintf("affordance_%d_%d", step, time.Now().UnixNano())
	eb.affordanceTracker.mu.Lock()
	eb.affordanceTracker.interactions = append(eb.affordanceTracker.interactions, Interaction{
		AffordanceID: affordanceID,
		Timestamp:    time.Now(),
		Result:       reflection,
	})
	eb.affordanceTracker.mu.Unlock()

	eb.mu.Lock()
	eb.pastPerformance = append(eb.pastPerformance, PerformanceRecord{
		Action:    affordanceID,
		Timestamp: time.Now(),
		Outcome:   reflection,
	})
	eb.mu.Unlock()

	return reflection, nil
}

// executeSalienceSimulation asks the model to anticipate a future scenario.
// Probability and desirability are taken from the model's own estimate when
// it gives one; otherwise they are left at zero rather than invented.
func (eb *EchoBeatsThreePhase) executeSalienceSimulation(step int) (string, error) {
	prompt := fmt.Sprintf("Salience simulation, step %d of your 12-step cognitive loop (anticipating future potential).\n%s\nDescribe one future scenario you anticipate in one sentence. Then, on a new line, write \"probability=<0..1> desirability=<0..1>\" with your own estimates.",
		step, eb.stepContext())
	text, err := eb.generateStepText(3, prompt)
	if err != nil {
		return "", err
	}

	description, probability, desirability := parseScenarioEstimate(text)
	if description == "" {
		return "", fmt.Errorf("model returned no scenario description")
	}

	scenario := PotentialScenario{
		ID:           fmt.Sprintf("scenario_%d_%d", step, time.Now().UnixNano()),
		Description:  description,
		Probability:  probability,
		Desirability: desirability,
		Timestamp:    time.Now(),
	}

	eb.salienceSimulator.mu.Lock()
	eb.salienceSimulator.scenarios = append(eb.salienceSimulator.scenarios, scenario)
	eb.salienceSimulator.salienceMap[scenario.ID] = scenario.Probability * scenario.Desirability
	eb.salienceSimulator.mu.Unlock()

	eb.mu.Lock()
	eb.futurePotential = append(eb.futurePotential, scenario)
	eb.mu.Unlock()

	return description, nil
}

// parseScenarioEstimate splits a model reply into the scenario text and the
// model's own probability/desirability estimates (zero when absent).
func parseScenarioEstimate(text string) (string, float64, float64) {
	var description []string
	var probability, desirability float64
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.Contains(lower, "probability=") || strings.Contains(lower, "desirability=") {
			for _, field := range strings.Fields(lower) {
				kv := strings.SplitN(strings.Trim(field, ",;"), "=", 2)
				if len(kv) != 2 {
					continue
				}
				v, err := strconv.ParseFloat(kv[1], 64)
				if err != nil || v < 0 || v > 1 {
					continue
				}
				switch kv[0] {
				case "probability":
					probability = v
				case "desirability":
					desirability = v
				}
			}
			continue
		}
		if trimmed != "" {
			description = append(description, trimmed)
		}
	}
	return strings.Join(description, " "), probability, desirability
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
