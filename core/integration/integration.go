// Package integration provides unified integration for the Deep Tree Echo cognitive architecture.
// It connects all subsystems including:
// - Core autonomous agent systems (deeptreeecho)
// - Orchestration layer for high-level coordination
// - Cognitive state management
// - Memory and consciousness integration
// - Entelechy and ontogenesis development tracking
//
// The integration package serves as the glue layer that enables seamless communication
// and state synchronization between all cognitive components of the Deep Tree Echo system.
package integration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
	"github.com/cogpy/echo9llama/orchestration"
)

// IntegratedDeepTreeEcho represents the fully integrated Deep Tree Echo system
// with all layers connected and synchronized.
type IntegratedDeepTreeEcho struct {
	mu sync.RWMutex
	ctx context.Context
	cancel context.CancelFunc

	// Integration hub - central coordinator
	hub *DeepTreeEchoHub

	// Cognitive bridge - connects core and orchestration
	bridge *CognitiveBridge

	// Orchestration layer
	orchestration *orchestration.DeepTreeEcho

	// Configuration
	config IntegrationConfig

	// Lifecycle state
	initialized bool
	started bool
	startTime time.Time

	// Lifecycle hooks
	onStartHooks []func() error
	onStopHooks []func() error
}

// IntegrationConfig holds configuration for the integrated system
type IntegrationConfig struct {
	// Identity
	SystemName string
	AgentID string
	SessionName string

	// LLM Configuration
	LLMProvider llm.LLMProvider

	// Hub Configuration
	HubConfig HubConfig

	// Bridge Configuration
	BridgeConfig BridgeConfig

	// Features
	EnableOrchestration bool
	EnableBridge bool
	EnableAutoRecovery bool

	// Timing
	StartupTimeout time.Duration
	ShutdownTimeout time.Duration
}

// DefaultIntegrationConfig returns the default integration configuration
func DefaultIntegrationConfig(llmProvider llm.LLMProvider) IntegrationConfig {
	ts := time.Now().Unix()
	return IntegrationConfig{
		SystemName: "DeepTreeEcho",
		AgentID: fmt.Sprintf("dte-%d", ts),
		SessionName: fmt.Sprintf("dte-session-%d", ts),
		LLMProvider: llmProvider,
		HubConfig: DefaultHubConfig(),
		BridgeConfig: DefaultBridgeConfig(),
		EnableOrchestration: true,
		EnableBridge: true,
		EnableAutoRecovery: true,
		StartupTimeout: 60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
	}
}

// NewIntegratedDeepTreeEcho creates a new fully integrated Deep Tree Echo system
func NewIntegratedDeepTreeEcho(config IntegrationConfig) (*IntegratedDeepTreeEcho, error) {
	if config.LLMProvider == nil {
		return nil, fmt.Errorf("LLM provider is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Configure hub
	hubConfig := config.HubConfig
	hubConfig.AgentID = config.AgentID
	hubConfig.SessionName = config.SessionName

	integrated := &IntegratedDeepTreeEcho{
		ctx: ctx,
		cancel: cancel,
		config: config,
		onStartHooks: make([]func() error, 0),
		onStopHooks: make([]func() error, 0),
	}

	// Initialize hub
	integrated.hub = NewDeepTreeEchoHub(config.LLMProvider, hubConfig)

	// Initialize orchestration layer if enabled
	if config.EnableOrchestration {
		integrated.orchestration = orchestration.NewDeepTreeEcho(config.SystemName)
	}

	// Initialize bridge if enabled
	if config.EnableBridge && integrated.hub.GetAutonomousAgent() != nil && integrated.orchestration != nil {
		integrated.bridge = NewCognitiveBridge(
			integrated.hub.GetAutonomousAgent(),
			integrated.orchestration,
			config.BridgeConfig,
		)
	}

	integrated.initialized = true

	return integrated, nil
}

// Start starts the integrated system
func (i *IntegratedDeepTreeEcho) Start(ctx context.Context) error {
	i.mu.Lock()
	if i.started {
		i.mu.Unlock()
		return fmt.Errorf("system already started")
	}
	i.mu.Unlock()

	fmt.Println()
	fmt.Println("============================================================")
	fmt.Println("      Deep Tree Echo - Integrated Cognitive System")
	fmt.Println("============================================================")
	fmt.Printf("  System: %s\n", i.config.SystemName)
	fmt.Printf("  Agent: %s\n", i.config.AgentID)
	fmt.Printf("  Session: %s\n", i.config.SessionName)
	fmt.Println("============================================================")
	fmt.Println()

	// Create startup context with timeout
	startCtx, startCancel := context.WithTimeout(ctx, i.config.StartupTimeout)
	defer startCancel()

	// Run pre-start hooks
	for _, hook := range i.onStartHooks {
		if err := hook(); err != nil {
			return fmt.Errorf("pre-start hook failed: %w", err)
		}
	}

	// Start orchestration layer first
	if i.orchestration != nil {
		fmt.Println("[1/4] Initializing orchestration layer...")
		if err := i.orchestration.InitializeDTECore(startCtx); err != nil {
			return fmt.Errorf("failed to initialize orchestration: %w", err)
		}
		fmt.Println("      Orchestration layer initialized")
	}

	// Start the hub
	fmt.Println("[2/4] Starting integration hub...")
	if err := i.hub.Start(); err != nil {
		return fmt.Errorf("failed to start hub: %w", err)
	}
	fmt.Println("      Integration hub started")

	// Start the bridge
	if i.bridge != nil {
		fmt.Println("[3/4] Starting cognitive bridge...")
		if err := i.bridge.Start(); err != nil {
			return fmt.Errorf("failed to start bridge: %w", err)
		}
		fmt.Println("      Cognitive bridge started")
	} else {
		fmt.Println("[3/4] Cognitive bridge disabled")
	}

	// Verify all systems
	fmt.Println("[4/4] Verifying system integration...")
	if err := i.verifyIntegration(); err != nil {
		return fmt.Errorf("integration verification failed: %w", err)
	}
	fmt.Println("      System integration verified")

	i.mu.Lock()
	i.started = true
	i.startTime = time.Now()
	i.mu.Unlock()

	fmt.Println()
	fmt.Println("============================================================")
	fmt.Println("      Deep Tree Echo is now fully operational")
	fmt.Println("============================================================")
	fmt.Println("  All cognitive systems integrated and synchronized")
	fmt.Println("  Autonomous operation enabled")
	fmt.Println("  Wisdom cultivation active")
	fmt.Println("============================================================")
	fmt.Println()

	return nil
}

// verifyIntegration verifies that all systems are properly integrated
func (i *IntegratedDeepTreeEcho) verifyIntegration() error {
	// Verify hub is running
	status := i.hub.GetStatus()
	if !status.Running {
		return fmt.Errorf("hub is not running")
	}

	// Verify autonomous agent
	if i.hub.GetAutonomousAgent() == nil {
		return fmt.Errorf("autonomous agent not initialized")
	}

	// Verify orchestration if enabled
	if i.orchestration != nil {
		if i.orchestration.CoreStatus != orchestration.CoreStatusActive {
			// Try to activate
			if err := i.orchestration.RefreshStatus(i.ctx); err != nil {
				return fmt.Errorf("failed to refresh orchestration status: %w", err)
			}
		}
	}

	// Verify bridge if enabled
	if i.bridge != nil && !i.bridge.IsRunning() {
		return fmt.Errorf("cognitive bridge not running")
	}

	return nil
}

// Stop stops the integrated system
func (i *IntegratedDeepTreeEcho) Stop() error {
	i.mu.Lock()
	if !i.started {
		i.mu.Unlock()
		return fmt.Errorf("system not started")
	}
	i.mu.Unlock()

	fmt.Println()
	fmt.Println("============================================================")
	fmt.Println("      Stopping Deep Tree Echo")
	fmt.Println("============================================================")

	// Create shutdown context with timeout
	stopCtx, stopCancel := context.WithTimeout(context.Background(), i.config.ShutdownTimeout)
	defer stopCancel()

	// Run pre-stop hooks
	for _, hook := range i.onStopHooks {
		if err := hook(); err != nil {
			fmt.Printf("[!] Pre-stop hook error: %v\n", err)
		}
	}

	// Stop in reverse order

	// Stop bridge
	if i.bridge != nil {
		fmt.Println("[1/3] Stopping cognitive bridge...")
		if err := i.bridge.Stop(); err != nil {
			fmt.Printf("      Warning: %v\n", err)
		} else {
			fmt.Println("      Bridge stopped")
		}
	}

	// Stop hub
	fmt.Println("[2/3] Stopping integration hub...")
	if err := i.hub.Stop(); err != nil {
		fmt.Printf("      Warning: %v\n", err)
	} else {
		fmt.Println("      Hub stopped")
	}

	// Run final diagnostics on orchestration
	if i.orchestration != nil {
		fmt.Println("[3/3] Running final diagnostics...")
		if _, err := i.orchestration.RunDiagnostics(stopCtx); err != nil {
			fmt.Printf("      Warning: %v\n", err)
		} else {
			fmt.Println("      Diagnostics complete")
		}
	}

	// Cancel context
	i.cancel()

	i.mu.Lock()
	uptime := time.Since(i.startTime)
	i.started = false
	i.mu.Unlock()

	fmt.Println()
	fmt.Printf("  Session duration: %s\n", uptime)
	fmt.Println("============================================================")
	fmt.Println("      Deep Tree Echo stopped successfully")
	fmt.Println("============================================================")
	fmt.Println()

	return nil
}

// GetStatus returns the current status of the integrated system
func (i *IntegratedDeepTreeEcho) GetStatus() map[string]interface{} {
	i.mu.RLock()
	defer i.mu.RUnlock()

	status := map[string]interface{}{
		"initialized": i.initialized,
		"started": i.started,
		"system_name": i.config.SystemName,
		"agent_id": i.config.AgentID,
		"session_name": i.config.SessionName,
	}

	if i.started {
		status["uptime"] = time.Since(i.startTime).String()
	}

	// Get hub status
	if i.hub != nil {
		hubStatus := i.hub.GetStatus()
		status["hub"] = map[string]interface{}{
			"running": hubStatus.Running,
			"total_events": hubStatus.Metrics.TotalEvents,
			"total_cycles": hubStatus.Metrics.TotalCycles,
		}
	}

	// Get orchestration status
	if i.orchestration != nil {
		status["orchestration"] = map[string]interface{}{
			"system_health": string(i.orchestration.SystemHealth),
			"core_status": string(i.orchestration.CoreStatus),
			"thought_count": i.orchestration.ThoughtCount,
			"evolution_stage": func() string {
				if i.orchestration.EvolutionTimeline != nil {
					return i.orchestration.EvolutionTimeline.CurrentStage
				}
				return "unknown"
			}(),
		}
	}

	// Get bridge status
	if i.bridge != nil {
		status["bridge"] = i.bridge.GetMetrics()
	}

	return status
}

// GetGestalt returns the combined gestalt from all subsystems
func (i *IntegratedDeepTreeEcho) GetGestalt() map[string]interface{} {
	gestalt := make(map[string]interface{})

	// Get hub gestalt
	if i.hub != nil {
		gestalt["hub"] = i.hub.GetGestalt()
	}

	// Get orchestration gestalt
	if i.orchestration != nil {
		gestalt["orchestration"] = map[string]interface{}{
			"identity_coherence": func() float64 {
				if i.orchestration.IdentityCoherence != nil {
					return i.orchestration.IdentityCoherence.OverallCoherence
				}
				return 0
			}(),
			"memory_coherence": func() float64 {
				if i.orchestration.MemoryResonance != nil {
					return i.orchestration.MemoryResonance.Coherence
				}
				return 0
			}(),
			"echo_patterns": func() map[string]float64 {
				patterns := make(map[string]float64)
				if i.orchestration.EchoPatterns != nil {
					if i.orchestration.EchoPatterns.RecursiveSelfImprovement != nil {
						patterns["recursive_self_improvement"] = i.orchestration.EchoPatterns.RecursiveSelfImprovement.Strength
					}
					if i.orchestration.EchoPatterns.CrossSystemSynthesis != nil {
						patterns["cross_system_synthesis"] = i.orchestration.EchoPatterns.CrossSystemSynthesis.Strength
					}
					if i.orchestration.EchoPatterns.IdentityPreservation != nil {
						patterns["identity_preservation"] = i.orchestration.EchoPatterns.IdentityPreservation.Strength
					}
				}
				return patterns
			}(),
		}
	}

	return gestalt
}

// OnStart registers a hook to be called before system starts
func (i *IntegratedDeepTreeEcho) OnStart(hook func() error) {
	i.onStartHooks = append(i.onStartHooks, hook)
}

// OnStop registers a hook to be called before system stops
func (i *IntegratedDeepTreeEcho) OnStop(hook func() error) {
	i.onStopHooks = append(i.onStopHooks, hook)
}

// GetHub returns the integration hub
func (i *IntegratedDeepTreeEcho) GetHub() *DeepTreeEchoHub {
	return i.hub
}

// GetBridge returns the cognitive bridge
func (i *IntegratedDeepTreeEcho) GetBridge() *CognitiveBridge {
	return i.bridge
}

// GetOrchestration returns the orchestration layer
func (i *IntegratedDeepTreeEcho) GetOrchestration() *orchestration.DeepTreeEcho {
	return i.orchestration
}

// GetAutonomousAgent returns the autonomous agent from the hub
func (i *IntegratedDeepTreeEcho) GetAutonomousAgent() *deeptreeecho.AutonomousAgent {
	if i.hub != nil {
		return i.hub.GetAutonomousAgent()
	}
	return nil
}

// InjectThought injects a thought into the cognitive system
func (i *IntegratedDeepTreeEcho) InjectThought(content string, tags []string) error {
	if !i.started {
		return fmt.Errorf("system not started")
	}

	stateManager := i.hub.GetStateManager()
	if stateManager != nil {
		stateManager.AddThought(content, "external", "api", 0.8, tags)
		return nil
	}

	return fmt.Errorf("state manager not available")
}

// StoreKnowledge stores knowledge in the knowledge graph
func (i *IntegratedDeepTreeEcho) StoreKnowledge(subject, predicate, object string) error {
	if !i.started {
		return fmt.Errorf("system not started")
	}

	agent := i.GetAutonomousAgent()
	if agent != nil {
		agent.StoreKnowledge(subject, predicate, object)
		return nil
	}

	return fmt.Errorf("autonomous agent not available")
}

// QueryKnowledge queries the knowledge graph
func (i *IntegratedDeepTreeEcho) QueryKnowledge(subject string) []*deeptreeecho.Quad {
	if !i.started {
		return nil
	}

	agent := i.GetAutonomousAgent()
	if agent != nil {
		return agent.QueryKnowledge(subject)
	}

	return nil
}
