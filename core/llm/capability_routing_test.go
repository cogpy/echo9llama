package llm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/cogpy/echo9llama/core/backendcap"
)

type routingTestProvider struct {
	mu          sync.Mutex
	name        string
	capability  backendcap.Capability
	result      string
	err         error
	chunks      []StreamChunk
	calls       int
	streamCalls int
}

func newRoutingTestProvider(name string, kind backendcap.BackendKind) *routingTestProvider {
	return &routingTestProvider{
		name: name,
		capability: backendcap.Capability{
			Name:           name,
			ProviderName:   name,
			Kind:           kind,
			Available:      true,
			Offline:        kind != backendcap.BackendRemoteAPI,
			Native:         kind == backendcap.BackendNativeCPU,
			Concrete:       kind != backendcap.BackendFallback,
			MemorySafe:     true,
			MemoryTier:     backendcap.MemoryConstrained,
			ContextLength:  8192,
			MaxConcurrency: 1,
		},
	}
}

func (provider *routingTestProvider) Generate(ctx context.Context, _ string, _ GenerateOptions) (string, error) {
	provider.mu.Lock()
	provider.calls++
	provider.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return provider.result, provider.err
}

func (provider *routingTestProvider) StreamGenerate(ctx context.Context, _ string, _ GenerateOptions) (<-chan StreamChunk, error) {
	provider.mu.Lock()
	provider.streamCalls++
	chunks := append([]StreamChunk(nil), provider.chunks...)
	err := provider.err
	provider.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err != nil && len(chunks) == 0 {
		return nil, err
	}
	output := make(chan StreamChunk, len(chunks))
	for _, chunk := range chunks {
		output <- chunk
	}
	close(output)
	return output, nil
}

func (provider *routingTestProvider) Name() string    { return provider.name }
func (provider *routingTestProvider) Available() bool { return provider.capability.Available }
func (provider *routingTestProvider) MaxTokens() int  { return provider.capability.ContextLength }
func (provider *routingTestProvider) CapabilitySnapshot() backendcap.Capability {
	return provider.capability
}

func newRoutingTestRouter(providers ...Provider) *MultiProviderLLM {
	router := &MultiProviderLLM{
		providers:    make([]Provider, 0, len(providers)),
		stats:        make(map[string]*ProviderStats),
		mode:         ProviderModeBalanced,
		routeResults: make(map[string]BackendRoutingState),
		routeOrder:   make([]string, 0, 16),
	}
	for _, provider := range providers {
		router.AddProvider(provider)
	}
	return router
}

func TestCapabilityRouterSelectsLocalForOfflineWorkload(t *testing.T) {
	local := newRoutingTestProvider("local", backendcap.BackendNativeCPU)
	local.result = "local wisdom"
	remote := newRoutingTestProvider("remote", backendcap.BackendRemoteAPI)
	remote.result = "remote wisdom"
	router := newRoutingTestRouter(local, remote, &SimpleFallbackProvider{})

	result, err := router.Generate(context.Background(), "reflect", GenerateOptions{
		MaxTokens: 128,
		Routing: RoutingOptions{
			TraceID:          "offline-wisdom",
			Intent:           "wisdom.synthesis",
			NeedOffline:      true,
			RequireRealModel: true,
			AllowFallback:    false,
			PolicyExplicit:   true,
		},
	})
	if err != nil {
		t.Fatalf("offline local generation failed: %v", err)
	}
	if result != "local wisdom" {
		t.Fatalf("unexpected result %q", result)
	}
	if local.calls != 1 || remote.calls != 0 {
		t.Fatalf("unexpected provider calls: local=%d remote=%d", local.calls, remote.calls)
	}
	state, ok := router.GetRouteResult("offline-wisdom")
	if !ok || state.SelectedProvider != "local" || state.SelectedKind != backendcap.BackendNativeCPU {
		t.Fatalf("unexpected route state: %#v", state)
	}
}

func TestCapabilityRouterFallsBackOnLocalQueueSaturation(t *testing.T) {
	local := newRoutingTestProvider("local", backendcap.BackendNativeCPU)
	local.err = ErrLocalGGUFQueueSaturated
	remote := newRoutingTestProvider("remote", backendcap.BackendRemoteAPI)
	remote.result = "remote continuity"
	router := newRoutingTestRouter(local, remote, &SimpleFallbackProvider{})

	result, err := router.Generate(context.Background(), "continue", GenerateOptions{
		MaxTokens: 64,
		Routing:   RoutingOptions{TraceID: "queue-failover", Intent: "echobeats.affordance"},
	})
	if err != nil || result != "remote continuity" {
		t.Fatalf("expected remote continuity after local saturation, result=%q err=%v", result, err)
	}
	state, ok := router.GetRouteResult("queue-failover")
	if !ok || state.SelectedProvider != "remote" || len(state.Attempts) != 2 {
		t.Fatalf("unexpected failover trace: %#v", state)
	}
	if state.Attempts[0].ErrorCategory != "queue_saturated" {
		t.Fatalf("first failure category=%q", state.Attempts[0].ErrorCategory)
	}
}

func TestCapabilityRouterRejectsFallbackForRealModelWorkload(t *testing.T) {
	router := newRoutingTestRouter(&SimpleFallbackProvider{})
	_, err := router.Generate(context.Background(), "deep synthesis", GenerateOptions{
		Routing: RoutingOptions{
			TraceID:          "real-model-only",
			RequireRealModel: true,
			PolicyExplicit:   true,
			AllowFallback:    false,
			AllowRemote:      true,
		},
	})
	if err == nil || !strings.Contains(err.Error(), "no provider satisfies workload") {
		t.Fatalf("expected hard real-model rejection, got %v", err)
	}
	state, ok := router.GetRouteResult("real-model-only")
	if !ok || len(state.Decision.Rejections) != 1 {
		t.Fatalf("expected explicit rejection telemetry: %#v", state)
	}
}

func TestCapabilityRouterStreamingFailsOverBeforeOutput(t *testing.T) {
	local := newRoutingTestProvider("local", backendcap.BackendNativeCPU)
	local.chunks = []StreamChunk{{Error: ErrLocalGGUFQueueSaturated, Done: true}}
	remote := newRoutingTestProvider("remote", backendcap.BackendRemoteAPI)
	remote.chunks = []StreamChunk{{Content: "remote"}, {Content: " stream"}, {Done: true}}
	router := newRoutingTestRouter(local, remote)

	stream, err := router.StreamGenerate(context.Background(), "stream", GenerateOptions{Routing: RoutingOptions{TraceID: "stream-failover"}})
	if err != nil {
		t.Fatalf("start stream route: %v", err)
	}
	var output strings.Builder
	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("unexpected terminal error: %v", chunk.Error)
		}
		output.WriteString(chunk.Content)
	}
	if output.String() != "remote stream" {
		t.Fatalf("unexpected stream %q", output.String())
	}
	state, _ := router.GetRouteResult("stream-failover")
	if state.SelectedProvider != "remote" || len(state.Attempts) != 2 {
		t.Fatalf("unexpected stream failover state: %#v", state)
	}
}

func TestCapabilityRouterDoesNotReplayAfterStreamingOutput(t *testing.T) {
	local := newRoutingTestProvider("local", backendcap.BackendNativeCPU)
	local.chunks = []StreamChunk{{Content: "partial"}, {Error: errors.New("decode failed"), Done: true}}
	remote := newRoutingTestProvider("remote", backendcap.BackendRemoteAPI)
	remote.chunks = []StreamChunk{{Content: "must not replay"}, {Done: true}}
	router := newRoutingTestRouter(local, remote)

	stream, err := router.StreamGenerate(context.Background(), "stream", GenerateOptions{Routing: RoutingOptions{TraceID: "stream-no-replay"}})
	if err != nil {
		t.Fatalf("start stream route: %v", err)
	}
	var output strings.Builder
	var terminalErr error
	for chunk := range stream {
		output.WriteString(chunk.Content)
		if chunk.Error != nil {
			terminalErr = chunk.Error
		}
	}
	if output.String() != "partial" || terminalErr == nil {
		t.Fatalf("unexpected partial stream result output=%q err=%v", output.String(), terminalErr)
	}
	if remote.streamCalls != 0 {
		t.Fatalf("remote stream replayed after content emission: %d", remote.streamCalls)
	}
}

func TestCapabilityRouterBoundsTraceResultsAndScrubsPaths(t *testing.T) {
	local := newRoutingTestProvider("local", backendcap.BackendNativeCPU)
	local.result = "ok"
	local.capability.ModelID = "opaque-model"
	local.capability.ModelPath = "/private/models/secret.gguf"
	router := newRoutingTestRouter(local)

	for index := range maxRouteResults + 3 {
		traceID := "trace-" + fmt.Sprint(index)
		if _, err := router.Generate(context.Background(), "x", GenerateOptions{Routing: RoutingOptions{TraceID: traceID}}); err != nil {
			t.Fatalf("trace %d: %v", index, err)
		}
	}
	if _, exists := router.GetRouteResult("trace-0"); exists {
		t.Fatal("oldest trace was not evicted")
	}
	latest, exists := router.GetRouteResult("trace-258")
	if !exists || latest.SelectedModelID != "opaque-model" {
		t.Fatalf("latest trace missing: %#v", latest)
	}
	if latest.Decision.Selected.ModelPath != "" {
		t.Fatalf("model path leaked through decision: %#v", latest.Decision.Selected)
	}
}

func TestCapabilityRouterProviderModes(t *testing.T) {
	tests := []struct {
		name     string
		mode     ProviderMode
		expected string
	}{
		{name: "balanced prefers native by a small margin", mode: ProviderModeBalanced, expected: "local"},
		{name: "local first strongly prefers native", mode: ProviderModeLocalFirst, expected: "local"},
		{name: "remote first prefers API", mode: ProviderModeRemoteFirst, expected: "remote"},
		{name: "offline excludes API", mode: ProviderModeOffline, expected: "local"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			local := newRoutingTestProvider("local", backendcap.BackendNativeCPU)
			local.result = "local"
			remote := newRoutingTestProvider("remote", backendcap.BackendRemoteAPI)
			remote.result = "remote"
			router := newRoutingTestRouter(local, remote, &SimpleFallbackProvider{})
			router.mode = test.mode
			result, err := router.Generate(context.Background(), "mode", GenerateOptions{Routing: RoutingOptions{TraceID: string(test.mode)}})
			if err != nil {
				t.Fatal(err)
			}
			if result != test.expected {
				t.Fatalf("mode %s selected %q, expected %q", test.mode, result, test.expected)
			}
			state, _ := router.GetRouteResult(string(test.mode))
			if state.SelectedProvider != test.expected {
				t.Fatalf("mode %s telemetry selected %q", test.mode, state.SelectedProvider)
			}
		})
	}
}

func TestCapabilityRouterZeroValueOptionsPreserveFallbackCompatibility(t *testing.T) {
	remote := newRoutingTestProvider("remote", backendcap.BackendRemoteAPI)
	remote.err = errors.New("remote unavailable")
	router := newRoutingTestRouter(remote, &SimpleFallbackProvider{})
	result, err := router.Generate(context.Background(), "legacy call", GenerateOptions{})
	if err != nil {
		t.Fatalf("zero-value legacy routing lost fallback continuity: %v", err)
	}
	if strings.TrimSpace(result) == "" {
		t.Fatal("deterministic fallback returned empty cognition")
	}
	state := router.GetBackendState()
	if state.SelectedKind != backendcap.BackendFallback || !state.Degraded || state.FallbackCount != 1 {
		t.Fatalf("zero-value compatibility telemetry incorrect: %#v", state)
	}
}

func TestParseProviderModeFallsBackSafely(t *testing.T) {
	if parseProviderMode("LOCAL_FIRST") != ProviderModeLocalFirst || parseProviderMode("remote_first") != ProviderModeRemoteFirst || parseProviderMode("offline") != ProviderModeOffline {
		t.Fatal("recognized provider modes were not parsed")
	}
	if parseProviderMode("unsupported") != ProviderModeBalanced {
		t.Fatal("unsupported mode must fall back to balanced")
	}
}

func TestNewMultiProviderLLMWiresModelPathsAndMode(t *testing.T) {
	t.Setenv("ECHO_PROVIDER_MODE", "offline")
	t.Setenv("ECHO_MODEL_PATHS", "missing-one.gguf,missing-two.gguf,missing-one.gguf")
	t.Setenv("LOCAL_MODEL_PATH", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENROUTER_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")

	router := NewMultiProviderLLM()
	if router.currentMode() != ProviderModeOffline {
		t.Fatalf("constructor ignored provider mode: %s", router.currentMode())
	}
	runtime := router.LocalRuntime()
	if runtime == nil {
		t.Fatal("configured model paths did not create an optional local registry")
	}
	state := runtime.State()
	if state.ConfiguredPathCount != 2 || len(state.DiscoveredModels) != 0 {
		t.Fatalf("unexpected path configuration state: %#v", state)
	}
	result, err := router.Generate(context.Background(), "continue offline", GenerateOptions{})
	if err != nil || strings.TrimSpace(result) == "" {
		t.Fatalf("fallback continuity failed with unavailable configured models: result=%q err=%v", result, err)
	}
	if err := router.Close(); err != nil {
		t.Fatal(err)
	}
	if !runtime.State().Closed {
		t.Fatal("router close did not terminate its registry-owned runtime")
	}
}

func TestSplitLocalPathListDeduplicatesSupportedSeparators(t *testing.T) {
	separator := string(os.PathListSeparator)
	paths := splitLocalPathList("one.gguf,two.gguf" + separator + "one.gguf\nthree.gguf")
	if len(paths) != 3 || paths[0] != "one.gguf" || paths[1] != "two.gguf" || paths[2] != "three.gguf" {
		t.Fatalf("unexpected path list: %#v", paths)
	}
}
