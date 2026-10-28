package echoself

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/cogpy/echo9llama/core/llm"
)

type fakeEchoselfGenerator struct {
	reply string
	err   error
	calls int
}

func (f *fakeEchoselfGenerator) Generate(context.Context, string, llm.GenerateOptions) (string, error) {
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	return f.reply, nil
}

func TestEchoselfProducesNoTextWithoutWorkingModel(t *testing.T) {
	ae := NewAutonomousEchoself()
	defer ae.cancel()
	ae.interestPatterns["hello"] = 1.0 // make incoming messages interesting enough to answer

	ae.generateAutonomousThought()
	ae.handleExternalMessage(ExternalMessage{ID: "m1", Source: "user", Content: "hello"})
	if len(ae.thoughtStream) != 0 {
		t.Fatalf("no model must mean no thoughts, got %d", len(ae.thoughtStream))
	}

	failing := &fakeEchoselfGenerator{err: errors.New("offline")}
	ae.SetTextGenerator(failing)
	ae.generateAutonomousThought()
	ae.handleExternalMessage(ExternalMessage{ID: "m2", Source: "user", Content: "hello"})
	if failing.calls != 2 || len(ae.thoughtStream) != 0 || ae.interactionsHandled != 0 {
		t.Fatalf("model failure must produce no text: calls=%d stream=%d handled=%d", failing.calls, len(ae.thoughtStream), ae.interactionsHandled)
	}

	thoughts := make([]Thought, 60)
	for i := range thoughts {
		thoughts[i] = Thought{Content: fmt.Sprintf("thought %d", i)}
	}
	if w := ae.extractWisdomFromThoughts(thoughts); w != nil {
		t.Fatalf("model failure must produce no wisdom, got %#v", w)
	}
}

func TestEchoselfUsesModelText(t *testing.T) {
	ae := NewAutonomousEchoself()
	defer ae.cancel()
	ae.interestPatterns["hello"] = 1.0 // make incoming messages interesting enough to answer
	ae.SetTextGenerator(&fakeEchoselfGenerator{reply: "model words"})

	ae.generateAutonomousThought()
	ae.handleExternalMessage(ExternalMessage{ID: "m1", Source: "user", Content: "hello"})
	if len(ae.thoughtStream) != 2 {
		t.Fatalf("expected two thoughts, got %d", len(ae.thoughtStream))
	}
	for range 2 {
		if got := (<-ae.thoughtStream).Content; got != "model words" {
			t.Fatalf("thought was not model text: %q", got)
		}
	}

	thoughts := make([]Thought, 60)
	if w := ae.extractWisdomFromThoughts(thoughts); w == nil || w.Content != "model words" {
		t.Fatalf("expected model wisdom, got %#v", w)
	}
}
