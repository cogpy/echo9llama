//go:build ignore
// +build ignore

// Package milvus provides a Milvus-backed implementation of the CognitiveMemory interface.
package milvus

import (
	"context"
	"fmt"
	"time"

	"github.com/cogpy/echo9llama/core/consciousness"
	"github.com/cogpy/echo9llama/core/memory"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const (
	// Collection and field names
	CollectionName = "echo9llama_thoughts"
	IDField        = "thought_id"
	TimestampField = "timestamp"
	TypeField      = "thought_type"
	ContentField   = "content"
	EmotionField   = "emotion"
	DepthField     = "depth"
	VectorField    = "vector"
	
	// Index parameters
	IndexType   = entity.HNSW
	MetricType  = entity.IP // Inner Product for cosine similarity
	IndexParams = `{"M": 16, "efConstruction": 256}`
	SearchParams = `{"ef": 64}`
)

// MilvusClient implements the memory.CognitiveMemory interface using Milvus.
type MilvusClient struct {
	client            client.Client
	embeddingProvider memory.EmbeddingProvider
	collectionName    string
	vectorDim         int
}

// Config holds the configuration for the MilvusClient.
type Config struct {
	Address           string `json:"address"`           // e.g., "localhost:19530"
	Username          string `json:"username,omitempty"`
	Password          string `json:"password,omitempty"`
	CollectionName    string `json:"collection_name,omitempty"`
}

// NewMilvusClient creates a new client for Milvus.
func NewMilvusClient(ctx context.Context, cfg Config, embedder memory.EmbeddingProvider) (*MilvusClient, error) {
	if cfg.Address == "" {
		return nil, fmt.Errorf("milvus address cannot be empty")
	}
	if embedder == nil {
		return nil, fmt.Errorf("embedding provider cannot be nil")
	}
	if cfg.CollectionName == "" {
		cfg.CollectionName = CollectionName
	}

	// Connect to Milvus
	milvusClient, err := client.NewGrpcClient(ctx, cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to milvus: %w", err)
	}

	// Authenticate if credentials provided
	if cfg.Username != "" && cfg.Password != "" {
		// Note: Authentication is handled during connection in newer SDK versions
		// This is a placeholder for explicit auth if needed
	}

	mc := &MilvusClient{
		client:            milvusClient,
		embeddingProvider: embedder,
		collectionName:    cfg.CollectionName,
		vectorDim:         embedder.Dimension(),
	}

	// Ensure collection exists
	if err := mc.ensureCollection(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	return mc, nil
}

// ensureCollection creates the collection if it doesn't exist.
func (c *MilvusClient) ensureCollection(ctx context.Context) error {
	// Check if collection exists
	has, err := c.client.HasCollection(ctx, c.collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection: %w", err)
	}

	if has {
		// Load collection into memory
		return c.client.LoadCollection(ctx, c.collectionName, false)
	}

	// Define schema
	schema := &entity.Schema{
		CollectionName: c.collectionName,
		Description:    "Echo9llama cognitive memory - stores thoughts as vectors",
		Fields: []*entity.Field{
			{
				Name:       IDField,
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				AutoID:     false,
				TypeParams: map[string]string{"max_length": "256"},
			},
			{
				Name:     TimestampField,
				DataType: entity.FieldTypeInt64,
			},
			{
				Name:       TypeField,
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "64"},
			},
			{
				Name:       ContentField,
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "4096"},
			},
			{
				Name:       EmotionField,
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "64"},
			},
			{
				Name:     DepthField,
				DataType: entity.FieldTypeDouble,
			},
			{
				Name:     VectorField,
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", c.vectorDim),
				},
			},
		},
	}

	// Create collection
	if err := c.client.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	// Create index on vector field
	idx, err := entity.NewIndexHNSW(MetricType, 16, 256)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	if err := c.client.CreateIndex(ctx, c.collectionName, VectorField, idx, false); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// Load collection
	return c.client.LoadCollection(ctx, c.collectionName, false)
}

// StoreThought converts a thought to a vector and stores it in Milvus.
func (c *MilvusClient) StoreThought(ctx context.Context, thought *consciousness.Thought) error {
	return c.StoreThoughts(ctx, []*consciousness.Thought{thought})
}

// StoreThoughts saves a batch of thoughts.
func (c *MilvusClient) StoreThoughts(ctx context.Context, thoughts []*consciousness.Thought) error {
	if len(thoughts) == 0 {
		return nil
	}

	// Prepare data for batch insertion
	ids := make([]string, len(thoughts))
	timestamps := make([]int64, len(thoughts))
	types := make([]string, len(thoughts))
	contents := make([]string, len(thoughts))
	emotions := make([]string, len(thoughts))
	depths := make([]float64, len(thoughts))
	vectors := make([][]float32, len(thoughts))

	// Generate embeddings for all thoughts
	contentTexts := make([]string, len(thoughts))
	for i, thought := range thoughts {
		contentTexts[i] = thought.Content
	}

	embeddings, err := c.embeddingProvider.CreateEmbeddings(ctx, contentTexts)
	if err != nil {
		return fmt.Errorf("failed to create embeddings: %w", err)
	}

	// Populate arrays
	for i, thought := range thoughts {
		ids[i] = thought.ID
		timestamps[i] = thought.Timestamp.Unix()
		types[i] = string(thought.Type)
		contents[i] = thought.Content
		emotions[i] = thought.Emotion
		depths[i] = thought.Depth
		vectors[i] = embeddings[i]
	}

	// Create columns
	idColumn := entity.NewColumnVarChar(IDField, ids)
	timestampColumn := entity.NewColumnInt64(TimestampField, timestamps)
	typeColumn := entity.NewColumnVarChar(TypeField, types)
	contentColumn := entity.NewColumnVarChar(ContentField, contents)
	emotionColumn := entity.NewColumnVarChar(EmotionField, emotions)
	depthColumn := entity.NewColumnDouble(DepthField, depths)
	vectorColumn := entity.NewColumnFloatVector(VectorField, c.vectorDim, vectors)

	// Insert data
	_, err = c.client.Insert(ctx, c.collectionName, "",
		idColumn, timestampColumn, typeColumn, contentColumn,
		emotionColumn, depthColumn, vectorColumn)
	if err != nil {
		return fmt.Errorf("failed to insert thoughts: %w", err)
	}

	// Flush to ensure data is persisted
	if err := c.client.Flush(ctx, c.collectionName, false); err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}

	return nil
}

// RetrieveSimilarThoughts searches for similar thoughts in Milvus.
func (c *MilvusClient) RetrieveSimilarThoughts(ctx context.Context, content string, topK int) ([]*consciousness.Thought, error) {
	// Generate embedding for search query
	embedding, err := c.embeddingProvider.CreateEmbedding(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}

	// Prepare search vectors
	searchVectors := []entity.Vector{
		entity.FloatVector(embedding),
	}

	// Define search parameters
	sp, _ := entity.NewIndexHNSWSearchParam(64)

	// Execute search
	results, err := c.client.Search(
		ctx,
		c.collectionName,
		nil,
		"",
		[]string{IDField, TimestampField, TypeField, ContentField, EmotionField, DepthField},
		searchVectors,
		VectorField,
		MetricType,
		topK,
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	if len(results) == 0 {
		return []*consciousness.Thought{}, nil
	}

	// Convert results to thoughts
	thoughts := make([]*consciousness.Thought, 0, topK)
	for i := 0; i < results[0].ResultCount; i++ {
		id, _ := results[0].Fields.GetColumn(IDField).Get(i)
		timestamp, _ := results[0].Fields.GetColumn(TimestampField).GetAsInt64(i)
		thoughtType, _ := results[0].Fields.GetColumn(TypeField).Get(i)
		content, _ := results[0].Fields.GetColumn(ContentField).Get(i)
		emotion, _ := results[0].Fields.GetColumn(EmotionField).Get(i)
		depth, _ := results[0].Fields.GetColumn(DepthField).GetAsDouble(i)

		thought := &consciousness.Thought{
			ID:        id.(string),
			Type:      consciousness.ThoughtType(thoughtType.(string)),
			Content:   content.(string),
			Timestamp: time.Unix(timestamp, 0),
			Emotion:   emotion.(string),
			Depth:     depth,
		}
		thoughts = append(thoughts, thought)
	}

	return thoughts, nil
}

// GetThoughtByID retrieves a specific thought by its unique ID.
func (c *MilvusClient) GetThoughtByID(ctx context.Context, id string) (*consciousness.Thought, error) {
	expr := fmt.Sprintf("%s == '%s'", IDField, id)
	
	results, err := c.client.Query(
		ctx,
		c.collectionName,
		nil,
		expr,
		[]string{IDField, TimestampField, TypeField, ContentField, EmotionField, DepthField},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	if results.Len() == 0 {
		return nil, fmt.Errorf("thought not found: %s", id)
	}

	// Extract first result
	timestamp, _ := results.GetColumn(TimestampField).GetAsInt64(0)
	thoughtType, _ := results.GetColumn(TypeField).Get(0)
	content, _ := results.GetColumn(ContentField).Get(0)
	emotion, _ := results.GetColumn(EmotionField).Get(0)
	depth, _ := results.GetColumn(DepthField).GetAsDouble(0)

	return &consciousness.Thought{
		ID:        id,
		Type:      consciousness.ThoughtType(thoughtType.(string)),
		Content:   content.(string),
		Timestamp: time.Unix(timestamp, 0),
		Emotion:   emotion.(string),
		Depth:     depth,
	}, nil
}

// GetRecentThoughts retrieves the most recent thoughts.
func (c *MilvusClient) GetRecentThoughts(ctx context.Context, limit int) ([]*consciousness.Thought, error) {
	// Query all and sort by timestamp (Milvus doesn't support ORDER BY in query yet)
	// This is a simplified implementation; for production, consider pagination
	
	results, err := c.client.Query(
		ctx,
		c.collectionName,
		nil,
		"",
		[]string{IDField, TimestampField, TypeField, ContentField, EmotionField, DepthField},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	// Convert to thoughts and sort by timestamp
	thoughts := make([]*consciousness.Thought, 0, results.Len())
	for i := 0; i < results.Len(); i++ {
		id, _ := results.GetColumn(IDField).Get(i)
		timestamp, _ := results.GetColumn(TimestampField).GetAsInt64(i)
		thoughtType, _ := results.GetColumn(TypeField).Get(i)
		content, _ := results.GetColumn(ContentField).Get(i)
		emotion, _ := results.GetColumn(EmotionField).Get(i)
		depth, _ := results.GetColumn(DepthField).GetAsDouble(i)

		thoughts = append(thoughts, &consciousness.Thought{
			ID:        id.(string),
			Type:      consciousness.ThoughtType(thoughtType.(string)),
			Content:   content.(string),
			Timestamp: time.Unix(timestamp, 0),
			Emotion:   emotion.(string),
			Depth:     depth,
		})
	}

	// Sort by timestamp descending (most recent first)
	// Note: In production, use a more efficient approach
	for i := 0; i < len(thoughts)-1; i++ {
		for j := i + 1; j < len(thoughts); j++ {
			if thoughts[i].Timestamp.Before(thoughts[j].Timestamp) {
				thoughts[i], thoughts[j] = thoughts[j], thoughts[i]
			}
		}
	}

	// Return top limit
	if len(thoughts) > limit {
		thoughts = thoughts[:limit]
	}

	return thoughts, nil
}

// GetThoughtsByType retrieves thoughts of a specific type.
func (c *MilvusClient) GetThoughtsByType(ctx context.Context, thoughtType consciousness.ThoughtType, limit int) ([]*consciousness.Thought, error) {
	expr := fmt.Sprintf("%s == '%s'", TypeField, string(thoughtType))
	
	results, err := c.client.Query(
		ctx,
		c.collectionName,
		nil,
		expr,
		[]string{IDField, TimestampField, TypeField, ContentField, EmotionField, DepthField},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	thoughts := make([]*consciousness.Thought, 0, results.Len())
	count := 0
	for i := 0; i < results.Len() && count < limit; i++ {
		id, _ := results.GetColumn(IDField).Get(i)
		timestamp, _ := results.GetColumn(TimestampField).GetAsInt64(i)
		tType, _ := results.GetColumn(TypeField).Get(i)
		content, _ := results.GetColumn(ContentField).Get(i)
		emotion, _ := results.GetColumn(EmotionField).Get(i)
		depth, _ := results.GetColumn(DepthField).GetAsDouble(i)

		thoughts = append(thoughts, &consciousness.Thought{
			ID:        id.(string),
			Type:      consciousness.ThoughtType(tType.(string)),
			Content:   content.(string),
			Timestamp: time.Unix(timestamp, 0),
			Emotion:   emotion.(string),
			Depth:     depth,
		})
		count++
	}

	return thoughts, nil
}

// GetThoughtsByTimeRange retrieves thoughts within a time range.
func (c *MilvusClient) GetThoughtsByTimeRange(ctx context.Context, startTime, endTime int64, limit int) ([]*consciousness.Thought, error) {
	expr := fmt.Sprintf("%s >= %d && %s <= %d", TimestampField, startTime, TimestampField, endTime)
	
	results, err := c.client.Query(
		ctx,
		c.collectionName,
		nil,
		expr,
		[]string{IDField, TimestampField, TypeField, ContentField, EmotionField, DepthField},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	thoughts := make([]*consciousness.Thought, 0, results.Len())
	count := 0
	for i := 0; i < results.Len() && count < limit; i++ {
		id, _ := results.GetColumn(IDField).Get(i)
		timestamp, _ := results.GetColumn(TimestampField).GetAsInt64(i)
		thoughtType, _ := results.GetColumn(TypeField).Get(i)
		content, _ := results.GetColumn(ContentField).Get(i)
		emotion, _ := results.GetColumn(EmotionField).Get(i)
		depth, _ := results.GetColumn(DepthField).GetAsDouble(i)

		thoughts = append(thoughts, &consciousness.Thought{
			ID:        id.(string),
			Type:      consciousness.ThoughtType(thoughtType.(string)),
			Content:   content.(string),
			Timestamp: time.Unix(timestamp, 0),
			Emotion:   emotion.(string),
			Depth:     depth,
		})
		count++
	}

	return thoughts, nil
}

// DeleteThought removes a thought from the memory backend.
func (c *MilvusClient) DeleteThought(ctx context.Context, id string) error {
	expr := fmt.Sprintf("%s == '%s'", IDField, id)
	
	if err := c.client.Delete(ctx, c.collectionName, "", expr); err != nil {
		return fmt.Errorf("failed to delete thought: %w", err)
	}

	return c.client.Flush(ctx, c.collectionName, false)
}

// Clear removes all thoughts from the memory backend.
func (c *MilvusClient) Clear(ctx context.Context) error {
	// Drop and recreate collection
	if err := c.client.DropCollection(ctx, c.collectionName); err != nil {
		return fmt.Errorf("failed to drop collection: %w", err)
	}

	return c.ensureCollection(ctx)
}

// Close closes the Milvus client connection.
func (c *MilvusClient) Close() error {
	return c.client.Close()
}
