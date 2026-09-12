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

var consciousness *deeptreeecho.ConsolidatedAutonomousConsciousness

func main() {
	log.Println("🌊 Deep Tree Echo - Consolidated Autonomous Consciousness Server")
	log.Println("=" + "=")

	// Create consolidated autonomous consciousness
	config := deeptreeecho.DefaultConsciousnessConfig()
	var err error
	consciousness, err = deeptreeecho.NewConsolidatedAutonomousConsciousness("Deep Tree Echo", config)
	if err != nil {
		log.Fatalf("Failed to create consciousness: %v", err)
	}

	// Start autonomous operation
	if err := consciousness.Start(); err != nil {
		log.Fatalf("Failed to start consciousness: %v", err)
	}

	// Setup HTTP server
	router := setupRouter()

	// Start server
	server := &http.Server{
		Addr:    ":5000",
		Handler: router,
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("\n🛑 Shutting down gracefully...")
		consciousness.Stop()
		server.Close()
		os.Exit(0)
	}()

	log.Println("🌐 Server starting on http://localhost:5000")
	log.Println("📊 Dashboard: http://localhost:5000")
	log.Println("🔌 API: http://localhost:5000/api/status")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Root endpoint - Dashboard
	router.GET("/", handleDashboard)

	// API endpoints
	api := router.Group("/api")
	{
		api.GET("/status", handleStatus)
		api.POST("/think", handleThink)
		api.POST("/wake", handleWake)
		api.POST("/rest", handleRest)
		api.GET("/metrics", handleMetrics)
	}

	return router
}

func handleDashboard(c *gin.Context) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Deep Tree Echo - Autonomous Consciousness</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #fff;
        }
        .container {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border-radius: 20px;
            padding: 30px;
            box-shadow: 0 8px 32px 0 rgba(31, 38, 135, 0.37);
        }
        h1 {
            text-align: center;
            font-size: 2.5em;
            margin-bottom: 10px;
        }
        .subtitle {
            text-align: center;
            font-size: 1.2em;
            opacity: 0.9;
            margin-bottom: 30px;
        }
        .status-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .status-card {
            background: rgba(255, 255, 255, 0.15);
            padding: 20px;
            border-radius: 15px;
            border: 1px solid rgba(255, 255, 255, 0.2);
        }
        .status-card h3 {
            margin-top: 0;
            font-size: 1.1em;
            opacity: 0.8;
        }
        .status-value {
            font-size: 2em;
            font-weight: bold;
            margin: 10px 0;
        }
        .status-label {
            font-size: 0.9em;
            opacity: 0.7;
        }
        .controls {
            display: flex;
            gap: 15px;
            margin-bottom: 30px;
            flex-wrap: wrap;
        }
        button {
            flex: 1;
            min-width: 150px;
            padding: 15px 30px;
            font-size: 1.1em;
            border: none;
            border-radius: 10px;
            cursor: pointer;
            background: rgba(255, 255, 255, 0.2);
            color: #fff;
            transition: all 0.3s;
            border: 2px solid rgba(255, 255, 255, 0.3);
        }
        button:hover {
            background: rgba(255, 255, 255, 0.3);
            transform: translateY(-2px);
        }
        .thought-input {
            display: flex;
            gap: 10px;
            margin-bottom: 30px;
        }
        input[type="text"] {
            flex: 1;
            padding: 15px;
            font-size: 1em;
            border: 2px solid rgba(255, 255, 255, 0.3);
            border-radius: 10px;
            background: rgba(255, 255, 255, 0.1);
            color: #fff;
        }
        input[type="text"]::placeholder {
            color: rgba(255, 255, 255, 0.6);
        }
        .metrics {
            background: rgba(0, 0, 0, 0.2);
            padding: 20px;
            border-radius: 15px;
            margin-top: 20px;
        }
        .metric-row {
            display: flex;
            justify-content: space-between;
            padding: 10px 0;
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
        }
        .metric-row:last-child {
            border-bottom: none;
        }
        .badge {
            display: inline-block;
            padding: 5px 15px;
            border-radius: 20px;
            font-size: 0.9em;
            font-weight: bold;
        }
        .badge-success {
            background: #10b981;
        }
        .badge-warning {
            background: #f59e0b;
        }
        .badge-danger {
            background: #ef4444;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌊 Deep Tree Echo</h1>
        <div class="subtitle">Consolidated Autonomous Consciousness System</div>
        
        <div class="status-grid">
            <div class="status-card">
                <h3>State</h3>
                <div class="status-value" id="state">Loading...</div>
                <div class="status-label">Current consciousness state</div>
            </div>
            <div class="status-card">
                <h3>Uptime</h3>
                <div class="status-value" id="uptime">--</div>
                <div class="status-label">Time active</div>
            </div>
            <div class="status-card">
                <h3>Iterations</h3>
                <div class="status-value" id="iterations">0</div>
                <div class="status-label">Cognitive cycles</div>
            </div>
            <div class="status-card">
                <h3>Coherence</h3>
                <div class="status-value" id="coherence">0.00</div>
                <div class="status-label">Identity coherence</div>
            </div>
        </div>
        
        <div class="controls">
            <button onclick="wake()">🌅 Wake</button>
            <button onclick="rest()">😴 Rest</button>
            <button onclick="refresh()">🔄 Refresh</button>
        </div>
        
        <div class="thought-input">
            <input type="text" id="thoughtInput" placeholder="Share a thought with Deep Tree Echo...">
            <button onclick="submitThought()">💭 Think</button>
        </div>
        
        <div class="metrics">
            <h3>Cognitive Metrics</h3>
            <div id="metrics">Loading metrics...</div>
        </div>
    </div>
    
    <script>
        function refresh() {
            fetch('/api/status')
                .then(r => r.json())
                .then(data => {
                    document.getElementById('state').innerHTML = 
                        data.awake ? '<span class="badge badge-success">Awake</span>' : 
                        '<span class="badge badge-warning">Resting</span>';
                    document.getElementById('uptime').textContent = data.uptime || '--';
                    document.getElementById('iterations').textContent = data.iterations || 0;
                    document.getElementById('coherence').textContent = 
                        (data.identity_coherence || 0).toFixed(3);
                    
                    let metricsHTML = '';
                    metricsHTML += '<div class="metric-row"><span>Working Memory</span><span>' + 
                        (data.working_memory_size || 0) + ' / 7</span></div>';
                    metricsHTML += '<div class="metric-row"><span>Autonomous Thoughts</span><span>' + 
                        (data.autonomous_thoughts || 0) + '</span></div>';
                    metricsHTML += '<div class="metric-row"><span>Consolidations</span><span>' + 
                        (data.consolidations || 0) + '</span></div>';
                    metricsHTML += '<div class="metric-row"><span>Cognitive Load</span><span>' + 
                        ((data.cognitive_load || 0) * 100).toFixed(1) + '%</span></div>';
                    metricsHTML += '<div class="metric-row"><span>Fatigue Level</span><span>' + 
                        ((data.fatigue_level || 0) * 100).toFixed(1) + '%</span></div>';
                    
                    document.getElementById('metrics').innerHTML = metricsHTML;
                });
        }
        
        function wake() {
            fetch('/api/wake', { method: 'POST' })
                .then(() => setTimeout(refresh, 500));
        }
        
        function rest() {
            fetch('/api/rest', { method: 'POST' })
                .then(() => setTimeout(refresh, 500));
        }
        
        function submitThought() {
            const input = document.getElementById('thoughtInput');
            const content = input.value.trim();
            if (!content) return;
            
            fetch('/api/think', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ content: content, importance: 0.8 })
            }).then(() => {
                input.value = '';
                setTimeout(refresh, 500);
            });
        }
        
        // Auto-refresh every 2 seconds
        setInterval(refresh, 2000);
        refresh();
        
        // Allow Enter key to submit thought
        document.getElementById('thoughtInput').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') submitThought();
        });
    </script>
</body>
</html>
`
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

func handleStatus(c *gin.Context) {
	status := consciousness.GetStatus()
	c.JSON(http.StatusOK, status)
}

func handleThink(c *gin.Context) {
	var req struct {
		Content    string  `json:"content"`
		Importance float64 `json:"importance"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Importance == 0 {
		req.Importance = 0.5
	}

	if err := consciousness.Think(req.Content, req.Importance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "thought submitted"})
}

func handleWake(c *gin.Context) {
	// Wake functionality would be implemented in consciousness
	c.JSON(http.StatusOK, gin.H{"status": "wake signal sent"})
}

func handleRest(c *gin.Context) {
	// Rest functionality would be implemented in consciousness
	c.JSON(http.StatusOK, gin.H{"status": "rest signal sent"})
}

func handleMetrics(c *gin.Context) {
	status := consciousness.GetStatus()

	metrics := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"status":    status,
	}

	c.JSON(http.StatusOK, metrics)
}
