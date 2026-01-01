package live2d

import (
	"math"
	"sync"
	"time"
)

// EnvironmentState represents the virtual environment's current state
type EnvironmentState struct {
	LightingIntensity  float64 `json:"lighting_intensity"`  // 0.0 to 1.0
	LightingColor      Color   `json:"lighting_color"`      // RGB color
	ParticleIntensity  float64 `json:"particle_intensity"`  // 0.0 to 1.0
	ParticleType       string  `json:"particle_type"`       // "sparkles", "energy", "thoughts", "emotions"
	AudioAmbience      float64 `json:"audio_ambience"`      // 0.0 to 1.0
	RealityDistortion  float64 `json:"reality_distortion"`  // 0.0 to 1.0 (for deep thought effects)
	EnvironmentMood    string  `json:"environment_mood"`    // "calm", "energetic", "mysterious", "contemplative"
}

// Color represents an RGB color
type Color struct {
	R, G, B float64 // 0.0 to 1.0
}

// CognitiveEnvironmentCoupler manages the bidirectional coupling between
// avatar cognitive state and virtual environment
type CognitiveEnvironmentCoupler struct {
	mu                  sync.RWMutex
	currentEnvState     EnvironmentState
	avatarState         AvatarState
	couplingStrength    float64 // 0.0 to 1.0
	updateRate          time.Duration
	running             bool
	stopChan            chan struct{}
	environmentCallback func(EnvironmentState)
}

// NewCognitiveEnvironmentCoupler creates a new environment coupler
func NewCognitiveEnvironmentCoupler(couplingStrength float64) *CognitiveEnvironmentCoupler {
	return &CognitiveEnvironmentCoupler{
		currentEnvState: EnvironmentState{
			LightingIntensity: 0.7,
			LightingColor:     Color{R: 1.0, G: 1.0, B: 1.0},
			ParticleIntensity: 0.3,
			ParticleType:      "sparkles",
			AudioAmbience:     0.5,
			RealityDistortion: 0.0,
			EnvironmentMood:   "calm",
		},
		couplingStrength: clamp(couplingStrength, 0.0, 1.0),
		updateRate:       50 * time.Millisecond, // 20 Hz
		running:          false,
		stopChan:         make(chan struct{}),
	}
}

// Start begins the environment coupling update loop
func (cec *CognitiveEnvironmentCoupler) Start() {
	cec.mu.Lock()
	if cec.running {
		cec.mu.Unlock()
		return
	}
	cec.running = true
	cec.mu.Unlock()
	
	go cec.updateLoop()
}

// Stop halts the environment coupling
func (cec *CognitiveEnvironmentCoupler) Stop() {
	cec.mu.Lock()
	defer cec.mu.Unlock()
	
	if !cec.running {
		return
	}
	
	cec.running = false
	close(cec.stopChan)
}

// SetEnvironmentCallback sets the callback function for environment updates
func (cec *CognitiveEnvironmentCoupler) SetEnvironmentCallback(callback func(EnvironmentState)) {
	cec.mu.Lock()
	defer cec.mu.Unlock()
	cec.environmentCallback = callback
}

// UpdateAvatarState updates the avatar state and triggers environment response
func (cec *CognitiveEnvironmentCoupler) UpdateAvatarState(state AvatarState) {
	cec.mu.Lock()
	cec.avatarState = state
	cec.mu.Unlock()
}

// GetEnvironmentState returns the current environment state
func (cec *CognitiveEnvironmentCoupler) GetEnvironmentState() EnvironmentState {
	cec.mu.RLock()
	defer cec.mu.RUnlock()
	return cec.currentEnvState
}

// updateLoop continuously updates environment based on avatar state
func (cec *CognitiveEnvironmentCoupler) updateLoop() {
	ticker := time.NewTicker(cec.updateRate)
	defer ticker.Stop()
	
	for {
		select {
		case <-cec.stopChan:
			return
		case <-ticker.C:
			cec.computeEnvironmentResponse()
		}
	}
}

// computeEnvironmentResponse calculates environment state from avatar state
func (cec *CognitiveEnvironmentCoupler) computeEnvironmentResponse() {
	cec.mu.Lock()
	defer cec.mu.Unlock()
	
	emotional := cec.avatarState.Emotional
	cognitive := cec.avatarState.Cognitive
	strength := cec.couplingStrength
	
	// === Lighting Response ===
	// Intensity based on arousal and energy
	targetIntensity := 0.5 + (emotional.Arousal*0.3 + cognitive.EnergyLevel*0.2) * strength
	cec.currentEnvState.LightingIntensity = lerp(
		cec.currentEnvState.LightingIntensity,
		targetIntensity,
		0.1,
	)
	
	// Color based on valence (warm = positive, cool = negative)
	if emotional.Valence > 0 {
		// Warm colors (yellow-orange)
		targetColor := Color{
			R: 1.0,
			G: 0.8 + emotional.Valence*0.2,
			B: 0.6 + emotional.Valence*0.2,
		}
		cec.currentEnvState.LightingColor = lerpColor(
			cec.currentEnvState.LightingColor,
			targetColor,
			0.05 * strength,
		)
	} else {
		// Cool colors (blue-cyan)
		targetColor := Color{
			R: 0.6 - emotional.Valence*0.3,
			G: 0.7 - emotional.Valence*0.2,
			B: 1.0,
		}
		cec.currentEnvState.LightingColor = lerpColor(
			cec.currentEnvState.LightingColor,
			targetColor,
			0.05 * strength,
		)
	}
	
	// === Particle Effects ===
	// Intensity based on cognitive activity
	targetParticleIntensity := (cognitive.Attention*0.4 + cognitive.Awareness*0.3 + emotional.Curiosity*0.3) * strength
	cec.currentEnvState.ParticleIntensity = lerp(
		cec.currentEnvState.ParticleIntensity,
		targetParticleIntensity,
		0.08,
	)
	
	// Particle type based on processing mode
	switch cognitive.ProcessingMode {
	case "contemplative":
		cec.currentEnvState.ParticleType = "thoughts"
	case "dynamic":
		cec.currentEnvState.ParticleType = "energy"
	case "creative":
		cec.currentEnvState.ParticleType = "sparkles"
	default:
		cec.currentEnvState.ParticleType = "emotions"
	}
	
	// === Audio Ambience ===
	// Volume based on arousal and cognitive load
	targetAudio := (emotional.Arousal*0.5 + cognitive.CognitiveLoad*0.5) * strength
	cec.currentEnvState.AudioAmbience = lerp(
		cec.currentEnvState.AudioAmbience,
		targetAudio,
		0.06,
	)
	
	// === Reality Distortion (Deep Thought Effect) ===
	// Distortion increases with high cognitive load and low coherence
	targetDistortion := 0.0
	if cognitive.CognitiveLoad > 0.7 && cognitive.Coherence < 0.5 {
		targetDistortion = (cognitive.CognitiveLoad - 0.7) * (1.0 - cognitive.Coherence) * strength
	}
	cec.currentEnvState.RealityDistortion = lerp(
		cec.currentEnvState.RealityDistortion,
		targetDistortion,
		0.04,
	)
	
	// === Environment Mood ===
	// Determine mood from emotional and cognitive states
	if emotional.Arousal > 0.7 {
		cec.currentEnvState.EnvironmentMood = "energetic"
	} else if cognitive.ProcessingMode == "contemplative" {
		cec.currentEnvState.EnvironmentMood = "contemplative"
	} else if emotional.Dominance > 0.7 && emotional.Confidence > 0.7 {
		cec.currentEnvState.EnvironmentMood = "mysterious"
	} else {
		cec.currentEnvState.EnvironmentMood = "calm"
	}
	
	// Trigger callback if set
	if cec.environmentCallback != nil {
		cec.environmentCallback(cec.currentEnvState)
	}
}

// SetCouplingStrength adjusts the coupling strength
func (cec *CognitiveEnvironmentCoupler) SetCouplingStrength(strength float64) {
	cec.mu.Lock()
	defer cec.mu.Unlock()
	cec.couplingStrength = clamp(strength, 0.0, 1.0)
}

// GetCouplingStrength returns the current coupling strength
func (cec *CognitiveEnvironmentCoupler) GetCouplingStrength() float64 {
	cec.mu.RLock()
	defer cec.mu.RUnlock()
	return cec.couplingStrength
}

// lerp performs linear interpolation
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// lerpColor performs linear interpolation on colors
func lerpColor(a, b Color, t float64) Color {
	return Color{
		R: lerp(a.R, b.R, t),
		G: lerp(a.G, b.G, t),
		B: lerp(a.B, b.B, t),
	}
}

// EnvironmentalStorytellingEngine generates narrative elements from environment state
type EnvironmentalStorytellingEngine struct {
	mu              sync.RWMutex
	currentNarrative string
	narrativeHistory []string
	historySize      int
}

// NewEnvironmentalStorytellingEngine creates a new storytelling engine
func NewEnvironmentalStorytellingEngine() *EnvironmentalStorytellingEngine {
	return &EnvironmentalStorytellingEngine{
		currentNarrative: "",
		narrativeHistory: make([]string, 0),
		historySize:      20,
	}
}

// GenerateNarrative creates a narrative description from environment state
func (ese *EnvironmentalStorytellingEngine) GenerateNarrative(envState EnvironmentState, avatarState AvatarState) string {
	ese.mu.Lock()
	defer ese.mu.Unlock()
	
	narrative := ""
	
	// Describe lighting
	if envState.LightingIntensity > 0.8 {
		narrative += "The space is bathed in brilliant light. "
	} else if envState.LightingIntensity < 0.3 {
		narrative += "Shadows dance in the dim ambience. "
	}
	
	// Describe color mood
	if envState.LightingColor.R > 0.8 && envState.LightingColor.G > 0.7 {
		narrative += "Warm golden hues fill the air. "
	} else if envState.LightingColor.B > 0.8 {
		narrative += "Cool azure tones create a serene atmosphere. "
	}
	
	// Describe particles
	if envState.ParticleIntensity > 0.6 {
		switch envState.ParticleType {
		case "thoughts":
			narrative += "Streams of thought-particles swirl around, visualizing the depth of contemplation. "
		case "energy":
			narrative += "Crackling energy particles pulse with vitality. "
		case "sparkles":
			narrative += "Sparkling motes of light dance playfully. "
		case "emotions":
			narrative += "Emotional resonances manifest as shimmering particles. "
		}
	}
	
	// Describe reality distortion
	if envState.RealityDistortion > 0.5 {
		narrative += "Reality itself seems to bend and warp, responding to the intensity of thought. "
	}
	
	// Describe overall mood
	switch envState.EnvironmentMood {
	case "energetic":
		narrative += "The environment pulses with dynamic energy. "
	case "contemplative":
		narrative += "A profound stillness invites deep reflection. "
	case "mysterious":
		narrative += "An enigmatic atmosphere surrounds everything. "
	case "calm":
		narrative += "Tranquility permeates the space. "
	}
	
	// Add to history
	ese.narrativeHistory = append(ese.narrativeHistory, narrative)
	if len(ese.narrativeHistory) > ese.historySize {
		ese.narrativeHistory = ese.narrativeHistory[1:]
	}
	
	ese.currentNarrative = narrative
	return narrative
}

// GetCurrentNarrative returns the current narrative description
func (ese *EnvironmentalStorytellingEngine) GetCurrentNarrative() string {
	ese.mu.RLock()
	defer ese.mu.RUnlock()
	return ese.currentNarrative
}

// GetNarrativeHistory returns the narrative history
func (ese *EnvironmentalStorytellingEngine) GetNarrativeHistory() []string {
	ese.mu.RLock()
	defer ese.mu.RUnlock()
	
	history := make([]string, len(ese.narrativeHistory))
	copy(history, ese.narrativeHistory)
	return history
}

// ProceduralAssetPlacement manages dynamic asset placement based on cognitive state
type ProceduralAssetPlacement struct {
	mu              sync.RWMutex
	assetPositions  map[string]Position3D
	assetTypes      []string
	placementRules  map[string]PlacementRule
}

// Position3D represents a 3D position
type Position3D struct {
	X, Y, Z float64
}

// PlacementRule defines how assets should be placed
type PlacementRule struct {
	MinDistance     float64
	MaxDistance     float64
	HeightRange     [2]float64
	DensityFactor   float64
	CognitiveAffinity string // Which cognitive state influences this asset
}

// NewProceduralAssetPlacement creates a new asset placement manager
func NewProceduralAssetPlacement() *ProceduralAssetPlacement {
	return &ProceduralAssetPlacement{
		assetPositions: make(map[string]Position3D),
		assetTypes:     []string{"thought_orb", "memory_crystal", "emotion_wisp", "idea_spark"},
		placementRules: map[string]PlacementRule{
			"thought_orb": {
				MinDistance:       2.0,
				MaxDistance:       10.0,
				HeightRange:       [2]float64{1.0, 3.0},
				DensityFactor:     0.5,
				CognitiveAffinity: "contemplative",
			},
			"memory_crystal": {
				MinDistance:       3.0,
				MaxDistance:       8.0,
				HeightRange:       [2]float64{0.5, 2.0},
				DensityFactor:     0.3,
				CognitiveAffinity: "awareness",
			},
			"emotion_wisp": {
				MinDistance:       1.0,
				MaxDistance:       6.0,
				HeightRange:       [2]float64{0.8, 2.5},
				DensityFactor:     0.7,
				CognitiveAffinity: "arousal",
			},
			"idea_spark": {
				MinDistance:       1.5,
				MaxDistance:       12.0,
				HeightRange:       [2]float64{1.5, 4.0},
				DensityFactor:     0.4,
				CognitiveAffinity: "creative",
			},
		},
	}
}

// UpdateAssetPlacement updates asset positions based on avatar state
func (pap *ProceduralAssetPlacement) UpdateAssetPlacement(avatarState AvatarState) map[string]Position3D {
	pap.mu.Lock()
	defer pap.mu.Unlock()
	
	// Clear old positions
	pap.assetPositions = make(map[string]Position3D)
	
	// Place assets based on cognitive state
	for _, assetType := range pap.assetTypes {
		rule := pap.placementRules[assetType]
		
		// Determine number of assets based on cognitive affinity
		var affinityValue float64
		switch rule.CognitiveAffinity {
		case "contemplative":
			if avatarState.Cognitive.ProcessingMode == "contemplative" {
				affinityValue = avatarState.Cognitive.CognitiveLoad
			}
		case "awareness":
			affinityValue = avatarState.Cognitive.Awareness
		case "arousal":
			affinityValue = avatarState.Emotional.Arousal
		case "creative":
			if avatarState.Cognitive.ProcessingMode == "creative" {
				affinityValue = avatarState.Emotional.Curiosity
			}
		}
		
		numAssets := int(math.Ceil(affinityValue * rule.DensityFactor * 10))
		
		// Place assets
		for i := 0; i < numAssets; i++ {
			assetID := assetType + "_" + string(rune(i))
			
			// Calculate position (simplified - would use proper spatial algorithms)
			angle := float64(i) * (2.0 * math.Pi / float64(numAssets))
			distance := rule.MinDistance + (rule.MaxDistance-rule.MinDistance)*affinityValue
			height := rule.HeightRange[0] + (rule.HeightRange[1]-rule.HeightRange[0])*affinityValue
			
			position := Position3D{
				X: distance * math.Cos(angle),
				Y: height,
				Z: distance * math.Sin(angle),
			}
			
			pap.assetPositions[assetID] = position
		}
	}
	
	return pap.assetPositions
}

// GetAssetPositions returns current asset positions
func (pap *ProceduralAssetPlacement) GetAssetPositions() map[string]Position3D {
	pap.mu.RLock()
	defer pap.mu.RUnlock()
	
	positions := make(map[string]Position3D)
	for k, v := range pap.assetPositions {
		positions[k] = v
	}
	return positions
}
