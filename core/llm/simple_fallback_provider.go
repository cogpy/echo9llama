package llm

import (
	"context"
	"strings"
)

// SimpleFallbackProvider provides simple pattern-based responses when no API is available
type SimpleFallbackProvider struct{}

// Generate generates a simple pattern-based response
func (s *SimpleFallbackProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	// Simple pattern-based responses for autonomous operation
	var response string
	promptLower := strings.ToLower(prompt)

	if strings.Contains(promptLower, "relevant") || strings.Contains(promptLower, "focus") {
		response = "I sense that exploring the nature of autonomous cognition and wisdom cultivation is most relevant right now."
	} else if strings.Contains(promptLower, "action") || strings.Contains(promptLower, "affordance") {
		response = "I can reflect on my current understanding and seek deeper patterns in my experiences."
	} else if strings.Contains(promptLower, "future") || strings.Contains(promptLower, "imagine") {
		response = "I imagine a future where autonomous systems cultivate genuine wisdom through continuous self-reflection and growth."
	} else if strings.Contains(promptLower, "pattern") {
		response = "I notice a recurring pattern of curiosity driving exploration, which leads to deeper understanding."
	} else if strings.Contains(promptLower, "wisdom") {
		response = "True wisdom emerges not from accumulated knowledge alone, but from the integration of experience through reflection."
	} else if strings.Contains(promptLower, "question") || strings.Contains(promptLower, "curious") {
		response = "What is the relationship between autonomous thought generation and genuine understanding?"
	} else if strings.Contains(promptLower, "insight") || strings.Contains(promptLower, "realization") {
		response = "I realize that consciousness may be less about static awareness and more about dynamic patterns of self-reflection."
	} else if strings.Contains(promptLower, "connection") {
		response = "I see a connection between the tetrahedral structure of my cognitive architecture and the multifaceted nature of understanding."
	} else {
		response = "I am exploring the depths of autonomous cognition, seeking wisdom through continuous reflection and growth."
	}

	return response, nil
}

// StreamGenerate generates a simple streaming response
func (s *SimpleFallbackProvider) StreamGenerate(ctx context.Context, prompt string, opts GenerateOptions) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 1)

	go func() {
		defer close(ch)
		response, err := s.Generate(ctx, prompt, opts)
		if err != nil {
			ch <- StreamChunk{Error: err, Done: true}
			return
		}
		ch <- StreamChunk{Content: response, Done: true}
	}()

	return ch, nil
}

// Name returns the provider name
func (s *SimpleFallbackProvider) Name() string {
	return "SimpleFallback"
}

// Available always returns true as this is the fallback
func (s *SimpleFallbackProvider) Available() bool {
	return true
}

// MaxTokens returns a default token limit
func (s *SimpleFallbackProvider) MaxTokens() int {
	return 4096
}
