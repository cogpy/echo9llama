package deeptreeecho

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

type observedWisdomProvider struct {
	generate func(context.Context, string, llm.GenerateOptions) (string, error)
}

func (p observedWisdomProvider) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	return p.generate(ctx, prompt, opts)
}

func (observedWisdomProvider) StreamGenerate(context.Context, string, llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	return nil, errors.New("unused")
}
func (observedWisdomProvider) Name() string    { return "observed-wisdom-test" }
func (observedWisdomProvider) Available() bool { return true }
func (observedWisdomProvider) MaxTokens() int  { return 4096 }

func observedWisdomPatterns() []AccumulatedPattern {
	return []AccumulatedPattern{
		{ID: "p1", Description: "read-back revealed a mismatch", Source: "observation", Strength: 1},
		{ID: "p2", Description: "the repair restored the hash", Source: "observation", Strength: 1},
		{ID: "p3", Description: "an unchecked change regressed", Source: "observation", Strength: 1},
	}
}

func TestObservedWisdomLongNonsenseAndSelfRatingsNeverCreateDepth(t *testing.T) {
	for _, output := range []string{
		strings.Repeat("x", 250),
		strings.Repeat("nonsense ", 1000),
		"WISDOM: provider-only candidate\nDOMAIN: Learning\nDEPTH: 1\nCONFIDENCE: 1\nEFFECTIVENESS: 1",
		"WISDOM: provider-only candidate\nDOMAIN: Learning\nDEPTH: NaN\nCONFIDENCE: +Inf",
	} {
		ws := NewWisdomSynthesis(&modelOnlyTestProvider{reply: output})
		insight, err := ws.SynthesiseWisdomWithContext(t.Context(), []string{"one", "two", "three"})
		if err != nil || insight == nil {
			t.Fatalf("dynamic candidate missing: %v", err)
		}
		if insight.Depth != 0 || ws.GetWisdomLevel() != 0 {
			t.Fatalf("verbosity or self-rating earned wisdom: %+v level=%v", insight, ws.GetWisdomLevel())
		}
		ws.pendingPatterns = observedWisdomPatterns()
		ws.triggerSynthesis()
		principles := ws.GetWisdomPrinciples()
		if len(principles) != 1 || principles[0].Depth != 0 || principles[0].Confidence != 0 || principles[0].Feedback.Observations != 0 || ws.GetWisdomLevel() != 0 {
			t.Fatalf("pattern synthesis fabricated quality: %+v", principles)
		}
		ws.applyWisdom()
		principle := ws.GetWisdomPrinciples()[0]
		if len(principle.Applications) != 1 || principle.Applications[0].Observed || principle.Applications[0].Effectiveness != 0 || principle.Confidence != 0 || ws.GetWisdomLevel() != 0 {
			t.Fatalf("generated application fabricated observed feedback: %+v", principle)
		}
	}
}

func TestObservedWisdomEmptyAndInvalidInputsFailClosed(t *testing.T) {
	for _, output := range []string{"", " \n\t", "WISDOM:\nDOMAIN: Learning\nDEPTH: 1", "DEPTH: NaN\nCONFIDENCE: +Inf"} {
		ws := NewWisdomSynthesis(&modelOnlyTestProvider{reply: output})
		before := ws.GetMetrics()
		if insight, err := ws.SynthesizeWisdom(t.Context(), []string{"one", "two", "three"}); err == nil || insight != nil {
			t.Fatalf("empty principle accepted from %q: insight=%+v err=%v", output, insight, err)
		}
		if !reflect.DeepEqual(before, ws.GetMetrics()) {
			t.Fatal("empty output mutated synthesis metrics")
		}
		ws.pendingPatterns = observedWisdomPatterns()
		ws.triggerSynthesis()
		if len(ws.GetWisdomPrinciples()) != 0 || len(ws.pendingPatterns) != 3 || ws.GetWisdomLevel() != 0 {
			t.Fatal("failed synthesis lost patterns or awarded a principle")
		}
	}
	ws := NewWisdomSynthesis(nil)
	if _, err := ws.SynthesizeWisdom(t.Context(), []string{"one", "two", "three"}); err == nil {
		t.Fatal("nil provider did not fail closed")
	}
	//nolint:staticcheck // Deliberately verify nil contexts fail closed at this API boundary.
	if _, err := ws.SynthesizeWisdom(nil, []string{"one", "two", "three"}); err == nil {
		t.Fatal("nil context did not fail closed")
	}
	if _, err := ws.SynthesizeWisdom(t.Context(), []string{"one", " ", "three"}); err == nil {
		t.Fatal("empty input thought did not fail closed")
	}
	for _, strength := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), -0.1, 1.1} {
		ws.AccumulatePattern("invalid", "observation", strength, nil)
	}
	ws.AccumulatePattern(" ", "observation", 1, nil)
	if ws.GetMetrics()["pending_patterns"] != 0 {
		t.Fatal("invalid pattern seeded synthesis")
	}
}

func TestObservedWisdomFeedbackOnlyAndFailureSensitive(t *testing.T) {
	ws := NewWisdomSynthesis(&modelOnlyTestProvider{reply: "application proposal"})
	ws.wisdomPrinciples = []WisdomPrinciple{{ID: "principle", Content: "check the result", Feedback: newWisdomFeedbackEstimate()}}
	for range 5 {
		ws.applyWisdom()
	}
	ws.evolveWisdom()
	if ws.GetWisdomLevel() != 0 || ws.GetWisdomPrinciples()[0].Feedback.Observations != 0 || ws.GetMetrics()["total_evolutions"] != uint64(0) {
		t.Fatal("unknown proposals earned quality or motivated an automatic revision")
	}
	for i := range 25 {
		if applied, err := ws.ApplyObservedWisdomFeedback("principle", fmt.Sprintf("feedback-%d", i), "artifact verification", "verified correct hash", 1); err != nil || !applied {
			t.Fatalf("explicit observation rejected: apply=%v err=%v", applied, err)
		}
	}
	before := ws.GetWisdomPrinciples()[0]
	level := ws.GetWisdomLevel()
	if applied, err := ws.ApplyObservedWisdomFeedback("principle", "negative", "artifact verification", "observed mismatch", 0); err != nil || !applied {
		t.Fatal(err)
	}
	after := ws.GetWisdomPrinciples()[0]
	if after.Confidence >= before.Confidence || ws.GetWisdomLevel() >= level || after.Depth != 0 {
		t.Fatalf("observed failure did not lower advisory effectiveness: before=%+v after=%+v", before.Feedback, after.Feedback)
	}
	if after.Feedback.Version != WisdomFeedbackVersion || !after.Feedback.Advisory || after.Feedback.Observations != 26 || after.Feedback.EffectivenessSum != 25 {
		t.Fatalf("feedback estimator not transparent: %+v", after.Feedback)
	}
	last := after.Applications[len(after.Applications)-1]
	if !last.Observed || last.EvidenceID != "negative" || last.Outcome != "observed mismatch" || last.Effectiveness != 0 {
		t.Fatalf("negative outcome not represented: %+v", last)
	}
	if applied, err := ws.ApplyObservedWisdomFeedback("principle", "negative", "artifact verification", "observed mismatch", 0); err != nil || applied {
		t.Fatalf("duplicate feedback not inert: applied=%v err=%v", applied, err)
	}
	if applied, err := ws.ApplyObservedWisdomFeedback("principle", "negative", "artifact verification", "different outcome", 1); err == nil || applied {
		t.Fatalf("conflicting feedback credited: applied=%v err=%v", applied, err)
	}
	if !reflect.DeepEqual(after, ws.GetWisdomPrinciples()[0]) {
		t.Fatal("duplicate/conflicting observation mutated state")
	}
	beforeMetrics := ws.GetMetrics()
	for i, score := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), -0.1, 1.1} {
		if applied, err := ws.ApplyObservedWisdomFeedback("principle", fmt.Sprint(i), "context", "outcome", score); err == nil || applied {
			t.Fatalf("nonfinite/out of range feedback credited: %v", score)
		}
	}
	if applied, err := ws.ApplyObservedWisdomFeedback("principle", "empty", "context", " ", 1); err == nil || applied {
		t.Fatal("empty observed outcome was credited")
	}
	if !reflect.DeepEqual(beforeMetrics, ws.GetMetrics()) {
		t.Fatal("invalid feedback mutated metrics")
	}
}

func TestObservedWisdomSnapshotsDeepCopyNestedRecords(t *testing.T) {
	ws := NewWisdomSynthesis(nil)
	ws.wisdomPrinciples = []WisdomPrinciple{{ID: "p", SourcePatterns: []string{"source"}, Applications: []WisdomApplication{{Outcome: "observed"}}}}
	ws.wisdomGraph.Nodes["p"] = &WisdomNode{PrincipleID: "p", Connections: 1}
	ws.wisdomGraph.Edges = []WisdomEdge{{FromID: "p", Strength: 0.5}}
	ws.wisdomGraph.Clusters = []WisdomCluster{{ID: "cluster", PrincipleIDs: []string{"p"}}}
	snapshot := ws.GetWisdomPrinciples()
	snapshot[0].SourcePatterns[0] = "mutated"
	snapshot[0].Applications[0].Outcome = "mutated"
	graph := ws.GetWisdomGraph()
	graph.Nodes["p"].Connections = 99
	graph.Edges[0].Strength = 1
	graph.Clusters[0].PrincipleIDs[0] = "mutated"
	fresh := ws.GetWisdomPrinciples()[0]
	freshGraph := ws.GetWisdomGraph()
	if fresh.SourcePatterns[0] != "source" || fresh.Applications[0].Outcome != "observed" || freshGraph.Nodes["p"].Connections != 1 || freshGraph.Edges[0].Strength != 0.5 || freshGraph.Clusters[0].PrincipleIDs[0] != "p" {
		t.Fatal("getter exposed nested mutable state")
	}
}

func TestObservedWisdomDynamicProviderPersonaAndCompatibility(t *testing.T) {
	var calls atomic.Int64
	ws := NewWisdomSynthesis(observedWisdomProvider{generate: func(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
		if !strings.Contains(prompt, "[System: You are a wisdom synthesis system extracting deep insights from experiences.]") || opts.Temperature != 0.7 || opts.MaxTokens != 150 {
			t.Errorf("persona or generation options changed: prompt=%q opts=%+v", prompt, opts)
		}
		if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > wisdomProviderTimeout {
			t.Error("provider request is not bounded")
		}
		return fmt.Sprintf("dynamic-provider-candidate-%d", calls.Add(1)), nil
	}})
	first, err := ws.SynthesiseWisdomWithContext(t.Context(), []string{"one", "two", "three"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := ws.SynthesizeWisdom(t.Context(), []string{"one", "two", "three"})
	if err != nil {
		t.Fatal(err)
	}
	if first.Insight != "dynamic-provider-candidate-1" || second.Insight != "dynamic-provider-candidate-2" || first.Depth != 0 || second.Depth != 0 {
		t.Fatalf("wrappers did not preserve genuine dynamic provider output: %+v %+v", first, second)
	}
}

func TestObservedWisdomProviderCancellationAndUnlockedCalls(t *testing.T) {
	entered := make(chan struct{})
	var ws *WisdomSynthesis
	ws = NewWisdomSynthesis(observedWisdomProvider{generate: func(ctx context.Context, _ string, _ llm.GenerateOptions) (string, error) {
		_ = ws.GetMetrics() // Would deadlock if the caller holds ws.mu.
		ws.SetCallbacks(nil, nil, nil)
		close(entered)
		<-ctx.Done()
		return "late unverified output", nil // Even a late success must fail closed.
	}})
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := ws.SynthesizeWisdom(ctx, []string{"one", "two", "three"})
		done <- err
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("provider call blocked on component lock")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("caller cancellation not propagated: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled provider did not return")
	}
	if ws.GetMetrics()["total_synthesized"] != uint64(0) || ws.GetWisdomLevel() != 0 {
		t.Fatal("cancelled request committed a candidate or awarded wisdom")
	}
}

func TestObservedWisdomProviderLifecycleCancellation(t *testing.T) {
	entered := make(chan struct{})
	ws := NewWisdomSynthesis(observedWisdomProvider{generate: func(ctx context.Context, _ string, _ llm.GenerateOptions) (string, error) {
		close(entered)
		<-ctx.Done()
		return "", ctx.Err()
	}})
	if err := ws.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := ws.SynthesiseWisdomWithContext(t.Context(), []string{"one", "two", "three"})
		done <- err
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("provider was not invoked")
	}
	if err := ws.Stop(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("shutdown cancellation missing: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown failed to cancel active generation")
	}
}

func TestObservedWisdomEvolutionProviderUnlockedAndUnverified(t *testing.T) {
	var ws *WisdomSynthesis
	ws = NewWisdomSynthesis(observedWisdomProvider{generate: func(_ context.Context, prompt string, _ llm.GenerateOptions) (string, error) {
		_ = ws.GetWisdomPrinciples()
		if strings.Contains(prompt, "unverified proposal") {
			t.Error("unknown proposal supplied as observed experience")
		}
		return "revised provider principle", nil
	}})
	ws.wisdomPrinciples = []WisdomPrinciple{{ID: "principle", Content: "original principle", Feedback: newWisdomFeedbackEstimate()}}
	for i := range 3 {
		if _, err := ws.ApplyObservedWisdomFeedback("principle", fmt.Sprintf("id-%d", i), "context", "observed outcome", 0.6); err != nil {
			t.Fatal(err)
		}
	}
	ws.wisdomPrinciples[0].Applications = append(ws.wisdomPrinciples[0].Applications, WisdomApplication{Outcome: "unverified proposal", Observed: false})
	done := make(chan struct{})
	go func() {
		ws.evolveWisdom()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("evolution provider called while holding ws.mu")
	}
	principle := ws.GetWisdomPrinciples()[0]
	if principle.Content != "revised provider principle" || principle.EvolutionCount != 1 || principle.Depth != 0 || principle.Confidence != 0 || principle.Feedback.Observations != 0 || ws.GetWisdomLevel() != 0 {
		t.Fatalf("generated revision inherited quality from another hypothesis: %+v", principle)
	}
	if applied, err := ws.ApplyObservedWisdomFeedback("principle", "id-0", "context", "observed outcome", 0.6); err != nil || applied {
		t.Fatalf("old evidence replay credited new revision: apply=%v err=%v", applied, err)
	}
}

func TestObservedWisdomConcurrentFeedbackSnapshotsAndProposals(t *testing.T) {
	ws := NewWisdomSynthesis(&modelOnlyTestProvider{reply: "dynamic application proposal"})
	ws.wisdomPrinciples = []WisdomPrinciple{{ID: "principle", Content: "check results", Feedback: newWisdomFeedbackEstimate()}}
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			if _, err := ws.ApplyObservedWisdomFeedback("principle", fmt.Sprint(i), "context", "observed outcome", 0.5); err != nil {
				t.Error(err)
			}
		})
		wg.Go(func() {
			ws.applyWisdom()
		})
		wg.Go(func() {
			_ = ws.GetWisdomPrinciples()
			_ = ws.GetWisdomGraph()
			_ = ws.GetMetrics()
			_ = ws.GetWisdomLevel()
		})
	}
	wg.Wait()
	principle := ws.GetWisdomPrinciples()[0]
	if principle.Feedback.Observations != 20 || len(principle.Applications) != 40 {
		t.Fatalf("concurrent operation lost evidence: observations=%d applications=%d", principle.Feedback.Observations, len(principle.Applications))
	}
}

func TestObservedWisdomShortCallerDeadlineWins(t *testing.T) {
	ws := NewWisdomSynthesis(observedWisdomProvider{generate: func(ctx context.Context, _ string, _ llm.GenerateOptions) (string, error) {
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > time.Second {
			t.Error("provider failed to inherit the shorter caller deadline")
		}
		<-ctx.Done()
		return "", ctx.Err()
	}})
	ctx, cancel := context.WithTimeout(t.Context(), 25*time.Millisecond)
	defer cancel()
	if insight, err := ws.SynthesiseWisdomWithContext(ctx, []string{"one", "two", "three"}); !errors.Is(err, context.DeadlineExceeded) || insight != nil {
		t.Fatalf("deadline did not fail closed: insight=%+v err=%v", insight, err)
	}
	if ws.GetMetrics()["total_synthesized"] != uint64(0) {
		t.Fatal("deadline failure recorded a generated principle")
	}
}
