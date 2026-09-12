package deeptreeecho

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/cognitivecore"
	"github.com/cogpy/echo9llama/core/persistence"
)

func TestCognitiveCoreObservationIsDurableAndDreamProjectedOnce(t *testing.T) {
	stateDir := t.TempDir()
	config := DefaultOrchestratorConfig()
	config.StateDirectory = stateDir
	config.EventStorePath = filepath.Join(stateDir, "events.db")
	config.WorkspaceDirectory = filepath.Join(stateDir, "workspace")
	config.EnableEnaction = false
	config.EnableSkillLearning = false
	config.EnableDiscussionMonitoring = false
	config.EnableWisdomSynthesis = false
	config.EnableStreamOfConsciousness = false
	config.EnableEchobeats = false
	config.AutoWakeRest = false
	config.MainLoopInterval = time.Hour
	config.GoalReviewInterval = time.Hour
	config.WisdomSynthesisInterval = time.Hour
	config.StateSyncInterval = time.Hour
	config.CognitiveCoreInterval = time.Hour

	orchestrator := NewUnifiedAutonomousOrchestrator(&lifecycleMockLLMProvider{}, config)
	if orchestrator.eventStore == nil {
		t.Fatal("cognitive event ledger should not depend on enaction")
	}
	if err := orchestrator.Awaken(); err != nil {
		t.Fatalf("Awaken: %v", err)
	}
	t.Cleanup(func() {
		if orchestrator.GetStatus().Running {
			_ = orchestrator.Sleep()
		}
	})

	orchestrator.performCognitiveCycle()
	orchestrator.performCognitiveCycle()
	status := orchestrator.GetStatus()
	if !status.CognitiveCoreEnabled || !status.CognitiveCore.Started || status.CognitiveCore.ObservationCount != 1 {
		t.Fatalf("unexpected cognitive core status: %+v", status.CognitiveCore)
	}
	if status.PendingExperiences != 1 {
		t.Fatalf("pending EchoDream experiences = %d, want 1", status.PendingExperiences)
	}

	events, err := orchestrator.eventStore.Query(context.Background(), persistence.CognitiveEventQuery{
		EventType: persistence.EventTypeCoreObserved,
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("query cognitive core evidence: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("cognitive core event count = %d, want 1", len(events))
	}
	var evidence cognitiveCoreEvidence
	if err := json.Unmarshal(events[0].PayloadJSON, &evidence); err != nil {
		t.Fatalf("decode cognitive core event: %v", err)
	}
	observation := evidence.Observation
	tampered := evidence
	tampered.Observation.Provenance.CatalogueSHA256 = strings.Repeat("0", 64)
	tamperedPayload, err := json.Marshal(tampered)
	if err != nil {
		t.Fatalf("marshal tampered evidence: %v", err)
	}
	if _, err := decodeCognitiveCoreEvidence(tamperedPayload); err == nil {
		t.Fatal("tampered cognitive-core catalogue provenance was accepted")
	}
	if observation.Provenance.SourceCommit != cognitivecore.SourceCommitSHA || observation.Provenance.ManifestSHA256 != cognitivecore.ManifestSHA256 {
		t.Fatalf("observation lacks immutable source binding: %+v", observation.Provenance)
	}
	if observation.Provenance.TrustGrade != cognitivecore.TrustGrade || len(observation.Devices) != 4 {
		t.Fatalf("observation lacks reviewed adapted evidence: %+v", observation)
	}

	if err := orchestrator.replayCognitiveEvidence(context.Background()); err != nil {
		t.Fatalf("replay cognitive evidence: %v", err)
	}
	for index := range maxExperienceLedgerEntries + 1 {
		orchestrator.markExperienceOnce(fmt.Sprintf("eviction-fixture-%d", index))
	}
	if err := orchestrator.replayCognitiveEvidence(context.Background()); err != nil {
		t.Fatalf("second replay cognitive evidence: %v", err)
	}
	if got := orchestrator.GetStatus().PendingExperiences; got != 1 {
		t.Fatalf("replay duplicated EchoDream input: got %d pending experiences", got)
	}

	if err := orchestrator.Sleep(); err != nil {
		t.Fatalf("first Sleep: %v", err)
	}
	if err := orchestrator.Awaken(); !errors.Is(err, ErrOrchestratorTerminated) {
		t.Fatalf("same-instance restart error = %v, want ErrOrchestratorTerminated", err)
	}
	restartConfig := config
	restarted := NewUnifiedAutonomousOrchestrator(&lifecycleMockLLMProvider{}, restartConfig)
	restarted.totalCycles = 0 // Simulate a missing or stale mutable state snapshot.
	if err := restarted.Awaken(); err != nil {
		t.Fatalf("restart Awaken: %v", err)
	}
	t.Cleanup(func() {
		if restarted.GetStatus().Running {
			_ = restarted.Sleep()
		}
	})
	restartedStatus := restarted.GetStatus()
	if restartedStatus.CognitiveCore.RehydratedInputs != 1 || restartedStatus.CognitiveCore.LastCycle != 1 {
		t.Fatalf("cognitive core was not rehydrated from durable input: %+v", restartedStatus.CognitiveCore)
	}
	if restartedStatus.TotalCycles != 1 {
		t.Fatalf("ledger high-water mark did not repair total cycles: got %d", restartedStatus.TotalCycles)
	}
}

func TestCognitiveCoreFollowsCanonicalWakeRestAuthority(t *testing.T) {
	config := DefaultOrchestratorConfig()
	config.EnablePersistence = false
	config.EnableCoreSelf = false
	config.EnableEnaction = false
	config.EnableSkillLearning = false
	config.EnableDiscussionMonitoring = false
	config.EnableWisdomSynthesis = false
	config.EnableStreamOfConsciousness = false
	config.EnableEchobeats = false
	config.AutoWakeRest = false
	config.MainLoopInterval = time.Hour
	config.GoalReviewInterval = time.Hour
	config.WisdomSynthesisInterval = time.Hour
	config.StateSyncInterval = time.Hour
	config.DreamLightDuration = time.Hour
	config.CognitiveCoreInterval = time.Hour
	config.DreamDeepDuration = time.Hour
	config.DreamREMDuration = time.Hour

	orchestrator := NewUnifiedAutonomousOrchestrator(&lifecycleMockLLMProvider{}, config)
	if err := orchestrator.Awaken(); err != nil {
		t.Fatalf("Awaken: %v", err)
	}
	t.Cleanup(func() {
		if orchestrator.GetStatus().Running {
			_ = orchestrator.Sleep()
		}
	})

	orchestrator.performCognitiveCycle()
	orchestrator.performCognitiveCycle()
	if got := orchestrator.GetStatus().CognitiveCore.ObservationCount; got != 1 {
		t.Fatalf("cadence did not suppress duplicate observation: got %d", got)
	}
	if err := orchestrator.onRest(); err != nil {
		t.Fatalf("onRest: %v", err)
	}
	status := orchestrator.GetStatus()
	if status.IsAwake || status.CognitiveCore.Awake {
		t.Fatalf("rest did not close both admission boundaries: %+v", status)
	}
	if err := orchestrator.onWake(); err != nil {
		t.Fatalf("onWake: %v", err)
	}
	status = orchestrator.GetStatus()
	if !status.IsAwake || !status.CognitiveCore.Awake {
		t.Fatalf("wake did not reopen both admission boundaries: %+v", status)
	}
	orchestrator.performCognitiveCycle()
	if got := orchestrator.GetStatus().CognitiveCore.ObservationCount; got != 2 {
		t.Fatalf("wake did not admit an immediate observation: got %d", got)
	}
}
