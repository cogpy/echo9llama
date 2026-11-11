package deeptreeecho

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
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

	// Running state
	running         bool
	practicing      bool
	activePractices int

	// IDs bind to the complete normalized payload, not just an award marker.
	appliedEvidence map[string]skillEvidencePayload
}

const (
	SkillCompetenceVersion   = "evaluator-ema-v1"
	SkillCompetenceEMAWeight = 0.2
)

// SkillCompetenceEstimate summarizes only observed evaluator evidence. The EMA
// starts at zero and uses score on a pass, zero on a failure. The Beta(1,1)
// posterior describes pass frequency, not general ability. Neither statistic
// certifies mastery or transfer beyond the supplied evaluator's tasks.
type SkillCompetenceEstimate struct {
	Version           string
	Advisory          bool
	Evaluations       uint64
	Passes            uint64
	Failures          uint64
	ScoreSum          float64
	MeanScore         float64
	EffectiveScoreEMA float64
	EMAWeight         float64
	PassPosteriorMean float64
}

type skillEvidencePayload struct {
	SkillName string
	Score     float64
	Passed    bool
	Feedback  string
}

func newSkillCompetenceEstimate() SkillCompetenceEstimate {
	return SkillCompetenceEstimate{
		Version: SkillCompetenceVersion, Advisory: true,
		EMAWeight: SkillCompetenceEMAWeight, PassPosteriorMean: 0.5,
	}
}

// Skill represents a learnable capability
type Skill struct {
	ID            string
	Name          string
	Description   string
	Category      SkillCategory
	Proficiency   float64 // Advisory evaluator EMA alias; never a mastery certificate.
	Competence    SkillCompetenceEstimate
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
	EvidenceID  string
	Status      string // unknown (generated rehearsal), passed, or failed.
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

	system := &SkillLearningSystem{
		ctx:             ctx,
		cancel:          cancel,
		skills:          make(map[string]*Skill),
		practiceQueue:   make([]*SkillPracticeTask, 0),
		llmProvider:     llmProvider,
		appliedEvidence: make(map[string]skillEvidencePayload),
	}
	system.initializeFoundationalSkills()
	return system
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

	// Idempotently ensure the foundational registry is present.
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
			Competence:    newSkillCompetenceEstimate(),
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
		skillID := stableSkillID(fs.name)
		if _, exists := sls.skills[skillID]; exists {
			continue
		}

		sls.skills[skillID] = &Skill{
			ID:            skillID,
			Name:          fs.name,
			Description:   fs.description,
			Category:      fs.category,
			Proficiency:   0, // Unknown until evaluator evidence is observed.
			Competence:    newSkillCompetenceEstimate(),
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

func stableSkillID(name string) string {
	value := strings.ToLower(strings.TrimSpace(name))
	value = strings.NewReplacer(" ", "_", "-", "_", "/", "_").Replace(value)
	return "skill_" + value
}

// PracticeSkill executes a practice session for a skill
func (sls *SkillLearningSystem) PracticeSkill(skillID string) error {
	sls.mu.Lock()
	skill, exists := sls.skills[skillID]
	if !exists {
		sls.mu.Unlock()
		return fmt.Errorf("skill not found: %s", skillID)
	}
	snapshot := cloneSkill(skill)
	provider := sls.llmProvider
	ctx, cancel := context.WithTimeout(sls.ctx, 30*time.Second)
	defer cancel()
	sls.activePractices++
	sls.practicing = true
	sls.mu.Unlock()
	defer func() {
		sls.mu.Lock()
		sls.activePractices--
		sls.practicing = sls.activePractices > 0
		sls.mu.Unlock()
	}()

	startTime := time.Now()

	// Generate practice task
	prompt := fmt.Sprintf(`You are practicing the skill: %s
Description: %s
Advisory evaluator estimate (not mastery): %.2f

Generate a practice exercise for this skill and attempt to complete it.
Provide your attempt; do not grade your own competence.`, snapshot.Name, snapshot.Description, snapshot.Proficiency)

	opts := llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   300,
	}

	if provider == nil {
		return fmt.Errorf("practice provider unavailable")
	}
	result, err := provider.Generate(ctx, prompt, opts)
	if err != nil {
		return fmt.Errorf("practice generation failed: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("practice generation cancelled: %w", err)
	}
	result = strings.TrimSpace(result)
	if result == "" {
		return fmt.Errorf("practice generation returned empty output")
	}

	// A generated rehearsal is useful context, but it is not outcome evidence.
	// Record it without changing proficiency. Only ApplyEvaluatorEvidence may
	// award skill growth from an observed deterministic evaluation.
	attempt := SkillAttempt{
		Timestamp:   time.Now(),
		Success:     false,
		Performance: 0,
		Feedback:    "unverified rehearsal candidate: " + result,
		Duration:    time.Since(startTime),
		Status:      "unknown",
	}

	sls.mu.Lock()
	skill.Attempts = append(skill.Attempts, attempt)
	skill.PracticeCount++
	skill.LastPracticed = time.Now()
	sls.totalPractices++
	proficiency := skill.Proficiency
	sls.mu.Unlock()

	fmt.Printf("🎯 Rehearsed: %s (Proficiency unchanged at %.2f; awaiting evaluator evidence)\n",
		snapshot.Name, proficiency)

	return nil
}

// ApplyEvaluatorEvidence applies one observed, independently evaluated result.
// Replaying the same evidence ID is an idempotent no-op.
func (sls *SkillLearningSystem) ApplyEvaluatorEvidence(
	skillName, evidenceID string,
	score float64,
	passed bool,
	feedback string,
) (bool, error) {
	skillName = strings.TrimSpace(skillName)
	evidenceID = strings.TrimSpace(evidenceID)
	if skillName == "" || evidenceID == "" {
		return false, fmt.Errorf("skill name and evidence ID are required")
	}
	if math.IsNaN(score) || math.IsInf(score, 0) || score < 0 || score > 1 {
		return false, fmt.Errorf("evaluator score must be finite and between 0 and 1")
	}
	payload := skillEvidencePayload{skillName, score, passed, strings.TrimSpace(feedback)}

	sls.mu.Lock()
	defer sls.mu.Unlock()
	if sls.appliedEvidence == nil {
		sls.appliedEvidence = make(map[string]skillEvidencePayload)
	}
	if previous, exists := sls.appliedEvidence[evidenceID]; exists {
		if previous != payload {
			return false, fmt.Errorf("evidence ID %q conflicts with its recorded payload", evidenceID)
		}
		return false, nil
	}

	skillID := stableSkillID(skillName)
	skill, exists := sls.skills[skillID]
	if !exists {
		skill = &Skill{
			ID:            skillID,
			Name:          skillName,
			Description:   "Evidence-grounded capability: " + skillName,
			Category:      SkillCategoryTechnical,
			Proficiency:   0,
			Competence:    newSkillCompetenceEstimate(),
			CreatedAt:     time.Now(),
			LearningRate:  0.1,
			Difficulty:    0.5,
			Prerequisites: make([]string, 0),
			Attempts:      make([]SkillAttempt, 0),
		}
		sls.skills[skillID] = skill
	}
	if skill.Name != skillName {
		return false, fmt.Errorf("skill name %q collides with registry entry %q", skillName, skill.Name)
	}

	attempt := SkillAttempt{
		Timestamp:   time.Now(),
		Success:     passed,
		Performance: score,
		Feedback:    fmt.Sprintf("evidence=%s: %s", evidenceID, payload.Feedback),
		EvidenceID:  evidenceID,
		Status:      "failed",
	}
	if passed {
		attempt.Status = "passed"
	}
	skill.Attempts = append(skill.Attempts, attempt)
	skill.PracticeCount++
	skill.LastPracticed = attempt.Timestamp
	if skill.Competence.Version != SkillCompetenceVersion {
		// Legacy proficiency is not evidence and cannot seed the estimator.
		skill.Competence = newSkillCompetenceEstimate()
	}
	estimate := &skill.Competence
	estimate.Evaluations++
	estimate.ScoreSum += score
	estimate.MeanScore = estimate.ScoreSum / float64(estimate.Evaluations)
	effectiveScore := 0.0
	if passed {
		estimate.Passes++
		effectiveScore = score
	} else {
		estimate.Failures++
	}
	estimate.EffectiveScoreEMA = (1-SkillCompetenceEMAWeight)*estimate.EffectiveScoreEMA + SkillCompetenceEMAWeight*effectiveScore
	estimate.PassPosteriorMean = (1 + float64(estimate.Passes)) / (2 + float64(estimate.Evaluations))
	skill.Proficiency = estimate.EffectiveScoreEMA
	sls.totalPractices++
	sls.appliedEvidence[evidenceID] = payload
	return true, nil
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
			selectedSkill = cloneSkill(skill)
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
		"total_skills":        len(sls.skills),
		"total_practices":     sls.totalPractices,
		"total_masteries":     uint64(0), // Retained for compatibility; evaluations do not certify mastery.
		"avg_proficiency":     avgProficiency,
		"competence_version":  SkillCompetenceVersion,
		"competence_advisory": true,
	}
}

// GetSkillProfile returns a summary of current skills
func (sls *SkillLearningSystem) GetSkillProfile() string {
	sls.mu.RLock()
	defer sls.mu.RUnlock()

	profile := "Current Skill Profile:\n"

	for _, skill := range sls.skills {
		status := "Unknown; no evaluator evidence"
		if skill.Competence.Evaluations > 0 {
			status = "Advisory evaluator estimate; transfer unverified"
		}

		profile += fmt.Sprintf("- %s: %.2f (%s) [%d practices; %d passes, %d failures]\n",
			skill.Name, skill.Proficiency, status, skill.PracticeCount, skill.Competence.Passes, skill.Competence.Failures)
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
		// Even high evaluator scores do not establish universal mastery.
		needing = append(needing, cloneSkill(skill))
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
