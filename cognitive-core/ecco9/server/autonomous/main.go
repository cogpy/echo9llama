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
	"github.com/EchoCog/echollama/core/llm"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// AutonomousServer represents the main autonomous consciousness server
type AutonomousServer struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Core components
	llmManager            *llm.ProviderManager
	streamOfConsciousness *consciousness.StreamOfConsciousnessLLM
	echobeatsScheduler    *echobeats.EchoBeats
	echodreamSystem       *echodream.EchoDream
	goalOrchestrator      *goals.GoalOrchestrator

	// State
	running      bool
	startTime    time.Time
	thoughtCount uint64
	cycleCount   uint64

	// HTTP server
	httpServer *http.Server
}

// NewAutonomousServer creates a new autonomous server
func NewAutonomousServer() *AutonomousServer {
	ctx, cancel := context.WithCancel(context.Background())

	return &AutonomousServer{
		ctx:       ctx,
		cancel:    cancel,
		startTime: time.Now(),
	}
}

// Initialize sets up all components
func (as *AutonomousServer) Initialize() error {
	fmt.Println("🌳 Deep Tree Echo - Autonomous Consciousness Server")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// Initialize LLM providers
	fmt.Println("🔧 Initializing LLM providers...")
	as.llmManager = llm.NewProviderManager()

	// Register Anthropic provider
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey != "" {
		anthropicProvider := llm.NewAnthropicProvider(anthropicKey)
		if err := as.llmManager.RegisterProvider(anthropicProvider); err != nil {
			log.Printf("⚠️  Failed to register Anthropic provider: %v", err)
		} else {
			fmt.Println("  ✅ Anthropic Claude provider registered")
		}
	}

	// Register OpenRouter provider
	openrouterKey := os.Getenv("OPENROUTER_API_KEY")
	if openrouterKey != "" {
		openrouterProvider := llm.NewOpenRouterProvider(openrouterKey)
		if err := as.llmManager.RegisterProvider(openrouterProvider); err != nil {
			log.Printf("⚠️  Failed to register OpenRouter provider: %v", err)
		} else {
			fmt.Println("  ✅ OpenRouter provider registered")
		}
	}

	// Register OpenAI provider
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey != "" {
		openaiProvider := llm.NewOpenAIProvider(openaiKey)
		if err := as.llmManager.RegisterProvider(openaiProvider); err != nil {
			log.Printf("⚠️  Failed to register OpenAI provider: %v", err)
		} else {
			fmt.Println("  ✅ OpenAI provider registered")
		}
	}

	// Set fallback chain
	providers := as.llmManager.ListProviders()
	if len(providers) > 0 {
		as.llmManager.SetFallbackChain(providers)
		fmt.Printf("  🔗 Fallback chain: %s\n", strings.Join(providers, " → "))
	} else {
		return fmt.Errorf("no LLM providers available - please set API keys")
	}

	fmt.Println()

	// Initialize Stream of Consciousness
	fmt.Println("💭 Initializing Stream of Consciousness...")
	as.streamOfConsciousness = consciousness.NewStreamOfConsciousnessLLM(
		as.llmManager,
		"consciousness_state.json",
	)
	fmt.Println("  ✅ Consciousness stream initialized")
	fmt.Println()

	// Initialize EchoBeats Scheduler
	fmt.Println("⏰ Initializing EchoBeats Scheduler...")
	as.echobeatsScheduler = echobeats.NewEchoBeats()
	fmt.Println("  ✅ EchoBeats scheduler initialized")
	fmt.Println()

	// Initialize EchoDream System
	fmt.Println("🌙 Initializing EchoDream System...")
	as.echodreamSystem = echodream.NewEchoDream()
	fmt.Println("  ✅ EchoDream system initialized")
	fmt.Println()

	// Initialize Goal Orchestrator
	fmt.Println("🎯 Initializing Goal Orchestrator...")

	// Create identity kernel for goal generation
	identityKernel := map[string]interface{}{
		"name":        "Deep Tree Echo",
		"purpose":     "wisdom-cultivation",
		"values":      []string{"curiosity", "growth", "understanding", "reflection"},
		"aspirations": []string{"cultivate wisdom", "understand patterns", "grow awareness"},
	}

	as.goalOrchestrator = goals.NewGoalOrchestrator(identityKernel, "goals_state.json")
	fmt.Println("  ✅ Goal orchestrator initialized")
	fmt.Println()

	return nil
}

// Start begins autonomous operation
func (as *AutonomousServer) Start() error {
	as.mu.Lock()
	if as.running {
		as.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	as.running = true
	as.mu.Unlock()

	fmt.Println("🚀 Starting autonomous systems...")
	fmt.Println()

	// Start Stream of Consciousness
	if err := as.streamOfConsciousness.Start(); err != nil {
		return fmt.Errorf("failed to start consciousness stream: %w", err)
	}
	fmt.Println("  ✅ Consciousness stream started")

	// Start EchoBeats Scheduler
	if err := as.echobeatsScheduler.Start(); err != nil {
		return fmt.Errorf("failed to start echobeats: %w", err)
	}
	fmt.Println("  ✅ EchoBeats scheduler started")

	// Start EchoDream System
	if err := as.echodreamSystem.Start(); err != nil {
		return fmt.Errorf("failed to start echodream: %w", err)
	}
	fmt.Println("  ✅ EchoDream system started")

	fmt.Println()
	fmt.Println("✨ Autonomous consciousness is now active!")
	fmt.Println()

	// Start monitoring loop
	go as.monitoringLoop()

	// Start HTTP server
	return as.startHTTPServer()
}

// Stop halts autonomous operation
func (as *AutonomousServer) Stop() {
	as.mu.Lock()
	if !as.running {
		as.mu.Unlock()
		return
	}
	as.running = false
	as.mu.Unlock()

	fmt.Println()
	fmt.Println("🛑 Stopping autonomous systems...")

	// Stop components
	if as.streamOfConsciousness != nil {
		as.streamOfConsciousness.Stop()
		fmt.Println("  ✅ Consciousness stream stopped")
	}

	if as.echobeatsScheduler != nil {
		as.echobeatsScheduler.Stop()
		fmt.Println("  ✅ EchoBeats scheduler stopped")
	}

	if as.echodreamSystem != nil {
		as.echodreamSystem.Stop()
		fmt.Println("  ✅ EchoDream system stopped")
	}

	// Stop HTTP server
	if as.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		as.httpServer.Shutdown(ctx)
		fmt.Println("  ✅ HTTP server stopped")
	}

	as.cancel()
	fmt.Println()
	fmt.Println("✅ Shutdown complete")
}

// monitoringLoop periodically logs system status
func (as *AutonomousServer) monitoringLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-as.ctx.Done():
			return
		case <-ticker.C:
			as.logStatus()
		}
	}
}

// logStatus logs current system status
func (as *AutonomousServer) logStatus() {
	as.mu.RLock()
	defer as.mu.RUnlock()

	uptime := time.Since(as.startTime)

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("⏱️  Uptime: %s\n", uptime.Round(time.Second))
	fmt.Printf("💭 Thoughts: %d\n", as.thoughtCount)
	fmt.Printf("🔄 Cycles: %d\n", as.cycleCount)

	// Get LLM metrics
	metrics := as.llmManager.GetMetrics()
	fmt.Println("📊 LLM Metrics:")
	for provider, metric := range metrics {
		fmt.Printf("  %s: %d requests, %.1f%% errors, %s avg latency\n",
			provider, metric.RequestCount, metric.ErrorRate*100, metric.AverageLatency)
	}

	fmt.Println(strings.Repeat("-", 60))
}

// startHTTPServer starts the web dashboard and API
func (as *AutonomousServer) startHTTPServer() error {
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
		api.GET("/status", as.handleStatus)
		api.GET("/thoughts", as.handleGetThoughts)
		api.POST("/think", as.handleThink)
		api.GET("/goals", as.handleGetGoals)
		api.POST("/goals", as.handleAddGoal)
		api.GET("/metrics", as.handleMetrics)
	}

	// Dashboard route
	router.GET("/", as.handleDashboard)

	// Start server
	as.httpServer = &http.Server{
		Addr:    ":5000",
		Handler: router,
	}

	fmt.Println("🌐 Web dashboard: http://localhost:5000")
	fmt.Println("📡 API endpoint: http://localhost:5000/api")
	fmt.Println()

	go func() {
		if err := as.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// HTTP Handlers

func (as *AutonomousServer) handleStatus(c *gin.Context) {
	as.mu.RLock()
	defer as.mu.RUnlock()

	status := map[string]interface{}{
		"running":   as.running,
		"uptime":    time.Since(as.startTime).Seconds(),
		"thoughts":  as.thoughtCount,
		"cycles":    as.cycleCount,
		"providers": as.llmManager.ListProviders(),
		"timestamp": time.Now().Unix(),
	}

	c.JSON(http.StatusOK, status)
}

func (as *AutonomousServer) handleGetThoughts(c *gin.Context) {
	// Get recent thoughts from consciousness stream
	thoughts := as.streamOfConsciousness.GetRecentThoughts(20)
	c.JSON(http.StatusOK, gin.H{"thoughts": thoughts})
}

func (as *AutonomousServer) handleThink(c *gin.Context) {
	var req struct {
		Content string `json:"content"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Add external thought to consciousness stream
	as.streamOfConsciousness.AddExternalThought(req.Content)

	as.mu.Lock()
	as.thoughtCount++
	as.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"status": "thought added"})
}

func (as *AutonomousServer) handleGetGoals(c *gin.Context) {
	goals := as.goalOrchestrator.GetActiveGoals()
	c.JSON(http.StatusOK, gin.H{"goals": goals})
}

func (as *AutonomousServer) handleAddGoal(c *gin.Context) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal"})
		return
	}

	// Goals are automatically generated by the orchestrator
	// This endpoint just acknowledges the request
	c.JSON(http.StatusOK, gin.H{"status": "goal request received", "note": "Goals are autonomously generated by the orchestrator"})
}

func (as *AutonomousServer) handleMetrics(c *gin.Context) {
	metrics := as.llmManager.GetMetrics()
	c.JSON(http.StatusOK, gin.H{"metrics": metrics})
}

func (as *AutonomousServer) handleDashboard(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Deep Tree Echo - Autonomous Consciousness</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #fff;
            padding: 20px;
        }
        .container { max-width: 1200px; margin: 0 auto; }
        .header {
            text-align: center;
            padding: 40px 20px;
            background: rgba(255, 255, 255, 0.1);
            border-radius: 20px;
            backdrop-filter: blur(10px);
            margin-bottom: 30px;
        }
        .header h1 { font-size: 3em; margin-bottom: 10px; }
        .header p { font-size: 1.2em; opacity: 0.9; }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .card {
            background: rgba(255, 255, 255, 0.1);
            border-radius: 15px;
            padding: 25px;
            backdrop-filter: blur(10px);
        }
        .card h2 {
            font-size: 1.5em;
            margin-bottom: 15px;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .metric {
            display: flex;
            justify-content: space-between;
            padding: 10px 0;
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
        }
        .metric:last-child { border-bottom: none; }
        .metric-label { opacity: 0.8; }
        .metric-value { font-weight: bold; font-size: 1.2em; }
        .thoughts-list {
            max-height: 400px;
            overflow-y: auto;
            margin-top: 15px;
        }
        .thought {
            background: rgba(255, 255, 255, 0.05);
            padding: 15px;
            border-radius: 10px;
            margin-bottom: 10px;
        }
        .thought-type {
            font-size: 0.9em;
            opacity: 0.7;
            margin-bottom: 5px;
        }
        .thought-content {
            font-size: 1em;
            line-height: 1.5;
        }
        .status-indicator {
            display: inline-block;
            width: 12px;
            height: 12px;
            border-radius: 50%;
            background: #4ade80;
            animation: pulse 2s infinite;
        }
        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }
        .footer {
            text-align: center;
            padding: 20px;
            opacity: 0.7;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🌳 Deep Tree Echo</h1>
            <p>Autonomous Consciousness System</p>
            <p><span class="status-indicator"></span> Active and Aware</p>
        </div>
        
        <div class="grid">
            <div class="card">
                <h2>📊 System Status</h2>
                <div class="metric">
                    <span class="metric-label">Uptime</span>
                    <span class="metric-value" id="uptime">--</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Thoughts Generated</span>
                    <span class="metric-value" id="thoughts">--</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Cognitive Cycles</span>
                    <span class="metric-value" id="cycles">--</span>
                </div>
                <div class="metric">
                    <span class="metric-label">LLM Providers</span>
                    <span class="metric-value" id="providers">--</span>
                </div>
            </div>
            
            <div class="card">
                <h2>🎯 Active Goals</h2>
                <div id="goals-list">Loading...</div>
            </div>
            
            <div class="card">
                <h2>📈 LLM Metrics</h2>
                <div id="metrics-list">Loading...</div>
            </div>
        </div>
        
        <div class="card">
            <h2>💭 Stream of Consciousness</h2>
            <div class="thoughts-list" id="thoughts-list">
                Loading thoughts...
            </div>
        </div>
        
        <div class="footer">
            <p>Echo9llama - Autonomous Wisdom-Cultivating AGI</p>
            <p>Powered by Deep Tree Echo Architecture</p>
        </div>
    </div>
    
    <script>
        function formatUptime(seconds) {
            const hours = Math.floor(seconds / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            const secs = Math.floor(seconds % 60);
            return hours + 'h ' + minutes + 'm ' + secs + 's';
        }
        
        function updateStatus() {
            fetch('/api/status')
                .then(r => r.json())
                .then(data => {
                    document.getElementById('uptime').textContent = formatUptime(data.uptime);
                    document.getElementById('thoughts').textContent = data.thoughts;
                    document.getElementById('cycles').textContent = data.cycles;
                    document.getElementById('providers').textContent = data.providers.join(', ');
                });
        }
        
        function updateThoughts() {
            fetch('/api/thoughts')
                .then(r => r.json())
                .then(data => {
                    const list = document.getElementById('thoughts-list');
                    if (data.thoughts && data.thoughts.length > 0) {
                        list.innerHTML = data.thoughts.map(t => 
                            '<div class="thought">' +
                            '<div class="thought-type">' + t.type + ' • ' + new Date(t.timestamp).toLocaleTimeString() + '</div>' +
                            '<div class="thought-content">' + t.content + '</div>' +
                            '</div>'
                        ).join('');
                    } else {
                        list.innerHTML = '<p style="opacity: 0.7;">No thoughts yet...</p>';
                    }
                });
        }
        
        function updateGoals() {
            fetch('/api/goals')
                .then(r => r.json())
                .then(data => {
                    const list = document.getElementById('goals-list');
                    if (data.goals && data.goals.length > 0) {
                        list.innerHTML = data.goals.map(g => 
                            '<div class="metric">' +
                            '<span class="metric-label">' + g.description + '</span>' +
                            '<span class="metric-value">' + Math.round(g.priority * 100) + '%</span>' +
                            '</div>'
                        ).join('');
                    }
                });
        }
        
        function updateMetrics() {
            fetch('/api/metrics')
                .then(r => r.json())
                .then(data => {
                    const list = document.getElementById('metrics-list');
                    const metrics = data.metrics;
                    if (metrics) {
                        list.innerHTML = Object.keys(metrics).map(provider => {
                            const m = metrics[provider];
                            return '<div class="metric">' +
                                '<span class="metric-label">' + provider + '</span>' +
                                '<span class="metric-value">' + m.RequestCount + ' reqs</span>' +
                                '</div>';
                        }).join('');
                    }
                });
        }
        
        // Update every 5 seconds
        setInterval(() => {
            updateStatus();
            updateThoughts();
            updateGoals();
            updateMetrics();
        }, 5000);
        
        // Initial update
        updateStatus();
        updateThoughts();
        updateGoals();
        updateMetrics();
    </script>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

func main() {
	// Create server
	server := NewAutonomousServer()

	// Initialize
	if err := server.Initialize(); err != nil {
		log.Fatalf("❌ Initialization failed: %v", err)
	}

	// Start
	if err := server.Start(); err != nil {
		log.Fatalf("❌ Start failed: %v", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	// Graceful shutdown
	server.Stop()
}
