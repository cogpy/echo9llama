// Package echodream implements knowledge integration during rest cycles
// This is where Deep Tree Echo consolidates memories, extracts patterns,
// and synthesizes wisdom during sleep and dream states

package echodream

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// SleepCyclePhase represents the phase of sleep cycle
type SleepCyclePhase int

const (
	PhaseLight    SleepCyclePhase = iota // Light sleep - transition phase
	PhaseDeep                            // Deep sleep - memory consolidation
	PhaseREMSleep                        // REM sleep - pattern extraction and dreaming
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
	mu sync.RWMutex

	// Ingested waking experiences awaiting consolidation
	pendingExperiences []DreamExperience

	// Pattern extraction
	extractedPatterns     []Pattern
	consolidatedKnowledge []Knowledge

	// Wisdom synthesis
	wisdomInsights  []SynthesizedWisdom
	synthesisCursor int // index of the first pattern not yet considered for wisdom

	// Metrics
	totalDreamCycles      uint64
	patternsExtracted     uint64
	knowledgeConsolidated uint64
	wisdomSynthesized     uint64
}

// DreamExperience is a waking experience queued for dream consolidation
type DreamExperience struct {
	ID         string
	Content    string
	Importance float64
	Tags       []string
	RecordedAt time.Time
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
	ID         string
	Domain     string
	Content    string
	Confidence float64
	Sources    []string
	CreatedAt  time.Time
}

// SynthesizedWisdom represents synthesized wisdom from dreams
type SynthesizedWisdom struct {
	ID            string
	Dimension     string // One of the 7 wisdom dimensions
	Insight       string
	Depth         float64
	RelatedTo     []string
	SynthesizedAt time.Time
}

// SleepWakeStateMachine manages sleep/wake transitions and dream processing
type SleepWakeStateMachine struct {
	mu sync.RWMutex
	// ctx is the owned lifecycle context used to interrupt every dream phase.
	ctx    context.Context //nolint:containedctx
	cancel context.CancelFunc

	// Current state
	isAsleep       bool
	currentPhase   SleepCyclePhase
	sleepStartTime time.Time

	// Configuration
	lightSleepDuration time.Duration
	deepSleepDuration  time.Duration
	remSleepDuration   time.Duration

	// Dream processor
	dreamProcessor *DreamProcessor

	// Metrics
	totalSleepCycles uint64
	totalSleepTime   time.Duration
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

// ConfigurePhaseDurations updates the light, deep, and REM dwell times.
// Non-positive values leave the corresponding defaults unchanged.
func (sm *SleepWakeStateMachine) ConfigurePhaseDurations(light, deep, rem time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if light > 0 {
		sm.lightSleepDuration = light
	}
	if deep > 0 {
		sm.deepSleepDuration = deep
	}
	if rem > 0 {
		sm.remSleepDuration = rem
	}
}

// NewDreamProcessor creates a new dream processor.
func NewDreamProcessor(_ context.Context) *DreamProcessor {
	return &DreamProcessor{
		pendingExperiences:    make([]DreamExperience, 0),
		extractedPatterns:     make([]Pattern, 0),
		consolidatedKnowledge: make([]Knowledge, 0),
		wisdomInsights:        make([]SynthesizedWisdom, 0),
	}
}

// IngestExperience queues a waking experience for consolidation during the
// next sleep cycle. This is the entry point through which the orchestrator
// (or any waking subsystem) hands raw experience to the dream system.
func (dp *DreamProcessor) IngestExperience(content string, importance float64, tags []string) string {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	exp := DreamExperience{
		ID:         fmt.Sprintf("exp_%d", time.Now().UnixNano()),
		Content:    content,
		Importance: importance,
		Tags:       tags,
		RecordedAt: time.Now(),
	}
	dp.pendingExperiences = append(dp.pendingExperiences, exp)

	// Keep the pending buffer bounded, dropping the least recent first
	if len(dp.pendingExperiences) > 1000 {
		dp.pendingExperiences = dp.pendingExperiences[len(dp.pendingExperiences)-1000:]
	}
	return exp.ID
}

// IngestExperienceForProcessor exposes ingestion at the state machine level
func (sm *SleepWakeStateMachine) IngestExperience(content string, importance float64, tags []string) string {
	return sm.dreamProcessor.IngestExperience(content, importance, tags)
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

	if len(dp.pendingExperiences) == 0 {
		fmt.Println("   No pending experiences to consolidate")
		return
	}

	// Group pending experiences by their dominant tag (domain). Experiences
	// sharing a domain are merged into a single consolidated Knowledge item
	// whose confidence reflects both volume and average importance.
	byDomain := make(map[string][]DreamExperience)
	for _, exp := range dp.pendingExperiences {
		domain := "general"
		if len(exp.Tags) > 0 {
			domain = exp.Tags[0]
		}
		byDomain[domain] = append(byDomain[domain], exp)
	}

	consolidated := 0
	for domain, exps := range byDomain {
		// Weight content by importance: keep the most important exemplars
		sort.Slice(exps, func(i, j int) bool { return exps[i].Importance > exps[j].Importance })

		totalImportance := 0.0
		sources := make([]string, 0, len(exps))
		exemplars := make([]string, 0, 3)
		for i, exp := range exps {
			totalImportance += exp.Importance
			sources = append(sources, exp.ID)
			if i < 3 {
				exemplars = append(exemplars, summarizeContent(exp.Content, 100))
			}
		}
		avgImportance := totalImportance / float64(len(exps))

		// Confidence grows with corroborating volume, capped at 0.95
		confidence := math.Min(0.95, avgImportance*0.6+math.Min(float64(len(exps))/10.0, 1.0)*0.35)

		knowledge := Knowledge{
			ID:         fmt.Sprintf("knowledge_%d_%s", time.Now().UnixNano(), domain),
			Domain:     domain,
			Content:    fmt.Sprintf("Consolidated understanding of %s from %d experiences: %s", domain, len(exps), strings.Join(exemplars, " | ")),
			Confidence: confidence,
			Sources:    sources,
			CreatedAt:  time.Now(),
		}
		dp.consolidatedKnowledge = append(dp.consolidatedKnowledge, knowledge)
		dp.knowledgeConsolidated++
		consolidated++
	}

	// Keep consolidated knowledge bounded
	if len(dp.consolidatedKnowledge) > 500 {
		dp.consolidatedKnowledge = dp.consolidatedKnowledge[len(dp.consolidatedKnowledge)-500:]
	}

	fmt.Printf("   Consolidated %d domains of knowledge from %d experiences\n", consolidated, len(dp.pendingExperiences))
}

// ExtractPatterns extracts patterns from episodic memories
func (dp *DreamProcessor) ExtractPatterns() {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	fmt.Println("🔍 Dream Processor: Extracting patterns...")

	if len(dp.pendingExperiences) == 0 {
		fmt.Println("   No experiences available for pattern extraction")
		return
	}

	// Count tag co-occurrence frequencies across pending experiences. Tags
	// recurring above threshold become conceptual patterns; recurring tag
	// pairs become relational patterns.
	tagFreq := make(map[string]int)
	tagImportance := make(map[string]float64)
	tagExamples := make(map[string][]string)
	pairFreq := make(map[string]int)

	for _, exp := range dp.pendingExperiences {
		for i, tag := range exp.Tags {
			tagFreq[tag]++
			tagImportance[tag] += exp.Importance
			if len(tagExamples[tag]) < 3 {
				tagExamples[tag] = append(tagExamples[tag], summarizeContent(exp.Content, 60))
			}
			for _, other := range exp.Tags[i+1:] {
				pair := tag + "+" + other
				if other < tag {
					pair = other + "+" + tag
				}
				pairFreq[pair]++
			}
		}
	}

	extracted := 0
	for tag, freq := range tagFreq {
		if freq < 2 {
			continue // require recurrence before it counts as a pattern
		}
		avgImp := tagImportance[tag] / float64(freq)
		strength := math.Min(0.95, avgImp*0.5+math.Min(float64(freq)/10.0, 1.0)*0.45)

		dp.extractedPatterns = append(dp.extractedPatterns, Pattern{
			ID:          fmt.Sprintf("pattern_%d_%s", time.Now().UnixNano(), tag),
			Type:        "conceptual",
			Description: fmt.Sprintf("Recurring theme '%s' across %d experiences", tag, freq),
			Frequency:   freq,
			Strength:    strength,
			Examples:    tagExamples[tag],
			ExtractedAt: time.Now(),
		})
		dp.patternsExtracted++
		extracted++
	}

	for pair, freq := range pairFreq {
		if freq < 2 {
			continue
		}
		dp.extractedPatterns = append(dp.extractedPatterns, Pattern{
			ID:          fmt.Sprintf("pattern_%d_%s", time.Now().UnixNano(), pair),
			Type:        "relational",
			Description: fmt.Sprintf("Concepts '%s' co-occur in %d experiences", pair, freq),
			Frequency:   freq,
			Strength:    math.Min(0.9, 0.4+float64(freq)*0.1),
			Examples:    []string{},
			ExtractedAt: time.Now(),
		})
		dp.patternsExtracted++
		extracted++
	}

	// Keep extracted patterns bounded and preserve the synthesis cursor relative
	// to the retained window.
	if len(dp.extractedPatterns) > 500 {
		removed := len(dp.extractedPatterns) - 500
		dp.extractedPatterns = dp.extractedPatterns[removed:]
		dp.synthesisCursor -= removed
		if dp.synthesisCursor < 0 {
			dp.synthesisCursor = 0
		}
	}

	fmt.Printf("   Extracted %d patterns (total: %d)\n", extracted, len(dp.extractedPatterns))
}

// SynthesizeWisdom synthesizes wisdom from patterns and knowledge
func (dp *DreamProcessor) SynthesizeWisdom() {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	fmt.Println("✨ Dream Processor: Synthesizing wisdom...")

	if dp.synthesisCursor > len(dp.extractedPatterns) {
		dp.synthesisCursor = len(dp.extractedPatterns)
	}
	if dp.synthesisCursor == len(dp.extractedPatterns) {
		fmt.Println("   No new patterns available for wisdom synthesis")
		dp.finishDreamCycle()
		return
	}

	// Synthesize only from patterns extracted since the previous dream cycle.
	// Historical patterns remain available for memory and inspection but cannot
	// inflate wisdom metrics by being emitted again under fresh IDs.
	recent := dp.extractedPatterns[dp.synthesisCursor:]
	dp.synthesisCursor = len(dp.extractedPatterns)
	if len(recent) > 20 {
		recent = recent[len(recent)-20:]
	}

	sorted := make([]Pattern, len(recent))
	copy(sorted, recent)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Strength > sorted[j].Strength })

	synthesized := 0
	for i, pattern := range sorted {
		if i >= 3 || pattern.Strength < 0.5 {
			break // synthesize at most 3 insights per cycle, from strong patterns only
		}

		dimension := wisdomDimensionForPattern(pattern)
		depth := math.Min(0.95, pattern.Strength*0.7+math.Min(float64(pattern.Frequency)/10.0, 1.0)*0.25)

		insight := SynthesizedWisdom{
			ID:            fmt.Sprintf("wisdom_%d_%d", time.Now().UnixNano(), i),
			Dimension:     dimension,
			Insight:       fmt.Sprintf("%s — this recurring structure suggests deeper significance worth attending to", pattern.Description),
			Depth:         depth,
			RelatedTo:     []string{pattern.ID},
			SynthesizedAt: time.Now(),
		}
		dp.wisdomInsights = append(dp.wisdomInsights, insight)
		dp.wisdomSynthesized++
		synthesized++
	}

	// Keep wisdom insights bounded
	if len(dp.wisdomInsights) > 200 {
		dp.wisdomInsights = dp.wisdomInsights[len(dp.wisdomInsights)-200:]
	}

	dp.finishDreamCycle()

	fmt.Printf("   Synthesized %d wisdom insights (total: %d)\n", synthesized, len(dp.wisdomInsights))
}

// finishDreamCycle clears consumed experiences and increments cycle count.
// Must be called with dp.mu held.
func (dp *DreamProcessor) finishDreamCycle() {
	dp.pendingExperiences = dp.pendingExperiences[:0]
	dp.totalDreamCycles++
}

// wisdomDimensionForPattern maps a pattern type to a wisdom dimension
func wisdomDimensionForPattern(p Pattern) string {
	switch p.Type {
	case "relational":
		return "Integrative Understanding"
	case "behavioral":
		return "Self-Reflection"
	case "temporal":
		return "Temporal Perspective"
	default:
		return "Conceptual Insight"
	}
}

// summarizeContent truncates content to maxLen runes for compact summaries
func summarizeContent(content string, maxLen int) string {
	runes := []rune(strings.TrimSpace(content))
	if len(runes) <= maxLen {
		return string(runes)
	}
	return string(runes[:maxLen]) + "…"
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
	isAsleep := sm.isAsleep
	phase := sm.currentPhase.String()
	totalCycles := sm.totalSleepCycles
	totalSleepTime := sm.totalSleepTime
	sm.mu.RUnlock()

	dp := sm.dreamProcessor
	dp.mu.RLock()
	pending := len(dp.pendingExperiences)
	patterns := dp.patternsExtracted
	knowledge := dp.knowledgeConsolidated
	wisdom := dp.wisdomSynthesized
	dp.mu.RUnlock()

	return map[string]interface{}{
		"is_asleep":              isAsleep,
		"current_phase":          phase,
		"total_sleep_cycles":     totalCycles,
		"total_sleep_time":       totalSleepTime.String(),
		"pending_experiences":    pending,
		"patterns_extracted":     patterns,
		"knowledge_consolidated": knowledge,
		"wisdom_synthesized":     wisdom,
	}
}

// GetWisdomInsights returns all synthesized wisdom insights
func (sm *SleepWakeStateMachine) GetWisdomInsights() []SynthesizedWisdom {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	dp := sm.dreamProcessor
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	insights := make([]SynthesizedWisdom, len(dp.wisdomInsights))
	copy(insights, dp.wisdomInsights)
	for i := range insights {
		insights[i].RelatedTo = append([]string(nil), insights[i].RelatedTo...)
	}
	return insights
}

// GetExtractedPatterns returns all extracted patterns
func (sm *SleepWakeStateMachine) GetExtractedPatterns() []Pattern {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	dp := sm.dreamProcessor
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	patterns := make([]Pattern, len(dp.extractedPatterns))
	copy(patterns, dp.extractedPatterns)
	for i := range patterns {
		patterns[i].Examples = append([]string(nil), patterns[i].Examples...)
	}
	return patterns
}

// GetConsolidatedKnowledge returns all consolidated knowledge
func (sm *SleepWakeStateMachine) GetConsolidatedKnowledge() []Knowledge {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	dp := sm.dreamProcessor
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	knowledge := make([]Knowledge, len(dp.consolidatedKnowledge))
	copy(knowledge, dp.consolidatedKnowledge)
	for i := range knowledge {
		knowledge[i].Sources = append([]string(nil), knowledge[i].Sources...)
	}
	return knowledge
}

// Shutdown gracefully shuts down the state machine
func (sm *SleepWakeStateMachine) Shutdown() {
	sm.cancel()
}
