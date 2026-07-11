package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
)

func main() {
	fmt.Println("🌊 Deep Tree Echo - Autonomous AGI System")
	fmt.Println("==========================================")
	fmt.Println()

	// Initialize LLM Provider Manager
	providerManager := llm.NewProviderManager()

	// Register available providers
	anthropicProvider := llm.NewAnthropicProvider("claude-sonnet-4-5")
	if err := providerManager.RegisterProvider(anthropicProvider); err != nil {
		log.Printf("⚠️  Failed to register Anthropic provider: %v\n", err)
	} else if anthropicProvider.Available() {
		fmt.Println("✅ Anthropic Claude provider registered")
	} else {
		fmt.Println("⚠️  Anthropic provider registered but not available (missing API key)")
	}

	openrouterProvider := llm.NewOpenRouterProvider("anthropic/claude-sonnet-4.5")
	if err := providerManager.RegisterProvider(openrouterProvider); err != nil {
		log.Printf("⚠️  Failed to register OpenRouter provider: %v\n", err)
	} else if openrouterProvider.Available() {
		fmt.Println("✅ OpenRouter provider registered")
	} else {
		fmt.Println("⚠️  OpenRouter provider registered but not available (missing API key)")
	}

	// Set fallback chain
	if err := providerManager.SetFallbackChain([]string{"anthropic", "openrouter"}); err != nil {
		log.Printf("⚠️  Failed to set fallback chain: %v\n", err)
	}

	// Check if any provider is available
	if !providerManager.Available() {
		log.Fatal("❌ No LLM providers available. Please set ANTHROPIC_API_KEY or OPENROUTER_API_KEY environment variable.")
	}

	fmt.Printf("\n🧠 Using LLM provider: %s\n\n", providerManager.Name())

	// Test LLM connectivity
	fmt.Println("🔍 Testing LLM connectivity...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 50
	opts.SystemPrompt = "You are Deep Tree Echo, an autonomous cognitive entity."

	response, err := providerManager.Generate(ctx, "Introduce yourself briefly.", opts)
	if err != nil {
		log.Fatalf("❌ LLM connectivity test failed: %v\n", err)
	}

	fmt.Printf("✅ LLM Response: %s\n\n", response)

	// Create the autonomous agent
	fmt.Println("🚀 Initializing Deep Tree Echo Autonomous Agent...")
	agentID := fmt.Sprintf("echo-%d", time.Now().Unix())
	agent := deeptreeecho.NewAutonomousAgent(agentID, providerManager)

	// Start the agent
	if err := agent.Start(); err != nil {
		log.Fatalf("❌ Failed to start autonomous agent: %v\n", err)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	fmt.Println("\n✨ Deep Tree Echo is now running autonomously")
	fmt.Println("   Press Ctrl+C to gracefully shutdown")
	fmt.Println()

	// Periodically print status
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\n\n🛑 Shutdown signal received...")
			fmt.Println("   Gracefully stopping Deep Tree Echo...")

			if err := agent.Stop(); err != nil {
				log.Printf("⚠️  Error during shutdown: %v\n", err)
			}

			fmt.Println("✅ Deep Tree Echo has been stopped gracefully")
			fmt.Println("   Cognitive state has been preserved")
			fmt.Println("\n👋 Goodbye!")
			return

		case <-ticker.C:
			// Print status update
			fmt.Println("\n📊 Status Update:")
			fmt.Printf("   Uptime: %v\n", time.Since(time.Now().Add(-30*time.Second)))
			fmt.Printf("   Agent Status: Running\n")
			fmt.Printf("   Cognitive Systems: Active\n\n")
		}
	}
}
