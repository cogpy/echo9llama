package orchestration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ollama/ollama/api"
)

func TestNewLearningSystem(t *testing.T) {
	ls := NewLearningSystem()
	
	if ls == nil {
		t.Fatal("NewLearningSystem should return a non-nil system")
	}
	
	if ls.performanceHistory == nil {
		t.Error("Performance history should be initialized")
	}
	
	if ls.learningModels == nil {
		t.Error("Learning models should be initialized")
	}
	
	if ls.adaptationEngine == nil {
		t.Error("Adaptation engine should be initialized")
	}
}

func TestRecordTaskPerformance(t *testing.T) {
	ls := NewLearningSystem()
	agentID := "test-agent-1"
	
	performance := &TaskPerformance{
		TaskID:     "task-1",
		TaskType:   TaskTypeGenerate,
		AgentID:    agentID,
		StartTime:  time.Now().Add(-5 * time.Second),
		EndTime:    time.Now(),
		Duration:   5 * time.Second,
		Success:    true,
		Quality:    0.8,
		Difficulty: 0.5,
		Context:    map[string]interface{}{"test": true},
	}
	
	ls.RecordTaskPerformance(performance)
	
	// Check that performance was recorded
	if history, exists := ls.performanceHistory[agentID]; !exists || len(history) == 0 {
		t.Error("Performance should be recorded in history")
	}
	
	// Check that learning model was created/updated
	model := ls.GetLearningModel(agentID)
	if model == nil {
		t.Error("Learning model should be created")
	}
	
	if model.AgentID != agentID {
		t.Errorf("Expected agent ID %s, got %s", agentID, model.AgentID)
	}
}

func TestGetLearningModel(t *testing.T) {
	ls := NewLearningSystem()
	agentID := "test-agent-2"
	
	// Test creating new model
	model := ls.GetLearningModel(agentID)
	if model == nil {
		t.Fatal("GetLearningModel should create a new model if none exists")
	}
	
	if model.AgentID != agentID {
		t.Errorf("Expected agent ID %s, got %s", agentID, model.AgentID)
	}
	
	if model.LearningTrajectory == nil {
		t.Error("Learning trajectory should be initialized")
	}
	
	// Test retrieving existing model
	model2 := ls.GetLearningModel(agentID)
	if model != model2 {
		t.Error("Should return the same model instance for the same agent ID")
	}
}

func TestCalculateRecentPerformance(t *testing.T) {
	ls := NewLearningSystem()
	
	// Create test performance history
	history := []*TaskPerformance{
		{Quality: 0.6, Success: true},
		{Quality: 0.7, Success: true},
		{Quality: 0.8, Success: true},
		{Quality: 0.5, Success: false},
		{Quality: 0.9, Success: true},
	}
	
	performance := ls.calculateRecentPerformance(history)
	
	// Expected: (0.6 + 0.7 + 0.8 + 0.5 + 0.9) / 5 * 0.7 + (4/5) * 0.3 = 0.7 * 0.7 + 0.8 * 0.3 = 0.49 + 0.24 = 0.73
	expectedMin := 0.7
	expectedMax := 0.75
	
	if performance < expectedMin || performance > expectedMax {
		t.Errorf("Expected performance between %.2f and %.2f, got %.2f", expectedMin, expectedMax, performance)
	}
}

func TestUpdateSpecializationAreas(t *testing.T) {
	ls := NewLearningSystem()
	model := ls.GetLearningModel("test-agent")
	
	// Create performance history with clear specializations
	history := []*TaskPerformance{
		{TaskType: TaskTypeGenerate, Quality: 0.9, Success: true},
		{TaskType: TaskTypeGenerate, Quality: 0.85, Success: true},
		{TaskType: TaskTypeGenerate, Quality: 0.88, Success: true},
		{TaskType: TaskTypeChat, Quality: 0.6, Success: true},
		{TaskType: TaskTypeChat, Quality: 0.65, Success: true},
		{TaskType: TaskTypeChat, Quality: 0.62, Success: true},
		{TaskType: TaskTypeEmbed, Quality: 0.95, Success: true},
		{TaskType: TaskTypeEmbed, Quality: 0.92, Success: true},
		{TaskType: TaskTypeEmbed, Quality: 0.93, Success: true},
	}
	
	ls.updateSpecializationAreas(model, history)
	
	// Should specialize in embed and generate (both > 0.8 avg)
	if len(model.SpecializationAreas) == 0 {
		t.Error("Should have identified specialization areas")
	}
	
	foundEmbed := false
	foundGenerate := false
	for _, area := range model.SpecializationAreas {
		if area == TaskTypeEmbed {
			foundEmbed = true
		}
		if area == TaskTypeGenerate {
			foundGenerate = true
		}
	}
	
	if !foundEmbed {
		t.Error("Should have identified embed as a specialization area")
	}
	if !foundGenerate {
		t.Error("Should have identified generate as a specialization area")
	}
}

func TestPredictOptimalAgent(t *testing.T) {
	ls := NewLearningSystem()
	ctx := context.Background()
	
	// Create agents with different specializations
	agent1 := &Agent{ID: "agent-1", Name: "Agent 1"}
	agent2 := &Agent{ID: "agent-2", Name: "Agent 2"}
	
	// Give agent1 better performance in generate tasks
	generateHistory := []*TaskPerformance{
		{TaskType: TaskTypeGenerate, Quality: 0.9, Success: true, AgentID: agent1.ID},
		{TaskType: TaskTypeGenerate, Quality: 0.85, Success: true, AgentID: agent1.ID},
		{TaskType: TaskTypeGenerate, Quality: 0.88, Success: true, AgentID: agent1.ID},
	}
	
	for _, perf := range generateHistory {
		ls.RecordTaskPerformance(perf)
	}
	
	// Give agent2 better performance in chat tasks
	chatHistory := []*TaskPerformance{
		{TaskType: TaskTypeChat, Quality: 0.92, Success: true, AgentID: agent2.ID},
		{TaskType: TaskTypeChat, Quality: 0.89, Success: true, AgentID: agent2.ID},
		{TaskType: TaskTypeChat, Quality: 0.91, Success: true, AgentID: agent2.ID},
	}
	
	for _, perf := range chatHistory {
		ls.RecordTaskPerformance(perf)
	}
	
	// Test prediction for generate task
	generateTask := &Task{ID: "test-task", Type: TaskTypeGenerate}
	bestAgent, score, err := ls.PredictOptimalAgent(ctx, generateTask, []*Agent{agent1, agent2})
	
	if err != nil {
		t.Fatalf("PredictOptimalAgent failed: %v", err)
	}
	
	if bestAgent.ID != agent1.ID {
		t.Errorf("Expected agent1 to be optimal for generate task, got %s", bestAgent.ID)
	}
	
	if score <= 0.5 {
		t.Errorf("Expected score > 0.5, got %.2f", score)
	}
	
	// Test prediction for chat task
	chatTask := &Task{ID: "test-task-2", Type: TaskTypeChat}
	bestAgent, score, err = ls.PredictOptimalAgent(ctx, chatTask, []*Agent{agent1, agent2})
	
	if err != nil {
		t.Fatalf("PredictOptimalAgent failed: %v", err)
	}
	
	if bestAgent.ID != agent2.ID {
		t.Errorf("Expected agent2 to be optimal for chat task, got %s", bestAgent.ID)
	}
}

func TestAdaptationEngine(t *testing.T) {
	ls := NewLearningSystem()
	ctx := context.Background()
	agent := &Agent{ID: "test-agent", Name: "Test Agent"}
	
	// Create performance history that should trigger adaptation
	poorPerformance := []*TaskPerformance{
		{TaskType: TaskTypeGenerate, Quality: 0.3, Success: false, AgentID: agent.ID, Context: map[string]interface{}{"failure_reason": "timeout"}},
		{TaskType: TaskTypeGenerate, Quality: 0.4, Success: false, AgentID: agent.ID, Context: map[string]interface{}{"failure_reason": "timeout"}},
		{TaskType: TaskTypeGenerate, Quality: 0.2, Success: false, AgentID: agent.ID, Context: map[string]interface{}{"failure_reason": "timeout"}},
		{TaskType: TaskTypeGenerate, Quality: 0.5, Success: true, AgentID: agent.ID},
		{TaskType: TaskTypeGenerate, Quality: 0.3, Success: false, AgentID: agent.ID, Context: map[string]interface{}{"failure_reason": "resource_constraint"}},
		{TaskType: TaskTypeGenerate, Quality: 0.4, Success: false, AgentID: agent.ID, Context: map[string]interface{}{"failure_reason": "resource_constraint"}},
	}
	
	for _, perf := range poorPerformance {
		ls.RecordTaskPerformance(perf)
	}
	
	result, err := ls.adaptationEngine.AdaptAgent(ctx, agent, ls)
	if err != nil {
		t.Fatalf("AdaptAgent failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("AdaptAgent should return a result")
	}
	
	if len(result.RecommendedActions) == 0 {
		t.Error("Should have recommended actions based on poor performance")
	}
	
	// Check for specific recommendations based on failure patterns
	foundTimeoutRecommendation := false
	foundResourceRecommendation := false
	
	for _, action := range result.RecommendedActions {
		if action == "Consider increasing task timeout limits" {
			foundTimeoutRecommendation = true
		}
		if action == "Consider allocating additional resources" {
			foundResourceRecommendation = true
		}
	}
	
	if !foundTimeoutRecommendation {
		t.Errorf("Should recommend timeout increase for timeout failures, got actions: %v", result.RecommendedActions)
	}
	if !foundResourceRecommendation {
		t.Errorf("Should recommend resource boost for resource constraint failures, got actions: %v", result.RecommendedActions)
	}
}

func TestLearningSystemIntegrationWithEngine(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()
	
	// Register default tools and plugins
	RegisterDefaultTools(engine)
	RegisterDefaultPlugins(engine)
	
	// Create an agent
	agent := &Agent{
		Name:        "learning-test-agent",
		Description: "Agent for testing learning integration",
		Type:        AgentTypeGeneral,
		Models:      []string{"llama3.2"},
		Tools:       []string{"calculator"},
	}
	
	err := engine.CreateAgent(ctx, agent)
	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}
	
	// Execute some tasks to generate learning data
	tasks := []*Task{
		{
			ID:      "learning-task-1",
			Type:    TaskTypeTool,
			Input:   "Calculate 5 + 3",
			Status:  TaskStatusPending,
			AgentID: agent.ID,
			Parameters: map[string]interface{}{
				"tool": map[string]interface{}{
					"name": "calculator",
					"parameters": map[string]interface{}{
						"operation": "add",
						"a":         5.0,
						"b":         3.0,
					},
				},
			},
		},
		{
			ID:      "learning-task-2",
			Type:    TaskTypeReflect,
			Input:   "Analyze recent performance",
			Status:  TaskStatusPending,
			AgentID: agent.ID,
		},
	}
	
	for _, task := range tasks {
		result, err := engine.ExecuteTask(ctx, task, agent)
		if err != nil {
			t.Errorf("ExecuteTask failed for task %s: %v", task.ID, err)
		}
		
		if result == nil {
			t.Errorf("ExecuteTask should return a result for task %s", task.ID)
		}
	}
	
	// Check that learning system has recorded the performance
	learningSystem := engine.GetLearningSystem()
	if learningSystem == nil {
		t.Fatal("Engine should have a learning system")
	}
	
	model := learningSystem.GetLearningModel(agent.ID)
	if model == nil {
		t.Fatal("Learning model should exist for the agent")
	}
	
	history := learningSystem.performanceHistory[agent.ID]
	if len(history) != len(tasks) {
		t.Errorf("Expected %d performance records, got %d", len(tasks), len(history))
	}
	
	// Test agent adaptation
	adaptationResult, err := engine.AdaptAgent(ctx, agent.ID)
	if err != nil {
		t.Fatalf("AdaptAgent failed: %v", err)
	}
	
	if adaptationResult == nil {
		t.Error("AdaptAgent should return a result")
	}
	
	// Test optimal agent prediction
	newTask := &Task{
		ID:   "prediction-task",
		Type: TaskTypeTool,
	}
	
	optimalAgent, confidence, err := engine.PredictOptimalAgentForTask(ctx, newTask)
	if err != nil {
		t.Fatalf("PredictOptimalAgentForTask failed: %v", err)
	}
	
	if optimalAgent == nil {
		t.Error("Should predict an optimal agent")
	}
	
	if confidence <= 0.0 || confidence > 1.0 {
		t.Errorf("Confidence should be between 0.0 and 1.0, got %.2f", confidence)
	}
}

func TestLearningTrajectoryEvolution(t *testing.T) {
	ls := NewLearningSystem()
	agentID := "trajectory-test-agent"
	
	// Simulate improving performance over time
	baseTime := time.Now().Add(-24 * time.Hour)
	
	for i := 0; i < 20; i++ {
		quality := 0.3 + float64(i)*0.03 // Gradually improving from 0.3 to 0.87
		
		performance := &TaskPerformance{
			TaskID:     fmt.Sprintf("task-%d", i),
			TaskType:   TaskTypeGenerate,
			AgentID:    agentID,
			StartTime:  baseTime.Add(time.Duration(i) * time.Hour),
			EndTime:    baseTime.Add(time.Duration(i) * time.Hour).Add(5 * time.Second),
			Duration:   5 * time.Second,
			Success:    quality > 0.5,
			Quality:    quality,
			Difficulty: 0.5,
			Context:    map[string]interface{}{"iteration": i},
		}
		
		ls.RecordTaskPerformance(performance)
	}
	
	model := ls.GetLearningModel(agentID)
	trajectory := model.LearningTrajectory
	
	if trajectory.InitialPerformance >= trajectory.CurrentPerformance {
		t.Error("Current performance should be higher than initial performance")
	}
	
	if trajectory.ImprovementRate <= 0 {
		t.Error("Improvement rate should be positive")
	}
	
	if len(trajectory.LearningMilestones) == 0 {
		t.Error("Should have recorded learning milestones")
	}
	
	// Check that peak performance milestone was recorded
	foundPeakMilestone := false
	for _, milestone := range trajectory.LearningMilestones {
		if milestone.Achievement == "Peak Performance Achieved" {
			foundPeakMilestone = true
			break
		}
	}
	
	if !foundPeakMilestone {
		t.Error("Should have recorded peak performance milestone")
	}
}