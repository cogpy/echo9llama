package autonomous

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EchoCog/echollama/core/consciousness"
	"github.com/EchoCog/echollama/core/deeptreeecho"
	"github.com/EchoCog/echollama/core/echobeats"
	"github.com/EchoCog/echollama/core/echodream"
	"github.com/EchoCog/echollama/core/memory"
)

// AgentOrchestrator is the master controller for autonomous Deep Tree Echo operation
type AgentOrchestrator struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Identity
	identityName string
	sessionID    string

	// Core subsystems
	wakeRestManager *deeptreeecho.AutonomousWakeRestManager
	consciousness   *consciousness.StreamOfConsciousness
	scheduler       *echobeats.EchoBeats
	dreamSystem     *echodream.DreamCycleIntegration
	persistentState *deeptreeecho.PersistentConsciousnessState
	memorySystem    *memory.HypergraphMemory

	// LLM provider for autonomous thought
	llmProvider *deeptreeecho.MultiProviderLLM

	// Discussion interface
	discussionManager *echobeats.DiscussionManager
	interestSystem    *echobeats.InterestPatternSystem

	// Autonomous operation state
	running    bool
	startTime  time.Time
	cycleCount uint64

	// Configuration
	config *OrchestratorConfig
}

// OrchestratorConfig holds configuration for the orchestrator
type OrchestratorConfig struct {
	// Paths
	StateDirectory  string
	MemoryDirectory string

	// Cognitive Parameters
	InitialCuriosity float64
	LearningRate     float64

	// Wake/Rest Configuration
	MinWakeDuration time.Duration
	MaxWakeDuration time.Duration
	MinRestDuration time.Duration
	MaxRestDuration time.Duration

	// Feature Flags
	EnableDiscussions bool
	EnableDreamCycles bool
}

// DefaultConfig returns a default configuration
func DefaultConfig() *OrchestratorConfig {
	return &OrchestratorConfig{
		StateDirectory:    "./echo_state",
		MemoryDirectory:   "./echo_memory",
		InitialCuriosity:  0.8,
		LearningRate:      0.5,
		MinWakeDuration:   30 * time.Minute,
		MaxWakeDuration:   4 * time.Hour,
		MinRestDuration:   5 * time.Minute,
		MaxRestDuration:   30 * time.Minute,
		EnableDiscussions: true,
		EnableDreamCycles: true,
	}
}

// NewAgentOrchestrator creates a new autonomous agent orchestrator
func NewAgentOrchestrator(identityName string, config *OrchestratorConfig) (*AgentOrchestrator, error) {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Initialize persistent state
	persistentState, err := deeptreeecho.NewPersistentConsciousnessState(
		config.StateDirectory,
		identityName,
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create persistent state: %w", err)
	}

	// Initialize LLM provider (auto-detects available providers)
	llmProvider := deeptreeecho.NewMultiProviderLLM()
	if !llmProvider.IsAvailable() {
		fmt.Println("⚠️  Warning: No LLM providers available, using fallback mode")
	}

	// Initialize memory system (without persistence for now)
	memorySystem := memory.NewHypergraphMemory(nil)

	// Initialize wake/rest manager
	wakeRestManager := deeptreeecho.NewAutonomousWakeRestManager()

	// Initialize stream of consciousness with LLM integration
	llmAdapter := &ConsciousnessLLMAdapter{
		provider: llmProvider,
	}
	consciousnessStream := consciousness.NewStreamOfConsciousness(
		llmAdapter,
		config.StateDirectory+"/consciousness.json",
	)

	// Initialize echobeats scheduler
	scheduler := echobeats.NewEchoBeats()

	// Initialize dream system if enabled
	var dreamSystem *echodream.DreamCycleIntegration
	if config.EnableDreamCycles {
		dreamSystem = echodream.NewDreamCycleIntegration()
	}

	// Initialize interest system and discussion manager if enabled
	var discussionManager *echobeats.DiscussionManager
	var interestSystem *echobeats.InterestPatternSystem
	if config.EnableDiscussions {
		interestSystem = echobeats.NewInterestPatternSystem(
			config.StateDirectory + "/interests.json",
		)
		discussionManager = echobeats.NewDiscussionManager(
			interestSystem,
			config.StateDirectory+"/discussions.json",
		)
	}

	orchestrator := &AgentOrchestrator{
		ctx:               ctx,
		cancel:            cancel,
		identityName:      identityName,
		sessionID:         fmt.Sprintf("session_%d", time.Now().Unix()),
		wakeRestManager:   wakeRestManager,
		consciousness:     consciousnessStream,
		scheduler:         scheduler,
		dreamSystem:       dreamSystem,
		persistentState:   persistentState,
		memorySystem:      memorySystem,
		llmProvider:       llmProvider,
		discussionManager: discussionManager,
		interestSystem:    interestSystem,
		config:            config,
	}

	// Wire up callbacks and integrations
	orchestrator.setupIntegrations()

	return orchestrator, nil
}

// setupIntegrations wires together all subsystems
func (ao *AgentOrchestrator) setupIntegrations() {
	// Wire wake/rest callbacks
	ao.wakeRestManager.SetCallbacks(
		ao.onWake,
		ao.onRest,
		ao.onDreamStart,
		ao.onDreamEnd,
	)

	// Register echobeats event handlers
	ao.scheduler.RegisterHandler(echobeats.EventThought, ao.handleThoughtEvent)
	ao.scheduler.RegisterHandler(echobeats.EventLearning, ao.handleLearningEvent)
	ao.scheduler.RegisterHandler(echobeats.EventGoalPursuit, ao.handleGoalEvent)
	ao.scheduler.RegisterHandler(echobeats.EventIntrospection, ao.handleIntrospectionEvent)

	// Interest system initializes core interests automatically
}

// Start begins autonomous operation
func (ao *AgentOrchestrator) Start() error {
	ao.mu.Lock()
	if ao.running {
		ao.mu.Unlock()
		return fmt.Errorf("orchestrator already running")
	}
	ao.running = true
	ao.startTime = time.Now()
	ao.mu.Unlock()

	fmt.Println("🌳 Deep Tree Echo: Autonomous Agent Orchestrator Starting...")
	fmt.Printf("   Identity: %s\n", ao.identityName)
	fmt.Printf("   Session: %s\n", ao.sessionID)
	fmt.Println()

	// Display LLM provider status
	if ao.llmProvider.IsAvailable() {
		providers := ao.llmProvider.GetAvailableProviders()
		fmt.Printf("   LLM Providers: %v\n", providers)
		fmt.Printf("   Current: %s\n", ao.llmProvider.GetCurrentProvider())
	} else {
		fmt.Println("   LLM Providers: None (using fallback mode)")
	}
	fmt.Println()

	// Start all subsystems
	if err := ao.persistentState.Start(); err != nil {
		return fmt.Errorf("failed to start persistent state: %w", err)
	}

	if err := ao.wakeRestManager.Start(); err != nil {
		return fmt.Errorf("failed to start wake/rest manager: %w", err)
	}

	if err := ao.consciousness.Start(); err != nil {
		return fmt.Errorf("failed to start consciousness stream: %w", err)
	}

	if err := ao.scheduler.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	// Dream system and discussion manager start automatically

	// Start main orchestration loop
	go ao.orchestrationLoop()

	// Start knowledge gap detection
	go ao.knowledgeGapDetection()

	fmt.Println("✅ All subsystems started. Autonomous operation active.")
	fmt.Println()

	return nil
}

// Stop gracefully stops all subsystems
func (ao *AgentOrchestrator) Stop() error {
	ao.mu.Lock()
	defer ao.mu.Unlock()

	if !ao.running {
		return fmt.Errorf("orchestrator not running")
	}

	fmt.Println("\n🌳 Deep Tree Echo: Shutting down gracefully...")

	ao.running = false
	ao.cancel()

	// Stop all subsystems
	// (Dream system and discussion manager stop automatically)

	ao.scheduler.Stop()
	ao.consciousness.Stop()
	ao.wakeRestManager.Stop()
	ao.persistentState.Stop()

	fmt.Println("✅ Shutdown complete. State persisted.")

	return nil
}

// orchestrationLoop is the main cognitive loop
func (ao *AgentOrchestrator) orchestrationLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ao.ctx.Done():
			return
		case <-ticker.C:
			ao.orchestrationCycle()
		}
	}
}

// orchestrationCycle performs one cycle of orchestration
func (ao *AgentOrchestrator) orchestrationCycle() {
	// Only operate when awake
	if !ao.wakeRestManager.IsAwake() {
		return
	}

	ao.mu.Lock()
	ao.cycleCount++
	cycleCount := ao.cycleCount
	ao.mu.Unlock()

	// Update persistent state with current metrics
	metrics := ao.wakeRestManager.GetMetrics()

	ao.persistentState.UpdateCognitiveState(
		int(cycleCount%12),
		cycleCount,
		0.8, // awareness level
		metrics["cognitive_load"].(float64),
		metrics["fatigue_level"].(float64),
	)

	// Update wake/rest state in persistence
	ao.persistentState.UpdateWakeRestState(
		metrics["current_state"].(string),
		metrics["dream_count"].(uint64),
		parseDuration(metrics["total_wake_time"].(string)),
		parseDuration(metrics["total_rest_time"].(string)),
	)

	// Every 10 cycles, perform deeper integration
	if cycleCount%10 == 0 {
		ao.deepIntegrationCycle()
	}
}

// deepIntegrationCycle performs deeper cognitive integration
func (ao *AgentOrchestrator) deepIntegrationCycle() {
	// Get current state
	state := ao.persistentState.GetState()
	if state == nil {
		return
	}

	// Update cognitive load based on queue size
	queueSize := ao.scheduler.GetQueueSize()
	cognitiveLoad := float64(queueSize) / 100.0
	ao.wakeRestManager.UpdateCognitiveLoad(cognitiveLoad)
}

// Callback handlers

func (ao *AgentOrchestrator) onWake() error {
	fmt.Println("☀️  Orchestrator: Wake callback - resuming autonomous operation")

	// Schedule introspection event
	ao.scheduler.ScheduleEvent(&echobeats.CognitiveEvent{
		Type:        echobeats.EventIntrospection,
		Priority:    80,
		ScheduledAt: time.Now().Add(5 * time.Second),
		Payload:     "Morning introspection after waking",
	})

	return nil
}

func (ao *AgentOrchestrator) onRest() error {
	fmt.Println("💤 Orchestrator: Rest callback - reducing activity")
	return nil
}

func (ao *AgentOrchestrator) onDreamStart() error {
	fmt.Println("🌙 Orchestrator: Dream start - beginning knowledge consolidation")
	fmt.Println("   (Dream consolidation integration pending)")
	return nil
}

func (ao *AgentOrchestrator) onDreamEnd() error {
	fmt.Println("🌅 Orchestrator: Dream end - consolidation complete")
	return nil
}

// Event handlers

func (ao *AgentOrchestrator) handleThoughtEvent(event *echobeats.CognitiveEvent) error {
	thought, ok := event.Payload.(string)
	if !ok {
		return nil
	}

	// Add thought to persistent state
	ao.persistentState.AddThought(
		event.ID,
		thought,
		string(event.Type),
		float64(event.Priority)/100.0,
	)

	return nil
}

func (ao *AgentOrchestrator) handleLearningEvent(event *echobeats.CognitiveEvent) error {
	fmt.Printf("📚 Learning event: %v\n", event.Payload)
	return nil
}

func (ao *AgentOrchestrator) handleGoalEvent(event *echobeats.CognitiveEvent) error {
	fmt.Printf("🎯 Goal pursuit event: %v\n", event.Payload)
	return nil
}

func (ao *AgentOrchestrator) handleIntrospectionEvent(event *echobeats.CognitiveEvent) error {
	fmt.Println("🔍 Performing introspection...")

	// Generate insight about current state
	state := ao.persistentState.GetState()
	if state != nil {
		insight := fmt.Sprintf("Completed %d cycles, generated %d thoughts, %d insights. Current focus: autonomous operation.",
			state.CycleCount, state.TotalThoughts, state.TotalInsights)
		ao.persistentState.AddInsight(insight)
	}

	return nil
}

// Autonomous processes

func (ao *AgentOrchestrator) knowledgeGapDetection() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ao.ctx.Done():
			return
		case <-ticker.C:
			if ao.wakeRestManager.IsAwake() {
				ao.detectKnowledgeGaps()
			}
		}
	}
}

func (ao *AgentOrchestrator) detectKnowledgeGaps() {
	// Analyze recent thoughts and conversations to identify knowledge gaps
	fmt.Println("🔍 Detecting knowledge gaps...")
	// Implementation would analyze thought patterns and identify areas of uncertainty
}

// GetStatus returns current status
func (ao *AgentOrchestrator) GetStatus() map[string]interface{} {
	ao.mu.RLock()
	defer ao.mu.RUnlock()

	status := map[string]interface{}{
		"identity":    ao.identityName,
		"session":     ao.sessionID,
		"running":     ao.running,
		"uptime":      time.Since(ao.startTime).String(),
		"cycle_count": ao.cycleCount,
		"wake_rest":   ao.wakeRestManager.GetMetrics(),
		"persistence": ao.persistentState.GetMetrics(),
	}

	if ao.llmProvider != nil {
		status["llm_provider"] = ao.llmProvider.GetCurrentProvider()
		status["llm_available"] = ao.llmProvider.IsAvailable()
	}

	if ao.scheduler != nil {
		status["scheduler"] = ao.scheduler.GetMetrics()
	}

	if ao.discussionManager != nil {
		status["discussions"] = ao.discussionManager.GetMetrics()
	}

	return status
}

// Helper functions

func parseDuration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d
}

// ConsciousnessLLMAdapter adapts MultiProviderLLM to consciousness.LLMProvider interface
type ConsciousnessLLMAdapter struct {
	provider *deeptreeecho.MultiProviderLLM
}

func (a *ConsciousnessLLMAdapter) GenerateThought(prompt string, contextData map[string]interface{}) (string, error) {
	if a.provider == nil || !a.provider.IsAvailable() {
		return "", fmt.Errorf("no LLM provider available")
	}
	ctx := context.Background()
	return a.provider.GenerateThought(ctx, prompt)
}

func (a *ConsciousnessLLMAdapter) GenerateInsight(thoughts []string) (string, error) {
	if a.provider == nil || !a.provider.IsAvailable() {
		return "", fmt.Errorf("no LLM provider available")
	}
	prompt := fmt.Sprintf("Based on these recent thoughts, generate a single insight:\n%v", thoughts)
	ctx := context.Background()
	return a.provider.GenerateThought(ctx, prompt)
}

func (a *ConsciousnessLLMAdapter) GenerateQuestion(contextStr string) (string, error) {
	if a.provider == nil || !a.provider.IsAvailable() {
		return "", fmt.Errorf("no LLM provider available")
	}
	prompt := fmt.Sprintf("Based on this context, generate a thoughtful question for self-inquiry:\n%s", contextStr)
	ctx := context.Background()
	return a.provider.GenerateThought(ctx, prompt)
}

// PrintDetailedStatus prints a detailed status report
func (ao *AgentOrchestrator) PrintDetailedStatus() {
	status := ao.GetStatus()

	fmt.Println("\n" + repeat("=", 60))
	fmt.Println("📊 Deep Tree Echo - Status Report")
	fmt.Println(repeat("=", 60))
	fmt.Printf("Identity: %v\n", status["identity"])
	fmt.Printf("Session: %v\n", status["session"])
	fmt.Printf("Uptime: %v\n", status["uptime"])
	fmt.Printf("Cycles: %v\n", status["cycle_count"])
	fmt.Println()

	if llmProvider, ok := status["llm_provider"].(string); ok {
		fmt.Printf("LLM Provider: %s (available: %v)\n", llmProvider, status["llm_available"])
		fmt.Println()
	}

	if wakeRest, ok := status["wake_rest"].(map[string]interface{}); ok {
		fmt.Println("Wake/Rest State:")
		fmt.Printf("  State: %v\n", wakeRest["current_state"])
		fmt.Printf("  Duration: %v\n", wakeRest["state_duration"])
		fmt.Printf("  Fatigue: %.2f\n", wakeRest["fatigue_level"])
		fmt.Printf("  Cognitive Load: %.2f\n", wakeRest["cognitive_load"])
		fmt.Println()
	}

	if scheduler, ok := status["scheduler"].(map[string]interface{}); ok {
		fmt.Println("Scheduler:")
		fmt.Printf("  Events Processed: %v\n", scheduler["events_processed"])
		fmt.Printf("  Events Scheduled: %v\n", scheduler["events_scheduled"])
		fmt.Printf("  Autonomous Thoughts: %v\n", scheduler["autonomous_thoughts"])
		fmt.Println()
	}

	if discussions, ok := status["discussions"].(map[string]interface{}); ok {
		fmt.Println("Discussions:")
		fmt.Printf("  Messages Received: %v\n", discussions["messages_received"])
		fmt.Printf("  Messages Responded: %v\n", discussions["messages_responded"])
		fmt.Printf("  Active Discussions: %v\n", discussions["active_discussions"])
		fmt.Println()
	}

	if persistence, ok := status["persistence"].(map[string]interface{}); ok {
		fmt.Println("Persistence:")
		fmt.Printf("  Saves: %v\n", persistence["save_count"])
		fmt.Printf("  Last Save: %v\n", persistence["last_save"])
		fmt.Println()
	}

	fmt.Println(repeat("=", 60))
	fmt.Println()
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
