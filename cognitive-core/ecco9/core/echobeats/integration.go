package echobeats

import (
	"fmt"
	"log"
)

// ConsciousnessAdapter adapts the 3-phase system to autonomous consciousness
type ConsciousnessAdapter struct {
	consciousness interface{} // Reference to AutonomousConsciousness
	coherence     float64
}

// NewConsciousnessAdapter creates a new adapter
func NewConsciousnessAdapter(consciousness interface{}) *ConsciousnessAdapter {
	return &ConsciousnessAdapter{
		consciousness: consciousness,
		coherence:     1.0,
	}
}

// Integrate integrates cognitive streams into consciousness
func (ca *ConsciousnessAdapter) Integrate(streams []*CognitiveStream, couplings []Coupling) error {
	if len(streams) == 0 {
		return nil
	}

	log.Printf("🌊 Integrating %d cognitive streams with %d couplings", len(streams), len(couplings))

	// Process each stream
	for _, stream := range streams {
		if err := ca.processStream(stream); err != nil {
			log.Printf("❌ Error processing stream: %v", err)
		}
	}

	// Process couplings
	for _, coupling := range couplings {
		if err := ca.processCoupling(&coupling, streams); err != nil {
			log.Printf("❌ Error processing coupling: %v", err)
		}
	}

	// Update coherence based on integration success
	ca.updateCoherence(streams, couplings)

	return nil
}

// GetCoherence returns current coherence level
func (ca *ConsciousnessAdapter) GetCoherence() float64 {
	return ca.coherence
}

// processStream processes a single cognitive stream
func (ca *ConsciousnessAdapter) processStream(stream *CognitiveStream) error {
	// Log the stream processing
	log.Printf("  📥 Phase %d: %v%v - %v", stream.PhaseID, stream.Term, stream.Mode, stream.Content)

	// Here we would integrate with the actual consciousness system
	// For now, we just log the integration
	return nil
}

// processCoupling processes a tensional coupling
func (ca *ConsciousnessAdapter) processCoupling(coupling *Coupling, streams []*CognitiveStream) error {
	switch coupling.Type {
	case PerceptionMemory:
		return ca.processPerceptionMemoryCoupling(coupling, streams)
	case AssessmentPlanning:
		return ca.processAssessmentPlanningCoupling(coupling, streams)
	case BalancedIntegration:
		return ca.processBalancedIntegrationCoupling(coupling, streams)
	default:
		return fmt.Errorf("unknown coupling type: %v", coupling.Type)
	}
}

// processPerceptionMemoryCoupling handles T4E ↔ T7R coupling
func (ca *ConsciousnessAdapter) processPerceptionMemoryCoupling(coupling *Coupling, streams []*CognitiveStream) error {
	log.Println("  🔗 Processing Perception-Memory coupling")

	// Find T4E and T7R streams
	var perceptionStream, memoryStream *CognitiveStream
	for _, stream := range streams {
		if stream.Term == T4_SensoryInput && stream.Mode == Expressive {
			perceptionStream = stream
		}
		if stream.Term == T7_MemoryEncoding && stream.Mode == Reflective {
			memoryStream = stream
		}
	}

	if perceptionStream != nil && memoryStream != nil {
		log.Println("  💫 Memory-guided perception: Enriching current perception with relevant memories")

		// Extract content
		perception := perceptionStream.Content
		memories := memoryStream.Content

		// Create enriched perception
		enriched := map[string]interface{}{
			"current_perception": perception,
			"relevant_memories":  memories,
			"coupling_strength":  coupling.Strength,
			"type":               "memory_guided_perception",
		}

		log.Printf("  ✨ Enriched perception created: %v", enriched)

		// Here we would send this to consciousness for pattern recognition
		// For now, we just log it
	}

	return nil
}

// processAssessmentPlanningCoupling handles T1R ↔ T2E coupling
func (ca *ConsciousnessAdapter) processAssessmentPlanningCoupling(coupling *Coupling, streams []*CognitiveStream) error {
	log.Println("  🔗 Processing Assessment-Planning coupling")

	// Find T1R and T2E streams
	var assessmentStream, planningStream *CognitiveStream
	for _, stream := range streams {
		if stream.Term == T1_Perception && stream.Mode == Reflective {
			assessmentStream = stream
		}
		if stream.Term == T2_IdeaFormation && stream.Mode == Expressive {
			planningStream = stream
		}
	}

	if assessmentStream != nil && planningStream != nil {
		log.Println("  💫 Simulation-based planning: Generating ideas based on need assessment")

		// Extract content
		assessment := assessmentStream.Content
		planning := planningStream.Content

		// Create goal-directed plan
		plan := map[string]interface{}{
			"needs_assessment":  assessment,
			"generated_ideas":   planning,
			"coupling_strength": coupling.Strength,
			"type":              "goal_directed_plan",
		}

		log.Printf("  ✨ Goal-directed plan created: %v", plan)

		// Here we would send this to consciousness for execution
		// For now, we just log it
	}

	return nil
}

// processBalancedIntegrationCoupling handles T8E coupling
func (ca *ConsciousnessAdapter) processBalancedIntegrationCoupling(coupling *Coupling, streams []*CognitiveStream) error {
	log.Println("  🔗 Processing Balanced Integration")

	// Find T8E stream
	var balanceStream *CognitiveStream
	for _, stream := range streams {
		if stream.Term == T8_BalancedResponse && stream.Mode == Expressive {
			balanceStream = stream
		}
	}

	if balanceStream != nil {
		log.Println("  💫 Balanced integration: Coordinating all cognitive streams")

		// Gather all streams for integration
		allContent := make([]interface{}, 0, len(streams))
		for _, stream := range streams {
			allContent = append(allContent, stream.Content)
		}

		// Create integrated response
		integrated := map[string]interface{}{
			"balance_state":     balanceStream.Content,
			"all_streams":       allContent,
			"coupling_strength": coupling.Strength,
			"type":              "integrated_response",
		}

		log.Printf("  ✨ Integrated response created with %d streams: %v", len(allContent), integrated)

		// Here we would send this to consciousness for coordinated action
		// For now, we just log it
	}

	return nil
}

// updateCoherence updates coherence based on integration quality
func (ca *ConsciousnessAdapter) updateCoherence(streams []*CognitiveStream, couplings []Coupling) {
	// Coherence increases with successful integrations
	// Coherence decreases with conflicts or errors

	// Simple model: coherence based on stream count and coupling strength
	streamFactor := float64(len(streams)) / 3.0 // Normalize by expected 3 streams
	couplingFactor := 0.0
	if len(couplings) > 0 {
		totalStrength := 0.0
		for _, coupling := range couplings {
			totalStrength += coupling.Strength
		}
		couplingFactor = totalStrength / float64(len(couplings))
	}

	// Update coherence with exponential moving average
	alpha := 0.3
	newCoherence := (streamFactor + couplingFactor) / 2.0
	ca.coherence = alpha*newCoherence + (1-alpha)*ca.coherence

	// Clamp to [0, 1]
	if ca.coherence < 0 {
		ca.coherence = 0
	} else if ca.coherence > 1 {
		ca.coherence = 1
	}

	log.Printf("  📊 Coherence updated: %.3f", ca.coherence)
}
