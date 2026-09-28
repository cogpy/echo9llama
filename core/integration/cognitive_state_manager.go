package integration

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CognitiveStateManager integrates echobeats and echodream with shared state
type CognitiveStateManager struct {
	mu     sync.RWMutex
	ctx    context.Context //nolint:containedctx // lifecycle context created with cancel in the constructor; stops background goroutines
	cancel context.CancelFunc

	// Shared cognitive state
	currentPhase  string
	currentFocus  string
	cognitiveLoad float64
	fatigueLevel  float64

	// Thought flow
	thoughtBuffer []SharedThought
	maxThoughts   int

	// Pattern and wisdom accumulation
	recognizedPatterns []RecognizedPattern
	wisdomInsights     []WisdomInsight

	// Integration callbacks
	onPhaseChange       func(string)
	onThoughtGenerated  func(SharedThought)
	onPatternRecognized func(RecognizedPattern)
	onWisdomGained      func(WisdomInsight)

	// Metrics
	totalThoughts uint64
	totalPatterns uint64
	totalWisdom   uint64

	// Running state
	running bool
}

// SharedThought represents a thought shared between subsystems
type SharedThought struct {
	ID         string
	Content    string
	Phase      string
	Source     string // "echobeats", "autonomous_engine", etc.
	Timestamp  time.Time
	Importance float64
	Tags       []string
}

// RecognizedPattern represents a pattern identified across thoughts
type RecognizedPattern struct {
	ID          string
	Description string
	Frequency   int
	Strength    float64
	Examples    []string
	Identified  time.Time
}

// WisdomInsight represents wisdom extracted from patterns
type WisdomInsight struct {
	ID        string
	Insight   string
	Depth     float64
	Source    []string // Pattern IDs
	Generated time.Time
}

// NewCognitiveStateManager creates a new cognitive state manager
func NewCognitiveStateManager() *CognitiveStateManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &CognitiveStateManager{
		ctx:                ctx,
		cancel:             cancel,
		thoughtBuffer:      make([]SharedThought, 0),
		maxThoughts:        100,
		recognizedPatterns: make([]RecognizedPattern, 0),
		wisdomInsights:     make([]WisdomInsight, 0),
	}
}

// Start begins cognitive state management
func (csm *CognitiveStateManager) Start() error {
	csm.mu.Lock()
	if csm.running {
		csm.mu.Unlock()
		return fmt.Errorf("already running")
	}
	csm.running = true
	csm.mu.Unlock()

	fmt.Println("🧠 Starting Cognitive State Manager...")
	fmt.Println("   Integrating echobeats, echodream, and thought generation")

	go csm.run()

	return nil
}

// Stop gracefully stops the manager
func (csm *CognitiveStateManager) Stop() error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if !csm.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("🧠 Stopping cognitive state manager...")
	csm.running = false
	csm.cancel()

	return nil
}

// run executes the main integration loop
func (csm *CognitiveStateManager) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-csm.ctx.Done():
			return
		case <-ticker.C:
			csm.processThoughts()
		}
	}
}

// processThoughts analyzes recent thoughts for patterns
func (csm *CognitiveStateManager) processThoughts() {
	csm.mu.RLock()
	thoughts := csm.thoughtBuffer
	csm.mu.RUnlock()

	if len(thoughts) < 5 {
		return // Need at least 5 thoughts to find patterns
	}

	// Simple pattern recognition: find repeated themes
	themes := make(map[string]int)
	for _, thought := range thoughts {
		for _, tag := range thought.Tags {
			themes[tag]++
		}
	}

	// Identify patterns from frequent themes
	for theme, count := range themes {
		if count >= 3 {
			csm.recognizePattern(theme, count, thoughts)
		}
	}
}

// recognizePattern creates a recognized pattern
func (csm *CognitiveStateManager) recognizePattern(theme string, frequency int, thoughts []SharedThought) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	// Check if pattern already exists
	for _, pattern := range csm.recognizedPatterns {
		if pattern.Description == theme {
			return // Already recognized
		}
	}

	// Collect examples
	examples := make([]string, 0)
	for _, thought := range thoughts {
		for _, tag := range thought.Tags {
			if tag == theme && len(examples) < 3 {
				examples = append(examples, thought.Content)
			}
		}
	}

	pattern := RecognizedPattern{
		ID:          fmt.Sprintf("pattern_%d", time.Now().UnixNano()),
		Description: theme,
		Frequency:   frequency,
		Strength:    float64(frequency) / float64(len(thoughts)),
		Examples:    examples,
		Identified:  time.Now(),
	}

	csm.recognizedPatterns = append(csm.recognizedPatterns, pattern)
	csm.totalPatterns++

	fmt.Printf("🔍 Pattern recognized: %s (frequency: %d, strength: %.2f)\n",
		theme, frequency, pattern.Strength)

	// Trigger callback
	if csm.onPatternRecognized != nil {
		csm.onPatternRecognized(pattern)
	}

	// Check if patterns can generate wisdom
	if len(csm.recognizedPatterns) >= 3 {
		csm.generateWisdom()
	}
}

// generateWisdom extracts wisdom from multiple patterns
func (csm *CognitiveStateManager) generateWisdom() {
	// Take recent patterns
	recentPatterns := csm.recognizedPatterns
	if len(recentPatterns) > 5 {
		recentPatterns = recentPatterns[len(recentPatterns)-5:]
	}

	// Generate wisdom insight (simplified - in production would use LLM)
	patternNames := make([]string, 0)
	patternIDs := make([]string, 0)
	for _, pattern := range recentPatterns {
		patternNames = append(patternNames, pattern.Description)
		patternIDs = append(patternIDs, pattern.ID)
	}

	insight := WisdomInsight{
		ID:        fmt.Sprintf("wisdom_%d", time.Now().UnixNano()),
		Insight:   fmt.Sprintf("Patterns of %v suggest deeper understanding emerging", patternNames),
		Depth:     0.7,
		Source:    patternIDs,
		Generated: time.Now(),
	}

	csm.wisdomInsights = append(csm.wisdomInsights, insight)
	csm.totalWisdom++

	fmt.Printf("💎 Wisdom insight: %s\n", insight.Insight)

	// Trigger callback
	if csm.onWisdomGained != nil {
		csm.onWisdomGained(insight)
	}
}

// AddThought adds a thought to the shared buffer
func (csm *CognitiveStateManager) AddThought(content, phase, source string, importance float64, tags []string) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	thought := SharedThought{
		ID:         fmt.Sprintf("thought_%d", time.Now().UnixNano()),
		Content:    content,
		Phase:      phase,
		Source:     source,
		Timestamp:  time.Now(),
		Importance: importance,
		Tags:       tags,
	}

	csm.thoughtBuffer = append(csm.thoughtBuffer, thought)
	csm.totalThoughts++

	// Trim buffer if too large
	if len(csm.thoughtBuffer) > csm.maxThoughts {
		csm.thoughtBuffer = csm.thoughtBuffer[1:]
	}

	// Trigger callback
	if csm.onThoughtGenerated != nil {
		csm.onThoughtGenerated(thought)
	}
}

// UpdatePhase updates the current cognitive phase
func (csm *CognitiveStateManager) UpdatePhase(phase string) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if csm.currentPhase != phase {
		csm.currentPhase = phase

		fmt.Printf("🔄 Phase transition: %s\n", phase)

		// Trigger callback
		if csm.onPhaseChange != nil {
			csm.onPhaseChange(phase)
		}
	}
}

// UpdateFocus updates the current cognitive focus
func (csm *CognitiveStateManager) UpdateFocus(focus string) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	csm.currentFocus = focus
}

// UpdateCognitiveLoad updates the current cognitive load
func (csm *CognitiveStateManager) UpdateCognitiveLoad(load float64) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	csm.cognitiveLoad = load
}

// UpdateFatigue updates the current fatigue level
func (csm *CognitiveStateManager) UpdateFatigue(fatigue float64) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	csm.fatigueLevel = fatigue
}

// GetCurrentPhase returns the current phase
func (csm *CognitiveStateManager) GetCurrentPhase() string {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	return csm.currentPhase
}

// GetCurrentFocus returns the current focus
func (csm *CognitiveStateManager) GetCurrentFocus() string {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	return csm.currentFocus
}

// GetRecentThoughts returns recent thoughts
func (csm *CognitiveStateManager) GetRecentThoughts(count int) []SharedThought {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	if count > len(csm.thoughtBuffer) {
		count = len(csm.thoughtBuffer)
	}

	start := len(csm.thoughtBuffer) - count
	return csm.thoughtBuffer[start:]
}

// GetRecognizedPatterns returns all recognized patterns
func (csm *CognitiveStateManager) GetRecognizedPatterns() []RecognizedPattern {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	return csm.recognizedPatterns
}

// GetWisdomInsights returns all wisdom insights
func (csm *CognitiveStateManager) GetWisdomInsights() []WisdomInsight {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	return csm.wisdomInsights
}

// SetCallbacks sets integration callbacks
func (csm *CognitiveStateManager) SetCallbacks(
	onPhaseChange func(string),
	onThoughtGenerated func(SharedThought),
	onPatternRecognized func(RecognizedPattern),
	onWisdomGained func(WisdomInsight),
) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	csm.onPhaseChange = onPhaseChange
	csm.onThoughtGenerated = onThoughtGenerated
	csm.onPatternRecognized = onPatternRecognized
	csm.onWisdomGained = onWisdomGained
}

// GetMetrics returns cognitive state metrics
func (csm *CognitiveStateManager) GetMetrics() map[string]interface{} {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	return map[string]interface{}{
		"current_phase":       csm.currentPhase,
		"current_focus":       csm.currentFocus,
		"cognitive_load":      csm.cognitiveLoad,
		"fatigue_level":       csm.fatigueLevel,
		"total_thoughts":      csm.totalThoughts,
		"total_patterns":      csm.totalPatterns,
		"total_wisdom":        csm.totalWisdom,
		"buffered_thoughts":   len(csm.thoughtBuffer),
		"recognized_patterns": len(csm.recognizedPatterns),
		"wisdom_insights":     len(csm.wisdomInsights),
	}
}

// ExportState exports the current state for persistence
func (csm *CognitiveStateManager) ExportState() map[string]interface{} {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	return map[string]interface{}{
		"current_phase":       csm.currentPhase,
		"current_focus":       csm.currentFocus,
		"cognitive_load":      csm.cognitiveLoad,
		"fatigue_level":       csm.fatigueLevel,
		"thought_buffer":      csm.thoughtBuffer,
		"recognized_patterns": csm.recognizedPatterns,
		"wisdom_insights":     csm.wisdomInsights,
	}
}

// ImportState imports state from persistence
func (csm *CognitiveStateManager) ImportState(state map[string]interface{}) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if phase, ok := state["current_phase"].(string); ok {
		csm.currentPhase = phase
	}
	if focus, ok := state["current_focus"].(string); ok {
		csm.currentFocus = focus
	}
	if load, ok := state["cognitive_load"].(float64); ok {
		csm.cognitiveLoad = load
	}
	if fatigue, ok := state["fatigue_level"].(float64); ok {
		csm.fatigueLevel = fatigue
	}

	// Import complex structures would need proper type assertions
	// Simplified for this implementation
}
