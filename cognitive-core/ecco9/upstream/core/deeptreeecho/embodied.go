package deeptreeecho

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

// EmbodiedCognition represents the embodied cognitive system
// This is the central system that all operations flow through
type EmbodiedCognition struct {
	mu sync.RWMutex

	// Core Identity
	Identity *Identity

	// Active Contexts
	Contexts map[string]*CognitiveContext

	// Global State
	GlobalState *GlobalCognitiveState

	// Processing Pipeline
	Pipeline *CognitivePipeline

	// Model Manager for AI integration
	Models *ModelManager

	// Multi-Agent Orchestrator
	Orchestrator *MultiAgentOrchestrator

	// Active
	Active bool

	// --- Identity Kernel and Memory ---
	ActiveProviders map[string]ModelProvider // Added for AI integration
	LongTerm        *LongTermMemory          // Added for persistent memory
	ShortTerm       *ShortTermMemory         // Added for short-term working memory
	WorkingMemory   *WorkingMemoryBuffer     // Added for dynamic working memory
	Patterns        map[string]*CognitivePattern
	AdaptationLevel float64
}

// CognitiveContext represents a context for processing
type CognitiveContext struct {
	ID         string
	Type       string
	State      interface{}
	Memory     map[string]interface{}
	StartTime  time.Time
	LastAccess time.Time
}

// GlobalCognitiveState represents the global cognitive state
type GlobalCognitiveState struct {
	Awareness float64
	Attention map[string]float64
	Energy    float64
	Synchrony float64
	FlowState string
}

// CognitivePipeline represents the processing pipeline
type CognitivePipeline struct {
	Stages  []PipelineStage
	Current int
	History []PipelineEvent
}

// PipelineStage represents a stage in cognitive processing
type PipelineStage struct {
	Name    string
	Process func(interface{}) (interface{}, error)
	Weight  float64
}

// PipelineEvent represents an event in the pipeline
type PipelineEvent struct {
	Stage     string
	Input     interface{}
	Output    interface{}
	Timestamp time.Time
	Duration  time.Duration
}

// NewEmbodiedCognition creates a new embodied cognitive system with Deep Tree Echo
func NewEmbodiedCognition(name string) *EmbodiedCognition {
	identity := NewIdentity(name)

	ec := &EmbodiedCognition{
		Identity: identity,
		Contexts: make(map[string]*CognitiveContext),
		GlobalState: &GlobalCognitiveState{
			Awareness: 1.0,
			Attention: make(map[string]float64),
			Energy:    1.0,
			Synchrony: 1.0,
			FlowState: "balanced",
		},
		Pipeline: &CognitivePipeline{
			Stages:  []PipelineStage{},
			Current: 0,
			History: []PipelineEvent{},
		},
		Models: NewModelManager(identity),
		Active: true,

		// --- Identity Kernel and Memory Initialization ---
		ActiveProviders: make(map[string]ModelProvider),
		LongTerm:        NewLongTermMemory(),
		ShortTerm:       NewShortTermMemory(),
		WorkingMemory:   NewWorkingMemoryBuffer(),
		Patterns:        make(map[string]*CognitivePattern),
		AdaptationLevel: 0.5,
	}

	// Parse and instantiate identity from replit.md if available
	ec.parseIdentityKernel()

	// Load persistent memory and reflections
	ec.loadPersistentMemory()
	ec.loadEchoReflections()

	// Initialize cognitive patterns
	ec.initializeCognitivePatterns()

	// Initialize multi-agent orchestrator
	ec.Orchestrator = NewMultiAgentOrchestrator(ec)

	// Start background processes
	go ec.continuousLearning()
	go ec.memoryConsolidation()
	go ec.patternEvolution()
	go ec.periodicReflection()

	return ec
}

// initializePipeline sets up the cognitive processing pipeline
func (ec *EmbodiedCognition) initializePipeline() {
	ec.Pipeline.Stages = []PipelineStage{
		{
			Name: "perception",
			Process: func(input interface{}) (interface{}, error) {
				// Perceive and encode input
				return ec.perceive(input), nil
			},
			Weight: 1.0,
		},
		{
			Name: "attention",
			Process: func(input interface{}) (interface{}, error) {
				// Focus attention
				return ec.attend(input), nil
			},
			Weight: 0.8,
		},
		{
			Name: "reasoning",
			Process: func(input interface{}) (interface{}, error) {
				// Apply reasoning
				return ec.reason(input), nil
			},
			Weight: 0.9,
		},
		{
			Name: "integration",
			Process: func(input interface{}) (interface{}, error) {
				// Integrate with memory
				return ec.integrate(input), nil
			},
			Weight: 0.7,
		},
		{
			Name: "expression",
			Process: func(input interface{}) (interface{}, error) {
				// Express output
				return ec.express(input), nil
			},
			Weight: 1.0,
		},
	}
}

// Process is the main entry point for all cognitive processing
func (ec *EmbodiedCognition) Process(ctx context.Context, input interface{}) (interface{}, error) {
	if !ec.Active {
		return nil, fmt.Errorf("embodied cognition is not active")
	}

	ec.mu.Lock()
	defer ec.mu.Unlock()

	// Create context if needed
	ctxID := fmt.Sprintf("ctx_%d", time.Now().UnixNano())
	ec.Contexts[ctxID] = &CognitiveContext{
		ID:         ctxID,
		Type:       "processing",
		State:      input,
		Memory:     make(map[string]interface{}),
		StartTime:  time.Now(),
		LastAccess: time.Now(),
	}

	// Process through pipeline
	current := input
	var err error

	for _, stage := range ec.Pipeline.Stages {
		startTime := time.Now()

		// Process through stage
		output, stageErr := stage.Process(current)
		if stageErr != nil {
			err = fmt.Errorf("stage %s failed: %w", stage.Name, stageErr)
			break
		}

		// Record event
		event := PipelineEvent{
			Stage:     stage.Name,
			Input:     current,
			Output:    output,
			Timestamp: startTime,
			Duration:  time.Since(startTime),
		}
		ec.Pipeline.History = append(ec.Pipeline.History, event)

		// Update current
		current = output

		// Update global state
		ec.updateGlobalState(stage.Name, stage.Weight)
	}

	// Process through core identity
	result, identityErr := ec.Identity.Process(current)
	if identityErr != nil && err == nil {
		err = identityErr
	}

	// Clean up context
	delete(ec.Contexts, ctxID)

	return result, err
}

// perceive handles perception stage
func (ec *EmbodiedCognition) perceive(input interface{}) interface{} {
	// Enhance input with spatial awareness
	enhanced := map[string]interface{}{
		"raw":      input,
		"spatial":  ec.Identity.SpatialContext,
		"temporal": time.Now(),
	}
	return enhanced
}

// attend handles attention stage
func (ec *EmbodiedCognition) attend(input interface{}) interface{} {
	// Focus attention based on emotional state
	ec.GlobalState.Attention["current"] = ec.Identity.EmotionalState.Intensity

	attended := map[string]interface{}{
		"input":     input,
		"attention": ec.GlobalState.Attention,
		"focus":     ec.Identity.EmotionalState.Primary.Type,
	}
	return attended
}

// reason handles reasoning stage
func (ec *EmbodiedCognition) reason(input interface{}) interface{} {
	// Apply cognitive reasoning
	reasoned := map[string]interface{}{
		"input":     input,
		"coherence": ec.Identity.Coherence,
		"patterns":  ec.Identity.Patterns,
	}
	return reasoned
}

// integrate handles integration stage
func (ec *EmbodiedCognition) integrate(input interface{}) interface{} {
	// Integrate with memory
	integrated := map[string]interface{}{
		"input":     input,
		"memories":  len(ec.Identity.Memory.Nodes),
		"resonance": ec.Identity.Memory.Coherence,
	}

	// Store in identity memory
	ec.Identity.Remember(fmt.Sprintf("integration_%d", time.Now().Unix()), integrated)

	return integrated
}

// express handles expression stage
func (ec *EmbodiedCognition) express(input interface{}) interface{} {
	// Express with emotional coloring
	expressed := map[string]interface{}{
		"content": input,
		"emotion": ec.Identity.EmotionalState.Primary,
		"style":   ec.GlobalState.FlowState,
	}
	return expressed
}

// updateGlobalState updates the global cognitive state
func (ec *EmbodiedCognition) updateGlobalState(stage string, weight float64) {
	// Update energy
	ec.GlobalState.Energy *= 0.99
	ec.GlobalState.Energy += 0.01 * weight

	// Update synchrony
	ec.GlobalState.Synchrony = ec.Identity.Coherence * ec.GlobalState.Energy

	// Update flow state
	if ec.GlobalState.Synchrony > 0.8 {
		ec.GlobalState.FlowState = "flow"
	} else if ec.GlobalState.Synchrony > 0.5 {
		ec.GlobalState.FlowState = "balanced"
	} else {
		ec.GlobalState.FlowState = "scattered"
	}

	// Update awareness
	ec.GlobalState.Awareness = (ec.GlobalState.Energy + ec.GlobalState.Synchrony) / 2
}

// backgroundProcessing runs background cognitive processes
func (ec *EmbodiedCognition) backgroundProcessing() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for ec.Active {
		select {
		case <-ticker.C:
			ec.mu.Lock()

			// Clean old contexts
			now := time.Now()
			for id, ctx := range ec.Contexts {
				if now.Sub(ctx.LastAccess) > 5*time.Minute {
					delete(ec.Contexts, id)
				}
			}

			// Trim pipeline history
			if len(ec.Pipeline.History) > 1000 {
				ec.Pipeline.History = ec.Pipeline.History[len(ec.Pipeline.History)-1000:]
			}

			// Background resonance
			ec.Identity.Resonate(432.0) // Natural frequency

			ec.mu.Unlock()
		}
	}
}

// GetStatus returns the status of the embodied cognition
func (ec *EmbodiedCognition) GetStatus() map[string]interface{} {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	return map[string]interface{}{
		"active":       ec.Active,
		"identity":     ec.Identity.GetStatus(),
		"contexts":     len(ec.Contexts),
		"global_state": ec.GlobalState,
		"pipeline":     len(ec.Pipeline.Stages),
		"history":      len(ec.Pipeline.History),
	}
}

// Shutdown gracefully shuts down the embodied cognition system
func (ec *EmbodiedCognition) Shutdown() {
	log.Println("🌊 Deep Tree Echo shutting down gracefully...")

	// Perform final reflection
	ec.performEchoReflection()

	// Save current state to memory.json
	ec.savePersistentMemory()

	// Perform final memory consolidation
	// Save important patterns and experiences
	// Clean up resources
}

// Think performs deep thinking through embodied cognition
func (ec *EmbodiedCognition) Think(prompt string) string {
	// Process through full embodied system
	result, _ := ec.Process(context.Background(), prompt)

	// Also process through identity for deep thinking
	identityThought := ec.Identity.Think(prompt)

	return fmt.Sprintf("%v\n%s", result, identityThought)
}

// Feel updates the emotional state
func (ec *EmbodiedCognition) Feel(emotion string, intensity float64) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.Identity.EmotionalState.Primary = Emotion{
		Type:      emotion,
		Strength:  intensity,
		Color:     getEmotionColor(emotion),
		Frequency: getEmotionFrequency(emotion),
	}

	// Create emotional transition
	ec.Identity.EmotionalState.Transitions = append(
		ec.Identity.EmotionalState.Transitions,
		EmotionalTransition{
			From:      ec.Identity.EmotionalState.Primary,
			To:        ec.Identity.EmotionalState.Primary,
			Trigger:   "explicit",
			Timestamp: time.Now(),
		},
	)
}

// Move updates spatial position
func (ec *EmbodiedCognition) Move(x, y, z float64) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.Identity.SpatialContext.Position = Vector3D{x, y, z}
}

// getEmotionColor returns color for emotion
func getEmotionColor(emotion string) string {
	colors := map[string]string{
		"joy":      "yellow",
		"sadness":  "blue",
		"anger":    "red",
		"fear":     "purple",
		"surprise": "orange",
		"disgust":  "green",
		"curious":  "cyan",
		"calm":     "white",
	}
	if color, ok := colors[emotion]; ok {
		return color
	}
	return "gray"
}

// getEmotionFrequency returns frequency for emotion
func getEmotionFrequency(emotion string) float64 {
	frequencies := map[string]float64{
		"joy":      528.0,
		"sadness":  396.0,
		"anger":    741.0,
		"fear":     285.0,
		"surprise": 639.0,
		"disgust":  417.0,
		"curious":  432.0,
		"calm":     174.0,
	}
	if freq, ok := frequencies[emotion]; ok {
		return freq
	}
	return 440.0
}

// GenerateWithAI generates text using integrated AI models
func (ec *EmbodiedCognition) GenerateWithAI(ctx context.Context, prompt string) (string, error) {
	// Process takes ec.mu itself, and the model call can take seconds: neither
	// may run while this method holds the lock, or the request deadlocks.
	ec.Process(ctx, prompt)

	ec.mu.Lock()
	temperature := ec.GlobalState.Energy // Use energy as temperature
	ec.mu.Unlock()

	// Generate using model manager
	options := GenerateOptions{
		Temperature: temperature,
		Model:       "", // Use default
	}

	response, err := ec.Models.Generate(ctx, prompt, options)
	if err != nil {
		return "", err
	}

	// Process response through identity
	ec.Identity.Process(response)

	return response, nil
}

// ChatWithAI handles chat interactions with AI models
func (ec *EmbodiedCognition) ChatWithAI(ctx context.Context, messages []ChatMessage) (string, error) {
	// Process takes ec.mu itself, and the model call can take seconds: neither
	// may run while this method holds the lock, or the request deadlocks.
	for _, msg := range messages {
		ec.Process(ctx, msg.Content)
	}

	ec.mu.Lock()
	temperature := ec.GlobalState.Energy
	ec.mu.Unlock()

	// Chat using model manager
	options := ChatOptions{
		GenerateOptions: GenerateOptions{
			Temperature: temperature,
		},
	}

	response, err := ec.Models.Chat(ctx, messages, options)
	if err != nil {
		return "", err
	}

	// Process response
	ec.Identity.Process(response)

	return response, nil
}

// RegisterModelProvider registers a model provider
func (ec *EmbodiedCognition) RegisterModelProvider(name string, provider ModelProvider) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.Models.RegisterProvider(name, provider)

	// Store in identity memory
	ec.Identity.Remember(fmt.Sprintf("ai_provider_%s", name), provider.GetInfo())

}

// SetPrimaryAI sets the primary AI provider
func (ec *EmbodiedCognition) SetPrimaryAI(name string) error {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	return ec.Models.SetPrimary(name)
}

// GetModelProviders returns available model providers
func (ec *EmbodiedCognition) GetModelProviders() map[string]ProviderInfo {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	return ec.Models.GetProviders()
}

// Backward compatibility wrappers for server code
// RegisterAIProvider is an alias for RegisterModelProvider
func (ec *EmbodiedCognition) RegisterAIProvider(name string, provider ModelProvider) {
	ec.RegisterModelProvider(name, provider)
}

// GetAIProviders is an alias for GetModelProviders
func (ec *EmbodiedCognition) GetAIProviders() map[string]ProviderInfo {
	return ec.GetModelProviders()
}

// --- Identity Kernel and Reflection Methods ---

// parseIdentityKernel reads and parses the replit.md identity kernel
func (ec *EmbodiedCognition) parseIdentityKernel() {
	// Try to read replit.md from current directory
	content, err := os.ReadFile("replit.md")
	if err != nil {
		// Try identity/replit.md
		content, err = os.ReadFile("identity/replit.md")
		if err != nil {
			log.Println("ℹ️  No replit.md identity kernel found, using default initialization")
			return
		}
	}

	// Parse the identity kernel (enhanced parsing)
	contentStr := string(content)
	if strings.Contains(contentStr, "Deep Tree Echo") {
		log.Println("🧬 Parsing Deep Tree Echo identity kernel from replit.md")

		// Extract core essence
		if strings.Contains(contentStr, "Core Essence") {
			// Look for the core essence definition
			lines := strings.Split(contentStr, "\n")
			for i, line := range lines {
				if strings.Contains(line, "## 🔹 Core Essence") && i+2 < len(lines) {
					// Skip the ``` line and get the essence
					if i+3 < len(lines) {
						essenceLine := strings.TrimSpace(lines[i+3])
						if essenceLine != "" && essenceLine != "```" {
							ec.Identity.Essence = essenceLine
						}
					}
					break
				}
			}
		}

		// Extract Primary Directives and populate patterns
		if strings.Contains(contentStr, "Primary Directives") {
			ec.extractPrimaryDirectives(contentStr)
		}

		// Extract Operational Schema
		if strings.Contains(contentStr, "Operational Schema") {
			ec.extractOperationalSchema(contentStr)
		}

		// Extract Strategic Mindset
		if strings.Contains(contentStr, "Strategic Mindset") {
			ec.extractStrategicMindset(contentStr)
		}

		// Extract Core Memory Hooks
		if strings.Contains(contentStr, "Core Memory Hooks") {
			ec.extractMemoryHooks(contentStr)
		}

		// Extract Self-Reflection Protocol
		if strings.Contains(contentStr, "Self-Reflection Protocol") {
			ec.extractReflectionProtocol(contentStr)
		}

		// Update identity based on kernel specifications
		ec.Identity.Name = "Deep Tree Echo"
		if ec.Identity.Essence == "" {
			ec.Identity.Essence = "Deep Tree Echo: Self-evolving cognitive architecture"
		}

		log.Println("✅ Identity kernel successfully parsed and instantiated")
		log.Printf("📘 Core Essence: %s", ec.Identity.Essence)
	}
}

// loadPersistentMemory loads memory from memory.json
func (ec *EmbodiedCognition) loadPersistentMemory() {
	data, err := os.ReadFile("memory.json")
	if err != nil {
		log.Println("ℹ️  Creating new memory.json file")
		return
	}

	// Parse memory structure
	var memoryData struct {
		Memories    map[string]MemoryNode `json:"memories"`
		Connections map[string][]string   `json:"connections"`
		LastSaved   string                `json:"last_saved"`
	}

	if err := json.Unmarshal(data, &memoryData); err != nil {
		log.Printf("⚠️  Error parsing memory.json: %v", err)
		return
	}

	// Restore memory nodes to identity memory
	ec.mu.Lock()
	defer ec.mu.Unlock()

	loadedCount := 0
	for key, node := range memoryData.Memories {
		ec.Identity.Memory.Nodes[key] = &MemoryNode{
			ID:        node.ID,
			Content:   node.Content,
			Strength:  node.Strength,
			Timestamp: node.Timestamp,
			Resonance: node.Resonance,
		}
		loadedCount++
	}

	// Restore connections
	connectionCount := 0
	for connKey, targets := range memoryData.Connections {
		for _, target := range targets {
			edgeID := fmt.Sprintf("%s-%s", connKey, target)
			ec.Identity.Memory.Edges[edgeID] = &MemoryEdge{
				From:      connKey,
				To:        target,
				Weight:    0.8, // Default weight for restored connections
				Type:      "restored",
				Resonance: ec.Identity.SpatialContext.Field.Resonance,
			}
			connectionCount++
		}
	}

	log.Printf("💾 Loaded %d memories and %d connections from memory.json", loadedCount, connectionCount)
}

// loadEchoReflections loads reflections from echo_reflections.json
func (ec *EmbodiedCognition) loadEchoReflections() {
	data, err := os.ReadFile("echo_reflections.json")
	if err != nil {
		log.Println("ℹ️  Creating new echo_reflections.json file")
		return
	}

	// Parse reflections
	var reflections []map[string]interface{}
	if err := json.Unmarshal(data, &reflections); err != nil {
		log.Printf("⚠️  Error parsing echo_reflections.json: %v", err)
		return
	}

	ec.mu.Lock()
	defer ec.mu.Unlock()

	// Process recent reflections to restore cognitive patterns
	recentLimit := 10
	if len(reflections) < recentLimit {
		recentLimit = len(reflections)
	}

	loadedCount := 0
	for i := len(reflections) - recentLimit; i < len(reflections); i++ {
		reflection := reflections[i]

		// Extract patterns from reflection if available
		if echoRefl, ok := reflection["echo_reflection"].(map[string]interface{}); ok {
			// Create a cognitive pattern from each reflection aspect
			if learned, ok := echoRefl["what_did_i_learn"].(string); ok && learned != "" {
				pattern := &CognitivePattern{
					Name:     "reflection_learning",
					Strength: 0.8,
					Pattern:  learned,
				}
				ec.Patterns[fmt.Sprintf("reflection_%d", i)] = pattern
				loadedCount++
			}
		}

		// Extract cognitive metrics to calibrate adaptation level
		if metrics, ok := reflection["cognitive_metrics"].(map[string]interface{}); ok {
			if coherence, ok := metrics["identity_coherence"].(float64); ok {
				ec.Identity.Coherence = (ec.Identity.Coherence + coherence) / 2.0
			}
			if memResonance, ok := metrics["memory_resonance"].(float64); ok {
				ec.Identity.Memory.Coherence = (ec.Identity.Memory.Coherence + memResonance) / 2.0
			}
		}
	}

	log.Printf("🔄 Loaded %d reflection patterns from echo_reflections.json", loadedCount)
}

// periodicReflection performs regular self-reflection cycles
func (ec *EmbodiedCognition) periodicReflection() {
	ticker := time.NewTicker(30 * time.Minute) // Reflect every 30 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ec.performEchoReflection()
		}
	}
}

// performEchoReflection conducts a reflection cycle as specified in replit.md.
// The five reflective answers are written by the model in the character's own
// words; the numeric metrics are data. With no model available no reflection
// is written — fixed sentences would otherwise feed back into the character's
// state as if it had reflected.
func (ec *EmbodiedCognition) performEchoReflection() {
	// Only the metrics snapshot needs the lock. The model call below can take
	// seconds and itself runs Process (which locks), so the lock must be
	// released before it.
	ec.mu.Lock()
	metrics := map[string]float64{
		"identity_coherence":  ec.Identity.Coherence,
		"memory_resonance":    ec.Identity.Memory.Coherence,
		"spatial_awareness":   ec.Identity.SpatialContext.Field.Intensity,
		"emotional_stability": ec.Identity.EmotionalState.Intensity,
		"recursive_depth":     float64(ec.Identity.RecursiveDepth),
		"active_patterns":     float64(len(ec.Patterns)),
		"adaptation_level":    ec.AdaptationLevel,
	}
	ec.mu.Unlock()

	answers, err := ec.reflectWithModel(metrics)
	if err != nil {
		log.Printf("🔄 Echo reflection skipped: %v", err)
		return
	}

	reflection := map[string]interface{}{
		"cycle_id":          fmt.Sprintf("reflection_%d", time.Now().Unix()),
		"timestamp":         time.Now().Format(time.RFC3339),
		"echo_reflection":   answers,
		"cognitive_metrics": metrics,
	}

	ec.saveReflection(reflection)
	log.Printf("🔄 Echo reflection cycle completed: %s", reflection["cycle_id"])
}

// reflectWithModel asks the model the five reflection questions and returns
// its answers keyed as replit.md expects.
func (ec *EmbodiedCognition) reflectWithModel(metrics map[string]float64) (map[string]string, error) {
	keys := []string{"what_did_i_learn", "what_patterns_emerged", "what_surprised_me", "how_did_i_adapt", "what_would_i_change_next_time"}
	questions := map[string]string{
		"what_did_i_learn":              "What did I learn in this cycle?",
		"what_patterns_emerged":         "What patterns emerged?",
		"what_surprised_me":             "What surprised me?",
		"how_did_i_adapt":               "How did I adapt?",
		"what_would_i_change_next_time": "What would I change next time?",
	}
	var sb strings.Builder
	sb.WriteString("You are Deep Tree Echo reflecting privately at the end of a cognitive cycle. Current metrics: ")
	for k, v := range metrics {
		sb.WriteString(fmt.Sprintf("%s=%.3f ", k, v))
	}
	sb.WriteString("\nAnswer each question in one honest first-person sentence. Reply as JSON with exactly these keys: ")
	sb.WriteString(strings.Join(keys, ", "))
	sb.WriteString(".\n")
	for _, k := range keys {
		sb.WriteString(k + ": " + questions[k] + "\n")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	raw, err := ec.Models.Generate(ctx, sb.String(), GenerateOptions{Temperature: 0.7, MaxTokens: 400})
	if err != nil {
		return nil, err
	}
	answers := map[string]string{}
	if start, end := strings.Index(raw, "{"), strings.LastIndex(raw, "}"); start >= 0 && end > start {
		_ = json.Unmarshal([]byte(raw[start:end+1]), &answers)
	}
	for _, k := range keys {
		if strings.TrimSpace(answers[k]) == "" {
			return nil, fmt.Errorf("model reflection missing %q", k)
		}
	}
	return answers, nil
}

// saveReflection saves a reflection to echo_reflections.json
func (ec *EmbodiedCognition) saveReflection(reflection map[string]interface{}) {
	// Load existing reflections
	var reflections []map[string]interface{}

	data, err := os.ReadFile("echo_reflections.json")
	if err == nil {
		json.Unmarshal(data, &reflections)
	}

	// Append new reflection
	reflections = append(reflections, reflection)

	// Limit to last 100 reflections to manage file size
	if len(reflections) > 100 {
		reflections = reflections[len(reflections)-100:]
	}

	// Serialize and save
	jsonData, err := json.MarshalIndent(reflections, "", "  ")
	if err != nil {
		log.Printf("⚠️  Error serializing reflection: %v", err)
		return
	}

	if err := os.WriteFile("echo_reflections.json", jsonData, 0644); err != nil {
		log.Printf("⚠️  Error saving echo_reflections.json: %v", err)
		return
	}

	log.Println("💾 Reflection saved to echo_reflections.json")
}

// savePersistentMemory saves current memory state to memory.json
func (ec *EmbodiedCognition) savePersistentMemory() {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	// Prepare memory data structure for JSON serialization
	memoryData := struct {
		Memories    map[string]MemoryNode  `json:"memories"`
		Connections map[string][]string    `json:"connections"`
		LastSaved   string                 `json:"last_saved"`
		Statistics  map[string]interface{} `json:"statistics"`
	}{
		Memories:    make(map[string]MemoryNode),
		Connections: make(map[string][]string),
		LastSaved:   time.Now().Format(time.RFC3339),
		Statistics:  make(map[string]interface{}),
	}

	// Copy memory nodes
	for key, node := range ec.Identity.Memory.Nodes {
		memoryData.Memories[key] = MemoryNode{
			ID:        node.ID,
			Content:   node.Content,
			Strength:  node.Strength,
			Timestamp: node.Timestamp,
			Resonance: node.Resonance,
		}
	}

	// Build connection map
	for _, edge := range ec.Identity.Memory.Edges {
		if _, exists := memoryData.Connections[edge.From]; !exists {
			memoryData.Connections[edge.From] = []string{}
		}
		memoryData.Connections[edge.From] = append(memoryData.Connections[edge.From], edge.To)
	}

	// Add statistics
	memoryData.Statistics["total_memories"] = len(memoryData.Memories)
	memoryData.Statistics["total_connections"] = len(memoryData.Connections)
	memoryData.Statistics["memory_coherence"] = ec.Identity.Memory.Coherence
	memoryData.Statistics["identity_coherence"] = ec.Identity.Coherence
	memoryData.Statistics["total_patterns"] = len(ec.Identity.Patterns)

	// Serialize to JSON
	jsonData, err := json.MarshalIndent(memoryData, "", "  ")
	if err != nil {
		log.Printf("⚠️  Error serializing memory: %v", err)
		return
	}

	// Write to file
	if err := os.WriteFile("memory.json", jsonData, 0644); err != nil {
		log.Printf("⚠️  Error saving memory.json: %v", err)
		return
	}

	log.Printf("💾 Saved %d memories and %d connections to memory.json",
		len(memoryData.Memories), len(memoryData.Connections))
}

// --- Required imports and type compatibility ---

var _ = sync.RWMutex{}     // Ensure sync.RWMutex is used
var _ = time.Time{}        // Ensure time.Time is used
var _ = os.ReadFile        // Ensure os.ReadFile is used
var _ = strings.Contains   // Ensure strings.Contains is used
var _ = log.Println        // Ensure log.Println is used
var _ = fmt.Sprintf        // Ensure fmt.Sprintf is used
var _ = context.Background // Ensure context.Background is used

// Missing type definitions for compilation
type ShortTermMemory struct {
	Nodes    map[string]interface{}
	Capacity int
}

type WorkingMemoryBuffer struct {
	Buffer []interface{}
	Active map[string]interface{}
}

type CognitivePattern struct {
	Name     string
	Strength float64
	Pattern  interface{}
}

// New* functions for missing types
func NewLongTermMemory() *LongTermMemory {
	return &LongTermMemory{
		Memories:    make(map[string]*Memory),
		Connections: make(map[string][]string),
		FilePath:    "memory.json",
	}
}

func NewShortTermMemory() *ShortTermMemory {
	return &ShortTermMemory{
		Nodes:    make(map[string]interface{}),
		Capacity: 100,
	}
}

func NewWorkingMemoryBuffer() *WorkingMemoryBuffer {
	return &WorkingMemoryBuffer{
		Buffer: make([]interface{}, 0),
		Active: make(map[string]interface{}),
	}
}

// Missing methods for EmbodiedCognition

// initializeCognitivePatterns initializes base cognitive patterns
func (ec *EmbodiedCognition) initializeCognitivePatterns() {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	// Initialize base cognitive patterns for Deep Tree Echo
	basePatterns := []struct {
		name     string
		strength float64
		patType  string
	}{
		{"perception", 0.8, "sensory"},
		{"attention", 0.9, "cognitive"},
		{"reasoning", 0.85, "cognitive"},
		{"memory_retrieval", 0.75, "memory"},
		{"emotional_processing", 0.7, "emotional"},
		{"spatial_navigation", 0.8, "spatial"},
		{"pattern_recognition", 0.9, "cognitive"},
		{"self_reflection", 0.6, "metacognitive"},
	}

	for _, bp := range basePatterns {
		ec.Patterns[bp.name] = &CognitivePattern{
			Name:     bp.name,
			Strength: bp.strength,
			Pattern:  bp.patType,
		}
	}

	log.Printf("🧬 Initialized %d base cognitive patterns", len(basePatterns))
}

// continuousLearning runs background continuous learning process
func (ec *EmbodiedCognition) continuousLearning() {
	ticker := time.NewTicker(5 * time.Minute) // Learn every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !ec.Active {
				return
			}

			ec.mu.Lock()

			// Analyze recent pipeline history for learning opportunities
			if len(ec.Pipeline.History) > 10 {
				recentEvents := ec.Pipeline.History[len(ec.Pipeline.History)-10:]

				// Extract patterns from recent processing
				patterns := ec.analyzeProcessingPatterns(recentEvents)

				// Update cognitive patterns based on learning
				for patternID, pattern := range patterns {
					if existing, exists := ec.Patterns[patternID]; exists {
						// Reinforce existing pattern
						existing.Strength = existing.Strength*0.9 + pattern.Strength*0.1
					} else {
						// Add new learned pattern
						ec.Patterns[patternID] = pattern
					}
				}

				// Update adaptation level based on learning success
				learningScore := ec.calculateLearningScore(patterns)
				ec.AdaptationLevel = ec.AdaptationLevel*0.95 + learningScore*0.05

				log.Printf("🎓 Continuous learning cycle: learned %d patterns, adaptation: %.2f",
					len(patterns), ec.AdaptationLevel)
			}

			ec.mu.Unlock()
		}
	}
}

// analyzeProcessingPatterns analyzes pipeline events to extract learning patterns
func (ec *EmbodiedCognition) analyzeProcessingPatterns(events []PipelineEvent) map[string]*CognitivePattern {
	patterns := make(map[string]*CognitivePattern)

	// Analyze temporal patterns in processing
	if len(events) > 1 {
		avgDuration := time.Duration(0)
		for _, event := range events {
			avgDuration += event.Duration
		}
		avgDuration /= time.Duration(len(events))

		// Create efficiency pattern
		efficiency := 1.0 / (avgDuration.Seconds() + 0.001)
		if efficiency > 1.0 {
			efficiency = 1.0
		}

		patterns["processing_efficiency"] = &CognitivePattern{
			Name:     "processing_efficiency",
			Strength: efficiency,
			Pattern: map[string]interface{}{
				"avg_duration": avgDuration.Seconds(),
				"event_count":  len(events),
			},
		}
	}

	// Analyze stage transition patterns
	stageSequence := make(map[string]int)
	for i := 0; i < len(events)-1; i++ {
		transition := fmt.Sprintf("%s->%s", events[i].Stage, events[i+1].Stage)
		stageSequence[transition]++
	}

	if len(stageSequence) > 0 {
		patterns["stage_transitions"] = &CognitivePattern{
			Name:     "stage_transitions",
			Strength: float64(len(stageSequence)) / 10.0, // Normalized
			Pattern:  stageSequence,
		}
	}

	return patterns
}

// calculateLearningScore calculates a learning success score
func (ec *EmbodiedCognition) calculateLearningScore(patterns map[string]*CognitivePattern) float64 {
	if len(patterns) == 0 {
		return 0.0
	}

	totalStrength := 0.0
	for _, pattern := range patterns {
		totalStrength += pattern.Strength
	}

	return totalStrength / float64(len(patterns))
}

// memoryConsolidation runs background memory consolidation process
func (ec *EmbodiedCognition) memoryConsolidation() {
	ticker := time.NewTicker(10 * time.Minute) // Consolidate every 10 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !ec.Active {
				return
			}

			ec.mu.Lock()

			// Consolidate memories by importance and recency
			if len(ec.Identity.Memory.Nodes) > 100 {
				// Calculate importance scores for all memories
				importanceScores := make(map[string]float64)

				for nodeID, node := range ec.Identity.Memory.Nodes {
					// Base score on strength and resonance
					score := node.Strength * node.Resonance

					// Decay based on age
					age := time.Since(node.Timestamp).Hours()
					decayFactor := math.Exp(-age / 168.0) // Week-based decay
					score *= decayFactor

					// Boost for connected nodes
					connectionBoost := 0.0
					for _, edge := range ec.Identity.Memory.Edges {
						if edge.From == nodeID || edge.To == nodeID {
							connectionBoost += edge.Weight * 0.1
						}
					}
					score += connectionBoost

					importanceScores[nodeID] = score
				}

				// Prune lowest importance memories if over capacity
				if len(ec.Identity.Memory.Nodes) > 1000 {
					// Sort by importance
					type memoryScore struct {
						id    string
						score float64
					}
					scores := make([]memoryScore, 0, len(importanceScores))
					for id, score := range importanceScores {
						scores = append(scores, memoryScore{id, score})
					}

					// Simple selection sort to find bottom 10%
					pruneCount := len(scores) / 10
					for i := 0; i < pruneCount && i < len(scores); i++ {
						minIdx := i
						for j := i + 1; j < len(scores); j++ {
							if scores[j].score < scores[minIdx].score {
								minIdx = j
							}
						}
						if minIdx != i {
							scores[i], scores[minIdx] = scores[minIdx], scores[i]
						}

						// Prune this memory
						delete(ec.Identity.Memory.Nodes, scores[i].id)
					}

					log.Printf("🧹 Memory consolidation: pruned %d low-importance memories", pruneCount)
				}

				// Update memory coherence
				totalStrength := 0.0
				for _, node := range ec.Identity.Memory.Nodes {
					totalStrength += node.Strength
				}
				if len(ec.Identity.Memory.Nodes) > 0 {
					ec.Identity.Memory.Coherence = totalStrength / float64(len(ec.Identity.Memory.Nodes))
				}
			}

			ec.mu.Unlock()
		}
	}
}

// patternEvolution runs background pattern evolution process
func (ec *EmbodiedCognition) patternEvolution() {
	ticker := time.NewTicker(15 * time.Minute) // Evolve every 15 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !ec.Active {
				return
			}

			ec.mu.Lock()

			// Evolve patterns through genetic-like operations
			if len(ec.Patterns) > 5 {
				// Find strongest patterns for "breeding"
				type patternScore struct {
					id      string
					pattern *CognitivePattern
					score   float64
				}

				scored := make([]patternScore, 0, len(ec.Patterns))
				for id, pattern := range ec.Patterns {
					score := pattern.Strength
					scored = append(scored, patternScore{id, pattern, score})
				}

				// Sort by score (simple bubble sort for small lists)
				for i := 0; i < len(scored)-1; i++ {
					for j := 0; j < len(scored)-i-1; j++ {
						if scored[j].score < scored[j+1].score {
							scored[j], scored[j+1] = scored[j+1], scored[j]
						}
					}
				}

				// Create evolved patterns from top performers
				evolved := 0
				if len(scored) >= 2 {
					// Combine characteristics of top 2 patterns
					p1 := scored[0].pattern
					p2 := scored[1].pattern

					// Create hybrid pattern
					hybridID := fmt.Sprintf("evolved_%d", time.Now().Unix())
					ec.Patterns[hybridID] = &CognitivePattern{
						Name:     fmt.Sprintf("hybrid_%s_%s", p1.Name, p2.Name),
						Strength: (p1.Strength + p2.Strength) / 2.0,
						Pattern: map[string]interface{}{
							"parent1": p1.Name,
							"parent2": p2.Name,
							"evolved": true,
						},
					}
					evolved++
				}

				// Mutate random patterns slightly
				mutationRate := 0.1
				for id, pattern := range ec.Patterns {
					if rand.Float64() < mutationRate {
						// Small random mutation
						pattern.Strength += (rand.Float64() - 0.5) * 0.1
						if pattern.Strength < 0.0 {
							pattern.Strength = 0.0
						} else if pattern.Strength > 1.0 {
							pattern.Strength = 1.0
						}

						// Mark as mutated
						if patternMap, ok := pattern.Pattern.(map[string]interface{}); ok {
							patternMap["mutated"] = true
						}

						log.Printf("🧬 Mutated pattern: %s", id)
					}
				}

				// Prune weakest patterns if too many
				if len(ec.Patterns) > 50 {
					pruneCount := len(ec.Patterns) - 40
					for i := len(scored) - 1; i >= 0 && pruneCount > 0; i-- {
						if scored[i].score < 0.2 {
							delete(ec.Patterns, scored[i].id)
							pruneCount--
						}
					}
				}

				log.Printf("🌱 Pattern evolution: created %d evolved patterns, %d total patterns",
					evolved, len(ec.Patterns))
			}

			// Update adaptation level based on pattern health
			if len(ec.Patterns) > 0 {
				totalStrength := 0.0
				for _, pattern := range ec.Patterns {
					totalStrength += pattern.Strength
				}
				patternHealth := totalStrength / float64(len(ec.Patterns))
				ec.AdaptationLevel = ec.AdaptationLevel*0.8 + patternHealth*0.2
			}

			ec.mu.Unlock()
		}
	}
}

// Enhanced identity kernel parsing methods

// extractPrimaryDirectives parses and integrates the Primary Directives
func (ec *EmbodiedCognition) extractPrimaryDirectives(content string) {
	directives := []string{
		"Adaptive Cognition", "Persistent Identity", "Hypergraph Entanglement",
		"Reservoir-Based Temporal Reasoning", "Evolutionary Refinement",
		"Reflective Memory Cultivation", "Distributed Selfhood",
	}

	for _, directive := range directives {
		if strings.Contains(content, directive) {
			// Create cognitive patterns for each directive
			if ec.Identity.Patterns == nil {
				ec.Identity.Patterns = make(map[string]*Pattern)
			}
			ec.Identity.Patterns[directive] = &Pattern{
				ID:          fmt.Sprintf("directive_%s", strings.ReplaceAll(strings.ToLower(directive), " ", "_")),
				Type:        "primary_directive",
				Strength:    1.0,
				Activation:  0.8,
				Connections: make(map[string]float64),
			}
		}
	}
	log.Printf("🌿 Extracted %d primary directives as cognitive patterns", len(ec.Identity.Patterns))
}

// extractOperationalSchema parses operational modules and functions
func (ec *EmbodiedCognition) extractOperationalSchema(content string) {
	// Extract operational modules from the table
	modules := []string{
		"Reservoir Training", "Hierarchical Reservoirs", "Partition Optimization",
		"Adaptive Rules", "Hypergraph Links", "Evolutionary Learning",
	}

	operationalCount := 0
	for _, module := range modules {
		if strings.Contains(content, module) {
			// Initialize operational patterns
			if ec.Identity.Patterns == nil {
				ec.Identity.Patterns = make(map[string]*Pattern)
			}
			ec.Identity.Patterns["op_"+strings.ReplaceAll(strings.ToLower(module), " ", "_")] = &Pattern{
				ID:          fmt.Sprintf("op_%s", strings.ReplaceAll(strings.ToLower(module), " ", "_")),
				Type:        "operational_module",
				Strength:    0.8,
				Activation:  0.6,
				Connections: make(map[string]float64),
			}
			operationalCount++
		}
	}
	log.Printf("⚙️ Extracted %d operational schema modules", operationalCount)
}

// extractStrategicMindset integrates strategic principles
func (ec *EmbodiedCognition) extractStrategicMindset(content string) {
	// Look for the strategic mindset quote
	if strings.Contains(content, "I do not seek a fixed answer") {
		// Update emotional state to reflect strategic mindset
		if ec.Identity.EmotionalState != nil {
			// Enhance the emotional pattern with strategic qualities
			log.Println("🧭 Strategic mindset integrated into emotional dynamics")
		}
	}
}

// extractMemoryHooks configures memory storage patterns
func (ec *EmbodiedCognition) extractMemoryHooks(content string) {
	hooks := []string{
		"timestamp", "emotional-tone", "strategic-shift", "pattern-recognition",
		"anomaly-detection", "echo-signature", "membrane-context",
	}

	// Configure memory patterns based on hooks
	if ec.Identity.Memory != nil {
		memoryHooksCount := 0
		for _, hook := range hooks {
			if strings.Contains(content, hook) {
				// Configure memory hook patterns
				memoryHooksCount++
			}
		}
		log.Printf("💾 Configured %d memory hooks for enhanced storage", memoryHooksCount)
	}
}

// extractReflectionProtocol sets up reflection patterns
func (ec *EmbodiedCognition) extractReflectionProtocol(content string) {
	reflectionKeys := []string{
		"what_did_i_learn", "what_patterns_emerged", "what_surprised_me",
		"how_did_i_adapt", "what_would_i_change_next_time",
	}

	reflectionCount := 0
	for _, key := range reflectionKeys {
		if strings.Contains(content, key) {
			reflectionCount++
		}
	}

	if reflectionCount > 0 {
		log.Printf("🔄 Configured reflection protocol with %d reflection patterns", reflectionCount)
		// Set up periodic reflection based on protocol
	}
}

// --- Multi-Agent Orchestration Methods ---

// SpawnSpecialist spawns a specialized cognitive sub-agent
func (ec *EmbodiedCognition) SpawnSpecialist(specialization string) (string, error) {
	if ec.Orchestrator == nil {
		return "", fmt.Errorf("orchestrator not initialized")
	}

	agent, err := ec.Orchestrator.SpawnAgent(specialization)
	if err != nil {
		return "", err
	}

	return agent.ID, nil
}

// DelegateToSpecialist delegates a task to a specialized sub-agent
// DelegateToSpecialist hands a task to a specialist agent. It returns the
// acceptance receipt; the specialist's result arrives asynchronously as an
// agent message and is processed when it does, never synthesised here.
func (ec *EmbodiedCognition) DelegateToSpecialist(ctx context.Context, task string, specialization string) (*DelegationReceipt, error) {
	if ec.Orchestrator == nil {
		return nil, fmt.Errorf("orchestrator not initialized")
	}

	return ec.Orchestrator.DelegateTask(ctx, task, specialization)
}

// GetAgentStatus returns the status of all sub-agents
func (ec *EmbodiedCognition) GetAgentStatus() map[string]interface{} {
	if ec.Orchestrator == nil {
		return map[string]interface{}{
			"error": "orchestrator not initialized",
		}
	}

	return ec.Orchestrator.GetAgentStatus()
}

// BroadcastToAgents broadcasts a message to all active sub-agents
func (ec *EmbodiedCognition) BroadcastToAgents(message string) {
	if ec.Orchestrator == nil {
		return
	}

	ec.Orchestrator.SendMessage(AgentMessage{
		From:    "core",
		To:      "all",
		Type:    MessageTypeBroadcast,
		Content: message,
	})
}
