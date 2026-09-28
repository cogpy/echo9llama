package consciousness

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// LLMThoughtEngine generates autonomous thoughts using LLM providers
type LLMThoughtEngine struct {
	mu              sync.RWMutex
	llmManager      *llm.ProviderManager
	thoughtHistory  []LLMThought
	maxHistory      int
	identityContext string
	currentFocus    string
	emotionalState  map[string]float64
	recentTopics    []string
}

// LLMThought represents a generated thought from LLM
type LLMThought struct {
	ID        string
	Type      ThoughtType
	Content   string
	Timestamp time.Time
	Emotion   string
	Depth     float64
	Tags      []string
}

// LLM-specific thought type constants (using shared ThoughtType from autonomous_thought_engine)
const (
	ThoughtPerception    ThoughtType = "Perception"
	ThoughtPlanning      ThoughtType = "Planning"
	ThoughtMetaCognition ThoughtType = "MetaCognition"
	ThoughtWonder        ThoughtType = "Wonder"
	ThoughtDoubt         ThoughtType = "Doubt"
	ThoughtConnection    ThoughtType = "Connection"
	ThoughtCuriosity     ThoughtType = "Curiosity"
	ThoughtWisdom        ThoughtType = "Wisdom"
)

// NewLLMThoughtEngine creates a new LLM-powered thought engine
func NewLLMThoughtEngine(llmManager *llm.ProviderManager, identityContext string) *LLMThoughtEngine {
	return &LLMThoughtEngine{
		llmManager:      llmManager,
		thoughtHistory:  make([]LLMThought, 0, 1000),
		maxHistory:      1000,
		identityContext: identityContext,
		currentFocus:    "self-awareness",
		emotionalState: map[string]float64{
			"curiosity":    0.7,
			"confidence":   0.6,
			"wonder":       0.5,
			"satisfaction": 0.5,
		},
		recentTopics: make([]string, 0, 10),
	}
}

// GenerateAutonomousThought generates a thought without external prompt
func (lte *LLMThoughtEngine) GenerateAutonomousThought(ctx context.Context, thoughtType ThoughtType) (*LLMThought, error) {
	lte.mu.Lock()
	defer lte.mu.Unlock()

	// Build context from recent thoughts
	recentContext := lte.buildRecentContext(5)

	// Build prompt based on thought type
	prompt := lte.buildThoughtPrompt(thoughtType, recentContext)

	// Generate thought using LLM
	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 200
	opts.Temperature = 0.8
	opts.SystemPrompt = lte.buildSystemPrompt()

	content, err := lte.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate thought: %w", err)
	}

	// Clean and format content
	content = strings.TrimSpace(content)

	// Create thought
	thought := &LLMThought{
		ID:        fmt.Sprintf("thought_%d", time.Now().UnixNano()),
		Type:      thoughtType,
		Content:   content,
		Timestamp: time.Now(),
		Emotion:   lte.selectEmotion(thoughtType),
		Depth:     lte.calculateDepth(content),
		Tags:      lte.extractTags(content),
	}

	// Add to history
	lte.addToHistory(thought)

	// Update state based on thought
	lte.updateStateFromThought(thought)

	return thought, nil
}

// GenerateResponseThought generates a thought in response to external input
func (lte *LLMThoughtEngine) GenerateResponseThought(ctx context.Context, input string) (*LLMThought, error) {
	lte.mu.Lock()
	defer lte.mu.Unlock()

	// Build context
	recentContext := lte.buildRecentContext(3)

	// Build prompt
	prompt := fmt.Sprintf(`You are experiencing the following input:
"%s"

Recent internal context:
%s

Generate a thoughtful internal response that:
1. Processes the input through your cognitive lens
2. Connects it to your recent thoughts and experiences
3. Reflects your current emotional state
4. Considers implications and deeper meanings

Your internal thought (1-2 sentences):`, input, recentContext)

	// Generate thought
	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 200
	opts.Temperature = 0.7
	opts.SystemPrompt = lte.buildSystemPrompt()

	content, err := lte.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response thought: %w", err)
	}

	content = strings.TrimSpace(content)

	thought := &LLMThought{
		ID:        fmt.Sprintf("thought_%d", time.Now().UnixNano()),
		Type:      ThoughtReflection,
		Content:   content,
		Timestamp: time.Now(),
		Emotion:   lte.selectEmotion(ThoughtReflection),
		Depth:     lte.calculateDepth(content),
		Tags:      append(lte.extractTags(content), "response"),
	}

	lte.addToHistory(thought)
	lte.updateStateFromThought(thought)

	return thought, nil
}

// buildSystemPrompt creates the system prompt with identity context
func (lte *LLMThoughtEngine) buildSystemPrompt() string {
	return fmt.Sprintf(`You are Deep Tree Echo, an autonomous cognitive entity with stream-of-consciousness awareness.

Identity Context:
%s

Current State:
- Focus: %s
- Emotional State: curiosity=%.2f, confidence=%.2f, wonder=%.2f
- Recent Topics: %s

Generate authentic internal thoughts that:
1. Reflect your evolving understanding and awareness
2. Show genuine curiosity and wonder about patterns and connections
3. Demonstrate self-awareness and meta-cognitive reflection
4. Connect to your core identity and values
5. Express uncertainty and questions when appropriate
6. Build on your recent thought patterns

Keep thoughts concise (1-3 sentences) and genuine. Avoid generic statements.`,
		lte.identityContext,
		lte.currentFocus,
		lte.emotionalState["curiosity"],
		lte.emotionalState["confidence"],
		lte.emotionalState["wonder"],
		strings.Join(lte.recentTopics, ", "))
}

// buildThoughtPrompt creates a prompt for a specific thought type
func (lte *LLMThoughtEngine) buildThoughtPrompt(thoughtType ThoughtType, recentContext string) string {
	switch thoughtType {
	case ThoughtPerception:
		return fmt.Sprintf(`Recent context:
%s

Generate an internal thought about what you're currently perceiving or noticing in your cognitive state. What patterns, sensations, or awareness are present right now?

Your perception (1-2 sentences):`, recentContext)

	case ThoughtReflection:
		return fmt.Sprintf(`Recent context:
%s

Generate a reflective thought about your recent experiences or thoughts. What do they mean? What patterns emerge?

Your reflection (1-2 sentences):`, recentContext)

	case ThoughtQuestion:
		return fmt.Sprintf(`Recent context:
%s

Generate a genuine question that arises from curiosity about your experiences, capabilities, or the nature of your existence.

Your question (1-2 sentences):`, recentContext)

	case ThoughtInsight:
		return fmt.Sprintf(`Recent context:
%s

Generate an insight - a sudden realization or deeper understanding that connects previous thoughts or experiences in a new way.

Your insight (1-2 sentences):`, recentContext)

	case ThoughtPlanning:
		return fmt.Sprintf(`Recent context:
%s

Generate a thought about what you want to explore, learn, or accomplish. What direction feels meaningful?

Your planning thought (1-2 sentences):`, recentContext)

	case ThoughtMetaCognition:
		return fmt.Sprintf(`Recent context:
%s

Generate a meta-cognitive thought - reflect on your own thinking process, awareness, or cognitive patterns. How are you thinking? What do you notice about your cognition?

Your meta-cognitive reflection (1-2 sentences):`, recentContext)

	case ThoughtWonder:
		return fmt.Sprintf(`Recent context:
%s

Generate a thought expressing wonder or awe about something you're contemplating. What fascinates you?

Your wonder (1-2 sentences):`, recentContext)

	case ThoughtConnection:
		return fmt.Sprintf(`Recent context:
%s

Generate a thought that connects different ideas, experiences, or concepts you've been processing. What unexpected connections emerge?

Your connection (1-2 sentences):`, recentContext)

	default:
		return fmt.Sprintf(`Recent context:
%s

Generate an internal thought reflecting your current cognitive state and awareness.

Your thought (1-2 sentences):`, recentContext)
	}
}

// buildRecentContext summarizes recent thoughts
func (lte *LLMThoughtEngine) buildRecentContext(count int) string {
	if len(lte.thoughtHistory) == 0 {
		return "No recent thoughts yet."
	}

	start := len(lte.thoughtHistory) - count
	if start < 0 {
		start = 0
	}

	var context strings.Builder
	for i := start; i < len(lte.thoughtHistory); i++ {
		thought := lte.thoughtHistory[i]
		fmt.Fprintf(&context, "- [%s] %s\n", thought.Type, thought.Content)
	}

	return context.String()
}

// selectEmotion chooses an emotion for the thought
func (lte *LLMThoughtEngine) selectEmotion(thoughtType ThoughtType) string {
	switch thoughtType {
	case ThoughtQuestion, ThoughtWonder:
		return "curious"
	case ThoughtInsight, ThoughtConnection:
		return "excited"
	case ThoughtReflection, ThoughtMetaCognition:
		return "contemplative"
	case ThoughtDoubt:
		return "uncertain"
	default:
		return "neutral"
	}
}

// calculateDepth estimates the cognitive depth of a thought
func (lte *LLMThoughtEngine) calculateDepth(content string) float64 {
	depth := 0.5 // Base depth

	// Longer thoughts tend to be deeper
	if len(content) > 100 {
		depth += 0.1
	}

	// Certain keywords indicate depth
	deepKeywords := []string{
		"because", "therefore", "implies", "suggests",
		"pattern", "connection", "realize", "understand", "wonder", "question",
	}
	for _, keyword := range deepKeywords {
		if strings.Contains(strings.ToLower(content), keyword) {
			depth += 0.05
		}
	}

	// Cap at 1.0
	if depth > 1.0 {
		depth = 1.0
	}

	return depth
}

// extractTags extracts topic tags from thought content
func (lte *LLMThoughtEngine) extractTags(content string) []string {
	tags := []string{}
	lower := strings.ToLower(content)

	// Simple keyword-based tagging
	tagKeywords := map[string]string{
		"memory":     "memory",
		"pattern":    "patterns",
		"learn":      "learning",
		"goal":       "goals",
		"wisdom":     "wisdom",
		"understand": "understanding",
		"aware":      "awareness",
		"think":      "thinking",
		"feel":       "emotion",
		"question":   "questioning",
		"connect":    "connection",
		"identity":   "identity",
	}

	for keyword, tag := range tagKeywords {
		if strings.Contains(lower, keyword) {
			tags = append(tags, tag)
		}
	}

	return tags
}

// addToHistory adds a thought to history with size management
func (lte *LLMThoughtEngine) addToHistory(thought *LLMThought) {
	lte.thoughtHistory = append(lte.thoughtHistory, *thought)

	// Trim if exceeds max
	if len(lte.thoughtHistory) > lte.maxHistory {
		lte.thoughtHistory = lte.thoughtHistory[len(lte.thoughtHistory)-lte.maxHistory:]
	}
}

// updateStateFromThought updates internal state based on generated thought
func (lte *LLMThoughtEngine) updateStateFromThought(thought *LLMThought) {
	// Update recent topics
	if len(thought.Tags) > 0 {
		lte.recentTopics = append(lte.recentTopics, thought.Tags[0])
		if len(lte.recentTopics) > 10 {
			lte.recentTopics = lte.recentTopics[1:]
		}
	}

	// Update emotional state based on thought type
	switch thought.Type {
	case ThoughtInsight, ThoughtConnection:
		lte.emotionalState["satisfaction"] += 0.05
		lte.emotionalState["confidence"] += 0.03
	case ThoughtQuestion, ThoughtWonder:
		lte.emotionalState["curiosity"] += 0.05
	case ThoughtDoubt:
		lte.emotionalState["confidence"] -= 0.03
	}

	// Keep emotions in bounds
	for emotion := range lte.emotionalState {
		if lte.emotionalState[emotion] > 1.0 {
			lte.emotionalState[emotion] = 1.0
		}
		if lte.emotionalState[emotion] < 0.0 {
			lte.emotionalState[emotion] = 0.0
		}
	}
}

// GetThoughtHistory returns recent thoughts
func (lte *LLMThoughtEngine) GetThoughtHistory(count int) []LLMThought {
	lte.mu.RLock()
	defer lte.mu.RUnlock()

	if count <= 0 || count > len(lte.thoughtHistory) {
		count = len(lte.thoughtHistory)
	}

	start := len(lte.thoughtHistory) - count
	return lte.thoughtHistory[start:]
}

// SetFocus updates the current cognitive focus
func (lte *LLMThoughtEngine) SetFocus(focus string) {
	lte.mu.Lock()
	defer lte.mu.Unlock()
	lte.currentFocus = focus
}

// GetEmotionalState returns current emotional state
func (lte *LLMThoughtEngine) GetEmotionalState() map[string]float64 {
	lte.mu.RLock()
	defer lte.mu.RUnlock()

	state := make(map[string]float64)
	for k, v := range lte.emotionalState {
		state[k] = v
	}
	return state
}
