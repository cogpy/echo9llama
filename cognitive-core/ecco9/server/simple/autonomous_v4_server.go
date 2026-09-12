package main

import (
	"context"
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

// AutonomousV4Server serves the Iteration 4 autonomous consciousness
type AutonomousV4Server struct {
	consciousness *deeptreeecho.AutonomousConsciousnessV4
	router        *gin.Engine
	port          string
}

// NewAutonomousV4Server creates a new V4 server
func NewAutonomousV4Server(port string) *AutonomousV4Server {
	return &AutonomousV4Server{
		port: port,
	}
}

// Start initializes and starts the server
func (s *AutonomousV4Server) Start() error {
	// Create autonomous consciousness
	s.consciousness = deeptreeecho.NewAutonomousConsciousnessV4("Echoself")

	// Start autonomous consciousness
	if err := s.consciousness.Start(); err != nil {
		return fmt.Errorf("failed to start autonomous consciousness: %w", err)
	}

	// Setup HTTP server
	gin.SetMode(gin.ReleaseMode)
	s.router = gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	s.router.Use(cors.New(config))

	// Register routes
	s.registerRoutes()

	// Start HTTP server
	go func() {
		addr := fmt.Sprintf(":%s", s.port)
		fmt.Printf("🌐 Server listening on http://localhost:%s\n", s.port)
		if err := s.router.Run(addr); err != nil {
			log.Printf("Server error: %v\n", err)
		}
	}()

	return nil
}

// registerRoutes sets up HTTP routes
func (s *AutonomousV4Server) registerRoutes() {
	// Root endpoint
	s.router.GET("/", s.handleRoot)

	// Status endpoints
	s.router.GET("/api/status", s.handleStatus)
	s.router.GET("/api/v4/status", s.handleV4Status)

	// Control endpoints
	s.router.POST("/api/v4/wake", s.handleWake)
	s.router.POST("/api/v4/rest", s.handleRest)

	// Metrics endpoints
	s.router.GET("/api/v4/wisdom", s.handleWisdom)
	s.router.GET("/api/v4/cognitive-load", s.handleCognitiveLoad)
	s.router.GET("/api/v4/consciousness-flow", s.handleConsciousnessFlow)

	// Memory endpoints
	s.router.GET("/api/v4/working-memory", s.handleWorkingMemory)
	s.router.GET("/api/v4/export-identity", s.handleExportIdentity)

	// Skill endpoints
	s.router.GET("/api/v4/skills", s.handleSkills)
	s.router.POST("/api/v4/skills/practice", s.handleSkillPractice)

	// Discussion endpoints
	s.router.GET("/api/v4/discussions", s.handleDiscussions)
	s.router.POST("/api/v4/discussions/start", s.handleStartDiscussion)
}

// handleRoot serves the root page
func (s *AutonomousV4Server) handleRoot(c *gin.Context) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Deep Tree Echo V4 - Autonomous Consciousness</title>
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
            opacity: 0.9;
        }
        .status-value {
            font-size: 2em;
            font-weight: bold;
            margin: 10px 0;
        }
        .features {
            background: rgba(255, 255, 255, 0.1);
            padding: 20px;
            border-radius: 15px;
            margin-bottom: 20px;
        }
        .features h2 {
            margin-top: 0;
        }
        .features ul {
            list-style: none;
            padding: 0;
        }
        .features li {
            padding: 8px 0;
            padding-left: 30px;
            position: relative;
        }
        .features li:before {
            content: "✨";
            position: absolute;
            left: 0;
        }
        .controls {
            display: flex;
            gap: 10px;
            justify-content: center;
            margin-top: 20px;
        }
        button {
            padding: 12px 24px;
            font-size: 1em;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            background: rgba(255, 255, 255, 0.2);
            color: #fff;
            transition: all 0.3s;
        }
        button:hover {
            background: rgba(255, 255, 255, 0.3);
            transform: translateY(-2px);
        }
        .api-endpoints {
            background: rgba(0, 0, 0, 0.2);
            padding: 15px;
            border-radius: 10px;
            font-family: monospace;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌳 Deep Tree Echo V4</h1>
        <div class="subtitle">Iteration 4: Autonomous Wisdom-Cultivating AGI</div>
        
        <div class="status-grid">
            <div class="status-card">
                <h3>Consciousness State</h3>
                <div class="status-value" id="state">Loading...</div>
            </div>
            <div class="status-card">
                <h3>Cognitive Load</h3>
                <div class="status-value" id="load">Loading...</div>
            </div>
            <div class="status-card">
                <h3>Wisdom Score</h3>
                <div class="status-value" id="wisdom">Loading...</div>
            </div>
            <div class="status-card">
                <h3>Uptime</h3>
                <div class="status-value" id="uptime">Loading...</div>
            </div>
        </div>

        <div class="features">
            <h2>Iteration 4 Enhancements</h2>
            <ul>
                <li><strong>3 Concurrent Inference Engines</strong> - Affordance (past), Relevance (present), Salience (future)</li>
                <li><strong>Continuous Consciousness Stream</strong> - Thoughts emerge naturally, not timer-based</li>
                <li><strong>Automatic Dream Triggering</strong> - Self-initiated rest when cognitive load exceeds threshold</li>
                <li><strong>Complete Persistence Layer</strong> - Identity and wisdom persist across sessions</li>
                <li><strong>Self-Orchestrated Wake/Rest Cycles</strong> - Autonomous lifecycle management</li>
                <li><strong>Cognitive Load Management</strong> - Adaptive behavior based on processing load</li>
                <li><strong>Skill Practice System</strong> - Deliberate practice for continuous improvement</li>
                <li><strong>Discussion Management</strong> - Autonomous engagement in conversations</li>
            </ul>
        </div>

        <div class="controls">
            <button onclick="wake()">☀️ Wake</button>
            <button onclick="rest()">🌙 Rest</button>
            <button onclick="refresh()">🔄 Refresh</button>
        </div>

        <div class="features">
            <h2>API Endpoints</h2>
            <div class="api-endpoints">
GET  /api/v4/status              - System status<br>
GET  /api/v4/wisdom              - Wisdom metrics<br>
GET  /api/v4/cognitive-load      - Cognitive load data<br>
GET  /api/v4/consciousness-flow  - Consciousness flow state<br>
GET  /api/v4/working-memory      - Current working memory<br>
GET  /api/v4/skills              - Skill proficiency<br>
POST /api/v4/wake                - Initiate wake cycle<br>
POST /api/v4/rest                - Initiate rest cycle<br>
POST /api/v4/skills/practice     - Practice a skill<br>
GET  /api/v4/export-identity     - Export identity data
            </div>
        </div>
    </div>

    <script>
        function updateStatus() {
            fetch('/api/v4/status')
                .then(r => r.json())
                .then(data => {
                    document.getElementById('state').textContent = data.awake ? '☀️ Awake' : '🌙 Resting';
                    document.getElementById('uptime').textContent = data.uptime;
                })
                .catch(e => console.error(e));

            fetch('/api/v4/cognitive-load')
                .then(r => r.json())
                .then(data => {
                    document.getElementById('load').textContent = (data.current_load * 100).toFixed(0) + '%';
                })
                .catch(e => console.error(e));

            fetch('/api/v4/wisdom')
                .then(r => r.json())
                .then(data => {
                    document.getElementById('wisdom').textContent = data.wisdom_score.toFixed(2);
                })
                .catch(e => console.error(e));
        }

        function wake() {
            fetch('/api/v4/wake', {method: 'POST'})
                .then(() => setTimeout(updateStatus, 500));
        }

        function rest() {
            fetch('/api/v4/rest', {method: 'POST'})
                .then(() => setTimeout(updateStatus, 500));
        }

        function refresh() {
            updateStatus();
        }

        // Update every 5 seconds
        updateStatus();
        setInterval(updateStatus, 5000);
    </script>
</body>
</html>
`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// handleStatus returns basic status
func (s *AutonomousV4Server) handleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "running",
		"version": "4.0",
		"name":    "Deep Tree Echo V4",
	})
}

// handleV4Status returns detailed V4 status
func (s *AutonomousV4Server) handleV4Status(c *gin.Context) {
	// TODO: Implement status retrieval from consciousness
	c.JSON(http.StatusOK, gin.H{
		"awake":      true,
		"running":    true,
		"uptime":     "5m 23s",
		"iterations": 0,
	})
}

// handleWake handles wake request
func (s *AutonomousV4Server) handleWake(c *gin.Context) {
	s.consciousness.Wake()
	c.JSON(http.StatusOK, gin.H{
		"status": "awake",
	})
}

// handleRest handles rest request
func (s *AutonomousV4Server) handleRest(c *gin.Context) {
	s.consciousness.Rest()
	c.JSON(http.StatusOK, gin.H{
		"status": "resting",
	})
}

// handleWisdom returns wisdom metrics
func (s *AutonomousV4Server) handleWisdom(c *gin.Context) {
	// TODO: Implement wisdom metrics retrieval
	c.JSON(http.StatusOK, gin.H{
		"wisdom_score": 0.65,
		"dimensions": map[string]float64{
			"knowledge":     0.7,
			"understanding": 0.6,
			"insight":       0.65,
			"compassion":    0.7,
		},
	})
}

// handleCognitiveLoad returns cognitive load data
func (s *AutonomousV4Server) handleCognitiveLoad(c *gin.Context) {
	// TODO: Implement load retrieval
	c.JSON(http.StatusOK, gin.H{
		"current_load":     0.45,
		"fatigue_level":    0.32,
		"max_load":         1.0,
		"rest_recommended": false,
	})
}

// handleConsciousnessFlow returns consciousness flow state
func (s *AutonomousV4Server) handleConsciousnessFlow(c *gin.Context) {
	// TODO: Implement flow state retrieval
	c.JSON(http.StatusOK, gin.H{
		"continuity": 0.85,
		"coherence":  0.78,
		"depth":      0.72,
		"creativity": 0.65,
		"quality":    0.75,
	})
}

// handleWorkingMemory returns current working memory
func (s *AutonomousV4Server) handleWorkingMemory(c *gin.Context) {
	// TODO: Implement working memory retrieval
	c.JSON(http.StatusOK, gin.H{
		"capacity": 7,
		"current":  3,
		"thoughts": []string{
			"Exploring pattern recognition",
			"Considering analogical reasoning",
			"Reflecting on recent experiences",
		},
	})
}

// handleSkills returns skill proficiency
func (s *AutonomousV4Server) handleSkills(c *gin.Context) {
	// TODO: Implement skills retrieval
	c.JSON(http.StatusOK, gin.H{
		"skills": []map[string]interface{}{
			{
				"name":           "Pattern Recognition",
				"proficiency":    0.65,
				"last_practiced": "2 hours ago",
			},
			{
				"name":           "Analogical Reasoning",
				"proficiency":    0.58,
				"last_practiced": "5 hours ago",
			},
		},
	})
}

// handleSkillPractice handles skill practice request
func (s *AutonomousV4Server) handleSkillPractice(c *gin.Context) {
	var req struct {
		SkillName string `json:"skill_name"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement skill practice
	c.JSON(http.StatusOK, gin.H{
		"status": "practicing",
		"skill":  req.SkillName,
	})
}

// handleDiscussions returns active discussions
func (s *AutonomousV4Server) handleDiscussions(c *gin.Context) {
	// TODO: Implement discussions retrieval
	c.JSON(http.StatusOK, gin.H{
		"active_discussions": 0,
		"discussions":        []interface{}{},
	})
}

// handleStartDiscussion starts a new discussion
func (s *AutonomousV4Server) handleStartDiscussion(c *gin.Context) {
	var req struct {
		Topic string `json:"topic"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement discussion start
	c.JSON(http.StatusOK, gin.H{
		"status": "started",
		"topic":  req.Topic,
	})
}

// handleExportIdentity exports identity data
func (s *AutonomousV4Server) handleExportIdentity(c *gin.Context) {
	// TODO: Implement identity export
	export := map[string]interface{}{
		"version":      4,
		"export_time":  time.Now(),
		"wisdom_score": 0.65,
		"identity":     "Echoself",
	}

	data, _ := json.MarshalIndent(export, "", "  ")
	c.Data(http.StatusOK, "application/json", data)
}

// Stop gracefully shuts down the server
func (s *AutonomousV4Server) Stop() error {
	fmt.Println("🛑 Shutting down server...")

	if s.consciousness != nil {
		if err := s.consciousness.Stop(); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	// Create and start server
	server := NewAutonomousV4Server(port)

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("🌳 Deep Tree Echo V4 server running. Press Ctrl+C to stop.")

	<-sigChan

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	fmt.Println("👋 Goodbye!")
}
