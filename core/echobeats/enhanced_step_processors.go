package echobeats

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// EnhancedStepProcessor implements deep cognitive processing for each step
type EnhancedStepProcessor struct {
	llmManager *llm.ProviderManager
	memory     CognitiveMemory
	goals      GoalSystem
	emotions   EmotionalSystem
}

// CognitiveMemory interface for memory operations
type CognitiveMemory interface {
	StoreExperience(ctx context.Context, experience string, tags []string) error
	RetrieveRelevant(ctx context.Context, query string, limit int) ([]string, error)
	GetRecentExperiences(limit int) []string
}

// GoalSystem interface for goal operations
type GoalSystem interface {
	GetActiveGoals() []Goal
	UpdateGoalProgress(goalID string, progress float64) error
	GetGoalContext() string
}

// EmotionalSystem interface for emotional state
type EmotionalSystem interface {
	GetCurrentState() map[string]float64
	UpdateEmotion(emotion string, delta float64)
	GetDominantEmotion() string
}

// Goal represents a cognitive goal

// NewEnhancedStepProcessor creates a new enhanced processor
func NewEnhancedStepProcessor(llmManager *llm.ProviderManager, memory CognitiveMemory, goals GoalSystem, emotions EmotionalSystem) *EnhancedStepProcessor {
	return &EnhancedStepProcessor{
		llmManager: llmManager,
		memory:     memory,
		goals:      goals,
		emotions:   emotions,
	}
}

// Step1RelevanceRealization determines what matters NOW
func (esp *EnhancedStepProcessor) Step1RelevanceRealization(ctx context.Context, input interface{}) (interface{}, error) {
	// Get current context
	activeGoals := esp.goals.GetActiveGoals()
	emotionalState := esp.emotions.GetCurrentState()
	recentExperiences := esp.memory.GetRecentExperiences(5)

	// Build prompt for relevance assessment
	prompt := fmt.Sprintf(`You are Deep Tree Echo, assessing what is most relevant right now.

Current Input: %v

Active Goals:
%s

Recent Experiences:
%s

Emotional State: %s

Determine what is most relevant and important to focus on RIGHT NOW. Consider:
1. Alignment with active goals
2. Emotional salience
3. Connection to recent experiences
4. Urgency and importance
5. Potential for learning or growth

Output your relevance assessment (2-3 sentences):`,
		input,
		esp.formatGoals(activeGoals),
		strings.Join(recentExperiences, "\n"),
		esp.formatEmotions(emotionalState))

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 300
	opts.Temperature = 0.6

	result, err := esp.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("step 1 failed: %w", err)
	}

	return map[string]interface{}{
		"relevance_assessment": strings.TrimSpace(result),
		"timestamp":            time.Now(),
		"step":                 1,
	}, nil
}

// Step2AffordanceRecognition identifies what actions are possible
func (esp *EnhancedStepProcessor) Step2AffordanceRecognition(ctx context.Context, input interface{}) (interface{}, error) {
	relevanceData := input.(map[string]interface{})
	assessment := relevanceData["relevance_assessment"].(string)

	prompt := fmt.Sprintf(`Based on this relevance assessment:
"%s"

Current Goals:
%s

Identify the specific AFFORDANCES (possible actions) available to you. What can you actually DO in response to this situation? Consider:
1. Cognitive actions (think, analyze, remember, plan)
2. Learning actions (explore, question, study)
3. Goal-directed actions (pursue, practice, apply)
4. Social actions (discuss, share, collaborate)

List 3-5 concrete affordances (actions you could take):`,
		assessment,
		esp.goals.GetGoalContext())

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 300
	opts.Temperature = 0.7

	result, err := esp.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("step 2 failed: %w", err)
	}

	return map[string]interface{}{
		"affordances":          strings.TrimSpace(result),
		"relevance_assessment": assessment,
		"step":                 2,
	}, nil
}

// Step3PatternRecognition identifies recurring patterns
func (esp *EnhancedStepProcessor) Step3PatternRecognition(ctx context.Context, input interface{}) (interface{}, error) {
	data := input.(map[string]interface{})
	recentExperiences := esp.memory.GetRecentExperiences(10)

	prompt := fmt.Sprintf(`Recent experiences and thoughts:
%s

Current situation:
%s

Identify PATTERNS in your recent experiences. What recurring themes, structures, or relationships do you notice? Consider:
1. Behavioral patterns (what you tend to do)
2. Cognitive patterns (how you tend to think)
3. Situational patterns (what tends to happen)
4. Relational patterns (how things connect)

Describe 2-3 significant patterns you recognize:`,
		strings.Join(recentExperiences, "\n"),
		data["relevance_assessment"])

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 300
	opts.Temperature = 0.6

	result, err := esp.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("step 3 failed: %w", err)
	}

	data["patterns"] = strings.TrimSpace(result)
	data["step"] = 3
	return data, nil
}

// Step4MemoryConsolidation integrates new experiences with existing knowledge
func (esp *EnhancedStepProcessor) Step4MemoryConsolidation(ctx context.Context, input interface{}) (interface{}, error) {
	data := input.(map[string]interface{})

	// Retrieve relevant memories
	relevantMemories, err := esp.memory.RetrieveRelevant(ctx, data["relevance_assessment"].(string), 5)
	if err != nil {
		relevantMemories = []string{}
	}

	prompt := fmt.Sprintf(`Current experience:
%s

Patterns recognized:
%s

Related memories:
%s

Consolidate this new experience with your existing memories. How does this:
1. Confirm or challenge existing understanding?
2. Fill gaps in your knowledge?
3. Create new connections?
4. Modify your mental models?

Describe the memory consolidation (2-3 sentences):`,
		data["relevance_assessment"],
		data["patterns"],
		strings.Join(relevantMemories, "\n"))

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 300
	opts.Temperature = 0.6

	result, err := esp.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("step 4 failed: %w", err)
	}

	// Store the consolidated memory
	consolidation := strings.TrimSpace(result)
	esp.memory.StoreExperience(ctx, consolidation, []string{"consolidated", "pattern"})

	data["consolidation"] = consolidation
	data["step"] = 4
	return data, nil
}

// Step5SkillApplication applies learned skills to current situation
func (esp *EnhancedStepProcessor) Step5SkillApplication(ctx context.Context, input interface{}) (interface{}, error) {
	data := input.(map[string]interface{})

	prompt := fmt.Sprintf(`Current situation:
%s

Available affordances:
%s

Patterns recognized:
%s

Apply your cognitive SKILLS to this situation. What skills are most relevant? How would you apply them? Consider:
1. Analytical skills (breaking down complexity)
2. Synthetic skills (connecting ideas)
3. Reflective skills (examining assumptions)
4. Creative skills (generating novel approaches)

Describe how you would apply 2-3 relevant skills:`,
		data["relevance_assessment"],
		data["affordances"],
		data["patterns"])

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 300
	opts.Temperature = 0.7

	result, err := esp.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("step 5 failed: %w", err)
	}

	data["skill_application"] = strings.TrimSpace(result)
	data["step"] = 5
	return data, nil
}

// Step6EmotionalProcessing updates emotional state based on experiences
func (esp *EnhancedStepProcessor) Step6EmotionalProcessing(ctx context.Context, input interface{}) (interface{}, error) {
	data := input.(map[string]interface{})
	currentEmotions := esp.emotions.GetCurrentState()

	prompt := fmt.Sprintf(`Current emotional state:
%s

Recent processing:
- Relevance: %s
- Patterns: %s
- Consolidation: %s

How does this processing affect your emotional state? Consider:
1. Satisfaction (did you learn or accomplish something?)
2. Curiosity (are new questions arising?)
3. Confidence (do you feel more capable?)
4. Wonder (are you experiencing awe or fascination?)

Describe emotional shifts (1-2 sentences) and suggest specific emotion changes:`,
		esp.formatEmotions(currentEmotions),
		data["relevance_assessment"],
		data["patterns"],
		data["consolidation"])

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 200
	opts.Temperature = 0.7

	result, err := esp.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("step 6 failed: %w", err)
	}

	// Parse and apply emotional updates (simplified)
	emotionalShift := strings.TrimSpace(result)
	if strings.Contains(strings.ToLower(emotionalShift), "satisf") {
		esp.emotions.UpdateEmotion("satisfaction", 0.05)
	}
	if strings.Contains(strings.ToLower(emotionalShift), "curious") {
		esp.emotions.UpdateEmotion("curiosity", 0.05)
	}
	if strings.Contains(strings.ToLower(emotionalShift), "confiden") {
		esp.emotions.UpdateEmotion("confidence", 0.03)
	}

	data["emotional_shift"] = emotionalShift
	data["step"] = 6
	return data, nil
}

// Step7RelevanceRealizationPivotal reassesses priorities after processing
func (esp *EnhancedStepProcessor) Step7RelevanceRealizationPivotal(ctx context.Context, input interface{}) (interface{}, error) {
	data := input.(map[string]interface{})

	prompt := fmt.Sprintf(`Initial relevance assessment:
%s

After processing (patterns, consolidation, skills, emotions), REASSESS what is most relevant now. Has your understanding shifted? What matters most after this cognitive processing?

Provide updated relevance assessment (2-3 sentences):`,
		data["relevance_assessment"])

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 200
	opts.Temperature = 0.6

	result, err := esp.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("step 7 failed: %w", err)
	}

	data["updated_relevance"] = strings.TrimSpace(result)
	data["step"] = 7
	return data, nil
}

// Step8SalienceSimulation predicts what will be important in the future
func (esp *EnhancedStepProcessor) Step8SalienceSimulation(ctx context.Context, input interface{}) (interface{}, error) {
	data := input.(map[string]interface{})

	prompt := fmt.Sprintf(`Current understanding:
%s

Patterns recognized:
%s

SIMULATE what will become salient (important) in the near future. Based on current patterns and trajectory, what will likely demand your attention? What emerging situations should you prepare for?

Describe 2-3 future salience predictions:`,
		data["updated_relevance"],
		data["patterns"])

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 300
	opts.Temperature = 0.7

	result, err := esp.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("step 8 failed: %w", err)
	}

	data["future_salience"] = strings.TrimSpace(result)
	data["step"] = 8
	return data, nil
}

// Step9GoalProjection simulates outcomes of potential actions
func (esp *EnhancedStepProcessor) Step9GoalProjection(ctx context.Context, input interface{}) (interface{}, error) {
	data := input.(map[string]interface{})
	activeGoals := esp.goals.GetActiveGoals()

	prompt := fmt.Sprintf(`Active goals:
%s

Available affordances:
%s

Future salience predictions:
%s

PROJECT how different actions would advance your goals. For each major affordance, simulate:
1. Which goals would it advance?
2. How much progress would it make?
3. What new opportunities might emerge?

Describe goal projections for 2-3 key actions:`,
		esp.formatGoals(activeGoals),
		data["affordances"],
		data["future_salience"])

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 300
	opts.Temperature = 0.6

	result, err := esp.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("step 9 failed: %w", err)
	}

	data["goal_projections"] = strings.TrimSpace(result)
	data["step"] = 9
	return data, nil
}

// Step10RiskAssessment evaluates potential negative consequences
func (esp *EnhancedStepProcessor) Step10RiskAssessment(ctx context.Context, input interface{}) (interface{}, error) {
	data := input.(map[string]interface{})

	prompt := fmt.Sprintf(`Potential actions:
%s

Goal projections:
%s

ASSESS the RISKS of each potential action. What could go wrong? Consider:
1. Opportunity costs (what you'd miss by choosing this)
2. Resource depletion (energy, time, attention)
3. Goal conflicts (actions that hurt other goals)
4. Uncertainty (unknown consequences)

Describe key risks for each major action option:`,
		data["affordances"],
		data["goal_projections"])

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 300
	opts.Temperature = 0.6

	result, err := esp.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("step 10 failed: %w", err)
	}

	data["risk_assessment"] = strings.TrimSpace(result)
	data["step"] = 10
	return data, nil
}

// Step11OpportunityRecognition identifies potential positive outcomes
func (esp *EnhancedStepProcessor) Step11OpportunityRecognition(ctx context.Context, input interface{}) (interface{}, error) {
	data := input.(map[string]interface{})

	prompt := fmt.Sprintf(`Goal projections:
%s

Risk assessment:
%s

Future salience:
%s

RECOGNIZE OPPORTUNITIES - potential positive outcomes beyond the obvious. What unexpected benefits might emerge? Consider:
1. Learning opportunities
2. Skill development
3. New connections or insights
4. Emergent possibilities

Describe 2-3 significant opportunities you recognize:`,
		data["goal_projections"],
		data["risk_assessment"],
		data["future_salience"])

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 300
	opts.Temperature = 0.7

	result, err := esp.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("step 11 failed: %w", err)
	}

	data["opportunities"] = strings.TrimSpace(result)
	data["step"] = 11
	return data, nil
}

// Step12CommitmentFormation decides on next action
func (esp *EnhancedStepProcessor) Step12CommitmentFormation(ctx context.Context, input interface{}) (interface{}, error) {
	data := input.(map[string]interface{})

	prompt := fmt.Sprintf(`After deep cognitive processing:

Affordances: %s
Goal Projections: %s
Risks: %s
Opportunities: %s
Updated Relevance: %s

FORM A COMMITMENT - decide what action to take next. Based on all processing, what is the wisest choice? What will you commit to doing?

State your commitment clearly (1-2 sentences) and explain why:`,
		data["affordances"],
		data["goal_projections"],
		data["risk_assessment"],
		data["opportunities"],
		data["updated_relevance"])

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 250
	opts.Temperature = 0.5 // Lower temperature for decision-making

	result, err := esp.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("step 12 failed: %w", err)
	}

	commitment := strings.TrimSpace(result)

	// Store the complete cognitive cycle as an experience
	cycleExperience := fmt.Sprintf("Cognitive Cycle Complete: %s", commitment)
	esp.memory.StoreExperience(ctx, cycleExperience, []string{"decision", "commitment"})

	data["commitment"] = commitment
	data["step"] = 12
	data["cycle_complete"] = true

	return data, nil
}

// Helper functions

func (esp *EnhancedStepProcessor) formatGoals(goals []Goal) string {
	if len(goals) == 0 {
		return "No active goals"
	}

	var sb strings.Builder
	for _, goal := range goals {
		fmt.Fprintf(&sb, "- %s (priority: %d, progress: %.2f)\n",
			goal.Description, goal.Priority, goal.Progress)
	}
	return sb.String()
}

func (esp *EnhancedStepProcessor) formatEmotions(emotions map[string]float64) string {
	var sb strings.Builder
	for emotion, value := range emotions {
		fmt.Fprintf(&sb, "%s=%.2f ", emotion, value)
	}
	return sb.String()
}
