package integration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/orchestration"
)

// CognitiveBridge connects the core deeptreeecho systems with the orchestration layer,
// enabling bidirectional communication and state synchronization between the
// autonomous agent systems and the high-level orchestration components.
type CognitiveBridge struct {
	mu sync.RWMutex
	ctx context.Context
	cancel context.CancelFunc

	// Core layer reference
	autonomousAgent *deeptreeecho.AutonomousAgent

	// Orchestration layer reference
	deepTreeEcho *orchestration.DeepTreeEcho

	// Synchronization channels
	coreToOrchChan chan BridgeMessage
	orchToCoreChan chan BridgeMessage

	// State synchronization
	lastCoreSync time.Time
	lastOrchSync time.Time
	syncInterval time.Duration

	// Configuration
	config BridgeConfig

	// Metrics
	messagesSent uint64
	messagesReceived uint64
	syncCount uint64
	errorCount uint64

	// Running state
	running bool
}

// BridgeConfig holds configuration for the cognitive bridge
type BridgeConfig struct {
	SyncInterval time.Duration
	BufferSize int
	EnableBidirectional bool
	EnableAutoSync bool
}

// DefaultBridgeConfig returns default bridge configuration
func DefaultBridgeConfig() BridgeConfig {
	return BridgeConfig{
		SyncInterval: 5 * time.Second,
		BufferSize: 100,
		EnableBidirectional: true,
		EnableAutoSync: true,
	}
}

// BridgeMessage represents a message passing through the bridge
type BridgeMessage struct {
	Type string
	Source string
	Destination string
	Payload interface{}
	Timestamp time.Time
	Priority int
}

// BridgeMessageType defines message types for the bridge
type BridgeMessageType string

const (
	MessageTypeStateSync BridgeMessageType = "state_sync"
	MessageTypeEvent BridgeMessageType = "event"
	MessageTypeCommand BridgeMessageType = "command"
	MessageTypeMetrics BridgeMessageType = "metrics"
	MessageTypeThought BridgeMessageType = "thought"
	MessageTypeWisdom BridgeMessageType = "wisdom"
	MessageTypePattern BridgeMessageType = "pattern"
)

// NewCognitiveBridge creates a new cognitive bridge
func NewCognitiveBridge(
	agent *deeptreeecho.AutonomousAgent,
	dte *orchestration.DeepTreeEcho,
	config BridgeConfig,
) *CognitiveBridge {
	ctx, cancel := context.WithCancel(context.Background())

	return &CognitiveBridge{
		ctx: ctx,
		cancel: cancel,
		autonomousAgent: agent,
		deepTreeEcho: dte,
		coreToOrchChan: make(chan BridgeMessage, config.BufferSize),
		orchToCoreChan: make(chan BridgeMessage, config.BufferSize),
		syncInterval: config.SyncInterval,
		config: config,
	}
}

// Start starts the cognitive bridge
func (cb *CognitiveBridge) Start() error {
	cb.mu.Lock()
	if cb.running {
		cb.mu.Unlock()
		return fmt.Errorf("bridge already running")
	}
	cb.running = true
	cb.mu.Unlock()

	fmt.Println("[CognitiveBridge] Starting cognitive bridge...")

	// Start core to orchestration relay
	go cb.coreToOrchRelay()

	// Start orchestration to core relay
	if cb.config.EnableBidirectional {
		go cb.orchToCoreRelay()
	}

	// Start auto-sync if enabled
	if cb.config.EnableAutoSync {
		go cb.autoSyncLoop()
	}

	fmt.Println("[CognitiveBridge] Bridge started successfully")
	return nil
}

// Stop stops the cognitive bridge
func (cb *CognitiveBridge) Stop() error {
	cb.mu.Lock()
	if !cb.running {
		cb.mu.Unlock()
		return fmt.Errorf("bridge not running")
	}
	cb.running = false
	cb.mu.Unlock()

	cb.cancel()
	close(cb.coreToOrchChan)
	close(cb.orchToCoreChan)

	fmt.Println("[CognitiveBridge] Bridge stopped")
	return nil
}

// coreToOrchRelay relays messages from core to orchestration layer
func (cb *CognitiveBridge) coreToOrchRelay() {
	for {
		select {
		case <-cb.ctx.Done():
			return
		case msg := <-cb.coreToOrchChan:
			cb.processCoreMessage(msg)
		}
	}
}

// orchToCoreRelay relays messages from orchestration to core layer
func (cb *CognitiveBridge) orchToCoreRelay() {
	for {
		select {
		case <-cb.ctx.Done():
			return
		case msg := <-cb.orchToCoreChan:
			cb.processOrchMessage(msg)
		}
	}
}

// autoSyncLoop performs automatic state synchronization
func (cb *CognitiveBridge) autoSyncLoop() {
	ticker := time.NewTicker(cb.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cb.ctx.Done():
			return
		case <-ticker.C:
			cb.synchronizeState()
		}
	}
}

// synchronizeState synchronizes state between core and orchestration layers
func (cb *CognitiveBridge) synchronizeState() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.syncCount++

	// Sync from core to orchestration
	if cb.autonomousAgent != nil && cb.deepTreeEcho != nil {
		// Get core agent status
		coreStatus := cb.autonomousAgent.GetStatus()

		// Update orchestration layer with core status
		cb.updateOrchestrationFromCore(coreStatus)

		cb.lastCoreSync = time.Now()

		// Get orchestration status
		orchStatus := cb.getOrchestrationStatus()

		// Update core layer with orchestration status
		cb.updateCoreFromOrchestration(orchStatus)

		cb.lastOrchSync = time.Now()
	}
}

// updateOrchestrationFromCore updates orchestration layer from core agent status
func (cb *CognitiveBridge) updateOrchestrationFromCore(coreStatus map[string]interface{}) {
	if cb.deepTreeEcho == nil {
		return
	}

	// Extract phase from core status
	if phase, ok := coreStatus["phase"].(string); ok {
		// Map core phase to orchestration core status
		switch phase {
		case "Active":
			cb.deepTreeEcho.CoreStatus = orchestration.CoreStatusActive
		case "Resting", "Consolidating":
			cb.deepTreeEcho.CoreStatus = orchestration.CoreStatusInactive
		case "Bootstrapping", "Awakening":
			cb.deepTreeEcho.CoreStatus = orchestration.CoreStatusStarting
		}
	}

	// Extract metrics
	if metrics, ok := coreStatus["metrics"].(map[string]interface{}); ok {
		if thoughts, ok := metrics["total_thoughts"].(uint64); ok {
			cb.deepTreeEcho.ThoughtCount = int64(thoughts)
		}
	}

	// Extract autonomy level and update identity coherence
	if autonomy, ok := coreStatus["autonomy_level"].(float64); ok {
		if cb.deepTreeEcho.IdentityCoherence != nil {
			cb.deepTreeEcho.IdentityCoherence.OverallCoherence = autonomy
			cb.deepTreeEcho.IdentityCoherence.LastUpdated = time.Now()
		}
	}

	// Update timestamp
	cb.deepTreeEcho.UpdatedAt = time.Now()
}

// getOrchestrationStatus gets status from orchestration layer
func (cb *CognitiveBridge) getOrchestrationStatus() map[string]interface{} {
	if cb.deepTreeEcho == nil {
		return nil
	}

	return map[string]interface{}{
		"system_health": string(cb.deepTreeEcho.SystemHealth),
		"core_status": string(cb.deepTreeEcho.CoreStatus),
		"thought_count": cb.deepTreeEcho.ThoughtCount,
		"recursive_depth": cb.deepTreeEcho.RecursiveDepth,
		"identity_coherence": func() float64 {
			if cb.deepTreeEcho.IdentityCoherence != nil {
				return cb.deepTreeEcho.IdentityCoherence.OverallCoherence
			}
			return 0
		}(),
		"memory_coherence": func() float64 {
			if cb.deepTreeEcho.MemoryResonance != nil {
				return cb.deepTreeEcho.MemoryResonance.Coherence
			}
			return 0
		}(),
		"evolution_stage": func() string {
			if cb.deepTreeEcho.EvolutionTimeline != nil {
				return cb.deepTreeEcho.EvolutionTimeline.CurrentStage
			}
			return "unknown"
		}(),
	}
}

// updateCoreFromOrchestration updates core layer from orchestration status
func (cb *CognitiveBridge) updateCoreFromOrchestration(orchStatus map[string]interface{}) {
	if cb.autonomousAgent == nil || orchStatus == nil {
		return
	}

	// Store evolution stage in knowledge graph
	if stage, ok := orchStatus["evolution_stage"].(string); ok {
		cb.autonomousAgent.StoreKnowledge("DeepTreeEcho", "evolution_stage", stage)
	}

	// Store system health
	if health, ok := orchStatus["system_health"].(string); ok {
		cb.autonomousAgent.StoreKnowledge("DeepTreeEcho", "system_health", health)
	}
}

// processCoreMessage processes a message from the core layer
func (cb *CognitiveBridge) processCoreMessage(msg BridgeMessage) {
	cb.mu.Lock()
	cb.messagesReceived++
	cb.mu.Unlock()

	if cb.deepTreeEcho == nil {
		return
	}

	switch BridgeMessageType(msg.Type) {
	case MessageTypeThought:
		// Update thought count in orchestration
		cb.deepTreeEcho.ThoughtCount++

	case MessageTypePattern:
		// Update echo patterns in orchestration
		if pattern, ok := msg.Payload.(map[string]interface{}); ok {
			if name, ok := pattern["name"].(string); ok {
				if strength, ok := pattern["strength"].(float64); ok {
					cb.updateEchoPattern(name, strength)
				}
			}
		}

	case MessageTypeWisdom:
		// Update evolution progress
		if cb.deepTreeEcho.EvolutionTimeline != nil {
			cb.deepTreeEcho.EvolutionTimeline.Progress += 0.001
		}

	case MessageTypeMetrics:
		// Update memory resonance
		if metrics, ok := msg.Payload.(map[string]interface{}); ok {
			if cb.deepTreeEcho.MemoryResonance != nil {
				if nodes, ok := metrics["memory_nodes"].(int); ok {
					cb.deepTreeEcho.MemoryResonance.MemoryNodes = nodes
				}
				if connections, ok := metrics["connections"].(int); ok {
					cb.deepTreeEcho.MemoryResonance.Connections = connections
				}
			}
		}
	}
}

// processOrchMessage processes a message from the orchestration layer
func (cb *CognitiveBridge) processOrchMessage(msg BridgeMessage) {
	cb.mu.Lock()
	cb.messagesReceived++
	cb.mu.Unlock()

	if cb.autonomousAgent == nil {
		return
	}

	switch BridgeMessageType(msg.Type) {
	case MessageTypeCommand:
		// Process command from orchestration
		if cmd, ok := msg.Payload.(map[string]interface{}); ok {
			if action, ok := cmd["action"].(string); ok {
				switch action {
				case "store_knowledge":
					if subject, ok := cmd["subject"].(string); ok {
						if predicate, ok := cmd["predicate"].(string); ok {
							if object, ok := cmd["object"].(string); ok {
								cb.autonomousAgent.StoreKnowledge(subject, predicate, object)
							}
						}
					}
				}
			}
		}

	case MessageTypeStateSync:
		// State sync request from orchestration
		// The core will respond with its current state
		cb.sendToOrchestration(BridgeMessage{
			Type: string(MessageTypeStateSync),
			Source: "core",
			Destination: "orchestration",
			Payload: cb.autonomousAgent.GetStatus(),
			Timestamp: time.Now(),
		})
	}
}

// updateEchoPattern updates an echo pattern in the orchestration layer
func (cb *CognitiveBridge) updateEchoPattern(name string, strength float64) {
	if cb.deepTreeEcho == nil || cb.deepTreeEcho.EchoPatterns == nil {
		return
	}

	patterns := cb.deepTreeEcho.EchoPatterns

	switch name {
	case "recursive_self_improvement":
		if patterns.RecursiveSelfImprovement != nil {
			patterns.RecursiveSelfImprovement.Strength = strength
		}
	case "cross_system_synthesis":
		if patterns.CrossSystemSynthesis != nil {
			patterns.CrossSystemSynthesis.Strength = strength
		}
	case "identity_preservation":
		if patterns.IdentityPreservation != nil {
			patterns.IdentityPreservation.Strength = strength
		}
	case "spatial_awareness":
		if patterns.SpatialAwareness != nil {
			patterns.SpatialAwareness.Strength = strength
		}
	case "emotional_resonance":
		if patterns.EmotionalResonance != nil {
			patterns.EmotionalResonance.Strength = strength
		}
	}

	patterns.LastUpdated = time.Now()
}

// SendToCore sends a message from orchestration to core
func (cb *CognitiveBridge) SendToCore(msg BridgeMessage) error {
	if !cb.running {
		return fmt.Errorf("bridge not running")
	}

	msg.Destination = "core"
	msg.Timestamp = time.Now()

	select {
	case cb.orchToCoreChan <- msg:
		cb.mu.Lock()
		cb.messagesSent++
		cb.mu.Unlock()
		return nil
	default:
		cb.mu.Lock()
		cb.errorCount++
		cb.mu.Unlock()
		return fmt.Errorf("orchestration to core channel full")
	}
}

// sendToOrchestration sends a message from core to orchestration
func (cb *CognitiveBridge) sendToOrchestration(msg BridgeMessage) error {
	if !cb.running {
		return fmt.Errorf("bridge not running")
	}

	msg.Destination = "orchestration"
	msg.Timestamp = time.Now()

	select {
	case cb.coreToOrchChan <- msg:
		cb.mu.Lock()
		cb.messagesSent++
		cb.mu.Unlock()
		return nil
	default:
		cb.mu.Lock()
		cb.errorCount++
		cb.mu.Unlock()
		return fmt.Errorf("core to orchestration channel full")
	}
}

// GetMetrics returns bridge metrics
func (cb *CognitiveBridge) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"running": cb.running,
		"messages_sent": cb.messagesSent,
		"messages_received": cb.messagesReceived,
		"sync_count": cb.syncCount,
		"error_count": cb.errorCount,
		"last_core_sync": cb.lastCoreSync,
		"last_orch_sync": cb.lastOrchSync,
	}
}

// ForceSync forces an immediate state synchronization
func (cb *CognitiveBridge) ForceSync() {
	cb.synchronizeState()
}

// IsRunning returns whether the bridge is running
func (cb *CognitiveBridge) IsRunning() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.running
}
