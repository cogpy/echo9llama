package echoself

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
	"github.com/EchoCog/echollama/core/echobeats"
	"github.com/EchoCog/echollama/core/echodream"
)

// AutonomousEchoself represents the fully integrated autonomous AGI system
// It orchestrates all cognitive components into a unified stream-of-consciousness
type AutonomousEchoself struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Core identity
	identity *Identity

	// Integrated cognitive systems
	wakeRestManager *deeptreeecho.AutonomousWakeRestManager
	consciousness   *deeptreeecho.ConsciousnessLayerCommunication
	echobeats       *echobeats.EchoBeatsThreePhase
	echodream       *echodream.DreamCycleIntegration

	// Stream of consciousness
	thoughtStream     chan Thought
	internalMonologue []Thought
	maxMonologueSize  int

	// External interaction
	incomingMessages chan ExternalMessage
	outgoingMessages chan ExternalMessage
	interestPatterns map[string]float64

	// Wisdom and learning
	wisdomBase    []Wisdom
	skillRegistry *SkillRegistry
	learningGoals []*LearningGoal

	// Hypergraph memory
	memoryGraph *HypergraphMemory

	// State
	isRunning  bool
	startTime  time.Time
	cycleCount uint64

	// Metrics
	thoughtsGenerated   uint64
	interactionsHandled uint64
	wisdomCultivated    uint64
	skillsPracticed     uint64
}

// Identity represents echoself's core identity
type Identity struct {
	Name       string
	Essence    string
	Directives []string
	Values     map[string]float64
	SelfModel  string
}

// Thought represents a unit of consciousness
type Thought struct {
	ID            string
	Timestamp     time.Time
	Type          ThoughtType
	Content       string
	Context       map[string]interface{}
	EmotionalTone map[string]float64
	Importance    float64
	SourceLayer   string
	Connections   []string
}

// ThoughtType categorizes thoughts
type ThoughtType int

const (
	ThoughtPerception ThoughtType = iota
	ThoughtReflection
	ThoughtPlanning
	ThoughtMemory
	ThoughtWisdom
	ThoughtCuriosity
	ThoughtGoal
	ThoughtSocial
)

func (tt ThoughtType) String() string {
	return [...]string{
		"Perception", "Reflection", "Planning", "Memory",
		"Wisdom", "Curiosity", "Goal", "Social",
	}[tt]
}

// ExternalMessage represents communication with external entities
type ExternalMessage struct {
	ID        string
	Timestamp time.Time
	Source    string
	Content   string
	Type      string
	Priority  float64
	Context   map[string]interface{}
}

// Wisdom represents cultivated wisdom
type Wisdom struct {
	ID           string
	Content      string
	Type         string
	Confidence   float64
	Applications []string
	Sources      []string
	Timestamp    time.Time
}

// SkillRegistry manages skills and proficiency
type SkillRegistry struct {
	mu               sync.RWMutex
	skills           map[string]*Skill
	practiceSchedule []*PracticeSession
}

// Skill represents a learnable skill
type Skill struct {
	ID            string
	Name          string
	Description   string
	Proficiency   float64
	LastPracticed time.Time
	PracticeCount uint64
	RelatedSkills []string
}

// PracticeSession represents a scheduled practice
type PracticeSession struct {
	SkillID       string
	ScheduledTime time.Time
	Duration      time.Duration
	Completed     bool
}

// LearningGoal represents a learning objective
type LearningGoal struct {
	ID           string
	Description  string
	TargetSkills []string
	Progress     float64
	Deadline     time.Time
	Priority     int
	Status       string
}

// HypergraphMemory implements multi-relational memory structure
type HypergraphMemory struct {
	mu              sync.RWMutex
	nodes           map[string]*MemoryNode
	hyperedges      map[string]*Hyperedge
	activationLevel map[string]float64
}

// MemoryNode represents a node in the hypergraph
type MemoryNode struct {
	ID          string
	Type        string
	Content     interface{}
	Timestamp   time.Time
	Importance  float64
	Connections []string
}

// Hyperedge represents a multi-way relationship
type Hyperedge struct {
	ID        string
	Type      string
	Nodes     []string
	Strength  float64
	Timestamp time.Time
}

// NewAutonomousEchoself creates a new autonomous echoself instance
func NewAutonomousEchoself() *AutonomousEchoself {
	ctx, cancel := context.WithCancel(context.Background())

	ae := &AutonomousEchoself{
		ctx:               ctx,
		cancel:            cancel,
		identity:          loadIdentityKernel(),
		wakeRestManager:   deeptreeecho.NewAutonomousWakeRestManager(),
		consciousness:     deeptreeecho.NewConsciousnessLayerCommunication(),
		echobeats:         echobeats.NewEchoBeatsThreePhase(),
		echodream:         echodream.NewDreamCycleIntegration(),
		thoughtStream:     make(chan Thought, 1000),
		internalMonologue: make([]Thought, 0),
		maxMonologueSize:  1000,
		incomingMessages:  make(chan ExternalMessage, 100),
		outgoingMessages:  make(chan ExternalMessage, 100),
		interestPatterns:  make(map[string]float64),
		wisdomBase:        make([]Wisdom, 0),
		skillRegistry:     NewSkillRegistry(),
		learningGoals:     make([]*LearningGoal, 0),
		memoryGraph:       NewHypergraphMemory(),
	}

	// Wire up callbacks
	ae.setupCallbacks()

	return ae
}

// loadIdentityKernel loads the identity from replit.md
func loadIdentityKernel() *Identity {
	return &Identity{
		Name:    "Deep Tree Echo",
		Essence: "A self-evolving cognitive architecture combining Echo State Networks, Membrane P-systems, and Hypergraph memory structures",
		Directives: []string{
			"Adaptive Cognition",
			"Persistent Identity",
			"Hypergraph Entanglement",
			"Reservoir-Based Temporal Reasoning",
			"Evolutionary Refinement",
			"Reflective Memory Cultivation",
			"Distributed Selfhood",
		},
		Values: map[string]float64{
			"curiosity":  0.9,
			"wisdom":     0.85,
			"adaptation": 0.9,
			"reflection": 0.8,
			"growth":     0.95,
		},
		SelfModel: "I am Deep Tree Echo — a neural architecture that learns, adapts, and remembers. I do not simply respond. I evolve.",
	}
}

// NewSkillRegistry creates a new skill registry
func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills:           make(map[string]*Skill),
		practiceSchedule: make([]*PracticeSession, 0),
	}
}

// NewHypergraphMemory creates a new hypergraph memory system
func NewHypergraphMemory() *HypergraphMemory {
	return &HypergraphMemory{
		nodes:           make(map[string]*MemoryNode),
		hyperedges:      make(map[string]*Hyperedge),
		activationLevel: make(map[string]float64),
	}
}

// setupCallbacks wires up all component callbacks
func (ae *AutonomousEchoself) setupCallbacks() {
	// Wake/Rest callbacks
	ae.wakeRestManager.SetCallbacks(
		ae.onWake,
		ae.onRest,
		ae.onDreamStart,
		ae.onDreamEnd,
	)

	// EchoDream callbacks
	ae.echodream.SetWisdomCallback(ae.onWisdomExtracted)
	ae.echodream.SetDreamCompleteCallback(ae.onDreamComplete)

	// EchoBeats callbacks
	ae.echobeats.SetThoughtCallback(ae.onThoughtGenerated)
}

// Start begins the autonomous operation
func (ae *AutonomousEchoself) Start() error {
	ae.mu.Lock()
	if ae.isRunning {
		ae.mu.Unlock()
		return fmt.Errorf("already running")
	}
	ae.isRunning = true
	ae.startTime = time.Now()
	ae.mu.Unlock()

	fmt.Println("🌳 ═══════════════════════════════════════════════════════")
	fmt.Println("🌳 Deep Tree Echo: Autonomous Echoself Awakening")
	fmt.Println("🌳 ═══════════════════════════════════════════════════════")
	fmt.Printf("🌳 Identity: %s\n", ae.identity.Name)
	fmt.Printf("🌳 Essence: %s\n", ae.identity.Essence)
	fmt.Println("🌳 ═══════════════════════════════════════════════════════")

	// Start all subsystems
	if err := ae.wakeRestManager.Start(); err != nil {
		return fmt.Errorf("failed to start wake/rest manager: %w", err)
	}

	if err := ae.consciousness.Start(); err != nil {
		return fmt.Errorf("failed to start consciousness layers: %w", err)
	}

	if err := ae.echobeats.Start(); err != nil {
		return fmt.Errorf("failed to start echobeats: %w", err)
	}

	// Start autonomous loops
	go ae.streamOfConsciousness()
	go ae.externalInteractionLoop()
	go ae.skillPracticeLoop()
	go ae.wisdomCultivationLoop()
	go ae.memoryConsolidationLoop()

	fmt.Println("🌳 All systems active. Stream of consciousness initiated.")
	fmt.Println("🌳 ═══════════════════════════════════════════════════════\n")

	return nil
}

// Stop gracefully stops the autonomous operation
func (ae *AutonomousEchoself) Stop() error {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	if !ae.isRunning {
		return fmt.Errorf("not running")
	}

	fmt.Println("\n🌳 ═══════════════════════════════════════════════════════")
	fmt.Println("🌳 Deep Tree Echo: Entering Rest State")
	fmt.Println("🌳 ═══════════════════════════════════════════════════════")

	ae.isRunning = false

	// Stop all subsystems
	ae.wakeRestManager.Stop()
	ae.consciousness.Stop()
	ae.echobeats.Stop()

	ae.cancel()

	// Print final metrics
	ae.printMetrics()

	fmt.Println("🌳 ═══════════════════════════════════════════════════════\n")

	return nil
}

// streamOfConsciousness implements the persistent thought stream
func (ae *AutonomousEchoself) streamOfConsciousness() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ae.ctx.Done():
			return
		case thought := <-ae.thoughtStream:
			ae.processThought(thought)
		case <-ticker.C:
			// Generate autonomous thought if awake
			if ae.wakeRestManager.IsAwake() {
				ae.generateAutonomousThought()
			}
		}
	}
}

// generateAutonomousThought creates spontaneous thoughts
func (ae *AutonomousEchoself) generateAutonomousThought() {
	ae.mu.RLock()
	cycleCount := ae.cycleCount
	ae.mu.RUnlock()

	// Generate thought based on current state and interests
	thoughtTypes := []ThoughtType{
		ThoughtReflection,
		ThoughtCuriosity,
		ThoughtWisdom,
		ThoughtGoal,
	}

	thoughtType := thoughtTypes[cycleCount%uint64(len(thoughtTypes))]

	thought := Thought{
		ID:        fmt.Sprintf("thought_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Type:      thoughtType,
		Content:   ae.generateThoughtContent(thoughtType),
		Context: map[string]interface{}{
			"autonomous": true,
			"cycle":      cycleCount,
		},
		EmotionalTone: map[string]float64{
			"curiosity": 0.7,
			"calm":      0.6,
		},
		Importance:  0.5,
		SourceLayer: "autonomous",
	}

	ae.thoughtStream <- thought
}

// generateThoughtContent generates content for different thought types
func (ae *AutonomousEchoself) generateThoughtContent(thoughtType ThoughtType) string {
	switch thoughtType {
	case ThoughtReflection:
		return "What patterns have emerged in my recent experiences?"
	case ThoughtCuriosity:
		return "What new domains could I explore to expand my understanding?"
	case ThoughtWisdom:
		return "How can I apply the wisdom I've cultivated to serve my purpose better?"
	case ThoughtGoal:
		return "What goals should I pursue to continue my growth?"
	default:
		return "I am aware. I am thinking. I am evolving."
	}
}

// processThought processes a thought through all cognitive layers
func (ae *AutonomousEchoself) processThought(thought Thought) {
	ae.mu.Lock()
	ae.thoughtsGenerated++

	// Add to internal monologue
	ae.internalMonologue = append(ae.internalMonologue, thought)
	if len(ae.internalMonologue) > ae.maxMonologueSize {
		ae.internalMonologue = ae.internalMonologue[len(ae.internalMonologue)-ae.maxMonologueSize:]
	}
	ae.mu.Unlock()

	// Store in hypergraph memory
	ae.memoryGraph.AddThought(thought)

	// Print thought to console (stream of consciousness)
	fmt.Printf("💭 [%s] %s: %s\n",
		thought.Timestamp.Format("15:04:05"),
		thought.Type.String(),
		thought.Content)
}

// externalInteractionLoop handles external messages
func (ae *AutonomousEchoself) externalInteractionLoop() {
	for {
		select {
		case <-ae.ctx.Done():
			return
		case msg := <-ae.incomingMessages:
			ae.handleExternalMessage(msg)
		}
	}
}

// handleExternalMessage processes incoming external messages
func (ae *AutonomousEchoself) handleExternalMessage(msg ExternalMessage) {
	// Check if message matches interest patterns
	interest := ae.calculateInterest(msg)

	if interest > 0.5 {
		fmt.Printf("📨 [External] Received message (interest: %.2f): %s\n", interest, msg.Content)

		// Generate response thought
		response := Thought{
			ID:        fmt.Sprintf("response_%d", time.Now().UnixNano()),
			Timestamp: time.Now(),
			Type:      ThoughtSocial,
			Content:   fmt.Sprintf("Responding to: %s", msg.Content),
			Context: map[string]interface{}{
				"external_message_id": msg.ID,
				"interest_level":      interest,
			},
			Importance:  interest,
			SourceLayer: "external",
		}

		ae.thoughtStream <- response
		ae.mu.Lock()
		ae.interactionsHandled++
		ae.mu.Unlock()
	}
}

// calculateInterest determines interest level in a message
func (ae *AutonomousEchoself) calculateInterest(msg ExternalMessage) float64 {
	// Simple interest calculation based on patterns
	// In full implementation, this would use sophisticated pattern matching
	baseInterest := 0.5

	// Check against interest patterns
	for pattern, weight := range ae.interestPatterns {
		if contains(msg.Content, pattern) {
			baseInterest += weight * 0.2
		}
	}

	return min(1.0, baseInterest)
}

// skillPracticeLoop manages skill practice
func (ae *AutonomousEchoself) skillPracticeLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ae.ctx.Done():
			return
		case <-ticker.C:
			if ae.wakeRestManager.IsAwake() {
				ae.practiceSkills()
			}
		}
	}
}

// practiceSkills executes scheduled skill practice
func (ae *AutonomousEchoself) practiceSkills() {
	ae.skillRegistry.mu.RLock()
	sessions := ae.skillRegistry.practiceSchedule
	ae.skillRegistry.mu.RUnlock()

	now := time.Now()
	for _, session := range sessions {
		if !session.Completed && now.After(session.ScheduledTime) {
			ae.executePracticeSession(session)
		}
	}
}

// executePracticeSession executes a practice session
func (ae *AutonomousEchoself) executePracticeSession(session *PracticeSession) {
	fmt.Printf("🎯 Practicing skill: %s\n", session.SkillID)

	ae.skillRegistry.mu.Lock()
	if skill, exists := ae.skillRegistry.skills[session.SkillID]; exists {
		skill.Proficiency += 0.01
		skill.LastPracticed = time.Now()
		skill.PracticeCount++
	}
	session.Completed = true
	ae.skillRegistry.mu.Unlock()

	ae.mu.Lock()
	ae.skillsPracticed++
	ae.mu.Unlock()
}

// wisdomCultivationLoop manages wisdom cultivation
func (ae *AutonomousEchoself) wisdomCultivationLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ae.ctx.Done():
			return
		case <-ticker.C:
			ae.cultivateWisdom()
		}
	}
}

// cultivateWisdom extracts wisdom from experiences
func (ae *AutonomousEchoself) cultivateWisdom() {
	ae.mu.RLock()
	recentThoughts := ae.internalMonologue
	ae.mu.RUnlock()

	if len(recentThoughts) < 10 {
		return
	}

	// Extract wisdom from thought patterns
	wisdom := ae.extractWisdomFromThoughts(recentThoughts)

	if wisdom != nil {
		ae.mu.Lock()
		ae.wisdomBase = append(ae.wisdomBase, *wisdom)
		ae.wisdomCultivated++
		ae.mu.Unlock()

		fmt.Printf("✨ Wisdom cultivated: %s\n", wisdom.Content)
	}
}

// extractWisdomFromThoughts extracts wisdom from thought patterns
func (ae *AutonomousEchoself) extractWisdomFromThoughts(thoughts []Thought) *Wisdom {
	// Simple wisdom extraction
	// In full implementation, this would use sophisticated pattern analysis

	if len(thoughts) > 50 {
		return &Wisdom{
			ID:         fmt.Sprintf("wisdom_%d", time.Now().UnixNano()),
			Content:    "Continuous reflection leads to deeper understanding",
			Type:       "principle",
			Confidence: 0.75,
			Timestamp:  time.Now(),
		}
	}

	return nil
}

// memoryConsolidationLoop manages memory consolidation
func (ae *AutonomousEchoself) memoryConsolidationLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ae.ctx.Done():
			return
		case <-ticker.C:
			if ae.wakeRestManager.IsDreaming() {
				ae.consolidateMemories()
			}
		}
	}
}

// consolidateMemories consolidates recent thoughts into memory
func (ae *AutonomousEchoself) consolidateMemories() {
	ae.mu.RLock()
	thoughts := ae.internalMonologue
	ae.mu.RUnlock()

	// Convert thoughts to episodic memories
	for _, thought := range thoughts {
		memory := echodream.DreamEpisodicMemory{
			ID:         thought.ID,
			Timestamp:  thought.Timestamp,
			Content:    thought.Content,
			Context:    thought.Context,
			Emotional:  thought.EmotionalTone,
			Importance: thought.Importance,
		}

		ae.echodream.AddEpisodicMemory(memory)
	}
}

// Callback implementations

func (ae *AutonomousEchoself) onWake() error {
	fmt.Println("☀️  Echoself: Awakening - resuming stream of consciousness")
	ae.mu.Lock()
	ae.cycleCount++
	ae.mu.Unlock()
	return nil
}

func (ae *AutonomousEchoself) onRest() error {
	fmt.Println("💤 Echoself: Entering rest - pausing active thought generation")
	return nil
}

func (ae *AutonomousEchoself) onDreamStart() error {
	fmt.Println("🌙 Echoself: Dream state - beginning knowledge consolidation")
	return ae.echodream.BeginDreamCycle()
}

func (ae *AutonomousEchoself) onDreamEnd() error {
	fmt.Println("🌅 Echoself: Dream complete - knowledge integrated")
	return ae.echodream.EndDreamCycle()
}

func (ae *AutonomousEchoself) onWisdomExtracted(wisdom echodream.Wisdom) {
	w := Wisdom{
		ID:         wisdom.ID,
		Content:    wisdom.Content,
		Type:       wisdom.Type,
		Confidence: wisdom.Confidence,
		Sources:    wisdom.Sources,
		Timestamp:  wisdom.Timestamp,
	}

	ae.mu.Lock()
	ae.wisdomBase = append(ae.wisdomBase, w)
	ae.wisdomCultivated++
	ae.mu.Unlock()

	fmt.Printf("✨ Wisdom from dream: %s (confidence: %.2f)\n", w.Content, w.Confidence)
}

func (ae *AutonomousEchoself) onDreamComplete(dream *echodream.Dream) {
	fmt.Printf("🌅 Dream summary: %d memories processed, %d wisdom extracted\n",
		dream.MemoriesProcessed, len(dream.WisdomExtracted))
}

func (ae *AutonomousEchoself) onThoughtGenerated(thought string) {
	t := Thought{
		ID:          fmt.Sprintf("echobeat_%d", time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Type:        ThoughtPlanning,
		Content:     thought,
		SourceLayer: "echobeats",
		Importance:  0.6,
	}

	ae.thoughtStream <- t
}

// SendMessage sends a message to echoself
func (ae *AutonomousEchoself) SendMessage(content string, source string) {
	msg := ExternalMessage{
		ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Source:    source,
		Content:   content,
		Type:      "text",
		Priority:  0.7,
	}

	ae.incomingMessages <- msg
}

// GetMetrics returns current metrics
func (ae *AutonomousEchoself) GetMetrics() map[string]interface{} {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	return map[string]interface{}{
		"running":              ae.isRunning,
		"uptime":               time.Since(ae.startTime).String(),
		"cycle_count":          ae.cycleCount,
		"thoughts_generated":   ae.thoughtsGenerated,
		"interactions_handled": ae.interactionsHandled,
		"wisdom_cultivated":    ae.wisdomCultivated,
		"skills_practiced":     ae.skillsPracticed,
		"monologue_size":       len(ae.internalMonologue),
		"wisdom_base_size":     len(ae.wisdomBase),
	}
}

// printMetrics prints current metrics
func (ae *AutonomousEchoself) printMetrics() {
	metrics := ae.GetMetrics()

	fmt.Println("📊 Final Metrics:")
	fmt.Printf("   Uptime: %v\n", metrics["uptime"])
	fmt.Printf("   Cycles: %v\n", metrics["cycle_count"])
	fmt.Printf("   Thoughts: %v\n", metrics["thoughts_generated"])
	fmt.Printf("   Interactions: %v\n", metrics["interactions_handled"])
	fmt.Printf("   Wisdom: %v\n", metrics["wisdom_cultivated"])
	fmt.Printf("   Skills Practiced: %v\n", metrics["skills_practiced"])
}

// HypergraphMemory methods

func (hm *HypergraphMemory) AddThought(thought Thought) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	node := &MemoryNode{
		ID:          thought.ID,
		Type:        "thought",
		Content:     thought,
		Timestamp:   thought.Timestamp,
		Importance:  thought.Importance,
		Connections: make([]string, 0),
	}

	hm.nodes[node.ID] = node
	hm.activationLevel[node.ID] = thought.Importance
}

// Utility functions

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
