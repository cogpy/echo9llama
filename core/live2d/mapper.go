package live2d

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// DefaultParameterMapper implements the ParameterMapper interface
// Maps Echo9's emotional and cognitive states to Live2D parameters
type DefaultParameterMapper struct {
	mu              sync.RWMutex
	smoothingFactor float64 // For parameter smoothing (0.0 = no smoothing, 1.0 = max smoothing)
	previousState   *AvatarState
}

// NewDefaultParameterMapper creates a new parameter mapper
func NewDefaultParameterMapper() *DefaultParameterMapper {
	return &DefaultParameterMapper{
		smoothingFactor: 0.3, // Smooth transitions
		previousState:   nil,
	}
}

// MapEmotionalState maps emotional state to Live2D parameters
func (m *DefaultParameterMapper) MapEmotionalState(state EmotionalState) []ModelParameter {
	params := []ModelParameter{}

	// Map valence and arousal to facial expressions
	// Positive valence -> smile
	smileIntensity := math.Max(0, state.Valence) // 0 to 1
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.MouthSmile,
		Value: smileIntensity,
		Min:   0.0,
		Max:   1.0,
	})
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.EyeSmileLeft,
		Value: smileIntensity * 0.8,
		Min:   0.0,
		Max:   1.0,
	})
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.EyeSmileRight,
		Value: smileIntensity * 0.8,
		Min:   0.0,
		Max:   1.0,
	})

	// Arousal affects eye openness and mouth
	eyeOpenness := 0.5 + (state.Arousal * 0.5) // 0.5 to 1.0
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.EyeOpenLeft,
		Value: eyeOpenness,
		Min:   0.0,
		Max:   1.0,
	})
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.EyeOpenRight,
		Value: eyeOpenness,
		Min:   0.0,
		Max:   1.0,
	})

	// Curiosity affects eye direction and head tilt
	if state.Curiosity > 0.6 {
		// Slight head tilt when curious
		params = append(params, ModelParameter{
			ID:    StandardParameterNames.AngleX,
			Value: (state.Curiosity - 0.6) * 20, // -10 to 10 degrees
			Min:   -30.0,
			Max:   30.0,
		})
	}

	// Confidence affects posture
	bodyAngle := (state.Confidence - 0.5) * 10 // -5 to 5 degrees
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.BodyAngleX,
		Value: bodyAngle,
		Min:   -30.0,
		Max:   30.0,
	})

	return params
}

// MapCognitiveState maps cognitive state to Live2D parameters
func (m *DefaultParameterMapper) MapCognitiveState(state CognitiveState) []ModelParameter {
	params := []ModelParameter{}

	// Energy level affects breathing rate
	breathingRate := 0.5 + (state.EnergyLevel * 0.5) // 0.5 to 1.0
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.Breathing,
		Value: breathingRate,
		Min:   0.0,
		Max:   1.0,
	})

	// Cognitive load affects eye blink rate (simulated)
	// High cognitive load = more frequent blinking
	blinkIntensity := 1.0 - (state.CognitiveLoad * 0.3) // 0.7 to 1.0
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.EyeOpenLeft,
		Value: blinkIntensity,
		Min:   0.0,
		Max:   1.0,
	})
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.EyeOpenRight,
		Value: blinkIntensity,
		Min:   0.0,
		Max:   1.0,
	})

	// Awareness affects gaze direction
	// Higher awareness = more direct gaze
	gazeDirectness := state.Awareness
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.EyeBallX,
		Value: (1.0 - gazeDirectness) * 5, // 0 to 5
		Min:   -30.0,
		Max:   30.0,
	})

	// Processing mode affects head position
	var headAngleY float64
	switch state.ProcessingMode {
	case "contemplative":
		headAngleY = -5.0 // Slight downward
	case "dynamic":
		headAngleY = 5.0 // Slight upward
	case "cautious":
		headAngleY = 0.0 // Neutral
	case "creative":
		headAngleY = 3.0 // Slight upward
	default:
		headAngleY = 0.0
	}
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.AngleY,
		Value: headAngleY,
		Min:   -30.0,
		Max:   30.0,
	})

	return params
}

// MapCombinedState maps both emotional and cognitive states
func (m *DefaultParameterMapper) MapCombinedState(state AvatarState) []ModelParameter {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get parameters from both mappers
	emotionalParams := m.MapEmotionalState(state.Emotional)
	cognitiveParams := m.MapCognitiveState(state.Cognitive)

	// Merge parameters - cognitive takes precedence for conflicts
	paramMap := make(map[string]ModelParameter)

	for _, param := range emotionalParams {
		paramMap[param.ID] = param
	}

	for _, param := range cognitiveParams {
		if existing, ok := paramMap[param.ID]; ok {
			// Blend the values
			blended := (existing.Value + param.Value) / 2.0
			param.Value = blended
		}
		paramMap[param.ID] = param
	}

	// Apply smoothing if we have a previous state
	if m.previousState != nil && m.smoothingFactor > 0 {
		// Apply exponential smoothing
		for id, param := range paramMap {
			// Smooth toward new value
			smoothed := param
			smoothed.Value = param.Value * (1.0 - m.smoothingFactor)
			paramMap[id] = smoothed
		}
	}

	// Store current state for next smoothing
	stateCopy := state
	m.previousState = &stateCopy

	// Convert map to slice
	result := make([]ModelParameter, 0, len(paramMap))
	for _, param := range paramMap {
		result = append(result, param)
	}

	return result
}

// SetSmoothingFactor sets the smoothing factor (0.0 to 1.0)
func (m *DefaultParameterMapper) SetSmoothingFactor(factor float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if factor < 0.0 {
		factor = 0.0
	} else if factor > 1.0 {
		factor = 1.0
	}

	m.smoothingFactor = factor
}

// clamp ensures value is within min and max bounds
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// NewLive2DModel creates a new Live2D model instance
func NewLive2DModel(name, modelPath string) *Live2DModel {
	return &Live2DModel{
		Name:       name,
		ModelPath:  modelPath,
		Parameters: make(map[string]*ModelParameter),
		CurrentState: AvatarState{
			Emotional: EmotionPresets["neutral"],
			Cognitive: CognitiveState{
				Awareness:      0.5,
				Attention:      0.5,
				CognitiveLoad:  0.3,
				Coherence:      0.7,
				EnergyLevel:    0.7,
				ProcessingMode: "contemplative",
			},
			Timestamp: time.Now(),
		},
		UpdateRate: 16 * time.Millisecond, // ~60 FPS
	}
}

// UpdateState updates the model's current state
func (m *Live2DModel) UpdateState(state AvatarState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state.Timestamp = time.Now()
	m.CurrentState = state

	return nil
}

// GetCurrentParameters returns the current parameter values
func (m *Live2DModel) GetCurrentParameters(mapper ParameterMapper) []ModelParameter {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return mapper.MapCombinedState(m.CurrentState)
}

// GetCurrentState returns the current avatar state
func (m *Live2DModel) GetCurrentState() AvatarState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.CurrentState
}

// String returns a string representation of the model
func (m *Live2DModel) String() string {
	return fmt.Sprintf("Live2DModel{Name: %s, Path: %s}", m.Name, m.ModelPath)
}
