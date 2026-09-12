package persistence

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore provides persistent storage for the autonomous system
type SQLiteStore struct {
	mu     sync.RWMutex
	db     *sql.DB
	dbPath string
	isOpen bool
}

// ThoughtRecord represents a persisted thought
type ThoughtRecord struct {
	ID         int64
	Content    string
	Type       string
	Timestamp  time.Time
	Context    string // JSON-encoded context
	Interests  string // JSON-encoded interests
	Importance float64
}

// MemoryRecord represents a persisted memory
type MemoryRecord struct {
	ID           int64
	Content      string
	Type         string
	Timestamp    time.Time
	Strength     float64
	Associations string // JSON-encoded associations
}

// StateRecord represents persisted system state
type StateRecord struct {
	ID        int64
	Key       string
	Value     string // JSON-encoded value
	UpdatedAt time.Time
}

// GoalRecord represents a persisted goal
type GoalRecord struct {
	ID          int64
	Description string
	Type        string
	Priority    float64
	Status      string
	CreatedAt   time.Time
	CompletedAt *time.Time
	Metadata    string // JSON-encoded metadata
}

// NewSQLiteStore creates a new SQLite-based persistence store
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	store := &SQLiteStore{
		dbPath: dbPath,
	}

	if err := store.Open(); err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// Open opens the database connection
func (s *SQLiteStore) Open() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isOpen {
		return nil
	}

	db, err := sql.Open("sqlite3", s.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite works best with single connection
	db.SetMaxIdleConns(1)

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	s.db = db
	s.isOpen = true

	return nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isOpen || s.db == nil {
		return nil
	}

	err := s.db.Close()
	s.isOpen = false
	return err
}

// initSchema creates the database schema if it doesn't exist
func (s *SQLiteStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS thoughts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		type TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		context TEXT,
		interests TEXT,
		importance REAL DEFAULT 0.5,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS memories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		type TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		strength REAL DEFAULT 1.0,
		associations TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS state (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT UNIQUE NOT NULL,
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS goals (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		description TEXT NOT NULL,
		type TEXT NOT NULL,
		priority REAL DEFAULT 0.5,
		status TEXT DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME,
		metadata TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_thoughts_timestamp ON thoughts(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_thoughts_type ON thoughts(type);
	CREATE INDEX IF NOT EXISTS idx_memories_timestamp ON memories(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_memories_strength ON memories(strength DESC);
	CREATE INDEX IF NOT EXISTS idx_goals_status ON goals(status);
	CREATE INDEX IF NOT EXISTS idx_goals_priority ON goals(priority DESC);
	`

	_, err := s.db.Exec(schema)
	return err
}

// SaveThought persists a thought to the database
func (s *SQLiteStore) SaveThought(thought *ThoughtRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isOpen {
		return fmt.Errorf("database not open")
	}

	query := `
		INSERT INTO thoughts (content, type, timestamp, context, interests, importance)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(query,
		thought.Content,
		thought.Type,
		thought.Timestamp,
		thought.Context,
		thought.Interests,
		thought.Importance,
	)

	if err != nil {
		return fmt.Errorf("failed to save thought: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		thought.ID = id
	}

	return nil
}

// GetRecentThoughts retrieves the most recent thoughts
func (s *SQLiteStore) GetRecentThoughts(limit int) ([]*ThoughtRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isOpen {
		return nil, fmt.Errorf("database not open")
	}

	query := `
		SELECT id, content, type, timestamp, context, interests, importance
		FROM thoughts
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query thoughts: %w", err)
	}
	defer rows.Close()

	thoughts := make([]*ThoughtRecord, 0, limit)
	for rows.Next() {
		thought := &ThoughtRecord{}
		err := rows.Scan(
			&thought.ID,
			&thought.Content,
			&thought.Type,
			&thought.Timestamp,
			&thought.Context,
			&thought.Interests,
			&thought.Importance,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan thought: %w", err)
		}
		thoughts = append(thoughts, thought)
	}

	return thoughts, nil
}

// SaveMemory persists a memory to the database
func (s *SQLiteStore) SaveMemory(memory *MemoryRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isOpen {
		return fmt.Errorf("database not open")
	}

	query := `
		INSERT INTO memories (content, type, timestamp, strength, associations)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(query,
		memory.Content,
		memory.Type,
		memory.Timestamp,
		memory.Strength,
		memory.Associations,
	)

	if err != nil {
		return fmt.Errorf("failed to save memory: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		memory.ID = id
	}

	return nil
}

// GetStrongMemories retrieves memories above a strength threshold
func (s *SQLiteStore) GetStrongMemories(minStrength float64, limit int) ([]*MemoryRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isOpen {
		return nil, fmt.Errorf("database not open")
	}

	query := `
		SELECT id, content, type, timestamp, strength, associations
		FROM memories
		WHERE strength >= ?
		ORDER BY strength DESC, timestamp DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, minStrength, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}
	defer rows.Close()

	memories := make([]*MemoryRecord, 0, limit)
	for rows.Next() {
		memory := &MemoryRecord{}
		err := rows.Scan(
			&memory.ID,
			&memory.Content,
			&memory.Type,
			&memory.Timestamp,
			&memory.Strength,
			&memory.Associations,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory: %w", err)
		}
		memories = append(memories, memory)
	}

	return memories, nil
}

// SaveState persists a key-value state
func (s *SQLiteStore) SaveState(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isOpen {
		return fmt.Errorf("database not open")
	}

	// Serialize value to JSON
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	query := `
		INSERT OR REPLACE INTO state (key, value, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`

	_, err = s.db.Exec(query, key, string(valueJSON))
	if err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// GetState retrieves a state value by key
func (s *SQLiteStore) GetState(key string, target interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isOpen {
		return fmt.Errorf("database not open")
	}

	query := `SELECT value FROM state WHERE key = ?`

	var valueJSON string
	err := s.db.QueryRow(query, key).Scan(&valueJSON)
	if err == sql.ErrNoRows {
		return fmt.Errorf("state key not found: %s", key)
	}
	if err != nil {
		return fmt.Errorf("failed to query state: %w", err)
	}

	// Deserialize JSON to target
	err = json.Unmarshal([]byte(valueJSON), target)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// SaveGoal persists a goal to the database
func (s *SQLiteStore) SaveGoal(goal *GoalRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isOpen {
		return fmt.Errorf("database not open")
	}

	query := `
		INSERT INTO goals (description, type, priority, status, metadata)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(query,
		goal.Description,
		goal.Type,
		goal.Priority,
		goal.Status,
		goal.Metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to save goal: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		goal.ID = id
	}

	return nil
}

// GetActiveGoals retrieves all active goals
func (s *SQLiteStore) GetActiveGoals() ([]*GoalRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isOpen {
		return nil, fmt.Errorf("database not open")
	}

	query := `
		SELECT id, description, type, priority, status, created_at, completed_at, metadata
		FROM goals
		WHERE status = 'active'
		ORDER BY priority DESC, created_at ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query goals: %w", err)
	}
	defer rows.Close()

	goals := make([]*GoalRecord, 0)
	for rows.Next() {
		goal := &GoalRecord{}
		err := rows.Scan(
			&goal.ID,
			&goal.Description,
			&goal.Type,
			&goal.Priority,
			&goal.Status,
			&goal.CreatedAt,
			&goal.CompletedAt,
			&goal.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan goal: %w", err)
		}
		goals = append(goals, goal)
	}

	return goals, nil
}

// UpdateGoalStatus updates the status of a goal
func (s *SQLiteStore) UpdateGoalStatus(goalID int64, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isOpen {
		return fmt.Errorf("database not open")
	}

	query := `UPDATE goals SET status = ?, completed_at = ? WHERE id = ?`

	var completedAt *time.Time
	if status == "completed" {
		now := time.Now()
		completedAt = &now
	}

	_, err := s.db.Exec(query, status, completedAt, goalID)
	if err != nil {
		return fmt.Errorf("failed to update goal status: %w", err)
	}

	return nil
}

// GetStats returns database statistics
func (s *SQLiteStore) GetStats() (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isOpen {
		return nil, fmt.Errorf("database not open")
	}

	stats := make(map[string]interface{})

	// Count thoughts
	var thoughtCount int
	err := s.db.QueryRow("SELECT COUNT(*) FROM thoughts").Scan(&thoughtCount)
	if err == nil {
		stats["thought_count"] = thoughtCount
	}

	// Count memories
	var memoryCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM memories").Scan(&memoryCount)
	if err == nil {
		stats["memory_count"] = memoryCount
	}

	// Count goals
	var goalCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM goals WHERE status = 'active'").Scan(&goalCount)
	if err == nil {
		stats["active_goal_count"] = goalCount
	}

	// Database size
	var pageCount, pageSize int64
	s.db.QueryRow("PRAGMA page_count").Scan(&pageCount)
	s.db.QueryRow("PRAGMA page_size").Scan(&pageSize)
	stats["db_size_bytes"] = pageCount * pageSize

	stats["db_path"] = s.dbPath

	return stats, nil
}
