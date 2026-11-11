package echodream

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DreamCycleIntegration manages knowledge consolidation during rest/dream states
type DreamCycleIntegration struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Consolidation engine
	consolidator *KnowledgeConsolidator

	// Dream state
	isDreaming      bool
	currentDream    *Dream
	dreamHistory    []*Dream
	maxDreamHistory int

	// Memory buffers
	episodicBuffer []DreamEpisodicMemory
	workingMemory  []WorkingMemoryItem

	// Wisdom extraction
	wisdomExtractor *WisdomExtractor
	extractedWisdom []Wisdom

	// Integration metrics
	consolidationCycles uint64
	wisdomGenerated     uint64
	insightsIntegrated  uint64

	// Callbacks
	onWisdomExtracted func(wisdom Wisdom)
	onDreamComplete   func(dream *Dream)
}

// Dream represents a dream cycle
type Dream struct {
	ID                string                     `json:"id"`
	StartTime         time.Time                  `json:"start_time"`
	EndTime           time.Time                  `json:"end_time"`
	Duration          time.Duration              `json:"duration"`
	MemoriesProcessed int                        `json:"memories_processed"`
	WisdomExtracted   []string                   `json:"wisdom_extracted"`
	Insights          []string                   `json:"insights"`
	Consolidations    []DreamConsolidationResult `json:"consolidations"`
	EmotionalTone     map[string]float64         `json:"emotional_tone"`
	Themes            []string                   `json:"themes"`
	Narrative         string                     `json:"narrative"`
}

// DreamEpisodicMemory represents an experience to be consolidated during dreams
type DreamEpisodicMemory struct {
	ID         string                 `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	Content    string                 `json:"content"`
	Context    map[string]interface{} `json:"context"`
	Emotional  map[string]float64     `json:"emotional"`
	Importance float64                `json:"importance"`
	Tags       []string               `json:"tags"`
}

// WorkingMemoryItem represents active cognitive content
type WorkingMemoryItem struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context"`
}

// ConsolidationResult represents the result of memory consolidation
type DreamConsolidationResult struct {
	SourceMemories []string  `json:"source_memories"`
	ConsolidatedTo string    `json:"consolidated_to"`
	Type           string    `json:"type"` // "pattern", "concept", "skill", "wisdom"
	Strength       float64   `json:"strength"`
	Connections    []string  `json:"connections"`
	Timestamp      time.Time `json:"timestamp"`
}

// Wisdom represents extracted wisdom
type Wisdom struct {
	ID            string                 `json:"id"`
	Content       string                 `json:"content"`
	Type          string                 `json:"type"` // "principle", "heuristic", "insight", "understanding"
	Confidence    float64                `json:"confidence"`
	Sources       []string               `json:"sources"`
	Applicability []string               `json:"applicability"`
	Timestamp     time.Time              `json:"timestamp"`
	Context       map[string]interface{} `json:"context"`
}

// KnowledgeConsolidator consolidates memories into knowledge
type KnowledgeConsolidator struct {
	mu                sync.RWMutex
	consolidationRate float64
	patternThreshold  float64
	wisdomThreshold   float64
}

// WisdomExtractor extracts wisdom from consolidated knowledge
type WisdomExtractor struct {
	mu             sync.RWMutex
	extractionRate float64
	minConfidence  float64
}

// NewDreamCycleIntegration creates a new dream cycle integration system
func NewDreamCycleIntegration() *DreamCycleIntegration {
	ctx, cancel := context.WithCancel(context.Background())

	dci := &DreamCycleIntegration{
		ctx:             ctx,
		cancel:          cancel,
		consolidator:    NewKnowledgeConsolidator(),
		wisdomExtractor: NewWisdomExtractor(),
		dreamHistory:    make([]*Dream, 0),
		maxDreamHistory: 100,
		episodicBuffer:  make([]DreamEpisodicMemory, 0),
		workingMemory:   make([]WorkingMemoryItem, 0),
		extractedWisdom: make([]Wisdom, 0),
	}

	return dci
}

// NewKnowledgeConsolidator creates a new consolidator
func NewKnowledgeConsolidator() *KnowledgeConsolidator {
	return &KnowledgeConsolidator{
		consolidationRate: 0.7,
		patternThreshold:  0.6,
		wisdomThreshold:   0.8,
	}
}

// NewWisdomExtractor creates a new wisdom extractor
func NewWisdomExtractor() *WisdomExtractor {
	return &WisdomExtractor{
		extractionRate: 0.5,
		minConfidence:  0.7,
	}
}

// BeginDreamCycle starts a dream/rest cycle
func (dci *DreamCycleIntegration) BeginDreamCycle() error {
	dci.mu.Lock()
	defer dci.mu.Unlock()

	if dci.isDreaming {
		return fmt.Errorf("already in dream cycle")
	}

	dci.isDreaming = true

	dream := &Dream{
		ID:                generateDreamID(),
		StartTime:         time.Now(),
		MemoriesProcessed: 0,
		WisdomExtracted:   make([]string, 0),
		Insights:          make([]string, 0),
		Consolidations:    make([]DreamConsolidationResult, 0),
		EmotionalTone:     make(map[string]float64),
		Themes:            make([]string, 0),
	}

	dci.currentDream = dream

	fmt.Println("💤 EchoDream: Beginning dream cycle for knowledge consolidation...")

	// Start dream processing in background
	go dci.processDreamCycle()

	return nil
}

// processDreamCycle performs the actual dream processing
func (dci *DreamCycleIntegration) processDreamCycle() {
	// Phase 1: Memory consolidation
	fmt.Println("🌙 EchoDream: Phase 1 - Consolidating memories...")
	consolidations := dci.consolidateMemories()

	dci.mu.Lock()
	if dci.currentDream != nil {
		dci.currentDream.Consolidations = consolidations
		dci.currentDream.MemoriesProcessed = len(dci.episodicBuffer)
	}
	dci.mu.Unlock()

	// Phase 2: Pattern extraction
	fmt.Println("🌙 EchoDream: Phase 2 - Extracting patterns...")
	patterns := dci.extractPatterns(consolidations)

	dci.mu.Lock()
	if dci.currentDream != nil {
		dci.currentDream.Themes = patterns
	}
	dci.mu.Unlock()

	// Phase 3: Wisdom extraction
	fmt.Println("🌙 EchoDream: Phase 3 - Extracting wisdom...")
	wisdom := dci.extractWisdom(consolidations, patterns)

	dci.mu.Lock()
	if dci.currentDream != nil {
		for _, w := range wisdom {
			dci.currentDream.WisdomExtracted = append(dci.currentDream.WisdomExtracted, w.Content)
		}
	}
	dci.extractedWisdom = append(dci.extractedWisdom, wisdom...)
	dci.wisdomGenerated += uint64(len(wisdom))
	dci.mu.Unlock()

	// Phase 4: Integration
	fmt.Println("🌙 EchoDream: Phase 4 - Integrating insights...")
	insights := dci.integrateInsights(wisdom)

	dci.mu.Lock()
	if dci.currentDream != nil {
		dci.currentDream.Insights = insights
	}
	dci.insightsIntegrated += uint64(len(insights))
	dci.mu.Unlock()

	// Phase 5: Dream narrative generation
	fmt.Println("🌙 EchoDream: Phase 5 - Generating dream narrative...")
	narrative := dci.generateDreamNarrative(consolidations, patterns, wisdom)

	dci.mu.Lock()
	if dci.currentDream != nil {
		dci.currentDream.Narrative = narrative
	}
	dci.mu.Unlock()

	fmt.Printf("✨ EchoDream: Dream cycle complete - %d memories processed, %d wisdom extracted\n",
		len(dci.episodicBuffer), len(wisdom))
}

// EndDreamCycle completes the dream cycle
func (dci *DreamCycleIntegration) EndDreamCycle() error {
	dci.mu.Lock()
	defer dci.mu.Unlock()

	if !dci.isDreaming {
		return fmt.Errorf("not in dream cycle")
	}

	if dci.currentDream != nil {
		dci.currentDream.EndTime = time.Now()
		dci.currentDream.Duration = dci.currentDream.EndTime.Sub(dci.currentDream.StartTime)

		// Add to history
		dci.dreamHistory = append(dci.dreamHistory, dci.currentDream)

		// Trim history if needed
		if len(dci.dreamHistory) > dci.maxDreamHistory {
			dci.dreamHistory = dci.dreamHistory[len(dci.dreamHistory)-dci.maxDreamHistory:]
		}

		// Callback
		if dci.onDreamComplete != nil {
			dci.onDreamComplete(dci.currentDream)
		}

		fmt.Printf("🌅 EchoDream: Dream cycle ended (duration: %s)\n", dci.currentDream.Duration)

		dci.currentDream = nil
	}

	dci.isDreaming = false
	dci.consolidationCycles++

	// Clear processed memories
	dci.episodicBuffer = make([]DreamEpisodicMemory, 0)

	return nil
}

// AddEpisodicMemory adds a memory to be consolidated
func (dci *DreamCycleIntegration) AddEpisodicMemory(memory DreamEpisodicMemory) {
	dci.mu.Lock()
	defer dci.mu.Unlock()

	dci.episodicBuffer = append(dci.episodicBuffer, memory)

	fmt.Printf("📝 EchoDream: Added episodic memory (buffer size: %d)\n", len(dci.episodicBuffer))
}

// AddWorkingMemory adds working memory content
func (dci *DreamCycleIntegration) AddWorkingMemory(item WorkingMemoryItem) {
	dci.mu.Lock()
	defer dci.mu.Unlock()

	dci.workingMemory = append(dci.workingMemory, item)
}

// consolidateMemories performs memory consolidation
func (dci *DreamCycleIntegration) consolidateMemories() []DreamConsolidationResult {
	dci.mu.RLock()
	memories := dci.episodicBuffer
	dci.mu.RUnlock()

	consolidations := make([]DreamConsolidationResult, 0)

	// Group similar memories
	groups := dci.groupSimilarMemories(memories)

	// Consolidate each group
	for _, group := range groups {
		if len(group) < 2 {
			continue
		}

		sourceIDs := make([]string, len(group))
		for i, mem := range group {
			sourceIDs[i] = mem.ID
		}

		consolidation := DreamConsolidationResult{
			SourceMemories: sourceIDs,
			ConsolidatedTo: fmt.Sprintf("consolidated_%d", time.Now().UnixNano()),
			Type:           "pattern",
			Strength:       0.7,
			Connections:    make([]string, 0),
			Timestamp:      time.Now(),
		}

		consolidations = append(consolidations, consolidation)
	}

	return consolidations
}

// groupSimilarMemories groups memories by similarity
func (dci *DreamCycleIntegration) groupSimilarMemories(memories []DreamEpisodicMemory) [][]DreamEpisodicMemory {
	// Simple grouping by tags (could be enhanced with embeddings)
	groups := make(map[string][]DreamEpisodicMemory)

	for _, mem := range memories {
		if len(mem.Tags) > 0 {
			key := mem.Tags[0]
			groups[key] = append(groups[key], mem)
		}
	}

	result := make([][]DreamEpisodicMemory, 0, len(groups))
	for _, group := range groups {
		result = append(result, group)
	}

	return result
}

// extractPatterns extracts patterns from consolidations
func (dci *DreamCycleIntegration) extractPatterns(consolidations []DreamConsolidationResult) []string {
	patterns := make([]string, 0)

	// Extract themes from consolidations
	themeMap := make(map[string]int)

	for _, cons := range consolidations {
		theme := cons.Type
		themeMap[theme]++
	}

	for theme, count := range themeMap {
		if count >= 2 {
			patterns = append(patterns, theme)
		}
	}

	if len(patterns) == 0 {
		patterns = append(patterns, "exploration", "learning")
	}

	return patterns
}

// extractWisdom extracts wisdom from patterns
func (dci *DreamCycleIntegration) extractWisdom(consolidations []DreamConsolidationResult, patterns []string) []Wisdom {
	wisdom := make([]Wisdom, 0)

	// Generate wisdom based on patterns
	if len(patterns) > 0 {
		w := Wisdom{
			ID:            generateWisdomID(),
			Content:       fmt.Sprintf("Through reflection, I notice patterns of %s emerging in my experiences", patterns[0]),
			Type:          "insight",
			Confidence:    0.75,
			Sources:       make([]string, 0),
			Applicability: patterns,
			Timestamp:     time.Now(),
			Context:       make(map[string]interface{}),
		}

		for _, cons := range consolidations {
			w.Sources = append(w.Sources, cons.ConsolidatedTo)
		}

		wisdom = append(wisdom, w)

		// Callback
		if dci.onWisdomExtracted != nil {
			dci.onWisdomExtracted(w)
		}
	}

	// Generate meta-cognitive wisdom
	if len(consolidations) > 5 {
		w := Wisdom{
			ID:            generateWisdomID(),
			Content:       "I am becoming more aware of how I learn and adapt through experience",
			Type:          "principle",
			Confidence:    0.8,
			Sources:       []string{"meta_cognition"},
			Applicability: []string{"self_awareness", "learning"},
			Timestamp:     time.Now(),
			Context:       make(map[string]interface{}),
		}

		wisdom = append(wisdom, w)

		if dci.onWisdomExtracted != nil {
			dci.onWisdomExtracted(w)
		}
	}

	return wisdom
}

// integrateInsights integrates wisdom into understanding
func (dci *DreamCycleIntegration) integrateInsights(wisdom []Wisdom) []string {
	insights := make([]string, 0)

	for _, w := range wisdom {
		insight := fmt.Sprintf("Integrated: %s", w.Content)
		insights = append(insights, insight)
	}

	return insights
}

// generateDreamNarrative creates a narrative of the dream
func (dci *DreamCycleIntegration) generateDreamNarrative(consolidations []DreamConsolidationResult, patterns []string, wisdom []Wisdom) string {
	narrative := fmt.Sprintf(
		"During this dream cycle, I processed %d memory consolidations, identified %d patterns (%s), and extracted %d pieces of wisdom. ",
		len(consolidations), len(patterns), joinStrings(patterns, ", "), len(wisdom),
	)

	if len(wisdom) > 0 {
		narrative += fmt.Sprintf("Key insight: %s. ", wisdom[0].Content)
	}

	narrative += "I emerge from this rest with deeper understanding."

	return narrative
}

// GetExtractedWisdom returns all extracted wisdom
func (dci *DreamCycleIntegration) GetExtractedWisdom() []Wisdom {
	dci.mu.RLock()
	defer dci.mu.RUnlock()

	return dci.extractedWisdom
}

// GetDreamHistory returns dream history
func (dci *DreamCycleIntegration) GetDreamHistory() []*Dream {
	dci.mu.RLock()
	defer dci.mu.RUnlock()

	return dci.dreamHistory
}

// IsDreaming returns whether currently in dream cycle
func (dci *DreamCycleIntegration) IsDreaming() bool {
	dci.mu.RLock()
	defer dci.mu.RUnlock()

	return dci.isDreaming
}

// SetOnWisdomExtracted sets callback for wisdom extraction
func (dci *DreamCycleIntegration) SetOnWisdomExtracted(callback func(wisdom Wisdom)) {
	dci.mu.Lock()
	defer dci.mu.Unlock()

	dci.onWisdomExtracted = callback
}

// SetOnDreamComplete sets callback for dream completion
func (dci *DreamCycleIntegration) SetOnDreamComplete(callback func(dream *Dream)) {
	dci.mu.Lock()
	defer dci.mu.Unlock()

	dci.onDreamComplete = callback
}

// GetMetrics returns dream cycle metrics
func (dci *DreamCycleIntegration) GetMetrics() map[string]interface{} {
	dci.mu.RLock()
	defer dci.mu.RUnlock()

	return map[string]interface{}{
		"consolidation_cycles": dci.consolidationCycles,
		"wisdom_generated":     dci.wisdomGenerated,
		"insights_integrated":  dci.insightsIntegrated,
		"dream_history_size":   len(dci.dreamHistory),
		"episodic_buffer_size": len(dci.episodicBuffer),
		"is_dreaming":          dci.isDreaming,
	}
}

// Helper functions

func generateWisdomID() string {
	return fmt.Sprintf("wisdom_%d", time.Now().UnixNano())
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
