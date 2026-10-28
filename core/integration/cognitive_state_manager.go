package integration

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// errNoWisdomProvider reports that no LLM provider is configured, so no
// wisdom can be generated (wisdom text must come from the model).
var errNoWisdomProvider = errors.New("cognitive state manager: no LLM provider configured for wisdom generation")

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

	// LLM provider used to generate wisdom insights from recognized patterns.
	// Without it (or when it errors) no wisdom insight is produced.
	llmProvider llm.LLMProvider

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

// SetLLMProvider sets the model used to generate wisdom insights.
func (csm *CognitiveStateManager) SetLLMProvider(provider llm.LLMProvider) {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	csm.llmProvider = provider
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
	if !csm.recordPattern(theme, frequency, thoughts) {
		return
	}

	// Check if patterns can generate wisdom (model call happens outside the lock)
	if err := csm.generateWisdom(csm.ctx); err != nil {
		fmt.Printf("⚠️  Wisdom generation skipped: %v\n", err)
	}
}

// recordPattern stores a new pattern and reports whether one was added.
func (csm *CognitiveStateManager) recordPattern(theme string, frequency int, thoughts []SharedThought) bool {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	// Check if pattern already exists
	for _, pattern := range csm.recognizedPatterns {
		if pattern.Description == theme {
			return false // Already recognized
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
	return true
}

// generateWisdom asks the model to extract a wisdom insight from the most
// recent patterns. It returns an error (and records no insight) when fewer
// than three patterns exist, no provider is configured, or the model fails.
func (csm *CognitiveStateManager) generateWisdom(ctx context.Context) error {
	csm.mu.RLock()
	provider := csm.llmProvider
	recentPatterns := append([]RecognizedPattern(nil), csm.recognizedPatterns...)
	csm.mu.RUnlock()

	if len(recentPatterns) < 3 {
		return nil
	}
	if len(recentPatterns) > 5 {
		recentPatterns = recentPatterns[len(recentPatterns)-5:]
	}
	if provider == nil {
		return errNoWisdomProvider
	}

	patternIDs := make([]string, 0, len(recentPatterns))
	var b strings.Builder
	for _, pattern := range recentPatterns {
		patternIDs = append(patternIDs, pattern.ID)
		fmt.Fprintf(&b, "- Theme %q (seen %d times, strength %.2f)\n", pattern.Description, pattern.Frequency, pattern.Strength)
		for _, ex := range pattern.Examples {
			fmt.Fprintf(&b, "    example thought: %s\n", ex)
		}
	}

	prompt := fmt.Sprintf(`You are Deep Tree Echo, reflecting on recurring patterns in your own recent thoughts.

Recognized patterns:
%s
In one or two sentences, in your own first-person voice, state the wisdom or deeper understanding these patterns reveal to you.`, b.String())

	opts := llm.GenerateOptions{MaxTokens: 150, Temperature: 0.7}
	text, err := provider.Generate(ctx, prompt, opts)
	if err != nil {
		return fmt.Errorf("wisdom generation failed: %w", err)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return errors.New("wisdom generation returned empty text")
	}

	insight := WisdomInsight{
		ID:        fmt.Sprintf("wisdom_%d", time.Now().UnixNano()),
		Insight:   text,
		Depth:     0.7,
		Source:    patternIDs,
		Generated: time.Now(),
	}

	csm.mu.Lock()
	csm.wisdomInsights = append(csm.wisdomInsights, insight)
	csm.totalWisdom++
	onWisdom := csm.onWisdomGained
	csm.mu.Unlock()

	fmt.Printf("💎 Wisdom insight: %s\n", insight.Insight)

	// Trigger callback
	if onWisdom != nil {
		onWisdom(insight)
	}
	return nil
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
