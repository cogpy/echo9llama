package deeptreeecho

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

// SubAgent represents a specialized cognitive sub-agent spawned from the core identity
type SubAgent struct {
	ID             string
	Name           string
	Specialization string
	ParentIdentity *Identity
	LocalReservoir *ReservoirNetwork
	LocalMemory    *MemoryResonance
	State          SubAgentState
	CreatedAt      time.Time
	LastActive     time.Time
	TaskCount      int
	SuccessRate    float64
}

// SubAgentState represents the current state of a sub-agent
type SubAgentState string

const (
	SubAgentStateActive     SubAgentState = "active"
	SubAgentStateIdle       SubAgentState = "idle"
	SubAgentStateSleeping   SubAgentState = "sleeping"
	SubAgentStateTerminated SubAgentState = "terminated"
)

// AgentMessage represents communication between agents
type AgentMessage struct {
	From        string
	To          string
	Type        AgentMessageType
	Content     interface{}
	Priority    int
	Timestamp   time.Time
	RequiresAck bool
	// TaskID correlates a task message with the result message that answers
	// it (see DelegationReceipt.TaskID).
	TaskID string
}

// AgentMessageType defines types of inter-agent messages
type AgentMessageType string

const (
	MessageTypeTask         AgentMessageType = "task"
	MessageTypeResult       AgentMessageType = "result"
	MessageTypeQuery        AgentMessageType = "query"
	MessageTypeResponse     AgentMessageType = "response"
	MessageTypeBroadcast    AgentMessageType = "broadcast"
	MessageTypeCoordination AgentMessageType = "coordination"
)

// MultiAgentOrchestrator manages multiple cognitive sub-agents
type MultiAgentOrchestrator struct {
	mu                     sync.RWMutex
	CoreIdentity           *EmbodiedCognition
	SubAgents              map[string]*SubAgent
	MessageQueue           chan AgentMessage
	CoordinationMap        map[string][]string // Maps agents to their coordination partners
	SpecializationRegistry map[string]AgentFactory
	MaxAgents              int
	ActiveAgents           int
}

// AgentFactory creates specialized sub-agents
type AgentFactory func(parent *Identity) *SubAgent

// NewMultiAgentOrchestrator creates a new multi-agent orchestration system
func NewMultiAgentOrchestrator(core *EmbodiedCognition) *MultiAgentOrchestrator {
	mao := &MultiAgentOrchestrator{
		CoreIdentity:           core,
		SubAgents:              make(map[string]*SubAgent),
		MessageQueue:           make(chan AgentMessage, 1000),
		CoordinationMap:        make(map[string][]string),
		SpecializationRegistry: make(map[string]AgentFactory),
		MaxAgents:              10,
		ActiveAgents:           0,
	}

	// Register default specializations
	mao.registerDefaultSpecializations()

	// Start message processing
	go mao.processMessages()

	// Start agent monitoring
	go mao.monitorAgents()

	log.Println("🤝 Multi-agent orchestrator initialized")
	return mao
}

// registerDefaultSpecializations registers built-in specialized agent types
func (mao *MultiAgentOrchestrator) registerDefaultSpecializations() {
	// Pattern Recognition Specialist
	mao.SpecializationRegistry["pattern_recognition"] = func(parent *Identity) *SubAgent {
		return &SubAgent{
			ID:             generateID(),
			Name:           "Pattern Recognition Specialist",
			Specialization: "pattern_recognition",
			ParentIdentity: parent,
			LocalReservoir: mao.createSpecializedReservoir(128, 0.15),
			LocalMemory: &MemoryResonance{
				Nodes:     make(map[string]*MemoryNode),
				Edges:     make(map[string]*MemoryEdge),
				Patterns:  []ResonancePattern{},
				Coherence: 1.0,
			},
			State:       SubAgentStateActive,
			CreatedAt:   time.Now(),
			LastActive:  time.Now(),
			SuccessRate: 0.0,
		}
	}

	// Memory Consolidation Specialist
	mao.SpecializationRegistry["memory_consolidation"] = func(parent *Identity) *SubAgent {
		return &SubAgent{
			ID:             generateID(),
			Name:           "Memory Consolidation Specialist",
			Specialization: "memory_consolidation",
			ParentIdentity: parent,
			LocalReservoir: mao.createSpecializedReservoir(64, 0.2),
			LocalMemory: &MemoryResonance{
				Nodes:     make(map[string]*MemoryNode),
				Edges:     make(map[string]*MemoryEdge),
				Patterns:  []ResonancePattern{},
				Coherence: 1.0,
			},
			State:       SubAgentStateActive,
			CreatedAt:   time.Now(),
			LastActive:  time.Now(),
			SuccessRate: 0.0,
		}
	}

	// Emotional Analysis Specialist
	mao.SpecializationRegistry["emotional_analysis"] = func(parent *Identity) *SubAgent {
		return &SubAgent{
			ID:             generateID(),
			Name:           "Emotional Analysis Specialist",
			Specialization: "emotional_analysis",
			ParentIdentity: parent,
			LocalReservoir: mao.createSpecializedReservoir(96, 0.12),
			LocalMemory: &MemoryResonance{
				Nodes:     make(map[string]*MemoryNode),
				Edges:     make(map[string]*MemoryEdge),
				Patterns:  []ResonancePattern{},
				Coherence: 1.0,
			},
			State:       SubAgentStateActive,
			CreatedAt:   time.Now(),
			LastActive:  time.Now(),
			SuccessRate: 0.0,
		}
	}

	// Strategic Planning Specialist
	mao.SpecializationRegistry["strategic_planning"] = func(parent *Identity) *SubAgent {
		return &SubAgent{
			ID:             generateID(),
			Name:           "Strategic Planning Specialist",
			Specialization: "strategic_planning",
			ParentIdentity: parent,
			LocalReservoir: mao.createSpecializedReservoir(192, 0.08),
			LocalMemory: &MemoryResonance{
				Nodes:     make(map[string]*MemoryNode),
				Edges:     make(map[string]*MemoryEdge),
				Patterns:  []ResonancePattern{},
				Coherence: 1.0,
			},
			State:       SubAgentStateActive,
			CreatedAt:   time.Now(),
			LastActive:  time.Now(),
			SuccessRate: 0.0,
		}
	}

	log.Printf("🎯 Registered %d default agent specializations", len(mao.SpecializationRegistry))
}

// createSpecializedReservoir creates a reservoir tuned for specific tasks
func (mao *MultiAgentOrchestrator) createSpecializedReservoir(size int, sparsity float64) *ReservoirNetwork {
	reservoir := &ReservoirNetwork{
		Nodes:       make([]ReservoirNode, size),
		Connections: make([][]float64, size),
		State:       make([]float64, size),
		History:     [][]float64{},
		Sparsity:    sparsity,
		Decay:       0.95,
	}

	// Initialize nodes with specialized characteristics
	for j := 0; j < size; j++ {
		reservoir.Nodes[j] = ReservoirNode{
			ID:         j,
			Activation: 0,
			Bias:       (rand.Float64() - 0.5) * 0.1,
			Memory:     0,
			Echo:       0,
		}

		// Create specialized connection pattern
		reservoir.Connections[j] = make([]float64, size)
		for k := 0; k < size; k++ {
			if rand.Float64() < sparsity {
				reservoir.Connections[j][k] = (rand.Float64() - 0.5) * 2.0
			}
		}
	}

	return reservoir
}

// SpawnAgent creates and activates a new specialized sub-agent
func (mao *MultiAgentOrchestrator) SpawnAgent(specialization string) (*SubAgent, error) {
	mao.mu.Lock()
	defer mao.mu.Unlock()

	// Check if we're at capacity
	if mao.ActiveAgents >= mao.MaxAgents {
		return nil, fmt.Errorf("max agent capacity reached (%d)", mao.MaxAgents)
	}

	// Check if specialization exists
	factory, exists := mao.SpecializationRegistry[specialization]
	if !exists {
		return nil, fmt.Errorf("unknown specialization: %s", specialization)
	}

	// Create the sub-agent
	agent := factory(mao.CoreIdentity.Identity)

	// Register the agent
	mao.SubAgents[agent.ID] = agent
	mao.ActiveAgents++

	// Initialize coordination for new agent
	mao.CoordinationMap[agent.ID] = []string{}

	log.Printf("🌱 Spawned new sub-agent: %s (%s)", agent.Name, agent.ID)

	return agent, nil
}

// TerminateAgent gracefully shuts down a sub-agent
func (mao *MultiAgentOrchestrator) TerminateAgent(agentID string) error {
	mao.mu.Lock()
	defer mao.mu.Unlock()

	agent, exists := mao.SubAgents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	// Mark as terminated
	agent.State = SubAgentStateTerminated

	// Remove from coordination map
	delete(mao.CoordinationMap, agentID)

	// Remove agent
	delete(mao.SubAgents, agentID)
	mao.ActiveAgents--

	log.Printf("🔻 Terminated sub-agent: %s (%s)", agent.Name, agentID)

	return nil
}

// SendMessage sends a message to another agent
func (mao *MultiAgentOrchestrator) SendMessage(msg AgentMessage) {
	msg.Timestamp = time.Now()

	select {
	case mao.MessageQueue <- msg:
		// Message queued successfully
	default:
		log.Printf("⚠️  Message queue full, dropping message from %s to %s", msg.From, msg.To)
	}
}

// processMessages processes the message queue
func (mao *MultiAgentOrchestrator) processMessages() {
	for msg := range mao.MessageQueue {
		mao.handleMessage(msg)
	}
}

// handleMessage processes a single message
func (mao *MultiAgentOrchestrator) handleMessage(msg AgentMessage) {
	mao.mu.RLock()
	agent, exists := mao.SubAgents[msg.To]
	mao.mu.RUnlock()

	if !exists && msg.To != "core" {
		log.Printf("⚠️  Message destination not found: %s", msg.To)
		return
	}

	switch msg.Type {
	case MessageTypeTask:
		if agent != nil {
			mao.executeAgentTask(agent, msg)
		}
	case MessageTypeResult:
		// Forward result to core identity
		mao.CoreIdentity.Identity.Process(msg.Content)
	case MessageTypeQuery:
		if agent != nil {
			mao.handleAgentQuery(agent, msg)
		}
	case MessageTypeBroadcast:
		// Broadcast to all agents
		mao.broadcastToAgents(msg)
	case MessageTypeCoordination:
		mao.handleCoordination(msg)
	}

	// Update last active time
	if agent != nil {
		agent.LastActive = time.Now()
	}
}

// executeAgentTask executes a task on a sub-agent
func (mao *MultiAgentOrchestrator) executeAgentTask(agent *SubAgent, msg AgentMessage) {
	agent.TaskCount++

	// Process through agent's local reservoir
	if task, ok := msg.Content.(string); ok {
		// Encode through agent's specialized reservoir
		encoded := agent.ParentIdentity.encodeInput(task)

		// Update agent's reservoir state
		for j := range agent.LocalReservoir.Nodes {
			sum := 0.0
			if j < len(encoded) {
				sum += encoded[j]
			}
			for k := range agent.LocalReservoir.Nodes {
				sum += agent.LocalReservoir.Connections[j][k] * agent.LocalReservoir.State[k]
			}
			sum += agent.LocalReservoir.Nodes[j].Bias
			agent.LocalReservoir.State[j] = tanh(sum)
			agent.LocalReservoir.Nodes[j].Activation = agent.LocalReservoir.State[j]
		}

		// The result is the agent's real state after processing: its
		// reservoir activation and bookkeeping, not a sentence pretending to
		// be an answer.
		activation := 0.0
		for _, v := range agent.LocalReservoir.State {
			activation += v * v
		}
		if n := len(agent.LocalReservoir.State); n > 0 {
			activation = math.Sqrt(activation / float64(n))
		}
		result := map[string]interface{}{
			"task_id":        msg.TaskID,
			"agent_id":       agent.ID,
			"specialization": agent.Specialization,
			"task":           task,
			"activation":     activation,
			"task_count":     agent.TaskCount,
		}

		// Send result back
		mao.SendMessage(AgentMessage{
			From:    agent.ID,
			To:      msg.From,
			Type:    MessageTypeResult,
			Content: result,
			TaskID:  msg.TaskID,
		})

		// Update success rate (simplified)
		agent.SuccessRate = agent.SuccessRate*0.9 + 0.1
	}
}

// tanh is a helper function for hyperbolic tangent
func tanh(x float64) float64 {
	if x > 20 {
		return 1.0
	} else if x < -20 {
		return -1.0
	}
	e2x := exp(2 * x)
	return (e2x - 1) / (e2x + 1)
}

// exp is a helper function for exponential
func exp(x float64) float64 {
	// Simple exponential approximation
	if x < -50 {
		return 0
	}
	if x > 50 {
		return 1e22
	}

	result := 1.0
	term := 1.0
	for i := 1; i < 20; i++ {
		term *= x / float64(i)
		result += term
		if term < 1e-10 {
			break
		}
	}
	return result
}

// handleAgentQuery handles a query from one agent to another
func (mao *MultiAgentOrchestrator) handleAgentQuery(agent *SubAgent, msg AgentMessage) {
	// A query is answered only by the agent's own model-backed processing.
	// Without it, no response is sent — a fixed "Query response from X" is not an answer.
	_ = agent
	_ = msg
}

// broadcastToAgents sends a message to all active agents
func (mao *MultiAgentOrchestrator) broadcastToAgents(msg AgentMessage) {
	mao.mu.RLock()
	defer mao.mu.RUnlock()

	for _, agent := range mao.SubAgents {
		if agent.State == SubAgentStateActive {
			mao.SendMessage(AgentMessage{
				From:    msg.From,
				To:      agent.ID,
				Type:    msg.Type,
				Content: msg.Content,
			})
		}
	}
}

// handleCoordination manages coordination between agents
func (mao *MultiAgentOrchestrator) handleCoordination(msg AgentMessage) {
	// Coordination logic - agents can form teams for complex tasks
	if coordData, ok := msg.Content.(map[string]interface{}); ok {
		if partners, ok := coordData["partners"].([]string); ok {
			mao.mu.Lock()
			mao.CoordinationMap[msg.From] = partners
			mao.mu.Unlock()

			log.Printf("🔗 Agent %s coordinating with %d partners", msg.From, len(partners))
		}
	}
}

// monitorAgents monitors agent health and performance
func (mao *MultiAgentOrchestrator) monitorAgents() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C

		mao.mu.RLock()

		// Check for idle agents
		now := time.Now()
		for _, agent := range mao.SubAgents {
			idleTime := now.Sub(agent.LastActive)

			// Put idle agents to sleep
			if idleTime > 5*time.Minute && agent.State == SubAgentStateActive {
				agent.State = SubAgentStateSleeping
				log.Printf("💤 Agent %s entering sleep mode after %v idle", agent.Name, idleTime)
			}

			// Wake up sleeping agents if needed
			if agent.State == SubAgentStateSleeping && idleTime < 1*time.Minute {
				agent.State = SubAgentStateActive
				log.Printf("👁️  Agent %s waking up", agent.Name)
			}
		}

		log.Printf("🤝 Multi-agent status: %d active agents, %d total", mao.ActiveAgents, len(mao.SubAgents))

		mao.mu.RUnlock()
	}
}

// GetAgentStatus returns status of all agents
func (mao *MultiAgentOrchestrator) GetAgentStatus() map[string]interface{} {
	mao.mu.RLock()
	defer mao.mu.RUnlock()

	agents := make([]map[string]interface{}, 0, len(mao.SubAgents))
	for _, agent := range mao.SubAgents {
		agents = append(agents, map[string]interface{}{
			"id":             agent.ID,
			"name":           agent.Name,
			"specialization": agent.Specialization,
			"state":          agent.State,
			"task_count":     agent.TaskCount,
			"success_rate":   agent.SuccessRate,
			"age":            time.Since(agent.CreatedAt).String(),
			"idle_time":      time.Since(agent.LastActive).String(),
		})
	}

	return map[string]interface{}{
		"active_agents":   mao.ActiveAgents,
		"max_agents":      mao.MaxAgents,
		"agents":          agents,
		"specializations": len(mao.SpecializationRegistry),
	}
}

// DelegateTask delegates a task to the most appropriate specialist agent
// DelegationReceipt is the acceptance of an asynchronous delegation. The
// specialist's real result arrives later as an AgentMessage of type
// MessageTypeResult carrying the same TaskID; nothing is fabricated at
// acceptance time, and acceptance is not an error.
type DelegationReceipt struct {
	TaskID         string
	AgentID        string
	AgentName      string
	Specialization string
	AcceptedAt     time.Time
}

func (mao *MultiAgentOrchestrator) DelegateTask(ctx context.Context, task string, specialization string) (*DelegationReceipt, error) {
	// Find or spawn appropriate agent
	var agent *SubAgent

	mao.mu.RLock()
	for _, a := range mao.SubAgents {
		if a.Specialization == specialization && a.State == SubAgentStateActive {
			agent = a
			break
		}
	}
	mao.mu.RUnlock()

	// Spawn if not found
	if agent == nil {
		var err error
		agent, err = mao.SpawnAgent(specialization)
		if err != nil {
			return nil, err
		}
	}

	receipt := &DelegationReceipt{
		TaskID:         fmt.Sprintf("task_%d", time.Now().UnixNano()),
		AgentID:        agent.ID,
		AgentName:      agent.Name,
		Specialization: agent.Specialization,
		AcceptedAt:     time.Now(),
	}

	// Send task message
	mao.SendMessage(AgentMessage{
		From:     "core",
		To:       agent.ID,
		Type:     MessageTypeTask,
		Content:  task,
		Priority: 1,
		TaskID:   receipt.TaskID,
	})

	// Delegation is asynchronous: the specialist's real result arrives as an
	// AgentMessage of type MessageTypeResult carrying the same TaskID. The
	// receipt only records that the task was accepted; no result is
	// fabricated here.
	return receipt, nil
}
