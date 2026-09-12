package echodream

import (
	"context"
	"fmt"
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

// processDreamPhase processes the current dream phase
func (ed *EchoDream) processDreamPhase() {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	switch ed.dreamPhase {
	case PhaseREM:
		// Process recent memories
		ed.processRecentMemories()
		ed.dreamPhase = PhaseDeepSleep

	case PhaseDeepSleep:
		// Consolidate memories into knowledge
		ed.consolidateMemories()
		ed.dreamPhase = PhaseConsolidation

	case PhaseConsolidation:
		// Extract wisdom from knowledge
		ed.extractWisdom()
		ed.dreamPhase = PhaseIntegration

	case PhaseIntegration:
		// Integrate wisdom into cognitive system
		ed.integrateWisdom()
		ed.dreamPhase = PhaseREM
	}
}

// processRecentMemories processes recent episodic memories
func (ed *EchoDream) processRecentMemories() {
	// In real implementation, would process actual memories
	// For now, simulate processing
	processed := 0
	for i := range ed.episodicMemories {
		if !ed.episodicMemories[i].Consolidated {
			ed.episodicMemories[i].Consolidated = true
			processed++
		}
	}
	ed.memoriesProcessed += uint64(processed)
}

// consolidateMemories consolidates memories into knowledge
func (ed *EchoDream) consolidateMemories() {
	// Simulate knowledge consolidation
	if len(ed.episodicMemories) > 0 {
		knowledge := KnowledgeItem{
			ID:         fmt.Sprintf("knowledge_%d", time.Now().UnixNano()),
			Content:    "Consolidated knowledge from recent experiences",
			Confidence: 0.8,
			Created:    time.Now(),
		}
		ed.consolidatedKnowledge = append(ed.consolidatedKnowledge, knowledge)
	}
}

// extractWisdom extracts wisdom from consolidated knowledge
func (ed *EchoDream) extractWisdom() {
	// Simulate wisdom extraction
	if len(ed.consolidatedKnowledge) > 0 {
		wisdom := WisdomInsight{
			ID:            fmt.Sprintf("wisdom_%d", time.Now().UnixNano()),
			Insight:       "Wisdom insight from integrated knowledge",
			Depth:         0.7,
			Applicability: 0.8,
			Created:       time.Now(),
		}
		ed.wisdomInsights = append(ed.wisdomInsights, wisdom)
		ed.wisdomExtracted++
	}
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
