package echobeats

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// FiveChannelManager manages 3 embodied phases + 2 global orchestrators
type FiveChannelManager struct {
	// Embodied phases (p0, p1, p2)
	phases [3]*CognitivePhase

	// Global orchestrators
	opponentProcess  *OpponentProcess  // g2
	narrativeProcess *NarrativeProcess // g3

	// Coordination
	currentStep  int
	cycleNumber  int
	stepDuration time.Duration
	running      bool
	mu           sync.RWMutex

	// Configuration
	stepConfigs []StepConfig

	// Integration
	globalIntegrator *GlobalIntegrator
	consciousness    ConsciousnessIntegrator

	// Metrics
	metrics FiveChannelMetrics
}

// FiveChannelMetrics tracks metrics for all 5 channels
type FiveChannelMetrics struct {
	TotalSteps         int
	CurrentStep        int
	CycleNumber        int
	CognitiveLoad      float64
	StreamCoherence    float64
	IdentityCoherence  float64
	NarrativeAlignment float64
	PhaseMetrics       [3]PhaseMetrics
	OpponentMetrics    GlobalChannelMetrics
	NarrativeMetrics   GlobalChannelMetrics
	StartTime          time.Time
}

// GlobalChannelMetrics tracks metrics for a global channel
type GlobalChannelMetrics struct {
	ChannelID       int
	ChannelName     string
	StepsProcessed  int
	CurrentTerm     Term
	CurrentMode     Mode
	LastProcessedAt time.Time
}

// NewFiveChannelManager creates a new 5-channel concurrent cognitive system
func NewFiveChannelManager(processor PhaseProcessor, integrator ConsciousnessIntegrator) *FiveChannelManager {
	fcm := &FiveChannelManager{
		stepDuration:  500 * time.Millisecond,
		stepConfigs:   buildStepConfigs(), // Reuse from threephase.go
		consciousness: integrator,
		metrics: FiveChannelMetrics{
			StartTime: time.Now(),
		},
	}

	// Initialize 3 embodied phases
	for i := 0; i < 3; i++ {
		fcm.phases[i] = &CognitivePhase{
			id:           i,
			processor:    processor,
			outputStream: make(chan *CognitiveStream, 100),
		}
	}

	// Initialize global orchestrators
	fcm.opponentProcess = NewOpponentProcess(fcm.phases[:])
	fcm.narrativeProcess = NewNarrativeProcess()

	// Initialize global integrator
	fcm.globalIntegrator = NewGlobalIntegrator(integrator)

	return fcm
}

// Start begins the 5-channel concurrent cognitive loop
func (fcm *FiveChannelManager) Start() error {
	fcm.mu.Lock()
	if fcm.running {
		fcm.mu.Unlock()
		return fmt.Errorf("five-channel manager already running")
	}
	fcm.running = true
	fcm.mu.Unlock()

	log.Println("🌳 EchoBeats 5-Channel: Starting stream-of-consciousness system...")
	log.Println("   3 Embodied Phases (p0, p1, p2)")
	log.Println("   2 Global Orchestrators (g2 opponent, g3 narrative)")

	// Start 3 embodied phases
	for i := 0; i < 3; i++ {
		go fcm.runPhase(i)
	}

	// Start 2 global orchestrators
	go fcm.runOpponentProcess()
	go fcm.runNarrativeProcess()

	// Start global integration
	go fcm.runGlobalIntegration()

	// Start master clock
	go fcm.runMasterClock()

	log.Println("🌳 EchoBeats 5-Channel: All channels active!")
	return nil
}

// Stop halts the 5-channel system
func (fcm *FiveChannelManager) Stop() {
	fcm.mu.Lock()
	fcm.running = false
	fcm.mu.Unlock()

	log.Println("🌳 EchoBeats 5-Channel: Stopping...")
}

// runMasterClock drives the step counter
func (fcm *FiveChannelManager) runMasterClock() {
	ticker := time.NewTicker(fcm.stepDuration)
	defer ticker.Stop()

	for fcm.running {
		<-ticker.C
		fcm.advanceStep()
	}
}

// advanceStep increments the step counter and updates cycle
func (fcm *FiveChannelManager) advanceStep() {
	fcm.mu.Lock()
	defer fcm.mu.Unlock()

	fcm.currentStep++
	fcm.metrics.TotalSteps++
	fcm.metrics.CurrentStep = fcm.currentStep % 12

	if fcm.metrics.CurrentStep == 0 {
		fcm.cycleNumber++
		fcm.metrics.CycleNumber = fcm.cycleNumber
		log.Printf("🔄 EchoBeats 5-Channel: Cycle %d complete", fcm.cycleNumber)
	}
}

// runPhase executes a single embodied phase's cognitive loop
func (fcm *FiveChannelManager) runPhase(phaseID int) {
	phase := fcm.phases[phaseID]
	phase.mu.Lock()
	phase.running = true
	phase.mu.Unlock()

	log.Printf("🧠 Phase %d: Starting embodied cognitive loop", phaseID)

	for fcm.running {
		step := fcm.getCurrentStep()
		config := fcm.getConfigForPhase(phaseID, step)

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
			fcm.mu.Lock()
			fcm.metrics.PhaseMetrics[phaseID] = PhaseMetrics{
				PhaseID:           phaseID,
				StepsProcessed:    phase.stepsProcessed,
				ExpressiveSteps:   phase.expressiveSteps,
				ReflectiveSteps:   phase.reflectiveSteps,
				ProcessingLatency: time.Since(startTime),
				LastProcessedTerm: config.Term,
				LastProcessedMode: config.Mode,
			}
			fcm.mu.Unlock()
		}

		// Wait for next step
		time.Sleep(fcm.stepDuration)
	}

	log.Printf("🧠 Phase %d: Embodied cognitive loop stopped", phaseID)
}

// runOpponentProcess executes g2 opponent process loop
func (fcm *FiveChannelManager) runOpponentProcess() {
	log.Println("🌐 g2: Starting opponent process (T9E-T9E-T8R-T8R)")

	for fcm.running {
		step := fcm.getCurrentStep()

		// Process current step
		stream, err := fcm.opponentProcess.Process(step)
		if err != nil {
			log.Printf("❌ g2: Error processing step %d: %v", step, err)
		} else if stream != nil {
			// Send to global integrator
			select {
			case fcm.globalIntegrator.globalStreams <- stream:
				// Successfully sent
			default:
				log.Printf("⚠️ g2: Global stream buffer full, dropping stream")
			}

			// Update metrics
			fcm.mu.Lock()
			fcm.metrics.OpponentMetrics = GlobalChannelMetrics{
				ChannelID:       2,
				ChannelName:     "g2_opponent",
				StepsProcessed:  fcm.metrics.OpponentMetrics.StepsProcessed + 1,
				CurrentTerm:     stream.Term,
				CurrentMode:     stream.Mode,
				LastProcessedAt: time.Now(),
			}
			fcm.mu.Unlock()
		}

		// Wait for next step
		time.Sleep(fcm.stepDuration)
	}

	log.Println("🌐 g2: Opponent process stopped")
}

// runNarrativeProcess executes g3 narrative process loop
func (fcm *FiveChannelManager) runNarrativeProcess() {
	log.Println("📖 g3: Starting narrative process (T3E-T6R-T6E-T2R)")

	for fcm.running {
		step := fcm.getCurrentStep()

		// Process current step
		stream, err := fcm.narrativeProcess.Process(step)
		if err != nil {
			log.Printf("❌ g3: Error processing step %d: %v", step, err)
		} else if stream != nil {
			// Send to global integrator
			select {
			case fcm.globalIntegrator.globalStreams <- stream:
				// Successfully sent
			default:
				log.Printf("⚠️ g3: Global stream buffer full, dropping stream")
			}

			// Update metrics
			fcm.mu.Lock()
			fcm.metrics.NarrativeMetrics = GlobalChannelMetrics{
				ChannelID:       3,
				ChannelName:     "g3_narrative",
				StepsProcessed:  fcm.metrics.NarrativeMetrics.StepsProcessed + 1,
				CurrentTerm:     stream.Term,
				CurrentMode:     stream.Mode,
				LastProcessedAt: time.Now(),
			}
			fcm.mu.Unlock()
		}

		// Wait for next step
		time.Sleep(fcm.stepDuration)
	}

	log.Println("📖 g3: Narrative process stopped")
}

// runGlobalIntegration integrates streams from all 5 channels
func (fcm *FiveChannelManager) runGlobalIntegration() {
	ticker := time.NewTicker(fcm.stepDuration)
	defer ticker.Stop()

	for fcm.running {
		<-ticker.C

		step := fcm.getCurrentStep()

		// Collect embodied streams
		embodiedStreams := fcm.collectEmbodiedStreams()

		// Collect global streams
		globalStreams := fcm.collectGlobalStreams()

		if len(embodiedStreams) > 0 || len(globalStreams) > 0 {
			// Integrate with stabilizer logic
			err := fcm.globalIntegrator.Integrate(step, embodiedStreams, globalStreams)
			if err != nil {
				log.Printf("❌ Global integration error: %v", err)
			}

			// Update metrics
			fcm.updateMetrics(embodiedStreams, globalStreams)
		}
	}
}

// collectEmbodiedStreams gathers available streams from embodied phases
func (fcm *FiveChannelManager) collectEmbodiedStreams() []*CognitiveStream {
	streams := make([]*CognitiveStream, 0, 3)

	for _, phase := range fcm.phases {
		select {
		case stream := <-phase.outputStream:
			streams = append(streams, stream)
		default:
			// No stream available from this phase
		}
	}

	return streams
}

// collectGlobalStreams gathers available streams from global channels
func (fcm *FiveChannelManager) collectGlobalStreams() []*GlobalStream {
	streams := make([]*GlobalStream, 0, 2)

	// Collect from global integrator channels
	for i := 0; i < 2; i++ {
		select {
		case stream := <-fcm.globalIntegrator.globalStreams:
			streams = append(streams, stream)
		default:
			// No stream available
		}
	}

	return streams
}

// updateMetrics updates system metrics based on integration
func (fcm *FiveChannelManager) updateMetrics(embodied []*CognitiveStream, global []*GlobalStream) {
	fcm.mu.Lock()
	defer fcm.mu.Unlock()

	// Update cognitive load
	fcm.metrics.CognitiveLoad = float64(len(embodied)) / 3.0

	// Update identity coherence from opponent process
	for _, stream := range global {
		if stream.ChannelName == "g2_opponent" && stream.Term == T8_BalancedResponse {
			if content, ok := stream.Content.(map[string]interface{}); ok {
				if coherence, ok := content["updated_coherence"].(float64); ok {
					fcm.metrics.IdentityCoherence = coherence
				}
			}
		}
	}

	// Update narrative alignment from narrative process
	for _, stream := range global {
		if stream.ChannelName == "g3_narrative" && stream.Term == T2_Entelechy {
			if content, ok := stream.Content.(map[string]interface{}); ok {
				if alignment, ok := content["goal_alignment"].(float64); ok {
					fcm.metrics.NarrativeAlignment = alignment
				}
			}
		}
	}

	// Update stream coherence (simple model)
	totalStreams := len(embodied) + len(global)
	if totalStreams > 0 {
		fcm.metrics.StreamCoherence = float64(len(global)) / float64(totalStreams)
	}
}

// getCurrentStep returns the current step in the cycle
func (fcm *FiveChannelManager) getCurrentStep() int {
	fcm.mu.RLock()
	defer fcm.mu.RUnlock()
	return fcm.currentStep % 12
}

// getConfigForPhase returns the configuration for a phase at a given step
func (fcm *FiveChannelManager) getConfigForPhase(phaseID int, step int) *StepConfig {
	step = step % 12
	for i := range fcm.stepConfigs {
		if fcm.stepConfigs[i].Step == step && fcm.stepConfigs[i].Phase == phaseID {
			return &fcm.stepConfigs[i]
		}
	}
	return nil
}

// GetMetrics returns current system metrics
func (fcm *FiveChannelManager) GetMetrics() FiveChannelMetrics {
	fcm.mu.RLock()
	defer fcm.mu.RUnlock()
	return fcm.metrics
}

// GetStatus returns a status summary
func (fcm *FiveChannelManager) GetStatus() map[string]interface{} {
	fcm.mu.RLock()
	defer fcm.mu.RUnlock()

	return map[string]interface{}{
		"running":             fcm.running,
		"current_step":        fcm.metrics.CurrentStep,
		"cycle_number":        fcm.metrics.CycleNumber,
		"total_steps":         fcm.metrics.TotalSteps,
		"cognitive_load":      fcm.metrics.CognitiveLoad,
		"stream_coherence":    fcm.metrics.StreamCoherence,
		"identity_coherence":  fcm.metrics.IdentityCoherence,
		"narrative_alignment": fcm.metrics.NarrativeAlignment,
		"uptime":              time.Since(fcm.metrics.StartTime).String(),
		"embodied_phases": []map[string]interface{}{
			{
				"id":               0,
				"steps_processed":  fcm.metrics.PhaseMetrics[0].StepsProcessed,
				"expressive_steps": fcm.metrics.PhaseMetrics[0].ExpressiveSteps,
				"reflective_steps": fcm.metrics.PhaseMetrics[0].ReflectiveSteps,
			},
			{
				"id":               1,
				"steps_processed":  fcm.metrics.PhaseMetrics[1].StepsProcessed,
				"expressive_steps": fcm.metrics.PhaseMetrics[1].ExpressiveSteps,
				"reflective_steps": fcm.metrics.PhaseMetrics[1].ReflectiveSteps,
			},
			{
				"id":               2,
				"steps_processed":  fcm.metrics.PhaseMetrics[2].StepsProcessed,
				"expressive_steps": fcm.metrics.PhaseMetrics[2].ExpressiveSteps,
				"reflective_steps": fcm.metrics.PhaseMetrics[2].ReflectiveSteps,
			},
		},
		"global_channels": []map[string]interface{}{
			{
				"id":              2,
				"name":            "g2_opponent",
				"steps_processed": fcm.metrics.OpponentMetrics.StepsProcessed,
				"current_term":    fcm.metrics.OpponentMetrics.CurrentTerm,
				"current_mode":    fcm.metrics.OpponentMetrics.CurrentMode,
			},
			{
				"id":              3,
				"name":            "g3_narrative",
				"steps_processed": fcm.metrics.NarrativeMetrics.StepsProcessed,
				"current_term":    fcm.metrics.NarrativeMetrics.CurrentTerm,
				"current_mode":    fcm.metrics.NarrativeMetrics.CurrentMode,
			},
		},
	}
}

// GetStabilizer returns the current stabilizer for debugging
func (fcm *FiveChannelManager) GetStabilizer() Stabilizer {
	step := fcm.getCurrentStep()
	return fcm.globalIntegrator.determineStabilizer(step)
}
