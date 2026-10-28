package llm

import (
	"context"
	"errors"
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

func TestMultiProviderFailsOverToNextModel(t *testing.T) {
	remote := &recoveryTestProvider{name: "failing-remote", available: true, err: errors.New("temporary upstream outage")}
	local := &recoveryTestProvider{name: "local-model", available: true, result: "wisdom from the local model"}
	router := &MultiProviderLLM{providers: make([]Provider, 0), stats: make(map[string]*ProviderStats)}
	router.AddProvider(remote)
	router.AddProvider(local)

	result, err := router.Generate(context.Background(), "What wisdom is relevant now?", GenerateOptions{})
	if err != nil {
		t.Fatalf("expected failover to the local model, got %v", err)
	}
	if result != "wisdom from the local model" {
		t.Fatalf("unexpected response: %q", result)
	}
	if remote.calls != 1 || local.calls != 1 {
		t.Fatalf("calls remote=%d local=%d", remote.calls, local.calls)
	}

	stats := router.GetStats()
	if stats["failing-remote"].FailedCalls != 1 {
		t.Fatalf("remote failure was not tracked: %#v", stats["failing-remote"])
	}
	if stats["local-model"].SuccessCalls != 1 {
		t.Fatalf("local success was not tracked: %#v", stats["local-model"])
	}
}

func TestMultiProviderHonorsCancellationBeforeFailover(t *testing.T) {
	remote := &recoveryTestProvider{name: "remote", available: true, result: "should not run"}
	router := &MultiProviderLLM{providers: make([]Provider, 0), stats: make(map[string]*ProviderStats)}
	router.AddProvider(remote)

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

func TestMultiProviderWithoutModelsHasNoProvider(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENROUTER_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ECHO_MODEL_PATHS", "")
	t.Setenv("LOCAL_MODEL_PATH", "")

	router := NewMultiProviderLLM()
	if len(router.providers) != 0 {
		t.Fatalf("expected no providers without a model, got %#v", router.providers)
	}
	if out, err := router.Generate(context.Background(), "hello", GenerateOptions{}); err == nil {
		t.Fatalf("expected an error without a model, got %q", out)
	}
}
