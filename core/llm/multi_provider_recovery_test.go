package llm

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type recoveryTestProvider struct {
	name      string
	available bool
	result    string
	err       error
	calls     int
}

func (provider *recoveryTestProvider) Generate(ctx context.Context, _ string, _ GenerateOptions) (string, error) {
	provider.calls++
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return provider.result, provider.err
}

func (provider *recoveryTestProvider) Name() string    { return provider.name }
func (provider *recoveryTestProvider) Available() bool { return provider.available }
func (provider *recoveryTestProvider) MaxTokens() int  { return 4096 }

func TestMultiProviderFallsBackToDeterministicLocalCognition(t *testing.T) {
	remote := &recoveryTestProvider{name: "failing-remote", available: true, err: errors.New("temporary upstream outage")}
	router := &MultiProviderLLM{providers: make([]Provider, 0), stats: make(map[string]*ProviderStats)}
	router.AddProvider(remote)
	router.AddProvider(&SimpleFallbackProvider{})

	result, err := router.Generate(context.Background(), "What wisdom is relevant now?", GenerateOptions{})
	if err != nil {
		t.Fatalf("expected local fallback after remote failure, got %v", err)
	}
	if !strings.Contains(strings.ToLower(result), "wisdom") {
		t.Fatalf("unexpected fallback response: %q", result)
	}
	if remote.calls != 1 {
		t.Fatalf("expected one remote attempt, got %d", remote.calls)
	}

	stats := router.GetStats()
	if stats["failing-remote"].FailedCalls != 1 {
		t.Fatalf("remote failure was not tracked: %#v", stats["failing-remote"])
	}
	if stats["SimpleFallback"].SuccessCalls != 1 {
		t.Fatalf("fallback success was not tracked: %#v", stats["SimpleFallback"])
	}
}

func TestMultiProviderHonorsCancellationBeforeFailover(t *testing.T) {
	remote := &recoveryTestProvider{name: "remote", available: true, result: "should not run"}
	router := &MultiProviderLLM{providers: make([]Provider, 0), stats: make(map[string]*ProviderStats)}
	router.AddProvider(remote)
	router.AddProvider(&SimpleFallbackProvider{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := router.Generate(ctx, "prompt", GenerateOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if remote.calls != 0 {
		t.Fatalf("provider was called after cancellation: %d", remote.calls)
	}
	if _, err := router.StreamGenerate(ctx, "prompt", GenerateOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected streaming context cancellation, got %v", err)
	}
}

func TestMultiProviderWithoutKeysStillHasAutonomousFallback(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENROUTER_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")

	router := NewMultiProviderLLM()
	if !router.Available() {
		t.Fatal("router should remain available through deterministic local cognition")
	}
	if len(router.providers) != 1 || router.providers[0].Name() != "SimpleFallback" {
		t.Fatalf("expected exactly one local fallback provider, got %#v", router.providers)
	}
}
