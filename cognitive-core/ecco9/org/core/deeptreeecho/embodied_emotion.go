//go:build orgdte
// +build orgdte

package deeptreeecho

import (
	"math"
	"sync"
	"time"
)

// DiscreteEmotionType represents categorical emotion types based on
// Izard's Differential Emotion Theory
type DiscreteEmotionType int

const (
	EmotionInterest DiscreteEmotionType = iota
	EmotionJoy
	EmotionSurprise
	EmotionSadness
	EmotionAnger
	EmotionDisgust
	EmotionContempt
	EmotionFear
	EmotionShame
	EmotionGuilt
	EmotionCuriosity
)

func (et DiscreteEmotionType) String() string {
	names := []string{
		"Interest", "Joy", "Surprise", "Sadness", "Anger",
		"Disgust", "Contempt", "Fear", "Shame", "Guilt", "Curiosity",
	}
	if int(et) < len(names) {
		return names[et]
	}
	return "Unknown"
}

// DiscreteEmotion represents a discrete emotional state with cognitive effects.
type DiscreteEmotion struct {
	Type              DiscreteEmotionType
	Intensity         float64
	Duration          time.Duration
	OnsetTime         time.Time
	AttentionScope    float64
	ProcessingDepth   float64
	ApproachAvoidance float64
	MemoryStrength    float64
	ExplorationBias   float64
}

// EmotionSystem manages embodied emotional states and their cognitive effects.
type EmotionSystem struct {
	mu              sync.RWMutex
	emotions        map[DiscreteEmotionType]*DiscreteEmotion
	dominantEmotion DiscreteEmotionType
	emotionBlend    map[DiscreteEmotionType]float64
	arousal         float64
	valence         float64
	emotionHistory  []EmotionEvent
	decayRate       float64
	blendingFactor  float64
}

type EmotionEvent struct {
	Timestamp time.Time
	Emotion   DiscreteEmotionType
	Intensity float64
	Trigger   string
	Context   map[string]interface{}
}

type CognitiveEffects struct {
	AttentionScope    float64
	ProcessingDepth   float64
	ApproachAvoidance float64
	MemoryStrength    float64
	ExplorationBias   float64
}

func NewEmotionSystem() *EmotionSystem {
	es := &EmotionSystem{
		emotions:       make(map[DiscreteEmotionType]*DiscreteEmotion),
		emotionBlend:   make(map[DiscreteEmotionType]float64),
		emotionHistory: make([]EmotionEvent, 0),
		decayRate:      0.1,
		blendingFactor: 0.3,
		arousal:        0.5,
		valence:        0.5,
	}
	es.initializeEmotions()
	return es
}

func (es *EmotionSystem) initializeEmotions() {
	for _, et := range []DiscreteEmotionType{
		EmotionInterest, EmotionJoy, EmotionSurprise, EmotionSadness,
		EmotionAnger, EmotionDisgust, EmotionContempt, EmotionFear,
		EmotionShame, EmotionGuilt, EmotionCuriosity,
	} {
		es.emotions[et] = es.createDiscreteEmotion(et, 0.1)
	}
	es.emotions[EmotionInterest].Intensity = 0.4
	es.dominantEmotion = EmotionInterest
}

func (es *EmotionSystem) createDiscreteEmotion(emotionType DiscreteEmotionType, intensity float64) *DiscreteEmotion {
	emotion := &DiscreteEmotion{Type: emotionType, Intensity: intensity, OnsetTime: time.Now()}
	switch emotionType {
	case EmotionInterest:
		emotion.AttentionScope, emotion.ProcessingDepth, emotion.ApproachAvoidance = 1.2, 1.3, 0.8
		emotion.MemoryStrength, emotion.ExplorationBias = 1.2, 0.6
	case EmotionJoy:
		emotion.AttentionScope, emotion.ProcessingDepth, emotion.ApproachAvoidance = 1.5, 1.0, 0.9
		emotion.MemoryStrength, emotion.ExplorationBias = 1.3, 0.8
	case EmotionSurprise:
		emotion.AttentionScope, emotion.ProcessingDepth, emotion.ApproachAvoidance = 1.8, 0.7, 0.0
		emotion.MemoryStrength, emotion.ExplorationBias = 1.5, 0.5
	case EmotionFear:
		emotion.AttentionScope, emotion.ProcessingDepth, emotion.ApproachAvoidance = 0.5, 0.6, -0.9
		emotion.MemoryStrength, emotion.ExplorationBias = 1.8, -0.7
	case EmotionAnger:
		emotion.AttentionScope, emotion.ProcessingDepth, emotion.ApproachAvoidance = 0.7, 0.8, 0.7
		emotion.MemoryStrength, emotion.ExplorationBias = 1.4, 0.3
	case EmotionSadness:
		emotion.AttentionScope, emotion.ProcessingDepth, emotion.ApproachAvoidance = 0.8, 1.4, -0.5
		emotion.MemoryStrength, emotion.ExplorationBias = 1.3, -0.4
	case EmotionDisgust:
		emotion.AttentionScope, emotion.ProcessingDepth, emotion.ApproachAvoidance = 0.6, 0.5, -0.8
		emotion.MemoryStrength, emotion.ExplorationBias = 1.2, -0.6
	case EmotionCuriosity:
		emotion.AttentionScope, emotion.ProcessingDepth, emotion.ApproachAvoidance = 1.6, 1.5, 0.9
		emotion.MemoryStrength, emotion.ExplorationBias = 1.4, 0.9
	default:
		emotion.AttentionScope, emotion.ProcessingDepth, emotion.ApproachAvoidance = 1.0, 1.0, 0.0
		emotion.MemoryStrength, emotion.ExplorationBias = 1.0, 0.0
	}
	return emotion
}

func (es *EmotionSystem) TriggerEmotion(emotionType DiscreteEmotionType, intensity float64, trigger string) {
	es.mu.Lock()
	defer es.mu.Unlock()
	intensity = math.Max(0.0, math.Min(1.0, intensity))
	if emotion, exists := es.emotions[emotionType]; exists {
		emotion.Intensity = intensity
		emotion.OnsetTime = time.Now()
	} else {
		es.emotions[emotionType] = es.createDiscreteEmotion(emotionType, intensity)
	}
	es.recordEmotionEvent(emotionType, intensity, trigger)
	es.updateDimensionalAffect()
	es.updateDominantEmotion()
}

func (es *EmotionSystem) UpdateEmotions(deltaTime time.Duration) {
	es.mu.Lock()
	defer es.mu.Unlock()
	dt := deltaTime.Seconds()
	for _, emotion := range es.emotions {
		emotion.Intensity *= math.Exp(-es.decayRate * dt)
		if emotion.Intensity < 0.1 {
			emotion.Intensity = 0.1
		}
	}
	es.updateDimensionalAffect()
	es.updateDominantEmotion()
}

func (es *EmotionSystem) updateDimensionalAffect() {
	arousalSum := es.emotions[EmotionJoy].Intensity*0.8 + es.emotions[EmotionFear].Intensity*0.9 +
		es.emotions[EmotionAnger].Intensity*0.9 + es.emotions[EmotionSurprise].Intensity*1.0 +
		es.emotions[EmotionInterest].Intensity*0.6 + es.emotions[EmotionCuriosity].Intensity*0.7 -
		es.emotions[EmotionSadness].Intensity*0.3
	es.arousal = math.Max(0.0, math.Min(1.0, arousalSum/6.0))

	valenceSum := es.emotions[EmotionJoy].Intensity*1.0 + es.emotions[EmotionInterest].Intensity*0.6 +
		es.emotions[EmotionCuriosity].Intensity*0.7 - es.emotions[EmotionSadness].Intensity*0.8 -
		es.emotions[EmotionFear].Intensity*0.9 - es.emotions[EmotionAnger].Intensity*0.7 -
		es.emotions[EmotionDisgust].Intensity*0.8
	es.valence = math.Max(-1.0, math.Min(1.0, valenceSum))
}

func (es *EmotionSystem) updateDominantEmotion() {
	maxIntensity := 0.0
	dominant := EmotionInterest
	for emotionType, emotion := range es.emotions {
		if emotion.Intensity > maxIntensity {
			maxIntensity = emotion.Intensity
			dominant = emotionType
		}
	}
	es.dominantEmotion = dominant
	es.emotionBlend = make(map[DiscreteEmotionType]float64)
	for emotionType, emotion := range es.emotions {
		if emotion.Intensity > 0.2 {
			es.emotionBlend[emotionType] = emotion.Intensity
		}
	}
}

func (es *EmotionSystem) GetCognitiveEffects() CognitiveEffects {
	es.mu.RLock()
	defer es.mu.RUnlock()
	effects := CognitiveEffects{AttentionScope: 1.0, ProcessingDepth: 1.0, MemoryStrength: 1.0}
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

func (es *EmotionSystem) recordEmotionEvent(emotionType DiscreteEmotionType, intensity float64, trigger string) {
	es.emotionHistory = append(es.emotionHistory, EmotionEvent{
		Timestamp: time.Now(), Emotion: emotionType, Intensity: intensity, Trigger: trigger,
	})
	if len(es.emotionHistory) > 100 {
		es.emotionHistory = es.emotionHistory[1:]
	}
}

func (es *EmotionSystem) GetEmotionalState() map[string]interface{} {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return map[string]interface{}{
		"dominant_emotion": es.dominantEmotion.String(),
		"arousal":          es.arousal, "valence": es.valence,
		"emotion_blend": es.emotionBlend, "history_size": len(es.emotionHistory),
	}
}

func (es *EmotionSystem) GetDominantEmotion() DiscreteEmotionType {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.dominantEmotion
}

func (es *EmotionSystem) GetArousal() float64 {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.arousal
}

func (es *EmotionSystem) GetValence() float64 {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.valence
}
