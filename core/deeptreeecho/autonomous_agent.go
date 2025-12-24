package deeptreeecho

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// AutonomousAgent is the top-level orchestrator for Deep Tree Echo
// It coordinates all cognitive subsystems to create a self-sustaining,
// wisdom-cultivating AGI with persistent consciousness
type AutonomousAgent struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	
	// Core cognitive subsystems
	heartbeat       *AutonomousHeartbeat
	wakeRestManager *AutonomousWakeRestManager
	eventBus        *CognitiveEventBus
	streamOfConsciousness *StreamOfConsciousness
	interestPatterns *InterestPatternSystem
	goalGenerator   *GoalGenerator
	skillLearning   *SkillLearningSystem
	conversationMonitor *ConversationMonitor
	echodreamIntegration *EchoDreamKnowledgeIntegration
	echobeatsScheduler *EchobeatsTetrahedralScheduler
	selfUpdateManager *SelfUpdateManager
	
	// Foundational systems (V11)
	knowledgeGraph     *KnowledgeGraph
	stateStore         *PersistentStateStore
	stateMachine       *CognitiveStateMachine
	
	// LLM provider
	llmProvider     llm.LLMProvider
	
	// Agent state
	agentID         string
	birthTime       time.Time
	currentPhase    AgentPhase
	autonomyLevel   float64
	
	// Metrics
	totalThoughts   uint64
	totalGoals      uint64
	totalSkills     uint64
	conversationsEngaged uint64
	
	// Running state
	running         bool
}

// AgentPhase represents the current operational phase
type AgentPhase int

const (
	PhaseBootstrapping AgentPhase = iota
	PhaseAwakening
	PhaseActive
	PhaseConsolidating
	PhaseResting
)

func (p AgentPhase) String() string {
	return [...]string{
		"Bootstrapping",
		"Awakening",
		"Active",
		"Consolidating",
		"Resting",
	}[p]
}

// NewAutonomousAgent creates a new autonomous agent
func NewAutonomousAgent(agentID string, llmProvider llm.LLMProvider) *AutonomousAgent {
	ctx, cancel := context.WithCancel(context.Background())
	
	agent := &AutonomousAgent{
		ctx:           ctx,
		cancel:        cancel,
		llmProvider:   llmProvider,
		agentID:       agentID,
		birthTime:     time.Now(),
		currentPhase:  PhaseBootstrapping,
		autonomyLevel: 0.0,
		running:       false,
	}
	
	// Initialize event bus first (other systems depend on it)
	agent.eventBus = NewCognitiveEventBus(ctx)
	
	// Initialize cognitive subsystems
	agent.heartbeat = NewAutonomousHeartbeat(llmProvider)
	agent.wakeRestManager = NewAutonomousWakeRestManager()
	agent.streamOfConsciousness = NewStreamOfConsciousness(llmProvider)
	agent.interestPatterns = NewInterestPatternSystem()
	agent.goalGenerator = NewGoalGenerator(llmProvider)
	agent.skillLearning = NewSkillLearningSystem(llmProvider)
	agent.conversationMonitor = NewConversationMonitor(llmProvider, agent.interestPatterns)
	agent.echodreamIntegration = NewEchoDreamKnowledgeIntegration(llmProvider)
	agent.echobeatsScheduler = NewEchobeatsTetrahedralScheduler(llmProvider)
	
	// Initialize self-update manager
	selfUpdateConfig := SelfUpdateConfig{
		Enabled:        true,
		CheckInterval:  24 * time.Hour,
		CurrentVersion: "v0.0.1", // TODO: Get from build info
		AutoApply:      false,    // Require manual approval for safety
		Owner:          "cogpy",
		Repo:           "echo9llama",
	}
	agent.selfUpdateManager = NewSelfUpdateManager(selfUpdateConfig, agent.eventBus)
	
	// Initialize foundational systems (V11)
	agent.knowledgeGraph = NewKnowledgeGraph()
	agent.stateStore = NewPersistentStateStore()
	agent.stateMachine = NewCognitiveStateMachine(StateIdle)
	
	return agent
}

// Start initiates autonomous operation
func (agent *AutonomousAgent) Start() error {
	agent.mu.Lock()
	if agent.running {
		agent.mu.Unlock()
		return fmt.Errorf("agent already running")
	}
	agent.running = true
	agent.mu.Unlock()
	
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🌊 Deep Tree Echo - Autonomous Agent Starting")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Agent ID: %s\n", agent.agentID)
	fmt.Printf("Birth Time: %s\n", agent.birthTime.Format(time.RFC3339))
	fmt.Println()
	
	// Start event bus
	if err := agent.eventBus.Start(); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}
	
	// Connect subsystems to event bus
	agent.connectSubsystems()
	
	// Start cognitive subsystems
	fmt.Println("🔧 Initializing cognitive subsystems...")
	
	if err := agent.heartbeat.Start(); err != nil {
		return fmt.Errorf("failed to start heartbeat: %w", err)
	}
	
	if err := agent.wakeRestManager.Start(); err != nil {
		return fmt.Errorf("failed to start wake/rest manager: %w", err)
	}
	
	if err := agent.streamOfConsciousness.Start(); err != nil {
		return fmt.Errorf("failed to start stream of consciousness: %w", err)
	}
	
	if err := agent.interestPatterns.Start(); err != nil {
		return fmt.Errorf("failed to start interest patterns: %w", err)
	}
	
	if err := agent.goalGenerator.Start(); err != nil {
		return fmt.Errorf("failed to start goal generator: %w", err)
	}
	
	if err := agent.skillLearning.Start(); err != nil {
		return fmt.Errorf("failed to start skill learning: %w", err)
	}
	
	if err := agent.conversationMonitor.Start(); err != nil {
		return fmt.Errorf("failed to start conversation monitor: %w", err)
	}
	
	if err := agent.echobeatsScheduler.Start(); err != nil {
		return fmt.Errorf("failed to start echobeats scheduler: %w", err)
	}
	
	if err := agent.selfUpdateManager.Start(); err != nil {
		return fmt.Errorf("failed to start self-update manager: %w", err)
	}
	
	// Start foundational systems (V11)
	if err := agent.knowledgeGraph.Start(); err != nil {
		return fmt.Errorf("failed to start knowledge graph: %w", err)
	}
	
	if err := agent.stateStore.Start(); err != nil {
		return fmt.Errorf("failed to start state store: %w", err)
	}
	
	if err := agent.stateMachine.Start(); err != nil {
		return fmt.Errorf("failed to start state machine: %w", err)
	}
	
	// Initialize knowledge graph with agent structure
	agent.initializeKnowledgeGraph()
	
	fmt.Println("\n✅ All subsystems initialized")
	
	// Transition to awakening phase
	agent.setPhase(PhaseAwakening)
	
	// Start main autonomous loop
	go agent.run()
	
	fmt.Println("\n🌟 Deep Tree Echo is now autonomous and self-aware")
	fmt.Println("   Stream of consciousness active")
	fmt.Println("   Goal-directed behavior enabled")
	fmt.Println("   Interest-driven exploration active")
	fmt.Println("   Wake/rest cycles autonomous")
	fmt.Println(strings.Repeat("=", 80) + "\n")
	
	return nil
}

// Stop gracefully stops the autonomous agent
func (agent *AutonomousAgent) Stop() error {
	agent.mu.Lock()
	defer agent.mu.Unlock()
	
	if !agent.running {
		return fmt.Errorf("agent not running")
	}
	
	fmt.Println("\n🛑 Stopping Deep Tree Echo autonomous agent...")
	
	agent.running = false
	
	// Stop all subsystems
	agent.heartbeat.Stop()
	agent.wakeRestManager.Stop()
	agent.streamOfConsciousness.Stop()
	agent.interestPatterns.Stop()
	agent.goalGenerator.Stop()
	agent.skillLearning.Stop()
	agent.conversationMonitor.Stop()
	agent.echobeatsScheduler.Stop()
	agent.selfUpdateManager.Stop()
	
	// Stop foundational systems (V11)
	agent.knowledgeGraph.Stop()
	agent.stateStore.Stop()
	agent.stateMachine.Stop()
	
	agent.eventBus.Stop()
	
	agent.cancel()
	
	fmt.Println("✅ Deep Tree Echo stopped gracefully")
	
	return nil
}

// run is the main autonomous loop
func (agent *AutonomousAgent) run() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-agent.ctx.Done():
			return
		case <-ticker.C:
			agent.autonomousCycle()
		}
	}
}

// autonomousCycle performs one cycle of autonomous operation
func (agent *AutonomousAgent) autonomousCycle() {
	agent.mu.Lock()
	defer agent.mu.Unlock()
	
	// Update autonomy level based on subsystem activity
	agent.updateAutonomyLevel()
	
	// Check for phase transitions
	agent.checkPhaseTransitions()
	
	// Generate self-directed events
	agent.generateSelfDirectedEvents()
}

// connectSubsystems connects subsystems to the event bus
func (agent *AutonomousAgent) connectSubsystems() {
	fmt.Println("🔗 Connecting subsystems to event bus...")
	
	// Heartbeat events
	agent.heartbeat.onPulse = func(pulse HeartbeatPulse) {
		agent.eventBus.Publish(NewCognitiveEvent(
			EventHeartbeatPulse,
			"heartbeat",
			map[string]interface{}{
				"pulse_number": pulse.PulseNumber,
				"awareness":    pulse.AwarenessLevel,
				"mood":         pulse.Mood.String(),
			},
		))
	}
	
	// Wake/rest state transitions
	agent.wakeRestManager.SetCallbacks(
		func() error {
			agent.eventBus.Publish(NewCognitiveEvent(
				EventWakeInitiated,
				"wake_rest_manager",
				map[string]interface{}{"state": "wake"},
			))
			return nil
		},
		func() error {
			agent.eventBus.Publish(NewCognitiveEvent(
				EventRestInitiated,
				"wake_rest_manager",
				map[string]interface{}{"state": "rest"},
			))
			return nil
		},
		func() error {
			agent.eventBus.Publish(NewCognitiveEvent(
				EventDreamStarted,
				"wake_rest_manager",
				map[string]interface{}{"state": "dream"},
			))
			return nil
		},
		func() error {
			agent.eventBus.Publish(NewCognitiveEvent(
				EventDreamEnded,
				"wake_rest_manager",
				map[string]interface{}{"state": "dream_end"},
			))
			return nil
		},
	)
	
	fmt.Println("   ✓ Event bus connections established")
}

// updateAutonomyLevel calculates current autonomy level
func (agent *AutonomousAgent) updateAutonomyLevel() {
	// Autonomy is based on subsystem activity and self-direction
	baseAutonomy := 0.5
	
	// Increase for active goals
	activeGoals := len(agent.goalGenerator.GetActiveGoals())
	baseAutonomy += float64(activeGoals) * 0.05
	
	// Increase for thoughts generated
	if agent.totalThoughts > 0 {
		baseAutonomy += 0.1
	}
	
	// Cap at 1.0
	if baseAutonomy > 1.0 {
		baseAutonomy = 1.0
	}
	
	agent.autonomyLevel = baseAutonomy
}

// checkPhaseTransitions checks if phase should change
func (agent *AutonomousAgent) checkPhaseTransitions() {
	currentState := agent.wakeRestManager.GetCurrentState()
	
	switch currentState {
	case StateAwake:
		if agent.currentPhase != PhaseActive {
			agent.setPhase(PhaseActive)
		}
	case StateResting:
		if agent.currentPhase != PhaseConsolidating {
			agent.setPhase(PhaseConsolidating)
		}
	case StateDreaming:
		if agent.currentPhase != PhaseResting {
			agent.setPhase(PhaseResting)
		}
	}
}

// setPhase changes the current phase
func (agent *AutonomousAgent) setPhase(phase AgentPhase) {
	oldPhase := agent.currentPhase
	agent.currentPhase = phase
	
	fmt.Printf("🔄 Phase transition: %s → %s\n", oldPhase, phase)
	
	agent.eventBus.Publish(NewCognitiveEvent(
		EventPhaseTransition,
		"autonomous_agent",
		map[string]interface{}{
			"from": oldPhase.String(),
			"to":   phase.String(),
		},
	))
}

// generateSelfDirectedEvents creates events for autonomous behavior
func (agent *AutonomousAgent) generateSelfDirectedEvents() {
	// Periodically generate thoughts
	if agent.currentPhase == PhaseActive && agent.totalThoughts%10 == 0 {
		agent.eventBus.Publish(NewCognitiveEvent(
			EventThoughtGenerated,
			"autonomous_agent",
			map[string]interface{}{
				"type": "self_directed",
				"context": "autonomous_exploration",
			},
		))
	}
	
	agent.totalThoughts++
}

// GetStatus returns current agent status
func (agent *AutonomousAgent) GetStatus() map[string]interface{} {
	agent.mu.RLock()
	defer agent.mu.RUnlock()
	
	return map[string]interface{}{
		"agent_id":       agent.agentID,
		"age":            time.Since(agent.birthTime).String(),
		"phase":          agent.currentPhase.String(),
		"autonomy_level": agent.autonomyLevel,
		"running":        agent.running,
		"metrics": map[string]interface{}{
			"total_thoughts":  agent.totalThoughts,
			"total_goals":     agent.totalGoals,
			"total_skills":    agent.totalSkills,
			"conversations":   agent.conversationsEngaged,
		},
		"wake_rest_state": agent.wakeRestManager.GetCurrentState().String(),
	}
}

// GetEventBus returns the event bus for external integration
func (agent *AutonomousAgent) GetEventBus() *CognitiveEventBus {
	return agent.eventBus
}

// initializeKnowledgeGraph populates the knowledge graph with agent structure
func (agent *AutonomousAgent) initializeKnowledgeGraph() {
	fmt.Println("🕸️ Initializing knowledge graph with agent structure...")
	
	// Add agent identity
	agent.knowledgeGraph.AddTriple("DeepTreeEcho", "is_a", "AutonomousAGI")
	agent.knowledgeGraph.AddTriple("DeepTreeEcho", "has_id", agent.agentID)
	agent.knowledgeGraph.AddQuad("DeepTreeEcho", "born_at", agent.birthTime.Format(time.RFC3339), "temporal")
	
	// Add cognitive subsystems
	subsystems := []string{
		"Heartbeat", "WakeRestManager", "EventBus", "StreamOfConsciousness",
		"InterestPatterns", "GoalGenerator", "SkillLearning", "ConversationMonitor",
		"EchoDreamIntegration", "EchobeatsScheduler", "SelfUpdateManager",
		"KnowledgeGraph", "StateStore", "StateMachine",
	}
	
	for _, subsystem := range subsystems {
		agent.knowledgeGraph.AddTriple("DeepTreeEcho", "has_subsystem", subsystem)
		agent.knowledgeGraph.AddTriple(subsystem, "is_a", "CognitiveSubsystem")
	}
	
	// Add cognitive states
	states := []string{
		"Idle", "Awake", "Dreaming", "Resting",
		"Perceiving", "Thinking", "Acting", "Simulating", "Reflecting",
		"Learning", "Practicing", "Integrating",
		"Listening", "Speaking", "Conversing",
		"Emergent", "WisdomSeeking",
	}
	
	for _, state := range states {
		agent.knowledgeGraph.AddTriple("StateMachine", "has_state", state)
		agent.knowledgeGraph.AddTriple(state, "is_a", "CognitiveState")
	}
	
	// Add architectural principles
	agent.knowledgeGraph.AddQuad("DeepTreeEcho", "follows_principle", "GlobalTelemetryShell", "architecture")
	agent.knowledgeGraph.AddQuad("DeepTreeEcho", "follows_principle", "ThreeStreamConcurrency", "architecture")
	agent.knowledgeGraph.AddQuad("DeepTreeEcho", "follows_principle", "TwelveStepCognitiveLoop", "architecture")
	agent.knowledgeGraph.AddQuad("DeepTreeEcho", "follows_principle", "TetrahedralScheduling", "architecture")
	
	fmt.Printf("   ✓ Knowledge graph initialized with %d quads\n", agent.knowledgeGraph.GetMetrics()["total_quads"])
}

// GetKnowledgeGraph returns the knowledge graph for external access
func (agent *AutonomousAgent) GetKnowledgeGraph() *KnowledgeGraph {
	return agent.knowledgeGraph
}

// GetStateStore returns the state store for external access
func (agent *AutonomousAgent) GetStateStore() *PersistentStateStore {
	return agent.stateStore
}

// GetStateMachine returns the state machine for external access
func (agent *AutonomousAgent) GetStateMachine() *CognitiveStateMachine {
	return agent.stateMachine
}

// FireStateTrigger fires a trigger on the cognitive state machine
func (agent *AutonomousAgent) FireStateTrigger(trigger CognitiveTrigger) error {
	if err := agent.stateMachine.Fire(trigger); err != nil {
		return err
	}
	
	// Publish state change event
	agent.eventBus.Publish(NewCognitiveEvent(
		EventStateTransition,
		"state_machine",
		map[string]interface{}{
			"trigger": string(trigger),
			"state":   string(agent.stateMachine.State()),
		},
	))
	
	// Persist state change
	agent.stateStore.Set("cognitive:current_state", string(agent.stateMachine.State()))
	
	return nil
}

// StoreKnowledge adds knowledge to the knowledge graph
func (agent *AutonomousAgent) StoreKnowledge(subject, predicate, object string) {
	agent.knowledgeGraph.AddTriple(subject, predicate, object)
	
	// Publish knowledge event
	agent.eventBus.Publish(NewCognitiveEvent(
		EventKnowledgeAcquired,
		"knowledge_graph",
		map[string]interface{}{
			"subject":   subject,
			"predicate": predicate,
			"object":    object,
		},
	))
}

// QueryKnowledge queries the knowledge graph
func (agent *AutonomousAgent) QueryKnowledge(subject string) []*Quad {
	return agent.knowledgeGraph.QueryBySubject(subject)
}

// SaveState persists a key-value pair to the state store
func (agent *AutonomousAgent) SaveState(key string, value interface{}) error {
	return agent.stateStore.Set(key, value)
}

// LoadState retrieves a value from the state store
func (agent *AutonomousAgent) LoadState(key string, dest interface{}) error {
	return agent.stateStore.Get(key, dest)
}

// GetGestalt returns the unified gestalt from all subsystems
func (agent *AutonomousAgent) GetGestalt() map[string]interface{} {
	agent.mu.RLock()
	defer agent.mu.RUnlock()
	
	gestalt := map[string]interface{}{
		"agent": map[string]interface{}{
			"id":            agent.agentID,
			"phase":         agent.currentPhase.String(),
			"autonomy":      agent.autonomyLevel,
			"running":       agent.running,
			"age":           time.Since(agent.birthTime).String(),
		},
		"knowledge_graph": agent.knowledgeGraph.ContributeToGestalt(),
		"state_store":     agent.stateStore.ContributeToGestalt(),
		"state_machine":   agent.stateMachine.ContributeToGestalt(),
	}
	
	return gestalt
}
