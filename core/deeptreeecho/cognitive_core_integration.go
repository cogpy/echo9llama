package deeptreeecho

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/cogpy/echo9llama/core/cognitivecore"
	"github.com/cogpy/echo9llama/core/persistence"
)

const (
	cognitiveCorePolicyVersion   = "ecco9-cognitive-core-adapter-v1"
	maxCognitiveCoreReplayInputs = 4096
)

type cognitiveCoreEvidence struct {
	Input       cognitivecore.Input       `json:"input"`
	Observation cognitivecore.Observation `json:"observation"`
}

func (uao *UnifiedAutonomousOrchestrator) initializeCognitiveCore() {
	if !uao.config.EnableCognitiveCore {
		return
	}
	bridge, err := cognitivecore.NewDefaultBridge()
	if err != nil {
		uao.cognitiveCoreInitErr = err
		fmt.Printf("⚠️  ecco9 cognitive core unavailable: %v\n", err)
		return
	}
	uao.cognitiveCore = bridge
	fmt.Printf("   ✓ ecco9 Cognitive Core bound (%d preserved Go files, %s)\n", cognitivecore.PreservedGoFileCount, cognitivecore.TrustGrade)
}

// rehydrateCognitiveCore reconstructs bounded source-driver temporal state from
// the latest durable canonical inputs before waking cycles resume.
func (uao *UnifiedAutonomousOrchestrator) rehydrateCognitiveCore(ctx context.Context) error {
	uao.mu.RLock()
	bridge := uao.cognitiveCore
	store := uao.eventStore
	identityID := uao.identityID
	uao.mu.RUnlock()
	if bridge == nil || store == nil {
		return nil
	}

	inputs := make([]cognitivecore.Input, 0, maxCognitiveCoreReplayInputs)
	afterSequence := int64(0)
	for {
		events, err := store.Query(ctx, persistence.CognitiveEventQuery{
			AfterSequence: afterSequence,
			Limit:         1000,
			IdentityID:    identityID,
			EventType:     persistence.EventTypeCoreObserved,
		})
		if err != nil {
			return fmt.Errorf("query cognitive core replay: %w", err)
		}
		if len(events) == 0 {
			break
		}
		for _, event := range events {
			evidence, err := decodeCognitiveCoreEvidence(event.PayloadJSON)
			if err != nil {
				return fmt.Errorf("decode cognitive core replay event %s: %w", event.EventID, err)
			}
			inputs = append(inputs, evidence.Input)
			if len(inputs) > maxCognitiveCoreReplayInputs {
				copy(inputs, inputs[len(inputs)-maxCognitiveCoreReplayInputs:])
				inputs = inputs[:maxCognitiveCoreReplayInputs]
			}
			afterSequence = event.Sequence
		}
		if len(events) < 1000 {
			break
		}
	}
	if len(inputs) == 0 {
		return nil
	}
	replayed, err := bridge.Rehydrate(ctx, inputs)
	if err != nil {
		return err
	}
	status := bridge.GetStatus()
	uao.mu.Lock()
	if status.LastSessionID == uao.sessionID && status.LastCycle > uao.totalCycles {
		uao.totalCycles = status.LastCycle
	}
	uao.mu.Unlock()
	fmt.Printf("   ✓ ecco9 Cognitive Core rehydrated from %d durable inputs\n", replayed)
	return nil
}

// observeCognitiveCore sends only bounded public state through the reviewed
// ecco9 devices. The result is appended to the canonical event spine before it
// becomes EchoDream input.
func (uao *UnifiedAutonomousOrchestrator) observeCognitiveCore(cycle uint64) {
	now := time.Now().UTC()
	uao.mu.Lock()
	bridge := uao.cognitiveCore
	awake := uao.isAwake
	identityID := uao.identityID
	sessionID := uao.sessionID
	cognitiveLoad := uao.cognitiveLoad
	wisdomDepth := uao.wisdomDepth
	if bridge != nil && awake && cycle != 0 && !uao.lastCognitiveCoreObservation.IsZero() && now.Sub(uao.lastCognitiveCoreObservation) < uao.config.CognitiveCoreInterval {
		uao.mu.Unlock()
		return
	}
	if bridge != nil && awake && cycle != 0 {
		uao.lastCognitiveCoreObservation = now
	}
	uao.mu.Unlock()
	if bridge == nil || !awake || cycle == 0 {
		return
	}

	interests := make(map[string]float64)
	if uao.interestPatterns != nil {
		for _, interest := range uao.interestPatterns.GetTopInterests(8) {
			interests[interest.Topic] = interest.Strength
		}
	}
	thoughts := make([]string, 0, 5)
	if uao.streamOfConsciousness != nil {
		for _, thought := range uao.streamOfConsciousness.GetRecentThoughts(5) {
			content := strings.TrimSpace(thought.Content)
			if content != "" {
				thoughts = append(thoughts, content)
			}
		}
	}

	input := cognitivecore.Input{
		IdentityID: identityID, SessionID: sessionID, Cycle: cycle,
		ObservedAt: now, Awake: true,
		CognitiveLoad: cognitiveLoad, WisdomDepth: wisdomDepth,
		TopInterests: interests, RecentThoughts: thoughts,
	}
	observation, err := bridge.Observe(uao.ctx, input)
	if err != nil {
		if !errors.Is(err, cognitivecore.ErrResting) && !errors.Is(err, context.Canceled) {
			fmt.Printf("⚠️  ecco9 cognitive core observation failed closed: %v\n", err)
		}
		return
	}

	eventID, err := uao.recordCognitiveCoreObservation(input, observation)
	if err != nil {
		fmt.Printf("⚠️  ecco9 cognitive core evidence append failed: %v\n", err)
		return
	}
	if eventID == "" {
		eventID = fmt.Sprintf("cognitive-core-volatile:%s:%d", cognitiveCoreSessionKey(sessionID), cycle)
	}
	uao.ingestDreamExperienceOnce(
		eventID,
		cognitiveCoreSummary(observation),
		cognitiveCoreImportance(observation),
		[]string{"cognitive_core", "adapted_observed", "reservoir", "memory", "emotion", "consciousness"},
	)
}

func (uao *UnifiedAutonomousOrchestrator) recordCognitiveCoreObservation(input cognitivecore.Input, observation cognitivecore.Observation) (string, error) {
	uao.mu.RLock()
	store := uao.eventStore
	identityID := uao.identityID
	sessionID := uao.sessionID
	uao.mu.RUnlock()
	if store == nil {
		return "", nil
	}
	payload, err := json.Marshal(cognitiveCoreEvidence{Input: input, Observation: observation})
	if err != nil {
		return "", fmt.Errorf("marshal cognitive core observation: %w", err)
	}
	sessionKey := cognitiveCoreSessionKey(sessionID)
	eventID := fmt.Sprintf("cognitive-core:%s:cycle-%d", sessionKey, observation.Cycle)
	event := persistence.NewCognitiveEvent(
		eventID,
		persistence.EventTypeCoreObserved,
		identityID,
		sessionID,
		cognitiveCorePolicyVersion,
		payload,
	)
	event.CorrelationID = fmt.Sprintf("cognitive-cycle:%s:%d", sessionKey, observation.Cycle)
	event.IdempotencyKey = eventID
	event.EvidenceClass = persistence.EvidenceClassSensitive
	if _, err := store.Append(uao.ctx, event); err != nil {
		return "", err
	}
	return eventID, nil
}

func cognitiveCoreSessionKey(sessionID string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(sessionID)))
}

func decodeCognitiveCoreEvidence(payload []byte) (cognitiveCoreEvidence, error) {
	var evidence cognitiveCoreEvidence
	if err := json.Unmarshal(payload, &evidence); err != nil {
		return evidence, err
	}
	if err := evidence.Input.Validate(); err != nil {
		return evidence, err
	}
	expectedProvenance := cognitivecore.DefaultProvenance()
	actualProvenance := evidence.Observation.Provenance
	if evidence.Input.Cycle != evidence.Observation.Cycle ||
		evidence.Observation.PreservedGoFiles != cognitivecore.PreservedGoFileCount ||
		actualProvenance.SourceRepository != expectedProvenance.SourceRepository ||
		actualProvenance.SourceCommit != expectedProvenance.SourceCommit ||
		actualProvenance.ManifestSHA256 != expectedProvenance.ManifestSHA256 ||
		actualProvenance.CatalogueSHA256 != expectedProvenance.CatalogueSHA256 ||
		actualProvenance.AdapterVersion != expectedProvenance.AdapterVersion ||
		actualProvenance.TrustGrade != expectedProvenance.TrustGrade ||
		!slices.Equal(actualProvenance.SourcePaths, expectedProvenance.SourcePaths) {
		return evidence, fmt.Errorf("cognitive core evidence provenance mismatch")
	}
	encodedInput, err := json.Marshal(evidence.Input)
	if err != nil {
		return evidence, err
	}
	inputDigest := fmt.Sprintf("%x", sha256.Sum256(encodedInput))
	if inputDigest != evidence.Observation.InputSHA256 {
		return evidence, fmt.Errorf("cognitive core evidence input digest mismatch")
	}
	return evidence, nil
}

func cognitiveCoreSummary(observation cognitivecore.Observation) string {
	devices := make([]string, 0, len(observation.Devices))
	for _, device := range observation.Devices {
		summary := strings.TrimSpace(device.OutputSummary)
		if len(summary) > 160 {
			summary = summary[:160] + "…"
		}
		devices = append(devices, fmt.Sprintf("%s=%s/%s [%s]", device.DeviceType, device.Status, device.Health, summary))
	}
	sort.Strings(devices)
	return fmt.Sprintf(
		"ecco9 cognitive-core cycle %d observed %s under trust grade %s; input evidence SHA-256 %s.",
		observation.Cycle,
		strings.Join(devices, ", "),
		observation.Provenance.TrustGrade,
		observation.InputSHA256,
	)
}

func cognitiveCoreImportance(observation cognitivecore.Observation) float64 {
	if len(observation.Devices) == 0 {
		return 0.3
	}
	healthy := 0
	for _, device := range observation.Devices {
		if device.Status == "ready" && device.Health == "healthy" {
			healthy++
		}
	}
	return 0.35 + 0.25*float64(healthy)/float64(len(observation.Devices))
}
