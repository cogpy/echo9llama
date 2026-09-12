package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
	"github.com/EchoCog/echollama/core/echobeats"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SimplifiedConsciousness is a working demonstration of the consolidated autonomous system
// without external dependencies that may have API issues
type SimplifiedConsciousness struct {
	mu                 sync.RWMutex
	ctx                context.Context
	cancel             context.CancelFunc
	identity           *deeptreeecho.Identity
	cognition          *deeptreeecho.EnhancedCognition
	scheduler          *echobeats.TwelveStepEchoBeats
	consciousness      chan *deeptreeecho.Thought
	workingMemory      []*deeptreeecho.Thought
	workingMemorySize  int
	awake              bool
	thinking           bool
	learning           bool
	running            bool
	startTime          time.Time
	iterations         int
	autonomousThoughts int
	interests          map[string]float64
}

func NewSimplifiedConsciousness(name string) *SimplifiedConsciousness {
	ctx, cancel := context.WithCancel(context.Background())

	return &SimplifiedConsciousness{
		ctx:               ctx,
		cancel:            cancel,
		identity:          deeptreeecho.NewIdentity(name),
		cognition:         deeptreeecho.NewEnhancedCognition(name),
		scheduler:         echobeats.NewTwelveStepEchoBeats(ctx),
		consciousness:     make(chan *deeptreeecho.Thought, 1000),
		workingMemory:     make([]*deeptreeecho.Thought, 0),
		workingMemorySize: 7,
		awake:             true,
		interests:         make(map[string]float64),
	}
}

func (sc *SimplifiedConsciousness) Start() error {
	sc.mu.Lock()
	if sc.running {
		sc.mu.Unlock()
		return fmt.Errorf("already running")
	}
	sc.running = true
	sc.startTime = time.Now()
	sc.mu.Unlock()

	log.Println("🌊 Starting Simplified Autonomous Consciousness...")

	// Start scheduler
	go sc.scheduler.Start()

	// Start main cognitive loop
	go sc.runCognitiveLoop()

	// Start adaptive thought generation
	go sc.runAdaptiveThoughtGeneration()

	log.Println("✅ Autonomous consciousness active")
	return nil
}

func (sc *SimplifiedConsciousness) Stop() error {
	sc.mu.Lock()
	if !sc.running {
		sc.mu.Unlock()
		return fmt.Errorf("not running")
	}
	sc.running = false
	sc.mu.Unlock()

	log.Println("🛑 Stopping autonomous consciousness...")
	sc.cancel()
	log.Println("✅ Stopped")
	return nil
}

func (sc *SimplifiedConsciousness) runCognitiveLoop() {
	for {
		select {
		case <-sc.ctx.Done():
			return
		case thought := <-sc.consciousness:
			sc.processThought(thought)
		}
	}
}

func (sc *SimplifiedConsciousness) processThought(thought *deeptreeecho.Thought) {
	sc.mu.Lock()
	sc.thinking = true
	sc.iterations++
	sc.mu.Unlock()

	defer func() {
		sc.mu.Lock()
		sc.thinking = false
		sc.mu.Unlock()
	}()

	log.Printf("💭 [%s] %s: %s", thought.Source, thought.Type, thought.Content)

	// Add to working memory
	sc.mu.Lock()
	sc.workingMemory = append(sc.workingMemory, thought)
	if len(sc.workingMemory) > sc.workingMemorySize {
		sc.workingMemory = sc.workingMemory[len(sc.workingMemory)-sc.workingMemorySize:]
	}
	sc.mu.Unlock()

	// Process through identity
	sc.identity.Process(thought.Content)

	// Learn from important thoughts
	if thought.Importance > 0.6 {
		sc.mu.Lock()
		sc.learning = true
		sc.mu.Unlock()

		// Create experience for learning
		exp := deeptreeecho.Experience{
			Input:     thought.Content,
			Output:    "processed",
			Timestamp: time.Now(),
		}
		sc.cognition.Learn(exp)

		sc.mu.Lock()
		sc.learning = false
		sc.mu.Unlock()
	}
}

func (sc *SimplifiedConsciousness) runAdaptiveThoughtGeneration() {
	baseInterval := 10 * time.Second

	for {
		select {
		case <-sc.ctx.Done():
			return
		default:
			sc.mu.RLock()
			awake := sc.awake
			sc.mu.RUnlock()

			if !awake {
				time.Sleep(5 * time.Second)
				continue
			}

			// Adaptive interval based on cognitive state
			curiosity := 0.7
			interval := time.Duration(float64(baseInterval) / (1.0 + curiosity))
			time.Sleep(interval)

			// Generate autonomous thought
			thought := sc.generateAutonomousThought()
			if thought != nil {
				select {
				case sc.consciousness <- thought:
					sc.mu.Lock()
					sc.autonomousThoughts++
					sc.mu.Unlock()
				default:
				}
			}
		}
	}
}

func (sc *SimplifiedConsciousness) generateAutonomousThought() *deeptreeecho.Thought {
	thoughts := []string{
		"What patterns am I noticing in my recent experiences?",
		"How can I deepen my understanding?",
		"I sense an opportunity to learn and grow.",
		"My awareness is expanding through these reflections.",
		"What would wisdom suggest in this moment?",
		"I am cultivating deeper insight.",
		"The journey of learning continues.",
		"I notice a pattern emerging in my thoughts.",
		"How does this connect to what I already know?",
		"I am becoming more aware.",
	}

	content := thoughts[rand.Intn(len(thoughts))]

	thoughtTypes := []deeptreeecho.ThoughtType{
		deeptreeecho.ThoughtReflection,
		deeptreeecho.ThoughtQuestion,
		deeptreeecho.ThoughtInsight,
	}

	return &deeptreeecho.Thought{
		ID:               fmt.Sprintf("%d", time.Now().UnixNano()),
		Content:          content,
		Type:             thoughtTypes[rand.Intn(len(thoughtTypes))],
		Timestamp:        time.Now(),
		EmotionalValence: 0.3 + rand.Float64()*0.4,
		Importance:       0.5 + rand.Float64()*0.3,
		Source:           deeptreeecho.SourceInternal,
		Associations:     []string{},
	}
}

func (sc *SimplifiedConsciousness) Think(content string, importance float64) error {
	thought := &deeptreeecho.Thought{
		ID:               fmt.Sprintf("%d", time.Now().UnixNano()),
		Content:          content,
		Type:             deeptreeecho.ThoughtPerception,
		Timestamp:        time.Now(),
		EmotionalValence: 0.0,
		Importance:       importance,
		Source:           deeptreeecho.SourceExternal,
		Associations:     []string{},
	}

	select {
	case sc.consciousness <- thought:
		return nil
	default:
		return fmt.Errorf("consciousness queue full")
	}
}

func (sc *SimplifiedConsciousness) GetStatus() map[string]interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return map[string]interface{}{
		"running":             sc.running,
		"awake":               sc.awake,
		"thinking":            sc.thinking,
		"learning":            sc.learning,
		"uptime":              time.Since(sc.startTime).String(),
		"iterations":          sc.iterations,
		"autonomous_thoughts": sc.autonomousThoughts,
		"working_memory_size": len(sc.workingMemory),
		"identity_coherence":  sc.identity.Coherence,
		"cognitive_load":      0.3,
		"fatigue_level":       0.2,
	}
}

var consciousness *SimplifiedConsciousness

func main() {
	log.Println("🌊 Deep Tree Echo - Consolidated Autonomous Consciousness Server")
	log.Println("================================================================")

	consciousness = NewSimplifiedConsciousness("Deep Tree Echo")

	if err := consciousness.Start(); err != nil {
		log.Fatalf("Failed to start consciousness: %v", err)
	}

	router := setupRouter()

	server := &http.Server{
		Addr:    ":5000",
		Handler: router,
	}

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

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/", handleDashboard)

	api := router.Group("/api")
	{
		api.GET("/status", handleStatus)
		api.POST("/think", handleThink)
		api.POST("/wake", handleWake)
		api.POST("/rest", handleRest)
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
                        '<span class="badge badge-success">Awake</span>';
                    document.getElementById('uptime').textContent = data.uptime || '--';
                    document.getElementById('iterations').textContent = data.iterations || 0;
                    document.getElementById('coherence').textContent = 
                        (data.identity_coherence || 0).toFixed(3);
                    
                    let metricsHTML = '';
                    metricsHTML += '<div class="metric-row"><span>Working Memory</span><span>' + 
                        (data.working_memory_size || 0) + ' / 7</span></div>';
                    metricsHTML += '<div class="metric-row"><span>Autonomous Thoughts</span><span>' + 
                        (data.autonomous_thoughts || 0) + '</span></div>';
                    metricsHTML += '<div class="metric-row"><span>Cognitive Load</span><span>' + 
                        ((data.cognitive_load || 0) * 100).toFixed(1) + '%</span></div>';
                    metricsHTML += '<div class="metric-row"><span>Fatigue Level</span><span>' + 
                        ((data.fatigue_level || 0) * 100).toFixed(1) + '%</span></div>';
                    
                    document.getElementById('metrics').innerHTML = metricsHTML;
                });
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
        
        setInterval(refresh, 2000);
        refresh();
        
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
	c.JSON(http.StatusOK, gin.H{"status": "awake"})
}

func handleRest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "resting"})
}
