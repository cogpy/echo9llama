package deeptreeecho

import (
	"fmt"
	"testing"
	"time"
)

func newCanonicalOrchestratorForTest(t *testing.T) *UnifiedAutonomousOrchestrator {
	t.Helper()
	config := DefaultOrchestratorConfig()
	config.EnablePersistence = false
	config.EnableSkillLearning = false
	config.EnableDiscussionMonitoring = false
	config.DreamLightDuration = time.Millisecond
	config.DreamDeepDuration = time.Millisecond
	config.DreamREMDuration = time.Millisecond
	config.WakeDuration = 100 * time.Millisecond
	config.RestDuration = 60 * time.Millisecond

	orchestrator := NewUnifiedAutonomousOrchestrator(&lifecycleMockLLMProvider{}, config)
	t.Cleanup(func() {
		if orchestrator.dreamCycle != nil {
			orchestrator.dreamCycle.Shutdown()
		}
	})
	return orchestrator
}

func waitForDreamWisdom(t *testing.T, orchestrator *UnifiedAutonomousOrchestrator) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if metrics := orchestrator.dreamCycle.GetMetrics(); metrics["wisdom_synthesized"].(uint64) > 0 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("EchoDream did not synthesize wisdom before deadline: %#v", orchestrator.dreamCycle.GetMetrics())
}

func seedRecurringDreamExperiences(t *testing.T, orchestrator *UnifiedAutonomousOrchestrator) {
	t.Helper()
	if !orchestrator.ingestDreamExperienceOnce("seed-1", "Careful practice revealed a stable feedback pattern", 0.9, []string{"practice", "feedback"}) {
		t.Fatal("expected first seed experience to be ingested")
	}
	if !orchestrator.ingestDreamExperienceOnce("seed-2", "Repeated practice strengthened the same feedback pattern", 0.85, []string{"practice", "feedback"}) {
		t.Fatal("expected second seed experience to be ingested")
	}
}

func TestUnifiedOrchestratorOwnsCanonicalEchoDream(t *testing.T) {
	orchestrator := newCanonicalOrchestratorForTest(t)
	if orchestrator.dreamCycle == nil {
		t.Fatal("expected canonical echodream.SleepWakeStateMachine")
	}

	status := orchestrator.GetStatus()
	if status.DreamPhase == "" {
		t.Fatal("expected canonical dream phase in orchestrator status")
	}
	if status.Provider != "lifecycle-mock" || !status.ProviderAvailable {
		t.Fatalf("unexpected provider status: name=%q available=%v", status.Provider, status.ProviderAvailable)
	}
}

func TestDreamExperienceIngestionIsIdempotentAndBounded(t *testing.T) {
	orchestrator := newCanonicalOrchestratorForTest(t)

	if !orchestrator.ingestDreamExperienceOnce("same-event", "a meaningful event", 2.0, []string{"test"}) {
		t.Fatal("expected first event ingestion to succeed")
	}
	if orchestrator.ingestDreamExperienceOnce("same-event", "a meaningful event", 2.0, []string{"test"}) {
		t.Fatal("duplicate event should not be ingested")
	}
	if got := orchestrator.GetStatus().PendingExperiences; got != 1 {
		t.Fatalf("expected one pending experience after duplicate suppression, got %d", got)
	}

	for i := 0; i < maxExperienceLedgerEntries+50; i++ {
		orchestrator.markExperienceOnce(fmt.Sprintf("ledger-%d", i))
	}
	if got := orchestrator.GetStatus().ExperienceLedgerSize; got != maxExperienceLedgerEntries {
		t.Fatalf("expected bounded ledger size %d, got %d", maxExperienceLedgerEntries, got)
	}
	if !orchestrator.markExperienceOnce("ledger-0") {
		t.Fatal("expected oldest evicted key to be admissible again")
	}
	if got := orchestrator.GetStatus().ExperienceLedgerSize; got != maxExperienceLedgerEntries {
		t.Fatalf("ledger exceeded bound after reinsertion: %d", got)
	}
}

func TestWakeRestManagerRoundTripIntegratesDreamWisdomOnce(t *testing.T) {
	orchestrator := newCanonicalOrchestratorForTest(t)
	seedRecurringDreamExperiences(t, orchestrator)

	orchestrator.wakeRestCycle.transitionToRest()
	if orchestrator.isAwake {
		t.Fatal("rest callback did not synchronize orchestrator state")
	}
	if !orchestrator.dreamCycle.IsAsleep() {
		t.Fatal("rest callback did not start canonical EchoDream")
	}

	orchestrator.wakeRestCycle.transitionToDream()
	waitForDreamWisdom(t, orchestrator)
	orchestrator.wakeRestCycle.transitionToWake()

	status := orchestrator.GetStatus()
	if !status.IsAwake || status.WakeRestState != StateAwake.String() {
		t.Fatalf("wake round trip did not synchronize state: %#v", status)
	}
	if status.TotalWisdom == 0 || status.WisdomDepth <= 0 {
		t.Fatalf("dream wisdom was not integrated: %#v", status)
	}
	if status.PendingExperiences != 0 {
		t.Fatalf("dream cycle did not consume pending experiences: %d", status.PendingExperiences)
	}

	wisdomCount := status.TotalWisdom
	if err := orchestrator.onWake(); err != nil {
		t.Fatalf("idempotent wake failed: %v", err)
	}
	if got := orchestrator.GetStatus().TotalWisdom; got != wisdomCount {
		t.Fatalf("wisdom was integrated twice: before=%d after=%d", wisdomCount, got)
	}
}

func TestThoughtDiscussionAndGoalOutcomesBecomeDreamExperiencesOnce(t *testing.T) {
	orchestrator := newCanonicalOrchestratorForTest(t)

	orchestrator.streamOfConsciousness.mu.Lock()
	orchestrator.streamOfConsciousness.thoughts = append(orchestrator.streamOfConsciousness.thoughts, AutonomousThought{
		ID:         "thought-source-1",
		Content:    "I should test whether reflection changes future attention.",
		Type:       ThoughtReflection,
		Timestamp:  time.Now(),
		Importance: 0.8,
		Tags:       []string{"reflection"},
	})
	orchestrator.streamOfConsciousness.mu.Unlock()

	if got := orchestrator.captureThoughtExperiences(); got != 1 {
		t.Fatalf("expected one thought experience, got %d", got)
	}
	if got := orchestrator.captureThoughtExperiences(); got != 0 {
		t.Fatalf("expected repeated thought capture to be idempotent, got %d", got)
	}

	conversation := &TrackedConversation{
		ID:            "conversation-1",
		Topic:         "wisdom cultivation",
		InterestScore: 0.9,
		Status:        ConversationActive,
		Messages: []ConversationMessage{
			{ID: "message-1", Sender: "Dan", Content: "What changed your mind?", Timestamp: time.Now(), Topics: []string{"reflection"}},
			{ID: "message-2", Sender: "Echo", Content: "The feedback loop did.", Timestamp: time.Now(), IsFromEcho: true, Topics: []string{"feedback"}},
		},
	}
	if got := orchestrator.captureConversationExperiences([]*TrackedConversation{conversation}); got != 2 {
		t.Fatalf("expected two discussion experiences, got %d", got)
	}
	if got := orchestrator.captureConversationExperiences([]*TrackedConversation{conversation}); got != 0 {
		t.Fatalf("expected repeated discussion capture to be idempotent, got %d", got)
	}

	goalID := orchestrator.echobeatsScheduler.AddGoal("verify a closed learning loop", 0.9)
	orchestrator.echobeatsScheduler.UpdateGoalProgress(goalID, 1.0)
	pendingAfterGoal := orchestrator.GetStatus().PendingExperiences
	orchestrator.echobeatsScheduler.UpdateGoalProgress(goalID, 1.0)
	if got := orchestrator.GetStatus().PendingExperiences; got != pendingAfterGoal {
		t.Fatalf("goal completion was ingested more than once: before=%d after=%d", pendingAfterGoal, got)
	}

	if pendingAfterGoal != 4 {
		t.Fatalf("expected thought + two messages + goal outcome (4), got %d", pendingAfterGoal)
	}
}

func TestCognitiveAccessorsReturnDefensiveSnapshots(t *testing.T) {
	stream := &StreamOfConsciousness{thoughts: []AutonomousThought{{ID: "t1", Tags: []string{"original"}}}}
	thoughts := stream.GetAllThoughts()
	thoughts[0].Tags[0] = "mutated"
	if got := stream.GetAllThoughts()[0].Tags[0]; got != "original" {
		t.Fatalf("thought tags leaked mutable storage: %q", got)
	}

	monitor := &ConversationMonitor{conversations: map[string]*TrackedConversation{
		"c1": {
			ID:           "c1",
			Status:       ConversationActive,
			Participants: []string{"Dan"},
			Messages:     []ConversationMessage{{ID: "m1", Topics: []string{"original"}}},
			Context:      ConversationContext{RelevantKnowledge: []string{"original"}},
		},
	}}
	conversations := monitor.GetActiveConversations()
	conversations[0].Participants[0] = "mutated"
	conversations[0].Messages[0].Topics[0] = "mutated"
	conversations[0].Context.RelevantKnowledge[0] = "mutated"
	fresh := monitor.GetActiveConversations()[0]
	if fresh.Participants[0] != "Dan" || fresh.Messages[0].Topics[0] != "original" || fresh.Context.RelevantKnowledge[0] != "original" {
		t.Fatalf("conversation snapshot leaked mutable storage: %#v", fresh)
	}

	skills := &SkillLearningSystem{skills: map[string]*Skill{
		"s1": {ID: "s1", Prerequisites: []string{"original"}, Attempts: []SkillAttempt{{Feedback: "original"}}},
	}}
	snapshot, err := skills.GetSkillByID("s1")
	if err != nil {
		t.Fatalf("GetSkillByID failed: %v", err)
	}
	snapshot.Prerequisites[0] = "mutated"
	snapshot.Attempts[0].Feedback = "mutated"
	freshSkill, _ := skills.GetSkillByID("s1")
	if freshSkill.Prerequisites[0] != "original" || freshSkill.Attempts[0].Feedback != "original" {
		t.Fatalf("skill snapshot leaked mutable storage: %#v", freshSkill)
	}
}

func TestPersistentInterestContinuityAcrossRestart(t *testing.T) {
	stateDirectory := t.TempDir()
	config := DefaultOrchestratorConfig()
	config.StateDirectory = stateDirectory
	config.EnableSkillLearning = false
	config.EnableDiscussionMonitoring = false

	first := NewUnifiedAutonomousOrchestrator(&lifecycleMockLLMProvider{}, config)
	first.interestPatterns.UpdateInterest("recursive_wisdom_practice", 0.91)
	first.syncPersistentState()
	first.dreamCycle.Shutdown()

	second := NewUnifiedAutonomousOrchestrator(&lifecycleMockLLMProvider{}, config)
	t.Cleanup(func() { second.dreamCycle.Shutdown() })
	if got := second.interestPatterns.GetInterestLevel("recursive_wisdom_practice"); got < 0.90 {
		t.Fatalf("learned interest was not hydrated: %.3f", got)
	}
	if err := second.interestPatterns.Start(); err != nil {
		t.Fatalf("start restored interest system: %v", err)
	}
	t.Cleanup(func() { _ = second.interestPatterns.Stop() })
	if got := second.interestPatterns.GetInterestLevel("recursive_wisdom_practice"); got < 0.90 {
		t.Fatalf("core-interest initialization overwrote restored strength: %.3f", got)
	}
}
