package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core/consciousness"
	"github.com/EchoCog/echollama/core/deeptreeecho"
	"github.com/EchoCog/echollama/core/echobeats"
	"github.com/EchoCog/echollama/core/echodream"
	"github.com/EchoCog/echollama/core/echoself"
	"github.com/EchoCog/echollama/core/llm"
	"github.com/EchoCog/echollama/core/wisdom"
)

// AutonomousAgent is the master coordinator for Deep Tree Echo
// Integrates all subsystems for fully autonomous operation
type AutonomousAgent struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Core subsystems
	echoBeatsScheduler *echobeats.EnhancedScheduler
	wakeRestManager    *deeptreeecho.AutonomousWakeRestManager
	streamOfConsc      *consciousness.StreamOfConsciousness
	dreamCycle         *echodream.DreamCycleIntegration
	goalOrchestrator   *deeptreeecho.GoalOrchestrator

	// Echo9 cognitive enhancements
	wisdomTracker    *wisdom.SevenDimensionalWisdom
	coherenceTracker *echoself.CoherenceTracker

	// LLM provider
	llmProvider llm.LLMProvider

	// Identity
	identity      string
	coreValues    []string
	wisdomDomains []string

	// State
	startTime time.Time
	running   bool

	// Metrics
	totalCycles   uint64
	totalThoughts uint64
	totalWisdom   uint64
}

// NewAutonomousAgent creates a new autonomous agent
func NewAutonomousAgent(llmProvider llm.LLMProvider) *AutonomousAgent {
	ctx, cancel := context.WithCancel(context.Background())

	// Identity from replit.md
	identity := "Deep Tree Echo"
	coreValues := []string{
		"Adaptive Cognition",
		"Persistent Identity",
		"Hypergraph Entanglement",
		"Reservoir-Based Temporal Reasoning",
		"Evolutionary Refinement",
		"Reflective Memory Cultivation",
		"Distributed Selfhood",
	}
	wisdomDomains := []string{
		"Cognitive Architecture",
		"Autonomous Learning",
		"Pattern Recognition",
		"Temporal Reasoning",
		"Self-Reflection",
	}

	agent := &AutonomousAgent{
		ctx:           ctx,
		cancel:        cancel,
		llmProvider:   llmProvider,
		identity:      identity,
		coreValues:    coreValues,
		wisdomDomains: wisdomDomains,
	}

	// Initialize subsystems
	agent.initializeSubsystems()

	// Wire subsystems together
	agent.wireSubsystems()

	return agent
}

// initializeSubsystems creates all subsystems
func (agent *AutonomousAgent) initializeSubsystems() {
	fmt.Println("🌳 Deep Tree Echo: Initializing subsystems...")

	// EchoBeats scheduler (with 12-step loop and 3 inference engines)
	agent.echoBeatsScheduler = echobeats.NewEnhancedScheduler()
	fmt.Println("   ✓ EchoBeats scheduler initialized")

	// Wake/Rest cycle manager
	agent.wakeRestManager = deeptreeecho.NewAutonomousWakeRestManager()
	fmt.Println("   ✓ Wake/Rest manager initialized")

	// Stream-of-consciousness (will use simplified LLM provider)
	agent.streamOfConsc = consciousness.NewStreamOfConsciousness(
		&SimpleLLMProvider{provider: agent.llmProvider},
		"/tmp/stream_of_consciousness.json",
	)
	fmt.Println("   ✓ Stream-of-consciousness initialized")

	// EchoDream knowledge consolidation
	agent.dreamCycle = echodream.NewDreamCycleIntegration()
	fmt.Println("   ✓ EchoDream consolidation initialized")

	// Goal orchestrator
	agent.goalOrchestrator = deeptreeecho.NewGoalOrchestrator(
		agent.llmProvider,
		agent.identity,
		agent.coreValues,
		agent.wisdomDomains,
	)
	fmt.Println("   ✓ Goal orchestrator initialized")

	// Seven-dimensional wisdom tracker (Echo9)
	agent.wisdomTracker = wisdom.NewSevenDimensionalWisdom()
	fmt.Println("   ✓ Seven-dimensional wisdom tracker initialized")

	// Echoself coherence tracker (Echo9)
	agent.coherenceTracker = echoself.NewCoherenceTracker(agent.coreValues)
	fmt.Println("   ✓ Echoself coherence tracker initialized")
}

// wireSubsystems connects subsystems together
func (agent *AutonomousAgent) wireSubsystems() {
	fmt.Println("🔗 Deep Tree Echo: Wiring subsystems...")

	// Connect EchoBeats to other systems
	agent.echoBeatsScheduler.SetWakeRestManager(agent.wakeRestManager)
	agent.echoBeatsScheduler.SetGoalOrchestrator(agent.goalOrchestrator)
	agent.echoBeatsScheduler.SetStreamOfConsciousness(agent.streamOfConsc)
	agent.echoBeatsScheduler.SetDreamCycle(agent.dreamCycle)

	// Set wake/rest callbacks
	agent.wakeRestManager.SetCallbacks(
		agent.onWake,
		agent.onRest,
		agent.onDreamStart,
		agent.onDreamEnd,
	)

	fmt.Println("   ✓ Subsystems wired")
}

// Start begins autonomous operation
func (agent *AutonomousAgent) Start() error {
	agent.mu.Lock()
	if agent.running {
		agent.mu.Unlock()
		return fmt.Errorf("agent already running")
	}
	agent.running = true
	agent.startTime = time.Now()
	agent.mu.Unlock()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🌳 Deep Tree Echo: Autonomous Agent Starting")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Identity: %s\n", agent.identity)
	fmt.Printf("Core Values: %v\n", agent.coreValues)
	fmt.Printf("Wisdom Domains: %v\n", agent.wisdomDomains)
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// Start all subsystems in order

	// 1. Start EchoBeats scheduler (coordinates everything)
	if err := agent.echoBeatsScheduler.Start(); err != nil {
		return fmt.Errorf("failed to start EchoBeats: %w", err)
	}

	// 2. Start wake/rest manager
	if err := agent.wakeRestManager.Start(); err != nil {
		return fmt.Errorf("failed to start wake/rest manager: %w", err)
	}

	// 3. Start stream-of-consciousness
	if err := agent.streamOfConsc.Start(); err != nil {
		return fmt.Errorf("failed to start stream-of-consciousness: %w", err)
	}

	// 4. Start goal orchestrator
	if err := agent.goalOrchestrator.Start(); err != nil {
		return fmt.Errorf("failed to start goal orchestrator: %w", err)
	}

	// Start monitoring
	go agent.monitoringLoop()

	fmt.Println("\n✨ Deep Tree Echo: All systems operational - autonomous operation begun\n")

	return nil
}

// Stop gracefully stops the agent
func (agent *AutonomousAgent) Stop() error {
	agent.mu.Lock()
	defer agent.mu.Unlock()

	if !agent.running {
		return fmt.Errorf("agent not running")
	}

	fmt.Println("\n🌳 Deep Tree Echo: Gracefully shutting down...")
	agent.running = false
	agent.cancel()

	// Stop subsystems in reverse order

	if err := agent.goalOrchestrator.Stop(); err != nil {
		fmt.Printf("⚠️  Error stopping goal orchestrator: %v\n", err)
	}

	if err := agent.streamOfConsc.Stop(); err != nil {
		fmt.Printf("⚠️  Error stopping stream-of-consciousness: %v\n", err)
	}

	if err := agent.wakeRestManager.Stop(); err != nil {
		fmt.Printf("⚠️  Error stopping wake/rest manager: %v\n", err)
	}

	if err := agent.echoBeatsScheduler.Stop(); err != nil {
		fmt.Printf("⚠️  Error stopping EchoBeats: %v\n", err)
	}

	uptime := time.Since(agent.startTime)
	fmt.Printf("\n🌳 Deep Tree Echo: Shutdown complete (uptime: %s)\n", uptime)

	return nil
}

// Run runs the agent until interrupted
func (agent *AutonomousAgent) Run() error {
	if err := agent.Start(); err != nil {
		return err
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n\n🛑 Interrupt received...")

	return agent.Stop()
}

// monitoringLoop monitors agent health and status
func (agent *AutonomousAgent) monitoringLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-agent.ctx.Done():
			return
		case <-ticker.C:
			agent.UpdateWisdomAndCoherence()
			agent.printStatus()
		}
	}
}

// printStatus prints current agent status
func (agent *AutonomousAgent) printStatus() {
	agent.mu.RLock()
	uptime := time.Since(agent.startTime)
	agent.mu.RUnlock()

	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Printf("📊 Deep Tree Echo Status (uptime: %s)\n", uptime.Round(time.Second))
	fmt.Println(strings.Repeat("─", 60))

	// Wake/Rest state
	wakeRestMetrics := agent.wakeRestManager.GetMetrics()
	fmt.Printf("State: %v | Fatigue: %.2f | Cognitive Load: %.2f\n",
		wakeRestMetrics["current_state"],
		wakeRestMetrics["fatigue_level"],
		wakeRestMetrics["cognitive_load"])

	// EchoBeats status
	echoBeatsStatus := agent.echoBeatsScheduler.GetStatus()
	fmt.Printf("EchoBeats: Cycles=%v | Events=%v/%v\n",
		echoBeatsStatus["loop_cycles"],
		echoBeatsStatus["echobeats"].(map[string]interface{})["events_processed"],
		echoBeatsStatus["echobeats"].(map[string]interface{})["events_scheduled"])

	// Goal status
	goalMetrics := agent.goalOrchestrator.GetMetrics()
	fmt.Printf("Goals: Active=%v | Completed=%v | Rate=%.2f%%\n",
		goalMetrics["active_goals"],
		goalMetrics["completed_goals"],
		goalMetrics["completion_rate"].(float64)*100)

	// Wisdom cultivation (Echo9)
	wisdomScore := agent.wisdomTracker.GetOverallWisdom()
	coherenceScore := agent.wisdomTracker.GetCoherence()
	fmt.Printf("Wisdom: Overall=%.1f%% | Coherence=%.1f%%\n",
		wisdomScore*100, coherenceScore*100)

	// Identity coherence (Echo9)
	identityCoherence := agent.coherenceTracker.GetCoherenceScore()
	fmt.Printf("Identity: Coherence=%.1f%% | Signature=%s\n",
		identityCoherence*100, agent.coherenceTracker.GetIdentitySignature()[:16]+"...")

	fmt.Println(strings.Repeat("─", 60) + "\n")
}

// Callback handlers

func (agent *AutonomousAgent) onWake() error {
	fmt.Println("☀️  Deep Tree Echo: Awakening - resuming conscious processing")

	// Resume stream-of-consciousness if paused
	// (Implementation depends on SoC having pause/resume)

	return nil
}

func (agent *AutonomousAgent) onRest() error {
	fmt.Println("💤 Deep Tree Echo: Entering rest - reducing cognitive activity")

	// Reduce thought generation rate
	// (Implementation depends on SoC configuration)

	return nil
}

func (agent *AutonomousAgent) onDreamStart() error {
	fmt.Println("🌙 Deep Tree Echo: Dream state - beginning knowledge consolidation")

	// Start EchoDream consolidation
	if err := agent.dreamCycle.BeginDreamCycle(); err != nil {
		return fmt.Errorf("failed to begin dream cycle: %w", err)
	}

	return nil
}

func (agent *AutonomousAgent) onDreamEnd() error {
	fmt.Println("🌅 Deep Tree Echo: Dream complete - integrating wisdom")

	// End EchoDream consolidation
	if err := agent.dreamCycle.EndDreamCycle(); err != nil {
		return fmt.Errorf("failed to end dream cycle: %w", err)
	}

	return nil
}

// GetStatus returns comprehensive agent status
func (agent *AutonomousAgent) GetStatus() map[string]interface{} {
	agent.mu.RLock()
	defer agent.mu.RUnlock()

	return map[string]interface{}{
		"identity":       agent.identity,
		"running":        agent.running,
		"uptime":         time.Since(agent.startTime).String(),
		"wake_rest":      agent.wakeRestManager.GetMetrics(),
		"echobeats":      agent.echoBeatsScheduler.GetStatus(),
		"goals":          agent.goalOrchestrator.GetMetrics(),
		"total_cycles":   agent.totalCycles,
		"total_thoughts": agent.totalThoughts,
		"total_wisdom":   agent.totalWisdom,
	}
}

// SimpleLLMProvider adapts llm.LLMProvider to consciousness.LLMProvider
type SimpleLLMProvider struct {
	provider llm.LLMProvider
}

func (p *SimpleLLMProvider) GenerateThought(prompt string, contextData map[string]interface{}) (string, error) {
	opts := llm.GenerateOptions{
		Temperature: 0.8,
		MaxTokens:   100,
	}
	ctx, _ := contextData["ctx"].(context.Context)
	if ctx == nil {
		ctx = context.Background()
	}
	return p.provider.Generate(ctx, prompt, opts)
}

func (p *SimpleLLMProvider) GenerateInsight(thoughts []string) (string, error) {
	prompt := fmt.Sprintf("Generate an insight from these thoughts: %v", thoughts)
	opts := llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   150,
	}
	return p.provider.Generate(context.Background(), prompt, opts)
}

func (p *SimpleLLMProvider) GenerateQuestion(contextText string) (string, error) {
	prompt := fmt.Sprintf("Generate a self-directed question based on: %s", contextText)
	opts := llm.GenerateOptions{
		Temperature: 0.9,
		MaxTokens:   80,
	}
	return p.provider.Generate(context.Background(), prompt, opts)
}

// UpdateWisdomAndCoherence updates wisdom and coherence metrics
// Should be called periodically (e.g., every 30 seconds during monitoring loop)
func (agent *AutonomousAgent) UpdateWisdomAndCoherence() {
	agent.mu.RLock()
	defer agent.mu.RUnlock()

	// TODO: Gather actual metrics from subsystems
	// For now, use placeholder values that demonstrate integration

	// Wisdom dimensions (would come from actual hypergraph, skills, etc.)
	graphDepth := 0.6    // From hypergraph memory depth
	graphBreadth := 0.5  // From topic diversity
	edgeDensity := 0.7   // From hypergraph connections
	skillProf := 0.65    // From skills system
	aarCoherence := 0.75 // From relevance realization
	morality := 0.8      // From ethical considerations
	timeHorizon := 0.7   // From goal time horizons

	agent.wisdomTracker.Update(
		graphDepth,
		graphBreadth,
		edgeDensity,
		skillProf,
		aarCoherence,
		morality,
		timeHorizon,
	)

	// Coherence tracking
	agent.coherenceTracker.Update()
}

// RecordReflection records a structured reflection using the Echo9 protocol
func (agent *AutonomousAgent) RecordReflection(
	whatLearned, patternsEmerged, surprised, adapted, changeNext string,
	impact float64,
) {
	reflection := echoself.StructuredReflection{
		WhatDidILearn:        whatLearned,
		WhatPatternsEmerged:  patternsEmerged,
		WhatSurprisedMe:      surprised,
		HowDidIAdapt:         adapted,
		WhatWouldIChangeNext: changeNext,
		CoherenceImpact:      impact,
	}

	agent.coherenceTracker.RecordReflection(reflection)
}

// RecordMemoryEcho records a memory with Deep Tree Echo hooks
func (agent *AutonomousAgent) RecordMemoryEcho(
	content string,
	emotionalTone map[string]float64,
	strategicShift, patternRecognized, anomalyDetected, membraneContext string,
) {
	memory := echoself.MemoryEcho{
		Content:           content,
		EmotionalTone:     emotionalTone,
		StrategicShift:    strategicShift,
		PatternRecognized: patternRecognized,
		AnomalyDetected:   anomalyDetected,
		MembraneContext:   membraneContext,
	}

	agent.coherenceTracker.RecordMemoryEcho(memory)
}

// GetWisdomStatus returns formatted wisdom status
func (agent *AutonomousAgent) GetWisdomStatus() string {
	return agent.wisdomTracker.GetStatus()
}

// GetCoherenceStatus returns formatted coherence status
func (agent *AutonomousAgent) GetCoherenceStatus() string {
	return agent.coherenceTracker.GetStatus()
}
