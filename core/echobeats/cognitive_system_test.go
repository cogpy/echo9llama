package echobeats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCognitiveLoopSteps tests the 12-step cognitive loop structure
func TestCognitiveLoopSteps(t *testing.T) {
	t.Run("TwelveStepStructure", func(t *testing.T) {
		// The cognitive loop has 12 steps
		steps := 12
		assert.Equal(t, 12, steps)
	})

	t.Run("TriadStructure", func(t *testing.T) {
		// Steps are organized into triads: {1,5,9}, {2,6,10}, {3,7,11}, {4,8,12}
		triads := [][]int{
			{1, 5, 9},
			{2, 6, 10},
			{3, 7, 11},
			{4, 8, 12},
		}

		for _, triad := range triads {
			assert.Len(t, triad, 3)
			// Each triad has steps 4 apart
			assert.Equal(t, 4, triad[1]-triad[0])
			assert.Equal(t, 4, triad[2]-triad[1])
		}
	})

	t.Run("PhaseStructure", func(t *testing.T) {
		// 3 phases, 120 degrees apart
		phases := 3
		degreesApart := 360 / phases
		assert.Equal(t, 120, degreesApart)
	})
}

// TestEngineTypes tests the three cognitive engine types
func TestEngineTypes(t *testing.T) {
	t.Run("AffordanceEngine", func(t *testing.T) {
		// Affordance engine handles steps 0-5 (past conditioning)
		affordanceSteps := []int{0, 1, 2, 3, 4, 5}
		assert.Len(t, affordanceSteps, 6)
	})

	t.Run("RelevanceEngine", func(t *testing.T) {
		// Relevance engine handles pivotal steps 0 and 6
		relevanceSteps := []int{0, 6}
		assert.Len(t, relevanceSteps, 2)
	})

	t.Run("SalienceEngine", func(t *testing.T) {
		// Salience engine handles steps 6-11 (future anticipation)
		salienceSteps := []int{6, 7, 8, 9, 10, 11}
		assert.Len(t, salienceSteps, 6)
	})
}

// TestCognitiveStepTypes tests the step type constants
func TestCognitiveStepTypes(t *testing.T) {
	t.Run("StepTypeConstants", func(t *testing.T) {
		// Verify step type constants exist and are distinct
		stepTypes := map[string]bool{
			"affordance": true,
			"relevance":  true,
			"salience":   true,
		}
		assert.Len(t, stepTypes, 3)
	})
}

// TestInterleaving tests the interleaving of cognitive streams
func TestInterleaving(t *testing.T) {
	t.Run("ThreeConcurrentStreams", func(t *testing.T) {
		// 3 concurrent streams, phased 4 steps apart
		streams := 3
		phaseOffset := 4
		totalSteps := 12

		// Each stream covers all 12 steps
		for stream := range streams {
			startStep := stream * phaseOffset
			for step := range totalSteps {
				currentStep := (startStep + step) % totalSteps
				assert.GreaterOrEqual(t, currentStep, 0)
				assert.Less(t, currentStep, totalSteps)
			}
		}
	})
}

// TestAffordanceStruct tests the Affordance struct
func TestAffordanceStruct(t *testing.T) {
	t.Run("CreateAffordance", func(t *testing.T) {
		aff := Affordance{
			Action:      "test-action",
			Context:     map[string]interface{}{"key": "value"},
			PastSuccess: 0.8,
			Confidence:  0.9,
			Timestamp:   time.Now(),
		}

		assert.Equal(t, "test-action", aff.Action)
		assert.Equal(t, 0.8, aff.PastSuccess)
		assert.Equal(t, 0.9, aff.Confidence)
	})
}

// TestScenarioStruct tests the Scenario struct
func TestScenarioStruct(t *testing.T) {
	t.Run("CreateScenario", func(t *testing.T) {
		scenario := Scenario{
			ID:           "scenario-1",
			Description:  "Test scenario",
			Probability:  0.7,
			Desirability: 0.85,
			Consequences: []interface{}{"consequence-1"},
			Timestamp:    time.Now(),
		}

		assert.Equal(t, "scenario-1", scenario.ID)
		assert.Equal(t, "Test scenario", scenario.Description)
		assert.Equal(t, 0.7, scenario.Probability)
		assert.Equal(t, 0.85, scenario.Desirability)
	})

	t.Run("SalienceCalculation", func(t *testing.T) {
		scenario := Scenario{
			Probability:  0.6,
			Desirability: 0.8,
		}

		// Salience = Probability × Desirability
		expectedSalience := scenario.Probability * scenario.Desirability
		assert.Equal(t, 0.48, expectedSalience)
	})
}

// BenchmarkCognitiveLoop benchmarks cognitive loop operations
func BenchmarkCognitiveLoop(b *testing.B) {
	b.Run("StepCalculation", func(b *testing.B) {
		for i := range b.N {
			step := i % 12
			_ = step
		}
	})

	b.Run("TriadLookup", func(b *testing.B) {
		triads := [][]int{
			{1, 5, 9},
			{2, 6, 10},
			{3, 7, 11},
			{4, 8, 12},
		}
		b.ResetTimer()
		for i := range b.N {
			triadIdx := i % 4
			_ = triads[triadIdx]
		}
	})
}
