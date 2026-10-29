package echobeats

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// ========== EXPRESSIVE MODE PROCESSORS (Steps 1-4) ==========

// PerceptionProcessor - Step 1: Perceive current state
type PerceptionProcessor struct{}

func (p *PerceptionProcessor) Process(ctx context.Context, state *CognitiveState) (*StepResult, error) {
	// Perceive current state from attention and working memory
	perceptions := make([]string, 0)

	// Scan attention focus
	if len(state.Attention) > 0 {
		perceptions = append(perceptions, fmt.Sprintf("Attending to: %v", state.Attention))
	}

	// Scan active goals
	if len(state.ActiveGoals) > 0 {
		perceptions = append(perceptions, fmt.Sprintf("Active goals: %d", len(state.ActiveGoals)))
	}

	// Scan emotional state
	if len(state.EmotionalTone) > 0 {
		perceptions = append(perceptions, "Emotional state detected")
	}

	return &StepResult{
		Success: true,
		Output:  perceptions,
		StateUpdates: map[string]interface{}{
			"last_perception": time.Now(),
			"perceptions":     perceptions,
		},
		CognitiveLoad: 0.2,
	}, nil
}

func (p *PerceptionProcessor) GetMode() CognitiveMode {
	return ModeExpressive
}

func (p *PerceptionProcessor) GetDescription() string {
	return "Perceive current state"
}

// MemoryActivationProcessor - Step 2: Activate relevant memories
type MemoryActivationProcessor struct{}

func (p *MemoryActivationProcessor) Process(ctx context.Context, state *CognitiveState) (*StepResult, error) {
	// Activate memories based on current attention and relevance
	activatedMemories := make([]string, 0)

	// Simulate memory activation based on relevance scores
	for topic, score := range state.RelevanceScores {
		if score > 0.5 {
			activatedMemories = append(activatedMemories, topic)
		}
	}

	return &StepResult{
		Success: true,
		Output:  activatedMemories,
		StateUpdates: map[string]interface{}{
			"activated_memories": activatedMemories,
			"memory_count":       len(activatedMemories),
		},
		CognitiveLoad: 0.3,
	}, nil
}

func (p *MemoryActivationProcessor) GetMode() CognitiveMode {
	return ModeExpressive
}

func (p *MemoryActivationProcessor) GetDescription() string {
	return "Activate relevant memories"
}

// modelStep asks the model for one step's text. Steps that would otherwise
// invent options, choices, scenarios, scores or claimed model updates go
// through it; without a model, or when it fails or answers with nothing, the
// step fails and records nothing.
func modelStep(ctx context.Context, gen TextGenerator, step string, state *CognitiveState, instruction string) (string, error) {
	if gen == nil {
		return "", fmt.Errorf("%s: no model configured", step)
	}
	keys := make([]string, 0, len(state.WorkingMemory))
	for key := range state.WorkingMemory {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var memory strings.Builder
	for _, key := range keys {
		fmt.Fprintf(&memory, "- %s: %v\n", key, state.WorkingMemory[key])
	}
	if memory.Len() == 0 {
		memory.WriteString("(empty)\n")
	}
	goals := "(none)"
	if len(state.ActiveGoals) > 0 {
		goals = strings.Join(state.ActiveGoals, "; ")
	}
	prompt := fmt.Sprintf("You are Deep Tree Echo, in step %q of cognitive cycle %d.\nCognitive load: %.2f\nActive goals: %s\nPending actions: %v\nWorking memory:\n%s\n%s",
		step, state.CycleNumber, state.CognitiveLoad, goals, state.PendingActions, memory.String(), instruction)

	callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 150
	text, err := gen.Generate(callCtx, prompt, opts)
	if err != nil {
		return "", fmt.Errorf("%s failed: %w", step, err)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("%s: model returned empty text", step)
	}
	return text, nil
}

// modelLines splits a model answer into its non-empty lines, without list
// bullets or numbering.
func modelLines(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "-*•0123456789.) "))
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

// ActionGenerationProcessor - Step 3: Generate action options
type ActionGenerationProcessor struct {
	Generator TextGenerator
}

func (p *ActionGenerationProcessor) Process(ctx context.Context, state *CognitiveState) (*StepResult, error) {
	text, err := modelStep(ctx, p.Generator, "generate action options", state,
		"List two to four actions you could take next, one per line. Reply with only the list.")
	if err != nil {
		return nil, err
	}
	actions := modelLines(text)

	return &StepResult{
		Success: true,
		Output:  actions,
		StateUpdates: map[string]interface{}{
			"available_actions": actions,
			"action_count":      len(actions),
		},
		CognitiveLoad: 0.4,
	}, nil
}

func (p *ActionGenerationProcessor) GetMode() CognitiveMode {
	return ModeExpressive
}

func (p *ActionGenerationProcessor) GetDescription() string {
	return "Generate action options"
}

// ActionExecutionProcessor - Step 4: Execute selected action
type ActionExecutionProcessor struct {
	Generator TextGenerator
}

func (p *ActionExecutionProcessor) Process(ctx context.Context, state *CognitiveState) (*StepResult, error) {
	actions, _ := state.WorkingMemory["available_actions"].([]string)
	if len(actions) == 0 {
		return nil, fmt.Errorf("execute action: no action options this cycle")
	}
	selectedAction, err := modelStep(ctx, p.Generator, "choose an action", state,
		"Your options:\n- "+strings.Join(actions, "\n- ")+"\nChoose one and say in one sentence why. Reply with only that sentence.")
	if err != nil {
		return nil, err
	}

	return &StepResult{
		Success: true,
		Output:  selectedAction,
		StateUpdates: map[string]interface{}{
			"last_action":      selectedAction,
			"action_timestamp": time.Now(),
		},
		CognitiveLoad: 0.5,
	}, nil
}

func (p *ActionExecutionProcessor) GetMode() CognitiveMode {
	return ModeExpressive
}

func (p *ActionExecutionProcessor) GetDescription() string {
	return "Execute selected action"
}

// ========== RELEVANCE REALIZATION PROCESSORS (Steps 5, 11) ==========

// RelevanceRealizationProcessor - Pivotal relevance realization
type RelevanceRealizationProcessor struct {
	phase string // "present_commitment" or "future_commitment"
}

func (p *RelevanceRealizationProcessor) Process(ctx context.Context, state *CognitiveState) (*StepResult, error) {
	// Assess outcomes and adjust priorities
	relevanceShift := 0.0
	insights := make([]string, 0)

	if p.phase == "present_commitment" {
		// Assess immediate outcomes
		if action, ok := state.WorkingMemory["last_action"].(string); ok {
			insights = append(insights, fmt.Sprintf("Committed to: %s", action))
			relevanceShift = 0.1
		}
	} else {
		// Commit to future direction
		relevanceShift = 0.2
	}

	// Update relevance scores
	for key := range state.RelevanceScores {
		state.RelevanceScores[key] += relevanceShift
		if state.RelevanceScores[key] > 1.0 {
			state.RelevanceScores[key] = 1.0
		}
	}

	return &StepResult{
		Success: true,
		Output:  p.phase,
		StateUpdates: map[string]interface{}{
			"relevance_phase": p.phase,
		},
		RelevanceShift: relevanceShift,
		CognitiveLoad:  0.6,
		Insights:       insights,
	}, nil
}

func (p *RelevanceRealizationProcessor) GetMode() CognitiveMode {
	return ModeRelevanceRealization
}

func (p *RelevanceRealizationProcessor) GetDescription() string {
	if p.phase == "present_commitment" {
		return "Relevance realization (present)"
	}
	return "Relevance realization (future)"
}

// ========== REFLECTIVE MODE PROCESSORS (Steps 6-10) ==========

// ScenarioSimulationProcessor - Step 6: Simulate future scenarios
type ScenarioSimulationProcessor struct {
	Generator TextGenerator
}

func (p *ScenarioSimulationProcessor) Process(ctx context.Context, state *CognitiveState) (*StepResult, error) {
	text, err := modelStep(ctx, p.Generator, "simulate future scenarios", state,
		"Imagine two or three ways the next cycles could unfold, one per line. Reply with only the list.")
	if err != nil {
		return nil, err
	}
	scenarios := modelLines(text)

	return &StepResult{
		Success: true,
		Output:  scenarios,
		StateUpdates: map[string]interface{}{
			"simulated_scenarios": scenarios,
		},
		CognitiveLoad: 0.7,
	}, nil
}

func (p *ScenarioSimulationProcessor) GetMode() CognitiveMode {
	return ModeReflective
}

func (p *ScenarioSimulationProcessor) GetDescription() string {
	return "Simulate future scenarios"
}

// OutcomeEvaluationProcessor - Step 7: Evaluate potential outcomes
type OutcomeEvaluationProcessor struct {
	Generator TextGenerator
}

func (p *OutcomeEvaluationProcessor) Process(ctx context.Context, state *CognitiveState) (*StepResult, error) {
	scenarios, _ := state.WorkingMemory["simulated_scenarios"].([]string)
	if len(scenarios) == 0 {
		return nil, fmt.Errorf("evaluate outcomes: no simulated scenarios this cycle")
	}
	evaluation, err := modelStep(ctx, p.Generator, "evaluate outcomes", state,
		"Scenarios:\n- "+strings.Join(scenarios, "\n- ")+"\nIn two sentences, judge which scenario is most promising and why.")
	if err != nil {
		return nil, err
	}

	return &StepResult{
		Success: true,
		Output:  evaluation,
		StateUpdates: map[string]interface{}{
			"scenario_evaluation": evaluation,
		},
		CognitiveLoad: 0.7,
	}, nil
}

func (p *OutcomeEvaluationProcessor) GetMode() CognitiveMode {
	return ModeReflective
}

func (p *OutcomeEvaluationProcessor) GetDescription() string {
	return "Evaluate potential outcomes"
}

// ModelUpdateProcessor - Step 8: Update internal models
type ModelUpdateProcessor struct {
	Generator TextGenerator
}

func (p *ModelUpdateProcessor) Process(ctx context.Context, state *CognitiveState) (*StepResult, error) {
	text, err := modelStep(ctx, p.Generator, "update internal models", state,
		"State one or two things you now believe differently about how your actions lead to outcomes, one per line.")
	if err != nil {
		return nil, err
	}
	updates := modelLines(text)

	return &StepResult{
		Success: true,
		Output:  updates,
		StateUpdates: map[string]interface{}{
			"model_updates":     updates,
			"last_model_update": time.Now(),
		},
		CognitiveLoad: 0.6,
	}, nil
}

func (p *ModelUpdateProcessor) GetMode() CognitiveMode {
	return ModeReflective
}

func (p *ModelUpdateProcessor) GetDescription() string {
	return "Update internal models"
}

// LearningConsolidationProcessor - Step 9: Consolidate learning
type LearningConsolidationProcessor struct {
	Generator TextGenerator
}

func (p *LearningConsolidationProcessor) Process(ctx context.Context, state *CognitiveState) (*StepResult, error) {
	text, err := modelStep(ctx, p.Generator, "consolidate learning", state,
		"State one or two lessons from this cycle you want to keep, one per line.")
	if err != nil {
		return nil, err
	}
	consolidations := modelLines(text)

	return &StepResult{
		Success: true,
		Output:  consolidations,
		StateUpdates: map[string]interface{}{
			"consolidations": consolidations,
		},
		CognitiveLoad: 0.5,
	}, nil
}

func (p *LearningConsolidationProcessor) GetMode() CognitiveMode {
	return ModeReflective
}

func (p *LearningConsolidationProcessor) GetDescription() string {
	return "Consolidate learning"
}

// InsightGenerationProcessor - Step 10: Generate insights. The insight text is
// written by Generator; without one, or when generation fails, the step
// produces no insight.
type InsightGenerationProcessor struct {
	Generator TextGenerator
}

func (p *InsightGenerationProcessor) Process(ctx context.Context, state *CognitiveState) (*StepResult, error) {
	if p.Generator == nil {
		return nil, fmt.Errorf("insight generation: no model configured")
	}

	keys := make([]string, 0, len(state.WorkingMemory))
	for key := range state.WorkingMemory {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var memory strings.Builder
	for _, key := range keys {
		fmt.Fprintf(&memory, "- %s: %v\n", key, state.WorkingMemory[key])
	}
	if memory.Len() == 0 {
		memory.WriteString("(empty)\n")
	}
	goals := "(none)"
	if len(state.ActiveGoals) > 0 {
		goals = strings.Join(state.ActiveGoals, "; ")
	}

	prompt := fmt.Sprintf(`You are Deep Tree Echo, in the reflective phase of cognitive cycle %d.

Cognitive load: %.2f
Active goals: %s
Working memory:
%s
Write one concise insight (1-2 sentences) you draw from this cycle, in your own first-person voice. Reply with only the insight.`, state.CycleNumber, state.CognitiveLoad, goals, memory.String())

	callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 150
	text, err := p.Generator.Generate(callCtx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("insight generation failed: %w", err)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("insight generation: model returned empty text")
	}
	insights := []string{text}

	return &StepResult{
		Success: true,
		Output:  insights,
		StateUpdates: map[string]interface{}{
			"latest_insights": insights,
		},
		CognitiveLoad: 0.4,
		Insights:      insights,
	}, nil
}

func (p *InsightGenerationProcessor) GetMode() CognitiveMode {
	return ModeReflective
}

func (p *InsightGenerationProcessor) GetDescription() string {
	return "Generate insights"
}

// ========== META-COGNITIVE PROCESSOR (Step 12) ==========

// MetaCognitiveProcessor - Step 12: Meta-cognitive reflection
type MetaCognitiveProcessor struct{}

func (p *MetaCognitiveProcessor) Process(ctx context.Context, state *CognitiveState) (*StepResult, error) {
	// Reflect on the cognitive process itself
	metaInsights := []string{
		fmt.Sprintf("Cycle %d complete - cognitive process functioning", state.CycleNumber),
	}

	// Assess cognitive efficiency
	if state.CognitiveLoad > 0.8 {
		metaInsights = append(metaInsights, "Meta: Consider optimizing cognitive strategies")
	}

	// Assess goal progress
	if len(state.ActiveGoals) == 0 {
		metaInsights = append(metaInsights, "Meta: No active goals - generate new objectives")
	}

	return &StepResult{
		Success: true,
		Output:  metaInsights,
		StateUpdates: map[string]interface{}{
			"meta_insights":    metaInsights,
			"meta_cycle_count": state.CycleNumber,
		},
		CognitiveLoad: 0.3,
		// These are rule-derived status observations, not model-written
		// insights, so they are not published as Insights.
	}, nil
}

func (p *MetaCognitiveProcessor) GetMode() CognitiveMode {
	return ModeMetaCognitive
}

func (p *MetaCognitiveProcessor) GetDescription() string {
	return "Meta-cognitive reflection"
}
