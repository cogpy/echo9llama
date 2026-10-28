package deeptreeecho

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/cogpy/echo9llama/core/llm"
)

// modelOnlyTestProvider returns a fixed reply or error for every call.
type modelOnlyTestProvider struct {
	mu    sync.Mutex
	reply string
	err   error
	calls int
}

func (p *modelOnlyTestProvider) Generate(context.Context, string, llm.GenerateOptions) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	if p.err != nil {
		return "", p.err
	}
	return p.reply, nil
}

func (p *modelOnlyTestProvider) StreamGenerate(context.Context, string, llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	return nil, errors.New("not implemented")
}
func (p *modelOnlyTestProvider) Name() string    { return "model-only-test" }
func (p *modelOnlyTestProvider) Available() bool { return true }
func (p *modelOnlyTestProvider) MaxTokens() int  { return 4096 }

func failingProvider() *modelOnlyTestProvider {
	return &modelOnlyTestProvider{err: llm.ErrNoEdgeModel}
}

func TestStreamOfConsciousnessModelErrorYieldsNoThought(t *testing.T) {
	soc := NewStreamOfConsciousness(failingProvider())
	soc.generateThought()
	if n := len(soc.GetAllThoughts()); n != 0 {
		t.Fatalf("expected no thought on model error, got %d", n)
	}

	ok := NewStreamOfConsciousness(&modelOnlyTestProvider{reply: "thought-from-model-91c"})
	ok.generateThought()
	thoughts := ok.GetAllThoughts()
	if len(thoughts) != 1 || thoughts[0].Content != "thought-from-model-91c" {
		t.Fatalf("expected the model's thought, got %+v", thoughts)
	}
}

func TestHeartbeatIntrospectionModelErrorYieldsNoText(t *testing.T) {
	ah := NewAutonomousHeartbeat(failingProvider())
	reflection, insight, focus := ah.performIntrospection()
	if reflection != "" || insight != nil || focus != "" {
		t.Fatalf("expected no reflection/insight/focus on model error, got %q %v %q", reflection, insight, focus)
	}

	ah = NewAutonomousHeartbeat(&modelOnlyTestProvider{
		reply: "REFLECTION: model-reflection-x | INSIGHT: none | FOCUS: model-focus-y",
	})
	reflection, insight, focus = ah.performIntrospection()
	if reflection != "model-reflection-x" {
		t.Fatalf("reflection = %q", reflection)
	}
	if insight != nil {
		t.Fatalf("expected no insight for 'none', got %+v", insight)
	}
	if focus != "model-focus-y" {
		t.Fatalf("focus = %q", focus)
	}
}

func TestEchobeatsModelErrorMarksTaskFailed(t *testing.T) {
	sched := NewEchobeatsScheduler(failingProvider())
	sched.relevanceRealization("what matters?")
	sched.salienceSimulation(1)
	sched.affordanceInteraction(2)

	sched.mu.RLock()
	commitment := sched.presentCommitment
	past := len(sched.pastPerformance)
	future := len(sched.futureAnticipation)
	sched.mu.RUnlock()
	if commitment != "" || past != 0 || future != 0 {
		t.Fatalf("expected no model-free text, got commitment=%q past=%d future=%d", commitment, past, future)
	}
	for _, engine := range []*InferenceEngine{sched.engine1, sched.engine2, sched.engine3} {
		for _, task := range engine.taskHistory {
			if task.Success || task.Result != "" {
				t.Fatalf("task %s: expected Success=false and empty result, got %+v", task.ID, task)
			}
		}
	}
}

func TestTetrahedralModelErrorMarksTaskFailed(t *testing.T) {
	sched := NewEchobeatsTetrahedralScheduler(failingProvider())
	sched.relevanceRealizationTetrahedral("what matters?", sched.engine1)
	sched.affordanceInteractionTetrahedral(2)
	sched.salienceSimulationTetrahedral(1)

	if sched.presentCommitment != "" || len(sched.pastPerformance) != 0 || len(sched.futureAnticipation) != 0 {
		t.Fatal("expected no model-free text after model errors")
	}
	for _, engine := range []*TetrahedralEngine{sched.engine1, sched.engine2, sched.engine3, sched.engine4} {
		for _, task := range engine.taskHistory {
			if task.Success || task.Result != "" {
				t.Fatalf("task %s: expected Success=false and empty result, got %+v", task.ID, task)
			}
		}
	}
}

func TestDiscussionModelErrorYieldsNoMessage(t *testing.T) {
	das := NewDiscussionAutonomySystem(failingProvider())
	das.initiateDiscussion(DiscussionTrigger{Topic: "trees", Urgency: 0.9})
	if n := len(das.activeDiscussions); n != 0 {
		t.Fatalf("expected no discussion without a model opening, got %d", n)
	}

	disc := &Discussion{ID: "d1", Topic: "trees", Active: true, InterestLevel: 1.0}
	das.activeDiscussions[disc.ID] = disc
	das.energyLevel, das.socialCapacity = 1, 1
	reply, err := das.RespondToMessage("d1", DiscussionMessage{From: "user", Content: "hello"})
	if err == nil || reply != "" {
		t.Fatalf("expected an error and no reply, got %q, %v", reply, err)
	}
	for _, m := range disc.Messages {
		if m.From == "echoself" {
			t.Fatalf("unexpected echoself message %q", m.Content)
		}
	}

	das.mu.Lock()
	das.endDiscussion(disc, "test")
	das.mu.Unlock()
	for _, m := range disc.Messages {
		if m.From == "echoself" {
			t.Fatalf("unexpected closing message %q", m.Content)
		}
	}
}

func TestEmergentConceptRequiresModel(t *testing.T) {
	c := emergenceCandidate{id: "em_a_b", key: "a_b", strength: 0.9, from: "x", to: "y"}

	edi := NewEchoDreamKnowledgeIntegration(failingProvider())
	if err := edi.realizeEmergentConcept(context.Background(), c); err == nil {
		t.Fatal("expected an error on model failure")
	}
	if n := len(edi.GetEmergentConcepts()); n != 0 {
		t.Fatalf("expected no emergent concept, got %d", n)
	}

	edi = NewEchoDreamKnowledgeIntegration(&modelOnlyTestProvider{reply: "concept-from-model-5e"})
	if err := edi.realizeEmergentConcept(context.Background(), c); err != nil {
		t.Fatal(err)
	}
	concepts := edi.GetEmergentConcepts()
	if len(concepts) != 1 || concepts[0].Concept != "concept-from-model-5e" {
		t.Fatalf("expected model concept, got %+v", concepts)
	}
}

func TestApplyWisdomOutcomeComesFromModel(t *testing.T) {
	principle := WisdomPrinciple{ID: "p1", Content: "integrate", Depth: 1, Confidence: 1}

	ws := NewWisdomSynthesis(failingProvider())
	ws.wisdomPrinciples = []WisdomPrinciple{principle}
	ws.applyWisdom()
	if ws.totalApplications != 0 || len(ws.wisdomPrinciples[0].Applications) != 0 {
		t.Fatal("expected no application recorded on model error")
	}

	ws = NewWisdomSynthesis(&modelOnlyTestProvider{reply: "outcome-from-model-2d"})
	ws.wisdomPrinciples = []WisdomPrinciple{principle}
	ws.applyWisdom()
	apps := ws.wisdomPrinciples[0].Applications
	if len(apps) != 1 || apps[0].Outcome != "outcome-from-model-2d" {
		t.Fatalf("expected model outcome, got %+v", apps)
	}
}
