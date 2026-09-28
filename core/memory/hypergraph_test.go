package memory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Define test-specific type constants that match the production types
const (
	NodeTypeEpisodic   NodeType = "episodic"
	NodeTypeSemantic   NodeType = "semantic"
	NodeTypeProcedural NodeType = "procedural"
	NodeTypeWorking    NodeType = "working"

	EdgeTypeAssociation EdgeType = "association"
	EdgeTypeCausal      EdgeType = "causal"
	EdgeTypeTemporal    EdgeType = "temporal"
	EdgeTypeSimilarity  EdgeType = "similarity"
)

// TestMemoryNode tests the MemoryNode struct and its methods
func TestMemoryNode(t *testing.T) {
	t.Run("CreateNode", func(t *testing.T) {
		node := &MemoryNode{
			ID:        "test-node-1",
			Type:      NodeTypeEpisodic,
			Content:   "Test memory content",
			Embedding: []float64{0.1, 0.2, 0.3},
			Metadata: map[string]interface{}{
				"source": "test",
			},
		}

		assert.Equal(t, "test-node-1", node.ID)
		assert.Equal(t, NodeTypeEpisodic, node.Type)
		assert.Equal(t, "Test memory content", node.Content)
		assert.Len(t, node.Embedding, 3)
	})

	t.Run("NodeTypes", func(t *testing.T) {
		types := []NodeType{
			NodeTypeEpisodic,
			NodeTypeSemantic,
			NodeTypeProcedural,
			NodeTypeWorking,
		}

		for _, nodeType := range types {
			assert.NotEmpty(t, string(nodeType))
		}
	})

	t.Run("NodeMetadata", func(t *testing.T) {
		node := &MemoryNode{
			ID:      "test-node-meta",
			Type:    NodeTypeEpisodic,
			Content: "Test content",
			Metadata: map[string]interface{}{
				"source":     "test",
				"importance": 0.8,
				"tags":       []string{"test", "memory"},
			},
		}

		assert.Equal(t, "test", node.Metadata["source"])
		assert.Equal(t, 0.8, node.Metadata["importance"])
	})

	t.Run("NodeTimestamps", func(t *testing.T) {
		now := time.Now()
		node := &MemoryNode{
			ID:        "test-node-time",
			Type:      NodeTypeEpisodic,
			Content:   "Test content",
			CreatedAt: now,
			UpdatedAt: now,
		}

		assert.Equal(t, now, node.CreatedAt)
		assert.Equal(t, now, node.UpdatedAt)
	})
}

// TestMemoryEdge tests the MemoryEdge struct
func TestMemoryEdge(t *testing.T) {
	t.Run("CreateEdge", func(t *testing.T) {
		edge := &MemoryEdge{
			ID:       "test-edge-1",
			Type:     EdgeTypeAssociation,
			SourceID: "node-1",
			TargetID: "node-2",
			Weight:   0.8,
			Metadata: map[string]interface{}{
				"created": time.Now(),
			},
		}

		assert.Equal(t, "test-edge-1", edge.ID)
		assert.Equal(t, EdgeTypeAssociation, edge.Type)
		assert.Equal(t, "node-1", edge.SourceID)
		assert.Equal(t, "node-2", edge.TargetID)
		assert.Equal(t, 0.8, edge.Weight)
	})

	t.Run("EdgeTypes", func(t *testing.T) {
		types := []EdgeType{
			EdgeTypeAssociation,
			EdgeTypeCausal,
			EdgeTypeTemporal,
			EdgeTypeSimilarity,
		}

		for _, edgeType := range types {
			assert.NotEmpty(t, string(edgeType))
		}
	})

	t.Run("EdgeWeight", func(t *testing.T) {
		edge := &MemoryEdge{
			ID:       "test-edge-weight",
			SourceID: "a",
			TargetID: "b",
			Type:     EdgeTypeAssociation,
			Weight:   0.5,
		}

		assert.Equal(t, 0.5, edge.Weight)
		assert.GreaterOrEqual(t, edge.Weight, 0.0)
		assert.LessOrEqual(t, edge.Weight, 1.0)
	})
}

// TestHyperEdge tests the HyperEdge struct
func TestHyperEdge(t *testing.T) {
	t.Run("CreateHyperEdge", func(t *testing.T) {
		hyperEdge := &HyperEdge{
			ID:      "test-hyperedge-1",
			NodeIDs: []string{"node-1", "node-2", "node-3"},
			Type:    "conceptual_cluster",
			Metadata: map[string]interface{}{
				"strength": 0.9,
			},
		}

		assert.Equal(t, "test-hyperedge-1", hyperEdge.ID)
		assert.Len(t, hyperEdge.NodeIDs, 3)
		assert.Equal(t, "conceptual_cluster", hyperEdge.Type)
	})

	t.Run("HyperEdgeNodeIDs", func(t *testing.T) {
		hyperEdge := &HyperEdge{
			ID:      "test-hyperedge-nodes",
			NodeIDs: []string{"A", "B", "C", "D"},
			Type:    "cluster",
		}

		assert.Contains(t, hyperEdge.NodeIDs, "A")
		assert.Contains(t, hyperEdge.NodeIDs, "B")
		assert.Contains(t, hyperEdge.NodeIDs, "C")
		assert.Contains(t, hyperEdge.NodeIDs, "D")
		assert.Len(t, hyperEdge.NodeIDs, 4)
	})
}

// TestEpisode tests the Episode struct
func TestEpisode(t *testing.T) {
	t.Run("CreateEpisode", func(t *testing.T) {
		now := time.Now()
		episode := &Episode{
			ID:         "test-episode-1",
			Timestamp:  now,
			Context:    "Test context",
			Importance: 0.7,
			NodeIDs:    []string{"node-1", "node-2"},
			Metadata: map[string]interface{}{
				"source": "test",
			},
		}

		assert.Equal(t, "test-episode-1", episode.ID)
		assert.Equal(t, now, episode.Timestamp)
		assert.Equal(t, "Test context", episode.Context)
		assert.Equal(t, 0.7, episode.Importance)
		assert.Len(t, episode.NodeIDs, 2)
	})
}

// TestIdentitySnapshot tests the IdentitySnapshot struct
func TestIdentitySnapshot(t *testing.T) {
	t.Run("CreateSnapshot", func(t *testing.T) {
		now := time.Now()
		snapshot := &IdentitySnapshot{
			ID:        "test-snapshot-1",
			Timestamp: now,
			Coherence: 0.95,
			State: map[string]interface{}{
				"mood":   "contemplative",
				"energy": 0.8,
			},
			Metadata: map[string]interface{}{
				"trigger": "reflection",
			},
		}

		assert.Equal(t, "test-snapshot-1", snapshot.ID)
		assert.Equal(t, now, snapshot.Timestamp)
		assert.Equal(t, 0.95, snapshot.Coherence)
		assert.Equal(t, "contemplative", snapshot.State["mood"])
	})
}

// TestDreamJournal tests the DreamJournal struct
func TestDreamJournal(t *testing.T) {
	t.Run("CreateDreamJournal", func(t *testing.T) {
		now := time.Now()
		journal := &DreamJournal{
			ID:                   "test-dream-1",
			Timestamp:            now,
			DreamState:           "deep_consolidation",
			MemoriesConsolidated: 15,
			PatternsSynthesized:  3,
			Insights:             []string{"Pattern A emerged", "Connection B strengthened"},
			Metadata: map[string]interface{}{
				"duration": 3600,
			},
		}

		assert.Equal(t, "test-dream-1", journal.ID)
		assert.Equal(t, "deep_consolidation", journal.DreamState)
		assert.Equal(t, 15, journal.MemoriesConsolidated)
		assert.Equal(t, 3, journal.PatternsSynthesized)
		assert.Len(t, journal.Insights, 2)
	})
}

// BenchmarkMemoryStructs benchmarks memory struct operations
func BenchmarkMemoryStructs(b *testing.B) {
	b.Run("CreateMemoryNode", func(b *testing.B) {
		for range b.N {
			_ = &MemoryNode{
				ID:        "benchmark-node",
				Type:      NodeTypeEpisodic,
				Content:   "Benchmark content",
				Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5},
				Metadata: map[string]interface{}{
					"source": "benchmark",
				},
			}
		}
	})

	b.Run("CreateMemoryEdge", func(b *testing.B) {
		for range b.N {
			_ = &MemoryEdge{
				ID:       "benchmark-edge",
				SourceID: "node-a",
				TargetID: "node-b",
				Type:     EdgeTypeAssociation,
				Weight:   0.5,
			}
		}
	})

	b.Run("CreateHyperEdge", func(b *testing.B) {
		for range b.N {
			_ = &HyperEdge{
				ID:      "benchmark-hyperedge",
				NodeIDs: []string{"a", "b", "c", "d", "e"},
				Type:    "cluster",
			}
		}
	})
}
