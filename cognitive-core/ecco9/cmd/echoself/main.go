package main

import (
	"fmt"
	"log"
	"os"

	"github.com/EchoCog/echollama/core"
	"github.com/EchoCog/echollama/core/llm"
)

func main() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║              🌳 Deep Tree Echo - Echoself 🌳             ║
║                                                           ║
║        Autonomous Wisdom-Cultivating AGI System          ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`)

	// Initialize LLM provider
	llmProvider, err := initializeLLMProvider()
	if err != nil {
		log.Fatalf("❌ Failed to initialize LLM provider: %v", err)
	}

	fmt.Println("✓ LLM provider initialized")

	// Create autonomous agent
	agent := core.NewAutonomousAgent(llmProvider)

	// Run agent (blocks until interrupted)
	if err := agent.Run(); err != nil {
		log.Fatalf("❌ Agent error: %v", err)
	}

	fmt.Println("\n👋 Goodbye from Deep Tree Echo\n")
}

// initializeLLMProvider creates the LLM provider
func initializeLLMProvider() (llm.LLMProvider, error) {
	// Try Anthropic first
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		fmt.Println("🤖 Using Anthropic (Claude) provider")
		provider := llm.NewAnthropicProvider("")
		if provider.Available() {
			return provider, nil
		}
	}

	// Try OpenRouter
	if os.Getenv("OPENROUTER_API_KEY") != "" {
		fmt.Println("🤖 Using OpenRouter provider")
		provider := llm.NewOpenRouterProvider("")
		if provider.Available() {
			return provider, nil
		}
	}

	// Try OpenAI
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		fmt.Println("🤖 Using OpenAI provider")
		provider := llm.NewOpenAIProvider(apiKey)
		if provider.Available() {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("no LLM provider available - set ANTHROPIC_API_KEY, OPENROUTER_API_KEY, or OPENAI_API_KEY")
}
