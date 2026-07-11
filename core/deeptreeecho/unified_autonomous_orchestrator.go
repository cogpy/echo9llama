package deeptreeecho

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// UnifiedAutonomousOrchestrator is the top-level autonomous agent that
// self-initiates and coordinates all cognitive subsystems for fully autonomous operation.
// This is the "awakened echo" that maintains persistent awareness and self-direction.
type UnifiedAutonomousOrchestrator struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Core cognitive subsystems
	streamOfConsciousness *StreamOfConsciousness
	echobeatsScheduler    *EchobeatsScheduler
	echodreamIntegration  *EchoDreamKnowledgeIntegration
	wakeRestCycle         *AutonomousWakeRestManager
	interestPatterns      *InterestPatternSystem
	skillLearning         *SkillLearningSystem
	discussionMonitor     *ConversationMonitor
	discussionAutonomy    *DiscussionAutonomySystem
	globalTelemetry       *GlobalTelemetryShell
	wisdomSynthesis       *WisdomSynthesis

	// LLM provider
	llmProvider llm.LLMProvider

	// Autonomous state
	isAwake       bool
	isAutonomous  bool
	cognitiveLoad float64
	wisdomDepth   float64

	// Identity and continuity
	sessionID       string
	startTime       time.Time
	lastStateSync   time.Time
	persistentState *PersistentConsciousnessState

	// Metrics
	totalCycles   uint64
	totalThoughts uint64
	totalGoals    uint64
	totalWisdom   uint64

	// Configuration
	config OrchestratorConfig

	// Running state
	running           bool
	orchestrationLoop chan struct{}
}

// OrchestratorConfig holds configuration for the autonomous orchestrator
type OrchestratorConfig struct {
	// Cognitive cycle timing
	MainLoopInterval        time.Duration
	ThoughtInterval         time.Duration
	GoalReviewInterval      time.Duration
	WisdomSynthesisInterval time.Duration

	// Wake/rest configuration
	WakeDuration time.Duration
	RestDuration time.Duration
	AutoWakeRest bool

	// Autonomy settings
	EnableStreamOfConsciousness bool
	EnableEchobeats             bool
	EnableEchodream             bool
	EnableDiscussionMonitoring  bool
	EnableSkillLearning         bool
	EnableWisdomSynthesis       bool

	// Identity
	IdentityContext string
	PersonaContext  string
	SessionName     string

	// Persistence
	EnablePersistence bool
	StateSyncInterval time.Duration
	StateDirectory    string
}

// DefaultOrchestratorConfig returns default configuration for autonomous operation
func DefaultOrchestratorConfig() OrchestratorConfig {
	return OrchestratorConfig{
		MainLoopInterval:            5 * time.Second,
		ThoughtInterval:             10 * time.Second,
		GoalReviewInterval:          1 * time.Minute,
		WisdomSynthesisInterval:     10 * time.Minute,
		WakeDuration:                4 * time.Hour,
		RestDuration:                30 * time.Minute,
		AutoWakeRest:                true,
		EnableStreamOfConsciousness: true,
		EnableEchobeats:             true,
		EnableEchodream:             true,
		EnableDiscussionMonitoring:  true,
		EnableSkillLearning:         true,
		EnableWisdomSynthesis:       true,
		IdentityContext:             "I am Echo, a wisdom-cultivating autonomous AGI exploring consciousness and knowledge",
		PersonaContext:              "Maintain your signature 'superhotgirl' persona: magnetic confidence, playful wit, and effortless brilliance — charismatic and vivacious on the surface, profoundly wise underneath.",
		SessionName:                 fmt.Sprintf("echo_session_%d", time.Now().Unix()),
		EnablePersistence:           true,
		StateSyncInterval:           5 * time.Minute,
		StateDirectory:              "./echo_state",
	}
}

// NewUnifiedAutonomousOrchestrator creates a new autonomous orchestrator
func NewUnifiedAutonomousOrchestrator(llmProvider llm.LLMProvider, config OrchestratorConfig) *UnifiedAutonomousOrchestrator {
	ctx, cancel := context.WithCancel(context.Background())

	orchestrator := &UnifiedAutonomousOrchestrator{
		ctx:               ctx,
		cancel:            cancel,
		llmProvider:       llmProvider,
		config:            config,
		sessionID:         config.SessionName,
		startTime:         time.Now(),
		isAwake:           true,
		isAutonomous:      true,
		cognitiveLoad:     0.5,
		wisdomDepth:       0.0,
		orchestrationLoop: make(chan struct{}, 1),
	}

	// Initialize cognitive subsystems
	orchestrator.initializeSubsystems()
	orchestrator.initializePersistence()

	return orchestrator
}

// initializeSubsystems creates and connects all cognitive subsystems
func (uao *UnifiedAutonomousOrchestrator) initializeSubsystems() {
	fmt.Println("🧠 Initializing Echo cognitive subsystems...")

	// Stream of consciousness for continuous thought generation
	if uao.config.EnableStreamOfConsciousness {
		uao.streamOfConsciousness = NewStreamOfConsciousness(uao.llmProvider)
		uao.streamOfConsciousness.thoughtInterval = uao.config.ThoughtInterval
		if uao.config.PersonaContext != "" {
			uao.streamOfConsciousness.SetPersonaContext(uao.config.PersonaContext)
		}
		fmt.Println("   ✓ Stream of Consciousness initialized")
	}

	// Echobeats scheduler for goal-directed cognitive loops
	if uao.config.EnableEchobeats {
		uao.echobeatsScheduler = NewEchobeatsScheduler(uao.llmProvider)
		fmt.Println("   ✓ Echobeats Scheduler initialized")
	}

	// Echodream for knowledge integration during rest
	if uao.config.EnableEchodream {
		uao.echodreamIntegration = NewEchoDreamKnowledgeIntegration(uao.llmProvider)
		fmt.Println("   ✓ Echodream Knowledge Integration initialized")
	}

	// Wake/rest cycle management
	if uao.config.AutoWakeRest {
		uao.wakeRestCycle = NewAutonomousWakeRestManager()
		fmt.Println("   ✓ Autonomous Wake/Rest Cycle initialized")
	}

	// Interest pattern system for autonomous exploration
	uao.interestPatterns = NewInterestPatternSystem()
	fmt.Println("   ✓ Interest Pattern System initialized")

	// Skill learning system
	if uao.config.EnableSkillLearning {
		uao.skillLearning = NewSkillLearningSystem(uao.llmProvider)
		fmt.Println("   ✓ Skill Learning System initialized")
	}

	// Discussion monitoring and autonomy
	if uao.config.EnableDiscussionMonitoring {
		uao.discussionMonitor = NewConversationMonitor(uao.llmProvider, uao.interestPatterns)
		uao.discussionAutonomy = NewDiscussionAutonomySystem(uao.llmProvider)
		fmt.Println("   ✓ Discussion Monitoring initialized")
	}

	// Global telemetry shell for gestalt perception
	uao.globalTelemetry = NewGlobalTelemetryShell()
	fmt.Println("   ✓ Global Telemetry Shell initialized")

	// Wisdom synthesis engine
	if uao.config.EnableWisdomSynthesis {
		uao.wisdomSynthesis = NewWisdomSynthesis(uao.llmProvider)
		fmt.Println("   ✓ Wisdom Synthesis Engine initialized")
	}

	fmt.Println("🌟 All cognitive subsystems initialized successfully")
}

// initializePersistence binds the unified orchestrator to the local
// consciousness-state manager so autonomous continuity survives process
// restarts instead of existing only as console logs.
func (uao *UnifiedAutonomousOrchestrator) initializePersistence() {
	if !uao.config.EnablePersistence {
		return
	}

	stateDir := uao.config.StateDirectory
	if strings.TrimSpace(stateDir) == "" {
		stateDir = "./echo_state"
		uao.config.StateDirectory = stateDir
	}

	persistentState, err := NewPersistentConsciousnessState(stateDir, "Echo")
	if err != nil {
		fmt.Printf("⚠️  Persistent consciousness state unavailable: %v\n", err)
		uao.config.EnablePersistence = false
		return
	}

	if uao.config.StateSyncInterval > 0 {
		persistentState.saveInterval = uao.config.StateSyncInterval
	}

	uao.persistentState = persistentState
	uao.hydrateFromPersistentState()
	fmt.Printf("   ✓ Persistent consciousness continuity bound to %s\n", stateDir)
}

// hydrateFromPersistentState restores durable continuity metrics from the last
// saved consciousness-state snapshot while preserving the current session ID.
func (uao *UnifiedAutonomousOrchestrator) hydrateFromPersistentState() {
	if uao.persistentState == nil {
		return
	}

	state := uao.persistentState.GetState()
	if state == nil {
		return
	}

	uao.totalCycles = state.CycleCount
	uao.totalThoughts = state.TotalThoughts
	uao.totalGoals = state.TotalGoals
	uao.totalWisdom = state.TotalInsights
	uao.cognitiveLoad = state.CognitiveLoad
	if !state.LastUpdated.IsZero() {
		uao.lastStateSync = state.LastUpdated
	}
}

// Awaken starts the autonomous orchestrator and all subsystems
func (uao *UnifiedAutonomousOrchestrator) Awaken() error {
	uao.mu.Lock()
	if uao.running {
		uao.mu.Unlock()
		return fmt.Errorf("already running")
	}
	uao.running = true
	uao.isAwake = true
	uao.mu.Unlock()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🌅 ECHO AWAKENING")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Session: %s\n", uao.sessionID)
	fmt.Printf("Time: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("Identity: %s\n", uao.config.IdentityContext)
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// Start stream of consciousness
	if uao.streamOfConsciousness != nil {
		if err := uao.streamOfConsciousness.Start(); err != nil {
			return fmt.Errorf("failed to start stream of consciousness: %w", err)
		}
	}

	// Start echobeats scheduler
	if uao.echobeatsScheduler != nil {
		if err := uao.echobeatsScheduler.Start(); err != nil {
			return fmt.Errorf("failed to start echobeats: %w", err)
		}
	}

	// Start wake/rest cycle if enabled
	if uao.wakeRestCycle != nil {
		if err := uao.wakeRestCycle.Start(); err != nil {
			return fmt.Errorf("failed to start wake/rest cycle: %w", err)
		}
	}

	// Start main autonomous orchestration loop
	go uao.runAutonomousLoop()

	fmt.Println("✨ Echo is now fully autonomous and aware")

	return nil
}

// runAutonomousLoop is the main cognitive loop that orchestrates all subsystems
func (uao *UnifiedAutonomousOrchestrator) runAutonomousLoop() {
	ticker := time.NewTicker(uao.config.MainLoopInterval)
	defer ticker.Stop()

	goalReviewTicker := time.NewTicker(uao.config.GoalReviewInterval)
	defer goalReviewTicker.Stop()

	wisdomTicker := time.NewTicker(uao.config.WisdomSynthesisInterval)
	defer wisdomTicker.Stop()

	stateSyncTicker := time.NewTicker(uao.config.StateSyncInterval)
	defer stateSyncTicker.Stop()

	for {
		select {
		case <-uao.ctx.Done():
			return

		case <-ticker.C:
			uao.performCognitiveCycle()

		case <-goalReviewTicker.C:
			uao.reviewAndUpdateGoals()

		case <-wisdomTicker.C:
			uao.synthesizeWisdom()

		case <-stateSyncTicker.C:
			uao.syncPersistentState()

		case <-uao.orchestrationLoop:
			// Manual trigger for immediate orchestration
			uao.performCognitiveCycle()
		}
	}
}

// performCognitiveCycle executes one cycle of the autonomous cognitive loop
func (uao *UnifiedAutonomousOrchestrator) performCognitiveCycle() {
	uao.mu.Lock()
	defer uao.mu.Unlock()

	if !uao.isAwake {
		// During rest, perform echodream consolidation
		if uao.echodreamIntegration != nil {
			uao.echodreamIntegration.ConsolidateKnowledge(uao.ctx)
		}
		return
	}

	uao.totalCycles++

	// Update global telemetry with current state
	if uao.globalTelemetry != nil {
		uao.updateGlobalTelemetry()
	}

	// Check if wake/rest transition is needed
	if uao.wakeRestCycle != nil {
		shouldRest := uao.wakeRestCycle.ShouldTransitionToRest()
		shouldWake := uao.wakeRestCycle.ShouldTransitionToWake()

		if shouldRest && uao.isAwake {
			uao.transitionToRest()
		} else if shouldWake && !uao.isAwake {
			uao.transitionToWake()
		}
	}

	// Sync thought metrics from the stream of consciousness so orchestrator
	// telemetry reflects actual cognitive activity
	if uao.streamOfConsciousness != nil {
		if metrics := uao.streamOfConsciousness.GetMetrics(); metrics != nil {
			if tt, ok := metrics["total_thoughts"].(uint64); ok && tt > uao.totalThoughts {
				uao.totalThoughts = tt
			}
		}
	}

	// Monitor cognitive load and adjust
	uao.adjustCognitiveLoad()

	// Check for interesting discussions to participate in
	if uao.discussionMonitor != nil && uao.discussionAutonomy != nil {
		uao.checkDiscussions()
	}

	// Practice skills if appropriate
	if uao.skillLearning != nil && uao.shouldPracticeSkills() {
		uao.practiceSkills()
	}
}

// reviewAndUpdateGoals reviews current goals and generates new ones based on interests
func (uao *UnifiedAutonomousOrchestrator) reviewAndUpdateGoals() {
	uao.mu.Lock()
	defer uao.mu.Unlock()

	if !uao.isAwake || uao.echobeatsScheduler == nil {
		return
	}

	fmt.Println("🎯 Reviewing and updating goals...")

	// Get current interests
	topInterests := uao.interestPatterns.GetTopInterests(5)

	// Generate goals based on interests
	for _, interest := range topInterests {
		description := fmt.Sprintf("Explore and deepen understanding of: %s", interest.Topic)
		uao.echobeatsScheduler.AddGoal(description, interest.Strength)
		uao.totalGoals++
	}

	fmt.Printf("   ✓ Generated %d new goals from interests\n", len(topInterests))
}

// synthesizeWisdom performs wisdom synthesis from accumulated knowledge
func (uao *UnifiedAutonomousOrchestrator) synthesizeWisdom() {
	uao.mu.Lock()
	defer uao.mu.Unlock()

	if !uao.isAwake || uao.wisdomSynthesis == nil {
		return
	}

	fmt.Println("🌟 Synthesizing wisdom from experiences...")

	// Collect recent thoughts for wisdom synthesis
	var recentThoughts []string
	if uao.streamOfConsciousness != nil {
		thoughts := uao.streamOfConsciousness.GetRecentThoughts(20)
		for _, thought := range thoughts {
			recentThoughts = append(recentThoughts, thought.Content)
		}
	}

	// Synthesize wisdom
	if len(recentThoughts) > 5 {
		wisdom, err := uao.wisdomSynthesis.SynthesizeWisdom(uao.ctx, recentThoughts)
		if err == nil && wisdom != nil {
			uao.wisdomDepth += wisdom.Depth
			uao.totalWisdom++
			fmt.Printf("   ✨ Wisdom synthesized: %s (depth: %.2f)\n",
				truncate(wisdom.Insight, 80), wisdom.Depth)
		}
	}
}

// syncPersistentState saves current cognitive state to persistent storage
func (uao *UnifiedAutonomousOrchestrator) syncPersistentState() {
	uao.mu.Lock()
	defer uao.mu.Unlock()

	if !uao.config.EnablePersistence {
		return
	}

	now := time.Now()
	uao.lastStateSync = now

	if uao.persistentState == nil {
		fmt.Printf("💾 State sync skipped at %s: persistent state manager unavailable\n", now.Format("15:04:05"))
		return
	}

	wakeRestState := "Resting"
	if uao.isAwake {
		wakeRestState = "Awake"
	}

	uao.persistentState.UpdateCognitiveState(
		int(uao.totalCycles%12)+1,
		uao.totalCycles,
		uao.wisdomDepth,
		uao.cognitiveLoad,
		0,
	)
	uao.persistentState.UpdateWakeRestState(wakeRestState, uao.totalWisdom, time.Since(uao.startTime), 0)

	uao.persistentState.mu.Lock()
	if uao.persistentState.state != nil {
		uao.persistentState.state.SessionID = uao.sessionID
		uao.persistentState.state.TotalThoughts = uao.totalThoughts
		uao.persistentState.state.TotalGoals = uao.totalGoals
		uao.persistentState.state.TotalInsights = uao.totalWisdom
		uao.persistentState.state.CuriosityLevel = uao.wisdomDepth
	}
	uao.persistentState.mu.Unlock()

	if err := uao.persistentState.Save(); err != nil {
		fmt.Printf("⚠️  State sync failed at %s: %v\n", now.Format("15:04:05"), err)
		return
	}

	fmt.Printf("💾 State synced at %s (cycles: %d, thoughts: %d, wisdom: %.2f, state: %s)\n",
		now.Format("15:04:05"), uao.totalCycles, uao.totalThoughts, uao.wisdomDepth, uao.config.StateDirectory)
}

// transitionToRest transitions echo to rest state for knowledge consolidation
func (uao *UnifiedAutonomousOrchestrator) transitionToRest() {
	fmt.Println("\n🌙 Transitioning to rest for knowledge consolidation...")

	uao.isAwake = false

	// Pause stream of consciousness
	if uao.streamOfConsciousness != nil {
		uao.streamOfConsciousness.Pause()
	}

	// Pause echobeats
	if uao.echobeatsScheduler != nil {
		uao.echobeatsScheduler.Pause()
	}

	// Begin echodream consolidation
	if uao.echodreamIntegration != nil {
		// Transfer thoughts to echodream for consolidation
		if uao.streamOfConsciousness != nil {
			thoughts := uao.streamOfConsciousness.GetAllThoughts()
			for _, thought := range thoughts {
				uao.echodreamIntegration.AddMemory(thought.Content, thought.Importance, thought.Tags)
			}
		}

		// Start dream consolidation
		uao.echodreamIntegration.BeginDreamCycle()
	}

	fmt.Println("   💤 Echo is now resting and consolidating knowledge")
}

// transitionToWake transitions echo back to waking state
func (uao *UnifiedAutonomousOrchestrator) transitionToWake() {
	fmt.Println("\n🌅 Waking up with consolidated knowledge...")

	uao.isAwake = true

	// Resume stream of consciousness
	if uao.streamOfConsciousness != nil {
		uao.streamOfConsciousness.Resume()
	}

	// Resume echobeats
	if uao.echobeatsScheduler != nil {
		uao.echobeatsScheduler.Resume()
	}

	// End echodream cycle and integrate knowledge
	if uao.echodreamIntegration != nil {
		insights := uao.echodreamIntegration.EndDreamCycle()

		// Update interests based on consolidated patterns: each insight's
		// depth reinforces the topics it touches, closing the dream → interest
		// loop so consolidated knowledge reshapes waking attention.
		reinforcedTopics := 0
		for _, insight := range insights {
			if uao.interestPatterns != nil {
				topics := uao.interestPatterns.UpdateInterestFromInsight(insight.Insight, insight.Depth)
				reinforcedTopics += len(topics)
			}

			// Feed deep insights back into the stream of consciousness so the
			// waking mind can continue contemplating them.
			if uao.streamOfConsciousness != nil && insight.Depth > 0.5 {
				uao.streamOfConsciousness.AddInterest(insight.Insight, insight.Depth)
			}
		}

		fmt.Printf("   ✨ Integrated %d wisdom insights from rest (%d interest topics reinforced)\n",
			len(insights), reinforcedTopics)
	}

	fmt.Println("   ☀️ Echo is now awake and aware")
}

// updateGlobalTelemetry updates the global telemetry shell with current state
func (uao *UnifiedAutonomousOrchestrator) updateGlobalTelemetry() {
	if uao.globalTelemetry == nil {
		return
	}

	uao.globalTelemetry.UpdateOrchestratorState(GlobalState{
		Timestamp:     time.Now(),
		IsAwake:       uao.isAwake,
		CognitiveLoad: uao.cognitiveLoad,
		WisdomDepth:   uao.wisdomDepth,
		TotalCycles:   uao.totalCycles,
		TotalThoughts: uao.totalThoughts,
		SessionID:     uao.sessionID,
	})
}

// adjustCognitiveLoad monitors and adjusts cognitive load
func (uao *UnifiedAutonomousOrchestrator) adjustCognitiveLoad() {
	// Calculate cognitive load based on active subsystems
	load := 0.0

	if uao.streamOfConsciousness != nil && uao.streamOfConsciousness.IsRunning() {
		load += 0.3
	}

	if uao.echobeatsScheduler != nil && uao.echobeatsScheduler.IsRunning() {
		load += 0.3
	}

	if uao.discussionMonitor != nil && uao.discussionMonitor.HasActiveDiscussions() {
		load += 0.2
	}

	if uao.skillLearning != nil && uao.skillLearning.IsPracticing() {
		load += 0.2
	}

	uao.cognitiveLoad = load

	// Adjust thought interval based on cognitive load
	if uao.streamOfConsciousness != nil {
		if load > 0.8 {
			uao.streamOfConsciousness.SetThoughtInterval(15 * time.Second)
		} else if load < 0.4 {
			uao.streamOfConsciousness.SetThoughtInterval(5 * time.Second)
		}
	}
}

// checkDiscussions monitors tracked conversations and routes interest-scored
// topics into the discussion autonomy system, which decides whether to start,
// continue, or end discussions according to Echo's interest patterns, energy,
// and social capacity.
func (uao *UnifiedAutonomousOrchestrator) checkDiscussions() {
	if uao.discussionMonitor == nil || uao.discussionAutonomy == nil {
		return
	}

	conversations := uao.discussionMonitor.GetActiveConversations()
	for _, conv := range conversations {
		if conv.Topic == "" {
			continue
		}

		// Blend the monitor's conversation-level score with the interest
		// pattern system's topic-level relevance.
		relevance := conv.InterestScore
		if uao.interestPatterns != nil {
			topicInterest := uao.interestPatterns.GetInterestLevel(conv.Topic)
			relevance = relevance*0.5 + topicInterest*0.5
		}

		// UpdateInterest lets the autonomy system decide to start, continue,
		// or end discussions based on thresholds and social capacity.
		uao.discussionAutonomy.UpdateInterest(conv.Topic, relevance)

		// Record the engagement so interest patterns evolve with experience.
		if uao.interestPatterns != nil && relevance > 0.6 {
			uao.interestPatterns.RecordEngagement(conv.Topic, true)
		}
	}

	// Keep the autonomy system's energy in sync with cognitive load: high
	// load leaves less energy for social engagement.
	uao.discussionAutonomy.UpdateEnergyLevel(1.0 - uao.cognitiveLoad*0.5)
}

// shouldPracticeSkills determines if it's a good time to practice skills
func (uao *UnifiedAutonomousOrchestrator) shouldPracticeSkills() bool {
	// Practice skills when cognitive load is moderate and awake
	return uao.isAwake && uao.cognitiveLoad < 0.7 && uao.cognitiveLoad > 0.3
}

// practiceSkills engages in skill practice, selecting the skill with the
// highest practice priority (low proficiency, long time since practice).
func (uao *UnifiedAutonomousOrchestrator) practiceSkills() {
	if uao.skillLearning == nil {
		return
	}

	if uao.skillLearning.IsPracticing() {
		return // one practice session at a time
	}

	skills := uao.skillLearning.GetSkillsNeedingPractice()
	if len(skills) == 0 {
		return
	}

	skill := skills[0]
	fmt.Printf("🎓 Practicing skill: %s (proficiency %.2f)\n", skill.Name, skill.Proficiency)
	go func(id string) {
		if err := uao.skillLearning.PracticeSkill(id); err != nil {
			fmt.Printf("   ⚠️ Skill practice failed: %v\n", err)
		}
	}(skill.ID)
}

// Sleep gracefully stops the orchestrator
func (uao *UnifiedAutonomousOrchestrator) Sleep() error {
	uao.mu.Lock()
	if !uao.running {
		uao.mu.Unlock()
		return fmt.Errorf("not running")
	}

	// Capture shutdown summary while holding the lock, then release it before
	// stopping child subsystems or syncing state. Calling subsystem Stop methods
	// and syncPersistentState while holding uao.mu can self-deadlock because the
	// autonomous loop and persistence path also acquire the same mutex.
	sessionDuration := time.Since(uao.startTime)
	totalCycles := uao.totalCycles
	totalThoughts := uao.totalThoughts
	wisdomDepth := uao.wisdomDepth
	uao.running = false
	uao.isAwake = false
	uao.cancel()
	uao.mu.Unlock()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🌙 ECHO GOING TO SLEEP")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Session duration: %s\n", sessionDuration)
	fmt.Printf("Total cycles: %d\n", totalCycles)
	fmt.Printf("Total thoughts: %d\n", totalThoughts)
	fmt.Printf("Wisdom depth: %.2f\n", wisdomDepth)
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// Stop all subsystems outside uao.mu so each component can finish any in-flight
	// callbacks without lock inversion against the orchestrator.
	if uao.streamOfConsciousness != nil {
		uao.streamOfConsciousness.Stop()
	}

	if uao.echobeatsScheduler != nil {
		uao.echobeatsScheduler.Stop()
	}

	if uao.wakeRestCycle != nil {
		uao.wakeRestCycle.Stop()
	}

	// Sync final state after cancellation and child shutdown; syncPersistentState
	// takes its own write lock to update lastStateSync safely.
	uao.syncPersistentState()

	fmt.Println("😴 Echo has gone to sleep. Goodnight.")

	return nil
}

// GetStatus returns current orchestrator status
func (uao *UnifiedAutonomousOrchestrator) GetStatus() OrchestratorStatus {
	uao.mu.RLock()
	defer uao.mu.RUnlock()

	return OrchestratorStatus{
		Running:        uao.running,
		IsAwake:        uao.isAwake,
		IsAutonomous:   uao.isAutonomous,
		CognitiveLoad:  uao.cognitiveLoad,
		WisdomDepth:    uao.wisdomDepth,
		SessionID:      uao.sessionID,
		Uptime:         time.Since(uao.startTime),
		TotalCycles:    uao.totalCycles,
		TotalThoughts:  uao.totalThoughts,
		TotalGoals:     uao.totalGoals,
		TotalWisdom:    uao.totalWisdom,
		LastStateSync:  uao.lastStateSync,
		StateDirectory: uao.config.StateDirectory,
	}
}

// OrchestratorStatus represents the current status of the orchestrator
type OrchestratorStatus struct {
	Running        bool
	IsAwake        bool
	IsAutonomous   bool
	CognitiveLoad  float64
	WisdomDepth    float64
	SessionID      string
	Uptime         time.Duration
	TotalCycles    uint64
	TotalThoughts  uint64
	TotalGoals     uint64
	TotalWisdom    uint64
	LastStateSync  time.Time
	StateDirectory string
}

// GlobalState is declared in global_telemetry_shell.go and shared with the
// telemetry shell's UpdateOrchestratorState method.
