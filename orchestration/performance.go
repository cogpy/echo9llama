package orchestration

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// PerformanceOptimizer provides advanced performance optimization capabilities
type PerformanceOptimizer struct {
	resourceManager   *ResourceManager
	taskScheduler     *IntelligentScheduler
	loadBalancer      *LoadBalancer
	performanceMonitor *PerformanceMonitor
}

// ResourceManager handles dynamic resource allocation
type ResourceManager struct {
	mu                sync.RWMutex
	availableResources *ResourcePool
	resourceUsage     map[string]*ResourceUsage
	reservations      map[string]*ResourceReservation
}

// ResourcePool represents available computational resources
type ResourcePool struct {
	CPUCores        int
	MemoryGB        float64
	NetworkBandwidthMbps int
	StorageGB       float64
	GPUCount        int
	LastUpdated     time.Time
}

// ResourceUsage tracks resource consumption by agents/tasks
type ResourceUsage struct {
	AgentID       string
	CPUUsage      float64  // 0.0 to 1.0 per core
	MemoryUsageGB float64
	NetworkUsageMbps float64
	StorageUsageGB float64
	GPUUsage      float64  // 0.0 to 1.0 per GPU
	LastUpdated   time.Time
}

// ResourceReservation represents a resource allocation
type ResourceReservation struct {
	ReservationID string
	AgentID       string
	TaskID        string
	Resources     *ResourceUsage
	StartTime     time.Time
	EndTime       time.Time
	Priority      ResourcePriority
}

// ResourcePriority defines resource allocation priority
type ResourcePriority int

const (
	ResourcePriorityLow ResourcePriority = iota
	ResourcePriorityNormal
	ResourcePriorityHigh
	ResourcePriorityCritical
)

// IntelligentScheduler provides advanced task scheduling
type IntelligentScheduler struct {
	mu               sync.RWMutex
	schedulingQueue  []*ScheduledTask
	executionHistory []*ExecutionRecord
	schedulingPolicies map[string]SchedulingPolicy
}

// ScheduledTask represents a task in the scheduling queue
type ScheduledTask struct {
	Task            *Task
	Agent           *Agent
	Priority        TaskPriority
	Deadline        time.Time
	EstimatedDuration time.Duration
	ResourceRequirements *ResourceUsage
	Dependencies    []string // Task IDs this task depends on
	ScheduledAt     time.Time
	Status          SchedulingStatus
}

// TaskPriority defines task execution priority
type TaskPriority int

const (
	TaskPriorityLow TaskPriority = iota
	TaskPriorityNormal
	TaskPriorityHigh
	TaskPriorityUrgent
)

// SchedulingStatus represents the status of a scheduled task
type SchedulingStatus string

const (
	SchedulingStatusPending   SchedulingStatus = "pending"
	SchedulingStatusScheduled SchedulingStatus = "scheduled"
	SchedulingStatusRunning   SchedulingStatus = "running"
	SchedulingStatusCompleted SchedulingStatus = "completed"
	SchedulingStatusFailed    SchedulingStatus = "failed"
)

// ExecutionRecord tracks task execution history
type ExecutionRecord struct {
	TaskID           string
	AgentID          string
	ScheduledTime    time.Time
	StartTime        time.Time
	EndTime          time.Time
	ActualDuration   time.Duration
	EstimatedDuration time.Duration
	ResourcesUsed    *ResourceUsage
	Success          bool
	PerformanceMetrics *ExecutionMetrics
}

// ExecutionMetrics provides detailed execution performance data
type ExecutionMetrics struct {
	ThroughputTPS    float64 // Tasks per second
	ResponseTime     time.Duration
	ErrorRate        float64
	ResourceEfficiency map[string]float64 // Resource type -> efficiency (0-1)
	QualityScore     float64
}

// SchedulingPolicy defines how tasks should be scheduled
type SchedulingPolicy interface {
	Name() string
	ScheduleTasks(tasks []*ScheduledTask, resources *ResourcePool) []*ScheduledTask
	Priority() int
}

// LoadBalancer handles load distribution across agents
type LoadBalancer struct {
	mu              sync.RWMutex
	agentLoads      map[string]*AgentLoad
	balancingStrategy BalancingStrategy
	healthChecker   *HealthChecker
}

// AgentLoad tracks the current load on an agent
type AgentLoad struct {
	AgentID           string
	ActiveTasks       int
	QueuedTasks       int
	ResourceUtilization *ResourceUsage
	PerformanceScore  float64
	HealthStatus      HealthStatus
	LastUpdated       time.Time
}

// BalancingStrategy defines load balancing approach
type BalancingStrategy string

const (
	BalancingStrategyRoundRobin     BalancingStrategy = "round_robin"
	BalancingStrategyLeastLoaded    BalancingStrategy = "least_loaded"
	BalancingStrategyWeightedRound  BalancingStrategy = "weighted_round_robin"
	BalancingStrategyPerformanceBased BalancingStrategy = "performance_based"
	BalancingStrategyAdaptive       BalancingStrategy = "adaptive"
)

// HealthStatus represents agent health
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthChecker monitors agent health
type HealthChecker struct {
	mu             sync.RWMutex
	healthChecks   map[string]*HealthCheck
	checkInterval  time.Duration
	healthHistory  map[string][]*HealthRecord
}

// HealthCheck defines a health check configuration
type HealthCheck struct {
	AgentID         string
	CheckType       HealthCheckType
	Interval        time.Duration
	Timeout         time.Duration
	Threshold       float64
	Enabled         bool
}

// HealthCheckType defines different types of health checks
type HealthCheckType string

const (
	HealthCheckTypeResponse    HealthCheckType = "response_time"
	HealthCheckTypeResource    HealthCheckType = "resource_usage"
	HealthCheckTypeError       HealthCheckType = "error_rate"
	HealthCheckTypeCapacity    HealthCheckType = "capacity"
	HealthCheckTypePerformance HealthCheckType = "performance"
)

// HealthRecord tracks health check results over time
type HealthRecord struct {
	Timestamp   time.Time
	CheckType   HealthCheckType
	Value       float64
	Status      HealthStatus
	Details     map[string]interface{}
}

// PerformanceMonitor tracks system-wide performance metrics
type PerformanceMonitor struct {
	mu              sync.RWMutex
	systemMetrics   *SystemMetrics
	alertRules      []*AlertRule
	metricHistory   map[string][]*MetricSnapshot
	alertHistory    []*Alert
}

// SystemMetrics represents overall system performance
type SystemMetrics struct {
	TotalTasks           int
	CompletedTasks       int
	FailedTasks          int
	AverageResponseTime  time.Duration
	ThroughputTPS        float64
	ResourceUtilization  *ResourceUsage
	SystemHealth         float64 // 0.0 to 1.0
	LastUpdated          time.Time
}

// MetricSnapshot captures metrics at a point in time
type MetricSnapshot struct {
	Timestamp time.Time
	Metrics   map[string]float64
}

// AlertRule defines conditions for triggering alerts
type AlertRule struct {
	ID          string
	Name        string
	MetricName  string
	Condition   AlertCondition
	Threshold   float64
	Duration    time.Duration
	Severity    AlertSeverity
	Enabled     bool
}

// AlertCondition defines when an alert should trigger
type AlertCondition string

const (
	AlertConditionGreaterThan AlertCondition = "greater_than"
	AlertConditionLessThan    AlertCondition = "less_than"
	AlertConditionEquals      AlertCondition = "equals"
	AlertConditionNotEquals   AlertCondition = "not_equals"
)

// AlertSeverity defines alert importance
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// Alert represents a triggered alert
type Alert struct {
	ID          string
	RuleID      string
	RuleName    string
	Message     string
	Severity    AlertSeverity
	MetricValue float64
	Threshold   float64
	Timestamp   time.Time
	Resolved    bool
	ResolvedAt  *time.Time
}

// NewPerformanceOptimizer creates a new performance optimizer
func NewPerformanceOptimizer() *PerformanceOptimizer {
	return &PerformanceOptimizer{
		resourceManager:    NewResourceManager(),
		taskScheduler:      NewIntelligentScheduler(),
		loadBalancer:       NewLoadBalancer(),
		performanceMonitor: NewPerformanceMonitor(),
	}
}

// NewResourceManager creates a new resource manager
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		availableResources: &ResourcePool{
			CPUCores:             8,
			MemoryGB:             32.0,
			NetworkBandwidthMbps: 1000,
			StorageGB:            1000.0,
			GPUCount:             2,
			LastUpdated:          time.Now(),
		},
		resourceUsage: make(map[string]*ResourceUsage),
		reservations:  make(map[string]*ResourceReservation),
	}
}

// NewIntelligentScheduler creates a new intelligent scheduler
func NewIntelligentScheduler() *IntelligentScheduler {
	scheduler := &IntelligentScheduler{
		schedulingQueue:    make([]*ScheduledTask, 0),
		executionHistory:   make([]*ExecutionRecord, 0),
		schedulingPolicies: make(map[string]SchedulingPolicy),
	}
	
	// Register default scheduling policies
	scheduler.RegisterPolicy(&PrioritySchedulingPolicy{})
	scheduler.RegisterPolicy(&DeadlineAwareSchedulingPolicy{})
	scheduler.RegisterPolicy(&ResourceOptimizedSchedulingPolicy{})
	
	return scheduler
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		agentLoads:        make(map[string]*AgentLoad),
		balancingStrategy: BalancingStrategyAdaptive,
		healthChecker:     NewHealthChecker(),
	}
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		healthChecks:  make(map[string]*HealthCheck),
		checkInterval: 30 * time.Second,
		healthHistory: make(map[string][]*HealthRecord),
	}
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	monitor := &PerformanceMonitor{
		systemMetrics: &SystemMetrics{
			LastUpdated: time.Now(),
		},
		alertRules:    make([]*AlertRule, 0),
		metricHistory: make(map[string][]*MetricSnapshot),
		alertHistory:  make([]*Alert, 0),
	}
	
	// Register default alert rules
	monitor.RegisterDefaultAlerts()
	
	return monitor
}

// ResourceManager methods

// AllocateResources allocates resources for a task
func (rm *ResourceManager) AllocateResources(ctx context.Context, taskID, agentID string, requirements *ResourceUsage, priority ResourcePriority) (*ResourceReservation, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	// Check if resources are available
	if !rm.hasAvailableResources(requirements) {
		return nil, fmt.Errorf("insufficient resources available")
	}
	
	reservationID := fmt.Sprintf("res_%s_%d", taskID, time.Now().Unix())
	reservation := &ResourceReservation{
		ReservationID: reservationID,
		AgentID:       agentID,
		TaskID:        taskID,
		Resources:     requirements,
		StartTime:     time.Now(),
		Priority:      priority,
	}
	
	rm.reservations[reservationID] = reservation
	rm.updateResourceUsage(agentID, requirements)
	
	return reservation, nil
}

// ReleaseResources releases allocated resources
func (rm *ResourceManager) ReleaseResources(ctx context.Context, reservationID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	reservation, exists := rm.reservations[reservationID]
	if !exists {
		return fmt.Errorf("reservation not found: %s", reservationID)
	}
	
	// Update resource usage
	if usage, exists := rm.resourceUsage[reservation.AgentID]; exists {
		usage.CPUUsage -= reservation.Resources.CPUUsage
		usage.MemoryUsageGB -= reservation.Resources.MemoryUsageGB
		usage.NetworkUsageMbps -= reservation.Resources.NetworkUsageMbps
		usage.StorageUsageGB -= reservation.Resources.StorageUsageGB
		usage.GPUUsage -= reservation.Resources.GPUUsage
		usage.LastUpdated = time.Now()
	}
	
	delete(rm.reservations, reservationID)
	return nil
}

// hasAvailableResources checks if requested resources are available
func (rm *ResourceManager) hasAvailableResources(requirements *ResourceUsage) bool {
	totalCPUUsage := 0.0
	totalMemoryUsage := 0.0
	totalNetworkUsage := 0.0
	totalStorageUsage := 0.0
	totalGPUUsage := 0.0
	
	for _, usage := range rm.resourceUsage {
		totalCPUUsage += usage.CPUUsage
		totalMemoryUsage += usage.MemoryUsageGB
		totalNetworkUsage += usage.NetworkUsageMbps
		totalStorageUsage += usage.StorageUsageGB
		totalGPUUsage += usage.GPUUsage
	}
	
	return (totalCPUUsage+requirements.CPUUsage) <= float64(rm.availableResources.CPUCores) &&
		(totalMemoryUsage+requirements.MemoryUsageGB) <= rm.availableResources.MemoryGB &&
		(totalNetworkUsage+requirements.NetworkUsageMbps) <= float64(rm.availableResources.NetworkBandwidthMbps) &&
		(totalStorageUsage+requirements.StorageUsageGB) <= rm.availableResources.StorageGB &&
		(totalGPUUsage+requirements.GPUUsage) <= float64(rm.availableResources.GPUCount)
}

// updateResourceUsage updates resource usage for an agent
func (rm *ResourceManager) updateResourceUsage(agentID string, requirements *ResourceUsage) {
	if usage, exists := rm.resourceUsage[agentID]; exists {
		usage.CPUUsage += requirements.CPUUsage
		usage.MemoryUsageGB += requirements.MemoryUsageGB
		usage.NetworkUsageMbps += requirements.NetworkUsageMbps
		usage.StorageUsageGB += requirements.StorageUsageGB
		usage.GPUUsage += requirements.GPUUsage
		usage.LastUpdated = time.Now()
	} else {
		rm.resourceUsage[agentID] = &ResourceUsage{
			AgentID:            agentID,
			CPUUsage:           requirements.CPUUsage,
			MemoryUsageGB:      requirements.MemoryUsageGB,
			NetworkUsageMbps:   requirements.NetworkUsageMbps,
			StorageUsageGB:     requirements.StorageUsageGB,
			GPUUsage:           requirements.GPUUsage,
			LastUpdated:        time.Now(),
		}
	}
}

// IntelligentScheduler methods

// ScheduleTask adds a task to the scheduling queue
func (is *IntelligentScheduler) ScheduleTask(task *Task, agent *Agent, priority TaskPriority, deadline time.Time) (*ScheduledTask, error) {
	is.mu.Lock()
	defer is.mu.Unlock()
	
	scheduledTask := &ScheduledTask{
		Task:                 task,
		Agent:                agent,
		Priority:             priority,
		Deadline:             deadline,
		EstimatedDuration:    is.estimateTaskDuration(task),
		ResourceRequirements: is.estimateResourceRequirements(task),
		Dependencies:         make([]string, 0),
		ScheduledAt:          time.Now(),
		Status:               SchedulingStatusPending,
	}
	
	is.schedulingQueue = append(is.schedulingQueue, scheduledTask)
	is.optimizeSchedule()
	
	return scheduledTask, nil
}

// GetNextTask returns the next task to execute
func (is *IntelligentScheduler) GetNextTask(ctx context.Context, agentID string) (*ScheduledTask, error) {
	is.mu.Lock()
	defer is.mu.Unlock()
	
	for i, scheduledTask := range is.schedulingQueue {
		if scheduledTask.Agent.ID == agentID && scheduledTask.Status == SchedulingStatusScheduled {
			// Remove from queue
			is.schedulingQueue = append(is.schedulingQueue[:i], is.schedulingQueue[i+1:]...)
			scheduledTask.Status = SchedulingStatusRunning
			return scheduledTask, nil
		}
	}
	
	return nil, fmt.Errorf("no tasks available for agent %s", agentID)
}

// optimizeSchedule applies scheduling policies to optimize task execution
func (is *IntelligentScheduler) optimizeSchedule() {
	// Apply scheduling policies in priority order
	policies := make([]SchedulingPolicy, 0, len(is.schedulingPolicies))
	for _, policy := range is.schedulingPolicies {
		policies = append(policies, policy)
	}
	
	// Sort by priority
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].Priority() > policies[j].Priority()
	})
	
	// Apply each policy
	for _, policy := range policies {
		is.schedulingQueue = policy.ScheduleTasks(is.schedulingQueue, nil)
	}
	
	// Mark top tasks as scheduled
	for i, task := range is.schedulingQueue {
		if i < 10 && task.Status == SchedulingStatusPending { // Schedule top 10
			task.Status = SchedulingStatusScheduled
		}
	}
}

// estimateTaskDuration estimates how long a task will take
func (is *IntelligentScheduler) estimateTaskDuration(task *Task) time.Duration {
	// Base durations by task type
	baseDuration := map[string]time.Duration{
		TaskTypeGenerate: 10 * time.Second,
		TaskTypeChat:     15 * time.Second,
		TaskTypeEmbed:    3 * time.Second,
		TaskTypeTool:     5 * time.Second,
		TaskTypeReflect:  2 * time.Second,
		TaskTypePlugin:   500 * time.Millisecond,
	}
	
	duration, exists := baseDuration[task.Type]
	if !exists {
		duration = 5 * time.Second
	}
	
	// Adjust based on input length
	inputFactor := float64(len(task.Input)) / 1000.0
	if inputFactor < 1.0 {
		inputFactor = 1.0
	}
	
	return time.Duration(float64(duration) * inputFactor)
}

// estimateResourceRequirements estimates resource needs for a task
func (is *IntelligentScheduler) estimateResourceRequirements(task *Task) *ResourceUsage {
	baseRequirements := map[string]*ResourceUsage{
		TaskTypeGenerate: {CPUUsage: 0.5, MemoryUsageGB: 2.0, NetworkUsageMbps: 10},
		TaskTypeChat:     {CPUUsage: 0.4, MemoryUsageGB: 1.5, NetworkUsageMbps: 15},
		TaskTypeEmbed:    {CPUUsage: 0.8, MemoryUsageGB: 4.0, NetworkUsageMbps: 5, GPUUsage: 0.3},
		TaskTypeTool:     {CPUUsage: 0.3, MemoryUsageGB: 1.0, NetworkUsageMbps: 20},
		TaskTypeReflect:  {CPUUsage: 0.2, MemoryUsageGB: 0.5, NetworkUsageMbps: 5},
		TaskTypePlugin:   {CPUUsage: 0.6, MemoryUsageGB: 3.0, NetworkUsageMbps: 8},
	}
	
	if requirements, exists := baseRequirements[task.Type]; exists {
		return requirements
	}
	
	return &ResourceUsage{
		CPUUsage:         0.4,
		MemoryUsageGB:    1.5,
		NetworkUsageMbps: 10,
	}
}

// RegisterPolicy registers a scheduling policy
func (is *IntelligentScheduler) RegisterPolicy(policy SchedulingPolicy) {
	is.mu.Lock()
	defer is.mu.Unlock()
	is.schedulingPolicies[policy.Name()] = policy
}

// Priority-based scheduling policy
type PrioritySchedulingPolicy struct{}

func (psp *PrioritySchedulingPolicy) Name() string { return "priority" }
func (psp *PrioritySchedulingPolicy) Priority() int { return 100 }

func (psp *PrioritySchedulingPolicy) ScheduleTasks(tasks []*ScheduledTask, resources *ResourcePool) []*ScheduledTask {
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Priority > tasks[j].Priority
	})
	return tasks
}

// Deadline-aware scheduling policy
type DeadlineAwareSchedulingPolicy struct{}

func (dasp *DeadlineAwareSchedulingPolicy) Name() string { return "deadline_aware" }
func (dasp *DeadlineAwareSchedulingPolicy) Priority() int { return 90 }

func (dasp *DeadlineAwareSchedulingPolicy) ScheduleTasks(tasks []*ScheduledTask, resources *ResourcePool) []*ScheduledTask {
	now := time.Now()
	sort.Slice(tasks, func(i, j int) bool {
		// Prioritize tasks closer to deadline
		deadlineUrgencyI := tasks[i].Deadline.Sub(now)
		deadlineUrgencyJ := tasks[j].Deadline.Sub(now)
		return deadlineUrgencyI < deadlineUrgencyJ
	})
	return tasks
}

// Resource-optimized scheduling policy
type ResourceOptimizedSchedulingPolicy struct{}

func (rosp *ResourceOptimizedSchedulingPolicy) Name() string { return "resource_optimized" }
func (rosp *ResourceOptimizedSchedulingPolicy) Priority() int { return 80 }

func (rosp *ResourceOptimizedSchedulingPolicy) ScheduleTasks(tasks []*ScheduledTask, resources *ResourcePool) []*ScheduledTask {
	sort.Slice(tasks, func(i, j int) bool {
		// Prioritize tasks with lower resource requirements when resources are scarce
		resourceScoreI := tasks[i].ResourceRequirements.CPUUsage + tasks[i].ResourceRequirements.MemoryUsageGB/10
		resourceScoreJ := tasks[j].ResourceRequirements.CPUUsage + tasks[j].ResourceRequirements.MemoryUsageGB/10
		return resourceScoreI < resourceScoreJ
	})
	return tasks
}

// LoadBalancer methods

// SelectOptimalAgent selects the best agent for a task based on current loads
func (lb *LoadBalancer) SelectOptimalAgent(ctx context.Context, task *Task, availableAgents []*Agent) (*Agent, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	if len(availableAgents) == 0 {
		return nil, fmt.Errorf("no available agents")
	}
	
	switch lb.balancingStrategy {
	case BalancingStrategyLeastLoaded:
		return lb.selectLeastLoadedAgent(availableAgents), nil
	case BalancingStrategyPerformanceBased:
		return lb.selectBestPerformingAgent(availableAgents), nil
	case BalancingStrategyAdaptive:
		return lb.selectAdaptiveAgent(availableAgents, task), nil
	default:
		return lb.selectRoundRobinAgent(availableAgents), nil
	}
}

// selectLeastLoadedAgent returns the agent with the lowest current load
func (lb *LoadBalancer) selectLeastLoadedAgent(agents []*Agent) *Agent {
	bestAgent := agents[0]
	bestLoad := lb.calculateAgentLoad(bestAgent)
	
	for _, agent := range agents[1:] {
		load := lb.calculateAgentLoad(agent)
		if load < bestLoad {
			bestLoad = load
			bestAgent = agent
		}
	}
	
	return bestAgent
}

// selectBestPerformingAgent returns the agent with the best performance score
func (lb *LoadBalancer) selectBestPerformingAgent(agents []*Agent) *Agent {
	bestAgent := agents[0]
	bestScore := lb.getPerformanceScore(bestAgent)
	
	for _, agent := range agents[1:] {
		score := lb.getPerformanceScore(agent)
		if score > bestScore {
			bestScore = score
			bestAgent = agent
		}
	}
	
	return bestAgent
}

// selectAdaptiveAgent uses adaptive logic combining load, performance, and health
func (lb *LoadBalancer) selectAdaptiveAgent(agents []*Agent, task *Task) *Agent {
	bestAgent := agents[0]
	bestScore := lb.calculateAdaptiveScore(bestAgent, task)
	
	for _, agent := range agents[1:] {
		score := lb.calculateAdaptiveScore(agent, task)
		if score > bestScore {
			bestScore = score
			bestAgent = agent
		}
	}
	
	return bestAgent
}

// selectRoundRobinAgent implements simple round-robin selection
func (lb *LoadBalancer) selectRoundRobinAgent(agents []*Agent) *Agent {
	// Simple round-robin implementation (in practice, would maintain state)
	return agents[time.Now().Unix()%int64(len(agents))]
}

// calculateAgentLoad calculates current load for an agent
func (lb *LoadBalancer) calculateAgentLoad(agent *Agent) float64 {
	if load, exists := lb.agentLoads[agent.ID]; exists {
		return float64(load.ActiveTasks+load.QueuedTasks) + load.ResourceUtilization.CPUUsage
	}
	return 0.0
}

// getPerformanceScore gets performance score for an agent
func (lb *LoadBalancer) getPerformanceScore(agent *Agent) float64 {
	if load, exists := lb.agentLoads[agent.ID]; exists {
		return load.PerformanceScore
	}
	return 0.5 // Default score
}

// calculateAdaptiveScore calculates adaptive score combining multiple factors
func (lb *LoadBalancer) calculateAdaptiveScore(agent *Agent, task *Task) float64 {
	load := lb.calculateAgentLoad(agent)
	performance := lb.getPerformanceScore(agent)
	health := lb.getHealthScore(agent)
	
	// Weighted combination: lower load is better, higher performance and health are better
	loadScore := 1.0 - (load / 10.0) // Normalize load
	return (loadScore * 0.3) + (performance * 0.4) + (health * 0.3)
}

// getHealthScore gets health score for an agent
func (lb *LoadBalancer) getHealthScore(agent *Agent) float64 {
	if load, exists := lb.agentLoads[agent.ID]; exists {
		switch load.HealthStatus {
		case HealthStatusHealthy:
			return 1.0
		case HealthStatusDegraded:
			return 0.7
		case HealthStatusUnhealthy:
			return 0.3
		default:
			return 0.5
		}
	}
	return 0.5 // Default health score
}

// UpdateAgentLoad updates the load information for an agent
func (lb *LoadBalancer) UpdateAgentLoad(agentID string, activeTasks, queuedTasks int, resourceUsage *ResourceUsage, performanceScore float64, healthStatus HealthStatus) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	lb.agentLoads[agentID] = &AgentLoad{
		AgentID:             agentID,
		ActiveTasks:         activeTasks,
		QueuedTasks:         queuedTasks,
		ResourceUtilization: resourceUsage,
		PerformanceScore:    performanceScore,
		HealthStatus:        healthStatus,
		LastUpdated:         time.Now(),
	}
}

// PerformanceMonitor methods

// RegisterDefaultAlerts registers common alert rules
func (pm *PerformanceMonitor) RegisterDefaultAlerts() {
	pm.alertRules = []*AlertRule{
		{
			ID:         "high_error_rate",
			Name:       "High Error Rate",
			MetricName: "error_rate",
			Condition:  AlertConditionGreaterThan,
			Threshold:  0.1, // 10% error rate
			Duration:   5 * time.Minute,
			Severity:   AlertSeverityWarning,
			Enabled:    true,
		},
		{
			ID:         "high_response_time",
			Name:       "High Response Time",
			MetricName: "avg_response_time",
			Condition:  AlertConditionGreaterThan,
			Threshold:  30000, // 30 seconds in milliseconds
			Duration:   3 * time.Minute,
			Severity:   AlertSeverityWarning,
			Enabled:    true,
		},
		{
			ID:         "low_throughput",
			Name:       "Low Throughput",
			MetricName: "throughput_tps",
			Condition:  AlertConditionLessThan,
			Threshold:  0.1, // Less than 0.1 tasks per second
			Duration:   10 * time.Minute,
			Severity:   AlertSeverityError,
			Enabled:    true,
		},
		{
			ID:         "system_health_critical",
			Name:       "Critical System Health",
			MetricName: "system_health",
			Condition:  AlertConditionLessThan,
			Threshold:  0.3, // Less than 30% system health
			Duration:   1 * time.Minute,
			Severity:   AlertSeverityCritical,
			Enabled:    true,
		},
	}
}

// UpdateSystemMetrics updates system-wide performance metrics
func (pm *PerformanceMonitor) UpdateSystemMetrics(metrics *SystemMetrics) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.systemMetrics = metrics
	pm.systemMetrics.LastUpdated = time.Now()
	
	// Check for alert conditions
	pm.checkAlerts()
	
	// Store metrics snapshot
	pm.storeMetricSnapshot()
}

// checkAlerts evaluates alert rules against current metrics
func (pm *PerformanceMonitor) checkAlerts() {
	now := time.Now()
	
	for _, rule := range pm.alertRules {
		if !rule.Enabled {
			continue
		}
		
		value := pm.getMetricValue(rule.MetricName)
		shouldAlert := false
		
		switch rule.Condition {
		case AlertConditionGreaterThan:
			shouldAlert = value > rule.Threshold
		case AlertConditionLessThan:
			shouldAlert = value < rule.Threshold
		case AlertConditionEquals:
			shouldAlert = value == rule.Threshold
		case AlertConditionNotEquals:
			shouldAlert = value != rule.Threshold
		}
		
		if shouldAlert {
			alert := &Alert{
				ID:          fmt.Sprintf("alert_%d", now.Unix()),
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				Message:     fmt.Sprintf("%s: %s is %.2f (threshold: %.2f)", rule.Name, rule.MetricName, value, rule.Threshold),
				Severity:    rule.Severity,
				MetricValue: value,
				Threshold:   rule.Threshold,
				Timestamp:   now,
				Resolved:    false,
			}
			
			pm.alertHistory = append(pm.alertHistory, alert)
		}
	}
}

// getMetricValue gets a specific metric value
func (pm *PerformanceMonitor) getMetricValue(metricName string) float64 {
	switch metricName {
	case "error_rate":
		if pm.systemMetrics.CompletedTasks+pm.systemMetrics.FailedTasks > 0 {
			return float64(pm.systemMetrics.FailedTasks) / float64(pm.systemMetrics.CompletedTasks+pm.systemMetrics.FailedTasks)
		}
		return 0.0
	case "avg_response_time":
		return float64(pm.systemMetrics.AverageResponseTime.Milliseconds())
	case "throughput_tps":
		return pm.systemMetrics.ThroughputTPS
	case "system_health":
		return pm.systemMetrics.SystemHealth
	default:
		return 0.0
	}
}

// storeMetricSnapshot stores current metrics for historical analysis
func (pm *PerformanceMonitor) storeMetricSnapshot() {
	snapshot := &MetricSnapshot{
		Timestamp: time.Now(),
		Metrics: map[string]float64{
			"total_tasks":         float64(pm.systemMetrics.TotalTasks),
			"completed_tasks":     float64(pm.systemMetrics.CompletedTasks),
			"failed_tasks":        float64(pm.systemMetrics.FailedTasks),
			"avg_response_time":   float64(pm.systemMetrics.AverageResponseTime.Milliseconds()),
			"throughput_tps":      pm.systemMetrics.ThroughputTPS,
			"system_health":       pm.systemMetrics.SystemHealth,
			"cpu_utilization":     pm.systemMetrics.ResourceUtilization.CPUUsage,
			"memory_utilization":  pm.systemMetrics.ResourceUtilization.MemoryUsageGB,
			"network_utilization": pm.systemMetrics.ResourceUtilization.NetworkUsageMbps,
		},
	}
	
	// Store in history (keep last 1000 snapshots)
	key := "system_metrics"
	if pm.metricHistory[key] == nil {
		pm.metricHistory[key] = make([]*MetricSnapshot, 0)
	}
	
	pm.metricHistory[key] = append(pm.metricHistory[key], snapshot)
	if len(pm.metricHistory[key]) > 1000 {
		pm.metricHistory[key] = pm.metricHistory[key][1:]
	}
}

// GetActiveAlerts returns currently active (unresolved) alerts
func (pm *PerformanceMonitor) GetActiveAlerts() []*Alert {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	activeAlerts := make([]*Alert, 0)
	for _, alert := range pm.alertHistory {
		if !alert.Resolved {
			activeAlerts = append(activeAlerts, alert)
		}
	}
	
	return activeAlerts
}

// GetSystemMetrics returns current system metrics
func (pm *PerformanceMonitor) GetSystemMetrics() *SystemMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.systemMetrics
}