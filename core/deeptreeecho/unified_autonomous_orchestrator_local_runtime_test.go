package deeptreeecho

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/backendcap"
	"github.com/cogpy/echo9llama/core/llm"
)

type orchestratorLocalRuntime struct {
	mu             sync.Mutex
	warmups        int
	cooldowns      int
	memoryChecks   int
	idleChecks     int
	closes         int
	warmErr        error
	memoryUnloaded bool
	state          llm.LocalModelRegistryState
}

func (runtime *orchestratorLocalRuntime) Warmup(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	runtime.mu.Lock()
	runtime.warmups++
	err := runtime.warmErr
	runtime.mu.Unlock()
	return err
}

func (runtime *orchestratorLocalRuntime) Cooldown(string) bool {
	runtime.mu.Lock()
	runtime.cooldowns++
	runtime.mu.Unlock()
	return true
}

func (runtime *orchestratorLocalRuntime) Refresh() llm.LocalModelRegistryState {
	return runtime.State()
}

func (runtime *orchestratorLocalRuntime) UnloadIdle(string) bool {
	runtime.mu.Lock()
	runtime.idleChecks++
	runtime.mu.Unlock()
	return true
}

func (runtime *orchestratorLocalRuntime) MaybeUnloadForMemoryPressure(string) bool {
	runtime.mu.Lock()
	runtime.memoryChecks++
	unloaded := runtime.memoryUnloaded
	runtime.mu.Unlock()
	return unloaded
}

func (runtime *orchestratorLocalRuntime) RuntimeReadiness() bool { return runtime.State().RuntimeReady }

func (runtime *orchestratorLocalRuntime) State() llm.LocalModelRegistryState {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	return runtime.state
}

func (runtime *orchestratorLocalRuntime) Close() error {
	runtime.mu.Lock()
	runtime.closes++
	runtime.mu.Unlock()
	return nil
}

func (runtime *orchestratorLocalRuntime) counts() (warmups, cooldowns, memoryChecks, idleChecks, closes int) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	return runtime.warmups, runtime.cooldowns, runtime.memoryChecks, runtime.idleChecks, runtime.closes
}

type orchestratorRoutedProvider struct {
	runtime *orchestratorLocalRuntime
	backend llm.BackendRoutingState
}

func (provider *orchestratorRoutedProvider) Generate(ctx context.Context, _ string, _ llm.GenerateOptions) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return "routed cognition", nil
}

func (provider *orchestratorRoutedProvider) StreamGenerate(ctx context.Context, _ string, _ llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	output := make(chan llm.StreamChunk, 2)
	output <- llm.StreamChunk{Content: "routed cognition"}
	output <- llm.StreamChunk{Done: true}
	close(output)
	return output, nil
}
func (provider *orchestratorRoutedProvider) Name() string    { return "capability-router" }
func (provider *orchestratorRoutedProvider) Available() bool { return true }
func (provider *orchestratorRoutedProvider) MaxTokens() int  { return 8192 }
func (provider *orchestratorRoutedProvider) LocalRuntime() llm.LocalRuntimeController {
	return provider.runtime
}

func (provider *orchestratorRoutedProvider) GetBackendState() llm.BackendRoutingState {
	state := provider.backend
	state.LocalModel = provider.runtime.State()
	return state
}

func localRuntimeTestConfig() OrchestratorConfig {
	config := DefaultOrchestratorConfig()
	config.EnableStreamOfConsciousness = false
	config.EnableEchobeats = false
	config.EnableEchodream = false
	config.EnableDiscussionMonitoring = false
	config.EnableSkillLearning = false
	config.EnableWisdomSynthesis = false
	config.EnablePersistence = false
	config.AutoWakeRest = false
	config.WarmLocalModelOnWake = true
	config.CoolLocalModelOnRest = true
	config.LocalModelWarmupTimeout = 100 * time.Millisecond
	config.MainLoopInterval = time.Hour
	config.GoalReviewInterval = time.Hour
	config.WisdomSynthesisInterval = time.Hour
	config.StateSyncInterval = time.Hour
	return config
}

func TestUnifiedOrchestratorCoordinatesLocalRuntimeLifecycle(t *testing.T) {
	runtime := &orchestratorLocalRuntime{state: llm.LocalModelRegistryState{
		RuntimeReady: true,
		MemorySafe:   true,
		Loaded:       true,
		SelectedModel: backendcap.Capability{
			ModelID: "echo-native",
			Kind:    backendcap.BackendNativeCPU,
		},
	}}
	provider := &orchestratorRoutedProvider{runtime: runtime}
	orchestrator := NewUnifiedAutonomousOrchestrator(provider, localRuntimeTestConfig())
	if orchestrator.localRuntime != runtime {
		t.Fatal("unified orchestrator did not bind the router-owned runtime")
	}
	if err := orchestrator.Awaken(); err != nil {
		t.Fatal(err)
	}
	if warmups, _, _, _, _ := runtime.counts(); warmups != 1 {
		t.Fatalf("wake warmups=%d", warmups)
	}
	if err := orchestrator.onRest(); err != nil {
		t.Fatal(err)
	}
	if _, cooldowns, _, _, _ := runtime.counts(); cooldowns != 1 {
		t.Fatalf("rest cooldowns=%d", cooldowns)
	}
	if err := orchestrator.onWake(); err != nil {
		t.Fatal(err)
	}
	if warmups, _, _, _, _ := runtime.counts(); warmups != 2 {
		t.Fatalf("post-dream warmups=%d", warmups)
	}
	orchestrator.performCognitiveCycle()
	if _, _, memoryChecks, idleChecks, _ := runtime.counts(); memoryChecks != 1 || idleChecks != 1 {
		t.Fatalf("unexpected residency maintenance memory=%d idle=%d", memoryChecks, idleChecks)
	}
	if err := orchestrator.Sleep(); err != nil {
		t.Fatal(err)
	}
	if _, _, _, _, closes := runtime.counts(); closes != 1 {
		t.Fatalf("terminal closes=%d", closes)
	}
}

func TestUnifiedOrchestratorWarmupFailureDegradesWithoutBlockingAwaken(t *testing.T) {
	runtime := &orchestratorLocalRuntime{warmErr: errors.New("model load failed")}
	provider := &orchestratorRoutedProvider{runtime: runtime}
	orchestrator := NewUnifiedAutonomousOrchestrator(provider, localRuntimeTestConfig())
	if err := orchestrator.Awaken(); err != nil {
		t.Fatalf("local warmup failure blocked routed autonomy: %v", err)
	}
	if !orchestrator.GetStatus().Running {
		t.Fatal("orchestrator did not remain running after local warmup failure")
	}
	if err := orchestrator.Sleep(); err != nil {
		t.Fatal(err)
	}
}

func TestUnifiedOrchestratorMemoryPressureSkipsIdlePass(t *testing.T) {
	runtime := &orchestratorLocalRuntime{memoryUnloaded: true}
	provider := &orchestratorRoutedProvider{runtime: runtime}
	config := localRuntimeTestConfig()
	config.WarmLocalModelOnWake = false
	orchestrator := NewUnifiedAutonomousOrchestrator(provider, config)
	orchestrator.performCognitiveCycle()
	_, _, memoryChecks, idleChecks, _ := runtime.counts()
	if memoryChecks != 1 || idleChecks != 0 {
		t.Fatalf("memory pressure should short-circuit idle policy: memory=%d idle=%d", memoryChecks, idleChecks)
	}
}

func TestUnifiedOrchestratorStatusIncludesScrubbedBackendState(t *testing.T) {
	runtime := &orchestratorLocalRuntime{state: llm.LocalModelRegistryState{
		SelectedModel: backendcap.Capability{ModelID: "opaque-native", ModelPath: "", Kind: backendcap.BackendNativeCPU},
		MemorySafe:    true,
		RuntimeReady:  true,
	}}
	provider := &orchestratorRoutedProvider{
		runtime: runtime,
		backend: llm.BackendRoutingState{
			SelectedProvider: "local_gguf",
			SelectedModelID:  "opaque-native",
			SelectedKind:     backendcap.BackendNativeCPU,
		},
	}
	orchestrator := NewUnifiedAutonomousOrchestrator(provider, localRuntimeTestConfig())
	status := orchestrator.GetStatus()
	if status.Backend.SelectedProvider != "local_gguf" || status.Backend.LocalModel.SelectedModel.ModelID != "opaque-native" {
		t.Fatalf("backend status missing: %#v", status.Backend)
	}
	if status.Backend.LocalModel.SelectedModel.ModelPath != "" {
		t.Fatalf("backend status leaked model path: %#v", status.Backend)
	}
}
