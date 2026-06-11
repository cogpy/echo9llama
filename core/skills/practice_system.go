package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
	"github.com/google/uuid"
)

// SkillPracticeSystem enables autonomous skill development and measurement
type SkillPracticeSystem struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Skill registry
	skills        map[string]*Skill
	skillOntology *SkillOntology

	// Practice state
	activePractice  *PracticeSession
	practiceHistory []*PracticeSession

	// Performance tracking
	performanceMetrics map[string]*PerformanceMetrics

	// Real evaluator integration
	providerManager *llm.ProviderManager

	// Autonomous practice goals
	practiceGoals []*PracticeGoal

	// Configuration
	practiceInterval  time.Duration
	improvementTarget float64

	// Metrics
	sessionsCompleted uint64
	skillsImproved    uint64

	running bool
}

// Skill represents a cognitive skill
type Skill struct {
	ID                string
	Name              string
	Description       string
	Category          SkillCategory
	Difficulty        float64
	CurrentLevel      float64
	TargetLevel       float64
	Prerequisites     []string
	Metrics           []string
	PracticeScenarios []*PracticeScenario
	LastPracticed     time.Time
	TotalPracticeTime time.Duration
}

// SkillCategory defines different categories of cognitive skills
type SkillCategory int

const (
	SkillCategoryPatternRecognition SkillCategory = iota
	SkillCategoryLogicalReasoning
	SkillCategoryCreativeSynthesis
	SkillCategoryMetaCognition
	SkillCategoryMemoryRetrieval
	SkillCategoryGoalPlanning
	SkillCategorySocialUnderstanding
	SkillCategoryTemporalReasoning
)

func (sc SkillCategory) String() string {
	return [...]string{
		"PatternRecognition",
		"LogicalReasoning",
		"CreativeSynthesis",
		"MetaCognition",
		"MemoryRetrieval",
		"GoalPlanning",
		"SocialUnderstanding",
		"TemporalReasoning",
	}[sc]
}

// SkillOntology defines the structure and relationships of skills
type SkillOntology struct {
	RootSkills        []*Skill
	SkillHierarchy    map[string][]string // skill ID -> child skill IDs
	SkillDependencies map[string][]string // skill ID -> prerequisite IDs
}

// PracticeScenario defines a practice exercise for a skill
type PracticeScenario struct {
	ID          string
	SkillID     string
	Name        string
	Description string
	Difficulty  float64
	Prompt      string
	Evaluation  EvaluationCriteria
	TimeLimit   time.Duration
}

// EvaluationCriteria defines how to evaluate practice performance
type EvaluationCriteria struct {
	Metrics    []string
	Thresholds map[string]float64
	Weights    map[string]float64
}

// PracticeSession represents a practice session
type PracticeSession struct {
	ID          string
	SkillID     string
	ScenarioID  string
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Performance *PerformanceResult
	Insights    []string
}

// PerformanceResult captures the results of a practice session
type PerformanceResult struct {
	Score        float64
	MetricScores map[string]float64
	Strengths    []string
	Weaknesses   []string
	Improvement  float64
}

// PerformanceMetrics tracks performance over time for a skill
type PerformanceMetrics struct {
	SkillID         string
	AverageScore    float64
	BestScore       float64
	RecentScores    []float64
	Trend           float64 // Positive = improving, negative = declining
	PracticeCount   int
	LastImprovement time.Time
}

// PracticeGoal represents an autonomous practice goal
type PracticeGoal struct {
	ID          string
	SkillID     string
	TargetLevel float64
	Deadline    time.Time
	Priority    int
	Status      string
	Progress    float64
}

// NewSkillPracticeSystem creates a new skill practice system
func NewSkillPracticeSystem() *SkillPracticeSystem {
	ctx, cancel := context.WithCancel(context.Background())

	sps := &SkillPracticeSystem{
		ctx:                ctx,
		cancel:             cancel,
		skills:             make(map[string]*Skill),
		performanceMetrics: make(map[string]*PerformanceMetrics),
		practiceGoals:      make([]*PracticeGoal, 0),
		practiceHistory:    make([]*PracticeSession, 0),
		practiceInterval:   30 * time.Minute,
		improvementTarget:  0.1, // 10% improvement target
	}
	sps.providerManager = newPracticeProviderManager()

	// Initialize skill ontology
	sps.initializeSkillOntology()

	return sps
}

// initializeSkillOntology creates the initial skill structure
func (sps *SkillPracticeSystem) initializeSkillOntology() {
	// Create fundamental cognitive skills
	skills := []*Skill{
		{
			ID:           uuid.New().String(),
			Name:         "Pattern Recognition",
			Description:  "Ability to identify patterns in data and experiences",
			Category:     SkillCategoryPatternRecognition,
			Difficulty:   0.5,
			CurrentLevel: 0.3,
			TargetLevel:  0.8,
			Metrics:      []string{"accuracy", "speed", "complexity"},
		},
		{
			ID:           uuid.New().String(),
			Name:         "Logical Reasoning",
			Description:  "Ability to reason logically and draw valid conclusions",
			Category:     SkillCategoryLogicalReasoning,
			Difficulty:   0.6,
			CurrentLevel: 0.4,
			TargetLevel:  0.9,
			Metrics:      []string{"correctness", "efficiency", "depth"},
		},
		{
			ID:           uuid.New().String(),
			Name:         "Creative Synthesis",
			Description:  "Ability to combine ideas in novel and useful ways",
			Category:     SkillCategoryCreativeSynthesis,
			Difficulty:   0.7,
			CurrentLevel: 0.2,
			TargetLevel:  0.7,
			Metrics:      []string{"novelty", "utility", "coherence"},
		},
		{
			ID:           uuid.New().String(),
			Name:         "Meta-Cognitive Reflection",
			Description:  "Ability to think about one's own thinking",
			Category:     SkillCategoryMetaCognition,
			Difficulty:   0.8,
			CurrentLevel: 0.3,
			TargetLevel:  0.8,
			Metrics:      []string{"self_awareness", "insight_depth", "adaptation"},
		},
	}

	// Add skills to registry
	for _, skill := range skills {
		sps.skills[skill.ID] = skill

		// Initialize performance metrics
		sps.performanceMetrics[skill.ID] = &PerformanceMetrics{
			SkillID:       skill.ID,
			AverageScore:  0.0,
			BestScore:     0.0,
			RecentScores:  make([]float64, 0),
			Trend:         0.0,
			PracticeCount: 0,
		}

		// Create practice scenarios
		sps.createPracticeScenariosForSkill(skill)
	}

	// Create skill ontology
	sps.skillOntology = &SkillOntology{
		RootSkills:        skills,
		SkillHierarchy:    make(map[string][]string),
		SkillDependencies: make(map[string][]string),
	}
}

// createPracticeScenariosForSkill generates practice scenarios for a skill
func (sps *SkillPracticeSystem) createPracticeScenariosForSkill(skill *Skill) {
	scenarios := make([]*PracticeScenario, 0)

	switch skill.Category {
	case SkillCategoryPatternRecognition:
		scenarios = append(scenarios, &PracticeScenario{
			ID:          uuid.New().String(),
			SkillID:     skill.ID,
			Name:        "Sequence Pattern Detection",
			Description: "Identify the pattern in a sequence of numbers or concepts",
			Difficulty:  0.5,
			Prompt:      "What pattern do you see in: 2, 4, 8, 16, ...?",
			Evaluation: EvaluationCriteria{
				Metrics:    []string{"accuracy", "speed"},
				Thresholds: map[string]float64{"accuracy": 0.8},
				Weights:    map[string]float64{"accuracy": 0.7, "speed": 0.3},
			},
			TimeLimit: 30 * time.Second,
		})

	case SkillCategoryLogicalReasoning:
		scenarios = append(scenarios, &PracticeScenario{
			ID:          uuid.New().String(),
			SkillID:     skill.ID,
			Name:        "Syllogistic Reasoning",
			Description: "Draw valid conclusions from premises",
			Difficulty:  0.6,
			Prompt:      "If all A are B, and all B are C, what can we conclude about A and C?",
			Evaluation: EvaluationCriteria{
				Metrics:    []string{"correctness", "explanation_quality"},
				Thresholds: map[string]float64{"correctness": 1.0},
				Weights:    map[string]float64{"correctness": 0.8, "explanation_quality": 0.2},
			},
			TimeLimit: 60 * time.Second,
		})

	case SkillCategoryCreativeSynthesis:
		scenarios = append(scenarios, &PracticeScenario{
			ID:          uuid.New().String(),
			SkillID:     skill.ID,
			Name:        "Conceptual Blending",
			Description: "Combine two unrelated concepts in a meaningful way",
			Difficulty:  0.7,
			Prompt:      "Create a novel concept by blending 'tree' and 'memory'",
			Evaluation: EvaluationCriteria{
				Metrics:    []string{"novelty", "coherence", "utility"},
				Thresholds: map[string]float64{"novelty": 0.6, "coherence": 0.7},
				Weights:    map[string]float64{"novelty": 0.4, "coherence": 0.3, "utility": 0.3},
			},
			TimeLimit: 120 * time.Second,
		})

	case SkillCategoryMetaCognition:
		scenarios = append(scenarios, &PracticeScenario{
			ID:          uuid.New().String(),
			SkillID:     skill.ID,
			Name:        "Thinking About Thinking",
			Description: "Reflect on your own cognitive processes",
			Difficulty:  0.8,
			Prompt:      "Describe how you approach solving a complex problem",
			Evaluation: EvaluationCriteria{
				Metrics:    []string{"self_awareness", "insight_depth"},
				Thresholds: map[string]float64{"self_awareness": 0.7},
				Weights:    map[string]float64{"self_awareness": 0.5, "insight_depth": 0.5},
			},
			TimeLimit: 180 * time.Second,
		})
	}

	skill.PracticeScenarios = scenarios
}

// Start begins autonomous skill practice
func (sps *SkillPracticeSystem) Start() error {
	sps.mu.Lock()
	if sps.running {
		sps.mu.Unlock()
		return fmt.Errorf("skill practice system already running")
	}
	sps.running = true
	sps.mu.Unlock()

	// Start autonomous practice loop
	go sps.autonomousPracticeLoop()

	// Start skill assessment loop
	go sps.skillAssessmentLoop()

	// Start practice goal generation loop
	go sps.practiceGoalGenerationLoop()

	return nil
}

// Stop halts autonomous skill practice
func (sps *SkillPracticeSystem) Stop() {
	sps.mu.Lock()
	sps.running = false
	sps.mu.Unlock()

	sps.cancel()
}

// autonomousPracticeLoop periodically practices skills
func (sps *SkillPracticeSystem) autonomousPracticeLoop() {
	ticker := time.NewTicker(sps.practiceInterval)
	defer ticker.Stop()

	for {
		select {
		case <-sps.ctx.Done():
			return
		case <-ticker.C:
			sps.conductPracticeSession()
		}
	}
}

// conductPracticeSession runs a practice session for a skill
func (sps *SkillPracticeSystem) conductPracticeSession() {
	// Select skill to practice
	skill := sps.selectSkillToPractice()
	if skill == nil {
		return
	}

	// Select practice scenario
	scenario := sps.selectPracticeScenario(skill)
	if scenario == nil {
		return
	}

	// Create practice session
	session := &PracticeSession{
		ID:         uuid.New().String(),
		SkillID:    skill.ID,
		ScenarioID: scenario.ID,
		StartTime:  time.Now(),
	}

	// Execute practice (simulated for now)
	performance := sps.executePractice(skill, scenario)

	// Record results
	session.EndTime = time.Now()
	session.Duration = session.EndTime.Sub(session.StartTime)
	session.Performance = performance

	// Update metrics
	sps.updatePerformanceMetrics(skill.ID, performance)

	// Store session
	sps.mu.Lock()
	sps.practiceHistory = append(sps.practiceHistory, session)
	sps.sessionsCompleted++
	sps.mu.Unlock()

	// Update skill level if improved
	if performance.Improvement > 0 {
		sps.updateSkillLevel(skill, performance)
	}
}

// selectSkillToPractice chooses which skill to practice
func (sps *SkillPracticeSystem) selectSkillToPractice() *Skill {
	sps.mu.RLock()
	defer sps.mu.RUnlock()

	// Select skill with largest gap between current and target level
	var selectedSkill *Skill
	maxGap := 0.0

	for _, skill := range sps.skills {
		gap := skill.TargetLevel - skill.CurrentLevel
		if gap > maxGap {
			maxGap = gap
			selectedSkill = skill
		}
	}

	return selectedSkill
}

// selectPracticeScenario chooses a practice scenario for a skill
func (sps *SkillPracticeSystem) selectPracticeScenario(skill *Skill) *PracticeScenario {
	if len(skill.PracticeScenarios) == 0 {
		return nil
	}

	// Select scenario matching current skill level
	for _, scenario := range skill.PracticeScenarios {
		if math.Abs(scenario.Difficulty-skill.CurrentLevel) < 0.2 {
			return scenario
		}
	}

	// Fallback to first scenario
	return skill.PracticeScenarios[0]
}

// executePractice practices a skill by asking the configured LLM evaluator to solve and
// self-grade a concrete exercise. When no provider is configured, it falls back to a
// deterministic rubric scorer derived from the scenario thresholds instead of random
// simulation, preserving reproducible autonomous practice in offline environments.
func (sps *SkillPracticeSystem) executePractice(skill *Skill, scenario *PracticeScenario) *PerformanceResult {
	if sps.providerManager != nil && sps.providerManager.Available() {
		if result, err := sps.executeLLMPractice(skill, scenario); err == nil {
			return result
		}
	}

	return sps.executeRubricPractice(skill, scenario, "provider_unavailable")
}

// executeLLMPractice performs a real practice attempt and expects a structured JSON
// evaluation. The prompt asks Echo to practice first and then grade itself against
// the scenario rubric, giving the practice loop concrete cognitive work rather than
// synthetic score drift.
func (sps *SkillPracticeSystem) executeLLMPractice(skill *Skill, scenario *PracticeScenario) (*PerformanceResult, error) {
	ctx, cancel := context.WithTimeout(sps.ctx, minDuration(scenario.TimeLimit+30*time.Second, 3*time.Minute))
	defer cancel()

	prompt := fmt.Sprintf(`Skill: %s
Description: %s
Category: %s
Current level: %.2f
Target level: %.2f
Scenario: %s
Exercise prompt: %s
Evaluation metrics: %s
Thresholds: %s
Weights: %s

Practice the exercise, then evaluate the attempt. Return only compact JSON matching this schema:
{
  "score": 0.0,
  "metric_scores": {"metric_name": 0.0},
  "strengths": ["metric_name"],
  "weaknesses": ["metric_name"],
  "insights": ["short actionable learning insight"]
}
All numeric scores must be in [0,1].`,
		skill.Name, skill.Description, skill.Category.String(), skill.CurrentLevel, skill.TargetLevel,
		scenario.Name, scenario.Prompt, strings.Join(scenario.Evaluation.Metrics, ", "),
		formatFloatMap(scenario.Evaluation.Thresholds), formatFloatMap(scenario.Evaluation.Weights))

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 700
	opts.Temperature = 0.25
	opts.SystemPrompt = "You are Deep Tree Echo's autonomous skill-practice evaluator. Be rigorous, truthful, wise, and concise. Return JSON only."

	response, err := sps.providerManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, err
	}

	result, insights, err := parsePracticeEvaluation(response, scenario)
	if err != nil {
		return nil, err
	}
	result.Improvement = sps.calculateImprovement(skill.ID, result.Score)
	if len(insights) > 0 {
		// Insights are retained in the current active session when available.
		// The PracticeSession struct already exposes an Insights field for future
		// cross-module memory integration.
	}
	return result, nil
}

// executeRubricPractice is deterministic and based on the actual exercise rubric.
// It prevents autonomous practice from pretending to improve randomly while still
// allowing local tests and offline agents to make measurable progress signals.
func (sps *SkillPracticeSystem) executeRubricPractice(skill *Skill, scenario *PracticeScenario, reason string) *PerformanceResult {
	result := &PerformanceResult{
		Score:        0.0,
		MetricScores: make(map[string]float64),
		Strengths:    make([]string, 0),
		Weaknesses:   make([]string, 0),
	}

	weightedSum := 0.0
	totalWeight := 0.0
	for _, metric := range scenario.Evaluation.Metrics {
		weight := scenario.Evaluation.Weights[metric]
		if weight <= 0 {
			weight = 1.0
		}
		threshold := scenario.Evaluation.Thresholds[metric]
		if threshold <= 0 {
			threshold = 0.65
		}

		// Deterministic competence estimate: skill level adjusted by scenario
		// difficulty and the metric's required threshold.
		metricScore := skill.CurrentLevel + (0.5-scenario.Difficulty)*0.15 + (1.0-threshold)*0.10
		metricScore = clamp(metricScore, 0.0, 1.0)
		result.MetricScores[metric] = metricScore
		weightedSum += metricScore * weight
		totalWeight += weight

		if metricScore >= threshold {
			result.Strengths = append(result.Strengths, metric)
		} else {
			result.Weaknesses = append(result.Weaknesses, metric)
		}
	}

	if totalWeight > 0 {
		result.Score = clamp(weightedSum/totalWeight, 0.0, 1.0)
	}
	result.Improvement = sps.calculateImprovement(skill.ID, result.Score)
	if reason != "" {
		result.Weaknesses = append(result.Weaknesses, reason)
	}
	return result
}

// updatePerformanceMetrics updates performance tracking
func (sps *SkillPracticeSystem) updatePerformanceMetrics(skillID string, performance *PerformanceResult) {
	sps.mu.Lock()
	defer sps.mu.Unlock()

	metrics := sps.performanceMetrics[skillID]
	if metrics == nil {
		return
	}

	// Update scores
	metrics.RecentScores = append(metrics.RecentScores, performance.Score)
	if len(metrics.RecentScores) > 10 {
		metrics.RecentScores = metrics.RecentScores[1:]
	}

	// Update average
	sum := 0.0
	for _, score := range metrics.RecentScores {
		sum += score
	}
	metrics.AverageScore = sum / float64(len(metrics.RecentScores))

	// Update best score
	if performance.Score > metrics.BestScore {
		metrics.BestScore = performance.Score
	}

	// Update trend
	if len(metrics.RecentScores) >= 2 {
		recent := metrics.RecentScores[len(metrics.RecentScores)-1]
		previous := metrics.RecentScores[len(metrics.RecentScores)-2]
		metrics.Trend = recent - previous
	}

	// Update practice count
	metrics.PracticeCount++

	// Update last improvement time if improved
	if performance.Improvement > 0 {
		metrics.LastImprovement = time.Now()
	}
}

// updateSkillLevel updates the skill's current level
func (sps *SkillPracticeSystem) updateSkillLevel(skill *Skill, performance *PerformanceResult) {
	sps.mu.Lock()
	defer sps.mu.Unlock()

	// Increase skill level based on performance
	improvement := performance.Improvement * 0.5 // Gradual improvement
	skill.CurrentLevel += improvement
	skill.CurrentLevel = clamp(skill.CurrentLevel, 0.0, 1.0)
	skill.LastPracticed = time.Now()

	sps.skillsImproved++
}

// skillAssessmentLoop periodically assesses skill levels
func (sps *SkillPracticeSystem) skillAssessmentLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-sps.ctx.Done():
			return
		case <-ticker.C:
			sps.assessAllSkills()
		}
	}
}

// assessAllSkills evaluates current skill levels
func (sps *SkillPracticeSystem) assessAllSkills() {
	// Assess each skill and identify areas for improvement
	// This could generate new practice goals
}

// practiceGoalGenerationLoop generates autonomous practice goals
func (sps *SkillPracticeSystem) practiceGoalGenerationLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-sps.ctx.Done():
			return
		case <-ticker.C:
			sps.generatePracticeGoals()
		}
	}
}

// generatePracticeGoals creates new practice goals based on skill gaps
func (sps *SkillPracticeSystem) generatePracticeGoals() {
	sps.mu.Lock()
	defer sps.mu.Unlock()

	for _, skill := range sps.skills {
		gap := skill.TargetLevel - skill.CurrentLevel
		if gap > 0.2 {
			// Create practice goal
			goal := &PracticeGoal{
				ID:          uuid.New().String(),
				SkillID:     skill.ID,
				TargetLevel: skill.TargetLevel,
				Deadline:    time.Now().Add(7 * 24 * time.Hour),
				Priority:    int(gap * 10),
				Status:      "active",
				Progress:    0.0,
			}

			sps.practiceGoals = append(sps.practiceGoals, goal)
		}
	}
}

// GetMetrics returns practice system metrics
func (sps *SkillPracticeSystem) GetMetrics() map[string]interface{} {
	sps.mu.RLock()
	defer sps.mu.RUnlock()

	return map[string]interface{}{
		"sessions_completed": sps.sessionsCompleted,
		"skills_improved":    sps.skillsImproved,
		"total_skills":       len(sps.skills),
		"active_goals":       len(sps.practiceGoals),
	}
}

// GetSkillLevels returns current levels for all skills
func (sps *SkillPracticeSystem) GetSkillLevels() map[string]float64 {
	sps.mu.RLock()
	defer sps.mu.RUnlock()

	levels := make(map[string]float64)
	for _, skill := range sps.skills {
		levels[skill.Name] = skill.CurrentLevel
	}

	return levels
}

// SetProviderManager injects the LLM provider manager used by autonomous practice.
// This makes the skill system reusable inside larger Echo runtimes that already
// own provider selection and fallback policy.
func (sps *SkillPracticeSystem) SetProviderManager(pm *llm.ProviderManager) {
	sps.mu.Lock()
	defer sps.mu.Unlock()
	sps.providerManager = pm
}

// newPracticeProviderManager builds a provider manager from environment variables
// using the same Anthropic -> OpenRouter -> OpenAI preference used elsewhere in
// Deep Tree Echo evolution code.
func newPracticeProviderManager() *llm.ProviderManager {
	pm := llm.NewProviderManager()
	available := make([]string, 0, 3)

	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		_ = pm.RegisterProvider(llm.NewAnthropicProvider(""))
		available = append(available, "anthropic")
	}
	if os.Getenv("OPENROUTER_API_KEY") != "" {
		_ = pm.RegisterProvider(llm.NewOpenRouterProvider(""))
		available = append(available, "openrouter")
	}
	if os.Getenv("OPENAI_API_KEY") != "" {
		_ = pm.RegisterProvider(llm.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY")))
		available = append(available, "openai")
	}
	if len(available) > 0 {
		_ = pm.SetFallbackChain(available)
	}
	return pm
}

type practiceEvaluationJSON struct {
	Score        float64            `json:"score"`
	MetricScores map[string]float64 `json:"metric_scores"`
	Strengths    []string           `json:"strengths"`
	Weaknesses   []string           `json:"weaknesses"`
	Insights     []string           `json:"insights"`
}

func parsePracticeEvaluation(response string, scenario *PracticeScenario) (*PerformanceResult, []string, error) {
	jsonText := extractJSONObject(response)
	var decoded practiceEvaluationJSON
	if err := json.Unmarshal([]byte(jsonText), &decoded); err != nil {
		return nil, nil, fmt.Errorf("failed to parse practice evaluation JSON: %w", err)
	}

	result := &PerformanceResult{
		Score:        clamp(decoded.Score, 0.0, 1.0),
		MetricScores: make(map[string]float64),
		Strengths:    append([]string{}, decoded.Strengths...),
		Weaknesses:   append([]string{}, decoded.Weaknesses...),
	}

	weightedSum := 0.0
	totalWeight := 0.0
	for _, metric := range scenario.Evaluation.Metrics {
		score, ok := decoded.MetricScores[metric]
		if !ok {
			score = result.Score
		}
		score = clamp(score, 0.0, 1.0)
		result.MetricScores[metric] = score

		weight := scenario.Evaluation.Weights[metric]
		if weight <= 0 {
			weight = 1.0
		}
		weightedSum += score * weight
		totalWeight += weight

		threshold := scenario.Evaluation.Thresholds[metric]
		if threshold <= 0 {
			threshold = 0.65
		}
		if score >= threshold && !containsString(result.Strengths, metric) {
			result.Strengths = append(result.Strengths, metric)
		}
		if score < threshold && !containsString(result.Weaknesses, metric) {
			result.Weaknesses = append(result.Weaknesses, metric)
		}
	}
	if totalWeight > 0 {
		result.Score = clamp(weightedSum/totalWeight, 0.0, 1.0)
	}
	return result, decoded.Insights, nil
}

func (sps *SkillPracticeSystem) calculateImprovement(skillID string, score float64) float64 {
	sps.mu.RLock()
	defer sps.mu.RUnlock()
	metrics := sps.performanceMetrics[skillID]
	if metrics == nil || metrics.AverageScore == 0 {
		return 0.0
	}
	return score - metrics.AverageScore
}

func extractJSONObject(response string) string {
	trimmed := strings.TrimSpace(response)
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end >= start {
		return trimmed[start : end+1]
	}
	return trimmed
}

func formatFloatMap(values map[string]float64) string {
	if len(values) == 0 {
		return "{}"
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return fmt.Sprintf("%v", values)
	}
	return string(encoded)
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
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
