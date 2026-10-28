package drivers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/EchoCog/echollama/core/ecco9"
	"github.com/EchoCog/echollama/core/llm"
)

// LLMDriver implements the driver for LLM providers. Every completion it
// exposes comes from a real language model: an injected generator, the given
// provider manager, or a backend configured through the environment (see
// ModelBackendFromEnv). Without one, generation fails and no text is produced.
type LLMDriver struct {
	mu        sync.RWMutex
	name      string
	version   string
	devices   map[string]*LLMDevice
	generator TextGenerator
}

// NewLLMDriver creates a new LLM driver backed by a provider manager. A nil
// manager, or one that still has no registered providers at Load time, is
// replaced by the environment-configured backend when one exists.
func NewLLMDriver(manager *llm.ProviderManager) *LLMDriver {
	var gen TextGenerator
	if manager != nil {
		gen = manager
	}
	return NewLLMDriverWithGenerator(gen)
}

// NewLLMDriverWithGenerator creates an LLM driver driving the given model.
func NewLLMDriverWithGenerator(generator TextGenerator) *LLMDriver {
	return &LLMDriver{
		name:      "llm",
		version:   "1.0.0",
		devices:   make(map[string]*LLMDevice),
		generator: generator,
	}
}

// Load implements Driver.Load. config may be a TextGenerator to inject a
// model; otherwise the driver's generator or the environment backend is used.
func (ld *LLMDriver) Load(config interface{}) error {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	if gen, ok := config.(TextGenerator); ok && gen != nil {
		ld.generator = gen
	}
	if pm, ok := ld.generator.(*llm.ProviderManager); ok && len(pm.ListProviders()) == 0 {
		if gen, err := ModelBackendFromEnv(); err == nil {
			ld.generator = gen
		}
	}
	if ld.generator == nil {
		if gen, err := ModelBackendFromEnv(); err == nil {
			ld.generator = gen
		}
	}

	device := NewLLMDeviceWithGenerator("llm0", ld.generator)
	ld.devices["llm0"] = device

	return nil
}

// Unload implements Driver.Unload
func (ld *LLMDriver) Unload() error {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	// Shutdown all devices
	for _, device := range ld.devices {
		ctx := context.Background()
		if err := device.Shutdown(ctx); err != nil {
			return err
		}
	}

	ld.devices = make(map[string]*LLMDevice)
	return nil
}

// GetDevice implements Driver.GetDevice
func (ld *LLMDriver) GetDevice(id string) (ecco9.CognitiveDevice, error) {
	ld.mu.RLock()
	defer ld.mu.RUnlock()

	device, exists := ld.devices[id]
	if !exists {
		return nil, fmt.Errorf("device %s not found", id)
	}
	return device, nil
}

// ListDevices implements Driver.ListDevices
func (ld *LLMDriver) ListDevices() []ecco9.CognitiveDevice {
	ld.mu.RLock()
	defer ld.mu.RUnlock()

	devices := make([]ecco9.CognitiveDevice, 0, len(ld.devices))
	for _, device := range ld.devices {
		devices = append(devices, device)
	}
	return devices
}

// GetName implements Driver.GetName
func (ld *LLMDriver) GetName() string {
	return ld.name
}

// GetVersion implements Driver.GetVersion
func (ld *LLMDriver) GetVersion() string {
	return ld.version
}

// GetCapabilities implements Driver.GetCapabilities
func (ld *LLMDriver) GetCapabilities() []string {
	return []string{
		"text-generation",
		"streaming",
		"multi-provider",
		"fallback-routing",
	}
}

// LLMDevice represents an LLM provider device. Write submits a prompt to
// the model; Read returns the model's completion.
type LLMDevice struct {
	mu        sync.RWMutex
	id        string
	name      string
	state     ecco9.DeviceState
	generator TextGenerator
	output    []byte
	metrics   ecco9.DeviceMetrics
	startTime time.Time
}

// NewLLMDevice creates a new LLM device backed by a provider manager.
func NewLLMDevice(id string, manager *llm.ProviderManager) *LLMDevice {
	var gen TextGenerator
	if manager != nil {
		gen = manager
	}
	return NewLLMDeviceWithGenerator(id, gen)
}

// NewLLMDeviceWithGenerator creates a new LLM device driving generator.
func NewLLMDeviceWithGenerator(id string, generator TextGenerator) *LLMDevice {
	return &LLMDevice{
		id:        id,
		name:      fmt.Sprintf("LLM Provider %s", id),
		generator: generator,
		state: ecco9.DeviceState{
			ID:     id,
			Name:   fmt.Sprintf("LLM %s", id),
			Status: ecco9.DeviceStatusOffline,
			Power:  ecco9.PowerStateOff,
			Health: ecco9.HealthStatusHealthy,
		},
		metrics: ecco9.DeviceMetrics{},
	}
}

// Initialize implements CognitiveDevice.Initialize
func (ld *LLMDevice) Initialize(ctx context.Context) error {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	ld.state.Status = ecco9.DeviceStatusInitializing
	ld.state.Power = ecco9.PowerStateActive

	if ld.generator == nil {
		ld.state.Status = ecco9.DeviceStatusError
		ld.state.Health = ecco9.HealthStatusFailed
		return ErrNoModelBackend
	}

	ld.startTime = time.Now()
	ld.state.Status = ecco9.DeviceStatusReady
	ld.state.LastUpdate = time.Now()

	return nil
}

// Shutdown implements CognitiveDevice.Shutdown
func (ld *LLMDevice) Shutdown(ctx context.Context) error {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	ld.state.Status = ecco9.DeviceStatusOffline
	ld.state.Power = ecco9.PowerStateOff

	return nil
}

// Reset implements CognitiveDevice.Reset
func (ld *LLMDevice) Reset(ctx context.Context) error {
	if err := ld.Shutdown(ctx); err != nil {
		return err
	}
	return ld.Initialize(ctx)
}

// GetState implements CognitiveDevice.GetState
func (ld *LLMDevice) GetState() (ecco9.DeviceState, error) {
	ld.mu.RLock()
	defer ld.mu.RUnlock()

	state := ld.state
	state.Uptime = time.Since(ld.startTime)
	state.Metrics = ld.metrics

	return state, nil
}

// SetState implements CognitiveDevice.SetState
func (ld *LLMDevice) SetState(state ecco9.DeviceState) error {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	ld.state = state
	return nil
}

// Read implements CognitiveDevice.Read. It drains the completion produced
// by the model for the last Write; with no pending completion it returns
// io.EOF and no bytes.
func (ld *LLMDevice) Read(buffer []byte) (int, error) {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	if len(ld.output) == 0 {
		return 0, io.EOF
	}
	n := copy(buffer, ld.output)
	ld.output = ld.output[n:]

	ld.metrics.OperationCount++
	ld.metrics.LastOperation = time.Now()

	return n, nil
}

// Write implements CognitiveDevice.Write. The buffer is sent to the model as
// a prompt and its completion becomes readable through Read. On failure the
// error is returned and no output is made available.
func (ld *LLMDevice) Write(buffer []byte) (int, error) {
	if _, err := ld.Generate(context.Background(), string(buffer), llm.DefaultGenerateOptions()); err != nil {
		return 0, err
	}
	return len(buffer), nil
}

// Generate queries the model with prompt and returns its completion, which
// also becomes readable through Read.
func (ld *LLMDevice) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	ld.mu.Lock()
	gen := ld.generator
	ld.output = nil
	ld.mu.Unlock()

	if gen == nil {
		ld.recordError()
		return "", ErrNoModelBackend
	}

	startTime := time.Now()
	out, err := generateText(ctx, gen, prompt, opts)
	if err != nil {
		ld.recordError()
		return "", fmt.Errorf("llm device %s: %w", ld.id, err)
	}

	ld.mu.Lock()
	defer ld.mu.Unlock()
	ld.output = []byte(out)
	ld.metrics.OperationCount++
	ld.metrics.LastOperation = time.Now()
	latency := time.Since(startTime)
	n := time.Duration(ld.metrics.OperationCount)
	ld.metrics.AverageLatency = (ld.metrics.AverageLatency*(n-1) + latency) / n

	return out, nil
}

func (ld *LLMDevice) recordError() {
	ld.mu.Lock()
	defer ld.mu.Unlock()
	ld.metrics.ErrorCount++
}

// SetGenerator replaces the model backend driven by this device.
func (ld *LLMDevice) SetGenerator(generator TextGenerator) error {
	if generator == nil {
		return errors.New("llm device: nil generator")
	}
	ld.mu.Lock()
	defer ld.mu.Unlock()
	ld.generator = generator
	return nil
}

// IoCtl implements CognitiveDevice.IoCtl
func (ld *LLMDevice) IoCtl(command uint32, arg interface{}) error {
	// Implement device-specific control operations
	return nil
}

// GetMetrics implements CognitiveDevice.GetMetrics
func (ld *LLMDevice) GetMetrics() (ecco9.DeviceMetrics, error) {
	ld.mu.RLock()
	defer ld.mu.RUnlock()

	return ld.metrics, nil
}

// GetHealth implements CognitiveDevice.GetHealth
func (ld *LLMDevice) GetHealth() (ecco9.HealthStatus, error) {
	ld.mu.RLock()
	defer ld.mu.RUnlock()

	return ld.state.Health, nil
}

// GetID implements CognitiveDevice.GetID
func (ld *LLMDevice) GetID() string {
	return ld.id
}

// GetName implements CognitiveDevice.GetName
func (ld *LLMDevice) GetName() string {
	return ld.name
}

// GetType implements CognitiveDevice.GetType
func (ld *LLMDevice) GetType() ecco9.DeviceType {
	return ecco9.DeviceTypeLLM
}
