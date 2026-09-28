package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// SkillGoalIntegration connects skill learning with goal generation
// enabling autonomous skill acquisition based on identified needs and interests
type SkillGoalIntegration struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// LLM provider
	llmProvider llm.LLMProvider

	// Connected systems
	skillLearning    *SkillLearningSystem
	goalGenerator    *GoalGenerator
	interestPatterns *InterestPatternSystem

	// Skill-goal mappings
	skillGoalMap map[string][]string // skill -> goals that require it
	goalSkillMap map[string][]string // goal -> skills needed

	// Learning queue
	learningQueue  []SkillLearningTask
	activeLearning *SkillLearningTask

	// Practice scheduling
	practiceSchedule []PracticeSession
	lastPractice     time.Time

	// Progress tracking
	skillProgress map[string]*SkillProgressTracker

	// Callbacks
	onSkillGoalCreated  func(skill string, goal ScheduledGoal)
	onPracticeScheduled func(session PracticeSession)
	onSkillMastered     func(skill string, level float64)

	// Metrics
	totalSkillsLearned    uint64
	totalPracticeSessions uint64
	totalGoalsFromSkills  uint64

	// Running state
	running bool
}

// SkillLearningTask represents a task to learn a skill
type SkillLearningTask struct {
	ID             string
	SkillName      string
	Reason         string
	Priority       float64
	SourceGoal     string
	SourceInterest string
	CreatedAt      time.Time
	StartedAt      *time.Time
	CompletedAt    *time.Time
	Status         LearningStatus
	Progress       float64
}

// LearningStatus represents the status of a learning task
type LearningStatus int

const (
	LearningQueued LearningStatus = iota
	LearningActive
	LearningPaused
	LearningCompleted
	LearningAbandoned
)

func (ls LearningStatus) String() string {
	return [...]string{"Queued", "Active", "Paused", "Completed", "Abandoned"}[ls]
}

// PracticeSession represents a scheduled practice session
type PracticeSession struct {
	ID          string
	SkillName   string
	ScheduledAt time.Time
	Duration    time.Duration
	Exercises   []PracticeExercise
	Completed   bool
	Performance float64
}

// PracticeExercise represents an exercise within a practice session
type PracticeExercise struct {
	ID          string
	Description string
	Difficulty  float64
	Completed   bool
	Score       float64
}

// SkillProgressTracker tracks progress for a specific skill
type SkillProgressTracker struct {
	SkillName         string
	CurrentLevel      float64
	TargetLevel       float64
	PracticeCount     int
	TotalPracticeTime time.Duration
	LastPractice      time.Time
	LearningRate      float64
	Milestones        []SkillMilestone
}

// SkillMilestone represents a milestone in skill development
type SkillMilestone struct {
	Level       float64
	Description string
	ReachedAt   *time.Time
}

// NewSkillGoalIntegration creates a new skill-goal integration system
func NewSkillGoalIntegration(
	llmProvider llm.LLMProvider,
	skillLearning *SkillLearningSystem,
	interestPatterns *InterestPatternSystem,
) *SkillGoalIntegration {
	ctx, cancel := context.WithCancel(context.Background())

	return &SkillGoalIntegration{
		ctx:              ctx,
		cancel:           cancel,
		llmProvider:      llmProvider,
		skillLearning:    skillLearning,
		interestPatterns: interestPatterns,
		skillGoalMap:     make(map[string][]string),
		goalSkillMap:     make(map[string][]string),
		learningQueue:    make([]SkillLearningTask, 0),
		practiceSchedule: make([]PracticeSession, 0),
		skillProgress:    make(map[string]*SkillProgressTracker),
	}
}

// Start begins the skill-goal integration system
func (sgi *SkillGoalIntegration) Start() error {
	sgi.mu.Lock()
	if sgi.running {
		sgi.mu.Unlock()
		return fmt.Errorf("skill-goal integration already running")
	}
	sgi.running = true
	sgi.mu.Unlock()

	fmt.Println("📚 Skill-Goal Integration starting...")

	// Start learning management loop
	go sgi.learningManagementLoop()

	// Start practice scheduling loop
	go sgi.practiceSchedulingLoop()

	// Start progress monitoring loop
	go sgi.progressMonitoringLoop()

	return nil
}

// Stop stops the skill-goal integration system
func (sgi *SkillGoalIntegration) Stop() error {
	sgi.mu.Lock()
	defer sgi.mu.Unlock()

	if !sgi.running {
		return fmt.Errorf("skill-goal integration not running")
	}

	sgi.running = false
	sgi.cancel()

	fmt.Println("📚 Skill-Goal Integration stopped")
	fmt.Printf("   Skills learned: %d\n", sgi.totalSkillsLearned)
	fmt.Printf("   Practice sessions: %d\n", sgi.totalPracticeSessions)

	return nil
}

// IdentifySkillsForGoal identifies skills needed to achieve a goal
func (sgi *SkillGoalIntegration) IdentifySkillsForGoal(goal ScheduledGoal) []string {
	sgi.mu.Lock()
	defer sgi.mu.Unlock()

	// Check cache first
	if skills, exists := sgi.goalSkillMap[goal.ID]; exists {
		return skills
	}

	// Use LLM to identify required skills
	prompt := fmt.Sprintf(`[System: You are a skill analysis system. Identify skills needed to achieve a goal.]

Goal: %s
Description: %s

What skills are needed to achieve this goal? List 1-3 specific, learnable skills.
Format each skill on a new line, e.g.:
- Programming in Go
- Understanding distributed systems
- Technical writing

Skills:`, goal.ID, goal.Description)

	opts := llm.GenerateOptions{
		Temperature: 0.5,
		MaxTokens:   100,
	}

	result, err := sgi.llmProvider.Generate(context.Background(), prompt, opts)
	if err != nil {
		return []string{}
	}

	// Parse skills from result
	skills := parseSkillList(result)

	// Cache the mapping
	sgi.goalSkillMap[goal.ID] = skills
	for _, skill := range skills {
		sgi.skillGoalMap[skill] = append(sgi.skillGoalMap[skill], goal.ID)
	}

	return skills
}

// CreateSkillAcquisitionGoal creates a goal for learning a skill
func (sgi *SkillGoalIntegration) CreateSkillAcquisitionGoal(skillName string, reason string, priority float64) ScheduledGoal {
	goal := ScheduledGoal{
		ID:          fmt.Sprintf("skill_goal_%s_%d", skillName, time.Now().UnixNano()),
		Description: fmt.Sprintf("Learn and develop proficiency in: %s", skillName),
		Priority:    priority,
		Status:      GoalPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create learning task
	task := SkillLearningTask{
		ID:         fmt.Sprintf("learn_%s_%d", skillName, time.Now().UnixNano()),
		SkillName:  skillName,
		Reason:     reason,
		Priority:   priority,
		SourceGoal: goal.ID,
		CreatedAt:  time.Now(),
		Status:     LearningQueued,
	}

	sgi.mu.Lock()
	sgi.learningQueue = append(sgi.learningQueue, task)
	sgi.totalGoalsFromSkills++
	sgi.mu.Unlock()

	fmt.Printf("📚 Skill acquisition goal created: %s (priority: %.2f)\n", skillName, priority)

	// Notify callback
	if sgi.onSkillGoalCreated != nil {
		go sgi.onSkillGoalCreated(skillName, goal)
	}

	return goal
}

// QueueSkillFromInterest queues a skill for learning based on interest
func (sgi *SkillGoalIntegration) QueueSkillFromInterest(interest string, strength float64) {
	// Determine relevant skill from interest
	skill := sgi.interestToSkill(interest)
	if skill == "" {
		return
	}

	// Check if already queued or learning
	sgi.mu.RLock()
	for _, task := range sgi.learningQueue {
		if task.SkillName == skill {
			sgi.mu.RUnlock()
			return
		}
	}
	sgi.mu.RUnlock()

	// Create learning task
	task := SkillLearningTask{
		ID:             fmt.Sprintf("learn_%s_%d", skill, time.Now().UnixNano()),
		SkillName:      skill,
		Reason:         fmt.Sprintf("Interest in: %s", interest),
		Priority:       strength,
		SourceInterest: interest,
		CreatedAt:      time.Now(),
		Status:         LearningQueued,
	}

	sgi.mu.Lock()
	sgi.learningQueue = append(sgi.learningQueue, task)
	sgi.mu.Unlock()

	fmt.Printf("📚 Skill queued from interest: %s (strength: %.2f)\n", skill, strength)
}

// interestToSkill maps an interest to a learnable skill
func (sgi *SkillGoalIntegration) interestToSkill(interest string) string {
	// Simple mapping - in production, use LLM or knowledge base
	interestSkillMap := map[string]string{
		"programming":     "Software Development",
		"AI":              "Machine Learning",
		"philosophy":      "Critical Thinking",
		"writing":         "Technical Writing",
		"communication":   "Effective Communication",
		"problem-solving": "Analytical Problem Solving",
		"creativity":      "Creative Thinking",
		"learning":        "Meta-Learning",
	}

	for key, skill := range interestSkillMap {
		if contains(interest, key) {
			return skill
		}
	}

	return ""
}

// learningManagementLoop manages the learning queue
func (sgi *SkillGoalIntegration) learningManagementLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sgi.ctx.Done():
			return
		case <-ticker.C:
			sgi.processLearningQueue()
		}
	}
}

// processLearningQueue processes the learning queue
func (sgi *SkillGoalIntegration) processLearningQueue() {
	sgi.mu.Lock()
	defer sgi.mu.Unlock()

	// Check if we have an active learning task
	if sgi.activeLearning != nil {
		// Update progress
		sgi.updateLearningProgress()
		return
	}

	// Find highest priority task
	if len(sgi.learningQueue) == 0 {
		return
	}

	// Sort by priority (simple selection)
	highestIdx := 0
	highestPriority := sgi.learningQueue[0].Priority
	for i, task := range sgi.learningQueue {
		if task.Priority > highestPriority {
			highestIdx = i
			highestPriority = task.Priority
		}
	}

	// Start learning
	task := &sgi.learningQueue[highestIdx]
	now := time.Now()
	task.StartedAt = &now
	task.Status = LearningActive
	sgi.activeLearning = task

	// Initialize progress tracker
	if _, exists := sgi.skillProgress[task.SkillName]; !exists {
		sgi.skillProgress[task.SkillName] = &SkillProgressTracker{
			SkillName:    task.SkillName,
			CurrentLevel: 0.0,
			TargetLevel:  1.0,
			LearningRate: 0.1,
			Milestones: []SkillMilestone{
				{Level: 0.25, Description: "Beginner"},
				{Level: 0.50, Description: "Intermediate"},
				{Level: 0.75, Description: "Advanced"},
				{Level: 1.00, Description: "Expert"},
			},
		}
	}

	fmt.Printf("📚 Started learning: %s\n", task.SkillName)
}

// updateLearningProgress updates progress on active learning
func (sgi *SkillGoalIntegration) updateLearningProgress() {
	if sgi.activeLearning == nil {
		return
	}

	task := sgi.activeLearning
	tracker := sgi.skillProgress[task.SkillName]

	// Simulate learning progress
	progressIncrement := tracker.LearningRate * 0.05
	tracker.CurrentLevel = min(tracker.CurrentLevel+progressIncrement, tracker.TargetLevel)
	task.Progress = tracker.CurrentLevel

	// Check for milestone completion
	for i, milestone := range tracker.Milestones {
		if milestone.ReachedAt == nil && tracker.CurrentLevel >= milestone.Level {
			now := time.Now()
			tracker.Milestones[i].ReachedAt = &now
			fmt.Printf("📚 Milestone reached: %s - %s (%.0f%%)\n",
				task.SkillName, milestone.Description, milestone.Level*100)
		}
	}

	// Check for completion
	if tracker.CurrentLevel >= tracker.TargetLevel {
		now := time.Now()
		task.CompletedAt = &now
		task.Status = LearningCompleted
		sgi.totalSkillsLearned++

		fmt.Printf("📚 Skill mastered: %s\n", task.SkillName)

		// Notify callback
		if sgi.onSkillMastered != nil {
			go sgi.onSkillMastered(task.SkillName, tracker.CurrentLevel)
		}

		// Remove from queue and clear active
		sgi.removeFromQueue(task.ID)
		sgi.activeLearning = nil
	}
}

// removeFromQueue removes a task from the learning queue
func (sgi *SkillGoalIntegration) removeFromQueue(taskID string) {
	newQueue := make([]SkillLearningTask, 0, len(sgi.learningQueue)-1)
	for _, task := range sgi.learningQueue {
		if task.ID != taskID {
			newQueue = append(newQueue, task)
		}
	}
	sgi.learningQueue = newQueue
}

// practiceSchedulingLoop schedules practice sessions
func (sgi *SkillGoalIntegration) practiceSchedulingLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-sgi.ctx.Done():
			return
		case <-ticker.C:
			sgi.schedulePractice()
		}
	}
}

// schedulePractice schedules practice sessions for skills being learned
func (sgi *SkillGoalIntegration) schedulePractice() {
	sgi.mu.Lock()
	defer sgi.mu.Unlock()

	// Check if enough time has passed since last practice
	if time.Since(sgi.lastPractice) < 5*time.Minute {
		return
	}

	// Schedule practice for active learning
	if sgi.activeLearning != nil {
		session := sgi.createPracticeSession(sgi.activeLearning.SkillName)
		sgi.practiceSchedule = append(sgi.practiceSchedule, session)
		sgi.lastPractice = time.Now()
		sgi.totalPracticeSessions++

		fmt.Printf("📚 Practice session scheduled: %s\n", session.SkillName)

		// Notify callback
		if sgi.onPracticeScheduled != nil {
			go sgi.onPracticeScheduled(session)
		}

		// Execute practice (simplified)
		go sgi.executePractice(session)
	}
}

// createPracticeSession creates a practice session for a skill
func (sgi *SkillGoalIntegration) createPracticeSession(skillName string) PracticeSession {
	tracker := sgi.skillProgress[skillName]
	difficulty := 0.3
	if tracker != nil {
		difficulty = tracker.CurrentLevel + 0.1
	}

	exercises := []PracticeExercise{
		{
			ID:          fmt.Sprintf("ex1_%d", time.Now().UnixNano()),
			Description: fmt.Sprintf("Basic %s exercise", skillName),
			Difficulty:  difficulty * 0.7,
		},
		{
			ID:          fmt.Sprintf("ex2_%d", time.Now().UnixNano()),
			Description: fmt.Sprintf("Intermediate %s challenge", skillName),
			Difficulty:  difficulty,
		},
		{
			ID:          fmt.Sprintf("ex3_%d", time.Now().UnixNano()),
			Description: fmt.Sprintf("Advanced %s problem", skillName),
			Difficulty:  difficulty * 1.3,
		},
	}

	return PracticeSession{
		ID:          fmt.Sprintf("practice_%s_%d", skillName, time.Now().UnixNano()),
		SkillName:   skillName,
		ScheduledAt: time.Now(),
		Duration:    15 * time.Minute,
		Exercises:   exercises,
	}
}

// executePractice executes a practice session
func (sgi *SkillGoalIntegration) executePractice(session PracticeSession) {
	// Simulate practice execution
	time.Sleep(2 * time.Second)

	sgi.mu.Lock()
	defer sgi.mu.Unlock()

	// Update session results
	for i := range session.Exercises {
		session.Exercises[i].Completed = true
		session.Exercises[i].Score = 0.7 + randFloat()*0.3
	}
	session.Completed = true

	// Calculate overall performance
	totalScore := 0.0
	for _, ex := range session.Exercises {
		totalScore += ex.Score
	}
	session.Performance = totalScore / float64(len(session.Exercises))

	// Update skill progress
	if tracker, exists := sgi.skillProgress[session.SkillName]; exists {
		tracker.PracticeCount++
		tracker.TotalPracticeTime += session.Duration
		tracker.LastPractice = time.Now()

		// Boost learning rate based on practice performance
		tracker.LearningRate = 0.1 + (session.Performance * 0.1)
	}

	fmt.Printf("📚 Practice completed: %s (performance: %.2f)\n",
		session.SkillName, session.Performance)
}

// progressMonitoringLoop monitors skill progress
func (sgi *SkillGoalIntegration) progressMonitoringLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-sgi.ctx.Done():
			return
		case <-ticker.C:
			sgi.reportProgress()
		}
	}
}

// reportProgress reports on skill learning progress
func (sgi *SkillGoalIntegration) reportProgress() {
	sgi.mu.RLock()
	defer sgi.mu.RUnlock()

	if len(sgi.skillProgress) == 0 {
		return
	}

	fmt.Println("\n📚 Skill Learning Progress Report:")
	for name, tracker := range sgi.skillProgress {
		fmt.Printf("   %s: %.0f%% (practices: %d, rate: %.2f)\n",
			name, tracker.CurrentLevel*100, tracker.PracticeCount, tracker.LearningRate)
	}
}

// SetCallbacks sets callback functions
func (sgi *SkillGoalIntegration) SetCallbacks(
	onGoalCreated func(string, ScheduledGoal),
	onPracticeScheduled func(PracticeSession),
	onSkillMastered func(string, float64),
) {
	sgi.mu.Lock()
	defer sgi.mu.Unlock()

	sgi.onSkillGoalCreated = onGoalCreated
	sgi.onPracticeScheduled = onPracticeScheduled
	sgi.onSkillMastered = onSkillMastered
}

// GetLearningQueue returns the current learning queue
func (sgi *SkillGoalIntegration) GetLearningQueue() []SkillLearningTask {
	sgi.mu.RLock()
	defer sgi.mu.RUnlock()

	result := make([]SkillLearningTask, len(sgi.learningQueue))
	copy(result, sgi.learningQueue)
	return result
}

// GetSkillProgress returns progress for a specific skill
func (sgi *SkillGoalIntegration) GetSkillProgress(skillName string) *SkillProgressTracker {
	sgi.mu.RLock()
	defer sgi.mu.RUnlock()

	return sgi.skillProgress[skillName]
}

// GetMetrics returns integration metrics
func (sgi *SkillGoalIntegration) GetMetrics() map[string]interface{} {
	sgi.mu.RLock()
	defer sgi.mu.RUnlock()

	return map[string]interface{}{
		"skills_learned":     sgi.totalSkillsLearned,
		"practice_sessions":  sgi.totalPracticeSessions,
		"goals_from_skills":  sgi.totalGoalsFromSkills,
		"queue_length":       len(sgi.learningQueue),
		"active_learning":    sgi.activeLearning != nil,
		"skills_in_progress": len(sgi.skillProgress),
		"running":            sgi.running,
	}
}

// Helper function to parse skill list from LLM response
func parseSkillList(response string) []string {
	skills := make([]string, 0)
	lines := splitLines(response)

	for _, line := range lines {
		// Remove bullet points and trim
		line = trimBullet(line)
		if len(line) > 3 && len(line) < 100 {
			skills = append(skills, line)
		}
	}

	return skills
}

func splitLines(s string) []string {
	lines := make([]string, 0)
	current := ""
	for _, c := range s {
		if c == '\n' {
			if current != "" {
				lines = append(lines, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func trimBullet(s string) string {
	// Remove leading whitespace, bullets, dashes
	start := 0
	for start < len(s) {
		r := rune(s[start])
		if r == ' ' || r == '-' || r == '*' || r == '•' {
			start++
		} else {
			break
		}
	}
	return s[start:]
}
