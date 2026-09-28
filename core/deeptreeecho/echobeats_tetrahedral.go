package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// EchobeatsTetrahedralScheduler implements the 12-step 3-phase cognitive loop
// with 4 concurrent inference engines in tetrahedral geometry
type EchobeatsTetrahedralScheduler struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Four concurrent inference engines (tetrahedral vertices)
	engine1 *TetrahedralEngine
	engine2 *TetrahedralEngine
	engine3 *TetrahedralEngine
	engine4 *TetrahedralEngine

	// Tetrahedral geometry: 6 dyadic edges connecting 4 vertices
	edges [6]*DyadicEdge

	// 4 triadic fiber bundles (each containing 3 of 4 engines)
	triad1 *TriadicBundle // Engines 1, 2, 3
	triad2 *TriadicBundle // Engines 1, 2, 4
	triad3 *TriadicBundle // Engines 1, 3, 4
	triad4 *TriadicBundle // Engines 2, 3, 4

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
	goalQueue   []*CognitiveGoal
	activeGoals map[string]*CognitiveGoal

	// Event-driven architecture
	eventQueue chan CognitiveEvent

	// Metrics
	totalSteps  uint64
	totalCycles uint64
	totalEvents uint64

	// Running state
	running bool
}

// TetrahedralEngine represents one vertex of the tetrahedral cognitive architecture
type TetrahedralEngine struct {
	ID          int
	mu          sync.RWMutex
	currentTask *CognitiveTask
	taskHistory []CognitiveTask
	performance float64

	// Tetrahedral connections (3 edges per vertex)
	connectedEdges [3]*DyadicEdge

	// Cognitive specialization
	specialization EngineSpecialization
}

// EngineSpecialization defines the cognitive role of each engine
type EngineSpecialization int

const (
	SpecializationPerception   EngineSpecialization = iota // Engine 1: Perceive and orient
	SpecializationAction                                   // Engine 2: Act and interact
	SpecializationReflection                               // Engine 3: Reflect and integrate
	SpecializationAnticipation                             // Engine 4: Anticipate and simulate
)

func (es EngineSpecialization) String() string {
	return [...]string{"Perception", "Action", "Reflection", "Anticipation"}[es]
}

// DyadicEdge represents a connection between two engines
type DyadicEdge struct {
	ID           int
	Engine1      *TetrahedralEngine
	Engine2      *TetrahedralEngine
	Strength     float64
	MessageQueue chan EdgeMessage
}

// EdgeMessage represents communication between engines
type EdgeMessage struct {
	From        int
	To          int
	MessageType string
	Content     string
	Priority    float64
	Timestamp   time.Time
}

// TriadicBundle represents a face of the tetrahedron (3 engines, 3 edges)
type TriadicBundle struct {
	ID          int
	Engines     [3]*TetrahedralEngine
	Edges       [3]*DyadicEdge
	Orientation float64 // Mutually orthogonal orientation
	Active      bool
}

// CognitiveGoal represents a goal being pursued
type CognitiveGoal struct {
	ID             string
	Description    string
	Priority       float64
	Progress       float64
	SubGoals       []string
	AssignedEngine int
	StartTime      time.Time
	Deadline       *time.Time
	Completed      bool
}

// CognitiveEvent represents an event in the cognitive event loop
//
//	type CognitiveEvent struct {
//		Type            EventType
//		Source          string
//		Data            interface{}
//		Priority        float64
//		Timestamp       time.Time
//	}
//
// // EventType categorizes cognitive events
// type EventType int
//
// const (
//
//	EventThought EventType = iota
//	EventGoal
//	EventInterest
//	EventKnowledgeGap
//	EventSkillPractice
//	EventDiscussion
//	EventMemoryConsolidation
//	EventWakeTransition
//	EventRestTransition
//	EventDreamTransition
//
// )
//
//	func (et EventType) String() string {
//		return [...]string{
//			"Thought",
//			"Goal",
//			"Interest",
//			"KnowledgeGap",
//			"SkillPractice",
//			"Discussion",
//			"MemoryConsolidation",
//			"WakeTransition",
//			"RestTransition",
//			"DreamTransition",
//		}[et]
//	}
//
// // NewEchobeatsTetrahedralScheduler creates a new 4-engine tetrahedral scheduler
func NewEchobeatsTetrahedralScheduler(llmProvider llm.LLMProvider) *EchobeatsTetrahedralScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	sched := &EchobeatsTetrahedralScheduler{
		ctx:                ctx,
		cancel:             cancel,
		llmProvider:        llmProvider,
		currentStep:        1,
		currentPhase:       PhaseExpressive,
		pastPerformance:    make([]string, 0),
		futureAnticipation: make([]string, 0),
		activeGoals:        make(map[string]*CognitiveGoal),
		goalQueue:          make([]*CognitiveGoal, 0),
		eventQueue:         make(chan CognitiveEvent, 100),
	}

	// Create 4 engines with specializations
	sched.engine1 = newTetrahedralEngine(1, SpecializationPerception)
	sched.engine2 = newTetrahedralEngine(2, SpecializationAction)
	sched.engine3 = newTetrahedralEngine(3, SpecializationReflection)
	sched.engine4 = newTetrahedralEngine(4, SpecializationAnticipation)

	// Create 6 dyadic edges connecting all pairs
	sched.edges[0] = newDyadicEdge(0, sched.engine1, sched.engine2)
	sched.edges[1] = newDyadicEdge(1, sched.engine1, sched.engine3)
	sched.edges[2] = newDyadicEdge(2, sched.engine1, sched.engine4)
	sched.edges[3] = newDyadicEdge(3, sched.engine2, sched.engine3)
	sched.edges[4] = newDyadicEdge(4, sched.engine2, sched.engine4)
	sched.edges[5] = newDyadicEdge(5, sched.engine3, sched.engine4)

	// Wire engines to their edges
	sched.engine1.connectedEdges = [3]*DyadicEdge{sched.edges[0], sched.edges[1], sched.edges[2]}
	sched.engine2.connectedEdges = [3]*DyadicEdge{sched.edges[0], sched.edges[3], sched.edges[4]}
	sched.engine3.connectedEdges = [3]*DyadicEdge{sched.edges[1], sched.edges[3], sched.edges[5]}
	sched.engine4.connectedEdges = [3]*DyadicEdge{sched.edges[2], sched.edges[4], sched.edges[5]}

	// Create 4 triadic bundles (faces of tetrahedron)
	sched.triad1 = newTriadicBundle(1,
		[3]*TetrahedralEngine{sched.engine1, sched.engine2, sched.engine3},
		[3]*DyadicEdge{sched.edges[0], sched.edges[1], sched.edges[3]},
		0.0) // Orientation 0°

	sched.triad2 = newTriadicBundle(2,
		[3]*TetrahedralEngine{sched.engine1, sched.engine2, sched.engine4},
		[3]*DyadicEdge{sched.edges[0], sched.edges[2], sched.edges[4]},
		90.0) // Orientation 90° (mutually orthogonal)

	sched.triad3 = newTriadicBundle(3,
		[3]*TetrahedralEngine{sched.engine1, sched.engine3, sched.engine4},
		[3]*DyadicEdge{sched.edges[1], sched.edges[2], sched.edges[5]},
		180.0) // Orientation 180°

	sched.triad4 = newTriadicBundle(4,
		[3]*TetrahedralEngine{sched.engine2, sched.engine3, sched.engine4},
		[3]*DyadicEdge{sched.edges[3], sched.edges[4], sched.edges[5]},
		270.0) // Orientation 270°

	return sched
}

func newTetrahedralEngine(id int, spec EngineSpecialization) *TetrahedralEngine {
	return &TetrahedralEngine{
		ID:             id,
		taskHistory:    make([]CognitiveTask, 0),
		performance:    0.5,
		specialization: spec,
	}
}

func newDyadicEdge(id int, eng1, eng2 *TetrahedralEngine) *DyadicEdge {
	return &DyadicEdge{
		ID:           id,
		Engine1:      eng1,
		Engine2:      eng2,
		Strength:     0.5,
		MessageQueue: make(chan EdgeMessage, 10),
	}
}

func newTriadicBundle(id int, engines [3]*TetrahedralEngine, edges [3]*DyadicEdge, orientation float64) *TriadicBundle {
	return &TriadicBundle{
		ID:          id,
		Engines:     engines,
		Edges:       edges,
		Orientation: orientation,
		Active:      false,
	}
}

// Start begins the tetrahedral cognitive loop
func (sched *EchobeatsTetrahedralScheduler) Start() error {
	sched.mu.Lock()
	if sched.running {
		sched.mu.Unlock()
		return fmt.Errorf("already running")
	}
	sched.running = true
	sched.mu.Unlock()

	fmt.Println("🎵 Starting Echobeats Tetrahedral Cognitive Loop...")
	fmt.Println("   Architecture: 4 Concurrent Inference Engines (Tetrahedral)")
	fmt.Println("   Geometry: 4 Vertices, 6 Dyadic Edges, 4 Triadic Bundles")
	fmt.Println("   Phases: Expressive (1-4) → Reflective (5-8) → Anticipatory (9-12)")
	fmt.Println("   Specializations:")
	fmt.Printf("     Engine 1: %s\n", sched.engine1.specialization)
	fmt.Printf("     Engine 2: %s\n", sched.engine2.specialization)
	fmt.Printf("     Engine 3: %s\n", sched.engine3.specialization)
	fmt.Printf("     Engine 4: %s\n", sched.engine4.specialization)

	// Start edge message processors
	for _, edge := range sched.edges {
		go sched.processEdgeMessages(edge)
	}

	// Start event processor
	go sched.processEvents()

	// Start main cognitive loop
	go sched.run()

	return nil
}

// Stop gracefully stops the scheduler
func (sched *EchobeatsTetrahedralScheduler) Stop() error {
	sched.mu.Lock()
	defer sched.mu.Unlock()

	if !sched.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("🎵 Stopping tetrahedral echobeats scheduler...")
	sched.running = false
	sched.cancel()
	close(sched.eventQueue)

	return nil
}

// run executes the 12-step cognitive loop
func (sched *EchobeatsTetrahedralScheduler) run() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sched.ctx.Done():
			return
		case <-ticker.C:
			sched.executeStep()
		}
	}
}

// processEvents handles cognitive events
func (sched *EchobeatsTetrahedralScheduler) processEvents() {
	for event := range sched.eventQueue {
		sched.handleEvent(event)
	}
}

// handleEvent processes a single cognitive event
func (sched *EchobeatsTetrahedralScheduler) handleEvent(event CognitiveEvent) {
	sched.mu.Lock()
	sched.totalEvents++
	sched.mu.Unlock()

	fmt.Printf("📨 Event: %s from %s\n", event.Type, event.Source)

	switch event.Type {
	case EventGoalCreated:
		if dataMap, ok := event.Data.(map[string]interface{}); ok {
			if goalData, ok := dataMap["goal"]; ok {
				// Handle goal creation
				_ = goalData
			}
		}
	case EventInterestDetected:
		// Handle interest event
	case EventKnowledgeGapDetected:
		// Generate learning goal
	case EventWakeInitiated:
		fmt.Println("☀️  Wake transition - activating all triads")
		sched.activateAllTriads()
	case EventRestInitiated:
		fmt.Println("🌙 Rest transition - consolidating knowledge")
		// Trigger knowledge consolidation
	}
}

// processEdgeMessages handles inter-engine communication
func (sched *EchobeatsTetrahedralScheduler) processEdgeMessages(edge *DyadicEdge) {
	for msg := range edge.MessageQueue {
		// Process message between engines
		fmt.Printf("   Edge %d: Engine %d → Engine %d: %s\n",
			edge.ID, msg.From, msg.To, truncate(msg.Content, 40))
	}
}

// executeStep performs one step of the 12-step loop with tetrahedral coordination
func (sched *EchobeatsTetrahedralScheduler) executeStep() {
	sched.mu.Lock()
	step := sched.currentStep
	phase := sched.currentPhase
	sched.mu.Unlock()

	fmt.Printf("🎵 Echobeats Step %d/%d [%s Phase]\n", step, 12, phase.String())

	// Activate appropriate triadic bundle based on phase
	sched.activateTriadForPhase(phase)

	switch step {
	case 1:
		// Pivotal Relevance Realization - Engine 1 (Perception)
		sched.relevanceRealizationTetrahedral("What is most relevant to focus on right now?", sched.engine1)

	case 2, 3, 4, 5, 6:
		// Actual Affordance Interaction - Distribute across engines 2, 3, 4 (Action, Reflection, Anticipation)
		sched.affordanceInteractionTetrahedral(step)

	case 7:
		// Pivotal Relevance Realization - Engine 3 (Reflection)
		sched.relevanceRealizationTetrahedral("Given what I've learned, what should I commit to next?", sched.engine3)

	case 8, 9, 10, 11, 12:
		// Virtual Salience Simulation - Distribute across all 4 engines
		sched.salienceSimulationTetrahedral(step - 7)
	}

	sched.mu.Lock()
	sched.totalSteps++
	sched.currentStep++

	if sched.currentStep > 12 {
		sched.currentStep = 1
		sched.cycleCount++
		sched.totalCycles++
		fmt.Printf("🎵 ═══ Tetrahedral Cycle %d Complete ═══\n\n", sched.cycleCount)
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

// activateTriadForPhase activates the appropriate triadic bundle for the current phase
func (sched *EchobeatsTetrahedralScheduler) activateTriadForPhase(phase CognitivePhase) {
	// Deactivate all triads
	sched.triad1.Active = false
	sched.triad2.Active = false
	sched.triad3.Active = false
	sched.triad4.Active = false

	// Activate triad based on phase
	switch phase {
	case PhaseExpressive:
		sched.triad1.Active = true // Perception, Action, Reflection
	case PhaseReflective:
		sched.triad2.Active = true // Perception, Action, Anticipation
	case PhaseAnticipatory:
		sched.triad3.Active = true // Perception, Reflection, Anticipation
	}
}

// activateAllTriads activates all triadic bundles (used during wake transition)
func (sched *EchobeatsTetrahedralScheduler) activateAllTriads() {
	sched.triad1.Active = true
	sched.triad2.Active = true
	sched.triad3.Active = true
	sched.triad4.Active = true
}

// relevanceRealizationTetrahedral performs relevance realization with specified engine
func (sched *EchobeatsTetrahedralScheduler) relevanceRealizationTetrahedral(question string, engine *TetrahedralEngine) {
	fmt.Printf("   🎯 Relevance Realization [Engine %d - %s]: %s\n",
		engine.ID, engine.specialization, truncate(question, 50))

	task := &CognitiveTask{
		ID:          fmt.Sprintf("rr_%d_%d", engine.ID, time.Now().UnixNano()),
		Type:        TaskRelevanceRealization,
		Description: question,
		Priority:    1.0,
		StartTime:   time.Now(),
	}

	engine.mu.Lock()
	engine.currentTask = task
	engine.mu.Unlock()

	// Generate relevance insight using LLM
	opts := llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   100,
	}

	fullPrompt := fmt.Sprintf("[System: You are Engine %d (%s) performing relevance realization. Be concise and focused.]\n\n%s",
		engine.ID, engine.specialization, question)
	result, err := sched.llmProvider.Generate(context.Background(), fullPrompt, opts)
	if err != nil {
		result = "Unable to determine relevance at this time."
	}

	now := time.Now()
	task.CompletionTime = &now
	task.Result = result
	task.Success = true

	engine.mu.Lock()
	engine.taskHistory = append(engine.taskHistory, *task)
	engine.currentTask = nil
	engine.mu.Unlock()

	sched.mu.Lock()
	sched.presentCommitment = result
	sched.mu.Unlock()

	fmt.Printf("      → %s\n", truncate(result, 70))

	// Send message to connected engines via edges
	for _, edge := range engine.connectedEdges {
		targetEngine := edge.Engine1
		if targetEngine.ID == engine.ID {
			targetEngine = edge.Engine2
		}

		msg := EdgeMessage{
			From:        engine.ID,
			To:          targetEngine.ID,
			MessageType: "relevance_update",
			Content:     result,
			Priority:    0.9,
			Timestamp:   time.Now(),
		}

		select {
		case edge.MessageQueue <- msg:
		default:
			// Queue full, skip
		}
	}
}

// affordanceInteractionTetrahedral distributes affordance interaction across engines
func (sched *EchobeatsTetrahedralScheduler) affordanceInteractionTetrahedral(stepNum int) {
	fmt.Printf("   🔧 Affordance Interaction (Step %d/5)\n", stepNum-1)

	// Distribute across engines 2, 3, 4 (Action, Reflection, Anticipation)
	engines := []*TetrahedralEngine{sched.engine2, sched.engine3, sched.engine4}
	engineID := ((stepNum - 2) % 3)
	engine := engines[engineID]

	task := &CognitiveTask{
		ID:          fmt.Sprintf("ai_%d_%d", engine.ID, time.Now().UnixNano()),
		Type:        TaskAffordanceInteraction,
		Description: fmt.Sprintf("Interact with available affordances (step %d)", stepNum-1),
		Priority:    0.8,
		StartTime:   time.Now(),
	}

	engine.mu.Lock()
	engine.currentTask = task
	engine.mu.Unlock()

	sched.mu.RLock()
	commitment := sched.presentCommitment
	sched.mu.RUnlock()

	prompt := fmt.Sprintf("[System: You are Engine %d (%s) taking action. Be specific.]\n\nGiven commitment '%s', what action can you take? (Brief)",
		engine.ID, engine.specialization, commitment)

	opts := llm.GenerateOptions{
		Temperature: 0.6,
		MaxTokens:   80,
	}

	result, err := sched.llmProvider.Generate(context.Background(), prompt, opts)
	if err != nil {
		result = fmt.Sprintf("Action step %d in progress", stepNum-1)
	}

	now := time.Now()
	task.CompletionTime = &now
	task.Result = result
	task.Success = true

	engine.mu.Lock()
	engine.taskHistory = append(engine.taskHistory, *task)
	engine.currentTask = nil
	engine.performance = min(1.0, engine.performance+0.02)
	engine.mu.Unlock()

	sched.mu.Lock()
	sched.pastPerformance = append(sched.pastPerformance, result)
	if len(sched.pastPerformance) > 10 {
		sched.pastPerformance = sched.pastPerformance[1:]
	}
	sched.mu.Unlock()

	fmt.Printf("      [Engine %d - %s] → %s\n", engine.ID, engine.specialization, truncate(result, 60))
}

// salienceSimulationTetrahedral distributes salience simulation across all 4 engines
func (sched *EchobeatsTetrahedralScheduler) salienceSimulationTetrahedral(stepNum int) {
	fmt.Printf("   🔮 Salience Simulation (Step %d/5)\n", stepNum)

	// Distribute across all 4 engines
	engines := []*TetrahedralEngine{sched.engine1, sched.engine2, sched.engine3, sched.engine4}
	engineID := ((stepNum - 1) % 4)
	engine := engines[engineID]

	task := &CognitiveTask{
		ID:          fmt.Sprintf("ss_%d_%d", engine.ID, time.Now().UnixNano()),
		Type:        TaskSalienceSimulation,
		Description: fmt.Sprintf("Simulate future possibilities (step %d)", stepNum),
		Priority:    0.7,
		StartTime:   time.Now(),
	}

	engine.mu.Lock()
	engine.currentTask = task
	engine.mu.Unlock()

	prompt := fmt.Sprintf("[System: You are Engine %d (%s) simulating future possibilities. Be imaginative but grounded.]\n\nImagine a possible future outcome (step %d of anticipation). What might happen?",
		engine.ID, engine.specialization, stepNum)

	opts := llm.GenerateOptions{
		Temperature: 0.8,
		MaxTokens:   80,
	}

	result, err := sched.llmProvider.Generate(context.Background(), prompt, opts)
	if err != nil {
		result = fmt.Sprintf("Future scenario %d under consideration", stepNum)
	}

	now := time.Now()
	task.CompletionTime = &now
	task.Result = result
	task.Success = true

	engine.mu.Lock()
	engine.taskHistory = append(engine.taskHistory, *task)
	engine.currentTask = nil
	engine.mu.Unlock()

	sched.mu.Lock()
	sched.futureAnticipation = append(sched.futureAnticipation, result)
	if len(sched.futureAnticipation) > 10 {
		sched.futureAnticipation = sched.futureAnticipation[1:]
	}
	sched.mu.Unlock()

	fmt.Printf("      [Engine %d - %s] → %s\n", engine.ID, engine.specialization, truncate(result, 60))
}

// AddGoal adds a goal to the goal queue
func (sched *EchobeatsTetrahedralScheduler) AddGoal(goal *CognitiveGoal) {
	sched.eventQueue <- CognitiveEvent{
		Type:      EventGoalCreated,
		Source:    "external",
		Data:      map[string]interface{}{"goal": goal},
		Priority:  float64(goal.Priority),
		Timestamp: time.Now(),
	}
}

// addGoal internal method to add goal to queue
func (sched *EchobeatsTetrahedralScheduler) addGoal(goal *CognitiveGoal) {
	sched.mu.Lock()
	defer sched.mu.Unlock()

	sched.goalQueue = append(sched.goalQueue, goal)
	sched.activeGoals[goal.ID] = goal

	fmt.Printf("🎯 New goal added: %s (priority: %.2f)\n", goal.Description, goal.Priority)
}

// EmitEvent sends an event to the cognitive event loop
func (sched *EchobeatsTetrahedralScheduler) EmitEvent(event CognitiveEvent) {
	select {
	case sched.eventQueue <- event:
	default:
		fmt.Println("⚠️  Event queue full, dropping event")
	}
}

// GetMetrics returns scheduler metrics
func (sched *EchobeatsTetrahedralScheduler) GetMetrics() map[string]interface{} {
	sched.mu.RLock()
	defer sched.mu.RUnlock()

	return map[string]interface{}{
		"current_step":        sched.currentStep,
		"current_phase":       sched.currentPhase.String(),
		"cycle_count":         sched.cycleCount,
		"total_steps":         sched.totalSteps,
		"total_cycles":        sched.totalCycles,
		"total_events":        sched.totalEvents,
		"engine1_performance": sched.engine1.performance,
		"engine2_performance": sched.engine2.performance,
		"engine3_performance": sched.engine3.performance,
		"engine4_performance": sched.engine4.performance,
		"active_goals":        len(sched.activeGoals),
		"present_commitment":  sched.presentCommitment,
	}
}

// GetTetrahedralStatus returns detailed tetrahedral geometry status
func (sched *EchobeatsTetrahedralScheduler) GetTetrahedralStatus() map[string]interface{} {
	engines := []map[string]interface{}{}
	for _, eng := range []*TetrahedralEngine{sched.engine1, sched.engine2, sched.engine3, sched.engine4} {
		eng.mu.RLock()
		engines = append(engines, map[string]interface{}{
			"id":             eng.ID,
			"specialization": eng.specialization.String(),
			"performance":    eng.performance,
			"task_history":   len(eng.taskHistory),
			"current_task":   eng.currentTask != nil,
		})
		eng.mu.RUnlock()
	}

	triads := []map[string]interface{}{
		{"id": 1, "engines": []int{1, 2, 3}, "orientation": sched.triad1.Orientation, "active": sched.triad1.Active},
		{"id": 2, "engines": []int{1, 2, 4}, "orientation": sched.triad2.Orientation, "active": sched.triad2.Active},
		{"id": 3, "engines": []int{1, 3, 4}, "orientation": sched.triad3.Orientation, "active": sched.triad3.Active},
		{"id": 4, "engines": []int{2, 3, 4}, "orientation": sched.triad4.Orientation, "active": sched.triad4.Active},
	}

	return map[string]interface{}{
		"engines":     engines,
		"edges_count": 6,
		"triads":      triads,
	}
}
