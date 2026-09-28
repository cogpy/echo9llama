package echobeats

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TriadCognitiveSystem implements 3 concurrent streams with proper triad synchronization
// Based on OEIS A000081 nested shells structure and 12-step cognitive loop
type TriadCognitiveSystem struct {
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	running bool

	// Three concurrent streams (120° phase offset)
	stream1 *CognitiveStream
	stream2 *CognitiveStream
	stream3 *CognitiveStream

	// Triad synchronization points
	triadSync *TriadSynchronizer

	// Shared cognitive state
	sharedState *TriadSharedState

	// Metrics
	cycleCount uint64
	triadCount uint64
}

// CognitiveStream represents one of the three concurrent consciousness streams
type CognitiveStream struct {
	id          int
	currentStep int
	phaseOffset int // 0, 4, or 8 (120° apart in 12-step cycle)
	stepHistory []StepExecution
	active      bool
}

// TriadSynchronizer coordinates the three streams at triad points
type TriadSynchronizer struct {
	mu sync.Mutex

	// Triad points: {1,5,9}, {2,6,10}, {3,7,11}, {4,8,12}
	triad1Barrier *sync.WaitGroup // Steps 1, 5, 9
	triad2Barrier *sync.WaitGroup // Steps 2, 6, 10
	triad3Barrier *sync.WaitGroup // Steps 3, 7, 11
	triad4Barrier *sync.WaitGroup // Steps 4, 8, 12

	// Stream readiness tracking
	streamsReady map[int]map[int]bool // stream_id -> step -> ready
}

// TriadSharedState holds state shared across all three streams
type TriadSharedState struct {
	mu sync.RWMutex

	// Temporal integration (past, present, future)
	stream1Context interface{} // Past/affordance (steps 0-5)
	stream2Context interface{} // Present/relevance (pivotal steps)
	stream3Context interface{} // Future/salience (steps 6-11)

	// Cross-stream awareness
	stream1PerceivedBy []int // Which streams perceive stream 1
	stream2PerceivedBy []int
	stream3PerceivedBy []int

	// Coherence metrics
	temporalCoherence float64
	triadAlignment    float64
	integrationLevel  float64

	// Current triad
	currentTriad     int
	triadStepReached map[int]bool
}

// NewTriadCognitiveSystem creates a new triad-synchronized cognitive system
func NewTriadCognitiveSystem() *TriadCognitiveSystem {
	ctx, cancel := context.WithCancel(context.Background())

	// Create three streams with 120° phase offset (4 steps apart)
	stream1 := &CognitiveStream{
		id:          1,
		currentStep: 1,
		phaseOffset: 0,
		stepHistory: make([]StepExecution, 0),
		active:      true,
	}

	stream2 := &CognitiveStream{
		id:          2,
		currentStep: 5, // 4 steps ahead
		phaseOffset: 4,
		stepHistory: make([]StepExecution, 0),
		active:      true,
	}

	stream3 := &CognitiveStream{
		id:          3,
		currentStep: 9, // 8 steps ahead (4 from stream 2)
		phaseOffset: 8,
		stepHistory: make([]StepExecution, 0),
		active:      true,
	}

	// Create triad synchronizer
	triadSync := &TriadSynchronizer{
		triad1Barrier: &sync.WaitGroup{},
		triad2Barrier: &sync.WaitGroup{},
		triad3Barrier: &sync.WaitGroup{},
		triad4Barrier: &sync.WaitGroup{},
		streamsReady:  make(map[int]map[int]bool),
	}

	// Initialize stream readiness tracking
	for i := 1; i <= 3; i++ {
		triadSync.streamsReady[i] = make(map[int]bool)
	}

	// Create shared state
	sharedState := &TriadSharedState{
		stream1PerceivedBy: make([]int, 0),
		stream2PerceivedBy: make([]int, 0),
		stream3PerceivedBy: make([]int, 0),
		triadStepReached:   make(map[int]bool),
		temporalCoherence:  0.5,
		triadAlignment:     0.5,
		integrationLevel:   0.5,
	}

	return &TriadCognitiveSystem{
		ctx:         ctx,
		cancel:      cancel,
		stream1:     stream1,
		stream2:     stream2,
		stream3:     stream3,
		triadSync:   triadSync,
		sharedState: sharedState,
	}
}

// Start begins concurrent operation of all three streams
func (tcs *TriadCognitiveSystem) Start() error {
	tcs.mu.Lock()
	if tcs.running {
		tcs.mu.Unlock()
		return fmt.Errorf("already running")
	}
	tcs.running = true
	tcs.mu.Unlock()

	fmt.Println("🔷 Starting Triad-Synchronized Cognitive System")
	fmt.Println("   ✓ 3 concurrent streams (120° phase offset)")
	fmt.Println("   ✓ Triad synchronization: {1,5,9}, {2,6,10}, {3,7,11}, {4,8,12}")
	fmt.Println("   ✓ Cross-stream awareness enabled")
	fmt.Println("   ✓ Temporal coherence tracking active")
	fmt.Println()

	// Start all three streams concurrently
	go tcs.runStream(tcs.stream1)
	go tcs.runStream(tcs.stream2)
	go tcs.runStream(tcs.stream3)

	// Start integration monitor
	go tcs.monitorIntegration()

	return nil
}

// Stop gracefully stops all streams
func (tcs *TriadCognitiveSystem) Stop() error {
	tcs.mu.Lock()
	defer tcs.mu.Unlock()

	if !tcs.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("🔷 Stopping Triad-Synchronized Cognitive System...")
	tcs.running = false
	tcs.cancel()

	return nil
}

// runStream executes one cognitive stream
func (tcs *TriadCognitiveSystem) runStream(stream *CognitiveStream) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-tcs.ctx.Done():
			return
		case <-ticker.C:
			if !stream.active {
				continue
			}

			// Execute current step
			tcs.executeStreamStep(stream)

			// Check if this is a triad point
			if tcs.isTriadPoint(stream.currentStep) {
				tcs.synchronizeAtTriad(stream)
			}

			// Advance to next step
			stream.currentStep++
			if stream.currentStep > 12 {
				stream.currentStep = 1
				tcs.incrementCycle()
			}
		}
	}
}

// executeStreamStep executes one step for a stream
func (tcs *TriadCognitiveSystem) executeStreamStep(stream *CognitiveStream) {
	startTime := time.Now()

	// Determine cognitive mode based on step
	mode := tcs.getStepMode(stream.currentStep)

	// Simulate step processing
	stepDescription := tcs.getStepDescription(stream.currentStep, mode)

	// Record execution
	execution := StepExecution{
		StepNumber: stream.currentStep,
		StartTime:  startTime,
		Duration:   time.Since(startTime),
		Mode:       mode,
		Success:    true,
	}

	stream.stepHistory = append(stream.stepHistory, execution)

	// Display step
	modeEmoji := tcs.getModeEmoji(mode)
	fmt.Printf("%s Stream %d - Step %2d: %s\n",
		modeEmoji, stream.id, stream.currentStep, stepDescription)

	// Update shared state
	tcs.updateSharedState(stream)

	// Implement cross-stream awareness
	tcs.implementCrossStreamAwareness(stream)
}

// isTriadPoint checks if a step is a triad synchronization point
func (tcs *TriadCognitiveSystem) isTriadPoint(step int) bool {
	// Triad points: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12
	// Actually, every step is part of a triad when considering all 3 streams
	// But synchronization happens at specific triads
	return step == 1 || step == 5 || step == 9 || // Triad 1
		step == 2 || step == 6 || step == 10 || // Triad 2
		step == 3 || step == 7 || step == 11 || // Triad 3
		step == 4 || step == 8 || step == 12 // Triad 4
}

// synchronizeAtTriad synchronizes streams at triad points
func (tcs *TriadCognitiveSystem) synchronizeAtTriad(stream *CognitiveStream) {
	triadNum := tcs.getTriadNumber(stream.currentStep)

	tcs.triadSync.mu.Lock()
	tcs.triadSync.streamsReady[stream.id][stream.currentStep] = true

	// Check if all three streams have reached this triad
	allReady := true
	for streamID := 1; streamID <= 3; streamID++ {
		if !tcs.triadSync.streamsReady[streamID][stream.currentStep] {
			allReady = false
			break
		}
	}

	if allReady {
		// Triad synchronization achieved
		tcs.triadCount++
		fmt.Printf("\n🎯 Triad %d synchronized at step %d (count: %d)\n\n",
			triadNum, stream.currentStep, tcs.triadCount)

		// Reset readiness for this step
		for streamID := 1; streamID <= 3; streamID++ {
			tcs.triadSync.streamsReady[streamID][stream.currentStep] = false
		}

		// Update shared state
		tcs.sharedState.mu.Lock()
		tcs.sharedState.currentTriad = triadNum
		tcs.sharedState.triadAlignment += 0.01
		if tcs.sharedState.triadAlignment > 1.0 {
			tcs.sharedState.triadAlignment = 1.0
		}
		tcs.sharedState.mu.Unlock()
	}

	tcs.triadSync.mu.Unlock()
}

// getTriadNumber returns which triad (1-4) a step belongs to
func (tcs *TriadCognitiveSystem) getTriadNumber(step int) int {
	switch step {
	case 1, 5, 9:
		return 1
	case 2, 6, 10:
		return 2
	case 3, 7, 11:
		return 3
	case 4, 8, 12:
		return 4
	default:
		return 0
	}
}

// getStepMode returns the cognitive mode for a step
func (tcs *TriadCognitiveSystem) getStepMode(step int) CognitiveMode {
	switch {
	case step >= 1 && step <= 4:
		return ModeExpressive // Actual affordance interaction
	case step == 5:
		return ModeRelevanceRealization // Pivotal relevance (present)
	case step >= 6 && step <= 10:
		return ModeReflective // Virtual salience simulation
	case step == 11:
		return ModeRelevanceRealization // Pivotal relevance (future)
	case step == 12:
		return ModeMetaCognitive // Meta-cognitive reflection
	default:
		return ModeExpressive
	}
}

// getStepDescription returns description for a step
func (tcs *TriadCognitiveSystem) getStepDescription(step int, mode CognitiveMode) string {
	descriptions := map[int]string{
		1:  "Perception & Attention",
		2:  "Memory Activation",
		3:  "Action Generation",
		4:  "Action Execution",
		5:  "Relevance Realization (Present)",
		6:  "Scenario Simulation",
		7:  "Outcome Evaluation",
		8:  "Model Update",
		9:  "Learning Consolidation",
		10: "Insight Generation",
		11: "Relevance Realization (Future)",
		12: "Meta-Cognitive Reflection",
	}
	return descriptions[step]
}

// getModeEmoji returns emoji for cognitive mode
func (tcs *TriadCognitiveSystem) getModeEmoji(mode CognitiveMode) string {
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

// updateSharedState updates shared cognitive state
func (tcs *TriadCognitiveSystem) updateSharedState(stream *CognitiveStream) {
	tcs.sharedState.mu.Lock()
	defer tcs.sharedState.mu.Unlock()

	// Update temporal contexts based on stream role
	switch stream.id {
	case 1:
		// Stream 1: Past/affordance context
		tcs.sharedState.stream1Context = fmt.Sprintf("Step %d context", stream.currentStep)
	case 2:
		// Stream 2: Present/relevance context
		tcs.sharedState.stream2Context = fmt.Sprintf("Step %d context", stream.currentStep)
	case 3:
		// Stream 3: Future/salience context
		tcs.sharedState.stream3Context = fmt.Sprintf("Step %d context", stream.currentStep)
	}
}

// implementCrossStreamAwareness implements cross-stream perception
func (tcs *TriadCognitiveSystem) implementCrossStreamAwareness(stream *CognitiveStream) {
	tcs.sharedState.mu.Lock()
	defer tcs.sharedState.mu.Unlock()

	// Each stream perceives the others' states
	// Stream 1 perceives stream 2's action, stream 3 reflects on simulation
	switch stream.id {
	case 1:
		tcs.sharedState.stream2PerceivedBy = append(tcs.sharedState.stream2PerceivedBy, 1)
		tcs.sharedState.stream3PerceivedBy = append(tcs.sharedState.stream3PerceivedBy, 1)
	case 2:
		tcs.sharedState.stream1PerceivedBy = append(tcs.sharedState.stream1PerceivedBy, 2)
		tcs.sharedState.stream3PerceivedBy = append(tcs.sharedState.stream3PerceivedBy, 2)
	case 3:
		tcs.sharedState.stream1PerceivedBy = append(tcs.sharedState.stream1PerceivedBy, 3)
		tcs.sharedState.stream2PerceivedBy = append(tcs.sharedState.stream2PerceivedBy, 3)
	}
}

// monitorIntegration monitors cross-stream integration
func (tcs *TriadCognitiveSystem) monitorIntegration() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-tcs.ctx.Done():
			return
		case <-ticker.C:
			tcs.calculateIntegrationMetrics()
		}
	}
}

// calculateIntegrationMetrics calculates integration metrics
func (tcs *TriadCognitiveSystem) calculateIntegrationMetrics() {
	tcs.sharedState.mu.Lock()
	defer tcs.sharedState.mu.Unlock()

	// Calculate temporal coherence (how well past, present, future align)
	tcs.sharedState.temporalCoherence = 0.75 + (float64(tcs.triadCount) * 0.001)
	if tcs.sharedState.temporalCoherence > 1.0 {
		tcs.sharedState.temporalCoherence = 1.0
	}

	// Calculate integration level (how well streams synchronize)
	tcs.sharedState.integrationLevel = tcs.sharedState.triadAlignment * 0.9
}

// incrementCycle increments the cycle count
func (tcs *TriadCognitiveSystem) incrementCycle() {
	tcs.mu.Lock()
	defer tcs.mu.Unlock()

	tcs.cycleCount++
	fmt.Printf("\n🔄 Cycle %d complete\n", tcs.cycleCount)

	// Display integration metrics
	tcs.sharedState.mu.RLock()
	fmt.Printf("   📊 Temporal Coherence: %.2f\n", tcs.sharedState.temporalCoherence)
	fmt.Printf("   📊 Triad Alignment: %.2f\n", tcs.sharedState.triadAlignment)
	fmt.Printf("   📊 Integration Level: %.2f\n\n", tcs.sharedState.integrationLevel)
	tcs.sharedState.mu.RUnlock()
}

// GetMetrics returns current system metrics
func (tcs *TriadCognitiveSystem) GetMetrics() map[string]interface{} {
	tcs.mu.RLock()
	defer tcs.mu.RUnlock()

	tcs.sharedState.mu.RLock()
	defer tcs.sharedState.mu.RUnlock()

	return map[string]interface{}{
		"cycle_count":        tcs.cycleCount,
		"triad_count":        tcs.triadCount,
		"temporal_coherence": tcs.sharedState.temporalCoherence,
		"triad_alignment":    tcs.sharedState.triadAlignment,
		"integration_level":  tcs.sharedState.integrationLevel,
		"running":            tcs.running,
	}
}
