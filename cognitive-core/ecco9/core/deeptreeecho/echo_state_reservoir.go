package deeptreeecho

import (
	"math"
	"math/rand"
	"sync"
)

// PersonaType represents different cognitive styles
type PersonaType int

const (
	PersonaContemplativeScholar PersonaType = iota
	PersonaDynamicExplorer
	PersonaCautiousAnalyst
	PersonaCreativeVisionary
)

func (pt PersonaType) String() string {
	names := []string{
		"Contemplative Scholar",
		"Dynamic Explorer",
		"Cautious Analyst",
		"Creative Visionary",
	}
	if int(pt) < len(names) {
		return names[pt]
	}
	return "Unknown"
}

// EchoStateReservoir implements reservoir computing with persona-based configurations
// Based on Echo State Networks for multi-timescale temporal processing
type EchoStateReservoir struct {
	mu sync.RWMutex

	// Reservoir parameters (persona-specific)
	spectralRadius float64 // Controls memory vs. responsiveness
	inputScaling   float64 // Input influence strength
	leakRate       float64 // State update speed

	// Reservoir structure
	size           int         // Number of reservoir neurons
	reservoirState []float64   // Current internal state
	weights        [][]float64 // Fixed random internal weights
	inputWeights   [][]float64 // Input-to-reservoir weights

	// Hierarchy
	level           int
	parentReservoir *EchoStateReservoir
	childReservoirs []*EchoStateReservoir

	// Persona
	persona PersonaType

	// History
	stateHistory [][]float64
	maxHistory   int

	// Metrics
	echoProperty float64 // Measure of echo state property
	complexity   float64 // State space complexity
}

// PersonaConfig defines reservoir parameters for each persona
type PersonaConfig struct {
	SpectralRadius float64
	InputScaling   float64
	LeakRate       float64
	Description    string
}

// GetPersonaConfig returns configuration for a persona type
func GetPersonaConfig(persona PersonaType) PersonaConfig {
	configs := map[PersonaType]PersonaConfig{
		PersonaContemplativeScholar: {
			SpectralRadius: 0.95,
			InputScaling:   0.3,
			LeakRate:       0.2,
			Description:    "Deep memory, slow deliberation, reflection over reaction",
		},
		PersonaDynamicExplorer: {
			SpectralRadius: 0.7,
			InputScaling:   0.8,
			LeakRate:       0.8,
			Description:    "Low memory, rapid adaptation, exploration over exploitation",
		},
		PersonaCautiousAnalyst: {
			SpectralRadius: 0.99,
			InputScaling:   0.2,
			LeakRate:       0.3,
			Description:    "Maximal stability, conservative, systematic processing",
		},
		PersonaCreativeVisionary: {
			SpectralRadius: 0.85,
			InputScaling:   0.7,
			LeakRate:       0.6,
			Description:    "Edge of chaos, flexible memory, transformation-seeking",
		},
	}

	if config, exists := configs[persona]; exists {
		return config
	}

	// Default to contemplative
	return configs[PersonaContemplativeScholar]
}

// NewEchoStateReservoir creates a new echo state reservoir with persona configuration
func NewEchoStateReservoir(size int, persona PersonaType, level int) *EchoStateReservoir {
	config := GetPersonaConfig(persona)

	esr := &EchoStateReservoir{
		spectralRadius:  config.SpectralRadius,
		inputScaling:    config.InputScaling,
		leakRate:        config.LeakRate,
		size:            size,
		reservoirState:  make([]float64, size),
		level:           level,
		persona:         persona,
		childReservoirs: make([]*EchoStateReservoir, 0),
		stateHistory:    make([][]float64, 0),
		maxHistory:      100,
	}

	// Initialize reservoir weights
	esr.initializeWeights()

	return esr
}

// initializeWeights creates random reservoir weights with spectral radius scaling
func (esr *EchoStateReservoir) initializeWeights() {
	// Create random internal weights
	esr.weights = make([][]float64, esr.size)
	for i := range esr.weights {
		esr.weights[i] = make([]float64, esr.size)
		for j := range esr.weights[i] {
			esr.weights[i][j] = (rand.Float64()*2.0 - 1.0) * 0.5
		}
	}

	// Scale to desired spectral radius
	esr.scaleToSpectralRadius()

	// Initialize input weights (will be set when input dimension is known)
	esr.inputWeights = make([][]float64, 0)
}

// scaleToSpectralRadius adjusts weights to achieve target spectral radius
func (esr *EchoStateReservoir) scaleToSpectralRadius() {
	// Simplified: scale by target spectral radius
	// In full implementation, would compute actual spectral radius and scale
	scale := esr.spectralRadius / 1.0 // Assuming initial spectral radius ~1.0

	for i := range esr.weights {
		for j := range esr.weights[i] {
			esr.weights[i][j] *= scale
		}
	}
}

// Update processes input and updates reservoir state
func (esr *EchoStateReservoir) Update(input []float64) []float64 {
	esr.mu.Lock()
	defer esr.mu.Unlock()

	// Ensure input weights are initialized
	if len(esr.inputWeights) == 0 {
		esr.initializeInputWeights(len(input))
	}

	// Compute new state
	newState := make([]float64, esr.size)

	for i := 0; i < esr.size; i++ {
		// Input contribution
		inputSum := 0.0
		for j := 0; j < len(input); j++ {
			if j < len(esr.inputWeights[i]) {
				inputSum += esr.inputWeights[i][j] * input[j]
			}
		}
		inputSum *= esr.inputScaling

		// Reservoir contribution
		reservoirSum := 0.0
		for j := 0; j < esr.size; j++ {
			reservoirSum += esr.weights[i][j] * esr.reservoirState[j]
		}

		// Leaky integration
		// x(t+1) = (1-α)x(t) + α·tanh(W_in·u(t) + W·x(t))
		newState[i] = (1.0-esr.leakRate)*esr.reservoirState[i] +
			esr.leakRate*math.Tanh(inputSum+reservoirSum)
	}

	// Update state
	esr.reservoirState = newState

	// Record in history
	esr.recordState(newState)

	// Update metrics
	esr.updateMetrics()

	return newState
}

// initializeInputWeights creates input-to-reservoir weights
func (esr *EchoStateReservoir) initializeInputWeights(inputDim int) {
	esr.inputWeights = make([][]float64, esr.size)
	for i := range esr.inputWeights {
		esr.inputWeights[i] = make([]float64, inputDim)
		for j := range esr.inputWeights[i] {
			esr.inputWeights[i][j] = (rand.Float64()*2.0 - 1.0) * 0.5
		}
	}
}

// recordState adds current state to history
func (esr *EchoStateReservoir) recordState(state []float64) {
	// Copy state
	stateCopy := make([]float64, len(state))
	copy(stateCopy, state)

	esr.stateHistory = append(esr.stateHistory, stateCopy)

	// Keep last maxHistory states
	if len(esr.stateHistory) > esr.maxHistory {
		esr.stateHistory = esr.stateHistory[1:]
	}
}

// updateMetrics calculates reservoir metrics
func (esr *EchoStateReservoir) updateMetrics() {
	// Echo property: measure of fading memory
	esr.echoProperty = esr.calculateEchoProperty()

	// Complexity: diversity of states
	esr.complexity = esr.calculateComplexity()
}

// calculateEchoProperty measures how well the reservoir exhibits echo state property
func (esr *EchoStateReservoir) calculateEchoProperty() float64 {
	if len(esr.stateHistory) < 2 {
		return 1.0
	}

	// Measure state change rate
	recent := esr.stateHistory[len(esr.stateHistory)-1]
	previous := esr.stateHistory[len(esr.stateHistory)-2]

	changeSum := 0.0
	for i := range recent {
		diff := recent[i] - previous[i]
		changeSum += diff * diff
	}

	// Echo property is good when changes are moderate (not too stable, not too chaotic)
	change := math.Sqrt(changeSum / float64(len(recent)))

	// Optimal change rate around 0.1-0.3
	optimal := 0.2
	echoProperty := 1.0 - math.Abs(change-optimal)/optimal

	return math.Max(0.0, math.Min(1.0, echoProperty))
}

// calculateComplexity measures state space diversity
func (esr *EchoStateReservoir) calculateComplexity() float64 {
	if len(esr.stateHistory) < 10 {
		return 0.5
	}

	// Calculate variance across history
	means := make([]float64, esr.size)
	for _, state := range esr.stateHistory {
		for i, val := range state {
			means[i] += val
		}
	}

	for i := range means {
		means[i] /= float64(len(esr.stateHistory))
	}

	variances := make([]float64, esr.size)
	for _, state := range esr.stateHistory {
		for i, val := range state {
			diff := val - means[i]
			variances[i] += diff * diff
		}
	}

	avgVariance := 0.0
	for i := range variances {
		variances[i] /= float64(len(esr.stateHistory))
		avgVariance += variances[i]
	}
	avgVariance /= float64(len(variances))

	// Normalize to 0-1 range
	complexity := math.Min(avgVariance*2.0, 1.0)

	return complexity
}

// GetState returns current reservoir state
func (esr *EchoStateReservoir) GetState() []float64 {
	esr.mu.RLock()
	defer esr.mu.RUnlock()

	state := make([]float64, len(esr.reservoirState))
	copy(state, esr.reservoirState)
	return state
}

// Reset clears reservoir state
func (esr *EchoStateReservoir) Reset() {
	esr.mu.Lock()
	defer esr.mu.Unlock()

	for i := range esr.reservoirState {
		esr.reservoirState[i] = 0.0
	}

	esr.stateHistory = make([][]float64, 0)
}

// AddChild adds a child reservoir for hierarchical processing
func (esr *EchoStateReservoir) AddChild(child *EchoStateReservoir) {
	esr.mu.Lock()
	defer esr.mu.Unlock()

	child.parentReservoir = esr
	child.level = esr.level + 1
	esr.childReservoirs = append(esr.childReservoirs, child)
}

// ProcessHierarchical processes input through hierarchical reservoir structure
func (esr *EchoStateReservoir) ProcessHierarchical(input []float64) map[int][]float64 {
	// Process at this level
	state := esr.Update(input)

	results := make(map[int][]float64)
	results[esr.level] = state

	// Process children with this level's state as input
	for _, child := range esr.childReservoirs {
		childResults := child.ProcessHierarchical(state)
		for level, childState := range childResults {
			results[level] = childState
		}
	}

	return results
}

// GetMetrics returns reservoir metrics
func (esr *EchoStateReservoir) GetMetrics() map[string]interface{} {
	esr.mu.RLock()
	defer esr.mu.RUnlock()

	return map[string]interface{}{
		"persona":         esr.persona.String(),
		"spectral_radius": esr.spectralRadius,
		"input_scaling":   esr.inputScaling,
		"leak_rate":       esr.leakRate,
		"size":            esr.size,
		"level":           esr.level,
		"echo_property":   esr.echoProperty,
		"complexity":      esr.complexity,
		"history_size":    len(esr.stateHistory),
		"child_count":     len(esr.childReservoirs),
	}
}

// HierarchicalReservoirSystem manages multiple reservoirs in hierarchy
type HierarchicalReservoirSystem struct {
	mu        sync.RWMutex
	root      *EchoStateReservoir
	allLevels map[int][]*EchoStateReservoir
}

// NewHierarchicalReservoirSystem creates a multi-level reservoir system
func NewHierarchicalReservoirSystem(
	levelsConfig []struct {
		Size    int
		Persona PersonaType
	},
) *HierarchicalReservoirSystem {
	hrs := &HierarchicalReservoirSystem{
		allLevels: make(map[int][]*EchoStateReservoir),
	}

	// Create root
	if len(levelsConfig) > 0 {
		hrs.root = NewEchoStateReservoir(
			levelsConfig[0].Size,
			levelsConfig[0].Persona,
			0,
		)
		hrs.allLevels[0] = []*EchoStateReservoir{hrs.root}

		// Create children
		parent := hrs.root
		for i := 1; i < len(levelsConfig); i++ {
			child := NewEchoStateReservoir(
				levelsConfig[i].Size,
				levelsConfig[i].Persona,
				i,
			)
			parent.AddChild(child)
			hrs.allLevels[i] = append(hrs.allLevels[i], child)
			parent = child
		}
	}

	return hrs
}

// Process input through entire hierarchy
func (hrs *HierarchicalReservoirSystem) Process(input []float64) map[int][]float64 {
	if hrs.root == nil {
		return make(map[int][]float64)
	}

	return hrs.root.ProcessHierarchical(input)
}

// GetSystemMetrics returns metrics for entire system
func (hrs *HierarchicalReservoirSystem) GetSystemMetrics() map[string]interface{} {
	hrs.mu.RLock()
	defer hrs.mu.RUnlock()

	levelMetrics := make(map[int]interface{})

	for level, reservoirs := range hrs.allLevels {
		metrics := make([]map[string]interface{}, len(reservoirs))
		for i, reservoir := range reservoirs {
			metrics[i] = reservoir.GetMetrics()
		}
		levelMetrics[level] = metrics
	}

	return map[string]interface{}{
		"total_levels":  len(hrs.allLevels),
		"level_metrics": levelMetrics,
	}
}
