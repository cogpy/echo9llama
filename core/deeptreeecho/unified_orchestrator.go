package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// UnifiedOrchestrator coordinates all cognitive subsystems for autonomous operation
// This is the central nervous system that ties together all components
type UnifiedOrchestrator struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc

	// Core subsystem references
	eventBus        *CognitiveEventBus
	knowledgeGraph  *KnowledgeGraph
	stateStore      *PersistentStateStore
	stateMachine    *CognitiveStateMachine
	patternProcessor *NeuralPatternProcessor
	semanticMemory  *SemanticMemory
	scheduler       *CognitiveScheduler

	// Orchestration state
	orchestrationMode OrchestratorMode
	cycleCount        uint64
	lastCycleTime     time.Time

	// Gestalt aggregation
	gestaltCache      map[string]interface{}
	gestaltTimestamp  time.Time

	// Event subscriptions
	subscriptions     []string

	// Metrics
	totalCycles       uint64
	totalEvents       uint64
	totalDecisions    uint64

	// Running state
	running           bool
}

// OrchestratorMode represents the current orchestration mode
type OrchestratorMode int

const (
	ModeIdle OrchestratorMode = iota
	ModeActive
	ModeReflective
	ModeConsolidating
	ModeEmergent
)

func (m OrchestratorMode) String() string {
	return [...]string{
		"Idle",
		"Active",
		"Reflective",
		"Consolidating",
		"Emergent",
	}[m]
}

// NewUnifiedOrchestrator creates a new unified orchestrator
func NewUnifiedOrchestrator(
	eventBus *CognitiveEventBus,
	knowledgeGraph *KnowledgeGraph,
	stateStore *PersistentStateStore,
	stateMachine *CognitiveStateMachine,
) *UnifiedOrchestrator {
	ctx, cancel := context.WithCancel(context.Background())

	uo := &UnifiedOrchestrator{
		ctx:              ctx,
		cancel:           cancel,
		eventBus:         eventBus,
		knowledgeGraph:   knowledgeGraph,
		stateStore:       stateStore,
		stateMachine:     stateMachine,
		orchestrationMode: ModeIdle,
		gestaltCache:     make(map[string]interface{}),
		subscriptions:    make([]string, 0),
	}

	return uo
}

// SetPatternProcessor sets the neural pattern processor
func (uo *UnifiedOrchestrator) SetPatternProcessor(pp *NeuralPatternProcessor) {
	uo.mu.Lock()
	defer uo.mu.Unlock()
	uo.patternProcessor = pp
}

// SetSemanticMemory sets the semantic memory system
func (uo *UnifiedOrchestrator) SetSemanticMemory(sm *SemanticMemory) {
	uo.mu.Lock()
	defer uo.mu.Unlock()
	uo.semanticMemory = sm
}

// SetScheduler sets the cognitive scheduler
func (uo *UnifiedOrchestrator) SetScheduler(cs *CognitiveScheduler) {
	uo.mu.Lock()
	defer uo.mu.Unlock()
	uo.scheduler = cs
}

// Start begins the unified orchestrator
func (uo *UnifiedOrchestrator) Start() error {
	uo.mu.Lock()
	defer uo.mu.Unlock()

	if uo.running {
		return fmt.Errorf("unified orchestrator already running")
	}

	uo.running = true
	fmt.Println("🎭 Unified Orchestrator started")

	// Subscribe to key events
	uo.subscribeToEvents()

	// Start orchestration loop
	go uo.orchestrationLoop()

	// Start gestalt aggregation loop
	go uo.gestaltAggregationLoop()

	return nil
}

// Stop gracefully stops the unified orchestrator
func (uo *UnifiedOrchestrator) Stop() error {
	uo.mu.Lock()
	defer uo.mu.Unlock()

	if !uo.running {
		return fmt.Errorf("unified orchestrator not running")
	}

	uo.cancel()
	uo.running = false
	fmt.Println("🎭 Unified Orchestrator stopped")

	return nil
}

// subscribeToEvents sets up event subscriptions
func (uo *UnifiedOrchestrator) subscribeToEvents() {
	// Subscribe to state transitions
	uo.eventBus.Subscribe(EventStateTransition, func(event CognitiveEvent) {
		uo.handleStateTransition(event)
	})

	// Subscribe to emergence detection
	uo.eventBus.Subscribe(EventEmergenceDetected, func(event CognitiveEvent) {
		uo.handleEmergence(event)
	})

	// Subscribe to knowledge acquisition
	uo.eventBus.Subscribe(EventKnowledgeAcquired, func(event CognitiveEvent) {
		uo.handleKnowledgeAcquisition(event)
	})

	// Subscribe to phase transitions
	uo.eventBus.Subscribe(EventPhaseTransition, func(event CognitiveEvent) {
		uo.handlePhaseTransition(event)
	})

	// Subscribe to thought generation
	uo.eventBus.Subscribe(EventThoughtGenerated, func(event CognitiveEvent) {
		uo.handleThoughtGenerated(event)
	})

	fmt.Println("   ✓ Event subscriptions established")
}

// orchestrationLoop is the main orchestration cycle
func (uo *UnifiedOrchestrator) orchestrationLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-uo.ctx.Done():
			return
		case <-ticker.C:
			uo.runOrchestrationCycle()
		}
	}
}

// runOrchestrationCycle performs one orchestration cycle
func (uo *UnifiedOrchestrator) runOrchestrationCycle() {
	uo.mu.Lock()
	defer uo.mu.Unlock()

	uo.cycleCount++
	uo.totalCycles++
	uo.lastCycleTime = time.Now()

	// 1. Sense: Gather current state from all subsystems
	currentState := uo.gatherCurrentState()

	// 2. Think: Analyze state and determine actions
	decisions := uo.analyzeAndDecide(currentState)

	// 3. Act: Execute decisions
	uo.executeDecisions(decisions)

	// 4. Reflect: Update mode based on activity
	uo.updateOrchestratorMode()

	// 5. Persist: Save state periodically
	if uo.cycleCount%10 == 0 {
		uo.persistState()
	}
}

// gatherCurrentState collects state from all subsystems
func (uo *UnifiedOrchestrator) gatherCurrentState() map[string]interface{} {
	state := make(map[string]interface{})

	// State machine state
	if uo.stateMachine != nil {
		state["cognitive_state"] = string(uo.stateMachine.State())
		state["permitted_triggers"] = uo.stateMachine.PermittedTriggers()
	}

	// Knowledge graph metrics
	if uo.knowledgeGraph != nil {
		state["knowledge_metrics"] = uo.knowledgeGraph.GetMetrics()
	}

	// Pattern processor metrics
	if uo.patternProcessor != nil {
		state["pattern_metrics"] = uo.patternProcessor.GetMetrics()
	}

	// Semantic memory metrics
	if uo.semanticMemory != nil {
		state["memory_metrics"] = uo.semanticMemory.GetMetrics()
	}

	// Scheduler metrics
	if uo.scheduler != nil {
		state["scheduler_metrics"] = uo.scheduler.GetMetrics()
	}

	state["orchestrator_mode"] = uo.orchestrationMode.String()
	state["cycle_count"] = uo.cycleCount

	return state
}

// analyzeAndDecide analyzes current state and makes decisions
func (uo *UnifiedOrchestrator) analyzeAndDecide(state map[string]interface{}) []OrchestratorDecision {
	decisions := make([]OrchestratorDecision, 0)

	// Check if we should transition cognitive state
	if cogState, ok := state["cognitive_state"].(string); ok {
		decision := uo.decideStateTransition(cogState)
		if decision != nil {
			decisions = append(decisions, *decision)
		}
	}

	// Check if we should learn new patterns
	if uo.patternProcessor != nil {
		if metrics, ok := state["pattern_metrics"].(map[string]interface{}); ok {
			if totalPatterns, ok := metrics["total_patterns"].(uint64); ok && totalPatterns < 10 {
				decisions = append(decisions, OrchestratorDecision{
					Type:     DecisionLearnPattern,
					Priority: 0.7,
					Data: map[string]interface{}{
						"reason": "pattern_count_low",
					},
				})
			}
		}
	}

	// Check if we should consolidate knowledge
	if uo.cycleCount%100 == 0 {
		decisions = append(decisions, OrchestratorDecision{
			Type:     DecisionConsolidate,
			Priority: 0.5,
			Data: map[string]interface{}{
				"reason": "periodic_consolidation",
			},
		})
	}

	return decisions
}

// decideStateTransition decides if a state transition should occur
func (uo *UnifiedOrchestrator) decideStateTransition(currentState string) *OrchestratorDecision {
	// Simple heuristic: transition based on cycle count
	switch currentState {
	case "idle":
		if uo.cycleCount > 5 {
			return &OrchestratorDecision{
				Type:     DecisionStateTransition,
				Priority: 0.9,
				Data: map[string]interface{}{
					"trigger": TriggerWake,
				},
			}
		}
	case "awake":
		if uo.cycleCount%50 == 0 {
			return &OrchestratorDecision{
				Type:     DecisionStateTransition,
				Priority: 0.6,
				Data: map[string]interface{}{
					"trigger": TriggerPerceive,
				},
			}
		}
	}

	return nil
}

// executeDecisions executes the orchestrator decisions
func (uo *UnifiedOrchestrator) executeDecisions(decisions []OrchestratorDecision) {
	for _, decision := range decisions {
		uo.totalDecisions++

		switch decision.Type {
		case DecisionStateTransition:
			if trigger, ok := decision.Data["trigger"].(CognitiveTrigger); ok {
				if uo.stateMachine != nil && uo.stateMachine.CanFire(trigger) {
					uo.stateMachine.Fire(trigger)
				}
			}

		case DecisionLearnPattern:
			if uo.patternProcessor != nil {
				// Learn a simple pattern based on current state
				features := []float64{float64(uo.cycleCount % 100) / 100.0}
				uo.patternProcessor.LearnPattern("orchestration", "cycle_pattern", features)
			}

		case DecisionConsolidate:
			// Trigger knowledge consolidation
			uo.consolidateKnowledge()

		case DecisionEmit:
			if eventType, ok := decision.Data["event_type"].(CognitiveEventType); ok {
				uo.eventBus.Publish(NewCognitiveEvent(
					eventType,
					"unified_orchestrator",
					decision.Data,
				))
			}
		}
	}
}

// updateOrchestratorMode updates the orchestration mode based on activity
func (uo *UnifiedOrchestrator) updateOrchestratorMode() {
	// Simple mode transitions based on cycle count
	switch {
	case uo.cycleCount < 10:
		uo.orchestrationMode = ModeIdle
	case uo.cycleCount%100 < 80:
		uo.orchestrationMode = ModeActive
	case uo.cycleCount%100 < 90:
		uo.orchestrationMode = ModeReflective
	default:
		uo.orchestrationMode = ModeConsolidating
	}
}

// persistState saves current state to the state store
func (uo *UnifiedOrchestrator) persistState() {
	if uo.stateStore == nil {
		return
	}

	state := map[string]interface{}{
		"mode":        uo.orchestrationMode.String(),
		"cycle_count": uo.cycleCount,
		"timestamp":   time.Now().Unix(),
	}

	uo.stateStore.Set("orchestrator:state", state)
}

// consolidateKnowledge performs knowledge consolidation
func (uo *UnifiedOrchestrator) consolidateKnowledge() {
	if uo.knowledgeGraph == nil {
		return
	}

	// Add consolidation record to knowledge graph
	uo.knowledgeGraph.AddQuad(
		"Orchestrator",
		"consolidated_at",
		time.Now().Format(time.RFC3339),
		"consolidation",
	)
}

// gestaltAggregationLoop periodically aggregates the gestalt
func (uo *UnifiedOrchestrator) gestaltAggregationLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-uo.ctx.Done():
			return
		case <-ticker.C:
			uo.aggregateGestalt()
		}
	}
}

// aggregateGestalt collects and aggregates the global gestalt
func (uo *UnifiedOrchestrator) aggregateGestalt() {
	uo.mu.Lock()
	defer uo.mu.Unlock()

	gestalt := make(map[string]interface{})

	// Orchestrator contribution
	gestalt["orchestrator"] = map[string]interface{}{
		"mode":           uo.orchestrationMode.String(),
		"cycle_count":    uo.cycleCount,
		"total_cycles":   uo.totalCycles,
		"total_events":   uo.totalEvents,
		"total_decisions": uo.totalDecisions,
	}

	// Knowledge graph contribution
	if uo.knowledgeGraph != nil {
		gestalt["knowledge_graph"] = uo.knowledgeGraph.ContributeToGestalt()
	}

	// State store contribution
	if uo.stateStore != nil {
		gestalt["state_store"] = uo.stateStore.ContributeToGestalt()
	}

	// State machine contribution
	if uo.stateMachine != nil {
		gestalt["state_machine"] = uo.stateMachine.ContributeToGestalt()
	}

	// Pattern processor contribution
	if uo.patternProcessor != nil {
		gestalt["pattern_processor"] = uo.patternProcessor.ContributeToGestalt()
	}

	// Semantic memory contribution
	if uo.semanticMemory != nil {
		gestalt["semantic_memory"] = uo.semanticMemory.ContributeToGestalt()
	}

	// Scheduler contribution
	if uo.scheduler != nil {
		gestalt["scheduler"] = uo.scheduler.ContributeToGestalt()
	}

	uo.gestaltCache = gestalt
	uo.gestaltTimestamp = time.Now()
}

// Event handlers

func (uo *UnifiedOrchestrator) handleStateTransition(event CognitiveEvent) {
	uo.mu.Lock()
	defer uo.mu.Unlock()
	uo.totalEvents++

	// Log state transition to knowledge graph
	if uo.knowledgeGraph != nil {
		if data, ok := event.Data.(map[string]interface{}); ok {
			if from, ok := data["from"].(string); ok {
				if to, ok := data["to"].(string); ok {
					uo.knowledgeGraph.AddQuad(
						from,
						"transitioned_to",
						to,
						"state_history",
					)
				}
			}
		}
	}
}

func (uo *UnifiedOrchestrator) handleEmergence(event CognitiveEvent) {
	uo.mu.Lock()
	defer uo.mu.Unlock()
	uo.totalEvents++

	// Switch to emergent mode
	uo.orchestrationMode = ModeEmergent

	// Record emergence in knowledge graph
	if uo.knowledgeGraph != nil {
		uo.knowledgeGraph.AddQuad(
			"Emergence",
			"detected_at",
			time.Now().Format(time.RFC3339),
			"emergence",
		)
	}
}

func (uo *UnifiedOrchestrator) handleKnowledgeAcquisition(event CognitiveEvent) {
	uo.mu.Lock()
	defer uo.mu.Unlock()
	uo.totalEvents++

	// Learn pattern from knowledge acquisition
	if uo.patternProcessor != nil {
		if data, ok := event.Data.(map[string]interface{}); ok {
			if subject, ok := data["subject"].(string); ok {
				features := []float64{float64(len(subject)) / 100.0}
				uo.patternProcessor.LearnPattern("knowledge", subject, features)
			}
		}
	}
}

func (uo *UnifiedOrchestrator) handlePhaseTransition(event CognitiveEvent) {
	uo.mu.Lock()
	defer uo.mu.Unlock()
	uo.totalEvents++

	// Update orchestrator mode based on phase
	if data, ok := event.Data.(map[string]interface{}); ok {
		if to, ok := data["to"].(string); ok {
			switch to {
			case "Active":
				uo.orchestrationMode = ModeActive
			case "Consolidating":
				uo.orchestrationMode = ModeConsolidating
			case "Resting":
				uo.orchestrationMode = ModeReflective
			}
		}
	}
}

func (uo *UnifiedOrchestrator) handleThoughtGenerated(event CognitiveEvent) {
	uo.mu.Lock()
	defer uo.mu.Unlock()
	uo.totalEvents++

	// Store thought in semantic memory
	if uo.semanticMemory != nil {
		if data, ok := event.Data.(map[string]interface{}); ok {
			if thoughtType, ok := data["type"].(string); ok {
				uo.semanticMemory.AddDocument("episodic", thoughtType, map[string]string{
					"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
					"source":    event.Source,
				})
			}
		}
	}
}

// OrchestratorDecision represents a decision made by the orchestrator
type OrchestratorDecision struct {
	Type     DecisionType
	Priority float64
	Data     map[string]interface{}
}

// DecisionType represents the type of orchestrator decision
type DecisionType int

const (
	DecisionStateTransition DecisionType = iota
	DecisionLearnPattern
	DecisionConsolidate
	DecisionEmit
	DecisionSchedule
)

// GetGestalt returns the current aggregated gestalt
func (uo *UnifiedOrchestrator) GetGestalt() map[string]interface{} {
	uo.mu.RLock()
	defer uo.mu.RUnlock()
	return uo.gestaltCache
}

// GetMetrics returns orchestrator metrics
func (uo *UnifiedOrchestrator) GetMetrics() map[string]interface{} {
	uo.mu.RLock()
	defer uo.mu.RUnlock()

	return map[string]interface{}{
		"running":         uo.running,
		"mode":            uo.orchestrationMode.String(),
		"cycle_count":     uo.cycleCount,
		"total_cycles":    uo.totalCycles,
		"total_events":    uo.totalEvents,
		"total_decisions": uo.totalDecisions,
		"last_cycle_time": uo.lastCycleTime,
	}
}

// ContributeToGestalt provides orchestrator state for the global gestalt
func (uo *UnifiedOrchestrator) ContributeToGestalt() map[string]interface{} {
	return uo.GetMetrics()
}
