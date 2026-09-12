package integration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EchoCog/echollama/core/consciousness"
	"github.com/EchoCog/echollama/core/echobeats"
	"github.com/EchoCog/echollama/core/echodream"
	"github.com/EchoCog/echollama/core/goals"
	"github.com/google/uuid"
)

// CognitiveEventLoopOrchestrator unifies consciousness, scheduling, and goal pursuit
// into a single persistent cognitive loop with bidirectional influence
type CognitiveEventLoopOrchestrator struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Core components
	consciousness    *consciousness.StreamOfConsciousnessLLM
	scheduler        *echobeats.EchoBeats
	dreamSystem      *echodream.EchoDream
	goalOrchestrator *goals.GoalOrchestrator

	// Event translation
	thoughtToEvent map[consciousness.ThoughtType]echobeats.EventType
	eventToThought map[echobeats.EventType]consciousness.ThoughtType

	// Cognitive state tracking
	currentFocus   string
	cognitiveLoad  float64
	fatigueLevel   float64
	awarenessLevel float64

	// Metrics
	eventsTriggered   uint64
	thoughtsTriggered uint64
	goalsGenerated    uint64
	cyclesCompleted   uint64

	running bool
}

// NewCognitiveEventLoopOrchestrator creates a new orchestrator
func NewCognitiveEventLoopOrchestrator(
	consciousness *consciousness.StreamOfConsciousnessLLM,
	scheduler *echobeats.EchoBeats,
	dreamSystem *echodream.EchoDream,
	goalOrchestrator *goals.GoalOrchestrator,
) *CognitiveEventLoopOrchestrator {
	ctx, cancel := context.WithCancel(context.Background())

	orchestrator := &CognitiveEventLoopOrchestrator{
		ctx:              ctx,
		cancel:           cancel,
		consciousness:    consciousness,
		scheduler:        scheduler,
		dreamSystem:      dreamSystem,
		goalOrchestrator: goalOrchestrator,
		cognitiveLoad:    0.3,
		fatigueLevel:     0.0,
		awarenessLevel:   0.7,
	}

	// Initialize thought-to-event mappings
	orchestrator.initializeMappings()

	return orchestrator
}

// initializeMappings sets up bidirectional mappings between thoughts and events
func (celo *CognitiveEventLoopOrchestrator) initializeMappings() {
	celo.thoughtToEvent = map[consciousness.ThoughtType]echobeats.EventType{
		consciousness.ThoughtTypeReflection:    echobeats.EventIntrospection,
		consciousness.ThoughtTypeQuestion:      echobeats.EventLearning,
		consciousness.ThoughtTypeInsight:       echobeats.EventMemoryConsolidation,
		consciousness.ThoughtTypePlanning:      echobeats.EventGoalPursuit,
		consciousness.ThoughtTypeMetaCognition: echobeats.EventIntrospection,
		consciousness.ThoughtTypePerception:    echobeats.EventPerception,
	}

	celo.eventToThought = map[echobeats.EventType]consciousness.ThoughtType{
		echobeats.EventIntrospection:       consciousness.ThoughtTypeReflection,
		echobeats.EventLearning:            consciousness.ThoughtTypeQuestion,
		echobeats.EventMemoryConsolidation: consciousness.ThoughtTypeInsight,
		echobeats.EventGoalPursuit:         consciousness.ThoughtTypePlanning,
		echobeats.EventPerception:          consciousness.ThoughtTypePerception,
	}
}

// Start begins the orchestrated cognitive loop
func (celo *CognitiveEventLoopOrchestrator) Start() error {
	celo.mu.Lock()
	if celo.running {
		celo.mu.Unlock()
		return fmt.Errorf("cognitive event loop orchestrator already running")
	}
	celo.running = true
	celo.mu.Unlock()

	// Start thought-to-event translation loop
	go celo.thoughtToEventLoop()

	// Start event-to-thought translation loop
	go celo.eventToThoughtLoop()

	// Start cognitive state monitoring loop
	go celo.cognitiveStateLoop()

	// Start goal-driven event generation loop
	go celo.goalDrivenEventLoop()

	// Start autonomous cognitive cycle loop
	go celo.autonomousCycleLoop()

	return nil
}

// Stop halts the orchestrated cognitive loop
func (celo *CognitiveEventLoopOrchestrator) Stop() {
	celo.mu.Lock()
	celo.running = false
	celo.mu.Unlock()

	celo.cancel()
}

// thoughtToEventLoop translates thoughts into scheduled events
func (celo *CognitiveEventLoopOrchestrator) thoughtToEventLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-celo.ctx.Done():
			return
		case <-ticker.C:
			celo.translateThoughtsToEvents()
		}
	}
}

// translateThoughtsToEvents converts recent thoughts into cognitive events
func (celo *CognitiveEventLoopOrchestrator) translateThoughtsToEvents() {
	// Get recent thoughts
	thoughts := celo.consciousness.GetRecentThoughts(3)

	for _, thought := range thoughts {
		// Check if this thought should trigger an event
		if celo.shouldTriggerEvent(thought) {
			event := celo.createEventFromThought(thought)
			if event != nil {
				// Schedule the event
				celo.scheduler.ScheduleEvent(event)

				celo.mu.Lock()
				celo.eventsTriggered++
				celo.mu.Unlock()
			}
		}
	}
}

// shouldTriggerEvent determines if a thought should create an event
func (celo *CognitiveEventLoopOrchestrator) shouldTriggerEvent(thought interface{}) bool {
	// Check thought type, confidence, and current cognitive state
	// For now, trigger events for high-confidence thoughts
	return true // Simplified
}

// createEventFromThought creates a cognitive event from a thought
func (celo *CognitiveEventLoopOrchestrator) createEventFromThought(thought interface{}) *echobeats.CognitiveEvent {
	// In real implementation, would extract thought type and map to event type
	// For now, create a generic thought event

	event := &echobeats.CognitiveEvent{
		ID:          uuid.New().String(),
		Type:        echobeats.EventThought,
		Priority:    5,
		Timestamp:   time.Now(),
		ScheduledAt: time.Now().Add(1 * time.Second),
		Payload:     thought,
		Context: map[string]interface{}{
			"source": "consciousness",
		},
		Recurring: false,
	}

	return event
}

// eventToThoughtLoop translates scheduled events into thoughts
func (celo *CognitiveEventLoopOrchestrator) eventToThoughtLoop() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-celo.ctx.Done():
			return
		case <-ticker.C:
			celo.translateEventsToThoughts()
		}
	}
}

// translateEventsToThoughts converts events into injected thoughts
func (celo *CognitiveEventLoopOrchestrator) translateEventsToThoughts() {
	// Get pending events from scheduler
	// In real implementation, would query scheduler for upcoming events

	// For now, periodically inject goal-related thoughts
	goals := celo.goalOrchestrator.GetActiveGoals()
	if len(goals) > 0 {
		goal := goals[0]
		thoughtContent := fmt.Sprintf("Pursuing goal: %s", goal.Title)
		celo.consciousness.AddExternalThought(thoughtContent)

		celo.mu.Lock()
		celo.thoughtsTriggered++
		celo.mu.Unlock()
	}
}

// cognitiveStateLoop monitors and updates cognitive state
func (celo *CognitiveEventLoopOrchestrator) cognitiveStateLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-celo.ctx.Done():
			return
		case <-ticker.C:
			celo.updateCognitiveState()
		}
	}
}

// updateCognitiveState updates cognitive load, fatigue, and awareness
func (celo *CognitiveEventLoopOrchestrator) updateCognitiveState() {
	celo.mu.Lock()
	defer celo.mu.Unlock()

	// Increase fatigue based on cognitive load
	celo.fatigueLevel += celo.cognitiveLoad * 0.01

	// Decrease awareness if fatigued
	if celo.fatigueLevel > 0.7 {
		celo.awarenessLevel -= 0.01
	}

	// Clamp values
	celo.fatigueLevel = clamp(celo.fatigueLevel, 0.0, 1.0)
	celo.awarenessLevel = clamp(celo.awarenessLevel, 0.0, 1.0)

	// If fatigue is high, trigger rest cycle
	if celo.fatigueLevel > 0.8 && celo.awarenessLevel < 0.5 {
		celo.triggerRestCycle()
	}
}

// triggerRestCycle initiates a rest/dream cycle
func (celo *CognitiveEventLoopOrchestrator) triggerRestCycle() {
	// Create rest event
	event := &echobeats.CognitiveEvent{
		ID:          uuid.New().String(),
		Type:        echobeats.EventRest,
		Priority:    10,
		Timestamp:   time.Now(),
		ScheduledAt: time.Now(),
		Context: map[string]interface{}{
			"fatigue_level": celo.fatigueLevel,
			"reason":        "autonomous_fatigue_management",
		},
	}

	celo.scheduler.ScheduleEvent(event)

	// Reset fatigue after rest
	celo.fatigueLevel = 0.0
	celo.awarenessLevel = 0.7
}

// goalDrivenEventLoop generates events based on active goals
func (celo *CognitiveEventLoopOrchestrator) goalDrivenEventLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-celo.ctx.Done():
			return
		case <-ticker.C:
			celo.generateGoalDrivenEvents()
		}
	}
}

// generateGoalDrivenEvents creates events to pursue active goals
func (celo *CognitiveEventLoopOrchestrator) generateGoalDrivenEvents() {
	goals := celo.goalOrchestrator.GetActiveGoals()

	for _, goal := range goals {
		// Create goal pursuit event
		event := &echobeats.CognitiveEvent{
			ID:          uuid.New().String(),
			Type:        echobeats.EventGoalPursuit,
			Priority:    goal.Priority,
			Timestamp:   time.Now(),
			ScheduledAt: time.Now().Add(5 * time.Second),
			Payload:     goal,
			Context: map[string]interface{}{
				"goal_id":   goal.ID,
				"goal_name": goal.Title,
			},
		}

		celo.scheduler.ScheduleEvent(event)
	}
}

// autonomousCycleLoop implements the 12-step cognitive cycle
func (celo *CognitiveEventLoopOrchestrator) autonomousCycleLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-celo.ctx.Done():
			return
		case <-ticker.C:
			celo.executeAutonomousCycle()
		}
	}
}

// executeAutonomousCycle runs one complete cognitive cycle
func (celo *CognitiveEventLoopOrchestrator) executeAutonomousCycle() {
	// 12-step cognitive cycle based on relevance realization
	// Phase 1: Orienting (Steps 1-4)
	celo.orientingPhase()

	// Phase 2: Conditioning (Steps 5-8)
	celo.conditioningPhase()

	// Phase 3: Anticipating (Steps 9-12)
	celo.anticipatingPhase()

	celo.mu.Lock()
	celo.cyclesCompleted++
	celo.mu.Unlock()
}

// orientingPhase: Relevance realization and present commitment
func (celo *CognitiveEventLoopOrchestrator) orientingPhase() {
	// Step 1: Relevance realization - What matters now?
	celo.consciousness.AddExternalThought("What is most relevant to my current goals?")

	// Steps 2-4: Orient to present context
	time.Sleep(1 * time.Second)
}

// conditioningPhase: Actual affordance interaction and past performance
func (celo *CognitiveEventLoopOrchestrator) conditioningPhase() {
	// Steps 5-8: Interact with actual affordances based on past learning
	celo.consciousness.AddExternalThought("What have I learned that applies here?")
	time.Sleep(1 * time.Second)
}

// anticipatingPhase: Virtual salience simulation and future potential
func (celo *CognitiveEventLoopOrchestrator) anticipatingPhase() {
	// Step 9: Relevance realization - What could matter?
	celo.consciousness.AddExternalThought("What possibilities should I explore?")

	// Steps 10-12: Simulate future scenarios
	time.Sleep(1 * time.Second)
}

// GetMetrics returns orchestrator metrics
func (celo *CognitiveEventLoopOrchestrator) GetMetrics() map[string]interface{} {
	celo.mu.RLock()
	defer celo.mu.RUnlock()

	return map[string]interface{}{
		"events_triggered":   celo.eventsTriggered,
		"thoughts_triggered": celo.thoughtsTriggered,
		"goals_generated":    celo.goalsGenerated,
		"cycles_completed":   celo.cyclesCompleted,
		"cognitive_load":     celo.cognitiveLoad,
		"fatigue_level":      celo.fatigueLevel,
		"awareness_level":    celo.awarenessLevel,
	}
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
