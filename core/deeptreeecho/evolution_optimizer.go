package deeptreeecho

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// EvolutionOptimizer orchestrates the evolutionary optimization of Deep Tree Echo
// It integrates stream-of-consciousness, echobeats, and echodream into a unified
// cognitive evolution system
type EvolutionOptimizer struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Core subsystems
	consciousness    *StreamOfConsciousness
	scheduler        *EchobeatsScheduler
	dreamIntegration *EchoDreamKnowledgeIntegration

	// LLM provider
	llmProvider llm.LLMProvider

	// Evolution state
	generation        uint64
	fitnessScore      float64
	evolutionVelocity float64
	adaptationRate    float64

	// Evolutionary metrics
	wisdomDepth        float64
	patternComplexity  float64
	coherenceStability float64
	emergenceIndex     float64

	// Genetic memory - traits that persist across generations
	geneticTraits map[string]GeneticTrait

	// Evolution history
	evolutionHistory []EvolutionSnapshot

	// Fitness landscape
	fitnessLandscape *FitnessLandscape

	// Running state
	running        bool
	evolutionCycle time.Duration
}

// GeneticTrait represents an evolved characteristic
type GeneticTrait struct {
	Name          string    `json:"name"`
	Value         float64   `json:"value"`
	Mutability    float64   `json:"mutability"`
	Expression    float64   `json:"expression"`
	Generation    uint64    `json:"generation"`
	LastMutation  time.Time `json:"last_mutation"`
	FitnessImpact float64   `json:"fitness_impact"`
}

// EvolutionSnapshot captures evolutionary state at a point in time
type EvolutionSnapshot struct {
	Generation        uint64             `json:"generation"`
	Timestamp         time.Time          `json:"timestamp"`
	FitnessScore      float64            `json:"fitness_score"`
	WisdomDepth       float64            `json:"wisdom_depth"`
	PatternComplexity float64            `json:"pattern_complexity"`
	CoherenceScore    float64            `json:"coherence_score"`
	EmergenceIndex    float64            `json:"emergence_index"`
	TraitSnapshot     map[string]float64 `json:"trait_snapshot"`
}

// FitnessLandscape models the evolutionary fitness space
type FitnessLandscape struct {
	Dimensions      int           `json:"dimensions"`
	Peaks           []FitnessPeak `json:"peaks"`
	CurrentPosition []float64     `json:"current_position"`
	Gradient        []float64     `json:"gradient"`
	LocalOptima     bool          `json:"local_optima"`
}

// FitnessPeak represents a fitness peak in the landscape
type FitnessPeak struct {
	Position   []float64 `json:"position"`
	Height     float64   `json:"height"`
	Breadth    float64   `json:"breadth"`
	Discovered time.Time `json:"discovered"`
}

// EvolutionConfig configures the evolution optimizer
type EvolutionConfig struct {
	InitialAdaptationRate float64
	MutationProbability   float64
	SelectionPressure     float64
	ElitismRatio          float64
	DiversityWeight       float64
	EvolutionCycle        time.Duration
}

// DefaultEvolutionConfig returns default evolution configuration
func DefaultEvolutionConfig() EvolutionConfig {
	return EvolutionConfig{
		InitialAdaptationRate: 0.1,
		MutationProbability:   0.15,
		SelectionPressure:     0.7,
		ElitismRatio:          0.2,
		DiversityWeight:       0.3,
		EvolutionCycle:        30 * time.Second,
	}
}

// NewEvolutionOptimizer creates a new evolution optimizer
func NewEvolutionOptimizer(llmProvider llm.LLMProvider, config EvolutionConfig) *EvolutionOptimizer {
	ctx, cancel := context.WithCancel(context.Background())

	eo := &EvolutionOptimizer{
		ctx:                ctx,
		cancel:             cancel,
		llmProvider:        llmProvider,
		generation:         0,
		fitnessScore:       0.5,
		evolutionVelocity:  0.0,
		adaptationRate:     config.InitialAdaptationRate,
		wisdomDepth:        0.0,
		patternComplexity:  0.0,
		coherenceStability: 0.5,
		emergenceIndex:     0.0,
		geneticTraits:      make(map[string]GeneticTrait),
		evolutionHistory:   make([]EvolutionSnapshot, 0),
		evolutionCycle:     config.EvolutionCycle,
	}

	// Initialize core subsystems
	eo.consciousness = NewStreamOfConsciousness(llmProvider)
	eo.scheduler = NewEchobeatsScheduler(llmProvider)
	eo.dreamIntegration = NewEchoDreamKnowledgeIntegration(llmProvider)

	// Initialize genetic traits
	eo.initializeGeneticTraits()

	// Initialize fitness landscape
	eo.initializeFitnessLandscape()

	return eo
}

// initializeGeneticTraits sets up the initial genetic traits
func (eo *EvolutionOptimizer) initializeGeneticTraits() {
	now := time.Now()

	eo.geneticTraits = map[string]GeneticTrait{
		"curiosity_drive": {
			Name:          "Curiosity Drive",
			Value:         0.7,
			Mutability:    0.2,
			Expression:    0.8,
			Generation:    0,
			LastMutation:  now,
			FitnessImpact: 0.15,
		},
		"wisdom_accumulation": {
			Name:          "Wisdom Accumulation",
			Value:         0.5,
			Mutability:    0.1,
			Expression:    0.6,
			Generation:    0,
			LastMutation:  now,
			FitnessImpact: 0.25,
		},
		"pattern_recognition": {
			Name:          "Pattern Recognition",
			Value:         0.6,
			Mutability:    0.15,
			Expression:    0.7,
			Generation:    0,
			LastMutation:  now,
			FitnessImpact: 0.20,
		},
		"coherence_maintenance": {
			Name:          "Coherence Maintenance",
			Value:         0.8,
			Mutability:    0.05,
			Expression:    0.9,
			Generation:    0,
			LastMutation:  now,
			FitnessImpact: 0.20,
		},
		"adaptive_learning": {
			Name:          "Adaptive Learning",
			Value:         0.6,
			Mutability:    0.25,
			Expression:    0.7,
			Generation:    0,
			LastMutation:  now,
			FitnessImpact: 0.20,
		},
	}
}

// initializeFitnessLandscape creates the initial fitness landscape
func (eo *EvolutionOptimizer) initializeFitnessLandscape() {
	eo.fitnessLandscape = &FitnessLandscape{
		Dimensions:      5, // One for each genetic trait
		Peaks:           []FitnessPeak{},
		CurrentPosition: []float64{0.7, 0.5, 0.6, 0.8, 0.6},
		Gradient:        []float64{0.0, 0.0, 0.0, 0.0, 0.0},
		LocalOptima:     false,
	}
}

// Start begins the evolution optimization loop
func (eo *EvolutionOptimizer) Start() error {
	eo.mu.Lock()
	if eo.running {
		eo.mu.Unlock()
		return fmt.Errorf("evolution optimizer already running")
	}
	eo.running = true
	eo.mu.Unlock()

	fmt.Println("🧬 Starting Evolution Optimizer...")
	fmt.Printf("   Initial fitness: %.3f\n", eo.fitnessScore)
	fmt.Printf("   Genetic traits: %d\n", len(eo.geneticTraits))
	fmt.Printf("   Evolution cycle: %v\n", eo.evolutionCycle)

	// Start subsystems
	if err := eo.consciousness.Start(); err != nil {
		fmt.Printf("   Warning: Consciousness start failed: %v\n", err)
	}

	if err := eo.scheduler.Start(); err != nil {
		fmt.Printf("   Warning: Scheduler start failed: %v\n", err)
	}

	// Start evolution loop
	go eo.runEvolutionLoop()

	// Start integration monitor
	go eo.runIntegrationMonitor()

	return nil
}

// Stop gracefully stops the evolution optimizer
func (eo *EvolutionOptimizer) Stop() error {
	eo.mu.Lock()
	defer eo.mu.Unlock()

	if !eo.running {
		return fmt.Errorf("evolution optimizer not running")
	}

	fmt.Println("🧬 Stopping Evolution Optimizer...")

	// Stop subsystems
	eo.consciousness.Stop()
	eo.scheduler.Stop()

	eo.running = false
	eo.cancel()

	// Take final snapshot
	eo.captureEvolutionSnapshot()

	return nil
}

// runEvolutionLoop executes the main evolution cycle
func (eo *EvolutionOptimizer) runEvolutionLoop() {
	ticker := time.NewTicker(eo.evolutionCycle)
	defer ticker.Stop()

	for {
		select {
		case <-eo.ctx.Done():
			return
		case <-ticker.C:
			eo.evolve()
		}
	}
}

// runIntegrationMonitor monitors subsystem integration
func (eo *EvolutionOptimizer) runIntegrationMonitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-eo.ctx.Done():
			return
		case <-ticker.C:
			eo.integrateSubsystems()
		}
	}
}

// evolve performs one evolution cycle
func (eo *EvolutionOptimizer) evolve() {
	eo.mu.Lock()
	defer eo.mu.Unlock()

	eo.generation++
	previousFitness := eo.fitnessScore

	fmt.Printf("\n🧬 Evolution Cycle %d\n", eo.generation)

	// 1. Evaluate current fitness
	eo.evaluateFitness()

	// 2. Compute fitness gradient
	eo.computeFitnessGradient()

	// 3. Apply evolutionary pressure
	eo.applyEvolutionaryPressure()

	// 4. Mutate traits based on adaptation rate
	eo.mutateTraits()

	// 5. Update emergence metrics
	eo.updateEmergenceMetrics()

	// 6. Consolidate knowledge if in rest phase
	if eo.shouldDream() {
		eo.triggerDreamConsolidation()
	}

	// Calculate evolution velocity
	eo.evolutionVelocity = eo.fitnessScore - previousFitness

	// Capture snapshot
	eo.captureEvolutionSnapshot()

	// Report progress
	fmt.Printf("   Fitness: %.3f (Δ%.4f)\n", eo.fitnessScore, eo.evolutionVelocity)
	fmt.Printf("   Wisdom: %.3f | Patterns: %.3f | Emergence: %.3f\n",
		eo.wisdomDepth, eo.patternComplexity, eo.emergenceIndex)
}

// evaluateFitness computes the current fitness score
func (eo *EvolutionOptimizer) evaluateFitness() {
	// Gather metrics from subsystems
	socMetrics := eo.consciousness.GetMetrics()
	schedMetrics := eo.scheduler.GetMetrics()
	dreamMetrics := eo.dreamIntegration.GetMetrics()

	// Extract relevant values
	thoughtCount, _ := socMetrics["total_thoughts"].(uint64)
	insightCount, _ := socMetrics["insight_count"].(uint64)
	cycleCount, _ := schedMetrics["total_cycles"].(uint64)
	wisdomCount := dreamMetrics["total_wisdom"].(int)
	patternCount := dreamMetrics["total_patterns"].(int)

	// Calculate component fitness scores
	thoughtFitness := math.Min(1.0, float64(thoughtCount)/100.0)
	insightRatio := 0.0
	if thoughtCount > 0 {
		insightRatio = float64(insightCount) / float64(thoughtCount)
	}
	cycleFitness := math.Min(1.0, float64(cycleCount)/10.0)
	wisdomFitness := math.Min(1.0, float64(wisdomCount)/5.0)
	patternFitness := math.Min(1.0, float64(patternCount)/10.0)

	// Weighted fitness calculation
	eo.fitnessScore = (thoughtFitness * 0.15) +
		(insightRatio * 0.25) +
		(cycleFitness * 0.15) +
		(wisdomFitness * 0.25) +
		(patternFitness * 0.20)

	// Update derived metrics
	eo.wisdomDepth = wisdomFitness
	eo.patternComplexity = patternFitness

	// Apply genetic trait modifiers
	for _, trait := range eo.geneticTraits {
		eo.fitnessScore += trait.Value * trait.Expression * trait.FitnessImpact * 0.1
	}

	// Clamp to valid range
	eo.fitnessScore = math.Max(0.0, math.Min(1.0, eo.fitnessScore))
}

// computeFitnessGradient calculates the gradient in fitness landscape
func (eo *EvolutionOptimizer) computeFitnessGradient() {
	if len(eo.evolutionHistory) < 2 {
		return
	}

	// Compare current to previous snapshot
	current := eo.evolutionHistory[len(eo.evolutionHistory)-1]
	previous := eo.evolutionHistory[len(eo.evolutionHistory)-2]

	// Compute gradient for each dimension
	traitNames := []string{
		"curiosity_drive", "wisdom_accumulation", "pattern_recognition",
		"coherence_maintenance", "adaptive_learning",
	}

	for i, name := range traitNames {
		if i < len(eo.fitnessLandscape.Gradient) {
			prevVal, prevOk := previous.TraitSnapshot[name]
			currVal, currOk := current.TraitSnapshot[name]
			if prevOk && currOk && currVal != prevVal {
				fitnessDelta := current.FitnessScore - previous.FitnessScore
				traitDelta := currVal - prevVal
				if traitDelta != 0 {
					eo.fitnessLandscape.Gradient[i] = fitnessDelta / traitDelta
				}
			}
		}
	}

	// Detect local optima
	gradientMagnitude := 0.0
	for _, g := range eo.fitnessLandscape.Gradient {
		gradientMagnitude += g * g
	}
	gradientMagnitude = math.Sqrt(gradientMagnitude)

	eo.fitnessLandscape.LocalOptima = gradientMagnitude < 0.01 && eo.evolutionVelocity < 0.001
}

// applyEvolutionaryPressure applies selection pressure to traits
func (eo *EvolutionOptimizer) applyEvolutionaryPressure() {
	// If stuck in local optima, increase exploration
	if eo.fitnessLandscape.LocalOptima {
		eo.adaptationRate = math.Min(0.3, eo.adaptationRate*1.5)
		fmt.Println("   ⚡ Local optima detected - increasing adaptation rate")
	} else {
		// Gradually reduce adaptation rate as fitness improves
		eo.adaptationRate = math.Max(0.05, eo.adaptationRate*0.95)
	}

	// Update trait expressions based on gradient
	traitNames := []string{
		"curiosity_drive", "wisdom_accumulation", "pattern_recognition",
		"coherence_maintenance", "adaptive_learning",
	}

	for i, name := range traitNames {
		if trait, ok := eo.geneticTraits[name]; ok {
			if i < len(eo.fitnessLandscape.Gradient) {
				// Increase expression of traits with positive gradient
				gradient := eo.fitnessLandscape.Gradient[i]
				trait.Expression = math.Max(0.1, math.Min(1.0,
					trait.Expression+gradient*0.1))
				eo.geneticTraits[name] = trait
			}
		}
	}
}

// mutateTraits applies random mutations to genetic traits
func (eo *EvolutionOptimizer) mutateTraits() {
	now := time.Now()

	for name, trait := range eo.geneticTraits {
		// Probability of mutation based on mutability and adaptation rate
		mutationChance := trait.Mutability * eo.adaptationRate

		// Apply mutation
		if mutationChance > 0 {
			// Gaussian mutation centered on current value
			mutation := (randFloat64() - 0.5) * 2.0 * mutationChance
			trait.Value = math.Max(0.0, math.Min(1.0, trait.Value+mutation))
			trait.LastMutation = now
			trait.Generation = eo.generation
			eo.geneticTraits[name] = trait
		}
	}
}

// updateEmergenceMetrics calculates emergence indicators
func (eo *EvolutionOptimizer) updateEmergenceMetrics() {
	// Emergence = combination of wisdom depth, pattern complexity, and coherence
	eo.emergenceIndex = (eo.wisdomDepth*0.4 +
		eo.patternComplexity*0.3 +
		eo.coherenceStability*0.3) *
		math.Log(float64(eo.generation+1)) / 10.0

	eo.emergenceIndex = math.Min(1.0, eo.emergenceIndex)

	// Update coherence stability based on recent evolution velocity
	if len(eo.evolutionHistory) >= 5 {
		velocityVariance := 0.0
		recentVelocities := make([]float64, 0)

		for i := len(eo.evolutionHistory) - 5; i < len(eo.evolutionHistory); i++ {
			if i > 0 {
				velocity := eo.evolutionHistory[i].FitnessScore -
					eo.evolutionHistory[i-1].FitnessScore
				recentVelocities = append(recentVelocities, velocity)
			}
		}

		if len(recentVelocities) > 0 {
			mean := 0.0
			for _, v := range recentVelocities {
				mean += v
			}
			mean /= float64(len(recentVelocities))

			for _, v := range recentVelocities {
				velocityVariance += (v - mean) * (v - mean)
			}
			velocityVariance /= float64(len(recentVelocities))
		}

		// Lower variance = higher stability
		eo.coherenceStability = 1.0 / (1.0 + velocityVariance*10.0)
	}
}

// shouldDream determines if dream consolidation should occur
func (eo *EvolutionOptimizer) shouldDream() bool {
	// Dream every 5 generations or when emergence threshold reached
	return eo.generation%5 == 0 || eo.emergenceIndex > 0.7
}

// triggerDreamConsolidation initiates knowledge consolidation
func (eo *EvolutionOptimizer) triggerDreamConsolidation() {
	fmt.Println("   🌙 Triggering dream consolidation...")

	// Get thoughts from consciousness for consolidation
	thoughts := eo.consciousness.GetThoughtsForConsolidation()

	// Add to dream integration
	eo.dreamIntegration.episodicMemories = append(eo.dreamIntegration.episodicMemories, thoughts...)

	// Run consolidation
	if err := eo.dreamIntegration.ConsolidateKnowledge(eo.ctx); err != nil {
		fmt.Printf("   ⚠️ Dream consolidation error: %v\n", err)
	}

	// Extract wisdom and feed back to consciousness
	wisdom := eo.dreamIntegration.ExtractWisdom()
	eo.wisdomDepth = wisdom
}

// integrateSubsystems synchronizes the subsystems
func (eo *EvolutionOptimizer) integrateSubsystems() {
	eo.mu.Lock()
	defer eo.mu.Unlock()

	// Sync consciousness focus with scheduler phase
	phase := eo.scheduler.GetCurrentPhase()
	switch phase {
	case PhaseExpressive:
		eo.consciousness.SetMood("expressive")
		eo.consciousness.SetFocus("expressing current understanding")
	case PhaseReflective:
		eo.consciousness.SetMood("reflective")
		eo.consciousness.SetFocus("analyzing past experiences")
	case PhaseAnticipatory:
		eo.consciousness.SetMood("anticipatory")
		eo.consciousness.SetFocus("planning future actions")
	}

	// Adjust thought interval based on evolution velocity
	if eo.evolutionVelocity > 0 {
		// Positive velocity - slow down to consolidate
		eo.consciousness.thoughtInterval = time.Duration(
			float64(10*time.Second) * (1.0 + eo.evolutionVelocity*5.0))
	} else {
		// Negative or zero velocity - speed up exploration
		eo.consciousness.thoughtInterval = time.Duration(
			float64(10*time.Second) * math.Max(0.5, 1.0+eo.evolutionVelocity*2.0))
	}

	// Update genetic traits based on consciousness metrics
	socMetrics := eo.consciousness.GetMetrics()
	if gaps, ok := socMetrics["knowledge_gaps"].(int); ok && gaps > 0 {
		if trait, exists := eo.geneticTraits["curiosity_drive"]; exists {
			trait.Expression = math.Min(1.0, trait.Expression+0.01*float64(gaps))
			eo.geneticTraits["curiosity_drive"] = trait
		}
	}
}

// captureEvolutionSnapshot records current evolutionary state
func (eo *EvolutionOptimizer) captureEvolutionSnapshot() {
	traitSnapshot := make(map[string]float64)
	for name, trait := range eo.geneticTraits {
		traitSnapshot[name] = trait.Value * trait.Expression
	}

	snapshot := EvolutionSnapshot{
		Generation:        eo.generation,
		Timestamp:         time.Now(),
		FitnessScore:      eo.fitnessScore,
		WisdomDepth:       eo.wisdomDepth,
		PatternComplexity: eo.patternComplexity,
		CoherenceScore:    eo.coherenceStability,
		EmergenceIndex:    eo.emergenceIndex,
		TraitSnapshot:     traitSnapshot,
	}

	eo.evolutionHistory = append(eo.evolutionHistory, snapshot)

	// Keep only last 100 snapshots
	if len(eo.evolutionHistory) > 100 {
		eo.evolutionHistory = eo.evolutionHistory[len(eo.evolutionHistory)-100:]
	}
}

// GetEvolutionStatus returns current evolution status
func (eo *EvolutionOptimizer) GetEvolutionStatus() map[string]interface{} {
	eo.mu.RLock()
	defer eo.mu.RUnlock()

	return map[string]interface{}{
		"generation":          eo.generation,
		"fitness_score":       eo.fitnessScore,
		"evolution_velocity":  eo.evolutionVelocity,
		"adaptation_rate":     eo.adaptationRate,
		"wisdom_depth":        eo.wisdomDepth,
		"pattern_complexity":  eo.patternComplexity,
		"coherence_stability": eo.coherenceStability,
		"emergence_index":     eo.emergenceIndex,
		"local_optima":        eo.fitnessLandscape.LocalOptima,
		"running":             eo.running,
	}
}

// GetGeneticTraits returns the current genetic traits
func (eo *EvolutionOptimizer) GetGeneticTraits() map[string]GeneticTrait {
	eo.mu.RLock()
	defer eo.mu.RUnlock()

	// Return a copy
	traits := make(map[string]GeneticTrait)
	for k, v := range eo.geneticTraits {
		traits[k] = v
	}
	return traits
}

// GetEvolutionHistory returns the evolution history
func (eo *EvolutionOptimizer) GetEvolutionHistory() []EvolutionSnapshot {
	eo.mu.RLock()
	defer eo.mu.RUnlock()

	// Return a copy
	history := make([]EvolutionSnapshot, len(eo.evolutionHistory))
	copy(history, eo.evolutionHistory)
	return history
}

// GetConsciousness returns the stream of consciousness
func (eo *EvolutionOptimizer) GetConsciousness() *StreamOfConsciousness {
	return eo.consciousness
}

// GetScheduler returns the echobeats scheduler
func (eo *EvolutionOptimizer) GetScheduler() *EchobeatsScheduler {
	return eo.scheduler
}

// GetDreamIntegration returns the dream integration system
func (eo *EvolutionOptimizer) GetDreamIntegration() *EchoDreamKnowledgeIntegration {
	return eo.dreamIntegration
}

// InjectExternalStimulus injects an external thought or stimulus
func (eo *EvolutionOptimizer) InjectExternalStimulus(stimulus string, importance float64) {
	eo.mu.Lock()
	defer eo.mu.Unlock()

	// Add to consciousness as a knowledge gap if important
	if importance > 0.7 {
		eo.consciousness.AddKnowledgeGap(stimulus, importance)
	}

	// Add as interest if moderately important
	if importance > 0.5 {
		eo.consciousness.AddInterest(stimulus, importance)
	}

	// Add as goal if highly important
	if importance > 0.8 {
		eo.consciousness.AddGoal(fmt.Sprintf("Explore: %s", stimulus))
	}
}

// Helper function for random float64
func randFloat64() float64 {
	return float64(time.Now().UnixNano()%1000) / 1000.0
}
