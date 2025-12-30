package echodream

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/consciousness"
	"github.com/cogpy/echo9llama/core/llm"
)

// DreamSystem is the main interface for knowledge integration during rest cycles
// It wraps EchoDream with LLM-based consolidation capabilities
type DreamSystem struct {
	mu                    sync.RWMutex
	ctx                   context.Context
	cancel                context.CancelFunc
	
	// Core dream processor
	echoDream             *EchoDream
	
	// LLM-based consolidation
	llmConsolidator       *LLMConsolidator
	llmProvider           llm.Provider
	
	// Thought buffer for consolidation
	thoughtBuffer         []*consciousness.Thought
	maxThoughtBuffer      int
	
	// Running state
	running               bool
}

// NewDreamSystem creates a new dream system with LLM-based consolidation
func NewDreamSystem() *DreamSystem {
	ctx, cancel := context.WithCancel(context.Background())
	
	echoDream := NewEchoDream()
	
	// Try to get LLM provider for consolidation
	var llmConsolidator *LLMConsolidator
	var llmProv llm.Provider
	
	// Try Anthropic first
	anthropicProvider := llm.NewAnthropicProvider("")
	if anthropicProvider.Available() {
		llmProv = anthropicProvider
		llmConsolidator = NewLLMConsolidator(anthropicProvider)
	} else {
		// Try OpenRouter as fallback
		openrouterProvider := llm.NewOpenRouterProvider("")
		if openrouterProvider.Available() {
			llmProv = openrouterProvider
			llmConsolidator = NewLLMConsolidator(openrouterProvider)
		}
	}
	
	return &DreamSystem{
		ctx:              ctx,
		cancel:           cancel,
		echoDream:        echoDream,
		llmConsolidator:  llmConsolidator,
		llmProvider:      llmProv,
		thoughtBuffer:    make([]*consciousness.Thought, 0),
		maxThoughtBuffer: 100,
	}
}

// Start begins dream processing with LLM-based consolidation
func (ds *DreamSystem) Start() error {
	ds.mu.Lock()
	if ds.running {
		ds.mu.Unlock()
		return fmt.Errorf("dream system already running")
	}
	ds.running = true
	ds.mu.Unlock()
	
	// Start the underlying EchoDream
	if err := ds.echoDream.Start(); err != nil {
		return err
	}
	
	// Perform LLM-based consolidation if available
	if ds.llmConsolidator != nil {
		go ds.performLLMConsolidation()
	}
	
	return nil
}

// Stop ends dream processing
func (ds *DreamSystem) Stop() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	
	if !ds.running {
		return fmt.Errorf("dream system not running")
	}
	
	ds.running = false
	
	return ds.echoDream.Stop()
}

// AddThought adds a thought to the buffer for later consolidation
func (ds *DreamSystem) AddThought(thought *consciousness.Thought) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	
	ds.thoughtBuffer = append(ds.thoughtBuffer, thought)
	
	// Trim buffer if too large
	if len(ds.thoughtBuffer) > ds.maxThoughtBuffer {
		ds.thoughtBuffer = ds.thoughtBuffer[len(ds.thoughtBuffer)-ds.maxThoughtBuffer:]
	}
	
	// Also add to EchoDream as episodic memory
	ds.echoDream.AddEpisodicMemory(thought.Content, 0.7)
}

// performLLMConsolidation performs real LLM-based knowledge consolidation
func (ds *DreamSystem) performLLMConsolidation() {
	ds.mu.RLock()
	thoughts := make([]*consciousness.Thought, len(ds.thoughtBuffer))
	copy(thoughts, ds.thoughtBuffer)
	ds.mu.RUnlock()
	
	if len(thoughts) == 0 {
		fmt.Println("   ⚠️  No thoughts to consolidate")
		return
	}
	
	fmt.Printf("   🧠 Consolidating %d thoughts into knowledge...\n", len(thoughts))
	
	// Phase 1: Consolidate thoughts into knowledge
	ctx, cancel := context.WithTimeout(ds.ctx, 30*time.Second)
	defer cancel()
	
	knowledgeItems, err := ds.llmConsolidator.ConsolidateThoughtsToKnowledge(ctx, thoughts)
	if err != nil {
		fmt.Printf("   ⚠️  Knowledge consolidation error: %v\n", err)
		return
	}
	
	fmt.Printf("   ✓ Extracted %d knowledge items\n", len(knowledgeItems))
	
	// Store knowledge in EchoDream
	ds.echoDream.mu.Lock()
	for _, item := range knowledgeItems {
		ds.echoDream.consolidatedKnowledge = append(ds.echoDream.consolidatedKnowledge, item)
	}
	ds.echoDream.mu.Unlock()
	
	// Display knowledge items
	for i, item := range knowledgeItems {
		fmt.Printf("   📚 Knowledge %d: %s\n", i+1, truncateString(item.Content, 80))
	}
	
	// Phase 2: Extract wisdom from knowledge
	fmt.Println("   💎 Extracting wisdom insights...")
	
	ctx2, cancel2 := context.WithTimeout(ds.ctx, 30*time.Second)
	defer cancel2()
	
	wisdomInsights, err := ds.llmConsolidator.ExtractWisdomFromKnowledge(ctx2, knowledgeItems)
	if err != nil {
		fmt.Printf("   ⚠️  Wisdom extraction error: %v\n", err)
		return
	}
	
	fmt.Printf("   ✓ Extracted %d wisdom insights\n", len(wisdomInsights))
	
	// Store wisdom in EchoDream
	ds.echoDream.mu.Lock()
	for _, wisdom := range wisdomInsights {
		ds.echoDream.wisdomInsights = append(ds.echoDream.wisdomInsights, wisdom)
		ds.echoDream.wisdomExtracted++
	}
	ds.echoDream.mu.Unlock()
	
	// Display wisdom insights
	for i, wisdom := range wisdomInsights {
		fmt.Printf("   ✨ Wisdom %d: %s (Depth: %.2f, Applicability: %.2f)\n", 
			i+1, truncateString(wisdom.Insight, 80), wisdom.Depth, wisdom.Applicability)
	}
}

// GetMetrics returns dream system metrics
func (ds *DreamSystem) GetMetrics() map[string]interface{} {
	ds.mu.RLock()
	thoughtCount := len(ds.thoughtBuffer)
	ds.mu.RUnlock()
	
	echoDreamMetrics := ds.echoDream.GetMetrics()
	
	return map[string]interface{}{
		"thoughts_buffered":   thoughtCount,
		"llm_enabled":         ds.llmConsolidator != nil,
		"echo_dream_metrics":  echoDreamMetrics,
	}
}

// GetWisdomInsights returns all extracted wisdom insights
func (ds *DreamSystem) GetWisdomInsights() []WisdomInsight {
	ds.echoDream.mu.RLock()
	defer ds.echoDream.mu.RUnlock()
	
	insights := make([]WisdomInsight, len(ds.echoDream.wisdomInsights))
	copy(insights, ds.echoDream.wisdomInsights)
	return insights
}

// GetKnowledgeItems returns all consolidated knowledge
func (ds *DreamSystem) GetKnowledgeItems() []KnowledgeItem {
	ds.echoDream.mu.RLock()
	defer ds.echoDream.mu.RUnlock()
	
	items := make([]KnowledgeItem, len(ds.echoDream.consolidatedKnowledge))
	copy(items, ds.echoDream.consolidatedKnowledge)
	return items
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
