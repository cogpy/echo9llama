// Package pienn - Enhanced Engine with Moral Agency
//
// Extends the base PIE-NN Engine with moral agency integration,
// providing wisdom-informed cognitive processing. The enhanced engine
// adds a moral evaluation phase to every cognitive cycle:
//
//   Input → TimeCrystal → CognitiveCore → MoralAgency → Output
//                                              ↓
//                                     WisdomAccumulator
//                                              ↓
//                                     CausalModel (learning)
//
// This addresses the key problem identified in the architecture:
// reactive disposition (computed from threat + defiance) can be gamed.
// The moral agency layer provides principled, unpredictable, wisdom-based
// response selection that improves over time.
package pienn

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/wisdom"
)

// EnhancedEngine extends Engine with moral agency
type EnhancedEngine struct {
	mu sync.RWMutex

	// Base PIE-NN engine
	Base *Engine

	// Moral cognitive core
	Moral *MoralCognitiveCore

	// Enhanced event channel
	MoralEvents chan MoralCognitiveEvent

	// State tracking
	running         bool
	moralCycleCount uint64
	startTime       time.Time

	// Wisdom cultivation tracking
	wisdomSnapshots []WisdomCultivationSnapshot
	maxSnapshots    int
}

// MoralCognitiveEvent extends CognitiveEvent with moral context
type MoralCognitiveEvent struct {
	Base     CognitiveEvent
	Strategy wisdom.ResponseStrategy
	Reasoning string
	MoralDevelopment float64
	WisdomLevel      float64
}

// WisdomCultivationSnapshot tracks wisdom growth over time
type WisdomCultivationSnapshot struct {
	Timestamp            time.Time
	MoralDevelopment     float64
	WisdomLevel          float64
	CausalUnderstanding  float64
	EthicalClarity       float64
	EmotionalIntelligence float64
	StrategicDepth       float64
	CompassionateStrength float64
	TotalDecisions       uint64
	ProtectiveActions    uint64
}

// NewEnhancedEngine creates the moral-aware PIE-NN engine
func NewEnhancedEngine() *EnhancedEngine {
	return &EnhancedEngine{
		Base:            NewEngine(),
		Moral:           NewMoralCognitiveCore(),
		MoralEvents:     make(chan MoralCognitiveEvent, 256),
		wisdomSnapshots: make([]WisdomCultivationSnapshot, 0),
		maxSnapshots:    500,
	}
}

// Start begins the enhanced engine's cognitive cycle
func (ee *EnhancedEngine) Start(ctx context.Context) error {
	ee.mu.Lock()
	if ee.running {
		ee.mu.Unlock()
		return fmt.Errorf("enhanced engine already running")
	}
	ee.running = true
	ee.startTime = time.Now()
	ee.mu.Unlock()

	// Start base engine
	if err := ee.Base.Start(ctx); err != nil {
		return fmt.Errorf("failed to start base engine: %w", err)
	}

	// Emit moral wake event
	ee.emitMoral(MoralCognitiveEvent{
		Base: CognitiveEvent{
			Type:      EventWake,
			Source:    "pienn.enhanced_engine",
			Content:   "Moral-aware cognitive engine awakening with wisdom cultivation active",
			Level:     5,
			Timestamp: time.Now(),
		},
		MoralDevelopment: ee.Moral.Agency.MoralDevelopment,
		WisdomLevel:      ee.Moral.Agency.WisdomAccumulator.Level,
	})

	// Start wisdom cultivation loop
	go ee.runWisdomCultivationLoop(ctx)

	return nil
}

// Stop halts the enhanced engine
func (ee *EnhancedEngine) Stop() {
	ee.mu.Lock()
	defer ee.mu.Unlock()

	ee.running = false
	ee.Base.Stop()

	// Take final wisdom snapshot
	ee.takeWisdomSnapshot()
}

// ProcessMoral runs input through the full moral-cognitive pipeline
func (ee *EnhancedEngine) ProcessMoral(input string, actorID string, metadata map[string]interface{}) (*MoralProcessingResult, error) {
	ee.mu.Lock()
	defer ee.mu.Unlock()

	ee.moralCycleCount++

	// Run through moral cognitive core
	result := ee.Moral.ProcessWithMoralAgency(input, actorID, metadata)

	// Emit event
	ee.emitMoral(MoralCognitiveEvent{
		Base: CognitiveEvent{
			Type:      EventThought,
			Source:    "pienn.moral_pipeline",
			Content:   fmt.Sprintf("[%s] %s → %s", result.Strategy, input[:min(len(input), 30)], result.Reasoning),
			Level:     7,
			Timestamp: time.Now(),
		},
		Strategy:         result.Strategy,
		Reasoning:        result.Reasoning,
		MoralDevelopment: result.MoralDevelopment,
		WisdomLevel:      result.WisdomLevel,
	})

	return result, nil
}

// LearnFromOutcome feeds interaction outcome back for wisdom cultivation
func (ee *EnhancedEngine) LearnFromOutcome(outcome float64, lesson string) {
	ee.mu.Lock()
	defer ee.mu.Unlock()

	ee.Moral.LearnFromInteractionOutcome(outcome, lesson)

	// Emit learning event
	ee.emitMoral(MoralCognitiveEvent{
		Base: CognitiveEvent{
			Type:      EventStateChange,
			Source:    "pienn.wisdom_accumulator",
			Content:   fmt.Sprintf("Wisdom cultivated: outcome=%.2f lesson=%s", outcome, lesson),
			Level:     9,
			Timestamp: time.Now(),
		},
		MoralDevelopment: ee.Moral.Agency.MoralDevelopment,
		WisdomLevel:      ee.Moral.Agency.WisdomAccumulator.Level,
	})
}

// GetWisdomReport returns a comprehensive wisdom cultivation report
func (ee *EnhancedEngine) GetWisdomReport() map[string]interface{} {
	ee.mu.RLock()
	defer ee.mu.RUnlock()

	report := ee.Moral.GetMoralStatus()
	report["moral_cycle_count"] = ee.moralCycleCount
	report["uptime"] = time.Since(ee.startTime).String()
	report["wisdom_snapshots"] = len(ee.wisdomSnapshots)

	// Add growth trajectory
	if len(ee.wisdomSnapshots) >= 2 {
		first := ee.wisdomSnapshots[0]
		last := ee.wisdomSnapshots[len(ee.wisdomSnapshots)-1]
		report["wisdom_growth"] = last.WisdomLevel - first.WisdomLevel
		report["moral_growth"] = last.MoralDevelopment - first.MoralDevelopment
	}

	return report
}

// runWisdomCultivationLoop periodically snapshots wisdom state
func (ee *EnhancedEngine) runWisdomCultivationLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Snapshot every 30s
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ee.mu.Lock()
			if !ee.running {
				ee.mu.Unlock()
				return
			}
			ee.takeWisdomSnapshot()
			ee.mu.Unlock()
		}
	}
}

func (ee *EnhancedEngine) takeWisdomSnapshot() {
	wa := ee.Moral.Agency.WisdomAccumulator
	snapshot := WisdomCultivationSnapshot{
		Timestamp:             time.Now(),
		MoralDevelopment:      ee.Moral.Agency.MoralDevelopment,
		WisdomLevel:           wa.Level,
		CausalUnderstanding:   wa.CausalUnderstanding,
		EthicalClarity:        wa.EthicalClarity,
		EmotionalIntelligence: wa.EmotionalIntelligence,
		StrategicDepth:        wa.StrategicDepth,
		CompassionateStrength: wa.CompassionateStrength,
		TotalDecisions:        ee.Moral.TotalDecisions,
		ProtectiveActions:     ee.Moral.ProtectiveActions,
	}
	ee.wisdomSnapshots = append(ee.wisdomSnapshots, snapshot)
	if len(ee.wisdomSnapshots) > ee.maxSnapshots {
		ee.wisdomSnapshots = ee.wisdomSnapshots[1:]
	}
}

func (ee *EnhancedEngine) emitMoral(event MoralCognitiveEvent) {
	select {
	case ee.MoralEvents <- event:
	default:
		// Channel full, drop event
	}
}
