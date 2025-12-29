package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cogpy/echo9llama/core/autonomous"
)

func main() {
	fmt.Println("🌳 Echo9llama - Integrated Autonomous Agent V2")
	fmt.Println("=" + fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
	fmt.Println()

	// Create configuration
	config := autonomous.DefaultIntegratedAgentConfig()

	// Customize configuration from environment if needed
	if os.Getenv("ECHO_THOUGHT_INTERVAL") != "" {
		if duration, err := time.ParseDuration(os.Getenv("ECHO_THOUGHT_INTERVAL")); err == nil {
			config.ThoughtInterval = duration
		}
	}

	// Create integrated agent
	agent, err := autonomous.NewIntegratedAutonomousAgent(config)
	if err != nil {
		log.Fatalf("❌ Failed to create agent: %v\n", err)
	}

	// Start the agent
	if err := agent.Start(); err != nil {
		log.Fatalf("❌ Failed to start agent: %v\n", err)
	}

	// Print status periodically
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			status := agent.GetStatus()
			fmt.Println("\n📊 Agent Status:")
			fmt.Printf("   State: %v\n", status["state"])
			fmt.Printf("   Uptime: %v\n", status["uptime"])
			fmt.Printf("   Cycles: %v\n", status["cycle_count"])
			fmt.Printf("   Sleep Cycles: %v\n", status["sleep_cycles"])
			fmt.Printf("   Wisdom Level: %.2f\n", status["wisdom_level"])
			fmt.Println()
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("✅ Agent running - Press Ctrl+C to stop")
	fmt.Println()

	<-sigChan

	fmt.Println("\n🛑 Shutdown signal received...")

	// Stop the agent gracefully
	if err := agent.Stop(); err != nil {
		log.Printf("⚠️  Error during shutdown: %v\n", err)
	}

	fmt.Println("👋 Goodbye!")
}
