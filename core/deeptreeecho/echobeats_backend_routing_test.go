package deeptreeecho

import (
	"context"
	"sync"
	"testing"

	"github.com/cogpy/echo9llama/core/backendcap"
	"github.com/cogpy/echo9llama/core/llm"
)

type tracedSchedulerProvider struct {
	mu      sync.Mutex
	options []llm.GenerateOptions
	result  string
}

func (provider *tracedSchedulerProvider) Generate(ctx context.Context, _ string, options llm.GenerateOptions) (string, error) {
	provider.mu.Lock()
	provider.options = append(provider.options, options)
	provider.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return provider.result, nil
}

func (provider *tracedSchedulerProvider) StreamGenerate(context.Context, string, llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	return nil, nil
}
func (provider *tracedSchedulerProvider) Name() string    { return "traced" }
func (provider *tracedSchedulerProvider) Available() bool { return true }
func (provider *tracedSchedulerProvider) MaxTokens() int  { return 8192 }
func (provider *tracedSchedulerProvider) GetRouteResult(traceID string) (llm.BackendRoutingState, bool) {
	return llm.BackendRoutingState{
		SelectedProvider: "local_gguf",
		SelectedModelID:  "echo-model",
		SelectedKind:     backendcap.BackendNativeCPU,
		Degraded:         true,
		Decision: backendcap.Decision{
			Reason: "native preferred by cognitive workload",
		},
		Attempts: []llm.RouteAttempt{{Provider: "local_gguf"}},
	}, traceID != ""
}

func (provider *tracedSchedulerProvider) lastOptions(t *testing.T) llm.GenerateOptions {
	t.Helper()
	provider.mu.Lock()
	defer provider.mu.Unlock()
	if len(provider.options) == 0 {
		t.Fatal("no generation options captured")
	}
	return provider.options[len(provider.options)-1]
}

func assertTaskBackendEvidence(t *testing.T, task CognitiveTask) {
	t.Helper()
	if task.Provider != "local_gguf" || task.ModelID != "echo-model" || task.BackendKind != backendcap.BackendNativeCPU {
		t.Fatalf("missing backend identity: %#v", task)
	}
	if !task.Degraded || task.RouteReason == "" || task.ProviderAttempts != 1 {
		t.Fatalf("missing route evidence: %#v", task)
	}
}

func TestEchobeatsRelevanceRoutingContract(t *testing.T) {
	provider := &tracedSchedulerProvider{result: "relevant direction"}
	scheduler := NewEchobeatsScheduler(provider)
	scheduler.relevanceRealization("What is most relevant now?")

	options := provider.lastOptions(t)
	if options.Routing.Intent != "echobeats.relevance" || !options.Routing.PreferNative || !options.Routing.AllowQueue {
		t.Fatalf("unexpected relevance routing: %#v", options.Routing)
	}
	if options.Routing.RequiredContextTokens != 512 || options.Routing.LatencyClass != backendcap.LatencyInteractive {
		t.Fatalf("unexpected relevance workload sizing: %#v", options.Routing)
	}
	scheduler.engine1.mu.RLock()
	task := scheduler.engine1.taskHistory[len(scheduler.engine1.taskHistory)-1]
	scheduler.engine1.mu.RUnlock()
	if options.Routing.TraceID != task.ID || !task.Success {
		t.Fatalf("task trace or success mismatch options=%#v task=%#v", options.Routing, task)
	}
	assertTaskBackendEvidence(t, task)
}

func TestEchobeatsAffordanceRoutingRejectsQueueing(t *testing.T) {
	provider := &tracedSchedulerProvider{result: "take action"}
	scheduler := NewEchobeatsScheduler(provider)
	scheduler.mu.Lock()
	scheduler.presentCommitment = "practice wisely"
	scheduler.mu.Unlock()
	scheduler.affordanceInteraction(3)

	options := provider.lastOptions(t)
	if options.Routing.Intent != "echobeats.affordance" || !options.Routing.PreferNative || options.Routing.AllowQueue {
		t.Fatalf("unexpected affordance routing: %#v", options.Routing)
	}
	scheduler.engine2.mu.RLock()
	task := scheduler.engine2.taskHistory[len(scheduler.engine2.taskHistory)-1]
	scheduler.engine2.mu.RUnlock()
	assertTaskBackendEvidence(t, task)
}

func TestEchobeatsSalienceRoutingAllowsBackgroundQueue(t *testing.T) {
	provider := &tracedSchedulerProvider{result: "possible future"}
	scheduler := NewEchobeatsScheduler(provider)
	scheduler.salienceSimulation(3)

	options := provider.lastOptions(t)
	if options.Routing.Intent != "echobeats.salience" || options.Routing.PreferNative || !options.Routing.AllowQueue {
		t.Fatalf("unexpected salience routing: %#v", options.Routing)
	}
	if options.Routing.LatencyClass != backendcap.LatencyNormal || options.Routing.RequiredContextTokens != 768 {
		t.Fatalf("unexpected salience workload sizing: %#v", options.Routing)
	}
	scheduler.engine3.mu.RLock()
	task := scheduler.engine3.taskHistory[len(scheduler.engine3.taskHistory)-1]
	scheduler.engine3.mu.RUnlock()
	assertTaskBackendEvidence(t, task)
}

func TestEchobeatsGenerationUsesSchedulerCancellation(t *testing.T) {
	provider := &tracedSchedulerProvider{result: "must not be used"}
	scheduler := NewEchobeatsScheduler(provider)
	scheduler.cancel()
	scheduler.relevanceRealization("What is most relevant now?")

	scheduler.engine1.mu.RLock()
	task := scheduler.engine1.taskHistory[len(scheduler.engine1.taskHistory)-1]
	scheduler.engine1.mu.RUnlock()
	if task.Success {
		t.Fatalf("canceled scheduler task reported success: %#v", task)
	}
	if task.Result == "must not be used" {
		t.Fatalf("canceled provider result was accepted: %#v", task)
	}
}
