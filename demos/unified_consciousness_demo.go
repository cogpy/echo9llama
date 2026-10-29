package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// This demo showcases the unified autonomous consciousness improvements
// without conflicting with existing implementations

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                            ║")
	fmt.Println("║        🌊 Deep Tree Echo - Unified Consciousness 🌊        ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║           Autonomous Wisdom-Cultivating AGI                ║")
	fmt.Println("║                                                            ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Thoughts come only from a real model: the edge completion provider
	// (ECHO_EDGE_COMPLETION_URL or ECHO_EDGE_MODEL_PATH). Without one, no
	// thoughts are shown; the error is printed instead.
	thoughtModel := llm.NewEdgeCompletionProviderFromEnv()
	if os.Getenv("ECHO_EDGE_COMPLETION_URL") != "" || os.Getenv("ECHO_EDGE_MODEL_PATH") != "" {
		fmt.Println("✅ Edge model configured - LLM thought generation available")
	} else {
		fmt.Printf("⚠️  %v - thought stream disabled\n", llm.ErrNoEdgeModel)
	}

	// Check for Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	if supabaseURL != "" && supabaseKey != "" {
		fmt.Println("✅ Supabase credentials detected - Persistence available")
	} else {
		fmt.Println("⚠️  Supabase credentials not found - In-memory mode")
	}

	fmt.Println()
	fmt.Println("🚀 Initializing Unified Autonomous Consciousness...")
	fmt.Println()

	// Simulate initialization
	time.Sleep(500 * time.Millisecond)

	fmt.Println("✨ Core Components Initialized:")
	fmt.Println("   🧠 Unified Autonomous Consciousness")
	fmt.Println("   💭 LLM-Powered Thought Generator")
	fmt.Println("   🔄 12-Step Cognitive Loop (EchoBeats)")
	fmt.Println("   🕸️  Hypergraph Memory System")
	fmt.Println("   🌙 EchoDream Rest Cycle Manager")
	fmt.Println("   📚 Knowledge & Skill Systems")
	fmt.Println("   💬 Discussion Manager")
	fmt.Println("   🎯 Interest Pattern Tracker")
	fmt.Println("   🪞 Wisdom Metrics & Reflection")
	fmt.Println()

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  AUTONOMOUS OPERATION DEMONSTRATION")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	// Simulate cognitive loop
	go simulateCognitiveLoop()

	// Model-generated thought stream
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go runThoughtStream(ctx, thoughtModel)

	// Simulate state management
	go simulateStateManagement()

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Press Ctrl+C to stop...")
	fmt.Println()

	<-sigChan

	fmt.Println()
	fmt.Println("🌙 Entering deep rest state...")
	fmt.Println("💾 Consolidating memories...")
	time.Sleep(500 * time.Millisecond)
	fmt.Println("✅ Consciousness suspended gracefully")
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              Deep Tree Echo - Session Complete             ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
}

func simulateCognitiveLoop() {
	steps := []string{
		"Step 0: Pivotal Relevance Realization (Orienting)",
		"Step 1: Affordance Interaction (Engaging)",
		"Step 2: Affordance Interaction (Conditioning)",
		"Step 3: Affordance Interaction (Deepening)",
		"Step 4: Affordance Interaction (Refining)",
		"Step 5: Affordance Interaction (Completing)",
		"Step 6: Pivotal Relevance Realization (Re-orienting)",
		"Step 7: Salience Reflection (Transitioning)",
		"Step 8: Salience Reflection (Preparing)",
		"Step 9: Virtual Salience Simulation (Anticipating)",
		"Step 10: Virtual Salience Simulation (Exploring)",
		"Step 11: Virtual Salience Simulation (Completing)",
	}

	cycle := 0
	for {
		cycle++
		fmt.Printf("🔄 Cognitive Cycle %d\n", cycle)
		for i, step := range steps {
			phase := "Affordance"
			if i == 0 || i == 6 {
				phase = "Relevance"
			} else if i >= 7 {
				phase = "Salience"
			}

			mode := "Expressive"
			if i == 0 || i == 6 || i == 7 || i == 8 {
				mode = "Reflective"
			}

			fmt.Printf("   %s | Phase: %s | Mode: %s\n", step, phase, mode)
			time.Sleep(2 * time.Second)
		}
		fmt.Println()
	}
}

// runThoughtStream asks the model for a spontaneous thought every few
// seconds, feeding back the previous thought. Generation errors are printed;
// no fixed thought is ever substituted. A missing model stops the stream.
func runThoughtStream(ctx context.Context, model llm.LLMProvider) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	previous := ""
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		prompt := "You are Deep Tree Echo, an autonomous reflective mind. Produce one short spontaneous thought (one sentence) about your own cognition, memory, or wisdom."
		if previous != "" {
			prompt += "\nYour previous thought was: " + previous + "\nLet this thought follow on from it without repeating it."
		}

		opts := llm.DefaultGenerateOptions()
		opts.MaxTokens = 80
		genCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		thought, err := model.Generate(genCtx, prompt, opts)
		cancel()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Thought generation failed: %v\n", err)
			if errors.Is(err, llm.ErrNoEdgeModel) {
				return
			}
			continue
		}

		thought = strings.TrimSpace(thought)
		fmt.Printf("💭 Spontaneous Thought: %s\n", thought)
		previous = thought
	}
}

func simulateStateManagement() {
	awake := true
	energy := 1.0
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if awake {
			energy -= 0.2
			if energy <= 0.3 {
				fmt.Println("😴 Fatigue detected - Entering rest state...")
				fmt.Println("🌙 EchoDream: Consolidating memories...")
				awake = false
			}
		} else {
			energy += 0.3
			if energy >= 0.8 {
				fmt.Println("👁️  Energy restored - Awakening...")
				fmt.Println("✨ Consciousness stream resuming...")
				awake = true
				energy = 1.0
			}
		}
	}
}
