package consciousness

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// LayerMessage represents a message passed between consciousness layers.
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

// LayerIdentifier identifies which layer a message is from/to.
type LayerIdentifier string

const (
	LayerBasic      LayerIdentifier = "basic"
	LayerReflective LayerIdentifier = "reflective"
	LayerMetaCog    LayerIdentifier = "meta_cognitive"
)

// MessageType categorizes inter-layer messages.
type MessageType string

const (
	// Bottom-up messages (basic → reflective → meta).
	MessagePerception MessageType = "perception" // Sensory input
	MessagePattern    MessageType = "pattern"    // Recognized pattern
	MessageAnomaly    MessageType = "anomaly"    // Unexpected observation
	MessageReflection MessageType = "reflection" // Reflective insight
	MessageQuestion   MessageType = "question"   // Inquiry from reflection

	// Top-down messages (meta → reflective → basic).
	MessageGoal       MessageType = "goal"       // High-level goal
	MessageAttention  MessageType = "attention"  // Focus directive
	MessageStrategy   MessageType = "strategy"   // Approach guidance
	MessageInhibition MessageType = "inhibition" // Suppress certain processing

	// Feedback messages.
	MessageFeedback  MessageType = "feedback"  // Response to previous message
	MessageEmergence MessageType = "emergence" // Emergent property detected
)

// LayerCommunicationHub manages message passing between consciousness layers.
type LayerCommunicationHub struct {
	mu     sync.RWMutex
	ctx    context.Context //nolint:containedctx // The hub owns this lifecycle context.
	cancel context.CancelFunc

	// Message channels for each layer.
	basicChannel      chan *LayerMessage
	reflectiveChannel chan *LayerMessage
	metaCogChannel    chan *LayerMessage

	// Message history.
	messageHistory []*LayerMessage
	maxHistorySize int

	// Layer handlers.
	basicHandler      LayerHandler
	reflectiveHandler LayerHandler
	metaCogHandler    LayerHandler

	// Metrics.
	messagesProcessed uint64
	emergenceDetected uint64

	// Control.
	running bool
}

// LayerHandler processes messages for a specific layer.
type LayerHandler interface {
	ProcessMessage(msg *LayerMessage) ([]*LayerMessage, error)
	GetLayerState() map[string]interface{}
}

// NewLayerCommunicationHub creates a new inter-layer communication system.
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

// RegisterHandler registers a handler for a specific layer.
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

// Start begins processing inter-layer messages.
func (hub *LayerCommunicationHub) Start() error {
	hub.mu.Lock()
	if hub.running {
		hub.mu.Unlock()
		return fmt.Errorf("communication hub already running")
	}
	hub.running = true
	hub.mu.Unlock()

	go hub.processBasicLayer()
	go hub.processReflectiveLayer()
	go hub.processMetaCogLayer()
	go hub.detectEmergence()

	return nil
}

// Stop halts message processing.
func (hub *LayerCommunicationHub) Stop() {
	hub.mu.Lock()
	if !hub.running {
		hub.mu.Unlock()
		return
	}
	hub.running = false
	hub.cancel()
	hub.mu.Unlock()
}

// SendMessage atomically admits and routes an immutable message snapshot. A
// full channel does not create a misleading history entry.
func (hub *LayerCommunicationHub) SendMessage(msg *LayerMessage) error {
	if msg == nil {
		return fmt.Errorf("layer message is required")
	}
	processingCopy, err := cloneLayerMessage(msg)
	if err != nil {
		return fmt.Errorf("clone layer message context: %w", err)
	}
	historyCopy, err := cloneLayerMessage(processingCopy)
	if err != nil {
		return fmt.Errorf("clone layer message history: %w", err)
	}

	hub.mu.Lock()
	defer hub.mu.Unlock()
	if !hub.running {
		return fmt.Errorf("communication hub not running")
	}

	switch processingCopy.ToLayer {
	case LayerBasic:
		select {
		case hub.basicChannel <- processingCopy:
		default:
			return fmt.Errorf("basic layer channel full")
		}
	case LayerReflective:
		select {
		case hub.reflectiveChannel <- processingCopy:
		default:
			return fmt.Errorf("reflective layer channel full")
		}
	case LayerMetaCog:
		select {
		case hub.metaCogChannel <- processingCopy:
		default:
			return fmt.Errorf("meta-cognitive layer channel full")
		}
	default:
		return fmt.Errorf("unknown layer: %s", processingCopy.ToLayer)
	}

	hub.messageHistory = append(hub.messageHistory, historyCopy)
	if len(hub.messageHistory) > hub.maxHistorySize {
		hub.messageHistory = hub.messageHistory[len(hub.messageHistory)-hub.maxHistorySize:]
	}
	return nil
}

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

func (hub *LayerCommunicationHub) processLayerMessage(layer LayerIdentifier, msg *LayerMessage) {
	if msg == nil {
		return
	}
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
		return
	}
	responses, err := handler.ProcessMessage(msg)
	if err != nil {
		return
	}
	for _, response := range responses {
		if response != nil {
			_ = hub.SendMessage(response)
		}
	}

	hub.mu.Lock()
	hub.messagesProcessed++
	hub.mu.Unlock()
}

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

// analyzeEmergence safely examines at most twenty messages. The earlier
// implementation sliced len-20 whenever history had ten entries, panicking for
// every history length from 10 through 19, and mutated metrics under an RLock.
func (hub *LayerCommunicationHub) analyzeEmergence() {
	hub.mu.RLock()
	if len(hub.messageHistory) < 10 {
		hub.mu.RUnlock()
		return
	}
	start := max(0, len(hub.messageHistory)-20)
	recentMessages := make([]*LayerMessage, len(hub.messageHistory)-start)
	copy(recentMessages, hub.messageHistory[start:])
	hub.mu.RUnlock()

	typeCount := make(map[MessageType]int)
	for _, msg := range recentMessages {
		if msg != nil {
			typeCount[msg.MessageType]++
		}
	}

	detected := uint64(0)
	if typeCount[MessageReflection] > 5 && typeCount[MessageQuestion] > 3 {
		detected++
		fmt.Println("🌟 Emergence detected: Reflective inquiry cascade")
	}
	if typeCount[MessagePattern] > 3 && typeCount[MessageAttention] > 2 {
		detected++
		fmt.Println("🌟 Emergence detected: Pattern-driven attention shift")
	}
	if detected > 0 {
		hub.mu.Lock()
		hub.emergenceDetected += detected
		hub.mu.Unlock()
	}
}

// GetMetrics returns communication metrics.
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

// GetRecentMessages returns immutable copies of recent inter-layer messages.
func (hub *LayerCommunicationHub) GetRecentMessages(n int) []*LayerMessage {
	if n <= 0 {
		return []*LayerMessage{}
	}
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	if len(hub.messageHistory) == 0 {
		return []*LayerMessage{}
	}

	start := max(0, len(hub.messageHistory)-n)
	messages := make([]*LayerMessage, 0, len(hub.messageHistory)-start)
	for _, message := range hub.messageHistory[start:] {
		cloned, err := cloneLayerMessage(message)
		if err == nil {
			messages = append(messages, cloned)
		}
	}
	return messages
}

// CreateMessage creates a new layer message.
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

func cloneLayerMessage(message *LayerMessage) (*LayerMessage, error) {
	if message == nil {
		return nil, nil
	}
	clone := *message
	if message.Context == nil {
		return &clone, nil
	}
	encoded, err := json.Marshal(message.Context)
	if err != nil {
		return nil, err
	}
	clone.Context = make(map[string]interface{}, len(message.Context))
	decoder := json.NewDecoder(strings.NewReader(string(encoded)))
	decoder.UseNumber()
	if err := decoder.Decode(&clone.Context); err != nil {
		return nil, err
	}
	return &clone, nil
}
