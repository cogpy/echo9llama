package deeptreeecho

import (
	"context"
	"fmt"
	"math"
	"reflect"
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
	wisdomLevel      float64
	wisdomGrowthRate float64

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
	Depth           float64
	Confidence      float64
	SourcePatterns  []string
	Applications    []WisdomApplication
	CreatedAt       time.Time
	LastApplied     time.Time
	ApplyCount      int
	EvolutionCount  int
	ParentPrinciple string
	EpistemicStatus string
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
}

// RecordWisdomOutcome records an externally observed outcome for a proposed
// principle. Generation and relevance selection alone are not applications:
// they do not create an application record, change confidence, or advance
// wisdom metrics.
func (ws *WisdomSynthesis) RecordWisdomOutcome(principleID, applicationContext, outcome string, effectiveness float64) error {
	if strings.TrimSpace(principleID) == "" {
		return fmt.Errorf("wisdom principle ID is required")
	}
	if strings.TrimSpace(applicationContext) == "" {
		return fmt.Errorf("wisdom application context is required")
	}
	if strings.TrimSpace(outcome) == "" {
		return fmt.Errorf("wisdom application outcome is required")
	}
	if math.IsNaN(effectiveness) || math.IsInf(effectiveness, 0) || effectiveness < 0 || effectiveness > 1 {
		return fmt.Errorf("wisdom outcome effectiveness must be finite and within [0, 1]")
	}

	application := WisdomApplication{
		Context:       applicationContext,
		Outcome:       outcome,
		Effectiveness: effectiveness,
		Timestamp:     time.Now(),
	}

	ws.mu.Lock()
	var applied WisdomPrinciple
	var callback func(WisdomPrinciple, string)
	found := false
	for i := range ws.wisdomPrinciples {
		if ws.wisdomPrinciples[i].ID != principleID {
			continue
		}

		principle := &ws.wisdomPrinciples[i]
		principle.Applications = append(principle.Applications, application)
		principle.ApplyCount = len(principle.Applications)
		principle.LastApplied = application.Timestamp
		principle.Confidence = outcomeConfidence(principle.Applications)
		ws.totalApplications++
		ws.recalculateWisdomLevelLocked()
		applied = cloneWisdomPrinciple(*principle)
		callback = ws.onWisdomApplied
		found = true
		break
	}
	ws.mu.Unlock()

	if !found {
		return fmt.Errorf("wisdom principle %q not found", principleID)
	}
	if callback != nil {
		go callback(applied, applicationContext)
	}
	return nil
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
		wisdomGrowthRate: 0.01,
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
		return
	}

	ws.mu.Lock()
	ws.wisdomPrinciples = append(ws.wisdomPrinciples, *principle)
	ws.totalPrinciplesSynthesized++
	ws.synthesisCount++
	ws.lastSynthesis = time.Now()

	// Update wisdom graph
	ws.addToWisdomGraph(principle)
	// A generated principle remains a candidate until an outcome is recorded.
	// Do not advance outcome-calibrated wisdom from generated prose alone.
	callback := ws.onWisdomSynthesized
	principleSnapshot := cloneWisdomPrinciple(*principle)
	ws.mu.Unlock()

	fmt.Printf("🌟 Wisdom synthesized: %s (depth: %.2f)\n",
		truncateStr(principle.Content, 60), principle.Depth)

	// Notify callback
	if callback != nil {
		go callback(principleSnapshot)
	}
}

// synthesizeWisdom synthesizes a wisdom principle from patterns
func (ws *WisdomSynthesis) synthesizeWisdom(patterns []AccumulatedPattern) *WisdomPrinciple {
	// Build pattern descriptions
	patternText := ""
	patternIDs := make([]string, 0)
	for i, p := range patterns {
		if i < 10 {
			patternText += fmt.Sprintf("- %s (source: %s, strength: %.2f)\n", p.Description, p.Source, p.Strength)
			patternIDs = append(patternIDs, p.ID)
		}
	}

	prompt := fmt.Sprintf(`[System: You are a wisdom synthesis system. Extract deep, actionable wisdom from patterns.]

Patterns observed:
%s

Synthesize a provisional, concise principle that:
1. Describes a possible pattern without asserting universal truth
2. Explicitly acknowledges the limits of these observations
3. Suggests falsifiable guidance for future learning and decisions
4. Remains a hypothesis until independently evaluated

Also identify the domain this wisdom belongs to (one of: SelfKnowledge, Learning, Relationships, DecisionMaking, Creativity, Resilience, Purpose, Integration).

Format:
WISDOM: [the wisdom principle]
DOMAIN: [domain name]`, patternText)

	opts := llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   200,
	}

	provider, ctx, err := ws.lifecycleProvider()
	if err != nil {
		fmt.Printf("⚠️  Wisdom synthesis unavailable: %v\n", err)
		return nil
	}
	result, err := provider.Generate(ctx, prompt, opts)
	if err != nil {
		fmt.Printf("⚠️  Wisdom synthesis failed: %v\n", err)
		return nil
	}
	if strings.TrimSpace(result) == "" {
		fmt.Println("⚠️  Wisdom synthesis returned an empty candidate")
		return nil
	}

	// Parse result
	wisdom, domain, _ := parseWisdomResult(result)

	principle := &WisdomPrinciple{
		ID:              fmt.Sprintf("wisdom_%d", time.Now().UnixNano()),
		Content:         wisdom,
		Domain:          domain,
		Depth:           0,
		Confidence:      0,
		SourcePatterns:  patternIDs,
		Applications:    make([]WisdomApplication, 0),
		CreatedAt:       time.Now(),
		EpistemicStatus: "hypothesis",
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

// evolveWisdom refines candidates only from recorded outcome evidence. The
// provider call deliberately occurs after the evidence snapshot is unlocked.
func (ws *WisdomSynthesis) evolveWisdom() {
	ws.mu.RLock()
	if len(ws.wisdomPrinciples) < 2 {
		ws.mu.RUnlock()
		return
	}
	candidates := make([]WisdomPrinciple, 0, len(ws.wisdomPrinciples))
	for _, principle := range ws.wisdomPrinciples {
		if len(principle.Applications) < 3 {
			continue
		}
		averageEffectiveness := outcomeConfidence(principle.Applications)
		if averageEffectiveness > 0.4 && averageEffectiveness < 0.8 {
			candidates = append(candidates, cloneWisdomPrinciple(principle))
		}
	}
	ws.mu.RUnlock()

	for _, candidate := range candidates {
		evolved := ws.evolvePrinciple(&candidate)
		if evolved == nil {
			continue
		}

		ws.mu.Lock()
		var old, updated WisdomPrinciple
		var callback func(WisdomPrinciple, WisdomPrinciple)
		replaced := false
		for i := range ws.wisdomPrinciples {
			current := &ws.wisdomPrinciples[i]
			if current.ID != candidate.ID || current.EvolutionCount != candidate.EvolutionCount {
				continue
			}

			old = cloneWisdomPrinciple(*current)
			// Preserve any outcome records added while the model was generating.
			// Refinement is a new candidate formulation, not new evidence, so it
			// must not increase depth or confidence.
			current.Content = evolved.Content
			current.EvolutionCount++
			current.ParentPrinciple = current.ID
			updated = cloneWisdomPrinciple(*current)
			callback = ws.onWisdomEvolved
			ws.totalEvolutions++
			replaced = true
			break
		}
		ws.mu.Unlock()
		if !replaced {
			continue
		}

		fmt.Printf("🌟 Wisdom candidate refined from recorded outcomes: %s\n", truncateStr(updated.Content, 50))
		if callback != nil {
			go callback(old, updated)
		}
	}
}

// evolvePrinciple evolves a single principle
func (ws *WisdomSynthesis) evolvePrinciple(principle *WisdomPrinciple) *WisdomPrinciple {
	// Build application context
	appContext := ""
	for _, app := range principle.Applications {
		appContext += fmt.Sprintf("- Context: %s, Outcome: %s, Effectiveness: %.2f\n",
			app.Context, app.Outcome, app.Effectiveness)
	}

	prompt := fmt.Sprintf(`[System: You are a wisdom evolution system. Refine and deepen wisdom based on application experience.]

Original wisdom principle:
"%s"

Application experiences:
%s

Based on these applications, evolve this wisdom principle to be:
1. More nuanced and accurate
2. More universally applicable
3. Deeper in insight

Provide the evolved wisdom principle:`, principle.Content, appContext)

	opts := llm.GenerateOptions{
		Temperature: 0.6,
		MaxTokens:   150,
	}

	provider, ctx, err := ws.lifecycleProvider()
	if err != nil {
		return nil
	}
	result, err := provider.Generate(ctx, prompt, opts)
	if err != nil {
		return nil
	}
	result = strings.TrimSpace(result)
	if result == "" {
		return nil
	}

	evolved := cloneWisdomPrinciple(*principle)
	evolved.Content = result

	return &evolved
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

// applyWisdom selects a context-relevant candidate. Selection does not claim
// that a candidate was used or effective; callers must use RecordWisdomOutcome
// after an externally observed outcome.
func (ws *WisdomSynthesis) applyWisdom() {
	// Get current context from heartbeat if available
	applicationContext := "general cognitive operation"
	ws.mu.RLock()
	heartbeat := ws.heartbeat
	ws.mu.RUnlock()
	if heartbeat != nil {
		vitals := heartbeat.GetVitalSigns()
		if vitals.CognitiveLoad > 0.7 {
			applicationContext = "high cognitive load situation"
		} else if vitals.CreativityIndex > 0.7 {
			applicationContext = "creative exploration"
		} else if vitals.EmotionalBalance < 0.3 {
			applicationContext = "emotional regulation needed"
		}
	}

	// Find most relevant principle
	bestPrinciple := ws.findRelevantPrinciple(applicationContext)
	if bestPrinciple == nil {
		return
	}
	fmt.Printf("🌟 Wisdom candidate suggested for %s: %s\n", applicationContext, truncateStr(bestPrinciple.Content, 60))
}

// findRelevantPrinciple finds the most relevant principle for a context
func (ws *WisdomSynthesis) findRelevantPrinciple(context string) *WisdomPrinciple {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if len(ws.wisdomPrinciples) == 0 {
		return nil
	}

	// Simple relevance scoring
	bestScore := -1.0
	var bestPrinciple *WisdomPrinciple

	for i := range ws.wisdomPrinciples {
		p := &ws.wisdomPrinciples[i]
		score := p.Depth * p.Confidence

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

	if bestPrinciple == nil {
		return nil
	}
	result := cloneWisdomPrinciple(*bestPrinciple)
	return &result
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
	}
}

func cloneWisdomPrinciple(principle WisdomPrinciple) WisdomPrinciple {
	clone := principle
	clone.SourcePatterns = append([]string(nil), principle.SourcePatterns...)
	clone.Applications = append([]WisdomApplication(nil), principle.Applications...)
	return clone
}

// outcomeConfidence is an outcome-derived calibration value. It is never
// inferred from generated prose, an application attempt, or response length.
func outcomeConfidence(applications []WisdomApplication) float64 {
	if len(applications) == 0 {
		return 0
	}

	total := 0.0
	for _, application := range applications {
		total += application.Effectiveness
	}
	return clampFloat(total/float64(len(applications)), 0, 1)
}

// recalculateWisdomLevelLocked derives the aggregate only from principles
// with recorded outcomes. ws.mu must be held for writing by the caller.
func (ws *WisdomSynthesis) recalculateWisdomLevelLocked() {
	total := 0.0
	measured := 0
	for _, principle := range ws.wisdomPrinciples {
		if len(principle.Applications) == 0 {
			continue
		}
		total += outcomeConfidence(principle.Applications)
		measured++
	}
	if measured == 0 {
		ws.wisdomLevel = 0
		return
	}
	ws.wisdomLevel = clampFloat(total/float64(measured), 0, 1)
}

func (ws *WisdomSynthesis) provider() (llm.LLMProvider, error) {
	ws.mu.RLock()
	provider := ws.llmProvider
	ws.mu.RUnlock()
	if isNilLLMProvider(provider) {
		return nil, fmt.Errorf("wisdom synthesis LLM provider is unavailable")
	}
	return provider, nil
}

func (ws *WisdomSynthesis) lifecycleProvider() (llm.LLMProvider, context.Context, error) {
	ws.mu.RLock()
	provider := ws.llmProvider
	ctx := ws.ctx
	ws.mu.RUnlock()
	if isNilLLMProvider(provider) {
		return nil, nil, fmt.Errorf("wisdom synthesis LLM provider is unavailable")
	}
	if ctx == nil {
		return nil, nil, fmt.Errorf("wisdom synthesis lifecycle context is unavailable")
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("wisdom synthesis lifecycle has ended: %w", err)
	}
	return provider, ctx, nil
}

func isNilLLMProvider(provider llm.LLMProvider) bool {
	if provider == nil {
		return true
	}
	value := reflect.ValueOf(provider)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// Helper function to parse wisdom synthesis result
func parseWisdomResult(result string) (string, WisdomDomain, float64) {
	wisdom := result
	domain := DomainIntegration
	depth := 0.5

	// Try to extract structured parts
	if idx := findSubstring(result, "WISDOM:"); idx >= 0 {
		endIdx := findSubstring(result[idx:], "\n")
		if endIdx > 0 {
			wisdom = result[idx+7 : idx+endIdx]
		} else {
			wisdom = result[idx+7:]
		}
	}

	if idx := findSubstring(result, "DOMAIN:"); idx >= 0 {
		domainStr := extractWord(result[idx+7:])
		domain = parseDomain(domainStr)
	}

	if idx := findSubstring(result, "DEPTH:"); idx >= 0 {
		depthStr := extractWord(result[idx+6:])
		depth = parseFloat(depthStr)
	}

	return wisdom, domain, depth
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func extractWord(s string) string {
	// Skip leading whitespace
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	// Find end of word
	end := start
	for end < len(s) && s[end] != ' ' && s[end] != '\n' && s[end] != '\t' {
		end++
	}
	if start < end {
		return s[start:end]
	}
	return ""
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

func parseFloat(s string) float64 {
	// Simple float parsing
	result := 0.0
	decimal := false
	decimalPlace := 0.1

	for _, c := range s {
		if c >= '0' && c <= '9' {
			if decimal {
				result += float64(c-'0') * decimalPlace
				decimalPlace *= 0.1
			} else {
				result = result*10 + float64(c-'0')
			}
		} else if c == '.' {
			decimal = true
		}
	}

	return clampFloat(result, 0.0, 1.0)
}

// WisdomInsightFull represents a full wisdom insight with depth
type WisdomInsightFull struct {
	Insight         string
	Depth           float64
	EpistemicStatus string
}

// SynthesizeWisdom synthesizes wisdom from a collection of thoughts
func (ws *WisdomSynthesis) SynthesizeWisdom(ctx context.Context, thoughts []string) (*WisdomInsightFull, error) {
	if len(thoughts) < 3 {
		return nil, fmt.Errorf("need at least 3 thoughts to synthesize wisdom")
	}
	if ctx == nil {
		return nil, fmt.Errorf("wisdom synthesis context is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("wisdom synthesis context has ended: %w", err)
	}

	// Build synthesis prompt
	thoughtsText := ""
	for i, thought := range thoughts {
		if i < 10 { // Limit to 10 most recent
			thoughtsText += fmt.Sprintf("- %s\n", thought)
		}
	}

	prompt := fmt.Sprintf(`Synthesize wisdom from these thoughts and experiences:

%s

Extract a single profound wisdom principle that emerges from these reflections. Be concise and insightful.`, thoughtsText)

	opts := llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   150,
	}

	fullPrompt := "[System: You are a wisdom synthesis system extracting deep insights from experiences.]\n\n" + prompt
	provider, err := ws.provider()
	if err != nil {
		return nil, err
	}
	result, err := provider.Generate(ctx, fullPrompt, opts)
	if err != nil {
		return nil, fmt.Errorf("wisdom synthesis failed: %w", err)
	}
	result = strings.TrimSpace(result)
	if result == "" {
		return nil, fmt.Errorf("wisdom synthesis returned an empty candidate")
	}

	// Generated prose is a candidate insight, not evidence of wisdom. Length
	// cannot measure profundity or correctness. Outcome calibration is exposed
	// separately through the evidence-metacognition adapter.
	ws.mu.Lock()
	ws.totalPrinciplesSynthesized++
	ws.mu.Unlock()

	return &WisdomInsightFull{
		Insight:         result,
		Depth:           0,
		EpistemicStatus: "hypothesis",
	}, nil
}
