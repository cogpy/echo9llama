package live2d

// Live2D Cubism SDK Integration for Echo9
// This package provides real-time avatar animation based on cognitive/emotional states

import (
	"sync"
	"time"
)

// ModelParameter represents a Live2D model parameter
type ModelParameter struct {
	ID    string  `json:"id"`
	Value float64 `json:"value"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
}

// ParameterUpdate represents a parameter change event
type ParameterUpdate struct {
	Timestamp  time.Time        `json:"timestamp"`
	Parameters []ModelParameter `json:"parameters"`
}

// EmotionalState represents the emotional state to map to avatar
type EmotionalState struct {
	Valence    float64 `json:"valence"`    // -1.0 (negative) to 1.0 (positive)
	Arousal    float64 `json:"arousal"`    // 0.0 (calm) to 1.0 (excited)
	Dominance  float64 `json:"dominance"`  // 0.0 (submissive) to 1.0 (dominant)
	Curiosity  float64 `json:"curiosity"`  // 0.0 to 1.0
	Confidence float64 `json:"confidence"` // 0.0 to 1.0
}

// CognitiveState represents cognitive processing state
type CognitiveState struct {
	Awareness      float64 `json:"awareness"`       // 0.0 to 1.0
	Attention      float64 `json:"attention"`       // 0.0 to 1.0
	CognitiveLoad  float64 `json:"cognitive_load"`  // 0.0 to 1.0
	Coherence      float64 `json:"coherence"`       // 0.0 to 1.0
	EnergyLevel    float64 `json:"energy_level"`    // 0.0 to 1.0
	Confidence     float64 `json:"confidence"`      // 0.0 to 1.0
	ProcessingMode string  `json:"processing_mode"` // "contemplative", "dynamic", "cautious", "creative"
}

// AvatarState combines emotional and cognitive states
type AvatarState struct {
	Emotional EmotionalState `json:"emotional"`
	Cognitive CognitiveState `json:"cognitive"`
	Timestamp time.Time      `json:"timestamp"`
}

// Live2DModel represents a Live2D Cubism model
type Live2DModel struct {
	mu          sync.RWMutex
	ModelPath   string                    `json:"model_path"`
	Name        string                    `json:"name"`
	Parameters  map[string]*ModelParameter `json:"parameters"`
	CurrentState AvatarState              `json:"current_state"`
	UpdateRate  time.Duration             `json:"update_rate"` // How often to update parameters
}

// ParameterMapper maps cognitive/emotional states to Live2D parameters
type ParameterMapper interface {
	// MapEmotionalState maps emotional state to parameter values
	MapEmotionalState(state EmotionalState) []ModelParameter
	
	// MapCognitiveState maps cognitive state to parameter values
	MapCognitiveState(state CognitiveState) []ModelParameter
	
	// MapCombinedState maps both states together
	MapCombinedState(state AvatarState) []ModelParameter
}

// StandardParameterNames defines common Live2D parameter IDs
var StandardParameterNames = struct {
	// Face expression parameters
	EyeOpenLeft      string
	EyeOpenRight     string
	EyeSmileLeft     string
	EyeSmileRight    string
	MouthOpenY       string
	MouthForm        string
	MouthSmile       string
	
	// Eye movement
	EyeBallX         string
	EyeBallY         string
	
	// Head movement
	AngleX           string
	AngleY           string
	AngleZ           string
	
	// Body
	BodyAngleX       string
	BodyAngleY       string
	BodyAngleZ       string
	
	// Breathing
	Breathing        string
}{
	// Standard Cubism parameter IDs
	EyeOpenLeft:   "ParamEyeLOpen",
	EyeOpenRight:  "ParamEyeROpen",
	EyeSmileLeft:  "ParamEyeLSmile",
	EyeSmileRight: "ParamEyeRSmile",
	MouthOpenY:    "ParamMouthOpenY",
	MouthForm:     "ParamMouthForm",
	MouthSmile:    "ParamMouthSmile",
	
	EyeBallX:      "ParamEyeBallX",
	EyeBallY:      "ParamEyeBallY",
	
	AngleX:        "ParamAngleX",
	AngleY:        "ParamAngleY",
	AngleZ:        "ParamAngleZ",
	
	BodyAngleX:    "ParamBodyAngleX",
	BodyAngleY:    "ParamBodyAngleY",
	BodyAngleZ:    "ParamBodyAngleZ",
	
	Breathing:     "ParamBreath",
}

// EmotionPresets defines preset parameter values for common emotions
var EmotionPresets = map[string]EmotionalState{
	"neutral": {
		Valence:    0.0,
		Arousal:    0.3,
		Dominance:  0.5,
		Curiosity:  0.3,
		Confidence: 0.5,
	},
	"happy": {
		Valence:    0.8,
		Arousal:    0.6,
		Dominance:  0.6,
		Curiosity:  0.4,
		Confidence: 0.7,
	},
	"sad": {
		Valence:    -0.6,
		Arousal:    0.2,
		Dominance:  0.3,
		Curiosity:  0.2,
		Confidence: 0.3,
	},
	"curious": {
		Valence:    0.3,
		Arousal:    0.5,
		Dominance:  0.4,
		Curiosity:  0.9,
		Confidence: 0.5,
	},
	"confident": {
		Valence:    0.5,
		Arousal:    0.5,
		Dominance:  0.8,
		Curiosity:  0.4,
		Confidence: 0.9,
	},
	"contemplative": {
		Valence:    0.2,
		Arousal:    0.3,
		Dominance:  0.5,
		Curiosity:  0.7,
		Confidence: 0.6,
	},
	"excited": {
		Valence:    0.7,
		Arousal:    0.9,
		Dominance:  0.7,
		Curiosity:  0.6,
		Confidence: 0.7,
	},
}
