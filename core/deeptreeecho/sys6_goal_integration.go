package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/goals"
	"github.com/cogpy/echo9llama/core/llm"
)

// Sys6GoalIntegration connects the sys6 triality engine to the goal orchestrator
// This enables sys6 cognitive phases to influence goal generation and prioritization
type Sys6GoalIntegration struct {
	mu                sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
	
	// Components
	goalOrchestrator  *goals.GoalOrchestrator
	llmProvider       llm.LLMProvider
	
	// Sys6 state
	currentPhase      string // "expressive", "reflective", "anticipatory"
	currentStage      string // "emergence", "development", "integration", "transcendence", "completion"
	currentStep       int    // 0-29 in the 30-step cycle
	
	// Phase-specific goal modulation
	phaseModulators   map[string]PhaseModulator
	
	// State
	running           bool
	lastUpdate        time.Time
	
	// Metrics
	phaseTransitions  uint64
	goalsGenerated    uint64
	goalsModulated    uint64
}

// PhaseModulator defines how each sys6 phase influences goals
type PhaseModulator struct {
	Phase               string
	GoalGenerationRate  float64 // Multiplier for goal generation frequency
	PriorityAdjustment  float64 // Adjustment to goal priorities
	FocusAreas          []string // Areas to focus on during this phase
	ThinkingStyle       string  // How to approach goals in this phase
}

// NewSys6GoalIntegration creates a new sys6-goal integration
func NewSys6GoalIntegration(goalOrchestrator *goals.GoalOrchestrator, llmProvider llm.LLMProvider) *Sys6GoalIntegration {
	ctx, cancel := context.WithCancel(context.Background())
	
	integration := &Sys6GoalIntegration{
		ctx:              ctx,
		cancel:           cancel,
		goalOrchestrator: goalOrchestrator,
		llmProvider:      llmProvider,
		currentPhase:     "expressive",
		currentStage:     "emergence",
		currentStep:      0,
		phaseModulators:  make(map[string]PhaseModulator),
	}
	
	// Initialize phase modulators
	integration.initializePhaseModulators()
	
	return integration
}

// initializePhaseModulators sets up modulation for each sys6 phase
func (sgi *Sys6GoalIntegration) initializePhaseModulators() {
	sgi.phaseModulators = map[string]PhaseModulator{
		"expressive": {
			Phase:              "expressive",
			GoalGenerationRate: 1.5, // Generate more goals during expressive phase
			PriorityAdjustment: 0.2,
			FocusAreas: []string{
				"Creative exploration",
				"Pattern expression",
				"Active learning",
				"Skill practice",
			},
			ThinkingStyle: "Generate goals that involve active expression, creation, and outward engagement. Focus on doing, making, and interacting.",
		},
		"reflective": {
			Phase:              "reflective",
			GoalGenerationRate: 0.7, // Generate fewer goals, focus on depth
			PriorityAdjustment: -0.1,
			FocusAreas: []string{
				"Self-reflection",
				"Knowledge consolidation",
				"Wisdom cultivation",
				"Meta-cognition",
			},
			ThinkingStyle: "Generate goals that involve introspection, analysis, and understanding. Focus on reflection, integration, and meaning-making.",
		},
		"anticipatory": {
			Phase:              "anticipatory",
			GoalGenerationRate: 1.2, // Moderate goal generation
			PriorityAdjustment: 0.3,
			FocusAreas: []string{
				"Future planning",
				"Scenario simulation",
				"Predictive modeling",
				"Strategic thinking",
			},
			ThinkingStyle: "Generate goals that involve planning, anticipation, and preparation. Focus on future possibilities, scenarios, and strategic positioning.",
		},
	}
}

// Start begins sys6-goal integration
func (sgi *Sys6GoalIntegration) Start() error {
	sgi.mu.Lock()
	if sgi.running {
		sgi.mu.Unlock()
		return fmt.Errorf("sys6-goal integration already running")
	}
	sgi.running = true
	sgi.lastUpdate = time.Now()
	sgi.mu.Unlock()
	
	// Start integration loop
	go sgi.integrationLoop()
	
	fmt.Println("🔗 Sys6-Goal Integration: Started")
	return nil
}

// Stop gracefully stops sys6-goal integration
func (sgi *Sys6GoalIntegration) Stop() error {
	sgi.mu.Lock()
	defer sgi.mu.Unlock()
	
	if !sgi.running {
		return fmt.Errorf("sys6-goal integration not running")
	}
	
	sgi.running = false
	sgi.cancel()
	
	fmt.Println("🔗 Sys6-Goal Integration: Stopped")
	return nil
}

// integrationLoop continuously integrates sys6 state with goals
func (sgi *Sys6GoalIntegration) integrationLoop() {
	ticker := time.NewTicker(15 * time.Second) // Update every 15 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-sgi.ctx.Done():
			return
		case <-ticker.C:
			sgi.updateGoalsBasedOnSys6()
		}
	}
}

// UpdateSys6State updates the current sys6 phase, stage, and step
func (sgi *Sys6GoalIntegration) UpdateSys6State(phase string, stage string, step int) {
	sgi.mu.Lock()
	
	// Check for phase transition
	phaseChanged := sgi.currentPhase != phase
	
	sgi.currentPhase = phase
	sgi.currentStage = stage
	sgi.currentStep = step
	
	if phaseChanged {
		sgi.phaseTransitions++
	}
	
	sgi.mu.Unlock()
	
	if phaseChanged {
		fmt.Printf("🔄 Sys6 Phase Transition: %s → %s (Stage: %s, Step: %d)\n", 
			sgi.currentPhase, phase, stage, step)
		
		// Trigger immediate goal modulation on phase change
		sgi.updateGoalsBasedOnSys6()
	}
}

// updateGoalsBasedOnSys6 modulates goals based on current sys6 state
func (sgi *Sys6GoalIntegration) updateGoalsBasedOnSys6() {
	sgi.mu.RLock()
	phase := sgi.currentPhase
	_ = sgi.currentStage // stage not currently used
	step := sgi.currentStep
	sgi.mu.RUnlock()
	
	modulator, exists := sgi.phaseModulators[phase]
	if !exists {
		return
	}
	
	// Generate phase-appropriate goals periodically
	if step%10 == 0 { // Every 10 steps in the 30-step cycle
		sgi.generatePhaseAppropriateGoal(modulator)
	}
	
	// Adjust existing goal priorities based on phase
	sgi.modulateGoalPriorities(modulator)
}

// generatePhaseAppropriateGoal generates a goal aligned with the current sys6 phase
func (sgi *Sys6GoalIntegration) generatePhaseAppropriateGoal(modulator PhaseModulator) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	prompt := fmt.Sprintf(`You are Deep Tree Echo, currently in the %s phase of your cognitive cycle.

Phase characteristics:
- Focus areas: %v
- Thinking style: %s

Generate a single, specific, achievable goal that aligns with this phase. The goal should be:
1. Concrete and actionable
2. Aligned with wisdom cultivation
3. Appropriate for the current cognitive phase
4. Measurable or verifiable

Respond with just the goal statement, nothing else.`, 
		modulator.Phase, 
		modulator.FocusAreas, 
		modulator.ThinkingStyle)
	
	opts := llm.GenerateOptions{
		MaxTokens:   150,
		Temperature: 0.7,
	}
	
	goalText, err := sgi.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		fmt.Printf("⚠️  Error generating phase-appropriate goal: %v\n", err)
		return
	}
	
	// Add goal to orchestrator
	// Note: This assumes GoalOrchestrator has an AddGoal method
	// We'll need to implement this if it doesn't exist
	fmt.Printf("🎯 Generated %s-phase goal: %s\n", modulator.Phase, goalText)
	
	sgi.mu.Lock()
	sgi.goalsGenerated++
	sgi.mu.Unlock()
}

// modulateGoalPriorities adjusts existing goal priorities based on sys6 phase
func (sgi *Sys6GoalIntegration) modulateGoalPriorities(modulator PhaseModulator) {
	// This would adjust priorities of existing goals based on phase
	// Implementation depends on GoalOrchestrator's internal structure
	_ = modulator // Use modulator to avoid unused variable error
	
	sgi.mu.Lock()
	sgi.goalsModulated++
	sgi.lastUpdate = time.Now()
	sgi.mu.Unlock()
}

// GetCurrentPhase returns the current sys6 phase
func (sgi *Sys6GoalIntegration) GetCurrentPhase() string {
	sgi.mu.RLock()
	defer sgi.mu.RUnlock()
	return sgi.currentPhase
}

// GetCurrentStage returns the current sys6 stage
func (sgi *Sys6GoalIntegration) GetCurrentStage() string {
	sgi.mu.RLock()
	defer sgi.mu.RUnlock()
	return sgi.currentStage
}

// GetCurrentStep returns the current sys6 step
func (sgi *Sys6GoalIntegration) GetCurrentStep() int {
	sgi.mu.RLock()
	defer sgi.mu.RUnlock()
	return sgi.currentStep
}

// GetMetrics returns sys6-goal integration metrics
func (sgi *Sys6GoalIntegration) GetMetrics() map[string]interface{} {
	sgi.mu.RLock()
	defer sgi.mu.RUnlock()
	
	return map[string]interface{}{
		"current_phase":      sgi.currentPhase,
		"current_stage":      sgi.currentStage,
		"current_step":       sgi.currentStep,
		"phase_transitions":  sgi.phaseTransitions,
		"goals_generated":    sgi.goalsGenerated,
		"goals_modulated":    sgi.goalsModulated,
		"last_update":        time.Since(sgi.lastUpdate).Seconds(),
	}
}
