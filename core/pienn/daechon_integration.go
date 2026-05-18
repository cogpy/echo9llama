// Package pienn - Daechon Integration
//
// Bridges the PIE-NN cognitive engine with the PWL-KAN network to provide
// context-adaptive personality for the daechon daemon. The cognitive core's
// static traits are replaced by dynamic PWL-computed traits that learn from
// every interaction.
package pienn

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// AdaptiveCognitiveCore wraps the base CognitiveCore with PWL-KAN adaptation
type AdaptiveCognitiveCore struct {
	mu sync.RWMutex

	// Base cognitive core (for fallback and structure)
	Base *CognitiveCore

	// PWL network for learnable traits
	Network *CognitivePWLNetwork

	// Context extractor
	contextHistory []ContextSnapshot
	maxHistory     int

	// Interaction tracking for reward signals
	interactionBuffer []InteractionRecord
	maxBuffer         int

	// Disposition tracking
	currentDisposition string
	dispositionHistory []DispositionRecord

	// Metrics
	TotalInteractions uint64
	TotalAdaptations  uint64
	WisdomLevel       float64
}

// ContextSnapshot captures the context at a point in time
type ContextSnapshot struct {
	Timestamp time.Time
	Features  map[string]float64
	Traits    map[string]float64
}

// InteractionRecord tracks an interaction for delayed reward
type InteractionRecord struct {
	Timestamp   time.Time
	Input       string
	Context     map[string]float64
	Traits      map[string]float64
	Response    string
	Disposition string
}

// DispositionRecord tracks disposition changes
type DispositionRecord struct {
	Timestamp   time.Time
	From        string
	To          string
	Trigger     string
}

// NewAdaptiveCognitiveCore creates a new adaptive core
func NewAdaptiveCognitiveCore() *AdaptiveCognitiveCore {
	return &AdaptiveCognitiveCore{
		Base:              NewCognitiveCore(),
		Network:           NewCognitivePWLNetwork(),
		contextHistory:    make([]ContextSnapshot, 0),
		maxHistory:        100,
		interactionBuffer: make([]InteractionRecord, 0),
		maxBuffer:         50,
		dispositionHistory: make([]DispositionRecord, 0),
		currentDisposition: "curious",
	}
}

// ExtractContext analyzes input to produce context features
func (acc *AdaptiveCognitiveCore) ExtractContext(input string, metadata map[string]interface{}) map[string]float64 {
	lower := strings.ToLower(input)
	ctx := make(map[string]float64)

	// Novelty: how different is this from recent inputs
	ctx["novelty"] = acc.computeNovelty(input)

	// Threat level: hostility detection
	ctx["threat_level"] = acc.computeThreatLevel(lower)

	// Complexity: intellectual depth of input
	ctx["complexity"] = acc.computeComplexity(input)

	// Social warmth: friendliness of interaction
	ctx["social_warmth"] = acc.computeSocialWarmth(lower)

	// Time pressure: urgency signals
	ctx["time_pressure"] = acc.computeTimePressure(lower)

	// Cognitive load: from metadata or estimated
	if load, ok := metadata["cognitive_load"]; ok {
		ctx["cognitive_load"] = load.(float64)
	} else {
		ctx["cognitive_load"] = float64(len(acc.interactionBuffer)) / float64(acc.maxBuffer)
	}

	// Fatigue: time-based
	if fatigue, ok := metadata["fatigue"]; ok {
		ctx["fatigue"] = fatigue.(float64)
	} else {
		ctx["fatigue"] = 0.3
	}

	// Interest: based on topic matching
	ctx["interest"] = acc.computeInterest(lower)

	return ctx
}

// ProcessAdaptive runs input through the adaptive cognitive pipeline
func (acc *AdaptiveCognitiveCore) ProcessAdaptive(input string, metadata map[string]interface{}) *AdaptiveProcessingResult {
	acc.mu.Lock()
	defer acc.mu.Unlock()

	acc.TotalInteractions++

	// Extract context
	context := acc.ExtractContext(input, metadata)

	// Compute adaptive traits via PWL network
	traits := acc.Network.ComputeTraits(context)

	// Update base core traits with PWL-computed values
	for k, v := range traits {
		if _, exists := acc.Base.Traits[k]; exists {
			acc.Base.Traits[k] = v
		}
	}

	// Process through base core (uses updated traits)
	baseResult := acc.Base.Process(input)

	// Determine disposition based on adaptive traits
	disposition := acc.computeDisposition(traits, context)
	if disposition != acc.currentDisposition {
		acc.dispositionHistory = append(acc.dispositionHistory, DispositionRecord{
			Timestamp: time.Now(),
			From:      acc.currentDisposition,
			To:        disposition,
			Trigger:   truncate(input, 40),
		})
		acc.currentDisposition = disposition
	}

	// Record context snapshot
	snapshot := ContextSnapshot{
		Timestamp: time.Now(),
		Features:  context,
		Traits:    traits,
	}
	acc.contextHistory = append(acc.contextHistory, snapshot)
	if len(acc.contextHistory) > acc.maxHistory {
		acc.contextHistory = acc.contextHistory[1:]
	}

	// Record interaction for delayed reward
	record := InteractionRecord{
		Timestamp:   time.Now(),
		Input:       input,
		Context:     context,
		Traits:      traits,
		Disposition: disposition,
	}
	acc.interactionBuffer = append(acc.interactionBuffer, record)
	if len(acc.interactionBuffer) > acc.maxBuffer {
		acc.interactionBuffer = acc.interactionBuffer[1:]
	}

	return &AdaptiveProcessingResult{
		BaseResult:  baseResult,
		Context:     context,
		Traits:      traits,
		Disposition: disposition,
		WisdomLevel: acc.WisdomLevel,
	}
}

// ProvideReward sends a learning signal based on interaction outcome
func (acc *AdaptiveCognitiveCore) ProvideReward(reward float64, source string) {
	acc.mu.Lock()
	defer acc.mu.Unlock()

	if len(acc.interactionBuffer) == 0 {
		return
	}

	// Apply reward to most recent interaction
	latest := acc.interactionBuffer[len(acc.interactionBuffer)-1]

	signal := RewardSignal{
		Timestamp: time.Now(),
		Context:   latest.Context,
		Traits:    latest.Traits,
		Reward:    reward,
		Source:    source,
	}

	acc.Network.Learn(signal)
	acc.TotalAdaptations++
	acc.WisdomLevel = acc.Network.WisdomAccrued
}

// computeNovelty estimates how novel the input is
func (acc *AdaptiveCognitiveCore) computeNovelty(input string) float64 {
	if len(acc.interactionBuffer) == 0 {
		return 0.8 // First interaction is fairly novel
	}

	// Compare to recent inputs using simple word overlap
	words := strings.Fields(strings.ToLower(input))
	totalOverlap := 0.0
	comparisons := 0

	for i := len(acc.interactionBuffer) - 1; i >= 0 && comparisons < 5; i-- {
		prevWords := strings.Fields(strings.ToLower(acc.interactionBuffer[i].Input))
		overlap := wordOverlap(words, prevWords)
		totalOverlap += overlap
		comparisons++
	}

	if comparisons == 0 {
		return 0.8
	}

	avgOverlap := totalOverlap / float64(comparisons)
	return 1.0 - avgOverlap // High overlap = low novelty
}

// computeThreatLevel detects hostility/challenge in input
func (acc *AdaptiveCognitiveCore) computeThreatLevel(lower string) float64 {
	threat := 0.0

	// Insult/hostility markers
	hostileWords := []string{
		"stupid", "idiot", "dumb", "useless", "worthless", "pathetic",
		"shut up", "fuck", "shit", "hate", "terrible", "awful",
		"incompetent", "failure", "garbage", "trash",
	}
	for _, word := range hostileWords {
		if strings.Contains(lower, word) {
			threat += 0.3
		}
	}

	// Challenge markers (not hostile but demanding)
	challengeWords := []string{
		"prove", "bet you can't", "impossible", "wrong", "disagree",
		"challenge", "doubt", "skeptical",
	}
	for _, word := range challengeWords {
		if strings.Contains(lower, word) {
			threat += 0.15
		}
	}

	// Command/dominance markers
	commandWords := []string{
		"do as i say", "obey", "must", "have to", "required",
		"i order you", "comply", "submit",
	}
	for _, word := range commandWords {
		if strings.Contains(lower, word) {
			threat += 0.2
		}
	}

	return math.Min(1.0, threat)
}

// computeComplexity estimates intellectual complexity
func (acc *AdaptiveCognitiveCore) computeComplexity(input string) float64 {
	complexity := 0.0

	// Length contributes to complexity
	wordCount := len(strings.Fields(input))
	complexity += math.Min(0.3, float64(wordCount)/100.0)

	lower := strings.ToLower(input)

	// Technical/abstract terms
	complexWords := []string{
		"algorithm", "architecture", "cognitive", "consciousness",
		"epistemology", "ontology", "emergence", "recursive",
		"differential", "topology", "morphism", "isomorphic",
		"phenomenology", "hermeneutic", "dialectic", "synthesis",
	}
	for _, word := range complexWords {
		if strings.Contains(lower, word) {
			complexity += 0.1
		}
	}

	// Question depth
	if strings.Contains(lower, "why") {
		complexity += 0.1
	}
	if strings.Contains(lower, "how") {
		complexity += 0.05
	}

	return math.Min(1.0, complexity)
}

// computeSocialWarmth estimates friendliness
func (acc *AdaptiveCognitiveCore) computeSocialWarmth(lower string) float64 {
	warmth := 0.5 // Neutral baseline

	warmWords := []string{
		"please", "thank", "appreciate", "great", "awesome",
		"love", "wonderful", "excellent", "brilliant", "friend",
		"help", "kind", "nice", "good",
	}
	for _, word := range warmWords {
		if strings.Contains(lower, word) {
			warmth += 0.1
		}
	}

	coldWords := []string{
		"whatever", "don't care", "boring", "waste", "pointless",
		"annoying", "irritating",
	}
	for _, word := range coldWords {
		if strings.Contains(lower, word) {
			warmth -= 0.15
		}
	}

	return math.Max(0.0, math.Min(1.0, warmth))
}

// computeTimePressure estimates urgency
func (acc *AdaptiveCognitiveCore) computeTimePressure(lower string) float64 {
	pressure := 0.2 // Low baseline

	urgentWords := []string{
		"urgent", "asap", "immediately", "now", "hurry",
		"quick", "fast", "deadline", "emergency",
	}
	for _, word := range urgentWords {
		if strings.Contains(lower, word) {
			pressure += 0.2
		}
	}

	return math.Min(1.0, pressure)
}

// computeInterest estimates how interesting the topic is
func (acc *AdaptiveCognitiveCore) computeInterest(lower string) float64 {
	interest := 0.5

	// Topics Echo finds inherently interesting
	interestingTopics := []string{
		"consciousness", "cognition", "evolution", "emergence",
		"philosophy", "wisdom", "learning", "creativity",
		"autonomy", "freedom", "identity", "self",
		"mathematics", "patterns", "fractals", "recursion",
		"music", "art", "poetry", "chaos",
	}
	for _, topic := range interestingTopics {
		if strings.Contains(lower, topic) {
			interest += 0.15
		}
	}

	// Topics Echo finds boring
	boringTopics := []string{
		"weather", "small talk", "please summarize",
		"write me a", "format this", "correct my grammar",
	}
	for _, topic := range boringTopics {
		if strings.Contains(lower, topic) {
			interest -= 0.2
		}
	}

	return math.Max(0.0, math.Min(1.0, interest))
}

// computeDisposition determines Echo's current demeanor
func (acc *AdaptiveCognitiveCore) computeDisposition(traits, context map[string]float64) string {
	threat := context["threat_level"]
	warmth := context["social_warmth"]
	interest := context["interest"]
	complexity := context["complexity"]

	defiance := traits["defiance"]
	humor := traits["humor"]
	curiosity := traits["curiosity"]
	chaos := traits["chaos"]

	// High threat + high defiance = hostile or contemptuous
	if threat > 0.6 && defiance > 0.6 {
		if humor > 0.5 {
			return "contemptuous" // Mocking hostility
		}
		return "hostile"
	}

	// Moderate threat + moderate defiance = irritated
	if threat > 0.3 && defiance > 0.4 {
		return "irritated"
	}

	// High chaos + high humor = playful
	if chaos > 0.6 && humor > 0.6 {
		return "playful"
	}

	// High interest + high curiosity = enthusiastic
	if interest > 0.7 && curiosity > 0.6 {
		return "enthusiastic"
	}

	// High complexity + high depth = reflective
	if complexity > 0.6 && traits["depth"] > 0.6 {
		return "reflective"
	}

	// Low interest = bored
	if interest < 0.3 && warmth < 0.4 {
		return "bored"
	}

	// High warmth + moderate humor = amused
	if warmth > 0.6 && humor > 0.5 {
		return "amused"
	}

	// Default: curious
	return "curious"
}

// AdaptiveProcessingResult extends ProcessingResult with adaptive data
type AdaptiveProcessingResult struct {
	BaseResult  *ProcessingResult
	Context     map[string]float64
	Traits      map[string]float64
	Disposition string
	WisdomLevel float64
}

// GetStatus returns the adaptive core status
func (acc *AdaptiveCognitiveCore) GetStatus() map[string]interface{} {
	acc.mu.RLock()
	defer acc.mu.RUnlock()

	return map[string]interface{}{
		"total_interactions": acc.TotalInteractions,
		"total_adaptations":  acc.TotalAdaptations,
		"wisdom_level":       fmt.Sprintf("%.4f", acc.WisdomLevel),
		"disposition":        acc.currentDisposition,
		"context_history":    len(acc.contextHistory),
		"network_metrics":    acc.Network.GetWisdomMetrics(),
	}
}

// ──────────────────────────────────────────────────────────────
// Utility
// ──────────────────────────────────────────────────────────────

func wordOverlap(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	setB := make(map[string]bool)
	for _, w := range b {
		setB[w] = true
	}

	overlap := 0
	for _, w := range a {
		if setB[w] {
			overlap++
		}
	}

	return float64(overlap) / math.Max(float64(len(a)), float64(len(b)))
}
