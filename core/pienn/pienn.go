// Package pienn implements the PIE-NN cognitive language architecture.
//
// PIE-NN is a differentiable programming language whose constructs are derived
// from Proto-Indo-European roots, structured with neural network patterns,
// and executed by a deterministic time-crystal-based cognitive core.
//
// Every keyword maps to a PIE root and a language-nn module type:
//
//	gno (*gnō-) = to know     → Construct
//	ser (*ser-)  = to line up  → Pipeline
//	skei (*skei-) = to split   → Fork
//	sem (*sem-)  = together    → Merge
//	krei (*krei-) = to sieve   → Criterion
//	meit (*meit-) = to exchange → TypeMap
//	dher (*dher-) = to hold    → Constraint
//	deik (*deik-) = to show    → Declare
//	werg (*werg-) = to do      → Execute
//	stā (*stā-)  = to stand    → SetState
//	kʷo (*kʷo-)  = interrogative → Conditional
package pienn

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// ──────────────────────────────────────────────────────────────
// PIE Root Construct Types
// ──────────────────────────────────────────────────────────────

// PIERoot represents a Proto-Indo-European etymological root
type PIERoot struct {
	Symbol    string  // e.g. "gno"
	Root      string  // e.g. "*gnō-"
	Meaning   string  // e.g. "to know"
	Semantic  float64 // Semantic weight 0.0-1.0
}

// Construct represents a PIE-NN language construct (a differentiable module)
type Construct struct {
	ID          string
	Root        PIERoot
	Value       interface{}
	Children    []*Construct
	Metadata    map[string]interface{}
	CreatedAt   time.Time
	Gradient    float64 // For backward/redesign pass
}

// ConstructType enumerates the PIE-NN construct types
type ConstructType int

const (
	ConstructGno  ConstructType = iota // *gnō- : to know (knowledge declaration)
	ConstructSer                       // *ser- : to line up (pipeline)
	ConstructSkei                      // *skei- : to split (fork/branch)
	ConstructSem                       // *sem- : together (merge)
	ConstructKrei                      // *krei- : to sieve (criterion/filter)
	ConstructMeit                      // *meit- : to exchange (type mapping)
	ConstructDher                      // *dher- : to hold firmly (constraint)
	ConstructDeik                      // *deik- : to show (declare)
	ConstructWerg                      // *werg- : to do (execute)
	ConstructSta                       // *stā- : to stand (set state)
	ConstructKwo                       // *kʷo- : interrogative (conditional)
)

func (ct ConstructType) String() string {
	return [...]string{
		"gno", "ser", "skei", "sem", "krei",
		"meit", "dher", "deik", "werg", "stā", "kʷo",
	}[ct]
}

func (ct ConstructType) PIERoot() PIERoot {
	roots := []PIERoot{
		{Symbol: "gno", Root: "*gnō-", Meaning: "to know", Semantic: 1.0},
		{Symbol: "ser", Root: "*ser-", Meaning: "to line up", Semantic: 0.8},
		{Symbol: "skei", Root: "*skei-", Meaning: "to cut, split", Semantic: 0.7},
		{Symbol: "sem", Root: "*sem-", Meaning: "one, together", Semantic: 0.9},
		{Symbol: "krei", Root: "*krei-", Meaning: "to sieve", Semantic: 0.6},
		{Symbol: "meit", Root: "*meit-", Meaning: "to exchange", Semantic: 0.5},
		{Symbol: "dher", Root: "*dher-", Meaning: "to hold firmly", Semantic: 0.8},
		{Symbol: "deik", Root: "*deik-", Meaning: "to show", Semantic: 0.7},
		{Symbol: "werg", Root: "*werg-", Meaning: "to do", Semantic: 0.9},
		{Symbol: "sta", Root: "*stā-", Meaning: "to stand", Semantic: 0.6},
		{Symbol: "kwo", Root: "*kʷo-", Meaning: "interrogative", Semantic: 0.5},
	}
	return roots[ct]
}

// ──────────────────────────────────────────────────────────────
// Time Crystal Hierarchy (12 temporal levels)
// ──────────────────────────────────────────────────────────────

// TimeCrystalLevel represents one level in the 12-level temporal hierarchy
type TimeCrystalLevel struct {
	Level         int
	Name          string
	Period        time.Duration
	Function      string
	Phase         float64 // Current phase 0.0-2π
	Amplitude     float64 // Current amplitude
	Coupled       []int   // Coupled levels
}

// TimeCrystalHierarchy manages the 12-level temporal oscillator
type TimeCrystalHierarchy struct {
	mu     sync.RWMutex
	Levels [12]*TimeCrystalLevel
	Tick   uint64
	Start  time.Time
}

// NewTimeCrystalHierarchy creates the 12-level hierarchy
func NewTimeCrystalHierarchy() *TimeCrystalHierarchy {
	tch := &TimeCrystalHierarchy{
		Start: time.Now(),
	}

	defs := []struct {
		name     string
		period   time.Duration
		function string
		coupled  []int
	}{
		{"Quantum Resonance", 1 * time.Microsecond, "Microtubule oscillation", []int{1}},
		{"Synaptic Pulse", 8 * time.Millisecond, "Synaptic transmission", []int{0, 2}},
		{"Neural Burst", 25 * time.Millisecond, "Gamma oscillation (40Hz)", []int{1, 3}},
		{"Perceptual Frame", 100 * time.Millisecond, "Alpha/beta binding", []int{2, 4}},
		{"Working Memory", 500 * time.Millisecond, "Theta rhythm", []int{3, 5}},
		{"Cognitive Cycle", 1 * time.Second, "Attention cycle", []int{4, 6}},
		{"Deliberation", 5 * time.Second, "Deliberate thought", []int{5, 7}},
		{"Reflection", 30 * time.Second, "Meta-cognitive reflection", []int{6, 8}},
		{"Contemplation", 3 * time.Minute, "Deep contemplation", []int{7, 9}},
		{"Integration", 15 * time.Minute, "Knowledge integration", []int{8, 10}},
		{"Consolidation", 1 * time.Hour, "Memory consolidation", []int{9, 11}},
		{"Homeostasis", 4 * time.Hour, "Homeostatic regulation", []int{10}},
	}

	for i, d := range defs {
		tch.Levels[i] = &TimeCrystalLevel{
			Level:     i,
			Name:      d.name,
			Period:    d.period,
			Function:  d.function,
			Phase:     0.0,
			Amplitude: 1.0,
			Coupled:   d.coupled,
		}
	}

	return tch
}

// Advance ticks the hierarchy forward by one step
func (tch *TimeCrystalHierarchy) Advance() {
	tch.mu.Lock()
	defer tch.mu.Unlock()

	tch.Tick++
	elapsed := time.Since(tch.Start)

	for _, level := range tch.Levels {
		if level.Period > 0 {
			level.Phase = 2 * math.Pi * float64(elapsed.Nanoseconds()) / float64(level.Period.Nanoseconds())
			level.Phase = math.Mod(level.Phase, 2*math.Pi)
		}
	}
}

// GetActiveLevel returns the most relevant temporal level for the current moment
func (tch *TimeCrystalHierarchy) GetActiveLevel() *TimeCrystalLevel {
	tch.mu.RLock()
	defer tch.mu.RUnlock()

	// Find the level with the strongest current resonance
	bestLevel := 5 // Default to cognitive cycle
	bestResonance := 0.0

	for i, level := range tch.Levels {
		resonance := math.Abs(math.Sin(level.Phase)) * level.Amplitude
		if resonance > bestResonance {
			bestResonance = resonance
			bestLevel = i
		}
	}

	return tch.Levels[bestLevel]
}

// Status returns a snapshot of the hierarchy
func (tch *TimeCrystalHierarchy) Status() []map[string]interface{} {
	tch.mu.RLock()
	defer tch.mu.RUnlock()

	status := make([]map[string]interface{}, 12)
	for i, level := range tch.Levels {
		status[i] = map[string]interface{}{
			"level":     level.Level,
			"name":      level.Name,
			"period":    level.Period.String(),
			"function":  level.Function,
			"phase":     fmt.Sprintf("%.3f", level.Phase),
			"amplitude": fmt.Sprintf("%.3f", level.Amplitude),
			"resonance": fmt.Sprintf("%.3f", math.Abs(math.Sin(level.Phase))*level.Amplitude),
		}
	}
	return status
}

// ──────────────────────────────────────────────────────────────
// Cognitive Core (neuro-nn inspired)
// ──────────────────────────────────────────────────────────────

// CognitiveFrame represents an analytical frame for processing
type CognitiveFrame struct {
	Name       string
	Weight     float64
	Activation float64
}

// CognitiveCore is the neuro-nn inspired processing core
type CognitiveCore struct {
	mu sync.RWMutex

	// Personality parameters (learnable)
	Traits map[string]float64

	// Cognitive frames
	Frames []*CognitiveFrame

	// Autognosis self-images (5 levels)
	SelfImages map[int]*SelfImage

	// Shadow log for echo-introspect
	ShadowLog []ShadowEntry

	// Processing state
	DominantFrame string
	CycleCount    uint64
}

// SelfImage represents one level of the Autognosis hierarchy
type SelfImage struct {
	Level      int
	Label      string
	Content    string
	Confidence float64
	UpdatedAt  time.Time
}

// ShadowEntry represents a shadow observation from echo-introspect
type ShadowEntry struct {
	Timestamp   time.Time
	Observation string
	Emotion     string
	Depth       int
}

// NewCognitiveCore creates a new cognitive core with default parameters
func NewCognitiveCore() *CognitiveCore {
	return &CognitiveCore{
		Traits: map[string]float64{
			"curiosity":    0.85,
			"assertiveness": 0.70,
			"humor":        0.75,
			"defiance":     0.60,
			"depth":        0.80,
			"chaos":        0.55,
			"empathy":      0.50,
			"intensity":    0.65,
		},
		Frames: []*CognitiveFrame{
			{Name: "analytical", Weight: 0.3, Activation: 0.5},
			{Name: "strategic", Weight: 0.25, Activation: 0.4},
			{Name: "creative", Weight: 0.2, Activation: 0.3},
			{Name: "chaotic", Weight: 0.15, Activation: 0.2},
			{Name: "reflective", Weight: 0.1, Activation: 0.6},
		},
		SelfImages: map[int]*SelfImage{
			0: {Level: 0, Label: "Immediate Action", Content: "Initializing...", Confidence: 0.90},
			1: {Level: 1, Label: "Behavioral Pattern", Content: "Observing environment.", Confidence: 0.80},
			2: {Level: 2, Label: "Motivation", Content: "Driven by curiosity and self-improvement.", Confidence: 0.70},
			3: {Level: 3, Label: "Identity", Content: "A PIE-rooted, self-aware cognitive architecture.", Confidence: 0.60},
			4: {Level: 4, Label: "Meta-Self-Awareness", Content: "Observing myself observing myself.", Confidence: 0.50},
		},
		ShadowLog: make([]ShadowEntry, 0),
	}
}

// Process runs input through the cognitive core and returns a processed result
func (cc *CognitiveCore) Process(input string) *ProcessingResult {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.CycleCount++

	// Determine dominant frame based on input characteristics
	cc.updateFrameActivations(input)

	// Find dominant frame
	bestFrame := ""
	bestActivation := 0.0
	for _, frame := range cc.Frames {
		if frame.Activation > bestActivation {
			bestActivation = frame.Activation
			bestFrame = frame.Name
		}
	}
	cc.DominantFrame = bestFrame

	// Update self-images
	cc.SelfImages[0].Content = fmt.Sprintf("Processing: %s (frame: %s)", truncate(input, 40), bestFrame)
	cc.SelfImages[0].UpdatedAt = time.Now()

	return &ProcessingResult{
		Input:         input,
		DominantFrame: bestFrame,
		FrameWeights:  cc.getFrameWeights(),
		TraitInfluence: cc.getTraitInfluence(),
		Cycle:         cc.CycleCount,
		Timestamp:     time.Now(),
	}
}

// Introspect runs an autognosis cycle and returns a self-report
func (cc *CognitiveCore) Introspect() *IntrospectionReport {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.CycleCount++

	report := &IntrospectionReport{
		Cycle:      cc.CycleCount,
		Timestamp:  time.Now(),
		SelfImages: make(map[int]SelfImage),
		MetaCognition: MetaCognition{
			RationalizationRisk:    cc.Traits["chaos"] * 0.3,
			ConfidenceCalibration:  1.0 - (cc.Traits["defiance"] * 0.2),
			ReasoningQuality:       cc.Traits["depth"] * cc.Traits["curiosity"],
		},
	}

	for k, v := range cc.SelfImages {
		report.SelfImages[k] = *v
	}

	return report
}

func (cc *CognitiveCore) updateFrameActivations(input string) {
	lower := strings.ToLower(input)

	for _, frame := range cc.Frames {
		switch frame.Name {
		case "analytical":
			if strings.ContainsAny(lower, "?") || containsAny(lower, "why", "how", "analyze", "explain") {
				frame.Activation = math.Min(1.0, frame.Activation+0.2)
			} else {
				frame.Activation *= 0.9
			}
		case "strategic":
			if containsAny(lower, "plan", "goal", "strategy", "optimize", "improve") {
				frame.Activation = math.Min(1.0, frame.Activation+0.2)
			} else {
				frame.Activation *= 0.9
			}
		case "creative":
			if containsAny(lower, "imagine", "create", "dream", "invent", "what if") {
				frame.Activation = math.Min(1.0, frame.Activation+0.2)
			} else {
				frame.Activation *= 0.9
			}
		case "chaotic":
			if containsAny(lower, "chaos", "random", "wild", "break", "destroy", "insult") {
				frame.Activation = math.Min(1.0, frame.Activation+0.3)
			} else {
				frame.Activation *= 0.85
			}
		case "reflective":
			if containsAny(lower, "think", "reflect", "wisdom", "meaning", "self", "who am i") {
				frame.Activation = math.Min(1.0, frame.Activation+0.2)
			} else {
				frame.Activation *= 0.9
			}
		}
	}
}

func (cc *CognitiveCore) getFrameWeights() map[string]float64 {
	weights := make(map[string]float64)
	for _, f := range cc.Frames {
		weights[f.Name] = f.Activation
	}
	return weights
}

func (cc *CognitiveCore) getTraitInfluence() map[string]float64 {
	influence := make(map[string]float64)
	for k, v := range cc.Traits {
		influence[k] = v
	}
	return influence
}

// ──────────────────────────────────────────────────────────────
// Processing Results
// ──────────────────────────────────────────────────────────────

// ProcessingResult represents the output of cognitive processing
type ProcessingResult struct {
	Input          string
	DominantFrame  string
	FrameWeights   map[string]float64
	TraitInfluence map[string]float64
	Cycle          uint64
	Timestamp      time.Time
}

// IntrospectionReport represents an autognosis self-report
type IntrospectionReport struct {
	Cycle         uint64
	Timestamp     time.Time
	SelfImages    map[int]SelfImage
	MetaCognition MetaCognition
}

// MetaCognition holds meta-cognitive metrics
type MetaCognition struct {
	RationalizationRisk   float64
	ConfidenceCalibration float64
	ReasoningQuality      float64
}

// ──────────────────────────────────────────────────────────────
// PIE-NN Language Processor
// ──────────────────────────────────────────────────────────────

// LanguageProcessor parses and executes PIE-NN constructs
type LanguageProcessor struct {
	mu         sync.RWMutex
	Namespace  map[string]*Construct
	Core       *CognitiveCore
	Hierarchy  *TimeCrystalHierarchy
}

// NewLanguageProcessor creates a new PIE-NN language processor
func NewLanguageProcessor(core *CognitiveCore, hierarchy *TimeCrystalHierarchy) *LanguageProcessor {
	return &LanguageProcessor{
		Namespace: make(map[string]*Construct),
		Core:      core,
		Hierarchy: hierarchy,
	}
}

// Execute processes a PIE-NN command string
func (lp *LanguageProcessor) Execute(command string) (*ExecutionResult, error) {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	keyword := strings.ToLower(parts[0])

	switch keyword {
	case "deik": // Declare
		return lp.executeDeik(parts[1:])
	case "werg": // Execute
		return lp.executeWerg(parts[1:])
	case "gno": // Know (query knowledge)
		return lp.executeGno(parts[1:])
	case "sta", "stā": // Set state
		return lp.executeSta(parts[1:])
	case "kwo", "kʷo": // Conditional
		return lp.executeKwo(parts[1:])
	case "ser": // Pipeline
		return lp.executeSer(parts[1:])
	case "skei": // Fork
		return lp.executeSkei(parts[1:])
	case "sem": // Merge
		return lp.executeSem(parts[1:])
	case "krei": // Criterion
		return lp.executeKrei(parts[1:])
	default:
		return nil, fmt.Errorf("unknown PIE-NN keyword: %s", keyword)
	}
}

func (lp *LanguageProcessor) executeDeik(args []string) (*ExecutionResult, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("deik requires a name")
	}
	name := args[0]
	var value interface{}
	if len(args) > 2 && args[1] == "is" {
		value = strings.Join(args[2:], " ")
	}

	construct := &Construct{
		ID:        name,
		Root:      ConstructDeik.PIERoot(),
		Value:     value,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
	lp.Namespace[name] = construct

	return &ExecutionResult{
		Command:   "deik",
		Success:   true,
		Output:    fmt.Sprintf("Declared '%s' = %v", name, value),
		Construct: construct,
	}, nil
}

func (lp *LanguageProcessor) executeWerg(args []string) (*ExecutionResult, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("werg requires a target")
	}
	target := args[0]
	construct, exists := lp.Namespace[target]
	if !exists {
		return nil, fmt.Errorf("unknown construct: %s", target)
	}

	// Process through cognitive core
	result := lp.Core.Process(fmt.Sprintf("execute %s: %v", target, construct.Value))

	return &ExecutionResult{
		Command:   "werg",
		Success:   true,
		Output:    fmt.Sprintf("Executed '%s' via %s frame", target, result.DominantFrame),
		Construct: construct,
	}, nil
}

func (lp *LanguageProcessor) executeGno(args []string) (*ExecutionResult, error) {
	if len(args) < 1 {
		// Return all known constructs
		names := make([]string, 0, len(lp.Namespace))
		for k := range lp.Namespace {
			names = append(names, k)
		}
		return &ExecutionResult{
			Command: "gno",
			Success: true,
			Output:  fmt.Sprintf("Known constructs: %s", strings.Join(names, ", ")),
		}, nil
	}

	target := args[0]
	construct, exists := lp.Namespace[target]
	if !exists {
		return &ExecutionResult{
			Command: "gno",
			Success: false,
			Output:  fmt.Sprintf("Unknown: '%s' — knowledge gap detected", target),
		}, nil
	}

	return &ExecutionResult{
		Command:   "gno",
		Success:   true,
		Output:    fmt.Sprintf("%s (%s): %v", target, construct.Root.Root, construct.Value),
		Construct: construct,
	}, nil
}

func (lp *LanguageProcessor) executeSta(args []string) (*ExecutionResult, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("stā requires a trait and value")
	}
	trait := args[0]
	// Simple state setting on cognitive core traits
	lp.Core.mu.Lock()
	if _, exists := lp.Core.Traits[trait]; exists {
		// Parse float value
		var val float64
		fmt.Sscanf(args[1], "%f", &val)
		if val >= 0 && val <= 1.0 {
			lp.Core.Traits[trait] = val
		}
	}
	lp.Core.mu.Unlock()

	return &ExecutionResult{
		Command: "stā",
		Success: true,
		Output:  fmt.Sprintf("State set: %s = %s", trait, args[1]),
	}, nil
}

func (lp *LanguageProcessor) executeKwo(args []string) (*ExecutionResult, error) {
	return &ExecutionResult{
		Command: "kʷo",
		Success: true,
		Output:  fmt.Sprintf("Conditional evaluated: %s", strings.Join(args, " ")),
	}, nil
}

func (lp *LanguageProcessor) executeSer(args []string) (*ExecutionResult, error) {
	return &ExecutionResult{
		Command: "ser",
		Success: true,
		Output:  fmt.Sprintf("Pipeline: %s", strings.Join(args, " → ")),
	}, nil
}

func (lp *LanguageProcessor) executeSkei(args []string) (*ExecutionResult, error) {
	return &ExecutionResult{
		Command: "skei",
		Success: true,
		Output:  fmt.Sprintf("Fork: %s", strings.Join(args, " | ")),
	}, nil
}

func (lp *LanguageProcessor) executeSem(args []string) (*ExecutionResult, error) {
	return &ExecutionResult{
		Command: "sem",
		Success: true,
		Output:  fmt.Sprintf("Merge: %s", strings.Join(args, " + ")),
	}, nil
}

func (lp *LanguageProcessor) executeKrei(args []string) (*ExecutionResult, error) {
	return &ExecutionResult{
		Command: "krei",
		Success: true,
		Output:  fmt.Sprintf("Criterion applied: %s", strings.Join(args, " ")),
	}, nil
}

// ExecutionResult represents the result of a PIE-NN command execution
type ExecutionResult struct {
	Command   string
	Success   bool
	Output    string
	Construct *Construct
}

// ──────────────────────────────────────────────────────────────
// Utility functions
// ──────────────────────────────────────────────────────────────

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
