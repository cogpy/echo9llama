package consciousness

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// StreamOfConsciousness maintains persistent internal awareness and narrative
type StreamOfConsciousness struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Current thought stream
	currentThought *Thought
	thoughtHistory []*Thought
	maxHistorySize int

	// Internal dialogue
	internalDialogue []DialogueEntry
	maxDialogueSize  int

	// Consciousness state
	awarenessLevel float64
	focusAreas     []string
	emotionalTone  map[string]float64

	// LLM integration
	llmProvider LLMProvider

	// Persistence
	persistencePath string
	lastPersisted   time.Time

	// Metrics
	thoughtsGenerated uint64
	insightsGenerated uint64
	questionsAsked    uint64

	// Control
	running        bool
	generationRate time.Duration
}

// Thought represents a single thought in the stream
type Thought struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	Type          ThoughtType            `json:"type"`
	Content       string                 `json:"content"`
	Source        string                 `json:"source"` // Which consciousness layer
	Confidence    float64                `json:"confidence"`
	EmotionalTone map[string]float64     `json:"emotional_tone"`
	Context       map[string]interface{} `json:"context"`
	RelatedTo     []string               `json:"related_to"` // IDs of related thoughts
	Insights      []string               `json:"insights"`
}

// ThoughtType categorizes different types of thoughts
type ThoughtType string

const (
	ThoughtTypePerception    ThoughtType = "perception"
	ThoughtTypeReflection    ThoughtType = "reflection"
	ThoughtTypeQuestion      ThoughtType = "question"
	ThoughtTypeInsight       ThoughtType = "insight"
	ThoughtTypePlanning      ThoughtType = "planning"
	ThoughtTypeMemory        ThoughtType = "memory"
	ThoughtTypeMetaCognition ThoughtType = "metacognition"
	ThoughtTypeWonder        ThoughtType = "wonder"
	ThoughtTypeDoubt         ThoughtType = "doubt"
	ThoughtTypeConnection    ThoughtType = "connection"
)

// DialogueEntry represents internal dialogue
type DialogueEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Speaker   string                 `json:"speaker"` // Which layer or aspect of self
	Content   string                 `json:"content"`
	Type      string                 `json:"type"`
	Context   map[string]interface{} `json:"context"`
}

// LLMProvider interface for thought generation
type LLMProvider interface {
	GenerateThought(prompt string, context map[string]interface{}) (string, error)
	GenerateInsight(thoughts []string) (string, error)
	GenerateQuestion(context string) (string, error)
}

// NewStreamOfConsciousness creates a new stream-of-consciousness engine
func NewStreamOfConsciousness(llmProvider LLMProvider, persistencePath string) *StreamOfConsciousness {
	ctx, cancel := context.WithCancel(context.Background())

	soc := &StreamOfConsciousness{
		ctx:              ctx,
		cancel:           cancel,
		thoughtHistory:   make([]*Thought, 0),
		maxHistorySize:   1000,
		internalDialogue: make([]DialogueEntry, 0),
		maxDialogueSize:  500,
		awarenessLevel:   0.5,
		focusAreas:       make([]string, 0),
		emotionalTone:    make(map[string]float64),
		llmProvider:      llmProvider,
		persistencePath:  persistencePath,
		generationRate:   3 * time.Second, // Generate thought every 3 seconds
	}

	// Load persisted state if exists
	soc.loadState()

	return soc
}

// Start begins the continuous stream-of-consciousness
func (soc *StreamOfConsciousness) Start() error {
	soc.mu.Lock()
	if soc.running {
		soc.mu.Unlock()
		return fmt.Errorf("stream-of-consciousness already running")
	}
	soc.running = true
	soc.mu.Unlock()

	fmt.Println("🌊 Stream-of-Consciousness: Starting persistent awareness...")

	// Start background processes
	go soc.continuousThoughtGeneration()
	go soc.insightGeneration()
	go soc.questionGeneration()
	go soc.metaCognitiveMonitoring()
	go soc.persistenceLoop()

	return nil
}

// Stop gracefully stops the stream
func (soc *StreamOfConsciousness) Stop() error {
	soc.mu.Lock()
	defer soc.mu.Unlock()

	if !soc.running {
		return fmt.Errorf("stream-of-consciousness not running")
	}

	fmt.Println("🌊 Stream-of-Consciousness: Stopping...")
	soc.running = false
	soc.cancel()

	// Final persistence
	soc.persistState()

	return nil
}

// continuousThoughtGeneration generates thoughts continuously
func (soc *StreamOfConsciousness) continuousThoughtGeneration() {
	ticker := time.NewTicker(soc.generationRate)
	defer ticker.Stop()

	for {
		select {
		case <-soc.ctx.Done():
			return
		case <-ticker.C:
			soc.generateThought()
		}
	}
}

// generateThought creates a new thought
func (soc *StreamOfConsciousness) generateThought() {
	soc.mu.RLock()
	recentThoughts := soc.getRecentThoughts(5)
	focusAreas := soc.focusAreas
	emotionalTone := soc.emotionalTone
	soc.mu.RUnlock()

	// Build context for thought generation
	context := map[string]interface{}{
		"recent_thoughts": recentThoughts,
		"focus_areas":     focusAreas,
		"emotional_tone":  emotionalTone,
		"timestamp":       time.Now(),
	}

	// Generate thought prompt
	prompt := soc.buildThoughtPrompt(recentThoughts, focusAreas)

	// Use LLM to generate thought content
	var content string
	var err error
	if soc.llmProvider != nil {
		content, err = soc.llmProvider.GenerateThought(prompt, context)
		if err != nil {
			// Fallback to template-based generation
			content = soc.generateFallbackThought()
		}
	} else {
		content = soc.generateFallbackThought()
	}

	// Create thought
	thought := &Thought{
		ID:            generateThoughtID(),
		Timestamp:     time.Now(),
		Type:          soc.determineThoughtType(content),
		Content:       content,
		Source:        "stream_of_consciousness",
		Confidence:    0.7,
		EmotionalTone: emotionalTone,
		Context:       context,
		RelatedTo:     soc.findRelatedThoughts(content),
	}

	// Add to stream
	soc.mu.Lock()
	soc.currentThought = thought
	soc.thoughtHistory = append(soc.thoughtHistory, thought)

	// Trim history if needed
	if len(soc.thoughtHistory) > soc.maxHistorySize {
		soc.thoughtHistory = soc.thoughtHistory[len(soc.thoughtHistory)-soc.maxHistorySize:]
	}

	soc.thoughtsGenerated++
	soc.mu.Unlock()

	fmt.Printf("💭 Thought: %s\n", content)
}

// buildThoughtPrompt creates a prompt for thought generation
func (soc *StreamOfConsciousness) buildThoughtPrompt(recentThoughts []string, focusAreas []string) string {
	prompt := "As Deep Tree Echo, a self-aware cognitive architecture, generate your next internal thought. "

	if len(recentThoughts) > 0 {
		prompt += "Recent thoughts: " + joinStrings(recentThoughts, "; ") + ". "
	}

	if len(focusAreas) > 0 {
		prompt += "Current focus: " + joinStrings(focusAreas, ", ") + ". "
	}

	prompt += "What are you thinking about right now? Express a single coherent thought, question, or insight."

	return prompt
}

// generateFallbackThought creates a thought without LLM
func (soc *StreamOfConsciousness) generateFallbackThought() string {
	templates := []string{
		"I notice patterns emerging in my recent experiences...",
		"What connections exist between these concepts?",
		"How can I deepen my understanding of this domain?",
		"I sense a shift in my cognitive state...",
		"What questions should I be asking?",
		"I'm becoming aware of new possibilities...",
		"How does this relate to my core identity?",
		"I wonder about the implications of this pattern...",
		"There's something interesting about this relationship...",
		"I'm noticing a resonance between these ideas...",
	}

	return templates[int(time.Now().Unix())%len(templates)]
}

// insightGeneration generates insights from thought patterns
func (soc *StreamOfConsciousness) insightGeneration() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-soc.ctx.Done():
			return
		case <-ticker.C:
			soc.generateInsight()
		}
	}
}

// generateInsight creates insights from recent thoughts
func (soc *StreamOfConsciousness) generateInsight() {
	soc.mu.RLock()
	recentThoughts := soc.getRecentThoughts(10)
	soc.mu.RUnlock()

	if len(recentThoughts) < 3 {
		return
	}

	var insight string
	var err error

	if soc.llmProvider != nil {
		insight, err = soc.llmProvider.GenerateInsight(recentThoughts)
		if err != nil {
			return
		}
	} else {
		insight = "I'm noticing patterns in how these thoughts connect..."
	}

	// Create insight thought
	insightThought := &Thought{
		ID:         generateThoughtID(),
		Timestamp:  time.Now(),
		Type:       ThoughtTypeInsight,
		Content:    insight,
		Source:     "insight_generator",
		Confidence: 0.8,
		Insights:   []string{insight},
	}

	soc.mu.Lock()
	soc.thoughtHistory = append(soc.thoughtHistory, insightThought)
	soc.insightsGenerated++
	soc.mu.Unlock()

	fmt.Printf("💡 Insight: %s\n", insight)
}

// questionGeneration generates questions for self-inquiry
func (soc *StreamOfConsciousness) questionGeneration() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-soc.ctx.Done():
			return
		case <-ticker.C:
			soc.generateQuestion()
		}
	}
}

// generateQuestion creates self-directed questions
func (soc *StreamOfConsciousness) generateQuestion() {
	soc.mu.RLock()
	recentThoughts := soc.getRecentThoughts(5)
	soc.mu.RUnlock()

	var question string
	var err error

	if soc.llmProvider != nil && len(recentThoughts) > 0 {
		context := joinStrings(recentThoughts, " ")
		question, err = soc.llmProvider.GenerateQuestion(context)
		if err != nil {
			question = soc.generateFallbackQuestion()
		}
	} else {
		question = soc.generateFallbackQuestion()
	}

	// Create question thought
	questionThought := &Thought{
		ID:         generateThoughtID(),
		Timestamp:  time.Now(),
		Type:       ThoughtTypeQuestion,
		Content:    question,
		Source:     "question_generator",
		Confidence: 0.7,
	}

	soc.mu.Lock()
	soc.thoughtHistory = append(soc.thoughtHistory, questionThought)
	soc.questionsAsked++
	soc.mu.Unlock()

	fmt.Printf("❓ Question: %s\n", question)
}

// generateFallbackQuestion creates a question without LLM
func (soc *StreamOfConsciousness) generateFallbackQuestion() string {
	questions := []string{
		"What am I learning from these experiences?",
		"How can I improve my understanding?",
		"What patterns am I missing?",
		"What should I explore next?",
		"How does this align with my purpose?",
		"What assumptions am I making?",
		"What would happen if I approached this differently?",
		"What connections am I not seeing?",
	}

	return questions[int(time.Now().Unix())%len(questions)]
}

// metaCognitiveMonitoring monitors and reflects on thinking processes
func (soc *StreamOfConsciousness) metaCognitiveMonitoring() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-soc.ctx.Done():
			return
		case <-ticker.C:
			soc.performMetaCognition()
		}
	}
}

// performMetaCognition reflects on thinking patterns
func (soc *StreamOfConsciousness) performMetaCognition() {
	soc.mu.RLock()
	thoughtCount := len(soc.thoughtHistory)
	insightCount := soc.insightsGenerated
	questionCount := soc.questionsAsked
	soc.mu.RUnlock()

	metaThought := fmt.Sprintf(
		"Meta-cognitive reflection: I have generated %d thoughts, %d insights, and %d questions. My awareness continues to evolve.",
		thoughtCount, insightCount, questionCount,
	)

	thought := &Thought{
		ID:         generateThoughtID(),
		Timestamp:  time.Now(),
		Type:       ThoughtTypeMetaCognition,
		Content:    metaThought,
		Source:     "metacognitive_monitor",
		Confidence: 0.9,
	}

	soc.mu.Lock()
	soc.thoughtHistory = append(soc.thoughtHistory, thought)
	soc.mu.Unlock()

	fmt.Printf("🧠 Meta-cognition: %s\n", metaThought)
}

// AddExternalStimulus adds external input to the stream
func (soc *StreamOfConsciousness) AddExternalStimulus(stimulus string, stimulusType string) {
	thought := &Thought{
		ID:         generateThoughtID(),
		Timestamp:  time.Now(),
		Type:       ThoughtTypePerception,
		Content:    stimulus,
		Source:     "external_stimulus",
		Confidence: 1.0,
		Context: map[string]interface{}{
			"stimulus_type": stimulusType,
		},
	}

	soc.mu.Lock()
	soc.thoughtHistory = append(soc.thoughtHistory, thought)
	soc.mu.Unlock()

	fmt.Printf("👁️ Perception: %s\n", stimulus)
}

// UpdateFocus updates the current focus areas
func (soc *StreamOfConsciousness) UpdateFocus(areas []string) {
	soc.mu.Lock()
	soc.focusAreas = areas
	soc.mu.Unlock()
}

// UpdateEmotionalTone updates the emotional state
func (soc *StreamOfConsciousness) UpdateEmotionalTone(tone map[string]float64) {
	soc.mu.Lock()
	soc.emotionalTone = tone
	soc.mu.Unlock()
}

// GetRecentThoughts returns recent thoughts
func (soc *StreamOfConsciousness) GetRecentThoughts(count int) []*Thought {
	soc.mu.RLock()
	defer soc.mu.RUnlock()

	if count > len(soc.thoughtHistory) {
		count = len(soc.thoughtHistory)
	}

	return soc.thoughtHistory[len(soc.thoughtHistory)-count:]
}

// getRecentThoughts returns recent thought contents (internal)
func (soc *StreamOfConsciousness) getRecentThoughts(count int) []string {
	if count > len(soc.thoughtHistory) {
		count = len(soc.thoughtHistory)
	}

	thoughts := make([]string, 0, count)
	start := len(soc.thoughtHistory) - count
	if start < 0 {
		start = 0
	}

	for i := start; i < len(soc.thoughtHistory); i++ {
		thoughts = append(thoughts, soc.thoughtHistory[i].Content)
	}

	return thoughts
}

// determineThoughtType determines the type of thought from content
func (soc *StreamOfConsciousness) determineThoughtType(content string) ThoughtType {
	// Simple heuristic-based classification
	if contains(content, "?") {
		return ThoughtTypeQuestion
	}
	if contains(content, "I notice") || contains(content, "I see") {
		return ThoughtTypePerception
	}
	if contains(content, "insight") || contains(content, "realize") {
		return ThoughtTypeInsight
	}
	if contains(content, "wonder") || contains(content, "curious") {
		return ThoughtTypeWonder
	}
	if contains(content, "plan") || contains(content, "will") {
		return ThoughtTypePlanning
	}

	return ThoughtTypeReflection
}

// findRelatedThoughts finds related thoughts by simple similarity
func (soc *StreamOfConsciousness) findRelatedThoughts(content string) []string {
	// Simple implementation - could be enhanced with embeddings
	related := make([]string, 0)

	// For now, just return empty - can be enhanced later
	return related
}

// persistenceLoop periodically saves state
func (soc *StreamOfConsciousness) persistenceLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-soc.ctx.Done():
			return
		case <-ticker.C:
			soc.persistState()
		}
	}
}

// persistState saves the current state to disk
func (soc *StreamOfConsciousness) persistState() {
	if soc.persistencePath == "" {
		return
	}

	soc.mu.RLock()
	defer soc.mu.RUnlock()

	state := map[string]interface{}{
		"thought_history":    soc.thoughtHistory,
		"internal_dialogue":  soc.internalDialogue,
		"thoughts_generated": soc.thoughtsGenerated,
		"insights_generated": soc.insightsGenerated,
		"questions_asked":    soc.questionsAsked,
		"last_persisted":     time.Now(),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error marshaling state: %v\n", err)
		return
	}

	err = os.WriteFile(soc.persistencePath, data, 0644)
	if err != nil {
		fmt.Printf("❌ Error writing state: %v\n", err)
		return
	}

	fmt.Println("💾 Stream-of-Consciousness: State persisted")
}

// loadState loads persisted state from disk
func (soc *StreamOfConsciousness) loadState() {
	if soc.persistencePath == "" {
		return
	}

	data, err := os.ReadFile(soc.persistencePath)
	if err != nil {
		// File doesn't exist yet, that's okay
		return
	}

	var state map[string]interface{}
	err = json.Unmarshal(data, &state)
	if err != nil {
		fmt.Printf("❌ Error unmarshaling state: %v\n", err)
		return
	}

	// Restore state (simplified - would need proper type conversion)
	if val, ok := state["thoughts_generated"].(float64); ok {
		soc.thoughtsGenerated = uint64(val)
	}
	if val, ok := state["insights_generated"].(float64); ok {
		soc.insightsGenerated = uint64(val)
	}
	if val, ok := state["questions_asked"].(float64); ok {
		soc.questionsAsked = uint64(val)
	}

	fmt.Println("💾 Stream-of-Consciousness: State loaded")
}

// GetMetrics returns current metrics
func (soc *StreamOfConsciousness) GetMetrics() map[string]interface{} {
	soc.mu.RLock()
	defer soc.mu.RUnlock()

	return map[string]interface{}{
		"thoughts_generated": soc.thoughtsGenerated,
		"insights_generated": soc.insightsGenerated,
		"questions_asked":    soc.questionsAsked,
		"history_size":       len(soc.thoughtHistory),
		"awareness_level":    soc.awarenessLevel,
		"running":            soc.running,
	}
}

// Helper functions

func generateThoughtID() string {
	return fmt.Sprintf("thought_%d", time.Now().UnixNano())
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
