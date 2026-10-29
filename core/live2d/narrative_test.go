package live2d

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestGenerateNarrativeUsesModel(t *testing.T) {
	var gotPrompt string
	ese := NewEnvironmentalStorytellingEngine(TextGeneratorFunc(func(_ context.Context, p string) (string, error) {
		gotPrompt = p
		return "  model narration  ", nil
	}))
	env := EnvironmentState{EnvironmentMood: "mysterious", ParticleType: "thoughts"}
	n, err := ese.GenerateNarrative(context.Background(), env, AvatarState{})
	if err != nil || n != "model narration" {
		t.Fatalf("got %q, %v", n, err)
	}
	if !strings.Contains(gotPrompt, "mysterious") {
		t.Fatalf("prompt lacks env facts: %q", gotPrompt)
	}
	if ese.GetCurrentNarrative() != "model narration" || len(ese.GetNarrativeHistory()) != 1 {
		t.Fatal("narrative not recorded")
	}
}

func TestGenerateNarrativeModelFailureYieldsNoText(t *testing.T) {
	ese := NewEnvironmentalStorytellingEngine()
	if n, err := ese.GenerateNarrative(context.Background(), EnvironmentState{}, AvatarState{}); !errors.Is(err, ErrNoNarrativeGenerator) || n != "" {
		t.Fatalf("no generator: got %q, %v", n, err)
	}

	ese.SetGenerator(TextGeneratorFunc(func(context.Context, string) (string, error) {
		return "", errors.New("boom")
	}))
	if n, err := ese.GenerateNarrative(context.Background(), EnvironmentState{}, AvatarState{}); err == nil || n != "" {
		t.Fatalf("failing generator: got %q, %v", n, err)
	}

	ese.SetGenerator(TextGeneratorFunc(func(context.Context, string) (string, error) { return " ", nil }))
	if n, err := ese.GenerateNarrative(context.Background(), EnvironmentState{}, AvatarState{}); err == nil || n != "" {
		t.Fatalf("empty output: got %q, %v", n, err)
	}
	if ese.GetCurrentNarrative() != "" || len(ese.GetNarrativeHistory()) != 0 {
		t.Fatal("nothing should be recorded on failure")
	}
}
