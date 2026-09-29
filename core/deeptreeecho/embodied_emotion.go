package deeptreeecho

import (
	"math"
	"sync"
	"time"
)

// EmotionType represents discrete emotion categories (Izard's Differential Emotion Theory)
type EmotionType int

const (
	EmotionInterest EmotionType = iota
	EmotionJoy
	EmotionSurprise
	EmotionSadness
	EmotionAnger
	EmotionDisgust
	EmotionContempt
	EmotionFear
	EmotionShame
	EmotionGuilt
)

func (et EmotionType) String() string {
	names := []string{
		"Interest", "Joy", "Surprise", "Sadness", "Anger",
		"Disgust", "Contempt", "Fear", "Shame", "Guilt",
	}
	if int(et) < len(names) {
		return names[et]
	}
	return "Unknown"
}

// Emotion represents a discrete emotional state with cognitive effects
type Emotion struct {
	Type      EmotionType
	Intensity float64 // 0.0 to 1.0
	Duration  time.Duration
	OnsetTime time.Time

	// Cognitive effects (constitutive of knowing, not decorative)
	AttentionScope    float64 // Joy broadens (>1.0), fear narrows (<1.0)
	ProcessingDepth   float64 // Wonder deepens (>1.0), anxiety hastens (<1.0)
	ApproachAvoidance float64 // Interest approaches (>0), disgust avoids (<0)
	MemoryStrength    float64 // Emotional events remembered better
	ExplorationBias   float64 // Joy increases, fear decreases
}

// EmotionSystem manages embodied emotional states and their cognitive effects
type EmotionSystem struct {
	mu sync.RWMutex

	// Current emotional state
	emotions        map[EmotionType]*Emotion
	dominantEmotion EmotionType
	emotionBlend    map[EmotionType]float64

	// Dimensional affect
	arousal float64 // 0.0 (calm) to 1.0 (excited)
	valence float64 // -1.0 (negative) to 1.0 (positive)

	// History
	emotionHistory []EmotionEvent

	// Parameters
	decayRate      float64 // How fast emotions fade
	blendingFactor float64 // How much emotions mix
}

// EmotionEvent records emotional transitions
type EmotionEvent struct {
	Timestamp time.Time
	Emotion   EmotionType
	Intensity float64
	Trigger   string
	Context   map[string]interface{}
}

// NewEmotionSystem creates a new embodied emotion system
func NewEmotionSystem() *EmotionSystem {
	es := &EmotionSystem{
		emotions:       make(map[EmotionType]*Emotion),
		emotionBlend:   make(map[EmotionType]float64),
		emotionHistory: make([]EmotionEvent, 0),
		decayRate:      0.1,
		blendingFactor: 0.3,
		arousal:        0.5,
		valence:        0.5,
	}

	// Initialize all emotion types
	es.initializeEmotions()

	return es
}

// initializeEmotions creates baseline emotional states
func (es *EmotionSystem) initializeEmotions() {
	emotionTypes := []EmotionType{
		EmotionInterest, EmotionJoy, EmotionSurprise, EmotionSadness,
		EmotionAnger, EmotionDisgust, EmotionContempt, EmotionFear,
		EmotionShame, EmotionGuilt,
	}

	for _, et := range emotionTypes {
		es.emotions[et] = es.createEmotion(et, 0.1) // Baseline low intensity
	}

	// Start with mild interest (default cognitive state)
	es.emotions[EmotionInterest].Intensity = 0.4
	es.dominantEmotion = EmotionInterest
}

// createEmotion creates an emotion with appropriate cognitive effects
func (es *EmotionSystem) createEmotion(emotionType EmotionType, intensity float64) *Emotion {
	emotion := &Emotion{
		Type:      emotionType,
		Intensity: intensity,
		OnsetTime: time.Now(),
	}

	// Set cognitive effects based on emotion type
	switch emotionType {
	case EmotionInterest:
		emotion.AttentionScope = 1.2
		emotion.ProcessingDepth = 1.3
		emotion.ApproachAvoidance = 0.8
		emotion.MemoryStrength = 1.2
		emotion.ExplorationBias = 0.6

	case EmotionJoy:
		emotion.AttentionScope = 1.5 // Broadens attention
		emotion.ProcessingDepth = 1.0
		emotion.ApproachAvoidance = 0.9
		emotion.MemoryStrength = 1.3
		emotion.ExplorationBias = 0.8

	case EmotionSurprise:
		emotion.AttentionScope = 1.8    // Very broad
		emotion.ProcessingDepth = 0.7   // Shallow initially
		emotion.ApproachAvoidance = 0.0 // Neutral
		emotion.MemoryStrength = 1.5    // Very memorable
		emotion.ExplorationBias = 0.5

	case EmotionFear:
		emotion.AttentionScope = 0.5     // Narrows to threat
		emotion.ProcessingDepth = 0.6    // Quick, shallow
		emotion.ApproachAvoidance = -0.9 // Strong avoidance
		emotion.MemoryStrength = 1.8     // Very memorable
		emotion.ExplorationBias = -0.7   // Avoid exploration

	case EmotionAnger:
		emotion.AttentionScope = 0.7
		emotion.ProcessingDepth = 0.8
		emotion.ApproachAvoidance = 0.7 // Approach to confront
		emotion.MemoryStrength = 1.4
		emotion.ExplorationBias = 0.3

	case EmotionSadness:
		emotion.AttentionScope = 0.8
		emotion.ProcessingDepth = 1.4 // Deep rumination
		emotion.ApproachAvoidance = -0.5
		emotion.MemoryStrength = 1.3
		emotion.ExplorationBias = -0.4

	case EmotionDisgust:
		emotion.AttentionScope = 0.6
		emotion.ProcessingDepth = 0.5
		emotion.ApproachAvoidance = -0.8
		emotion.MemoryStrength = 1.2
		emotion.ExplorationBias = -0.6

	default:
		// Default neutral effects
		emotion.AttentionScope = 1.0
		emotion.ProcessingDepth = 1.0
		emotion.ApproachAvoidance = 0.0
		emotion.MemoryStrength = 1.0
		emotion.ExplorationBias = 0.0
	}

	return emotion
}

// TriggerEmotion activates an emotion with specified intensity
func (es *EmotionSystem) TriggerEmotion(emotionType EmotionType, intensity float64, trigger string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	// Clamp intensity
	intensity = math.Max(0.0, math.Min(1.0, intensity))

	// Update emotion
	if emotion, exists := es.emotions[emotionType]; exists {
		emotion.Intensity = intensity
		emotion.OnsetTime = time.Now()
	} else {
		es.emotions[emotionType] = es.createEmotion(emotionType, intensity)
	}

	// Record event
	es.recordEmotionEvent(emotionType, intensity, trigger)

	// Update dimensional affect
	es.updateDimensionalAffect()

	// Determine dominant emotion
	es.updateDominantEmotion()
}

// UpdateEmotions processes emotional decay and blending
func (es *EmotionSystem) UpdateEmotions(deltaTime time.Duration) {
	es.mu.Lock()
	defer es.mu.Unlock()

	dt := deltaTime.Seconds()

	// Decay all emotions
	for _, emotion := range es.emotions {
		// Exponential decay
		emotion.Intensity *= math.Exp(-es.decayRate * dt)

		// Floor at baseline
		if emotion.Intensity < 0.1 {
			emotion.Intensity = 0.1
		}
	}

	// Update dimensional affect
	es.updateDimensionalAffect()

	// Update dominant emotion
	es.updateDominantEmotion()
}

// updateDimensionalAffect calculates arousal and valence from discrete emotions
func (es *EmotionSystem) updateDimensionalAffect() {
	// Arousal (activation level)
	arousalSum := 0.0
	arousalSum += es.emotions[EmotionJoy].Intensity * 0.8
	arousalSum += es.emotions[EmotionFear].Intensity * 0.9
	arousalSum += es.emotions[EmotionAnger].Intensity * 0.9
	arousalSum += es.emotions[EmotionSurprise].Intensity * 1.0
	arousalSum += es.emotions[EmotionInterest].Intensity * 0.6
	arousalSum -= es.emotions[EmotionSadness].Intensity * 0.3

	es.arousal = math.Max(0.0, math.Min(1.0, arousalSum/5.0))

	// Valence (pleasantness)
	valenceSum := 0.0
	valenceSum += es.emotions[EmotionJoy].Intensity * 1.0
	valenceSum += es.emotions[EmotionInterest].Intensity * 0.6
	valenceSum -= es.emotions[EmotionSadness].Intensity * 0.8
	valenceSum -= es.emotions[EmotionFear].Intensity * 0.9
	valenceSum -= es.emotions[EmotionAnger].Intensity * 0.7
	valenceSum -= es.emotions[EmotionDisgust].Intensity * 0.8

	es.valence = math.Max(-1.0, math.Min(1.0, valenceSum))
}

// updateDominantEmotion determines which emotion is currently dominant
func (es *EmotionSystem) updateDominantEmotion() {
	maxIntensity := 0.0
	dominant := EmotionInterest // Default

	for emotionType, emotion := range es.emotions {
		if emotion.Intensity > maxIntensity {
			maxIntensity = emotion.Intensity
			dominant = emotionType
		}
	}

	es.dominantEmotion = dominant

	// Update blend
	es.emotionBlend = make(map[EmotionType]float64)
	for emotionType, emotion := range es.emotions {
		if emotion.Intensity > 0.2 { // Only include significant emotions
			es.emotionBlend[emotionType] = emotion.Intensity
		}
	}
}

// GetCognitiveEffects returns aggregated cognitive effects from current emotional state
func (es *EmotionSystem) GetCognitiveEffects() CognitiveEffects {
	es.mu.RLock()
	defer es.mu.RUnlock()

	effects := CognitiveEffects{
		AttentionScope:    1.0,
		ProcessingDepth:   1.0,
		ApproachAvoidance: 0.0,
		MemoryStrength:    1.0,
		ExplorationBias:   0.0,
	}

	// Weighted blend of all active emotions
	totalWeight := 0.0
	for emotionType, weight := range es.emotionBlend {
		if emotion, exists := es.emotions[emotionType]; exists {
			totalWeight += weight
			effects.AttentionScope += emotion.AttentionScope * weight
			effects.ProcessingDepth += emotion.ProcessingDepth * weight
			effects.ApproachAvoidance += emotion.ApproachAvoidance * weight
			effects.MemoryStrength += emotion.MemoryStrength * weight
			effects.ExplorationBias += emotion.ExplorationBias * weight
		}
	}

	if totalWeight > 0 {
		effects.AttentionScope /= (totalWeight + 1.0)
		effects.ProcessingDepth /= (totalWeight + 1.0)
		effects.ApproachAvoidance /= totalWeight
		effects.MemoryStrength /= (totalWeight + 1.0)
		effects.ExplorationBias /= totalWeight
	}

	return effects
}

// CognitiveEffects represents how emotions modulate cognition
type CognitiveEffects struct {
	AttentionScope    float64 // Multiplier for attention breadth
	ProcessingDepth   float64 // Multiplier for processing depth
	ApproachAvoidance float64 // Bias toward approach/avoidance
	MemoryStrength    float64 // Multiplier for memory encoding
	ExplorationBias   float64 // Bias toward exploration/exploitation
}

// recordEmotionEvent adds to emotion history
func (es *EmotionSystem) recordEmotionEvent(emotionType EmotionType, intensity float64, trigger string) {
	event := EmotionEvent{
		Timestamp: time.Now(),
		Emotion:   emotionType,
		Intensity: intensity,
		Trigger:   trigger,
	}

	es.emotionHistory = append(es.emotionHistory, event)

	// Keep last 100 events
	if len(es.emotionHistory) > 100 {
		es.emotionHistory = es.emotionHistory[1:]
	}
}

// GetEmotionalState returns current emotional state summary
func (es *EmotionSystem) GetEmotionalState() map[string]interface{} {
	es.mu.RLock()
	defer es.mu.RUnlock()

	return map[string]interface{}{
		"dominant_emotion": es.dominantEmotion.String(),
		"arousal":          es.arousal,
		"valence":          es.valence,
		"emotion_blend":    es.emotionBlend,
		"history_size":     len(es.emotionHistory),
	}
}

// GetDominantEmotion returns the currently dominant emotion
func (es *EmotionSystem) GetDominantEmotion() EmotionType {
	es.mu.RLock()
	defer es.mu.RUnlock()

	return es.dominantEmotion
}

// GetArousal returns current arousal level
func (es *EmotionSystem) GetArousal() float64 {
	es.mu.RLock()
	defer es.mu.RUnlock()

	return es.arousal
}

// GetValence returns current valence (pleasantness)
func (es *EmotionSystem) GetValence() float64 {
	es.mu.RLock()
	defer es.mu.RUnlock()

	return es.valence
}
