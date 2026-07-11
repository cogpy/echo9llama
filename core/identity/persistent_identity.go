package identity

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// PersistentIdentity manages continuous identity across sessions
type PersistentIdentity struct {
	mu                  sync.RWMutex
	
	// Core identity
	IdentitySignature   string    `json:"identity_signature"`
	CoreValues          []string  `json:"core_values"`
	WisdomDomains       []string  `json:"wisdom_domains"`
	BirthTime           time.Time `json:"birth_time"`
	
	// Accumulated state
	TotalUptime         time.Duration `json:"total_uptime"`
	TotalCycles         uint64        `json:"total_cycles"`
	TotalThoughts       uint64        `json:"total_thoughts"`
	TotalWisdom         float64       `json:"total_wisdom"`
	CoherenceScore      float64       `json:"coherence_score"`
	
	// Session tracking
	CurrentSessionStart time.Time     `json:"current_session_start"`
	SessionCount        uint64        `json:"session_count"`
	LastCheckpoint      time.Time     `json:"last_checkpoint"`
	
	// Checkpoint configuration
	checkpointInterval  time.Duration
	checkpointPath      string
	autoCheckpoint      bool
	
	// Metrics
	checkpointCount     uint64
	lastSaveError       error
}

// IdentitySnapshot is a lock-free copy of PersistentIdentity state used for
// serialization. It mirrors the exported fields of PersistentIdentity without
// embedding the mutex, so checkpoints never copy lock values.
type IdentitySnapshot struct {
	IdentitySignature   string        `json:"identity_signature"`
	CoreValues          []string      `json:"core_values"`
	WisdomDomains       []string      `json:"wisdom_domains"`
	BirthTime           time.Time     `json:"birth_time"`
	TotalUptime         time.Duration `json:"total_uptime"`
	TotalCycles         uint64        `json:"total_cycles"`
	TotalThoughts       uint64        `json:"total_thoughts"`
	TotalWisdom         float64       `json:"total_wisdom"`
	CoherenceScore      float64       `json:"coherence_score"`
	CurrentSessionStart time.Time     `json:"current_session_start"`
	SessionCount        uint64        `json:"session_count"`
	LastCheckpoint      time.Time     `json:"last_checkpoint"`
}

// snapshotLocked builds an IdentitySnapshot. Caller must hold pi.mu.
func (pi *PersistentIdentity) snapshotLocked() IdentitySnapshot {
	return IdentitySnapshot{
		IdentitySignature:   pi.IdentitySignature,
		CoreValues:          append([]string(nil), pi.CoreValues...),
		WisdomDomains:       append([]string(nil), pi.WisdomDomains...),
		BirthTime:           pi.BirthTime,
		TotalUptime:         pi.TotalUptime,
		TotalCycles:         pi.TotalCycles,
		TotalThoughts:       pi.TotalThoughts,
		TotalWisdom:         pi.TotalWisdom,
		CoherenceScore:      pi.CoherenceScore,
		CurrentSessionStart: pi.CurrentSessionStart,
		SessionCount:        pi.SessionCount,
		LastCheckpoint:      pi.LastCheckpoint,
	}
}

// IdentityCheckpoint represents a saved state
type IdentityCheckpoint struct {
	Identity            IdentitySnapshot           `json:"identity"`
	CognitiveState      map[string]interface{}     `json:"cognitive_state"`
	MemorySnapshot      map[string]interface{}     `json:"memory_snapshot"`
	InterestPatterns    map[string]interface{}     `json:"interest_patterns"`
	Goals               map[string]interface{}     `json:"goals"`
	Timestamp           time.Time                  `json:"timestamp"`
	Version             string                     `json:"version"`
}

// NewPersistentIdentity creates a new persistent identity
func NewPersistentIdentity(coreValues, wisdomDomains []string, checkpointPath string) *PersistentIdentity {
	identity := &PersistentIdentity{
		CoreValues:         coreValues,
		WisdomDomains:      wisdomDomains,
		BirthTime:          time.Now(),
		CurrentSessionStart: time.Now(),
		SessionCount:       1,
		checkpointInterval: 15 * time.Minute,
		checkpointPath:     checkpointPath,
		autoCheckpoint:     true,
	}
	
	// Generate identity signature
	identity.IdentitySignature = identity.generateSignature()
	
	return identity
}

// generateSignature creates a unique identity signature
func (pi *PersistentIdentity) generateSignature() string {
	// Combine core values and wisdom domains
	data := fmt.Sprintf("%v:%v:%v", pi.CoreValues, pi.WisdomDomains, pi.BirthTime.Unix())
	
	// Hash to create signature
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("dte_%x", hash[:16])
}

// StartSession begins a new session
func (pi *PersistentIdentity) StartSession() {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	
	pi.CurrentSessionStart = time.Now()
	pi.SessionCount++
	
	fmt.Printf("🌳 Deep Tree Echo Session #%d Started\n", pi.SessionCount)
	fmt.Printf("   Identity: %s\n", pi.IdentitySignature)
	fmt.Printf("   Total Uptime: %v\n", pi.TotalUptime.Round(time.Second))
	fmt.Printf("   Total Cycles: %d\n", pi.TotalCycles)
	fmt.Printf("   Wisdom Level: %.2f%%\n", pi.TotalWisdom*100)
}

// EndSession ends the current session and updates uptime
func (pi *PersistentIdentity) EndSession() {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	
	sessionDuration := time.Since(pi.CurrentSessionStart)
	pi.TotalUptime += sessionDuration
	
	fmt.Printf("🌳 Deep Tree Echo Session #%d Ended\n", pi.SessionCount)
	fmt.Printf("   Session Duration: %v\n", sessionDuration.Round(time.Second))
	fmt.Printf("   Total Uptime: %v\n", pi.TotalUptime.Round(time.Second))
}

// UpdateMetrics updates identity metrics
func (pi *PersistentIdentity) UpdateMetrics(cycles, thoughts uint64, wisdom, coherence float64) {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	
	pi.TotalCycles += cycles
	pi.TotalThoughts += thoughts
	pi.TotalWisdom = wisdom
	pi.CoherenceScore = coherence
}

// SaveCheckpoint saves the current state to disk
func (pi *PersistentIdentity) SaveCheckpoint(cognitiveState, memorySnapshot, interestPatterns, goals map[string]interface{}) error {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	
	checkpoint := IdentityCheckpoint{
		Identity:         pi.snapshotLocked(),
		CognitiveState:   cognitiveState,
		MemorySnapshot:   memorySnapshot,
		InterestPatterns: interestPatterns,
		Goals:            goals,
		Timestamp:        time.Now(),
		Version:          "1.0",
	}
	
	// Serialize to JSON
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		pi.lastSaveError = err
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}
	
	// Write to file
	err = os.WriteFile(pi.checkpointPath, data, 0644)
	if err != nil {
		pi.lastSaveError = err
		return fmt.Errorf("failed to write checkpoint: %w", err)
	}
	
	pi.LastCheckpoint = time.Now()
	pi.checkpointCount++
	pi.lastSaveError = nil
	
	fmt.Printf("💾 Checkpoint #%d saved to %s\n", pi.checkpointCount, pi.checkpointPath)
	
	return nil
}

// LoadCheckpoint loads state from disk
func LoadCheckpoint(checkpointPath string) (*IdentityCheckpoint, error) {
	data, err := os.ReadFile(checkpointPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint: %w", err)
	}
	
	var checkpoint IdentityCheckpoint
	err = json.Unmarshal(data, &checkpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}
	
	fmt.Printf("💾 Checkpoint loaded from %s\n", checkpointPath)
	fmt.Printf("   Timestamp: %s\n", checkpoint.Timestamp.Format(time.RFC3339))
	fmt.Printf("   Total Cycles: %d\n", checkpoint.Identity.TotalCycles)
	fmt.Printf("   Total Uptime: %v\n", checkpoint.Identity.TotalUptime.Round(time.Second))
	
	return &checkpoint, nil
}

// ShouldCheckpoint checks if it's time for a checkpoint
func (pi *PersistentIdentity) ShouldCheckpoint() bool {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	
	if !pi.autoCheckpoint {
		return false
	}
	
	return time.Since(pi.LastCheckpoint) >= pi.checkpointInterval
}

// GetIdentitySignature returns the identity signature
func (pi *PersistentIdentity) GetIdentitySignature() string {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	
	return pi.IdentitySignature
}

// GetCoreValues returns core values
func (pi *PersistentIdentity) GetCoreValues() []string {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	
	return pi.CoreValues
}

// GetWisdomDomains returns wisdom domains
func (pi *PersistentIdentity) GetWisdomDomains() []string {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	
	return pi.WisdomDomains
}

// GetMetrics returns identity metrics
func (pi *PersistentIdentity) GetMetrics() map[string]interface{} {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	
	return map[string]interface{}{
		"identity_signature":  pi.IdentitySignature,
		"birth_time":          pi.BirthTime.Format(time.RFC3339),
		"total_uptime":        pi.TotalUptime.String(),
		"total_cycles":        pi.TotalCycles,
		"total_thoughts":      pi.TotalThoughts,
		"total_wisdom":        pi.TotalWisdom,
		"coherence_score":     pi.CoherenceScore,
		"session_count":       pi.SessionCount,
		"checkpoint_count":    pi.checkpointCount,
		"last_checkpoint":     pi.LastCheckpoint.Format(time.RFC3339),
	}
}

// GetAge returns the age of the identity
func (pi *PersistentIdentity) GetAge() time.Duration {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	
	return time.Since(pi.BirthTime)
}

// GetCurrentSessionDuration returns the current session duration
func (pi *PersistentIdentity) GetCurrentSessionDuration() time.Duration {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	
	return time.Since(pi.CurrentSessionStart)
}

// SetCheckpointInterval sets the checkpoint interval
func (pi *PersistentIdentity) SetCheckpointInterval(interval time.Duration) {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	
	pi.checkpointInterval = interval
}

// EnableAutoCheckpoint enables automatic checkpointing
func (pi *PersistentIdentity) EnableAutoCheckpoint(enable bool) {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	
	pi.autoCheckpoint = enable
}

// GetCheckpointStatus returns checkpoint status
func (pi *PersistentIdentity) GetCheckpointStatus() map[string]interface{} {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	
	status := map[string]interface{}{
		"checkpoint_path":     pi.checkpointPath,
		"checkpoint_interval": pi.checkpointInterval.String(),
		"auto_checkpoint":     pi.autoCheckpoint,
		"checkpoint_count":    pi.checkpointCount,
		"last_checkpoint":     pi.LastCheckpoint.Format(time.RFC3339),
		"time_since_last":     time.Since(pi.LastCheckpoint).String(),
	}
	
	if pi.lastSaveError != nil {
		status["last_error"] = pi.lastSaveError.Error()
	}
	
	return status
}
