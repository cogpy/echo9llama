package deeptreeecho

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventBusV2 provides an enhanced event-driven messaging system for Deep Tree Echo
// This is inspired by watermill's architecture for event-driven applications
type EventBusV2 struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc

	// Topic subscriptions
	subscriptions   map[string][]*Subscription
	subsLock        sync.RWMutex

	// Message router
	router          *MessageRouter

	// Message channels
	publishCh       chan *EventMessage
	
	// Metrics
	totalPublished  uint64
	totalDelivered  uint64
	totalAcked      uint64
	totalNacked     uint64

	// Running state
	running         bool
}

// EventMessage is the enhanced message type for cognitive events
type EventMessage struct {
	UUID        string
	Topic       string
	Metadata    map[string]string
	Payload     []byte
	Timestamp   time.Time
	Priority    float64
	
	// Acknowledgement channels
	ack         chan struct{}
	nack        chan struct{}
	ackOnce     sync.Once
	nackOnce    sync.Once
	
	// Context
	ctx         context.Context
}

// Subscription represents a topic subscription
type Subscription struct {
	ID          string
	Topic       string
	Handler     MessageHandler
	Filter      MessageFilter
	Active      bool
}

// MessageHandler processes a cognitive message
type MessageHandler func(msg *EventMessage) error

// MessageFilter determines if a message should be processed
type MessageFilter func(msg *EventMessage) bool

// MessageRouter routes messages between topics
type MessageRouter struct {
	mu          sync.RWMutex
	routes      map[string][]string // source topic -> destination topics
	transforms  map[string]MessageTransform
}

// MessageTransform transforms a message during routing
type MessageTransform func(msg *EventMessage) *EventMessage

// NewEventBusV2 creates a new enhanced event bus
func NewEventBusV2() *EventBusV2 {
	ctx, cancel := context.WithCancel(context.Background())

	return &EventBusV2{
		ctx:           ctx,
		cancel:        cancel,
		subscriptions: make(map[string][]*Subscription),
		router: &MessageRouter{
			routes:     make(map[string][]string),
			transforms: make(map[string]MessageTransform),
		},
		publishCh: make(chan *EventMessage, 1000),
	}
}

// Start begins the event bus
func (eb *EventBusV2) Start() error {
	eb.mu.Lock()
	if eb.running {
		eb.mu.Unlock()
		return fmt.Errorf("event bus already running")
	}
	eb.running = true
	eb.mu.Unlock()

	fmt.Println("📡 Event Bus V2 started")

	// Start the message processing loop
	go eb.processMessages()

	return nil
}

// Stop gracefully stops the event bus
func (eb *EventBusV2) Stop() error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if !eb.running {
		return fmt.Errorf("event bus not running")
	}

	eb.cancel()
	eb.running = false
	fmt.Println("📡 Event Bus V2 stopped")

	return nil
}

// processMessages is the main message processing loop
func (eb *EventBusV2) processMessages() {
	for {
		select {
		case <-eb.ctx.Done():
			return

		case msg := <-eb.publishCh:
			eb.deliverMessage(msg)
		}
	}
}

// deliverMessage delivers a message to all subscribers
func (eb *EventBusV2) deliverMessage(msg *EventMessage) {
	eb.subsLock.RLock()
	subs, exists := eb.subscriptions[msg.Topic]
	eb.subsLock.RUnlock()

	if !exists || len(subs) == 0 {
		return
	}

	for _, sub := range subs {
		if !sub.Active {
			continue
		}

		// Apply filter if present
		if sub.Filter != nil && !sub.Filter(msg) {
			continue
		}

		// Deliver to handler
		go func(s *Subscription, m *EventMessage) {
			err := s.Handler(m)
			if err != nil {
				m.Nack()
			} else {
				m.Ack()
			}
		}(sub, msg)

		eb.mu.Lock()
		eb.totalDelivered++
		eb.mu.Unlock()
	}

	// Route to other topics if configured
	eb.routeMessage(msg)
}

// routeMessage routes a message to other topics based on router configuration
func (eb *EventBusV2) routeMessage(msg *EventMessage) {
	eb.router.mu.RLock()
	destinations, exists := eb.router.routes[msg.Topic]
	eb.router.mu.RUnlock()

	if !exists {
		return
	}

	for _, destTopic := range destinations {
		// Create a copy of the message for the new topic
		newMsg := msg.Copy()
		newMsg.Topic = destTopic

		// Apply transform if present
		eb.router.mu.RLock()
		transform, hasTransform := eb.router.transforms[msg.Topic+"->"+destTopic]
		eb.router.mu.RUnlock()

		if hasTransform && transform != nil {
			newMsg = transform(newMsg)
		}

		// Publish to the new topic
		eb.publishCh <- newMsg
	}
}

// NewMessage creates a new cognitive message
func NewEventMessage(topic string, payload interface{}) (*EventMessage, error) {
	var payloadBytes []byte
	var err error

	switch p := payload.(type) {
	case []byte:
		payloadBytes = p
	case string:
		payloadBytes = []byte(p)
	default:
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	return &EventMessage{
		UUID:      uuid.New().String(),
		Topic:     topic,
		Metadata:  make(map[string]string),
		Payload:   payloadBytes,
		Timestamp: time.Now(),
		Priority:  0.5,
		ack:       make(chan struct{}),
		nack:      make(chan struct{}),
		ctx:       context.Background(),
	}, nil
}

// Ack acknowledges the message
func (m *EventMessage) Ack() {
	m.ackOnce.Do(func() {
		close(m.ack)
	})
}

// Nack negatively acknowledges the message
func (m *EventMessage) Nack() {
	m.nackOnce.Do(func() {
		close(m.nack)
	})
}

// Acked returns a channel that closes when the message is acknowledged
func (m *EventMessage) Acked() <-chan struct{} {
	return m.ack
}

// Nacked returns a channel that closes when the message is negatively acknowledged
func (m *EventMessage) Nacked() <-chan struct{} {
	return m.nack
}

// Copy creates a copy of the message
func (m *EventMessage) Copy() *EventMessage {
	newMsg := &EventMessage{
		UUID:      uuid.New().String(),
		Topic:     m.Topic,
		Metadata:  make(map[string]string),
		Payload:   make([]byte, len(m.Payload)),
		Timestamp: time.Now(),
		Priority:  m.Priority,
		ack:       make(chan struct{}),
		nack:      make(chan struct{}),
		ctx:       m.ctx,
	}
	copy(newMsg.Payload, m.Payload)
	for k, v := range m.Metadata {
		newMsg.Metadata[k] = v
	}
	return newMsg
}

// SetMetadata sets a metadata value
func (m *EventMessage) SetMetadata(key, value string) {
	m.Metadata[key] = value
}

// GetMetadata gets a metadata value
func (m *EventMessage) GetMetadata(key string) string {
	return m.Metadata[key]
}

// Publish publishes a message to a topic
func (eb *EventBusV2) Publish(topic string, payload interface{}) error {
	msg, err := NewEventMessage(topic, payload)
	if err != nil {
		return err
	}

	eb.publishCh <- msg

	eb.mu.Lock()
	eb.totalPublished++
	eb.mu.Unlock()

	return nil
}

// PublishWithPriority publishes a message with a specific priority
func (eb *EventBusV2) PublishWithPriority(topic string, payload interface{}, priority float64) error {
	msg, err := NewEventMessage(topic, payload)
	if err != nil {
		return err
	}
	msg.Priority = priority

	eb.publishCh <- msg

	eb.mu.Lock()
	eb.totalPublished++
	eb.mu.Unlock()

	return nil
}

// Subscribe subscribes to a topic
func (eb *EventBusV2) Subscribe(topic string, handler MessageHandler) (string, error) {
	sub := &Subscription{
		ID:      uuid.New().String(),
		Topic:   topic,
		Handler: handler,
		Active:  true,
	}

	eb.subsLock.Lock()
	eb.subscriptions[topic] = append(eb.subscriptions[topic], sub)
	eb.subsLock.Unlock()

	return sub.ID, nil
}

// SubscribeWithFilter subscribes to a topic with a filter
func (eb *EventBusV2) SubscribeWithFilter(topic string, handler MessageHandler, filter MessageFilter) (string, error) {
	sub := &Subscription{
		ID:      uuid.New().String(),
		Topic:   topic,
		Handler: handler,
		Filter:  filter,
		Active:  true,
	}

	eb.subsLock.Lock()
	eb.subscriptions[topic] = append(eb.subscriptions[topic], sub)
	eb.subsLock.Unlock()

	return sub.ID, nil
}

// Unsubscribe removes a subscription
func (eb *EventBusV2) Unsubscribe(subscriptionID string) error {
	eb.subsLock.Lock()
	defer eb.subsLock.Unlock()

	for topic, subs := range eb.subscriptions {
		for i, sub := range subs {
			if sub.ID == subscriptionID {
				eb.subscriptions[topic] = append(subs[:i], subs[i+1:]...)
				return nil
			}
		}
	}

	return fmt.Errorf("subscription not found: %s", subscriptionID)
}

// AddRoute adds a routing rule from source to destination topic
func (eb *EventBusV2) AddRoute(sourceTopic, destTopic string) {
	eb.router.mu.Lock()
	eb.router.routes[sourceTopic] = append(eb.router.routes[sourceTopic], destTopic)
	eb.router.mu.Unlock()
}

// AddRouteWithTransform adds a routing rule with a transform function
func (eb *EventBusV2) AddRouteWithTransform(sourceTopic, destTopic string, transform MessageTransform) {
	eb.router.mu.Lock()
	eb.router.routes[sourceTopic] = append(eb.router.routes[sourceTopic], destTopic)
	eb.router.transforms[sourceTopic+"->"+destTopic] = transform
	eb.router.mu.Unlock()
}

// GetMetrics returns event bus metrics
func (eb *EventBusV2) GetMetrics() map[string]interface{} {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	eb.subsLock.RLock()
	topicCount := len(eb.subscriptions)
	totalSubs := 0
	for _, subs := range eb.subscriptions {
		totalSubs += len(subs)
	}
	eb.subsLock.RUnlock()

	return map[string]interface{}{
		"running":         eb.running,
		"topics":          topicCount,
		"subscriptions":   totalSubs,
		"total_published": eb.totalPublished,
		"total_delivered": eb.totalDelivered,
		"total_acked":     eb.totalAcked,
		"total_nacked":    eb.totalNacked,
	}
}

// ContributeToGestalt provides event bus state for the global gestalt
func (eb *EventBusV2) ContributeToGestalt() map[string]interface{} {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	eb.subsLock.RLock()
	topics := make([]string, 0, len(eb.subscriptions))
	for topic := range eb.subscriptions {
		topics = append(topics, topic)
	}
	eb.subsLock.RUnlock()

	return map[string]interface{}{
		"running":         eb.running,
		"active_topics":   topics,
		"total_published": eb.totalPublished,
		"total_delivered": eb.totalDelivered,
	}
}

// Predefined cognitive topics
const (
	TopicPerception       = "cognitive.perception"
	TopicAction           = "cognitive.action"
	TopicSimulation       = "cognitive.simulation"
	TopicEmergence        = "cognitive.emergence"
	TopicWisdom           = "cognitive.wisdom"
	TopicSkillLearning    = "cognitive.skill_learning"
	TopicDiscussion       = "cognitive.discussion"
	TopicGoalScheduled    = "cognitive.goal_scheduled"
	TopicGoalCompleted    = "cognitive.goal_completed"
	TopicMemoryStored     = "cognitive.memory_stored"
	TopicMemoryRecalled   = "cognitive.memory_recalled"
	TopicStateChange      = "cognitive.state_change"
	TopicWakeEvent        = "cognitive.wake"
	TopicRestEvent        = "cognitive.rest"
	TopicSelfUpdate       = "cognitive.self_update"
)

// SetupCognitiveRoutes sets up default routing for cognitive events
func (eb *EventBusV2) SetupCognitiveRoutes() {
	// Route perception events to simulation for reflection
	eb.AddRoute(TopicPerception, TopicSimulation)

	// Route emergence events to wisdom for integration
	eb.AddRoute(TopicEmergence, TopicWisdom)

	// Route skill learning events to memory for storage
	eb.AddRoute(TopicSkillLearning, TopicMemoryStored)

	// Route goal completion to wisdom for pattern extraction
	eb.AddRoute(TopicGoalCompleted, TopicWisdom)

	fmt.Println("   Cognitive routes configured")
}

// PublishCognitiveEvent publishes a cognitive event with standard metadata
func (eb *EventBusV2) PublishCognitiveEvent(topic string, eventType string, data interface{}) error {
	msg, err := NewEventMessage(topic, data)
	if err != nil {
		return err
	}

	msg.SetMetadata("event_type", eventType)
	msg.SetMetadata("source", "deep_tree_echo")
	msg.SetMetadata("timestamp", time.Now().Format(time.RFC3339))

	eb.publishCh <- msg

	eb.mu.Lock()
	eb.totalPublished++
	eb.mu.Unlock()

	return nil
}
