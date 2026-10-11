//go:build cgo

package deeptreeecho

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/backendcap"
	"github.com/cogpy/echo9llama/core/echodream"
	"github.com/cogpy/echo9llama/core/llm"
	"github.com/cogpy/echo9llama/core/metacognition"
	"github.com/cogpy/echo9llama/core/persistence"
)

// Scripted test fixtures validate plumbing, not cognition or live inference.
type reflectionFixtureProvider struct {
	calls     atomic.Int64
	invalid   bool
	degraded  bool
	block     chan struct{}
	entered   chan struct{}
	liveBase  string
	liveModel string
	liveRaw   string
}

func (p *reflectionFixtureProvider) Name() string {
	if p.liveModel != "" {
		return "live-openai-compatible-test-client"
	}
	return "reflection-test-fixture"
}
func (p *reflectionFixtureProvider) Available() bool { return true }
func (p *reflectionFixtureProvider) MaxTokens() int  { return 16384 }
func (p *reflectionFixtureProvider) StreamGenerate(context.Context, string, llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	return nil, errors.New("not used")
}

func (p *reflectionFixtureProvider) GetRouteResult(id string) (llm.BackendRoutingState, bool) {
	model := "fixture-model"
	if p.liveModel != "" {
		model = p.liveModel
	}
	return llm.BackendRoutingState{SelectedProvider: p.Name(), SelectedKind: backendcap.BackendRemoteAPI, SelectedModelID: model, Degraded: p.degraded}, true
}

func (p *reflectionFixtureProvider) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	p.calls.Add(1)
	if !opts.Routing.RequireRealModel || opts.Routing.AllowFallback || opts.Routing.Intent != "echobeats.evidence_reflection" {
		return "", errors.New("reflection routing contract not enforced")
	}
	if p.entered != nil {
		close(p.entered)
	}
	if p.block != nil {
		select {
		case <-p.block:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	if p.liveBase != "" {
		// Gemini's budget includes reasoning tokens; leave room for complete JSON.
		payload, err := json.Marshal(map[string]any{"model": p.liveModel, "messages": []map[string]string{{"role": "user", "content": prompt}}, "max_tokens": 8192})
		if err != nil {
			return "", err
		}
		request, err := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(p.liveBase, "/")+"/chat/completions", bytes.NewReader(payload))
		if err != nil {
			return "", err
		}
		request.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
		request.Header.Set("Content-Type", "application/json")
		response, err := (&http.Client{Timeout: 40 * time.Second}).Do(request)
		if err != nil {
			return "", errors.New("live model request failed")
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return "", fmt.Errorf("live model HTTP %d", response.StatusCode)
		}
		var result struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.NewDecoder(io.LimitReader(response.Body, 128*1024)).Decode(&result); err != nil {
			return "", err
		}
		if len(result.Choices) == 0 {
			return "", errors.New("live model returned no choices")
		}
		p.liveRaw = result.Choices[0].Message.Content
		return result.Choices[0].Message.Content, nil
	}
	index := strings.Index(prompt, "MEASURED EVIDENCE:\n")
	if index < 0 {
		return "", errors.New("missing measured context")
	}
	var report metacognition.Report
	if err := json.Unmarshal([]byte(prompt[index+len("MEASURED EVIDENCE:\n"):]), &report); err != nil {
		return "", err
	}
	if len(report.Evidence) == 0 {
		return "", errors.New("empty evidence should not generate")
	}
	source := report.Evidence[0].EventID
	if p.invalid {
		source = "invented-source"
	}
	result, err := json.Marshal(metacognition.Reflection{Lesson: "Exact read-back matched the requested artifact in this trial.", Uncertainty: "This verifies bytes, not the semantic correctness of the note.", Experiment: "Create a private note specifying a falsifiable evaluation for the next practice trial.", SourceEventIDs: []string{source}})
	return string(result), err
}

func newMeasuredReflectionFixture(t *testing.T, provider *reflectionFixtureProvider) (*UnifiedAutonomousOrchestrator, *EnactionPipeline) {
	t.Helper()
	planner := &scriptedActionPlanner{raw: testPlanJSON(t, "briefs/reflection.md", "# Evidence\n\nWhat would falsify this claim?\n"), route: PlanRouteEvidence{TraceID: "fixture", Provider: "fixture", BackendKind: "remote_api", ModelID: "fixture-model"}}
	pipeline, store, _ := newTestPipeline(t, EnactionLocalSandbox, planner)
	u := &UnifiedAutonomousOrchestrator{ctx: context.Background(), isAwake: true, identityID: "echo-test", sessionID: "session-test", eventStore: store, llmProvider: provider, echobeatsScheduler: NewEchobeatsScheduler(provider), dreamCycle: echodream.NewSleepWakeStateMachine(), ingestedExperienceIDs: make(map[string]struct{})}
	u.echobeatsScheduler.SetOnReflection(u.runEvidenceReflection)
	t.Cleanup(func() { u.dreamCycle.Shutdown() })
	if _, err := pipeline.RunGoal(context.Background(), testEnactionRequest()); err != nil {
		t.Fatal(err)
	}
	return u, pipeline
}

func TestEvidenceReflectionDurableReplayAndNoWisdomAward(t *testing.T) {
	provider := &reflectionFixtureProvider{}
	u, _ := newMeasuredReflectionFixture(t, provider)
	if err := u.runEvidenceReflection(8); err != nil {
		t.Fatal(err)
	}
	if err := u.runEvidenceReflection(8); err != nil {
		t.Fatal(err)
	}
	if provider.calls.Load() != 1 {
		t.Fatalf("same evidence generated %d calls", provider.calls.Load())
	}
	status := u.GetStatus()
	if status.Reflection.Evaluated != 1 || status.Reflection.Passed != 1 || status.Reflection.CalibrationSamples != 1 {
		t.Fatalf("wrong measured status: %+v", status.Reflection)
	}
	if status.Reflection.BrierScore == nil || *status.Reflection.BrierScore < .0099 || *status.Reflection.BrierScore > .0101 {
		t.Fatalf("wrong Brier: %+v", status.Reflection)
	}
	if status.TotalWisdom != 0 || status.WisdomDepth != 0 {
		t.Fatal("generated lesson awarded wisdom")
	}
	goals := u.echobeatsScheduler.GetGoalQueue()
	if len(goals) != 1 || goals[0].Progress != 0 || goals[0].Status != GoalPending {
		t.Fatalf("reflection must only propose a goal: %+v", goals)
	}
	if status.PendingExperiences != 1 {
		t.Fatalf("reflection not fed once to dream: %d", status.PendingExperiences)
	}
	// Simulate a fresh process session bound to the same durable identity/store.
	second := &UnifiedAutonomousOrchestrator{ctx: context.Background(), isAwake: true, identityID: u.identityID, sessionID: "next-session", eventStore: u.eventStore, llmProvider: provider, echobeatsScheduler: NewEchobeatsScheduler(provider), dreamCycle: echodream.NewSleepWakeStateMachine(), ingestedExperienceIDs: make(map[string]struct{})}
	defer second.dreamCycle.Shutdown()
	if err := second.replayCognitiveEvidence(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := second.runEvidenceReflection(8); err != nil {
		t.Fatal(err)
	}
	if err := second.runEvidenceReflection(8); err != nil {
		t.Fatal(err)
	}
	if provider.calls.Load() != 1 {
		t.Fatal("restart regenerated persisted reflection")
	}
	if second.GetStatus().Reflection.Evaluated != 1 || len(second.echobeatsScheduler.GetGoalQueue()) != 2 {
		t.Fatalf("replay lost measured evidence/goal: %+v", second.GetStatus().Reflection)
	}
	events, err := u.eventStore.Query(context.Background(), persistence.CognitiveEventQuery{EventType: persistence.EventTypeReflectionProposed, Limit: 10})
	if err != nil || len(events) != 1 {
		t.Fatalf("durable reflections: %d %v", len(events), err)
	}
	if events[0].EvidenceClass != persistence.EvidenceClassSensitive {
		t.Fatal("reflection handling not sensitive")
	}
	// Tampered claim must fail even when the event envelope gets a new hash.
	var evidence reflectionEvidence
	if err := json.Unmarshal(events[0].PayloadJSON, &evidence); err != nil {
		t.Fatal(err)
	}
	evidence.Report.Passed = 0
	events[0].PayloadJSON, _ = json.Marshal(evidence)
	if err := u.projectReflection(context.Background(), events[0]); err == nil {
		t.Fatal("tampered report accepted")
	}
}

func TestEvidenceReflectionRejectsInvalidAndDegradedOutputOnce(t *testing.T) {
	for _, tc := range []struct {
		name              string
		invalid, degraded bool
	}{{"invented citation", true, false}, {"degraded route", false, true}} {
		t.Run(tc.name, func(t *testing.T) {
			p := &reflectionFixtureProvider{invalid: tc.invalid, degraded: tc.degraded}
			u, _ := newMeasuredReflectionFixture(t, p)
			if err := u.runEvidenceReflection(8); err == nil {
				t.Fatal("invalid output accepted")
			}
			if err := u.runEvidenceReflection(8); err != nil {
				t.Fatal(err)
			}
			if p.calls.Load() != 1 || len(u.echobeatsScheduler.GetGoalQueue()) != 0 {
				t.Fatal("failure repeated or scheduled action")
			}
			count, err := u.eventStore.Count(context.Background(), persistence.CognitiveEventQuery{EventType: persistence.EventTypeErrorRecorded})
			if err != nil || count != 1 {
				t.Fatalf("failure not durable: %d %v", count, err)
			}
			// Failed reflection itself cannot recursively become measured evidence.
			if u.GetStatus().Reflection.EvidenceCount != 1 {
				t.Fatal("self-feedback contaminated outcome evidence")
			}
		})
	}
}

func TestEvidenceReflectionStatusAndPauseDoNotRaceInference(t *testing.T) {
	p := &reflectionFixtureProvider{block: make(chan struct{}), entered: make(chan struct{})}
	u, _ := newMeasuredReflectionFixture(t, p)
	u.echobeatsScheduler.currentStep = 8
	done := make(chan struct{})
	go func() { u.echobeatsScheduler.executeStep(); close(done) }()
	<-p.entered
	statusDone := make(chan struct{})
	go func() { _ = u.GetStatus(); close(statusDone) }()
	select {
	case <-statusDone:
	case <-time.After(time.Second):
		t.Fatal("status blocked on inference")
	}
	pauseDone := make(chan struct{})
	go func() { u.echobeatsScheduler.Pause(); close(pauseDone) }()
	select {
	case <-pauseDone:
		t.Fatal("Pause returned before reflection quiesced")
	case <-time.After(20 * time.Millisecond):
	}
	close(p.block)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("step did not finish")
	}
	select {
	case <-pauseDone:
	case <-time.After(time.Second):
		t.Fatal("Pause did not finish")
	}
	callsAfterPause := p.calls.Load()
	u.echobeatsScheduler.executeStep()
	if p.calls.Load() != callsAfterPause {
		t.Fatal("paused scheduler admitted inference")
	}
}

func TestEvidenceReflectionEmptyRestAndForeignEvidenceAreIgnored(t *testing.T) {
	p := &reflectionFixtureProvider{}
	u, _ := newMeasuredReflectionFixture(t, p)
	u.isAwake = false
	if err := u.runEvidenceReflection(8); err != nil {
		t.Fatal(err)
	}
	if p.calls.Load() != 0 {
		t.Fatal("rest admitted reflection")
	}
	u.isAwake = true
	u.identityID = "another-identity"
	if err := u.runEvidenceReflection(8); err != nil {
		t.Fatal(err)
	}
	if p.calls.Load() != 0 || u.GetStatus().Reflection.SuccessRate != nil {
		t.Fatal("foreign/empty evidence invented rate or inference")
	}
}

func TestEvidenceReflectionRejectsMissingCompletion(t *testing.T) {
	p := &reflectionFixtureProvider{}
	u, _ := newMeasuredReflectionFixture(t, p)
	payload, _ := json.Marshal(ActionEvaluation{Passed: true, Score: 1, Criteria: []string{"a", "b", "c", "d"}, EvidenceEventID: "orphan:evaluation", Reason: "test"})
	event := persistence.NewCognitiveEvent("orphan:evaluation", persistence.EventTypeEvaluationRecorded, u.identityID, u.sessionID, EnactionPolicyVersion, payload)
	event.ActionID = "orphan"
	event.CausationID = "orphan:completed"
	result, err := u.eventStore.Append(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := u.measuredReflectionEvidence(context.Background(), persistence.StoredCognitiveEvent{Sequence: result.Sequence, CognitiveEvent: event}); err == nil {
		t.Fatal("orphan evaluation admitted")
	}
}

// Optional live test performs a real HTTP inference over real local filesystem
// read-back evidence; it makes no claim that the generated lesson is true.
func TestEvidenceReflectionLiveModel(t *testing.T) {
	model := os.Getenv("ECHO_REFLECTION_LIVE_MODEL")
	if model == "" {
		t.Skip("set ECHO_REFLECTION_LIVE_MODEL and OPENAI_API_BASE for live inference")
	}
	if os.Getenv("OPENAI_API_KEY") == "" || os.Getenv("OPENAI_API_BASE") == "" {
		t.Fatal("live endpoint and key required")
	}
	p := &reflectionFixtureProvider{liveBase: os.Getenv("OPENAI_API_BASE"), liveModel: model}
	u, _ := newMeasuredReflectionFixture(t, p)
	if err := u.runEvidenceReflection(8); err != nil {
		t.Logf("Rejected public model output: %s", p.liveRaw)
		t.Fatal(err)
	}
	events, err := u.eventStore.Query(context.Background(), persistence.CognitiveEventQuery{EventType: persistence.EventTypeReflectionProposed, Limit: 10})
	if err != nil || len(events) != 1 {
		t.Fatalf("live reflection persistence: %v %d", err, len(events))
	}
	var evidence reflectionEvidence
	if err := json.Unmarshal(events[0].PayloadJSON, &evidence); err != nil {
		t.Fatal(err)
	}
	t.Logf("Live model %s: schema-valid public lesson; %d valid source citations; artifact Brier %.6f; no competence award", model, len(evidence.Reflection.SourceEventIDs), *evidence.Report.BrierScore)
	t.Logf("Proposed lesson: %s\nUncertainty: %s\nProposed practice: %s", evidence.Reflection.Lesson, evidence.Reflection.Uncertainty, evidence.Reflection.Experiment)
}
