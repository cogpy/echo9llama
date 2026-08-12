package backendcap

import (
	"reflect"
	"strings"
	"testing"
)

func TestSnapshotContainsOnlyBuildIndependentContinuity(t *testing.T) {
	capabilities := Snapshot()
	if len(capabilities) != 1 || capabilities[0].Name != "simple_fallback" {
		t.Fatalf("expected only deterministic continuity fallback, got %+v", capabilities)
	}
	if !capabilities[0].Available || !capabilities[0].Offline || !capabilities[0].MemorySafe {
		t.Fatalf("fallback must remain safe and available: %+v", capabilities[0])
	}
}

func TestZeroWorkloadPreservesRemoteThenFallbackCompatibility(t *testing.T) {
	capabilities := []Capability{
		remoteCapability("anthropic", 30),
		fallbackCapability(),
	}
	decision := SelectFromCapabilities(Workload{}, capabilities)
	if decision.Selected.ProviderName != "anthropic" {
		t.Fatalf("zero workload should retain real remote provider preference, got %+v", decision)
	}
	if !reflect.DeepEqual(decision.Route, []string{"anthropic", "simple_fallback"}) {
		t.Fatalf("unexpected route: %+v", decision.Route)
	}
}

func TestSelectPrefersSafeConcreteNativeForOfflineWorkload(t *testing.T) {
	local := localCapability("echo-local", 4096)
	decision := SelectFromCapabilities(Workload{
		Intent:                "echobeats.affordance",
		NeedOffline:           true,
		PreferNative:          true,
		AllowFallback:         true,
		RequiredContextTokens: 512,
		ExpectedOutputTokens:  80,
	}, []Capability{remoteCapability("anthropic", 100), local, fallbackCapability()})
	if decision.Selected.ModelID != "echo-local" || !decision.Selected.Native {
		t.Fatalf("expected concrete local model, got %+v", decision)
	}
	for _, name := range decision.Route {
		if name == "anthropic" {
			t.Fatalf("offline route must not include remote provider: %+v", decision.Route)
		}
	}
}

func TestSelectRequiresRealModel(t *testing.T) {
	decision := SelectFromCapabilities(Workload{
		Intent:           "wisdom.synthesis",
		RequireRealModel: true,
		AllowFallback:    true,
	}, []Capability{fallbackCapability()})
	if decision.Selected.Name != "no_backend_available" || !decision.Degraded {
		t.Fatalf("expected no real backend, got %+v", decision)
	}
	if len(decision.Rejections) != 1 || !strings.Contains(decision.Rejections[0].Reason, "real model") {
		t.Fatalf("expected real-model rejection, got %+v", decision.Rejections)
	}
}

func TestSelectAccountsForPromptAndOutputContext(t *testing.T) {
	local := localCapability("small", 1024)
	decision := SelectFromCapabilities(Workload{
		Intent:                "context-check",
		NeedOffline:           true,
		AllowFallback:         false,
		RequiredContextTokens: 900,
		ExpectedOutputTokens:  200,
	}, []Capability{local})
	if decision.Selected.Name != "no_backend_available" {
		t.Fatalf("model should fail combined context requirement, got %+v", decision)
	}
	if len(decision.Rejections) != 1 || !strings.Contains(decision.Rejections[0].Reason, "context length") {
		t.Fatalf("expected context rejection, got %+v", decision.Rejections)
	}
}

func TestSelectRejectsSaturatedProviderWhenQueueDisallowed(t *testing.T) {
	local := localCapability("busy", 4096)
	local.MaxConcurrency = 1
	local.CurrentInFlight = 1
	decision := SelectFromCapabilities(Workload{
		Intent:        "interactive",
		PreferNative:  true,
		AllowRemote:   true,
		AllowFallback: true,
		AllowQueue:    false,
	}, []Capability{local, remoteCapability("openrouter", 10), fallbackCapability()})
	if decision.Selected.ProviderName != "openrouter" {
		t.Fatalf("expected remote capacity fallback, got %+v", decision)
	}
	if len(decision.Rejections) == 0 || !strings.Contains(decision.Rejections[0].Reason, "saturated") {
		t.Fatalf("expected saturation rejection, got %+v", decision.Rejections)
	}
}

func TestSelectStableTiePreservesConfiguredOrder(t *testing.T) {
	first := remoteCapability("first", 10)
	second := remoteCapability("second", 10)
	decision := SelectFromCapabilities(Workload{AllowRemote: true}, []Capability{first, second})
	if decision.Selected.ProviderName != "first" {
		t.Fatalf("stable tie should preserve input order, got %+v", decision.Route)
	}
}

func localCapability(modelID string, contextLength int) Capability {
	return Capability{
		Name:           "model:" + modelID,
		ProviderName:   "local_gguf",
		ModelID:        modelID,
		Kind:           BackendNativeCPU,
		Available:      true,
		Native:         true,
		Offline:        true,
		Concrete:       true,
		MemorySafe:     true,
		MemoryTier:     MemoryConstrained,
		ModelPath:      "/private/model.gguf",
		ContextLength:  contextLength,
		MaxConcurrency: 1,
		Priority:       20,
	}
}

func remoteCapability(name string, priority int) Capability {
	return Capability{
		Name:          name,
		ProviderName:  name,
		Kind:          BackendRemoteAPI,
		Available:     true,
		Concrete:      true,
		MemorySafe:    true,
		MemoryTier:    MemoryConstrained,
		ContextLength: 128_000,
		Priority:      priority,
	}
}
