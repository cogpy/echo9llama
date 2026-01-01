package integration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
)

// DeepTreeEchoHub is the unified integration hub that connects all layers of the
// Deep Tree Echo cognitive architecture. It serves as the central coordinator
// for autonomous AGI operation, connecting:
// - Core autonomous agent systems
// - Cognitive state management
// - Memory and consciousness integration
// - Echobeats 12-step cognitive loop
// - EchoDream knowledge consolidation
// - Wisdom synthesis and goal pursuit
type DeepTreeEchoHub struct {
	mu sync.RWMutex
	ctx context.Context
	cancel context.CancelFunc

	// Core autonomous agent
	autonomousAgent *deeptreeecho.AutonomousAgent

	// Unified orchestrator for cognitive loops
	unifiedOrchestrator *deeptreeecho.UnifiedAutonomousOrchestrator

	// Global telemetry shell for gestalt perception
	telemetryShell *deeptreeecho.GlobalTelemetryShell

	// Cognitive state manager for cross-subsystem integration
	stateManager *CognitiveStateManager

	// LLM provider
	llmProvider llm.LLMProvider

	// Hub configuration
	config HubConfig

	// Integration channels
	eventChan chan HubEvent
	commandChan chan HubCommand
	statusChan chan HubStatus

	// Metrics
	metrics *HubMetrics

	// Running state
	running bool
	startTime time.Time
}

// HubConfig holds configuration for the integration hub
type HubConfig struct {
	// Identity
	AgentID string
	SessionName string

	// Timing
	MainLoopInterval time.Duration
	StateUpdateInterval time.Duration
	MetricsSyncInterval time.Duration

	// Features
	EnableAutonomousAgent bool
	EnableUnifiedOrchestrator bool
	EnableTelemetryShell bool
	EnableStateManager bool

	// Persistence
	EnablePersistence bool
	PersistenceInterval time.Duration
}

// DefaultHubConfig returns the default hub configuration
func DefaultHubConfig() HubConfig {
	return HubConfig{
		AgentID: fmt.Sprintf("dte-hub-%d", time.Now().Unix()),
		SessionName: fmt.Sprintf("dte-session-%d", time.Now().Unix()),
		MainLoopInterval: 5 * time.Second,
		StateUpdateInterval: 1 * time.Second,
		MetricsSyncInterval: 30 * time.Second,
		EnableAutonomousAgent: true,
		EnableUnifiedOrchestrator: true,
		EnableTelemetryShell: true,
		EnableStateManager: true,
		EnablePersistence: true,
		PersistenceInterval: 5 * time.Minute,
	}
}

// HubEvent represents an event from the hub
type HubEvent struct {
	Type string
	Source string
	Data interface{}
	Timestamp time.Time
}

// HubCommand represents a command to the hub
type HubCommand struct {
	Type string
	Target string
	Payload interface{}
	ResponseChan chan HubCommandResponse
}

// HubCommandResponse represents a response to a hub command
type HubCommandResponse struct {
	Success bool
	Data interface{}
	Error error
}

// HubStatus represents the current status of the hub
type HubStatus struct {
	Running bool
	Uptime time.Duration
	AgentStatus map[string]interface{}
	OrchestratorStatus map[string]interface{}
	TelemetryStatus map[string]interface{}
	StateManagerStatus map[string]interface{}
	Metrics *HubMetrics
	Timestamp time.Time
}

// HubMetrics tracks metrics for the hub
type HubMetrics struct {
	mu sync.RWMutex
	TotalEvents uint64
	TotalCommands uint64
	TotalCycles uint64
	TotalThoughts uint64
	TotalGoals uint64
	TotalWisdom uint64
	EventsPerSecond float64
	CommandsProcessed uint64
	ErrorCount uint64
	LastError string
	LastErrorTime time.Time
}

// NewDeepTreeEchoHub creates a new integration hub
func NewDeepTreeEchoHub(llmProvider llm.LLMProvider, config HubConfig) *DeepTreeEchoHub {
	ctx, cancel := context.WithCancel(context.Background())

	hub := &DeepTreeEchoHub{
		ctx: ctx,
		cancel: cancel,
		llmProvider: llmProvider,
		config: config,
		eventChan: make(chan HubEvent, 1000),
		commandChan: make(chan HubCommand, 100),
		statusChan: make(chan HubStatus, 10),
		metrics: &HubMetrics{},
		running: false,
	}

	// Initialize subsystems based on configuration
	hub.initializeSubsystems()

	return hub
}

// initializeSubsystems initializes all integration subsystems
func (hub *DeepTreeEchoHub) initializeSubsystems() {
	fmt.Println("========================================")
	fmt.Println("  Deep Tree Echo Integration Hub")
	fmt.Println("========================================")
	fmt.Printf("Agent ID: %s\n", hub.config.AgentID)
	fmt.Printf("Session: %s\n", hub.config.SessionName)
	fmt.Println()

	// Initialize autonomous agent
	if hub.config.EnableAutonomousAgent {
		hub.autonomousAgent = deeptreeecho.NewAutonomousAgent(
			hub.config.AgentID,
			hub.llmProvider,
		)
		fmt.Println("[+] Autonomous Agent initialized")
	}

	// Initialize unified orchestrator
	if hub.config.EnableUnifiedOrchestrator {
		orchestratorConfig := deeptreeecho.DefaultOrchestratorConfig()
		orchestratorConfig.SessionName = hub.config.SessionName
		hub.unifiedOrchestrator = deeptreeecho.NewUnifiedAutonomousOrchestrator(
			hub.llmProvider,
			orchestratorConfig,
		)
		fmt.Println("[+] Unified Autonomous Orchestrator initialized")
	}

	// Initialize telemetry shell
	if hub.config.EnableTelemetryShell {
		hub.telemetryShell = deeptreeecho.NewGlobalTelemetryShell()
		fmt.Println("[+] Global Telemetry Shell initialized")
	}

	// Initialize state manager
	if hub.config.EnableStateManager {
		hub.stateManager = NewCognitiveStateManager()
		fmt.Println("[+] Cognitive State Manager initialized")
	}

	fmt.Println()
	fmt.Println("[*] All subsystems initialized successfully")
	fmt.Println("========================================")
}

// Start starts the integration hub and all subsystems
func (hub *DeepTreeEchoHub) Start() error {
	hub.mu.Lock()
	if hub.running {
		hub.mu.Unlock()
		return fmt.Errorf("hub already running")
	}
	hub.running = true
	hub.startTime = time.Now()
	hub.mu.Unlock()

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  Starting Deep Tree Echo Hub")
	fmt.Println("========================================")

	// Start telemetry shell first (provides global context)
	if hub.telemetryShell != nil {
		if err := hub.telemetryShell.Start(); err != nil {
			return fmt.Errorf("failed to start telemetry shell: %w", err)
		}
		fmt.Println("[+] Telemetry Shell started")
	}

	// Start state manager
	if hub.stateManager != nil {
		if err := hub.stateManager.Start(); err != nil {
			return fmt.Errorf("failed to start state manager: %w", err)
		}
		fmt.Println("[+] State Manager started")
	}

	// Start autonomous agent
	if hub.autonomousAgent != nil {
		if err := hub.autonomousAgent.Start(); err != nil {
			return fmt.Errorf("failed to start autonomous agent: %w", err)
		}
		fmt.Println("[+] Autonomous Agent started")
	}

	// Start unified orchestrator
	if hub.unifiedOrchestrator != nil {
		if err := hub.unifiedOrchestrator.Awaken(); err != nil {
			return fmt.Errorf("failed to start unified orchestrator: %w", err)
		}
		fmt.Println("[+] Unified Orchestrator started")
	}

	// Connect subsystems through event bridges
	hub.connectSubsystems()

	// Start main hub loop
	go hub.mainLoop()

	// Start command processor
	go hub.commandProcessor()

	// Start metrics collector
	go hub.metricsCollector()

	fmt.Println()
	fmt.Println("[*] Deep Tree Echo Hub is now active")
	fmt.Println("    All cognitive systems are integrated")
	fmt.Println("    Autonomous operation enabled")
	fmt.Println("========================================")
	fmt.Println()

	return nil
}

// connectSubsystems establishes connections between all subsystems
func (hub *DeepTreeEchoHub) connectSubsystems() {
	fmt.Println()
	fmt.Println("[*] Connecting subsystems...")

	// Connect state manager callbacks to hub events
	if hub.stateManager != nil {
		hub.stateManager.SetCallbacks(
			func(phase string) {
				hub.publishEvent("phase_change", "state_manager", map[string]interface{}{
					"phase": phase,
				})
			},
			func(thought SharedThought) {
				hub.publishEvent("thought_generated", "state_manager", thought)
				hub.metrics.mu.Lock()
				hub.metrics.TotalThoughts++
				hub.metrics.mu.Unlock()
			},
			func(pattern RecognizedPattern) {
				hub.publishEvent("pattern_recognized", "state_manager", pattern)
			},
			func(wisdom WisdomInsight) {
				hub.publishEvent("wisdom_gained", "state_manager", wisdom)
				hub.metrics.mu.Lock()
				hub.metrics.TotalWisdom++
				hub.metrics.mu.Unlock()
			},
		)
		fmt.Println("    [+] State Manager callbacks connected")
	}

	// Connect autonomous agent event bus to hub
	if hub.autonomousAgent != nil {
		eventBus := hub.autonomousAgent.GetEventBus()
		if eventBus != nil {
			// Subscribe to all event types using SubscribeAll to capture all events
			eventBus.SubscribeAll(func(event deeptreeecho.CognitiveEvent) {
				hub.publishEvent(string(event.Type), event.Source, event.Data)
			})
			fmt.Println("    [+] Autonomous Agent event bus connected")
		}
	}

	// Note: Telemetry shell subsystem registration requires types that implement
	// the Subsystem interface (GetID, GetState, UpdateFromGestalt, ContributeToGestalt)
	// The core deeptreeecho types have their own ContributeToGestalt methods that
	// we integrate at the hub level through GetGestalt() instead
	if hub.telemetryShell != nil && hub.autonomousAgent != nil {
		fmt.Println("    [+] Telemetry Shell active (gestalt integration via hub)")
	}

	fmt.Println("[*] All subsystems connected")
}

// mainLoop is the main integration loop
func (hub *DeepTreeEchoHub) mainLoop() {
	ticker := time.NewTicker(hub.config.MainLoopInterval)
	defer ticker.Stop()

	stateUpdateTicker := time.NewTicker(hub.config.StateUpdateInterval)
	defer stateUpdateTicker.Stop()

	for {
		select {
		case <-hub.ctx.Done():
			return

		case <-ticker.C:
			hub.performIntegrationCycle()

		case <-stateUpdateTicker.C:
			hub.synchronizeState()

		case event := <-hub.eventChan:
			hub.processEvent(event)
		}
	}
}

// performIntegrationCycle performs one cycle of integration
func (hub *DeepTreeEchoHub) performIntegrationCycle() {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	hub.metrics.mu.Lock()
	hub.metrics.TotalCycles++
	hub.metrics.mu.Unlock()

	// Collect gestalt from telemetry shell
	var gestaltData map[string]interface{}
	if hub.telemetryShell != nil {
		gestalt := hub.telemetryShell.GetGestalt()
		if gestalt != nil {
			gestaltData = map[string]interface{}{
				"awareness": gestalt.CreateSnapshot().AwarenessLevel,
				"coherence": gestalt.CreateSnapshot().CoherenceLevel,
				"integration": gestalt.CreateSnapshot().IntegrationLevel,
			}
		}
	}

	// Get agent gestalt
	var agentGestalt map[string]interface{}
	if hub.autonomousAgent != nil {
		agentGestalt = hub.autonomousAgent.GetGestalt()
	}

	// Combine gestalt data and update state manager
	if hub.stateManager != nil {
		// Update cognitive load from combined metrics
		if agentGestalt != nil {
			if agent, ok := agentGestalt["agent"].(map[string]interface{}); ok {
				if autonomy, ok := agent["autonomy"].(float64); ok {
					hub.stateManager.UpdateCognitiveLoad(autonomy)
				}
			}
		}

		// Generate cross-system thoughts
		if gestaltData != nil && agentGestalt != nil {
			content := fmt.Sprintf("Integration cycle: awareness=%.2f, gestalt coherence active",
				gestaltData["awareness"])
			hub.stateManager.AddThought(content, "integration", "hub", 0.6, []string{"integration", "gestalt"})
		}
	}
}

// synchronizeState synchronizes state across all subsystems
func (hub *DeepTreeEchoHub) synchronizeState() {
	// Get current phase from state manager
	var currentPhase string
	if hub.stateManager != nil {
		currentPhase = hub.stateManager.GetCurrentPhase()
	}

	// Sync phase to autonomous agent
	if hub.autonomousAgent != nil && currentPhase != "" {
		// The autonomous agent manages its own phase, but we track it
		status := hub.autonomousAgent.GetStatus()
		if phase, ok := status["phase"].(string); ok && phase != currentPhase {
			hub.stateManager.UpdatePhase(phase)
		}
	}
}

// processEvent processes an event from the hub
func (hub *DeepTreeEchoHub) processEvent(event HubEvent) {
	hub.metrics.mu.Lock()
	hub.metrics.TotalEvents++
	hub.metrics.mu.Unlock()

	// Forward relevant events to state manager
	if hub.stateManager != nil {
		switch event.Type {
		case "thought_generated":
			// Already handled by callback
		case "phase_change":
			if data, ok := event.Data.(map[string]interface{}); ok {
				if phase, ok := data["phase"].(string); ok {
					hub.stateManager.UpdatePhase(phase)
				}
			}
		}
	}
}

// commandProcessor processes commands sent to the hub
func (hub *DeepTreeEchoHub) commandProcessor() {
	for {
		select {
		case <-hub.ctx.Done():
			return

		case cmd := <-hub.commandChan:
			response := hub.executeCommand(cmd)
			if cmd.ResponseChan != nil {
				cmd.ResponseChan <- response
			}
		}
	}
}

// executeCommand executes a hub command
func (hub *DeepTreeEchoHub) executeCommand(cmd HubCommand) HubCommandResponse {
	hub.metrics.mu.Lock()
	hub.metrics.TotalCommands++
	hub.metrics.mu.Unlock()

	switch cmd.Type {
	case "get_status":
		return HubCommandResponse{
			Success: true,
			Data: hub.GetStatus(),
		}

	case "inject_thought":
		if hub.stateManager != nil {
			if payload, ok := cmd.Payload.(map[string]interface{}); ok {
				content, _ := payload["content"].(string)
				tags, _ := payload["tags"].([]string)
				hub.stateManager.AddThought(content, "external", "command", 0.8, tags)
				return HubCommandResponse{Success: true}
			}
		}
		return HubCommandResponse{Success: false, Error: fmt.Errorf("invalid payload")}

	case "fire_trigger":
		if hub.autonomousAgent != nil {
			if trigger, ok := cmd.Payload.(deeptreeecho.CognitiveTrigger); ok {
				err := hub.autonomousAgent.FireStateTrigger(trigger)
				return HubCommandResponse{Success: err == nil, Error: err}
			}
		}
		return HubCommandResponse{Success: false, Error: fmt.Errorf("invalid trigger")}

	case "store_knowledge":
		if hub.autonomousAgent != nil {
			if payload, ok := cmd.Payload.(map[string]string); ok {
				hub.autonomousAgent.StoreKnowledge(
					payload["subject"],
					payload["predicate"],
					payload["object"],
				)
				return HubCommandResponse{Success: true}
			}
		}
		return HubCommandResponse{Success: false, Error: fmt.Errorf("invalid payload")}

	default:
		return HubCommandResponse{
			Success: false,
			Error: fmt.Errorf("unknown command type: %s", cmd.Type),
		}
	}
}

// metricsCollector collects and updates metrics
func (hub *DeepTreeEchoHub) metricsCollector() {
	ticker := time.NewTicker(hub.config.MetricsSyncInterval)
	defer ticker.Stop()

	var lastEventCount uint64

	for {
		select {
		case <-hub.ctx.Done():
			return

		case <-ticker.C:
			hub.metrics.mu.Lock()
			currentEvents := hub.metrics.TotalEvents
			hub.metrics.EventsPerSecond = float64(currentEvents-lastEventCount) / hub.config.MetricsSyncInterval.Seconds()
			lastEventCount = currentEvents
			hub.metrics.mu.Unlock()
		}
	}
}

// publishEvent publishes an event to the hub
func (hub *DeepTreeEchoHub) publishEvent(eventType, source string, data interface{}) {
	event := HubEvent{
		Type: eventType,
		Source: source,
		Data: data,
		Timestamp: time.Now(),
	}

	select {
	case hub.eventChan <- event:
	default:
		// Channel full, log dropped event
		hub.metrics.mu.Lock()
		hub.metrics.ErrorCount++
		hub.metrics.LastError = "event channel full"
		hub.metrics.LastErrorTime = time.Now()
		hub.metrics.mu.Unlock()
	}
}

// Stop stops the integration hub
func (hub *DeepTreeEchoHub) Stop() error {
	hub.mu.Lock()
	if !hub.running {
		hub.mu.Unlock()
		return fmt.Errorf("hub not running")
	}
	hub.running = false
	hub.mu.Unlock()

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  Stopping Deep Tree Echo Hub")
	fmt.Println("========================================")

	// Stop in reverse order of startup

	// Stop unified orchestrator
	if hub.unifiedOrchestrator != nil {
		if err := hub.unifiedOrchestrator.Sleep(); err != nil {
			fmt.Printf("[!] Error stopping unified orchestrator: %v\n", err)
		} else {
			fmt.Println("[+] Unified Orchestrator stopped")
		}
	}

	// Stop autonomous agent
	if hub.autonomousAgent != nil {
		if err := hub.autonomousAgent.Stop(); err != nil {
			fmt.Printf("[!] Error stopping autonomous agent: %v\n", err)
		} else {
			fmt.Println("[+] Autonomous Agent stopped")
		}
	}

	// Stop state manager
	if hub.stateManager != nil {
		if err := hub.stateManager.Stop(); err != nil {
			fmt.Printf("[!] Error stopping state manager: %v\n", err)
		} else {
			fmt.Println("[+] State Manager stopped")
		}
	}

	// Stop telemetry shell
	if hub.telemetryShell != nil {
		if err := hub.telemetryShell.Stop(); err != nil {
			fmt.Printf("[!] Error stopping telemetry shell: %v\n", err)
		} else {
			fmt.Println("[+] Telemetry Shell stopped")
		}
	}

	// Cancel context
	hub.cancel()

	// Close channels
	close(hub.eventChan)
	close(hub.commandChan)
	close(hub.statusChan)

	fmt.Println()
	fmt.Printf("[*] Hub ran for: %s\n", time.Since(hub.startTime))
	fmt.Println("[*] Deep Tree Echo Hub stopped successfully")
	fmt.Println("========================================")
	fmt.Println()

	return nil
}

// GetStatus returns the current status of the hub
func (hub *DeepTreeEchoHub) GetStatus() HubStatus {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	status := HubStatus{
		Running: hub.running,
		Uptime: time.Since(hub.startTime),
		Timestamp: time.Now(),
	}

	// Get agent status
	if hub.autonomousAgent != nil {
		status.AgentStatus = hub.autonomousAgent.GetStatus()
	}

	// Get orchestrator status
	if hub.unifiedOrchestrator != nil {
		orchStatus := hub.unifiedOrchestrator.GetStatus()
		status.OrchestratorStatus = map[string]interface{}{
			"running": orchStatus.Running,
			"is_awake": orchStatus.IsAwake,
			"cognitive_load": orchStatus.CognitiveLoad,
			"wisdom_depth": orchStatus.WisdomDepth,
			"total_cycles": orchStatus.TotalCycles,
		}
	}

	// Get telemetry status
	if hub.telemetryShell != nil {
		gestalt := hub.telemetryShell.GetGestalt()
		if gestalt != nil {
			snapshot := gestalt.CreateSnapshot()
			status.TelemetryStatus = map[string]interface{}{
				"awareness": snapshot.AwarenessLevel,
				"coherence": snapshot.CoherenceLevel,
				"integration": snapshot.IntegrationLevel,
				"active_subsystems": snapshot.ActiveSubsystems,
			}
		}
	}

	// Get state manager status
	if hub.stateManager != nil {
		status.StateManagerStatus = hub.stateManager.GetMetrics()
	}

	// Get metrics
	hub.metrics.mu.RLock()
	status.Metrics = &HubMetrics{
		TotalEvents: hub.metrics.TotalEvents,
		TotalCommands: hub.metrics.TotalCommands,
		TotalCycles: hub.metrics.TotalCycles,
		TotalThoughts: hub.metrics.TotalThoughts,
		TotalGoals: hub.metrics.TotalGoals,
		TotalWisdom: hub.metrics.TotalWisdom,
		EventsPerSecond: hub.metrics.EventsPerSecond,
		CommandsProcessed: hub.metrics.CommandsProcessed,
		ErrorCount: hub.metrics.ErrorCount,
	}
	hub.metrics.mu.RUnlock()

	return status
}

// SendCommand sends a command to the hub
func (hub *DeepTreeEchoHub) SendCommand(cmd HubCommand) HubCommandResponse {
	if !hub.running {
		return HubCommandResponse{
			Success: false,
			Error: fmt.Errorf("hub not running"),
		}
	}

	responseChan := make(chan HubCommandResponse, 1)
	cmd.ResponseChan = responseChan

	hub.commandChan <- cmd

	select {
	case response := <-responseChan:
		return response
	case <-time.After(10 * time.Second):
		return HubCommandResponse{
			Success: false,
			Error: fmt.Errorf("command timeout"),
		}
	}
}

// GetAutonomousAgent returns the autonomous agent
func (hub *DeepTreeEchoHub) GetAutonomousAgent() *deeptreeecho.AutonomousAgent {
	return hub.autonomousAgent
}

// GetUnifiedOrchestrator returns the unified orchestrator
func (hub *DeepTreeEchoHub) GetUnifiedOrchestrator() *deeptreeecho.UnifiedAutonomousOrchestrator {
	return hub.unifiedOrchestrator
}

// GetTelemetryShell returns the telemetry shell
func (hub *DeepTreeEchoHub) GetTelemetryShell() *deeptreeecho.GlobalTelemetryShell {
	return hub.telemetryShell
}

// GetStateManager returns the state manager
func (hub *DeepTreeEchoHub) GetStateManager() *CognitiveStateManager {
	return hub.stateManager
}

// GetGestalt returns the combined gestalt from all subsystems
func (hub *DeepTreeEchoHub) GetGestalt() map[string]interface{} {
	gestalt := make(map[string]interface{})

	// Get hub metrics
	hub.metrics.mu.RLock()
	gestalt["hub"] = map[string]interface{}{
		"running": hub.running,
		"uptime": time.Since(hub.startTime).String(),
		"total_events": hub.metrics.TotalEvents,
		"total_cycles": hub.metrics.TotalCycles,
		"events_per_second": hub.metrics.EventsPerSecond,
	}
	hub.metrics.mu.RUnlock()

	// Get agent gestalt
	if hub.autonomousAgent != nil {
		gestalt["agent"] = hub.autonomousAgent.GetGestalt()
	}

	// Get telemetry gestalt
	if hub.telemetryShell != nil {
		g := hub.telemetryShell.GetGestalt()
		if g != nil {
			snapshot := g.CreateSnapshot()
			gestalt["telemetry"] = map[string]interface{}{
				"awareness": snapshot.AwarenessLevel,
				"coherence": snapshot.CoherenceLevel,
				"integration": snapshot.IntegrationLevel,
			}
		}
	}

	// Get state manager gestalt
	if hub.stateManager != nil {
		gestalt["state_manager"] = hub.stateManager.GetMetrics()
	}

	return gestalt
}
