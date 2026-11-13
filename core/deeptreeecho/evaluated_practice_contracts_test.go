//go:build cgo

package deeptreeecho

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/cogpy/echo9llama/core/llm"
	"github.com/cogpy/echo9llama/core/persistence"
	"github.com/cogpy/echo9llama/core/tools"
)

func testArtifactPracticeTask() ArtifactPracticeTask {
	return ArtifactPracticeTask{TaskID: "practice-1", Version: ArithmeticArtifactVersion, A: -37, B: 12}
}

func testArtifactPracticeRequest() EnactionRequest {
	request := testEnactionRequest()
	task := testArtifactPracticeTask()
	request.Practice = &task
	return request
}

func TestArtifactPracticeExactJSON(t *testing.T) {
	task := testArtifactPracticeTask()
	for _, test := range []struct {
		name, content, reason string
		passed                bool
	}{
		{"object", `{"task_id":"practice-1","sum":-25,"product":-444}`, "exact_arithmetic_verified", true},
		{"fence", "```json\n{\"product\":-444,\"sum\":-25,\"task_id\":\"practice-1\"}\n```", "exact_arithmetic_verified", true},
		{"wrong_sum", `{"task_id":"practice-1","sum":25,"product":-444}`, "arithmetic_answer_mismatch", false},
		{"wrong_product", `{"task_id":"practice-1","sum":-25,"product":444}`, "arithmetic_answer_mismatch", false},
		{"task_mismatch", `{"task_id":"another","sum":-25,"product":-444}`, "task_id_mismatch", false},
		{"unknown", `{"task_id":"practice-1","sum":-25,"product":-444,"score":1}`, "invalid_arithmetic_json", false},
		{"duplicate", `{"task_id":"practice-1","sum":0,"sum":-25,"product":-444}`, "invalid_arithmetic_json", false},
		{"escaped_duplicate", `{"task_id":"practice-1","sum":0,"s\u0075m":-25,"product":-444}`, "invalid_arithmetic_json", false},
		{"trailing", `{"task_id":"practice-1","sum":-25,"product":-444} {}`, "invalid_arithmetic_json", false},
		{"trailing_prose", `{"task_id":"practice-1","sum":-25,"product":-444} done`, "invalid_arithmetic_json", false},
		{"missing", `{"sum":-25,"product":-444}`, "invalid_arithmetic_json", false},
		{"null", `{"task_id":"practice-1","sum":null,"product":-444}`, "invalid_arithmetic_json", false},
		{"fraction", `{"task_id":"practice-1","sum":-25.0,"product":-444}`, "invalid_arithmetic_json", false},
		{"exponent", `{"task_id":"practice-1","sum":-25,"product":-4.44e2}`, "invalid_arithmetic_json", false},
		{"quoted_number", `{"task_id":"practice-1","sum":"-25","product":-444}`, "invalid_arithmetic_json", false},
		{"array", `[{"task_id":"practice-1","sum":-25,"product":-444}]`, "invalid_arithmetic_json", false},
		{"prose", "The answer is -25 and -444.", "invalid_arithmetic_json", false},
		{"bare_fence", "```\n{}\n```", "invalid_arithmetic_json", false},
		{"multiple_fences", "```json\n{}\n```\n```json\n{}\n```", "invalid_arithmetic_json", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			passed, reason, err := EvaluateArtifactPracticeContent(task, []byte(test.content))
			if err != nil || passed != test.passed || reason != test.reason {
				t.Fatalf("evaluation=(%v,%q,%v), want (%v,%q,nil)", passed, reason, err, test.passed, test.reason)
			}
		})
	}
	for _, operands := range [][2]int64{{1_000_000, 1_000_000}, {-1_000_000, 1_000_000}, {-1_000_000, -1_000_000}, {0, 0}} {
		task.A, task.B = operands[0], operands[1]
		content, _ := json.Marshal(map[string]any{"task_id": task.TaskID, "sum": task.A + task.B, "product": task.A * task.B})
		if passed, _, err := EvaluateArtifactPracticeContent(task, content); err != nil || !passed {
			t.Fatalf("boundary exact arithmetic: passed=%v err=%v", passed, err)
		}
	}
	for _, bad := range []ArtifactPracticeTask{
		{TaskID: "x", Version: ArithmeticArtifactVersion, A: 1_000_001},
		{TaskID: "x", Version: ArithmeticArtifactVersion, B: -1_000_001},
		{TaskID: "x", Version: "model-scored"},
		{Version: ArithmeticArtifactVersion},
	} {
		if _, reason, err := EvaluateArtifactPracticeContent(bad, []byte(`{}`)); err == nil || reason != "" {
			t.Fatalf("invalid task acquired a correctness label: task=%#v reason=%s err=%v", bad, reason, err)
		}
	}
}

// HTTP fixture models only the provider boundary, not model intelligence. The
// production router, HTTP provider, event store, policy, and workspace tool run
// unchanged. No production fallback or fake arithmetic evaluator is installed.
func newRoutedPracticePipeline(t *testing.T, mode EnactionMode, content string, status int, onCall func()) (*EnactionPipeline, *persistence.CognitiveEventStore, string, *atomic.Int32, *atomic.Value) {
	t.Helper()
	plan := testPlanJSON(t, "briefs/arithmetic-1.md", content)
	calls := &atomic.Int32{}
	prompt := &atomic.Value{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var body struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Messages) == 0 {
			http.Error(w, "invalid provider request", http.StatusBadRequest)
			return
		}
		prompt.Store(body.Messages[len(body.Messages)-1].Content)
		if onCall != nil {
			onCall()
		}
		if status != http.StatusOK {
			http.Error(w, "provider unavailable", status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": plan}}}})
	}))
	t.Cleanup(server.Close)
	for _, key := range []string{"ANTHROPIC_API_KEY", "OPENROUTER_API_KEY", "OPENAI_API_KEY", "ECHO_MODEL_PATHS", "ECHO_MODEL_ROOTS", "LOCAL_MODEL_PATH"} {
		t.Setenv(key, "")
	}
	t.Setenv("OPENAI_BASE_URL", server.URL)
	router := llm.NewMultiProviderLLM()
	router.AddProvider(llm.NewOpenAIProvider("http-fixture-not-a-real-secret"))
	t.Cleanup(func() { _ = router.Close() })
	pipeline, store, root := newTestPipeline(t, mode, &scriptedActionPlanner{})
	pipeline.planner = LLMActionPlanner{Provider: router}
	return pipeline, store, root, calls, prompt
}

func getPracticeOutcome(t *testing.T, store *persistence.CognitiveEventStore, actionID string) (ArtifactPracticeOutcome, persistence.StoredCognitiveEvent) {
	t.Helper()
	event, err := store.Get(context.Background(), stageEventID(actionID, "practice-evaluation"))
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := ValidateArtifactPracticeOutcome(context.Background(), store, event)
	if err != nil {
		t.Fatalf("practice outcome provenance validation: %v", err)
	}
	return outcome, event
}

func TestArtifactPracticeActualPublicationAndCorrectnessAreSeparate(t *testing.T) {
	for _, test := range []struct {
		name, content string
		passed        bool
	}{
		{"correct", "```json\n{\"task_id\":\"practice-1\",\"sum\":-25,\"product\":-444}\n```", true},
		{"wrong_answer", `{"task_id":"practice-1","sum":25,"product":-444}`, false},
		{"invalid_json", "This is not arithmetic JSON.", false},
		{"duplicate", `{"task_id":"practice-1","sum":9,"sum":-25,"product":-444}`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			pipeline, store, root, calls, prompt := newRoutedPracticePipeline(t, EnactionLocalSandbox, test.content, http.StatusOK, nil)
			request := testArtifactPracticeRequest()
			result, err := pipeline.RunGoal(context.Background(), request)
			if err != nil || !result.Verified || !result.Executed {
				t.Fatalf("file publication failed: result=%#v err=%v", result, err)
			}
			practice, event := getPracticeOutcome(t, store, result.ActionID)
			if practice.Passed == nil || *practice.Passed != test.passed {
				t.Fatalf("incorrect exact label: %#v", practice)
			}
			actual, err := os.ReadFile(filepath.Join(root, "briefs/arithmetic-1.md"))
			if err != nil || string(actual) != test.content || sha256Hex(actual) != practice.ArtifactSHA256 {
				t.Fatalf("actual artifact/hash: %q err=%v", actual, err)
			}
			info, _ := os.Stat(filepath.Join(root, "briefs/arithmetic-1.md"))
			if info.Mode().Perm() != 0o600 {
				t.Fatalf("artifact mode=%v", info.Mode())
			}
			if calls.Load() != 1 || !strings.Contains(prompt.Load().(string), "ARITHMETIC PRACTICE OVERRIDE") {
				t.Fatal("practice did not use routed arithmetic prompt")
			}
			if event.CausationID != result.EvaluationEventID {
				t.Fatal("practice lacks canonical evaluation causal link")
			}
			ctxEvent, _ := store.Get(context.Background(), stageEventID(result.ActionID, "context"))
			storedTask, err := storedArtifactPractice(ctxEvent)
			if err != nil || storedTask == nil || *storedTask != *request.Practice {
				t.Fatal("task not durably preserved")
			}
			skills, _ := store.Query(context.Background(), persistence.CognitiveEventQuery{ActionID: result.ActionID, EventType: persistence.EventTypeSkillEvidence, Limit: 1000})
			if len(skills) != 1 || !strings.Contains(string(skills[0].PayloadJSON), structuredArtifactSkill) || strings.Contains(string(skills[0].PayloadJSON), ArithmeticArtifactSkillName) {
				t.Fatal("file construction was conflated with arithmetic correctness")
			}
		})
	}
}

func TestArtifactPracticeUnavailableAndObserveAreUnknown(t *testing.T) {
	for _, test := range []struct {
		name   string
		mode   EnactionMode
		status int
		cancel bool
	}{
		{"observe", EnactionObserve, http.StatusOK, false},
		{"provider_outage", EnactionLocalSandbox, http.StatusServiceUnavailable, false},
		{"provider_cancel", EnactionLocalSandbox, http.StatusOK, true},
		{"execution_unavailable", EnactionLocalSandbox, http.StatusOK, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var onCall func()
			if test.cancel {
				onCall = cancel
			}
			pipeline, store, root, _, _ := newRoutedPracticePipeline(t, test.mode, `{"task_id":"practice-1","sum":-25,"product":-444}`, test.status, onCall)
			if test.name == "execution_unavailable" {
				pipeline.tool.Root = "not-an-absolute-private-root"
			}
			result, err := pipeline.RunGoal(ctx, testArtifactPracticeRequest())
			if test.name != "observe" && err == nil {
				t.Fatal("unavailable boundary should fail closed")
			}
			practice, event := getPracticeOutcome(t, store, result.ActionID)
			if practice.Passed != nil || practice.ArtifactSHA256 != "" || practice.EvaluationEventID != "" || !strings.HasPrefix(practice.Reason, "not_evaluated:") {
				t.Fatalf("unknown acquired label/evidence: %#v", practice)
			}
			if strings.Contains(practice.Reason, "arithmetic_answer_mismatch") || strings.Contains(practice.Reason, "invalid_arithmetic_json") {
				t.Fatal("unevaluated content got a wrong-content label")
			}
			if _, err := os.Stat(filepath.Join(root, "briefs/arithmetic-1.md")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("unexpected file effect: %v", err)
			}
			for _, kind := range []string{persistence.EventTypeEvaluationRecorded, persistence.EventTypeSkillEvidence, persistence.EventTypeGoalProgressed} {
				count, err := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: result.ActionID, EventType: kind})
				if err != nil || count != 0 {
					t.Fatalf("unknown has %s labels: %d %v", kind, count, err)
				}
			}
			if event.CausationID == "" {
				t.Fatal("unknown lacks error/denial cause")
			}
		})
	}
}

func TestArtifactPracticeRestartExactlyOnceAndImmutableTask(t *testing.T) {
	pipeline, store, root, calls, _ := newRoutedPracticePipeline(t, EnactionLocalSandbox, `{"task_id":"practice-1","sum":-25,"product":-444}`, http.StatusOK, nil)
	ledgerPath := filepath.Join(t.TempDir(), "restart-ledger.db")
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err := persistence.NewCognitiveEventStore(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	pipeline.store = store
	request := testArtifactPracticeRequest()
	first, err := pipeline.RunGoal(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(filepath.Join(root, "briefs/arithmetic-1.md"))
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = persistence.NewCognitiveEventStore(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	config := DefaultActionPolicyConfig()
	config.Mode = EnactionLocalSandbox
	restarted, err := NewEnactionPipeline(store, pipeline.tool, NewActionPolicy(config), pipeline.planner, request.IdentityID, "restart-session")
	if err != nil {
		t.Fatal(err)
	}
	request.SessionID = "restart-session"
	second, err := restarted.RunGoal(context.Background(), request)
	if err != nil || second.ActionID != first.ActionID || !second.Recovered {
		t.Fatalf("restart: result=%#v err=%v", second, err)
	}
	for _, kind := range []string{"practice.evaluated", persistence.EventTypeActionStarted, persistence.EventTypeActionCompleted} {
		count, err := store.Count(context.Background(), persistence.CognitiveEventQuery{GoalID: request.GoalID, EventType: kind})
		if err != nil || count != 1 {
			t.Fatalf("restart %s count=%d err=%v", kind, count, err)
		}
	}
	after, _ := os.Stat(filepath.Join(root, "briefs/arithmetic-1.md"))
	if !os.SameFile(info, after) || !info.ModTime().Equal(after.ModTime()) || calls.Load() != 1 {
		t.Fatal("restart repeated effect/model call")
	}
	changed := *request.Practice
	changed.A++
	request.Practice = &changed
	if _, err := restarted.RunGoal(context.Background(), request); err == nil {
		t.Fatal("mutable task substitution accepted")
	}
	request.Practice = nil
	if _, err := restarted.RunGoal(context.Background(), request); err == nil {
		t.Fatal("dropping committed practice task accepted")
	}
	fresh, err := NewEnactionPipeline(store, pipeline.tool, NewActionPolicy(config), pipeline.planner, request.IdentityID, "fresh-without-task")
	if err != nil {
		t.Fatal(err)
	}
	request.SessionID = "fresh-without-task"
	if _, err := fresh.RunGoal(context.Background(), request); err == nil {
		t.Fatal("fresh restart substituted a generic task for committed practice")
	}
	if calls.Load() != 1 {
		t.Fatal("task substitution crossed planner boundary")
	}
}

func seedTerminalPracticeWithoutOutcome(t *testing.T, pipeline *EnactionPipeline, request EnactionRequest) (string, ActionPlan) {
	t.Helper()
	ctx := context.Background()
	actionID := "action:" + request.GoalID + ":workspace.create_note:local-sandbox:iteration-1"
	correlationID := "enaction:" + request.GoalID
	plan, err := ParseActionPlanStrict(testPlanJSON(t, "briefs/replay-practice-1.md", `{"task_id":"practice-1","sum":-25,"product":-444}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pipeline.appendEvent(ctx, stageEventID(actionID, "context"), persistence.EventTypeContextRetrieved, correlationID, "", request.GoalID, actionID, persistence.EvidenceClassSensitive, map[string]any{"practice": *request.Practice, "goal": request.Goal}, PlanRouteEvidence{}); err != nil {
		t.Fatal(err)
	}
	route := PlanRouteEvidence{TraceID: "seeded-durable-route", Provider: "openai", BackendKind: "remote_api", ModelID: "routed-fixture-model"}
	seedReplayChain(t, pipeline, plan, route, request, actionID, correlationID)
	observed, err := pipeline.tool.Create(ctx, tools.WorkspaceNoteRequest{Path: plan.RelativePath, Content: []byte(plan.Content)})
	if err != nil {
		t.Fatal(err)
	}
	observation := observationFromTool(plan, observed, false, nil)
	if err := pipeline.appendCompletion(ctx, actionID, correlationID, request.GoalID, observation, route); err != nil {
		t.Fatal(err)
	}
	evaluation, err := pipeline.recordEvaluation(ctx, plan, observation, actionID, correlationID, request.GoalID)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := pipeline.recordSuccessProjections(ctx, plan, observation, evaluation, actionID, correlationID, request.GoalID, request.Goal); err != nil {
		t.Fatal(err)
	}
	return actionID, plan
}

func TestArtifactPracticeTerminalReplayFinalizesMissingOutcomeWithoutEffects(t *testing.T) {
	for _, state := range []string{"present", "missing", "conflicting"} {
		t.Run(state, func(t *testing.T) {
			planner := &scriptedActionPlanner{err: errors.New("must not infer on terminal replay")}
			pipeline, store, root := newTestPipeline(t, EnactionLocalSandbox, planner)
			request := testArtifactPracticeRequest()
			actionID, plan := seedTerminalPracticeWithoutOutcome(t, pipeline, request)
			path := filepath.Join(root, filepath.FromSlash(plan.RelativePath))
			if state == "missing" {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			}
			if state == "conflicting" {
				if err := os.WriteFile(path, []byte("unverified conflicting content"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			config := DefaultActionPolicyConfig()
			config.Mode = EnactionLocalSandbox
			restarted, err := NewEnactionPipeline(store, pipeline.tool, NewActionPolicy(config), planner, request.IdentityID, "terminal-restart-session")
			if err != nil {
				t.Fatal(err)
			}
			pipeline = restarted
			request.SessionID = "terminal-restart-session"
			result, err := pipeline.RunGoal(context.Background(), request)
			if err != nil || !result.Recovered {
				t.Fatalf("terminal replay=%#v err=%v", result, err)
			}
			practice, practiceEvent := getPracticeOutcome(t, store, actionID)
			if practiceEvent.SessionID != "session-test" {
				t.Fatal("practice finalization lost the immutable action session")
			}
			if state != "present" && practice.Passed != nil {
				t.Fatalf("unverified actual content acquired label: %#v", practice)
			}
			if state == "present" && (practice.Passed == nil || !*practice.Passed) {
				t.Fatalf("verified terminal content not evaluated: %#v", practice)
			}
			if planner.calls != 0 {
				t.Fatal("terminal replay called model")
			}
			if state == "missing" {
				if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("terminal replay recreated missing file")
				}
			}
			if state == "conflicting" {
				actual, err := os.ReadFile(path)
				if err != nil || string(actual) != "unverified conflicting content" {
					t.Fatal("terminal replay rewrote conflicting file")
				}
			}
			if _, err := pipeline.RunGoal(context.Background(), request); err != nil {
				t.Fatal(err)
			}
			count, _ := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: actionID, EventType: "practice.evaluated"})
			if count != 1 {
				t.Fatalf("outcome count=%d", count)
			}
		})
	}
}

func TestArtifactPracticeProvenanceRejectsForgedDurableLabel(t *testing.T) {
	pipeline, store, _ := newTestPipeline(t, EnactionLocalSandbox, &scriptedActionPlanner{})
	request := testArtifactPracticeRequest()
	actionID, plan := seedTerminalPracticeWithoutOutcome(t, pipeline, request)
	wrong := false
	payload := ArtifactPracticeOutcome{Task: *request.Practice, Passed: &wrong, Reason: "arithmetic_answer_mismatch", EvaluatorVersion: ArithmeticArtifactVersion, EvaluationEventID: stageEventID(actionID, "evaluation"), ArtifactSHA256: sha256Hex([]byte(plan.Content))}
	if _, err := pipeline.appendEvent(context.Background(), stageEventID(actionID, "practice-evaluation"), "practice.evaluated", "enaction:"+request.GoalID, payload.EvaluationEventID, request.GoalID, actionID, persistence.EvidenceClassSensitive, payload, PlanRouteEvidence{}); err != nil {
		t.Fatal(err)
	}
	event, err := store.Get(context.Background(), stageEventID(actionID, "practice-evaluation"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateArtifactPracticeOutcome(context.Background(), store, event); err == nil {
		t.Fatal("durable wrong label was admitted despite exact evidence")
	}
}

func TestArtifactPracticeGenericPlannerPromptUnchangedAndNoPracticeEvent(t *testing.T) {
	pipeline, store, _, _, prompt := newRoutedPracticePipeline(t, EnactionLocalSandbox, "# Generic evidence questions", http.StatusOK, nil)
	result, err := pipeline.RunGoal(context.Background(), testEnactionRequest())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(prompt.Load().(string), "ARITHMETIC PRACTICE OVERRIDE") {
		t.Fatal("practice enabled for generic request")
	}
	count, err := store.Count(context.Background(), persistence.CognitiveEventQuery{ActionID: result.ActionID, EventType: "practice.evaluated"})
	if err != nil || count != 0 {
		t.Fatalf("generic practice events=%d err=%v", count, err)
	}
}

func TestArtifactPracticeInvalidTaskAndForeignContextFailBeforeEffects(t *testing.T) {
	for _, foreignContext := range []bool{false, true} {
		t.Run(map[bool]string{false: "invalid_bounds", true: "foreign_context"}[foreignContext], func(t *testing.T) {
			planner := &scriptedActionPlanner{raw: testPlanJSON(t, "briefs/must-not-exist.md", `{"task_id":"practice-1","sum":-25,"product":-444}`), route: PlanRouteEvidence{TraceID: "must-not-route", Provider: "openai", BackendKind: "remote_api"}}
			pipeline, store, root := newTestPipeline(t, EnactionLocalSandbox, planner)
			request := testArtifactPracticeRequest()
			if foreignContext {
				actionID := "action:" + request.GoalID + ":workspace.create_note:local-sandbox:iteration-1"
				payload, _ := json.Marshal(map[string]any{"practice": *request.Practice, "goal": request.Goal})
				event := persistence.NewCognitiveEvent(stageEventID(actionID, "context"), persistence.EventTypeContextRetrieved, "another-identity", "another-session", EnactionPolicyVersion, payload)
				event.ActionID, event.GoalID, event.CorrelationID = actionID, request.GoalID, "enaction:"+request.GoalID
				event.IdempotencyKey = event.EventID
				if _, err := store.Append(context.Background(), event); err != nil {
					t.Fatal(err)
				}
			} else {
				request.Practice.A = ArtifactPracticeOperandLimit + 1
			}
			if _, err := pipeline.RunGoal(context.Background(), request); err == nil {
				t.Fatal("invalid task/source crossed admission")
			}
			if planner.calls != 0 {
				t.Fatal("invalid task/source called planner")
			}
			if _, err := os.Stat(root); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("invalid task/source created workspace: %v", err)
			}
			count, err := store.Count(context.Background(), persistence.CognitiveEventQuery{EventType: "practice.evaluated"})
			if err != nil || count != 0 {
				t.Fatalf("invalid canonical task received a label: %d %v", count, err)
			}
		})
	}
}
