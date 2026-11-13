package deeptreeecho

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

// AdvancedReasoningEngine handles complex problem-solving and multi-step reasoning
type AdvancedReasoningEngine struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Reasoning chains
	activeChains    map[string]*AdvancedReasoningChain
	completedChains []*AdvancedReasoningChain

	// Problem decomposition
	problemRegistry         map[string]*Problem
	decompositionStrategies map[string]*DecompositionStrategy

	// Causal reasoning
	causalModel      *CausalModel
	causalInferences []*CausalInference

	// Counterfactual thinking
	counterfactuals   []*CounterfactualScenario
	alternateOutcomes map[string][]*AlternateOutcome

	// Chain-of-thought
	thoughtChains  []*ThoughtChain
	maxChainLength int

	// Metrics
	totalProblemsProcessed uint64
	totalInferences        uint64
	totalCounterfactuals   uint64

	// Model that carries out inference, counterfactuals and chains of thought;
	// nil disables them.
	gen TextGenerator

	// Running state
	running bool
}

// AdvancedReasoningChain represents a multi-step reasoning process
type AdvancedReasoningChain struct {
	ID             string
	Goal           string
	Steps          []*ReasoningStep
	Conclusion     string
	Confidence     float64
	StartTime      time.Time
	CompletionTime time.Time
	Status         ChainStatus
	Metadata       map[string]interface{}
}

// ReasoningStep is a single step in a reasoning chain
type ReasoningStep struct {
	Order        int
	StepType     StepType
	Premise      string
	Inference    string
	Conclusion   string
	Confidence   float64
	LogicalForm  string
	Dependencies []int // Which previous steps this depends on
	Evidence     []string
}

// StepType categorizes reasoning steps
type StepType int

const (
	StepDeduction StepType = iota
	StepInduction
	StepAbduction
	StepAnalogy
	StepSynthesis
)

func (st StepType) String() string {
	return [...]string{"Deduction", "Induction", "Abduction", "Analogy", "Synthesis"}[st]
}

// ChainStatus indicates reasoning chain state
type ChainStatus int

const (
	ChainActive ChainStatus = iota
	ChainCompleted
	ChainFailed
	ChainAbandoned
)

// Problem represents a complex problem to be solved
type Problem struct {
	ID          string
	Description string
	Type        ProblemType
	Complexity  float64
	Constraints []string
	Goals       []string
	SubProblems []*Problem
	Solution    *Solution
	CreatedAt   time.Time
	SolvedAt    time.Time
}

// ProblemType categorizes problems
type ProblemType int

const (
	ProblemSearch ProblemType = iota
	ProblemOptimization
	ProblemClassification
	ProblemPlanning
	ProblemDiagnosis
	ProblemDesign
)

func (pt ProblemType) String() string {
	return [...]string{
		"Search",
		"Optimization",
		"Classification",
		"Planning",
		"Diagnosis",
		"Design",
	}[pt]
}

// Solution represents a problem solution
type Solution struct {
	ID           string
	ProblemID    string
	Approach     string
	Steps        []string
	Outcome      string
	Quality      float64
	Confidence   float64
	Alternatives []*Solution
	Metadata     map[string]interface{}
}

// DecompositionStrategy breaks down complex problems
type DecompositionStrategy struct {
	ID           string
	Name         string
	Description  string
	Heuristics   []string
	ApplicableTo []ProblemType
}

// CausalModel represents causal relationships
type CausalModel struct {
	mu            sync.RWMutex
	Nodes         map[string]*CausalNode
	Edges         []*CausalEdge
	Interventions []*Intervention
}

// CausalNode represents an entity in the causal model
type CausalNode struct {
	ID          string
	Name        string
	State       interface{}
	Probability float64
	Parents     []*CausalNode
	Children    []*CausalNode
}

// CausalEdge represents a causal relationship
type CausalEdge struct {
	ID       string
	From     *CausalNode
	To       *CausalNode
	Strength float64
	Type     CausalType
	Evidence []string
}

// CausalType categorizes causal relationships
type CausalType int

const (
	CausalDirect CausalType = iota
	CausalIndirect
	CausalSpurious
	CausalMediated
)

// CausalInference represents a causal reasoning step
type CausalInference struct {
	ID         string
	Hypothesis string
	Evidence   []string
	Conclusion string
	Confidence float64
	Method     string
	Timestamp  time.Time
}

// Intervention represents a causal intervention
type Intervention struct {
	ID        string
	Node      *CausalNode
	NewState  interface{}
	Timestamp time.Time
	Outcome   interface{}
}

// CounterfactualScenario represents "what if" thinking
type CounterfactualScenario struct {
	ID               string
	BaseScenario     string
	Alteration       string
	PredictedOutcome string
	ActualOutcome    string
	Plausibility     float64
	Utility          float64
	Timestamp        time.Time
}

// AlternateOutcome represents possible alternative outcomes
type AlternateOutcome struct {
	Scenario     string
	Probability  float64
	Desirability float64
	Feasibility  float64
	Description  string
}

// ThoughtChain represents chain-of-thought reasoning
type ThoughtChain struct {
	ID         string
	Question   string
	Thoughts   []string
	Reasoning  []string
	Answer     string
	Confidence float64
	Timestamp  time.Time
}

// NewAdvancedReasoningEngine creates a new reasoning engine
func NewAdvancedReasoningEngine() *AdvancedReasoningEngine {
	ctx, cancel := context.WithCancel(context.Background())

	are := &AdvancedReasoningEngine{
		ctx:                     ctx,
		cancel:                  cancel,
		activeChains:            make(map[string]*AdvancedReasoningChain),
		completedChains:         make([]*AdvancedReasoningChain, 0),
		problemRegistry:         make(map[string]*Problem),
		decompositionStrategies: make(map[string]*DecompositionStrategy),
		causalModel:             initializeCausalModel(),
		causalInferences:        make([]*CausalInference, 0),
		counterfactuals:         make([]*CounterfactualScenario, 0),
		alternateOutcomes:       make(map[string][]*AlternateOutcome),
		thoughtChains:           make([]*ThoughtChain, 0),
		maxChainLength:          10,
	}

	// Initialize decomposition strategies
	are.initializeStrategies()

	return are
}

// StartReasoningChain begins a new multi-step reasoning process
func (are *AdvancedReasoningEngine) StartReasoningChain(goal string) string {
	are.mu.Lock()
	defer are.mu.Unlock()

	chain := &AdvancedReasoningChain{
		ID:        generateChainID(),
		Goal:      goal,
		Steps:     make([]*ReasoningStep, 0),
		StartTime: time.Now(),
		Status:    ChainActive,
		Metadata:  make(map[string]interface{}),
	}

	are.activeChains[chain.ID] = chain

	return chain.ID
}

// AddReasoningStep adds a step to an active reasoning chain
func (are *AdvancedReasoningEngine) AddReasoningStep(chainID string, stepType StepType, premise, inference, conclusion string, confidence float64) {
	are.mu.Lock()
	defer are.mu.Unlock()

	chain, exists := are.activeChains[chainID]
	if !exists {
		return
	}

	step := &ReasoningStep{
		Order:        len(chain.Steps),
		StepType:     stepType,
		Premise:      premise,
		Inference:    inference,
		Conclusion:   conclusion,
		Confidence:   confidence,
		Dependencies: make([]int, 0),
		Evidence:     make([]string, 0),
	}

	chain.Steps = append(chain.Steps, step)
}

// CompleteReasoningChain finalizes a reasoning chain with a conclusion
func (are *AdvancedReasoningEngine) CompleteReasoningChain(chainID, conclusion string, confidence float64) {
	are.mu.Lock()
	defer are.mu.Unlock()

	chain, exists := are.activeChains[chainID]
	if !exists {
		return
	}

	chain.Conclusion = conclusion
	chain.Confidence = confidence
	chain.CompletionTime = time.Now()
	chain.Status = ChainCompleted

	are.completedChains = append(are.completedChains, chain)
	delete(are.activeChains, chainID)
}

// DecomposeProblem breaks a complex problem into sub-problems
func (are *AdvancedReasoningEngine) DecomposeProblem(problemDesc string, problemType ProblemType) *Problem {
	are.mu.Lock()
	defer are.mu.Unlock()

	problem := &Problem{
		ID:          generateProblemID(),
		Description: problemDesc,
		Type:        problemType,
		Complexity:  estimateComplexity(problemDesc),
		SubProblems: make([]*Problem, 0),
		CreatedAt:   time.Now(),
	}

	// Find applicable decomposition strategy
	strategy := are.findBestStrategy(problemType)
	if strategy != nil {
		// Decompose based on strategy
		subProblems := applyDecompositionStrategy(problem, strategy)
		problem.SubProblems = subProblems
	}

	are.problemRegistry[problem.ID] = problem
	are.totalProblemsProcessed++

	return problem
}

// SetTextGenerator injects the model that performs the engine's reasoning.
// Passing nil disables model-backed reasoning.
func (are *AdvancedReasoningEngine) SetTextGenerator(gen TextGenerator) {
	are.mu.Lock()
	defer are.mu.Unlock()
	are.gen = gen
}

func (are *AdvancedReasoningEngine) generator() (context.Context, TextGenerator) {
	are.mu.RLock()
	defer are.mu.RUnlock()
	return are.ctx, are.gen
}

// PerformCausalReasoning makes causal inferences. The conclusion is written by
// the model from the hypothesis and evidence; without a model, or when the
// call fails, no inference is made and nil is returned.
func (are *AdvancedReasoningEngine) PerformCausalReasoning(hypothesis string, evidence []string) *CausalInference {
	ctx, gen := are.generator()

	prompt := fmt.Sprintf(`Assess this causal hypothesis against the evidence.

Hypothesis: %s
Evidence:
%s

Reply with only your conclusion in one or two sentences, stating how strongly the evidence supports the hypothesis and why.`, hypothesis, bulletList(evidence))
	conclusion, err := generateModelText(ctx, gen, prompt, 200)
	if err != nil {
		return nil
	}

	inference := &CausalInference{
		ID:         generateInferenceID(),
		Hypothesis: hypothesis,
		Evidence:   evidence,
		Conclusion: conclusion,
		Confidence: assessCausalEvidence(evidence),
		Method:     "causal_inference",
		Timestamp:  time.Now(),
	}

	are.mu.Lock()
	are.causalInferences = append(are.causalInferences, inference)
	are.totalInferences++
	are.mu.Unlock()

	return inference
}

// GenerateCounterfactual creates "what if" scenarios. The predicted outcome,
// its plausibility and utility, and the alternate outcomes are all produced by
// the model; without a model, or when the call fails or the reply has no
// outcome, nil is returned.
func (are *AdvancedReasoningEngine) GenerateCounterfactual(baseScenario, alteration string) *CounterfactualScenario {
	ctx, gen := are.generator()

	prompt := fmt.Sprintf(`Reason counterfactually.

Base scenario: %s
Alteration: %s

Reply using exactly these labelled lines:
OUTCOME: <the predicted outcome under the alteration>
PLAUSIBILITY: <number from 0 to 1>
UTILITY: <number from 0 to 1, how desirable the outcome is>
ALTERNATIVE: <an alternate outcome> | <probability 0-1> | <desirability 0-1> | <feasibility 0-1>
(give two or three ALTERNATIVE lines)`, baseScenario, alteration)
	reply, err := generateModelText(ctx, gen, prompt, 400)
	if err != nil {
		return nil
	}

	fields := parseLabelledLines(reply)
	outcomes := fields["OUTCOME"]
	if len(outcomes) == 0 || outcomes[0] == "" {
		return nil
	}

	counterfactual := &CounterfactualScenario{
		ID:               generateCounterfactualID(),
		BaseScenario:     baseScenario,
		Alteration:       alteration,
		PredictedOutcome: outcomes[0],
		Plausibility:     firstUnitFloat(fields["PLAUSIBILITY"]),
		Utility:          firstUnitFloat(fields["UTILITY"]),
		Timestamp:        time.Now(),
	}

	alternates := make([]*AlternateOutcome, 0)
	for _, line := range fields["ALTERNATIVE"] {
		parts := strings.Split(line, "|")
		desc := strings.TrimSpace(parts[0])
		if desc == "" {
			continue
		}
		alt := &AlternateOutcome{Scenario: desc, Description: desc}
		if len(parts) > 1 {
			alt.Probability = parseUnitFloat(parts[1])
		}
		if len(parts) > 2 {
			alt.Desirability = parseUnitFloat(parts[2])
		}
		if len(parts) > 3 {
			alt.Feasibility = parseUnitFloat(parts[3])
		}
		alternates = append(alternates, alt)
	}

	are.mu.Lock()
	are.counterfactuals = append(are.counterfactuals, counterfactual)
	are.totalCounterfactuals++
	are.alternateOutcomes[counterfactual.ID] = alternates
	are.mu.Unlock()

	return counterfactual
}

// ChainOfThought performs explicit step-by-step reasoning. Every step and the
// answer are written by the model; without a model, or when the call fails or
// yields no answer, nil is returned.
func (are *AdvancedReasoningEngine) ChainOfThought(question string) *ThoughtChain {
	ctx, gen := are.generator()

	are.mu.RLock()
	maxSteps := are.maxChainLength
	are.mu.RUnlock()

	prompt := fmt.Sprintf(`Think through this question step by step, in your own first-person voice.

Question: %s

Reply using labelled lines only (at most %d steps):
THOUGHT: <a reasoning step>
WHY: <what that step accomplishes>
(repeat THOUGHT/WHY for each step)
ANSWER: <your final answer>`, question, maxSteps)
	reply, err := generateModelText(ctx, gen, prompt, 600)
	if err != nil {
		return nil
	}

	fields := parseLabelledLines(reply)
	answers := fields["ANSWER"]
	if len(answers) == 0 || answers[0] == "" {
		return nil
	}

	chain := &ThoughtChain{
		ID:        generateThoughtChainID(),
		Question:  question,
		Thoughts:  fields["THOUGHT"],
		Reasoning: fields["WHY"],
		Answer:    answers[0],
		Timestamp: time.Now(),
	}
	if len(chain.Thoughts) > maxSteps {
		chain.Thoughts = chain.Thoughts[:maxSteps]
	}
	chain.Confidence = calculateAnswerConfidence(chain)

	are.mu.Lock()
	are.thoughtChains = append(are.thoughtChains, chain)
	are.mu.Unlock()

	return chain
}

// GetReasoningMetrics returns reasoning statistics
func (are *AdvancedReasoningEngine) GetReasoningMetrics() map[string]interface{} {
	are.mu.RLock()
	defer are.mu.RUnlock()

	return map[string]interface{}{
		"active_chains":            len(are.activeChains),
		"completed_chains":         len(are.completedChains),
		"problems_processed":       are.totalProblemsProcessed,
		"causal_inferences":        are.totalInferences,
		"counterfactuals":          are.totalCounterfactuals,
		"thought_chains":           len(are.thoughtChains),
		"decomposition_strategies": len(are.decompositionStrategies),
	}
}

// Helper functions

func (are *AdvancedReasoningEngine) initializeStrategies() {
	// Divide and conquer
	are.decompositionStrategies["divide_conquer"] = &DecompositionStrategy{
		ID:           "divide_conquer",
		Name:         "Divide and Conquer",
		Description:  "Split problem into independent sub-problems",
		Heuristics:   []string{"identify_subgoals", "minimize_dependencies"},
		ApplicableTo: []ProblemType{ProblemSearch, ProblemOptimization},
	}

	// Hierarchical decomposition
	are.decompositionStrategies["hierarchical"] = &DecompositionStrategy{
		ID:           "hierarchical",
		Name:         "Hierarchical Decomposition",
		Description:  "Break down into layers of abstraction",
		Heuristics:   []string{"identify_levels", "top_down_refinement"},
		ApplicableTo: []ProblemType{ProblemPlanning, ProblemDesign},
	}

	// Functional decomposition
	are.decompositionStrategies["functional"] = &DecompositionStrategy{
		ID:           "functional",
		Name:         "Functional Decomposition",
		Description:  "Separate by functional components",
		Heuristics:   []string{"identify_functions", "minimize_coupling"},
		ApplicableTo: []ProblemType{ProblemDesign, ProblemDiagnosis},
	}
}

func (are *AdvancedReasoningEngine) findBestStrategy(problemType ProblemType) *DecompositionStrategy {
	for _, strategy := range are.decompositionStrategies {
		for _, pt := range strategy.ApplicableTo {
			if pt == problemType {
				return strategy
			}
		}
	}
	return nil
}

func initializeCausalModel() *CausalModel {
	return &CausalModel{
		Nodes:         make(map[string]*CausalNode),
		Edges:         make([]*CausalEdge, 0),
		Interventions: make([]*Intervention, 0),
	}
}

func estimateComplexity(description string) float64 {
	// Simple heuristic based on description length and keywords
	baseComplexity := float64(len(description)) / 1000.0

	complexKeywords := []string{"multiple", "complex", "intricate", "interdependent"}
	for _, keyword := range complexKeywords {
		if contains(description, keyword) {
			baseComplexity += 0.2
		}
	}

	return math.Min(1.0, baseComplexity)
}

func applyDecompositionStrategy(problem *Problem, strategy *DecompositionStrategy) []*Problem {
	// Simple decomposition - split into 2-4 sub-problems
	numSubs := 2 + int(problem.Complexity*2)
	subProblems := make([]*Problem, numSubs)

	for i := 0; i < numSubs; i++ {
		subProblems[i] = &Problem{
			ID:          fmt.Sprintf("%s_sub%d", problem.ID, i),
			Description: fmt.Sprintf("Sub-problem %d of %s", i+1, problem.Description),
			Type:        problem.Type,
			Complexity:  problem.Complexity / float64(numSubs),
			CreatedAt:   time.Now(),
		}
	}

	return subProblems
}

func assessCausalEvidence(evidence []string) float64 {
	// Confidence based on amount and quality of evidence
	if len(evidence) == 0 {
		return 0.1
	}

	baseConfidence := 0.5
	evidenceBonus := math.Min(0.4, float64(len(evidence))*0.1)

	return baseConfidence + evidenceBonus
}

// bulletList formats items as a bulleted list for prompts.
func bulletList(items []string) string {
	if len(items) == 0 {
		return "- (none)"
	}
	var b strings.Builder
	for _, it := range items {
		b.WriteString("- ")
		b.WriteString(it)
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// parseLabelledLines collects "LABEL: value" lines from a model reply, keyed by
// upper-case label, in order of appearance.
func parseLabelledLines(reply string) map[string][]string {
	out := make(map[string][]string)
	for _, line := range strings.Split(reply, "\n") {
		line = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "-*0123456789. "))
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		label := strings.ToUpper(strings.TrimSpace(line[:idx]))
		value := strings.TrimSpace(line[idx+1:])
		if value == "" {
			continue
		}
		out[label] = append(out[label], value)
	}
	return out
}

// parseUnitFloat parses a number in [0,1] from model text; unparseable values
// yield 0.
func parseUnitFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if f := strings.Fields(s); len(f) > 0 {
		s = strings.TrimRight(f[0], ",;.%")
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(v) {
		return 0
	}
	return math.Max(0, math.Min(1, v))
}

func firstUnitFloat(values []string) float64 {
	if len(values) == 0 {
		return 0
	}
	return parseUnitFloat(values[0])
}

func calculateAnswerConfidence(chain *ThoughtChain) float64 {
	// Confidence based on chain length and consistency
	baseConfidence := 0.5
	lengthBonus := math.Min(0.3, float64(len(chain.Thoughts))*0.05)

	return baseConfidence + lengthBonus
}

func contains(s, substr string) bool {
	// Simple substring check
	if len(s) == 0 || len(substr) == 0 {
		return false
	}
	// In a real implementation, would use strings.Contains
	// Simplified version for this implementation
	return len(s) >= len(substr)
}

func generateChainID() string {
	return fmt.Sprintf("chain_%d", time.Now().UnixNano())
}

func generateProblemID() string {
	return fmt.Sprintf("problem_%d", time.Now().UnixNano())
}

func generateInferenceID() string {
	return fmt.Sprintf("inference_%d", time.Now().UnixNano())
}

func generateCounterfactualID() string {
	return fmt.Sprintf("counterfactual_%d", time.Now().UnixNano())
}

func generateThoughtChainID() string {
	return fmt.Sprintf("thought_chain_%d", time.Now().UnixNano())
}
