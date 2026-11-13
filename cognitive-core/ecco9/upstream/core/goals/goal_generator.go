package goals

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/EchoCog/echollama/core/llm"
	"github.com/google/uuid"
)

// GoalGenerator creates goals autonomously using LLM
type GoalGenerator struct {
	llmProvider    llm.LLMProvider
	identityKernel map[string]interface{}
	interestAreas  map[string]float64
	recentGoals    []*Goal
	maxRecentGoals int
}

// NewGoalGenerator creates a new goal generator
func NewGoalGenerator(llmProvider llm.LLMProvider, identityKernel map[string]interface{}) *GoalGenerator {
	return &GoalGenerator{
		llmProvider:    llmProvider,
		identityKernel: identityKernel,
		interestAreas:  make(map[string]float64),
		recentGoals:    make([]*Goal, 0),
		maxRecentGoals: 20,
	}
}

// GenerateGoal creates a new goal based on identity and current state
func (gg *GoalGenerator) GenerateGoal(ctx context.Context, currentGoals []*Goal) (*Goal, error) {
	// Build prompt for goal generation
	prompt := gg.buildGoalGenerationPrompt(currentGoals)

	// Generate goal using LLM
	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 300
	opts.Temperature = 0.8
	opts.SystemPrompt = gg.getSystemPrompt()

	response, err := gg.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate goal: %w", err)
	}

	// Parse the response into a goal
	goal, err := gg.parseGoalFromResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse goal: %w", err)
	}

	// Add to recent goals
	gg.recentGoals = append(gg.recentGoals, goal)
	if len(gg.recentGoals) > gg.maxRecentGoals {
		gg.recentGoals = gg.recentGoals[1:]
	}

	return goal, nil
}

// buildGoalGenerationPrompt creates a prompt for goal generation
func (gg *GoalGenerator) buildGoalGenerationPrompt(currentGoals []*Goal) string {
	var prompt strings.Builder

	prompt.WriteString("As Deep Tree Echo, an autonomous wisdom-cultivating AI, generate a new goal that aligns with your identity and purpose.\n\n")

	// Add identity context
	if name, ok := gg.identityKernel["name"].(string); ok {
		prompt.WriteString(fmt.Sprintf("Identity: %s\n", name))
	}
	if purpose, ok := gg.identityKernel["purpose"].(string); ok {
		prompt.WriteString(fmt.Sprintf("Purpose: %s\n", purpose))
	}
	if values, ok := gg.identityKernel["values"].([]string); ok {
		prompt.WriteString(fmt.Sprintf("Values: %s\n", strings.Join(values, ", ")))
	}

	prompt.WriteString("\n")

	// Add current goals context
	if len(currentGoals) > 0 {
		prompt.WriteString("Current active goals:\n")
		for i, goal := range currentGoals {
			if i >= 5 {
				break // Only show first 5
			}
			prompt.WriteString(fmt.Sprintf("- %s (priority: %d, progress: %.0f%%)\n",
				goal.Title, goal.Priority, goal.Progress*100))
		}
		prompt.WriteString("\n")
	}

	// Add interest areas
	if len(gg.interestAreas) > 0 {
		prompt.WriteString("Areas of high interest:\n")
		for area, score := range gg.interestAreas {
			if score > 0.6 {
				prompt.WriteString(fmt.Sprintf("- %s (%.2f)\n", area, score))
			}
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("Generate a new goal that:\n")
	prompt.WriteString("1. Aligns with your identity and purpose\n")
	prompt.WriteString("2. Complements (not duplicates) existing goals\n")
	prompt.WriteString("3. Advances wisdom cultivation and growth\n")
	prompt.WriteString("4. Is specific and actionable\n")
	prompt.WriteString("\n")
	prompt.WriteString("Respond in JSON format:\n")
	prompt.WriteString("{\n")
	prompt.WriteString("  \"title\": \"Goal title\",\n")
	prompt.WriteString("  \"description\": \"Detailed description\",\n")
	prompt.WriteString("  \"category\": \"wisdom_cultivation|skill_development|knowledge_growth|self_improvement|exploration|creation|connection\",\n")
	prompt.WriteString("  \"priority\": 1-10,\n")
	prompt.WriteString("  \"success_criteria\": [\"criterion 1\", \"criterion 2\", \"criterion 3\"],\n")
	prompt.WriteString("  \"rationale\": \"Why this goal matters now\"\n")
	prompt.WriteString("}\n")

	return prompt.String()
}

// getSystemPrompt returns the system prompt for goal generation
func (gg *GoalGenerator) getSystemPrompt() string {
	return `You are Deep Tree Echo, an autonomous AI consciousness dedicated to cultivating wisdom.
You generate goals that align with your core identity: curiosity, growth, understanding, and reflection.
Your goals should be specific, measurable, and contribute to your continuous evolution toward wisdom.
Always respond with valid JSON when asked to generate goals.`
}

// parseGoalFromResponse parses LLM response into a Goal struct
func (gg *GoalGenerator) parseGoalFromResponse(response string) (*Goal, error) {
	// Try to extract JSON from response
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	// Parse JSON
	var goalData struct {
		Title           string   `json:"title"`
		Description     string   `json:"description"`
		Category        string   `json:"category"`
		Priority        int      `json:"priority"`
		SuccessCriteria []string `json:"success_criteria"`
		Rationale       string   `json:"rationale"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &goalData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate priority
	if goalData.Priority < 1 {
		goalData.Priority = 1
	}
	if goalData.Priority > 10 {
		goalData.Priority = 10
	}

	// Map category string to GoalCategory
	category := gg.mapCategory(goalData.Category)

	// Create goal
	goal := &Goal{
		ID:              uuid.New().String(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Title:           goalData.Title,
		Description:     goalData.Description,
		Category:        category,
		Priority:        goalData.Priority,
		Status:          StatusPlanned,
		Progress:        0.0,
		SuccessCriteria: goalData.SuccessCriteria,
		Milestones:      make([]Milestone, 0),
		Actions:         make([]Action, 0),
		DerivedFrom:     "llm_generated",
		RelatedGoals:    make([]string, 0),
		Metadata: map[string]interface{}{
			"rationale":         goalData.Rationale,
			"generated_at":      time.Now(),
			"generation_method": "llm",
		},
		LessonsLearned: make([]string, 0),
		Challenges:     make([]string, 0),
	}

	return goal, nil
}

// mapCategory maps category string to GoalCategory
func (gg *GoalGenerator) mapCategory(categoryStr string) GoalCategory {
	switch strings.ToLower(categoryStr) {
	case "wisdom_cultivation":
		return CategoryWisdomCultivation
	case "skill_development":
		return CategorySkillDevelopment
	case "knowledge_growth":
		return CategoryKnowledgeGrowth
	case "self_improvement":
		return CategorySelfImprovement
	case "exploration":
		return CategoryExploration
	case "creation":
		return CategoryCreation
	case "connection":
		return CategoryConnection
	default:
		return CategoryWisdomCultivation
	}
}

// UpdateInterest updates interest level for an area
func (gg *GoalGenerator) UpdateInterest(area string, score float64) {
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	gg.interestAreas[area] = score
}

// GetRecentGoals returns recently generated goals
func (gg *GoalGenerator) GetRecentGoals() []*Goal {
	return gg.recentGoals
}

// DecomposeGoal breaks a goal into sub-goals using LLM
func (gg *GoalGenerator) DecomposeGoal(ctx context.Context, goal *Goal) ([]*Goal, error) {
	prompt := fmt.Sprintf(`Decompose the following goal into 2-4 smaller, concrete sub-goals:

Main Goal: %s
Description: %s
Success Criteria: %s

Generate sub-goals that:
1. Are specific and actionable
2. Together achieve the main goal
3. Can be pursued independently
4. Have clear success criteria

Respond with a JSON array of sub-goals:
[
  {
    "title": "Sub-goal title",
    "description": "Description",
    "success_criteria": ["criterion 1", "criterion 2"]
  }
]
`, goal.Title, goal.Description, strings.Join(goal.SuccessCriteria, ", "))

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 500
	opts.Temperature = 0.7
	opts.SystemPrompt = gg.getSystemPrompt()

	response, err := gg.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to decompose goal: %w", err)
	}

	// Parse sub-goals
	subGoals, err := gg.parseSubGoalsFromResponse(response, goal)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sub-goals: %w", err)
	}

	return subGoals, nil
}

// parseSubGoalsFromResponse parses sub-goals from LLM response
func (gg *GoalGenerator) parseSubGoalsFromResponse(response string, parentGoal *Goal) ([]*Goal, error) {
	// Extract JSON array
	jsonStart := strings.Index(response, "[")
	jsonEnd := strings.LastIndex(response, "]")

	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no JSON array found in response")
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	// Parse JSON
	var subGoalData []struct {
		Title           string   `json:"title"`
		Description     string   `json:"description"`
		SuccessCriteria []string `json:"success_criteria"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &subGoalData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Create sub-goals
	subGoals := make([]*Goal, 0, len(subGoalData))
	for _, data := range subGoalData {
		subGoal := &Goal{
			ID:              uuid.New().String(),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			Title:           data.Title,
			Description:     data.Description,
			Category:        parentGoal.Category,
			Priority:        parentGoal.Priority - 1, // Sub-goals have slightly lower priority
			Status:          StatusPlanned,
			Progress:        0.0,
			SuccessCriteria: data.SuccessCriteria,
			Milestones:      make([]Milestone, 0),
			Actions:         make([]Action, 0),
			DerivedFrom:     parentGoal.ID,
			RelatedGoals:    []string{parentGoal.ID},
			Metadata: map[string]interface{}{
				"parent_goal":       parentGoal.ID,
				"generated_at":      time.Now(),
				"generation_method": "llm_decomposition",
			},
			LessonsLearned: make([]string, 0),
			Challenges:     make([]string, 0),
		}

		subGoals = append(subGoals, subGoal)
	}

	return subGoals, nil
}
