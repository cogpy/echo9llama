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