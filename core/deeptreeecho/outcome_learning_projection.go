package deeptreeecho

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cogpy/echo9llama/core/cognitivecore"
	"github.com/cogpy/echo9llama/core/persistence"
)

// The auxiliary learner belongs to the canonical spine, not the imported legacy
// driver clock. It cannot produce text, authorize actions, or tune the LLM.
func (uao *UnifiedAutonomousOrchestrator) initializeOutcomeLearning() {
	if uao.config.EnableCognitiveCore {
		uao.outcomeLearner = cognitivecore.NewOutcomeLearner()
	}
}

// replayOutcomeLearning processes ALL relevant events in ledger order rather
// than racing a core-observation callback against later action projection.
// Startup reconstructs from the full authoritative ledger. Checkpoint files are
// bounded exports, not trusted authority: an editable cache cannot invent labels.
func (uao *UnifiedAutonomousOrchestrator) replayOutcomeLearning(ctx context.Context) error {
	if uao.outcomeLearner == nil || uao.eventStore == nil {
		return nil
	}
	uao.learningMu.Lock()
	defer uao.learningMu.Unlock()
	for {
		events, err := uao.eventStore.Query(ctx, persistence.CognitiveEventQuery{AfterSequence: uao.learningCursor, IdentityID: uao.identityID, Limit: 1000})
		if err != nil {
			return err
		}
		if len(events) == 0 {
			return nil
		}
		for _, event := range events {
			var example *cognitivecore.LearningExample
			switch event.EventType {
			case persistence.EventTypeCoreObserved:
				evidence, err := decodeCognitiveCoreEvidence(event.PayloadJSON)
				if err != nil {
					uao.learningRejected++
					uao.learningCursor = event.Sequence
					continue
				}
				interests := make([]string, 0, len(evidence.Input.TopInterests))
				for topic := range evidence.Input.TopInterests {
					interests = append(interests, topic)
				}
				sort.Strings(interests)
				// Semantic content FIRST: no truncated JSON session/identity/cycle prefix.
				text := strings.Join(evidence.Input.RecentThoughts, "\n") + "\n" + strings.Join(interests, " ")
				example = &cognitivecore.LearningExample{Sequence: event.Sequence, EventID: event.EventID, Text: text, TaskKind: cognitivecore.OutcomeTaskGeneral}
			case persistence.EventTypePracticeEvaluated:
				outcome, plan, err := uao.validatedPractice(ctx, event)
				if err != nil {
					if !errors.Is(err, errRejectedPracticeEvidence) {
						return err
					}
					uao.learningRejected++
					uao.learningCursor = event.Sequence
					continue
				}
				// Content is the model's actual candidate, not evaluator prose or target.
				example = &cognitivecore.LearningExample{Sequence: event.Sequence, EventID: event.EventID, Text: plan.Content, TaskKind: cognitivecore.OutcomeTaskArithmeticArtifact, Numeric: cognitivecore.TaskNumericFeatures{A: float64(outcome.Task.A), B: float64(outcome.Task.B)}, Passed: outcome.Passed, HeldOut: practiceHeldOut(outcome.Task.TaskID)}
			}
			if example != nil {
				if _, err := uao.outcomeLearner.Apply(*example); err != nil {
					uao.learningRejected++
					fmt.Printf("Auxiliary learner rejected event %s: %v\n", event.EventID, err)
				}
			}
			uao.learningCursor = event.Sequence
		}
		if len(events) < 1000 {
			return nil
		}
	}
}

// Split depends exclusively on the immutable task ID, before its model answer.
func practiceHeldOut(taskID string) bool {
	digest := sha256.Sum256([]byte(taskID))
	return digest[0]%5 == 0
}

type outcomeCheckpointExport struct {
	Version            string                               `json:"version"`
	IdentityID         string                               `json:"identity_id"`
	ProjectionSequence int64                                `json:"projection_sequence"`
	RecoveryAuthority  string                               `json:"recovery_authority"`
	Learner            cognitivecore.OutcomeLearnerSnapshot `json:"learner"`
}

func (uao *UnifiedAutonomousOrchestrator) saveOutcomeCheckpoint() error {
	if uao.outcomeLearner == nil || !uao.config.EnablePersistence {
		return nil
	}
	uao.learningMu.Lock()
	defer uao.learningMu.Unlock()
	payload, err := json.Marshal(outcomeCheckpointExport{Version: "outcome-checkpoint-export-v1", IdentityID: uao.identityID, ProjectionSequence: uao.learningCursor, RecoveryAuthority: "full-validated-event-replay", Learner: uao.outcomeLearner.Snapshot()})
	if err != nil {
		return err
	}
	dir := uao.config.StateDirectory
	if strings.TrimSpace(dir) == "" {
		return fmt.Errorf("checkpoint state directory required")
	}
	f, err := os.CreateTemp(dir, ".outcome-checkpoint-*")
	if err != nil {
		return err
	}
	temp := f.Name()
	defer func() { _ = os.Remove(temp) }()
	if err := f.Chmod(0o600); err != nil {
		_ = f.Close()
		return err
	}
	if _, err := f.Write(append(payload, '\n')); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(temp, filepath.Join(dir, "auxiliary_learner_checkpoint.json")); err != nil {
		return err
	}
	directory, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func() { _ = directory.Close() }()
	return directory.Sync()
}

type OutcomeLearningStatus struct {
	Enabled             bool     `json:"enabled"`
	ModelVersion        string   `json:"model_version"`
	FeatureVersion      string   `json:"feature_version"`
	ReadoutVersion      uint64   `json:"readout_version"`
	Observed            uint64   `json:"observed"`
	Training            uint64   `json:"training"`
	HeldOut             uint64   `json:"held_out"`
	Unknown             uint64   `json:"unknown"`
	MeanPrequentialLoss *float64 `json:"mean_prequential_loss"`
	MeanHeldOutLoss     *float64 `json:"mean_held_out_loss"`
	Rejected            uint64   `json:"rejected"`
	ProjectionSequence  int64    `json:"projection_sequence"`
	DataSHA256          string   `json:"data_sha256"`
	RecoveryAuthority   string   `json:"recovery_authority"`
	Authority           string   `json:"authority"`
}

func (uao *UnifiedAutonomousOrchestrator) outcomeLearningStatus() OutcomeLearningStatus {
	status := OutcomeLearningStatus{Enabled: uao.outcomeLearner != nil, RecoveryAuthority: "full-validated-event-replay", Authority: "advisory-only; no text, policy or LLM-weight authority"}
	if uao.outcomeLearner == nil {
		return status
	}
	uao.learningMu.Lock()
	defer uao.learningMu.Unlock()
	snapshot := uao.outcomeLearner.Snapshot()
	status.ModelVersion = snapshot.ModelVersion
	status.FeatureVersion = snapshot.FeatureVersion
	status.ReadoutVersion = snapshot.ReadoutVersion
	status.Observed = snapshot.Counts.Observed
	status.Training = snapshot.Counts.Training
	status.HeldOut = snapshot.Counts.HeldOut
	status.Unknown = snapshot.Counts.Unknown
	if status.Training > 0 {
		loss := snapshot.Counts.PrequentialLossSum / float64(status.Training)
		status.MeanPrequentialLoss = &loss
	}
	if status.HeldOut > 0 {
		loss := snapshot.Counts.HeldOutLossSum / float64(status.HeldOut)
		status.MeanHeldOutLoss = &loss
	}
	status.Rejected = uao.learningRejected
	status.ProjectionSequence = uao.learningCursor
	status.DataSHA256 = snapshot.DataSHA256
	return status
}
