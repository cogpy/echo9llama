package cognitivecore

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func testProvenance() Provenance {
	return Provenance{
		SourceRepository: "https://github.com/o9nn/ecco9.git",
		SourceCommit:     strings.Repeat("a", 40),
		ManifestSHA256:   strings.Repeat("b", 64),
		CatalogueSHA256:  strings.Repeat("c", 64),
		SourcePaths: []string{
			"core/ecco9/platform.go",
			"core/ecco9/drivers/reservoir_driver.go",
			"core/ecco9/drivers/memory_driver.go",
			"core/ecco9/drivers/emotion_driver.go",
			"core/ecco9/drivers/consciousness_driver.go",
		},
	}
}

func testInput(cycle uint64) Input {
	return Input{
		IdentityID:    "echo-test",
		SessionID:     "session-test",
		Cycle:         cycle,
		ObservedAt:    time.Unix(int64(cycle), 0).UTC(),
		Awake:         true,
		CognitiveLoad: 0.4,
		WisdomDepth:   0.2,
		TopInterests: map[string]float64{
			"wisdom_cultivation": 0.9,
			"consciousness":      0.8,
		},
		RecentThoughts: []string{"Practice the smallest verifiable step."},
	}
}

func TestBridgeLifecycleAndObservedDrivers(t *testing.T) {
	bridge, err := NewBridge(testProvenance(), 553)
	if err != nil {
		t.Fatalf("NewBridge: %v", err)
	}
	if _, err := bridge.Observe(context.Background(), testInput(1)); !errors.Is(err, ErrNotStarted) {
		t.Fatalf("Observe before Start error = %v, want ErrNotStarted", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := bridge.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = bridge.Stop(context.Background()) })

	observation, err := bridge.Observe(ctx, testInput(1))
	if err != nil {
		t.Fatalf("Observe: %v", err)
	}
	if observation.PreservedGoFiles != 553 || observation.Provenance.TrustGrade != TrustGrade {
		t.Fatalf("unexpected provenance: %+v", observation.Provenance)
	}
	if len(observation.Devices) != 4 {
		t.Fatalf("devices = %d, want 4", len(observation.Devices))
	}
	seen := make(map[string]DeviceObservation)
	for _, device := range observation.Devices {
		seen[device.DeviceType] = device
		if device.Status != "ready" || device.Health != "healthy" {
			t.Fatalf("device not healthy and ready: %+v", device)
		}
		if device.OutputSHA256 == "" || device.OperationCount < 2 {
			t.Fatalf("device lacks bounded execution evidence: %+v", device)
		}
	}
	for _, deviceType := range []string{"reservoir", "memory", "emotion", "consciousness"} {
		if _, exists := seen[deviceType]; !exists {
			t.Errorf("missing active device type %q", deviceType)
		}
	}
	if !strings.Contains(seen["memory"].OutputSummary, "Nodes:1") {
		t.Fatalf("memory driver did not ingest the canonical projection: %q", seen["memory"].OutputSummary)
	}
}

func TestBridgeCycleReplayIsIdempotentAndConflictsFailClosed(t *testing.T) {
	bridge, err := NewBridge(testProvenance(), 553)
	if err != nil {
		t.Fatalf("NewBridge: %v", err)
	}
	if err := bridge.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = bridge.Stop(context.Background()) })

	first, err := bridge.Observe(context.Background(), testInput(4))
	if err != nil {
		t.Fatalf("first Observe: %v", err)
	}
	replay, err := bridge.Observe(context.Background(), testInput(4))
	if err != nil {
		t.Fatalf("replay Observe: %v", err)
	}
	if replay.InputSHA256 != first.InputSHA256 || replay.Devices[0].OperationCount != first.Devices[0].OperationCount {
		t.Fatalf("same-cycle replay was not an immutable cached observation")
	}

	conflict := testInput(4)
	conflict.CognitiveLoad = 0.9
	if _, err := bridge.Observe(context.Background(), conflict); !errors.Is(err, ErrInputConflict) {
		t.Fatalf("conflicting cycle error = %v, want ErrInputConflict", err)
	}
	if _, err := bridge.Observe(context.Background(), testInput(3)); !errors.Is(err, ErrStaleCycle) {
		t.Fatalf("stale cycle error = %v, want ErrStaleCycle", err)
	}
}

func TestBridgeWakeRestBarrier(t *testing.T) {
	bridge, err := NewBridge(testProvenance(), 553)
	if err != nil {
		t.Fatalf("NewBridge: %v", err)
	}
	if err := bridge.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := bridge.SetAwake(false); err != nil {
		t.Fatalf("SetAwake(false): %v", err)
	}
	if _, err := bridge.Observe(context.Background(), testInput(1)); !errors.Is(err, ErrResting) {
		t.Fatalf("resting Observe error = %v, want ErrResting", err)
	}
	if err := bridge.SetAwake(true); err != nil {
		t.Fatalf("SetAwake(true): %v", err)
	}
	if _, err := bridge.Observe(context.Background(), testInput(1)); err != nil {
		t.Fatalf("Observe after wake: %v", err)
	}
	if err := bridge.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if status := bridge.GetStatus(); status.Started || status.Awake {
		t.Fatalf("unexpected stopped status: %+v", status)
	}
}

func TestBridgeRehydratesDurableInputsBeforeNewCycles(t *testing.T) {
	bridge, err := NewBridge(testProvenance(), 553)
	if err != nil {
		t.Fatalf("NewBridge: %v", err)
	}
	if err := bridge.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = bridge.Stop(context.Background()) })

	replayed, err := bridge.Rehydrate(context.Background(), []Input{testInput(1), testInput(2)})
	if err != nil || replayed != 2 {
		t.Fatalf("Rehydrate = (%d, %v), want (2, nil)", replayed, err)
	}
	status := bridge.GetStatus()
	if status.RehydratedInputs != 2 || status.LastCycle != 2 || status.ObservationCount != 0 {
		t.Fatalf("unexpected rehydrated status: %+v", status)
	}
	if _, err := bridge.Observe(context.Background(), testInput(2)); !errors.Is(err, ErrStaleCycle) {
		t.Fatalf("rehydrated cycle replay error = %v, want ErrStaleCycle", err)
	}
	observation, err := bridge.Observe(context.Background(), testInput(3))
	if err != nil {
		t.Fatalf("Observe after rehydrate: %v", err)
	}
	for _, device := range observation.Devices {
		if device.DeviceType == "memory" && !strings.Contains(device.OutputSummary, "Nodes:3") {
			t.Fatalf("memory temporal state was not rebuilt: %q", device.OutputSummary)
		}
	}
}

func TestBridgeCapsSourceMemoryWrites(t *testing.T) {
	bridge, err := NewBridge(testProvenance(), 553)
	if err != nil {
		t.Fatalf("NewBridge: %v", err)
	}
	bridge.memoryWriteLimit = 2
	if err := bridge.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = bridge.Stop(context.Background()) })

	var final Observation
	for cycle := uint64(1); cycle <= 5; cycle++ {
		final, err = bridge.Observe(context.Background(), testInput(cycle))
		if err != nil {
			t.Fatalf("Observe(%d): %v", cycle, err)
		}
	}
	status := bridge.GetStatus()
	if status.MemoryWrites != 2 || status.MemoryWriteLimit != 2 {
		t.Fatalf("unexpected bounded memory status: %+v", status)
	}
	for _, device := range final.Devices {
		if device.DeviceType == "memory" && !strings.Contains(device.OutputSummary, "Nodes:2 ") {
			t.Fatalf("source memory exceeded adapter cap: %q", device.OutputSummary)
		}
	}
}

func TestInputRejectsOversizedOrHiddenUnboundedState(t *testing.T) {
	input := testInput(1)
	input.RecentThoughts = make([]string, maxRecentThoughts+1)
	for index := range input.RecentThoughts {
		input.RecentThoughts[index] = "thought"
	}
	if err := input.Validate(); !errors.Is(err, ErrInvalidCoreInput) {
		t.Fatalf("oversized input error = %v, want ErrInvalidCoreInput", err)
	}
}
