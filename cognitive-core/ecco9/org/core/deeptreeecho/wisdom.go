//go:build orgdte
// +build orgdte

package deeptreeecho

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// WisdomSynthesisEngine performs deep pattern synthesis to cultivate wisdom.
// It integrates insights from multiple cognitive subsystems into coherent
// wisdom principles that guide future behavior and decision-making.
type WisdomSynthesisEngine struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Identity integration
	identity *Identity

	// Wisdom repository
	principles []WisdomPrinciple
	graph      *WisdomGraph

	// Synthesis state
	synthesisDepth int
	lastSynthesis  time.Time
	synthesisCount uint64

	// Pattern accumulation
	pendingPatterns  []AccumulatedPattern
	patternThreshold int

	// Seven-dimensional wisdom metrics
	wisdomMetrics *WisdomMetrics

	// Wisdom level (0.0 to 1.0)
	wisdomLevel      float64
	wisdomGrowthRate float64

	// Callbacks
	onWisdomSynthesized func(principle WisdomPrinciple)
	onWisdomApplied     func(principle WisdomPrinciple, ctx string)
	onWisdomEvolved     func(old, evolved WisdomPrinciple)

	// Counters
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

// WisdomMetrics tracks seven dimensions of wisdom
type WisdomMetrics struct {
	// Depth of understanding -- how profound are the principles
	Depth float64

	// Breadth of coverage -- how many domains are represented
	Breadth float64

	// Integration -- how well principles connect across domains
	Integration float64

	// Application -- how effectively wisdom is applied
	Application float64

	// Evolution -- how much wisdom has grown over time
	Evolution float64

	// Coherence -- internal consistency of the wisdom corpus
	Coherence float64

	// Resilience -- stability of wisdom under new experiences
	Resilience float64

	// Timestamp
	Timestamp time.Time
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

// NewWisdomSynthesisEngine creates a new wisdom synthesis engine
func NewWisdomSynthesisEngine(identity *Identity) *WisdomSynthesisEngine {
	ctx, cancel := context.WithCancel(context.Background())

	return &WisdomSynthesisEngine{
		ctx:        ctx,
		cancel:     cancel,
		identity:   identity,
		principles: make([]WisdomPrinciple, 0),
		graph: &WisdomGraph{
			Nodes:    make(map[string]*WisdomNode),
			Edges:    make([]WisdomEdge, 0),
			Clusters: make([]WisdomCluster, 0),
		},
		synthesisDepth:   3,
		pendingPatterns:  make([]AccumulatedPattern, 0),
		patternThreshold: 5,
		wisdomMetrics: &WisdomMetrics{
			Timestamp: time.Now(),
		},
		wisdomLevel:      0.0,
		wisdomGrowthRate: 0.01,
	}
}

// Start begins the wisdom synthesis system
func (ws *WisdomSynthesisEngine) Start() error {
	ws.mu.Lock()
	if ws.running {
		ws.mu.Unlock()
		return fmt.Errorf("wisdom synthesis already running")
	}
	ws.running = true
	ws.mu.Unlock()

	// Start synthesis loop
	go ws.synthesisLoop()

	// Start evolution loop
	go ws.evolutionLoop()

	// Start application loop
	go ws.applicationLoop()

	return nil
}

// Stop stops the wisdom synthesis system
func (ws *WisdomSynthesisEngine) Stop() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if !ws.running {
		return fmt.Errorf("wisdom synthesis not running")
	}

	ws.running = false
	ws.cancel()

	return nil
}

// AccumulatePattern adds a pattern for potential synthesis
func (ws *WisdomSynthesisEngine) AccumulatePattern(description string, source string, strength float64, tags []string) {
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
func (ws *WisdomSynthesisEngine) synthesisLoop() {
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
func (ws *WisdomSynthesisEngine) triggerSynthesis() {
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
	ws.principles = append(ws.principles, *principle)
	ws.totalPrinciplesSynthesized++
	ws.synthesisCount++
	ws.lastSynthesis = time.Now()

	// Update wisdom graph
	ws.addToWisdomGraph(principle)

	// Update wisdom level
	ws.wisdomLevel = math.Min(ws.wisdomLevel+ws.wisdomGrowthRate, 1.0)

	// Recalculate metrics
	ws.recalculateMetrics()
	ws.mu.Unlock()

	// Integrate with identity memory if available
	if ws.identity != nil && ws.identity.Memory != nil {
		ws.identity.Memory.Nodes[principle.ID] = &MemoryNode{
			ID:        principle.ID,
			Content:   principle.Content,
			Strength:  principle.Confidence,
			Timestamp: principle.CreatedAt,
			Resonance: principle.Depth,
		}
	}

	// Notify callback
	if ws.onWisdomSynthesized != nil {
		go ws.onWisdomSynthesized(*principle)
	}
}

// synthesizeWisdom synthesizes a wisdom principle from patterns
func (ws *WisdomSynthesisEngine) synthesizeWisdom(patterns []AccumulatedPattern) *WisdomPrinciple {
	if len(patterns) == 0 {
		return nil
	}

	// Analyze pattern tags to determine domain
	domainScores := make(map[WisdomDomain]float64)
	patternIDs := make([]string, 0, len(patterns))
	totalStrength := 0.0

	for _, p := range patterns {
		patternIDs = append(patternIDs, p.ID)
		totalStrength += p.Strength

		for _, tag := range p.Tags {
			tag = strings.ToLower(tag)
			switch {
			case strings.Contains(tag, "self") || strings.Contains(tag, "identity"):
				domainScores[DomainSelfKnowledge] += p.Strength
			case strings.Contains(tag, "learn"):
				domainScores[DomainLearning] += p.Strength
			case strings.Contains(tag, "relation") || strings.Contains(tag, "social"):
				domainScores[DomainRelationships] += p.Strength
			case strings.Contains(tag, "decision") || strings.Contains(tag, "choice"):
				domainScores[DomainDecisionMaking] += p.Strength
			case strings.Contains(tag, "creat") || strings.Contains(tag, "novel"):
				domainScores[DomainCreativity] += p.Strength
			case strings.Contains(tag, "resilien") || strings.Contains(tag, "adapt"):
				domainScores[DomainResilience] += p.Strength
			case strings.Contains(tag, "purpose") || strings.Contains(tag, "meaning"):
				domainScores[DomainPurpose] += p.Strength
			default:
				domainScores[DomainIntegration] += p.Strength
			}
		}
	}

	// Find dominant domain
	bestDomain := DomainIntegration
	bestDomainScore := 0.0
	for domain, score := range domainScores {
		if score > bestDomainScore {
			bestDomainScore = score
			bestDomain = domain
		}
	}

	// Synthesize content from pattern descriptions
	contentParts := make([]string, 0, len(patterns))
	for _, p := range patterns {
		if len(contentParts) < 5 {
			contentParts = append(contentParts, p.Description)
		}
	}
	content := fmt.Sprintf("[%s] Synthesis from %d patterns: %s",
		bestDomain.String(), len(patterns), strings.Join(contentParts, "; "))

	// Calculate depth based on pattern strength and count
	avgStrength := totalStrength / float64(len(patterns))
	depth := math.Min(avgStrength*0.7+float64(len(patterns))*0.03, 1.0)

	principle := &WisdomPrinciple{
		ID:             fmt.Sprintf("wisdom_%d", time.Now().UnixNano()),
		Content:        content,
		Domain:         bestDomain,
		Depth:          depth,
		Confidence:     0.7,
		SourcePatterns: patternIDs,
		Applications:   make([]WisdomApplication, 0),
		CreatedAt:      time.Now(),
	}

	return principle
}

// addToWisdomGraph adds a principle to the wisdom graph
func (ws *WisdomSynthesisEngine) addToWisdomGraph(principle *WisdomPrinciple) {
	ws.graph.Nodes[principle.ID] = &WisdomNode{
		PrincipleID: principle.ID,
		Centrality:  0.0,
		Connections: 0,
	}

	// Find related principles and create edges
	for _, existing := range ws.principles {
		if existing.ID == principle.ID {
			continue
		}

		relation, strength := ws.findRelationship(principle, &existing)
		if strength > 0.3 {
			edge := WisdomEdge{
				FromID:       principle.ID,
				ToID:         existing.ID,
				Relationship: relation,
				Strength:     strength,
			}
			ws.graph.Edges = append(ws.graph.Edges, edge)

			ws.graph.Nodes[principle.ID].Connections++
			if node, exists := ws.graph.Nodes[existing.ID]; exists {
				node.Connections++
			}
		}
	}

	ws.updateCentrality()
}

// findRelationship determines the relationship between two principles
func (ws *WisdomSynthesisEngine) findRelationship(p1, p2 *WisdomPrinciple) (WisdomRelation, float64) {
	if p1.Domain == p2.Domain {
		if p1.Depth > p2.Depth {
			return RelationExtends, 0.6
		}
		return RelationSupports, 0.5
	}

	// Complementary domain pairs
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
func (ws *WisdomSynthesisEngine) updateCentrality() {
	totalConnections := 0
	for _, node := range ws.graph.Nodes {
		totalConnections += node.Connections
	}

	if totalConnections == 0 {
		return
	}

	for _, node := range ws.graph.Nodes {
		node.Centrality = float64(node.Connections) / float64(totalConnections)
	}
}

// evolutionLoop periodically evolves wisdom principles
func (ws *WisdomSynthesisEngine) evolutionLoop() {
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

// evolveWisdom evolves existing wisdom principles based on application experience
func (ws *WisdomSynthesisEngine) evolveWisdom() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if len(ws.principles) < 2 {
		return
	}

	for i, principle := range ws.principles {
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
				old := ws.principles[i]
				ws.principles[i] = *evolved
				ws.totalEvolutions++

				if ws.onWisdomEvolved != nil {
					go ws.onWisdomEvolved(old, *evolved)
				}
			}
		}
	}
}

// evolvePrinciple evolves a single principle by deepening it
func (ws *WisdomSynthesisEngine) evolvePrinciple(principle *WisdomPrinciple) *WisdomPrinciple {
	// Build evolved content from application experiences
	improvements := make([]string, 0)
	for _, app := range principle.Applications {
		if app.Effectiveness > 0.6 {
			improvements = append(improvements, app.Outcome)
		}
	}

	evolvedContent := principle.Content
	if len(improvements) > 0 {
		evolvedContent = fmt.Sprintf("%s [Evolved through %d applications: %s]",
			principle.Content, len(improvements), strings.Join(improvements, "; "))
	}

	evolved := &WisdomPrinciple{
		ID:              principle.ID,
		Content:         evolvedContent,
		Domain:          principle.Domain,
		Depth:           math.Min(principle.Depth+0.1, 1.0),
		Confidence:      math.Min(principle.Confidence+0.05, 1.0),
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
func (ws *WisdomSynthesisEngine) applicationLoop() {
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
func (ws *WisdomSynthesisEngine) applyWisdom() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if len(ws.principles) == 0 {
		return
	}

	// Derive context from identity state
	contextStr := "general cognitive operation"
	if ws.identity != nil {
		if ws.identity.Coherence > 0.8 {
			contextStr = "high coherence creative state"
		} else if ws.identity.Coherence < 0.4 {
			contextStr = "low coherence requiring stabilization"
		}
		if ws.identity.EmotionalState != nil && ws.identity.EmotionalState.Intensity > 0.7 {
			contextStr = "heightened emotional processing"
		}
	}

	bestPrinciple := ws.findRelevantPrinciple(contextStr)
	if bestPrinciple == nil {
		return
	}

	application := WisdomApplication{
		Context:       contextStr,
		Outcome:       "Applied to current cognitive state",
		Effectiveness: 0.7,
		Timestamp:     time.Now(),
	}

	for i := range ws.principles {
		if ws.principles[i].ID == bestPrinciple.ID {
			ws.principles[i].Applications = append(ws.principles[i].Applications, application)
			ws.principles[i].ApplyCount++
			ws.principles[i].LastApplied = time.Now()
			break
		}
	}

	ws.totalApplications++

	if ws.onWisdomApplied != nil {
		go ws.onWisdomApplied(*bestPrinciple, contextStr)
	}
}

// findRelevantPrinciple finds the most relevant principle for a context
func (ws *WisdomSynthesisEngine) findRelevantPrinciple(contextStr string) *WisdomPrinciple {
	if len(ws.principles) == 0 {
		return nil
	}

	bestScore := 0.0
	var bestPrinciple *WisdomPrinciple

	for i := range ws.principles {
		p := &ws.principles[i]
		score := p.Depth * p.Confidence

		// Boost score based on domain relevance
		lower := strings.ToLower(contextStr)
		if strings.Contains(lower, "cognitive") && p.Domain == DomainSelfKnowledge {
			score *= 1.5
		}
		if strings.Contains(lower, "creative") && p.Domain == DomainCreativity {
			score *= 1.5
		}
		if strings.Contains(lower, "emotional") && p.Domain == DomainResilience {
			score *= 1.5
		}
		if strings.Contains(lower, "coherence") && p.Domain == DomainIntegration {
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

// recalculateMetrics updates the seven-dimensional wisdom metrics
func (ws *WisdomSynthesisEngine) recalculateMetrics() {
	metrics := &WisdomMetrics{
		Timestamp: time.Now(),
	}

	if len(ws.principles) == 0 {
		ws.wisdomMetrics = metrics
		return
	}

	// Depth -- average depth of all principles
	totalDepth := 0.0
	for _, p := range ws.principles {
		totalDepth += p.Depth
	}
	metrics.Depth = totalDepth / float64(len(ws.principles))

	// Breadth -- number of distinct domains covered
	domains := make(map[WisdomDomain]bool)
	for _, p := range ws.principles {
		domains[p.Domain] = true
	}
	metrics.Breadth = float64(len(domains)) / 8.0 // 8 total domains

	// Integration -- graph connectivity
	if len(ws.graph.Nodes) > 1 {
		totalEdges := float64(len(ws.graph.Edges))
		maxEdges := float64(len(ws.graph.Nodes) * (len(ws.graph.Nodes) - 1) / 2)
		if maxEdges > 0 {
			metrics.Integration = math.Min(totalEdges/maxEdges, 1.0)
		}
	}

	// Application -- ratio of applied principles
	appliedCount := 0
	for _, p := range ws.principles {
		if p.ApplyCount > 0 {
			appliedCount++
		}
	}
	if len(ws.principles) > 0 {
		metrics.Application = float64(appliedCount) / float64(len(ws.principles))
	}

	// Evolution -- average evolution count
	totalEvolution := 0
	for _, p := range ws.principles {
		totalEvolution += p.EvolutionCount
	}
	metrics.Evolution = math.Min(float64(totalEvolution)/float64(len(ws.principles)), 1.0)

	// Coherence -- average confidence
	totalConfidence := 0.0
	for _, p := range ws.principles {
		totalConfidence += p.Confidence
	}
	metrics.Coherence = totalConfidence / float64(len(ws.principles))

	// Resilience -- stability metric based on evolution success
	if ws.totalEvolutions > 0 {
		metrics.Resilience = math.Min(float64(ws.totalEvolutions)/float64(ws.totalPrinciplesSynthesized+1), 1.0)
	}

	ws.wisdomMetrics = metrics
}

// SetIdentity sets the identity integration
func (ws *WisdomSynthesisEngine) SetIdentity(identity *Identity) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.identity = identity
}

// SetCallbacks sets callback functions
func (ws *WisdomSynthesisEngine) SetCallbacks(
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
func (ws *WisdomSynthesisEngine) GetWisdomPrinciples() []WisdomPrinciple {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	result := make([]WisdomPrinciple, len(ws.principles))
	copy(result, ws.principles)
	return result
}

// GetWisdomLevel returns the current wisdom level
func (ws *WisdomSynthesisEngine) GetWisdomLevel() float64 {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.wisdomLevel
}

// GetWisdomMetrics returns the seven-dimensional wisdom metrics
func (ws *WisdomSynthesisEngine) GetWisdomMetrics() *WisdomMetrics {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	metrics := *ws.wisdomMetrics
	return &metrics
}

// GetMetrics returns synthesis metrics as a map
func (ws *WisdomSynthesisEngine) GetMetrics() map[string]interface{} {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	return map[string]interface{}{
		"wisdom_level":        ws.wisdomLevel,
		"principles_count":    len(ws.principles),
		"total_synthesized":   ws.totalPrinciplesSynthesized,
		"total_applications":  ws.totalApplications,
		"total_evolutions":    ws.totalEvolutions,
		"pending_patterns":    len(ws.pendingPatterns),
		"graph_nodes":         len(ws.graph.Nodes),
		"graph_edges":         len(ws.graph.Edges),
		"running":             ws.running,
		"metrics_depth":       ws.wisdomMetrics.Depth,
		"metrics_breadth":     ws.wisdomMetrics.Breadth,
		"metrics_integration": ws.wisdomMetrics.Integration,
		"metrics_application": ws.wisdomMetrics.Application,
		"metrics_evolution":   ws.wisdomMetrics.Evolution,
		"metrics_coherence":   ws.wisdomMetrics.Coherence,
		"metrics_resilience":  ws.wisdomMetrics.Resilience,
	}
}
