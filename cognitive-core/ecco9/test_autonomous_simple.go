package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core/autonomous"
)

func main() {
	fmt.Println("🌳 Deep Tree Echo: Autonomous Agent Test")
	fmt.Println("==========================================")
	fmt.Println()

	// Check for OpenAI API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("⚠️  Warning: OPENAI_API_KEY not set")
		fmt.Println("   The system will use fallback templates for thought generation")
		fmt.Println("   Set OPENAI_API_KEY environment variable for LLM-powered thoughts")
		fmt.Println()
	} else {
		fmt.Println("✅ OpenAI API key detected")
		fmt.Println()
	}

	// Create configuration
	config := autonomous.DefaultConfig()
	config.StateDirectory = "./echo_state_test"
	config.MinWakeDuration = 2 * time.Minute
	config.MaxWakeDuration = 5 * time.Minute
	config.MinRestDuration = 30 * time.Second
	config.MaxRestDuration = 1 * time.Minute

	// Create orchestrator
	fmt.Println("🔧 Initializing autonomous agent orchestrator...")
	orchestrator, err := autonomous.NewAgentOrchestrator("Deep Tree Echo", config)
	if err != nil {
		fmt.Printf("❌ Failed to create orchestrator: %v\n", err)
		os.Exit(1)
	}

	// Start autonomous operation
	fmt.Println("🚀 Starting autonomous operation...")
	if err := orchestrator.Start(); err != nil {
		fmt.Printf("❌ Failed to start orchestrator: %v\n", err)
		os.Exit(1)
	}

	// Status monitoring
	go monitorStatus(orchestrator)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("✅ Autonomous agent running. Press Ctrl+C to stop.")
	fmt.Println()

	<-sigChan

	fmt.Println("\n🛑 Shutdown signal received...")

	// Stop orchestrator
	if err := orchestrator.Stop(); err != nil {
		fmt.Printf("⚠️  Error during shutdown: %v\n", err)
	}

	fmt.Println("👋 Goodbye!")
}

// monitorStatus periodically prints status
func monitorStatus(orchestrator *autonomous.AgentOrchestrator) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		orchestrator.PrintDetailedStatus()
	}
}
