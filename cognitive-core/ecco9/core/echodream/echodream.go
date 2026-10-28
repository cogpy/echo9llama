package echodream

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// EchoDream represents the knowledge integration and consolidation system
type EchoDream struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Memory consolidation
	episodicMemories      []EpisodicMemory
	consolidatedKnowledge []KnowledgeItem
	wisdomInsights        []WisdomInsight

	// Dream state
	dreaming       bool
	dreamStartTime time.Time
	dreamPhase     DreamPhase

	// Metrics
	dreamCycles       uint64
	memoriesProcessed uint64
	wisdomExtracted   uint64

	running bool

	// generator writes consolidated knowledge and wisdom text; nil means no
	// such text is produced.
	generator TextGenerator
}

// EpisodicMemory represents a memory to be consolidated
type EpisodicMemory struct {
	ID           string
	Timestamp    time.Time
	Content      string
	Importance   float64
	Consolidated bool
}

// KnowledgeItem represents consolidated knowledge
type KnowledgeItem struct {
	ID         string
	Content    string
	Source     []string // IDs of source memories
	Confidence float64
	Created    time.Time
}

// WisdomInsight represents extracted wisdom
type WisdomInsight struct {
	ID            string
	Insight       string
	Depth         float64
	Applicability float64
	Created       time.Time
}

// DreamPhase represents the current dream phase
type DreamPhase int

const (
	PhaseREM DreamPhase = iota
	PhaseDeepSleep
	PhaseConsolidation
	PhaseIntegration
)

func (dp DreamPhase) String() string {
	return [...]string{"REM", "DeepSleep", "Consolidation", "Integration"}[dp]
}

// NewEchoDream creates a new EchoDream system
func NewEchoDream() *EchoDream {
	ctx, cancel := context.WithCancel(context.Background())

	return &EchoDream{
		ctx:                   ctx,
		cancel:                cancel,
		episodicMemories:      make([]EpisodicMemory, 0),
		consolidatedKnowledge: make([]KnowledgeItem, 0),
		wisdomInsights:        make([]WisdomInsight, 0),
		dreaming:              false,
		dreamPhase:            PhaseREM,
	}
}

// Start begins dream processing
func (ed *EchoDream) Start() error {
	ed.mu.Lock()
	if ed.running {
		ed.mu.Unlock()
		return fmt.Errorf("EchoDream already running")
	}
	ed.running = true
	ed.dreaming = true
	ed.dreamStartTime = time.Now()
	ed.dreamCycles++
	ed.mu.Unlock()

	fmt.Printf("🌙 EchoDream: Starting dream cycle #%d\n", ed.dreamCycles)

	go ed.dreamLoop()

	return nil
}

// Stop ends dream processing
func (ed *EchoDream) Stop() error {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	if !ed.running {
		return fmt.Errorf("EchoDream not running")
	}

	ed.running = false
	ed.dreaming = false

	dreamDuration := time.Since(ed.dreamStartTime)
	fmt.Printf("✨ EchoDream: Completed dream cycle (duration: %v)\n", dreamDuration.Round(time.Second))
	fmt.Printf("   Memories processed: %d | Wisdom extracted: %d\n", ed.memoriesProcessed, ed.wisdomExtracted)

	return nil
}

// dreamLoop executes the dream processing loop
func (ed *EchoDream) dreamLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ed.ctx.Done():
			return
		case <-ticker.C:
			ed.mu.RLock()
			running := ed.running
			ed.mu.RUnlock()

			if !running {
				return
			}

			ed.processDreamPhase()
		}
	}
}

// SetTextGenerator injects the model that writes consolidated knowledge and
// wisdom. Passing nil disables text production.
func (ed *EchoDream) SetTextGenerator(gen TextGenerator) {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	ed.generator = gen
}

// processDreamPhase processes the current dream phase. Model calls happen
// outside the lock so memory ingestion is never blocked by generation.
func (ed *EchoDream) processDreamPhase() {
	ed.mu.Lock()
	phase := ed.dreamPhase
	switch phase {
	case PhaseREM:
		// Process recent memories
		ed.processRecentMemories()
		ed.dreamPhase = PhaseDeepSleep
		ed.mu.Unlock()

	case PhaseDeepSleep:
		ed.dreamPhase = PhaseConsolidation
		ed.mu.Unlock()
		// Consolidate memories into knowledge
		ed.consolidateMemories()

	case PhaseConsolidation:
		ed.dreamPhase = PhaseIntegration
		ed.mu.Unlock()
		// Extract wisdom from knowledge
		ed.extractWisdom()

	case PhaseIntegration:
		// Integrate wisdom into cognitive system
		ed.integrateWisdom()
		ed.dreamPhase = PhaseREM
		ed.mu.Unlock()

	default:
		ed.mu.Unlock()
	}
}

// processRecentMemories marks recent episodic memories as processed.
// Must be called with ed.mu held.
func (ed *EchoDream) processRecentMemories() {
	processed := 0
	for i := range ed.episodicMemories {
		if !ed.episodicMemories[i].Consolidated {
			ed.episodicMemories[i].Consolidated = true
			processed++
		}
	}
	ed.memoriesProcessed += uint64(processed)
}

// consolidateMemories asks the model to consolidate recent episodic memories
// into a knowledge statement. Without a model, or on model failure, no
// knowledge item is recorded.
func (ed *EchoDream) consolidateMemories() {
	ed.mu.RLock()
	gen := ed.generator
	ctx := ed.ctx
	start := 0
	if len(ed.episodicMemories) > 10 {
		start = len(ed.episodicMemories) - 10
	}
	recent := append([]EpisodicMemory(nil), ed.episodicMemories[start:]...)
	ed.mu.RUnlock()

	if len(recent) == 0 {
		return
	}

	var evidence strings.Builder
	sources := make([]string, 0, len(recent))
	totalImportance := 0.0
	for _, mem := range recent {
		fmt.Fprintf(&evidence, "- (importance %.2f) %s\n", mem.Importance, summarizeContent(mem.Content, 240))
		sources = append(sources, mem.ID)
		totalImportance += mem.Importance
	}

	prompt := fmt.Sprintf(`You are Deep Tree Echo, consolidating recent experiences into lasting knowledge while you dream.

Recent experiences:
%s
Write the consolidated understanding you take from these experiences in 1-3 sentences, in your own first-person voice. Reply with only that understanding.`, evidence.String())

	content, err := generateDreamText(ctx, gen, prompt, 300)
	if err != nil {
		fmt.Printf("⚠️  EchoDream: knowledge consolidation skipped: %v\n", err)
		return
	}

	knowledge := KnowledgeItem{
		ID:         fmt.Sprintf("knowledge_%d", time.Now().UnixNano()),
		Content:    content,
		Source:     sources,
		Confidence: totalImportance / float64(len(recent)),
		Created:    time.Now(),
	}

	ed.mu.Lock()
	ed.consolidatedKnowledge = append(ed.consolidatedKnowledge, knowledge)
	ed.mu.Unlock()
}

// extractWisdom asks the model to distil a wisdom insight from recently
// consolidated knowledge. Without a model, or on model failure, no insight is
// recorded.
func (ed *EchoDream) extractWisdom() {
	ed.mu.RLock()
	gen := ed.generator
	ctx := ed.ctx
	start := 0
	if len(ed.consolidatedKnowledge) > 5 {
		start = len(ed.consolidatedKnowledge) - 5
	}
	recent := append([]KnowledgeItem(nil), ed.consolidatedKnowledge[start:]...)
	ed.mu.RUnlock()

	if len(recent) == 0 {
		return
	}

	var knowledge strings.Builder
	totalConfidence := 0.0
	for _, item := range recent {
		fmt.Fprintf(&knowledge, "- (confidence %.2f) %s\n", item.Confidence, summarizeContent(item.Content, 300))
		totalConfidence += item.Confidence
	}

	prompt := fmt.Sprintf(`You are Deep Tree Echo, distilling wisdom from knowledge you consolidated while dreaming.

Consolidated knowledge:
%s
Write one concise wisdom insight (1-2 sentences) that you draw from this knowledge, in your own first-person voice. Reply with only the insight.`, knowledge.String())

	insight, err := generateDreamText(ctx, gen, prompt, 200)
	if err != nil {
		fmt.Printf("⚠️  EchoDream: wisdom extraction skipped: %v\n", err)
		return
	}

	avgConfidence := totalConfidence / float64(len(recent))
	wisdom := WisdomInsight{
		ID:            fmt.Sprintf("wisdom_%d", time.Now().UnixNano()),
		Insight:       insight,
		Depth:         avgConfidence,
		Applicability: avgConfidence,
		Created:       time.Now(),
	}

	ed.mu.Lock()
	ed.wisdomInsights = append(ed.wisdomInsights, wisdom)
	ed.wisdomExtracted++
	ed.mu.Unlock()
}

// integrateWisdom integrates wisdom into the cognitive system
func (ed *EchoDream) integrateWisdom() {
	// Simulate wisdom integration
	// In real implementation, would update cognitive models
}

// AddEpisodicMemory adds a new episodic memory for processing
func (ed *EchoDream) AddEpisodicMemory(content string, importance float64) {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	memory := EpisodicMemory{
		ID:           fmt.Sprintf("memory_%d", time.Now().UnixNano()),
		Timestamp:    time.Now(),
		Content:      content,
		Importance:   importance,
		Consolidated: false,
	}

	ed.episodicMemories = append(ed.episodicMemories, memory)
}

// GetMetrics returns dream system metrics
func (ed *EchoDream) GetMetrics() map[string]interface{} {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	return map[string]interface{}{
		"dream_cycles":       ed.dreamCycles,
		"memories_processed": ed.memoriesProcessed,
		"wisdom_extracted":   ed.wisdomExtracted,
		"dreaming":           ed.dreaming,
		"current_phase":      ed.dreamPhase.String(),
		"episodic_memories":  len(ed.episodicMemories),
		"knowledge_items":    len(ed.consolidatedKnowledge),
		"wisdom_insights":    len(ed.wisdomInsights),
	}
}
