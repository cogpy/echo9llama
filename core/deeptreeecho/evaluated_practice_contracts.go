package deeptreeecho

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cogpy/echo9llama/core/backendcap"
	"github.com/cogpy/echo9llama/core/persistence"
	"github.com/cogpy/echo9llama/core/tools"
)

const (
	ArithmeticArtifactVersion          = "arithmetic-artifact-v1"
	ArithmeticArtifactSkillName        = "Arithmetic Artifact Construction"
	ArtifactPracticeOperandLimit int64 = 1_000_000
)

// ArtifactPracticeTask is an opt-in, immutable specification. Expected answers
// are computed by the evaluator, never supplied by or scored by the planner.
type ArtifactPracticeTask struct {
	TaskID  string `json:"task_id"`
	Version string `json:"version"`
	A       int64  `json:"a"`
	B       int64  `json:"b"`
}

func (task ArtifactPracticeTask) Validate() error {
	if task.TaskID == "" || task.TaskID != strings.TrimSpace(task.TaskID) || len(task.TaskID) > 256 || !utf8.ValidString(task.TaskID) {
		return fmt.Errorf("task_id must be nonempty, bounded, and unpadded")
	}
	if task.Version != ArithmeticArtifactVersion {
		return fmt.Errorf("unsupported arithmetic artifact version %q", task.Version)
	}
	if task.A < -ArtifactPracticeOperandLimit || task.A > ArtifactPracticeOperandLimit || task.B < -ArtifactPracticeOperandLimit || task.B > ArtifactPracticeOperandLimit {
		return fmt.Errorf("arithmetic operands must be within +/-1000000")
	}
	return nil
}

// ArtifactPracticeOutcome is distinct from successful file publication. A nil
// Passed means no canonical correctness judgment was possible, not a failure.
type ArtifactPracticeOutcome struct {
	Task              ArtifactPracticeTask `json:"task"`
	EvaluationEventID string               `json:"evaluation_event_id"`
	ArtifactSHA256    string               `json:"artifact_sha256"`
	Passed            *bool                `json:"passed"`
	Reason            string               `json:"reason"`
	EvaluatorVersion  string               `json:"evaluator_version"`
}

// EvaluateArtifactPractice is the independent exact semantic evaluator exposed
// to replay admission. Only verified artifact bytes may supply learning labels.
func EvaluateArtifactPractice(task ArtifactPracticeTask, content []byte) (bool, string, error) {
	return EvaluateArtifactPracticeContent(task, content)
}

// EvaluateArtifactPracticeContent performs exact arithmetic on already verified
// artifact bytes. Callers must not interpret a task validation error as a label.
func EvaluateArtifactPracticeContent(task ArtifactPracticeTask, content []byte) (bool, string, error) {
	if err := task.Validate(); err != nil {
		return false, "", err
	}
	if !utf8.Valid(content) {
		return false, "invalid_arithmetic_json", nil
	}
	value := strings.TrimSpace(string(content))
	if strings.HasPrefix(value, "```") {
		line := strings.IndexByte(value, '\n')
		if line < 0 || strings.TrimSpace(value[:line]) != "```json" || !strings.HasSuffix(value, "```") {
			return false, "invalid_arithmetic_json", nil
		}
		value = strings.TrimSpace(strings.TrimSuffix(value[line+1:], "```"))
		if strings.Contains(value, "```") {
			return false, "invalid_arithmetic_json", nil
		}
	}
	fields, err := strictPracticeObject([]byte(value))
	if err != nil || len(fields) != 3 {
		return false, "invalid_arithmetic_json", nil //nolint:nilerr // Already-read-back malformed content is a known negative task outcome, not an operational error.
	}
	var taskID string
	var sum, product int64
	if fields["task_id"] == nil || fields["sum"] == nil || fields["product"] == nil ||
		bytes.Equal(bytes.TrimSpace(fields["task_id"]), []byte("null")) ||
		bytes.Equal(bytes.TrimSpace(fields["sum"]), []byte("null")) ||
		bytes.Equal(bytes.TrimSpace(fields["product"]), []byte("null")) ||
		json.Unmarshal(fields["task_id"], &taskID) != nil || json.Unmarshal(fields["sum"], &sum) != nil || json.Unmarshal(fields["product"], &product) != nil {
		return false, "invalid_arithmetic_json", nil //nolint:nilerr // Invalid answer field types are a measured task failure; no inference/runtime failure is concealed.
	}
	if taskID != task.TaskID {
		return false, "task_id_mismatch", nil
	}
	// Operand validation proves these operations cannot overflow int64.
	if sum != task.A+task.B || product != task.A*task.B {
		return false, "arithmetic_answer_mismatch", nil
	}
	return true, "exact_arithmetic_verified", nil
}

// strictPracticeObject rejects duplicate keys (including escaped equivalents),
// nonobjects, and trailing material. Typed decoding then enforces field types.
func strictPracticeObject(raw []byte) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, fmt.Errorf("expected a JSON object")
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := token.(string)
		if !ok {
			return nil, fmt.Errorf("invalid object key")
		}
		if _, exists := fields[key]; exists {
			return nil, fmt.Errorf("duplicate field %q", key)
		}
		var rawValue json.RawMessage
		if err := decoder.Decode(&rawValue); err != nil {
			return nil, err
		}
		fields[key] = rawValue
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("trailing JSON material")
	}
	return fields, nil
}

func storedArtifactPractice(event persistence.StoredCognitiveEvent) (*ArtifactPracticeTask, error) {
	fields, err := strictPracticeObject(event.PayloadJSON)
	if err != nil {
		return nil, err
	}
	raw, exists := fields["practice"]
	if !exists || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	return decodeArtifactPracticeTask(raw)
}

func decodeArtifactPracticeTask(raw json.RawMessage) (*ArtifactPracticeTask, error) {
	taskFields, err := strictPracticeObject(raw)
	if err != nil || len(taskFields) != 4 || taskFields["task_id"] == nil || taskFields["version"] == nil || taskFields["a"] == nil || taskFields["b"] == nil {
		return nil, fmt.Errorf("invalid stored arithmetic task")
	}
	for _, value := range taskFields {
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, fmt.Errorf("null stored arithmetic task field")
		}
	}
	var task ArtifactPracticeTask
	if err := json.Unmarshal(raw, &task); err != nil {
		return nil, err
	}
	if err := task.Validate(); err != nil {
		return nil, err
	}
	return &task, nil
}

func practiceTaskParity(request EnactionRequest, indexed map[string]persistence.StoredCognitiveEvent, actionID string) error {
	stored, exists := indexed[persistence.EventTypeContextRetrieved]
	if !exists {
		if request.Practice != nil && len(indexed) > 0 {
			return fmt.Errorf("practice action lacks its immutable context task")
		}
		return nil
	}
	if stored.EventID != stageEventID(actionID, "context") || stored.IdentityID != request.IdentityID || stored.GoalID != request.GoalID || stored.ActionID != actionID || stored.CorrelationID != "enaction:"+request.GoalID {
		return fmt.Errorf("practice context source parity mismatch")
	}
	task, err := storedArtifactPractice(stored)
	if err != nil {
		return fmt.Errorf("decode immutable practice context: %w", err)
	}
	if (task == nil) != (request.Practice == nil) || (task != nil && *task != *request.Practice) {
		return fmt.Errorf("request practice differs from immutable context task")
	}
	return nil
}

// ValidateArtifactPracticeOutcome is the event-admission boundary for learning.
// It verifies durable membership, immutable task and causal source parity, and
// known-label correctness against the plan bytes whose hash the tool verified.
// It does not call a model or access the workspace; the original evaluator does
// mandatory tool read-back before creating the durable correctness outcome.
func ValidateArtifactPracticeOutcome(ctx context.Context, store *persistence.CognitiveEventStore, event persistence.StoredCognitiveEvent) (ArtifactPracticeOutcome, error) {
	var outcome ArtifactPracticeOutcome
	if store == nil || event.EventType != "practice.evaluated" || event.ActionID == "" || event.EventID != stageEventID(event.ActionID, "practice-evaluation") || event.IdempotencyKey != event.EventID {
		return outcome, fmt.Errorf("invalid practice event envelope")
	}
	stored, err := store.Get(ctx, event.EventID)
	if err != nil || stored.ContentSHA256 != event.ContentSHA256 || stored.Sequence != event.Sequence || event.Validate() != nil {
		return outcome, fmt.Errorf("practice event is not canonical durable evidence")
	}
	fields, err := strictPracticeObject(event.PayloadJSON)
	if err != nil || len(fields) != 6 || fields["task"] == nil || fields["passed"] == nil || fields["evaluation_event_id"] == nil || fields["artifact_sha256"] == nil || fields["reason"] == nil || fields["evaluator_version"] == nil {
		return outcome, fmt.Errorf("invalid practice outcome payload")
	}
	if err := json.Unmarshal(event.PayloadJSON, &outcome); err != nil {
		return outcome, err
	}
	payloadTask, err := decodeArtifactPracticeTask(fields["task"])
	if err != nil || payloadTask == nil || *payloadTask != outcome.Task {
		return outcome, fmt.Errorf("invalid practice outcome task")
	}
	if outcome.EvaluatorVersion != ArithmeticArtifactVersion || outcome.Reason == "" {
		return outcome, fmt.Errorf("invalid evaluator version or missing reason")
	}
	getSource := func(id, kind string) (persistence.StoredCognitiveEvent, error) {
		source, err := store.Get(ctx, id)
		if err != nil {
			return source, err
		}
		if source.EventType != kind || source.IdentityID != event.IdentityID || source.SessionID != event.SessionID || source.ActionID != event.ActionID || source.GoalID != event.GoalID || source.CorrelationID != event.CorrelationID || source.PolicyVersion != event.PolicyVersion || source.Sequence >= event.Sequence {
			return source, fmt.Errorf("practice source parity mismatch")
		}
		return source, nil
	}
	contextEvent, err := getSource(stageEventID(event.ActionID, "context"), persistence.EventTypeContextRetrieved)
	if err != nil {
		return outcome, err
	}
	task, err := storedArtifactPractice(contextEvent)
	if err != nil || task == nil || *task != outcome.Task {
		return outcome, fmt.Errorf("practice outcome differs from immutable context task")
	}
	if outcome.Passed == nil {
		if outcome.EvaluationEventID != "" || outcome.ArtifactSHA256 != "" || !strings.HasPrefix(outcome.Reason, "not_evaluated:") {
			return outcome, fmt.Errorf("unknown practice outcome must not claim evaluated artifact evidence")
		}
		source, err := store.Get(ctx, event.CausationID)
		if err != nil {
			return outcome, err
		}
		switch source.EventType {
		case persistence.EventTypeErrorRecorded, persistence.EventTypeActionFailed:
		case persistence.EventTypePolicyDecision:
			var decision ActionPolicyDecision
			if json.Unmarshal(source.PayloadJSON, &decision) != nil || decision.Allowed {
				return outcome, fmt.Errorf("unknown outcome requires denied policy")
			}
		default:
			return outcome, fmt.Errorf("unknown outcome lacks terminal error/denial provenance")
		}
		_, err = getSource(source.EventID, source.EventType)
		return outcome, err
	}
	if outcome.EvaluationEventID != stageEventID(event.ActionID, "evaluation") || event.CausationID != outcome.EvaluationEventID {
		return outcome, fmt.Errorf("practice outcome lacks canonical evaluation link")
	}
	evaluationEvent, err := getSource(outcome.EvaluationEventID, persistence.EventTypeEvaluationRecorded)
	if err != nil {
		return outcome, err
	}
	var evaluation ActionEvaluation
	if json.Unmarshal(evaluationEvent.PayloadJSON, &evaluation) != nil || !evaluation.Passed || evaluation.EvidenceEventID != evaluationEvent.EventID || evaluationEvent.CausationID != stageEventID(event.ActionID, "completed") {
		return outcome, fmt.Errorf("practice outcome lacks verified artifact evaluation")
	}
	completion, err := getSource(stageEventID(event.ActionID, "completed"), persistence.EventTypeActionCompleted)
	if err != nil {
		return outcome, err
	}
	planEvent, err := getSource(stageEventID(event.ActionID, "plan"), persistence.EventTypePlanProposed)
	if err != nil {
		return outcome, err
	}
	var plan ActionPlan
	var observation ActionObservation
	if json.Unmarshal(planEvent.PayloadJSON, &plan) != nil || plan.Validate() != nil || json.Unmarshal(completion.PayloadJSON, &observation) != nil {
		return outcome, fmt.Errorf("invalid practice artifact provenance")
	}
	if observation.Tool != WorkspaceCreateNote || observation.Conflict || (!observation.Created && !observation.Recovered) || observation.RelativePath != plan.RelativePath || observation.ContentSHA256 != sha256Hex([]byte(plan.Content)) || observation.ObservedSHA256 != observation.ContentSHA256 || outcome.ArtifactSHA256 != observation.ObservedSHA256 || completion.Sequence >= evaluationEvent.Sequence || planEvent.Sequence >= completion.Sequence {
		return outcome, fmt.Errorf("practice artifact hash/source mismatch")
	}
	routeEvent, err := getSource(stageEventID(event.ActionID, "route"), persistence.EventTypeProviderRouted)
	if err != nil {
		return outcome, err
	}
	policyEvent, err := getSource(stageEventID(event.ActionID, "policy"), persistence.EventTypePolicyDecision)
	if err != nil {
		return outcome, err
	}
	var decision ActionPolicyDecision
	if json.Unmarshal(policyEvent.PayloadJSON, &decision) != nil || !decision.Allowed {
		return outcome, fmt.Errorf("practice lacks allowed tool policy")
	}
	var route PlanRouteEvidence
	if json.Unmarshal(routeEvent.PayloadJSON, &route) != nil || route.Provider == "" || route.BackendKind == "" || route.Degraded || route.BackendKind == string(backendcap.BackendFallback) || routeEvent.Degraded || contextEvent.Sequence >= planEvent.Sequence || routeEvent.Sequence >= planEvent.Sequence || policyEvent.Sequence >= completion.Sequence {
		return outcome, fmt.Errorf("practice requires a nondegraded routed model")
	}
	passed, reason, err := EvaluateArtifactPracticeContent(outcome.Task, []byte(plan.Content))
	if err != nil || passed != *outcome.Passed || reason != outcome.Reason {
		return outcome, fmt.Errorf("practice correctness label disagrees with exact evaluator")
	}
	return outcome, nil
}

// completeArtifactPractice is called on every terminal path and on replay. It
// performs only durable reads and read-only tool reconciliation, never creation
// or inference. Cancellation stops effects, not the bounded audit append.
func (pipeline *EnactionPipeline) completeArtifactPractice(ctx context.Context, request EnactionRequest, actionID, correlationID string, runErr error) error {
	if request.Practice == nil {
		return nil
	}
	auditCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	events, err := pipeline.store.Query(auditCtx, persistence.CognitiveEventQuery{ActionID: actionID, Limit: 1000})
	if err != nil {
		return err
	}
	indexed := indexActionEvents(events)
	if err := practiceTaskParity(request, indexed, actionID); err != nil {
		return err
	}
	if existing, exists := indexed["practice.evaluated"]; exists {
		_, err := ValidateArtifactPracticeOutcome(auditCtx, pipeline.store, existing)
		return err
	}
	// Without committed context there is no admissible canonical task to label.
	if _, exists := indexed[persistence.EventTypeContextRetrieved]; !exists {
		return nil
	}
	outcome := ArtifactPracticeOutcome{Task: *request.Practice, EvaluatorVersion: ArithmeticArtifactVersion}
	causation := ""
	for _, kind := range []string{persistence.EventTypeActionFailed, persistence.EventTypeErrorRecorded, persistence.EventTypePolicyDecision} {
		if source, exists := indexed[kind]; exists {
			if kind == persistence.EventTypePolicyDecision {
				var decision ActionPolicyDecision
				if json.Unmarshal(source.PayloadJSON, &decision) != nil || decision.Allowed {
					continue
				}
				outcome.Reason = "not_evaluated:policy_denied"
			} else {
				outcome.Reason = "not_evaluated:" + kind
			}
			causation = source.EventID
			break
		}
	}
	if causation == "" {
		var plan ActionPlan
		var evaluation ActionEvaluation
		var observation ActionObservation
		planEvent, planned := indexed[persistence.EventTypePlanProposed]
		evaluationEvent, evaluated := indexed[persistence.EventTypeEvaluationRecorded]
		completion, completed := indexed[persistence.EventTypeActionCompleted]
		canonicalSources := true
		for _, source := range []struct{ kind, stage string }{
			{persistence.EventTypePlanProposed, "plan"},
			{persistence.EventTypeProviderRouted, "route"},
			{persistence.EventTypePolicyDecision, "policy"},
			{persistence.EventTypeActionCompleted, "completed"},
			{persistence.EventTypeEvaluationRecorded, "evaluation"},
		} {
			stored, exists := indexed[source.kind]
			if !exists || stored.EventID != stageEventID(actionID, source.stage) || stored.IdentityID != request.IdentityID || stored.ActionID != actionID || stored.GoalID != request.GoalID || stored.CorrelationID != correlationID || stored.PolicyVersion != EnactionPolicyVersion {
				canonicalSources = false
			}
		}
		route := routeFromEvents(indexed)
		var decision ActionPolicyDecision
		policyEvent := indexed[persistence.EventTypePolicyDecision]
		canonicalSources = canonicalSources && route.Provider != "" && route.BackendKind != "" && !route.Degraded && route.BackendKind != string(backendcap.BackendFallback) && !indexed[persistence.EventTypeProviderRouted].Degraded && json.Unmarshal(policyEvent.PayloadJSON, &decision) == nil && decision.Allowed && planEvent.Sequence < completion.Sequence && completion.Sequence < evaluationEvent.Sequence && evaluationEvent.CausationID == completion.EventID
		if canonicalSources && planned && evaluated && completed && json.Unmarshal(planEvent.PayloadJSON, &plan) == nil && plan.Validate() == nil && json.Unmarshal(evaluationEvent.PayloadJSON, &evaluation) == nil && evaluation.Passed && evaluation.EvidenceEventID == evaluationEvent.EventID && json.Unmarshal(completion.PayloadJSON, &observation) == nil && observation.Tool == WorkspaceCreateNote && observation.RelativePath == plan.RelativePath && observation.ContentSHA256 == sha256Hex([]byte(plan.Content)) && observation.ObservedSHA256 == observation.ContentSHA256 && !observation.Conflict && (observation.Created || observation.Recovered) {
			readBack, verifyErr := pipeline.tool.ReconcileContext(ctx, tools.WorkspaceNoteRequest{Path: plan.RelativePath, Content: []byte(plan.Content)})
			if verifyErr == nil && readBack.Status == tools.WorkspaceNoteMatchingExisting && !readBack.Blocked && readBack.SHA256 == observation.ObservedSHA256 && sha256Hex(readBack.ReadBack) == readBack.SHA256 {
				passed, reason, evaluateErr := EvaluateArtifactPracticeContent(*request.Practice, readBack.ReadBack)
				if evaluateErr != nil {
					return evaluateErr
				}
				outcome.Passed = &passed
				outcome.Reason = reason
				outcome.ArtifactSHA256 = readBack.SHA256
				outcome.EvaluationEventID = evaluationEvent.EventID
				causation = evaluationEvent.EventID
			} else {
				if verifyErr == nil {
					verifyErr = fmt.Errorf("practice artifact read-back unavailable or conflicting")
				}
				runErr = verifyErr
			}
		}
		if causation == "" && evaluated && completed && runErr == nil {
			runErr = fmt.Errorf("canonical artifact integrity evaluation unavailable")
		}
		if causation == "" {
			if runErr == nil {
				return nil // An incomplete action has not reached a labelable terminal.
			}
			if err := pipeline.appendError(auditCtx, actionID, correlationID, request.GoalID, "practice-verification", classifyEnactionError(runErr), runErr); err != nil {
				return err
			}
			causation = stageEventID(actionID, "error-practice-verification")
			outcome.Reason = "not_evaluated:" + classifyEnactionError(runErr)
		}
	}
	_, err = pipeline.appendEvent(auditCtx, stageEventID(actionID, "practice-evaluation"), "practice.evaluated", correlationID, causation, request.GoalID, actionID, persistence.EvidenceClassSensitive, outcome, PlanRouteEvidence{})
	return err
}

// recoverArtifactPracticeAction prevents a restarted immutable task from being
// mistaken for the next generic artifact iteration after goal.progressed.
func (pipeline *EnactionPipeline) recoverArtifactPracticeAction(ctx context.Context, request EnactionRequest, mode EnactionMode) (string, error) {
	events, err := pipeline.store.Query(ctx, persistence.CognitiveEventQuery{IdentityID: request.IdentityID, GoalID: request.GoalID, EventType: persistence.EventTypeContextRetrieved, Limit: 1000})
	if err != nil {
		return "", err
	}
	var recovered string
	for _, event := range events {
		if !strings.Contains(event.ActionID, ":workspace.create_note:"+string(mode)+":iteration-") {
			continue
		}
		task, err := storedArtifactPractice(event)
		if err != nil {
			return "", err
		}
		if task == nil && request.Practice == nil {
			continue // Preserve ordinary generic artifact iteration behavior.
		}
		if task == nil || request.Practice == nil || *task != *request.Practice {
			return "", fmt.Errorf("practice goal already has a different immutable context task")
		}
		if recovered != "" && recovered != event.ActionID {
			return "", errors.New("practice task has ambiguous action provenance")
		}
		recovered = event.ActionID
	}
	return recovered, nil
}
