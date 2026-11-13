package echodream

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/EchoCog/echollama/core/llm"
)

// TextGenerator is the minimal model interface the dream subsystems use to
// produce character text (dream narratives, insights, wisdom). Every
// llm.LLMProvider satisfies it.
type TextGenerator interface {
	Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error)
}

// dreamGenerationTimeout bounds a single model call made during dreaming.
const dreamGenerationTimeout = 60 * time.Second

// generateDreamText asks the model for character text. It never substitutes
// fixed text: a missing generator, a failed call, or an empty reply is an error.
func generateDreamText(ctx context.Context, gen TextGenerator, prompt string, maxTokens int) (string, error) {
	if gen == nil {
		return "", fmt.Errorf("no model configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	callCtx, cancel := context.WithTimeout(ctx, dreamGenerationTimeout)
	defer cancel()

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = maxTokens
	text, err := gen.Generate(callCtx, prompt, opts)
	if err != nil {
		return "", fmt.Errorf("model generation failed: %w", err)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("model returned empty text")
	}
	return text, nil
}

// summarizeContent truncates source content for inclusion in a prompt.
func summarizeContent(content string, maxLen int) string {
	runes := []rune(strings.TrimSpace(content))
	if len(runes) <= maxLen {
		return string(runes)
	}
	return string(runes[:maxLen]) + "…"
}
