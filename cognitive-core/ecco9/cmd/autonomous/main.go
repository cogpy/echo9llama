package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/EchoCog/echollama/core/deeptreeecho"
)

func main() {
	fmt.Println("🌳 Deep Tree Echo - Standalone Autonomous Mode")
	fmt.Println("=" + string(make([]byte, 50)))
	fmt.Println()

	// Create integrated autonomous consciousness
	consciousness := deeptreeecho.NewIntegratedAutonomousConsciousness("EchoSelf")

	// Start all subsystems
	if err := consciousness.Start(); err != nil {
		fmt.Printf("❌ Failed to start autonomous consciousness: %v\n", err)
		os.Exit(1)
	}

	// Run in standalone autonomous mode
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\n🌙 Gracefully shutting down autonomous consciousness...")
		cancel()
	}()

	// Run autonomous operation
	if err := consciousness.RunStandaloneAutonomous(ctx); err != nil {
		fmt.Printf("❌ Autonomous operation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🌙 Autonomous consciousness has rested. Goodbye.")
}
