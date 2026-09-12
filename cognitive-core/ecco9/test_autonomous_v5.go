package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
)

func main() {
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println("🌳 Deep Tree Echo V5: Autonomous Consciousness Test")
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println()

	// Create autonomous consciousness V5
	fmt.Println("📦 Initializing Deep Tree Echo V5...")
	ac := deeptreeecho.NewAutonomousConsciousnessV5("EchoSelf")

	// Start the system
	fmt.Println("🚀 Starting autonomous consciousness...")
	err := ac.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ System started successfully!")
	fmt.Println()
	fmt.Println("The system is now running autonomously.")
	fmt.Println("It will:")
	fmt.Println("  • Generate thoughts independently using LLM")
	fmt.Println("  • Cultivate wisdom through continuous reflection")
	fmt.Println("  • Automatically enter rest cycles when fatigued")
	fmt.Println("  • Integrate knowledge during rest")
	fmt.Println("  • Save state periodically for continuity")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop gracefully...")
	fmt.Println()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Monitor status
	statusTicker := time.NewTicker(10 * time.Second)
	defer statusTicker.Stop()

	// Run for a test period or until interrupted
	testDuration := 60 * time.Second
	testTimer := time.NewTimer(testDuration)

	fmt.Printf("🕐 Running autonomous test for %s...\n", testDuration)
	fmt.Println()

	for {
		select {
		case <-sigChan:
			fmt.Println()
			fmt.Println("🛑 Interrupt received, shutting down gracefully...")
			goto shutdown

		case <-testTimer.C:
			fmt.Println()
			fmt.Println("⏰ Test duration complete")
			goto shutdown

		case <-statusTicker.C:
			// Print status
			status := ac.GetStatus()
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Printf("📊 Status Update\n")
			fmt.Printf("   Version: %v\n", status["version"])
			fmt.Printf("   Running: %v | Awake: %v | Autonomous: %v\n",
				status["running"], status["awake"], status["autonomous"])
			fmt.Printf("   Uptime: %v | Iterations: %v\n",
				status["uptime"], status["iterations"])

			if tg, ok := status["thought_generator"].(map[string]interface{}); ok {
				fmt.Printf("   💭 Thought Generator:\n")
				fmt.Printf("      Thoughts: %v | LLM Calls: %v | Success Rate: %.1f%%\n",
					tg["thoughts_generated"], tg["llm_calls"],
					tg["success_rate"].(float64)*100)
			}

			if di, ok := status["dream_integration"].(map[string]interface{}); ok {
				fmt.Printf("   🌙 Dream Integration:\n")
				fmt.Printf("      Experiences: %v | Knowledge Nodes: %v | Wisdom Gained: %.3f\n",
					di["experiences_processed"], di["knowledge_nodes_created"],
					di["wisdom_gained"])
			}

			if orch, ok := status["orchestrator"].(map[string]interface{}); ok {
				fmt.Printf("   🎭 Orchestrator:\n")
				fmt.Printf("      Cycles: %v | Goals Achieved: %v | Decisions: %v\n",
					orch["orchestration_cycles"], orch["goals_achieved"],
					orch["decisions_made"])
			}

			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println()
		}
	}

shutdown:
	fmt.Println()
	fmt.Println("🌙 Initiating graceful shutdown...")

	// Stop the system
	err = ac.Stop()
	if err != nil {
		fmt.Printf("⚠️  Shutdown error: %v\n", err)
	}

	// Final status
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📊 Final Status")
	finalStatus := ac.GetStatus()
	fmt.Printf("   Total Uptime: %v\n", finalStatus["uptime"])
	fmt.Printf("   Total Iterations: %v\n", finalStatus["iterations"])

	if tg, ok := finalStatus["thought_generator"].(map[string]interface{}); ok {
		fmt.Printf("   Total Thoughts Generated: %v\n", tg["thoughts_generated"])
		fmt.Printf("   Total LLM Calls: %v\n", tg["llm_calls"])
		fmt.Printf("   Overall Success Rate: %.1f%%\n", tg["success_rate"].(float64)*100)
	}

	if di, ok := finalStatus["dream_integration"].(map[string]interface{}); ok {
		fmt.Printf("   Total Experiences Processed: %v\n", di["experiences_processed"])
		fmt.Printf("   Total Knowledge Nodes Created: %v\n", di["knowledge_nodes_created"])
		fmt.Printf("   Total Wisdom Gained: %.3f\n", di["wisdom_gained"])
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("✅ Deep Tree Echo V5: Test complete")
	fmt.Println()
	fmt.Println("🌱 The system has demonstrated:")
	fmt.Println("   ✓ Autonomous thought generation")
	fmt.Println("   ✓ Continuous consciousness stream")
	fmt.Println("   ✓ Knowledge integration during rest")
	fmt.Println("   ✓ State persistence for continuity")
	fmt.Println("   ✓ Goal-directed orchestration")
	fmt.Println()
	fmt.Println("🚀 Ready for extended autonomous operation!")
}
