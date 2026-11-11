package deeptreeecho

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/cogpy/echo9llama/core/persistence"
)

func reviewFixturePlanner() *scriptedActionPlanner {
	return &scriptedActionPlanner{route: PlanRouteEvidence{TraceID: "review-fixture", Provider: "explicit-unit-fixture", BackendKind: "remote_api", ModelID: "unit-test-fixture"}}
}

func reviewGoal(t *testing.T, rig *UnifiedAutonomousOrchestrator, id string) ScheduledGoal {
	t.Helper()
	for _, goal := range rig.echobeatsScheduler.GetGoalQueue() {
		if goal.ID == id {
			return goal
		}
	}
	t.Fatalf("missing goal %s", id)
	return ScheduledGoal{}
}

func reviewRun(t *testing.T, rig *UnifiedAutonomousOrchestrator, pipeline *EnactionPipeline, planner *scriptedActionPlanner, request EnactionRequest, answer string) persistence.StoredCognitiveEvent {
	t.Helper()
	if planner.err == nil {
		planner.raw = testPlanJSON(t, fmt.Sprintf("briefs/review-%s.md", request.Practice.TaskID), answer)
	}
	_, err := pipeline.RunGoal(context.Background(), request)
	if (err != nil) != (planner.err != nil) {
		t.Fatalf("unexpected run error: %v", err)
	}
	events, err := rig.eventStore.Query(context.Background(), persistence.CognitiveEventQuery{ActionID: stageActionIDForPractice(rig, request.GoalID), Limit: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if err := rig.projectEnactionEvents(events); err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.EventType == persistence.EventTypePracticeEvaluated {
			return event
		}
	}
	t.Fatal("missing terminal")
	return persistence.StoredCognitiveEvent{}
}

func TestPracticeReviewPassOnlyAchievement(t *testing.T) {
	for _, kind := range []string{"pass", "fail", "unknown"} {
		t.Run(kind, func(t *testing.T) {
			planner := reviewFixturePlanner()
			rig, pipeline := practiceProjectionRig(t, planner)
			request := admittedPracticeRequest(t, rig)
			answer := `{"task_id":"` + request.Practice.TaskID + `","sum":0,"product":0}`
			if kind == "pass" {
				raw, _ := json.Marshal(map[string]any{"task_id": request.Practice.TaskID, "sum": request.Practice.A + request.Practice.B, "product": request.Practice.A * request.Practice.B})
				answer = string(raw)
			}
			if kind == "unknown" {
				planner.err = context.DeadlineExceeded
			}
			event := reviewRun(t, rig, pipeline, planner, request, answer)
			goal := reviewGoal(t, rig, request.GoalID)
			if kind == "pass" {
				if goal.Status != GoalCompleted || goal.Progress != 1 || rig.echobeatsScheduler.goalsCompleted != 1 {
					t.Fatalf("pass not achieved: %#v", goal)
				}
			} else if goal.Status != GoalAbandoned || goal.Progress != 0 || rig.echobeatsScheduler.goalsCompleted != 0 {
				t.Fatalf("nonpass invented achievement: %#v", goal)
			}
			if !rig.practiceTerminal[request.GoalID] || len(rig.practiceTasks) != 0 {
				t.Fatal("trial not retired")
			}
			if err := rig.projectPracticeTerminal(event); err != nil {
				t.Fatal(err)
			}
			if kind == "pass" && rig.echobeatsScheduler.goalsCompleted != 1 {
				t.Fatal("duplicate terminal counted twice")
			}
		})
	}
}

func TestPracticeReviewOrdinaryTurnBetweenOutages(t *testing.T) {
	planner := reviewFixturePlanner()
	planner.err = context.DeadlineExceeded
	rig, pipeline := practiceProjectionRig(t, planner)
	ordinaryID := rig.echobeatsScheduler.AddGoal("ordinary pending work", 0.9)
	for i := range 3 {
		request := admittedPracticeRequest(t, rig)
		reviewRun(t, rig, pipeline, planner, request, "")
		before := rig.practiceIndex
		for range 3 {
			if err := rig.queueEvaluatedPractice(); err != nil {
				t.Fatal(err)
			}
		}
		if rig.practiceIndex != before || rig.echobeatsScheduler.GetActiveGoal().ID != ordinaryID {
			t.Fatal("practice monopolized ordinary lane")
		}
		// A real ordinary pipeline attempt durably records context even on outage.
		_, _ = pipeline.RunGoal(context.Background(), EnactionRequest{IdentityID: rig.identityID, SessionID: rig.sessionID, GoalID: fmt.Sprintf("ordinary-turn-%d", i), Goal: "ordinary work"})
		events, err := rig.eventStore.Query(context.Background(), persistence.CognitiveEventQuery{GoalID: fmt.Sprintf("ordinary-turn-%d", i), Limit: 1000})
		if err != nil {
			t.Fatal(err)
		}
		if err := rig.projectEnactionEvents(events); err != nil {
			t.Fatal(err)
		}
		if rig.lastOrdinaryEnactionSequence <= rig.lastPracticeTerminalSequence {
			t.Fatal("ordinary durable turn did not release budget")
		}
		if goal := reviewGoal(t, rig, ordinaryID); goal.Progress != 0 {
			t.Fatal("ordinary pending goal progress fabricated")
		}
	}
}

func TestPracticeReviewMalformedThenValidReplay(t *testing.T) {
	planner := reviewFixturePlanner()
	rig, pipeline := practiceProjectionRig(t, planner)
	bad := persistence.NewCognitiveEvent("malformed-admission", persistence.EventTypeGoalCreated, rig.identityID, rig.sessionID, evaluatedPracticeVersion, []byte(`{}`))
	bad.GoalID = bad.EventID
	bad.IdempotencyKey = bad.EventID
	if _, err := rig.eventStore.Append(context.Background(), bad); err != nil {
		t.Fatal(err)
	}
	request := admittedPracticeRequest(t, rig)
	raw, _ := json.Marshal(map[string]any{"task_id": request.Practice.TaskID, "sum": request.Practice.A + request.Practice.B, "product": request.Practice.A * request.Practice.B})
	reviewRun(t, rig, pipeline, planner, request, string(raw))
	fresh := &UnifiedAutonomousOrchestrator{ctx: context.Background(), identityID: rig.identityID, sessionID: "restart-session", eventStore: rig.eventStore, echobeatsScheduler: NewEchobeatsScheduler(nil), skillLearning: NewSkillLearningSystem(nil), config: rig.config}
	fresh.initializeOutcomeLearning()
	// Labels must work directly from the ledger before cognitive reconstruction.
	if err := fresh.replayOutcomeLearning(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := fresh.outcomeLearner.Snapshot().Counts; got.Training+got.HeldOut != 1 {
		t.Fatalf("valid later outcome lost: %#v", got)
	}
	if err := fresh.replayCognitiveEvidence(context.Background()); err != nil {
		t.Fatal(err)
	}
	if fresh.cognitiveReplaySequence == 0 {
		t.Fatal("malformed admission stopped replay")
	}
	stored, err := rig.eventStore.Get(context.Background(), bad.EventID)
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.projectPracticeGoal(stored); !errors.Is(err, errRejectedPracticeEvidence) {
		t.Fatalf("malformed admission not quarantinable: %v", err)
	}
}

func TestPracticeReviewDurableAdmissionPredecessor(t *testing.T) {
	t.Run("late", func(t *testing.T) {
		planner := reviewFixturePlanner()
		rig, pipeline := practiceProjectionRig(t, planner)
		task := generatedPracticeTask(rig.identityID, 1)
		goalID := "practice-goal:" + task.TaskID
		request := EnactionRequest{IdentityID: rig.identityID, SessionID: rig.sessionID, GoalID: goalID, Goal: practiceDescription(task), Practice: &task}
		planner.raw = testPlanJSON(t, "briefs/late-review.md", `{"task_id":"`+task.TaskID+`","sum":0,"product":0}`)
		if _, err := pipeline.RunGoal(context.Background(), request); err != nil {
			t.Fatal(err)
		}
		if err := rig.queueEvaluatedPractice(); err != nil {
			t.Fatal(err)
		}
		events, err := rig.eventStore.Query(context.Background(), persistence.CognitiveEventQuery{EventType: persistence.EventTypePracticeEvaluated, Limit: 1})
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := rig.validatedPractice(context.Background(), events[0]); !errors.Is(err, errRejectedPracticeEvidence) {
			t.Fatalf("late admission authorized action: %v", err)
		}
	})
	t.Run("noncanonical-and-transient", func(t *testing.T) {
		rig, _ := practiceProjectionRig(t, reviewFixturePlanner())
		request := admittedPracticeRequest(t, rig)
		admission, err := rig.eventStore.Get(context.Background(), request.GoalID)
		if err != nil {
			t.Fatal(err)
		}
		for _, mutate := range []func(*persistence.StoredCognitiveEvent){func(e *persistence.StoredCognitiveEvent) { e.Sequence++ }, func(e *persistence.StoredCognitiveEvent) { e.IdentityID = "foreign"; e.ContentSHA256 = "" }, func(e *persistence.StoredCognitiveEvent) { e.IdempotencyKey = "foreign"; e.ContentSHA256 = "" }, func(e *persistence.StoredCognitiveEvent) {
			e.PayloadJSON = []byte(`{"description":"bad","priority":1,"practice_index":1,"practice_task":{}}`)
			e.ContentSHA256 = ""
		}} {
			copy := admission
			mutate(&copy)
			if err := rig.projectPracticeGoal(copy); !errors.Is(err, errRejectedPracticeEvidence) {
				t.Fatalf("invalid admission accepted: %v", err)
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := rig.validatePracticeAdmission(ctx, admission); err == nil || errors.Is(err, errRejectedPracticeEvidence) {
			t.Fatalf("transient cancellation quarantined: %v", err)
		}
	})
}

func TestPracticeReviewDisabledRestorationAndHistoricalLearning(t *testing.T) {
	planner := reviewFixturePlanner()
	rig, pipeline := practiceProjectionRig(t, planner)
	request := admittedPracticeRequest(t, rig)
	fresh := &UnifiedAutonomousOrchestrator{ctx: context.Background(), identityID: rig.identityID, sessionID: "new-session", isAwake: true, eventStore: rig.eventStore, enactionPipeline: pipeline, echobeatsScheduler: NewEchobeatsScheduler(nil), skillLearning: NewSkillLearningSystem(nil), config: rig.config}
	fresh.config.EnableEvaluatedPractice = false
	fresh.initializeOutcomeLearning()
	if err := fresh.replayCognitiveEvidence(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(fresh.echobeatsScheduler.GetGoalQueue()) != 0 {
		t.Fatal("disabled restored task is runnable")
	}
	before := planner.calls
	fresh.runEnactionCycle()
	if planner.calls != before {
		t.Fatal("disabled restoration called planner")
	}
	planner.err = context.DeadlineExceeded
	reviewRun(t, rig, pipeline, planner, request, "")
	fresh.cognitiveEvidenceReplayed = false
	if err := fresh.replayCognitiveEvidence(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := fresh.replayOutcomeLearning(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !fresh.practiceTerminal[request.GoalID] || fresh.outcomeLearner.Snapshot().Counts.Unknown != 1 {
		t.Fatal("disabled history not reconstructed")
	}
}

func TestPracticeReviewBoundedReceiptsAndDurableLabels(t *testing.T) {
	planner := reviewFixturePlanner()
	rig, pipeline := practiceProjectionRig(t, planner)
	request := admittedPracticeRequest(t, rig)
	raw, _ := json.Marshal(map[string]any{"task_id": request.Practice.TaskID, "sum": request.Practice.A + request.Practice.B, "product": request.Practice.A * request.Practice.B})
	event := reviewRun(t, rig, pipeline, planner, request, string(raw))
	if err := rig.replayOutcomeLearning(context.Background()); err != nil {
		t.Fatal(err)
	}
	before := rig.outcomeLearner.Snapshot()
	rig.practiceMu.Lock()
	for i := range 300 {
		rig.rememberPracticeTerminalLocked(fmt.Sprintf("practice-goal:receipt-%d", i), int64(i))
	}
	rig.practiceMu.Unlock()
	if len(rig.practiceTerminal) > 256 || len(rig.practiceTerminalOrder) > 256 || len(rig.practiceTasks) != 0 {
		t.Fatal("practice state unbounded")
	}
	if _, _, err := rig.validatedPractice(context.Background(), event); err != nil {
		t.Fatalf("evicted receipt invalidated label: %v", err)
	}
	if err := rig.replayOutcomeLearning(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, rig.outcomeLearner.Snapshot()) {
		t.Fatal("receipt eviction retrained learner")
	}
}
