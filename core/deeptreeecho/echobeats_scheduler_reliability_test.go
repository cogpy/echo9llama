package deeptreeecho

import (
	"context"
	"errors"
	"math"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

type schedulerFixedProvider struct {
	mu       sync.Mutex
	response string
	err      error
	calls    int
}

func (provider *schedulerFixedProvider) Generate(ctx context.Context, _ string, _ llm.GenerateOptions) (string, error) {
	provider.mu.Lock()
	provider.calls++
	provider.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if provider.err != nil {
		return "", provider.err
	}
	return provider.response, nil
}

func (provider *schedulerFixedProvider) StreamGenerate(context.Context, string, llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	return nil, nil
}
func (provider *schedulerFixedProvider) Name() string    { return "scheduler-fixed" }
func (provider *schedulerFixedProvider) Available() bool { return true }
func (provider *schedulerFixedProvider) MaxTokens() int  { return 4096 }

type schedulerBlockingProvider struct {
	started chan struct{}
	once    sync.Once
}

func (provider *schedulerBlockingProvider) Generate(ctx context.Context, _ string, _ llm.GenerateOptions) (string, error) {
	provider.once.Do(func() { close(provider.started) })
	<-ctx.Done()
	return "", ctx.Err()
}

func (provider *schedulerBlockingProvider) StreamGenerate(context.Context, string, llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	return nil, nil
}
func (provider *schedulerBlockingProvider) Name() string    { return "scheduler-blocking" }
func (provider *schedulerBlockingProvider) Available() bool { return true }
func (provider *schedulerBlockingProvider) MaxTokens() int  { return 4096 }

func latestEngineTask(t *testing.T, engine *InferenceEngine) CognitiveTask {
	t.Helper()
	engine.mu.RLock()
	defer engine.mu.RUnlock()
	if len(engine.taskHistory) == 0 {
		t.Fatal("engine has no recorded task")
	}
	return engine.taskHistory[len(engine.taskHistory)-1]
}

func TestEchobeatsSchedulerBoundsEngineTaskHistory(t *testing.T) {
	scheduler := NewEchobeatsScheduler(&schedulerFixedProvider{response: "grounded observation"})
	for range inferenceEngineTaskHistoryLimit + 5 {
		scheduler.relevanceRealization("retain only bounded task evidence")
	}

	scheduler.engine1.mu.RLock()
	defer scheduler.engine1.mu.RUnlock()
	if got := len(scheduler.engine1.taskHistory); got != inferenceEngineTaskHistoryLimit {
		t.Fatalf("engine task history length = %d, want %d", got, inferenceEngineTaskHistoryLimit)
	}
	if scheduler.engine1.taskHistory[0].CompletionTime == nil || scheduler.engine1.taskHistory[inferenceEngineTaskHistoryLimit-1].CompletionTime == nil {
		t.Fatal("bounded task history did not retain completed task evidence")
	}
}

func TestEchobeatsSchedulerRecordsProviderFailuresWithoutSyntheticThought(t *testing.T) {
	provider := &schedulerFixedProvider{err: errors.New("provider unavailable")}
	scheduler := NewEchobeatsScheduler(provider)

	scheduler.relevanceRealization("focus")
	scheduler.affordanceInteraction(3)
	scheduler.salienceSimulation(3)

	for _, engine := range []*InferenceEngine{scheduler.engine1, scheduler.engine2, scheduler.engine3} {
		task := latestEngineTask(t, engine)
		if task.Success {
			t.Fatalf("provider failure was marked successful: %#v", task)
		}
		if task.Result != "generation failed: provider unavailable" {
			t.Fatalf("provider failure result = %q, want explicit failure record", task.Result)
		}
		if strings.Contains(task.Result, "in progress") || strings.Contains(task.Result, "under consideration") {
			t.Fatalf("provider failure retained success-like fallback prose: %#v", task)
		}
	}

	scheduler.mu.RLock()
	defer scheduler.mu.RUnlock()
	if scheduler.presentCommitment != "" || len(scheduler.pastPerformance) != 0 || len(scheduler.futureAnticipation) != 0 {
		t.Fatalf("failed provider output entered cognitive state: commitment=%q past=%#v future=%#v", scheduler.presentCommitment, scheduler.pastPerformance, scheduler.futureAnticipation)
	}
}

func TestEchobeatsSchedulerReflectionHookRunsAtStepEightOutsideSchedulerMutex(t *testing.T) {
	scheduler := NewEchobeatsScheduler(&schedulerFixedProvider{response: "bounded simulation"})
	var reflectionCalls int
	scheduler.SetOnReflection(func(step int) error {
		if step != 8 {
			t.Errorf("reflection step = %d, want 8", step)
		}
		// This takes sched.mu. It proves user code is not invoked while the
		// scheduler mutex is held, while executeStep still holds stepMu.
		scheduler.SetOnAffordance(nil)
		reflectionCalls++
		return errors.New("evidence unavailable")
	})

	if err := scheduler.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	t.Cleanup(func() {
		if scheduler.IsRunning() {
			if err := scheduler.Stop(); err != nil {
				t.Errorf("Stop() error: %v", err)
			}
		}
	})

	for range 24 {
		scheduler.executeStep()
	}
	if reflectionCalls != 2 {
		t.Fatalf("reflection callback calls = %d, want one callback per cycle", reflectionCalls)
	}
}

func TestEchobeatsSchedulerStopWaitsForContextAwareStepAndRestartGetsFreshContext(t *testing.T) {
	blockingProvider := &schedulerBlockingProvider{started: make(chan struct{})}
	scheduler := NewEchobeatsScheduler(blockingProvider)
	if err := scheduler.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	stepDone := make(chan struct{})
	go func() {
		scheduler.executeStep()
		close(stepDone)
	}()
	select {
	case <-blockingProvider.started:
	case <-time.After(time.Second):
		t.Fatal("provider call did not start")
	}

	stopDone := make(chan error, 1)
	go func() { stopDone <- scheduler.Stop() }()
	select {
	case err := <-stopDone:
		if err != nil {
			t.Fatalf("Stop() error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Stop() did not cancel and quiesce the active provider call")
	}
	select {
	case <-stepDone:
	case <-time.After(time.Second):
		t.Fatal("active step remained after Stop() returned")
	}
	if scheduler.IsRunning() {
		t.Fatal("scheduler still reports running after Stop()")
	}

	restartingProvider := &schedulerFixedProvider{response: "fresh context response"}
	restartingScheduler := NewEchobeatsScheduler(restartingProvider)
	if err := restartingScheduler.Start(); err != nil {
		t.Fatalf("initial Start() error: %v", err)
	}
	if err := restartingScheduler.Stop(); err != nil {
		t.Fatalf("initial Stop() error: %v", err)
	}
	if err := restartingScheduler.Start(); err != nil {
		t.Fatalf("restart Start() error: %v", err)
	}
	t.Cleanup(func() {
		if restartingScheduler.IsRunning() {
			if err := restartingScheduler.Stop(); err != nil {
				t.Errorf("restart Stop() error: %v", err)
			}
		}
	})
	restartingScheduler.executeStep()
	if task := latestEngineTask(t, restartingScheduler.engine1); !task.Success || task.Result != "fresh context response" {
		t.Fatalf("restart did not use a fresh live context: %#v", task)
	}
}

func TestEchobeatsSchedulerEnsureGoalIsIdempotentAndValidated(t *testing.T) {
	scheduler := NewEchobeatsScheduler(&schedulerFixedProvider{response: "unused"})
	if !scheduler.EnsureGoal("reflection-e4", "Persist reflection replay evidence", 0.75) {
		t.Fatal("EnsureGoal() rejected valid explicit-ID goal")
	}
	if !scheduler.EnsureGoal("reflection-e4", "later replay must not replace the original", 0.1) {
		t.Fatal("EnsureGoal() did not treat existing ID as already ensured")
	}

	queue := scheduler.GetGoalQueue()
	if len(queue) != 1 {
		t.Fatalf("goal queue length = %d, want 1 after idempotent replay", len(queue))
	}
	goal := queue[0]
	if goal.ID != "reflection-e4" || goal.Description != "Persist reflection replay evidence" || goal.Priority != 0.75 || goal.Progress != 0 || goal.Status != GoalPending {
		t.Fatalf("unexpected ensured goal: %#v", goal)
	}

	for _, invalid := range []struct {
		id, description string
		priority        float64
	}{
		{"", "description", 0.5},
		{"goal", "", 0.5},
		{"goal", "description", -0.01},
		{"goal", "description", 1.01},
		{"goal", "description", math.NaN()},
		{"goal", "description", math.Inf(1)},
	} {
		if scheduler.EnsureGoal(invalid.id, invalid.description, invalid.priority) {
			t.Fatalf("EnsureGoal() accepted invalid input: %#v", invalid)
		}
	}
}

func TestEchobeatsSchedulerSnapshotsNestedGoalsAndLocksEngineMetrics(t *testing.T) {
	scheduler := NewEchobeatsScheduler(&schedulerFixedProvider{response: "ok"})
	goalID := scheduler.AddGoal("defensive queue snapshot", 1)
	deadline := time.Now().Add(time.Hour)
	scheduler.mu.Lock()
	scheduler.goalQueue[0].Deadline = &deadline
	scheduler.goalQueue[0].Dependencies = []string{"durable evidence"}
	scheduler.mu.Unlock()

	queue := scheduler.GetGoalQueue()
	if len(queue) != 1 || queue[0].ID != goalID {
		t.Fatalf("unexpected goal queue snapshot: %#v", queue)
	}
	queue[0].Dependencies[0] = "mutated snapshot"
	*queue[0].Deadline = queue[0].Deadline.Add(time.Hour)

	scheduler.mu.RLock()
	internal := scheduler.goalQueue[0]
	scheduler.mu.RUnlock()
	if internal.Dependencies[0] != "durable evidence" || !internal.Deadline.Equal(deadline) {
		t.Fatalf("nested goal fields leaked through snapshot: %#v", internal)
	}

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		for i := range 1000 {
			scheduler.engine1.mu.Lock()
			scheduler.engine1.performance = float64(i) / 1000
			scheduler.engine1.mu.Unlock()
		}
	}()
	for range 1000 {
		_ = scheduler.GetMetrics()
		_ = scheduler.GetCompleteCycleMetrics()
	}
	<-writerDone
}
