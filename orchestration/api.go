package orchestration

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIServer provides REST endpoints for the Deep Tree Echo system
type APIServer struct {
	engine *Engine
	router *gin.Engine
}

// NewAPIServer creates a new API server for the orchestration engine
func NewAPIServer(engine *Engine) *APIServer {
	router := gin.Default()
	
	server := &APIServer{
		engine: engine,
		router: router,
	}
	
	server.setupRoutes()
	return server
}

// setupRoutes configures all API routes
func (s *APIServer) setupRoutes() {
	// Serve static dashboard
	s.router.StaticFile("/dashboard.html", "./examples/dashboard.html")
	s.router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/dashboard.html")
	})
	
	// Deep Tree Echo routes
	dte := s.router.Group("/api/deep-tree-echo")
	{
		dte.GET("/status", s.getDTEStatus)
		dte.GET("/dashboard", s.getDTEDashboard)
		dte.POST("/initialize", s.initializeDTE)
		dte.POST("/diagnostics", s.runDTEDiagnostics)
		dte.POST("/refresh", s.refreshDTEStatus)
		dte.POST("/introspection", s.performDTEIntrospection)
	}
	
	// Agent management routes
	agents := s.router.Group("/api/agents")
	{
		agents.GET("/", s.listAgents)
		agents.POST("/", s.createAgent)
		agents.GET("/:id", s.getAgent)
		agents.PUT("/:id", s.updateAgent)
		agents.DELETE("/:id", s.deleteAgent)
		agents.POST("/:id/tasks", s.executeTask)
	}
	
	// Orchestration routes
	orchestration := s.router.Group("/api/orchestration")
	{
		orchestration.POST("/", s.orchestrateTasks)
		orchestration.GET("/tools", s.getAvailableTools)
		orchestration.GET("/plugins", s.getAvailablePlugins)
	}
	
	// Learning System routes
	learning := s.router.Group("/api/learning")
	{
		learning.GET("/agents/:id/model", s.getLearningModel)
		learning.GET("/agents/:id/performance", s.getAgentPerformance)
		learning.POST("/agents/:id/adapt", s.adaptAgent)
		learning.POST("/predict-optimal-agent", s.predictOptimalAgent)
		learning.GET("/system/metrics", s.getLearningSystemMetrics)
	}
	
	// Performance Optimization routes
	performance := s.router.Group("/api/performance")
	{
		performance.GET("/metrics", s.getSystemMetrics)
		performance.GET("/alerts", s.getActiveAlerts)
		performance.GET("/resources", s.getResourceUsage)
		performance.GET("/agents/loads", s.getAgentLoads)
		performance.POST("/tasks/execute-optimized", s.executeTaskOptimized)
	}
}

// Run starts the API server on the specified port
func (s *APIServer) Run(port int) error {
	return s.router.Run(fmt.Sprintf(":%d", port))
}

// Deep Tree Echo API Handlers

func (s *APIServer) getDTEStatus(c *gin.Context) {
	status := s.engine.GetDeepTreeEchoStatus()
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   status,
	})
}

func (s *APIServer) getDTEDashboard(c *gin.Context) {
	dashboardData := s.engine.GetDeepTreeEchoDashboardData()
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   dashboardData,
	})
}

func (s *APIServer) initializeDTE(c *gin.Context) {
	err := s.engine.InitializeDeepTreeEcho(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Deep Tree Echo system initialized successfully",
	})
}

func (s *APIServer) runDTEDiagnostics(c *gin.Context) {
	diagnostics, err := s.engine.RunDeepTreeEchoDiagnostics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   diagnostics,
	})
}

func (s *APIServer) refreshDTEStatus(c *gin.Context) {
	err := s.engine.RefreshDeepTreeEchoStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Deep Tree Echo status refreshed successfully",
	})
}

func (s *APIServer) performDTEIntrospection(c *gin.Context) {
	var req struct {
		RepositoryRoot string  `json:"repository_root"`
		CurrentLoad    float64 `json:"current_load"`
		RecentActivity float64 `json:"recent_activity"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request body",
		})
		return
	}
	
	result, err := s.engine.PerformDeepTreeEchoIntrospection(
		c.Request.Context(),
		req.RepositoryRoot,
		req.CurrentLoad,
		req.RecentActivity,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})
}

// Agent Management API Handlers

func (s *APIServer) listAgents(c *gin.Context) {
	agents, err := s.engine.ListAgents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   agents,
	})
}

func (s *APIServer) createAgent(c *gin.Context) {
	var agent Agent
	if err := c.ShouldBindJSON(&agent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request body",
		})
		return
	}
	
	err := s.engine.CreateAgent(c.Request.Context(), &agent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   agent,
	})
}

func (s *APIServer) getAgent(c *gin.Context) {
	id := c.Param("id")
	
	agent, err := s.engine.GetAgent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   agent,
	})
}

func (s *APIServer) updateAgent(c *gin.Context) {
	id := c.Param("id")
	
	var agent Agent
	if err := c.ShouldBindJSON(&agent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request body",
		})
		return
	}
	
	agent.ID = id
	err := s.engine.UpdateAgent(c.Request.Context(), &agent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   agent,
	})
}

func (s *APIServer) deleteAgent(c *gin.Context) {
	id := c.Param("id")
	
	err := s.engine.DeleteAgent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Agent deleted successfully",
	})
}

func (s *APIServer) executeTask(c *gin.Context) {
	agentID := c.Param("id")
	
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request body",
		})
		return
	}
	
	task.AgentID = agentID
	
	agent, err := s.engine.GetAgent(c.Request.Context(), agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "Agent not found",
		})
		return
	}
	
	result, err := s.engine.ExecuteTask(c.Request.Context(), &task, agent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"task":   task,
			"result": result,
		},
	})
}

// Orchestration API Handlers

func (s *APIServer) orchestrateTasks(c *gin.Context) {
	var req OrchestrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request body",
		})
		return
	}
	
	response, err := s.engine.OrchestrateTasks(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

func (s *APIServer) getAvailableTools(c *gin.Context) {
	tools := s.engine.GetAvailableTools()
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   tools,
	})
}

func (s *APIServer) getAvailablePlugins(c *gin.Context) {
	plugins := s.engine.GetAvailablePlugins()
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   plugins,
	})
}

// Helper functions for common response patterns

func (s *APIServer) sendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"status": "error",
		"error":  message,
	})
}

func (s *APIServer) sendSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   data,
	})
}

// Dashboard Data Formatters

// FormatDashboardMetrics formats system metrics for dashboard display
func FormatDashboardMetrics(dte *DeepTreeEcho) map[string]interface{} {
	return map[string]interface{}{
		"systemHealth": map[string]interface{}{
			"status": dte.SystemHealth,
			"color":  getHealthColor(dte.SystemHealth),
		},
		"dteCore": map[string]interface{}{
			"status": dte.CoreStatus,
			"color":  getCoreStatusColor(dte.CoreStatus),
		},
		"thoughtCount": dte.ThoughtCount,
		"recursiveDepth": dte.RecursiveDepth,
	}
}

// FormatIdentityCoherence formats identity coherence data for dashboard
func FormatIdentityCoherence(coherence *IdentityCoherence) map[string]interface{} {
	return map[string]interface{}{
		"overallCoherence": fmt.Sprintf("%.0f%%", coherence.OverallCoherence*100),
		"maintainingCore": "Maintaining core essence while adapting",
		"factors": coherence.Factors,
	}
}

// FormatMemoryResonance formats memory resonance data for dashboard
func FormatMemoryResonance(resonance *MemoryResonance) map[string]interface{} {
	return map[string]interface{}{
		"memoryNodes":      resonance.MemoryNodes,
		"connections":      resonance.Connections,
		"coherence":        fmt.Sprintf("%.0f%%", resonance.Coherence*100),
		"activePatterns":   resonance.ActivePatterns,
		"resonancePattern": resonance.ResonancePattern,
	}
}

// FormatEchoPatterns formats echo patterns data for dashboard
func FormatEchoPatterns(patterns *EchoPatterns) map[string]interface{} {
	return map[string]interface{}{
		"recursiveSelfImprovement": map[string]interface{}{
			"name":        patterns.RecursiveSelfImprovement.Name,
			"description": patterns.RecursiveSelfImprovement.Description,
			"strength":    fmt.Sprintf("%.0f%%", patterns.RecursiveSelfImprovement.Strength*100),
			"frequency":   patterns.RecursiveSelfImprovement.Frequency,
		},
		"crossSystemSynthesis": map[string]interface{}{
			"name":        patterns.CrossSystemSynthesis.Name,
			"description": patterns.CrossSystemSynthesis.Description,
			"strength":    fmt.Sprintf("%.0f%%", patterns.CrossSystemSynthesis.Strength*100),
			"frequency":   patterns.CrossSystemSynthesis.Frequency,
		},
		"identityPreservation": map[string]interface{}{
			"name":        patterns.IdentityPreservation.Name,
			"description": patterns.IdentityPreservation.Description,
			"strength":    fmt.Sprintf("%.0f%%", patterns.IdentityPreservation.Strength*100),
			"frequency":   patterns.IdentityPreservation.Frequency,
		},
	}
}

// Helper functions for status colors

func getHealthColor(health SystemHealthStatus) string {
	switch health {
	case SystemHealthOptimal:
		return "green"
	case SystemHealthStable:
		return "blue"
	case SystemHealthDegraded:
		return "orange"
	case SystemHealthInactive:
		return "red"
	default:
		return "gray"
	}
}

func getCoreStatusColor(status CoreStatus) string {
	switch status {
	case CoreStatusActive:
		return "green"
	case CoreStatusStarting:
		return "yellow"
	case CoreStatusInactive:
		return "orange"
	case CoreStatusError:
		return "red"
	default:
		return "gray"
	}
}

// Learning System API Handlers

func (s *APIServer) getLearningModel(c *gin.Context) {
	agentID := c.Param("id")
	
	learningSystem := s.engine.GetLearningSystem()
	model := learningSystem.GetLearningModel(agentID)
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   model,
	})
}

func (s *APIServer) getAgentPerformance(c *gin.Context) {
	agentID := c.Param("id")
	
	learningSystem := s.engine.GetLearningSystem()
	history := learningSystem.performanceHistory[agentID]
	
	if history == nil {
		history = make([]*TaskPerformance, 0)
	}
	
	// Return recent performance (last 20 records)
	recentHistory := history
	if len(history) > 20 {
		recentHistory = history[len(history)-20:]
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": map[string]interface{}{
			"agent_id":           agentID,
			"total_tasks":        len(history),
			"recent_performance": recentHistory,
		},
	})
}

func (s *APIServer) adaptAgent(c *gin.Context) {
	agentID := c.Param("id")
	
	result, err := s.engine.AdaptAgent(c.Request.Context(), agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})
}

func (s *APIServer) predictOptimalAgent(c *gin.Context) {
	var req struct {
		TaskType   string                 `json:"task_type"`
		Input      string                 `json:"input"`
		Parameters map[string]interface{} `json:"parameters"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request body",
		})
		return
	}
	
	task := &Task{
		Type:       req.TaskType,
		Input:      req.Input,
		Parameters: req.Parameters,
	}
	
	agent, confidence, err := s.engine.PredictOptimalAgentForTask(c.Request.Context(), task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": map[string]interface{}{
			"optimal_agent": agent,
			"confidence":    confidence,
		},
	})
}

func (s *APIServer) getLearningSystemMetrics(c *gin.Context) {
	learningSystem := s.engine.GetLearningSystem()
	
	// Calculate system-wide learning metrics
	totalAgents := len(learningSystem.learningModels)
	totalPerformanceRecords := 0
	avgLearningRate := 0.0
	avgCurrentPerformance := 0.0
	
	for _, model := range learningSystem.learningModels {
		if history, exists := learningSystem.performanceHistory[model.AgentID]; exists {
			totalPerformanceRecords += len(history)
		}
		avgLearningRate += model.LearningRate
		avgCurrentPerformance += model.LearningTrajectory.CurrentPerformance
	}
	
	if totalAgents > 0 {
		avgLearningRate /= float64(totalAgents)
		avgCurrentPerformance /= float64(totalAgents)
	}
	
	// Get adaptation strategies count
	adaptationStrategiesCount := len(learningSystem.adaptationEngine.adaptationStrategies)
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": map[string]interface{}{
			"total_agents":                totalAgents,
			"total_performance_records":   totalPerformanceRecords,
			"average_learning_rate":       avgLearningRate,
			"average_current_performance": avgCurrentPerformance,
			"adaptation_strategies_count": adaptationStrategiesCount,
			"environment_factors":         learningSystem.adaptationEngine.environmentFactors,
		},
	})
}

// Performance Optimization API Handlers

func (s *APIServer) getSystemMetrics(c *gin.Context) {
	metrics := s.engine.GetSystemMetrics()
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   metrics,
	})
}

func (s *APIServer) getActiveAlerts(c *gin.Context) {
	alerts := s.engine.GetActiveAlerts()
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   alerts,
	})
}

func (s *APIServer) getResourceUsage(c *gin.Context) {
	usage := s.engine.GetResourceUsage()
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   usage,
	})
}

func (s *APIServer) getAgentLoads(c *gin.Context) {
	loads := s.engine.GetAgentLoads()
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   loads,
	})
}

func (s *APIServer) executeTaskOptimized(c *gin.Context) {
	var req struct {
		Type       string                 `json:"type"`
		Input      string                 `json:"input"`
		ModelName  string                 `json:"model_name"`
		Parameters map[string]interface{} `json:"parameters"`
		Priority   string                 `json:"priority"`   // "low", "normal", "high", "urgent"
		Deadline   string                 `json:"deadline"`   // ISO 8601 timestamp
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request body",
		})
		return
	}
	
	// Parse priority
	var priority TaskPriority
	switch req.Priority {
	case "low":
		priority = TaskPriorityLow
	case "normal":
		priority = TaskPriorityNormal
	case "high":
		priority = TaskPriorityHigh
	case "urgent":
		priority = TaskPriorityUrgent
	default:
		priority = TaskPriorityNormal
	}
	
	// Parse deadline
	deadline := time.Now().Add(30 * time.Minute) // Default deadline
	if req.Deadline != "" {
		if parsedDeadline, err := time.Parse(time.RFC3339, req.Deadline); err == nil {
			deadline = parsedDeadline
		}
	}
	
	// Create task
	task := &Task{
		ID:         fmt.Sprintf("opt-task-%d", time.Now().Unix()),
		Type:       req.Type,
		Input:      req.Input,
		ModelName:  req.ModelName,
		Parameters: req.Parameters,
		Status:     TaskStatusPending,
		CreatedAt:  time.Now(),
	}
	
	// Execute task with optimization
	result, err := s.engine.ExecuteTaskOptimized(c.Request.Context(), task, priority, deadline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})
}