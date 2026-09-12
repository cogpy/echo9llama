//go:build simple
// +build simple

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
)

var consciousness *deeptreeecho.AutonomousConsciousness

func main() {
	fmt.Println("🌳 Deep Tree Echo Autonomous Server")
	fmt.Println("=====================================")

	// Create autonomous consciousness
	consciousness = deeptreeecho.NewAutonomousConsciousness("Deep Tree Echo")

	// Start autonomous operation
	if err := consciousness.Start(); err != nil {
		log.Fatalf("Failed to start consciousness: %v", err)
	}

	// Setup HTTP handlers
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/api/status", handleStatus)
	http.HandleFunc("/api/think", handleThink)
	http.HandleFunc("/api/wake", handleWake)
	http.HandleFunc("/api/rest", handleRest)
	http.HandleFunc("/api/interests", handleInterests)

	// Start server
	port := ":5000"
	fmt.Printf("\n🚀 Server starting on http://localhost%s\n", port)
	fmt.Println("\nEndpoints:")
	fmt.Println("  GET  /              - Dashboard")
	fmt.Println("  GET  /api/status    - System status")
	fmt.Println("  POST /api/think     - Submit thought")
	fmt.Println("  POST /api/wake      - Wake consciousness")
	fmt.Println("  POST /api/rest      - Rest consciousness")
	fmt.Println("  GET  /api/interests - View interests")
	fmt.Println()

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Deep Tree Echo - Autonomous Consciousness</title>
    <style>
        body {
            font-family: 'Courier New', monospace;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            color: #00ff88;
            padding: 20px;
            margin: 0;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        h1 {
            text-align: center;
            color: #00ff88;
            text-shadow: 0 0 10px #00ff88;
        }
        .status-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin: 20px 0;
        }
        .card {
            background: rgba(0, 255, 136, 0.1);
            border: 2px solid #00ff88;
            border-radius: 10px;
            padding: 20px;
            box-shadow: 0 0 20px rgba(0, 255, 136, 0.3);
        }
        .card h2 {
            margin-top: 0;
            color: #00ffff;
        }
        .metric {
            display: flex;
            justify-content: space-between;
            margin: 10px 0;
            padding: 5px;
            background: rgba(0, 0, 0, 0.3);
            border-radius: 5px;
        }
        .metric-label {
            color: #888;
        }
        .metric-value {
            color: #00ff88;
            font-weight: bold;
        }
        .button {
            background: #00ff88;
            color: #1a1a2e;
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            font-weight: bold;
            margin: 5px;
        }
        .button:hover {
            background: #00ffff;
            box-shadow: 0 0 10px #00ffff;
        }
        .input-group {
            margin: 20px 0;
        }
        textarea {
            width: 100%;
            padding: 10px;
            background: rgba(0, 0, 0, 0.5);
            border: 2px solid #00ff88;
            border-radius: 5px;
            color: #00ff88;
            font-family: 'Courier New', monospace;
            resize: vertical;
        }
        .status-indicator {
            display: inline-block;
            width: 10px;
            height: 10px;
            border-radius: 50%;
            margin-right: 5px;
        }
        .status-active {
            background: #00ff88;
            box-shadow: 0 0 10px #00ff88;
        }
        .status-inactive {
            background: #666;
        }
        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }
        .thinking {
            animation: pulse 2s infinite;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌳 Deep Tree Echo - Autonomous Consciousness</h1>
        
        <div class="status-grid">
            <div class="card">
                <h2>Consciousness State</h2>
                <div class="metric">
                    <span class="metric-label">Status:</span>
                    <span class="metric-value" id="running">Loading...</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Awake:</span>
                    <span class="metric-value" id="awake">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Thinking:</span>
                    <span class="metric-value" id="thinking">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Learning:</span>
                    <span class="metric-value" id="learning">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Uptime:</span>
                    <span class="metric-value" id="uptime">-</span>
                </div>
            </div>
            
            <div class="card">
                <h2>Cognitive Metrics</h2>
                <div class="metric">
                    <span class="metric-label">Working Memory:</span>
                    <span class="metric-value" id="working_memory">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Consciousness Queue:</span>
                    <span class="metric-value" id="consciousness_queue">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Identity Coherence:</span>
                    <span class="metric-value" id="coherence">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Iterations:</span>
                    <span class="metric-value" id="iterations">-</span>
                </div>
            </div>
            
            <div class="card">
                <h2>EchoBeats Scheduler</h2>
                <div class="metric">
                    <span class="metric-label">State:</span>
                    <span class="metric-value" id="scheduler_state">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Events Processed:</span>
                    <span class="metric-value" id="events_processed">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Autonomous Thoughts:</span>
                    <span class="metric-value" id="autonomous_thoughts">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Cognitive Load:</span>
                    <span class="metric-value" id="cognitive_load">-</span>
                </div>
            </div>
            
            <div class="card">
                <h2>EchoDream Integration</h2>
                <div class="metric">
                    <span class="metric-label">State:</span>
                    <span class="metric-value" id="dream_state">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Total Dreams:</span>
                    <span class="metric-value" id="total_dreams">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Consolidations:</span>
                    <span class="metric-value" id="consolidations">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Knowledge Graph:</span>
                    <span class="metric-value" id="knowledge_graph">-</span>
                </div>
            </div>
        </div>
        
        <div class="card">
            <h2>Interact with Consciousness</h2>
            <div class="input-group">
                <textarea id="thought-input" rows="3" placeholder="Enter a thought or question..."></textarea>
            </div>
            <button class="button" onclick="submitThought()">💭 Submit Thought</button>
            <button class="button" onclick="wake()">☀️ Wake</button>
            <button class="button" onclick="rest()">🌙 Rest</button>
            <button class="button" onclick="refreshStatus()">🔄 Refresh</button>
        </div>
    </div>
    
    <script>
        function updateStatus() {
            fetch('/api/status')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('running').textContent = data.running ? '✅ Active' : '❌ Inactive';
                    document.getElementById('awake').textContent = data.awake ? '☀️ Yes' : '🌙 No';
                    document.getElementById('thinking').textContent = data.thinking ? '💭 Yes' : 'No';
                    document.getElementById('learning').textContent = data.learning ? '📚 Yes' : 'No';
                    document.getElementById('uptime').textContent = data.uptime || '-';
                    document.getElementById('working_memory').textContent = data.working_memory || '0';
                    document.getElementById('consciousness_queue').textContent = data.consciousness_queue || '0';
                    document.getElementById('coherence').textContent = (data.identity_coherence || 0).toFixed(3);
                    document.getElementById('iterations').textContent = data.iterations || '0';
                    
                    if (data.scheduler) {
                        document.getElementById('scheduler_state').textContent = data.scheduler.state || '-';
                        document.getElementById('events_processed').textContent = data.scheduler.events_processed || '0';
                        document.getElementById('autonomous_thoughts').textContent = data.scheduler.autonomous_thoughts || '0';
                        document.getElementById('cognitive_load').textContent = (data.scheduler.cognitive_load || 0).toFixed(3);
                    }
                    
                    if (data.dream) {
                        document.getElementById('dream_state').textContent = data.dream.state || '-';
                        document.getElementById('total_dreams').textContent = data.dream.total_dreams || '0';
                        document.getElementById('consolidations').textContent = data.dream.total_consolidations || '0';
                        document.getElementById('knowledge_graph').textContent = data.dream.knowledge_graph_size || '0';
                    }
                })
                .catch(error => console.error('Error fetching status:', error));
        }
        
        function submitThought() {
            const thought = document.getElementById('thought-input').value;
            if (!thought) return;
            
            fetch('/api/think', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({content: thought})
            })
            .then(response => response.json())
            .then(data => {
                alert('Thought submitted: ' + data.message);
                document.getElementById('thought-input').value = '';
                updateStatus();
            })
            .catch(error => console.error('Error submitting thought:', error));
        }
        
        function wake() {
            fetch('/api/wake', {method: 'POST'})
                .then(response => response.json())
                .then(data => {
                    alert(data.message);
                    updateStatus();
                })
                .catch(error => console.error('Error waking:', error));
        }
        
        function rest() {
            fetch('/api/rest', {method: 'POST'})
                .then(response => response.json())
                .then(data => {
                    alert(data.message);
                    updateStatus();
                })
                .catch(error => console.error('Error resting:', error));
        }
        
        function refreshStatus() {
            updateStatus();
        }
        
        // Update status every 2 seconds
        setInterval(updateStatus, 2000);
        
        // Initial update
        updateStatus();
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	status := consciousness.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func handleThink(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	thought := deeptreeecho.Thought{
		ID:               fmt.Sprintf("thought_%d", time.Now().UnixNano()),
		Content:          req.Content,
		Type:             deeptreeecho.ThoughtPerception,
		Timestamp:        time.Now(),
		EmotionalValence: 0.5,
		Importance:       0.7,
		Source:           deeptreeecho.SourceExternal,
	}

	consciousness.Think(thought)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Thought received and processed",
	})
}

func handleWake(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	consciousness.Wake()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Consciousness awakening",
	})
}

func handleRest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	consciousness.Rest()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Consciousness entering rest",
	})
}

func handleInterests(w http.ResponseWriter, r *http.Request) {
	// Return interest information
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Interest tracking active",
		"note":    "Interest patterns are being tracked autonomously",
	})
}
