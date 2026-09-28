package echodream

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AutonomousWakeRestController manages self-directed wake and rest cycles
// based on cognitive load, fatigue, and knowledge integration needs
type AutonomousWakeRestController struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Core components
	dreamSystem *EchoDream

	// Cognitive state monitoring
	cognitiveLoad      float64
	fatigueLevel       float64
	integrationBacklog int
	consolidationNeed  float64

	// Wake/rest state
	currentState    WakeRestState
	lastStateChange time.Time
	wakeDuration    time.Duration
	restDuration    time.Duration

	// Thresholds for autonomous decisions
	fatigueThreshold       float64
	consolidationThreshold float64
	minWakeDuration        time.Duration
	minRestDuration        time.Duration

	// Metrics
	wakeEpisodes    uint64
	restEpisodes    uint64
	autonomousWakes uint64
	autonomousRests uint64
	totalWakeTime   time.Duration
	totalRestTime   time.Duration

	running bool
}

// WakeRestState represents the current wake/rest state
type WakeRestState int

const (
	StateAwake WakeRestState = iota
	StateResting
	StateDreaming
	StateTransitioning
)

func (wrs WakeRestState) String() string {
	return [...]string{"Awake", "Resting", "Dreaming", "Transitioning"}[wrs]
}

// NewAutonomousWakeRestController creates a new autonomous controller
func NewAutonomousWakeRestController(dreamSystem *EchoDream) *AutonomousWakeRestController {
	ctx, cancel := context.WithCancel(context.Background())

	return &AutonomousWakeRestController{
		ctx:                    ctx,
		cancel:                 cancel,
		dreamSystem:            dreamSystem,
		cognitiveLoad:          0.3,
		fatigueLevel:           0.0,
		integrationBacklog:     0,
		consolidationNeed:      0.0,
		currentState:           StateAwake,
		lastStateChange:        time.Now(),
		fatigueThreshold:       0.7,
		consolidationThreshold: 0.6,
		minWakeDuration:        5 * time.Minute,
		minRestDuration:        2 * time.Minute,
	}
}

// Start begins autonomous wake/rest management
func (awrc *AutonomousWakeRestController) Start() error {
	awrc.mu.Lock()
	if awrc.running {
		awrc.mu.Unlock()
		return fmt.Errorf("autonomous wake/rest controller already running")
	}
	awrc.running = true
	awrc.mu.Unlock()

	// Start cognitive state monitoring
	go awrc.cognitiveStateMonitoringLoop()

	// Start autonomous decision loop
	go awrc.autonomousDecisionLoop()

	// Start integration assessment loop
	go awrc.integrationAssessmentLoop()

	return nil
}

// Stop halts autonomous wake/rest management
func (awrc *AutonomousWakeRestController) Stop() {
	awrc.mu.Lock()
	awrc.running = false
	awrc.mu.Unlock()

	awrc.cancel()
}

// cognitiveStateMonitoringLoop continuously monitors cognitive state
func (awrc *AutonomousWakeRestController) cognitiveStateMonitoringLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-awrc.ctx.Done():
			return
		case <-ticker.C:
			awrc.updateCognitiveState()
		}
	}
}

// updateCognitiveState updates cognitive load and fatigue levels
func (awrc *AutonomousWakeRestController) updateCognitiveState() {
	awrc.mu.Lock()
	defer awrc.mu.Unlock()

	// Update based on current state
	switch awrc.currentState {
	case StateAwake:
		// Increase fatigue while awake
		awrc.fatigueLevel += 0.01

		// Cognitive load varies based on activity
		// In real implementation, would measure actual cognitive activity
		awrc.cognitiveLoad = 0.5 + 0.3*awrc.fatigueLevel

		// Increase consolidation need over time
		awrc.consolidationNeed += 0.005

	case StateResting, StateDreaming:
		// Decrease fatigue while resting
		awrc.fatigueLevel -= 0.02

		// Decrease cognitive load
		awrc.cognitiveLoad -= 0.03

		// Decrease consolidation need as memories are processed
		awrc.consolidationNeed -= 0.01
	}

	// Clamp values
	awrc.fatigueLevel = clamp(awrc.fatigueLevel, 0.0, 1.0)
	awrc.cognitiveLoad = clamp(awrc.cognitiveLoad, 0.0, 1.0)
	awrc.consolidationNeed = clamp(awrc.consolidationNeed, 0.0, 1.0)
}

// autonomousDecisionLoop makes autonomous wake/rest decisions
func (awrc *AutonomousWakeRestController) autonomousDecisionLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-awrc.ctx.Done():
			return
		case <-ticker.C:
			awrc.makeAutonomousDecision()
		}
	}
}

// makeAutonomousDecision decides whether to wake or rest
func (awrc *AutonomousWakeRestController) makeAutonomousDecision() {
	awrc.mu.RLock()
	currentState := awrc.currentState
	fatigue := awrc.fatigueLevel
	consolidation := awrc.consolidationNeed
	timeSinceChange := time.Since(awrc.lastStateChange)
	awrc.mu.RUnlock()

	switch currentState {
	case StateAwake:
		// Check if should rest
		shouldRest := awrc.shouldEnterRest(fatigue, consolidation, timeSinceChange)
		if shouldRest {
			awrc.initiateRest()
		}

	case StateResting, StateDreaming:
		// Check if should wake
		shouldWake := awrc.shouldWake(fatigue, consolidation, timeSinceChange)
		if shouldWake {
			awrc.initiateWake()
		}
	}
}

// shouldEnterRest determines if the system should enter rest state
func (awrc *AutonomousWakeRestController) shouldEnterRest(fatigue, consolidation float64, timeSinceChange time.Duration) bool {
	// Must be awake for minimum duration
	if timeSinceChange < awrc.minWakeDuration {
		return false
	}

	// Rest if fatigue is high
	if fatigue > awrc.fatigueThreshold {
		return true
	}

	// Rest if consolidation need is high
	if consolidation > awrc.consolidationThreshold {
		return true
	}

	// Rest if both are moderately high
	if fatigue > 0.5 && consolidation > 0.4 {
		return true
	}

	return false
}

// shouldWake determines if the system should wake from rest
func (awrc *AutonomousWakeRestController) shouldWake(fatigue, consolidation float64, timeSinceChange time.Duration) bool {
	// Must rest for minimum duration
	if timeSinceChange < awrc.minRestDuration {
		return false
	}

	// Wake if fatigue is low
	if fatigue < 0.2 {
		return true
	}

	// Wake if consolidation need is low
	if consolidation < 0.2 {
		return true
	}

	// Wake if both are sufficiently low
	if fatigue < 0.4 && consolidation < 0.3 {
		return true
	}

	return false
}

// initiateRest begins a rest/dream cycle
func (awrc *AutonomousWakeRestController) initiateRest() {
	awrc.mu.Lock()

	// Record wake episode
	wakeDuration := time.Since(awrc.lastStateChange)
	awrc.totalWakeTime += wakeDuration
	awrc.wakeEpisodes++

	// Transition to resting
	awrc.currentState = StateResting
	awrc.lastStateChange = time.Now()
	awrc.autonomousRests++

	awrc.mu.Unlock()

	// Start dream system
	if awrc.dreamSystem != nil {
		awrc.dreamSystem.Start()
	}

	fmt.Println("🌙 Autonomous Rest: Entering rest/dream cycle for knowledge integration")
	fmt.Printf("   Fatigue: %.2f, Consolidation Need: %.2f\n", awrc.fatigueLevel, awrc.consolidationNeed)
}

// initiateWake begins a wake cycle
func (awrc *AutonomousWakeRestController) initiateWake() {
	awrc.mu.Lock()

	// Record rest episode
	restDuration := time.Since(awrc.lastStateChange)
	awrc.totalRestTime += restDuration
	awrc.restEpisodes++

	// Transition to awake
	awrc.currentState = StateAwake
	awrc.lastStateChange = time.Now()
	awrc.autonomousWakes++

	awrc.mu.Unlock()

	// Stop dream system
	if awrc.dreamSystem != nil {
		awrc.dreamSystem.Stop()
	}

	fmt.Println("☀️  Autonomous Wake: Emerging from rest, ready for new experiences")
	fmt.Printf("   Fatigue: %.2f, Consolidation Need: %.2f\n", awrc.fatigueLevel, awrc.consolidationNeed)
}

// integrationAssessmentLoop assesses knowledge integration needs
func (awrc *AutonomousWakeRestController) integrationAssessmentLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-awrc.ctx.Done():
			return
		case <-ticker.C:
			awrc.assessIntegrationNeeds()
		}
	}
}

// assessIntegrationNeeds evaluates the need for knowledge consolidation
func (awrc *AutonomousWakeRestController) assessIntegrationNeeds() {
	awrc.mu.Lock()
	defer awrc.mu.Unlock()

	// In real implementation, would query memory system for:
	// - Unconsolidated short-term memories
	// - Unintegrated experiences
	// - Pending pattern syntheses

	// For now, simulate assessment
	awrc.integrationBacklog = int(awrc.consolidationNeed * 100)
}

// GetState returns the current wake/rest state
func (awrc *AutonomousWakeRestController) GetState() WakeRestState {
	awrc.mu.RLock()
	defer awrc.mu.RUnlock()
	return awrc.currentState
}

// GetCognitiveMetrics returns current cognitive state metrics
func (awrc *AutonomousWakeRestController) GetCognitiveMetrics() map[string]interface{} {
	awrc.mu.RLock()
	defer awrc.mu.RUnlock()

	return map[string]interface{}{
		"state":               awrc.currentState.String(),
		"cognitive_load":      awrc.cognitiveLoad,
		"fatigue_level":       awrc.fatigueLevel,
		"consolidation_need":  awrc.consolidationNeed,
		"integration_backlog": awrc.integrationBacklog,
		"time_in_state":       time.Since(awrc.lastStateChange).Seconds(),
	}
}

// GetMetrics returns controller metrics
func (awrc *AutonomousWakeRestController) GetMetrics() map[string]interface{} {
	awrc.mu.RLock()
	defer awrc.mu.RUnlock()

	return map[string]interface{}{
		"wake_episodes":     awrc.wakeEpisodes,
		"rest_episodes":     awrc.restEpisodes,
		"autonomous_wakes":  awrc.autonomousWakes,
		"autonomous_rests":  awrc.autonomousRests,
		"total_wake_time":   awrc.totalWakeTime.Seconds(),
		"total_rest_time":   awrc.totalRestTime.Seconds(),
		"avg_wake_duration": awrc.getAverageWakeDuration(),
		"avg_rest_duration": awrc.getAverageRestDuration(),
	}
}

func (awrc *AutonomousWakeRestController) getAverageWakeDuration() float64 {
	if awrc.wakeEpisodes == 0 {
		return 0.0
	}
	return awrc.totalWakeTime.Seconds() / float64(awrc.wakeEpisodes)
}

func (awrc *AutonomousWakeRestController) getAverageRestDuration() float64 {
	if awrc.restEpisodes == 0 {
		return 0.0
	}
	return awrc.totalRestTime.Seconds() / float64(awrc.restEpisodes)
}

// SetFatigueThreshold allows adjusting the fatigue threshold
func (awrc *AutonomousWakeRestController) SetFatigueThreshold(threshold float64) {
	awrc.mu.Lock()
	defer awrc.mu.Unlock()
	awrc.fatigueThreshold = clamp(threshold, 0.0, 1.0)
}

// SetConsolidationThreshold allows adjusting the consolidation threshold
func (awrc *AutonomousWakeRestController) SetConsolidationThreshold(threshold float64) {
	awrc.mu.Lock()
	defer awrc.mu.Unlock()
	awrc.consolidationThreshold = clamp(threshold, 0.0, 1.0)
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
