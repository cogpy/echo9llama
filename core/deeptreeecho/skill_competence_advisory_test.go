package deeptreeecho

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestAdvisoryCompetenceFailureAfterManySuccessesLowersEstimate(t *testing.T) {
	system := NewSkillLearningSystem(verbosePracticeProvider{})
	const name = "Verified Artifact Creation"
	for i := range 50 {
		if applied, err := system.ApplyEvaluatorEvidence(name, fmt.Sprintf("pass-%d", i), 1, true, "verified read-back"); err != nil || !applied {
			t.Fatalf("record success: applied=%v err=%v", applied, err)
		}
	}
	before, _ := system.GetSkillByID(stableSkillID(name))
	if applied, err := system.ApplyEvaluatorEvidence(name, "failure", 0.95, false, "read-back mismatch"); err != nil || !applied {
		t.Fatalf("record failure: applied=%v err=%v", applied, err)
	}
	after, _ := system.GetSkillByID(stableSkillID(name))
	if after.Proficiency >= before.Proficiency || after.Competence.PassPosteriorMean >= before.Competence.PassPosteriorMean {
		t.Fatalf("failure did not lower advisory estimates: before=%+v after=%+v", before.Competence, after.Competence)
	}
	if math.Abs(after.Proficiency-before.Proficiency*(1-SkillCompetenceEMAWeight)) > 1e-12 {
		t.Fatalf("failure did not follow documented zero-signal EMA: %v", after.Proficiency)
	}
	estimate := after.Competence
	if estimate.Version != SkillCompetenceVersion || !estimate.Advisory || estimate.Evaluations != 51 || estimate.Passes != 50 || estimate.Failures != 1 {
		t.Fatalf("evidence counters/version missing: %+v", estimate)
	}
	if estimate.ScoreSum != 50.95 || math.Abs(estimate.MeanScore-50.95/51) > 1e-12 {
		t.Fatalf("raw observed scores were lost: %+v", estimate)
	}
	last := after.Attempts[len(after.Attempts)-1]
	if last.Status != "failed" || last.Success || last.EvidenceID != "failure" || last.Performance != 0.95 {
		t.Fatalf("negative evidence not recorded transparently: %+v", last)
	}
	if system.GetMetrics()["total_masteries"] != uint64(0) || strings.Contains(system.GetSkillProfile(), "Mastered") {
		t.Fatal("repeated evaluator success was promoted to universal mastery")
	}
	found := false
	for _, skill := range system.GetSkillsNeedingPractice() {
		if skill.ID == before.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("high advisory score incorrectly exempts skill from further practice")
	}
}

func TestAdvisoryCompetenceReplayAndCollision(t *testing.T) {
	system := NewSkillLearningSystem(verbosePracticeProvider{})
	const name = "Evidence Bound Skill"
	if applied, err := system.ApplyEvaluatorEvidence(name, "id", 0.8, true, "actual feedback"); err != nil || !applied {
		t.Fatalf("initial apply=%v err=%v", applied, err)
	}
	before, _ := system.GetSkillByID(stableSkillID(name))
	if applied, err := system.ApplyEvaluatorEvidence(" "+name+" ", " id ", 0.8, true, " actual feedback "); err != nil || applied {
		t.Fatalf("identical normalized replay not inert: apply=%v err=%v", applied, err)
	}
	for _, collision := range []struct {
		name     string
		score    float64
		passed   bool
		feedback string
	}{
		{name, 0.7, true, "actual feedback"},
		{name, 0.8, false, "actual feedback"},
		{name, 0.8, true, "different feedback"},
		{"Different Skill", 0.8, true, "actual feedback"},
	} {
		if applied, err := system.ApplyEvaluatorEvidence(collision.name, "id", collision.score, collision.passed, collision.feedback); err == nil || applied {
			t.Fatalf("conflicting payload accepted: %+v apply=%v err=%v", collision, applied, err)
		}
	}
	after, _ := system.GetSkillByID(stableSkillID(name))
	if !reflect.DeepEqual(before, after) {
		t.Fatal("replay or collision changed recorded skill")
	}
	if _, err := system.GetSkillByID(stableSkillID("Different Skill")); err == nil {
		t.Fatal("conflicting ID created a new skill")
	}
	if applied, err := system.ApplyEvaluatorEvidence("A-B", "name-first", 1, true, "verified"); err != nil || !applied {
		t.Fatal(err)
	}
	if applied, err := system.ApplyEvaluatorEvidence("A B", "name-collision", 1, true, "verified"); err == nil || applied {
		t.Fatal("distinct names sharing a stable ID received each other's credit")
	}
}

func TestAdvisoryCompetenceInvalidScoresFailClosed(t *testing.T) {
	system := NewSkillLearningSystem(nil)
	before := system.GetMetrics()
	for i, score := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), -0.1, 1.1} {
		if applied, err := system.ApplyEvaluatorEvidence("invalid", fmt.Sprint(i), score, true, "claim"); err == nil || applied {
			t.Fatalf("invalid score %v applied=%v err=%v", score, applied, err)
		}
	}
	if !reflect.DeepEqual(before, system.GetMetrics()) {
		t.Fatal("invalid evidence mutated metrics")
	}
	if err := system.PracticeSkill(stableSkillID("Pattern Recognition")); err == nil {
		t.Fatal("nil rehearsal provider did not fail closed")
	}
}

func TestAdvisoryCompetenceUnknownRehearsalAndSnapshots(t *testing.T) {
	system := NewSkillLearningSystem(verbosePracticeProvider{})
	id := stableSkillID("Pattern Recognition")
	before, _ := system.GetSkillByID(id)
	if before.Proficiency != 0 || before.Competence.Evaluations != 0 {
		t.Fatal("foundational skill starts with fabricated competence")
	}
	if err := system.PracticeSkill(id); err != nil {
		t.Fatal(err)
	}
	after, _ := system.GetSkillByID(id)
	if after.Competence != before.Competence || after.Proficiency != before.Proficiency || after.Attempts[0].Status != "unknown" {
		t.Fatalf("rehearsal awarded competence or lost unknown status: %+v", after)
	}
	after.Attempts[0].Feedback = "mutated"
	after.Competence.EffectiveScoreEMA = 1
	fresh, _ := system.GetSkillByID(id)
	if fresh.Attempts[0].Feedback == "mutated" || fresh.Competence.EffectiveScoreEMA != 0 {
		t.Fatal("getter exposes mutable skill state")
	}
}

func TestAdvisoryCompetenceConcurrentEvidenceAndRehearsal(t *testing.T) {
	system := NewSkillLearningSystem(verbosePracticeProvider{})
	const name = "Pattern Recognition"
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			if _, err := system.ApplyEvaluatorEvidence(name, fmt.Sprintf("concurrent-%d", i), 0.5, i%2 == 0, "observed"); err != nil {
				t.Error(err)
			}
		})
		wg.Go(func() {
			if err := system.PracticeSkill(stableSkillID(name)); err != nil {
				t.Error(err)
			}
		})
		wg.Go(func() {
			_ = system.GetSkills()
			_ = system.GetSkillsNeedingPractice()
			_ = system.GetMetrics()
		})
	}
	wg.Wait()
	skill, _ := system.GetSkillByID(stableSkillID(name))
	if skill.Competence.Evaluations != 20 || skill.Competence.Passes != 10 || skill.Competence.Failures != 10 || len(skill.Attempts) != 40 {
		t.Fatalf("lost concurrent evidence: %+v attempts=%d", skill.Competence, len(skill.Attempts))
	}
}
