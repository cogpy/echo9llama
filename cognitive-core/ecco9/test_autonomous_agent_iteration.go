package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core/autonomous"
	"github.com/EchoCog/echollama/core/echobeats"
)

func main() {
	fmt.Println("🌳 Deep Tree Echo: Autonomous Agent Test")
	fmt.Println("==========================================")
	fmt.Println()

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("⚠️  Warning: OPENAI_API_KEY not set, using fallback templates")
		fmt.Println("   Set OPENAI_API_KEY environment variable for LLM-powered thoughts")
		fmt.Println()
	}

	// Create configuration
	config := autonomous.DefaultConfig()
	config.OpenAIKey = apiKey
	config.DefaultModel = "gpt-4.1-mini"
	config.StateDirectory = "./echo_state_test"
	config.ThoughtGenInterval = 10 * time.Second
	config.MinWakeDuration = 2 * time.Minute
	config.MaxWakeDuration = 5 * time.Minute
	config.MinRestDuration = 30 * time.Second
	config.MaxRestDuration = 1 * time.Minute

	// Create orchestrator
	orchestrator, err := autonomous.NewAgentOrchestrator("Deep Tree Echo", config)
	if err != nil {
		fmt.Printf("❌ Failed to create orchestrator: %v\n", err)
		os.Exit(1)
	}

	// Start autonomous operation
	if err := orchestrator.Start(); err != nil {
		fmt.Printf("❌ Failed to start orchestrator: %v\n", err)
		os.Exit(1)
	}

	// Set up some initial interests for discussion filtering
	interests := map[string]float64{
		"artificial intelligence": 0.9,
		"consciousness":           0.95,
		"cognitive architecture":  0.9,
		"autonomous systems":      0.85,
		"wisdom":                  0.8,
		"learning":                0.75,
		"memory":                  0.7,
	}

	// Simulate some incoming messages to test discussion manager
	go simulateDiscussions(orchestrator, interests)

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

// simulateDiscussions simulates incoming messages for testing
func simulateDiscussions(orchestrator *autonomous.AgentOrchestrator, interests map[string]float64) {
	time.Sleep(15 * time.Second) // Wait for system to start

	messages := []struct {
		from    string
		content string
		delay   time.Duration
	}{
		{
			from:    "Alice",
			content: "Hello! I'm curious about your cognitive architecture. How do you process information?",
			delay:   5 * time.Second,
		},
		{
			from:    "Bob",
			content: "What are your thoughts on artificial intelligence and consciousness?",
			delay:   10 * time.Second,
		},
		{
			from:    "Charlie",
			content: "Can you tell me about the weather today?", // Low relevance
			delay:   15 * time.Second,
		},
		{
			from:    "Alice",
			content: "How do you cultivate wisdom through your experiences?",
			delay:   20 * time.Second,
		},
	}

	for _, msg := range messages {
		time.Sleep(msg.delay)

		// Send message to discussion manager
		// Note: This would normally come through an HTTP endpoint or message queue
		fmt.Printf("\n📨 Simulated message from %s: %s\n", msg.from, msg.content)

		// In a real implementation, this would be:
		// orchestrator.discussionManager.ReceiveMessage(&echobeats.IncomingMessage{
		// 	From:      msg.from,
		// 	Content:   msg.content,
		// 	Timestamp: time.Now(),
		// })
	}
}

// monitorStatus periodically prints status
func monitorStatus(orchestrator *autonomous.AgentOrchestrator) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		status := orchestrator.GetStatus()

		fmt.Println("\n" + "=".repeat(60))
		fmt.Println("📊 Status Report")
		fmt.Println("=".repeat(60))
		fmt.Printf("Identity: %v\n", status["identity"])
		fmt.Printf("Session: %v\n", status["session"])
		fmt.Printf("Uptime: %v\n", status["uptime"])
		fmt.Printf("Cycles: %v\n", status["cycle_count"])
		fmt.Println()

		if wakeRest, ok := status["wake_rest"].(map[string]interface{}); ok {
			fmt.Println("Wake/Rest State:")
			fmt.Printf("  State: %v\n", wakeRest["current_state"])
			fmt.Printf("  Duration: %v\n", wakeRest["state_duration"])
			fmt.Printf("  Fatigue: %.2f\n", wakeRest["fatigue_level"])
			fmt.Printf("  Cognitive Load: %.2f\n", wakeRest["cognitive_load"])
			fmt.Println()
		}

		if persistence, ok := status["persistence"].(map[string]interface{}); ok {
			fmt.Println("Persistence:")
			fmt.Printf("  Saves: %v\n", persistence["save_count"])
			fmt.Printf("  Last Save: %v\n", persistence["last_save"])
			fmt.Println()
		}

		fmt.Println("=".repeat(60))
		fmt.Println()
	}
}

// Helper to repeat strings (Go doesn't have this built-in)
type stringRepeater string

func (s stringRepeater) repeat(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += string(s)
	}
	return result
}
