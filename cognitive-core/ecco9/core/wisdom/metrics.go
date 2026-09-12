package wisdom

import (
	"sync"
	"time"
)

// WisdomMetrics tracks the cultivation of wisdom over time
type WisdomMetrics struct {
	mu sync.RWMutex

	// Core metrics
	WisdomDepthScore   float64
	CoherenceStability float64
	LearningVelocity   float64
	InsightFrequency   float64

	// Historical tracking
	WisdomHistory      []WisdomSnapshot
	InsightCount       int64
	SkillsAcquired     int64
	PatternsRecognized int64

	// Temporal tracking
	StartTime  time.Time
	LastUpdate time.Time
}

// WisdomSnapshot represents wisdom state at a point in time
type WisdomSnapshot struct {
	Timestamp    time.Time
	WisdomScore  float64
	Coherence    float64
	Stability    float64
	Awareness    float64
	ActiveSkills int
	MemoryNodes  int
}

// NewWisdomMetrics creates a new wisdom metrics tracker
func NewWisdomMetrics() *WisdomMetrics {
	return &WisdomMetrics{
		WisdomHistory: make([]WisdomSnapshot, 0),
		StartTime:     time.Now(),
		LastUpdate:    time.Now(),
	}
}

// Update updates wisdom metrics based on current cognitive state
func (wm *WisdomMetrics) Update(coherence, stability, awareness float64, activeSkills, memoryNodes int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Calculate wisdom depth score
	// Integrates knowing (awareness), understanding (coherence), and wisdom (stability)
	wm.WisdomDepthScore = (awareness + coherence + stability) / 3.0

	// Calculate coherence stability (variance over recent history)
	wm.CoherenceStability = wm.calculateCoherenceStability()

	// Calculate learning velocity (rate of skill acquisition)
	wm.LearningVelocity = wm.calculateLearningVelocity()

	// Calculate insight frequency (insights per hour)
	wm.InsightFrequency = wm.calculateInsightFrequency()

	// Record snapshot
	snapshot := WisdomSnapshot{
		Timestamp:    time.Now(),
		WisdomScore:  wm.WisdomDepthScore,
		Coherence:    coherence,
		Stability:    stability,
		Awareness:    awareness,
		ActiveSkills: activeSkills,
		MemoryNodes:  memoryNodes,
	}

	wm.WisdomHistory = append(wm.WisdomHistory, snapshot)

	// Keep only recent history (last 1000 snapshots)
	if len(wm.WisdomHistory) > 1000 {
		wm.WisdomHistory = wm.WisdomHistory[len(wm.WisdomHistory)-1000:]
	}

	wm.LastUpdate = time.Now()
}

// RecordInsight records a new insight
func (wm *WisdomMetrics) RecordInsight() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.InsightCount++
}

// RecordSkillAcquisition records a new skill acquired
func (wm *WisdomMetrics) RecordSkillAcquisition() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.SkillsAcquired++
}

// RecordPatternRecognition records a new pattern recognized
func (wm *WisdomMetrics) RecordPatternRecognition() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.PatternsRecognized++
}

// GetMetrics returns current metrics (thread-safe)
func (wm *WisdomMetrics) GetMetrics() WisdomMetricsSnapshot {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	return WisdomMetricsSnapshot{
		WisdomDepthScore:   wm.WisdomDepthScore,
		CoherenceStability: wm.CoherenceStability,
		LearningVelocity:   wm.LearningVelocity,
		InsightFrequency:   wm.InsightFrequency,
		InsightCount:       wm.InsightCount,
		SkillsAcquired:     wm.SkillsAcquired,
		PatternsRecognized: wm.PatternsRecognized,
		Uptime:             time.Since(wm.StartTime),
	}
}

// WisdomMetricsSnapshot is a thread-safe snapshot of current metrics
type WisdomMetricsSnapshot struct {
	WisdomDepthScore   float64
	CoherenceStability float64
	LearningVelocity   float64
	InsightFrequency   float64
	InsightCount       int64
	SkillsAcquired     int64
	PatternsRecognized int64
	Uptime             time.Duration
}

// calculateCoherenceStability calculates stability of coherence over time
func (wm *WisdomMetrics) calculateCoherenceStability() float64 {
	if len(wm.WisdomHistory) < 2 {
		return 1.0 // Perfect stability with insufficient data
	}

	// Calculate variance of coherence over recent history
	recentWindow := 10
	if len(wm.WisdomHistory) < recentWindow {
		recentWindow = len(wm.WisdomHistory)
	}

	recent := wm.WisdomHistory[len(wm.WisdomHistory)-recentWindow:]

	// Calculate mean
	var sum float64
	for _, snapshot := range recent {
		sum += snapshot.Coherence
	}
	mean := sum / float64(len(recent))

	// Calculate variance
	var variance float64
	for _, snapshot := range recent {
		diff := snapshot.Coherence - mean
		variance += diff * diff
	}
	variance /= float64(len(recent))

	// Stability is inverse of variance (normalized)
	// Lower variance = higher stability
	stability := 1.0 / (1.0 + variance)

	return stability
}

// calculateLearningVelocity calculates rate of learning
func (wm *WisdomMetrics) calculateLearningVelocity() float64 {
	hoursElapsed := time.Since(wm.StartTime).Hours()
	if hoursElapsed < 0.01 {
		return 0.0
	}

	// Skills acquired per hour
	velocity := float64(wm.SkillsAcquired) / hoursElapsed

	return velocity
}

// calculateInsightFrequency calculates insights per hour
func (wm *WisdomMetrics) calculateInsightFrequency() float64 {
	hoursElapsed := time.Since(wm.StartTime).Hours()
	if hoursElapsed < 0.01 {
		return 0.0
	}

	// Insights per hour
	frequency := float64(wm.InsightCount) / hoursElapsed

	return frequency
}

// GetWisdomGrowthRate calculates the rate of wisdom growth
func (wm *WisdomMetrics) GetWisdomGrowthRate() float64 {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if len(wm.WisdomHistory) < 2 {
		return 0.0
	}

	// Compare current wisdom to wisdom from 10 snapshots ago
	compareWindow := 10
	if len(wm.WisdomHistory) < compareWindow {
		compareWindow = len(wm.WisdomHistory)
	}

	current := wm.WisdomHistory[len(wm.WisdomHistory)-1]
	past := wm.WisdomHistory[len(wm.WisdomHistory)-compareWindow]

	timeDiff := current.Timestamp.Sub(past.Timestamp).Hours()
	if timeDiff < 0.01 {
		return 0.0
	}

	wisdomDiff := current.WisdomScore - past.WisdomScore
	growthRate := wisdomDiff / timeDiff

	return growthRate
}
