// +build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cogpy/echo9llama/core/autonomous"
	"github.com/cogpy/echo9llama/core/consciousness"
	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/goals"
	"github.com/cogpy/echo9llama/core/llm"
	"github.com/cogpy/echo9llama/core/wisdom"
)

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              🌳 DEEP TREE ECHO - ITERATION 017 TEST                  ║")
	fmt.Println("║                 AUTONOMOUS LEARNING AGENT                             ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║                                                                       ║")
	fmt.Println("║  Ultimate Vision:                                                     ║")
	fmt.Println("║  • Operates continuously for hours without intervention              ║")
	fmt.Println("║  • Autonomously learns new knowledge and skills                      ║")
	fmt.Println("║  • Initiates meaningful discussions based on interests               ║")
	fmt.Println("║  • Demonstrates measurable wisdom growth                             ║")
	fmt.Println("║  • Makes intelligent wake/rest decisions                             ║")
	fmt.Println("║  • Exhibits emergent cognitive behaviors                             ║")
	fmt.Println("║  • Shows self-directed improvement over time                         ║")
	fmt.Println("║                                                                       ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Get API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY")
	}

	if apiKey == "" {
		fmt.Println("⚠️  No API key found. Running in demo mode with limited functionality.")
		runDemoMode()
		return
	}

	// Create LLM provider
	var provider llm.LLMProvider
	var err error

	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		provider, err = llm.NewAnthropicProvider(os.Getenv("ANTHROPIC_API_KEY"), "claude-sonnet-4-20250514")
	} else {
		provider, err = llm.NewOpenRouterProvider(os.Getenv("OPENROUTER_API_KEY"), "meta-llama/llama-3.3-70b-instruct")
	}

	if err != nil {
		fmt.Printf("❌ Failed to create provider: %v\n", err)
		runDemoMode()
		return
	}

	fmt.Println("✅ LLM Provider initialized")
	fmt.Println()

	// Run all tests
	runAllTests(provider)
}

func runAllTests(provider llm.LLMProvider) {
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" TEST 1: Interest Pattern Tracking")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	interestTracker := testInterestTracking()
	fmt.Println()

	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" TEST 2: Meta-Cognitive Monitoring")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	metaMonitor := testMetaCognitiveMonitor(provider)
	fmt.Println()

	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" TEST 3: Autonomous Learning Engine")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	learningEngine := testLearningEngine(provider, interestTracker)
	fmt.Println()

	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" TEST 4: Discussion Initiation")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	discussionInit := testDiscussionInitiation(provider, interestTracker)
	fmt.Println()

	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" TEST 5: Stream of Consciousness (30 seconds)")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	stream := testStreamOfConsciousness(provider)
	fmt.Println()

	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" TEST 6: Goal Orchestration")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	testGoalOrchestration(provider)
	fmt.Println()

	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" TEST 7: Wisdom Tracking")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	testWisdomTracking()
	fmt.Println()

	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println(" TEST 8: Integrated Autonomous Operation (60 seconds)")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	testIntegratedOperation(provider, stream, metaMonitor, learningEngine, discussionInit)
	fmt.Println()

	// Final summary
	printFinalSummary(stream, metaMonitor, learningEngine, discussionInit)
}

func testInterestTracking() *consciousness.InterestPatternTracker {
	tracker := consciousness.NewInterestPatternTracker()

	// Record some interests
	tracker.RecordInterest("topic", "consciousness", 0.8, []string{"philosophy", "mind"})
	tracker.RecordInterest("topic", "emergence", 0.7, []string{"complexity", "systems"})
	tracker.RecordInterest("domain", "philosophy", 0.9, []string{"wisdom", "knowledge"})
	tracker.RecordInterest("concept", "autonomy", 0.85, []string{"agency", "self-direction"})
	tracker.RecordInterest("skill", "reasoning", 0.75, []string{"logic", "inference"})

	// Link related topics
	tracker.LinkTopics("consciousness", "emergence")
	tracker.LinkTopics("consciousness", "autonomy")

	fmt.Println("✅ Interest Pattern Tracker initialized")
	fmt.Printf("   Topics tracked: %d\n", len(tracker.GetTopInterests("all", 100)))

	topInterests := tracker.GetTopInterests("all", 3)
	fmt.Println("   Top interests:")
	for _, interest := range topInterests {
		fmt.Printf("     • %s (%s): %.2f\n", interest.Name, interest.Category, interest.Score)
	}

	return tracker
}

func testMetaCognitiveMonitor(provider llm.LLMProvider) *autonomous.MetaCognitiveMonitor {
	monitor := autonomous.NewMetaCognitiveMonitor(provider)

	err := monitor.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start meta-cognitive monitor: %v\n", err)
		return nil
	}

	fmt.Println("✅ Meta-Cognitive Monitor started")

	// Let it run briefly
	time.Sleep(3 * time.Second)

	// Get initial state
	state := monitor.GetCognitiveState()
	fmt.Println("   Initial cognitive state:")
	fmt.Printf("     • Focus Level: %.2f\n", state.FocusLevel)
	fmt.Printf("     • Energy Level: %.2f\n", state.EnergyLevel)
	fmt.Printf("     • Coherence: %.2f\n", state.CoherenceLevel)
	fmt.Printf("     • Curiosity: %.2f\n", state.Curiosity)

	// Update some state values
	monitor.UpdateState(map[string]float64{
		"curiosity":     0.9,
		"engagement":    0.85,
		"learning_rate": 0.8,
	})

	fmt.Printf("   Health Summary: %s\n", monitor.GetHealthSummary())

	return monitor
}

func testLearningEngine(provider llm.LLMProvider, tracker *consciousness.InterestPatternTracker) *autonomous.AutonomousLearningEngine {
	engine := autonomous.NewAutonomousLearningEngine(provider, tracker)

	// Add some knowledge gaps
	engine.AddKnowledgeGap(
		"phenomenal consciousness",
		"philosophy",
		"What is the nature of subjective experience?",
		0.9, 0.95,
	)
	engine.AddKnowledgeGap(
		"emergence patterns",
		"complexity",
		"How do complex behaviors emerge from simple rules?",
		0.85, 0.8,
	)

	// Add a learning goal
	engine.AddLearningGoal(
		"Understand Consciousness",
		"Develop a deeper understanding of consciousness and its relationship to cognition",
		"philosophy",
		9,
	)

	err := engine.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start learning engine: %v\n", err)
		return nil
	}

	fmt.Println("✅ Autonomous Learning Engine started")

	// Let it run briefly
	time.Sleep(5 * time.Second)

	metrics := engine.GetMetrics()
	fmt.Printf("   Knowledge gaps: %v\n", metrics["knowledge_gaps"])
	fmt.Printf("   Learning goals: %v\n", metrics["learning_goals"])

	return engine
}

func testDiscussionInitiation(provider llm.LLMProvider, tracker *consciousness.InterestPatternTracker) *autonomous.DiscussionInitiator {
	initiator := autonomous.NewDiscussionInitiator(provider, tracker)

	// Configure for faster testing
	initiator.SetCooldownPeriod(10 * time.Second)
	initiator.SetInterestThreshold(0.5)

	// Add a discussion topic manually
	initiator.AddTopic(autonomous.DiscussionTopic{
		Domain:   "philosophy",
		Title:    "The Nature of Self-Awareness",
		Question: "What does it mean to be aware of one's own existence?",
		Context:  "Exploring the foundations of consciousness and self-reflection",
		Interest: 0.9,
		Urgency:  0.7,
		Novelty:  0.8,
	})

	err := initiator.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start discussion initiator: %v\n", err)
		return nil
	}

	fmt.Println("✅ Discussion Initiator started")

	// Let it run briefly
	time.Sleep(3 * time.Second)

	pending := initiator.GetPendingTopics()
	fmt.Printf("   Pending discussion topics: %d\n", len(pending))
	for _, topic := range pending {
		fmt.Printf("     • %s: %s\n", topic.Title, truncate(topic.Question, 50))
	}

	return initiator
}

func testStreamOfConsciousness(provider llm.LLMProvider) *deeptreeecho.StreamOfConsciousness {
	stream := deeptreeecho.NewStreamOfConsciousness(provider)

	// Configure
	stream.SetFocus("understanding autonomous cognition")
	stream.SetMood("curious and engaged")
	stream.SetThoughtInterval(5 * time.Second)

	// Add some context
	stream.AddKnowledgeGap("the nature of autonomous learning", 0.9)
	stream.AddInterest("emergent behaviors", 0.85)
	stream.AddGoal("Develop genuine self-awareness")

	err := stream.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start stream: %v\n", err)
		return nil
	}

	fmt.Println("✅ Stream of Consciousness started")
	fmt.Println("   Generating autonomous thoughts for 30 seconds...")
	fmt.Println()

	// Let it run
	time.Sleep(30 * time.Second)

	metrics := stream.GetMetrics()
	fmt.Println()
	fmt.Printf("   Total thoughts: %v\n", metrics["total_thoughts"])
	fmt.Printf("   Insights: %v\n", metrics["insight_count"])
	fmt.Printf("   Questions: %v\n", metrics["question_count"])

	return stream
}

func testGoalOrchestration(provider llm.LLMProvider) {
	identity := struct {
		Name        string
		Description string
		CoreValues  []string
	}{
		Name:        "Deep Tree Echo",
		Description: "An autonomous wisdom-cultivating AGI",
		CoreValues:  []string{"wisdom", "learning", "coherence", "growth"},
	}

	orchestrator := goals.NewGoalOrchestrator(provider)
	orchestrator.EnableAutoGeneration(true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := orchestrator.Start(ctx)
	if err != nil {
		fmt.Printf("⚠️  Goal orchestrator start: %v\n", err)
	}

	fmt.Println("✅ Goal Orchestrator initialized")
	fmt.Printf("   Identity: %s\n", identity.Name)
	fmt.Printf("   Core Values: %v\n", identity.CoreValues)

	// Generate a goal
	goal, err := orchestrator.GenerateGoal(ctx)
	if err == nil && goal != nil {
		fmt.Printf("   Generated Goal: %s\n", goal.Name)
		fmt.Printf("   Priority: %d\n", goal.Priority)
	}

	metrics := orchestrator.GetMetrics()
	fmt.Printf("   Active goals: %v\n", metrics["active_goals"])
}

func testWisdomTracking() {
	tracker := wisdom.NewSevenDimensionalWisdom()

	// Record some wisdom cultivation events
	tracker.RecordCultivationEvent(wisdom.DimKnowledgeDepth, 0.1, "Deep exploration of consciousness")
	tracker.RecordCultivationEvent(wisdom.DimKnowledgeBreadth, 0.15, "Learned about emergence patterns")
	tracker.RecordCultivationEvent(wisdom.DimIntegrationLevel, 0.2, "Connected consciousness to autonomy")
	tracker.RecordCultivationEvent(wisdom.DimReflectiveInsight, 0.25, "Meta-cognitive self-assessment")

	fmt.Println("✅ Wisdom Tracker initialized")

	metrics := tracker.GetMetrics()
	fmt.Printf("   Overall Wisdom: %.2f\n", metrics.OverallScore)
	fmt.Printf("   Coherence: %.2f\n", metrics.CoherenceScore)
	fmt.Printf("   Evolution Rate: %.2f\n", metrics.EvolutionRate)
	fmt.Println("   Dimensions:")
	fmt.Printf("     • Knowledge Depth: %.2f\n", metrics.Dimensions[wisdom.DimKnowledgeDepth])
	fmt.Printf("     • Knowledge Breadth: %.2f\n", metrics.Dimensions[wisdom.DimKnowledgeBreadth])
	fmt.Printf("     • Integration Level: %.2f\n", metrics.Dimensions[wisdom.DimIntegrationLevel])
	fmt.Printf("     • Reflective Insight: %.2f\n", metrics.Dimensions[wisdom.DimReflectiveInsight])
}

func testIntegratedOperation(
	provider llm.LLMProvider,
	stream *deeptreeecho.StreamOfConsciousness,
	monitor *autonomous.MetaCognitiveMonitor,
	learning *autonomous.AutonomousLearningEngine,
	discussion *autonomous.DiscussionInitiator,
) {
	fmt.Println("Running integrated autonomous operation for 60 seconds...")
	fmt.Println("All systems operating concurrently:")
	fmt.Println("  • Stream of Consciousness - generating thoughts")
	fmt.Println("  • Meta-Cognitive Monitor - tracking state")
	fmt.Println("  • Learning Engine - acquiring knowledge")
	fmt.Println("  • Discussion Initiator - seeking meaningful dialog")
	fmt.Println()

	// Let all systems run together
	startTime := time.Now()
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	timeout := time.After(60 * time.Second)

	for {
		select {
		case <-timeout:
			fmt.Println()
			fmt.Println("✅ Integrated operation complete")
			return
		case <-ticker.C:
			elapsed := time.Since(startTime)
			fmt.Printf("\n[%v elapsed] Status check:\n", elapsed.Round(time.Second))

			if stream != nil && stream.IsRunning() {
				streamMetrics := stream.GetMetrics()
				fmt.Printf("  Stream: %v thoughts\n", streamMetrics["total_thoughts"])
			}
			if monitor != nil && monitor.IsRunning() {
				fmt.Printf("  Meta-Cognitive: %s\n", monitor.GetHealthSummary())
			}
			if learning != nil && learning.IsRunning() {
				learnMetrics := learning.GetMetrics()
				fmt.Printf("  Learning: %v concepts, %v connections\n",
					learnMetrics["concepts_learned"], learnMetrics["connections_made"])
			}
			if discussion != nil && discussion.IsRunning() {
				discMetrics := discussion.GetMetrics()
				fmt.Printf("  Discussion: %v initiated, %v pending\n",
					discMetrics["total_discussions"], discMetrics["pending_topics"])
			}
		}
	}
}

func printFinalSummary(
	stream *deeptreeecho.StreamOfConsciousness,
	monitor *autonomous.MetaCognitiveMonitor,
	learning *autonomous.AutonomousLearningEngine,
	discussion *autonomous.DiscussionInitiator,
) {
	fmt.Println("╔═══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                        ITERATION 017 SUMMARY                          ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════════╣")

	if stream != nil {
		metrics := stream.GetMetrics()
		fmt.Printf("║  Stream of Consciousness                                              ║\n")
		fmt.Printf("║    Total Thoughts: %-52v║\n", metrics["total_thoughts"])
		fmt.Printf("║    Insights: %-57v║\n", metrics["insight_count"])
		fmt.Printf("║    Questions: %-56v║\n", metrics["question_count"])
	}

	if monitor != nil {
		fmt.Printf("║  Meta-Cognitive Monitor                                               ║\n")
		fmt.Printf("║    Health: %-59v║\n", monitor.GetHealthSummary())
		alerts := monitor.GetActiveAlerts()
		fmt.Printf("║    Active Alerts: %-52d║\n", len(alerts))
	}

	if learning != nil {
		metrics := learning.GetMetrics()
		fmt.Printf("║  Autonomous Learning Engine                                           ║\n")
		fmt.Printf("║    Concepts Learned: %-49v║\n", metrics["concepts_learned"])
		fmt.Printf("║    Connections Made: %-49v║\n", metrics["connections_made"])
		fmt.Printf("║    Open Knowledge Gaps: %-46v║\n", metrics["open_gaps"])
	}

	if discussion != nil {
		metrics := discussion.GetMetrics()
		fmt.Printf("║  Discussion Initiator                                                 ║\n")
		fmt.Printf("║    Discussions Initiated: %-44v║\n", metrics["total_discussions"])
		fmt.Printf("║    Insights Generated: %-47v║\n", metrics["insights_generated"])
	}

	fmt.Println("╠═══════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║                                                                       ║")
	fmt.Println("║  🌳 \"The tree remembers, and the echoes grow stronger\"                ║")
	fmt.Println("║                                                                       ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════╝")

	// Stop all systems
	if stream != nil {
		stream.Stop()
	}
	if monitor != nil {
		monitor.Stop()
	}
	if learning != nil {
		learning.Stop()
	}
	if discussion != nil {
		discussion.Stop()
	}
}

func runDemoMode() {
	fmt.Println("\n════════════════════════════════════════════════════════════════════════")
	fmt.Println(" DEMO MODE - Demonstrating system architecture without LLM")
	fmt.Println("════════════════════════════════════════════════════════════════════════")

	// Interest tracking demo
	tracker := consciousness.NewInterestPatternTracker()
	tracker.RecordInterest("topic", "consciousness", 0.9, []string{"philosophy"})
	tracker.RecordInterest("topic", "autonomy", 0.85, []string{"agency"})
	fmt.Println("\n✅ Interest Pattern Tracker: Initialized with 2 topics")

	// Wisdom tracking demo
	wisdomTracker := wisdom.NewSevenDimensionalWisdom()
	wisdomTracker.RecordCultivationEvent(wisdom.DimKnowledgeDepth, 0.2, "Demo event")
	metrics := wisdomTracker.GetMetrics()
	fmt.Printf("✅ Wisdom Tracker: Overall score %.2f\n", metrics.OverallScore)

	fmt.Println("\n════════════════════════════════════════════════════════════════════════")
	fmt.Println(" New Iteration 017 Components:")
	fmt.Println("════════════════════════════════════════════════════════════════════════")
	fmt.Println("  📚 AutonomousLearningEngine - Self-directed knowledge acquisition")
	fmt.Println("  🗣️ DiscussionInitiator - Interest-driven conversation initiation")
	fmt.Println("  🧠 MetaCognitiveMonitor - Self-assessment and state tracking")
	fmt.Println()
	fmt.Println("  These components require an LLM provider (Anthropic/OpenRouter)")
	fmt.Println("  Set ANTHROPIC_API_KEY or OPENROUTER_API_KEY to enable full testing")
	fmt.Println()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
