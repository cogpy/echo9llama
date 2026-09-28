// Package memory provides interfaces and implementations for cognitive memory systems.
package memory

import (
	"context"

	"github.com/cogpy/echo9llama/core/consciousness"
)

// CognitiveMemory defines the interface for a persistent, searchable memory system.
// Implementations of this interface provide storage and retrieval of thoughts
// with semantic search capabilities.
type CognitiveMemory interface {
	// StoreThought saves a single thought to the memory backend.
	StoreThought(ctx context.Context, thought *consciousness.Thought) error

	// StoreThoughts saves a batch of thoughts for efficiency.
	StoreThoughts(ctx context.Context, thoughts []*consciousness.Thought) error

	// RetrieveSimilarThoughts finds thoughts semantically similar to the given content.
	// Returns up to topK thoughts ordered by similarity (most similar first).
	RetrieveSimilarThoughts(ctx context.Context, content string, topK int) ([]*consciousness.Thought, error)

	// GetThoughtByID retrieves a specific thought by its unique ID.
	GetThoughtByID(ctx context.Context, id string) (*consciousness.Thought, error)

	// GetRecentThoughts retrieves the most recent thoughts up to the specified limit.
	GetRecentThoughts(ctx context.Context, limit int) ([]*consciousness.Thought, error)

	// GetThoughtsByType retrieves thoughts of a specific type.
	GetThoughtsByType(ctx context.Context, thoughtType consciousness.ThoughtType, limit int) ([]*consciousness.Thought, error)

	// GetThoughtsByTimeRange retrieves thoughts within a time range.
	GetThoughtsByTimeRange(ctx context.Context, startTime, endTime int64, limit int) ([]*consciousness.Thought, error)

	// DeleteThought removes a thought from the memory backend.
	DeleteThought(ctx context.Context, id string) error

	// Clear removes all thoughts from the memory backend.
	Clear(ctx context.Context) error
}

// EmbeddingProvider defines the interface for converting text to vector embeddings.
// This abstraction allows for different embedding models to be used.
type EmbeddingProvider interface {
	// CreateEmbedding generates a vector embedding for the given text.
	// Returns a float32 slice representing the embedding vector.
	CreateEmbedding(ctx context.Context, text string) ([]float32, error)

	// CreateEmbeddings generates embeddings for multiple texts in a batch.
	CreateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)

	// Dimension returns the dimensionality of the embeddings produced.
	Dimension() int
}

// MemoryMetrics contains statistics about memory usage.
type MemoryMetrics struct {
	TotalThoughts  int64
	ThoughtsByType map[consciousness.ThoughtType]int64
	OldestThought  int64 // Unix timestamp
	NewestThought  int64 // Unix timestamp
	AverageDepth   float64
	StorageSize    int64 // Bytes
}

// MetricsProvider is an optional interface for memory backends that support metrics.
type MetricsProvider interface {
	// GetMetrics returns usage statistics for the memory system.
	GetMetrics(ctx context.Context) (*MemoryMetrics, error)
}
