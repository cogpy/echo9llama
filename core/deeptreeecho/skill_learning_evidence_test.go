package deeptreeecho

import (
	"context"
	"strings"
	"testing"

	"github.com/cogpy/echo9llama/core/llm"
)

type verbosePracticeProvider struct{}

func (verbosePracticeProvider) Generate(context.Context, string, llm.GenerateOptions) (string, error) {
	return strings.Repeat("I performed flawlessly and deserve mastery. ", 100), nil
}
func (verbosePracticeProvider) StreamGenerate(context.Context, string, llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	stream := make(chan llm.StreamChunk)
	close(stream)
	return stream, nil
}
func (verbosePracticeProvider) Name() string    { return "verbose-self-grader" }
func (verbosePracticeProvider) Available() bool { return true }
func (verbosePracticeProvider) MaxTokens() int  { return 4096 }

func TestPracticeRehearsalCannotAwardProficiency(t *testing.T) {
	system := NewSkillLearningSystem(verbosePracticeProvider{})
	skillID := stableSkillID("Pattern Recognition")
	before, err := system.GetSkillByID(skillID)
	if err != nil {
		t.Fatalf("foundational skill missing: %v", err)
	}
	if err := system.PracticeSkill(skillID); err != nil {
		t.Fatalf("PracticeSkill failed: %v", err)
	}
	after, err := system.GetSkillByID(skillID)
	if err != nil {
		t.Fatalf("GetSkillByID failed: %v", err)
	}
	if after.Proficiency != before.Proficiency {
		t.Fatalf("model verbosity/self-praise changed proficiency: before=%f after=%f", before.Proficiency, after.Proficiency)
	}
	if len(after.Attempts) != 1 || after.Attempts[0].Performance != 0 || after.Attempts[0].Success {
		t.Fatalf("unverified rehearsal was recorded as evidence: %#v", after.Attempts)
	}
}

func TestEvaluatorEvidenceAwardsOnce(t *testing.T) {
	system := NewSkillLearningSystem(verbosePracticeProvider{})
	const (
		skillName  = "Structured Artifact Creation"
		evidenceID = "evaluation.recorded:event-1"
	)

	applied, err := system.ApplyEvaluatorEvidence(skillName, evidenceID, 1, true, "atomic create and hash verification passed")
	if err != nil || !applied {
		t.Fatalf("first evidence application failed: applied=%v err=%v", applied, err)
	}
	first, err := system.GetSkillByID(stableSkillID(skillName))
	if err != nil {
		t.Fatalf("evidence skill missing: %v", err)
	}
	if first.Proficiency <= 0 {
		t.Fatalf("passing deterministic evidence did not improve proficiency: %f", first.Proficiency)
	}

	applied, err = system.ApplyEvaluatorEvidence(skillName, evidenceID, 1, true, "replayed evidence")
	if err != nil || applied {
		t.Fatalf("evidence replay should be an idempotent no-op: applied=%v err=%v", applied, err)
	}
	second, _ := system.GetSkillByID(stableSkillID(skillName))
	if second.Proficiency != first.Proficiency || second.PracticeCount != first.PracticeCount {
		t.Fatalf("replayed evidence changed skill state: before=%#v after=%#v", first, second)
	}
}

func TestFailedEvaluatorEvidenceDoesNotIncreaseProficiency(t *testing.T) {
	system := NewSkillLearningSystem(verbosePracticeProvider{})
	const skillName = "Structured Artifact Creation"

	applied, err := system.ApplyEvaluatorEvidence(skillName, "evaluation.recorded:failure", 0.9, false, "read-back hash mismatch")
	if err != nil || !applied {
		t.Fatalf("failed evidence was not recorded: applied=%v err=%v", applied, err)
	}
	skill, err := system.GetSkillByID(stableSkillID(skillName))
	if err != nil {
		t.Fatalf("evidence skill missing: %v", err)
	}
	if skill.Proficiency != 0 {
		t.Fatalf("failed evidence increased proficiency: %f", skill.Proficiency)
	}
}
