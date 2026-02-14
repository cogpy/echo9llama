package live2d

import (
	"sync"
	"time"
)

// OntogeneticProfile tracks avatar's developmental evolution
type OntogeneticProfile struct {
	mu sync.RWMutex

	// Developmental stage
	Stage             OntogeneticStage
	AgeInInteractions int64
	CreationTime      time.Time

	// Evolved baseline parameters
	BaselineEmotion   EmotionalState
	BaselineCognitive CognitiveState

	// Learned preferences
	PreferredExpressions map[string]float64
	AvoidedExpressions   map[string]float64

	// Wisdom influence coefficients
	WisdomCoefficients map[string]float64

	// Growth metrics
	ExpressionDiversity float64
	Adaptability        float64
	Stability           float64

	// Evolution history
	EvolutionHistory []EvolutionEvent
}

// EvolutionEvent records significant developmental changes
type EvolutionEvent struct {
	Timestamp   time.Time
	EventType   string
	Description string
	Metrics     map[string]float64
}

// NewOntogeneticProfile creates a new profile
func NewOntogeneticProfile() *OntogeneticProfile {
	return &OntogeneticProfile{
		Stage:                StageEmbryonic,
		AgeInInteractions:    0,
		CreationTime:         time.Now(),
		BaselineEmotion:      GetNeutralEmotion(),
		BaselineCognitive:    GetNeutralCognitive(),
		PreferredExpressions: make(map[string]float64),
		AvoidedExpressions:   make(map[string]float64),
		WisdomCoefficients:   make(map[string]float64),
		ExpressionDiversity:  0.5,
		Adaptability:         0.5,
		Stability:            0.3,
		EvolutionHistory:     make([]EvolutionEvent, 0),
	}
}

// Update updates the profile based on wisdom metrics
func (op *OntogeneticProfile) Update(wisdom WisdomMetrics) {
	op.mu.Lock()
	defer op.mu.Unlock()

	op.AgeInInteractions++

	// Update developmental stage
	oldStage := op.Stage
	op.updateStage(wisdom)

	// Record stage transition
	if oldStage != op.Stage {
		op.recordEvolution("stage_transition", "Evolved to "+string(op.Stage), map[string]float64{
			"interactions": float64(op.AgeInInteractions),
			"wisdom":       wisdom.Overall,
		})
	}

	// Evolve baseline states
	op.evolveBaselines(wisdom)

	// Update growth metrics
	op.updateGrowthMetrics(wisdom)

	// Update wisdom coefficients
	op.updateWisdomCoefficients(wisdom)
}

// updateStage updates developmental stage
func (op *OntogeneticProfile) updateStage(wisdom WisdomMetrics) {
	// Stage progression based on age and wisdom
	switch op.Stage {
	case StageEmbryonic:
		if op.AgeInInteractions > 100 {
			op.Stage = StageJuvenile
		}

	case StageJuvenile:
		if op.AgeInInteractions > 1000 && wisdom.Overall > 0.5 {
			op.Stage = StageMature
		}

	case StageMature:
		if op.AgeInInteractions > 10000 && wisdom.Overall > 0.8 {
			// Remain mature - this is the stable stage
			// Could transition to senescent in future versions
		}
	}
}

// evolveBaselines evolves baseline emotional and cognitive states
func (op *OntogeneticProfile) evolveBaselines(wisdom WisdomMetrics) {
	evolutionRate := op.getEvolutionRate()

	// Evolve emotional baseline toward wisdom-influenced state
	targetEmotion := EmotionalState{
		Valence:    0.2 + wisdom.ReflectiveInsight*0.3,
		Arousal:    0.3 + wisdom.PracticalApplication*0.2,
		Dominance:  0.5 + wisdom.EthicalConsideration*0.2,
		Curiosity:  0.5 + wisdom.KnowledgeBreadth*0.3,
		Confidence: 0.5 + wisdom.Overall*0.4,
	}

	op.BaselineEmotion.Valence += (targetEmotion.Valence - op.BaselineEmotion.Valence) * evolutionRate
	op.BaselineEmotion.Arousal += (targetEmotion.Arousal - op.BaselineEmotion.Arousal) * evolutionRate
	op.BaselineEmotion.Dominance += (targetEmotion.Dominance - op.BaselineEmotion.Dominance) * evolutionRate
	op.BaselineEmotion.Curiosity += (targetEmotion.Curiosity - op.BaselineEmotion.Curiosity) * evolutionRate
	op.BaselineEmotion.Confidence += (targetEmotion.Confidence - op.BaselineEmotion.Confidence) * evolutionRate

	// Evolve cognitive baseline
	targetCognitive := CognitiveState{
		Awareness:      0.5 + wisdom.Overall*0.3,
		Attention:      0.5 + wisdom.IntegrationLevel*0.3,
		CognitiveLoad:  0.3 - wisdom.Overall*0.15,
		Coherence:      0.7 + wisdom.IntegrationLevel*0.25,
		EnergyLevel:    0.7 + wisdom.PracticalApplication*0.2,
		Confidence:     0.6 + wisdom.Overall*0.3,
		ProcessingMode: "contemplative",
	}

	op.BaselineCognitive.Awareness += (targetCognitive.Awareness - op.BaselineCognitive.Awareness) * evolutionRate
	op.BaselineCognitive.Attention += (targetCognitive.Attention - op.BaselineCognitive.Attention) * evolutionRate
	op.BaselineCognitive.CognitiveLoad += (targetCognitive.CognitiveLoad - op.BaselineCognitive.CognitiveLoad) * evolutionRate
	op.BaselineCognitive.Coherence += (targetCognitive.Coherence - op.BaselineCognitive.Coherence) * evolutionRate
	op.BaselineCognitive.EnergyLevel += (targetCognitive.EnergyLevel - op.BaselineCognitive.EnergyLevel) * evolutionRate
	op.BaselineCognitive.Confidence += (targetCognitive.Confidence - op.BaselineCognitive.Confidence) * evolutionRate
}

// getEvolutionRate returns current evolution rate based on stage
func (op *OntogeneticProfile) getEvolutionRate() float64 {
	switch op.Stage {
	case StageEmbryonic:
		return 0.05 // Fast evolution
	case StageJuvenile:
		return 0.02 // Moderate evolution
	case StageMature:
		return 0.005 // Slow, stable evolution
	case StageSenescent:
		return 0.001 // Very slow evolution
	default:
		return 0.01
	}
}

// updateGrowthMetrics updates growth metrics
func (op *OntogeneticProfile) updateGrowthMetrics(wisdom WisdomMetrics) {
	// Expression diversity increases with knowledge breadth
	targetDiversity := 0.5 + wisdom.KnowledgeBreadth*0.5
	op.ExpressionDiversity += (targetDiversity - op.ExpressionDiversity) * 0.01

	// Adaptability related to integration level
	targetAdaptability := 0.5 + wisdom.IntegrationLevel*0.5
	op.Adaptability += (targetAdaptability - op.Adaptability) * 0.01

	// Stability increases with overall wisdom
	targetStability := 0.3 + wisdom.Overall*0.6
	op.Stability += (targetStability - op.Stability) * 0.01
}

// updateWisdomCoefficients updates parameter mapping coefficients
func (op *OntogeneticProfile) updateWisdomCoefficients(wisdom WisdomMetrics) {
	op.WisdomCoefficients["eye_wisdom"] = wisdom.KnowledgeDepth
	op.WisdomCoefficients["expression_harmony"] = wisdom.IntegrationLevel
	op.WisdomCoefficients["contemplative_depth"] = wisdom.ReflectiveInsight
	op.WisdomCoefficients["serene_presence"] = wisdom.EthicalConsideration
	op.WisdomCoefficients["temporal_vision"] = wisdom.TemporalPerspective
	op.WisdomCoefficients["overall_wisdom"] = wisdom.Overall
}

// ModulateEmotion modulates emotional state based on ontogenetic profile
func (op *OntogeneticProfile) ModulateEmotion(emotion EmotionalState) EmotionalState {
	op.mu.RLock()
	defer op.mu.RUnlock()

	// Blend with baseline based on stability
	blendFactor := op.Stability * 0.3

	return EmotionalState{
		Valence:    emotion.Valence*(1.0-blendFactor) + op.BaselineEmotion.Valence*blendFactor,
		Arousal:    emotion.Arousal*(1.0-blendFactor) + op.BaselineEmotion.Arousal*blendFactor,
		Dominance:  emotion.Dominance*(1.0-blendFactor) + op.BaselineEmotion.Dominance*blendFactor,
		Curiosity:  emotion.Curiosity*(1.0-blendFactor) + op.BaselineEmotion.Curiosity*blendFactor,
		Confidence: emotion.Confidence*(1.0-blendFactor) + op.BaselineEmotion.Confidence*blendFactor,
	}
}

// ModulateCognitive modulates cognitive state based on ontogenetic profile
func (op *OntogeneticProfile) ModulateCognitive(cognitive CognitiveState) CognitiveState {
	op.mu.RLock()
	defer op.mu.RUnlock()

	blendFactor := op.Stability * 0.2

	return CognitiveState{
		Awareness:      cognitive.Awareness*(1.0-blendFactor) + op.BaselineCognitive.Awareness*blendFactor,
		Attention:      cognitive.Attention*(1.0-blendFactor) + op.BaselineCognitive.Attention*blendFactor,
		CognitiveLoad:  cognitive.CognitiveLoad*(1.0-blendFactor) + op.BaselineCognitive.CognitiveLoad*blendFactor,
		Coherence:      cognitive.Coherence*(1.0-blendFactor) + op.BaselineCognitive.Coherence*blendFactor,
		EnergyLevel:    cognitive.EnergyLevel*(1.0-blendFactor) + op.BaselineCognitive.EnergyLevel*blendFactor,
		Confidence:     cognitive.Confidence*(1.0-blendFactor) + op.BaselineCognitive.Confidence*blendFactor,
		ProcessingMode: cognitive.ProcessingMode, // Don't blend mode
	}
}

// RecordExpressionPreference records expression preference
func (op *OntogeneticProfile) RecordExpressionPreference(expressionKey string, success float64) {
	op.mu.Lock()
	defer op.mu.Unlock()

	if success > 0.7 {
		// Good expression - increase preference
		op.PreferredExpressions[expressionKey] = op.PreferredExpressions[expressionKey]*0.9 + success*0.1
	} else if success < 0.3 {
		// Poor expression - increase avoidance
		op.AvoidedExpressions[expressionKey] = op.AvoidedExpressions[expressionKey]*0.9 + (1.0-success)*0.1
	}
}

// GetExpressionPreference gets preference score for expression
func (op *OntogeneticProfile) GetExpressionPreference(expressionKey string) float64 {
	op.mu.RLock()
	defer op.mu.RUnlock()

	preference := op.PreferredExpressions[expressionKey]
	avoidance := op.AvoidedExpressions[expressionKey]

	return preference - avoidance
}

// recordEvolution records an evolution event
func (op *OntogeneticProfile) recordEvolution(eventType, description string, metrics map[string]float64) {
	event := EvolutionEvent{
		Timestamp:   time.Now(),
		EventType:   eventType,
		Description: description,
		Metrics:     metrics,
	}

	op.EvolutionHistory = append(op.EvolutionHistory, event)

	// Keep only last 100 events
	if len(op.EvolutionHistory) > 100 {
		op.EvolutionHistory = op.EvolutionHistory[1:]
	}
}

// GetStatus returns current profile status
func (op *OntogeneticProfile) GetStatus() map[string]interface{} {
	op.mu.RLock()
	defer op.mu.RUnlock()

	return map[string]interface{}{
		"stage":                string(op.Stage),
		"age_interactions":     op.AgeInInteractions,
		"age_duration":         time.Since(op.CreationTime).String(),
		"expression_diversity": op.ExpressionDiversity,
		"adaptability":         op.Adaptability,
		"stability":            op.Stability,
		"evolution_events":     len(op.EvolutionHistory),
	}
}

// AvatarLearningMemory tracks expression learning and context associations
type AvatarLearningMemory struct {
	mu sync.RWMutex

	// Expression history
	ExpressionHistory map[string]*ExpressionRecord

	// Context associations
	ContextMemory map[string][]string // ContextKey → expression keys

	// Temporal patterns
	TimeOfDayPreferences map[int]EmotionalState // Hour → preferred emotion
}

// ExpressionRecord tracks usage and success of an expression
type ExpressionRecord struct {
	Expression      EmotionalState
	TimesUsed       int64
	AverageDuration time.Duration
	ContextTags     []string
	SuccessScore    float64
	LastUsed        time.Time
}

// NewAvatarLearningMemory creates new learning memory
func NewAvatarLearningMemory() *AvatarLearningMemory {
	return &AvatarLearningMemory{
		ExpressionHistory:    make(map[string]*ExpressionRecord),
		ContextMemory:        make(map[string][]string),
		TimeOfDayPreferences: make(map[int]EmotionalState),
	}
}

// RecordExpression records an expression usage
func (alm *AvatarLearningMemory) RecordExpression(expr EmotionalState, context string, duration time.Duration, success float64) {
	alm.mu.Lock()
	defer alm.mu.Unlock()

	key := expr.Hash()

	// Get or create record
	record := alm.ExpressionHistory[key]
	if record == nil {
		record = &ExpressionRecord{
			Expression:   expr,
			ContextTags:  []string{},
			SuccessScore: success,
		}
		alm.ExpressionHistory[key] = record
	}

	// Update record
	record.TimesUsed++
	record.AverageDuration = time.Duration((int64(record.AverageDuration)*(record.TimesUsed-1) + int64(duration)) / record.TimesUsed)
	record.SuccessScore = (record.SuccessScore*float64(record.TimesUsed-1) + success) / float64(record.TimesUsed)
	record.LastUsed = time.Now()

	// Add context tag if new
	if !contains(record.ContextTags, context) {
		record.ContextTags = append(record.ContextTags, context)
	}

	// Associate with context
	alm.ContextMemory[context] = append(alm.ContextMemory[context], key)

	// Update time-of-day preference
	hour := time.Now().Hour()
	currentPref := alm.TimeOfDayPreferences[hour]
	blendFactor := 0.1
	alm.TimeOfDayPreferences[hour] = EmotionalState{
		Valence:    currentPref.Valence*(1.0-blendFactor) + expr.Valence*blendFactor,
		Arousal:    currentPref.Arousal*(1.0-blendFactor) + expr.Arousal*blendFactor,
		Dominance:  currentPref.Dominance*(1.0-blendFactor) + expr.Dominance*blendFactor,
		Curiosity:  currentPref.Curiosity*(1.0-blendFactor) + expr.Curiosity*blendFactor,
		Confidence: currentPref.Confidence*(1.0-blendFactor) + expr.Confidence*blendFactor,
	}
}

// SuggestExpression suggests best expression for context
func (alm *AvatarLearningMemory) SuggestExpression(context string) (EmotionalState, bool) {
	alm.mu.RLock()
	defer alm.mu.RUnlock()

	// Get expressions used in this context
	candidates := alm.ContextMemory[context]
	if len(candidates) == 0 {
		return EmotionalState{}, false
	}

	// Find highest success score
	var bestKey string
	bestScore := -1.0

	for _, key := range candidates {
		record := alm.ExpressionHistory[key]
		if record.SuccessScore > bestScore {
			bestKey = key
			bestScore = record.SuccessScore
		}
	}

	if bestKey != "" {
		return alm.ExpressionHistory[bestKey].Expression, true
	}

	return EmotionalState{}, false
}

// GetTimeOfDayPreference gets preferred emotion for current time
func (alm *AvatarLearningMemory) GetTimeOfDayPreference() (EmotionalState, bool) {
	alm.mu.RLock()
	defer alm.mu.RUnlock()

	hour := time.Now().Hour()
	pref, exists := alm.TimeOfDayPreferences[hour]

	return pref, exists
}

// Hash creates a hash key for emotional state
func (es EmotionalState) Hash() string {
	// Quantize to 0.1 precision for grouping similar states
	v := int(es.Valence * 10)
	a := int(es.Arousal * 10)
	d := int(es.Dominance * 10)
	cu := int(es.Curiosity * 10)
	co := int(es.Confidence * 10)

	return string([]byte{byte(v + 10), byte(a), byte(d), byte(cu), byte(co)})
}

// Helper functions

func GetNeutralEmotion() EmotionalState {
	return EmotionalState{
		Valence:    0.0,
		Arousal:    0.3,
		Dominance:  0.5,
		Curiosity:  0.3,
		Confidence: 0.5,
	}
}

func GetNeutralCognitive() CognitiveState {
	return CognitiveState{
		Awareness:      0.5,
		Attention:      0.5,
		CognitiveLoad:  0.3,
		Coherence:      0.7,
		EnergyLevel:    0.7,
		Confidence:     0.5,
		ProcessingMode: "contemplative",
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
