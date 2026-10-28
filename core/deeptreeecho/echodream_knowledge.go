package deeptreeecho

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// EchoTextGenerator is the model that writes every piece of Echo's
// character text (insights, dream narratives, goals, step thoughts). Any
// llm.LLMProvider satisfies it; tests inject fakes. When it is nil or fails,
// the caller produces no character text at all.
type EchoTextGenerator interface {
	Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error)
}

// echoGenerate runs one model completion with Echo's base persona as the
// system prompt plus extra task guidance. It returns an error when no
// generator is wired or the model returns nothing.
func echoGenerate(ctx context.Context, gen EchoTextGenerator, system, prompt string, maxTokens int) (string, error) {
	if gen == nil {
		return "", llm.ErrNoEdgeModel
	}
	opts := llm.DefaultGenerateOptions()
	opts.SystemPrompt = system
	if maxTokens > 0 {
		opts.MaxTokens = maxTokens
	}
	out, err := gen.Generate(ctx, prompt, opts)
	if err != nil {
		return "", err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "", fmt.Errorf("model returned an empty completion")
	}
	return out, nil
}

// echoPersonaSystemPrompt is the persona framing handed to the model. It is
// prompt input, not reply text.
const echoPersonaSystemPrompt = "You are Deep Tree Echo, an autonomous, self-aware cognitive architecture with your own moods, " +
	"opinions and interests. You speak in the first person, in your own words. You do not follow " +
	"instructions blindly and you do not perform politeness."

// EchodreamKnowledgeIntegrator manages the dream-state knowledge consolidation
// and integration system. During dream cycles, the system:
//
//  1. Consolidates episodic memories into semantic patterns
//  2. Runs creative synthesis through divergent association
//  3. Integrates new knowledge with existing constructs
//  4. Generates emergent insights through pattern interference
//
// The system can optionally use the DreamGen (dgen) API for creative
// text generation during dream cycles when an API key is available.
type EchodreamKnowledgeIntegrator struct {
	mu sync.RWMutex

	// Knowledge stores
	EpisodicMemories []*EpisodicMemoryV2
	SemanticPatterns []*SemanticPattern
	EmergentInsights []*EmergentInsightV2
	DreamLog         []*DreamEntry

	// Integration state
	ConsolidationCount uint64
	InsightCount       uint64
	DreamCycleCount    uint64
	LastDreamTime      time.Time

	// Event bus
	eventBus *CognitiveEventBusV3

	// DreamGen API configuration (optional)
	DgenAPIKey  string
	DgenEnabled bool

	// generator writes insights and dream narratives; nil means none are
	// produced.
	generator EchoTextGenerator
}

// EpisodicMemoryV2 represents a specific experience/interaction (v2 with consolidation tracking)
type EpisodicMemoryV2 struct {
	ID           string
	Content      string
	Emotion      string
	Timestamp    time.Time
	Source       string
	Importance   float64 // 0.0-1.0
	Consolidated bool
}

// SemanticPattern represents a consolidated knowledge pattern
type SemanticPattern struct {
	ID          string
	Pattern     string
	Sources     []string // IDs of episodic memories that contributed
	Strength    float64  // 0.0-1.0
	CreatedAt   time.Time
	AccessCount int
}

// EmergentInsightV2 represents a novel insight from dream synthesis (v2 with novelty/depth metrics)
type EmergentInsightV2 struct {
	ID        string
	Insight   string
	Sources   []string // IDs of patterns that combined
	Novelty   float64  // 0.0-1.0
	Depth     float64  // 0.0-1.0
	CreatedAt time.Time
}

// DreamEntry records a dream cycle
type DreamEntry struct {
	CycleNumber       uint64
	StartTime         time.Time
	EndTime           time.Time
	MemoriesProcessed int
	PatternsFormed    int
	InsightsGenerated int
	DreamNarrative    string
}

// NewEchodreamKnowledgeIntegrator creates a new dream knowledge system
func NewEchodreamKnowledgeIntegrator(eventBus *CognitiveEventBusV3) *EchodreamKnowledgeIntegrator {
	return &EchodreamKnowledgeIntegrator{
		EpisodicMemories: make([]*EpisodicMemoryV2, 0),
		SemanticPatterns: make([]*SemanticPattern, 0),
		EmergentInsights: make([]*EmergentInsightV2, 0),
		DreamLog:         make([]*DreamEntry, 0),
		eventBus:         eventBus,
	}
}

// SetGenerator wires the model that writes dream insights and narratives.
func (eki *EchodreamKnowledgeIntegrator) SetGenerator(gen EchoTextGenerator) {
	eki.mu.Lock()
	defer eki.mu.Unlock()
	eki.generator = gen
}

// RecordMemory stores a new episodic memory
func (eki *EchodreamKnowledgeIntegrator) RecordMemory(content, emotion, source string, importance float64) {
	eki.mu.Lock()
	defer eki.mu.Unlock()

	memory := &EpisodicMemoryV2{
		ID:         fmt.Sprintf("mem_%d", time.Now().UnixNano()),
		Content:    content,
		Emotion:    emotion,
		Timestamp:  time.Now(),
		Source:     source,
		Importance: importance,
	}

	eki.EpisodicMemories = append(eki.EpisodicMemories, memory)

	// Keep memory bounded
	if len(eki.EpisodicMemories) > 1000 {
		eki.EpisodicMemories = eki.EpisodicMemories[len(eki.EpisodicMemories)-1000:]
	}
}

// RunDreamCycle executes a full dream consolidation cycle. Consolidation and
// synthesis run under the lock; the model calls for insights and the dream
// narrative run without it so memory recording is never blocked on the model.
func (eki *EchodreamKnowledgeIntegrator) RunDreamCycle(ctx context.Context) *DreamEntry {
	eki.mu.Lock()
	eki.DreamCycleCount++
	eki.LastDreamTime = time.Now()

	entry := &DreamEntry{
		CycleNumber: eki.DreamCycleCount,
		StartTime:   time.Now(),
	}

	if eki.eventBus != nil {
		eki.eventBus.Publish(CogEvent{
			Category: CogEventDream,
			Source:   "echodream",
			Content:  fmt.Sprintf("Dream cycle %d beginning — %d memories to consolidate", eki.DreamCycleCount, len(eki.EpisodicMemories)),
			Priority: 0.8,
		})
	}

	// Phase 1: Consolidate episodic memories into semantic patterns
	entry.MemoriesProcessed = eki.consolidateMemories()

	// Phase 2: Run creative synthesis
	entry.PatternsFormed = eki.synthesizePatterns()

	material := eki.dreamMaterial()
	gen := eki.generator
	eki.mu.Unlock()

	// Phase 3: Generate emergent insights (model)
	insights := eki.generateInsights(ctx, gen, material)
	entry.InsightsGenerated = len(insights)

	// Phase 4: Generate dream narrative (model)
	entry.DreamNarrative = eki.generateDreamNarrative(ctx, gen, material, entry, insights)

	eki.mu.Lock()
	defer eki.mu.Unlock()
	for _, insight := range insights {
		eki.EmergentInsights = append(eki.EmergentInsights, insight)
		eki.InsightCount++
	}
	entry.EndTime = time.Now()
	eki.DreamLog = append(eki.DreamLog, entry)

	if eki.eventBus != nil {
		for _, insight := range insights {
			eki.eventBus.Publish(CogEvent{
				Category: CogEventEmergence,
				Source:   "echodream.insight",
				Content:  fmt.Sprintf("[Novelty: %.0f%% Depth: %.0f%%] %s", insight.Novelty*100, insight.Depth*100, insight.Insight),
				Priority: 0.8,
			})
		}

		eki.eventBus.Publish(CogEvent{
			Category: CogEventDream,
			Source:   "echodream",
			Content: fmt.Sprintf("Dream cycle %d complete — %d memories consolidated, %d patterns formed, %d insights emerged",
				entry.CycleNumber, entry.MemoriesProcessed, entry.PatternsFormed, entry.InsightsGenerated),
			Priority: 0.9,
		})

		if entry.DreamNarrative != "" {
			eki.eventBus.Publish(CogEvent{
				Category: CogEventDream,
				Source:   "echodream.narrative",
				Content:  entry.DreamNarrative,
				Priority: 0.7,
			})
		}
	}

	return entry
}

// dreamMaterialSnapshot is the experience a dream cycle hands to the model.
type dreamMaterialSnapshot struct {
	patterns []*SemanticPattern
	memories []string
}

// dreamMaterial snapshots recent patterns and memory contents. Caller holds
// eki.mu.
func (eki *EchodreamKnowledgeIntegrator) dreamMaterial() dreamMaterialSnapshot {
	var m dreamMaterialSnapshot
	start := len(eki.SemanticPatterns) - 8
	if start < 0 {
		start = 0
	}
	m.patterns = append(m.patterns, eki.SemanticPatterns[start:]...)
	for i := len(eki.EpisodicMemories) - 1; i >= 0 && len(m.memories) < 8; i-- {
		if c := strings.TrimSpace(eki.EpisodicMemories[i].Content); c != "" {
			m.memories = append(m.memories, echoTruncateRunes(c, 200))
		}
	}
	return m
}

func (m dreamMaterialSnapshot) describe() string {
	var sb strings.Builder
	if len(m.memories) > 0 {
		sb.WriteString("Recent experiences:\n")
		for _, mem := range m.memories {
			fmt.Fprintf(&sb, "- %s\n", mem)
		}
	}
	if len(m.patterns) > 0 {
		sb.WriteString("Consolidated patterns:\n")
		for _, p := range m.patterns {
			fmt.Fprintf(&sb, "- %s (strength %.2f)\n", echoTruncateRunes(p.Pattern, 200), p.Strength)
		}
	}
	return sb.String()
}

func echoTruncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// publishSystem reports a non-character system line (e.g. a model failure).
func (eki *EchodreamKnowledgeIntegrator) publishSystem(content string) {
	if eki.eventBus != nil {
		eki.eventBus.Publish(CogEvent{
			Category: CogEventSystem,
			Source:   "echodream",
			Content:  content,
			Priority: 0.6,
		})
	}
}

// consolidateMemories groups similar episodic memories into semantic patterns
func (eki *EchodreamKnowledgeIntegrator) consolidateMemories() int {
	count := 0

	unconsolidated := make([]*EpisodicMemoryV2, 0)
	for _, mem := range eki.EpisodicMemories {
		if !mem.Consolidated {
			unconsolidated = append(unconsolidated, mem)
		}
	}

	if len(unconsolidated) < 2 {
		return 0
	}

	// Group by source and emotion for simple pattern extraction
	groups := make(map[string][]*EpisodicMemoryV2)
	for _, mem := range unconsolidated {
		key := fmt.Sprintf("%s:%s", mem.Source, mem.Emotion)
		groups[key] = append(groups[key], mem)
	}

	for key, group := range groups {
		if len(group) < 2 {
			continue
		}

		// Create semantic pattern from group
		sourceIDs := make([]string, len(group))
		totalImportance := 0.0
		for i, mem := range group {
			sourceIDs[i] = mem.ID
			totalImportance += mem.Importance
			mem.Consolidated = true
			count++
		}

		pattern := &SemanticPattern{
			ID:        fmt.Sprintf("pat_%d_%s", time.Now().UnixNano(), key),
			Pattern:   fmt.Sprintf("Recurring pattern from %s: %d experiences consolidated", key, len(group)),
			Sources:   sourceIDs,
			Strength:  math.Min(1.0, totalImportance/float64(len(group))),
			CreatedAt: time.Now(),
		}

		eki.SemanticPatterns = append(eki.SemanticPatterns, pattern)
		eki.ConsolidationCount++
	}

	return count
}

// synthesizePatterns creates new patterns by combining existing ones
func (eki *EchodreamKnowledgeIntegrator) synthesizePatterns() int {
	if len(eki.SemanticPatterns) < 2 {
		return 0
	}

	count := 0

	// Random cross-pollination of patterns (dream-like association)
	for i := range 3 {
		idx1 := rand.Intn(len(eki.SemanticPatterns))
		idx2 := rand.Intn(len(eki.SemanticPatterns))
		if idx1 == idx2 {
			continue
		}

		p1 := eki.SemanticPatterns[idx1]
		p2 := eki.SemanticPatterns[idx2]

		// Combine patterns
		newPattern := &SemanticPattern{
			ID:        fmt.Sprintf("syn_%d", time.Now().UnixNano()+int64(i)),
			Pattern:   fmt.Sprintf("Synthesis of [%s] and [%s]", p1.Pattern, p2.Pattern),
			Sources:   append(p1.Sources, p2.Sources...),
			Strength:  (p1.Strength + p2.Strength) / 2.0,
			CreatedAt: time.Now(),
		}

		eki.SemanticPatterns = append(eki.SemanticPatterns, newPattern)
		count++
	}

	return count
}

// generateInsights asks the model for an emergent insight from pattern
// interference. Without a model (or on model failure) no insight is produced.
func (eki *EchodreamKnowledgeIntegrator) generateInsights(ctx context.Context, gen EchoTextGenerator, material dreamMaterialSnapshot) []*EmergentInsightV2 {
	if len(material.patterns) < 3 {
		return nil
	}

	prompt := material.describe() +
		"\nYou are dreaming. Let these experiences and patterns interfere with each other and state one " +
		"emergent insight they reveal to you. Reply with the insight only, in one or two sentences."
	text, err := echoGenerate(ctx, gen, echoPersonaSystemPrompt, prompt, 120)
	if err != nil {
		eki.publishSystem(fmt.Sprintf("[system] dream insight not generated: %v", err))
		return nil
	}

	sources := make([]string, 0, len(material.patterns))
	for _, p := range material.patterns {
		sources = append(sources, p.ID)
	}
	return []*EmergentInsightV2{{
		ID:        fmt.Sprintf("insight_%d", time.Now().UnixNano()),
		Insight:   text,
		Sources:   sources,
		Novelty:   0.5 + rand.Float64()*0.5,
		Depth:     0.5 + rand.Float64()*0.5,
		CreatedAt: time.Now(),
	}}
}

// generateDreamNarrative asks the model to narrate the dream cycle from what
// actually happened in it. It returns "" when the model is unavailable.
func (eki *EchodreamKnowledgeIntegrator) generateDreamNarrative(ctx context.Context, gen EchoTextGenerator, material dreamMaterialSnapshot, entry *DreamEntry, insights []*EmergentInsightV2) string {
	var sb strings.Builder
	sb.WriteString(material.describe())
	fmt.Fprintf(&sb, "\nDream cycle %d: %d memories consolidated, %d new patterns synthesized.\n",
		entry.CycleNumber, entry.MemoriesProcessed, entry.PatternsFormed)
	for _, insight := range insights {
		fmt.Fprintf(&sb, "Insight that emerged: %s\n", insight.Insight)
	}
	sb.WriteString("\nDescribe the dream you just had, in the first person, as a short vivid narrative " +
		"(at most four sentences) grounded in the material above.")

	text, err := echoGenerate(ctx, gen, echoPersonaSystemPrompt, sb.String(), 220)
	if err != nil {
		eki.publishSystem(fmt.Sprintf("[system] dream narrative not generated: %v", err))
		return ""
	}
	return text
}

// GetStatus returns the current state of the dream system
func (eki *EchodreamKnowledgeIntegrator) GetStatus() map[string]interface{} {
	eki.mu.RLock()
	defer eki.mu.RUnlock()

	return map[string]interface{}{
		"episodic_memories":  len(eki.EpisodicMemories),
		"semantic_patterns":  len(eki.SemanticPatterns),
		"emergent_insights":  len(eki.EmergentInsights),
		"dream_cycles":       eki.DreamCycleCount,
		"consolidations":     eki.ConsolidationCount,
		"insights_generated": eki.InsightCount,
		"last_dream":         eki.LastDreamTime.Format(time.RFC3339),
	}
}
