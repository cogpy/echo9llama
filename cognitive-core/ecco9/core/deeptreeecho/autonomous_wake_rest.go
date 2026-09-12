package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AutonomousWakeRestManager manages autonomous wake/rest cycles
// Integrates with echodream for knowledge consolidation during rest
type AutonomousWakeRestManager struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// State
	currentState   WakeRestState
	stateStartTime time.Time
	cycleCount     uint64

	// Configuration
	minWakeDuration time.Duration
	maxWakeDuration time.Duration
	minRestDuration time.Duration
	maxRestDuration time.Duration

	// Cognitive load tracking
	cognitiveLoad float64
	fatigueLevel  float64
	learningRate  float64

	// Decision thresholds
	restThreshold float64
	wakeThreshold float64

	// Callbacks
	onWake       func() error
	onRest       func() error
	onDreamStart func() error
	onDreamEnd   func() error

	// Metrics
	totalWakeTime time.Duration
	totalRestTime time.Duration
	dreamCount    uint64

	// Running state
	running bool
}

// WakeRestState represents the current state
type WakeRestState int

const (
	StateAwake WakeRestState = iota
	StateResting
	StateDreaming
	StateTransitioning
)

func (s WakeRestState) String() string {
	return [...]string{"Awake", "Resting", "Dreaming", "Transitioning"}[s]
}

// NewAutonomousWakeRestManager creates a new wake/rest manager
func NewAutonomousWakeRestManager() *AutonomousWakeRestManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &AutonomousWakeRestManager{
		ctx:             ctx,
		cancel:          cancel,
		currentState:    StateAwake,
		stateStartTime:  time.Now(),
		minWakeDuration: 30 * time.Minute,
		maxWakeDuration: 4 * time.Hour,
		minRestDuration: 5 * time.Minute,
		maxRestDuration: 30 * time.Minute,
		cognitiveLoad:   0.0,
		fatigueLevel:    0.0,
		learningRate:    0.5,
		restThreshold:   0.75, // Rest when fatigue > 0.75
		wakeThreshold:   0.25, // Wake when fatigue < 0.25
	}
}

// SetCallbacks sets the wake/rest/dream callbacks
func (m *AutonomousWakeRestManager) SetCallbacks(
	onWake, onRest, onDreamStart, onDreamEnd func() error,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.onWake = onWake
	m.onRest = onRest
	m.onDreamStart = onDreamStart
	m.onDreamEnd = onDreamEnd
}

// Start begins autonomous wake/rest cycle management
func (m *AutonomousWakeRestManager) Start() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("already running")
	}
	m.running = true
	m.mu.Unlock()

	fmt.Println("🌙 Starting Autonomous Wake/Rest Cycle Manager...")
	fmt.Printf("   Wake Duration: %v - %v\n", m.minWakeDuration, m.maxWakeDuration)
	fmt.Printf("   Rest Duration: %v - %v\n", m.minRestDuration, m.maxRestDuration)
	fmt.Printf("   Rest Threshold: %.2f | Wake Threshold: %.2f\n", m.restThreshold, m.wakeThreshold)

	go m.run()

	return nil
}

// Stop gracefully stops the wake/rest manager
func (m *AutonomousWakeRestManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("🌙 Stopping wake/rest cycle manager...")
	m.running = false
	m.cancel()

	return nil
}

// run executes the main wake/rest cycle loop
func (m *AutonomousWakeRestManager) run() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.evaluateStateTransition()
		}
	}
}

// evaluateStateTransition checks if state should change
func (m *AutonomousWakeRestManager) evaluateStateTransition() {
	m.mu.Lock()
	currentState := m.currentState
	stateTime := time.Since(m.stateStartTime)
	m.mu.Unlock()

	switch currentState {
	case StateAwake:
		m.evaluateNeedForRest(stateTime)
	case StateResting:
		m.evaluateNeedForDream(stateTime)
	case StateDreaming:
		m.evaluateNeedForWake(stateTime)
	}
}

// evaluateNeedForRest checks if system should rest
func (m *AutonomousWakeRestManager) evaluateNeedForRest(awakeTime time.Duration) {
	m.mu.Lock()
	fatigue := m.fatigueLevel
	cogLoad := m.cognitiveLoad
	m.mu.Unlock()

	// Decision logic for resting
	shouldRest := false

	// High fatigue
	if fatigue > m.restThreshold {
		shouldRest = true
	}

	// High cognitive load for extended period
	if cogLoad > 0.8 && awakeTime > m.minWakeDuration {
		shouldRest = true
	}

	// Maximum wake duration reached
	if awakeTime > m.maxWakeDuration {
		shouldRest = true
	}

	if shouldRest {
		m.transitionToRest()
	}
}

// evaluateNeedForDream checks if system should enter dream state
func (m *AutonomousWakeRestManager) evaluateNeedForDream(restTime time.Duration) {
	// After minimum rest time, enter dream state for knowledge consolidation
	if restTime > m.minRestDuration/2 {
		m.transitionToDream()
	}
}

// evaluateNeedForWake checks if system should wake
func (m *AutonomousWakeRestManager) evaluateNeedForWake(dreamTime time.Duration) {
	m.mu.Lock()
	fatigue := m.fatigueLevel
	m.mu.Unlock()

	// Decision logic for waking
	shouldWake := false

	// Fatigue recovered
	if fatigue < m.wakeThreshold {
		shouldWake = true
	}

	// Minimum rest duration reached and fatigue low enough
	if dreamTime > m.minRestDuration && fatigue < 0.5 {
		shouldWake = true
	}

	// Maximum rest duration reached
	if dreamTime > m.maxRestDuration {
		shouldWake = true
	}

	if shouldWake {
		m.transitionToWake()
	}
}

// transitionToRest transitions to rest state
func (m *AutonomousWakeRestManager) transitionToRest() {
	m.mu.Lock()
	if m.currentState != StateAwake {
		m.mu.Unlock()
		return
	}

	awakeTime := time.Since(m.stateStartTime)
	m.totalWakeTime += awakeTime

	m.currentState = StateResting
	m.stateStartTime = time.Now()
	m.mu.Unlock()

	fmt.Printf("\n💤 Transitioning to REST (awake for %v)\n", awakeTime.Round(time.Second))
	fmt.Printf("   Fatigue: %.2f | Cognitive Load: %.2f\n", m.fatigueLevel, m.cognitiveLoad)

	if m.onRest != nil {
		if err := m.onRest(); err != nil {
			fmt.Printf("⚠️  Rest callback error: %v\n", err)
		}
	}
}

// transitionToDream transitions to dream state
func (m *AutonomousWakeRestManager) transitionToDream() {
	m.mu.Lock()
	if m.currentState != StateResting {
		m.mu.Unlock()
		return
	}

	m.currentState = StateDreaming
	m.dreamCount++
	m.mu.Unlock()

	fmt.Printf("\n🌙 Entering DREAM state (dream #%d)\n", m.dreamCount)
	fmt.Println("   Consolidating knowledge and integrating experiences...")

	if m.onDreamStart != nil {
		if err := m.onDreamStart(); err != nil {
			fmt.Printf("⚠️  Dream start callback error: %v\n", err)
		}
	}
}

// transitionToWake transitions to wake state
func (m *AutonomousWakeRestManager) transitionToWake() {
	m.mu.Lock()
	if m.currentState != StateDreaming {
		m.mu.Unlock()
		return
	}

	restTime := time.Since(m.stateStartTime)
	m.totalRestTime += restTime

	m.currentState = StateAwake
	m.stateStartTime = time.Now()
	m.cycleCount++

	// Reduce fatigue after rest
	m.fatigueLevel *= 0.3
	m.mu.Unlock()

	fmt.Printf("\n☀️  AWAKENING (rested for %v, cycle #%d)\n", restTime.Round(time.Second), m.cycleCount)
	fmt.Printf("   Fatigue: %.2f | Ready for new experiences\n", m.fatigueLevel)

	if m.onDreamEnd != nil {
		if err := m.onDreamEnd(); err != nil {
			fmt.Printf("⚠️  Dream end callback error: %v\n", err)
		}
	}

	if m.onWake != nil {
		if err := m.onWake(); err != nil {
			fmt.Printf("⚠️  Wake callback error: %v\n", err)
		}
	}
}

// UpdateCognitiveLoad updates the current cognitive load
func (m *AutonomousWakeRestManager) UpdateCognitiveLoad(load float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cognitiveLoad = load

	// Increase fatigue based on cognitive load
	if m.currentState == StateAwake {
		fatigueIncrease := load * 0.01 * m.learningRate
		m.fatigueLevel = min(1.0, m.fatigueLevel+fatigueIncrease)
	}
}

// UpdateLearningRate updates the learning rate (affects fatigue accumulation)
func (m *AutonomousWakeRestManager) UpdateLearningRate(rate float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.learningRate = rate
}

// GetState returns the current state
func (m *AutonomousWakeRestManager) GetState() WakeRestState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.currentState
}

// GetMetrics returns current metrics
func (m *AutonomousWakeRestManager) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"current_state":   m.currentState.String(),
		"state_duration":  time.Since(m.stateStartTime).Round(time.Second).String(),
		"cycle_count":     m.cycleCount,
		"dream_count":     m.dreamCount,
		"fatigue_level":   m.fatigueLevel,
		"cognitive_load":  m.cognitiveLoad,
		"total_wake_time": m.totalWakeTime.Round(time.Second).String(),
		"total_rest_time": m.totalRestTime.Round(time.Second).String(),
	}
}

// IsAwake returns true if currently awake
func (m *AutonomousWakeRestManager) IsAwake() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.currentState == StateAwake
}

// IsDreaming returns true if currently dreaming
func (m *AutonomousWakeRestManager) IsDreaming() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.currentState == StateDreaming
}
