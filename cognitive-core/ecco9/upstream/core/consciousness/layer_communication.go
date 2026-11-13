package consciousness

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LayerMessage represents a message passed between consciousness layers
type LayerMessage struct {
	ID          string
	Timestamp   time.Time
	FromLayer   LayerIdentifier
	ToLayer     LayerIdentifier
	MessageType MessageType
	Content     string
	Priority    float64
	Context     map[string]interface{}
}

// LayerIdentifier identifies which layer a message is from/to
type LayerIdentifier string

const (
	LayerBasic      LayerIdentifier = "basic"
	LayerReflective LayerIdentifier = "reflective"
	LayerMetaCog    LayerIdentifier = "meta_cognitive"
)

// MessageType categorizes inter-layer messages
type MessageType string

const (
	// Bottom-up messages (basic → reflective → meta)
	MessagePerception MessageType = "perception" // Sensory input
	MessagePattern    MessageType = "pattern"    // Recognized pattern
	MessageAnomaly    MessageType = "anomaly"    // Unexpected observation
	MessageReflection MessageType = "reflection" // Reflective insight
	MessageQuestion   MessageType = "question"   // Inquiry from reflection

	// Top-down messages (meta → reflective → basic)
	MessageGoal       MessageType = "goal"       // High-level goal
	MessageAttention  MessageType = "attention"  // Focus directive
	MessageStrategy   MessageType = "strategy"   // Approach guidance
	MessageInhibition MessageType = "inhibition" // Suppress certain processing

	// Feedback messages
	MessageFeedback  MessageType = "feedback"  // Response to previous message
	MessageEmergence MessageType = "emergence" // Emergent property detected
)

// LayerCommunicationHub manages message passing between consciousness layers
type LayerCommunicationHub struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Message channels for each layer
	basicChannel      chan *LayerMessage
	reflectiveChannel chan *LayerMessage
	metaCogChannel    chan *LayerMessage

	// Message history
	messageHistory []*LayerMessage
	maxHistorySize int

	// Layer handlers
	basicHandler      LayerHandler
	reflectiveHandler LayerHandler
	metaCogHandler    LayerHandler

	// Metrics
	messagesProcessed uint64
	emergenceDetected uint64

	// Control
	running bool
}

// LayerHandler processes messages for a specific layer
type LayerHandler interface {
	ProcessMessage(msg *LayerMessage) ([]*LayerMessage, error)
	GetLayerState() map[string]interface{}
}

// NewLayerCommunicationHub creates a new inter-layer communication system
func NewLayerCommunicationHub() *LayerCommunicationHub {
	ctx, cancel := context.WithCancel(context.Background())

	return &LayerCommunicationHub{
		ctx:               ctx,
		cancel:            cancel,
		basicChannel:      make(chan *LayerMessage, 100),
		reflectiveChannel: make(chan *LayerMessage, 100),
		metaCogChannel:    make(chan *LayerMessage, 100),
		messageHistory:    make([]*LayerMessage, 0),
		maxHistorySize:    1000,
	}
}

// RegisterHandler registers a handler for a specific layer
func (hub *LayerCommunicationHub) RegisterHandler(layer LayerIdentifier, handler LayerHandler) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	switch layer {
	case LayerBasic:
		hub.basicHandler = handler
	case LayerReflective:
		hub.reflectiveHandler = handler
	case LayerMetaCog:
		hub.metaCogHandler = handler
	}
}

// Start begins processing inter-layer messages
func (hub *LayerCommunicationHub) Start() error {
	hub.mu.Lock()
	if hub.running {
		hub.mu.Unlock()
		return fmt.Errorf("communication hub already running")
	}
	hub.running = true
	hub.mu.Unlock()

	// Start message processors for each layer
	go hub.processBasicLayer()
	go hub.processReflectiveLayer()
	go hub.processMetaCogLayer()

	// Start emergence detector
	go hub.detectEmergence()

	return nil
}

// Stop halts message processing
func (hub *LayerCommunicationHub) Stop() {
	hub.mu.Lock()
	if !hub.running {
		hub.mu.Unlock()
		return
	}
	hub.running = false
	hub.mu.Unlock()

	hub.cancel()
}

// SendMessage sends a message to a specific layer
func (hub *LayerCommunicationHub) SendMessage(msg *LayerMessage) error {
	hub.mu.RLock()
	if !hub.running {
		hub.mu.RUnlock()
		return fmt.Errorf("communication hub not running")
	}
	hub.mu.RUnlock()

	// Add to history
	hub.mu.Lock()
	hub.messageHistory = append(hub.messageHistory, msg)
	if len(hub.messageHistory) > hub.maxHistorySize {
		hub.messageHistory = hub.messageHistory[1:]
	}
	hub.mu.Unlock()

	// Route to appropriate channel
	switch msg.ToLayer {
	case LayerBasic:
		select {
		case hub.basicChannel <- msg:
		default:
			return fmt.Errorf("basic layer channel full")
		}
	case LayerReflective:
		select {
		case hub.reflectiveChannel <- msg:
		default:
			return fmt.Errorf("reflective layer channel full")
		}
	case LayerMetaCog:
		select {
		case hub.metaCogChannel <- msg:
		default:
			return fmt.Errorf("meta-cognitive layer channel full")
		}
	default:
		return fmt.Errorf("unknown layer: %s", msg.ToLayer)
	}

	return nil
}

// processBasicLayer processes messages for the basic consciousness layer
func (hub *LayerCommunicationHub) processBasicLayer() {
	for {
		select {
		case <-hub.ctx.Done():
			return
		case msg := <-hub.basicChannel:
			hub.processLayerMessage(LayerBasic, msg)
		}
	}
}

// processReflectiveLayer processes messages for the reflective consciousness layer
func (hub *LayerCommunicationHub) processReflectiveLayer() {
	for {
		select {
		case <-hub.ctx.Done():
			return
		case msg := <-hub.reflectiveChannel:
			hub.processLayerMessage(LayerReflective, msg)
		}
	}
}

// processMetaCogLayer processes messages for the meta-cognitive consciousness layer
func (hub *LayerCommunicationHub) processMetaCogLayer() {
	for {
		select {
		case <-hub.ctx.Done():
			return
		case msg := <-hub.metaCogChannel:
			hub.processLayerMessage(LayerMetaCog, msg)
		}
	}
}

// processLayerMessage processes a message for a specific layer
func (hub *LayerCommunicationHub) processLayerMessage(layer LayerIdentifier, msg *LayerMessage) {
	hub.mu.RLock()
	var handler LayerHandler
	switch layer {
	case LayerBasic:
		handler = hub.basicHandler
	case LayerReflective:
		handler = hub.reflectiveHandler
	case LayerMetaCog:
		handler = hub.metaCogHandler
	}
	hub.mu.RUnlock()

	if handler == nil {
		return // No handler registered
	}

	// Process message and get response messages
	responses, err := handler.ProcessMessage(msg)
	if err != nil {
		return
	}

	// Send response messages
	for _, response := range responses {
		hub.SendMessage(response)
	}

	hub.mu.Lock()
	hub.messagesProcessed++
	hub.mu.Unlock()
}

// detectEmergence monitors for emergent patterns in layer interactions
func (hub *LayerCommunicationHub) detectEmergence() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-hub.ctx.Done():
			return
		case <-ticker.C:
			hub.analyzeEmergence()
		}
	}
}

// analyzeEmergence looks for emergent patterns in message history
func (hub *LayerCommunicationHub) analyzeEmergence() {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	if len(hub.messageHistory) < 10 {
		return
	}

	// Look for feedback loops
	recentMessages := hub.messageHistory[len(hub.messageHistory)-20:]

	// Count message types
	typeCount := make(map[MessageType]int)
	for _, msg := range recentMessages {
		typeCount[msg.MessageType]++
	}

	// Detect emergence patterns
	// Example: High reflection activity followed by meta-cognitive insights
	if typeCount[MessageReflection] > 5 && typeCount[MessageQuestion] > 3 {
		hub.emergenceDetected++
		fmt.Println("🌟 Emergence detected: Reflective inquiry cascade")
	}

	// Example: Bottom-up pattern recognition triggering top-down attention
	if typeCount[MessagePattern] > 3 && typeCount[MessageAttention] > 2 {
		hub.emergenceDetected++
		fmt.Println("🌟 Emergence detected: Pattern-driven attention shift")
	}
}

// GetMetrics returns communication metrics
func (hub *LayerCommunicationHub) GetMetrics() map[string]interface{} {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	return map[string]interface{}{
		"messages_processed":   hub.messagesProcessed,
		"emergence_detected":   hub.emergenceDetected,
		"message_history_size": len(hub.messageHistory),
		"basic_queue":          len(hub.basicChannel),
		"reflective_queue":     len(hub.reflectiveChannel),
		"meta_cog_queue":       len(hub.metaCogChannel),
	}
}

// GetRecentMessages returns recent inter-layer messages
func (hub *LayerCommunicationHub) GetRecentMessages(n int) []*LayerMessage {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	if len(hub.messageHistory) == 0 {
		return []*LayerMessage{}
	}

	start := len(hub.messageHistory) - n
	if start < 0 {
		start = 0
	}

	messages := make([]*LayerMessage, len(hub.messageHistory)-start)
	copy(messages, hub.messageHistory[start:])
	return messages
}

// CreateMessage creates a new layer message
func CreateMessage(from, to LayerIdentifier, msgType MessageType, content string, priority float64) *LayerMessage {
	return &LayerMessage{
		ID:          fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Timestamp:   time.Now(),
		FromLayer:   from,
		ToLayer:     to,
		MessageType: msgType,
		Content:     content,
		Priority:    priority,
		Context:     make(map[string]interface{}),
	}
}
