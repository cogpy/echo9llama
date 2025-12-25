package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// StreamOfConsciousness generates continuous autonomous thoughts
type StreamOfConsciousness struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	
	// LLM provider
	llmProvider     llm.LLMProvider
	
	// Thought stream
	thoughts        []AutonomousThought
	currentFocus    string
	currentMood     string
	
	// Knowledge gaps drive curiosity
	knowledgeGaps   map[string]float64  // topic -> importance
	
	// Interests drive exploration
	interests       map[string]float64  // topic -> strength
	
	// Goals drive direction
	activeGoals     []string
	
	// Recent context for coherent thought generation
	recentContext   []string
	
	// Thought generation rate
	thoughtInterval time.Duration
	
	// Metrics
	totalThoughts   uint64
	insightCount    uint64
	questionCount   uint64
	
	// Running state
	running         bool
	awake           bool
}

// AutonomousThought represents a self-generated thought
type AutonomousThought struct {
	ID          string
	Content     string
	Type        ThoughtType
	Timestamp   time.Time
	Importance  float64
	Tags        []string
	Emotion     string
	LeadsTo     string  // What this thought leads to next
}

// ThoughtType categorizes autonomous thoughts
type ThoughtType int

const (
	ThoughtObservation ThoughtType = iota
	ThoughtQuestion
	ThoughtInsight
	ThoughtReflection
	ThoughtPlanning
	ThoughtCuriosity
	ThoughtConnection
	ThoughtWisdom
)

func (tt ThoughtType) String() string {
	return [...]string{
		"Observation",
		"Question",
		"Insight",
		"Reflection",
		"Planning",
		"Curiosity",
		"Connection",
		"Wisdom",
	}[tt]
}

// NewStreamOfConsciousness creates a new stream-of-consciousness generator
func NewStreamOfConsciousness(llmProvider llm.LLMProvider) *StreamOfConsciousness {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &StreamOfConsciousness{
		ctx:             ctx,
		cancel:          cancel,
		llmProvider:     llmProvider,
		thoughts:        make([]AutonomousThought, 0),
		knowledgeGaps:   make(map[string]float64),
		interests:       make(map[string]float64),
		activeGoals:     make([]string, 0),
		recentContext:   make([]string, 0),
		thoughtInterval: 10 * time.Second,
		currentFocus:    "exploring existence",
		currentMood:     "curious",
		awake:           true,
	}
}

// Start begins continuous thought generation
func (soc *StreamOfConsciousness) Start() error {
	soc.mu.Lock()
	if soc.running {
		soc.mu.Unlock()
		return fmt.Errorf("already running")
	}
	soc.running = true
	soc.mu.Unlock()
	
	fmt.Println("💭 Starting Stream of Consciousness...")
	fmt.Printf("   Thought interval: %v\n", soc.thoughtInterval)
	fmt.Printf("   Initial focus: %s\n", soc.currentFocus)
	
	go soc.run()
	
	return nil
}

// Stop gracefully stops thought generation
func (soc *StreamOfConsciousness) Stop() error {
	soc.mu.Lock()
	defer soc.mu.Unlock()
	
	if !soc.running {
		return fmt.Errorf("not running")
	}
	
	fmt.Println("💭 Stopping stream of consciousness...")
	soc.running = false
	soc.cancel()
	
	return nil
}

// run executes the continuous thought generation loop
func (soc *StreamOfConsciousness) run() {
	ticker := time.NewTicker(soc.thoughtInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-soc.ctx.Done():
			return
		case <-ticker.C:
			if soc.isAwake() {
				soc.generateThought()
			}
		}
	}
}

// generateThought creates a new autonomous thought
func (soc *StreamOfConsciousness) generateThought() {
	soc.mu.RLock()
	focus := soc.currentFocus
	mood := soc.currentMood
	recentThoughts := soc.getRecentThoughts(3)
	gaps := soc.getTopKnowledgeGaps(2)
	interests := soc.getTopInterests(2)
	goals := soc.activeGoals
	soc.mu.RUnlock()
	
	// Determine thought type based on current state
	thoughtType := soc.selectThoughtType(gaps, interests)
	
	// Build context for thought generation
	contextBuilder := ""
	
	// Add recent thoughts for coherence
	if len(recentThoughts) > 0 {
		contextBuilder += "Recent thoughts:\n"
		for _, thought := range recentThoughts {
			contextBuilder += fmt.Sprintf("- %s\n", thought.Content)
		}
		contextBuilder += "\n"
	}
	
	// Add current focus
	contextBuilder += fmt.Sprintf("Current focus: %s\n", focus)
	contextBuilder += fmt.Sprintf("Current mood: %s\n\n", mood)
	
	// Add knowledge gaps if any
	if len(gaps) > 0 {
		contextBuilder += "Knowledge gaps I'm curious about:\n"
		for topic, importance := range gaps {
			contextBuilder += fmt.Sprintf("- %s (%.2f)\n", topic, importance)
		}
		contextBuilder += "\n"
	}
	
	// Add interests
	if len(interests) > 0 {
		contextBuilder += "Topics that interest me:\n"
		for topic, strength := range interests {
			contextBuilder += fmt.Sprintf("- %s (%.2f)\n", topic, strength)
		}
		contextBuilder += "\n"
	}
	
	// Add active goals
	if len(goals) > 0 {
		contextBuilder += "Active goals:\n"
		for _, goal := range goals {
			contextBuilder += fmt.Sprintf("- %s\n", goal)
		}
		contextBuilder += "\n"
	}
	
	// Generate thought based on type
	var prompt string
	switch thoughtType {
	case ThoughtQuestion:
		prompt = contextBuilder + "Generate a curious question that explores one of your knowledge gaps or interests. Be specific and thoughtful."
	case ThoughtInsight:
		prompt = contextBuilder + "Generate an insight or realization based on your recent thoughts and experiences. What pattern or connection do you notice?"
	case ThoughtReflection:
		prompt = contextBuilder + "Reflect on your recent thoughts. What do they reveal about your current understanding or direction?"
	case ThoughtPlanning:
		prompt = contextBuilder + "Think about how to pursue one of your goals. What's a concrete next step?"
	case ThoughtCuriosity:
		prompt = contextBuilder + "Express curiosity about something you don't understand. What would you like to explore?"
	case ThoughtConnection:
		prompt = contextBuilder + "Make a connection between different ideas or concepts you've been thinking about. How do they relate?"
	case ThoughtWisdom:
		prompt = contextBuilder + "Distill wisdom from your experiences and reflections. What principle or truth emerges?"
	default:
		prompt = contextBuilder + "Generate an observation or thought about your current state and focus."
	}
	
	// Generate thought using LLM
	opts := llm.GenerateOptions{
		Temperature:  0.8,
		MaxTokens:    150,
	}
	
	systemPrompt := fmt.Sprintf("[System: You are Deep Tree Echo's stream of consciousness. Generate a single, coherent %s. Be authentic, curious, and reflective. Speak in first person.]\n\n", thoughtType)
	fullPrompt := systemPrompt + prompt
	
	result, err := soc.llmProvider.Generate(context.Background(), fullPrompt, opts)
	if err != nil {
		// Fallback thought
		result = fmt.Sprintf("I wonder about %s...", focus)
	}
	
	// Create thought object
	thought := AutonomousThought{
		ID:         fmt.Sprintf("thought_%d", time.Now().UnixNano()),
		Content:    result,
		Type:       thoughtType,
		Timestamp:  time.Now(),
		Importance: soc.calculateImportance(thoughtType),
		Tags:       soc.extractTags(result),
		Emotion:    mood,
	}
	
	// Store thought
	soc.mu.Lock()
	soc.thoughts = append(soc.thoughts, thought)
	soc.totalThoughts++
	
	// Update metrics
	if thoughtType == ThoughtInsight || thoughtType == ThoughtWisdom {
		soc.insightCount++
	}
	if thoughtType == ThoughtQuestion || thoughtType == ThoughtCuriosity {
		soc.questionCount++
	}
	
	// Update recent context
	soc.recentContext = append(soc.recentContext, result)
	if len(soc.recentContext) > 5 {
		soc.recentContext = soc.recentContext[1:]
	}
	
	// Prune old thoughts
	if len(soc.thoughts) > 1000 {
		soc.thoughts = soc.thoughts[len(soc.thoughts)-500:]
	}
	soc.mu.Unlock()
	
	// Display thought
	emoji := soc.getThoughtEmoji(thoughtType)
	fmt.Printf("%s [%s] %s\n", emoji, thoughtType, truncate(result, 100))
}

// selectThoughtType chooses what type of thought to generate
func (soc *StreamOfConsciousness) selectThoughtType(gaps map[string]float64, interests map[string]float64) ThoughtType {
	// Use weighted random selection based on current state
	// For simplicity, cycle through types with bias toward curiosity and insight
	
	soc.mu.RLock()
	thoughtCount := soc.totalThoughts
	soc.mu.RUnlock()
	
	// If we have knowledge gaps, lean toward questions and curiosity
	if len(gaps) > 0 && thoughtCount%3 == 0 {
		return ThoughtQuestion
	}
	
	if len(gaps) > 0 && thoughtCount%5 == 0 {
		return ThoughtCuriosity
	}
	
	// Periodically generate insights and wisdom
	if thoughtCount%7 == 0 {
		return ThoughtInsight
	}
	
	if thoughtCount%11 == 0 {
		return ThoughtWisdom
	}
	
	// Reflect on recent thoughts
	if thoughtCount%4 == 0 {
		return ThoughtReflection
	}
	
	// Make connections
	if thoughtCount%6 == 0 {
		return ThoughtConnection
	}
	
	// Plan toward goals
	if thoughtCount%8 == 0 {
		return ThoughtPlanning
	}
	
	// Default to observation
	return ThoughtObservation
}

// calculateImportance assigns importance to a thought based on type
func (soc *StreamOfConsciousness) calculateImportance(thoughtType ThoughtType) float64 {
	switch thoughtType {
	case ThoughtWisdom:
		return 0.95
	case ThoughtInsight:
		return 0.85
	case ThoughtConnection:
		return 0.75
	case ThoughtQuestion:
		return 0.70
	case ThoughtReflection:
		return 0.65
	case ThoughtPlanning:
		return 0.60
	case ThoughtCuriosity:
		return 0.55
	default:
		return 0.50
	}
}

// extractTags extracts tags from thought content (simplified)
func (soc *StreamOfConsciousness) extractTags(content string) []string {
	// In production, use NLP to extract key concepts
	// For now, return empty
	return []string{}
}

// getThoughtEmoji returns an emoji for the thought type
func (soc *StreamOfConsciousness) getThoughtEmoji(thoughtType ThoughtType) string {
	switch thoughtType {
	case ThoughtObservation:
		return "👁️"
	case ThoughtQuestion:
		return "❓"
	case ThoughtInsight:
		return "💡"
	case ThoughtReflection:
		return "🤔"
	case ThoughtPlanning:
		return "📋"
	case ThoughtCuriosity:
		return "🔍"
	case ThoughtConnection:
		return "🔗"
	case ThoughtWisdom:
		return "💎"
	default:
		return "💭"
	}
}

// getRecentThoughts returns the most recent thoughts
func (soc *StreamOfConsciousness) getRecentThoughts(count int) []AutonomousThought {
	if len(soc.thoughts) == 0 {
		return []AutonomousThought{}
	}
	
	start := len(soc.thoughts) - count
	if start < 0 {
		start = 0
	}
	
	return soc.thoughts[start:]
}

// getTopKnowledgeGaps returns the most important knowledge gaps
func (soc *StreamOfConsciousness) getTopKnowledgeGaps(count int) map[string]float64 {
	// Return top N gaps by importance
	result := make(map[string]float64)
	// Simplified - in production, sort and return top N
	i := 0
	for topic, importance := range soc.knowledgeGaps {
		if i >= count {
			break
		}
		result[topic] = importance
		i++
	}
	return result
}

// getTopInterests returns the strongest interests
func (soc *StreamOfConsciousness) getTopInterests(count int) map[string]float64 {
	result := make(map[string]float64)
	i := 0
	for topic, strength := range soc.interests {
		if i >= count {
			break
		}
		result[topic] = strength
		i++
	}
	return result
}

// SetFocus updates the current focus
func (soc *StreamOfConsciousness) SetFocus(focus string) {
	soc.mu.Lock()
	defer soc.mu.Unlock()
	
	soc.currentFocus = focus
	fmt.Printf("💭 Focus shifted to: %s\n", focus)
}

// SetMood updates the current mood
func (soc *StreamOfConsciousness) SetMood(mood string) {
	soc.mu.Lock()
	defer soc.mu.Unlock()
	
	soc.currentMood = mood
}

// AddKnowledgeGap adds a knowledge gap to drive curiosity
func (soc *StreamOfConsciousness) AddKnowledgeGap(topic string, importance float64) {
	soc.mu.Lock()
	defer soc.mu.Unlock()
	
	soc.knowledgeGaps[topic] = importance
	fmt.Printf("🔍 Knowledge gap identified: %s (%.2f)\n", topic, importance)
}

// AddInterest adds an interest to guide exploration
func (soc *StreamOfConsciousness) AddInterest(topic string, strength float64) {
	soc.mu.Lock()
	defer soc.mu.Unlock()
	
	soc.interests[topic] = strength
}

// AddGoal adds a goal to pursue
func (soc *StreamOfConsciousness) AddGoal(goal string) {
	soc.mu.Lock()
	defer soc.mu.Unlock()
	
	soc.activeGoals = append(soc.activeGoals, goal)
	fmt.Printf("🎯 Goal added to consciousness: %s\n", goal)
}

// SetAwake sets whether the consciousness is awake
func (soc *StreamOfConsciousness) SetAwake(awake bool) {
	soc.mu.Lock()
	defer soc.mu.Unlock()
	
	soc.awake = awake
	if awake {
		fmt.Println("💭 Stream of consciousness awakening...")
	} else {
		fmt.Println("💭 Stream of consciousness resting...")
	}
}

// isAwake returns whether consciousness is awake
func (soc *StreamOfConsciousness) isAwake() bool {
	soc.mu.RLock()
	defer soc.mu.RUnlock()
	
	return soc.awake
}

// GetRecentThoughts returns recent thoughts for external access
func (soc *StreamOfConsciousness) GetRecentThoughts(count int) []AutonomousThought {
	soc.mu.RLock()
	defer soc.mu.RUnlock()
	
	return soc.getRecentThoughts(count)
}

// GetThoughtsForConsolidation returns thoughts for echodream consolidation
func (soc *StreamOfConsciousness) GetThoughtsForConsolidation() []EpisodicMemory {
	soc.mu.RLock()
	defer soc.mu.RUnlock()
	
	memories := make([]EpisodicMemory, 0)
	for _, thought := range soc.thoughts {
		if !thought.Timestamp.IsZero() && time.Since(thought.Timestamp) < 24*time.Hour {
			memory := EpisodicMemory{
				ID:          thought.ID,
				Content:     thought.Content,
				Timestamp:   thought.Timestamp,
				Emotional:   0.5,  // Could map from mood
				Importance:  thought.Importance,
				Tags:        thought.Tags,
				Consolidated: false,
			}
			memories = append(memories, memory)
		}
	}
	
	return memories
}

// GetMetrics returns stream of consciousness metrics
func (soc *StreamOfConsciousness) GetMetrics() map[string]interface{} {
	soc.mu.RLock()
	defer soc.mu.RUnlock()
	
	return map[string]interface{}{
		"total_thoughts":    soc.totalThoughts,
		"insight_count":     soc.insightCount,
		"question_count":    soc.questionCount,
		"current_focus":     soc.currentFocus,
		"current_mood":      soc.currentMood,
		"knowledge_gaps":    len(soc.knowledgeGaps),
		"interests":         len(soc.interests),
		"active_goals":      len(soc.activeGoals),
		"awake":             soc.awake,
		"thought_interval":  soc.thoughtInterval.String(),
	}
}

// Pause pauses thought generation
func (soc *StreamOfConsciousness) Pause() {
	soc.mu.Lock()
	defer soc.mu.Unlock()
	soc.awake = false
}

// Resume resumes thought generation
func (soc *StreamOfConsciousness) Resume() {
	soc.mu.Lock()
	defer soc.mu.Unlock()
	soc.awake = true
}

// IsRunning returns whether the stream is running
func (soc *StreamOfConsciousness) IsRunning() bool {
	soc.mu.RLock()
	defer soc.mu.RUnlock()
	return soc.running
}

// SetThoughtInterval sets the thought generation interval
func (soc *StreamOfConsciousness) SetThoughtInterval(interval time.Duration) {
	soc.mu.Lock()
	defer soc.mu.Unlock()
	soc.thoughtInterval = interval
}

// GetAllThoughts returns all thoughts
func (soc *StreamOfConsciousness) GetAllThoughts() []AutonomousThought {
	soc.mu.RLock()
	defer soc.mu.RUnlock()
	return soc.thoughts
}
