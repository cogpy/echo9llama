package deeptreeecho

import (
	"context"
	"os"
	"path/filepath"
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

func TestUnifiedAutonomousOrchestratorPersistsContinuitySnapshot(t *testing.T) {
	stateDir := t.TempDir()

	config := DefaultOrchestratorConfig()
	config.SessionName = "test-continuity-a"
	config.StateDirectory = stateDir
	config.EnablePersistence = true
	config.StateSyncInterval = time.Hour

	orchestrator := NewUnifiedAutonomousOrchestrator(&lifecycleMockLLMProvider{}, config)
	if orchestrator.persistentState == nil {
		t.Fatal("expected persistent consciousness state manager to be initialized")
	}

	orchestrator.totalCycles = 7
	orchestrator.totalThoughts = 11
	orchestrator.totalGoals = 3
	orchestrator.totalWisdom = 2
	orchestrator.cognitiveLoad = 0.42
	orchestrator.wisdomDepth = 0.73
	orchestrator.isAwake = true
	orchestrator.syncPersistentState()

	if _, err := os.Stat(filepath.Join(stateDir, "consciousness_state.json")); err != nil {
		t.Fatalf("expected consciousness state file to be written: %v", err)
	}

	persisted := orchestrator.persistentState.GetState()
	if persisted == nil {
		t.Fatal("expected persisted state snapshot")
	}
	if persisted.SessionID != "test-continuity-a" {
		t.Fatalf("expected saved session ID to match current session, got %q", persisted.SessionID)
	}
	if persisted.CycleCount != 7 || persisted.TotalThoughts != 11 || persisted.TotalGoals != 3 || persisted.TotalInsights != 2 {
		t.Fatalf("unexpected persisted counters: cycles=%d thoughts=%d goals=%d insights=%d",
			persisted.CycleCount, persisted.TotalThoughts, persisted.TotalGoals, persisted.TotalInsights)
	}
	if persisted.WakeRestState != "Awake" {
		t.Fatalf("expected persisted wake/rest state to be Awake, got %q", persisted.WakeRestState)
	}

	secondConfig := DefaultOrchestratorConfig()
	secondConfig.SessionName = "test-continuity-b"
	secondConfig.StateDirectory = stateDir
	secondConfig.EnablePersistence = true
	secondConfig.StateSyncInterval = time.Hour

	rehydrated := NewUnifiedAutonomousOrchestrator(&lifecycleMockLLMProvider{}, secondConfig)
	status := rehydrated.GetStatus()
	if status.TotalCycles != 7 || status.TotalThoughts != 11 || status.TotalGoals != 3 || status.TotalWisdom != 2 {
		t.Fatalf("unexpected hydrated counters: cycles=%d thoughts=%d goals=%d wisdom=%d",
			status.TotalCycles, status.TotalThoughts, status.TotalGoals, status.TotalWisdom)
	}
	if status.CognitiveLoad != 0.42 {
		t.Fatalf("expected hydrated cognitive load 0.42, got %.2f", status.CognitiveLoad)
	}
	if status.StateDirectory != stateDir {
		t.Fatalf("expected status state directory %q, got %q", stateDir, status.StateDirectory)
	}
	if status.LastStateSync.IsZero() {
		t.Fatal("expected hydrated status to expose last state sync time")
	}
}

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
