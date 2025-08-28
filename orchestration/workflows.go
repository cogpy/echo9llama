package orchestration

import (
	"context"
	"fmt"
	"strings"
)

// DefaultAgent creates a default orchestration agent with common models
func (e *Engine) CreateDefaultAgent(ctx context.Context) (*Agent, error) {
	agent := &Agent{
		Name:        "default",
		Description: "Default orchestration agent for general tasks",
		Models:      []string{"llama3.2", "llama2", "codellama"},
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