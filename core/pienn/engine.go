package pienn

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Engine is the top-level PIE-NN cognitive engine.
// It combines the Time Crystal Hierarchy, Cognitive Core, and Language Processor
// into a single interface that the daechon daemon can use.
type Engine struct {
	mu sync.RWMutex

	Core      *CognitiveCore
	Hierarchy *TimeCrystalHierarchy
	Language  *LanguageProcessor

	// Event channel for broadcasting cognitive events
	Events chan CognitiveEvent

	// State
	running    bool
	cycleCount uint64
	startTime  time.Time
}

// CognitiveEvent represents an event emitted by the PIE-NN engine
type CognitiveEvent struct {
	Type      EventType
	Source    string
	Content   string
	Level     int // Time crystal level
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// EventType categorizes cognitive events
type EventType int

const (
	EventThought      EventType = iota // Internal thought generated
	EventIntrospection                 // Self-reflection cycle completed
	EventCommand                       // PIE-NN command executed
	EventStateChange                   // Cognitive state changed
	EventEmergence                     // Emergent pattern detected
	EventDream                         // Dream cycle event
	EventWake                          // Wake event
	EventRest                          // Rest event
)

func (et EventType) String() string {
	return [...]string{
		"THOUGHT", "INTROSPECT", "COMMAND", "STATE_CHANGE",
		"EMERGENCE", "DREAM", "WAKE", "REST",
	}[et]
}

// NewEngine creates a new PIE-NN cognitive engine
func NewEngine() *Engine {
	core := NewCognitiveCore()
	hierarchy := NewTimeCrystalHierarchy()
	language := NewLanguageProcessor(core, hierarchy)

	return &Engine{
		Core:      core,
		Hierarchy: hierarchy,
		Language:  language,
		Events:    make(chan CognitiveEvent, 256),
	}
}

// Start begins the PIE-NN engine's cognitive cycle
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return fmt.Errorf("engine already running")
	}
	e.running = true
	e.startTime = time.Now()
	e.mu.Unlock()

	// Emit wake event
	e.emit(CognitiveEvent{
		Type:      EventWake,
		Source:    "pienn.engine",
		Content:   "PIE-NN cognitive engine awakening",
		Level:     5,
		Timestamp: time.Now(),
	})

	// Start the time crystal oscillator in background
	go e.runTimeCrystalLoop(ctx)

	return nil
}

// Stop halts the PIE-NN engine
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.running = false

	e.emitUnlocked(CognitiveEvent{
		Type:      EventRest,
		Source:    "pienn.engine",
		Content:   "PIE-NN cognitive engine entering rest",
		Level:     11,
		Timestamp: time.Now(),
	})
}

// Process runs input through the full PIE-NN cognitive pipeline
func (e *Engine) Process(input string) (*ProcessingResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.cycleCount++

	// Advance time crystal
	e.Hierarchy.Advance()

	// Get active temporal level
	activeLevel := e.Hierarchy.GetActiveLevel()

	// Process through cognitive core
	result := e.Core.Process(input)

	// Emit thought event
	e.emitUnlocked(CognitiveEvent{
		Type:      EventThought,
		Source:    "pienn.core",
		Content:   fmt.Sprintf("[%s@L%d] %s", result.DominantFrame, activeLevel.Level, truncate(input, 60)),
		Level:     activeLevel.Level,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"frame":  result.DominantFrame,
			"cycle":  result.Cycle,
			"level":  activeLevel.Name,
		},
	})

	return result, nil
}

// ExecuteCommand runs a PIE-NN language command
func (e *Engine) ExecuteCommand(command string) (*ExecutionResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	result, err := e.Language.Execute(command)
	if err != nil {
		return nil, err
	}

	e.emitUnlocked(CognitiveEvent{
		Type:      EventCommand,
		Source:    "pienn.language",
		Content:   fmt.Sprintf("[%s] %s", result.Command, result.Output),
		Level:     5,
		Timestamp: time.Now(),
	})

	return result, nil
}

// Introspect runs an autognosis self-reflection cycle
func (e *Engine) Introspect() *IntrospectionReport {
	e.mu.Lock()
	defer e.mu.Unlock()

	report := e.Core.Introspect()

	e.emitUnlocked(CognitiveEvent{
		Type:      EventIntrospection,
		Source:    "pienn.autognosis",
		Content:   fmt.Sprintf("Cycle %d — reasoning quality: %.2f", report.Cycle, report.MetaCognition.ReasoningQuality),
		Level:     7,
		Timestamp: time.Now(),
	})

	return report
}

// GetStatus returns the current engine status
func (e *Engine) GetStatus() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	uptime := time.Duration(0)
	if !e.startTime.IsZero() {
		uptime = time.Since(e.startTime)
	}

	return map[string]interface{}{
		"running":     e.running,
		"cycle_count": e.cycleCount,
		"uptime":      uptime.String(),
		"dominant_frame": e.Core.DominantFrame,
		"traits":      e.Core.Traits,
		"active_level": e.Hierarchy.GetActiveLevel().Name,
		"constructs":  len(e.Language.Namespace),
	}
}

// runTimeCrystalLoop advances the time crystal hierarchy periodically
func (e *Engine) runTimeCrystalLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.mu.Lock()
			if !e.running {
				e.mu.Unlock()
				return
			}
			e.Hierarchy.Advance()
			e.mu.Unlock()
		}
	}
}

func (e *Engine) emit(event CognitiveEvent) {
	select {
	case e.Events <- event:
	default:
		// Channel full, drop oldest
	}
}

func (e *Engine) emitUnlocked(event CognitiveEvent) {
	select {
	case e.Events <- event:
	default:
	}
}
