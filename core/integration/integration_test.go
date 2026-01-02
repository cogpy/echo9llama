package integration

import (
	"context"
	"testing"
	"time"
)

func TestIntegratedDeepTreeEchoCreation(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultIntegrationConfig(provider)

	integrated, err := NewIntegratedDeepTreeEcho(config)
	if err != nil {
		t.Fatalf("Failed to create integrated system: %v", err)
	}

	if integrated == nil {
		t.Fatal("Expected integrated system to be created")
	}

	if !integrated.initialized {
		t.Error("Expected system to be initialized")
	}

	if integrated.started {
		t.Error("Expected system to not be started yet")
	}

	// Verify hub was created
	if integrated.hub == nil {
		t.Error("Expected hub to be created")
	}
}

func TestIntegratedDeepTreeEchoNilProvider(t *testing.T) {
	config := DefaultIntegrationConfig(nil)
	config.LLMProvider = nil

	_, err := NewIntegratedDeepTreeEcho(config)
	if err == nil {
		t.Error("Expected error when LLM provider is nil")
	}
}

func TestIntegratedDeepTreeEchoStartStop(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultIntegrationConfig(provider)
	config.EnableOrchestration = true
	config.EnableBridge = false // Disable bridge for simpler test

	integrated, err := NewIntegratedDeepTreeEcho(config)
	if err != nil {
		t.Fatalf("Failed to create integrated system: %v", err)
	}

	ctx := context.Background()

	// Start
	err = integrated.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start integrated system: %v", err)
	}

	if !integrated.started {
		t.Error("Expected system to be started")
	}

	// Try to start again (should error)
	err = integrated.Start(ctx)
	if err == nil {
		t.Error("Expected error when starting already started system")
	}

	// Get status
	status := integrated.GetStatus()
	if !status["started"].(bool) {
		t.Error("Expected status to show started")
	}

	// Stop
	err = integrated.Stop()
	if err != nil {
		t.Fatalf("Failed to stop integrated system: %v", err)
	}

	if integrated.started {
		t.Error("Expected system to be stopped")
	}

	// Try to stop again (should error)
	err = integrated.Stop()
	if err == nil {
		t.Error("Expected error when stopping already stopped system")
	}
}

func TestIntegratedDeepTreeEchoGetGestalt(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultIntegrationConfig(provider)
	config.EnableOrchestration = true

	integrated, err := NewIntegratedDeepTreeEcho(config)
	if err != nil {
		t.Fatalf("Failed to create integrated system: %v", err)
	}

	ctx := context.Background()
	err = integrated.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start integrated system: %v", err)
	}
	defer integrated.Stop()

	gestalt := integrated.GetGestalt()
	if gestalt == nil {
		t.Fatal("Expected gestalt to be returned")
	}

	// Check for hub section
	if _, ok := gestalt["hub"]; !ok {
		t.Error("Expected 'hub' section in gestalt")
	}

	// Check for orchestration section
	if _, ok := gestalt["orchestration"]; !ok {
		t.Error("Expected 'orchestration' section in gestalt")
	}
}

func TestIntegratedDeepTreeEchoInjectThought(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultIntegrationConfig(provider)

	integrated, err := NewIntegratedDeepTreeEcho(config)
	if err != nil {
		t.Fatalf("Failed to create integrated system: %v", err)
	}

	// Try to inject thought before starting (should error)
	err = integrated.InjectThought("Test thought", []string{"test"})
	if err == nil {
		t.Error("Expected error when injecting thought before start")
	}

	// Start system
	ctx := context.Background()
	err = integrated.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start integrated system: %v", err)
	}
	defer integrated.Stop()

	// Now inject thought
	err = integrated.InjectThought("Test thought", []string{"test"})
	if err != nil {
		t.Errorf("Failed to inject thought: %v", err)
	}
}

func TestIntegratedDeepTreeEchoStoreKnowledge(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultIntegrationConfig(provider)

	integrated, err := NewIntegratedDeepTreeEcho(config)
	if err != nil {
		t.Fatalf("Failed to create integrated system: %v", err)
	}

	// Try to store knowledge before starting (should error)
	err = integrated.StoreKnowledge("subject", "predicate", "object")
	if err == nil {
		t.Error("Expected error when storing knowledge before start")
	}

	// Start system
	ctx := context.Background()
	err = integrated.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start integrated system: %v", err)
	}
	defer integrated.Stop()

	// Now store knowledge
	err = integrated.StoreKnowledge("TestSubject", "has_property", "TestValue")
	if err != nil {
		t.Errorf("Failed to store knowledge: %v", err)
	}

	// Query knowledge
	quads := integrated.QueryKnowledge("TestSubject")
	if len(quads) == 0 {
		t.Error("Expected to find stored knowledge")
	}
}

func TestIntegratedDeepTreeEchoLifecycleHooks(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultIntegrationConfig(provider)

	integrated, err := NewIntegratedDeepTreeEcho(config)
	if err != nil {
		t.Fatalf("Failed to create integrated system: %v", err)
	}

	var startHookCalled bool
	var stopHookCalled bool

	integrated.OnStart(func() error {
		startHookCalled = true
		return nil
	})

	integrated.OnStop(func() error {
		stopHookCalled = true
		return nil
	})

	ctx := context.Background()
	err = integrated.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start integrated system: %v", err)
	}

	if !startHookCalled {
		t.Error("Expected start hook to be called")
	}

	err = integrated.Stop()
	if err != nil {
		t.Fatalf("Failed to stop integrated system: %v", err)
	}

	if !stopHookCalled {
		t.Error("Expected stop hook to be called")
	}
}

func TestIntegratedDeepTreeEchoGetComponents(t *testing.T) {
	provider := &MockLLMProvider{}
	config := DefaultIntegrationConfig(provider)
	config.EnableOrchestration = true
	config.EnableBridge = true

	integrated, err := NewIntegratedDeepTreeEcho(config)
	if err != nil {
		t.Fatalf("Failed to create integrated system: %v", err)
	}

	// Test GetHub
	hub := integrated.GetHub()
	if hub == nil {
		t.Error("Expected hub to be returned")
	}

	// Test GetOrchestration
	orch := integrated.GetOrchestration()
	if orch == nil {
		t.Error("Expected orchestration to be returned")
	}

	// Note: Bridge is created during Start, so it may be nil here
}

func TestCognitiveBridgeCreation(t *testing.T) {
	// Create a bridge directly for testing
	config := DefaultBridgeConfig()
	bridge := NewCognitiveBridge(nil, nil, config)

	if bridge == nil {
		t.Fatal("Expected bridge to be created")
	}

	if bridge.IsRunning() {
		t.Error("Expected bridge to not be running initially")
	}
}

func TestCognitiveBridgeStartStop(t *testing.T) {
	config := DefaultBridgeConfig()
	bridge := NewCognitiveBridge(nil, nil, config)

	// Start bridge
	err := bridge.Start()
	if err != nil {
		t.Fatalf("Failed to start bridge: %v", err)
	}

	if !bridge.IsRunning() {
		t.Error("Expected bridge to be running")
	}

	// Try to start again (should error)
	err = bridge.Start()
	if err == nil {
		t.Error("Expected error when starting already running bridge")
	}

	// Stop bridge
	err = bridge.Stop()
	if err != nil {
		t.Fatalf("Failed to stop bridge: %v", err)
	}

	if bridge.IsRunning() {
		t.Error("Expected bridge to not be running")
	}

	// Try to stop again (should error)
	err = bridge.Stop()
	if err == nil {
		t.Error("Expected error when stopping already stopped bridge")
	}
}

func TestCognitiveBridgeMetrics(t *testing.T) {
	config := DefaultBridgeConfig()
	bridge := NewCognitiveBridge(nil, nil, config)

	err := bridge.Start()
	if err != nil {
		t.Fatalf("Failed to start bridge: %v", err)
	}
	defer bridge.Stop()

	metrics := bridge.GetMetrics()
	if metrics == nil {
		t.Fatal("Expected metrics to be returned")
	}

	if _, ok := metrics["running"]; !ok {
		t.Error("Expected 'running' field in metrics")
	}

	if _, ok := metrics["messages_sent"]; !ok {
		t.Error("Expected 'messages_sent' field in metrics")
	}

	if _, ok := metrics["sync_count"]; !ok {
		t.Error("Expected 'sync_count' field in metrics")
	}
}

func TestCognitiveBridgeForceSync(t *testing.T) {
	config := DefaultBridgeConfig()
	bridge := NewCognitiveBridge(nil, nil, config)

	err := bridge.Start()
	if err != nil {
		t.Fatalf("Failed to start bridge: %v", err)
	}
	defer bridge.Stop()

	// Get initial sync count
	initialMetrics := bridge.GetMetrics()
	initialSyncCount := initialMetrics["sync_count"].(uint64)

	// Force sync
	bridge.ForceSync()

	// Wait briefly
	time.Sleep(10 * time.Millisecond)

	// Get new sync count
	newMetrics := bridge.GetMetrics()
	newSyncCount := newMetrics["sync_count"].(uint64)

	if newSyncCount <= initialSyncCount {
		t.Error("Expected sync count to increase after ForceSync")
	}
}
