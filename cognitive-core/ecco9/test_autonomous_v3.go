package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core"
)

func main() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("🌳 Deep Tree Echo V3 - Autonomous System Test")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Create configuration
	config := core.DefaultEchoselfConfigV3()

	// Override for testing
	config.ThoughtInterval = 10 * time.Second
	config.ReflectionInterval = 30 * time.Second
	config.IntrospectionInterval = 1 * time.Minute
	config.PersistenceInterval = 20 * time.Second

	fmt.Println("Configuration:")
	fmt.Printf("  Persistence DB: %s\n", config.PersistenceDBPath)
	fmt.Printf("  Repository Root: %s\n", config.RepositoryRoot)
	fmt.Printf("  Thought Interval: %v\n", config.ThoughtInterval)
	fmt.Printf("  Reflection Interval: %v\n", config.ReflectionInterval)
	fmt.Printf("  Introspection Interval: %v\n", config.IntrospectionInterval)
	fmt.Println()

	// Create autonomous system
	fmt.Println("Initializing autonomous system...")
	ae, err := core.NewAutonomousEchoselfV3(config)
	if err != nil {
		log.Fatalf("❌ Failed to create autonomous system: %v", err)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))

	// Start the system
	if err := ae.Start(); err != nil {
		log.Fatalf("❌ Failed to start system: %v", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start statistics reporter
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				printStats(ae)
			}
		}
	}()

	// Wait for interrupt signal
	fmt.Println()
	fmt.Println("System running. Press Ctrl+C to stop...")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	<-sigChan

	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("Shutdown signal received...")

	// Stop the system
	if err := ae.Stop(); err != nil {
		log.Printf("⚠️  Error during shutdown: %v", err)
	}

	// Print final statistics
	fmt.Println()
	printStats(ae)

	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("✅ Autonomous system test completed")
	fmt.Println(strings.Repeat("=", 80))
}

func printStats(ae *core.AutonomousEchoselfV3) {
	stats := ae.GetStats()

	fmt.Println(strings.Repeat("-", 80))
	fmt.Println("📊 System Statistics")
	fmt.Println(strings.Repeat("-", 80))

	fmt.Printf("State: %v\n", stats["state"])
	fmt.Printf("Running: %v\n", stats["is_running"])
	fmt.Printf("Uptime: %v\n", stats["uptime"])
	fmt.Printf("Cycles: %v\n", stats["cycle_count"])
	fmt.Println()

	fmt.Println("Cognitive Activity:")
	fmt.Printf("  Thoughts Generated: %v\n", stats["thoughts_generated"])
	fmt.Printf("  Memories Formed: %v\n", stats["memories_formed"])
	fmt.Printf("  Goals Created: %v\n", stats["goals_created"])
	fmt.Println()

	if llmProviders, ok := stats["llm_providers"]; ok {
		fmt.Printf("LLM Providers: %v\n", llmProviders)
		fmt.Printf("Current Provider: %v\n", stats["llm_current"])
		fmt.Println()
	}

	if dbThoughts, ok := stats["db_thought_count"]; ok {
		fmt.Println("Database:")
		fmt.Printf("  Thoughts: %v\n", dbThoughts)
		fmt.Printf("  Memories: %v\n", stats["db_memory_count"])
		fmt.Printf("  Active Goals: %v\n", stats["db_active_goal_count"])
		if dbSize, ok := stats["db_db_size_bytes"]; ok {
			fmt.Printf("  Size: %v bytes\n", dbSize)
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("-", 80))
}
