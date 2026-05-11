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
	sessionID     string
	startTime     time.Time
	lastStateSync time.Time

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
	SessionName     string

	// Persistence
	EnablePersistence bool
	StateSyncInterval time.Duration
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
		SessionName:                 fmt.Sprintf("echo_session_%d", time.Now().Unix()),
		EnablePersistence:           true,
		StateSyncInterval:           5 * time.Minute,
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

	return orchestrator
}

// initializeSubsystems creates and connects all cognitive subsystems
func (uao *UnifiedAutonomousOrchestrator) initializeSubsystems() {
	fmt.Println("🧠 Initializing Echo cognitive subsystems...")

	// Stream of consciousness for continuous thought generation
	if uao.config.EnableStreamOfConsciousness {
		uao.streamOfConsciousness = NewStreamOfConsciousness(uao.llmProvider)
		uao.streamOfConsciousness.thoughtInterval = uao.config.ThoughtInterval
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

	// TODO: Implement actual persistence to Supabase or local storage
	now := time.Now()
	uao.lastStateSync = now

	// For now, just log the sync
	fmt.Printf("💾 State synced at %s (cycles: %d, thoughts: %d, wisdom: %.2f)\n",
		now.Format("15:04:05"), uao.totalCycles, uao.totalThoughts, uao.wisdomDepth)
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

		// Update interests based on consolidated patterns
		for range insights {
			// Extract topics from wisdom insights and update interest patterns
			if uao.interestPatterns != nil {
				// TODO: Implement UpdateInterest method
				// uao.interestPatterns.UpdateInterest(insight.Insight, insight.Depth)
			}
		}

		fmt.Printf("   ✨ Integrated %d wisdom insights from rest\n", len(insights))
	}

	fmt.Println("   ☀️ Echo is now awake and aware")
}

// updateGlobalTelemetry updates the global telemetry shell with current state
func (uao *UnifiedAutonomousOrchestrator) updateGlobalTelemetry() {
	if uao.globalTelemetry == nil {
		return
	}

	// TODO: Implement UpdateState method on GlobalTelemetryShell
	// telemetryState := GlobalState{
	// 	Timestamp:     time.Now(),
	// 	IsAwake:       uao.isAwake,
	// 	CognitiveLoad: uao.cognitiveLoad,
	// 	WisdomDepth:   uao.wisdomDepth,
	// 	TotalCycles:   uao.totalCycles,
	// 	TotalThoughts: uao.totalThoughts,
	// 	SessionID:     uao.sessionID,
	// }
	// uao.globalTelemetry.UpdateState(telemetryState)
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

	// TODO: Implement HasActiveDiscussions and IsPracticing methods
	// if uao.discussionMonitor != nil && uao.discussionMonitor.HasActiveDiscussions() {
	// 	load += 0.2
	// }
	//
	// if uao.skillLearning != nil && uao.skillLearning.IsPracticing() {
	// 	load += 0.2
	// }

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

// checkDiscussions monitors discussions and decides on participation
func (uao *UnifiedAutonomousOrchestrator) checkDiscussions() {
	// TODO: Implement discussion monitoring methods
	// discussions := uao.discussionMonitor.GetActiveDiscussions()
	// for _, discussion := range discussions {
	// 	relevance := uao.interestPatterns.CalculateRelevance(discussion.Topic)
	// 	if relevance > 0.6 && uao.discussionAutonomy != nil {
	// 		shouldParticipate := uao.discussionAutonomy.ShouldParticipate(discussion, relevance)
	// 		if shouldParticipate {
	// 			uao.discussionAutonomy.Participate(uao.ctx, discussion)
	// 		}
	// 	}
	// }
}

// shouldPracticeSkills determines if it's a good time to practice skills
func (uao *UnifiedAutonomousOrchestrator) shouldPracticeSkills() bool {
	// Practice skills when cognitive load is moderate and awake
	return uao.isAwake && uao.cognitiveLoad < 0.7 && uao.cognitiveLoad > 0.3
}

// practiceSkills engages in skill practice
func (uao *UnifiedAutonomousOrchestrator) practiceSkills() {
	// TODO: Implement skill practice methods
	// if uao.skillLearning == nil {
	// 	return
	// }
	// skills := uao.skillLearning.GetSkillsNeedingPractice()
	// if len(skills) > 0 {
	// 	skill := skills[0]
	// 	fmt.Printf("🎓 Practicing skill: %s\n", skill.Name)
	// 	uao.skillLearning.PracticeSkill(uao.ctx, skill)
	// }
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
		Running:       uao.running,
		IsAwake:       uao.isAwake,
		IsAutonomous:  uao.isAutonomous,
		CognitiveLoad: uao.cognitiveLoad,
		WisdomDepth:   uao.wisdomDepth,
		SessionID:     uao.sessionID,
		Uptime:        time.Since(uao.startTime),
		TotalCycles:   uao.totalCycles,
		TotalThoughts: uao.totalThoughts,
		TotalGoals:    uao.totalGoals,
		TotalWisdom:   uao.totalWisdom,
	}
}

// OrchestratorStatus represents the current status of the orchestrator
type OrchestratorStatus struct {
	Running       bool
	IsAwake       bool
	IsAutonomous  bool
	CognitiveLoad float64
	WisdomDepth   float64
	SessionID     string
	Uptime        time.Duration
	TotalCycles   uint64
	TotalThoughts uint64
	TotalGoals    uint64
	TotalWisdom   uint64
}

// GlobalState represents the global telemetry state
type GlobalState struct {
	Timestamp     time.Time
	IsAwake       bool
	CognitiveLoad float64
	WisdomDepth   float64
	TotalCycles   uint64
	TotalThoughts uint64
	SessionID     string
}
