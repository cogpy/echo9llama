package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EmbodiedCognition represents the embodied cognitive system.
// This is the central system that all operations flow through, wrapping the
// core Identity with processing pipeline, global state, and AI integration.
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

	// Active
	Active bool
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
			FlowState: "initializing",
		},
		Pipeline: &CognitivePipeline{
			Stages:  make([]PipelineStage, 0),
			Current: 0,
			History: make([]PipelineEvent, 0),
		},
		Active: false,
	}

	// Initialize default pipeline stages
	ec.initializePipeline()

	return ec
}

// initializePipeline sets up the default cognitive processing pipeline
func (ec *EmbodiedCognition) initializePipeline() {
	ec.Pipeline.Stages = []PipelineStage{
		{
			Name: "perceive",
			Process: func(input interface{}) (interface{}, error) {
				return ec.perceive(input), nil
			},
			Weight: 1.0,
		},
		{
			Name: "attend",
			Process: func(input interface{}) (interface{}, error) {
				return ec.attend(input), nil
			},
			Weight: 0.9,
		},
		{
			Name: "reason",
			Process: func(input interface{}) (interface{}, error) {
				return ec.reason(input), nil
			},
			Weight: 0.8,
		},
		{
			Name: "respond",
			Process: func(input interface{}) (interface{}, error) {
				return ec.respond(input), nil
			},
			Weight: 0.7,
		},
	}
}

// Process processes input through the embodied cognitive pipeline
func (ec *EmbodiedCognition) Process(ctx context.Context, input interface{}) (interface{}, error) {
	ec.mu.Lock()
	ec.Active = true
	ec.mu.Unlock()

	current := input
	for _, stage := range ec.Pipeline.Stages {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		start := time.Now()
		result, err := stage.Process(current)
		duration := time.Since(start)

		ec.Pipeline.History = append(ec.Pipeline.History, PipelineEvent{
			Stage:     stage.Name,
			Input:     current,
			Output:    result,
			Timestamp: start,
			Duration:  duration,
		})

		if err != nil {
			return nil, fmt.Errorf("pipeline stage '%s' failed: %w", stage.Name, err)
		}
		current = result
	}

	return current, nil
}

// perceive processes raw input into perceptual representation
func (ec *EmbodiedCognition) perceive(input interface{}) interface{} {
	ec.GlobalState.FlowState = "perceiving"
	return fmt.Sprintf("perceived: %v", input)
}

// attend focuses attention on relevant aspects
func (ec *EmbodiedCognition) attend(input interface{}) interface{} {
	ec.GlobalState.FlowState = "attending"
	return input
}

// reason applies cognitive reasoning
func (ec *EmbodiedCognition) reason(input interface{}) interface{} {
	ec.GlobalState.FlowState = "reasoning"
	return fmt.Sprintf("reasoned: %v", input)
}

// respond generates a response
func (ec *EmbodiedCognition) respond(input interface{}) interface{} {
	ec.GlobalState.FlowState = "responding"
	return fmt.Sprintf("response: %v", input)
}

// GetStatus returns comprehensive system status
func (ec *EmbodiedCognition) GetStatus() map[string]interface{} {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	return map[string]interface{}{
		"active":        ec.Active,
		"identity":      ec.Identity.GetStatus(),
		"contexts":      len(ec.Contexts),
		"awareness":     ec.GlobalState.Awareness,
		"energy":        ec.GlobalState.Energy,
		"synchrony":     ec.GlobalState.Synchrony,
		"flow_state":    ec.GlobalState.FlowState,
		"pipeline_stages": len(ec.Pipeline.Stages),
		"pipeline_events": len(ec.Pipeline.History),
	}
}
