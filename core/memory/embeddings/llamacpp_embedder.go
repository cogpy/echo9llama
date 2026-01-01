//go:build ignore
// +build ignore

// Package embeddings provides embedding providers for the memory system.
package embeddings

import (
	"context"
	"fmt"

	"github.com/cogpy/echo9llama/core/llm/providers/llamacpp"
	"github.com/cogpy/echo9llama/core/memory"
)

// LlamaCppEmbedder implements the memory.EmbeddingProvider interface using llama.cpp.
// This enables local, on-device embedding generation without external API dependencies.
type LlamaCppEmbedder struct {
	provider  *llamacpp.LlamaCppProvider
	dimension int
}

// NewLlamaCppEmbedder creates a new embedding provider using a llama.cpp model.
// The model should be an embedding model (e.g., nomic-embed, bge, etc.) for best results.
func NewLlamaCppEmbedder(modelPath string, dimension int) (*LlamaCppEmbedder, error) {
	config := llamacpp.DefaultConfig(modelPath)
	config.Name = "llamacpp-embedder"
	config.ContextSize = 512 // Embedding models typically need less context
	config.GPULayers = -1    // Auto-detect GPU

	provider, err := llamacpp.NewLlamaCppProvider(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create llama.cpp provider: %w", err)
	}

	return &LlamaCppEmbedder{
		provider:  provider,
		dimension: dimension,
	}, nil
}

// CreateEmbedding generates a vector embedding for the given text.
func (e *LlamaCppEmbedder) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	embedding, err := e.provider.GetEmbedding(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Verify dimension matches
	if len(embedding) != e.dimension {
		return nil, fmt.Errorf("embedding dimension mismatch: got %d, expected %d", len(embedding), e.dimension)
	}

	return embedding, nil
}

// CreateEmbeddings generates embeddings for multiple texts in a batch.
func (e *LlamaCppEmbedder) CreateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	
	for i, text := range texts {
		embedding, err := e.CreateEmbedding(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to create embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// Dimension returns the dimensionality of the embeddings produced.
func (e *LlamaCppEmbedder) Dimension() int {
	return e.dimension
}

// Close releases the resources used by the embedder.
func (e *LlamaCppEmbedder) Close() error {
	return e.provider.Close()
}
