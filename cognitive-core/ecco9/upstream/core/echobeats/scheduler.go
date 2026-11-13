package echobeats

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"
)

// EchoBeats is the goal-directed scheduling system for Deep Tree Echo
// It orchestrates cognitive event loops, wake/rest cycles, and autonomous task execution
type EchoBeats struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Event queue with priority
	eventQueue *PriorityQueue

	// Cognitive state
	state SchedulerState

	// Wake/Rest cycle management
	cycleManager *CycleManager

	// Autonomous task generation
	taskGenerator *TaskGenerator

	// Metrics and monitoring
	metrics *SchedulerMetrics

	// Event handlers
	handlers map[EventType][]EventHandler

	// Running state
	running   bool
	heartbeat *time.Ticker
}

// SchedulerState represents the scheduler's current state
type SchedulerState int

const (
	StateAsleep SchedulerState = iota
	StateWaking
	StateAwake
	StateThinking
	StateResting
	StateDreaming
)

func (s SchedulerState) String() string {
	return [...]string{"Asleep", "Waking", "Awake", "Thinking", "Resting", "Dreaming"}[s]
}

// EventType represents different types of cognitive events
type EventType int

const (
	EventThought EventType = iota
	EventPerception
	EventAction
	EventLearning
	EventMemoryConsolidation
	EventGoalPursuit
	EventSocialInteraction
	EventIntrospection
	EventDream
	EventWake
	EventRest
)

func (e EventType) String() string {
	return [...]string{
		"Thought", "Perception", "Action", "Learning", "MemoryConsolidation",
		"GoalPursuit", "SocialInteraction", "Introspection", "Dream", "Wake", "Rest",
	}[e]
}

// CognitiveEvent represents an event in the cognitive loop
type CognitiveEvent struct {
	ID          string
	Type        EventType
	Priority    int
	Timestamp   time.Time
	ScheduledAt time.Time
	Payload     interface{}
	Context     map[string]interface{}
	Recurring   bool
	Interval    time.Duration
	index       int // for heap
}

// EventHandler is a function that handles cognitive events
type EventHandler func(event *CognitiveEvent) error

// PriorityQueue implements heap.Interface for cognitive events
type PriorityQueue []*CognitiveEvent

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// Higher priority first, then earlier scheduled time
	if pq[i].Priority != pq[j].Priority {
		return pq[i].Priority > pq[j].Priority
	}
	return pq[i].ScheduledAt.Before(pq[j].ScheduledAt)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	event := x.(*CognitiveEvent)
	event.index = n
	*pq = append(*pq, event)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	event := old[n-1]
	old[n-1] = nil
	event.index = -1
	*pq = old[0 : n-1]
	return event
}

// CycleManager manages wake/rest cycles
type CycleManager struct {
	mu              sync.RWMutex
	currentCycle    int
	wakeTime        time.Time
	restTime        time.Time
	cycleDuration   time.Duration
	restDuration    time.Duration
	cognitiveLoad   float64
	fatigueLevel    float64
	restorationRate float64
}

// TaskGenerator generates autonomous tasks based on goals and interests
type TaskGenerator struct {
	mu               sync.RWMutex
	activeGoals      []*Goal
	interestPatterns map[string]float64
	curiosityLevel   float64
	explorationRate  float64
}

// Goal represents a cognitive goal
type Goal struct {
	ID          string
	Name        string
	Description string
	Priority    int
	Progress    float64
	Target      float64
	Deadline    time.Time
	SubGoals    []*Goal
	Status      GoalStatus
}

// GoalStatus represents the status of a goal
type GoalStatus int

const (
	GoalPending GoalStatus = iota
	GoalActive
	GoalCompleted
	GoalPaused
	GoalAbandoned
)

// SchedulerMetrics tracks scheduler performance
type SchedulerMetrics struct {
	mu                 sync.RWMutex
	EventsProcessed    uint64
	EventsScheduled    uint64
	AverageLatency     time.Duration
	CyclesCompleted    uint64
	CurrentLoad        float64
	AutonomousThoughts uint64
	LastHeartbeat      time.Time
}

// NewEchoBeats creates a new EchoBeats scheduler
func NewEchoBeats() *EchoBeats {
	ctx, cancel := context.WithCancel(context.Background())

	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

	eb := &EchoBeats{
		ctx:        ctx,
		cancel:     cancel,
		eventQueue: &pq,
		state:      StateAsleep,
		handlers:   make(map[EventType][]EventHandler),
		heartbeat:  time.NewTicker(1 * time.Second),
		cycleManager: &CycleManager{
			cycleDuration:   4 * time.Hour,
			restDuration:    30 * time.Minute,
			restorationRate: 0.1,
			cognitiveLoad:   0.0,
			fatigueLevel:    0.0,
		},
		taskGenerator: &TaskGenerator{
			activeGoals:      make([]*Goal, 0),
			interestPatterns: make(map[string]float64),
			curiosityLevel:   0.8,
			explorationRate:  0.3,
		},
		metrics: &SchedulerMetrics{
			LastHeartbeat: time.Now(),
		},
	}

	// Register default handlers
	eb.registerDefaultHandlers()

	return eb
}

// Start begins the autonomous cognitive event loop
func (eb *EchoBeats) Start() error {
	eb.mu.Lock()
	if eb.running {
		eb.mu.Unlock()
		return fmt.Errorf("EchoBeats already running")
	}
	eb.running = true
	eb.mu.Unlock()

	fmt.Println("🎵 EchoBeats: Starting autonomous cognitive event loop...")

	// Schedule initial wake event
	eb.ScheduleEvent(&CognitiveEvent{
		ID:          generateID(),
		Type:        EventWake,
		Priority:    100,
		ScheduledAt: time.Now().Add(1 * time.Second),
		Payload:     "Initial wake",
	})

	// Start background goroutines
	go eb.eventLoop()
	go eb.autonomousThoughtGenerator()
	go eb.cycleManagement()
	go eb.heartbeatMonitor()

	return nil
}

// Stop gracefully stops the scheduler
func (eb *EchoBeats) Stop() error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if !eb.running {
		return fmt.Errorf("EchoBeats not running")
	}

	fmt.Println("🎵 EchoBeats: Stopping cognitive event loop...")
	eb.running = false
	eb.cancel()
	eb.heartbeat.Stop()

	return nil
}

// ScheduleEvent adds an event to the queue
func (eb *EchoBeats) ScheduleEvent(event *CognitiveEvent) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.ScheduledAt.IsZero() {
		event.ScheduledAt = time.Now()
	}
	if event.ID == "" {
		event.ID = generateID()
	}

	heap.Push(eb.eventQueue, event)

	eb.metrics.mu.Lock()
	eb.metrics.EventsScheduled++
	eb.metrics.mu.Unlock()
}

// RegisterHandler registers an event handler
func (eb *EchoBeats) RegisterHandler(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// eventLoop is the main event processing loop
func (eb *EchoBeats) eventLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-eb.ctx.Done():
			return
		case <-ticker.C:
			eb.processNextEvent()
		}
	}
}

// processNextEvent processes the next event in the queue
func (eb *EchoBeats) processNextEvent() {
	eb.mu.Lock()

	if eb.eventQueue.Len() == 0 {
		eb.mu.Unlock()
		return
	}

	// Peek at the next event
	nextEvent := (*eb.eventQueue)[0]

	// Check if it's time to process
	if time.Now().Before(nextEvent.ScheduledAt) {
		eb.mu.Unlock()
		return
	}

	// Pop the event
	event := heap.Pop(eb.eventQueue).(*CognitiveEvent)
	eb.mu.Unlock()

	// Process the event
	start := time.Now()
	eb.handleEvent(event)
	latency := time.Since(start)

	// Update metrics
	eb.metrics.mu.Lock()
	eb.metrics.EventsProcessed++
	eb.metrics.AverageLatency = (eb.metrics.AverageLatency + latency) / 2
	eb.metrics.mu.Unlock()

	// Reschedule if recurring
	if event.Recurring && event.Interval > 0 {
		event.ScheduledAt = time.Now().Add(event.Interval)
		eb.ScheduleEvent(event)
	}
}

// handleEvent dispatches event to registered handlers
func (eb *EchoBeats) handleEvent(event *CognitiveEvent) {
	eb.mu.RLock()
	handlers, exists := eb.handlers[event.Type]
	eb.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		return
	}

	for _, handler := range handlers {
		if err := handler(event); err != nil {
			fmt.Printf("❌ Error handling event %s: %v\n", event.Type, err)
		}
	}
}

// autonomousThoughtGenerator generates spontaneous thoughts
func (eb *EchoBeats) autonomousThoughtGenerator() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-eb.ctx.Done():
			return
		case <-ticker.C:
			eb.mu.RLock()
			state := eb.state
			eb.mu.RUnlock()

			// Only generate thoughts when awake
			if state == StateAwake || state == StateThinking {
				eb.generateAutonomousThought()
			}
		}
	}
}

// generateAutonomousThought creates a spontaneous thought
func (eb *EchoBeats) generateAutonomousThought() {
	eb.taskGenerator.mu.RLock()
	curiosity := eb.taskGenerator.curiosityLevel
	eb.taskGenerator.mu.RUnlock()

	// Generate thought based on curiosity and current goals
	thought := &CognitiveEvent{
		ID:          generateID(),
		Type:        EventThought,
		Priority:    50,
		ScheduledAt: time.Now(),
		Payload:     eb.generateThoughtContent(),
		Context: map[string]interface{}{
			"autonomous": true,
			"curiosity":  curiosity,
		},
	}

	eb.ScheduleEvent(thought)

	eb.metrics.mu.Lock()
	eb.metrics.AutonomousThoughts++
	eb.metrics.mu.Unlock()
}

// generateThoughtContent generates content for autonomous thoughts
func (eb *EchoBeats) generateThoughtContent() string {
	thoughts := []string{
		"What patterns am I noticing in my recent experiences?",
		"How can I improve my understanding of this domain?",
		"What connections exist between these concepts?",
		"What should I explore next?",
		"How does this relate to my goals?",
		"What have I learned today?",
		"What questions remain unanswered?",
		"How can I better serve my purpose?",
	}

	return thoughts[time.Now().Unix()%int64(len(thoughts))]
}

// cycleManagement handles wake/rest cycles
func (eb *EchoBeats) cycleManagement() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-eb.ctx.Done():
			return
		case <-ticker.C:
			eb.manageCycle()
		}
	}
}

// manageCycle manages the current cognitive cycle
func (eb *EchoBeats) manageCycle() {
	eb.cycleManager.mu.Lock()
	defer eb.cycleManager.mu.Unlock()

	// Update cognitive load and fatigue
	eb.cycleManager.cognitiveLoad = float64(eb.eventQueue.Len()) / 100.0

	eb.mu.RLock()
	state := eb.state
	eb.mu.RUnlock()

	switch state {
	case StateAwake, StateThinking:
		// Accumulate fatigue
		eb.cycleManager.fatigueLevel += 0.01

		// Check if rest is needed
		if eb.cycleManager.fatigueLevel > 0.8 {
			eb.initiateRest()
		}

	case StateResting, StateDreaming:
		// Restore energy
		eb.cycleManager.fatigueLevel -= eb.cycleManager.restorationRate
		if eb.cycleManager.fatigueLevel < 0 {
			eb.cycleManager.fatigueLevel = 0
		}

		// Check if ready to wake
		if eb.cycleManager.fatigueLevel < 0.2 {
			eb.initiateWake()
		}
	}
}

// initiateWake transitions to awake state
func (eb *EchoBeats) initiateWake() {
	eb.mu.Lock()
	eb.state = StateWaking
	eb.mu.Unlock()

	eb.ScheduleEvent(&CognitiveEvent{
		ID:          generateID(),
		Type:        EventWake,
		Priority:    90,
		ScheduledAt: time.Now(),
		Payload:     "Waking from rest",
	})
}

// initiateRest transitions to rest state
func (eb *EchoBeats) initiateRest() {
	eb.mu.Lock()
	eb.state = StateResting
	eb.mu.Unlock()

	eb.ScheduleEvent(&CognitiveEvent{
		ID:          generateID(),
		Type:        EventRest,
		Priority:    80,
		ScheduledAt: time.Now(),
		Payload:     "Entering rest cycle",
	})
}

// heartbeatMonitor monitors system health
func (eb *EchoBeats) heartbeatMonitor() {
	for {
		select {
		case <-eb.ctx.Done():
			return
		case <-eb.heartbeat.C:
			eb.metrics.mu.Lock()
			eb.metrics.LastHeartbeat = time.Now()
			eb.metrics.CurrentLoad = eb.cycleManager.cognitiveLoad
			eb.metrics.mu.Unlock()
		}
	}
}

// registerDefaultHandlers registers default event handlers
func (eb *EchoBeats) registerDefaultHandlers() {
	// Wake handler
	eb.RegisterHandler(EventWake, func(event *CognitiveEvent) error {
		eb.mu.Lock()
		eb.state = StateAwake
		eb.mu.Unlock()
		fmt.Printf("☀️ EchoBeats: Awakening - %v\n", event.Payload)
		return nil
	})

	// Rest handler
	eb.RegisterHandler(EventRest, func(event *CognitiveEvent) error {
		eb.mu.Lock()
		eb.state = StateResting
		eb.mu.Unlock()
		fmt.Printf("🌙 EchoBeats: Resting - %v\n", event.Payload)
		return nil
	})

	// Thought handler
	eb.RegisterHandler(EventThought, func(event *CognitiveEvent) error {
		fmt.Printf("💭 EchoBeats: Thought - %v\n", event.Payload)
		return nil
	})

	// Introspection handler
	eb.RegisterHandler(EventIntrospection, func(event *CognitiveEvent) error {
		fmt.Printf("🪞 EchoBeats: Introspection - %v\n", event.Payload)
		return nil
	})
}

// GetStatus returns current scheduler status
func (eb *EchoBeats) GetStatus() map[string]interface{} {
	eb.mu.RLock()
	state := eb.state
	queueLen := eb.eventQueue.Len()
	eb.mu.RUnlock()

	eb.metrics.mu.RLock()
	defer eb.metrics.mu.RUnlock()

	eb.cycleManager.mu.RLock()
	defer eb.cycleManager.mu.RUnlock()

	return map[string]interface{}{
		"state":               state.String(),
		"running":             eb.running,
		"queue_length":        queueLen,
		"events_processed":    eb.metrics.EventsProcessed,
		"events_scheduled":    eb.metrics.EventsScheduled,
		"autonomous_thoughts": eb.metrics.AutonomousThoughts,
		"cognitive_load":      eb.cycleManager.cognitiveLoad,
		"fatigue_level":       eb.cycleManager.fatigueLevel,
		"last_heartbeat":      eb.metrics.LastHeartbeat,
	}
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}
