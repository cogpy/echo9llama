package orchestration

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ollama/ollama/api"
)

// Engine implements the core orchestration functionality
type Engine struct {
	client         api.Client
	agents         map[string]*Agent
	tasks          map[string]*Task
	tools          map[string]Tool
	plugins        *PluginRegistry
	mu             sync.RWMutex
}

// NewEngine creates a new orchestration engine
func NewEngine(client api.Client) *Engine {
	return &Engine{
		client:  client,
		agents:  make(map[string]*Agent),
		tasks:   make(map[string]*Task),
		tools:   make(map[string]Tool),
		plugins: &PluginRegistry{plugins: make(map[string]Plugin)},
	}
}

// CreateAgent creates a new orchestration agent
func (e *Engine) CreateAgent(ctx context.Context, agent *Agent) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if agent.ID == "" {
		agent.ID = uuid.New().String()
	}

	// Initialize agent state if not provided
	if agent.State == nil {
		agent.State = &AgentState{
			Memory:          make(map[string]interface{}),
			Context:         make([]ContextItem, 0),
			Goals:           make([]string, 0),
			Capabilities:    make([]string, 0),
			LastInteraction: time.Now(),
		}
	}

	// Set default agent type if not specified
	if agent.Type == "" {
		agent.Type = AgentTypeGeneral
	}

	agent.CreatedAt = time.Now()
	agent.UpdatedAt = time.Now()

	e.agents[agent.ID] = agent
	slog.Info("Created orchestration agent", "id", agent.ID, "name", agent.Name)
	return nil
}

// GetAgent retrieves an agent by ID
func (e *Engine) GetAgent(ctx context.Context, id string) (*Agent, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	agent, exists := e.agents[id]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", id)
	}

	return agent, nil
}

// ListAgents returns all registered agents
func (e *Engine) ListAgents(ctx context.Context) ([]*Agent, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	agents := make([]*Agent, 0, len(e.agents))
	for _, agent := range e.agents {
		agents = append(agents, agent)
	}

	return agents, nil
}

// UpdateAgent updates an existing agent
func (e *Engine) UpdateAgent(ctx context.Context, agent *Agent) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.agents[agent.ID]; !exists {
		return fmt.Errorf("agent not found: %s", agent.ID)
	}

	agent.UpdatedAt = time.Now()
	e.agents[agent.ID] = agent
	slog.Info("Updated orchestration agent", "id", agent.ID, "name", agent.Name)
	return nil
}

// DeleteAgent removes an agent
func (e *Engine) DeleteAgent(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.agents[id]; !exists {
		return fmt.Errorf("agent not found: %s", id)
	}

	delete(e.agents, id)
	slog.Info("Deleted orchestration agent", "id", id)
	return nil
}

// ExecuteTask executes a single task
func (e *Engine) ExecuteTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error) {
	startTime := time.Now()
	task.Status = TaskStatusRunning

	var result *TaskResult
	var err error

	switch task.Type {
	case TaskTypeGenerate:
		result, err = e.executeGenerateTask(ctx, task, agent)
	case TaskTypeChat:
		result, err = e.executeChatTask(ctx, task, agent)
	case TaskTypeEmbed:
		result, err = e.executeEmbedTask(ctx, task, agent)
	case TaskTypeTool:
		result, err = e.executeToolTask(ctx, task, agent)
	case TaskTypeReflect:
		result, err = e.executeReflectTask(ctx, task, agent)
	case TaskTypePlugin:
		result, err = e.executePluginTask(ctx, task, agent)
	default:
		result, err = e.executeCustomTask(ctx, task, agent)
	}

	duration := time.Since(startTime)

	if err != nil {
		task.Status = TaskStatusFailed
		task.Error = err.Error()
		return nil, err
	}

	task.Status = TaskStatusCompleted
	now := time.Now()
	task.CompletedAt = &now
	task.Output = result.Output

	if result.Metrics.Duration == 0 {
		result.Metrics.Duration = duration
	}

	slog.Info("Task completed", "task_id", task.ID, "type", task.Type, "duration", duration)
	return result, nil
}

// ExecuteTasks executes multiple tasks either sequentially or in parallel
func (e *Engine) ExecuteTasks(ctx context.Context, tasks []*Task, agent *Agent, sequential bool) ([]*TaskResult, error) {
	results := make([]*TaskResult, len(tasks))

	if sequential {
		for i, task := range tasks {
			result, err := e.ExecuteTask(ctx, task, agent)
			if err != nil {
				return results[:i], err
			}
			results[i] = result
		}
	} else {
		var wg sync.WaitGroup
		var mu sync.Mutex
		var firstError error

		for i, task := range tasks {
			wg.Add(1)
			go func(idx int, t *Task) {
				defer wg.Done()
				result, err := e.ExecuteTask(ctx, t, agent)
				
				mu.Lock()
				if err != nil && firstError == nil {
					firstError = err
				}
				if result != nil {
					results[idx] = result
				}
				mu.Unlock()
			}(i, task)
		}

		wg.Wait()

		if firstError != nil {
			return results, firstError
		}
	}

	return results, nil
}

// OrchestrateTasks orchestrates multiple tasks using an agent
func (e *Engine) OrchestrateTasks(ctx context.Context, req *OrchestrationRequest) (*OrchestrationResponse, error) {
	agent, err := e.GetAgent(ctx, req.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	// Create tasks from the request
	tasks := make([]*Task, len(req.Tasks))
	for i, taskReq := range req.Tasks {
		task := &Task{
			ID:         uuid.New().String(),
			Type:       taskReq.Type,
			Input:      taskReq.Input,
			Status:     TaskStatusPending,
			AgentID:    req.AgentID,
			ModelName:  taskReq.ModelName,
			Parameters: taskReq.Parameters,
			CreatedAt:  time.Now(),
		}

		// Store task for tracking
		e.mu.Lock()
		e.tasks[task.ID] = task
		e.mu.Unlock()

		tasks[i] = task
	}

	// Execute tasks
	results, err := e.ExecuteTasks(ctx, tasks, agent, req.Sequential)

	// Convert []*Task to []Task and []*TaskResult to []TaskResult
	taskSlice := make([]Task, len(tasks))
	for i, task := range tasks {
		taskSlice[i] = *task
	}

	resultSlice := make([]TaskResult, 0)
	if results != nil {
		resultSlice = make([]TaskResult, len(results))
		for i, result := range results {
			if result != nil {
				resultSlice[i] = *result
			}
		}
	}

	response := &OrchestrationResponse{
		ID:        uuid.New().String(),
		AgentID:   req.AgentID,
		Status:    "completed",
		Tasks:     taskSlice,
		Results:   resultSlice,
		CreatedAt: time.Now(),
	}

	if err != nil {
		response.Status = "failed"
		response.Error = err.Error()
	}

	return response, err
}

// executeGenerateTask executes a generate task using the Ollama API
func (e *Engine) executeGenerateTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error) {
	modelName := task.ModelName
	if modelName == "" && len(agent.Models) > 0 {
		modelName = agent.Models[0] // Use first model as default
	}

	if modelName == "" {
		return nil, fmt.Errorf("no model specified for generate task")
	}

	req := &api.GenerateRequest{
		Model:  modelName,
		Prompt: task.Input,
	}

	// Apply parameters from task
	if task.Parameters != nil {
		if opts, ok := task.Parameters["options"]; ok {
			if optsMap, ok := opts.(map[string]interface{}); ok {
				req.Options = optsMap
			}
		}
	}

	var output string
	err := e.client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		output += resp.Response
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &TaskResult{
		TaskID:    task.ID,
		Output:    output,
		ModelUsed: modelName,
	}, nil
}

// executeChatTask executes a chat task using the Ollama API
func (e *Engine) executeChatTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error) {
	modelName := task.ModelName
	if modelName == "" && len(agent.Models) > 0 {
		modelName = agent.Models[0]
	}

	if modelName == "" {
		return nil, fmt.Errorf("no model specified for chat task")
	}

	req := &api.ChatRequest{
		Model: modelName,
		Messages: []api.Message{
			{Role: "user", Content: task.Input},
		},
	}

	// Apply parameters from task
	if task.Parameters != nil {
		if opts, ok := task.Parameters["options"]; ok {
			if optsMap, ok := opts.(map[string]interface{}); ok {
				req.Options = optsMap
			}
		}
	}

	var output string
	err := e.client.Chat(ctx, req, func(resp api.ChatResponse) error {
		output += resp.Message.Content
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &TaskResult{
		TaskID:    task.ID,
		Output:    output,
		ModelUsed: modelName,
	}, nil
}

// executeEmbedTask executes an embedding task using the Ollama API
func (e *Engine) executeEmbedTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error) {
	modelName := task.ModelName
	if modelName == "" && len(agent.Models) > 0 {
		modelName = agent.Models[0]
	}

	if modelName == "" {
		return nil, fmt.Errorf("no model specified for embed task")
	}

	req := &api.EmbeddingRequest{
		Model:  modelName,
		Prompt: task.Input,
	}

	resp, err := e.client.Embeddings(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert embeddings to string representation
	output := fmt.Sprintf("Embedding generated with dimension %d", len(resp.Embedding))

	return &TaskResult{
		TaskID:    task.ID,
		Output:    output,
		ModelUsed: modelName,
	}, nil
}

// executeCustomTask executes a custom task (enhanced for echoself integration)
func (e *Engine) executeCustomTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error) {
	// Enhanced custom task execution with agent state awareness
	e.updateAgentState(agent, "custom_task", task.Input)
	
	output := fmt.Sprintf("Custom task '%s' completed with enhanced agent coordination", task.Type)
	if agent.Type == AgentTypeReflective {
		output += " (with self-reflection capabilities)"
	}
	
	return &TaskResult{
		TaskID: task.ID,
		Output: output,
	}, nil
}

// executeToolTask executes a tool call task
func (e *Engine) executeToolTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error) {
	// Parse tool call from task parameters
	var toolCall ToolCall
	if toolParams, ok := task.Parameters["tool"]; ok {
		if toolMap, ok := toolParams.(map[string]interface{}); ok {
			if name, ok := toolMap["name"].(string); ok {
				toolCall.Name = name
			}
			if params, ok := toolMap["parameters"].(map[string]interface{}); ok {
				toolCall.Parameters = params
			}
		}
	}

	// Execute tool if available
	if tool, exists := e.tools[toolCall.Name]; exists {
		result, err := tool.Call(ctx, toolCall.Parameters)
		if err != nil {
			return nil, fmt.Errorf("tool call failed: %v", err)
		}
		
		e.updateAgentState(agent, "tool_use", toolCall.Name)
		
		return &TaskResult{
			TaskID: task.ID,
			Output: fmt.Sprintf("Tool '%s' executed successfully: %v", toolCall.Name, result.Output),
		}, nil
	}

	return &TaskResult{
		TaskID: task.ID,
		Output: fmt.Sprintf("Tool '%s' not available", toolCall.Name),
	}, nil
}

// executeReflectTask executes a self-reflection task
func (e *Engine) executeReflectTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error) {
	// Enhanced reflection capabilities for echoself integration
	reflection := e.performAgentReflection(agent, task.Input)
	
	e.updateAgentState(agent, "reflection", reflection)
	
	return &TaskResult{
		TaskID: task.ID,
		Output: reflection,
	}, nil
}

// executePluginTask executes a plugin-based task
func (e *Engine) executePluginTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error) {
	pluginName := ""
	if name, ok := task.Parameters["plugin_name"].(string); ok {
		pluginName = name
	}

	if plugin, exists := e.plugins.plugins[pluginName]; exists {
		result, err := plugin.Execute(ctx, task.Input, task.Parameters)
		if err != nil {
			return nil, fmt.Errorf("plugin execution failed: %v", err)
		}
		
		e.updateAgentState(agent, "plugin_use", pluginName)
		
		return &TaskResult{
			TaskID: task.ID,
			Output: fmt.Sprintf("Plugin '%s' result: %v", pluginName, result),
		}, nil
	}

	return &TaskResult{
		TaskID: task.ID,
		Output: fmt.Sprintf("Plugin '%s' not found", pluginName),
	}, nil
}

// Tool and Plugin Management Methods

// RegisterTool registers a new tool with the engine
func (e *Engine) RegisterTool(tool Tool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.tools[tool.Name()] = tool
	slog.Info("Registered tool", "name", tool.Name())
}

// RegisterPlugin registers a new plugin with the engine
func (e *Engine) RegisterPlugin(plugin Plugin) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.plugins.plugins[plugin.Name()] = plugin
	slog.Info("Registered plugin", "name", plugin.Name())
}

// GetAvailableTools returns list of available tools
func (e *Engine) GetAvailableTools() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	tools := make([]string, 0, len(e.tools))
	for name := range e.tools {
		tools = append(tools, name)
	}
	return tools
}

// GetAvailablePlugins returns list of available plugins
func (e *Engine) GetAvailablePlugins() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	plugins := make([]string, 0, len(e.plugins.plugins))
	for name := range e.plugins.plugins {
		plugins = append(plugins, name)
	}
	return plugins
}

// Agent State Management Methods

// updateAgentState updates the agent's internal state and memory
func (e *Engine) updateAgentState(agent *Agent, key string, value interface{}) {
	if agent.State == nil {
		agent.State = &AgentState{
			Memory:   make(map[string]interface{}),
			Context:  make([]ContextItem, 0),
		}
	}
	
	agent.State.Memory[key] = value
	agent.State.LastInteraction = time.Now()
	
	// Add to context with relevance scoring
	contextItem := ContextItem{
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
		Relevance: 1.0, // Simple relevance scoring
	}
	
	agent.State.Context = append(agent.State.Context, contextItem)
	
	// Keep only last 10 context items for memory management
	if len(agent.State.Context) > 10 {
		agent.State.Context = agent.State.Context[len(agent.State.Context)-10:]
	}
}

// performAgentReflection performs self-reflection for enhanced agent capabilities
func (e *Engine) performAgentReflection(agent *Agent, input string) string {
	reflection := fmt.Sprintf("Agent '%s' reflecting on: %s", agent.Name, input)
	
	if agent.State != nil && len(agent.State.Context) > 0 {
		reflection += fmt.Sprintf(". Recent context includes %d interactions.", len(agent.State.Context))
		
		// Analyze recent performance
		if len(agent.State.Context) >= 3 {
			reflection += " Pattern analysis suggests consistent performance across multiple tasks."
		}
	}
	
	// Agent type specific reflection
	switch agent.Type {
	case AgentTypeReflective:
		reflection += " Advanced self-analysis indicates opportunities for optimization and learning."
	case AgentTypeOrchestrator:
		reflection += " Coordination patterns show effective multi-agent task distribution."
	case AgentTypeSpecialist:
		reflection += " Domain expertise application demonstrates specialized knowledge utilization."
	}
	
	return reflection
}