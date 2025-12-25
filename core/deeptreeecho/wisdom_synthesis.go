package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// WisdomSynthesis performs deep pattern synthesis to cultivate wisdom
// It integrates insights from multiple cognitive subsystems into coherent
// wisdom principles that guide future behavior and decision-making
type WisdomSynthesis struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc

	// LLM provider
	llmProvider     llm.LLMProvider

	// Wisdom repository
	wisdomPrinciples []WisdomPrinciple
	wisdomGraph      *WisdomGraph
	
	// Synthesis state
	synthesisDepth   int
	lastSynthesis    time.Time
	synthesisCount   uint64
	
	// Pattern accumulation
	pendingPatterns  []AccumulatedPattern
	patternThreshold int
	
	// Wisdom levels
	wisdomLevel      float64
	wisdomGrowthRate float64
	
	// Integration with other systems
	echoDream        *EchoDreamKnowledgeIntegration
	heartbeat        *AutonomousHeartbeat
	
	// Callbacks
	onWisdomSynthesized func(principle WisdomPrinciple)
	onWisdomApplied     func(principle WisdomPrinciple, context string)
	onWisdomEvolved     func(old, new WisdomPrinciple)
	
	// Metrics
	totalPrinciplesSynthesized uint64
	totalApplications          uint64
	totalEvolutions            uint64
	
	// Running state
	running         bool
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
	Context     string
	Outcome     string
	Effectiveness float64
	Timestamp   time.Time
}

// WisdomGraph represents relationships between wisdom principles
type WisdomGraph struct {
	Nodes       map[string]*WisdomNode
	Edges       []WisdomEdge
	Clusters    []WisdomCluster
}

// WisdomNode represents a node in the wisdom graph
type WisdomNode struct {
	PrincipleID string
	Centrality  float64
	Connections int
}

// WisdomEdge represents a relationship between principles
type WisdomEdge struct {
	FromID      string
	ToID        string
	Relationship WisdomRelation
	Strength    float64
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
	ID          string
	Theme       string
	PrincipleIDs []string
	Coherence   float64
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
		Tags:        tags,
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

	// Update wisdom level
	ws.wisdomLevel = min(ws.wisdomLevel+ws.wisdomGrowthRate, 1.0)
	ws.mu.Unlock()

	fmt.Printf("🌟 Wisdom synthesized: %s (depth: %.2f)\n",
		truncateStr(principle.Content, 60), principle.Depth)

	// Notify callback
	if ws.onWisdomSynthesized != nil {
		go ws.onWisdomSynthesized(*principle)
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

Synthesize these patterns into a single, profound wisdom principle that:
1. Captures the essential truth underlying these patterns
2. Is universally applicable across contexts
3. Provides guidance for future decisions and growth
4. Is expressed concisely but deeply

Also identify the domain this wisdom belongs to (one of: SelfKnowledge, Learning, Relationships, DecisionMaking, Creativity, Resilience, Purpose, Integration).

Format:
WISDOM: [the wisdom principle]
DOMAIN: [domain name]
DEPTH: [0.0-1.0 indicating profundity]`, patternText)

	opts := llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   200,
	}

	result, err := ws.llmProvider.Generate(context.Background(), prompt, opts)
	if err != nil {
		fmt.Printf("⚠️  Wisdom synthesis failed: %v\n", err)
		return nil
	}

	// Parse result
	wisdom, domain, depth := parseWisdomResult(result)

	principle := &WisdomPrinciple{
		ID:             fmt.Sprintf("wisdom_%d", time.Now().UnixNano()),
		Content:        wisdom,
		Domain:         domain,
		Depth:          depth,
		Confidence:     0.7,
		SourcePatterns: patternIDs,
		Applications:   make([]WisdomApplication, 0),
		CreatedAt:      time.Now(),
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
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if len(ws.wisdomPrinciples) < 2 {
		return
	}

	// Find principles that could be evolved
	for i, principle := range ws.wisdomPrinciples {
		// Check if principle has been applied enough to warrant evolution
		if principle.ApplyCount < 3 {
			continue
		}

		// Calculate effectiveness
		avgEffectiveness := 0.0
		for _, app := range principle.Applications {
			avgEffectiveness += app.Effectiveness
		}
		avgEffectiveness /= float64(len(principle.Applications))

		// Evolve if effectiveness is moderate (room for improvement)
		if avgEffectiveness > 0.4 && avgEffectiveness < 0.8 {
			evolved := ws.evolvePrinciple(&principle)
			if evolved != nil {
				ws.wisdomPrinciples[i] = *evolved
				ws.totalEvolutions++

				fmt.Printf("🌟 Wisdom evolved: %s\n", truncateStr(evolved.Content, 50))

				if ws.onWisdomEvolved != nil {
					go ws.onWisdomEvolved(principle, *evolved)
				}
			}
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

	result, err := ws.llmProvider.Generate(context.Background(), prompt, opts)
	if err != nil {
		return nil
	}

	evolved := &WisdomPrinciple{
		ID:              principle.ID,
		Content:         result,
		Domain:          principle.Domain,
		Depth:           min(principle.Depth+0.1, 1.0),
		Confidence:      min(principle.Confidence+0.05, 1.0),
		SourcePatterns:  principle.SourcePatterns,
		Applications:    principle.Applications,
		CreatedAt:       principle.CreatedAt,
		LastApplied:     principle.LastApplied,
		ApplyCount:      principle.ApplyCount,
		EvolutionCount:  principle.EvolutionCount + 1,
		ParentPrinciple: principle.ID,
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

// applyWisdom applies relevant wisdom to current context
func (ws *WisdomSynthesis) applyWisdom() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if len(ws.wisdomPrinciples) == 0 {
		return
	}

	// Get current context from heartbeat if available
	context := "general cognitive operation"
	if ws.heartbeat != nil {
		vitals := ws.heartbeat.GetVitalSigns()
		if vitals.CognitiveLoad > 0.7 {
			context = "high cognitive load situation"
		} else if vitals.CreativityIndex > 0.7 {
			context = "creative exploration"
		} else if vitals.EmotionalBalance < 0.3 {
			context = "emotional regulation needed"
		}
	}

	// Find most relevant principle
	bestPrinciple := ws.findRelevantPrinciple(context)
	if bestPrinciple == nil {
		return
	}

	// Record application
	application := WisdomApplication{
		Context:       context,
		Outcome:       "Applied to current state",
		Effectiveness: 0.7,
		Timestamp:     time.Now(),
	}

	// Update principle
	for i := range ws.wisdomPrinciples {
		if ws.wisdomPrinciples[i].ID == bestPrinciple.ID {
			ws.wisdomPrinciples[i].Applications = append(ws.wisdomPrinciples[i].Applications, application)
			ws.wisdomPrinciples[i].ApplyCount++
			ws.wisdomPrinciples[i].LastApplied = time.Now()
			break
		}
	}

	ws.totalApplications++

	// Notify callback
	if ws.onWisdomApplied != nil {
		go ws.onWisdomApplied(*bestPrinciple, context)
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
	copy(result, ws.wisdomPrinciples)
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
		"wisdom_level":           ws.wisdomLevel,
		"principles_count":       len(ws.wisdomPrinciples),
		"total_synthesized":      ws.totalPrinciplesSynthesized,
		"total_applications":     ws.totalApplications,
		"total_evolutions":       ws.totalEvolutions,
		"pending_patterns":       len(ws.pendingPatterns),
		"graph_nodes":            len(ws.wisdomGraph.Nodes),
		"graph_edges":            len(ws.wisdomGraph.Edges),
		"running":                ws.running,
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
	Insight string
	Depth   float64
}

// SynthesizeWisdom synthesizes wisdom from a collection of thoughts
func (ws *WisdomSynthesis) SynthesizeWisdom(ctx context.Context, thoughts []string) (*WisdomInsightFull, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	
	if len(thoughts) < 3 {
		return nil, fmt.Errorf("need at least 3 thoughts to synthesize wisdom")
	}
	
	// Build synthesis prompt
	thoughtsText := ""
	for i, thought := range thoughts {
		if i < 10 {  // Limit to 10 most recent
			thoughtsText += fmt.Sprintf("- %s\n", thought)
		}
	}
	
	prompt := fmt.Sprintf(`Synthesize wisdom from these thoughts and experiences:

%s

Extract a single profound wisdom principle that emerges from these reflections. Be concise and insightful.`, thoughtsText)
	
	opts := llm.GenerateOptions{
		Temperature:  0.7,
		MaxTokens:    150,
	}
	
	fullPrompt := "[System: You are a wisdom synthesis system extracting deep insights from experiences.]\n\n" + prompt
	result, err := ws.llmProvider.Generate(ctx, fullPrompt, opts)
	if err != nil {
		return nil, fmt.Errorf("wisdom synthesis failed: %w", err)
	}
	
	// Calculate depth based on length and complexity (simplified)
	depth := 0.5 + float64(len(result))/500.0
	if depth > 1.0 {
		depth = 1.0
	}
	
	ws.totalPrinciplesSynthesized++
	
	return &WisdomInsightFull{
		Insight: result,
		Depth:   depth,
	}, nil
}
