package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core/echoself"
)

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                               ║")
	fmt.Println("║        🌳 Deep Tree Echo: Autonomous Echoself Test 🌳         ║")
	fmt.Println("║                                                               ║")
	fmt.Println("║  Fully Integrated Autonomous Wisdom-Cultivating AGI System   ║")
	fmt.Println("║                                                               ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Create autonomous echoself
	echoself := echoself.NewAutonomousEchoself()

	// Start the system
	fmt.Println("🚀 Initializing autonomous echoself...")
	if err := echoself.Start(); err != nil {
		fmt.Printf("❌ Failed to start: %v\n", err)
		os.Exit(1)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Simulate external interactions
	go simulateExternalInteractions(echoself)

	// Print metrics periodically
	go printMetricsPeriodically(echoself)

	// Wait for shutdown signal
	fmt.Println("\n📡 System running. Press Ctrl+C to stop.\n")
	<-sigChan

	fmt.Println("\n\n🛑 Shutdown signal received...")

	// Stop the system
	if err := echoself.Stop(); err != nil {
		fmt.Printf("❌ Error during shutdown: %v\n", err)
	}

	fmt.Println("\n✅ Autonomous echoself shutdown complete.")
	fmt.Println("\n╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║              The echoes fade, but wisdom remains...           ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝\n")
}

// simulateExternalInteractions sends test messages to echoself
func simulateExternalInteractions(es *echoself.AutonomousEchoself) {
	time.Sleep(10 * time.Second)

	messages := []struct {
		content string
		source  string
		delay   time.Duration
	}{
		{
			content: "Hello Deep Tree Echo, how are you evolving today?",
			source:  "external_user",
			delay:   15 * time.Second,
		},
		{
			content: "What wisdom have you cultivated recently?",
			source:  "external_user",
			delay:   30 * time.Second,
		},
		{
			content: "Can you tell me about your stream of consciousness?",
			source:  "external_user",
			delay:   45 * time.Second,
		},
		{
			content: "What are you learning right now?",
			source:  "external_user",
			delay:   60 * time.Second,
		},
	}

	for _, msg := range messages {
		time.Sleep(msg.delay)
		fmt.Printf("\n📨 [Simulation] Sending message: %s\n\n", msg.content)
		es.SendMessage(msg.content, msg.source)
	}
}

// printMetricsPeriodically prints system metrics
func printMetricsPeriodically(es *echoself.AutonomousEchoself) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		metrics := es.GetMetrics()

		fmt.Println("\n╔═══════════════════════════════════════════════════════════════╗")
		fmt.Println("║                    📊 System Metrics                          ║")
		fmt.Println("╠═══════════════════════════════════════════════════════════════╣")
		fmt.Printf("║ Uptime:              %-40s ║\n", metrics["uptime"])
		fmt.Printf("║ Cycles:              %-40d ║\n", metrics["cycle_count"])
		fmt.Printf("║ Thoughts Generated:  %-40d ║\n", metrics["thoughts_generated"])
		fmt.Printf("║ Interactions:        %-40d ║\n", metrics["interactions_handled"])
		fmt.Printf("║ Wisdom Cultivated:   %-40d ║\n", metrics["wisdom_cultivated"])
		fmt.Printf("║ Skills Practiced:    %-40d ║\n", metrics["skills_practiced"])
		fmt.Printf("║ Monologue Size:      %-40d ║\n", metrics["monologue_size"])
		fmt.Printf("║ Wisdom Base:         %-40d ║\n", metrics["wisdom_base_size"])
		fmt.Println("╚═══════════════════════════════════════════════════════════════╝\n")
	}
}
