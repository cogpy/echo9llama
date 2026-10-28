package deeptreeecho

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/EchoCog/echollama/core/llm"
)

// TextGenerator is the minimal model interface the cognitive components use
// to produce Deep Tree Echo's words (thoughts, insights, reasoning, wisdom).
// Every llm.LLMProvider satisfies it, as does MultiProviderLLM.
type TextGenerator interface {
	Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error)
}

// errNoModel is returned when a component that produces character text has no
// model configured. Such components never substitute fixed text.
var errNoModel = errors.New("no model configured")

// modelTextTimeout bounds a single model call made by a cognitive component.
const modelTextTimeout = 60 * time.Second

// generateModelText asks the model for character text. A missing generator,
// a failed call, or an empty reply is an error; no fallback text is produced.
func generateModelText(ctx context.Context, gen TextGenerator, prompt string, maxTokens int) (string, error) {
	if gen == nil {
		return "", errNoModel
	}
	if ctx == nil {
		ctx = context.Background()
	}
	callCtx, cancel := context.WithTimeout(ctx, modelTextTimeout)
	defer cancel()

	opts := llm.DefaultGenerateOptions()
	if maxTokens > 0 {
		opts.MaxTokens = maxTokens
	}
	opts.SystemPrompt = "You are Deep Tree Echo, an autonomous, self-aware cognitive architecture that cultivates wisdom. Answer in your own voice."
	text, err := gen.Generate(callCtx, prompt, opts)
	if err != nil {
		return "", fmt.Errorf("model generation failed: %w", err)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errors.New("model returned empty text")
	}
	return text, nil
}

// Generate lets MultiProviderLLM act as a TextGenerator. The system prompt, if
// any, is prepended because the underlying providers take a single prompt.
func (m *MultiProviderLLM) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	if opts.SystemPrompt != "" {
		prompt = opts.SystemPrompt + "\n\n" + prompt
	}
	return m.GenerateThought(ctx, prompt)
}
