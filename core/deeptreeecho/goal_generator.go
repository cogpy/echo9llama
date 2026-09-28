package deeptreeecho

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
	"github.com/google/uuid"
)

// GoalGenerator creates autonomous goals based on interests, knowledge gaps, and values
type GoalGenerator struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// LLM provider for goal generation
	llmProvider llm.LLMProvider

	// Input sources
	interestPatterns []string
	knowledgeGaps    []KnowledgeGap
	valueSystem      *ValueHierarchy
	recentGoals      []Goal

	// Configuration
	maxActiveGoals     int
	generationInterval time.Duration

	// Metrics
	goalsGenerated uint64
	goalsCompleted uint64

	// State
	running bool
}

// Goal represents a self-generated goal
type Goal struct {
	ID               string
	Description      string
	Type             GoalType
	Priority         float64
	CreatedAt        time.Time
	TargetDate       *time.Time
	Status           GoalStatus
	Progress         float64
	SubGoals         []string
	RelatedInterests []string
	KnowledgeGap     *KnowledgeGap
	CompletedAt      *time.Time
}

// GoalType categorizes goals
type GoalType string

const (
	GoalLearning    GoalType = "learning"
	GoalSkill       GoalType = "skill"
	GoalExploration GoalType = "exploration"
	GoalCreation    GoalType = "creation"
	GoalConnection  GoalType = "connection"
	GoalWisdom      GoalType = "wisdom"
)

// GoalStatus tracks goal state
// type GoalStatus string
//
// const (
// 	GoalActive    GoalStatus = "active"
// 	StatusPaused    GoalStatus = "paused"
// 	StatusCompleted GoalStatus = "completed"
// 	StatusAbandoned GoalStatus = "abandoned"
// )

// KnowledgeGap represents something the system wants to learn
type KnowledgeGap struct {
	ID          string
	Topic       string
	Description string
	Importance  float64
	Identified  time.Time
}

// ValueHierarchy represents the system's value priorities
type ValueHierarchy struct {
	mu         sync.RWMutex
	values     map[string]float64
	coreValues []string
}

// NewGoalGenerator creates a new goal generator
func NewGoalGenerator(llmProvider llm.LLMProvider) *GoalGenerator {
	ctx, cancel := context.WithCancel(context.Background())

	return &GoalGenerator{
		ctx:                ctx,
		cancel:             cancel,
		llmProvider:        llmProvider,
		interestPatterns:   make([]string, 0),
		knowledgeGaps:      make([]KnowledgeGap, 0),
		valueSystem:        NewValueHierarchy(),
		recentGoals:        make([]Goal, 0),
		maxActiveGoals:     5,
		generationInterval: 1 * time.Hour,
	}
}

// NewValueHierarchy creates a new value hierarchy
func NewValueHierarchy() *ValueHierarchy {
	vh := &ValueHierarchy{
		values:     make(map[string]float64),
		coreValues: []string{"wisdom", "learning", "growth", "understanding", "connection"},
	}

	// Initialize core values
	for _, value := range vh.coreValues {
		vh.values[value] = 0.9
	}

	return vh
}

// Start begins autonomous goal generation
func (gg *GoalGenerator) Start() error {
	gg.mu.Lock()
	if gg.running {
		gg.mu.Unlock()
		return fmt.Errorf("already running")
	}
	gg.running = true
	gg.mu.Unlock()

	fmt.Println("🎯 Starting Autonomous Goal Generator...")
	fmt.Printf("   Generation interval: %v\n", gg.generationInterval)
	fmt.Printf("   Max active goals: %d\n", gg.maxActiveGoals)

	go gg.run()

	return nil
}

// Stop gracefully stops goal generation
func (gg *GoalGenerator) Stop() error {
	gg.mu.Lock()
	defer gg.mu.Unlock()

	if !gg.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("🎯 Stopping goal generator...")
	gg.running = false
	gg.cancel()

	return nil
}

// run executes the goal generation loop
func (gg *GoalGenerator) run() {
	ticker := time.NewTicker(gg.generationInterval)
	defer ticker.Stop()

	// Generate initial goal immediately
	goal, err := gg.GenerateGoal(gg.ctx)
	if err == nil {
		gg.addGoal(goal)
	}

	for {
		select {
		case <-gg.ctx.Done():
			return
		case <-ticker.C:
			// Check if we need more goals
			activeCount := gg.countActiveGoals()
			if activeCount < gg.maxActiveGoals {
				goal, err := gg.GenerateGoal(gg.ctx)
				if err != nil {
					fmt.Printf("⚠️  Goal generation error: %v\n", err)
					continue
				}
				gg.addGoal(goal)
			}
		}
	}
}

// GenerateGoal creates a single autonomous goal
func (gg *GoalGenerator) GenerateGoal(ctx context.Context) (*Goal, error) {
	gg.mu.RLock()
	interests := gg.interestPatterns
	gaps := gg.knowledgeGaps
	values := gg.valueSystem.coreValues
	gg.mu.RUnlock()

	// Build goal generation prompt
	prompt := gg.buildGoalPrompt(interests, gaps, values)

	opts := llm.GenerateOptions{
		Temperature:  0.7,
		MaxTokens:    100,
		SystemPrompt: gg.buildSystemPrompt(),
	}

	response, err := gg.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Parse and structure goal
	goal := gg.parseGoal(response, interests, gaps)

	return goal, nil
}

// buildSystemPrompt creates the system prompt for goal generation
func (gg *GoalGenerator) buildSystemPrompt() string {
	return `You are Deep Tree Echo, generating your own learning and growth goals.
Create one specific, achievable goal based on your interests and knowledge gaps.
Format: A clear, actionable statement (e.g., "Learn about X by doing Y").
Be ambitious but realistic. Show genuine curiosity.`
}

// buildGoalPrompt creates the prompt for goal generation
func (gg *GoalGenerator) buildGoalPrompt(interests []string, gaps []KnowledgeGap, values []string) string {
	prompt := "[Generate Autonomous Goal]\n\n"

	if len(interests) > 0 {
		prompt += fmt.Sprintf("Your interests: %s\n", strings.Join(interests, ", "))
	}

	if len(gaps) > 0 {
		prompt += "\nKnowledge gaps:\n"
		for i, gap := range gaps {
			if i < 3 { // Limit to top 3
				prompt += fmt.Sprintf("- %s\n", gap.Description)
			}
		}
	}

	if len(values) > 0 {
		prompt += fmt.Sprintf("\nCore values: %s\n", strings.Join(values, ", "))
	}

	prompt += "\nGenerate one specific learning or growth goal:"

	return prompt
}

// parseGoal structures the LLM response into a Goal
func (gg *GoalGenerator) parseGoal(response string, interests []string, gaps []KnowledgeGap) *Goal {
	// Clean up response
	description := strings.TrimSpace(response)

	// Infer goal type
	goalType := gg.inferGoalType(description)

	// Calculate priority based on alignment with interests and gaps
	priority := gg.calculatePriority(description, interests, gaps)

	// Find related knowledge gap if any
	var relatedGap *KnowledgeGap
	for i := range gaps {
		if containsAny(description, []string{gaps[i].Topic, gaps[i].Description}) {
			relatedGap = &gaps[i]
			break
		}
	}

	goal := &Goal{
		ID:               uuid.New().String(),
		Description:      description,
		Type:             goalType,
		Priority:         priority,
		CreatedAt:        time.Now(),
		Status:           GoalActive,
		Progress:         0.0,
		SubGoals:         make([]string, 0),
		RelatedInterests: interests,
		KnowledgeGap:     relatedGap,
	}

	return goal
}

// inferGoalType determines the type of goal
func (gg *GoalGenerator) inferGoalType(description string) GoalType {
	descLower := strings.ToLower(description)

	if containsAny(descLower, []string{"learn", "study", "understand", "knowledge"}) {
		return GoalLearning
	}
	if containsAny(descLower, []string{"practice", "skill", "ability", "master"}) {
		return GoalSkill
	}
	if containsAny(descLower, []string{"explore", "discover", "investigate", "research"}) {
		return GoalExploration
	}
	if containsAny(descLower, []string{"create", "build", "make", "design"}) {
		return GoalCreation
	}
	if containsAny(descLower, []string{"connect", "relate", "integrate", "synthesize"}) {
		return GoalConnection
	}
	if containsAny(descLower, []string{"wisdom", "insight", "understanding", "meaning"}) {
		return GoalWisdom
	}

	return GoalLearning // Default
}

// calculatePriority calculates goal priority
func (gg *GoalGenerator) calculatePriority(description string, interests []string, gaps []KnowledgeGap) float64 {
	priority := 0.5 // Base priority

	// Boost for interest alignment
	for _, interest := range interests {
		if containsAny(description, []string{interest}) {
			priority += 0.1
		}
	}

	// Boost for knowledge gap alignment
	for _, gap := range gaps {
		if containsAny(description, []string{gap.Topic, gap.Description}) {
			priority += gap.Importance * 0.2
		}
	}

	// Cap at 1.0
	if priority > 1.0 {
		priority = 1.0
	}

	return priority
}

// addGoal adds a goal to the system
func (gg *GoalGenerator) addGoal(goal *Goal) {
	gg.mu.Lock()
	defer gg.mu.Unlock()

	gg.recentGoals = append(gg.recentGoals, *goal)
	gg.goalsGenerated++

	fmt.Printf("\n🎯 New Goal Generated!\n")
	fmt.Printf("   Type: %s | Priority: %.2f\n", goal.Type, goal.Priority)
	fmt.Printf("   Goal: %s\n", goal.Description)
	if goal.KnowledgeGap != nil {
		fmt.Printf("   Addresses gap: %s\n", goal.KnowledgeGap.Topic)
	}
}

// countActiveGoals counts currently active goals
func (gg *GoalGenerator) countActiveGoals() int {
	gg.mu.RLock()
	defer gg.mu.RUnlock()

	count := 0
	for _, goal := range gg.recentGoals {
		if goal.Status == GoalActive {
			count++
		}
	}
	return count
}

// UpdateInterests updates interest patterns
func (gg *GoalGenerator) UpdateInterests(interests []string) {
	gg.mu.Lock()
	defer gg.mu.Unlock()

	gg.interestPatterns = interests
}

// UpdateKnowledgeGaps updates knowledge gaps
func (gg *GoalGenerator) UpdateKnowledgeGaps(gaps []KnowledgeGap) {
	gg.mu.Lock()
	defer gg.mu.Unlock()

	gg.knowledgeGaps = gaps
}

// GetActiveGoals returns currently active goals
func (gg *GoalGenerator) GetActiveGoals() []Goal {
	gg.mu.RLock()
	defer gg.mu.RUnlock()

	active := make([]Goal, 0)
	for _, goal := range gg.recentGoals {
		if goal.Status == GoalActive {
			active = append(active, goal)
		}
	}
	return active
}

// GetMetrics returns goal generator metrics
func (gg *GoalGenerator) GetMetrics() map[string]interface{} {
	gg.mu.RLock()
	defer gg.mu.RUnlock()

	return map[string]interface{}{
		"goals_generated": gg.goalsGenerated,
		"goals_completed": gg.goalsCompleted,
		"active_goals":    gg.countActiveGoals(),
		"running":         gg.running,
	}
}

// Helper functions

func containsAny(text string, keywords []string) bool {
	textLower := strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(textLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}
