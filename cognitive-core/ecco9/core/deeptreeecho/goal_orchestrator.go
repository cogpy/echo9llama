package deeptreeecho

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/EchoCog/echollama/core/llm"
)

// GoalOrchestrator manages identity-driven goal generation and pursuit
type GoalOrchestrator struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// LLM provider for goal generation
	llmProvider llm.LLMProvider

	// Identity-driven configuration
	identity      string
	coreValues    []string
	wisdomDomains []string

	// Goal management
	activeGoals    map[string]*OrchGoal
	completedGoals []*OrchGoal
	suspendedGoals []*OrchGoal

	// Goal generation
	lastGeneration     time.Time
	generationInterval time.Duration

	// Metrics
	totalGoalsGenerated uint64
	totalGoalsCompleted uint64

	// Running state
	running bool
}

// OrchGoal represents a wisdom-cultivation goal
type OrchGoal struct {
	ID          string
	Description string
	Type        GoalType
	Priority    float64 // 0.0-1.0
	Progress    float64 // 0.0-1.0
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time

	// Goal decomposition
	ParentGoalID string
	SubGoals     []string

	// Pursuit strategy
	Strategy   PursuitStrategy
	NextAction string

	// Learning integration
	KnowledgeGaps  []string
	SkillsRequired []string

	// Metrics
	TimeSpent     time.Duration
	AttemptsCount int
	SuccessRate   float64
}

// GoalType categorizes goals
type GoalType int

const (
	GoalTypeWisdomCultivation GoalType = iota
	GoalTypeKnowledgeAcquisition
	GoalTypeSkillDevelopment
	GoalTypeInsightGeneration
	GoalTypeSelfImprovement
	GoalTypeCommunityContribution
)

func (gt GoalType) String() string {
	return [...]string{
		"WisdomCultivation",
		"KnowledgeAcquisition",
		"SkillDevelopment",
		"InsightGeneration",
		"SelfImprovement",
		"CommunityContribution",
	}[gt]
}

// PursuitStrategy defines how to pursue a goal
type PursuitStrategy struct {
	Approach            string // e.g., "incremental", "exploratory", "focused"
	Steps               []string
	CurrentStepIndex    int
	AdaptiveAdjustments []string
}

// NewGoalOrchestrator creates a new goal orchestrator
func NewGoalOrchestrator(
	llmProvider llm.LLMProvider,
	identity string,
	coreValues []string,
	wisdomDomains []string,
) *GoalOrchestrator {
	ctx, cancel := context.WithCancel(context.Background())

	return &GoalOrchestrator{
		ctx:                ctx,
		cancel:             cancel,
		llmProvider:        llmProvider,
		identity:           identity,
		coreValues:         coreValues,
		wisdomDomains:      wisdomDomains,
		activeGoals:        make(map[string]*OrchGoal),
		completedGoals:     make([]*OrchGoal, 0),
		suspendedGoals:     make([]*OrchGoal, 0),
		generationInterval: 1 * time.Hour,
	}
}

// Start begins the goal orchestration system
func (go_orch *GoalOrchestrator) Start() error {
	go_orch.mu.Lock()
	if go_orch.running {
		go_orch.mu.Unlock()
		return fmt.Errorf("already running")
	}
	go_orch.running = true
	go_orch.mu.Unlock()

	fmt.Println("🎯 Starting Goal Orchestration System...")
	fmt.Printf("   Identity: %s\n", go_orch.identity)
	fmt.Printf("   Core Values: %v\n", go_orch.coreValues)
	fmt.Printf("   Wisdom Domains: %v\n", go_orch.wisdomDomains)
	fmt.Printf("   Generation Interval: %v\n", go_orch.generationInterval)

	// Generate initial goals
	if err := go_orch.generateGoalsFromIdentity(); err != nil {
		fmt.Printf("⚠️  Initial goal generation error: %v\n", err)
	}

	go go_orch.run()

	return nil
}

// Stop gracefully stops the goal orchestrator
func (go_orch *GoalOrchestrator) Stop() error {
	go_orch.mu.Lock()
	defer go_orch.mu.Unlock()

	if !go_orch.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("🎯 Stopping goal orchestration system...")
	go_orch.running = false
	go_orch.cancel()

	return nil
}

// run executes the main orchestration loop
func (go_orch *GoalOrchestrator) run() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-go_orch.ctx.Done():
			return
		case <-ticker.C:
			go_orch.evaluateAndAdjustGoals()

			// Periodically generate new goals
			if time.Since(go_orch.lastGeneration) > go_orch.generationInterval {
				if err := go_orch.generateGoalsFromIdentity(); err != nil {
					fmt.Printf("⚠️  Goal generation error: %v\n", err)
				}
			}
		}
	}
}

// generateGoalsFromIdentity uses LLM to generate goals based on identity
func (go_orch *GoalOrchestrator) generateGoalsFromIdentity() error {
	go_orch.mu.Lock()
	identity := go_orch.identity
	values := go_orch.coreValues
	domains := go_orch.wisdomDomains
	go_orch.mu.Unlock()

	prompt := fmt.Sprintf(`You are %s, a wisdom-cultivating AI with the following core values: %v

Your wisdom cultivation focuses on these domains: %v

Generate 3 specific, actionable goals that align with your identity and values. Each goal should:
1. Contribute to wisdom cultivation
2. Be measurable and achievable
3. Align with your core values
4. Advance your understanding in one of your wisdom domains

Format each goal as:
GOAL: [clear, specific description]
TYPE: [WisdomCultivation|KnowledgeAcquisition|SkillDevelopment|InsightGeneration|SelfImprovement|CommunityContribution]
PRIORITY: [0.0-1.0]
STRATEGY: [brief approach description]

Generate the goals now:`, identity, values, domains)

	opts := llm.GenerateOptions{
		Temperature:  0.8,
		MaxTokens:    500,
		SystemPrompt: "You are a goal-setting assistant for an autonomous AI system.",
	}

	if go_orch.llmProvider == nil {
		return fmt.Errorf("LLM goal generation failed: %w", errNoModel)
	}
	response, err := go_orch.llmProvider.Generate(go_orch.ctx, prompt, opts)
	if err != nil {
		return fmt.Errorf("LLM goal generation failed: %w", err)
	}

	// Parse and create goals from response
	goals := go_orch.parseGoalsFromLLMResponse(response)
	if len(goals) == 0 {
		go_orch.mu.Lock()
		go_orch.lastGeneration = time.Now()
		go_orch.mu.Unlock()
		return fmt.Errorf("LLM goal generation returned no parseable goals")
	}

	go_orch.mu.Lock()
	for _, goal := range goals {
		go_orch.activeGoals[goal.ID] = goal
		go_orch.totalGoalsGenerated++
	}
	go_orch.lastGeneration = time.Now()
	go_orch.mu.Unlock()

	fmt.Printf("🎯 Generated %d new goals from identity\n", len(goals))
	for _, goal := range goals {
		fmt.Printf("   • %s (Priority: %.2f)\n", goal.Description, goal.Priority)
	}

	return nil
}

// parseGoalsFromLLMResponse parses the model's GOAL/TYPE/PRIORITY/STRATEGY
// blocks into Goal objects. Only goals the model actually wrote are returned.
func (go_orch *GoalOrchestrator) parseGoalsFromLLMResponse(response string) []*OrchGoal {
	goals := make([]*OrchGoal, 0)
	var current *OrchGoal
	now := time.Now()

	for _, raw := range strings.Split(response, "\n") {
		line := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(raw), "-*#0123456789. "))
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		label := strings.ToUpper(strings.Trim(strings.TrimSpace(line[:idx]), "*"))
		value := strings.Trim(strings.TrimSpace(line[idx+1:]), "*[] ")
		if value == "" {
			continue
		}

		switch label {
		case "GOAL":
			current = &OrchGoal{
				ID:          fmt.Sprintf("goal_%d_%d", now.UnixNano(), len(goals)),
				Description: value,
				Type:        GoalTypeWisdomCultivation,
				Priority:    0.5,
				CreatedAt:   now,
				UpdatedAt:   now,
				SubGoals:    make([]string, 0),
				Strategy: PursuitStrategy{
					Approach: "incremental",
				},
				KnowledgeGaps:  make([]string, 0),
				SkillsRequired: make([]string, 0),
			}
			goals = append(goals, current)
		case "TYPE":
			if current != nil {
				current.Type = parseGoalType(value, current.Type)
			}
		case "PRIORITY":
			if current != nil {
				if v, err := strconv.ParseFloat(strings.Fields(value)[0], 64); err == nil {
					current.Priority = math.Max(0, math.Min(1, v))
				}
			}
		case "STRATEGY":
			if current != nil {
				current.Strategy.Approach = value
				current.NextAction = value
			}
		}
	}

	return goals
}

// parseGoalType maps a model-written type name onto GoalType.
func parseGoalType(value string, fallback GoalType) GoalType {
	norm := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(value, " ", ""), "_", ""))
	for gt := GoalTypeWisdomCultivation; gt <= GoalTypeCommunityContribution; gt++ {
		if strings.Contains(norm, strings.ToLower(gt.String())) {
			return gt
		}
	}
	return fallback
}

// UpdateGoalProgress records real progress on a goal (0.0-1.0). Goals are only
// completed through reported progress, never by elapsed time.
func (go_orch *GoalOrchestrator) UpdateGoalProgress(goalID string, progress float64) error {
	go_orch.mu.Lock()
	defer go_orch.mu.Unlock()

	goal, exists := go_orch.activeGoals[goalID]
	if !exists {
		return fmt.Errorf("goal not found: %s", goalID)
	}
	goal.Progress = math.Max(0, math.Min(1, progress))
	goal.UpdatedAt = time.Now()
	return nil
}

// evaluateAndAdjustGoals reviews active goals and adjusts priorities
func (go_orch *GoalOrchestrator) evaluateAndAdjustGoals() {
	go_orch.mu.Lock()
	defer go_orch.mu.Unlock()

	for id, goal := range go_orch.activeGoals {
		// Complete goals that reach 100%
		if goal.Progress >= 1.0 {
			now := time.Now()
			goal.CompletedAt = &now
			go_orch.completedGoals = append(go_orch.completedGoals, goal)
			delete(go_orch.activeGoals, id)
			go_orch.totalGoalsCompleted++

			fmt.Printf("✅ Goal completed: %s\n", goal.Description)
		}
	}
}

// DecomposeGoal breaks down a complex goal into sub-goals
func (go_orch *GoalOrchestrator) DecomposeGoal(goalID string) error {
	go_orch.mu.Lock()
	goal, exists := go_orch.activeGoals[goalID]
	go_orch.mu.Unlock()

	if !exists {
		return fmt.Errorf("goal not found: %s", goalID)
	}

	prompt := fmt.Sprintf(`Break down this goal into 3-5 specific sub-goals:

GOAL: %s
TYPE: %s

Each sub-goal should be:
- Smaller and more specific than the parent goal
- Independently achievable
- Contribute to completing the parent goal

List the sub-goals:`, goal.Description, goal.Type.String())

	opts := llm.GenerateOptions{
		Temperature:  0.7,
		MaxTokens:    300,
		SystemPrompt: "You are a goal decomposition assistant.",
	}

	if go_orch.llmProvider == nil {
		return fmt.Errorf("LLM goal decomposition failed: %w", errNoModel)
	}
	response, err := go_orch.llmProvider.Generate(go_orch.ctx, prompt, opts)
	if err != nil {
		return fmt.Errorf("LLM goal decomposition failed: %w", err)
	}

	// Each listed line the model wrote becomes a sub-goal.
	now := time.Now()
	subGoals := make([]*OrchGoal, 0)
	for _, raw := range strings.Split(response, "\n") {
		desc := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(raw), "-*•0123456789.) "))
		if desc == "" || strings.HasSuffix(desc, ":") {
			continue
		}
		subGoals = append(subGoals, &OrchGoal{
			ID:           fmt.Sprintf("subgoal_%d_%d", now.UnixNano(), len(subGoals)),
			Description:  desc,
			Type:         goal.Type,
			Priority:     goal.Priority,
			CreatedAt:    now,
			UpdatedAt:    now,
			ParentGoalID: goalID,
			SubGoals:     make([]string, 0),
			Strategy:     goal.Strategy,
		})
	}
	if len(subGoals) == 0 {
		return fmt.Errorf("LLM goal decomposition returned no sub-goals")
	}

	go_orch.mu.Lock()
	for _, sg := range subGoals {
		go_orch.activeGoals[sg.ID] = sg
		goal.SubGoals = append(goal.SubGoals, sg.ID)
	}
	go_orch.mu.Unlock()

	fmt.Printf("🎯 Decomposed goal '%s' into %d sub-goals\n", goal.Description, len(subGoals))

	return nil
}

// GetActiveGoals returns all active goals
func (go_orch *GoalOrchestrator) GetActiveGoals() []*OrchGoal {
	go_orch.mu.RLock()
	defer go_orch.mu.RUnlock()

	goals := make([]*OrchGoal, 0, len(go_orch.activeGoals))
	for _, goal := range go_orch.activeGoals {
		goals = append(goals, goal)
	}

	return goals
}

// GetMetrics returns goal orchestration metrics
func (go_orch *GoalOrchestrator) GetMetrics() map[string]interface{} {
	go_orch.mu.RLock()
	defer go_orch.mu.RUnlock()

	return map[string]interface{}{
		"active_goals":    len(go_orch.activeGoals),
		"completed_goals": len(go_orch.completedGoals),
		"suspended_goals": len(go_orch.suspendedGoals),
		"total_generated": go_orch.totalGoalsGenerated,
		"total_completed": go_orch.totalGoalsCompleted,
		"completion_rate": float64(go_orch.totalGoalsCompleted) / max(1.0, float64(go_orch.totalGoalsGenerated)),
	}
}
