package pienn

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cogpy/echo9llama/core/llm"
)

type selfImageFake struct {
	err     error
	prompts []string
}

func (f *selfImageFake) Generate(_ context.Context, prompt string, _ llm.GenerateOptions) (string, error) {
	f.prompts = append(f.prompts, prompt)
	if f.err != nil {
		return "", f.err
	}
	return "MODEL-SELF-IMAGE", nil
}

func TestSelfImagesStartEmpty(t *testing.T) {
	cc := NewCognitiveCore()
	cc.Process("why do rivers remember?")
	for lvl, img := range cc.SelfImages {
		if img.Content != "" {
			t.Fatalf("L%d has fixed content %q", lvl, img.Content)
		}
	}
}

func TestIntrospectWithModelWritesSelfImages(t *testing.T) {
	e := NewEngine()
	e.Core.Process("why do rivers remember?")
	fake := &selfImageFake{}
	report, err := e.IntrospectWithModel(context.Background(), fake)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.SelfImages) != 5 {
		t.Fatalf("self-images = %d", len(report.SelfImages))
	}
	for lvl, img := range report.SelfImages {
		if img.Content != "MODEL-SELF-IMAGE" {
			t.Fatalf("L%d content %q not from model", lvl, img.Content)
		}
	}
	if !strings.Contains(fake.prompts[0], "rivers remember") {
		t.Fatalf("prompt lacks actual state: %q", fake.prompts[0])
	}
}

func TestIntrospectWithModelFailureLeavesNoSelfImage(t *testing.T) {
	e := NewEngine()
	if _, err := e.IntrospectWithModel(context.Background(), &selfImageFake{}); err != nil {
		t.Fatal(err)
	}
	for name, gen := range map[string]SelfImageGenerator{"nil": nil, "error": &selfImageFake{err: errors.New("down")}} {
		report, err := e.IntrospectWithModel(context.Background(), gen)
		if err == nil {
			t.Fatalf("%s: expected an error", name)
		}
		for lvl, img := range report.SelfImages {
			if img.Content != "" {
				t.Fatalf("%s: L%d kept content %q after model failure", name, lvl, img.Content)
			}
		}
	}
}
