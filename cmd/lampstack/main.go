// cmd/lampstack/main.go
// LAMPSTACK Echo Server - Unified cognitive integration stack
//
// Usage:
//   go run cmd/lampstack/main.go
//
// Environment Variables:
//   ANTHROPIC_API_KEY    - Anthropic API key for Claude
//   OPENROUTER_API_KEY   - OpenRouter API key
//   OPENAI_API_KEY       - OpenAI API key
//   LAMPSTACK_PORT       - Server port (default: 8080)
//   LAMPSTACK_DATA_DIR   - Data directory (default: data/lampstack)
//
// Endpoints:
//   GET  /health              - Health check
//   GET  /info                - Server information
//   GET  /metrics             - Stack metrics
//   POST /api/generate        - Text generation
//   POST /api/chat            - Chat completion
//   POST /api/echo/think      - Cognitive thought generation
//   POST /api/echo/reflect    - Meta-cognitive reflection
//   GET  /api/echo/state      - Current cognitive state
//   POST /api/memory/store    - Store memory
//   POST /api/memory/retrieve - Retrieve memories
//   GET  /api/memory/stats    - Memory statistics
//   GET/POST /api/snapshot    - State snapshots

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cogpy/echo9llama/core/lampstack"
)

func main() {
	// Parse flags
	port := flag.String("port", getEnvOrDefault("LAMPSTACK_PORT", "8080"), "Server port")
	dataDir := flag.String("data", getEnvOrDefault("LAMPSTACK_DATA_DIR", "data/lampstack"), "Data directory")
	flag.Parse()

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("              LAMPSTACK Echo - Cognitive Integration Stack")
	fmt.Printf("                         Version %s\n", lampstack.Version)
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create configuration
	config := lampstack.DefaultConfig()
	config.DataDirectory = *dataDir

	// Create stack
	stack, err := lampstack.NewStack(config)
	if err != nil {
		fmt.Printf("❌ Failed to create stack: %v\n", err)
		os.Exit(1)
	}

	// Initialize stack
	fmt.Println("Initializing LAMPSTACK...")
	if err := stack.Initialize(ctx); err != nil {
		fmt.Printf("❌ Failed to initialize stack: %v\n", err)
		os.Exit(1)
	}

	// Start stack
	if err := stack.Start(ctx); err != nil {
		fmt.Printf("❌ Failed to start stack: %v\n", err)
		os.Exit(1)
	}

	// Create and start server
	serverConfig := lampstack.DefaultServerConfig()
	serverConfig.Address = ":" + *port

	server := lampstack.NewServer(stack, serverConfig)
	if err := server.Start(ctx); err != nil {
		fmt.Printf("❌ Failed to start server: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("  Server running at http://localhost:%s\n", *port)
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /health              - Health check")
	fmt.Println("  GET  /info                - Server information")
	fmt.Println("  POST /api/generate        - Text generation")
	fmt.Println("  POST /api/chat            - Chat completion")
	fmt.Println("  POST /api/echo/think      - Cognitive thought")
	fmt.Println("  POST /api/echo/reflect    - Meta-cognitive reflection")
	fmt.Println("  GET  /api/echo/state      - Cognitive state")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	fmt.Println("Shutting down...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop server
	if err := server.Stop(shutdownCtx); err != nil {
		fmt.Printf("⚠️  Server shutdown error: %v\n", err)
	}

	// Stop stack
	if err := stack.Stop(shutdownCtx); err != nil {
		fmt.Printf("⚠️  Stack shutdown error: %v\n", err)
	}

	fmt.Println("Goodbye!")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
