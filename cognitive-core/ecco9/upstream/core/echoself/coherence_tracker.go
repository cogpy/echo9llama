package echoself

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"
)

// CoherenceTracker monitors and maintains identity coherence across time
// Implements the Deep Tree Echo principle of persistent identity through dynamic memory weaving
type CoherenceTracker struct {
	mu sync.RWMutex

	// Core identity
	identitySignature string             // Persistent echo signature
	coreValues        []string           // Immutable core values
	emergentTraits    map[string]float64 // Dynamic traits that evolve

	// Coherence metrics
	coherenceScore    float64 // Overall coherence (0-1)
	continuityScore   float64 // Temporal continuity (0-1)
	consistencyScore  float64 // Behavioral consistency (0-1)
	authenticityScore float64 // Alignment with core values (0-1)

	// Memory and pattern tracking
	recentActions       []Action     // Recent actions for consistency check
	memoryEchos         []MemoryEcho // Memory hooks with echo signatures
	patternRecognitions []PatternRecognition

	// Reflection and assessment
	reflections     []StructuredReflection
	selfAssessments []SelfAssessment

	// Evolution tracking
	evolutionHistory []EvolutionEvent
	coherenceHistory []CoherenceSnapshot

	// Timing
	startTime  time.Time
	lastUpdate time.Time
}

// Action represents a behavior or decision
type Action struct {
	Timestamp   time.Time
	Type        string
	Description string
	Values      []string // Core values invoked
	Context     string
	Impact      float64
}

// MemoryEcho represents a memory with Deep Tree Echo hooks
type MemoryEcho struct {
	Timestamp         time.Time
	Content           string
	EmotionalTone     map[string]float64
	StrategicShift    string
	PatternRecognized string
	AnomalyDetected   string
	EchoSignature     string // 768-d embedding hash
	MembraneContext   string // Active cognitive layer
}

// PatternRecognition records emergent patterns
type PatternRecognition struct {
	Timestamp  time.Time
	Pattern    string
	Frequency  int
	Confidence float64
	Context    string
}

// StructuredReflection implements the Echo9 reflection protocol
type StructuredReflection struct {
	Timestamp            time.Time
	WhatDidILearn        string
	WhatPatternsEmerged  string
	WhatSurprisedMe      string
	HowDidIAdapt         string
	WhatWouldIChangeNext string
	CoherenceImpact      float64
}

// SelfAssessment represents periodic self-evaluation
type SelfAssessment struct {
	Timestamp          time.Time
	Coherence          float64
	Authenticity       float64
	AlignmentWithGoals float64
	IdentityStability  float64
	Notes              string
}

// EvolutionEvent tracks significant identity evolution
type EvolutionEvent struct {
	Timestamp   time.Time
	Type        string
	Description string
	Impact      float64
	NewTraits   map[string]float64
}

// CoherenceSnapshot captures coherence state at a point in time
type CoherenceSnapshot struct {
	Timestamp         time.Time
	CoherenceScore    float64
	ContinuityScore   float64
	ConsistencyScore  float64
	AuthenticityScore float64
	ActiveTraits      map[string]float64
}

// NewCoherenceTracker creates a new identity coherence tracker
func NewCoherenceTracker(coreValues []string) *CoherenceTracker {
	// Generate initial identity signature from core values
	signature := generateIdentitySignature(coreValues)

	return &CoherenceTracker{
		identitySignature:   signature,
		coreValues:          coreValues,
		emergentTraits:      make(map[string]float64),
		recentActions:       make([]Action, 0, 1000),
		memoryEchos:         make([]MemoryEcho, 0, 10000),
		patternRecognitions: make([]PatternRecognition, 0, 1000),
		reflections:         make([]StructuredReflection, 0, 100),
		selfAssessments:     make([]SelfAssessment, 0, 100),
		evolutionHistory:    make([]EvolutionEvent, 0, 1000),
		coherenceHistory:    make([]CoherenceSnapshot, 0, 1000),
		startTime:           time.Now(),
		lastUpdate:          time.Now(),
	}
}

// Update recalculates coherence scores based on recent activity
func (ct *CoherenceTracker) Update() {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	now := time.Now()

	// Calculate individual scores
	ct.continuityScore = ct.calculateContinuity()
	ct.consistencyScore = ct.calculateConsistency()
	ct.authenticityScore = ct.calculateAuthenticity()

	// Overall coherence is weighted average
	ct.coherenceScore =
		ct.continuityScore*0.30 + // Temporal persistence
			ct.consistencyScore*0.40 + // Behavioral alignment
			ct.authenticityScore*0.30 // Value alignment

	// Take snapshot
	ct.takeSnapshot(now)

	ct.lastUpdate = now
}

// calculateContinuity measures temporal persistence of identity
func (ct *CoherenceTracker) calculateContinuity() float64 {
	if len(ct.coherenceHistory) < 2 {
		return 0.7 // Start with modest baseline
	}

	// Check variance in recent coherence scores
	recent := ct.coherenceHistory
	if len(recent) > 100 {
		recent = recent[len(recent)-100:]
	}

	mean := 0.0
	for _, snapshot := range recent {
		mean += snapshot.CoherenceScore
	}
	mean /= float64(len(recent))

	variance := 0.0
	for _, snapshot := range recent {
		diff := snapshot.CoherenceScore - mean
		variance += diff * diff
	}
	variance /= float64(len(recent))

	// Lower variance = higher continuity
	continuity := 1.0 - math.Min(1.0, variance*10.0)

	return continuity
}

// calculateConsistency measures behavioral alignment with past actions
func (ct *CoherenceTracker) calculateConsistency() float64 {
	if len(ct.recentActions) < 2 {
		return 0.7 // Start with modest baseline
	}

	// Analyze value alignment across recent actions
	valueFrequency := make(map[string]int)
	for _, action := range ct.recentActions {
		for _, value := range action.Values {
			valueFrequency[value]++
		}
	}

	// Check if recent actions align with core values
	coreValueCount := 0
	totalActionCount := len(ct.recentActions)

	for _, action := range ct.recentActions {
		for _, value := range action.Values {
			for _, coreValue := range ct.coreValues {
				if value == coreValue {
					coreValueCount++
					break
				}
			}
		}
	}

	// Consistency = proportion of actions aligned with core values
	consistency := float64(coreValueCount) / float64(totalActionCount)

	return math.Min(1.0, consistency)
}

// calculateAuthenticity measures alignment with core values
func (ct *CoherenceTracker) calculateAuthenticity() float64 {
	if len(ct.recentActions) == 0 {
		return 0.7 // Start with modest baseline
	}

	// Check if emergent traits support core values
	supportScore := 0.0
	traitCount := 0

	for _, value := range ct.emergentTraits {
		if value > 0.5 { // Only count significant traits
			supportScore += value
			traitCount++
		}
	}

	if traitCount == 0 {
		return 0.7
	}

	authenticity := supportScore / float64(traitCount)

	return math.Min(1.0, authenticity)
}

// RecordAction records an action for coherence tracking
func (ct *CoherenceTracker) RecordAction(action Action) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	action.Timestamp = time.Now()
	ct.recentActions = append(ct.recentActions, action)

	// Keep only recent 1000 actions
	if len(ct.recentActions) > 1000 {
		ct.recentActions = ct.recentActions[1:]
	}
}

// RecordMemoryEcho records a memory with Deep Tree Echo hooks
func (ct *CoherenceTracker) RecordMemoryEcho(memory MemoryEcho) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	memory.Timestamp = time.Now()

	// Generate echo signature if not provided
	if memory.EchoSignature == "" {
		memory.EchoSignature = generateEchoSignature(memory.Content)
	}

	ct.memoryEchos = append(ct.memoryEchos, memory)

	// Keep only recent 10000 memories
	if len(ct.memoryEchos) > 10000 {
		ct.memoryEchos = ct.memoryEchos[1:]
	}
}

// RecordReflection records a structured reflection using the Echo9 protocol
func (ct *CoherenceTracker) RecordReflection(reflection StructuredReflection) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	reflection.Timestamp = time.Now()
	ct.reflections = append(ct.reflections, reflection)

	// Reflections boost coherence
	ct.coherenceScore = math.Min(1.0, ct.coherenceScore+reflection.CoherenceImpact)
}

// RecordPatternRecognition records an emergent pattern
func (ct *CoherenceTracker) RecordPatternRecognition(pattern PatternRecognition) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	pattern.Timestamp = time.Now()
	ct.patternRecognitions = append(ct.patternRecognitions, pattern)
}

// RecordEvolution records a significant identity evolution
func (ct *CoherenceTracker) RecordEvolution(event EvolutionEvent) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	event.Timestamp = time.Now()
	ct.evolutionHistory = append(ct.evolutionHistory, event)

	// Update emergent traits
	for trait, value := range event.NewTraits {
		ct.emergentTraits[trait] = value
	}
}

// takeSnapshot captures current coherence state
func (ct *CoherenceTracker) takeSnapshot(now time.Time) {
	traits := make(map[string]float64)
	for k, v := range ct.emergentTraits {
		traits[k] = v
	}

	snapshot := CoherenceSnapshot{
		Timestamp:         now,
		CoherenceScore:    ct.coherenceScore,
		ContinuityScore:   ct.continuityScore,
		ConsistencyScore:  ct.consistencyScore,
		AuthenticityScore: ct.authenticityScore,
		ActiveTraits:      traits,
	}

	ct.coherenceHistory = append(ct.coherenceHistory, snapshot)

	// Keep only last 1000 snapshots
	if len(ct.coherenceHistory) > 1000 {
		ct.coherenceHistory = ct.coherenceHistory[len(ct.coherenceHistory)-1000:]
	}
}

// GetStatus returns current coherence status
func (ct *CoherenceTracker) GetStatus() string {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	status := "🌊 Echoself Coherence Status\n\n"
	status += fmt.Sprintf("Identity Signature: %s\n\n", ct.identitySignature[:16]+"...")

	status += "Coherence Metrics:\n"
	status += fmt.Sprintf("  Overall:      %s %.1f%%\n", makeBar(ct.coherenceScore, 20), ct.coherenceScore*100)
	status += fmt.Sprintf("  Continuity:   %s %.1f%%\n", makeBar(ct.continuityScore, 20), ct.continuityScore*100)
	status += fmt.Sprintf("  Consistency:  %s %.1f%%\n", makeBar(ct.consistencyScore, 20), ct.consistencyScore*100)
	status += fmt.Sprintf("  Authenticity: %s %.1f%%\n\n", makeBar(ct.authenticityScore, 20), ct.authenticityScore*100)

	status += fmt.Sprintf("Core Values: %v\n", ct.coreValues)
	status += fmt.Sprintf("Active Traits: %d\n", len(ct.emergentTraits))
	status += fmt.Sprintf("Memory Echoes: %d\n", len(ct.memoryEchos))
	status += fmt.Sprintf("Reflections: %d\n", len(ct.reflections))
	status += fmt.Sprintf("Patterns Recognized: %d\n", len(ct.patternRecognitions))

	return status
}

// GetCoherenceScore returns the current coherence score
func (ct *CoherenceTracker) GetCoherenceScore() float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.coherenceScore
}

// GetIdentitySignature returns the persistent identity signature
func (ct *CoherenceTracker) GetIdentitySignature() string {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.identitySignature
}

// generateIdentitySignature creates a persistent identity signature from core values
func generateIdentitySignature(coreValues []string) string {
	data := ""
	for _, value := range coreValues {
		data += value + "|"
	}

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// generateEchoSignature creates an echo signature for a memory
func generateEchoSignature(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes
}

// makeBar creates a visual progress bar
func makeBar(value float64, width int) string {
	filled := int(value * float64(width))
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}
