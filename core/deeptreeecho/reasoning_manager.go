package deeptreeecho

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
	"github.com/google/uuid"
)

// generateReasoningID generates a unique ID for reasoning components
func generateReasoningID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, uuid.New().String()[:8])
}

// ReasoningMode represents the mode of reasoning
type ReasoningMode int

const (
	ReasoningModeReactive ReasoningMode = iota
	ReasoningModeDeliberative
	ReasoningModeReflective
	ReasoningModeCreative
)

func (rm ReasoningMode) String() string {
	switch rm {
	case ReasoningModeReactive:
		return "reactive"
	case ReasoningModeDeliberative:
		return "deliberative"
	case ReasoningModeReflective:
		return "reflective"
	case ReasoningModeCreative:
		return "creative"
	default:
		return "unknown"
	}
}

// ReasoningStep represents a single step in a reasoning chain
type ReasoningStep struct {
	ID          string
	Thought     string
	Action      string
	ActionInput string
	Observation string
	Timestamp   time.Time
}

// LangchainReasoningChain represents a complete chain of reasoning for langchain integration
type LangchainReasoningChain struct {
	ID        string
	Goal      string
	Mode      ReasoningMode
	Steps     []ReasoningStep
	Result    string
	Success   bool
	StartTime time.Time
	EndTime   time.Time
}

// ReasoningTask represents a task to be reasoned about
type ReasoningTask struct {
	ID          string
	Input       string
	Mode        ReasoningMode
	Priority    float64
	MaxSteps    int
	Timeout     time.Duration
	Context     map[string]interface{}
	ResultChan  chan ReasoningResult
}

// ReasoningResult represents the result of a reasoning task
type ReasoningResult struct {
	TaskID  string
	Chain   *LangchainReasoningChain
	Output  string
	Error   error
}

// ReasoningManager orchestrates complex reasoning using cognitive tools
type ReasoningManager struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	
	// LLM provider for reasoning
	llmProvider     llm.LLMProvider
	
	// Available cognitive tools
	tools           map[string]CognitiveTool
	
	// Task queue
	taskQueue       chan ReasoningTask
	
	// Active chains
	activeChains    map[string]*LangchainReasoningChain
	completedChains []*LangchainReasoningChain
	
	// Configuration
	maxConcurrent   int
	defaultMaxSteps int
	defaultTimeout  time.Duration
	
	// Metrics
	totalTasks      uint64
	successfulTasks uint64
	failedTasks     uint64
	totalSteps      uint64
	
	// Running state
	running         bool
	
	// Callbacks
	onChainStart    func(chain *LangchainReasoningChain)
	onStepComplete  func(chain *LangchainReasoningChain, step ReasoningStep)
	onChainComplete func(chain *LangchainReasoningChain)
}

// NewReasoningManager creates a new reasoning manager
func NewReasoningManager(llmProvider llm.LLMProvider, tools []CognitiveTool) *ReasoningManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	toolMap := make(map[string]CognitiveTool)
	for _, tool := range tools {
		toolMap[strings.ToUpper(tool.Name())] = tool
	}
	
	return &ReasoningManager{
		ctx:             ctx,
		cancel:          cancel,
		llmProvider:     llmProvider,
		tools:           toolMap,
		taskQueue:       make(chan ReasoningTask, 100),
		activeChains:    make(map[string]*LangchainReasoningChain),
		completedChains: make([]*LangchainReasoningChain, 0),
		maxConcurrent:   3,
		defaultMaxSteps: 10,
		defaultTimeout:  2 * time.Minute,
	}
}

// SetCallbacks sets the reasoning manager callbacks
func (rm *ReasoningManager) SetCallbacks(
	onChainStart func(chain *LangchainReasoningChain),
	onStepComplete func(chain *LangchainReasoningChain, step ReasoningStep),
	onChainComplete func(chain *LangchainReasoningChain),
) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.onChainStart = onChainStart
	rm.onStepComplete = onStepComplete
	rm.onChainComplete = onChainComplete
}

// Start begins the reasoning manager
func (rm *ReasoningManager) Start() error {
	rm.mu.Lock()
	if rm.running {
		rm.mu.Unlock()
		return fmt.Errorf("reasoning manager already running")
	}
	rm.running = true
	rm.mu.Unlock()
	
	// Start worker goroutines
	for i := 0; i < rm.maxConcurrent; i++ {
		go rm.reasoningWorker(i)
	}
	
	fmt.Println("🧠 Reasoning Manager started")
	return nil
}

// Stop stops the reasoning manager
func (rm *ReasoningManager) Stop() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if !rm.running {
		return fmt.Errorf("reasoning manager not running")
	}
	
	rm.running = false
	rm.cancel()
	close(rm.taskQueue)
	
	fmt.Println("🧠 Reasoning Manager stopped")
	return nil
}

// Reason performs a reasoning task synchronously
func (rm *ReasoningManager) Reason(ctx context.Context, input string, mode ReasoningMode) (string, error) {
	resultChan := make(chan ReasoningResult, 1)
	
	task := ReasoningTask{
		ID:         generateReasoningID("reason"),
		Input:      input,
		Mode:       mode,
		Priority:   0.5,
		MaxSteps:   rm.defaultMaxSteps,
		Timeout:    rm.defaultTimeout,
		Context:    make(map[string]interface{}),
		ResultChan: resultChan,
	}
	
	// Submit task
	select {
	case rm.taskQueue <- task:
	case <-ctx.Done():
		return "", ctx.Err()
	}
	
	// Wait for result
	select {
	case result := <-resultChan:
		if result.Error != nil {
			return "", result.Error
		}
		return result.Output, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// ReasonAsync performs a reasoning task asynchronously
func (rm *ReasoningManager) ReasonAsync(input string, mode ReasoningMode, priority float64) (string, chan ReasoningResult) {
	resultChan := make(chan ReasoningResult, 1)
	
	taskID := generateReasoningID("reason")
	task := ReasoningTask{
		ID:         taskID,
		Input:      input,
		Mode:       mode,
		Priority:   priority,
		MaxSteps:   rm.defaultMaxSteps,
		Timeout:    rm.defaultTimeout,
		Context:    make(map[string]interface{}),
		ResultChan: resultChan,
	}
	
	// Submit task non-blocking
	select {
	case rm.taskQueue <- task:
	default:
		// Queue full, return error
		go func() {
			resultChan <- ReasoningResult{
				TaskID: taskID,
				Error:  fmt.Errorf("reasoning queue full"),
			}
		}()
	}
	
	return taskID, resultChan
}

// reasoningWorker processes reasoning tasks
func (rm *ReasoningManager) reasoningWorker(workerID int) {
	for task := range rm.taskQueue {
		select {
		case <-rm.ctx.Done():
			return
		default:
			result := rm.executeReasoningTask(task)
			if task.ResultChan != nil {
				task.ResultChan <- result
			}
		}
	}
}

// executeReasoningTask executes a single reasoning task
func (rm *ReasoningManager) executeReasoningTask(task ReasoningTask) ReasoningResult {
	rm.mu.Lock()
	rm.totalTasks++
	rm.mu.Unlock()
	
	// Create reasoning chain
	chain := &LangchainReasoningChain{
		ID:        task.ID,
		Goal:      task.Input,
		Mode:      task.Mode,
		Steps:     make([]ReasoningStep, 0),
		StartTime: time.Now(),
	}
	
	rm.mu.Lock()
	rm.activeChains[chain.ID] = chain
	rm.mu.Unlock()
	
	// Notify chain start
	if rm.onChainStart != nil {
		rm.onChainStart(chain)
	}
	
	// Create timeout context
	ctx, cancel := context.WithTimeout(rm.ctx, task.Timeout)
	defer cancel()
	
	// Execute reasoning loop
	var finalOutput string
	var err error
	
	for step := 0; step < task.MaxSteps; step++ {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			break
		default:
		}
		
		if err != nil {
			break
		}
		
		// Generate next reasoning step
		reasoningStep, finished, stepErr := rm.generateReasoningStep(ctx, chain, task)
		if stepErr != nil {
			err = stepErr
			break
		}
		
		// Add step to chain
		chain.Steps = append(chain.Steps, reasoningStep)
		
		rm.mu.Lock()
		rm.totalSteps++
		rm.mu.Unlock()
		
		// Notify step complete
		if rm.onStepComplete != nil {
			rm.onStepComplete(chain, reasoningStep)
		}
		
		// Check if reasoning is complete
		if finished {
			finalOutput = reasoningStep.Observation
			break
		}
	}
	
	// Complete chain
	chain.EndTime = time.Now()
	chain.Result = finalOutput
	chain.Success = err == nil && finalOutput != ""
	
	// Move to completed
	rm.mu.Lock()
	delete(rm.activeChains, chain.ID)
	rm.completedChains = append(rm.completedChains, chain)
	if chain.Success {
		rm.successfulTasks++
	} else {
		rm.failedTasks++
	}
	rm.mu.Unlock()
	
	// Notify chain complete
	if rm.onChainComplete != nil {
		rm.onChainComplete(chain)
	}
	
	return ReasoningResult{
		TaskID: task.ID,
		Chain:  chain,
		Output: finalOutput,
		Error:  err,
	}
}

// generateReasoningStep generates a single reasoning step using LLM
func (rm *ReasoningManager) generateReasoningStep(ctx context.Context, chain *LangchainReasoningChain, task ReasoningTask) (ReasoningStep, bool, error) {
	step := ReasoningStep{
		ID:        generateReasoningID("step"),
		Timestamp: time.Now(),
	}
	
	// Build prompt based on mode and current chain state
	prompt := rm.buildReasoningPrompt(chain, task)
	
	// Call LLM
	response, err := rm.llmProvider.Generate(ctx, prompt, llm.DefaultGenerateOptions())
	if err != nil {
		return step, false, fmt.Errorf("LLM completion failed: %w", err)
	}
	
	// Parse response to extract thought, action, and action input
	thought, action, actionInput, finished := rm.parseReasoningResponse(response)
	
	step.Thought = thought
	step.Action = action
	step.ActionInput = actionInput
	
	// If finished, return the final answer
	if finished {
		step.Observation = actionInput // Final answer is in actionInput for "Final Answer" action
		return step, true, nil
	}
	
	// Execute action using appropriate tool
	observation, err := rm.executeAction(ctx, action, actionInput)
	if err != nil {
		step.Observation = fmt.Sprintf("Error: %v", err)
	} else {
		step.Observation = observation
	}
	
	return step, false, nil
}

// buildReasoningPrompt builds the prompt for reasoning
func (rm *ReasoningManager) buildReasoningPrompt(chain *LangchainReasoningChain, task ReasoningTask) string {
	var sb strings.Builder
	
	// System context based on mode
	switch task.Mode {
	case ReasoningModeReactive:
		sb.WriteString("You are a reactive reasoning agent. Respond quickly and directly.\n\n")
	case ReasoningModeDeliberative:
		sb.WriteString("You are a deliberative reasoning agent. Think carefully through each step.\n\n")
	case ReasoningModeReflective:
		sb.WriteString("You are a reflective reasoning agent. Consider multiple perspectives and learn from the process.\n\n")
	case ReasoningModeCreative:
		sb.WriteString("You are a creative reasoning agent. Explore novel solutions and connections.\n\n")
	}
	
	// Available tools
	sb.WriteString("Available tools:\n")
	for name, tool := range rm.tools {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", name, tool.Description()))
	}
	sb.WriteString("\n")
	
	// Format instructions
	sb.WriteString("Use the following format:\n")
	sb.WriteString("Thought: your reasoning about what to do\n")
	sb.WriteString("Action: the tool to use (one of: ")
	toolNames := make([]string, 0, len(rm.tools))
	for name := range rm.tools {
		toolNames = append(toolNames, name)
	}
	sb.WriteString(strings.Join(toolNames, ", "))
	sb.WriteString(", Final Answer)\n")
	sb.WriteString("Action Input: the input to the tool\n")
	sb.WriteString("Observation: the result of the action\n\n")
	sb.WriteString("When you have the final answer, use:\n")
	sb.WriteString("Thought: I now have the final answer\n")
	sb.WriteString("Action: Final Answer\n")
	sb.WriteString("Action Input: [your final answer]\n\n")
	
	// Goal
	sb.WriteString(fmt.Sprintf("Goal: %s\n\n", chain.Goal))
	
	// Previous steps
	if len(chain.Steps) > 0 {
		sb.WriteString("Previous reasoning:\n")
		for _, s := range chain.Steps {
			sb.WriteString(fmt.Sprintf("Thought: %s\n", s.Thought))
			sb.WriteString(fmt.Sprintf("Action: %s\n", s.Action))
			sb.WriteString(fmt.Sprintf("Action Input: %s\n", s.ActionInput))
			sb.WriteString(fmt.Sprintf("Observation: %s\n\n", s.Observation))
		}
	}
	
	sb.WriteString("Continue reasoning:\n")
	
	return sb.String()
}

// parseReasoningResponse parses the LLM response to extract reasoning components
func (rm *ReasoningManager) parseReasoningResponse(response string) (thought, action, actionInput string, finished bool) {
	lines := strings.Split(response, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "Thought:") {
			thought = strings.TrimSpace(strings.TrimPrefix(line, "Thought:"))
		} else if strings.HasPrefix(line, "Action:") {
			action = strings.TrimSpace(strings.TrimPrefix(line, "Action:"))
		} else if strings.HasPrefix(line, "Action Input:") {
			actionInput = strings.TrimSpace(strings.TrimPrefix(line, "Action Input:"))
		}
	}
	
	// Check if this is the final answer
	if strings.ToUpper(action) == "FINAL ANSWER" {
		finished = true
	}
	
	return
}

// executeAction executes an action using the appropriate tool
func (rm *ReasoningManager) executeAction(ctx context.Context, action, input string) (string, error) {
	tool, exists := rm.tools[strings.ToUpper(action)]
	if !exists {
		return "", fmt.Errorf("unknown tool: %s", action)
	}
	
	return tool.Call(ctx, input)
}

// GetActiveChains returns currently active reasoning chains
func (rm *ReasoningManager) GetActiveChains() []*LangchainReasoningChain {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	chains := make([]*LangchainReasoningChain, 0, len(rm.activeChains))
	for _, chain := range rm.activeChains {
		chains = append(chains, chain)
	}
	return chains
}

// GetMetrics returns reasoning metrics
func (rm *ReasoningManager) GetMetrics() map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	return map[string]interface{}{
		"total_tasks":      rm.totalTasks,
		"successful_tasks": rm.successfulTasks,
		"failed_tasks":     rm.failedTasks,
		"total_steps":      rm.totalSteps,
		"active_chains":    len(rm.activeChains),
		"completed_chains": len(rm.completedChains),
	}
}

// ContributeToGestalt returns the reasoning manager's contribution to the global gestalt
func (rm *ReasoningManager) ContributeToGestalt() map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	successRate := float64(0)
	if rm.totalTasks > 0 {
		successRate = float64(rm.successfulTasks) / float64(rm.totalTasks)
	}
	
	avgSteps := float64(0)
	if rm.totalTasks > 0 {
		avgSteps = float64(rm.totalSteps) / float64(rm.totalTasks)
	}
	
	return map[string]interface{}{
		"subsystem":       "reasoning_manager",
		"running":         rm.running,
		"success_rate":    successRate,
		"avg_steps":       avgSteps,
		"active_chains":   len(rm.activeChains),
		"available_tools": len(rm.tools),
	}
}
