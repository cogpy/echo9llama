package echobeats

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CognitiveLoop implements the 12-step cognitive processing loop
// Based on Kawaii Hexapod System 4 architecture
// 7 expressive mode steps + 5 reflective mode steps
type CognitiveLoop struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Loop state
	currentStep int
	cycleCount  uint64
	stepHistory []StepExecution
	maxHistory  int

	// Step processors
	stepProcessors map[int]StepProcessor

	// Cognitive state
	currentState *CognitiveState
	stateHistory []*CognitiveState

	// Timing
	stepDuration   time.Duration
	cycleStartTime time.Time

	// Callbacks
	onStepComplete  func(step int, result *StepResult)
	onCycleComplete func(cycle uint64)

	// Metrics
	totalSteps  uint64
	avgStepTime time.Duration

	// Control
	running bool
	paused  bool
}

// CognitiveState represents the current cognitive state
type CognitiveState struct {
	Timestamp       time.Time              `json:"timestamp"`
	CycleNumber     uint64                 `json:"cycle_number"`
	StepNumber      int                    `json:"step_number"`
	Mode            CognitiveMode          `json:"mode"`
	Attention       []string               `json:"attention"`
	WorkingMemory   map[string]interface{} `json:"working_memory"`
	EmotionalTone   map[string]float64     `json:"emotional_tone"`
	CognitiveLoad   float64                `json:"cognitive_load"`
	RelevanceScores map[string]float64     `json:"relevance_scores"`
	ActiveGoals     []string               `json:"active_goals"`
	PendingActions  []string               `json:"pending_actions"`
	Insights        []string               `json:"insights"`
}

// CognitiveMode represents the processing mode
type CognitiveMode string

const (
	ModeExpressive           CognitiveMode = "expressive"
	ModeReflective           CognitiveMode = "reflective"
	ModeRelevanceRealization CognitiveMode = "relevance_realization"
	ModeMetaCognitive        CognitiveMode = "metacognitive"
)

// StepExecution is now defined in shared_types.go to avoid redeclaration

// StepResult contains the result of step processing
type StepResult struct {
	Success        bool
	Output         interface{}
	StateUpdates   map[string]interface{}
	NextStepHint   int
	RelevanceShift float64
	CognitiveLoad  float64
	Insights       []string
	Error          error
}

// StepProcessor defines the interface for step processing
type StepProcessor interface {
	Process(ctx context.Context, state *CognitiveState) (*StepResult, error)
	GetMode() CognitiveMode
	GetDescription() string
}

// NewCognitiveLoop creates a new 12-step cognitive loop
func NewCognitiveLoop() *CognitiveLoop {
	ctx, cancel := context.WithCancel(context.Background())

	cl := &CognitiveLoop{
		ctx:            ctx,
		cancel:         cancel,
		currentStep:    1,
		cycleCount:     0,
		stepHistory:    make([]StepExecution, 0),
		maxHistory:     1000,
		stepProcessors: make(map[int]StepProcessor),
		stateHistory:   make([]*CognitiveState, 0),
		stepDuration:   2 * time.Second,
	}

	// Initialize cognitive state
	cl.currentState = &CognitiveState{
		Timestamp:       time.Now(),
		CycleNumber:     0,
		StepNumber:      1,
		Mode:            ModeExpressive,
		Attention:       make([]string, 0),
		WorkingMemory:   make(map[string]interface{}),
		EmotionalTone:   make(map[string]float64),
		CognitiveLoad:   0.0,
		RelevanceScores: make(map[string]float64),
		ActiveGoals:     make([]string, 0),
		PendingActions:  make([]string, 0),
		Insights:        make([]string, 0),
	}

	// Register default step processors
	cl.registerDefaultProcessors()

	return cl
}

// registerDefaultProcessors registers the 12 step processors
func (cl *CognitiveLoop) registerDefaultProcessors() {
	// Steps 1-4: Expressive Mode (Actual Affordance Interaction)
	cl.stepProcessors[1] = &PerceptionProcessor{}
	cl.stepProcessors[2] = &MemoryActivationProcessor{}
	cl.stepProcessors[3] = &ActionGenerationProcessor{}
	cl.stepProcessors[4] = &ActionExecutionProcessor{}

	// Step 5: Pivotal Relevance Realization
	cl.stepProcessors[5] = &RelevanceRealizationProcessor{phase: "present_commitment"}

	// Steps 6-10: Reflective Mode (Virtual Salience Simulation)
	cl.stepProcessors[6] = &ScenarioSimulationProcessor{}
	cl.stepProcessors[7] = &OutcomeEvaluationProcessor{}
	cl.stepProcessors[8] = &ModelUpdateProcessor{}
	cl.stepProcessors[9] = &LearningConsolidationProcessor{}
	cl.stepProcessors[10] = &InsightGenerationProcessor{}

	// Step 11: Pivotal Relevance Realization
	cl.stepProcessors[11] = &RelevanceRealizationProcessor{phase: "future_commitment"}

	// Step 12: Meta-Cognitive Reflection
	cl.stepProcessors[12] = &MetaCognitiveProcessor{}
}

// RegisterStepProcessor registers a custom step processor
func (cl *CognitiveLoop) RegisterStepProcessor(step int, processor StepProcessor) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.stepProcessors[step] = processor
}

// Start begins the cognitive loop
func (cl *CognitiveLoop) Start() error {
	cl.mu.Lock()
	if cl.running {
		cl.mu.Unlock()
		return fmt.Errorf("cognitive loop already running")
	}
	cl.running = true
	cl.cycleStartTime = time.Now()
	cl.mu.Unlock()

	fmt.Println("🔄 CognitiveLoop: Starting 12-step cognitive processing...")
	fmt.Printf("   Step Duration: %v\n", cl.stepDuration)
	fmt.Println("   Mode Sequence: Expressive(1-4) → Relevance(5) → Reflective(6-10) → Relevance(11) → MetaCognitive(12)")

	go cl.run()

	return nil
}

// Stop gracefully stops the cognitive loop
func (cl *CognitiveLoop) Stop() error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if !cl.running {
		return fmt.Errorf("cognitive loop not running")
	}

	fmt.Println("🔄 CognitiveLoop: Stopping...")
	cl.running = false
	cl.cancel()

	return nil
}

// Pause pauses the cognitive loop
func (cl *CognitiveLoop) Pause() {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.paused = true
	fmt.Println("⏸️  CognitiveLoop: Paused")
}

// Resume resumes the cognitive loop
func (cl *CognitiveLoop) Resume() {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.paused = false
	fmt.Println("▶️  CognitiveLoop: Resumed")
}

// run executes the main cognitive loop
func (cl *CognitiveLoop) run() {
	ticker := time.NewTicker(cl.stepDuration)
	defer ticker.Stop()

	for {
		select {
		case <-cl.ctx.Done():
			return
		case <-ticker.C:
			cl.mu.RLock()
			isPaused := cl.paused
			cl.mu.RUnlock()

			if !isPaused {
				cl.executeStep()
			}
		}
	}
}

// executeStep executes the current step
func (cl *CognitiveLoop) executeStep() {
	cl.mu.Lock()
	step := cl.currentStep
	processor := cl.stepProcessors[step]
	state := cl.currentState
	cl.mu.Unlock()

	if processor == nil {
		fmt.Printf("⚠️  CognitiveLoop: No processor for step %d\n", step)
		cl.advanceStep()
		return
	}

	startTime := time.Now()

	// Update state mode
	state.StepNumber = step
	state.Mode = processor.GetMode()
	state.Timestamp = startTime

	// Execute step
	result, err := processor.Process(cl.ctx, state)

	duration := time.Since(startTime)

	// Record execution
	execution := StepExecution{
		StepNumber: step,
		StartTime:  startTime,
		Duration:   duration,
		Mode:       processor.GetMode(),
		Success:    err == nil && result != nil && result.Success,
		Output:     nil,
		Error:      err,
	}

	if result != nil {
		execution.Output = result.Output

		// Apply state updates
		cl.applyStateUpdates(result.StateUpdates)

		// Update cognitive load
		state.CognitiveLoad = result.CognitiveLoad

		// Add insights
		if len(result.Insights) > 0 {
			state.Insights = append(state.Insights, result.Insights...)
		}
	}

	cl.mu.Lock()
	cl.stepHistory = append(cl.stepHistory, execution)
	if len(cl.stepHistory) > cl.maxHistory {
		cl.stepHistory = cl.stepHistory[len(cl.stepHistory)-cl.maxHistory:]
	}
	cl.totalSteps++
	cl.mu.Unlock()

	// Callback
	if cl.onStepComplete != nil {
		cl.onStepComplete(step, result)
	}

	// Log step completion
	modeEmoji := cl.getModeEmoji(processor.GetMode())
	fmt.Printf("%s Step %2d/%2d: %s (%.2fs)\n",
		modeEmoji, step, 12, processor.GetDescription(), duration.Seconds())

	// Advance to next step
	cl.advanceStep()
}

// advanceStep moves to the next step
func (cl *CognitiveLoop) advanceStep() {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.currentStep++

	if cl.currentStep > 12 {
		// Cycle complete
		cl.currentStep = 1
		cl.cycleCount++

		cycleDuration := time.Since(cl.cycleStartTime)
		cl.cycleStartTime = time.Now()

		fmt.Printf("\n🔄 Cycle %d complete (duration: %s)\n", cl.cycleCount, cycleDuration)
		fmt.Printf("   Insights generated: %d\n", len(cl.currentState.Insights))
		fmt.Printf("   Cognitive load: %.2f\n\n", cl.currentState.CognitiveLoad)

		// Save state snapshot
		stateCopy := *cl.currentState
		cl.stateHistory = append(cl.stateHistory, &stateCopy)

		// Reset cycle-specific state
		cl.currentState.Insights = make([]string, 0)
		cl.currentState.CycleNumber = cl.cycleCount

		// Callback
		if cl.onCycleComplete != nil {
			cl.onCycleComplete(cl.cycleCount)
		}
	}
}

// applyStateUpdates applies updates to cognitive state
func (cl *CognitiveLoop) applyStateUpdates(updates map[string]interface{}) {
	if updates == nil {
		return
	}

	for key, value := range updates {
		cl.currentState.WorkingMemory[key] = value
	}
}

// getModeEmoji returns emoji for cognitive mode
func (cl *CognitiveLoop) getModeEmoji(mode CognitiveMode) string {
	switch mode {
	case ModeExpressive:
		return "🎭"
	case ModeReflective:
		return "🤔"
	case ModeRelevanceRealization:
		return "🎯"
	case ModeMetaCognitive:
		return "🧠"
	default:
		return "⚙️"
	}
}

// GetCurrentState returns the current cognitive state
func (cl *CognitiveLoop) GetCurrentState() *CognitiveState {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	stateCopy := *cl.currentState
	return &stateCopy
}

// GetMetrics returns cognitive loop metrics
func (cl *CognitiveLoop) GetMetrics() map[string]interface{} {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	return map[string]interface{}{
		"current_step":   cl.currentStep,
		"cycle_count":    cl.cycleCount,
		"total_steps":    cl.totalSteps,
		"current_mode":   cl.currentState.Mode,
		"cognitive_load": cl.currentState.CognitiveLoad,
		"insights_count": len(cl.currentState.Insights),
		"running":        cl.running,
		"paused":         cl.paused,
	}
}

// SetStepDuration sets the duration for each step
func (cl *CognitiveLoop) SetStepDuration(duration time.Duration) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.stepDuration = duration
}

// SetCallbacks sets the callback functions
func (cl *CognitiveLoop) SetCallbacks(
	onStepComplete func(step int, result *StepResult),
	onCycleComplete func(cycle uint64),
) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.onStepComplete = onStepComplete
	cl.onCycleComplete = onCycleComplete
}
