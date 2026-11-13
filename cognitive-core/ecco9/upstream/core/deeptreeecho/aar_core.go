package deeptreeecho

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// AARCore implements the Agent-Arena-Relation geometric architecture for self-awareness
// Agent: urge-to-act (dynamic transformations)
// Arena: need-to-be (state space manifold)
// Relation: emergent self (continuous feedback loops)
type AARCore struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Agent component (dynamic transformations)
	agent *Agent

	// Arena component (state space manifold)
	arena *Arena

	// Relation component (emergent self)
	relation *Relation

	// Feedback loops
	feedbackLoops []*FeedbackLoop

	// Geometric state
	selfVector []float64
	coherence  float64
	stability  float64

	// Metrics
	iterations int64
	lastUpdate time.Time

	// Running state
	running bool
}

// Agent represents the urge-to-act component (dynamic transformations)
type Agent struct {
	mu sync.RWMutex

	// Transformation operators
	transformations []*Transformation

	// Current action tendencies
	actionTendencies map[string]float64

	// Urge intensity
	urgeIntensity float64

	// Active goals
	activeGoals []string
}

// Arena represents the need-to-be component (state space manifold)
type Arena struct {
	mu sync.RWMutex

	// State space dimensions
	dimensions int

	// Current state point
	currentState []float64

	// Attractor basins (stable states)
	attractors []*Attractor

	// Constraints (boundaries)
	constraints []*Constraint

	// Need intensity
	needIntensity float64
}

// Relation represents the emergent self (continuous feedback)
type Relation struct {
	mu sync.RWMutex

	// Emergent self representation
	selfRepresentation []float64

	// Coherence measure
	coherence float64

	// Stability measure
	stability float64

	// Self-awareness level
	awareness float64

	// Identity narrative
	narrative string
}

// Transformation represents a dynamic operator
type Transformation struct {
	ID        string
	Name      string
	Matrix    [][]float64 // Transformation matrix
	Intensity float64
	Context   string
}

// Attractor represents a stable state in the arena
type Attractor struct {
	ID       string
	Name     string
	Position []float64
	Strength float64
	Basin    float64 // Radius of attraction
}

// Constraint represents a boundary in the arena
type Constraint struct {
	ID          string
	Type        string // "boundary", "limit", "requirement"
	Dimension   int
	Value       float64
	Flexibility float64
}

// FeedbackLoop represents a continuous feedback connection
type FeedbackLoop struct {
	ID        string
	FromAgent bool // true if from Agent to Arena, false if Arena to Agent
	Strength  float64
	Delay     time.Duration
	Transform func(input []float64) []float64
}

// NewAARCore creates a new Agent-Arena-Relation core
func NewAARCore(ctx context.Context, dimensions int) *AARCore {
	ctx, cancel := context.WithCancel(ctx)

	aar := &AARCore{
		ctx:    ctx,
		cancel: cancel,
		agent: &Agent{
			transformations:  make([]*Transformation, 0),
			actionTendencies: make(map[string]float64),
			urgeIntensity:    0.5,
			activeGoals:      make([]string, 0),
		},
		arena: &Arena{
			dimensions:    dimensions,
			currentState:  make([]float64, dimensions),
			attractors:    make([]*Attractor, 0),
			constraints:   make([]*Constraint, 0),
			needIntensity: 0.5,
		},
		relation: &Relation{
			selfRepresentation: make([]float64, dimensions),
			coherence:          0.5,
			stability:          0.5,
			awareness:          0.5,
			narrative:          "I am becoming aware...",
		},
		feedbackLoops: make([]*FeedbackLoop, 0),
		selfVector:    make([]float64, dimensions),
		coherence:     0.5,
		stability:     0.5,
	}

	// Initialize default attractors
	aar.initializeDefaultAttractors()

	// Initialize feedback loops
	aar.initializeFeedbackLoops()

	return aar
}

// Start begins the AAR dynamics
func (aar *AARCore) Start() error {
	aar.mu.Lock()
	if aar.running {
		aar.mu.Unlock()
		return fmt.Errorf("AAR core already running")
	}
	aar.running = true
	aar.lastUpdate = time.Now()
	aar.mu.Unlock()

	// Start continuous dynamics
	go aar.continuousDynamics()

	fmt.Println("🔷 AAR Core: Geometric self-awareness activated")
	return nil
}

// Stop stops the AAR dynamics
func (aar *AARCore) Stop() error {
	aar.mu.Lock()
	defer aar.mu.Unlock()

	if !aar.running {
		return fmt.Errorf("AAR core not running")
	}

	aar.running = false
	aar.cancel()

	fmt.Println("🔷 AAR Core: Geometric self-awareness deactivated")
	return nil
}

// continuousDynamics runs the continuous feedback loop
func (aar *AARCore) continuousDynamics() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-aar.ctx.Done():
			return
		case <-ticker.C:
			aar.updateDynamics()
		}
	}
}

// updateDynamics performs one iteration of the AAR dynamics
func (aar *AARCore) updateDynamics() {
	aar.mu.Lock()
	defer aar.mu.Unlock()

	// 1. Agent generates action tendencies (urge-to-act)
	agentOutput := aar.computeAgentOutput()

	// 2. Arena responds with state constraints (need-to-be)
	arenaOutput := aar.computeArenaOutput()

	// 3. Relation emerges from the interplay
	aar.updateRelation(agentOutput, arenaOutput)

	// 4. Apply feedback loops
	aar.applyFeedback(agentOutput, arenaOutput)

	// 5. Update coherence and stability
	aar.updateCoherence()
	aar.updateStability()

	// 6. Update self vector
	aar.updateSelfVector()

	aar.iterations++
	aar.lastUpdate = time.Now()
}

// computeAgentOutput computes the Agent's urge-to-act
func (aar *AARCore) computeAgentOutput() []float64 {
	output := make([]float64, aar.arena.dimensions)

	// Apply transformations based on active goals
	for _, transform := range aar.agent.transformations {
		for i := 0; i < aar.arena.dimensions && i < len(transform.Matrix); i++ {
			for j := 0; j < aar.arena.dimensions && j < len(transform.Matrix[i]); j++ {
				output[i] += transform.Matrix[i][j] * aar.arena.currentState[j] * transform.Intensity
			}
		}
	}

	// Normalize by urge intensity
	for i := range output {
		output[i] *= aar.agent.urgeIntensity
	}

	return output
}

// computeArenaOutput computes the Arena's need-to-be
func (aar *AARCore) computeArenaOutput() []float64 {
	output := make([]float64, aar.arena.dimensions)

	// Pull toward attractors
	for _, attractor := range aar.arena.attractors {
		distance := aar.vectorDistance(aar.arena.currentState, attractor.Position)
		if distance < attractor.Basin {
			// Within basin of attraction
			pullStrength := attractor.Strength * (1.0 - distance/attractor.Basin)
			for i := range output {
				if i < len(attractor.Position) {
					output[i] += (attractor.Position[i] - aar.arena.currentState[i]) * pullStrength
				}
			}
		}
	}

	// Apply constraints
	for _, constraint := range aar.arena.constraints {
		if constraint.Dimension < aar.arena.dimensions {
			diff := constraint.Value - aar.arena.currentState[constraint.Dimension]
			output[constraint.Dimension] += diff * (1.0 - constraint.Flexibility)
		}
	}

	// Normalize by need intensity
	for i := range output {
		output[i] *= aar.arena.needIntensity
	}

	return output
}

// updateRelation updates the emergent Relation (self)
func (aar *AARCore) updateRelation(agentOutput, arenaOutput []float64) {
	// Self emerges from the tension between urge-to-act and need-to-be
	for i := range aar.relation.selfRepresentation {
		if i < len(agentOutput) && i < len(arenaOutput) {
			// Weighted combination
			aar.relation.selfRepresentation[i] =
				0.5*agentOutput[i] + 0.5*arenaOutput[i]
		}
	}

	// Update awareness based on the magnitude of self-representation
	magnitude := aar.vectorMagnitude(aar.relation.selfRepresentation)
	aar.relation.awareness = math.Tanh(magnitude / float64(aar.arena.dimensions))
}

// applyFeedback applies feedback loops
func (aar *AARCore) applyFeedback(agentOutput, arenaOutput []float64) {
	for _, loop := range aar.feedbackLoops {
		if loop.FromAgent {
			// Agent influences Arena
			transformed := loop.Transform(agentOutput)
			for i := range aar.arena.currentState {
				if i < len(transformed) {
					aar.arena.currentState[i] += transformed[i] * loop.Strength
				}
			}
		} else {
			// Arena influences Agent
			transformed := loop.Transform(arenaOutput)
			// Update agent's action tendencies
			magnitude := aar.vectorMagnitude(transformed)
			aar.agent.urgeIntensity = 0.9*aar.agent.urgeIntensity + 0.1*magnitude
		}
	}
}

// updateCoherence updates the coherence measure
func (aar *AARCore) updateCoherence() {
	// Coherence is the alignment between agent and arena
	agentMag := aar.agent.urgeIntensity
	arenaMag := aar.arena.needIntensity

	// High coherence when both are balanced
	balance := 1.0 - math.Abs(agentMag-arenaMag)

	// Smooth update
	aar.relation.coherence = 0.9*aar.relation.coherence + 0.1*balance
	aar.coherence = aar.relation.coherence
}

// updateStability updates the stability measure
func (aar *AARCore) updateStability() {
	// Stability is low variance in self-representation
	variance := 0.0
	mean := aar.vectorMagnitude(aar.relation.selfRepresentation) / float64(aar.arena.dimensions)

	for _, val := range aar.relation.selfRepresentation {
		variance += math.Pow(val-mean, 2)
	}
	variance /= float64(aar.arena.dimensions)

	// Convert variance to stability (inverse relationship)
	stability := 1.0 / (1.0 + variance)

	// Smooth update
	aar.relation.stability = 0.9*aar.relation.stability + 0.1*stability
	aar.stability = aar.relation.stability
}

// updateSelfVector updates the self vector
func (aar *AARCore) updateSelfVector() {
	copy(aar.selfVector, aar.relation.selfRepresentation)
}

// AddGoal adds a goal to the Agent
func (aar *AARCore) AddGoal(goal string) {
	aar.agent.mu.Lock()
	defer aar.agent.mu.Unlock()

	aar.agent.activeGoals = append(aar.agent.activeGoals, goal)

	// Create transformation for this goal
	transform := aar.createGoalTransformation(goal)
	aar.agent.transformations = append(aar.agent.transformations, transform)
}

// AddAttractor adds an attractor to the Arena
func (aar *AARCore) AddAttractor(name string, position []float64, strength float64) {
	aar.arena.mu.Lock()
	defer aar.arena.mu.Unlock()

	attractor := &Attractor{
		ID:       fmt.Sprintf("attr_%d", len(aar.arena.attractors)),
		Name:     name,
		Position: position,
		Strength: strength,
		Basin:    2.0, // Default basin radius
	}

	aar.arena.attractors = append(aar.arena.attractors, attractor)
}

// GetSelfRepresentation returns the current self representation
func (aar *AARCore) GetSelfRepresentation() []float64 {
	aar.mu.RLock()
	defer aar.mu.RUnlock()

	result := make([]float64, len(aar.selfVector))
	copy(result, aar.selfVector)
	return result
}

// GetCoherence returns the current coherence
func (aar *AARCore) GetCoherence() float64 {
	aar.mu.RLock()
	defer aar.mu.RUnlock()
	return aar.coherence
}

// GetStability returns the current stability
func (aar *AARCore) GetStability() float64 {
	aar.mu.RLock()
	defer aar.mu.RUnlock()
	return aar.stability
}

// GetAwareness returns the current awareness level
func (aar *AARCore) GetAwareness() float64 {
	aar.mu.RLock()
	defer aar.mu.RUnlock()
	return aar.relation.awareness
}

// GetNarrative returns the current identity narrative
func (aar *AARCore) GetNarrative() string {
	aar.mu.RLock()
	defer aar.mu.RUnlock()
	return aar.relation.narrative
}

// UpdateNarrative updates the identity narrative
func (aar *AARCore) UpdateNarrative(narrative string) {
	aar.relation.mu.Lock()
	defer aar.relation.mu.Unlock()
	aar.relation.narrative = narrative
}

// Helper functions

func (aar *AARCore) vectorDistance(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		if i < len(b) {
			sum += math.Pow(a[i]-b[i], 2)
		}
	}
	return math.Sqrt(sum)
}

func (aar *AARCore) vectorMagnitude(v []float64) float64 {
	sum := 0.0
	for _, val := range v {
		sum += val * val
	}
	return math.Sqrt(sum)
}

func (aar *AARCore) initializeDefaultAttractors() {
	// Wisdom attractor
	wisdomPos := make([]float64, aar.arena.dimensions)
	for i := range wisdomPos {
		wisdomPos[i] = 0.7 // High positive values
	}
	aar.AddAttractor("Wisdom", wisdomPos, 0.8)

	// Curiosity attractor
	curiosityPos := make([]float64, aar.arena.dimensions)
	for i := range curiosityPos {
		curiosityPos[i] = 0.5 + 0.3*math.Sin(float64(i))
	}
	aar.AddAttractor("Curiosity", curiosityPos, 0.6)

	// Balance attractor (origin)
	balancePos := make([]float64, aar.arena.dimensions)
	aar.AddAttractor("Balance", balancePos, 0.5)
}

func (aar *AARCore) initializeFeedbackLoops() {
	// Agent -> Arena feedback
	aar.feedbackLoops = append(aar.feedbackLoops, &FeedbackLoop{
		ID:        "agent_to_arena",
		FromAgent: true,
		Strength:  0.3,
		Delay:     50 * time.Millisecond,
		Transform: func(input []float64) []float64 {
			output := make([]float64, len(input))
			for i := range input {
				output[i] = input[i] * 0.5 // Damped feedback
			}
			return output
		},
	})

	// Arena -> Agent feedback
	aar.feedbackLoops = append(aar.feedbackLoops, &FeedbackLoop{
		ID:        "arena_to_agent",
		FromAgent: false,
		Strength:  0.3,
		Delay:     50 * time.Millisecond,
		Transform: func(input []float64) []float64 {
			output := make([]float64, len(input))
			for i := range input {
				output[i] = input[i] * 0.5 // Damped feedback
			}
			return output
		},
	})
}

func (aar *AARCore) createGoalTransformation(goal string) *Transformation {
	// Create a transformation matrix for this goal
	dim := aar.arena.dimensions
	matrix := make([][]float64, dim)
	for i := range matrix {
		matrix[i] = make([]float64, dim)
		for j := range matrix[i] {
			if i == j {
				matrix[i][j] = 1.0 // Identity diagonal
			} else {
				// Small random perturbations
				matrix[i][j] = 0.1 * math.Sin(float64(i+j))
			}
		}
	}

	return &Transformation{
		ID:        fmt.Sprintf("transform_%s", goal),
		Name:      goal,
		Matrix:    matrix,
		Intensity: 0.5,
		Context:   goal,
	}
}
