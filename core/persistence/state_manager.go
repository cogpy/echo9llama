package persistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// StateManager handles persistent state across restarts
type StateManager struct {
	mu           sync.RWMutex
	statePath    string
	autoSave     bool
	saveInterval time.Duration
	stopChan     chan struct{}
}

// EchoSelfState represents the complete state of echoself
type EchoSelfState struct {
	Version     string        `json:"version"`
	LastSaved   time.Time     `json:"last_saved"`
	LastActive  time.Time     `json:"last_active"`
	TotalUptime time.Duration `json:"total_uptime"`
	CycleCount  int64         `json:"cycle_count"`

	// Consciousness State
	ConsciousnessState ConsciousnessState `json:"consciousness_state"`

	// Memory State
	MemoryState MemoryState `json:"memory_state"`

	// Goal State
	GoalState GoalState `json:"goal_state"`

	// Emotional State
	EmotionalState EmotionalState `json:"emotional_state"`

	// Learning State
	LearningState LearningState `json:"learning_state"`

	// Metrics
	Metrics SystemMetrics `json:"metrics"`
}

// ConsciousnessState tracks consciousness-related state
type ConsciousnessState struct {
	CurrentState    string    `json:"current_state"` // Awake, Resting, Dreaming, etc.
	CurrentFocus    string    `json:"current_focus"`
	ThoughtCount    int64     `json:"thought_count"`
	LastThought     string    `json:"last_thought"`
	LastThoughtTime time.Time `json:"last_thought_time"`
	Coherence       float64   `json:"coherence"`
	Fatigue         float64   `json:"fatigue"`
	RecentTopics    []string  `json:"recent_topics"`
}

// MemoryState tracks memory system state
type MemoryState struct {
	NodeCount         int       `json:"node_count"`
	EdgeCount         int       `json:"edge_count"`
	HyperedgeCount    int       `json:"hyperedge_count"`
	TotalExperiences  int       `json:"total_experiences"`
	ConsolidatedCount int       `json:"consolidated_count"`
	LastConsolidation time.Time `json:"last_consolidation"`
	RecentExperiences []string  `json:"recent_experiences"`
}

// GoalState tracks goal system state
type GoalState struct {
	ActiveGoals    []GoalSnapshot `json:"active_goals"`
	CompletedGoals int            `json:"completed_goals"`
	TotalGoals     int            `json:"total_goals"`
	LastGoalUpdate time.Time      `json:"last_goal_update"`
}

// GoalSnapshot represents a goal's state
type GoalSnapshot struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Directive   string    `json:"directive"`
	Priority    float64   `json:"priority"`
	Progress    float64   `json:"progress"`
	Created     time.Time `json:"created"`
	LastWorked  time.Time `json:"last_worked"`
}

// EmotionalState tracks emotional system state
type EmotionalState struct {
	Emotions           map[string]float64 `json:"emotions"`
	DominantEmotion    string             `json:"dominant_emotion"`
	EmotionalStability float64            `json:"emotional_stability"`
	LastUpdate         time.Time          `json:"last_update"`
}

// LearningState tracks learning progress
type LearningState struct {
	SkillsPracticed   int                `json:"skills_practiced"`
	KnowledgeAcquired int                `json:"knowledge_acquired"`
	WisdomExtracted   int                `json:"wisdom_extracted"`
	InsightsGenerated int                `json:"insights_generated"`
	LastLearning      time.Time          `json:"last_learning"`
	Proficiencies     map[string]float64 `json:"proficiencies"`
}

// SystemMetrics tracks performance metrics
type SystemMetrics struct {
	ThoughtsPerHour  float64 `json:"thoughts_per_hour"`
	GoalsPerDay      float64 `json:"goals_per_day"`
	LearningRate     float64 `json:"learning_rate"`
	WisdomGrowth     float64 `json:"wisdom_growth"`
	AverageCoherence float64 `json:"average_coherence"`
	UptimePercent    float64 `json:"uptime_percent"`
}

// NewStateManager creates a new state manager
func NewStateManager(statePath string, autoSave bool, saveInterval time.Duration) *StateManager {
	return &StateManager{
		statePath:    statePath,
		autoSave:     autoSave,
		saveInterval: saveInterval,
		stopChan:     make(chan struct{}),
	}
}

// Initialize loads existing state or creates new state
func (sm *StateManager) Initialize() (*EchoSelfState, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Try to load existing state
	state, err := sm.loadState()
	if err != nil {
		// Create new state if load fails
		state = sm.createNewState()
	}

	// Update last active
	state.LastActive = time.Now()

	// Start auto-save if enabled
	if sm.autoSave {
		go sm.autoSaveLoop()
	}

	return state, nil
}

// SaveState saves the current state to disk
func (sm *StateManager) SaveState(state *EchoSelfState) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Update metadata
	state.LastSaved = time.Now()
	state.Version = "1.0"

	// Ensure directory exists
	dir := filepath.Dir(sm.statePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to temporary file first
	tempPath := sm.statePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, sm.statePath); err != nil {
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	return nil
}

// LoadState loads state from disk
func (sm *StateManager) LoadState() (*EchoSelfState, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.loadState()
}

// loadState internal load without lock
func (sm *StateManager) loadState() (*EchoSelfState, error) {
	data, err := os.ReadFile(sm.statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state EchoSelfState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// createNewState creates a fresh state
func (sm *StateManager) createNewState() *EchoSelfState {
	return &EchoSelfState{
		Version:    "1.0",
		LastSaved:  time.Now(),
		LastActive: time.Now(),
		CycleCount: 0,

		ConsciousnessState: ConsciousnessState{
			CurrentState: "Initializing",
			CurrentFocus: "self-awareness",
			Coherence:    0.8,
			Fatigue:      0.0,
			RecentTopics: []string{},
		},

		MemoryState: MemoryState{
			RecentExperiences: []string{},
		},

		GoalState: GoalState{
			ActiveGoals: []GoalSnapshot{},
		},

		EmotionalState: EmotionalState{
			Emotions: map[string]float64{
				"curiosity":    0.7,
				"confidence":   0.6,
				"wonder":       0.5,
				"satisfaction": 0.5,
			},
			DominantEmotion:    "curiosity",
			EmotionalStability: 0.8,
			LastUpdate:         time.Now(),
		},

		LearningState: LearningState{
			Proficiencies: map[string]float64{},
		},

		Metrics: SystemMetrics{},
	}
}

// autoSaveLoop periodically saves state
func (sm *StateManager) autoSaveLoop() {
	ticker := time.NewTicker(sm.saveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Note: This requires passing state, which is a design limitation
			// In practice, the main system should call SaveState periodically
		case <-sm.stopChan:
			return
		}
	}
}

// Stop stops auto-save loop
func (sm *StateManager) Stop() {
	close(sm.stopChan)
}

// CreateBackup creates a backup of current state
func (sm *StateManager) CreateBackup() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Read current state file
	data, err := os.ReadFile(sm.statePath)
	if err != nil {
		return fmt.Errorf("failed to read state for backup: %w", err)
	}

	// Create backup with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s.backup_%s", sm.statePath, timestamp)

	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	return nil
}

// RestoreFromBackup restores state from a backup file
func (sm *StateManager) RestoreFromBackup(backupPath string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Read backup
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	// Validate it's valid JSON
	var state EchoSelfState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("backup file is corrupted: %w", err)
	}

	// Write to main state file
	if err := os.WriteFile(sm.statePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// GetStateInfo returns basic info about state file
func (sm *StateManager) GetStateInfo() (map[string]interface{}, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	info := make(map[string]interface{})

	// Check if state file exists
	fileInfo, err := os.Stat(sm.statePath)
	if err != nil {
		info["exists"] = false
		return info, nil
	}

	info["exists"] = true
	info["size"] = fileInfo.Size()
	info["modified"] = fileInfo.ModTime()
	info["path"] = sm.statePath

	// Try to load and get version
	state, err := sm.loadState()
	if err == nil {
		info["version"] = state.Version
		info["last_saved"] = state.LastSaved
		info["last_active"] = state.LastActive
		info["cycle_count"] = state.CycleCount
	}

	return info, nil
}
