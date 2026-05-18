// Package pienn - PWL Cognitive Network
//
// Integrates Piecewise-Linear Kolmogorov-Arnold Networks (PWL-KAN) into the
// PIE-NN cognitive architecture. Instead of fixed trait values, cognitive
// parameters become learnable functions that adapt through experience.
//
// The PIE-NN cognitive core's traits (curiosity, assertiveness, humor, etc.)
// are no longer static floats but piecewise-linear functions of context,
// allowing the system to develop nuanced, context-dependent personality.
package pienn

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// ──────────────────────────────────────────────────────────────
// PWL-KAN Cognitive Network
// ──────────────────────────────────────────────────────────────

// PWLFunction represents a learnable piecewise-linear function
// that maps a scalar input to a scalar output via split points.
type PWLFunction struct {
	// SplitPoints are the x-coordinates where the function changes slope
	SplitPoints []float64
	// ControlValues are the y-values at each split point
	ControlValues []float64
	// Gradients accumulated for backpropagation
	SplitGradients   []float64
	ControlGradients []float64
	// Learning rate for this function
	LearningRate float64
	// Whether split points can move
	SplitsTrainable bool
	// Extrapolation mode
	LinearExtrapolation bool
}

// NewPWLFunction creates a new piecewise-linear function with uniform splits
func NewPWLFunction(numSplits int, rangeMin, rangeMax float64) *PWLFunction {
	splits := make([]float64, numSplits)
	controls := make([]float64, numSplits)
	splitGrads := make([]float64, numSplits)
	controlGrads := make([]float64, numSplits)

	step := (rangeMax - rangeMin) / float64(numSplits-1)
	for i := 0; i < numSplits; i++ {
		splits[i] = rangeMin + float64(i)*step
		controls[i] = 0.5 // Initialize to midpoint
	}

	return &PWLFunction{
		SplitPoints:         splits,
		ControlValues:       controls,
		SplitGradients:      splitGrads,
		ControlGradients:    controlGrads,
		LearningRate:        0.01,
		SplitsTrainable:     false,
		LinearExtrapolation: true,
	}
}

// Evaluate computes the PWL function value at input x
func (f *PWLFunction) Evaluate(x float64) float64 {
	n := len(f.SplitPoints)
	if n == 0 {
		return 0.5
	}

	// Below range
	if x <= f.SplitPoints[0] {
		if f.LinearExtrapolation && n > 1 {
			slope := (f.ControlValues[1] - f.ControlValues[0]) / (f.SplitPoints[1] - f.SplitPoints[0])
			return f.ControlValues[0] + slope*(x-f.SplitPoints[0])
		}
		return f.ControlValues[0]
	}

	// Above range
	if x >= f.SplitPoints[n-1] {
		if f.LinearExtrapolation && n > 1 {
			slope := (f.ControlValues[n-1] - f.ControlValues[n-2]) / (f.SplitPoints[n-1] - f.SplitPoints[n-2])
			return f.ControlValues[n-1] + slope*(x-f.SplitPoints[n-1])
		}
		return f.ControlValues[n-1]
	}

	// Find the segment
	for i := 0; i < n-1; i++ {
		if x >= f.SplitPoints[i] && x < f.SplitPoints[i+1] {
			// Linear interpolation within segment
			t := (x - f.SplitPoints[i]) / (f.SplitPoints[i+1] - f.SplitPoints[i])
			return f.ControlValues[i]*(1-t) + f.ControlValues[i+1]*t
		}
	}

	return f.ControlValues[n-1]
}

// AccumulateGradient records a gradient signal for learning
func (f *PWLFunction) AccumulateGradient(x float64, gradient float64) {
	n := len(f.SplitPoints)
	if n == 0 {
		return
	}

	// Find affected segment and accumulate gradient to control points
	for i := 0; i < n-1; i++ {
		if x >= f.SplitPoints[i] && x < f.SplitPoints[i+1] {
			t := (x - f.SplitPoints[i]) / (f.SplitPoints[i+1] - f.SplitPoints[i])
			f.ControlGradients[i] += gradient * (1 - t)
			f.ControlGradients[i+1] += gradient * t
			return
		}
	}

	// Edge cases
	if x <= f.SplitPoints[0] {
		f.ControlGradients[0] += gradient
	} else {
		f.ControlGradients[n-1] += gradient
	}
}

// ApplyGradients updates the function parameters based on accumulated gradients
func (f *PWLFunction) ApplyGradients() {
	for i := range f.ControlValues {
		f.ControlValues[i] += f.LearningRate * f.ControlGradients[i]
		// Clamp to [0, 1] for trait functions
		f.ControlValues[i] = math.Max(0.0, math.Min(1.0, f.ControlValues[i]))
		f.ControlGradients[i] = 0 // Reset gradient
	}

	if f.SplitsTrainable {
		for i := range f.SplitPoints {
			f.SplitPoints[i] += f.LearningRate * 0.1 * f.SplitGradients[i]
			f.SplitGradients[i] = 0
		}
		// Enforce monotonicity of split points
		f.enforceSplitMonotonicity(0.01)
	}
}

func (f *PWLFunction) enforceSplitMonotonicity(margin float64) {
	for i := 1; i < len(f.SplitPoints); i++ {
		if f.SplitPoints[i] <= f.SplitPoints[i-1]+margin {
			f.SplitPoints[i] = f.SplitPoints[i-1] + margin
		}
	}
}

// ──────────────────────────────────────────────────────────────
// Cognitive PWL Network
// ──────────────────────────────────────────────────────────────

// CognitivePWLNetwork is a network of PWL functions that map context
// features to cognitive trait activations. This makes personality
// context-dependent and learnable.
type CognitivePWLNetwork struct {
	mu sync.RWMutex

	// Trait functions: trait_name -> context_feature -> PWL function
	// e.g., "curiosity" -> "novelty" -> PWLFunction
	TraitFunctions map[string]map[string]*PWLFunction

	// Context features currently active
	ContextFeatures map[string]float64

	// Output trait values (computed from PWL network)
	ComputedTraits map[string]float64

	// Learning signal accumulator
	RewardHistory []RewardSignal

	// Network metadata
	TotalUpdates   uint64
	LastUpdate     time.Time
	WisdomAccrued  float64
}

// RewardSignal represents a learning signal from experience
type RewardSignal struct {
	Timestamp time.Time
	Context   map[string]float64
	Traits    map[string]float64
	Reward    float64 // Positive = good outcome, Negative = bad outcome
	Source    string  // What generated this signal
}

// NewCognitivePWLNetwork creates a new cognitive PWL network
func NewCognitivePWLNetwork() *CognitivePWLNetwork {
	network := &CognitivePWLNetwork{
		TraitFunctions:  make(map[string]map[string]*PWLFunction),
		ContextFeatures: make(map[string]float64),
		ComputedTraits:  make(map[string]float64),
		RewardHistory:   make([]RewardSignal, 0),
	}

	// Initialize trait functions for each cognitive trait
	traits := []string{
		"curiosity", "assertiveness", "humor", "defiance",
		"depth", "chaos", "empathy", "intensity",
		"patience", "wisdom", "creativity", "focus",
	}

	// Context features that modulate traits
	contexts := []string{
		"novelty",        // How novel is the current input
		"threat_level",   // Perceived hostility/challenge
		"complexity",     // Intellectual complexity of situation
		"social_warmth",  // Warmth of the interaction
		"time_pressure",  // Urgency of the situation
		"cognitive_load", // Current mental load
		"fatigue",        // Current fatigue level
		"interest",       // Current interest level
	}

	for _, trait := range traits {
		network.TraitFunctions[trait] = make(map[string]*PWLFunction)
		for _, ctx := range contexts {
			pwl := NewPWLFunction(8, 0.0, 1.0) // 8 split points over [0,1]
			// Initialize with sensible defaults based on trait-context pairing
			network.initializeTraitContext(pwl, trait, ctx)
			network.TraitFunctions[trait][ctx] = pwl
		}
	}

	return network
}

// initializeTraitContext sets initial PWL values based on trait-context semantics
func (n *CognitivePWLNetwork) initializeTraitContext(pwl *PWLFunction, trait, context string) {
	// Set initial control values based on expected relationships
	switch {
	case trait == "curiosity" && context == "novelty":
		// Curiosity increases with novelty
		for i := range pwl.ControlValues {
			pwl.ControlValues[i] = 0.3 + 0.6*float64(i)/float64(len(pwl.ControlValues)-1)
		}
	case trait == "assertiveness" && context == "threat_level":
		// Assertiveness increases with threat
		for i := range pwl.ControlValues {
			pwl.ControlValues[i] = 0.4 + 0.5*float64(i)/float64(len(pwl.ControlValues)-1)
		}
	case trait == "defiance" && context == "threat_level":
		// Defiance spikes at high threat
		for i := range pwl.ControlValues {
			t := float64(i) / float64(len(pwl.ControlValues)-1)
			pwl.ControlValues[i] = 0.3 + 0.6*t*t // Quadratic increase
		}
	case trait == "humor" && context == "social_warmth":
		// Humor increases with social warmth
		for i := range pwl.ControlValues {
			pwl.ControlValues[i] = 0.2 + 0.6*float64(i)/float64(len(pwl.ControlValues)-1)
		}
	case trait == "depth" && context == "complexity":
		// Depth increases with complexity
		for i := range pwl.ControlValues {
			pwl.ControlValues[i] = 0.4 + 0.5*float64(i)/float64(len(pwl.ControlValues)-1)
		}
	case trait == "patience" && context == "fatigue":
		// Patience decreases with fatigue
		for i := range pwl.ControlValues {
			pwl.ControlValues[i] = 0.8 - 0.6*float64(i)/float64(len(pwl.ControlValues)-1)
		}
	case trait == "wisdom" && context == "complexity":
		// Wisdom activates more with complexity
		for i := range pwl.ControlValues {
			t := float64(i) / float64(len(pwl.ControlValues)-1)
			pwl.ControlValues[i] = 0.3 + 0.5*math.Sqrt(t)
		}
	default:
		// Default: moderate flat response
		for i := range pwl.ControlValues {
			pwl.ControlValues[i] = 0.5
		}
	}
}

// ComputeTraits evaluates all trait functions given current context
func (n *CognitivePWLNetwork) ComputeTraits(context map[string]float64) map[string]float64 {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ContextFeatures = context
	traits := make(map[string]float64)

	for traitName, contextFunctions := range n.TraitFunctions {
		// Weighted average of all context contributions
		totalWeight := 0.0
		weightedSum := 0.0

		for ctxName, pwlFunc := range contextFunctions {
			ctxValue, exists := context[ctxName]
			if !exists {
				ctxValue = 0.5 // Default neutral context
			}
			output := pwlFunc.Evaluate(ctxValue)
			weight := 1.0 // Could be learned too
			weightedSum += output * weight
			totalWeight += weight
		}

		if totalWeight > 0 {
			traits[traitName] = weightedSum / totalWeight
		} else {
			traits[traitName] = 0.5
		}
	}

	n.ComputedTraits = traits
	return traits
}

// Learn applies a reward signal to update the PWL functions
func (n *CognitivePWLNetwork) Learn(signal RewardSignal) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.RewardHistory = append(n.RewardHistory, signal)
	if len(n.RewardHistory) > 1000 {
		n.RewardHistory = n.RewardHistory[len(n.RewardHistory)-500:]
	}

	// Backpropagate reward to PWL functions
	for traitName, contextFunctions := range n.TraitFunctions {
		traitValue, exists := signal.Traits[traitName]
		if !exists {
			continue
		}

		for ctxName, pwlFunc := range contextFunctions {
			ctxValue, exists := signal.Context[ctxName]
			if !exists {
				ctxValue = 0.5
			}

			// Gradient: if reward is positive and trait was active, reinforce
			// If reward is negative and trait was active, diminish
			currentOutput := pwlFunc.Evaluate(ctxValue)
			gradient := signal.Reward * (traitValue - currentOutput) * 0.1
			pwlFunc.AccumulateGradient(ctxValue, gradient)
		}
	}

	// Apply accumulated gradients
	for _, contextFunctions := range n.TraitFunctions {
		for _, pwlFunc := range contextFunctions {
			pwlFunc.ApplyGradients()
		}
	}

	n.TotalUpdates++
	n.LastUpdate = time.Now()
	n.WisdomAccrued += math.Abs(signal.Reward) * 0.01
}

// GetWisdomMetrics returns metrics about the network's learning progress
func (n *CognitivePWLNetwork) GetWisdomMetrics() map[string]interface{} {
	n.mu.RLock()
	defer n.mu.RUnlock()

	avgReward := 0.0
	if len(n.RewardHistory) > 0 {
		for _, r := range n.RewardHistory {
			avgReward += r.Reward
		}
		avgReward /= float64(len(n.RewardHistory))
	}

	return map[string]interface{}{
		"total_updates":   n.TotalUpdates,
		"wisdom_accrued":  fmt.Sprintf("%.4f", n.WisdomAccrued),
		"reward_history":  len(n.RewardHistory),
		"avg_reward":      fmt.Sprintf("%.4f", avgReward),
		"last_update":     n.LastUpdate.Format(time.RFC3339),
		"trait_count":     len(n.TraitFunctions),
		"context_count":   len(n.ContextFeatures),
	}
}

// Serialize returns a snapshot of the network state for persistence
func (n *CognitivePWLNetwork) Serialize() *PWLNetworkSnapshot {
	n.mu.RLock()
	defer n.mu.RUnlock()

	snapshot := &PWLNetworkSnapshot{
		Timestamp:     time.Now(),
		TotalUpdates:  n.TotalUpdates,
		WisdomAccrued: n.WisdomAccrued,
		Traits:        make(map[string]map[string]PWLSnapshot),
	}

	for traitName, contextFunctions := range n.TraitFunctions {
		snapshot.Traits[traitName] = make(map[string]PWLSnapshot)
		for ctxName, pwlFunc := range contextFunctions {
			snapshot.Traits[traitName][ctxName] = PWLSnapshot{
				SplitPoints:   append([]float64{}, pwlFunc.SplitPoints...),
				ControlValues: append([]float64{}, pwlFunc.ControlValues...),
			}
		}
	}

	return snapshot
}

// PWLNetworkSnapshot is a serializable snapshot of the network
type PWLNetworkSnapshot struct {
	Timestamp     time.Time
	TotalUpdates  uint64
	WisdomAccrued float64
	Traits        map[string]map[string]PWLSnapshot
}

// PWLSnapshot is a serializable snapshot of a single PWL function
type PWLSnapshot struct {
	SplitPoints   []float64
	ControlValues []float64
}
