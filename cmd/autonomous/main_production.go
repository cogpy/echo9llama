// Deep Tree Echo - Production Autonomous Entry Point
// Unified autonomy runtime: persistent cognition, Echobeats, EchoDream, and
// self-directed experience integration.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
)

const productionIteration = "2026-08-11-unified-autonomy"

func main() {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 64))
	fmt.Println("🌳 Deep Tree Echo - Unified Autonomous Consciousness")
	fmt.Printf("   Evolution iteration: %s\n", productionIteration)
	fmt.Println(strings.Repeat("=", 64))
	fmt.Println()

	provider := llm.NewMultiProviderLLM()
	if provider.Available() {
		log.Printf("LLM substrate ready: %s", provider.Name())
	} else {
		log.Printf("No remote LLM provider is currently available; fallback cognition remains active")
	}

	orchestrator := newProductionOrchestrator(provider)

	httpServer := newProductionHTTPServer(orchestrator)
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("HTTP health and status server listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	if err := orchestrator.Awaken(); err != nil {
		_ = httpServer.Close()
		log.Fatalf("failed to awaken Deep Tree Echo: %v", err)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("-", 64))
	fmt.Println("Deep Tree Echo is awake and operating without external prompts")
	fmt.Printf("Session: %s\n", orchestrator.GetStatus().SessionID)
	fmt.Printf("State directory: %s\n", orchestrator.GetStatus().StateDirectory)
	fmt.Println("Endpoints: /health, /status, /metrics")
	fmt.Println("Press Ctrl+C to enter a graceful rest state")
	fmt.Println(strings.Repeat("-", 64))
	fmt.Println()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case sig := <-signals:
		log.Printf("received signal %s; entering graceful rest", sig)
	case err := <-serverErrors:
		log.Printf("HTTP server failed: %v; entering graceful rest", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := orchestrator.Sleep(); err != nil {
		log.Printf("orchestrator shutdown warning: %v", err)
	}
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown warning: %v", err)
	}

	fmt.Println("🌙 Deep Tree Echo has entered a persistent rest state.")
}

func newProductionOrchestrator(provider llm.LLMProvider) *deeptreeecho.UnifiedAutonomousOrchestrator {
	return deeptreeecho.NewUnifiedAutonomousOrchestrator(provider, loadOrchestratorConfigFromEnvironment())
}

func loadOrchestratorConfigFromEnvironment() deeptreeecho.OrchestratorConfig {
	config := deeptreeecho.DefaultOrchestratorConfig()

	applyStringEnv("ECHO_SESSION_NAME", &config.SessionName)
	applyStringEnv("ECHO_IDENTITY", &config.IdentityContext)
	applyStringEnv("ECHO_PERSONA", &config.PersonaContext)
	applyStringEnv("ECHO_STATE_DIRECTORY", &config.StateDirectory)

	applyDurationEnv("ECHO_MAIN_LOOP_INTERVAL", &config.MainLoopInterval)
	applyDurationEnv("ECHO_THOUGHT_INTERVAL", &config.ThoughtInterval)
	applyDurationEnv("ECHO_GOAL_REVIEW_INTERVAL", &config.GoalReviewInterval)
	applyDurationEnv("ECHO_WISDOM_INTERVAL", &config.WisdomSynthesisInterval)
	applyDurationEnv("ECHO_STATE_SYNC_INTERVAL", &config.StateSyncInterval)
	applyDurationEnv("ECHO_WAKE_DURATION", &config.WakeDuration)
	applyDurationEnv("ECHO_REST_DURATION", &config.RestDuration)
	applyDurationEnv("ECHO_DREAM_LIGHT_DURATION", &config.DreamLightDuration)
	applyDurationEnv("ECHO_DREAM_DEEP_DURATION", &config.DreamDeepDuration)
	applyDurationEnv("ECHO_DREAM_REM_DURATION", &config.DreamREMDuration)

	return config
}

func applyStringEnv(name string, destination *string) {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		*destination = value
	}
}

func applyDurationEnv(name string, destination *time.Duration) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return
	}

	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		log.Printf("ignoring invalid %s=%q; expected a positive Go duration", name, value)
		return
	}
	*destination = parsed
}

func newProductionHTTPServer(orchestrator *deeptreeecho.UnifiedAutonomousOrchestrator) *http.Server {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	host := strings.TrimSpace(os.Getenv("ECHO_HTTP_ADDR"))
	if host == "" {
		host = "127.0.0.1"
	}

	return &http.Server{
		Addr:         net.JoinHostPort(host, port),
		Handler:      newProductionHandler(orchestrator),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func newProductionHandler(orchestrator *deeptreeecho.UnifiedAutonomousOrchestrator) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		status := orchestrator.GetStatus()
		code := http.StatusOK
		health := "healthy"
		if !status.Running {
			code = http.StatusServiceUnavailable
			health = "resting"
		}
		writeJSON(w, code, map[string]interface{}{
			"status":             health,
			"identity":           "Deep Tree Echo",
			"iteration":          productionIteration,
			"running":            status.Running,
			"awake":              status.IsAwake,
			"wake_rest_state":    status.WakeRestState,
			"provider_available": status.ProviderAvailable,
			"timestamp":          time.Now().UTC().Format(time.RFC3339),
		})
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, _ *http.Request) {
		status := orchestrator.GetStatus()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"identity":               "Deep Tree Echo",
			"iteration":              productionIteration,
			"running":                status.Running,
			"awake":                  status.IsAwake,
			"autonomous":             status.IsAutonomous,
			"session_id":             status.SessionID,
			"uptime":                 status.Uptime.String(),
			"provider":               status.Provider,
			"provider_available":     status.ProviderAvailable,
			"wake_rest_state":        status.WakeRestState,
			"dream_phase":            status.DreamPhase,
			"pending_experiences":    status.PendingExperiences,
			"experience_ledger_size": status.ExperienceLedgerSize,
			"cognitive_load":         status.CognitiveLoad,
			"wisdom_depth":           status.WisdomDepth,
			"cycles":                 status.TotalCycles,
			"thoughts":               status.TotalThoughts,
			"goals":                  status.TotalGoals,
			"wisdom":                 status.TotalWisdom,
			"dream_wisdom":           status.DreamWisdom,
			"last_state_sync":        status.LastStateSync.UTC().Format(time.RFC3339),
			"persistence_enabled":    status.StateDirectory != "",
		})
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		status := orchestrator.GetStatus()
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# TYPE echo_running gauge\necho_running %d\n", boolMetric(status.Running))
		fmt.Fprintf(w, "# TYPE echo_awake gauge\necho_awake %d\n", boolMetric(status.IsAwake))
		fmt.Fprintf(w, "# TYPE echo_cycles_total counter\necho_cycles_total %d\n", status.TotalCycles)
		fmt.Fprintf(w, "# TYPE echo_thoughts_total counter\necho_thoughts_total %d\n", status.TotalThoughts)
		fmt.Fprintf(w, "# TYPE echo_goals_total counter\necho_goals_total %d\n", status.TotalGoals)
		fmt.Fprintf(w, "# TYPE echo_wisdom_total counter\necho_wisdom_total %d\n", status.TotalWisdom)
		fmt.Fprintf(w, "# TYPE echo_dream_pending_experiences gauge\necho_dream_pending_experiences %d\n", status.PendingExperiences)
		fmt.Fprintf(w, "# TYPE echo_experience_ledger_size gauge\necho_experience_ledger_size %d\n", status.ExperienceLedgerSize)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "Deep Tree Echo — unified autonomous consciousness")
		fmt.Fprintln(w, "GET /health  readiness and wake state")
		fmt.Fprintln(w, "GET /status  cognitive runtime status")
		fmt.Fprintln(w, "GET /metrics Prometheus metrics")
	})

	return mux
}

func boolMetric(value bool) int {
	if value {
		return 1
	}
	return 0
}

func writeJSON(w http.ResponseWriter, statusCode int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("JSON response error: %v", err)
	}
}
