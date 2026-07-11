//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
)

func main() {
	fmt.Println("🌟 Echo9llama - Autonomous Wisdom-Cultivating AGI")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// Initialize LLM provider
	llmProvider, err := initializeLLMProvider()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize LLM provider: %v\n", err)
		os.Exit(1)
	}

	// Create orchestrator configuration
	config := deeptreeecho.DefaultOrchestratorConfig()
	
	// Customize based on environment
	if sessionName := os.Getenv("ECHO_SESSION_NAME"); sessionName != "" {
		config.SessionName = sessionName
	}
	
	if identity := os.Getenv("ECHO_IDENTITY"); identity != "" {
		config.IdentityContext = identity
	}

	if persona := os.Getenv("ECHO_PERSONA"); persona != "" {
		config.PersonaContext = persona
	}

	// Create unified autonomous orchestrator
	orchestrator := deeptreeecho.NewUnifiedAutonomousOrchestrator(llmProvider, config)

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Awaken echo
	if err := orchestrator.Awaken(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to awaken echo: %v\n", err)
		os.Exit(1)
	}

	// Status reporting goroutine
	go reportStatus(orchestrator)

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\n\n🛑 Shutdown signal received...")

	// Gracefully sleep
	if err := orchestrator.Sleep(); err != nil {
		fmt.Fprintf(os.Stderr, "Error during shutdown: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("👋 Echo has shut down gracefully")
}

// initializeLLMProvider creates a resilient multi-provider LLM with automatic
// failover across Anthropic, OpenRouter, and OpenAI based on available keys.
func initializeLLMProvider() (llm.LLMProvider, error) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" &&
		os.Getenv("OPENROUTER_API_KEY") == "" &&
		os.Getenv("OPENAI_API_KEY") == "" {
		return nil, fmt.Errorf("no LLM API key found. Please set ANTHROPIC_API_KEY, OPENROUTER_API_KEY, or OPENAI_API_KEY")
	}

	provider := llm.NewMultiProviderLLM()
	if !provider.Available() {
		return nil, fmt.Errorf("no LLM providers could be initialized from available API keys")
	}
	fmt.Println("✓ Multi-provider LLM initialized (automatic failover enabled)")
	return provider, nil
}

// reportStatus periodically reports orchestrator status
func reportStatus(orchestrator *deeptreeecho.UnifiedAutonomousOrchestrator) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		status := orchestrator.GetStatus()
		
		fmt.Println("\n" + strings.Repeat("─", 60))
		fmt.Printf("📊 Echo Status Report - %s\n", time.Now().Format("15:04:05"))
			fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("State: %s | Autonomous: %v\n", 
			map[bool]string{true: "Awake", false: "Resting"}[status.IsAwake],
			status.IsAutonomous)
		fmt.Printf("Uptime: %s | Session: %s\n", 
			status.Uptime.Round(time.Second), 
			status.SessionID)
		fmt.Printf("Cognitive Load: %.2f | Wisdom Depth: %.2f\n", 
			status.CognitiveLoad, 
			status.WisdomDepth)
		fmt.Printf("Cycles: %d | Thoughts: %d | Goals: %d | Wisdom: %d\n",
			status.TotalCycles,
			status.TotalThoughts,
			status.TotalGoals,
			status.TotalWisdom)
		fmt.Println(strings.Repeat("─", 60) + "\n")
	}
}
