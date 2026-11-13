package deeptreeecho

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/cogpy/echo9llama/core/persistence"
)

const (
	evaluatedPracticeVersion = "arithmetic-artifact-v1"
	evaluatedPracticeSkill   = "Arithmetic Artifact Construction"
)

// errRejectedPracticeEvidence denotes permanent semantic rejection, never a
// transient ledger read failure. Replay may quarantine only this class.
var errRejectedPracticeEvidence = errors.New("rejected practice evidence")

// ErrRejectedPracticeEvidence exposes the same sentinel to external callers.
var ErrRejectedPracticeEvidence = errRejectedPracticeEvidence

func rejectPractice(format string, args ...any) error {
	return fmt.Errorf("%w: %s", errRejectedPracticeEvidence, fmt.Sprintf(format, args...))
}

type practiceGoalEvidence struct {
	Description string               `json:"description"`
	Priority    float64              `json:"priority"`
	Index       int64                `json:"practice_index"`
	Task        ArtifactPracticeTask `json:"practice_task"`
}

func generatedPracticeTask(identity string, index int64) ArtifactPracticeTask {
	// Reduce before multiplication so even a malicious large index cannot overflow.
	return ArtifactPracticeTask{TaskID: fmt.Sprintf("practice-%s-%d", shortHash(identity), index), Version: evaluatedPracticeVersion, A: ((index%900001)*7919)%900001 + 101, B: ((index%900001)*3571)%900001 + 103}
}

func practiceDescription(task ArtifactPracticeTask) string {
	return fmt.Sprintf("Construct and verify arithmetic JSON for task %s: sum and product of %d and %d. Include task_id, sum, product only.", task.TaskID, task.A, task.B)
}

func (uao *UnifiedAutonomousOrchestrator) practiceEvidence(ctx context.Context, id string) (persistence.StoredCognitiveEvent, error) {
	if id == "" {
		return persistence.StoredCognitiveEvent{}, rejectPractice("missing source ID")
	}
	stored, err := uao.eventStore.Get(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return stored, rejectPractice("missing durable source %s", id)
	}
	return stored, err
}

// validatePracticeAdmission verifies the admission's own immutable envelope and
// deterministic task. The current session is deliberately not an authority:
// admissions and action contexts retain their original (possibly earlier) sessions.
func (uao *UnifiedAutonomousOrchestrator) validatePracticeAdmission(ctx context.Context, event persistence.StoredCognitiveEvent) (practiceGoalEvidence, error) {
	var payload practiceGoalEvidence
	if uao.eventStore == nil {
		return payload, fmt.Errorf("practice event store unavailable")
	}
	if err := event.Validate(); err != nil {
		return payload, rejectPractice("invalid admission envelope: %v", err)
	}
	if event.Sequence <= 0 || event.ContentSHA256 == "" || event.EventType != persistence.EventTypeGoalCreated || event.IdentityID != uao.identityID || event.PolicyVersion != evaluatedPracticeVersion || event.EventID != event.GoalID || event.IdempotencyKey != event.EventID || event.ActionID != "" || event.CausationID != "" || event.CorrelationID != "" || event.Degraded || event.Provider != "" || event.BackendKind != "" || event.ModelID != "" || event.RouteReason != "" {
		return payload, rejectPractice("invalid practice admission binding %s", event.EventID)
	}
	stored, err := uao.practiceEvidence(ctx, event.EventID)
	if err != nil {
		return payload, err
	}
	if stored.Sequence != event.Sequence || stored.ContentSHA256 != event.ContentSHA256 {
		return payload, rejectPractice("admission is not canonical durable evidence")
	}
	fields, err := strictPracticeObject(event.PayloadJSON)
	if err != nil || len(fields) != 4 || fields["description"] == nil || fields["priority"] == nil || fields["practice_index"] == nil || fields["practice_task"] == nil {
		return payload, rejectPractice("invalid practice admission fields")
	}
	for _, value := range fields {
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return payload, rejectPractice("null practice admission field")
		}
	}
	if err := json.Unmarshal(event.PayloadJSON, &payload); err != nil {
		return payload, rejectPractice("decode practice admission: %v", err)
	}
	task, err := decodeArtifactPracticeTask(fields["practice_task"])
	if err != nil || task == nil || payload.Index <= 0 || payload.Priority != 1 || *task != payload.Task || payload.Task != generatedPracticeTask(event.IdentityID, payload.Index) || event.GoalID != "practice-goal:"+payload.Task.TaskID || payload.Description != practiceDescription(payload.Task) {
		return payload, rejectPractice("invalid deterministic practice task %s", event.EventID)
	}
	return payload, nil
}

// ValidateTaskAdmission requires a durable admission strictly before the canonical
// action context. The bounded runnable-task map is not learning-label authority.
func (uao *UnifiedAutonomousOrchestrator) ValidateTaskAdmission(ctx context.Context, task ArtifactPracticeTask, actionContext persistence.StoredCognitiveEvent) error {
	if uao.eventStore == nil {
		return fmt.Errorf("practice event store unavailable")
	}
	if err := task.Validate(); err != nil {
		return rejectPractice("invalid task: %v", err)
	}
	goalID := "practice-goal:" + task.TaskID
	admission, err := uao.practiceEvidence(ctx, goalID)
	if err != nil {
		return err
	}
	payload, err := uao.validatePracticeAdmission(ctx, admission)
	if err != nil {
		return err
	}
	if err := actionContext.Validate(); err != nil {
		return rejectPractice("invalid action context: %v", err)
	}
	canonical, err := uao.practiceEvidence(ctx, actionContext.EventID)
	if err != nil {
		return err
	}
	if actionContext.ContentSHA256 == "" || canonical.ContentSHA256 != actionContext.ContentSHA256 || canonical.Sequence != actionContext.Sequence || actionContext.EventType != persistence.EventTypeContextRetrieved || actionContext.ActionID == "" || actionContext.EventID != stageEventID(actionContext.ActionID, "context") || actionContext.IdempotencyKey != actionContext.EventID || actionContext.PolicyVersion != EnactionPolicyVersion || actionContext.IdentityID != admission.IdentityID || actionContext.GoalID != goalID || admission.Sequence >= actionContext.Sequence || payload.Task != task {
		return rejectPractice("practice admission is not the immutable action predecessor")
	}
	contextTask, err := storedArtifactPractice(actionContext)
	if err != nil || contextTask == nil || *contextTask != task {
		return rejectPractice("practice task differs from action context")
	}
	return nil
}

func (uao *UnifiedAutonomousOrchestrator) queueEvaluatedPractice() error {
	uao.mu.RLock()
	awake := uao.isAwake
	enabled := uao.config.EnableEvaluatedPractice && uao.config.EnableEnaction && uao.config.EnactionMode == EnactionLocalSandbox
	store := uao.eventStore
	scheduler := uao.echobeatsScheduler
	identity, session := uao.identityID, uao.sessionID
	ctx := uao.ctx
	uao.mu.RUnlock()
	if !awake || !enabled || store == nil || scheduler == nil {
		return nil
	}
	uao.practiceMu.Lock()
	defer uao.practiceMu.Unlock()
	for id := range uao.practiceTasks {
		if !uao.practiceTerminal[id] {
			return nil
		}
	}
	if uao.lastPracticeTerminalSequence > uao.lastOrdinaryEnactionSequence && scheduler.HasEligibleOrdinaryGoal() {
		return nil
	}
	index := uao.practiceIndex + 1
	if index <= 0 {
		return fmt.Errorf("practice index exhausted")
	}
	// These are immutable exercise inputs, not canned character output or scores.
	task := generatedPracticeTask(identity, index)
	description := practiceDescription(task)
	payload, err := json.Marshal(practiceGoalEvidence{Description: description, Priority: 1, Index: index, Task: task})
	if err != nil {
		return err
	}
	goalID := "practice-goal:" + task.TaskID
	event := persistence.NewCognitiveEvent(goalID, persistence.EventTypeGoalCreated, identity, session, evaluatedPracticeVersion, payload)
	event.GoalID = goalID
	event.IdempotencyKey = goalID
	result, err := store.Append(ctx, event)
	if err != nil {
		return err
	}
	stored, err := store.Get(ctx, goalID)
	if err != nil {
		return err
	}
	if result.Sequence != stored.Sequence {
		return rejectPractice("practice admission sequence mismatch")
	}
	if _, err := uao.validatePracticeAdmission(ctx, stored); err != nil {
		return err
	}
	if err := scheduler.AdmitPracticeGoal(goalID, description, 1, stored.OccurredAt); err != nil {
		return rejectPractice("scheduler admission: %v", err)
	}
	if uao.practiceTasks == nil {
		uao.practiceTasks = make(map[string]ArtifactPracticeTask)
	}
	uao.practiceTasks[goalID] = task
	uao.practiceIndex = index
	return nil
}

func (uao *UnifiedAutonomousOrchestrator) projectPracticeGoal(event persistence.StoredCognitiveEvent) error {
	if event.PolicyVersion != evaluatedPracticeVersion {
		return nil
	}
	payload, err := uao.validatePracticeAdmission(context.Background(), event)
	if err != nil {
		return err
	}
	uao.mu.RLock()
	enabled, scheduler := uao.config.EnableEvaluatedPractice, uao.echobeatsScheduler
	uao.mu.RUnlock()
	uao.practiceMu.Lock()
	defer uao.practiceMu.Unlock()
	if payload.Index > uao.practiceIndex {
		uao.practiceIndex = payload.Index
	}
	if !enabled || uao.practiceTerminal[event.GoalID] {
		return nil
	}
	if uao.practiceTasks == nil {
		uao.practiceTasks = make(map[string]ArtifactPracticeTask)
	}
	if old, exists := uao.practiceTasks[event.GoalID]; exists && old != payload.Task {
		return rejectPractice("practice task identity conflict")
	}
	if scheduler != nil {
		if err := scheduler.AdmitPracticeGoal(event.GoalID, payload.Description, payload.Priority, event.OccurredAt); err != nil {
			return rejectPractice("scheduler admission: %v", err)
		}
	}
	uao.practiceTasks[event.GoalID] = payload.Task
	return nil
}

// validatedPractice verifies durable evaluation ancestry before a target can
// influence competence or readout weights. Unknown operational outcomes remain
// unknown; they never become invented semantic failures.
func (uao *UnifiedAutonomousOrchestrator) validatedPractice(ctx context.Context, event persistence.StoredCognitiveEvent) (ArtifactPracticeOutcome, ActionPlan, error) {
	var outcome ArtifactPracticeOutcome
	if uao.eventStore == nil {
		return outcome, ActionPlan{}, fmt.Errorf("practice event store unavailable")
	}
	if err := event.Validate(); err != nil {
		return outcome, ActionPlan{}, rejectPractice("invalid outcome envelope: %v", err)
	}
	stored, err := uao.practiceEvidence(ctx, event.EventID)
	if err != nil {
		return outcome, ActionPlan{}, err
	}
	if event.ContentSHA256 == "" || stored.Sequence != event.Sequence || stored.ContentSHA256 != event.ContentSHA256 {
		return outcome, ActionPlan{}, rejectPractice("noncanonical outcome membership")
	}
	outcome, err = ValidateArtifactPracticeOutcome(ctx, uao.eventStore, event)
	if err != nil {
		// The shared contract reports semantic errors without a sentinel. Probe
		// membership again to preserve store/cancellation failures as retryable.
		if _, readErr := uao.practiceEvidence(ctx, event.EventID); readErr != nil {
			return outcome, ActionPlan{}, readErr
		}
		if errors.Is(err, sql.ErrNoRows) {
			return outcome, ActionPlan{}, rejectPractice("missing outcome source")
		}
		switch err.Error() {
		case "invalid practice event envelope", "practice event is not canonical durable evidence", "invalid practice outcome payload", "invalid practice outcome task", "invalid evaluator version or missing reason", "practice source parity mismatch", "practice outcome differs from immutable context task", "unknown practice outcome must not claim evaluated artifact evidence", "unknown outcome requires denied policy", "unknown outcome lacks terminal error/denial provenance", "practice outcome lacks canonical evaluation link", "practice outcome lacks verified artifact evaluation", "invalid practice artifact provenance", "practice artifact hash/source mismatch", "practice lacks allowed tool policy", "practice requires a nondegraded routed model", "practice correctness label disagrees with exact evaluator":
			return outcome, ActionPlan{}, rejectPractice("invalid outcome: %v", err)
		}
		var syntax *json.SyntaxError
		var typeError *json.UnmarshalTypeError
		if errors.As(err, &syntax) || errors.As(err, &typeError) {
			return outcome, ActionPlan{}, rejectPractice("invalid outcome JSON: %v", err)
		}
		return outcome, ActionPlan{}, err
	}
	if event.IdentityID != uao.identityID || event.ActionID == "" || outcome.Task.Version != evaluatedPracticeVersion {
		return outcome, ActionPlan{}, rejectPractice("invalid practice outcome binding")
	}
	actionContext, err := uao.practiceEvidence(ctx, stageEventID(event.ActionID, "context"))
	if err != nil {
		return outcome, ActionPlan{}, err
	}
	if err := uao.ValidateTaskAdmission(ctx, outcome.Task, actionContext); err != nil {
		return outcome, ActionPlan{}, err
	}
	source, err := uao.practiceEvidence(ctx, event.CausationID)
	if err != nil {
		return outcome, ActionPlan{}, err
	}
	if source.Sequence >= event.Sequence || source.IdentityID != event.IdentityID || source.SessionID != event.SessionID || source.ActionID != event.ActionID || source.GoalID != event.GoalID || source.CorrelationID != event.CorrelationID {
		return outcome, ActionPlan{}, rejectPractice("practice outcome source mismatch")
	}
	if outcome.Passed == nil {
		if source.EventType != persistence.EventTypeErrorRecorded && source.EventType != persistence.EventTypeActionFailed && source.EventType != persistence.EventTypePolicyDecision {
			return outcome, ActionPlan{}, rejectPractice("unknown practice lacks operational source")
		}
		return outcome, ActionPlan{}, nil
	}
	if source.EventType != persistence.EventTypeEvaluationRecorded || source.EventID != outcome.EvaluationEventID {
		return outcome, ActionPlan{}, rejectPractice("practice lacks evaluation source")
	}
	var evaluation ActionEvaluation
	if err := json.Unmarshal(source.PayloadJSON, &evaluation); err != nil {
		return outcome, ActionPlan{}, rejectPractice("invalid evaluation: %v", err)
	}
	if !evaluation.Passed || evaluation.Score != 1 {
		return outcome, ActionPlan{}, rejectPractice("unverified artifact cannot be task evidence")
	}
	events, err := uao.eventStore.Query(ctx, persistence.CognitiveEventQuery{ActionID: event.ActionID, Limit: 1000})
	if err != nil {
		return outcome, ActionPlan{}, err
	}
	indexed := indexActionEvents(events)
	for _, kind := range []string{persistence.EventTypeContextRetrieved, persistence.EventTypeProviderRouted, persistence.EventTypePlanProposed, persistence.EventTypeActionRequested, persistence.EventTypePolicyDecision, persistence.EventTypeActionStarted, persistence.EventTypeActionCompleted} {
		item, ok := indexed[kind]
		if !ok || item.Sequence >= source.Sequence || item.IdentityID != event.IdentityID || item.SessionID != event.SessionID || item.GoalID != event.GoalID || item.CorrelationID != event.CorrelationID {
			return outcome, ActionPlan{}, rejectPractice("practice missing/mismatched causal stage %s", kind)
		}
	}
	stages := []struct{ kind, stage string }{
		{persistence.EventTypeContextRetrieved, "context"},
		{persistence.EventTypeProviderRouted, "route"},
		{persistence.EventTypePlanProposed, "plan"},
		{persistence.EventTypeActionRequested, "requested"},
		{persistence.EventTypePolicyDecision, "policy"},
		{persistence.EventTypeActionStarted, "started"},
		{persistence.EventTypeActionCompleted, "completed"},
	}
	previousID := ""
	previousSequence := int64(0)
	for _, stage := range stages {
		item := indexed[stage.kind]
		if item.EventID != stageEventID(event.ActionID, stage.stage) || item.CausationID != previousID || item.Sequence <= previousSequence || item.PolicyVersion != EnactionPolicyVersion {
			return outcome, ActionPlan{}, rejectPractice("noncanonical practice causal link %s", stage.kind)
		}
		previousID = item.EventID
		previousSequence = item.Sequence
	}
	if source.CausationID != previousID || source.EventID != stageEventID(event.ActionID, "evaluation") || source.PolicyVersion != EnactionPolicyVersion {
		return outcome, ActionPlan{}, rejectPractice("noncanonical practice evaluation link")
	}
	contextTask, err := storedArtifactPractice(indexed[persistence.EventTypeContextRetrieved])
	if err != nil || contextTask == nil || !reflect.DeepEqual(*contextTask, outcome.Task) {
		return outcome, ActionPlan{}, rejectPractice("practice task differs from canonical action context")
	}
	var plan ActionPlan
	if err := json.Unmarshal(indexed[persistence.EventTypePlanProposed].PayloadJSON, &plan); err != nil {
		return outcome, plan, rejectPractice("invalid plan: %v", err)
	}
	if err := plan.Validate(); err != nil {
		return outcome, plan, rejectPractice("invalid plan: %v", err)
	}
	var observation ActionObservation
	if err := json.Unmarshal(indexed[persistence.EventTypeActionCompleted].PayloadJSON, &observation); err != nil {
		return outcome, plan, rejectPractice("invalid observation: %v", err)
	}
	var request struct {
		Tool            string `json:"tool"`
		RelativePath    string `json:"relative_path"`
		RequestedSHA256 string `json:"requested_sha256"`
	}
	if err := json.Unmarshal(indexed[persistence.EventTypeActionRequested].PayloadJSON, &request); err != nil {
		return outcome, plan, rejectPractice("invalid request: %v", err)
	}
	if request.Tool != plan.Tool || request.RelativePath != plan.RelativePath || request.RequestedSHA256 != sha256Hex([]byte(plan.Content)) || observation.RelativePath != plan.RelativePath {
		return outcome, plan, rejectPractice("practice request/plan mismatch")
	}
	if observation.Tool != WorkspaceCreateNote || observation.Conflict || (!observation.Created && !observation.Recovered) || observation.ContentSHA256 != sha256Hex([]byte(plan.Content)) || observation.ObservedSHA256 != observation.ContentSHA256 || outcome.ArtifactSHA256 != observation.ObservedSHA256 {
		return outcome, plan, rejectPractice("practice artifact hash mismatch")
	}
	var decision ActionPolicyDecision
	if err := json.Unmarshal(indexed[persistence.EventTypePolicyDecision].PayloadJSON, &decision); err != nil {
		return outcome, plan, rejectPractice("invalid decision: %v", err)
	}
	if !decision.Allowed || decision.Mode != EnactionLocalSandbox {
		return outcome, plan, rejectPractice("practice lacks local authorization")
	}
	var route PlanRouteEvidence
	if err := json.Unmarshal(indexed[persistence.EventTypeProviderRouted].PayloadJSON, &route); err != nil {
		return outcome, plan, rejectPractice("invalid route: %v", err)
	}
	if route.Degraded || route.Provider == "" || route.BackendKind == "fallback" {
		return outcome, plan, rejectPractice("practice source is not a real model route")
	}
	passed, reason, err := EvaluateArtifactPracticeContent(outcome.Task, []byte(plan.Content))
	if err != nil || passed != *outcome.Passed || reason != outcome.Reason || outcome.EvaluatorVersion != evaluatedPracticeVersion {
		return outcome, plan, rejectPractice("practice evaluator result does not match independently recomputed content")
	}
	return outcome, plan, nil
}

// rememberPracticeTerminalLocked bounds advisory receipts. Eviction cannot
// invalidate durable labels or reactivate a retired task; the ledger is authority.
func (uao *UnifiedAutonomousOrchestrator) rememberPracticeTerminalLocked(goalID string, sequence int64) {
	if uao.practiceTerminal == nil {
		uao.practiceTerminal = make(map[string]bool)
	}
	if !uao.practiceTerminal[goalID] {
		uao.practiceTerminalOrder = append(uao.practiceTerminalOrder, goalID)
	}
	uao.practiceTerminal[goalID] = true
	delete(uao.practiceTasks, goalID)
	if sequence > uao.lastPracticeTerminalSequence {
		uao.lastPracticeTerminalSequence = sequence
	}
	for len(uao.practiceTerminalOrder) > 256 {
		delete(uao.practiceTerminal, uao.practiceTerminalOrder[0])
		uao.practiceTerminalOrder = uao.practiceTerminalOrder[1:]
	}
}

func (uao *UnifiedAutonomousOrchestrator) projectPracticeTerminal(event persistence.StoredCognitiveEvent) error {
	outcome, _, err := uao.validatedPractice(context.Background(), event)
	if err != nil {
		return err
	}
	if outcome.Passed != nil && uao.skillLearning != nil {
		score := 0.0
		if *outcome.Passed {
			score = 1
		}
		if _, err := uao.skillLearning.ApplyEvaluatorEvidence(evaluatedPracticeSkill, event.EventID, score, *outcome.Passed, outcome.Reason); err != nil {
			return err
		}
	}
	uao.practiceMu.Lock()
	uao.rememberPracticeTerminalLocked(event.GoalID, event.Sequence)
	uao.practiceMu.Unlock()
	if uao.echobeatsScheduler != nil {
		uao.echobeatsScheduler.FinishPracticeGoal(event.GoalID, event.EventID, outcome.Passed != nil && *outcome.Passed)
	}
	return nil
}
