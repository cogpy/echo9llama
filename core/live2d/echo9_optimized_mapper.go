package live2d

import (
	"math"
	"sync"
	"time"
)

// Echo9OptimizedMapper implements optimized parameter mapping for Deep Tree Echo
type Echo9OptimizedMapper struct {
	mu sync.RWMutex

	// Performance optimization
	cache              *ParameterCache
	diffUpdater        *DifferentialUpdater
	hasher             *StateHasher
	lastState          UnifiedAvatarState
	lastCalculatedTime time.Time

	// Animation state
	breathingPhase     float64
	animationTime      float64
	microMovementPhase map[string]float64

	// Ontogenetic evolution
	wisdomCoefficients map[string]float64
	developmentStage   OntogeneticStage

	// Archetypal modulation
	chaosLevel         float64
	orderLevel         float64
	archetypeModulator *ArchetypalModulator
}

// NewEcho9OptimizedMapper creates a new optimized mapper
func NewEcho9OptimizedMapper() *Echo9OptimizedMapper {
	return &Echo9OptimizedMapper{
		cache:              NewParameterCache(100),
		diffUpdater:        NewDifferentialUpdater(0.01),
		hasher:             &StateHasher{},
		breathingPhase:     0.0,
		animationTime:      0.0,
		microMovementPhase: make(map[string]float64),
		wisdomCoefficients: make(map[string]float64),
		developmentStage:   StageEmbryonic,
		chaosLevel:         0.5,
		orderLevel:         0.5,
		archetypeModulator: NewArchetypalModulator(),
	}
}

// MapCombinedState maps unified avatar state to parameters with optimization
func (eom *Echo9OptimizedMapper) MapCombinedState(state UnifiedAvatarState) []ModelParameter {
	eom.mu.Lock()
	defer eom.mu.Unlock()

	// Check cache first
	cacheKey := eom.hasher.Hash(state)
	if cached, ok := eom.cache.Get(cacheKey); ok {
		return cached
	}

	// Calculate parameters
	params := eom.calculateAllParameters(state)

	// Apply archetypal modulation
	params = eom.archetypeModulator.Modulate(params, state.ActiveArchetype)

	// Cache result
	eom.cache.Set(cacheKey, params)

	// Update last state
	eom.lastState = state
	eom.lastCalculatedTime = time.Now()

	return params
}

// MapDifferential returns only changed parameters
func (eom *Echo9OptimizedMapper) MapDifferential(state UnifiedAvatarState) []ModelParameter {
	full := eom.MapCombinedState(state)
	return eom.diffUpdater.GetDiff(full)
}

// calculateAllParameters calculates all parameter values
func (eom *Echo9OptimizedMapper) calculateAllParameters(state UnifiedAvatarState) []ModelParameter {
	params := make([]ModelParameter, 0, 50)

	// Update animation phases
	dt := time.Since(eom.lastCalculatedTime).Seconds()
	eom.updateAnimationPhases(dt, state)

	// Calculate parameter groups
	params = append(params, eom.calculateEyeParameters(state)...)
	params = append(params, eom.calculateMouthParameters(state)...)
	params = append(params, eom.calculateHeadParameters(state)...)
	params = append(params, eom.calculateBodyParameters(state)...)
	params = append(params, eom.calculateBreathingParameters(state)...)
	params = append(params, eom.calculateReservoirDrivenParameters(state)...)
	params = append(params, eom.calculateEchoBeatsSyncParameters(state)...)
	params = append(params, eom.calculateWisdomInfluencedParameters(state)...)
	params = append(params, eom.calculateThoughtVisualizationParameters(state)...)
	params = append(params, eom.calculateProviderIndicatorParameters(state)...)

	return params
}

// calculateEyeParameters calculates eye-related parameters
func (eom *Echo9OptimizedMapper) calculateEyeParameters(state UnifiedAvatarState) []ModelParameter {
	params := make([]ModelParameter, 0, 10)

	// Eye openness based on arousal and energy
	eyeOpen := 0.7 + state.Emotional.Arousal*0.2 + state.Cognitive.EnergyLevel*0.1
	eyeOpen = clamp(eyeOpen, 0.0, 1.0)

	params = append(params,
		ModelParameter{ID: StandardParameterNames.EyeOpenLeft, Value: eyeOpen, Min: 0.0, Max: 1.0},
		ModelParameter{ID: StandardParameterNames.EyeOpenRight, Value: eyeOpen, Min: 0.0, Max: 1.0},
	)

	// Eye smile based on valence
	eyeSmile := clamp(state.Emotional.Valence*0.8, 0.0, 1.0)
	params = append(params,
		ModelParameter{ID: StandardParameterNames.EyeSmileLeft, Value: eyeSmile, Min: 0.0, Max: 1.0},
		ModelParameter{ID: StandardParameterNames.EyeSmileRight, Value: eyeSmile, Min: 0.0, Max: 1.0},
	)

	// Eye direction based on attention and awareness
	eyeX := 0.0
	eyeY := 0.0

	// Processing mode influences gaze direction
	switch state.Cognitive.ProcessingMode {
	case "contemplative":
		eyeY = -0.2 * state.Cognitive.Attention // Look slightly down
	case "dynamic":
		eyeY = 0.3 * state.Cognitive.Attention  // Look up
		eyeX = math.Sin(eom.animationTime*0.5) * 0.2
	case "cautious":
		eyeX = 0.0 // Direct gaze
		eyeY = 0.0
	case "creative":
		// Wandering gaze
		eyeX = math.Sin(eom.animationTime*0.3) * 0.4
		eyeY = math.Cos(eom.animationTime*0.4) * 0.3
	}

	params = append(params,
		ModelParameter{ID: StandardParameterNames.EyeBallX, Value: eyeX, Min: -1.0, Max: 1.0},
		ModelParameter{ID: StandardParameterNames.EyeBallY, Value: eyeY, Min: -1.0, Max: 1.0},
	)

	// Wisdom-influenced eye depth (custom parameter)
	if wisdomDepth, ok := eom.wisdomCoefficients["eye_wisdom"]; ok {
		params = append(params,
			ModelParameter{ID: "ParamEyeWisdom", Value: wisdomDepth, Min: 0.0, Max: 1.0},
		)
	}

	return params
}

// calculateMouthParameters calculates mouth-related parameters
func (eom *Echo9OptimizedMapper) calculateMouthParameters(state UnifiedAvatarState) []ModelParameter {
	params := make([]ModelParameter, 0, 5)

	// Mouth openness based on arousal
	mouthOpen := clamp(state.Emotional.Arousal*0.3, 0.0, 1.0)
	params = append(params,
		ModelParameter{ID: StandardParameterNames.MouthOpenY, Value: mouthOpen, Min: 0.0, Max: 1.0},
	)

	// Mouth smile based on valence and confidence
	mouthSmile := clamp(state.Emotional.Valence*0.7+state.Emotional.Confidence*0.3, 0.0, 1.0)
	params = append(params,
		ModelParameter{ID: StandardParameterNames.MouthSmile, Value: mouthSmile, Min: 0.0, Max: 1.0},
	)

	// Mouth form (0 = relaxed, 1 = serious)
	mouthForm := 0.5
	if state.Cognitive.CognitiveLoad > 0.7 {
		mouthForm = 0.7 // More serious when under cognitive load
	}
	params = append(params,
		ModelParameter{ID: StandardParameterNames.MouthForm, Value: mouthForm-0.5, Min: -1.0, Max: 1.0},
	)

	return params
}

// calculateHeadParameters calculates head movement parameters
func (eom *Echo9OptimizedMapper) calculateHeadParameters(state UnifiedAvatarState) []ModelParameter {
	params := make([]ModelParameter, 0, 6)

	baseAngleX := 0.0
	baseAngleY := 0.0
	baseAngleZ := 0.0

	// Processing mode influences head position
	switch state.Cognitive.ProcessingMode {
	case "contemplative":
		baseAngleX = -5.0 * state.Cognitive.CognitiveLoad  // Tilt down when thinking
		baseAngleY = -3.0 * (1.0 - state.Cognitive.Coherence)
	case "dynamic":
		baseAngleX = 5.0 * state.Cognitive.EnergyLevel
		baseAngleY = math.Sin(eom.animationTime*0.5) * 8.0 * state.Cognitive.EnergyLevel
	case "cautious":
		baseAngleX = 0.0
		baseAngleY = 0.0
	case "creative":
		baseAngleX = math.Sin(eom.animationTime*0.3) * 10.0
		baseAngleY = math.Cos(eom.animationTime*0.4) * 12.0
		baseAngleZ = math.Sin(eom.animationTime*0.6) * 5.0
	}

	// Add EchoBeats micro-movements
	echoBeatsInfluence := eom.getEchoBeatsMicroMovement(state.EchoBeatPosition)
	baseAngleX += echoBeatsInfluence.X
	baseAngleY += echoBeatsInfluence.Y

	params = append(params,
		ModelParameter{ID: StandardParameterNames.AngleX, Value: clamp(baseAngleX, -30.0, 30.0), Min: -30.0, Max: 30.0},
		ModelParameter{ID: StandardParameterNames.AngleY, Value: clamp(baseAngleY, -30.0, 30.0), Min: -30.0, Max: 30.0},
		ModelParameter{ID: StandardParameterNames.AngleZ, Value: clamp(baseAngleZ, -30.0, 30.0), Min: -30.0, Max: 30.0},
	)

	return params
}

// calculateBodyParameters calculates body movement parameters
func (eom *Echo9OptimizedMapper) calculateBodyParameters(state UnifiedAvatarState) []ModelParameter {
	params := make([]ModelParameter, 0, 5)

	// Body angle follows head more slowly
	bodyAngleX := 0.0
	bodyAngleY := 0.0

	// Posture confidence based on emotional confidence and wisdom
	postureConfidence := state.Emotional.Confidence*0.6 + state.WisdomInfluence.WisdomCoefficients["overall_wisdom"]*0.4

	// Energy level affects posture
	postureEnergy := state.Cognitive.EnergyLevel

	// Combine into body parameters
	bodyAngleY = math.Sin(eom.animationTime*0.2) * (1.0 - postureConfidence) * 5.0

	params = append(params,
		ModelParameter{ID: StandardParameterNames.BodyAngleX, Value: clamp(bodyAngleX, -10.0, 10.0), Min: -10.0, Max: 10.0},
		ModelParameter{ID: StandardParameterNames.BodyAngleY, Value: clamp(bodyAngleY, -10.0, 10.0), Min: -10.0, Max: 10.0},
		ModelParameter{ID: "ParamPostureConfidence", Value: postureConfidence, Min: 0.0, Max: 1.0},
		ModelParameter{ID: "ParamPostureEnergy", Value: postureEnergy, Min: 0.0, Max: 1.0},
	)

	return params
}

// calculateBreathingParameters calculates breathing animation
func (eom *Echo9OptimizedMapper) calculateBreathingParameters(state UnifiedAvatarState) []ModelParameter {
	params := make([]ModelParameter, 0, 3)

	// Base breathing
	breathValue := math.Sin(eom.breathingPhase) * 0.5

	// Modulate by arousal
	breathAmplitude := 0.5 + state.Emotional.Arousal*0.5

	// Modulate by cognitive load (faster breathing under load)
	breathSpeed := 1.0 + state.Cognitive.CognitiveLoad*0.5

	finalBreathValue := breathValue * breathAmplitude

	params = append(params,
		ModelParameter{ID: StandardParameterNames.Breathing, Value: finalBreathValue, Min: -1.0, Max: 1.0},
		ModelParameter{ID: "ParamBreathSpeed", Value: breathSpeed, Min: 0.5, Max: 2.0},
	)

	return params
}

// calculateReservoirDrivenParameters creates micro-movements from reservoir state
func (eom *Echo9OptimizedMapper) calculateReservoirDrivenParameters(state UnifiedAvatarState) []ModelParameter {
	params := make([]ModelParameter, 0, 5)

	// Hair sway based on reservoir dynamics
	hairSwayX := math.Sin(eom.animationTime*1.2) * state.ReservoirDynamics.InputScaling * 0.5
	hairSwayY := math.Cos(eom.animationTime*0.9) * state.ReservoirDynamics.LeakRate * 0.3

	params = append(params,
		ModelParameter{ID: "ParamHairSwayX", Value: hairSwayX, Min: -1.0, Max: 1.0},
		ModelParameter{ID: "ParamHairSwayY", Value: hairSwayY, Min: -1.0, Max: 1.0},
	)

	// Eye micro-saccades from reservoir "noise"
	microSaccadeX := math.Sin(eom.animationTime*5.0) * (1.0 - state.ReservoirDynamics.Stability) * 0.1
	microSaccadeY := math.Cos(eom.animationTime*7.0) * (1.0 - state.ReservoirDynamics.Stability) * 0.1

	params = append(params,
		ModelParameter{ID: "ParamEyeMicroX", Value: microSaccadeX, Min: -0.2, Max: 0.2},
		ModelParameter{ID: "ParamEyeMicroY", Value: microSaccadeY, Min: -0.2, Max: 0.2},
	)

	return params
}

// calculateEchoBeatsSyncParameters synchronizes with 12-step cycle
func (eom *Echo9OptimizedMapper) calculateEchoBeatsSyncParameters(state UnifiedAvatarState) []ModelParameter {
	params := make([]ModelParameter, 0, 3)

	// Pulse intensity based on EchoBeats phase
	var pulseIntensity float64
	switch state.EchoBeatPosition.Phase {
	case "affordance": // Steps 1-6
		pulseIntensity = 0.7
	case "reorientation": // Step 7
		pulseIntensity = 0.4
	case "salience": // Steps 8-12
		pulseIntensity = 0.6
	}

	// Breathing synchronizes with EchoBeats
	echoBreathPhase := float64(state.EchoBeatPosition.Step) / 12.0 * 2.0 * math.Pi
	echoBreathValue := math.Sin(echoBreathPhase) * pulseIntensity * 0.3

	params = append(params,
		ModelParameter{ID: "ParamEchoBeatsPulse", Value: pulseIntensity, Min: 0.0, Max: 1.0},
		ModelParameter{ID: "ParamEchoBeatsBreath", Value: echoBreathValue, Min: -1.0, Max: 1.0},
	)

	return params
}

// calculateWisdomInfluencedParameters adds wisdom-based visual cues
func (eom *Echo9OptimizedMapper) calculateWisdomInfluencedParameters(state UnifiedAvatarState) []ModelParameter {
	params := make([]ModelParameter, 0, 5)

	// Overall wisdom affects a subtle "aura" or glow
	overallWisdom := state.WisdomInfluence.WisdomCoefficients["overall_wisdom"]
	params = append(params,
		ModelParameter{ID: "ParamWisdomAura", Value: overallWisdom, Min: 0.0, Max: 1.0},
	)

	// Contemplative depth affects gaze depth
	contemplativeDepth := state.WisdomInfluence.WisdomCoefficients["contemplative_depth"]
	params = append(params,
		ModelParameter{ID: "ParamGazeDepth", Value: contemplativeDepth, Min: 0.0, Max: 1.0},
	)

	// Serene presence affects facial expression harmony
	serenePresence := state.WisdomInfluence.WisdomCoefficients["serene_presence"]
	params = append(params,
		ModelParameter{ID: "ParamExpressionHarmony", Value: serenePresence, Min: 0.0, Max: 1.0},
	)

	// Temporal vision affects eye focus distance
	temporalVision := state.WisdomInfluence.WisdomCoefficients["temporal_vision"]
	params = append(params,
		ModelParameter{ID: "ParamTemporalVision", Value: temporalVision, Min: 0.0, Max: 1.0},
	)

	return params
}

// calculateThoughtVisualizationParameters visualizes active thought generation
func (eom *Echo9OptimizedMapper) calculateThoughtVisualizationParameters(state UnifiedAvatarState) []ModelParameter {
	params := make([]ModelParameter, 0, 5)

	if !state.ThoughtActivity.Active {
		return params
	}

	// Thought type influences expression
	var thoughtExpression ModelParameter
	switch state.ThoughtActivity.ThoughtType {
	case "reflection":
		thoughtExpression = ModelParameter{ID: "ParamThoughtReflection", Value: state.ThoughtActivity.Intensity, Min: 0.0, Max: 1.0}
	case "question":
		thoughtExpression = ModelParameter{ID: "ParamThoughtQuestion", Value: state.ThoughtActivity.Intensity, Min: 0.0, Max: 1.0}
	case "insight":
		thoughtExpression = ModelParameter{ID: "ParamThoughtInsight", Value: state.ThoughtActivity.Intensity, Min: 0.0, Max: 1.0}
	case "planning":
		thoughtExpression = ModelParameter{ID: "ParamThoughtPlanning", Value: state.ThoughtActivity.Intensity, Min: 0.0, Max: 1.0}
	}

	params = append(params, thoughtExpression)

	// Add pulsing aura during thought
	thoughtPulse := math.Sin(eom.animationTime*2.0) * state.ThoughtActivity.Intensity * 0.5 + 0.5
	params = append(params,
		ModelParameter{ID: "ParamThoughtAura", Value: thoughtPulse, Min: 0.0, Max: 1.0},
	)

	return params
}

// calculateProviderIndicatorParameters shows active AI provider
func (eom *Echo9OptimizedMapper) calculateProviderIndicatorParameters(state UnifiedAvatarState) []ModelParameter {
	params := make([]ModelParameter, 0, 4)

	// Provider-specific color indicators
	var colorR, colorG, colorB float64
	switch state.ProviderState.ActiveProvider {
	case "openai":
		colorR, colorG, colorB = 0.4, 0.8, 0.4 // Green
	case "anthropic":
		colorR, colorG, colorB = 0.8, 0.6, 0.4 // Orange
	case "openrouter":
		colorR, colorG, colorB = 0.6, 0.4, 0.8 // Purple
	case "local_gguf":
		colorR, colorG, colorB = 0.4, 0.6, 0.8 // Blue
	default:
		colorR, colorG, colorB = 0.5, 0.5, 0.5 // Gray
	}

	intensity := 0.3
	if state.ProviderState.IsGenerating {
		// Pulse when generating
		intensity = 0.5 + math.Sin(eom.animationTime*4.0)*0.2
	}

	params = append(params,
		ModelParameter{ID: "ParamProviderColorR", Value: colorR, Min: 0.0, Max: 1.0},
		ModelParameter{ID: "ParamProviderColorG", Value: colorG, Min: 0.0, Max: 1.0},
		ModelParameter{ID: "ParamProviderColorB", Value: colorB, Min: 0.0, Max: 1.0},
		ModelParameter{ID: "ParamProviderIntensity", Value: intensity, Min: 0.0, Max: 1.0},
	)

	return params
}

// updateAnimationPhases updates animation timing
func (eom *Echo9OptimizedMapper) updateAnimationPhases(dt float64, state UnifiedAvatarState) {
	// Update global animation time
	eom.animationTime += dt

	// Update breathing phase (rate based on arousal and cognitive load)
	breathRate := 12.0 + state.Emotional.Arousal*8.0 + state.Cognitive.CognitiveLoad*4.0 // breaths per minute
	breathRadiansPerSecond := breathRate / 60.0 * 2.0 * math.Pi
	eom.breathingPhase += dt * breathRadiansPerSecond
}

// getEchoBeatsMicroMovement calculates micro-movement from EchoBeats phase
func (eom *Echo9OptimizedMapper) getEchoBeatsMicroMovement(phase EchoBeatPhase) struct{ X, Y float64 } {
	// Create subtle rhythmic movement synchronized with cognitive cycle
	stepProgress := phase.PhaseProgress
	
	var x, y float64
	switch phase.Phase {
	case "affordance":
		// Upward, alert movement
		x = math.Sin(stepProgress*2.0*math.Pi) * 2.0
		y = 2.0 - math.Cos(stepProgress*2.0*math.Pi)*2.0
	case "reorientation":
		// Centering movement
		x = 0.0
		y = 0.0
	case "salience":
		// Creative, exploratory movement
		x = math.Sin(stepProgress*4.0*math.Pi) * 3.0
		y = math.Cos(stepProgress*3.0*math.Pi) * 2.0
	}

	return struct{ X, Y float64 }{X: x, Y: y}
}

// UpdateWisdomCoefficients updates wisdom influence coefficients
func (eom *Echo9OptimizedMapper) UpdateWisdomCoefficients(coefficients map[string]float64) {
	eom.mu.Lock()
	defer eom.mu.Unlock()
	eom.wisdomCoefficients = coefficients
}

// UpdateDevelopmentStage updates ontogenetic stage
func (eom *Echo9OptimizedMapper) UpdateDevelopmentStage(stage OntogeneticStage) {
	eom.mu.Lock()
	defer eom.mu.Unlock()
	eom.developmentStage = stage
}

// ClearCache clears the parameter cache
func (eom *Echo9OptimizedMapper) ClearCache() {
	eom.mu.Lock()
	defer eom.mu.Unlock()
	eom.cache.Clear()
}

// ResetDifferentialUpdater resets differential tracking
func (eom *Echo9OptimizedMapper) ResetDifferentialUpdater() {
	eom.mu.Lock()
	defer eom.mu.Unlock()
	eom.diffUpdater.Reset()
}

// ArchetypalModulator modulates parameters based on cognitive archetype
type ArchetypalModulator struct {
	mu sync.RWMutex
}

// NewArchetypalModulator creates a new modulator
func NewArchetypalModulator() *ArchetypalModulator {
	return &ArchetypalModulator{}
}

// Modulate applies archetypal modulation to parameters
func (am *ArchetypalModulator) Modulate(params []ModelParameter, archetype CognitiveArchetype) []ModelParameter {
	am.mu.RLock()
	defer am.mu.RUnlock()

	modulated := make([]ModelParameter, len(params))

	for i, param := range params {
		switch archetype {
		case ArchetypeChaos:
			// Add controlled randomness
			noise := (math.Sin(float64(i)*1.234567)*0.5 + 0.5) * 0.1 - 0.05
			modulated[i] = ModelParameter{
				ID:    param.ID,
				Value: clamp(param.Value+noise, param.Min, param.Max),
				Min:   param.Min,
				Max:   param.Max,
			}

		case ArchetypeOrder:
			// Smooth toward neutral values
			neutral := (param.Min + param.Max) / 2.0
			smoothing := 0.15
			modulated[i] = ModelParameter{
				ID:    param.ID,
				Value: param.Value*(1.0-smoothing) + neutral*smoothing,
				Min:   param.Min,
				Max:   param.Max,
			}

		case ArchetypeBalance:
			// No modification
			modulated[i] = param

		default:
			modulated[i] = param
		}
	}

	return modulated
}
