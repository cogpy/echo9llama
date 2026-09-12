package drivers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EchoCog/echollama/core/ecco9"
	"github.com/EchoCog/echollama/core/llm"
)

// LLMDriver implements the driver for LLM providers
type LLMDriver struct {
	mu      sync.RWMutex
	name    string
	version string
	devices map[string]*LLMDevice
	manager *llm.ProviderManager
}

// NewLLMDriver creates a new LLM driver
func NewLLMDriver(manager *llm.ProviderManager) *LLMDriver {
	return &LLMDriver{
		name:    "llm",
		version: "1.0.0",
		devices: make(map[string]*LLMDevice),
		manager: manager,
	}
}

// Load implements Driver.Load
func (ld *LLMDriver) Load(config interface{}) error {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	// Create LLM device for the provider manager
	device := NewLLMDevice("llm0", ld.manager)
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

// LLMDevice represents an LLM provider device
type LLMDevice struct {
	mu        sync.RWMutex
	id        string
	name      string
	state     ecco9.DeviceState
	manager   *llm.ProviderManager
	metrics   ecco9.DeviceMetrics
	startTime time.Time
}

// NewLLMDevice creates a new LLM device
func NewLLMDevice(id string, manager *llm.ProviderManager) *LLMDevice {
	return &LLMDevice{
		id:      id,
		name:    fmt.Sprintf("LLM Provider %s", id),
		manager: manager,
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

	if ld.manager == nil {
		ld.state.Status = ecco9.DeviceStatusError
		ld.state.Health = ecco9.HealthStatusFailed
		return fmt.Errorf("provider manager not configured")
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

// Read implements CognitiveDevice.Read
func (ld *LLMDevice) Read(buffer []byte) (int, error) {
	ld.mu.RLock()
	defer ld.mu.RUnlock()

	// Return provider status as bytes
	status := "LLM Provider Manager Ready"
	if ld.manager == nil {
		status = "LLM Provider Manager Not Initialized"
	}
	statusBytes := []byte(status)
	n := copy(buffer, statusBytes)

	ld.metrics.OperationCount++
	ld.metrics.LastOperation = time.Now()

	return n, nil
}

// Write implements CognitiveDevice.Write
func (ld *LLMDevice) Write(buffer []byte) (int, error) {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	startTime := time.Now()

	// Process prompt through LLM provider
	prompt := string(buffer)
	ctx := context.Background()

	_, err := ld.manager.Generate(ctx, prompt, llm.DefaultGenerateOptions())
	if err != nil {
		ld.metrics.ErrorCount++
		return 0, err
	}

	ld.metrics.OperationCount++
	ld.metrics.LastOperation = time.Now()

	// Update average latency
	latency := time.Since(startTime)
	ld.metrics.AverageLatency = (ld.metrics.AverageLatency + latency) / 2

	return len(buffer), nil
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
