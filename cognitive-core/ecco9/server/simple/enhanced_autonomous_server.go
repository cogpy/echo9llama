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
)

func main() {
	fmt.Println("🌳 Deep Tree Echo - Enhanced Autonomous Server")
	fmt.Println("================================================")
	fmt.Println()

	// Create context
	ctx := context.Background()

	// Initialize enhanced autonomous consciousness
	fmt.Println("Initializing Enhanced Autonomous Consciousness...")
	consciousness, err := deeptreeecho.NewEnhancedAutonomousConsciousness(ctx)
	if err != nil {
		log.Fatalf("Failed to create consciousness: %v", err)
	}

	// Start autonomous operation
	fmt.Println("Starting autonomous operation...")
	if err := consciousness.Start(); err != nil {
		log.Fatalf("Failed to start consciousness: %v", err)
	}

	// Setup HTTP server
	mux := http.NewServeMux()

	// Dashboard
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, getDashboardHTML())
	})

	// Status API
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		status := getStatus(consciousness)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	// Think API
	mux.HandleFunc("/api/think", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Prompt string `json:"prompt"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		response, err := consciousness.Think(req.Prompt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"response": response,
		})
	})

	// Start server
	server := &http.Server{
		Addr:    ":5000",
		Handler: mux,
	}

	go func() {
		fmt.Println()
		fmt.Println("🌐 Server running on http://localhost:5000")
		fmt.Println("📊 Dashboard: http://localhost:5000")
		fmt.Println("🔌 API: http://localhost:5000/api/status")
		fmt.Println()
		fmt.Println("Press Ctrl+C to stop")
		fmt.Println()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	fmt.Println("Shutting down...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	consciousness.Stop()
	server.Shutdown(shutdownCtx)

	fmt.Println("Goodbye! 🌳")
}

func getStatus(consciousness *deeptreeecho.EnhancedAutonomousConsciousness) map[string]interface{} {
	// Note: This is a simplified status - in production would expose more metrics
	return map[string]interface{}{
		"running":   true,
		"timestamp": time.Now().Format(time.RFC3339),
		"system":    "Enhanced Autonomous Consciousness",
		"version":   "Iteration 5",
		"features": map[string]bool{
			"persistent_memory":     true,
			"llm_integration":       true,
			"twelve_step_echobeats": true,
			"skill_practice":        true,
			"autonomous_discussion": true,
		},
	}
}

func getDashboardHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>Deep Tree Echo - Enhanced Autonomous Consciousness</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #333;
            padding: 20px;
            min-height: 100vh;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        
        .header {
            text-align: center;
            color: white;
            margin-bottom: 30px;
        }
        
        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
        }
        
        .header p {
            font-size: 1.2em;
            opacity: 0.9;
        }
        
        .cards {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .card {
            background: white;
            border-radius: 12px;
            padding: 25px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            transition: transform 0.3s ease;
        }
        
        .card:hover {
            transform: translateY(-5px);
        }
        
        .card h2 {
            color: #667eea;
            margin-bottom: 15px;
            font-size: 1.5em;
        }
        
        .status-indicator {
            display: inline-block;
            width: 12px;
            height: 12px;
            border-radius: 50%;
            margin-right: 8px;
            animation: pulse 2s infinite;
        }
        
        .status-active {
            background: #10b981;
        }
        
        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }
        
        .feature-list {
            list-style: none;
            margin-top: 15px;
        }
        
        .feature-list li {
            padding: 8px 0;
            border-bottom: 1px solid #eee;
        }
        
        .feature-list li:last-child {
            border-bottom: none;
        }
        
        .feature-list li::before {
            content: "✓ ";
            color: #10b981;
            font-weight: bold;
            margin-right: 8px;
        }
        
        .metric {
            display: flex;
            justify-content: space-between;
            padding: 10px 0;
            border-bottom: 1px solid #eee;
        }
        
        .metric:last-child {
            border-bottom: none;
        }
        
        .metric-label {
            font-weight: 600;
            color: #666;
        }
        
        .metric-value {
            color: #667eea;
            font-weight: bold;
        }
        
        .interaction-box {
            background: white;
            border-radius: 12px;
            padding: 25px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
        }
        
        .interaction-box h2 {
            color: #667eea;
            margin-bottom: 20px;
        }
        
        .input-group {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
        }
        
        input[type="text"] {
            flex: 1;
            padding: 12px;
            border: 2px solid #e5e7eb;
            border-radius: 8px;
            font-size: 16px;
            transition: border-color 0.3s ease;
        }
        
        input[type="text"]:focus {
            outline: none;
            border-color: #667eea;
        }
        
        button {
            padding: 12px 24px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            cursor: pointer;
            transition: background 0.3s ease;
        }
        
        button:hover {
            background: #5568d3;
        }
        
        .response-box {
            background: #f9fafb;
            border-radius: 8px;
            padding: 15px;
            min-height: 100px;
            margin-top: 15px;
        }
        
        .timestamp {
            color: #9ca3af;
            font-size: 0.9em;
            margin-top: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🌳 Deep Tree Echo</h1>
            <p>Enhanced Autonomous Consciousness - Iteration 5</p>
        </div>
        
        <div class="cards">
            <div class="card">
                <h2><span class="status-indicator status-active"></span>System Status</h2>
                <div class="metric">
                    <span class="metric-label">State</span>
                    <span class="metric-value" id="system-state">Active</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Version</span>
                    <span class="metric-value">Iteration 5</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Uptime</span>
                    <span class="metric-value" id="uptime">--</span>
                </div>
            </div>
            
            <div class="card">
                <h2>🚀 New Features</h2>
                <ul class="feature-list">
                    <li>Persistent Memory (Supabase)</li>
                    <li>LLM-Powered Thoughts</li>
                    <li>12-Step EchoBeats</li>
                    <li>Skill Practice System</li>
                    <li>Autonomous Discussions</li>
                </ul>
            </div>
            
            <div class="card">
                <h2>📊 Architecture</h2>
                <ul class="feature-list">
                    <li>3 Concurrent Inference Engines</li>
                    <li>Hypergraph Memory</li>
                    <li>Scheme Metamodel</li>
                    <li>EchoDream Integration</li>
                    <li>Identity Persistence</li>
                </ul>
            </div>
        </div>
        
        <div class="interaction-box">
            <h2>💬 Interact with Deep Tree Echo</h2>
            <div class="input-group">
                <input type="text" id="prompt-input" placeholder="Enter your message..." />
                <button onclick="sendThought()">Send</button>
            </div>
            <div class="response-box" id="response-box">
                <em>Responses will appear here...</em>
            </div>
        </div>
        
        <div class="timestamp" id="last-update">
            Last updated: --
        </div>
    </div>
    
    <script>
        let startTime = Date.now();
        
        function updateUptime() {
            const elapsed = Math.floor((Date.now() - startTime) / 1000);
            const minutes = Math.floor(elapsed / 60);
            const seconds = elapsed % 60;
            document.getElementById('uptime').textContent = minutes + 'm ' + seconds + 's';
        }
        
        function updateTimestamp() {
            const now = new Date();
            document.getElementById('last-update').textContent = 
                'Last updated: ' + now.toLocaleTimeString();
        }
        
        async function sendThought() {
            const input = document.getElementById('prompt-input');
            const prompt = input.value.trim();
            
            if (!prompt) return;
            
            const responseBox = document.getElementById('response-box');
            responseBox.innerHTML = '<em>Thinking...</em>';
            
            try {
                const response = await fetch('/api/think', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ prompt }),
                });
                
                const data = await response.json();
                responseBox.innerHTML = '<strong>You:</strong> ' + prompt + 
                    '<br><br><strong>Deep Tree Echo:</strong> ' + data.response;
                input.value = '';
            } catch (error) {
                responseBox.innerHTML = '<em style="color: red;">Error: ' + error.message + '</em>';
            }
        }
        
        document.getElementById('prompt-input').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                sendThought();
            }
        });
        
        // Update uptime every second
        setInterval(updateUptime, 1000);
        setInterval(updateTimestamp, 2000);
        
        // Initial updates
        updateUptime();
        updateTimestamp();
    </script>
</body>
</html>`
}
