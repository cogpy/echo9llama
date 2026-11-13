package memory

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

// HypergraphMemory implements a true hypergraph structure for knowledge representation
type HypergraphMemory struct {
	mu sync.RWMutex

	// Core graph structures
	nodes      map[string]*MemoryNode
	edges      map[string]*MemoryEdge
	hyperedges map[string]*HyperEdge

	// Adjacency lists for fast traversal
	outgoing map[string][]string // node ID -> edge IDs
	incoming map[string][]string // node ID -> edge IDs

	// Indices for fast lookup
	typeIndex map[NodeType][]string // type -> node IDs
	timeIndex []string              // sorted by creation time

	// Semantic embeddings (for future similarity search)
	embeddings map[string][]float64

	// Persistence layer
	persistence *SupabasePersistence
}

// NewHypergraphMemory creates a new hypergraph memory structure
func NewHypergraphMemory(persistence *SupabasePersistence) *HypergraphMemory {
	return &HypergraphMemory{
		nodes:       make(map[string]*MemoryNode),
		edges:       make(map[string]*MemoryEdge),
		hyperedges:  make(map[string]*HyperEdge),
		outgoing:    make(map[string][]string),
		incoming:    make(map[string][]string),
		typeIndex:   make(map[NodeType][]string),
		timeIndex:   make([]string, 0),
		embeddings:  make(map[string][]float64),
		persistence: persistence,
	}
}

// AddNode adds a node to the hypergraph
func (hg *HypergraphMemory) AddNode(node *MemoryNode) error {
	hg.mu.Lock()
	defer hg.mu.Unlock()

	if node.ID == "" {
		node.ID = uuid.New().String()
	}
	if node.CreatedAt.IsZero() {
		node.CreatedAt = time.Now()
	}
	node.UpdatedAt = time.Now()

	// Add to main storage
	hg.nodes[node.ID] = node

	// Update type index
	hg.typeIndex[node.Type] = append(hg.typeIndex[node.Type], node.ID)

	// Update time index (insert in sorted order)
	hg.timeIndex = append(hg.timeIndex, node.ID)

	// Initialize adjacency lists
	if _, exists := hg.outgoing[node.ID]; !exists {
		hg.outgoing[node.ID] = make([]string, 0)
	}
	if _, exists := hg.incoming[node.ID]; !exists {
		hg.incoming[node.ID] = make([]string, 0)
	}

	// Persist if available
	if hg.persistence != nil {
		if err := hg.persistence.StoreNode(node); err != nil {
			return fmt.Errorf("failed to persist node: %w", err)
		}
	}

	return nil
}

// AddEdge adds an edge to the hypergraph
func (hg *HypergraphMemory) AddEdge(edge *MemoryEdge) error {
	hg.mu.Lock()
	defer hg.mu.Unlock()

	if edge.ID == "" {
		edge.ID = uuid.New().String()
	}
	if edge.CreatedAt.IsZero() {
		edge.CreatedAt = time.Now()
	}

	// Validate nodes exist
	if _, exists := hg.nodes[edge.SourceID]; !exists {
		return fmt.Errorf("source node not found: %s", edge.SourceID)
	}
	if _, exists := hg.nodes[edge.TargetID]; !exists {
		return fmt.Errorf("target node not found: %s", edge.TargetID)
	}

	// Add to main storage
	hg.edges[edge.ID] = edge

	// Update adjacency lists
	hg.outgoing[edge.SourceID] = append(hg.outgoing[edge.SourceID], edge.ID)
	hg.incoming[edge.TargetID] = append(hg.incoming[edge.TargetID], edge.ID)

	// Persist if available
	if hg.persistence != nil {
		if err := hg.persistence.StoreEdge(edge); err != nil {
			return fmt.Errorf("failed to persist edge: %w", err)
		}
	}

	return nil
}

// AddHyperEdge adds a hyperedge (multi-way relationship) to the hypergraph
func (hg *HypergraphMemory) AddHyperEdge(hyperedge *HyperEdge) error {
	hg.mu.Lock()
	defer hg.mu.Unlock()

	if hyperedge.ID == "" {
		hyperedge.ID = uuid.New().String()
	}
	if hyperedge.CreatedAt.IsZero() {
		hyperedge.CreatedAt = time.Now()
	}

	// Validate all nodes exist
	for _, nodeID := range hyperedge.NodeIDs {
		if _, exists := hg.nodes[nodeID]; !exists {
			return fmt.Errorf("node not found in hyperedge: %s", nodeID)
		}
	}

	// Add to main storage
	hg.hyperedges[hyperedge.ID] = hyperedge

	return nil
}

// GetNode retrieves a node by ID
func (hg *HypergraphMemory) GetNode(id string) (*MemoryNode, error) {
	hg.mu.RLock()
	defer hg.mu.RUnlock()

	node, exists := hg.nodes[id]
	if !exists {
		return nil, fmt.Errorf("node not found: %s", id)
	}

	return node, nil
}

// GetNodesByType retrieves all nodes of a specific type
func (hg *HypergraphMemory) GetNodesByType(nodeType NodeType) []*MemoryNode {
	hg.mu.RLock()
	defer hg.mu.RUnlock()

	nodeIDs := hg.typeIndex[nodeType]
	nodes := make([]*MemoryNode, 0, len(nodeIDs))

	for _, id := range nodeIDs {
		if node, exists := hg.nodes[id]; exists {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// GetOutgoingEdges retrieves all edges originating from a node
func (hg *HypergraphMemory) GetOutgoingEdges(nodeID string) []*MemoryEdge {
	hg.mu.RLock()
	defer hg.mu.RUnlock()

	edgeIDs := hg.outgoing[nodeID]
	edges := make([]*MemoryEdge, 0, len(edgeIDs))

	for _, id := range edgeIDs {
		if edge, exists := hg.edges[id]; exists {
			edges = append(edges, edge)
		}
	}

	return edges
}

// GetIncomingEdges retrieves all edges pointing to a node
func (hg *HypergraphMemory) GetIncomingEdges(nodeID string) []*MemoryEdge {
	hg.mu.RLock()
	defer hg.mu.RUnlock()

	edgeIDs := hg.incoming[nodeID]
	edges := make([]*MemoryEdge, 0, len(edgeIDs))

	for _, id := range edgeIDs {
		if edge, exists := hg.edges[id]; exists {
			edges = append(edges, edge)
		}
	}

	return edges
}

// TraverseBFS performs breadth-first search from a starting node
func (hg *HypergraphMemory) TraverseBFS(startID string, maxDepth int, edgeTypes []EdgeType) ([]*MemoryNode, error) {
	hg.mu.RLock()
	defer hg.mu.RUnlock()

	if _, exists := hg.nodes[startID]; !exists {
		return nil, fmt.Errorf("start node not found: %s", startID)
	}

	visited := make(map[string]bool)
	result := make([]*MemoryNode, 0)

	type queueItem struct {
		nodeID string
		depth  int
	}

	queue := []queueItem{{nodeID: startID, depth: 0}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.nodeID] || current.depth > maxDepth {
			continue
		}

		visited[current.nodeID] = true

		if node, exists := hg.nodes[current.nodeID]; exists {
			result = append(result, node)
		}

		// Add neighbors
		for _, edgeID := range hg.outgoing[current.nodeID] {
			edge := hg.edges[edgeID]

			// Filter by edge type if specified
			if len(edgeTypes) > 0 {
				matchType := false
				for _, et := range edgeTypes {
					if edge.Type == et {
						matchType = true
						break
					}
				}
				if !matchType {
					continue
				}
			}

			queue = append(queue, queueItem{
				nodeID: edge.TargetID,
				depth:  current.depth + 1,
			})
		}
	}

	return result, nil
}

// TraverseDFS performs depth-first search from a starting node
func (hg *HypergraphMemory) TraverseDFS(startID string, maxDepth int, edgeTypes []EdgeType) ([]*MemoryNode, error) {
	hg.mu.RLock()
	defer hg.mu.RUnlock()

	if _, exists := hg.nodes[startID]; !exists {
		return nil, fmt.Errorf("start node not found: %s", startID)
	}

	visited := make(map[string]bool)
	result := make([]*MemoryNode, 0)

	var dfs func(nodeID string, depth int)
	dfs = func(nodeID string, depth int) {
		if visited[nodeID] || depth > maxDepth {
			return
		}

		visited[nodeID] = true

		if node, exists := hg.nodes[nodeID]; exists {
			result = append(result, node)
		}

		for _, edgeID := range hg.outgoing[nodeID] {
			edge := hg.edges[edgeID]

			// Filter by edge type if specified
			if len(edgeTypes) > 0 {
				matchType := false
				for _, et := range edgeTypes {
					if edge.Type == et {
						matchType = true
						break
					}
				}
				if !matchType {
					continue
				}
			}

			dfs(edge.TargetID, depth+1)
		}
	}

	dfs(startID, 0)
	return result, nil
}

// FindShortestPath finds the shortest path between two nodes
func (hg *HypergraphMemory) FindShortestPath(startID, endID string) ([]*MemoryNode, error) {
	hg.mu.RLock()
	defer hg.mu.RUnlock()

	if _, exists := hg.nodes[startID]; !exists {
		return nil, fmt.Errorf("start node not found: %s", startID)
	}
	if _, exists := hg.nodes[endID]; !exists {
		return nil, fmt.Errorf("end node not found: %s", endID)
	}

	visited := make(map[string]bool)
	parent := make(map[string]string)

	type queueItem struct {
		nodeID string
	}

	queue := []queueItem{{nodeID: startID}}
	visited[startID] = true

	found := false

	for len(queue) > 0 && !found {
		current := queue[0]
		queue = queue[1:]

		if current.nodeID == endID {
			found = true
			break
		}

		for _, edgeID := range hg.outgoing[current.nodeID] {
			edge := hg.edges[edgeID]

			if !visited[edge.TargetID] {
				visited[edge.TargetID] = true
				parent[edge.TargetID] = current.nodeID
				queue = append(queue, queueItem{nodeID: edge.TargetID})
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("no path found from %s to %s", startID, endID)
	}

	// Reconstruct path
	path := make([]*MemoryNode, 0)
	current := endID

	for current != "" {
		if node, exists := hg.nodes[current]; exists {
			path = append([]*MemoryNode{node}, path...)
		}
		current = parent[current]
	}

	return path, nil
}

// FindRelatedByType finds nodes related through specific edge types
func (hg *HypergraphMemory) FindRelatedByType(nodeID string, edgeTypes []EdgeType, maxResults int) []*MemoryNode {
	hg.mu.RLock()
	defer hg.mu.RUnlock()

	result := make([]*MemoryNode, 0)

	for _, edgeID := range hg.outgoing[nodeID] {
		if len(result) >= maxResults {
			break
		}

		edge := hg.edges[edgeID]

		// Check edge type
		matchType := false
		for _, et := range edgeTypes {
			if edge.Type == et {
				matchType = true
				break
			}
		}

		if matchType {
			if node, exists := hg.nodes[edge.TargetID]; exists {
				result = append(result, node)
			}
		}
	}

	return result
}

// FindSimilarNodes finds nodes similar to a given node based on embeddings
func (hg *HypergraphMemory) FindSimilarNodes(nodeID string, topK int) ([]*MemoryNode, error) {
	hg.mu.RLock()
	defer hg.mu.RUnlock()

	embedding, exists := hg.embeddings[nodeID]
	if !exists {
		return nil, fmt.Errorf("no embedding for node: %s", nodeID)
	}

	type similarity struct {
		nodeID string
		score  float64
	}

	similarities := make([]similarity, 0)

	for id, emb := range hg.embeddings {
		if id == nodeID {
			continue
		}

		score := cosineSimilarity(embedding, emb)
		similarities = append(similarities, similarity{nodeID: id, score: score})
	}

	// Sort by score descending
	for i := 0; i < len(similarities)-1; i++ {
		for j := i + 1; j < len(similarities); j++ {
			if similarities[j].score > similarities[i].score {
				similarities[i], similarities[j] = similarities[j], similarities[i]
			}
		}
	}

	// Get top K
	result := make([]*MemoryNode, 0, topK)
	for i := 0; i < topK && i < len(similarities); i++ {
		if node, exists := hg.nodes[similarities[i].nodeID]; exists {
			result = append(result, node)
		}
	}

	return result, nil
}

// GetRecentNodes retrieves the most recently created nodes
func (hg *HypergraphMemory) GetRecentNodes(limit int) []*MemoryNode {
	hg.mu.RLock()
	defer hg.mu.RUnlock()

	result := make([]*MemoryNode, 0, limit)

	// Iterate from end of time index (most recent)
	start := len(hg.timeIndex) - limit
	if start < 0 {
		start = 0
	}

	for i := len(hg.timeIndex) - 1; i >= start; i-- {
		if node, exists := hg.nodes[hg.timeIndex[i]]; exists {
			result = append(result, node)
		}
	}

	return result
}

// GetNodeCount returns the total number of nodes
func (hg *HypergraphMemory) GetNodeCount() int {
	hg.mu.RLock()
	defer hg.mu.RUnlock()
	return len(hg.nodes)
}

// GetEdgeCount returns the total number of edges
func (hg *HypergraphMemory) GetEdgeCount() int {
	hg.mu.RLock()
	defer hg.mu.RUnlock()
	return len(hg.edges)
}

// Helper function: cosine similarity
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	dotProduct := 0.0
	magA := 0.0
	magB := 0.0

	for i := range a {
		dotProduct += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}

	magA = math.Sqrt(magA)
	magB = math.Sqrt(magB)

	if magA == 0 || magB == 0 {
		return 0.0
	}

	return dotProduct / (magA * magB)
}
