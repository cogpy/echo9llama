package autonomous

import (
	"context"
	"errors"
	"testing"

	"github.com/cogpy/echo9llama/core/llm"
)

// fakeLLM is an llm.LLMProvider test double.
type fakeLLM struct {
	reply string
	err   error
	calls int
}

func (f *fakeLLM) Generate(context.Context, string, llm.GenerateOptions) (string, error) {
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	return f.reply, nil
}

func (f *fakeLLM) StreamGenerate(context.Context, string, llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeLLM) Name() string    { return "fake" }
func (f *fakeLLM) Available() bool { return true }
func (f *fakeLLM) MaxTokens() int  { return 1000 }

func TestDiscussionInitiatorNoOpeningOnModelFailure(t *testing.T) {
	di := NewDiscussionInitiator(&fakeLLM{err: errors.New("offline")}, nil)
	defer di.cancel()
	topic := DiscussionTopic{Title: "recursion", Domain: "ideas", Question: "What is recursion?"}
	di.pendingTopics = append(di.pendingTopics, topic)

	di.initiateDiscussion(topic)
	if di.activeDiscussion != nil || di.totalDiscussions != 0 {
		t.Fatalf("model failure must not start a discussion: %#v", di.activeDiscussion)
	}
	if len(di.pendingTopics) != 1 {
		t.Fatal("topic should stay pending after a failed opening")
	}
}

func TestDiscussionInitiatorUsesModelText(t *testing.T) {
	gen := &fakeLLM{reply: "model opening"}
	di := NewDiscussionInitiator(gen, nil)
	defer di.cancel()
	topic := DiscussionTopic{Title: "recursion", Domain: "ideas", Question: "What is recursion?"}

	di.initiateDiscussion(topic)
	if di.activeDiscussion == nil || di.activeDiscussion.Messages[0].Content != "model opening" {
		t.Fatalf("expected model opening, got %#v", di.activeDiscussion)
	}

	gen.err = errors.New("offline")
	if err := di.CompleteDiscussion(); err == nil {
		t.Fatal("expected error when insights cannot be generated")
	}
	if di.activeDiscussion == nil || len(di.activeDiscussion.Insights) != 0 || di.insightsGenerated != 0 {
		t.Fatalf("failed completion must record no insight: %#v", di.activeDiscussion)
	}
}

func TestSelfAssessmentRequiresModel(t *testing.T) {
	mcm := NewMetaCognitiveMonitor(&fakeLLM{err: errors.New("offline")})
	defer mcm.cancel()
	if a := mcm.PerformSelfAssessment(); a != nil {
		t.Fatalf("model failure must produce no assessment, got %#v", a)
	}
	if len(mcm.assessmentHistory) != 0 {
		t.Fatal("model failure must record no assessment")
	}

	mcm = NewMetaCognitiveMonitor(&fakeLLM{reply: "I am steady.\nSTRENGTHS: focus; curiosity\nCHALLENGES: fatigue\nRECOMMENDATIONS: rest"})
	defer mcm.cancel()
	a := mcm.PerformSelfAssessment()
	if a == nil || a.Reflection != "I am steady." {
		t.Fatalf("expected model reflection, got %#v", a)
	}
	if len(a.Strengths) != 2 || a.Challenges[0] != "fatigue" || a.Recommendations[0] != "rest" {
		t.Fatalf("sections not taken from the model: %#v", a)
	}
}
