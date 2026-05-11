// sys6_thread_multiplexing.go - Thread-Level Multiplexing for Sys6
// Implements the cycling permutations of particular sets and
// complementary triads with entangled qubit order 2 concurrency

package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// =============================================================================
// THREAD MULTIPLEXING CONSTANTS
// =============================================================================

const (
	// Number of particular sets (threads)
	NumParticularSets = 4

	// Number of dyadic permutations: C(4,2) = 6
	NumDyadicPermutations = 6

	// Number of triadic permutations: C(4,3) = 4
	NumTriadicPermutations = 4
)

// =============================================================================
// PARTICULAR SET PERMUTATIONS
// =============================================================================

// DyadicPermutation represents a pair of particular sets
type DyadicPermutation struct {
	P1 int // First particular set (1-4)
	P2 int // Second particular set (1-4)
}

// TriadicPermutation represents a triple of particular sets
type TriadicPermutation struct {
	P1 int // First particular set (1-4)
	P2 int // Second particular set (1-4)
	P3 int // Third particular set (1-4)
}

// GetDyadicPermutations returns all 6 dyadic permutations
// P(1,2)→P(1,3)→P(1,4)→P(2,3)→P(2,4)→P(3,4)
func GetDyadicPermutations() []DyadicPermutation {
	return []DyadicPermutation{
		{1, 2}, {1, 3}, {1, 4},
		{2, 3}, {2, 4}, {3, 4},
	}
}

// GetTriadicPermutations returns all 4 triadic permutations
// MP1: P[1,2,3]→P[1,2,4]→P[1,3,4]→P[2,3,4]
func GetTriadicPermutations() []TriadicPermutation {
	return []TriadicPermutation{
		{1, 2, 3}, {1, 2, 4}, {1, 3, 4}, {2, 3, 4},
	}
}

// GetComplementaryTriadicPermutations returns the complementary cycle
// MP2: P[1,3,4]→P[2,3,4]→P[1,2,3]→P[1,2,4]
func GetComplementaryTriadicPermutations() []TriadicPermutation {
	return []TriadicPermutation{
		{1, 3, 4}, {2, 3, 4}, {1, 2, 3}, {1, 2, 4},
	}
}

// =============================================================================
// ENTANGLED STATE MANAGER
// =============================================================================

// Sys6EntangledState represents a shared state with order 2 entanglement
// Two processes can access simultaneously without blocking
type Sys6EntangledState struct {
	mu         sync.RWMutex
	value      atomic.Value
	accessLog  []Sys6EntangledAccess
	maxLogSize int
}

// Sys6EntangledAccess records an access to the entangled state
type Sys6EntangledAccess struct {
	ProcessID int
	Timestamp time.Time
	Operation string
	OldValue  interface{}
	NewValue  interface{}
}

// NewSys6EntangledState creates a new entangled state
func NewSys6EntangledState(initialValue interface{}) *Sys6EntangledState {
	es := &Sys6EntangledState{
		accessLog:  make([]Sys6EntangledAccess, 0, 100),
		maxLogSize: 100,
	}
	es.value.Store(initialValue)
	return es
}

// Read performs an entangled read operation
func (es *Sys6EntangledState) Read(processID int) interface{} {
	value := es.value.Load()

	es.mu.Lock()
	es.logAccess(processID, "read", value, value)
	es.mu.Unlock()

	return value
}

// Write performs an entangled write operation
func (es *Sys6EntangledState) Write(processID int, newValue interface{}) {
	oldValue := es.value.Load()
	es.value.Store(newValue)

	es.mu.Lock()
	es.logAccess(processID, "write", oldValue, newValue)
	es.mu.Unlock()
}

// CompareAndSwap performs an atomic compare-and-swap
func (es *Sys6EntangledState) CompareAndSwap(processID int, expected, newValue interface{}) bool {
	es.mu.Lock()
	defer es.mu.Unlock()

	current := es.value.Load()
	if current == expected {
		es.value.Store(newValue)
		es.logAccess(processID, "cas_success", current, newValue)
		return true
	}
	es.logAccess(processID, "cas_fail", current, newValue)
	return false
}

// logAccess logs an access to the entangled state
func (es *Sys6EntangledState) logAccess(processID int, operation string, oldValue, newValue interface{}) {
	access := Sys6EntangledAccess{
		ProcessID: processID,
		Timestamp: time.Now(),
		Operation: operation,
		OldValue:  oldValue,
		NewValue:  newValue,
	}

	if len(es.accessLog) >= es.maxLogSize {
		es.accessLog = es.accessLog[1:]
	}
	es.accessLog = append(es.accessLog, access)
}

// GetAccessLog returns the access log
func (es *Sys6EntangledState) GetAccessLog() []Sys6EntangledAccess {
	es.mu.RLock()
	defer es.mu.RUnlock()

	log := make([]Sys6EntangledAccess, len(es.accessLog))
	copy(log, es.accessLog)
	return log
}

// =============================================================================
// THREAD MULTIPLEXER
// =============================================================================

// Sys6ThreadMultiplexer manages the cycling through permutations
// of the four particular sets with entangled concurrency
type Sys6ThreadMultiplexer struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Permutation cycles
	dyadicPerms   []DyadicPermutation
	triadicPerms1 []TriadicPermutation // MP1
	triadicPerms2 []TriadicPermutation // MP2 (complementary)

	// Current indices
	dyadicIndex   int
	triadicIndex1 int
	triadicIndex2 int

	// Entangled states for each particular set
	particularStates [4]*Sys6EntangledState

	// Thread workers
	workers [4]*Sys6ThreadWorker

	// Metrics
	multiplexCycles uint64
	entanglements   uint64

	// Running state
	running bool
}

// Sys6ThreadWorker represents a single thread in the multiplexer
type Sys6ThreadWorker struct {
	mu        sync.RWMutex
	id        int
	state     *Sys6EntangledState
	active    bool
	processed uint64
}

// NewSys6ThreadMultiplexer creates a new thread multiplexer
func NewSys6ThreadMultiplexer(ctx context.Context) *Sys6ThreadMultiplexer {
	muxCtx, cancel := context.WithCancel(ctx)

	tm := &Sys6ThreadMultiplexer{
		ctx:           muxCtx,
		cancel:        cancel,
		dyadicPerms:   GetDyadicPermutations(),
		triadicPerms1: GetTriadicPermutations(),
		triadicPerms2: GetComplementaryTriadicPermutations(),
	}

	// Initialize particular states
	for i := 0; i < 4; i++ {
		tm.particularStates[i] = NewSys6EntangledState(0.5)
	}

	// Initialize workers
	for i := 0; i < 4; i++ {
		tm.workers[i] = &Sys6ThreadWorker{
			id:    i + 1,
			state: tm.particularStates[i],
		}
	}

	return tm
}

// Start begins the thread multiplexer
func (tm *Sys6ThreadMultiplexer) Start() error {
	tm.mu.Lock()
	if tm.running {
		tm.mu.Unlock()
		return fmt.Errorf("thread multiplexer already running")
	}
	tm.running = true
	tm.mu.Unlock()

	fmt.Println("🔄 Thread Multiplexer starting...")
	fmt.Printf("   Dyadic permutations: %d\n", len(tm.dyadicPerms))
	fmt.Printf("   Triadic permutations (MP1): %d\n", len(tm.triadicPerms1))
	fmt.Printf("   Triadic permutations (MP2): %d\n", len(tm.triadicPerms2))

	// Activate all workers
	for _, worker := range tm.workers {
		worker.mu.Lock()
		worker.active = true
		worker.mu.Unlock()
	}

	return nil
}

// Stop halts the thread multiplexer
func (tm *Sys6ThreadMultiplexer) Stop() error {
	tm.mu.Lock()
	if !tm.running {
		tm.mu.Unlock()
		return nil
	}
	tm.running = false
	tm.mu.Unlock()

	tm.cancel()

	// Deactivate all workers
	for _, worker := range tm.workers {
		worker.mu.Lock()
		worker.active = false
		worker.mu.Unlock()
	}

	fmt.Println("🔄 Thread Multiplexer stopped")
	fmt.Printf("   Total multiplex cycles: %d\n", tm.multiplexCycles)
	fmt.Printf("   Total entanglements: %d\n", tm.entanglements)

	return nil
}

// ExecuteDyadicCycle executes one dyadic permutation cycle
func (tm *Sys6ThreadMultiplexer) ExecuteDyadicCycle(input float64) (float64, float64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.running {
		return 0, 0
	}

	// Get current dyadic permutation
	perm := tm.dyadicPerms[tm.dyadicIndex]

	// Execute entangled access on both particular sets
	var wg sync.WaitGroup
	var result1, result2 float64

	wg.Add(2)

	// Process 1 accesses P1
	go func() {
		defer wg.Done()
		state := tm.particularStates[perm.P1-1]
		current := state.Read(1).(float64)
		newVal := (current + input) / 2
		state.Write(1, newVal)
		result1 = newVal
		tm.workers[perm.P1-1].processed++
	}()

	// Process 2 accesses P2 (entangled - simultaneous)
	go func() {
		defer wg.Done()
		state := tm.particularStates[perm.P2-1]
		current := state.Read(2).(float64)
		newVal := (current + input) / 2
		state.Write(2, newVal)
		result2 = newVal
		tm.workers[perm.P2-1].processed++
	}()

	wg.Wait()

	// Advance dyadic index
	tm.dyadicIndex = (tm.dyadicIndex + 1) % len(tm.dyadicPerms)
	tm.entanglements++

	if tm.dyadicIndex == 0 {
		tm.multiplexCycles++
	}

	return result1, result2
}

// ExecuteTriadicCycle executes one triadic permutation cycle
func (tm *Sys6ThreadMultiplexer) ExecuteTriadicCycle(input float64) [3]float64 {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.running {
		return [3]float64{}
	}

	// Get current triadic permutations from both MP1 and MP2
	perm1 := tm.triadicPerms1[tm.triadicIndex1]
	perm2 := tm.triadicPerms2[tm.triadicIndex2]

	var result [3]float64
	var wg sync.WaitGroup

	// Execute MP1 triad
	wg.Add(3)
	for i, p := range []int{perm1.P1, perm1.P2, perm1.P3} {
		go func(idx, particularSet int) {
			defer wg.Done()
			state := tm.particularStates[particularSet-1]
			current := state.Read(1).(float64)
			newVal := (current + input) / 2
			state.Write(1, newVal)

			// Each MP1 goroutine writes a distinct result index. The parent
			// ExecuteTriadicCycle call already holds tm.mu while waiting for this
			// worker group, so re-entering tm.mu here deadlocks the Sys6 engine.
			result[idx] = newVal

			tm.workers[particularSet-1].processed++
		}(i, p)
	}

	// Execute MP2 triad (complementary, entangled with MP1)
	for _, p := range []int{perm2.P1, perm2.P2, perm2.P3} {
		go func(particularSet int) {
			state := tm.particularStates[particularSet-1]
			current := state.Read(2).(float64)
			newVal := (current + input*0.5) / 2
			state.Write(2, newVal)
			tm.workers[particularSet-1].processed++
		}(p)
	}

	wg.Wait()

	// Advance triadic indices
	tm.triadicIndex1 = (tm.triadicIndex1 + 1) % len(tm.triadicPerms1)
	tm.triadicIndex2 = (tm.triadicIndex2 + 1) % len(tm.triadicPerms2)
	tm.entanglements += 2 // Two entangled triads

	return result
}

// GetCurrentPermutations returns the current permutation state
func (tm *Sys6ThreadMultiplexer) GetCurrentPermutations() map[string]interface{} {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return map[string]interface{}{
		"dyadic_index":        tm.dyadicIndex,
		"dyadic_current":      tm.dyadicPerms[tm.dyadicIndex],
		"triadic_mp1_index":   tm.triadicIndex1,
		"triadic_mp1_current": tm.triadicPerms1[tm.triadicIndex1],
		"triadic_mp2_index":   tm.triadicIndex2,
		"triadic_mp2_current": tm.triadicPerms2[tm.triadicIndex2],
	}
}

// GetMetrics returns the multiplexer metrics
func (tm *Sys6ThreadMultiplexer) GetMetrics() map[string]interface{} {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	workerStats := make([]map[string]interface{}, 4)
	for i, worker := range tm.workers {
		worker.mu.RLock()
		workerStats[i] = map[string]interface{}{
			"id":        worker.id,
			"active":    worker.active,
			"processed": worker.processed,
		}
		worker.mu.RUnlock()
	}

	return map[string]interface{}{
		"multiplex_cycles": tm.multiplexCycles,
		"entanglements":    tm.entanglements,
		"running":          tm.running,
		"workers":          workerStats,
	}
}

// GetParticularState returns the state of a particular set
func (tm *Sys6ThreadMultiplexer) GetParticularState(setID int) interface{} {
	if setID < 1 || setID > 4 {
		return nil
	}
	return tm.particularStates[setID-1].Read(0)
}

// SetParticularState sets the state of a particular set
func (tm *Sys6ThreadMultiplexer) SetParticularState(setID int, value interface{}) {
	if setID < 1 || setID > 4 {
		return
	}
	tm.particularStates[setID-1].Write(0, value)
}

// =============================================================================
// INTEGRATED SYS6 MULTIPLEXED ENGINE
// =============================================================================

// Sys6MultiplexedEngine combines the triality engine with thread multiplexing
type Sys6MultiplexedEngine struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Core engines
	triality    *Sys6TrialityEngine
	multiplexer *Sys6ThreadMultiplexer

	// Integration state
	integrationVector [4]float64

	// Metrics
	totalOperations uint64
	coherence       float64

	// Running state
	running bool
}

// NewSys6MultiplexedEngine creates a new integrated sys6 engine
func NewSys6MultiplexedEngine() *Sys6MultiplexedEngine {
	ctx, cancel := context.WithCancel(context.Background())

	return &Sys6MultiplexedEngine{
		ctx:               ctx,
		cancel:            cancel,
		triality:          NewSys6TrialityEngine(),
		multiplexer:       NewSys6ThreadMultiplexer(ctx),
		integrationVector: [4]float64{0.5, 0.5, 0.5, 0.5},
		coherence:         1.0,
	}
}

// Start begins the integrated sys6 engine
func (s6m *Sys6MultiplexedEngine) Start() error {
	s6m.mu.Lock()
	if s6m.running {
		s6m.mu.Unlock()
		return fmt.Errorf("sys6 multiplexed engine already running")
	}
	s6m.running = true
	s6m.mu.Unlock()

	fmt.Println("⚡ Sys6 Multiplexed Engine starting...")

	// Start triality engine
	if err := s6m.triality.Start(); err != nil {
		return err
	}

	// Start multiplexer
	if err := s6m.multiplexer.Start(); err != nil {
		s6m.triality.Stop()
		return err
	}

	// Set up integration callbacks
	s6m.triality.SetCallbacks(
		func(phase Sys6Phase) {
			s6m.onPhaseChange(phase)
		},
		func(stage TransformationStage) {
			s6m.onStageChange(stage)
		},
		func(cycle uint64) {
			s6m.onCycleComplete(cycle)
		},
		func(gestalt [3]float64) {
			s6m.onEmergence(gestalt)
		},
	)

	fmt.Println("⚡ Sys6 Multiplexed Engine started")

	return nil
}

// Stop halts the integrated sys6 engine
func (s6m *Sys6MultiplexedEngine) Stop() error {
	s6m.mu.Lock()
	if !s6m.running {
		s6m.mu.Unlock()
		return nil
	}
	s6m.running = false
	s6m.mu.Unlock()

	s6m.cancel()
	s6m.triality.Stop()
	s6m.multiplexer.Stop()

	fmt.Println("⚡ Sys6 Multiplexed Engine stopped")

	return nil
}

// onPhaseChange handles phase transitions
func (s6m *Sys6MultiplexedEngine) onPhaseChange(phase Sys6Phase) {
	// Execute dyadic cycle on phase change
	gestalt := s6m.triality.GetGestaltVector()
	input := (gestalt[0] + gestalt[1] + gestalt[2]) / 3
	s6m.multiplexer.ExecuteDyadicCycle(input)
}

// onStageChange handles stage transitions
func (s6m *Sys6MultiplexedEngine) onStageChange(stage TransformationStage) {
	// Execute triadic cycle on stage change
	gestalt := s6m.triality.GetGestaltVector()
	input := (gestalt[0] + gestalt[1] + gestalt[2]) / 3
	result := s6m.multiplexer.ExecuteTriadicCycle(input)

	// Update integration vector
	s6m.mu.Lock()
	for i := 0; i < 3; i++ {
		s6m.integrationVector[i] = result[i]
	}
	s6m.integrationVector[3] = input
	s6m.mu.Unlock()
}

// onCycleComplete handles cycle completion
func (s6m *Sys6MultiplexedEngine) onCycleComplete(cycle uint64) {
	s6m.mu.Lock()
	s6m.totalOperations++
	s6m.mu.Unlock()
}

// onEmergence handles emergence events
func (s6m *Sys6MultiplexedEngine) onEmergence(gestalt [3]float64) {
	// Inject emergence into triality engine
	s6m.triality.InjectInput(gestalt)
}

// GetState returns the current state of the integrated engine
func (s6m *Sys6MultiplexedEngine) GetState() map[string]interface{} {
	s6m.mu.RLock()
	defer s6m.mu.RUnlock()

	return map[string]interface{}{
		"triality":           s6m.triality.GetState(),
		"multiplexer":        s6m.multiplexer.GetMetrics(),
		"integration_vector": s6m.integrationVector,
		"total_operations":   s6m.totalOperations,
		"coherence":          s6m.coherence,
		"running":            s6m.running,
	}
}

// GetMetrics returns the integrated engine metrics
func (s6m *Sys6MultiplexedEngine) GetMetrics() map[string]interface{} {
	s6m.mu.RLock()
	defer s6m.mu.RUnlock()

	return map[string]interface{}{
		"triality_metrics":    s6m.triality.GetMetrics(),
		"multiplexer_metrics": s6m.multiplexer.GetMetrics(),
		"total_operations":    s6m.totalOperations,
		"coherence":           s6m.coherence,
	}
}

// ContributeToGestalt returns the sys6 contribution to the global gestalt
func (s6m *Sys6MultiplexedEngine) ContributeToGestalt() map[string]interface{} {
	s6m.mu.RLock()
	defer s6m.mu.RUnlock()

	return map[string]interface{}{
		"sys6_multiplexed": map[string]interface{}{
			"triality":           s6m.triality.ContributeToGestalt(),
			"integration_vector": s6m.integrationVector,
			"coherence":          s6m.coherence,
			"operations":         s6m.totalOperations,
		},
	}
}
