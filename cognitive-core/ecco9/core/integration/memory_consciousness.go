package integration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EchoCog/echollama/core/consciousness"
	"github.com/EchoCog/echollama/core/memory"
	"github.com/google/uuid"
)

// MemoryConsciousnessIntegrator bridges the stream of consciousness with hypergraph memory
// It enables thoughts to query memory, store insights, and build persistent knowledge
type MemoryConsciousnessIntegrator struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Core components
	consciousness *consciousness.StreamOfConsciousnessLLM
	memory        *memory.HypergraphMemory

	// Integration state
	activeQueries  map[string]*MemoryQuery
	recentInsights []*StoredInsight
	activationMap  map[string]float64 // node ID -> activation level

	// Configuration
	queryThreshold  float64 // Minimum confidence to trigger memory query
	storeThreshold  float64 // Minimum confidence to store as memory
	activationDecay float64 // How fast activation decays

	// Metrics
	queriesExecuted uint64
	insightsStored  uint64
	patternsFound   uint64

	running bool
}

// MemoryQuery represents a query from consciousness to memory
type MemoryQuery struct {
	ID        string
	Timestamp time.Time
	ThoughtID string
	QueryType QueryType
	Keywords  []string
	Context   map[string]interface{}
	Results   []*memory.MemoryNode
	Relevance float64
}

// QueryType defines different types of memory queries
type QueryType int

const (
	QueryTypeRecall      QueryType = iota // Recall specific facts
	QueryTypePattern                      // Find patterns
	QueryTypeAssociation                  // Find associations
	QueryTypeEpisodic                     // Recall experiences
	QueryTypeProcedural                   // Recall how to do something
)

// StoredInsight represents an insight stored in memory
type StoredInsight struct {
	ID          string
	ThoughtID   string
	NodeID      string
	Content     string
	Timestamp   time.Time
	Importance  float64
	Connections []string
}

// NewMemoryConsciousnessIntegrator creates a new integrator
func NewMemoryConsciousnessIntegrator(
	consciousness *consciousness.StreamOfConsciousnessLLM,
	memory *memory.HypergraphMemory,
) *MemoryConsciousnessIntegrator {
	ctx, cancel := context.WithCancel(context.Background())

	return &MemoryConsciousnessIntegrator{
		ctx:             ctx,
		cancel:          cancel,
		consciousness:   consciousness,
		memory:          memory,
		activeQueries:   make(map[string]*MemoryQuery),
		recentInsights:  make([]*StoredInsight, 0),
		activationMap:   make(map[string]float64),
		queryThreshold:  0.6,
		storeThreshold:  0.7,
		activationDecay: 0.95,
	}
}

// Start begins the integration process
func (mci *MemoryConsciousnessIntegrator) Start() error {
	mci.mu.Lock()
	if mci.running {
		mci.mu.Unlock()
		return fmt.Errorf("memory-consciousness integrator already running")
	}
	mci.running = true
	mci.mu.Unlock()

	// Start thought monitoring loop
	go mci.thoughtMonitoringLoop()

	// Start activation decay loop
	go mci.activationDecayLoop()

	// Start pattern recognition loop
	go mci.patternRecognitionLoop()

	return nil
}

// Stop halts the integration process
func (mci *MemoryConsciousnessIntegrator) Stop() {
	mci.mu.Lock()
	mci.running = false
	mci.mu.Unlock()

	mci.cancel()
}

// thoughtMonitoringLoop monitors thoughts and triggers memory operations
func (mci *MemoryConsciousnessIntegrator) thoughtMonitoringLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mci.ctx.Done():
			return
		case <-ticker.C:
			mci.processRecentThoughts()
		}
	}
}

// processRecentThoughts processes recent thoughts for memory integration
func (mci *MemoryConsciousnessIntegrator) processRecentThoughts() {
	// Get recent thoughts from consciousness stream
	thoughts := mci.consciousness.GetRecentThoughts(5)

	for _, thought := range thoughts {
		// Check if thought should trigger memory query
		if mci.shouldQueryMemory(thought) {
			mci.queryMemoryForThought(thought)
		}

		// Check if thought should be stored as insight
		if mci.shouldStoreAsInsight(thought) {
			mci.storeThoughtAsInsight(thought)
		}

		// Update activation based on thought
		mci.updateActivationFromThought(thought)
	}
}

// shouldQueryMemory determines if a thought should trigger a memory query
func (mci *MemoryConsciousnessIntegrator) shouldQueryMemory(thought interface{}) bool {
	// Type assertion to get thought details
	// In real implementation, would check thought type, confidence, and content
	// For now, query for reflection, question, and connection thoughts
	return true // Simplified for initial implementation
}

// queryMemoryForThought queries memory based on thought content
func (mci *MemoryConsciousnessIntegrator) queryMemoryForThought(thought interface{}) {
	// Extract keywords from thought
	keywords := mci.extractKeywords(thought)

	// Determine query type based on thought type
	queryType := mci.determineQueryType(thought)

	// Create query
	query := &MemoryQuery{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		QueryType: queryType,
		Keywords:  keywords,
		Results:   make([]*memory.MemoryNode, 0),
	}

	// Execute query based on type
	switch queryType {
	case QueryTypeRecall:
		query.Results = mci.recallMemory(keywords)
	case QueryTypePattern:
		query.Results = mci.findPatterns(keywords)
	case QueryTypeAssociation:
		query.Results = mci.findAssociations(keywords)
	case QueryTypeEpisodic:
		query.Results = mci.recallEpisodes(keywords)
	case QueryTypeProcedural:
		query.Results = mci.recallProcedures(keywords)
	}

	// Store query
	mci.mu.Lock()
	mci.activeQueries[query.ID] = query
	mci.queriesExecuted++
	mci.mu.Unlock()

	// Inject results back into consciousness if relevant
	if len(query.Results) > 0 {
		mci.injectMemoryIntoConsciousness(query.Results)
	}
}

// shouldStoreAsInsight determines if a thought should be stored as memory
func (mci *MemoryConsciousnessIntegrator) shouldStoreAsInsight(thought interface{}) bool {
	// Store insights, realizations, and important connections
	// In real implementation, would check thought type and confidence
	return false // Simplified - only store specific thought types
}

// storeThoughtAsInsight stores a thought as a memory node
func (mci *MemoryConsciousnessIntegrator) storeThoughtAsInsight(thought interface{}) {
	// Create memory node from thought
	node := &memory.MemoryNode{
		ID:        uuid.New().String(),
		Type:      memory.NodeThought,
		Content:   mci.extractContent(thought),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"source":     "consciousness",
			"thought_id": mci.extractThoughtID(thought),
		},
	}

	// Add to hypergraph
	if err := mci.memory.AddNode(node); err != nil {
		// Log error but continue
		return
	}

	// Create insight record
	insight := &StoredInsight{
		ID:         uuid.New().String(),
		NodeID:     node.ID,
		Content:    node.Content,
		Timestamp:  time.Now(),
		Importance: 0.8,
	}

	// Store insight
	mci.mu.Lock()
	mci.recentInsights = append(mci.recentInsights, insight)
	mci.insightsStored++
	mci.mu.Unlock()
}

// updateActivationFromThought updates node activation based on thought content
func (mci *MemoryConsciousnessIntegrator) updateActivationFromThought(thought interface{}) {
	// Extract concepts from thought
	concepts := mci.extractKeywords(thought)

	// Find matching nodes
	for _, concept := range concepts {
		// Search for nodes matching this concept
		// In real implementation, would use semantic search
		nodes := mci.findNodesForConcept(concept)

		// Increase activation for matching nodes
		mci.mu.Lock()
		for _, node := range nodes {
			currentActivation := mci.activationMap[node.ID]
			mci.activationMap[node.ID] = min(1.0, currentActivation+0.1)
		}
		mci.mu.Unlock()
	}
}

// activationDecayLoop gradually decays activation levels
func (mci *MemoryConsciousnessIntegrator) activationDecayLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mci.ctx.Done():
			return
		case <-ticker.C:
			mci.decayActivation()
		}
	}
}

// decayActivation applies decay to all activation levels
func (mci *MemoryConsciousnessIntegrator) decayActivation() {
	mci.mu.Lock()
	defer mci.mu.Unlock()

	for nodeID, activation := range mci.activationMap {
		newActivation := activation * mci.activationDecay
		if newActivation < 0.01 {
			delete(mci.activationMap, nodeID)
		} else {
			mci.activationMap[nodeID] = newActivation
		}
	}
}

// patternRecognitionLoop looks for emerging patterns in activation
func (mci *MemoryConsciousnessIntegrator) patternRecognitionLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mci.ctx.Done():
			return
		case <-ticker.C:
			mci.recognizePatterns()
		}
	}
}

// recognizePatterns identifies patterns in highly activated nodes
func (mci *MemoryConsciousnessIntegrator) recognizePatterns() {
	mci.mu.RLock()

	// Find highly activated nodes
	highlyActivated := make([]string, 0)
	for nodeID, activation := range mci.activationMap {
		if activation > 0.7 {
			highlyActivated = append(highlyActivated, nodeID)
		}
	}
	mci.mu.RUnlock()

	// If multiple nodes are highly activated, look for patterns
	if len(highlyActivated) >= 2 {
		// Find connections between activated nodes
		// This could reveal emerging patterns of thought
		mci.mu.Lock()
		mci.patternsFound++
		mci.mu.Unlock()
	}
}

// Helper methods for memory operations

func (mci *MemoryConsciousnessIntegrator) extractKeywords(thought interface{}) []string {
	// Simplified keyword extraction
	// In real implementation, would use NLP or LLM
	return []string{"wisdom", "pattern", "understanding"}
}

func (mci *MemoryConsciousnessIntegrator) determineQueryType(thought interface{}) QueryType {
	// Determine query type based on thought type
	return QueryTypePattern
}

func (mci *MemoryConsciousnessIntegrator) extractContent(thought interface{}) string {
	// Extract content from thought
	return "Insight from consciousness"
}

func (mci *MemoryConsciousnessIntegrator) extractThoughtID(thought interface{}) string {
	// Extract thought ID
	return uuid.New().String()
}

func (mci *MemoryConsciousnessIntegrator) recallMemory(keywords []string) []*memory.MemoryNode {
	// Query memory for matching nodes
	return make([]*memory.MemoryNode, 0)
}

func (mci *MemoryConsciousnessIntegrator) findPatterns(keywords []string) []*memory.MemoryNode {
	// Find pattern nodes
	return make([]*memory.MemoryNode, 0)
}

func (mci *MemoryConsciousnessIntegrator) findAssociations(keywords []string) []*memory.MemoryNode {
	// Find associated nodes
	return make([]*memory.MemoryNode, 0)
}

func (mci *MemoryConsciousnessIntegrator) recallEpisodes(keywords []string) []*memory.MemoryNode {
	// Recall episodic memories
	return make([]*memory.MemoryNode, 0)
}

func (mci *MemoryConsciousnessIntegrator) recallProcedures(keywords []string) []*memory.MemoryNode {
	// Recall procedural memories
	return make([]*memory.MemoryNode, 0)
}

func (mci *MemoryConsciousnessIntegrator) findNodesForConcept(concept string) []*memory.MemoryNode {
	// Find nodes matching concept
	return make([]*memory.MemoryNode, 0)
}

func (mci *MemoryConsciousnessIntegrator) injectMemoryIntoConsciousness(nodes []*memory.MemoryNode) {
	// Inject memory results back into consciousness stream
	// This would add external thoughts based on retrieved memories
	for _, node := range nodes {
		content := fmt.Sprintf("Remembering: %s", node.Content)
		mci.consciousness.AddExternalThought(content)
	}
}

// GetMetrics returns integration metrics
func (mci *MemoryConsciousnessIntegrator) GetMetrics() map[string]interface{} {
	mci.mu.RLock()
	defer mci.mu.RUnlock()

	return map[string]interface{}{
		"queries_executed": mci.queriesExecuted,
		"insights_stored":  mci.insightsStored,
		"patterns_found":   mci.patternsFound,
		"active_nodes":     len(mci.activationMap),
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
