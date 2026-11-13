package providers

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/cogpy/echo9llama/core/llm"
)

func TestNativeRuntimeConcurrentFirstUseCreatesOneModel(t *testing.T) {
	var r nativeRuntime
	path := filepath.Join(t.TempDir(), "absent.gguf")
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, release, err := r.acquire(path); err == nil {
				release()
			}
		}()
	}
	wg.Wait()
	first := r.model
	if first == nil || r.path != path {
		t.Fatalf("runtime not created for %s", path)
	}
	_, _, _ = r.acquire(path)
	if r.model != first {
		t.Fatal("a second runtime was created for the same model")
	}
}

func TestNativeRuntimeWithoutServedModelIsAnError(t *testing.T) {
	var r nativeRuntime
	path := filepath.Join(t.TempDir(), "absent.gguf")
	if _, err := r.generate(context.Background(), path, "hi", llm.DefaultGenerateOptions()); err == nil {
		t.Fatal("generated text without a served model")
	}
	if _, err := r.stream(context.Background(), path, "hi", llm.DefaultGenerateOptions()); err == nil {
		t.Fatal("streamed without a served model")
	}
}

func TestNativeRuntimeResetReplacesTheModel(t *testing.T) {
	var r nativeRuntime
	a := filepath.Join(t.TempDir(), "a.gguf")
	_, _, _ = r.acquire(a)
	r.reset()
	if r.model != nil || r.path != "" {
		t.Fatal("reset left the old runtime in place")
	}
	b := filepath.Join(t.TempDir(), "b.gguf")
	_, _, _ = r.acquire(b)
	if r.path != b {
		t.Fatalf("path = %s", r.path)
	}
}
