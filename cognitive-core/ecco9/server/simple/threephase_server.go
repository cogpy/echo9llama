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

var threePhaseManager *echobeats.ThreePhaseManager

func main() {
	fmt.Println("🌳 Deep Tree Echo - 3-Phase Concurrent Inference Engine")
	fmt.Println("========================================================")

	// Create processor and integrator
	processor := echobeats.NewDefaultPhaseProcessor()
	integrator := echobeats.NewConsciousnessAdapter(nil)

	// Create 3-phase manager
	threePhaseManager = echobeats.NewThreePhaseManager(processor, integrator)

	// Start the 3-phase system
	if err := threePhaseManager.Start(); err != nil {
		log.Fatalf("Failed to start 3-phase system: %v", err)
	}

	// Setup HTTP handlers
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/api/status", handleStatus)
	http.HandleFunc("/api/metrics", handleMetrics)
	http.HandleFunc("/api/stop", handleStop)

	// Start server
	port := ":5001"
	fmt.Printf("\n🚀 Server starting on http://localhost%s\n", port)
	fmt.Println("\nEndpoints:")
	fmt.Println("  GET  /              - Dashboard")
	fmt.Println("  GET  /api/status    - System status")
	fmt.Println("  GET  /api/metrics   - Detailed metrics")
	fmt.Println("  POST /api/stop      - Stop system")
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
    <title>Deep Tree Echo - 3-Phase Concurrent Inference</title>
    <style>
        body {
            font-family: 'Courier New', monospace;
            background: linear-gradient(135deg, #0f0c29 0%, #302b63 50%, #24243e 100%);
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
            text-shadow: 0 0 20px #00ff88;
            font-size: 2em;
        }
        .subtitle {
            text-align: center;
            color: #00ffff;
            font-size: 1.2em;
            margin-bottom: 30px;
        }
        .phase-grid {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 20px;
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
            padding: 20px;
            box-shadow: 0 0 20px rgba(0, 255, 136, 0.3);
        }
        .card h2 {
            margin-top: 0;
            color: #00ffff;
            font-size: 1.5em;
        }
        .metric {
            display: flex;
            justify-content: space-between;
            margin: 10px 0;
            padding: 8px;
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
        .phase-indicator {
            width: 100%;
            height: 40px;
            background: rgba(0, 0, 0, 0.5);
            border-radius: 5px;
            margin: 10px 0;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 1.2em;
            font-weight: bold;
        }
        .phase-active {
            background: linear-gradient(90deg, #00ff88, #00ffff);
            color: #000;
            animation: pulse 1s infinite;
        }
        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.7; }
        }
        .step-cycle {
            display: grid;
            grid-template-columns: repeat(12, 1fr);
            gap: 5px;
            margin: 20px 0;
        }
        .step {
            height: 60px;
            background: rgba(0, 0, 0, 0.5);
            border: 1px solid #444;
            border-radius: 5px;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            font-size: 0.8em;
        }
        .step-current {
            border: 3px solid #00ff88;
            box-shadow: 0 0 15px #00ff88;
        }
        .step-expressive {
            background: rgba(255, 100, 100, 0.3);
        }
        .step-reflective {
            background: rgba(100, 100, 255, 0.3);
        }
        .coupling-indicator {
            padding: 10px;
            margin: 10px 0;
            background: rgba(255, 215, 0, 0.2);
            border-left: 4px solid #ffd700;
            border-radius: 5px;
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
        <h1>🌳 Deep Tree Echo - 3-Phase Concurrent Inference Engine</h1>
        <div class="subtitle">Inspired by Kawaii Hexapod System 4 Tripod Gait</div>
        
        <div class="card">
            <h2>12-Step Cognitive Cycle</h2>
            <div class="step-cycle" id="step-cycle"></div>
            <div class="metric">
                <span class="metric-label">Current Step:</span>
                <span class="metric-value" id="current-step">-</span>
            </div>
            <div class="metric">
                <span class="metric-label">Cycle Number:</span>
                <span class="metric-value" id="cycle-number">-</span>
            </div>
        </div>
        
        <div class="phase-grid">
            <div class="card">
                <h2>Phase 0</h2>
                <div class="phase-indicator" id="phase-0-indicator">Idle</div>
                <div class="metric">
                    <span class="metric-label">Steps Processed:</span>
                    <span class="metric-value" id="phase-0-steps">0</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Expressive:</span>
                    <span class="metric-value" id="phase-0-expressive">0</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Reflective:</span>
                    <span class="metric-value" id="phase-0-reflective">0</span>
                </div>
            </div>
            
            <div class="card">
                <h2>Phase 1</h2>
                <div class="phase-indicator" id="phase-1-indicator">Idle</div>
                <div class="metric">
                    <span class="metric-label">Steps Processed:</span>
                    <span class="metric-value" id="phase-1-steps">0</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Expressive:</span>
                    <span class="metric-value" id="phase-1-expressive">0</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Reflective:</span>
                    <span class="metric-value" id="phase-1-reflective">0</span>
                </div>
            </div>
            
            <div class="card">
                <h2>Phase 2</h2>
                <div class="phase-indicator" id="phase-2-indicator">Idle</div>
                <div class="metric">
                    <span class="metric-label">Steps Processed:</span>
                    <span class="metric-value" id="phase-2-steps">0</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Expressive:</span>
                    <span class="metric-value" id="phase-2-expressive">0</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Reflective:</span>
                    <span class="metric-value" id="phase-2-reflective">0</span>
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
                    <span class="metric-label">Total Steps:</span>
                    <span class="metric-value" id="total-steps">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Cognitive Load:</span>
                    <span class="metric-value" id="cognitive-load">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Stream Coherence:</span>
                    <span class="metric-value" id="coherence">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Uptime:</span>
                    <span class="metric-value" id="uptime">-</span>
                </div>
            </div>
            
            <div class="card">
                <h2>Active Couplings</h2>
                <div id="couplings-container">
                    <div style="color: #888;">No active couplings</div>
                </div>
            </div>
        </div>
        
        <div class="card" style="text-align: center;">
            <button class="button" onclick="refreshStatus()">🔄 Refresh</button>
            <button class="button" onclick="stopSystem()">⏹️ Stop System</button>
        </div>
    </div>
    
    <script>
        const stepConfigs = [
            {step: 0, phase: 0, term: 'T4', mode: 'E'},
            {step: 1, phase: 1, term: 'T1', mode: 'R'},
            {step: 2, phase: 2, term: 'T2', mode: 'E'},
            {step: 3, phase: 0, term: 'T7', mode: 'R'},
            {step: 4, phase: 1, term: 'T4', mode: 'E'},
            {step: 5, phase: 2, term: 'T1', mode: 'R'},
            {step: 6, phase: 0, term: 'T2', mode: 'E'},
            {step: 7, phase: 1, term: 'T5', mode: 'E'},
            {step: 8, phase: 2, term: 'T8', mode: 'E'},
            {step: 9, phase: 0, term: 'T8', mode: 'E'},
            {step: 10, phase: 1, term: 'T7', mode: 'R'},
            {step: 11, phase: 2, term: 'T5', mode: 'E'}
        ];
        
        function initStepCycle() {
            const container = document.getElementById('step-cycle');
            stepConfigs.forEach(config => {
                const step = document.createElement('div');
                step.className = 'step ' + (config.mode === 'E' ? 'step-expressive' : 'step-reflective');
                step.id = 'step-' + config.step;
                step.innerHTML = '<div>Step ' + config.step + '</div><div>' + config.term + config.mode + '</div>';
                container.appendChild(step);
            });
        }
        
        function updateStatus() {
            fetch('/api/status')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('running').textContent = data.running ? '✅ Active' : '❌ Stopped';
                    document.getElementById('current-step').textContent = data.current_step;
                    document.getElementById('cycle-number').textContent = data.cycle_number;
                    document.getElementById('total-steps').textContent = data.total_steps;
                    document.getElementById('cognitive-load').textContent = data.cognitive_load.toFixed(3);
                    document.getElementById('coherence').textContent = data.stream_coherence.toFixed(3);
                    document.getElementById('uptime').textContent = data.uptime;
                    
                    // Update step cycle
                    for (let i = 0; i < 12; i++) {
                        const stepEl = document.getElementById('step-' + i);
                        if (i === data.current_step) {
                            stepEl.classList.add('step-current');
                        } else {
                            stepEl.classList.remove('step-current');
                        }
                    }
                    
                    // Update phases
                    data.phases.forEach((phase, idx) => {
                        document.getElementById('phase-' + idx + '-steps').textContent = phase.steps_processed;
                        document.getElementById('phase-' + idx + '-expressive').textContent = phase.expressive_steps;
                        document.getElementById('phase-' + idx + '-reflective').textContent = phase.reflective_steps;
                        
                        // Update phase indicator
                        const indicator = document.getElementById('phase-' + idx + '-indicator');
                        const config = stepConfigs[data.current_step];
                        if (config.phase === idx) {
                            indicator.textContent = config.term + config.mode + ' Active';
                            indicator.classList.add('phase-active');
                        } else {
                            indicator.textContent = 'Idle';
                            indicator.classList.remove('phase-active');
                        }
                    });
                    
                    // Update couplings
                    const couplingsContainer = document.getElementById('couplings-container');
                    if (data.active_couplings > 0) {
                        // In a real implementation, we'd show coupling details
                        couplingsContainer.innerHTML = '<div class="coupling-indicator">🔗 ' + data.active_couplings + ' coupling(s) active</div>';
                    } else {
                        couplingsContainer.innerHTML = '<div style="color: #888;">No active couplings</div>';
                    }
                })
                .catch(error => console.error('Error fetching status:', error));
        }
        
        function refreshStatus() {
            updateStatus();
        }
        
        function stopSystem() {
            if (confirm('Stop the 3-phase system?')) {
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
        initStepCycle();
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
	status := threePhaseManager.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := threePhaseManager.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	threePhaseManager.Stop()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "3-phase system stopped",
	})
}
