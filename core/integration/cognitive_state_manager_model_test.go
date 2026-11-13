package integration

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cogpy/echo9llama/core/llm"
)

// scriptedProvider returns a fixed reply or error and records prompts.
type scriptedProvider struct {
	reply   string
	err     error
	prompts []string
}

func (p *scriptedProvider) Generate(_ context.Context, prompt string, _ llm.GenerateOptions) (string, error) {
	p.prompts = append(p.prompts, prompt)
	if p.err != nil {
		return "", p.err
	}
	return p.reply, nil
}

func (p *scriptedProvider) StreamGenerate(context.Context, string, llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	return nil, errors.New("not implemented")
}
func (p *scriptedProvider) Name() string    { return "scripted" }
func (p *scriptedProvider) Available() bool { return true }
func (p *scriptedProvider) MaxTokens() int  { return 4096 }

func seedPatterns(csm *CognitiveStateManager) {
	thoughts := []SharedThought{
		{Content: "a", Tags: []string{"alpha"}},
		{Content: "b", Tags: []string{"beta"}},
		{Content: "c", Tags: []string{"gamma"}},
	}
	for _, theme := range []string{"alpha", "beta", "gamma"} {
		csm.recordPattern(theme, 3, thoughts)
	}
}

func TestGenerateWisdomUsesModelText(t *testing.T) {
	csm := NewCognitiveStateManager()
	provider := &scriptedProvider{reply: "  model-wisdom-7f3a  "}
	csm.SetLLMProvider(provider)
	seedPatterns(csm)

	if err := csm.generateWisdom(context.Background()); err != nil {
		t.Fatalf("generateWisdom: %v", err)
	}
	insights := csm.GetWisdomInsights()
	if len(insights) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(insights))
	}
	if insights[0].Insight != "model-wisdom-7f3a" {
		t.Fatalf("insight must be the model's text, got %q", insights[0].Insight)
	}
	if len(provider.prompts) != 1 || !strings.Contains(provider.prompts[0], "alpha") {
		t.Fatalf("prompt should describe the patterns, got %v", provider.prompts)
	}
}

func TestGenerateWisdomProviderErrorYieldsNoInsight(t *testing.T) {
	csm := NewCognitiveStateManager()
	csm.SetLLMProvider(&scriptedProvider{err: llm.ErrEdgeBusy})
	seedPatterns(csm)

	if err := csm.generateWisdom(context.Background()); !errors.Is(err, llm.ErrEdgeBusy) {
		t.Fatalf("expected ErrEdgeBusy, got %v", err)
	}
	if n := len(csm.GetWisdomInsights()); n != 0 {
		t.Fatalf("expected no wisdom on model error, got %d", n)
	}
}

func TestGenerateWisdomWithoutProviderYieldsNoInsight(t *testing.T) {
	csm := NewCognitiveStateManager()
	seedPatterns(csm)

	if err := csm.generateWisdom(context.Background()); err == nil {
		t.Fatal("expected an error without a provider")
	}
	if n := len(csm.GetWisdomInsights()); n != 0 {
		t.Fatalf("expected no wisdom without a provider, got %d", n)
	}
}
