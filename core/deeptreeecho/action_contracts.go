package deeptreeecho

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

const (
	EnactionPolicyVersion = "e1.0"
	WorkspaceCreateNote   = "workspace.create_note"
)

// EnactionMode controls whether model-proposed actions may cross the local tool
// boundary. Observe mode is deliberately the production default.
type EnactionMode string

const (
	EnactionObserve      EnactionMode = "observe"
	EnactionLocalSandbox EnactionMode = "local-sandbox"
)

func ParseEnactionMode(value string) EnactionMode {
	switch EnactionMode(strings.ToLower(strings.TrimSpace(value))) {
	case EnactionLocalSandbox:
		return EnactionLocalSandbox
	default:
		return EnactionObserve
	}
}

// ActionPlan is the only structured model output accepted by the E1 enaction
// boundary. Unknown fields, trailing data, unsupported tools, and unsafe paths
// are rejected before policy evaluation.
type ActionPlan struct {
	Tool             string   `json:"tool"`
	Purpose          string   `json:"purpose"`
	RelativePath     string   `json:"relative_path"`
	Content          string   `json:"content"`
	Confidence       float64  `json:"confidence"`
	Reversible       bool     `json:"reversible"`
	AffectedParties  []string `json:"affected_parties"`
	SuccessCriteria  []string `json:"success_criteria"`
	PredictedEffects []string `json:"predicted_effects"`
}

// PlanRouteEvidence records the selected inference substrate without exposing
// prompts, credentials, or private model paths.
type PlanRouteEvidence struct {
	TraceID      string `json:"trace_id"`
	Provider     string `json:"provider"`
	BackendKind  string `json:"backend_kind"`
	ModelID      string `json:"model_id"`
	RouteReason  string `json:"route_reason"`
	Degraded     bool   `json:"degraded"`
	AttemptCount int    `json:"attempt_count"`
}

// ActionObservation is an immutable description of an attempted tool effect.
type ActionObservation struct {
	Tool           string        `json:"tool"`
	RelativePath   string        `json:"relative_path"`
	ContentSHA256  string        `json:"content_sha256"`
	ObservedSHA256 string        `json:"observed_sha256,omitempty"`
	BytesWritten   int64         `json:"bytes_written"`
	Created        bool          `json:"created"`
	Recovered      bool          `json:"recovered"`
	Conflict       bool          `json:"conflict"`
	Duration       time.Duration `json:"duration"`
	ErrorCategory  string        `json:"error_category,omitempty"`
	ErrorMessage   string        `json:"error_message,omitempty"`
}

// ActionEvaluation is produced only by deterministic verification code. Model
// confidence and self-assessment never determine Passed or Score.
type ActionEvaluation struct {
	Passed          bool     `json:"passed"`
	Score           float64  `json:"score"`
	Criteria        []string `json:"criteria"`
	EvidenceEventID string   `json:"evidence_event_id"`
	Reason          string   `json:"reason"`
}

// EnactionRequest carries a bounded goal and only explicit, inspectable context
// into the planner. It intentionally excludes hidden chain-of-thought.
type EnactionRequest struct {
	IdentityID      string   `json:"identity_id"`
	SessionID       string   `json:"session_id"`
	GoalID          string   `json:"goal_id"`
	Goal            string   `json:"goal"`
	ActionIteration int64    `json:"action_iteration"`
	TopInterests    []string `json:"top_interests"`
	RecentContext   []string `json:"recent_context"`
}

// EnactionOutcome is the terminal result returned to the orchestrator.
type EnactionOutcome struct {
	CorrelationID        string
	ActionID             string
	GoalID               string
	TerminalEventID      string
	EvaluationEventID    string
	SkillEvidenceEventID string
	Allowed              bool
	Executed             bool
	Verified             bool
	Recovered            bool
	Degraded             bool
	Score                float64
	ErrorCategory        string
	Summary              string
}

func ParseActionPlanStrict(raw string) (ActionPlan, error) {
	var plan ActionPlan
	normalized, err := normalizeActionPlanEnvelope(raw)
	if err != nil {
		return ActionPlan{}, err
	}
	decoder := json.NewDecoder(bytes.NewBufferString(normalized))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&plan); err != nil {
		return ActionPlan{}, fmt.Errorf("decode action plan: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return ActionPlan{}, fmt.Errorf("decode action plan: trailing JSON value")
		}
		return ActionPlan{}, fmt.Errorf("decode action plan: trailing data: %w", err)
	}
	if err := plan.Validate(); err != nil {
		return ActionPlan{}, err
	}
	return plan, nil
}

func normalizeActionPlanEnvelope(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if !strings.HasPrefix(value, "```") {
		return value, nil
	}
	firstLine := strings.IndexByte(value, '\n')
	if firstLine < 0 {
		return "", fmt.Errorf("decode action plan: incomplete JSON fence")
	}
	language := strings.TrimSpace(strings.TrimPrefix(value[:firstLine], "```"))
	if language != "" && !strings.EqualFold(language, "json") {
		return "", fmt.Errorf("decode action plan: unsupported fence language %q", language)
	}
	if !strings.HasSuffix(value, "```") {
		return "", fmt.Errorf("decode action plan: incomplete JSON fence")
	}
	inner := strings.TrimSpace(strings.TrimSuffix(value[firstLine+1:], "```"))
	if strings.Contains(inner, "```") {
		return "", fmt.Errorf("decode action plan: multiple code fences are not allowed")
	}
	return inner, nil
}

func (plan ActionPlan) Validate() error {
	if strings.TrimSpace(plan.Tool) != WorkspaceCreateNote {
		return fmt.Errorf("unsupported tool %q", plan.Tool)
	}
	if strings.TrimSpace(plan.Purpose) == "" {
		return fmt.Errorf("purpose is required")
	}
	if strings.TrimSpace(plan.Content) == "" {
		return fmt.Errorf("content is required")
	}
	if plan.Confidence < 0 || plan.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}
	if filepath.IsAbs(plan.RelativePath) {
		return fmt.Errorf("relative_path must be relative")
	}
	cleaned := filepath.Clean(strings.TrimSpace(plan.RelativePath))
	if cleaned == "." || cleaned == "" || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return fmt.Errorf("relative_path escapes the workspace")
	}
	if !strings.HasPrefix(filepath.ToSlash(cleaned), "briefs/") || !strings.EqualFold(filepath.Ext(cleaned), ".md") {
		return fmt.Errorf("relative_path must name a Markdown file beneath briefs/")
	}
	if len(plan.AffectedParties) == 0 {
		return fmt.Errorf("affected_parties is required")
	}
	if len(plan.SuccessCriteria) == 0 {
		return fmt.Errorf("success_criteria is required")
	}
	return nil
}

func (request EnactionRequest) Validate() error {
	if strings.TrimSpace(request.IdentityID) == "" {
		return fmt.Errorf("identity_id is required")
	}
	if strings.TrimSpace(request.SessionID) == "" {
		return fmt.Errorf("session_id is required")
	}
	if strings.TrimSpace(request.GoalID) == "" || strings.TrimSpace(request.Goal) == "" {
		return fmt.Errorf("goal identity and description are required")
	}
	return nil
}

func clampUnit(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
