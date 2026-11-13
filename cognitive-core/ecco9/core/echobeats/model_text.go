package echobeats

import (
	"context"

	"github.com/EchoCog/echollama/core/llm"
)

// TextGenerator is the minimal model interface echobeats uses to write Deep
// Tree Echo's thoughts, reflections and replies. llm.LLMProvider and
// llm.ProviderManager both satisfy it.
type TextGenerator interface {
	Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error)
}
