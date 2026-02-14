package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CognitiveEventBusV3 is the central nervous system of the daechon.
// It replaces timer-based polling with event-driven cognitive processing.
// All subsystems publish and subscribe to cognitive events through this bus.
type CognitiveEventBusV3 struct {
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc

	// Subscribers keyed by event category
	subscribers map[CogEventCategory][]CogEventHandler

	// Event queue
	eventQueue  chan CogEvent

	// Activity feed - the public-facing log of cognitive activity
	activityFeed chan ActivityEntry

	// Metrics
	totalEvents    uint64
	eventsByType   map[CogEventCategory]uint64

	// Running state
	running bool
}

// CogEventCategory categorizes cognitive events
type CogEventCategory int

const (
	CogEventThought       CogEventCategory = iota // Internal thought generated
	CogEventEmotion                                // Emotional state change
	CogEventGoal                                   // Goal created/completed/failed
	CogEventMemory                                 // Memory stored/retrieved
	CogEventDream                                  // Dream cycle event
	CogEventWakeRest                               // Wake/rest transition
	CogEventConversation                           // Conversation event
	CogEventSkill                                  // Skill learning event
	CogEventIntrospection                          // Self-reflection event
	CogEventEmergence                              // Emergent pattern detected
	CogEventPIENN                                  // PIE-NN language event
	CogEventScheduler                              // Echobeats scheduler event
	CogEventDisposition                            // Disposition/mood change
	CogEventSystem                                 // System-level event
)

func (c CogEventCategory) String() string {
	return [...]string{
		"THOUGHT", "EMOTION", "GOAL", "MEMORY", "DREAM",
		"WAKE_REST", "CONVERSATION", "SKILL", "INTROSPECTION",
		"EMERGENCE", "PIENN", "SCHEDULER", "DISPOSITION", "SYSTEM",
	}[c]
}

// CogEvent represents a cognitive event flowing through the bus
type CogEvent struct {
	ID        string
	Category  CogEventCategory
	Source    string
	Content   string
	Priority  float64 // 0.0-1.0
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// CogEventHandler is a function that handles cognitive events
type CogEventHandler func(event CogEvent)

// ActivityEntry is a formatted entry for the activity feed console
type ActivityEntry struct {
	Timestamp time.Time
	Category  CogEventCategory
	Source    string
	Content   string
	Priority  float64
}

// NewCognitiveEventBusV3 creates a new event bus
func NewCognitiveEventBusV3() *CognitiveEventBusV3 {
	ctx, cancel := context.WithCancel(context.Background())
	return &CognitiveEventBusV3{
		ctx:          ctx,
		cancel:       cancel,
		subscribers:  make(map[CogEventCategory][]CogEventHandler),
		eventQueue:   make(chan CogEvent, 1024),
		activityFeed: make(chan ActivityEntry, 512),
		eventsByType: make(map[CogEventCategory]uint64),
	}
}

// Start begins processing events
func (bus *CognitiveEventBusV3) Start() error {
	bus.mu.Lock()
	if bus.running {
		bus.mu.Unlock()
		return fmt.Errorf("event bus already running")
	}
	bus.running = true
	bus.mu.Unlock()

	go bus.processLoop()
	return nil
}

// Stop halts event processing
func (bus *CognitiveEventBusV3) Stop() {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.running = false
	bus.cancel()
}

// Publish sends an event to the bus
func (bus *CognitiveEventBusV3) Publish(event CogEvent) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.ID == "" {
		event.ID = fmt.Sprintf("evt_%d_%d", event.Timestamp.UnixNano(), event.Category)
	}

	select {
	case bus.eventQueue <- event:
	default:
		// Queue full — drop low priority events
		if event.Priority > 0.5 {
			// Force push by draining one
			select {
			case <-bus.eventQueue:
			default:
			}
			bus.eventQueue <- event
		}
	}
}

// Subscribe registers a handler for a specific event category
func (bus *CognitiveEventBusV3) Subscribe(category CogEventCategory, handler CogEventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.subscribers[category] = append(bus.subscribers[category], handler)
}

// SubscribeAll registers a handler for all event categories
func (bus *CognitiveEventBusV3) SubscribeAll(handler CogEventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	for cat := CogEventThought; cat <= CogEventSystem; cat++ {
		bus.subscribers[cat] = append(bus.subscribers[cat], handler)
	}
}

// ActivityFeed returns the channel for reading activity entries
func (bus *CognitiveEventBusV3) ActivityFeed() <-chan ActivityEntry {
	return bus.activityFeed
}

// GetMetrics returns event processing metrics
func (bus *CognitiveEventBusV3) GetMetrics() map[string]interface{} {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	metrics := map[string]interface{}{
		"total_events": bus.totalEvents,
		"running":      bus.running,
		"queue_size":   len(bus.eventQueue),
	}
	for cat, count := range bus.eventsByType {
		metrics[fmt.Sprintf("events_%s", cat.String())] = count
	}
	return metrics
}

// processLoop is the main event processing goroutine
func (bus *CognitiveEventBusV3) processLoop() {
	for {
		select {
		case <-bus.ctx.Done():
			return
		case event := <-bus.eventQueue:
			bus.dispatch(event)
		}
	}
}

// dispatch sends an event to all registered handlers and the activity feed
func (bus *CognitiveEventBusV3) dispatch(event CogEvent) {
	bus.mu.Lock()
	bus.totalEvents++
	bus.eventsByType[event.Category]++
	handlers := make([]CogEventHandler, len(bus.subscribers[event.Category]))
	copy(handlers, bus.subscribers[event.Category])
	bus.mu.Unlock()

	// Dispatch to handlers
	for _, handler := range handlers {
		handler(event)
	}

	// Push to activity feed
	entry := ActivityEntry{
		Timestamp: event.Timestamp,
		Category:  event.Category,
		Source:    event.Source,
		Content:   event.Content,
		Priority:  event.Priority,
	}

	select {
	case bus.activityFeed <- entry:
	default:
		// Feed full, drop
	}
}
