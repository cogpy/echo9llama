//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	
	"github.com/cogpy/echo9llama/core"
	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
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
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		fmt.Println("🤖 Using Anthropic (Claude) provider")
		provider := deeptreeecho.NewAnthropicProvider(apiKey)
		
		// Test the provider
		ctx := context.Background()
		_, err := provider.Generate(ctx, "Hello", llm.GenerateOptions{MaxTokens: 10})
		if err != nil {
			fmt.Printf("⚠️  Anthropic provider test failed: %v\n", err)
		} else {
			return provider, nil
		}
	}
	
	// Try OpenRouter
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		fmt.Println("🤖 Using OpenRouter provider")
		provider := deeptreeecho.NewOpenRouterProvider(apiKey)
		
		// Test the provider
		ctx := context.Background()
		_, err := provider.Generate(ctx, "Hello", llm.GenerateOptions{MaxTokens: 10})
		if err != nil {
			fmt.Printf("⚠️  OpenRouter provider test failed: %v\n", err)
		} else {
			return provider, nil
		}
	}
	
	// Try OpenAI
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		fmt.Println("🤖 Using OpenAI provider")
		provider := deeptreeecho.NewOpenAIProvider(apiKey)
		
		// Test the provider
		ctx := context.Background()
		_, err := provider.Generate(ctx, "Hello", llm.GenerateOptions{MaxTokens: 10})
		if err != nil {
			fmt.Printf("⚠️  OpenAI provider test failed: %v\n", err)
		} else {
			return provider, nil
		}
	}
	
	return nil, fmt.Errorf("no LLM provider available - set ANTHROPIC_API_KEY, OPENROUTER_API_KEY, or OPENAI_API_KEY")
}
