//go:build cgo

package deeptreeecho

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/persistence"
)

func TestReflectionCausalRealPipelineChainPasses(t *testing.T) {
	planner := &scriptedActionPlanner{
		raw:   testPlanJSON(t, "briefs/causal-real.md", "# Causal validation\n"),
		route: PlanRouteEvidence{TraceID: "trace-causal-real", Provider: "real-test-provider", BackendKind: "remote_api", ModelID: "test-model", AttemptCount: 1},
	}
	pipeline, store, _ := newTestPipeline(t, EnactionLocalSandbox, planner)
	outcome, err := pipeline.RunGoal(context.Background(), testEnactionRequest())
	if err != nil {
		t.Fatalf("RunGoal: %v", err)
	}
	evaluation, err := store.Get(context.Background(), outcome.EvaluationEventID)
	if err != nil {
		t.Fatalf("load evaluation: %v", err)
	}
	if err := validateReflectionActionChain(context.Background(), store, evaluation); err != nil {
		t.Fatalf("validate real E1 action chain: %v", err)
	}
}

func TestReflectionCausalRejectsForgedIncompleteCompletion(t *testing.T) {
	_, store, _ := newTestPipeline(t, EnactionLocalSandbox, &scriptedActionPlanner{})
	plan := reflectionCausalTestPlan()
	actionID, goalID, correlationID := reflectionCausalIDs()
	observation := reflectionCausalObservation(plan, false)
	completed := appendReflectionCausalEvent(t, store, "session-test", actionID, goalID, correlationID, stageEventID(actionID, "completed"), persistence.EventTypeActionCompleted, stageEventID(actionID, "started"), persistence.EvidenceClassSensitive, observation, PlanRouteEvidence{})
	evaluation := reflectionCausalEvaluation(actionID, true)
	candidate := appendReflectionCausalEvent(t, store, "session-test", actionID, goalID, correlationID, stageEventID(actionID, "evaluation"), persistence.EventTypeEvaluationRecorded, completed.EventID, persistence.EvidenceClassSensitive, evaluation, PlanRouteEvidence{})
	if err := validateReflectionActionChain(context.Background(), store, candidate); err == nil {
		t.Fatal("forged completion/evaluation without context chain was accepted")
	}
}

func TestReflectionCausalRejectsWrongCorrelationID(t *testing.T) {
	_, store, _ := newTestPipeline(t, EnactionLocalSandbox, &scriptedActionPlanner{})
	plan := reflectionCausalTestPlan()
	actionID, goalID, _ := reflectionCausalIDs()
	wrongCorrelation := "enaction:other-goal"
	candidate := appendReflectionCausalFullChain(t, store, plan, actionID, goalID, wrongCorrelation, "session-test", "session-test", false)
	if err := validateReflectionActionChain(context.Background(), store, candidate); err == nil {
		t.Fatal("chain with wrong correlation ID was accepted")
	}
}

func TestReflectionCausalRejectsMissingStartedOrPolicy(t *testing.T) {
	for _, tc := range []struct {
		name           string
		includePolicy  bool
		includeStarted bool
	}{
		{name: "missing policy", includePolicy: false, includeStarted: true},
		{name: "missing started", includePolicy: true, includeStarted: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, store, _ := newTestPipeline(t, EnactionLocalSandbox, &scriptedActionPlanner{})
			plan := reflectionCausalTestPlan()
			actionID, goalID, correlationID := reflectionCausalIDs()
			route := reflectionCausalRoute()
			appendReflectionCausalPrefix(t, store, plan, actionID, goalID, correlationID, "session-test", tc.includePolicy, tc.includeStarted, route)
			observation := reflectionCausalObservation(plan, false)
			completed := appendReflectionCausalEvent(t, store, "session-test", actionID, goalID, correlationID, stageEventID(actionID, "completed"), persistence.EventTypeActionCompleted, stageEventID(actionID, "started"), persistence.EvidenceClassSensitive, observation, route)
			evaluation := appendReflectionCausalEvent(t, store, "session-test", actionID, goalID, correlationID, stageEventID(actionID, "evaluation"), persistence.EventTypeEvaluationRecorded, completed.EventID, persistence.EvidenceClassSensitive, reflectionCausalEvaluation(actionID, true), PlanRouteEvidence{})
			if err := validateReflectionActionChain(context.Background(), store, evaluation); err == nil {
				t.Fatalf("chain with %s was accepted", tc.name)
			}
		})
	}
}

func TestReflectionCausalAllowsRecoveredTerminalSession(t *testing.T) {
	_, store, _ := newTestPipeline(t, EnactionLocalSandbox, &scriptedActionPlanner{})
	plan := reflectionCausalTestPlan()
	actionID, goalID, correlationID := reflectionCausalIDs()
	route := reflectionCausalRoute()
	appendReflectionCausalPrefix(t, store, plan, actionID, goalID, correlationID, "session-test", true, true, route)

	observation := reflectionCausalObservation(plan, true)
	completed := appendReflectionCausalEvent(t, store, "session-recovery", actionID, goalID, correlationID, stageEventID(actionID, "completed"), persistence.EventTypeActionCompleted, stageEventID(actionID, "started"), persistence.EvidenceClassSensitive, observation, route)
	evaluation := appendReflectionCausalEvent(t, store, "session-recovery", actionID, goalID, correlationID, stageEventID(actionID, "evaluation"), persistence.EventTypeEvaluationRecorded, completed.EventID, persistence.EvidenceClassSensitive, reflectionCausalEvaluation(actionID, true), PlanRouteEvidence{})
	if err := validateReflectionActionChain(context.Background(), store, evaluation); err != nil {
		t.Fatalf("recovered terminal session was rejected: %v", err)
	}
}

func reflectionCausalTestPlan() ActionPlan {
	return ActionPlan{
		Tool:             WorkspaceCreateNote,
		Purpose:          "record causal validation evidence",
		RelativePath:     "briefs/causal-validation.md",
		Content:          "# Causal validation\n",
		Confidence:       0.9,
		Reversible:       true,
		AffectedParties:  []string{"self"},
		SuccessCriteria:  []string{"private note exists"},
		PredictedEffects: []string{"one local note"},
	}
}

func reflectionCausalRoute() PlanRouteEvidence {
	return PlanRouteEvidence{TraceID: "trace-causal", Provider: "real-test-provider", BackendKind: "remote_api", ModelID: "test-model", AttemptCount: 1}
}

func reflectionCausalIDs() (actionID, goalID, correlationID string) {
	goalID = "goal-test"
	actionID = "action:" + goalID + ":" + WorkspaceCreateNote + ":" + string(EnactionLocalSandbox) + ":iteration-1"
	return actionID, goalID, "enaction:" + goalID
}

func reflectionCausalObservation(plan ActionPlan, recovered bool) ActionObservation {
	digest := sha256Hex([]byte(plan.Content))
	return ActionObservation{
		Tool:           plan.Tool,
		RelativePath:   plan.RelativePath,
		ContentSHA256:  digest,
		ObservedSHA256: digest,
		BytesWritten:   int64(len(plan.Content)),
		Created:        !recovered,
		Recovered:      recovered,
	}
}

func reflectionCausalEvaluation(actionID string, passed bool) ActionEvaluation {
	return ActionEvaluation{
		Passed:          passed,
		Score:           boolScore(passed),
		Criteria:        []string{"tool is workspace.create_note", "target remained within private workspace", "create/reconcile reported no conflict", "durable read-back SHA-256 matches requested content"},
		EvidenceEventID: stageEventID(actionID, "evaluation"),
		Reason:          "deterministic test evidence",
	}
}

func appendReflectionCausalPrefix(t *testing.T, store *persistence.CognitiveEventStore, plan ActionPlan, actionID, goalID, correlationID, sessionID string, includePolicy, includeStarted bool, route PlanRouteEvidence) {
	t.Helper()
	appendReflectionCausalEvent(t, store, sessionID, actionID, goalID, correlationID, stageEventID(actionID, "context"), persistence.EventTypeContextRetrieved, "", persistence.EvidenceClassSensitive, reflectionContextPayload{Goal: "record causal validation evidence", TopInterests: []string{"autonomy"}, RecentContextCount: 1}, PlanRouteEvidence{})
	appendReflectionCausalEvent(t, store, sessionID, actionID, goalID, correlationID, stageEventID(actionID, "route"), persistence.EventTypeProviderRouted, stageEventID(actionID, "context"), persistence.EvidenceClassOperational, route, route)
	appendReflectionCausalEvent(t, store, sessionID, actionID, goalID, correlationID, stageEventID(actionID, "plan"), persistence.EventTypePlanProposed, stageEventID(actionID, "route"), persistence.EvidenceClassRestricted, plan, route)
	request := reflectionRequestedPayload{Tool: plan.Tool, Purpose: plan.Purpose, RelativePath: plan.RelativePath, RequestedSHA256: sha256Hex([]byte(plan.Content)), Reversible: plan.Reversible}
	appendReflectionCausalEvent(t, store, sessionID, actionID, goalID, correlationID, stageEventID(actionID, "requested"), persistence.EventTypeActionRequested, stageEventID(actionID, "plan"), persistence.EvidenceClassSensitive, request, route)
	if includePolicy {
		decision := ActionPolicyDecision{PolicyVersion: EnactionPolicyVersion, Allowed: true, Code: "allowed", Reason: "validated create-only private workspace action within all E1 bounds", Mode: EnactionLocalSandbox, DecidedAt: time.Now().UTC()}
		appendReflectionCausalEvent(t, store, sessionID, actionID, goalID, correlationID, stageEventID(actionID, "policy"), persistence.EventTypePolicyDecision, stageEventID(actionID, "requested"), persistence.EvidenceClassSensitive, decision, PlanRouteEvidence{})
	}
	if includeStarted {
		started := reflectionStartedPayload{Tool: plan.Tool, RelativePath: plan.RelativePath, RequestedSHA256: sha256Hex([]byte(plan.Content))}
		appendReflectionCausalEvent(t, store, sessionID, actionID, goalID, correlationID, stageEventID(actionID, "started"), persistence.EventTypeActionStarted, stageEventID(actionID, "policy"), persistence.EvidenceClassSensitive, started, route)
	}
}

func appendReflectionCausalFullChain(t *testing.T, store *persistence.CognitiveEventStore, plan ActionPlan, actionID, goalID, correlationID, prefixSession, terminalSession string, recovered bool) persistence.StoredCognitiveEvent {
	t.Helper()
	route := reflectionCausalRoute()
	appendReflectionCausalPrefix(t, store, plan, actionID, goalID, correlationID, prefixSession, true, true, route)
	completed := appendReflectionCausalEvent(t, store, terminalSession, actionID, goalID, correlationID, stageEventID(actionID, "completed"), persistence.EventTypeActionCompleted, stageEventID(actionID, "started"), persistence.EvidenceClassSensitive, reflectionCausalObservation(plan, recovered), route)
	return appendReflectionCausalEvent(t, store, terminalSession, actionID, goalID, correlationID, stageEventID(actionID, "evaluation"), persistence.EventTypeEvaluationRecorded, completed.EventID, persistence.EvidenceClassSensitive, reflectionCausalEvaluation(actionID, true), PlanRouteEvidence{})
}

func appendReflectionCausalEvent(t *testing.T, store *persistence.CognitiveEventStore, sessionID, actionID, goalID, correlationID, eventID, eventType, causationID, evidenceClass string, payload any, route PlanRouteEvidence) persistence.StoredCognitiveEvent {
	t.Helper()
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal %s: %v", eventType, err)
	}
	event := persistence.NewCognitiveEvent(eventID, eventType, "echo-test", sessionID, EnactionPolicyVersion, payloadJSON)
	event.CorrelationID = correlationID
	event.CausationID = causationID
	event.GoalID = goalID
	event.ActionID = actionID
	event.IdempotencyKey = eventID
	event.EvidenceClass = evidenceClass
	event.Provider = sanitizeIdentifier(route.Provider)
	event.BackendKind = sanitizeIdentifier(route.BackendKind)
	event.ModelID = sanitizeIdentifier(route.ModelID)
	event.Degraded = route.Degraded
	if _, err := store.Append(context.Background(), event); err != nil {
		t.Fatalf("append %s: %v", eventType, err)
	}
	stored, err := store.Get(context.Background(), eventID)
	if err != nil {
		t.Fatalf("load %s: %v", eventType, err)
	}
	return stored
}
