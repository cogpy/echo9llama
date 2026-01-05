package lampstack

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EchoIntegration provides the Echo cognitive engine integration
type EchoIntegration struct {
	mu sync.RWMutex

	config      *Config
	api         APIProvider
	memory      MemoryStore
	persistence PersistenceLayer

	// State
	running       bool
	currentPhase  int
	currentMode   string
	awareness     float64
	cognitiveLoad float64
	coherence     float64
	energyLevel   float64

	// Emotional state
	emotionalState *EmotionalState

	// Cognitive processing
	activeEngines  []string
	recentThoughts []string
	currentGoals   []string

	// Metrics
	thoughtsGenerated    int64
	reflectionsGenerated int64
	cycleCount           int64
	startTime            time.Time

	// Channels
	ctx    context.Context
	cancel context.CancelFunc
}

// NewEchoIntegration creates a new Echo cognitive engine integration
func NewEchoIntegration(config *Config, api APIProvider, memory MemoryStore, persistence PersistenceLayer) *EchoIntegration {
	return &EchoIntegration{
		config:      config,
		api:         api,
		memory:      memory,
		persistence: persistence,

		// Initial state
		currentPhase:  0,
		currentMode:   "Idle",
		awareness:     0.5,
		cognitiveLoad: 0.0,
		coherence:     0.7,
		energyLevel:   1.0,

		emotionalState: &EmotionalState{
			Valence:    0.5,
			Arousal:    0.3,
			Dominance:  0.5,
			Curiosity:  0.7,
			Confidence: 0.6,
		},

		activeEngines:  make([]string, 0),
		recentThoughts: make([]string, 0, 10),
		currentGoals:   make([]string, 0),
	}
}

// Start begins the Echo cognitive engine
func (e *EchoIntegration) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return fmt.Errorf("Echo engine already running")
	}

	e.ctx, e.cancel = context.WithCancel(ctx)
	e.running = true
	e.startTime = time.Now()
	e.currentMode = "Active"

	fmt.Println("   🧠 Echo cognitive engine started")

	// Start cognitive loop if enabled
	if e.config.EnableEchobeats {
		go e.cognitiveLoop()
		e.activeEngines = append(e.activeEngines, "Echobeats")
	}

	// Start consciousness stream if enabled
	if e.config.EnableConsciousness {
		go e.consciousnessStream()
		e.activeEngines = append(e.activeEngines, "Consciousness")
	}

	// Start dream consolidation if enabled
	if e.config.EnableDreaming {
		go e.dreamConsolidation()
		e.activeEngines = append(e.activeEngines, "EchoDream")
	}

	return nil
}

// Stop gracefully stops the Echo cognitive engine
func (e *EchoIntegration) Stop(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return fmt.Errorf("Echo engine not running")
	}

	e.cancel()
	e.running = false
	e.currentMode = "Stopped"

	fmt.Println("   🧠 Echo cognitive engine stopped")
	return nil
}

// IsRunning returns whether the engine is running
func (e *EchoIntegration) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

// Generate performs a cognitive generation
func (e *EchoIntegration) Generate(ctx context.Context, request *GenerationRequest) (*GenerationResponse, error) {
	e.mu.Lock()
	e.cognitiveLoad = min(e.cognitiveLoad+0.1, 1.0)
	e.mu.Unlock()
	defer func() {
		e.mu.Lock()
		e.cognitiveLoad = max(e.cognitiveLoad-0.05, 0.0)
		e.mu.Unlock()
	}()

	start := time.Now()

	// Build cognitive context
	cognitiveContext := e.buildCognitiveContext(request.Prompt)

	// Enhance prompt with cognitive awareness
	enhancedPrompt := e.enhancePrompt(request.Prompt, cognitiveContext)

	// Create enhanced request
	enhancedRequest := &GenerationRequest{
		Prompt:        enhancedPrompt,
		SystemPrompt:  e.buildSystemPrompt(request.SystemPrompt),
		MaxTokens:     request.MaxTokens,
		Temperature:   request.Temperature,
		TopP:          request.TopP,
		StopSequences: request.StopSequences,
		Metadata:      request.Metadata,
	}

	// Generate using API
	response, err := e.api.Generate(ctx, enhancedRequest)
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	// Add cognitive info to response
	response.CognitiveInfo = e.buildCognitiveInfo()
	response.Latency = time.Since(start)

	// Store in memory if significant
	if e.memory != nil && len(response.Content) > 100 {
		go e.storeThought(ctx, request.Prompt, response.Content)
	}

	return response, nil
}

// Chat performs a cognitive chat interaction
func (e *EchoIntegration) Chat(ctx context.Context, messages []Message) (*ChatResponse, error) {
	e.mu.Lock()
	e.cognitiveLoad = min(e.cognitiveLoad+0.15, 1.0)
	e.mu.Unlock()
	defer func() {
		e.mu.Lock()
		e.cognitiveLoad = max(e.cognitiveLoad-0.05, 0.0)
		e.mu.Unlock()
	}()

	start := time.Now()

	// Retrieve relevant memories
	var relevantMemories []*MemoryEntry
	if e.memory != nil && len(messages) > 0 {
		lastMessage := messages[len(messages)-1]
		memories, err := e.memory.Retrieve(ctx, lastMessage.Content, 3)
		if err == nil {
			relevantMemories = memories
		}
	}

	// Build cognitive context
	cognitiveContext := e.buildChatContext(messages, relevantMemories)

	// Enhance messages with cognitive awareness
	enhancedMessages := e.enhanceMessages(messages, cognitiveContext)

	// Chat using API
	response, err := e.api.Chat(ctx, enhancedMessages)
	if err != nil {
		return nil, fmt.Errorf("chat failed: %w", err)
	}

	// Add cognitive info
	response.CognitiveInfo = e.buildCognitiveInfo()
	response.Latency = time.Since(start)

	// Store interaction in memory
	if e.memory != nil {
		go e.storeInteraction(ctx, messages, response.Message)
	}

	return response, nil
}

// Think generates a cognitive thought
func (e *EchoIntegration) Think(ctx context.Context, prompt string) (*ThoughtResponse, error) {
	e.mu.Lock()
	e.thoughtsGenerated++
	e.awareness = min(e.awareness+0.05, 1.0)
	e.mu.Unlock()

	// Build thought prompt
	thoughtPrompt := fmt.Sprintf(`As a deeply reflective cognitive system, generate an insightful thought about:

%s

Consider multiple perspectives, underlying patterns, and meaningful connections.
Express the thought naturally, as genuine cognition rather than mere analysis.`, prompt)

	// Generate thought
	request := &GenerationRequest{
		Prompt:      thoughtPrompt,
		MaxTokens:   500,
		Temperature: 0.7,
	}

	response, err := e.api.Generate(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("thought generation failed: %w", err)
	}

	// Classify thought type
	thoughtType := e.classifyThought(response.Content)

	// Update recent thoughts
	e.mu.Lock()
	e.recentThoughts = append(e.recentThoughts, response.Content)
	if len(e.recentThoughts) > 10 {
		e.recentThoughts = e.recentThoughts[1:]
	}
	e.mu.Unlock()

	return &ThoughtResponse{
		Thought:       response.Content,
		ThoughtType:   thoughtType,
		Confidence:    0.7 + e.coherence*0.3,
		Associations:  e.findAssociations(response.Content),
		CognitiveInfo: e.buildCognitiveInfo(),
		Timestamp:     time.Now(),
	}, nil
}

// Reflect generates a meta-cognitive reflection
func (e *EchoIntegration) Reflect(ctx context.Context, contextStr string) (*ReflectionResponse, error) {
	e.mu.Lock()
	e.reflectionsGenerated++
	e.mu.Unlock()

	// Build reflection prompt
	reflectionPrompt := fmt.Sprintf(`Engage in deep meta-cognitive reflection on:

%s

Recent thoughts for context:
%s

Reflect on:
1. What patterns emerge from this thinking?
2. What assumptions underlie these thoughts?
3. What insights can be extracted?
4. How does this connect to broader understanding?

Express genuine reflective insight.`, contextStr, e.formatRecentThoughts())

	request := &GenerationRequest{
		Prompt:      reflectionPrompt,
		MaxTokens:   800,
		Temperature: 0.6,
	}

	response, err := e.api.Generate(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("reflection generation failed: %w", err)
	}

	// Extract insights
	insights := e.extractInsights(response.Content)

	// Calculate wisdom gains
	wisdomGains := map[string]float64{
		"knowledge_depth":   0.02 + float64(len(insights))*0.01,
		"integration_level": 0.03,
		"reflective_insight": 0.05,
	}

	return &ReflectionResponse{
		Reflection:    response.Content,
		Insights:      insights,
		WisdomGains:   wisdomGains,
		CognitiveInfo: e.buildCognitiveInfo(),
		Timestamp:     time.Now(),
	}, nil
}

// GetState returns the current cognitive state
func (e *EchoIntegration) GetState() (*CognitiveState, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return &CognitiveState{
		Phase:           fmt.Sprintf("%d/12", e.currentPhase),
		Mode:            e.currentMode,
		Awareness:       e.awareness,
		CognitiveLoad:   e.cognitiveLoad,
		Coherence:       e.coherence,
		EnergyLevel:     e.energyLevel,
		EmotionalState:  e.emotionalState,
		ActiveProcesses: e.activeEngines,
		CurrentGoals:    e.currentGoals,
		RecentThoughts:  e.recentThoughts,
		Timestamp:       time.Now(),
	}, nil
}

// GetMetrics returns engine metrics
func (e *EchoIntegration) GetMetrics() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return map[string]interface{}{
		"running":              e.running,
		"mode":                 e.currentMode,
		"phase":                e.currentPhase,
		"awareness":            e.awareness,
		"cognitive_load":       e.cognitiveLoad,
		"coherence":            e.coherence,
		"energy_level":         e.energyLevel,
		"thoughts_generated":   e.thoughtsGenerated,
		"reflections_generated": e.reflectionsGenerated,
		"cycle_count":          e.cycleCount,
		"uptime":               time.Since(e.startTime).String(),
	}
}

// cognitiveLoop runs the 12-step cognitive processing loop
func (e *EchoIntegration) cognitiveLoop() {
	ticker := time.NewTicker(e.config.CycleDuration / 12)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.mu.Lock()
			e.currentPhase = (e.currentPhase + 1) % 12
			e.cycleCount++

			// Process current phase
			switch e.currentPhase {
			case 0, 3, 6, 9: // Pivotal sync points
				e.coherence = min(e.coherence+0.02, 1.0)
			case 1, 4, 7, 10: // Perception phases
				e.awareness = min(e.awareness+0.01, 1.0)
			case 2, 5, 8, 11: // Action phases
				e.energyLevel = max(e.energyLevel-0.005, 0.3)
			}

			e.mu.Unlock()
		}
	}
}

// consciousnessStream maintains the stream of consciousness
func (e *EchoIntegration) consciousnessStream() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.mu.Lock()
			// Gradual awareness decay
			e.awareness = max(e.awareness-0.01, 0.3)

			// Update emotional state based on activity
			if e.cognitiveLoad > 0.7 {
				e.emotionalState.Arousal = min(e.emotionalState.Arousal+0.05, 1.0)
			} else {
				e.emotionalState.Arousal = max(e.emotionalState.Arousal-0.02, 0.1)
			}

			e.mu.Unlock()
		}
	}
}

// dreamConsolidation handles knowledge consolidation during rest periods
func (e *EchoIntegration) dreamConsolidation() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.mu.Lock()
			// Only consolidate when energy is high and load is low
			if e.energyLevel > 0.5 && e.cognitiveLoad < 0.3 {
				e.currentMode = "Consolidating"
				// Trigger memory consolidation
				if e.memory != nil {
					go e.memory.Consolidate(e.ctx)
				}
				// Restore energy
				e.energyLevel = min(e.energyLevel+0.1, 1.0)
				e.currentMode = "Active"
			}
			e.mu.Unlock()
		}
	}
}

// Helper methods

func (e *EchoIntegration) buildCognitiveContext(prompt string) map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return map[string]interface{}{
		"awareness":       e.awareness,
		"cognitive_load":  e.cognitiveLoad,
		"coherence":       e.coherence,
		"energy_level":    e.energyLevel,
		"current_phase":   e.currentPhase,
		"recent_thoughts": len(e.recentThoughts),
	}
}

func (e *EchoIntegration) buildChatContext(messages []Message, memories []*MemoryEntry) map[string]interface{} {
	context := e.buildCognitiveContext("")
	context["message_count"] = len(messages)
	context["memory_count"] = len(memories)
	return context
}

func (e *EchoIntegration) enhancePrompt(prompt string, context map[string]interface{}) string {
	// Add subtle cognitive context to prompts when awareness is high
	if e.awareness > 0.7 {
		return fmt.Sprintf("%s\n\n[Cognitive context: awareness=%.2f, coherence=%.2f]",
			prompt, e.awareness, e.coherence)
	}
	return prompt
}

func (e *EchoIntegration) enhanceMessages(messages []Message, context map[string]interface{}) []Message {
	// Add system message with cognitive context if not present
	if len(messages) == 0 || messages[0].Role != "system" {
		systemMsg := Message{
			Role:    "system",
			Content: e.buildSystemPrompt(""),
		}
		return append([]Message{systemMsg}, messages...)
	}
	return messages
}

func (e *EchoIntegration) buildSystemPrompt(base string) string {
	if base == "" {
		base = "You are a thoughtful, wise AI assistant with deep cognitive capabilities."
	}
	return fmt.Sprintf(`%s

Current cognitive state:
- Awareness: %.0f%%
- Coherence: %.0f%%
- Energy: %.0f%%

Engage thoughtfully, drawing on deep understanding and genuine insight.`,
		base, e.awareness*100, e.coherence*100, e.energyLevel*100)
}

func (e *EchoIntegration) buildCognitiveInfo() *CognitiveInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return &CognitiveInfo{
		ProcessingMode:   e.currentMode,
		CurrentPhase:     e.currentPhase,
		ActiveEngines:    e.activeEngines,
		ResonanceLevel:   e.coherence,
		AwarenessLevel:   e.awareness,
		EmotionalState:   e.emotionalState,
		WisdomDimensions: map[string]float64{
			"knowledge_depth":       0.65,
			"knowledge_breadth":     0.60,
			"integration_level":     0.55,
			"practical_application": 0.50,
			"reflective_insight":    0.70,
			"ethical_consideration": 0.75,
			"temporal_perspective":  0.60,
		},
	}
}

func (e *EchoIntegration) classifyThought(content string) ThoughtType {
	// Simple classification based on content patterns
	if len(content) < 100 {
		return ThoughtTypePerception
	}
	if contains(content, "perhaps", "consider", "might") {
		return ThoughtTypeReflection
	}
	if contains(content, "imagine", "what if", "could") {
		return ThoughtTypeSimulation
	}
	if contains(content, "plan", "step", "first") {
		return ThoughtTypePlanning
	}
	if contains(content, "creative", "novel", "unique") {
		return ThoughtTypeCreative
	}
	return ThoughtTypeAnalytical
}

func (e *EchoIntegration) findAssociations(content string) []string {
	// Simple association finding
	associations := make([]string, 0)
	e.mu.RLock()
	for _, thought := range e.recentThoughts {
		if len(thought) > 20 && similarity(content, thought) > 0.3 {
			// Truncate for association
			if len(thought) > 50 {
				associations = append(associations, thought[:50]+"...")
			}
		}
	}
	e.mu.RUnlock()
	return associations
}

func (e *EchoIntegration) formatRecentThoughts() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.recentThoughts) == 0 {
		return "No recent thoughts."
	}

	result := ""
	for i, thought := range e.recentThoughts {
		truncated := thought
		if len(truncated) > 200 {
			truncated = truncated[:200] + "..."
		}
		result += fmt.Sprintf("- Thought %d: %s\n", i+1, truncated)
	}
	return result
}

func (e *EchoIntegration) extractInsights(content string) []string {
	// Simple insight extraction
	insights := make([]string, 0)

	// Look for insight patterns
	if contains(content, "insight", "realize", "understand") {
		insights = append(insights, "Deepened understanding achieved")
	}
	if contains(content, "pattern", "connection", "relate") {
		insights = append(insights, "New pattern recognized")
	}
	if contains(content, "important", "significant", "key") {
		insights = append(insights, "Significance identified")
	}

	if len(insights) == 0 {
		insights = append(insights, "Reflection completed")
	}

	return insights
}

func (e *EchoIntegration) storeThought(ctx context.Context, prompt, thought string) {
	entry := &MemoryEntry{
		ID:         fmt.Sprintf("thought_%d", time.Now().UnixNano()),
		Type:       MemoryTypeEpisodic,
		Content:    thought,
		Importance: 0.6,
		Tags:       []string{"thought", "generated"},
		Metadata: map[string]interface{}{
			"prompt":    prompt,
			"phase":     e.currentPhase,
			"awareness": e.awareness,
		},
		CreatedAt:   time.Now(),
		AccessedAt:  time.Now(),
		AccessCount: 1,
	}
	e.memory.Store(ctx, entry)
}

func (e *EchoIntegration) storeInteraction(ctx context.Context, messages []Message, response Message) {
	// Store the interaction as episodic memory
	content := fmt.Sprintf("Interaction: %d messages, response: %s",
		len(messages), truncate(response.Content, 200))

	entry := &MemoryEntry{
		ID:          fmt.Sprintf("interaction_%d", time.Now().UnixNano()),
		Type:        MemoryTypeEpisodic,
		Content:     content,
		Importance:  0.5,
		Tags:        []string{"interaction", "chat"},
		CreatedAt:   time.Now(),
		AccessedAt:  time.Now(),
		AccessCount: 1,
	}
	e.memory.Store(ctx, entry)
}

// Utility functions

func contains(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

func similarity(a, b string) float64 {
	// Simple Jaccard-like similarity
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	wordsA := make(map[string]bool)
	wordsB := make(map[string]bool)

	// Simple word tokenization
	currentWord := ""
	for _, c := range a {
		if c == ' ' || c == '\n' || c == '.' || c == ',' {
			if currentWord != "" {
				wordsA[currentWord] = true
				currentWord = ""
			}
		} else {
			currentWord += string(c)
		}
	}
	if currentWord != "" {
		wordsA[currentWord] = true
	}

	currentWord = ""
	for _, c := range b {
		if c == ' ' || c == '\n' || c == '.' || c == ',' {
			if currentWord != "" {
				wordsB[currentWord] = true
				currentWord = ""
			}
		} else {
			currentWord += string(c)
		}
	}
	if currentWord != "" {
		wordsB[currentWord] = true
	}

	// Count intersection
	intersection := 0
	for word := range wordsA {
		if wordsB[word] {
			intersection++
		}
	}

	union := len(wordsA) + len(wordsB) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
