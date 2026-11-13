package echobeats

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// errNoInferenceModel reports that an engine has no model to infer with.
var (
	errNoInferenceModel    = errors.New("inference engine has no model")
	errModelReturnedNoText = errors.New("model returned no text")
)

// InferenceEngine represents a concurrent inference engine
// EchoBeats runs 3 of these in parallel
type InferenceEngine struct {
	mu     sync.RWMutex
	ctx    context.Context //nolint:containedctx // lifecycle context created with cancel in constructor; cancelled on Stop to end goroutines
	cancel context.CancelFunc

	// Identity
	id   int
	name string

	// Cognitive loop
	cognitiveLoop *CognitiveLoop

	// Processing state
	currentTask    *InferenceTask
	taskQueue      []*InferenceTask
	completedTasks []*InferenceTask
	maxQueueSize   int

	// Specialization
	specialization InferenceSpecialization

	// Metrics
	tasksProcessed  uint64
	totalInferences uint64

	// gen is the model that performs inference; nil means tasks fail.
	gen TextGenerator

	// Control
	running bool
	paused  bool
}

// SetTextGenerator injects the model this engine infers with. Without one,
// tasks complete with Success=false and no output.
func (ie *InferenceEngine) SetTextGenerator(gen TextGenerator) {
	ie.mu.Lock()
	ie.gen = gen
	ie.mu.Unlock()
	if ie.cognitiveLoop != nil {
		ie.cognitiveLoop.SetTextGenerator(gen)
	}
}

// InferenceSpecialization defines what the engine specializes in
type InferenceSpecialization string

const (
	SpecializationPerception InferenceSpecialization = "perception" // Sensory and perceptual processing
	SpecializationCognition  InferenceSpecialization = "cognition"  // Reasoning and problem-solving
	SpecializationAction     InferenceSpecialization = "action"     // Action planning and execution
)

// InferenceTask represents a task for inference
type InferenceTask struct {
	ID          string
	Type        string
	Input       interface{}
	Context     map[string]interface{}
	Priority    float64
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	Result      *InferenceResult
	Error       error
}

// InferenceResult contains the result of inference
type InferenceResult struct {
	Success        bool
	Output         interface{}
	Confidence     float64
	ProcessingTime time.Duration
	Insights       []string
	NextActions    []string
}

// NewInferenceEngine creates a new inference engine
func NewInferenceEngine(id int, specialization InferenceSpecialization) *InferenceEngine {
	ctx, cancel := context.WithCancel(context.Background())

	name := fmt.Sprintf("InferenceEngine-%d-%s", id, specialization)

	ie := &InferenceEngine{
		ctx:            ctx,
		cancel:         cancel,
		id:             id,
		name:           name,
		cognitiveLoop:  NewCognitiveLoop(),
		taskQueue:      make([]*InferenceTask, 0),
		completedTasks: make([]*InferenceTask, 0),
		maxQueueSize:   100,
		specialization: specialization,
	}

	// Configure cognitive loop for this engine
	ie.cognitiveLoop.SetStepDuration(1 * time.Second)

	return ie
}

// Start begins the inference engine
func (ie *InferenceEngine) Start() error {
	ie.mu.Lock()
	if ie.running {
		ie.mu.Unlock()
		return fmt.Errorf("inference engine already running")
	}
	ie.running = true
	ie.mu.Unlock()

	fmt.Printf("🧠 %s: Starting (specialization: %s)...\n", ie.name, ie.specialization)

	// Start cognitive loop
	if err := ie.cognitiveLoop.Start(); err != nil {
		return fmt.Errorf("failed to start cognitive loop: %w", err)
	}

	// Start task processing
	go ie.processTaskQueue()

	return nil
}

// Stop gracefully stops the inference engine
func (ie *InferenceEngine) Stop() error {
	ie.mu.Lock()
	defer ie.mu.Unlock()

	if !ie.running {
		return fmt.Errorf("inference engine not running")
	}

	fmt.Printf("🧠 %s: Stopping...\n", ie.name)
	ie.running = false
	ie.cancel()

	// Stop cognitive loop
	if err := ie.cognitiveLoop.Stop(); err != nil {
		fmt.Printf("⚠️  %s: Error stopping cognitive loop: %v\n", ie.name, err)
	}

	return nil
}

// Pause pauses the inference engine
func (ie *InferenceEngine) Pause() {
	ie.mu.Lock()
	defer ie.mu.Unlock()
	ie.paused = true
	ie.cognitiveLoop.Pause()
	fmt.Printf("⏸️  %s: Paused\n", ie.name)
}

// Resume resumes the inference engine
func (ie *InferenceEngine) Resume() {
	ie.mu.Lock()
	defer ie.mu.Unlock()
	ie.paused = false
	ie.cognitiveLoop.Resume()
	fmt.Printf("▶️  %s: Resumed\n", ie.name)
}

// SubmitTask submits a task for inference
func (ie *InferenceEngine) SubmitTask(task *InferenceTask) error {
	ie.mu.Lock()
	defer ie.mu.Unlock()

	if len(ie.taskQueue) >= ie.maxQueueSize {
		return fmt.Errorf("task queue full")
	}

	task.CreatedAt = time.Now()
	ie.taskQueue = append(ie.taskQueue, task)

	// Sort by priority (higher first)
	ie.sortTaskQueue()

	return nil
}

// sortTaskQueue sorts tasks by priority
func (ie *InferenceEngine) sortTaskQueue() {
	// Simple bubble sort for small queues
	n := len(ie.taskQueue)
	for i := range n - 1 {
		for j := range n - i - 1 {
			if ie.taskQueue[j].Priority < ie.taskQueue[j+1].Priority {
				ie.taskQueue[j], ie.taskQueue[j+1] = ie.taskQueue[j+1], ie.taskQueue[j]
			}
		}
	}
}

// processTaskQueue processes tasks from the queue
func (ie *InferenceEngine) processTaskQueue() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ie.ctx.Done():
			return
		case <-ticker.C:
			ie.mu.RLock()
			isPaused := ie.paused
			queueLen := len(ie.taskQueue)
			ie.mu.RUnlock()

			if !isPaused && queueLen > 0 {
				ie.processNextTask()
			}
		}
	}
}

// processNextTask processes the next task in the queue
func (ie *InferenceEngine) processNextTask() {
	ie.mu.Lock()
	if len(ie.taskQueue) == 0 {
		ie.mu.Unlock()
		return
	}

	// Get highest priority task
	task := ie.taskQueue[0]
	ie.taskQueue = ie.taskQueue[1:]
	ie.currentTask = task
	ie.mu.Unlock()

	// Process task
	startTime := time.Now()
	now := time.Now()
	task.StartedAt = &now

	result := ie.performInference(task)

	processingTime := time.Since(startTime)
	result.ProcessingTime = processingTime

	// Update task
	completedTime := time.Now()
	task.CompletedAt = &completedTime
	task.Result = result

	ie.mu.Lock()
	ie.completedTasks = append(ie.completedTasks, task)
	ie.currentTask = nil
	ie.tasksProcessed++
	ie.totalInferences++
	ie.mu.Unlock()

	if result.Success {
		fmt.Printf("🔍 %s: Completed task %s (%.2fs)\n", ie.name, task.Type, processingTime.Seconds())
	} else {
		fmt.Printf("🔍 %s: Task %s produced no inference: %v\n", ie.name, task.Type, task.Error)
	}
}

// performInference asks the model to carry out the task for this engine's
// specialization, given the task and the engine's real cognitive state. No
// result is invented: without a model, or when it fails or answers with
// nothing, the task records the error and Success=false.
func (ie *InferenceEngine) performInference(task *InferenceTask) *InferenceResult {
	ie.mu.RLock()
	gen := ie.gen
	ctx := ie.ctx
	ie.mu.RUnlock()
	if gen == nil {
		task.Error = errNoInferenceModel
		return &InferenceResult{}
	}

	cogState := ie.cognitiveLoop.GetCurrentState()
	facts := fmt.Sprintf("Task type: %s. Input: %v. Context: %v.", task.Type, task.Input, task.Context)
	if cogState != nil {
		facts += fmt.Sprintf(" Attention: %v. Working memory items: %d. Cognitive load: %.2f. Pending actions: %v.",
			cogState.Attention, len(cogState.WorkingMemory), cogState.CognitiveLoad, cogState.PendingActions)
	}
	var instruction string
	switch ie.specialization {
	case SpecializationPerception:
		instruction = "As Deep Tree Echo's perception engine, describe in one or two sentences what is salient in this input and why."
	case SpecializationAction:
		instruction = "As Deep Tree Echo's action engine, propose in one or two sentences the next concrete action and its expected effect."
	default:
		instruction = "As Deep Tree Echo's reasoning engine, state in one or two sentences what you infer from this input."
	}

	out, err := gen.Generate(ctx, facts+"\n"+instruction, llm.GenerateOptions{MaxTokens: 120, Temperature: 0.6})
	out = strings.TrimSpace(out)
	if err == nil && out == "" {
		err = errModelReturnedNoText
	}
	if err != nil {
		task.Error = err
		return &InferenceResult{}
	}
	return &InferenceResult{Success: true, Output: out}
}

// GetMetrics returns inference engine metrics
func (ie *InferenceEngine) GetMetrics() map[string]interface{} {
	ie.mu.RLock()
	defer ie.mu.RUnlock()

	return map[string]interface{}{
		"id":               ie.id,
		"name":             ie.name,
		"specialization":   ie.specialization,
		"running":          ie.running,
		"paused":           ie.paused,
		"tasks_processed":  ie.tasksProcessed,
		"total_inferences": ie.totalInferences,
		"queue_length":     len(ie.taskQueue),
		"completed_tasks":  len(ie.completedTasks),
		"current_task":     ie.currentTask != nil,
	}
}

// GetCognitiveState returns the current cognitive state
func (ie *InferenceEngine) GetCognitiveState() *CognitiveState {
	return ie.cognitiveLoop.GetCurrentState()
}

// GetQueueLength returns the current queue length
func (ie *InferenceEngine) GetQueueLength() int {
	ie.mu.RLock()
	defer ie.mu.RUnlock()
	return len(ie.taskQueue)
}
