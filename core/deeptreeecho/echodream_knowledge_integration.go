package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// EchoDreamKnowledgeIntegration handles knowledge consolidation during dream state
type EchoDreamKnowledgeIntegration struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc

	// LLM provider for knowledge processing
	llmProvider     llm.LLMProvider

	// Knowledge structures
	episodicMemories    []EpisodicMemory
	consolidatedPatterns []Pattern
	wisdomInsights      []WisdomInsight

	// Semantic memory network
	semanticNetwork     map[string]*SemanticNode
	patternLinks        []PatternLink

	// Wisdom depth tracking
	wisdomDepth         float64
	wisdomGrowthRate    float64
	maxWisdomDepth      float64

	// Dream cycle state
	dreamCycleActive    bool
	dreamIntensity      float64
	dreamPhase          DreamPhase

	// Consolidation state
	lastConsolidation   time.Time
	consolidationCount  uint64

	// Cross-pattern emergence
	emergentConcepts    []EmergentConcept

	// Metrics
	totalMemoriesProcessed uint64
	totalPatternsExtracted uint64
	totalWisdomGenerated   uint64
	semanticNodesCreated   uint64
	emergenceEvents        uint64

	// Running state
	running         bool
}

// EpisodicMemory represents a memory of an experience
type EpisodicMemory struct {
	ID          string
	Content     string
	Timestamp   time.Time
	Emotional   float64
	Importance  float64
	Tags        []string
	Consolidated bool
}

// Pattern represents an extracted pattern from experiences and cognitive processes
type Pattern struct {
	ID          string
	Description string
	Frequency   int
	Strength    float64
	Examples    []string
	CreatedAt   time.Time
	// Identity-related fields for cognitive pattern tracking
	Type        string    // Pattern type (cognitive, semantic, temporal, activation_cluster)
	Nodes       []string  // Connected nodes in hypergraph
	FirstSeen   time.Time // When pattern was first observed
	LastSeen    time.Time // When pattern was last active
	Occurrences int       // Number of times pattern has been activated
}

// WisdomInsight represents wisdom extracted from patterns
type WisdomInsight struct {
	ID            string
	Insight       string
	Source        []string  // Pattern IDs
	Depth         float64
	Applicability float64
	CreatedAt     time.Time
}

// SemanticNode represents a node in the semantic memory network
type SemanticNode struct {
	ID          string
	Concept     string
	Activation  float64
	Connections []string  // IDs of connected nodes
	Strength    float64
	CreatedAt   time.Time
	LastAccess  time.Time
	AccessCount int
}

// PatternLink represents a link between patterns
type PatternLink struct {
	ID          string
	FromPattern string
	ToPattern   string
	LinkType    string   // "causal", "temporal", "semantic", "contrast"
	Strength    float64
	CreatedAt   time.Time
}

// DreamPhase represents the phases of a dream cycle
type DreamPhase int

const (
	PhaseREM DreamPhase = iota    // Active pattern processing
	PhaseNREM1                     // Light consolidation
	PhaseNREM2                     // Intermediate consolidation
	PhaseNREM3                     // Deep consolidation
	PhaseWaking                    // Transition to wakefulness
)

func (dp DreamPhase) String() string {
	return [...]string{"REM", "NREM1", "NREM2", "NREM3", "Waking"}[dp]
}

// EmergentConcept represents a concept that emerged from pattern combination
type EmergentConcept struct {
	ID          string
	Concept     string
	SourcePatterns []string
	EmergenceStrength float64
	Novelty     float64
	Utility     float64
	CreatedAt   time.Time
}

// NewEchoDreamKnowledgeIntegration creates a new knowledge integration system
func NewEchoDreamKnowledgeIntegration(llmProvider llm.LLMProvider) *EchoDreamKnowledgeIntegration {
	ctx, cancel := context.WithCancel(context.Background())

	return &EchoDreamKnowledgeIntegration{
		ctx:                  ctx,
		cancel:               cancel,
		llmProvider:          llmProvider,
		episodicMemories:     make([]EpisodicMemory, 0),
		consolidatedPatterns: make([]Pattern, 0),
		wisdomInsights:       make([]WisdomInsight, 0),
		semanticNetwork:      make(map[string]*SemanticNode),
		patternLinks:         make([]PatternLink, 0),
		emergentConcepts:     make([]EmergentConcept, 0),
		wisdomDepth:          0.0,
		wisdomGrowthRate:     0.05,
		maxWisdomDepth:       1.0,
		dreamPhase:           PhaseWaking,
		dreamIntensity:       0.0,
	}
}

// ConsolidateKnowledge processes thoughts and experiences during dream state
func (edi *EchoDreamKnowledgeIntegration) ConsolidateKnowledge(ctx context.Context) error {
	edi.mu.Lock()
	defer edi.mu.Unlock()
	
	fmt.Println("🌙 Echodream: Beginning knowledge consolidation...")
	
	// Consolidate existing memories
	thoughtCount := len(edi.episodicMemories)
	edi.totalMemoriesProcessed += uint64(thoughtCount)
	
	// Extract patterns from recent memories
	if err := edi.extractPatterns(); err != nil {
		fmt.Printf("⚠️  Pattern extraction error: %v\n", err)
	}
	
	// Consolidate memories by importance
	edi.consolidateMemories()
	
	// Generate wisdom insights
	if err := edi.generateWisdomInsights(); err != nil {
		fmt.Printf("⚠️  Wisdom generation error: %v\n", err)
	}
	
	edi.lastConsolidation = time.Now()
	edi.consolidationCount++
	
	fmt.Printf("   ✓ Processed %d memories\n", thoughtCount)
	fmt.Printf("   ✓ Extracted %d patterns\n", len(edi.consolidatedPatterns))
	fmt.Printf("   ✓ Generated %d wisdom insights\n", len(edi.wisdomInsights))
	
	return nil
}

// extractPatterns identifies recurring patterns in memories
func (edi *EchoDreamKnowledgeIntegration) extractPatterns() error {
	// Collect recent unconsolidated memories
	recentMemories := make([]EpisodicMemory, 0)
	for _, mem := range edi.episodicMemories {
		if !mem.Consolidated && time.Since(mem.Timestamp) < 24*time.Hour {
			recentMemories = append(recentMemories, mem)
		}
	}
	
	if len(recentMemories) < 3 {
		return nil  // Need at least 3 memories to extract patterns
	}
	
	// Construct pattern extraction prompt
	memoryTexts := ""
	for i, mem := range recentMemories {
		if i < 10 {  // Limit to 10 most recent
			memoryTexts += fmt.Sprintf("- %s\n", mem.Content)
		}
	}
	
	prompt := fmt.Sprintf(`Analyze these recent experiences and identify recurring patterns or themes:

%s

Identify 1-3 key patterns. For each pattern, provide:
1. A brief description
2. Why it's significant

Be concise.`, memoryTexts)
	
	opts := llm.GenerateOptions{
		Temperature:  0.6,
		MaxTokens:    200,
	}
	
	fullPrompt := "[System: You are a pattern recognition system analyzing cognitive experiences.]\n\n" + prompt
	result, err := edi.llmProvider.Generate(context.Background(), fullPrompt, opts)
	if err != nil {
		return fmt.Errorf("pattern extraction failed: %w", err)
	}
	
	// Create pattern object (simplified - in production, parse the result)
	pattern := Pattern{
		ID:          fmt.Sprintf("pattern_%d", time.Now().UnixNano()),
		Description: result,
		Frequency:   len(recentMemories),
		Strength:    0.7,
		Examples:    make([]string, 0),
		CreatedAt:   time.Now(),
	}
	
	edi.consolidatedPatterns = append(edi.consolidatedPatterns, pattern)
	edi.totalPatternsExtracted++
	
	fmt.Printf("   🔍 Pattern Identified: %s\n", truncate(result, 70))
	
	return nil
}

// consolidateMemories prunes low-importance memories
func (edi *EchoDreamKnowledgeIntegration) consolidateMemories() {
	// Mark memories as consolidated
	consolidatedCount := 0
	for i := range edi.episodicMemories {
		if !edi.episodicMemories[i].Consolidated {
			edi.episodicMemories[i].Consolidated = true
			consolidatedCount++
		}
	}
	
	// Prune low-importance memories if we have too many
	if len(edi.episodicMemories) > 500 {
		// Keep only high-importance memories
		kept := make([]EpisodicMemory, 0)
		for _, mem := range edi.episodicMemories {
			if mem.Importance > 0.6 || time.Since(mem.Timestamp) < 24*time.Hour {
				kept = append(kept, mem)
			}
		}
		
		pruned := len(edi.episodicMemories) - len(kept)
		edi.episodicMemories = kept
		
		if pruned > 0 {
			fmt.Printf("   🗑️  Pruned %d low-importance memories\n", pruned)
		}
	}
	
	fmt.Printf("   📦 Consolidated %d memories\n", consolidatedCount)
}

// generateWisdomInsights extracts wisdom from patterns
func (edi *EchoDreamKnowledgeIntegration) generateWisdomInsights() error {
	if len(edi.consolidatedPatterns) < 2 {
		return nil  // Need at least 2 patterns to generate wisdom
	}
	
	// Take recent patterns
	recentPatterns := edi.consolidatedPatterns
	if len(recentPatterns) > 5 {
		recentPatterns = recentPatterns[len(recentPatterns)-5:]
	}
	
	patternTexts := ""
	patternIDs := make([]string, 0)
	for _, pattern := range recentPatterns {
		patternTexts += fmt.Sprintf("- %s\n", pattern.Description)
		patternIDs = append(patternIDs, pattern.ID)
	}
	
	prompt := fmt.Sprintf(`Reflect on these patterns from recent experiences:

%s

What wisdom or deeper understanding emerges from these patterns? 
What principle or insight can guide future growth?

Provide a concise wisdom insight:`, patternTexts)
	
	opts := llm.GenerateOptions{
		Temperature:  0.7,
		MaxTokens:    150,
	}
	
	fullPrompt := "[System: You are a wisdom extraction system. Generate deep, actionable insights.]\n\n" + prompt
	result, err := edi.llmProvider.Generate(context.Background(), fullPrompt, opts)
	if err != nil {
		return fmt.Errorf("wisdom generation failed: %w", err)
	}
	
	wisdom := WisdomInsight{
		ID:            fmt.Sprintf("wisdom_%d", time.Now().UnixNano()),
		Insight:       result,
		Source:        patternIDs,
		Depth:         0.7,
		Applicability: 0.8,
		CreatedAt:     time.Now(),
	}
	
	edi.wisdomInsights = append(edi.wisdomInsights, wisdom)
	edi.totalWisdomGenerated++
	
	fmt.Printf("   💎 Wisdom Insight: %s\n", truncate(result, 70))
	
	return nil
}

// ExtractWisdom returns the accumulated wisdom level
func (edi *EchoDreamKnowledgeIntegration) ExtractWisdom() float64 {
	edi.mu.RLock()
	defer edi.mu.RUnlock()
	
	if len(edi.wisdomInsights) == 0 {
		return 0.0
	}
	
	// Calculate average depth of recent wisdom insights
	totalDepth := 0.0
	count := 0
	for i := len(edi.wisdomInsights) - 1; i >= 0 && count < 5; i-- {
		totalDepth += edi.wisdomInsights[i].Depth
		count++
	}
	
	return totalDepth / float64(count)
}

// GetRecentWisdom returns recent wisdom insights
func (edi *EchoDreamKnowledgeIntegration) GetRecentWisdom(limit int) []WisdomInsight {
	edi.mu.RLock()
	defer edi.mu.RUnlock()
	
	if len(edi.wisdomInsights) == 0 {
		return []WisdomInsight{}
	}
	
	start := len(edi.wisdomInsights) - limit
	if start < 0 {
		start = 0
	}
	
	return edi.wisdomInsights[start:]
}

// GetPatterns returns all consolidated patterns
func (edi *EchoDreamKnowledgeIntegration) GetPatterns() []Pattern {
	edi.mu.RLock()
	defer edi.mu.RUnlock()
	
	return edi.consolidatedPatterns
}

// GetMetrics returns echodream metrics
func (edi *EchoDreamKnowledgeIntegration) GetMetrics() map[string]interface{} {
	edi.mu.RLock()
	defer edi.mu.RUnlock()

	return map[string]interface{}{
		"total_memories":         len(edi.episodicMemories),
		"total_patterns":         len(edi.consolidatedPatterns),
		"total_wisdom":           len(edi.wisdomInsights),
		"consolidation_count":    edi.consolidationCount,
		"last_consolidation":     edi.lastConsolidation.Format(time.RFC3339),
		"memories_processed":     edi.totalMemoriesProcessed,
		"patterns_extracted":     edi.totalPatternsExtracted,
		"wisdom_generated":       edi.totalWisdomGenerated,
		"semantic_nodes":         len(edi.semanticNetwork),
		"pattern_links":          len(edi.patternLinks),
		"emergent_concepts":      len(edi.emergentConcepts),
		"wisdom_depth":           edi.wisdomDepth,
		"dream_phase":            edi.dreamPhase.String(),
		"dream_intensity":        edi.dreamIntensity,
	}
}

// StartDreamCycle initiates an autonomous dream cycle
func (edi *EchoDreamKnowledgeIntegration) StartDreamCycle() error {
	edi.mu.Lock()
	defer edi.mu.Unlock()

	if edi.dreamCycleActive {
		return fmt.Errorf("dream cycle already active")
	}

	edi.dreamCycleActive = true
	edi.dreamPhase = PhaseNREM1
	edi.dreamIntensity = 0.3

	fmt.Println("🌙 Starting dream cycle...")

	go edi.runDreamCycle()

	return nil
}

// runDreamCycle executes the autonomous dream cycle
func (edi *EchoDreamKnowledgeIntegration) runDreamCycle() {
	phases := []DreamPhase{PhaseNREM1, PhaseNREM2, PhaseNREM3, PhaseREM}
	intensities := []float64{0.3, 0.5, 0.8, 1.0}

	for i, phase := range phases {
		select {
		case <-edi.ctx.Done():
			return
		default:
			edi.mu.Lock()
			edi.dreamPhase = phase
			edi.dreamIntensity = intensities[i]
			edi.mu.Unlock()

			fmt.Printf("   🌙 Dream phase: %s (intensity: %.1f)\n", phase.String(), intensities[i])

			// Process based on phase
			switch phase {
			case PhaseNREM1:
				edi.lightConsolidation()
			case PhaseNREM2:
				edi.intermediateConsolidation()
			case PhaseNREM3:
				edi.deepConsolidation()
			case PhaseREM:
				edi.activePatternProcessing()
			}

			time.Sleep(2 * time.Second)
		}
	}

	// Transition to waking
	edi.mu.Lock()
	edi.dreamPhase = PhaseWaking
	edi.dreamIntensity = 0.0
	edi.dreamCycleActive = false
	edi.mu.Unlock()

	fmt.Println("   🌅 Dream cycle complete - transitioning to wakefulness")
}

// lightConsolidation performs light memory consolidation
func (edi *EchoDreamKnowledgeIntegration) lightConsolidation() {
	edi.mu.Lock()
	defer edi.mu.Unlock()

	// Decay weak memories slightly
	for i := range edi.episodicMemories {
		if edi.episodicMemories[i].Importance < 0.3 {
			edi.episodicMemories[i].Importance *= 0.95
		}
	}
}

// intermediateConsolidation performs intermediate memory consolidation
func (edi *EchoDreamKnowledgeIntegration) intermediateConsolidation() {
	edi.mu.Lock()
	defer edi.mu.Unlock()

	// Strengthen moderate importance memories
	for i := range edi.episodicMemories {
		if edi.episodicMemories[i].Importance >= 0.5 && edi.episodicMemories[i].Importance < 0.8 {
			edi.episodicMemories[i].Importance *= 1.05
			if edi.episodicMemories[i].Importance > 1.0 {
				edi.episodicMemories[i].Importance = 1.0
			}
		}
	}
}

// deepConsolidation performs deep memory consolidation
func (edi *EchoDreamKnowledgeIntegration) deepConsolidation() {
	edi.mu.Lock()
	defer edi.mu.Unlock()

	// Build semantic connections for high-importance memories
	for _, mem := range edi.episodicMemories {
		if mem.Importance >= 0.7 && !mem.Consolidated {
			edi.createSemanticNode(mem)
		}
	}
}

// activePatternProcessing performs REM-like active pattern processing
func (edi *EchoDreamKnowledgeIntegration) activePatternProcessing() {
	edi.mu.Lock()
	defer edi.mu.Unlock()

	// Link patterns and detect emergence
	edi.linkPatterns()
	edi.detectEmergence()

	// Grow wisdom depth
	if len(edi.wisdomInsights) > 0 {
		edi.wisdomDepth += edi.wisdomGrowthRate
		if edi.wisdomDepth > edi.maxWisdomDepth {
			edi.wisdomDepth = edi.maxWisdomDepth
		}
	}
}

// createSemanticNode creates a semantic node from a memory
func (edi *EchoDreamKnowledgeIntegration) createSemanticNode(mem EpisodicMemory) {
	nodeID := fmt.Sprintf("sem_%s", mem.ID)

	if _, exists := edi.semanticNetwork[nodeID]; !exists {
		node := &SemanticNode{
			ID:          nodeID,
			Concept:     mem.Content,
			Activation:  mem.Importance,
			Connections: make([]string, 0),
			Strength:    mem.Importance,
			CreatedAt:   time.Now(),
			LastAccess:  time.Now(),
			AccessCount: 1,
		}
		edi.semanticNetwork[nodeID] = node
		edi.semanticNodesCreated++
	}
}

// linkPatterns creates links between related patterns
func (edi *EchoDreamKnowledgeIntegration) linkPatterns() {
	// Link patterns based on temporal proximity and strength
	for i := 0; i < len(edi.consolidatedPatterns)-1; i++ {
		for j := i + 1; j < len(edi.consolidatedPatterns); j++ {
			p1 := edi.consolidatedPatterns[i]
			p2 := edi.consolidatedPatterns[j]

			// Calculate similarity based on time proximity
			timeDiff := p2.CreatedAt.Sub(p1.CreatedAt)
			if timeDiff < 24*time.Hour {
				strength := (p1.Strength + p2.Strength) / 2.0 * (1.0 - float64(timeDiff.Hours())/24.0)

				if strength > 0.5 {
					link := PatternLink{
						ID:          fmt.Sprintf("link_%d", time.Now().UnixNano()),
						FromPattern: p1.ID,
						ToPattern:   p2.ID,
						LinkType:    "temporal",
						Strength:    strength,
						CreatedAt:   time.Now(),
					}
					edi.patternLinks = append(edi.patternLinks, link)
				}
			}
		}
	}
}

// detectEmergence identifies emergent concepts from pattern combinations
func (edi *EchoDreamKnowledgeIntegration) detectEmergence() {
	// Look for strongly linked patterns
	linkStrengths := make(map[string]float64)
	for _, link := range edi.patternLinks {
		if link.Strength > 0.7 {
			key := link.FromPattern + "_" + link.ToPattern
			linkStrengths[key] = link.Strength
		}
	}

	// Create emergent concepts from strong links
	for key, strength := range linkStrengths {
		// Check if we already have this concept
		exists := false
		for _, ec := range edi.emergentConcepts {
			if ec.ID == "em_"+key {
				exists = true
				break
			}
		}

		if !exists {
			emergent := EmergentConcept{
				ID:               "em_" + key,
				Concept:          fmt.Sprintf("Emergent concept from pattern combination"),
				SourcePatterns:   []string{key},
				EmergenceStrength: strength,
				Novelty:          0.7,
				Utility:          0.6,
				CreatedAt:        time.Now(),
			}
			edi.emergentConcepts = append(edi.emergentConcepts, emergent)
			edi.emergenceEvents++
			fmt.Printf("   ✨ Emergence detected: strength %.2f\n", strength)
		}
	}
}

// GetWisdomDepth returns the current wisdom depth
func (edi *EchoDreamKnowledgeIntegration) GetWisdomDepth() float64 {
	edi.mu.RLock()
	defer edi.mu.RUnlock()
	return edi.wisdomDepth
}

// GetSemanticNetwork returns the semantic network
func (edi *EchoDreamKnowledgeIntegration) GetSemanticNetwork() map[string]*SemanticNode {
	edi.mu.RLock()
	defer edi.mu.RUnlock()

	// Return a copy
	network := make(map[string]*SemanticNode)
	for k, v := range edi.semanticNetwork {
		network[k] = v
	}
	return network
}

// GetEmergentConcepts returns emergent concepts
func (edi *EchoDreamKnowledgeIntegration) GetEmergentConcepts() []EmergentConcept {
	edi.mu.RLock()
	defer edi.mu.RUnlock()

	concepts := make([]EmergentConcept, len(edi.emergentConcepts))
	copy(concepts, edi.emergentConcepts)
	return concepts
}

// GetDreamPhase returns the current dream phase
func (edi *EchoDreamKnowledgeIntegration) GetDreamPhase() DreamPhase {
	edi.mu.RLock()
	defer edi.mu.RUnlock()
	return edi.dreamPhase
}

// IsDreaming returns whether a dream cycle is active
func (edi *EchoDreamKnowledgeIntegration) IsDreaming() bool {
	edi.mu.RLock()
	defer edi.mu.RUnlock()
	return edi.dreamCycleActive
}

// AddMemory adds an episodic memory to the integration system
func (edi *EchoDreamKnowledgeIntegration) AddMemory(content string, importance float64, tags []string) string {
	edi.mu.Lock()
	defer edi.mu.Unlock()

	mem := EpisodicMemory{
		ID:          fmt.Sprintf("mem_%d", time.Now().UnixNano()),
		Content:     content,
		Timestamp:   time.Now(),
		Emotional:   0.5,
		Importance:  importance,
		Tags:        tags,
		Consolidated: false,
	}

	edi.episodicMemories = append(edi.episodicMemories, mem)
	return mem.ID
}

// BeginDreamCycle starts a dream cycle for knowledge consolidation
func (edi *EchoDreamKnowledgeIntegration) BeginDreamCycle() {
	edi.mu.Lock()
	defer edi.mu.Unlock()
	edi.dreamCycleActive = true
	edi.dreamPhase = PhaseREM
}

// EndDreamCycle ends the dream cycle and returns wisdom insights
func (edi *EchoDreamKnowledgeIntegration) EndDreamCycle() []WisdomInsight {
	edi.mu.Lock()
	defer edi.mu.Unlock()
	edi.dreamCycleActive = false
	edi.dreamPhase = PhaseWaking
	return edi.wisdomInsights
}
