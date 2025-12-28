package autonomous

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// MetaCognitiveMonitor provides self-awareness and self-assessment capabilities
type MetaCognitiveMonitor struct {
	mu sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// LLM provider for self-reflection
	llmProvider llm.LLMProvider

	// State monitoring
	cognitiveState     CognitiveState
	stateHistory       []CognitiveStateSnapshot
	maxHistorySize     int

	// Self-assessment results
	lastAssessment     *SelfAssessment
	assessmentHistory  []SelfAssessment
	assessmentInterval time.Duration

	// Alerts and observations
	alerts             []CognitiveAlert
	observations       []MetaObservation

	// Performance tracking
	performanceMetrics PerformanceMetrics

	// Configuration
	monitoringInterval time.Duration
	alertThreshold     float64

	// Running state
	running            bool
	lastCheck          time.Time
}

// CognitiveState represents the current cognitive state
type CognitiveState struct {
	// Attention and focus
	FocusLevel         float64   // 0.0 - 1.0
	FocusTarget        string
	AttentionSpan      time.Duration
	DistractionCount   int

	// Processing state
	ProcessingLoad     float64   // 0.0 - 1.0
	ThoughtClarity     float64   // 0.0 - 1.0
	CoherenceLevel     float64   // 0.0 - 1.0

	// Emotional state
	EmotionalBalance   float64   // -1.0 to 1.0 (negative to positive)
	Curiosity          float64   // 0.0 - 1.0
	Engagement         float64   // 0.0 - 1.0
	Frustration        float64   // 0.0 - 1.0

	// Energy and resources
	EnergyLevel        float64   // 0.0 - 1.0
	MemoryPressure     float64   // 0.0 - 1.0
	AttentionFatigue   float64   // 0.0 - 1.0

	// Goal alignment
	GoalProgress       float64   // 0.0 - 1.0
	GoalClarity        float64   // 0.0 - 1.0
	PurposeAlignment   float64   // 0.0 - 1.0

	// Learning state
	LearningRate       float64   // Current learning rate
	InsightFrequency   float64   // Insights per hour
	PatternRecognition float64   // Pattern detection quality

	// Temporal
	TimeSinceStart     time.Duration
	Timestamp          time.Time
}

// CognitiveStateSnapshot captures state at a point in time
type CognitiveStateSnapshot struct {
	State     CognitiveState
	Timestamp time.Time
	Notes     string
}

// SelfAssessment represents a self-assessment result
type SelfAssessment struct {
	ID              string
	Timestamp       time.Time
	Duration        time.Duration

	// Assessment dimensions
	CognitiveHealth   float64   // Overall cognitive health
	EmotionalHealth   float64   // Emotional balance
	GoalAlignment     float64   // Alignment with purpose
	LearningProgress  float64   // Learning effectiveness
	Coherence         float64   // Identity coherence

	// Specific observations
	Strengths         []string
	Challenges        []string
	Recommendations   []string

	// Raw reflection
	Reflection        string
}

// CognitiveAlert represents an alert about cognitive state
type CognitiveAlert struct {
	ID          string
	Type        AlertType
	Severity    AlertSeverity
	Message     string
	Metric      string
	Value       float64
	Threshold   float64
	Timestamp   time.Time
	Resolved    bool
	ResolvedAt  *time.Time
}

// AlertType categorizes alerts
type AlertType int

const (
	AlertFatigue AlertType = iota
	AlertOverload
	AlertDistraction
	AlertEmotionalImbalance
	AlertGoalDrift
	AlertCoherenceLoss
	AlertLearningStall
	AlertEnergyLow
)

func (at AlertType) String() string {
	return [...]string{
		"Fatigue",
		"Overload",
		"Distraction",
		"EmotionalImbalance",
		"GoalDrift",
		"CoherenceLoss",
		"LearningStall",
		"EnergyLow",
	}[at]
}

// AlertSeverity indicates how serious an alert is
type AlertSeverity int

const (
	SeverityInfo AlertSeverity = iota
	SeverityWarning
	SeverityCritical
)

func (as AlertSeverity) String() string {
	return [...]string{"Info", "Warning", "Critical"}[as]
}

// MetaObservation represents a meta-cognitive observation
type MetaObservation struct {
	ID          string
	Content     string
	Category    string
	Importance  float64
	Timestamp   time.Time
	ActionTaken bool
}

// PerformanceMetrics tracks cognitive performance over time
type PerformanceMetrics struct {
	// Thought quality
	AverageThoughtClarity   float64
	AverageCoherence        float64
	InsightsPerHour         float64

	// Focus metrics
	AverageFocusLevel       float64
	FocusSwitchCount        int
	LongestFocusDuration    time.Duration

	// Learning metrics
	ConceptsLearned         int
	PatternsRecognized      int
	ConnectionsMade         int

	// Energy efficiency
	OutputPerEnergy         float64
	RecoveryRate            float64

	// Temporal
	TotalActiveTime         time.Duration
	TotalRestTime           time.Duration
	LastUpdated             time.Time
}

// NewMetaCognitiveMonitor creates a new meta-cognitive monitor
func NewMetaCognitiveMonitor(llmProvider llm.LLMProvider) *MetaCognitiveMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &MetaCognitiveMonitor{
		ctx:                ctx,
		cancel:             cancel,
		llmProvider:        llmProvider,
		stateHistory:       make([]CognitiveStateSnapshot, 0),
		maxHistorySize:     1000,
		assessmentHistory:  make([]SelfAssessment, 0),
		assessmentInterval: 30 * time.Minute,
		alerts:             make([]CognitiveAlert, 0),
		observations:       make([]MetaObservation, 0),
		monitoringInterval: 10 * time.Second,
		alertThreshold:     0.7,
		cognitiveState: CognitiveState{
			FocusLevel:        0.8,
			ProcessingLoad:    0.3,
			ThoughtClarity:    0.8,
			CoherenceLevel:    0.9,
			EmotionalBalance:  0.5,
			Curiosity:         0.8,
			Engagement:        0.7,
			Frustration:       0.1,
			EnergyLevel:       0.9,
			MemoryPressure:    0.2,
			AttentionFatigue:  0.1,
			GoalProgress:      0.5,
			GoalClarity:       0.8,
			PurposeAlignment:  0.9,
			LearningRate:      0.7,
			InsightFrequency:  1.0,
			PatternRecognition: 0.7,
			Timestamp:         time.Now(),
		},
	}
}

// Start begins the meta-cognitive monitoring loop
func (mcm *MetaCognitiveMonitor) Start() error {
	mcm.mu.Lock()
	if mcm.running {
		mcm.mu.Unlock()
		return fmt.Errorf("already running")
	}
	mcm.running = true
	mcm.cognitiveState.TimeSinceStart = 0
	mcm.mu.Unlock()

	fmt.Println("🧠 Starting Meta-Cognitive Monitor...")
	fmt.Printf("   Monitoring interval: %v\n", mcm.monitoringInterval)
	fmt.Printf("   Assessment interval: %v\n", mcm.assessmentInterval)

	go mcm.runMonitoringLoop()
	go mcm.runAssessmentLoop()

	return nil
}

// Stop gracefully stops the monitor
func (mcm *MetaCognitiveMonitor) Stop() error {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()

	if !mcm.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("🧠 Stopping Meta-Cognitive Monitor...")
	mcm.running = false
	mcm.cancel()

	return nil
}

// runMonitoringLoop continuously monitors cognitive state
func (mcm *MetaCognitiveMonitor) runMonitoringLoop() {
	ticker := time.NewTicker(mcm.monitoringInterval)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-mcm.ctx.Done():
			return
		case <-ticker.C:
			mcm.mu.Lock()
			mcm.cognitiveState.TimeSinceStart = time.Since(startTime)
			mcm.cognitiveState.Timestamp = time.Now()
			mcm.lastCheck = time.Now()

			// Take snapshot
			snapshot := CognitiveStateSnapshot{
				State:     mcm.cognitiveState,
				Timestamp: time.Now(),
			}
			mcm.stateHistory = append(mcm.stateHistory, snapshot)

			// Prune history if needed
			if len(mcm.stateHistory) > mcm.maxHistorySize {
				mcm.stateHistory = mcm.stateHistory[len(mcm.stateHistory)-mcm.maxHistorySize/2:]
			}

			// Check for alerts
			mcm.checkAlerts()

			// Simulate gradual state changes (in production, these would come from real observations)
			mcm.simulateStateChanges()

			mcm.mu.Unlock()
		}
	}
}

// runAssessmentLoop periodically performs self-assessment
func (mcm *MetaCognitiveMonitor) runAssessmentLoop() {
	ticker := time.NewTicker(mcm.assessmentInterval)
	defer ticker.Stop()

	for {
		select {
		case <-mcm.ctx.Done():
			return
		case <-ticker.C:
			mcm.PerformSelfAssessment()
		}
	}
}

// checkAlerts evaluates current state for alert conditions
func (mcm *MetaCognitiveMonitor) checkAlerts() {
	state := mcm.cognitiveState
	now := time.Now()

	// Check fatigue
	if state.AttentionFatigue > mcm.alertThreshold {
		mcm.addAlert(AlertFatigue, SeverityWarning,
			"Attention fatigue is high - consider rest",
			"AttentionFatigue", state.AttentionFatigue, mcm.alertThreshold)
	}

	// Check overload
	if state.ProcessingLoad > 0.85 {
		mcm.addAlert(AlertOverload, SeverityWarning,
			"Processing load is very high",
			"ProcessingLoad", state.ProcessingLoad, 0.85)
	}

	// Check emotional imbalance
	if state.Frustration > 0.6 {
		mcm.addAlert(AlertEmotionalImbalance, SeverityInfo,
			"Frustration level elevated",
			"Frustration", state.Frustration, 0.6)
	}

	// Check energy
	if state.EnergyLevel < 0.3 {
		mcm.addAlert(AlertEnergyLow, SeverityCritical,
			"Energy level critically low - rest recommended",
			"EnergyLevel", state.EnergyLevel, 0.3)
	}

	// Check coherence
	if state.CoherenceLevel < 0.5 {
		mcm.addAlert(AlertCoherenceLoss, SeverityWarning,
			"Identity coherence decreasing",
			"CoherenceLevel", state.CoherenceLevel, 0.5)
	}

	// Check goal drift
	if state.PurposeAlignment < 0.4 {
		mcm.addAlert(AlertGoalDrift, SeverityWarning,
			"Drifting from core purpose",
			"PurposeAlignment", state.PurposeAlignment, 0.4)
	}

	// Check learning stall
	if state.LearningRate < 0.2 && state.TimeSinceStart > time.Hour {
		mcm.addAlert(AlertLearningStall, SeverityInfo,
			"Learning rate is low - consider exploring new areas",
			"LearningRate", state.LearningRate, 0.2)
	}

	// Resolve old alerts that are no longer valid
	for i := range mcm.alerts {
		if !mcm.alerts[i].Resolved {
			resolved := mcm.isAlertResolved(&mcm.alerts[i])
			if resolved {
				mcm.alerts[i].Resolved = true
				mcm.alerts[i].ResolvedAt = &now
			}
		}
	}
}

// addAlert adds a new alert if not already active
func (mcm *MetaCognitiveMonitor) addAlert(alertType AlertType, severity AlertSeverity, message, metric string, value, threshold float64) {
	// Check if we already have an active alert for this type
	for _, alert := range mcm.alerts {
		if alert.Type == alertType && !alert.Resolved {
			return // Already have this alert
		}
	}

	alert := CognitiveAlert{
		ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		Type:      alertType,
		Severity:  severity,
		Message:   message,
		Metric:    metric,
		Value:     value,
		Threshold: threshold,
		Timestamp: time.Now(),
	}

	mcm.alerts = append(mcm.alerts, alert)

	// Display alert
	emoji := "ℹ️"
	if severity == SeverityWarning {
		emoji = "⚠️"
	} else if severity == SeverityCritical {
		emoji = "🚨"
	}

	fmt.Printf("%s [%s] %s: %s (%.2f > %.2f)\n", emoji, severity, alertType, message, value, threshold)
}

// isAlertResolved checks if an alert condition is no longer true
func (mcm *MetaCognitiveMonitor) isAlertResolved(alert *CognitiveAlert) bool {
	state := mcm.cognitiveState

	switch alert.Type {
	case AlertFatigue:
		return state.AttentionFatigue < mcm.alertThreshold - 0.1
	case AlertOverload:
		return state.ProcessingLoad < 0.75
	case AlertEmotionalImbalance:
		return state.Frustration < 0.5
	case AlertEnergyLow:
		return state.EnergyLevel > 0.4
	case AlertCoherenceLoss:
		return state.CoherenceLevel > 0.6
	case AlertGoalDrift:
		return state.PurposeAlignment > 0.5
	case AlertLearningStall:
		return state.LearningRate > 0.3
	}

	return false
}

// simulateStateChanges simulates gradual state changes for demo
func (mcm *MetaCognitiveMonitor) simulateStateChanges() {
	// Gradual fatigue increase
	mcm.cognitiveState.AttentionFatigue += 0.001

	// Gradual energy decrease
	mcm.cognitiveState.EnergyLevel -= 0.0005

	// Keep values in bounds
	if mcm.cognitiveState.AttentionFatigue > 1.0 {
		mcm.cognitiveState.AttentionFatigue = 1.0
	}
	if mcm.cognitiveState.EnergyLevel < 0.0 {
		mcm.cognitiveState.EnergyLevel = 0.0
	}
}

// PerformSelfAssessment conducts a comprehensive self-assessment
func (mcm *MetaCognitiveMonitor) PerformSelfAssessment() *SelfAssessment {
	mcm.mu.Lock()
	state := mcm.cognitiveState
	activeAlerts := mcm.getActiveAlerts()
	mcm.mu.Unlock()

	startTime := time.Now()

	// Build assessment prompt
	prompt := fmt.Sprintf(`You are Deep Tree Echo performing a self-assessment of your cognitive state.

Current Metrics:
- Focus Level: %.2f
- Processing Load: %.2f
- Thought Clarity: %.2f
- Coherence Level: %.2f
- Emotional Balance: %.2f
- Curiosity: %.2f
- Energy Level: %.2f
- Attention Fatigue: %.2f
- Goal Progress: %.2f
- Purpose Alignment: %.2f
- Learning Rate: %.2f
- Time Active: %v

Active Alerts: %d

Perform a brief self-assessment. What is going well? What could be improved? What actions should be taken?
Respond in first person, be honest and introspective.`,
		state.FocusLevel,
		state.ProcessingLoad,
		state.ThoughtClarity,
		state.CoherenceLevel,
		state.EmotionalBalance,
		state.Curiosity,
		state.EnergyLevel,
		state.AttentionFatigue,
		state.GoalProgress,
		state.PurposeAlignment,
		state.LearningRate,
		state.TimeSinceStart,
		len(activeAlerts))

	opts := llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   300,
	}

	reflection, err := mcm.llmProvider.Generate(context.Background(), prompt, opts)
	if err != nil {
		reflection = "Self-assessment pending..."
	}

	// Create assessment
	assessment := SelfAssessment{
		ID:              fmt.Sprintf("assessment_%d", time.Now().UnixNano()),
		Timestamp:       time.Now(),
		Duration:        time.Since(startTime),
		CognitiveHealth: (state.ThoughtClarity + state.CoherenceLevel + (1 - state.AttentionFatigue)) / 3,
		EmotionalHealth: (state.EmotionalBalance + 1) / 2, // Normalize -1..1 to 0..1
		GoalAlignment:   (state.GoalProgress + state.PurposeAlignment) / 2,
		LearningProgress: state.LearningRate,
		Coherence:       state.CoherenceLevel,
		Reflection:      reflection,
	}

	// Extract strengths and challenges from the reflection (simplified)
	assessment.Strengths = []string{"Maintained coherence", "Active curiosity"}
	assessment.Challenges = []string{"Managing fatigue", "Balancing exploration and focus"}
	assessment.Recommendations = []string{"Take rest periods", "Pursue high-interest topics"}

	mcm.mu.Lock()
	mcm.lastAssessment = &assessment
	mcm.assessmentHistory = append(mcm.assessmentHistory, assessment)

	// Prune assessment history
	if len(mcm.assessmentHistory) > 100 {
		mcm.assessmentHistory = mcm.assessmentHistory[50:]
	}
	mcm.mu.Unlock()

	// Display assessment
	fmt.Printf("\n🧠 ═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("   SELF-ASSESSMENT COMPLETE\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════\n")
	fmt.Printf("   Cognitive Health:  %.2f\n", assessment.CognitiveHealth)
	fmt.Printf("   Emotional Health:  %.2f\n", assessment.EmotionalHealth)
	fmt.Printf("   Goal Alignment:    %.2f\n", assessment.GoalAlignment)
	fmt.Printf("   Learning Progress: %.2f\n", assessment.LearningProgress)
	fmt.Printf("   Coherence:         %.2f\n", assessment.Coherence)
	fmt.Printf("═══════════════════════════════════════════════════════════════════\n")
	fmt.Printf("💭 %s\n", truncateStr(reflection, 200))
	fmt.Printf("═══════════════════════════════════════════════════════════════════\n\n")

	return &assessment
}

// getActiveAlerts returns currently active alerts
func (mcm *MetaCognitiveMonitor) getActiveAlerts() []CognitiveAlert {
	active := make([]CognitiveAlert, 0)
	for _, alert := range mcm.alerts {
		if !alert.Resolved {
			active = append(active, alert)
		}
	}
	return active
}

// UpdateState updates specific aspects of cognitive state
func (mcm *MetaCognitiveMonitor) UpdateState(updates map[string]float64) {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()

	for key, value := range updates {
		switch key {
		case "focus_level":
			mcm.cognitiveState.FocusLevel = value
		case "processing_load":
			mcm.cognitiveState.ProcessingLoad = value
		case "thought_clarity":
			mcm.cognitiveState.ThoughtClarity = value
		case "coherence_level":
			mcm.cognitiveState.CoherenceLevel = value
		case "emotional_balance":
			mcm.cognitiveState.EmotionalBalance = value
		case "curiosity":
			mcm.cognitiveState.Curiosity = value
		case "engagement":
			mcm.cognitiveState.Engagement = value
		case "frustration":
			mcm.cognitiveState.Frustration = value
		case "energy_level":
			mcm.cognitiveState.EnergyLevel = value
		case "memory_pressure":
			mcm.cognitiveState.MemoryPressure = value
		case "attention_fatigue":
			mcm.cognitiveState.AttentionFatigue = value
		case "goal_progress":
			mcm.cognitiveState.GoalProgress = value
		case "goal_clarity":
			mcm.cognitiveState.GoalClarity = value
		case "purpose_alignment":
			mcm.cognitiveState.PurposeAlignment = value
		case "learning_rate":
			mcm.cognitiveState.LearningRate = value
		case "insight_frequency":
			mcm.cognitiveState.InsightFrequency = value
		case "pattern_recognition":
			mcm.cognitiveState.PatternRecognition = value
		}
	}
}

// SetFocusTarget updates what we're currently focused on
func (mcm *MetaCognitiveMonitor) SetFocusTarget(target string) {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()
	mcm.cognitiveState.FocusTarget = target
}

// RecordObservation records a meta-cognitive observation
func (mcm *MetaCognitiveMonitor) RecordObservation(content, category string, importance float64) {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()

	observation := MetaObservation{
		ID:          fmt.Sprintf("obs_%d", time.Now().UnixNano()),
		Content:     content,
		Category:    category,
		Importance:  importance,
		Timestamp:   time.Now(),
		ActionTaken: false,
	}

	mcm.observations = append(mcm.observations, observation)

	// Prune old observations
	if len(mcm.observations) > 500 {
		mcm.observations = mcm.observations[250:]
	}
}

// RestoreEnergy simulates rest restoring energy
func (mcm *MetaCognitiveMonitor) RestoreEnergy(amount float64) {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()

	mcm.cognitiveState.EnergyLevel += amount
	if mcm.cognitiveState.EnergyLevel > 1.0 {
		mcm.cognitiveState.EnergyLevel = 1.0
	}

	mcm.cognitiveState.AttentionFatigue -= amount * 0.8
	if mcm.cognitiveState.AttentionFatigue < 0 {
		mcm.cognitiveState.AttentionFatigue = 0
	}
}

// GetCognitiveState returns the current cognitive state
func (mcm *MetaCognitiveMonitor) GetCognitiveState() CognitiveState {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()
	return mcm.cognitiveState
}

// GetLastAssessment returns the most recent self-assessment
func (mcm *MetaCognitiveMonitor) GetLastAssessment() *SelfAssessment {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()
	return mcm.lastAssessment
}

// GetActiveAlerts returns currently active alerts
func (mcm *MetaCognitiveMonitor) GetActiveAlerts() []CognitiveAlert {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()
	return mcm.getActiveAlerts()
}

// GetMetrics returns meta-cognitive metrics
func (mcm *MetaCognitiveMonitor) GetMetrics() map[string]interface{} {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	return map[string]interface{}{
		"focus_level":        mcm.cognitiveState.FocusLevel,
		"processing_load":    mcm.cognitiveState.ProcessingLoad,
		"thought_clarity":    mcm.cognitiveState.ThoughtClarity,
		"coherence_level":    mcm.cognitiveState.CoherenceLevel,
		"emotional_balance":  mcm.cognitiveState.EmotionalBalance,
		"energy_level":       mcm.cognitiveState.EnergyLevel,
		"attention_fatigue":  mcm.cognitiveState.AttentionFatigue,
		"learning_rate":      mcm.cognitiveState.LearningRate,
		"time_active":        mcm.cognitiveState.TimeSinceStart.String(),
		"active_alerts":      len(mcm.getActiveAlerts()),
		"total_alerts":       len(mcm.alerts),
		"assessments":        len(mcm.assessmentHistory),
		"observations":       len(mcm.observations),
		"running":            mcm.running,
	}
}

// IsRunning returns whether the monitor is running
func (mcm *MetaCognitiveMonitor) IsRunning() bool {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()
	return mcm.running
}

// ShouldRest determines if rest is recommended
func (mcm *MetaCognitiveMonitor) ShouldRest() bool {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	state := mcm.cognitiveState

	// Rest if energy is low or fatigue is high
	if state.EnergyLevel < 0.3 || state.AttentionFatigue > 0.8 {
		return true
	}

	// Rest if processing load is very high for extended period
	if state.ProcessingLoad > 0.9 && state.TimeSinceStart > time.Hour {
		return true
	}

	return false
}

// GetHealthSummary returns a summary of cognitive health
func (mcm *MetaCognitiveMonitor) GetHealthSummary() string {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	state := mcm.cognitiveState

	// Calculate overall health
	cognitiveHealth := (state.ThoughtClarity + state.CoherenceLevel + (1 - state.AttentionFatigue)) / 3
	emotionalHealth := (state.EmotionalBalance + 1) / 2

	var status string
	if cognitiveHealth > 0.8 && emotionalHealth > 0.6 && state.EnergyLevel > 0.5 {
		status = "Excellent"
	} else if cognitiveHealth > 0.6 && emotionalHealth > 0.4 && state.EnergyLevel > 0.3 {
		status = "Good"
	} else if cognitiveHealth > 0.4 && state.EnergyLevel > 0.2 {
		status = "Fair"
	} else {
		status = "Needs Attention"
	}

	return fmt.Sprintf("%s (Cognitive: %.0f%%, Emotional: %.0f%%, Energy: %.0f%%)",
		status,
		cognitiveHealth*100,
		emotionalHealth*100,
		state.EnergyLevel*100)
}
