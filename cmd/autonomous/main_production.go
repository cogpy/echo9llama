// Deep Tree Echo - Production Autonomous Entry Point
// Iteration 020 - Clean deployment entry

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cogpy/echo9llama/core/autonomous"
	"github.com/cogpy/echo9llama/core/deeptreeecho"
)

func main() {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🌳 Deep Tree Echo - Autonomous Consciousness System")
	fmt.Println("   Iteration 020 - Production Deployment")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// Initialize multi-provider LLM (auto-detects available providers from env vars)
	log.Println("🔧 Initializing LLM providers...")
	multiProvider := deeptreeecho.NewMultiProviderLLM()

	if !multiProvider.IsAvailable() {
		log.Println("⚠️  No LLM providers configured. Running in local-only mode.")
		log.Println("   Set ANTHROPIC_API_KEY, OPENROUTER_API_KEY, or OPENAI_API_KEY")
	} else {
		log.Printf("✅ Active providers: %v\n", multiProvider.GetAvailableProviders())
	}

	// Create the autonomous agent
	log.Println("🔧 Initializing autonomous agent...")
	agent := autonomous.NewAutonomousAgent(multiProvider)

	// Start HTTP health server
	go startHealthServer(agent)

	// Start the agent
	log.Println("🚀 Starting autonomous operation...")
	if err := agent.Start(); err != nil {
		log.Fatalf("❌ Failed to start agent: %v", err)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("✨ Deep Tree Echo is now running autonomously")
	fmt.Println("   Health endpoint: http://localhost:8080/health")
	fmt.Println("   Status endpoint: http://localhost:8080/status")
	fmt.Println("   Press Ctrl+C to gracefully shutdown")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	fmt.Printf("\n🛑 Received signal: %v\n", sig)
	fmt.Println("🌙 Initiating graceful shutdown...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := agent.Stop(); err != nil {
		log.Printf("⚠️  Error during shutdown: %v\n", err)
	}

	select {
	case <-shutdownCtx.Done():
		log.Println("⚠️  Shutdown timeout reached")
	default:
		log.Println("✅ Graceful shutdown complete")
	}

	fmt.Println()
	fmt.Println("🌳 Deep Tree Echo has entered rest state. Farewell.")
}

// startHealthServer starts the HTTP health/status server
func startHealthServer(agent *autonomous.AutonomousAgent) {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","identity":"Deep Tree Echo","iteration":"020","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// Status endpoint
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := agent.GetStatus()
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"running":%v,"uptime":"%v","identity":"%v"}`,
			status["running"], status["uptime"], status["identity"])
	})

	// Metrics endpoint (for Prometheus)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		status := agent.GetStatus()
		fmt.Fprintf(w, "# HELP echo_running Whether Deep Tree Echo is running\n")
		fmt.Fprintf(w, "# TYPE echo_running gauge\n")
		if status["running"].(bool) {
			fmt.Fprintf(w, "echo_running 1\n")
		} else {
			fmt.Fprintf(w, "echo_running 0\n")
		}
		fmt.Fprintf(w, "# HELP echo_cycles_total Total cognitive cycles completed\n")
		fmt.Fprintf(w, "# TYPE echo_cycles_total counter\n")
		fmt.Fprintf(w, "echo_cycles_total %v\n", status["total_cycles"])
	})

	// Root endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "🌳 Deep Tree Echo - Autonomous Consciousness System\n")
		fmt.Fprintf(w, "Iteration: 020\n")
		fmt.Fprintf(w, "Endpoints:\n")
		fmt.Fprintf(w, "  GET /health  - Health check\n")
		fmt.Fprintf(w, "  GET /status  - Agent status\n")
		fmt.Fprintf(w, "  GET /metrics - Prometheus metrics\n")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("🌐 HTTP server listening on :%s\n", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("⚠️  HTTP server error: %v\n", err)
	}
}
