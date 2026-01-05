package lampstack

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// JSONPersistence provides JSON file-based state persistence
type JSONPersistence struct {
	mu sync.RWMutex

	config    *Config
	dataDir   string
	state     map[string]interface{}
	dirty     map[string]bool
	autoSave  bool
	closed    bool

	// Auto-save control
	ctx    context.Context
	cancel context.CancelFunc
}

// NewJSONPersistence creates a new JSON persistence layer
func NewJSONPersistence(config *Config) *JSONPersistence {
	ctx, cancel := context.WithCancel(context.Background())

	p := &JSONPersistence{
		config:   config,
		dataDir:  config.DataDirectory,
		state:    make(map[string]interface{}),
		dirty:    make(map[string]bool),
		autoSave: config.AutoSave,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Ensure data directory exists
	if err := os.MkdirAll(p.dataDir, 0755); err != nil {
		fmt.Printf("   ⚠️  Failed to create data directory: %v\n", err)
	}

	// Start auto-save goroutine if enabled
	if p.autoSave && config.SaveInterval > 0 {
		go p.autoSaveLoop(config.SaveInterval)
	}

	return p
}

// Save persists a single key-value pair
func (p *JSONPersistence) Save(ctx context.Context, key string, value interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("persistence layer is closed")
	}

	p.state[key] = value
	p.dirty[key] = true

	// Write to file immediately if auto-save is disabled
	if !p.autoSave {
		return p.writeFile(key, value)
	}

	return nil
}

// SaveAll persists all dirty state to disk
func (p *JSONPersistence) SaveAll(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("persistence layer is closed")
	}

	errors := make([]error, 0)

	for key := range p.dirty {
		if value, ok := p.state[key]; ok {
			if err := p.writeFile(key, value); err != nil {
				errors = append(errors, fmt.Errorf("failed to save %s: %w", key, err))
			} else {
				delete(p.dirty, key)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("save errors: %v", errors)
	}

	return nil
}

// Load retrieves a value from persistence
func (p *JSONPersistence) Load(ctx context.Context, key string, value interface{}) error {
	p.mu.RLock()
	// Check in-memory cache first
	if cached, ok := p.state[key]; ok {
		p.mu.RUnlock()
		// Copy cached value to destination
		data, err := json.Marshal(cached)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, value)
	}
	p.mu.RUnlock()

	// Load from file
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := p.readFile(key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, value); err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", key, err)
	}

	// Cache in memory
	var cached interface{}
	if err := json.Unmarshal(data, &cached); err == nil {
		p.state[key] = cached
	}

	return nil
}

// LoadAll loads all persisted state from disk
func (p *JSONPersistence) LoadAll(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("persistence layer is closed")
	}

	// List all JSON files in data directory
	entries, err := os.ReadDir(p.dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No data yet
		}
		return fmt.Errorf("failed to read data directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if len(name) < 5 || name[len(name)-5:] != ".json" {
			continue
		}

		key := name[:len(name)-5]
		data, err := p.readFile(key)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to load %s: %v\n", key, err)
			continue
		}

		var value interface{}
		if err := json.Unmarshal(data, &value); err != nil {
			fmt.Printf("   ⚠️  Failed to parse %s: %v\n", key, err)
			continue
		}

		p.state[key] = value
	}

	return nil
}

// Keys returns all persisted keys
func (p *JSONPersistence) Keys(ctx context.Context) ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Start with in-memory keys
	keySet := make(map[string]bool)
	for key := range p.state {
		keySet[key] = true
	}

	// Add keys from disk
	entries, err := os.ReadDir(p.dataDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if len(name) >= 5 && name[len(name)-5:] == ".json" {
				keySet[name[:len(name)-5]] = true
			}
		}
	}

	keys := make([]string, 0, len(keySet))
	for key := range keySet {
		keys = append(keys, key)
	}

	return keys, nil
}

// Exists checks if a key exists in persistence
func (p *JSONPersistence) Exists(ctx context.Context, key string) (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Check in-memory
	if _, ok := p.state[key]; ok {
		return true, nil
	}

	// Check on disk
	filePath := p.getFilePath(key)
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Delete removes a key from persistence
func (p *JSONPersistence) Delete(ctx context.Context, key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("persistence layer is closed")
	}

	// Remove from memory
	delete(p.state, key)
	delete(p.dirty, key)

	// Remove from disk
	filePath := p.getFilePath(key)
	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete %s: %w", key, err)
	}

	return nil
}

// Close shuts down the persistence layer
func (p *JSONPersistence) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.cancel()

	// Final save of dirty state
	for key := range p.dirty {
		if value, ok := p.state[key]; ok {
			if err := p.writeFile(key, value); err != nil {
				fmt.Printf("   ⚠️  Failed to save %s on close: %v\n", key, err)
			}
		}
	}

	p.closed = true
	return nil
}

// Helper methods

func (p *JSONPersistence) getFilePath(key string) string {
	// Sanitize key for filename
	safeKey := sanitizeFilename(key)
	return filepath.Join(p.dataDir, safeKey+".json")
}

func (p *JSONPersistence) writeFile(key string, value interface{}) error {
	filePath := p.getFilePath(key)

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", key, err)
	}

	// Write to temp file first, then rename (atomic)
	tempPath := filePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tempPath, filePath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func (p *JSONPersistence) readFile(key string) ([]byte, error) {
	filePath := p.getFilePath(key)
	return os.ReadFile(filePath)
}

func (p *JSONPersistence) autoSaveLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			if err := p.SaveAll(p.ctx); err != nil {
				fmt.Printf("   ⚠️  Auto-save failed: %v\n", err)
			}
		}
	}
}

// sanitizeFilename removes/replaces characters that are invalid in filenames
func sanitizeFilename(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			result = append(result, '_')
		default:
			result = append(result, c)
		}
	}
	return string(result)
}

// StateSnapshot represents a complete snapshot of the system state
type StateSnapshot struct {
	Version      string                 `json:"version"`
	Timestamp    time.Time              `json:"timestamp"`
	CognitiveState *CognitiveState       `json:"cognitive_state,omitempty"`
	Memories     []*MemoryEntry         `json:"memories,omitempty"`
	Metrics      *StackMetrics          `json:"metrics,omitempty"`
	Custom       map[string]interface{} `json:"custom,omitempty"`
}

// CreateSnapshot creates a complete state snapshot
func (p *JSONPersistence) CreateSnapshot(ctx context.Context, stack *Stack) (*StateSnapshot, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	snapshot := &StateSnapshot{
		Version:   Version,
		Timestamp: time.Now(),
		Custom:    make(map[string]interface{}),
	}

	// Get cognitive state
	if stack.echo != nil {
		state, err := stack.echo.GetState()
		if err == nil {
			snapshot.CognitiveState = state
		}
	}

	// Get metrics
	snapshot.Metrics = stack.GetMetrics()

	// Get recent memories
	if stack.memory != nil {
		memories, err := stack.memory.RetrieveRecent(ctx, 100)
		if err == nil {
			snapshot.Memories = memories
		}
	}

	return snapshot, nil
}

// SaveSnapshot saves a complete state snapshot
func (p *JSONPersistence) SaveSnapshot(ctx context.Context, snapshot *StateSnapshot) error {
	key := fmt.Sprintf("snapshot_%d", time.Now().Unix())
	return p.Save(ctx, key, snapshot)
}

// LoadLatestSnapshot loads the most recent snapshot
func (p *JSONPersistence) LoadLatestSnapshot(ctx context.Context) (*StateSnapshot, error) {
	keys, err := p.Keys(ctx)
	if err != nil {
		return nil, err
	}

	// Find latest snapshot
	var latestKey string
	var latestTime int64

	for _, key := range keys {
		if len(key) > 9 && key[:9] == "snapshot_" {
			var timestamp int64
			if _, err := fmt.Sscanf(key[9:], "%d", &timestamp); err == nil {
				if timestamp > latestTime {
					latestTime = timestamp
					latestKey = key
				}
			}
		}
	}

	if latestKey == "" {
		return nil, fmt.Errorf("no snapshots found")
	}

	var snapshot StateSnapshot
	if err := p.Load(ctx, latestKey, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

// ListSnapshots returns all available snapshots
func (p *JSONPersistence) ListSnapshots(ctx context.Context) ([]string, error) {
	keys, err := p.Keys(ctx)
	if err != nil {
		return nil, err
	}

	snapshots := make([]string, 0)
	for _, key := range keys {
		if len(key) > 9 && key[:9] == "snapshot_" {
			snapshots = append(snapshots, key)
		}
	}

	return snapshots, nil
}

// PruneSnapshots removes old snapshots, keeping only the most recent N
func (p *JSONPersistence) PruneSnapshots(ctx context.Context, keep int) error {
	snapshots, err := p.ListSnapshots(ctx)
	if err != nil {
		return err
	}

	if len(snapshots) <= keep {
		return nil
	}

	// Parse timestamps and sort
	type snapshotInfo struct {
		key       string
		timestamp int64
	}

	infos := make([]snapshotInfo, 0, len(snapshots))
	for _, key := range snapshots {
		var timestamp int64
		if _, err := fmt.Sscanf(key[9:], "%d", &timestamp); err == nil {
			infos = append(infos, snapshotInfo{key, timestamp})
		}
	}

	// Sort by timestamp descending
	for i := 0; i < len(infos)-1; i++ {
		for j := i + 1; j < len(infos); j++ {
			if infos[j].timestamp > infos[i].timestamp {
				infos[i], infos[j] = infos[j], infos[i]
			}
		}
	}

	// Delete older snapshots
	for i := keep; i < len(infos); i++ {
		if err := p.Delete(ctx, infos[i].key); err != nil {
			fmt.Printf("   ⚠️  Failed to prune snapshot %s: %v\n", infos[i].key, err)
		}
	}

	return nil
}
