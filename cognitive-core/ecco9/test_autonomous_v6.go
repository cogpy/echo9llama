package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("🚀 Starting Deep Tree Echo V6 Test")
	log.Printf("=" + "==========================================================")

	// Create autonomous consciousness V6
	ac, err := deeptreeecho.NewAutonomousConsciousnessV6("Deep Tree Echo")
	if err != nil {
		log.Fatalf("Failed to create autonomous consciousness: %v", err)
	}

	// Start the system
	if err := ac.Start(); err != nil {
		log.Fatalf("Failed to start autonomous consciousness: %v", err)
	}

	// Monitor status
	go monitorStatus(ac)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("✅ System running. Press Ctrl+C to stop.")
	log.Printf("=" + "==========================================================")

	<-sigChan

	log.Printf("\n🛑 Shutdown signal received")

	// Stop the system
	if err := ac.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Printf("✅ Test complete")
}

func monitorStatus(ac *deeptreeecho.AutonomousConsciousnessV6) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		status := ac.GetStatus()
		log.Printf("\n" + "=" + "========== STATUS REPORT ==========")
		log.Printf("Running: %v", status["running"])
		log.Printf("Awake: %v", status["awake"])
		log.Printf("Thinking: %v", status["thinking"])
		log.Printf("Dreaming: %v", status["dreaming"])
		log.Printf("Uptime: %v", status["uptime"])
		log.Printf("Iterations: %v", status["iterations"])
		log.Printf("Thoughts: %v", status["thought_count"])
		log.Printf("Autonomous Thoughts: %v", status["autonomous_thoughts"])
		log.Printf("Working Memory: %v items", status["working_memory_items"])
		log.Printf("Identity Coherence: %.3f", status["identity_coherence"])
		log.Printf("Fatigue: %.2f", status["fatigue"])
		log.Printf("Interests: %v", status["interests"])
		log.Printf("=" + "====================================\n")

		// Print summary
		fmt.Printf("\n📊 Quick Stats: Thoughts=%v | Coherence=%.3f | Fatigue=%.2f\n\n",
			status["thought_count"],
			status["identity_coherence"],
			status["fatigue"])
	}
}
