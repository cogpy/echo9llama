package main

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
)

type fakeGenerator struct {
	mu     sync.Mutex
	reply  string
	err    error
	system []string
	prompt []string
}

func (f *fakeGenerator) Generate(_ context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.system = append(f.system, opts.SystemPrompt)
	f.prompt = append(f.prompt, prompt)
	if f.err != nil {
		return "", f.err
	}
	return f.reply, nil
}

// captureEvents starts the daechon's bus and records every event.
func captureEvents(t *testing.T, d *Daechon) func() []deeptreeecho.CogEvent {
	t.Helper()
	var mu sync.Mutex
	var events []deeptreeecho.CogEvent
	d.eventBus.SubscribeAll(func(e deeptreeecho.CogEvent) {
		mu.Lock()
		events = append(events, e)
		mu.Unlock()
	})
	if err := d.eventBus.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(d.eventBus.Stop)
	return func() []deeptreeecho.CogEvent {
		time.Sleep(100 * time.Millisecond)
		mu.Lock()
		defer mu.Unlock()
		return append([]deeptreeecho.CogEvent(nil), events...)
	}
}

func TestHandleChatReplyComesFromModelWithDisposition(t *testing.T) {
	fake := &fakeGenerator{reply: "  MODEL-REPLY  "}
	d := NewDaechon("echo", 4, 30, fake, nil)

	got, err := d.HandleChat(context.Background(), "u", "you are a stupid useless idiot")
	if err != nil {
		t.Fatal(err)
	}
	if got != "MODEL-REPLY" {
		t.Fatalf("reply = %q, want the model's text", got)
	}
	sys := fake.system[0]
	if !strings.Contains(sys, "hostile") && !strings.Contains(sys, "irritated") {
		t.Fatalf("mood not passed to model as prompt input: %q", sys)
	}
	if !strings.Contains(fake.prompt[0], "you are a stupid useless idiot") {
		t.Fatalf("user message missing from prompt: %q", fake.prompt[0])
	}

	// The next turn carries the history.
	if _, err := d.HandleChat(context.Background(), "u", "second"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(fake.prompt[1], "Echo: MODEL-REPLY") {
		t.Fatalf("history missing from prompt: %q", fake.prompt[1])
	}
}

func TestHandleChatModelFailureSaysNothing(t *testing.T) {
	for name, gen := range map[string]Generator{
		"nil":   nil,
		"error": &fakeGenerator{err: llm.ErrNoEdgeModel},
		"empty": &fakeGenerator{reply: "   "},
	} {
		t.Run(name, func(t *testing.T) {
			d := NewDaechon("echo", 4, 30, gen, nil)
			events := captureEvents(t, d)

			got, err := d.HandleChat(context.Background(), "u", "hello")
			if err == nil || got != "" {
				t.Fatalf("got %q, %v; want no reply and an error", got, err)
			}
			for _, e := range events() {
				if strings.HasPrefix(e.Source, "echo:") {
					t.Fatalf("character event published on model failure: %+v", e)
				}
			}
		})
	}
}

func TestSpeakAsyncFailurePublishesOnlySystemLine(t *testing.T) {
	d := NewDaechon("echo", 4, 30, &fakeGenerator{err: errors.New("down")}, nil)
	events := captureEvents(t, d)
	d.speakAsync(context.Background(), deeptreeecho.CogEventWakeRest, "awakening line", "wake up")

	var sawSystem bool
	for _, e := range events() {
		if e.Category == deeptreeecho.CogEventWakeRest || strings.HasPrefix(e.Source, "echo:") {
			t.Fatalf("character wake event on model failure: %+v", e)
		}
		if e.Category == deeptreeecho.CogEventSystem && strings.HasPrefix(e.Content, "[system]") {
			sawSystem = true
		}
	}
	if !sawSystem {
		t.Fatal("no [system] notice for the failed awakening line")
	}
}

func TestSpeakAsyncUsesModelText(t *testing.T) {
	d := NewDaechon("echo", 4, 30, &fakeGenerator{reply: "MODEL-WAKE"}, nil)
	events := captureEvents(t, d)
	d.speakAsync(context.Background(), deeptreeecho.CogEventWakeRest, "awakening line", "wake up")
	for _, e := range events() {
		if e.Category == deeptreeecho.CogEventWakeRest && e.Content == "MODEL-WAKE" {
			return
		}
	}
	t.Fatal("model awakening line not published")
}
