package live2d

import (
	"sync"
)

// CircularBuffer is a fixed-size circular buffer for state history
type CircularBuffer struct {
	mu      sync.RWMutex
	data    []interface{}
	head    int
	size    int
	maxSize int
}

// NewCircularBuffer creates a new circular buffer
func NewCircularBuffer(maxSize int) *CircularBuffer {
	return &CircularBuffer{
		data:    make([]interface{}, maxSize),
		head:    0,
		size:    0,
		maxSize: maxSize,
	}
}

// Add adds an item to the buffer
func (cb *CircularBuffer) Add(item interface{}) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.data[cb.head] = item
	cb.head = (cb.head + 1) % cb.maxSize

	if cb.size < cb.maxSize {
		cb.size++
	}
}

// GetLast returns the last n items
func (cb *CircularBuffer) GetLast(n int) []interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if n > cb.size {
		n = cb.size
	}

	result := make([]interface{}, n)
	for i := 0; i < n; i++ {
		idx := (cb.head - 1 - i + cb.maxSize) % cb.maxSize
		result[n-1-i] = cb.data[idx]
	}

	return result
}

// Len returns the current size
func (cb *CircularBuffer) Len() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.size
}

// StateDiffCalculator calculates differences between states
type StateDiffCalculator struct {
	threshold float64
}

// NewStateDiffCalculator creates a new diff calculator
func NewStateDiffCalculator() *StateDiffCalculator {
	return &StateDiffCalculator{
		threshold: 0.01, // Minimum change to consider
	}
}

// StateDiff represents differences between states
type StateDiff struct {
	ChangedDimensions []string
	ChangeMAgnitude   float64
	EmotionalChange   float64
	CognitiveChange   float64
}

// Calculate calculates differences between two states
func (sdc *StateDiffCalculator) Calculate(state1, state2 UnifiedAvatarState) StateDiff {
	diff := StateDiff{
		ChangedDimensions: make([]string, 0),
	}

	// Calculate emotional changes
	emotionalChange := 0.0
	if abs(state1.Emotional.Valence-state2.Emotional.Valence) > sdc.threshold {
		diff.ChangedDimensions = append(diff.ChangedDimensions, "emotional.valence")
		emotionalChange += abs(state1.Emotional.Valence - state2.Emotional.Valence)
	}
	if abs(state1.Emotional.Arousal-state2.Emotional.Arousal) > sdc.threshold {
		diff.ChangedDimensions = append(diff.ChangedDimensions, "emotional.arousal")
		emotionalChange += abs(state1.Emotional.Arousal - state2.Emotional.Arousal)
	}
	if abs(state1.Emotional.Dominance-state2.Emotional.Dominance) > sdc.threshold {
		diff.ChangedDimensions = append(diff.ChangedDimensions, "emotional.dominance")
		emotionalChange += abs(state1.Emotional.Dominance - state2.Emotional.Dominance)
	}
	if abs(state1.Emotional.Curiosity-state2.Emotional.Curiosity) > sdc.threshold {
		diff.ChangedDimensions = append(diff.ChangedDimensions, "emotional.curiosity")
		emotionalChange += abs(state1.Emotional.Curiosity - state2.Emotional.Curiosity)
	}
	if abs(state1.Emotional.Confidence-state2.Emotional.Confidence) > sdc.threshold {
		diff.ChangedDimensions = append(diff.ChangedDimensions, "emotional.confidence")
		emotionalChange += abs(state1.Emotional.Confidence - state2.Emotional.Confidence)
	}
	diff.EmotionalChange = emotionalChange

	// Calculate cognitive changes
	cognitiveChange := 0.0
	if abs(state1.Cognitive.Awareness-state2.Cognitive.Awareness) > sdc.threshold {
		diff.ChangedDimensions = append(diff.ChangedDimensions, "cognitive.awareness")
		cognitiveChange += abs(state1.Cognitive.Awareness - state2.Cognitive.Awareness)
	}
	if abs(state1.Cognitive.Attention-state2.Cognitive.Attention) > sdc.threshold {
		diff.ChangedDimensions = append(diff.ChangedDimensions, "cognitive.attention")
		cognitiveChange += abs(state1.Cognitive.Attention - state2.Cognitive.Attention)
	}
	if abs(state1.Cognitive.CognitiveLoad-state2.Cognitive.CognitiveLoad) > sdc.threshold {
		diff.ChangedDimensions = append(diff.ChangedDimensions, "cognitive.load")
		cognitiveChange += abs(state1.Cognitive.CognitiveLoad - state2.Cognitive.CognitiveLoad)
	}
	if abs(state1.Cognitive.Coherence-state2.Cognitive.Coherence) > sdc.threshold {
		diff.ChangedDimensions = append(diff.ChangedDimensions, "cognitive.coherence")
		cognitiveChange += abs(state1.Cognitive.Coherence - state2.Cognitive.Coherence)
	}
	if abs(state1.Cognitive.EnergyLevel-state2.Cognitive.EnergyLevel) > sdc.threshold {
		diff.ChangedDimensions = append(diff.ChangedDimensions, "cognitive.energy")
		cognitiveChange += abs(state1.Cognitive.EnergyLevel - state2.Cognitive.EnergyLevel)
	}
	if state1.Cognitive.ProcessingMode != state2.Cognitive.ProcessingMode {
		diff.ChangedDimensions = append(diff.ChangedDimensions, "cognitive.mode")
		cognitiveChange += 0.5 // Mode change has fixed weight
	}
	diff.CognitiveChange = cognitiveChange

	// Calculate reservoir changes
	if abs(state1.ReservoirDynamics.SpectralRadius-state2.ReservoirDynamics.SpectralRadius) > sdc.threshold {
		diff.ChangedDimensions = append(diff.ChangedDimensions, "reservoir.spectral_radius")
	}
	if abs(state1.ReservoirDynamics.InputScaling-state2.ReservoirDynamics.InputScaling) > sdc.threshold {
		diff.ChangedDimensions = append(diff.ChangedDimensions, "reservoir.input_scaling")
	}
	if abs(state1.ReservoirDynamics.LeakRate-state2.ReservoirDynamics.LeakRate) > sdc.threshold {
		diff.ChangedDimensions = append(diff.ChangedDimensions, "reservoir.leak_rate")
	}

	// EchoBeats phase change
	if state1.EchoBeatPosition.Step != state2.EchoBeatPosition.Step {
		diff.ChangedDimensions = append(diff.ChangedDimensions, "echobeats.step")
	}

	// Thought activity change
	if state1.ThoughtActivity.Active != state2.ThoughtActivity.Active {
		diff.ChangedDimensions = append(diff.ChangedDimensions, "thought.active")
	}

	// Overall magnitude
	diff.ChangeMAgnitude = emotionalChange + cognitiveChange

	return diff
}

// RateLimiter limits update rate
type RateLimiter struct {
	mu              sync.Mutex
	minInterval     float64 // Milliseconds
	maxInterval     float64 // Milliseconds
	currentInterval float64
	lastUpdate      int64
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(minInterval, maxInterval float64) *RateLimiter {
	return &RateLimiter{
		minInterval:     minInterval,
		maxInterval:     maxInterval,
		currentInterval: minInterval,
		lastUpdate:      0,
	}
}

// ShouldUpdate checks if an update should occur
func (rl *RateLimiter) ShouldUpdate(changeMA float64, nowMS int64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Adjust interval based on change magnitude
	if changeMA > 0.5 {
		rl.currentInterval = rl.minInterval
	} else if changeMA < 0.1 {
		rl.currentInterval = rl.maxInterval
	} else {
		// Interpolate
		factor := (0.5 - changeMA) / 0.4
		rl.currentInterval = rl.minInterval + (rl.maxInterval-rl.minInterval)*factor
	}

	// Check if enough time has passed
	elapsed := float64(nowMS - rl.lastUpdate)
	if elapsed >= rl.currentInterval {
		rl.lastUpdate = nowMS
		return true
	}

	return false
}

// DifferentialUpdater tracks parameter changes
type DifferentialUpdater struct {
	mu              sync.RWMutex
	lastParameters  map[string]float64
	changeThreshold float64
}

// NewDifferentialUpdater creates a new updater
func NewDifferentialUpdater(threshold float64) *DifferentialUpdater {
	return &DifferentialUpdater{
		lastParameters:  make(map[string]float64),
		changeThreshold: threshold,
	}
}

// GetDiff returns only changed parameters
func (du *DifferentialUpdater) GetDiff(newParameters []ModelParameter) []ModelParameter {
	du.mu.Lock()
	defer du.mu.Unlock()

	diff := make([]ModelParameter, 0)

	for _, param := range newParameters {
		lastValue, exists := du.lastParameters[param.ID]

		// Include if doesn't exist or changed significantly
		if !exists || abs(param.Value-lastValue) > du.changeThreshold {
			diff = append(diff, param)
			du.lastParameters[param.ID] = param.Value
		}
	}

	return diff
}

// Reset resets the differential updater
func (du *DifferentialUpdater) Reset() {
	du.mu.Lock()
	defer du.mu.Unlock()
	du.lastParameters = make(map[string]float64)
}

// ParameterCache caches parameter calculations
type ParameterCache struct {
	mu      sync.RWMutex
	cache   map[string][]ModelParameter
	maxSize int
	keys    []string // LRU tracking
}

// NewParameterCache creates a new cache
func NewParameterCache(maxSize int) *ParameterCache {
	return &ParameterCache{
		cache:   make(map[string][]ModelParameter),
		maxSize: maxSize,
		keys:    make([]string, 0),
	}
}

// Get retrieves from cache
func (pc *ParameterCache) Get(key string) ([]ModelParameter, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	params, exists := pc.cache[key]
	return params, exists
}

// Set stores in cache
func (pc *ParameterCache) Set(key string, params []ModelParameter) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Add to cache
	pc.cache[key] = params

	// Track key
	pc.keys = append(pc.keys, key)

	// Evict oldest if over size
	if len(pc.keys) > pc.maxSize {
		oldestKey := pc.keys[0]
		delete(pc.cache, oldestKey)
		pc.keys = pc.keys[1:]
	}
}

// Clear clears the cache
func (pc *ParameterCache) Clear() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.cache = make(map[string][]ModelParameter)
	pc.keys = make([]string, 0)
}

// StateHasher creates hash keys from states
type StateHasher struct{}

// Hash creates a hash from a state
func (sh *StateHasher) Hash(state UnifiedAvatarState) string {
	// Quantize to reduce cache misses
	hash := ""

	// Emotional (quantized to 0.1)
	hash += string([]byte{
		byte(int(state.Emotional.Valence*10) + 10),
		byte(int(state.Emotional.Arousal * 10)),
		byte(int(state.Emotional.Dominance * 10)),
		byte(int(state.Emotional.Curiosity * 10)),
		byte(int(state.Emotional.Confidence * 10)),
	})

	// Cognitive (quantized to 0.1)
	hash += string([]byte{
		byte(int(state.Cognitive.Awareness * 10)),
		byte(int(state.Cognitive.Attention * 10)),
		byte(int(state.Cognitive.CognitiveLoad * 10)),
		byte(int(state.Cognitive.Coherence * 10)),
		byte(int(state.Cognitive.EnergyLevel * 10)),
	})

	// Processing mode
	modeMap := map[string]byte{
		"contemplative": 0,
		"dynamic":       1,
		"cautious":      2,
		"creative":      3,
	}
	hash += string([]byte{modeMap[state.Cognitive.ProcessingMode]})

	// EchoBeats step
	hash += string([]byte{byte(state.EchoBeatPosition.Step)})

	// Archetype
	archetypeMap := map[CognitiveArchetype]byte{
		ArchetypeChaos:   0,
		ArchetypeOrder:   1,
		ArchetypeBalance: 2,
	}
	hash += string([]byte{archetypeMap[state.ActiveArchetype]})

	return hash
}
