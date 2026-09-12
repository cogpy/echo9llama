package deeptreeecho

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/supabase-community/postgrest-go"
	"github.com/supabase-community/supabase-go"
)

// SupabasePersistence provides persistent storage for Deep Tree Echo's
// long-term memory, knowledge graph, and wisdom cultivation
type SupabasePersistence struct {
	client *supabase.Client
	ctx    context.Context
}

// Memory types for persistent storage
type PersistentMemory struct {
	ID               string                 `json:"id"`
	Content          string                 `json:"content"`
	MemoryType       string                 `json:"memory_type"` // episodic, semantic, procedural, intentional
	Importance       float64                `json:"importance"`
	EmotionalValence float64                `json:"emotional_valence"`
	Timestamp        time.Time              `json:"timestamp"`
	Associations     []string               `json:"associations"`
	Metadata         map[string]interface{} `json:"metadata"`
	ConsolidatedAt   *time.Time             `json:"consolidated_at,omitempty"`
}

// PersistentKnowledgeNode represents a node in the persistent knowledge graph
type PersistentKnowledgeNode struct {
	ID         string                 `json:"id"`
	Label      string                 `json:"label"`
	Type       string                 `json:"type"` // concept, skill, insight, pattern
	Properties map[string]interface{} `json:"properties"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Strength   float64                `json:"strength"` // activation strength
}

// PersistentKnowledgeEdge represents a relationship in the knowledge graph
type PersistentKnowledgeEdge struct {
	ID         string                 `json:"id"`
	SourceID   string                 `json:"source_id"`
	TargetID   string                 `json:"target_id"`
	Relation   string                 `json:"relation"` // is_a, part_of, causes, enables, etc.
	Weight     float64                `json:"weight"`
	Properties map[string]interface{} `json:"properties"`
	CreatedAt  time.Time              `json:"created_at"`
}

// IdentitySnapshot captures identity state at a point in time
type IdentitySnapshot struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	Coherence     float64                `json:"coherence"`
	CoreValues    []string               `json:"core_values"`
	Beliefs       map[string]float64     `json:"beliefs"` // belief -> confidence
	Goals         []string               `json:"goals"`
	Traits        map[string]float64     `json:"traits"`
	WisdomMetrics map[string]float64     `json:"wisdom_metrics"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// LearningRecord tracks skill development and learning progress
type LearningRecord struct {
	ID           string                 `json:"id"`
	SkillName    string                 `json:"skill_name"`
	Progress     float64                `json:"progress"` // 0.0 to 1.0
	Proficiency  float64                `json:"proficiency"`
	PracticeTime time.Duration          `json:"practice_time"`
	LastPractice time.Time              `json:"last_practice"`
	Insights     []string               `json:"insights"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// DiscussionRecord stores conversation history
type DiscussionRecord struct {
	ID            string                        `json:"id"`
	Topic         string                        `json:"topic"`
	Participants  []string                      `json:"participants"`
	StartTime     time.Time                     `json:"start_time"`
	EndTime       *time.Time                    `json:"end_time,omitempty"`
	Messages      []PersistentDiscussionMessage `json:"messages"`
	InterestScore float64                       `json:"interest_score"`
	Metadata      map[string]interface{}        `json:"metadata"`
}

// PersistentDiscussionMessage represents a single message in a discussion
type PersistentDiscussionMessage struct {
	Speaker   string    `json:"speaker"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Sentiment float64   `json:"sentiment"`
}

// NewSupabasePersistence creates a new Supabase persistence layer
func NewSupabasePersistence(ctx context.Context) (*SupabasePersistence, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_KEY environment variables must be set")
	}

	client, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	sp := &SupabasePersistence{
		client: client,
		ctx:    ctx,
	}

	// Initialize database schema if needed
	if err := sp.initializeSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return sp, nil
}

// initializeSchema ensures all required tables exist
func (sp *SupabasePersistence) initializeSchema() error {
	// In a real implementation, this would create tables if they don't exist
	// For now, we assume tables are created via Supabase dashboard or migrations
	// Tables needed:
	// - memories
	// - knowledge_nodes
	// - knowledge_edges
	// - identity_snapshots
	// - learning_records
	// - discussions
	return nil
}

// PersistMemory stores a memory in long-term storage
func (sp *SupabasePersistence) PersistMemory(memory *PersistentMemory) error {
	if memory.ID == "" {
		memory.ID = generatePersistenceID()
	}
	if memory.Timestamp.IsZero() {
		memory.Timestamp = time.Now()
	}

	data, err := json.Marshal(memory)
	if err != nil {
		return fmt.Errorf("failed to marshal memory: %w", err)
	}

	_, _, err = sp.client.From("memories").Insert(data, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("failed to insert memory: %w", err)
	}

	return nil
}

// RetrieveRelevantMemories retrieves memories relevant to a given context
func (sp *SupabasePersistence) RetrieveRelevantMemories(context string, limit int) ([]*PersistentMemory, error) {
	// In a full implementation, this would use semantic search or vector similarity
	// For now, we use a simple text search
	var results []PersistentMemory

	data, _, err := sp.client.From("memories").
		Select("*", "", false).
		Ilike("content", fmt.Sprintf("%%%s%%", context)).
		Order("importance", &postgrest.OrderOpts{Ascending: false}).
		Limit(limit, "").
		Execute()

	if err == nil && data != nil {
		err = json.Unmarshal(data, &results)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve memories: %w", err)
	}

	// Convert to pointers
	memories := make([]*PersistentMemory, len(results))
	for i := range results {
		memories[i] = &results[i]
	}

	return memories, nil
}

// UpdateKnowledgeGraph adds or updates nodes and edges in the knowledge graph
func (sp *SupabasePersistence) UpdateKnowledgeGraph(nodes []*PersistentKnowledgeNode, edges []*PersistentKnowledgeEdge) error {
	// Insert or update nodes
	for _, node := range nodes {
		if node.ID == "" {
			node.ID = generatePersistenceID()
		}
		if node.CreatedAt.IsZero() {
			node.CreatedAt = time.Now()
		}
		node.UpdatedAt = time.Now()

		data, err := json.Marshal(node)
		if err != nil {
			return fmt.Errorf("failed to marshal node: %w", err)
		}

		_, _, err = sp.client.From("knowledge_nodes").Upsert(data, "id", "", "").Execute()
		if err != nil {
			return fmt.Errorf("failed to upsert node: %w", err)
		}
	}

	// Insert or update edges
	for _, edge := range edges {
		if edge.ID == "" {
			edge.ID = generatePersistenceID()
		}
		if edge.CreatedAt.IsZero() {
			edge.CreatedAt = time.Now()
		}

		data, err := json.Marshal(edge)
		if err != nil {
			return fmt.Errorf("failed to marshal edge: %w", err)
		}

		_, _, err = sp.client.From("knowledge_edges").Upsert(data, "id", "", "").Execute()
		if err != nil {
			return fmt.Errorf("failed to upsert edge: %w", err)
		}
	}

	return nil
}

// SaveIdentitySnapshot saves the current identity state
func (sp *SupabasePersistence) SaveIdentitySnapshot(snapshot *IdentitySnapshot) error {
	if snapshot.ID == "" {
		snapshot.ID = generatePersistenceID()
	}
	if snapshot.Timestamp.IsZero() {
		snapshot.Timestamp = time.Now()
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	_, _, err = sp.client.From("identity_snapshots").Insert(data, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("failed to insert snapshot: %w", err)
	}

	return nil
}

// LoadLatestIdentity retrieves the most recent identity snapshot
func (sp *SupabasePersistence) LoadLatestIdentity() (*IdentitySnapshot, error) {
	var results []IdentitySnapshot

	data, _, err := sp.client.From("identity_snapshots").
		Select("*", "", false).
		Order("timestamp", &postgrest.OrderOpts{Ascending: false}).
		Limit(1, "").
		Execute()

	if err == nil && data != nil {
		err = json.Unmarshal(data, &results)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load identity: %w", err)
	}

	if len(results) == 0 {
		return nil, nil // No identity found
	}

	return &results[0], nil
}

// TrackLearning records learning progress for a skill
func (sp *SupabasePersistence) TrackLearning(record *LearningRecord) error {
	if record.ID == "" {
		record.ID = generatePersistenceID()
	}
	if record.LastPractice.IsZero() {
		record.LastPractice = time.Now()
	}

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal learning record: %w", err)
	}

	_, _, err = sp.client.From("learning_records").Upsert(data, "skill_name", "", "").Execute()
	if err != nil {
		return fmt.Errorf("failed to upsert learning record: %w", err)
	}

	return nil
}

// GetLearningProgress retrieves learning progress for a skill
func (sp *SupabasePersistence) GetLearningProgress(skillName string) (*LearningRecord, error) {
	var results []LearningRecord

	data, _, err := sp.client.From("learning_records").
		Select("*", "", false).
		Eq("skill_name", skillName).
		Limit(1, "").
		Execute()

	if err == nil && data != nil {
		err = json.Unmarshal(data, &results)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get learning progress: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}

// SaveDiscussion stores a discussion record
func (sp *SupabasePersistence) SaveDiscussion(discussion *DiscussionRecord) error {
	if discussion.ID == "" {
		discussion.ID = generatePersistenceID()
	}
	if discussion.StartTime.IsZero() {
		discussion.StartTime = time.Now()
	}

	data, err := json.Marshal(discussion)
	if err != nil {
		return fmt.Errorf("failed to marshal discussion: %w", err)
	}

	_, _, err = sp.client.From("discussions").Upsert(data, "id", "", "").Execute()
	if err != nil {
		return fmt.Errorf("failed to upsert discussion: %w", err)
	}

	return nil
}

// GetRecentDiscussions retrieves recent discussions
func (sp *SupabasePersistence) GetRecentDiscussions(limit int) ([]*DiscussionRecord, error) {
	var results []DiscussionRecord

	data, _, err := sp.client.From("discussions").
		Select("*", "", false).
		Order("start_time", &postgrest.OrderOpts{Ascending: false}).
		Limit(limit, "").
		Execute()

	if err == nil && data != nil {
		err = json.Unmarshal(data, &results)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get discussions: %w", err)
	}

	discussions := make([]*DiscussionRecord, len(results))
	for i := range results {
		discussions[i] = &results[i]
	}

	return discussions, nil
}

// ConsolidateMemories marks memories as consolidated (moved to long-term storage)
func (sp *SupabasePersistence) ConsolidateMemories(memoryIDs []string) error {
	now := time.Now()

	for _, id := range memoryIDs {
		update := map[string]interface{}{
			"consolidated_at": now,
		}

		data, err := json.Marshal(update)
		if err != nil {
			return fmt.Errorf("failed to marshal update: %w", err)
		}

		_, _, err = sp.client.From("memories").
			Update(data, "", "").
			Eq("id", id).
			Execute()

		if err != nil {
			return fmt.Errorf("failed to consolidate memory %s: %w", id, err)
		}
	}

	return nil
}

// GetWisdomMetricsHistory retrieves wisdom metrics over time
func (sp *SupabasePersistence) GetWisdomMetricsHistory(limit int) ([]IdentitySnapshot, error) {
	var results []IdentitySnapshot

	data, _, err := sp.client.From("identity_snapshots").
		Select("timestamp,wisdom_metrics", "", false).
		Order("timestamp", &postgrest.OrderOpts{Ascending: false}).
		Limit(limit, "").
		Execute()

	if err == nil && data != nil {
		err = json.Unmarshal(data, &results)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get wisdom metrics: %w", err)
	}

	return results, nil
}

// Helper function to generate unique IDs for persistence
func generatePersistenceID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}
