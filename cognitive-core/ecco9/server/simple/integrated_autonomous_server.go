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

var consciousness *deeptreeecho.IntegratedAutonomousConsciousness

func main() {
	fmt.Println("🌳 Deep Tree Echo - Integrated Autonomous Server")
	fmt.Println("==================================================")
	fmt.Println()

	// Create integrated autonomous consciousness
	consciousness = deeptreeecho.NewIntegratedAutonomousConsciousness("Deep Tree Echo")

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
	http.HandleFunc("/api/skills", handleSkills)
	http.HandleFunc("/api/memory", handleMemory)

	// Start server
	port := ":5000"
	fmt.Printf("\n🚀 Server starting on http://localhost%s\n", port)
	fmt.Println("\nEndpoints:")
	fmt.Println("  GET  /              - Dashboard")
	fmt.Println("  GET  /api/status    - System status")
	fmt.Println("  POST /api/think     - Submit thought")
	fmt.Println("  POST /api/wake      - Wake consciousness")
	fmt.Println("  POST /api/rest      - Rest consciousness")
	fmt.Println("  GET  /api/skills    - View skills")
	fmt.Println("  GET  /api/memory    - View memory graph")
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
    <title>Deep Tree Echo - Integrated Autonomous Consciousness</title>
    <style>
        body {
            font-family: 'Courier New', monospace;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            color: #00ff88;
            padding: 20px;
            margin: 0;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
        }
        h1 {
            text-align: center;
            color: #00ff88;
            text-shadow: 0 0 10px #00ff88;
        }
        .badge {
            display: inline-block;
            background: rgba(0, 255, 136, 0.2);
            border: 1px solid #00ff88;
            padding: 5px 10px;
            border-radius: 5px;
            margin: 5px;
            font-size: 0.9em;
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
        <h1>🌳 Deep Tree Echo - Integrated Autonomous Consciousness</h1>
        
        <div style="text-align: center; margin: 20px 0;">
            <span class="badge">✓ AAR Geometric Self-Awareness</span>
            <span class="badge">✓ 12-Step EchoBeats Loop</span>
            <span class="badge">✓ Enhanced LLM Integration</span>
            <span class="badge">✓ Hypergraph Memory</span>
            <span class="badge">✓ Persistent Wisdom</span>
        </div>
        
        <div class="status-grid">
            <div class="card">
                <h2>Consciousness State</h2>
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
                    <span class="metric-label">Iterations:</span>
                    <span class="metric-value" id="iterations">-</span>
                </div>
            </div>
            
            <div class="card">
                <h2>AAR Self-Awareness</h2>
                <div class="metric">
                    <span class="metric-label">Coherence:</span>
                    <span class="metric-value" id="coherence">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Stability:</span>
                    <span class="metric-value" id="stability">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Awareness:</span>
                    <span class="metric-value" id="awareness">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Narrative:</span>
                    <span class="metric-value" id="narrative" style="font-size: 0.8em;">-</span>
                </div>
            </div>
            
            <div class="card">
                <h2>Memory & Knowledge</h2>
                <div class="metric">
                    <span class="metric-label">Hypergraph Nodes:</span>
                    <span class="metric-value" id="nodes">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Hypergraph Edges:</span>
                    <span class="metric-value" id="edges">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Working Memory:</span>
                    <span class="metric-value" id="working_memory">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Identity Coherence:</span>
                    <span class="metric-value" id="identity_coherence">-</span>
                </div>
            </div>
            
            <div class="card">
                <h2>Skills & Practice</h2>
                <div class="metric">
                    <span class="metric-label">Total Skills:</span>
                    <span class="metric-value" id="total_skills">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Avg Proficiency:</span>
                    <span class="metric-value" id="avg_proficiency">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Practice Sessions:</span>
                    <span class="metric-value" id="practice_sessions">-</span>
                </div>
            </div>
        </div>
        
        <div class="card">
            <h2>Interact with Consciousness</h2>
            <div>
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
                    document.getElementById('awake').textContent = data.awake ? '☀️ Yes' : '🌙 No';
                    document.getElementById('thinking').textContent = data.thinking ? '💭 Yes' : 'No';
                    document.getElementById('learning').textContent = data.learning ? '📚 Yes' : 'No';
                    document.getElementById('iterations').textContent = data.iterations || '0';
                    
                    if (data.aar) {
                        document.getElementById('coherence').textContent = (data.aar.coherence || 0).toFixed(3);
                        document.getElementById('stability').textContent = (data.aar.stability || 0).toFixed(3);
                        document.getElementById('awareness').textContent = (data.aar.awareness || 0).toFixed(3);
                        document.getElementById('narrative').textContent = data.aar.narrative || '-';
                    }
                    
                    if (data.memory) {
                        document.getElementById('nodes').textContent = data.memory.nodes || '0';
                        document.getElementById('edges').textContent = data.memory.edges || '0';
                    }
                    
                    document.getElementById('working_memory').textContent = data.working_memory || '0';
                    document.getElementById('identity_coherence').textContent = (data.identity_coherence || 0).toFixed(3);
                    
                    if (data.skills) {
                        document.getElementById('total_skills').textContent = data.skills.total || '0';
                        document.getElementById('avg_proficiency').textContent = (data.skills.avg_proficiency || 0).toFixed(2);
                        document.getElementById('practice_sessions').textContent = data.skills.practice_count || '0';
                    }
                })
                .catch(error => console.error('Error:', error));
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
                document.getElementById('thought-input').value = '';
                updateStatus();
            })
            .catch(error => console.error('Error:', error));
        }
        
        function wake() {
            fetch('/api/wake', {method: 'POST'})
                .then(() => updateStatus());
        }
        
        function rest() {
            fetch('/api/rest', {method: 'POST'})
                .then(() => updateStatus());
        }
        
        function refreshStatus() {
            updateStatus();
        }
        
        // Auto-refresh every 2 seconds
        setInterval(updateStatus, 2000);
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	thought := deeptreeecho.Thought{
		Content:    req.Content,
		Type:       deeptreeecho.ThoughtPerception,
		Timestamp:  time.Now(),
		Importance: 0.7,
		Source:     deeptreeecho.SourceExternal,
	}

	consciousness.Think(thought)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "thought received",
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
		"status": "awakening",
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
		"status": "resting",
	})
}

func handleSkills(w http.ResponseWriter, r *http.Request) {
	status := consciousness.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"skills": status["skills"],
	})
}

func handleMemory(w http.ResponseWriter, r *http.Request) {
	status := consciousness.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"memory": status["memory"],
	})
}
