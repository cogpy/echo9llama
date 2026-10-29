package deeptreeecho

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/EchoCog/echollama/core/echobeats"
)

// N5CognitiveSystem integrates all Iteration N+5 enhancements
type N5CognitiveSystem struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// N+5 Components
	interestEvolution    *echobeats.InterestEvolutionEngine
	metacognitiveMonitor *MetaCognitiveMonitor
	advancedReasoning    *AdvancedReasoningEngine
	wisdomApplication    *WisdomApplicationEngine

	// Integration state
	running    bool
	lastUpdate time.Time

	// Metrics
	totalOperations    uint64
	integrationMetrics map[string]interface{}
}

// NewN5CognitiveSystem creates a new Iteration N+5 integrated cognitive system
func NewN5CognitiveSystem() *N5CognitiveSystem {
	ctx, cancel := context.WithCancel(context.Background())

	return &N5CognitiveSystem{
		ctx:                  ctx,
		cancel:               cancel,
		interestEvolution:    echobeats.NewInterestEvolutionEngine(),
		metacognitiveMonitor: NewMetaCognitiveMonitor(),
		advancedReasoning:    NewAdvancedReasoningEngine(),
		wisdomApplication:    NewWisdomApplicationEngine(),
		integrationMetrics:   make(map[string]interface{}),
		lastUpdate:           time.Now(),
	}
}

// Start begins autonomous operation of the N+5 system
func (n5 *N5CognitiveSystem) Start() error {
	n5.mu.Lock()
	defer n5.mu.Unlock()

	if n5.running {
		return fmt.Errorf("N5 system already running")
	}

	n5.running = true

	// Start autonomous cognitive loops
	go n5.autonomousCognitiveLoop()

	return nil
}

// Stop gracefully stops the N+5 system
func (n5 *N5CognitiveSystem) Stop() {
	n5.mu.Lock()
	defer n5.mu.Unlock()

	if !n5.running {
		return
	}

	n5.running = false
	n5.cancel()
}

// SetTextGenerator injects the model into every N+5 component that produces
// Deep Tree Echo's words (meta-thoughts, reasoning, wisdom). Passing nil
// disables them.
func (n5 *N5CognitiveSystem) SetTextGenerator(gen TextGenerator) {
	n5.metacognitiveMonitor.SetTextGenerator(gen)
	n5.advancedReasoning.SetTextGenerator(gen)
	n5.wisdomApplication.SetTextGenerator(gen)
}

// ProcessWithMetaCognition processes a task with full meta-cognitive awareness.
// The reasoning and the meta-insight come from the model; when it is missing
// or fails, the process is marked failed and Success is false.
func (n5 *N5CognitiveSystem) ProcessWithMetaCognition(task string) *MetaCognitiveResult {
	n5.mu.Lock()
	n5.totalOperations++
	n5.mu.Unlock()

	result := &MetaCognitiveResult{
		Task:      task,
		StartTime: time.Now(),
	}

	// 1. Monitor the process
	processID := n5.metacognitiveMonitor.StartProcess(task, ProcessProblemSolving)

	// 2. Select cognitive strategy
	strategy := n5.metacognitiveMonitor.SelectStrategy(ProcessProblemSolving, map[string]interface{}{
		"accuracy_required": true,
	})
	if strategy != nil {
		result.StrategyUsed = strategy.Name
	}

	// 3. Find relevant wisdom
	wisdomMatches := n5.wisdomApplication.FindRelevantWisdom(task, 3)
	result.WisdomApplied = len(wisdomMatches)

	// 4. Perform model-backed reasoning
	chainID := n5.advancedReasoning.StartReasoningChain(fmt.Sprintf("Solve: %s", task))
	thoughtChain := n5.advancedReasoning.ChainOfThought(task)
	if thoughtChain == nil {
		n5.advancedReasoning.mu.Lock()
		if chain, ok := n5.advancedReasoning.activeChains[chainID]; ok {
			chain.Status = ChainFailed
			chain.CompletionTime = time.Now()
			delete(n5.advancedReasoning.activeChains, chainID)
		}
		n5.advancedReasoning.mu.Unlock()
		n5.metacognitiveMonitor.UpdateProcess(processID, 0, StatusFailed)
		result.EndTime = time.Now()
		return result
	}
	for i, thought := range thoughtChain.Thoughts {
		why := ""
		if i < len(thoughtChain.Reasoning) {
			why = thoughtChain.Reasoning[i]
		}
		n5.advancedReasoning.AddReasoningStep(chainID, StepDeduction, task, why, thought, thoughtChain.Confidence)
	}
	n5.advancedReasoning.CompleteReasoningChain(chainID, thoughtChain.Answer, thoughtChain.Confidence)
	result.Answer = thoughtChain.Answer

	// 5. Model-written meta-thought about how the task was processed
	metaThought := n5.metacognitiveMonitor.GenerateMetaThought(
		"task_processing",
		fmt.Sprintf("Task: %s\nStrategy: %s\nMy reasoning: %s\nMy answer: %s",
			task, result.StrategyUsed, strings.Join(thoughtChain.Thoughts, " / "), thoughtChain.Answer),
		0,
	)
	if metaThought == nil {
		n5.metacognitiveMonitor.UpdateProcess(processID, 1.0, StatusFailed)
		result.EndTime = time.Now()
		return result
	}
	result.MetaInsight = metaThought.Insight

	n5.metacognitiveMonitor.UpdateProcess(processID, 1.0, StatusCompleted)

	result.EndTime = time.Now()
	result.Success = true

	return result
}

// LearnFromEngagement updates interest patterns based on engagement
func (n5 *N5CognitiveSystem) LearnFromEngagement(topic string, duration time.Duration, satisfaction float64) {
	n5.mu.RLock()
	defer n5.mu.RUnlock()

	// Create engagement outcome
	outcome := &echobeats.EngagementOutcome{
		InterestID:   topic,
		Timestamp:    time.Now(),
		Duration:     duration,
		Reward:       satisfaction,
		LearningGain: 0.6,
		Satisfaction: satisfaction,
		Competence:   0.05, // Small competence gain
		NoveltyValue: 0.4,
	}

	// Apply reinforcement learning (would need to get actual interest from system)
	// This is a simplified version - in practice would integrate with InterestPatternSystem
	_ = outcome
}

// ReasonAboutProblem uses advanced reasoning for a problem. The answer, the
// causal conclusion and the alternative scenario are produced by the model;
// whatever the model could not produce is left empty.
func (n5 *N5CognitiveSystem) ReasonAboutProblem(problem string, problemType ProblemType) *ReasoningResult {
	result := &ReasoningResult{
		Problem:   problem,
		StartTime: time.Now(),
	}

	// Decompose problem
	decomposed := n5.advancedReasoning.DecomposeProblem(problem, problemType)
	result.SubProblems = len(decomposed.SubProblems)

	// Chain of thought reasoning
	thoughtChain := n5.advancedReasoning.ChainOfThought(problem)
	if thoughtChain != nil {
		result.ReasoningSteps = len(thoughtChain.Thoughts)
		result.Answer = thoughtChain.Answer
		result.Confidence = thoughtChain.Confidence

		// Causal reasoning over the model's own reasoning steps
		inference := n5.advancedReasoning.PerformCausalReasoning(
			fmt.Sprintf("Solving %s requires addressing its root causes", problem),
			thoughtChain.Thoughts,
		)
		if inference != nil {
			result.CausalConclusion = inference.Conclusion
		}
	}

	// Counterfactual
	counterfactual := n5.advancedReasoning.GenerateCounterfactual(
		problem,
		"The problem is approached in a substantially different way",
	)
	if counterfactual != nil {
		result.AlternativeScenarios = 1
		result.BestAlternative = counterfactual.PredictedOutcome
	}

	result.EndTime = time.Now()

	return result
}

// ApplyWisdomToContext finds relevant wisdom and has the model apply it. No
// feedback is fabricated: the application stays pending until real feedback
// arrives via the wisdom engine's ProvideFeedback.
func (n5 *N5CognitiveSystem) ApplyWisdomToContext(context string) *WisdomApplicationResult {
	result := &WisdomApplicationResult{
		Context:   context,
		StartTime: time.Now(),
	}

	// Find relevant wisdom
	matches := n5.wisdomApplication.FindRelevantWisdom(context, 5)
	result.WisdomFound = len(matches)

	if len(matches) > 0 {
		bestMatch := matches[0]
		result.WisdomApplied = bestMatch.Wisdom.Content
		result.Relevance = bestMatch.RelevanceScore

		if application := n5.wisdomApplication.ApplyWisdom(bestMatch.WisdomID, context); application != nil {
			result.Guidance = application.Guidance
			result.Success = true
		}
	}

	// Try synthesizing new wisdom if we found multiple matches
	if len(matches) >= 2 {
		sourceIDs := []string{matches[0].WisdomID, matches[1].WisdomID}
		if synthesized := n5.wisdomApplication.SynthesizeWisdom(sourceIDs); synthesized != nil {
			result.SynthesizedWisdom = synthesized.Content
		}
	}

	result.EndTime = time.Now()

	return result
}

// GetSystemStatus returns comprehensive N+5 system status
func (n5 *N5CognitiveSystem) GetSystemStatus() map[string]interface{} {
	n5.mu.RLock()
	defer n5.mu.RUnlock()

	return map[string]interface{}{
		"running":          n5.running,
		"total_operations": n5.totalOperations,
		"last_update":      n5.lastUpdate,
		"metacognitive":    n5.metacognitiveMonitor.GetSelfAwareness(),
		"reasoning":        n5.advancedReasoning.GetReasoningMetrics(),
		"wisdom":           n5.wisdomApplication.GetWisdomMetrics(),
	}
}

// autonomousCognitiveLoop runs autonomous cognitive processes
func (n5 *N5CognitiveSystem) autonomousCognitiveLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-n5.ctx.Done():
			return
		case <-ticker.C:
			n5.autonomousCognitiveCycle()
		}
	}
}

// autonomousCognitiveCycle performs one cycle of autonomous cognition
func (n5 *N5CognitiveSystem) autonomousCognitiveCycle() {
	n5.mu.RLock()
	running := n5.running
	n5.mu.RUnlock()
	if !running {
		return
	}

	// Model-written meta-thought about the current system state (none if the
	// model is unavailable)
	awareness := n5.metacognitiveMonitor.GetSelfAwareness()
	n5.metacognitiveMonitor.GenerateMetaThought(
		"system_state",
		fmt.Sprintf("My current state: awareness %v, confidence %v, processes monitored %v, decisions reviewed %v, meta-thoughts %v",
			awareness["awareness_level"], awareness["confidence_level"], awareness["total_processes"],
			awareness["total_decisions"], awareness["meta_thoughts"]),
		0,
	)

	n5.mu.Lock()
	defer n5.mu.Unlock()

	// Update metrics
	n5.integrationMetrics["last_autonomous_cycle"] = time.Now()
	n5.integrationMetrics["awareness_level"] = awareness["awareness_level"]

	n5.lastUpdate = time.Now()
}

// Result types

// MetaCognitiveResult represents the result of meta-cognitive processing
type MetaCognitiveResult struct {
	Task          string
	StartTime     time.Time
	EndTime       time.Time
	StrategyUsed  string
	WisdomApplied int
	Answer        string
	MetaInsight   string
	Success       bool
}

// ReasoningResult represents advanced reasoning output
type ReasoningResult struct {
	Problem              string
	StartTime            time.Time
	EndTime              time.Time
	SubProblems          int
	CausalConclusion     string
	AlternativeScenarios int
	BestAlternative      string
	ReasoningSteps       int
	Answer               string
	Confidence           float64
}

// WisdomApplicationResult represents wisdom application output
type WisdomApplicationResult struct {
	Context           string
	StartTime         time.Time
	EndTime           time.Time
	WisdomFound       int
	WisdomApplied     string
	Relevance         float64
	Effectiveness     float64
	Guidance          string
	Success           bool
	SynthesizedWisdom string
}
