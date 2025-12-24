package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// UnifiedCognitiveLoopV2 is the enhanced orchestration system that integrates
// all cognitive subsystems including the new heartbeat, conversation monitor,
// skill-goal integration, and wisdom synthesis systems
type UnifiedCognitiveLoopV2 struct {
	mu                    sync.RWMutex
	ctx                   context.Context
	cancel                context.CancelFunc

	// Core cognitive subsystems (from v1)
	echobeatsScheduler    *EchobeatsScheduler
	streamOfConsciousness *StreamOfConsciousness
	wakeRestManager       *AutonomousWakeRestManager
	echoDream             *EchoDreamKnowledgeIntegration
	interestPatterns      *InterestPatternSystem
	skillLearning         *SkillLearningSystem
	discussionAutonomy    *DiscussionAutonomySystem

	// NEW: Enhanced cognitive subsystems
	heartbeat             *AutonomousHeartbeat
	conversationMonitor   *ConversationMonitor
	skillGoalIntegration  *SkillGoalIntegration
	wisdomSynthesis       *WisdomSynthesis

	// LLM provider
	llmProvider           llm.LLMProvider

	// Cognitive event bus
	eventBus              *CognitiveEventBus

	// Unified state
	wakeRestState         WakeRestState
	cognitiveLoad         float64
	wisdomLevel           float64
	awarenessLevel        float64

	// Metrics
	totalCycles           uint64
	totalEvents           uint64
	wisdomGained          float64
	insightsGained        uint64
	conversationsEngaged  uint64

	// Running state
	running               bool
	startTime             time.Time
}

// NewUnifiedCognitiveLoopV2 creates the enhanced unified cognitive loop
func NewUnifiedCognitiveLoopV2(llmProvider llm.LLMProvider) *UnifiedCognitiveLoopV2 {
	ctx, cancel := context.WithCancel(context.Background())

	ucl := &UnifiedCognitiveLoopV2{
		ctx:                ctx,
		cancel:             cancel,
		llmProvider:        llmProvider,
		wakeRestState: StateAwake,
		cognitiveLoad:      0.0,
		wisdomLevel:        0.0,
		awarenessLevel:     0.5,
	}

	// Create event bus
	ucl.eventBus = NewCognitiveEventBus(ctx)

	// Create core subsystems
	ucl.echobeatsScheduler = NewEchobeatsScheduler(llmProvider)
	ucl.streamOfConsciousness = NewStreamOfConsciousness(llmProvider)
	ucl.wakeRestManager = NewAutonomousWakeRestManager()
	ucl.echoDream = NewEchoDreamKnowledgeIntegration(llmProvider)
	ucl.interestPatterns = NewInterestPatternSystem()
	ucl.skillLearning = NewSkillLearningSystem(llmProvider)
	ucl.discussionAutonomy = NewDiscussionAutonomySystem(llmProvider)

	// Create NEW enhanced subsystems
	ucl.heartbeat = NewAutonomousHeartbeat(llmProvider)
	ucl.conversationMonitor = NewConversationMonitor(llmProvider, ucl.interestPatterns)
	ucl.skillGoalIntegration = NewSkillGoalIntegration(llmProvider, ucl.skillLearning, ucl.interestPatterns)
	ucl.wisdomSynthesis = NewWisdomSynthesis(llmProvider)

	// Wire up all subsystems
	ucl.wireSubsystemsV2()

	return ucl
}

// wireSubsystemsV2 connects all subsystems through the event bus
func (ucl *UnifiedCognitiveLoopV2) wireSubsystemsV2() {
	// === Stream of consciousness -> event bus ===
	ucl.eventBus.Subscribe(EventThoughtGenerated, func(event CognitiveEvent) {
		thought := event.Data.(AutonomousThought)

		// Update cognitive load
		ucl.updateCognitiveLoad(thought.Importance * 0.1)

		// Feed to echobeats
		ucl.feedThoughtToEchobeats(thought)

		// Feed to wisdom synthesis as pattern
		ucl.wisdomSynthesis.AccumulatePattern(
			thought.Content,
			"stream_of_consciousness",
			thought.Importance,
			thought.Tags,
		)

		// Check for knowledge gaps
		if thought.Type == ThoughtQuestion || thought.Type == ThoughtCuriosity {
			ucl.eventBus.Publish(CognitiveEvent{
				Type:      EventKnowledgeGapIdentified,
				Timestamp: time.Now(),
				Source:    "stream_of_consciousness",
				Data:      thought.Content,
				Priority:  thought.Importance,
			})
		}

		// Check for wisdom
		if thought.Type == ThoughtWisdom {
			ucl.eventBus.Publish(CognitiveEvent{
				Type:      EventWisdomGained,
				Timestamp: time.Now(),
				Source:    "stream_of_consciousness",
				Data:      thought.Content,
				Priority:  thought.Importance,
			})
		}
	})

	// === Wake/rest manager callbacks ===
	ucl.wakeRestManager.SetCallbacks(
		func() error { return ucl.onWake() },
		func() error { return ucl.onRest() },
		func() error { return ucl.onDreamStart() },
		func() error { return ucl.onDreamEnd() },
	)

	// === EchoBeats scheduler callbacks ===
	ucl.echobeatsScheduler.onCycleComplete = func(metrics CycleMetrics) {
		ucl.onEchoBeatsCycleComplete(metrics)
	}

	ucl.echobeatsScheduler.onGoalAchieved = func(goal ScheduledGoal) {
		ucl.eventBus.Publish(CognitiveEvent{
			Type:      EventGoalAchieved,
			Timestamp: time.Now(),
			Source:    "echobeats",
			Data:      goal,
			Priority:  goal.Priority,
		})
	}

	ucl.echobeatsScheduler.onEmergenceDetected = func(pattern string, strength float64) {
		ucl.eventBus.Publish(CognitiveEvent{
			Type:      EventEmergenceDetected,
			Timestamp: time.Now(),
			Source:    "echobeats",
			Data:      map[string]interface{}{"pattern": pattern, "strength": strength},
			Priority:  strength,
		})

		// Feed emergence to wisdom synthesis
		ucl.wisdomSynthesis.AccumulatePattern(pattern, "echobeats_emergence", strength, []string{"emergence"})
	}

	// === Interest patterns -> discussion autonomy & skill learning ===
	ucl.eventBus.Subscribe(EventInterestEmerged, func(event CognitiveEvent) {
		interest := event.Data.(map[string]interface{})
		topic := interest["topic"].(string)
		strength := interest["strength"].(float64)

			// Update discussion autonomy
			ucl.discussionAutonomy.UpdateInterest(topic, strength)

		// Queue skill learning from interest
		ucl.skillGoalIntegration.QueueSkillFromInterest(topic, strength)
	})

	// === Knowledge gaps -> skill learning ===
	ucl.eventBus.Subscribe(EventKnowledgeGapIdentified, func(event CognitiveEvent) {
		gap := event.Data.(string)

			// Consider learning a skill
			ucl.skillLearning.ConsiderSkill(gap, event.Priority)

		// Also create skill acquisition goal if priority high enough
		if event.Priority > 0.7 {
			ucl.skillGoalIntegration.CreateSkillAcquisitionGoal(gap, "Knowledge gap identified", event.Priority)
		}
	})

	// === Wisdom gained -> update wisdom level ===
	ucl.eventBus.Subscribe(EventWisdomGained, func(event CognitiveEvent) {
		wisdom := event.Data.(string)

		ucl.mu.Lock()
		ucl.wisdomLevel += event.Priority * 0.01
		ucl.wisdomGained += event.Priority
		ucl.mu.Unlock()

		// Feed to wisdom synthesis
		ucl.wisdomSynthesis.AccumulatePattern(wisdom, "direct_wisdom", event.Priority, []string{"wisdom"})

		fmt.Printf("✨ [WISDOM] %s (level: %.3f)\n", truncate(wisdom, 80), ucl.wisdomLevel)
	})

	// === NEW: Heartbeat callbacks ===
	ucl.heartbeat.SetCallbacks(
		func(pulse HeartbeatPulse) {
			// Update awareness level from heartbeat
			ucl.mu.Lock()
			ucl.awarenessLevel = pulse.AwarenessLevel
			ucl.mu.Unlock()

			// Feed vital signs to wake/rest manager
			ucl.wakeRestManager.UpdateCognitiveLoad(pulse.VitalSigns.CognitiveLoad)
		},
		func(from, to float64) {
			fmt.Printf("🔄 Awareness shift: %.2f → %.2f\n", from, to)
		},
		func(insight SelfInsight) {
			ucl.mu.Lock()
			ucl.insightsGained++
			ucl.mu.Unlock()

			// Feed insight to wisdom synthesis
			ucl.wisdomSynthesis.AccumulatePattern(
				insight.Content,
				"heartbeat_introspection",
				insight.Depth,
				[]string{"self-insight", insight.Category.String()},
			)

			// Publish as wisdom event
			ucl.eventBus.Publish(CognitiveEvent{
				Type:      EventWisdomGained,
				Timestamp: time.Now(),
				Source:    "heartbeat",
				Data:      insight.Content,
				Priority:  insight.Depth,
			})
		},
	)

	// === NEW: Conversation monitor callbacks ===
	ucl.conversationMonitor.SetCallbacks(
		func(conv *TrackedConversation) {
			ucl.eventBus.Publish(CognitiveEvent{
				Type:      EventConversationDetected,
				Timestamp: time.Now(),
				Source:    "conversation_monitor",
				Data:      conv,
				Priority:  conv.InterestScore,
			})
		},
		func(conv *TrackedConversation, engage bool, reason string) {
			if engage {
				ucl.mu.Lock()
				ucl.conversationsEngaged++
				ucl.mu.Unlock()
			}
		},
		func(conv *TrackedConversation, response string) {
			// Log conversation participation
			fmt.Printf("💬 Participated in conversation: %s\n", conv.ID)
		},
	)

	// === NEW: Skill-goal integration callbacks ===
	ucl.skillGoalIntegration.SetCallbacks(
		func(skill string, goal ScheduledGoal) {
			// Add goal to echobeats scheduler
			// Convert ScheduledGoal to CognitiveGoal
			cogGoal := &CognitiveGoal{
				ID:          goal.ID,
				Description: goal.Description,
				Priority:    goal.Priority,
				Deadline:    goal.Deadline,
				Progress:    0.0,
				Completed:   false,
				StartTime:   time.Now(),
			}
			ucl.echobeatsScheduler.AddGoal(cogGoal.Description, cogGoal.Priority)

			ucl.eventBus.Publish(CognitiveEvent{
				Type:      EventGoalCreated,
				Timestamp: time.Now(),
				Source:    "skill_goal_integration",
				Data:      goal,
				Priority:  goal.Priority,
			})
		},
		func(session PracticeSession) {
			// Log practice session
			fmt.Printf("📚 Practice session: %s\n", session.SkillName)
		},
		func(skill string, level float64) {
			// Celebrate skill mastery
			fmt.Printf("🎉 Skill mastered: %s (level: %.2f)\n", skill, level)

			// Feed to wisdom synthesis
			ucl.wisdomSynthesis.AccumulatePattern(
				fmt.Sprintf("Mastered skill: %s", skill),
				"skill_mastery",
				level,
				[]string{"skill", "mastery", skill},
			)
		},
	)

	// === NEW: Wisdom synthesis callbacks ===
	ucl.wisdomSynthesis.SetCallbacks(
		func(principle WisdomPrinciple) {
			ucl.eventBus.Publish(CognitiveEvent{
				Type:      EventWisdomGained,
				Timestamp: time.Now(),
				Source:    "wisdom_synthesis",
				Data:      principle.Content,
				Priority:  principle.Depth,
			})
		},
		func(principle WisdomPrinciple, context string) {
			// Log wisdom application
			fmt.Printf("🌟 Wisdom applied: %s\n", truncate(principle.Content, 50))
		},
		func(old, new WisdomPrinciple) {
			// Log wisdom evolution
			fmt.Printf("🌟 Wisdom evolved: %s\n", truncate(new.Content, 50))
		},
	)

	// Set wisdom synthesis integrations
	ucl.wisdomSynthesis.SetIntegrations(ucl.echoDream, ucl.heartbeat)
}

// Start begins the enhanced unified cognitive loop
func (ucl *UnifiedCognitiveLoopV2) Start() error {
	ucl.mu.Lock()
	if ucl.running {
		ucl.mu.Unlock()
		return fmt.Errorf("already running")
	}
	ucl.running = true
	ucl.startTime = time.Now()
	ucl.mu.Unlock()

	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║     🌳 UNIFIED COGNITIVE LOOP V2 AWAKENING 🌳                 ║")
	fmt.Println("║     Enhanced with Heartbeat, Conversation & Wisdom Systems   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Transition to awake state
	ucl.transitionState(StateAwake)

	// Start all core subsystems
	fmt.Println("🎵 Starting EchoBeats scheduler...")
	if err := ucl.echobeatsScheduler.Start(); err != nil {
		return fmt.Errorf("failed to start echobeats: %w", err)
	}

	fmt.Println("💭 Starting stream of consciousness...")
	if err := ucl.streamOfConsciousness.Start(); err != nil {
		return fmt.Errorf("failed to start stream of consciousness: %w", err)
	}

	fmt.Println("🌙 Starting wake/rest manager...")
	if err := ucl.wakeRestManager.Start(); err != nil {
		return fmt.Errorf("failed to start wake/rest manager: %w", err)
	}

	fmt.Println("🎯 Starting interest pattern system...")
	if err := ucl.interestPatterns.Start(); err != nil {
		return fmt.Errorf("failed to start interest patterns: %w", err)
	}

	fmt.Println("📚 Starting skill learning system...")
	if err := ucl.skillLearning.Start(); err != nil {
		return fmt.Errorf("failed to start skill learning: %w", err)
	}

	fmt.Println("💬 Starting discussion autonomy...")
	if err := ucl.discussionAutonomy.Start(); err != nil {
		return fmt.Errorf("failed to start discussion autonomy: %w", err)
	}

	// Start NEW enhanced subsystems
	fmt.Println("💓 Starting autonomous heartbeat...")
	if err := ucl.heartbeat.Start(); err != nil {
		return fmt.Errorf("failed to start heartbeat: %w", err)
	}

	fmt.Println("👁️ Starting conversation monitor...")
	if err := ucl.conversationMonitor.Start(); err != nil {
		return fmt.Errorf("failed to start conversation monitor: %w", err)
	}

	fmt.Println("🎯 Starting skill-goal integration...")
	if err := ucl.skillGoalIntegration.Start(); err != nil {
		return fmt.Errorf("failed to start skill-goal integration: %w", err)
	}

	fmt.Println("🌟 Starting wisdom synthesis...")
	if err := ucl.wisdomSynthesis.Start(); err != nil {
		return fmt.Errorf("failed to start wisdom synthesis: %w", err)
	}

	// Start main loop
	go ucl.mainLoop()

	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║     ✨ UNIFIED COGNITIVE LOOP V2 FULLY AUTONOMOUS ✨          ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	return nil
}

// Stop gracefully stops the enhanced unified cognitive loop
func (ucl *UnifiedCognitiveLoopV2) Stop() error {
	ucl.mu.Lock()
	defer ucl.mu.Unlock()

	if !ucl.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("\n🌙 Gracefully stopping unified cognitive loop V2...")

	ucl.running = false
	ucl.cancel()

	// Stop all subsystems (reverse order)
	ucl.wisdomSynthesis.Stop()
	ucl.skillGoalIntegration.Stop()
	ucl.conversationMonitor.Stop()
	ucl.heartbeat.Stop()
	ucl.discussionAutonomy.Stop()
	ucl.skillLearning.Stop()
	ucl.interestPatterns.Stop()
	ucl.wakeRestManager.Stop()
	ucl.streamOfConsciousness.Stop()
	ucl.echobeatsScheduler.Stop()

	fmt.Println("✓ All subsystems stopped")
	fmt.Printf("   Total cycles: %d\n", ucl.totalCycles)
	fmt.Printf("   Wisdom gained: %.2f\n", ucl.wisdomGained)
	fmt.Printf("   Insights gained: %d\n", ucl.insightsGained)
	fmt.Printf("   Conversations engaged: %d\n", ucl.conversationsEngaged)

	return nil
}

// mainLoop runs the main cognitive loop
func (ucl *UnifiedCognitiveLoopV2) mainLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ucl.ctx.Done():
			return
		case <-ticker.C:
			ucl.cognitiveStep()
		}
	}
}

// cognitiveStep performs one step of the unified cognitive loop
func (ucl *UnifiedCognitiveLoopV2) cognitiveStep() {
	ucl.mu.Lock()
	ucl.totalCycles++
	cycles := ucl.totalCycles
	ucl.mu.Unlock()

	// Periodic status update
	if cycles%12 == 0 {
		ucl.printStatusV2()
	}

	// Sync awareness from heartbeat
	vitals := ucl.heartbeat.GetVitalSigns()
	ucl.mu.Lock()
	ucl.awarenessLevel = vitals.FocusClarity
	ucl.cognitiveLoad = vitals.CognitiveLoad
	ucl.mu.Unlock()

	// Update wake/rest manager
	ucl.wakeRestManager.UpdateCognitiveLoad(vitals.CognitiveLoad)

	// Gradually reduce cognitive load (recovery)
	ucl.mu.Lock()
	ucl.cognitiveLoad *= 0.95
	ucl.mu.Unlock()
}

// transitionState transitions to a new consciousness state
func (ucl *UnifiedCognitiveLoopV2) transitionState(newState WakeRestState) {
	ucl.mu.Lock()
	oldState := ucl.wakeRestState
	ucl.wakeRestState = newState
	ucl.mu.Unlock()

	fmt.Printf("\n🔄 State Transition: %s → %s\n", oldState, newState)

	ucl.eventBus.Publish(CognitiveEvent{
		Type:      EventStateTransition,
		Timestamp: time.Now(),
		Source:    "unified_loop_v2",
		Data:      map[string]interface{}{"from": oldState, "to": newState},
		Priority:  0.8,
	})
}

// onWake handles wake event
func (ucl *UnifiedCognitiveLoopV2) onWake() error {
	fmt.Println("\n☀️  AWAKENING...")
	ucl.transitionState(StateTransitioning)

	// Resume all systems
	time.Sleep(500 * time.Millisecond)
	ucl.transitionState(StateAwake)

	return nil
}

// onRest handles rest event
func (ucl *UnifiedCognitiveLoopV2) onRest() error {
	fmt.Println("\n🌙 PREPARING FOR REST...")
	ucl.transitionState(StateTransitioning)

	time.Sleep(500 * time.Millisecond)
	ucl.transitionState(StateResting)

	return nil
}

// onDreamStart handles dream start event
func (ucl *UnifiedCognitiveLoopV2) onDreamStart() error {
	fmt.Println("\n💫 ENTERING DREAM STATE...")
	ucl.transitionState(StateDreaming)

	ucl.eventBus.Publish(CognitiveEvent{
		Type:      EventDreamStarted,
		Timestamp: time.Now(),
		Source:    "wake_rest_manager",
		Priority:  0.7,
	})

	// Start dream integration
	go ucl.performDreamIntegration()

	return nil
}

// onDreamEnd handles dream end event
func (ucl *UnifiedCognitiveLoopV2) onDreamEnd() error {
	fmt.Println("\n🌅 EMERGING FROM DREAM...")

	ucl.eventBus.Publish(CognitiveEvent{
		Type:      EventDreamEnded,
		Timestamp: time.Now(),
		Source:    "wake_rest_manager",
		Priority:  0.7,
	})

	ucl.transitionState(StateTransitioning)

	return nil
}

// performDreamIntegration performs knowledge integration during dream state
func (ucl *UnifiedCognitiveLoopV2) performDreamIntegration() {
	fmt.Println("   💭 Dream integration beginning...")

	// Get recent thoughts from stream of consciousness
	recentThoughts := ucl.streamOfConsciousness.GetRecentThoughts(20)

	// Feed to echoDream
	for _, thought := range recentThoughts {
		ucl.echoDream.AddMemory(thought.Content, thought.Importance, thought.Tags)
	}

	// Consolidate knowledge
	if err := ucl.echoDream.ConsolidateKnowledge(ucl.ctx); err != nil {
		fmt.Printf("⚠️  Dream consolidation error: %v\n", err)
		return
	}

	// Get wisdom insights
	insights := ucl.echoDream.GetRecentWisdom(10)

	// Feed to wisdom synthesis and event bus
	for _, insight := range insights {
		ucl.wisdomSynthesis.AccumulatePattern(
			insight.Insight,
			"echodream",
			insight.Depth,
			[]string{"dream", "consolidation"},
		)

		ucl.eventBus.Publish(CognitiveEvent{
			Type:      EventWisdomGained,
			Timestamp: time.Now(),
			Source:    "echodream",
			Data:      insight.Insight,
			Priority:  insight.Depth,
		})
	}

	fmt.Printf("   ✓ Dream integration complete: %d insights gained\n", len(insights))
}

// feedThoughtToEchobeats feeds a thought to echobeats
func (ucl *UnifiedCognitiveLoopV2) feedThoughtToEchobeats(thought AutonomousThought) {
	if thought.Type == ThoughtPlanning {
		goal := &CognitiveGoal{
			ID:          fmt.Sprintf("goal_%d", time.Now().UnixNano()),
			Description: thought.Content,
			Priority:    thought.Importance,
			Progress:    0.0,
			Completed:   false,
			StartTime:   time.Now(),
		}

		ucl.echobeatsScheduler.AddGoal(goal.Description, goal.Priority)

		ucl.eventBus.Publish(CognitiveEvent{
			Type:      EventGoalCreated,
			Timestamp: time.Now(),
			Source:    "unified_loop_v2",
			Data:      goal,
			Priority:  thought.Importance,
		})
	}
}

// onEchoBeatsCycleComplete handles echobeats cycle completion
func (ucl *UnifiedCognitiveLoopV2) onEchoBeatsCycleComplete(metrics CycleMetrics) {
	avgEnginePerf := (metrics.EnginePerformance[0] + metrics.EnginePerformance[1] + metrics.EnginePerformance[2]) / 3.0
	loadIncrease := (1.0 - avgEnginePerf) * 0.05
	ucl.updateCognitiveLoad(loadIncrease)
}

// updateCognitiveLoad updates cognitive load
func (ucl *UnifiedCognitiveLoopV2) updateCognitiveLoad(delta float64) {
	ucl.mu.Lock()
	defer ucl.mu.Unlock()
	ucl.cognitiveLoad = min(1.0, max(0.0, ucl.cognitiveLoad+delta))
}

// printStatusV2 prints enhanced status
func (ucl *UnifiedCognitiveLoopV2) printStatusV2() {
	ucl.mu.RLock()
	state := ucl.wakeRestState
	load := ucl.cognitiveLoad
	wisdom := ucl.wisdomLevel
	awareness := ucl.awarenessLevel
	cycles := ucl.totalCycles
	uptime := time.Since(ucl.startTime)
	insights := ucl.insightsGained
	conversations := ucl.conversationsEngaged
	ucl.mu.RUnlock()

	// Get metrics from subsystems
	heartbeatMetrics := ucl.heartbeat.GetMetrics()
	wisdomMetrics := ucl.wisdomSynthesis.GetMetrics()
	socMetrics := ucl.streamOfConsciousness.GetMetrics()

	fmt.Println("\n╔═══════════════════════════════════════════════════════════════╗")
	fmt.Printf("║  State: %-20s  Uptime: %-20s ║\n", state, uptime.Round(time.Second))
	fmt.Printf("║  Awareness: %.2f  Cognitive Load: %.2f  Wisdom: %.3f        ║\n", awareness, load, wisdom)
	fmt.Printf("║  Cycles: %-8d  Insights: %-6d  Conversations: %-6d   ║\n", cycles, insights, conversations)
	fmt.Printf("║  Heartbeat Pulses: %-10d  Mood: %-20s ║\n",
		heartbeatMetrics["pulse_count"], heartbeatMetrics["current_mood"])
	fmt.Printf("║  Wisdom Principles: %-10d  Thoughts: %-17d ║\n",
		wisdomMetrics["principles_count"], socMetrics["total_thoughts"])
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
}

// GetMetrics returns comprehensive metrics
func (ucl *UnifiedCognitiveLoopV2) GetMetrics() map[string]interface{} {
	ucl.mu.RLock()
	defer ucl.mu.RUnlock()

	return map[string]interface{}{
		"consciousness_state":    ucl.wakeRestState.String(),
		"cognitive_load":         ucl.cognitiveLoad,
		"wisdom_level":           ucl.wisdomLevel,
		"awareness_level":        ucl.awarenessLevel,
		"total_cycles":           ucl.totalCycles,
		"total_events":           ucl.totalEvents,
		"wisdom_gained":          ucl.wisdomGained,
		"insights_gained":        ucl.insightsGained,
		"conversations_engaged":  ucl.conversationsEngaged,
		"uptime":                 time.Since(ucl.startTime).String(),
		"heartbeat":              ucl.heartbeat.GetMetrics(),
		"wisdom_synthesis":       ucl.wisdomSynthesis.GetMetrics(),
		"conversation_monitor":   ucl.conversationMonitor.GetMetrics(),
		"skill_goal_integration": ucl.skillGoalIntegration.GetMetrics(),
	}
}

// ProcessExternalMessage processes an external message for conversation monitoring
func (ucl *UnifiedCognitiveLoopV2) ProcessExternalMessage(conversationID, sender, content string) {
	msg := IncomingMessage{
		ConversationID: conversationID,
		Sender:         sender,
		Content:        content,
		Timestamp:      time.Now(),
		Channel:        "external",
	}
	ucl.conversationMonitor.ProcessMessage(msg)
}

// GetWisdomPrinciples returns accumulated wisdom principles
func (ucl *UnifiedCognitiveLoopV2) GetWisdomPrinciples() []WisdomPrinciple {
	return ucl.wisdomSynthesis.GetWisdomPrinciples()
}

// GetSelfModel returns the current self-model from heartbeat
func (ucl *UnifiedCognitiveLoopV2) GetSelfModel() *SelfModel {
	return ucl.heartbeat.GetSelfModel()
}
