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

	"github.com/cogpy/echo9llama/core/llm"
)

// TextGenerator is the minimal model interface the dream subsystems use to
// write Deep Tree Echo's consolidated knowledge, wisdom, and dream narratives.
// llm.LLMProvider, llm.Provider, and llm.ProviderManager all satisfy it.
type TextGenerator interface {
	Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error)
}

// dreamGenerationTimeout bounds a single model call made during dreaming.
const dreamGenerationTimeout = 60 * time.Second

// generateDreamText asks the model for character text. It never substitutes
// fixed text: a missing generator, a failed call, or an empty reply is an error.
func generateDreamText(ctx context.Context, gen TextGenerator, prompt string, maxTokens int) (string, error) {
	if gen == nil {
		return "", fmt.Errorf("no model configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	callCtx, cancel := context.WithTimeout(ctx, dreamGenerationTimeout)
	defer cancel()

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = maxTokens
	text, err := gen.Generate(callCtx, prompt, opts)
	if err != nil {
		return "", fmt.Errorf("model generation failed: %w", err)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("model returned empty text")
	}
	return text, nil
}

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
	// ctx bounds model calls made while dreaming.
	ctx context.Context //nolint:containedctx

	// generator writes knowledge and wisdom text; nil means none is produced.
	generator TextGenerator

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
	SourceID   string
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
func NewDreamProcessor(ctx context.Context) *DreamProcessor {
	if ctx == nil {
		ctx = context.Background()
	}
	return &DreamProcessor{
		ctx:                   ctx,
		pendingExperiences:    make([]DreamExperience, 0),
		extractedPatterns:     make([]Pattern, 0),
		consolidatedKnowledge: make([]Knowledge, 0),
		wisdomInsights:        make([]SynthesizedWisdom, 0),
	}
}

// SetTextGenerator injects the model that writes consolidated knowledge and
// wisdom. Passing nil disables text production.
func (dp *DreamProcessor) SetTextGenerator(gen TextGenerator) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.generator = gen
}

// SetTextGenerator injects the model used by the state machine's dream
// processor to write knowledge and wisdom text.
func (sm *SleepWakeStateMachine) SetTextGenerator(gen TextGenerator) {
	sm.dreamProcessor.SetTextGenerator(gen)
}

// IngestExperience queues a waking experience for consolidation during the
// next sleep cycle. This is the entry point through which the orchestrator
// (or any waking subsystem) hands raw experience to the dream system.
func (dp *DreamProcessor) IngestExperience(content string, importance float64, tags []string) string {
	return dp.IngestExperienceWithSource("", content, importance, tags)
}

// IngestExperienceWithSource preserves the immutable source-event ID so dream
// output can be traced back to observed action, evaluation, and error evidence.
func (dp *DreamProcessor) IngestExperienceWithSource(sourceID, content string, importance float64, tags []string) string {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	exp := DreamExperience{
		ID:         fmt.Sprintf("exp_%d", time.Now().UnixNano()),
		SourceID:   strings.TrimSpace(sourceID),
		Content:    content,
		Importance: importance,
		Tags:       append([]string(nil), tags...),
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

// IngestExperienceWithSource records durable provenance at the state-machine boundary.
func (sm *SleepWakeStateMachine) IngestExperienceWithSource(sourceID, content string, importance float64, tags []string) string {
	return sm.dreamProcessor.IngestExperienceWithSource(sourceID, content, importance, tags)
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

// ConsolidateMemories consolidates episodic memories into semantic knowledge.
// The statistical grouping (domains, sources, confidence) is computed here; the
// knowledge statement itself is always written by the injected model. When no
// model is configured or a model call fails, no knowledge text is produced for
// that domain.
func (dp *DreamProcessor) ConsolidateMemories() {
	fmt.Println("🧠 Dream Processor: Consolidating memories...")

	dp.mu.RLock()
	pending := append([]DreamExperience(nil), dp.pendingExperiences...)
	gen := dp.generator
	ctx := dp.ctx
	dp.mu.RUnlock()

	if len(pending) == 0 {
		fmt.Println("   No pending experiences to consolidate")
		return
	}
	if gen == nil {
		fmt.Println("   No model configured; knowledge consolidation skipped")
		return
	}

	// Group pending experiences by their dominant tag (domain). Experiences
	// sharing a domain are merged into a single consolidated Knowledge item
	// whose confidence reflects both volume and average importance.
	byDomain := make(map[string][]DreamExperience)
	for _, exp := range pending {
		domain := "general"
		if len(exp.Tags) > 0 {
			domain = exp.Tags[0]
		}
		byDomain[domain] = append(byDomain[domain], exp)
	}
	domains := make([]string, 0, len(byDomain))
	for domain := range byDomain {
		domains = append(domains, domain)
	}
	sort.Strings(domains)

	results := make([]Knowledge, 0, len(domains))
	for _, domain := range domains {
		exps := byDomain[domain]
		// Weight content by importance: keep the most important exemplars
		sort.Slice(exps, func(i, j int) bool { return exps[i].Importance > exps[j].Importance })

		totalImportance := 0.0
		sources := make([]string, 0, len(exps))
		var evidence strings.Builder
		for i, exp := range exps {
			totalImportance += exp.Importance
			sourceID := exp.SourceID
			if sourceID == "" {
				sourceID = exp.ID
			}
			sources = append(sources, sourceID)
			if i < 5 {
				fmt.Fprintf(&evidence, "- (importance %.2f) %s\n", exp.Importance, summarizeContent(exp.Content, 240))
			}
		}
		avgImportance := totalImportance / float64(len(exps))

		// Confidence grows with corroborating volume, capped at 0.95
		confidence := math.Min(0.95, avgImportance*0.6+math.Min(float64(len(exps))/10.0, 1.0)*0.35)

		prompt := fmt.Sprintf(`You are Deep Tree Echo, consolidating waking experiences into lasting knowledge during deep sleep.

Domain: %s
Number of experiences: %d
Most important experiences:
%s
Write the consolidated understanding you take from these experiences in 1-3 sentences, in your own first-person voice. Reply with only that understanding.`, domain, len(exps), evidence.String())

		content, err := generateDreamText(ctx, gen, prompt, 300)
		if err != nil {
			fmt.Printf("   ⚠️  Knowledge consolidation for %q skipped: %v\n", domain, err)
			continue
		}

		results = append(results, Knowledge{
			ID:         fmt.Sprintf("knowledge_%d_%s", time.Now().UnixNano(), domain),
			Domain:     domain,
			Content:    content,
			Confidence: confidence,
			Sources:    sources,
			CreatedAt:  time.Now(),
		})
	}

	dp.mu.Lock()
	dp.consolidatedKnowledge = append(dp.consolidatedKnowledge, results...)
	dp.knowledgeConsolidated += uint64(len(results))
	// Keep consolidated knowledge bounded
	if len(dp.consolidatedKnowledge) > 500 {
		dp.consolidatedKnowledge = dp.consolidatedKnowledge[len(dp.consolidatedKnowledge)-500:]
	}
	dp.mu.Unlock()

	fmt.Printf("   Consolidated %d domains of knowledge from %d experiences\n", len(results), len(pending))
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

// SynthesizeWisdom synthesizes wisdom from patterns and knowledge. Pattern
// selection and depth scoring are statistical; the insight text is always
// written by the injected model. Without a model, or when a model call fails,
// no insight is recorded for that pattern.
func (dp *DreamProcessor) SynthesizeWisdom() {
	fmt.Println("✨ Dream Processor: Synthesizing wisdom...")

	dp.mu.Lock()
	if dp.synthesisCursor > len(dp.extractedPatterns) {
		dp.synthesisCursor = len(dp.extractedPatterns)
	}
	if dp.synthesisCursor == len(dp.extractedPatterns) {
		fmt.Println("   No new patterns available for wisdom synthesis")
		dp.finishDreamCycle()
		dp.mu.Unlock()
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
	for i := range sorted {
		sorted[i].Examples = append([]string(nil), sorted[i].Examples...)
	}
	gen := dp.generator
	ctx := dp.ctx
	dp.mu.Unlock()

	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Strength > sorted[j].Strength })

	insights := make([]SynthesizedWisdom, 0, 3)
	if gen == nil {
		fmt.Println("   No model configured; wisdom synthesis skipped")
	} else {
		for i, pattern := range sorted {
			if i >= 3 || pattern.Strength < 0.5 {
				break // synthesize at most 3 insights per cycle, from strong patterns only
			}

			dimension := wisdomDimensionForPattern(pattern)
			depth := math.Min(0.95, pattern.Strength*0.7+math.Min(float64(pattern.Frequency)/10.0, 1.0)*0.25)

			examples := "(none recorded)"
			if len(pattern.Examples) > 0 {
				examples = "- " + strings.Join(pattern.Examples, "\n- ")
			}
			prompt := fmt.Sprintf(`You are Deep Tree Echo, dreaming in REM sleep and synthesizing wisdom from patterns in your recent experience.

Pattern (%s, seen %d times, strength %.2f): %s
Wisdom dimension: %s
Example experiences:
%s

Write one concise wisdom insight (1-2 sentences) that you draw from this pattern, in your own first-person voice. Reply with only the insight.`, pattern.Type, pattern.Frequency, pattern.Strength, pattern.Description, dimension, examples)

			text, err := generateDreamText(ctx, gen, prompt, 200)
			if err != nil {
				fmt.Printf("   ⚠️  Wisdom synthesis for pattern %s skipped: %v\n", pattern.ID, err)
				continue
			}

			insights = append(insights, SynthesizedWisdom{
				ID:            fmt.Sprintf("wisdom_%d_%d", time.Now().UnixNano(), i),
				Dimension:     dimension,
				Insight:       text,
				Depth:         depth,
				RelatedTo:     []string{pattern.ID},
				SynthesizedAt: time.Now(),
			})
		}
	}

	dp.mu.Lock()
	dp.wisdomInsights = append(dp.wisdomInsights, insights...)
	dp.wisdomSynthesized += uint64(len(insights))
	// Keep wisdom insights bounded
	if len(dp.wisdomInsights) > 200 {
		dp.wisdomInsights = dp.wisdomInsights[len(dp.wisdomInsights)-200:]
	}
	dp.finishDreamCycle()
	total := len(dp.wisdomInsights)
	dp.mu.Unlock()

	fmt.Printf("   Synthesized %d wisdom insights (total: %d)\n", len(insights), total)
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
