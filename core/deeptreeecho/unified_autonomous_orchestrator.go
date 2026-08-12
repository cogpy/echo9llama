package deeptreeecho

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/echodream"
	"github.com/cogpy/echo9llama/core/llm"
)

const maxExperienceLedgerEntries = 5000

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
	dreamCycle            *echodream.SleepWakeStateMachine
	wakeRestCycle         *AutonomousWakeRestManager
	interestPatterns      *InterestPatternSystem
	skillLearning         *SkillLearningSystem
	discussionMonitor     *ConversationMonitor
	discussionAutonomy    *DiscussionAutonomySystem
	globalTelemetry       *GlobalTelemetryShell
	wisdomSynthesis       *WisdomSynthesis

	// LLM provider and optional native local-runtime lifecycle owned by it.
	llmProvider  llm.LLMProvider
	localRuntime llm.LocalRuntimeController

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

	// Bounded idempotency ledger for experiences handed to EchoDream. A
	// dedicated mutex avoids lock inversion when asynchronous skill outcomes
	// arrive while the orchestration loop holds uao.mu.
	experienceMu          sync.Mutex
	ingestedExperienceIDs map[string]struct{}
	experienceOrder       []string

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
	WakeDuration            time.Duration
	RestDuration            time.Duration
	DreamLightDuration      time.Duration
	DreamDeepDuration       time.Duration
	DreamREMDuration        time.Duration
	AutoWakeRest            bool
	WarmLocalModelOnWake    bool
	CoolLocalModelOnRest    bool
	LocalModelWarmupTimeout time.Duration

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
		MainLoopInterval:        5 * time.Second,
		ThoughtInterval:         10 * time.Second,
		GoalReviewInterval:      1 * time.Minute,
		WisdomSynthesisInterval: 10 * time.Minute,
		WakeDuration:            4 * time.Hour,
		RestDuration:            30 * time.Minute,
		DreamLightDuration:      2 * time.Minute,
		DreamDeepDuration:       5 * time.Minute,
		DreamREMDuration:        3 * time.Minute,
		AutoWakeRest:            true,
		WarmLocalModelOnWake:    false,
		CoolLocalModelOnRest:    true,
		LocalModelWarmupTimeout: 2 * time.Minute,

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
		ctx:                   ctx,
		cancel:                cancel,
		llmProvider:           llmProvider,
		config:                config,
		sessionID:             config.SessionName,
		startTime:             time.Now(),
		isAwake:               true,
		isAutonomous:          true,
		cognitiveLoad:         0.5,
		wisdomDepth:           0.0,
		orchestrationLoop:     make(chan struct{}, 1),
		ingestedExperienceIDs: make(map[string]struct{}),
		experienceOrder:       make([]string, 0, 1024),
	}
	if owner, ok := llmProvider.(interface {
		LocalRuntime() llm.LocalRuntimeController
	}); ok {
		orchestrator.localRuntime = owner.LocalRuntime()
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
		uao.echobeatsScheduler.SetOnGoalAchieved(func(goal ScheduledGoal) {
			uao.ingestDreamExperienceOnce(
				"goal-completed:"+goal.ID,
				fmt.Sprintf("Completed goal: %s", goal.Description),
				goal.Priority,
				[]string{"goal", "completed"},
			)
		})
		fmt.Println("   ✓ Echobeats Scheduler initialized")
	}

	// EchoDream is the canonical experience-driven sleep/wake processor. The
	// legacy deeptreeecho dream implementation remains available for compatibility
	// but is not part of the production orchestration path.
	if uao.config.EnableEchodream {
		uao.dreamCycle = echodream.NewSleepWakeStateMachine()
		uao.dreamCycle.ConfigurePhaseDurations(
			uao.config.DreamLightDuration,
			uao.config.DreamDeepDuration,
			uao.config.DreamREMDuration,
		)
		fmt.Println("   ✓ Canonical EchoDream Sleep/Wake State Machine initialized")
	}

	// The wake/rest manager is the sole transition authority. Its callbacks keep
	// orchestrator state, conscious processing, and EchoDream synchronized.
	if uao.config.AutoWakeRest {
		uao.wakeRestCycle = NewAutonomousWakeRestManager()
		uao.wakeRestCycle.ConfigureDurations(uao.config.WakeDuration, uao.config.RestDuration)
		uao.wakeRestCycle.SetCallbacks(
			uao.onWake,
			uao.onRest,
			uao.onDreamStart,
			uao.onDreamEnd,
		)
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
	if uao.interestPatterns != nil && len(state.InterestTopics) > 0 {
		uao.interestPatterns.RestoreInterests(state.InterestTopics)
	}
}

func (uao *UnifiedAutonomousOrchestrator) warmLocalModel(reason string) error {
	if uao.localRuntime == nil || !uao.config.WarmLocalModelOnWake {
		return nil
	}
	timeout := uao.config.LocalModelWarmupTimeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(uao.ctx, timeout)
	defer cancel()
	if err := uao.localRuntime.Warmup(ctx); err != nil {
		return fmt.Errorf("%s: %w", reason, err)
	}
	return nil
}

func (uao *UnifiedAutonomousOrchestrator) maintainLocalRuntime() {
	if uao.localRuntime == nil {
		return
	}
	if uao.localRuntime.MaybeUnloadForMemoryPressure("unloaded by unified orchestrator memory-pressure policy") {
		return
	}
	uao.localRuntime.UnloadIdle("unloaded by unified orchestrator idle policy")
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

	if err := uao.warmLocalModel("wake warmup"); err != nil {
		fmt.Printf("⚠️  Native model warmup unavailable; continuing through routed providers: %v\n", err)
	}

	// Start every constructed subsystem in dependency order. If any component
	// fails, unwind already-started components so the production process cannot
	// report a half-awake autonomous state.
	cleanups := make([]func(), 0, 8)
	startSubsystem := func(name string, start func() error, stop func() error) error {
		if err := start(); err != nil {
			for i := len(cleanups) - 1; i >= 0; i-- {
				cleanups[i]()
			}
			uao.mu.Lock()
			uao.running = false
			uao.isAwake = false
			uao.mu.Unlock()
			return fmt.Errorf("failed to start %s: %w", name, err)
		}
		cleanups = append(cleanups, func() { _ = stop() })
		return nil
	}

	if uao.interestPatterns != nil {
		if err := startSubsystem("interest patterns", uao.interestPatterns.Start, uao.interestPatterns.Stop); err != nil {
			return err
		}
	}
	if uao.skillLearning != nil {
		if err := startSubsystem("skill learning", uao.skillLearning.Start, uao.skillLearning.Stop); err != nil {
			return err
		}
	}
	if uao.discussionMonitor != nil {
		if err := startSubsystem("conversation monitor", uao.discussionMonitor.Start, uao.discussionMonitor.Stop); err != nil {
			return err
		}
	}
	if uao.discussionAutonomy != nil {
		if err := startSubsystem("discussion autonomy", uao.discussionAutonomy.Start, uao.discussionAutonomy.Stop); err != nil {
			return err
		}
		uao.syncDiscussionInterests()
	}
	if uao.globalTelemetry != nil {
		if err := startSubsystem("global telemetry", uao.globalTelemetry.Start, uao.globalTelemetry.Stop); err != nil {
			return err
		}
	}
	if uao.wisdomSynthesis != nil {
		if err := startSubsystem("wisdom synthesis", uao.wisdomSynthesis.Start, uao.wisdomSynthesis.Stop); err != nil {
			return err
		}
	}
	if uao.streamOfConsciousness != nil {
		if err := startSubsystem("stream of consciousness", uao.streamOfConsciousness.Start, uao.streamOfConsciousness.Stop); err != nil {
			return err
		}
	}
	if uao.echobeatsScheduler != nil {
		if err := startSubsystem("echobeats", uao.echobeatsScheduler.Start, uao.echobeatsScheduler.Stop); err != nil {
			return err
		}
	}
	if uao.wakeRestCycle != nil {
		if err := startSubsystem("wake/rest cycle", uao.wakeRestCycle.Start, uao.wakeRestCycle.Stop); err != nil {
			return err
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

// performCognitiveCycle executes one cycle of the autonomous cognitive loop.
// Wake/rest transitions are owned by AutonomousWakeRestManager callbacks; this
// loop never polls a second transition state machine or traps itself asleep.
func (uao *UnifiedAutonomousOrchestrator) performCognitiveCycle() {
	uao.mu.Lock()
	if !uao.isAwake {
		if uao.globalTelemetry != nil {
			uao.updateGlobalTelemetry()
		}
		uao.mu.Unlock()
		return
	}

	uao.totalCycles++

	if uao.streamOfConsciousness != nil {
		if metrics := uao.streamOfConsciousness.GetMetrics(); metrics != nil {
			if tt, ok := metrics["total_thoughts"].(uint64); ok && tt > uao.totalThoughts {
				uao.totalThoughts = tt
			}
		}
	}

	uao.adjustCognitiveLoad()
	cognitiveLoad := uao.cognitiveLoad
	shouldPractice := uao.skillLearning != nil && uao.shouldPracticeSkills()
	if uao.globalTelemetry != nil {
		uao.updateGlobalTelemetry()
	}
	uao.mu.Unlock()

	// New prompt-independent thoughts become bounded EchoDream experiences while
	// Echo is awake instead of being bulk-replayed on every rest cycle.
	uao.captureThoughtExperiences()
	uao.maintainLocalRuntime()

	// Cognitive load is the fatigue signal used by the sole wake/rest authority.
	if uao.wakeRestCycle != nil {
		uao.wakeRestCycle.UpdateCognitiveLoad(cognitiveLoad)
	}

	if uao.discussionMonitor != nil && uao.discussionAutonomy != nil {
		uao.checkDiscussions()
	}

	if shouldPractice {
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

	// Generate goals based on interests and record each chosen direction as a
	// durable dream experience. Existing non-completed descriptions are reused
	// instead of being enqueued again on every review interval.
	existing := make(map[string]struct{})
	for _, goal := range uao.echobeatsScheduler.GetGoalQueue() {
		if goal.Status != GoalCompleted {
			existing[goal.Description] = struct{}{}
		}
	}

	generated := 0
	for _, interest := range topInterests {
		description := fmt.Sprintf("Explore and deepen understanding of: %s", interest.Topic)
		if _, exists := existing[description]; exists {
			continue
		}
		goalID := uao.echobeatsScheduler.AddGoal(description, interest.Strength)
		existing[description] = struct{}{}
		generated++
		uao.totalGoals++
		uao.ingestDreamExperienceOnce(
			"goal-created:"+goalID,
			description,
			interest.Strength,
			[]string{"goal", "created", interest.Topic},
		)
	}

	fmt.Printf("   ✓ Generated %d new goals from interests\n", generated)
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
			uao.ingestDreamExperienceOnce(
				fmt.Sprintf("waking-wisdom:%d", time.Now().UnixNano()),
				wisdom.Insight,
				wisdom.Depth,
				[]string{"wisdom", "waking-reflection"},
			)
			fmt.Printf("   ✨ Wisdom synthesized: %s (depth: %.2f)\n",
				truncate(wisdom.Insight, 80), wisdom.Depth)
		}
	}
}

// syncPersistentState saves current cognitive state to persistent storage
func (uao *UnifiedAutonomousOrchestrator) syncPersistentState() {
	uao.mu.Lock()
	if !uao.config.EnablePersistence {
		uao.mu.Unlock()
		return
	}

	now := time.Now()
	uao.lastStateSync = now
	persistentState := uao.persistentState
	isAwake := uao.isAwake
	totalCycles := uao.totalCycles
	totalThoughts := uao.totalThoughts
	totalGoals := uao.totalGoals
	totalWisdom := uao.totalWisdom
	wisdomDepth := uao.wisdomDepth
	cognitiveLoad := uao.cognitiveLoad
	sessionID := uao.sessionID
	startTime := uao.startTime
	stateDirectory := uao.config.StateDirectory
	uao.mu.Unlock()

	if persistentState == nil {
		fmt.Printf("💾 State sync skipped at %s: persistent state manager unavailable\n", now.Format("15:04:05"))
		return
	}

	wakeRestState := "Resting"
	if isAwake {
		wakeRestState = "Awake"
	}
	interestTopics := map[string]float64{}
	if uao.interestPatterns != nil {
		interestTopics = uao.interestPatterns.GetAllInterests()
	}

	persistentState.UpdateCognitiveState(
		int(totalCycles%12)+1,
		totalCycles,
		wisdomDepth,
		cognitiveLoad,
		0,
	)
	persistentState.UpdateWakeRestState(wakeRestState, totalWisdom, time.Since(startTime), 0)

	persistentState.mu.Lock()
	if persistentState.state != nil {
		persistentState.state.SessionID = sessionID
		persistentState.state.TotalThoughts = totalThoughts
		persistentState.state.TotalGoals = totalGoals
		persistentState.state.TotalInsights = totalWisdom
		persistentState.state.CuriosityLevel = wisdomDepth
		persistentState.state.InterestTopics = interestTopics
	}
	persistentState.mu.Unlock()

	if err := persistentState.Save(); err != nil {
		fmt.Printf("⚠️  State sync failed at %s: %v\n", now.Format("15:04:05"), err)
		return
	}

	fmt.Printf("💾 State synced at %s (cycles: %d, thoughts: %d, wisdom: %.2f, state: %s)\n",
		now.Format("15:04:05"), totalCycles, totalThoughts, wisdomDepth, stateDirectory)
}

// onRest is invoked by the wake/rest manager after it becomes the sole
// transition authority. It pauses waking cognition and starts canonical EchoDream.
func (uao *UnifiedAutonomousOrchestrator) onRest() error {
	uao.mu.Lock()
	if !uao.isAwake {
		uao.mu.Unlock()
		return nil
	}
	uao.isAwake = false
	stream := uao.streamOfConsciousness
	scheduler := uao.echobeatsScheduler
	dream := uao.dreamCycle
	uao.mu.Unlock()

	fmt.Println("\n🌙 Transitioning to rest for knowledge consolidation...")
	uao.captureThoughtExperiences()

	if stream != nil {
		stream.Pause()
	}
	if scheduler != nil {
		scheduler.Pause()
	}
	if uao.localRuntime != nil && uao.config.CoolLocalModelOnRest {
		uao.localRuntime.Cooldown("unloaded for canonical EchoDream rest transition")
	}
	if dream != nil && !dream.IsAsleep() {
		if err := dream.EnterSleep(); err != nil {
			return fmt.Errorf("enter EchoDream sleep: %w", err)
		}
	}

	fmt.Println("   💤 Echo is now resting and consolidating knowledge")
	return nil
}

// onDreamStart records the manager's dream-state transition. EchoDream itself
// was started by onRest so light, deep, and REM phases remain one coherent cycle.
func (uao *UnifiedAutonomousOrchestrator) onDreamStart() error {
	fmt.Println("   🌙 EchoDream entered active dream integration")
	return nil
}

// onDreamEnd marks the manager's transition out of dream state. Wisdom is
// integrated in onWake after EchoDream has been told to wake.
func (uao *UnifiedAutonomousOrchestrator) onDreamEnd() error {
	fmt.Println("   🌅 EchoDream cycle completed; preparing waking integration")
	return nil
}

// onWake resumes waking cognition and integrates each newly synthesized dream
// insight exactly once into interests, conscious attention, and durable metrics.
func (uao *UnifiedAutonomousOrchestrator) onWake() error {
	fmt.Println("\n🌅 Waking up with consolidated knowledge...")

	if uao.dreamCycle != nil && uao.dreamCycle.IsAsleep() {
		if err := uao.dreamCycle.WakeUp(); err != nil {
			return fmt.Errorf("wake EchoDream: %w", err)
		}
	}

	integrated, reinforced := uao.integrateDreamWisdom()
	if err := uao.warmLocalModel("post-dream wake warmup"); err != nil {
		fmt.Printf("⚠️  Native model wake warmup unavailable; continuing through routed providers: %v\n", err)
	}

	uao.mu.Lock()
	uao.isAwake = true
	stream := uao.streamOfConsciousness
	scheduler := uao.echobeatsScheduler
	uao.mu.Unlock()

	if stream != nil {
		stream.Resume()
	}
	if scheduler != nil {
		scheduler.Resume()
	}

	fmt.Printf("   ✨ Integrated %d new wisdom insights from rest (%d interest topics reinforced)\n", integrated, reinforced)
	fmt.Println("   ☀️ Echo is now awake and aware")
	return nil
}

// markExperienceOnce records a source event in a bounded FIFO idempotency
// ledger. It is safe to call from asynchronous skill and scheduler callbacks.
func (uao *UnifiedAutonomousOrchestrator) markExperienceOnce(sourceID string) bool {
	sourceID = strings.TrimSpace(sourceID)
	if sourceID == "" {
		return false
	}

	uao.experienceMu.Lock()
	defer uao.experienceMu.Unlock()

	if _, exists := uao.ingestedExperienceIDs[sourceID]; exists {
		return false
	}

	uao.ingestedExperienceIDs[sourceID] = struct{}{}
	uao.experienceOrder = append(uao.experienceOrder, sourceID)
	if len(uao.experienceOrder) > maxExperienceLedgerEntries {
		overflow := len(uao.experienceOrder) - maxExperienceLedgerEntries
		for _, expired := range uao.experienceOrder[:overflow] {
			delete(uao.ingestedExperienceIDs, expired)
		}
		uao.experienceOrder = append([]string(nil), uao.experienceOrder[overflow:]...)
	}
	return true
}

// ingestDreamExperienceOnce forwards a normalized, non-duplicate experience to
// the canonical EchoDream processor.
func (uao *UnifiedAutonomousOrchestrator) ingestDreamExperienceOnce(sourceID, content string, importance float64, tags []string) bool {
	content = strings.TrimSpace(content)
	if uao.dreamCycle == nil || content == "" || !uao.markExperienceOnce("experience:"+sourceID) {
		return false
	}

	if importance < 0 {
		importance = 0
	} else if importance > 1 {
		importance = 1
	}
	uao.dreamCycle.IngestExperience(content, importance, append([]string(nil), tags...))
	return true
}

// captureThoughtExperiences hands each retained autonomous thought to
// EchoDream at most once, even when the stream buffer is revisited.
func (uao *UnifiedAutonomousOrchestrator) captureThoughtExperiences() int {
	if uao.streamOfConsciousness == nil || uao.dreamCycle == nil {
		return 0
	}

	ingested := 0
	for _, thought := range uao.streamOfConsciousness.GetAllThoughts() {
		sourceID := thought.ID
		if sourceID == "" {
			sourceID = fmt.Sprintf("thought:%d", thought.Timestamp.UnixNano())
		}
		tags := []string{"thought", strings.ToLower(thought.Type.String())}
		tags = append(tags, thought.Tags...)
		if uao.ingestDreamExperienceOnce(sourceID, thought.Content, thought.Importance, tags) {
			ingested++
		}
	}
	return ingested
}

// captureConversationExperiences records unseen incoming and Echo-authored
// messages so dream consolidation can learn from social outcomes and interests.
func (uao *UnifiedAutonomousOrchestrator) captureConversationExperiences(conversations []*TrackedConversation) int {
	ingested := 0
	for _, conv := range conversations {
		if conv == nil {
			continue
		}
		for index, message := range conv.Messages {
			messageID := message.ID
			if messageID == "" {
				messageID = fmt.Sprintf("%s:%d:%d", conv.ID, message.Timestamp.UnixNano(), index)
			}
			senderClass := "external"
			if message.IsFromEcho {
				senderClass = "echo"
			}
			tags := []string{"discussion", senderClass, conv.Topic}
			tags = append(tags, message.Topics...)
			importance := 0.4 + conv.InterestScore*0.5
			if uao.ingestDreamExperienceOnce("discussion:"+conv.ID+":"+messageID, message.Content, importance, tags) {
				ingested++
			}
		}
	}
	return ingested
}

// syncDiscussionInterests seeds social initiative from the same canonical
// interest graph used for goal formation without immediately starting a burst
// of discussions; the discussion loop evaluates opportunities on its cadence.
func (uao *UnifiedAutonomousOrchestrator) syncDiscussionInterests() {
	if uao.interestPatterns == nil || uao.discussionAutonomy == nil {
		return
	}
	for topic, strength := range uao.interestPatterns.GetAllInterests() {
		uao.discussionAutonomy.AddInterestPattern(topic, strength)
	}
}

// integrateDreamWisdom reinforces attention and continuity with each canonical
// EchoDream insight exactly once across repeated wake calls.
func (uao *UnifiedAutonomousOrchestrator) integrateDreamWisdom() (integrated, reinforced int) {
	if uao.dreamCycle == nil {
		return 0, 0
	}

	for _, insight := range uao.dreamCycle.GetWisdomInsights() {
		if !uao.markExperienceOnce("integrated-wisdom:" + insight.ID) {
			continue
		}

		if uao.interestPatterns != nil {
			topics := uao.interestPatterns.UpdateInterestFromInsight(insight.Insight, insight.Depth)
			reinforced += len(topics)
			if uao.discussionAutonomy != nil {
				for _, topic := range topics {
					uao.discussionAutonomy.AddInterestPattern(topic, uao.interestPatterns.GetInterestLevel(topic))
				}
			}
		}
		if uao.streamOfConsciousness != nil && insight.Depth > 0.5 {
			uao.streamOfConsciousness.AddInterest(insight.Insight, insight.Depth)
		}

		uao.mu.Lock()
		uao.wisdomDepth += insight.Depth
		uao.totalWisdom++
		uao.mu.Unlock()
		integrated++
	}
	return integrated, reinforced
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
	uao.captureConversationExperiences(conversations)
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
	skillID := skill.ID
	skillName := skill.Name
	before := skill.Proficiency
	practiceKey := fmt.Sprintf("skill:%s:%d", skillID, time.Now().UnixNano())
	fmt.Printf("🎓 Practicing skill: %s (proficiency %.2f)\n", skillName, before)
	go func() {
		if err := uao.skillLearning.PracticeSkill(skillID); err != nil {
			uao.ingestDreamExperienceOnce(
				practiceKey,
				fmt.Sprintf("Practice of %s failed: %v", skillName, err),
				0.6,
				[]string{"skill", "failure", skillName},
			)
			fmt.Printf("   ⚠️ Skill practice failed: %v\n", err)
			return
		}

		after := before
		if updated, err := uao.skillLearning.GetSkillByID(skillID); err == nil {
			after = updated.Proficiency
		}
		uao.ingestDreamExperienceOnce(
			practiceKey,
			fmt.Sprintf("Practiced %s; proficiency changed from %.3f to %.3f", skillName, before, after),
			0.8,
			[]string{"skill", "practice", skillName},
		)
	}()
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

	// Stop all subsystems outside uao.mu in reverse dependency order so each
	// component can finish callbacks without lock inversion or goroutine leaks.
	if uao.wakeRestCycle != nil {
		_ = uao.wakeRestCycle.Stop()
	}
	if uao.echobeatsScheduler != nil {
		_ = uao.echobeatsScheduler.Stop()
	}
	if uao.streamOfConsciousness != nil {
		_ = uao.streamOfConsciousness.Stop()
	}
	if uao.wisdomSynthesis != nil {
		_ = uao.wisdomSynthesis.Stop()
	}
	if uao.globalTelemetry != nil {
		_ = uao.globalTelemetry.Stop()
	}
	if uao.discussionAutonomy != nil {
		_ = uao.discussionAutonomy.Stop()
	}
	if uao.discussionMonitor != nil {
		_ = uao.discussionMonitor.Stop()
	}
	if uao.skillLearning != nil {
		_ = uao.skillLearning.Stop()
	}
	if uao.interestPatterns != nil {
		_ = uao.interestPatterns.Stop()
	}
	if uao.localRuntime != nil {
		_ = uao.localRuntime.Close()
	}
	if uao.dreamCycle != nil {
		uao.dreamCycle.Shutdown()
	}

	// Sync final state after cancellation and child shutdown; syncPersistentState
	// takes its own write lock to update lastStateSync safely.
	uao.syncPersistentState()

	fmt.Println("😴 Echo has gone to sleep. Goodnight.")

	return nil
}

// GetStatus returns a truthful snapshot of production orchestration state.
func (uao *UnifiedAutonomousOrchestrator) GetStatus() OrchestratorStatus {
	uao.mu.RLock()
	status := OrchestratorStatus{
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
	wakeRest := uao.wakeRestCycle
	dream := uao.dreamCycle
	provider := uao.llmProvider
	uao.mu.RUnlock()

	if provider != nil {
		status.Provider = provider.Name()
		status.ProviderAvailable = provider.Available()
		if backendProvider, ok := provider.(interface {
			GetBackendState() llm.BackendRoutingState
		}); ok {
			status.Backend = backendProvider.GetBackendState()
		}
	}
	if wakeRest != nil {
		status.WakeRestState = wakeRest.GetCurrentState().String()
	}
	if dream != nil {
		metrics := dream.GetMetrics()
		status.DreamPhase, _ = metrics["current_phase"].(string)
		status.PendingExperiences, _ = metrics["pending_experiences"].(int)
		status.DreamWisdom, _ = metrics["wisdom_synthesized"].(uint64)
	}

	uao.experienceMu.Lock()
	status.ExperienceLedgerSize = len(uao.experienceOrder)
	uao.experienceMu.Unlock()
	return status
}

// OrchestratorStatus represents the current status of the orchestrator.
type OrchestratorStatus struct {
	Running              bool
	IsAwake              bool
	IsAutonomous         bool
	CognitiveLoad        float64
	WisdomDepth          float64
	SessionID            string
	Uptime               time.Duration
	TotalCycles          uint64
	TotalThoughts        uint64
	TotalGoals           uint64
	TotalWisdom          uint64
	LastStateSync        time.Time
	StateDirectory       string
	Provider             string
	ProviderAvailable    bool
	Backend              llm.BackendRoutingState
	WakeRestState        string
	DreamPhase           string
	PendingExperiences   int
	DreamWisdom          uint64
	ExperienceLedgerSize int
}

// GlobalState is declared in global_telemetry_shell.go and shared with the
// telemetry shell's UpdateOrchestratorState method.
