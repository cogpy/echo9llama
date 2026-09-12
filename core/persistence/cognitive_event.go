package persistence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

// CognitiveEventSchemaVersion is the current version of the durable event
// envelope. Schema changes must introduce a new version and migration.
const CognitiveEventSchemaVersion = 1

// Stable event types. Event payloads record externally observable inputs,
// decisions, actions, and outcomes; they must not contain hidden reasoning.
const (
	EventTypeSessionStarted      = "session.started"
	EventTypeSessionEnded        = "session.ended"
	EventTypeObservationRecorded = "observation.recorded"
	EventTypeMemoryRecorded      = "memory.recorded"
	EventTypeGoalCreated         = "goal.created"
	EventTypeGoalCompleted       = "goal.completed"
	EventTypeActionRequested     = "action.requested"
	EventTypeActionStarted       = "action.started"
	EventTypeActionCompleted     = "action.completed"
	EventTypeActionFailed        = "action.failed"
	EventTypePolicyDecision      = "policy.decision"
	EventTypeProviderRouted      = "provider.routed"
	EventTypeEvidenceAttached    = "evidence.attached"
	EventTypeContextRetrieved    = "context.retrieved"
	EventTypePlanProposed        = "plan.proposed"
	EventTypeEvaluationRecorded  = "evaluation.recorded"
	EventTypeGoalProgressed      = "goal.progressed"
	EventTypeSkillEvidence       = "skill.evidence_recorded"
	EventTypeDreamExperience     = "dream.experience_queued"
	EventTypeCoreObserved        = "cognitive_core.observed"
	EventTypeErrorRecorded       = "error.recorded"
	EventTypeDegradedModeEntered = "degraded_mode.entered"
	EventTypeDegradedModeExited  = "degraded_mode.exited"
)

// Stable evidence classes, ordered from least to most restricted. The class
// describes handling requirements, not the truth or quality of an event.
const (
	EvidenceClassPublic      = "public"
	EvidenceClassOperational = "operational"
	EvidenceClassSensitive   = "sensitive"
	EvidenceClassRestricted  = "restricted"
)

var (
	// ErrInvalidCognitiveEvent wraps an event whose immutable envelope is not
	// safe or complete enough to persist.
	ErrInvalidCognitiveEvent = errors.New("invalid cognitive event")
	// ErrUnsupportedCognitiveEventSchemaVersion identifies an envelope version
	// for which this binary has no canonicalization or migration contract.
	ErrUnsupportedCognitiveEventSchemaVersion = errors.New("unsupported cognitive event schema version")
	// ErrInvalidCognitiveEventType identifies a non-stable event type.
	ErrInvalidCognitiveEventType = errors.New("invalid cognitive event type")
	// ErrInvalidEvidenceClass identifies an unsupported evidence class.
	ErrInvalidEvidenceClass = errors.New("invalid evidence class")
	// ErrInvalidCognitiveEventPayload identifies malformed, non-object, or
	// hidden-reasoning-bearing JSON payloads.
	ErrInvalidCognitiveEventPayload = errors.New("invalid cognitive event payload")
	// ErrCognitiveEventConflict means event_id or idempotency_key is already
	// associated with different immutable material.
	ErrCognitiveEventConflict = errors.New("cognitive event conflict")
	// ErrCognitiveEventNotFound is returned for absent durable event records.
	ErrCognitiveEventNotFound = errors.New("cognitive event not found")
	// ErrProjectionCursorNotFound is returned for an uninitialized projection.
	ErrProjectionCursorNotFound = errors.New("projection cursor not found")
	// ErrProjectionCursorRegression forbids moving a durable cursor backwards.
	ErrProjectionCursorRegression = errors.New("projection cursor regression")
)

var identifierPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:@/-]{0,255}$`)

var stableEventTypes = map[string]struct{}{
	EventTypeSessionStarted:      {},
	EventTypeSessionEnded:        {},
	EventTypeObservationRecorded: {},
	EventTypeMemoryRecorded:      {},
	EventTypeGoalCreated:         {},
	EventTypeGoalCompleted:       {},
	EventTypeActionRequested:     {},
	EventTypeActionStarted:       {},
	EventTypeActionCompleted:     {},
	EventTypeActionFailed:        {},
	EventTypePolicyDecision:      {},
	EventTypeProviderRouted:      {},
	EventTypeEvidenceAttached:    {},
	EventTypeContextRetrieved:    {},
	EventTypePlanProposed:        {},
	EventTypeEvaluationRecorded:  {},
	EventTypeGoalProgressed:      {},
	EventTypeSkillEvidence:       {},
	EventTypeDreamExperience:     {},
	EventTypeCoreObserved:        {},
	EventTypeErrorRecorded:       {},
	EventTypeDegradedModeEntered: {},
	EventTypeDegradedModeExited:  {},
}

var stableEvidenceClasses = map[string]struct{}{
	EvidenceClassPublic:      {},
	EvidenceClassOperational: {},
	EvidenceClassSensitive:   {},
	EvidenceClassRestricted:  {},
}

// CognitiveEvent is the versioned, immutable persistence envelope for an
// externally observable cognitive-system event. PayloadJSON deliberately has
// no field for chain-of-thought or hidden provider reasoning.
type CognitiveEvent struct {
	SchemaVersion  int             `json:"schema_version"`
	EventID        string          `json:"event_id"`
	EventType      string          `json:"event_type"`
	OccurredAt     time.Time       `json:"occurred_at"`
	IdentityID     string          `json:"identity_id"`
	SessionID      string          `json:"session_id"`
	CorrelationID  string          `json:"correlation_id,omitempty"`
	CausationID    string          `json:"causation_id,omitempty"`
	GoalID         string          `json:"goal_id,omitempty"`
	ActionID       string          `json:"action_id,omitempty"`
	IdempotencyKey string          `json:"idempotency_key,omitempty"`
	EvidenceClass  string          `json:"evidence_class"`
	PayloadJSON    json.RawMessage `json:"payload_json"`
	Provider       string          `json:"provider,omitempty"`
	BackendKind    string          `json:"backend_kind,omitempty"`
	ModelID        string          `json:"model_id,omitempty"`
	RouteReason    string          `json:"route_reason,omitempty"`
	Degraded       bool            `json:"degraded"`
	PolicyVersion  string          `json:"policy_version"`
	ContentSHA256  string          `json:"content_sha256"`
}

// Validate checks envelope semantics, canonical payload validity, and an
// optional supplied content hash. It does not mutate the event.
func (e CognitiveEvent) Validate() error {
	if e.SchemaVersion != CognitiveEventSchemaVersion {
		return fmt.Errorf("%w: got %d, want %d", ErrUnsupportedCognitiveEventSchemaVersion, e.SchemaVersion, CognitiveEventSchemaVersion)
	}
	for field, value := range map[string]string{
		"event_id":       e.EventID,
		"identity_id":    e.IdentityID,
		"session_id":     e.SessionID,
		"policy_version": e.PolicyVersion,
	} {
		if err := validateRequiredIdentifier(field, value); err != nil {
			return err
		}
	}
	for field, value := range map[string]string{
		"correlation_id":  e.CorrelationID,
		"causation_id":    e.CausationID,
		"goal_id":         e.GoalID,
		"action_id":       e.ActionID,
		"idempotency_key": e.IdempotencyKey,
		"provider":        e.Provider,
		"backend_kind":    e.BackendKind,
		"model_id":        e.ModelID,
		"route_reason":    e.RouteReason,
	} {
		if err := validateOptionalIdentifier(field, value); err != nil {
			return err
		}
	}
	if e.OccurredAt.IsZero() {
		return fmt.Errorf("%w: occurred_at is required", ErrInvalidCognitiveEvent)
	}
	if _, ok := stableEventTypes[e.EventType]; !ok {
		return fmt.Errorf("%w: %q", ErrInvalidCognitiveEventType, e.EventType)
	}
	if _, ok := stableEvidenceClasses[e.EvidenceClass]; !ok {
		return fmt.Errorf("%w: %q", ErrInvalidEvidenceClass, e.EvidenceClass)
	}
	if _, err := canonicalPayloadJSON(e.PayloadJSON); err != nil {
		return err
	}
	if e.ContentSHA256 != "" {
		hash, err := e.CanonicalHash()
		if err != nil {
			return err
		}
		if e.ContentSHA256 != hash {
			return fmt.Errorf("%w: content_sha256 does not match canonical event", ErrInvalidCognitiveEvent)
		}
	}
	return nil
}

// CanonicalBytes returns the deterministic JSON representation that is hashed
// and compared for immutable duplicate detection. The content_sha256 field is
// excluded so the digest cannot recursively include itself.
func (e CognitiveEvent) CanonicalBytes() ([]byte, error) {
	copy := e
	copy.ContentSHA256 = ""
	if err := copy.validateWithoutHash(); err != nil {
		return nil, err
	}
	payload, err := canonicalPayloadJSON(copy.PayloadJSON)
	if err != nil {
		return nil, err
	}

	// A struct, rather than a map, fixes key order. RawMessage retains the
	// already-normalized payload JSON without an additional quoted layer.
	canonical := struct {
		SchemaVersion  int             `json:"schema_version"`
		EventID        string          `json:"event_id"`
		EventType      string          `json:"event_type"`
		OccurredAt     string          `json:"occurred_at"`
		IdentityID     string          `json:"identity_id"`
		SessionID      string          `json:"session_id"`
		CorrelationID  string          `json:"correlation_id"`
		CausationID    string          `json:"causation_id"`
		GoalID         string          `json:"goal_id"`
		ActionID       string          `json:"action_id"`
		IdempotencyKey string          `json:"idempotency_key"`
		EvidenceClass  string          `json:"evidence_class"`
		PayloadJSON    json.RawMessage `json:"payload_json"`
		Provider       string          `json:"provider"`
		BackendKind    string          `json:"backend_kind"`
		ModelID        string          `json:"model_id"`
		RouteReason    string          `json:"route_reason"`
		Degraded       bool            `json:"degraded"`
		PolicyVersion  string          `json:"policy_version"`
	}{
		SchemaVersion:  copy.SchemaVersion,
		EventID:        copy.EventID,
		EventType:      copy.EventType,
		OccurredAt:     copy.OccurredAt.UTC().Format(time.RFC3339Nano),
		IdentityID:     copy.IdentityID,
		SessionID:      copy.SessionID,
		CorrelationID:  copy.CorrelationID,
		CausationID:    copy.CausationID,
		GoalID:         copy.GoalID,
		ActionID:       copy.ActionID,
		IdempotencyKey: copy.IdempotencyKey,
		EvidenceClass:  copy.EvidenceClass,
		PayloadJSON:    payload,
		Provider:       copy.Provider,
		BackendKind:    copy.BackendKind,
		ModelID:        copy.ModelID,
		RouteReason:    copy.RouteReason,
		Degraded:       copy.Degraded,
		PolicyVersion:  copy.PolicyVersion,
	}
	return json.Marshal(canonical)
}

// CanonicalHash returns a stable lower-case SHA-256 digest of CanonicalBytes.
func (e CognitiveEvent) CanonicalHash() (string, error) {
	canonical, err := e.CanonicalBytes()
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}

// WithCanonicalHash validates an event and returns a copy with its immutable
// content digest populated. Callers retain ownership of the original value.
func (e CognitiveEvent) WithCanonicalHash() (CognitiveEvent, error) {
	e.ContentSHA256 = ""
	if err := e.Validate(); err != nil {
		return CognitiveEvent{}, err
	}
	hash, err := e.CanonicalHash()
	if err != nil {
		return CognitiveEvent{}, err
	}
	e.ContentSHA256 = hash
	return e, nil
}

func (e CognitiveEvent) validateWithoutHash() error {
	e.ContentSHA256 = ""
	return e.Validate()
}

func validateRequiredIdentifier(field, value string) error {
	if value == "" {
		return fmt.Errorf("%w: %s is required", ErrInvalidCognitiveEvent, field)
	}
	return validateOptionalIdentifier(field, value)
}

func validateOptionalIdentifier(field, value string) error {
	if value == "" {
		return nil
	}
	if value != strings.TrimSpace(value) || !identifierPattern.MatchString(value) {
		return fmt.Errorf("%w: invalid %s", ErrInvalidCognitiveEvent, field)
	}
	return nil
}

func canonicalPayloadJSON(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("%w: payload_json is required", ErrInvalidCognitiveEventPayload)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("%w: malformed JSON: %v", ErrInvalidCognitiveEventPayload, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("%w: payload_json contains trailing JSON values", ErrInvalidCognitiveEventPayload)
	}
	if _, ok := value.(map[string]any); !ok {
		return nil, fmt.Errorf("%w: payload_json must be a JSON object", ErrInvalidCognitiveEventPayload)
	}
	if containsHiddenReasoningKey(value) {
		return nil, fmt.Errorf("%w: payload_json must not contain hidden reasoning", ErrInvalidCognitiveEventPayload)
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("%w: canonicalize payload JSON: %v", ErrInvalidCognitiveEventPayload, err)
	}
	return json.RawMessage(canonical), nil
}

func containsHiddenReasoningKey(value any) bool {
	switch v := value.(type) {
	case map[string]any:
		for key, nested := range v {
			normalized := strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.ToLower(key))
			switch normalized {
			case "chainofthought", "hiddenchainofthought", "reasoningtrace", "privatereasoning", "scratchpad":
				return true
			}
			if containsHiddenReasoningKey(nested) {
				return true
			}
		}
	case []any:
		for _, nested := range v {
			if containsHiddenReasoningKey(nested) {
				return true
			}
		}
	}
	return false
}

// CognitiveEventConflictError identifies the uniqueness constraint that was
// violated without exposing the persisted event payload.
type CognitiveEventConflictError struct {
	Kind string
	Key  string
}

func (e *CognitiveEventConflictError) Error() string {
	return fmt.Sprintf("%s conflict for %q: %v", e.Kind, e.Key, ErrCognitiveEventConflict)
}

func (e *CognitiveEventConflictError) Unwrap() error { return ErrCognitiveEventConflict }

// DuplicateCognitiveEventResult is a typed successful result for retrying the
// exact immutable event. Duplicate is never used for conflicting material.
type DuplicateCognitiveEventResult struct {
	Sequence  int64
	Duplicate bool
}

// NewCognitiveEvent returns a minimally populated version-one event. IDs and
// policy version remain explicit because they are part of the durable contract.
func NewCognitiveEvent(eventID, eventType, identityID, sessionID, policyVersion string, payload json.RawMessage) CognitiveEvent {
	return CognitiveEvent{
		SchemaVersion: CognitiveEventSchemaVersion,
		EventID:       eventID,
		EventType:     eventType,
		OccurredAt:    time.Now().UTC(),
		IdentityID:    identityID,
		SessionID:     sessionID,
		EvidenceClass: EvidenceClassOperational,
		PayloadJSON:   payload,
		PolicyVersion: policyVersion,
	}
}

// IsCognitiveEventDuplicate reports whether an append result represents an
// exact, already-persisted immutable event.
func IsCognitiveEventDuplicate(result DuplicateCognitiveEventResult) bool { return result.Duplicate }

// Ensure a change to the envelope cannot silently turn an error contract into
// an unused declaration during future refactors.
var _ = errors.Is
