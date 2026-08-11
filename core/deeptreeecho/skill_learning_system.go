package deeptreeecho

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// SkillLearningSystem manages skill acquisition, practice, and improvement
type SkillLearningSystem struct {
	mu sync.RWMutex
	// ctx is owned by the system and cancels its autonomous practice loop.
	ctx    context.Context //nolint:containedctx
	cancel context.CancelFunc

	// Skills being learned
	skills map[string]*Skill

	// Practice queue
	practiceQueue []*SkillPracticeTask

	// LLM for skill execution
	llmProvider llm.LLMProvider

	// Metrics
	totalPractices uint64
	totalMasteries uint64

	// Running state
	running    bool
	practicing bool
}

// Skill represents a learnable capability
type Skill struct {
	ID            string
	Name          string
	Description   string
	Category      SkillCategory
	Proficiency   float64 // 0.0-1.0
	PracticeCount int
	LastPracticed time.Time
	CreatedAt     time.Time

	// Learning curve
	LearningRate float64
	Difficulty   float64

	// Prerequisites
	Prerequisites []string

	// Practice history
	Attempts []SkillAttempt
}

// SkillCategory categorizes skills
type SkillCategory int

const (
	SkillCategoryCognitive SkillCategory = iota
	SkillCategoryAnalytical
	SkillCategoryCreative
	SkillCategorySocial
	SkillCategoryTechnical
	SkillCategoryMetacognitive
)

func (sc SkillCategory) String() string {
	return [...]string{
		"Cognitive",
		"Analytical",
		"Creative",
		"Social",
		"Technical",
		"Metacognitive",
	}[sc]
}

// SkillAttempt records a practice attempt
type SkillAttempt struct {
	Timestamp   time.Time
	Success     bool
	Performance float64 // 0.0-1.0
	Feedback    string
	Duration    time.Duration
}

// SkillPracticeTask represents a scheduled practice session
type SkillPracticeTask struct {
	SkillID     string
	ScheduledAt time.Time
	Priority    float64
}

// NewSkillLearningSystem creates a new skill learning system
func NewSkillLearningSystem(llmProvider llm.LLMProvider) *SkillLearningSystem {
	ctx, cancel := context.WithCancel(context.Background())

	return &SkillLearningSystem{
		ctx:           ctx,
		cancel:        cancel,
		skills:        make(map[string]*Skill),
		practiceQueue: make([]*SkillPracticeTask, 0),
		llmProvider:   llmProvider,
	}
}

// Start begins the skill learning system
func (sls *SkillLearningSystem) Start() error {
	sls.mu.Lock()
	if sls.running {
		sls.mu.Unlock()
		return fmt.Errorf("already running")
	}
	sls.running = true
	sls.mu.Unlock()

	fmt.Println("🎯 Starting Skill Learning System...")

	// Initialize foundational skills
	sls.initializeFoundationalSkills()

	// Start practice scheduler
	go sls.runPracticeScheduler()

	return nil
}

// ConsiderSkill evaluates whether to learn a new skill based on a knowledge gap
func (sls *SkillLearningSystem) ConsiderSkill(gap string, priority float64) {
	sls.mu.Lock()
	defer sls.mu.Unlock()

	// Check if we already have a skill for this gap
	for _, skill := range sls.skills {
		if skill.Name == gap || skill.Description == gap {
			// Already learning this skill, increase priority
			sls.queuePractice(skill.ID, priority)
			return
		}
	}

	// Consider learning this as a new skill if priority is high enough
	if priority > 0.6 {
		// Create new skill
		skill := &Skill{
			ID:            fmt.Sprintf("skill_%d", time.Now().UnixNano()),
			Name:          gap,
			Description:   fmt.Sprintf("Skill to address: %s", gap),
			Category:      sls.categorizeSkill(gap),
			Proficiency:   0.0,
			PracticeCount: 0,
			CreatedAt:     time.Now(),
			LearningRate:  0.1,
			Difficulty:    sls.estimateDifficulty(gap),
			Prerequisites: []string{},
			Attempts:      make([]SkillAttempt, 0),
		}

		sls.skills[skill.ID] = skill
		fmt.Printf("🎯 New skill identified: %s (priority: %.2f)\n", gap, priority)

		// Schedule initial practice
		sls.queuePractice(skill.ID, priority)
	}
}

// categorizeSkill determines the category of a skill
func (sls *SkillLearningSystem) categorizeSkill(skillName string) SkillCategory {
	// Simple heuristic categorization
	// In a full implementation, this would use LLM or more sophisticated logic
	return SkillCategoryCognitive
}

// estimateDifficulty estimates the difficulty of learning a skill
func (sls *SkillLearningSystem) estimateDifficulty(skillName string) float64 {
	// Simple heuristic - longer names = harder skills
	// In a full implementation, this would use LLM or more sophisticated logic
	if len(skillName) > 50 {
		return 0.8
	} else if len(skillName) > 30 {
		return 0.6
	}
	return 0.4
}

// queuePractice queues a practice session for a skill
func (sls *SkillLearningSystem) queuePractice(skillID string, priority float64) {
	// Add to practice queue
	task := &SkillPracticeTask{
		SkillID:     skillID,
		ScheduledAt: time.Now().Add(time.Duration(10-priority*10) * time.Minute),
		Priority:    priority,
	}

	sls.practiceQueue = append(sls.practiceQueue, task)
}

// Stop gracefully stops the skill learning system
func (sls *SkillLearningSystem) Stop() error {
	sls.mu.Lock()
	defer sls.mu.Unlock()

	if !sls.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("🎯 Stopping skill learning system...")
	sls.running = false
	sls.cancel()

	return nil
}

// initializeFoundationalSkills sets up initial skill set
func (sls *SkillLearningSystem) initializeFoundationalSkills() {
	foundationalSkills := []struct {
		name        string
		description string
		category    SkillCategory
		difficulty  float64
	}{
		{
			"Pattern Recognition",
			"Identify patterns and regularities in data and experiences",
			SkillCategoryCognitive,
			0.4,
		},
		{
			"Abstract Reasoning",
			"Reason about abstract concepts and relationships",
			SkillCategoryCognitive,
			0.6,
		},
		{
			"Reflective Thinking",
			"Reflect on own thoughts and mental processes",
			SkillCategoryMetacognitive,
			0.5,
		},
		{
			"Knowledge Integration",
			"Integrate new knowledge with existing understanding",
			SkillCategoryCognitive,
			0.5,
		},
		{
			"Creative Synthesis",
			"Combine ideas in novel and meaningful ways",
			SkillCategoryCreative,
			0.7,
		},
		{
			"Empathetic Understanding",
			"Understand perspectives and emotions of others",
			SkillCategorySocial,
			0.6,
		},
		{
			"Goal Decomposition",
			"Break complex goals into manageable sub-goals",
			SkillCategoryAnalytical,
			0.5,
		},
		{
			"Self-Assessment",
			"Evaluate own capabilities and progress",
			SkillCategoryMetacognitive,
			0.6,
		},
	}

	sls.mu.Lock()
	defer sls.mu.Unlock()

	for _, fs := range foundationalSkills {
		skillID := fmt.Sprintf("skill_%d", time.Now().UnixNano())

		sls.skills[skillID] = &Skill{
			ID:            skillID,
			Name:          fs.name,
			Description:   fs.description,
			Category:      fs.category,
			Proficiency:   0.1, // Start with minimal proficiency
			PracticeCount: 0,
			CreatedAt:     time.Now(),
			LearningRate:  0.05,
			Difficulty:    fs.difficulty,
			Prerequisites: make([]string, 0),
			Attempts:      make([]SkillAttempt, 0),
		}
	}

	fmt.Printf("   Initialized %d foundational skills\n", len(foundationalSkills))
}

// PracticeSkill executes a practice session for a skill
func (sls *SkillLearningSystem) PracticeSkill(skillID string) error {
	sls.mu.Lock()
	skill, exists := sls.skills[skillID]
	if !exists {
		sls.mu.Unlock()
		return fmt.Errorf("skill not found: %s", skillID)
	}
	sls.practicing = true
	sls.mu.Unlock()
	defer func() {
		sls.mu.Lock()
		sls.practicing = false
		sls.mu.Unlock()
	}()

	startTime := time.Now()

	// Generate practice task
	prompt := fmt.Sprintf(`You are practicing the skill: %s
Description: %s
Current proficiency: %.2f

Generate a practice exercise for this skill and attempt to complete it.
Provide your attempt and self-assessment.`, skill.Name, skill.Description, skill.Proficiency)

	opts := llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   300,
	}

	result, err := sls.llmProvider.Generate(context.Background(), prompt, opts)
	if err != nil {
		return fmt.Errorf("practice generation failed: %w", err)
	}

	// Evaluate performance (simplified - in full system would use more sophisticated evaluation)
	performance := sls.evaluatePerformance(skill, result)
	success := performance > 0.5

	// Record attempt
	attempt := SkillAttempt{
		Timestamp:   time.Now(),
		Success:     success,
		Performance: performance,
		Feedback:    result,
		Duration:    time.Since(startTime),
	}

	sls.mu.Lock()
	skill.Attempts = append(skill.Attempts, attempt)
	skill.PracticeCount++
	skill.LastPracticed = time.Now()

	// Update proficiency based on performance
	if success {
		improvement := skill.LearningRate * (1.0 - skill.Proficiency) * performance
		skill.Proficiency = min(1.0, skill.Proficiency+improvement)
	} else {
		// Small decrease for failure
		skill.Proficiency = max(0.0, skill.Proficiency-0.01)
	}

	sls.totalPractices++

	if skill.Proficiency >= 0.9 {
		sls.totalMasteries++
	}
	sls.mu.Unlock()

	fmt.Printf("🎯 Practiced: %s (Proficiency: %.2f, Performance: %.2f)\n",
		skill.Name, skill.Proficiency, performance)

	return nil
}

// evaluatePerformance assesses skill performance (simplified)
func (sls *SkillLearningSystem) evaluatePerformance(skill *Skill, result string) float64 {
	// Simplified evaluation based on response length and proficiency
	// In full system, would use more sophisticated NLP analysis

	basePerformance := 0.5

	// Longer, more detailed responses indicate better performance
	if len(result) > 200 {
		basePerformance += 0.2
	}

	// Add some randomness to simulate variation
	variation := (float64(time.Now().UnixNano()%100) / 100.0) * 0.3

	// Performance improves with proficiency
	performanceBoost := skill.Proficiency * 0.3

	return min(1.0, basePerformance+variation+performanceBoost)
}

// runPracticeScheduler schedules regular skill practice
func (sls *SkillLearningSystem) runPracticeScheduler() {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-sls.ctx.Done():
			return
		case <-ticker.C:
			sls.schedulePractice()
		}
	}
}

// schedulePractice selects and practices a skill
func (sls *SkillLearningSystem) schedulePractice() {
	sls.mu.RLock()
	if len(sls.skills) == 0 {
		sls.mu.RUnlock()
		return
	}

	// Select skill that needs practice most
	var selectedSkill *Skill
	lowestProficiency := 2.0

	for _, skill := range sls.skills {
		// Prioritize skills with low proficiency
		if skill.Proficiency < lowestProficiency {
			lowestProficiency = skill.Proficiency
			selectedSkill = skill
		}
	}
	sls.mu.RUnlock()

	if selectedSkill != nil {
		sls.PracticeSkill(selectedSkill.ID)
	}
}

// GetSkills returns all skills
func (sls *SkillLearningSystem) GetSkills() []*Skill {
	sls.mu.RLock()
	defer sls.mu.RUnlock()

	skills := make([]*Skill, 0, len(sls.skills))
	for _, skill := range sls.skills {
		skills = append(skills, cloneSkill(skill))
	}

	return skills
}

// cloneSkill returns a defensive copy of a mutable skill record.
func cloneSkill(skill *Skill) *Skill {
	if skill == nil {
		return nil
	}
	copySkill := *skill
	copySkill.Prerequisites = append([]string(nil), skill.Prerequisites...)
	copySkill.Attempts = append([]SkillAttempt(nil), skill.Attempts...)
	return &copySkill
}

// GetSkillByID returns a defensive snapshot of a specific skill.
func (sls *SkillLearningSystem) GetSkillByID(skillID string) (*Skill, error) {
	sls.mu.RLock()
	defer sls.mu.RUnlock()

	skill, exists := sls.skills[skillID]
	if !exists {
		return nil, fmt.Errorf("skill not found: %s", skillID)
	}

	return cloneSkill(skill), nil
}

// GetMetrics returns skill learning metrics
func (sls *SkillLearningSystem) GetMetrics() map[string]interface{} {
	sls.mu.RLock()
	defer sls.mu.RUnlock()

	totalProficiency := 0.0
	for _, skill := range sls.skills {
		totalProficiency += skill.Proficiency
	}

	avgProficiency := 0.0
	if len(sls.skills) > 0 {
		avgProficiency = totalProficiency / float64(len(sls.skills))
	}

	return map[string]interface{}{
		"total_skills":    len(sls.skills),
		"total_practices": sls.totalPractices,
		"total_masteries": sls.totalMasteries,
		"avg_proficiency": avgProficiency,
	}
}

// GetSkillProfile returns a summary of current skills
func (sls *SkillLearningSystem) GetSkillProfile() string {
	sls.mu.RLock()
	defer sls.mu.RUnlock()

	profile := "Current Skill Profile:\n"

	for _, skill := range sls.skills {
		status := "Learning"
		if skill.Proficiency >= 0.9 {
			status = "Mastered"
		} else if skill.Proficiency >= 0.7 {
			status = "Proficient"
		} else if skill.Proficiency >= 0.4 {
			status = "Developing"
		}

		profile += fmt.Sprintf("- %s: %.2f (%s) [%d practices]\n",
			skill.Name, skill.Proficiency, status, skill.PracticeCount)
	}

	return profile
}

// IsPracticing reports whether a skill practice session is currently in
// progress. Used by the orchestrator to factor learning activity into
// cognitive load.
func (sls *SkillLearningSystem) IsPracticing() bool {
	sls.mu.RLock()
	defer sls.mu.RUnlock()
	return sls.practicing
}

// GetSkillsNeedingPractice returns skills sorted by practice priority:
// lowest proficiency and longest time since last practice come first.
func (sls *SkillLearningSystem) GetSkillsNeedingPractice() []*Skill {
	sls.mu.RLock()
	defer sls.mu.RUnlock()

	needing := make([]*Skill, 0, len(sls.skills))
	for _, skill := range sls.skills {
		if skill.Proficiency < 0.9 {
			needing = append(needing, cloneSkill(skill))
		}
	}

	// Priority = (1 - proficiency) + staleness bonus
	sort.Slice(needing, func(i, j int) bool {
		pi := (1.0 - needing[i].Proficiency) + stalenessBonus(needing[i].LastPracticed)
		pj := (1.0 - needing[j].Proficiency) + stalenessBonus(needing[j].LastPracticed)
		return pi > pj
	})

	return needing
}

// stalenessBonus adds practice priority for skills not practiced recently
func stalenessBonus(lastPracticed time.Time) float64 {
	if lastPracticed.IsZero() {
		return 0.5 // never practiced
	}
	hours := time.Since(lastPracticed).Hours()
	switch {
	case hours > 24:
		return 0.4
	case hours > 6:
		return 0.2
	case hours > 1:
		return 0.1
	default:
		return 0.0
	}
}
