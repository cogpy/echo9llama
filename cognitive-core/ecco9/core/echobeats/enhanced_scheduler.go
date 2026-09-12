package echobeats

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EnhancedScheduler extends EchoBeats with 12-step cognitive loop and 3 inference engines
type EnhancedScheduler struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Original EchoBeats scheduler
	echoBeats *EchoBeats

	// New components: 3 concurrent inference engines
	engines []*InferenceEngine

	// 12-step cognitive loop (shared across engines)
	masterLoop *CognitiveLoop

	// Integration points
	wakeRestManager  interface{} // *deeptreeecho.AutonomousWakeRestManager
	goalOrchestrator interface{} // *deeptreeecho.GoalOrchestrator
	streamOfConsc    interface{} // *consciousness.StreamOfConsciousness
	dreamCycle       interface{} // *echodream.DreamCycleIntegration

	// Enhanced metrics
	loopCycles  uint64
	engineTasks uint64

	// Control
	running bool
}

// NewEnhancedScheduler creates an enhanced scheduler
func NewEnhancedScheduler() *EnhancedScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	es := &EnhancedScheduler{
		ctx:       ctx,
		cancel:    cancel,
		echoBeats: NewEchoBeats(),
		engines:   make([]*InferenceEngine, 0, 3),
	}

	// Create 3 concurrent inference engines with different specializations
	es.engines = append(es.engines, NewInferenceEngine(1, SpecializationPerception))
	es.engines = append(es.engines, NewInferenceEngine(2, SpecializationCognition))
	es.engines = append(es.engines, NewInferenceEngine(3, SpecializationAction))

	// Create master 12-step cognitive loop
	es.masterLoop = NewCognitiveLoop()

	// Set up callbacks to coordinate systems
	es.setupCallbacks()

	return es
}

// setupCallbacks configures coordination between components
func (es *EnhancedScheduler) setupCallbacks() {
	// Cognitive loop callbacks
	es.masterLoop.SetCallbacks(
		func(step int, result *StepResult) {
			// On step complete, potentially generate events
			es.onCognitiveStepComplete(step, result)
		},
		func(cycle uint64) {
			// On cycle complete
			es.mu.Lock()
			es.loopCycles++
			es.mu.Unlock()

			fmt.Printf("🔄 Enhanced Scheduler: Cognitive cycle %d complete\n", cycle)
		},
	)

	// Register EchoBeats handlers that route to inference engines
	es.registerEnhancedHandlers()
}

// registerEnhancedHandlers registers handlers that use inference engines
func (es *EnhancedScheduler) registerEnhancedHandlers() {
	// Thought generation handler - route to perception engine
	es.echoBeats.RegisterHandler(EventThought, func(event *CognitiveEvent) error {
		task := &InferenceTask{
			ID:       event.ID,
			Type:     "thought_generation",
			Input:    event.Payload,
			Priority: float64(event.Priority) / 100.0,
			Context:  event.Context,
		}
		return es.engines[0].SubmitTask(task)
	})

	// Goal pursuit handler - route to action engine
	es.echoBeats.RegisterHandler(EventGoalPursuit, func(event *CognitiveEvent) error {
		task := &InferenceTask{
			ID:       event.ID,
			Type:     "goal_pursuit",
			Input:    event.Payload,
			Priority: float64(event.Priority) / 100.0,
			Context:  event.Context,
		}
		return es.engines[2].SubmitTask(task)
	})

	// Introspection handler - route to cognition engine
	es.echoBeats.RegisterHandler(EventIntrospection, func(event *CognitiveEvent) error {
		task := &InferenceTask{
			ID:       event.ID,
			Type:     "introspection",
			Input:    event.Payload,
			Priority: float64(event.Priority) / 100.0,
			Context:  event.Context,
		}
		return es.engines[1].SubmitTask(task)
	})

	// Learning handler - route to cognition engine
	es.echoBeats.RegisterHandler(EventLearning, func(event *CognitiveEvent) error {
		task := &InferenceTask{
			ID:       event.ID,
			Type:     "learning",
			Input:    event.Payload,
			Priority: float64(event.Priority) / 100.0,
			Context:  event.Context,
		}
		return es.engines[1].SubmitTask(task)
	})
}

// onCognitiveStepComplete handles cognitive loop step completion
func (es *EnhancedScheduler) onCognitiveStepComplete(step int, result *StepResult) {
	if result == nil || !result.Success {
		return
	}

	// Generate insights can trigger new events
	if len(result.Insights) > 0 {
		for _, insight := range result.Insights {
			es.echoBeats.ScheduleEvent(&CognitiveEvent{
				ID:          fmt.Sprintf("insight_%d_%d", step, time.Now().UnixNano()),
				Type:        EventIntrospection,
				Priority:    70,
				ScheduledAt: time.Now().Add(1 * time.Second),
				Payload:     insight,
				Context: map[string]interface{}{
					"source_step": step,
					"from_loop":   true,
				},
			})
		}
	}

	// High cognitive load can trigger rest
	if result.CognitiveLoad > 0.8 {
		es.echoBeats.ScheduleEvent(&CognitiveEvent{
			ID:          fmt.Sprintf("rest_trigger_%d", time.Now().UnixNano()),
			Type:        EventRest,
			Priority:    85,
			ScheduledAt: time.Now().Add(5 * time.Second),
			Payload:     "High cognitive load detected",
		})
	}
}

// Start begins the enhanced scheduler
func (es *EnhancedScheduler) Start() error {
	es.mu.Lock()
	if es.running {
		es.mu.Unlock()
		return fmt.Errorf("enhanced scheduler already running")
	}
	es.running = true
	es.mu.Unlock()

	fmt.Println("🎵 Enhanced EchoBeats Scheduler: Starting...")
	fmt.Println("   Components:")
	fmt.Println("   • Original EchoBeats event scheduler")
	fmt.Println("   • 3 concurrent inference engines")
	fmt.Println("   • 12-step cognitive loop")

	// Start original EchoBeats
	if err := es.echoBeats.Start(); err != nil {
		return fmt.Errorf("failed to start EchoBeats: %w", err)
	}

	// Start all inference engines
	for _, engine := range es.engines {
		if err := engine.Start(); err != nil {
			return fmt.Errorf("failed to start inference engine: %w", err)
		}
	}

	// Start master cognitive loop
	if err := es.masterLoop.Start(); err != nil {
		return fmt.Errorf("failed to start cognitive loop: %w", err)
	}

	fmt.Println("🎵 Enhanced EchoBeats Scheduler: All systems operational!")

	return nil
}

// Stop gracefully stops the enhanced scheduler
func (es *EnhancedScheduler) Stop() error {
	es.mu.Lock()
	defer es.mu.Unlock()

	if !es.running {
		return fmt.Errorf("enhanced scheduler not running")
	}

	fmt.Println("🎵 Enhanced EchoBeats Scheduler: Stopping...")
	es.running = false
	es.cancel()

	// Stop cognitive loop
	if err := es.masterLoop.Stop(); err != nil {
		fmt.Printf("⚠️  Error stopping cognitive loop: %v\n", err)
	}

	// Stop all inference engines
	for _, engine := range es.engines {
		if err := engine.Stop(); err != nil {
			fmt.Printf("⚠️  Error stopping inference engine: %v\n", err)
		}
	}

	// Stop original EchoBeats
	if err := es.echoBeats.Stop(); err != nil {
		fmt.Printf("⚠️  Error stopping EchoBeats: %v\n", err)
	}

	return nil
}

// ScheduleEvent schedules an event through EchoBeats
func (es *EnhancedScheduler) ScheduleEvent(event *CognitiveEvent) {
	es.echoBeats.ScheduleEvent(event)
}

// SetWakeRestManager sets the wake/rest manager
func (es *EnhancedScheduler) SetWakeRestManager(manager interface{}) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.wakeRestManager = manager
}

// SetGoalOrchestrator sets the goal orchestrator
func (es *EnhancedScheduler) SetGoalOrchestrator(orchestrator interface{}) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.goalOrchestrator = orchestrator
}

// SetStreamOfConsciousness sets the stream-of-consciousness
func (es *EnhancedScheduler) SetStreamOfConsciousness(soc interface{}) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.streamOfConsc = soc
}

// SetDreamCycle sets the dream cycle integration
func (es *EnhancedScheduler) SetDreamCycle(dc interface{}) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.dreamCycle = dc
}

// GetStatus returns comprehensive status
func (es *EnhancedScheduler) GetStatus() map[string]interface{} {
	es.mu.RLock()
	defer es.mu.RUnlock()

	// Get EchoBeats status
	echoBeatsStatus := es.echoBeats.GetStatus()

	// Get cognitive loop metrics
	loopMetrics := es.masterLoop.GetMetrics()

	// Get engine metrics
	engineMetrics := make([]map[string]interface{}, len(es.engines))
	for i, engine := range es.engines {
		engineMetrics[i] = engine.GetMetrics()
	}

	return map[string]interface{}{
		"running":           es.running,
		"loop_cycles":       es.loopCycles,
		"engine_tasks":      es.engineTasks,
		"echobeats":         echoBeatsStatus,
		"cognitive_loop":    loopMetrics,
		"inference_engines": engineMetrics,
	}
}

// GetCognitiveState returns current cognitive state
func (es *EnhancedScheduler) GetCognitiveState() *CognitiveState {
	return es.masterLoop.GetCurrentState()
}

// GetEchoBeats returns the underlying EchoBeats scheduler
func (es *EnhancedScheduler) GetEchoBeats() *EchoBeats {
	return es.echoBeats
}
