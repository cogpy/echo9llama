//go:build !cgo || nollama

package llm

import (
	"context"
	"errors"
	"testing"

	"github.com/cogpy/echo9llama/core/backendcap"
)

func TestLocalGGUFStubPreservesUnavailableContract(t *testing.T) {
	provider := NewLocalGGUFProviderFromCapability(backendcap.Capability{
		ProviderName:   "local_gguf",
		ModelID:        "stub-model",
		ModelPath:      "/private/stub.gguf",
		Kind:           backendcap.BackendNativeCPU,
		Available:      true,
		MemorySafe:     true,
		ContextLength:  4096,
		MaxConcurrency: 1,
	})
	if provider.Available() {
		t.Fatal("stub must never claim native availability")
	}
	capability := provider.CapabilitySnapshot()
	if capability.Available || capability.ModelPath != "" || capability.ModelID != "stub-model" {
		t.Fatalf("unexpected stub capability: %#v", capability)
	}
	if _, err := provider.Generate(context.Background(), "x", GenerateOptions{}); !errors.Is(err, ErrLocalGGUFUnavailable) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
	stream, err := provider.StreamGenerate(context.Background(), "x", GenerateOptions{})
	if !errors.Is(err, ErrLocalGGUFUnavailable) || stream != nil {
		t.Fatalf("stub streaming must fail before output so the router can fail over: stream=%v err=%v", stream, err)
	}
	state := provider.State()
	if state.NativeBuild || state.Available || state.ModelID != "stub-model" {
		t.Fatalf("unexpected stub state: %#v", state)
	}
	if err := provider.Close(); err != nil {
		t.Fatal(err)
	}
	if !provider.State().Closed {
		t.Fatal("stub close state not retained")
	}
	if err := provider.Reset(); err != nil {
		t.Fatal(err)
	}
	if provider.State().Closed {
		t.Fatal("stub reset did not clear terminal state")
	}
}

func TestLocalGGUFStubHonorsCancellation(t *testing.T) {
	provider := NewLocalGGUFProvider("missing.gguf")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := provider.Generate(ctx, "x", GenerateOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}
