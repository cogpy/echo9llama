package deeptreeecho

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cogpy/echo9llama/core/persistence"
)

// validateReflectionActionChain proves only the structural provenance of an E1
// reflection source against the append-only event ledger. It deliberately does
// not mutate the ledger, award authority, or synthesize a score.
//
// Trust limitation: this verifies that durable records have the causal shape
// emitted by enaction_pipeline.go. It is not cryptographic provenance against
// a hostile writer that is already authorized to create arbitrary valid ledger
// entries. Such protection requires a separate signer/attestation boundary.
func validateReflectionActionChain(ctx context.Context, store *persistence.CognitiveEventStore, event persistence.StoredCognitiveEvent) error {
	if store == nil {
		return fmt.Errorf("reflection causal validation requires an event store")
	}
	if event.EventID == "" || event.Sequence <= 0 {
		return fmt.Errorf("reflection source is not a durable ledger event")
	}

	// Reload the candidate so a caller cannot validate a detached event-shaped
	// value and then use it as if it were the durable terminal record.
	persisted, err := store.Get(ctx, event.EventID)
	if err != nil {
		return fmt.Errorf("load reflection source %q: %w", event.EventID, err)
	}
	if persisted.Sequence != event.Sequence || persisted.ContentSHA256 != event.ContentSHA256 {
		return fmt.Errorf("reflection source does not match durable ledger record")
	}
	event = persisted

	switch event.EventType {
	case persistence.EventTypeEvaluationRecorded:
		return validateReflectionEvaluationChain(ctx, store, event)
	case persistence.EventTypeActionFailed:
		return validateReflectionFailedActionChain(ctx, store, event)
	case persistence.EventTypeErrorRecorded:
		return validateReflectionPlanningFailureChain(ctx, store, event)
	default:
		return fmt.Errorf("unsupported reflection causal source type %q", event.EventType)
	}
}

type reflectionContextPayload struct {
	Goal               string   `json:"goal"`
	TopInterests       []string `json:"top_interests"`
	RecentContextCount int      `json:"recent_context_count"`
}

type reflectionRequestedPayload struct {
	Tool            string `json:"tool"`
	Purpose         string `json:"purpose"`
	RelativePath    string `json:"relative_path"`
	RequestedSHA256 string `json:"requested_sha256"`
	Reversible      bool   `json:"reversible"`
}

type reflectionStartedPayload struct {
	Tool            string `json:"tool"`
	RelativePath    string `json:"relative_path"`
	RequestedSHA256 string `json:"requested_sha256"`
}

type reflectionPlanningErrorPayload struct {
	Stage    string `json:"stage"`
	Category string `json:"category"`
	Message  string `json:"message"`
}

func validateReflectionEvaluationChain(ctx context.Context, store *persistence.CognitiveEventStore, evaluation persistence.StoredCognitiveEvent) error {
	if err := validateReflectionActionBase(evaluation, true); err != nil {
		return err
	}
	if err := validateReflectionStage(evaluation, "evaluation", persistence.EventTypeEvaluationRecorded, stageEventID(evaluation.ActionID, "completed"), persistence.EvidenceClassSensitive); err != nil {
		return err
	}
	var evaluationPayload ActionEvaluation
	if err := strictReflectionJSON(evaluation.PayloadJSON, &evaluationPayload); err != nil {
		return fmt.Errorf("decode action evaluation: %w", err)
	}
	if evaluationPayload.EvidenceEventID != evaluation.EventID || !finiteUnit(evaluationPayload.Score) || evaluationPayload.Score != boolScore(evaluationPayload.Passed) || len(evaluationPayload.Criteria) != 4 || strings.TrimSpace(evaluationPayload.Reason) == "" {
		return fmt.Errorf("invalid deterministic action evaluation")
	}
	if err := validateReflectionUnroutedEnvelope(evaluation); err != nil {
		return fmt.Errorf("evaluation envelope: %w", err)
	}

	chain, plan, route, originalSession, err := loadReflectionActionPrefix(ctx, store, evaluation)
	if err != nil {
		return err
	}
	completed, err := reflectionLoadStage(ctx, store, evaluation.ActionID, "completed")
	if err != nil {
		return err
	}
	if err := validateReflectionRelated(evaluation, completed); err != nil {
		return fmt.Errorf("completed causal binding: %w", err)
	}
	if completed.EventType != persistence.EventTypeActionCompleted {
		return fmt.Errorf("evaluation requires action.completed, got %q", completed.EventType)
	}
	if err := validateReflectionStage(completed, "completed", persistence.EventTypeActionCompleted, stageEventID(evaluation.ActionID, "started"), persistence.EvidenceClassSensitive); err != nil {
		return err
	}
	var observation ActionObservation
	if err := strictReflectionJSON(completed.PayloadJSON, &observation); err != nil {
		return fmt.Errorf("decode action completion: %w", err)
	}
	if err := validateReflectionObservation(observation, plan); err != nil {
		return fmt.Errorf("invalid action completion: %w", err)
	}
	if err := validateReflectionTerminalSessions(originalSession, chain.policy, chain.started, completed, evaluation, observation); err != nil {
		return err
	}
	if completed.Sequence <= chain.started.Sequence || evaluation.Sequence <= completed.Sequence {
		return fmt.Errorf("completion/evaluation sequence must follow action.started")
	}
	if evaluation.SessionID != completed.SessionID {
		return fmt.Errorf("evaluation session must match completed action session")
	}
	if err := validateReflectionRoutedEnvelope(completed, route); err != nil {
		return fmt.Errorf("completed route binding: %w", err)
	}
	return nil
}

func validateReflectionFailedActionChain(ctx context.Context, store *persistence.CognitiveEventStore, failed persistence.StoredCognitiveEvent) error {
	if err := validateReflectionActionBase(failed, true); err != nil {
		return err
	}
	if err := validateReflectionStage(failed, "failed", persistence.EventTypeActionFailed, stageEventID(failed.ActionID, "started"), persistence.EvidenceClassSensitive); err != nil {
		return err
	}
	var observation ActionObservation
	if err := strictReflectionJSON(failed.PayloadJSON, &observation); err != nil {
		return fmt.Errorf("decode action failure: %w", err)
	}
	if strings.TrimSpace(observation.ErrorCategory) == "" {
		return fmt.Errorf("action failure has no error category")
	}

	chain, plan, route, originalSession, err := loadReflectionActionPrefix(ctx, store, failed)
	if err != nil {
		return err
	}
	if err := validateReflectionObservation(observation, plan); err != nil {
		return fmt.Errorf("invalid action failure: %w", err)
	}
	if err := validateReflectionTerminalSessions(originalSession, chain.policy, chain.started, failed, persistence.StoredCognitiveEvent{}, observation); err != nil {
		return err
	}
	if failed.Sequence <= chain.started.Sequence {
		return fmt.Errorf("failed action sequence must follow action.started")
	}
	if err := validateReflectionRoutedEnvelope(failed, route); err != nil {
		return fmt.Errorf("failed action route binding: %w", err)
	}
	return nil
}

func validateReflectionPlanningFailureChain(ctx context.Context, store *persistence.CognitiveEventStore, planningError persistence.StoredCognitiveEvent) error {
	if err := validateReflectionActionBase(planningError, false); err != nil {
		return err
	}
	if err := validateReflectionStage(planningError, "error-planning", persistence.EventTypeErrorRecorded, "", persistence.EvidenceClassOperational); err != nil {
		return err
	}
	var payload reflectionPlanningErrorPayload
	if err := strictReflectionJSON(planningError.PayloadJSON, &payload); err != nil {
		return fmt.Errorf("decode planning failure: %w", err)
	}
	if payload.Stage != "planning" || strings.TrimSpace(payload.Category) == "" || strings.TrimSpace(payload.Message) == "" {
		return fmt.Errorf("invalid planning failure payload")
	}
	if err := validateReflectionUnroutedEnvelope(planningError); err != nil {
		return fmt.Errorf("planning failure envelope: %w", err)
	}

	contextEvent, err := reflectionLoadStage(ctx, store, planningError.ActionID, "context")
	if err != nil {
		return err
	}
	if err := validateReflectionRelated(planningError, contextEvent); err != nil {
		return fmt.Errorf("planning context binding: %w", err)
	}
	if err := validateReflectionStage(contextEvent, "context", persistence.EventTypeContextRetrieved, "", persistence.EvidenceClassSensitive); err != nil {
		return err
	}
	if err := validateReflectionContext(contextEvent); err != nil {
		return err
	}
	if contextEvent.Sequence >= planningError.Sequence {
		return fmt.Errorf("planning context must precede planning failure")
	}
	if contextEvent.SessionID != planningError.SessionID {
		return fmt.Errorf("planning failure session must match context session")
	}

	// A route is emitted only after the planner returns route evidence. It is
	// consequently optional for early planner failures, but must be a valid
	// context-bound producer record whenever it exists.
	routeEvent, err := store.Get(ctx, stageEventID(planningError.ActionID, "route"))
	if err != nil {
		if persistenceErrNotFound(err) {
			return nil
		}
		return fmt.Errorf("load optional planning route: %w", err)
	}
	if err := validateReflectionRelated(planningError, routeEvent); err != nil {
		return fmt.Errorf("planning route binding: %w", err)
	}
	if err := validateReflectionStage(routeEvent, "route", persistence.EventTypeProviderRouted, contextEvent.EventID, persistence.EvidenceClassOperational); err != nil {
		return err
	}
	if routeEvent.Sequence <= contextEvent.Sequence || routeEvent.Sequence >= planningError.Sequence {
		return fmt.Errorf("planning route sequence is outside context/failure boundary")
	}
	if routeEvent.SessionID != contextEvent.SessionID {
		return fmt.Errorf("planning route session must match context session")
	}
	if err := validateReflectionRoute(routeEvent, false); err != nil {
		return err
	}
	return nil
}

type reflectionActionChain struct {
	context   persistence.StoredCognitiveEvent
	route     persistence.StoredCognitiveEvent
	plan      persistence.StoredCognitiveEvent
	requested persistence.StoredCognitiveEvent
	policy    persistence.StoredCognitiveEvent
	started   persistence.StoredCognitiveEvent
}

// loadReflectionActionPrefix reads the E1 producer prefix in exact causal
// order. The terminal stage is validated by the caller because evaluations and
// failures have different terminal event types.
func loadReflectionActionPrefix(ctx context.Context, store *persistence.CognitiveEventStore, terminal persistence.StoredCognitiveEvent) (reflectionActionChain, ActionPlan, PlanRouteEvidence, string, error) {
	var chain reflectionActionChain
	stages := []struct {
		name  string
		dest  *persistence.StoredCognitiveEvent
		typ   string
		cause string
		class string
	}{
		{"context", &chain.context, persistence.EventTypeContextRetrieved, "", persistence.EvidenceClassSensitive},
		{"route", &chain.route, persistence.EventTypeProviderRouted, stageEventID(terminal.ActionID, "context"), persistence.EvidenceClassOperational},
		{"plan", &chain.plan, persistence.EventTypePlanProposed, stageEventID(terminal.ActionID, "route"), persistence.EvidenceClassRestricted},
		{"requested", &chain.requested, persistence.EventTypeActionRequested, stageEventID(terminal.ActionID, "plan"), persistence.EvidenceClassSensitive},
		{"policy", &chain.policy, persistence.EventTypePolicyDecision, stageEventID(terminal.ActionID, "requested"), persistence.EvidenceClassSensitive},
		{"started", &chain.started, persistence.EventTypeActionStarted, stageEventID(terminal.ActionID, "policy"), persistence.EvidenceClassSensitive},
	}
	for _, stage := range stages {
		loaded, err := reflectionLoadStage(ctx, store, terminal.ActionID, stage.name)
		if err != nil {
			return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", err
		}
		*stage.dest = loaded
		if err := validateReflectionRelated(terminal, loaded); err != nil {
			return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("%s causal binding: %w", stage.name, err)
		}
		if err := validateReflectionStage(loaded, stage.name, stage.typ, stage.cause, stage.class); err != nil {
			return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", err
		}
	}
	if chain.context.Sequence >= chain.route.Sequence || chain.route.Sequence >= chain.plan.Sequence || chain.plan.Sequence >= chain.requested.Sequence || chain.requested.Sequence >= chain.policy.Sequence || chain.policy.Sequence >= chain.started.Sequence || chain.started.Sequence >= terminal.Sequence {
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("E1 causal stages are not in increasing ledger sequence")
	}

	if err := validateReflectionContext(chain.context); err != nil {
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", err
	}
	var route PlanRouteEvidence
	if err := strictReflectionJSON(chain.route.PayloadJSON, &route); err != nil {
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("decode provider route: %w", err)
	}
	if err := validateReflectionRoutePayload(chain.route, route, true); err != nil {
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", err
	}
	var plan ActionPlan
	if err := strictReflectionJSON(chain.plan.PayloadJSON, &plan); err != nil {
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("decode action plan: %w", err)
	}
	if err := plan.Validate(); err != nil || !finiteUnit(plan.Confidence) {
		if err != nil {
			return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("invalid action plan: %w", err)
		}
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("action plan confidence is not finite and unit-bounded")
	}
	if err := validateReflectionRoutedEnvelope(chain.plan, route); err != nil {
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("plan route binding: %w", err)
	}

	var requested reflectionRequestedPayload
	if err := strictReflectionJSON(chain.requested.PayloadJSON, &requested); err != nil {
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("decode action request: %w", err)
	}
	if requested.Tool != plan.Tool || requested.Purpose != plan.Purpose || requested.RelativePath != plan.RelativePath || requested.RequestedSHA256 != sha256Hex([]byte(plan.Content)) || requested.Reversible != plan.Reversible {
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("action request does not bind the proposed plan")
	}
	if err := validateReflectionRoutedEnvelope(chain.requested, route); err != nil {
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("requested route binding: %w", err)
	}

	var decision ActionPolicyDecision
	if err := strictReflectionJSON(chain.policy.PayloadJSON, &decision); err != nil {
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("decode policy decision: %w", err)
	}
	if decision.PolicyVersion != EnactionPolicyVersion || !decision.Allowed || decision.Mode != EnactionLocalSandbox || strings.TrimSpace(decision.Code) == "" || strings.TrimSpace(decision.Reason) == "" || decision.DecidedAt.IsZero() {
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("policy decision is not an allowed E1 local-sandbox decision")
	}
	if err := validateReflectionUnroutedEnvelope(chain.policy); err != nil {
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("policy envelope: %w", err)
	}

	var started reflectionStartedPayload
	if err := strictReflectionJSON(chain.started.PayloadJSON, &started); err != nil {
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("decode action started: %w", err)
	}
	if started.Tool != plan.Tool || started.RelativePath != plan.RelativePath || started.RequestedSHA256 != sha256Hex([]byte(plan.Content)) {
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("action.started does not bind the proposed plan")
	}
	if err := validateReflectionRoutedEnvelope(chain.started, route); err != nil {
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("started route binding: %w", err)
	}

	originalSession := chain.context.SessionID
	if originalSession == "" || chain.route.SessionID != originalSession || chain.plan.SessionID != originalSession || chain.requested.SessionID != originalSession {
		return reflectionActionChain{}, ActionPlan{}, PlanRouteEvidence{}, "", fmt.Errorf("context, route, plan, and request must share a non-empty original session")
	}
	return chain, plan, route, originalSession, nil
}

func validateReflectionTerminalSessions(originalSession string, policy, started, terminal, evaluation persistence.StoredCognitiveEvent, observation ActionObservation) error {
	if terminal.SessionID == "" {
		return fmt.Errorf("terminal action has no session")
	}
	if terminal.SessionID != originalSession && !observation.Recovered {
		return fmt.Errorf("terminal action changed session without recovery evidence")
	}
	for _, stage := range []struct {
		name    string
		session string
	}{{"policy", policy.SessionID}, {"started", started.SessionID}} {
		if stage.session != originalSession && stage.session != terminal.SessionID {
			return fmt.Errorf("%s session is neither original nor terminal session", stage.name)
		}
	}
	if evaluation.EventID != "" && evaluation.SessionID != terminal.SessionID {
		return fmt.Errorf("evaluation session must match terminal action session")
	}
	return nil
}

func validateReflectionActionBase(event persistence.StoredCognitiveEvent, requireLocalSandboxAction bool) error {
	if event.IdentityID == "" || event.ActionID == "" || event.GoalID == "" || event.SessionID == "" {
		return fmt.Errorf("E1 reflection source has incomplete identity/action/goal/session binding")
	}
	if event.CorrelationID != "enaction:"+event.GoalID {
		return fmt.Errorf("E1 reflection source has wrong correlation ID")
	}
	if event.PolicyVersion != EnactionPolicyVersion {
		return fmt.Errorf("E1 reflection source has policy version %q", event.PolicyVersion)
	}
	mode, err := reflectionActionMode(event.ActionID, event.GoalID)
	if err != nil {
		return err
	}
	if requireLocalSandboxAction && mode != EnactionLocalSandbox {
		return fmt.Errorf("executed E1 action mode must be %q", EnactionLocalSandbox)
	}
	return nil
}

func reflectionActionMode(actionID, goalID string) (EnactionMode, error) {
	prefix := "action:" + goalID + ":" + WorkspaceCreateNote + ":"
	if !strings.HasPrefix(actionID, prefix) {
		return "", fmt.Errorf("action ID does not have the deterministic E1 prefix")
	}
	rest := strings.TrimPrefix(actionID, prefix)
	for _, mode := range []EnactionMode{EnactionLocalSandbox, EnactionObserve} {
		iterationPrefix := string(mode) + ":iteration-"
		if !strings.HasPrefix(rest, iterationPrefix) {
			continue
		}
		iteration, err := strconv.ParseInt(strings.TrimPrefix(rest, iterationPrefix), 10, 64)
		if err != nil || iteration < 1 {
			return "", fmt.Errorf("action ID has an invalid E1 iteration")
		}
		return mode, nil
	}
	return "", fmt.Errorf("action ID does not name an E1 mode")
}

func reflectionLoadStage(ctx context.Context, store *persistence.CognitiveEventStore, actionID, stage string) (persistence.StoredCognitiveEvent, error) {
	event, err := store.Get(ctx, stageEventID(actionID, stage))
	if err != nil {
		return persistence.StoredCognitiveEvent{}, fmt.Errorf("load E1 %s stage: %w", stage, err)
	}
	return event, nil
}

func validateReflectionRelated(root, related persistence.StoredCognitiveEvent) error {
	if related.IdentityID != root.IdentityID || related.ActionID != root.ActionID || related.GoalID != root.GoalID || related.CorrelationID != "enaction:"+root.GoalID || related.PolicyVersion != EnactionPolicyVersion || related.SessionID == "" {
		return fmt.Errorf("identity, action, goal, correlation, policy, or session mismatch")
	}
	return nil
}

func validateReflectionStage(event persistence.StoredCognitiveEvent, stage, eventType, causationID, evidenceClass string) error {
	if event.EventID != stageEventID(event.ActionID, stage) {
		return fmt.Errorf("%s event ID is not its deterministic stage ID", stage)
	}
	if event.EventType != eventType || event.CausationID != causationID || event.IdempotencyKey != event.EventID || event.EvidenceClass != evidenceClass || event.PolicyVersion != EnactionPolicyVersion || event.RouteReason != "" {
		return fmt.Errorf("%s envelope does not match the E1 producer contract", stage)
	}
	return nil
}

func validateReflectionContext(event persistence.StoredCognitiveEvent) error {
	var payload reflectionContextPayload
	if err := strictReflectionJSON(event.PayloadJSON, &payload); err != nil {
		return fmt.Errorf("decode retrieved context: %w", err)
	}
	if strings.TrimSpace(payload.Goal) == "" || payload.RecentContextCount < 0 {
		return fmt.Errorf("invalid retrieved context payload")
	}
	if err := validateReflectionUnroutedEnvelope(event); err != nil {
		return fmt.Errorf("context envelope: %w", err)
	}
	return nil
}

func validateReflectionObservation(observation ActionObservation, plan ActionPlan) error {
	if observation.Tool != plan.Tool || observation.Tool != WorkspaceCreateNote || observation.RelativePath != plan.RelativePath || observation.ContentSHA256 != sha256Hex([]byte(plan.Content)) {
		return fmt.Errorf("tool, path, or content hash mismatch")
	}
	if observation.BytesWritten < 0 {
		return fmt.Errorf("negative bytes written")
	}
	return nil
}

func validateReflectionRoute(event persistence.StoredCognitiveEvent, requireUsableRoute bool) error {
	var route PlanRouteEvidence
	if err := strictReflectionJSON(event.PayloadJSON, &route); err != nil {
		return fmt.Errorf("decode provider route: %w", err)
	}
	return validateReflectionRoutePayload(event, route, requireUsableRoute)
}

func validateReflectionRoutePayload(event persistence.StoredCognitiveEvent, route PlanRouteEvidence, requireUsableRoute bool) error {
	if strings.TrimSpace(route.TraceID) == "" || route.AttemptCount < 0 {
		return fmt.Errorf("invalid provider route payload")
	}
	if err := validateReflectionRoutedEnvelope(event, route); err != nil {
		return err
	}
	if requireUsableRoute && (strings.TrimSpace(route.Provider) == "" || strings.TrimSpace(route.BackendKind) == "" || strings.TrimSpace(route.ModelID) == "" || route.Degraded || route.BackendKind == "fallback") {
		return fmt.Errorf("route is degraded or lacks real provider/backend/model evidence")
	}
	return nil
}

func validateReflectionRoutedEnvelope(event persistence.StoredCognitiveEvent, route PlanRouteEvidence) error {
	if event.Provider != sanitizeIdentifier(route.Provider) || event.BackendKind != sanitizeIdentifier(route.BackendKind) || event.ModelID != sanitizeIdentifier(route.ModelID) || event.Degraded != route.Degraded || event.RouteReason != "" {
		return fmt.Errorf("route envelope differs from producer route evidence")
	}
	return nil
}

func validateReflectionUnroutedEnvelope(event persistence.StoredCognitiveEvent) error {
	if event.Provider != "" || event.BackendKind != "" || event.ModelID != "" || event.RouteReason != "" || event.Degraded {
		return fmt.Errorf("event unexpectedly carries provider route metadata")
	}
	return nil
}

func persistenceErrNotFound(err error) bool {
	return errors.Is(err, persistence.ErrCognitiveEventNotFound)
}
