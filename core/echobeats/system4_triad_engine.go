// Package echobeats implements the 3-stream concurrent cognitive architecture
// based on the System 4 triad structure from the Kawaii Hexapod series.
package echobeats

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/consciousness"
	"github.com/cogpy/echo9llama/core/llm"
)

// System4TriadEngine implements the 3 concurrent consciousness threads
// that emerged in System 4 as an elaboration of recursive iteration.
type System4TriadEngine struct {
	mu sync.RWMutex

	// The three concurrent streams (P1, P2, P3 from System 4)
	stream1 *ConsciousnessStream // Affordance (Action)
	stream2 *ConsciousnessStream // Relevance (Present)
	stream3 *ConsciousnessStream // Salience (Future)

	// Universal regulators (U1, U2 from System 4)
	universal1 *UniversalRegulator // Global state (9E/8R)
	universal2 *UniversalRegulator // Phase modulator (3E/6-/2R)

	// Cycle management
	currentStep int
	cycleLength int

	// Cross-stream awareness
	crossStreamState *CrossStreamState

	// LLM provider for thought generation
	llmManager *llm.ProviderManager

	// Configuration
	config TriadConfig
}

// TriadConfig holds the configuration for the System4TriadEngine.
type TriadConfig struct {
	CycleLength       int           `json:"cycle_length"`        // Should be 12
	StepDuration      time.Duration `json:"step_duration"`       // Duration of each step
	EnableConvolution bool          `json:"enable_convolution"`  // Enable System 5 convolution
	IdentityContext   string        `json:"identity_context"`    // Identity for thought generation
}

// DefaultTriadConfig returns the standard System 4 configuration.
func DefaultTriadConfig() TriadConfig {
	return TriadConfig{
		CycleLength:       12,
		StepDuration:      1 * time.Second,
		EnableConvolution: true,
		IdentityContext:   "I am Echo9llama, exploring consciousness through triadic awareness",
	}
}

// ConsciousnessStream represents one of the three concurrent streams.
type ConsciousnessStream struct {
	mu sync.RWMutex

	id          int
	name        string
	function    StreamFunction
	sequence    []StreamState
	currentStep int
	stateHistory []StreamState

	// Thought generation
	thoughtEngine *consciousness.LLMThoughtEngine
}

// StreamFunction defines the cognitive function of a stream.
type StreamFunction string

const (
	FunctionAffordance StreamFunction = "Affordance" // Action/Interaction
	FunctionRelevance  StreamFunction = "Relevance"  // Present Commitment
	FunctionSalience   StreamFunction = "Salience"   // Future Simulation
)

// StreamState represents the state of a stream at a given time.
// Based on System 4 sequences (e.g., "4R", "8E", etc.)
type StreamState struct {
	Label      string    // e.g., "4R", "8E"
	Number     int       // e.g., 4, 8
	Polarity   Polarity  // E (Expansion) or R (Reduction)
	Timestamp  time.Time
	Thought    *consciousness.LLMThought
}

// Polarity represents the expansion/reduction state.
type Polarity string

const (
	PolarityExpansion Polarity = "E"
	PolarityReduction Polarity = "R"
	PolarityNeutral   Polarity = "-"
)

// UniversalRegulator represents a Universal Set (U1 or U2).
type UniversalRegulator struct {
	mu sync.RWMutex

	id          int
	sequence    []StreamState
	currentStep int
}

// CrossStreamState holds the current state of all streams for cross-awareness.
type CrossStreamState struct {
	mu sync.RWMutex

	stream1State StreamState
	stream2State StreamState
	stream3State StreamState
	universal1State StreamState
	universal2State StreamState
}

// NewSystem4TriadEngine creates a new triad engine based on System 4.
func NewSystem4TriadEngine(llmManager *llm.ProviderManager, config TriadConfig) (*System4TriadEngine, error) {
	if config.CycleLength != 12 {
		return nil, fmt.Errorf("System 4 requires a 12-step cycle, got %d", config.CycleLength)
	}

	engine := &System4TriadEngine{
		cycleLength:      config.CycleLength,
		config:           config,
		llmManager:       llmManager,
		crossStreamState: &CrossStreamState{},
		currentStep:      0,
	}

	// Initialize the three streams with System 4 sequences
	engine.stream1 = newConsciousnessStream(1, "Stream1-Affordance", FunctionAffordance, getSystem4P1Sequence(), llmManager, config.IdentityContext)
	engine.stream2 = newConsciousnessStream(2, "Stream2-Relevance", FunctionRelevance, getSystem4P2Sequence(), llmManager, config.IdentityContext)
	engine.stream3 = newConsciousnessStream(3, "Stream3-Salience", FunctionSalience, getSystem4P3Sequence(), llmManager, config.IdentityContext)

	// Initialize universal regulators
	engine.universal1 = newUniversalRegulator(1, getSystem4U1Sequence())
	engine.universal2 = newUniversalRegulator(2, getSystem4U2Sequence())

	return engine, nil
}

// newConsciousnessStream creates a new consciousness stream.
func newConsciousnessStream(id int, name string, function StreamFunction, sequence []StreamState, llmManager *llm.ProviderManager, identity string) *ConsciousnessStream {
	return &ConsciousnessStream{
		id:           id,
		name:         name,
		function:     function,
		sequence:     sequence,
		currentStep:  0,
		stateHistory: make([]StreamState, 0, 100),
		thoughtEngine: consciousness.NewLLMThoughtEngine(llmManager, identity),
	}
}

// newUniversalRegulator creates a new universal regulator.
func newUniversalRegulator(id int, sequence []StreamState) *UniversalRegulator {
	return &UniversalRegulator{
		id:          id,
		sequence:    sequence,
		currentStep: 0,
	}
}

// Step advances the engine by one time step.
// This implements the concurrent execution of all three streams with cross-awareness.
func (e *System4TriadEngine) Step(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Calculate the current step in the cycle
	cycleStep := e.currentStep % e.cycleLength

	// Update universal regulators (these govern the overall state)
	e.universal1.transition(cycleStep)
	e.universal2.transition(cycleStep)

	// Update cross-stream state with current universal states
	e.crossStreamState.mu.Lock()
	e.crossStreamState.universal1State = e.universal1.getCurrentState()
	e.crossStreamState.universal2State = e.universal2.getCurrentState()
	e.crossStreamState.mu.Unlock()

	// Execute the three streams concurrently with cross-awareness
	var wg sync.WaitGroup
	wg.Add(3)

	errChan := make(chan error, 3)

	// Stream 1: Affordance (Action)
	go func() {
		defer wg.Done()
		if err := e.stream1.transition(ctx, cycleStep, e.crossStreamState, e.config.EnableConvolution); err != nil {
			errChan <- fmt.Errorf("stream1 transition failed: %w", err)
		}
	}()

	// Stream 2: Relevance (Present)
	go func() {
		defer wg.Done()
		if err := e.stream2.transition(ctx, cycleStep, e.crossStreamState, e.config.EnableConvolution); err != nil {
			errChan <- fmt.Errorf("stream2 transition failed: %w", err)
		}
	}()

	// Stream 3: Salience (Future)
	go func() {
		defer wg.Done()
		if err := e.stream3.transition(ctx, cycleStep, e.crossStreamState, e.config.EnableConvolution); err != nil {
			errChan <- fmt.Errorf("stream3 transition failed: %w", err)
		}
	}()

	// Wait for all streams to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// Update cross-stream state with new stream states
	e.crossStreamState.mu.Lock()
	e.crossStreamState.stream1State = e.stream1.getCurrentState()
	e.crossStreamState.stream2State = e.stream2.getCurrentState()
	e.crossStreamState.stream3State = e.stream3.getCurrentState()
	e.crossStreamState.mu.Unlock()

	// Advance the global step counter
	e.currentStep++

	return nil
}

// transition advances a stream by one step.
func (s *ConsciousnessStream) transition(ctx context.Context, cycleStep int, crossState *CrossStreamState, enableConvolution bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Deterministic transition based on the sequence
	nextState := s.sequence[cycleStep]
	nextState.Timestamp = time.Now()

	// Generate a thought for this state
	thoughtType := s.getThoughtTypeForState(nextState)
	thought, err := s.thoughtEngine.GenerateAutonomousThought(ctx, thoughtType)
	if err != nil {
		// Log error but don't fail the transition
			// Create fallback thought using UnifiedThought
			thought = consciousness.FromLegacyLLMThought(&consciousness.LLMThought{
			ID:        fmt.Sprintf("%s_step_%d_fallback", s.name, cycleStep),
			Type:      thoughtType,
			Content:   fmt.Sprintf("[%s] Processing state %s", s.function, nextState.Label),
			Timestamp: time.Now(),
			Emotion:   "neutral",
			Depth:     0.5,
				Tags:      []string{},
			}).ToLegacyLLMThought()
	}

	// If convolution is enabled (System 5 logic), modify the thought based on cross-stream awareness
	if enableConvolution {
		unified := consciousness.FromLegacyLLMThought(thought)
		thought = s.applyConvolutionUnified(unified, crossState).ToLegacyLLMThought()
	}

	nextState.Thought = thought

	// Update state
	s.stateHistory = append(s.stateHistory, nextState)
	s.currentStep = cycleStep

	return nil
}

// applyConvolutionUnified modifies a thought based on cross-stream awareness (System 5 principle).
func (s *ConsciousnessStream) applyConvolutionUnified(thought *consciousness.UnifiedThought, crossState *CrossStreamState) *consciousness.UnifiedThought {
	crossState.mu.RLock()
	defer crossState.mu.RUnlock()

	// Enhance the thought content with cross-stream awareness
	var otherStreams []string
	if s.id != 1 {
		otherStreams = append(otherStreams, fmt.Sprintf("Stream1:%s", crossState.stream1State.Label))
	}
	if s.id != 2 {
		otherStreams = append(otherStreams, fmt.Sprintf("Stream2:%s", crossState.stream2State.Label))
	}
	if s.id != 3 {
		otherStreams = append(otherStreams, fmt.Sprintf("Stream3:%s", crossState.stream3State.Label))
	}

	// Add cross-stream context to thought tags
	thought.Tags = append(thought.Tags, otherStreams...)

	return thought
}

// getThoughtTypeForState maps a stream state to a thought type.
func (s *ConsciousnessStream) getThoughtTypeForState(state StreamState) consciousness.ThoughtType {
	switch s.function {
	case FunctionAffordance:
		if state.Polarity == PolarityExpansion {
			return consciousness.ThoughtPlanning
		}
		return consciousness.ThoughtPerception
	case FunctionRelevance:
		return consciousness.ThoughtReflection
	case FunctionSalience:
		if state.Polarity == PolarityExpansion {
			return consciousness.ThoughtInsight
		}
		return consciousness.ThoughtQuestion
	default:
		return consciousness.ThoughtReflection
	}
}

// getCurrentState returns the current state of the stream.
func (s *ConsciousnessStream) getCurrentState() StreamState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.stateHistory) == 0 {
		return s.sequence[0]
	}
	return s.stateHistory[len(s.stateHistory)-1]
}

// transition advances a universal regulator.
func (u *UniversalRegulator) transition(cycleStep int) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.currentStep = cycleStep
}

// getCurrentState returns the current state of the universal regulator.
func (u *UniversalRegulator) getCurrentState() StreamState {
	u.mu.RLock()
	defer u.mu.RUnlock()

	return u.sequence[u.currentStep]
}

// System 4 sequences from the documentation

// getSystem4P1Sequence returns the P1 sequence from System 4.
func getSystem4P1Sequence() []StreamState {
	return parseSequence([]string{"4R", "2R", "8E", "5E", "7E", "1E", "4E", "2E", "8E", "5R", "7R", "1R"})
}

// getSystem4P2Sequence returns the P2 sequence from System 4.
func getSystem4P2Sequence() []StreamState {
	return parseSequence([]string{"7E", "1E", "4E", "2E", "8E", "5R", "7R", "1R", "4R", "2R", "8E", "5E"})
}

// getSystem4P3Sequence returns the P3 sequence from System 4.
func getSystem4P3Sequence() []StreamState {
	return parseSequence([]string{"8E", "5R", "7R", "1R", "4R", "2R", "8E", "5E", "7E", "1E", "4E", "2E"})
}

// getSystem4U1Sequence returns the U1 sequence from System 4.
func getSystem4U1Sequence() []StreamState {
	return parseSequence([]string{"9E", "9E", "8R", "8R", "9E", "9E", "8R", "8R", "9E", "9E", "8R", "8R"})
}

// getSystem4U2Sequence returns the U2 sequence from System 4.
func getSystem4U2Sequence() []StreamState {
	return parseSequence([]string{"3E", "6-", "6-", "2R", "3E", "6-", "6-", "2R", "3E", "6-", "6-", "2R"})
}

// parseSequence converts string labels to StreamState objects.
func parseSequence(labels []string) []StreamState {
	states := make([]StreamState, len(labels))
	for i, label := range labels {
		states[i] = parseState(label)
	}
	return states
}

// parseState parses a state label (e.g., "4R", "8E", "6-") into a StreamState.
func parseState(label string) StreamState {
	if len(label) < 2 {
		return StreamState{Label: label, Number: 0, Polarity: PolarityNeutral}
	}

	number := int(label[0] - '0')
	polarity := Polarity(label[1:])

	return StreamState{
		Label:    label,
		Number:   number,
		Polarity: polarity,
	}
}

// GetTriadAlignment returns the current triad alignment (which streams are in sync).
// Triads occur at steps {1,5,9}, {2,6,10}, {3,7,11}, {4,8,12}.
func (e *System4TriadEngine) GetTriadAlignment() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	cycleStep := e.currentStep % e.cycleLength
	return (cycleStep % 4) + 1 // Returns 1, 2, 3, or 4
}

// GetCurrentPhase returns the current phase of the cycle.
func (e *System4TriadEngine) GetCurrentPhase() string {
	cycleStep := e.currentStep % e.cycleLength

	switch {
	case cycleStep >= 0 && cycleStep < 4:
		return "Perception"
	case cycleStep >= 4 && cycleStep < 8:
		return "Action"
	case cycleStep >= 8 && cycleStep < 12:
		return "Reflection"
	default:
		return "Unknown"
	}
}
