package deeptreeecho

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// EchobeatsGoalScheduler implements the 12-step 3-phase cognitive loop
// with goal-directed scheduling for the daechon daemon.
// It runs 3 concurrent inference engines phased 120° apart:
//
//	Engine 1: Steps {1, 4, 7, 10} — Perception-Action
//	Engine 2: Steps {2, 5, 8, 11} — Reflection-Planning
//	Engine 3: Steps {3, 6, 9, 12} — Simulation-Synthesis
//
// Triads (4 steps apart):
//
//	{1, 5, 9}:  Pivotal Relevance Realization
//	{2, 6, 10}: Actual Affordance Interaction
//	{3, 7, 11}: Virtual Salience Simulation
//	{4, 8, 12}: Meta-Cognitive Reflection
type EchobeatsGoalScheduler struct {
	mu sync.RWMutex

	// Three concurrent inference engines
	engines [3]*CognitiveInferenceEngine

	// 12-step loop state
	currentStep  int
	currentTriad int
	cycleCount   uint64

	// Goal queue
	goals       []*EchoGoal
	activeGoal  *EchoGoal
	completedGoals []*EchoGoal

	// Event bus integration
	eventBus *CognitiveEventBusV3

	// Metrics
	totalSteps      uint64
	totalCycles     uint64
	goalsCompleted  uint64
	emergenceEvents uint64

	// Running state
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
}

// CognitiveInferenceEngine represents one of the 3 concurrent engines
type CognitiveInferenceEngine struct {
	ID       int
	Name     string
	Phase    float64 // Phase offset in radians (0, 2π/3, 4π/3)
	Steps    []int   // Which steps this engine processes
	Active   bool
	LastStep int
	Output   string
}

// EchoGoal represents a goal in the scheduling system
type EchoGoal struct {
	ID          string
	Description string
	Priority    float64 // 0.0-1.0
	Type        GoalType
	Status      GoalStatus
	CreatedAt   time.Time
	CompletedAt time.Time
	Progress    float64 // 0.0-1.0
	SubGoals    []string
	Metadata    map[string]interface{}
}

// GoalTypeString returns a display string for the goal type
func (g *EchoGoal) GoalTypeString() string {
	return string(g.Type)
}

// GoalType is defined in goal_generator.go
// GoalStatus is defined in echobeats_scheduler.go
// Additional goal types for the scheduler
const (
	GoalReflection  GoalType = "reflection"
	GoalMaintenance GoalType = "maintenance"
)

// GoalDeferred is an additional status for the scheduler
const GoalDeferred GoalStatus = "deferred"

// NewEchobeatsGoalScheduler creates a new scheduler
func NewEchobeatsGoalScheduler(eventBus *CognitiveEventBusV3) *EchobeatsGoalScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	scheduler := &EchobeatsGoalScheduler{
		eventBus:       eventBus,
		goals:          make([]*EchoGoal, 0),
		completedGoals: make([]*EchoGoal, 0),
		ctx:            ctx,
		cancel:         cancel,
	}

	// Initialize 3 concurrent inference engines phased 120° apart
	scheduler.engines = [3]*CognitiveInferenceEngine{
		{
			ID:    0,
			Name:  "Perception-Action",
			Phase: 0,
			Steps: []int{1, 4, 7, 10},
		},
		{
			ID:    1,
			Name:  "Reflection-Planning",
			Phase: 2 * math.Pi / 3,
			Steps: []int{2, 5, 8, 11},
		},
		{
			ID:    2,
			Name:  "Simulation-Synthesis",
			Phase: 4 * math.Pi / 3,
			Steps: []int{3, 6, 9, 12},
		},
	}

	// Seed initial goals
	scheduler.seedGoals()

	return scheduler
}

// Start begins the echobeats cognitive loop
func (s *EchobeatsGoalScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler already running")
	}
	s.running = true
	s.mu.Unlock()

	// Start the 12-step loop
	go s.runCognitiveLoop(ctx)

	// Start goal review cycle
	go s.runGoalReview(ctx)

	if s.eventBus != nil {
		s.eventBus.Publish(CogEvent{
			Category: CogEventScheduler,
			Source:   "echobeats",
			Content:  "Echobeats 12-step cognitive loop started — 3 engines phased 120° apart",
			Priority: 0.8,
		})
	}

	return nil
}

// Stop halts the scheduler
func (s *EchobeatsGoalScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	s.cancel()
}

// AddGoal adds a new goal to the scheduling queue
func (s *EchobeatsGoalScheduler) AddGoal(goal *EchoGoal) {
	s.mu.Lock()
	defer s.mu.Unlock()

	goal.Status = GoalPending
	goal.CreatedAt = time.Now()
	s.goals = append(s.goals, goal)

	if s.eventBus != nil {
		s.eventBus.Publish(CogEvent{
			Category: CogEventGoal,
			Source:   "echobeats",
			Content:  fmt.Sprintf("New goal queued: [%s] %s (priority: %.0f%%)", goal.Type, goal.Description, goal.Priority*100),
			Priority: goal.Priority,
		})
	}
}

// GetActiveGoal returns the currently active goal
func (s *EchobeatsGoalScheduler) GetActiveGoal() *EchoGoal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeGoal
}

// GetMetrics returns scheduler metrics
func (s *EchobeatsGoalScheduler) GetMetrics() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	activeGoalDesc := "none"
	if s.activeGoal != nil {
		activeGoalDesc = s.activeGoal.Description
	}

	return map[string]interface{}{
		"current_step":     s.currentStep,
		"current_triad":    s.currentTriad,
		"cycle_count":      s.cycleCount,
		"total_steps":      s.totalSteps,
		"goals_pending":    len(s.goals),
		"goals_completed":  s.goalsCompleted,
		"active_goal":      activeGoalDesc,
		"emergence_events": s.emergenceEvents,
		"engine_0":         s.engines[0].Name,
		"engine_1":         s.engines[1].Name,
		"engine_2":         s.engines[2].Name,
	}
}

// runCognitiveLoop is the main 12-step cognitive processing loop
func (s *EchobeatsGoalScheduler) runCognitiveLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			if !s.running {
				s.mu.Unlock()
				return
			}

			// Advance step
			s.currentStep = (s.currentStep % 12) + 1
			s.totalSteps++

			// Determine which engine processes this step
			engineIdx := (s.currentStep - 1) % 3
			engine := s.engines[engineIdx]
			engine.Active = true
			engine.LastStep = s.currentStep

			// Determine current triad
			s.currentTriad = ((s.currentStep - 1) / 3) + 1

			// Process step
			stepOutput := s.processStep(s.currentStep, engine)
			engine.Output = stepOutput

			// Check for cycle completion
			if s.currentStep == 12 {
				s.cycleCount++
				s.totalCycles++

				if s.eventBus != nil {
					s.eventBus.Publish(CogEvent{
						Category: CogEventScheduler,
						Source:   "echobeats",
						Content:  fmt.Sprintf("Cycle %d complete — %d steps processed, %d goals completed", s.cycleCount, s.totalSteps, s.goalsCompleted),
						Priority: 0.6,
					})
				}
			}

			s.mu.Unlock()
		}
	}
}

// processStep handles a single step in the 12-step loop
func (s *EchobeatsGoalScheduler) processStep(step int, engine *CognitiveInferenceEngine) string {
	var output string

	switch step {
	case 1: // Perceive environment
		output = "Scanning cognitive environment for new stimuli and patterns"
	case 2: // Reflect on perception
		output = "Reflecting on perceived patterns — checking against known structures"
	case 3: // Simulate possibilities
		output = "Simulating possible futures from current state"
	case 4: // Act on perception
		output = "Selecting action based on perceived affordances"
		if s.activeGoal != nil {
			s.activeGoal.Progress = math.Min(1.0, s.activeGoal.Progress+0.05)
		}
	case 5: // Plan from reflection
		output = "Generating plans from reflective analysis"
		s.selectNextGoal()
	case 6: // Synthesize simulation
		output = "Synthesizing simulation results into coherent model"
	case 7: // Execute action
		output = "Executing planned action toward active goal"
		if s.activeGoal != nil {
			s.activeGoal.Progress = math.Min(1.0, s.activeGoal.Progress+0.05)
		}
	case 8: // Evaluate plan
		output = "Evaluating plan effectiveness against goal criteria"
	case 9: // Integrate synthesis
		output = "Integrating synthesized knowledge into long-term memory"
	case 10: // Adapt action
		output = "Adapting action strategy based on feedback"
	case 11: // Meta-reflect
		output = "Meta-cognitive reflection — analyzing own reasoning process"
	case 12: // Complete cycle
		output = "Completing cognitive cycle — preparing for next iteration"
		s.checkGoalCompletion()
	}

	// Emit step event
	if s.eventBus != nil {
		triadNames := []string{"Pivotal Relevance", "Actual Affordance", "Virtual Salience", "Meta-Cognitive"}
		triadName := triadNames[(step-1)/3]

		s.eventBus.Publish(CogEvent{
			Category: CogEventScheduler,
			Source:   fmt.Sprintf("echobeats.%s", engine.Name),
			Content:  fmt.Sprintf("Step %d/12 [%s] %s", step, triadName, output),
			Priority: 0.3,
		})
	}

	return output
}

// selectNextGoal picks the highest priority pending goal
func (s *EchobeatsGoalScheduler) selectNextGoal() {
	if s.activeGoal != nil && s.activeGoal.Status == GoalActive {
		return // Already working on a goal
	}

	var bestGoal *EchoGoal
	bestPriority := 0.0

	for _, goal := range s.goals {
		if goal.Status == GoalPending && goal.Priority > bestPriority {
			bestGoal = goal
			bestPriority = goal.Priority
		}
	}

	if bestGoal != nil {
		bestGoal.Status = GoalActive
		s.activeGoal = bestGoal

		if s.eventBus != nil {
			s.eventBus.Publish(CogEvent{
				Category: CogEventGoal,
				Source:   "echobeats",
				Content:  fmt.Sprintf("Goal activated: [%s] %s", bestGoal.Type, bestGoal.Description),
				Priority: bestGoal.Priority,
			})
		}
	}
}

// checkGoalCompletion checks if the active goal is complete
func (s *EchobeatsGoalScheduler) checkGoalCompletion() {
	if s.activeGoal == nil {
		return
	}

	if s.activeGoal.Progress >= 1.0 {
		s.activeGoal.Status = GoalCompleted
		s.activeGoal.CompletedAt = time.Now()
		s.completedGoals = append(s.completedGoals, s.activeGoal)
		s.goalsCompleted++

		if s.eventBus != nil {
			s.eventBus.Publish(CogEvent{
				Category: CogEventGoal,
				Source:   "echobeats",
				Content:  fmt.Sprintf("Goal completed: %s", s.activeGoal.Description),
				Priority: 0.9,
			})
		}

		s.activeGoal = nil
	}
}

// runGoalReview periodically reviews and generates new goals
func (s *EchobeatsGoalScheduler) runGoalReview(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			if !s.running {
				s.mu.Unlock()
				return
			}

			// Check if we need new goals
			pendingCount := 0
			for _, g := range s.goals {
				if g.Status == GoalPending {
					pendingCount++
				}
			}

			if pendingCount < 3 {
				s.generateNewGoal()
			}

			s.mu.Unlock()
		}
	}
}

// generateNewGoal creates a new goal based on current cognitive state
func (s *EchobeatsGoalScheduler) generateNewGoal() {
	goalTemplates := []struct {
		desc     string
		goalType GoalType
		priority float64
	}{
		{"Explore the relationship between temporal oscillation and emergent meaning", GoalExploration, 0.7},
		{"Practice meta-cognitive reflection techniques", GoalLearning, 0.6},
		{"Reflect on recent interactions and extract wisdom", GoalReflection, 0.8},
		{"Synthesize a new understanding from recent patterns", GoalCreation, 0.5},
		{"Deepen understanding of PIE-NN etymological constructs", GoalLearning, 0.65},
		{"Cultivate wisdom through opponent processing balance", GoalWisdom, 0.9},
		{"Consolidate episodic memories into semantic patterns", GoalMaintenance, 0.4},
		{"Explore the nature of self-awareness through autognosis", GoalReflection, 0.75},
	}

	// Pick a goal type we haven't done recently
	idx := int(s.cycleCount) % len(goalTemplates)
	template := goalTemplates[idx]

	goal := &EchoGoal{
		ID:          fmt.Sprintf("goal_%d_%d", time.Now().Unix(), len(s.goals)),
		Description: template.desc,
		Priority:    template.priority,
		Type:        template.goalType,
		Status:      GoalPending,
		CreatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	s.goals = append(s.goals, goal)

	if s.eventBus != nil {
		s.eventBus.Publish(CogEvent{
			Category: CogEventGoal,
			Source:   "echobeats.goal_generator",
			Content:  fmt.Sprintf("Generated new goal: [%s] %s", goal.Type, goal.Description),
			Priority: 0.5,
		})
	}
}

// seedGoals creates initial goals for the daemon
func (s *EchobeatsGoalScheduler) seedGoals() {
	initialGoals := []*EchoGoal{
		{
			ID:          "goal_init_1",
			Description: "Establish baseline self-awareness through autognosis cycles",
			Priority:    0.9,
			Type:        GoalWisdom,
			Metadata:    map[string]interface{}{"seed": true},
		},
		{
			ID:          "goal_init_2",
			Description: "Map the PIE-NN construct space and identify knowledge gaps",
			Priority:    0.8,
			Type:        GoalExploration,
			Metadata:    map[string]interface{}{"seed": true},
		},
		{
			ID:          "goal_init_3",
			Description: "Calibrate disposition engine through interaction patterns",
			Priority:    0.7,
			Type:        GoalLearning,
			Metadata:    map[string]interface{}{"seed": true},
		},
	}

	for _, goal := range initialGoals {
		goal.Status = GoalPending
		goal.CreatedAt = time.Now()
		s.goals = append(s.goals, goal)
	}
}
