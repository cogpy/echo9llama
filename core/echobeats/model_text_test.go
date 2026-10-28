package echobeats

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/cogpy/echo9llama/core/llm"
)

type fakeTextGenerator struct {
	mu    sync.Mutex
	reply string
	err   error
	calls int
}

func (f *fakeTextGenerator) Generate(_ context.Context, _ string, _ llm.GenerateOptions) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	return f.reply, nil
}

type fixedInterest float64

func (f fixedInterest) GetInterestScore(string, string) float64 { return float64(f) }
func (f fixedInterest) IsInterested(_, _ string, threshold float64) bool {
	return float64(f) >= threshold
}

func TestEchoBeatsAutonomousThoughtRequiresModel(t *testing.T) {
	eb := NewEchoBeats()
	defer eb.cancel()

	eb.generateAutonomousThought()
	if eb.GetQueueSize() != 0 {
		t.Fatal("no model must mean no thought is scheduled")
	}

	failing := &fakeTextGenerator{err: errors.New("offline")}
	eb.SetThoughtGenerator(failing)
	eb.generateAutonomousThought()
	if failing.calls != 1 || eb.GetQueueSize() != 0 {
		t.Fatalf("model failure must schedule no thought: calls=%d queue=%d", failing.calls, eb.GetQueueSize())
	}

	eb.SetThoughtGenerator(&fakeTextGenerator{reply: "  a model thought  "})
	eb.generateAutonomousThought()
	if eb.GetQueueSize() != 1 {
		t.Fatalf("expected one scheduled thought, got %d", eb.GetQueueSize())
	}
	eb.mu.RLock()
	payload := (*eb.eventQueue)[0].Payload
	eb.mu.RUnlock()
	if payload != "a model thought" {
		t.Fatalf("thought payload was not model text: %#v", payload)
	}
}

func TestEchobeatsStepsRecordNoTextOnModelFailure(t *testing.T) {
	e := NewEchobeats()
	e.SetTextGenerator(&fakeTextGenerator{err: errors.New("offline")})
	cycle := &EchobeatsCycle{Steps: make([]*CognitiveStep, 12), Results: map[string]interface{}{}}
	steps := e.defineSteps()

	e.executeExpressiveStep(context.Background(), cycle, steps[0])
	e.executeReflectiveStep(context.Background(), cycle, steps[11])
	for _, step := range []*CognitiveStep{steps[0], steps[11]} {
		result := step.Result.(map[string]interface{})
		if _, ok := result["thought"]; ok {
			t.Fatalf("step %d recorded a thought despite model failure: %#v", step.StepNum, result)
		}
		if _, ok := result["reflection"]; ok {
			t.Fatalf("step %d recorded a reflection despite model failure: %#v", step.StepNum, result)
		}
		if result["error"] == nil {
			t.Fatalf("step %d should record the generation error: %#v", step.StepNum, result)
		}
	}

	e.SetTextGenerator(&fakeTextGenerator{reply: "model reflection"})
	e.executeReflectiveStep(context.Background(), cycle, steps[11])
	if got := steps[11].Result.(map[string]interface{})["reflection"]; got != "model reflection" {
		t.Fatalf("expected model reflection, got %#v", got)
	}
}

func TestDiscussionManagerSendsOnlyModelText(t *testing.T) {
	adm := NewAutonomousDiscussionManager(fixedInterest(0.9))
	defer adm.cancel()
	msg := IncomingMessage{Source: "user", Topic: "memory", Content: "What do you remember?"}

	adm.engageInDiscussion(msg)
	if out, ok := adm.GetOutgoingMessage(); ok {
		t.Fatalf("no generator must mean no reply, got %#v", out)
	}

	adm.SetResponseGenerator(func(context.Context, string, string) (string, error) {
		return "", errors.New("offline")
	})
	adm.engageInDiscussion(msg)
	if out, ok := adm.GetOutgoingMessage(); ok {
		t.Fatalf("generator failure must mean no reply, got %#v", out)
	}

	adm.SetResponseGenerator(func(context.Context, string, string) (string, error) {
		return "model reply", nil
	})
	adm.engageInDiscussion(msg)
	out, ok := adm.GetOutgoingMessage()
	if !ok || out.Content != "model reply" {
		t.Fatalf("expected model reply, got %#v ok=%v", out, ok)
	}
}

func TestInitiateDiscussionRequiresModelOpening(t *testing.T) {
	adm := NewAutonomousDiscussionManager(fixedInterest(0.9))
	defer adm.cancel()

	if err := adm.InitiateDiscussion("recursion", "peer"); err == nil {
		t.Fatal("expected error without an opening generator")
	}
	adm.SetOpeningGenerator(func(context.Context, string) (string, error) { return "", errors.New("offline") })
	if err := adm.InitiateDiscussion("recursion", "peer"); err == nil {
		t.Fatal("expected error when the opening generator fails")
	}
	if len(adm.activeDiscussions) != 0 || adm.discussionsInitiated != 0 {
		t.Fatal("failed openings must not start a discussion")
	}
	if out, ok := adm.GetOutgoingMessage(); ok {
		t.Fatalf("failed openings must send nothing, got %#v", out)
	}

	adm.SetOpeningGenerator(func(context.Context, string) (string, error) { return "model opening", nil })
	if err := adm.InitiateDiscussion("recursion", "peer"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := adm.GetOutgoingMessage()
	if !ok || out.Content != "model opening" {
		t.Fatalf("expected model opening, got %#v ok=%v", out, ok)
	}
}

func TestInsightProcessorRequiresModel(t *testing.T) {
	state := &CognitiveState{WorkingMemory: map[string]interface{}{"focus": "trees"}}

	if _, err := (&InsightGenerationProcessor{}).Process(context.Background(), state); err == nil {
		t.Fatal("expected error without a model")
	}
	if _, err := (&InsightGenerationProcessor{Generator: &fakeTextGenerator{err: errors.New("offline")}}).Process(context.Background(), state); err == nil {
		t.Fatal("expected error when the model fails")
	}
	result, err := (&InsightGenerationProcessor{Generator: &fakeTextGenerator{reply: "model insight"}}).Process(context.Background(), state)
	if err != nil || len(result.Insights) != 1 || result.Insights[0] != "model insight" {
		t.Fatalf("expected model insight, got %#v err=%v", result, err)
	}
}
