//go:build cgo && !nollama

package llm

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/backendcap"
)

func nativeProviderFixture() *LocalGGUFProvider {
	return NewLocalGGUFProviderWithConfig(LocalGGUFProviderConfig{
		Name:      "native-test",
		ModelPath: "/private/models/test.gguf",
		QueueWait: 50 * time.Millisecond,
		Capability: backendcap.Capability{
			Name:           "model:test",
			ProviderName:   "native-test",
			ModelID:        "test-model",
			ModelPath:      "/private/models/test.gguf",
			Kind:           backendcap.BackendNativeCPU,
			Available:      true,
			Native:         true,
			Offline:        true,
			Concrete:       true,
			MemorySafe:     true,
			ContextLength:  4096,
			MaxConcurrency: 1,
		},
	})
}

func TestLocalGGUFNativeCapabilitySnapshotScrubsPath(t *testing.T) {
	provider := nativeProviderFixture()
	capability := provider.CapabilitySnapshot()
	if capability.ModelID != "test-model" || capability.ModelPath != "" || !capability.Available {
		t.Fatalf("unexpected capability snapshot: %#v", capability)
	}
	state := provider.State()
	if state.ModelID != "test-model" || !state.NativeBuild || state.ContextSize != 4096 {
		t.Fatalf("unexpected provider state: %#v", state)
	}
}

func TestLocalGGUFNativeQueuePolicyAndCancellation(t *testing.T) {
	provider := nativeProviderFixture()
	provider.slot <- struct{}{}
	start := time.Now()
	err := provider.acquireSlot(context.Background(), RoutingOptions{PolicyExplicit: true, AllowQueue: false})
	if !errors.Is(err, ErrLocalGGUFQueueSaturated) {
		t.Fatalf("expected immediate queue saturation, got %v", err)
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Fatalf("nonqueueing acquisition waited too long: %s", time.Since(start))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancel()
	err = provider.acquireSlot(ctx, RoutingOptions{PolicyExplicit: true, AllowQueue: true, QueueWait: time.Second})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected caller cancellation to win over queue timeout, got %v", err)
	}
	<-provider.slot
}

func TestLocalGGUFNativeCloseDrainsActiveSlotAndReset(t *testing.T) {
	provider := nativeProviderFixture()
	provider.slot <- struct{}{}
	closed := make(chan error, 1)
	go func() { closed <- provider.Close() }()
	select {
	case err := <-closed:
		t.Fatalf("close returned before active slot drained: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	<-provider.slot
	select {
	case err := <-closed:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("close did not finish after slot release")
	}
	if !provider.State().Closed || provider.Available() {
		t.Fatalf("provider remained available after close: %#v", provider.State())
	}
	if err := provider.Close(); err != nil {
		t.Fatalf("idempotent close: %v", err)
	}
	if err := provider.Reset(); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if provider.State().Closed {
		t.Fatalf("reset did not reopen provider state: %#v", provider.State())
	}
}

func TestLocalGGUFNativeGenerateHonorsCanceledContextBeforeLoad(t *testing.T) {
	provider := nativeProviderFixture()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := provider.Generate(ctx, "do not load", GenerateOptions{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation before load, got %v", err)
	}
	if provider.State().Loading || provider.State().Loaded {
		t.Fatalf("canceled request touched native load state: %#v", provider.State())
	}
}

func TestStopSequenceEmitterMatchesAcrossChunkBoundaries(t *testing.T) {
	var output strings.Builder
	emitter := newStopSequenceEmitter([]string{"<STOP>", "END"}, func(content string) error {
		output.WriteString(content)
		return nil
	})
	stopped, err := emitter.Write("alpha<ST")
	if err != nil || stopped {
		t.Fatalf("unexpected first write stopped=%v err=%v", stopped, err)
	}
	stopped, err = emitter.Write("OP>ignored")
	if err != nil || !stopped {
		t.Fatalf("expected cross-chunk stop stopped=%v err=%v", stopped, err)
	}
	if err := emitter.Flush(); err != nil {
		t.Fatal(err)
	}
	if output.String() != "alpha" {
		t.Fatalf("stop sequence leaked or prefix lost: %q", output.String())
	}
}

func TestStopSequenceEmitterFlushesWithoutStop(t *testing.T) {
	var output strings.Builder
	emitter := newStopSequenceEmitter([]string{"XYZ"}, func(content string) error {
		output.WriteString(content)
		return nil
	})
	for _, chunk := range []string{"wis", "dom", " grows"} {
		stopped, err := emitter.Write(chunk)
		if err != nil || stopped {
			t.Fatalf("write %q stopped=%v err=%v", chunk, stopped, err)
		}
	}
	if err := emitter.Flush(); err != nil {
		t.Fatal(err)
	}
	if output.String() != "wisdom grows" {
		t.Fatalf("unexpected flushed content %q", output.String())
	}
}
