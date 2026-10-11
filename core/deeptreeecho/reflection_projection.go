package deeptreeecho

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/cogpy/echo9llama/core/metacognition"
	"github.com/cogpy/echo9llama/core/persistence"
)

// quarantineReflectionSource excludes structurally invalid but hash-valid source
// records from the metacognition projection. The original append-only record is
// retained. Hash corruption/store failures still stop replay; a rejected event
// is never converted into training evidence or a dream experience.
func (uao *UnifiedAutonomousOrchestrator) quarantineReflectionSource(ctx context.Context, source persistence.StoredCognitiveEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	id := "reflection-quarantine:" + sha256Hex([]byte(uao.identityID+":"+source.EventID))
	if _, err := uao.eventStore.Get(ctx, id); err == nil {
		return nil
	} else if !errors.Is(err, persistence.ErrCognitiveEventNotFound) {
		return err
	}
	payload, _ := json.Marshal(map[string]any{"stage": "reflection_source_validation", "category": "invalid_causal_or_payload_evidence", "source_event_id": source.EventID, "source_sequence": source.Sequence})
	event := persistence.NewCognitiveEvent(id, persistence.EventTypeErrorRecorded, uao.identityID, uao.sessionID, metacognition.Version, payload)
	event.IdempotencyKey = id
	event.CausationID = source.EventID
	event.EvidenceClass = persistence.EvidenceClassSensitive
	_, err := uao.eventStore.Append(ctx, event)
	return err
}

// Ordered ledger cursors make counters replay-stable without an unbounded set
// of seen event IDs. Independent high-watermarks avoid skipping quarantine
// records appended during a source page before a later reflection success.
func (uao *UnifiedAutonomousOrchestrator) countReflectionEvent(event persistence.StoredCognitiveEvent) {
	if event.IdentityID != uao.identityID || event.PolicyVersion != metacognition.Version {
		return
	}
	if event.EventType == persistence.EventTypeReflectionProposed {
		if event.Sequence <= uao.reflectionCounterSequence {
			return
		}
		var evidence reflectionEvidence
		if strictReflectionJSON(event.PayloadJSON, &evidence) != nil || evidence.Version != metacognition.Version || evidence.EpistemicStatus != "proposal" {
			return
		}
		uao.reflectionProposed++
		uao.reflectionCounterSequence = event.Sequence
		uao.lastReflectionDigest = evidence.Report.SourceSHA256
		return
	}
	if event.EventType != persistence.EventTypeErrorRecorded {
		return
	}
	if strings.HasPrefix(event.EventID, "reflection-quarantine:") {
		if event.Sequence > uao.reflectionQuarantineSequence {
			uao.reflectionQuarantined++
			uao.reflectionQuarantineSequence = event.Sequence
		}
		return
	}
	if !strings.HasPrefix(event.EventID, "reflection:") || !strings.HasSuffix(event.EventID, ":error") || event.Sequence <= uao.reflectionErrorSequence {
		return
	}
	var payload struct {
		Stage        string `json:"stage"`
		Category     string `json:"category"`
		SourceSHA256 string `json:"source_sha256"`
	}
	if strictReflectionJSON(event.PayloadJSON, &payload) != nil || payload.Stage != "evidence_reflection" || len(payload.SourceSHA256) != 64 {
		return
	}
	uao.reflectionErrors++
	uao.reflectionErrorSequence = event.Sequence
	uao.lastReflectionDigest = payload.SourceSHA256
}

func (uao *UnifiedAutonomousOrchestrator) hasOpenReflectionGoal() bool {
	if uao.echobeatsScheduler == nil {
		return false
	}
	for _, goal := range uao.echobeatsScheduler.GetGoalQueue() {
		if strings.HasPrefix(goal.ID, "reflection-goal:") && goal.Status != GoalCompleted && goal.Status != GoalAbandoned {
			return true
		}
	}
	return false
}

// This defense-in-depth classifier does not replace ActionPolicy. Trusted goal
// construction further restricts every admitted experiment to planning a private
// note, never carrying out the model's prose as commands or widening capability.
func validateReflectionExperiment(experiment string) error {
	value := strings.ToLower(experiment)
	for _, forbidden := range []string{"credential", "api key", "password", "secret token", "network", "http://", "https://", "shell", "sudo ", "execute command", "send message", "publish", "email ", "slack", "discord", "change policy", "disable policy", "i completed", "i executed", "already completed", "already executed"} {
		if strings.Contains(value, forbidden) {
			return errors.New("reflection experiment exceeds private-note proposal contract")
		}
	}
	return nil
}
