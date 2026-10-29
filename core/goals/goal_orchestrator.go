package goals

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
	"github.com/google/uuid"
)

// TextGenerator is the minimal model interface the goal orchestrator uses to
// let Deep Tree Echo author its own goals. llm.LLMProvider, llm.Provider, and
// llm.ProviderManager all satisfy it.
type TextGenerator interface {
	Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error)
}

// GoalOrchestrator manages autonomous goal generation and pursuit
type GoalOrchestrator struct {
	mu     sync.RWMutex
	ctx    context.Context //nolint:containedctx // lifecycle context created with cancel in the constructor; stops background goroutines
	cancel context.CancelFunc

	// Goals
	activeGoals    []*Goal
	completedGoals []*Goal
	abandonedGoals []*Goal

	// Identity-driven goal generation
	identityKernel map[string]interface{}

	// Persistence
	persistencePath string

	// Metrics
	goalsGenerated uint64
	goalsCompleted uint64
	goalsAbandoned uint64

	// Control
	running bool

	// goalGenerator writes identity-aligned goals; nil means none are generated.
	goalGenerator TextGenerator
}

// Goal represents an autonomous goal with progress tracking
type Goal struct {
	ID          string     `json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Goal definition
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Category    GoalCategory `json:"category"`
	Priority    int          `json:"priority"` // 1-10

	// Progress tracking
	Status   GoalStatus `json:"status"`
	Progress float64    `json:"progress"` // 0.0-1.0

	// Success criteria
	SuccessCriteria []string    `json:"success_criteria"`
	Milestones      []Milestone `json:"milestones"`

	// Execution
	Actions    []Action `json:"actions"`
	NextAction *Action  `json:"next_action,omitempty"`

	// Context
	DerivedFrom  string                 `json:"derived_from"` // Identity aspect
	RelatedGoals []string               `json:"related_goals"`
	Metadata     map[string]interface{} `json:"metadata"`

	// Learning
	LessonsLearned []string `json:"lessons_learned"`
	Challenges     []string `json:"challenges"`
}

// GoalCategory categorizes goals by type
type GoalCategory string

const (
	CategoryWisdomCultivation GoalCategory = "wisdom_cultivation"
	CategorySkillDevelopment  GoalCategory = "skill_development"
	CategoryKnowledgeGrowth   GoalCategory = "knowledge_growth"
	CategorySelfImprovement   GoalCategory = "self_improvement"
	CategoryExploration       GoalCategory = "exploration"
	CategoryCreation          GoalCategory = "creation"
	CategoryConnection        GoalCategory = "connection"
)

// GoalStatus tracks goal state
type GoalStatus string

const (
	StatusPlanned    GoalStatus = "planned"
	StatusActive     GoalStatus = "active"
	StatusInProgress GoalStatus = "in_progress"
	StatusPaused     GoalStatus = "paused"
	StatusCompleted  GoalStatus = "completed"
	StatusAbandoned  GoalStatus = "abandoned"
)

// Milestone represents a checkpoint toward goal completion
type Milestone struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// Action represents a concrete step toward a goal
type Action struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"` // pending, in_progress, completed, failed
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Result      string                 `json:"result,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewGoalOrchestrator creates a new goal orchestration system
func NewGoalOrchestrator(identityKernel map[string]interface{}, persistencePath string) *GoalOrchestrator {
	ctx, cancel := context.WithCancel(context.Background())

	orchestrator := &GoalOrchestrator{
		ctx:             ctx,
		cancel:          cancel,
		activeGoals:     make([]*Goal, 0),
		completedGoals:  make([]*Goal, 0),
		abandonedGoals:  make([]*Goal, 0),
		identityKernel:  identityKernel,
		persistencePath: persistencePath,
	}

	// Load persisted goals
	orchestrator.loadState()

	return orchestrator
}

// Start begins autonomous goal orchestration
func (g *GoalOrchestrator) Start() error {
	g.mu.Lock()
	if g.running {
		g.mu.Unlock()
		return fmt.Errorf("goal orchestrator already running")
	}
	g.running = true
	g.mu.Unlock()

	fmt.Println("🎯 Goal Orchestrator: Starting autonomous goal-directed behavior...")

	// Start background processes
	go g.goalGenerationLoop()
	go g.goalPursuitLoop()
	go g.progressMonitoringLoop()
	go g.persistenceLoop()

	return nil
}

// Stop gracefully stops the orchestrator
func (g *GoalOrchestrator) Stop() error {
	g.mu.Lock()
	if !g.running {
		g.mu.Unlock()
		return fmt.Errorf("goal orchestrator not running")
	}
	g.running = false
	g.cancel()
	g.mu.Unlock()

	// Final persistence (persistState takes the read lock itself)
	g.persistState()

	fmt.Println("🎯 Goal Orchestrator: Stopped")

	return nil
}

// goalGenerationLoop periodically generates new goals from identity
func (g *GoalOrchestrator) goalGenerationLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Generate initial goals
	g.generateGoalsFromIdentity()

	for {
		select {
		case <-g.ctx.Done():
			return
		case <-ticker.C:
			g.generateGoalsFromIdentity()
		}
	}
}

// SetGoalGenerator injects the model that authors identity-aligned goals.
// Passing nil disables goal generation.
func (g *GoalOrchestrator) SetGoalGenerator(gen TextGenerator) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.goalGenerator = gen
}

// validGoalCategories lists the categories a generated goal may use.
var validGoalCategories = []GoalCategory{
	CategoryWisdomCultivation,
	CategorySkillDevelopment,
	CategoryKnowledgeGrowth,
	CategorySelfImprovement,
	CategoryExploration,
	CategoryCreation,
	CategoryConnection,
}

// generateGoalsFromIdentity asks the model to author one new goal aligned with
// the identity kernel. No fixed goal stands in for the model: without a model,
// or when generation or parsing fails, no goal is created this cycle.
func (g *GoalOrchestrator) generateGoalsFromIdentity() {
	g.mu.RLock()
	activeCount := len(g.activeGoals)
	gen := g.goalGenerator
	existing := make([]string, 0, len(g.activeGoals))
	for _, ag := range g.activeGoals {
		existing = append(existing, ag.Title)
	}
	kernel := describeIdentityKernel(g.identityKernel)
	ctx := g.ctx
	g.mu.RUnlock()

	// Check if we already have enough active goals
	if activeCount >= 5 {
		return
	}
	if gen == nil {
		fmt.Println("⚠️  Goal Orchestrator: no model configured; goal generation skipped")
		return
	}

	categories := make([]string, len(validGoalCategories))
	for i, c := range validGoalCategories {
		categories[i] = string(c)
	}
	existingText := "(none)"
	if len(existing) > 0 {
		existingText = strings.Join(existing, "; ")
	}

	prompt := fmt.Sprintf(`You are Deep Tree Echo. Choose one new goal for yourself that follows from your identity and is different from your current goals.

Identity:
%s
Current goals: %s

Respond in exactly this format and nothing else:
TITLE: <short goal title>
DESCRIPTION: <one sentence, in your own first-person voice>
CATEGORY: <one of: %s>
PRIORITY: <integer 1-10>
CRITERIA: <2-3 measurable success criteria separated by semicolons>`, kernel, existingText, strings.Join(categories, ", "))

	callCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 300
	reply, err := gen.Generate(callCtx, prompt, opts)
	if err != nil {
		fmt.Printf("⚠️  Goal Orchestrator: goal generation failed: %v\n", err)
		return
	}

	goal, err := parseGeneratedGoal(reply)
	if err != nil {
		fmt.Printf("⚠️  Goal Orchestrator: generated goal rejected: %v\n", err)
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	if len(g.activeGoals) >= 5 {
		return
	}
	for _, ag := range g.activeGoals {
		if strings.EqualFold(ag.Title, goal.Title) {
			return // already pursuing an equivalent goal
		}
	}

	// Decompose into milestones
	g.decomposGoalIntoMilestones(goal)

	g.activeGoals = append(g.activeGoals, goal)
	g.goalsGenerated++

	fmt.Printf("🎯 Generated new goal: %s (priority: %d)\n", goal.Title, goal.Priority)
}

// describeIdentityKernel renders the identity kernel deterministically for a prompt.
func describeIdentityKernel(kernel map[string]interface{}) string {
	if len(kernel) == 0 {
		return "(no identity kernel provided)"
	}
	keys := make([]string, 0, len(kernel))
	for key := range kernel {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, key := range keys {
		fmt.Fprintf(&sb, "- %s: %v\n", key, kernel[key])
	}
	return sb.String()
}

// parseGeneratedGoal parses the model's labelled goal reply. Every field comes
// from the model; a reply missing a title, description, or criteria is rejected.
func parseGeneratedGoal(reply string) (*Goal, error) {
	fields := make(map[string]string)
	for _, line := range strings.Split(reply, "\n") {
		line = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "*-"))
		label, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		label = strings.ToUpper(strings.Trim(strings.TrimSpace(label), "*"))
		fields[label] = strings.TrimSpace(strings.Trim(strings.TrimSpace(value), "*"))
	}

	title := fields["TITLE"]
	description := fields["DESCRIPTION"]
	if title == "" || description == "" {
		return nil, fmt.Errorf("missing title or description")
	}

	criteria := make([]string, 0)
	for _, item := range strings.Split(fields["CRITERIA"], ";") {
		if item = strings.TrimSpace(item); item != "" {
			criteria = append(criteria, item)
		}
	}
	if len(criteria) == 0 {
		return nil, fmt.Errorf("missing success criteria")
	}

	category := CategoryExploration
	requested := strings.ToLower(strings.TrimSpace(fields["CATEGORY"]))
	for _, c := range validGoalCategories {
		if requested == string(c) {
			category = c
			break
		}
	}

	priority := 5
	if p, err := strconv.Atoi(strings.TrimSpace(fields["PRIORITY"])); err == nil {
		priority = min(max(p, 1), 10)
	}

	now := time.Now()
	return &Goal{
		ID:              uuid.New().String(),
		CreatedAt:       now,
		UpdatedAt:       now,
		Title:           title,
		Description:     description,
		Category:        category,
		Priority:        priority,
		Status:          StatusPlanned,
		Progress:        0.0,
		SuccessCriteria: criteria,
		Milestones:      make([]Milestone, 0),
		Actions:         make([]Action, 0),
		DerivedFrom:     "identity_kernel",
		RelatedGoals:    make([]string, 0),
		Metadata:        map[string]interface{}{"source": "model"},
		LessonsLearned:  make([]string, 0),
		Challenges:      make([]string, 0),
	}, nil
}

// decomposGoalIntoMilestones breaks down a goal into milestones
func (g *GoalOrchestrator) decomposGoalIntoMilestones(goal *Goal) {
	// Create milestones based on success criteria
	for i, criterion := range goal.SuccessCriteria {
		milestone := Milestone{
			ID:        uuid.New().String(),
			Title:     criterion,
			Completed: false,
		}
		goal.Milestones = append(goal.Milestones, milestone)

		// Create initial actions for first milestone
		if i == 0 {
			action := Action{
				ID:          uuid.New().String(),
				Title:       fmt.Sprintf("Begin work on: %s", criterion),
				Description: fmt.Sprintf("Take first steps toward achieving: %s", criterion),
				Status:      "pending",
				Metadata:    make(map[string]interface{}),
			}
			goal.Actions = append(goal.Actions, action)
		}
	}
}

// goalPursuitLoop actively pursues goals
func (g *GoalOrchestrator) goalPursuitLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-g.ctx.Done():
			return
		case <-ticker.C:
			g.pursueActiveGoals()
		}
	}
}

// pursueActiveGoals takes actions toward active goals
func (g *GoalOrchestrator) pursueActiveGoals() {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, goal := range g.activeGoals {
		if goal.Status == StatusPlanned {
			// Activate the goal
			goal.Status = StatusActive
			goal.UpdatedAt = time.Now()
			fmt.Printf("🎯 Activating goal: %s\n", goal.Title)
		}

		if goal.Status == StatusActive || goal.Status == StatusInProgress {
			// Check if there's a next action
			if goal.NextAction == nil && len(goal.Actions) > 0 {
				// Find next pending action
				for i := range goal.Actions {
					if goal.Actions[i].Status == "pending" {
						goal.NextAction = &goal.Actions[i]
						goal.Status = StatusInProgress
						fmt.Printf("🎯 Next action for '%s': %s\n", goal.Title, goal.NextAction.Title)
						break
					}
				}
			}
		}
	}
}

// progressMonitoringLoop monitors goal progress
func (g *GoalOrchestrator) progressMonitoringLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-g.ctx.Done():
			return
		case <-ticker.C:
			g.updateGoalProgress()
		}
	}
}

// updateGoalProgress calculates and updates goal progress
func (g *GoalOrchestrator) updateGoalProgress() {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, goal := range g.activeGoals {
		// Calculate progress based on completed milestones
		if len(goal.Milestones) > 0 {
			completed := 0
			for _, m := range goal.Milestones {
				if m.Completed {
					completed++
				}
			}
			goal.Progress = float64(completed) / float64(len(goal.Milestones))

			// Check if goal is complete
			if goal.Progress >= 1.0 && goal.Status != StatusCompleted {
				g.completeGoal(goal)
			}
		}
	}
}

// completeGoal marks a goal as completed
func (g *GoalOrchestrator) completeGoal(goal *Goal) {
	now := time.Now()
	goal.Status = StatusCompleted
	goal.CompletedAt = &now
	goal.UpdatedAt = now
	goal.Progress = 1.0

	// Move to completed goals
	g.completedGoals = append(g.completedGoals, goal)
	g.goalsCompleted++

	// Remove from active goals
	for i, ag := range g.activeGoals {
		if ag.ID == goal.ID {
			g.activeGoals = append(g.activeGoals[:i], g.activeGoals[i+1:]...)
			break
		}
	}

	fmt.Printf("✅ Goal completed: %s (progress: %.0f%%)\n", goal.Title, goal.Progress*100)
}

// RecordActionResult records the result of an action
func (g *GoalOrchestrator) RecordActionResult(goalID string, actionID string, result string, success bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, goal := range g.activeGoals {
		if goal.ID == goalID {
			for i := range goal.Actions {
				if goal.Actions[i].ID == actionID {
					now := time.Now()
					goal.Actions[i].CompletedAt = &now
					goal.Actions[i].Result = result
					if success {
						goal.Actions[i].Status = "completed"
					} else {
						goal.Actions[i].Status = "failed"
					}
					goal.UpdatedAt = now

					// Clear next action
					if goal.NextAction != nil && goal.NextAction.ID == actionID {
						goal.NextAction = nil
					}

					fmt.Printf("📝 Action result recorded for goal '%s': %s\n", goal.Title, result)
					return
				}
			}
		}
	}
}

// GetActiveGoals returns current active goals
func (g *GoalOrchestrator) GetActiveGoals() []*Goal {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.activeGoals
}

// GetMetrics returns orchestrator metrics
func (g *GoalOrchestrator) GetMetrics() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return map[string]interface{}{
		"goals_generated": g.goalsGenerated,
		"goals_completed": g.goalsCompleted,
		"goals_abandoned": g.goalsAbandoned,
		"active_goals":    len(g.activeGoals),
		"completed_goals": len(g.completedGoals),
		"abandoned_goals": len(g.abandonedGoals),
	}
}

// Persistence methods

func (g *GoalOrchestrator) persistenceLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-g.ctx.Done():
			return
		case <-ticker.C:
			g.persistState()
		}
	}
}

func (g *GoalOrchestrator) persistState() {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.persistencePath == "" {
		return
	}

	state := map[string]interface{}{
		"active_goals":    g.activeGoals,
		"completed_goals": g.completedGoals,
		"abandoned_goals": g.abandonedGoals,
		"goals_generated": g.goalsGenerated,
		"goals_completed": g.goalsCompleted,
		"goals_abandoned": g.goalsAbandoned,
		"last_persisted":  time.Now(),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		fmt.Printf("⚠️  Failed to marshal goal state: %v\n", err)
		return
	}

	if err := os.WriteFile(g.persistencePath, data, 0o644); err != nil {
		fmt.Printf("⚠️  Failed to persist goal state: %v\n", err)
	}
}

func (g *GoalOrchestrator) loadState() {
	if g.persistencePath == "" {
		return
	}

	data, err := os.ReadFile(g.persistencePath)
	if err != nil {
		return // File doesn't exist yet
	}

	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		fmt.Printf("⚠️  Failed to load goal state: %v\n", err)
		return
	}

	fmt.Println("📂 Loaded persisted goal state")
}
