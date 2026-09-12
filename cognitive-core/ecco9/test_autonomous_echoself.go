package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core"
)

func main() {
	fmt.Println("🌳 Deep Tree Echo: Autonomous Echoself Test")
	fmt.Println("=" + repeat("=", 50))
	fmt.Println()

	// Create configuration
	config := core.DefaultEchoselfConfig()
	config.PersistenceDir = "./echoself_data"

	// Ensure persistence directory exists
	os.MkdirAll(config.PersistenceDir, 0755)

	// Create autonomous echoself
	fmt.Println("🔧 Initializing Autonomous Echoself...")
	echoself := core.NewAutonomousEchoself(config)

	// Start autonomous operation
	fmt.Println("🚀 Starting autonomous operation...")
	if err := echoself.Start(); err != nil {
		fmt.Printf("❌ Error starting echoself: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("✅ Echoself is now running autonomously!")
	fmt.Println("   - Stream of consciousness active")
	fmt.Println("   - Interest patterns developing")
	fmt.Println("   - Wake/rest cycles managing autonomously")
	fmt.Println("   - Dream cycles consolidating knowledge")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop...")
	fmt.Println()

	// Simulate some external interactions
	go simulateInteractions(echoself)

	// Monitor and display status
	go monitorStatus(echoself)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	fmt.Println("🛑 Shutdown signal received...")

	// Stop echoself
	if err := echoself.Stop(); err != nil {
		fmt.Printf("❌ Error stopping echoself: %v\n", err)
	}

	// Display final metrics
	displayFinalMetrics(echoself)

	fmt.Println()
	fmt.Println("👋 Echoself shutdown complete. Until next time...")
}

func simulateInteractions(echoself *core.AutonomousEchoself) {
	// Wait for startup
	time.Sleep(5 * time.Second)

	interactions := []struct {
		input     string
		inputType string
		delay     time.Duration
	}{
		{
			input:     "Exploring cognitive architectures and their potential",
			inputType: "topic",
			delay:     10 * time.Second,
		},
		{
			input:     "How do memory systems consolidate knowledge?",
			inputType: "question",
			delay:     15 * time.Second,
		},
		{
			input:     "Pattern recognition in complex systems",
			inputType: "topic",
			delay:     20 * time.Second,
		},
		{
			input:     "The nature of consciousness and self-awareness",
			inputType: "topic",
			delay:     25 * time.Second,
		},
		{
			input:     "Wisdom cultivation through experience",
			inputType: "topic",
			delay:     30 * time.Second,
		},
	}

	for _, interaction := range interactions {
		time.Sleep(interaction.delay)

		fmt.Printf("\n📥 External input: %s\n", interaction.input)
		echoself.ProcessExternalInput(interaction.input, interaction.inputType)

		// Evaluate if echoself would engage in discussion
		decision := echoself.EvaluateDiscussionTopic(interaction.input)
		if decision.ShouldEngage {
			fmt.Printf("   ✅ Would engage: %s (confidence: %.2f)\n", decision.Reason, decision.Confidence)
		} else {
			fmt.Printf("   ⏭️  Would not engage: %s\n", decision.Reason)
		}
	}
}

func monitorStatus(echoself *core.AutonomousEchoself) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		state := echoself.GetCurrentState()
		metrics := echoself.GetMetrics()

		fmt.Println()
		fmt.Println("📊 Status Update")
		fmt.Println("   " + repeat("-", 45))
		fmt.Printf("   State: %s\n", state)
		fmt.Printf("   Uptime: %v\n", metrics["uptime"])
		fmt.Printf("   Cycles: %d\n", metrics["cycles_completed"])
		fmt.Printf("   Wisdom: %d\n", metrics["wisdom_cultivated"])
		fmt.Printf("   Actions: %d\n", metrics["autonomous_actions"])

		// Stream of consciousness metrics
		if socMetrics, ok := metrics["stream_of_consciousness"].(map[string]interface{}); ok {
			fmt.Printf("   Thoughts: %v\n", socMetrics["thoughts_generated"])
			fmt.Printf("   Insights: %v\n", socMetrics["insights_generated"])
			fmt.Printf("   Questions: %v\n", socMetrics["questions_asked"])
		}

		// Interest metrics
		if interestMetrics, ok := metrics["interest_patterns"].(map[string]interface{}); ok {
			fmt.Printf("   Interests: %v\n", interestMetrics["total_interests"])
			fmt.Printf("   Curiosity: %.2f\n", interestMetrics["curiosity_level"])
		}

		// Display recent thoughts
		recentThoughts := echoself.GetRecentThoughts(3)
		if len(recentThoughts) > 0 {
			fmt.Println("   Recent thoughts:")
			for _, thought := range recentThoughts {
				fmt.Printf("     • %s\n", truncate(thought.Content, 60))
			}
		}

		// Display top interests
		topInterests := echoself.GetTopInterests(3)
		if len(topInterests) > 0 {
			fmt.Println("   Top interests:")
			for _, interest := range topInterests {
				fmt.Printf("     • %s (%.2f)\n", interest.Name, interest.Salience)
			}
		}

		fmt.Println()
	}
}

func displayFinalMetrics(echoself *core.AutonomousEchoself) {
	fmt.Println()
	fmt.Println("📈 Final Metrics")
	fmt.Println("=" + repeat("=", 50))

	metrics := echoself.GetMetrics()

	fmt.Printf("Uptime: %v\n", metrics["uptime"])
	fmt.Printf("Cycles Completed: %d\n", metrics["cycles_completed"])
	fmt.Printf("Wisdom Cultivated: %d\n", metrics["wisdom_cultivated"])
	fmt.Printf("Autonomous Actions: %d\n", metrics["autonomous_actions"])
	fmt.Println()

	// Stream of consciousness
	if socMetrics, ok := metrics["stream_of_consciousness"].(map[string]interface{}); ok {
		fmt.Println("Stream of Consciousness:")
		fmt.Printf("  Thoughts Generated: %v\n", socMetrics["thoughts_generated"])
		fmt.Printf("  Insights Generated: %v\n", socMetrics["insights_generated"])
		fmt.Printf("  Questions Asked: %v\n", socMetrics["questions_asked"])
		fmt.Printf("  History Size: %v\n", socMetrics["history_size"])
		fmt.Println()
	}

	// Interest patterns
	if interestMetrics, ok := metrics["interest_patterns"].(map[string]interface{}); ok {
		fmt.Println("Interest Patterns:")
		fmt.Printf("  Total Interests: %v\n", interestMetrics["total_interests"])
		fmt.Printf("  Average Strength: %.2f\n", interestMetrics["avg_strength"])
		fmt.Printf("  Average Salience: %.2f\n", interestMetrics["avg_salience"])
		fmt.Printf("  Curiosity Level: %.2f\n", interestMetrics["curiosity_level"])
		fmt.Println()
	}

	// Discussions
	if discussionMetrics, ok := metrics["discussions"].(map[string]interface{}); ok {
		fmt.Println("Discussions:")
		fmt.Printf("  Joined: %v\n", discussionMetrics["discussions_joined"])
		fmt.Printf("  Initiated: %v\n", discussionMetrics["discussions_initiated"])
		fmt.Printf("  Messages Processed: %v\n", discussionMetrics["messages_processed"])
		fmt.Println()
	}

	// Dream cycles
	if dreamMetrics, ok := metrics["dream_cycles"].(map[string]interface{}); ok {
		fmt.Println("Dream Cycles:")
		fmt.Printf("  Consolidation Cycles: %v\n", dreamMetrics["consolidation_cycles"])
		fmt.Printf("  Wisdom Generated: %v\n", dreamMetrics["wisdom_generated"])
		fmt.Printf("  Insights Integrated: %v\n", dreamMetrics["insights_integrated"])
		fmt.Println()
	}

	// Display extracted wisdom
	wisdom := echoself.GetExtractedWisdom()
	if len(wisdom) > 0 {
		fmt.Println("Extracted Wisdom:")
		for i, w := range wisdom {
			if i >= 5 {
				break
			}
			fmt.Printf("  %d. %s (confidence: %.2f)\n", i+1, w.Content, w.Confidence)
		}
		fmt.Println()
	}

	// Display final thoughts
	recentThoughts := echoself.GetRecentThoughts(5)
	if len(recentThoughts) > 0 {
		fmt.Println("Final Thoughts:")
		for i, thought := range recentThoughts {
			fmt.Printf("  %d. [%s] %s\n", i+1, thought.Type, thought.Content)
		}
		fmt.Println()
	}

	// Display top interests
	topInterests := echoself.GetTopInterests(5)
	if len(topInterests) > 0 {
		fmt.Println("Top Interests:")
		for i, interest := range topInterests {
			fmt.Printf("  %d. %s (strength: %.2f, salience: %.2f)\n",
				i+1, interest.Name, interest.Strength, interest.Salience)
		}
		fmt.Println()
	}
}

// Helper functions

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
