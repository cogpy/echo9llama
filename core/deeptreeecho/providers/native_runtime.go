package providers

import (
	"context"
	"fmt"
	"sync"

	"github.com/cogpy/echo9llama/core/llm"
)

// nativeRuntime owns the in-process llama.cpp model a provider serves.
// Requests hold a read lock while they use the model, so a reload (which
// takes the write lock) never closes a model under a running request, and
// concurrent first requests create one runtime, not several.
type nativeRuntime struct {
	mu    sync.RWMutex
	path  string
	model *llm.LocalGGUFProvider

	errMu     sync.Mutex
	streamErr error
}

// acquire returns the native model for path with the read lock held; the
// caller calls release when it no longer uses the model.
func (r *nativeRuntime) acquire(path string) (model *llm.LocalGGUFProvider, release func(), err error) {
	for {
		r.mu.RLock()
		if r.model != nil && r.path == path {
			if !r.model.Available() {
				r.mu.RUnlock()
				return nil, nil, fmt.Errorf("local GGUF model %s is not served: native llama.cpp bindings unavailable", path)
			}
			return r.model, r.mu.RUnlock, nil
		}
		r.mu.RUnlock()

		r.mu.Lock()
		if r.model == nil || r.path != path {
			if r.model != nil {
				_ = r.model.Close()
			}
			r.model, r.path = llm.NewLocalGGUFProvider(path), path
		}
		r.mu.Unlock()
	}
}

// reset closes the current model once no request is using it.
func (r *nativeRuntime) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.model != nil {
		_ = r.model.Close()
	}
	r.model, r.path = nil, ""
}

// stream forwards the model's own tokens. A load or generation failure
// reported in the stream ends it and is kept for StreamErr; it is never
// turned into a normal end of stream silently.
func (r *nativeRuntime) stream(ctx context.Context, path, prompt string, opts llm.GenerateOptions) (<-chan string, error) {
	model, release, err := r.acquire(path)
	if err != nil {
		return nil, err
	}
	chunks, err := model.StreamGenerate(ctx, prompt, opts)
	if err != nil {
		release()
		return nil, err
	}
	r.setStreamErr(nil)
	ch := make(chan string, 16)
	go func() {
		defer release()
		defer close(ch)
		for chunk := range chunks {
			if chunk.Error != nil {
				r.setStreamErr(chunk.Error)
				return
			}
			if chunk.Content == "" {
				continue
			}
			select {
			case <-ctx.Done():
				r.setStreamErr(ctx.Err())
				return
			case ch <- chunk.Content:
			}
		}
	}()
	return ch, nil
}

func (r *nativeRuntime) setStreamErr(err error) {
	r.errMu.Lock()
	r.streamErr = err
	r.errMu.Unlock()
}

// StreamErr reports why the last stream ended early (nil when it completed).
func (r *nativeRuntime) StreamErr() error {
	r.errMu.Lock()
	defer r.errMu.Unlock()
	return r.streamErr
}

// generate runs one completion on the native model for path.
func (r *nativeRuntime) generate(ctx context.Context, path, prompt string, opts llm.GenerateOptions) (string, error) {
	model, release, err := r.acquire(path)
	if err != nil {
		return "", err
	}
	defer release()
	return model.Generate(ctx, prompt, opts)
}
