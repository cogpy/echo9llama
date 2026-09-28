package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/persistence"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"github.com/google/uuid"
)

// DgraphHypergraph implements persistent hypergraph storage using Dgraph
// This replaces the in-memory HypergraphMemory for production use
type DgraphHypergraph struct {
	mu           sync.RWMutex
	client       *persistence.DgraphClient
	ctx          context.Context
	cancel       context.CancelFunc
	schemaLoaded bool

	// Local cache for frequently accessed nodes
	nodeCache map[string]*MemoryNode
	cacheSize int
	cacheTTL  time.Duration
}

// DgraphNode represents a node in Dgraph format
type DgraphNode struct {
	UID         string    `json:"uid,omitempty"`
	DType       []string  `json:"dgraph.type,omitempty"`
	NodeID      string    `json:"node_id,omitempty"`
	NodeType    string    `json:"node_type,omitempty"`
	Content     string    `json:"content,omitempty"`
	Embedding   []float64 `json:"embedding,omitempty"`
	Activation  float64   `json:"activation,omitempty"`
	Importance  float64   `json:"importance,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	AccessedAt  time.Time `json:"accessed_at,omitempty"`
	AccessCount int       `json:"access_count,omitempty"`
	Metadata    string    `json:"metadata,omitempty"`
}

// DgraphEdge represents an edge in Dgraph format
type DgraphEdge struct {
	UID        string      `json:"uid,omitempty"`
	DType      []string    `json:"dgraph.type,omitempty"`
	EdgeID     string      `json:"edge_id,omitempty"`
	EdgeType   string      `json:"edge_type,omitempty"`
	Weight     float64     `json:"weight,omitempty"`
	Confidence float64     `json:"confidence,omitempty"`
	CreatedAt  time.Time   `json:"created_at,omitempty"`
	Metadata   string      `json:"metadata,omitempty"`
	Source     *DgraphNode `json:"source,omitempty"`
	Target     *DgraphNode `json:"target,omitempty"`
}

// DgraphHyperEdge represents a hyperedge in Dgraph format
type DgraphHyperEdge struct {
	UID           string        `json:"uid,omitempty"`
	DType         []string      `json:"dgraph.type,omitempty"`
	HyperEdgeID   string        `json:"hyperedge_id,omitempty"`
	HyperEdgeType string        `json:"hyperedge_type,omitempty"`
	Weight        float64       `json:"weight,omitempty"`
	CreatedAt     time.Time     `json:"created_at,omitempty"`
	Metadata      string        `json:"metadata,omitempty"`
	Members       []*DgraphNode `json:"members,omitempty"`
}

// NewDgraphHypergraph creates a new Dgraph-backed hypergraph
func NewDgraphHypergraph(client *persistence.DgraphClient) (*DgraphHypergraph, error) {
	if client == nil {
		return nil, fmt.Errorf("dgraph client is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	hg := &DgraphHypergraph{
		client:    client,
		ctx:       ctx,
		cancel:    cancel,
		nodeCache: make(map[string]*MemoryNode),
		cacheSize: 1000,
		cacheTTL:  time.Minute * 5,
	}

	return hg, nil
}

// LoadSchema loads the Dgraph schema from the schema.dql file
func (hg *DgraphHypergraph) LoadSchema(schema string) error {
	hg.mu.Lock()
	defer hg.mu.Unlock()

	if err := hg.client.SetSchema(schema); err != nil {
		return fmt.Errorf("failed to set schema: %w", err)
	}

	hg.schemaLoaded = true
	return nil
}

// AddNode adds a new node to the hypergraph
func (hg *DgraphHypergraph) AddNode(node *MemoryNode) error {
	hg.mu.Lock()
	defer hg.mu.Unlock()

	if node.ID == "" {
		node.ID = uuid.New().String()
	}

	// Convert metadata to JSON string
	metadataJSON, err := json.Marshal(node.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	dgNode := &DgraphNode{
		DType:       []string{"MemoryNode"},
		NodeID:      node.ID,
		NodeType:    string(node.Type),
		Content:     node.Content,
		Embedding:   node.Embedding,
		Activation:  0.5,
		Importance:  0.5,
		CreatedAt:   time.Now(),
		AccessedAt:  time.Now(),
		AccessCount: 0,
		Metadata:    string(metadataJSON),
	}

	data, err := json.Marshal(dgNode)
	if err != nil {
		return fmt.Errorf("failed to marshal node: %w", err)
	}

	mu := &api.Mutation{
		SetJson:   data,
		CommitNow: true,
	}

	_, err = hg.client.Mutate(hg.ctx, mu)
	if err != nil {
		return fmt.Errorf("failed to add node: %w", err)
	}

	// Update cache
	hg.nodeCache[node.ID] = node

	return nil
}

// GetNode retrieves a node by ID
func (hg *DgraphHypergraph) GetNode(nodeID string) (*MemoryNode, error) {
	hg.mu.RLock()
	// Check cache first
	if node, ok := hg.nodeCache[nodeID]; ok {
		hg.mu.RUnlock()
		return node, nil
	}
	hg.mu.RUnlock()

	query := `query GetNode($nodeID: string) {
		node(func: eq(node_id, $nodeID)) {
			uid
			node_id
			node_type
			content
			embedding
			activation
			importance
			created_at
			accessed_at
			access_count
			metadata
		}
	}`

	vars := map[string]string{"$nodeID": nodeID}
	resp, err := hg.client.Query(hg.ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to query node: %w", err)
	}

	var result struct {
		Node []DgraphNode `json:"node"`
	}
	if err := json.Unmarshal(resp.Json, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(result.Node) == 0 {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}

	dgNode := result.Node[0]

	// Parse metadata
	var metadata map[string]interface{}
	if dgNode.Metadata != "" {
		json.Unmarshal([]byte(dgNode.Metadata), &metadata)
	}

	node := &MemoryNode{
		ID:        dgNode.NodeID,
		Type:      NodeType(dgNode.NodeType),
		Content:   dgNode.Content,
		Embedding: dgNode.Embedding,
		Metadata:  metadata,
	}

	// Update cache
	hg.mu.Lock()
	hg.nodeCache[nodeID] = node
	hg.mu.Unlock()

	// Update access time
	go hg.updateAccessTime(dgNode.UID)

	return node, nil
}

// AddEdge adds a binary edge between two nodes
func (hg *DgraphHypergraph) AddEdge(edge *MemoryEdge) error {
	hg.mu.Lock()
	defer hg.mu.Unlock()

	if edge.ID == "" {
		edge.ID = uuid.New().String()
	}

	// First, get UIDs for source and target nodes
	sourceUID, err := hg.getNodeUID(edge.SourceID)
	if err != nil {
		return fmt.Errorf("source node not found: %w", err)
	}

	targetUID, err := hg.getNodeUID(edge.TargetID)
	if err != nil {
		return fmt.Errorf("target node not found: %w", err)
	}

	metadataJSON, _ := json.Marshal(edge.Metadata)

	dgEdge := map[string]interface{}{
		"dgraph.type": []string{"MemoryEdge"},
		"edge_id":     edge.ID,
		"edge_type":   string(edge.Type),
		"weight":      edge.Weight,
		"confidence":  1.0,
		"created_at":  time.Now(),
		"metadata":    string(metadataJSON),
		"source":      map[string]string{"uid": sourceUID},
		"target":      map[string]string{"uid": targetUID},
	}

	data, err := json.Marshal(dgEdge)
	if err != nil {
		return fmt.Errorf("failed to marshal edge: %w", err)
	}

	mu := &api.Mutation{
		SetJson:   data,
		CommitNow: true,
	}

	_, err = hg.client.Mutate(hg.ctx, mu)
	if err != nil {
		return fmt.Errorf("failed to add edge: %w", err)
	}

	return nil
}

// AddHyperEdge adds a hyperedge connecting multiple nodes
func (hg *DgraphHypergraph) AddHyperEdge(hyperEdge *HyperEdge) error {
	hg.mu.Lock()
	defer hg.mu.Unlock()

	if hyperEdge.ID == "" {
		hyperEdge.ID = uuid.New().String()
	}

	// Get UIDs for all member nodes
	members := make([]map[string]string, 0, len(hyperEdge.NodeIDs))
	for _, nodeID := range hyperEdge.NodeIDs {
		uid, err := hg.getNodeUID(nodeID)
		if err != nil {
			return fmt.Errorf("member node not found: %s: %w", nodeID, err)
		}
		members = append(members, map[string]string{"uid": uid})
	}

	metadataJSON, _ := json.Marshal(hyperEdge.Metadata)

	dgHyperEdge := map[string]interface{}{
		"dgraph.type":    []string{"HyperEdge"},
		"hyperedge_id":   hyperEdge.ID,
		"hyperedge_type": string(hyperEdge.Type),
		"weight":         0.0, // HyperEdge has no Weight field
		"created_at":     time.Now(),
		"metadata":       string(metadataJSON),
		"members":        members,
	}

	data, err := json.Marshal(dgHyperEdge)
	if err != nil {
		return fmt.Errorf("failed to marshal hyperedge: %w", err)
	}

	mu := &api.Mutation{
		SetJson:   data,
		CommitNow: true,
	}

	_, err = hg.client.Mutate(hg.ctx, mu)
	if err != nil {
		return fmt.Errorf("failed to add hyperedge: %w", err)
	}

	return nil
}

// Traverse performs a graph traversal from a starting node
func (hg *DgraphHypergraph) Traverse(startNodeID string, maxDepth int, edgeTypes []EdgeType) ([]*MemoryNode, error) {
	// Build edge type filter
	edgeFilter := ""
	if len(edgeTypes) > 0 {
		types := make([]string, len(edgeTypes))
		for i, t := range edgeTypes {
			types[i] = fmt.Sprintf(`"%s"`, t)
		}
		edgeFilter = fmt.Sprintf("@filter(eq(edge_type, [%s]))", types)
	}

	query := fmt.Sprintf(`query Traverse($nodeID: string) {
		var(func: eq(node_id, $nodeID)) {
			start as uid
		}
		
		traverse(func: uid(start)) @recurse(depth: %d) {
			uid
			node_id
			node_type
			content
			target %s {
				uid
				node_id
				node_type
				content
			}
		}
	}`, maxDepth, edgeFilter)

	vars := map[string]string{"$nodeID": startNodeID}
	resp, err := hg.client.Query(hg.ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to traverse: %w", err)
	}

	var result struct {
		Traverse []DgraphNode `json:"traverse"`
	}
	if err := json.Unmarshal(resp.Json, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal traverse result: %w", err)
	}

	nodes := make([]*MemoryNode, 0, len(result.Traverse))
	for _, dgNode := range result.Traverse {
		nodes = append(nodes, &MemoryNode{
			ID:      dgNode.NodeID,
			Type:    NodeType(dgNode.NodeType),
			Content: dgNode.Content,
		})
	}

	return nodes, nil
}

// SearchByContent performs full-text search on node content
func (hg *DgraphHypergraph) SearchByContent(query string, limit int) ([]*MemoryNode, error) {
	dgQuery := fmt.Sprintf(`{
		search(func: alloftext(content, "%s"), first: %d) {
			uid
			node_id
			node_type
			content
			activation
			importance
		}
	}`, query, limit)

	resp, err := hg.client.Query(hg.ctx, dgQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	var result struct {
		Search []DgraphNode `json:"search"`
	}
	if err := json.Unmarshal(resp.Json, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search result: %w", err)
	}

	nodes := make([]*MemoryNode, 0, len(result.Search))
	for _, dgNode := range result.Search {
		nodes = append(nodes, &MemoryNode{
			ID:      dgNode.NodeID,
			Type:    NodeType(dgNode.NodeType),
			Content: dgNode.Content,
		})
	}

	return nodes, nil
}

// GetNodesByType retrieves all nodes of a specific type
func (hg *DgraphHypergraph) GetNodesByType(nodeType NodeType, limit int) ([]*MemoryNode, error) {
	query := fmt.Sprintf(`{
		nodes(func: eq(node_type, "%s"), first: %d, orderasc: created_at) {
			uid
			node_id
			node_type
			content
			activation
			importance
			created_at
		}
	}`, nodeType, limit)

	resp, err := hg.client.Query(hg.ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes by type: %w", err)
	}

	var result struct {
		Nodes []DgraphNode `json:"nodes"`
	}
	if err := json.Unmarshal(resp.Json, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	nodes := make([]*MemoryNode, 0, len(result.Nodes))
	for _, dgNode := range result.Nodes {
		nodes = append(nodes, &MemoryNode{
			ID:      dgNode.NodeID,
			Type:    NodeType(dgNode.NodeType),
			Content: dgNode.Content,
		})
	}

	return nodes, nil
}

// UpdateNodeActivation updates the activation level of a node
func (hg *DgraphHypergraph) UpdateNodeActivation(nodeID string, activation float64) error {
	uid, err := hg.getNodeUID(nodeID)
	if err != nil {
		return err
	}

	update := map[string]interface{}{
		"uid":        uid,
		"activation": activation,
	}

	data, _ := json.Marshal(update)
	mu := &api.Mutation{
		SetJson:   data,
		CommitNow: true,
	}

	_, err = hg.client.Mutate(hg.ctx, mu)
	return err
}

// Close closes the hypergraph and releases resources
func (hg *DgraphHypergraph) Close() error {
	hg.cancel()
	return nil
}

// Helper functions

func (hg *DgraphHypergraph) getNodeUID(nodeID string) (string, error) {
	query := `query GetUID($nodeID: string) {
		node(func: eq(node_id, $nodeID)) {
			uid
		}
	}`

	vars := map[string]string{"$nodeID": nodeID}
	resp, err := hg.client.Query(hg.ctx, query, vars)
	if err != nil {
		return "", err
	}

	var result struct {
		Node []struct {
			UID string `json:"uid"`
		} `json:"node"`
	}
	if err := json.Unmarshal(resp.Json, &result); err != nil {
		return "", err
	}

	if len(result.Node) == 0 {
		return "", fmt.Errorf("node not found: %s", nodeID)
	}

	return result.Node[0].UID, nil
}

func (hg *DgraphHypergraph) updateAccessTime(uid string) {
	update := map[string]interface{}{
		"uid":         uid,
		"accessed_at": time.Now(),
	}

	data, _ := json.Marshal(update)
	mu := &api.Mutation{
		SetJson:   data,
		CommitNow: true,
	}

	hg.client.Mutate(hg.ctx, mu)
}
