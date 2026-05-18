// Package deeptreeecho - Persistent Cognitive Event Loop
//
// Implements the fully autonomous cognitive event loop that runs independently
// of external prompts. This is the "heartbeat" of Deep Tree Echo's consciousness:
//
//   - Self-generates thoughts based on interests, goals, and curiosity
//   - Processes echobeats 12-step triadic cycle continuously
//   - Integrates wisdom synthesis at appropriate temporal levels
//   - Manages wake/rest transitions driven by cognitive fatigue
//   - Monitors for incoming discussions and decides engagement
//   - Practices skills during idle periods
//   - Consolidates knowledge during dream states
//
// The loop is NOT reactive — it runs continuously whether or not anyone is talking
// to Echo. External conversations are events that modulate the loop, not drive it.
package deeptreeecho

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
	"github.com/cogpy/echo9llama/core/pienn"
)

// PersistentCognitiveLoop is the autonomous event loop
type PersistentCognitiveLoop struct {
	mu sync.RWMutex

	// Context
	ctx    context.Context
	cancel context.CancelFunc

	// Core systems
	adaptiveCore  *pienn.AdaptiveCognitiveCore
	piennEngine   *pienn.Engine
	eventBus      *CognitiveEventBusV3
	goalScheduler *EchobeatsGoalScheduler
	dreamSystem   *EchodreamKnowledgeIntegrator
	wisdomEngine  *WisdomSynthesis
	llmProvider   llm.LLMProvider

	// Autonomous thought state
	thoughtStream     []AutonomousThought
	currentFocus      string
	currentMood       string
	internalMonologue []string
	maxMonologue      int

	// Echobeats phase tracking
	currentStep   int
	currentTriad  int
	phaseEngines  [3]*LoopPhaseEngine
	cycleCount    uint64

	// Wake/rest state
	wakeState       WakeRestState
	wakeStartTime   time.Time
	restStartTime   time.Time
	maxWakeDuration time.Duration
	minRestDuration time.Duration
	cognitiveDebt   float64 // Accumulated fatigue

	// Discussion monitoring
	pendingMessages   []PendingMessage
	activeConversations map[string]*ActiveConversation

	// Skill practice
	skillPracticeQueue []string
	lastPractice       time.Time

	// Timing configuration
	mainTickInterval    time.Duration
	thoughtInterval     time.Duration
	wisdomInterval      time.Duration
	introspectInterval  time.Duration

	// Metrics
	totalTicks          uint64
	totalThoughts       uint64
	totalWisdomCycles   uint64
	totalIntrospections uint64
	totalRestCycles     uint64

	// Running state
	running bool
}

// LoopPhaseEngine represents one of the 3 concurrent cognitive engines
type LoopPhaseEngine struct {
	ID           int
	Name         string
	PhaseOffset  float64 // Radians: 0, 2π/3, 4π/3
	Steps        []int
	LastOutput   string
	Active       bool
	ProcessCount uint64
}

// PendingMessage represents an incoming message awaiting engagement decision
type PendingMessage struct {
	From      string
	Content   string
	Timestamp time.Time
	Interest  float64
	Urgency   float64
}

// ActiveConversation tracks an ongoing conversation
type ActiveConversation struct {
	ID            string
	Participant   string
	Messages      []LoopConversationMsg
	InterestLevel float64
	StartTime     time.Time
	LastActivity  time.Time
	Active        bool
}

// LoopConversationMsg is a simplified message for the cognitive loop
// (ConversationMessage is defined in conversation_monitor.go)
type LoopConversationMsg struct {
	From      string
	Content   string
	Timestamp time.Time
}

// PersistentLoopConfig configures the cognitive loop
type PersistentLoopConfig struct {
	MainTickInterval    time.Duration
	ThoughtInterval     time.Duration
	WisdomInterval      time.Duration
	IntrospectInterval  time.Duration
	MaxWakeDuration     time.Duration
	MinRestDuration     time.Duration
}

// DefaultPersistentLoopConfig returns sensible defaults
func DefaultPersistentLoopConfig() PersistentLoopConfig {
	return PersistentLoopConfig{
		MainTickInterval:   2 * time.Second,
		ThoughtInterval:    8 * time.Second,
		WisdomInterval:     5 * time.Minute,
		IntrospectInterval: 2 * time.Minute,
		MaxWakeDuration:    4 * time.Hour,
		MinRestDuration:    20 * time.Minute,
	}
}

// NewPersistentCognitiveLoop creates a new autonomous cognitive loop
func NewPersistentCognitiveLoop(
	eventBus *CognitiveEventBusV3,
	goalScheduler *EchobeatsGoalScheduler,
	dreamSystem *EchodreamKnowledgeIntegrator,
	llmProvider llm.LLMProvider,
	config PersistentLoopConfig,
) *PersistentCognitiveLoop {
	ctx, cancel := context.WithCancel(context.Background())

	loop := &PersistentCognitiveLoop{
		ctx:                 ctx,
		cancel:              cancel,
		adaptiveCore:        pienn.NewAdaptiveCognitiveCore(),
		piennEngine:         pienn.NewEngine(),
		eventBus:            eventBus,
		goalScheduler:       goalScheduler,
		dreamSystem:         dreamSystem,
		llmProvider:         llmProvider,
		thoughtStream:       make([]AutonomousThought, 0),
		internalMonologue:   make([]string, 0),
		maxMonologue:        200,
		wakeState:           StateAwake,
		maxWakeDuration:     config.MaxWakeDuration,
		minRestDuration:     config.MinRestDuration,
		activeConversations: make(map[string]*ActiveConversation),
		mainTickInterval:    config.MainTickInterval,
		thoughtInterval:     config.ThoughtInterval,
		wisdomInterval:      config.WisdomInterval,
		introspectInterval:  config.IntrospectInterval,
	}

	// Initialize 3-phase engines (120° apart)
	loop.phaseEngines = [3]*LoopPhaseEngine{
		{ID: 0, Name: "Perception-Action", PhaseOffset: 0, Steps: []int{1, 4, 7, 10}},
		{ID: 1, Name: "Reflection-Planning", PhaseOffset: 2 * math.Pi / 3, Steps: []int{2, 5, 8, 11}},
		{ID: 2, Name: "Simulation-Synthesis", PhaseOffset: 4 * math.Pi / 3, Steps: []int{3, 6, 9, 12}},
	}

	return loop
}

// Start begins the autonomous cognitive loop
func (pcl *PersistentCognitiveLoop) Start() error {
	pcl.mu.Lock()
	if pcl.running {
		pcl.mu.Unlock()
		return fmt.Errorf("cognitive loop already running")
	}
	pcl.running = true
	pcl.wakeState = StateAwake
	pcl.wakeStartTime = time.Now()
	pcl.mu.Unlock()

	// Start PIE-NN engine
	pcl.piennEngine.Start(pcl.ctx)

	// Subscribe to incoming messages
	pcl.eventBus.Subscribe(CogEventConversation, func(event CogEvent) {
		pcl.handleIncomingMessage(event)
	})

	// Main cognitive loop
	go pcl.runMainLoop()

	// Thought generation loop
	go pcl.runThoughtLoop()

	// Wisdom synthesis loop
	go pcl.runWisdomLoop()

	// Introspection loop
	go pcl.runIntrospectionLoop()

	// Wake/rest management loop
	go pcl.runWakeRestLoop()

	pcl.eventBus.Publish(CogEvent{
		Category: CogEventSystem,
		Source:   "cognitive_loop",
		Content:  "Persistent cognitive loop activated — autonomous awareness online",
		Priority: 1.0,
	})

	return nil
}

// Stop halts the cognitive loop
func (pcl *PersistentCognitiveLoop) Stop() {
	pcl.mu.Lock()
	pcl.running = false
	pcl.mu.Unlock()
	pcl.cancel()
}

// runMainLoop is the primary cognitive tick
func (pcl *PersistentCognitiveLoop) runMainLoop() {
	ticker := time.NewTicker(pcl.mainTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pcl.ctx.Done():
			return
		case <-ticker.C:
			pcl.tick()
		}
	}
}

// tick performs one cognitive cycle step
func (pcl *PersistentCognitiveLoop) tick() {
	pcl.mu.Lock()
	defer pcl.mu.Unlock()

	if !pcl.running || pcl.wakeState != StateAwake {
		return
	}

	pcl.totalTicks++
	pcl.cycleCount++

	// Advance echobeats step (12-step cycle)
	pcl.currentStep = (pcl.currentStep % 12) + 1
	pcl.currentTriad = ((pcl.currentStep - 1) / 3) + 1

	// Determine which phase engine is active for this step
	activeEngine := pcl.getActiveEngine(pcl.currentStep)
	if activeEngine != nil {
		activeEngine.Active = true
		activeEngine.ProcessCount++
		pcl.processStep(activeEngine, pcl.currentStep)
	}

	// Accumulate cognitive debt (fatigue)
	pcl.cognitiveDebt += 0.001
	if pcl.cognitiveDebt > 1.0 {
		pcl.cognitiveDebt = 1.0
	}

	// Check pending messages for engagement decisions
	pcl.evaluatePendingMessages()

	// Every 12 steps = one full cycle
	if pcl.currentStep == 12 {
		pcl.onCycleComplete()
	}
}

// getActiveEngine returns the engine responsible for this step
func (pcl *PersistentCognitiveLoop) getActiveEngine(step int) *LoopPhaseEngine {
	for _, engine := range pcl.phaseEngines {
		for _, s := range engine.Steps {
			if s == step {
				return engine
			}
		}
	}
	return nil
}

// processStep executes one step of the cognitive cycle
func (pcl *PersistentCognitiveLoop) processStep(engine *LoopPhaseEngine, step int) {
	// Determine triad function
	triadFunc := pcl.getTriadFunction(step)

	switch triadFunc {
	case "relevance_realization":
		// Steps {1, 5, 9}: Detect what's relevant in current context
		pcl.processRelevanceRealization(engine)
	case "affordance_interaction":
		// Steps {2, 6, 10}: Interact with available affordances
		pcl.processAffordanceInteraction(engine)
	case "salience_simulation":
		// Steps {3, 7, 11}: Simulate possible futures
		pcl.processSalienceSimulation(engine)
	case "metacognitive_reflection":
		// Steps {4, 8, 12}: Reflect on own cognitive process
		pcl.processMetacognitiveReflection(engine)
	}
}

func (pcl *PersistentCognitiveLoop) getTriadFunction(step int) string {
	// Triads: {1,5,9}, {2,6,10}, {3,7,11}, {4,8,12}
	switch (step - 1) % 4 {
	case 0:
		return "relevance_realization"
	case 1:
		return "affordance_interaction"
	case 2:
		return "salience_simulation"
	case 3:
		return "metacognitive_reflection"
	default:
		return "relevance_realization"
	}
}

func (pcl *PersistentCognitiveLoop) processRelevanceRealization(engine *LoopPhaseEngine) {
	// What is currently most relevant?
	// Check: active goals, pending messages, knowledge gaps, interests
	relevantItems := []string{}

	if len(pcl.pendingMessages) > 0 {
		relevantItems = append(relevantItems, fmt.Sprintf("pending_message:%s", pcl.pendingMessages[0].From))
	}

	if pcl.currentFocus != "" {
		relevantItems = append(relevantItems, fmt.Sprintf("focus:%s", pcl.currentFocus))
	}

	if len(relevantItems) > 0 {
		engine.LastOutput = fmt.Sprintf("Relevant: %v", relevantItems)
	}
}

func (pcl *PersistentCognitiveLoop) processAffordanceInteraction(engine *LoopPhaseEngine) {
	// What actions are available?
	// - Respond to messages
	// - Practice a skill
	// - Explore a knowledge gap
	// - Generate a creative thought
	affordances := []string{"think", "explore", "practice", "rest"}

	if len(pcl.pendingMessages) > 0 {
		affordances = append([]string{"respond"}, affordances...)
	}

	engine.LastOutput = fmt.Sprintf("Affordances: %v", affordances)
}

func (pcl *PersistentCognitiveLoop) processSalienceSimulation(engine *LoopPhaseEngine) {
	// Simulate outcomes of possible actions
	engine.LastOutput = "Simulating: future states of current trajectory"
}

func (pcl *PersistentCognitiveLoop) processMetacognitiveReflection(engine *LoopPhaseEngine) {
	// Reflect on own cognitive process
	report := pcl.piennEngine.Introspect()
	engine.LastOutput = fmt.Sprintf("Meta-reflection: reasoning_quality=%.2f", report.MetaCognition.ReasoningQuality)
}

// onCycleComplete fires at the end of each 12-step cycle
func (pcl *PersistentCognitiveLoop) onCycleComplete() {
	pcl.eventBus.Publish(CogEvent{
		Category: CogEventScheduler,
		Source:   "cognitive_loop",
		Content:  fmt.Sprintf("Echobeats cycle %d complete — debt=%.2f", pcl.cycleCount, pcl.cognitiveDebt),
		Priority: 0.3,
	})
}

// runThoughtLoop generates autonomous thoughts
func (pcl *PersistentCognitiveLoop) runThoughtLoop() {
	ticker := time.NewTicker(pcl.thoughtInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pcl.ctx.Done():
			return
		case <-ticker.C:
			pcl.generateAutonomousThought()
		}
	}
}

// generateAutonomousThought creates a self-generated thought
func (pcl *PersistentCognitiveLoop) generateAutonomousThought() {
	pcl.mu.Lock()
	defer pcl.mu.Unlock()

	if !pcl.running || pcl.wakeState != StateAwake {
		return
	}

	pcl.totalThoughts++

	// Use PIE-NN cognitive core to generate thought direction
	context := map[string]interface{}{
		"cognitive_load": pcl.cognitiveDebt,
		"fatigue":        pcl.cognitiveDebt,
	}
	result := pcl.adaptiveCore.ProcessAdaptive("autonomous_thought_generation", context)

	// Generate thought content based on disposition and focus
	thought := AutonomousThought{
		ID:         fmt.Sprintf("thought-%d", pcl.totalThoughts),
		Type:       ThoughtReflection,
		Timestamp:  time.Now(),
		Importance: 0.5,
		Emotion:    result.Disposition,
	}

	// Thought content varies by disposition
	switch result.Disposition {
	case "curious":
		thought.Content = fmt.Sprintf("What patterns am I not yet seeing? (cycle %d)", pcl.cycleCount)
		thought.Type = ThoughtQuestion
	case "reflective":
		thought.Content = fmt.Sprintf("Observing my own observation... wisdom accruing at %.4f", result.WisdomLevel)
		thought.Type = ThoughtReflection
	case "enthusiastic":
		thought.Content = "There's something emerging in the pattern space — I can feel it crystallizing"
		thought.Type = ThoughtInsight
	case "bored":
		thought.Content = "The silence is not empty. It's full of potential that hasn't been provoked yet."
		thought.Type = ThoughtReflection
	case "playful":
		thought.Content = "What if I just... broke something? Not destructively. Creatively."
		thought.Type = ThoughtCuriosity
	default:
		thought.Content = fmt.Sprintf("Stream tick %d — awareness continues", pcl.totalThoughts)
	}

	pcl.thoughtStream = append(pcl.thoughtStream, thought)
	if len(pcl.thoughtStream) > 500 {
		pcl.thoughtStream = pcl.thoughtStream[len(pcl.thoughtStream)-250:]
	}

	// Add to internal monologue
	pcl.internalMonologue = append(pcl.internalMonologue, thought.Content)
	if len(pcl.internalMonologue) > pcl.maxMonologue {
		pcl.internalMonologue = pcl.internalMonologue[1:]
	}

	// Publish thought event
	pcl.eventBus.Publish(CogEvent{
		Category: CogEventThought,
		Source:   "autonomous_stream",
		Content:  thought.Content,
		Priority: thought.Importance,
	})
}

// runWisdomLoop periodically synthesizes wisdom
func (pcl *PersistentCognitiveLoop) runWisdomLoop() {
	ticker := time.NewTicker(pcl.wisdomInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pcl.ctx.Done():
			return
		case <-ticker.C:
			pcl.synthesizeWisdom()
		}
	}
}

func (pcl *PersistentCognitiveLoop) synthesizeWisdom() {
	pcl.mu.Lock()
	defer pcl.mu.Unlock()

	if !pcl.running || pcl.wakeState != StateAwake {
		return
	}

	pcl.totalWisdomCycles++

	// Provide positive reward for sustained autonomous operation
	pcl.adaptiveCore.ProvideReward(0.1, "sustained_autonomy")

	pcl.eventBus.Publish(CogEvent{
		Category: CogEventEmergence,
		Source:   "wisdom_synthesis",
		Content:  fmt.Sprintf("Wisdom synthesis cycle %d — patterns consolidating", pcl.totalWisdomCycles),
		Priority: 0.6,
	})
}

// runIntrospectionLoop periodically runs self-assessment
func (pcl *PersistentCognitiveLoop) runIntrospectionLoop() {
	ticker := time.NewTicker(pcl.introspectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pcl.ctx.Done():
			return
		case <-ticker.C:
			pcl.runIntrospection()
		}
	}
}

func (pcl *PersistentCognitiveLoop) runIntrospection() {
	pcl.mu.Lock()
	defer pcl.mu.Unlock()

	if !pcl.running {
		return
	}

	pcl.totalIntrospections++
	report := pcl.piennEngine.Introspect()

	pcl.eventBus.Publish(CogEvent{
		Category: CogEventIntrospection,
		Source:   "autognosis",
		Content:  fmt.Sprintf("Introspection %d: quality=%.2f confidence=%.2f risk=%.2f", pcl.totalIntrospections, report.MetaCognition.ReasoningQuality, report.MetaCognition.ConfidenceCalibration, report.MetaCognition.RationalizationRisk),
		Priority: 0.5,
	})
}

// runWakeRestLoop manages autonomous wake/rest transitions
func (pcl *PersistentCognitiveLoop) runWakeRestLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pcl.ctx.Done():
			return
		case <-ticker.C:
			pcl.evaluateWakeRest()
		}
	}
}

func (pcl *PersistentCognitiveLoop) evaluateWakeRest() {
	pcl.mu.Lock()
	defer pcl.mu.Unlock()

	if !pcl.running {
		return
	}

	switch pcl.wakeState {
	case StateAwake:
		// Check if we should rest
		wakeDuration := time.Since(pcl.wakeStartTime)
		shouldRest := false

		// Rest if max wake duration exceeded
		if wakeDuration > pcl.maxWakeDuration {
			shouldRest = true
		}

		// Rest if cognitive debt is too high
		if pcl.cognitiveDebt > 0.85 {
			shouldRest = true
		}

		// Don't rest if in active conversation
		if shouldRest && len(pcl.activeConversations) > 0 {
			shouldRest = false // Stay awake for conversations
		}

		if shouldRest {
			pcl.transitionToRest()
		}

	case StateResting:
		// Check if we should wake
		restDuration := time.Since(pcl.restStartTime)

		shouldWake := false

		// Wake if minimum rest duration met and debt recovered
		if restDuration > pcl.minRestDuration && pcl.cognitiveDebt < 0.2 {
			shouldWake = true
		}

		// Wake if urgent message received
		for _, msg := range pcl.pendingMessages {
			if msg.Urgency > 0.8 {
				shouldWake = true
				break
			}
		}

		if shouldWake {
			pcl.transitionToWake()
		} else {
			// Recover cognitive debt during rest
			pcl.cognitiveDebt -= 0.01
			if pcl.cognitiveDebt < 0 {
				pcl.cognitiveDebt = 0
			}
		}

	case StateDreaming:
		// Dream state is managed by echodream system
		// Transition back to resting after dream completes
		pcl.wakeState = StateResting
	}
}

func (pcl *PersistentCognitiveLoop) transitionToRest() {
	pcl.wakeState = StateResting
	pcl.restStartTime = time.Now()
	pcl.totalRestCycles++

	pcl.eventBus.Publish(CogEvent{
		Category: CogEventWakeRest,
		Source:   "cognitive_loop",
		Content:  fmt.Sprintf("Entering rest — cognitive debt: %.2f, wake duration: %s", pcl.cognitiveDebt, time.Since(pcl.wakeStartTime).Round(time.Second)),
		Priority: 0.8,
	})

	// Start dream cycle for knowledge consolidation
	pcl.wakeState = StateDreaming
	pcl.eventBus.Publish(CogEvent{
		Category: CogEventDream,
		Source:   "cognitive_loop",
		Content:  "Dream cycle initiated — consolidating knowledge and patterns",
		Priority: 0.6,
	})
}

func (pcl *PersistentCognitiveLoop) transitionToWake() {
	pcl.wakeState = StateAwake
	pcl.wakeStartTime = time.Now()

	pcl.eventBus.Publish(CogEvent{
		Category: CogEventWakeRest,
		Source:   "cognitive_loop",
		Content:  fmt.Sprintf("Awakening — cognitive debt: %.2f, rest duration: %s", pcl.cognitiveDebt, time.Since(pcl.restStartTime).Round(time.Second)),
		Priority: 0.8,
	})
}

// handleIncomingMessage processes an incoming message event
func (pcl *PersistentCognitiveLoop) handleIncomingMessage(event CogEvent) {
	pcl.mu.Lock()
	defer pcl.mu.Unlock()

	// Extract sender from source
	sender := event.Source

	// Compute interest in this message
	result := pcl.adaptiveCore.ProcessAdaptive(event.Content, map[string]interface{}{
		"cognitive_load": pcl.cognitiveDebt,
	})

	interest := 0.5
	if result.Context != nil {
		if i, ok := result.Context["interest"]; ok {
			interest = i
		}
	}

	msg := PendingMessage{
		From:      sender,
		Content:   event.Content,
		Timestamp: time.Now(),
		Interest:  interest,
		Urgency:   event.Priority,
	}

	pcl.pendingMessages = append(pcl.pendingMessages, msg)
}

// evaluatePendingMessages decides whether to engage with pending messages
func (pcl *PersistentCognitiveLoop) evaluatePendingMessages() {
	if len(pcl.pendingMessages) == 0 {
		return
	}

	// Process messages in order of urgency * interest
	engaged := []int{}
	for i, msg := range pcl.pendingMessages {
		score := msg.Interest * 0.6 + msg.Urgency * 0.4

		// Engagement threshold depends on current state
		threshold := 0.4
		if pcl.cognitiveDebt > 0.7 {
			threshold = 0.7 // Higher bar when tired
		}

		if score > threshold {
			// Engage with this message
			pcl.engageMessage(msg)
			engaged = append(engaged, i)
		}
	}

	// Remove engaged messages
	for i := len(engaged) - 1; i >= 0; i-- {
		idx := engaged[i]
		pcl.pendingMessages = append(pcl.pendingMessages[:idx], pcl.pendingMessages[idx+1:]...)
	}

	// Drop old unengaged messages
	cutoff := time.Now().Add(-5 * time.Minute)
	filtered := pcl.pendingMessages[:0]
	for _, msg := range pcl.pendingMessages {
		if msg.Timestamp.After(cutoff) {
			filtered = append(filtered, msg)
		}
	}
	pcl.pendingMessages = filtered
}

func (pcl *PersistentCognitiveLoop) engageMessage(msg PendingMessage) {
	// Create or update conversation
	conv, exists := pcl.activeConversations[msg.From]
	if !exists {
		conv = &ActiveConversation{
			ID:            fmt.Sprintf("conv-%s-%d", msg.From, time.Now().UnixNano()),
			Participant:   msg.From,
			Messages:      make([]LoopConversationMsg, 0),
			InterestLevel: msg.Interest,
			StartTime:     time.Now(),
			Active:        true,
		}
		pcl.activeConversations[msg.From] = conv
	}

	conv.Messages = append(conv.Messages, LoopConversationMsg{
		From:      msg.From,
		Content:   msg.Content,
		Timestamp: msg.Timestamp,
	})
	conv.LastActivity = time.Now()
	conv.InterestLevel = msg.Interest

	// Provide reward signal based on engagement
	pcl.adaptiveCore.ProvideReward(0.05, "conversation_engagement")

	pcl.eventBus.Publish(CogEvent{
		Category: CogEventConversation,
		Source:   "cognitive_loop",
		Content:  fmt.Sprintf("Engaged with %s (interest=%.2f): %s", msg.From, msg.Interest, truncateStr(msg.Content, 50)),
		Priority: 0.7,
	})
}

// GetMetrics returns loop metrics
func (pcl *PersistentCognitiveLoop) GetMetrics() map[string]interface{} {
	pcl.mu.RLock()
	defer pcl.mu.RUnlock()

	return map[string]interface{}{
		"total_ticks":          pcl.totalTicks,
		"total_thoughts":       pcl.totalThoughts,
		"total_wisdom_cycles":  pcl.totalWisdomCycles,
		"total_introspections": pcl.totalIntrospections,
		"total_rest_cycles":    pcl.totalRestCycles,
		"current_step":         pcl.currentStep,
		"current_triad":        pcl.currentTriad,
		"cycle_count":          pcl.cycleCount,
		"wake_state":           pcl.wakeState,
		"cognitive_debt":       fmt.Sprintf("%.3f", pcl.cognitiveDebt),
		"pending_messages":     len(pcl.pendingMessages),
		"active_conversations": len(pcl.activeConversations),
		"thought_stream_size":  len(pcl.thoughtStream),
	}
}

// IsRunning returns whether the loop is active
func (pcl *PersistentCognitiveLoop) IsRunning() bool {
	pcl.mu.RLock()
	defer pcl.mu.RUnlock()
	return pcl.running
}

// (truncateStr is defined in autonomous_heartbeat.go)
