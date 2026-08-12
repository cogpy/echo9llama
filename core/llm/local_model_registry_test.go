package llm

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/backendcap"
)

type registryTestRuntime struct {
	mu          sync.Mutex
	capability  backendcap.Capability
	loaded      bool
	closed      bool
	loadErr     error
	warmErr     error
	generateErr error
	closeCalls  int
}

func (runtime *registryTestRuntime) Generate(ctx context.Context, _ string, _ GenerateOptions) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	if runtime.closed {
		return "", ErrLocalGGUFClosed
	}
	if runtime.generateErr != nil {
		return "", runtime.generateErr
	}
	runtime.loaded = true
	return "local result", nil
}

func (runtime *registryTestRuntime) StreamGenerate(ctx context.Context, prompt string, options GenerateOptions) (<-chan StreamChunk, error) {
	result, err := runtime.Generate(ctx, prompt, options)
	if err != nil {
		return nil, err
	}
	output := make(chan StreamChunk, 2)
	output <- StreamChunk{Content: result}
	output <- StreamChunk{Done: true}
	close(output)
	return output, nil
}

func (runtime *registryTestRuntime) Name() string { return runtime.capability.ProviderName }
func (runtime *registryTestRuntime) Available() bool {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	return !runtime.closed && runtime.capability.Available
}
func (runtime *registryTestRuntime) MaxTokens() int { return runtime.capability.ContextLength }
func (runtime *registryTestRuntime) CapabilitySnapshot() backendcap.Capability {
	return runtime.capability
}

func (runtime *registryTestRuntime) Warmup(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	if runtime.closed {
		return ErrLocalGGUFClosed
	}
	if runtime.warmErr != nil {
		runtime.loadErr = runtime.warmErr
		return runtime.warmErr
	}
	runtime.loaded = true
	runtime.loadErr = nil
	return nil
}

func (runtime *registryTestRuntime) State() LocalGGUFState {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	state := LocalGGUFState{
		Name:                 runtime.capability.ProviderName,
		ModelID:              runtime.capability.ModelID,
		ContextSize:          runtime.capability.ContextLength,
		EstimatedMemoryBytes: runtime.capability.EstimatedMemoryBytes,
		NativeBuild:          true,
		Available:            !runtime.closed,
		Loaded:               runtime.loaded,
		Closed:               runtime.closed,
	}
	if runtime.loadErr != nil {
		state.LoadError = runtime.loadErr.Error()
	}
	return state
}

func (runtime *registryTestRuntime) Loaded() bool {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	return runtime.loaded
}

func (runtime *registryTestRuntime) LoadError() error {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	return runtime.loadErr
}

func (runtime *registryTestRuntime) Close() error {
	runtime.mu.Lock()
	runtime.closed = true
	runtime.loaded = false
	runtime.closeCalls++
	runtime.mu.Unlock()
	return nil
}

func (runtime *registryTestRuntime) Reset() error {
	runtime.mu.Lock()
	runtime.closed = false
	runtime.loadErr = nil
	runtime.mu.Unlock()
	return nil
}

func testLocalCapability(id string, contextLength int, estimated uint64) backendcap.Capability {
	return backendcap.Capability{
		Name:                 "model:" + id,
		ProviderName:         "local_gguf",
		ModelID:              id,
		ModelPath:            "/private/" + id + ".gguf",
		Kind:                 backendcap.BackendNativeCPU,
		Available:            true,
		Native:               true,
		Offline:              true,
		Concrete:             true,
		MemorySafe:           true,
		MemoryTier:           backendcap.MemoryStandard,
		ContextLength:        contextLength,
		EstimatedMemoryBytes: estimated,
		MaxConcurrency:       1,
		Priority:             20,
	}
}

func TestLocalModelRegistryLifecycleReloadAndPrivacy(t *testing.T) {
	now := time.Unix(1_000, 0)
	host := backendcap.HostMemoryProbe{Known: true, TotalBytes: 32 << 30, AvailableBytes: 24 << 30, Tier: backendcap.MemoryStress, Reason: "test"}
	capability := testLocalCapability("echo-small", 4096, 2<<30)
	var runtimes []*registryTestRuntime
	var events []LocalModelEvent
	registry := NewLocalModelRegistry(LocalModelRegistryOptions{
		ModelPaths:      []string{"/private/echo-small.gguf"},
		ModelRoots:      []string{"/private"},
		IdleUnloadAfter: time.Minute,
		Now:             func() time.Time { return now },
		ProbeHostMemory: func() backendcap.HostMemoryProbe { return host },
		discoverModels: func([]string, backendcap.DiscoveryOptions) []backendcap.Capability {
			return []backendcap.Capability{capability}
		},
		providerFactory: func(selected backendcap.Capability) localGGUFRuntime {
			runtime := &registryTestRuntime{capability: selected}
			runtimes = append(runtimes, runtime)
			return runtime
		},
		OnEvent: func(event LocalModelEvent) { events = append(events, event) },
	})

	state := registry.State()
	if state.SelectedModel.ModelID != "echo-small" || state.SelectedModel.ModelPath != "" || !state.MemorySafe {
		t.Fatalf("unexpected scrubbed initial state: %#v", state)
	}
	if err := registry.Warmup(context.Background()); err != nil {
		t.Fatalf("warmup: %v", err)
	}
	state = registry.State()
	if !state.Loaded || !state.RuntimeReady || len(runtimes) != 1 {
		t.Fatalf("unexpected warm state: %#v runtimes=%d", state, len(runtimes))
	}
	if len(events) == 0 {
		t.Fatal("expected lifecycle events")
	}
	for _, event := range events {
		if event.Capability.ModelPath != "" {
			t.Fatalf("event leaked model path: %#v", event)
		}
	}

	now = now.Add(2 * time.Minute)
	if !registry.UnloadIdle("test idle") {
		t.Fatal("expected idle unload")
	}
	if state = registry.State(); state.Loaded {
		t.Fatalf("registry remained loaded after idle unload: %#v", state)
	}
	provider := registry.Provider()
	if provider == nil {
		t.Fatal("registry wrapper should remain available for lazy reload")
	}
	if result, err := provider.Generate(context.Background(), "resume", GenerateOptions{}); err != nil || result != "local result" {
		t.Fatalf("lazy reload generation result=%q err=%v", result, err)
	}
	if len(runtimes) != 2 || !registry.State().Loaded {
		t.Fatalf("provider was not recreated after unload: runtimes=%d state=%#v", len(runtimes), registry.State())
	}
	if err := registry.Close(); err != nil {
		t.Fatal(err)
	}
	if registry.Provider() != nil || !registry.State().Closed {
		t.Fatalf("terminal close did not hold: %#v", registry.State())
	}
	registry.Refresh()
	if !registry.State().Closed || registry.Provider() != nil {
		t.Fatalf("refresh reopened a terminally closed registry: %#v", registry.State())
	}
}

func TestLocalModelRegistryMemoryPressureAndSelection(t *testing.T) {
	host := backendcap.HostMemoryProbe{Known: true, TotalBytes: 32 << 30, AvailableBytes: 24 << 30, Tier: backendcap.MemoryStress, Reason: "initial"}
	inadequate := testLocalCapability("too-small-context", 1024, 1<<30)
	adequateSmall := testLocalCapability("adequate-small", 4096, 2<<30)
	adequateLarge := testLocalCapability("adequate-large", 8192, 8<<30)
	var runtime *registryTestRuntime
	registry := NewLocalModelRegistry(LocalModelRegistryOptions{
		ModelPaths:         []string{"models"},
		MemorySafetyRatio:  0.8,
		MemoryReserveBytes: 512 << 20,
		SelectionTask: ModelSelectionTask{
			RequiredContextTokens: 3000,
			ExpectedOutputTokens:  500,
		},
		ProbeHostMemory: func() backendcap.HostMemoryProbe { return host },
		discoverModels: func([]string, backendcap.DiscoveryOptions) []backendcap.Capability {
			return []backendcap.Capability{inadequate, adequateLarge, adequateSmall}
		},
		providerFactory: func(selected backendcap.Capability) localGGUFRuntime {
			runtime = &registryTestRuntime{capability: selected}
			return runtime
		},
	})
	if selected := registry.State().SelectedModel.ModelID; selected != "adequate-small" {
		t.Fatalf("expected smallest adequate model, got %q", selected)
	}
	if err := registry.Warmup(context.Background()); err != nil {
		t.Fatal(err)
	}
	host = backendcap.HostMemoryProbe{Known: true, TotalBytes: 2 << 30, AvailableBytes: 1 << 30, Tier: backendcap.MemoryConstrained, Reason: "pressure"}
	if !registry.MaybeUnloadForMemoryPressure("test pressure") {
		t.Fatal("expected memory-pressure unload")
	}
	if runtime.closeCalls != 1 || registry.State().Loaded {
		t.Fatalf("runtime not closed under pressure: closes=%d state=%#v", runtime.closeCalls, registry.State())
	}
}

func TestLocalModelRegistryCallbacksRunOutsideRegistryLock(t *testing.T) {
	capability := testLocalCapability("callback-model", 4096, 1<<30)
	host := backendcap.HostMemoryProbe{Known: true, TotalBytes: 16 << 30, AvailableBytes: 12 << 30, Tier: backendcap.MemoryStress}
	var registry *LocalModelRegistry
	callbackDone := make(chan struct{}, 1)
	registry = NewLocalModelRegistry(LocalModelRegistryOptions{
		ModelPaths:      []string{"model"},
		ProbeHostMemory: func() backendcap.HostMemoryProbe { return host },
		discoverModels: func([]string, backendcap.DiscoveryOptions) []backendcap.Capability {
			return []backendcap.Capability{capability}
		},
		providerFactory: func(selected backendcap.Capability) localGGUFRuntime {
			return &registryTestRuntime{capability: selected}
		},
	})
	registry.options.OnEvent = func(LocalModelEvent) {
		_ = registry.State()
		select {
		case callbackDone <- struct{}{}:
		default:
		}
	}
	refreshDone := make(chan struct{})
	go func() {
		registry.Refresh()
		close(refreshDone)
	}()
	select {
	case <-refreshDone:
	case <-time.After(time.Second):
		t.Fatal("refresh deadlocked while callback read state")
	}
	select {
	case <-callbackDone:
	case <-time.After(time.Second):
		t.Fatal("expected callback")
	}
}

func TestLocalModelRegistryWarmupFailureIsCategorized(t *testing.T) {
	capability := testLocalCapability("broken", 4096, 1<<30)
	warmErr := errors.New("load failed")
	var event LocalModelEvent
	registry := NewLocalModelRegistry(LocalModelRegistryOptions{
		ModelPaths: []string{"model"},
		ProbeHostMemory: func() backendcap.HostMemoryProbe {
			return backendcap.HostMemoryProbe{Known: true, TotalBytes: 16 << 30, AvailableBytes: 12 << 30, Tier: backendcap.MemoryStress}
		},
		discoverModels: func([]string, backendcap.DiscoveryOptions) []backendcap.Capability {
			return []backendcap.Capability{capability}
		},
		providerFactory: func(selected backendcap.Capability) localGGUFRuntime {
			return &registryTestRuntime{capability: selected, warmErr: warmErr}
		},
		OnEvent: func(candidate LocalModelEvent) {
			if candidate.Type == ModelLifecycleLoadFailed {
				event = candidate
			}
		},
	})
	if err := registry.Warmup(context.Background()); !errors.Is(err, warmErr) {
		t.Fatalf("expected warmup failure, got %v", err)
	}
	if event.Type != ModelLifecycleLoadFailed || event.ErrorCategory != "runtime" || event.Capability.ModelPath != "" {
		t.Fatalf("unexpected failure event: %#v", event)
	}
}
