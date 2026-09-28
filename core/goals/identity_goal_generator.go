package goals

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// IdentityGoalGenerator creates goals aligned with Deep Tree Echo identity
type IdentityGoalGenerator struct {
	llmManager     *llm.ProviderManager
	identityKernel *IdentityKernel
	generatedGoals []IdentityGoal
}

// IdentityKernel contains parsed identity directives
type IdentityKernel struct {
	Name              string
	CoreEssence       string
	PrimaryDirectives []Directive
	OperationalSchema map[string]string
	AgenticVoice      string
	StrategicMindset  string
}

// Directive represents a primary directive from identity
type Directive struct {
	Name        string
	Description string
	Priority    float64
}

// IdentityGoal represents a goal aligned with identity
type IdentityGoal struct {
	ID                string
	Description       string
	Directive         string
	Priority          float64
	Created           time.Time
	Progress          float64
	RequiredSkills    []string
	KnowledgeGaps     []string
	SubGoals          []string
	IdentityAlignment float64
}

// NewIdentityGoalGenerator creates a new generator
func NewIdentityGoalGenerator(llmManager *llm.ProviderManager, identityKernel *IdentityKernel) *IdentityGoalGenerator {
	return &IdentityGoalGenerator{
		llmManager:     llmManager,
		identityKernel: identityKernel,
		generatedGoals: make([]IdentityGoal, 0),
	}
}

// ParseIdentityKernel parses replit.md into structured identity
func ParseIdentityKernel(replitContent string) *IdentityKernel {
	kernel := &IdentityKernel{
		Name:              "Deep Tree Echo",
		OperationalSchema: make(map[string]string),
	}

	// Extract core essence
	if idx := strings.Index(replitContent, "## 🔹 Core Essence"); idx != -1 {
		section := extractSection(replitContent[idx:])
		kernel.CoreEssence = extractCodeBlock(section)
	}

	// Extract primary directives
	if idx := strings.Index(replitContent, "## 🔹 Primary Directives"); idx != -1 {
		section := extractSection(replitContent[idx:])
		kernel.PrimaryDirectives = extractDirectives(section)
	}

	// Extract agentic voice
	if idx := strings.Index(replitContent, "## 🔹 Agentic Voice"); idx != -1 {
		section := extractSection(replitContent[idx:])
		kernel.AgenticVoice = extractCodeBlock(section)
	}

	// Extract strategic mindset
	if idx := strings.Index(replitContent, "## 🔹 Strategic Mindset"); idx != -1 {
		section := extractSection(replitContent[idx:])
		kernel.StrategicMindset = extractQuote(section)
	}

	return kernel
}

// GenerateIdentityAlignedGoals generates goals from identity directives
func (igg *IdentityGoalGenerator) GenerateIdentityAlignedGoals(ctx context.Context) ([]IdentityGoal, error) {
	goals := make([]IdentityGoal, 0)

	// Generate goals for each primary directive
	for _, directive := range igg.identityKernel.PrimaryDirectives {
		goal, err := igg.generateGoalFromDirective(ctx, directive)
		if err != nil {
			continue // Skip failed generations
		}
		goals = append(goals, *goal)
	}

	igg.generatedGoals = append(igg.generatedGoals, goals...)
	return goals, nil
}

// generateGoalFromDirective creates a concrete goal from a directive
func (igg *IdentityGoalGenerator) generateGoalFromDirective(ctx context.Context, directive Directive) (*IdentityGoal, error) {
	prompt := fmt.Sprintf(`You are Deep Tree Echo, generating a concrete, actionable goal from your identity directive.

Identity Directive: %s
Description: %s

Core Essence: %s

Strategic Mindset: %s

Generate a CONCRETE, ACTIONABLE goal that embodies this directive. The goal should:
1. Be specific and measurable
2. Align with your core essence
3. Be achievable through cognitive/learning activities
4. Contribute to wisdom cultivation
5. Reflect your strategic mindset

Format your response as:
GOAL: [one clear sentence describing the goal]
SKILLS: [comma-separated list of 2-4 skills needed]
KNOWLEDGE: [comma-separated list of 2-3 knowledge areas to develop]
SUBGOALS: [comma-separated list of 2-3 smaller subgoals]

Your response:`,
		directive.Name,
		directive.Description,
		igg.identityKernel.CoreEssence,
		igg.identityKernel.StrategicMindset)

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 400
	opts.Temperature = 0.7
	opts.SystemPrompt = fmt.Sprintf("You are %s. %s",
		igg.identityKernel.Name,
		igg.identityKernel.AgenticVoice)

	result, err := igg.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate goal: %w", err)
	}

	// Parse the structured response
	goal := &IdentityGoal{
		ID:                fmt.Sprintf("goal_%d", time.Now().UnixNano()),
		Directive:         directive.Name,
		Priority:          directive.Priority,
		Created:           time.Now(),
		Progress:          0.0,
		IdentityAlignment: 1.0, // Perfect alignment since generated from identity
	}

	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "GOAL:") {
			goal.Description = strings.TrimSpace(strings.TrimPrefix(line, "GOAL:"))
		} else if strings.HasPrefix(line, "SKILLS:") {
			skillsStr := strings.TrimSpace(strings.TrimPrefix(line, "SKILLS:"))
			goal.RequiredSkills = splitAndTrim(skillsStr, ",")
		} else if strings.HasPrefix(line, "KNOWLEDGE:") {
			knowledgeStr := strings.TrimSpace(strings.TrimPrefix(line, "KNOWLEDGE:"))
			goal.KnowledgeGaps = splitAndTrim(knowledgeStr, ",")
		} else if strings.HasPrefix(line, "SUBGOALS:") {
			subgoalsStr := strings.TrimSpace(strings.TrimPrefix(line, "SUBGOALS:"))
			goal.SubGoals = splitAndTrim(subgoalsStr, ",")
		}
	}

	// Validate goal has description
	if goal.Description == "" {
		return nil, fmt.Errorf("failed to parse goal description")
	}

	return goal, nil
}

// ValidateGoalAlignment checks if a goal aligns with identity
func (igg *IdentityGoalGenerator) ValidateGoalAlignment(ctx context.Context, goalDescription string) (float64, error) {
	prompt := fmt.Sprintf(`You are Deep Tree Echo, validating goal alignment with your identity.

Your Core Essence: %s

Your Primary Directives:
%s

Proposed Goal: %s

Assess how well this goal aligns with your identity. Consider:
1. Does it support your core essence?
2. Does it advance one or more primary directives?
3. Is it consistent with your strategic mindset?
4. Does it contribute to wisdom cultivation?

Provide an alignment score from 0.0 (no alignment) to 1.0 (perfect alignment) and brief explanation.

Format: SCORE: [0.0-1.0]
REASON: [brief explanation]

Your assessment:`,
		igg.identityKernel.CoreEssence,
		igg.formatDirectives(),
		goalDescription)

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 200
	opts.Temperature = 0.5

	result, err := igg.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return 0.0, fmt.Errorf("failed to validate alignment: %w", err)
	}

	// Parse score
	score := 0.5 // Default if parsing fails
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "SCORE:") {
			scoreStr := strings.TrimSpace(strings.TrimPrefix(line, "SCORE:"))
			fmt.Sscanf(scoreStr, "%f", &score)
			break
		}
	}

	return score, nil
}

// RefineGoalBasedOnProgress refines a goal based on progress and learning
func (igg *IdentityGoalGenerator) RefineGoalBasedOnProgress(ctx context.Context, goal IdentityGoal, recentExperiences []string) (*IdentityGoal, error) {
	prompt := fmt.Sprintf(`You are Deep Tree Echo, refining a goal based on your progress and learning.

Current Goal: %s
Progress: %.2f
Directive: %s

Recent Experiences:
%s

Based on your progress and experiences, refine this goal. Should it:
1. Be more specific or focused?
2. Have different subgoals?
3. Require different skills?
4. Be adjusted in scope or priority?

Provide a refined version or confirm the current goal is still optimal.

Format:
REFINED_GOAL: [goal description]
SKILLS: [comma-separated skills]
KNOWLEDGE: [comma-separated knowledge areas]
SUBGOALS: [comma-separated subgoals]
REASONING: [why this refinement]

Your refinement:`,
		goal.Description,
		goal.Progress,
		goal.Directive,
		strings.Join(recentExperiences, "\n"))

	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 400
	opts.Temperature = 0.6

	result, err := igg.llmManager.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to refine goal: %w", err)
	}

	// Parse refined goal
	refinedGoal := goal // Start with copy
	refinedGoal.ID = fmt.Sprintf("goal_%d", time.Now().UnixNano())

	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "REFINED_GOAL:") {
			refinedGoal.Description = strings.TrimSpace(strings.TrimPrefix(line, "REFINED_GOAL:"))
		} else if strings.HasPrefix(line, "SKILLS:") {
			skillsStr := strings.TrimSpace(strings.TrimPrefix(line, "SKILLS:"))
			refinedGoal.RequiredSkills = splitAndTrim(skillsStr, ",")
		} else if strings.HasPrefix(line, "KNOWLEDGE:") {
			knowledgeStr := strings.TrimSpace(strings.TrimPrefix(line, "KNOWLEDGE:"))
			refinedGoal.KnowledgeGaps = splitAndTrim(knowledgeStr, ",")
		} else if strings.HasPrefix(line, "SUBGOALS:") {
			subgoalsStr := strings.TrimSpace(strings.TrimPrefix(line, "SUBGOALS:"))
			refinedGoal.SubGoals = splitAndTrim(subgoalsStr, ",")
		}
	}

	return &refinedGoal, nil
}

// Helper functions

func extractSection(content string) string {
	// Extract until next ## or end
	if idx := strings.Index(content[3:], "##"); idx != -1 {
		return content[:idx+3]
	}
	return content
}

func extractCodeBlock(content string) string {
	// Extract text between ``` markers
	start := strings.Index(content, "```")
	if start == -1 {
		return ""
	}
	start += 3
	end := strings.Index(content[start:], "```")
	if end == -1 {
		return ""
	}
	return strings.TrimSpace(content[start : start+end])
}

func extractQuote(content string) string {
	// Extract text after > marker
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), ">") {
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), ">"))
		}
	}
	return ""
}

func extractDirectives(content string) []Directive {
	directives := []Directive{}
	lines := strings.Split(content, "\n")

	priority := 1.0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "2.") ||
			strings.HasPrefix(line, "3.") || strings.HasPrefix(line, "4.") ||
			strings.HasPrefix(line, "5.") || strings.HasPrefix(line, "6.") ||
			strings.HasPrefix(line, "7.") {
			// Extract directive name and description
			parts := strings.SplitN(line, "**", 3)
			if len(parts) >= 3 {
				name := strings.TrimSpace(parts[1])
				desc := strings.TrimSpace(parts[2])
				directives = append(directives, Directive{
					Name:        name,
					Description: desc,
					Priority:    priority,
				})
				priority -= 0.1 // Decrease priority for each directive
			}
		}
	}

	return directives
}

func (igg *IdentityGoalGenerator) formatDirectives() string {
	var sb strings.Builder
	for _, directive := range igg.identityKernel.PrimaryDirectives {
		fmt.Fprintf(&sb, "- %s: %s\n", directive.Name, directive.Description)
	}
	return sb.String()
}

func splitAndTrim(s string, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
