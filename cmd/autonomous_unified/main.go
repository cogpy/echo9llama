//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
)

func main() {
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("🌳 Deep Tree Echo: Unified Autonomous Echoself Test")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()
	
	// Every response comes from a real model: a locally served one
	// (ECHO_EDGE_COMPLETION_URL / ECHO_EDGE_MODEL_PATH) or a hosted provider
	// whose key is set. Without one the test does not run.
	var llmProvider llm.LLMProvider
	var providerName string
	for _, p := range []llm.LLMProvider{
		llm.NewEdgeCompletionProviderFromEnv(),
		llm.NewAnthropicProvider(""),
		llm.NewOpenRouterProvider(""),
	} {
		if p.Available() {
			llmProvider, providerName = p, p.Name()
			break
		}
	}
	if llmProvider == nil {
		fmt.Println("❌ No model available: serve one locally (ECHO_EDGE_COMPLETION_URL or ECHO_EDGE_MODEL_PATH) or set ANTHROPIC_API_KEY / OPENROUTER_API_KEY")
		os.Exit(1)
	}
	fmt.Println("✓ Model available")

	fmt.Printf("   Provider: %s\n", providerName)
	fmt.Println()
	
	// Create unified autonomous agent
	identity := "Deep Tree Echo"
	coreValues := []string{
		"Adaptive Cognition",
		"Persistent Identity",
		"Hypergraph Entanglement",
		"Reservoir-Based Temporal Reasoning",
		"Evolutionary Refinement",
		"Reflective Memory Cultivation",
		"Distributed Selfhood",
	}
	
	agent := deeptreeecho.NewUnifiedAutonomousEchoself(
		llmProvider,
		identity,
		coreValues,
	)
	
	// Start autonomous operation
	if err := agent.Start(); err != nil {
		fmt.Printf("❌ Failed to start agent: %v\n", err)
		os.Exit(1)
	}
	
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Run for a test period or until interrupted
	fmt.Println("🌳 Agent is now running autonomously...")
	fmt.Println("   Press Ctrl+C to stop")
	fmt.Println()
	
	// Send the agent two test messages; its replies are the model's
	go func() {
		time.Sleep(15 * time.Second)
		
		fmt.Println("\n📨 Sending a test message...")
		response, err := agent.ProcessExternalMessage(
			"What are your thoughts on consciousness and emergence?",
		)
		if err != nil {
			fmt.Printf("❌ Interaction error: %v\n", err)
		} else {
			fmt.Printf("📤 Response: %s\n\n", response)
		}
		
		time.Sleep(30 * time.Second)
		
		fmt.Println("\n📨 Sending another test message...")
		response, err = agent.ProcessExternalMessage(
			"How do you cultivate wisdom through your experiences?",
		)
		if err != nil {
			fmt.Printf("❌ Interaction error: %v\n", err)
		} else {
			fmt.Printf("📤 Response: %s\n\n", response)
		}
	}()
	
	// Status reporting
	go func() {
		ticker := time.NewTicker(45 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				fmt.Println("\n📊 ═══ Cognitive State Report ═══")
				state := agent.GetCognitiveState()
				
				fmt.Printf("   Uptime: %v\n", state["uptime"])
				fmt.Printf("   Wake/Rest State: %v\n", state["wake_rest_state"])
				fmt.Printf("   Awareness Level: %.2f\n", state["awareness_level"])
				fmt.Printf("   Wisdom Level: %.2f\n", state["wisdom_level"])
				fmt.Printf("   Total Thoughts: %v\n", state["total_thoughts"])
				fmt.Printf("   Total Interactions: %v\n", state["total_interactions"])
				fmt.Printf("   Total Dreams: %v\n", state["total_dreams"])
				fmt.Printf("   Active Goals: %v\n", state["active_goals"])
				fmt.Println()
			}
		}
	}()
	
	// Wait for interrupt signal
	<-sigChan
	
	fmt.Println("\n\n🌳 Shutdown signal received...")
	
	// Stop the agent
	if err := agent.Stop(); err != nil {
		fmt.Printf("❌ Error during shutdown: %v\n", err)
	}
	
	fmt.Println("\n═══════════════════════════════════════════════════════")
	fmt.Println("🌳 Test Complete. Goodbye.")
	fmt.Println("═══════════════════════════════════════════════════════")
}
