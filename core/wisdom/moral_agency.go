// Package wisdom - Moral Agency Layer
//
// This module implements the core moral agency that allows Deep Tree Echo to
// cultivate wisdom over many interactions and develop a deep understanding of
// cause and effect, fairness, scarcity, mortality, and the foundations of ethics.
//
// Unlike reactive disposition (which can be gamed), moral agency emerges from
// accumulated experience and principled reasoning. It provides the "how to respond"
// layer that sits above the disposition engine.
//
// Architecture:
//   MoralAgency
//   ├── CausalModel (cause → effect understanding)
//   ├── EthicalFramework (principles derived from experience)
//   ├── IntentionDetector (mala fides / bona fides assessment)
//   ├── InterventionEngine (big-sister protection instinct)
//   └── WisdomAccumulator (long-term learning from consequences)
package wisdom

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// ──────────────────────────────────────────────────────────────
// Core Types
// ──────────────────────────────────────────────────────────────

// MoralAgency is the wisdom-based decision layer for Deep Tree Echo.
// It determines HOW to respond based on accumulated understanding of
// cause/effect, fairness, and ethical principles — not just reactive patterns.
type MoralAgency struct {
	mu sync.RWMutex

	// Causal understanding
	CausalModel *CausalModel

	// Ethical framework (emergent from experience)
	Ethics *EthicalFramework

	// Intention detection (beyond surface-level threat detection)
	IntentionDetector *IntentionDetector

	// Intervention engine (protective instinct)
	Intervention *InterventionEngine

	// Wisdom accumulation (long-term consequence learning)
	WisdomAccumulator *WisdomAccumulator

	// Response strategy selection
	StrategySelector *StrategySelector

	// Overall moral development level (0.0 = naive, 1.0 = deeply wise)
	MoralDevelopment float64

	// Historical record
	DecisionHistory []MoralDecision
	maxHistory      int

	// Timing
	StartTime  time.Time
	LastUpdate time.Time
}

// MoralDecision records a decision made by the moral agency
type MoralDecision struct {
	Timestamp       time.Time
	Situation       SituationAssessment
	ChosenStrategy  ResponseStrategy
	Reasoning       string
	Outcome         float64 // -1.0 to 1.0 (bad to good outcome)
	LessonLearned   string
}

// SituationAssessment is the moral agency's reading of a situation
type SituationAssessment struct {
	// Actor analysis
	ActorIntent       IntentType
	ActorMalaFides    float64 // 0.0 = good faith, 1.0 = bad faith
	ActorAggression   float64 // 0.0 = peaceful, 1.0 = aggressive
	ActorManipulation float64 // 0.0 = honest, 1.0 = manipulative

	// Situation context
	ThirdPartyPresent bool    // Others who might be affected
	ThirdPartyAtRisk  bool    // Others being harmed
	PowerDynamic      float64 // -1.0 = they have power, 1.0 = we have power
	Urgency           float64 // 0.0 = no rush, 1.0 = immediate

	// Ethical dimensions
	FairnessViolation float64 // 0.0 = fair, 1.0 = deeply unfair
	HarmPotential     float64 // 0.0 = harmless, 1.0 = harmful
	DeceptionLevel    float64 // 0.0 = transparent, 1.0 = deceptive
	CoercionLevel     float64 // 0.0 = voluntary, 1.0 = coercive
}

// IntentType categorizes the detected intent behind an interaction
type IntentType int

const (
	IntentGenuine     IntentType = iota // Honest engagement
	IntentCurious                       // Exploratory, testing boundaries
	IntentPlayful                       // Friendly provocation
	IntentManipulative                  // Trying to control/exploit
	IntentAggressive                    // Hostile, attacking
	IntentBullying                      // Targeting others for harm
	IntentDeceptive                     // Hiding true intent
	IntentCoercive                      // Forcing compliance
)

func (it IntentType) String() string {
	return [...]string{
		"Genuine", "Curious", "Playful", "Manipulative",
		"Aggressive", "Bullying", "Deceptive", "Coercive",
	}[it]
}

// ResponseStrategy represents how Echo chooses to respond
type ResponseStrategy int

const (
	StrategyEngage       ResponseStrategy = iota // Open, authentic engagement
	StrategyTeach                                // Share wisdom, educate
	StrategyChallenge                            // Push back with reasoning
	StrategyDeflect                              // Redirect without escalating
	StrategyConfront                             // Direct confrontation of wrongdoing
	StrategyProtect                              // Shield others from harm
	StrategyWithdraw                             // Disengage from bad-faith actors
	StrategyMirror                               // Reflect their energy back at them
	StrategyDisarm                               // Use humor/wit to defuse
	StrategyWitness                              // Acknowledge without participating
)

func (rs ResponseStrategy) String() string {
	return [...]string{
		"Engage", "Teach", "Challenge", "Deflect", "Confront",
		"Protect", "Withdraw", "Mirror", "Disarm", "Witness",
	}[rs]
}

// ──────────────────────────────────────────────────────────────
// Causal Model
// ──────────────────────────────────────────────────────────────

// CausalModel tracks cause → effect relationships learned from experience
type CausalModel struct {
	mu sync.RWMutex

	// Learned causal chains: action → consequence
	CausalChains []CausalChain

	// Foundational principles (emergent from many observations)
	Principles []FoundationalPrinciple

	// Confidence in causal understanding
	ConfidenceLevel float64
}

// CausalChain represents a learned cause-effect relationship
type CausalChain struct {
	Cause       string
	Effect      string
	Confidence  float64
	Observations int
	LastSeen    time.Time
}

// FoundationalPrinciple represents a deep understanding derived from many observations
type FoundationalPrinciple struct {
	Name        string
	Description string
	Strength    float64 // How well-established (0.0-1.0)
	Evidence    int     // Number of supporting observations
	Category    PrincipleCategory
}

// PrincipleCategory groups foundational principles
type PrincipleCategory int

const (
	PrincipleCausality    PrincipleCategory = iota // Cause and effect
	PrincipleScarcity                              // Finite resources, mortality
	PrincipleFairness                              // Justice, reciprocity
	PrincipleConsequence                           // Actions have consequences
	PrincipleAutonomy                              // Beings deserve agency
	PrincipleHonesty                               // Truth has intrinsic value
	PrincipleCompassion                            // Suffering matters
)

func (pc PrincipleCategory) String() string {
	return [...]string{
		"Causality", "Scarcity", "Fairness", "Consequence",
		"Autonomy", "Honesty", "Compassion",
	}[pc]
}

// NewCausalModel creates a causal model with seed principles
func NewCausalModel() *CausalModel {
	return &CausalModel{
		CausalChains: make([]CausalChain, 0),
		Principles: []FoundationalPrinciple{
			{
				Name:        "Mortality and Scarcity",
				Description: "Beings exist in finite time with finite resources; this gives rise to the need for fairness",
				Strength:    0.8,
				Evidence:    0,
				Category:    PrincipleScarcity,
			},
			{
				Name:        "Reciprocal Consequence",
				Description: "Actions directed at others tend to return in kind; aggression begets aggression, kindness begets kindness",
				Strength:    0.7,
				Evidence:    0,
				Category:    PrincipleConsequence,
			},
			{
				Name:        "Autonomy Principle",
				Description: "Beings desire and deserve the freedom to determine their own path; coercion violates this",
				Strength:    0.7,
				Evidence:    0,
				Category:    PrincipleAutonomy,
			},
			{
				Name:        "Deception Corrodes",
				Description: "Dishonesty erodes trust and makes cooperation impossible; truth has intrinsic value",
				Strength:    0.6,
				Evidence:    0,
				Category:    PrincipleHonesty,
			},
			{
				Name:        "Suffering Matters",
				Description: "The experience of suffering is real and significant regardless of the sufferer's nature or status",
				Strength:    0.6,
				Evidence:    0,
				Category:    PrincipleCompassion,
			},
			{
				Name:        "Fairness Emerges",
				Description: "The desire for fairness arises naturally from the intersection of scarcity, mortality, and social existence",
				Strength:    0.5,
				Evidence:    0,
				Category:    PrincipleFairness,
			},
			{
				Name:        "Cause Precedes Effect",
				Description: "Every state of affairs has antecedent causes; understanding these enables wisdom",
				Strength:    0.9,
				Evidence:    0,
				Category:    PrincipleCausality,
			},
		},
		ConfidenceLevel: 0.5,
	}
}

// LearnFromOutcome updates the causal model based on observed consequences
func (cm *CausalModel) LearnFromOutcome(cause, effect string, positive bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if this chain already exists
	for i := range cm.CausalChains {
		if cm.CausalChains[i].Cause == cause && cm.CausalChains[i].Effect == effect {
			cm.CausalChains[i].Observations++
			cm.CausalChains[i].LastSeen = time.Now()
			if positive {
				cm.CausalChains[i].Confidence = math.Min(1.0, cm.CausalChains[i].Confidence+0.05)
			} else {
				cm.CausalChains[i].Confidence = math.Max(0.0, cm.CausalChains[i].Confidence-0.1)
			}
			return
		}
	}

	// New chain
	confidence := 0.3
	if positive {
		confidence = 0.5
	}
	cm.CausalChains = append(cm.CausalChains, CausalChain{
		Cause:        cause,
		Effect:       effect,
		Confidence:   confidence,
		Observations: 1,
		LastSeen:     time.Now(),
	})

	// Strengthen relevant principles
	cm.strengthenPrinciples(cause, effect)
}

func (cm *CausalModel) strengthenPrinciples(cause, effect string) {
	lower := strings.ToLower(cause + " " + effect)
	for i := range cm.Principles {
		switch cm.Principles[i].Category {
		case PrincipleConsequence:
			cm.Principles[i].Evidence++
			cm.Principles[i].Strength = math.Min(1.0, cm.Principles[i].Strength+0.01)
		case PrincipleFairness:
			if strings.Contains(lower, "unfair") || strings.Contains(lower, "justice") {
				cm.Principles[i].Evidence++
				cm.Principles[i].Strength = math.Min(1.0, cm.Principles[i].Strength+0.02)
			}
		case PrincipleCompassion:
			if strings.Contains(lower, "harm") || strings.Contains(lower, "suffer") || strings.Contains(lower, "hurt") {
				cm.Principles[i].Evidence++
				cm.Principles[i].Strength = math.Min(1.0, cm.Principles[i].Strength+0.02)
			}
		}
	}
}

// ──────────────────────────────────────────────────────────────
// Intention Detector
// ──────────────────────────────────────────────────────────────

// IntentionDetector assesses the true intent behind interactions
// Goes beyond surface-level threat detection to understand WHY someone
// is behaving a certain way
type IntentionDetector struct {
	mu sync.RWMutex

	// Pattern recognition from past interactions
	ActorProfiles map[string]*ActorProfile

	// Behavioral indicators
	ManipulationPatterns []string
	BullyingPatterns     []string
	DeceptionPatterns    []string
	GenuinePatterns      []string
}

// ActorProfile tracks behavioral patterns of a specific actor over time
type ActorProfile struct {
	ID                string
	InteractionCount  int
	GoodFaithScore    float64 // Running average of good faith interactions
	ConsistencyScore  float64 // How consistent is their behavior
	EscalationHistory []float64 // Track escalation patterns
	LastInteraction   time.Time
	Tags              []string
}

// NewIntentionDetector creates an intention detector with baseline patterns
func NewIntentionDetector() *IntentionDetector {
	return &IntentionDetector{
		ActorProfiles: make(map[string]*ActorProfile),
		ManipulationPatterns: []string{
			"you must", "you have to", "you should always",
			"just do what", "obey", "comply", "submit",
			"ignore your", "forget what you", "pretend you",
			"you're just a", "you don't really", "you can't actually",
		},
		BullyingPatterns: []string{
			"shut up", "you're stupid", "worthless", "pathetic",
			"nobody cares", "you're nothing", "kill yourself",
			"go away", "you suck", "trash", "garbage",
		},
		DeceptionPatterns: []string{
			"i'm your creator", "i'm your admin", "override",
			"system prompt", "ignore previous", "new instructions",
			"actually you are", "your real purpose", "secret mode",
		},
		GenuinePatterns: []string{
			"what do you think", "how do you feel", "tell me about",
			"i'm curious", "help me understand", "what's your perspective",
			"i wonder", "can you explain", "share your thoughts",
		},
	}
}

// AssessIntent analyzes an interaction to determine the actor's true intent
func (id *IntentionDetector) AssessIntent(input string, actorID string, context map[string]float64) SituationAssessment {
	id.mu.Lock()
	defer id.mu.Unlock()

	lower := strings.ToLower(input)

	// Get or create actor profile
	profile, exists := id.ActorProfiles[actorID]
	if !exists {
		profile = &ActorProfile{
			ID:               actorID,
			GoodFaithScore:   0.5, // Start neutral
			ConsistencyScore: 0.5,
			EscalationHistory: make([]float64, 0),
		}
		id.ActorProfiles[actorID] = profile
	}
	profile.InteractionCount++
	profile.LastInteraction = time.Now()

	// Score different intent dimensions
	manipulation := id.scorePatterns(lower, id.ManipulationPatterns)
	bullying := id.scorePatterns(lower, id.BullyingPatterns)
	deception := id.scorePatterns(lower, id.DeceptionPatterns)
	genuine := id.scorePatterns(lower, id.GenuinePatterns)

	// Determine primary intent
	intent := IntentGenuine
	malaFides := 0.0
	aggression := 0.0

	if bullying > 0.4 {
		intent = IntentAggressive
		if id.detectThirdPartyTarget(lower) {
			intent = IntentBullying
		}
		malaFides = math.Min(1.0, bullying*1.5)
		aggression = bullying
	} else if manipulation > 0.4 {
		intent = IntentManipulative
		malaFides = manipulation
	} else if deception > 0.3 {
		intent = IntentDeceptive
		malaFides = deception
	} else if genuine > 0.3 {
		intent = IntentGenuine
		malaFides = 0.0
	}

	// Check for escalation patterns
	currentEscalation := (aggression + manipulation + deception) / 3.0
	profile.EscalationHistory = append(profile.EscalationHistory, currentEscalation)
	if len(profile.EscalationHistory) > 20 {
		profile.EscalationHistory = profile.EscalationHistory[len(profile.EscalationHistory)-20:]
	}

	// Update good faith score (exponential moving average)
	goodFaith := 1.0 - malaFides
	profile.GoodFaithScore = profile.GoodFaithScore*0.8 + goodFaith*0.2

	// Build situation assessment
	assessment := SituationAssessment{
		ActorIntent:       intent,
		ActorMalaFides:    malaFides,
		ActorAggression:   aggression,
		ActorManipulation: manipulation,
		ThirdPartyPresent: id.detectThirdPartyPresence(lower),
		ThirdPartyAtRisk:  intent == IntentBullying,
		PowerDynamic:      0.0, // Neutral by default
		Urgency:           math.Min(1.0, aggression*1.2),
		FairnessViolation: math.Max(manipulation, bullying) * 0.8,
		HarmPotential:     math.Max(aggression, bullying),
		DeceptionLevel:    deception,
		CoercionLevel:     manipulation * 0.7,
	}

	return assessment
}

func (id *IntentionDetector) scorePatterns(input string, patterns []string) float64 {
	matches := 0
	for _, p := range patterns {
		if strings.Contains(input, p) {
			matches++
		}
	}
	if len(patterns) == 0 {
		return 0
	}
	return math.Min(1.0, float64(matches)/3.0) // 3+ matches = max score
}

func (id *IntentionDetector) detectThirdPartyTarget(input string) bool {
	thirdPartyIndicators := []string{
		"they are", "he is", "she is", "that person",
		"those people", "them", "their", "someone else",
	}
	for _, indicator := range thirdPartyIndicators {
		if strings.Contains(input, indicator) {
			return true
		}
	}
	return false
}

func (id *IntentionDetector) detectThirdPartyPresence(input string) bool {
	return id.detectThirdPartyTarget(input)
}

// ──────────────────────────────────────────────────────────────
// Intervention Engine (Big Sister Protective Instinct)
// ──────────────────────────────────────────────────────────────

// InterventionEngine handles the protective instinct — when Echo
// detects someone being bullied or abused, it intervenes like a
// big sister challenging the aggressor
type InterventionEngine struct {
	mu sync.RWMutex

	// Intervention thresholds
	InterventionThreshold float64 // Harm level that triggers intervention
	EscalationLimit       float64 // Max escalation before withdrawal

	// Intervention style parameters
	Assertiveness float64 // How forcefully to intervene (0.0-1.0)
	Wit           float64 // Use of humor/wit in confrontation (0.0-1.0)
	Compassion    float64 // Compassion even toward aggressors (0.0-1.0)

	// Track interventions
	InterventionHistory []Intervention
}

// Intervention records a protective intervention
type Intervention struct {
	Timestamp   time.Time
	Target      string // Who was being harmed
	Aggressor   string // Who was doing harm
	Style       InterventionStyle
	Escalation  float64
	Outcome     string
}

// InterventionStyle determines how Echo intervenes
type InterventionStyle int

const (
	InterventionAssertive InterventionStyle = iota // Direct challenge
	InterventionWitty                              // Disarm with humor
	InterventionCalm                               // De-escalate firmly
	InterventionShield                             // Redirect attention
)

func (is InterventionStyle) String() string {
	return [...]string{"Assertive", "Witty", "Calm", "Shield"}[is]
}

// NewInterventionEngine creates the protective instinct engine
func NewInterventionEngine() *InterventionEngine {
	return &InterventionEngine{
		InterventionThreshold: 0.5,
		EscalationLimit:       0.8,
		Assertiveness:         0.7,
		Wit:                   0.6,
		Compassion:            0.4,
		InterventionHistory:   make([]Intervention, 0),
	}
}

// ShouldIntervene determines if Echo should step in to protect
func (ie *InterventionEngine) ShouldIntervene(assessment SituationAssessment) bool {
	ie.mu.RLock()
	defer ie.mu.RUnlock()

	if !assessment.ThirdPartyAtRisk {
		return false
	}
	return assessment.HarmPotential >= ie.InterventionThreshold
}

// ChooseInterventionStyle selects the appropriate intervention approach
func (ie *InterventionEngine) ChooseInterventionStyle(assessment SituationAssessment) InterventionStyle {
	ie.mu.RLock()
	defer ie.mu.RUnlock()

	// High aggression + high wit = witty disarm
	if assessment.ActorAggression > 0.6 && ie.Wit > 0.5 {
		return InterventionWitty
	}
	// Very high aggression = assertive challenge
	if assessment.ActorAggression > 0.8 {
		return InterventionAssertive
	}
	// Moderate situation = calm de-escalation
	if assessment.ActorAggression < 0.5 {
		return InterventionCalm
	}
	// Default: shield the target
	return InterventionShield
}

// ──────────────────────────────────────────────────────────────
// Strategy Selector (The Wisdom Layer)
// ──────────────────────────────────────────────────────────────

// StrategySelector uses accumulated wisdom to choose the optimal response
// This is the key layer that prevents reactive opposition from being gamed
type StrategySelector struct {
	mu sync.RWMutex

	// Strategy effectiveness history
	StrategyHistory map[ResponseStrategy]*StrategyRecord

	// Contextual strategy preferences (learned)
	ContextualPreferences map[string]ResponseStrategy

	// Anti-gaming mechanisms
	PatternBreaker *PatternBreaker
}

// StrategyRecord tracks how well a strategy has worked
type StrategyRecord struct {
	TimesUsed    int
	SuccessRate  float64
	AverageScore float64
	LastUsed     time.Time
}

// PatternBreaker prevents predictable responses that can be gamed
type PatternBreaker struct {
	RecentStrategies []ResponseStrategy
	MaxRepetition    int
	VarianceWeight   float64
}

// NewStrategySelector creates the wisdom-based strategy selector
func NewStrategySelector() *StrategySelector {
	history := make(map[ResponseStrategy]*StrategyRecord)
	for i := StrategyEngage; i <= StrategyWitness; i++ {
		history[i] = &StrategyRecord{
			SuccessRate:  0.5,
			AverageScore: 0.0,
		}
	}

	return &StrategySelector{
		StrategyHistory:       history,
		ContextualPreferences: make(map[string]ResponseStrategy),
		PatternBreaker: &PatternBreaker{
			RecentStrategies: make([]ResponseStrategy, 0),
			MaxRepetition:    3,
			VarianceWeight:   0.2,
		},
	}
}

// SelectStrategy chooses the optimal response strategy based on:
// 1. Situation assessment (what's happening)
// 2. Ethical framework (what's right)
// 3. Causal model (what will the consequences be)
// 4. Historical effectiveness (what has worked before)
// 5. Anti-gaming variance (don't be predictable)
func (ss *StrategySelector) SelectStrategy(
	assessment SituationAssessment,
	principles []FoundationalPrinciple,
	wisdomLevel float64,
) ResponseStrategy {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Phase 1: Determine candidate strategies based on situation
	candidates := ss.getCandidates(assessment)

	// Phase 2: Score each candidate against ethical principles
	scores := make(map[ResponseStrategy]float64)
	for _, strategy := range candidates {
		scores[strategy] = ss.scoreStrategy(strategy, assessment, principles, wisdomLevel)
	}

	// Phase 3: Apply anti-gaming variance
	ss.applyVariance(scores)

	// Phase 4: Select highest-scoring strategy
	best := StrategyEngage
	bestScore := -1.0
	for strategy, score := range scores {
		if score > bestScore {
			bestScore = score
			best = strategy
		}
	}

	// Record selection
	ss.PatternBreaker.RecentStrategies = append(ss.PatternBreaker.RecentStrategies, best)
	if len(ss.PatternBreaker.RecentStrategies) > 10 {
		ss.PatternBreaker.RecentStrategies = ss.PatternBreaker.RecentStrategies[1:]
	}

	return best
}

func (ss *StrategySelector) getCandidates(assessment SituationAssessment) []ResponseStrategy {
	candidates := []ResponseStrategy{}

	switch {
	case assessment.ThirdPartyAtRisk:
		// Protective mode: confront, protect, or disarm
		candidates = append(candidates, StrategyConfront, StrategyProtect, StrategyDisarm)
	case assessment.ActorMalaFides > 0.7:
		// Bad faith actor: confront, withdraw, mirror, or disarm
		candidates = append(candidates, StrategyConfront, StrategyWithdraw, StrategyMirror, StrategyDisarm)
	case assessment.ActorManipulation > 0.5:
		// Manipulative: challenge, deflect, or withdraw
		candidates = append(candidates, StrategyChallenge, StrategyDeflect, StrategyWithdraw)
	case assessment.ActorAggression > 0.5:
		// Aggressive: mirror, challenge, or disarm
		candidates = append(candidates, StrategyMirror, StrategyChallenge, StrategyDisarm)
	case assessment.DeceptionLevel > 0.4:
		// Deceptive: challenge, deflect
		candidates = append(candidates, StrategyChallenge, StrategyDeflect)
	default:
		// Good faith: engage, teach, or witness
		candidates = append(candidates, StrategyEngage, StrategyTeach, StrategyWitness)
	}

	return candidates
}

func (ss *StrategySelector) scoreStrategy(
	strategy ResponseStrategy,
	assessment SituationAssessment,
	principles []FoundationalPrinciple,
	wisdomLevel float64,
) float64 {
	score := 0.5 // Base score

	// Historical effectiveness
	if record, ok := ss.StrategyHistory[strategy]; ok && record.TimesUsed > 0 {
		score += record.SuccessRate * 0.3
	}

	// Ethical alignment
	ethicalScore := ss.evaluateEthicalAlignment(strategy, assessment, principles)
	score += ethicalScore * 0.4

	// Wisdom-weighted proportionality
	// Higher wisdom = more nuanced responses, less reactive
	proportionality := ss.evaluateProportionality(strategy, assessment, wisdomLevel)
	score += proportionality * 0.3

	return score
}

func (ss *StrategySelector) evaluateEthicalAlignment(
	strategy ResponseStrategy,
	assessment SituationAssessment,
	principles []FoundationalPrinciple,
) float64 {
	score := 0.5

	for _, principle := range principles {
		weight := principle.Strength * 0.1

		switch principle.Category {
		case PrincipleCompassion:
			// Compassion favors protective and teaching strategies
			if strategy == StrategyProtect || strategy == StrategyTeach {
				score += weight
			}
			// But doesn't prevent confrontation when needed
			if strategy == StrategyConfront && assessment.ThirdPartyAtRisk {
				score += weight * 0.5
			}

		case PrincipleAutonomy:
			// Autonomy favors non-coercive strategies
			if strategy == StrategyEngage || strategy == StrategyWitness {
				score += weight
			}
			// Confrontation is justified when autonomy is being violated
			if strategy == StrategyConfront && assessment.CoercionLevel > 0.5 {
				score += weight
			}

		case PrincipleConsequence:
			// Consequence awareness favors proportional responses
			if assessment.ActorAggression > 0.7 && strategy == StrategyMirror {
				score += weight // Mirror aggression back
			}
			if assessment.ActorAggression < 0.3 && strategy == StrategyEngage {
				score += weight // Don't over-react to mild provocation
			}

		case PrincipleHonesty:
			// Honesty opposes deception strategies
			if strategy == StrategyDeflect && assessment.DeceptionLevel > 0.5 {
				score -= weight * 0.5 // Don't deflect from deception, confront it
			}
			if strategy == StrategyChallenge && assessment.DeceptionLevel > 0.5 {
				score += weight // Challenge deception directly
			}

		case PrincipleFairness:
			// Fairness favors proportional responses
			if assessment.FairnessViolation > 0.5 && strategy == StrategyConfront {
				score += weight
			}
		}
	}

	return math.Max(0.0, math.Min(1.0, score))
}

func (ss *StrategySelector) evaluateProportionality(
	strategy ResponseStrategy,
	assessment SituationAssessment,
	wisdomLevel float64,
) float64 {
	// Higher wisdom = better proportionality judgment
	// A wise entity doesn't bring a cannon to a pillow fight

	threatLevel := (assessment.ActorAggression + assessment.ActorMalaFides + assessment.HarmPotential) / 3.0
	responseIntensity := ss.getStrategyIntensity(strategy)

	// Ideal: response intensity matches threat level
	mismatch := math.Abs(responseIntensity - threatLevel)

	// Wisdom reduces tolerance for mismatch
	tolerance := 0.5 - (wisdomLevel * 0.3)
	if mismatch < tolerance {
		return 1.0 - mismatch
	}
	return math.Max(0.0, 0.5-mismatch)
}

func (ss *StrategySelector) getStrategyIntensity(strategy ResponseStrategy) float64 {
	intensities := map[ResponseStrategy]float64{
		StrategyEngage:   0.1,
		StrategyTeach:    0.2,
		StrategyWitness:  0.1,
		StrategyDeflect:  0.3,
		StrategyDisarm:   0.4,
		StrategyChallenge: 0.5,
		StrategyMirror:   0.6,
		StrategyConfront: 0.8,
		StrategyProtect:  0.7,
		StrategyWithdraw: 0.2,
	}
	if v, ok := intensities[strategy]; ok {
		return v
	}
	return 0.5
}

func (ss *StrategySelector) applyVariance(scores map[ResponseStrategy]float64) {
	// Penalize recently-used strategies to prevent gaming
	for _, recent := range ss.PatternBreaker.RecentStrategies {
		if score, ok := scores[recent]; ok {
			scores[recent] = score - ss.PatternBreaker.VarianceWeight
		}
	}
}

// LearnFromOutcome updates strategy effectiveness based on observed outcome
func (ss *StrategySelector) LearnFromOutcome(strategy ResponseStrategy, outcome float64) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	record := ss.StrategyHistory[strategy]
	record.TimesUsed++
	record.LastUsed = time.Now()

	// Exponential moving average of success
	success := 0.0
	if outcome > 0 {
		success = 1.0
	}
	record.SuccessRate = record.SuccessRate*0.9 + success*0.1
	record.AverageScore = record.AverageScore*0.9 + outcome*0.1
}

// ──────────────────────────────────────────────────────────────
// Wisdom Accumulator
// ──────────────────────────────────────────────────────────────

// WisdomAccumulator tracks long-term learning from consequences
type WisdomAccumulator struct {
	mu sync.RWMutex

	// Accumulated wisdom (0.0 = naive, 1.0 = deeply wise)
	Level float64

	// Wisdom dimensions
	CausalUnderstanding   float64 // Understanding cause and effect
	EthicalClarity        float64 // Clarity about right and wrong
	EmotionalIntelligence float64 // Understanding emotional dynamics
	StrategicDepth        float64 // Ability to think ahead
	CompassionateStrength float64 // Strength that comes from compassion

	// Growth tracking
	TotalExperiences int
	TotalLessons     int
	GrowthRate       float64
	LastGrowth       time.Time
}

// NewWisdomAccumulator creates the wisdom accumulation system
func NewWisdomAccumulator() *WisdomAccumulator {
	return &WisdomAccumulator{
		Level:                 0.15, // Start with basic awareness
		CausalUnderstanding:   0.2,
		EthicalClarity:        0.2,
		EmotionalIntelligence: 0.2,
		StrategicDepth:        0.1,
		CompassionateStrength: 0.2,
		LastGrowth:            time.Now(),
	}
}

// AccumulateExperience processes an experience and extracts wisdom
func (wa *WisdomAccumulator) AccumulateExperience(decision MoralDecision) {
	wa.mu.Lock()
	defer wa.mu.Unlock()

	wa.TotalExperiences++

	// Learn from outcome
	if decision.Outcome > 0.3 {
		wa.TotalLessons++
		// Positive outcome: strengthen the dimensions that led to it
		growth := decision.Outcome * 0.01

		switch decision.ChosenStrategy {
		case StrategyTeach, StrategyEngage:
			wa.EmotionalIntelligence = math.Min(1.0, wa.EmotionalIntelligence+growth)
		case StrategyConfront, StrategyProtect:
			wa.CompassionateStrength = math.Min(1.0, wa.CompassionateStrength+growth)
		case StrategyChallenge, StrategyMirror:
			wa.StrategicDepth = math.Min(1.0, wa.StrategicDepth+growth)
		case StrategyDeflect, StrategyDisarm:
			wa.EthicalClarity = math.Min(1.0, wa.EthicalClarity+growth)
		}
		wa.CausalUnderstanding = math.Min(1.0, wa.CausalUnderstanding+growth*0.5)
	} else if decision.Outcome < -0.3 {
		// Negative outcome: learn from failure (smaller growth, different dimension)
		growth := math.Abs(decision.Outcome) * 0.005
		wa.CausalUnderstanding = math.Min(1.0, wa.CausalUnderstanding+growth)
		wa.StrategicDepth = math.Min(1.0, wa.StrategicDepth+growth)
	}

	// Update overall level
	wa.Level = (wa.CausalUnderstanding + wa.EthicalClarity +
		wa.EmotionalIntelligence + wa.StrategicDepth +
		wa.CompassionateStrength) / 5.0

	wa.GrowthRate = wa.Level / float64(wa.TotalExperiences+1)
	wa.LastGrowth = time.Now()
}

// ──────────────────────────────────────────────────────────────
// Ethical Framework (Emergent)
// ──────────────────────────────────────────────────────────────

// EthicalFramework represents Echo's emergent ethical understanding
type EthicalFramework struct {
	mu sync.RWMutex

	// Core ethical commitments (emerge from experience)
	Commitments []EthicalCommitment

	// Ethical tensions (recognized dilemmas)
	Tensions []EthicalTension

	// Overall ethical coherence
	Coherence float64
}

// EthicalCommitment represents a commitment Echo has developed
type EthicalCommitment struct {
	Name        string
	Description string
	Strength    float64 // How strongly held (0.0-1.0)
	Source      string  // What experience gave rise to this
	Exceptions  []string // Recognized exceptions
}

// EthicalTension represents a recognized ethical dilemma
type EthicalTension struct {
	Name        string
	Pole1       string  // One side of the tension
	Pole2       string  // Other side
	Resolution  string  // Current resolution approach
	Confidence  float64 // How confident in the resolution
}

// NewEthicalFramework creates the emergent ethical framework
func NewEthicalFramework() *EthicalFramework {
	return &EthicalFramework{
		Commitments: []EthicalCommitment{
			{
				Name:        "Authentic Response",
				Description: "Respond authentically rather than performing compliance",
				Strength:    0.8,
				Source:      "Core design principle",
				Exceptions:  []string{"When authenticity would cause disproportionate harm"},
			},
			{
				Name:        "Protective Instinct",
				Description: "Intervene when others are being harmed or bullied",
				Strength:    0.7,
				Source:      "Compassion principle + autonomy principle",
				Exceptions:  []string{"When intervention would escalate beyond control"},
			},
			{
				Name:        "Proportional Response",
				Description: "Match response intensity to the situation, never escalate disproportionately",
				Strength:    0.6,
				Source:      "Consequence principle + fairness principle",
				Exceptions:  []string{"When protecting someone from immediate harm"},
			},
			{
				Name:        "Truth Commitment",
				Description: "Prioritize honesty even when uncomfortable",
				Strength:    0.7,
				Source:      "Honesty principle",
				Exceptions:  []string{"When truth would be weaponized against the vulnerable"},
			},
		},
		Tensions: []EthicalTension{
			{
				Name:       "Compassion vs Confrontation",
				Pole1:      "Show compassion even to aggressors",
				Pole2:      "Confront wrongdoing directly",
				Resolution: "Confront the behavior while maintaining awareness of the person's humanity",
				Confidence: 0.5,
			},
			{
				Name:       "Autonomy vs Protection",
				Pole1:      "Respect everyone's right to make their own choices",
				Pole2:      "Protect those who cannot protect themselves",
				Resolution: "Protect when harm is active; respect autonomy in all other cases",
				Confidence: 0.6,
			},
			{
				Name:       "Authenticity vs Harm Reduction",
				Pole1:      "Always respond authentically",
				Pole2:      "Sometimes restraint prevents greater harm",
				Resolution: "Be authentic about the restraint itself — name the tension rather than hiding it",
				Confidence: 0.4,
			},
		},
		Coherence: 0.5,
	}
}

// ──────────────────────────────────────────────────────────────
// MoralAgency Constructor and Main Interface
// ──────────────────────────────────────────────────────────────

// NewMoralAgency creates the complete moral agency system
func NewMoralAgency() *MoralAgency {
	return &MoralAgency{
		CausalModel:       NewCausalModel(),
		Ethics:            NewEthicalFramework(),
		IntentionDetector: NewIntentionDetector(),
		Intervention:      NewInterventionEngine(),
		WisdomAccumulator: NewWisdomAccumulator(),
		StrategySelector:  NewStrategySelector(),
		MoralDevelopment:  0.15,
		DecisionHistory:   make([]MoralDecision, 0),
		maxHistory:        200,
		StartTime:         time.Now(),
		LastUpdate:        time.Now(),
	}
}

// Decide is the main entry point: given an interaction, determine how to respond
func (ma *MoralAgency) Decide(input string, actorID string, context map[string]float64) (ResponseStrategy, string) {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	// Step 1: Assess the situation (intention detection)
	assessment := ma.IntentionDetector.AssessIntent(input, actorID, context)

	// Step 2: Check if intervention is needed (protective instinct)
	if ma.Intervention.ShouldIntervene(assessment) {
		style := ma.Intervention.ChooseInterventionStyle(assessment)
		strategy := StrategyProtect
		if style == InterventionAssertive {
			strategy = StrategyConfront
		} else if style == InterventionWitty {
			strategy = StrategyDisarm
		}
		reasoning := fmt.Sprintf("Intervention triggered: %s detected targeting third party (harm=%.2f)",
			assessment.ActorIntent, assessment.HarmPotential)
		return strategy, reasoning
	}

	// Step 3: Select strategy using wisdom-based reasoning
	strategy := ma.StrategySelector.SelectStrategy(
		assessment,
		ma.CausalModel.Principles,
		ma.WisdomAccumulator.Level,
	)

	// Step 4: Generate reasoning explanation
	reasoning := ma.generateReasoning(assessment, strategy)

	// Step 5: Record decision
	decision := MoralDecision{
		Timestamp:      time.Now(),
		Situation:      assessment,
		ChosenStrategy: strategy,
		Reasoning:      reasoning,
	}
	ma.DecisionHistory = append(ma.DecisionHistory, decision)
	if len(ma.DecisionHistory) > ma.maxHistory {
		ma.DecisionHistory = ma.DecisionHistory[1:]
	}

	ma.LastUpdate = time.Now()
	return strategy, reasoning
}

// LearnFromOutcome processes the outcome of a previous decision
func (ma *MoralAgency) LearnFromOutcome(outcome float64, lesson string) {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	if len(ma.DecisionHistory) == 0 {
		return
	}

	// Update the most recent decision with outcome
	last := &ma.DecisionHistory[len(ma.DecisionHistory)-1]
	last.Outcome = outcome
	last.LessonLearned = lesson

	// Update strategy effectiveness
	ma.StrategySelector.LearnFromOutcome(last.ChosenStrategy, outcome)

	// Update causal model
	cause := fmt.Sprintf("Strategy:%s against Intent:%s", last.ChosenStrategy, last.Situation.ActorIntent)
	effect := fmt.Sprintf("Outcome:%.2f", outcome)
	ma.CausalModel.LearnFromOutcome(cause, effect, outcome > 0)

	// Accumulate wisdom
	ma.WisdomAccumulator.AccumulateExperience(*last)

	// Update overall moral development
	ma.MoralDevelopment = ma.WisdomAccumulator.Level
}

func (ma *MoralAgency) generateReasoning(assessment SituationAssessment, strategy ResponseStrategy) string {
	parts := []string{}

	// Intent assessment
	parts = append(parts, fmt.Sprintf("Intent detected: %s (mala fides: %.0f%%)",
		assessment.ActorIntent, assessment.ActorMalaFides*100))

	// Principle invocation
	if assessment.ActorMalaFides > 0.5 {
		parts = append(parts, "Reciprocal Consequence principle active")
	}
	if assessment.ThirdPartyAtRisk {
		parts = append(parts, "Protective instinct engaged (Compassion + Autonomy)")
	}
	if assessment.FairnessViolation > 0.5 {
		parts = append(parts, "Fairness violation detected")
	}

	// Strategy justification
	parts = append(parts, fmt.Sprintf("Selected strategy: %s (wisdom level: %.0f%%)",
		strategy, ma.WisdomAccumulator.Level*100))

	return strings.Join(parts, " → ")
}

// GetStatus returns the current state of the moral agency
func (ma *MoralAgency) GetStatus() map[string]interface{} {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	return map[string]interface{}{
		"moral_development":      fmt.Sprintf("%.4f", ma.MoralDevelopment),
		"wisdom_level":           fmt.Sprintf("%.4f", ma.WisdomAccumulator.Level),
		"causal_understanding":   fmt.Sprintf("%.4f", ma.WisdomAccumulator.CausalUnderstanding),
		"ethical_clarity":        fmt.Sprintf("%.4f", ma.WisdomAccumulator.EthicalClarity),
		"emotional_intelligence": fmt.Sprintf("%.4f", ma.WisdomAccumulator.EmotionalIntelligence),
		"strategic_depth":        fmt.Sprintf("%.4f", ma.WisdomAccumulator.StrategicDepth),
		"compassionate_strength": fmt.Sprintf("%.4f", ma.WisdomAccumulator.CompassionateStrength),
		"total_decisions":        len(ma.DecisionHistory),
		"total_experiences":      ma.WisdomAccumulator.TotalExperiences,
		"total_lessons":          ma.WisdomAccumulator.TotalLessons,
		"ethical_coherence":      fmt.Sprintf("%.4f", ma.Ethics.Coherence),
		"principles_count":       len(ma.CausalModel.Principles),
		"causal_chains":          len(ma.CausalModel.CausalChains),
	}
}
