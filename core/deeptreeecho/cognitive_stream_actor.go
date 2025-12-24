package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// generateStreamID generates a unique ID for stream actors
func generateStreamID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, uuid.New().String()[:8])
}

// StreamType represents the type of cognitive stream
type StreamType int

const (
	StreamTypePerception StreamType = iota
	StreamTypeAction
	StreamTypeSimulation
)

func (st StreamType) String() string {
	switch st {
	case StreamTypePerception:
		return "perception"
	case StreamTypeAction:
		return "action"
	case StreamTypeSimulation:
		return "simulation"
	default:
		return "unknown"
	}
}

// StreamPhase represents the current phase within the 12-step cycle
type StreamPhase int

const (
	// Triads occur every 4 steps: {1,5,9}, {2,6,10}, {3,7,11}, {4,8,12}
	PhaseTriad1 StreamPhase = iota // Steps 1, 5, 9
	PhaseTriad2                     // Steps 2, 6, 10
	PhaseTriad3                     // Steps 3, 7, 11
	PhaseTriad4                     // Steps 4, 8, 12
)

func (sp StreamPhase) String() string {
	switch sp {
	case PhaseTriad1:
		return "triad_1"
	case PhaseTriad2:
		return "triad_2"
	case PhaseTriad3:
		return "triad_3"
	case PhaseTriad4:
		return "triad_4"
	default:
		return "unknown"
	}
}

// CognitiveMessage represents a message passed between cognitive stream actors
type CognitiveMessage struct {
	Type      CognitiveMessageType
	Source    StreamType
	Target    StreamType
	Step      int
	Phase     StreamPhase
	Content   interface{}
	Timestamp time.Time
}

// CognitiveMessageType represents the type of cognitive message
type CognitiveMessageType int

const (
	MsgTypePerception CognitiveMessageType = iota
	MsgTypeAction
	MsgTypeSimulation
	MsgTypeSyncRequest
	MsgTypeSyncResponse
	MsgTypeStateUpdate
	MsgTypeEmergence
)

// StreamState represents the internal state of a cognitive stream
type StreamState struct {
	CurrentStep     int
	CurrentPhase    StreamPhase
	ProcessingLoad  float64
	LastActivity    time.Time
	AccumulatedData []interface{}
	Insights        []string
}

// CognitiveStreamActor represents an actor-based cognitive stream
// This implements the ergo actor pattern for the 3 concurrent cognitive loops
type CognitiveStreamActor struct {
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	
	// Identity
	id            string
	streamType    StreamType
	phaseOffset   int // 0, 4, or 8 for the 120-degree phase offset
	
	// State
	state         StreamState
	
	// Communication channels (actor mailbox simulation)
	inbox         chan CognitiveMessage
	outbox        chan CognitiveMessage
	
	// Peer references for inter-stream communication
	peers         map[StreamType]chan CognitiveMessage
	
	// Processing callbacks
	onPerceive    func(data interface{}) interface{}
	onAct         func(data interface{}) interface{}
	onSimulate    func(data interface{}) interface{}
	onEmergence   func(pattern string, strength float64)
	
	// Running state
	running       bool
}

// NewCognitiveStreamActor creates a new cognitive stream actor
func NewCognitiveStreamActor(streamType StreamType) *CognitiveStreamActor {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Calculate phase offset based on stream type
	// Streams are phased 4 steps apart (120 degrees) over the 12-step cycle
	var phaseOffset int
	switch streamType {
	case StreamTypePerception:
		phaseOffset = 0 // Steps 1, 2, 3, 4
	case StreamTypeAction:
		phaseOffset = 4 // Steps 5, 6, 7, 8
	case StreamTypeSimulation:
		phaseOffset = 8 // Steps 9, 10, 11, 12
	}
	
	return &CognitiveStreamActor{
		ctx:         ctx,
		cancel:      cancel,
		id:          generateStreamID(streamType.String()),
		streamType:  streamType,
		phaseOffset: phaseOffset,
		state: StreamState{
			CurrentStep:     0,
			CurrentPhase:    PhaseTriad1,
			ProcessingLoad:  0.0,
			LastActivity:    time.Now(),
			AccumulatedData: make([]interface{}, 0),
			Insights:        make([]string, 0),
		},
		inbox:  make(chan CognitiveMessage, 100),
		outbox: make(chan CognitiveMessage, 100),
		peers:  make(map[StreamType]chan CognitiveMessage),
	}
}

// SetPeers sets the peer stream references for inter-stream communication
func (csa *CognitiveStreamActor) SetPeers(peers map[StreamType]chan CognitiveMessage) {
	csa.mu.Lock()
	defer csa.mu.Unlock()
	csa.peers = peers
}

// SetCallbacks sets the processing callbacks
func (csa *CognitiveStreamActor) SetCallbacks(
	onPerceive func(data interface{}) interface{},
	onAct func(data interface{}) interface{},
	onSimulate func(data interface{}) interface{},
	onEmergence func(pattern string, strength float64),
) {
	csa.mu.Lock()
	defer csa.mu.Unlock()
	csa.onPerceive = onPerceive
	csa.onAct = onAct
	csa.onSimulate = onSimulate
	csa.onEmergence = onEmergence
}

// GetInbox returns the inbox channel for receiving messages
func (csa *CognitiveStreamActor) GetInbox() chan CognitiveMessage {
	return csa.inbox
}

// GetOutbox returns the outbox channel for sending messages
func (csa *CognitiveStreamActor) GetOutbox() chan CognitiveMessage {
	return csa.outbox
}

// Start begins the cognitive stream actor
func (csa *CognitiveStreamActor) Start() error {
	csa.mu.Lock()
	if csa.running {
		csa.mu.Unlock()
		return fmt.Errorf("stream actor %s already running", csa.id)
	}
	csa.running = true
	csa.mu.Unlock()
	
	// Start message processing loop
	go csa.messageLoop()
	
	// Start cognitive processing loop
	go csa.cognitiveLoop()
	
	fmt.Printf("🌊 Cognitive Stream Actor [%s] started (phase offset: %d)\n", csa.streamType, csa.phaseOffset)
	return nil
}

// Stop stops the cognitive stream actor
func (csa *CognitiveStreamActor) Stop() error {
	csa.mu.Lock()
	defer csa.mu.Unlock()
	
	if !csa.running {
		return fmt.Errorf("stream actor %s not running", csa.id)
	}
	
	csa.running = false
	csa.cancel()
	
	fmt.Printf("🌊 Cognitive Stream Actor [%s] stopped\n", csa.streamType)
	return nil
}

// messageLoop processes incoming messages
func (csa *CognitiveStreamActor) messageLoop() {
	for {
		select {
		case <-csa.ctx.Done():
			return
		case msg := <-csa.inbox:
			csa.handleMessage(msg)
		}
	}
}

// handleMessage handles an incoming cognitive message
func (csa *CognitiveStreamActor) handleMessage(msg CognitiveMessage) {
	csa.mu.Lock()
	defer csa.mu.Unlock()
	
	switch msg.Type {
	case MsgTypePerception:
		// Process perception data from peer
		csa.state.AccumulatedData = append(csa.state.AccumulatedData, msg.Content)
		
	case MsgTypeAction:
		// Process action result from peer
		csa.state.AccumulatedData = append(csa.state.AccumulatedData, msg.Content)
		
	case MsgTypeSimulation:
		// Process simulation result from peer
		csa.state.AccumulatedData = append(csa.state.AccumulatedData, msg.Content)
		
	case MsgTypeSyncRequest:
		// Respond with current state
		csa.sendToPeer(msg.Source, CognitiveMessage{
			Type:      MsgTypeSyncResponse,
			Source:    csa.streamType,
			Target:    msg.Source,
			Step:      csa.state.CurrentStep,
			Phase:     csa.state.CurrentPhase,
			Content:   csa.state,
			Timestamp: time.Now(),
		})
		
	case MsgTypeStateUpdate:
		// Update based on peer state
		// This enables the interdependent self-balancing mechanism
		
	case MsgTypeEmergence:
		// Handle emergence detection from peer
		if pattern, ok := msg.Content.(string); ok {
			csa.state.Insights = append(csa.state.Insights, pattern)
		}
	}
	
	csa.state.LastActivity = time.Now()
}

// cognitiveLoop runs the main cognitive processing cycle
func (csa *CognitiveStreamActor) cognitiveLoop() {
	// Ticker for the 12-step cycle
	// Each step is approximately 500ms, full cycle is 6 seconds
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-csa.ctx.Done():
			return
		case <-ticker.C:
			csa.executeCognitiveStep()
		}
	}
}

// executeCognitiveStep executes one step of the cognitive cycle
func (csa *CognitiveStreamActor) executeCognitiveStep() {
	csa.mu.Lock()
	defer csa.mu.Unlock()
	
	// Advance step (0-11 for 12-step cycle)
	csa.state.CurrentStep = (csa.state.CurrentStep + 1) % 12
	
	// Calculate effective step with phase offset
	effectiveStep := (csa.state.CurrentStep + csa.phaseOffset) % 12
	
	// Determine current phase based on effective step
	// Triads: {1,5,9}, {2,6,10}, {3,7,11}, {4,8,12} (1-indexed)
	switch effectiveStep % 4 {
	case 0:
		csa.state.CurrentPhase = PhaseTriad4
	case 1:
		csa.state.CurrentPhase = PhaseTriad1
	case 2:
		csa.state.CurrentPhase = PhaseTriad2
	case 3:
		csa.state.CurrentPhase = PhaseTriad3
	}
	
	// Execute stream-specific processing
	var result interface{}
	switch csa.streamType {
	case StreamTypePerception:
		if csa.onPerceive != nil {
			result = csa.onPerceive(csa.state.AccumulatedData)
		}
	case StreamTypeAction:
		if csa.onAct != nil {
			result = csa.onAct(csa.state.AccumulatedData)
		}
	case StreamTypeSimulation:
		if csa.onSimulate != nil {
			result = csa.onSimulate(csa.state.AccumulatedData)
		}
	}
	
	// Broadcast result to peers
	if result != nil {
		csa.broadcastToPeers(CognitiveMessage{
			Type:      csa.getMessageType(),
			Source:    csa.streamType,
			Step:      csa.state.CurrentStep,
			Phase:     csa.state.CurrentPhase,
			Content:   result,
			Timestamp: time.Now(),
		})
	}
	
	// Check for emergent patterns
	csa.detectEmergence()
	
	// Update processing load
	csa.updateProcessingLoad()
}

// getMessageType returns the message type for this stream
func (csa *CognitiveStreamActor) getMessageType() CognitiveMessageType {
	switch csa.streamType {
	case StreamTypePerception:
		return MsgTypePerception
	case StreamTypeAction:
		return MsgTypeAction
	case StreamTypeSimulation:
		return MsgTypeSimulation
	default:
		return MsgTypeStateUpdate
	}
}

// sendToPeer sends a message to a specific peer
func (csa *CognitiveStreamActor) sendToPeer(target StreamType, msg CognitiveMessage) {
	if peer, exists := csa.peers[target]; exists {
		select {
		case peer <- msg:
		default:
			// Peer inbox full, drop message
		}
	}
}

// broadcastToPeers sends a message to all peers
func (csa *CognitiveStreamActor) broadcastToPeers(msg CognitiveMessage) {
	for peerType, peer := range csa.peers {
		if peerType != csa.streamType {
			msgCopy := msg
			msgCopy.Target = peerType
			select {
			case peer <- msgCopy:
			default:
				// Peer inbox full, drop message
			}
		}
	}
}

// detectEmergence checks for emergent patterns in accumulated data
func (csa *CognitiveStreamActor) detectEmergence() {
	// Simple emergence detection based on data accumulation
	if len(csa.state.AccumulatedData) > 10 {
		// Detect pattern (simplified)
		pattern := fmt.Sprintf("emergence_%s_%d", csa.streamType, csa.state.CurrentStep)
		strength := float64(len(csa.state.AccumulatedData)) / 20.0
		if strength > 1.0 {
			strength = 1.0
		}
		
		if csa.onEmergence != nil {
			csa.onEmergence(pattern, strength)
		}
		
		// Broadcast emergence to peers
		csa.broadcastToPeers(CognitiveMessage{
			Type:      MsgTypeEmergence,
			Source:    csa.streamType,
			Step:      csa.state.CurrentStep,
			Phase:     csa.state.CurrentPhase,
			Content:   pattern,
			Timestamp: time.Now(),
		})
		
		// Clear accumulated data after emergence
		csa.state.AccumulatedData = csa.state.AccumulatedData[:0]
	}
}

// updateProcessingLoad updates the processing load metric
func (csa *CognitiveStreamActor) updateProcessingLoad() {
	// Calculate load based on accumulated data and message frequency
	dataLoad := float64(len(csa.state.AccumulatedData)) / 100.0
	if dataLoad > 1.0 {
		dataLoad = 1.0
	}
	
	// Smooth load update
	csa.state.ProcessingLoad = csa.state.ProcessingLoad*0.9 + dataLoad*0.1
}

// GetState returns the current stream state
func (csa *CognitiveStreamActor) GetState() StreamState {
	csa.mu.RLock()
	defer csa.mu.RUnlock()
	return csa.state
}

// GetStreamType returns the stream type
func (csa *CognitiveStreamActor) GetStreamType() StreamType {
	return csa.streamType
}

// ContributeToGestalt returns the stream's contribution to the global gestalt
func (csa *CognitiveStreamActor) ContributeToGestalt() map[string]interface{} {
	csa.mu.RLock()
	defer csa.mu.RUnlock()
	
	return map[string]interface{}{
		"stream_type":     csa.streamType.String(),
		"current_step":    csa.state.CurrentStep,
		"current_phase":   csa.state.CurrentPhase.String(),
		"processing_load": csa.state.ProcessingLoad,
		"accumulated":     len(csa.state.AccumulatedData),
		"insights":        len(csa.state.Insights),
		"running":         csa.running,
	}
}

// ActorSupervisor manages the three cognitive stream actors
type ActorSupervisor struct {
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	
	// The three cognitive streams
	perception  *CognitiveStreamActor
	action      *CognitiveStreamActor
	simulation  *CognitiveStreamActor
	
	// Unified outbox for external communication
	outbox      chan CognitiveMessage
	
	// Running state
	running     bool
}

// NewActorSupervisor creates a new actor supervisor
func NewActorSupervisor() *ActorSupervisor {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create the three streams
	perception := NewCognitiveStreamActor(StreamTypePerception)
	action := NewCognitiveStreamActor(StreamTypeAction)
	simulation := NewCognitiveStreamActor(StreamTypeSimulation)
	
	// Wire up peer communication
	peers := map[StreamType]chan CognitiveMessage{
		StreamTypePerception: perception.GetInbox(),
		StreamTypeAction:     action.GetInbox(),
		StreamTypeSimulation: simulation.GetInbox(),
	}
	
	perception.SetPeers(peers)
	action.SetPeers(peers)
	simulation.SetPeers(peers)
	
	return &ActorSupervisor{
		ctx:        ctx,
		cancel:     cancel,
		perception: perception,
		action:     action,
		simulation: simulation,
		outbox:     make(chan CognitiveMessage, 100),
	}
}

// SetStreamCallbacks sets the callbacks for all streams
func (as *ActorSupervisor) SetStreamCallbacks(
	onPerceive func(data interface{}) interface{},
	onAct func(data interface{}) interface{},
	onSimulate func(data interface{}) interface{},
	onEmergence func(pattern string, strength float64),
) {
	as.perception.SetCallbacks(onPerceive, nil, nil, onEmergence)
	as.action.SetCallbacks(nil, onAct, nil, onEmergence)
	as.simulation.SetCallbacks(nil, nil, onSimulate, onEmergence)
}

// Start starts all cognitive stream actors
func (as *ActorSupervisor) Start() error {
	as.mu.Lock()
	if as.running {
		as.mu.Unlock()
		return fmt.Errorf("actor supervisor already running")
	}
	as.running = true
	as.mu.Unlock()
	
	// Start all streams
	if err := as.perception.Start(); err != nil {
		return err
	}
	if err := as.action.Start(); err != nil {
		return err
	}
	if err := as.simulation.Start(); err != nil {
		return err
	}
	
	fmt.Println("🎭 Actor Supervisor started - 3 concurrent cognitive streams active")
	return nil
}

// Stop stops all cognitive stream actors
func (as *ActorSupervisor) Stop() error {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	if !as.running {
		return fmt.Errorf("actor supervisor not running")
	}
	
	as.running = false
	as.cancel()
	
	// Stop all streams
	as.perception.Stop()
	as.action.Stop()
	as.simulation.Stop()
	
	fmt.Println("🎭 Actor Supervisor stopped")
	return nil
}

// SendToStream sends a message to a specific stream
func (as *ActorSupervisor) SendToStream(streamType StreamType, msg CognitiveMessage) {
	switch streamType {
	case StreamTypePerception:
		as.perception.GetInbox() <- msg
	case StreamTypeAction:
		as.action.GetInbox() <- msg
	case StreamTypeSimulation:
		as.simulation.GetInbox() <- msg
	}
}

// GetStreamStates returns the states of all streams
func (as *ActorSupervisor) GetStreamStates() map[StreamType]StreamState {
	return map[StreamType]StreamState{
		StreamTypePerception: as.perception.GetState(),
		StreamTypeAction:     as.action.GetState(),
		StreamTypeSimulation: as.simulation.GetState(),
	}
}

// ContributeToGestalt returns the supervisor's contribution to the global gestalt
func (as *ActorSupervisor) ContributeToGestalt() map[string]interface{} {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	return map[string]interface{}{
		"subsystem":   "actor_supervisor",
		"running":     as.running,
		"perception":  as.perception.ContributeToGestalt(),
		"action":      as.action.ContributeToGestalt(),
		"simulation":  as.simulation.ContributeToGestalt(),
	}
}
