package live2d

import (
	"math"
	"sync"
)

// SuperHotGirlPresets defines emotional presets optimized for charismatic/attractive avatar behavior
var SuperHotGirlPresets = map[string]EmotionalState{
	"alluring": {
		Valence:    0.7,
		Arousal:    0.6,
		Dominance:  0.7,
		Curiosity:  0.5,
		Confidence: 0.85,
	},
	"playful": {
		Valence:    0.8,
		Arousal:    0.7,
		Dominance:  0.6,
		Curiosity:  0.7,
		Confidence: 0.75,
	},
	"mysterious": {
		Valence:    0.4,
		Arousal:    0.5,
		Dominance:  0.8,
		Curiosity:  0.6,
		Confidence: 0.9,
	},
	"seductive": {
		Valence:    0.6,
		Arousal:    0.7,
		Dominance:  0.75,
		Curiosity:  0.4,
		Confidence: 0.85,
	},
	"charming": {
		Valence:    0.85,
		Arousal:    0.6,
		Dominance:  0.65,
		Curiosity:  0.6,
		Confidence: 0.8,
	},
	"elegant": {
		Valence:    0.5,
		Arousal:    0.4,
		Dominance:  0.8,
		Curiosity:  0.4,
		Confidence: 0.9,
	},
	"flirtatious": {
		Valence:    0.75,
		Arousal:    0.65,
		Dominance:  0.6,
		Curiosity:  0.7,
		Confidence: 0.75,
	},
}

// EnhancedParameterNames extends standard parameters with advanced avatar controls
var EnhancedParameterNames = struct {
	// Hair dynamics (physics-based)
	HairFront        string
	HairSide         string
	HairBack         string
	HairSwayX        string
	HairSwayY        string
	HairBounce       string
	
	// Advanced eye expressions
	EyeSeductive     string
	EyeSparkle       string
	EyeGaze          string
	EyeHighlight     string
	
	// Mouth and lip expressions
	LipPout          string
	SmileCharm       string
	LipBite          string
	LipGloss         string
	
	// Body language and posture
	ShoulderTilt     string
	HipSway          string
	PostureConfidence string
	BodyCurve        string
	
	// Clothing dynamics
	ClothingWrinkle  string
	ClothingTension  string
	
	// Accessories
	EarringSwing     string
	NecklaceMove     string
	
	// Skin and blush
	BlushIntensity   string
	SkinGlow         string
	
	// Advanced breathing
	ChestBreath      string
	ShoulderBreath   string
}{
	HairFront:         "ParamHairFront",
	HairSide:          "ParamHairSide",
	HairBack:          "ParamHairBack",
	HairSwayX:         "ParamHairSwayX",
	HairSwayY:         "ParamHairSwayY",
	HairBounce:        "ParamHairBounce",
	
	EyeSeductive:      "ParamEyeSeductive",
	EyeSparkle:        "ParamEyeSparkle",
	EyeGaze:           "ParamEyeGaze",
	EyeHighlight:      "ParamEyeHighlight",
	
	LipPout:           "ParamLipPout",
	SmileCharm:        "ParamSmileCharm",
	LipBite:           "ParamLipBite",
	LipGloss:          "ParamLipGloss",
	
	ShoulderTilt:      "ParamShoulderTilt",
	HipSway:           "ParamHipSway",
	PostureConfidence: "ParamPostureConfidence",
	BodyCurve:         "ParamBodyCurve",
	
	ClothingWrinkle:   "ParamClothingWrinkle",
	ClothingTension:   "ParamClothingTension",
	
	EarringSwing:      "ParamEarringSwing",
	NecklaceMove:      "ParamNecklaceMove",
	
	BlushIntensity:    "ParamBlushIntensity",
	SkinGlow:          "ParamSkinGlow",
	
	ChestBreath:       "ParamChestBreath",
	ShoulderBreath:    "ParamShoulderBreath",
}

// EnhancedParameterMapper implements sophisticated parameter mapping for super-hot-girl avatars
type EnhancedParameterMapper struct {
	mu                sync.RWMutex
	charmIntensity    float64
	eleganceLevel     float64
	playfulnessLevel  float64
	mysteryLevel      float64
	breathingPhase    float64
	gazeTarget        GazeTarget
	expressionHistory []EmotionalState
	historySize       int
}

// GazeTarget represents where the avatar is looking
type GazeTarget struct {
	X, Y     float64 // -1.0 to 1.0
	Intensity float64 // 0.0 to 1.0
}

// NewEnhancedParameterMapper creates a new enhanced mapper
func NewEnhancedParameterMapper() *EnhancedParameterMapper {
	return &EnhancedParameterMapper{
		charmIntensity:    0.7,
		eleganceLevel:     0.6,
		playfulnessLevel:  0.5,
		mysteryLevel:      0.4,
		breathingPhase:    0.0,
		gazeTarget:        GazeTarget{X: 0.0, Y: 0.0, Intensity: 0.5},
		expressionHistory: make([]EmotionalState, 0),
		historySize:       10,
	}
}

// MapEmotionalState maps emotional state to enhanced parameters
func (epm *EnhancedParameterMapper) MapEmotionalState(state EmotionalState) []ModelParameter {
	epm.mu.Lock()
	defer epm.mu.Unlock()
	
	// Add to history
	epm.expressionHistory = append(epm.expressionHistory, state)
	if len(epm.expressionHistory) > epm.historySize {
		epm.expressionHistory = epm.expressionHistory[1:]
	}
	
	params := []ModelParameter{}
	
	// === Eyes ===
	// Eye openness based on arousal and valence
	eyeOpen := 0.8 + state.Arousal*0.2
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.EyeOpenLeft,
		Value: eyeOpen,
		Min:   0.0,
		Max:   1.0,
	})
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.EyeOpenRight,
		Value: eyeOpen,
		Min:   0.0,
		Max:   1.0,
	})
	
	// Eye smile based on valence
	eyeSmile := clamp(state.Valence*0.8, 0.0, 1.0)
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.EyeSmileLeft,
		Value: eyeSmile,
		Min:   0.0,
		Max:   1.0,
	})
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.EyeSmileRight,
		Value: eyeSmile,
		Min:   0.0,
		Max:   1.0,
	})
	
	// Seductive eye expression (half-lidded with high confidence)
	seductiveEye := clamp(state.Confidence*state.Dominance*0.6, 0.0, 1.0)
	params = append(params, ModelParameter{
		ID:    EnhancedParameterNames.EyeSeductive,
		Value: seductiveEye,
		Min:   0.0,
		Max:   1.0,
	})
	
	// Eye sparkle based on curiosity and arousal
	sparkle := clamp((state.Curiosity+state.Arousal)*0.5, 0.0, 1.0)
	params = append(params, ModelParameter{
		ID:    EnhancedParameterNames.EyeSparkle,
		Value: sparkle,
		Min:   0.0,
		Max:   1.0,
	})
	
	// === Mouth ===
	// Mouth openness
	mouthOpen := clamp(state.Arousal*0.3, 0.0, 1.0)
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.MouthOpenY,
		Value: mouthOpen,
		Min:   0.0,
		Max:   1.0,
	})
	
	// Smile charm (enhanced smile with confidence)
	smileCharm := clamp(state.Valence*state.Confidence*0.9, 0.0, 1.0)
	params = append(params, ModelParameter{
		ID:    EnhancedParameterNames.SmileCharm,
		Value: smileCharm,
		Min:   0.0,
		Max:   1.0,
	})
	
	// Lip pout (playful/seductive)
	lipPout := clamp((1.0-state.Dominance)*state.Confidence*0.5, 0.0, 1.0)
	params = append(params, ModelParameter{
		ID:    EnhancedParameterNames.LipPout,
		Value: lipPout,
		Min:   0.0,
		Max:   1.0,
	})
	
	// === Blush ===
	// Blush intensity based on arousal and valence
	blush := clamp(state.Arousal*0.7+state.Valence*0.3, 0.0, 1.0)
	params = append(params, ModelParameter{
		ID:    EnhancedParameterNames.BlushIntensity,
		Value: blush,
		Min:   0.0,
		Max:   1.0,
	})
	
	// Skin glow (confidence and positive valence)
	glow := clamp((state.Confidence+state.Valence)*0.4, 0.0, 1.0)
	params = append(params, ModelParameter{
		ID:    EnhancedParameterNames.SkinGlow,
		Value: glow,
		Min:   0.0,
		Max:   1.0,
	})
	
	// === Body Language ===
	// Posture confidence
	postureConf := clamp(state.Confidence*state.Dominance, 0.0, 1.0)
	params = append(params, ModelParameter{
		ID:    EnhancedParameterNames.PostureConfidence,
		Value: postureConf,
		Min:   0.0,
		Max:   1.0,
	})
	
	// Shoulder tilt (playful/flirtatious)
	shoulderTilt := math.Sin(epm.breathingPhase*0.5) * state.Arousal * 0.3
	params = append(params, ModelParameter{
		ID:    EnhancedParameterNames.ShoulderTilt,
		Value: shoulderTilt,
		Min:   -1.0,
		Max:   1.0,
	})
	
	// Hip sway (elegance and confidence)
	hipSway := math.Sin(epm.breathingPhase*0.7) * epm.eleganceLevel * 0.4
	params = append(params, ModelParameter{
		ID:    EnhancedParameterNames.HipSway,
		Value: hipSway,
		Min:   -1.0,
		Max:   1.0,
	})
	
	// === Hair Dynamics ===
	// Hair sway based on movement and arousal
	hairSwayX := math.Sin(epm.breathingPhase*1.2) * state.Arousal * 0.5
	hairSwayY := math.Cos(epm.breathingPhase*0.9) * state.Arousal * 0.3
	params = append(params, ModelParameter{
		ID:    EnhancedParameterNames.HairSwayX,
		Value: hairSwayX,
		Min:   -1.0,
		Max:   1.0,
	})
	params = append(params, ModelParameter{
		ID:    EnhancedParameterNames.HairSwayY,
		Value: hairSwayY,
		Min:   -1.0,
		Max:   1.0,
	})
	
	// Hair bounce (energy level)
	hairBounce := math.Abs(math.Sin(epm.breathingPhase*2.0)) * state.Arousal * 0.6
	params = append(params, ModelParameter{
		ID:    EnhancedParameterNames.HairBounce,
		Value: hairBounce,
		Min:   0.0,
		Max:   1.0,
	})
	
	// === Breathing ===
	epm.breathingPhase += 0.05 // Increment breathing phase
	breathValue := math.Sin(epm.breathingPhase) * 0.5
	
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.Breathing,
		Value: breathValue,
		Min:   -1.0,
		Max:   1.0,
	})
	
	// Chest breathing (more pronounced with arousal)
	chestBreath := breathValue * (0.5 + state.Arousal*0.5)
	params = append(params, ModelParameter{
		ID:    EnhancedParameterNames.ChestBreath,
		Value: chestBreath,
		Min:   -1.0,
		Max:   1.0,
	})
	
	return params
}

// MapCognitiveState maps cognitive state to parameters
func (epm *EnhancedParameterMapper) MapCognitiveState(state CognitiveState) []ModelParameter {
	epm.mu.Lock()
	defer epm.mu.Unlock()
	
	params := []ModelParameter{}
	
	// === Gaze and Attention ===
	// Eye gaze intensity based on attention
	gazeIntensity := state.Attention
	params = append(params, ModelParameter{
		ID:    EnhancedParameterNames.EyeGaze,
		Value: gazeIntensity,
		Min:   0.0,
		Max:   1.0,
	})
	
	// Eye position based on awareness
	eyeX := epm.gazeTarget.X * state.Awareness
	eyeY := epm.gazeTarget.Y * state.Awareness
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.EyeBallX,
		Value: eyeX,
		Min:   -1.0,
		Max:   1.0,
	})
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.EyeBallY,
		Value: eyeY,
		Min:   -1.0,
		Max:   1.0,
	})
	
	// === Head Movement ===
	// Head angle based on cognitive load and processing mode
	var headAngleX, headAngleY float64
	
	switch state.ProcessingMode {
	case "contemplative":
		headAngleX = -10.0 * state.CognitiveLoad // Look down when thinking
		headAngleY = 5.0 * (1.0 - state.Coherence) // Slight tilt
	case "dynamic":
		headAngleX = 5.0 * state.EnergyLevel
		headAngleY = math.Sin(epm.breathingPhase*0.5) * 10.0 * state.EnergyLevel
	case "cautious":
		headAngleX = 0.0
		headAngleY = -5.0 * (1.0 - state.Confidence)
	case "creative":
		headAngleX = math.Sin(epm.breathingPhase*0.3) * 8.0
		headAngleY = math.Cos(epm.breathingPhase*0.4) * 12.0
	default:
		headAngleX = 0.0
		headAngleY = 0.0
	}
	
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.AngleX,
		Value: headAngleX,
		Min:   -30.0,
		Max:   30.0,
	})
	params = append(params, ModelParameter{
		ID:    StandardParameterNames.AngleY,
		Value: headAngleY,
		Min:   -30.0,
		Max:   30.0,
	})
	
	// === Energy and Vitality ===
	// Posture based on energy level
	postureEnergy := state.EnergyLevel * 0.8
	params = append(params, ModelParameter{
		ID:    EnhancedParameterNames.PostureConfidence,
		Value: postureEnergy,
		Min:   0.0,
		Max:   1.0,
	})
	
	return params
}

// MapCombinedState maps both emotional and cognitive states
func (epm *EnhancedParameterMapper) MapCombinedState(state AvatarState) []ModelParameter {
	emotionalParams := epm.MapEmotionalState(state.Emotional)
	cognitiveParams := epm.MapCognitiveState(state.Cognitive)
	
	// Merge parameters, cognitive overrides emotional for conflicts
	paramMap := make(map[string]ModelParameter)
	
	for _, p := range emotionalParams {
		paramMap[p.ID] = p
	}
	
	for _, p := range cognitiveParams {
		paramMap[p.ID] = p // Override
	}
	
	// Convert back to slice
	result := make([]ModelParameter, 0, len(paramMap))
	for _, p := range paramMap {
		result = append(result, p)
	}
	
	return result
}

// SetGazeTarget sets where the avatar should look
func (epm *EnhancedParameterMapper) SetGazeTarget(x, y, intensity float64) {
	epm.mu.Lock()
	defer epm.mu.Unlock()
	
	epm.gazeTarget = GazeTarget{
		X:         clamp(x, -1.0, 1.0),
		Y:         clamp(y, -1.0, 1.0),
		Intensity: clamp(intensity, 0.0, 1.0),
	}
}

// SetCharmIntensity adjusts the overall charm level
func (epm *EnhancedParameterMapper) SetCharmIntensity(intensity float64) {
	epm.mu.Lock()
	defer epm.mu.Unlock()
	epm.charmIntensity = clamp(intensity, 0.0, 1.0)
}

// SetEleganceLevel adjusts elegance in movements
func (epm *EnhancedParameterMapper) SetEleganceLevel(level float64) {
	epm.mu.Lock()
	defer epm.mu.Unlock()
	epm.eleganceLevel = clamp(level, 0.0, 1.0)
}

// GetExpressionTrend analyzes recent emotional history
func (epm *EnhancedParameterMapper) GetExpressionTrend() (avgValence, avgArousal float64) {
	epm.mu.RLock()
	defer epm.mu.RUnlock()
	
	if len(epm.expressionHistory) == 0 {
		return 0.0, 0.0
	}
	
	totalValence := 0.0
	totalArousal := 0.0
	
	for _, state := range epm.expressionHistory {
		totalValence += state.Valence
		totalArousal += state.Arousal
	}
	
	count := float64(len(epm.expressionHistory))
	return totalValence / count, totalArousal / count
}
