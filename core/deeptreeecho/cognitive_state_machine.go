package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CognitiveStateMachine provides a finite state machine for managing cognitive states
// This is inspired by the Stateless FSM library for managing state transitions
type CognitiveStateMachine struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc

	// Current state
	currentState    CognitiveState
	stateArgs       []interface{}

	// State configuration
	stateConfigs    map[CognitiveState]*StateConfiguration

	// Trigger configuration
	triggerConfigs  map[CognitiveTrigger]*TriggerConfiguration

	// Event handlers
	onTransitioning []TransitionHandler
	onTransitioned  []TransitionHandler
	onUnhandled     UnhandledTriggerHandler

	// Transition history
	history         []*StateTransition
	historyLimit    int

	// Metrics
	totalTransitions uint64

	// Running state
	running         bool
}

// CognitiveState represents a cognitive state
type CognitiveState string

// CognitiveTrigger represents a trigger that causes state transitions
type CognitiveTrigger string

// Cognitive states for Deep Tree Echo
const (
	// Primary consciousness states
	StateIdle           CognitiveState = "idle"
	StateCogAwake       CognitiveState = "awake"
	StateCogDreaming    CognitiveState = "dreaming"
	StateCogResting     CognitiveState = "resting"

	// Cognitive processing states
	StatePerceiving     CognitiveState = "perceiving"
	StateThinking       CognitiveState = "thinking"
	StateActing         CognitiveState = "acting"
	StateSimulating     CognitiveState = "simulating"
	StateReflecting     CognitiveState = "reflecting"

	// Learning states
	StateLearning       CognitiveState = "learning"
	StatePracticing     CognitiveState = "practicing"
	StateIntegrating    CognitiveState = "integrating"

	// Social states
	StateListening      CognitiveState = "listening"
	StateSpeaking       CognitiveState = "speaking"
	StateConversing     CognitiveState = "conversing"

	// Meta states
	StateEmergent       CognitiveState = "emergent"
	StateWisdomSeeking  CognitiveState = "wisdom_seeking"
)

// Cognitive triggers
const (
	TriggerWake         CognitiveTrigger = "wake"
	TriggerSleep        CognitiveTrigger = "sleep"
	TriggerDream        CognitiveTrigger = "dream"
	TriggerRest         CognitiveTrigger = "rest"

	TriggerPerceive     CognitiveTrigger = "perceive"
	TriggerThink        CognitiveTrigger = "think"
	TriggerAct          CognitiveTrigger = "act"
	TriggerSimulate     CognitiveTrigger = "simulate"
	TriggerReflect      CognitiveTrigger = "reflect"

	TriggerLearn        CognitiveTrigger = "learn"
	TriggerPractice     CognitiveTrigger = "practice"
	TriggerIntegrate    CognitiveTrigger = "integrate"

	TriggerListen       CognitiveTrigger = "listen"
	TriggerSpeak        CognitiveTrigger = "speak"
	TriggerConverse     CognitiveTrigger = "converse"

	TriggerEmerge       CognitiveTrigger = "emerge"
	TriggerSeekWisdom   CognitiveTrigger = "seek_wisdom"
	TriggerComplete     CognitiveTrigger = "complete"
)

// StateConfiguration holds configuration for a state
type StateConfiguration struct {
	State           CognitiveState
	Substates       []CognitiveState
	Superstate      CognitiveState
	EntryActions    []StateAction
	ExitActions     []StateAction
	InternalActions map[CognitiveTrigger]StateAction
	Transitions     map[CognitiveTrigger]*TransitionConfig
	Guards          map[CognitiveTrigger][]GuardCondition
}

// TransitionConfig holds configuration for a transition
type TransitionConfig struct {
	Destination CognitiveState
	Actions     []TransitionAction
	Guards      []GuardCondition
}

// TriggerConfiguration holds configuration for a trigger
type TriggerConfiguration struct {
	Trigger     CognitiveTrigger
	Parameters  []string
}

// StateTransition represents a state transition
type StateTransition struct {
	ID          string
	Source      CognitiveState
	Destination CognitiveState
	Trigger     CognitiveTrigger
	Args        []interface{}
	Timestamp   time.Time
	Duration    time.Duration
}

// StateAction is a function executed on state entry/exit
type StateAction func(ctx context.Context, args ...interface{}) error

// TransitionAction is a function executed during transition
type TransitionAction func(ctx context.Context, transition *StateTransition) error

// GuardCondition is a function that determines if a transition is allowed
type GuardCondition func(ctx context.Context, args ...interface{}) (bool, string)

// TransitionHandler is called on state transitions
type TransitionHandler func(ctx context.Context, transition *StateTransition)

// UnhandledTriggerHandler is called when a trigger is not handled
type UnhandledTriggerHandler func(ctx context.Context, state CognitiveState, trigger CognitiveTrigger, unmetGuards []string) error

// NewCognitiveStateMachine creates a new cognitive state machine
func NewCognitiveStateMachine(initialState CognitiveState) *CognitiveStateMachine {
	ctx, cancel := context.WithCancel(context.Background())

	csm := &CognitiveStateMachine{
		ctx:             ctx,
		cancel:          cancel,
		currentState:    initialState,
		stateConfigs:    make(map[CognitiveState]*StateConfiguration),
		triggerConfigs:  make(map[CognitiveTrigger]*TriggerConfiguration),
		onTransitioning: make([]TransitionHandler, 0),
		onTransitioned:  make([]TransitionHandler, 0),
		history:         make([]*StateTransition, 0),
		historyLimit:    100,
	}

	// Set default unhandled trigger handler
	csm.onUnhandled = func(ctx context.Context, state CognitiveState, trigger CognitiveTrigger, unmetGuards []string) error {
		if len(unmetGuards) > 0 {
			return fmt.Errorf("trigger '%s' not allowed from state '%s': guards not met: %v", trigger, state, unmetGuards)
		}
		return fmt.Errorf("trigger '%s' not valid from state '%s'", trigger, state)
	}

	return csm
}

// Start begins the cognitive state machine
func (csm *CognitiveStateMachine) Start() error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if csm.running {
		return fmt.Errorf("cognitive state machine already running")
	}

	csm.running = true
	fmt.Println("🧠 Cognitive State Machine started")

	// Configure default states and transitions
	csm.configureDefaultStates()

	return nil
}

// Stop gracefully stops the cognitive state machine
func (csm *CognitiveStateMachine) Stop() error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if !csm.running {
		return fmt.Errorf("cognitive state machine not running")
	}

	csm.cancel()
	csm.running = false
	fmt.Println("🧠 Cognitive State Machine stopped")

	return nil
}

// configureDefaultStates sets up the default state configuration
func (csm *CognitiveStateMachine) configureDefaultStates() {
	// Configure Idle state
	csm.Configure(StateIdle).
		Permit(TriggerWake, StateCogAwake).
		Permit(TriggerDream, StateCogDreaming)

	// Configure Awake state
	csm.Configure(StateCogAwake).
		Permit(TriggerSleep, StateIdle).
		Permit(TriggerRest, StateCogResting).
		Permit(TriggerPerceive, StatePerceiving).
		Permit(TriggerThink, StateThinking).
		Permit(TriggerLearn, StateLearning).
		Permit(TriggerListen, StateListening).
		Permit(TriggerSeekWisdom, StateWisdomSeeking)

	// Configure Dreaming state
	csm.Configure(StateCogDreaming).
		Permit(TriggerWake, StateCogAwake).
		Permit(TriggerIntegrate, StateIntegrating)

	// Configure Resting state
	csm.Configure(StateCogResting).
		Permit(TriggerWake, StateCogAwake).
		Permit(TriggerSleep, StateIdle)

	// Configure Perceiving state
	csm.Configure(StatePerceiving).
		Permit(TriggerThink, StateThinking).
		Permit(TriggerAct, StateActing).
		Permit(TriggerComplete, StateCogAwake)

	// Configure Thinking state
	csm.Configure(StateThinking).
		Permit(TriggerAct, StateActing).
		Permit(TriggerSimulate, StateSimulating).
		Permit(TriggerReflect, StateReflecting).
		Permit(TriggerComplete, StateCogAwake)

	// Configure Acting state
	csm.Configure(StateActing).
		Permit(TriggerPerceive, StatePerceiving).
		Permit(TriggerReflect, StateReflecting).
		Permit(TriggerComplete, StateCogAwake)

	// Configure Simulating state
	csm.Configure(StateSimulating).
		Permit(TriggerThink, StateThinking).
		Permit(TriggerReflect, StateReflecting).
		Permit(TriggerEmerge, StateEmergent).
		Permit(TriggerComplete, StateCogAwake)

	// Configure Reflecting state
	csm.Configure(StateReflecting).
		Permit(TriggerThink, StateThinking).
		Permit(TriggerLearn, StateLearning).
		Permit(TriggerSeekWisdom, StateWisdomSeeking).
		Permit(TriggerComplete, StateCogAwake)

	// Configure Learning state
	csm.Configure(StateLearning).
		Permit(TriggerPractice, StatePracticing).
		Permit(TriggerIntegrate, StateIntegrating).
		Permit(TriggerComplete, StateCogAwake)

	// Configure Practicing state
	csm.Configure(StatePracticing).
		Permit(TriggerLearn, StateLearning).
		Permit(TriggerComplete, StateCogAwake)

	// Configure Integrating state
	csm.Configure(StateIntegrating).
		Permit(TriggerComplete, StateCogAwake).
		Permit(TriggerSeekWisdom, StateWisdomSeeking)

	// Configure Listening state
	csm.Configure(StateListening).
		Permit(TriggerSpeak, StateSpeaking).
		Permit(TriggerConverse, StateConversing).
		Permit(TriggerComplete, StateCogAwake)

	// Configure Speaking state
	csm.Configure(StateSpeaking).
		Permit(TriggerListen, StateListening).
		Permit(TriggerConverse, StateConversing).
		Permit(TriggerComplete, StateCogAwake)

	// Configure Conversing state
	csm.Configure(StateConversing).
		Permit(TriggerListen, StateListening).
		Permit(TriggerSpeak, StateSpeaking).
		Permit(TriggerComplete, StateCogAwake)

	// Configure Emergent state
	csm.Configure(StateEmergent).
		Permit(TriggerIntegrate, StateIntegrating).
		Permit(TriggerSeekWisdom, StateWisdomSeeking).
		Permit(TriggerComplete, StateCogAwake)

	// Configure WisdomSeeking state
	csm.Configure(StateWisdomSeeking).
		Permit(TriggerReflect, StateReflecting).
		Permit(TriggerIntegrate, StateIntegrating).
		Permit(TriggerComplete, StateCogAwake)

	fmt.Printf("   Configured %d cognitive states\n", len(csm.stateConfigs))
}

// StateConfigBuilder provides a fluent API for configuring states
type StateConfigBuilder struct {
	csm    *CognitiveStateMachine
	config *StateConfiguration
}

// Configure starts configuring a state
func (csm *CognitiveStateMachine) Configure(state CognitiveState) *StateConfigBuilder {
	config, exists := csm.stateConfigs[state]
	if !exists {
		config = &StateConfiguration{
			State:           state,
			Substates:       make([]CognitiveState, 0),
			EntryActions:    make([]StateAction, 0),
			ExitActions:     make([]StateAction, 0),
			InternalActions: make(map[CognitiveTrigger]StateAction),
			Transitions:     make(map[CognitiveTrigger]*TransitionConfig),
			Guards:          make(map[CognitiveTrigger][]GuardCondition),
		}
		csm.stateConfigs[state] = config
	}

	return &StateConfigBuilder{
		csm:    csm,
		config: config,
	}
}

// Permit allows a transition from this state to another
func (scb *StateConfigBuilder) Permit(trigger CognitiveTrigger, destination CognitiveState) *StateConfigBuilder {
	scb.config.Transitions[trigger] = &TransitionConfig{
		Destination: destination,
		Actions:     make([]TransitionAction, 0),
		Guards:      make([]GuardCondition, 0),
	}
	return scb
}

// PermitIf allows a transition with a guard condition
func (scb *StateConfigBuilder) PermitIf(trigger CognitiveTrigger, destination CognitiveState, guard GuardCondition) *StateConfigBuilder {
	scb.config.Transitions[trigger] = &TransitionConfig{
		Destination: destination,
		Actions:     make([]TransitionAction, 0),
		Guards:      []GuardCondition{guard},
	}
	return scb
}

// OnEntry adds an entry action
func (scb *StateConfigBuilder) OnEntry(action StateAction) *StateConfigBuilder {
	scb.config.EntryActions = append(scb.config.EntryActions, action)
	return scb
}

// OnExit adds an exit action
func (scb *StateConfigBuilder) OnExit(action StateAction) *StateConfigBuilder {
	scb.config.ExitActions = append(scb.config.ExitActions, action)
	return scb
}

// SubstateOf sets the superstate
func (scb *StateConfigBuilder) SubstateOf(superstate CognitiveState) *StateConfigBuilder {
	scb.config.Superstate = superstate

	// Add this state to superstate's substates
	if superConfig, exists := scb.csm.stateConfigs[superstate]; exists {
		superConfig.Substates = append(superConfig.Substates, scb.config.State)
	}

	return scb
}

// Fire triggers a state transition
func (csm *CognitiveStateMachine) Fire(trigger CognitiveTrigger, args ...interface{}) error {
	return csm.FireCtx(csm.ctx, trigger, args...)
}

// FireCtx triggers a state transition with context
func (csm *CognitiveStateMachine) FireCtx(ctx context.Context, trigger CognitiveTrigger, args ...interface{}) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	startTime := time.Now()

	// Get current state config
	config, exists := csm.stateConfigs[csm.currentState]
	if !exists {
		return fmt.Errorf("no configuration for state: %s", csm.currentState)
	}

	// Check if transition is allowed
	transConfig, exists := config.Transitions[trigger]
	if !exists {
		return csm.onUnhandled(ctx, csm.currentState, trigger, nil)
	}

	// Check guards
	unmetGuards := make([]string, 0)
	for _, guard := range transConfig.Guards {
		allowed, reason := guard(ctx, args...)
		if !allowed {
			unmetGuards = append(unmetGuards, reason)
		}
	}

	if len(unmetGuards) > 0 {
		return csm.onUnhandled(ctx, csm.currentState, trigger, unmetGuards)
	}

	// Create transition record
	transition := &StateTransition{
		ID:          fmt.Sprintf("trans_%s", uuid.New().String()[:8]),
		Source:      csm.currentState,
		Destination: transConfig.Destination,
		Trigger:     trigger,
		Args:        args,
		Timestamp:   time.Now(),
	}

	// Call transitioning handlers
	for _, handler := range csm.onTransitioning {
		handler(ctx, transition)
	}

	// Execute exit actions
	for _, action := range config.ExitActions {
		if err := action(ctx, args...); err != nil {
			return fmt.Errorf("exit action failed: %w", err)
		}
	}

	// Execute transition actions
	for _, action := range transConfig.Actions {
		if err := action(ctx, transition); err != nil {
			return fmt.Errorf("transition action failed: %w", err)
		}
	}

	// Update state
	csm.currentState = transConfig.Destination
	csm.stateArgs = args

	// Execute entry actions
	if destConfig, exists := csm.stateConfigs[transConfig.Destination]; exists {
		for _, action := range destConfig.EntryActions {
			if err := action(ctx, args...); err != nil {
				return fmt.Errorf("entry action failed: %w", err)
			}
		}
	}

	// Record transition
	transition.Duration = time.Since(startTime)
	csm.history = append(csm.history, transition)
	if len(csm.history) > csm.historyLimit {
		csm.history = csm.history[1:]
	}
	csm.totalTransitions++

	// Call transitioned handlers
	for _, handler := range csm.onTransitioned {
		handler(ctx, transition)
	}

	return nil
}

// State returns the current state
func (csm *CognitiveStateMachine) State() CognitiveState {
	csm.mu.RLock()
	defer csm.mu.RUnlock()
	return csm.currentState
}

// CanFire checks if a trigger can be fired from the current state
func (csm *CognitiveStateMachine) CanFire(trigger CognitiveTrigger) bool {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	config, exists := csm.stateConfigs[csm.currentState]
	if !exists {
		return false
	}

	_, exists = config.Transitions[trigger]
	return exists
}

// PermittedTriggers returns all triggers that can be fired from the current state
func (csm *CognitiveStateMachine) PermittedTriggers() []CognitiveTrigger {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	config, exists := csm.stateConfigs[csm.currentState]
	if !exists {
		return nil
	}

	triggers := make([]CognitiveTrigger, 0, len(config.Transitions))
	for trigger := range config.Transitions {
		triggers = append(triggers, trigger)
	}
	return triggers
}

// OnTransitioning adds a handler called before transitions
func (csm *CognitiveStateMachine) OnTransitioning(handler TransitionHandler) {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	csm.onTransitioning = append(csm.onTransitioning, handler)
}

// OnTransitioned adds a handler called after transitions
func (csm *CognitiveStateMachine) OnTransitioned(handler TransitionHandler) {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	csm.onTransitioned = append(csm.onTransitioned, handler)
}

// GetHistory returns the transition history
func (csm *CognitiveStateMachine) GetHistory() []*StateTransition {
	csm.mu.RLock()
	defer csm.mu.RUnlock()
	return csm.history
}

// GetMetrics returns state machine metrics
func (csm *CognitiveStateMachine) GetMetrics() map[string]interface{} {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	return map[string]interface{}{
		"running":           csm.running,
		"current_state":     csm.currentState,
		"total_states":      len(csm.stateConfigs),
		"total_transitions": csm.totalTransitions,
		"history_size":      len(csm.history),
	}
}

// ContributeToGestalt provides state machine state for the global gestalt
func (csm *CognitiveStateMachine) ContributeToGestalt() map[string]interface{} {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	var lastTransition *StateTransition
	if len(csm.history) > 0 {
		lastTransition = csm.history[len(csm.history)-1]
	}

	gestalt := map[string]interface{}{
		"running":           csm.running,
		"current_state":     csm.currentState,
		"permitted_triggers": csm.PermittedTriggers(),
		"total_transitions": csm.totalTransitions,
	}

	if lastTransition != nil {
		gestalt["last_transition"] = map[string]interface{}{
			"source":      lastTransition.Source,
			"destination": lastTransition.Destination,
			"trigger":     lastTransition.Trigger,
			"timestamp":   lastTransition.Timestamp,
		}
	}

	return gestalt
}

// GenerateDOT generates a DOT graph representation of the state machine
func (csm *CognitiveStateMachine) GenerateDOT() string {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	dot := "digraph CognitiveStateMachine {\n"
	dot += "    rankdir=LR;\n"
	dot += "    node [shape=ellipse];\n\n"

	// Mark current state
	dot += fmt.Sprintf("    \"%s\" [style=filled, fillcolor=lightblue];\n\n", csm.currentState)

	// Add transitions
	for state, config := range csm.stateConfigs {
		for trigger, transConfig := range config.Transitions {
			dot += fmt.Sprintf("    \"%s\" -> \"%s\" [label=\"%s\"];\n", state, transConfig.Destination, trigger)
		}
	}

	dot += "}\n"
	return dot
}
