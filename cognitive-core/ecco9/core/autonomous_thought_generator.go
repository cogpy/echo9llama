package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ThoughtGenerator generates autonomous thoughts using LLM and context
type ThoughtGenerator struct {
	mu               sync.RWMutex
	llmProvider      LLMProvider
	workingMemory    []string
	interestPatterns map[string]float64
	thoughtHistory   []GeneratedThought
	maxHistorySize   int
	generationCount  int64
}

// LLMProvider interface for thought generation
type LLMProvider interface {
	GenerateThought(ctx context.Context, prompt string) (string, error)
	GenerateReflection(ctx context.Context, context string) (string, error)
	IsAvailable() bool
}

// GeneratedThought represents a generated thought with metadata
type GeneratedThought struct {
	Content    string
	Type       string
	Timestamp  time.Time
	Context    []string
	Interests  []string
	Importance float64
}

// NewThoughtGenerator creates a new thought generator
func NewThoughtGenerator(llmProvider LLMProvider) *ThoughtGenerator {
	return &ThoughtGenerator{
		llmProvider:      llmProvider,
		workingMemory:    make([]string, 0, 7), // 7-item working memory
		interestPatterns: make(map[string]float64),
		thoughtHistory:   make([]GeneratedThought, 0),
		maxHistorySize:   100,
	}
}

// GenerateAutonomousThought generates a thought based on current context
func (tg *ThoughtGenerator) GenerateAutonomousThought() (*GeneratedThought, error) {
	tg.mu.RLock()
	if tg.llmProvider == nil || !tg.llmProvider.IsAvailable() {
		tg.mu.RUnlock()
		return nil, fmt.Errorf("LLM provider not available")
	}
	tg.mu.RUnlock()

	// Build context-aware prompt
	prompt := tg.buildContextualPrompt()

	// Generate thought
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	content, err := tg.llmProvider.GenerateThought(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate thought: %w", err)
	}

	// Create thought object
	thought := &GeneratedThought{
		Content:    content,
		Type:       tg.determineThoughtType(content),
		Timestamp:  time.Now(),
		Context:    tg.getWorkingMemoryCopy(),
		Interests:  tg.getTopInterests(3),
		Importance: tg.calculateImportance(content),
	}

	// Add to history
	tg.addToHistory(thought)

	// Update working memory
	tg.updateWorkingMemory(content)

	tg.mu.Lock()
	tg.generationCount++
	tg.mu.Unlock()

	return thought, nil
}

// GenerateReflection generates a reflection on recent thoughts
func (tg *ThoughtGenerator) GenerateReflection() (*GeneratedThought, error) {
	tg.mu.RLock()
	if tg.llmProvider == nil || !tg.llmProvider.IsAvailable() {
		tg.mu.RUnlock()
		return nil, fmt.Errorf("LLM provider not available")
	}

	// Get recent thoughts for context
	recentThoughts := tg.getRecentThoughts(5)
	tg.mu.RUnlock()

	if len(recentThoughts) == 0 {
		return nil, fmt.Errorf("no recent thoughts to reflect on")
	}

	// Build reflection context
	contextStr := tg.buildReflectionContext(recentThoughts)

	// Generate reflection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	content, err := tg.llmProvider.GenerateReflection(ctx, contextStr)
	if err != nil {
		return nil, fmt.Errorf("failed to generate reflection: %w", err)
	}

	// Create thought object
	thought := &GeneratedThought{
		Content:    content,
		Type:       "reflection",
		Timestamp:  time.Now(),
		Context:    tg.getWorkingMemoryCopy(),
		Interests:  tg.getTopInterests(3),
		Importance: 0.8, // Reflections are generally important
	}

	// Add to history
	tg.addToHistory(thought)

	// Update working memory
	tg.updateWorkingMemory(content)

	return thought, nil
}

// buildContextualPrompt builds a prompt based on current context
func (tg *ThoughtGenerator) buildContextualPrompt() string {
	tg.mu.RLock()
	defer tg.mu.RUnlock()

	var sb strings.Builder

	// Add working memory context
	if len(tg.workingMemory) > 0 {
		sb.WriteString("Recent thoughts in working memory:\n")
		for i, mem := range tg.workingMemory {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, mem))
		}
		sb.WriteString("\n")
	}

	// Add interest patterns
	if len(tg.interestPatterns) > 0 {
		sb.WriteString("Current interests: ")
		interests := tg.getTopInterests(3)
		sb.WriteString(strings.Join(interests, ", "))
		sb.WriteString("\n\n")
	}

	// Add generation prompt
	sb.WriteString("Generate a new thought that builds on this context, explores a curiosity, or offers a fresh insight.")

	return sb.String()
}

// buildReflectionContext builds context for reflection
func (tg *ThoughtGenerator) buildReflectionContext(recentThoughts []GeneratedThought) string {
	var sb strings.Builder

	sb.WriteString("Recent thoughts:\n")
	for i, thought := range recentThoughts {
		sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, thought.Type, thought.Content))
	}

	return sb.String()
}

// determineThoughtType analyzes content to determine thought type
func (tg *ThoughtGenerator) determineThoughtType(content string) string {
	lower := strings.ToLower(content)

	if strings.Contains(lower, "wonder") || strings.Contains(lower, "curious") || strings.Contains(lower, "?") {
		return "curiosity"
	}
	if strings.Contains(lower, "realize") || strings.Contains(lower, "understand") || strings.Contains(lower, "insight") {
		return "insight"
	}
	if strings.Contains(lower, "reflect") || strings.Contains(lower, "consider") || strings.Contains(lower, "ponder") {
		return "reflection"
	}
	if strings.Contains(lower, "connect") || strings.Contains(lower, "relate") || strings.Contains(lower, "similar") {
		return "association"
	}

	return "observation"
}

// calculateImportance estimates the importance of a thought
func (tg *ThoughtGenerator) calculateImportance(content string) float64 {
	importance := 0.5 // Base importance

	// Longer thoughts tend to be more developed
	wordCount := len(strings.Fields(content))
	if wordCount > 20 {
		importance += 0.1
	}
	if wordCount > 40 {
		importance += 0.1
	}

	// Questions indicate curiosity
	if strings.Contains(content, "?") {
		importance += 0.15
	}

	// Keywords that indicate depth
	deepKeywords := []string{"wisdom", "insight", "realize", "understand", "pattern", "connection"}
	for _, keyword := range deepKeywords {
		if strings.Contains(strings.ToLower(content), keyword) {
			importance += 0.05
		}
	}

	// Cap at 1.0
	if importance > 1.0 {
		importance = 1.0
	}

	return importance
}

// updateWorkingMemory adds content to working memory (7-item limit)
func (tg *ThoughtGenerator) updateWorkingMemory(content string) {
	tg.mu.Lock()
	defer tg.mu.Unlock()

	// Add to working memory
	tg.workingMemory = append(tg.workingMemory, content)

	// Keep only last 7 items
	if len(tg.workingMemory) > 7 {
		tg.workingMemory = tg.workingMemory[len(tg.workingMemory)-7:]
	}
}

// getWorkingMemoryCopy returns a copy of working memory
func (tg *ThoughtGenerator) getWorkingMemoryCopy() []string {
	tg.mu.RLock()
	defer tg.mu.RUnlock()

	copy := make([]string, len(tg.workingMemory))
	for i, mem := range tg.workingMemory {
		copy[i] = mem
	}
	return copy
}

// addToHistory adds a thought to history
func (tg *ThoughtGenerator) addToHistory(thought *GeneratedThought) {
	tg.mu.Lock()
	defer tg.mu.Unlock()

	tg.thoughtHistory = append(tg.thoughtHistory, *thought)

	// Prune history if too large
	if len(tg.thoughtHistory) > tg.maxHistorySize {
		tg.thoughtHistory = tg.thoughtHistory[len(tg.thoughtHistory)-tg.maxHistorySize:]
	}
}

// getRecentThoughts returns recent thoughts
func (tg *ThoughtGenerator) getRecentThoughts(count int) []GeneratedThought {
	if len(tg.thoughtHistory) == 0 {
		return []GeneratedThought{}
	}

	start := len(tg.thoughtHistory) - count
	if start < 0 {
		start = 0
	}

	return tg.thoughtHistory[start:]
}

// AddInterest adds or updates an interest pattern
func (tg *ThoughtGenerator) AddInterest(topic string, strength float64) {
	tg.mu.Lock()
	defer tg.mu.Unlock()

	tg.interestPatterns[topic] = strength
}

// getTopInterests returns top N interests
func (tg *ThoughtGenerator) getTopInterests(n int) []string {
	if len(tg.interestPatterns) == 0 {
		return []string{}
	}

	// Simple selection of top interests
	interests := make([]string, 0, n)
	for topic := range tg.interestPatterns {
		interests = append(interests, topic)
		if len(interests) >= n {
			break
		}
	}

	return interests
}

// GetStats returns generation statistics
func (tg *ThoughtGenerator) GetStats() map[string]interface{} {
	tg.mu.RLock()
	defer tg.mu.RUnlock()

	return map[string]interface{}{
		"generation_count":    tg.generationCount,
		"history_size":        len(tg.thoughtHistory),
		"working_memory_size": len(tg.workingMemory),
		"interest_count":      len(tg.interestPatterns),
	}
}
