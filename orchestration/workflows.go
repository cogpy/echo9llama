package orchestration

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CreateDefaultAgent creates a default orchestration agent with common models
func (e *Engine) CreateDefaultAgent(ctx context.Context) (*Agent, error) {
	agent := &Agent{
		Name:        "default",
		Description: "Default orchestration agent for general tasks",
		Type:        AgentTypeGeneral,
		Models:      []string{"llama3.2", "llama2", "codellama"},
		Tools:       []string{"web_search", "calculator"},
		Config: map[string]interface{}{
			"max_concurrent_tasks": 3,
			"default_model":        "llama3.2",
			"timeout_seconds":      300,
		},
	}

	err := e.CreateAgent(ctx, agent)
	if err != nil {
		return nil, err
	}

	return agent, nil
}

// CreateSpecializedAgent creates an agent with specialized capabilities
func (e *Engine) CreateSpecializedAgent(ctx context.Context, agentType AgentType, domain string) (*Agent, error) {
	var agent *Agent
	
	switch agentType {
	case AgentTypeReflective:
		agent = &Agent{
			Name:        fmt.Sprintf("reflective-%s", domain),
			Description: fmt.Sprintf("Self-reflective agent specialized in %s", domain),
			Type:        AgentTypeReflective,
			Models:      []string{"llama3.2", "llama2"},
			Tools:       []string{"data_analysis"},
			Config: map[string]interface{}{
				"reflection_interval": 300,
				"learning_rate":       0.1,
				"domain":             domain,
			},
		}
	case AgentTypeOrchestrator:
		agent = &Agent{
			Name:        "orchestrator-coordinator",
			Description: "Orchestrator agent for coordinating multiple agents and complex workflows",
			Type:        AgentTypeOrchestrator,
			Models:      []string{"llama3.2"},
			Tools:       []string{"web_search", "calculator", "data_analysis"},
			Config: map[string]interface{}{
				"max_sub_agents":    5,
				"coordination_mode": "hierarchical",
			},
		}
	case AgentTypeSpecialist:
		agent = &Agent{
			Name:        fmt.Sprintf("specialist-%s", domain),
			Description: fmt.Sprintf("Specialist agent for %s domain tasks", domain),
			Type:        AgentTypeSpecialist,
			Models:      []string{"llama3.2", "codellama"},
			Tools:       []string{"web_search", "data_analysis"},
			Config: map[string]interface{}{
				"specialization": domain,
				"expertise_level": "advanced",
			},
		}
	default:
		return nil, fmt.Errorf("unsupported agent type: %s", agentType)
	}

	err := e.CreateAgent(ctx, agent)
	if err != nil {
		return nil, err
	}

	return agent, nil
}

// SmartRouting intelligently routes tasks to appropriate models based on task type and content
func (e *Engine) SmartRouting(ctx context.Context, agentID string, input string, taskType string) (*TaskResult, error) {
	agent, err := e.GetAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}

	// Determine best model for the task
	modelName := e.selectBestModel(agent, taskType, input)

	task := &Task{
		Type:      taskType,
		Input:     input,
		Status:    TaskStatusPending,
		AgentID:   agentID,
		ModelName: modelName,
	}

	// Store task for tracking
	e.mu.Lock()
	e.tasks[task.ID] = task
	e.mu.Unlock()

	return e.ExecuteTask(ctx, task, agent)
}

// selectBestModel chooses the most appropriate model for a given task
func (e *Engine) selectBestModel(agent *Agent, taskType, input string) string {
	if len(agent.Models) == 0 {
		return ""
	}

	// Simple routing logic - this could be made much more sophisticated
	switch taskType {
	case TaskTypeGenerate:
		// For code-related content, prefer codellama
		if strings.Contains(strings.ToLower(input), "code") ||
			strings.Contains(strings.ToLower(input), "function") ||
			strings.Contains(strings.ToLower(input), "programming") {
			for _, model := range agent.Models {
				if strings.Contains(strings.ToLower(model), "code") {
					return model
				}
			}
		}
	case TaskTypeChat:
		// For conversational tasks, prefer general purpose models
		for _, model := range agent.Models {
			if strings.Contains(strings.ToLower(model), "llama") &&
				!strings.Contains(strings.ToLower(model), "code") {
				return model
			}
		}
	}

	// Default to first model or configured default
	if defaultModel, ok := agent.Config["default_model"].(string); ok {
		for _, model := range agent.Models {
			if model == defaultModel {
				return model
			}
		}
	}

	return agent.Models[0]
}

// MultiStepWorkflow executes a multi-step workflow with dependency management
func (e *Engine) MultiStepWorkflow(ctx context.Context, agentID string, steps []WorkflowStep) (*WorkflowResult, error) {
	agent, err := e.GetAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}

	result := &WorkflowResult{
		Steps:   make([]WorkflowStepResult, len(steps)),
		Success: true,
	}

	context := make(map[string]string)

	for i, step := range steps {
		// Replace placeholders with previous results
		input := e.replacePlaceholders(step.Input, context)

		task := &Task{
			Type:      step.Type,
			Input:     input,
			Status:    TaskStatusPending,
			AgentID:   agentID,
			ModelName: step.ModelName,
		}

		if task.ModelName == "" {
			task.ModelName = e.selectBestModel(agent, step.Type, input)
		}

		stepResult, err := e.ExecuteTask(ctx, task, agent)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Step %d failed: %v", i+1, err)
			break
		}

		// Store result for future steps
		context[fmt.Sprintf("step%d", i+1)] = stepResult.Output
		context[step.Name] = stepResult.Output

		result.Steps[i] = WorkflowStepResult{
			Name:      step.Name,
			Type:      step.Type,
			Input:     input,
			Output:    stepResult.Output,
			ModelUsed: stepResult.ModelUsed,
			Success:   true,
		}
	}

	return result, nil
}

// EnhancedCoordinatedWorkflow implements advanced agent coordination patterns
func (e *Engine) EnhancedCoordinatedWorkflow(ctx context.Context, coordinatorID string, tasks []CoordinatedTask) (*CoordinatedWorkflowResult, error) {
	coordinator, err := e.GetAgent(ctx, coordinatorID)
	if err != nil {
		return nil, err
	}

	if coordinator.Type != AgentTypeOrchestrator {
		return nil, fmt.Errorf("agent must be of type orchestrator for coordinated workflows")
	}

	result := &CoordinatedWorkflowResult{
		CoordinatorID: coordinatorID,
		Tasks:         make([]CoordinatedTaskResult, len(tasks)),
		Success:       true,
		StartTime:     time.Now(),
	}

	// Phase 1: Task Analysis and Agent Assignment
	for i, task := range tasks {
		// Intelligent agent selection based on task requirements
		selectedAgent, err := e.selectOptimalAgent(ctx, task)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Agent selection failed for task %d: %v", i, err)
			break
		}

		// Phase 2: Task Execution with Coordination
		executionResult, err := e.executeCoordinatedTask(ctx, task, selectedAgent, coordinator)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Task %d execution failed: %v", i, err)
			break
		}

		result.Tasks[i] = CoordinatedTaskResult{
			TaskID:      task.ID,
			AgentID:     selectedAgent.ID,
			Type:        task.Type,
			Input:       task.Input,
			Output:      executionResult.Output,
			Success:     true,
			Coordination: fmt.Sprintf("Coordinated by %s", coordinator.Name),
		}

		// Update coordinator state with task completion
		e.updateAgentState(coordinator, fmt.Sprintf("coordinated_task_%d", i), executionResult.Output)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Phase 3: Post-execution reflection and learning
	if coordinator.Type == AgentTypeReflective || 
	   (coordinator.Config != nil && coordinator.Config["enable_reflection"] == true) {
		reflection := e.performCoordinationReflection(coordinator, result)
		e.updateAgentState(coordinator, "workflow_reflection", reflection)
	}

	return result, nil
}

// selectOptimalAgent intelligently selects the best agent for a given task
func (e *Engine) selectOptimalAgent(ctx context.Context, task CoordinatedTask) (*Agent, error) {
	agents, err := e.ListAgents(ctx)
	if err != nil {
		return nil, err
	}

	var bestAgent *Agent
	var bestScore float64

	for _, agent := range agents {
		score := e.calculateAgentTaskFit(agent, task)
		if score > bestScore {
			bestScore = score
			bestAgent = agent
		}
	}

	if bestAgent == nil {
		return nil, fmt.Errorf("no suitable agent found for task type: %s", task.Type)
	}

	return bestAgent, nil
}

// calculateAgentTaskFit calculates how well an agent fits a particular task
func (e *Engine) calculateAgentTaskFit(agent *Agent, task CoordinatedTask) float64 {
	score := 0.0

	// Base score for agent type compatibility
	switch task.Type {
	case TaskTypeGenerate, TaskTypeChat:
		if agent.Type == AgentTypeGeneral || agent.Type == AgentTypeSpecialist {
			score += 0.3
		}
	case TaskTypeTool:
		if len(agent.Tools) > 0 {
			score += 0.4
		}
	case TaskTypeReflect:
		if agent.Type == AgentTypeReflective {
			score += 0.5
		}
	}

	// Model availability score
	if len(agent.Models) > 0 {
		score += 0.2
	}

	// Tool compatibility score
	if task.RequiredTools != nil {
		for _, requiredTool := range task.RequiredTools {
			for _, agentTool := range agent.Tools {
				if agentTool == requiredTool {
					score += 0.1
				}
			}
		}
	}

	// Recent performance score (based on state)
	if agent.State != nil && len(agent.State.Context) > 0 {
		// Recent activity indicates agent readiness
		timeSinceLastInteraction := time.Since(agent.State.LastInteraction)
		if timeSinceLastInteraction < time.Hour {
			score += 0.1
		}
	}

	return score
}

// executeCoordinatedTask executes a task with coordination oversight
func (e *Engine) executeCoordinatedTask(ctx context.Context, task CoordinatedTask, agent *Agent, coordinator *Agent) (*TaskResult, error) {
	// Create a regular Task from CoordinatedTask
	regularTask := &Task{
		ID:         task.ID,
		Type:       task.Type,
		Input:      task.Input,
		Status:     TaskStatusPending,
		AgentID:    agent.ID,
		Parameters: task.Parameters,
	}

	// Add coordination context
	if regularTask.Parameters == nil {
		regularTask.Parameters = make(map[string]interface{})
	}
	regularTask.Parameters["coordinator_id"] = coordinator.ID
	regularTask.Parameters["coordination_mode"] = "enhanced"

	// Execute the task
	result, err := e.ExecuteTask(ctx, regularTask, agent)
	if err != nil {
		return nil, err
	}

	// Update both agent and coordinator states
	e.updateAgentState(agent, "coordinated_execution", result.Output)
	e.updateAgentState(coordinator, "coordination_oversight", fmt.Sprintf("Supervised %s task by %s", task.Type, agent.Name))

	return result, nil
}

// performCoordinationReflection performs reflection on coordination patterns
func (e *Engine) performCoordinationReflection(coordinator *Agent, result *CoordinatedWorkflowResult) string {
	reflection := fmt.Sprintf("Coordination session completed: %d tasks in %v", 
		len(result.Tasks), result.Duration)

	successRate := 0.0
	for _, task := range result.Tasks {
		if task.Success {
			successRate += 1.0
		}
	}
	successRate = successRate / float64(len(result.Tasks)) * 100

	reflection += fmt.Sprintf(". Success rate: %.1f%%", successRate)

	if successRate >= 90 {
		reflection += ". Excellent coordination performance - agents worked efficiently together."
	} else if successRate >= 70 {
		reflection += ". Good coordination with room for optimization in agent selection."
	} else {
		reflection += ". Coordination challenges identified - reviewing agent assignment strategies."
	}

	return reflection
}

// CoordinatedTask represents a task within a coordinated workflow
type CoordinatedTask struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Input         string                 `json:"input"`
	RequiredTools []string               `json:"required_tools,omitempty"`
	Parameters    map[string]interface{} `json:"parameters,omitempty"`
	Priority      int                    `json:"priority,omitempty"`
}

// CoordinatedWorkflowResult represents the result of a coordinated workflow
type CoordinatedWorkflowResult struct {
	CoordinatorID string                   `json:"coordinator_id"`
	Tasks         []CoordinatedTaskResult  `json:"tasks"`
	Success       bool                     `json:"success"`
	Error         string                   `json:"error,omitempty"`
	StartTime     time.Time                `json:"start_time"`
	EndTime       time.Time                `json:"end_time"`
	Duration      time.Duration            `json:"duration"`
}

// CoordinatedTaskResult represents the result of a single coordinated task
type CoordinatedTaskResult struct {
	TaskID       string `json:"task_id"`
	AgentID      string `json:"agent_id"`
	Type         string `json:"type"`
	Input        string `json:"input"`
	Output       string `json:"output"`
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	Coordination string `json:"coordination"`
}

// replacePlaceholders replaces {{step1}}, {{step2}}, etc. with actual results
func (e *Engine) replacePlaceholders(input string, context map[string]string) string {
	result := input
	for key, value := range context {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// WorkflowStep represents a single step in a multi-step workflow
type WorkflowStep struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Input     string `json:"input"`
	ModelName string `json:"model_name,omitempty"`
}

// WorkflowResult represents the result of a multi-step workflow
type WorkflowResult struct {
	Steps   []WorkflowStepResult `json:"steps"`
	Success bool                 `json:"success"`
	Error   string               `json:"error,omitempty"`
}

// WorkflowStepResult represents the result of a single workflow step
type WorkflowStepResult struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Input     string `json:"input"`
	Output    string `json:"output"`
	ModelUsed string `json:"model_used"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}