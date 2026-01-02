package integration

import (
	"context"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// MockLLMProvider implements the full LLMProvider interface for testing
type MockLLMProvider struct{}

func (m *MockLLMProvider) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	return "Mock response for: " + prompt, nil
}

func (m *MockLLMProvider) StreamGenerate(ctx context.Context, prompt string, opts llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	go func() {
		ch <- llm.StreamChunk{Content: "Mock streaming response", Done: true}
		close(ch)
	}()
	return ch, nil
}

func (m *MockLLMProvider) Name() string {
	return "mock"
}

func (m *MockLLMProvider) Available() bool {
	return true
}

func (m *MockLLMProvider) MaxTokens() int {
	return 4096
}

func TestDeepTreeEchoHubCreation(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultHubConfig()

	hub := NewDeepTreeEchoHub(provider, config)
	if hub == nil {
		t.Fatal("Expected hub to be created")
	}

	// Verify initial state
	if hub.running {
		t.Error("Hub should not be running initially")
	}

	// Verify subsystems are initialized based on config
	if config.EnableAutonomousAgent && hub.autonomousAgent == nil {
		t.Error("Expected autonomous agent to be initialized")
	}

	if config.EnableUnifiedOrchestrator && hub.unifiedOrchestrator == nil {
		t.Error("Expected unified orchestrator to be initialized")
	}

	if config.EnableTelemetryShell && hub.telemetryShell == nil {
		t.Error("Expected telemetry shell to be initialized")
	}

	if config.EnableStateManager && hub.stateManager == nil {
		t.Error("Expected state manager to be initialized")
	}
}

func TestDeepTreeEchoHubStartStop(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultHubConfig()

	hub := NewDeepTreeEchoHub(provider, config)

	// Start hub
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}

	// Verify running
	if !hub.running {
		t.Error("Hub should be running after Start")
	}

	// Get status
	status := hub.GetStatus()
	if !status.Running {
		t.Error("Status should show running")
	}

	// Try to start again (should error)
	err = hub.Start()
	if err == nil {
		t.Error("Expected error when starting already running hub")
	}

	// Stop hub
	err = hub.Stop()
	if err != nil {
		t.Fatalf("Failed to stop hub: %v", err)
	}

	// Verify stopped
	if hub.running {
		t.Error("Hub should not be running after Stop")
	}

	// Try to stop again (should error)
	err = hub.Stop()
	if err == nil {
		t.Error("Expected error when stopping already stopped hub")
	}
}

func TestDeepTreeEchoHubGetGestalt(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultHubConfig()

	hub := NewDeepTreeEchoHub(provider, config)

	// Start hub
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	// Get gestalt
	gestalt := hub.GetGestalt()
	if gestalt == nil {
		t.Fatal("Expected gestalt to be returned")
	}

	// Verify hub section exists
	if _, ok := gestalt["hub"]; !ok {
		t.Error("Expected 'hub' section in gestalt")
	}
}

func TestDeepTreeEchoHubCommand(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultHubConfig()

	hub := NewDeepTreeEchoHub(provider, config)

	// Start hub
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	// Test get_status command
	response := hub.SendCommand(HubCommand{
		Type: "get_status",
	})

	if !response.Success {
		t.Error("Expected get_status command to succeed")
	}

	// Test inject_thought command
	response = hub.SendCommand(HubCommand{
		Type: "inject_thought",
		Payload: map[string]interface{}{
			"content": "Test thought",
			"tags": []string{"test"},
		},
	})

	if !response.Success {
		t.Errorf("Expected inject_thought command to succeed: %v", response.Error)
	}

	// Test unknown command
	response = hub.SendCommand(HubCommand{
		Type: "unknown_command",
	})

	if response.Success {
		t.Error("Expected unknown command to fail")
	}
}

func TestDeepTreeEchoHubMetrics(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultHubConfig()
	config.MainLoopInterval = 100 * time.Millisecond

	hub := NewDeepTreeEchoHub(provider, config)

	// Start hub
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	// Wait for a few cycles
	time.Sleep(500 * time.Millisecond)

	// Get status and check metrics
	status := hub.GetStatus()
	if status.Metrics == nil {
		t.Fatal("Expected metrics to be present")
	}

	// Verify some metrics are tracked
	if status.Metrics.TotalCycles == 0 {
		t.Error("Expected some cycles to have run")
	}
}

func TestCognitiveStateManagerBasics(t *testing.T) {
	csm := NewCognitiveStateManager()

	// Start
	err := csm.Start()
	if err != nil {
		t.Fatalf("Failed to start state manager: %v", err)
	}
	defer csm.Stop()

	// Add a thought
	csm.AddThought("Test thought", "active", "test", 0.8, []string{"test", "unit"})

	// Get recent thoughts
	thoughts := csm.GetRecentThoughts(10)
	if len(thoughts) != 1 {
		t.Errorf("Expected 1 thought, got %d", len(thoughts))
	}

	if thoughts[0].Content != "Test thought" {
		t.Errorf("Expected 'Test thought', got '%s'", thoughts[0].Content)
	}

	// Update phase
	csm.UpdatePhase("active")
	if csm.GetCurrentPhase() != "active" {
		t.Errorf("Expected phase 'active', got '%s'", csm.GetCurrentPhase())
	}

	// Get metrics
	metrics := csm.GetMetrics()
	if metrics["total_thoughts"].(uint64) != 1 {
		t.Error("Expected total_thoughts to be 1")
	}
}

func TestCognitiveStateManagerCallbacks(t *testing.T) {
	csm := NewCognitiveStateManager()

	var phaseChanged bool
	var thoughtGenerated bool

	csm.SetCallbacks(
		func(phase string) {
			phaseChanged = true
		},
		func(thought SharedThought) {
			thoughtGenerated = true
		},
		nil,
		nil,
	)

	// Start
	err := csm.Start()
	if err != nil {
		t.Fatalf("Failed to start state manager: %v", err)
	}
	defer csm.Stop()

	// Trigger phase change
	csm.UpdatePhase("test_phase")
	if !phaseChanged {
		t.Error("Expected phase change callback to be triggered")
	}

	// Trigger thought
	csm.AddThought("Test", "test", "test", 0.5, nil)
	if !thoughtGenerated {
		t.Error("Expected thought callback to be triggered")
	}
}

func TestDefaultConfigs(t *testing.T) {
	// Test HubConfig defaults
	hubConfig := DefaultHubConfig()
	if hubConfig.MainLoopInterval == 0 {
		t.Error("Expected MainLoopInterval to be set")
	}
	if !hubConfig.EnableAutonomousAgent {
		t.Error("Expected EnableAutonomousAgent to be true by default")
	}

	// Test BridgeConfig defaults
	bridgeConfig := DefaultBridgeConfig()
	if bridgeConfig.SyncInterval == 0 {
		t.Error("Expected SyncInterval to be set")
	}
	if !bridgeConfig.EnableBidirectional {
		t.Error("Expected EnableBidirectional to be true by default")
	}
}
