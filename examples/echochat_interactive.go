package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/EchoCog/echollama/api"
	"github.com/EchoCog/echollama/orchestration"
)

// EchoChatInteractiveDemo starts an interactive EchoChat session
func main() {
	fmt.Println("🌊 EchoChat - Interactive Deep Tree Echo Shell Assistant")
	fmt.Println("========================================================")
	fmt.Println("Starting interactive session with Deep Tree Echo intelligence...")
	fmt.Println()

	// Initialize orchestration engine
	client := api.Client{}
	engine := orchestration.NewEngine(client)
	ctx := context.Background()

	// Register default tools and plugins
	orchestration.RegisterDefaultTools(engine)
	orchestration.RegisterDefaultPlugins(engine)

	// Initialize Deep Tree Echo System
	fmt.Println("🧠 Initializing Deep Tree Echo...")
	err := engine.InitializeDeepTreeEcho(ctx)
	if err != nil {
		log.Printf("Warning: Failed to initialize Deep Tree Echo: %v", err)
		fmt.Println("⚠️  Running in limited mode without Deep Tree Echo")
	} else {
		fmt.Println("✅ Deep Tree Echo system ready")
	}

	// Show initial status
	fmt.Println("\n📊 Initial Deep Tree Echo Status:")
	status := engine.GetDeepTreeEchoStatus()
	
	if health, ok := status["system_health"].(string); ok {
		fmt.Printf("   🏥 System Health: %s\n", health)
	}
	
	if coreStatus, ok := status["core_status"].(string); ok {
		fmt.Printf("   🧠 Core Status: %s\n", coreStatus)
	}

	// Create EchoChat instance
	echoChat := orchestration.NewEchoChat(engine)
	
	// Start interactive session
	fmt.Println()
	if err := echoChat.StartInteractiveSession(ctx); err != nil {
		fmt.Printf("Error in interactive session: %v\n", err)
		os.Exit(1)
	}
}