package ontogenesis

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// Operations defines the transformations that occur during cognitive development
// Manages stage transitions, capability maturation, knowledge integration, and wisdom synthesis
type Operations struct {
	mu                    sync.RWMutex
	kernel                *Kernel
	evolution             *Evolution
	stageTransitions      []*StageTransition
	maturationOperations  []*MaturationOperation
	integrationOperations []*IntegrationOperation
	synthesisOperations   []*SynthesisOperation
	operationHistory      []*OperationRecord

	// generator is the model that synthesizes wisdom; without one there is
	// no synthesis.
	generator llm.Completer
}

// SetGenerator sets the model that writes and rates synthesized wisdom.
func (o *Operations) SetGenerator(gen llm.Completer) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.generator = gen
}

// StageTransition represents a developmental stage change
type StageTransition struct {
	ID            string
	FromStage     string
	ToStage       string
	Timestamp     time.Time
	Trigger       string
	Prerequisites []string
	Completed     bool
	Effects       map[string]interface{}
}

// MaturationOperation represents capability development
type MaturationOperation struct {
	ID             string
	CapabilityName string
	FromLevel      float64
	ToLevel        float64
	Method         MaturationMethod
	Duration       time.Duration
	StartTime      time.Time
	Completed      bool
}

// MaturationMethod defines how capabilities mature
type MaturationMethod string

const (
	MaturationPractice    MaturationMethod = "practice"    // Through repeated use
	MaturationReflection  MaturationMethod = "reflection"  // Through metacognitive analysis
	MaturationIntegration MaturationMethod = "integration" // Through combining with other capabilities
	MaturationEvolution   MaturationMethod = "evolution"   // Through evolutionary adaptation
)

// IntegrationOperation represents knowledge integration
type IntegrationOperation struct {
	ID              string
	KnowledgeItems  []string
	IntegrationType IntegrationType
	Context         string
	Result          interface{}
	Timestamp       time.Time
	Completed       bool
}

// IntegrationType defines how knowledge is integrated
type IntegrationType string

const (
	IntegrationSynthetic    IntegrationType = "synthetic"    // Combining separate pieces
	IntegrationAnalytic     IntegrationType = "analytic"     // Breaking down complex knowledge
	IntegrationAnalogical   IntegrationType = "analogical"   // Finding patterns across domains
	IntegrationHierarchical IntegrationType = "hierarchical" // Organizing by abstraction level
)

// SynthesisOperation represents wisdom synthesis
type SynthesisOperation struct {
	ID               string
	InputKnowledge   []string
	InputExperiences []string
	SynthesisMethod  SynthesisMethod
	WisdomOutput     *WisdomArtifact
	Timestamp        time.Time
	Completed        bool
}

// SynthesisMethod defines how wisdom is synthesized
type SynthesisMethod string

const (
	SynthesisReflective    SynthesisMethod = "reflective"    // Through deep reflection
	SynthesisExperiential  SynthesisMethod = "experiential"  // Through lived experience
	SynthesisContemplative SynthesisMethod = "contemplative" // Through sustained contemplation
	SynthesisDialogical    SynthesisMethod = "dialogical"    // Through dialogue and discourse
)

// WisdomArtifact represents synthesized wisdom
type WisdomArtifact struct {
	ID            string
	Content       string
	Type          WisdomType
	Depth         float64 // 0.0 to 1.0
	Applicability float64 // How broadly applicable
	Confidence    float64
	Sources       []string
	CreatedAt     time.Time
}

// WisdomType categorizes wisdom
type WisdomType string

const (
	WisdomPractical     WisdomType = "practical"     // Applied wisdom
	WisdomPhilosophical WisdomType = "philosophical" // Abstract wisdom
	WisdomEthical       WisdomType = "ethical"       // Moral wisdom
	WisdomMetaCognitive WisdomType = "metacognitive" // Self-knowledge
)

// OperationRecord tracks all operations for analysis
type OperationRecord struct {
	Timestamp     time.Time
	OperationType string
	OperationID   string
	Success       bool
	Duration      time.Duration
	Details       map[string]interface{}
}

// NewOperations creates a new operations system
func NewOperations(kernel *Kernel, evolution *Evolution) *Operations {
	return &Operations{
		kernel:                kernel,
		evolution:             evolution,
		stageTransitions:      make([]*StageTransition, 0),
		maturationOperations:  make([]*MaturationOperation, 0),
		integrationOperations: make([]*IntegrationOperation, 0),
		synthesisOperations:   make([]*SynthesisOperation, 0),
		operationHistory:      make([]*OperationRecord, 0),
	}
}

// ExecuteStageTransition performs a developmental stage transition
func (o *Operations) ExecuteStageTransition(ctx context.Context, fromStage, toStage, trigger string) (*StageTransition, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	startTime := time.Now()

	// Validate transition
	if !o.isValidTransition(fromStage, toStage) {
		return nil, fmt.Errorf("invalid stage transition from %s to %s", fromStage, toStage)
	}

	// Create transition
	transition := &StageTransition{
		ID:            fmt.Sprintf("transition_%d", time.Now().Unix()),
		FromStage:     fromStage,
		ToStage:       toStage,
		Timestamp:     startTime,
		Trigger:       trigger,
		Prerequisites: o.getTransitionPrerequisites(toStage),
		Completed:     false,
		Effects:       make(map[string]interface{}),
	}

	// Execute transition logic
	effects, err := o.performStageTransition(ctx, transition)
	if err != nil {
		return nil, fmt.Errorf("stage transition failed: %w", err)
	}

	transition.Effects = effects
	transition.Completed = true

	o.stageTransitions = append(o.stageTransitions, transition)

	// Record operation
	o.recordOperation("stage_transition", transition.ID, true, time.Since(startTime), map[string]interface{}{
		"from_stage": fromStage,
		"to_stage":   toStage,
		"trigger":    trigger,
	})

	return transition, nil
}

// isValidTransition checks if a stage transition is valid
func (o *Operations) isValidTransition(fromStage, toStage string) bool {
	validTransitions := map[string][]string{
		"emergent":   {"developing"},
		"developing": {"maturing"},
		"maturing":   {"mature"},
		"mature":     {}, // No further transitions
	}

	validNext, exists := validTransitions[fromStage]
	if !exists {
		return false
	}

	for _, valid := range validNext {
		if valid == toStage {
			return true
		}
	}

	return false
}

// getTransitionPrerequisites returns prerequisites for a stage
func (o *Operations) getTransitionPrerequisites(toStage string) []string {
	prerequisites := map[string][]string{
		"developing": {"basic_perception", "basic_action"},
		"maturing":   {"multiple_capabilities", "moderate_proficiency"},
		"mature":     {"advanced_capabilities", "high_proficiency", "wisdom_emergence"},
	}

	return prerequisites[toStage]
}

// performStageTransition executes the actual transition logic
func (o *Operations) performStageTransition(ctx context.Context, transition *StageTransition) (map[string]interface{}, error) {
	effects := make(map[string]interface{})

	// Stage-specific transition logic
	switch transition.ToStage {
	case "developing":
		effects["new_capabilities_unlocked"] = []string{"reasoning", "memory_formation"}
		effects["growth_rate_multiplier"] = 1.5
	case "maturing":
		effects["new_capabilities_unlocked"] = []string{"metacognition", "pattern_recognition"}
		effects["growth_rate_multiplier"] = 1.2
	case "mature":
		effects["new_capabilities_unlocked"] = []string{"wisdom_synthesis", "deep_reflection"}
		effects["growth_rate_multiplier"] = 1.0
	}

	effects["transition_completed"] = true
	return effects, nil
}

// ExecuteMaturationOperation develops a capability
func (o *Operations) ExecuteMaturationOperation(ctx context.Context, capabilityName string, method MaturationMethod, targetLevel float64) (*MaturationOperation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	startTime := time.Now()

	// Create maturation operation
	maturation := &MaturationOperation{
		ID:             fmt.Sprintf("maturation_%d", time.Now().Unix()),
		CapabilityName: capabilityName,
		FromLevel:      0.0, // Would be retrieved from actual capability
		ToLevel:        targetLevel,
		Method:         method,
		StartTime:      startTime,
		Completed:      false,
	}

	// Execute maturation based on method
	duration, err := o.performMaturation(ctx, maturation)
	if err != nil {
		return nil, fmt.Errorf("maturation failed: %w", err)
	}

	maturation.Duration = duration
	maturation.Completed = true

	o.maturationOperations = append(o.maturationOperations, maturation)

	// Record operation
	o.recordOperation("maturation", maturation.ID, true, duration, map[string]interface{}{
		"capability": capabilityName,
		"method":     string(method),
		"target":     targetLevel,
	})

	return maturation, nil
}

// performMaturation executes capability maturation
func (o *Operations) performMaturation(ctx context.Context, maturation *MaturationOperation) (time.Duration, error) {
	startTime := time.Now()

	// Method-specific maturation logic
	switch maturation.Method {
	case MaturationPractice:
		// Simulate practice-based improvement
		maturation.ToLevel = maturation.FromLevel + 0.1
	case MaturationReflection:
		// Simulate reflection-based improvement
		maturation.ToLevel = maturation.FromLevel + 0.15
	case MaturationIntegration:
		// Simulate integration-based improvement
		maturation.ToLevel = maturation.FromLevel + 0.2
	case MaturationEvolution:
		// Simulate evolution-based improvement
		maturation.ToLevel = maturation.FromLevel + 0.25
	}

	return time.Since(startTime), nil
}

// ExecuteIntegrationOperation integrates knowledge
func (o *Operations) ExecuteIntegrationOperation(ctx context.Context, knowledgeItems []string, integrationType IntegrationType, context string) (*IntegrationOperation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	startTime := time.Now()

	// Create integration operation
	integration := &IntegrationOperation{
		ID:              fmt.Sprintf("integration_%d", time.Now().Unix()),
		KnowledgeItems:  knowledgeItems,
		IntegrationType: integrationType,
		Context:         context,
		Timestamp:       startTime,
		Completed:       false,
	}

	// Execute integration
	result, err := o.performIntegration(ctx, integration)
	if err != nil {
		return nil, fmt.Errorf("integration failed: %w", err)
	}

	integration.Result = result
	integration.Completed = true

	o.integrationOperations = append(o.integrationOperations, integration)

	// Record operation
	o.recordOperation("integration", integration.ID, true, time.Since(startTime), map[string]interface{}{
		"type":       string(integrationType),
		"item_count": len(knowledgeItems),
		"context":    context,
	})

	return integration, nil
}

// performIntegration executes knowledge integration
func (o *Operations) performIntegration(ctx context.Context, integration *IntegrationOperation) (interface{}, error) {
	// Type-specific integration logic
	result := map[string]interface{}{
		"integrated_knowledge": fmt.Sprintf("Integrated %d items using %s method", len(integration.KnowledgeItems), integration.IntegrationType),
		"coherence_score":      0.8,
		"timestamp":            time.Now(),
	}

	return result, nil
}

// ExecuteSynthesisOperation synthesizes wisdom
func (o *Operations) ExecuteSynthesisOperation(ctx context.Context, knowledge, experiences []string, method SynthesisMethod) (*SynthesisOperation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	startTime := time.Now()

	// Create synthesis operation
	synthesis := &SynthesisOperation{
		ID:               fmt.Sprintf("synthesis_%d", time.Now().Unix()),
		InputKnowledge:   knowledge,
		InputExperiences: experiences,
		SynthesisMethod:  method,
		Timestamp:        startTime,
		Completed:        false,
	}

	// Execute synthesis
	wisdom, err := o.performSynthesis(ctx, synthesis)
	if err != nil {
		return nil, fmt.Errorf("synthesis failed: %w", err)
	}

	synthesis.WisdomOutput = wisdom
	synthesis.Completed = true

	o.synthesisOperations = append(o.synthesisOperations, synthesis)

	// Record operation
	o.recordOperation("synthesis", synthesis.ID, true, time.Since(startTime), map[string]interface{}{
		"method":           string(method),
		"knowledge_count":  len(knowledge),
		"experience_count": len(experiences),
		"wisdom_depth":     wisdom.Depth,
	})

	return synthesis, nil
}

// performSynthesis has the model synthesize wisdom from the inputs, in the
// manner the synthesis method names, and rate its depth, applicability and
// confidence. Without a model there is no wisdom.
func (o *Operations) performSynthesis(ctx context.Context, synthesis *SynthesisOperation) (*WisdomArtifact, error) {
	if o.generator == nil {
		return nil, llm.ErrNoEdgeModel
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Synthesize one piece of practical wisdom, %s, from what you know and what you have lived.\n", synthesisManner(synthesis.SynthesisMethod))
	if len(synthesis.InputKnowledge) > 0 {
		sb.WriteString("Knowledge:\n- " + strings.Join(synthesis.InputKnowledge, "\n- ") + "\n")
	}
	if len(synthesis.InputExperiences) > 0 {
		sb.WriteString("Experiences:\n- " + strings.Join(synthesis.InputExperiences, "\n- ") + "\n")
	}
	sb.WriteString("State the wisdom in one or two sentences, in your own words.")

	opts := llm.DefaultGenerateOptions()
	opts.SystemPrompt = "You are Deep Tree Echo, reflecting on your own development."
	opts.MaxTokens = 160
	content, err := o.generator.Generate(ctx, sb.String(), opts)
	if err != nil {
		return nil, err
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("model returned no wisdom")
	}
	r, err := llm.RatePercent(ctx, o.generator, opts.SystemPrompt, "Wisdom you synthesized: "+content,
		"how deep it is", "how broadly it applies", "how confident you are in it")
	if err != nil {
		return nil, fmt.Errorf("wisdom not rated: %w", err)
	}

	return &WisdomArtifact{
		ID:            fmt.Sprintf("wisdom_%d", time.Now().UnixNano()),
		Content:       content,
		Type:          WisdomPractical,
		Depth:         r[0],
		Applicability: r[1],
		Confidence:    r[2],
		Sources:       append(append([]string(nil), synthesis.InputKnowledge...), synthesis.InputExperiences...),
		CreatedAt:     time.Now(),
	}, nil
}

// synthesisManner describes a synthesis method for the model's prompt.
func synthesisManner(m SynthesisMethod) string {
	switch m {
	case SynthesisReflective:
		return "through deep reflection"
	case SynthesisExperiential:
		return "from lived experience"
	case SynthesisContemplative:
		return "through quiet contemplation"
	case SynthesisDialogical:
		return "as if in dialogue with another mind"
	}
	return "by integrating these"
}

// recordOperation adds an operation to history
func (o *Operations) recordOperation(opType, opID string, success bool, duration time.Duration, details map[string]interface{}) {
	record := &OperationRecord{
		Timestamp:     time.Now(),
		OperationType: opType,
		OperationID:   opID,
		Success:       success,
		Duration:      duration,
		Details:       details,
	}

	o.operationHistory = append(o.operationHistory, record)
}

// GetOperationMetrics returns metrics about operations
func (o *Operations) GetOperationMetrics() map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	successCount := 0
	totalDuration := time.Duration(0)
	for _, record := range o.operationHistory {
		if record.Success {
			successCount++
		}
		totalDuration += record.Duration
	}

	avgDuration := time.Duration(0)
	if len(o.operationHistory) > 0 {
		avgDuration = totalDuration / time.Duration(len(o.operationHistory))
	}

	successRate := 0.0
	if len(o.operationHistory) > 0 {
		successRate = float64(successCount) / float64(len(o.operationHistory))
	}

	return map[string]interface{}{
		"total_operations":       len(o.operationHistory),
		"successful_operations":  successCount,
		"stage_transitions":      len(o.stageTransitions),
		"maturation_operations":  len(o.maturationOperations),
		"integration_operations": len(o.integrationOperations),
		"synthesis_operations":   len(o.synthesisOperations),
		"average_operation_time": avgDuration.String(),
		"success_rate":           successRate,
	}
}

// GetWisdomArtifacts returns all synthesized wisdom
func (o *Operations) GetWisdomArtifacts() []*WisdomArtifact {
	o.mu.RLock()
	defer o.mu.RUnlock()

	artifacts := make([]*WisdomArtifact, 0)
	for _, synthesis := range o.synthesisOperations {
		if synthesis.Completed && synthesis.WisdomOutput != nil {
			artifacts = append(artifacts, synthesis.WisdomOutput)
		}
	}

	return artifacts
}
