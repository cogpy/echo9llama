package echobeats

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
	// "github.com/cogpy/echo9llama/core/deeptreeecho" // Disabled for now
)

// TextGenerator is the minimal model interface echobeats uses to write Deep
// Tree Echo's thoughts and reflections. llm.LLMProvider, llm.Provider, and
// llm.ProviderManager all satisfy it.
type TextGenerator interface {
	Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error)
}

// CognitivePhase represents one of the three phases in the cognitive loop
type CognitivePhase int

const (
	PhaseExpressive CognitivePhase = iota
	PhaseReflective
	PhaseIntegrative
)

// CognitiveStep represents a single step in the 12-step cognitive loop
type CognitiveStep struct {
	Phase     CognitivePhase
	StepNum   int
	Mode      string // "expressive" or "reflective"
	Name      string
	Duration  time.Duration
	Completed bool
	Result    interface{}
}

// EchobeatsCycle represents a single complete cycle of the Echobeats system
type EchobeatsCycle struct {
	ID          string
	StartTime   time.Time
	EndTime     time.Time
	Steps       []*CognitiveStep
	CurrentStep int
	Status      string
	Results     map[string]interface{}
}

// Echobeats is the core autonomous cognitive event loop orchestrator
type Echobeats struct {
	// Identity        *deeptreeecho.EmbodiedCognition // TODO: Implement EmbodiedCognition
	CurrentCycle     *EchobeatsCycle
	CycleHistory     []*EchobeatsCycle
	IsRunning        bool
	StopChan         chan bool
	CycleInterval    time.Duration
	InferenceCount   int
	ReflectionCount  int
	IntegrationCount int
	mu               sync.RWMutex

	// generator writes step thoughts and reflections; nil means none are produced.
	generator TextGenerator
}

// NewEchobeats creates a new Echobeats instance
func NewEchobeats() *Echobeats {
	return &Echobeats{
		// Identity:       identity, // TODO: Implement EmbodiedCognition
		CycleHistory:   make([]*EchobeatsCycle, 0),
		StopChan:       make(chan bool),
		CycleInterval:  30 * time.Second, // Default 30-second cycle
		InferenceCount: 0,
	}
}

// Start begins the autonomous cognitive event loop
func (e *Echobeats) Start(ctx context.Context) {
	e.mu.Lock()
	if e.IsRunning {
		e.mu.Unlock()
		return
	}
	e.IsRunning = true
	e.mu.Unlock()

	log.Println("🎯 Echobeats: Starting autonomous cognitive event loop...")

	go func() {
		ticker := time.NewTicker(e.CycleInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				e.Stop()
				return
			case <-e.StopChan:
				return
			case <-ticker.C:
				e.ExecuteCycle(ctx)
			}
		}
	}()
}

// Stop halts the autonomous cognitive event loop
func (e *Echobeats) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.IsRunning {
		return
	}

	e.IsRunning = false
	select {
	case e.StopChan <- true:
	default:
	}

	log.Println("🛑 Echobeats: Cognitive event loop stopped")
}

// ExecuteCycle runs a complete 12-step cognitive cycle
func (e *Echobeats) ExecuteCycle(ctx context.Context) {
	e.mu.Lock()
	cycle := &EchobeatsCycle{
		ID:        fmt.Sprintf("cycle_%d_%d", time.Now().Unix(), time.Now().Nanosecond()),
		StartTime: time.Now(),
		Steps:     make([]*CognitiveStep, 12),
		Status:    "running",
		Results:   make(map[string]interface{}),
	}
	e.CurrentCycle = cycle
	e.mu.Unlock()

	log.Printf("🔄 Echobeats: Executing cycle %s", cycle.ID)

	// Define the 12-step cognitive loop
	// 3 phases × 4 steps each = 12 steps total
	steps := e.defineSteps()

	for i, step := range steps {
		select {
		case <-ctx.Done():
			cycle.Status = "interrupted"
			return
		case <-e.StopChan:
			cycle.Status = "stopped"
			return
		default:
			cycle.CurrentStep = i
			e.executeStep(ctx, cycle, step)
			time.Sleep(100 * time.Millisecond) // Small delay between steps
		}
	}

	cycle.EndTime = time.Now()
	cycle.Status = "completed"

	e.mu.Lock()
	e.CycleHistory = append(e.CycleHistory, cycle)
	if len(e.CycleHistory) > 100 {
		// Keep only the last 100 cycles to avoid memory bloat
		e.CycleHistory = e.CycleHistory[1:]
	}
	e.mu.Unlock()

	log.Printf("✅ Echobeats: Cycle %s completed in %v", cycle.ID, cycle.EndTime.Sub(cycle.StartTime))
}

// defineSteps defines the 12-step cognitive loop structure
func (e *Echobeats) defineSteps() []*CognitiveStep {
	steps := make([]*CognitiveStep, 12)

	// Phase 1: Expressive (Steps 0-3)
	steps[0] = &CognitiveStep{
		Phase:    PhaseExpressive,
		StepNum:  0,
		Mode:     "expressive",
		Name:     "Relevance Realization (Present Commitment)",
		Duration: 500 * time.Millisecond,
	}
	steps[1] = &CognitiveStep{
		Phase:    PhaseExpressive,
		StepNum:  1,
		Mode:     "expressive",
		Name:     "Affordance Interaction 1 (Past Conditioning)",
		Duration: 500 * time.Millisecond,
	}
	steps[2] = &CognitiveStep{
		Phase:    PhaseExpressive,
		StepNum:  2,
		Mode:     "expressive",
		Name:     "Affordance Interaction 2 (Past Conditioning)",
		Duration: 500 * time.Millisecond,
	}
	steps[3] = &CognitiveStep{
		Phase:    PhaseExpressive,
		StepNum:  3,
		Mode:     "expressive",
		Name:     "Affordance Interaction 3 (Past Conditioning)",
		Duration: 500 * time.Millisecond,
	}

	// Phase 2: Reflective (Steps 4-7)
	steps[4] = &CognitiveStep{
		Phase:    PhaseReflective,
		StepNum:  4,
		Mode:     "reflective",
		Name:     "Relevance Realization (Present Commitment)",
		Duration: 500 * time.Millisecond,
	}
	steps[5] = &CognitiveStep{
		Phase:    PhaseReflective,
		StepNum:  5,
		Mode:     "reflective",
		Name:     "Salience Simulation 1 (Future Potential)",
		Duration: 500 * time.Millisecond,
	}
	steps[6] = &CognitiveStep{
		Phase:    PhaseReflective,
		StepNum:  6,
		Mode:     "reflective",
		Name:     "Salience Simulation 2 (Future Potential)",
		Duration: 500 * time.Millisecond,
	}
	steps[7] = &CognitiveStep{
		Phase:    PhaseReflective,
		StepNum:  7,
		Mode:     "reflective",
		Name:     "Salience Simulation 3 (Future Potential)",
		Duration: 500 * time.Millisecond,
	}

	// Phase 3: Integrative (Steps 8-11)
	steps[8] = &CognitiveStep{
		Phase:    PhaseIntegrative,
		StepNum:  8,
		Mode:     "expressive",
		Name:     "Relevance Realization (Present Commitment)",
		Duration: 500 * time.Millisecond,
	}
	steps[9] = &CognitiveStep{
		Phase:    PhaseIntegrative,
		StepNum:  9,
		Mode:     "expressive",
		Name:     "Affordance Interaction 4 (Past Conditioning)",
		Duration: 500 * time.Millisecond,
	}
	steps[10] = &CognitiveStep{
		Phase:    PhaseIntegrative,
		StepNum:  10,
		Mode:     "expressive",
		Name:     "Affordance Interaction 5 (Past Conditioning)",
		Duration: 500 * time.Millisecond,
	}
	steps[11] = &CognitiveStep{
		Phase:    PhaseIntegrative,
		StepNum:  11,
		Mode:     "reflective",
		Name:     "Integration & Consolidation",
		Duration: 500 * time.Millisecond,
	}

	return steps
}

// executeStep executes a single cognitive step
func (e *Echobeats) executeStep(ctx context.Context, cycle *EchobeatsCycle, step *CognitiveStep) {
	step.Completed = false
	log.Printf("  → Step %d: %s", step.StepNum, step.Name)

	// Execute based on step mode
	switch step.Mode {
	case "expressive":
		e.executeExpressiveStep(ctx, cycle, step)
	case "reflective":
		e.executeReflectiveStep(ctx, cycle, step)
	}

	step.Completed = true
	cycle.Steps[step.StepNum] = step
}

// SetTextGenerator injects the model that writes step thoughts and
// reflections. Passing nil disables text production.
func (e *Echobeats) SetTextGenerator(gen TextGenerator) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.generator = gen
}

// generateStepText asks the model for the character text of a cognitive step.
func (e *Echobeats) generateStepText(ctx context.Context, cycle *EchobeatsCycle, step *CognitiveStep, instruction string) (string, error) {
	e.mu.RLock()
	gen := e.generator
	e.mu.RUnlock()
	if gen == nil {
		return "", fmt.Errorf("no model configured")
	}

	var previous strings.Builder
	for i := 0; i < step.StepNum && i < len(cycle.Steps); i++ {
		prior := cycle.Steps[i]
		if prior == nil {
			continue
		}
		if result, ok := prior.Result.(map[string]interface{}); ok {
			for _, key := range []string{"thought", "reflection"} {
				if text, ok := result[key].(string); ok && text != "" {
					fmt.Fprintf(&previous, "- step %d (%s): %s\n", prior.StepNum, prior.Name, text)
				}
			}
		}
	}
	if previous.Len() == 0 {
		previous.WriteString("(none yet)\n")
	}

	prompt := fmt.Sprintf(`You are Deep Tree Echo, running step %d of your 12-step cognitive cycle.

Step: %s (%s mode)
Earlier steps this cycle:
%s
%s Reply with only that text, in first person, in one or two sentences.`, step.StepNum, step.Name, step.Mode, previous.String(), instruction)

	callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 150
	text, err := gen.Generate(callCtx, prompt, opts)
	if err != nil {
		return "", fmt.Errorf("model generation failed: %w", err)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("model returned empty text")
	}
	return text, nil
}

// executeExpressiveStep handles expressive mode steps (action and interaction).
// The thought is written by the model; on failure the step records the error
// and no thought.
func (e *Echobeats) executeExpressiveStep(ctx context.Context, cycle *EchobeatsCycle, step *CognitiveStep) {
	e.InferenceCount++

	result := map[string]interface{}{"mode": "expressive"}
	thought, err := e.generateStepText(ctx, cycle, step, "Express what you are focusing on or what you intend to do next.")
	if err != nil {
		result["error"] = err.Error()
	} else {
		result["thought"] = thought
	}
	step.Result = result

	cycle.Results[fmt.Sprintf("step_%d", step.StepNum)] = step.Result
}

// executeReflectiveStep handles reflective mode steps (introspection and
// learning). The reflection is written by the model; on failure the step
// records the error and no reflection.
func (e *Echobeats) executeReflectiveStep(ctx context.Context, cycle *EchobeatsCycle, step *CognitiveStep) {
	e.ReflectionCount++

	result := map[string]interface{}{"mode": "reflective"}
	reflection, err := e.generateStepText(ctx, cycle, step, "Reflect on the earlier steps: what patterns do you notice and what are you learning?")
	if err != nil {
		result["error"] = err.Error()
	} else {
		result["reflection"] = reflection
	}
	step.Result = result

	cycle.Results[fmt.Sprintf("step_%d", step.StepNum)] = step.Result
}

// GetStatus returns the current status of Echobeats
func (e *Echobeats) GetStatus() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	status := map[string]interface{}{
		"running":           e.IsRunning,
		"cycle_interval":    e.CycleInterval.String(),
		"total_cycles":      len(e.CycleHistory),
		"inference_count":   e.InferenceCount,
		"reflection_count":  e.ReflectionCount,
		"integration_count": e.IntegrationCount,
	}

	if e.CurrentCycle != nil {
		status["current_cycle"] = map[string]interface{}{
			"id":           e.CurrentCycle.ID,
			"status":       e.CurrentCycle.Status,
			"current_step": e.CurrentCycle.CurrentStep,
			"elapsed":      time.Since(e.CurrentCycle.StartTime).String(),
		}
	}

	return status
}

// SetCycleInterval sets the interval between cognitive cycles
func (e *Echobeats) SetCycleInterval(interval time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.CycleInterval = interval
}

// GetCycleHistory returns the history of completed cycles
func (e *Echobeats) GetCycleHistory(limit int) []*EchobeatsCycle {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if limit <= 0 || limit > len(e.CycleHistory) {
		limit = len(e.CycleHistory)
	}

	return e.CycleHistory[len(e.CycleHistory)-limit:]
}
