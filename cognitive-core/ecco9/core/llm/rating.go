package llm

import (
	"context"
	"fmt"
	"regexp"
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

// ParsePercentRatings reads the first n numbers in reply. Each must be a
// whole integer in [0,100] — a sign, a decimal or an out-of-range value
// makes the reply unreadable rather than a different rating. It returns
// them scaled to [0,1].
func ParsePercentRatings(reply string, n int) ([]float64, bool) {
	nums := numericLiteral.FindAllString(reply, n)
	if len(nums) < n {
		return nil, false
	}
	out := make([]float64, 0, n)
	for _, lit := range nums {
		v, err := strconv.Atoi(lit)
		if err != nil || v < 0 || v > 100 {
			return nil, false
		}
		out = append(out, float64(v)/100)
	}
	return out, true
}

// numericLiteral matches a whole number token, sign and decimals included,
// so "-20" or "75.5" is seen as one (invalid) value, not as "20" or "75","5".
var numericLiteral = regexp.MustCompile(`[-+]?\d+(?:\.\d+)?`)
