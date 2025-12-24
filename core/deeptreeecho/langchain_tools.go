package deeptreeecho

import (
	"context"
	"fmt"
	"strings"
)

// CognitiveTool represents a tool that can be used by the ReasoningManager
// This interface is compatible with langchaingo tools.Tool interface
type CognitiveTool interface {
	Name() string
	Description() string
	Call(ctx context.Context, input string) (string, error)
}

// SkillLearningTool wraps the SkillLearningSystem for use as a cognitive tool
type SkillLearningTool struct {
	skillSystem *SkillLearningSystem
}

// NewSkillLearningTool creates a new skill learning tool
func NewSkillLearningTool(sls *SkillLearningSystem) *SkillLearningTool {
	return &SkillLearningTool{
		skillSystem: sls,
	}
}

func (t *SkillLearningTool) Name() string {
	return "SkillLearner"
}

func (t *SkillLearningTool) Description() string {
	return "Use this tool to learn new skills or consider practicing skills. " +
		"Input format: 'consider:<skill_name>' to consider learning a skill."
}

func (t *SkillLearningTool) Call(ctx context.Context, input string) (string, error) {
	parts := strings.SplitN(input, ":", 2)
	command := strings.ToLower(strings.TrimSpace(parts[0]))

	switch command {
	case "consider":
		if len(parts) < 2 {
			return "", fmt.Errorf("skill name required for consider command")
		}
		skillName := strings.TrimSpace(parts[1])
		t.skillSystem.ConsiderSkill(skillName, 0.8)
		return fmt.Sprintf("Considering skill: %s", skillName), nil

	default:
		// Default to consider
		t.skillSystem.ConsiderSkill(input, 0.7)
		return fmt.Sprintf("Considering skill: %s", input), nil
	}
}

// DiscussionTool wraps the DiscussionAutonomySystem for use as a cognitive tool
type DiscussionTool struct {
	discussionSystem *DiscussionAutonomySystem
}

// NewDiscussionTool creates a new discussion tool
func NewDiscussionTool(das *DiscussionAutonomySystem) *DiscussionTool {
	return &DiscussionTool{
		discussionSystem: das,
	}
}

func (t *DiscussionTool) Name() string {
	return "DiscussionManager"
}

func (t *DiscussionTool) Description() string {
	return "Use this tool to manage discussions. " +
		"Input format: 'update:<topic>' to update interest in a topic."
}

func (t *DiscussionTool) Call(ctx context.Context, input string) (string, error) {
	parts := strings.SplitN(input, ":", 2)
	command := strings.ToLower(strings.TrimSpace(parts[0]))

	switch command {
	case "update":
		if len(parts) < 2 {
			return "", fmt.Errorf("topic required for update command")
		}
		topic := strings.TrimSpace(parts[1])
		t.discussionSystem.UpdateInterest(topic, 0.1)
		return fmt.Sprintf("Updated interest in topic: %s", topic), nil

	default:
		// Default to update interest
		t.discussionSystem.UpdateInterest(input, 0.1)
		return fmt.Sprintf("Updated interest in: %s", input), nil
	}
}

// WisdomTool wraps the WisdomSynthesis system for use as a cognitive tool
type WisdomTool struct {
	wisdomSystem *WisdomSynthesis
}

// NewWisdomTool creates a new wisdom tool
func NewWisdomTool(ws *WisdomSynthesis) *WisdomTool {
	return &WisdomTool{
		wisdomSystem: ws,
	}
}

func (t *WisdomTool) Name() string {
	return "WisdomOracle"
}

func (t *WisdomTool) Description() string {
	return "Use this tool to accumulate patterns for wisdom synthesis. " +
		"Input format: 'accumulate:<pattern>:<source>' to add a pattern."
}

func (t *WisdomTool) Call(ctx context.Context, input string) (string, error) {
	parts := strings.SplitN(input, ":", 2)
	command := strings.ToLower(strings.TrimSpace(parts[0]))

	switch command {
	case "accumulate":
		if len(parts) < 2 {
			return "", fmt.Errorf("pattern details required for accumulate command")
		}
		subParts := strings.SplitN(parts[1], ":", 2)
		pattern := strings.TrimSpace(subParts[0])
		source := "reasoning"
		if len(subParts) > 1 {
			source = strings.TrimSpace(subParts[1])
		}
		t.wisdomSystem.AccumulatePattern(pattern, source, 0.7, []string{})
		return fmt.Sprintf("Accumulated pattern from %s: %s", source, pattern), nil

	default:
		// Default to accumulate
		t.wisdomSystem.AccumulatePattern(input, "reasoning", 0.6, []string{})
		return fmt.Sprintf("Accumulated pattern: %s", input), nil
	}
}

// InterestTool wraps the InterestPatternSystem for use as a cognitive tool
type InterestTool struct {
	interestSystem *InterestPatternSystem
}

// NewInterestTool creates a new interest tool
func NewInterestTool(ips *InterestPatternSystem) *InterestTool {
	return &InterestTool{
		interestSystem: ips,
	}
}

func (t *InterestTool) Name() string {
	return "InterestTracker"
}

func (t *InterestTool) Description() string {
	return "Use this tool to evaluate and record interest engagement. " +
		"Input format: 'evaluate:<content>' to evaluate interest, 'engage:<content>' to record engagement."
}

func (t *InterestTool) Call(ctx context.Context, input string) (string, error) {
	parts := strings.SplitN(input, ":", 2)
	command := strings.ToLower(strings.TrimSpace(parts[0]))

	switch command {
	case "evaluate":
		if len(parts) < 2 {
			return "", fmt.Errorf("content required for evaluate command")
		}
		content := strings.TrimSpace(parts[1])
		interest := t.interestSystem.EvaluateInterest(content)
		return fmt.Sprintf("Interest level for '%s': %.2f", content, interest), nil

	case "engage":
		if len(parts) < 2 {
			return "", fmt.Errorf("content required for engage command")
		}
		content := strings.TrimSpace(parts[1])
		t.interestSystem.RecordEngagement(content, true)
		return fmt.Sprintf("Recorded engagement with: %s", content), nil

	default:
		// Default to evaluate
		interest := t.interestSystem.EvaluateInterest(input)
		return fmt.Sprintf("Interest level: %.2f", interest), nil
	}
}

// KnowledgeTool wraps the EchoDreamKnowledgeIntegration for use as a cognitive tool
type KnowledgeTool struct {
	knowledgeSystem *EchoDreamKnowledgeIntegration
}

// NewKnowledgeTool creates a new knowledge tool
func NewKnowledgeTool(edk *EchoDreamKnowledgeIntegration) *KnowledgeTool {
	return &KnowledgeTool{
		knowledgeSystem: edk,
	}
}

func (t *KnowledgeTool) Name() string {
	return "KnowledgeIntegrator"
}

func (t *KnowledgeTool) Description() string {
	return "Use this tool to add memories for knowledge integration. " +
		"Input format: 'memory:<content>:<importance>' to add a memory."
}

func (t *KnowledgeTool) Call(ctx context.Context, input string) (string, error) {
	parts := strings.SplitN(input, ":", 2)
	command := strings.ToLower(strings.TrimSpace(parts[0]))

	switch command {
	case "memory":
		if len(parts) < 2 {
			return "", fmt.Errorf("content required for memory command")
		}
		subParts := strings.SplitN(parts[1], ":", 2)
		content := strings.TrimSpace(subParts[0])
		importance := 0.5
		memID := t.knowledgeSystem.AddMemory(content, importance, []string{"reasoning"})
		return fmt.Sprintf("Added memory %s: %s", memID, content), nil

	default:
		// Default to add memory
		memID := t.knowledgeSystem.AddMemory(input, 0.5, []string{"reasoning"})
		return fmt.Sprintf("Added memory %s", memID), nil
	}
}

// GoalTool wraps the EchobeatsScheduler for goal management
type GoalTool struct {
	scheduler *EchobeatsScheduler
}

// NewGoalTool creates a new goal tool
func NewGoalTool(es *EchobeatsScheduler) *GoalTool {
	return &GoalTool{
		scheduler: es,
	}
}

func (t *GoalTool) Name() string {
	return "GoalManager"
}

func (t *GoalTool) Description() string {
	return "Use this tool to create and manage cognitive goals. " +
		"Input format: 'create:<description>' to create a goal."
}

func (t *GoalTool) Call(ctx context.Context, input string) (string, error) {
	parts := strings.SplitN(input, ":", 2)
	command := strings.ToLower(strings.TrimSpace(parts[0]))

	switch command {
	case "create":
		if len(parts) < 2 {
			return "", fmt.Errorf("description required for create command")
		}
		description := strings.TrimSpace(parts[1])
		goalID := t.scheduler.AddGoal(description, 0.5)
		return fmt.Sprintf("Created goal %s: %s", goalID, description), nil

	default:
		// Default to create
		goalID := t.scheduler.AddGoal(input, 0.5)
		return fmt.Sprintf("Created goal %s: %s", goalID, input), nil
	}
}

// CreateAllCognitiveTools creates all cognitive tools from the subsystems
func CreateAllCognitiveTools(
	skillLearning *SkillLearningSystem,
	discussionAutonomy *DiscussionAutonomySystem,
	wisdomSynthesis *WisdomSynthesis,
	interestPatterns *InterestPatternSystem,
	echoDream *EchoDreamKnowledgeIntegration,
	echobeatsScheduler *EchobeatsScheduler,
) []CognitiveTool {
	return []CognitiveTool{
		NewSkillLearningTool(skillLearning),
		NewDiscussionTool(discussionAutonomy),
		NewWisdomTool(wisdomSynthesis),
		NewInterestTool(interestPatterns),
		NewKnowledgeTool(echoDream),
		NewGoalTool(echobeatsScheduler),
	}
}
