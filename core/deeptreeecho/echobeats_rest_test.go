package deeptreeecho

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

type echobeatsCountingProvider struct {
	mu    sync.Mutex
	calls int
}

func (provider *echobeatsCountingProvider) Generate(context.Context, string, llm.GenerateOptions) (string, error) {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	provider.calls++
	return "counted response", nil
}

func (provider *echobeatsCountingProvider) StreamGenerate(context.Context, string, llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	stream := make(chan llm.StreamChunk)
	close(stream)
	return stream, nil
}

func (provider *echobeatsCountingProvider) Name() string    { return "echobeats-counting" }
func (provider *echobeatsCountingProvider) Available() bool { return true }
func (provider *echobeatsCountingProvider) MaxTokens() int  { return 4096 }

func (provider *echobeatsCountingProvider) Calls() int {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	return provider.calls
}

func TestEchobeatsPauseResumeGuardsStepsAndProviderCalls(t *testing.T) {
	provider := &echobeatsCountingProvider{}
	scheduler := NewEchobeatsScheduler(provider)

	// Pause is safe and idempotent before Start.
	scheduler.Pause()
	scheduler.Pause()
	if !scheduler.IsPaused() {
		t.Fatal("scheduler should be paused before Start")
	}

	if err := scheduler.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer func() {
		if err := scheduler.Stop(); err != nil {
			t.Errorf("Stop() error: %v", err)
		}
	}()

	// Direct calls avoid ticker timing while proving a paused scheduler performs
	// neither a step nor an LLM-provider call after Start.
	scheduler.Pause()
	scheduler.Pause()
	scheduler.executeStep()
	scheduler.executeStep()

	metrics := scheduler.GetMetrics()
	if got := metrics["total_steps"].(uint64); got != 0 {
		t.Errorf("paused total steps = %d, want 0", got)
	}
	if got := metrics["current_step"].(int); got != 1 {
		t.Errorf("paused current step = %d, want 1", got)
	}
	if got := provider.Calls(); got != 0 {
		t.Errorf("paused provider calls = %d, want 0", got)
	}

	// Resume is idempotent and continues from the unchanged step without
	// repeating prior work.
	scheduler.Resume()
	scheduler.Resume()
	if scheduler.IsPaused() {
		t.Fatal("scheduler should not be paused after Resume")
	}
	scheduler.executeStep()

	metrics = scheduler.GetMetrics()
	if got := metrics["total_steps"].(uint64); got != 1 {
		t.Errorf("resumed total steps = %d, want 1", got)
	}
	if got := metrics["current_step"].(int); got != 2 {
		t.Errorf("resumed current step = %d, want 2", got)
	}
	if got := provider.Calls(); got != 1 {
		t.Errorf("resumed provider calls = %d, want 1", got)
	}

	scheduler.Pause()
	scheduler.executeStep()
	if got := provider.Calls(); got != 1 {
		t.Errorf("provider calls after re-pause = %d, want 1", got)
	}

	scheduler.Resume()
	scheduler.executeStep()
	metrics = scheduler.GetMetrics()
	if got := metrics["total_steps"].(uint64); got != 2 {
		t.Errorf("second resumed total steps = %d, want 2", got)
	}
	if got := metrics["current_step"].(int); got != 3 {
		t.Errorf("second resumed current step = %d, want 3", got)
	}
	if got := provider.Calls(); got != 2 {
		t.Errorf("second resumed provider calls = %d, want 2", got)
	}
}

func TestEchobeatsGetActiveGoalReturnsDefensiveCopy(t *testing.T) {
	scheduler := NewEchobeatsScheduler(&echobeatsCountingProvider{})
	scheduler.AddGoal("defensive snapshot", 1.0)

	deadline := time.Now().Add(time.Hour)
	scheduler.mu.Lock()
	scheduler.goalQueue[0].Dependencies = []string{"original dependency"}
	scheduler.goalQueue[0].Deadline = &deadline
	scheduler.mu.Unlock()

	snapshot := scheduler.GetActiveGoal()
	if snapshot == nil {
		t.Fatal("GetActiveGoal() returned nil")
	}
	snapshot.Description = "mutated snapshot"
	snapshot.Dependencies[0] = "mutated dependency"
	*snapshot.Deadline = snapshot.Deadline.Add(time.Hour)

	scheduler.mu.RLock()
	internal := scheduler.goalQueue[0]
	scheduler.mu.RUnlock()

	if internal.Description != "defensive snapshot" {
		t.Errorf("internal description changed to %q", internal.Description)
	}
	if internal.Dependencies[0] != "original dependency" {
		t.Errorf("internal dependency changed to %q", internal.Dependencies[0])
	}
	if !internal.Deadline.Equal(deadline) {
		t.Errorf("internal deadline changed to %v, want %v", internal.Deadline, deadline)
	}
}

func TestEchobeatsAffordanceUsesObservedEnactionCallback(t *testing.T) {
	provider := &echobeatsCountingProvider{}
	scheduler := NewEchobeatsScheduler(provider)
	callbackCalls := 0
	scheduler.SetOnAffordance(func(step int) (string, bool) {
		callbackCalls++
		if step != 2 {
			t.Fatalf("unexpected affordance step: %d", step)
		}
		return "verified bounded effect", true
	})

	scheduler.affordanceInteraction(2)
	if callbackCalls != 1 {
		t.Fatalf("affordance callback calls=%d, want 1", callbackCalls)
	}
	if provider.Calls() != 0 {
		t.Fatalf("prose-only affordance provider was called %d times", provider.Calls())
	}
	scheduler.engine1.mu.RLock()
	defer scheduler.engine1.mu.RUnlock()
	if len(scheduler.engine1.taskHistory) != 1 || !scheduler.engine1.taskHistory[0].Success {
		t.Fatalf("verified action was not recorded in engine history: %#v", scheduler.engine1.taskHistory)
	}
}
