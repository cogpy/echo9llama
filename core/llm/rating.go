package llm

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// Completer is anything that completes a prompt with a model.
type Completer interface {
	Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error)
}

// RatePercent asks the model to rate subject on each named scale from 0 to
// 100 and returns the ratings scaled to [0,1], in the order of scales. The
// numbers are the model's: an unreadable reply is an error, never a default.
func RatePercent(ctx context.Context, gen Completer, system, subject string, scales ...string) ([]float64, error) {
	if gen == nil {
		return nil, ErrNoEdgeModel
	}
	if len(scales) == 0 {
		return nil, fmt.Errorf("no rating scales given")
	}
	if strings.TrimSpace(system) == "" {
		// A turn-framed request: a bare prompt lets a small completion model
		// ramble instead of answering.
		system = "You are Deep Tree Echo. You rate things honestly and answer with numbers only."
	}
	opts := DefaultGenerateOptions()
	opts.SystemPrompt = system
	opts.MaxTokens = 8 * len(scales)
	opts.Temperature = 0.2
	prompt := fmt.Sprintf("%s\nRate it honestly. Reply with exactly %d integers from 0 to 100, separated by spaces, in this order: %s.",
		strings.TrimSpace(subject), len(scales), strings.Join(scales, ", "))
	// An unreadable reply is asked for once more; it is never filled in.
	var reply string
	for range 2 {
		out, err := gen.Generate(ctx, prompt, opts)
		if err != nil {
			return nil, err
		}
		reply = out
		if ratings, ok := ParsePercentRatings(out, len(scales)); ok {
			return ratings, nil
		}
	}
	return nil, fmt.Errorf("model rating unreadable: %q", strings.TrimSpace(reply))
}

// ParsePercentRatings reads the first n integers from reply; each must be in
// [0,100]. It returns them scaled to [0,1].
func ParsePercentRatings(reply string, n int) ([]float64, bool) {
	out := make([]float64, 0, n)
	for _, f := range strings.FieldsFunc(reply, func(r rune) bool { return r < '0' || r > '9' }) {
		v, err := strconv.Atoi(f)
		if err != nil || v > 100 {
			return nil, false
		}
		out = append(out, float64(v)/100)
		if len(out) == n {
			return out, true
		}
	}
	return nil, false
}
