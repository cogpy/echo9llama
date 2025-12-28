//go:build !cgo
// +build !cgo

package llamarunner

import (
	"context"
	"errors"

	"github.com/cogpy/echo9llama/api"
	"github.com/cogpy/echo9llama/llm"
)

// Stub implementations when CGO is not available
// This allows the package to compile but local model execution won't work

// Runner is a stub runner
type Runner struct{}

// New creates a stub runner
func New(ctx context.Context, opts api.Options) (*Runner, error) {
	return nil, errors.New("local model runner not available (CGO disabled) - use API-based providers instead")
}

// Completion is a stub
func (r *Runner) Completion(ctx context.Context, req llm.CompletionRequest, fn func(llm.CompletionResponse)) error {
	return errors.New("local model runner not available (CGO disabled)")
}

// Embedding is a stub
func (r *Runner) Embedding(ctx context.Context, input string) ([]float32, error) {
	return nil, errors.New("local model runner not available (CGO disabled)")
}

// Close is a stub
func (r *Runner) Close() error {
	return nil
}
