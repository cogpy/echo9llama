package cognitivecore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func outcomeTestLabel(value bool) *bool { return &value }
func outcomeTestApply(t *testing.T, learner *OutcomeLearner, example LearningExample) OutcomeMetrics {
	t.Helper()
	metrics, err := learner.Apply(example)
	if err != nil {
		t.Fatalf("Apply(%d, %q): %v", example.Sequence, example.EventID, err)
	}
	return metrics
}

func outcomeTestCheckpoint(t *testing.T, learner *OutcomeLearner) []byte {
	t.Helper()
	encoded, err := json.Marshal(learner.Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func outcomeTestReseal(t *testing.T, snapshot *OutcomeLearnerSnapshot) {
	t.Helper()
	checksum, err := outcomeSnapshotChecksum(*snapshot)
	if err != nil {
		t.Fatal(err)
	}
	snapshot.CanonicalSHA256 = checksum
}

func outcomeTestMaxDistance(a, b []float64) float64 {
	distance := 0.0
	for i := range a {
		distance = math.Max(distance, math.Abs(a[i]-b[i]))
	}
	return distance
}

func TestOutcomeLearnerAllContentFeaturesAndNoTargetLeakage(t *testing.T) {
	prefix := strings.Repeat("m", 100)
	first := LearningExample{Sequence: 1, EventID: "content", Text: prefix + " astronomy telescope nebula", TaskKind: OutcomeTaskGeneral}
	second := first
	second.Text = prefix + " botany photosynthesis chlorophyll"
	x, err := OutcomeFeatures(first)
	if err != nil {
		t.Fatal(err)
	}
	y, err := OutcomeFeatures(second)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(x, y) {
		t.Fatal("content after identical 100-byte prefix was discarded")
	}
	left, right := NewOutcomeLearner(), NewOutcomeLearner()
	outcomeTestApply(t, left, first)
	outcomeTestApply(t, right, second)
	if reflect.DeepEqual(left.Snapshot().State, right.Snapshot().State) {
		t.Fatal("distinct semantic content did not change numerical state")
	}

	// Change all target/metadata fields but leave predictor inputs fixed.
	second = first
	second.Sequence = 900
	second.EventID = "other-metadata"
	second.Passed = outcomeTestLabel(true)
	second.HeldOut = true
	y, err = OutcomeFeatures(second)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(x, y) {
		t.Fatal("label/split/identity/sequence leaked into features")
	}
	passed, failed := first, first
	passed.Passed, failed.Passed = outcomeTestLabel(true), outcomeTestLabel(false)
	p := outcomeTestApply(t, NewOutcomeLearner(), passed)
	f := outcomeTestApply(t, NewOutcomeLearner(), failed)
	if p.Probability != f.Probability || p.Logit != f.Logit || p.PredictionID != f.PredictionID || p.Probability != 0.5 {
		t.Fatal("prediction was made after consuming current label")
	}
	if p.Loss == nil || *p.Loss != math.Log(2) || !p.Trained || p.VersionBefore != 0 || p.VersionAfter != 1 {
		t.Fatalf("incorrect pre-update metrics: %+v", p)
	}

	arithmetic := first
	arithmetic.TaskKind = OutcomeTaskArithmeticArtifact
	arithmetic.Numeric.A, arithmetic.Numeric.B = 17, -4
	z, err := OutcomeFeatures(arithmetic)
	if err != nil {
		t.Fatal(err)
	}
	if z[68] != 1 || z[70] != math.Tanh(17.0/16) || z[71] != math.Tanh(-4.0/16) {
		t.Fatal("typed arithmetic features not used")
	}
}

func TestOutcomeLearnerUsesActualWinWrecAndHistory(t *testing.T) {
	initial := NewOutcomeLearner().Snapshot()
	if !reflect.DeepEqual(initial, NewOutcomeLearner().Snapshot()) {
		t.Fatal("fixed seed is not deterministic")
	}
	input := LearningExample{Sequence: 1, EventID: "initial", Text: "signal frequency analysis", TaskKind: OutcomeTaskNumeric, Numeric: TaskNumericFeatures{Value: 0.7}}
	x, err := OutcomeFeatures(input)
	if err != nil {
		t.Fatal(err)
	}
	base := NewOutcomeLearner()
	outcomeTestApply(t, base, input)
	h := base.Snapshot().State
	for i := range OutcomeUnits {
		activation := 0.0
		for j, value := range x {
			activation += initial.Win[i][j] * value
		}
		want := initial.Leak * math.Tanh(activation)
		if h[i] != want {
			t.Fatalf("unit %d recurrence did not use actual Win: got %.17g, want %.17g", i, h[i], want)
		}
	}
	zeroWin := NewOutcomeLearner()
	for i := range zeroWin.state.Win {
		clear(zeroWin.state.Win[i])
	}
	outcomeTestApply(t, zeroWin, input)
	if reflect.DeepEqual(h, zeroWin.Snapshot().State) {
		t.Fatal("Win is merely decorative")
	}

	withRec, withoutRec := NewOutcomeLearner(), NewOutcomeLearner()
	outcomeTestApply(t, withRec, input)
	outcomeTestApply(t, withoutRec, input)
	for i := range withoutRec.state.Wrec {
		clear(withoutRec.state.Wrec[i])
	}
	next := LearningExample{Sequence: 2, EventID: "next", Text: "independent measurement"}
	outcomeTestApply(t, withRec, next)
	outcomeTestApply(t, withoutRec, next)
	if reflect.DeepEqual(withRec.Snapshot().State, withoutRec.Snapshot().State) {
		t.Fatal("Wrec is merely decorative")
	}
	noHistory := NewOutcomeLearner()
	outcomeTestApply(t, noHistory, next)
	if reflect.DeepEqual(withRec.Snapshot().State, noHistory.Snapshot().State) {
		t.Fatal("state ignores input history")
	}
}

func TestOutcomeLearnerContractivityAndFiniteBounds(t *testing.T) {
	left, right := NewOutcomeLearner(), NewOutcomeLearner()
	maxRowNorm := 0.0
	for _, row := range left.state.Wrec {
		norm := 0.0
		for _, weight := range row {
			norm += math.Abs(weight)
		}
		if norm >= 1 {
			t.Fatalf("noncontractive row norm %.17g", norm)
		}
		maxRowNorm = math.Max(maxRowNorm, norm)
	}
	contraction := (1 - outcomeLeak) + outcomeLeak*maxRowNorm
	if contraction >= 1 {
		t.Fatalf("contraction factor = %g", contraction)
	}
	for i := range right.state.State {
		right.state.State[i] = 0.9
	}
	for sequence := int64(1); sequence <= 512; sequence++ {
		before := outcomeTestMaxDistance(left.state.State, right.state.State)
		example := LearningExample{Sequence: sequence, EventID: fmt.Sprintf("bound-%d", sequence), Text: strings.Repeat("strong signal ", 200), Numeric: TaskNumericFeatures{Value: math.MaxFloat64, A: -math.MaxFloat64, B: math.MaxFloat64, Complexity: 1, InputUnits: math.MaxFloat64, BudgetUnits: math.MaxFloat64}, Passed: outcomeTestLabel(sequence%2 == 0)}
		metrics := outcomeTestApply(t, left, example)
		outcomeTestApply(t, right, example)
		after := outcomeTestMaxDistance(left.state.State, right.state.State)
		// Floating-point roundoff is allowed after histories nearly coalesce.
		if after > contraction*before+1e-14 {
			t.Fatalf("contraction violated: %.17g > %.17g", after, contraction*before)
		}
		for _, value := range left.state.State {
			if !outcomeFinite(value) || math.Abs(value) > 1 {
				t.Fatalf("state escaped finite invariant: %g", value)
			}
		}
		for _, value := range left.state.Readout {
			if !outcomeFinite(value) || math.Abs(value) > outcomeReadoutBound {
				t.Fatalf("readout escaped projection: %g", value)
			}
		}
		if metrics.Loss == nil || !outcomeFinite(*metrics.Loss) || !outcomeFinite(metrics.Logit) || !outcomeFinite(metrics.Probability) {
			t.Fatal("nonfinite prediction/loss")
		}
	}
	for _, z := range []float64{-528, 0, 528} {
		p := outcomeSigmoid(z)
		if !outcomeFinite(p) || p < 0 || p > 1 {
			t.Fatalf("unstable sigmoid(%g)", z)
		}
	}
	t.Logf("computed ||Wrec||_inf=%.17g; recurrence Lipschitz bound=%.17g", maxRowNorm, contraction)
}

func TestOutcomeLearnerSGDDeltaUnknownAndHeldOut(t *testing.T) {
	learner := NewOutcomeLearner()
	before := learner.Snapshot()
	example := LearningExample{Sequence: 1, EventID: "explicit-pass", Text: "evaluate this task", Passed: outcomeTestLabel(true)}
	metrics := outcomeTestApply(t, learner, example)
	after := learner.Snapshot()
	for i := range OutcomeUnits {
		want := outcomeLearningRate * 0.5 * after.State[i]
		if after.Readout[i]-before.Readout[i] != want {
			t.Fatalf("unit %d actual SGD delta not gradient: %g vs %g", i, after.Readout[i]-before.Readout[i], want)
		}
	}
	if after.Readout[OutcomeUnits] != outcomeLearningRate*0.5 || reflect.DeepEqual(before.Readout, after.Readout) || !metrics.Trained {
		t.Fatal("explicit independent outcome did not train actual weights")
	}
	if !reflect.DeepEqual(before.Win, after.Win) || !reflect.DeepEqual(before.Wrec, after.Wrec) {
		t.Fatal("fixed reservoir weights were trained")
	}

	unknown := LearningExample{Sequence: 2, EventID: "unknown", Text: "PASS success certainty failed", HeldOut: false}
	m := outcomeTestApply(t, learner, unknown)
	unknownSnapshot := learner.Snapshot()
	if m.Trained || m.Loss != nil || !reflect.DeepEqual(after.Readout, unknownSnapshot.Readout) || after.Optimizer != unknownSnapshot.Optimizer || after.ReadoutVersion != unknownSnapshot.ReadoutVersion {
		t.Fatal("unknown or textual success proxy trained")
	}
	if reflect.DeepEqual(after.State, unknownSnapshot.State) || unknownSnapshot.Counts.Unknown != 1 {
		t.Fatal("unknown observation did not advance documented history")
	}

	heldOut := LearningExample{Sequence: 3, EventID: "held-out", Text: "fixed evaluation input", Passed: outcomeTestLabel(false), HeldOut: true}
	m = outcomeTestApply(t, learner, heldOut)
	heldSnapshot := learner.Snapshot()
	if m.Trained || m.Loss == nil || m.MeanHeldOutLoss == nil || heldSnapshot.Counts.HeldOut != 1 || !reflect.DeepEqual(unknownSnapshot.Readout, heldSnapshot.Readout) || unknownSnapshot.Optimizer != heldSnapshot.Optimizer {
		t.Fatal("held-out label trained or missing evaluation metric")
	}
	if reflect.DeepEqual(unknownSnapshot.State, heldSnapshot.State) {
		t.Fatal("held-out state did not advance as documented")
	}

	// Opposite held-out labels cannot influence subsequent state, prediction or weights.
	left, right := NewOutcomeLearner(), NewOutcomeLearner()
	heldOut.Sequence, heldOut.EventID = 1, "split"
	heldOut.Passed = outcomeTestLabel(true)
	outcomeTestApply(t, left, heldOut)
	heldOut.Passed = outcomeTestLabel(false)
	outcomeTestApply(t, right, heldOut)
	next := LearningExample{Sequence: 2, EventID: "after-split", Text: "next unchanged task"}
	a, b := outcomeTestApply(t, left, next), outcomeTestApply(t, right, next)
	if a.Probability != b.Probability || a.Logit != b.Logit || !reflect.DeepEqual(left.Snapshot().State, right.Snapshot().State) || !reflect.DeepEqual(left.Snapshot().Readout, right.Snapshot().Readout) {
		t.Fatal("held-out label leaked into predictor")
	}
}

func TestOutcomeLearnerFixedNumericFixtureLoss(t *testing.T) {
	// Fixed simple signed-scalar input/explicit-label fixture. These labels are not
	// model outputs, evaluator proxies, production examples, or any AGI benchmark.
	values := []float64{-2, 2, -1, 1, -0.5, 0.5, -0.25, 0.25}
	labels := []bool{false, true, false, true, false, true, false, true}
	learner := NewOutcomeLearner()
	sequence := int64(0)
	firstEpochLoss, lastEpochLoss, allLoss := 0.0, 0.0, 0.0
	const epochs = 120
	for epoch := range epochs {
		for index, value := range values {
			sequence++
			m := outcomeTestApply(t, learner, LearningExample{Sequence: sequence, EventID: fmt.Sprintf("numeric-train-%d", sequence), TaskKind: OutcomeTaskNumeric, Numeric: TaskNumericFeatures{Value: value}, Passed: outcomeTestLabel(labels[index])})
			if m.Loss == nil {
				t.Fatal("explicit fixture label missing loss")
			}
			allLoss += *m.Loss
			if epoch == 0 {
				firstEpochLoss += *m.Loss
			}
			if epoch == epochs-1 {
				lastEpochLoss += *m.Loss
			}
		}
	}
	frozenWeights := append([]float64(nil), learner.Snapshot().Readout...)
	heldOutLoss := 0.0
	for index, value := range values {
		sequence++
		m := outcomeTestApply(t, learner, LearningExample{Sequence: sequence, EventID: fmt.Sprintf("numeric-heldout-%d", sequence), TaskKind: OutcomeTaskNumeric, Numeric: TaskNumericFeatures{Value: value}, Passed: outcomeTestLabel(labels[index]), HeldOut: true})
		heldOutLoss += *m.Loss
	}
	s := learner.Snapshot()
	meanFirst := firstEpochLoss / float64(len(values))
	meanLast := lastEpochLoss / float64(len(values))
	meanHeldOut := heldOutLoss / float64(len(values))
	meanAll := allLoss / float64(epochs*len(values))
	if !(meanLast < meanFirst && meanHeldOut < math.Log(2)) {
		t.Fatalf("fixed fixture loss did not improve: first=%g last=%g heldout=%g", meanFirst, meanLast, meanHeldOut)
	}
	if s.Counts.PrequentialLossSum != allLoss || s.Counts.HeldOutLossSum != heldOutLoss || !reflect.DeepEqual(frozenWeights, s.Readout) {
		t.Fatal("split loss accounting or held-out isolation mismatch")
	}
	t.Logf("fixed labelled scalar fixture only: first-epoch prequential loss=%.6f; last=%.6f; all=%.6f (n=%d); chronological held-out loss=%.6f (n=%d)", meanFirst, meanLast, meanAll, epochs*len(values), meanHeldOut, len(values))
}

func TestOutcomeLearnerDedupConflictsAndBoundedProof(t *testing.T) {
	learner := NewOutcomeLearner()
	example := LearningExample{Sequence: 1, EventID: "immutable", Text: "source task", Passed: outcomeTestLabel(true)}
	first := outcomeTestApply(t, learner, example)
	checkpoint := outcomeTestCheckpoint(t, learner)
	replayed := outcomeTestApply(t, learner, example)
	if !replayed.Duplicate {
		t.Fatal("duplicate not identified")
	}
	replayed.Duplicate = false
	if !reflect.DeepEqual(first, replayed) || !bytes.Equal(checkpoint, outcomeTestCheckpoint(t, learner)) {
		t.Fatal("duplicate changed metrics or state")
	}
	for _, changed := range []LearningExample{
		{Sequence: 1, EventID: "immutable", Text: "altered task", Passed: outcomeTestLabel(true)},
		{Sequence: 1, EventID: "immutable", Text: "source task", Passed: outcomeTestLabel(false)},
		{Sequence: 2, EventID: "immutable", Text: "source task", Passed: outcomeTestLabel(true)},
	} {
		if _, err := learner.Apply(changed); !errors.Is(err, ErrOutcomeConflict) {
			t.Fatalf("conflicting ID not rejected: %v", err)
		}
		if !bytes.Equal(checkpoint, outcomeTestCheckpoint(t, learner)) {
			t.Fatal("conflict mutated checkpoint")
		}
	}
	unknown := LearningExample{Sequence: 2, EventID: "unknown-immutable"}
	outcomeTestApply(t, learner, unknown)
	unknown.Passed = outcomeTestLabel(true)
	if _, err := learner.Apply(unknown); !errors.Is(err, ErrOutcomeConflict) {
		t.Fatalf("unknown upgraded later and trained: %v", err)
	}
	for seq := int64(3); seq <= OutcomeRecentLimit+5; seq++ {
		outcomeTestApply(t, learner, LearningExample{Sequence: seq, EventID: fmt.Sprintf("window-%d", seq)})
	}
	s := learner.Snapshot()
	if len(s.Cursor.Recent) != OutcomeRecentLimit || len(learner.recent) != OutcomeRecentLimit || len(s.Cursor.SeenIDBits) != outcomeSeenWords {
		t.Fatal("deduplication memory is not bounded")
	}
	checkpoint = outcomeTestCheckpoint(t, learner)
	if _, err := learner.Apply(example); !errors.Is(err, ErrOutcomeStale) {
		t.Fatalf("evicted sequence accepted: %v", err)
	}
	example.Sequence = s.Cursor.LastSequence + 1
	if _, err := learner.Apply(example); !errors.Is(err, ErrOutcomeIDPossiblySeen) {
		t.Fatalf("evicted ID reused at fresh sequence: %v", err)
	}
	if !bytes.Equal(checkpoint, outcomeTestCheckpoint(t, learner)) {
		t.Fatal("evicted duplicate mutated state")
	}

	// Force a false positive. Guard must reject, never take a probabilistic update.
	positions := outcomeIDPositions("never-accepted")
	for _, bit := range positions {
		learner.state.Cursor.SeenIDBits[bit/64] |= uint64(1) << (bit % 64)
	}
	checkpoint = outcomeTestCheckpoint(t, learner)
	if _, err := learner.Apply(LearningExample{Sequence: 1000, EventID: "never-accepted", Passed: outcomeTestLabel(true)}); !errors.Is(err, ErrOutcomeIDPossiblySeen) {
		t.Fatalf("guard failed open: %v", err)
	}
	if !bytes.Equal(checkpoint, outcomeTestCheckpoint(t, learner)) {
		t.Fatal("guard rejection trained")
	}
}

func TestOutcomeLearnerCheckpointSuffixExactReplay(t *testing.T) {
	original := NewOutcomeLearner()
	var examples []LearningExample
	for seq := int64(1); seq <= 340; seq++ {
		example := LearningExample{Sequence: seq * 3, EventID: fmt.Sprintf("replay-%d", seq), Text: fmt.Sprintf("task input %d numerical study", seq), TaskKind: OutcomeTaskNumeric, Numeric: TaskNumericFeatures{Value: float64(seq%9) - 4}, HeldOut: seq%7 == 0}
		if seq%5 != 0 {
			example.Passed = outcomeTestLabel(seq%2 == 0)
		}
		examples = append(examples, example)
	}
	for _, example := range examples[:290] {
		outcomeTestApply(t, original, example)
	}
	checkpoint := outcomeTestCheckpoint(t, original)
	restored := NewOutcomeLearner()
	if err := restored.RestoreJSON(checkpoint); err != nil {
		t.Fatalf("restore valid JSON: %v", err)
	}
	if !bytes.Equal(checkpoint, outcomeTestCheckpoint(t, restored)) {
		t.Fatal("restored checkpoint differs bitwise")
	}
	beforeDuplicate := outcomeTestCheckpoint(t, restored)
	m := outcomeTestApply(t, restored, examples[289])
	if !m.Duplicate || !bytes.Equal(beforeDuplicate, outcomeTestCheckpoint(t, restored)) {
		t.Fatal("restored recent duplicate retrained")
	}
	if _, err := restored.Apply(examples[0]); !errors.Is(err, ErrOutcomeStale) {
		t.Fatalf("restored evicted duplicate accepted: %v", err)
	}
	evicted := examples[0]
	evicted.Sequence = 10000
	if _, err := restored.Apply(evicted); !errors.Is(err, ErrOutcomeIDPossiblySeen) {
		t.Fatalf("restored lifetime ID guard lost evicted ID: %v", err)
	}
	for _, example := range examples[290:] {
		a, b := outcomeTestApply(t, original, example), outcomeTestApply(t, restored, example)
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("prediction/loss suffix differs at %d", example.Sequence)
		}
		if !bytes.Equal(outcomeTestCheckpoint(t, original), outcomeTestCheckpoint(t, restored)) {
			t.Fatalf("numerical/checksum suffix differs at %d", example.Sequence)
		}
	}
}

func TestOutcomeLearnerStrictRestoreCorruption(t *testing.T) {
	learner := NewOutcomeLearner()
	outcomeTestApply(t, learner, LearningExample{Sequence: 3, EventID: "checkpoint", Text: "bounded input", Passed: outcomeTestLabel(true)})
	checkpoint := outcomeTestCheckpoint(t, learner)
	snapshot := learner.Snapshot()
	snapshot.State[0] += 0.01
	if err := learner.Restore(snapshot); !errors.Is(err, ErrInvalidOutcomeSnapshot) {
		t.Fatalf("corruption without checksum rejected incorrectly: %v", err)
	}
	if !bytes.Equal(checkpoint, outcomeTestCheckpoint(t, learner)) {
		t.Fatal("failed restore mutated existing learner")
	}
	mutations := []struct {
		name   string
		modify func(*OutcomeLearnerSnapshot)
	}{
		{"version", func(s *OutcomeLearnerSnapshot) { s.SnapshotVersion++ }},
		{"feature-version", func(s *OutcomeLearnerSnapshot) { s.FeatureVersion = "other" }},
		{"state-dimension", func(s *OutcomeLearnerSnapshot) { s.State = s.State[:31] }},
		{"extra-matrix-column", func(s *OutcomeLearnerSnapshot) { s.Win[0] = append(s.Win[0], 0) }},
		{"missing-rec-row", func(s *OutcomeLearnerSnapshot) { s.Wrec = s.Wrec[:31] }},
		{"readout-dimension", func(s *OutcomeLearnerSnapshot) { s.Readout = append(s.Readout, 0) }},
		{"state-bound", func(s *OutcomeLearnerSnapshot) { s.State[0] = 1.1 }},
		{"readout-bound", func(s *OutcomeLearnerSnapshot) { s.Readout[0] = 17 }},
		{"noncontractive", func(s *OutcomeLearnerSnapshot) { s.Wrec[0][0] = 1 }},
		{"counter", func(s *OutcomeLearnerSnapshot) { s.Optimizer.Steps++ }},
		{"cursor", func(s *OutcomeLearnerSnapshot) { s.Cursor.LastSequence++ }},
		{"receipt-guard", func(s *OutcomeLearnerSnapshot) { clear(s.Cursor.SeenIDBits) }},
		{"receipt-label", func(s *OutcomeLearnerSnapshot) { s.Cursor.Recent[0].Metrics.HeldOut = true }},
		{"loss-availability", func(s *OutcomeLearnerSnapshot) { s.Cursor.Recent[0].Metrics.MeanPrequentialLoss = nil }},
		{"data-hash", func(s *OutcomeLearnerSnapshot) { s.DataSHA256 = "not-a-hash" }},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			s := learner.Snapshot()
			mutation.modify(&s)
			outcomeTestReseal(t, &s) // Structural rejection even with a recomputed checksum.
			if err := learner.Restore(s); !errors.Is(err, ErrInvalidOutcomeSnapshot) {
				t.Fatalf("malformed resealed snapshot accepted: %v", err)
			}
			if !bytes.Equal(checkpoint, outcomeTestCheckpoint(t, learner)) {
				t.Fatal("invalid restore mutated learner")
			}
		})
	}
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		s := learner.Snapshot()
		s.Wrec[0][0] = value
		if err := learner.Restore(s); !errors.Is(err, ErrInvalidOutcomeSnapshot) {
			t.Fatalf("nonfinite snapshot accepted: %v", err)
		}
	}
	for _, badJSON := range [][]byte{
		append(append([]byte(nil), checkpoint...), []byte(" {}")...),
		bytes.Replace(checkpoint, []byte("\"snapshot_version\":"), []byte("\"unknown_field\":1,\"snapshot_version\":"), 1),
		[]byte("{\"state\":[NaN]}"),
	} {
		if err := learner.RestoreJSON(badJSON); !errors.Is(err, ErrInvalidOutcomeSnapshot) {
			t.Fatalf("invalid JSON accepted: %v", err)
		}
	}

	// Detached snapshots and receipts must not share mutable slices/pointers.
	detached := learner.Snapshot()
	detached.Readout[0] = 12
	detached.Cursor.SeenIDBits[0] = 0
	*detached.Cursor.Recent[0].Metrics.Loss = 100
	if !bytes.Equal(checkpoint, outcomeTestCheckpoint(t, learner)) {
		t.Fatal("Snapshot aliases learner state")
	}
	restored := NewOutcomeLearner()
	detached = learner.Snapshot()
	if err := restored.Restore(detached); err != nil {
		t.Fatal(err)
	}
	detached.Wrec[0][0] = 100
	*detached.Cursor.Recent[0].Metrics.Loss = 100
	if !bytes.Equal(checkpoint, outcomeTestCheckpoint(t, restored)) {
		t.Fatal("Restore aliases supplied state")
	}
}

func TestOutcomeLearnerInvalidExamplesDoNotMutate(t *testing.T) {
	learner := NewOutcomeLearner()
	checkpoint := outcomeTestCheckpoint(t, learner)
	invalid := []LearningExample{
		{Sequence: 0, EventID: "zero"},
		{Sequence: -1, EventID: "negative"},
		{Sequence: 1, EventID: ""},
		{Sequence: 1, EventID: " whitespace "},
		{Sequence: 1, EventID: strings.Repeat("x", 257)},
		{Sequence: 1, EventID: "oversized", Text: strings.Repeat("x", outcomeMaxTextBytes+1)},
		{Sequence: 1, EventID: "utf8", Text: string([]byte{0xff})},
		{Sequence: 1, EventID: "kind", TaskKind: "success-proxy"},
		{Sequence: 1, EventID: "nan", Numeric: TaskNumericFeatures{Value: math.NaN()}},
		{Sequence: 1, EventID: "inf", Numeric: TaskNumericFeatures{A: math.Inf(1)}},
		{Sequence: 1, EventID: "complexity", Numeric: TaskNumericFeatures{Complexity: 2}},
		{Sequence: 1, EventID: "negative-units", Numeric: TaskNumericFeatures{InputUnits: -1}},
	}
	for _, example := range invalid {
		if _, err := learner.Apply(example); !errors.Is(err, ErrInvalidLearningExample) {
			t.Fatalf("invalid example accepted: %+v, %v", example, err)
		}
		if !bytes.Equal(checkpoint, outcomeTestCheckpoint(t, learner)) {
			t.Fatal("invalid input mutated state")
		}
	}
}

func TestOutcomeLearnerConcurrentSnapshots(t *testing.T) {
	learner := NewOutcomeLearner()
	var workers sync.WaitGroup
	workers.Add(3)
	for range 3 {
		go func() {
			defer workers.Done()
			for range 30 {
				snapshot := learner.Snapshot()
				if err := NewOutcomeLearner().Restore(snapshot); err != nil {
					t.Errorf("concurrent snapshot invalid: %v", err)
				}
			}
		}()
	}
	for seq := int64(1); seq <= 100; seq++ {
		outcomeTestApply(t, learner, LearningExample{Sequence: seq, EventID: fmt.Sprintf("concurrent-%d", seq), Passed: outcomeTestLabel(seq%2 == 0)})
	}
	workers.Wait()
}
