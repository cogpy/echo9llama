package deeptreeecho

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PersistentStateStore provides a fast key-value store for Deep Tree Echo
// This is inspired by Badger's architecture for high-performance persistent storage
type PersistentStateStore struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc

	// In-memory storage (memtable)
	memtable        map[string]*StateEntry
	memtableLock    sync.RWMutex

	// Immutable memtables (for flush)
	immutableTables []*ImmutableTable

	// Namespaces for organizing data
	namespaces      map[string]bool

	// Transaction support
	transactions    map[string]*Transaction
	txnLock         sync.RWMutex

	// Configuration
	memtableSize    int64
	currentSize     int64
	persistPath     string

	// Metrics
	totalGets       uint64
	totalSets       uint64
	totalDeletes    uint64

	// Running state
	running         bool
}

// StateEntry represents a key-value entry in the store
type StateEntry struct {
	Key       string
	Value     []byte
	Version   uint64
	ExpiresAt time.Time
	Metadata  map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
	Deleted   bool
}

// ImmutableTable represents a frozen memtable ready for persistence
type ImmutableTable struct {
	ID        string
	Entries   map[string]*StateEntry
	CreatedAt time.Time
	Size      int64
}

// Transaction represents an atomic transaction
type Transaction struct {
	ID        string
	Entries   map[string]*StateEntry
	ReadSet   map[string]uint64 // key -> version at read time
	WriteSet  map[string]*StateEntry
	StartTime time.Time
	Committed bool
	Discarded bool
}

// NewPersistentStateStore creates a new persistent state store
func NewPersistentStateStore() *PersistentStateStore {
	ctx, cancel := context.WithCancel(context.Background())

	return &PersistentStateStore{
		ctx:             ctx,
		cancel:          cancel,
		memtable:        make(map[string]*StateEntry),
		immutableTables: make([]*ImmutableTable, 0),
		namespaces:      make(map[string]bool),
		transactions:    make(map[string]*Transaction),
		memtableSize:    64 * 1024 * 1024, // 64MB default
		persistPath:     "./state_store",
	}
}

// Start begins the persistent state store
func (pss *PersistentStateStore) Start() error {
	pss.mu.Lock()
	defer pss.mu.Unlock()

	if pss.running {
		return fmt.Errorf("persistent state store already running")
	}

	pss.running = true
	fmt.Println("💾 Persistent State Store started")

	// Initialize default namespaces
	pss.initializeDefaultNamespaces()

	// Start background processes
	go pss.runCompactionLoop()
	go pss.runExpirationLoop()

	return nil
}

// Stop gracefully stops the persistent state store
func (pss *PersistentStateStore) Stop() error {
	pss.mu.Lock()
	defer pss.mu.Unlock()

	if !pss.running {
		return fmt.Errorf("persistent state store not running")
	}

	pss.cancel()
	pss.running = false
	fmt.Println("💾 Persistent State Store stopped")

	return nil
}

// initializeDefaultNamespaces creates default namespaces
func (pss *PersistentStateStore) initializeDefaultNamespaces() {
	defaultNamespaces := []string{
		"consciousness",  // Consciousness state
		"cognitive",      // Cognitive loop state
		"memory",         // Memory state
		"goals",          // Goal state
		"skills",         // Skill state
		"discussions",    // Discussion state
		"telemetry",      // Telemetry data
		"cache",          // Cache data
	}

	for _, ns := range defaultNamespaces {
		pss.namespaces[ns] = true
	}

	fmt.Printf("   Initialized %d default namespaces\n", len(defaultNamespaces))
}

// runCompactionLoop runs the background compaction process
func (pss *PersistentStateStore) runCompactionLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pss.ctx.Done():
			return
		case <-ticker.C:
			pss.maybeFlushMemtable()
		}
	}
}

// runExpirationLoop runs the background expiration process
func (pss *PersistentStateStore) runExpirationLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pss.ctx.Done():
			return
		case <-ticker.C:
			pss.expireEntries()
		}
	}
}

// maybeFlushMemtable flushes the memtable if it exceeds the size limit
func (pss *PersistentStateStore) maybeFlushMemtable() {
	pss.mu.RLock()
	shouldFlush := pss.currentSize >= pss.memtableSize
	pss.mu.RUnlock()

	if shouldFlush {
		pss.flushMemtable()
	}
}

// flushMemtable freezes the current memtable and creates a new one
func (pss *PersistentStateStore) flushMemtable() {
	pss.memtableLock.Lock()
	defer pss.memtableLock.Unlock()

	if len(pss.memtable) == 0 {
		return
	}

	// Create immutable table
	immTable := &ImmutableTable{
		ID:        fmt.Sprintf("imm_%s", uuid.New().String()[:8]),
		Entries:   pss.memtable,
		CreatedAt: time.Now(),
		Size:      pss.currentSize,
	}

	pss.mu.Lock()
	pss.immutableTables = append(pss.immutableTables, immTable)
	pss.mu.Unlock()

	// Create new memtable
	pss.memtable = make(map[string]*StateEntry)
	pss.mu.Lock()
	pss.currentSize = 0
	pss.mu.Unlock()
}

// expireEntries removes expired entries
func (pss *PersistentStateStore) expireEntries() {
	now := time.Now()

	pss.memtableLock.Lock()
	for key, entry := range pss.memtable {
		if !entry.ExpiresAt.IsZero() && entry.ExpiresAt.Before(now) {
			delete(pss.memtable, key)
		}
	}
	pss.memtableLock.Unlock()
}

// Set stores a key-value pair
func (pss *PersistentStateStore) Set(key string, value interface{}) error {
	return pss.SetWithTTL(key, value, 0)
}

// SetWithTTL stores a key-value pair with a time-to-live
func (pss *PersistentStateStore) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	pss.memtableLock.Lock()
	defer pss.memtableLock.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	entry := &StateEntry{
		Key:       key,
		Value:     valueBytes,
		Version:   1,
		ExpiresAt: expiresAt,
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Deleted:   false,
	}

	// Update version if key exists
	if existing, exists := pss.memtable[key]; exists {
		entry.Version = existing.Version + 1
		entry.CreatedAt = existing.CreatedAt
	}

	pss.memtable[key] = entry

	pss.mu.Lock()
	pss.currentSize += int64(len(valueBytes))
	pss.totalSets++
	pss.mu.Unlock()

	return nil
}

// Get retrieves a value by key
func (pss *PersistentStateStore) Get(key string, dest interface{}) error {
	pss.mu.Lock()
	pss.totalGets++
	pss.mu.Unlock()

	// Check memtable first
	pss.memtableLock.RLock()
	if entry, exists := pss.memtable[key]; exists && !entry.Deleted {
		pss.memtableLock.RUnlock()
		if !entry.ExpiresAt.IsZero() && entry.ExpiresAt.Before(time.Now()) {
			return fmt.Errorf("key expired: %s", key)
		}
		return json.Unmarshal(entry.Value, dest)
	}
	pss.memtableLock.RUnlock()

	// Check immutable tables
	pss.mu.RLock()
	for i := len(pss.immutableTables) - 1; i >= 0; i-- {
		if entry, exists := pss.immutableTables[i].Entries[key]; exists && !entry.Deleted {
			pss.mu.RUnlock()
			if !entry.ExpiresAt.IsZero() && entry.ExpiresAt.Before(time.Now()) {
				return fmt.Errorf("key expired: %s", key)
			}
			return json.Unmarshal(entry.Value, dest)
		}
	}
	pss.mu.RUnlock()

	return fmt.Errorf("key not found: %s", key)
}

// Delete removes a key
func (pss *PersistentStateStore) Delete(key string) error {
	pss.memtableLock.Lock()
	defer pss.memtableLock.Unlock()

	if entry, exists := pss.memtable[key]; exists {
		entry.Deleted = true
		entry.UpdatedAt = time.Now()
	} else {
		// Create tombstone entry
		pss.memtable[key] = &StateEntry{
			Key:       key,
			Deleted:   true,
			UpdatedAt: time.Now(),
		}
	}

	pss.mu.Lock()
	pss.totalDeletes++
	pss.mu.Unlock()

	return nil
}

// Exists checks if a key exists
func (pss *PersistentStateStore) Exists(key string) bool {
	pss.memtableLock.RLock()
	if entry, exists := pss.memtable[key]; exists && !entry.Deleted {
		pss.memtableLock.RUnlock()
		if !entry.ExpiresAt.IsZero() && entry.ExpiresAt.Before(time.Now()) {
			return false
		}
		return true
	}
	pss.memtableLock.RUnlock()

	pss.mu.RLock()
	for i := len(pss.immutableTables) - 1; i >= 0; i-- {
		if entry, exists := pss.immutableTables[i].Entries[key]; exists && !entry.Deleted {
			pss.mu.RUnlock()
			if !entry.ExpiresAt.IsZero() && entry.ExpiresAt.Before(time.Now()) {
				return false
			}
			return true
		}
	}
	pss.mu.RUnlock()

	return false
}

// BeginTransaction starts a new transaction
func (pss *PersistentStateStore) BeginTransaction() *Transaction {
	txn := &Transaction{
		ID:        fmt.Sprintf("txn_%s", uuid.New().String()[:8]),
		Entries:   make(map[string]*StateEntry),
		ReadSet:   make(map[string]uint64),
		WriteSet:  make(map[string]*StateEntry),
		StartTime: time.Now(),
	}

	pss.txnLock.Lock()
	pss.transactions[txn.ID] = txn
	pss.txnLock.Unlock()

	return txn
}

// TxnGet gets a value within a transaction
func (pss *PersistentStateStore) TxnGet(txn *Transaction, key string, dest interface{}) error {
	// Check write set first
	if entry, exists := txn.WriteSet[key]; exists {
		if entry.Deleted {
			return fmt.Errorf("key deleted in transaction: %s", key)
		}
		txn.ReadSet[key] = entry.Version
		return json.Unmarshal(entry.Value, dest)
	}

	// Get from store
	var entry StateEntry
	err := pss.Get(key, &entry)
	if err != nil {
		return err
	}

	txn.ReadSet[key] = entry.Version
	return json.Unmarshal(entry.Value, dest)
}

// TxnSet sets a value within a transaction
func (pss *PersistentStateStore) TxnSet(txn *Transaction, key string, value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	entry := &StateEntry{
		Key:       key,
		Value:     valueBytes,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	txn.WriteSet[key] = entry
	return nil
}

// TxnDelete deletes a key within a transaction
func (pss *PersistentStateStore) TxnDelete(txn *Transaction, key string) {
	txn.WriteSet[key] = &StateEntry{
		Key:     key,
		Deleted: true,
	}
}

// CommitTransaction commits a transaction
func (pss *PersistentStateStore) CommitTransaction(txn *Transaction) error {
	if txn.Committed || txn.Discarded {
		return fmt.Errorf("transaction already finalized")
	}

	// Validate read set (optimistic concurrency control)
	pss.memtableLock.RLock()
	for key, readVersion := range txn.ReadSet {
		if entry, exists := pss.memtable[key]; exists {
			if entry.Version != readVersion {
				pss.memtableLock.RUnlock()
				txn.Discarded = true
				return fmt.Errorf("transaction conflict on key: %s", key)
			}
		}
	}
	pss.memtableLock.RUnlock()

	// Apply write set
	pss.memtableLock.Lock()
	for key, entry := range txn.WriteSet {
		if existing, exists := pss.memtable[key]; exists {
			entry.Version = existing.Version + 1
			entry.CreatedAt = existing.CreatedAt
		}
		pss.memtable[key] = entry
	}
	pss.memtableLock.Unlock()

	txn.Committed = true

	// Remove transaction
	pss.txnLock.Lock()
	delete(pss.transactions, txn.ID)
	pss.txnLock.Unlock()

	return nil
}

// DiscardTransaction discards a transaction
func (pss *PersistentStateStore) DiscardTransaction(txn *Transaction) {
	txn.Discarded = true

	pss.txnLock.Lock()
	delete(pss.transactions, txn.ID)
	pss.txnLock.Unlock()
}

// GetMetrics returns state store metrics
func (pss *PersistentStateStore) GetMetrics() map[string]interface{} {
	pss.mu.RLock()
	defer pss.mu.RUnlock()

	pss.memtableLock.RLock()
	memtableEntries := len(pss.memtable)
	pss.memtableLock.RUnlock()

	return map[string]interface{}{
		"running":           pss.running,
		"memtable_entries":  memtableEntries,
		"memtable_size":     pss.currentSize,
		"immutable_tables":  len(pss.immutableTables),
		"total_gets":        pss.totalGets,
		"total_sets":        pss.totalSets,
		"total_deletes":     pss.totalDeletes,
	}
}

// ContributeToGestalt provides state store state for the global gestalt
func (pss *PersistentStateStore) ContributeToGestalt() map[string]interface{} {
	pss.mu.RLock()
	defer pss.mu.RUnlock()

	pss.memtableLock.RLock()
	memtableEntries := len(pss.memtable)
	pss.memtableLock.RUnlock()

	return map[string]interface{}{
		"running":          pss.running,
		"memtable_entries": memtableEntries,
		"total_operations": pss.totalGets + pss.totalSets + pss.totalDeletes,
	}
}

// SaveConsciousnessState saves the consciousness state
func (pss *PersistentStateStore) SaveConsciousnessState(state map[string]interface{}) error {
	return pss.Set("consciousness:current", state)
}

// LoadConsciousnessState loads the consciousness state
func (pss *PersistentStateStore) LoadConsciousnessState() (map[string]interface{}, error) {
	var state map[string]interface{}
	err := pss.Get("consciousness:current", &state)
	return state, err
}

// SaveCognitiveLoopState saves the cognitive loop state
func (pss *PersistentStateStore) SaveCognitiveLoopState(loopID string, state map[string]interface{}) error {
	return pss.Set(fmt.Sprintf("cognitive:loop:%s", loopID), state)
}

// LoadCognitiveLoopState loads the cognitive loop state
func (pss *PersistentStateStore) LoadCognitiveLoopState(loopID string) (map[string]interface{}, error) {
	var state map[string]interface{}
	err := pss.Get(fmt.Sprintf("cognitive:loop:%s", loopID), &state)
	return state, err
}

// CacheValue caches a value with TTL
func (pss *PersistentStateStore) CacheValue(key string, value interface{}, ttl time.Duration) error {
	return pss.SetWithTTL(fmt.Sprintf("cache:%s", key), value, ttl)
}

// GetCachedValue retrieves a cached value
func (pss *PersistentStateStore) GetCachedValue(key string, dest interface{}) error {
	return pss.Get(fmt.Sprintf("cache:%s", key), dest)
}
