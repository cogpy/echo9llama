package live2d

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HTTPBridge provides HTTP API for Unreal Engine to communicate with Live2D backend
type HTTPBridge struct {
	mu                  sync.RWMutex
	avatarManager       *AvatarManager
	environmentCoupler  *CognitiveEnvironmentCoupler
	storytellingEngine  *EnvironmentalStorytellingEngine
	assetPlacement      *ProceduralAssetPlacement
	server              *http.Server
	port                int
}

// NewHTTPBridge creates a new HTTP bridge
func NewHTTPBridge(port int) *HTTPBridge {
	return &HTTPBridge{
		avatarManager:      NewAvatarManager("DeepTreeEcho", "/models/default"),
		environmentCoupler: NewCognitiveEnvironmentCoupler(0.8),
		storytellingEngine: NewEnvironmentalStorytellingEngine(),
		assetPlacement:     NewProceduralAssetPlacement(),
		port:               port,
	}
}

// Start starts the HTTP server
func (hb *HTTPBridge) Start() error {
	// Start avatar manager
	if err := hb.avatarManager.Start(); err != nil {
		return fmt.Errorf("failed to start avatar manager: %w", err)
	}
	
	// Start environment coupler
	hb.environmentCoupler.Start()
	
	// Setup environment callback
	hb.environmentCoupler.SetEnvironmentCallback(func(envState EnvironmentState) {
		// Environment state updated
	})
	
	// Setup HTTP routes
	mux := http.NewServeMux()
	
	// Avatar endpoints
	mux.HandleFunc("/api/avatar/state", hb.handleGetAvatarState)
	mux.HandleFunc("/api/avatar/emotional", hb.handleSetEmotionalState)
	mux.HandleFunc("/api/avatar/cognitive", hb.handleSetCognitiveState)
	mux.HandleFunc("/api/avatar/preset", hb.handleSetPreset)
	mux.HandleFunc("/api/avatar/parameters", hb.handleGetParameters)
	
	// Environment endpoints
	mux.HandleFunc("/api/environment/state", hb.handleGetEnvironmentState)
	mux.HandleFunc("/api/environment/narrative", hb.handleGetNarrative)
	mux.HandleFunc("/api/environment/assets", hb.handleGetAssetPositions)
	
	// Chaos engine endpoints
	mux.HandleFunc("/api/chaos/intensity", hb.handleSetChaosIntensity)
	
	// Health check
	mux.HandleFunc("/health", hb.handleHealth)
	
	// Create server
	hb.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", hb.port),
		Handler:      hb.corsMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	
	// Start server
	go func() {
		if err := hb.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()
	
	fmt.Printf("HTTP Bridge started on port %d\n", hb.port)
	return nil
}

// Stop stops the HTTP server
func (hb *HTTPBridge) Stop() error {
	hb.avatarManager.Stop()
	hb.environmentCoupler.Stop()
	
	if hb.server != nil {
		return hb.server.Close()
	}
	return nil
}

// corsMiddleware adds CORS headers
func (hb *HTTPBridge) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// handleGetAvatarState returns current avatar state
func (hb *HTTPBridge) handleGetAvatarState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	state := hb.avatarManager.GetCurrentState()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

// handleSetEmotionalState updates emotional state
func (hb *HTTPBridge) handleSetEmotionalState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var state EmotionalState
	if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := hb.avatarManager.UpdateEmotionalState(state); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Update environment
	avatarState := hb.avatarManager.GetCurrentState()
	hb.environmentCoupler.UpdateAvatarState(avatarState)
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleSetCognitiveState updates cognitive state
func (hb *HTTPBridge) handleSetCognitiveState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var state CognitiveState
	if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := hb.avatarManager.UpdateCognitiveState(state); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Update environment
	avatarState := hb.avatarManager.GetCurrentState()
	hb.environmentCoupler.UpdateAvatarState(avatarState)
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleSetPreset sets emotion preset
func (hb *HTTPBridge) handleSetPreset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		Preset string `json:"preset"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Check if preset exists in standard or super-hot-girl presets
	var state EmotionalState
	var found bool
	
	if s, ok := EmotionPresets[req.Preset]; ok {
		state = s
		found = true
	} else if s, ok := SuperHotGirlPresets[req.Preset]; ok {
		state = s
		found = true
	}
	
	if !found {
		http.Error(w, "Preset not found", http.StatusNotFound)
		return
	}
	
	if err := hb.avatarManager.UpdateEmotionalState(state); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Update environment
	avatarState := hb.avatarManager.GetCurrentState()
	hb.environmentCoupler.UpdateAvatarState(avatarState)
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "preset": req.Preset})
}

// handleGetParameters returns current parameter values
func (hb *HTTPBridge) handleGetParameters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	params, err := hb.avatarManager.GetCurrentParameters()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(params)
}

// handleGetEnvironmentState returns current environment state
func (hb *HTTPBridge) handleGetEnvironmentState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	envState := hb.environmentCoupler.GetEnvironmentState()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(envState)
}

// handleGetNarrative returns environmental narrative
func (hb *HTTPBridge) handleGetNarrative(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	avatarState := hb.avatarManager.GetCurrentState()
	envState := hb.environmentCoupler.GetEnvironmentState()
	
	narrative := hb.storytellingEngine.GenerateNarrative(envState, avatarState)
	
	response := map[string]interface{}{
		"narrative": narrative,
		"history":   hb.storytellingEngine.GetNarrativeHistory(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetAssetPositions returns procedural asset positions
func (hb *HTTPBridge) handleGetAssetPositions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	avatarState := hb.avatarManager.GetCurrentState()
	positions := hb.assetPlacement.UpdateAssetPlacement(avatarState)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(positions)
}

// handleSetChaosIntensity sets chaos engine intensity
func (hb *HTTPBridge) handleSetChaosIntensity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		Intensity float64 `json:"intensity"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// This would need access to the chaos mapper
	// For now, just acknowledge
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"intensity": req.Intensity,
	})
}

// handleHealth returns health status
func (hb *HTTPBridge) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":           "healthy",
		"avatar_running":   hb.avatarManager.IsRunning(),
		"environment_active": true,
		"timestamp":        time.Now().Unix(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
