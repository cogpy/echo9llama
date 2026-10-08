package server

import (
	"testing"

	"github.com/cogpy/echo9llama/core/embodiment"
)

func TestBuildChatUsesEmbodiedPromptWhenNoSystemMessage(t *testing.T) {
	hub := embodiment.NewHub(0.2, 8)
	msgs := []chatMessage{{Role: "user", Content: "how do you feel?"}}

	prompt, system := buildChat(msgs, hub)
	if prompt != "User: how do you feel?" || system != "" {
		t.Fatalf("before any frame: prompt=%q system=%q", prompt, system)
	}

	ref, err := hub.Ingest(embodiment.Frame{Contract: embodiment.Contract, Persona: "Lucy", Valence: -0.4, Arousal: 0.8, Flow: 0.1})
	if err != nil {
		t.Fatal(err)
	}
	if _, system = buildChat(msgs, hub); system != ref.SystemPrompt {
		t.Fatalf("system = %q, want embodied prompt", system)
	}

	withSystem := append([]chatMessage{{Role: "system", Content: "be brief"}}, msgs...)
	prompt, system = buildChat(withSystem, hub)
	if system != "be brief" || prompt != "User: how do you feel?" {
		t.Fatalf("caller system not honoured: prompt=%q system=%q", prompt, system)
	}

	if _, system = buildChat(msgs, nil); system != "" {
		t.Fatalf("nil hub: system = %q", system)
	}
}

func TestBuildChatLabelsTurnsForTheEdgePrompt(t *testing.T) {
	prompt, _ := buildChat([]chatMessage{
		{Role: "user", Content: "hi"},
		{Role: "assistant", Content: "hello"},
		{Role: "user", Content: "how is the match?"},
	}, nil)
	want := "User: hi\nEcho: hello\nUser: how is the match?"
	if prompt != want {
		t.Fatalf("prompt = %q, want %q", prompt, want)
	}
}
