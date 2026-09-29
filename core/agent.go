package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// AutonomousAgent represents the main autonomous agent
type AutonomousAgent struct {
	llmProvider llm.LLMProvider
	ctx         context.Context //nolint:containedctx // lifecycle context created with cancel in the constructor; stops background goroutines
	cancel      context.CancelFunc
}

// NewAutonomousAgent creates a new autonomous agent
func NewAutonomousAgent(provider llm.LLMProvider) *AutonomousAgent {
	ctx, cancel := context.WithCancel(context.Background())
	return &AutonomousAgent{
		llmProvider: provider,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Run starts the autonomous agent
func (a *AutonomousAgent) Run() error {
	fmt.Println("🚀 Starting Deep Tree Echo autonomous agent...")

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Main loop
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return nil
		case sig := <-sigChan:
			fmt.Printf("\n⚡ Received signal: %v\n", sig)
			a.cancel()
			return nil
		case <-ticker.C:
			// Heartbeat
			fmt.Println("💓 Agent heartbeat...")
		}
	}
}

// Stop stops the autonomous agent
func (a *AutonomousAgent) Stop() {
	a.cancel()
}
