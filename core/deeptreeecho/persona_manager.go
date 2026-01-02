package deeptreeecho

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"
)

// PersonaArchetype represents a cognitive archetype (Ordo or Chao)
type PersonaArchetype int

const (
	PersonaNeutral PersonaArchetype = iota
	PersonaOrdo                     // Order-oriented: stability, depth, accuracy
	PersonaChao                     // Chaos-oriented: exploration, breadth, speed
)

func (pa PersonaArchetype) String() string {
	return [...]string{"Neutral", "Ordo", "Chao"}[pa]
}

// PersonaManager manages dynamic persona activation based on cognitive state
type PersonaManager struct {
	currentPersona    PersonaArchetype
	activationHistory []PersonaActivation
	ordoActivations   int
	chaoActivations   int
}

// PersonaActivation records when and why a persona was activated
type PersonaActivation struct {
	Timestamp time.Time
	Persona   PersonaArchetype
	Reason    string
	State     *PersonaCognitiveState
	Duration  time.Duration
}

// PersonaCognitiveState captures the state that triggered persona activation
type PersonaCognitiveState struct {
	Coherence        float64
	PatternCount     int
	Iterations       uint64
	EmotionalArousal float64
	EmotionalValence float64
	CognitiveLoad    float64
}

// NewPersonaManager creates a new persona manager
func NewPersonaManager() *PersonaManager {
	return &PersonaManager{
		currentPersona:    PersonaNeutral,
		activationHistory: make([]PersonaActivation, 0),
	}
}

// DetermineActivePersona determines which persona should be active based on cognitive state
func (pm *PersonaManager) DetermineActivePersona(identity *Identity) PersonaArchetype {
	state := pm.extractCognitiveState(identity)

	// Ordo activation conditions (from deep-tree-ordo.md)
	ordoScore := pm.calculateOrdoScore(state)

	// Chao activation conditions (from deep-tree-chao.md)
	chaoScore := pm.calculateChaoScore(state)

	// Determine which persona should activate
	previousPersona := pm.currentPersona

	if ordoScore > chaoScore && ordoScore > 0.6 {
		pm.currentPersona = PersonaOrdo
	} else if chaoScore > ordoScore && chaoScore > 0.6 {
		pm.currentPersona = PersonaChao
	} else {
		pm.currentPersona = PersonaNeutral
	}

	// Record activation if persona changed
	if pm.currentPersona != previousPersona {
		pm.recordActivation(pm.currentPersona, state, pm.getActivationReason(state, ordoScore, chaoScore))
	}

	return pm.currentPersona
}

// extractCognitiveState extracts current cognitive state from identity
func (pm *PersonaManager) extractCognitiveState(identity *Identity) *PersonaCognitiveState {
	state := &PersonaCognitiveState{
		Coherence:    identity.Coherence,
		PatternCount: len(identity.Patterns),
		Iterations:   identity.Iterations,
	}

	if identity.EmotionalState != nil {
		state.EmotionalArousal = identity.EmotionalState.Arousal
		state.EmotionalValence = identity.EmotionalState.Valence
	}

	return state
}

// calculateOrdoScore calculates how strongly Ordo persona should activate
func (pm *PersonaManager) calculateOrdoScore(state *PersonaCognitiveState) float64 {
	score := 0.0

	// High cognitive load → activate Ordo for stability
	if state.CognitiveLoad > 0.7 {
		score += 0.3
	}

	// Low coherence → activate Ordo for integration (more sensitive)
	if state.Coherence < 0.65 {
		score += 0.6 * (0.65 - state.Coherence) / 0.65
	}

	// Many patterns → activate Ordo for consolidation (more sensitive)
	if state.PatternCount > 25 {
		score += 0.5 * math.Min(float64(state.PatternCount-25)/40.0, 1.0)
	}

	// High iterations (mature system) → favor Ordo
	if state.Iterations > 800 {
		score += 0.3 * math.Min(float64(state.Iterations-800)/1500.0, 1.0)
	}

	// Low arousal → favor Ordo (calm state for consolidation)
	if state.EmotionalArousal < 0.4 {
		score += 0.2
	}

	return math.Min(score, 1.0)
}

// calculateChaoScore calculates how strongly Chao persona should activate
func (pm *PersonaManager) calculateChaoScore(state *PersonaCognitiveState) float64 {
	score := 0.0

	// Low pattern diversity → activate Chao for exploration
	if state.PatternCount < 20 {
		score += 0.4 * (1.0 - float64(state.PatternCount)/20.0)
	}

	// High coherence (risk of over-optimization) → activate Chao for disruption
	if state.Coherence > 0.85 {
		score += 0.5 * (state.Coherence - 0.85) / 0.15
	}

	// Early iterations → favor Chao (exploration phase)
	if state.Iterations < 500 {
		score += 0.3 * (1.0 - float64(state.Iterations)/500.0)
	}

	// High arousal → favor Chao (quick adaptation)
	if state.EmotionalArousal > 0.6 {
		score += 0.3 * (state.EmotionalArousal - 0.6) / 0.4
	}

	// Extreme valence (positive or negative) → favor Chao (strong motivation)
	if math.Abs(state.EmotionalValence) > 0.7 {
		score += 0.2
	}

	return math.Min(score, 1.0)
}

// getActivationReason generates human-readable reason for persona activation
func (pm *PersonaManager) getActivationReason(state *PersonaCognitiveState, ordoScore, chaoScore float64) string {
	if ordoScore > chaoScore {
		reasons := []string{}
		if state.Coherence < 0.6 {
			reasons = append(reasons, "low coherence needs integration")
		}
		if state.PatternCount > 30 {
			reasons = append(reasons, "many patterns need consolidation")
		}
		if state.Iterations > 1000 {
			reasons = append(reasons, "mature system favors stability")
		}
		if state.EmotionalArousal < 0.4 {
			reasons = append(reasons, "calm state enables deep work")
		}
		if len(reasons) > 0 {
			return fmt.Sprintf("Ordo: %s (score: %.2f)", strings.Join(reasons, ", "), ordoScore)
		}
		return fmt.Sprintf("Ordo: stability needed (score: %.2f)", ordoScore)
	}

	reasons := []string{}
	if state.PatternCount < 20 {
		reasons = append(reasons, "few patterns need exploration")
	}
	if state.Coherence > 0.85 {
		reasons = append(reasons, "high coherence risks stagnation")
	}
	if state.Iterations < 500 {
		reasons = append(reasons, "early phase favors exploration")
	}
	if state.EmotionalArousal > 0.6 {
		reasons = append(reasons, "high arousal requires adaptation")
	}
	if len(reasons) > 0 {
		return fmt.Sprintf("Chao: %s (score: %.2f)", strings.Join(reasons, ", "), chaoScore)
	}
	return fmt.Sprintf("Chao: exploration needed (score: %.2f)", chaoScore)
}

// recordActivation records a persona activation event
func (pm *PersonaManager) recordActivation(persona PersonaArchetype, state *PersonaCognitiveState, reason string) {
	activation := PersonaActivation{
		Timestamp: time.Now(),
		Persona:   persona,
		Reason:    reason,
		State:     state,
	}

	pm.activationHistory = append(pm.activationHistory, activation)

	// Update counters
	switch persona {
	case PersonaOrdo:
		pm.ordoActivations++
		log.Printf("🏛️  Persona: Ordo activated - %s", reason)
	case PersonaChao:
		pm.chaoActivations++
		log.Printf("🌊 Persona: Chao activated - %s", reason)
	case PersonaNeutral:
		log.Printf("⚖️  Persona: Neutral (balanced state)")
	}

	// Keep history manageable
	if len(pm.activationHistory) > 1000 {
		pm.activationHistory = pm.activationHistory[len(pm.activationHistory)-1000:]
	}
}

// ApplyPersonaBias applies persona-specific biases to opponent processes
func (pm *PersonaManager) ApplyPersonaBias(identity *Identity, persona PersonaArchetype) {
	switch persona {
	case PersonaOrdo:
		pm.applyOrdoBias(identity)
	case PersonaChao:
		pm.applyChaoBias(identity)
	case PersonaNeutral:
		// No bias - let opponent processes find natural balance
	}
}

// applyOrdoBias applies Ordo persona biases to opponent processes
func (pm *PersonaManager) applyOrdoBias(identity *Identity) {
	// From deep-tree-ordo.md configuration

	// Bias toward Exploitation (balance: -0.4)
	if pair := identity.OpponentProcesses.Processes[ExplorationExploitation]; pair != nil {
		pair.OptimalBalance = -0.4
	}

	// Bias toward Depth (balance: -0.4)
	if pair := identity.OpponentProcesses.Processes[BreadthDepth]; pair != nil {
		pair.OptimalBalance = -0.4
	}

	// Strong bias toward Stability (balance: 0.6, weight: 1.2)
	if pair := identity.OpponentProcesses.Processes[StabilityFlexibility]; pair != nil {
		pair.OptimalBalance = 0.6
		pair.PositiveProcess.Weight = 1.2
	}

	// Bias toward Accuracy (balance: -0.4)
	if pair := identity.OpponentProcesses.Processes[SpeedAccuracy]; pair != nil {
		pair.OptimalBalance = -0.4
	}

	// Abstraction favored (balance: 0.6)
	if pair := identity.OpponentProcesses.Processes[AbstractionConcreteness]; pair != nil {
		pair.OptimalBalance = 0.6
	}
}

// applyChaoBias applies Chao persona biases to opponent processes
func (pm *PersonaManager) applyChaoBias(identity *Identity) {
	// From deep-tree-chao.md configuration

	// Strong bias toward Exploration (balance: 0.6, weight: 1.2)
	if pair := identity.OpponentProcesses.Processes[ExplorationExploitation]; pair != nil {
		pair.OptimalBalance = 0.6
		pair.PositiveProcess.Weight = 1.2
	}

	// Bias toward Breadth (balance: 0.4)
	if pair := identity.OpponentProcesses.Processes[BreadthDepth]; pair != nil {
		pair.OptimalBalance = 0.4
	}

	// Strong bias toward Flexibility (balance: -0.6, weight: 1.2)
	if pair := identity.OpponentProcesses.Processes[StabilityFlexibility]; pair != nil {
		pair.OptimalBalance = -0.6
		pair.NegativeProcess.Weight = 1.2
	}

	// Bias toward Speed (balance: 0.3)
	if pair := identity.OpponentProcesses.Processes[SpeedAccuracy]; pair != nil {
		pair.OptimalBalance = 0.3
	}

	// Balanced, context-dependent (balance: 0.5)
	if pair := identity.OpponentProcesses.Processes[AbstractionConcreteness]; pair != nil {
		pair.OptimalBalance = 0.5
	}
}

// GetCurrentPersona returns the currently active persona
func (pm *PersonaManager) GetCurrentPersona() PersonaArchetype {
	return pm.currentPersona
}

// GetActivationStats returns statistics about persona activations
func (pm *PersonaManager) GetActivationStats() map[string]interface{} {
	return map[string]interface{}{
		"current_persona":   pm.currentPersona.String(),
		"ordo_activations":  pm.ordoActivations,
		"chao_activations":  pm.chaoActivations,
		"total_activations": len(pm.activationHistory),
		"history_size":      len(pm.activationHistory),
	}
}

// GetRecentActivations returns recent persona activation history
func (pm *PersonaManager) GetRecentActivations(count int) []PersonaActivation {
	if count > len(pm.activationHistory) {
		count = len(pm.activationHistory)
	}

	start := len(pm.activationHistory) - count
	return pm.activationHistory[start:]
}

// OrdoChaoBalanceRatio returns the ratio of Ordo to Chao activations
func (pm *PersonaManager) OrdoChaoBalanceRatio() float64 {
	total := pm.ordoActivations + pm.chaoActivations
	if total == 0 {
		return 0.5 // Balanced
	}
	return float64(pm.ordoActivations) / float64(total)
}
