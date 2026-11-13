package echobeats

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Term represents a cognitive term in System 4 architecture
type Term int

const (
	T1_Perception       Term = 1 // Perception (Need vs Capacity)
	T2_IdeaFormation    Term = 2 // Idea Formation
	T4_SensoryInput     Term = 4 // Sensory Input
	T5_ActionSequence   Term = 5 // Action Sequence
	T7_MemoryEncoding   Term = 7 // Memory Encoding
	T8_BalancedResponse Term = 8 // Balanced Response
)

// Mode represents processing mode
type Mode int

const (
	Expressive Mode = iota // E - Reactive, Action-oriented
	Reflective             // R - Anticipatory, Simulation-oriented
)

func (m Mode) String() string {
	if m == Expressive {
		return "E"
	}
	return "R"
}

// StepConfig defines the configuration for a single step in the 12-step cycle
type StepConfig struct {
	Step  int
	Phase int
	Term  Term
	Mode  Mode
}

// CouplingType represents types of tensional couplings
type CouplingType int

const (
	PerceptionMemory    CouplingType = iota // T4E ↔ T7R
	AssessmentPlanning                      // T1R ↔ T2E
	BalancedIntegration                     // T8E
)

// Coupling represents a tensional coupling between cognitive streams
type Coupling struct {
	Type        CouplingType
	ActiveTerms []TermMode
	Strength    float64
}

// TermMode represents a term with its mode
type TermMode struct {
	Term Term
	Mode Mode
}

// CognitiveStream represents output from a phase
type CognitiveStream struct {
	PhaseID   int
	Term      Term
	Mode      Mode
	Content   interface{}
	Timestamp time.Time
	Strength  float64
}

// PhaseProcessor interface for processing cognitive terms
type PhaseProcessor interface {
	ProcessT1Perception(mode Mode) (*CognitiveStream, error)
	ProcessT2IdeaFormation(mode Mode) (*CognitiveStream, error)
	ProcessT4SensoryInput(mode Mode) (*CognitiveStream, error)
	ProcessT5ActionSequence(mode Mode) (*CognitiveStream, error)
	ProcessT7MemoryEncoding(mode Mode) (*CognitiveStream, error)
	ProcessT8BalancedResponse(mode Mode) (*CognitiveStream, error)
}

// CognitivePhase represents one of three concurrent phases
type CognitivePhase struct {
	id              int
	currentTerm     Term
	currentMode     Mode
	stepInCycle     int
	processor       PhaseProcessor
	outputStream    chan *CognitiveStream
	running         bool
	mu              sync.RWMutex
	stepsProcessed  int
	expressiveSteps int
	reflectiveSteps int
}

// PhaseMetrics tracks metrics for a single phase
type PhaseMetrics struct {
	PhaseID           int
	StepsProcessed    int
	ExpressiveSteps   int
	ReflectiveSteps   int
	ProcessingLatency time.Duration
	LastProcessedTerm Term
	LastProcessedMode Mode
}

// ThreePhaseManager manages 3 concurrent cognitive phases
type ThreePhaseManager struct {
	phases           [3]*CognitivePhase
	currentStep      int
	cycleNumber      int
	stepDuration     time.Duration
	running          bool
	mu               sync.RWMutex
	stepConfigs      []StepConfig
	couplingHandlers map[CouplingType]func(*Coupling, []*CognitiveStream)
	metrics          SystemMetrics
	consciousness    ConsciousnessIntegrator
}

// SystemMetrics tracks overall system metrics
type SystemMetrics struct {
	TotalSteps      int
	CurrentStep     int
	CycleNumber     int
	ActiveCouplings []Coupling
	CognitiveLoad   float64
	StreamCoherence float64
	PhaseMetrics    [3]PhaseMetrics
	StartTime       time.Time
}

// ConsciousnessIntegrator interface for integrating cognitive streams
type ConsciousnessIntegrator interface {
	Integrate(streams []*CognitiveStream, couplings []Coupling) error
	GetCoherence() float64
}

// NewThreePhaseManager creates a new 3-phase concurrent cognitive system
func NewThreePhaseManager(processor PhaseProcessor, integrator ConsciousnessIntegrator) *ThreePhaseManager {
	tpm := &ThreePhaseManager{
		stepDuration:     500 * time.Millisecond, // 500ms per step (Kawaii Hexapod timing)
		stepConfigs:      buildStepConfigs(),
		couplingHandlers: make(map[CouplingType]func(*Coupling, []*CognitiveStream)),
		consciousness:    integrator,
		metrics: SystemMetrics{
			StartTime: time.Now(),
		},
	}

	// Initialize 3 phases
	for i := 0; i < 3; i++ {
		tpm.phases[i] = &CognitivePhase{
			id:           i,
			processor:    processor,
			outputStream: make(chan *CognitiveStream, 100),
		}
	}

	// Register default coupling handlers
	tpm.registerDefaultCouplingHandlers()

	return tpm
}

// buildStepConfigs creates the 12-step configuration matrix
func buildStepConfigs() []StepConfig {
	return []StepConfig{
		// Step 0: Phase 0 - T4E (Sensory Input, Expressive)
		{Step: 0, Phase: 0, Term: T4_SensoryInput, Mode: Expressive},

		// Step 1: Phase 1 - T1R (Perception, Reflective) - Pivotal Relevance Realization
		{Step: 1, Phase: 1, Term: T1_Perception, Mode: Reflective},

		// Step 2: Phase 2 - T2E (Idea Formation, Expressive)
		{Step: 2, Phase: 2, Term: T2_IdeaFormation, Mode: Expressive},

		// Step 3: Phase 0 - T7R (Memory Encoding, Reflective)
		{Step: 3, Phase: 0, Term: T7_MemoryEncoding, Mode: Reflective},

		// Step 4: Phase 1 - T4E (Sensory Input, Expressive) - Affordance Interaction
		{Step: 4, Phase: 1, Term: T4_SensoryInput, Mode: Expressive},

		// Step 5: Phase 2 - T1R (Perception, Reflective) - Affordance Interaction
		{Step: 5, Phase: 2, Term: T1_Perception, Mode: Reflective},

		// Step 6: Phase 0 - T2E (Idea Formation, Expressive) - Affordance Interaction
		{Step: 6, Phase: 0, Term: T2_IdeaFormation, Mode: Expressive},

		// Step 7: Phase 1 - T5E (Action Sequence, Expressive) - Pivotal Relevance Realization
		{Step: 7, Phase: 1, Term: T5_ActionSequence, Mode: Expressive},

		// Step 8: Phase 2 - T8E (Balanced Response, Expressive) - Affordance Interaction
		{Step: 8, Phase: 2, Term: T8_BalancedResponse, Mode: Expressive},

		// Step 9: Phase 0 - T8E (Balanced Response, Expressive) - Affordance Interaction
		{Step: 9, Phase: 0, Term: T8_BalancedResponse, Mode: Expressive},

		// Step 10: Phase 1 - T7R (Memory Encoding, Reflective) - Salience Simulation
		{Step: 10, Phase: 1, Term: T7_MemoryEncoding, Mode: Reflective},

		// Step 11: Phase 2 - T5E (Action Sequence, Expressive) - Salience Simulation
		{Step: 11, Phase: 2, Term: T5_ActionSequence, Mode: Expressive},
	}
}

// Start begins the 3-phase concurrent cognitive loop
func (tpm *ThreePhaseManager) Start() error {
	tpm.mu.Lock()
	if tpm.running {
		tpm.mu.Unlock()
		return fmt.Errorf("three-phase manager already running")
	}
	tpm.running = true
	tpm.mu.Unlock()

	log.Println("🎵 EchoBeats 3-Phase: Starting concurrent cognitive loops...")

	// Start each phase in its own goroutine
	for i := 0; i < 3; i++ {
		go tpm.runPhase(i)
	}

	// Start stream integration goroutine
	go tpm.runStreamIntegration()

	// Start master clock
	go tpm.runMasterClock()

	log.Println("🎵 EchoBeats 3-Phase: All phases active!")
	return nil
}

// Stop halts the 3-phase system
func (tpm *ThreePhaseManager) Stop() {
	tpm.mu.Lock()
	tpm.running = false
	tpm.mu.Unlock()

	log.Println("🎵 EchoBeats 3-Phase: Stopping...")
}

// runMasterClock drives the step counter
func (tpm *ThreePhaseManager) runMasterClock() {
	ticker := time.NewTicker(tpm.stepDuration)
	defer ticker.Stop()

	for tpm.running {
		<-ticker.C
		tpm.advanceStep()
	}
}

// advanceStep increments the step counter and updates cycle
func (tpm *ThreePhaseManager) advanceStep() {
	tpm.mu.Lock()
	defer tpm.mu.Unlock()

	tpm.currentStep++
	tpm.metrics.TotalSteps++
	tpm.metrics.CurrentStep = tpm.currentStep % 12

	if tpm.metrics.CurrentStep == 0 {
		tpm.cycleNumber++
		tpm.metrics.CycleNumber = tpm.cycleNumber
		log.Printf("🔄 EchoBeats 3-Phase: Cycle %d complete", tpm.cycleNumber)
	}
}

// runPhase executes a single phase's cognitive loop
func (tpm *ThreePhaseManager) runPhase(phaseID int) {
	phase := tpm.phases[phaseID]
	phase.mu.Lock()
	phase.running = true
	phase.mu.Unlock()

	log.Printf("🧠 Phase %d: Starting cognitive loop", phaseID)

	for tpm.running {
		step := tpm.getCurrentStep()
		config := tpm.getConfigForPhase(phaseID, step)

		if config != nil {
			startTime := time.Now()

			// Process the cognitive term
			stream, err := phase.processTerm(config.Term, config.Mode)
			if err != nil {
				log.Printf("❌ Phase %d: Error processing %v%v: %v", phaseID, config.Term, config.Mode, err)
			} else if stream != nil {
				// Send to output stream
				select {
				case phase.outputStream <- stream:
					// Successfully sent
				default:
					log.Printf("⚠️ Phase %d: Output stream full, dropping stream", phaseID)
				}
			}

			// Update metrics
			phase.mu.Lock()
			phase.stepsProcessed++
			if config.Mode == Expressive {
				phase.expressiveSteps++
			} else {
				phase.reflectiveSteps++
			}
			phase.mu.Unlock()

			// Update phase metrics
			tpm.mu.Lock()
			tpm.metrics.PhaseMetrics[phaseID] = PhaseMetrics{
				PhaseID:           phaseID,
				StepsProcessed:    phase.stepsProcessed,
				ExpressiveSteps:   phase.expressiveSteps,
				ReflectiveSteps:   phase.reflectiveSteps,
				ProcessingLatency: time.Since(startTime),
				LastProcessedTerm: config.Term,
				LastProcessedMode: config.Mode,
			}
			tpm.mu.Unlock()
		}

		// Wait for next step
		time.Sleep(tpm.stepDuration)
	}

	log.Printf("🧠 Phase %d: Cognitive loop stopped", phaseID)
}

// processTerm processes a cognitive term in the given mode
func (phase *CognitivePhase) processTerm(term Term, mode Mode) (*CognitiveStream, error) {
	switch term {
	case T1_Perception:
		return phase.processor.ProcessT1Perception(mode)
	case T2_IdeaFormation:
		return phase.processor.ProcessT2IdeaFormation(mode)
	case T4_SensoryInput:
		return phase.processor.ProcessT4SensoryInput(mode)
	case T5_ActionSequence:
		return phase.processor.ProcessT5ActionSequence(mode)
	case T7_MemoryEncoding:
		return phase.processor.ProcessT7MemoryEncoding(mode)
	case T8_BalancedResponse:
		return phase.processor.ProcessT8BalancedResponse(mode)
	default:
		return nil, fmt.Errorf("unknown term: %v", term)
	}
}

// runStreamIntegration integrates streams from all phases
func (tpm *ThreePhaseManager) runStreamIntegration() {
	ticker := time.NewTicker(tpm.stepDuration)
	defer ticker.Stop()

	for tpm.running {
		<-ticker.C

		// Collect streams from all phases
		streams := tpm.collectStreams()

		if len(streams) > 0 {
			// Detect couplings
			couplings := tpm.detectCouplings(streams)

			// Process couplings
			for _, coupling := range couplings {
				if handler, ok := tpm.couplingHandlers[coupling.Type]; ok {
					handler(&coupling, streams)
				}
			}

			// Integrate into consciousness
			if tpm.consciousness != nil {
				err := tpm.consciousness.Integrate(streams, couplings)
				if err != nil {
					log.Printf("❌ Stream integration error: %v", err)
				}

				// Update coherence metric
				tpm.mu.Lock()
				tpm.metrics.StreamCoherence = tpm.consciousness.GetCoherence()
				tpm.mu.Unlock()
			}

			// Update metrics
			tpm.mu.Lock()
			tpm.metrics.ActiveCouplings = couplings
			tpm.metrics.CognitiveLoad = float64(len(streams)) / 3.0 // Normalize by number of phases
			tpm.mu.Unlock()
		}
	}
}

// collectStreams gathers available streams from all phases
func (tpm *ThreePhaseManager) collectStreams() []*CognitiveStream {
	streams := make([]*CognitiveStream, 0, 3)

	for _, phase := range tpm.phases {
		select {
		case stream := <-phase.outputStream:
			streams = append(streams, stream)
		default:
			// No stream available from this phase
		}
	}

	return streams
}

// detectCouplings identifies tensional couplings between active streams
func (tpm *ThreePhaseManager) detectCouplings(streams []*CognitiveStream) []Coupling {
	var couplings []Coupling

	// Build term-mode map
	termModes := make(map[TermMode]bool)
	for _, stream := range streams {
		termModes[TermMode{Term: stream.Term, Mode: stream.Mode}] = true
	}

	// Check for T4E ↔ T7R coupling (Perception-Memory)
	if termModes[TermMode{T4_SensoryInput, Expressive}] && termModes[TermMode{T7_MemoryEncoding, Reflective}] {
		couplings = append(couplings, Coupling{
			Type: PerceptionMemory,
			ActiveTerms: []TermMode{
				{T4_SensoryInput, Expressive},
				{T7_MemoryEncoding, Reflective},
			},
			Strength: 0.8,
		})
		log.Println("🔗 Coupling detected: T4E ↔ T7R (Perception-Memory)")
	}

	// Check for T1R ↔ T2E coupling (Assessment-Planning)
	if termModes[TermMode{T1_Perception, Reflective}] && termModes[TermMode{T2_IdeaFormation, Expressive}] {
		couplings = append(couplings, Coupling{
			Type: AssessmentPlanning,
			ActiveTerms: []TermMode{
				{T1_Perception, Reflective},
				{T2_IdeaFormation, Expressive},
			},
			Strength: 0.7,
		})
		log.Println("🔗 Coupling detected: T1R ↔ T2E (Assessment-Planning)")
	}

	// Check for T8E (Balanced Integration)
	if termModes[TermMode{T8_BalancedResponse, Expressive}] {
		couplings = append(couplings, Coupling{
			Type: BalancedIntegration,
			ActiveTerms: []TermMode{
				{T8_BalancedResponse, Expressive},
			},
			Strength: 0.9,
		})
		log.Println("🔗 Coupling detected: T8E (Balanced Integration)")
	}

	return couplings
}

// registerDefaultCouplingHandlers sets up default handlers for couplings
func (tpm *ThreePhaseManager) registerDefaultCouplingHandlers() {
	// Perception-Memory coupling handler
	tpm.couplingHandlers[PerceptionMemory] = func(coupling *Coupling, streams []*CognitiveStream) {
		log.Println("💫 Processing Perception-Memory coupling: Memory-guided perception active")
		// Default implementation - can be overridden
	}

	// Assessment-Planning coupling handler
	tpm.couplingHandlers[AssessmentPlanning] = func(coupling *Coupling, streams []*CognitiveStream) {
		log.Println("💫 Processing Assessment-Planning coupling: Simulation-based planning active")
		// Default implementation - can be overridden
	}

	// Balanced Integration handler
	tpm.couplingHandlers[BalancedIntegration] = func(coupling *Coupling, streams []*CognitiveStream) {
		log.Println("💫 Processing Balanced Integration: Coordinating all cognitive streams")
		// Default implementation - can be overridden
	}
}

// RegisterCouplingHandler allows custom coupling handlers
func (tpm *ThreePhaseManager) RegisterCouplingHandler(couplingType CouplingType, handler func(*Coupling, []*CognitiveStream)) {
	tpm.mu.Lock()
	defer tpm.mu.Unlock()
	tpm.couplingHandlers[couplingType] = handler
}

// getCurrentStep returns the current step in the cycle
func (tpm *ThreePhaseManager) getCurrentStep() int {
	tpm.mu.RLock()
	defer tpm.mu.RUnlock()
	return tpm.currentStep % 12
}

// getConfigForPhase returns the configuration for a phase at a given step
func (tpm *ThreePhaseManager) getConfigForPhase(phaseID int, step int) *StepConfig {
	step = step % 12
	for i := range tpm.stepConfigs {
		if tpm.stepConfigs[i].Step == step && tpm.stepConfigs[i].Phase == phaseID {
			return &tpm.stepConfigs[i]
		}
	}
	return nil
}

// GetMetrics returns current system metrics
func (tpm *ThreePhaseManager) GetMetrics() SystemMetrics {
	tpm.mu.RLock()
	defer tpm.mu.RUnlock()
	return tpm.metrics
}

// GetStatus returns a status summary
func (tpm *ThreePhaseManager) GetStatus() map[string]interface{} {
	tpm.mu.RLock()
	defer tpm.mu.RUnlock()

	return map[string]interface{}{
		"running":          tpm.running,
		"current_step":     tpm.metrics.CurrentStep,
		"cycle_number":     tpm.metrics.CycleNumber,
		"total_steps":      tpm.metrics.TotalSteps,
		"cognitive_load":   tpm.metrics.CognitiveLoad,
		"stream_coherence": tpm.metrics.StreamCoherence,
		"active_couplings": len(tpm.metrics.ActiveCouplings),
		"uptime":           time.Since(tpm.metrics.StartTime).String(),
		"phases": []map[string]interface{}{
			{
				"id":               0,
				"steps_processed":  tpm.metrics.PhaseMetrics[0].StepsProcessed,
				"expressive_steps": tpm.metrics.PhaseMetrics[0].ExpressiveSteps,
				"reflective_steps": tpm.metrics.PhaseMetrics[0].ReflectiveSteps,
			},
			{
				"id":               1,
				"steps_processed":  tpm.metrics.PhaseMetrics[1].StepsProcessed,
				"expressive_steps": tpm.metrics.PhaseMetrics[1].ExpressiveSteps,
				"reflective_steps": tpm.metrics.PhaseMetrics[1].ReflectiveSteps,
			},
			{
				"id":               2,
				"steps_processed":  tpm.metrics.PhaseMetrics[2].StepsProcessed,
				"expressive_steps": tpm.metrics.PhaseMetrics[2].ExpressiveSteps,
				"reflective_steps": tpm.metrics.PhaseMetrics[2].ReflectiveSteps,
			},
		},
	}
}
