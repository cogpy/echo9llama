package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CognitiveScheduler provides advanced scheduling for Deep Tree Echo cognitive cycles
// This is inspired by gocron's architecture for flexible job scheduling
type CognitiveScheduler struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc

	// Scheduled jobs
	jobs            map[uuid.UUID]*CognitiveJob
	jobsLock        sync.RWMutex

	// Job channels
	newJobCh        chan *CognitiveJob
	removeJobCh     chan uuid.UUID
	runJobCh        chan uuid.UUID

	// Configuration
	location        *time.Location
	defaultTimeout  time.Duration

	// Metrics
	totalJobsRun    uint64
	totalJobsFailed uint64

	// Running state
	running         bool
	started         bool
}

// CognitiveJob represents a scheduled cognitive task
type CognitiveJob struct {
	ID              uuid.UUID
	Name            string
	Description     string
	Definition      JobDefinition
	Task            func(ctx context.Context) error
	Tags            []string
	Priority        float64
	
	// Scheduling state
	NextRun         time.Time
	LastRun         time.Time
	RunCount        uint64
	FailCount       uint64
	
	// Job options
	Timeout         time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
	
	// Context
	ctx             context.Context
	cancel          context.CancelFunc
}

// JobDefinition defines when a job should run
type JobDefinition interface {
	NextRun(lastRun time.Time, now time.Time) time.Time
	String() string
}

// DurationJobDefinition runs a job at fixed intervals
type DurationJobDefinition struct {
	Duration time.Duration
}

func (d DurationJobDefinition) NextRun(lastRun time.Time, now time.Time) time.Time {
	if lastRun.IsZero() {
		return now.Add(d.Duration)
	}
	return lastRun.Add(d.Duration)
}

func (d DurationJobDefinition) String() string {
	return fmt.Sprintf("every %s", d.Duration)
}

// CronJobDefinition runs a job based on cron expression
type CronJobDefinition struct {
	Expression string
	// Simplified cron fields (minute, hour, day, month, weekday)
	Minute     int
	Hour       int
	Day        int
	Month      int
	Weekday    int
}

func (c CronJobDefinition) NextRun(lastRun time.Time, now time.Time) time.Time {
	// Simplified cron calculation - just use hour and minute
	next := time.Date(now.Year(), now.Month(), now.Day(), c.Hour, c.Minute, 0, 0, now.Location())
	if next.Before(now) || next.Equal(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

func (c CronJobDefinition) String() string {
	return fmt.Sprintf("cron(%s)", c.Expression)
}

// CognitivePhaseJobDefinition runs a job based on cognitive cycle phases
type CognitivePhaseJobDefinition struct {
	Phase       int // 1-12 for the 12-step cognitive cycle
	CycleLength time.Duration
}

func (c CognitivePhaseJobDefinition) NextRun(lastRun time.Time, now time.Time) time.Time {
	phaseOffset := time.Duration(c.Phase-1) * (c.CycleLength / 12)
	cycleStart := now.Truncate(c.CycleLength)
	nextPhase := cycleStart.Add(phaseOffset)
	if nextPhase.Before(now) || nextPhase.Equal(now) {
		nextPhase = nextPhase.Add(c.CycleLength)
	}
	return nextPhase
}

func (c CognitivePhaseJobDefinition) String() string {
	return fmt.Sprintf("cognitive_phase(%d)", c.Phase)
}

// WakeRestJobDefinition runs a job based on wake/rest cycles
type WakeRestJobDefinition struct {
	WakeTime  time.Duration // Time after midnight to wake
	RestTime  time.Duration // Time after midnight to rest
	RunDuring string        // "wake" or "rest"
	Interval  time.Duration // Interval during the active period
}

func (w WakeRestJobDefinition) NextRun(lastRun time.Time, now time.Time) time.Time {
	// Calculate today's wake and rest times
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	wakeTime := midnight.Add(w.WakeTime)
	restTime := midnight.Add(w.RestTime)

	isAwake := now.After(wakeTime) && now.Before(restTime)

	if w.RunDuring == "wake" && isAwake {
		if lastRun.IsZero() || lastRun.Before(wakeTime) {
			return now.Add(w.Interval)
		}
		return lastRun.Add(w.Interval)
	} else if w.RunDuring == "rest" && !isAwake {
		if lastRun.IsZero() || lastRun.Before(restTime) {
			return now.Add(w.Interval)
		}
		return lastRun.Add(w.Interval)
	}

	// Wait for the appropriate period
	if w.RunDuring == "wake" {
		if now.Before(wakeTime) {
			return wakeTime
		}
		return wakeTime.Add(24 * time.Hour)
	}
	if now.Before(restTime) {
		return restTime
	}
	return restTime.Add(24 * time.Hour)
}

func (w WakeRestJobDefinition) String() string {
	return fmt.Sprintf("wake_rest(%s, every %s)", w.RunDuring, w.Interval)
}

// NewCognitiveScheduler creates a new cognitive scheduler
func NewCognitiveScheduler() *CognitiveScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &CognitiveScheduler{
		ctx:            ctx,
		cancel:         cancel,
		jobs:           make(map[uuid.UUID]*CognitiveJob),
		newJobCh:       make(chan *CognitiveJob, 100),
		removeJobCh:    make(chan uuid.UUID, 100),
		runJobCh:       make(chan uuid.UUID, 100),
		location:       time.Local,
		defaultTimeout: 30 * time.Second,
	}
}

// Start begins the cognitive scheduler
func (cs *CognitiveScheduler) Start() error {
	cs.mu.Lock()
	if cs.running {
		cs.mu.Unlock()
		return fmt.Errorf("cognitive scheduler already running")
	}
	cs.running = true
	cs.started = true
	cs.mu.Unlock()

	fmt.Println("⏰ Cognitive Scheduler started")

	// Start the scheduler loop
	go cs.runSchedulerLoop()

	return nil
}

// Stop gracefully stops the cognitive scheduler
func (cs *CognitiveScheduler) Stop() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if !cs.running {
		return fmt.Errorf("cognitive scheduler not running")
	}

	cs.cancel()
	cs.running = false
	fmt.Println("⏰ Cognitive Scheduler stopped")

	return nil
}

// runSchedulerLoop is the main scheduler loop
func (cs *CognitiveScheduler) runSchedulerLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-cs.ctx.Done():
			return

		case job := <-cs.newJobCh:
			cs.jobsLock.Lock()
			cs.jobs[job.ID] = job
			cs.jobsLock.Unlock()

		case jobID := <-cs.removeJobCh:
			cs.jobsLock.Lock()
			if job, exists := cs.jobs[jobID]; exists {
				if job.cancel != nil {
					job.cancel()
				}
				delete(cs.jobs, jobID)
			}
			cs.jobsLock.Unlock()

		case jobID := <-cs.runJobCh:
			cs.jobsLock.RLock()
			if job, exists := cs.jobs[jobID]; exists {
				go cs.executeJob(job)
			}
			cs.jobsLock.RUnlock()

		case <-ticker.C:
			cs.checkAndRunJobs()
		}
	}
}

// checkAndRunJobs checks all jobs and runs those that are due
func (cs *CognitiveScheduler) checkAndRunJobs() {
	now := time.Now()

	cs.jobsLock.RLock()
	jobsToRun := make([]*CognitiveJob, 0)
	for _, job := range cs.jobs {
		if !job.NextRun.IsZero() && job.NextRun.Before(now) {
			jobsToRun = append(jobsToRun, job)
		}
	}
	cs.jobsLock.RUnlock()

	for _, job := range jobsToRun {
		go cs.executeJob(job)
	}
}

// executeJob executes a single job
func (cs *CognitiveScheduler) executeJob(job *CognitiveJob) {
	// Create job context with timeout
	ctx, cancel := context.WithTimeout(cs.ctx, job.Timeout)
	defer cancel()

	// Update job state
	cs.jobsLock.Lock()
	job.LastRun = time.Now()
	job.NextRun = job.Definition.NextRun(job.LastRun, time.Now())
	cs.jobsLock.Unlock()

	// Execute the task
	err := job.Task(ctx)

	cs.jobsLock.Lock()
	job.RunCount++
	if err != nil {
		job.FailCount++
		cs.mu.Lock()
		cs.totalJobsFailed++
		cs.mu.Unlock()
	}
	cs.jobsLock.Unlock()

	cs.mu.Lock()
	cs.totalJobsRun++
	cs.mu.Unlock()
}

// NewJob creates a new cognitive job
func (cs *CognitiveScheduler) NewJob(name string, definition JobDefinition, task func(ctx context.Context) error, options ...JobOption) (*CognitiveJob, error) {
	ctx, cancel := context.WithCancel(cs.ctx)

	job := &CognitiveJob{
		ID:          uuid.New(),
		Name:        name,
		Definition:  definition,
		Task:        task,
		Tags:        []string{},
		Priority:    0.5,
		NextRun:     definition.NextRun(time.Time{}, time.Now()),
		Timeout:     cs.defaultTimeout,
		MaxRetries:  3,
		RetryDelay:  time.Second,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Apply options
	for _, opt := range options {
		opt(job)
	}

	// Add to scheduler
	cs.newJobCh <- job

	return job, nil
}

// JobOption is a function that configures a CognitiveJob
type JobOption func(*CognitiveJob)

// WithTags sets tags for the job
func WithTags(tags ...string) JobOption {
	return func(j *CognitiveJob) {
		j.Tags = tags
	}
}

// WithPriority sets the priority for the job
func WithPriority(priority float64) JobOption {
	return func(j *CognitiveJob) {
		j.Priority = priority
	}
}

// WithTimeout sets the timeout for the job
func WithTimeout(timeout time.Duration) JobOption {
	return func(j *CognitiveJob) {
		j.Timeout = timeout
	}
}

// WithDescription sets the description for the job
func WithDescription(desc string) JobOption {
	return func(j *CognitiveJob) {
		j.Description = desc
	}
}

// RemoveJob removes a job by ID
func (cs *CognitiveScheduler) RemoveJob(id uuid.UUID) error {
	cs.removeJobCh <- id
	return nil
}

// RemoveByTags removes all jobs with any of the specified tags
func (cs *CognitiveScheduler) RemoveByTags(tags ...string) {
	cs.jobsLock.RLock()
	toRemove := make([]uuid.UUID, 0)
	for id, job := range cs.jobs {
		for _, tag := range tags {
			for _, jobTag := range job.Tags {
				if tag == jobTag {
					toRemove = append(toRemove, id)
					break
				}
			}
		}
	}
	cs.jobsLock.RUnlock()

	for _, id := range toRemove {
		cs.removeJobCh <- id
	}
}

// RunJobNow runs a job immediately
func (cs *CognitiveScheduler) RunJobNow(id uuid.UUID) error {
	cs.runJobCh <- id
	return nil
}

// Jobs returns all jobs
func (cs *CognitiveScheduler) Jobs() []*CognitiveJob {
	cs.jobsLock.RLock()
	defer cs.jobsLock.RUnlock()

	jobs := make([]*CognitiveJob, 0, len(cs.jobs))
	for _, job := range cs.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// GetMetrics returns scheduler metrics
func (cs *CognitiveScheduler) GetMetrics() map[string]interface{} {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	cs.jobsLock.RLock()
	jobCount := len(cs.jobs)
	cs.jobsLock.RUnlock()

	return map[string]interface{}{
		"running":          cs.running,
		"job_count":        jobCount,
		"total_jobs_run":   cs.totalJobsRun,
		"total_jobs_failed": cs.totalJobsFailed,
	}
}

// ContributeToGestalt provides scheduler state for the global gestalt
func (cs *CognitiveScheduler) ContributeToGestalt() map[string]interface{} {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	cs.jobsLock.RLock()
	upcomingJobs := make([]map[string]interface{}, 0)
	for _, job := range cs.jobs {
		if !job.NextRun.IsZero() {
			upcomingJobs = append(upcomingJobs, map[string]interface{}{
				"name":     job.Name,
				"next_run": job.NextRun.Format(time.RFC3339),
				"priority": job.Priority,
			})
		}
	}
	cs.jobsLock.RUnlock()

	return map[string]interface{}{
		"running":        cs.running,
		"upcoming_jobs":  upcomingJobs,
		"total_jobs_run": cs.totalJobsRun,
	}
}

// ScheduleCognitivePhase schedules a job to run at a specific cognitive phase
func (cs *CognitiveScheduler) ScheduleCognitivePhase(name string, phase int, cycleLength time.Duration, task func(ctx context.Context) error) (*CognitiveJob, error) {
	definition := CognitivePhaseJobDefinition{
		Phase:       phase,
		CycleLength: cycleLength,
	}
	return cs.NewJob(name, definition, task, WithTags("cognitive_phase", fmt.Sprintf("phase_%d", phase)))
}

// ScheduleWakeTask schedules a job to run during wake periods
func (cs *CognitiveScheduler) ScheduleWakeTask(name string, interval time.Duration, task func(ctx context.Context) error) (*CognitiveJob, error) {
	definition := WakeRestJobDefinition{
		WakeTime:  8 * time.Hour,  // 8 AM
		RestTime:  22 * time.Hour, // 10 PM
		RunDuring: "wake",
		Interval:  interval,
	}
	return cs.NewJob(name, definition, task, WithTags("wake_task"))
}

// ScheduleRestTask schedules a job to run during rest periods
func (cs *CognitiveScheduler) ScheduleRestTask(name string, interval time.Duration, task func(ctx context.Context) error) (*CognitiveJob, error) {
	definition := WakeRestJobDefinition{
		WakeTime:  8 * time.Hour,
		RestTime:  22 * time.Hour,
		RunDuring: "rest",
		Interval:  interval,
	}
	return cs.NewJob(name, definition, task, WithTags("rest_task"))
}

// ScheduleInterval schedules a job to run at fixed intervals
func (cs *CognitiveScheduler) ScheduleInterval(name string, interval time.Duration, task func(ctx context.Context) error) (*CognitiveJob, error) {
	definition := DurationJobDefinition{Duration: interval}
	return cs.NewJob(name, definition, task, WithTags("interval_task"))
}
