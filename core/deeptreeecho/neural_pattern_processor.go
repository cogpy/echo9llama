package deeptreeecho

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

// NeuralPatternProcessor provides neural network-inspired pattern processing
// This is inspired by GoMLX's tensor and graph computation approach
type NeuralPatternProcessor struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc

	// Pattern storage
	patterns        map[string]*CognitivePattern
	patternIndex    map[string][]string // category -> pattern IDs

	// Neural layers
	layers          []*NeuralLayer

	// Learning parameters
	learningRate    float64
	momentum        float64
	decayRate       float64

	// Activation history
	activations     []*PatternActivation
	activationLimit int

	// Metrics
	totalPatterns   uint64
	totalActivations uint64
	totalLearnings  uint64

	// Running state
	running         bool
}

// CognitivePattern represents a learned pattern
type CognitivePattern struct {
	ID          string
	Category    string
	Name        string
	Features    []float64
	Weights     []float64
	Bias        float64
	Strength    float64
	Frequency   uint64
	LastActive  time.Time
	CreatedAt   time.Time
	Metadata    map[string]interface{}
}

// NeuralLayer represents a layer in the neural processing pipeline
type NeuralLayer struct {
	ID          string
	Name        string
	Type        NeuralLayerType
	InputSize   int
	OutputSize  int
	Weights     [][]float64
	Biases      []float64
	Activation  ActivationFunc
}

// NeuralLayerType represents the type of neural layer
type NeuralLayerType string

const (
	NeuralLayerDense       NeuralLayerType = "dense"
	NeuralLayerAttention   NeuralLayerType = "attention"
	NeuralLayerRecurrent   NeuralLayerType = "recurrent"
	NeuralLayerConvolution NeuralLayerType = "convolution"
)

// ActivationFunc represents an activation function
type ActivationFunc string

const (
	ActivationReLU    ActivationFunc = "relu"
	ActivationSigmoid ActivationFunc = "sigmoid"
	ActivationTanh    ActivationFunc = "tanh"
	ActivationSoftmax ActivationFunc = "softmax"
	ActivationGELU    ActivationFunc = "gelu"
)

// PatternActivation records when a pattern was activated
type PatternActivation struct {
	PatternID   string
	Timestamp   time.Time
	Strength    float64
	Context     string
	Input       []float64
	Output      []float64
}

// NewNeuralPatternProcessor creates a new neural pattern processor
func NewNeuralPatternProcessor() *NeuralPatternProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	npp := &NeuralPatternProcessor{
		ctx:             ctx,
		cancel:          cancel,
		patterns:        make(map[string]*CognitivePattern),
		patternIndex:    make(map[string][]string),
		layers:          make([]*NeuralLayer, 0),
		learningRate:    0.01,
		momentum:        0.9,
		decayRate:       0.001,
		activations:     make([]*PatternActivation, 0),
		activationLimit: 1000,
	}

	return npp
}

// Start begins the neural pattern processor
func (npp *NeuralPatternProcessor) Start() error {
	npp.mu.Lock()
	defer npp.mu.Unlock()

	if npp.running {
		return fmt.Errorf("neural pattern processor already running")
	}

	npp.running = true
	fmt.Println("🧬 Neural Pattern Processor started")

	// Initialize default layers for cognitive processing
	npp.initializeDefaultLayers()

	// Initialize default pattern categories
	npp.initializeDefaultCategories()

	return nil
}

// Stop gracefully stops the neural pattern processor
func (npp *NeuralPatternProcessor) Stop() error {
	npp.mu.Lock()
	defer npp.mu.Unlock()

	if !npp.running {
		return fmt.Errorf("neural pattern processor not running")
	}

	npp.cancel()
	npp.running = false
	fmt.Println("🧬 Neural Pattern Processor stopped")

	return nil
}

// initializeDefaultLayers sets up the default neural layers
func (npp *NeuralPatternProcessor) initializeDefaultLayers() {
	// Perception layer - processes raw input
	npp.AddLayer(&NeuralLayer{
		ID:         "layer_perception",
		Name:       "Perception",
		Type:       NeuralLayerDense,
		InputSize:  128,
		OutputSize: 64,
		Activation: ActivationReLU,
	})

	// Attention layer - focuses on relevant features
	npp.AddLayer(&NeuralLayer{
		ID:         "layer_attention",
		Name:       "Attention",
		Type:       NeuralLayerAttention,
		InputSize:  64,
		OutputSize: 64,
		Activation: ActivationSoftmax,
	})

	// Integration layer - combines features
	npp.AddLayer(&NeuralLayer{
		ID:         "layer_integration",
		Name:       "Integration",
		Type:       NeuralLayerDense,
		InputSize:  64,
		OutputSize: 32,
		Activation: ActivationGELU,
	})

	// Pattern layer - recognizes patterns
	npp.AddLayer(&NeuralLayer{
		ID:         "layer_pattern",
		Name:       "Pattern",
		Type:       NeuralLayerDense,
		InputSize:  32,
		OutputSize: 16,
		Activation: ActivationTanh,
	})

	fmt.Printf("   Initialized %d neural layers\n", len(npp.layers))
}

// initializeDefaultCategories sets up default pattern categories
func (npp *NeuralPatternProcessor) initializeDefaultCategories() {
	categories := []string{
		"perception",
		"action",
		"simulation",
		"emotion",
		"memory",
		"reasoning",
		"social",
		"creative",
		"wisdom",
	}

	for _, cat := range categories {
		npp.patternIndex[cat] = make([]string, 0)
	}

	fmt.Printf("   Initialized %d pattern categories\n", len(categories))
}

// AddLayer adds a neural layer to the processor
func (npp *NeuralPatternProcessor) AddLayer(layer *NeuralLayer) {
	// Initialize weights if not set
	if layer.Weights == nil {
		layer.Weights = npp.initializeWeights(layer.InputSize, layer.OutputSize)
	}
	if layer.Biases == nil {
		layer.Biases = make([]float64, layer.OutputSize)
	}

	npp.layers = append(npp.layers, layer)
}

// initializeWeights creates Xavier-initialized weights
func (npp *NeuralPatternProcessor) initializeWeights(inputSize, outputSize int) [][]float64 {
	weights := make([][]float64, outputSize)
	scale := math.Sqrt(2.0 / float64(inputSize+outputSize))

	for i := 0; i < outputSize; i++ {
		weights[i] = make([]float64, inputSize)
		for j := 0; j < inputSize; j++ {
			// Simple pseudo-random initialization
			weights[i][j] = (float64((i*inputSize+j)%1000)/1000.0 - 0.5) * scale
		}
	}

	return weights
}

// LearnPattern learns a new pattern from input
func (npp *NeuralPatternProcessor) LearnPattern(category, name string, features []float64) *CognitivePattern {
	npp.mu.Lock()
	defer npp.mu.Unlock()

	patternID := fmt.Sprintf("pattern_%s", uuid.New().String()[:8])

	pattern := &CognitivePattern{
		ID:         patternID,
		Category:   category,
		Name:       name,
		Features:   features,
		Weights:    npp.computePatternWeights(features),
		Bias:       0.0,
		Strength:   1.0,
		Frequency:  1,
		LastActive: time.Now(),
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	npp.patterns[patternID] = pattern
	npp.patternIndex[category] = append(npp.patternIndex[category], patternID)
	npp.totalPatterns++
	npp.totalLearnings++

	return pattern
}

// computePatternWeights computes weights for a pattern
func (npp *NeuralPatternProcessor) computePatternWeights(features []float64) []float64 {
	weights := make([]float64, len(features))
	sum := 0.0

	for _, f := range features {
		sum += math.Abs(f)
	}

	if sum > 0 {
		for i, f := range features {
			weights[i] = f / sum
		}
	}

	return weights
}

// RecognizePattern attempts to recognize a pattern from input
func (npp *NeuralPatternProcessor) RecognizePattern(input []float64) (*CognitivePattern, float64) {
	npp.mu.RLock()
	defer npp.mu.RUnlock()

	var bestPattern *CognitivePattern
	bestScore := 0.0

	for _, pattern := range npp.patterns {
		score := npp.computeSimilarity(input, pattern.Features)
		if score > bestScore {
			bestScore = score
			bestPattern = pattern
		}
	}

	if bestPattern != nil && bestScore > 0.5 {
		// Record activation
		npp.recordActivation(bestPattern.ID, bestScore, input)
		return bestPattern, bestScore
	}

	return nil, 0.0
}

// computeSimilarity computes cosine similarity between two vectors
func (npp *NeuralPatternProcessor) computeSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Pad shorter vector
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	aPadded := make([]float64, maxLen)
	bPadded := make([]float64, maxLen)
	copy(aPadded, a)
	copy(bPadded, b)

	// Compute cosine similarity
	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := 0; i < maxLen; i++ {
		dotProduct += aPadded[i] * bPadded[i]
		normA += aPadded[i] * aPadded[i]
		normB += bPadded[i] * bPadded[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// recordActivation records a pattern activation
func (npp *NeuralPatternProcessor) recordActivation(patternID string, strength float64, input []float64) {
	activation := &PatternActivation{
		PatternID: patternID,
		Timestamp: time.Now(),
		Strength:  strength,
		Input:     input,
	}

	npp.activations = append(npp.activations, activation)
	if len(npp.activations) > npp.activationLimit {
		npp.activations = npp.activations[1:]
	}

	npp.totalActivations++

	// Update pattern frequency
	if pattern, exists := npp.patterns[patternID]; exists {
		pattern.Frequency++
		pattern.LastActive = time.Now()
	}
}

// Forward performs forward propagation through all layers
func (npp *NeuralPatternProcessor) Forward(input []float64) []float64 {
	npp.mu.RLock()
	defer npp.mu.RUnlock()

	current := input

	for _, layer := range npp.layers {
		current = npp.forwardLayer(layer, current)
	}

	return current
}

// forwardLayer performs forward propagation through a single layer
func (npp *NeuralPatternProcessor) forwardLayer(layer *NeuralLayer, input []float64) []float64 {
	// Pad or truncate input to match layer input size
	paddedInput := make([]float64, layer.InputSize)
	for i := 0; i < len(paddedInput) && i < len(input); i++ {
		paddedInput[i] = input[i]
	}

	output := make([]float64, layer.OutputSize)

	// Matrix multiplication
	for i := 0; i < layer.OutputSize; i++ {
		sum := layer.Biases[i]
		for j := 0; j < layer.InputSize; j++ {
			sum += layer.Weights[i][j] * paddedInput[j]
		}
		output[i] = sum
	}

	// Apply activation function
	return npp.applyActivation(output, layer.Activation)
}

// applyActivation applies an activation function to a vector
func (npp *NeuralPatternProcessor) applyActivation(x []float64, activation ActivationFunc) []float64 {
	result := make([]float64, len(x))

	switch activation {
	case ActivationReLU:
		for i, v := range x {
			if v > 0 {
				result[i] = v
			} else {
				result[i] = 0
			}
		}
	case ActivationSigmoid:
		for i, v := range x {
			result[i] = 1.0 / (1.0 + math.Exp(-v))
		}
	case ActivationTanh:
		for i, v := range x {
			result[i] = math.Tanh(v)
		}
	case ActivationSoftmax:
		maxVal := x[0]
		for _, v := range x {
			if v > maxVal {
				maxVal = v
			}
		}
		sum := 0.0
		for i, v := range x {
			result[i] = math.Exp(v - maxVal)
			sum += result[i]
		}
		for i := range result {
			result[i] /= sum
		}
	case ActivationGELU:
		for i, v := range x {
			// Approximate GELU
			result[i] = 0.5 * v * (1.0 + math.Tanh(math.Sqrt(2.0/math.Pi)*(v+0.044715*v*v*v)))
		}
	default:
		copy(result, x)
	}

	return result
}

// ReinforceLearning reinforces a pattern based on feedback
func (npp *NeuralPatternProcessor) ReinforceLearning(patternID string, reward float64) error {
	npp.mu.Lock()
	defer npp.mu.Unlock()

	pattern, exists := npp.patterns[patternID]
	if !exists {
		return fmt.Errorf("pattern not found: %s", patternID)
	}

	// Update pattern strength based on reward
	pattern.Strength += npp.learningRate * reward * (1.0 - pattern.Strength)

	// Decay other patterns slightly
	for id, p := range npp.patterns {
		if id != patternID {
			p.Strength *= (1.0 - npp.decayRate)
		}
	}

	return nil
}

// GetPatternsByCategory returns patterns in a category
func (npp *NeuralPatternProcessor) GetPatternsByCategory(category string) []*CognitivePattern {
	npp.mu.RLock()
	defer npp.mu.RUnlock()

	patternIDs, exists := npp.patternIndex[category]
	if !exists {
		return nil
	}

	patterns := make([]*CognitivePattern, 0, len(patternIDs))
	for _, id := range patternIDs {
		if pattern, exists := npp.patterns[id]; exists {
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

// GetStrongestPatterns returns the N strongest patterns
func (npp *NeuralPatternProcessor) GetStrongestPatterns(n int) []*CognitivePattern {
	npp.mu.RLock()
	defer npp.mu.RUnlock()

	// Collect all patterns
	allPatterns := make([]*CognitivePattern, 0, len(npp.patterns))
	for _, p := range npp.patterns {
		allPatterns = append(allPatterns, p)
	}

	// Sort by strength (simple bubble sort for small N)
	for i := 0; i < len(allPatterns)-1; i++ {
		for j := i + 1; j < len(allPatterns); j++ {
			if allPatterns[j].Strength > allPatterns[i].Strength {
				allPatterns[i], allPatterns[j] = allPatterns[j], allPatterns[i]
			}
		}
	}

	if n > len(allPatterns) {
		n = len(allPatterns)
	}

	return allPatterns[:n]
}

// GetMetrics returns processor metrics
func (npp *NeuralPatternProcessor) GetMetrics() map[string]interface{} {
	npp.mu.RLock()
	defer npp.mu.RUnlock()

	return map[string]interface{}{
		"running":           npp.running,
		"total_patterns":    npp.totalPatterns,
		"total_activations": npp.totalActivations,
		"total_learnings":   npp.totalLearnings,
		"layer_count":       len(npp.layers),
		"category_count":    len(npp.patternIndex),
		"learning_rate":     npp.learningRate,
	}
}

// ContributeToGestalt provides processor state for the global gestalt
func (npp *NeuralPatternProcessor) ContributeToGestalt() map[string]interface{} {
	npp.mu.RLock()
	defer npp.mu.RUnlock()

	// Get strongest patterns
	strongestPatterns := make([]map[string]interface{}, 0)
	for _, p := range npp.GetStrongestPatterns(5) {
		strongestPatterns = append(strongestPatterns, map[string]interface{}{
			"id":       p.ID,
			"name":     p.Name,
			"category": p.Category,
			"strength": p.Strength,
		})
	}

	return map[string]interface{}{
		"running":            npp.running,
		"total_patterns":     npp.totalPatterns,
		"total_activations":  npp.totalActivations,
		"layer_count":        len(npp.layers),
		"strongest_patterns": strongestPatterns,
	}
}
