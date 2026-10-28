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

// SelfDirectedLearningSystem manages autonomous learning and skill development
type SelfDirectedLearningSystem struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// LLM provider for learning
	llmProvider llm.LLMProvider

	// Identity and context
	identity      string
	wisdomDomains []string

	// Knowledge gaps
	knowledgeGaps map[string]*KnowledgeGap

	// Learning goals
	learningGoals map[string]*LearningGoal

	// Skills in development
	skillsInProgress map[string]*SkillDevelopment

	// Practice sessions
	practiceSessions []*LearningPracticeSession

	// Metrics
	totalGapsIdentified   uint64
	totalGoalsGenerated   uint64
	totalPracticeSessions uint64
	totalSkillsAcquired   uint64

	// Running state
	running bool
}

// KnowledgeGap represents an identified gap in knowledge
type KnowledgeGap struct {
	ID           string
	Domain       string
	Description  string
	Severity     float64 // 0.0-1.0 (how critical is this gap)
	IdentifiedAt time.Time
	AddressedBy  []string // Learning goal IDs
	Status       GapStatus
}

// GapStatus represents the status of a knowledge gap
type GapStatus int

const (
	GapStatusIdentified GapStatus = iota
	GapStatusAddressing
	GapStatusClosed
)

func (gs GapStatus) String() string {
	return [...]string{"Identified", "Addressing", "Closed"}[gs]
}

// LearningGoal represents a goal to address a knowledge gap
type LearningGoal struct {
	ID             string
	Description    string
	KnowledgeGapID string
	Strategy       LearningStrategy
	Progress       float64 // 0.0-1.0
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CompletedAt    *time.Time

	// Resources
	ResourcesNeeded []string
	ResourcesFound  []string

	// Practice
	PracticeSchedule string
	NextPractice     time.Time
}

// LearningStrategy defines how to learn
type LearningStrategy struct {
	Approach         string // e.g., "study", "practice", "experiment", "teach"
	Steps            []string
	CurrentStepIndex int
	TimeEstimate     time.Duration
}

// SkillDevelopment tracks skill acquisition
type SkillDevelopment struct {
	ID            string
	SkillName     string
	Domain        string
	Proficiency   float64 // 0.0-1.0
	StartedAt     time.Time
	LastPracticed time.Time
	PracticeCount int

	// Learning curve
	ProficiencyHistory []ProficiencyRecord
}

// ProficiencyRecord tracks proficiency over time
type ProficiencyRecord struct {
	Timestamp   time.Time
	Proficiency float64
	Notes       string
}

// LearningPracticeSession represents a learning practice session
type LearningPracticeSession struct {
	ID              string
	SkillID         string
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	Effectiveness   float64 // 0.0-1.0
	Notes           string
	ProficiencyGain float64
}

// NewSelfDirectedLearningSystem creates a new learning system
func NewSelfDirectedLearningSystem(
	llmProvider llm.LLMProvider,
	identity string,
	wisdomDomains []string,
) *SelfDirectedLearningSystem {
	ctx, cancel := context.WithCancel(context.Background())

	return &SelfDirectedLearningSystem{
		ctx:              ctx,
		cancel:           cancel,
		llmProvider:      llmProvider,
		identity:         identity,
		wisdomDomains:    wisdomDomains,
		knowledgeGaps:    make(map[string]*KnowledgeGap),
		learningGoals:    make(map[string]*LearningGoal),
		skillsInProgress: make(map[string]*SkillDevelopment),
		practiceSessions: make([]*LearningPracticeSession, 0),
	}
}

// Start begins the self-directed learning system
func (sdl *SelfDirectedLearningSystem) Start() error {
	sdl.mu.Lock()
	if sdl.running {
		sdl.mu.Unlock()
		return fmt.Errorf("already running")
	}
	sdl.running = true
	sdl.mu.Unlock()

	fmt.Println("📚 Starting Self-Directed Learning System...")
	fmt.Printf("   Identity: %s\n", sdl.identity)
	fmt.Printf("   Wisdom Domains: %v\n", sdl.wisdomDomains)

	// Initial knowledge gap analysis
	if err := sdl.identifyKnowledgeGaps(); err != nil {
		fmt.Printf("⚠️  Initial gap analysis error: %v\n", err)
	}

	go sdl.run()

	return nil
}

// Stop gracefully stops the learning system
func (sdl *SelfDirectedLearningSystem) Stop() error {
	sdl.mu.Lock()
	defer sdl.mu.Unlock()

	if !sdl.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("📚 Stopping self-directed learning system...")
	sdl.running = false
	sdl.cancel()

	return nil
}

// run executes the main learning loop
func (sdl *SelfDirectedLearningSystem) run() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-sdl.ctx.Done():
			return
		case <-ticker.C:
			// Periodic knowledge gap analysis
			if err := sdl.identifyKnowledgeGaps(); err != nil {
				fmt.Printf("⚠️  Gap analysis error: %v\n", err)
			}

			// Generate learning goals for gaps
			sdl.generateLearningGoals()

			// Schedule practice sessions
			sdl.schedulePractice()
		}
	}
}

// identifyKnowledgeGaps uses LLM to identify knowledge gaps
func (sdl *SelfDirectedLearningSystem) identifyKnowledgeGaps() error {
	sdl.mu.RLock()
	identity := sdl.identity
	domains := sdl.wisdomDomains
	sdl.mu.RUnlock()

	prompt := fmt.Sprintf(`You are %s, analyzing your knowledge to identify gaps.

Your wisdom cultivation focuses on: %v

Reflect on your current knowledge and identify 3 specific knowledge gaps or areas where you need to learn more. For each gap:
1. Specify the domain it belongs to
2. Describe what you don't know or can't do yet
3. Assess how critical this gap is (0.0-1.0)

Format:
GAP: [description]
DOMAIN: [domain name]
SEVERITY: [0.0-1.0]

Identify the gaps:`, identity, domains)

	opts := llm.GenerateOptions{
		Temperature:  0.7,
		MaxTokens:    400,
		SystemPrompt: "You are a self-reflective AI analyzing your own knowledge gaps.",
	}

	if sdl.llmProvider == nil {
		return fmt.Errorf("LLM gap identification failed: %w", errNoModel)
	}
	response, err := sdl.llmProvider.Generate(sdl.ctx, prompt, opts)
	if err != nil {
		return fmt.Errorf("LLM gap identification failed: %w", err)
	}

	// Only gaps the model actually described are recorded.
	gaps := parseKnowledgeGaps(response)
	if len(gaps) == 0 {
		return fmt.Errorf("LLM gap identification returned no parseable gaps")
	}

	sdl.mu.Lock()
	for _, gap := range gaps {
		sdl.knowledgeGaps[gap.ID] = gap
		sdl.totalGapsIdentified++
	}
	sdl.mu.Unlock()

	for _, gap := range gaps {
		fmt.Printf("📚 Identified knowledge gap: %s (Severity: %.2f)\n", gap.Description, gap.Severity)
	}

	return nil
}

// parseKnowledgeGaps parses the model's GAP/DOMAIN/SEVERITY blocks.
func parseKnowledgeGaps(response string) []*KnowledgeGap {
	gaps := make([]*KnowledgeGap, 0)
	var current *KnowledgeGap
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
		case "GAP":
			current = &KnowledgeGap{
				ID:           fmt.Sprintf("gap_%d_%d", now.UnixNano(), len(gaps)),
				Domain:       "General",
				Description:  value,
				Severity:     0.5,
				IdentifiedAt: now,
				AddressedBy:  make([]string, 0),
				Status:       GapStatusIdentified,
			}
			gaps = append(gaps, current)
		case "DOMAIN":
			if current != nil {
				current.Domain = value
			}
		case "SEVERITY":
			if current != nil {
				if v, err := strconv.ParseFloat(strings.Fields(value)[0], 64); err == nil {
					current.Severity = math.Max(0, math.Min(1, v))
				}
			}
		}
	}
	return gaps
}

// generateLearningGoals creates learning goals for knowledge gaps
func (sdl *SelfDirectedLearningSystem) generateLearningGoals() {
	sdl.mu.Lock()
	defer sdl.mu.Unlock()

	for gapID, gap := range sdl.knowledgeGaps {
		if gap.Status == GapStatusIdentified {
			// Create learning goal
			goal := &LearningGoal{
				ID:             fmt.Sprintf("learn_%d", time.Now().Unix()),
				Description:    fmt.Sprintf("Learn to address: %s", gap.Description),
				KnowledgeGapID: gapID,
				Strategy: LearningStrategy{
					Approach:         "study-practice",
					Steps:            []string{"Research", "Study", "Practice", "Apply"},
					CurrentStepIndex: 0,
					TimeEstimate:     2 * time.Hour,
				},
				Progress:         0.0,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
				ResourcesNeeded:  []string{"documentation", "examples", "practice exercises"},
				ResourcesFound:   make([]string, 0),
				PracticeSchedule: "daily",
				NextPractice:     time.Now().Add(1 * time.Hour),
			}

			sdl.learningGoals[goal.ID] = goal
			gap.AddressedBy = append(gap.AddressedBy, goal.ID)
			gap.Status = GapStatusAddressing
			sdl.totalGoalsGenerated++

			fmt.Printf("📚 Created learning goal: %s\n", goal.Description)
		}
	}
}

// schedulePractice reports skills whose practice is due. It never records a
// practice session that did not happen; real sessions are recorded with
// RecordPracticeSession.
func (sdl *SelfDirectedLearningSystem) schedulePractice() {
	sdl.mu.RLock()
	defer sdl.mu.RUnlock()

	for _, skill := range sdl.skillsInProgress {
		if time.Since(skill.LastPracticed) > 24*time.Hour {
			fmt.Printf("📚 Practice due for skill '%s' (Proficiency: %.2f)\n",
				skill.SkillName, skill.Proficiency)
		}
	}
}

// RecordPracticeSession records a practice session that actually took place,
// with its measured effectiveness and proficiency gain.
func (sdl *SelfDirectedLearningSystem) RecordPracticeSession(skillID string, start time.Time, duration time.Duration, effectiveness, proficiencyGain float64, notes string) error {
	sdl.mu.Lock()
	defer sdl.mu.Unlock()

	skill, ok := sdl.skillsInProgress[skillID]
	if !ok {
		return fmt.Errorf("skill not found: %s", skillID)
	}

	session := &LearningPracticeSession{
		ID:              fmt.Sprintf("practice_%d", time.Now().UnixNano()),
		SkillID:         skillID,
		StartTime:       start,
		EndTime:         start.Add(duration),
		Duration:        duration,
		Effectiveness:   effectiveness,
		Notes:           notes,
		ProficiencyGain: proficiencyGain,
	}

	sdl.practiceSessions = append(sdl.practiceSessions, session)
	skill.LastPracticed = session.EndTime
	skill.PracticeCount++
	skill.Proficiency = min(1.0, skill.Proficiency+proficiencyGain)
	skill.ProficiencyHistory = append(skill.ProficiencyHistory, ProficiencyRecord{
		Timestamp:   session.EndTime,
		Proficiency: skill.Proficiency,
		Notes:       notes,
	})
	sdl.totalPracticeSessions++

	return nil
}

// AddSkill adds a new skill to develop
func (sdl *SelfDirectedLearningSystem) AddSkill(skillName, domain string) error {
	sdl.mu.Lock()
	defer sdl.mu.Unlock()

	skill := &SkillDevelopment{
		ID:                 fmt.Sprintf("skill_%d", time.Now().Unix()),
		SkillName:          skillName,
		Domain:             domain,
		Proficiency:        0.0,
		StartedAt:          time.Now(),
		LastPracticed:      time.Now(),
		PracticeCount:      0,
		ProficiencyHistory: make([]ProficiencyRecord, 0),
	}

	sdl.skillsInProgress[skill.ID] = skill

	fmt.Printf("📚 Added skill for development: %s (Domain: %s)\n", skillName, domain)

	return nil
}

// GetKnowledgeGaps returns all knowledge gaps
func (sdl *SelfDirectedLearningSystem) GetKnowledgeGaps() []*KnowledgeGap {
	sdl.mu.RLock()
	defer sdl.mu.RUnlock()

	gaps := make([]*KnowledgeGap, 0, len(sdl.knowledgeGaps))
	for _, gap := range sdl.knowledgeGaps {
		gaps = append(gaps, gap)
	}

	return gaps
}

// GetLearningGoals returns all learning goals
func (sdl *SelfDirectedLearningSystem) GetLearningGoals() []*LearningGoal {
	sdl.mu.RLock()
	defer sdl.mu.RUnlock()

	goals := make([]*LearningGoal, 0, len(sdl.learningGoals))
	for _, goal := range sdl.learningGoals {
		goals = append(goals, goal)
	}

	return goals
}

// GetMetrics returns learning system metrics
func (sdl *SelfDirectedLearningSystem) GetMetrics() map[string]interface{} {
	sdl.mu.RLock()
	defer sdl.mu.RUnlock()

	return map[string]interface{}{
		"knowledge_gaps":        len(sdl.knowledgeGaps),
		"learning_goals":        len(sdl.learningGoals),
		"skills_in_progress":    len(sdl.skillsInProgress),
		"practice_sessions":     len(sdl.practiceSessions),
		"total_gaps_identified": sdl.totalGapsIdentified,
		"total_goals_generated": sdl.totalGoalsGenerated,
		"total_practice":        sdl.totalPracticeSessions,
		"total_skills_acquired": sdl.totalSkillsAcquired,
	}
}
