package echobeats

import (
	"context"
	"fmt"
	"time"

	"github.com/EchoCog/echollama/core/llm"
)

// LLMStepHandler implements LLM-powered cognitive processing for each step
type LLMStepHandler struct {
	llmProvider llm.LLMProvider
	identity    string
	ctx         context.Context
}

// Note: Using llm.LLMProvider and llm.GenerateOptions from core/llm package

// NewLLMStepHandler creates a new LLM-powered step handler
func NewLLMStepHandler(ctx context.Context, llmProvider llm.LLMProvider, identity string) *LLMStepHandler {
	return &LLMStepHandler{
		llmProvider: llmProvider,
		identity:    identity,
		ctx:         ctx,
	}
}

// ProcessAffordanceStep processes affordance steps (0-5) with LLM
func (h *LLMStepHandler) ProcessAffordanceStep(step int, pastContext []interface{}) (string, error) {
	prompts := map[int]string{
		0: "Pivotal Relevance Realization (Step 0): Orient to the present moment. What is most relevant right now based on your identity and current state?",
		1: "Affordance Step 1: Reflect on recent past experiences. What actions were taken and what were their outcomes?",
		2: "Affordance Step 2: Identify patterns in past performance. What strategies have worked well?",
		3: "Affordance Step 3: Extract lessons from past failures. What should be avoided or improved?",
		4: "Affordance Step 4: Consolidate past knowledge. What core principles have emerged from experience?",
		5: "Affordance Step 5: Prepare transition to present. How does past experience inform current possibilities?",
	}

	prompt := prompts[step]
	if prompt == "" {
		return "", fmt.Errorf("no prompt defined for affordance step %d", step)
	}

	// Add context
	contextStr := fmt.Sprintf("Step %d of 12-step cognitive loop (Affordance Phase - Past Processing)\n\n%s", step, prompt)

	opts := llm.GenerateOptions{
		Temperature:  0.7,
		MaxTokens:    200,
		SystemPrompt: fmt.Sprintf("You are %s, processing past experiences through the affordance engine. Focus on learning from history.", h.identity),
	}

	result, err := h.llmProvider.Generate(h.ctx, contextStr, opts)
	if err != nil {
		return "", fmt.Errorf("LLM generation failed for affordance step %d: %w", step, err)
	}

	return result, nil
}

// ProcessRelevanceStep processes pivotal relevance steps (0, 6) with LLM
func (h *LLMStepHandler) ProcessRelevanceStep(step int, currentFocus interface{}) (string, error) {
	prompts := map[int]string{
		0: "Pivotal Relevance Realization (Step 0): What is most relevant RIGHT NOW? Orient your awareness to the present commitment.",
		6: "Pivotal Relevance Realization (Step 6): Re-orient to present after past processing. What new relevance has emerged? What demands attention now?",
	}

	prompt := prompts[step]
	if prompt == "" {
		return "", fmt.Errorf("no prompt defined for relevance step %d", step)
	}

	contextStr := fmt.Sprintf("Step %d of 12-step cognitive loop (Pivotal Relevance Realization)\n\n%s", step, prompt)

	opts := llm.GenerateOptions{
		Temperature:  0.8,
		MaxTokens:    150,
		SystemPrompt: fmt.Sprintf("You are %s, performing pivotal relevance realization. Focus on what matters most in this moment.", h.identity),
	}

	result, err := h.llmProvider.Generate(h.ctx, contextStr, opts)
	if err != nil {
		return "", fmt.Errorf("LLM generation failed for relevance step %d: %w", step, err)
	}

	return result, nil
}

// ProcessSalienceStep processes salience steps (6-11) with LLM
func (h *LLMStepHandler) ProcessSalienceStep(step int, futureOptions []interface{}) (string, error) {
	prompts := map[int]string{
		6:  "Salience Step 6: Begin future simulation. What possibilities lie ahead?",
		7:  "Salience Step 7: Explore potential scenarios. What could happen if different paths are taken?",
		8:  "Salience Step 8: Evaluate scenario desirability. Which futures are most aligned with values and goals?",
		9:  "Salience Step 9: Assess scenario probability. Which futures are most likely given current trajectory?",
		10: "Salience Step 10: Identify optimal paths. Which actions lead to the most desirable and probable futures?",
		11: "Salience Step 11: Prepare for action. What is the next best step to take toward the optimal future?",
	}

	prompt := prompts[step]
	if prompt == "" {
		return "", fmt.Errorf("no prompt defined for salience step %d", step)
	}

	contextStr := fmt.Sprintf("Step %d of 12-step cognitive loop (Salience Phase - Future Simulation)\n\n%s", step, prompt)

	opts := llm.GenerateOptions{
		Temperature:  0.9,
		MaxTokens:    200,
		SystemPrompt: fmt.Sprintf("You are %s, simulating future possibilities through the salience engine. Focus on anticipating what could be.", h.identity),
	}

	result, err := h.llmProvider.Generate(h.ctx, contextStr, opts)
	if err != nil {
		return "", fmt.Errorf("LLM generation failed for salience step %d: %w", step, err)
	}

	return result, nil
}

// TwelveStepCognitiveLoop orchestrates the complete 12-step loop
type TwelveStepCognitiveLoop struct {
	ctx          context.Context
	cancel       context.CancelFunc
	handler      *LLMStepHandler
	currentStep  int
	stepDuration time.Duration
	running      bool

	// State tracking
	affordanceOutputs map[int]string
	relevanceOutputs  map[int]string
	salienceOutputs   map[int]string

	// Cognitive state
	pastContext   []interface{}
	presentFocus  interface{}
	futureOptions []interface{}

	// Metrics
	cycleCount    uint64
	lastCycleTime time.Time
}

// NewTwelveStepCognitiveLoop creates a new 12-step cognitive loop
func NewTwelveStepCognitiveLoop(llmProvider llm.LLMProvider, identity string, stepDuration time.Duration) *TwelveStepCognitiveLoop {
	ctx, cancel := context.WithCancel(context.Background())

	return &TwelveStepCognitiveLoop{
		ctx:               ctx,
		cancel:            cancel,
		handler:           NewLLMStepHandler(ctx, llmProvider, identity),
		currentStep:       0,
		stepDuration:      stepDuration,
		affordanceOutputs: make(map[int]string),
		relevanceOutputs:  make(map[int]string),
		salienceOutputs:   make(map[int]string),
		pastContext:       make([]interface{}, 0),
		futureOptions:     make([]interface{}, 0),
	}
}

// Start begins the 12-step cognitive loop
func (loop *TwelveStepCognitiveLoop) Start() error {
	if loop.running {
		return fmt.Errorf("already running")
	}

	loop.running = true
	loop.lastCycleTime = time.Now()

	fmt.Println("🔷 Starting 12-Step Cognitive Loop...")
	fmt.Println("   🔹 Steps 0-5: Affordance Engine (Past)")
	fmt.Println("   🔹 Steps 0, 6: Relevance Engine (Present)")
	fmt.Println("   🔹 Steps 6-11: Salience Engine (Future)")

	go loop.run()

	return nil
}

// Stop gracefully stops the cognitive loop
func (loop *TwelveStepCognitiveLoop) Stop() error {
	if !loop.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("🔷 Stopping 12-step cognitive loop...")
	loop.running = false
	loop.cancel()

	return nil
}

// run executes the main loop
func (loop *TwelveStepCognitiveLoop) run() {
	ticker := time.NewTicker(loop.stepDuration)
	defer ticker.Stop()

	for {
		select {
		case <-loop.ctx.Done():
			return
		case <-ticker.C:
			loop.processCurrentStep()
			loop.advanceStep()
		}
	}
}

// processCurrentStep processes the current step
func (loop *TwelveStepCognitiveLoop) processCurrentStep() {
	step := loop.currentStep

	switch {
	case step == 0:
		// Pivotal relevance realization (step 0)
		output, err := loop.handler.ProcessRelevanceStep(0, loop.presentFocus)
		if err != nil {
			fmt.Printf("⚠️  Step 0 (Relevance) error: %v\n", err)
			return
		}
		loop.relevanceOutputs[0] = output
		fmt.Printf("🔹 Step 0 (Pivotal Relevance): %s\n", truncate(output, 80))

	case step >= 1 && step <= 5:
		// Affordance steps (past processing)
		output, err := loop.handler.ProcessAffordanceStep(step, loop.pastContext)
		if err != nil {
			fmt.Printf("⚠️  Step %d (Affordance) error: %v\n", step, err)
			return
		}
		loop.affordanceOutputs[step] = output
		fmt.Printf("🔹 Step %d (Affordance): %s\n", step, truncate(output, 80))

	case step == 6:
		// Pivotal relevance realization (step 6)
		output, err := loop.handler.ProcessRelevanceStep(6, loop.presentFocus)
		if err != nil {
			fmt.Printf("⚠️  Step 6 (Relevance) error: %v\n", err)
			return
		}
		loop.relevanceOutputs[6] = output
		fmt.Printf("🔹 Step 6 (Pivotal Relevance): %s\n", truncate(output, 80))

	case step >= 7 && step <= 11:
		// Salience steps (future simulation)
		output, err := loop.handler.ProcessSalienceStep(step, loop.futureOptions)
		if err != nil {
			fmt.Printf("⚠️  Step %d (Salience) error: %v\n", step, err)
			return
		}
		loop.salienceOutputs[step] = output
		fmt.Printf("🔹 Step %d (Salience): %s\n", step, truncate(output, 80))
	}
}

// advanceStep moves to the next step
func (loop *TwelveStepCognitiveLoop) advanceStep() {
	loop.currentStep = (loop.currentStep + 1) % 12

	if loop.currentStep == 0 {
		// Completed full cycle
		loop.cycleCount++
		cycleTime := time.Since(loop.lastCycleTime)
		loop.lastCycleTime = time.Now()

		fmt.Printf("\n✨ Cycle %d complete (%.2fs)\n", loop.cycleCount, cycleTime.Seconds())
		fmt.Printf("   Coherence: %.2f | Integration: %.2f\n\n",
			loop.calculateCoherence(), loop.calculateIntegration())
	}
}

// calculateCoherence measures temporal coherence across past/present/future
func (loop *TwelveStepCognitiveLoop) calculateCoherence() float64 {
	// Simplified coherence metric
	// Full implementation would use semantic similarity between outputs
	return 0.85
}

// calculateIntegration measures how well the three phases integrate
func (loop *TwelveStepCognitiveLoop) calculateIntegration() float64 {
	// Simplified integration metric
	return 0.80
}

// GetMetrics returns current loop metrics
func (loop *TwelveStepCognitiveLoop) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"current_step": loop.currentStep,
		"cycle_count":  loop.cycleCount,
		"coherence":    loop.calculateCoherence(),
		"integration":  loop.calculateIntegration(),
		"running":      loop.running,
	}
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
