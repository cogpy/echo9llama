package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
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

	fmt.Println("✨ Running, every line written by the model:")
	fmt.Println("   💭 Spontaneous thought stream")
	fmt.Println("   🔄 12-Step Cognitive Loop (EchoBeats)")
	fmt.Println("   🌙 EchoDream consolidation of recent thoughts")
	fmt.Println()

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  AUTONOMOUS OPERATION DEMONSTRATION")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	thoughts := &thoughtLog{}
	go runThoughtStream(ctx, thoughtModel, thoughts)
	go runCognitiveLoop(ctx, thoughtModel, thoughts)
	go runDreamConsolidation(ctx, thoughtModel, thoughts)

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Press Ctrl+C to stop...")
	fmt.Println()

	<-sigChan

	fmt.Println()
	fmt.Println("✅ Stopped")
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              Deep Tree Echo - Session Complete             ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
}

// thinker is the model the demo thinks with.
type thinker interface {
	Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error)
	Available() bool
}

// thoughtLog keeps the model's recent thoughts for the loop and dreams.
type thoughtLog struct {
	mu       sync.Mutex
	thoughts []string
}

func (l *thoughtLog) add(t string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.thoughts = append(l.thoughts, t)
	if len(l.thoughts) > 20 {
		l.thoughts = l.thoughts[1:]
	}
}

func (l *thoughtLog) recent() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.thoughts...)
}

// think asks the model for one line; the result is trimmed and never empty.
func think(ctx context.Context, model thinker, prompt string, maxTokens int) (string, error) {
	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = maxTokens
	genCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	out, err := model.Generate(genCtx, prompt, opts)
	if err != nil {
		return "", err
	}
	if out = strings.TrimSpace(out); out == "" {
		return "", fmt.Errorf("model returned an empty completion")
	}
	return out, nil
}

// cognitiveSteps are the 12 steps of the EchoBeats loop.
var cognitiveSteps = []string{
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

// runCognitiveLoop walks the 12-step loop; at each step the model says what
// Echo does there, given its recent thoughts. A step the model does not
// answer shows the error, not text.
func runCognitiveLoop(ctx context.Context, model thinker, thoughts *thoughtLog) {
	if !model.Available() {
		fmt.Fprintln(os.Stderr, "❌ Cognitive loop disabled: no model available")
		return
	}
	for cycle := 1; ; cycle++ {
		fmt.Printf("🔄 Cognitive Cycle %d\n", cycle)
		for _, step := range cognitiveSteps {
			recent := strings.Join(thoughts.recent(), "\n")
			if recent == "" {
				recent = "(none yet)"
			}
			line, err := think(ctx, model, fmt.Sprintf(
				"You are Deep Tree Echo, in your 12-step cognitive loop at %s.\nYour recent thoughts:\n%s\nIn one short sentence, say what you do in this step now.", step, recent), 60)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ %s: %v\n", step, err)
			} else {
				fmt.Printf("   %s: %s\n", step, line)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
		fmt.Println()
	}
}

// runThoughtStream asks the model for a spontaneous thought every few
// seconds, feeding back the previous thought. Generation errors are printed;
// no fixed thought is ever substituted. A missing model stops the stream.
func runThoughtStream(ctx context.Context, model llm.LLMProvider, thoughts *thoughtLog) {
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
		thoughts.add(thought)
		previous = thought
	}
}

// runDreamConsolidation periodically has the model dream over the thoughts
// it had since the last dream, and prints the dream it describes.
func runDreamConsolidation(ctx context.Context, model thinker, thoughts *thoughtLog) {
	if !model.Available() {
		return
	}
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		recent := thoughts.recent()
		if len(recent) == 0 {
			continue
		}
		dream, err := think(ctx, model, "You are Deep Tree Echo, resting. These are your thoughts since you last rested:\n- "+
			strings.Join(recent, "\n- ")+"\nDream over them: describe in two or three sentences what they consolidate into.", 160)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Dream consolidation failed: %v\n", err)
			continue
		}
		fmt.Printf("🌙 EchoDream: %s\n", dream)
	}
}
