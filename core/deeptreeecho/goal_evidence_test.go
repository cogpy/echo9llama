package deeptreeecho

import "testing"

func TestApplyGoalEvidenceIsIdempotentAndRestoresGoal(t *testing.T) {
	scheduler := NewEchobeatsScheduler(verbosePracticeProvider{})
	if !scheduler.ApplyGoalEvidence("goal-restored", "Restored evidence goal", "goal.progressed:event-1", 0.25) {
		t.Fatal("first goal evidence was not applied")
	}
	if scheduler.ApplyGoalEvidence("goal-restored", "Restored evidence goal", "goal.progressed:event-1", 0.25) {
		t.Fatal("duplicate goal evidence was applied twice")
	}
	goal := scheduler.GetActiveGoal()
	if goal == nil || goal.ID != "goal-restored" || goal.Progress != 0.25 || goal.Status != GoalActive {
		t.Fatalf("goal projection mismatch: %#v", goal)
	}
}
