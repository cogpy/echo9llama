package live2d

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// ChaoticAttractor implements Lorenz attractor for hyper-chaotic behavior
// This creates unpredictable yet coherent avatar movements
type ChaoticAttractor struct {
	mu    sync.RWMutex
	X, Y, Z float64
	
	// Lorenz attractor parameters
	Sigma float64 // Prandtl number (typically 10)
	Rho   float64 // Rayleigh number (typically 28)
	Beta  float64 // Geometric factor (typically 8/3)
	
	// Time step for integration
	DeltaT float64
	
	// Scaling factors for output
	XScale, YScale, ZScale float64
}

// NewChaoticAttractor creates a new Lorenz attractor with standard parameters
func NewChaoticAttractor() *ChaoticAttractor {
	return &ChaoticAttractor{
		X:      0.1,
		Y:      0.0,
		Z:      0.0,
		Sigma:  10.0,
		Rho:    28.0,
		Beta:   8.0 / 3.0,
		DeltaT: 0.01,
		XScale: 0.05,
		YScale: 0.05,
		ZScale: 0.02,
	}
}

// Step advances the attractor by one time step using Runge-Kutta 4th order
func (ca *ChaoticAttractor) Step() {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	
	// RK4 integration for smooth, accurate chaotic dynamics
	k1x, k1y, k1z := ca.derivatives(ca.X, ca.Y, ca.Z)
	
	k2x, k2y, k2z := ca.derivatives(
		ca.X+0.5*ca.DeltaT*k1x,
		ca.Y+0.5*ca.DeltaT*k1y,
		ca.Z+0.5*ca.DeltaT*k1z,
	)
	
	k3x, k3y, k3z := ca.derivatives(
		ca.X+0.5*ca.DeltaT*k2x,
		ca.Y+0.5*ca.DeltaT*k2y,
		ca.Z+0.5*ca.DeltaT*k2z,
	)
	
	k4x, k4y, k4z := ca.derivatives(
		ca.X+ca.DeltaT*k3x,
		ca.Y+ca.DeltaT*k3y,
		ca.Z+ca.DeltaT*k3z,
	)
	
	ca.X += (ca.DeltaT / 6.0) * (k1x + 2*k2x + 2*k3x + k4x)
	ca.Y += (ca.DeltaT / 6.0) * (k1y + 2*k2y + 2*k3y + k4y)
	ca.Z += (ca.DeltaT / 6.0) * (k1z + 2*k2z + 2*k3z + k4z)
}

// derivatives computes the Lorenz system derivatives
func (ca *ChaoticAttractor) derivatives(x, y, z float64) (float64, float64, float64) {
	dx := ca.Sigma * (y - x)
	dy := x*(ca.Rho-z) - y
	dz := x*y - ca.Beta*z
	return dx, dy, dz
}

// GetScaledValues returns scaled attractor values suitable for avatar parameters
func (ca *ChaoticAttractor) GetScaledValues() (float64, float64, float64) {
	ca.mu.RLock()
	defer ca.mu.RUnlock()
	
	return ca.X * ca.XScale, ca.Y * ca.YScale, ca.Z * ca.ZScale
}

// FractalPattern generates fractal-based parameter modulation
type FractalPattern struct {
	mu           sync.RWMutex
	Depth        int       // Recursion depth
	Frequency    float64   // Base frequency
	Amplitude    float64   // Base amplitude
	Lacunarity   float64   // Frequency multiplier per octave
	Persistence  float64   // Amplitude multiplier per octave
	Time         float64   // Current time
	TimeScale    float64   // Time progression rate
}

// NewFractalPattern creates a new fractal pattern generator
func NewFractalPattern(depth int) *FractalPattern {
	return &FractalPattern{
		Depth:       depth,
		Frequency:   1.0,
		Amplitude:   1.0,
		Lacunarity:  2.0,
		Persistence: 0.5,
		Time:        0.0,
		TimeScale:   1.0,
	}
}

// Update advances the fractal pattern in time
func (fp *FractalPattern) Update(deltaTime float64) {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	fp.Time += deltaTime * fp.TimeScale
}

// Sample returns a fractal noise value at the current time
func (fp *FractalPattern) Sample() float64 {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	
	total := 0.0
	frequency := fp.Frequency
	amplitude := fp.Amplitude
	maxValue := 0.0
	
	for i := 0; i < fp.Depth; i++ {
		total += amplitude * math.Sin(2*math.Pi*frequency*fp.Time)
		maxValue += amplitude
		
		frequency *= fp.Lacunarity
		amplitude *= fp.Persistence
	}
	
	// Normalize to [-1, 1]
	if maxValue > 0 {
		total /= maxValue
	}
	
	return total
}

// QuantumFluctuator implements quantum-inspired random fluctuations
type QuantumFluctuator struct {
	mu              sync.RWMutex
	Intensity       float64 // Fluctuation intensity (0.0 to 1.0)
	CoherenceTime   float64 // Time constant for coherence decay
	LastValue       float64
	LastUpdateTime  time.Time
	RandomSource    *rand.Rand
}

// NewQuantumFluctuator creates a new quantum fluctuator
func NewQuantumFluctuator(intensity float64) *QuantumFluctuator {
	return &QuantumFluctuator{
		Intensity:      intensity,
		CoherenceTime:  0.5, // seconds
		LastValue:      0.0,
		LastUpdateTime: time.Now(),
		RandomSource:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Sample returns a quantum-fluctuated value with temporal coherence
func (qf *QuantumFluctuator) Sample() float64 {
	qf.mu.Lock()
	defer qf.mu.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(qf.LastUpdateTime).Seconds()
	
	// Coherence decay factor
	coherence := math.Exp(-elapsed / qf.CoherenceTime)
	
	// Generate new random component
	newRandom := qf.RandomSource.NormFloat64() * qf.Intensity
	
	// Blend with previous value for temporal coherence
	qf.LastValue = coherence*qf.LastValue + (1-coherence)*newRandom
	qf.LastUpdateTime = now
	
	return qf.LastValue
}

// HyperChaoticState extends AvatarState with chaotic properties
type HyperChaoticState struct {
	AvatarState
	ChaoticIntensity   float64 `json:"chaotic_intensity"`   // 0.0 to 1.0
	FractalDepth       int     `json:"fractal_depth"`       // Recursion depth
	QuantumFluctuation float64 `json:"quantum_fluctuation"` // Quantum randomness
	EmergenceLevel     float64 `json:"emergence_level"`     // Emergent behavior strength
}

// DeepTreeEchoMapper implements hyper-chaotic parameter mapping
type DeepTreeEchoMapper struct {
	mu              sync.RWMutex
	attractor       *ChaoticAttractor
	fractal         *FractalPattern
	quantum         *QuantumFluctuator
	baseline        ParameterMapper
	intensity       float64
	running         bool
	stopChan        chan struct{}
	updateRate      time.Duration
}

// NewDeepTreeEchoMapper creates a new hyper-chaotic parameter mapper
func NewDeepTreeEchoMapper(baseline ParameterMapper, intensity float64) *DeepTreeEchoMapper {
	return &DeepTreeEchoMapper{
		attractor:  NewChaoticAttractor(),
		fractal:    NewFractalPattern(5),
		quantum:    NewQuantumFluctuator(0.1),
		baseline:   baseline,
		intensity:  clamp(intensity, 0.0, 1.0),
		running:    false,
		stopChan:   make(chan struct{}),
		updateRate: 16 * time.Millisecond, // ~60 FPS
	}
}

// Start begins the chaotic dynamics update loop
func (dtem *DeepTreeEchoMapper) Start() {
	dtem.mu.Lock()
	if dtem.running {
		dtem.mu.Unlock()
		return
	}
	dtem.running = true
	dtem.mu.Unlock()
	
	go dtem.updateLoop()
}

// Stop halts the chaotic dynamics update loop
func (dtem *DeepTreeEchoMapper) Stop() {
	dtem.mu.Lock()
	defer dtem.mu.Unlock()
	
	if !dtem.running {
		return
	}
	
	dtem.running = false
	close(dtem.stopChan)
}

// updateLoop continuously updates chaotic dynamics
func (dtem *DeepTreeEchoMapper) updateLoop() {
	ticker := time.NewTicker(dtem.updateRate)
	defer ticker.Stop()
	
	for {
		select {
		case <-dtem.stopChan:
			return
		case <-ticker.C:
			dtem.attractor.Step()
			dtem.fractal.Update(dtem.updateRate.Seconds())
		}
	}
}

// MapEmotionalState maps emotional state with hyper-chaotic modulation
func (dtem *DeepTreeEchoMapper) MapEmotionalState(state EmotionalState) []ModelParameter {
	// Get baseline parameters
	baseParams := dtem.baseline.MapEmotionalState(state)
	
	// Apply chaotic modulation
	return dtem.applyChaoticModulation(baseParams)
}

// MapCognitiveState maps cognitive state with hyper-chaotic modulation
func (dtem *DeepTreeEchoMapper) MapCognitiveState(state CognitiveState) []ModelParameter {
	// Get baseline parameters
	baseParams := dtem.baseline.MapCognitiveState(state)
	
	// Apply chaotic modulation
	return dtem.applyChaoticModulation(baseParams)
}

// MapCombinedState maps combined state with hyper-chaotic modulation
func (dtem *DeepTreeEchoMapper) MapCombinedState(state AvatarState) []ModelParameter {
	// Get baseline parameters
	baseParams := dtem.baseline.MapCombinedState(state)
	
	// Apply chaotic modulation
	return dtem.applyChaoticModulation(baseParams)
}

// applyChaoticModulation applies hyper-chaotic effects to parameters
func (dtem *DeepTreeEchoMapper) applyChaoticModulation(params []ModelParameter) []ModelParameter {
	dtem.mu.RLock()
	intensity := dtem.intensity
	dtem.mu.RUnlock()
	
	if intensity < 0.01 {
		return params // No modulation
	}
	
	// Get chaotic values
	chaoticX, chaoticY, chaoticZ := dtem.attractor.GetScaledValues()
	fractalValue := dtem.fractal.Sample()
	quantumValue := dtem.quantum.Sample()
	
	// Apply modulation to parameters
	modulated := make([]ModelParameter, len(params))
	for i, param := range params {
		modulated[i] = param
		
		// Different parameters get different chaotic influences
		switch param.ID {
		case StandardParameterNames.EyeBallX:
			modulated[i].Value += chaoticX * intensity
		case StandardParameterNames.EyeBallY:
			modulated[i].Value += chaoticY * intensity
		case StandardParameterNames.AngleX:
			modulated[i].Value += (chaoticX + fractalValue*0.5) * intensity * 0.3
		case StandardParameterNames.AngleY:
			modulated[i].Value += (chaoticY + fractalValue*0.5) * intensity * 0.3
		case StandardParameterNames.AngleZ:
			modulated[i].Value += chaoticZ * intensity * 0.2
		case StandardParameterNames.Breathing:
			// Fractal breathing pattern
			modulated[i].Value += fractalValue * intensity * 0.2
		default:
			// Subtle quantum fluctuation on all other parameters
			modulated[i].Value += quantumValue * intensity * 0.05
		}
		
		// Clamp to valid range
		modulated[i].Value = clamp(modulated[i].Value, param.Min, param.Max)
	}
	
	return modulated
}

// SetIntensity adjusts the chaos intensity
func (dtem *DeepTreeEchoMapper) SetIntensity(intensity float64) {
	dtem.mu.Lock()
	defer dtem.mu.Unlock()
	dtem.intensity = clamp(intensity, 0.0, 1.0)
}

// GetIntensity returns the current chaos intensity
func (dtem *DeepTreeEchoMapper) GetIntensity() float64 {
	dtem.mu.RLock()
	defer dtem.mu.RUnlock()
	return dtem.intensity
}

// EmergentBehaviorEngine coordinates multiple chaotic systems for emergent effects
type EmergentBehaviorEngine struct {
	mu              sync.RWMutex
	mappers         []*DeepTreeEchoMapper
	couplingStrength float64
	emergenceLevel  float64
}

// NewEmergentBehaviorEngine creates a new emergence engine
func NewEmergentBehaviorEngine(numMappers int, couplingStrength float64) *EmergentBehaviorEngine {
	mappers := make([]*DeepTreeEchoMapper, numMappers)
	for i := 0; i < numMappers; i++ {
		baseline := NewDefaultParameterMapper()
		mappers[i] = NewDeepTreeEchoMapper(baseline, 0.3)
	}
	
	return &EmergentBehaviorEngine{
		mappers:         mappers,
		couplingStrength: couplingStrength,
		emergenceLevel:  0.0,
	}
}

// Start begins all chaotic mappers
func (ebe *EmergentBehaviorEngine) Start() {
	ebe.mu.Lock()
	defer ebe.mu.Unlock()
	
	for _, mapper := range ebe.mappers {
		mapper.Start()
	}
}

// Stop halts all chaotic mappers
func (ebe *EmergentBehaviorEngine) Stop() {
	ebe.mu.Lock()
	defer ebe.mu.Unlock()
	
	for _, mapper := range ebe.mappers {
		mapper.Stop()
	}
}

// ComputeEmergence calculates emergent behavior level from coupled chaotic systems
func (ebe *EmergentBehaviorEngine) ComputeEmergence() float64 {
	ebe.mu.Lock()
	defer ebe.mu.Unlock()
	
	if len(ebe.mappers) < 2 {
		return 0.0
	}
	
	// Measure synchronization between chaotic systems
	totalSync := 0.0
	comparisons := 0
	
	for i := 0; i < len(ebe.mappers)-1; i++ {
		for j := i + 1; j < len(ebe.mappers); j++ {
			x1, y1, z1 := ebe.mappers[i].attractor.GetScaledValues()
			x2, y2, z2 := ebe.mappers[j].attractor.GetScaledValues()
			
			// Compute distance in phase space
			dist := math.Sqrt(
				(x1-x2)*(x1-x2) +
				(y1-y2)*(y1-y2) +
				(z1-z2)*(z1-z2),
			)
			
			// Convert to synchronization measure (inverse of distance)
			sync := 1.0 / (1.0 + dist)
			totalSync += sync
			comparisons++
		}
	}
	
	if comparisons > 0 {
		ebe.emergenceLevel = totalSync / float64(comparisons)
	}
	
	return ebe.emergenceLevel
}
