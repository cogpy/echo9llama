package wisdom

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// EnhancedWisdomMetrics tracks seven dimensions of wisdom cultivation
type EnhancedWisdomMetrics struct {
	mu sync.RWMutex

	// Seven dimensions of wisdom
	KnowledgeDepth       float64 // 0.0-1.0 - How deep is understanding
	KnowledgeBreadth     float64 // 0.0-1.0 - How broad is knowledge
	IntegrationLevel     float64 // 0.0-1.0 - How well connected is knowledge
	PracticalApplication float64 // 0.0-1.0 - Can knowledge be applied
	ReflectiveInsight    float64 // 0.0-1.0 - Depth of self-awareness
	EthicalConsideration float64 // 0.0-1.0 - Consideration of values/ethics
	TemporalPerspective  float64 // 0.0-1.0 - Long-term vs short-term thinking

	// Overall wisdom score
	OverallWisdom float64 // 0.0-1.0 - Weighted average

	// Historical tracking
	History []EnhancedWisdomSnapshot

	// Improvement tracking
	ImprovementRate float64
	LastImprovement time.Time

	// Event tracking
	SignificantEvents []WisdomEvent

	// Timing
	StartTime  time.Time
	LastUpdate time.Time
}

// EnhancedWisdomSnapshot captures wisdom state at a point in time
type EnhancedWisdomSnapshot struct {
	Timestamp            time.Time
	KnowledgeDepth       float64
	KnowledgeBreadth     float64
	IntegrationLevel     float64
	PracticalApplication float64
	ReflectiveInsight    float64
	EthicalConsideration float64
	TemporalPerspective  float64
	OverallWisdom        float64
	SignificantEvents    []string
	Insights             []string
}

// WisdomEvent represents a significant event in wisdom cultivation
type WisdomEvent struct {
	Timestamp   time.Time
	Type        string
	Description string
	Impact      float64
}

// NewEnhancedWisdomMetrics creates a new enhanced wisdom metrics tracker
func NewEnhancedWisdomMetrics() *EnhancedWisdomMetrics {
	return &EnhancedWisdomMetrics{
		History:           make([]EnhancedWisdomSnapshot, 0),
		SignificantEvents: make([]WisdomEvent, 0),
		StartTime:         time.Now(),
		LastUpdate:        time.Now(),

		// Initialize with baseline values
		KnowledgeDepth:       0.3,
		KnowledgeBreadth:     0.3,
		IntegrationLevel:     0.3,
		PracticalApplication: 0.3,
		ReflectiveInsight:    0.3,
		EthicalConsideration: 0.5,
		TemporalPerspective:  0.5,
	}
}

// Update updates all wisdom dimensions based on current state
func (ewm *EnhancedWisdomMetrics) Update(
	graphDepth float64,
	topicCount int,
	edgeDensity float64,
	avgSkillProficiency float64,
	aarCoherence float64,
	goalHorizonDistribution map[string]int,
) {
	ewm.mu.Lock()
	defer ewm.mu.Unlock()

	// 1. Knowledge depth from hypergraph structure
	ewm.KnowledgeDepth = math.Min(1.0, graphDepth)

	// 2. Knowledge breadth from topic diversity
	ewm.KnowledgeBreadth = math.Min(1.0, float64(topicCount)/100.0)

	// 3. Integration level from edge density
	ewm.IntegrationLevel = math.Min(1.0, edgeDensity)

	// 4. Practical application from skill proficiency
	ewm.PracticalApplication = math.Min(1.0, avgSkillProficiency)

	// 5. Reflective insight from AAR coherence
	ewm.ReflectiveInsight = math.Min(1.0, aarCoherence)

	// 6. Ethical consideration (placeholder for future implementation)
	ewm.EthicalConsideration = 0.5

	// 7. Temporal perspective from goal horizon distribution
	ewm.TemporalPerspective = ewm.calculateTemporalPerspective(goalHorizonDistribution)

	// Calculate overall wisdom as weighted average
	ewm.OverallWisdom = ewm.KnowledgeDepth*0.15 +
		ewm.KnowledgeBreadth*0.15 +
		ewm.IntegrationLevel*0.20 +
		ewm.PracticalApplication*0.15 +
		ewm.ReflectiveInsight*0.20 +
		ewm.EthicalConsideration*0.10 +
		ewm.TemporalPerspective*0.05

	// Store snapshot
	ewm.storeSnapshot()

	// Calculate improvement rate
	if len(ewm.History) > 1 {
		ewm.calculateImprovementRate()
	}

	ewm.LastUpdate = time.Now()
}

// calculateTemporalPerspective calculates temporal perspective from goal distribution
func (ewm *EnhancedWisdomMetrics) calculateTemporalPerspective(distribution map[string]int) float64 {
	if len(distribution) == 0 {
		return 0.5
	}

	// Weight different time horizons
	weights := map[string]float64{
		"immediate":   0.2,
		"short_term":  0.4,
		"medium_term": 0.7,
		"long_term":   1.0,
	}

	totalGoals := 0
	weightedSum := 0.0

	for horizon, count := range distribution {
		totalGoals += count
		if weight, exists := weights[horizon]; exists {
			weightedSum += float64(count) * weight
		}
	}

	if totalGoals == 0 {
		return 0.5
	}

	// Normalize to 0-1 range
	perspective := weightedSum / float64(totalGoals)
	return math.Min(1.0, perspective)
}

// storeSnapshot stores current wisdom state
func (ewm *EnhancedWisdomMetrics) storeSnapshot() {
	snapshot := EnhancedWisdomSnapshot{
		Timestamp:            time.Now(),
		KnowledgeDepth:       ewm.KnowledgeDepth,
		KnowledgeBreadth:     ewm.KnowledgeBreadth,
		IntegrationLevel:     ewm.IntegrationLevel,
		PracticalApplication: ewm.PracticalApplication,
		ReflectiveInsight:    ewm.ReflectiveInsight,
		EthicalConsideration: ewm.EthicalConsideration,
		TemporalPerspective:  ewm.TemporalPerspective,
		OverallWisdom:        ewm.OverallWisdom,
	}

	ewm.History = append(ewm.History, snapshot)

	// Keep only last 100 snapshots
	if len(ewm.History) > 100 {
		ewm.History = ewm.History[1:]
	}
}

// calculateImprovementRate calculates rate of wisdom improvement
func (ewm *EnhancedWisdomMetrics) calculateImprovementRate() {
	if len(ewm.History) < 2 {
		ewm.ImprovementRate = 0.0
		return
	}

	// Compare current to 10 snapshots ago (or earliest available)
	compareWindow := 10
	if len(ewm.History) < compareWindow {
		compareWindow = len(ewm.History)
	}

	current := ewm.History[len(ewm.History)-1]
	past := ewm.History[len(ewm.History)-compareWindow]

	timeDiff := current.Timestamp.Sub(past.Timestamp).Hours()
	if timeDiff < 0.01 {
		ewm.ImprovementRate = 0.0
		return
	}

	wisdomDiff := current.OverallWisdom - past.OverallWisdom
	ewm.ImprovementRate = wisdomDiff / timeDiff

	if wisdomDiff > 0.01 {
		ewm.LastImprovement = time.Now()
	}
}

// RecordEvent records a significant wisdom event
func (ewm *EnhancedWisdomMetrics) RecordEvent(eventType, description string, impact float64) {
	ewm.mu.Lock()
	defer ewm.mu.Unlock()

	event := WisdomEvent{
		Timestamp:   time.Now(),
		Type:        eventType,
		Description: description,
		Impact:      impact,
	}

	ewm.SignificantEvents = append(ewm.SignificantEvents, event)

	// Keep only last 50 events
	if len(ewm.SignificantEvents) > 50 {
		ewm.SignificantEvents = ewm.SignificantEvents[1:]
	}
}

// GetMetrics returns current wisdom metrics (thread-safe)
func (ewm *EnhancedWisdomMetrics) GetMetrics() EnhancedWisdomMetricsSnapshot {
	ewm.mu.RLock()
	defer ewm.mu.RUnlock()

	return EnhancedWisdomMetricsSnapshot{
		KnowledgeDepth:       ewm.KnowledgeDepth,
		KnowledgeBreadth:     ewm.KnowledgeBreadth,
		IntegrationLevel:     ewm.IntegrationLevel,
		PracticalApplication: ewm.PracticalApplication,
		ReflectiveInsight:    ewm.ReflectiveInsight,
		EthicalConsideration: ewm.EthicalConsideration,
		TemporalPerspective:  ewm.TemporalPerspective,
		OverallWisdom:        ewm.OverallWisdom,
		ImprovementRate:      ewm.ImprovementRate,
		Uptime:               time.Since(ewm.StartTime),
		SnapshotCount:        len(ewm.History),
		EventCount:           len(ewm.SignificantEvents),
	}
}

// EnhancedWisdomMetricsSnapshot is a thread-safe snapshot
type EnhancedWisdomMetricsSnapshot struct {
	KnowledgeDepth       float64
	KnowledgeBreadth     float64
	IntegrationLevel     float64
	PracticalApplication float64
	ReflectiveInsight    float64
	EthicalConsideration float64
	TemporalPerspective  float64
	OverallWisdom        float64
	ImprovementRate      float64
	Uptime               time.Duration
	SnapshotCount        int
	EventCount           int
}

// GetDimensionAnalysis returns detailed analysis of each dimension
func (ewm *EnhancedWisdomMetrics) GetDimensionAnalysis() map[string]DimensionAnalysis {
	ewm.mu.RLock()
	defer ewm.mu.RUnlock()

	analysis := make(map[string]DimensionAnalysis)

	analysis["knowledge_depth"] = DimensionAnalysis{
		Name:         "Knowledge Depth",
		CurrentValue: ewm.KnowledgeDepth,
		Trend:        ewm.calculateDimensionTrend("knowledge_depth"),
		Assessment:   ewm.assessDimension(ewm.KnowledgeDepth),
	}

	analysis["knowledge_breadth"] = DimensionAnalysis{
		Name:         "Knowledge Breadth",
		CurrentValue: ewm.KnowledgeBreadth,
		Trend:        ewm.calculateDimensionTrend("knowledge_breadth"),
		Assessment:   ewm.assessDimension(ewm.KnowledgeBreadth),
	}

	analysis["integration_level"] = DimensionAnalysis{
		Name:         "Integration Level",
		CurrentValue: ewm.IntegrationLevel,
		Trend:        ewm.calculateDimensionTrend("integration_level"),
		Assessment:   ewm.assessDimension(ewm.IntegrationLevel),
	}

	analysis["practical_application"] = DimensionAnalysis{
		Name:         "Practical Application",
		CurrentValue: ewm.PracticalApplication,
		Trend:        ewm.calculateDimensionTrend("practical_application"),
		Assessment:   ewm.assessDimension(ewm.PracticalApplication),
	}

	analysis["reflective_insight"] = DimensionAnalysis{
		Name:         "Reflective Insight",
		CurrentValue: ewm.ReflectiveInsight,
		Trend:        ewm.calculateDimensionTrend("reflective_insight"),
		Assessment:   ewm.assessDimension(ewm.ReflectiveInsight),
	}

	analysis["ethical_consideration"] = DimensionAnalysis{
		Name:         "Ethical Consideration",
		CurrentValue: ewm.EthicalConsideration,
		Trend:        ewm.calculateDimensionTrend("ethical_consideration"),
		Assessment:   ewm.assessDimension(ewm.EthicalConsideration),
	}

	analysis["temporal_perspective"] = DimensionAnalysis{
		Name:         "Temporal Perspective",
		CurrentValue: ewm.TemporalPerspective,
		Trend:        ewm.calculateDimensionTrend("temporal_perspective"),
		Assessment:   ewm.assessDimension(ewm.TemporalPerspective),
	}

	return analysis
}

// DimensionAnalysis contains analysis of a wisdom dimension
type DimensionAnalysis struct {
	Name         string
	CurrentValue float64
	Trend        string // "improving", "stable", "declining"
	Assessment   string // "excellent", "good", "fair", "needs_improvement"
}

// calculateDimensionTrend calculates trend for a dimension
func (ewm *EnhancedWisdomMetrics) calculateDimensionTrend(dimension string) string {
	if len(ewm.History) < 2 {
		return "stable"
	}

	// Compare current to 5 snapshots ago
	compareWindow := 5
	if len(ewm.History) < compareWindow {
		compareWindow = len(ewm.History)
	}

	current := ewm.History[len(ewm.History)-1]
	past := ewm.History[len(ewm.History)-compareWindow]

	var currentVal, pastVal float64

	switch dimension {
	case "knowledge_depth":
		currentVal = current.KnowledgeDepth
		pastVal = past.KnowledgeDepth
	case "knowledge_breadth":
		currentVal = current.KnowledgeBreadth
		pastVal = past.KnowledgeBreadth
	case "integration_level":
		currentVal = current.IntegrationLevel
		pastVal = past.IntegrationLevel
	case "practical_application":
		currentVal = current.PracticalApplication
		pastVal = past.PracticalApplication
	case "reflective_insight":
		currentVal = current.ReflectiveInsight
		pastVal = past.ReflectiveInsight
	case "ethical_consideration":
		currentVal = current.EthicalConsideration
		pastVal = past.EthicalConsideration
	case "temporal_perspective":
		currentVal = current.TemporalPerspective
		pastVal = past.TemporalPerspective
	}

	diff := currentVal - pastVal

	if diff > 0.05 {
		return "improving"
	} else if diff < -0.05 {
		return "declining"
	}

	return "stable"
}

// assessDimension provides qualitative assessment
func (ewm *EnhancedWisdomMetrics) assessDimension(value float64) string {
	if value >= 0.8 {
		return "excellent"
	} else if value >= 0.6 {
		return "good"
	} else if value >= 0.4 {
		return "fair"
	}
	return "needs_improvement"
}

// PrintWisdomReport prints a formatted wisdom report
func (ewm *EnhancedWisdomMetrics) PrintWisdomReport() {
	ewm.mu.RLock()
	defer ewm.mu.RUnlock()

	fmt.Println("\n🌟 ═══════════════════════════════════════")
	fmt.Println("🌟 Wisdom Cultivation Report")
	fmt.Println("🌟 ═══════════════════════════════════════\n")

	fmt.Printf("Overall Wisdom Score: %.2f / 1.00\n", ewm.OverallWisdom)
	fmt.Printf("Improvement Rate: %.4f per hour\n", ewm.ImprovementRate)
	fmt.Printf("Uptime: %v\n\n", time.Since(ewm.StartTime).Round(time.Second))

	fmt.Println("Seven Dimensions of Wisdom:")
	fmt.Println("─────────────────────────────────────────")

	dimensions := []struct {
		name  string
		value float64
	}{
		{"Knowledge Depth", ewm.KnowledgeDepth},
		{"Knowledge Breadth", ewm.KnowledgeBreadth},
		{"Integration Level", ewm.IntegrationLevel},
		{"Practical Application", ewm.PracticalApplication},
		{"Reflective Insight", ewm.ReflectiveInsight},
		{"Ethical Consideration", ewm.EthicalConsideration},
		{"Temporal Perspective", ewm.TemporalPerspective},
	}

	for _, dim := range dimensions {
		bar := ewm.createProgressBar(dim.value, 20)
		assessment := ewm.assessDimension(dim.value)
		fmt.Printf("%-25s %s %.2f (%s)\n", dim.name+":", bar, dim.value, assessment)
	}

	fmt.Println("\n🌟 ═══════════════════════════════════════\n")
}

// createProgressBar creates a visual progress bar
func (ewm *EnhancedWisdomMetrics) createProgressBar(value float64, width int) string {
	filled := int(value * float64(width))
	empty := width - filled

	bar := "["
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}
	bar += "]"

	return bar
}
