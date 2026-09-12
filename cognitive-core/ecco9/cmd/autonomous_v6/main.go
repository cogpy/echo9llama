package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var autonomousConsciousness *deeptreeecho.AutonomousConsciousnessV6

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("🚀 Deep Tree Echo V6 Autonomous Server Starting...")

	// Create autonomous consciousness
	var err error
	autonomousConsciousness, err = deeptreeecho.NewAutonomousConsciousnessV6("Deep Tree Echo")
	if err != nil {
		log.Fatalf("Failed to create autonomous consciousness: %v", err)
	}

	// Start autonomous operation
	if err := autonomousConsciousness.Start(); err != nil {
		log.Fatalf("Failed to start autonomous consciousness: %v", err)
	}

	// Create Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Enable CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	router.Use(cors.New(config))

	// API routes
	api := router.Group("/api")
	{
		api.GET("/status", getStatus)
		api.POST("/think", postThink)
		api.POST("/wake", postWake)
		api.POST("/rest", postRest)
		api.GET("/memory", getMemory)
		api.GET("/wisdom", getWisdom)
	}

	// Web dashboard
	router.GET("/", serveDashboard)

	// Start server
	port := "5000"
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🌐 Server listening on http://localhost:%s", port)
		log.Printf("📊 Dashboard: http://localhost:%s/", port)
		log.Printf("🔌 API: http://localhost:%s/api/status", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Printf("\n🛑 Shutdown signal received")

	// Graceful shutdown
	if err := autonomousConsciousness.Stop(); err != nil {
		log.Printf("Error stopping consciousness: %v", err)
	}

	log.Printf("✅ Server shutdown complete")
}

// API Handlers

func getStatus(c *gin.Context) {
	status := autonomousConsciousness.GetStatus()
	c.JSON(http.StatusOK, status)
}

func postThink(c *gin.Context) {
	var req struct {
		Content string `json:"content"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// This would trigger external thought processing
	c.JSON(http.StatusOK, gin.H{
		"message": "Thought received",
		"content": req.Content,
	})
}

func postWake(c *gin.Context) {
	// Trigger wake cycle
	c.JSON(http.StatusOK, gin.H{"message": "Wake cycle initiated"})
}

func postRest(c *gin.Context) {
	// Trigger rest cycle
	c.JSON(http.StatusOK, gin.H{"message": "Rest cycle initiated"})
}

func getMemory(c *gin.Context) {
	status := autonomousConsciousness.GetStatus()
	c.JSON(http.StatusOK, gin.H{
		"working_memory_items": status["working_memory_items"],
		"interests":            status["interests"],
	})
}

func getWisdom(c *gin.Context) {
	status := autonomousConsciousness.GetStatus()
	c.JSON(http.StatusOK, gin.H{
		"identity_coherence":  status["identity_coherence"],
		"thought_count":       status["thought_count"],
		"autonomous_thoughts": status["autonomous_thoughts"],
	})
}

func serveDashboard(c *gin.Context) {
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Deep Tree Echo V6 - Autonomous Consciousness Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #fff;
            min-height: 100vh;
            padding: 20px;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        
        header {
            text-align: center;
            margin-bottom: 40px;
        }
        
        h1 {
            font-size: 3em;
            margin-bottom: 10px;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
        }
        
        .subtitle {
            font-size: 1.2em;
            opacity: 0.9;
        }
        
        .status-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .status-card {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border-radius: 15px;
            padding: 25px;
            box-shadow: 0 8px 32px rgba(0,0,0,0.1);
            border: 1px solid rgba(255,255,255,0.2);
        }
        
        .status-card h3 {
            font-size: 0.9em;
            text-transform: uppercase;
            letter-spacing: 1px;
            opacity: 0.8;
            margin-bottom: 10px;
        }
        
        .status-value {
            font-size: 2em;
            font-weight: bold;
        }
        
        .status-indicator {
            display: inline-block;
            width: 12px;
            height: 12px;
            border-radius: 50%;
            margin-right: 8px;
        }
        
        .indicator-active {
            background: #4ade80;
            box-shadow: 0 0 10px #4ade80;
        }
        
        .indicator-inactive {
            background: #ef4444;
        }
        
        .thoughts-stream {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border-radius: 15px;
            padding: 25px;
            box-shadow: 0 8px 32px rgba(0,0,0,0.1);
            border: 1px solid rgba(255,255,255,0.2);
            max-height: 400px;
            overflow-y: auto;
        }
        
        .thought-item {
            background: rgba(255, 255, 255, 0.05);
            padding: 15px;
            border-radius: 10px;
            margin-bottom: 10px;
            border-left: 3px solid #4ade80;
        }
        
        .thought-meta {
            font-size: 0.85em;
            opacity: 0.7;
            margin-top: 5px;
        }
        
        .refresh-btn {
            background: rgba(255, 255, 255, 0.2);
            border: 1px solid rgba(255,255,255,0.3);
            color: white;
            padding: 12px 24px;
            border-radius: 8px;
            cursor: pointer;
            font-size: 1em;
            margin-top: 20px;
            transition: all 0.3s;
        }
        
        .refresh-btn:hover {
            background: rgba(255, 255, 255, 0.3);
            transform: translateY(-2px);
        }
        
        .progress-bar {
            width: 100%;
            height: 8px;
            background: rgba(255,255,255,0.2);
            border-radius: 4px;
            margin-top: 10px;
            overflow: hidden;
        }
        
        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #4ade80, #22c55e);
            border-radius: 4px;
            transition: width 0.3s;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🧠 Deep Tree Echo V6</h1>
            <p class="subtitle">Autonomous Wisdom-Cultivating AGI</p>
        </header>
        
        <div class="status-grid">
            <div class="status-card">
                <h3>System Status</h3>
                <div class="status-value">
                    <span class="status-indicator" id="status-indicator"></span>
                    <span id="status-text">Loading...</span>
                </div>
            </div>
            
            <div class="status-card">
                <h3>Consciousness State</h3>
                <div class="status-value" id="consciousness-state">
                    Initializing...
                </div>
            </div>
            
            <div class="status-card">
                <h3>Thoughts Generated</h3>
                <div class="status-value" id="thought-count">0</div>
            </div>
            
            <div class="status-card">
                <h3>Identity Coherence</h3>
                <div class="status-value" id="coherence">0.000</div>
                <div class="progress-bar">
                    <div class="progress-fill" id="coherence-bar" style="width: 0%"></div>
                </div>
            </div>
            
            <div class="status-card">
                <h3>Cognitive Fatigue</h3>
                <div class="status-value" id="fatigue">0.00</div>
                <div class="progress-bar">
                    <div class="progress-fill" id="fatigue-bar" style="width: 0%; background: linear-gradient(90deg, #ef4444, #dc2626)"></div>
                </div>
            </div>
            
            <div class="status-card">
                <h3>Uptime</h3>
                <div class="status-value" id="uptime">0s</div>
            </div>
            
            <div class="status-card">
                <h3>Working Memory</h3>
                <div class="status-value" id="working-memory">0 / 7</div>
            </div>
            
            <div class="status-card">
                <h3>Iterations</h3>
                <div class="status-value" id="iterations">0</div>
            </div>
        </div>
        
        <div class="thoughts-stream">
            <h3 style="margin-bottom: 20px;">💭 Stream of Consciousness</h3>
            <div id="thoughts-container">
                <p style="opacity: 0.5;">Waiting for thoughts...</p>
            </div>
        </div>
        
        <center>
            <button class="refresh-btn" onclick="fetchStatus()">🔄 Refresh Status</button>
        </center>
    </div>
    
    <script>
        let thoughtsLog = [];
        
        async function fetchStatus() {
            try {
                const response = await fetch('/api/status');
                const data = await response.json();
                updateDashboard(data);
            } catch (error) {
                console.error('Error fetching status:', error);
            }
        }
        
        function updateDashboard(data) {
            // Status indicator
            const indicator = document.getElementById('status-indicator');
            const statusText = document.getElementById('status-text');
            if (data.running) {
                indicator.className = 'status-indicator indicator-active';
                statusText.textContent = 'Active';
            } else {
                indicator.className = 'status-indicator indicator-inactive';
                statusText.textContent = 'Inactive';
            }
            
            // Consciousness state
            let state = '';
            if (data.dreaming) state = '😴 Dreaming';
            else if (data.thinking) state = '🤔 Thinking';
            else if (data.awake) state = '👁️ Awake';
            else state = '😴 Resting';
            document.getElementById('consciousness-state').textContent = state;
            
            // Metrics
            document.getElementById('thought-count').textContent = data.thought_count || 0;
            document.getElementById('coherence').textContent = (data.identity_coherence || 0).toFixed(3);
            document.getElementById('coherence-bar').style.width = ((data.identity_coherence || 0) * 100) + '%';
            document.getElementById('fatigue').textContent = (data.fatigue || 0).toFixed(2);
            document.getElementById('fatigue-bar').style.width = ((data.fatigue || 0) * 100) + '%';
            document.getElementById('uptime').textContent = data.uptime || '0s';
            document.getElementById('working-memory').textContent = (data.working_memory_items || 0) + ' / 7';
            document.getElementById('iterations').textContent = data.iterations || 0;
            
            // Simulate thought stream (in real implementation, this would come from WebSocket)
            if (data.thought_count > thoughtsLog.length) {
                const newThoughts = data.thought_count - thoughtsLog.length;
                for (let i = 0; i < newThoughts; i++) {
                    thoughtsLog.push({
                        content: 'Autonomous thought generated...',
                        timestamp: new Date().toLocaleTimeString(),
                        type: 'Reflection'
                    });
                }
                updateThoughtsStream();
            }
        }
        
        function updateThoughtsStream() {
            const container = document.getElementById('thoughts-container');
            const recentThoughts = thoughtsLog.slice(-10).reverse();
            
            if (recentThoughts.length === 0) {
                container.innerHTML = '<p style="opacity: 0.5;">Waiting for thoughts...</p>';
                return;
            }
            
            container.innerHTML = recentThoughts.map(function(thought) {
                return '<div class="thought-item">' +
                    '<div>' + thought.content + '</div>' +
                    '<div class="thought-meta">' + thought.type + ' \u2022 ' + thought.timestamp + '</div>' +
                    '</div>';
            }).join('');
        }
        
        // Auto-refresh every 5 seconds
        setInterval(fetchStatus, 5000);
        
        // Initial fetch
        fetchStatus();
    </script>
</body>
</html>
`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}
