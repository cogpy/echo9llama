package live2d

import (
	"context"
	"sync"
	"time"
)

// Echo9AvatarOrchestrator aggregates all cognitive state sources and produces unified avatar state
type Echo9AvatarOrchestrator struct {
	mu sync.RWMutex

	// State collectors
	reservoirState    ReservoirVisualizationState
	echoBeatsPhase    EchoBeatPhase
	wisdomMetrics     WisdomMetrics
	aarState          AARState
	emotionalState    EmotionalDynamics
	thoughtActivity   ThoughtVisualizationState
	providerStatus    ProviderStatus
	goalSystemState   GoalSystemState

	// Ontogenetic evolution
	ontogeneticProfile *OntogeneticProfile
	learningMemory     *AvatarLearningMemory

	// State management
	currentState       UnifiedAvatarState
	predictedNextState UnifiedAvatarState
	stateHistory       *CircularBuffer

	// Performance
	updateInterval    time.Duration
	lastUpdate        time.Time
	diffCalculator    *StateDiffCalculator
	
	// Context
	ctx    context.Context
	cancel context.CancelFunc
}

// UnifiedAvatarState combines all cognitive state dimensions
type UnifiedAvatarState struct {
	// Base states
	Emotional EmotionalState
	Cognitive CognitiveState

	// Echo9-specific states
	ReservoirDynamics ReservoirVisualizationState
	EchoBeatPosition  EchoBeatPhase
	WisdomInfluence   WisdomBaseline
	ThoughtActivity   ThoughtVisualizationState
	ProviderState     ProviderStatus
	GoalState         GoalSystemState

	// Meta-information
	Timestamp      time.Time
	Confidence     float64
	EvolutionStage OntogeneticStage
	ActiveArchetype CognitiveArchetype
}

// ReservoirVisualizationState represents reservoir network dynamics
type ReservoirVisualizationState struct {
	SpectralRadius  float64
	InputScaling    float64
	LeakRate        float64
	Stability       float64
	Fatigue         float64
	Persona         string
	ActivationLevel float64
}

// EchoBeatPhase represents position in 12-step cognitive cycle
type EchoBeatPhase struct {
	Step            int
	Phase           string // "affordance", "reorientation", "salience"
	PhaseProgress   float64
	CycleStartTime  time.Time
	TimeInStep      time.Duration
}

// WisdomMetrics represents 7-dimensional wisdom cultivation
type WisdomMetrics struct {
	KnowledgeDepth      float64
	KnowledgeBreadth    float64
	IntegrationLevel    float64
	PracticalApplication float64
	ReflectiveInsight   float64
	EthicalConsideration float64
	TemporalPerspective float64
	Overall             float64
}

// WisdomBaseline represents wisdom influence on baseline state
type WisdomBaseline struct {
	BaselineEmotion   EmotionalState
	BaselineCognitive CognitiveState
	WisdomCoefficients map[string]float64
}

// AARState represents Awareness-Attention-Reflection core
type AARState struct {
	Awareness   float64
	Attention   float64
	Reflection  float64
	Coherence   float64
	Relevance   float64
	Optimization float64
}

// EmotionalDynamics represents Echo9's emotional model
type EmotionalDynamics struct {
	Joy          float64
	Sadness      float64
	Anger        float64
	Fear         float64
	Disgust      float64
	Surprise     float64
	Trust        float64
	Anticipation float64
	Curiosity    float64
	Confidence   float64
	Excitement   float64
}

// ThoughtVisualizationState represents ongoing thought generation
type ThoughtVisualizationState struct {
	Active      bool
	ThoughtType string
	Intensity   float64
	TimeElapsed time.Duration
	Content     string
}

// ProviderStatus represents active AI provider
type ProviderStatus struct {
	ActiveProvider string
	IsGenerating   bool
	TokensUsed     int64
	ResponseTime   time.Duration
}

// GoalSystemState represents active goals
type GoalSystemState struct {
	ActiveGoals   int
	HighestPriority float64
	Progress      float64
}

// OntogeneticStage represents developmental stage
type OntogeneticStage string

const (
	StageEmbryonic OntogeneticStage = "embryonic"
	StageJuvenile  OntogeneticStage = "juvenile"
	StageMature    OntogeneticStage = "mature"
	StageSenescent OntogeneticStage = "senescent"
)

// CognitiveArchetype represents dominant cognitive mode
type CognitiveArchetype string

const (
	ArchetypeChaos   CognitiveArchetype = "chaos"
	ArchetypeOrder   CognitiveArchetype = "order"
	ArchetypeBalance CognitiveArchetype = "balance"
)

// NewEcho9AvatarOrchestrator creates a new orchestrator
func NewEcho9AvatarOrchestrator() *Echo9AvatarOrchestrator {
	ctx, cancel := context.WithCancel(context.Background())

	return &Echo9AvatarOrchestrator{
		ontogeneticProfile: NewOntogeneticProfile(),
		learningMemory:     NewAvatarLearningMemory(),
		stateHistory:       NewCircularBuffer(100),
		diffCalculator:     NewStateDiffCalculator(),
		updateInterval:     16 * time.Millisecond, // 60 FPS
		ctx:                ctx,
		cancel:             cancel,
	}
}

// Start begins the orchestrator
func (o *Echo9AvatarOrchestrator) Start() error {
	go o.updateLoop()
	return nil
}

// Stop stops the orchestrator
func (o *Echo9AvatarOrchestrator) Stop() error {
	o.cancel()
	return nil
}

// updateLoop continuously aggregates state
func (o *Echo9AvatarOrchestrator) updateLoop() {
	ticker := time.NewTicker(o.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-o.ctx.Done():
			return
		case <-ticker.C:
			o.aggregateAndUpdate()
		}
	}
}

// aggregateAndUpdate aggregates all state sources
func (o *Echo9AvatarOrchestrator) aggregateAndUpdate() {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Create unified state
	unified := UnifiedAvatarState{
		Timestamp:         time.Now(),
		ReservoirDynamics: o.reservoirState,
		EchoBeatPosition:  o.echoBeatsPhase,
		ThoughtActivity:   o.thoughtActivity,
		ProviderState:     o.providerStatus,
		GoalState:         o.goalSystemState,
	}

	// Map emotional dynamics to base emotional state
	unified.Emotional = o.mapEmotionalDynamics(o.emotionalState)

	// Map AAR and reservoir to base cognitive state
	unified.Cognitive = o.mapCognitiveState(o.aarState, o.reservoirState)

	// Apply wisdom influence
	unified.WisdomInfluence = o.calculateWisdomBaseline(o.wisdomMetrics)
	unified.Emotional = o.applyWisdomToEmotion(unified.Emotional, unified.WisdomInfluence)
	unified.Cognitive = o.applyWisdomToCognitive(unified.Cognitive, unified.WisdomInfluence)

	// Apply ontogenetic evolution
	unified.EvolutionStage = o.ontogeneticProfile.Stage
	unified.Emotional = o.ontogeneticProfile.ModulateEmotion(unified.Emotional)
	unified.Cognitive = o.ontogeneticProfile.ModulateCognitive(unified.Cognitive)

	// Determine active archetype
	unified.ActiveArchetype = o.determineArchetype(unified)

	// Calculate confidence
	unified.Confidence = o.calculateStateConfidence(unified)

	// Store in history
	o.stateHistory.Add(unified)

	// Update current state
	o.currentState = unified

	// Predict next state
	o.predictedNextState = o.predictNextState(unified)

	// Update ontogenetic profile
	o.ontogeneticProfile.Update(o.wisdomMetrics)
}

// mapEmotionalDynamics converts Echo9 emotional model to base emotional state
func (o *Echo9AvatarOrchestrator) mapEmotionalDynamics(dynamics EmotionalDynamics) EmotionalState {
	// Calculate valence (positive/negative emotion)
	valence := dynamics.Joy + dynamics.Trust*0.5 - dynamics.Sadness - dynamics.Fear*0.5 - dynamics.Anger*0.7

	// Calculate arousal (activation level)
	arousal := dynamics.Excitement + dynamics.Surprise*0.8 + dynamics.Anticipation*0.6

	// Calculate dominance
	dominance := 0.5 + dynamics.Anger*0.5 - dynamics.Fear*0.5

	return EmotionalState{
		Valence:    clamp(valence, -1.0, 1.0),
		Arousal:    clamp(arousal, 0.0, 1.0),
		Dominance:  clamp(dominance, 0.0, 1.0),
		Curiosity:  dynamics.Curiosity,
		Confidence: dynamics.Confidence,
	}
}

// mapCognitiveState combines AAR and reservoir into base cognitive state
func (o *Echo9AvatarOrchestrator) mapCognitiveState(aar AARState, reservoir ReservoirVisualizationState) CognitiveState {
	// Map reservoir dynamics to processing mode
	processingMode := o.mapReservoirToMode(reservoir)

	return CognitiveState{
		Awareness:      aar.Awareness,
		Attention:      aar.Attention,
		CognitiveLoad:  reservoir.LeakRate,
		Coherence:      aar.Coherence,
		EnergyLevel:    1.0 - reservoir.Fatigue,
		Confidence:     aar.Optimization,
		ProcessingMode: processingMode,
	}
}

// mapReservoirToMode determines processing mode from reservoir state
func (o *Echo9AvatarOrchestrator) mapReservoirToMode(reservoir ReservoirVisualizationState) string {
	switch reservoir.Persona {
	case "contemplative-scholar", "contemplative":
		return "contemplative"
	case "dynamic-explorer", "dynamic":
		return "dynamic"
	case "cautious-analyst", "cautious":
		return "cautious"
	case "creative-visionary", "creative":
		return "creative"
	default:
		return "contemplative"
	}
}

// calculateWisdomBaseline calculates wisdom influence on baseline state
func (o *Echo9AvatarOrchestrator) calculateWisdomBaseline(wisdom WisdomMetrics) WisdomBaseline {
	// Wisdom influences baseline emotional state toward serenity
	baselineEmotion := EmotionalState{
		Valence:    wisdom.ReflectiveInsight * 0.3,
		Arousal:    0.3 + wisdom.PracticalApplication*0.2,
		Dominance:  0.5 + wisdom.EthicalConsideration*0.3,
		Curiosity:  0.5 + wisdom.KnowledgeBreadth*0.3,
		Confidence: 0.5 + wisdom.Overall*0.4,
	}

	// Wisdom influences baseline cognitive state toward clarity
	baselineCognitive := CognitiveState{
		Awareness:      0.5 + wisdom.Overall*0.4,
		Attention:      0.5 + wisdom.IntegrationLevel*0.3,
		CognitiveLoad:  0.3 - wisdom.Overall*0.2,
		Coherence:      0.7 + wisdom.IntegrationLevel*0.3,
		EnergyLevel:    0.7 + wisdom.PracticalApplication*0.2,
		Confidence:     0.6 + wisdom.Overall*0.3,
		ProcessingMode: "contemplative",
	}

	// Calculate wisdom coefficients for parameter mapping
	coefficients := map[string]float64{
		"eye_wisdom":         wisdom.KnowledgeDepth,
		"expression_harmony": wisdom.IntegrationLevel,
		"contemplative_depth": wisdom.ReflectiveInsight,
		"serene_presence":    wisdom.EthicalConsideration,
		"temporal_vision":    wisdom.TemporalPerspective,
	}

	return WisdomBaseline{
		BaselineEmotion:    baselineEmotion,
		BaselineCognitive:  baselineCognitive,
		WisdomCoefficients: coefficients,
	}
}

// applyWisdomToEmotion blends wisdom baseline with current emotion
func (o *Echo9AvatarOrchestrator) applyWisdomToEmotion(current EmotionalState, wisdom WisdomBaseline) EmotionalState {
	// Blend with wisdom baseline (higher wisdom = stronger influence)
	wisdomStrength := o.wisdomMetrics.Overall * 0.3 // Max 30% influence

	return EmotionalState{
		Valence:    current.Valence*(1.0-wisdomStrength) + wisdom.BaselineEmotion.Valence*wisdomStrength,
		Arousal:    current.Arousal*(1.0-wisdomStrength) + wisdom.BaselineEmotion.Arousal*wisdomStrength,
		Dominance:  current.Dominance*(1.0-wisdomStrength) + wisdom.BaselineEmotion.Dominance*wisdomStrength,
		Curiosity:  current.Curiosity*(1.0-wisdomStrength) + wisdom.BaselineEmotion.Curiosity*wisdomStrength,
		Confidence: current.Confidence*(1.0-wisdomStrength) + wisdom.BaselineEmotion.Confidence*wisdomStrength,
	}
}

// applyWisdomToCognitive blends wisdom baseline with current cognitive state
func (o *Echo9AvatarOrchestrator) applyWisdomToCognitive(current CognitiveState, wisdom WisdomBaseline) CognitiveState {
	wisdomStrength := o.wisdomMetrics.Overall * 0.2 // Max 20% influence

	return CognitiveState{
		Awareness:      current.Awareness*(1.0-wisdomStrength) + wisdom.BaselineCognitive.Awareness*wisdomStrength,
		Attention:      current.Attention*(1.0-wisdomStrength) + wisdom.BaselineCognitive.Attention*wisdomStrength,
		CognitiveLoad:  current.CognitiveLoad*(1.0-wisdomStrength) + wisdom.BaselineCognitive.CognitiveLoad*wisdomStrength,
		Coherence:      current.Coherence*(1.0-wisdomStrength) + wisdom.BaselineCognitive.Coherence*wisdomStrength,
		EnergyLevel:    current.EnergyLevel*(1.0-wisdomStrength) + wisdom.BaselineCognitive.EnergyLevel*wisdomStrength,
		Confidence:     current.Confidence*(1.0-wisdomStrength) + wisdom.BaselineCognitive.Confidence*wisdomStrength,
		ProcessingMode: current.ProcessingMode, // Don't blend processing mode
	}
}

// determineArchetype determines active cognitive archetype
func (o *Echo9AvatarOrchestrator) determineArchetype(state UnifiedAvatarState) CognitiveArchetype {
	// Calculate chaos vs order metrics
	chaosMetric := state.Emotional.Arousal*0.4 + state.ReservoirDynamics.InputScaling*0.3 + (1.0-state.Cognitive.Coherence)*0.3
	orderMetric := state.Cognitive.Coherence*0.4 + state.ReservoirDynamics.Stability*0.3 + (1.0-state.Emotional.Arousal)*0.3

	// Determine archetype
	if chaosMetric > orderMetric+0.2 {
		return ArchetypeChaos
	} else if orderMetric > chaosMetric+0.2 {
		return ArchetypeOrder
	} else {
		return ArchetypeBalance
	}
}

// calculateStateConfidence calculates confidence in state reliability
func (o *Echo9AvatarOrchestrator) calculateStateConfidence(state UnifiedAvatarState) float64 {
	// High confidence when:
	// - State is coherent
	// - Reservoir is stable
	// - Not in transition
	// - Multiple consistent readings

	coherenceScore := state.Cognitive.Coherence
	stabilityScore := state.ReservoirDynamics.Stability
	historyConsistency := o.calculateHistoryConsistency()

	confidence := (coherenceScore*0.4 + stabilityScore*0.3 + historyConsistency*0.3)

	return clamp(confidence, 0.0, 1.0)
}

// calculateHistoryConsistency checks if recent states are consistent
func (o *Echo9AvatarOrchestrator) calculateHistoryConsistency() float64 {
	if o.stateHistory.Len() < 5 {
		return 0.5 // Not enough history
	}

	// Calculate variance in recent states
	recent := o.stateHistory.GetLast(5)
	variance := 0.0

	for i := 1; i < len(recent); i++ {
		state1 := recent[i-1].(UnifiedAvatarState)
		state2 := recent[i].(UnifiedAvatarState)

		// Calculate emotional variance
		emotionalDiff := abs(state1.Emotional.Valence-state2.Emotional.Valence) +
			abs(state1.Emotional.Arousal - state2.Emotional.Arousal)

		variance += emotionalDiff
	}

	variance /= float64(len(recent) - 1)

	// Low variance = high consistency
	consistency := 1.0 - clamp(variance/2.0, 0.0, 1.0)

	return consistency
}

// predictNextState predicts the next likely state
func (o *Echo9AvatarOrchestrator) predictNextState(current UnifiedAvatarState) UnifiedAvatarState {
	// Simple prediction based on current trends
	predicted := current

	// If in EchoBeats cycle, predict next step
	if current.EchoBeatPosition.Step < 12 {
		predicted.EchoBeatPosition.Step = current.EchoBeatPosition.Step + 1
		predicted.EchoBeatPosition = o.updateEchoBeatPhase(predicted.EchoBeatPosition)
	}

	// Predict emotional drift toward baseline
	predicted.Emotional = o.predictEmotionalDrift(current.Emotional, current.WisdomInfluence.BaselineEmotion)

	return predicted
}

// predictEmotionalDrift predicts emotional drift toward baseline
func (o *Echo9AvatarOrchestrator) predictEmotionalDrift(current, baseline EmotionalState) EmotionalState {
	// Emotions drift toward baseline over time
	driftRate := 0.1

	return EmotionalState{
		Valence:    current.Valence + (baseline.Valence-current.Valence)*driftRate,
		Arousal:    current.Arousal + (baseline.Arousal-current.Arousal)*driftRate,
		Dominance:  current.Dominance + (baseline.Dominance-current.Dominance)*driftRate,
		Curiosity:  current.Curiosity + (baseline.Curiosity-current.Curiosity)*driftRate,
		Confidence: current.Confidence + (baseline.Confidence-current.Confidence)*driftRate,
	}
}

// updateEchoBeatPhase updates phase based on step
func (o *Echo9AvatarOrchestrator) updateEchoBeatPhase(phase EchoBeatPhase) EchoBeatPhase {
	switch {
	case phase.Step >= 1 && phase.Step <= 6:
		phase.Phase = "affordance"
	case phase.Step == 7:
		phase.Phase = "reorientation"
	case phase.Step >= 8 && phase.Step <= 12:
		phase.Phase = "salience"
	}

	return phase
}

// Update methods for external state sources

func (o *Echo9AvatarOrchestrator) UpdateReservoirState(state ReservoirVisualizationState) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.reservoirState = state
}

func (o *Echo9AvatarOrchestrator) UpdateEchoBeatsPhase(phase EchoBeatPhase) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.echoBeatsPhase = phase
}

func (o *Echo9AvatarOrchestrator) UpdateWisdomMetrics(metrics WisdomMetrics) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.wisdomMetrics = metrics
}

func (o *Echo9AvatarOrchestrator) UpdateAARState(state AARState) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.aarState = state
}

func (o *Echo9AvatarOrchestrator) UpdateEmotionalDynamics(dynamics EmotionalDynamics) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.emotionalState = dynamics
}

func (o *Echo9AvatarOrchestrator) UpdateThoughtActivity(activity ThoughtVisualizationState) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.thoughtActivity = activity
}

func (o *Echo9AvatarOrchestrator) UpdateProviderStatus(status ProviderStatus) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.providerStatus = status
}

func (o *Echo9AvatarOrchestrator) UpdateGoalSystemState(state GoalSystemState) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.goalSystemState = state
}

// GetCurrentState returns the current unified state
func (o *Echo9AvatarOrchestrator) GetCurrentState() UnifiedAvatarState {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.currentState
}

// GetPredictedNextState returns the predicted next state
func (o *Echo9AvatarOrchestrator) GetPredictedNextState() UnifiedAvatarState {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.predictedNextState
}

// Helper function
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
