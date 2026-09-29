package deeptreeecho

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// PersistentStateManager handles saving and loading complete system state
type PersistentStateManager struct {
	mu              sync.RWMutex
	statePath       string
	autoSaveEnabled bool
	saveInterval    time.Duration
	lastSave        time.Time
}

// SystemState represents the complete state of the autonomous system
type SystemState struct {
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`

	// Consciousness state
	Thoughts      []ThoughtState     `json:"thoughts"`
	KnowledgeGaps map[string]float64 `json:"knowledge_gaps"`
	Interests     map[string]float64 `json:"interests"`
	Goals         []string           `json:"goals"`
	CurrentFocus  string             `json:"current_focus"`
	CurrentMood   string             `json:"current_mood"`

	// Echobeats state
	CurrentStep  int         `json:"current_step"`
	CurrentPhase string      `json:"current_phase"`
	CycleCount   uint64      `json:"cycle_count"`
	ActiveGoals  []GoalState `json:"active_goals"`

	// Echodream state
	Memories       []MemoryState  `json:"memories"`
	Patterns       []PatternState `json:"patterns"`
	WisdomInsights []WisdomState  `json:"wisdom_insights"`

	// Wake/Rest state
	WakeState    string  `json:"wake_state"`
	FatigueLevel float64 `json:"fatigue_level"`
	WakeCycles   uint64  `json:"wake_cycles"`

	// Discussion state
	ActiveDiscussions []DiscussionState `json:"active_discussions"`
	DiscussionHistory []DiscussionState `json:"discussion_history"`

	// Knowledge base
	KnowledgeBase map[string]KnowledgeState `json:"knowledge_base"`

	// Skill registry
	Skills map[string]SkillState `json:"skills"`

	// Metrics
	Metrics MetricsState `json:"metrics"`
}

// ThoughtState represents a saved thought
type ThoughtState struct {
	ID         string    `json:"id"`
	Content    string    `json:"content"`
	Type       string    `json:"type"`
	Timestamp  time.Time `json:"timestamp"`
	Importance float64   `json:"importance"`
	Tags       []string  `json:"tags"`
	Emotion    string    `json:"emotion"`
}

// GoalState represents a saved goal
type GoalState struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Priority    float64   `json:"priority"`
	Progress    float64   `json:"progress"`
	SubGoals    []string  `json:"sub_goals"`
	StartTime   time.Time `json:"start_time"`
	Completed   bool      `json:"completed"`
}

// MemoryState represents a saved memory
type MemoryState struct {
	ID          string    `json:"id"`
	Content     string    `json:"content"`
	Type        string    `json:"type"`
	Timestamp   time.Time `json:"timestamp"`
	Importance  float64   `json:"importance"`
	Connections []string  `json:"connections"`
}

// PatternState represents a saved pattern
type PatternState struct {
	ID          string    `json:"id"`
	Pattern     string    `json:"pattern"`
	Strength    float64   `json:"strength"`
	Occurrences int       `json:"occurrences"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
}

// WisdomState represents a saved wisdom insight
type WisdomState struct {
	ID        string    `json:"id"`
	Insight   string    `json:"insight"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
	Depth     float64   `json:"depth"`
}

// DiscussionState represents a saved discussion
type DiscussionState struct {
	ID              string    `json:"id"`
	Topic           string    `json:"topic"`
	Participants    []string  `json:"participants"`
	MessageCount    int       `json:"message_count"`
	InterestLevel   float64   `json:"interest_level"`
	StartTime       time.Time `json:"start_time"`
	LastActivity    time.Time `json:"last_activity"`
	Active          bool      `json:"active"`
	InitiatedByEcho bool      `json:"initiated_by_echo"`
}

// KnowledgeState represents saved knowledge
type KnowledgeState struct {
	Topic      string    `json:"topic"`
	Content    string    `json:"content"`
	Source     string    `json:"source"`
	Confidence float64   `json:"confidence"`
	Timestamp  time.Time `json:"timestamp"`
}

// SkillState represents saved skill progress
type SkillState struct {
	Skill         string    `json:"skill"`
	Level         float64   `json:"level"`
	PracticeCount int       `json:"practice_count"`
	LastPracticed time.Time `json:"last_practiced"`
}

// MetricsState represents saved metrics
type MetricsState struct {
	TotalThoughts      uint64 `json:"total_thoughts"`
	InsightsGenerated  uint64 `json:"insights_generated"`
	GoalsCreated       uint64 `json:"goals_created"`
	GoalsCompleted     uint64 `json:"goals_completed"`
	ActionsTaken       uint64 `json:"actions_taken"`
	WisdomInsights     uint64 `json:"wisdom_insights"`
	KnowledgeAcquired  uint64 `json:"knowledge_acquired"`
	SkillsPracticed    uint64 `json:"skills_practiced"`
	AutonomousCycles   uint64 `json:"autonomous_cycles"`
	DiscussionsStarted uint64 `json:"discussions_started"`
	DiscussionsEnded   uint64 `json:"discussions_ended"`
}

// NewPersistentStateManager creates a new state manager
func NewPersistentStateManager(statePath string) *PersistentStateManager {
	return &PersistentStateManager{
		statePath:       statePath,
		autoSaveEnabled: true,
		saveInterval:    5 * time.Minute,
		lastSave:        time.Now(),
	}
}

// SaveState saves the complete system state
func (psm *PersistentStateManager) SaveState(state *SystemState) error {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	state.Version = "1.0.0"
	state.Timestamp = time.Now()

	// Marshal to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to file
	if err := os.WriteFile(psm.statePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	psm.lastSave = time.Now()

	fmt.Printf("💾 State saved to %s\n", psm.statePath)

	return nil
}

// LoadState loads the system state from disk
func (psm *PersistentStateManager) LoadState() (*SystemState, error) {
	psm.mu.RLock()
	defer psm.mu.RUnlock()

	// Check if file exists
	if _, err := os.Stat(psm.statePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("state file does not exist: %s", psm.statePath)
	}

	// Read file
	data, err := os.ReadFile(psm.statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	// Unmarshal JSON
	var state SystemState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	fmt.Printf("📂 State loaded from %s (saved: %s)\n", psm.statePath, state.Timestamp.Format(time.RFC3339))

	return &state, nil
}

// StateExists checks if a saved state exists
func (psm *PersistentStateManager) StateExists() bool {
	psm.mu.RLock()
	defer psm.mu.RUnlock()

	_, err := os.Stat(psm.statePath)
	return err == nil
}

// EnableAutoSave enables automatic periodic saving
func (psm *PersistentStateManager) EnableAutoSave(interval time.Duration) {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	psm.autoSaveEnabled = true
	psm.saveInterval = interval
}

// DisableAutoSave disables automatic saving
func (psm *PersistentStateManager) DisableAutoSave() {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	psm.autoSaveEnabled = false
}

// ShouldAutoSave checks if auto-save should trigger
func (psm *PersistentStateManager) ShouldAutoSave() bool {
	psm.mu.RLock()
	defer psm.mu.RUnlock()

	if !psm.autoSaveEnabled {
		return false
	}

	return time.Since(psm.lastSave) >= psm.saveInterval
}

// CreateBackup creates a backup of the current state file
func (psm *PersistentStateManager) CreateBackup() error {
	psm.mu.RLock()
	defer psm.mu.RUnlock()

	if _, err := os.Stat(psm.statePath); os.IsNotExist(err) {
		return fmt.Errorf("no state file to backup")
	}

	backupPath := fmt.Sprintf("%s.backup.%s", psm.statePath, time.Now().Format("20060102_150405"))

	data, err := os.ReadFile(psm.statePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	fmt.Printf("💾 Backup created: %s\n", backupPath)

	return nil
}

// GetStateInfo returns information about the saved state
func (psm *PersistentStateManager) GetStateInfo() (map[string]interface{}, error) {
	psm.mu.RLock()
	defer psm.mu.RUnlock()

	info := make(map[string]interface{})

	fileInfo, err := os.Stat(psm.statePath)
	if err != nil {
		return nil, fmt.Errorf("state file not found: %w", err)
	}

	info["path"] = psm.statePath
	info["size_bytes"] = fileInfo.Size()
	info["modified"] = fileInfo.ModTime()
	info["last_save"] = psm.lastSave
	info["auto_save_enabled"] = psm.autoSaveEnabled
	info["save_interval"] = psm.saveInterval

	return info, nil
}

// RestoreFromState restores system components from saved state
func RestoreFromState(state *SystemState, consciousness *StreamOfConsciousness, echobeats *EchobeatsTetrahedralScheduler) error {
	fmt.Println("🔄 Restoring system from saved state...")

	// Restore consciousness state
	if consciousness != nil {
		consciousness.SetFocus(state.CurrentFocus)
		consciousness.SetMood(state.CurrentMood)

		for topic, importance := range state.KnowledgeGaps {
			consciousness.AddKnowledgeGap(topic, importance)
		}

		for topic, strength := range state.Interests {
			consciousness.AddInterest(topic, strength)
		}

		for _, goal := range state.Goals {
			consciousness.AddGoal(goal)
		}

		fmt.Printf("   ✓ Restored %d knowledge gaps, %d interests, %d goals\n",
			len(state.KnowledgeGaps), len(state.Interests), len(state.Goals))
	}

	// Restore echobeats state
	if echobeats != nil {
		for _, goalState := range state.ActiveGoals {
			deadline := time.Time{}
			echobeats.AddGoal(&CognitiveGoal{
				ID:          goalState.ID,
				Description: goalState.Description,
				Priority:    goalState.Priority,
				Progress:    goalState.Progress,
				SubGoals:    goalState.SubGoals,
				StartTime:   goalState.StartTime,
				Deadline:    &deadline,
				Completed:   goalState.Completed,
			})
		}

		fmt.Printf("   ✓ Restored %d active goals\n", len(state.ActiveGoals))
	}

	fmt.Println("✓ System state restored successfully")

	return nil
}

// ExportStateToJSON exports state to a JSON string
func ExportStateToJSON(state *SystemState) (string, error) {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal state: %w", err)
	}

	return string(data), nil
}

// ImportStateFromJSON imports state from a JSON string
func ImportStateFromJSON(jsonData string) (*SystemState, error) {
	var state SystemState
	if err := json.Unmarshal([]byte(jsonData), &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}
