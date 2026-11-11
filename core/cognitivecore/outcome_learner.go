package cognitivecore

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

const (
	OutcomeLearnerVersion  = "auxiliary-outcome-learner-v1"
	OutcomeFeatureVersion  = "unicode-token-hash-task-numeric-v1"
	OutcomeSnapshotVersion = 1
	OutcomeUnits           = 32
	OutcomeFeatureCount    = 75 // 64 token bins, five task kinds, six numeric fields.
	OutcomeRecentLimit     = 256
	outcomeTokenBins       = 64
	outcomeSeenWords       = 4096 // 262144 bits; bounded fail-closed lifetime ID guard.
	outcomeMaxTextBytes    = 64 * 1024
	outcomeLeak            = 0.35
	outcomeLearningRate    = 0.15
	outcomeReadoutBound    = 16.0
)

var (
	ErrInvalidLearningExample = errors.New("invalid outcome learning example")
	ErrOutcomeConflict        = errors.New("outcome event conflicts with its immutable receipt")
	ErrOutcomeStale           = errors.New("outcome sequence is at or below the applied cursor")
	ErrOutcomeIDPossiblySeen  = errors.New("outcome event ID was previously seen or collided with the fail-closed ID guard")
	ErrInvalidOutcomeSnapshot = errors.New("invalid outcome learner snapshot")
)

// OutcomeTaskKind identifies pre-outcome task categories, not success categories.
// The empty kind is general. Arbitrary outcome-derived category strings are rejected.
type OutcomeTaskKind string

const (
	OutcomeTaskGeneral            OutcomeTaskKind = "general"
	OutcomeTaskNumeric            OutcomeTaskKind = "numeric"
	OutcomeTaskTool               OutcomeTaskKind = "tool"
	OutcomeTaskPlanning           OutcomeTaskKind = "planning"
	OutcomeTaskArithmeticArtifact OutcomeTaskKind = "arithmetic-artifact-v1"
)

// TaskNumericFeatures contains only facts available before independent evaluation.
// Value is a signed task input; Complexity is in [0,1]; units are nonnegative.
// These are task descriptors, never an evaluator score or an encoded target.
type TaskNumericFeatures struct {
	Value       float64 `json:"value"`
	A           float64 `json:"a"`
	B           float64 `json:"b"`
	Complexity  float64 `json:"complexity"`
	InputUnits  float64 `json:"input_units"`
	BudgetUnits float64 `json:"budget_units"`
}

// LearningExample is one immutable, ordered event from an authoritative spine.
// Text must be the complete accepted PRE-OUTCOME task content, not evaluator output.
// Passed must come exclusively from an explicit independent pass/fail evaluator;
// nil means unknown and cannot be upgraded later under the same event ID. The caller
// must validate evidence, temporal causality, and independence before calling Apply.
// This numerical lane does not verify evaluator evidence or infer labels from text.
type LearningExample struct {
	Sequence int64               `json:"sequence"`
	EventID  string              `json:"event_id"`
	Text     string              `json:"text"`
	TaskKind OutcomeTaskKind     `json:"task_kind"`
	Numeric  TaskNumericFeatures `json:"numeric"`
	Passed   *bool               `json:"passed"`
	HeldOut  bool                `json:"held_out"`
}

// OutcomeMetrics reports the prediction BEFORE this example's readout update.
// Loss is absent for unknowns; means are absent until that split has explicit labels.
// These are task-outcome metrics, not measures of intelligence or consciousness.
type OutcomeMetrics struct {
	Sequence            int64    `json:"sequence"`
	EventID             string   `json:"event_id"`
	PredictionID        string   `json:"prediction_id"`
	FeatureSHA256       string   `json:"feature_sha256"`
	Probability         float64  `json:"probability"`
	Logit               float64  `json:"logit"`
	Loss                *float64 `json:"loss"`
	HeldOut             bool     `json:"held_out"`
	Trained             bool     `json:"trained"`
	Duplicate           bool     `json:"duplicate"`
	VersionBefore       uint64   `json:"version_before"`
	VersionAfter        uint64   `json:"version_after"`
	TrainingExamples    uint64   `json:"training_examples"`
	HeldOutExamples     uint64   `json:"held_out_examples"`
	MeanPrequentialLoss *float64 `json:"mean_prequential_loss"`
	MeanHeldOutLoss     *float64 `json:"mean_held_out_loss"`
}

type OutcomeOptimizer struct {
	Version      string  `json:"version"`
	LearningRate float64 `json:"learning_rate"`
	Steps        uint64  `json:"steps"`
}

type OutcomeCounts struct {
	Observed           uint64  `json:"observed"`
	Training           uint64  `json:"training"`
	HeldOut            uint64  `json:"held_out"`
	Unknown            uint64  `json:"unknown"`
	PrequentialLossSum float64 `json:"prequential_loss_sum"`
	HeldOutLossSum     float64 `json:"held_out_loss_sum"`
}

// OutcomeReceipt binds the full example (including its explicit label) to a cached
// immutable prediction. Only hashes, not task text or evaluator evidence, persist.
type OutcomeReceipt struct {
	Sequence      int64          `json:"sequence"`
	EventID       string         `json:"event_id"`
	ExampleSHA256 string         `json:"example_sha256"`
	Metrics       OutcomeMetrics `json:"metrics"`
}

// OutcomeCursor has an exact recent receipt window and an irrevocable high-water
// sequence. Evicted old sequences always fail closed. SeenIDBits is a four-probe
// Bloom guard: it has NO false negatives for accepted IDs, but may reject a genuinely
// new ID. Bits are never cleared, including on restore. Saturation reduces availability
// rather than permitting retraining; no unbounded map or probabilistic acceptance.
type OutcomeCursor struct {
	LastSequence int64            `json:"last_sequence"`
	Recent       []OutcomeReceipt `json:"recent"`
	SeenIDBits   []uint64         `json:"seen_id_bits"`
}

// OutcomeLearnerSnapshot is a complete JSON checkpoint, including fixed numerical
// matrices and SGD state. CanonicalSHA256 hashes standard-library JSON serialization
// of this typed struct with CanonicalSHA256 empty, prefixed by a versioned domain.
// Struct field order and shortest round-trippable finite float encoding are stable;
// no maps, wall clocks or external state enter the checksum. This is integrity, NOT
// authentication: callers must protect checkpoints against malicious re-signing.
type OutcomeLearnerSnapshot struct {
	SnapshotVersion int              `json:"snapshot_version"`
	ModelVersion    string           `json:"model_version"`
	FeatureVersion  string           `json:"feature_version"`
	ReadoutVersion  uint64           `json:"readout_version"`
	Leak            float64          `json:"leak"`
	State           []float64        `json:"state"`
	Win             [][]float64      `json:"win"`
	Wrec            [][]float64      `json:"wrec"`
	Readout         []float64        `json:"readout"` // 32 state coefficients followed by bias.
	Optimizer       OutcomeOptimizer `json:"optimizer"`
	Counts          OutcomeCounts    `json:"counts"`
	Cursor          OutcomeCursor    `json:"cursor"`
	DataSHA256      string           `json:"data_sha256"`
	CanonicalSHA256 string           `json:"canonical_sha256"`
}

// OutcomeLearner is a small auxiliary reservoir/logistic readout. It creates no
// goroutines, clocks, providers or synthetic task text. Its mutex permits concurrent
// snapshots, but callers must submit Apply in authoritative sequence order.
//
// h'=(1-a)h+a*tanh(Win*x+Wrec*h), a=0.35. Initialization scales every Wrec row
// to absolute row sum 0.65, so its infinity norm is <1. As tanh is 1-Lipschitz,
// histories under the same input contract by at most (1-a)+a*||Wrec||inf <1;
// [-1,1]^32 is invariant. The fixed reservoir is not trained. Explicit non-held-out
// labels alone drive projected logistic SGD w'=clip(w+eta*(y-p)*[h',1],-16,16).
// Projection bounds logits and numerical loss; all reported loss is pre-update.
// Held-out/unknown events advance h and the cursor, but NOT readout or optimizer.
// Thus held-out loss is chronological online evaluation, not an isolated reset-state
// benchmark; held-out labels do not affect later states or predictor features.
type OutcomeLearner struct {
	mu     sync.Mutex
	state  OutcomeLearnerSnapshot
	recent map[string]OutcomeReceipt
}

// NewOutcomeLearner constructs the same fixed-seed reservoir on every invocation.
func NewOutcomeLearner() *OutcomeLearner {
	s := OutcomeLearnerSnapshot{
		SnapshotVersion: OutcomeSnapshotVersion, ModelVersion: OutcomeLearnerVersion,
		FeatureVersion: OutcomeFeatureVersion, Leak: outcomeLeak,
		State: make([]float64, OutcomeUnits), Win: make([][]float64, OutcomeUnits),
		Wrec: make([][]float64, OutcomeUnits), Readout: make([]float64, OutcomeUnits+1),
		Optimizer:  OutcomeOptimizer{Version: "projected-logistic-sgd-v1", LearningRate: outcomeLearningRate},
		Cursor:     OutcomeCursor{Recent: make([]OutcomeReceipt, 0), SeenIDBits: make([]uint64, outcomeSeenWords)},
		DataSHA256: outcomeSHA([]byte("outcome-data-v1\x00")),
	}
	// Local SplitMix64 avoids dependence on math/rand algorithms or shared RNG state.
	seed := uint64(0x6563686f39617578)
	next := func() float64 {
		seed += 0x9e3779b97f4a7c15
		z := seed
		z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
		z = (z ^ (z >> 27)) * 0x94d049bb133111eb
		z ^= z >> 31
		return 2*(float64(z>>11)/float64(uint64(1)<<53)) - 1
	}
	for i := range OutcomeUnits {
		s.Win[i] = make([]float64, OutcomeFeatureCount)
		for j := range OutcomeFeatureCount {
			s.Win[i][j] = 0.5 * next()
		}
		s.Wrec[i] = make([]float64, OutcomeUnits)
		norm := 0.0
		for j := range OutcomeUnits {
			s.Wrec[i][j] = next()
			norm += math.Abs(s.Wrec[i][j])
		}
		for j := range OutcomeUnits {
			s.Wrec[i][j] *= 0.65 / norm
		}
	}
	return &OutcomeLearner{state: s, recent: make(map[string]OutcomeReceipt)}
}

// OutcomeFeatures hashes every Unicode word/number token in the accepted content,
// case-folded with Unicode lowercase, into signed normalized bag-of-token bins.
// It is a lightweight lexical semantic proxy, not an embedding or an understanding
// claim. Text is never prefix-truncated: oversize/invalid UTF-8 input is rejected.
// IDs, sequence, label and HeldOut are deliberately excluded from this function.
func OutcomeFeatures(example LearningExample) ([]float64, error) {
	if len(example.Text) > outcomeMaxTextBytes || !utf8.ValidString(example.Text) {
		return nil, ErrInvalidLearningExample
	}
	n := example.Numeric
	if !outcomeFinite(n.Value) || !outcomeFinite(n.A) || !outcomeFinite(n.B) || !outcomeFinite(n.Complexity) || !outcomeFinite(n.InputUnits) || !outcomeFinite(n.BudgetUnits) || n.Complexity < 0 || n.Complexity > 1 || n.InputUnits < 0 || n.BudgetUnits < 0 {
		return nil, ErrInvalidLearningExample
	}
	kind := 0
	switch example.TaskKind {
	case "", OutcomeTaskGeneral:
	case OutcomeTaskNumeric:
		kind = 1
	case OutcomeTaskTool:
		kind = 2
	case OutcomeTaskPlanning:
		kind = 3
	case OutcomeTaskArithmeticArtifact:
		kind = 4
	default:
		return nil, ErrInvalidLearningExample
	}
	x := make([]float64, OutcomeFeatureCount)
	for _, token := range strings.FieldsFunc(strings.ToLower(example.Text), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) }) {
		h := sha256.Sum256([]byte(token))
		bin := int(binary.LittleEndian.Uint64(h[:8]) % outcomeTokenBins)
		sign := 1.0
		if h[8]&1 != 0 {
			sign = -1
		}
		x[bin] += sign
	}
	norm := 0.0
	for _, value := range x[:outcomeTokenBins] {
		norm += value * value
	}
	if norm > 0 {
		norm = math.Sqrt(norm)
		for i := range outcomeTokenBins {
			x[i] /= norm
		}
	}
	x[outcomeTokenBins+kind] = 1
	x[69] = math.Tanh(n.Value)
	x[70] = math.Tanh(n.A / 16)
	x[71] = math.Tanh(n.B / 16)
	x[72] = n.Complexity
	x[73] = math.Tanh(math.Log1p(n.InputUnits) / 8)
	x[74] = math.Tanh(math.Log1p(n.BudgetUnits) / 8)
	return x, nil
}

// Apply consumes one immutable example atomically. Duplicate recent examples return
// their original metrics with Duplicate=true and cause no numerical/cursor mutation.
// A changed label/text/sequence on a cached ID is a conflict. Any sequence <= cursor
// outside the cache is stale; an ID reused beyond the cache is rejected by the guard.
// Sequence gaps are allowed, reversals are not. Invalid/conflicting inputs never mutate.
func (learner *OutcomeLearner) Apply(example LearningExample) (OutcomeMetrics, error) {
	// Copy caller-owned label once; concurrent mutation by the caller is not supported.
	if example.Passed != nil {
		label := *example.Passed
		example.Passed = &label
	}
	if example.Sequence <= 0 || len(example.EventID) == 0 || len(example.EventID) > 256 || strings.TrimSpace(example.EventID) != example.EventID || !utf8.ValidString(example.EventID) {
		return OutcomeMetrics{}, ErrInvalidLearningExample
	}
	x, err := OutcomeFeatures(example)
	if err != nil {
		return OutcomeMetrics{}, err
	}
	encoded, err := json.Marshal(example)
	if err != nil {
		return OutcomeMetrics{}, ErrInvalidLearningExample
	}
	exampleHash := outcomeSHA(encoded)
	featureBytes, _ := json.Marshal(x)
	featureHash := outcomeSHA(featureBytes)
	learner.mu.Lock()
	defer learner.mu.Unlock()
	s := &learner.state
	if cached, ok := learner.recent[example.EventID]; ok {
		if cached.ExampleSHA256 != exampleHash || cached.Sequence != example.Sequence {
			return OutcomeMetrics{}, ErrOutcomeConflict
		}
		metrics := cloneOutcomeMetrics(cached.Metrics)
		metrics.Duplicate = true
		return metrics, nil
	}
	if example.Sequence <= s.Cursor.LastSequence {
		return OutcomeMetrics{}, ErrOutcomeStale
	}
	positions := outcomeIDPositions(example.EventID)
	if outcomeSeen(s.Cursor.SeenIDBits, positions) {
		return OutcomeMetrics{}, ErrOutcomeIDPossiblySeen
	}

	// Predictor uses pre-outcome features and previous state/weights only.
	nextState := make([]float64, OutcomeUnits)
	for i := range OutcomeUnits {
		activation := 0.0
		for j, value := range x {
			activation += s.Win[i][j] * value
		}
		for j, value := range s.State {
			activation += s.Wrec[i][j] * value
		}
		nextState[i] = (1-s.Leak)*s.State[i] + s.Leak*math.Tanh(activation)
	}
	logit := s.Readout[OutcomeUnits]
	for i, value := range nextState {
		logit += s.Readout[i] * value
	}
	probability := outcomeSigmoid(logit)
	predictionBytes, _ := json.Marshal(struct {
		Sequence                                 int64
		EventID, FeatureSHA256, PrefixDataSHA256 string
		ReadoutVersion                           uint64
	}{example.Sequence, example.EventID, featureHash, s.DataSHA256, s.ReadoutVersion})
	metrics := OutcomeMetrics{
		Sequence: example.Sequence, EventID: example.EventID,
		PredictionID: outcomeSHA(predictionBytes), FeatureSHA256: featureHash,
		Probability: probability, Logit: logit, HeldOut: example.HeldOut,
		VersionBefore: s.ReadoutVersion,
	}
	s.State = nextState
	s.Counts.Observed++
	if example.Passed == nil {
		s.Counts.Unknown++
	} else {
		y := 0.0
		if *example.Passed {
			y = 1
		}
		// softplus(-z) for pass, softplus(z) for fail: no log(0) or cancellation.
		lossLogit := logit
		if *example.Passed {
			lossLogit = -logit
		}
		loss := math.Max(lossLogit, 0) + math.Log1p(math.Exp(-math.Abs(lossLogit)))
		metrics.Loss = &loss
		if example.HeldOut {
			s.Counts.HeldOut++
			s.Counts.HeldOutLossSum += loss
		} else {
			step := s.Optimizer.LearningRate * (y - probability)
			for i, value := range nextState {
				s.Readout[i] = outcomeClip(s.Readout[i] + step*value)
			}
			s.Readout[OutcomeUnits] = outcomeClip(s.Readout[OutcomeUnits] + step)
			s.Optimizer.Steps++
			s.ReadoutVersion++
			s.Counts.Training++
			s.Counts.PrequentialLossSum += loss
			metrics.Trained = true
		}
	}
	metrics.VersionAfter = s.ReadoutVersion
	metrics.TrainingExamples = s.Counts.Training
	metrics.HeldOutExamples = s.Counts.HeldOut
	if s.Counts.Training > 0 {
		mean := s.Counts.PrequentialLossSum / float64(s.Counts.Training)
		metrics.MeanPrequentialLoss = &mean
	}
	if s.Counts.HeldOut > 0 {
		mean := s.Counts.HeldOutLossSum / float64(s.Counts.HeldOut)
		metrics.MeanHeldOutLoss = &mean
	}
	s.DataSHA256 = outcomeSHA([]byte("outcome-data-v1\x00" + s.DataSHA256 + exampleHash))
	s.Cursor.LastSequence = example.Sequence
	for _, bit := range positions {
		s.Cursor.SeenIDBits[bit/64] |= uint64(1) << (bit % 64)
	}
	receipt := OutcomeReceipt{Sequence: example.Sequence, EventID: example.EventID, ExampleSHA256: exampleHash, Metrics: cloneOutcomeMetrics(metrics)}
	learner.recent[example.EventID] = receipt
	s.Cursor.Recent = append(s.Cursor.Recent, receipt)
	if len(s.Cursor.Recent) > OutcomeRecentLimit {
		delete(learner.recent, s.Cursor.Recent[0].EventID)
		copy(s.Cursor.Recent, s.Cursor.Recent[1:])
		s.Cursor.Recent = s.Cursor.Recent[:OutcomeRecentLimit]
	}
	return metrics, nil
}

// Snapshot returns a detached canonical checkpoint; modifying it cannot mutate the learner.
func (learner *OutcomeLearner) Snapshot() OutcomeLearnerSnapshot {
	learner.mu.Lock()
	defer learner.mu.Unlock()
	s := cloneOutcomeSnapshot(learner.state)
	s.CanonicalSHA256, _ = outcomeSnapshotChecksum(s)
	return s
}

// Restore validates completely before replacing state, and does not alias caller data.
func (learner *OutcomeLearner) Restore(snapshot OutcomeLearnerSnapshot) error {
	s := cloneOutcomeSnapshot(snapshot)
	if err := validateOutcomeSnapshot(s); err != nil {
		return err
	}
	recent := make(map[string]OutcomeReceipt, len(s.Cursor.Recent))
	for _, receipt := range s.Cursor.Recent {
		recent[receipt.EventID] = receipt
	}
	learner.mu.Lock()
	defer learner.mu.Unlock()
	s.CanonicalSHA256 = ""
	learner.state = s
	learner.recent = recent
	return nil
}

// RestoreJSON rejects unknown fields, malformed dimensions, and trailing JSON values.
// A checkpoint must fit a bounded envelope before decoding allocation-heavy slices.
func (learner *OutcomeLearner) RestoreJSON(data []byte) error {
	if len(data) > 1024*1024 {
		return ErrInvalidOutcomeSnapshot
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var snapshot OutcomeLearnerSnapshot
	if err := decoder.Decode(&snapshot); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidOutcomeSnapshot, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return ErrInvalidOutcomeSnapshot
	}
	return learner.Restore(snapshot)
}

func validateOutcomeSnapshot(s OutcomeLearnerSnapshot) error {
	invalid := func(reason string) error { return fmt.Errorf("%w: %s", ErrInvalidOutcomeSnapshot, reason) }
	if s.SnapshotVersion != OutcomeSnapshotVersion || s.ModelVersion != OutcomeLearnerVersion || s.FeatureVersion != OutcomeFeatureVersion || s.Leak != outcomeLeak || s.Optimizer.Version != "projected-logistic-sgd-v1" || s.Optimizer.LearningRate != outcomeLearningRate {
		return invalid("version or algorithm contract")
	}
	if !outcomeValidHash(s.CanonicalSHA256) || !outcomeValidHash(s.DataSHA256) {
		return invalid("hash encoding")
	}
	checksum, err := outcomeSnapshotChecksum(s)
	if err != nil || checksum != s.CanonicalSHA256 {
		return invalid("canonical checksum")
	}
	if len(s.State) != OutcomeUnits || len(s.Win) != OutcomeUnits || len(s.Wrec) != OutcomeUnits || len(s.Readout) != OutcomeUnits+1 || len(s.Cursor.SeenIDBits) != outcomeSeenWords {
		return invalid("dimensions")
	}
	for i := range OutcomeUnits {
		if !outcomeFinite(s.State[i]) || math.Abs(s.State[i]) > 1 || len(s.Win[i]) != OutcomeFeatureCount || len(s.Wrec[i]) != OutcomeUnits {
			return invalid("state or matrix dimensions")
		}
		for _, weight := range s.Win[i] {
			if !outcomeFinite(weight) || math.Abs(weight) > 0.5 {
				return invalid("input weight")
			}
		}
		norm := 0.0
		for _, weight := range s.Wrec[i] {
			if !outcomeFinite(weight) {
				return invalid("recurrent weight")
			}
			norm += math.Abs(weight)
		}
		if !outcomeFinite(norm) || norm >= 1 {
			return invalid("noncontractive recurrent row")
		}
	}
	for _, weight := range s.Readout {
		if !outcomeFinite(weight) || math.Abs(weight) > outcomeReadoutBound {
			return invalid("readout weight")
		}
	}
	c := s.Counts
	if s.Cursor.LastSequence < 0 || c.Observed > uint64(s.Cursor.LastSequence) || c.Training > c.Observed || c.HeldOut > c.Observed-c.Training || c.Unknown != c.Observed-c.Training-c.HeldOut || s.Optimizer.Steps != c.Training || s.ReadoutVersion != c.Training {
		return invalid("counts or cursor")
	}
	for _, sum := range []float64{c.PrequentialLossSum, c.HeldOutLossSum} {
		if !outcomeFinite(sum) || sum < 0 {
			return invalid("loss sums")
		}
	}
	if (c.Training == 0 && c.PrequentialLossSum != 0) || (c.HeldOut == 0 && c.HeldOutLossSum != 0) {
		return invalid("unlabelled loss")
	}
	wantRecent := c.Observed
	if wantRecent > OutcomeRecentLimit {
		wantRecent = OutcomeRecentLimit
	}
	if uint64(len(s.Cursor.Recent)) != wantRecent {
		return invalid("receipt window")
	}
	ids := make(map[string]bool, len(s.Cursor.Recent))
	last := int64(0)
	for _, r := range s.Cursor.Recent {
		m := r.Metrics
		if r.Sequence <= last || r.Sequence > s.Cursor.LastSequence || r.EventID == "" || len(r.EventID) > 256 || strings.TrimSpace(r.EventID) != r.EventID || !utf8.ValidString(r.EventID) || ids[r.EventID] || !outcomeValidHash(r.ExampleSHA256) || !outcomeSeen(s.Cursor.SeenIDBits, outcomeIDPositions(r.EventID)) {
			return invalid("receipt identity")
		}
		if m.Sequence != r.Sequence || m.EventID != r.EventID || m.Duplicate || !outcomeValidHash(m.PredictionID) || !outcomeValidHash(m.FeatureSHA256) || !outcomeFinite(m.Probability) || m.Probability < 0 || m.Probability > 1 || !outcomeFinite(m.Logit) || math.Abs(m.Logit) > float64(OutcomeUnits+1)*outcomeReadoutBound {
			return invalid("receipt prediction")
		}
		if m.VersionBefore > m.VersionAfter || m.VersionAfter != m.TrainingExamples || m.TrainingExamples > c.Training || m.HeldOutExamples > c.HeldOut || (m.Trained && (m.HeldOut || m.Loss == nil || m.VersionAfter-m.VersionBefore != 1)) || (!m.Trained && m.VersionAfter != m.VersionBefore) || (m.Loss != nil && !m.Trained && !m.HeldOut) {
			return invalid("receipt label or optimizer")
		}
		for _, value := range []*float64{m.Loss, m.MeanPrequentialLoss, m.MeanHeldOutLoss} {
			if value != nil && (!outcomeFinite(*value) || *value < 0) {
				return invalid("receipt loss")
			}
		}
		if (m.TrainingExamples == 0) != (m.MeanPrequentialLoss == nil) || (m.HeldOutExamples == 0) != (m.MeanHeldOutLoss == nil) {
			return invalid("receipt loss availability")
		}
		ids[r.EventID] = true
		last = r.Sequence
	}
	if last != s.Cursor.LastSequence {
		return invalid("last receipt cursor")
	}
	if c.Observed == 0 {
		if s.DataSHA256 != outcomeSHA([]byte("outcome-data-v1\x00")) {
			return invalid("empty data hash")
		}
		for _, word := range s.Cursor.SeenIDBits {
			if word != 0 {
				return invalid("empty ID guard")
			}
		}
	}
	return nil
}

func outcomeSnapshotChecksum(s OutcomeLearnerSnapshot) (string, error) {
	s.CanonicalSHA256 = ""
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return outcomeSHA(append([]byte("outcome-checkpoint-v1\x00"), data...)), nil
}

func outcomeSHA(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func outcomeValidHash(hash string) bool {
	decoded, err := hex.DecodeString(hash)
	return len(hash) == 64 && err == nil && len(decoded) == sha256.Size && strings.ToLower(hash) == hash
}
func outcomeFinite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
func outcomeClip(value float64) float64 {
	return math.Max(-outcomeReadoutBound, math.Min(outcomeReadoutBound, value))
}

func outcomeSigmoid(z float64) float64 {
	if z >= 0 {
		return 1 / (1 + math.Exp(-z))
	}
	exp := math.Exp(z)
	return exp / (1 + exp)
}

func outcomeIDPositions(id string) [4]uint64 {
	h := sha256.Sum256([]byte("outcome-id-v1\x00" + id))
	var positions [4]uint64
	for i := range positions {
		positions[i] = binary.LittleEndian.Uint64(h[i*8:(i+1)*8]) % (outcomeSeenWords * 64)
	}
	return positions
}

func outcomeSeen(words []uint64, positions [4]uint64) bool {
	for _, bit := range positions {
		if words[bit/64]&(uint64(1)<<(bit%64)) == 0 {
			return false
		}
	}
	return true
}

func cloneOutcomeMetrics(m OutcomeMetrics) OutcomeMetrics {
	if m.Loss != nil {
		value := *m.Loss
		m.Loss = &value
	}
	if m.MeanPrequentialLoss != nil {
		value := *m.MeanPrequentialLoss
		m.MeanPrequentialLoss = &value
	}
	if m.MeanHeldOutLoss != nil {
		value := *m.MeanHeldOutLoss
		m.MeanHeldOutLoss = &value
	}
	return m
}

func cloneOutcomeSnapshot(s OutcomeLearnerSnapshot) OutcomeLearnerSnapshot {
	s.State = append([]float64(nil), s.State...)
	s.Readout = append([]float64(nil), s.Readout...)
	win, wrec := make([][]float64, len(s.Win)), make([][]float64, len(s.Wrec))
	for i, row := range s.Win {
		win[i] = append([]float64(nil), row...)
	}
	for i, row := range s.Wrec {
		wrec[i] = append([]float64(nil), row...)
	}
	s.Win, s.Wrec = win, wrec
	s.Cursor.SeenIDBits = append([]uint64(nil), s.Cursor.SeenIDBits...)
	recent := make([]OutcomeReceipt, len(s.Cursor.Recent))
	for i, r := range s.Cursor.Recent {
		r.Metrics = cloneOutcomeMetrics(r.Metrics)
		recent[i] = r
	}
	s.Cursor.Recent = recent
	return s
}
