package deeptreeecho

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

type blockingThoughtProvider struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once
	mu      sync.Mutex
	calls   int
}

func (provider *blockingThoughtProvider) Generate(context.Context, string, llm.GenerateOptions) (string, error) {
	provider.mu.Lock()
	provider.calls++
	provider.mu.Unlock()
	provider.once.Do(func() { close(provider.entered) })
	<-provider.release
	return "A completed thought", nil
}

func (*blockingThoughtProvider) StreamGenerate(context.Context, string, llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	stream := make(chan llm.StreamChunk)
	close(stream)
	return stream, nil
}
func (*blockingThoughtProvider) Name() string    { return "blocking-thought" }
func (*blockingThoughtProvider) Available() bool { return true }
func (*blockingThoughtProvider) MaxTokens() int  { return 4096 }
func (provider *blockingThoughtProvider) Calls() int {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	return provider.calls
}

func TestStreamPauseIsGenerationQuiescenceBarrier(t *testing.T) {
	provider := &blockingThoughtProvider{entered: make(chan struct{}), release: make(chan struct{})}
	stream := NewStreamOfConsciousness(provider)
	generated := make(chan struct{})
	go func() {
		stream.generateThought()
		close(generated)
	}()

	select {
	case <-provider.entered:
	case <-time.After(time.Second):
		t.Fatal("provider call was not entered")
	}
	paused := make(chan struct{})
	go func() {
		stream.Pause()
		close(paused)
	}()
	select {
	case <-paused:
		t.Fatal("Pause returned while generation was still in flight")
	case <-time.After(25 * time.Millisecond):
	}
	close(provider.release)
	select {
	case <-generated:
	case <-time.After(time.Second):
		t.Fatal("generation did not finish")
	}
	select {
	case <-paused:
	case <-time.After(time.Second):
		t.Fatal("Pause did not return after generation quiesced")
	}

	calls := provider.Calls()
	stream.generateThought()
	if provider.Calls() != calls {
		t.Fatalf("paused stream admitted another provider call: before=%d after=%d", calls, provider.Calls())
	}
}
