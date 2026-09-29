// Package void implements the orchestra - the coordination system
package void

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Orchestra coordinates all event loops and manages the 12-step cognitive cycle.
type Orchestra struct {
	// Scheduling
	scheduler  *Scheduler
	eventLoops map[string]*EventLoop

	// Protocols
	protocols map[string]*Protocol

	// Timing
	stepDuration time.Duration
	currentStep  int

	// State
	running bool

	// Synchronization
	mu       sync.RWMutex
	stepChan chan int
	stopChan chan struct{}
}

// Scheduler manages scheduled tasks.
type Scheduler struct {
	tasks []*ScheduledTask
	mu    sync.RWMutex
}

// ScheduledTask represents a task scheduled for a specific step.
type ScheduledTask struct {
	ID       string
	CoreID   string
	Step     int
	Priority int
	Callback func(context.Context) error
}

// EventLoop represents an event loop for a core.
type EventLoop struct {
	ID        string
	CoreID    string
	Frequency time.Duration
	Handler   func(context.Context) error
	running   bool
	stopChan  chan struct{}
}

// Protocol defines a communication protocol.
type Protocol struct {
	ID        string
	Name      string
	Version   string
	Schema    interface{}
	Validator func(interface{}) error
}

// NewOrchestra creates a new orchestra.
func NewOrchestra(stepDuration time.Duration) *Orchestra {
	return &Orchestra{
		scheduler:    NewScheduler(),
		eventLoops:   make(map[string]*EventLoop),
		protocols:    make(map[string]*Protocol),
		stepDuration: stepDuration,
		currentStep:  0,
		running:      false,
		stepChan:     make(chan int, 100),
		stopChan:     make(chan struct{}),
	}
}

// Start starts the orchestra.
func (o *Orchestra) Start(ctx context.Context) error {
	o.mu.Lock()
	if o.running {
		o.mu.Unlock()
		return fmt.Errorf("orchestra already running")
	}
	o.running = true
	o.mu.Unlock()

	// Start all event loops
	for _, loop := range o.eventLoops {
		go o.runEventLoop(ctx, loop)
	}

	return nil
}

// Stop stops the orchestra.
func (o *Orchestra) Stop() error {
	o.mu.Lock()
	if !o.running {
		o.mu.Unlock()
		return fmt.Errorf("orchestra not running")
	}
	o.running = false
	o.mu.Unlock()

	// Signal stop
	close(o.stopChan)

	// Stop all event loops
	for _, loop := range o.eventLoops {
		if loop.running {
			close(loop.stopChan)
		}
	}

	return nil
}

// runEventLoop runs an event loop.
func (o *Orchestra) runEventLoop(ctx context.Context, loop *EventLoop) {
	loop.running = true
	ticker := time.NewTicker(loop.Frequency)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-loop.stopChan:
			return
		case <-ticker.C:
			if err := loop.Handler(ctx); err != nil {
				fmt.Printf("Error in event loop %s: %v\n", loop.ID, err)
			}
		}
	}
}

// AdvanceStep advances to the next step in the cognitive cycle.
func (o *Orchestra) AdvanceStep() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.currentStep = (o.currentStep + 1) % 12

	// Notify step change
	select {
	case o.stepChan <- o.currentStep:
	default:
		// Channel full, skip notification
	}

	// Execute scheduled tasks for this step
	tasks := o.scheduler.GetTasksForStep(o.currentStep)
	for _, task := range tasks {
		go func(t *ScheduledTask) {
			ctx := context.Background()
			if err := t.Callback(ctx); err != nil {
				fmt.Printf("Error executing task %s: %v\n", t.ID, err)
			}
		}(task)
	}

	return nil
}

// GetCurrentStep returns the current step.
func (o *Orchestra) GetCurrentStep() int {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.currentStep
}

// RegisterEventLoop registers an event loop.
func (o *Orchestra) RegisterEventLoop(loop *EventLoop) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, exists := o.eventLoops[loop.ID]; exists {
		return fmt.Errorf("event loop %s already registered", loop.ID)
	}

	loop.stopChan = make(chan struct{})
	o.eventLoops[loop.ID] = loop

	// If orchestra is already running, start the loop
	if o.running {
		go o.runEventLoop(context.Background(), loop)
	}

	return nil
}

// RegisterProtocol registers a protocol.
func (o *Orchestra) RegisterProtocol(protocol *Protocol) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, exists := o.protocols[protocol.ID]; exists {
		return fmt.Errorf("protocol %s already registered", protocol.ID)
	}

	o.protocols[protocol.ID] = protocol
	return nil
}

// ScheduleTask schedules a task for a specific step.
func (o *Orchestra) ScheduleTask(task *ScheduledTask) error {
	return o.scheduler.AddTask(task)
}

// GetProtocol retrieves a protocol by ID.
func (o *Orchestra) GetProtocol(id string) (*Protocol, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	protocol, exists := o.protocols[id]
	return protocol, exists
}

// NewScheduler creates a new scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{
		tasks: make([]*ScheduledTask, 0),
	}
}

// AddTask adds a task to the scheduler.
func (s *Scheduler) AddTask(task *ScheduledTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks = append(s.tasks, task)
	return nil
}

// GetTasksForStep retrieves all tasks scheduled for a specific step.
func (s *Scheduler) GetTasksForStep(step int) []*ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*ScheduledTask, 0)
	for _, task := range s.tasks {
		if task.Step == step {
			result = append(result, task)
		}
	}
	return result
}

// RemoveTask removes a task from the scheduler.
func (s *Scheduler) RemoveTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, task := range s.tasks {
		if task.ID == taskID {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("task %s not found", taskID)
}
