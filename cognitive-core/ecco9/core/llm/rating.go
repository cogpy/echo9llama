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
		return nil, fmt.Errorf("no model available to rate")
	}
	if len(scales) == 0 {
		return nil, fmt.Errorf("no rating scales given")
	}
	opts := DefaultGenerateOptions()
	opts.SystemPrompt = system
	opts.MaxTokens = 8 * len(scales)
	opts.Temperature = 0.2
	prompt := fmt.Sprintf("%s\nRate it honestly. Reply with exactly %d integers from 0 to 100, separated by spaces, in this order: %s.",
		strings.TrimSpace(subject), len(scales), strings.Join(scales, ", "))
	reply, err := gen.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, err
	}
	ratings, ok := ParsePercentRatings(reply, len(scales))
	if !ok {
		return nil, fmt.Errorf("model rating unreadable: %q", strings.TrimSpace(reply))
	}
	return ratings, nil
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
