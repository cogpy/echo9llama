package lampstack

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// UnifiedMemoryStore provides a unified memory system combining
// hypergraph, episodic, and semantic memory types
type UnifiedMemoryStore struct {
	mu sync.RWMutex

	config *Config

	// Memory storage by type
	episodic   map[string]*MemoryEntry
	semantic   map[string]*MemoryEntry
	procedural map[string]*MemoryEntry
	working    map[string]*MemoryEntry

	// Indices for fast lookup
	byTag      map[string][]string // tag -> memory IDs
	byTime     []*MemoryEntry      // sorted by creation time
	importance []*MemoryEntry      // sorted by importance

	// Embeddings cache
	embeddings map[string][]float64

	// Stats
	stats MemoryStats
}

// MemoryStats tracks memory system statistics
type MemoryStats struct {
	TotalMemories     int64
	EpisodicCount     int64
	SemanticCount     int64
	ProceduralCount   int64
	WorkingCount      int64
	Consolidations    int64
	LastConsolidation time.Time
}

// NewUnifiedMemoryStore creates a new unified memory store
func NewUnifiedMemoryStore(config *Config) *UnifiedMemoryStore {
	return &UnifiedMemoryStore{
		config: config,

		episodic:   make(map[string]*MemoryEntry),
		semantic:   make(map[string]*MemoryEntry),
		procedural: make(map[string]*MemoryEntry),
		working:    make(map[string]*MemoryEntry),

		byTag:      make(map[string][]string),
		byTime:     make([]*MemoryEntry, 0),
		importance: make([]*MemoryEntry, 0),

		embeddings: make(map[string][]float64),
	}
}

// Store saves a memory entry
func (m *UnifiedMemoryStore) Store(ctx context.Context, entry *MemoryEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate ID if not set
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	// Set timestamps
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	entry.AccessedAt = time.Now()

	// Store in appropriate type map
	store := m.getStoreForType(entry.Type)
	store[entry.ID] = entry

	// Update indices
	m.updateIndices(entry)

	// Update stats
	m.updateStats(entry.Type, 1)

	// Check capacity
	if int(m.stats.TotalMemories) > m.config.MemoryCapacity {
		m.pruneOldMemories()
	}

	return nil
}

// Retrieve finds memories relevant to a query
func (m *UnifiedMemoryStore) Retrieve(ctx context.Context, query string, limit int) ([]*MemoryEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 {
		limit = 10
	}

	// Score all memories by relevance
	scored := make([]*scoredMemory, 0)

	for _, entry := range m.episodic {
		score := m.calculateRelevance(entry, query)
		if score > 0.1 {
			scored = append(scored, &scoredMemory{entry: entry, score: score})
		}
	}

	for _, entry := range m.semantic {
		score := m.calculateRelevance(entry, query)
		if score > 0.1 {
			scored = append(scored, &scoredMemory{entry: entry, score: score})
		}
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Return top results
	results := make([]*MemoryEntry, 0, limit)
	for i, sm := range scored {
		if i >= limit {
			break
		}
		// Update access count
		sm.entry.AccessCount++
		sm.entry.AccessedAt = time.Now()
		results = append(results, sm.entry)
	}

	return results, nil
}

// Get retrieves a specific memory by ID
func (m *UnifiedMemoryStore) Get(ctx context.Context, id string) (*MemoryEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check all stores
	if entry, ok := m.episodic[id]; ok {
		entry.AccessCount++
		entry.AccessedAt = time.Now()
		return entry, nil
	}
	if entry, ok := m.semantic[id]; ok {
		entry.AccessCount++
		entry.AccessedAt = time.Now()
		return entry, nil
	}
	if entry, ok := m.procedural[id]; ok {
		entry.AccessCount++
		entry.AccessedAt = time.Now()
		return entry, nil
	}
	if entry, ok := m.working[id]; ok {
		entry.AccessCount++
		entry.AccessedAt = time.Now()
		return entry, nil
	}

	return nil, fmt.Errorf("memory not found: %s", id)
}

// Update modifies an existing memory
func (m *UnifiedMemoryStore) Update(ctx context.Context, entry *MemoryEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	store := m.getStoreForType(entry.Type)
	if _, ok := store[entry.ID]; !ok {
		return fmt.Errorf("memory not found: %s", entry.ID)
	}

	entry.AccessedAt = time.Now()
	store[entry.ID] = entry
	m.updateIndices(entry)

	return nil
}

// Delete removes a memory
func (m *UnifiedMemoryStore) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check all stores
	for _, store := range []map[string]*MemoryEntry{m.episodic, m.semantic, m.procedural, m.working} {
		if entry, ok := store[id]; ok {
			// Remove from indices
			m.removeFromIndices(entry)
			delete(store, id)
			m.stats.TotalMemories--
			return nil
		}
	}

	return fmt.Errorf("memory not found: %s", id)
}

// RetrieveByType retrieves memories of a specific type
func (m *UnifiedMemoryStore) RetrieveByType(ctx context.Context, memType MemoryType, limit int) ([]*MemoryEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	store := m.getStoreForType(memType)
	results := make([]*MemoryEntry, 0, limit)

	// Convert to slice and sort by access time
	entries := make([]*MemoryEntry, 0, len(store))
	for _, entry := range store {
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].AccessedAt.After(entries[j].AccessedAt)
	})

	for i, entry := range entries {
		if i >= limit {
			break
		}
		results = append(results, entry)
	}

	return results, nil
}

// RetrieveByTags retrieves memories with specific tags
func (m *UnifiedMemoryStore) RetrieveByTags(ctx context.Context, tags []string, limit int) ([]*MemoryEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Find memories with any of the tags
	idSet := make(map[string]int) // ID -> count of matching tags

	for _, tag := range tags {
		if ids, ok := m.byTag[tag]; ok {
			for _, id := range ids {
				idSet[id]++
			}
		}
	}

	// Get entries sorted by matching tag count
	type idScore struct {
		id    string
		score int
	}
	scored := make([]idScore, 0, len(idSet))
	for id, count := range idSet {
		scored = append(scored, idScore{id, count})
	}
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Retrieve actual entries
	results := make([]*MemoryEntry, 0, limit)
	for i, s := range scored {
		if i >= limit {
			break
		}
		entry, err := m.Get(ctx, s.id)
		if err == nil {
			results = append(results, entry)
		}
	}

	return results, nil
}

// RetrieveRecent retrieves the most recently accessed memories
func (m *UnifiedMemoryStore) RetrieveRecent(ctx context.Context, limit int) ([]*MemoryEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect all memories
	all := make([]*MemoryEntry, 0)
	for _, entry := range m.episodic {
		all = append(all, entry)
	}
	for _, entry := range m.semantic {
		all = append(all, entry)
	}
	for _, entry := range m.procedural {
		all = append(all, entry)
	}
	for _, entry := range m.working {
		all = append(all, entry)
	}

	// Sort by access time
	sort.Slice(all, func(i, j int) bool {
		return all[i].AccessedAt.After(all[j].AccessedAt)
	})

	if len(all) > limit {
		all = all[:limit]
	}

	return all, nil
}

// Consolidate performs memory consolidation
func (m *UnifiedMemoryStore) Consolidate(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Move frequently accessed episodic memories to semantic
	for id, entry := range m.episodic {
		if entry.AccessCount > 5 && entry.Importance > 0.6 {
			// Promote to semantic memory
			entry.Type = MemoryTypeSemantic
			m.semantic[id] = entry
			delete(m.episodic, id)
			m.stats.EpisodicCount--
			m.stats.SemanticCount++
		}
	}

	// Clear old working memory
	now := time.Now()
	for id, entry := range m.working {
		if now.Sub(entry.AccessedAt) > 10*time.Minute {
			delete(m.working, id)
			m.stats.WorkingCount--
			m.stats.TotalMemories--
		}
	}

	m.stats.Consolidations++
	m.stats.LastConsolidation = time.Now()

	return nil
}

// GetStats returns memory system statistics
func (m *UnifiedMemoryStore) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_memories":      m.stats.TotalMemories,
		"episodic_count":      m.stats.EpisodicCount,
		"semantic_count":      m.stats.SemanticCount,
		"procedural_count":    m.stats.ProceduralCount,
		"working_count":       m.stats.WorkingCount,
		"consolidations":      m.stats.Consolidations,
		"last_consolidation":  m.stats.LastConsolidation,
		"capacity":            m.config.MemoryCapacity,
		"utilization_percent": float64(m.stats.TotalMemories) / float64(m.config.MemoryCapacity) * 100,
	}
}

// Helper types

type scoredMemory struct {
	entry *MemoryEntry
	score float64
}

// Helper methods

func (m *UnifiedMemoryStore) getStoreForType(memType MemoryType) map[string]*MemoryEntry {
	switch memType {
	case MemoryTypeEpisodic:
		return m.episodic
	case MemoryTypeSemantic:
		return m.semantic
	case MemoryTypeProcedural:
		return m.procedural
	case MemoryTypeWorking:
		return m.working
	default:
		return m.episodic
	}
}

func (m *UnifiedMemoryStore) updateIndices(entry *MemoryEntry) {
	// Update tag index
	for _, tag := range entry.Tags {
		m.byTag[tag] = append(m.byTag[tag], entry.ID)
	}

	// Update time index
	m.byTime = append(m.byTime, entry)
	sort.Slice(m.byTime, func(i, j int) bool {
		return m.byTime[i].CreatedAt.After(m.byTime[j].CreatedAt)
	})

	// Update importance index
	m.importance = append(m.importance, entry)
	sort.Slice(m.importance, func(i, j int) bool {
		return m.importance[i].Importance > m.importance[j].Importance
	})
}

func (m *UnifiedMemoryStore) removeFromIndices(entry *MemoryEntry) {
	// Remove from tag index
	for _, tag := range entry.Tags {
		if ids, ok := m.byTag[tag]; ok {
			newIDs := make([]string, 0, len(ids)-1)
			for _, id := range ids {
				if id != entry.ID {
					newIDs = append(newIDs, id)
				}
			}
			m.byTag[tag] = newIDs
		}
	}

	// Remove from time index
	newByTime := make([]*MemoryEntry, 0, len(m.byTime)-1)
	for _, e := range m.byTime {
		if e.ID != entry.ID {
			newByTime = append(newByTime, e)
		}
	}
	m.byTime = newByTime

	// Remove from importance index
	newImportance := make([]*MemoryEntry, 0, len(m.importance)-1)
	for _, e := range m.importance {
		if e.ID != entry.ID {
			newImportance = append(newImportance, e)
		}
	}
	m.importance = newImportance
}

func (m *UnifiedMemoryStore) updateStats(memType MemoryType, delta int64) {
	m.stats.TotalMemories += delta
	switch memType {
	case MemoryTypeEpisodic:
		m.stats.EpisodicCount += delta
	case MemoryTypeSemantic:
		m.stats.SemanticCount += delta
	case MemoryTypeProcedural:
		m.stats.ProceduralCount += delta
	case MemoryTypeWorking:
		m.stats.WorkingCount += delta
	}
}

func (m *UnifiedMemoryStore) calculateRelevance(entry *MemoryEntry, query string) float64 {
	// Simple text similarity
	contentSim := similarity(entry.Content, query)

	// Recency boost
	age := time.Since(entry.AccessedAt).Hours()
	recencyBoost := 1.0 / (1.0 + age/24.0) // Decay over days

	// Importance factor
	importanceFactor := entry.Importance

	// Access frequency factor
	accessFactor := float64(entry.AccessCount) / (float64(entry.AccessCount) + 10.0)

	// Combined score
	return contentSim*0.5 + recencyBoost*0.2 + importanceFactor*0.2 + accessFactor*0.1
}

func (m *UnifiedMemoryStore) pruneOldMemories() {
	// Remove least important, oldest memories
	target := m.config.MemoryCapacity / 10 // Remove 10%

	// Start with working memory
	if len(m.working) > 0 {
		for id := range m.working {
			if target <= 0 {
				break
			}
			delete(m.working, id)
			m.stats.WorkingCount--
			m.stats.TotalMemories--
			target--
		}
	}

	// Then episodic memory
	if target > 0 && len(m.byTime) > 0 {
		// Get oldest entries
		for i := len(m.byTime) - 1; i >= 0 && target > 0; i-- {
			entry := m.byTime[i]
			if entry.Importance < 0.5 && entry.AccessCount < 3 {
				store := m.getStoreForType(entry.Type)
				delete(store, entry.ID)
				m.removeFromIndices(entry)
				m.updateStats(entry.Type, -1)
				target--
			}
		}
	}
}

// AddRelation creates a relation between two memories (hypergraph edge)
func (m *UnifiedMemoryStore) AddRelation(ctx context.Context, fromID, toID, relationType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	from, err := m.Get(ctx, fromID)
	if err != nil {
		return fmt.Errorf("source memory not found: %w", err)
	}

	_, err = m.Get(ctx, toID)
	if err != nil {
		return fmt.Errorf("target memory not found: %w", err)
	}

	// Add relation as association
	if from.Associations == nil {
		from.Associations = make([]string, 0)
	}
	from.Associations = append(from.Associations, fmt.Sprintf("%s:%s", relationType, toID))

	return nil
}

// GetRelations returns all relations for a memory
func (m *UnifiedMemoryStore) GetRelations(ctx context.Context, id string) (map[string][]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, err := m.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	relations := make(map[string][]string)
	for _, assoc := range entry.Associations {
		// Parse "relationType:targetID" format
		for i, c := range assoc {
			if c == ':' {
				relType := assoc[:i]
				targetID := assoc[i+1:]
				relations[relType] = append(relations[relType], targetID)
				break
			}
		}
	}

	return relations, nil
}
