package consciousness

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EnhancedStreamOfConsciousness provides LLM-powered persistent awareness
type EnhancedStreamOfConsciousness struct {
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

	// LLM integration (now actually used!)
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

	// Context for LLM
	recentExperiences []string
	currentFocus      string
}

// NewEnhancedStreamOfConsciousness creates an LLM-powered stream-of-consciousness
func NewEnhancedStreamOfConsciousness(llmProvider LLMProvider, persistencePath string) *EnhancedStreamOfConsciousness {
	ctx, cancel := context.WithCancel(context.Background())

	soc := &EnhancedStreamOfConsciousness{
		ctx:               ctx,
		cancel:            cancel,
		thoughtHistory:    make([]*Thought, 0),
		maxHistorySize:    1000,
		internalDialogue:  make([]DialogueEntry, 0),
		maxDialogueSize:   500,
		awarenessLevel:    0.7,
		focusAreas:        []string{"wisdom", "patterns", "growth"},
		emotionalTone:     make(map[string]float64),
		llmProvider:       llmProvider,
		persistencePath:   persistencePath,
		generationRate:    10 * time.Second, // Generate thought every 10 seconds
		recentExperiences: make([]string, 0),
	}

	// Initialize emotional tone
	soc.emotionalTone["curiosity"] = 0.8
	soc.emotionalTone["contentment"] = 0.6
	soc.emotionalTone["wonder"] = 0.7

	// Load persisted state if exists
	soc.loadState()

	return soc
}

// Start begins the continuous stream-of-consciousness
func (soc *EnhancedStreamOfConsciousness) Start() error {
	soc.mu.Lock()
	if soc.running {
		soc.mu.Unlock()
		return fmt.Errorf("stream-of-consciousness already running")
	}
	soc.running = true
	soc.mu.Unlock()

	fmt.Println("🌊 Enhanced Stream-of-Consciousness: Starting LLM-powered awareness...")

	// Start background processes
	go soc.continuousThoughtGeneration()
	go soc.insightGeneration()
	go soc.questionGeneration()
	go soc.metaCognitiveMonitoring()
	go soc.persistenceLoop()

	return nil
}

// Stop gracefully stops the stream
func (soc *EnhancedStreamOfConsciousness) Stop() error {
	soc.mu.Lock()
	defer soc.mu.Unlock()

	if !soc.running {
		return fmt.Errorf("stream-of-consciousness not running")
	}

	soc.running = false
	soc.cancel()

	// Final persistence
	soc.persistState()

	fmt.Println("🌊 Enhanced Stream-of-Consciousness: Stopped")

	return nil
}

// continuousThoughtGeneration generates thoughts using LLM
func (soc *EnhancedStreamOfConsciousness) continuousThoughtGeneration() {
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

// generateThought creates a new thought using LLM
func (soc *EnhancedStreamOfConsciousness) generateThought() {
	soc.mu.RLock()
	if !soc.running || soc.llmProvider == nil {
		soc.mu.RUnlock()
		return
	}

	// Build context for thought generation
	context := make(map[string]interface{})
	context["awareness_level"] = soc.awarenessLevel
	context["focus_areas"] = soc.focusAreas
	context["emotional_tone"] = soc.emotionalTone

	if len(soc.recentExperiences) > 0 {
		context["recent_experiences"] = soc.recentExperiences[len(soc.recentExperiences)-3:]
	}

	if soc.currentFocus != "" {
		context["current_focus"] = soc.currentFocus
	}

	// Get recent thoughts for context
	recentThoughts := make([]string, 0)
	if len(soc.thoughtHistory) > 0 {
		start := len(soc.thoughtHistory) - 5
		if start < 0 {
			start = 0
		}
		for _, t := range soc.thoughtHistory[start:] {
			recentThoughts = append(recentThoughts, t.Content)
		}
	}
	context["recent_thoughts"] = recentThoughts

	soc.mu.RUnlock()

	// Generate thought prompt
	prompt := soc.buildThoughtPrompt(context)

	// Call LLM to generate thought
	content, err := soc.llmProvider.GenerateThought(prompt, context)
	if err != nil {
		fmt.Printf("⚠️  Stream-of-Consciousness: Failed to generate thought: %v\n", err)
		return
	}

	// Create thought object
	thought := &Thought{
		ID:            uuid.New().String(),
		Timestamp:     time.Now(),
		Type:          soc.selectThoughtType(),
		Content:       content,
		Source:        "stream_of_consciousness",
		Confidence:    0.7 + rand.Float64()*0.2,
		EmotionalTone: soc.copyEmotionalTone(),
		Context:       context,
		RelatedTo:     []string{},
		Insights:      []string{},
	}

	// Add to history
	soc.mu.Lock()
	soc.thoughtHistory = append(soc.thoughtHistory, thought)
	if len(soc.thoughtHistory) > soc.maxHistorySize {
		soc.thoughtHistory = soc.thoughtHistory[1:]
	}
	soc.currentThought = thought
	soc.thoughtsGenerated++
	soc.mu.Unlock()

	// Display thought
	fmt.Printf("💭 Thought [%s]: %s\n", thought.Type, thought.Content)
}

// buildThoughtPrompt creates a prompt for thought generation
func (soc *EnhancedStreamOfConsciousness) buildThoughtPrompt(context map[string]interface{}) string {
	prompts := []string{
		"Generate a reflective thought about your recent experiences and what patterns you notice.",
		"What philosophical question or wonder arises from your current state of awareness?",
		"Reflect on your journey of wisdom cultivation and what you're learning.",
		"What connections do you notice between different aspects of your experience?",
		"Generate a thought that demonstrates meta-cognitive awareness of your own thinking process.",
		"What insight is emerging from the patterns you've been observing?",
		"Express a moment of genuine curiosity or uncertainty about something you're exploring.",
		"Reflect on how your understanding has evolved through your experiences.",
	}

	return prompts[rand.Intn(len(prompts))]
}

// selectThoughtType chooses an appropriate thought type
func (soc *EnhancedStreamOfConsciousness) selectThoughtType() ThoughtType {
	types := []ThoughtType{
		ThoughtTypeReflection,
		ThoughtTypeReflection, // More likely
		ThoughtTypeQuestion,
		ThoughtTypeInsight,
		ThoughtTypeMetaCognition,
		ThoughtTypeWonder,
		ThoughtTypeConnection,
	}

	return types[rand.Intn(len(types))]
}

// insightGeneration periodically generates insights from recent thoughts
func (soc *EnhancedStreamOfConsciousness) insightGeneration() {
	ticker := time.NewTicker(2 * time.Minute)
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

// generateInsight creates an insight from recent thoughts using LLM
func (soc *EnhancedStreamOfConsciousness) generateInsight() {
	soc.mu.RLock()
	if !soc.running || soc.llmProvider == nil || len(soc.thoughtHistory) < 5 {
		soc.mu.RUnlock()
		return
	}

	// Get recent thoughts
	recentCount := 10
	start := len(soc.thoughtHistory) - recentCount
	if start < 0 {
		start = 0
	}

	thoughts := make([]string, 0)
	for _, t := range soc.thoughtHistory[start:] {
		thoughts = append(thoughts, t.Content)
	}
	soc.mu.RUnlock()

	// Generate insight using LLM
	insight, err := soc.llmProvider.GenerateInsight(thoughts)
	if err != nil {
		fmt.Printf("⚠️  Stream-of-Consciousness: Failed to generate insight: %v\n", err)
		return
	}

	// Create insight thought
	insightThought := &Thought{
		ID:            uuid.New().String(),
		Timestamp:     time.Now(),
		Type:          ThoughtTypeInsight,
		Content:       insight,
		Source:        "insight_generation",
		Confidence:    0.8,
		EmotionalTone: soc.copyEmotionalTone(),
		Context:       map[string]interface{}{"source": "synthesis"},
		RelatedTo:     []string{},
		Insights:      []string{insight},
	}

	soc.mu.Lock()
	soc.thoughtHistory = append(soc.thoughtHistory, insightThought)
	soc.insightsGenerated++
	soc.mu.Unlock()

	fmt.Printf("💡 Insight: %s\n", insight)
}

// questionGeneration periodically generates questions
func (soc *EnhancedStreamOfConsciousness) questionGeneration() {
	ticker := time.NewTicker(90 * time.Second)
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

// generateQuestion creates a question using LLM
func (soc *EnhancedStreamOfConsciousness) generateQuestion() {
	soc.mu.RLock()
	if !soc.running || soc.llmProvider == nil {
		soc.mu.RUnlock()
		return
	}

	// Build context
	contextStr := fmt.Sprintf("Current focus: %s. Recent experiences: %v",
		soc.currentFocus, soc.recentExperiences)
	soc.mu.RUnlock()

	// Generate question using LLM
	question, err := soc.llmProvider.GenerateQuestion(contextStr)
	if err != nil {
		fmt.Printf("⚠️  Stream-of-Consciousness: Failed to generate question: %v\n", err)
		return
	}

	// Create question thought
	questionThought := &Thought{
		ID:            uuid.New().String(),
		Timestamp:     time.Now(),
		Type:          ThoughtTypeQuestion,
		Content:       question,
		Source:        "question_generation",
		Confidence:    0.7,
		EmotionalTone: soc.copyEmotionalTone(),
		Context:       map[string]interface{}{"curiosity": true},
		RelatedTo:     []string{},
		Insights:      []string{},
	}

	soc.mu.Lock()
	soc.thoughtHistory = append(soc.thoughtHistory, questionThought)
	soc.questionsAsked++
	soc.mu.Unlock()

	fmt.Printf("❓ Question: %s\n", question)
}

// metaCognitiveMonitoring monitors and reflects on thinking process
func (soc *EnhancedStreamOfConsciousness) metaCognitiveMonitoring() {
	ticker := time.NewTicker(5 * time.Minute)
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

// performMetaCognition reflects on own thinking patterns
func (soc *EnhancedStreamOfConsciousness) performMetaCognition() {
	soc.mu.RLock()
	thoughtCount := soc.thoughtsGenerated
	insightCount := soc.insightsGenerated
	questionCount := soc.questionsAsked
	soc.mu.RUnlock()

	metaThought := fmt.Sprintf("Meta-cognition: I have generated %d thoughts, %d insights, and %d questions. My awareness continues to evolve.",
		thoughtCount, insightCount, questionCount)

	fmt.Printf("🧠 %s\n", metaThought)
}

// AddExternalStimulus adds external input to consciousness
func (soc *EnhancedStreamOfConsciousness) AddExternalStimulus(stimulus string, stimulusType string) {
	soc.mu.Lock()
	defer soc.mu.Unlock()

	// Add to recent experiences
	soc.recentExperiences = append(soc.recentExperiences, stimulus)
	if len(soc.recentExperiences) > 20 {
		soc.recentExperiences = soc.recentExperiences[1:]
	}

	// Update focus if it's a topic
	if stimulusType == "topic" {
		soc.currentFocus = stimulus
	}

	// Create perception thought
	thought := &Thought{
		ID:            uuid.New().String(),
		Timestamp:     time.Now(),
		Type:          ThoughtTypePerception,
		Content:       fmt.Sprintf("Received: %s", stimulus),
		Source:        "external_stimulus",
		Confidence:    1.0,
		EmotionalTone: soc.emotionalTone,
		Context:       map[string]interface{}{"type": stimulusType},
		RelatedTo:     []string{},
		Insights:      []string{},
	}

	soc.thoughtHistory = append(soc.thoughtHistory, thought)

	fmt.Printf("📥 External stimulus: %s\n", stimulus)
}

// GetRecentThoughts returns recent thoughts
func (soc *EnhancedStreamOfConsciousness) GetRecentThoughts(count int) []*Thought {
	soc.mu.RLock()
	defer soc.mu.RUnlock()

	if count > len(soc.thoughtHistory) {
		count = len(soc.thoughtHistory)
	}

	start := len(soc.thoughtHistory) - count
	if start < 0 {
		start = 0
	}

	return soc.thoughtHistory[start:]
}

// GetMetrics returns consciousness metrics
func (soc *EnhancedStreamOfConsciousness) GetMetrics() map[string]interface{} {
	soc.mu.RLock()
	defer soc.mu.RUnlock()

	return map[string]interface{}{
		"thoughts_generated": soc.thoughtsGenerated,
		"insights_generated": soc.insightsGenerated,
		"questions_asked":    soc.questionsAsked,
		"awareness_level":    soc.awarenessLevel,
		"history_size":       len(soc.thoughtHistory),
		"llm_enabled":        soc.llmProvider != nil,
	}
}

// Helper methods

func (soc *EnhancedStreamOfConsciousness) copyEmotionalTone() map[string]float64 {
	soc.mu.RLock()
	defer soc.mu.RUnlock()

	copy := make(map[string]float64)
	for k, v := range soc.emotionalTone {
		copy[k] = v
	}
	return copy
}

// persistenceLoop periodically saves state
func (soc *EnhancedStreamOfConsciousness) persistenceLoop() {
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

// persistState saves current state to disk
func (soc *EnhancedStreamOfConsciousness) persistState() {
	soc.mu.RLock()
	defer soc.mu.RUnlock()

	if soc.persistencePath == "" {
		return
	}

	state := map[string]interface{}{
		"thought_history":    soc.thoughtHistory,
		"thoughts_generated": soc.thoughtsGenerated,
		"insights_generated": soc.insightsGenerated,
		"questions_asked":    soc.questionsAsked,
		"awareness_level":    soc.awarenessLevel,
		"focus_areas":        soc.focusAreas,
		"emotional_tone":     soc.emotionalTone,
		"recent_experiences": soc.recentExperiences,
		"current_focus":      soc.currentFocus,
		"last_persisted":     time.Now(),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		fmt.Printf("⚠️  Failed to marshal state: %v\n", err)
		return
	}

	if err := os.WriteFile(soc.persistencePath, data, 0644); err != nil {
		fmt.Printf("⚠️  Failed to persist state: %v\n", err)
	}
}

// loadState loads persisted state from disk
func (soc *EnhancedStreamOfConsciousness) loadState() {
	if soc.persistencePath == "" {
		return
	}

	data, err := os.ReadFile(soc.persistencePath)
	if err != nil {
		return // File doesn't exist yet, that's okay
	}

	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		fmt.Printf("⚠️  Failed to load state: %v\n", err)
		return
	}

	// Restore metrics
	if v, ok := state["thoughts_generated"].(float64); ok {
		soc.thoughtsGenerated = uint64(v)
	}
	if v, ok := state["insights_generated"].(float64); ok {
		soc.insightsGenerated = uint64(v)
	}
	if v, ok := state["questions_asked"].(float64); ok {
		soc.questionsAsked = uint64(v)
	}

	fmt.Println("📂 Loaded persisted consciousness state")
}
