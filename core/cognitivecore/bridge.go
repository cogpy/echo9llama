// Package cognitivecore adapts reviewed o9nn/ecco9 cognitive hardware drivers
// into the canonical echo9llama runtime without granting the lineage module
// scheduling, persistence, provider-routing, or external-action authority.
package cognitivecore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	ecco9 "github.com/EchoCog/echollama/core/ecco9"
	"github.com/EchoCog/echollama/core/ecco9/drivers"
)

const (
	// AdapterVersion changes when the semantics of the canonical bridge change.
	AdapterVersion = "ecco9-cognitive-core-v1"
	// TrustGrade records that an observation was produced by executing reviewed
	// source drivers through this adapter. It is evidence, not verified wisdom.
	TrustGrade = "adapted_observed"

	maxInputBytes       = 64 * 1024
	maxRecentThoughts   = 8
	maxInterests        = 32
	maxCachedCycles     = 256
	maxMemoryWrites     = 4096
	deviceReadBufferLen = 4096
)

var (
	ErrNotStarted        = errors.New("cognitive core is not started")
	ErrResting           = errors.New("cognitive core is resting")
	ErrInputConflict     = errors.New("cognitive core cycle input conflict")
	ErrStaleCycle        = errors.New("cognitive core cycle is stale")
	ErrInvalidCoreInput  = errors.New("invalid cognitive core input")
	ErrInvalidProvenance = errors.New("invalid cognitive core provenance")
)

// Provenance binds every observation to the immutable imported source and the
// narrow adapter that activated it.
type Provenance struct {
	SourceRepository string   `json:"source_repository"`
	SourceCommit     string   `json:"source_commit"`
	ManifestSHA256   string   `json:"manifest_sha256"`
	CatalogueSHA256  string   `json:"catalogue_sha256"`
	SourcePaths      []string `json:"source_paths"`
	AdapterVersion   string   `json:"adapter_version"`
	TrustGrade       string   `json:"trust_grade"`
}

func (p Provenance) Validate() error {
	if strings.TrimSpace(p.SourceRepository) == "" || len(strings.TrimSpace(p.SourceCommit)) != 40 || len(strings.TrimSpace(p.ManifestSHA256)) != 64 || len(strings.TrimSpace(p.CatalogueSHA256)) != 64 {
		return ErrInvalidProvenance
	}
	if len(p.SourcePaths) == 0 {
		return ErrInvalidProvenance
	}
	return nil
}

// Input is a bounded public cognitive-state projection. It deliberately
// excludes hidden model reasoning and credentials.
type Input struct {
	IdentityID     string             `json:"identity_id"`
	SessionID      string             `json:"session_id"`
	Cycle          uint64             `json:"cycle"`
	ObservedAt     time.Time          `json:"observed_at"`
	Awake          bool               `json:"awake"`
	CognitiveLoad  float64            `json:"cognitive_load"`
	WisdomDepth    float64            `json:"wisdom_depth"`
	TopInterests   map[string]float64 `json:"top_interests,omitempty"`
	RecentThoughts []string           `json:"recent_thoughts,omitempty"`
}

// Validate enforces the bounded public-state schema used for live observation
// and durable replay.
func (input Input) Validate() error {
	if strings.TrimSpace(input.IdentityID) == "" || strings.TrimSpace(input.SessionID) == "" || input.Cycle == 0 || input.ObservedAt.IsZero() {
		return ErrInvalidCoreInput
	}
	if input.CognitiveLoad < 0 || input.CognitiveLoad > 1 || input.WisdomDepth < 0 {
		return ErrInvalidCoreInput
	}
	if len(input.TopInterests) > maxInterests || len(input.RecentThoughts) > maxRecentThoughts {
		return ErrInvalidCoreInput
	}
	for topic, strength := range input.TopInterests {
		if strings.TrimSpace(topic) == "" || strength < 0 || strength > 1 {
			return ErrInvalidCoreInput
		}
	}
	for _, thought := range input.RecentThoughts {
		if strings.TrimSpace(thought) == "" {
			return ErrInvalidCoreInput
		}
	}
	encoded, err := json.Marshal(input)
	if err != nil || len(encoded) > maxInputBytes {
		return ErrInvalidCoreInput
	}
	return nil
}

// DeviceObservation is the bounded public result from one source driver.
type DeviceObservation struct {
	DeviceID       string `json:"device_id"`
	DeviceType     string `json:"device_type"`
	Status         string `json:"status"`
	Health         string `json:"health"`
	OutputSummary  string `json:"output_summary"`
	OutputSHA256   string `json:"output_sha256"`
	OperationCount uint64 `json:"operation_count"`
	ErrorCount     uint64 `json:"error_count"`
}

// Observation is the typed, evidence-ready result of a cognitive-core cycle.
type Observation struct {
	Cycle            uint64              `json:"cycle"`
	ObservedAt       time.Time           `json:"observed_at"`
	InputSHA256      string              `json:"input_sha256"`
	PlatformStatus   map[string]string   `json:"platform_status"`
	Devices          []DeviceObservation `json:"devices"`
	ActiveDrivers    []string            `json:"active_drivers"`
	PreservedGoFiles int                 `json:"preserved_go_files"`
	Provenance       Provenance          `json:"provenance"`
}

// Status reports the bridge state without exposing cognitive payloads.
type Status struct {
	Started          bool       `json:"started"`
	Awake            bool       `json:"awake"`
	LastSessionID    string     `json:"last_session_id"`
	LastCycle        uint64     `json:"last_cycle"`
	ObservationCount uint64     `json:"observation_count"`
	RehydratedInputs uint64     `json:"rehydrated_inputs"`
	MemoryWrites     uint64     `json:"memory_writes"`
	MemoryWriteLimit uint64     `json:"memory_write_limit"`
	DeviceCount      int        `json:"device_count"`
	Provenance       Provenance `json:"provenance"`
}

// Bridge owns one source platform and its reviewed active device subset.
type Bridge struct {
	mu sync.Mutex

	platform         *ecco9.Platform
	devices          []ecco9.CognitiveDevice
	cancel           context.CancelFunc
	started          bool
	awake            bool
	lastSessionID    string
	lastCycle        uint64
	observations     uint64
	rehydratedInputs uint64
	memoryWrites     uint64
	memoryWriteLimit uint64
	cache            map[uint64]Observation
	cacheInput       map[uint64]string
	cacheOrder       []uint64
	provenance       Provenance
	preservedGoFiles int
}

// NewBridge constructs an inactive bridge. Device goroutines start only when
// Start succeeds.
func NewBridge(provenance Provenance, preservedGoFiles int) (*Bridge, error) {
	if err := provenance.Validate(); err != nil || preservedGoFiles <= 0 {
		return nil, ErrInvalidProvenance
	}
	provenance.SourcePaths = append([]string(nil), provenance.SourcePaths...)
	provenance.AdapterVersion = AdapterVersion
	provenance.TrustGrade = TrustGrade

	platform := ecco9.NewPlatform(ecco9.DefaultConfiguration())
	devices := []ecco9.CognitiveDevice{
		drivers.NewReservoirDevice("reservoir0", drivers.DefaultReservoirConfig()),
		drivers.NewMemoryDevice("memory0", drivers.DefaultMemoryConfig()),
		drivers.NewEmotionDevice("emotion0", drivers.DefaultEmotionConfig()),
		drivers.NewConsciousnessDevice("consciousness0", drivers.DefaultConsciousnessConfig()),
	}
	for _, device := range devices {
		if err := platform.RegisterDevice(device); err != nil {
			return nil, fmt.Errorf("register %s: %w", device.GetID(), err)
		}
	}

	return &Bridge{
		platform: platform, devices: devices, provenance: provenance,
		preservedGoFiles: preservedGoFiles,
		memoryWriteLimit: maxMemoryWrites,
		cache:            make(map[uint64]Observation), cacheInput: make(map[uint64]string),
	}, nil
}

// Start initializes the reviewed source drivers. The source platform's
// sleep-based demonstration boot sequence is intentionally not authoritative;
// the adapter records readiness only after every real device initializes.
func (bridge *Bridge) Start(parent context.Context) error {
	bridge.mu.Lock()
	defer bridge.mu.Unlock()
	if bridge.started {
		return fmt.Errorf("cognitive core already started")
	}
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	initialized := make([]ecco9.CognitiveDevice, 0, len(bridge.devices))
	for _, device := range bridge.devices {
		if err := device.Initialize(ctx); err != nil {
			cancel()
			for index := len(initialized) - 1; index >= 0; index-- {
				_ = initialized[index].Shutdown(context.Background())
			}
			return fmt.Errorf("initialize %s: %w", device.GetID(), err)
		}
		initialized = append(initialized, device)
	}
	bridge.cancel = cancel
	bridge.platform.BootTime = time.Now().UTC()
	bridge.platform.Firmware.BootStage = ecco9.BootStageReady
	bridge.started = true
	bridge.awake = true
	return nil
}

// SetAwake is the wake/rest authority boundary for source cognition.
func (bridge *Bridge) SetAwake(awake bool) error {
	bridge.mu.Lock()
	defer bridge.mu.Unlock()
	if !bridge.started {
		return ErrNotStarted
	}
	bridge.awake = awake
	return nil
}

// Rehydrate replays durable public inputs through the active devices in ledger
// order before normal cycles resume. It does not create fresh observations or
// mutate the per-process idempotency cache.
func (bridge *Bridge) Rehydrate(ctx context.Context, inputs []Input) (int, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	bridge.mu.Lock()
	defer bridge.mu.Unlock()
	if !bridge.started {
		return 0, ErrNotStarted
	}
	if !bridge.awake {
		return 0, ErrResting
	}

	replayed := 0
	for _, input := range inputs {
		if err := input.Validate(); err != nil {
			return replayed, err
		}
		select {
		case <-ctx.Done():
			return replayed, ctx.Err()
		default:
		}
		payload, err := json.Marshal(input)
		if err != nil {
			return replayed, fmt.Errorf("marshal rehydration input: %w", err)
		}
		for _, device := range bridge.devices {
			if err := bridge.writeDevice(device, payload); err != nil {
				return replayed, fmt.Errorf("rehydrate %s: %w", device.GetID(), err)
			}
		}
		if input.SessionID != bridge.lastSessionID {
			bridge.lastSessionID = input.SessionID
			bridge.lastCycle = 0
		}
		if input.Cycle > bridge.lastCycle {
			bridge.lastCycle = input.Cycle
		}
		replayed++
	}
	bridge.rehydratedInputs += uint64(replayed)
	return replayed, nil
}

// Observe executes one bounded state projection through each reviewed source
// driver. Replaying the same cycle and input is idempotent; conflicting or stale
// cycles fail closed.
func (bridge *Bridge) Observe(ctx context.Context, input Input) (Observation, error) {
	if err := input.Validate(); err != nil {
		return Observation{}, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-ctx.Done():
		return Observation{}, ctx.Err()
	default:
	}

	payload, err := json.Marshal(input)
	if err != nil {
		return Observation{}, fmt.Errorf("marshal cognitive core input: %w", err)
	}
	inputHash := digest(payload)

	bridge.mu.Lock()
	defer bridge.mu.Unlock()
	if !bridge.started {
		return Observation{}, ErrNotStarted
	}
	if !bridge.awake || !input.Awake {
		return Observation{}, ErrResting
	}
	if input.SessionID != bridge.lastSessionID {
		bridge.lastSessionID = input.SessionID
		bridge.lastCycle = 0
		bridge.cache = make(map[uint64]Observation)
		bridge.cacheInput = make(map[uint64]string)
		bridge.cacheOrder = nil
	}
	if cached, exists := bridge.cache[input.Cycle]; exists {
		if bridge.cacheInput[input.Cycle] != inputHash {
			return Observation{}, ErrInputConflict
		}
		return cloneObservation(cached), nil
	}
	if input.Cycle <= bridge.lastCycle {
		return Observation{}, ErrStaleCycle
	}

	observations := make([]DeviceObservation, 0, len(bridge.devices))
	for _, device := range bridge.devices {
		if err := bridge.writeDevice(device, payload); err != nil {
			return Observation{}, fmt.Errorf("write %s: %w", device.GetID(), err)
		}
		observed, err := observeDevice(device)
		if err != nil {
			return Observation{}, err
		}
		observations = append(observations, observed)
	}
	sort.Slice(observations, func(i, j int) bool { return observations[i].DeviceID < observations[j].DeviceID })

	observation := Observation{
		Cycle:       input.Cycle,
		ObservedAt:  input.ObservedAt.UTC(),
		InputSHA256: inputHash,
		PlatformStatus: map[string]string{
			"boot_stage":       bridge.platform.Firmware.BootStage.String(),
			"firmware_version": bridge.platform.Firmware.Version,
			"kernel_version":   bridge.platform.Firmware.KernelVersion,
		},
		Devices: observations,
		ActiveDrivers: []string{
			"consciousness", "emotion", "memory", "reservoir",
		},
		PreservedGoFiles: bridge.preservedGoFiles,
		Provenance:       bridge.provenance,
	}
	bridge.lastCycle = input.Cycle
	bridge.observations++
	bridge.cache[input.Cycle] = cloneObservation(observation)
	bridge.cacheInput[input.Cycle] = inputHash
	bridge.cacheOrder = append(bridge.cacheOrder, input.Cycle)
	if len(bridge.cacheOrder) > maxCachedCycles {
		overflow := len(bridge.cacheOrder) - maxCachedCycles
		for _, cycle := range bridge.cacheOrder[:overflow] {
			delete(bridge.cache, cycle)
			delete(bridge.cacheInput, cycle)
		}
		bridge.cacheOrder = append([]uint64(nil), bridge.cacheOrder[overflow:]...)
	}
	return cloneObservation(observation), nil
}

func observeDevice(device ecco9.CognitiveDevice) (DeviceObservation, error) {
	buffer := make([]byte, deviceReadBufferLen)
	count, err := device.Read(buffer)
	if err != nil {
		return DeviceObservation{}, fmt.Errorf("read %s: %w", device.GetID(), err)
	}
	if count < 0 || count > len(buffer) {
		return DeviceObservation{}, fmt.Errorf("read %s returned invalid byte count %d", device.GetID(), count)
	}
	state, err := device.GetState()
	if err != nil {
		return DeviceObservation{}, fmt.Errorf("state %s: %w", device.GetID(), err)
	}
	health, err := device.GetHealth()
	if err != nil {
		return DeviceObservation{}, fmt.Errorf("health %s: %w", device.GetID(), err)
	}
	metrics, err := device.GetMetrics()
	if err != nil {
		return DeviceObservation{}, fmt.Errorf("metrics %s: %w", device.GetID(), err)
	}
	output := strings.TrimSpace(string(buffer[:count]))
	return DeviceObservation{
		DeviceID: device.GetID(), DeviceType: string(device.GetType()),
		Status: string(state.Status), Health: string(health),
		OutputSummary: output, OutputSHA256: digest([]byte(output)),
		OperationCount: metrics.OperationCount, ErrorCount: metrics.ErrorCount,
	}, nil
}

// GetStatus returns an immutable status snapshot.
func (bridge *Bridge) GetStatus() Status {
	bridge.mu.Lock()
	defer bridge.mu.Unlock()
	status := Status{
		Started: bridge.started, Awake: bridge.awake, LastSessionID: bridge.lastSessionID,
		LastCycle:        bridge.lastCycle,
		ObservationCount: bridge.observations, RehydratedInputs: bridge.rehydratedInputs,
		MemoryWrites: bridge.memoryWrites, MemoryWriteLimit: bridge.memoryWriteLimit,
		DeviceCount: len(bridge.devices),
		Provenance:  bridge.provenance,
	}
	status.Provenance.SourcePaths = append([]string(nil), bridge.provenance.SourcePaths...)
	return status
}

// Stop cancels all source-driver loops and powers down every active device.
func (bridge *Bridge) Stop(ctx context.Context) error {
	bridge.mu.Lock()
	defer bridge.mu.Unlock()
	if !bridge.started {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if bridge.cancel != nil {
		bridge.cancel()
	}
	var errs []error
	for index := len(bridge.devices) - 1; index >= 0; index-- {
		if err := bridge.devices[index].Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown %s: %w", bridge.devices[index].GetID(), err))
		}
	}
	bridge.started = false
	bridge.awake = false
	return errors.Join(errs...)
}

func (bridge *Bridge) writeDevice(device ecco9.CognitiveDevice, payload []byte) error {
	devicePayload := payload
	if device.GetType() == ecco9.DeviceTypeMemory {
		if bridge.memoryWrites >= bridge.memoryWriteLimit {
			return nil
		}
		devicePayload = []byte(fmt.Sprintf(
			`{"input_sha256":%q,"ordinal":%d}`,
			digest(payload),
			bridge.memoryWrites+1,
		))
	}
	if _, err := device.Write(devicePayload); err != nil {
		return err
	}
	if device.GetType() == ecco9.DeviceTypeMemory {
		bridge.memoryWrites++
	}
	return nil
}

func digest(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func cloneObservation(observation Observation) Observation {
	observation.PlatformStatus = cloneStringMap(observation.PlatformStatus)
	observation.Devices = append([]DeviceObservation(nil), observation.Devices...)
	observation.ActiveDrivers = append([]string(nil), observation.ActiveDrivers...)
	observation.Provenance.SourcePaths = append([]string(nil), observation.Provenance.SourcePaths...)
	return observation
}

func cloneStringMap(input map[string]string) map[string]string {
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
