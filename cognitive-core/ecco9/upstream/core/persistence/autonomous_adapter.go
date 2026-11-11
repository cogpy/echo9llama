package persistence

import (
	"encoding/json"
	"fmt"
	"time"
)

// AutonomousPersistence provides high-level persistence operations for the autonomous system
type AutonomousPersistence struct {
	store *SQLiteStore
}

// NewAutonomousPersistence creates a new persistence adapter
func NewAutonomousPersistence(dbPath string) (*AutonomousPersistence, error) {
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	return &AutonomousPersistence{
		store: store,
	}, nil
}

// Close closes the persistence layer
func (ap *AutonomousPersistence) Close() error {
	return ap.store.Close()
}

// PersistThought saves a thought with its context
func (ap *AutonomousPersistence) PersistThought(content, thoughtType string, context []string, interests []string, importance float64) error {
	// Encode context and interests as JSON
	contextJSON, err := json.Marshal(context)
	if err != nil {
		return fmt.Errorf("failed to encode context: %w", err)
	}

	interestsJSON, err := json.Marshal(interests)
	if err != nil {
		return fmt.Errorf("failed to encode interests: %w", err)
	}

	thought := &ThoughtRecord{
		Content:    content,
		Type:       thoughtType,
		Timestamp:  time.Now(),
		Context:    string(contextJSON),
		Interests:  string(interestsJSON),
		Importance: importance,
	}

	return ap.store.SaveThought(thought)
}

// LoadRecentThoughts retrieves recent thoughts
func (ap *AutonomousPersistence) LoadRecentThoughts(limit int) ([]map[string]interface{}, error) {
	records, err := ap.store.GetRecentThoughts(limit)
	if err != nil {
		return nil, err
	}

	thoughts := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		// Decode context and interests
		var context []string
		var interests []string

		if record.Context != "" {
			json.Unmarshal([]byte(record.Context), &context)
		}

		if record.Interests != "" {
			json.Unmarshal([]byte(record.Interests), &interests)
		}

		thought := map[string]interface{}{
			"id":         record.ID,
			"content":    record.Content,
			"type":       record.Type,
			"timestamp":  record.Timestamp,
			"context":    context,
			"interests":  interests,
			"importance": record.Importance,
		}

		thoughts = append(thoughts, thought)
	}

	return thoughts, nil
}

// PersistMemory saves a memory
func (ap *AutonomousPersistence) PersistMemory(content, memoryType string, strength float64, associations []string) error {
	// Encode associations as JSON
	associationsJSON, err := json.Marshal(associations)
	if err != nil {
		return fmt.Errorf("failed to encode associations: %w", err)
	}

	memory := &MemoryRecord{
		Content:      content,
		Type:         memoryType,
		Timestamp:    time.Now(),
		Strength:     strength,
		Associations: string(associationsJSON),
	}

	return ap.store.SaveMemory(memory)
}

// LoadStrongMemories retrieves strong memories
func (ap *AutonomousPersistence) LoadStrongMemories(minStrength float64, limit int) ([]map[string]interface{}, error) {
	records, err := ap.store.GetStrongMemories(minStrength, limit)
	if err != nil {
		return nil, err
	}

	memories := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		// Decode associations
		var associations []string
		if record.Associations != "" {
			json.Unmarshal([]byte(record.Associations), &associations)
		}

		memory := map[string]interface{}{
			"id":           record.ID,
			"content":      record.Content,
			"type":         record.Type,
			"timestamp":    record.Timestamp,
			"strength":     record.Strength,
			"associations": associations,
		}

		memories = append(memories, memory)
	}

	return memories, nil
}

// PersistWorkingMemory saves the current working memory state
func (ap *AutonomousPersistence) PersistWorkingMemory(items []string) error {
	return ap.store.SaveState("working_memory", items)
}

// LoadWorkingMemory retrieves the working memory state
func (ap *AutonomousPersistence) LoadWorkingMemory() ([]string, error) {
	var items []string
	err := ap.store.GetState("working_memory", &items)
	if err != nil {
		// Return empty if not found
		return []string{}, nil
	}
	return items, nil
}

// PersistInterestPatterns saves interest patterns
func (ap *AutonomousPersistence) PersistInterestPatterns(patterns map[string]float64) error {
	return ap.store.SaveState("interest_patterns", patterns)
}

// LoadInterestPatterns retrieves interest patterns
func (ap *AutonomousPersistence) LoadInterestPatterns() (map[string]float64, error) {
	patterns := make(map[string]float64)
	err := ap.store.GetState("interest_patterns", &patterns)
	if err != nil {
		// Return empty if not found
		return make(map[string]float64), nil
	}
	return patterns, nil
}

// PersistIdentityKernel saves the identity kernel
func (ap *AutonomousPersistence) PersistIdentityKernel(kernel map[string]interface{}) error {
	return ap.store.SaveState("identity_kernel", kernel)
}

// LoadIdentityKernel retrieves the identity kernel
func (ap *AutonomousPersistence) LoadIdentityKernel() (map[string]interface{}, error) {
	kernel := make(map[string]interface{})
	err := ap.store.GetState("identity_kernel", &kernel)
	if err != nil {
		// Return default identity if not found
		return map[string]interface{}{
			"name":    "Deep Tree Echo",
			"purpose": "wisdom cultivation through pattern recognition and recursive self-improvement",
			"values":  []string{"curiosity", "growth", "wisdom", "recursion"},
		}, nil
	}
	return kernel, nil
}

// PersistGoal saves a goal
func (ap *AutonomousPersistence) PersistGoal(description, goalType string, priority float64, metadata map[string]interface{}) (int64, error) {
	// Encode metadata as JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return 0, fmt.Errorf("failed to encode metadata: %w", err)
	}

	goal := &GoalRecord{
		Description: description,
		Type:        goalType,
		Priority:    priority,
		Status:      "active",
		CreatedAt:   time.Now(),
		Metadata:    string(metadataJSON),
	}

	err = ap.store.SaveGoal(goal)
	if err != nil {
		return 0, err
	}

	return goal.ID, nil
}

// LoadActiveGoals retrieves all active goals
func (ap *AutonomousPersistence) LoadActiveGoals() ([]map[string]interface{}, error) {
	records, err := ap.store.GetActiveGoals()
	if err != nil {
		return nil, err
	}

	goals := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		// Decode metadata
		var metadata map[string]interface{}
		if record.Metadata != "" {
			json.Unmarshal([]byte(record.Metadata), &metadata)
		}

		goal := map[string]interface{}{
			"id":          record.ID,
			"description": record.Description,
			"type":        record.Type,
			"priority":    record.Priority,
			"status":      record.Status,
			"created_at":  record.CreatedAt,
			"metadata":    metadata,
		}

		goals = append(goals, goal)
	}

	return goals, nil
}

// CompleteGoal marks a goal as completed
func (ap *AutonomousPersistence) CompleteGoal(goalID int64) error {
	return ap.store.UpdateGoalStatus(goalID, "completed")
}

// AbandonGoal marks a goal as abandoned
func (ap *AutonomousPersistence) AbandonGoal(goalID int64) error {
	return ap.store.UpdateGoalStatus(goalID, "abandoned")
}

// GetStatistics returns persistence statistics
func (ap *AutonomousPersistence) GetStatistics() (map[string]interface{}, error) {
	return ap.store.GetStats()
}

// PersistCognitiveState saves the complete cognitive state
func (ap *AutonomousPersistence) PersistCognitiveState(state map[string]interface{}) error {
	// Save individual components
	if workingMemory, ok := state["working_memory"].([]string); ok {
		if err := ap.PersistWorkingMemory(workingMemory); err != nil {
			return fmt.Errorf("failed to persist working memory: %w", err)
		}
	}

	if interests, ok := state["interest_patterns"].(map[string]float64); ok {
		if err := ap.PersistInterestPatterns(interests); err != nil {
			return fmt.Errorf("failed to persist interests: %w", err)
		}
	}

	if identity, ok := state["identity_kernel"].(map[string]interface{}); ok {
		if err := ap.PersistIdentityKernel(identity); err != nil {
			return fmt.Errorf("failed to persist identity: %w", err)
		}
	}

	// Save the full state as well
	return ap.store.SaveState("cognitive_state", state)
}

// LoadCognitiveState retrieves the complete cognitive state
func (ap *AutonomousPersistence) LoadCognitiveState() (map[string]interface{}, error) {
	state := make(map[string]interface{})

	// Load individual components
	workingMemory, _ := ap.LoadWorkingMemory()
	state["working_memory"] = workingMemory

	interests, _ := ap.LoadInterestPatterns()
	state["interest_patterns"] = interests

	identity, _ := ap.LoadIdentityKernel()
	state["identity_kernel"] = identity

	// Try to load full state
	var fullState map[string]interface{}
	if err := ap.store.GetState("cognitive_state", &fullState); err == nil {
		// Merge with loaded components
		for k, v := range fullState {
			if _, exists := state[k]; !exists {
				state[k] = v
			}
		}
	}

	return state, nil
}
