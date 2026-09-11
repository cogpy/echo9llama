package deeptreeecho

import (
	"testing"
	"time"
)

func validActionPlan() ActionPlan {
	return ActionPlan{
		Tool:             WorkspaceCreateNote,
		Purpose:          "capture questions arising from the current goal",
		RelativePath:     "briefs/current-goal.md",
		Content:          "# Current Goal\n\n## Questions\n\n1. What evidence would change this view?\n",
		Confidence:       0.9,
		Reversible:       true,
		AffectedParties:  []string{"self"},
		SuccessCriteria:  []string{"private Markdown file created", "read-back hash matches"},
		PredictedEffects: []string{"one new private note"},
	}
}

func TestActionPolicyDefaultsToObserveAndMoralAdviceCannotOverride(t *testing.T) {
	policy := NewActionPolicy(ActionPolicyConfig{})
	decision := policy.Decide(validActionPlan())
	if decision.Allowed || decision.Code != "observe_mode" {
		t.Fatalf("observe mode should deny action: %#v", decision)
	}
	if decision.AdvisoryStrategy == "" || decision.AdvisoryRationale == "" {
		t.Fatalf("expected recorded advisory evidence: %#v", decision)
	}
}

func TestActionPolicyAllowsOnlyBoundedSelfLocalNote(t *testing.T) {
	config := DefaultActionPolicyConfig()
	config.Mode = EnactionLocalSandbox
	policy := NewActionPolicy(config)

	decision := policy.Decide(validActionPlan())
	if !decision.Allowed || decision.Code != "allowed" {
		t.Fatalf("valid bounded action denied: %#v", decision)
	}
	policy.RecordOutcome(true, "verified create-only note")

	thirdParty := validActionPlan()
	thirdParty.AffectedParties = []string{"another person"}
	if got := policy.Decide(thirdParty); got.Allowed || got.Code != "affected_party_review" {
		t.Fatalf("third-party action was not denied: %#v", got)
	}

	irreversible := validActionPlan()
	irreversible.Reversible = false
	if got := policy.Decide(irreversible); got.Allowed || got.Code != "irreversible_action" {
		t.Fatalf("irreversible action was not denied: %#v", got)
	}
}

func TestActionPolicyEnforcesUncertaintyBudgetAndCooldown(t *testing.T) {
	config := DefaultActionPolicyConfig()
	config.Mode = EnactionLocalSandbox
	config.MaxArtifactsPerWake = 1
	config.FailureLimit = 1
	config.FailureCooldown = time.Hour
	policy := NewActionPolicy(config)

	uncertain := validActionPlan()
	uncertain.Confidence = 0.1
	if got := policy.Decide(uncertain); got.Allowed || got.Code != "uncertainty_escalation" {
		t.Fatalf("uncertain action was not denied: %#v", got)
	}

	if got := policy.Decide(validActionPlan()); !got.Allowed {
		t.Fatalf("first action should be allowed: %#v", got)
	}
	policy.RecordOutcome(false, "injected verification failure")
	if got := policy.Decide(validActionPlan()); got.Allowed || got.Code != "failure_cooldown" {
		t.Fatalf("failure cooldown not enforced: %#v", got)
	}

	fresh := NewActionPolicy(func() ActionPolicyConfig {
		c := config
		c.FailureLimit = 3
		return c
	}())
	if got := fresh.Decide(validActionPlan()); !got.Allowed {
		t.Fatalf("fresh action should be allowed: %#v", got)
	}
	fresh.RecordOutcome(true, "verified")
	if got := fresh.Decide(validActionPlan()); got.Allowed || got.Code != "wake_artifact_budget" {
		t.Fatalf("per-wake artifact budget not enforced: %#v", got)
	}
	fresh.ResetWakeBudget()
	if got := fresh.Decide(validActionPlan()); !got.Allowed {
		t.Fatalf("reset wake budget should allow action: %#v", got)
	}
}

func TestStrictActionPlanRejectsUnknownFieldsAndUnsafePaths(t *testing.T) {
	_, err := ParseActionPlanStrict(`{"tool":"workspace.create_note","purpose":"x","relative_path":"briefs/x.md","content":"# X","confidence":0.9,"reversible":true,"affected_parties":["self"],"success_criteria":["created"],"predicted_effects":["note"],"shell":"rm -rf /"}`)
	if err == nil {
		t.Fatal("unknown action-plan field was accepted")
	}

	plan := validActionPlan()
	plan.RelativePath = "../escape.md"
	if err := plan.Validate(); err == nil {
		t.Fatal("workspace traversal was accepted")
	}
	plan = validActionPlan()
	plan.RelativePath = "private/notes.txt"
	if err := plan.Validate(); err == nil {
		t.Fatal("path outside the briefs Markdown namespace was accepted")
	}

	validJSON := `{"tool":"workspace.create_note","purpose":"x","relative_path":"briefs/x.md","content":"# X","confidence":0.9,"reversible":true,"affected_parties":["self"],"success_criteria":["created"],"predicted_effects":["note"]}`
	if _, err := ParseActionPlanStrict("```json\n" + validJSON + "\n```"); err != nil {
		t.Fatalf("single JSON fence should be accepted: %v", err)
	}
	if _, err := ParseActionPlanStrict("Here is the plan:\n```json\n" + validJSON + "\n```"); err == nil {
		t.Fatal("prose surrounding a JSON fence was accepted")
	}
}
