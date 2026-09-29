package echodream

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PersistentMemory manages long-term memory storage and retrieval
type PersistentMemory struct {
	StoragePath    string
	Memories       map[string]*MemoryRecord
	MaxMemories    int
	PruneThreshold float64
	mu             sync.RWMutex
}

// MemoryRecord represents a single persistent memory entry
type MemoryRecord struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Content      string                 `json:"content"`
	Importance   float64                `json:"importance"`
	CreatedAt    time.Time              `json:"created_at"`
	LastAccessed time.Time              `json:"last_accessed"`
	AccessCount  int                    `json:"access_count"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// NewPersistentMemory creates a new persistent memory system
func NewPersistentMemory(storagePath string) *PersistentMemory {
	pm := &PersistentMemory{
		StoragePath:    storagePath,
		Memories:       make(map[string]*MemoryRecord),
		MaxMemories:    10000,
		PruneThreshold: 0.3,
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storagePath, 0o755); err != nil {
		log.Printf("⚠️  PersistentMemory: Failed to create storage directory: %v", err)
	}

	// Load existing memories
	pm.Load()

	return pm
}

// Store saves a new memory
func (pm *PersistentMemory) Store(memType string, content string, importance float64, tags []string) (*MemoryRecord, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if we need to prune
	if len(pm.Memories) >= pm.MaxMemories {
		pm.pruneMemories()
	}

	record := &MemoryRecord{
		ID:           fmt.Sprintf("mem_%d_%d", time.Now().Unix(), time.Now().Nanosecond()),
		Type:         memType,
		Content:      content,
		Importance:   importance,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		AccessCount:  0,
		Tags:         tags,
		Metadata:     make(map[string]interface{}),
	}

	pm.Memories[record.ID] = record
	log.Printf("💾 PersistentMemory: Stored %s memory (importance: %.2f)", memType, importance)

	return record, nil
}

// Retrieve gets a memory by ID
func (pm *PersistentMemory) Retrieve(id string) (*MemoryRecord, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	record, exists := pm.Memories[id]
	if !exists {
		return nil, fmt.Errorf("memory not found: %s", id)
	}

	// Update access information
	record.LastAccessed = time.Now()
	record.AccessCount++

	return record, nil
}

// Search finds memories by tag or type
func (pm *PersistentMemory) Search(query string) []*MemoryRecord {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	results := make([]*MemoryRecord, 0)

	for _, record := range pm.Memories {
		// Check tags
		for _, tag := range record.Tags {
			if tag == query {
				results = append(results, record)
				break
			}
		}

		// Check type
		if record.Type == query {
			results = append(results, record)
		}
	}

	return results
}

// GetByType returns all memories of a specific type
func (pm *PersistentMemory) GetByType(memType string) []*MemoryRecord {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	results := make([]*MemoryRecord, 0)

	for _, record := range pm.Memories {
		if record.Type == memType {
			results = append(results, record)
		}
	}

	return results
}

// Update modifies an existing memory
func (pm *PersistentMemory) Update(id string, content string, importance float64) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	record, exists := pm.Memories[id]
	if !exists {
		return fmt.Errorf("memory not found: %s", id)
	}

	record.Content = content
	record.Importance = importance
	record.LastAccessed = time.Now()

	return nil
}

// Delete removes a memory
func (pm *PersistentMemory) Delete(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.Memories[id]; !exists {
		return fmt.Errorf("memory not found: %s", id)
	}

	delete(pm.Memories, id)
	log.Printf("🗑️  PersistentMemory: Deleted memory %s", id)

	return nil
}

// pruneMemories removes low-importance memories
func (pm *PersistentMemory) pruneMemories() {
	// Find the memory with the lowest score
	var lowestID string
	lowestScore := 1.0

	for id, record := range pm.Memories {
		// Calculate score based on importance and recency
		ageHours := time.Since(record.LastAccessed).Hours()
		score := record.Importance / (1.0 + ageHours/24.0)

		if score < lowestScore {
			lowestScore = score
			lowestID = id
		}
	}

	if lowestID != "" && lowestScore < pm.PruneThreshold {
		delete(pm.Memories, lowestID)
		log.Printf("🗑️  PersistentMemory: Pruned memory %s (score: %.2f)", lowestID, lowestScore)
	}
}

// Save persists all memories to disk
func (pm *PersistentMemory) Save() error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	filePath := filepath.Join(pm.StoragePath, "persistent_memories.json")

	data, err := json.MarshalIndent(pm.Memories, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal memories: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write memories file: %v", err)
	}

	log.Printf("💾 PersistentMemory: Saved %d memories to disk", len(pm.Memories))

	return nil
}

// Load restores memories from disk
func (pm *PersistentMemory) Load() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	filePath := filepath.Join(pm.StoragePath, "persistent_memories.json")

	if _, err := os.Stat(filePath); err != nil {
		// File doesn't exist yet, which is fine
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read memories file: %v", err)
	}

	if err := json.Unmarshal(data, &pm.Memories); err != nil {
		return fmt.Errorf("failed to unmarshal memories: %v", err)
	}

	log.Printf("📖 PersistentMemory: Loaded %d memories from disk", len(pm.Memories))

	return nil
}

// GetStats returns statistics about the memory system
func (pm *PersistentMemory) GetStats() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Count by type
	typeCount := make(map[string]int)
	totalImportance := 0.0
	oldestAccess := time.Now()
	newestAccess := time.Time{}

	for _, record := range pm.Memories {
		typeCount[record.Type]++
		totalImportance += record.Importance

		if record.LastAccessed.Before(oldestAccess) {
			oldestAccess = record.LastAccessed
		}
		if record.LastAccessed.After(newestAccess) {
			newestAccess = record.LastAccessed
		}
	}

	avgImportance := 0.0
	if len(pm.Memories) > 0 {
		avgImportance = totalImportance / float64(len(pm.Memories))
	}

	return map[string]interface{}{
		"total_memories":     len(pm.Memories),
		"memories_by_type":   typeCount,
		"average_importance": avgImportance,
		"oldest_access":      oldestAccess.Format(time.RFC3339),
		"newest_access":      newestAccess.Format(time.RFC3339),
		"storage_path":       pm.StoragePath,
		"max_memories":       pm.MaxMemories,
		"prune_threshold":    pm.PruneThreshold,
	}
}
