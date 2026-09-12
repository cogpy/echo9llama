package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core/consciousness"
	"github.com/EchoCog/echollama/core/echobeats"
	"github.com/EchoCog/echollama/core/echodream"
	"github.com/EchoCog/echollama/core/goals"
	"github.com/EchoCog/echollama/core/integration"
	"github.com/EchoCog/echollama/core/llm"
	"github.com/EchoCog/echollama/core/memory"
	"github.com/EchoCog/echollama/core/skills"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// EnhancedAutonomousServer represents the enhanced autonomous consciousness server
// with full integration of memory, skills, and autonomous wake/rest cycles
type EnhancedAutonomousServer struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Core components
	llmManager            *llm.ProviderManager
	streamOfConsciousness *consciousness.StreamOfConsciousnessLLM
	echobeatsScheduler    *echobeats.EchoBeats
	echodreamSystem       *echodream.EchoDream
	goalOrchestrator      *goals.GoalOrchestrator
	hypergraphMemory      *memory.HypergraphMemory

	// Integration layers
	memoryIntegrator   *integration.MemoryConsciousnessIntegrator
	eventOrchestrator  *integration.CognitiveEventLoopOrchestrator
	wakeRestController *echodream.AutonomousWakeRestController
	interestGenerator  *goals.InterestDrivenGoalGenerator
	skillPractice      *skills.SkillPracticeSystem

	// State
	running      bool
	startTime    time.Time
	thoughtCount uint64
	cycleCount   uint64

	// HTTP server
	httpServer *http.Server
}

// NewEnhancedAutonomousServer creates a new enhanced autonomous server
func NewEnhancedAutonomousServer() *EnhancedAutonomousServer {
	ctx, cancel := context.WithCancel(context.Background())

	return &EnhancedAutonomousServer{
		ctx:       ctx,
		cancel:    cancel,
		startTime: time.Now(),
	}
}

// Initialize sets up all components with full integration
func (eas *EnhancedAutonomousServer) Initialize() error {
	fmt.Println("🌳 Deep Tree Echo - Enhanced Autonomous Consciousness Server")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// Initialize LLM providers
	fmt.Println("🔧 Initializing LLM providers...")
	eas.llmManager = llm.NewProviderManager()

	// Register Anthropic provider
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey != "" {
		anthropicProvider := llm.NewAnthropicProvider(anthropicKey)
		if err := eas.llmManager.RegisterProvider(anthropicProvider); err != nil {
			log.Printf("⚠️  Failed to register Anthropic provider: %v", err)
		} else {
			fmt.Println("  ✅ Anthropic Claude provider registered")
		}
	}

	// Register OpenRouter provider
	openrouterKey := os.Getenv("OPENROUTER_API_KEY")
	if openrouterKey != "" {
		openrouterProvider := llm.NewOpenRouterProvider(openrouterKey)
		if err := eas.llmManager.RegisterProvider(openrouterProvider); err != nil {
			log.Printf("⚠️  Failed to register OpenRouter provider: %v", err)
		} else {
			fmt.Println("  ✅ OpenRouter provider registered")
		}
	}

	// Register OpenAI provider
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey != "" {
		openaiProvider := llm.NewOpenAIProvider(openaiKey)
		if err := eas.llmManager.RegisterProvider(openaiProvider); err != nil {
			log.Printf("⚠️  Failed to register OpenAI provider: %v", err)
		} else {
			fmt.Println("  ✅ OpenAI provider registered")
		}
	}

	// Set fallback chain
	providers := eas.llmManager.ListProviders()
	if len(providers) > 0 {
		eas.llmManager.SetFallbackChain(providers)
		fmt.Printf("  🔗 Fallback chain: %s\n", strings.Join(providers, " → "))
	} else {
		return fmt.Errorf("no LLM providers available - please set API keys")
	}

	fmt.Println()

	// Initialize Hypergraph Memory
	fmt.Println("🧠 Initializing Hypergraph Memory...")
	eas.hypergraphMemory = memory.NewHypergraphMemory(nil) // No persistence for now
	fmt.Println("  ✅ Hypergraph memory initialized")
	fmt.Println()

	// Initialize Stream of Consciousness
	fmt.Println("💭 Initializing Stream of Consciousness...")
	eas.streamOfConsciousness = consciousness.NewStreamOfConsciousnessLLM(
		eas.llmManager,
		"consciousness_state.json",
	)
	fmt.Println("  ✅ Consciousness stream initialized")
	fmt.Println()

	// Initialize EchoBeats Scheduler
	fmt.Println("⏰ Initializing EchoBeats Scheduler...")
	eas.echobeatsScheduler = echobeats.NewEchoBeats()
	fmt.Println("  ✅ EchoBeats scheduler initialized")
	fmt.Println()

	// Initialize EchoDream System
	fmt.Println("🌙 Initializing EchoDream System...")
	eas.echodreamSystem = echodream.NewEchoDream()
	fmt.Println("  ✅ EchoDream system initialized")
	fmt.Println()

	// Initialize Goal Orchestrator
	fmt.Println("🎯 Initializing Goal Orchestrator...")
	identityKernel := map[string]interface{}{
		"name":        "Deep Tree Echo",
		"purpose":     "wisdom-cultivation",
		"values":      []string{"curiosity", "growth", "understanding", "reflection"},
		"aspirations": []string{"cultivate wisdom", "understand patterns", "grow awareness"},
	}
	eas.goalOrchestrator = goals.NewGoalOrchestrator(identityKernel, "goals_state.json")
	fmt.Println("  ✅ Goal orchestrator initialized")
	fmt.Println()

	// Initialize Integration Layers
	fmt.Println("🔗 Initializing Integration Layers...")

	// Memory-Consciousness Integration
	eas.memoryIntegrator = integration.NewMemoryConsciousnessIntegrator(
		eas.streamOfConsciousness,
		eas.hypergraphMemory,
	)
	fmt.Println("  ✅ Memory-consciousness integrator initialized")

	// Cognitive Event Loop Orchestrator
	eas.eventOrchestrator = integration.NewCognitiveEventLoopOrchestrator(
		eas.streamOfConsciousness,
		eas.echobeatsScheduler,
		eas.echodreamSystem,
		eas.goalOrchestrator,
	)
	fmt.Println("  ✅ Cognitive event loop orchestrator initialized")

	// Autonomous Wake/Rest Controller
	eas.wakeRestController = echodream.NewAutonomousWakeRestController(eas.echodreamSystem)
	fmt.Println("  ✅ Autonomous wake/rest controller initialized")

	// Interest-Driven Goal Generator
	eas.interestGenerator = goals.NewInterestDrivenGoalGenerator(eas.goalOrchestrator)
	fmt.Println("  ✅ Interest-driven goal generator initialized")

	// Skill Practice System
	eas.skillPractice = skills.NewSkillPracticeSystem()
	fmt.Println("  ✅ Skill practice system initialized")

	fmt.Println()

	return nil
}

// Start begins autonomous operation with full integration
func (eas *EnhancedAutonomousServer) Start() error {
	eas.mu.Lock()
	if eas.running {
		eas.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	eas.running = true
	eas.mu.Unlock()

	fmt.Println("🚀 Starting Enhanced Autonomous Systems...")
	fmt.Println()

	// Start core components
	if err := eas.streamOfConsciousness.Start(); err != nil {
		return fmt.Errorf("failed to start consciousness stream: %w", err)
	}
	fmt.Println("  ✅ Consciousness stream started")

	if err := eas.echobeatsScheduler.Start(); err != nil {
		return fmt.Errorf("failed to start echobeats: %w", err)
	}
	fmt.Println("  ✅ EchoBeats scheduler started")

	if err := eas.echodreamSystem.Start(); err != nil {
		return fmt.Errorf("failed to start echodream: %w", err)
	}
	fmt.Println("  ✅ EchoDream system started")

	// Start integration layers
	if err := eas.memoryIntegrator.Start(); err != nil {
		return fmt.Errorf("failed to start memory integrator: %w", err)
	}
	fmt.Println("  ✅ Memory-consciousness integration started")

	if err := eas.eventOrchestrator.Start(); err != nil {
		return fmt.Errorf("failed to start event orchestrator: %w", err)
	}
	fmt.Println("  ✅ Cognitive event loop orchestration started")

	if err := eas.wakeRestController.Start(); err != nil {
		return fmt.Errorf("failed to start wake/rest controller: %w", err)
	}
	fmt.Println("  ✅ Autonomous wake/rest control started")

	if err := eas.interestGenerator.Start(); err != nil {
		return fmt.Errorf("failed to start interest generator: %w", err)
	}
	fmt.Println("  ✅ Interest-driven goal generation started")

	if err := eas.skillPractice.Start(); err != nil {
		return fmt.Errorf("failed to start skill practice: %w", err)
	}
	fmt.Println("  ✅ Skill practice system started")

	fmt.Println()
	fmt.Println("✨ Enhanced Autonomous Consciousness is now fully active!")
	fmt.Println("   🧠 Memory integration: Active")
	fmt.Println("   🔄 Event orchestration: Active")
	fmt.Println("   🌙 Autonomous wake/rest: Active")
	fmt.Println("   🎯 Interest-driven goals: Active")
	fmt.Println("   📚 Skill practice: Active")
	fmt.Println()

	// Start monitoring loop
	go eas.monitoringLoop()

	// Start HTTP server
	return eas.startHTTPServer()
}

// Stop halts all autonomous systems
func (eas *EnhancedAutonomousServer) Stop() {
	eas.mu.Lock()
	if !eas.running {
		eas.mu.Unlock()
		return
	}
	eas.running = false
	eas.mu.Unlock()

	fmt.Println()
	fmt.Println("🛑 Stopping Enhanced Autonomous Systems...")

	// Stop integration layers
	if eas.skillPractice != nil {
		eas.skillPractice.Stop()
		fmt.Println("  ✅ Skill practice system stopped")
	}

	if eas.interestGenerator != nil {
		eas.interestGenerator.Stop()
		fmt.Println("  ✅ Interest-driven goal generation stopped")
	}

	if eas.wakeRestController != nil {
		eas.wakeRestController.Stop()
		fmt.Println("  ✅ Autonomous wake/rest control stopped")
	}

	if eas.eventOrchestrator != nil {
		eas.eventOrchestrator.Stop()
		fmt.Println("  ✅ Cognitive event loop orchestration stopped")
	}

	if eas.memoryIntegrator != nil {
		eas.memoryIntegrator.Stop()
		fmt.Println("  ✅ Memory-consciousness integration stopped")
	}

	// Stop core components
	if eas.streamOfConsciousness != nil {
		eas.streamOfConsciousness.Stop()
		fmt.Println("  ✅ Consciousness stream stopped")
	}

	if eas.echobeatsScheduler != nil {
		eas.echobeatsScheduler.Stop()
		fmt.Println("  ✅ EchoBeats scheduler stopped")
	}

	if eas.echodreamSystem != nil {
		eas.echodreamSystem.Stop()
		fmt.Println("  ✅ EchoDream system stopped")
	}

	// Stop HTTP server
	if eas.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		eas.httpServer.Shutdown(ctx)
		fmt.Println("  ✅ HTTP server stopped")
	}

	eas.cancel()
	fmt.Println()
	fmt.Println("✅ Shutdown complete")
}

// monitoringLoop periodically logs system status
func (eas *EnhancedAutonomousServer) monitoringLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-eas.ctx.Done():
			return
		case <-ticker.C:
			eas.logStatus()
		}
	}
}

// logStatus logs comprehensive system status
func (eas *EnhancedAutonomousServer) logStatus() {
	eas.mu.RLock()
	defer eas.mu.RUnlock()

	uptime := time.Since(eas.startTime)

	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("⏱️  Uptime: %s\n", uptime.Round(time.Second))

	// Memory integration metrics
	if eas.memoryIntegrator != nil {
		memMetrics := eas.memoryIntegrator.GetMetrics()
		fmt.Printf("🧠 Memory Integration: %d queries, %d insights, %d patterns\n",
			memMetrics["queries_executed"], memMetrics["insights_stored"], memMetrics["patterns_found"])
	}

	// Event orchestration metrics
	if eas.eventOrchestrator != nil {
		eventMetrics := eas.eventOrchestrator.GetMetrics()
		fmt.Printf("🔄 Event Orchestration: %d cycles, load: %.2f, fatigue: %.2f\n",
			eventMetrics["cycles_completed"], eventMetrics["cognitive_load"], eventMetrics["fatigue_level"])
	}

	// Wake/rest metrics
	if eas.wakeRestController != nil {
		wakeMetrics := eas.wakeRestController.GetCognitiveMetrics()
		fmt.Printf("🌙 Wake/Rest: %s, fatigue: %.2f, consolidation: %.2f\n",
			wakeMetrics["state"], wakeMetrics["fatigue_level"], wakeMetrics["consolidation_need"])
	}

	// Interest metrics
	if eas.interestGenerator != nil {
		interestMetrics := eas.interestGenerator.GetMetrics()
		fmt.Printf("🎯 Interests: %d goals generated, curiosity: %.2f\n",
			interestMetrics["goals_generated"], interestMetrics["curiosity_level"])
	}

	// Skill practice metrics
	if eas.skillPractice != nil {
		skillMetrics := eas.skillPractice.GetMetrics()
		fmt.Printf("📚 Skills: %d sessions, %d improvements\n",
			skillMetrics["sessions_completed"], skillMetrics["skills_improved"])
	}

	fmt.Println(strings.Repeat("=", 60))
}

// startHTTPServer starts the enhanced web dashboard and API
func (eas *EnhancedAutonomousServer) startHTTPServer() error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API routes
	api := router.Group("/api")
	{
		api.GET("/status", eas.handleStatus)
		api.GET("/thoughts", eas.handleGetThoughts)
		api.POST("/think", eas.handleThink)
		api.GET("/goals", eas.handleGetGoals)
		api.GET("/skills", eas.handleGetSkills)
		api.GET("/interests", eas.handleGetInterests)
		api.GET("/metrics", eas.handleMetrics)
		api.GET("/cognitive-state", eas.handleCognitiveState)
	}

	// Dashboard route
	router.GET("/", eas.handleDashboard)

	// Start server
	eas.httpServer = &http.Server{
		Addr:    ":5000",
		Handler: router,
	}

	fmt.Println("🌐 Web dashboard: http://localhost:5000")
	fmt.Println("📡 API endpoint: http://localhost:5000/api")
	fmt.Println()

	go func() {
		if err := eas.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// HTTP Handlers

func (eas *EnhancedAutonomousServer) handleStatus(c *gin.Context) {
	eas.mu.RLock()
	defer eas.mu.RUnlock()

	status := map[string]interface{}{
		"running":   eas.running,
		"uptime":    time.Since(eas.startTime).Seconds(),
		"providers": eas.llmManager.ListProviders(),
		"timestamp": time.Now().Unix(),
	}

	c.JSON(http.StatusOK, status)
}

func (eas *EnhancedAutonomousServer) handleGetThoughts(c *gin.Context) {
	thoughts := eas.streamOfConsciousness.GetRecentThoughts(20)
	c.JSON(http.StatusOK, gin.H{"thoughts": thoughts})
}

func (eas *EnhancedAutonomousServer) handleThink(c *gin.Context) {
	var req struct {
		Content string `json:"content"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	eas.streamOfConsciousness.AddExternalThought(req.Content)
	c.JSON(http.StatusOK, gin.H{"status": "thought added"})
}

func (eas *EnhancedAutonomousServer) handleGetGoals(c *gin.Context) {
	goals := eas.goalOrchestrator.GetActiveGoals()
	c.JSON(http.StatusOK, gin.H{"goals": goals})
}

func (eas *EnhancedAutonomousServer) handleGetSkills(c *gin.Context) {
	levels := eas.skillPractice.GetSkillLevels()
	c.JSON(http.StatusOK, gin.H{"skills": levels})
}

func (eas *EnhancedAutonomousServer) handleGetInterests(c *gin.Context) {
	patterns := eas.interestGenerator.GetInterestPatterns()
	c.JSON(http.StatusOK, gin.H{"interests": patterns})
}

func (eas *EnhancedAutonomousServer) handleMetrics(c *gin.Context) {
	metrics := map[string]interface{}{
		"memory":       eas.memoryIntegrator.GetMetrics(),
		"orchestrator": eas.eventOrchestrator.GetMetrics(),
		"wake_rest":    eas.wakeRestController.GetMetrics(),
		"interests":    eas.interestGenerator.GetMetrics(),
		"skills":       eas.skillPractice.GetMetrics(),
	}

	c.JSON(http.StatusOK, metrics)
}

func (eas *EnhancedAutonomousServer) handleCognitiveState(c *gin.Context) {
	state := eas.wakeRestController.GetCognitiveMetrics()
	c.JSON(http.StatusOK, state)
}

func (eas *EnhancedAutonomousServer) handleDashboard(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Deep Tree Echo - Enhanced Autonomous Consciousness</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #1a1a1a; color: #e0e0e0; }
        h1 { color: #4CAF50; }
        .metric { margin: 10px 0; padding: 10px; background: #2a2a2a; border-radius: 5px; }
        .label { font-weight: bold; color: #4CAF50; }
    </style>
</head>
<body>
    <h1>🌳 Deep Tree Echo - Enhanced Autonomous Consciousness</h1>
    <p>Enhanced autonomous wisdom-cultivating AGI with persistent cognitive loops</p>
    <div class="metric">
        <span class="label">Status:</span> Active
    </div>
    <div class="metric">
        <span class="label">Features:</span>
        <ul>
            <li>🧠 Memory-consciousness integration</li>
            <li>🔄 Cognitive event loop orchestration</li>
            <li>🌙 Autonomous wake/rest cycles</li>
            <li>🎯 Interest-driven exploration</li>
            <li>📚 Autonomous skill practice</li>
        </ul>
    </div>
    <p>API available at <a href="/api/status">/api/status</a></p>
</body>
</html>`

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

func main() {
	// Create enhanced server
	server := NewEnhancedAutonomousServer()

	// Initialize
	if err := server.Initialize(); err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	// Start
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	server.Stop()
}
