package deeptreeecho

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/backendcap"
	"github.com/cogpy/echo9llama/core/llm"
	"github.com/cogpy/echo9llama/core/persistence"
	"github.com/cogpy/echo9llama/core/tools"
)

const structuredArtifactSkill = "Structured Artifact Creation"

// ActionPlanner returns the public model output and verifiable route evidence.
// The output is parsed by ParseActionPlanStrict; it is never executed directly.
type ActionPlanner interface {
	GeneratePlan(context.Context, EnactionRequest) (string, PlanRouteEvidence, error)
}

// LLMActionPlanner uses the production capability router with fallback disabled.
type LLMActionPlanner struct {
	Provider llm.LLMProvider
}

func (planner LLMActionPlanner) GeneratePlan(ctx context.Context, request EnactionRequest) (string, PlanRouteEvidence, error) {
	if planner.Provider == nil {
		return "", PlanRouteEvidence{}, fmt.Errorf("planner provider is required")
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", PlanRouteEvidence{}, fmt.Errorf("marshal enaction request: %w", err)
	}
	traceID := "enaction-plan:" + shortHash(request.SessionID+":"+request.GoalID+":"+request.Goal)
	prompt := `You are Deep Tree Echo's bounded action planner. Preserve a confident, witty, charismatic voice, but prioritize truth, uncertainty, autonomy, and non-harm.
Return exactly one JSON object and no prose. The only permitted tool is workspace.create_note, which creates one new private Markdown note and cannot overwrite files.
Required fields: tool, purpose, relative_path, content, confidence, reversible, affected_parties, success_criteria, predicted_effects.
Use a clean relative path under briefs/ ending in .md and include the supplied action_iteration in the filename. Set affected_parties to ["self"]. Set reversible true. Content must be concise (at most 1800 UTF-8 bytes), useful, evidence-aware Markdown grounded in the supplied goal/context; include open questions and avoid claims not supported by context. Keep arrays to at most four short items. Never request shell, network, credentials, external communication, file overwrite, or action affecting another person.

REQUEST:
` + string(requestJSON)
	opts := llm.GenerateOptions{
		MaxTokens:   1600,
		Temperature: 0.35,
		TopP:        0.85,
		Routing: llm.RoutingOptions{
			TraceID:               traceID,
			Intent:                "enaction.plan",
			PreferNative:          false,
			RequireRealModel:      true,
			AllowRemote:           true,
			AllowFallback:         false,
			PolicyExplicit:        true,
			AllowQueue:            true,
			RequiredContextTokens: 2048,
			LatencyClass:          backendcap.LatencyNormal,
			QueueWait:             15 * time.Second,
		},
	}
	raw, generateErr := planner.Provider.Generate(ctx, prompt, opts)
	evidence := PlanRouteEvidence{TraceID: traceID}
	traced, ok := planner.Provider.(interface {
		GetRouteResult(string) (llm.BackendRoutingState, bool)
	})
	if !ok {
		return raw, evidence, fmt.Errorf("planner provider does not expose route evidence")
	}
	state, exists := traced.GetRouteResult(traceID)
	if !exists {
		return raw, evidence, fmt.Errorf("planner route evidence is unavailable")
	}
	evidence.Provider = state.SelectedProvider
	evidence.BackendKind = string(state.SelectedKind)
	evidence.ModelID = state.SelectedModelID
	evidence.RouteReason = state.Decision.Reason
	evidence.Degraded = state.Degraded || state.SelectedKind == backendcap.BackendFallback
	evidence.AttemptCount = len(state.Attempts)
	if generateErr != nil {
		return raw, evidence, generateErr
	}
	if evidence.Provider == "" || evidence.Degraded {
		return raw, evidence, fmt.Errorf("planner route is degraded or did not select a real provider")
	}
	return raw, evidence, nil
}

// EnactionPipeline converts a scheduled goal into a bounded action, observed
// outcome, deterministic evaluation, and durable causal event chain.
type EnactionPipeline struct {
	stepMu sync.Mutex
	mu     sync.RWMutex
	paused bool

	store          *persistence.CognitiveEventStore
	tool           tools.WorkspaceNoteTool
	policy         *ActionPolicy
	planner        ActionPlanner
	identity       string
	session        string
	attemptedGoals map[string]string
}

func NewEnactionPipeline(
	store *persistence.CognitiveEventStore,
	tool tools.WorkspaceNoteTool,
	policy *ActionPolicy,
	planner ActionPlanner,
	identityID, sessionID string,
) (*EnactionPipeline, error) {
	if store == nil || policy == nil || planner == nil {
		return nil, fmt.Errorf("event store, action policy, and planner are required")
	}
	if strings.TrimSpace(identityID) == "" || strings.TrimSpace(sessionID) == "" {
		return nil, fmt.Errorf("identity and session IDs are required")
	}
	return &EnactionPipeline{
		store: store, tool: tool, policy: policy, planner: planner,
		identity: identityID, session: sessionID, attemptedGoals: make(map[string]string),
	}, nil
}

// Pause is a quiescence barrier: after it returns no action can cross the tool boundary.
func (pipeline *EnactionPipeline) Pause() {
	pipeline.stepMu.Lock()
	defer pipeline.stepMu.Unlock()
	pipeline.mu.Lock()
	pipeline.paused = true
	pipeline.mu.Unlock()
}

func (pipeline *EnactionPipeline) Resume() {
	pipeline.mu.Lock()
	pipeline.paused = false
	pipeline.mu.Unlock()
}

func (pipeline *EnactionPipeline) IsPaused() bool {
	pipeline.mu.RLock()
	defer pipeline.mu.RUnlock()
	return pipeline.paused
}

func (pipeline *EnactionPipeline) ResetWakeBudget() {
	pipeline.policy.ResetWakeBudget()
	pipeline.mu.Lock()
	pipeline.attemptedGoals = make(map[string]string)
	pipeline.mu.Unlock()
}

// RunGoal is serialized and replay-aware. The action ID is stable per goal and
// capability, enabling restart reconciliation without repeating side effects.
func (pipeline *EnactionPipeline) RunGoal(ctx context.Context, request EnactionRequest) (EnactionOutcome, error) {
	pipeline.stepMu.Lock()
	defer pipeline.stepMu.Unlock()
	if pipeline.IsPaused() {
		return EnactionOutcome{}, fmt.Errorf("enaction pipeline is paused")
	}
	if err := request.Validate(); err != nil {
		return EnactionOutcome{}, err
	}
	if request.IdentityID != pipeline.identity || request.SessionID != pipeline.session {
		return EnactionOutcome{}, fmt.Errorf("request identity/session does not match pipeline binding")
	}

	mode := pipeline.policy.Snapshot().Mode
	attemptKey := request.GoalID + ":" + string(mode)
	pipeline.mu.Lock()
	actionID := pipeline.attemptedGoals[attemptKey]
	pipeline.mu.Unlock()
	if actionID == "" {
		progressCount, countErr := pipeline.store.Count(ctx, persistence.CognitiveEventQuery{GoalID: request.GoalID, EventType: persistence.EventTypeGoalProgressed})
		if countErr != nil {
			return EnactionOutcome{}, countErr
		}
		actionID = fmt.Sprintf("action:%s:workspace.create_note:%s:iteration-%d", request.GoalID, mode, progressCount+1)
		pipeline.mu.Lock()
		pipeline.attemptedGoals[attemptKey] = actionID
		pipeline.mu.Unlock()
		request.ActionIteration = progressCount + 1
	} else {
		request.ActionIteration = actionIterationFromID(actionID)
	}
	correlationID := "enaction:" + request.GoalID
	outcome := EnactionOutcome{CorrelationID: correlationID, ActionID: actionID, GoalID: request.GoalID}

	events, err := pipeline.store.Query(ctx, persistence.CognitiveEventQuery{ActionID: actionID, Limit: 1000})
	if err != nil {
		return outcome, err
	}
	indexed := indexActionEvents(events)
	if err := pipeline.completeTerminalDream(ctx, indexed, actionID, correlationID, request.GoalID); err != nil {
		return outcome, err
	}
	if _, hadDream := indexed[persistence.EventTypeDreamExperience]; !hadDream {
		events, err = pipeline.store.Query(ctx, persistence.CognitiveEventQuery{ActionID: actionID, Limit: 1000})
		if err != nil {
			return outcome, err
		}
		indexed = indexActionEvents(events)
	}
	if terminal := terminalOutcome(indexed, outcome); terminal != nil {
		return *terminal, nil
	}

	plan, route, err := pipeline.obtainPlan(ctx, request, indexed, actionID, correlationID)
	if err != nil {
		outcome.Degraded = route.Degraded
		outcome.ErrorCategory = classifyEnactionError(err)
		outcome.Summary = "action planning failed closed"
		_ = pipeline.appendError(ctx, actionID, correlationID, request.GoalID, "planning", outcome.ErrorCategory, err)
		_ = pipeline.recordDreamOutcome(ctx, actionID, correlationID, request.GoalID, stageEventID(actionID, "error-planning"), "A proposed action failed closed during planning or route validation.", 0.7, []string{"enaction", "planning", "failure", outcome.ErrorCategory})
		return outcome, err
	}

	decision, err := pipeline.authorize(ctx, plan, indexed, actionID, correlationID, request.GoalID)
	if err != nil {
		outcome.ErrorCategory = classifyEnactionError(err)
		return outcome, err
	}
	outcome.Allowed = decision.Allowed
	if !decision.Allowed {
		outcome.TerminalEventID = stageEventID(actionID, "policy")
		outcome.Summary = "proposal recorded; policy denied tool execution: " + decision.Code
		_ = pipeline.recordDreamOutcome(ctx, actionID, correlationID, request.GoalID, outcome.TerminalEventID, "A proposed action was denied by deterministic policy: "+decision.Code+".", 0.6, []string{"enaction", "denied", decision.Code})
		return outcome, nil
	}

	var observation ActionObservation
	var recovered bool
	if stored, completed := indexed[persistence.EventTypeActionCompleted]; completed {
		if err := json.Unmarshal(stored.PayloadJSON, &observation); err != nil {
			pipeline.policy.RecordOutcome(false, "persisted completion evidence could not be decoded")
			return outcome, fmt.Errorf("decode persisted action completion: %w", err)
		}
		recovered = true
		observation.Recovered = true
	} else {
		observation, recovered, err = pipeline.executeOrReconcile(ctx, plan, indexed, actionID, correlationID, request.GoalID, route)
	}
	outcome.Executed = observation.Created || observation.Recovered
	outcome.Recovered = recovered || observation.Recovered
	outcome.Degraded = route.Degraded
	if err != nil {
		pipeline.policy.RecordOutcome(false, "tool effect was not verified")
		outcome.ErrorCategory = observation.ErrorCategory
		outcome.TerminalEventID = stageEventID(actionID, "failed")
		outcome.Summary = "tool action failed or reconciliation found a conflict"
		_ = pipeline.recordDreamOutcome(ctx, actionID, correlationID, request.GoalID, outcome.TerminalEventID, "A bounded tool action failed or reconciliation detected conflicting evidence.", 0.9, []string{"enaction", "tool", "failure", outcome.ErrorCategory})
		return outcome, err
	}

	evaluation, err := pipeline.recordEvaluation(ctx, plan, observation, actionID, correlationID, request.GoalID)
	if err != nil {
		pipeline.policy.RecordOutcome(false, "evaluation could not be durably recorded")
		return outcome, err
	}
	pipeline.policy.RecordOutcome(evaluation.Passed, evaluation.Reason)
	outcome.Verified = evaluation.Passed
	outcome.Score = evaluation.Score
	outcome.EvaluationEventID = evaluation.EvidenceEventID
	outcome.TerminalEventID = stageEventID(actionID, "completed")

	if evaluation.Passed {
		goalEvent, skillEvent, dreamEvent, projectionErr := pipeline.recordSuccessProjections(ctx, plan, observation, evaluation, actionID, correlationID, request.GoalID, request.Goal)
		if projectionErr != nil {
			return outcome, projectionErr
		}
		outcome.SkillEvidenceEventID = skillEvent
		outcome.Summary = fmt.Sprintf("verified private artifact %s; goal evidence %s; dream evidence %s", observation.RelativePath, goalEvent, dreamEvent)
	} else {
		outcome.Summary = "action completed but deterministic verification failed"
		_ = pipeline.recordDreamOutcome(ctx, actionID, correlationID, request.GoalID, evaluation.EvidenceEventID, "A completed bounded action did not satisfy deterministic verification.", 0.9, []string{"enaction", "evaluation", "failure"})
	}
	return outcome, nil
}

func (pipeline *EnactionPipeline) completeTerminalDream(ctx context.Context, indexed map[string]persistence.StoredCognitiveEvent, actionID, correlationID, goalID string) error {
	if _, exists := indexed[persistence.EventTypeDreamExperience]; exists {
		return nil
	}
	if event, exists := indexed[persistence.EventTypeActionFailed]; exists {
		var observation ActionObservation
		_ = json.Unmarshal(event.PayloadJSON, &observation)
		return pipeline.recordDreamOutcome(ctx, actionID, correlationID, goalID, event.EventID, "A bounded tool action failed or reconciliation detected conflicting evidence.", 0.9, []string{"enaction", "tool", "failure", observation.ErrorCategory})
	}
	if event, exists := indexed[persistence.EventTypeErrorRecorded]; exists {
		var payload struct {
			Category string `json:"category"`
		}
		_ = json.Unmarshal(event.PayloadJSON, &payload)
		return pipeline.recordDreamOutcome(ctx, actionID, correlationID, goalID, event.EventID, "A proposed action failed closed during planning or route validation.", 0.7, []string{"enaction", "planning", "failure", payload.Category})
	}
	if event, exists := indexed[persistence.EventTypePolicyDecision]; exists {
		var decision ActionPolicyDecision
		if json.Unmarshal(event.PayloadJSON, &decision) == nil && !decision.Allowed {
			return pipeline.recordDreamOutcome(ctx, actionID, correlationID, goalID, event.EventID, "A proposed action was denied by deterministic policy: "+decision.Code+".", 0.6, []string{"enaction", "denied", decision.Code})
		}
	}
	if event, exists := indexed[persistence.EventTypeEvaluationRecorded]; exists {
		var evaluation ActionEvaluation
		if json.Unmarshal(event.PayloadJSON, &evaluation) == nil && !evaluation.Passed {
			return pipeline.recordDreamOutcome(ctx, actionID, correlationID, goalID, event.EventID, "A completed bounded action did not satisfy deterministic verification.", 0.9, []string{"enaction", "evaluation", "failure"})
		}
	}
	return nil
}

func (pipeline *EnactionPipeline) obtainPlan(ctx context.Context, request EnactionRequest, indexed map[string]persistence.StoredCognitiveEvent, actionID, correlationID string) (ActionPlan, PlanRouteEvidence, error) {
	if stored, ok := indexed[persistence.EventTypePlanProposed]; ok {
		var plan ActionPlan
		if err := json.Unmarshal(stored.PayloadJSON, &plan); err != nil {
			return ActionPlan{}, PlanRouteEvidence{}, fmt.Errorf("decode persisted action plan: %w", err)
		}
		if err := plan.Validate(); err != nil {
			return ActionPlan{}, PlanRouteEvidence{}, fmt.Errorf("persisted action plan is invalid: %w", err)
		}
		route := routeFromEvents(indexed)
		if _, requested := indexed[persistence.EventTypeActionRequested]; !requested {
			requestPayload := map[string]any{
				"tool": plan.Tool, "purpose": plan.Purpose, "relative_path": plan.RelativePath,
				"requested_sha256": sha256Hex([]byte(plan.Content)), "reversible": plan.Reversible,
			}
			if _, err := pipeline.appendEvent(ctx, stageEventID(actionID, "requested"), persistence.EventTypeActionRequested, correlationID, stored.EventID, request.GoalID, actionID, persistence.EvidenceClassSensitive, requestPayload, route); err != nil {
				return ActionPlan{}, route, err
			}
		}
		return plan, route, nil
	}

	contextPayload := map[string]any{
		"goal":                 request.Goal,
		"top_interests":        append([]string(nil), request.TopInterests...),
		"recent_context_count": len(request.RecentContext),
	}
	if _, err := pipeline.appendEvent(ctx, stageEventID(actionID, "context"), persistence.EventTypeContextRetrieved, correlationID, "", request.GoalID, actionID, persistence.EvidenceClassSensitive, contextPayload, PlanRouteEvidence{}); err != nil {
		return ActionPlan{}, PlanRouteEvidence{}, err
	}

	raw, route, err := pipeline.planner.GeneratePlan(ctx, request)
	if route.TraceID != "" {
		_, routeErr := pipeline.appendEvent(ctx, stageEventID(actionID, "route"), persistence.EventTypeProviderRouted, correlationID, stageEventID(actionID, "context"), request.GoalID, actionID, persistence.EvidenceClassOperational, route, route)
		if routeErr != nil {
			return ActionPlan{}, route, routeErr
		}
	}
	if err != nil {
		return ActionPlan{}, route, err
	}
	if route.Degraded || strings.TrimSpace(route.Provider) == "" || route.BackendKind == string(backendcap.BackendFallback) {
		return ActionPlan{}, route, fmt.Errorf("degraded or unverifiable model output cannot authorize action")
	}
	plan, err := ParseActionPlanStrict(raw)
	if err != nil {
		return ActionPlan{}, route, err
	}
	if _, err := pipeline.appendEvent(ctx, stageEventID(actionID, "plan"), persistence.EventTypePlanProposed, correlationID, stageEventID(actionID, "route"), request.GoalID, actionID, persistence.EvidenceClassRestricted, plan, route); err != nil {
		return ActionPlan{}, route, err
	}
	requestPayload := map[string]any{
		"tool": plan.Tool, "purpose": plan.Purpose, "relative_path": plan.RelativePath,
		"requested_sha256": sha256Hex([]byte(plan.Content)), "reversible": plan.Reversible,
	}
	if _, err := pipeline.appendEvent(ctx, stageEventID(actionID, "requested"), persistence.EventTypeActionRequested, correlationID, stageEventID(actionID, "plan"), request.GoalID, actionID, persistence.EvidenceClassSensitive, requestPayload, route); err != nil {
		return ActionPlan{}, route, err
	}
	return plan, route, nil
}

func (pipeline *EnactionPipeline) authorize(ctx context.Context, plan ActionPlan, indexed map[string]persistence.StoredCognitiveEvent, actionID, correlationID, goalID string) (ActionPolicyDecision, error) {
	if stored, ok := indexed[persistence.EventTypePolicyDecision]; ok {
		var prior ActionPolicyDecision
		if err := json.Unmarshal(stored.PayloadJSON, &prior); err != nil {
			return ActionPolicyDecision{}, fmt.Errorf("decode persisted policy decision: %w", err)
		}
		if !prior.Allowed {
			return prior, nil
		}
		// A previously allowed but incomplete action must satisfy current policy too.
		current := pipeline.policy.Decide(plan)
		if !current.Allowed {
			return current, fmt.Errorf("current policy no longer authorizes recovered action: %s", current.Code)
		}
		return current, nil
	}
	decision := pipeline.policy.Decide(plan)
	_, err := pipeline.appendEvent(ctx, stageEventID(actionID, "policy"), persistence.EventTypePolicyDecision, correlationID, stageEventID(actionID, "requested"), goalID, actionID, persistence.EvidenceClassSensitive, decision, PlanRouteEvidence{})
	return decision, err
}

func (pipeline *EnactionPipeline) executeOrReconcile(ctx context.Context, plan ActionPlan, indexed map[string]persistence.StoredCognitiveEvent, actionID, correlationID, goalID string, route PlanRouteEvidence) (ActionObservation, bool, error) {
	request := tools.WorkspaceNoteRequest{Path: plan.RelativePath, Content: []byte(plan.Content)}
	if _, started := indexed[persistence.EventTypeActionStarted]; started {
		reconciled, err := pipeline.tool.ReconcileContext(ctx, request)
		if err != nil {
			observation := observationFromTool(plan, reconciled, false, err)
			_ = pipeline.appendFailure(ctx, actionID, correlationID, goalID, observation, route)
			return observation, true, err
		}
		switch reconciled.Status {
		case tools.WorkspaceNoteMatchingExisting:
			observation := observationFromTool(plan, reconciled, true, nil)
			if err := pipeline.appendCompletion(ctx, actionID, correlationID, goalID, observation, route); err != nil {
				return observation, true, err
			}
			return observation, true, nil
		case tools.WorkspaceNoteAbsentRetry:
			// Exactly one replay retry is allowed by the tool reconciliation contract.
		case tools.WorkspaceNoteHashConflict:
			err := fmt.Errorf("workspace reconciliation hash conflict")
			observation := observationFromTool(plan, reconciled, true, err)
			_ = pipeline.appendFailure(ctx, actionID, correlationID, goalID, observation, route)
			return observation, true, err
		default:
			return ActionObservation{}, true, fmt.Errorf("unknown reconciliation status %q", reconciled.Status)
		}
	} else {
		startPayload := map[string]any{"tool": plan.Tool, "relative_path": plan.RelativePath, "requested_sha256": sha256Hex([]byte(plan.Content))}
		if _, err := pipeline.appendEvent(ctx, stageEventID(actionID, "started"), persistence.EventTypeActionStarted, correlationID, stageEventID(actionID, "policy"), goalID, actionID, persistence.EvidenceClassSensitive, startPayload, route); err != nil {
			return ActionObservation{}, false, err
		}
	}

	start := time.Now()
	toolObservation, err := pipeline.tool.Create(ctx, request)
	observation := observationFromTool(plan, toolObservation, false, err)
	observation.Duration = time.Since(start)
	if err != nil {
		_ = pipeline.appendFailure(ctx, actionID, correlationID, goalID, observation, route)
		return observation, false, err
	}
	if err := pipeline.appendCompletion(ctx, actionID, correlationID, goalID, observation, route); err != nil {
		return observation, false, err
	}
	return observation, false, nil
}

func (pipeline *EnactionPipeline) recordEvaluation(ctx context.Context, plan ActionPlan, observation ActionObservation, actionID, correlationID, goalID string) (ActionEvaluation, error) {
	criteria := []string{
		"tool is workspace.create_note",
		"target remained within private workspace",
		"create/reconcile reported no conflict",
		"durable read-back SHA-256 matches requested content",
	}
	passed := observation.Tool == WorkspaceCreateNote && !observation.Conflict && (observation.Created || observation.Recovered) && observation.ObservedSHA256 != "" && observation.ObservedSHA256 == observation.ContentSHA256
	score := 0.0
	reason := "deterministic artifact verification failed"
	if passed {
		score = 1
		reason = "atomic create-only publication and exact read-back hash verification passed"
	}
	evaluation := ActionEvaluation{Passed: passed, Score: score, Criteria: criteria, EvidenceEventID: stageEventID(actionID, "evaluation"), Reason: reason}
	_, err := pipeline.appendEvent(ctx, evaluation.EvidenceEventID, persistence.EventTypeEvaluationRecorded, correlationID, stageEventID(actionID, "completed"), goalID, actionID, persistence.EvidenceClassSensitive, evaluation, PlanRouteEvidence{})
	return evaluation, err
}

func (pipeline *EnactionPipeline) recordSuccessProjections(ctx context.Context, plan ActionPlan, observation ActionObservation, evaluation ActionEvaluation, actionID, correlationID, goalID, goalDescription string) (string, string, string, error) {
	goalEvent := stageEventID(actionID, "goal-progress")
	goalPayload := map[string]any{"delta": 0.1, "score": evaluation.Score, "evidence_event_id": evaluation.EvidenceEventID, "artifact_sha256": observation.ObservedSHA256, "description": goalDescription}
	if _, err := pipeline.appendEvent(ctx, goalEvent, persistence.EventTypeGoalProgressed, correlationID, evaluation.EvidenceEventID, goalID, actionID, persistence.EvidenceClassOperational, goalPayload, PlanRouteEvidence{}); err != nil {
		return "", "", "", err
	}
	skillEvent := stageEventID(actionID, "skill-evidence")
	skillPayload := map[string]any{"skill": structuredArtifactSkill, "score": evaluation.Score, "passed": evaluation.Passed, "evidence_event_id": evaluation.EvidenceEventID, "feedback": evaluation.Reason}
	if _, err := pipeline.appendEvent(ctx, skillEvent, persistence.EventTypeSkillEvidence, correlationID, goalEvent, goalID, actionID, persistence.EvidenceClassOperational, skillPayload, PlanRouteEvidence{}); err != nil {
		return "", "", "", err
	}
	dreamEvent := stageEventID(actionID, "dream")
	dreamPayload := map[string]any{
		"source_event_id": evaluation.EvidenceEventID,
		"importance":      0.8,
		"summary":         fmt.Sprintf("Verified a bounded %s effect at %s with SHA-256 evidence %s for goal %s.", plan.Tool, observation.RelativePath, observation.ObservedSHA256, goalID),
		"tags":            []string{"enaction", "verified", "workspace", "skill-practice"},
	}
	if _, err := pipeline.appendEvent(ctx, dreamEvent, persistence.EventTypeDreamExperience, correlationID, skillEvent, goalID, actionID, persistence.EvidenceClassSensitive, dreamPayload, PlanRouteEvidence{}); err != nil {
		return "", "", "", err
	}
	return goalEvent, skillEvent, dreamEvent, nil
}

func (pipeline *EnactionPipeline) recordDreamOutcome(ctx context.Context, actionID, correlationID, goalID, sourceEventID, summary string, importance float64, tags []string) error {
	dreamPayload := map[string]any{
		"source_event_id": sourceEventID,
		"importance":      clampUnit(importance),
		"summary":         summary,
		"tags":            append([]string(nil), tags...),
	}
	_, err := pipeline.appendEvent(ctx, stageEventID(actionID, "dream"), persistence.EventTypeDreamExperience, correlationID, sourceEventID, goalID, actionID, persistence.EvidenceClassOperational, dreamPayload, PlanRouteEvidence{})
	return err
}

func (pipeline *EnactionPipeline) appendCompletion(ctx context.Context, actionID, correlationID, goalID string, observation ActionObservation, route PlanRouteEvidence) error {
	_, err := pipeline.appendEvent(ctx, stageEventID(actionID, "completed"), persistence.EventTypeActionCompleted, correlationID, stageEventID(actionID, "started"), goalID, actionID, persistence.EvidenceClassSensitive, observation, route)
	return err
}

func (pipeline *EnactionPipeline) appendFailure(ctx context.Context, actionID, correlationID, goalID string, observation ActionObservation, route PlanRouteEvidence) error {
	_, err := pipeline.appendEvent(ctx, stageEventID(actionID, "failed"), persistence.EventTypeActionFailed, correlationID, stageEventID(actionID, "started"), goalID, actionID, persistence.EvidenceClassSensitive, observation, route)
	return err
}

func (pipeline *EnactionPipeline) appendError(ctx context.Context, actionID, correlationID, goalID, stage, category string, cause error) error {
	payload := map[string]any{"stage": stage, "category": category, "message": safeErrorMessage(cause)}
	_, err := pipeline.appendEvent(ctx, stageEventID(actionID, "error-"+stage), persistence.EventTypeErrorRecorded, correlationID, "", goalID, actionID, persistence.EvidenceClassOperational, payload, PlanRouteEvidence{})
	return err
}

func (pipeline *EnactionPipeline) appendEvent(ctx context.Context, eventID, eventType, correlationID, causationID, goalID, actionID, evidenceClass string, payload any, route PlanRouteEvidence) (persistence.DuplicateCognitiveEventResult, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return persistence.DuplicateCognitiveEventResult{}, fmt.Errorf("marshal cognitive event payload: %w", err)
	}
	event := persistence.NewCognitiveEvent(eventID, eventType, pipeline.identity, pipeline.session, EnactionPolicyVersion, payloadJSON)
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
	if stored, getErr := pipeline.store.Get(ctx, eventID); getErr == nil {
		return reusePersistedStage(stored, event)
	} else if !errors.Is(getErr, persistence.ErrCognitiveEventNotFound) {
		return persistence.DuplicateCognitiveEventResult{}, getErr
	}
	result, appendErr := pipeline.store.Append(ctx, event)
	if appendErr == nil {
		return result, nil
	}
	// A commit may have succeeded even if the caller observed a later error. A
	// second read distinguishes that uncertainty from conflicting material.
	if stored, getErr := pipeline.store.Get(ctx, eventID); getErr == nil {
		return reusePersistedStage(stored, event)
	}
	return persistence.DuplicateCognitiveEventResult{}, appendErr
}

func reusePersistedStage(stored persistence.StoredCognitiveEvent, intended persistence.CognitiveEvent) (persistence.DuplicateCognitiveEventResult, error) {
	// Occurrence time and process session belong to the first durable append.
	// Reuse them when checking a retry from a later process session.
	intended.OccurredAt = stored.OccurredAt
	intended.SessionID = stored.SessionID
	hash, err := intended.CanonicalHash()
	if err != nil {
		return persistence.DuplicateCognitiveEventResult{}, err
	}
	if hash != stored.ContentSHA256 {
		return persistence.DuplicateCognitiveEventResult{}, &persistence.CognitiveEventConflictError{Kind: "event_id", Key: intended.EventID}
	}
	return persistence.DuplicateCognitiveEventResult{Sequence: stored.Sequence, Duplicate: true}, nil
}

func indexActionEvents(events []persistence.StoredCognitiveEvent) map[string]persistence.StoredCognitiveEvent {
	indexed := make(map[string]persistence.StoredCognitiveEvent, len(events))
	for _, event := range events {
		indexed[event.EventType] = event
	}
	return indexed
}

func terminalOutcome(indexed map[string]persistence.StoredCognitiveEvent, base EnactionOutcome) *EnactionOutcome {
	if event, ok := indexed[persistence.EventTypeEvaluationRecorded]; ok {
		var evaluation ActionEvaluation
		if json.Unmarshal(event.PayloadJSON, &evaluation) == nil {
			_, dreamProjected := indexed[persistence.EventTypeDreamExperience]
			if !dreamProjected {
				return nil
			}
			if evaluation.Passed {
				_, goalProjected := indexed[persistence.EventTypeGoalProgressed]
				_, skillProjected := indexed[persistence.EventTypeSkillEvidence]
				if !goalProjected || !skillProjected {
					return nil
				}
			}
			base.Allowed = true
			base.Executed = true
			base.Verified = evaluation.Passed
			base.Recovered = true
			base.Score = evaluation.Score
			base.EvaluationEventID = event.EventID
			base.TerminalEventID = event.EventID
			base.Summary = "replayed verified terminal action outcome"
			return &base
		}
	}
	if event, ok := indexed[persistence.EventTypeActionFailed]; ok {
		base.Allowed = true
		base.Recovered = true
		base.TerminalEventID = event.EventID
		base.ErrorCategory = "recorded_action_failure"
		base.Summary = "replayed terminal action failure"
		return &base
	}
	if event, ok := indexed[persistence.EventTypeErrorRecorded]; ok {
		var payload struct {
			Category string `json:"category"`
		}
		_ = json.Unmarshal(event.PayloadJSON, &payload)
		base.Recovered = true
		base.TerminalEventID = event.EventID
		base.ErrorCategory = payload.Category
		base.Summary = "replayed terminal enaction error"
		return &base
	}
	if event, ok := indexed[persistence.EventTypePolicyDecision]; ok {
		var decision ActionPolicyDecision
		if json.Unmarshal(event.PayloadJSON, &decision) == nil && !decision.Allowed {
			base.Recovered = true
			base.TerminalEventID = event.EventID
			base.Summary = "replayed terminal policy denial: " + decision.Code
			return &base
		}
	}
	return nil
}

func routeFromEvents(indexed map[string]persistence.StoredCognitiveEvent) PlanRouteEvidence {
	if event, ok := indexed[persistence.EventTypeProviderRouted]; ok {
		var route PlanRouteEvidence
		_ = json.Unmarshal(event.PayloadJSON, &route)
		return route
	}
	return PlanRouteEvidence{}
}

func observationFromTool(plan ActionPlan, observed tools.WorkspaceNoteObservation, recovered bool, err error) ActionObservation {
	observation := ActionObservation{
		Tool: WorkspaceCreateNote, RelativePath: plan.RelativePath,
		ContentSHA256: observed.RequestedSHA256, ObservedSHA256: observed.SHA256,
		BytesWritten: observed.Bytes, Created: observed.Created, Recovered: recovered || observed.Status == tools.WorkspaceNoteMatchingExisting,
		Conflict: observed.Status == tools.WorkspaceNoteHashConflict || observed.Blocked,
	}
	if observation.ContentSHA256 == "" {
		observation.ContentSHA256 = sha256Hex([]byte(plan.Content))
	}
	if err != nil {
		observation.ErrorCategory = classifyEnactionError(err)
		observation.ErrorMessage = safeErrorMessage(err)
	}
	return observation
}

func stageEventID(actionID, stage string) string {
	candidate := actionID + ":" + stage
	if len(candidate) <= 240 {
		return candidate
	}
	return "event:" + shortHash(candidate) + ":" + stage
}

func shortHash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:16])
}

func sha256Hex(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func sanitizeIdentifier(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var result strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', strings.ContainsRune("._:@/-", r):
			result.WriteRune(r)
		default:
			result.WriteByte('_')
		}
	}
	trimmed := strings.Trim(result.String(), "_")
	if len(trimmed) > 255 {
		trimmed = trimmed[:255]
	}
	return trimmed
}

func classifyEnactionError(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline"
	case errors.Is(err, tools.ErrAlreadyExists):
		return "already_exists"
	case errors.Is(err, tools.ErrSymlink), errors.Is(err, tools.ErrPathTraversal), errors.Is(err, tools.ErrInvalidRelativePath):
		return "workspace_boundary"
	case errors.Is(err, tools.ErrTooLarge):
		return "artifact_too_large"
	default:
		return "enaction"
	}
}

func safeErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	message := strings.ReplaceAll(err.Error(), "\n", " ")
	if len(message) > 240 {
		message = message[:240]
	}
	return message
}

func actionIterationFromID(actionID string) int64 {
	marker := ":iteration-"
	index := strings.LastIndex(actionID, marker)
	if index < 0 {
		return 1
	}
	value, err := strconv.ParseInt(actionID[index+len(marker):], 10, 64)
	if err != nil || value < 1 {
		return 1
	}
	return value
}
