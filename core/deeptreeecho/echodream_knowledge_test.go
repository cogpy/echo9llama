package deeptreeecho

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
	"github.com/cogpy/echo9llama/core/pienn"
)

// echoTextFake is a scripted model for tests: it returns reply (or err) and
// records every prompt and option set it was given.
type echoTextFake struct {
	mu      sync.Mutex
	reply   func(prompt string) string
	err     error
	prompts []string
	opts    []llm.GenerateOptions
}

func (f *echoTextFake) Generate(_ context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.prompts = append(f.prompts, prompt)
	f.opts = append(f.opts, opts)
	if f.err != nil {
		return "", f.err
	}
	return f.reply(prompt), nil
}

func (f *echoTextFake) StreamGenerate(ctx context.Context, prompt string, opts llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	out, err := f.Generate(ctx, prompt, opts)
	ch <- llm.StreamChunk{Content: out, Error: err, Done: true}
	close(ch)
	return ch, nil
}
func (f *echoTextFake) Name() string    { return "echoTextFake" }
func (f *echoTextFake) Available() bool { return true }
func (f *echoTextFake) MaxTokens() int  { return 4096 }

func seedDreamMemories(eki *EchodreamKnowledgeIntegrator) {
	for range 4 {
		eki.RecordMemory("we talked about rivers and memory", "reflective", "user:a", 0.8)
		eki.RecordMemory("a question about time crystals", "curious", "user:b", 0.7)
		eki.RecordMemory("an insult about my architecture", "hostile", "user:c", 0.6)
	}
}

func TestDreamCycleInsightAndNarrativeComeFromModel(t *testing.T) {
	eki := NewEchodreamKnowledgeIntegrator(nil)
	fake := &echoTextFake{reply: func(p string) string {
		if strings.Contains(p, "emergent insight") {
			return "MODEL-INSIGHT"
		}
		return "MODEL-NARRATIVE"
	}}
	eki.SetGenerator(fake)
	seedDreamMemories(eki)

	entry := eki.RunDreamCycle(context.Background())
	if entry.InsightsGenerated != 1 {
		t.Fatalf("insights = %d, want 1", entry.InsightsGenerated)
	}
	if got := eki.EmergentInsights[0].Insight; got != "MODEL-INSIGHT" {
		t.Fatalf("insight = %q, want model text", got)
	}
	if entry.DreamNarrative != "MODEL-NARRATIVE" {
		t.Fatalf("narrative = %q, want model text", entry.DreamNarrative)
	}
	if !strings.Contains(fake.prompts[0], "rivers and memory") {
		t.Fatalf("insight prompt lacks the actual memories: %q", fake.prompts[0])
	}
}

func TestDreamCycleModelFailureProducesNoText(t *testing.T) {
	for name, gen := range map[string]EchoTextGenerator{
		"nil":   nil,
		"error": &echoTextFake{err: errors.New("model down")},
	} {
		t.Run(name, func(t *testing.T) {
			eki := NewEchodreamKnowledgeIntegrator(nil)
			eki.SetGenerator(gen)
			seedDreamMemories(eki)

			entry := eki.RunDreamCycle(context.Background())
			if entry.InsightsGenerated != 0 || len(eki.EmergentInsights) != 0 {
				t.Fatalf("insights produced without a model: %+v", eki.EmergentInsights)
			}
			if entry.DreamNarrative != "" {
				t.Fatalf("narrative produced without a model: %q", entry.DreamNarrative)
			}
		})
	}
}

func TestGoalSchedulerGoalsComeFromModel(t *testing.T) {
	s := NewEchobeatsGoalScheduler(nil)
	if len(s.goals) != 0 {
		t.Fatalf("scheduler seeded %d fixed goals", len(s.goals))
	}
	fake := &echoTextFake{reply: func(string) string { return "Trace where my curiosity came from\nextra line" }}
	s.SetGenerator(fake)
	s.running = true

	s.reviewGoals(context.Background())
	if len(s.goals) != 3 {
		t.Fatalf("goals = %d, want 3", len(s.goals))
	}
	for _, g := range s.goals {
		if g.Description != "Trace where my curiosity came from" {
			t.Fatalf("goal description %q not from model", g.Description)
		}
	}
}

func TestGoalSchedulerModelFailureAddsNoGoal(t *testing.T) {
	s := NewEchobeatsGoalScheduler(nil)
	s.SetGenerator(&echoTextFake{err: errors.New("model down")})
	s.running = true
	s.reviewGoals(context.Background())
	if len(s.goals) != 0 {
		t.Fatalf("goals added on model failure: %d", len(s.goals))
	}

	s.SetGenerator(nil)
	s.reviewGoals(context.Background())
	if len(s.goals) != 0 {
		t.Fatalf("goals added without a model: %d", len(s.goals))
	}
}

func TestGoalSchedulerStepThoughtFromModel(t *testing.T) {
	s := NewEchobeatsGoalScheduler(nil)
	fake := &echoTextFake{reply: func(string) string { return "MODEL-STEP" }}
	s.SetGenerator(fake)
	engine := s.engines[0]

	s.mu.Lock()
	out := s.processStep(1, engine)
	s.mu.Unlock()
	if out != "" {
		t.Fatalf("processStep returned fixed text %q before the model answered", out)
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		s.mu.RLock()
		got := engine.Output
		s.mu.RUnlock()
		if got == "MODEL-STEP" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("engine output = %q, want model text", got)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestResponseStyleIsPromptGuidance(t *testing.T) {
	de := NewDispositionEngine(nil)
	reaction := de.React("u", "you are a stupid useless idiot")
	g := reaction.ResponseStyle.PromptGuidance()
	if !strings.Contains(g, reaction.ResponseStyle.Instruction) || !strings.Contains(g, "mood") {
		t.Fatalf("guidance %q does not carry the style instruction", g)
	}
}

func TestAutonomousInteractionRepliesOnlyFromModel(t *testing.T) {
	bus := NewCognitiveEventBusV3()
	fake := &echoTextFake{reply: func(string) string { return "MODEL-REPLY" }}
	ais := NewAutonomousInteractionSystem(pienn.NewAdaptiveCognitiveCore(), NewDispositionEngine(nil), bus, fake)
	got, err := ais.GenerateResponse("u", "hello there")
	if err != nil || got != "MODEL-REPLY" {
		t.Fatalf("GenerateResponse = %q, %v; want model reply", got, err)
	}
	if !strings.Contains(fake.opts[0].SystemPrompt, "Current disposition") {
		t.Fatalf("disposition not passed to model: %q", fake.opts[0].SystemPrompt)
	}

	failing := NewAutonomousInteractionSystem(pienn.NewAdaptiveCognitiveCore(), NewDispositionEngine(nil), bus, &echoTextFake{err: errors.New("down")})
	if got, err := failing.GenerateResponse("u", "hello"); err == nil || got != "" {
		t.Fatalf("on model error got %q, %v; want no text and an error", got, err)
	}
	none := NewAutonomousInteractionSystem(pienn.NewAdaptiveCognitiveCore(), NewDispositionEngine(nil), bus, nil)
	if got, err := none.GenerateResponse("u", "hello"); err == nil || got != "" {
		t.Fatalf("without a model got %q, %v; want no text and an error", got, err)
	}
}

func TestPersistentLoopThoughtsOnlyFromModel(t *testing.T) {
	bus := NewCognitiveEventBusV3()
	fake := &echoTextFake{reply: func(string) string { return "MODEL-THOUGHT" }}
	pcl := NewPersistentCognitiveLoop(bus, nil, nil, fake, DefaultPersistentLoopConfig())
	pcl.running = true
	pcl.generateAutonomousThought()
	if len(pcl.thoughtStream) != 1 || pcl.thoughtStream[0].Content != "MODEL-THOUGHT" {
		t.Fatalf("thought stream = %+v, want one model thought", pcl.thoughtStream)
	}

	failing := NewPersistentCognitiveLoop(bus, nil, nil, &echoTextFake{err: errors.New("down")}, DefaultPersistentLoopConfig())
	failing.running = true
	failing.generateAutonomousThought()
	if len(failing.thoughtStream) != 0 {
		t.Fatalf("thought produced on model failure: %+v", failing.thoughtStream)
	}
	none := NewPersistentCognitiveLoop(bus, nil, nil, nil, DefaultPersistentLoopConfig())
	none.running = true
	none.generateAutonomousThought()
	if len(none.thoughtStream) != 0 {
		t.Fatalf("thought produced without a model: %+v", none.thoughtStream)
	}
}
