//go:build cgo

package deeptreeecho

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/persistence"
	"github.com/cogpy/echo9llama/core/tools"
)

type scriptedActionPlanner struct {
	raw   string
	route PlanRouteEvidence
	err   error
	calls int
}

func (planner *scriptedActionPlanner) GeneratePlan(context.Context, EnactionRequest) (string, PlanRouteEvidence, error) {
	planner.calls++
	return planner.raw, planner.route, planner.err
}

func testPlanJSON(t *testing.T, path, content string) string {
	t.Helper()
	payload, err := json.Marshal(ActionPlan{
		Tool: WorkspaceCreateNote, Purpose: "capture evidence-aware questions",
		RelativePath: path, Content: content, Confidence: 0.9, Reversible: true,
		AffectedParties:  []string{"self"},
		SuccessCriteria:  []string{"private note created", "read-back hash matches"},
		PredictedEffects: []string{"one local note"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}

func newTestPipeline(t *testing.T, mode EnactionMode, planner *scriptedActionPlanner) (*EnactionPipeline, *persistence.CognitiveEventStore, string) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "workspace")
	store, err := persistence.NewCognitiveEventStore(filepath.Join(t.TempDir(), "events.db"))
	if err != nil {
		t.Fatalf("NewCognitiveEventStore: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	config := DefaultActionPolicyConfig()
	config.Mode = mode
	policy := NewActionPolicy(config)
	pipeline, err := NewEnactionPipeline(store, tools.WorkspaceNoteTool{Root: root, MaxBytes: int64(config.MaxArtifactBytes), Timeout: time.Second}, policy, planner, "echo-test", "session-test")
	if err != nil {
		t.Fatalf("NewEnactionPipeline: %v", err)
	}
	return pipeline, store, root
}

func testEnactionRequest() EnactionRequest {
	return EnactionRequest{
		IdentityID: "echo-test", SessionID: "session-test", GoalID: "goal-test",
		Goal:         "Clarify what evidence should guide the next autonomy iteration.",
		TopInterests: []string{"wisdom", "autonomy"}, RecentContext: []string{"event ledger is healthy"},
	}
}

func TestEnactionObserveModeRecordsProposalWithoutToolEffect(t *testing.T) {
	planner := &scriptedActionPlanner{
		raw:   testPlanJSON(t, "briefs/observe.md", "# Observe\n"),
		route: PlanRouteEvidence{TraceID: "trace-observe", Provider: "real-test-provider", BackendKind: "remote_api", ModelID: "test-model"},
	}
	pipeline, store, root := newTestPipeline(t, EnactionObserve, planner)
	outcome, err := pipeline.RunGoal(context.Background(), testEnactionRequest())
	if err != nil {
		t.Fatalf("RunGoal: %v", err)
	}
	if outcome.Allowed || outcome.Executed || outcome.Verified {
		t.Fatalf("observe mode crossed tool boundary: %#v", outcome)
	}
	if _, err := os.Stat(filepath.Join(root, "briefs", "observe.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("observe mode created a file: %v", err)
	}
	started, err := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: outcome.ActionID, EventType: persistence.EventTypeActionStarted})
	if err != nil || started != 0 {
		t.Fatalf("unexpected action.started count=%d err=%v", started, err)
	}
	dreams, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: outcome.ActionID, EventType: persistence.EventTypeDreamExperience})
	goals, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: outcome.ActionID, EventType: persistence.EventTypeGoalProgressed})
	skills, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: outcome.ActionID, EventType: persistence.EventTypeSkillEvidence})
	if dreams != 1 || goals != 0 || skills != 0 {
		t.Fatalf("denial learning projection mismatch: dreams=%d goals=%d skills=%d", dreams, goals, skills)
	}
}

func TestEnactionVerifiedActionAndReplayAreExactlyOnce(t *testing.T) {
	planner := &scriptedActionPlanner{
		raw:   testPlanJSON(t, "briefs/verified.md", "# Verified\n\nWhat evidence would falsify this hypothesis?\n"),
		route: PlanRouteEvidence{TraceID: "trace-verified", Provider: "real-test-provider", BackendKind: "remote_api", ModelID: "test-model", AttemptCount: 1},
	}
	pipeline, store, root := newTestPipeline(t, EnactionLocalSandbox, planner)
	request := testEnactionRequest()
	first, err := pipeline.RunGoal(context.Background(), request)
	if err != nil {
		t.Fatalf("first RunGoal: %v", err)
	}
	if !first.Allowed || !first.Executed || !first.Verified || first.Score != 1 {
		t.Fatalf("verified action did not complete: %#v", first)
	}
	info, err := os.Stat(filepath.Join(root, "briefs", "verified.md"))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("private artifact mode/exists: info=%v err=%v", info, err)
	}

	second, err := pipeline.RunGoal(context.Background(), request)
	if err != nil {
		t.Fatalf("replay RunGoal: %v", err)
	}
	if !second.Recovered || !second.Verified || planner.calls != 1 {
		t.Fatalf("replay was not event-driven/idempotent: outcome=%#v calls=%d", second, planner.calls)
	}
	completed, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: first.ActionID, EventType: persistence.EventTypeActionCompleted})
	evaluated, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: first.ActionID, EventType: persistence.EventTypeEvaluationRecorded})
	goals, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: first.ActionID, EventType: persistence.EventTypeGoalProgressed})
	skills, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: first.ActionID, EventType: persistence.EventTypeSkillEvidence})
	dreams, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: first.ActionID, EventType: persistence.EventTypeDreamExperience})
	if completed != 1 || evaluated != 1 || goals != 1 || skills != 1 || dreams != 1 {
		t.Fatalf("causal projection counts: completed=%d evaluated=%d goals=%d skills=%d dreams=%d", completed, evaluated, goals, skills, dreams)
	}
}

func TestEnactionCrashRecoveryReconcilesMatchingArtifact(t *testing.T) {
	plan := ActionPlan{
		Tool: WorkspaceCreateNote, Purpose: "recovery", RelativePath: "briefs/recover.md", Content: "# Recover\n",
		Confidence: 0.9, Reversible: true, AffectedParties: []string{"self"}, SuccessCriteria: []string{"verified"}, PredictedEffects: []string{"note"},
	}
	planner := &scriptedActionPlanner{raw: testPlanJSON(t, plan.RelativePath, plan.Content), route: PlanRouteEvidence{TraceID: "unused", Provider: "unused", BackendKind: "remote_api"}}
	pipeline, _, root := newTestPipeline(t, EnactionLocalSandbox, planner)
	request := testEnactionRequest()
	actionID := "action:" + request.GoalID + ":workspace.create_note:" + string(EnactionLocalSandbox) + ":iteration-1"
	correlationID := "enaction:" + request.GoalID
	route := PlanRouteEvidence{TraceID: "trace-recover", Provider: "real-test-provider", BackendKind: "remote_api", ModelID: "test-model"}

	if _, err := pipeline.tool.Create(context.Background(), tools.WorkspaceNoteRequest{Path: plan.RelativePath, Content: []byte(plan.Content)}); err != nil {
		t.Fatalf("seed artifact: %v", err)
	}
	seedReplayChain(t, pipeline, plan, route, request, actionID, correlationID)

	outcome, err := pipeline.RunGoal(context.Background(), request)
	if err != nil {
		t.Fatalf("RunGoal recovery: %v", err)
	}
	if !outcome.Recovered || !outcome.Verified || planner.calls != 0 {
		t.Fatalf("matching recovery failed: %#v calls=%d", outcome, planner.calls)
	}
	content, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(plan.RelativePath)))
	if string(content) != plan.Content {
		t.Fatalf("recovery rewrote artifact: %q", content)
	}
}

func TestEnactionCrashRecoveryRetriesMissingEffectOnce(t *testing.T) {
	plan := ActionPlan{
		Tool: WorkspaceCreateNote, Purpose: "recovery", RelativePath: "briefs/retry-missing.md", Content: "# Retry Missing Effect\n",
		Confidence: 0.9, Reversible: true, AffectedParties: []string{"self"}, SuccessCriteria: []string{"verified"}, PredictedEffects: []string{"note"},
	}
	planner := &scriptedActionPlanner{raw: testPlanJSON(t, plan.RelativePath, plan.Content), route: PlanRouteEvidence{Provider: "unused", BackendKind: "remote_api"}}
	pipeline, store, root := newTestPipeline(t, EnactionLocalSandbox, planner)
	request := testEnactionRequest()
	actionID := "action:" + request.GoalID + ":workspace.create_note:" + string(EnactionLocalSandbox) + ":iteration-1"
	correlationID := "enaction:" + request.GoalID
	route := PlanRouteEvidence{TraceID: "trace-retry", Provider: "real-test-provider", BackendKind: "remote_api", ModelID: "test-model"}
	seedReplayChain(t, pipeline, plan, route, request, actionID, correlationID)

	outcome, err := pipeline.RunGoal(context.Background(), request)
	if err != nil {
		t.Fatalf("RunGoal missing-effect recovery: %v", err)
	}
	if outcome.Recovered || !outcome.Verified || planner.calls != 0 {
		t.Fatalf("missing effect was not recovered exactly once: %#v calls=%d", outcome, planner.calls)
	}
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(plan.RelativePath)))
	if err != nil || string(content) != plan.Content {
		t.Fatalf("recovery artifact mismatch: content=%q err=%v", content, err)
	}
	completed, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: actionID, EventType: persistence.EventTypeActionCompleted})
	if completed != 1 {
		t.Fatalf("recovery completion count=%d, want 1", completed)
	}
}

func TestEnactionCrashRecoveryBlocksHashConflict(t *testing.T) {
	plan := ActionPlan{
		Tool: WorkspaceCreateNote, Purpose: "recovery", RelativePath: "briefs/conflict.md", Content: "# Intended\n",
		Confidence: 0.9, Reversible: true, AffectedParties: []string{"self"}, SuccessCriteria: []string{"verified"}, PredictedEffects: []string{"note"},
	}
	planner := &scriptedActionPlanner{raw: testPlanJSON(t, plan.RelativePath, plan.Content), route: PlanRouteEvidence{Provider: "unused", BackendKind: "remote_api"}}
	pipeline, store, root := newTestPipeline(t, EnactionLocalSandbox, planner)
	request := testEnactionRequest()
	actionID := "action:" + request.GoalID + ":workspace.create_note:" + string(EnactionLocalSandbox) + ":iteration-1"
	correlationID := "enaction:" + request.GoalID
	route := PlanRouteEvidence{TraceID: "trace-conflict", Provider: "real-test-provider", BackendKind: "remote_api", ModelID: "test-model"}

	if _, err := pipeline.tool.Create(context.Background(), tools.WorkspaceNoteRequest{Path: plan.RelativePath, Content: []byte("# Different\n")}); err != nil {
		t.Fatalf("seed conflicting artifact: %v", err)
	}
	seedReplayChain(t, pipeline, plan, route, request, actionID, correlationID)

	outcome, err := pipeline.RunGoal(context.Background(), request)
	if err == nil || outcome.Verified {
		t.Fatalf("hash conflict did not fail closed: outcome=%#v err=%v", outcome, err)
	}
	failed, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: actionID, EventType: persistence.EventTypeActionFailed})
	if failed != 1 {
		t.Fatalf("expected one durable action failure, got %d", failed)
	}
	content, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(plan.RelativePath)))
	if string(content) != "# Different\n" {
		t.Fatalf("conflicting artifact was overwritten: %q", content)
	}
}

func TestEnactionRejectsDegradedPlannerOutput(t *testing.T) {
	planner := &scriptedActionPlanner{
		raw:   testPlanJSON(t, "briefs/degraded.md", "# Degraded\n"),
		route: PlanRouteEvidence{TraceID: "trace-degraded", Provider: "simple-fallback", BackendKind: string("fallback"), Degraded: true},
	}
	pipeline, store, root := newTestPipeline(t, EnactionLocalSandbox, planner)
	outcome, err := pipeline.RunGoal(context.Background(), testEnactionRequest())
	if err == nil || outcome.Executed {
		t.Fatalf("degraded plan was enacted: outcome=%#v err=%v", outcome, err)
	}
	plans, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: outcome.ActionID, EventType: persistence.EventTypePlanProposed})
	if plans != 0 {
		t.Fatalf("degraded output was persisted as an actionable plan")
	}
	if _, err := os.Stat(filepath.Join(root, "briefs", "degraded.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("degraded output created a file: %v", err)
	}
	dreams, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: outcome.ActionID, EventType: persistence.EventTypeDreamExperience})
	goals, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: outcome.ActionID, EventType: persistence.EventTypeGoalProgressed})
	skills, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: outcome.ActionID, EventType: persistence.EventTypeSkillEvidence})
	if dreams != 1 || goals != 0 || skills != 0 {
		t.Fatalf("degraded failure projection mismatch: dreams=%d goals=%d skills=%d", dreams, goals, skills)
	}
}

func TestEnactionRestartReusesCommittedContextAcrossSessions(t *testing.T) {
	planner := &scriptedActionPlanner{
		raw:   testPlanJSON(t, "briefs/restart-context-1.md", "# Restart Context\n"),
		route: PlanRouteEvidence{TraceID: "trace-context-restart", Provider: "real-test-provider", BackendKind: "remote_api", ModelID: "test-model"},
	}
	first, store, root := newTestPipeline(t, EnactionLocalSandbox, planner)
	request := testEnactionRequest()
	actionID := "action:" + request.GoalID + ":workspace.create_note:" + string(EnactionLocalSandbox) + ":iteration-1"
	contextPayload := map[string]any{
		"goal": request.Goal, "top_interests": append([]string(nil), request.TopInterests...),
		"recent_context_count": len(request.RecentContext),
	}
	if _, err := first.appendEvent(context.Background(), stageEventID(actionID, "context"), persistence.EventTypeContextRetrieved, "enaction:"+request.GoalID, "", request.GoalID, actionID, persistence.EvidenceClassSensitive, contextPayload, PlanRouteEvidence{}); err != nil {
		t.Fatalf("seed context stage: %v", err)
	}

	config := DefaultActionPolicyConfig()
	config.Mode = EnactionLocalSandbox
	restarted, err := NewEnactionPipeline(store, tools.WorkspaceNoteTool{Root: root, MaxBytes: int64(config.MaxArtifactBytes), Timeout: time.Second}, NewActionPolicy(config), planner, "echo-test", "session-restart")
	if err != nil {
		t.Fatalf("NewEnactionPipeline restart: %v", err)
	}
	request.SessionID = "session-restart"
	outcome, err := restarted.RunGoal(context.Background(), request)
	if err != nil || !outcome.Verified {
		t.Fatalf("restart after context did not resume: outcome=%#v err=%v", outcome, err)
	}
	contexts, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: actionID, EventType: persistence.EventTypeContextRetrieved})
	if contexts != 1 {
		t.Fatalf("committed context was duplicated: %d", contexts)
	}
	stored, err := store.Get(context.Background(), stageEventID(actionID, "context"))
	if err != nil || stored.SessionID != "session-test" {
		t.Fatalf("original context envelope was not reused: event=%#v err=%v", stored, err)
	}
}

func TestEnactionRestartCompletesMissingEvidenceProjections(t *testing.T) {
	plan := ActionPlan{
		Tool: WorkspaceCreateNote, Purpose: "projection recovery", RelativePath: "briefs/projection-recovery.md", Content: "# Projection Recovery\n",
		Confidence: 0.9, Reversible: true, AffectedParties: []string{"self"}, SuccessCriteria: []string{"verified"}, PredictedEffects: []string{"note"},
	}
	planner := &scriptedActionPlanner{raw: testPlanJSON(t, plan.RelativePath, plan.Content), route: PlanRouteEvidence{Provider: "unused", BackendKind: "remote_api"}}
	pipeline, store, _ := newTestPipeline(t, EnactionLocalSandbox, planner)
	request := testEnactionRequest()
	actionID := "action:" + request.GoalID + ":workspace.create_note:" + string(EnactionLocalSandbox) + ":iteration-1"
	correlationID := "enaction:" + request.GoalID
	route := PlanRouteEvidence{TraceID: "trace-projection-restart", Provider: "real-test-provider", BackendKind: "remote_api", ModelID: "test-model"}
	seedReplayChain(t, pipeline, plan, route, request, actionID, correlationID)
	observed, err := pipeline.tool.Create(context.Background(), tools.WorkspaceNoteRequest{Path: plan.RelativePath, Content: []byte(plan.Content)})
	if err != nil {
		t.Fatalf("seed completed effect: %v", err)
	}
	observation := observationFromTool(plan, observed, false, nil)
	if err := pipeline.appendCompletion(context.Background(), actionID, correlationID, request.GoalID, observation, route); err != nil {
		t.Fatalf("seed completion: %v", err)
	}
	if _, err := pipeline.recordEvaluation(context.Background(), plan, observation, actionID, correlationID, request.GoalID); err != nil {
		t.Fatalf("seed evaluation: %v", err)
	}

	outcome, err := pipeline.RunGoal(context.Background(), request)
	if err != nil || !outcome.Verified || !outcome.Recovered {
		t.Fatalf("projection recovery failed: outcome=%#v err=%v", outcome, err)
	}
	goals, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: actionID, EventType: persistence.EventTypeGoalProgressed})
	skills, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: actionID, EventType: persistence.EventTypeSkillEvidence})
	dreams, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: actionID, EventType: persistence.EventTypeDreamExperience})
	if goals != 1 || skills != 1 || dreams != 1 {
		t.Fatalf("missing projections were not completed exactly once: goals=%d skills=%d dreams=%d", goals, skills, dreams)
	}
}

func TestEnactionPausePreventsPlanningAndToolCalls(t *testing.T) {
	planner := &scriptedActionPlanner{
		raw:   testPlanJSON(t, "briefs/after-wake.md", "# After Wake\n"),
		route: PlanRouteEvidence{TraceID: "trace-pause", Provider: "real-test-provider", BackendKind: "remote_api", ModelID: "test-model"},
	}
	pipeline, _, root := newTestPipeline(t, EnactionLocalSandbox, planner)
	pipeline.Pause()
	if !pipeline.IsPaused() {
		t.Fatal("pipeline did not report paused state")
	}
	if _, err := pipeline.RunGoal(context.Background(), testEnactionRequest()); err == nil {
		t.Fatal("paused pipeline accepted a goal")
	}
	if planner.calls != 0 {
		t.Fatalf("paused pipeline called planner %d times", planner.calls)
	}
	if _, err := os.Stat(filepath.Join(root, "briefs", "after-wake.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("paused pipeline created a file: %v", err)
	}
	pipeline.Resume()
	if _, err := pipeline.RunGoal(context.Background(), testEnactionRequest()); err != nil {
		t.Fatalf("resumed pipeline failed: %v", err)
	}
	if planner.calls != 1 {
		t.Fatalf("resumed pipeline planner calls=%d", planner.calls)
	}
}

func seedReplayChain(t *testing.T, pipeline *EnactionPipeline, plan ActionPlan, route PlanRouteEvidence, request EnactionRequest, actionID, correlationID string) {
	t.Helper()
	ctx := context.Background()
	mustAppend := func(eventID, eventType, causation, class string, payload any) {
		t.Helper()
		if _, err := pipeline.appendEvent(ctx, eventID, eventType, correlationID, causation, request.GoalID, actionID, class, payload, route); err != nil {
			t.Fatalf("append %s: %v", eventType, err)
		}
	}
	mustAppend(stageEventID(actionID, "route"), persistence.EventTypeProviderRouted, "", persistence.EvidenceClassOperational, route)
	mustAppend(stageEventID(actionID, "plan"), persistence.EventTypePlanProposed, stageEventID(actionID, "route"), persistence.EvidenceClassRestricted, plan)
	mustAppend(stageEventID(actionID, "requested"), persistence.EventTypeActionRequested, stageEventID(actionID, "plan"), persistence.EvidenceClassSensitive, map[string]any{"tool": plan.Tool, "relative_path": plan.RelativePath, "requested_sha256": sha256Hex([]byte(plan.Content))})
	mustAppend(stageEventID(actionID, "policy"), persistence.EventTypePolicyDecision, stageEventID(actionID, "requested"), persistence.EvidenceClassSensitive, ActionPolicyDecision{PolicyVersion: EnactionPolicyVersion, Allowed: true, Code: "allowed", Reason: "seeded allowed decision", Mode: EnactionLocalSandbox, DecidedAt: time.Now().UTC()})
	mustAppend(stageEventID(actionID, "started"), persistence.EventTypeActionStarted, stageEventID(actionID, "policy"), persistence.EvidenceClassSensitive, map[string]any{"tool": plan.Tool, "relative_path": plan.RelativePath, "requested_sha256": sha256Hex([]byte(plan.Content))})
}
