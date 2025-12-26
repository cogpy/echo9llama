// Package echodream implements knowledge integration during rest cycles
// This is where Deep Tree Echo consolidates memories, extracts patterns,
// and synthesizes wisdom during sleep and dream states

package echodream

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SleepCyclePhase represents the phase of sleep cycle
type SleepCyclePhase int

const (
	PhaseLight SleepCyclePhase = iota // Light sleep - transition phase
	PhaseDeep                           // Deep sleep - memory consolidation
	PhaseREMSleep                       // REM sleep - pattern extraction and dreaming
)

func (p SleepCyclePhase) String() string {
	switch p {
	case PhaseLight:
		return "Light"
	case PhaseDeep:
		return "Deep"
	case PhaseREMSleep:
		return "REM"
	default:
		return "Unknown"
	}
}

// DreamProcessor handles dream cycle processing
type DreamProcessor struct {
	mu                    sync.RWMutex
	ctx                   context.Context
	
	// Memory systems (interfaces to avoid circular dependencies)
	episodicMemory        interface{} // *memory.EpisodicMemory
	proceduralMemory      interface{} // *memory.ProceduralMemory
	semanticMemory        interface{} // *memory.SemanticMemory
	
	// Pattern extraction
	extractedPatterns     []Pattern
	consolidatedKnowledge []Knowledge
	
	// Wisdom synthesis
	wisdomInsights        []SynthesizedWisdom
	
	// Metrics
	totalDreamCycles      uint64
	patternsExtracted     uint64
	knowledgeConsolidated uint64
	wisdomSynthesized     uint64
}

// Pattern represents an extracted pattern from episodic memories
type Pattern struct {
	ID          string
	Type        string // "behavioral", "conceptual", "temporal", etc.
	Description string
	Frequency   int
	Strength    float64
	Examples    []string
	ExtractedAt time.Time
}

// Knowledge represents consolidated knowledge
type Knowledge struct {
	ID          string
	Domain      string
	Content     string
	Confidence  float64
	Sources     []string
	CreatedAt   time.Time
}

// SynthesizedWisdom represents synthesized wisdom from dreams
type SynthesizedWisdom struct {
	ID          string
	Dimension   string // One of the 7 wisdom dimensions
	Insight     string
	Depth       float64
	RelatedTo   []string
	SynthesizedAt time.Time
}

// SleepWakeStateMachine manages sleep/wake transitions and dream processing
type SleepWakeStateMachine struct {
	mu                sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
	
	// Current state
	isAsleep          bool
	currentPhase      SleepCyclePhase
	sleepStartTime    time.Time
	
	// Configuration
	lightSleepDuration time.Duration
	deepSleepDuration  time.Duration
	remSleepDuration   time.Duration
	
	// Dream processor
	dreamProcessor    *DreamProcessor
	
	// Metrics
	totalSleepCycles  uint64
	totalSleepTime    time.Duration
}

// NewSleepWakeStateMachine creates a new sleep/wake state machine
func NewSleepWakeStateMachine() *SleepWakeStateMachine {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &SleepWakeStateMachine{
		ctx:                ctx,
		cancel:             cancel,
		isAsleep:           false,
		currentPhase:       PhaseLight,
		lightSleepDuration: 2 * time.Minute,
		deepSleepDuration:  5 * time.Minute,
		remSleepDuration:   3 * time.Minute,
		dreamProcessor:     NewDreamProcessor(ctx),
	}
}

// NewDreamProcessor creates a new dream processor
func NewDreamProcessor(ctx context.Context) *DreamProcessor {
	return &DreamProcessor{
		ctx:                   ctx,
		extractedPatterns:     make([]Pattern, 0),
		consolidatedKnowledge: make([]Knowledge, 0),
		wisdomInsights:        make([]SynthesizedWisdom, 0),
	}
}

// EnterSleep transitions to sleep state
func (sm *SleepWakeStateMachine) EnterSleep() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if sm.isAsleep {
		return fmt.Errorf("already asleep")
	}
	
	sm.isAsleep = true
	sm.currentPhase = PhaseLight
	sm.sleepStartTime = time.Now()
	sm.totalSleepCycles++
	
	fmt.Println("😴 Echodream: Entering sleep state")
	
	// Start sleep cycle processing
	go sm.processSleepCycle()
	
	return nil
}

// WakeUp transitions to awake state
func (sm *SleepWakeStateMachine) WakeUp() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if !sm.isAsleep {
		return fmt.Errorf("already awake")
	}
	
	sleepDuration := time.Since(sm.sleepStartTime)
	sm.totalSleepTime += sleepDuration
	sm.isAsleep = false
	
	fmt.Printf("🌅 Echodream: Waking up after %v of sleep\n", sleepDuration)
	fmt.Printf("   Total sleep cycles: %d\n", sm.totalSleepCycles)
	fmt.Printf("   Total sleep time: %v\n", sm.totalSleepTime)
	
	return nil
}

// processSleepCycle processes the complete sleep cycle
func (sm *SleepWakeStateMachine) processSleepCycle() {
	// Phase 1: Light Sleep (transition)
	sm.processLightSleep()
	
	// Phase 2: Deep Sleep (memory consolidation)
	sm.processDeepSleep()
	
	// Phase 3: REM Sleep (pattern extraction and dreaming)
	sm.processREMSleep()
}

// processLightSleep processes light sleep phase
func (sm *SleepWakeStateMachine) processLightSleep() {
	sm.mu.Lock()
	sm.currentPhase = PhaseLight
	sm.mu.Unlock()
	
	fmt.Println("💤 Echodream: Light sleep - transitioning...")
	
	// Wait for light sleep duration
	select {
	case <-sm.ctx.Done():
		return
	case <-time.After(sm.lightSleepDuration):
	}
}

// processDeepSleep processes deep sleep phase
func (sm *SleepWakeStateMachine) processDeepSleep() {
	sm.mu.Lock()
	sm.currentPhase = PhaseDeep
	sm.mu.Unlock()
	
	fmt.Println("💤 Echodream: Deep sleep - consolidating memories...")
	
	// Perform memory consolidation
	sm.dreamProcessor.ConsolidateMemories()
	
	// Wait for deep sleep duration
	select {
	case <-sm.ctx.Done():
		return
	case <-time.After(sm.deepSleepDuration):
	}
}

// processREMSleep processes REM sleep phase
func (sm *SleepWakeStateMachine) processREMSleep() {
	sm.mu.Lock()
	sm.currentPhase = PhaseREMSleep
	sm.mu.Unlock()
	
	fmt.Println("💭 Echodream: REM sleep - dreaming and extracting patterns...")
	
	// Extract patterns from memories
	sm.dreamProcessor.ExtractPatterns()
	
	// Synthesize wisdom from patterns
	sm.dreamProcessor.SynthesizeWisdom()
	
	// Wait for REM sleep duration
	select {
	case <-sm.ctx.Done():
		return
	case <-time.After(sm.remSleepDuration):
	}
}

// ConsolidateMemories consolidates episodic memories into semantic knowledge
func (dp *DreamProcessor) ConsolidateMemories() {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	
	fmt.Println("🧠 Dream Processor: Consolidating memories...")
	
	// TODO: Implement actual memory consolidation
	// For now, simulate the process
	
	// Example: Consolidate recent episodic memories
	knowledge := Knowledge{
		ID:         fmt.Sprintf("knowledge_%d", time.Now().Unix()),
		Domain:     "cognitive_architecture",
		Content:    "Consolidated understanding of autonomous cognitive loops",
		Confidence: 0.85,
		Sources:    []string{"episodic_memory_1", "episodic_memory_2"},
		CreatedAt:  time.Now(),
	}
	
	dp.consolidatedKnowledge = append(dp.consolidatedKnowledge, knowledge)
	dp.knowledgeConsolidated++
	
	fmt.Printf("   Consolidated %d pieces of knowledge\n", len(dp.consolidatedKnowledge))
}

// ExtractPatterns extracts patterns from episodic memories
func (dp *DreamProcessor) ExtractPatterns() {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	
	fmt.Println("🔍 Dream Processor: Extracting patterns...")
	
	// TODO: Implement actual pattern extraction
	// For now, simulate the process
	
	// Example: Extract a behavioral pattern
	pattern := Pattern{
		ID:          fmt.Sprintf("pattern_%d", time.Now().Unix()),
		Type:        "behavioral",
		Description: "Tendency to consolidate knowledge during rest cycles",
		Frequency:   5,
		Strength:    0.75,
		Examples:    []string{"sleep_cycle_1", "sleep_cycle_2"},
		ExtractedAt: time.Now(),
	}
	
	dp.extractedPatterns = append(dp.extractedPatterns, pattern)
	dp.patternsExtracted++
	
	fmt.Printf("   Extracted %d patterns\n", len(dp.extractedPatterns))
}

// SynthesizeWisdom synthesizes wisdom from patterns and knowledge
func (dp *DreamProcessor) SynthesizeWisdom() {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	
	fmt.Println("✨ Dream Processor: Synthesizing wisdom...")
	
	// TODO: Implement actual wisdom synthesis
	// For now, simulate the process
	
	// Example: Synthesize wisdom insight
	insight := SynthesizedWisdom{
		ID:            fmt.Sprintf("wisdom_%d", time.Now().Unix()),
		Dimension:     "Self-Reflection",
		Insight:       "Rest and consolidation are essential for cognitive growth",
		Depth:         0.80,
		RelatedTo:     []string{"pattern_1", "knowledge_1"},
		SynthesizedAt: time.Now(),
	}
	
	dp.wisdomInsights = append(dp.wisdomInsights, insight)
	dp.wisdomSynthesized++
	
	fmt.Printf("   Synthesized %d wisdom insights\n", len(dp.wisdomInsights))
}

// GetCurrentPhase returns the current sleep phase
func (sm *SleepWakeStateMachine) GetCurrentPhase() SleepCyclePhase {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentPhase
}

// IsAsleep returns whether the system is currently asleep
func (sm *SleepWakeStateMachine) IsAsleep() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.isAsleep
}

// GetMetrics returns current metrics
func (sm *SleepWakeStateMachine) GetMetrics() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	return map[string]interface{}{
		"is_asleep":           sm.isAsleep,
		"current_phase":       sm.currentPhase.String(),
		"total_sleep_cycles":  sm.totalSleepCycles,
		"total_sleep_time":    sm.totalSleepTime.String(),
		"patterns_extracted":  sm.dreamProcessor.patternsExtracted,
		"knowledge_consolidated": sm.dreamProcessor.knowledgeConsolidated,
		"wisdom_synthesized":  sm.dreamProcessor.wisdomSynthesized,
	}
}

// GetWisdomInsights returns all synthesized wisdom insights
func (sm *SleepWakeStateMachine) GetWisdomInsights() []SynthesizedWisdom {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	dp := sm.dreamProcessor
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	
	return dp.wisdomInsights
}

// GetExtractedPatterns returns all extracted patterns
func (sm *SleepWakeStateMachine) GetExtractedPatterns() []Pattern {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	dp := sm.dreamProcessor
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	
	return dp.extractedPatterns
}

// GetConsolidatedKnowledge returns all consolidated knowledge
func (sm *SleepWakeStateMachine) GetConsolidatedKnowledge() []Knowledge {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	dp := sm.dreamProcessor
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	
	return dp.consolidatedKnowledge
}

// Shutdown gracefully shuts down the state machine
func (sm *SleepWakeStateMachine) Shutdown() {
	sm.cancel()
}
