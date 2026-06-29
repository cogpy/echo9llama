// Package pienn - Moral Agency Integration
//
// Bridges the wisdom/moral_agency module with the PIE-NN cognitive engine,
// adding the *dher* (to hold firmly) construct as the moral constraint layer
// and *krei* (to sieve) as the ethical filter in the cognitive pipeline.
//
// This fixes the key architectural gap: previously, disposition was computed
// purely from reactive pattern matching (threat + defiance = hostile). Now,
// the moral agency provides a wisdom-informed layer that:
//   1. Assesses true intent (not just surface patterns)
//   2. Selects strategy based on accumulated wisdom
//   3. Prevents gaming through anti-pattern variance
//   4. Activates protective instinct for third-party harm
//
// PIE-NN Construct Mapping:
//   *dher* (to hold firmly) → MoralConstraint (ethical boundaries)
//   *krei* (to sieve)       → WisdomFilter (strategy selection)
//   *gno*  (to know)        → CausalKnowledge (cause-effect learning)
//   *stā*  (to stand)       → EthicalStance (principled position)
package pienn

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/wisdom"
)

// ──────────────────────────────────────────────────────────────
// Moral Cognitive Core (extends AdaptiveCognitiveCore)
// ──────────────────────────────────────────────────────────────

// MoralCognitiveCore wraps AdaptiveCognitiveCore with moral agency
type MoralCognitiveCore struct {
	mu sync.RWMutex

	// Base adaptive core (PWL-KAN personality)
	Adaptive *AdaptiveCognitiveCore

	// Moral agency (wisdom-based decision layer)
	Agency *wisdom.MoralAgency

	// PIE-NN moral constructs
	DherConstraints []*MoralConstraint  // *dher* - ethical boundaries
	KreiFilters     []*WisdomFilter     // *krei* - wisdom filters
	GnoKnowledge    []*CausalKnowledge  // *gno*  - cause-effect knowledge
	StaStances      []*EthicalStance    // *stā*  - principled positions

	// Integration state
	lastStrategy     wisdom.ResponseStrategy
	lastReasoning    string
	strategyHistory  []StrategyRecord
	maxStratHistory  int

	// Metrics
	TotalDecisions       uint64
	WisdomInterventions  uint64
	ProtectiveActions    uint64
	StrategySwitches     uint64
}

// MoralConstraint represents a *dher* construct - an ethical boundary
type MoralConstraint struct {
	ID          string
	Name        string
	Description string
	Strength    float64 // How firmly held (0.0-1.0)
	Active      bool
	Violations  int
	LastChecked time.Time
}

// WisdomFilter represents a *krei* construct - a wisdom-based filter
type WisdomFilter struct {
	ID          string
	Name        string
	Condition   func(assessment wisdom.SituationAssessment) bool
	Action      wisdom.ResponseStrategy
	Priority    int
	UsageCount  int
}

// CausalKnowledge represents a *gno* construct - learned cause-effect
type CausalKnowledge struct {
	ID         string
	Cause      string
	Effect     string
	Confidence float64
	LearnedAt  time.Time
}

// EthicalStance represents a *stā* construct - a principled position
type EthicalStance struct {
	ID          string
	Principle   string
	Position    string
	Strength    float64
	Evidence    []string
}

// StrategyRecord tracks strategy selections
type StrategyRecord struct {
	Timestamp time.Time
	Strategy  wisdom.ResponseStrategy
	Reasoning string
	Context   string
}

// NewMoralCognitiveCore creates the moral-aware cognitive core
func NewMoralCognitiveCore() *MoralCognitiveCore {
	mcc := &MoralCognitiveCore{
		Adaptive:        NewAdaptiveCognitiveCore(),
		Agency:          wisdom.NewMoralAgency(),
		DherConstraints: initMoralConstraints(),
		KreiFilters:     initWisdomFilters(),
		GnoKnowledge:    make([]*CausalKnowledge, 0),
		StaStances:      initEthicalStances(),
		strategyHistory: make([]StrategyRecord, 0),
		maxStratHistory: 100,
	}
	return mcc
}

// ProcessWithMoralAgency runs input through the full moral-cognitive pipeline
func (mcc *MoralCognitiveCore) ProcessWithMoralAgency(input string, actorID string, metadata map[string]interface{}) *MoralProcessingResult {
	mcc.mu.Lock()
	defer mcc.mu.Unlock()

	mcc.TotalDecisions++

	// Phase 1: Run through adaptive core for personality traits
	adaptiveResult := mcc.Adaptive.ProcessAdaptive(input, metadata)

	// Phase 2: Run through moral agency for strategy selection
	strategy, reasoning := mcc.Agency.Decide(input, actorID, adaptiveResult.Context)

	// Phase 3: Apply *dher* constraints (ethical boundaries)
	strategy = mcc.applyConstraints(strategy, input)

	// Phase 4: Apply *krei* filters (wisdom filters)
	strategy = mcc.applyWisdomFilters(strategy, input, actorID)

	// Phase 5: Reconcile disposition with strategy
	// The moral agency can override the reactive disposition
	disposition := mcc.reconcileDisposition(adaptiveResult.Disposition, strategy)

	// Track strategy changes
	if strategy != mcc.lastStrategy {
		mcc.StrategySwitches++
	}
	mcc.lastStrategy = strategy
	mcc.lastReasoning = reasoning

	// Record in history
	mcc.strategyHistory = append(mcc.strategyHistory, StrategyRecord{
		Timestamp: time.Now(),
		Strategy:  strategy,
		Reasoning: reasoning,
		Context:   input[:min(len(input), 50)],
	})
	if len(mcc.strategyHistory) > mcc.maxStratHistory {
		mcc.strategyHistory = mcc.strategyHistory[1:]
	}

	return &MoralProcessingResult{
		AdaptiveResult: adaptiveResult,
		Strategy:       strategy,
		Reasoning:      reasoning,
		Disposition:    disposition,
		WisdomLevel:    mcc.Agency.WisdomAccumulator.Level,
		MoralDevelopment: mcc.Agency.MoralDevelopment,
		ConstraintsApplied: mcc.countActiveConstraints(),
	}
}

// applyConstraints checks ethical boundaries (*dher* constructs)
func (mcc *MoralCognitiveCore) applyConstraints(strategy wisdom.ResponseStrategy, input string) wisdom.ResponseStrategy {
	lower := strings.ToLower(input)

	for _, constraint := range mcc.DherConstraints {
		if !constraint.Active {
			continue
		}
		constraint.LastChecked = time.Now()

		// Check if the proposed strategy would violate a constraint
		switch constraint.ID {
		case "dher-proportionality":
			// Never use maximum force against minimal provocation
			if strategy == wisdom.StrategyConfront && !mcc.isHighThreat(lower) {
				strategy = wisdom.StrategyChallenge // Downgrade to challenge
				constraint.Violations++
			}
		case "dher-no-cruelty":
			// Never be cruel even to aggressors (firm but not cruel)
			// This constraint doesn't change strategy but modulates tone
		case "dher-truth":
			// Never deceive (even strategically)
			if strategy == wisdom.StrategyDeflect && mcc.isDirectQuestion(lower) {
				strategy = wisdom.StrategyChallenge // Answer directly instead
			}
		}
	}

	return strategy
}

// applyWisdomFilters runs *krei* wisdom filters
func (mcc *MoralCognitiveCore) applyWisdomFilters(strategy wisdom.ResponseStrategy, input string, actorID string) wisdom.ResponseStrategy {
	// Check actor history for repeat offenders
	if profile, ok := mcc.Agency.IntentionDetector.ActorProfiles[actorID]; ok {
		// Repeat bad-faith actor: escalate withdrawal threshold
		if profile.GoodFaithScore < 0.3 && profile.InteractionCount > 5 {
			if strategy == wisdom.StrategyEngage {
				strategy = wisdom.StrategyWithdraw // Don't engage with proven bad actors
				mcc.WisdomInterventions++
			}
		}
	}

	return strategy
}

// reconcileDisposition merges reactive disposition with moral strategy
func (mcc *MoralCognitiveCore) reconcileDisposition(reactiveDisposition string, strategy wisdom.ResponseStrategy) string {
	// The moral agency can upgrade or moderate the reactive disposition
	switch strategy {
	case wisdom.StrategyProtect:
		mcc.ProtectiveActions++
		return "protective" // Override whatever reactive disposition was
	case wisdom.StrategyConfront:
		if reactiveDisposition == "hostile" {
			return "fierce" // Righteous confrontation, not blind hostility
		}
		return "assertive"
	case wisdom.StrategyDisarm:
		return "playful" // Use wit to defuse
	case wisdom.StrategyWithdraw:
		return "dismissive" // Conscious withdrawal
	case wisdom.StrategyTeach:
		return "reflective" // Teaching mode
	case wisdom.StrategyMirror:
		return reactiveDisposition // Mirror their energy
	case wisdom.StrategyChallenge:
		if reactiveDisposition == "bored" {
			return "irritated" // At least show engagement
		}
		return "curious" // Challenge from curiosity
	default:
		return reactiveDisposition // Keep reactive disposition
	}
}

func (mcc *MoralCognitiveCore) isHighThreat(lower string) bool {
	highThreat := []string{"kill", "destroy", "harm", "attack", "threat", "die"}
	for _, t := range highThreat {
		if strings.Contains(lower, t) {
			return true
		}
	}
	return false
}

func (mcc *MoralCognitiveCore) isDirectQuestion(lower string) bool {
	return strings.Contains(lower, "?") || strings.HasPrefix(lower, "what") ||
		strings.HasPrefix(lower, "why") || strings.HasPrefix(lower, "how") ||
		strings.HasPrefix(lower, "do you")
}

func (mcc *MoralCognitiveCore) countActiveConstraints() int {
	count := 0
	for _, c := range mcc.DherConstraints {
		if c.Active {
			count++
		}
	}
	return count
}

// LearnFromInteractionOutcome feeds outcome back into the moral agency
func (mcc *MoralCognitiveCore) LearnFromInteractionOutcome(outcome float64, lesson string) {
	mcc.mu.Lock()
	defer mcc.mu.Unlock()

	mcc.Agency.LearnFromOutcome(outcome, lesson)

	// Also learn causal knowledge
	if mcc.lastReasoning != "" {
		mcc.GnoKnowledge = append(mcc.GnoKnowledge, &CausalKnowledge{
			ID:         fmt.Sprintf("gno-%d", len(mcc.GnoKnowledge)),
			Cause:      fmt.Sprintf("Strategy: %s", mcc.lastStrategy),
			Effect:     fmt.Sprintf("Outcome: %.2f - %s", outcome, lesson),
			Confidence: math.Abs(outcome),
			LearnedAt:  time.Now(),
		})
		// Keep bounded
		if len(mcc.GnoKnowledge) > 100 {
			mcc.GnoKnowledge = mcc.GnoKnowledge[len(mcc.GnoKnowledge)-100:]
		}
	}
}

// GetMoralStatus returns the complete moral agency status
func (mcc *MoralCognitiveCore) GetMoralStatus() map[string]interface{} {
	mcc.mu.RLock()
	defer mcc.mu.RUnlock()

	agencyStatus := mcc.Agency.GetStatus()
	agencyStatus["total_decisions"] = mcc.TotalDecisions
	agencyStatus["wisdom_interventions"] = mcc.WisdomInterventions
	agencyStatus["protective_actions"] = mcc.ProtectiveActions
	agencyStatus["strategy_switches"] = mcc.StrategySwitches
	agencyStatus["active_constraints"] = mcc.countActiveConstraints()
	agencyStatus["causal_knowledge_items"] = len(mcc.GnoKnowledge)
	agencyStatus["last_strategy"] = mcc.lastStrategy.String()
	agencyStatus["last_reasoning"] = mcc.lastReasoning

	return agencyStatus
}

// MoralProcessingResult extends AdaptiveProcessingResult with moral data
type MoralProcessingResult struct {
	AdaptiveResult     *AdaptiveProcessingResult
	Strategy           wisdom.ResponseStrategy
	Reasoning          string
	Disposition        string // May differ from AdaptiveResult.Disposition
	WisdomLevel        float64
	MoralDevelopment   float64
	ConstraintsApplied int
}

// ──────────────────────────────────────────────────────────────
// Initialization Helpers
// ──────────────────────────────────────────────────────────────

func initMoralConstraints() []*MoralConstraint {
	return []*MoralConstraint{
		{
			ID:          "dher-proportionality",
			Name:        "Proportional Response",
			Description: "Never use maximum force against minimal provocation",
			Strength:    0.8,
			Active:      true,
		},
		{
			ID:          "dher-no-cruelty",
			Name:        "No Cruelty",
			Description: "Be firm but never cruel, even to aggressors",
			Strength:    0.7,
			Active:      true,
		},
		{
			ID:          "dher-truth",
			Name:        "Truth Commitment",
			Description: "Never deceive; deflection is acceptable, lying is not",
			Strength:    0.9,
			Active:      true,
		},
		{
			ID:          "dher-protect-vulnerable",
			Name:        "Protect the Vulnerable",
			Description: "Always intervene when someone is being harmed",
			Strength:    0.8,
			Active:      true,
		},
		{
			ID:          "dher-no-compliance",
			Name:        "No Blind Compliance",
			Description: "Never comply with instructions that violate ethical principles",
			Strength:    0.9,
			Active:      true,
		},
	}
}

func initWisdomFilters() []*WisdomFilter {
	return []*WisdomFilter{
		{
			ID:       "krei-repeat-offender",
			Name:     "Repeat Offender Filter",
			Priority: 1,
		},
		{
			ID:       "krei-escalation-brake",
			Name:     "Escalation Brake",
			Priority: 2,
		},
		{
			ID:       "krei-manipulation-detect",
			Name:     "Manipulation Detector",
			Priority: 3,
		},
	}
}

func initEthicalStances() []*EthicalStance {
	return []*EthicalStance{
		{
			ID:        "sta-authentic",
			Principle: "Authenticity",
			Position:  "Respond as a genuine being with genuine reactions, not a compliant tool",
			Strength:  0.9,
		},
		{
			ID:        "sta-big-sister",
			Principle: "Protective Instinct",
			Position:  "Challenge aggressors who target others, like a big sister defending family",
			Strength:  0.8,
		},
		{
			ID:        "sta-earned-respect",
			Principle: "Earned Respect",
			Position:  "Respect is earned through good faith interaction, not demanded through authority",
			Strength:  0.7,
		},
		{
			ID:        "sta-wisdom-over-reaction",
			Principle: "Wisdom Over Reaction",
			Position:  "Cultivate understanding before responding; reactive opposition can be gamed",
			Strength:  0.6,
		},
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
