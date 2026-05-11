package deeptreeecho

import (
	"context"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

type lifecycleMockLLMProvider struct{}

func (m *lifecycleMockLLMProvider) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	return "mock lifecycle response", nil
}

func (m *lifecycleMockLLMProvider) StreamGenerate(ctx context.Context, prompt string, opts llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: "mock lifecycle stream", Done: true}
	close(ch)
	return ch, nil
}

func (m *lifecycleMockLLMProvider) Name() string    { return "lifecycle-mock" }
func (m *lifecycleMockLLMProvider) Available() bool { return true }
func (m *lifecycleMockLLMProvider) MaxTokens() int  { return 4096 }

func TestUnifiedAutonomousOrchestratorSleepDoesNotDeadlock(t *testing.T) {
	config := DefaultOrchestratorConfig()
	config.SessionName = "test-sleep-no-deadlock"
	config.MainLoopInterval = 25 * time.Millisecond
	config.ThoughtInterval = 25 * time.Millisecond
	config.GoalReviewInterval = 50 * time.Millisecond
	config.WisdomSynthesisInterval = 75 * time.Millisecond
	config.StateSyncInterval = 50 * time.Millisecond
	config.WakeDuration = 100 * time.Millisecond
	config.RestDuration = 100 * time.Millisecond
	config.EnablePersistence = true

	orchestrator := NewUnifiedAutonomousOrchestrator(&lifecycleMockLLMProvider{}, config)
	if err := orchestrator.Awaken(); err != nil {
		t.Fatalf("Awaken failed: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- orchestrator.Sleep()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Sleep returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Sleep deadlocked or exceeded shutdown deadline")
	}

	status := orchestrator.GetStatus()
	if status.Running {
		t.Fatal("orchestrator should not be running after Sleep")
	}
	if status.IsAwake {
		t.Fatal("orchestrator should not be awake after Sleep")
	}
}
