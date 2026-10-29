package llm

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type ratingFake struct {
	reply  string
	err    error
	prompt string
}

func (f *ratingFake) Generate(_ context.Context, prompt string, _ GenerateOptions) (string, error) {
	f.prompt = prompt
	return f.reply, f.err
}

func TestRatePercentReturnsTheModelsNumbers(t *testing.T) {
	f := &ratingFake{reply: "Novelty 72, depth 15."}
	r, err := RatePercent(context.Background(), f, "", "An insight.", "novelty", "depth")
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 2 || r[0] != 0.72 || r[1] != 0.15 {
		t.Fatalf("ratings = %v", r)
	}
	if !strings.Contains(f.prompt, "An insight.") || !strings.Contains(f.prompt, "novelty, depth") {
		t.Fatalf("prompt = %q", f.prompt)
	}
}

func TestRatePercentNeverInventsNumbers(t *testing.T) {
	for name, gen := range map[string]Completer{
		"nil":        nil,
		"error":      &ratingFake{err: errors.New("down")},
		"no numbers": &ratingFake{reply: "quite deep"},
		"too few":    &ratingFake{reply: "80"},
		"over 100":   &ratingFake{reply: "180 20"},
	} {
		t.Run(name, func(t *testing.T) {
			if r, err := RatePercent(context.Background(), gen, "", "x", "a", "b"); err == nil {
				t.Fatalf("got ratings %v without a readable model rating", r)
			}
		})
	}
}
