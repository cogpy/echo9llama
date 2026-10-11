// Package metacognition computes bounded, deterministic analytics over validated
// evidence-ledger records. It does not generate lessons, predictions, or model
// responses; ParseReflection only validates a separately supplied public response.
package metacognition

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	// Version identifies the deterministic evidence analytics contract.
	Version = "evidence-metacognition-v1"

	// MaxEvidence bounds the number of ledger records accepted in one analysis.
	MaxEvidence = 64

	maxEventIDBytes        = 256
	maxReflectionTextBytes = 800
	maxReflectionSources   = 8
)

const (
	kindEvaluation      = "evaluation"
	kindActionFailure   = "action_failure"
	kindPlanningFailure = "planning_failure"
)

// Evidence is one validated, ledger-derived record. Only evaluations may include
// outcome or confidence values: Passed is required and Confidence is optional.
type Evidence struct {
	EventID    string   `json:"event_id"`
	Sequence   int64    `json:"sequence"`
	Kind       string   `json:"kind"`
	Confidence *float64 `json:"confidence,omitempty"`
	Passed     *bool    `json:"passed,omitempty"`
}

// Report is an immutable snapshot of deterministic analytics calculated from
// Evidence. Its optional rate fields remain nil when their denominator is zero.
type Report struct {
	Version            string     `json:"version"`
	Evidence           []Evidence `json:"evidence"`
	Evaluated          int        `json:"evaluated"`
	Passed             int        `json:"passed"`
	Failed             int        `json:"failed"`
	PlanningFailures   int        `json:"planning_failures"`
	ActionFailures     int        `json:"action_failures"`
	MissingPredictions int        `json:"missing_predictions"`
	CalibrationSamples int        `json:"calibration_samples"`
	BrierScore         *float64   `json:"brier_score,omitempty"`
	SuccessRate        *float64   `json:"success_rate,omitempty"`
	SourceSHA256       string     `json:"source_sha256"`
	LastSequence       int64      `json:"last_sequence"`
}

// Reflection is a model-supplied public statement that has passed structural and
// evidence-citation validation. Analyze never produces this value.
type Reflection struct {
	Lesson         string   `json:"lesson"`
	Uncertainty    string   `json:"uncertainty"`
	Experiment     string   `json:"experiment"`
	SourceEventIDs []string `json:"source_event_ids"`
}

// Analyze validates a bounded ledger slice and returns a sorted, deep-copied
// snapshot with evaluation-only success and calibration measurements.
//
// SourceSHA256 is exactly the lowercase hexadecimal SHA-256 digest of the bytes
// returned by json.Marshal on the validated, deep-copied Evidence slice after it
// has been sorted in ascending Sequence order. The marshalled slice (not Report,
// its metrics, or input order) is the complete canonical digest preimage.
func Analyze(records []Evidence) (Report, error) {
	if len(records) > MaxEvidence {
		return Report{}, fmt.Errorf("evidence count %d exceeds maximum %d", len(records), MaxEvidence)
	}

	evidence := make([]Evidence, len(records))
	seenEventIDs := make(map[string]struct{}, len(records))
	seenSequences := make(map[int64]struct{}, len(records))

	for i, record := range records {
		if err := validateEvidence(record); err != nil {
			return Report{}, fmt.Errorf("evidence[%d]: %w", i, err)
		}
		if _, exists := seenEventIDs[record.EventID]; exists {
			return Report{}, fmt.Errorf("evidence[%d]: duplicate event_id %q", i, record.EventID)
		}
		if _, exists := seenSequences[record.Sequence]; exists {
			return Report{}, fmt.Errorf("evidence[%d]: duplicate sequence %d", i, record.Sequence)
		}
		seenEventIDs[record.EventID] = struct{}{}
		seenSequences[record.Sequence] = struct{}{}
		evidence[i] = cloneEvidence(record)
	}

	sort.Slice(evidence, func(i, j int) bool {
		return evidence[i].Sequence < evidence[j].Sequence
	})

	canonical, err := json.Marshal(evidence)
	if err != nil {
		return Report{}, fmt.Errorf("marshal canonical evidence: %w", err)
	}
	digest := sha256.Sum256(canonical)
	report := Report{
		Version:      Version,
		Evidence:     evidence,
		SourceSHA256: hex.EncodeToString(digest[:]),
	}
	if len(evidence) > 0 {
		report.LastSequence = evidence[len(evidence)-1].Sequence
	}

	var brierTotal float64
	for _, record := range evidence {
		switch record.Kind {
		case kindPlanningFailure:
			report.PlanningFailures++
		case kindActionFailure:
			report.ActionFailures++
		case kindEvaluation:
			report.Evaluated++
			if *record.Passed {
				report.Passed++
			} else {
				report.Failed++
			}
			if record.Confidence == nil {
				report.MissingPredictions++
				continue
			}

			outcome := 0.0
			if *record.Passed {
				outcome = 1.0
			}
			delta := *record.Confidence - outcome
			brierTotal += delta * delta
			report.CalibrationSamples++
		}
	}

	if report.Evaluated > 0 {
		rate := float64(report.Passed) / float64(report.Evaluated)
		if !isFinite(rate) {
			return Report{}, fmt.Errorf("non-finite success rate")
		}
		report.SuccessRate = &rate
	}
	if report.CalibrationSamples > 0 {
		score := brierTotal / float64(report.CalibrationSamples)
		if !isFinite(score) {
			return Report{}, fmt.Errorf("non-finite Brier score")
		}
		report.BrierScore = &score
	}

	return report, nil
}

func validateEvidence(record Evidence) error {
	if !utf8.ValidString(record.EventID) {
		return fmt.Errorf("event_id must be valid UTF-8")
	}
	if strings.TrimSpace(record.EventID) == "" {
		return fmt.Errorf("event_id is required")
	}
	if len(record.EventID) > maxEventIDBytes {
		return fmt.Errorf("event_id exceeds %d bytes", maxEventIDBytes)
	}
	if record.Sequence <= 0 {
		return fmt.Errorf("sequence must be positive")
	}

	switch record.Kind {
	case kindEvaluation:
		if record.Passed == nil {
			return fmt.Errorf("evaluation passed is required")
		}
		if record.Confidence != nil && !isUnitInterval(*record.Confidence) {
			return fmt.Errorf("evaluation confidence must be finite and within [0,1]")
		}
	case kindActionFailure, kindPlanningFailure:
		if record.Passed != nil {
			return fmt.Errorf("%s must not include passed", record.Kind)
		}
		if record.Confidence != nil {
			return fmt.Errorf("%s must not include confidence", record.Kind)
		}
	default:
		return fmt.Errorf("unsupported kind %q", record.Kind)
	}
	return nil
}

func cloneEvidence(record Evidence) Evidence {
	clone := Evidence{
		EventID:  record.EventID,
		Sequence: record.Sequence,
		Kind:     record.Kind,
	}
	if record.Confidence != nil {
		confidence := *record.Confidence
		clone.Confidence = &confidence
	}
	if record.Passed != nil {
		passed := *record.Passed
		clone.Passed = &passed
	}
	return clone
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func isUnitInterval(value float64) bool {
	return isFinite(value) && value >= 0 && value <= 1
}

// ParseReflection accepts only one strict JSON object containing a public
// reflection. It never generates or amends reflection text; it trims and validates
// the supplied prose and verifies every cited event against report.Evidence.
func ParseReflection(raw string, report Report) (Reflection, error) {
	if len(report.Evidence) == 0 {
		return Reflection{}, fmt.Errorf("reflection requires non-empty report evidence")
	}
	if !utf8.ValidString(raw) {
		return Reflection{}, fmt.Errorf("reflection must be valid UTF-8")
	}
	if strings.Contains(raw, "```") {
		return Reflection{}, fmt.Errorf("reflection markdown fences are not allowed")
	}
	if err := ValidateJSON(raw); err != nil {
		return Reflection{}, err
	}

	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.DisallowUnknownFields()
	var reflection Reflection
	if err := decoder.Decode(&reflection); err != nil {
		return Reflection{}, fmt.Errorf("decode reflection: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Reflection{}, fmt.Errorf("decode reflection: trailing JSON value")
		}
		return Reflection{}, fmt.Errorf("decode reflection: trailing data: %w", err)
	}

	lesson, err := validateReflectionText("lesson", reflection.Lesson)
	if err != nil {
		return Reflection{}, err
	}
	uncertainty, err := validateReflectionText("uncertainty", reflection.Uncertainty)
	if err != nil {
		return Reflection{}, err
	}
	experiment, err := validateReflectionText("experiment", reflection.Experiment)
	if err != nil {
		return Reflection{}, err
	}

	if len(reflection.SourceEventIDs) == 0 || len(reflection.SourceEventIDs) > maxReflectionSources {
		return Reflection{}, fmt.Errorf("source_event_ids must contain 1 to %d IDs", maxReflectionSources)
	}
	knownEventIDs := make(map[string]struct{}, len(report.Evidence))
	for _, evidence := range report.Evidence {
		knownEventIDs[evidence.EventID] = struct{}{}
	}

	sources := make([]string, len(reflection.SourceEventIDs))
	seenSources := make(map[string]struct{}, len(reflection.SourceEventIDs))
	for i, eventID := range reflection.SourceEventIDs {
		if !utf8.ValidString(eventID) {
			return Reflection{}, fmt.Errorf("source_event_ids[%d] must be valid UTF-8", i)
		}
		if strings.TrimSpace(eventID) == "" {
			return Reflection{}, fmt.Errorf("source_event_ids[%d] is required", i)
		}
		if len(eventID) > maxReflectionTextBytes {
			return Reflection{}, fmt.Errorf("source_event_ids[%d] exceeds %d bytes", i, maxReflectionTextBytes)
		}
		if _, duplicate := seenSources[eventID]; duplicate {
			return Reflection{}, fmt.Errorf("source_event_ids[%d] duplicates %q", i, eventID)
		}
		if _, known := knownEventIDs[eventID]; !known {
			return Reflection{}, fmt.Errorf("source_event_ids[%d] %q is not in report evidence", i, eventID)
		}
		seenSources[eventID] = struct{}{}
		sources[i] = eventID
	}

	return Reflection{
		Lesson:         lesson,
		Uncertainty:    uncertainty,
		Experiment:     experiment,
		SourceEventIDs: sources,
	}, nil
}

func validateReflectionText(field, value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", fmt.Errorf("%s must be valid UTF-8", field)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	if len(value) > maxReflectionTextBytes {
		return "", fmt.Errorf("%s exceeds %d bytes", field, maxReflectionTextBytes)
	}
	return value, nil
}
