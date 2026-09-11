package deeptreeecho

import (
	"encoding/json"
	"testing"

	"github.com/cogpy/echo9llama/core/echodream"
	"github.com/cogpy/echo9llama/core/persistence"
)

func TestProjectEnactionEventsRebuildsEvidenceStateExactlyOnce(t *testing.T) {
	orchestrator := &UnifiedAutonomousOrchestrator{
		identityID:            "echo-projection",
		echobeatsScheduler:    NewEchobeatsScheduler(verbosePracticeProvider{}),
		skillLearning:         NewSkillLearningSystem(verbosePracticeProvider{}),
		dreamCycle:            echodream.NewSleepWakeStateMachine(),
		ingestedExperienceIDs: make(map[string]struct{}),
	}
	marshal := func(value any) json.RawMessage {
		payload, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return payload
	}
	events := []persistence.StoredCognitiveEvent{
		{Sequence: 1, CognitiveEvent: persistence.CognitiveEvent{
			EventID: "goal-progress:event-1", EventType: persistence.EventTypeGoalProgressed,
			IdentityID: "echo-projection", GoalID: "goal-1",
			PayloadJSON: marshal(map[string]any{"delta": 0.2, "description": "Evidence-restored goal"}),
		}},
		{Sequence: 2, CognitiveEvent: persistence.CognitiveEvent{
			EventID: "skill-evidence:event-1", EventType: persistence.EventTypeSkillEvidence,
			IdentityID:  "echo-projection",
			PayloadJSON: marshal(map[string]any{"skill": structuredArtifactSkill, "score": 1.0, "passed": true, "feedback": "verified"}),
		}},
		{Sequence: 3, CognitiveEvent: persistence.CognitiveEvent{
			EventID: "dream:event-1", EventType: persistence.EventTypeDreamExperience,
			IdentityID:  "echo-projection",
			PayloadJSON: marshal(map[string]any{"source_event_id": "evaluation:event-1", "importance": 0.8, "summary": "Verified one bounded artifact.", "tags": []string{"verified"}}),
		}},
	}
	if err := orchestrator.projectEnactionEvents(events); err != nil {
		t.Fatalf("first projection: %v", err)
	}
	if err := orchestrator.projectEnactionEvents(events); err != nil {
		t.Fatalf("second projection: %v", err)
	}
	goal := orchestrator.echobeatsScheduler.GetActiveGoal()
	if goal == nil || goal.Progress != 0.2 {
		t.Fatalf("goal evidence duplicated or missing: %#v", goal)
	}
	skill, err := orchestrator.skillLearning.GetSkillByID(stableSkillID(structuredArtifactSkill))
	if err != nil {
		t.Fatalf("skill projection missing: %v", err)
	}
	if skill.PracticeCount != 1 || skill.Proficiency <= 0 {
		t.Fatalf("skill evidence duplicated or missing: %#v", skill)
	}
	metrics := orchestrator.dreamCycle.GetMetrics()
	if pending, _ := metrics["pending_experiences"].(int); pending != 1 {
		t.Fatalf("dream evidence duplicated or missing: %#v", metrics)
	}
}
