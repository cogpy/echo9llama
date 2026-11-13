package deeptreeecho

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/EchoCog/echollama/core/llm"
)

// fakeN5Model answers N+5 prompts in the formats the components ask for.
type fakeN5Model struct{ err error }

func (f fakeN5Model) Generate(_ context.Context, prompt string, _ llm.GenerateOptions) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	switch {
	case strings.Contains(prompt, "ANSWER:"):
		return "THOUGHT: model step one\nWHY: model why one\nTHOUGHT: model step two\nWHY: model why two\nANSWER: model answer", nil
	case strings.Contains(prompt, "OUTCOME:"):
		return "OUTCOME: model outcome\nPLAUSIBILITY: 0.6\nUTILITY: 0.4\nALTERNATIVE: model alt | 0.3 | 0.8 | 0.5", nil
	default:
		return "model text", nil
	}
}

func seededWisdomEngine() *WisdomApplicationEngine {
	wae := NewWisdomApplicationEngine()
	wae.SetTextGenerator(fakeN5Model{})
	wae.AddWisdom("model wisdom one", "learning", []string{"learning", "decisions"}, 0.8)
	wae.AddWisdom("model wisdom two", "strategy", []string{"learning", "problem"}, 0.8)
	return wae
}

// Test MetaCognitiveMonitor

func TestMetaCognitiveMonitor_ProcessTracking(t *testing.T) {
	mcm := NewMetaCognitiveMonitor()

	// Start a process
	processID := mcm.StartProcess("test_thinking", ProcessThinking)
	if processID == "" {
		t.Fatal("Expected non-empty process ID")
	}

	// Update process
	mcm.UpdateProcess(processID, 0.5, StatusActive)

	// Complete process
	mcm.UpdateProcess(processID, 1.0, StatusCompleted)

	// Check metrics
	awareness := mcm.GetSelfAwareness()
	if totalProcs, ok := awareness["total_processes"].(uint64); !ok || totalProcs != 1 {
		t.Errorf("Expected 1 total process, got %v", awareness["total_processes"])
	}
}

func TestMetaCognitiveMonitor_DecisionQuality(t *testing.T) {
	mcm := NewMetaCognitiveMonitor()

	// Record a decision
	decisionID := mcm.RecordDecision(
		"choosing cognitive strategy",
		"deliberate reasoning",
		"task requires careful analysis",
		[]string{"deliberate reasoning", "intuitive processing"},
		0.8,
	)

	if decisionID == "" {
		t.Fatal("Expected non-empty decision ID")
	}

	// Assess decision quality
	outcome := &DecisionOutcome{
		Success:         true,
		ActualBenefit:   0.9,
		ExpectedBenefit: 0.8,
		SideEffects:     []string{},
		LessonsLearned:  []string{"Deliberate reasoning effective for complex tasks"},
	}

	mcm.AssessDecisionQuality(decisionID, outcome)

	// Verify assessment exists
	awareness := mcm.GetSelfAwareness()
	if totalDecisions, ok := awareness["total_decisions"].(uint64); !ok || totalDecisions != 1 {
		t.Errorf("Expected 1 total decision, got %v", awareness["total_decisions"])
	}
}

func TestMetaCognitiveMonitor_StrategySelection(t *testing.T) {
	mcm := NewMetaCognitiveMonitor()

	// Select strategy for problem solving
	strategy := mcm.SelectStrategy(ProcessProblemSolving, map[string]interface{}{
		"accuracy_required": true,
	})

	if strategy == nil {
		t.Fatal("Expected strategy to be selected")
	}

	if strategy.Name == "" {
		t.Error("Expected strategy to have a name")
	}
}

func TestMetaCognitiveMonitor_MetaThoughts(t *testing.T) {
	mcm := NewMetaCognitiveMonitor()
	mcm.SetTextGenerator(fakeN5Model{})

	// Generate meta-thought
	thought := mcm.GenerateMetaThought(
		"decision_making",
		"I'm thinking about how I make decisions",
		0,
	)

	if thought == nil {
		t.Fatal("Expected meta-thought to be generated")
	}

	if thought.Depth != 0 {
		t.Errorf("Expected depth 0, got %d", thought.Depth)
	}

	if thought.Insight != "model text" {
		t.Errorf("Expected model-written insight, got %q", thought.Insight)
	}

	// Check recursive thinking
	awareness := mcm.GetSelfAwareness()
	if metaThoughts, ok := awareness["meta_thoughts"].(int); !ok || metaThoughts < 1 {
		t.Errorf("Expected at least 1 meta-thought, got %v", awareness["meta_thoughts"])
	}
}

func TestMetaCognitiveMonitor_NoModelNoMetaThought(t *testing.T) {
	mcm := NewMetaCognitiveMonitor()
	if thought := mcm.GenerateMetaThought("x", "y", 0); thought != nil {
		t.Fatalf("expected no meta-thought without a model, got %+v", thought)
	}
	mcm.SetTextGenerator(fakeN5Model{err: errors.New("down")})
	if thought := mcm.GenerateMetaThought("x", "y", 0); thought != nil {
		t.Fatalf("expected no meta-thought on model error, got %+v", thought)
	}
}

// Test AdvancedReasoningEngine

func TestAdvancedReasoning_ReasoningChain(t *testing.T) {
	are := NewAdvancedReasoningEngine()

	// Start reasoning chain
	chainID := are.StartReasoningChain("Determine optimal learning strategy")
	if chainID == "" {
		t.Fatal("Expected non-empty chain ID")
	}

	// Add reasoning steps
	are.AddReasoningStep(chainID, StepDeduction,
		"Learning requires practice",
		"More practice leads to better skills",
		"Therefore, increase practice frequency",
		0.9,
	)

	are.AddReasoningStep(chainID, StepInduction,
		"Observed patterns in past learning",
		"Successful learning followed structured approach",
		"Structure aids learning",
		0.8,
	)

	// Complete chain
	are.CompleteReasoningChain(chainID, "Use structured practice for optimal learning", 0.85)

	// Check metrics
	metrics := are.GetReasoningMetrics()
	if completed, ok := metrics["completed_chains"].(int); !ok || completed != 1 {
		t.Errorf("Expected 1 completed chain, got %v", metrics["completed_chains"])
	}
}

func TestAdvancedReasoning_ProblemDecomposition(t *testing.T) {
	are := NewAdvancedReasoningEngine()

	// Decompose a problem
	problem := are.DecomposeProblem(
		"Build an autonomous learning system",
		ProblemDesign,
	)

	if problem == nil {
		t.Fatal("Expected problem to be created")
	}

	if len(problem.SubProblems) == 0 {
		t.Error("Expected problem to be decomposed into sub-problems")
	}

	if problem.Complexity <= 0 {
		t.Error("Expected problem to have non-zero complexity")
	}
}

func TestAdvancedReasoning_CausalInference(t *testing.T) {
	are := NewAdvancedReasoningEngine()
	are.SetTextGenerator(fakeN5Model{})

	// Perform causal reasoning
	evidence := []string{
		"System performance improved after change",
		"Change preceded improvement",
		"Improvement consistent across trials",
	}

	inference := are.PerformCausalReasoning(
		"The change caused the improvement",
		evidence,
	)

	if inference == nil {
		t.Fatal("Expected causal inference to be generated")
	}

	if inference.Confidence <= 0 {
		t.Error("Expected positive confidence in inference")
	}

	if inference.Conclusion == "" {
		t.Error("Expected conclusion to be drawn")
	}
}

func TestAdvancedReasoning_Counterfactual(t *testing.T) {
	are := NewAdvancedReasoningEngine()
	are.SetTextGenerator(fakeN5Model{})

	// Generate counterfactual
	counterfactual := are.GenerateCounterfactual(
		"Current learning approach is slow",
		"If we used reinforcement learning instead",
	)

	if counterfactual == nil {
		t.Fatal("Expected counterfactual to be generated")
	}

	if counterfactual.Plausibility <= 0 || counterfactual.Plausibility > 1 {
		t.Errorf("Expected plausibility in [0,1], got %f", counterfactual.Plausibility)
	}
}

func TestAdvancedReasoning_ChainOfThought(t *testing.T) {
	are := NewAdvancedReasoningEngine()
	are.SetTextGenerator(fakeN5Model{})

	// Perform chain-of-thought reasoning
	chain := are.ChainOfThought("How can I improve my learning efficiency?")

	if chain == nil {
		t.Fatal("Expected thought chain to be generated")
	}

	if len(chain.Thoughts) == 0 {
		t.Error("Expected thoughts to be generated")
	}

	if len(chain.Reasoning) == 0 {
		t.Error("Expected reasoning steps to be generated")
	}

	if chain.Answer != "model answer" {
		t.Errorf("Expected model answer, got %q", chain.Answer)
	}

	if chain.Confidence <= 0 || chain.Confidence > 1 {
		t.Errorf("Expected confidence in [0,1], got %f", chain.Confidence)
	}
}

func TestAdvancedReasoning_NoModelNoReasoning(t *testing.T) {
	are := NewAdvancedReasoningEngine()
	if are.ChainOfThought("q") != nil || are.PerformCausalReasoning("h", []string{"e"}) != nil || are.GenerateCounterfactual("b", "a") != nil {
		t.Fatal("expected no reasoning output without a model")
	}
}

// Test WisdomApplicationEngine

func TestWisdomApplication_FindRelevant(t *testing.T) {
	wae := seededWisdomEngine()

	// Find relevant wisdom
	matches := wae.FindRelevantWisdom("learning from experience", 3)

	if len(matches) == 0 {
		t.Error("Expected to find at least one relevant wisdom entry")
	}

	// Check match properties
	for _, match := range matches {
		if match.Wisdom == nil {
			t.Error("Expected match to contain wisdom")
		}

		if match.RelevanceScore <= 0 {
			t.Error("Expected positive relevance score")
		}

		if match.Explanation == "" {
			t.Error("Expected match explanation")
		}
	}
}

func TestWisdomApplication_Apply(t *testing.T) {
	wae := seededWisdomEngine()

	// Find wisdom
	matches := wae.FindRelevantWisdom("making decisions", 1)
	if len(matches) == 0 {
		t.Skip("No wisdom found for test")
	}

	// Apply wisdom
	application := wae.ApplyWisdom(matches[0].WisdomID, "deciding on learning strategy")

	if application == nil {
		t.Fatal("Expected wisdom application to be created")
	}

	if application.Guidance != "model text" {
		t.Errorf("Expected model-written guidance, got %q", application.Guidance)
	}

	if application.Outcome != OutcomePending {
		t.Errorf("Expected pending outcome before feedback, got %s", application.Outcome)
	}

	// Check metrics
	metrics := wae.GetWisdomMetrics()
	if apps, ok := metrics["total_applications"].(uint64); !ok || apps != 1 {
		t.Errorf("Expected 1 application, got %v", metrics["total_applications"])
	}
}

func TestWisdomApplication_Synthesis(t *testing.T) {
	wae := seededWisdomEngine()

	// Find multiple wisdom entries
	matches := wae.FindRelevantWisdom("learning and adaptation", 5)
	if len(matches) < 2 {
		t.Skip("Need at least 2 wisdom entries for synthesis test")
	}

	// Synthesize new wisdom
	sourceIDs := []string{matches[0].WisdomID, matches[1].WisdomID}
	synthesized := wae.SynthesizeWisdom(sourceIDs)

	if synthesized == nil {
		t.Fatal("Expected synthesized wisdom to be created")
	}

	if synthesized.Category != "synthesized" {
		t.Errorf("Expected category 'synthesized', got %s", synthesized.Category)
	}

	if len(synthesized.RelatedWisdom) != 2 {
		t.Errorf("Expected 2 related wisdom entries, got %d", len(synthesized.RelatedWisdom))
	}
}

func TestWisdomApplication_Feedback(t *testing.T) {
	wae := seededWisdomEngine()

	// Apply wisdom
	matches := wae.FindRelevantWisdom("problem solving", 1)
	if len(matches) == 0 {
		t.Skip("No wisdom found for test")
	}

	application := wae.ApplyWisdom(matches[0].WisdomID, "solving a complex problem")

	// Provide feedback
	wae.ProvideFeedback(application.ID, 0.9, "Very helpful wisdom")

	// Check metrics
	metrics := wae.GetWisdomMetrics()
	if feedback, ok := metrics["feedback_entries"].(int); !ok || feedback != 1 {
		t.Errorf("Expected 1 feedback entry, got %v", metrics["feedback_entries"])
	}
}

func TestN5Integration_NoModelNoSuccess(t *testing.T) {
	n5 := NewN5CognitiveSystem()
	result := n5.ProcessWithMetaCognition("task")
	if result.Success || result.MetaInsight != "" || result.Answer != "" {
		t.Fatalf("expected failure without a model, got %+v", result)
	}
}

// Test N5 Integration

func TestN5Integration_Creation(t *testing.T) {
	n5 := NewN5CognitiveSystem()

	if n5 == nil {
		t.Fatal("Expected N5 system to be created")
	}

	status := n5.GetSystemStatus()
	if status["running"].(bool) {
		t.Error("Expected system to start in stopped state")
	}
}

func TestN5Integration_StartStop(t *testing.T) {
	n5 := NewN5CognitiveSystem()

	// Start system
	err := n5.Start()
	if err != nil {
		t.Fatalf("Failed to start N5 system: %v", err)
	}

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	status := n5.GetSystemStatus()
	if !status["running"].(bool) {
		t.Error("Expected system to be running")
	}

	// Stop system
	n5.Stop()

	status = n5.GetSystemStatus()
	if status["running"].(bool) {
		t.Error("Expected system to be stopped")
	}
}

func TestN5Integration_MetaCognitiveProcessing(t *testing.T) {
	n5 := NewN5CognitiveSystem()
	n5.SetTextGenerator(fakeN5Model{})

	// Process a task with meta-cognition
	result := n5.ProcessWithMetaCognition("Optimize learning strategy")

	if result == nil {
		t.Fatal("Expected result from meta-cognitive processing")
	}

	if !result.Success {
		t.Error("Expected processing to succeed")
	}

	if result.StrategyUsed == "" {
		t.Error("Expected strategy to be selected")
	}

	if result.MetaInsight == "" {
		t.Error("Expected meta-insight to be generated")
	}
}

func TestN5Integration_AdvancedReasoning(t *testing.T) {
	n5 := NewN5CognitiveSystem()
	n5.SetTextGenerator(fakeN5Model{})

	// Reason about a problem
	result := n5.ReasonAboutProblem(
		"How to improve autonomous learning",
		ProblemPlanning,
	)

	if result == nil {
		t.Fatal("Expected reasoning result")
	}

	if result.SubProblems == 0 {
		t.Error("Expected problem to be decomposed")
	}

	if result.CausalConclusion == "" {
		t.Error("Expected causal conclusion")
	}

	if result.Answer == "" {
		t.Error("Expected answer to be generated")
	}
}

func TestN5Integration_WisdomApplication(t *testing.T) {
	n5 := NewN5CognitiveSystem()
	n5.SetTextGenerator(fakeN5Model{})
	n5.wisdomApplication.AddWisdom("model wisdom", "learning", []string{"learning"}, 0.8)

	// Apply wisdom to context
	result := n5.ApplyWisdomToContext("learning new cognitive strategies")

	if result == nil {
		t.Fatal("Expected wisdom application result")
	}

	if result.WisdomFound == 0 {
		t.Error("Expected to find relevant wisdom")
	}
}

func TestN5Integration_SystemStatus(t *testing.T) {
	n5 := NewN5CognitiveSystem()

	// Get system status
	status := n5.GetSystemStatus()

	if status == nil {
		t.Fatal("Expected system status")
	}

	// Check required fields
	requiredFields := []string{"running", "total_operations", "metacognitive", "reasoning", "wisdom"}
	for _, field := range requiredFields {
		if _, exists := status[field]; !exists {
			t.Errorf("Expected status to contain field: %s", field)
		}
	}
}
