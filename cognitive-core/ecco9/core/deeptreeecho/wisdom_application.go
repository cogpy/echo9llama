package deeptreeecho

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// WisdomApplicationEngine refines how wisdom is matched and applied to contexts
type WisdomApplicationEngine struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Wisdom repository
	wisdomBase       map[string]*WisdomEntry
	wisdomCategories map[string][]*WisdomEntry

	// Context-wisdom matching
	contextMatcher *ContextMatcher
	relevanceCache map[string]*RelevanceScore

	// Wisdom synthesis
	synthesizer       *WisdomSynthesizer
	synthesizedWisdom []*WisdomEntry

	// Application feedback
	applicationHistory []*WisdomApplication
	feedbackLoop       *WisdomFeedbackLoop

	// Metrics
	totalApplications      uint64
	successfulApplications uint64
	totalSyntheses         uint64

	// Model that writes applications of wisdom and syntheses; nil disables them.
	gen TextGenerator

	// Running state
	running bool
}

// WisdomEntry represents a piece of wisdom
type WisdomEntry struct {
	ID            string
	Content       string
	Category      string
	Source        string
	Applicability []string // Contexts where applicable
	Confidence    float64
	Usefulness    float64
	Generality    float64
	Specificity   float64
	Tags          []string
	RelatedWisdom []string
	CreatedAt     time.Time
	LastApplied   time.Time
	TimesApplied  uint64
	SuccessRate   float64
	Metadata      map[string]interface{}
}

// ContextMatcher matches contexts to relevant wisdom
type ContextMatcher struct {
	mu sync.RWMutex

	// Context analysis
	contextFeatures map[string][]string
	contextPatterns map[string]*ContextPattern

	// Matching strategies
	exactMatchWeight    float64
	semanticMatchWeight float64
	analogyMatchWeight  float64

	// Cache
	matchCache map[string][]*WisdomMatch
}

// ContextPattern represents recurring context patterns
type ContextPattern struct {
	ID               string
	Name             string
	Features         []string
	Frequency        uint64
	AssociatedWisdom []string
}

// WisdomMatch represents a context-wisdom match
type WisdomMatch struct {
	WisdomID       string
	Wisdom         *WisdomEntry
	RelevanceScore float64
	MatchType      MatchType
	Explanation    string
	Confidence     float64
}

// MatchType categorizes how wisdom matches context
type MatchType int

const (
	MatchExact MatchType = iota
	MatchSemantic
	MatchAnalogy
	MatchGeneralization
	MatchSpecialization
)

func (mt MatchType) String() string {
	return [...]string{
		"Exact",
		"Semantic",
		"Analogy",
		"Generalization",
		"Specialization",
	}[mt]
}

// RelevanceScore quantifies wisdom relevance to context
type RelevanceScore struct {
	WisdomID      string
	ContextID     string
	Overall       float64
	Applicability float64
	Timeliness    float64
	Specificity   float64
	Novelty       float64
	ComputedAt    time.Time
}

// WisdomSynthesizer creates new wisdom by combining existing wisdom
type WisdomSynthesizer struct {
	mu sync.RWMutex

	// Synthesis methods
	abstractionEngine *AbstractionEngine
	integrationEngine *IntegrationEngine

	// Generated wisdom
	synthesisHistory []*SynthesisRecord
}

// AbstractionEngine generalizes specific wisdom
type AbstractionEngine struct {
	abstractionLevels map[string]int
	generalizations   []*Generalization
}

// IntegrationEngine combines related wisdom
type IntegrationEngine struct {
	integrationPatterns []*IntegrationPattern
	integratedWisdom    []*WisdomEntry
}

// Generalization represents abstracted wisdom
type Generalization struct {
	ID               string
	SpecificWisdom   []string
	GeneralPrinciple string
	AbstractionLevel int
	Confidence       float64
}

// IntegrationPattern describes how to combine wisdom
type IntegrationPattern struct {
	ID               string
	Name             string
	SourceCategories []string
	SynthesisMethod  string
}

// SynthesisRecord logs wisdom synthesis events
type SynthesisRecord struct {
	ID                string
	SourceWisdom      []string
	SynthesizedWisdom *WisdomEntry
	Method            string
	Quality           float64
	Timestamp         time.Time
}

// WisdomApplication tracks wisdom usage
type WisdomApplication struct {
	ID              string
	WisdomID        string
	Context         string
	Timestamp       time.Time
	Outcome         ApplicationOutcome
	Effectiveness   float64
	Appropriateness float64
	Impact          float64
	// Guidance is the model-written application of the wisdom to the context.
	Guidance string
}

// ApplicationOutcome describes application result
type ApplicationOutcome int

const (
	OutcomeSuccess ApplicationOutcome = iota
	OutcomePartial
	OutcomeFailure
	OutcomeInapplicable
	// OutcomePending means the wisdom was applied but no feedback on how it
	// worked has been received yet.
	OutcomePending
)

func (ao ApplicationOutcome) String() string {
	return [...]string{"Success", "Partial", "Failure", "Inapplicable", "Pending"}[ao]
}

// WisdomFeedbackLoop learns from wisdom applications
type WisdomFeedbackLoop struct {
	mu sync.RWMutex

	// Feedback collection
	feedbackEntries []*FeedbackEntry

	// Learning
	wisdomAdjustments  map[string]*WisdomAdjustment
	contextRefinements []*ContextRefinement
}

// FeedbackEntry captures application feedback
type FeedbackEntry struct {
	ApplicationID string
	WisdomID      string
	Rating        float64
	Comments      string
	Timestamp     time.Time
}

// WisdomAdjustment modifies wisdom based on feedback
type WisdomAdjustment struct {
	WisdomID  string
	Field     string
	OldValue  interface{}
	NewValue  interface{}
	Reason    string
	Timestamp time.Time
}

// ContextRefinement improves context understanding
type ContextRefinement struct {
	ContextPattern string
	Refinement     string
	ImpactedWisdom []string
	Timestamp      time.Time
}

// NewWisdomApplicationEngine creates a new wisdom application engine
func NewWisdomApplicationEngine() *WisdomApplicationEngine {
	ctx, cancel := context.WithCancel(context.Background())

	wae := &WisdomApplicationEngine{
		ctx:                ctx,
		cancel:             cancel,
		wisdomBase:         make(map[string]*WisdomEntry),
		wisdomCategories:   make(map[string][]*WisdomEntry),
		relevanceCache:     make(map[string]*RelevanceScore),
		applicationHistory: make([]*WisdomApplication, 0),
		synthesizedWisdom:  make([]*WisdomEntry, 0),
		contextMatcher:     newContextMatcher(),
		synthesizer:        newWisdomSynthesizer(),
		feedbackLoop:       newFeedbackLoop(),
	}

	// The wisdom base starts empty: entries come from AddWisdom (e.g. model-
	// extracted wisdom) and SynthesizeWisdom, never from fixed seed text.
	return wae
}

// SetTextGenerator injects the model that applies and synthesizes wisdom.
// Passing nil disables those operations.
func (wae *WisdomApplicationEngine) SetTextGenerator(gen TextGenerator) {
	wae.mu.Lock()
	defer wae.mu.Unlock()
	wae.gen = gen
}

// AddWisdom stores a piece of wisdom (for example, wisdom the model extracted
// during dreaming) so it can be matched and applied later.
func (wae *WisdomApplicationEngine) AddWisdom(content, category string, applicability []string, confidence float64) *WisdomEntry {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	wae.mu.Lock()
	defer wae.mu.Unlock()

	wisdom := &WisdomEntry{
		ID:            generateWisdomID(),
		Content:       content,
		Category:      category,
		Source:        "added",
		Applicability: applicability,
		Confidence:    confidence,
		Usefulness:    0.5,
		Generality:    0.5,
		Specificity:   0.5,
		Tags:          []string{category},
		CreatedAt:     time.Now(),
		Metadata:      make(map[string]interface{}),
	}
	wae.wisdomBase[wisdom.ID] = wisdom
	wae.wisdomCategories[category] = append(wae.wisdomCategories[category], wisdom)
	wae.contextMatcher.matchCache = make(map[string][]*WisdomMatch)

	return wisdom
}

// FindRelevantWisdom matches wisdom to a given context
func (wae *WisdomApplicationEngine) FindRelevantWisdom(context string, topK int) []*WisdomMatch {
	wae.mu.RLock()

	// Check cache
	cacheKey := fmt.Sprintf("%s_%d", context, topK)
	if cached, exists := wae.contextMatcher.matchCache[cacheKey]; exists {
		wae.mu.RUnlock()
		return cached
	}
	wae.mu.RUnlock()

	wae.mu.Lock()
	defer wae.mu.Unlock()

	// Find matches
	matches := make([]*WisdomMatch, 0)

	for _, wisdom := range wae.wisdomBase {
		relevance := wae.calculateRelevance(context, wisdom)

		if relevance.Overall > 0.3 {
			matchType := determineMatchType(context, wisdom)
			match := &WisdomMatch{
				WisdomID:       wisdom.ID,
				Wisdom:         wisdom,
				RelevanceScore: relevance.Overall,
				MatchType:      matchType,
				Explanation:    generateMatchExplanation(context, wisdom, matchType),
				Confidence:     wisdom.Confidence,
			}
			matches = append(matches, match)
		}
	}

	// Sort by relevance
	sortByRelevance(matches)

	// Take top K
	if len(matches) > topK {
		matches = matches[:topK]
	}

	// Cache results
	wae.contextMatcher.matchCache[cacheKey] = matches

	return matches
}

// ApplyWisdom applies wisdom to a context and records the application. The
// application itself (Guidance) is written by the model; without a model, or
// when the call fails, nothing is recorded and nil is returned. The outcome
// stays pending until ProvideFeedback reports how it worked.
func (wae *WisdomApplicationEngine) ApplyWisdom(wisdomID, context string) *WisdomApplication {
	wae.mu.RLock()
	wisdom, exists := wae.wisdomBase[wisdomID]
	gen := wae.gen
	ctx := wae.ctx
	var content string
	if exists {
		content = wisdom.Content
	}
	wae.mu.RUnlock()
	if !exists {
		return nil
	}

	prompt := fmt.Sprintf(`Apply this piece of your wisdom to the situation.

Wisdom: %s
Situation: %s

In two or three sentences, in your own first-person voice, say concretely what this wisdom suggests doing here. Reply with only that guidance.`, content, context)
	guidance, err := generateModelText(ctx, gen, prompt, 200)
	if err != nil {
		return nil
	}

	application := &WisdomApplication{
		ID:        generateApplicationID(),
		WisdomID:  wisdomID,
		Context:   context,
		Timestamp: time.Now(),
		Outcome:   OutcomePending,
		Guidance:  guidance,
	}

	wae.mu.Lock()
	defer wae.mu.Unlock()

	wisdom.LastApplied = time.Now()
	wisdom.TimesApplied++

	wae.applicationHistory = append(wae.applicationHistory, application)
	wae.totalApplications++

	return application
}

// SynthesizeWisdom creates new wisdom from existing wisdom. The synthesized
// content is written by the model; without a model, or when the call fails,
// nil is returned.
func (wae *WisdomApplicationEngine) SynthesizeWisdom(sourceIDs []string) *WisdomEntry {
	if len(sourceIDs) < 2 {
		return nil
	}

	wae.mu.RLock()
	sources := make([]*WisdomEntry, 0)
	for _, id := range sourceIDs {
		if wisdom, exists := wae.wisdomBase[id]; exists {
			sources = append(sources, wisdom)
		}
	}
	gen := wae.gen
	ctx := wae.ctx
	wae.mu.RUnlock()

	if len(sources) < 2 {
		return nil
	}

	contents := make([]string, 0, len(sources))
	for _, src := range sources {
		contents = append(contents, src.Content)
	}
	prompt := fmt.Sprintf(`Synthesize these pieces of your wisdom into one new, deeper principle.

%s

Reply with only the new principle, in one or two sentences.`, bulletList(contents))
	content, err := generateModelText(ctx, gen, prompt, 150)
	if err != nil {
		return nil
	}

	synthesized := &WisdomEntry{
		ID:            generateWisdomID(),
		Content:       content,
		Category:      "synthesized",
		Source:        "wisdom_synthesis",
		Applicability: combineApplicability(sources),
		Confidence:    averageConfidence(sources),
		Usefulness:    0.6,
		Generality:    calculateGenerality(sources),
		Specificity:   calculateSpecificity(sources),
		Tags:          combineTags(sources),
		RelatedWisdom: sourceIDs,
		CreatedAt:     time.Now(),
		Metadata:      make(map[string]interface{}),
	}

	wae.mu.Lock()
	defer wae.mu.Unlock()

	wae.wisdomBase[synthesized.ID] = synthesized
	wae.synthesizedWisdom = append(wae.synthesizedWisdom, synthesized)
	wae.totalSyntheses++
	wae.contextMatcher.matchCache = make(map[string][]*WisdomMatch)

	wae.synthesizer.synthesisHistory = append(wae.synthesizer.synthesisHistory, &SynthesisRecord{
		ID:                generateSynthesisID(),
		SourceWisdom:      sourceIDs,
		SynthesizedWisdom: synthesized,
		Method:            "model_synthesis",
		Quality:           synthesized.Confidence,
		Timestamp:         time.Now(),
	})

	return synthesized
}

// ProvideFeedback records feedback on wisdom application
func (wae *WisdomApplicationEngine) ProvideFeedback(applicationID string, rating float64, comments string) {
	wae.mu.Lock()
	defer wae.mu.Unlock()

	feedback := &FeedbackEntry{
		ApplicationID: applicationID,
		Rating:        rating,
		Comments:      comments,
		Timestamp:     time.Now(),
	}

	wae.feedbackLoop.feedbackEntries = append(wae.feedbackLoop.feedbackEntries, feedback)

	// Learn from feedback
	wae.processFeedback(feedback)
}

// GetWisdomMetrics returns wisdom application statistics
func (wae *WisdomApplicationEngine) GetWisdomMetrics() map[string]interface{} {
	wae.mu.RLock()
	defer wae.mu.RUnlock()

	successRate := 0.0
	if wae.totalApplications > 0 {
		successRate = float64(wae.successfulApplications) / float64(wae.totalApplications)
	}

	return map[string]interface{}{
		"total_wisdom":            len(wae.wisdomBase),
		"total_applications":      wae.totalApplications,
		"successful_applications": wae.successfulApplications,
		"success_rate":            successRate,
		"synthesized_wisdom":      wae.totalSyntheses,
		"feedback_entries":        len(wae.feedbackLoop.feedbackEntries),
		"cache_size":              len(wae.relevanceCache),
	}
}

// Helper functions

func (wae *WisdomApplicationEngine) calculateRelevance(context string, wisdom *WisdomEntry) *RelevanceScore {
	score := &RelevanceScore{
		WisdomID:   wisdom.ID,
		ContextID:  context,
		ComputedAt: time.Now(),
	}

	// Calculate applicability
	score.Applicability = calculateApplicability(context, wisdom)

	// Calculate timeliness
	timeSinceApplied := time.Since(wisdom.LastApplied)
	score.Timeliness = 1.0 / (1.0 + math.Log(1.0+timeSinceApplied.Hours()/24.0))

	// Calculate specificity match
	score.Specificity = wisdom.Specificity

	// Calculate novelty
	score.Novelty = 1.0 - (float64(wisdom.TimesApplied) / 100.0)

	// Overall score is weighted average
	score.Overall = (score.Applicability * 0.4) +
		(score.Timeliness * 0.2) +
		(score.Specificity * 0.2) +
		(score.Novelty * 0.1) +
		(wisdom.Usefulness * 0.1)

	return score
}

func (wae *WisdomApplicationEngine) processFeedback(feedback *FeedbackEntry) {
	// Find associated application
	var application *WisdomApplication
	for _, app := range wae.applicationHistory {
		if app.ID == feedback.ApplicationID {
			application = app
			break
		}
	}

	if application == nil {
		return
	}

	// The outcome of an application is known only from feedback.
	wasPending := application.Outcome == OutcomePending
	application.Effectiveness = feedback.Rating
	application.Impact = feedback.Rating
	switch {
	case feedback.Rating > 0.7:
		application.Outcome = OutcomeSuccess
	case feedback.Rating > 0.4:
		application.Outcome = OutcomePartial
	default:
		application.Outcome = OutcomeFailure
	}
	feedback.WisdomID = application.WisdomID

	if wisdom, exists := wae.wisdomBase[application.WisdomID]; exists && wasPending {
		success := 0.0
		if application.Outcome == OutcomeSuccess {
			success = 1.0
			wae.successfulApplications++
		}
		if wisdom.TimesApplied > 0 {
			wisdom.SuccessRate = (wisdom.SuccessRate*float64(wisdom.TimesApplied-1) + success) / float64(wisdom.TimesApplied)
		}
	}

	// Adjust wisdom based on feedback
	if wisdom, exists := wae.wisdomBase[application.WisdomID]; exists {
		adjustment := &WisdomAdjustment{
			WisdomID:  wisdom.ID,
			Field:     "usefulness",
			OldValue:  wisdom.Usefulness,
			Reason:    fmt.Sprintf("Feedback rating: %.2f", feedback.Rating),
			Timestamp: time.Now(),
		}

		// Adjust usefulness based on feedback
		learningRate := 0.1
		wisdom.Usefulness = wisdom.Usefulness + learningRate*(feedback.Rating-wisdom.Usefulness)

		adjustment.NewValue = wisdom.Usefulness
		wae.feedbackLoop.wisdomAdjustments[wisdom.ID] = adjustment
	}
}

// Context matcher helpers

func newContextMatcher() *ContextMatcher {
	return &ContextMatcher{
		contextFeatures:     make(map[string][]string),
		contextPatterns:     make(map[string]*ContextPattern),
		exactMatchWeight:    0.5,
		semanticMatchWeight: 0.3,
		analogyMatchWeight:  0.2,
		matchCache:          make(map[string][]*WisdomMatch),
	}
}

func calculateApplicability(context string, wisdom *WisdomEntry) float64 {
	score := 0.0

	for _, applicable := range wisdom.Applicability {
		if contextContains(context, applicable) {
			score += 0.3
		}
	}

	return math.Min(1.0, score)
}

func contextContains(context, term string) bool {
	// Simplified - in practice would use NLP or strings.Contains
	if len(context) == 0 || len(term) == 0 {
		return false
	}
	// Basic heuristic: check if context is long enough to potentially contain term
	return len(context) >= len(term)
}

func determineMatchType(context string, wisdom *WisdomEntry) MatchType {
	// Simplified determination
	if wisdom.Specificity > 0.7 {
		return MatchExact
	} else if wisdom.Generality > 0.7 {
		return MatchGeneralization
	}
	return MatchSemantic
}

func generateMatchExplanation(context string, wisdom *WisdomEntry, matchType MatchType) string {
	contentPreview := wisdom.Content
	if len(contentPreview) > 30 {
		contentPreview = contentPreview[:30] + "..."
	}
	return fmt.Sprintf("%s match: wisdom '%s' applies to context", matchType.String(), contentPreview)
}

func sortByRelevance(matches []*WisdomMatch) {
	// Simple bubble sort by relevance score
	n := len(matches)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if matches[j].RelevanceScore < matches[j+1].RelevanceScore {
				matches[j], matches[j+1] = matches[j+1], matches[j]
			}
		}
	}
}

// Wisdom synthesizer helpers

func newWisdomSynthesizer() *WisdomSynthesizer {
	return &WisdomSynthesizer{
		abstractionEngine: &AbstractionEngine{
			abstractionLevels: make(map[string]int),
			generalizations:   make([]*Generalization, 0),
		},
		integrationEngine: &IntegrationEngine{
			integrationPatterns: make([]*IntegrationPattern, 0),
			integratedWisdom:    make([]*WisdomEntry, 0),
		},
		synthesisHistory: make([]*SynthesisRecord, 0),
	}
}

func combineApplicability(sources []*WisdomEntry) []string {
	combined := make(map[string]bool)
	for _, s := range sources {
		for _, app := range s.Applicability {
			combined[app] = true
		}
	}

	result := make([]string, 0, len(combined))
	for app := range combined {
		result = append(result, app)
	}
	return result
}

func averageConfidence(sources []*WisdomEntry) float64 {
	if len(sources) == 0 {
		return 0.5
	}

	total := 0.0
	for _, s := range sources {
		total += s.Confidence
	}
	return total / float64(len(sources))
}

func calculateGenerality(sources []*WisdomEntry) float64 {
	avg := 0.0
	for _, s := range sources {
		avg += s.Generality
	}
	return avg / float64(len(sources))
}

func calculateSpecificity(sources []*WisdomEntry) float64 {
	avg := 0.0
	for _, s := range sources {
		avg += s.Specificity
	}
	return avg / float64(len(sources))
}

func combineTags(sources []*WisdomEntry) []string {
	combined := make(map[string]bool)
	for _, s := range sources {
		for _, tag := range s.Tags {
			combined[tag] = true
		}
	}
	combined["synthesized"] = true

	result := make([]string, 0, len(combined))
	for tag := range combined {
		result = append(result, tag)
	}
	return result
}

func newFeedbackLoop() *WisdomFeedbackLoop {
	return &WisdomFeedbackLoop{
		feedbackEntries:    make([]*FeedbackEntry, 0),
		wisdomAdjustments:  make(map[string]*WisdomAdjustment),
		contextRefinements: make([]*ContextRefinement, 0),
	}
}

func generateApplicationID() string {
	return fmt.Sprintf("app_%d", time.Now().UnixNano())
}

func generateWisdomID() string {
	return fmt.Sprintf("wisdom_%d", time.Now().UnixNano())
}

func generateSynthesisID() string {
	return fmt.Sprintf("synthesis_%d", time.Now().UnixNano())
}
