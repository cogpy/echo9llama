package orchestration

import (
	"context"
	"testing"
	"time"

	"github.com/ollama/ollama/api"
)

func TestNewEngine(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)

	if engine == nil {
		t.Error("NewEngine should return a non-nil engine")
	}

	if engine.agents == nil {
		t.Error("Engine should have initialized agents map")
	}

	if engine.tasks == nil {
		t.Error("Engine should have initialized tasks map")
	}
}

func TestCreateAgent(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	agent := &Agent{
		Name:        "test-agent",
		Description: "Test agent for unit testing",
		Models:      []string{"llama2"},
		Config:      map[string]interface{}{"key": "value"},
	}

	err := engine.CreateAgent(ctx, agent)
	if err != nil {
		t.Errorf("CreateAgent failed: %v", err)
	}

	if agent.ID == "" {
		t.Error("Agent ID should be generated")
	}

	if agent.CreatedAt.IsZero() {
		t.Error("Agent CreatedAt should be set")
	}

	if agent.UpdatedAt.IsZero() {
		t.Error("Agent UpdatedAt should be set")
	}

	// Verify agent was stored
	stored, err := engine.GetAgent(ctx, agent.ID)
	if err != nil {
		t.Errorf("GetAgent failed: %v", err)
	}

	if stored.Name != agent.Name {
		t.Errorf("Expected agent name %s, got %s", agent.Name, stored.Name)
	}
}

func TestListAgents(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Initially should be empty
	agents, err := engine.ListAgents(ctx)
	if err != nil {
		t.Errorf("ListAgents failed: %v", err)
	}

	if len(agents) != 0 {
		t.Errorf("Expected 0 agents, got %d", len(agents))
	}

	// Create an agent
	agent := &Agent{
		Name:        "test-agent",
		Description: "Test agent",
		Models:      []string{"llama2"},
	}

	err = engine.CreateAgent(ctx, agent)
	if err != nil {
		t.Errorf("CreateAgent failed: %v", err)
	}

	// Now should have one agent
	agents, err = engine.ListAgents(ctx)
	if err != nil {
		t.Errorf("ListAgents failed: %v", err)
	}

	if len(agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(agents))
	}

	if agents[0].Name != agent.Name {
		t.Errorf("Expected agent name %s, got %s", agent.Name, agents[0].Name)
	}
}

func TestUpdateAgent(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Create an agent first
	agent := &Agent{
		Name:        "test-agent",
		Description: "Original description",
		Models:      []string{"llama2"},
	}

	err := engine.CreateAgent(ctx, agent)
	if err != nil {
		t.Errorf("CreateAgent failed: %v", err)
	}

	originalUpdatedAt := agent.UpdatedAt

	// Update the agent
	time.Sleep(time.Millisecond) // Ensure time difference
	agent.Description = "Updated description"
	err = engine.UpdateAgent(ctx, agent)
	if err != nil {
		t.Errorf("UpdateAgent failed: %v", err)
	}

	// Verify update
	updated, err := engine.GetAgent(ctx, agent.ID)
	if err != nil {
		t.Errorf("GetAgent failed: %v", err)
	}

	if updated.Description != "Updated description" {
		t.Errorf("Expected description 'Updated description', got '%s'", updated.Description)
	}

	if !updated.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated")
	}
}

func TestDeleteAgent(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Create an agent first
	agent := &Agent{
		Name:        "test-agent",
		Description: "To be deleted",
		Models:      []string{"llama2"},
	}

	err := engine.CreateAgent(ctx, agent)
	if err != nil {
		t.Errorf("CreateAgent failed: %v", err)
	}

	// Delete the agent
	err = engine.DeleteAgent(ctx, agent.ID)
	if err != nil {
		t.Errorf("DeleteAgent failed: %v", err)
	}

	// Verify agent is gone
	_, err = engine.GetAgent(ctx, agent.ID)
	if err == nil {
		t.Error("GetAgent should have failed for deleted agent")
	}
}

func TestGetNonExistentAgent(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	_, err := engine.GetAgent(ctx, "non-existent-id")
	if err == nil {
		t.Error("GetAgent should have failed for non-existent agent")
	}
}

func TestDeleteNonExistentAgent(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	err := engine.DeleteAgent(ctx, "non-existent-id")
	if err == nil {
		t.Error("DeleteAgent should have failed for non-existent agent")
	}
}

func TestUpdateNonExistentAgent(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	agent := &Agent{
		ID:          "non-existent-id",
		Name:        "test-agent",
		Description: "Test description",
		Models:      []string{"llama2"},
	}

	err := engine.UpdateAgent(ctx, agent)
	if err == nil {
		t.Error("UpdateAgent should have failed for non-existent agent")
	}
}

func TestNewAgentTypes(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Test creating different agent types
	testCases := []struct {
		agentType AgentType
		domain    string
	}{
		{AgentTypeReflective, "analysis"},
		{AgentTypeOrchestrator, "coordination"},
		{AgentTypeSpecialist, "coding"},
	}

	for _, tc := range testCases {
		agent, err := engine.CreateSpecializedAgent(ctx, tc.agentType, tc.domain)
		if err != nil {
			t.Errorf("CreateSpecializedAgent failed for type %s: %v", tc.agentType, err)
			continue
		}

		if agent.Type != tc.agentType {
			t.Errorf("Expected agent type %s, got %s", tc.agentType, agent.Type)
		}

		if agent.State == nil {
			t.Error("Agent state should be initialized")
		}

		if agent.State.Memory == nil {
			t.Error("Agent memory should be initialized")
		}
	}
}

func TestToolRegistration(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)

	// Register default tools
	RegisterDefaultTools(engine)

	tools := engine.GetAvailableTools()
	if len(tools) == 0 {
		t.Error("Expected tools to be registered")
	}

	// Check for specific tools
	expectedTools := []string{"web_search", "calculator"}
	for _, expectedTool := range expectedTools {
		found := false
		for _, tool := range tools {
			if tool == expectedTool {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tool '%s' not found in registered tools", expectedTool)
		}
	}
}

func TestPluginRegistration(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)

	// Register default plugins
	RegisterDefaultPlugins(engine)

	plugins := engine.GetAvailablePlugins()
	if len(plugins) == 0 {
		t.Error("Expected plugins to be registered")
	}

	// Check for specific plugin
	expectedPlugin := "data_analysis"
	found := false
	for _, plugin := range plugins {
		if plugin == expectedPlugin {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected plugin '%s' not found in registered plugins", expectedPlugin)
	}
}

func TestEnhancedTaskExecution(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Register tools and plugins for testing
	RegisterDefaultTools(engine)
	RegisterDefaultPlugins(engine)

	// Create a general agent
	agent := &Agent{
		Name:        "test-agent",
		Description: "Test agent for enhanced features",
		Type:        AgentTypeGeneral,
		Models:      []string{"llama2"},
		Tools:       []string{"calculator"},
	}

	err := engine.CreateAgent(ctx, agent)
	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}

	// Test tool task execution
	toolTask := &Task{
		ID:      "tool-task-1",
		Type:    TaskTypeTool,
		Input:   "Calculate 2 + 3",
		Status:  TaskStatusPending,
		AgentID: agent.ID,
		Parameters: map[string]interface{}{
			"tool": map[string]interface{}{
				"name": "calculator",
				"parameters": map[string]interface{}{
					"operation": "add",
					"a":         2.0,
					"b":         3.0,
				},
			},
		},
	}

	result, err := engine.ExecuteTask(ctx, toolTask, agent)
	if err != nil {
		t.Errorf("Tool task execution failed: %v", err)
	}

	if result == nil {
		t.Error("Tool task result should not be nil")
	}

	// Test plugin task execution
	pluginTask := &Task{
		ID:      "plugin-task-1",
		Type:    TaskTypePlugin,
		Input:   "Analyze this sample data",
		Status:  TaskStatusPending,
		AgentID: agent.ID,
		Parameters: map[string]interface{}{
			"plugin_name": "data_analysis",
			"type":        "summary",
		},
	}

	result, err = engine.ExecuteTask(ctx, pluginTask, agent)
	if err != nil {
		t.Errorf("Plugin task execution failed: %v", err)
	}

	if result == nil {
		t.Error("Plugin task result should not be nil")
	}
}

func TestReflectiveAgent(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Create a reflective agent
	agent, err := engine.CreateSpecializedAgent(ctx, AgentTypeReflective, "self-analysis")
	if err != nil {
		t.Fatalf("CreateSpecializedAgent failed: %v", err)
	}

	// Test reflection task
	reflectTask := &Task{
		ID:      "reflect-task-1",
		Type:    TaskTypeReflect,
		Input:   "Analyze recent performance and learning patterns",
		Status:  TaskStatusPending,
		AgentID: agent.ID,
	}

	result, err := engine.ExecuteTask(ctx, reflectTask, agent)
	if err != nil {
		t.Errorf("Reflection task execution failed: %v", err)
	}

	if result == nil {
		t.Error("Reflection task result should not be nil")
	}

	// Check that agent state was updated
	updatedAgent, err := engine.GetAgent(ctx, agent.ID)
	if err != nil {
		t.Errorf("GetAgent failed: %v", err)
	}

	if updatedAgent.State == nil || len(updatedAgent.State.Context) == 0 {
		t.Error("Agent state should be updated with reflection context")
	}
}