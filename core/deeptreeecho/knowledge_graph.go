package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// KnowledgeGraph provides a graph-based knowledge representation for Deep Tree Echo
// This is inspired by Cayley's quad store architecture for storing and querying linked data
type KnowledgeGraph struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc

	// Quad storage (Subject, Predicate, Object, Label)
	quads           map[string]*Quad
	quadsLock       sync.RWMutex

	// Indexes for fast lookups
	subjectIndex    map[string][]string // subject -> quad IDs
	predicateIndex  map[string][]string // predicate -> quad IDs
	objectIndex     map[string][]string // object -> quad IDs
	labelIndex      map[string][]string // label -> quad IDs

	// Node storage
	nodes           map[string]*GraphNode
	nodesLock       sync.RWMutex

	// Metrics
	totalQuads      uint64
	totalNodes      uint64
	totalQueries    uint64

	// Running state
	running         bool
}

// Quad represents a quad (Subject, Predicate, Object, Label) in the knowledge graph
type Quad struct {
	ID        string
	Subject   string
	Predicate string
	Object    string
	Label     string // Context/graph label
	CreatedAt time.Time
	Metadata  map[string]string
}

// GraphNode represents a node in the knowledge graph
type GraphNode struct {
	ID         string
	Value      string
	NodeType   string
	Properties map[string]interface{}
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// QuadDirection represents the direction of a quad relationship
type QuadDirection int

const (
	DirectionSubject QuadDirection = iota
	DirectionPredicate
	DirectionObject
	DirectionLabel
)

// QueryPath represents a path query through the graph
type QueryPath struct {
	Steps []QueryStep
}

// QueryStep represents a single step in a path query
type QueryStep struct {
	Direction QuadDirection
	Value     string
	Filter    func(*Quad) bool
}

// QueryResult represents a result from a graph query
type GraphQueryResult struct {
	Quads   []*Quad
	Nodes   []*GraphNode
	Paths   [][]string
	Count   int
}

// NewKnowledgeGraph creates a new knowledge graph
func NewKnowledgeGraph() *KnowledgeGraph {
	ctx, cancel := context.WithCancel(context.Background())

	return &KnowledgeGraph{
		ctx:            ctx,
		cancel:         cancel,
		quads:          make(map[string]*Quad),
		subjectIndex:   make(map[string][]string),
		predicateIndex: make(map[string][]string),
		objectIndex:    make(map[string][]string),
		labelIndex:     make(map[string][]string),
		nodes:          make(map[string]*GraphNode),
	}
}

// Start begins the knowledge graph system
func (kg *KnowledgeGraph) Start() error {
	kg.mu.Lock()
	defer kg.mu.Unlock()

	if kg.running {
		return fmt.Errorf("knowledge graph already running")
	}

	kg.running = true
	fmt.Println("🕸️ Knowledge Graph started")

	// Initialize default graph labels (contexts)
	kg.initializeDefaultLabels()

	return nil
}

// Stop gracefully stops the knowledge graph
func (kg *KnowledgeGraph) Stop() error {
	kg.mu.Lock()
	defer kg.mu.Unlock()

	if !kg.running {
		return fmt.Errorf("knowledge graph not running")
	}

	kg.cancel()
	kg.running = false
	fmt.Println("🕸️ Knowledge Graph stopped")

	return nil
}

// initializeDefaultLabels creates default graph labels
func (kg *KnowledgeGraph) initializeDefaultLabels() {
	// Create default labels for different knowledge domains
	defaultLabels := []string{
		"core",       // Core knowledge
		"episodic",   // Episodic memories
		"semantic",   // Semantic facts
		"procedural", // Procedural knowledge
		"wisdom",     // Wisdom insights
		"relations",  // Entity relationships
		"temporal",   // Temporal knowledge
	}

	for _, label := range defaultLabels {
		kg.labelIndex[label] = []string{}
	}

	fmt.Printf("   Initialized %d default graph labels\n", len(defaultLabels))
}

// AddQuad adds a quad to the knowledge graph
func (kg *KnowledgeGraph) AddQuad(subject, predicate, object, label string) (string, error) {
	kg.quadsLock.Lock()
	defer kg.quadsLock.Unlock()

	quadID := fmt.Sprintf("quad_%s", uuid.New().String()[:8])

	quad := &Quad{
		ID:        quadID,
		Subject:   subject,
		Predicate: predicate,
		Object:    object,
		Label:     label,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]string),
	}

	// Store quad
	kg.quads[quadID] = quad

	// Update indexes
	kg.subjectIndex[subject] = append(kg.subjectIndex[subject], quadID)
	kg.predicateIndex[predicate] = append(kg.predicateIndex[predicate], quadID)
	kg.objectIndex[object] = append(kg.objectIndex[object], quadID)
	kg.labelIndex[label] = append(kg.labelIndex[label], quadID)

	// Ensure nodes exist
	kg.ensureNode(subject, "entity")
	kg.ensureNode(object, "entity")
	kg.ensureNode(predicate, "predicate")

	kg.mu.Lock()
	kg.totalQuads++
	kg.mu.Unlock()

	return quadID, nil
}

// ensureNode ensures a node exists in the graph
func (kg *KnowledgeGraph) ensureNode(value, nodeType string) {
	kg.nodesLock.Lock()
	defer kg.nodesLock.Unlock()

	if _, exists := kg.nodes[value]; !exists {
		kg.nodes[value] = &GraphNode{
			ID:         fmt.Sprintf("node_%s", uuid.New().String()[:8]),
			Value:      value,
			NodeType:   nodeType,
			Properties: make(map[string]interface{}),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		kg.mu.Lock()
		kg.totalNodes++
		kg.mu.Unlock()
	}
}

// AddTriple adds a triple (quad with default label) to the knowledge graph
func (kg *KnowledgeGraph) AddTriple(subject, predicate, object string) (string, error) {
	return kg.AddQuad(subject, predicate, object, "core")
}

// RemoveQuad removes a quad from the knowledge graph
func (kg *KnowledgeGraph) RemoveQuad(quadID string) error {
	kg.quadsLock.Lock()
	defer kg.quadsLock.Unlock()

	quad, exists := kg.quads[quadID]
	if !exists {
		return fmt.Errorf("quad not found: %s", quadID)
	}

	// Remove from indexes
	kg.removeFromIndex(kg.subjectIndex, quad.Subject, quadID)
	kg.removeFromIndex(kg.predicateIndex, quad.Predicate, quadID)
	kg.removeFromIndex(kg.objectIndex, quad.Object, quadID)
	kg.removeFromIndex(kg.labelIndex, quad.Label, quadID)

	// Remove quad
	delete(kg.quads, quadID)

	return nil
}

// removeFromIndex removes a quad ID from an index
func (kg *KnowledgeGraph) removeFromIndex(index map[string][]string, key, quadID string) {
	ids := index[key]
	for i, id := range ids {
		if id == quadID {
			index[key] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
}

// QueryBySubject returns all quads with the given subject
func (kg *KnowledgeGraph) QueryBySubject(subject string) []*Quad {
	kg.quadsLock.RLock()
	defer kg.quadsLock.RUnlock()

	kg.mu.Lock()
	kg.totalQueries++
	kg.mu.Unlock()

	quadIDs := kg.subjectIndex[subject]
	return kg.getQuadsByIDs(quadIDs)
}

// QueryByPredicate returns all quads with the given predicate
func (kg *KnowledgeGraph) QueryByPredicate(predicate string) []*Quad {
	kg.quadsLock.RLock()
	defer kg.quadsLock.RUnlock()

	kg.mu.Lock()
	kg.totalQueries++
	kg.mu.Unlock()

	quadIDs := kg.predicateIndex[predicate]
	return kg.getQuadsByIDs(quadIDs)
}

// QueryByObject returns all quads with the given object
func (kg *KnowledgeGraph) QueryByObject(object string) []*Quad {
	kg.quadsLock.RLock()
	defer kg.quadsLock.RUnlock()

	kg.mu.Lock()
	kg.totalQueries++
	kg.mu.Unlock()

	quadIDs := kg.objectIndex[object]
	return kg.getQuadsByIDs(quadIDs)
}

// QueryByLabel returns all quads with the given label
func (kg *KnowledgeGraph) QueryByLabel(label string) []*Quad {
	kg.quadsLock.RLock()
	defer kg.quadsLock.RUnlock()

	kg.mu.Lock()
	kg.totalQueries++
	kg.mu.Unlock()

	quadIDs := kg.labelIndex[label]
	return kg.getQuadsByIDs(quadIDs)
}

// getQuadsByIDs returns quads for the given IDs
func (kg *KnowledgeGraph) getQuadsByIDs(quadIDs []string) []*Quad {
	quads := make([]*Quad, 0, len(quadIDs))
	for _, id := range quadIDs {
		if quad, exists := kg.quads[id]; exists {
			quads = append(quads, quad)
		}
	}
	return quads
}

// TraversePath traverses the graph following a path
func (kg *KnowledgeGraph) TraversePath(startNode string, predicates []string) []string {
	kg.quadsLock.RLock()
	defer kg.quadsLock.RUnlock()

	kg.mu.Lock()
	kg.totalQueries++
	kg.mu.Unlock()

	currentNodes := []string{startNode}

	for _, predicate := range predicates {
		nextNodes := make([]string, 0)
		for _, node := range currentNodes {
			// Find quads where this node is the subject and predicate matches
			quadIDs := kg.subjectIndex[node]
			for _, quadID := range quadIDs {
				if quad, exists := kg.quads[quadID]; exists {
					if quad.Predicate == predicate {
						nextNodes = append(nextNodes, quad.Object)
					}
				}
			}
		}
		currentNodes = nextNodes
		if len(currentNodes) == 0 {
			break
		}
	}

	return currentNodes
}

// FindRelated finds all nodes related to the given node
func (kg *KnowledgeGraph) FindRelated(node string, maxDepth int) map[string]int {
	kg.quadsLock.RLock()
	defer kg.quadsLock.RUnlock()

	kg.mu.Lock()
	kg.totalQueries++
	kg.mu.Unlock()

	related := make(map[string]int)
	visited := make(map[string]bool)
	queue := []struct {
		node  string
		depth int
	}{{node, 0}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.node] || current.depth > maxDepth {
			continue
		}
		visited[current.node] = true

		// Find outgoing relationships
		for _, quadID := range kg.subjectIndex[current.node] {
			if quad, exists := kg.quads[quadID]; exists {
				if !visited[quad.Object] {
					related[quad.Object] = current.depth + 1
					queue = append(queue, struct {
						node  string
						depth int
					}{quad.Object, current.depth + 1})
				}
			}
		}

		// Find incoming relationships
		for _, quadID := range kg.objectIndex[current.node] {
			if quad, exists := kg.quads[quadID]; exists {
				if !visited[quad.Subject] {
					related[quad.Subject] = current.depth + 1
					queue = append(queue, struct {
						node  string
						depth int
					}{quad.Subject, current.depth + 1})
				}
			}
		}
	}

	return related
}

// GetNode returns a node by value
func (kg *KnowledgeGraph) GetNode(value string) (*GraphNode, bool) {
	kg.nodesLock.RLock()
	defer kg.nodesLock.RUnlock()
	node, exists := kg.nodes[value]
	return node, exists
}

// SetNodeProperty sets a property on a node
func (kg *KnowledgeGraph) SetNodeProperty(nodeValue, key string, value interface{}) error {
	kg.nodesLock.Lock()
	defer kg.nodesLock.Unlock()

	node, exists := kg.nodes[nodeValue]
	if !exists {
		return fmt.Errorf("node not found: %s", nodeValue)
	}

	node.Properties[key] = value
	node.UpdatedAt = time.Now()
	return nil
}

// GetMetrics returns knowledge graph metrics
func (kg *KnowledgeGraph) GetMetrics() map[string]interface{} {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	return map[string]interface{}{
		"running":       kg.running,
		"total_quads":   kg.totalQuads,
		"total_nodes":   kg.totalNodes,
		"total_queries": kg.totalQueries,
	}
}

// ContributeToGestalt provides knowledge graph state for the global gestalt
func (kg *KnowledgeGraph) ContributeToGestalt() map[string]interface{} {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	kg.quadsLock.RLock()
	labelStats := make(map[string]int)
	for label, quadIDs := range kg.labelIndex {
		labelStats[label] = len(quadIDs)
	}
	kg.quadsLock.RUnlock()

	return map[string]interface{}{
		"running":       kg.running,
		"total_quads":   kg.totalQuads,
		"total_nodes":   kg.totalNodes,
		"label_stats":   labelStats,
	}
}

// StoreKnowledge stores a piece of knowledge as a quad
func (kg *KnowledgeGraph) StoreKnowledge(entity, relation, value, context string) (string, error) {
	return kg.AddQuad(entity, relation, value, context)
}

// QueryKnowledge queries knowledge about an entity
func (kg *KnowledgeGraph) QueryKnowledge(entity string) map[string][]string {
	quads := kg.QueryBySubject(entity)
	
	knowledge := make(map[string][]string)
	for _, quad := range quads {
		knowledge[quad.Predicate] = append(knowledge[quad.Predicate], quad.Object)
	}
	
	return knowledge
}

// InferRelationship infers a relationship between two entities
func (kg *KnowledgeGraph) InferRelationship(entity1, entity2 string) []string {
	// Find common predicates
	quads1 := kg.QueryBySubject(entity1)
	quads2 := kg.QueryBySubject(entity2)

	predicates1 := make(map[string]bool)
	for _, q := range quads1 {
		predicates1[q.Predicate] = true
	}

	commonPredicates := make([]string, 0)
	for _, q := range quads2 {
		if predicates1[q.Predicate] {
			commonPredicates = append(commonPredicates, q.Predicate)
		}
	}

	return commonPredicates
}

// FindPath finds a path between two nodes
func (kg *KnowledgeGraph) FindPath(start, end string, maxDepth int) []string {
	kg.quadsLock.RLock()
	defer kg.quadsLock.RUnlock()

	kg.mu.Lock()
	kg.totalQueries++
	kg.mu.Unlock()

	// BFS to find shortest path
	visited := make(map[string]string) // node -> previous node
	queue := []string{start}
	visited[start] = ""

	for len(queue) > 0 && len(visited) < maxDepth*10 {
		current := queue[0]
		queue = queue[1:]

		if current == end {
			// Reconstruct path
			path := []string{end}
			for node := end; visited[node] != ""; node = visited[node] {
				path = append([]string{visited[node]}, path...)
			}
			return path
		}

		// Explore neighbors
		for _, quadID := range kg.subjectIndex[current] {
			if quad, exists := kg.quads[quadID]; exists {
				if _, seen := visited[quad.Object]; !seen {
					visited[quad.Object] = current
					queue = append(queue, quad.Object)
				}
			}
		}
	}

	return nil // No path found
}
