//go:build cgo

package deeptreeecho

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/backendcap"
	"github.com/cogpy/echo9llama/core/llm"
)

type routedEnactionProvider struct {
	response string
	calls    int
}

func (provider *routedEnactionProvider) Generate(_ context.Context, _ string, _ llm.GenerateOptions) (string, error) {
	provider.calls++
	return provider.response, nil
}

func (provider *routedEnactionProvider) StreamGenerate(context.Context, string, llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	stream := make(chan llm.StreamChunk)
	close(stream)
	return stream, nil
}
func (*routedEnactionProvider) Name() string    { return "routed-enaction-test" }
func (*routedEnactionProvider) Available() bool { return true }
func (*routedEnactionProvider) MaxTokens() int  { return 4096 }
func (*routedEnactionProvider) GetRouteResult(traceID string) (llm.BackendRoutingState, bool) {
	return llm.BackendRoutingState{
		TraceID: traceID, SelectedProvider: "routed-enaction-test", SelectedKind: backendcap.BackendRemoteAPI,
		SelectedModelID: "test-model", Decision: backendcap.Decision{Reason: "test real route"},
		Attempts: []llm.RouteAttempt{{Provider: "routed-enaction-test", BackendKind: backendcap.BackendRemoteAPI, Success: true}},
	}, true
}

func TestUnifiedOrchestratorProjectsVerifiedEnactionAndQuiescesAtRest(t *testing.T) {
	provider := &routedEnactionProvider{response: testPlanJSON(t, "briefs/orchestrator-1.md", "# Orchestrated Evidence\n\nWhat observation should be tested next?\n")}
	config := DefaultOrchestratorConfig()
	config.StateDirectory = t.TempDir()
	config.SessionName = "orchestrator-e1-test"
	config.EnactionMode = EnactionLocalSandbox
	config.EnableStreamOfConsciousness = false
	config.EnableDiscussionMonitoring = false
	config.EnableWisdomSynthesis = false
	config.AutoWakeRest = false
	config.MainLoopInterval = time.Hour
	config.GoalReviewInterval = time.Hour
	config.StateSyncInterval = time.Hour

	orchestrator := NewUnifiedAutonomousOrchestrator(provider, config)
	if orchestrator.eventStore == nil || orchestrator.enactionPipeline == nil {
		t.Fatal("production enaction dependencies were not initialized")
	}
	t.Cleanup(func() {
		orchestrator.dreamCycle.Shutdown()
		_ = orchestrator.eventStore.Close()
	})
	goalID := orchestrator.echobeatsScheduler.AddGoal("Test persistent autonomous cognition safely", 0.9)
	orchestrator.echobeatsScheduler.mu.RLock()
	affordance := orchestrator.echobeatsScheduler.onAffordance
	orchestrator.echobeatsScheduler.mu.RUnlock()
	if affordance == nil {
		t.Fatal("Echobeats enaction callback was not bound")
	}
	if summary, verified := affordance(2); !verified || summary == "" {
		t.Fatalf("Echobeats-bound enaction did not verify: %q", summary)
	}

	goal := orchestrator.echobeatsScheduler.GetActiveGoal()
	if goal == nil || goal.ID != goalID || goal.Progress != 0.1 {
		t.Fatalf("verified evidence did not advance selected goal: %#v", goal)
	}
	skill, err := orchestrator.skillLearning.GetSkillByID(stableSkillID(structuredArtifactSkill))
	if err != nil || skill.Proficiency <= 0 || skill.PracticeCount != 1 {
		t.Fatalf("verified evidence did not update skill: skill=%#v err=%v", skill, err)
	}
	if _, err := os.Stat(filepath.Join(config.StateDirectory, "workspace", "briefs", "orchestrator-1.md")); err != nil {
		t.Fatalf("verified artifact missing: %v", err)
	}
	status := orchestrator.GetStatus()
	if !status.EventLedgerReady || status.EventCount != 11 || !status.EnactionEnabled {
		t.Fatalf("unexpected E1 status: %#v", status)
	}

	affordance(3)
	if provider.calls != 1 {
		t.Fatalf("same-wake goal was planned more than once: %d", provider.calls)
	}
	goal = orchestrator.echobeatsScheduler.GetActiveGoal()
	if goal.Progress != 0.1 {
		t.Fatalf("replayed evidence advanced goal twice: %#v", goal)
	}

	orchestrator.mu.Lock()
	orchestrator.isAwake = false
	orchestrator.mu.Unlock()
	orchestrator.enactionPipeline.Pause()
	affordance(4)
	if provider.calls != 1 {
		t.Fatalf("resting orchestrator called action planner: %d", provider.calls)
	}
}
