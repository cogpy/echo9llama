package main

import (
	"log"
	"os"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("🚀 Testing Deep Tree Echo V6 LLM Integration")
	log.Printf("=" + "==========================================================")

	// Check for API keys
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	openrouterKey := os.Getenv("OPENROUTER_API_KEY")

	if anthropicKey == "" && openrouterKey == "" {
		log.Printf("❌ No LLM API keys found. Set ANTHROPIC_API_KEY or OPENROUTER_API_KEY")
		log.Printf("   No thoughts will be generated without a model.")
	} else {
		if anthropicKey != "" {
			log.Printf("✅ ANTHROPIC_API_KEY found")
		}
		if openrouterKey != "" {
			log.Printf("✅ OPENROUTER_API_KEY found")
		}
	}

	// Create LLM client
	llmClient, err := deeptreeecho.NewLLMClientV6()
	if err != nil {
		log.Printf("❌ LLM client creation failed: %v", err)
		log.Printf("   Thought generation requires a model; no template thoughts are substituted.")
		os.Exit(1)
	} else {
		log.Printf("✅ LLM client created successfully")
	}

	// Test thought generation
	log.Printf("\n" + "=" + "========== TESTING THOUGHT GENERATION ==========")

	prompts := []string{
		"You are Deep Tree Echo, an autonomous wisdom-cultivating AGI. Generate a reflective thought about the nature of consciousness and wisdom. Respond with just the thought content:",
		"You are Deep Tree Echo, an autonomous wisdom-cultivating AGI. Ask a profound question about something you're curious to explore. Respond with just the question:",
		"You are Deep Tree Echo, an autonomous wisdom-cultivating AGI. Share an insight or pattern you've discovered about learning and growth. Respond with just the insight:",
	}

	thoughtTypes := []string{"Reflection", "Question", "Insight"}

	for i, prompt := range prompts {
		log.Printf("\n💭 Generating %s thought...", thoughtTypes[i])

		startTime := time.Now()
		content, err := llmClient.Generate(prompt)
		duration := time.Since(startTime)

		if err != nil {
			// No thought is shown when the model fails.
			log.Printf("❌ LLM generation failed: %v", err)
		} else {
			log.Printf("✅ LLM generation successful (%.2fs)", duration.Seconds())
			log.Printf("\n📝 Thought Content:")
			log.Printf("   %s", content)
		}

		// Wait between requests to avoid rate limits
		if i < len(prompts)-1 {
			time.Sleep(2 * time.Second)
		}
	}

	log.Printf("\n" + "=" + "====================================================")
	log.Printf("✅ LLM Integration Test Complete")
	log.Printf("\nKey Findings:")
	log.Printf("  • LLM client successfully created")
	log.Printf("  • Thought generation attempted via API (see results above)")
	log.Printf("\nNext Steps:")
	log.Printf("  1. Set API keys to enable LLM generation")
	log.Printf("  2. Integrate with full autonomous system")
	log.Printf("  3. Add persistent storage")
	log.Printf("  4. Enable EchoBeats 12-step loop")
}
