package drivers

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/EchoCog/echollama/core/ecco9"
)

// EmotionDriver implements the driver for Emotion Processing Units
type EmotionDriver struct {
	mu      sync.RWMutex
	name    string
	version string
	devices map[string]*EmotionDevice
	config  *EmotionConfig
}

// EmotionConfig holds configuration for emotion processing
type EmotionConfig struct {
	NumChannels   int
	DecayRate     float64
	BlendingAlpha float64
}

// NewEmotionDriver creates a new emotion driver
func NewEmotionDriver() *EmotionDriver {
	return &EmotionDriver{
		name:    "emotion",
		version: "1.0.0",
		devices: make(map[string]*EmotionDevice),
		config:  DefaultEmotionConfig(),
	}
}

// DefaultEmotionConfig returns default emotion configuration
func DefaultEmotionConfig() *EmotionConfig {
	return &EmotionConfig{
		NumChannels:   10, // Izard's Differential Emotion Theory
		DecayRate:     0.05,
		BlendingAlpha: 0.3,
	}
}

// Load implements Driver.Load
func (ed *EmotionDriver) Load(config interface{}) error {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	if cfg, ok := config.(*EmotionConfig); ok {
		ed.config = cfg
	}

	// Create emotion processing device
	device := NewEmotionDevice("emotion0", ed.config)
	ed.devices["emotion0"] = device

	return nil
}

// Unload implements Driver.Unload
func (ed *EmotionDriver) Unload() error {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	for _, device := range ed.devices {
		ctx := context.Background()
		if err := device.Shutdown(ctx); err != nil {
			return err
		}
	}

	ed.devices = make(map[string]*EmotionDevice)
	return nil
}

// GetDevice implements Driver.GetDevice
func (ed *EmotionDriver) GetDevice(id string) (ecco9.CognitiveDevice, error) {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	device, exists := ed.devices[id]
	if !exists {
		return nil, fmt.Errorf("device %s not found", id)
	}
	return device, nil
}

// ListDevices implements Driver.ListDevices
func (ed *EmotionDriver) ListDevices() []ecco9.CognitiveDevice {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	devices := make([]ecco9.CognitiveDevice, 0, len(ed.devices))
	for _, device := range ed.devices {
		devices = append(devices, device)
	}
	return devices
}

// GetName implements Driver.GetName
func (ed *EmotionDriver) GetName() string {
	return ed.name
}

// GetVersion implements Driver.GetVersion
func (ed *EmotionDriver) GetVersion() string {
	return ed.version
}

// GetCapabilities implements Driver.GetCapabilities
func (ed *EmotionDriver) GetCapabilities() []string {
	return []string{
		"discrete-emotions",
		"dimensional-affect",
		"emotion-blending",
		"cognitive-modulation",
		"automatic-decay",
	}
}

// EmotionDevice represents an Emotion Processing Unit
type EmotionDevice struct {
	mu        sync.RWMutex
	id        string
	name      string
	state     ecco9.DeviceState
	config    *EmotionConfig
	emotions  map[string]*EmotionChannel
	arousal   float64 // Dimensional affect: arousal
	valence   float64 // Dimensional affect: valence
	metrics   ecco9.DeviceMetrics
	startTime time.Time
}

// EmotionChannel represents a discrete emotion channel
type EmotionChannel struct {
	Name       string
	Intensity  float64
	LastUpdate time.Time
}

// NewEmotionDevice creates a new emotion device
func NewEmotionDevice(id string, config *EmotionConfig) *EmotionDevice {
	return &EmotionDevice{
		id:       id,
		name:     fmt.Sprintf("Emotion Processing Unit %s", id),
		config:   config,
		emotions: make(map[string]*EmotionChannel),
		arousal:  0.5,
		valence:  0.5,
		state: ecco9.DeviceState{
			ID:     id,
			Name:   fmt.Sprintf("Emotion %s", id),
			Status: ecco9.DeviceStatusOffline,
			Power:  ecco9.PowerStateOff,
			Health: ecco9.HealthStatusHealthy,
		},
		metrics: ecco9.DeviceMetrics{},
	}
}

// Initialize implements CognitiveDevice.Initialize
func (ed *EmotionDevice) Initialize(ctx context.Context) error {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	ed.state.Status = ecco9.DeviceStatusInitializing
	ed.state.Power = ecco9.PowerStateActive

	// Initialize emotion channels (Izard's 10 basic emotions)
	emotionNames := []string{
		"interest",
		"joy",
		"surprise",
		"sadness",
		"anger",
		"disgust",
		"contempt",
		"fear",
		"shame",
		"guilt",
	}

	for _, name := range emotionNames {
		ed.emotions[name] = &EmotionChannel{
			Name:       name,
			Intensity:  0.0,
			LastUpdate: time.Now(),
		}
	}

	ed.startTime = time.Now()
	ed.state.Status = ecco9.DeviceStatusReady
	ed.state.LastUpdate = time.Now()

	// Start decay ticker
	go ed.runDecayLoop(ctx)

	return nil
}

// runDecayLoop applies automatic emotion decay
func (ed *EmotionDevice) runDecayLoop(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ed.applyDecay()
		}
	}
}

// applyDecay reduces emotion intensities over time
func (ed *EmotionDevice) applyDecay() {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	for _, emotion := range ed.emotions {
		emotion.Intensity *= (1.0 - ed.config.DecayRate)
		if emotion.Intensity < 0.01 {
			emotion.Intensity = 0.0
		}
	}

	// Decay arousal and valence toward neutral
	ed.arousal += (0.5 - ed.arousal) * ed.config.DecayRate
	ed.valence += (0.5 - ed.valence) * ed.config.DecayRate
}

// Shutdown implements CognitiveDevice.Shutdown
func (ed *EmotionDevice) Shutdown(ctx context.Context) error {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	ed.state.Status = ecco9.DeviceStatusOffline
	ed.state.Power = ecco9.PowerStateOff
	ed.emotions = make(map[string]*EmotionChannel)

	return nil
}

// Reset implements CognitiveDevice.Reset
func (ed *EmotionDevice) Reset(ctx context.Context) error {
	if err := ed.Shutdown(ctx); err != nil {
		return err
	}
	return ed.Initialize(ctx)
}

// GetState implements CognitiveDevice.GetState
func (ed *EmotionDevice) GetState() (ecco9.DeviceState, error) {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	state := ed.state
	state.Uptime = time.Since(ed.startTime)
	state.Metrics = ed.metrics

	return state, nil
}

// SetState implements CognitiveDevice.SetState
func (ed *EmotionDevice) SetState(state ecco9.DeviceState) error {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	ed.state = state
	return nil
}

// Read implements CognitiveDevice.Read
func (ed *EmotionDevice) Read(buffer []byte) (int, error) {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	if ed.emotions == nil {
		return 0, fmt.Errorf("emotion channels not initialized")
	}

	// Read current emotion state
	status := fmt.Sprintf("Arousal:%.2f Valence:%.2f | ", ed.arousal, ed.valence)

	// Add top 3 emotions
	type emotionIntensity struct {
		name      string
		intensity float64
	}

	emotions := make([]emotionIntensity, 0)
	for name, channel := range ed.emotions {
		if channel.Intensity > 0.01 {
			emotions = append(emotions, emotionIntensity{name, channel.Intensity})
		}
	}

	// Sort by intensity (simple bubble sort for small arrays)
	for i := 0; i < len(emotions); i++ {
		for j := i + 1; j < len(emotions); j++ {
			if emotions[j].intensity > emotions[i].intensity {
				emotions[i], emotions[j] = emotions[j], emotions[i]
			}
		}
	}

	// Add top emotions to status
	for i := 0; i < len(emotions) && i < 3; i++ {
		status += fmt.Sprintf("%s:%.2f ", emotions[i].name, emotions[i].intensity)
	}

	n := copy(buffer, []byte(status))
	ed.metrics.OperationCount++
	ed.metrics.LastOperation = time.Now()

	return n, nil
}

// Write implements CognitiveDevice.Write
func (ed *EmotionDevice) Write(buffer []byte) (int, error) {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	if ed.emotions == nil {
		return 0, fmt.Errorf("emotion channels not initialized")
	}

	startTime := time.Now()

	// Parse input to trigger emotions
	input := string(buffer)
	inputLen := float64(len(input))

	// Simple heuristic: trigger emotions based on input characteristics
	if inputLen > 0 {
		// Trigger interest based on input length
		ed.setEmotion("interest", math.Min(inputLen/100.0, 1.0))

		// Update arousal and valence
		ed.arousal = math.Min(0.5+inputLen/200.0, 1.0)
		ed.valence = 0.5 + 0.2 // Slight positive bias
	}

	ed.metrics.OperationCount++
	ed.metrics.LastOperation = time.Now()

	// Update average latency
	latency := time.Since(startTime)
	ed.metrics.AverageLatency = (ed.metrics.AverageLatency + latency) / 2

	return len(buffer), nil
}

// setEmotion updates an emotion channel with blending
func (ed *EmotionDevice) setEmotion(name string, intensity float64) {
	if channel, exists := ed.emotions[name]; exists {
		// Blend new intensity with existing
		channel.Intensity = channel.Intensity*(1.0-ed.config.BlendingAlpha) +
			intensity*ed.config.BlendingAlpha
		channel.LastUpdate = time.Now()
	}
}

// IoCtl implements CognitiveDevice.IoCtl
func (ed *EmotionDevice) IoCtl(command uint32, arg interface{}) error {
	// Implement device-specific control operations
	return nil
}

// GetMetrics implements CognitiveDevice.GetMetrics
func (ed *EmotionDevice) GetMetrics() (ecco9.DeviceMetrics, error) {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	return ed.metrics, nil
}

// GetHealth implements CognitiveDevice.GetHealth
func (ed *EmotionDevice) GetHealth() (ecco9.HealthStatus, error) {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	return ed.state.Health, nil
}

// GetID implements CognitiveDevice.GetID
func (ed *EmotionDevice) GetID() string {
	return ed.id
}

// GetName implements CognitiveDevice.GetName
func (ed *EmotionDevice) GetName() string {
	return ed.name
}

// GetType implements CognitiveDevice.GetType
func (ed *EmotionDevice) GetType() ecco9.DeviceType {
	return ecco9.DeviceTypeEmotion
}
