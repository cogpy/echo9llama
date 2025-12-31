package echobeats

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EchoBeatsThreePhase implements the 3 concurrent streams with 12-step cognitive loop
// Based on Deep Tree Echo architecture with 120-degree phase offset
type EchoBeatsThreePhase struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	
	// Three concurrent cognitive streams
	stream1         *ThreePhaseStream
	stream2         *ThreePhaseStream
	stream3         *ThreePhaseStream
	
	// Shared cognitive state
	sharedState     *SharedCognitiveState
	
	// Synchronization
	synchronizer    *StreamSynchronizer
	
	// Timing
	stepDuration    time.Duration
	cycleCount      uint64
	
	// Metrics
	metrics         *ThreePhaseMetrics
	
	// Control
	running         bool
	paused          bool
}

// ThreePhaseStream represents one of the three concurrent streams
type ThreePhaseStream struct {
	mu              sync.RWMutex
	id              int
	ctx             context.Context
	
	// Current state
	currentStep     int              // 1-12
	phaseOffset     int              // 0, 4, or 8 (120-degree offset)
	mode            CognitiveMode
	
	// Step processing
	stepProcessors  map[int]StepProcessor
	
	// Stream state
	state           *CognitiveState
	
	// Communication
	sharedState     *SharedCognitiveState
	outputChannel   chan StreamOutput
	
	// Metrics
	stepsProcessed  uint64
	cyclesCompleted uint64
}

// StreamSynchronizer coordinates the three streams at triad points
type StreamSynchronizer struct {
	mu              sync.Mutex
	
	// Triad barriers: {1,5,9}, {2,6,10}, {3,7,11}, {4,8,12}
	triadBarriers   map[int]*sync.WaitGroup
	
	// Stream readiness
	streamsReady    map[int]bool
	
	// Triad coordination
	triadSteps      [][]int
}

// StreamOutput represents output from a cognitive stream
type StreamOutput struct {
	StreamID        int
	Step            int
	Mode            CognitiveMode
	Output          interface{}
	StateUpdates    map[string]interface{}
	Timestamp       time.Time
}

// ThreePhaseMetrics tracks metrics for the three-phase system
type ThreePhaseMetrics struct {
	mu                  sync.RWMutex
	totalCycles         uint64
	totalSteps          uint64
	triadSynchronizations uint64
	averageCycleDuration time.Duration
	streamMetrics       map[int]*StreamMetrics
}

// StreamMetrics tracks metrics for individual streams
type StreamMetrics struct {
	StreamID        int
	StepsProcessed  uint64
	CyclesCompleted uint64
	AverageStepTime time.Duration
	LastStepTime    time.Time
}

// NewEchoBeatsThreePhase creates a new three-phase cognitive system
func NewEchoBeatsThreePhase() *EchoBeatsThreePhase {
	ctx, cancel := context.WithCancel(context.Background())
	
	sharedState := &SharedCognitiveState{
		currentAttention: nil,
		attentionWeight:  0.0,
		pastContext:      make([]interface{}, 0),
		presentFocus:     nil,
		futureOptions:    make([]interface{}, 0),
		coherenceScore:   1.0,
		integrationLevel: 0.0,
		currentStep:      1,
		pivotalStepReached: false,
	}
	
	synchronizer := NewStreamSynchronizer()
	
	ebtp := &EchoBeatsThreePhase{
		ctx:          ctx,
		cancel:       cancel,
		sharedState:  sharedState,
		synchronizer: synchronizer,
		stepDuration: 2 * time.Second,
		cycleCount:   0,
		metrics:      NewThreePhaseMetrics(),
		running:      false,
		paused:       false,
	}
	
	// Create three streams with 120-degree phase offset (4 steps apart)
	ebtp.stream1 = ebtp.createStream(1, 0, ctx)  // Starts at step 1
	ebtp.stream2 = ebtp.createStream(2, 4, ctx)  // Starts at step 5 (4 steps ahead)
	ebtp.stream3 = ebtp.createStream(3, 8, ctx)  // Starts at step 9 (8 steps ahead)
	
	return ebtp
}

// createStream creates a new cognitive stream with phase offset
func (ebtp *EchoBeatsThreePhase) createStream(id int, phaseOffset int, ctx context.Context) *ThreePhaseStream {
	stream := &ThreePhaseStream{
		id:             id,
		ctx:            ctx,
		currentStep:    (phaseOffset % 12) + 1,
		phaseOffset:    phaseOffset,
		mode:           ModeExpressive,
		stepProcessors: make(map[int]StepProcessor),
		state: &CognitiveState{
			Timestamp:       time.Now(),
			CycleNumber:     0,
			StepNumber:      (phaseOffset % 12) + 1,
			Mode:            ModeExpressive,
			Attention:       make([]string, 0),
			WorkingMemory:   make(map[string]interface{}),
			EmotionalTone:   make(map[string]float64),
			CognitiveLoad:   0.0,
			RelevanceScores: make(map[string]float64),
			ActiveGoals:     make([]string, 0),
			PendingActions:  make([]string, 0),
			Insights:        make([]string, 0),
		},
		sharedState:    ebtp.sharedState,
		outputChannel:  make(chan StreamOutput, 100),
		stepsProcessed: 0,
		cyclesCompleted: 0,
	}
	
	// Register default step processors
	stream.registerDefaultProcessors()
	
	return stream
}

// registerDefaultProcessors registers the 12 step processors for a stream
func (cs *ThreePhaseStream) registerDefaultProcessors() {
	// Steps 1-4: Expressive Mode (Actual Affordance Interaction)
	cs.stepProcessors[1] = &PerceptionProcessor{}
	cs.stepProcessors[2] = &MemoryActivationProcessor{}
	cs.stepProcessors[3] = &ActionGenerationProcessor{}
	cs.stepProcessors[4] = &ActionExecutionProcessor{}
	
	// Step 5: Pivotal Relevance Realization
	cs.stepProcessors[5] = &RelevanceRealizationProcessor{phase: "present_commitment"}
	
	// Steps 6-10: Reflective Mode (Virtual Salience Simulation)
	cs.stepProcessors[6] = &ScenarioSimulationProcessor{}
	cs.stepProcessors[7] = &OutcomeEvaluationProcessor{}
	cs.stepProcessors[8] = &ModelUpdateProcessor{}
	cs.stepProcessors[9] = &LearningConsolidationProcessor{}
	cs.stepProcessors[10] = &InsightGenerationProcessor{}
	
	// Step 11: Pivotal Relevance Realization
	cs.stepProcessors[11] = &RelevanceRealizationProcessor{phase: "future_commitment"}
	
	// Step 12: Meta-Cognitive Reflection
	cs.stepProcessors[12] = &MetaCognitiveProcessor{}
}

// NewStreamSynchronizer creates a new stream synchronizer
func NewStreamSynchronizer() *StreamSynchronizer {
	ss := &StreamSynchronizer{
		triadBarriers: make(map[int]*sync.WaitGroup),
		streamsReady:  make(map[int]bool),
		triadSteps: [][]int{
			{1, 5, 9},
			{2, 6, 10},
			{3, 7, 11},
			{4, 8, 12},
		},
	}
	
	// Initialize barriers for each triad
	for i := 1; i <= 12; i++ {
		ss.triadBarriers[i] = &sync.WaitGroup{}
	}
	
	return ss
}

// Start begins the three-phase cognitive system
func (ebtp *EchoBeatsThreePhase) Start() error {
	ebtp.mu.Lock()
	if ebtp.running {
		ebtp.mu.Unlock()
		return fmt.Errorf("three-phase system already running")
	}
	ebtp.running = true
	ebtp.mu.Unlock()
	
	fmt.Println("🌀 EchoBeatsThreePhase: Starting 3 concurrent cognitive streams...")
	fmt.Printf("   Stream 1: Starts at step 1 (0° phase)\n")
	fmt.Printf("   Stream 2: Starts at step 5 (120° phase)\n")
	fmt.Printf("   Stream 3: Starts at step 9 (240° phase)\n")
	fmt.Printf("   Triad synchronization at: {1,5,9}, {2,6,10}, {3,7,11}, {4,8,12}\n")
	fmt.Printf("   Step Duration: %v\n\n", ebtp.stepDuration)
	
	// Start all three streams concurrently
	go ebtp.runStream(ebtp.stream1)
	go ebtp.runStream(ebtp.stream2)
	go ebtp.runStream(ebtp.stream3)
	
	// Start metrics collector
	go ebtp.collectMetrics()
	
	return nil
}

// Stop gracefully stops the three-phase system
func (ebtp *EchoBeatsThreePhase) Stop() error {
	ebtp.mu.Lock()
	defer ebtp.mu.Unlock()
	
	if !ebtp.running {
		return fmt.Errorf("three-phase system not running")
	}
	
	fmt.Println("🌀 EchoBeatsThreePhase: Stopping...")
	ebtp.running = false
	ebtp.cancel()
	
	return nil
}

// runStream executes a cognitive stream
func (ebtp *EchoBeatsThreePhase) runStream(stream *ThreePhaseStream) {
	ticker := time.NewTicker(ebtp.stepDuration)
	defer ticker.Stop()
	
	for {
		select {
		case <-ebtp.ctx.Done():
			return
		case <-ticker.C:
			ebtp.mu.RLock()
			isPaused := ebtp.paused
			ebtp.mu.RUnlock()
			
			if !isPaused {
				ebtp.executeStreamStep(stream)
			}
		}
	}
}

// executeStreamStep executes one step in a cognitive stream
func (ebtp *EchoBeatsThreePhase) executeStreamStep(stream *ThreePhaseStream) {
	stream.mu.Lock()
	step := stream.currentStep
	processor := stream.stepProcessors[step]
	state := stream.state
	stream.mu.Unlock()
	
	if processor == nil {
		fmt.Printf("⚠️  Stream %d: No processor for step %d\n", stream.id, step)
		ebtp.advanceStreamStep(stream)
		return
	}
	
	startTime := time.Now()
	
	// Update state
	state.StepNumber = step
	state.Mode = processor.GetMode()
	state.Timestamp = startTime
	
	// Check if this is a triad synchronization point
	ebtp.synchronizer.WaitForTriad(step, stream.id)
	
	// Execute step
	result, err := processor.Process(ebtp.ctx, state)
	
	duration := time.Since(startTime)
	
	// Record execution
	if result != nil {
		// Apply state updates
		ebtp.applyStreamStateUpdates(stream, result.StateUpdates)
		
		// Update cognitive load
		state.CognitiveLoad = result.CognitiveLoad
		
		// Add insights
		if len(result.Insights) > 0 {
			state.Insights = append(state.Insights, result.Insights...)
		}
		
		// Send output
		output := StreamOutput{
			StreamID:     stream.id,
			Step:         step,
			Mode:         processor.GetMode(),
			Output:       result.Output,
			StateUpdates: result.StateUpdates,
			Timestamp:    time.Now(),
		}
		
		select {
		case stream.outputChannel <- output:
		default:
			// Channel full, skip
		}
	}
	
	// Update metrics
	stream.mu.Lock()
	stream.stepsProcessed++
	stream.mu.Unlock()
	
	// Log step completion
	modeEmoji := ebtp.getModeEmoji(processor.GetMode())
	fmt.Printf("%s Stream %d - Step %2d/%2d: %s (%.2fs)\n",
		modeEmoji, stream.id, step, 12, processor.GetDescription(), duration.Seconds())
	
	if err != nil {
		fmt.Printf("   ⚠️  Error: %v\n", err)
	}
	
	// Advance to next step
	ebtp.advanceStreamStep(stream)
}

// advanceStreamStep moves a stream to the next step
func (ebtp *EchoBeatsThreePhase) advanceStreamStep(stream *ThreePhaseStream) {
	stream.mu.Lock()
	defer stream.mu.Unlock()
	
	stream.currentStep++
	
	if stream.currentStep > 12 {
		// Cycle complete for this stream
		stream.currentStep = 1
		stream.cyclesCompleted++
		stream.state.CycleNumber = stream.cyclesCompleted
		
		fmt.Printf("\n🔄 Stream %d: Cycle %d complete\n", stream.id, stream.cyclesCompleted)
		fmt.Printf("   Insights: %d, Cognitive Load: %.2f\n\n", 
			len(stream.state.Insights), stream.state.CognitiveLoad)
		
		// Reset cycle-specific state
		stream.state.Insights = make([]string, 0)
		
		// Update global cycle count (only stream 1 increments)
		if stream.id == 1 {
			ebtp.mu.Lock()
			ebtp.cycleCount++
			ebtp.mu.Unlock()
		}
	}
}

// applyStreamStateUpdates applies updates to stream state
func (ebtp *EchoBeatsThreePhase) applyStreamStateUpdates(stream *ThreePhaseStream, updates map[string]interface{}) {
	if updates == nil {
		return
	}
	
	stream.mu.Lock()
	defer stream.mu.Unlock()
	
	for key, value := range updates {
		stream.state.WorkingMemory[key] = value
	}
}

// WaitForTriad synchronizes streams at triad points
func (ss *StreamSynchronizer) WaitForTriad(step int, streamID int) {
	// Check if this step is a triad synchronization point
	isTriadPoint := false
	for _, triad := range ss.triadSteps {
		for _, triadStep := range triad {
			if triadStep == step {
				isTriadPoint = true
				break
			}
		}
		if isTriadPoint {
			break
		}
	}
	
	if !isTriadPoint {
		return
	}
	
	// Wait for all three streams at this triad point
	ss.mu.Lock()
	barrier := ss.triadBarriers[step]
	barrier.Add(1)
	ss.mu.Unlock()
	
	barrier.Done()
	barrier.Wait()
}

// collectMetrics collects metrics from all streams
func (ebtp *EchoBeatsThreePhase) collectMetrics() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ebtp.ctx.Done():
			return
		case <-ticker.C:
			ebtp.updateMetrics()
		}
	}
}

// updateMetrics updates system metrics
func (ebtp *EchoBeatsThreePhase) updateMetrics() {
	ebtp.metrics.mu.Lock()
	defer ebtp.metrics.mu.Unlock()
	
	ebtp.metrics.totalCycles = ebtp.cycleCount
	ebtp.metrics.totalSteps = ebtp.stream1.stepsProcessed + 
		ebtp.stream2.stepsProcessed + ebtp.stream3.stepsProcessed
}

// getModeEmoji returns emoji for cognitive mode
func (ebtp *EchoBeatsThreePhase) getModeEmoji(mode CognitiveMode) string {
	switch mode {
	case ModeExpressive:
		return "🎭"
	case ModeReflective:
		return "🤔"
	case ModeRelevanceRealization:
		return "🎯"
	case ModeMetaCognitive:
		return "🧠"
	default:
		return "⚙️"
	}
}

// GetMetrics returns three-phase system metrics
func (ebtp *EchoBeatsThreePhase) GetMetrics() map[string]interface{} {
	ebtp.mu.RLock()
	defer ebtp.mu.RUnlock()
	
	ebtp.metrics.mu.RLock()
	defer ebtp.metrics.mu.RUnlock()
	
	return map[string]interface{}{
		"cycle_count":    ebtp.cycleCount,
		"total_steps":    ebtp.metrics.totalSteps,
		"stream1_cycles": ebtp.stream1.cyclesCompleted,
		"stream2_cycles": ebtp.stream2.cyclesCompleted,
		"stream3_cycles": ebtp.stream3.cyclesCompleted,
		"running":        ebtp.running,
		"paused":         ebtp.paused,
	}
}

// NewThreePhaseMetrics creates new metrics tracker
func NewThreePhaseMetrics() *ThreePhaseMetrics {
	return &ThreePhaseMetrics{
		totalCycles:           0,
		totalSteps:            0,
		triadSynchronizations: 0,
		averageCycleDuration:  0,
		streamMetrics:         make(map[int]*StreamMetrics),
	}
}
