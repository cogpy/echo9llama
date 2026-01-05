package lampstack

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server provides an HTTP API for the LAMPSTACK using Echo framework
type Server struct {
	mu sync.RWMutex

	stack   *Stack
	config  *ServerConfig
	echo    *echo.Echo

	// State
	running   bool
	startTime time.Time
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Address         string        `json:"address"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	EnableCORS      bool          `json:"enable_cors"`
	EnableMetrics   bool          `json:"enable_metrics"`
	EnableLogging   bool          `json:"enable_logging"`
	EnableRecover   bool          `json:"enable_recover"`
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Address:       ":8080",
		ReadTimeout:   30 * time.Second,
		WriteTimeout:  60 * time.Second,
		EnableCORS:    true,
		EnableMetrics: true,
		EnableLogging: true,
		EnableRecover: true,
	}
}

// NewServer creates a new LAMPSTACK server using Echo framework
func NewServer(stack *Stack, config *ServerConfig) *Server {
	if config == nil {
		config = DefaultServerConfig()
	}

	return &Server{
		stack:  stack,
		config: config,
	}
}

// Start begins serving HTTP requests
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.mu.Unlock()

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Configure middleware
	s.configureMiddleware(e)

	// Register routes
	s.registerRoutes(e)

	s.echo = e

	s.mu.Lock()
	s.running = true
	s.startTime = time.Now()
	s.mu.Unlock()

	fmt.Printf("🔷 LAMPSTACK Echo Server starting on %s\n", s.config.Address)

	// Start server in goroutine
	go func() {
		if err := e.Start(s.config.Address); err != nil && err != http.ErrServerClosed {
			fmt.Printf("   ⚠️  Server error: %v\n", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	fmt.Println("🔷 LAMPSTACK Echo Server stopping...")

	if s.echo != nil {
		if err := s.echo.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
	}

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	fmt.Println("🔷 LAMPSTACK Echo Server stopped")
	return nil
}

// configureMiddleware sets up Echo middleware
func (s *Server) configureMiddleware(e *echo.Echo) {
	// Recovery middleware
	if s.config.EnableRecover {
		e.Use(middleware.Recover())
	}

	// Logging middleware
	if s.config.EnableLogging {
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "  ${time_rfc3339} ${method} ${uri} ${status} ${latency_human}\n",
		}))
	}

	// CORS middleware
	if s.config.EnableCORS {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		}))
	}

	// Request ID middleware
	e.Use(middleware.RequestID())

	// Timeout middleware
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: s.config.WriteTimeout,
	}))
}

// registerRoutes sets up HTTP routes using Echo
func (s *Server) registerRoutes(e *echo.Echo) {
	// Health and info
	e.GET("/health", s.handleHealth)
	e.GET("/info", s.handleInfo)
	e.GET("/metrics", s.handleMetrics)

	// API group
	api := e.Group("/api")

	// Generation API
	api.POST("/generate", s.handleGenerate)
	api.POST("/chat", s.handleChat)

	// Cognitive API (Echo endpoints)
	echoAPI := api.Group("/echo")
	echoAPI.POST("/think", s.handleThink)
	echoAPI.POST("/reflect", s.handleReflect)
	echoAPI.GET("/state", s.handleCognitiveState)

	// Memory API
	memAPI := api.Group("/memory")
	memAPI.POST("/store", s.handleMemoryStore)
	memAPI.POST("/retrieve", s.handleMemoryRetrieve)
	memAPI.GET("/stats", s.handleMemoryStats)

	// Persistence API
	api.GET("/snapshot", s.handleSnapshotGet)
	api.POST("/snapshot", s.handleSnapshotCreate)
}

// Handlers using Echo context

func (s *Server) handleHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"version": Version,
	})
}

func (s *Server) handleInfo(c echo.Context) error {
	s.mu.RLock()
	uptime := time.Since(s.startTime)
	s.mu.RUnlock()

	providers := []string{}
	if s.stack.api != nil {
		for _, p := range s.stack.api.GetProviders() {
			if p.Available {
				providers = append(providers, p.Name)
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"name":      StackName,
		"version":   Version,
		"uptime":    uptime.String(),
		"providers": providers,
		"features": map[string]bool{
			"echo":        s.stack.echo != nil,
			"memory":      s.stack.memory != nil,
			"persistence": s.stack.persistence != nil,
		},
	})
}

func (s *Server) handleMetrics(c echo.Context) error {
	metrics := s.stack.GetMetrics()
	return c.JSON(http.StatusOK, metrics)
}

func (s *Server) handleGenerate(c echo.Context) error {
	var request GenerationRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Invalid request body"))
	}

	response, err := s.stack.Generate(c.Request().Context(), &request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) handleChat(c echo.Context) error {
	var request struct {
		Messages []Message `json:"messages"`
	}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Invalid request body"))
	}

	response, err := s.stack.Chat(c.Request().Context(), request.Messages)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) handleThink(c echo.Context) error {
	var request struct {
		Prompt string `json:"prompt"`
	}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Invalid request body"))
	}

	response, err := s.stack.Think(c.Request().Context(), request.Prompt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) handleReflect(c echo.Context) error {
	var request struct {
		Context string `json:"context"`
	}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Invalid request body"))
	}

	response, err := s.stack.Reflect(c.Request().Context(), request.Context)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) handleCognitiveState(c echo.Context) error {
	state, err := s.stack.GetCognitiveState()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusOK, state)
}

func (s *Server) handleMemoryStore(c echo.Context) error {
	var entry MemoryEntry
	if err := c.Bind(&entry); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Invalid request body"))
	}

	if err := s.stack.StoreMemory(c.Request().Context(), &entry); err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "stored",
		"id":     entry.ID,
	})
}

func (s *Server) handleMemoryRetrieve(c echo.Context) error {
	var request struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Invalid request body"))
	}

	if request.Limit <= 0 {
		request.Limit = 10
	}

	memories, err := s.stack.RetrieveMemory(c.Request().Context(), request.Query, request.Limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"memories": memories,
		"count":    len(memories),
	})
}

func (s *Server) handleMemoryStats(c echo.Context) error {
	if s.stack.memory == nil {
		return c.JSON(http.StatusServiceUnavailable, errorResponse("Memory not available"))
	}

	stats := s.stack.memory.GetStats()
	return c.JSON(http.StatusOK, stats)
}

func (s *Server) handleSnapshotCreate(c echo.Context) error {
	persistence, ok := s.stack.persistence.(*JSONPersistence)
	if !ok {
		return c.JSON(http.StatusServiceUnavailable, errorResponse("Persistence not available"))
	}

	snapshot, err := persistence.CreateSnapshot(c.Request().Context(), s.stack)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	if err := persistence.SaveSnapshot(c.Request().Context(), snapshot); err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "snapshot_created",
		"timestamp": snapshot.Timestamp,
	})
}

func (s *Server) handleSnapshotGet(c echo.Context) error {
	persistence, ok := s.stack.persistence.(*JSONPersistence)
	if !ok {
		return c.JSON(http.StatusServiceUnavailable, errorResponse("Persistence not available"))
	}

	snapshot, err := persistence.LoadLatestSnapshot(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusNotFound, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusOK, snapshot)
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code,omitempty"`
}

func errorResponse(message string) ErrorResponse {
	return ErrorResponse{Error: message}
}

// QuickStart creates and starts a LAMPSTACK server with default configuration
func QuickStart(ctx context.Context, address string) (*Server, error) {
	// Create stack with default config
	stack, err := NewStack(DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create stack: %w", err)
	}

	// Initialize the stack
	if err := stack.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize stack: %w", err)
	}

	// Start the stack
	if err := stack.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start stack: %w", err)
	}

	// Create and start server
	serverConfig := DefaultServerConfig()
	if address != "" {
		serverConfig.Address = address
	}

	server := NewServer(stack, serverConfig)
	if err := server.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	return server, nil
}

// Echo returns the underlying Echo instance for advanced configuration
func (s *Server) Echo() *echo.Echo {
	return s.echo
}

// Stack returns the LAMPSTACK instance
func (s *Server) Stack() *Stack {
	return s.stack
}
