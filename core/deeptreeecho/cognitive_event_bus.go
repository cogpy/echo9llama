package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CognitiveEventType represents the type of cognitive event
type CognitiveEventType string

const (
	// Thought and consciousness events
	EventThoughtGenerated      CognitiveEventType = "thought_generated"
	EventInsightGained         CognitiveEventType = "insight_gained"
	EventPatternRecognized     CognitiveEventType = "pattern_recognized"
	EventKnowledgeGapDetected  CognitiveEventType = "knowledge_gap_detected"
	
	// Goal and planning events
	EventGoalCreated           CognitiveEventType = "goal_created"
	EventGoalAchieved          CognitiveEventType = "goal_achieved"
	EventGoalFailed            CognitiveEventType = "goal_failed"
	EventGoalUpdated           CognitiveEventType = "goal_updated"
	
	// Skill and learning events
	EventSkillLearned          CognitiveEventType = "skill_learned"
	EventSkillPracticed        CognitiveEventType = "skill_practiced"
	EventSkillMastered         CognitiveEventType = "skill_mastered"
	EventSkillAcquisitionGoal  CognitiveEventType = "skill_acquisition_goal"
	
	// Conversation and social events
	EventConversationDetected  CognitiveEventType = "conversation_detected"
	EventConversationEngaged   CognitiveEventType = "conversation_engaged"
	EventConversationEnded     CognitiveEventType = "conversation_ended"
	EventMessageReceived       CognitiveEventType = "message_received"
	
	// Wisdom and synthesis events
	EventWisdomPrincipleCreated CognitiveEventType = "wisdom_principle_created"
	EventWisdomEvolved         CognitiveEventType = "wisdom_evolved"
	EventWisdomApplied         CognitiveEventType = "wisdom_applied"
	
	// State transition events
	EventStateTransition       CognitiveEventType = "state_transition"
	EventWakeInitiated         CognitiveEventType = "wake_initiated"
	EventRestInitiated         CognitiveEventType = "rest_initiated"
	EventDreamStarted          CognitiveEventType = "dream_started"
	EventDreamEnded            CognitiveEventType = "dream_ended"
	
	// Heartbeat and monitoring events
	EventHeartbeatPulse        CognitiveEventType = "heartbeat_pulse"
	EventSelfInsight           CognitiveEventType = "self_insight"
	EventVitalSignsUpdate      CognitiveEventType = "vital_signs_update"
	
	// Interest and attention events
	EventInterestDetected      CognitiveEventType = "interest_detected"
	EventInterestEmerged       CognitiveEventType = "interest_emerged"
	EventAttentionShift        CognitiveEventType = "attention_shift"
	EventTopicEmergence        CognitiveEventType = "topic_emergence"
	
	// Memory and consolidation events
	EventMemoryConsolidated    CognitiveEventType = "memory_consolidated"
	EventMemoryRecalled        CognitiveEventType = "memory_recalled"
	EventMemoryPruned          CognitiveEventType = "memory_pruned"
	
	// EchoBeats scheduler events
	EventPhaseTransition       CognitiveEventType = "phase_transition"
	EventCycleCompleted        CognitiveEventType = "cycle_completed"
	EventTriadSynchronized     CognitiveEventType = "triad_synchronized"
	EventEmergenceDetected     CognitiveEventType = "emergence_detected"
	
	// Knowledge and wisdom events
	EventKnowledgeGapIdentified CognitiveEventType = "knowledge_gap_identified"
	EventKnowledgeAcquired      CognitiveEventType = "knowledge_acquired"
	EventWisdomGained          CognitiveEventType = "wisdom_gained"
	
	// Self-update events
	EventUpdateAvailable       CognitiveEventType = "update_available"
	EventUpdateApplied         CognitiveEventType = "update_applied"
)

// CognitiveEvent represents a single event in the cognitive system
type CognitiveEvent struct {
	Type      CognitiveEventType
	Timestamp time.Time
	Source    string
	Data      interface{} // Can be any type - map, struct, string, etc.
	Priority  float64 // 0.0 = normal, higher = more important
}

// NewCognitiveEvent creates a new cognitive event
func NewCognitiveEvent(eventType CognitiveEventType, source string, data map[string]interface{}) CognitiveEvent {
	return CognitiveEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Source:    source,
		Data:      data,
		Priority:  0,
	}
}

// NewPriorityCognitiveEvent creates a new cognitive event with priority
func NewPriorityCognitiveEvent(eventType CognitiveEventType, source string, data map[string]interface{}, priority float64) CognitiveEvent {
	return CognitiveEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Source:    source,
		Data:      data,
		Priority:  priority,
	}
}

// EventHandler is a function that handles cognitive events
type EventHandler func(event CognitiveEvent)

// CognitiveEventBus manages event distribution across cognitive subsystems
type CognitiveEventBus struct {
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	
	// Event channels
	eventQueue   chan CognitiveEvent
	
	// Subscribers
	subscribers  map[CognitiveEventType][]EventHandler
	globalSubs   []EventHandler // Handlers that receive all events
	
	// Metrics
	totalEvents  uint64
	eventCounts  map[CognitiveEventType]uint64
	
	// Event history (limited size for debugging)
	eventHistory []CognitiveEvent
	maxHistory   int
	
	// Running state
	running      bool
}

// NewCognitiveEventBus creates a new cognitive event bus
func NewCognitiveEventBus(ctx context.Context) *CognitiveEventBus {
	busCtx, cancel := context.WithCancel(ctx)
	
	bus := &CognitiveEventBus{
		ctx:          busCtx,
		cancel:       cancel,
		eventQueue:   make(chan CognitiveEvent, 1000), // Buffer for 1000 events
		subscribers:  make(map[CognitiveEventType][]EventHandler),
		globalSubs:   make([]EventHandler, 0),
		eventCounts:  make(map[CognitiveEventType]uint64),
		eventHistory: make([]CognitiveEvent, 0),
		maxHistory:   100, // Keep last 100 events
		running:      false,
	}
	
	return bus
}

// Start begins processing events
func (bus *CognitiveEventBus) Start() error {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	
	if bus.running {
		return fmt.Errorf("event bus already running")
	}
	
	bus.running = true
	
	// Start event processing goroutine
	go bus.processEvents()
	
	return nil
}

// Stop stops the event bus
func (bus *CognitiveEventBus) Stop() error {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	
	if !bus.running {
		return fmt.Errorf("event bus not running")
	}
	
	bus.running = false
	bus.cancel()
	close(bus.eventQueue)
	
	return nil
}

// Subscribe registers a handler for a specific event type
func (bus *CognitiveEventBus) Subscribe(eventType CognitiveEventType, handler EventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	
	if bus.subscribers[eventType] == nil {
		bus.subscribers[eventType] = make([]EventHandler, 0)
	}
	
	bus.subscribers[eventType] = append(bus.subscribers[eventType], handler)
}

// SubscribeAll registers a handler for all events
func (bus *CognitiveEventBus) SubscribeAll(handler EventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	
	bus.globalSubs = append(bus.globalSubs, handler)
}

// Publish sends an event to all subscribers
func (bus *CognitiveEventBus) Publish(event CognitiveEvent) {
	select {
	case bus.eventQueue <- event:
		// Event queued successfully
	case <-bus.ctx.Done():
		// Context cancelled, ignore event
	default:
		// Queue full, log warning (in production would use proper logging)
		fmt.Printf("⚠️  Event queue full, dropping event: %s\n", event.Type)
	}
}

// PublishSync publishes an event and waits for all handlers to complete
func (bus *CognitiveEventBus) PublishSync(event CognitiveEvent) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	
	// Update metrics
	bus.totalEvents++
	bus.eventCounts[event.Type]++
	
	// Add to history
	bus.addToHistory(event)
	
	// Call global subscribers
	for _, handler := range bus.globalSubs {
		handler(event)
	}
	
	// Call type-specific subscribers
	if handlers, ok := bus.subscribers[event.Type]; ok {
		for _, handler := range handlers {
			handler(event)
		}
	}
}

// processEvents is the main event processing loop
func (bus *CognitiveEventBus) processEvents() {
	for {
		select {
		case event, ok := <-bus.eventQueue:
			if !ok {
				// Channel closed, exit
				return
			}
			
			// Process event
			bus.handleEvent(event)
			
		case <-bus.ctx.Done():
			// Context cancelled, exit
			return
		}
	}
}

// handleEvent processes a single event
func (bus *CognitiveEventBus) handleEvent(event CognitiveEvent) {
	bus.mu.Lock()
	
	// Update metrics
	bus.totalEvents++
	bus.eventCounts[event.Type]++
	
	// Add to history
	bus.addToHistory(event)
	
	// Get handlers (copy to avoid holding lock during execution)
	globalHandlers := make([]EventHandler, len(bus.globalSubs))
	copy(globalHandlers, bus.globalSubs)
	
	var typeHandlers []EventHandler
	if handlers, ok := bus.subscribers[event.Type]; ok {
		typeHandlers = make([]EventHandler, len(handlers))
		copy(typeHandlers, handlers)
	}
	
	bus.mu.Unlock()
	
	// Execute handlers (without holding lock)
	for _, handler := range globalHandlers {
		safeExecuteHandler(handler, event)
	}
	
	for _, handler := range typeHandlers {
		safeExecuteHandler(handler, event)
	}
}

// safeExecuteHandler executes a handler with panic recovery
func safeExecuteHandler(handler EventHandler, event CognitiveEvent) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("⚠️  Event handler panic for %s: %v\n", event.Type, r)
		}
	}()
	
	handler(event)
}

// addToHistory adds an event to the history buffer
func (bus *CognitiveEventBus) addToHistory(event CognitiveEvent) {
	bus.eventHistory = append(bus.eventHistory, event)
	
	// Trim history if too large
	if len(bus.eventHistory) > bus.maxHistory {
		bus.eventHistory = bus.eventHistory[len(bus.eventHistory)-bus.maxHistory:]
	}
}

// GetMetrics returns event bus metrics
func (bus *CognitiveEventBus) GetMetrics() map[string]interface{} {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	
	// Copy event counts
	eventCounts := make(map[string]uint64)
	for eventType, count := range bus.eventCounts {
		eventCounts[string(eventType)] = count
	}
	
	return map[string]interface{}{
		"total_events":     bus.totalEvents,
		"event_counts":     eventCounts,
		"queue_size":       len(bus.eventQueue),
		"subscriber_count": len(bus.subscribers),
		"global_sub_count": len(bus.globalSubs),
		"running":          bus.running,
	}
}

// GetRecentEvents returns the most recent events
func (bus *CognitiveEventBus) GetRecentEvents(count int) []CognitiveEvent {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	
	if count > len(bus.eventHistory) {
		count = len(bus.eventHistory)
	}
	
	start := len(bus.eventHistory) - count
	events := make([]CognitiveEvent, count)
	copy(events, bus.eventHistory[start:])
	
	return events
}

// GetEventCount returns the count for a specific event type
func (bus *CognitiveEventBus) GetEventCount(eventType CognitiveEventType) uint64 {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	
	return bus.eventCounts[eventType]
}

// ClearHistory clears the event history
func (bus *CognitiveEventBus) ClearHistory() {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	
	bus.eventHistory = make([]CognitiveEvent, 0)
}
