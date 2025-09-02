package orchestration

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// LearningSystem provides advanced learning capabilities for agents
type LearningSystem struct {
	performanceHistory map[string][]*TaskPerformance
	learningModels     map[string]*AgentLearningModel
	adaptationEngine   *AdaptationEngine
}

// TaskPerformance represents historical performance data for a task
type TaskPerformance struct {
	TaskID      string
	TaskType    string
	AgentID     string
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Success     bool
	Quality     float64 // 0.0 to 1.0
	Difficulty  float64 // 0.0 to 1.0
	Context     map[string]interface{}
	Feedback    *PerformanceFeedback
}

// PerformanceFeedback provides detailed feedback about task execution
type PerformanceFeedback struct {
	Accuracy        float64 // How accurate was the output
	Efficiency      float64 // How efficiently was the task completed
	Creativity      float64 // How creative/innovative was the approach
	Adaptability    float64 // How well did the agent adapt to challenges
	Collaboration   float64 // How well did the agent collaborate with others
	LearningRate    float64 // How quickly is the agent improving
}

// AgentLearningModel represents an agent's learning patterns and capabilities
type AgentLearningModel struct {
	AgentID               string
	LearningRate          float64
	AdaptationCapability  float64
	SpecializationAreas   []string
	WeakAreas            []string
	PreferredTaskTypes   []string
	CollaborationStyle   CollaborationStyle
	LearningTrajectory   *LearningTrajectory
	LastUpdated          time.Time
}

// CollaborationStyle defines how an agent prefers to collaborate
type CollaborationStyle string

const (
	CollaborationStyleIndependent CollaborationStyle = "independent"
	CollaborationStyleCooperative CollaborationStyle = "cooperative"
	CollaborationStyleDirective   CollaborationStyle = "directive"
	CollaborationStyleSupportive  CollaborationStyle = "supportive"
)

// LearningTrajectory tracks an agent's learning progress over time
type LearningTrajectory struct {
	InitialPerformance  float64
	CurrentPerformance  float64
	PeakPerformance     float64
	ImprovementRate     float64
	LearningMilestones  []*LearningMilestone
	PredictedTrajectory []float64
}

// LearningMilestone represents a significant learning achievement
type LearningMilestone struct {
	Timestamp   time.Time
	Achievement string
	Performance float64
	Context     map[string]interface{}
}

// AdaptationEngine handles dynamic adaptation of agent capabilities
type AdaptationEngine struct {
	adaptationStrategies map[string]AdaptationStrategy
	environmentFactors   *EnvironmentFactors
}

// AdaptationStrategy defines how agents should adapt to different situations
type AdaptationStrategy interface {
	Name() string
	ShouldAdapt(agent *Agent, context *AdaptationContext) bool
	Adapt(agent *Agent, context *AdaptationContext) (*AdaptationResult, error)
}

// AdaptationContext provides context for adaptation decisions
type AdaptationContext struct {
	RecentPerformance    []*TaskPerformance
	EnvironmentChanges   map[string]interface{}
	CollaborationHistory []*CollaborationRecord
	TaskDemands          map[string]float64
}

// AdaptationResult represents the result of an adaptation process
type AdaptationResult struct {
	Changes         map[string]interface{}
	ExpectedImprovement float64
	RiskLevel       string
	RecommendedActions []string
}

// CollaborationRecord tracks collaboration interactions between agents
type CollaborationRecord struct {
	Timestamp     time.Time
	Participants  []string
	TaskID        string
	Outcome       CollaborationOutcome
	Satisfaction  float64
	Lessons       []string
}

// CollaborationOutcome defines the result of a collaboration
type CollaborationOutcome string

const (
	CollaborationOutcomeSuccess     CollaborationOutcome = "success"
	CollaborationOutcomePartial     CollaborationOutcome = "partial"
	CollaborationOutcomeConflict    CollaborationOutcome = "conflict"
	CollaborationOutcomeFailed      CollaborationOutcome = "failed"
)

// EnvironmentFactors represents environmental factors affecting learning
type EnvironmentFactors struct {
	SystemLoad       float64
	TaskComplexity   float64
	TimeConstraints  float64
	ResourceAvailability float64
	CollaborationDensity float64
}

// NewLearningSystem creates a new learning system
func NewLearningSystem() *LearningSystem {
	return &LearningSystem{
		performanceHistory: make(map[string][]*TaskPerformance),
		learningModels:     make(map[string]*AgentLearningModel),
		adaptationEngine:   NewAdaptationEngine(),
	}
}

// NewAdaptationEngine creates a new adaptation engine
func NewAdaptationEngine() *AdaptationEngine {
	engine := &AdaptationEngine{
		adaptationStrategies: make(map[string]AdaptationStrategy),
		environmentFactors: &EnvironmentFactors{
			SystemLoad:           0.5,
			TaskComplexity:       0.5,
			TimeConstraints:      0.5,
			ResourceAvailability: 0.8,
			CollaborationDensity: 0.3,
		},
	}

	// Register default adaptation strategies
	engine.RegisterStrategy(&PerformanceBasedAdaptation{})
	engine.RegisterStrategy(&CollaborationAdaptation{})
	engine.RegisterStrategy(&SpecializationAdaptation{})
	
	return engine
}

// RecordTaskPerformance records performance data for learning
func (ls *LearningSystem) RecordTaskPerformance(performance *TaskPerformance) {
	agentID := performance.AgentID
	if ls.performanceHistory[agentID] == nil {
		ls.performanceHistory[agentID] = make([]*TaskPerformance, 0)
	}
	
	ls.performanceHistory[agentID] = append(ls.performanceHistory[agentID], performance)
	
	// Maintain a rolling window of recent performance (keep last 100 records)
	if len(ls.performanceHistory[agentID]) > 100 {
		ls.performanceHistory[agentID] = ls.performanceHistory[agentID][1:]
	}
	
	// Update learning model
	ls.updateLearningModel(agentID)
}

// GetLearningModel returns the learning model for an agent
func (ls *LearningSystem) GetLearningModel(agentID string) *AgentLearningModel {
	if model, exists := ls.learningModels[agentID]; exists {
		return model
	}
	
	// Create new learning model if it doesn't exist
	model := &AgentLearningModel{
		AgentID:               agentID,
		LearningRate:          0.1,
		AdaptationCapability:  0.5,
		SpecializationAreas:   []string{},
		WeakAreas:            []string{},
		PreferredTaskTypes:   []string{},
		CollaborationStyle:   CollaborationStyleCooperative,
		LearningTrajectory:   &LearningTrajectory{
			LearningMilestones:  []*LearningMilestone{},
			PredictedTrajectory: []float64{},
		},
		LastUpdated: time.Now(),
	}
	
	ls.learningModels[agentID] = model
	return model
}

// updateLearningModel updates an agent's learning model based on recent performance
func (ls *LearningSystem) updateLearningModel(agentID string) {
	model := ls.GetLearningModel(agentID)
	history := ls.performanceHistory[agentID]
	
	if len(history) < 2 {
		return
	}
	
	// Calculate current performance metrics
	recentPerformance := ls.calculateRecentPerformance(history)
	
	// Update learning trajectory
	if model.LearningTrajectory.InitialPerformance == 0 {
		model.LearningTrajectory.InitialPerformance = recentPerformance
	}
	
	previousPerformance := model.LearningTrajectory.CurrentPerformance
	model.LearningTrajectory.CurrentPerformance = recentPerformance
	
	if recentPerformance > model.LearningTrajectory.PeakPerformance {
		model.LearningTrajectory.PeakPerformance = recentPerformance
		
		// Record milestone for peak performance
		milestone := &LearningMilestone{
			Timestamp:   time.Now(),
			Achievement: "Peak Performance Achieved",
			Performance: recentPerformance,
			Context: map[string]interface{}{
				"improvement": recentPerformance - previousPerformance,
			},
		}
		model.LearningTrajectory.LearningMilestones = append(
			model.LearningTrajectory.LearningMilestones, milestone)
	}
	
	// Calculate improvement rate
	if previousPerformance > 0 {
		model.LearningTrajectory.ImprovementRate = (recentPerformance - previousPerformance) / previousPerformance
	}
	
	// Update specialization areas
	ls.updateSpecializationAreas(model, history)
	
	// Update collaboration style based on recent interactions
	ls.updateCollaborationStyle(model, history)
	
	// Update learning rate based on consistency of improvement
	ls.updateLearningRate(model, history)
	
	model.LastUpdated = time.Now()
}

// calculateRecentPerformance calculates performance based on recent task history
func (ls *LearningSystem) calculateRecentPerformance(history []*TaskPerformance) float64 {
	if len(history) == 0 {
		return 0.0
	}
	
	// Consider last 10 tasks for recent performance
	recentCount := int(math.Min(10, float64(len(history))))
	recentHistory := history[len(history)-recentCount:]
	
	totalQuality := 0.0
	successCount := 0
	
	for _, perf := range recentHistory {
		totalQuality += perf.Quality
		if perf.Success {
			successCount++
		}
	}
	
	avgQuality := totalQuality / float64(len(recentHistory))
	successRate := float64(successCount) / float64(len(recentHistory))
	
	// Combine quality and success rate with weights
	return (avgQuality * 0.7) + (successRate * 0.3)
}

// updateSpecializationAreas identifies areas where the agent excels
func (ls *LearningSystem) updateSpecializationAreas(model *AgentLearningModel, history []*TaskPerformance) {
	taskTypePerformance := make(map[string][]float64)
	
	// Group performance by task type
	for _, perf := range history {
		if perf.Success && perf.Quality > 0.7 { // Only consider high-quality successful tasks
			taskTypePerformance[perf.TaskType] = append(taskTypePerformance[perf.TaskType], perf.Quality)
		}
	}
	
	// Calculate average performance per task type
	avgPerformance := make([]struct {
		TaskType string
		AvgQuality float64
		Count int
	}, 0)
	
	for taskType, qualities := range taskTypePerformance {
		if len(qualities) >= 3 { // Need at least 3 instances
			sum := 0.0
			for _, quality := range qualities {
				sum += quality
			}
			avg := sum / float64(len(qualities))
			avgPerformance = append(avgPerformance, struct {
				TaskType string
				AvgQuality float64
				Count int
			}{taskType, avg, len(qualities)})
		}
	}
	
	// Sort by average quality
	sort.Slice(avgPerformance, func(i, j int) bool {
		return avgPerformance[i].AvgQuality > avgPerformance[j].AvgQuality
	})
	
	// Update specialization areas (top performers)
	model.SpecializationAreas = make([]string, 0)
	model.PreferredTaskTypes = make([]string, 0)
	for i, perf := range avgPerformance {
		if i < 3 && perf.AvgQuality > 0.8 { // Top 3 with quality > 80%
			model.SpecializationAreas = append(model.SpecializationAreas, perf.TaskType)
		}
		if perf.AvgQuality > 0.7 {
			model.PreferredTaskTypes = append(model.PreferredTaskTypes, perf.TaskType)
		}
	}
}

// updateCollaborationStyle updates collaboration preferences based on outcomes
func (ls *LearningSystem) updateCollaborationStyle(model *AgentLearningModel, history []*TaskPerformance) {
	// This is a simplified implementation - in reality, we'd track collaboration records
	collaborativeTasks := 0
	independentTasks := 0
	collaborativeSuccess := 0.0
	independentSuccess := 0.0
	
	for _, perf := range history {
		if perf.Context != nil {
			if isCollaborative, ok := perf.Context["collaborative"].(bool); ok && isCollaborative {
				collaborativeTasks++
				if perf.Success {
					collaborativeSuccess += perf.Quality
				}
			} else {
				independentTasks++
				if perf.Success {
					independentSuccess += perf.Quality
				}
			}
		}
	}
	
	if collaborativeTasks > 0 && independentTasks > 0 {
		collaborativeAvg := collaborativeSuccess / float64(collaborativeTasks)
		independentAvg := independentSuccess / float64(independentTasks)
		
		if collaborativeAvg > independentAvg*1.1 { // 10% better in collaboration
			model.CollaborationStyle = CollaborationStyleCooperative
		} else if independentAvg > collaborativeAvg*1.1 {
			model.CollaborationStyle = CollaborationStyleIndependent
		}
	}
}

// updateLearningRate adjusts learning rate based on improvement consistency
func (ls *LearningSystem) updateLearningRate(model *AgentLearningModel, history []*TaskPerformance) {
	if len(history) < 10 {
		return
	}
	
	// Look at improvement trajectory over recent history
	recentHistory := history[len(history)-10:]
	improvements := make([]float64, 0)
	
	for i := 1; i < len(recentHistory); i++ {
		prev := recentHistory[i-1]
		curr := recentHistory[i]
		
		if prev.Quality > 0 {
			improvement := (curr.Quality - prev.Quality) / prev.Quality
			improvements = append(improvements, improvement)
		}
	}
	
	if len(improvements) == 0 {
		return
	}
	
	// Calculate consistency of improvements
	avgImprovement := 0.0
	for _, imp := range improvements {
		avgImprovement += imp
	}
	avgImprovement /= float64(len(improvements))
	
	variance := 0.0
	for _, imp := range improvements {
		variance += (imp - avgImprovement) * (imp - avgImprovement)
	}
	variance /= float64(len(improvements))
	
	consistency := 1.0 / (1.0 + variance) // Higher consistency = lower variance
	
	// Adjust learning rate based on consistency and average improvement
	if avgImprovement > 0.05 && consistency > 0.8 { // Good, consistent improvement
		model.LearningRate = math.Min(0.3, model.LearningRate*1.1)
	} else if avgImprovement < -0.05 { // Performance declining
		model.LearningRate = math.Max(0.05, model.LearningRate*0.9)
	}
}

// PredictOptimalAgent predicts which agent would be best for a specific task
func (ls *LearningSystem) PredictOptimalAgent(ctx context.Context, task *Task, availableAgents []*Agent) (*Agent, float64, error) {
	if len(availableAgents) == 0 {
		return nil, 0.0, fmt.Errorf("no available agents")
	}
	
	bestAgent := availableAgents[0]
	bestScore := 0.0
	
	for _, agent := range availableAgents {
		score := ls.calculateTaskFitScore(agent, task)
		if score > bestScore {
			bestScore = score
			bestAgent = agent
		}
	}
	
	return bestAgent, bestScore, nil
}

// calculateTaskFitScore calculates how well an agent fits a specific task
func (ls *LearningSystem) calculateTaskFitScore(agent *Agent, task *Task) float64 {
	model := ls.GetLearningModel(agent.ID)
	score := 0.5 // Base score
	
	// Boost score for specialization areas
	for _, specialization := range model.SpecializationAreas {
		if specialization == task.Type {
			score += 0.3
			break
		}
	}
	
	// Boost score for preferred task types
	for _, preferred := range model.PreferredTaskTypes {
		if preferred == task.Type {
			score += 0.1
			break
		}
	}
	
	// Factor in current performance and learning trajectory
	if model.LearningTrajectory.CurrentPerformance > 0 {
		score += model.LearningTrajectory.CurrentPerformance * 0.2
	}
	
	// Factor in improvement rate (positive trend)
	if model.LearningTrajectory.ImprovementRate > 0 {
		score += model.LearningTrajectory.ImprovementRate * 0.1
	}
	
	return math.Min(1.0, score)
}

// RegisterStrategy registers an adaptation strategy
func (ae *AdaptationEngine) RegisterStrategy(strategy AdaptationStrategy) {
	ae.adaptationStrategies[strategy.Name()] = strategy
}

// AdaptAgent performs adaptation for an agent based on recent performance
func (ae *AdaptationEngine) AdaptAgent(ctx context.Context, agent *Agent, learningSystem *LearningSystem) (*AdaptationResult, error) {
	history := learningSystem.performanceHistory[agent.ID]
	
	if len(history) < 5 {
		return &AdaptationResult{
			Changes: make(map[string]interface{}),
			ExpectedImprovement: 0.0,
			RiskLevel: "low",
			RecommendedActions: []string{"Continue monitoring performance"},
		}, nil
	}
	
	adaptationContext := &AdaptationContext{
		RecentPerformance: history,
		EnvironmentChanges: make(map[string]interface{}),
		TaskDemands: make(map[string]float64),
	}
	
	// Limit recent performance to available history
	if len(history) > 10 {
		adaptationContext.RecentPerformance = history[len(history)-10:]
	}
	
	totalChanges := make(map[string]interface{})
	totalImprovement := 0.0
	recommendedActions := make([]string, 0)
	
	// Apply each adaptation strategy
	for _, strategy := range ae.adaptationStrategies {
		if strategy.ShouldAdapt(agent, adaptationContext) {
			result, err := strategy.Adapt(agent, adaptationContext)
			if err != nil {
				continue
			}
			
			for k, v := range result.Changes {
				totalChanges[k] = v
			}
			totalImprovement += result.ExpectedImprovement
			recommendedActions = append(recommendedActions, result.RecommendedActions...)
		}
	}
	
	return &AdaptationResult{
		Changes:             totalChanges,
		ExpectedImprovement: totalImprovement,
		RiskLevel:          "low",
		RecommendedActions: recommendedActions,
	}, nil
}

// Performance-based adaptation strategy
type PerformanceBasedAdaptation struct{}

func (pba *PerformanceBasedAdaptation) Name() string {
	return "performance_based"
}

func (pba *PerformanceBasedAdaptation) ShouldAdapt(agent *Agent, context *AdaptationContext) bool {
	if len(context.RecentPerformance) < 5 {
		return false
	}
	
	// Check if performance is declining
	recentQuality := 0.0
	for _, perf := range context.RecentPerformance {
		recentQuality += perf.Quality
	}
	avgRecentQuality := recentQuality / float64(len(context.RecentPerformance))
	
	return avgRecentQuality < 0.6 // Adapt if performance is below 60%
}

func (pba *PerformanceBasedAdaptation) Adapt(agent *Agent, context *AdaptationContext) (*AdaptationResult, error) {
	changes := make(map[string]interface{})
	actions := []string{}
	
	// Analyze failure patterns
	failureTypes := make(map[string]int)
	for _, perf := range context.RecentPerformance {
		if !perf.Success {
			if reason, exists := perf.Context["failure_reason"].(string); exists {
				failureTypes[reason]++
			}
		}
	}
	
	// Recommend specific improvements based on failure patterns
	if failureTypes["timeout"] > 2 {
		changes["recommended_timeout_increase"] = true
		actions = append(actions, "Consider increasing task timeout limits")
	}
	
	if failureTypes["resource_constraint"] > 1 {
		changes["recommended_resource_boost"] = true
		actions = append(actions, "Consider allocating additional resources")
	}
	
	return &AdaptationResult{
		Changes:             changes,
		ExpectedImprovement: 0.15,
		RiskLevel:          "low",
		RecommendedActions: actions,
	}, nil
}

// Collaboration adaptation strategy
type CollaborationAdaptation struct{}

func (ca *CollaborationAdaptation) Name() string {
	return "collaboration_based"
}

func (ca *CollaborationAdaptation) ShouldAdapt(agent *Agent, context *AdaptationContext) bool {
	// Check if agent has collaboration experience
	collaborativeTasks := 0
	for _, perf := range context.RecentPerformance {
		if isCollab, exists := perf.Context["collaborative"].(bool); exists && isCollab {
			collaborativeTasks++
		}
	}
	return collaborativeTasks > 3
}

func (ca *CollaborationAdaptation) Adapt(agent *Agent, context *AdaptationContext) (*AdaptationResult, error) {
	changes := make(map[string]interface{})
	actions := []string{}
	
	// Analyze collaboration effectiveness
	collabSuccess := 0
	totalCollab := 0
	for _, perf := range context.RecentPerformance {
		if isCollab, exists := perf.Context["collaborative"].(bool); exists && isCollab {
			totalCollab++
			if perf.Success && perf.Quality > 0.7 {
				collabSuccess++
			}
		}
	}
	
	if totalCollab > 0 {
		collabRate := float64(collabSuccess) / float64(totalCollab)
		if collabRate > 0.8 {
			changes["collaboration_preference"] = "high"
			actions = append(actions, "Agent shows strong collaboration skills - prioritize collaborative tasks")
		} else if collabRate < 0.5 {
			changes["collaboration_preference"] = "low"
			actions = append(actions, "Agent may benefit from independent work - reduce collaborative assignments")
		}
	}
	
	return &AdaptationResult{
		Changes:             changes,
		ExpectedImprovement: 0.1,
		RiskLevel:          "low",
		RecommendedActions: actions,
	}, nil
}

// Specialization adaptation strategy
type SpecializationAdaptation struct{}

func (sa *SpecializationAdaptation) Name() string {
	return "specialization_based"
}

func (sa *SpecializationAdaptation) ShouldAdapt(agent *Agent, context *AdaptationContext) bool {
	return len(context.RecentPerformance) >= 10
}

func (sa *SpecializationAdaptation) Adapt(agent *Agent, context *AdaptationContext) (*AdaptationResult, error) {
	changes := make(map[string]interface{})
	actions := []string{}
	
	// Analyze task type performance
	taskPerformance := make(map[string][]float64)
	for _, perf := range context.RecentPerformance {
		taskPerformance[perf.TaskType] = append(taskPerformance[perf.TaskType], perf.Quality)
	}
	
	// Find strongest and weakest task types
	bestTaskType := ""
	bestAvg := 0.0
	worstTaskType := ""
	worstAvg := 1.0
	
	for taskType, qualities := range taskPerformance {
		if len(qualities) >= 3 { // Need minimum sample size
			sum := 0.0
			for _, q := range qualities {
				sum += q
			}
			avg := sum / float64(len(qualities))
			
			if avg > bestAvg {
				bestAvg = avg
				bestTaskType = taskType
			}
			if avg < worstAvg {
				worstAvg = avg
				worstTaskType = taskType
			}
		}
	}
	
	if bestTaskType != "" && bestAvg > 0.8 {
		changes["recommended_specialization"] = bestTaskType
		actions = append(actions, fmt.Sprintf("Agent excels at %s tasks - consider specialization", bestTaskType))
	}
	
	if worstTaskType != "" && worstAvg < 0.5 {
		changes["avoid_task_type"] = worstTaskType
		actions = append(actions, fmt.Sprintf("Agent struggles with %s tasks - consider additional training or avoidance", worstTaskType))
	}
	
	return &AdaptationResult{
		Changes:             changes,
		ExpectedImprovement: 0.2,
		RiskLevel:          "medium",
		RecommendedActions: actions,
	}, nil
}