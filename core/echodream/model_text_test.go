package echodream

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/cogpy/echo9llama/core/llm"
)

// fakeDreamGenerator is a test double for the injected model.
type fakeDreamGenerator struct {
	mu      sync.Mutex
	reply   string
	err     error
	prompts []string
}

func (f *fakeDreamGenerator) Generate(_ context.Context, prompt string, _ llm.GenerateOptions) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.prompts = append(f.prompts, prompt)
	if f.err != nil {
		return "", f.err
	}
	return f.reply, nil
}

func (f *fakeDreamGenerator) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.prompts)
}

func ingestRecurring(dp *DreamProcessor) {
	dp.IngestExperience("Reflected on recursion", 0.9, []string{"recursion", "wisdom"})
	dp.IngestExperience("Explored recursive trees", 0.85, []string{"recursion", "wisdom"})
	dp.IngestExperience("Considered wisdom in rest", 0.8, []string{"wisdom", "rest"})
}

func TestDreamProcessorProducesNoTextWhenModelFails(t *testing.T) {
	dp := NewDreamProcessor(context.Background())
	gen := &fakeDreamGenerator{err: errors.New("provider down")}
	dp.SetTextGenerator(gen)
	ingestRecurring(dp)

	dp.ConsolidateMemories()
	dp.ExtractPatterns()
	dp.SynthesizeWisdom()

	if gen.calls() == 0 {
		t.Fatal("expected the model to be asked for text")
	}
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	if len(dp.consolidatedKnowledge) != 0 || dp.knowledgeConsolidated != 0 {
		t.Fatalf("model failure must produce no knowledge text, got %#v", dp.consolidatedKnowledge)
	}
	if len(dp.wisdomInsights) != 0 || dp.wisdomSynthesized != 0 {
		t.Fatalf("model failure must produce no wisdom text, got %#v", dp.wisdomInsights)
	}
	if dp.totalDreamCycles != 1 || len(dp.pendingExperiences) != 0 {
		t.Fatalf("dream cycle bookkeeping should still complete: cycles=%d pending=%d", dp.totalDreamCycles, len(dp.pendingExperiences))
	}
}

func TestDreamProcessorProducesNoTextWithoutModel(t *testing.T) {
	dp := NewDreamProcessor(context.Background())
	ingestRecurring(dp)

	dp.ConsolidateMemories()
	dp.ExtractPatterns()
	dp.SynthesizeWisdom()

	dp.mu.RLock()
	defer dp.mu.RUnlock()
	if len(dp.consolidatedKnowledge) != 0 || len(dp.wisdomInsights) != 0 {
		t.Fatalf("no model must mean no character text: knowledge=%d wisdom=%d", len(dp.consolidatedKnowledge), len(dp.wisdomInsights))
	}
}

func TestDreamProcessorRejectsEmptyModelReply(t *testing.T) {
	dp := NewDreamProcessor(context.Background())
	dp.SetTextGenerator(&fakeDreamGenerator{reply: "   "})
	ingestRecurring(dp)

	dp.ConsolidateMemories()
	dp.ExtractPatterns()
	dp.SynthesizeWisdom()

	dp.mu.RLock()
	defer dp.mu.RUnlock()
	if len(dp.consolidatedKnowledge) != 0 || len(dp.wisdomInsights) != 0 {
		t.Fatalf("empty model replies must not be recorded: knowledge=%d wisdom=%d", len(dp.consolidatedKnowledge), len(dp.wisdomInsights))
	}
}

func TestEchoDreamUsesModelTextOnly(t *testing.T) {
	ed := NewEchoDream()
	defer ed.cancel()
	ed.AddEpisodicMemory("Noticed a feedback loop", 0.8)

	// Without a model nothing is written.
	ed.consolidateMemories()
	if len(ed.consolidatedKnowledge) != 0 {
		t.Fatalf("expected no knowledge without a model, got %#v", ed.consolidatedKnowledge)
	}

	failing := &fakeDreamGenerator{err: errors.New("offline")}
	ed.SetTextGenerator(failing)
	ed.consolidateMemories()
	if len(ed.consolidatedKnowledge) != 0 || failing.calls() != 1 {
		t.Fatalf("expected one failed call and no knowledge, calls=%d knowledge=%d", failing.calls(), len(ed.consolidatedKnowledge))
	}

	ed.SetTextGenerator(&fakeDreamGenerator{reply: "model knowledge"})
	ed.consolidateMemories()
	if len(ed.consolidatedKnowledge) != 1 || ed.consolidatedKnowledge[0].Content != "model knowledge" {
		t.Fatalf("expected model-written knowledge, got %#v", ed.consolidatedKnowledge)
	}

	ed.SetTextGenerator(&fakeDreamGenerator{err: errors.New("offline")})
	ed.extractWisdom()
	if len(ed.wisdomInsights) != 0 || ed.wisdomExtracted != 0 {
		t.Fatalf("expected no wisdom on model failure, got %#v", ed.wisdomInsights)
	}

	ed.SetTextGenerator(&fakeDreamGenerator{reply: "model wisdom"})
	ed.extractWisdom()
	if len(ed.wisdomInsights) != 1 || ed.wisdomInsights[0].Insight != "model wisdom" {
		t.Fatalf("expected model-written wisdom, got %#v", ed.wisdomInsights)
	}
}

func TestDreamCycleIntegrationModelFailureProducesNoText(t *testing.T) {
	dci := NewDreamCycleIntegration()
	dci.SetTextGenerator(&fakeDreamGenerator{err: errors.New("offline")})
	consolidations := make([]DreamConsolidationResult, 6)
	for i := range consolidations {
		consolidations[i] = DreamConsolidationResult{ConsolidatedTo: "c", Type: "pattern"}
	}

	wisdom := dci.extractWisdom(consolidations, []string{"pattern"})
	if len(wisdom) != 0 {
		t.Fatalf("expected no wisdom on model failure, got %#v", wisdom)
	}
	if narrative := dci.generateDreamNarrative(consolidations, []string{"pattern"}, nil); narrative != "" {
		t.Fatalf("expected no narrative on model failure, got %q", narrative)
	}
	if patterns := dci.extractPatterns(nil); len(patterns) != 0 {
		t.Fatalf("expected no invented themes, got %v", patterns)
	}
}

func TestDreamCycleIntegrationUsesModelText(t *testing.T) {
	dci := NewDreamCycleIntegration()
	gen := &fakeDreamGenerator{reply: "model text"}
	dci.SetTextGenerator(gen)
	var callbackContent []string
	dci.SetOnWisdomExtracted(func(w Wisdom) { callbackContent = append(callbackContent, w.Content) })
	consolidations := make([]DreamConsolidationResult, 6)

	wisdom := dci.extractWisdom(consolidations, []string{"pattern"})
	if len(wisdom) != 2 {
		t.Fatalf("expected insight and principle, got %d", len(wisdom))
	}
	for _, w := range wisdom {
		if w.Content != "model text" {
			t.Fatalf("wisdom was not model text: %q", w.Content)
		}
	}
	if len(callbackContent) != 2 {
		t.Fatalf("expected wisdom callbacks, got %v", callbackContent)
	}
	if narrative := dci.generateDreamNarrative(consolidations, []string{"pattern"}, wisdom); narrative != "model text" {
		t.Fatalf("narrative was not model text: %q", narrative)
	}
}
