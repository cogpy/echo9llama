package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cogpy/echo9llama/core/autonomous"
	"github.com/cogpy/echo9llama/core/llm"
)

func main() {
	fmt.Println("=" * 60)
	fmt.Println("🌳 Deep Tree Echo - Autonomous Agent Test")
	fmt.Println("=" * 60)
	fmt.Println()

	// Check for API keys
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	openrouterKey := os.Getenv("OPENROUTER_API_KEY")
	
	if anthropicKey == "" && openrouterKey == "" {
		log.Fatal("❌ No API keys found. Please set ANTHROPIC_API_KEY or OPENROUTER_API_KEY")
	}

	// Create provider manager
	providerManager := llm.NewProviderManager()

	// Register available providers
	if anthropicKey != "" {
		anthropicProvider, err := llm.NewAPIProvider("anthropic", "")
		if err != nil {
			log.Printf("⚠️  Failed to create Anthropic provider: %v", err)
		} else {
			if err := providerManager.RegisterProvider(anthropicProvider); err != nil {
				log.Printf("⚠️  Failed to register Anthropic provider: %v", err)
			} else {
				fmt.Println("✅ Anthropic provider registered")
			}
		}
	}

	if openrouterKey != "" {
		openrouterProvider, err := llm.NewAPIProvider("openrouter", "")
		if err != nil {
			log.Printf("⚠️  Failed to create OpenRouter provider: %v", err)
		} else {
			if err := providerManager.RegisterProvider(openrouterProvider); err != nil {
				log.Printf("⚠️  Failed to register OpenRouter provider: %v", err)
			} else {
				fmt.Println("✅ OpenRouter provider registered")
			}
		}
	}

	// Set fallback chain
	if anthropicKey != "" && openrouterKey != "" {
		providerManager.SetFallbackChain([]string{"anthropic", "openrouter"})
	} else if anthropicKey != "" {
		providerManager.SetFallbackChain([]string{"anthropic"})
	} else {
		providerManager.SetFallbackChain([]string{"openrouter"})
	}

	if !providerManager.Available() {
		log.Fatal("❌ No LLM providers available")
	}

	fmt.Println()
	fmt.Println("🚀 Creating autonomous agent...")

	// Create autonomous agent
	agent := autonomous.NewAutonomousAgent(providerManager)

	fmt.Println()
	fmt.Println("▶️  Starting autonomous agent...")
	fmt.Println("   (The agent will run autonomously for 5 minutes)")
	fmt.Println("   (Press Ctrl+C to stop early)")
	fmt.Println()

	// Start the agent
	if err := agent.Start(); err != nil {
		log.Fatalf("❌ Failed to start agent: %v", err)
	}

	// Run for 5 minutes as a test
	time.Sleep(5 * time.Minute)

	fmt.Println()
	fmt.Println("⏱️  Test duration complete. Stopping agent...")

	// Stop the agent
	if err := agent.Stop(); err != nil {
		log.Printf("⚠️  Error stopping agent: %v", err)
	}

	fmt.Println()
	fmt.Println("✅ Test complete!")
}
