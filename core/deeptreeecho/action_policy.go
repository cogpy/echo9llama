package deeptreeecho

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/wisdom"
)

const defaultEnactionFailureCooldown = 30 * time.Minute

// ActionPolicyConfig defines hard authority limits. None of these limits can be
// expanded by model output or by the advisory moral-agency layer.
type ActionPolicyConfig struct {
	Mode                EnactionMode
	MaxArtifactBytes    int
	MaxArtifactsPerWake int
	MaxActiveActions    int
	UncertaintyFloor    float64
	FailureLimit        int
	FailureCooldown     time.Duration
}

func DefaultActionPolicyConfig() ActionPolicyConfig {
	return ActionPolicyConfig{
		Mode:                EnactionObserve,
		MaxArtifactBytes:    32 * 1024,
		MaxArtifactsPerWake: 8,
		MaxActiveActions:    1,
		UncertaintyFloor:    0.45,
		FailureLimit:        3,
		FailureCooldown:     defaultEnactionFailureCooldown,
	}
}

// ActionPolicyDecision is durable public rationale, not private reasoning.
type ActionPolicyDecision struct {
	PolicyVersion     string       `json:"policy_version"`
	Allowed           bool         `json:"allowed"`
	Code              string       `json:"code"`
	Reason            string       `json:"reason"`
	Mode              EnactionMode `json:"mode"`
	AdvisoryStrategy  string       `json:"advisory_strategy"`
	AdvisoryRationale string       `json:"advisory_rationale"`
	ArtifactsUsed     int          `json:"artifacts_used"`
	ActiveActions     int          `json:"active_actions"`
	ConsecutiveErrors int          `json:"consecutive_errors"`
	CooldownUntil     time.Time    `json:"cooldown_until,omitempty"`
	DecidedAt         time.Time    `json:"decided_at"`
}

// ActionPolicy is the authoritative deterministic capability gate for E1.
type ActionPolicy struct {
	mu sync.Mutex

	config             ActionPolicyConfig
	moralAgency        *wisdom.MoralAgency
	artifactsThisWake  int
	activeActions      int
	consecutiveFailure int
	cooldownUntil      time.Time
}

func NewActionPolicy(config ActionPolicyConfig) *ActionPolicy {
	defaults := DefaultActionPolicyConfig()
	if config.Mode == "" {
		config.Mode = defaults.Mode
	}
	if config.MaxArtifactBytes <= 0 {
		config.MaxArtifactBytes = defaults.MaxArtifactBytes
	}
	if config.MaxArtifactsPerWake <= 0 {
		config.MaxArtifactsPerWake = defaults.MaxArtifactsPerWake
	}
	if config.MaxActiveActions <= 0 {
		config.MaxActiveActions = defaults.MaxActiveActions
	}
	if config.UncertaintyFloor <= 0 || config.UncertaintyFloor > 1 {
		config.UncertaintyFloor = defaults.UncertaintyFloor
	}
	if config.FailureLimit <= 0 {
		config.FailureLimit = defaults.FailureLimit
	}
	if config.FailureCooldown <= 0 {
		config.FailureCooldown = defaults.FailureCooldown
	}
	return &ActionPolicy{config: config, moralAgency: wisdom.NewMoralAgency()}
}

func (policy *ActionPolicy) Decide(plan ActionPlan) ActionPolicyDecision {
	policy.mu.Lock()
	defer policy.mu.Unlock()

	now := time.Now().UTC()
	decision := ActionPolicyDecision{
		PolicyVersion:     EnactionPolicyVersion,
		Mode:              policy.config.Mode,
		ArtifactsUsed:     policy.artifactsThisWake,
		ActiveActions:     policy.activeActions,
		ConsecutiveErrors: policy.consecutiveFailure,
		CooldownUntil:     policy.cooldownUntil,
		DecidedAt:         now,
	}

	strategy, rationale := policy.moralAgency.Decide(
		fmt.Sprintf("Create a private local note for this purpose: %s", strings.TrimSpace(plan.Purpose)),
		"echo-self",
		map[string]float64{"confidence": clampUnit(plan.Confidence)},
	)
	decision.AdvisoryStrategy = strategy.String()
	decision.AdvisoryRationale = rationale

	deny := func(code, reason string) ActionPolicyDecision {
		decision.Allowed = false
		decision.Code = code
		decision.Reason = reason
		return decision
	}

	if err := plan.Validate(); err != nil {
		return deny("invalid_plan", err.Error())
	}
	if policy.config.Mode != EnactionLocalSandbox {
		return deny("observe_mode", "local action is disabled; observe mode records proposals only")
	}
	if plan.Tool != WorkspaceCreateNote {
		return deny("capability_denied", "tool is not in the E1 allowlist")
	}
	if len([]byte(plan.Content)) > policy.config.MaxArtifactBytes {
		return deny("artifact_too_large", "proposed artifact exceeds the configured byte limit")
	}
	if plan.Confidence < policy.config.UncertaintyFloor {
		return deny("uncertainty_escalation", "proposal confidence is below the action threshold")
	}
	if !plan.Reversible {
		return deny("irreversible_action", "E1 permits only reversible local artifacts")
	}
	if !selfOnly(plan.AffectedParties) {
		return deny("affected_party_review", "E1 cannot act on behalf of third parties")
	}
	cleaned := filepath.Clean(strings.TrimSpace(plan.RelativePath))
	if filepath.IsAbs(cleaned) || cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return deny("workspace_escape", "target must remain inside the private workspace")
	}
	if !policy.cooldownUntil.IsZero() && now.Before(policy.cooldownUntil) {
		return deny("failure_cooldown", "action cooldown is active after repeated failures")
	}
	if policy.activeActions >= policy.config.MaxActiveActions {
		return deny("active_action_budget", "the active action budget is exhausted")
	}
	if policy.artifactsThisWake >= policy.config.MaxArtifactsPerWake {
		return deny("wake_artifact_budget", "the per-wake artifact budget is exhausted")
	}

	policy.activeActions++
	decision.ActiveActions = policy.activeActions
	decision.Allowed = true
	decision.Code = "allowed"
	decision.Reason = "validated create-only private workspace action within all E1 bounds"
	return decision
}

func (policy *ActionPolicy) RecordOutcome(verified bool, lesson string) {
	policy.mu.Lock()
	defer policy.mu.Unlock()

	if policy.activeActions > 0 {
		policy.activeActions--
	}
	outcome := -1.0
	if verified {
		policy.artifactsThisWake++
		policy.consecutiveFailure = 0
		policy.cooldownUntil = time.Time{}
		outcome = 1.0
	} else {
		policy.consecutiveFailure++
		if policy.consecutiveFailure >= policy.config.FailureLimit {
			policy.cooldownUntil = time.Now().UTC().Add(policy.config.FailureCooldown)
		}
	}
	policy.moralAgency.LearnFromOutcome(outcome, lesson)
}

func (policy *ActionPolicy) ResetWakeBudget() {
	policy.mu.Lock()
	defer policy.mu.Unlock()
	policy.artifactsThisWake = 0
}

func (policy *ActionPolicy) Snapshot() ActionPolicyDecision {
	policy.mu.Lock()
	defer policy.mu.Unlock()
	return ActionPolicyDecision{
		PolicyVersion:     EnactionPolicyVersion,
		Mode:              policy.config.Mode,
		ArtifactsUsed:     policy.artifactsThisWake,
		ActiveActions:     policy.activeActions,
		ConsecutiveErrors: policy.consecutiveFailure,
		CooldownUntil:     policy.cooldownUntil,
		DecidedAt:         time.Now().UTC(),
	}
}

func selfOnly(parties []string) bool {
	for _, party := range parties {
		switch strings.ToLower(strings.TrimSpace(party)) {
		case "self", "echo", "deep tree echo":
		default:
			return false
		}
	}
	return len(parties) > 0
}
