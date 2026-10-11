package deeptreeecho

import (
	"context"
	"errors"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

type wisdomCallerContextKey struct{}

type wisdomEvidenceTestProvider struct {
	generate func(context.Context, string, llm.GenerateOptions) (string, error)
}

func (p *wisdomEvidenceTestProvider) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	if p.generate == nil {
		return "", errors.New("test provider was not configured")
	}
	return p.generate(ctx, prompt, opts)
}

func (p *wisdomEvidenceTestProvider) StreamGenerate(context.Context, string, llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk)
	close(ch)
	return ch, nil
}

func (p *wisdomEvidenceTestProvider) Name() string    { return "wisdom-evidence-test" }
func (p *wisdomEvidenceTestProvider) Available() bool { return true }
func (p *wisdomEvidenceTestProvider) MaxTokens() int  { return 1024 }

func TestWisdomCandidatesDoNotCreateFabricatedApplications(t *testing.T) {
	provider := &wisdomEvidenceTestProvider{generate: func(context.Context, string, llm.GenerateOptions) (string, error) {
		return "WISDOM: Test observations before generalizing\nDOMAIN: Learning\nDEPTH: 0.9", nil
	}}
	ws := NewWisdomSynthesis(provider)
	candidate := ws.synthesizeWisdom([]AccumulatedPattern{
		{ID: "p1", Description: "first observation"},
		{ID: "p2", Description: "second observation"},
		{ID: "p3", Description: "third observation"},
	})
	if candidate == nil {
		t.Fatal("expected a dynamically generated candidate")
	}
	if candidate.Confidence != 0 {
		t.Fatalf("generated candidate received unearned confidence: %.2f", candidate.Confidence)
	}
	if candidate.Depth != 0 || candidate.EpistemicStatus != "hypothesis" {
		t.Fatalf("model-awarded candidate profundity accepted: %#v", candidate)
	}

	ws.mu.Lock()
	ws.wisdomPrinciples = append(ws.wisdomPrinciples, *candidate)
	ws.mu.Unlock()

	applied := make(chan WisdomPrinciple, 1)
	ws.SetCallbacks(nil, func(principle WisdomPrinciple, _ string) { applied <- principle }, nil)
	ws.applyWisdom()

	principles := ws.GetWisdomPrinciples()
	if len(principles) != 1 {
		t.Fatalf("expected one candidate, got %d", len(principles))
	}
	if principles[0].ApplyCount != 0 || len(principles[0].Applications) != 0 || !principles[0].LastApplied.IsZero() {
		t.Fatalf("candidate selection fabricated an application: %#v", principles[0])
	}
	metrics := ws.GetMetrics()
	if applications := metrics["total_applications"].(uint64); applications != 0 {
		t.Fatalf("candidate selection increased applications: %d", applications)
	}
	if level := ws.GetWisdomLevel(); level != 0 {
		t.Fatalf("candidate selection increased outcome-calibrated wisdom level: %.2f", level)
	}
	select {
	case <-applied:
		t.Fatal("candidate selection invoked an application callback")
	case <-time.After(30 * time.Millisecond):
	}
}

func TestWisdomRecordOutcomeCalibratesOnlyFromObservedInput(t *testing.T) {
	ws := NewWisdomSynthesis(&wisdomEvidenceTestProvider{})
	ws.mu.Lock()
	ws.wisdomPrinciples = []WisdomPrinciple{{
		ID:         "candidate-1",
		Content:    "Try the measured practice",
		Depth:      0.9,
		Confidence: 0,
	}}
	ws.mu.Unlock()

	if err := ws.RecordWisdomOutcome("candidate-1", "practice run", "artifact passed verification", 0.8); err != nil {
		t.Fatalf("record outcome: %v", err)
	}
	if err := ws.RecordWisdomOutcome("candidate-1", "practice run", "bad score", math.NaN()); err == nil {
		t.Fatal("accepted non-finite outcome effectiveness")
	}

	principle := ws.GetWisdomPrinciples()[0]
	if principle.ApplyCount != 1 || len(principle.Applications) != 1 {
		t.Fatalf("expected exactly one recorded outcome, got %#v", principle)
	}
	if got := principle.Applications[0]; got.Context != "practice run" || got.Outcome != "artifact passed verification" || got.Effectiveness != 0.8 {
		t.Fatalf("outcome was not preserved exactly: %#v", got)
	}
	if principle.Confidence != 0.8 {
		t.Fatalf("confidence did not derive from the recorded outcome: %.2f", principle.Confidence)
	}
	if level := ws.GetWisdomLevel(); level != 0.8 {
		t.Fatalf("wisdom level did not derive from recorded outcome evidence: %.2f", level)
	}
	if applications := ws.GetMetrics()["total_applications"].(uint64); applications != 1 {
		t.Fatalf("expected one outcome-backed application, got %d", applications)
	}
}

func TestWisdomEvolutionDoesNotHoldLockOrInflateEvidenceMetrics(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	provider := &wisdomEvidenceTestProvider{generate: func(ctx context.Context, _ string, _ llm.GenerateOptions) (string, error) {
		once.Do(func() { close(entered) })
		<-release
		if ctx == context.Background() {
			return "", errors.New("evolution used a background context")
		}
		return "Refined candidate based on the supplied outcomes.", nil
	}}
	ws := NewWisdomSynthesis(provider)
	observations := []WisdomApplication{
		{Context: "run 1", Outcome: "measured", Effectiveness: 0.6},
		{Context: "run 2", Outcome: "measured", Effectiveness: 0.6},
		{Context: "run 3", Outcome: "measured", Effectiveness: 0.6},
	}
	ws.mu.Lock()
	ws.wisdomPrinciples = []WisdomPrinciple{
		{ID: "evolvable", Content: "Original candidate", Depth: 0.4, Confidence: 0.6, Applications: observations, ApplyCount: len(observations)},
		{ID: "other", Content: "Other candidate", Depth: 0.7, Confidence: 0},
	}
	ws.mu.Unlock()

	evolved := make(chan struct{})
	go func() {
		ws.evolveWisdom()
		close(evolved)
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("evolution did not invoke the provider")
	}

	metricsRead := make(chan struct{})
	go func() {
		_ = ws.GetMetrics()
		close(metricsRead)
	}()
	select {
	case <-metricsRead:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("provider generation held the wisdom lock")
	}

	close(release)
	select {
	case <-evolved:
	case <-time.After(time.Second):
		t.Fatal("evolution did not complete")
	}

	principle := ws.GetWisdomPrinciples()[0]
	if principle.Content != "Refined candidate based on the supplied outcomes." {
		t.Fatalf("expected refined candidate content, got %q", principle.Content)
	}
	if principle.Depth != 0.4 || principle.Confidence != 0.6 {
		t.Fatalf("generated refinement inflated evidence metrics: depth=%.2f confidence=%.2f", principle.Depth, principle.Confidence)
	}
	if principle.EvolutionCount != 1 {
		t.Fatalf("expected one outcome-gated refinement, got %d", principle.EvolutionCount)
	}
}

func TestWisdomSynthesisRejectsMissingOrEmptyProviderOutput(t *testing.T) {
	thoughts := []string{"one", "two", "three"}
	if _, err := NewWisdomSynthesis(nil).SynthesizeWisdom(context.Background(), thoughts); err == nil {
		t.Fatal("nil provider did not return an error")
	}

	emptyProvider := &wisdomEvidenceTestProvider{generate: func(context.Context, string, llm.GenerateOptions) (string, error) {
		return " \t\n", nil
	}}
	if _, err := NewWisdomSynthesis(emptyProvider).SynthesizeWisdom(context.Background(), thoughts); err == nil {
		t.Fatal("empty provider output did not return an error")
	}

	provider := &wisdomEvidenceTestProvider{generate: func(ctx context.Context, _ string, _ llm.GenerateOptions) (string, error) {
		if ctx.Value(wisdomCallerContextKey{}) != "preserved" {
			return "", errors.New("caller context was not passed to provider")
		}
		return "A provisional insight", nil
	}}
	ctx := context.WithValue(context.Background(), wisdomCallerContextKey{}, "preserved")
	insight, err := NewWisdomSynthesis(provider).SynthesizeWisdom(ctx, thoughts)
	if err != nil {
		t.Fatalf("synthesize wisdom: %v", err)
	}
	if insight.EpistemicStatus != "hypothesis" || insight.Depth != 0 {
		t.Fatalf("generated insight lost its hypothesis status: %#v", insight)
	}
}

func TestWisdomPrincipleSnapshotsDoNotAliasInternalSlices(t *testing.T) {
	ws := NewWisdomSynthesis(nil)
	ws.mu.Lock()
	ws.wisdomPrinciples = []WisdomPrinciple{{
		ID:             "snapshot",
		SourcePatterns: []string{"original-pattern"},
		Applications:   []WisdomApplication{{Context: "original-context", Outcome: "original-outcome", Effectiveness: 0.5}},
	}}
	ws.mu.Unlock()

	snapshot := ws.GetWisdomPrinciples()
	snapshot[0].SourcePatterns[0] = "mutated-pattern"
	snapshot[0].Applications[0].Context = "mutated-context"
	snapshot[0].Applications = append(snapshot[0].Applications, WisdomApplication{Context: "new"})

	fresh := ws.GetWisdomPrinciples()[0]
	if fresh.SourcePatterns[0] != "original-pattern" || fresh.Applications[0].Context != "original-context" || len(fresh.Applications) != 1 {
		t.Fatalf("wisdom snapshot leaked internal storage: %#v", fresh)
	}
}
