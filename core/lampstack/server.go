package lampstack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Server provides an HTTP API for the LAMPSTACK
type Server struct {
	mu sync.RWMutex

	stack   *Stack
	address string
	server  *http.Server

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
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Address:      ":8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		EnableCORS:   true,
		EnableMetrics: true,
	}
}

// NewServer creates a new LAMPSTACK server
func NewServer(stack *Stack, config *ServerConfig) *Server {
	if config == nil {
		config = DefaultServerConfig()
	}

	return &Server{
		stack:   stack,
		address: config.Address,
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

	// Create router
	mux := http.NewServeMux()

	// Register routes
	s.registerRoutes(mux)

	// Create server
	s.server = &http.Server{
		Addr:         s.address,
		Handler:      s.corsMiddleware(mux),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	s.mu.Lock()
	s.running = true
	s.startTime = time.Now()
	s.mu.Unlock()

	fmt.Printf("🔷 LAMPSTACK Server starting on %s\n", s.address)

	// Start server
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

	fmt.Println("🔷 LAMPSTACK Server stopping...")

	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
	}

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	fmt.Println("🔷 LAMPSTACK Server stopped")
	return nil
}

// registerRoutes sets up HTTP routes
func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Health and info
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/info", s.handleInfo)
	mux.HandleFunc("/metrics", s.handleMetrics)

	// Generation API
	mux.HandleFunc("/api/generate", s.handleGenerate)
	mux.HandleFunc("/api/chat", s.handleChat)

	// Cognitive API
	mux.HandleFunc("/api/echo/think", s.handleThink)
	mux.HandleFunc("/api/echo/reflect", s.handleReflect)
	mux.HandleFunc("/api/echo/state", s.handleCognitiveState)

	// Memory API
	mux.HandleFunc("/api/memory/store", s.handleMemoryStore)
	mux.HandleFunc("/api/memory/retrieve", s.handleMemoryRetrieve)
	mux.HandleFunc("/api/memory/stats", s.handleMemoryStats)

	// Persistence API
	mux.HandleFunc("/api/snapshot", s.handleSnapshot)
}

// CORS middleware
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Handlers

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, map[string]interface{}{
		"status":  "ok",
		"version": Version,
	})
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
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

	s.jsonResponse(w, map[string]interface{}{
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

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := s.stack.GetMetrics()
	s.jsonResponse(w, metrics)
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "POST required")
		return
	}

	var request GenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := s.stack.Generate(r.Context(), &request)
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.jsonResponse(w, response)
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "POST required")
		return
	}

	var request struct {
		Messages []Message `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := s.stack.Chat(r.Context(), request.Messages)
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.jsonResponse(w, response)
}

func (s *Server) handleThink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "POST required")
		return
	}

	var request struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := s.stack.Think(r.Context(), request.Prompt)
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.jsonResponse(w, response)
}

func (s *Server) handleReflect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "POST required")
		return
	}

	var request struct {
		Context string `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := s.stack.Reflect(r.Context(), request.Context)
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.jsonResponse(w, response)
}

func (s *Server) handleCognitiveState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "GET required")
		return
	}

	state, err := s.stack.GetCognitiveState()
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.jsonResponse(w, state)
}

func (s *Server) handleMemoryStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "POST required")
		return
	}

	var entry MemoryEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := s.stack.StoreMemory(r.Context(), &entry); err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.jsonResponse(w, map[string]interface{}{
		"status": "stored",
		"id":     entry.ID,
	})
}

func (s *Server) handleMemoryRetrieve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "POST required")
		return
	}

	var request struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if request.Limit <= 0 {
		request.Limit = 10
	}

	memories, err := s.stack.RetrieveMemory(r.Context(), request.Query, request.Limit)
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.jsonResponse(w, map[string]interface{}{
		"memories": memories,
		"count":    len(memories),
	})
}

func (s *Server) handleMemoryStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "GET required")
		return
	}

	if s.stack.memory == nil {
		s.errorResponse(w, http.StatusServiceUnavailable, "Memory not available")
		return
	}

	stats := s.stack.memory.GetStats()
	s.jsonResponse(w, stats)
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Create snapshot
		persistence, ok := s.stack.persistence.(*JSONPersistence)
		if !ok {
			s.errorResponse(w, http.StatusServiceUnavailable, "Persistence not available")
			return
		}

		snapshot, err := persistence.CreateSnapshot(r.Context(), s.stack)
		if err != nil {
			s.errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if err := persistence.SaveSnapshot(r.Context(), snapshot); err != nil {
			s.errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		s.jsonResponse(w, map[string]interface{}{
			"status":    "snapshot_created",
			"timestamp": snapshot.Timestamp,
		})
		return
	}

	if r.Method == http.MethodGet {
		// Get latest snapshot
		persistence, ok := s.stack.persistence.(*JSONPersistence)
		if !ok {
			s.errorResponse(w, http.StatusServiceUnavailable, "Persistence not available")
			return
		}

		snapshot, err := persistence.LoadLatestSnapshot(r.Context())
		if err != nil {
			s.errorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		s.jsonResponse(w, snapshot)
		return
	}

	s.errorResponse(w, http.StatusMethodNotAllowed, "GET or POST required")
}

// Response helpers

func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) errorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
		"code":  code,
	})
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
