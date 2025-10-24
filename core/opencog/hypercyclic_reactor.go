package opencog

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// HypercyclicReactor implements the hypercyclic cognitive fusion reactor
// with autocatalytic inference engine accelerator dynamics
type HypercyclicReactor struct {
	mu sync.RWMutex
	
	// Core reactor components
	ID                string
	AtomSpace         *AtomSpace
	ReactionCycles    map[string]*ReactionCycle
	CatalystPool      map[string]*Catalyst
	
	// Autocatalytic dynamics
	AutocatalyticRate float64
	FusionEnergy      float64
	ReactionThreshold float64
	
	// Inference engine
	InferenceEngine   *InferenceEngine
	InferenceQueue    chan *InferenceTask
	
	// Temporal compression
	TemporalCompressor *TemporalCompressor
	CompressionRatio   float64
	
	// Parallel execution
	WorkerPool        *WorkerPool
	MaxConcurrency    int
	
	// Performance metrics
	Metrics           *ReactorMetrics
	
	// Lifecycle
	Running           bool
	Created           time.Time
	LastReaction      time.Time
}

// ReactionCycle represents a hypercyclic reaction cycle
type ReactionCycle struct {
	ID            string
	Reactants     []string // Atom IDs
	Products      []string // Atom IDs
	Catalysts     []string // Catalyst IDs
	Rate          float64
	Energy        float64
	Iterations    int64
	LastExecution time.Time
	Active        bool
}

// Catalyst represents an autocatalytic agent
type Catalyst struct {
	ID           string
	Type         CatalystType
	Efficiency   float64
	Specificity  map[string]float64 // Reaction ID -> specificity
	State        CatalystState
	Created      time.Time
	LastActive   time.Time
}

// CatalystType defines catalyst types
type CatalystType string

const (
	MetabolicCatalyst    CatalystType = "Metabolic"
	ReplicativeCatalyst  CatalystType = "Replicative"
	RegulatatoryCatalyst CatalystType = "Regulatory"
	InformationalCatalyst CatalystType = "Informational"
)

// CatalystState represents catalyst state
type CatalystState string

const (
	ActiveState   CatalystState = "Active"
	InactiveState CatalystState = "Inactive"
	SaturatedState CatalystState = "Saturated"
)

// InferenceEngine performs massively parallel distributed inference
type InferenceEngine struct {
	mu sync.RWMutex
	
	// Inference rules
	Rules         map[string]*InferenceRule
	RuleChains    map[string][]*InferenceRule
	
	// PLN (Probabilistic Logic Networks)
	PLNEngine     *PLNEngine
	
	// Forward/Backward chaining
	ForwardChain  *ChainEngine
	BackwardChain *ChainEngine
	
	// Performance
	InferenceCount int64
	LastInference  time.Time
}

// InferenceTask represents an inference task
type InferenceTask struct {
	ID         string
	Type       InferenceType
	Input      []string // Atom IDs
	Goal       string   // Target atom ID
	Context    map[string]interface{}
	Priority   int
	Deadline   time.Time
	ResultChan chan *InferenceResult
}

// InferenceType defines inference types
type InferenceType string

const (
	ForwardInference  InferenceType = "Forward"
	BackwardInference InferenceType = "Backward"
	AbductiveInference InferenceType = "Abductive"
	InductiveInference InferenceType = "Inductive"
	DeductiveInference InferenceType = "Deductive"
)

// InferenceResult represents inference result
type InferenceResult struct {
	TaskID     string
	Success    bool
	Output     []string // Derived atom IDs
	TruthValue *TruthValue
	Cost       float64
	Duration   time.Duration
	Error      error
}

// InferenceRule represents a logical inference rule
type InferenceRule struct {
	ID         string
	Name       string
	Premises   []string // Pattern strings
	Conclusion string   // Pattern string
	TruthValueFormula func([]*TruthValue) *TruthValue
	Cost       float64
	Priority   int
}

// PLNEngine implements Probabilistic Logic Networks
type PLNEngine struct {
	mu sync.RWMutex
	
	// PLN rules
	DeductionRules    []*PLNRule
	InductionRules    []*PLNRule
	AbductionRules    []*PLNRule
	
	// Strength/confidence parameters
	DefaultStrength   float64
	DefaultConfidence float64
	
	// Evidence accumulation
	EvidenceThreshold float64
}

// PLNRule represents a PLN inference rule
type PLNRule struct {
	Name           string
	PremiseTypes   []LinkType
	ConclusionType LinkType
	Formula        func([]*TruthValue) *TruthValue
}

// ChainEngine performs forward/backward chaining
type ChainEngine struct {
	mu sync.RWMutex
	
	Mode         ChainMode
	MaxDepth     int
	MaxBranching int
	Visited      map[string]bool
}

// ChainMode defines chaining mode
type ChainMode string

const (
	ForwardChainMode  ChainMode = "Forward"
	BackwardChainMode ChainMode = "Backward"
)

// TemporalCompressor performs temporal compression for accelerated inference
type TemporalCompressor struct {
	mu sync.RWMutex
	
	// Compression parameters
	CompressionRatio  float64
	BufferSize        int
	Buffer            []*CompressedEvent
	
	// Parallelization strategy
	ParallelStreams   int
	StreamBuffers     map[int][]*CompressedEvent
	
	// Performance
	EventsProcessed   int64
	CompressionGain   float64
}

// CompressedEvent represents a temporally compressed event
type CompressedEvent struct {
	OriginalTime   time.Time
	CompressedTime time.Time
	Event          interface{}
	CompressionFactor float64
}

// WorkerPool manages parallel inference workers
type WorkerPool struct {
	mu sync.RWMutex
	
	Workers       []*InferenceWorker
	TaskQueue     chan *InferenceTask
	ResultQueue   chan *InferenceResult
	MaxWorkers    int
	ActiveWorkers int
	
	// Load balancing
	LoadBalancer  *LoadBalancer
}

// InferenceWorker performs inference in parallel
type InferenceWorker struct {
	ID        int
	TaskCount int64
	Busy      bool
	LastTask  time.Time
}

// LoadBalancer distributes tasks across workers
type LoadBalancer struct {
	Strategy  LoadBalancingStrategy
	WorkLoad  map[int]int64
}

// LoadBalancingStrategy defines load balancing strategies
type LoadBalancingStrategy string

const (
	RoundRobin   LoadBalancingStrategy = "RoundRobin"
	LeastLoaded  LoadBalancingStrategy = "LeastLoaded"
	WeightedLoad LoadBalancingStrategy = "WeightedLoad"
)

// ReactorMetrics tracks reactor performance
type ReactorMetrics struct {
	mu sync.RWMutex
	
	TotalReactions     int64
	ReactionsPerSecond float64
	AverageEnergy      float64
	CompressionGain    float64
	ParallelEfficiency float64
	ThroughputGain     float64
	
	StartTime          time.Time
	LastUpdate         time.Time
}

// NewHypercyclicReactor creates a new hypercyclic cognitive fusion reactor
func NewHypercyclicReactor(atomSpace *AtomSpace, maxConcurrency int) *HypercyclicReactor {
	reactor := &HypercyclicReactor{
		ID:                 fmt.Sprintf("reactor_%d", time.Now().UnixNano()),
		AtomSpace:          atomSpace,
		ReactionCycles:     make(map[string]*ReactionCycle),
		CatalystPool:       make(map[string]*Catalyst),
		AutocatalyticRate:  1.5,
		FusionEnergy:       1.0,
		ReactionThreshold:  0.5,
		InferenceQueue:     make(chan *InferenceTask, 10000),
		MaxConcurrency:     maxConcurrency,
		CompressionRatio:   1000.0, // 1000x acceleration
		Created:            time.Now(),
		Running:            false,
	}
	
	// Initialize inference engine
	reactor.InferenceEngine = NewInferenceEngine()
	
	// Initialize temporal compressor
	reactor.TemporalCompressor = NewTemporalCompressor(1000.0, 1000)
	
	// Initialize worker pool
	reactor.WorkerPool = NewWorkerPool(maxConcurrency)
	
	// Initialize metrics
	reactor.Metrics = &ReactorMetrics{
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
	}
	
	return reactor
}

// NewInferenceEngine creates a new inference engine
func NewInferenceEngine() *InferenceEngine {
	return &InferenceEngine{
		Rules:      make(map[string]*InferenceRule),
		RuleChains: make(map[string][]*InferenceRule),
		PLNEngine:  NewPLNEngine(),
		ForwardChain: &ChainEngine{
			Mode:         ForwardChainMode,
			MaxDepth:     10,
			MaxBranching: 5,
			Visited:      make(map[string]bool),
		},
		BackwardChain: &ChainEngine{
			Mode:         BackwardChainMode,
			MaxDepth:     10,
			MaxBranching: 5,
			Visited:      make(map[string]bool),
		},
	}
}

// NewPLNEngine creates a new PLN engine
func NewPLNEngine() *PLNEngine {
	pln := &PLNEngine{
		DeductionRules:    []*PLNRule{},
		InductionRules:    []*PLNRule{},
		AbductionRules:    []*PLNRule{},
		DefaultStrength:   0.5,
		DefaultConfidence: 0.5,
		EvidenceThreshold: 0.3,
	}
	
	// Add standard PLN rules
	pln.initializeStandardRules()
	
	return pln
}

// initializeStandardRules initializes standard PLN rules
func (pln *PLNEngine) initializeStandardRules() {
	// Deduction: (A→B, B→C) ⊢ A→C
	pln.DeductionRules = append(pln.DeductionRules, &PLNRule{
		Name:           "Deduction",
		PremiseTypes:   []LinkType{ImplicationLink, ImplicationLink},
		ConclusionType: ImplicationLink,
		Formula: func(tvs []*TruthValue) *TruthValue {
			if len(tvs) < 2 {
				return &TruthValue{Strength: 0.5, Confidence: 0.0, Count: 0.0}
			}
			s1, s2 := tvs[0].Strength, tvs[1].Strength
			c1, c2 := tvs[0].Confidence, tvs[1].Confidence
			return &TruthValue{
				Strength:   s1 * s2,
				Confidence: c1 * c2,
				Count:      tvs[0].Count + tvs[1].Count,
			}
		},
	})
	
	// Induction: Multiple instances of A→B increase confidence
	pln.InductionRules = append(pln.InductionRules, &PLNRule{
		Name:           "Induction",
		PremiseTypes:   []LinkType{EvaluationLink},
		ConclusionType: ImplicationLink,
		Formula: func(tvs []*TruthValue) *TruthValue {
			if len(tvs) == 0 {
				return &TruthValue{Strength: 0.5, Confidence: 0.0, Count: 0.0}
			}
			// Accumulate evidence
			totalCount := 0.0
			totalStrength := 0.0
			for _, tv := range tvs {
				totalCount += tv.Count
				totalStrength += tv.Strength * tv.Count
			}
			avgStrength := totalStrength / math.Max(totalCount, 1.0)
			confidence := math.Min(totalCount/100.0, 1.0)
			return &TruthValue{
				Strength:   avgStrength,
				Confidence: confidence,
				Count:      totalCount,
			}
		},
	})
	
	// Abduction: (A→B, B) ⊢ A (with lower confidence)
	// Note: Simplified - using EvaluationLink instead of raw concept
	pln.AbductionRules = append(pln.AbductionRules, &PLNRule{
		Name:           "Abduction",
		PremiseTypes:   []LinkType{ImplicationLink, EvaluationLink},
		ConclusionType: EvaluationLink,
		Formula: func(tvs []*TruthValue) *TruthValue {
			if len(tvs) < 2 {
				return &TruthValue{Strength: 0.5, Confidence: 0.0, Count: 0.0}
			}
			s1, s2 := tvs[0].Strength, tvs[1].Strength
			c1 := tvs[0].Confidence
			// Abduction has lower confidence
			return &TruthValue{
				Strength:   s1 * s2,
				Confidence: c1 * 0.5,
				Count:      tvs[0].Count,
			}
		},
	})
}

// NewTemporalCompressor creates a new temporal compressor
func NewTemporalCompressor(ratio float64, bufferSize int) *TemporalCompressor {
	return &TemporalCompressor{
		CompressionRatio: ratio,
		BufferSize:       bufferSize,
		Buffer:           make([]*CompressedEvent, 0, bufferSize),
		ParallelStreams:  8,
		StreamBuffers:    make(map[int][]*CompressedEvent),
	}
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(maxWorkers int) *WorkerPool {
	pool := &WorkerPool{
		Workers:      make([]*InferenceWorker, maxWorkers),
		TaskQueue:    make(chan *InferenceTask, maxWorkers*10),
		ResultQueue:  make(chan *InferenceResult, maxWorkers*10),
		MaxWorkers:   maxWorkers,
		LoadBalancer: &LoadBalancer{
			Strategy: LeastLoaded,
			WorkLoad: make(map[int]int64),
		},
	}
	
	// Initialize workers
	for i := 0; i < maxWorkers; i++ {
		pool.Workers[i] = &InferenceWorker{
			ID:   i,
			Busy: false,
		}
		pool.LoadBalancer.WorkLoad[i] = 0
	}
	
	return pool
}

// Start starts the hypercyclic reactor
func (hr *HypercyclicReactor) Start(ctx context.Context) error {
	hr.mu.Lock()
	if hr.Running {
		hr.mu.Unlock()
		return fmt.Errorf("reactor already running")
	}
	hr.Running = true
	hr.mu.Unlock()
	
	// Start worker pool
	hr.WorkerPool.Start(ctx, hr)
	
	// Start reaction cycles
	go hr.runReactionCycles(ctx)
	
	// Start inference engine
	go hr.runInferenceEngine(ctx)
	
	// Start temporal compression
	go hr.runTemporalCompression(ctx)
	
	// Start metrics collection
	go hr.collectMetrics(ctx)
	
	return nil
}

// Stop stops the hypercyclic reactor
func (hr *HypercyclicReactor) Stop() {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	
	hr.Running = false
}

// AddReactionCycle adds a new reaction cycle
func (hr *HypercyclicReactor) AddReactionCycle(reactants, products, catalysts []string, rate float64) (*ReactionCycle, error) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	
	id := fmt.Sprintf("cycle_%d", time.Now().UnixNano())
	cycle := &ReactionCycle{
		ID:            id,
		Reactants:     reactants,
		Products:      products,
		Catalysts:     catalysts,
		Rate:          rate,
		Energy:        1.0,
		Iterations:    0,
		LastExecution: time.Time{},
		Active:        true,
	}
	
	hr.ReactionCycles[id] = cycle
	return cycle, nil
}

// AddCatalyst adds a catalyst to the pool
func (hr *HypercyclicReactor) AddCatalyst(catalystType CatalystType, efficiency float64) (*Catalyst, error) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	
	id := fmt.Sprintf("catalyst_%d", time.Now().UnixNano())
	catalyst := &Catalyst{
		ID:          id,
		Type:        catalystType,
		Efficiency:  efficiency,
		Specificity: make(map[string]float64),
		State:       ActiveState,
		Created:     time.Now(),
		LastActive:  time.Now(),
	}
	
	hr.CatalystPool[id] = catalyst
	return catalyst, nil
}

// SubmitInference submits an inference task
func (hr *HypercyclicReactor) SubmitInference(task *InferenceTask) error {
	if !hr.Running {
		return fmt.Errorf("reactor not running")
	}
	
	select {
	case hr.InferenceQueue <- task:
		return nil
	default:
		return fmt.Errorf("inference queue full")
	}
}

// runReactionCycles runs autocatalytic reaction cycles
func (hr *HypercyclicReactor) runReactionCycles(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Millisecond) // High frequency for temporal compression
	defer ticker.Stop()
	
	for hr.Running {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hr.executeReactionCycles()
		}
	}
}

// executeReactionCycles executes all active reaction cycles
func (hr *HypercyclicReactor) executeReactionCycles() {
	hr.mu.RLock()
	cycles := make([]*ReactionCycle, 0, len(hr.ReactionCycles))
	for _, cycle := range hr.ReactionCycles {
		if cycle.Active {
			cycles = append(cycles, cycle)
		}
	}
	hr.mu.RUnlock()
	
	// Execute cycles in parallel
	var wg sync.WaitGroup
	for _, cycle := range cycles {
		wg.Add(1)
		go func(c *ReactionCycle) {
			defer wg.Done()
			hr.executeReactionCycle(c)
		}(cycle)
	}
	wg.Wait()
	
	hr.mu.Lock()
	hr.LastReaction = time.Now()
	hr.mu.Unlock()
}

// executeReactionCycle executes a single reaction cycle
func (hr *HypercyclicReactor) executeReactionCycle(cycle *ReactionCycle) {
	// Apply autocatalytic dynamics
	catalystBoost := 1.0
	for _, catalystID := range cycle.Catalysts {
		if catalyst, exists := hr.CatalystPool[catalystID]; exists {
			if catalyst.State == ActiveState {
				catalystBoost *= (1.0 + catalyst.Efficiency)
			}
		}
	}
	
	effectiveRate := cycle.Rate * catalystBoost * hr.AutocatalyticRate
	
	// Check if reaction should fire
	if effectiveRate > hr.ReactionThreshold {
		// Perform reaction: transform reactants to products in AtomSpace
		for i, reactantID := range cycle.Reactants {
			if i < len(cycle.Products) {
				productID := cycle.Products[i]
				// Update truth values (fusion)
				if atom1, exists1 := hr.AtomSpace.GetAtom(reactantID); exists1 {
					if atom2, exists2 := hr.AtomSpace.GetAtom(productID); exists2 {
						// Fuse truth values
						fusedTV := ComputeTruthValue(atom1.TruthValue, atom2.TruthValue, "and")
						hr.AtomSpace.UpdateTruthValue(productID, fusedTV)
					}
				}
			}
		}
		
		cycle.Iterations++
		cycle.LastExecution = time.Now()
		
		// Energy accumulation
		cycle.Energy *= 0.99
		cycle.Energy += effectiveRate * 0.01
		
		hr.FusionEnergy += effectiveRate * 0.001
	}
}

// runInferenceEngine runs the massively parallel inference engine
func (hr *HypercyclicReactor) runInferenceEngine(ctx context.Context) {
	for hr.Running {
		select {
		case <-ctx.Done():
			return
		case task := <-hr.InferenceQueue:
			// Dispatch to worker pool
			hr.WorkerPool.SubmitTask(task)
		}
	}
}

// runTemporalCompression runs temporal compression
func (hr *HypercyclicReactor) runTemporalCompression(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for hr.Running {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hr.TemporalCompressor.Compress()
		}
	}
}

// collectMetrics collects reactor metrics
func (hr *HypercyclicReactor) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	lastReactions := int64(0)
	
	for hr.Running {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hr.Metrics.mu.Lock()
			
			// Count total reactions
			totalReactions := int64(0)
			for _, cycle := range hr.ReactionCycles {
				totalReactions += cycle.Iterations
			}
			hr.Metrics.TotalReactions = totalReactions
			
			// Calculate reactions per second
			reactionsDelta := totalReactions - lastReactions
			hr.Metrics.ReactionsPerSecond = float64(reactionsDelta)
			lastReactions = totalReactions
			
			// Average energy
			hr.Metrics.AverageEnergy = hr.FusionEnergy / math.Max(float64(len(hr.ReactionCycles)), 1.0)
			
			// Compression gain
			hr.Metrics.CompressionGain = hr.TemporalCompressor.CompressionGain
			
			// Parallel efficiency
			if hr.MaxConcurrency > 0 {
				hr.Metrics.ParallelEfficiency = float64(hr.WorkerPool.ActiveWorkers) / float64(hr.MaxConcurrency)
			}
			
			// Throughput gain (reactions * compression * parallelism)
			hr.Metrics.ThroughputGain = hr.Metrics.ReactionsPerSecond * hr.CompressionRatio * hr.Metrics.ParallelEfficiency
			
			hr.Metrics.LastUpdate = time.Now()
			hr.Metrics.mu.Unlock()
		}
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start(ctx context.Context, reactor *HypercyclicReactor) {
	for i := 0; i < wp.MaxWorkers; i++ {
		go wp.runWorker(ctx, i, reactor)
	}
}

// runWorker runs a single worker
func (wp *WorkerPool) runWorker(ctx context.Context, workerID int, reactor *HypercyclicReactor) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-wp.TaskQueue:
			wp.mu.Lock()
			wp.Workers[workerID].Busy = true
			wp.ActiveWorkers++
			wp.mu.Unlock()
			
			// Execute inference task
			result := wp.executeInferenceTask(task, reactor)
			
			wp.mu.Lock()
			wp.Workers[workerID].Busy = false
			wp.Workers[workerID].TaskCount++
			wp.Workers[workerID].LastTask = time.Now()
			wp.ActiveWorkers--
			wp.LoadBalancer.WorkLoad[workerID]++
			wp.mu.Unlock()
			
			// Send result
			if task.ResultChan != nil {
				select {
				case task.ResultChan <- result:
				default:
				}
			}
		}
	}
}

// SubmitTask submits a task to the worker pool
func (wp *WorkerPool) SubmitTask(task *InferenceTask) {
	select {
	case wp.TaskQueue <- task:
	default:
		// Queue full - handle gracefully
	}
}

// executeInferenceTask executes an inference task
func (wp *WorkerPool) executeInferenceTask(task *InferenceTask, reactor *HypercyclicReactor) *InferenceResult {
	startTime := time.Now()
	
	result := &InferenceResult{
		TaskID:   task.ID,
		Success:  false,
		Output:   []string{},
		Cost:     0.0,
		Duration: 0,
	}
	
	// Execute based on inference type
	switch task.Type {
	case ForwardInference:
		output, tv, err := reactor.InferenceEngine.ForwardChain.Execute(reactor.AtomSpace, task.Input, task.Goal)
		result.Output = output
		result.TruthValue = tv
		result.Error = err
		result.Success = err == nil
		
	case BackwardInference:
		output, tv, err := reactor.InferenceEngine.BackwardChain.Execute(reactor.AtomSpace, task.Input, task.Goal)
		result.Output = output
		result.TruthValue = tv
		result.Error = err
		result.Success = err == nil
		
	case DeductiveInference:
		output, tv := reactor.InferenceEngine.PLNEngine.ApplyDeduction(reactor.AtomSpace, task.Input)
		result.Output = output
		result.TruthValue = tv
		result.Success = len(output) > 0
		
	case InductiveInference:
		output, tv := reactor.InferenceEngine.PLNEngine.ApplyInduction(reactor.AtomSpace, task.Input)
		result.Output = output
		result.TruthValue = tv
		result.Success = len(output) > 0
		
	case AbductiveInference:
		output, tv := reactor.InferenceEngine.PLNEngine.ApplyAbduction(reactor.AtomSpace, task.Input)
		result.Output = output
		result.TruthValue = tv
		result.Success = len(output) > 0
	}
	
	result.Duration = time.Since(startTime)
	result.Cost = result.Duration.Seconds()
	
	reactor.InferenceEngine.InferenceCount++
	reactor.InferenceEngine.LastInference = time.Now()
	
	return result
}

// Execute executes forward/backward chaining
func (ce *ChainEngine) Execute(as *AtomSpace, input []string, goal string) ([]string, *TruthValue, error) {
	// Simplified chaining implementation
	ce.mu.Lock()
	defer ce.mu.Unlock()
	
	ce.Visited = make(map[string]bool)
	
	if ce.Mode == ForwardChainMode {
		return ce.forwardChain(as, input, goal, 0)
	}
	return ce.backwardChain(as, input, goal, 0)
}

// forwardChain performs forward chaining
func (ce *ChainEngine) forwardChain(as *AtomSpace, current []string, goal string, depth int) ([]string, *TruthValue, error) {
	if depth >= ce.MaxDepth {
		return []string{}, nil, fmt.Errorf("max depth reached")
	}
	
	// Check if goal is reached
	for _, atomID := range current {
		if atomID == goal {
			if atom, exists := as.GetAtom(atomID); exists {
				return []string{atomID}, atom.TruthValue, nil
			}
		}
	}
	
	// Find applicable rules and continue chaining
	derived := []string{}
	for _, atomID := range current {
		if !ce.Visited[atomID] {
			ce.Visited[atomID] = true
			// Get incoming links and derive new atoms
			incoming := as.GetIncoming(atomID)
			for _, linkID := range incoming {
				if link, exists := as.GetLink(linkID); exists {
					for _, outgoing := range link.Outgoing {
						if !ce.Visited[outgoing] {
							derived = append(derived, outgoing)
						}
					}
				}
			}
		}
	}
	
	if len(derived) > 0 {
		return ce.forwardChain(as, derived, goal, depth+1)
	}
	
	return []string{}, nil, fmt.Errorf("goal not reached")
}

// backwardChain performs backward chaining
func (ce *ChainEngine) backwardChain(as *AtomSpace, current []string, goal string, depth int) ([]string, *TruthValue, error) {
	if depth >= ce.MaxDepth {
		return []string{}, nil, fmt.Errorf("max depth reached")
	}
	
	// Start from goal and work backwards
	if len(current) == 0 {
		current = []string{goal}
	}
	
	// Check if current premises are satisfied
	satisfied := true
	for _, atomID := range current {
		if atom, exists := as.GetAtom(atomID); exists {
			if atom.TruthValue.Strength < 0.5 {
				satisfied = false
				break
			}
		} else {
			satisfied = false
			break
		}
	}
	
	if satisfied {
		if len(current) > 0 {
			if atom, exists := as.GetAtom(current[0]); exists {
				return current, atom.TruthValue, nil
			}
		}
	}
	
	// Find atoms that could derive current atoms
	premises := []string{}
	for _, atomID := range current {
		incoming := as.GetIncoming(atomID)
		for _, linkID := range incoming {
			if link, exists := as.GetLink(linkID); exists {
				premises = append(premises, link.Outgoing...)
			}
		}
	}
	
	if len(premises) > 0 {
		return ce.backwardChain(as, premises, goal, depth+1)
	}
	
	return []string{}, nil, fmt.Errorf("goal not provable")
}

// ApplyDeduction applies deduction rules
func (pln *PLNEngine) ApplyDeduction(as *AtomSpace, input []string) ([]string, *TruthValue) {
	// Simplified deduction
	output := []string{}
	tvs := []*TruthValue{}
	
	for _, atomID := range input {
		if atom, exists := as.GetAtom(atomID); exists {
			tvs = append(tvs, atom.TruthValue)
		}
	}
	
	if len(pln.DeductionRules) > 0 && len(tvs) > 0 {
		resultTV := pln.DeductionRules[0].Formula(tvs)
		return output, resultTV
	}
	
	return output, &TruthValue{Strength: 0.5, Confidence: 0.0, Count: 0.0}
}

// ApplyInduction applies induction rules
func (pln *PLNEngine) ApplyInduction(as *AtomSpace, input []string) ([]string, *TruthValue) {
	output := []string{}
	tvs := []*TruthValue{}
	
	for _, atomID := range input {
		if atom, exists := as.GetAtom(atomID); exists {
			tvs = append(tvs, atom.TruthValue)
		}
	}
	
	if len(pln.InductionRules) > 0 && len(tvs) > 0 {
		resultTV := pln.InductionRules[0].Formula(tvs)
		return output, resultTV
	}
	
	return output, &TruthValue{Strength: 0.5, Confidence: 0.0, Count: 0.0}
}

// ApplyAbduction applies abduction rules
func (pln *PLNEngine) ApplyAbduction(as *AtomSpace, input []string) ([]string, *TruthValue) {
	output := []string{}
	tvs := []*TruthValue{}
	
	for _, atomID := range input {
		if atom, exists := as.GetAtom(atomID); exists {
			tvs = append(tvs, atom.TruthValue)
		}
	}
	
	if len(pln.AbductionRules) > 0 && len(tvs) > 0 {
		resultTV := pln.AbductionRules[0].Formula(tvs)
		return output, resultTV
	}
	
	return output, &TruthValue{Strength: 0.5, Confidence: 0.0, Count: 0.0}
}

// Compress performs temporal compression
func (tc *TemporalCompressor) Compress() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	
	// Compress events in buffer
	for _, event := range tc.Buffer {
		compressedTime := time.Now().Add(time.Duration(float64(time.Since(event.OriginalTime).Nanoseconds()) / tc.CompressionRatio))
		event.CompressedTime = compressedTime
		event.CompressionFactor = tc.CompressionRatio
	}
	
	tc.EventsProcessed += int64(len(tc.Buffer))
	tc.CompressionGain = tc.CompressionRatio
	
	// Clear buffer
	tc.Buffer = tc.Buffer[:0]
}

// GetMetrics returns reactor metrics
func (hr *HypercyclicReactor) GetMetrics() map[string]interface{} {
	hr.Metrics.mu.RLock()
	defer hr.Metrics.mu.RUnlock()
	
	return map[string]interface{}{
		"total_reactions":      hr.Metrics.TotalReactions,
		"reactions_per_second": hr.Metrics.ReactionsPerSecond,
		"average_energy":       hr.Metrics.AverageEnergy,
		"compression_gain":     hr.Metrics.CompressionGain,
		"parallel_efficiency":  hr.Metrics.ParallelEfficiency,
		"throughput_gain":      hr.Metrics.ThroughputGain,
		"fusion_energy":        hr.FusionEnergy,
		"running":              hr.Running,
		"cycles":               len(hr.ReactionCycles),
		"catalysts":            len(hr.CatalystPool),
		"workers":              hr.MaxConcurrency,
		"active_workers":       hr.WorkerPool.ActiveWorkers,
		"inference_count":      hr.InferenceEngine.InferenceCount,
	}
}
