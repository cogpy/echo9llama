package live2d

import (
	"time"
)

// EchoStateBridge bridges Deep Tree Echo cognitive states to Live2D avatar
type EchoStateBridge struct {
	avatarManager *AvatarManager
}

// NewEchoStateBridge creates a new bridge between Echo9 and Live2D
func NewEchoStateBridge(avatarManager *AvatarManager) *EchoStateBridge {
	return &EchoStateBridge{
		avatarManager: avatarManager,
	}
}

// UpdateFromEchoEmotion updates avatar from Echo9's emotional dynamics
// Maps Echo9's emotional model to Live2D avatar parameters
func (eb *EchoStateBridge) UpdateFromEchoEmotion(echoEmotion map[string]float64) error {
	// Extract emotional dimensions from Echo9's emotion system
	// Echo9 uses: joy, sadness, anger, fear, disgust, surprise, trust, anticipation

	// Calculate valence (positive/negative emotion)
	valence := 0.0
	if joy, ok := echoEmotion["joy"]; ok {
		valence += joy
	}
	if trust, ok := echoEmotion["trust"]; ok {
		valence += trust * 0.5
	}
	if sadness, ok := echoEmotion["sadness"]; ok {
		valence -= sadness
	}
	if fear, ok := echoEmotion["fear"]; ok {
		valence -= fear * 0.5
	}
	if anger, ok := echoEmotion["anger"]; ok {
		valence -= anger * 0.7
	}

	// Calculate arousal (activation level)
	arousal := 0.0
	if excitement, ok := echoEmotion["excitement"]; ok {
		arousal = excitement
	}
	if surprise, ok := echoEmotion["surprise"]; ok {
		arousal += surprise * 0.8
	}
	if anticipation, ok := echoEmotion["anticipation"]; ok {
		arousal += anticipation * 0.6
	}

	// Calculate dominance
	dominance := 0.5 // Default
	if anger, ok := echoEmotion["anger"]; ok {
		dominance += anger * 0.5
	}
	if fear, ok := echoEmotion["fear"]; ok {
		dominance -= fear * 0.5
	}

	// Get curiosity and confidence if available
	curiosity := 0.3
	if c, ok := echoEmotion["curiosity"]; ok {
		curiosity = c
	}

	confidence := 0.5
	if c, ok := echoEmotion["confidence"]; ok {
		confidence = c
	}

	// Create Live2D emotional state
	emotional := EmotionalState{
		Valence:    clamp(valence, -1.0, 1.0),
		Arousal:    clamp(arousal, 0.0, 1.0),
		Dominance:  clamp(dominance, 0.0, 1.0),
		Curiosity:  clamp(curiosity, 0.0, 1.0),
		Confidence: clamp(confidence, 0.0, 1.0),
	}

	return eb.avatarManager.UpdateEmotionalState(emotional)
}

// UpdateFromEchoCognitive updates avatar from Echo9's cognitive state
func (eb *EchoStateBridge) UpdateFromEchoCognitive(echoCognitive map[string]interface{}) error {
	// Extract cognitive metrics from Echo9
	cognitive := CognitiveState{
		Awareness:      extractFloat(echoCognitive, "awareness", 0.5),
		Attention:      extractFloat(echoCognitive, "attention", 0.5),
		CognitiveLoad:  extractFloat(echoCognitive, "cognitive_load", 0.3),
		Coherence:      extractFloat(echoCognitive, "coherence", 0.7),
		EnergyLevel:    extractFloat(echoCognitive, "energy_level", 0.7),
		ProcessingMode: extractString(echoCognitive, "processing_mode", "contemplative"),
	}

	return eb.avatarManager.UpdateCognitiveState(cognitive)
}

// UpdateFromEchoReservoir updates based on reservoir network state
func (eb *EchoStateBridge) UpdateFromEchoReservoir(reservoirState map[string]interface{}) error {
	// Map reservoir dynamics to cognitive state
	cognitive := CognitiveState{
		Awareness:      extractFloat(reservoirState, "spectral_radius", 0.5),
		Attention:      extractFloat(reservoirState, "input_scaling", 0.5),
		CognitiveLoad:  extractFloat(reservoirState, "leak_rate", 0.3),
		Coherence:      extractFloat(reservoirState, "stability", 0.7),
		EnergyLevel:    1.0 - extractFloat(reservoirState, "fatigue", 0.3),
		ProcessingMode: mapPersonaToMode(extractString(reservoirState, "persona", "contemplative")),
	}

	return eb.avatarManager.UpdateCognitiveState(cognitive)
}

// UpdateFromEchoBeats updates based on EchoBeats cognitive cycle
func (eb *EchoStateBridge) UpdateFromEchoBeats(step int, phase string) error {
	// Map EchoBeats 12-step cycle to avatar behavior
	currentState := eb.avatarManager.GetCurrentState()
	cognitive := currentState.Cognitive

	// Adjust attention based on phase
	switch phase {
	case "affordance": // Steps 1-6
		cognitive.Attention = 0.8
		cognitive.ProcessingMode = "dynamic"
	case "reorientation": // Step 7
		cognitive.Attention = 0.6
		cognitive.ProcessingMode = "contemplative"
	case "salience": // Steps 8-12
		cognitive.Attention = 0.7
		cognitive.ProcessingMode = "creative"
	}

	// Subtle head movement based on step
	// This creates a natural rhythmic motion

	return eb.avatarManager.UpdateCognitiveState(cognitive)
}

// UpdateFromWisdomMetrics updates based on 7-dimensional wisdom
func (eb *EchoStateBridge) UpdateFromWisdomMetrics(wisdom map[string]float64) error {
	// Calculate overall wisdom influence on avatar
	avgWisdom := 0.0
	count := 0
	for _, value := range wisdom {
		avgWisdom += value
		count++
	}
	if count > 0 {
		avgWisdom /= float64(count)
	}

	// Higher wisdom -> calmer, more confident demeanor
	emotional := eb.avatarManager.GetCurrentState().Emotional
	emotional.Confidence = clamp(emotional.Confidence+avgWisdom*0.2, 0.0, 1.0)
	emotional.Valence = clamp(emotional.Valence+avgWisdom*0.1, -1.0, 1.0)

	return eb.avatarManager.UpdateEmotionalState(emotional)
}

// SyncWithThoughtGeneration updates avatar during thought generation
func (eb *EchoStateBridge) SyncWithThoughtGeneration(thoughtType string, inProgress bool) error {
	cognitive := eb.avatarManager.GetCurrentState().Cognitive
	emotional := eb.avatarManager.GetCurrentState().Emotional

	if inProgress {
		// Increase cognitive load and attention during thought generation
		cognitive.CognitiveLoad = clamp(cognitive.CognitiveLoad+0.2, 0.0, 1.0)
		cognitive.Attention = clamp(cognitive.Attention+0.1, 0.0, 1.0)

		// Adjust emotional state based on thought type
		switch thoughtType {
		case "reflection":
			emotional.Curiosity = clamp(emotional.Curiosity+0.1, 0.0, 1.0)
			cognitive.ProcessingMode = "contemplative"
		case "question":
			emotional.Curiosity = clamp(emotional.Curiosity+0.3, 0.0, 1.0)
			cognitive.ProcessingMode = "dynamic"
		case "insight":
			emotional.Arousal = clamp(emotional.Arousal+0.2, 0.0, 1.0)
			emotional.Valence = clamp(emotional.Valence+0.2, -1.0, 1.0)
			cognitive.ProcessingMode = "creative"
		case "planning":
			cognitive.Attention = clamp(cognitive.Attention+0.2, 0.0, 1.0)
			cognitive.ProcessingMode = "cautious"
		}
	} else {
		// Reset to baseline after thought generation
		cognitive.CognitiveLoad = clamp(cognitive.CognitiveLoad-0.1, 0.0, 1.0)
	}

	// Update both states
	if err := eb.avatarManager.UpdateCognitiveState(cognitive); err != nil {
		return err
	}
	return eb.avatarManager.UpdateEmotionalState(emotional)
}

// Helper functions

func extractFloat(data map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		}
	}
	return defaultVal
}

func extractString(data map[string]interface{}, key string, defaultVal string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultVal
}

func mapPersonaToMode(persona string) string {
	switch persona {
	case "contemplative-scholar", "contemplative":
		return "contemplative"
	case "dynamic-explorer", "dynamic":
		return "dynamic"
	case "cautious-analyst", "cautious":
		return "cautious"
	case "creative-visionary", "creative":
		return "creative"
	default:
		return "contemplative"
	}
}

// PeriodicSync continuously syncs Echo9 state to avatar
func (eb *EchoStateBridge) PeriodicSync(syncFunc func() (AvatarState, error), interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if state, err := syncFunc(); err == nil {
			eb.avatarManager.UpdateFullState(state)
		}
	}
}
