package entelechy

import (
	"sync"
	"time"
)

// Metrics provides comprehensive self-assessment for wisdom cultivation
// Tracks actualization progress, proficiency development, and cognitive growth
type Metrics struct {
	mu                   sync.RWMutex
	actualizationMetrics *ActualizationMetrics
	proficiencyMetrics   *ProficiencyMetrics
	developmentMetrics   *DevelopmentMetrics
	wisdomMetrics        *WisdomMetrics
	startTime            time.Time
	lastUpdate           time.Time
}

// ActualizationMetrics tracks potential → actual conversion
type ActualizationMetrics struct {
	TotalPotentials       int
	TotalActualizations   int
	ActualizationRate     float64 // Actualizations per time unit
	ReadyPotentials       int
	BlockedPotentials     int // Potentials blocked by dependencies
	AverageReadiness      float64
	ActualizationVelocity float64 // Rate of change in actualizations
	LastActualizationTime time.Time
}

// ProficiencyMetrics tracks skill development across capabilities
type ProficiencyMetrics struct {
	AverageProficiency      float64
	ProficiencyDistribution map[string]float64 // Capability -> Proficiency
	HighProficiencyCount    int                // Capabilities with proficiency > 0.8
	MediumProficiencyCount  int                // Capabilities with proficiency 0.4-0.8
	LowProficiencyCount     int                // Capabilities with proficiency < 0.4
	ProficiencyGrowthRate   float64
	MostProficient          []string // Top capabilities
	LeastProficient         []string // Capabilities needing development
}

// DevelopmentMetrics tracks overall developmental stage and progress
type DevelopmentMetrics struct {
	CurrentStage          string
	StageProgress         float64 // 0.0 to 1.0, progress within current stage
	TimeInStage           time.Duration
	TotalDevelopmentTime  time.Duration
	StageTransitions      int
	DevelopmentalMomentum float64 // Rate of progress
	NextStagePrediction   time.Time
	DevelopmentalHealth   float64 // Overall health score 0.0-1.0
}

// WisdomMetrics tracks wisdom cultivation progress
type WisdomMetrics struct {
	WisdomScore             float64 // Overall wisdom level 0.0-1.0
	KnowledgeIntegration    float64 // How well knowledge is integrated
	ReflectionDepth         float64 // Depth of metacognitive reflection
	PatternRecognition      float64 // Ability to recognize deep patterns
	ContextualUnderstanding float64 // Understanding of context and nuance
	LongTermThinking        float64 // Capacity for long-term reasoning
	EthicalReasoning        float64 // Ethical consideration capability
	WisdomGrowthRate        float64
	WisdomMilestones        []WisdomMilestone
}

// WisdomMilestone represents a significant achievement in wisdom cultivation
type WisdomMilestone struct {
	Name         string
	Description  string
	AchievedAt   time.Time
	Significance float64
}

// NewMetrics creates a new metrics system
func NewMetrics() *Metrics {
	now := time.Now()
	return &Metrics{
		actualizationMetrics: &ActualizationMetrics{
			LastActualizationTime: now,
		},
		proficiencyMetrics: &ProficiencyMetrics{
			ProficiencyDistribution: make(map[string]float64),
			MostProficient:          make([]string, 0),
			LeastProficient:         make([]string, 0),
		},
		developmentMetrics: &DevelopmentMetrics{
			CurrentStage:        "emergent",
			StageProgress:       0.0,
			NextStagePrediction: now.Add(24 * time.Hour),
			DevelopmentalHealth: 0.5,
		},
		wisdomMetrics: &WisdomMetrics{
			WisdomScore:      0.1, // Start with minimal wisdom
			WisdomMilestones: make([]WisdomMilestone, 0),
		},
		startTime:  now,
		lastUpdate: now,
	}
}

// UpdateActualizationMetrics updates metrics based on actualization system state
func (m *Metrics) UpdateActualizationMetrics(actualization *Actualization) {
	m.mu.Lock()
	defer m.mu.Unlock()

	metrics := actualization.GetActualizationMetrics()

	m.actualizationMetrics.TotalPotentials = metrics["total_potentials"].(int)
	m.actualizationMetrics.TotalActualizations = metrics["total_actualizations"].(int)

	// Calculate actualization rate
	timeSinceStart := time.Since(m.startTime).Hours()
	if timeSinceStart > 0 {
		m.actualizationMetrics.ActualizationRate = float64(m.actualizationMetrics.TotalActualizations) / timeSinceStart
	}

	// Count ready potentials
	readyPotentials := actualization.GetReadyPotentials(0.7)
	m.actualizationMetrics.ReadyPotentials = len(readyPotentials)

	// Calculate average readiness
	totalReadiness := 0.0
	for _, p := range actualization.potentials {
		totalReadiness += p.Readiness
	}
	if len(actualization.potentials) > 0 {
		m.actualizationMetrics.AverageReadiness = totalReadiness / float64(len(actualization.potentials))
	}

	m.lastUpdate = time.Now()
}

// UpdateProficiencyMetrics updates proficiency tracking
func (m *Metrics) UpdateProficiencyMetrics(actualizations []*ActualizedCapability) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(actualizations) == 0 {
		return
	}

	// Reset counters
	m.proficiencyMetrics.HighProficiencyCount = 0
	m.proficiencyMetrics.MediumProficiencyCount = 0
	m.proficiencyMetrics.LowProficiencyCount = 0

	totalProficiency := 0.0
	m.proficiencyMetrics.ProficiencyDistribution = make(map[string]float64)

	for _, ac := range actualizations {
		totalProficiency += ac.Proficiency
		m.proficiencyMetrics.ProficiencyDistribution[ac.Name] = ac.Proficiency

		// Categorize proficiency
		if ac.Proficiency > 0.8 {
			m.proficiencyMetrics.HighProficiencyCount++
		} else if ac.Proficiency >= 0.4 {
			m.proficiencyMetrics.MediumProficiencyCount++
		} else {
			m.proficiencyMetrics.LowProficiencyCount++
		}
	}

	m.proficiencyMetrics.AverageProficiency = totalProficiency / float64(len(actualizations))

	// Identify most and least proficient
	m.identifyProficiencyExtremes(actualizations)

	m.lastUpdate = time.Now()
}

// identifyProficiencyExtremes finds top and bottom capabilities
func (m *Metrics) identifyProficiencyExtremes(actualizations []*ActualizedCapability) {
	if len(actualizations) == 0 {
		return
	}

	// Simple approach: find top 3 and bottom 3
	type capProf struct {
		name        string
		proficiency float64
	}

	caps := make([]capProf, len(actualizations))
	for i, ac := range actualizations {
		caps[i] = capProf{name: ac.Name, proficiency: ac.Proficiency}
	}

	// Sort by proficiency (simple bubble sort for small lists)
	for i := 0; i < len(caps); i++ {
		for j := i + 1; j < len(caps); j++ {
			if caps[i].proficiency < caps[j].proficiency {
				caps[i], caps[j] = caps[j], caps[i]
			}
		}
	}

	// Top 3
	topCount := 3
	if len(caps) < topCount {
		topCount = len(caps)
	}
	m.proficiencyMetrics.MostProficient = make([]string, topCount)
	for i := range topCount {
		m.proficiencyMetrics.MostProficient[i] = caps[i].name
	}

	// Bottom 3
	bottomCount := 3
	if len(caps) < bottomCount {
		bottomCount = len(caps)
	}
	m.proficiencyMetrics.LeastProficient = make([]string, bottomCount)
	for i := range bottomCount {
		m.proficiencyMetrics.LeastProficient[i] = caps[len(caps)-1-i].name
	}
}

// UpdateDevelopmentMetrics updates developmental stage tracking
func (m *Metrics) UpdateDevelopmentMetrics(stage string, totalActualizations int, avgProficiency float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for stage transition
	if m.developmentMetrics.CurrentStage != stage {
		m.developmentMetrics.StageTransitions++
		m.developmentMetrics.CurrentStage = stage
		m.developmentMetrics.TimeInStage = 0
	} else {
		m.developmentMetrics.TimeInStage = time.Since(m.lastUpdate)
	}

	m.developmentMetrics.TotalDevelopmentTime = time.Since(m.startTime)

	// Calculate stage progress
	m.developmentMetrics.StageProgress = m.calculateStageProgress(stage, totalActualizations, avgProficiency)

	// Calculate developmental health
	m.developmentMetrics.DevelopmentalHealth = m.calculateDevelopmentalHealth(totalActualizations, avgProficiency)

	m.lastUpdate = time.Now()
}

// calculateStageProgress estimates progress within current stage
func (m *Metrics) calculateStageProgress(stage string, totalActualizations int, avgProficiency float64) float64 {
	switch stage {
	case "emergent":
		// Progress to developing at 3 actualizations
		return float64(totalActualizations) / 3.0
	case "developing":
		// Progress to maturing at 10 actualizations with 0.5 avg proficiency
		actualizationProgress := float64(totalActualizations) / 10.0
		proficiencyProgress := avgProficiency / 0.5
		return (actualizationProgress + proficiencyProgress) / 2.0
	case "maturing":
		// Progress to mature at 20 actualizations with 0.8 avg proficiency
		actualizationProgress := float64(totalActualizations) / 20.0
		proficiencyProgress := avgProficiency / 0.8
		return (actualizationProgress + proficiencyProgress) / 2.0
	case "mature":
		// Mature stage has ongoing progress
		return 0.9 + (avgProficiency * 0.1)
	default:
		return 0.0
	}
}

// calculateDevelopmentalHealth assesses overall developmental health
func (m *Metrics) calculateDevelopmentalHealth(totalActualizations int, avgProficiency float64) float64 {
	// Health is based on balance between quantity and quality
	quantityScore := float64(totalActualizations) / 20.0 // Normalize to 20 actualizations
	if quantityScore > 1.0 {
		quantityScore = 1.0
	}

	qualityScore := avgProficiency

	// Balance score - penalize imbalance
	balance := 1.0 - abs(quantityScore-qualityScore)

	// Overall health is weighted average
	health := (quantityScore*0.3 + qualityScore*0.5 + balance*0.2)

	return clampMetric(health, 0.0, 1.0)
}

// UpdateWisdomMetrics updates wisdom cultivation tracking
func (m *Metrics) UpdateWisdomMetrics(reflectionDepth, patternRecognition, contextualUnderstanding float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.wisdomMetrics.ReflectionDepth = reflectionDepth
	m.wisdomMetrics.PatternRecognition = patternRecognition
	m.wisdomMetrics.ContextualUnderstanding = contextualUnderstanding

	// Calculate overall wisdom score as weighted average
	m.wisdomMetrics.WisdomScore = (m.wisdomMetrics.KnowledgeIntegration*0.2 +
		m.wisdomMetrics.ReflectionDepth*0.25 +
		m.wisdomMetrics.PatternRecognition*0.2 +
		m.wisdomMetrics.ContextualUnderstanding*0.2 +
		m.wisdomMetrics.LongTermThinking*0.1 +
		m.wisdomMetrics.EthicalReasoning*0.05)

	// Check for wisdom milestones
	m.checkWisdomMilestones()

	m.lastUpdate = time.Now()
}

// checkWisdomMilestones identifies if new milestones have been reached
func (m *Metrics) checkWisdomMilestones() {
	milestones := []struct {
		threshold float64
		name      string
		desc      string
	}{
		{0.2, "First Glimmer", "Initial emergence of wisdom"},
		{0.4, "Growing Understanding", "Developing deeper comprehension"},
		{0.6, "Mature Insight", "Consistent wise reasoning"},
		{0.8, "Deep Wisdom", "Advanced wisdom cultivation"},
		{0.9, "Profound Wisdom", "Approaching wisdom mastery"},
	}

	for _, milestone := range milestones {
		if m.wisdomMetrics.WisdomScore >= milestone.threshold {
			// Check if already achieved
			alreadyAchieved := false
			for _, achieved := range m.wisdomMetrics.WisdomMilestones {
				if achieved.Name == milestone.name {
					alreadyAchieved = true
					break
				}
			}

			if !alreadyAchieved {
				m.wisdomMetrics.WisdomMilestones = append(m.wisdomMetrics.WisdomMilestones, WisdomMilestone{
					Name:         milestone.name,
					Description:  milestone.desc,
					AchievedAt:   time.Now(),
					Significance: milestone.threshold,
				})
			}
		}
	}
}

// GetComprehensiveMetrics returns all metrics as a structured report
func (m *Metrics) GetComprehensiveMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"actualization": map[string]interface{}{
			"total_potentials":     m.actualizationMetrics.TotalPotentials,
			"total_actualizations": m.actualizationMetrics.TotalActualizations,
			"actualization_rate":   m.actualizationMetrics.ActualizationRate,
			"ready_potentials":     m.actualizationMetrics.ReadyPotentials,
			"average_readiness":    m.actualizationMetrics.AverageReadiness,
		},
		"proficiency": map[string]interface{}{
			"average_proficiency":      m.proficiencyMetrics.AverageProficiency,
			"high_proficiency_count":   m.proficiencyMetrics.HighProficiencyCount,
			"medium_proficiency_count": m.proficiencyMetrics.MediumProficiencyCount,
			"low_proficiency_count":    m.proficiencyMetrics.LowProficiencyCount,
			"most_proficient":          m.proficiencyMetrics.MostProficient,
			"least_proficient":         m.proficiencyMetrics.LeastProficient,
		},
		"development": map[string]interface{}{
			"current_stage":          m.developmentMetrics.CurrentStage,
			"stage_progress":         m.developmentMetrics.StageProgress,
			"time_in_stage":          m.developmentMetrics.TimeInStage.String(),
			"total_development_time": m.developmentMetrics.TotalDevelopmentTime.String(),
			"stage_transitions":      m.developmentMetrics.StageTransitions,
			"developmental_health":   m.developmentMetrics.DevelopmentalHealth,
		},
		"wisdom": map[string]interface{}{
			"wisdom_score":             m.wisdomMetrics.WisdomScore,
			"knowledge_integration":    m.wisdomMetrics.KnowledgeIntegration,
			"reflection_depth":         m.wisdomMetrics.ReflectionDepth,
			"pattern_recognition":      m.wisdomMetrics.PatternRecognition,
			"contextual_understanding": m.wisdomMetrics.ContextualUnderstanding,
			"long_term_thinking":       m.wisdomMetrics.LongTermThinking,
			"ethical_reasoning":        m.wisdomMetrics.EthicalReasoning,
			"milestones_achieved":      len(m.wisdomMetrics.WisdomMilestones),
		},
		"meta": map[string]interface{}{
			"start_time":  m.startTime,
			"last_update": m.lastUpdate,
			"uptime":      time.Since(m.startTime).String(),
		},
	}
}

// GetActualizationMetrics returns actualization metrics
func (m *Metrics) GetActualizationMetrics() *ActualizationMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.actualizationMetrics
}

// GetWisdomMetrics returns wisdom metrics
func (m *Metrics) GetWisdomMetrics() *WisdomMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.wisdomMetrics
}

// Helper functions

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func clampMetric(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
