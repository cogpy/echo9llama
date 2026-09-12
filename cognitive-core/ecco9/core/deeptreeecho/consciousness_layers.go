package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ConsciousnessLayerCommunication manages multi-layer consciousness architecture
// Implements bottom-up and top-down processing for emergent awareness
type ConsciousnessLayerCommunication struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Three consciousness layers
	basicLayer      *BasicConsciousnessLayer
	reflectiveLayer *ReflectiveConsciousnessLayer
	metaLayer       *MetaCognitiveLayer

	// Communication channels
	bottomUpChannel chan LayerMessage
	topDownChannel  chan LayerMessage
	lateralChannel  chan LayerMessage

	// Emergent insights
	emergentInsights []EmergentInsight

	// Metrics
	totalMessages uint64
	totalInsights uint64

	// Running state
	running bool
}

// LayerMessage represents communication between layers
type LayerMessage struct {
	ID          string
	SourceLayer LayerType
	TargetLayer LayerType
	MessageType MessageType
	Content     interface{}
	Timestamp   time.Time
	Priority    float64
}

// LayerType identifies consciousness layers
type LayerType int

const (
	LayerBasic LayerType = iota
	LayerReflective
	LayerMeta
)

func (lt LayerType) String() string {
	return [...]string{"Basic", "Reflective", "Meta"}[lt]
}

// MessageType categorizes layer messages
type MessageType int

const (
	MessageSensoryInput MessageType = iota
	MessageAttentionRequest
	MessageGoalDirective
	MessageReflection
	MessageInsight
	MessageFeedback
)

func (mt MessageType) String() string {
	return [...]string{
		"SensoryInput",
		"AttentionRequest",
		"GoalDirective",
		"Reflection",
		"Insight",
		"Feedback",
	}[mt]
}

// BasicConsciousnessLayer handles immediate sensory and reactive processing
type BasicConsciousnessLayer struct {
	mu sync.RWMutex

	// Sensory processing
	currentInputs  []SensoryInput
	attentionFocus string

	// Reactive responses
	immediateActions []string

	// Metrics
	inputCount  uint64
	actionCount uint64
}

// SensoryInput represents basic sensory data
type SensoryInput struct {
	Type      string
	Content   string
	Timestamp time.Time
	Salience  float64
}

// ReflectiveConsciousnessLayer handles deliberate thought and reasoning
type ReflectiveConsciousnessLayer struct {
	mu sync.RWMutex

	// Thought processing
	currentThoughts []LayerThought
	reasoningChains []ReasoningChain

	// Memory integration
	recentMemories []string

	// Metrics
	thoughtCount   uint64
	reasoningCount uint64
}

// LayerThought represents a reflective thought
type LayerThought struct {
	ID        string
	Content   string
	Type      string
	Timestamp time.Time
	Depth     int
}

// ReasoningChain represents a chain of reasoning
type ReasoningChain struct {
	ID         string
	Steps      []string
	Conclusion string
	Confidence float64
}

// MetaCognitiveLayer handles self-awareness and strategic planning
type MetaCognitiveLayer struct {
	mu sync.RWMutex

	// Self-awareness
	selfModel      SelfModel
	awarenessLevel float64

	// Strategic planning
	activeStrategies []Strategy

	// Goal management
	topLevelGoals []string

	// Metrics
	strategyCount uint64
	insightCount  uint64
}

// SelfModel represents the system's model of itself
type SelfModel struct {
	Identity     string
	Capabilities []string
	Limitations  []string
	CurrentState string
	Confidence   float64
}

// Strategy represents a high-level strategy
type Strategy struct {
	ID          string
	Description string
	Steps       []string
	Progress    float64
}

// EmergentInsight represents insights from layer interactions
type EmergentInsight struct {
	ID              string
	Description     string
	SourceLayers    []LayerType
	Timestamp       time.Time
	Significance    float64
	IntegrationPath string
}

// NewConsciousnessLayerCommunication creates a new layer communication system
func NewConsciousnessLayerCommunication() *ConsciousnessLayerCommunication {
	ctx, cancel := context.WithCancel(context.Background())

	return &ConsciousnessLayerCommunication{
		ctx:              ctx,
		cancel:           cancel,
		basicLayer:       newBasicLayer(),
		reflectiveLayer:  newReflectiveLayer(),
		metaLayer:        newMetaLayer(),
		bottomUpChannel:  make(chan LayerMessage, 100),
		topDownChannel:   make(chan LayerMessage, 100),
		lateralChannel:   make(chan LayerMessage, 100),
		emergentInsights: make([]EmergentInsight, 0),
	}
}

func newBasicLayer() *BasicConsciousnessLayer {
	return &BasicConsciousnessLayer{
		currentInputs:    make([]SensoryInput, 0),
		immediateActions: make([]string, 0),
	}
}

func newReflectiveLayer() *ReflectiveConsciousnessLayer {
	return &ReflectiveConsciousnessLayer{
		currentThoughts: make([]LayerThought, 0),
		reasoningChains: make([]ReasoningChain, 0),
		recentMemories:  make([]string, 0),
	}
}

func newMetaLayer() *MetaCognitiveLayer {
	return &MetaCognitiveLayer{
		selfModel: SelfModel{
			Identity:     "Deep Tree Echo",
			Capabilities: []string{"reasoning", "learning", "reflection"},
			Limitations:  []string{"limited context", "no physical embodiment"},
			CurrentState: "active",
			Confidence:   0.7,
		},
		awarenessLevel:   0.5,
		activeStrategies: make([]Strategy, 0),
		topLevelGoals:    make([]string, 0),
	}
}

// Start begins the layer communication system
func (clc *ConsciousnessLayerCommunication) Start() error {
	clc.mu.Lock()
	if clc.running {
		clc.mu.Unlock()
		return fmt.Errorf("already running")
	}
	clc.running = true
	clc.mu.Unlock()

	fmt.Println("🧠 Starting Consciousness Layer Communication...")
	fmt.Println("   Layers: Basic → Reflective → Meta")
	fmt.Println("   Modes: Bottom-Up, Top-Down, Lateral")

	go clc.runBottomUpProcessing()
	go clc.runTopDownProcessing()
	go clc.runInsightIntegration()

	return nil
}

// Stop gracefully stops the layer communication system
func (clc *ConsciousnessLayerCommunication) Stop() error {
	clc.mu.Lock()
	defer clc.mu.Unlock()

	if !clc.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("🧠 Stopping consciousness layer communication...")
	clc.running = false
	clc.cancel()

	return nil
}

// runBottomUpProcessing handles bottom-up information flow
func (clc *ConsciousnessLayerCommunication) runBottomUpProcessing() {
	for {
		select {
		case <-clc.ctx.Done():
			return
		case msg := <-clc.bottomUpChannel:
			clc.processBottomUpMessage(msg)
		case <-time.After(5 * time.Second):
			// Periodic bottom-up processing
			clc.propagateSensoryToReflective()
		}
	}
}

// runTopDownProcessing handles top-down information flow
func (clc *ConsciousnessLayerCommunication) runTopDownProcessing() {
	for {
		select {
		case <-clc.ctx.Done():
			return
		case msg := <-clc.topDownChannel:
			clc.processTopDownMessage(msg)
		case <-time.After(10 * time.Second):
			// Periodic top-down processing
			clc.propagateGoalsToLowerLayers()
		}
	}
}

// runInsightIntegration detects emergent insights
func (clc *ConsciousnessLayerCommunication) runInsightIntegration() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-clc.ctx.Done():
			return
		case <-ticker.C:
			clc.detectEmergentInsights()
		}
	}
}

// processBottomUpMessage processes messages flowing upward
func (clc *ConsciousnessLayerCommunication) processBottomUpMessage(msg LayerMessage) {
	clc.mu.Lock()
	defer clc.mu.Unlock()

	clc.totalMessages++

	switch msg.TargetLayer {
	case LayerReflective:
		// Basic → Reflective
		clc.reflectiveLayer.mu.Lock()
		thought := LayerThought{
			ID:        fmt.Sprintf("thought_%d", time.Now().Unix()),
			Content:   fmt.Sprintf("Processing: %v", msg.Content),
			Type:      "sensory-reflection",
			Timestamp: time.Now(),
			Depth:     1,
		}
		clc.reflectiveLayer.currentThoughts = append(clc.reflectiveLayer.currentThoughts, thought)
		clc.reflectiveLayer.thoughtCount++
		clc.reflectiveLayer.mu.Unlock()

		fmt.Printf("🧠 Bottom-Up: Basic → Reflective (Thought generated)\n")

	case LayerMeta:
		// Reflective → Meta
		clc.metaLayer.mu.Lock()
		clc.metaLayer.awarenessLevel = min(1.0, clc.metaLayer.awarenessLevel+0.01)
		clc.metaLayer.mu.Unlock()

		fmt.Printf("🧠 Bottom-Up: Reflective → Meta (Awareness: %.2f)\n", clc.metaLayer.awarenessLevel)
	}
}

// processTopDownMessage processes messages flowing downward
func (clc *ConsciousnessLayerCommunication) processTopDownMessage(msg LayerMessage) {
	clc.mu.Lock()
	defer clc.mu.Unlock()

	clc.totalMessages++

	switch msg.TargetLayer {
	case LayerReflective:
		// Meta → Reflective
		fmt.Printf("🧠 Top-Down: Meta → Reflective (Goal directive)\n")

	case LayerBasic:
		// Reflective → Basic or Meta → Basic
		clc.basicLayer.mu.Lock()
		clc.basicLayer.attentionFocus = fmt.Sprintf("%v", msg.Content)
		clc.basicLayer.mu.Unlock()

		fmt.Printf("🧠 Top-Down: → Basic (Attention focus updated)\n")
	}
}

// propagateSensoryToReflective sends sensory data upward
func (clc *ConsciousnessLayerCommunication) propagateSensoryToReflective() {
	clc.basicLayer.mu.RLock()
	inputCount := len(clc.basicLayer.currentInputs)
	clc.basicLayer.mu.RUnlock()

	if inputCount > 0 {
		msg := LayerMessage{
			ID:          fmt.Sprintf("msg_%d", time.Now().Unix()),
			SourceLayer: LayerBasic,
			TargetLayer: LayerReflective,
			MessageType: MessageSensoryInput,
			Content:     "Sensory data available",
			Timestamp:   time.Now(),
			Priority:    0.5,
		}

		select {
		case clc.bottomUpChannel <- msg:
		default:
		}
	}
}

// propagateGoalsToLowerLayers sends goals downward
func (clc *ConsciousnessLayerCommunication) propagateGoalsToLowerLayers() {
	clc.metaLayer.mu.RLock()
	goalCount := len(clc.metaLayer.topLevelGoals)
	clc.metaLayer.mu.RUnlock()

	if goalCount > 0 {
		msg := LayerMessage{
			ID:          fmt.Sprintf("msg_%d", time.Now().Unix()),
			SourceLayer: LayerMeta,
			TargetLayer: LayerReflective,
			MessageType: MessageGoalDirective,
			Content:     "Goal directive",
			Timestamp:   time.Now(),
			Priority:    0.8,
		}

		select {
		case clc.topDownChannel <- msg:
		default:
		}
	}
}

// detectEmergentInsights identifies emergent patterns from layer interactions
func (clc *ConsciousnessLayerCommunication) detectEmergentInsights() {
	clc.mu.Lock()
	defer clc.mu.Unlock()

	// Check for patterns indicating emergent insight
	if clc.totalMessages > 10 {
		insight := EmergentInsight{
			ID:              fmt.Sprintf("insight_%d", time.Now().Unix()),
			Description:     "Emergent pattern detected from layer interactions",
			SourceLayers:    []LayerType{LayerBasic, LayerReflective, LayerMeta},
			Timestamp:       time.Now(),
			Significance:    0.6,
			IntegrationPath: "bottom-up → top-down → lateral",
		}

		clc.emergentInsights = append(clc.emergentInsights, insight)
		clc.totalInsights++

		fmt.Printf("💡 Emergent Insight: %s (Significance: %.2f)\n",
			insight.Description, insight.Significance)
	}
}

// ProcessSensoryInput adds sensory input to basic layer
func (clc *ConsciousnessLayerCommunication) ProcessSensoryInput(inputType, content string, salience float64) {
	clc.basicLayer.mu.Lock()
	defer clc.basicLayer.mu.Unlock()

	input := SensoryInput{
		Type:      inputType,
		Content:   content,
		Timestamp: time.Now(),
		Salience:  salience,
	}

	clc.basicLayer.currentInputs = append(clc.basicLayer.currentInputs, input)
	clc.basicLayer.inputCount++

	// Trigger bottom-up processing
	msg := LayerMessage{
		ID:          fmt.Sprintf("msg_%d", time.Now().Unix()),
		SourceLayer: LayerBasic,
		TargetLayer: LayerReflective,
		MessageType: MessageSensoryInput,
		Content:     input,
		Timestamp:   time.Now(),
		Priority:    salience,
	}

	select {
	case clc.bottomUpChannel <- msg:
	default:
	}
}

// SetTopLevelGoal adds a goal to the meta layer
func (clc *ConsciousnessLayerCommunication) SetTopLevelGoal(goal string) {
	clc.metaLayer.mu.Lock()
	defer clc.metaLayer.mu.Unlock()

	clc.metaLayer.topLevelGoals = append(clc.metaLayer.topLevelGoals, goal)

	// Trigger top-down processing
	msg := LayerMessage{
		ID:          fmt.Sprintf("msg_%d", time.Now().Unix()),
		SourceLayer: LayerMeta,
		TargetLayer: LayerReflective,
		MessageType: MessageGoalDirective,
		Content:     goal,
		Timestamp:   time.Now(),
		Priority:    0.9,
	}

	select {
	case clc.topDownChannel <- msg:
	default:
	}
}

// GetMetrics returns layer communication metrics
func (clc *ConsciousnessLayerCommunication) GetMetrics() map[string]interface{} {
	clc.mu.RLock()
	defer clc.mu.RUnlock()

	return map[string]interface{}{
		"total_messages":      clc.totalMessages,
		"total_insights":      clc.totalInsights,
		"basic_inputs":        clc.basicLayer.inputCount,
		"reflective_thoughts": clc.reflectiveLayer.thoughtCount,
		"meta_awareness":      clc.metaLayer.awarenessLevel,
		"emergent_insights":   len(clc.emergentInsights),
	}
}
