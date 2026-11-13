package ontogenesis

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cogpy/echo9llama/core/llm"
)

type fakeCompleter struct {
	reply   func(prompt string) string
	err     error
	prompts []string
}

func (f *fakeCompleter) Generate(_ context.Context, prompt string, _ llm.GenerateOptions) (string, error) {
	f.prompts = append(f.prompts, prompt)
	if f.err != nil {
		return "", f.err
	}
	return f.reply(prompt), nil
}

func TestSynthesisWisdomIsTheModels(t *testing.T) {
	o := NewOperations(nil, nil)
	fake := &fakeCompleter{reply: func(p string) string {
		if strings.Contains(p, "Rate it honestly") {
			return "70 40 90"
		}
		return "Patience lets patterns show themselves."
	}}
	o.SetGenerator(fake)

	op, err := o.ExecuteSynthesisOperation(context.Background(), []string{"rivers carve slowly"}, []string{"I waited and saw"}, SynthesisContemplative)
	if err != nil {
		t.Fatal(err)
	}
	w := op.WisdomOutput
	if w.Content != "Patience lets patterns show themselves." || w.Depth != 0.7 || w.Applicability != 0.4 || w.Confidence != 0.9 {
		t.Fatalf("wisdom not the model's: %+v", w)
	}
	if !strings.Contains(fake.prompts[0], "rivers carve slowly") || !strings.Contains(fake.prompts[0], "contemplation") {
		t.Fatalf("prompt lacks the inputs: %q", fake.prompts[0])
	}
}

func TestSynthesisWithoutModelFails(t *testing.T) {
	for name, gen := range map[string]llm.Completer{
		"none":    nil,
		"error":   &fakeCompleter{err: errors.New("model down")},
		"unrated": &fakeCompleter{reply: func(string) string { return "some wisdom" }},
	} {
		t.Run(name, func(t *testing.T) {
			o := NewOperations(nil, nil)
			if gen != nil {
				o.SetGenerator(gen)
			}
			if _, err := o.ExecuteSynthesisOperation(context.Background(), []string{"k"}, nil, SynthesisReflective); err == nil {
				t.Fatal("synthesis succeeded without model wisdom")
			}
			if len(o.GetWisdomArtifacts()) != 0 {
				t.Fatal("wisdom artifact recorded without the model")
			}
		})
	}
}
