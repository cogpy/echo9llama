package deeptreeecho

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/backendcap"
	"github.com/cogpy/echo9llama/core/llm"
)

// EchobeatsScheduler implements the 12-step 3-phase cognitive loop
// with three logical inference engines on a sequential cadence for goal-directed scheduling.
type EchobeatsScheduler struct {
	mu          sync.RWMutex
	stepMu      sync.Mutex      // Serializes pause requests with active steps.
	lifecycleMu sync.Mutex      // Serializes Start/Stop so an old loop cannot overlap a restart.
	ctx         context.Context //nolint:containedctx // Scheduler owns this lifecycle context from Start through Stop.
	cancel      context.CancelFunc
	runDone     chan struct{}

	// Three logical inference engines on a sequential cadence
	engine1 *InferenceEngine
	engine2 *InferenceEngine
	engine3 *InferenceEngine

	// 12-step cognitive loop state
	currentStep  int
	currentPhase CognitivePhase
	cycleCount   uint64

	// LLM provider
	llmProvider llm.LLMProvider

	// Cognitive state
	presentCommitment  string
	pastPerformance    []string
	futureAnticipation []string

	// Goal-directed scheduling
	goalQueue           []ScheduledGoal
	priorityMatrix      map[string]float64
	temporalPlan        []TemporalEvent
	appliedGoalEvidence map[string]struct{}

	// Triad processing (tetrahedral synchronization)
	triadStates   [4]TriadState
	phaseRotation float64

	// Evolution integration hooks
	onCycleComplete     func(metrics CycleMetrics)
	onGoalAchieved      func(goal ScheduledGoal)
	onEmergenceDetected func(pattern string, strength float64)
	onAffordance        func(step int) (summary string, verified bool)
	onReflection        func(step int) error

	// Metrics
	totalSteps      uint64
	totalCycles     uint64
	goalsCompleted  uint64
	emergenceEvents uint64

	// Running state
	running      bool
	paused       bool
	adaptiveMode bool
}

// InferenceEngine represents one of three logical engines on a sequential cadence.
type InferenceEngine struct {
	ID          int
	mu          sync.RWMutex
	currentTask *CognitiveTask
	taskHistory []CognitiveTask
	performance float64
}

// CognitiveTask represents work for an inference engine
type CognitiveTask struct {
	ID               string
	Type             TaskType
	Description      string
	Priority         float64
	StartTime        time.Time
	CompletionTime   *time.Time
	Result           string
	Success          bool
	Provider         string
	BackendKind      backendcap.BackendKind
	ModelID          string
	RouteReason      string
	Degraded         bool
	ProviderAttempts int
}

// TaskType categorizes cognitive tasks
type TaskType int

const (
	TaskRelevanceRealization TaskType = iota
	TaskAffordanceInteraction
	TaskSalienceSimulation
	TaskPatternRecognition
	TaskGoalPursuit
	TaskMemoryIntegration
)

// inferenceEngineTaskHistoryLimit bounds retained task evidence per engine.
// The current task is stored separately and is not part of this retention budget.
const inferenceEngineTaskHistoryLimit = 100

func (tt TaskType) String() string {
	return [...]string{
		"RelevanceRealization",
		"AffordanceInteraction",
		"SalienceSimulation",
		"PatternRecognition",
		"GoalPursuit",
		"MemoryIntegration",
	}[tt]
}

// CognitivePhase represents the three phases of the 12-step loop
type CognitivePhase int

const (
	PhaseExpressive   CognitivePhase = iota // Steps 1-4
	PhaseReflective                         // Steps 5-8
	PhaseAnticipatory                       // Steps 9-12
)

func (cp CognitivePhase) String() string {
	return [...]string{"Expressive", "Reflective", "Anticipatory"}[cp]
}

// ScheduledGoal represents a goal in the scheduling queue
type ScheduledGoal struct {
	ID           string
	Description  string
	Priority     float64
	Deadline     *time.Time
	Progress     float64
	Status       GoalStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Dependencies []string
}

// GoalStatus represents the status of a scheduled goal
type GoalStatus string

const (
	GoalPending   GoalStatus = "pending"
	GoalActive    GoalStatus = "active"
	GoalCompleted GoalStatus = "completed"
	GoalBlocked   GoalStatus = "blocked"
	GoalAbandoned GoalStatus = "abandoned"
)

// TemporalEvent represents a planned event in the temporal timeline
type TemporalEvent struct {
	ID          string
	Type        string
	Description string
	ScheduledAt time.Time
	Duration    time.Duration
	Priority    float64
	Executed    bool
}

// TriadState represents the state of a tetrahedral triad
type TriadState struct {
	ID          int
	Name        string
	Steps       [3]int
	Activation  float64
	Coherence   float64
	LastUpdated time.Time
}

// CycleMetrics captures metrics for a complete cognitive cycle
type CycleMetrics struct {
	CycleNumber       uint64
	Duration          time.Duration
	StepsCompleted    int
	GoalsAdvanced     int
	EmergenceLevel    float64
	EnginePerformance [3]float64
	TriadCoherence    [4]float64
	Timestamp         time.Time
}

// NewEchobeatsScheduler creates a new 12-step scheduler
func NewEchobeatsScheduler(llmProvider llm.LLMProvider) *EchobeatsScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	sched := &EchobeatsScheduler{
		ctx:                 ctx,
		cancel:              cancel,
		llmProvider:         llmProvider,
		engine1:             newInferenceEngine(1),
		engine2:             newInferenceEngine(2),
		engine3:             newInferenceEngine(3),
		currentStep:         1,
		currentPhase:        PhaseExpressive,
		pastPerformance:     make([]string, 0),
		futureAnticipation:  make([]string, 0),
		goalQueue:           make([]ScheduledGoal, 0),
		priorityMatrix:      make(map[string]float64),
		temporalPlan:        make([]TemporalEvent, 0),
		appliedGoalEvidence: make(map[string]struct{}),
		phaseRotation:       0.0,
		adaptiveMode:        true,
	}

	// Initialize tetrahedral triads
	sched.initializeTriads()

	return sched
}

// initializeTriads sets up the four tetrahedral triads
func (sched *EchobeatsScheduler) initializeTriads() {
	now := time.Now()

	sched.triadStates = [4]TriadState{
		{
			ID:          0,
			Name:        "RelevanceRealization",
			Steps:       [3]int{1, 5, 9},
			Activation:  0.5,
			Coherence:   0.5,
			LastUpdated: now,
		},
		{
			ID:          1,
			Name:        "AffordanceInteraction",
			Steps:       [3]int{2, 6, 10},
			Activation:  0.5,
			Coherence:   0.5,
			LastUpdated: now,
		},
		{
			ID:          2,
			Name:        "SalienceSimulation",
			Steps:       [3]int{3, 7, 11},
			Activation:  0.5,
			Coherence:   0.5,
			LastUpdated: now,
		},
		{
			ID:          3,
			Name:        "MetaCognitiveReflection",
			Steps:       [3]int{4, 8, 12},
			Activation:  0.5,
			Coherence:   0.5,
			LastUpdated: now,
		},
	}
}

func newInferenceEngine(id int) *InferenceEngine {
	return &InferenceEngine{
		ID:          id,
		taskHistory: make([]CognitiveTask, 0, inferenceEngineTaskHistoryLimit),
		performance: 0.5,
	}
}

// appendBoundedTaskHistory appends a completed task while retaining only the
// most recent bounded history. Callers must hold the engine mutex.
func appendBoundedTaskHistory(history []CognitiveTask, task CognitiveTask) []CognitiveTask {
	history = append(history, task)
	if overflow := len(history) - inferenceEngineTaskHistoryLimit; overflow > 0 {
		copy(history, history[overflow:])
		clear(history[len(history)-overflow:])
		history = history[:inferenceEngineTaskHistoryLimit]
	}
	return history
}

func (engine *InferenceEngine) performanceSnapshot() float64 {
	engine.mu.RLock()
	defer engine.mu.RUnlock()
	return engine.performance
}

// Start begins the 12-step cognitive loop
func (sched *EchobeatsScheduler) Start() error {
	sched.lifecycleMu.Lock()
	defer sched.lifecycleMu.Unlock()

	sched.mu.Lock()
	if sched.running {
		sched.mu.Unlock()
		return fmt.Errorf("already running")
	}
	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan struct{})
	sched.ctx = ctx
	sched.cancel = cancel
	sched.runDone = runDone
	sched.running = true
	sched.mu.Unlock()

	fmt.Println("🎵 Starting Echobeats 12-Step Cognitive Loop...")
	fmt.Println("   Architecture: three logical inference engines on a sequential cadence")
	fmt.Println("   Phases: Expressive (1-4) → Reflective (5-8) → Anticipatory (9-12)")
	fmt.Println("   Pattern: 7 Expressive + 5 Reflective Steps")

	go sched.run(ctx, runDone)

	return nil
}

// Stop gracefully stops the scheduler
func (sched *EchobeatsScheduler) Stop() error {
	sched.lifecycleMu.Lock()
	defer sched.lifecycleMu.Unlock()

	sched.mu.Lock()
	if !sched.running {
		sched.mu.Unlock()
		return fmt.Errorf("not running")
	}

	fmt.Println("🎵 Stopping echobeats scheduler...")
	sched.running = false
	cancel := sched.cancel
	runDone := sched.runDone
	sched.mu.Unlock()

	// Cancel first so a provider that honors the scheduler context can finish.
	// Then use stepMu as a barrier before waiting for the loop to exit.
	cancel()
	sched.stepMu.Lock()
	//nolint:staticcheck // SA2001: intentional quiescence barrier for in-flight steps.
	sched.stepMu.Unlock()
	if runDone != nil {
		<-runDone
	}

	return nil
}

// run executes the 12-step cognitive loop
func (sched *EchobeatsScheduler) run(ctx context.Context, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sched.executeStep()
		}
	}
}

// executeStep performs one step of the 12-step loop
func (sched *EchobeatsScheduler) executeStep() {
	// Serializing with Pause makes a completed Pause a barrier: after it
	// returns, no steps or provider calls can begin until Resume.
	sched.stepMu.Lock()
	defer sched.stepMu.Unlock()

	sched.mu.RLock()
	if sched.paused || sched.ctx.Err() != nil {
		sched.mu.RUnlock()
		return
	}
	step := sched.currentStep
	phase := sched.currentPhase
	sched.mu.RUnlock()

	fmt.Printf("🎵 Echobeats Step %d/%d [%s Phase]\n", step, 12, phase.String())

	switch step {
	case 1:
		// Pivotal Relevance Realization - Orient Present Commitment
		sched.relevanceRealization("What is most relevant to focus on right now?")

	case 2, 3, 4, 5, 6:
		// Actual Affordance Interaction - Condition Past Performance (5 steps)
		sched.affordanceInteraction(step)

	case 7:
		// Pivotal Relevance Realization - Orient Present Commitment
		sched.relevanceRealization("Given what I've learned, what should I commit to next?")

	case 8:
		// The canonical evidence-grounded metacognition hook runs while this
		// step remains quiescent, but never while sched.mu is held.
		if !sched.runReflection(step) {
			sched.salienceSimulation(step - 7)
		}

	case 9, 10, 11, 12:
		// Virtual Salience Simulation - Anticipate Future Potential (5 steps)
		sched.salienceSimulation(step - 7)
	}

	sched.mu.Lock()
	sched.totalSteps++
	sched.currentStep++

	if sched.currentStep > 12 {
		sched.currentStep = 1
		sched.cycleCount++
		sched.totalCycles++
		fmt.Printf("🎵 ═══ Cycle %d Complete ═══\n\n", sched.cycleCount)
	}

	// Update phase based on step
	if sched.currentStep >= 1 && sched.currentStep <= 4 {
		sched.currentPhase = PhaseExpressive
	} else if sched.currentStep >= 5 && sched.currentStep <= 8 {
		sched.currentPhase = PhaseReflective
	} else {
		sched.currentPhase = PhaseAnticipatory
	}

	sched.mu.Unlock()
}

// runReflection invokes the evidence-grounded metacognition hook for the
// reflective pivot. executeStep calls this with stepMu held, but sched.mu is
// deliberately released before user code runs.
func (sched *EchobeatsScheduler) runReflection(step int) bool {
	sched.mu.RLock()
	callback := sched.onReflection
	sched.mu.RUnlock()
	if callback == nil {
		return false
	}
	if err := callback(step); err != nil {
		fmt.Printf("   ! Reflection callback failed at step %d: %v\n", step, err)
	}
	// Installing a reflection hook replaces this pivot's generic salience
	// call; no second inference is admitted even on missing/invalid evidence.
	return true
}

// relevanceRealization performs pivotal relevance realization
func (sched *EchobeatsScheduler) relevanceRealization(question string) {
	fmt.Printf("   🎯 Relevance Realization: %s\n", question)

	// Assign to engine 1 (primary for relevance realization)
	task := &CognitiveTask{
		ID:          fmt.Sprintf("rr_%d", time.Now().UnixNano()),
		Type:        TaskRelevanceRealization,
		Description: question,
		Priority:    1.0,
		StartTime:   time.Now(),
	}

	sched.engine1.mu.Lock()
	sched.engine1.currentTask = task
	sched.engine1.mu.Unlock()

	// Generate relevance insight using LLM
	opts := llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   100,
		Routing: llm.RoutingOptions{
			TraceID:               task.ID,
			Intent:                "echobeats.relevance",
			PreferNative:          true,
			AllowRemote:           true,
			AllowFallback:         true,
			PolicyExplicit:        true,
			AllowQueue:            true,
			RequiredContextTokens: 512,
			LatencyClass:          backendcap.LatencyInteractive,
			QueueWait:             5 * time.Second,
		},
	}

	fullPrompt := "[System: You are performing relevance realization. Be concise and focused.]\n\n" + question
	result, err := sched.llmProvider.Generate(sched.ctx, fullPrompt, opts)
	if err != nil {
		fmt.Printf("      ! relevance generation failed: %v\n", err)
		result = fmt.Sprintf("generation failed: %v", err)
	}

	now := time.Now()
	task.CompletionTime = &now
	task.Result = result
	task.Success = err == nil
	sched.attachTaskRoute(task)

	sched.engine1.mu.Lock()
	sched.engine1.taskHistory = appendBoundedTaskHistory(sched.engine1.taskHistory, *task)
	sched.engine1.currentTask = nil
	sched.engine1.mu.Unlock()

	if task.Success {
		sched.mu.Lock()
		sched.presentCommitment = result
		sched.mu.Unlock()
	}

	fmt.Printf("      → %s\n", truncate(result, 70))
}

// affordanceInteraction performs actual affordance interaction
func (sched *EchobeatsScheduler) affordanceInteraction(stepNum int) {
	fmt.Printf("   🔧 Affordance Interaction (Step %d/5)\n", stepNum-1)

	// Distribute across three engines
	engineID := ((stepNum - 2) % 3) + 1
	var engine *InferenceEngine

	switch engineID {
	case 1:
		engine = sched.engine1
	case 2:
		engine = sched.engine2
	case 3:
		engine = sched.engine3
	}

	task := &CognitiveTask{
		ID:          fmt.Sprintf("ai_%d", time.Now().UnixNano()),
		Type:        TaskAffordanceInteraction,
		Description: fmt.Sprintf("Interact with available affordances (step %d)", stepNum-1),
		Priority:    0.8,
		StartTime:   time.Now(),
	}

	engine.mu.Lock()
	engine.currentTask = task
	engine.mu.Unlock()

	// The production orchestrator supplies a bounded, policy-gated enaction
	// callback. Without one, this remains reflective affordance planning only.
	sched.mu.RLock()
	onAffordance := sched.onAffordance
	sched.mu.RUnlock()
	if onAffordance != nil {
		result, verified := onAffordance(stepNum)
		now := time.Now()
		task.CompletionTime = &now
		task.Result = result
		task.Success = verified
		engine.mu.Lock()
		engine.taskHistory = appendBoundedTaskHistory(engine.taskHistory, *task)
		engine.currentTask = nil
		if verified {
			engine.performance = engine.performance*0.9 + 0.1
		}
		engine.mu.Unlock()
		sched.mu.Lock()
		sched.pastPerformance = append(sched.pastPerformance, result)
		if len(sched.pastPerformance) > 10 {
			sched.pastPerformance = sched.pastPerformance[1:]
		}
		sched.mu.Unlock()
		fmt.Printf("      [Engine %d] → %s\n", engineID, truncate(result, 60))
		return
	}

	// Reflect on an affordance without claiming an external effect.
	sched.mu.RLock()
	commitment := sched.presentCommitment
	sched.mu.RUnlock()

	prompt := fmt.Sprintf("[System: You are considering a possible affordance. Describe a proposal, not a completed action.]\n\nGiven commitment '%s', what bounded action could be considered? (Brief)", commitment)

	opts := llm.GenerateOptions{
		Temperature: 0.6,
		MaxTokens:   80,
		Routing: llm.RoutingOptions{
			TraceID:               task.ID,
			Intent:                "echobeats.affordance",
			PreferNative:          true,
			AllowRemote:           true,
			AllowFallback:         true,
			PolicyExplicit:        true,
			AllowQueue:            false,
			RequiredContextTokens: 384,
			LatencyClass:          backendcap.LatencyInteractive,
			QueueWait:             3 * time.Second,
		},
	}

	result, err := sched.llmProvider.Generate(sched.ctx, prompt, opts)
	if err != nil {
		fmt.Printf("      ! affordance generation failed: %v\n", err)
		result = fmt.Sprintf("generation failed: %v", err)
	}

	now := time.Now()
	task.CompletionTime = &now
	task.Result = result
	task.Success = err == nil
	sched.attachTaskRoute(task)

	engine.mu.Lock()
	engine.taskHistory = appendBoundedTaskHistory(engine.taskHistory, *task)
	engine.currentTask = nil
	engine.mu.Unlock()

	if task.Success {
		sched.mu.Lock()
		sched.pastPerformance = append(sched.pastPerformance, result)
		if len(sched.pastPerformance) > 10 {
			sched.pastPerformance = sched.pastPerformance[1:]
		}
		sched.mu.Unlock()
	}

	fmt.Printf("      [Engine %d] → %s\n", engineID, truncate(result, 60))
}

// salienceSimulation performs virtual salience simulation
func (sched *EchobeatsScheduler) salienceSimulation(stepNum int) {
	fmt.Printf("   🔮 Salience Simulation (Step %d/5)\n", stepNum)

	// Distribute across three engines
	engineID := ((stepNum - 1) % 3) + 1
	var engine *InferenceEngine

	switch engineID {
	case 1:
		engine = sched.engine1
	case 2:
		engine = sched.engine2
	case 3:
		engine = sched.engine3
	}

	task := &CognitiveTask{
		ID:          fmt.Sprintf("ss_%d", time.Now().UnixNano()),
		Type:        TaskSalienceSimulation,
		Description: fmt.Sprintf("Simulate future possibilities (step %d)", stepNum),
		Priority:    0.7,
		StartTime:   time.Now(),
	}

	engine.mu.Lock()
	engine.currentTask = task
	engine.mu.Unlock()

	// Simulate future possibilities
	prompt := fmt.Sprintf("[System: You are simulating future possibilities. Be imaginative but grounded.]\n\nImagine a possible future outcome (step %d of anticipation). What might happen?", stepNum)

	opts := llm.GenerateOptions{
		Temperature: 0.8,
		MaxTokens:   80,
		Routing: llm.RoutingOptions{
			TraceID:               task.ID,
			Intent:                "echobeats.salience",
			PreferNative:          false,
			AllowRemote:           true,
			AllowFallback:         true,
			PolicyExplicit:        true,
			AllowQueue:            true,
			RequiredContextTokens: 768,
			LatencyClass:          backendcap.LatencyNormal,
			QueueWait:             10 * time.Second,
		},
	}

	result, err := sched.llmProvider.Generate(sched.ctx, prompt, opts)
	if err != nil {
		fmt.Printf("      ! salience generation failed: %v\n", err)
		result = fmt.Sprintf("generation failed: %v", err)
	}

	now := time.Now()
	task.CompletionTime = &now
	task.Result = result
	task.Success = err == nil
	sched.attachTaskRoute(task)

	engine.mu.Lock()
	engine.taskHistory = appendBoundedTaskHistory(engine.taskHistory, *task)
	engine.currentTask = nil
	engine.mu.Unlock()

	if task.Success {
		sched.mu.Lock()
		sched.futureAnticipation = append(sched.futureAnticipation, result)
		if len(sched.futureAnticipation) > 10 {
			sched.futureAnticipation = sched.futureAnticipation[1:]
		}
		sched.mu.Unlock()
	}

	fmt.Printf("      [Engine %d] → %s\n", engineID, truncate(result, 60))
}

func (sched *EchobeatsScheduler) attachTaskRoute(task *CognitiveTask) {
	if task == nil {
		return
	}
	tracedProvider, ok := sched.llmProvider.(interface {
		GetRouteResult(string) (llm.BackendRoutingState, bool)
	})
	if !ok {
		return
	}
	state, exists := tracedProvider.GetRouteResult(task.ID)
	if !exists {
		return
	}
	task.Provider = state.SelectedProvider
	task.BackendKind = state.SelectedKind
	task.ModelID = state.SelectedModelID
	task.RouteReason = state.Decision.Reason
	task.Degraded = state.Degraded
	task.ProviderAttempts = len(state.Attempts)
}

// GetMetrics returns scheduler metrics
func (sched *EchobeatsScheduler) GetMetrics() map[string]interface{} {
	sched.mu.RLock()
	currentStep := sched.currentStep
	currentPhase := sched.currentPhase.String()
	cycleCount := sched.cycleCount
	totalSteps := sched.totalSteps
	totalCycles := sched.totalCycles
	presentCommitment := sched.presentCommitment
	sched.mu.RUnlock()

	return map[string]interface{}{
		"current_step":        currentStep,
		"current_phase":       currentPhase,
		"cycle_count":         cycleCount,
		"total_steps":         totalSteps,
		"total_cycles":        totalCycles,
		"engine1_performance": sched.engine1.performanceSnapshot(),
		"engine2_performance": sched.engine2.performanceSnapshot(),
		"engine3_performance": sched.engine3.performanceSnapshot(),
		"present_commitment":  presentCommitment,
	}
}

// GetCurrentPhase returns the current cognitive phase
func (sched *EchobeatsScheduler) GetCurrentPhase() CognitivePhase {
	sched.mu.RLock()
	defer sched.mu.RUnlock()

	return sched.currentPhase
}

// GetEngineStatus returns status of all three engines
func (sched *EchobeatsScheduler) GetEngineStatus() []map[string]interface{} {
	engines := []*InferenceEngine{sched.engine1, sched.engine2, sched.engine3}
	status := make([]map[string]interface{}, 3)

	for i, engine := range engines {
		engine.mu.RLock()
		status[i] = map[string]interface{}{
			"id":           engine.ID,
			"performance":  engine.performance,
			"task_history": len(engine.taskHistory),
			"current_task": engine.currentTask != nil,
		}
		engine.mu.RUnlock()
	}

	return status
}

// Helper function to truncate strings
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// AddGoal adds a goal to the scheduling queue
func (sched *EchobeatsScheduler) AddGoal(description string, priority float64) string {
	sched.mu.Lock()
	defer sched.mu.Unlock()

	goal := ScheduledGoal{
		ID:          fmt.Sprintf("goal_%d", time.Now().UnixNano()),
		Description: description,
		Priority:    priority,
		Progress:    0.0,
		Status:      GoalPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	sched.goalQueue = append(sched.goalQueue, goal)
	sched.priorityMatrix[goal.ID] = priority

	// Sort goals by priority
	sched.sortGoalsByPriority()

	fmt.Printf("🎯 Goal scheduled: %s (priority: %.2f)\n", description, priority)

	return goal.ID
}

// EnsureGoal idempotently records a pending explicit-ID goal for durable
// replay. It returns false only for invalid input; a pre-existing valid ID is
// already ensured and returns true without changing its progress or state.
func (sched *EchobeatsScheduler) EnsureGoal(goalID, description string, priority float64) bool {
	if strings.TrimSpace(goalID) == "" || strings.TrimSpace(description) == "" ||
		math.IsNaN(priority) || math.IsInf(priority, 0) || priority < 0 || priority > 1 {
		return false
	}

	sched.mu.Lock()
	defer sched.mu.Unlock()
	for i := range sched.goalQueue {
		if sched.goalQueue[i].ID == goalID {
			return true
		}
	}
	if sched.priorityMatrix == nil {
		sched.priorityMatrix = make(map[string]float64)
	}

	now := time.Now()
	sched.goalQueue = append(sched.goalQueue, ScheduledGoal{
		ID:          goalID,
		Description: description,
		Priority:    priority,
		Status:      GoalPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	sched.priorityMatrix[goalID] = priority
	sched.sortGoalsByPriority()
	return true
}

// sortGoalsByPriority sorts the goal queue by priority (highest first)
func (sched *EchobeatsScheduler) sortGoalsByPriority() {
	for i := range len(sched.goalQueue) - 1 {
		for j := i + 1; j < len(sched.goalQueue); j++ {
			if sched.goalQueue[j].Priority > sched.goalQueue[i].Priority {
				sched.goalQueue[i], sched.goalQueue[j] = sched.goalQueue[j], sched.goalQueue[i]
			}
		}
	}
}

// GetActiveGoal returns the current highest priority active goal
func (sched *EchobeatsScheduler) GetActiveGoal() *ScheduledGoal {
	sched.mu.RLock()
	defer sched.mu.RUnlock()

	for i := range sched.goalQueue {
		if sched.goalQueue[i].Status == GoalActive || sched.goalQueue[i].Status == GoalPending {
			goal := cloneScheduledGoal(sched.goalQueue[i])
			return &goal
		}
	}
	return nil
}

// UpdateGoalProgress updates the progress of a goal
func (sched *EchobeatsScheduler) UpdateGoalProgress(goalID string, progress float64) {
	sched.mu.Lock()
	var callback func(ScheduledGoal)
	var achievedGoal *ScheduledGoal

	for i := range sched.goalQueue {
		if sched.goalQueue[i].ID == goalID {
			sched.goalQueue[i].Progress = progress
			sched.goalQueue[i].UpdatedAt = time.Now()

			if progress >= 1.0 {
				if sched.goalQueue[i].Status != GoalCompleted {
					sched.goalQueue[i].Status = GoalCompleted
					sched.goalsCompleted++
					fmt.Printf("✅ Goal completed: %s\n", sched.goalQueue[i].Description)
					goal := cloneScheduledGoal(sched.goalQueue[i])
					achievedGoal = &goal
					callback = sched.onGoalAchieved
				}
			} else if progress > 0 {
				sched.goalQueue[i].Status = GoalActive
			}
			break
		}
	}
	sched.mu.Unlock()

	if callback != nil && achievedGoal != nil {
		callback(*achievedGoal)
	}
}

// ApplyGoalEvidence projects one durable evaluator-backed progress event. The
// same evidence ID is never applied twice, including during repeated replay.
func (sched *EchobeatsScheduler) ApplyGoalEvidence(goalID, description, evidenceID string, delta float64) bool {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	if strings.TrimSpace(goalID) == "" || strings.TrimSpace(evidenceID) == "" || delta <= 0 {
		return false
	}
	if sched.appliedGoalEvidence == nil {
		sched.appliedGoalEvidence = make(map[string]struct{})
	}
	if _, exists := sched.appliedGoalEvidence[evidenceID]; exists {
		return false
	}
	index := -1
	for i := range sched.goalQueue {
		if sched.goalQueue[i].ID == goalID {
			index = i
			break
		}
	}
	if index < 0 {
		now := time.Now()
		sched.goalQueue = append(sched.goalQueue, ScheduledGoal{
			ID: goalID, Description: description, Priority: 0.5,
			Status: GoalPending, CreatedAt: now, UpdatedAt: now,
		})
		index = len(sched.goalQueue) - 1
		sched.priorityMatrix[goalID] = 0.5
	}
	goal := &sched.goalQueue[index]
	goal.Progress = min(1, goal.Progress+delta)
	goal.UpdatedAt = time.Now()
	if goal.Progress >= 1 {
		if goal.Status != GoalCompleted {
			goal.Status = GoalCompleted
			sched.goalsCompleted++
		}
	} else {
		goal.Status = GoalActive
	}
	sched.appliedGoalEvidence[evidenceID] = struct{}{}
	return true
}

// ScheduleTemporalEvent adds an event to the temporal plan
func (sched *EchobeatsScheduler) ScheduleTemporalEvent(eventType, description string, scheduledAt time.Time, duration time.Duration, priority float64) string {
	sched.mu.Lock()
	defer sched.mu.Unlock()

	event := TemporalEvent{
		ID:          fmt.Sprintf("event_%d", time.Now().UnixNano()),
		Type:        eventType,
		Description: description,
		ScheduledAt: scheduledAt,
		Duration:    duration,
		Priority:    priority,
		Executed:    false,
	}

	sched.temporalPlan = append(sched.temporalPlan, event)

	return event.ID
}

// ProcessTemporalPlan checks and executes due temporal events
func (sched *EchobeatsScheduler) ProcessTemporalPlan() {
	sched.mu.Lock()
	defer sched.mu.Unlock()

	now := time.Now()
	for i := range sched.temporalPlan {
		if !sched.temporalPlan[i].Executed && now.After(sched.temporalPlan[i].ScheduledAt) {
			fmt.Printf("⏰ Executing temporal event: %s\n", sched.temporalPlan[i].Description)
			sched.temporalPlan[i].Executed = true
		}
	}
}

// UpdateTriadState updates a triad's activation and coherence
func (sched *EchobeatsScheduler) UpdateTriadState(triadID int, activation, coherence float64) {
	sched.mu.Lock()
	defer sched.mu.Unlock()

	if triadID >= 0 && triadID < 4 {
		sched.triadStates[triadID].Activation = activation
		sched.triadStates[triadID].Coherence = coherence
		sched.triadStates[triadID].LastUpdated = time.Now()
	}
}

// GetTriadStates returns the current triad states
func (sched *EchobeatsScheduler) GetTriadStates() [4]TriadState {
	sched.mu.RLock()
	defer sched.mu.RUnlock()

	return sched.triadStates
}

// DetectEmergence checks for emergence patterns across triads
func (sched *EchobeatsScheduler) DetectEmergence() (bool, string, float64) {
	sched.mu.Lock()

	// Calculate average triad coherence
	totalCoherence := 0.0
	for _, triad := range sched.triadStates {
		totalCoherence += triad.Coherence
	}
	avgCoherence := totalCoherence / 4.0

	// Calculate average activation
	totalActivation := 0.0
	for _, triad := range sched.triadStates {
		totalActivation += triad.Activation
	}
	avgActivation := totalActivation / 4.0

	// Emergence threshold
	emergenceStrength := avgCoherence * avgActivation

	if emergenceStrength > 0.7 {
		pattern := "cross-triad-synchronization"
		sched.emergenceEvents++
		callback := sched.onEmergenceDetected
		sched.mu.Unlock()
		if callback != nil {
			callback(pattern, emergenceStrength)
		}
		return true, pattern, emergenceStrength
	}

	sched.mu.Unlock()
	return false, "", 0.0
}

// SetOnCycleComplete sets the callback for cycle completion
func (sched *EchobeatsScheduler) SetOnCycleComplete(callback func(CycleMetrics)) {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.onCycleComplete = callback
}

// SetOnGoalAchieved sets the callback for goal achievement
func (sched *EchobeatsScheduler) SetOnGoalAchieved(callback func(ScheduledGoal)) {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.onGoalAchieved = callback
}

// SetOnEmergenceDetected sets the callback for emergence detection
func (sched *EchobeatsScheduler) SetOnEmergenceDetected(callback func(string, float64)) {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.onEmergenceDetected = callback
}

// SetOnAffordance binds the scheduler's action phase to an observed enaction
// implementation. The callback must return true only for independently verified effects.
func (sched *EchobeatsScheduler) SetOnAffordance(callback func(step int) (string, bool)) {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.onAffordance = callback
}

// SetOnReflection binds evidence-grounded metacognition to the reflective
// pivot at step 8. Callback errors are logged and do not become cognitive output.
func (sched *EchobeatsScheduler) SetOnReflection(callback func(step int) error) {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.onReflection = callback
}

// RotatePhase advances the phase rotation for tetrahedral synchronization
func (sched *EchobeatsScheduler) RotatePhase() {
	sched.mu.Lock()
	defer sched.mu.Unlock()

	// Rotate by 30 degrees (pi/6 radians) per step
	sched.phaseRotation += 3.14159265 / 6.0

	// Normalize to 0-2pi
	if sched.phaseRotation > 2*3.14159265 {
		sched.phaseRotation -= 2 * 3.14159265
	}

	// Update triad activations based on phase
	for i := range sched.triadStates {
		// Each triad is 90 degrees (pi/2) apart
		triadPhase := sched.phaseRotation + float64(i)*3.14159265/2.0
		// Sinusoidal activation pattern
		activation := 0.5 + 0.5*sin(triadPhase)
		sched.triadStates[i].Activation = activation
		sched.triadStates[i].LastUpdated = time.Now()
	}
}

// sin helper function
func sin(x float64) float64 {
	// Taylor series approximation for sin
	x = x - float64(int(x/(2*3.14159265)))*2*3.14159265
	result := x
	term := x
	for i := 1; i < 10; i++ {
		term *= -x * x / float64((2*i)*(2*i+1))
		result += term
	}
	return result
}

// GetGoalQueue returns a copy of the goal queue
func (sched *EchobeatsScheduler) GetGoalQueue() []ScheduledGoal {
	sched.mu.RLock()
	defer sched.mu.RUnlock()

	queue := make([]ScheduledGoal, len(sched.goalQueue))
	for i, goal := range sched.goalQueue {
		queue[i] = cloneScheduledGoal(goal)
	}
	return queue
}

func cloneScheduledGoal(goal ScheduledGoal) ScheduledGoal {
	copy := goal
	if goal.Deadline != nil {
		deadline := *goal.Deadline
		copy.Deadline = &deadline
	}
	copy.Dependencies = append([]string(nil), goal.Dependencies...)
	return copy
}

// GetCompleteCycleMetrics returns metrics for the current/last cycle
func (sched *EchobeatsScheduler) GetCompleteCycleMetrics() CycleMetrics {
	sched.mu.RLock()
	// Count active goals that advanced
	goalsAdvanced := 0
	for _, goal := range sched.goalQueue {
		if goal.Status == GoalActive {
			goalsAdvanced++
		}
	}

	// Calculate emergence level
	totalCoherence := 0.0
	triadCoherence := [4]float64{}
	for i, triad := range sched.triadStates {
		totalCoherence += triad.Coherence
		triadCoherence[i] = triad.Coherence
	}
	emergenceLevel := totalCoherence / 4.0
	cycleCount := sched.cycleCount
	sched.mu.RUnlock()

	return CycleMetrics{
		CycleNumber:    cycleCount,
		StepsCompleted: 12,
		GoalsAdvanced:  goalsAdvanced,
		EmergenceLevel: emergenceLevel,
		EnginePerformance: [3]float64{
			sched.engine1.performanceSnapshot(),
			sched.engine2.performanceSnapshot(),
			sched.engine3.performanceSnapshot(),
		},
		TriadCoherence: triadCoherence,
		Timestamp:      time.Now(),
	}
}

// Pause pauses the scheduler
func (sched *EchobeatsScheduler) Pause() {
	// Wait for any current step before recording the paused state so that no
	// provider call remains in flight once Pause returns.
	sched.stepMu.Lock()
	defer sched.stepMu.Unlock()

	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.paused = true
}

// Resume resumes the scheduler
func (sched *EchobeatsScheduler) Resume() {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.paused = false
}

// IsPaused returns whether the scheduler is paused.
func (sched *EchobeatsScheduler) IsPaused() bool {
	sched.mu.RLock()
	defer sched.mu.RUnlock()
	return sched.paused
}

// IsRunning returns whether the scheduler is running
func (sched *EchobeatsScheduler) IsRunning() bool {
	sched.mu.RLock()
	defer sched.mu.RUnlock()
	return sched.running
}

// SetCallbacks sets the callback functions for echobeats events
func (sched *EchobeatsScheduler) SetCallbacks(
	onCycleComplete func(metrics CycleMetrics),
	onGoalAchieved func(goal ScheduledGoal),
	onEmergenceDetected func(pattern string, strength float64),
) {
	sched.mu.Lock()
	defer sched.mu.Unlock()

	sched.onCycleComplete = onCycleComplete
	sched.onGoalAchieved = onGoalAchieved
	sched.onEmergenceDetected = onEmergenceDetected
}
