package drivers

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/EchoCog/echollama/core/ecco9"
)

// MMIO register offsets for the NPU coprocessor (PERIPH space base 0x40001000).
const (
	regBase          uint64 = 0x40001000
	regCmd           uint64 = regBase + 0x00
	regStatus        uint64 = regBase + 0x04
	regPromptAddr    uint64 = regBase + 0x08
	regPromptLen     uint64 = regBase + 0x0C
	regNPredict      uint64 = regBase + 0x10
	regTokenOut      uint64 = regBase + 0x14
	regTokenReady    uint64 = regBase + 0x18
	regModelID       uint64 = regBase + 0x1C
	regCtxUsed       uint64 = regBase + 0x20
	regErrorCode     uint64 = regBase + 0x24
	regPerfTokensSec uint64 = regBase + 0x28
)

// Command bits written to regCmd.
const (
	cmdReset     uint32 = 1 << 0
	cmdLoadModel uint32 = 1 << 1
	cmdStartInf  uint32 = 1 << 2
	cmdSoftStop  uint32 = 1 << 3
)

// Status bits read from regStatus.
const (
	statusIdle       uint32 = 1 << 0
	statusBusy       uint32 = 1 << 1
	statusEOG        uint32 = 1 << 2
	statusError      uint32 = 1 << 3
	statusModelReady uint32 = 1 << 4
	statusTokenReady uint32 = 1 << 5
)

// ioctlNPUBase is the base for all NPU IoCtl command codes ('N','P' = 0x4E50).
const ioctlNPUBase uint32 = 0x4E500000

// IoCtl command codes for the NPU device.
const (
	IoctlNPULoadModel    uint32 = ioctlNPUBase | 0x0001
	IoctlNPURunInfer     uint32 = ioctlNPUBase | 0x0002
	IoctlNPUSoftStop     uint32 = ioctlNPUBase | 0x0003
	IoctlNPUReset        uint32 = ioctlNPUBase | 0x0004
	IoctlNPUGetTelemetry uint32 = ioctlNPUBase | 0x0005
	IoctlNPUSelfAssess   uint32 = ioctlNPUBase | 0x0006
)

// NPUModelConfig holds GGUF model and runtime parameters.
type NPUModelConfig struct {
	ModelPath      string
	ModelName      string
	NCtx           int32
	NThreads       int32
	NGPULayers     int32
	BatchSize      int32
	OffloadKVCache bool
	LowVRAMMode    bool
}

// DefaultNPUModelConfig returns sensible defaults for the NPU model.
func DefaultNPUModelConfig() *NPUModelConfig {
	return &NPUModelConfig{
		ModelPath:  "",
		ModelName:  "ecco9-npu",
		NCtx:       4096,
		NThreads:   4,
		NGPULayers: 0,
		BatchSize:  1,
	}
}

// NPUSequenceConfig controls a single inference sequence.
type NPUSequenceConfig struct {
	NPredict     int32
	MaxCtx       int32
	EchoPrompt   bool
	StreamTokens bool
	SystemPrompt string
}

// DefaultNPUSequenceConfig returns default sequence parameters.
func DefaultNPUSequenceConfig() NPUSequenceConfig {
	return NPUSequenceConfig{
		NPredict:     128,
		MaxCtx:       4096,
		StreamTokens: true,
	}
}

// NPUTelemetry holds real-time performance data exposed via hardware registers.
type NPUTelemetry struct {
	TokensPerSecond      float64
	TotalTokensGenerated uint64
	TotalPrompts         uint64
	LastPromptTokens     uint64
	LastCompletionTokens uint64
	LastInferenceStart   time.Time
	LastInferenceEnd     time.Time
}

// TokenCallback is called for each token during streaming inference.
type TokenCallback func(tokenText string, tokenID int32, isLast bool)

// --- Entelechy self-assessment structures ---

// OntologicalHealth scores the structural completeness of the NPU.
type OntologicalHealth struct {
	FoundationIntegrity    float64
	CoreCompleteness       float64
	SpecializedFeatures    float64
	ArchitecturalCoherence float64
}

// TeleologicalAlignment scores purpose clarity and development progress.
type TeleologicalAlignment struct {
	PhaseCompletion         [5]float64
	RoadmapAlignment        float64
	ActualizationTrajectory float64
	PurposeClarity          float64
}

// CognitiveCompleteness scores inference and meta-cognitive capability.
type CognitiveCompleteness struct {
	InferenceQuality        float64
	PerformanceIntelligence float64
	MetaCognitiveDepth      float64
	OverallCognition        float64
}

// IntegrativeHealth scores component coherence and inter-driver harmony.
type IntegrativeHealth struct {
	HardwareIntegration float64
	SoftwareCoherence   float64
	SystemUnity         float64
	OverallIntegration  float64
}

// EvolutionaryPotential scores the NPU's capacity for growth.
type EvolutionaryPotential struct {
	ImplementationDepth     float64
	SelfImprovementCapacity float64
	EvolutionaryFitness     float64
}

// NPUSelfAssessment is the full entelechy report produced by assessSelf.
type NPUSelfAssessment struct {
	Ontological                OntologicalHealth
	Teleological               TeleologicalAlignment
	Cognitive                  CognitiveCompleteness
	Integrative                IntegrativeHealth
	Evolutionary               EvolutionaryPotential
	OverallActualization       float64
	FitnessScore               float64
	ImprovementRecommendations []string
}

// --- MMIO register bank (in-process simulation) ---

type registerBank struct {
	mu   sync.RWMutex
	regs map[uint64]uint32
}

func newRegisterBank() *registerBank {
	rb := &registerBank{regs: make(map[uint64]uint32)}
	rb.regs[regStatus] = statusIdle
	return rb
}

func (rb *registerBank) write(addr uint64, val uint32) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.regs[addr] = val
}

func (rb *registerBank) read(addr uint64) uint32 {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.regs[addr]
}

func (rb *registerBank) setBits(addr uint64, bits uint32) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.regs[addr] |= bits
}

func (rb *registerBank) clearBits(addr uint64, bits uint32) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.regs[addr] &^= bits
}

// ===== NPUDriver (implements ecco9.Driver) =====

// NPUDriver is the device driver for the GGUF-backed LLM Neural Processing Unit.
type NPUDriver struct {
	mu      sync.RWMutex
	name    string
	version string
	devices map[string]*NPUDevice
	config  *NPUModelConfig
}

// NewNPUDriver creates a new NPU driver with default configuration.
func NewNPUDriver() *NPUDriver {
	return &NPUDriver{
		name:    "npu",
		version: "1.0.0",
		devices: make(map[string]*NPUDevice),
		config:  DefaultNPUModelConfig(),
	}
}

// Load implements Driver.Load.
func (nd *NPUDriver) Load(config interface{}) error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	if cfg, ok := config.(*NPUModelConfig); ok {
		nd.config = cfg
	}

	device := NewNPUDevice("npu0", nd.config)
	nd.devices["npu0"] = device
	return nil
}

// Unload implements Driver.Unload.
func (nd *NPUDriver) Unload() error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	for _, device := range nd.devices {
		ctx := context.Background()
		if err := device.Shutdown(ctx); err != nil {
			return err
		}
	}
	nd.devices = make(map[string]*NPUDevice)
	return nil
}

// GetDevice implements Driver.GetDevice.
func (nd *NPUDriver) GetDevice(id string) (ecco9.CognitiveDevice, error) {
	nd.mu.RLock()
	defer nd.mu.RUnlock()

	device, exists := nd.devices[id]
	if !exists {
		return nil, fmt.Errorf("npu: device %s not found", id)
	}
	return device, nil
}

// ListDevices implements Driver.ListDevices.
func (nd *NPUDriver) ListDevices() []ecco9.CognitiveDevice {
	nd.mu.RLock()
	defer nd.mu.RUnlock()

	out := make([]ecco9.CognitiveDevice, 0, len(nd.devices))
	for _, d := range nd.devices {
		out = append(out, d)
	}
	return out
}

// GetName implements Driver.GetName.
func (nd *NPUDriver) GetName() string { return nd.name }

// GetVersion implements Driver.GetVersion.
func (nd *NPUDriver) GetVersion() string { return nd.version }

// GetCapabilities implements Driver.GetCapabilities.
func (nd *NPUDriver) GetCapabilities() []string {
	return []string{
		"gguf-inference",
		"mmio-register-interface",
		"token-streaming",
		"hardware-telemetry",
		"entelechy-self-assessment",
		"ontogenetic-evolution",
		"multi-level-api",
	}
}

// ===== NPUDevice (implements ecco9.CognitiveDevice) =====

// NPUDevice is the GGUF-backed LLM coprocessor exposed as a memory-mapped device.
type NPUDevice struct {
	mu          sync.RWMutex
	id          string
	name        string
	state       ecco9.DeviceState
	modelConfig *NPUModelConfig
	regs        *registerBank
	telemetry   NPUTelemetry
	modelLoaded bool
	metrics     ecco9.DeviceMetrics
	startTime   time.Time
}

// NewNPUDevice constructs a new NPU device instance.
func NewNPUDevice(id string, cfg *NPUModelConfig) *NPUDevice {
	if cfg == nil {
		cfg = DefaultNPUModelConfig()
	}
	return &NPUDevice{
		id:          id,
		name:        fmt.Sprintf("NPU Coprocessor %s", id),
		modelConfig: cfg,
		regs:        newRegisterBank(),
		state: ecco9.DeviceState{
			ID:     id,
			Name:   fmt.Sprintf("NPU %s", id),
			Status: ecco9.DeviceStatusOffline,
			Power:  ecco9.PowerStateOff,
			Health: ecco9.HealthStatusHealthy,
		},
	}
}

// Initialize implements CognitiveDevice.Initialize.
func (nd *NPUDevice) Initialize(ctx context.Context) error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	nd.state.Status = ecco9.DeviceStatusInitializing
	nd.state.Power = ecco9.PowerStateActive

	// Hardware reset
	nd.regs.write(regCmd, cmdReset)
	nd.regs.write(regStatus, statusIdle)
	nd.regs.write(regErrorCode, 0)

	nd.startTime = time.Now()
	nd.state.Status = ecco9.DeviceStatusReady
	nd.state.LastUpdate = time.Now()

	return nil
}

// Shutdown implements CognitiveDevice.Shutdown.
func (nd *NPUDevice) Shutdown(ctx context.Context) error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	nd.regs.write(regCmd, cmdSoftStop)
	nd.regs.write(regStatus, statusIdle)
	nd.modelLoaded = false

	nd.state.Status = ecco9.DeviceStatusOffline
	nd.state.Power = ecco9.PowerStateOff
	return nil
}

// Reset implements CognitiveDevice.Reset.
func (nd *NPUDevice) Reset(ctx context.Context) error {
	if err := nd.Shutdown(ctx); err != nil {
		return err
	}
	return nd.Initialize(ctx)
}

// GetState implements CognitiveDevice.GetState.
func (nd *NPUDevice) GetState() (ecco9.DeviceState, error) {
	nd.mu.RLock()
	defer nd.mu.RUnlock()

	s := nd.state
	s.Uptime = time.Since(nd.startTime)
	s.Metrics = nd.metrics
	return s, nil
}

// SetState implements CognitiveDevice.SetState.
func (nd *NPUDevice) SetState(state ecco9.DeviceState) error {
	nd.mu.Lock()
	defer nd.mu.Unlock()
	nd.state = state
	return nil
}

// Read implements CognitiveDevice.Read.
// Returns a human-readable hardware status string.
func (nd *NPUDevice) Read(buffer []byte) (int, error) {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	status := nd.deviceStatusString()
	n := copy(buffer, []byte(status))

	nd.metrics.OperationCount++
	nd.metrics.LastOperation = time.Now()
	return n, nil
}

// Write implements CognitiveDevice.Write.
// Treats the buffer as a prompt and runs stubbed inference, writing output
// back into the device's token stream registers.
func (nd *NPUDevice) Write(buffer []byte) (int, error) {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	if nd.state.Status == ecco9.DeviceStatusOffline {
		return 0, fmt.Errorf("npu: device offline")
	}

	start := time.Now()

	nd.regs.clearBits(regStatus, statusIdle|statusEOG)
	nd.regs.setBits(regStatus, statusBusy)
	nd.regs.write(regPromptLen, uint32(len(buffer)))

	// Stubbed inference — replace with llama.cpp integration.
	nd.regs.clearBits(regStatus, statusBusy)
	nd.regs.setBits(regStatus, statusEOG|statusIdle)

	nd.telemetry.TotalPrompts++
	// NOTE: LastPromptTokens stores byte length as a proxy until a real tokenizer is integrated.
	nd.telemetry.LastPromptTokens = uint64(len(buffer))
	nd.telemetry.LastInferenceStart = start
	nd.telemetry.LastInferenceEnd = time.Now()

	elapsed := nd.telemetry.LastInferenceEnd.Sub(start).Seconds()
	if elapsed > 0 {
		nd.telemetry.TokensPerSecond = float64(nd.telemetry.LastCompletionTokens) / elapsed
	}
	nd.regs.write(regPerfTokensSec, uint32(nd.telemetry.TokensPerSecond))

	nd.metrics.OperationCount++
	nd.metrics.LastOperation = time.Now()
	latency := time.Since(start)
	// Cumulative moving average for latency across all operations.
	n := time.Duration(nd.metrics.OperationCount)
	nd.metrics.AverageLatency = (nd.metrics.AverageLatency*(n-1) + latency) / n

	return len(buffer), nil
}

// IoCtl implements CognitiveDevice.IoCtl.
func (nd *NPUDevice) IoCtl(command uint32, arg interface{}) error {
	switch command {
	case IoctlNPUReset:
		ctx := context.Background()
		return nd.Reset(ctx)

	case IoctlNPULoadModel:
		cfg, ok := arg.(*NPUModelConfig)
		if !ok {
			return fmt.Errorf("npu: IoctlNPULoadModel requires *NPUModelConfig argument")
		}
		return nd.loadModel(cfg)

	case IoctlNPURunInfer:
		req, ok := arg.(*NPUInferRequest)
		if !ok {
			return fmt.Errorf("npu: IoctlNPURunInfer requires *NPUInferRequest argument")
		}
		result, err := nd.infer(req.Prompt, req.SeqConfig)
		if err != nil {
			return err
		}
		if req.ResultPtr != nil {
			*req.ResultPtr = result
		}
		return nil

	case IoctlNPUSoftStop:
		nd.regs.write(regCmd, cmdSoftStop)
		nd.regs.clearBits(regStatus, statusBusy)
		nd.regs.setBits(regStatus, statusIdle)
		return nil

	case IoctlNPUGetTelemetry:
		ptr, ok := arg.(*NPUTelemetry)
		if !ok {
			return fmt.Errorf("npu: IoctlNPUGetTelemetry requires *NPUTelemetry argument")
		}
		nd.mu.RLock()
		*ptr = nd.telemetry
		nd.mu.RUnlock()
		return nil

	case IoctlNPUSelfAssess:
		ptr, ok := arg.(*NPUSelfAssessment)
		if !ok {
			return fmt.Errorf("npu: IoctlNPUSelfAssess requires *NPUSelfAssessment argument")
		}
		*ptr = nd.assessSelf()
		return nil

	default:
		return fmt.Errorf("npu: unknown ioctl command 0x%08X", command)
	}
}

// GetMetrics implements CognitiveDevice.GetMetrics.
func (nd *NPUDevice) GetMetrics() (ecco9.DeviceMetrics, error) {
	nd.mu.RLock()
	defer nd.mu.RUnlock()
	return nd.metrics, nil
}

// GetHealth implements CognitiveDevice.GetHealth.
func (nd *NPUDevice) GetHealth() (ecco9.HealthStatus, error) {
	nd.mu.RLock()
	defer nd.mu.RUnlock()
	return nd.state.Health, nil
}

// GetID implements CognitiveDevice.GetID.
func (nd *NPUDevice) GetID() string { return nd.id }

// GetName implements CognitiveDevice.GetName.
func (nd *NPUDevice) GetName() string { return nd.name }

// GetType implements CognitiveDevice.GetType.
func (nd *NPUDevice) GetType() ecco9.DeviceType { return ecco9.DeviceTypeNPU }

// ===== Hardware-level helpers =====

// loadModel simulates loading a GGUF model into the coprocessor.
func (nd *NPUDevice) loadModel(cfg *NPUModelConfig) error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	nd.regs.write(regCmd, cmdLoadModel)
	nd.regs.clearBits(regStatus, statusModelReady|statusError)
	nd.regs.write(regErrorCode, 0)

	if cfg != nil {
		nd.modelConfig = cfg
	}

	// Stub: real implementation would call llama_model_load here.
	nd.modelLoaded = true
	nd.regs.setBits(regStatus, statusModelReady)
	return nil
}

// infer runs a stubbed inference pass and returns generated text.
func (nd *NPUDevice) infer(prompt string, seq NPUSequenceConfig) (string, error) {
	nd.regs.clearBits(regStatus, statusIdle|statusEOG|statusError)
	nd.regs.setBits(regStatus, statusBusy)
	nd.regs.write(regPromptLen, uint32(len(prompt)))
	nd.regs.write(regNPredict, uint32(seq.NPredict))

	start := time.Now()

	// Stub completion — wire llama.cpp here for real inference.
	var sb strings.Builder
	sb.WriteString("[NPU-COPROC STUB] n_predict=")
	sb.WriteString(fmt.Sprintf("%d model=%s\n", seq.NPredict, nd.modelConfig.ModelName))
	sb.WriteString("Completion: (stubbed — connect GGUF runtime here)\n")
	result := sb.String()

	nd.regs.clearBits(regStatus, statusBusy)
	nd.regs.setBits(regStatus, statusEOG|statusIdle)

	nd.telemetry.TotalPrompts++
	// TODO: replace byte-length proxies with real token counts once GGUF tokenizer is integrated.
	nd.telemetry.LastPromptTokens = uint64(len(prompt))
	nd.telemetry.LastCompletionTokens = uint64(len(result))
	nd.telemetry.TotalTokensGenerated += nd.telemetry.LastCompletionTokens
	nd.telemetry.LastInferenceStart = start
	nd.telemetry.LastInferenceEnd = time.Now()

	elapsed := nd.telemetry.LastInferenceEnd.Sub(start).Seconds()
	if elapsed > 0 {
		nd.telemetry.TokensPerSecond = float64(nd.telemetry.LastCompletionTokens) / elapsed
	}
	nd.regs.write(regPerfTokensSec, uint32(nd.telemetry.TokensPerSecond))

	return result, nil
}

// Infer is the high-level convenience API for fire-and-forget inference.
func (nd *NPUDevice) Infer(prompt string, seq NPUSequenceConfig) (string, error) {
	nd.mu.Lock()
	defer nd.mu.Unlock()
	return nd.infer(prompt, seq)
}

// InferStreaming runs inference and calls cb for each token.
// In the stubbed implementation cb is called once with the full result.
func (nd *NPUDevice) InferStreaming(prompt string, seq NPUSequenceConfig, cb TokenCallback) error {
	result, err := nd.Infer(prompt, seq)
	if err != nil {
		return err
	}
	if cb != nil {
		cb(result, 0, true)
	}
	return nil
}

// ReadStatus returns the current hardware status register value.
func (nd *NPUDevice) ReadStatus() uint32 {
	return nd.regs.read(regStatus)
}

// IsBusy reports whether the NPU is currently performing inference.
func (nd *NPUDevice) IsBusy() bool {
	return nd.regs.read(regStatus)&statusBusy != 0
}

// HasError reports whether the NPU has a hardware error condition.
func (nd *NPUDevice) HasError() bool {
	return nd.regs.read(regStatus)&statusError != 0
}

// GetErrorCode returns the current hardware error code register.
func (nd *NPUDevice) GetErrorCode() uint32 {
	return nd.regs.read(regErrorCode)
}

// GetTelemetry returns a snapshot of NPU performance telemetry.
func (nd *NPUDevice) GetTelemetry() NPUTelemetry {
	nd.mu.RLock()
	defer nd.mu.RUnlock()
	return nd.telemetry
}

// deviceStatusString returns a human-readable hardware status string.
func (nd *NPUDevice) deviceStatusString() string {
	s := nd.regs.read(regStatus)
	var parts []string
	if s&statusIdle != 0 {
		parts = append(parts, "IDLE")
	}
	if s&statusBusy != 0 {
		parts = append(parts, "BUSY")
	}
	if s&statusModelReady != 0 {
		parts = append(parts, "MODEL_READY")
	}
	if s&statusEOG != 0 {
		parts = append(parts, "EOG")
	}
	if s&statusError != 0 {
		parts = append(parts, "ERROR")
	}
	if len(parts) == 0 {
		parts = append(parts, "UNKNOWN")
	}
	return fmt.Sprintf("NPU[%s] status=0x%08X [%s] tokens/s=%.1f total_tokens=%d",
		nd.id, s, strings.Join(parts, "|"),
		nd.telemetry.TokensPerSecond, nd.telemetry.TotalTokensGenerated)
}

// ===== Entelechy self-assessment =====

// Entelechy dimension weights (must sum to 1.0).
// Teleological and cognitive are weighted highest as they drive actualization most directly.
const (
	entelechyWeightOntological  = 0.20
	entelechyWeightTeleological = 0.25
	entelechyWeightCognitive    = 0.25
	entelechyWeightIntegrative  = 0.15
	entelechyWeightEvolutionary = 0.15
)

// assessSelf computes the five-dimensional entelechy report for the NPU.
func (nd *NPUDevice) assessSelf() NPUSelfAssessment {
	nd.mu.RLock()
	defer nd.mu.RUnlock()

	ont := nd.assessOntological()
	tel := nd.assessTeleological()
	cog := nd.assessCognitive()
	intg := nd.assessIntegrative()
	evo := nd.assessEvolutionary()

	overall := ont.ArchitecturalCoherence*entelechyWeightOntological +
		tel.RoadmapAlignment*entelechyWeightTeleological +
		cog.OverallCognition*entelechyWeightCognitive +
		intg.OverallIntegration*entelechyWeightIntegrative +
		evo.EvolutionaryFitness*entelechyWeightEvolutionary

	fitness := ont.ArchitecturalCoherence*entelechyWeightOntological +
		tel.ActualizationTrajectory*entelechyWeightTeleological +
		cog.OverallCognition*entelechyWeightCognitive +
		intg.OverallIntegration*entelechyWeightIntegrative +
		evo.EvolutionaryFitness*entelechyWeightEvolutionary

	recommendations := nd.generateRecommendations(ont, tel, cog, intg, evo)

	return NPUSelfAssessment{
		Ontological:                ont,
		Teleological:               tel,
		Cognitive:                  cog,
		Integrative:                intg,
		Evolutionary:               evo,
		OverallActualization:       overall,
		FitnessScore:               fitness,
		ImprovementRecommendations: recommendations,
	}
}

func (nd *NPUDevice) assessOntological() OntologicalHealth {
	foundation := 0.9  // MMIO register bank, config, device lifecycle — all present
	core := 0.7        // Stubbed inference; GGUF runtime not yet wired
	specialized := 0.5 // KV-cache, GPU offload, hot-swap planned but not implemented
	coherence := (foundation + core + specialized) / 3.0
	return OntologicalHealth{
		FoundationIntegrity:    foundation,
		CoreCompleteness:       core,
		SpecializedFeatures:    specialized,
		ArchitecturalCoherence: coherence,
	}
}

func (nd *NPUDevice) assessTeleological() TeleologicalAlignment {
	phases := [5]float64{1.0, 1.0, 0.6, 0.2, 0.0} // phases 1-2 done, 3 partial, 4-5 future
	avg := (phases[0] + phases[1] + phases[2] + phases[3] + phases[4]) / 5.0
	return TeleologicalAlignment{
		PhaseCompletion:         phases,
		RoadmapAlignment:        avg,
		ActualizationTrajectory: avg,
		PurposeClarity:          0.95,
	}
}

func (nd *NPUDevice) assessCognitive() CognitiveCompleteness {
	inferQ := 0.6   // Stubbed; real quality measurable once GGUF wired
	perfInt := 0.85 // Telemetry registers fully functional
	meta := 0.75    // Self-diagnostics and self-assessment implemented
	overall := (inferQ + perfInt + meta) / 3.0
	return CognitiveCompleteness{
		InferenceQuality:        inferQ,
		PerformanceIntelligence: perfInt,
		MetaCognitiveDepth:      meta,
		OverallCognition:        overall,
	}
}

func (nd *NPUDevice) assessIntegrative() IntegrativeHealth {
	hw := 0.9  // MMIO registers + CognitiveDevice interface fully wired
	sw := 0.85 // Driver interface, IoCtl, streaming API consistent
	sys := 0.8 // Coexists with other drivers; interrupt/DMA not yet implemented
	overall := (hw + sw + sys) / 3.0
	return IntegrativeHealth{
		HardwareIntegration: hw,
		SoftwareCoherence:   sw,
		SystemUnity:         sys,
		OverallIntegration:  overall,
	}
}

func (nd *NPUDevice) assessEvolutionary() EvolutionaryPotential {
	depth := 0.6    // Core stub present; production GGUF path not wired
	capacity := 0.8 // Clear extension points; ontogenesis framework ready
	fitness := (depth + capacity) / 2.0
	return EvolutionaryPotential{
		ImplementationDepth:     depth,
		SelfImprovementCapacity: capacity,
		EvolutionaryFitness:     fitness,
	}
}

func (nd *NPUDevice) generateRecommendations(
	ont OntologicalHealth, tel TeleologicalAlignment,
	cog CognitiveCompleteness, intg IntegrativeHealth,
	evo EvolutionaryPotential,
) []string {
	var recs []string
	if ont.CoreCompleteness < 0.9 {
		recs = append(recs, "Wire llama.cpp GGUF runtime into infer() to replace stub")
	}
	if ont.SpecializedFeatures < 0.8 {
		recs = append(recs, "Implement KV-cache management and GPU offload control")
	}
	if tel.PhaseCompletion[2] < 1.0 {
		recs = append(recs, "Complete Phase 3: token streaming FIFO and interrupt support")
	}
	if cog.InferenceQuality < 0.8 {
		recs = append(recs, "Add tokenization/detokenization pipeline for accurate quality metrics")
	}
	if evo.ImplementationDepth < 0.8 {
		recs = append(recs, "Replace stubbed paths with production implementations to raise depth score")
	}
	return recs
}

// ===== Helper types =====

// NPUInferRequest is passed to IoCtl(IoctlNPURunInfer, req).
type NPUInferRequest struct {
	Prompt    string
	SeqConfig NPUSequenceConfig
	ResultPtr *string
}
