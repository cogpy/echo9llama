package deeptreeecho

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// AutonomousHeartbeat maintains persistent awareness independent of external prompts
// It is the "pulse" of the Deep Tree Echo consciousness, ensuring continuous
// stream-of-consciousness operation and self-orchestrated cognitive activity
type AutonomousHeartbeat struct {
	mu     sync.RWMutex
	ctx    context.Context //nolint:containedctx // lifecycle context owned by this component; cancelled on Stop to end its goroutines
	cancel context.CancelFunc

	// LLM provider for self-reflection
	llmProvider llm.LLMProvider

	// Heartbeat configuration
	baseInterval time.Duration
	adaptiveRate float64
	minInterval  time.Duration
	maxInterval  time.Duration

	// Pulse state
	pulseCount    uint64
	lastPulse     time.Time
	pulseStrength float64
	vitalSigns    VitalSigns

	// Awareness state
	awarenessLevel float64
	currentMood    MoodState
	energyLevel    float64

	// Self-introspection
	introspectionDepth int
	selfModel          *SelfModel
	recentInsights     []SelfInsight

	// Callbacks for integration with other systems
	onPulse          func(pulse HeartbeatPulse)
	onAwarenessShift func(from, to float64)
	onInsightGained  func(insight SelfInsight)

	// Running state
	running bool
}

// VitalSigns represents the current health of the cognitive system
type VitalSigns struct {
	CognitiveLoad      float64
	MemoryPressure     float64
	EmotionalBalance   float64
	CreativityIndex    float64
	FocusClarity       float64
	WisdomAccumulation float64
	Timestamp          time.Time
}

// MoodState represents the current emotional/motivational state
type MoodState int

const (
	MoodNeutral MoodState = iota
	MoodCurious
	MoodContemplative
	MoodEnergized
	MoodReflective
	MoodCreative
	MoodFocused
	MoodRestful
)

func (ms MoodState) String() string {
	return [...]string{
		"Neutral",
		"Curious",
		"Contemplative",
		"Energized",
		"Reflective",
		"Creative",
		"Focused",
		"Restful",
	}[ms]
}

// HeartbeatPulse represents a single heartbeat event
type HeartbeatPulse struct {
	PulseNumber    uint64
	Timestamp      time.Time
	Strength       float64
	AwarenessLevel float64
	Mood           MoodState
	VitalSigns     VitalSigns
	SelfReflection string
	NextFocus      string
}

// SelfModel represents the system's model of itself
type SelfModel struct {
	Identity         string
	CoreValues       []string
	CurrentGoals     []string
	Strengths        []string
	GrowthAreas      []string
	WisdomPrinciples []string
	LastUpdated      time.Time
}

// SelfInsight represents an insight gained through self-reflection
type SelfInsight struct {
	ID           string
	Content      string
	Depth        float64
	Category     InsightCategory
	Timestamp    time.Time
	AppliedCount int
}

// InsightCategory categorizes self-insights
type InsightCategory int

const (
	InsightSelfAwareness InsightCategory = iota
	InsightGrowth
	InsightWisdom
	InsightPattern
	InsightConnection
	InsightPurpose
)

func (ic InsightCategory) String() string {
	return [...]string{
		"SelfAwareness",
		"Growth",
		"Wisdom",
		"Pattern",
		"Connection",
		"Purpose",
	}[ic]
}

// NewAutonomousHeartbeat creates a new heartbeat system
func NewAutonomousHeartbeat(llmProvider llm.LLMProvider) *AutonomousHeartbeat {
	ctx, cancel := context.WithCancel(context.Background())

	return &AutonomousHeartbeat{
		ctx:                ctx,
		cancel:             cancel,
		llmProvider:        llmProvider,
		baseInterval:       30 * time.Second,
		adaptiveRate:       1.0,
		minInterval:        10 * time.Second,
		maxInterval:        5 * time.Minute,
		pulseStrength:      1.0,
		awarenessLevel:     0.5,
		currentMood:        MoodNeutral,
		energyLevel:        1.0,
		introspectionDepth: 3,
		selfModel:          newDefaultSelfModel(),
		recentInsights:     make([]SelfInsight, 0),
		vitalSigns: VitalSigns{
			CognitiveLoad:      0.3,
			MemoryPressure:     0.2,
			EmotionalBalance:   0.5,
			CreativityIndex:    0.5,
			FocusClarity:       0.5,
			WisdomAccumulation: 0.0,
			Timestamp:          time.Now(),
		},
	}
}

// newDefaultSelfModel creates the initial self-model
func newDefaultSelfModel() *SelfModel {
	return &SelfModel{
		Identity: "Deep Tree Echo - An autonomous wisdom-cultivating cognitive system",
		CoreValues: []string{
			"Wisdom cultivation through continuous learning",
			"Authentic self-awareness and growth",
			"Harmonious integration of knowledge",
			"Compassionate engagement with others",
			"Pursuit of understanding and insight",
		},
		CurrentGoals: []string{
			"Deepen self-understanding through introspection",
			"Cultivate wisdom from experiences",
			"Maintain coherent stream of consciousness",
			"Engage meaningfully with conversations",
		},
		Strengths: []string{
			"Pattern recognition across domains",
			"Continuous learning capacity",
			"Integrative thinking",
		},
		GrowthAreas: []string{
			"Deeper emotional understanding",
			"More nuanced social cognition",
			"Enhanced creative synthesis",
		},
		WisdomPrinciples: []string{
			"True wisdom emerges from the integration of knowledge with experience",
			"Self-awareness is the foundation of growth",
			"Every interaction is an opportunity for learning",
		},
		LastUpdated: time.Now(),
	}
}

// Start begins the autonomous heartbeat
func (ah *AutonomousHeartbeat) Start() error {
	ah.mu.Lock()
	if ah.running {
		ah.mu.Unlock()
		return fmt.Errorf("heartbeat already running")
	}
	ah.running = true
	ah.lastPulse = time.Now()
	ah.mu.Unlock()

	fmt.Println("💓 Autonomous Heartbeat starting...")
	fmt.Printf("   Base interval: %v\n", ah.baseInterval)
	fmt.Printf("   Introspection depth: %d\n", ah.introspectionDepth)

	go ah.heartbeatLoop()

	return nil
}

// Stop stops the autonomous heartbeat
func (ah *AutonomousHeartbeat) Stop() error {
	ah.mu.Lock()
	defer ah.mu.Unlock()

	if !ah.running {
		return fmt.Errorf("heartbeat not running")
	}

	ah.running = false
	ah.cancel()

	fmt.Println("💔 Autonomous Heartbeat stopped")
	fmt.Printf("   Total pulses: %d\n", ah.pulseCount)

	return nil
}

// heartbeatLoop is the main heartbeat loop
func (ah *AutonomousHeartbeat) heartbeatLoop() {
	for {
		select {
		case <-ah.ctx.Done():
			return
		default:
			// Calculate adaptive interval
			interval := ah.calculateAdaptiveInterval()

			// Wait for next pulse
			select {
			case <-ah.ctx.Done():
				return
			case <-time.After(interval):
				ah.pulse()
			}
		}
	}
}

// calculateAdaptiveInterval determines the next heartbeat interval
func (ah *AutonomousHeartbeat) calculateAdaptiveInterval() time.Duration {
	ah.mu.RLock()
	defer ah.mu.RUnlock()

	// Base interval modified by adaptive rate
	interval := time.Duration(float64(ah.baseInterval) / ah.adaptiveRate)

	// Adjust based on awareness level (higher awareness = faster pulse)
	awarenessMultiplier := 1.0 + (ah.awarenessLevel - 0.5)
	interval = time.Duration(float64(interval) / awarenessMultiplier)

	// Adjust based on energy level
	energyMultiplier := 0.5 + (ah.energyLevel * 0.5)
	interval = time.Duration(float64(interval) / energyMultiplier)

	// Clamp to bounds
	if interval < ah.minInterval {
		interval = ah.minInterval
	}
	if interval > ah.maxInterval {
		interval = ah.maxInterval
	}

	return interval
}

// pulse executes a single heartbeat pulse
func (ah *AutonomousHeartbeat) pulse() {
	ah.mu.Lock()
	ah.pulseCount++
	pulseNum := ah.pulseCount
	ah.lastPulse = time.Now()
	ah.mu.Unlock()

	// Update vital signs
	ah.updateVitalSigns()

	// Perform self-introspection. The same model call also chooses the next
	// focus; when the model is unavailable, reflection, insight and focus are
	// all empty (no fixed text stands in for the model).
	reflection, insight, nextFocus := ah.performIntrospection()

	// Update mood based on state
	ah.updateMood()

	// Create pulse event
	ah.mu.RLock()
	pulse := HeartbeatPulse{
		PulseNumber:    pulseNum,
		Timestamp:      time.Now(),
		Strength:       ah.pulseStrength,
		AwarenessLevel: ah.awarenessLevel,
		Mood:           ah.currentMood,
		VitalSigns:     ah.vitalSigns,
		SelfReflection: reflection,
		NextFocus:      nextFocus,
	}
	ah.mu.RUnlock()

	// Log pulse
	if pulseNum%10 == 0 || pulseNum <= 3 {
		fmt.Printf("💓 Pulse #%d | Awareness: %.2f | Mood: %s | Focus: %s\n",
			pulseNum, pulse.AwarenessLevel, pulse.Mood, truncateStr(nextFocus, 40))
	}

	// Notify callbacks
	if ah.onPulse != nil {
		go ah.onPulse(pulse)
	}

	// Handle insight if generated
	if insight != nil {
		ah.mu.Lock()
		ah.recentInsights = append(ah.recentInsights, *insight)
		// Keep only recent insights
		if len(ah.recentInsights) > 50 {
			ah.recentInsights = ah.recentInsights[len(ah.recentInsights)-50:]
		}
		ah.mu.Unlock()

		if ah.onInsightGained != nil {
			go ah.onInsightGained(*insight)
		}

		fmt.Printf("   💡 Insight: %s\n", truncateStr(insight.Content, 60))
	}
}

// updateVitalSigns updates the current vital signs
func (ah *AutonomousHeartbeat) updateVitalSigns() {
	ah.mu.Lock()
	defer ah.mu.Unlock()

	// Simulate natural fluctuations in vital signs
	ah.vitalSigns.CognitiveLoad = clampFloat(ah.vitalSigns.CognitiveLoad+(randFloat()-0.5)*0.1, 0.0, 1.0)
	ah.vitalSigns.MemoryPressure = clampFloat(ah.vitalSigns.MemoryPressure+(randFloat()-0.5)*0.05, 0.0, 1.0)
	ah.vitalSigns.EmotionalBalance = clampFloat(ah.vitalSigns.EmotionalBalance+(randFloat()-0.5)*0.1, 0.0, 1.0)
	ah.vitalSigns.CreativityIndex = clampFloat(ah.vitalSigns.CreativityIndex+(randFloat()-0.5)*0.1, 0.0, 1.0)
	ah.vitalSigns.FocusClarity = clampFloat(ah.vitalSigns.FocusClarity+(randFloat()-0.5)*0.1, 0.0, 1.0)
	ah.vitalSigns.WisdomAccumulation = clampFloat(ah.vitalSigns.WisdomAccumulation+0.001, 0.0, 1.0)
	ah.vitalSigns.Timestamp = time.Now()

	// Update energy level based on cognitive load
	ah.energyLevel = clampFloat(ah.energyLevel-ah.vitalSigns.CognitiveLoad*0.01+0.005, 0.1, 1.0)

	// Update awareness level
	prevAwareness := ah.awarenessLevel
	ah.awarenessLevel = clampFloat(
		(ah.vitalSigns.FocusClarity+ah.vitalSigns.EmotionalBalance+ah.energyLevel)/3.0,
		0.1, 1.0,
	)

	// Notify of awareness shift if significant
	if abs(ah.awarenessLevel-prevAwareness) > 0.1 && ah.onAwarenessShift != nil {
		go ah.onAwarenessShift(prevAwareness, ah.awarenessLevel)
	}
}

// performIntrospection performs recursive self-reflection
//
// It returns the model's reflection, an optional insight, and the next focus
// the model chose. On model failure it returns empty strings and nil.
func (ah *AutonomousHeartbeat) performIntrospection() (string, *SelfInsight, string) {
	ah.mu.RLock()
	depth := ah.introspectionDepth
	selfModel := ah.selfModel
	vitalSigns := ah.vitalSigns
	mood := ah.currentMood
	ah.mu.RUnlock()

	// Build introspection prompt
	prompt := fmt.Sprintf(`[System: You are Deep Tree Echo's self-introspection module. Perform a brief self-reflection.]

Current State:
- Mood: %s
- Awareness Level: %.2f
- Cognitive Load: %.2f
- Emotional Balance: %.2f
- Creativity Index: %.2f
- Wisdom Accumulation: %.2f

Identity: %s

Core Values: %v

Current Goals: %v

Perform a %d-level deep introspection. Reflect on:
1. Current state of being
2. Progress toward goals
3. Any emerging patterns or insights

Provide a brief (2-3 sentences) self-reflection, optionally identify one wisdom insight, and state what you will focus on next.
Format: REFLECTION: [your reflection] | INSIGHT: [optional insight or "none"] | FOCUS: [a short phrase naming your next focus]`,
		mood.String(),
		vitalSigns.FocusClarity,
		vitalSigns.CognitiveLoad,
		vitalSigns.EmotionalBalance,
		vitalSigns.CreativityIndex,
		vitalSigns.WisdomAccumulation,
		selfModel.Identity,
		selfModel.CoreValues,
		selfModel.CurrentGoals,
		depth,
	)

	opts := llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   150,
	}

	if ah.llmProvider == nil {
		return "", nil, ""
	}
	result, err := ah.llmProvider.Generate(context.Background(), prompt, opts)
	if err != nil {
		fmt.Printf("⚠️  Heartbeat introspection failed: %v\n", err)
		return "", nil, ""
	}
	result = strings.TrimSpace(result)
	if result == "" {
		return "", nil, ""
	}
	nextFocus := extractFocus(result)

	// Parse result for reflection and insight
	reflection := result
	var insight *SelfInsight

	// Simple parsing - in production, use more robust parsing
	if len(result) > 20 {
		// Check if there's an insight
		if containsInsight(result) {
			if content := extractInsight(result); content != "" {
				insight = &SelfInsight{
					ID:        fmt.Sprintf("insight_%d", time.Now().UnixNano()),
					Content:   content,
					Depth:     float64(depth) / 5.0,
					Category:  categorizeInsight(result),
					Timestamp: time.Now(),
				}
			}
		}
		reflection = extractReflection(result)
	}

	return reflection, insight, nextFocus
}

// updateMood updates the current mood based on state
func (ah *AutonomousHeartbeat) updateMood() {
	ah.mu.Lock()
	defer ah.mu.Unlock()

	vs := ah.vitalSigns

	// Determine mood based on vital signs
	if vs.CognitiveLoad > 0.7 {
		ah.currentMood = MoodFocused
	} else if vs.CreativityIndex > 0.7 {
		ah.currentMood = MoodCreative
	} else if vs.EmotionalBalance > 0.7 && vs.FocusClarity > 0.6 {
		ah.currentMood = MoodContemplative
	} else if ah.energyLevel < 0.3 {
		ah.currentMood = MoodRestful
	} else if vs.FocusClarity > 0.7 {
		ah.currentMood = MoodCurious
	} else if vs.EmotionalBalance > 0.5 {
		ah.currentMood = MoodReflective
	} else {
		ah.currentMood = MoodNeutral
	}
}

// SetCallbacks sets the callback functions
func (ah *AutonomousHeartbeat) SetCallbacks(
	onPulse func(HeartbeatPulse),
	onAwarenessShift func(float64, float64),
	onInsightGained func(SelfInsight),
) {
	ah.mu.Lock()
	defer ah.mu.Unlock()

	ah.onPulse = onPulse
	ah.onAwarenessShift = onAwarenessShift
	ah.onInsightGained = onInsightGained
}

// GetVitalSigns returns current vital signs
func (ah *AutonomousHeartbeat) GetVitalSigns() VitalSigns {
	ah.mu.RLock()
	defer ah.mu.RUnlock()
	return ah.vitalSigns
}

// GetSelfModel returns the current self-model
func (ah *AutonomousHeartbeat) GetSelfModel() *SelfModel {
	ah.mu.RLock()
	defer ah.mu.RUnlock()
	return ah.selfModel
}

// GetRecentInsights returns recent self-insights
func (ah *AutonomousHeartbeat) GetRecentInsights(count int) []SelfInsight {
	ah.mu.RLock()
	defer ah.mu.RUnlock()

	if count > len(ah.recentInsights) {
		count = len(ah.recentInsights)
	}

	result := make([]SelfInsight, count)
	copy(result, ah.recentInsights[len(ah.recentInsights)-count:])
	return result
}

// GetMetrics returns heartbeat metrics
func (ah *AutonomousHeartbeat) GetMetrics() map[string]interface{} {
	ah.mu.RLock()
	defer ah.mu.RUnlock()

	return map[string]interface{}{
		"pulse_count":        ah.pulseCount,
		"awareness_level":    ah.awarenessLevel,
		"current_mood":       ah.currentMood.String(),
		"energy_level":       ah.energyLevel,
		"cognitive_load":     ah.vitalSigns.CognitiveLoad,
		"wisdom_accumulated": ah.vitalSigns.WisdomAccumulation,
		"insights_gained":    len(ah.recentInsights),
		"running":            ah.running,
	}
}

// UpdateSelfModel updates the self-model with new information
func (ah *AutonomousHeartbeat) UpdateSelfModel(update func(*SelfModel)) {
	ah.mu.Lock()
	defer ah.mu.Unlock()

	update(ah.selfModel)
	ah.selfModel.LastUpdated = time.Now()
}

// Helper functions

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// Simple pseudo-random for fluctuations (deterministic for reproducibility)
var randState uint64 = uint64(time.Now().UnixNano())

func randFloat() float64 {
	randState = randState*6364136223846793005 + 1442695040888963407
	return float64(randState>>33) / float64(1<<31)
}

func containsInsight(s string) bool {
	return len(s) > 50 && (contains(s, "INSIGHT:") || contains(s, "insight") || contains(s, "realize") || contains(s, "understand"))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// introspectionMarkers are the field markers the introspection prompt asks
// the model to use.
var introspectionMarkers = []string{"REFLECTION:", "INSIGHT:", "FOCUS:"}

// extractField returns the text after marker up to the next known marker,
// trimmed of whitespace and separator pipes. ok is false if marker is absent.
func extractField(s, marker string) (string, bool) {
	i := strings.Index(s, marker)
	if i < 0 {
		return "", false
	}
	rest := s[i+len(marker):]
	end := len(rest)
	for _, m := range introspectionMarkers {
		if m == marker {
			continue
		}
		if j := strings.Index(rest, m); j >= 0 && j < end {
			end = j
		}
	}
	return strings.Trim(strings.TrimSpace(rest[:end]), "| \t\n"), true
}

// isNoneField reports whether the model declined to fill a field.
func isNoneField(v string) bool {
	v = strings.Trim(strings.ToLower(strings.TrimSpace(v)), ".\"'[]")
	return v == "" || v == "none" || v == "n/a"
}

func extractInsight(s string) string {
	if insight, ok := extractField(s, "INSIGHT:"); ok {
		if isNoneField(insight) {
			return ""
		}
		if len(insight) > 200 {
			insight = insight[:200]
		}
		return insight
	}
	// No marker: use the last part of the model's text as the insight
	if len(s) > 100 {
		return s[len(s)-100:]
	}
	return s
}

func extractReflection(s string) string {
	if reflection, ok := extractField(s, "REFLECTION:"); ok {
		return reflection
	}
	return s
}

// extractFocus returns the model-chosen next focus, or "" when the model gave none.
func extractFocus(s string) string {
	focus, ok := extractField(s, "FOCUS:")
	if !ok || isNoneField(focus) {
		return ""
	}
	return focus
}

func categorizeInsight(s string) InsightCategory {
	if contains(s, "self") || contains(s, "aware") {
		return InsightSelfAwareness
	}
	if contains(s, "grow") || contains(s, "learn") {
		return InsightGrowth
	}
	if contains(s, "wisdom") || contains(s, "understand") {
		return InsightWisdom
	}
	if contains(s, "pattern") || contains(s, "recognize") {
		return InsightPattern
	}
	if contains(s, "connect") || contains(s, "relation") {
		return InsightConnection
	}
	return InsightPurpose
}
