package deeptreeecho

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/cognitivecore"
	"github.com/cogpy/echo9llama/core/persistence"
	"github.com/cogpy/echo9llama/core/tools"
)

func practiceProjectionRig(t *testing.T, planner ActionPlanner) (*UnifiedAutonomousOrchestrator, *EnactionPipeline) {
	t.Helper()
	root := t.TempDir()
	store, err := persistence.NewCognitiveEventStore(filepath.Join(root, "events.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	policy := DefaultActionPolicyConfig()
	policy.Mode = EnactionLocalSandbox
	pipeline, err := NewEnactionPipeline(store, tools.WorkspaceNoteTool{Root: root, MaxBytes: 4096, Timeout: 5 * time.Second}, NewActionPolicy(policy), planner, "echo-test", "session-test")
	if err != nil {
		t.Fatal(err)
	}
	rig := &UnifiedAutonomousOrchestrator{ctx: context.Background(), identityID: "echo-test", sessionID: "session-test", isAwake: true, eventStore: store, enactionPipeline: pipeline, echobeatsScheduler: NewEchobeatsScheduler(nil), skillLearning: NewSkillLearningSystem(nil), config: OrchestratorConfig{EnableCognitiveCore: true, EnableEnaction: true, EnableEvaluatedPractice: true, EnactionMode: EnactionLocalSandbox, EnablePersistence: true, StateDirectory: root}}
	rig.initializeOutcomeLearning()
	return rig, pipeline
}

func admittedPracticeRequest(t *testing.T, rig *UnifiedAutonomousOrchestrator) EnactionRequest {
	t.Helper()
	if err := rig.queueEvaluatedPractice(); err != nil {
		t.Fatal(err)
	}
	goal := rig.echobeatsScheduler.GetActiveGoal()
	if goal == nil {
		t.Fatal("practice goal not admitted")
	}
	task := rig.practiceTasks[goal.ID]
	return EnactionRequest{IdentityID: rig.identityID, SessionID: rig.sessionID, GoalID: goal.ID, Goal: goal.Description, Practice: &task}
}

func TestOutcomeProjectionAdmissionNegativeReplayAndCheckpoint(t *testing.T) {
	planner := &scriptedActionPlanner{route: PlanRouteEvidence{TraceID: "fixture", Provider: "explicit-unit-fixture", BackendKind: "remote_api", ModelID: "unit-test-fixture"}}
	rig, pipeline := practiceProjectionRig(t, planner)
	request := admittedPracticeRequest(t, rig)
	candidate := map[string]any{"task_id": request.Practice.TaskID, "sum": request.Practice.A + request.Practice.B, "product": request.Practice.A * request.Practice.B}
	bytes, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	planner.raw = testPlanJSON(t, "briefs/projection-trial-1.md", string(bytes))
	if _, err := pipeline.RunGoal(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if err := rig.replayCognitiveEvidence(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := rig.replayOutcomeLearning(context.Background()); err != nil {
		t.Fatal(err)
	}
	first := rig.outcomeLearner.Snapshot()
	if first.Counts.Training+first.Counts.HeldOut != 1 {
		t.Fatalf("verified task did not produce one explicit outcome: %#v", first.Counts)
	}
	skill, err := rig.skillLearning.GetSkillByID(stableSkillID(evaluatedPracticeSkill))
	if err != nil {
		t.Fatal(err)
	}
	before := skill.Proficiency
	// A second, independently checked wrong candidate is a real negative label.
	request = admittedPracticeRequest(t, rig)
	planner.raw = testPlanJSON(t, "briefs/projection-trial-2.md", `{"task_id":"`+request.Practice.TaskID+`","sum":0,"product":0}`)
	if _, err := pipeline.RunGoal(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	events, err := rig.eventStore.Query(context.Background(), persistence.CognitiveEventQuery{ActionID: stageActionIDForPractice(rig, request.GoalID), Limit: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if err := rig.projectEnactionEvents(events); err != nil {
		t.Fatal(err)
	}
	if err := rig.replayOutcomeLearning(context.Background()); err != nil {
		t.Fatal(err)
	}
	after, err := rig.skillLearning.GetSkillByID(stableSkillID(evaluatedPracticeSkill))
	if err != nil {
		t.Fatal(err)
	}
	if after.Proficiency >= before {
		t.Fatalf("negative correctness did not lower advisory competence: before=%g after=%g", before, after.Proficiency)
	}
	// An outage retires a third trial but supplies no invented negative label.
	request = admittedPracticeRequest(t, rig)
	planner.err = context.DeadlineExceeded
	if _, err := pipeline.RunGoal(context.Background(), request); err == nil {
		t.Fatal("outage falsely succeeded")
	}
	unknownEvents, err := rig.eventStore.Query(context.Background(), persistence.CognitiveEventQuery{ActionID: stageActionIDForPractice(rig, request.GoalID), Limit: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if err := rig.projectEnactionEvents(unknownEvents); err != nil {
		t.Fatal(err)
	}
	if err := rig.replayOutcomeLearning(context.Background()); err != nil {
		t.Fatal(err)
	}
	snapshot := rig.outcomeLearner.Snapshot()
	if snapshot.Counts.Unknown != 1 || snapshot.Counts.Training+snapshot.Counts.HeldOut != 2 {
		t.Fatalf("known/negative/unknown evidence changed labels: %#v", snapshot.Counts)
	}
	if err := rig.replayOutcomeLearning(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(snapshot, rig.outcomeLearner.Snapshot()) {
		t.Fatal("duplicate projection retrained or advanced recurrence")
	}
	if err := rig.saveOutcomeCheckpoint(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(rig.config.StateDirectory, "auxiliary_learner_checkpoint.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("private checkpoint mode: %v", info.Mode())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var checkpoint outcomeCheckpointExport
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		t.Fatal(err)
	}
	clone := cognitivecore.NewOutcomeLearner()
	if err := clone.Restore(checkpoint.Learner); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(snapshot, clone.Snapshot()) {
		t.Fatal("checkpoint numerical state diverged")
	}
	// Corrupt cache cannot become authority: startup reconstruction ignores it.
	if err := os.WriteFile(path, []byte("corrupt\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	restarted := &UnifiedAutonomousOrchestrator{ctx: context.Background(), identityID: rig.identityID, eventStore: rig.eventStore, echobeatsScheduler: NewEchobeatsScheduler(nil), skillLearning: NewSkillLearningSystem(nil), config: rig.config}
	restarted.initializeOutcomeLearning()
	if err := restarted.replayCognitiveEvidence(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := restarted.replayOutcomeLearning(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(snapshot, restarted.outcomeLearner.Snapshot()) {
		t.Fatal("canonical full replay changed numerical state")
	}
}

func stageActionIDForPractice(rig *UnifiedAutonomousOrchestrator, goalID string) string {
	return "action:" + goalID + ":workspace.create_note:local-sandbox:iteration-1"
}

func TestOutcomeProjectionUnknownAndOptIn(t *testing.T) {
	planner := &scriptedActionPlanner{err: context.DeadlineExceeded, route: PlanRouteEvidence{TraceID: "fixture-outage", Provider: "explicit-unit-fixture", BackendKind: "remote_api", ModelID: "unit-test-fixture"}}
	rig, pipeline := practiceProjectionRig(t, planner)
	rig.config.EnableEvaluatedPractice = false
	if err := rig.queueEvaluatedPractice(); err != nil {
		t.Fatal(err)
	}
	if len(rig.echobeatsScheduler.GetGoalQueue()) != 0 {
		t.Fatal("practice became active without opt-in")
	}
	rig.config.EnableEvaluatedPractice = true
	request := admittedPracticeRequest(t, rig)
	if _, err := pipeline.RunGoal(context.Background(), request); err == nil {
		t.Fatal("outage falsely succeeded")
	}
	if err := rig.replayCognitiveEvidence(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := rig.replayOutcomeLearning(context.Background()); err != nil {
		t.Fatal(err)
	}
	snapshot := rig.outcomeLearner.Snapshot()
	if snapshot.Counts.Unknown != 1 || snapshot.Counts.Training != 0 || snapshot.Counts.HeldOut != 0 || snapshot.Optimizer.Steps != 0 {
		t.Fatalf("outage invented a target: %#v", snapshot.Counts)
	}
	if !rig.practiceTerminal[request.GoalID] {
		t.Fatal("unknown trial was not closed; practice would stall")
	}
	goal := reviewGoal(t, rig, request.GoalID)
	if goal.Status != GoalAbandoned || goal.Progress != 0 || rig.echobeatsScheduler.goalsCompleted != 0 {
		t.Fatalf("unknown trial invented arithmetic achievement: %#v", goal)
	}
}
