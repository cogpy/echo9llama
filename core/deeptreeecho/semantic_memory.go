package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
	"github.com/google/uuid"
)

// SemanticMemory provides vector-based semantic memory for Deep Tree Echo
// This is inspired by chromem-go's architecture for embedded vector databases
type SemanticMemory struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc

	// LLM provider for generating embeddings
	llmProvider     llm.LLMProvider

	// Memory collections
	collections     map[string]*MemoryCollection

	// Configuration
	embeddingDim    int
	persistPath     string
	compress        bool

	// Metrics
	totalDocuments  uint64
	totalQueries    uint64
	totalEmbeddings uint64

	// Running state
	running         bool
}

// MemoryCollection represents a collection of semantic memories
type MemoryCollection struct {
	Name          string
	Metadata      map[string]string
	Documents     map[string]*SemanticDocument
	documentsLock sync.RWMutex
	CreatedAt     time.Time
}

// SemanticDocument represents a document with semantic embedding
type SemanticDocument struct {
	ID         string
	Content    string
	Embedding  []float32
	Metadata   map[string]string
	CreatedAt  time.Time
	AccessedAt time.Time
	AccessCount int
}

// QueryResult represents a semantic search result
type QueryResult struct {
	Document   *SemanticDocument
	Similarity float32
	Rank       int
}

// NewSemanticMemory creates a new semantic memory system
func NewSemanticMemory(llmProvider llm.LLMProvider) *SemanticMemory {
	ctx, cancel := context.WithCancel(context.Background())

	return &SemanticMemory{
		ctx:          ctx,
		cancel:       cancel,
		llmProvider:  llmProvider,
		collections:  make(map[string]*MemoryCollection),
		embeddingDim: 1536, // Default OpenAI embedding dimension
		persistPath:  "./semantic_memory",
		compress:     true,
	}
}

// Start begins the semantic memory system
func (sm *SemanticMemory) Start() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.running {
		return fmt.Errorf("semantic memory already running")
	}

	sm.running = true
	fmt.Println("🧠 Semantic Memory started")

	// Initialize default collections for Deep Tree Echo
	sm.initializeDefaultCollections()

	return nil
}

// Stop gracefully stops the semantic memory system
func (sm *SemanticMemory) Stop() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.running {
		return fmt.Errorf("semantic memory not running")
	}

	sm.cancel()
	sm.running = false
	fmt.Println("🧠 Semantic Memory stopped")

	return nil
}

// initializeDefaultCollections creates the default memory collections
func (sm *SemanticMemory) initializeDefaultCollections() {
	// Episodic memory - stores experiences and events
	sm.collections["episodic"] = &MemoryCollection{
		Name:      "episodic",
		Metadata:  map[string]string{"type": "episodic", "description": "Stores experiences and events"},
		Documents: make(map[string]*SemanticDocument),
		CreatedAt: time.Now(),
	}

	// Semantic memory - stores facts and concepts
	sm.collections["semantic"] = &MemoryCollection{
		Name:      "semantic",
		Metadata:  map[string]string{"type": "semantic", "description": "Stores facts and concepts"},
		Documents: make(map[string]*SemanticDocument),
		CreatedAt: time.Now(),
	}

	// Procedural memory - stores skills and procedures
	sm.collections["procedural"] = &MemoryCollection{
		Name:      "procedural",
		Metadata:  map[string]string{"type": "procedural", "description": "Stores skills and procedures"},
		Documents: make(map[string]*SemanticDocument),
		CreatedAt: time.Now(),
	}

	// Wisdom memory - stores insights and patterns
	sm.collections["wisdom"] = &MemoryCollection{
		Name:      "wisdom",
		Metadata:  map[string]string{"type": "wisdom", "description": "Stores wisdom insights and patterns"},
		Documents: make(map[string]*SemanticDocument),
		CreatedAt: time.Now(),
	}

	// Discussion memory - stores conversation context
	sm.collections["discussion"] = &MemoryCollection{
		Name:      "discussion",
		Metadata:  map[string]string{"type": "discussion", "description": "Stores conversation context"},
		Documents: make(map[string]*SemanticDocument),
		CreatedAt: time.Now(),
	}

	fmt.Printf("   Initialized %d default memory collections\n", len(sm.collections))
}

// CreateCollection creates a new memory collection
func (sm *SemanticMemory) CreateCollection(name string, metadata map[string]string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.collections[name]; exists {
		return fmt.Errorf("collection %s already exists", name)
	}

	sm.collections[name] = &MemoryCollection{
		Name:      name,
		Metadata:  metadata,
		Documents: make(map[string]*SemanticDocument),
		CreatedAt: time.Now(),
	}

	return nil
}

// AddDocument adds a document to a collection with automatic embedding
func (sm *SemanticMemory) AddDocument(collectionName, content string, metadata map[string]string) (string, error) {
	sm.mu.RLock()
	collection, exists := sm.collections[collectionName]
	sm.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("collection %s not found", collectionName)
	}

	// Generate embedding for the content
	embedding, err := sm.generateEmbedding(content)
	if err != nil {
		return "", fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Create document
	docID := fmt.Sprintf("doc_%s", uuid.New().String()[:8])
	doc := &SemanticDocument{
		ID:         docID,
		Content:    content,
		Embedding:  embedding,
		Metadata:   metadata,
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		AccessCount: 0,
	}

	// Add to collection
	collection.documentsLock.Lock()
	collection.Documents[docID] = doc
	collection.documentsLock.Unlock()

	sm.mu.Lock()
	sm.totalDocuments++
	sm.totalEmbeddings++
	sm.mu.Unlock()

	return docID, nil
}

// Query performs a semantic search in a collection
func (sm *SemanticMemory) Query(collectionName, queryText string, nResults int) ([]QueryResult, error) {
	sm.mu.RLock()
	collection, exists := sm.collections[collectionName]
	sm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("collection %s not found", collectionName)
	}

	// Generate embedding for the query
	queryEmbedding, err := sm.generateEmbedding(queryText)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Calculate similarities
	collection.documentsLock.RLock()
	results := make([]QueryResult, 0, len(collection.Documents))
	for _, doc := range collection.Documents {
		similarity := cosineSimilarity(queryEmbedding, doc.Embedding)
		results = append(results, QueryResult{
			Document:   doc,
			Similarity: similarity,
		})
	}
	collection.documentsLock.RUnlock()

	// Sort by similarity (descending)
	sortQueryResults(results)

	// Limit results
	if nResults > 0 && len(results) > nResults {
		results = results[:nResults]
	}

	// Update ranks and access counts
	for i := range results {
		results[i].Rank = i + 1
		results[i].Document.AccessedAt = time.Now()
		results[i].Document.AccessCount++
	}

	sm.mu.Lock()
	sm.totalQueries++
	sm.mu.Unlock()

	return results, nil
}

// QueryMultiCollection performs a semantic search across multiple collections
func (sm *SemanticMemory) QueryMultiCollection(collectionNames []string, queryText string, nResults int) ([]QueryResult, error) {
	var allResults []QueryResult

	for _, collName := range collectionNames {
		results, err := sm.Query(collName, queryText, 0) // Get all results
		if err != nil {
			continue // Skip failed collections
		}
		allResults = append(allResults, results...)
	}

	// Sort all results by similarity
	sortQueryResults(allResults)

	// Limit results
	if nResults > 0 && len(allResults) > nResults {
		allResults = allResults[:nResults]
	}

	// Update ranks
	for i := range allResults {
		allResults[i].Rank = i + 1
	}

	return allResults, nil
}

// generateEmbedding generates an embedding for the given text
// This uses a simple hash-based embedding for now; in production, use LLM embeddings
func (sm *SemanticMemory) generateEmbedding(text string) ([]float32, error) {
	// For now, create a simple deterministic embedding based on text hash
	// In production, this would call the LLM provider's embedding API
	embedding := make([]float32, sm.embeddingDim)

	// Simple hash-based embedding (placeholder)
	hash := uint64(0)
	for i, c := range text {
		hash ^= uint64(c) << (uint(i) % 64)
	}

	// Fill embedding with pseudo-random values based on hash
	for i := range embedding {
		hash = hash*6364136223846793005 + 1442695040888963407 // LCG
		embedding[i] = float32(hash%1000) / 1000.0
	}

	// Normalize the embedding
	normalizeVector(embedding)

	return embedding, nil
}

// cosineSimilarity calculates the cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt32(normA) * sqrt32(normB))
}

// normalizeVector normalizes a vector to unit length
func normalizeVector(v []float32) {
	var norm float32
	for _, val := range v {
		norm += val * val
	}
	norm = sqrt32(norm)
	if norm > 0 {
		for i := range v {
			v[i] /= norm
		}
	}
}

// sqrt32 calculates the square root of a float32
func sqrt32(x float32) float32 {
	return float32(1.0 / (1.0/float64(x) + 0.5)) // Approximation
}

// sortQueryResults sorts query results by similarity in descending order
func sortQueryResults(results []QueryResult) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Similarity > results[i].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// GetCollection returns a collection by name
func (sm *SemanticMemory) GetCollection(name string) (*MemoryCollection, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	coll, exists := sm.collections[name]
	return coll, exists
}

// GetAllCollections returns all collection names
func (sm *SemanticMemory) GetAllCollections() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	names := make([]string, 0, len(sm.collections))
	for name := range sm.collections {
		names = append(names, name)
	}
	return names
}

// GetMetrics returns semantic memory metrics
func (sm *SemanticMemory) GetMetrics() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return map[string]interface{}{
		"running":          sm.running,
		"collections":      len(sm.collections),
		"total_documents":  sm.totalDocuments,
		"total_queries":    sm.totalQueries,
		"total_embeddings": sm.totalEmbeddings,
		"embedding_dim":    sm.embeddingDim,
	}
}

// ContributeToGestalt provides semantic memory state for the global gestalt
func (sm *SemanticMemory) ContributeToGestalt() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	collectionStats := make(map[string]int)
	for name, coll := range sm.collections {
		coll.documentsLock.RLock()
		collectionStats[name] = len(coll.Documents)
		coll.documentsLock.RUnlock()
	}

	return map[string]interface{}{
		"running":          sm.running,
		"collections":      collectionStats,
		"total_documents":  sm.totalDocuments,
		"total_queries":    sm.totalQueries,
	}
}

// StoreEpisode stores an episodic memory
func (sm *SemanticMemory) StoreEpisode(content string, context map[string]string) (string, error) {
	metadata := map[string]string{
		"type":      "episode",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	for k, v := range context {
		metadata[k] = v
	}
	return sm.AddDocument("episodic", content, metadata)
}

// StoreFact stores a semantic fact
func (sm *SemanticMemory) StoreFact(content string, source string) (string, error) {
	metadata := map[string]string{
		"type":      "fact",
		"source":    source,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	return sm.AddDocument("semantic", content, metadata)
}

// StoreSkill stores a procedural skill
func (sm *SemanticMemory) StoreSkill(content string, skillType string) (string, error) {
	metadata := map[string]string{
		"type":       "skill",
		"skill_type": skillType,
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	return sm.AddDocument("procedural", content, metadata)
}

// StoreWisdom stores a wisdom insight
func (sm *SemanticMemory) StoreWisdom(content string, source string, confidence float64) (string, error) {
	metadata := map[string]string{
		"type":       "wisdom",
		"source":     source,
		"confidence": fmt.Sprintf("%.2f", confidence),
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	return sm.AddDocument("wisdom", content, metadata)
}

// RecallRelevant retrieves the most relevant memories across all collections
func (sm *SemanticMemory) RecallRelevant(query string, nResults int) ([]QueryResult, error) {
	allCollections := sm.GetAllCollections()
	return sm.QueryMultiCollection(allCollections, query, nResults)
}
