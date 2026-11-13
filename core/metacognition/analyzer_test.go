package metacognition

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func boolPtr(value bool) *bool { return &value }

func floatPtr(value float64) *float64 { return &value }

func TestAnalyzeExactBrierAndDigest(t *testing.T) {
	passed := true
	failed := false
	passPrediction := 0.8
	failPrediction := 0.25
	records := []Evidence{
		{EventID: "evaluation-two", Sequence: 2, Kind: kindEvaluation, Confidence: &failPrediction, Passed: &failed},
		{EventID: "evaluation-one", Sequence: 1, Kind: kindEvaluation, Confidence: &passPrediction, Passed: &passed},
		{EventID: "action-one", Sequence: 3, Kind: kindActionFailure},
		{EventID: "plan-one", Sequence: 4, Kind: kindPlanningFailure},
	}

	report, err := Analyze(records)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if report.Version != Version {
		t.Fatalf("Version = %q, want %q", report.Version, Version)
	}
	if got, want := report.Evaluated, 2; got != want {
		t.Errorf("Evaluated = %d, want %d", got, want)
	}
	if got, want := report.Passed, 1; got != want {
		t.Errorf("Passed = %d, want %d", got, want)
	}
	if got, want := report.Failed, 1; got != want {
		t.Errorf("Failed = %d, want %d", got, want)
	}
	if got, want := report.ActionFailures, 1; got != want {
		t.Errorf("ActionFailures = %d, want %d", got, want)
	}
	if got, want := report.PlanningFailures, 1; got != want {
		t.Errorf("PlanningFailures = %d, want %d", got, want)
	}
	if got, want := report.LastSequence, int64(4); got != want {
		t.Errorf("LastSequence = %d, want %d", got, want)
	}
	if report.BrierScore == nil {
		t.Fatal("BrierScore = nil")
	}
	// ((0.8 - 1)^2 + (0.25 - 0)^2) / 2 = 0.05125.
	if got, want := *report.BrierScore, 0.05125; math.Abs(got-want) > 1e-12 {
		t.Errorf("BrierScore = %.17g, want %.17g", got, want)
	}
	if report.SuccessRate == nil || *report.SuccessRate != 0.5 {
		t.Errorf("SuccessRate = %v, want 0.5", report.SuccessRate)
	}
	if got, want := report.CalibrationSamples, 2; got != want {
		t.Errorf("CalibrationSamples = %d, want %d", got, want)
	}
	if got, want := report.MissingPredictions, 0; got != want {
		t.Errorf("MissingPredictions = %d, want %d", got, want)
	}

	if got, want := report.Evidence[0].Sequence, int64(1); got != want {
		t.Fatalf("first sorted sequence = %d, want %d", got, want)
	}
	canonical, err := json.Marshal(report.Evidence)
	if err != nil {
		t.Fatalf("marshal canonical evidence: %v", err)
	}
	digest := sha256.Sum256(canonical)
	if got, want := report.SourceSHA256, hex.EncodeToString(digest[:]); got != want {
		t.Errorf("SourceSHA256 = %q, want %q", got, want)
	}
}

func TestAnalyzeMissingPredictionsAndEmptyEvidence(t *testing.T) {
	records := []Evidence{
		{EventID: "one", Sequence: 1, Kind: kindEvaluation, Passed: boolPtr(true)},
		{EventID: "two", Sequence: 2, Kind: kindEvaluation, Passed: boolPtr(false), Confidence: floatPtr(0.25)},
		{EventID: "three", Sequence: 3, Kind: kindEvaluation, Passed: boolPtr(true)},
		{EventID: "failure", Sequence: 4, Kind: kindActionFailure},
	}
	report, err := Analyze(records)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if got, want := report.MissingPredictions, 2; got != want {
		t.Errorf("MissingPredictions = %d, want %d", got, want)
	}
	if got, want := report.CalibrationSamples, 1; got != want {
		t.Errorf("CalibrationSamples = %d, want %d", got, want)
	}
	if report.BrierScore == nil || *report.BrierScore != 0.0625 {
		t.Errorf("BrierScore = %v, want 0.0625", report.BrierScore)
	}
	if report.SuccessRate == nil || *report.SuccessRate != 2.0/3.0 {
		t.Errorf("SuccessRate = %v, want %v", report.SuccessRate, 2.0/3.0)
	}

	empty, err := Analyze(nil)
	if err != nil {
		t.Fatalf("Analyze(nil) error = %v", err)
	}
	if empty.Evidence == nil {
		t.Error("empty Evidence = nil, want canonical non-nil empty slice")
	}
	if empty.Evaluated != 0 || empty.Passed != 0 || empty.Failed != 0 || empty.LastSequence != 0 {
		t.Errorf("empty report has non-zero aggregates: %+v", empty)
	}
	if empty.BrierScore != nil || empty.SuccessRate != nil {
		t.Errorf("empty report invented rates: BrierScore=%v SuccessRate=%v", empty.BrierScore, empty.SuccessRate)
	}
	canonical, err := json.Marshal([]Evidence{})
	if err != nil {
		t.Fatalf("marshal empty evidence: %v", err)
	}
	digest := sha256.Sum256(canonical)
	if got, want := empty.SourceSHA256, hex.EncodeToString(digest[:]); got != want {
		t.Errorf("empty SourceSHA256 = %q, want %q", got, want)
	}
}

func TestAnalyzeRejectsBoundedInvalidAndDuplicateEvidence(t *testing.T) {
	valid := Evidence{EventID: "valid", Sequence: 1, Kind: kindEvaluation, Passed: boolPtr(true)}
	tooMany := make([]Evidence, MaxEvidence+1)
	for i := range tooMany {
		tooMany[i] = Evidence{EventID: "record-" + string(rune('a'+i)), Sequence: int64(i + 1), Kind: kindActionFailure}
	}

	tests := []struct {
		name    string
		records []Evidence
	}{
		{name: "too many", records: tooMany},
		{name: "empty ID", records: []Evidence{{EventID: " \t", Sequence: 1, Kind: kindActionFailure}}},
		{name: "oversized ID", records: []Evidence{{EventID: strings.Repeat("a", maxEventIDBytes+1), Sequence: 1, Kind: kindActionFailure}}},
		{name: "non-positive sequence", records: []Evidence{{EventID: "one", Sequence: 0, Kind: kindActionFailure}}},
		{name: "invalid kind", records: []Evidence{{EventID: "one", Sequence: 1, Kind: "opinion"}}},
		{name: "evaluation needs outcome", records: []Evidence{{EventID: "one", Sequence: 1, Kind: kindEvaluation}}},
		{name: "failure does not take outcome", records: []Evidence{{EventID: "one", Sequence: 1, Kind: kindActionFailure, Passed: boolPtr(false)}}},
		{name: "failure does not take confidence", records: []Evidence{{EventID: "one", Sequence: 1, Kind: kindPlanningFailure, Confidence: floatPtr(0.5)}}},
		{name: "NaN confidence", records: []Evidence{{EventID: "one", Sequence: 1, Kind: kindEvaluation, Passed: boolPtr(true), Confidence: floatPtr(math.NaN())}}},
		{name: "infinite confidence", records: []Evidence{{EventID: "one", Sequence: 1, Kind: kindEvaluation, Passed: boolPtr(true), Confidence: floatPtr(math.Inf(1))}}},
		{name: "out-of-range confidence", records: []Evidence{{EventID: "one", Sequence: 1, Kind: kindEvaluation, Passed: boolPtr(true), Confidence: floatPtr(-0.01)}}},
		{name: "duplicate event ID", records: []Evidence{valid, {EventID: "valid", Sequence: 2, Kind: kindActionFailure}}},
		{name: "duplicate sequence", records: []Evidence{valid, {EventID: "another", Sequence: 1, Kind: kindActionFailure}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Analyze(test.records); err == nil {
				t.Fatal("Analyze() error = nil, want rejection")
			}
		})
	}
}

func TestAnalyzeSortsWithoutAliasingCaller(t *testing.T) {
	confidence := 0.2
	passed := true
	input := []Evidence{
		{EventID: "later", Sequence: 2, Kind: kindEvaluation, Confidence: &confidence, Passed: &passed},
		{EventID: "first", Sequence: 1, Kind: kindActionFailure},
	}
	report, err := Analyze(input)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if got, want := input[0].Sequence, int64(2); got != want {
		t.Errorf("Analyze mutated caller order: input[0].Sequence = %d, want %d", got, want)
	}
	if got, want := report.Evidence[0].EventID, "first"; got != want {
		t.Errorf("sorted first EventID = %q, want %q", got, want)
	}
	if report.Evidence[1].Confidence == input[0].Confidence || report.Evidence[1].Passed == input[0].Passed {
		t.Fatal("report pointer fields alias caller pointer fields")
	}

	input[0].EventID = "mutated"
	confidence = 0.9
	passed = false
	if got, want := report.Evidence[1].EventID, "later"; got != want {
		t.Errorf("snapshot EventID changed to %q, want %q", got, want)
	}
	if got, want := *report.Evidence[1].Confidence, 0.2; got != want {
		t.Errorf("snapshot Confidence changed to %v, want %v", got, want)
	}
	if got, want := *report.Evidence[1].Passed, true; got != want {
		t.Errorf("snapshot Passed changed to %v, want %v", got, want)
	}
}

func TestAnalyzeDigestIsDeterministicAcrossInputOrder(t *testing.T) {
	first := []Evidence{
		{EventID: "two", Sequence: 2, Kind: kindActionFailure},
		{EventID: "one", Sequence: 1, Kind: kindEvaluation, Passed: boolPtr(true), Confidence: floatPtr(0.9)},
	}
	second := []Evidence{cloneEvidence(first[1]), cloneEvidence(first[0])}

	a, err := Analyze(first)
	if err != nil {
		t.Fatalf("Analyze(first) error = %v", err)
	}
	b, err := Analyze(second)
	if err != nil {
		t.Fatalf("Analyze(second) error = %v", err)
	}
	if a.SourceSHA256 != b.SourceSHA256 {
		t.Errorf("digest differs for equal evidence in a different input order: %q != %q", a.SourceSHA256, b.SourceSHA256)
	}
	if !sameEvidence(a.Evidence, b.Evidence) {
		t.Errorf("sorted evidence differs: %#v != %#v", a.Evidence, b.Evidence)
	}
}

func TestParseReflectionStrictValidation(t *testing.T) {
	report, err := Analyze([]Evidence{
		{EventID: "evaluation-1", Sequence: 1, Kind: kindEvaluation, Passed: boolPtr(true)},
		{EventID: "action-2", Sequence: 2, Kind: kindActionFailure},
	})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	valid := `{"lesson":"  Verify the artifact before relying on it.  ","uncertainty":"  Coverage is narrow. ","experiment":"  Repeat with one independent fixture. ","source_event_ids":["evaluation-1","action-2"]}`
	reflection, err := ParseReflection(valid, report)
	if err != nil {
		t.Fatalf("ParseReflection(valid) error = %v", err)
	}
	if got, want := reflection.Lesson, "Verify the artifact before relying on it."; got != want {
		t.Errorf("Lesson = %q, want %q", got, want)
	}
	if got, want := reflection.Uncertainty, "Coverage is narrow."; got != want {
		t.Errorf("Uncertainty = %q, want %q", got, want)
	}
	if got, want := reflection.Experiment, "Repeat with one independent fixture."; got != want {
		t.Errorf("Experiment = %q, want %q", got, want)
	}

	tests := []struct {
		name string
		raw  string
	}{
		{name: "unknown field", raw: `{"lesson":"x","uncertainty":"y","experiment":"z","source_event_ids":["evaluation-1"],"extra":true}`},
		{name: "trailing JSON", raw: `{"lesson":"x","uncertainty":"y","experiment":"z","source_event_ids":["evaluation-1"]} {}`},
		{name: "markdown fence", raw: "```json\n" + valid + "\n```"},
		{name: "empty prose", raw: `{"lesson":" ","uncertainty":"y","experiment":"z","source_event_ids":["evaluation-1"]}`},
		{name: "unknown source", raw: `{"lesson":"x","uncertainty":"y","experiment":"z","source_event_ids":["not-in-report"]}`},
		{name: "duplicate source", raw: `{"lesson":"x","uncertainty":"y","experiment":"z","source_event_ids":["evaluation-1","evaluation-1"]}`},
		{name: "no sources", raw: `{"lesson":"x","uncertainty":"y","experiment":"z","source_event_ids":[]}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ParseReflection(test.raw, report); err == nil {
				t.Fatal("ParseReflection() error = nil, want rejection")
			}
		})
	}

	emptyReport, err := Analyze(nil)
	if err != nil {
		t.Fatalf("Analyze(nil) error = %v", err)
	}
	if _, err := ParseReflection(valid, emptyReport); err == nil {
		t.Fatal("ParseReflection(empty report) error = nil, want rejection")
	}
}

func sameEvidence(left, right []Evidence) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i].EventID != right[i].EventID || left[i].Sequence != right[i].Sequence || left[i].Kind != right[i].Kind {
			return false
		}
		if (left[i].Confidence == nil) != (right[i].Confidence == nil) || (left[i].Passed == nil) != (right[i].Passed == nil) {
			return false
		}
		if left[i].Confidence != nil && *left[i].Confidence != *right[i].Confidence {
			return false
		}
		if left[i].Passed != nil && *left[i].Passed != *right[i].Passed {
			return false
		}
	}
	return true
}
