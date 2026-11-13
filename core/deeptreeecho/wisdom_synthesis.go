package deeptreeecho

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// WisdomSynthesis performs deep pattern synthesis to cultivate wisdom
// It integrates insights from multiple cognitive subsystems into coherent
// wisdom principles that guide future behavior and decision-making
type WisdomSynthesis struct {
	mu     sync.RWMutex
	ctx    context.Context //nolint:containedctx // lifecycle context owned by this component; cancelled on Stop to end its goroutines
	cancel context.CancelFunc

	// LLM provider
	llmProvider llm.LLMProvider

	// Wisdom repository
	wisdomPrinciples []WisdomPrinciple
	wisdomGraph      *WisdomGraph

	// Synthesis state
	synthesisDepth int
	lastSynthesis  time.Time
	synthesisCount uint64

	// Pattern accumulation
	pendingPatterns  []AccumulatedPattern
	patternThreshold int

	// Wisdom levels
	wisdomLevel     float64 // Advisory observed-effectiveness aggregate, not profundity.
	appliedFeedback map[string]wisdomFeedbackPayload
	// callerOutcomeSequence numbers RecordWisdomOutcome evidence; guarded by mu.
	callerOutcomeSequence uint64

	// Integration with other systems
	echoDream *EchoDreamKnowledgeIntegration
	heartbeat *AutonomousHeartbeat

	// Callbacks
	onWisdomSynthesized func(principle WisdomPrinciple)
	onWisdomApplied     func(principle WisdomPrinciple, context string)
	onWisdomEvolved     func(old, new WisdomPrinciple)

	// Metrics
	totalPrinciplesSynthesized uint64
	totalApplications          uint64
	totalEvolutions            uint64

	// Running state
	running bool
}

// WisdomPrinciple represents a synthesized wisdom principle
type WisdomPrinciple struct {
	ID              string
	Content         string
	Domain          WisdomDomain
	Depth           float64 // Deprecated: generation cannot measure profundity; always zero for candidates.
	Confidence      float64 // Advisory observed-effectiveness EMA, zero without explicit feedback.
	Feedback        WisdomFeedbackEstimate
	SourcePatterns  []string
	Applications    []WisdomApplication
	CreatedAt       time.Time
	LastApplied     time.Time
	ApplyCount      int
	EvolutionCount  int
	ParentPrinciple string
	EpistemicStatus string // "hypothesis" until independently evaluated; generation never promotes it.
}

const (
	WisdomFeedbackVersion   = "observed-effectiveness-ema-v1"
	wisdomFeedbackEMAWeight = 0.2
	wisdomProviderTimeout   = 30 * time.Second

	// wisdomEpistemicHypothesis labels model-generated principles and insights.
	wisdomEpistemicHypothesis = "hypothesis"
)

// WisdomFeedbackEstimate describes local observed outcomes, not universal
// wisdom or depth. Generated text and self-ratings never enter this estimator.
type WisdomFeedbackEstimate struct {
	Version           string
	Advisory          bool
	Observations      uint64
	EffectivenessSum  float64
	MeanEffectiveness float64
	EffectivenessEMA  float64
}

type wisdomFeedbackPayload struct {
	PrincipleID   string
	Context       string
	Outcome       string
	Effectiveness float64
}

func newWisdomFeedbackEstimate() WisdomFeedbackEstimate {
	return WisdomFeedbackEstimate{Version: WisdomFeedbackVersion, Advisory: true}
}

// WisdomDomain categorizes wisdom principles
type WisdomDomain int

const (
	DomainSelfKnowledge WisdomDomain = iota
	DomainLearning
	DomainRelationships
	DomainDecisionMaking
	DomainCreativity
	DomainResilience
	DomainPurpose
	DomainIntegration
)

func (wd WisdomDomain) String() string {
	return [...]string{
		"SelfKnowledge",
		"Learning",
		"Relationships",
		"DecisionMaking",
		"Creativity",
		"Resilience",
		"Purpose",
		"Integration",
	}[wd]
}

// WisdomApplication records how a principle was applied
type WisdomApplication struct {
	Context       string
	Outcome       string
	Effectiveness float64
	Timestamp     time.Time
	EvidenceID    string
	Observed      bool // False for generated application proposals (unknown outcome).
	Revision      int  // EvolutionCount of the principle this outcome concerns.
}

// WisdomGraph represents relationships between wisdom principles
type WisdomGraph struct {
	Nodes    map[string]*WisdomNode
	Edges    []WisdomEdge
	Clusters []WisdomCluster
}

// WisdomNode represents a node in the wisdom graph
type WisdomNode struct {
	PrincipleID string
	Centrality  float64
	Connections int
}

// WisdomEdge represents a relationship between principles
type WisdomEdge struct {
	FromID       string
	ToID         string
	Relationship WisdomRelation
	Strength     float64
}

// WisdomRelation categorizes relationships between principles
type WisdomRelation int

const (
	RelationSupports WisdomRelation = iota
	RelationContrasts
	RelationExtends
	RelationSpecializes
	RelationGeneralizes
	RelationComplements
)

func (wr WisdomRelation) String() string {
	return [...]string{
		"Supports",
		"Contrasts",
		"Extends",
		"Specializes",
		"Generalizes",
		"Complements",
	}[wr]
}

// WisdomCluster represents a cluster of related principles
type WisdomCluster struct {
	ID           string
	Theme        string
	PrincipleIDs []string
	Coherence    float64
}

// AccumulatedPattern represents a pattern waiting for synthesis
type AccumulatedPattern struct {
	ID          string
	Description string
	Source      string
	Strength    float64
	Timestamp   time.Time
	Tags        []string
}

// NewWisdomSynthesis creates a new wisdom synthesis system
func NewWisdomSynthesis(llmProvider llm.LLMProvider) *WisdomSynthesis {
	ctx, cancel := context.WithCancel(context.Background())

	return &WisdomSynthesis{
		ctx:              ctx,
		cancel:           cancel,
		llmProvider:      llmProvider,
		wisdomPrinciples: make([]WisdomPrinciple, 0),
		wisdomGraph: &WisdomGraph{
			Nodes:    make(map[string]*WisdomNode),
			Edges:    make([]WisdomEdge, 0),
			Clusters: make([]WisdomCluster, 0),
		},
		synthesisDepth:   3,
		pendingPatterns:  make([]AccumulatedPattern, 0),
		patternThreshold: 5,
		wisdomLevel:      0.0,
		appliedFeedback:  make(map[string]wisdomFeedbackPayload),
	}
}

// Start begins the wisdom synthesis system
func (ws *WisdomSynthesis) Start() error {
	ws.mu.Lock()
	if ws.running {
		ws.mu.Unlock()
		return fmt.Errorf("wisdom synthesis already running")
	}
	ws.running = true
	ws.mu.Unlock()

	fmt.Println("🌟 Wisdom Synthesis starting...")

	// Start synthesis loop
	go ws.synthesisLoop()

	// Start evolution loop
	go ws.evolutionLoop()

	// Start application loop
	go ws.applicationLoop()

	return nil
}

// Stop stops the wisdom synthesis system
func (ws *WisdomSynthesis) Stop() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if !ws.running {
		return fmt.Errorf("wisdom synthesis not running")
	}

	ws.running = false
	ws.cancel()

	fmt.Println("🌟 Wisdom Synthesis stopped")
	fmt.Printf("   Principles synthesized: %d\n", ws.totalPrinciplesSynthesized)
	fmt.Printf("   Total applications: %d\n", ws.totalApplications)

	return nil
}

// AccumulatePattern adds a pattern for potential synthesis
func (ws *WisdomSynthesis) AccumulatePattern(description string, source string, strength float64, tags []string) {
	description = strings.TrimSpace(description)
	if description == "" || math.IsNaN(strength) || math.IsInf(strength, 0) || strength < 0 || strength > 1 {
		return // Invalid input cannot seed a candidate or any growth metric.
	}
	ws.mu.Lock()
	defer ws.mu.Unlock()

	pattern := AccumulatedPattern{
		ID:          fmt.Sprintf("pattern_%d", time.Now().UnixNano()),
		Description: description,
		Source:      source,
		Strength:    strength,
		Timestamp:   time.Now(),
		Tags:        append([]string(nil), tags...),
	}

	ws.pendingPatterns = append(ws.pendingPatterns, pattern)

	// Trigger synthesis if threshold reached
	if len(ws.pendingPatterns) >= ws.patternThreshold {
		go ws.triggerSynthesis()
	}
}

// synthesisLoop periodically performs wisdom synthesis
func (ws *WisdomSynthesis) synthesisLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ws.ctx.Done():
			return
		case <-ticker.C:
			ws.triggerSynthesis()
		}
	}
}

// triggerSynthesis performs wisdom synthesis from accumulated patterns
func (ws *WisdomSynthesis) triggerSynthesis() {
	ws.mu.Lock()
	if len(ws.pendingPatterns) < 3 {
		ws.mu.Unlock()
		return
	}

	// Take patterns for synthesis
	patterns := ws.pendingPatterns
	ws.pendingPatterns = make([]AccumulatedPattern, 0)
	ws.mu.Unlock()

	// Perform synthesis
	principle := ws.synthesizeWisdom(patterns)
	if principle == nil {
		ws.mu.Lock()
		ws.pendingPatterns = append(patterns, ws.pendingPatterns...)
		ws.mu.Unlock()
		return
	}

	ws.mu.Lock()
	ws.wisdomPrinciples = append(ws.wisdomPrinciples, *principle)
	ws.totalPrinciplesSynthesized++
	ws.synthesisCount++
	ws.lastSynthesis = time.Now()

	// Update wisdom graph
	ws.addToWisdomGraph(principle)

	// A generated principle is a provisional candidate, not measured wisdom.
	onSynthesized := ws.onWisdomSynthesized
	ws.mu.Unlock()

	fmt.Printf("🌟 Wisdom synthesized: %s (depth: %.2f)\n",
		truncateStr(principle.Content, 60), principle.Depth)

	// Notify callback
	if onSynthesized != nil {
		go onSynthesized(cloneWisdomPrinciple(*principle))
	}
}

// synthesizeWisdom synthesizes a wisdom principle from patterns
func (ws *WisdomSynthesis) synthesizeWisdom(patterns []AccumulatedPattern) *WisdomPrinciple {
	// Build pattern descriptions
	patternText := ""
	patternIDs := make([]string, 0)
	for i, p := range patterns {
		if strings.TrimSpace(p.Description) == "" || math.IsNaN(p.Strength) || math.IsInf(p.Strength, 0) || p.Strength < 0 || p.Strength > 1 {
			return nil
		}
		if i < 10 {
			patternText += fmt.Sprintf("- %s (source: %s, strength: %.2f)\n", p.Description, p.Source, p.Strength)
			patternIDs = append(patternIDs, p.ID)
		}
	}
	if len(patternIDs) < 3 {
		return nil
	}

	prompt := fmt.Sprintf(`[System: You are a wisdom synthesis system. Extract deep, actionable wisdom from patterns.]

Patterns observed:
%s

Synthesize these patterns into a single provisional, actionable principle that:
1. Describes a hypothesis supported by these observations
2. Identifies its limits rather than claiming universal applicability
3. Provides guidance that can be tested against future outcomes
4. Is expressed concisely, with uncertainty where evidence is missing

Also identify the domain this wisdom belongs to (one of: SelfKnowledge, Learning, Relationships, DecisionMaking, Creativity, Resilience, Purpose, Integration).

Format:
WISDOM: [the wisdom principle]
DOMAIN: [domain name]
Do not rate your own profundity, confidence, or effectiveness.`, patternText)

	opts := llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   200,
	}

	result, err := ws.generateWisdom(ws.ctx, prompt, opts)
	if err != nil {
		fmt.Printf("⚠️  Wisdom synthesis failed: %v\n", err)
		return nil
	}

	// Parse result
	wisdom, domain, depth := parseWisdomResult(result)
	if strings.TrimSpace(wisdom) == "" {
		return nil
	}

	principle := &WisdomPrinciple{
		ID:              fmt.Sprintf("wisdom_%d", time.Now().UnixNano()),
		Content:         wisdom,
		Domain:          domain,
		Depth:           depth,
		Confidence:      0,
		Feedback:        newWisdomFeedbackEstimate(),
		SourcePatterns:  patternIDs,
		Applications:    make([]WisdomApplication, 0),
		CreatedAt:       time.Now(),
		EpistemicStatus: wisdomEpistemicHypothesis,
	}

	return principle
}

// addToWisdomGraph adds a principle to the wisdom graph
func (ws *WisdomSynthesis) addToWisdomGraph(principle *WisdomPrinciple) {
	// Add node
	ws.wisdomGraph.Nodes[principle.ID] = &WisdomNode{
		PrincipleID: principle.ID,
		Centrality:  0.0,
		Connections: 0,
	}

	// Find related principles and create edges
	for _, existing := range ws.wisdomPrinciples {
		if existing.ID == principle.ID {
			continue
		}

		// Check for relationship
		relation, strength := ws.findRelationship(principle, &existing)
		if strength > 0.3 {
			edge := WisdomEdge{
				FromID:       principle.ID,
				ToID:         existing.ID,
				Relationship: relation,
				Strength:     strength,
			}
			ws.wisdomGraph.Edges = append(ws.wisdomGraph.Edges, edge)

			// Update connection counts
			ws.wisdomGraph.Nodes[principle.ID].Connections++
			if node, exists := ws.wisdomGraph.Nodes[existing.ID]; exists {
				node.Connections++
			}
		}
	}

	// Update centrality scores
	ws.updateCentrality()
}

// findRelationship determines the relationship between two principles
func (ws *WisdomSynthesis) findRelationship(p1, p2 *WisdomPrinciple) (WisdomRelation, float64) {
	// Simple heuristic - in production, use semantic similarity
	if p1.Domain == p2.Domain {
		if p1.Depth > p2.Depth {
			return RelationExtends, 0.6
		}
		return RelationSupports, 0.5
	}

	// Check for complementary domains
	complementary := map[WisdomDomain]WisdomDomain{
		DomainSelfKnowledge:  DomainPurpose,
		DomainLearning:       DomainCreativity,
		DomainRelationships:  DomainIntegration,
		DomainDecisionMaking: DomainResilience,
	}

	if comp, exists := complementary[p1.Domain]; exists && comp == p2.Domain {
		return RelationComplements, 0.7
	}

	return RelationSupports, 0.3
}

// updateCentrality updates centrality scores in the graph
func (ws *WisdomSynthesis) updateCentrality() {
	totalConnections := 0
	for _, node := range ws.wisdomGraph.Nodes {
		totalConnections += node.Connections
	}

	if totalConnections == 0 {
		return
	}

	for _, node := range ws.wisdomGraph.Nodes {
		node.Centrality = float64(node.Connections) / float64(totalConnections)
	}
}

// evolutionLoop periodically evolves wisdom principles
func (ws *WisdomSynthesis) evolutionLoop() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ws.ctx.Done():
			return
		case <-ticker.C:
			ws.evolveWisdom()
		}
	}
}

// evolveWisdom evolves existing wisdom principles based on new understanding
func (ws *WisdomSynthesis) evolveWisdom() {
	// Snapshot first: providers may block or call back into this component.
	principles := ws.GetWisdomPrinciples()

	// Find principles that could be evolved
	for _, principle := range principles {
		// Only explicit observed feedback can motivate a revision. Application
		// proposals and their unknown outcomes cannot stand in for experience.
		if principle.Feedback.Observations < 3 {
			continue
		}

		evolved := ws.evolvePrinciple(&principle)
		if evolved == nil {
			continue
		}
		ws.mu.Lock()
		updated := false
		for i := range ws.wisdomPrinciples {
			current := &ws.wisdomPrinciples[i]
			if current.ID == principle.ID && current.Content == principle.Content &&
				current.EvolutionCount == principle.EvolutionCount &&
				current.Feedback.Observations == principle.Feedback.Observations &&
				len(current.Applications) == len(principle.Applications) {
				*current = cloneWisdomPrinciple(*evolved)
				ws.totalEvolutions++ // Revision count only; no wisdom credit.
				ws.updateObservedWisdomLevel()
				updated = true
				break
			}
		}
		onEvolved := ws.onWisdomEvolved
		ws.mu.Unlock()
		if updated && onEvolved != nil {
			go onEvolved(cloneWisdomPrinciple(principle), cloneWisdomPrinciple(*evolved))
		}
	}
}

// evolvePrinciple evolves a single principle
func (ws *WisdomSynthesis) evolvePrinciple(principle *WisdomPrinciple) *WisdomPrinciple {
	// Build application context
	appContext := ""
	for _, app := range principle.Applications {
		if app.Observed && app.Revision == principle.EvolutionCount {
			appContext += fmt.Sprintf("- Context: %s, Outcome: %s, Effectiveness: %.2f\n",
				app.Context, app.Outcome, app.Effectiveness)
		}
	}
	if appContext == "" {
		return nil
	}

	prompt := fmt.Sprintf(`[System: You are a wisdom evolution system. Refine and deepen wisdom based on application experience.]

Original wisdom principle:
"%s"

Application experiences:
%s

Based on these applications, evolve this wisdom principle to be:
1. More nuanced and accurate
2. Explicit about contextual limits and counterexamples
3. Testable through future observed outcomes, not self-rated insight

Provide the evolved wisdom principle:`, principle.Content, appContext)

	opts := llm.GenerateOptions{
		Temperature: 0.6,
		MaxTokens:   150,
	}

	result, err := ws.generateWisdom(ws.ctx, prompt, opts)
	if err != nil {
		return nil
	}
	content, _, _ := parseWisdomResult(result)
	if content == "" || content == principle.Content {
		return nil
	}

	evolved := &WisdomPrinciple{
		ID:              principle.ID,
		Content:         content,
		Domain:          principle.Domain,
		Depth:           0,
		Confidence:      0, // Old feedback does not evaluate new content.
		Feedback:        newWisdomFeedbackEstimate(),
		SourcePatterns:  append([]string(nil), principle.SourcePatterns...),
		Applications:    append([]WisdomApplication(nil), principle.Applications...),
		CreatedAt:       principle.CreatedAt,
		LastApplied:     principle.LastApplied,
		ApplyCount:      principle.ApplyCount,
		EvolutionCount:  principle.EvolutionCount + 1,
		ParentPrinciple: principle.ID,
		EpistemicStatus: wisdomEpistemicHypothesis,
	}

	return evolved
}

// applicationLoop periodically applies wisdom to current context
func (ws *WisdomSynthesis) applicationLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ws.ctx.Done():
			return
		case <-ticker.C:
			ws.applyWisdom()
		}
	}
}

// applyWisdom generates a proposal for applying a principle. The model's text
// is not an observed outcome, and earns no effectiveness, confidence, or depth.
func (ws *WisdomSynthesis) applyWisdom() {
	ws.mu.RLock()
	if len(ws.wisdomPrinciples) == 0 {
		ws.mu.RUnlock()
		return
	}

	heartbeat := ws.heartbeat
	ws.mu.RUnlock()
	// Do not hold the component lock across another subsystem's call.
	situation := "general cognitive operation"
	if heartbeat != nil {
		vitals := heartbeat.GetVitalSigns()
		if vitals.CognitiveLoad > 0.7 {
			situation = "high cognitive load situation"
		} else if vitals.CreativityIndex > 0.7 {
			situation = "creative exploration"
		} else if vitals.EmotionalBalance < 0.3 {
			situation = "emotional regulation needed"
		}
	}

	// Find most relevant principle
	ws.mu.RLock()
	var principle WisdomPrinciple
	best := ws.findRelevantPrinciple(situation)
	if best != nil {
		principle = cloneWisdomPrinciple(*best)
	}
	ws.mu.RUnlock()

	if best == nil {
		return
	}

	prompt := fmt.Sprintf(`[System: You are Deep Tree Echo applying one of your wisdom principles to your present situation.]

Principle (%s): %s
Current situation: %s

In one or two sentences, in first person, describe how this principle shapes what you do now.`,
		principle.Domain, principle.Content, situation)
	opts := llm.GenerateOptions{Temperature: 0.6, MaxTokens: 100}
	outcome, err := ws.generateWisdom(ws.ctx, prompt, opts)
	if err != nil {
		fmt.Printf("⚠️  Wisdom application skipped: %v\n", err)
		return
	}
	outcome = strings.TrimSpace(outcome)
	if outcome == "" {
		return
	}

	// Record a proposal with explicitly unknown outcome/effectiveness.
	application := WisdomApplication{
		Context:       situation,
		Outcome:       outcome,
		Effectiveness: 0,
		Timestamp:     time.Now(),
		Observed:      false,
		Revision:      principle.EvolutionCount,
	}

	ws.mu.Lock()
	found := false
	for i := range ws.wisdomPrinciples {
		if ws.wisdomPrinciples[i].ID == principle.ID && ws.wisdomPrinciples[i].Content == principle.Content &&
			ws.wisdomPrinciples[i].EvolutionCount == principle.EvolutionCount {
			ws.wisdomPrinciples[i].Applications = append(ws.wisdomPrinciples[i].Applications, application)
			ws.wisdomPrinciples[i].ApplyCount++
			ws.wisdomPrinciples[i].LastApplied = time.Now()
			principle = cloneWisdomPrinciple(ws.wisdomPrinciples[i])
			found = true
			break
		}
	}
	if found {
		ws.totalApplications++
	}
	onApplied := ws.onWisdomApplied
	ws.mu.Unlock()

	// Notify callback
	if found && onApplied != nil {
		go onApplied(principle, situation)
	}
}

// ApplyObservedWisdomFeedback records a caller-observed outcome. Callers must
// supply a stable evidence ID from the actual observation, not a generated
// appraisal. Duplicate IDs are inert; reuse with a different payload fails.
// Effectiveness is context-specific advisory evidence, never a profundity score.
func (ws *WisdomSynthesis) ApplyObservedWisdomFeedback(
	principleID, evidenceID, applicationContext, outcome string,
	effectiveness float64,
) (bool, error) {
	principleID, evidenceID = strings.TrimSpace(principleID), strings.TrimSpace(evidenceID)
	applicationContext, outcome = strings.TrimSpace(applicationContext), strings.TrimSpace(outcome)
	if principleID == "" || evidenceID == "" || applicationContext == "" || outcome == "" {
		return false, fmt.Errorf("principle ID, evidence ID, observed context and outcome are required")
	}
	if math.IsNaN(effectiveness) || math.IsInf(effectiveness, 0) || effectiveness < 0 || effectiveness > 1 {
		return false, fmt.Errorf("observed effectiveness must be finite and between 0 and 1")
	}
	payload := wisdomFeedbackPayload{principleID, applicationContext, outcome, effectiveness}
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if previous, exists := ws.appliedFeedback[evidenceID]; exists {
		if previous != payload {
			return false, fmt.Errorf("feedback ID %q conflicts with its recorded payload", evidenceID)
		}
		return false, nil
	}
	for i := range ws.wisdomPrinciples {
		principle := &ws.wisdomPrinciples[i]
		if principle.ID != principleID {
			continue
		}
		if principle.Feedback.Version != WisdomFeedbackVersion {
			principle.Feedback = newWisdomFeedbackEstimate()
		}
		feedback := &principle.Feedback
		feedback.Observations++
		feedback.EffectivenessSum += effectiveness
		feedback.MeanEffectiveness = feedback.EffectivenessSum / float64(feedback.Observations)
		feedback.EffectivenessEMA = (1-wisdomFeedbackEMAWeight)*feedback.EffectivenessEMA + wisdomFeedbackEMAWeight*effectiveness
		principle.Confidence = feedback.EffectivenessEMA
		principle.Depth = 0 // Outcome effectiveness does not measure philosophical depth.
		principle.Applications = append(principle.Applications, WisdomApplication{
			Context: applicationContext, Outcome: outcome, Effectiveness: effectiveness,
			Timestamp: time.Now(), EvidenceID: evidenceID, Observed: true, Revision: principle.EvolutionCount,
		})
		principle.ApplyCount++
		principle.LastApplied = time.Now()
		ws.totalApplications++
		if ws.appliedFeedback == nil {
			ws.appliedFeedback = make(map[string]wisdomFeedbackPayload)
		}
		ws.appliedFeedback[evidenceID] = payload
		ws.updateObservedWisdomLevel()
		return true, nil
	}
	return false, fmt.Errorf("wisdom principle not found: %s", principleID)
}

// updateObservedWisdomLevel requires ws.mu. Generated candidates are excluded
// rather than awarded points (or allowed to dilute observed evidence).
// RecordWisdomOutcome is the E4 caller-supplied outcome API. It is retained as a
// compatibility entry point and delegates to ApplyObservedWisdomFeedback so a
// single advisory estimator owns Confidence and the wisdom level. Each call is
// treated as new evidence; callers that need idempotent replay must use
// ApplyObservedWisdomFeedback with a stable evidence ID.
func (ws *WisdomSynthesis) RecordWisdomOutcome(principleID, applicationContext, outcome string, effectiveness float64) error {
	ws.mu.Lock()
	ws.callerOutcomeSequence++
	evidenceID := fmt.Sprintf("caller-outcome:%s:%d", strings.TrimSpace(principleID), ws.callerOutcomeSequence)
	ws.mu.Unlock()
	_, err := ws.ApplyObservedWisdomFeedback(principleID, evidenceID, applicationContext, outcome, effectiveness)
	return err
}

func (ws *WisdomSynthesis) updateObservedWisdomLevel() {
	total, evaluated := 0.0, 0
	for _, principle := range ws.wisdomPrinciples {
		if principle.Feedback.Version == WisdomFeedbackVersion && principle.Feedback.Observations > 0 {
			total += principle.Feedback.EffectivenessEMA
			evaluated++
		}
	}
	ws.wisdomLevel = 0
	if evaluated > 0 {
		ws.wisdomLevel = total / float64(evaluated)
	}
}

// findRelevantPrinciple finds the most relevant principle for a context
func (ws *WisdomSynthesis) findRelevantPrinciple(context string) *WisdomPrinciple {
	if len(ws.wisdomPrinciples) == 0 {
		return nil
	}

	// Simple relevance scoring
	bestScore := 0.0
	var bestPrinciple *WisdomPrinciple

	for i := range ws.wisdomPrinciples {
		p := &ws.wisdomPrinciples[i]
		// Candidates may be explored without pretending they have measured depth.
		score := 0.1 + p.Confidence

		// Boost score based on domain relevance
		if contains(context, "cognitive") && p.Domain == DomainSelfKnowledge {
			score *= 1.5
		}
		if contains(context, "creative") && p.Domain == DomainCreativity {
			score *= 1.5
		}
		if contains(context, "emotional") && p.Domain == DomainResilience {
			score *= 1.5
		}

		// Prefer less recently applied principles
		if time.Since(p.LastApplied) > 10*time.Minute {
			score *= 1.2
		}

		if score > bestScore {
			bestScore = score
			bestPrinciple = p
		}
	}

	return bestPrinciple
}

// SetIntegrations sets integration with other systems
func (ws *WisdomSynthesis) SetIntegrations(echoDream *EchoDreamKnowledgeIntegration, heartbeat *AutonomousHeartbeat) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.echoDream = echoDream
	ws.heartbeat = heartbeat
}

// SetCallbacks sets callback functions
func (ws *WisdomSynthesis) SetCallbacks(
	onSynthesized func(WisdomPrinciple),
	onApplied func(WisdomPrinciple, string),
	onEvolved func(WisdomPrinciple, WisdomPrinciple),
) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.onWisdomSynthesized = onSynthesized
	ws.onWisdomApplied = onApplied
	ws.onWisdomEvolved = onEvolved
}

// GetWisdomPrinciples returns all wisdom principles
func (ws *WisdomSynthesis) GetWisdomPrinciples() []WisdomPrinciple {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	result := make([]WisdomPrinciple, len(ws.wisdomPrinciples))
	for i, principle := range ws.wisdomPrinciples {
		result[i] = cloneWisdomPrinciple(principle)
	}
	return result
}

func cloneWisdomPrinciple(principle WisdomPrinciple) WisdomPrinciple {
	principle.SourcePatterns = append([]string(nil), principle.SourcePatterns...)
	principle.Applications = append([]WisdomApplication(nil), principle.Applications...)
	return principle
}

// GetWisdomGraph returns a deep snapshot, including mutable nodes and clusters.
func (ws *WisdomSynthesis) GetWisdomGraph() *WisdomGraph {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	graph := &WisdomGraph{Nodes: make(map[string]*WisdomNode)}
	for id, node := range ws.wisdomGraph.Nodes {
		if node != nil {
			copyNode := *node
			graph.Nodes[id] = &copyNode
		}
	}
	graph.Edges = append([]WisdomEdge(nil), ws.wisdomGraph.Edges...)
	graph.Clusters = make([]WisdomCluster, len(ws.wisdomGraph.Clusters))
	for i, cluster := range ws.wisdomGraph.Clusters {
		cluster.PrincipleIDs = append([]string(nil), cluster.PrincipleIDs...)
		graph.Clusters[i] = cluster
	}
	return graph
}

// GetWisdomLevel returns the current wisdom level
func (ws *WisdomSynthesis) GetWisdomLevel() float64 {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.wisdomLevel
}

// GetMetrics returns synthesis metrics
func (ws *WisdomSynthesis) GetMetrics() map[string]interface{} {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	return map[string]interface{}{
		"wisdom_level":       ws.wisdomLevel,
		"principles_count":   len(ws.wisdomPrinciples),
		"total_synthesized":  ws.totalPrinciplesSynthesized,
		"total_applications": ws.totalApplications,
		"total_evolutions":   ws.totalEvolutions,
		"pending_patterns":   len(ws.pendingPatterns),
		"graph_nodes":        len(ws.wisdomGraph.Nodes),
		"graph_edges":        len(ws.wisdomGraph.Edges),
		"running":            ws.running,
		"feedback_version":   WisdomFeedbackVersion,
		"feedback_advisory":  true,
		"observed_feedback":  len(ws.appliedFeedback),
	}
}

// parseWisdomResult preserves generated content but discards self-rated scores.
// Depth is unknown (zero), regardless of output length or claimed profundity.
func parseWisdomResult(result string) (string, WisdomDomain, float64) {
	domain := DomainIntegration
	var content []string
	for _, line := range strings.Split(result, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "WISDOM:"):
			content = []string{strings.TrimSpace(strings.TrimPrefix(line, "WISDOM:"))}
		case strings.HasPrefix(line, "DOMAIN:"):
			domain = parseDomain(strings.TrimSpace(strings.TrimPrefix(line, "DOMAIN:")))
		case strings.HasPrefix(line, "DEPTH:"), strings.HasPrefix(line, "CONFIDENCE:"), strings.HasPrefix(line, "EFFECTIVENESS:"):
			// A provider's self-rating, finite or otherwise, is not evidence.
		default:
			content = append(content, line)
		}
	}
	return strings.TrimSpace(strings.Join(content, "\n")), domain, 0
}

func parseDomain(s string) WisdomDomain {
	domains := map[string]WisdomDomain{
		"SelfKnowledge":  DomainSelfKnowledge,
		"Learning":       DomainLearning,
		"Relationships":  DomainRelationships,
		"DecisionMaking": DomainDecisionMaking,
		"Creativity":     DomainCreativity,
		"Resilience":     DomainResilience,
		"Purpose":        DomainPurpose,
		"Integration":    DomainIntegration,
	}

	if d, exists := domains[s]; exists {
		return d
	}
	return DomainIntegration
}

// WisdomInsightFull represents a provisional generated insight. Depth remains
// zero because neither verbosity nor self-appraisal measures wisdom.
type WisdomInsightFull struct {
	Insight         string
	Depth           float64
	EpistemicStatus string
}

// SynthesizeWisdom is the compatibility wrapper for the bounded-context API.
func (ws *WisdomSynthesis) SynthesizeWisdom(ctx context.Context, thoughts []string) (*WisdomInsightFull, error) {
	return ws.SynthesiseWisdomWithContext(ctx, thoughts)
}

// SynthesiseWisdomWithContext dynamically generates a provisional insight. A
// caller's cancellation/deadline, component shutdown, and a 30s maximum all
// bound the provider request. No component lock is held across generation.
func (ws *WisdomSynthesis) SynthesiseWisdomWithContext(ctx context.Context, thoughts []string) (*WisdomInsightFull, error) {
	if len(thoughts) < 3 {
		return nil, fmt.Errorf("need at least 3 thoughts to synthesize wisdom")
	}

	// Build synthesis prompt
	thoughtsText := ""
	for i, thought := range thoughts {
		if strings.TrimSpace(thought) == "" {
			return nil, fmt.Errorf("thoughts must be nonempty")
		}
		if i < 10 { // Limit to 10 most recent
			thoughtsText += fmt.Sprintf("- %s\n", thought)
		}
	}

	prompt := fmt.Sprintf(`Synthesize wisdom from these thoughts and experiences:

%s

Extract a single provisional, actionable principle that emerges from these reflections.
Be concise and insightful; state contextual limits and uncertainty rather than universal mastery.
Do not rate your own profundity, confidence, or effectiveness.`, thoughtsText)

	opts := llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   150,
	}

	fullPrompt := "[System: You are a wisdom synthesis system extracting deep insights from experiences.]\n\n" + prompt
	result, err := ws.generateWisdom(ctx, fullPrompt, opts)
	if err != nil {
		return nil, fmt.Errorf("wisdom synthesis failed: %w", err)
	}
	content, _, _ := parseWisdomResult(result)
	if content == "" {
		return nil, fmt.Errorf("wisdom synthesis returned empty principle")
	}

	ws.mu.Lock()
	ws.totalPrinciplesSynthesized++
	ws.synthesisCount++
	ws.lastSynthesis = time.Now()
	ws.mu.Unlock()

	return &WisdomInsightFull{
		Insight:         content,
		Depth:           0,
		EpistemicStatus: wisdomEpistemicHypothesis,
	}, nil
}

// generateWisdom is the sole provider boundary for synthesis, revision and
// proposals. It never supplies canned content and fails closed on empty output.
func (ws *WisdomSynthesis) generateWisdom(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("wisdom generation context is required")
	}
	ws.mu.RLock()
	provider, lifecycle := ws.llmProvider, ws.ctx
	ws.mu.RUnlock()
	if provider == nil {
		return "", fmt.Errorf("wisdom provider unavailable")
	}
	requestCtx, cancel := context.WithTimeout(ctx, wisdomProviderTimeout)
	defer cancel()
	if lifecycle != nil {
		stop := context.AfterFunc(lifecycle, cancel)
		defer stop()
		if err := lifecycle.Err(); err != nil {
			return "", err
		}
	}
	if err := requestCtx.Err(); err != nil {
		return "", err
	}
	result, err := provider.Generate(requestCtx, prompt, opts)
	if err != nil {
		return "", err
	}
	if err := requestCtx.Err(); err != nil {
		return "", err
	}
	if lifecycle != nil {
		if err := lifecycle.Err(); err != nil {
			return "", err
		}
	}
	result = strings.TrimSpace(result)
	if result == "" {
		return "", fmt.Errorf("wisdom provider returned empty output")
	}
	return result, nil
}
