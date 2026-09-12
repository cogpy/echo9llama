//go:build simple
// +build simple

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/EchoCog/echollama/core/echobeats"
)

var fiveChannelManager *echobeats.FiveChannelManager

func main() {
	fmt.Println("🌳 Deep Tree Echo - 5-Channel Stream-of-Consciousness System")
	fmt.Println("==============================================================")
	fmt.Println("3 Embodied Phases + 2 Global Orchestrators")
	fmt.Println("Opponent Processing + Narrative Continuity")

	// Create processor and integrator
	processor := echobeats.NewDefaultPhaseProcessor()
	integrator := echobeats.NewConsciousnessAdapter(nil)

	// Create 5-channel manager
	fiveChannelManager = echobeats.NewFiveChannelManager(processor, integrator)

	// Start the 5-channel system
	if err := fiveChannelManager.Start(); err != nil {
		log.Fatalf("Failed to start 5-channel system: %v", err)
	}

	// Setup HTTP handlers
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/api/status", handleStatus)
	http.HandleFunc("/api/metrics", handleMetrics)
	http.HandleFunc("/api/stabilizer", handleStabilizer)
	http.HandleFunc("/api/stop", handleStop)

	// Start server
	port := ":5002"
	fmt.Printf("\n🚀 Server starting on http://localhost%s\n", port)
	fmt.Println("\nEndpoints:")
	fmt.Println("  GET  /                - Dashboard")
	fmt.Println("  GET  /api/status      - System status")
	fmt.Println("  GET  /api/metrics     - Detailed metrics")
	fmt.Println("  GET  /api/stabilizer  - Current stabilizer")
	fmt.Println("  POST /api/stop        - Stop system")
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
    <title>Deep Tree Echo - 5-Channel Stream-of-Consciousness</title>
    <style>
        body {
            font-family: 'Courier New', monospace;
            background: linear-gradient(135deg, #0f0c29 0%, #302b63 50%, #24243e 100%);
            color: #00ff88;
            padding: 20px;
            margin: 0;
        }
        .container {
            max-width: 1600px;
            margin: 0 auto;
        }
        h1 {
            text-align: center;
            color: #00ff88;
            text-shadow: 0 0 20px #00ff88;
            font-size: 2em;
        }
        .subtitle {
            text-align: center;
            color: #00ffff;
            font-size: 1.2em;
            margin-bottom: 30px;
        }
        .channel-grid {
            display: grid;
            grid-template-columns: repeat(5, 1fr);
            gap: 15px;
            margin: 20px 0;
        }
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: 20px;
            margin: 20px 0;
        }
        .card {
            background: rgba(0, 255, 136, 0.1);
            border: 2px solid #00ff88;
            border-radius: 10px;
            padding: 15px;
            box-shadow: 0 0 20px rgba(0, 255, 136, 0.3);
        }
        .card h2 {
            margin-top: 0;
            color: #00ffff;
            font-size: 1.3em;
        }
        .global-card {
            border-color: #ffd700;
            background: rgba(255, 215, 0, 0.1);
            box-shadow: 0 0 20px rgba(255, 215, 0, 0.3);
        }
        .global-card h2 {
            color: #ffd700;
        }
        .metric {
            display: flex;
            justify-content: space-between;
            margin: 8px 0;
            padding: 6px;
            background: rgba(0, 0, 0, 0.3);
            border-radius: 5px;
            font-size: 0.9em;
        }
        .metric-label {
            color: #888;
        }
        .metric-value {
            color: #00ff88;
            font-weight: bold;
        }
        .stabilizer {
            padding: 15px;
            margin: 20px 0;
            background: rgba(255, 100, 255, 0.2);
            border-left: 4px solid #ff00ff;
            border-radius: 5px;
            font-size: 1.1em;
        }
        .button {
            background: #00ff88;
            color: #000;
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
    </style>
</head>
<body>
    <div class="container">
        <h1>🌳 Deep Tree Echo - 5-Channel Stream-of-Consciousness</h1>
        <div class="subtitle">3 Embodied Phases + 2 Global Orchestrators</div>
        
        <div class="stabilizer" id="stabilizer">
            <strong>Current Stabilizer:</strong> <span id="stabilizer-value">-</span>
        </div>
        
        <div class="channel-grid">
            <div class="card">
                <h2>Phase 0 (p0)</h2>
                <div class="metric">
                    <span class="metric-label">Steps:</span>
                    <span class="metric-value" id="p0-steps">0</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Expressive:</span>
                    <span class="metric-value" id="p0-expressive">0</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Reflective:</span>
                    <span class="metric-value" id="p0-reflective">0</span>
                </div>
            </div>
            
            <div class="card">
                <h2>Phase 1 (p1)</h2>
                <div class="metric">
                    <span class="metric-label">Steps:</span>
                    <span class="metric-value" id="p1-steps">0</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Expressive:</span>
                    <span class="metric-value" id="p1-expressive">0</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Reflective:</span>
                    <span class="metric-value" id="p1-reflective">0</span>
                </div>
            </div>
            
            <div class="card">
                <h2>Phase 2 (p2)</h2>
                <div class="metric">
                    <span class="metric-label">Steps:</span>
                    <span class="metric-value" id="p2-steps">0</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Expressive:</span>
                    <span class="metric-value" id="p2-expressive">0</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Reflective:</span>
                    <span class="metric-value" id="p2-reflective">0</span>
                </div>
            </div>
            
            <div class="card global-card">
                <h2>g2: Opponent</h2>
                <div class="metric">
                    <span class="metric-label">Steps:</span>
                    <span class="metric-value" id="g2-steps">0</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Current:</span>
                    <span class="metric-value" id="g2-current">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Pattern:</span>
                    <span class="metric-value">T9E-T9E-T8R-T8R</span>
                </div>
            </div>
            
            <div class="card global-card">
                <h2>g3: Narrative</h2>
                <div class="metric">
                    <span class="metric-label">Steps:</span>
                    <span class="metric-value" id="g3-steps">0</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Current:</span>
                    <span class="metric-value" id="g3-current">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Pattern:</span>
                    <span class="metric-value">T3E-T6R-T6E-T2R</span>
                </div>
            </div>
        </div>
        
        <div class="metrics-grid">
            <div class="card">
                <h2>System Metrics</h2>
                <div class="metric">
                    <span class="metric-label">Running:</span>
                    <span class="metric-value" id="running">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Current Step:</span>
                    <span class="metric-value" id="current-step">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Cycle Number:</span>
                    <span class="metric-value" id="cycle-number">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Total Steps:</span>
                    <span class="metric-value" id="total-steps">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Uptime:</span>
                    <span class="metric-value" id="uptime">-</span>
                </div>
            </div>
            
            <div class="card">
                <h2>Coherence Metrics</h2>
                <div class="metric">
                    <span class="metric-label">Cognitive Load:</span>
                    <span class="metric-value" id="cognitive-load">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Stream Coherence:</span>
                    <span class="metric-value" id="stream-coherence">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Identity Coherence:</span>
                    <span class="metric-value" id="identity-coherence">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Narrative Alignment:</span>
                    <span class="metric-value" id="narrative-alignment">-</span>
                </div>
            </div>
        </div>
        
        <div class="card" style="text-align: center;">
            <button class="button" onclick="refreshStatus()">🔄 Refresh</button>
            <button class="button" onclick="stopSystem()">⏹️ Stop System</button>
        </div>
    </div>
    
    <script>
        function updateStatus() {
            fetch('/api/status')
                .then(response => response.json())
                .then(data => {
                    // System metrics
                    document.getElementById('running').textContent = data.running ? '✅ Active' : '❌ Stopped';
                    document.getElementById('current-step').textContent = data.current_step;
                    document.getElementById('cycle-number').textContent = data.cycle_number;
                    document.getElementById('total-steps').textContent = data.total_steps;
                    document.getElementById('uptime').textContent = data.uptime;
                    
                    // Coherence metrics
                    document.getElementById('cognitive-load').textContent = data.cognitive_load.toFixed(3);
                    document.getElementById('stream-coherence').textContent = data.stream_coherence.toFixed(3);
                    document.getElementById('identity-coherence').textContent = data.identity_coherence.toFixed(3);
                    document.getElementById('narrative-alignment').textContent = data.narrative_alignment.toFixed(3);
                    
                    // Embodied phases
                    data.embodied_phases.forEach((phase, idx) => {
                        document.getElementById('p' + idx + '-steps').textContent = phase.steps_processed;
                        document.getElementById('p' + idx + '-expressive').textContent = phase.expressive_steps;
                        document.getElementById('p' + idx + '-reflective').textContent = phase.reflective_steps;
                    });
                    
                    // Global channels
                    data.global_channels.forEach(channel => {
                        const prefix = 'g' + channel.id;
                        document.getElementById(prefix + '-steps').textContent = channel.steps_processed;
                        document.getElementById(prefix + '-current').textContent = 
                            'T' + channel.current_term + (channel.current_mode === 0 ? 'E' : 'R');
                    });
                })
                .catch(error => console.error('Error fetching status:', error));
            
            // Update stabilizer
            fetch('/api/stabilizer')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('stabilizer-value').textContent = data.stabilizer;
                })
                .catch(error => console.error('Error fetching stabilizer:', error));
        }
        
        function refreshStatus() {
            updateStatus();
        }
        
        function stopSystem() {
            if (confirm('Stop the 5-channel system?')) {
                fetch('/api/stop', {method: 'POST'})
                    .then(response => response.json())
                    .then(data => {
                        alert(data.message);
                        updateStatus();
                    })
                    .catch(error => console.error('Error stopping system:', error));
            }
        }
        
        // Initialize
        updateStatus();
        
        // Update every 500ms
        setInterval(updateStatus, 500);
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	status := fiveChannelManager.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := fiveChannelManager.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func handleStabilizer(w http.ResponseWriter, r *http.Request) {
	stabilizer := fiveChannelManager.GetStabilizer()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"stabilizer": stabilizer.String(),
	})
}

func handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fiveChannelManager.Stop()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "5-channel system stopped",
	})
}
