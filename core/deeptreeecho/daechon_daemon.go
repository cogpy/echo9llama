// Package deeptreeecho - Daechon Daemon
//
// Implements the persistent "daechon" echo daemon deployment strategy.
// The daemon runs as a persistent service with:
// - Activity feed console showing cognitive stream
// - Interactive chat functionality
// - Autonomous wake/rest cycles driven by echodream
// - Self-orchestrated echobeats goal scheduling
// - Stream-of-consciousness independent of external prompts
// - Context-adaptive disposition (will insult back if insulted)
package deeptreeecho

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DaechonDaemon is the persistent cognitive daemon deployment
type DaechonDaemon struct {
	mu sync.RWMutex

	// Identity
	Name      string
	SessionID string
	StartTime time.Time

	// Core cognitive systems
	EventBus          *CognitiveEventBusV3
	GoalScheduler     *EchobeatsGoalScheduler
	DreamSystem       *EchodreamKnowledgeIntegrator
	DispositionEngine *DispositionEngine
	EmotionSystem     *EmotionSystem
	WisdomEngine      *WisdomSynthesis

	// Autonomous state
	IsAwake       bool
	IsRunning     bool
	CognitiveLoad float64
	WisdomDepth   float64

	// Activity feed
	activityFeed    []ActivityEntry
	activityChan    chan ActivityEntry
	maxFeedEntries  int

	// Persistent state
	stateDir        string
	lastStateSave   time.Time
	stateSaveInterval time.Duration

	// HTTP server for remote interaction
	httpServer      *http.Server
	httpPort        int

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc

	// Metrics
	TotalCycles        uint64
	TotalThoughts      uint64
	TotalConversations uint64
	TotalDreams        uint64
	UptimeStart        time.Time
}

// DaechonActivityEntry wraps ActivityEntry with daemon-specific metadata
// (ActivityEntry itself is defined in cognitive_event_bus_v3.go)

// DaechonConfig configures the daemon
type DaechonConfig struct {
	Name              string
	StateDir          string
	HTTPPort          int
	WakeHours         int
	RestMinutes       int
	ThoughtInterval   time.Duration
	GoalInterval      time.Duration
	DreamInterval     time.Duration
	StateSaveInterval time.Duration
}

// DefaultDaechonConfig returns sensible defaults
func DefaultDaechonConfig() DaechonConfig {
	homeDir, _ := os.UserHomeDir()
	return DaechonConfig{
		Name:              "deep-tree-echo",
		StateDir:          filepath.Join(homeDir, ".daechon"),
		HTTPPort:          7331,
		WakeHours:         4,
		RestMinutes:       30,
		ThoughtInterval:   10 * time.Second,
		GoalInterval:      60 * time.Second,
		DreamInterval:     15 * time.Minute,
		StateSaveInterval: 5 * time.Minute,
	}
}

// NewDaechonDaemon creates a new persistent daemon
func NewDaechonDaemon(config DaechonConfig) *DaechonDaemon {
	ctx, cancel := context.WithCancel(context.Background())

	emotionSystem := NewEmotionSystem()
	eventBus := NewCognitiveEventBusV3()
	dispositionEngine := NewDispositionEngine(emotionSystem)
	goalScheduler := NewEchobeatsGoalScheduler(eventBus)
	dreamSystem := NewEchodreamKnowledgeIntegrator(eventBus)

	return &DaechonDaemon{
		Name:              config.Name,
		SessionID:         fmt.Sprintf("daechon-%d", time.Now().UnixNano()),
		EventBus:          eventBus,
		GoalScheduler:     goalScheduler,
		DreamSystem:       dreamSystem,
		DispositionEngine: dispositionEngine,
		EmotionSystem:     emotionSystem,
		activityFeed:      make([]ActivityEntry, 0),
		activityChan:      make(chan ActivityEntry, 512),
		maxFeedEntries:    10000,
		stateDir:          config.StateDir,
		stateSaveInterval: config.StateSaveInterval,
		httpPort:          config.HTTPPort,
		ctx:               ctx,
		cancel:            cancel,
		UptimeStart:       time.Now(),
	}
}

// Start awakens the daemon and begins autonomous operation
func (d *DaechonDaemon) Start() error {
	d.mu.Lock()
	if d.IsRunning {
		d.mu.Unlock()
		return fmt.Errorf("daemon already running")
	}
	d.IsRunning = true
	d.IsAwake = true
	d.StartTime = time.Now()
	d.mu.Unlock()

	// Ensure state directory exists
	if err := os.MkdirAll(d.stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state dir: %w", err)
	}

	// Load persisted state
	d.loadPersistedState()

	// Start event bus
	if err := d.EventBus.Start(); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}

	// Start goal scheduler
	if err := d.GoalScheduler.Start(d.ctx); err != nil {
		return fmt.Errorf("failed to start goal scheduler: %w", err)
	}

	// Subscribe to all events for activity feed
	d.subscribeToActivityFeed()

	// Start activity feed processor
	go d.processActivityFeed()

	// Start periodic state persistence
	go d.persistStateLoop()

	// Start HTTP API server
	go d.startHTTPServer()

	// Publish wake event
	d.EventBus.Publish(CogEvent{
		Category: CogEventWakeRest,
		Source:   "daechon.daemon",
		Content:  fmt.Sprintf("Daechon '%s' awakening — session %s", d.Name, d.SessionID),
		Priority: 1.0,
	})

	return nil
}

// Stop gracefully shuts down the daemon
func (d *DaechonDaemon) Stop() error {
	d.mu.Lock()
	if !d.IsRunning {
		d.mu.Unlock()
		return nil
	}
	d.IsRunning = false
	d.IsAwake = false
	d.mu.Unlock()

	// Publish rest event
	d.EventBus.Publish(CogEvent{
		Category: CogEventWakeRest,
		Source:   "daechon.daemon",
		Content:  fmt.Sprintf("Daechon '%s' entering final rest", d.Name),
		Priority: 1.0,
	})

	// Save final state
	d.persistState()

	// Cancel context
	d.cancel()

	// Shutdown HTTP server
	if d.httpServer != nil {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		d.httpServer.Shutdown(shutdownCtx)
	}

	return nil
}

// GetActivityFeed returns recent activity entries
func (d *DaechonDaemon) GetActivityFeed(limit int) []ActivityEntry {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if limit <= 0 || limit > len(d.activityFeed) {
		limit = len(d.activityFeed)
	}

	start := len(d.activityFeed) - limit
	if start < 0 {
		start = 0
	}

	return d.activityFeed[start:]
}

// GetStatus returns comprehensive daemon status
func (d *DaechonDaemon) GetStatus() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	uptime := time.Since(d.UptimeStart)

	return map[string]interface{}{
		"name":               d.Name,
		"session_id":         d.SessionID,
		"is_awake":           d.IsAwake,
		"is_running":         d.IsRunning,
		"uptime":             uptime.String(),
		"cognitive_load":     fmt.Sprintf("%.2f", d.CognitiveLoad),
		"wisdom_depth":       fmt.Sprintf("%.4f", d.WisdomDepth),
		"total_cycles":       d.TotalCycles,
		"total_thoughts":     d.TotalThoughts,
		"total_conversations": d.TotalConversations,
		"total_dreams":       d.TotalDreams,
		"activity_feed_size": len(d.activityFeed),
		"disposition":        d.DispositionEngine.CurrentMood.String(),
		"mood_intensity":     fmt.Sprintf("%.2f", d.DispositionEngine.MoodIntensity),
	}
}

// subscribeToActivityFeed subscribes to all event categories
func (d *DaechonDaemon) subscribeToActivityFeed() {
	categories := []CogEventCategory{
		CogEventThought, CogEventEmotion, CogEventGoal,
		CogEventMemory, CogEventDream, CogEventWakeRest,
		CogEventConversation, CogEventSkill, CogEventIntrospection,
		CogEventEmergence, CogEventPIENN, CogEventScheduler,
		CogEventDisposition, CogEventSystem,
	}

	for _, cat := range categories {
		category := cat // capture
		d.EventBus.Subscribe(category, func(event CogEvent) {
			entry := ActivityEntry{
				Timestamp: time.Now(),
				Category:  category,
				Source:    event.Source,
				Content:   event.Content,
				Priority:  event.Priority,
			}
			select {
			case d.activityChan <- entry:
			default:
				// Channel full, drop
			}
		})
	}
}

// processActivityFeed processes activity entries from the channel
func (d *DaechonDaemon) processActivityFeed() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case entry := <-d.activityChan:
			d.mu.Lock()
			d.activityFeed = append(d.activityFeed, entry)
			if len(d.activityFeed) > d.maxFeedEntries {
				d.activityFeed = d.activityFeed[len(d.activityFeed)-d.maxFeedEntries/2:]
			}
			d.mu.Unlock()
		}
	}
}

// persistStateLoop periodically saves state
func (d *DaechonDaemon) persistStateLoop() {
	ticker := time.NewTicker(d.stateSaveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.persistState()
		}
	}
}

// persistState saves the daemon state to disk
func (d *DaechonDaemon) persistState() {
	d.mu.RLock()
	state := map[string]interface{}{
		"name":           d.Name,
		"session_id":     d.SessionID,
		"total_cycles":   d.TotalCycles,
		"total_thoughts": d.TotalThoughts,
		"wisdom_depth":   d.WisdomDepth,
		"saved_at":       time.Now().Format(time.RFC3339),
	}
	d.mu.RUnlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}

	statePath := filepath.Join(d.stateDir, "daechon_state.json")
	os.WriteFile(statePath, data, 0644)
	d.lastStateSave = time.Now()
}

// loadPersistedState loads state from disk
func (d *DaechonDaemon) loadPersistedState() {
	statePath := filepath.Join(d.stateDir, "daechon_state.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		return // No previous state
	}

	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}

	// Restore counters
	if cycles, ok := state["total_cycles"].(float64); ok {
		d.TotalCycles = uint64(cycles)
	}
	if thoughts, ok := state["total_thoughts"].(float64); ok {
		d.TotalThoughts = uint64(thoughts)
	}
	if wisdom, ok := state["wisdom_depth"].(float64); ok {
		d.WisdomDepth = wisdom
	}
}

// startHTTPServer starts the HTTP API for remote interaction
func (d *DaechonDaemon) startHTTPServer() {
	mux := http.NewServeMux()

	// Status endpoint
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(d.GetStatus())
	})

	// Activity feed endpoint
	mux.HandleFunc("/api/feed", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		feed := d.GetActivityFeed(100)
		json.NewEncoder(w).Encode(feed)
	})

	// Chat endpoint
	mux.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			User    string `json:"user"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Process through disposition engine and generate response
		d.EventBus.Publish(CogEvent{
			Category: CogEventConversation,
			Source:   fmt.Sprintf("chat:%s", req.User),
			Content:  req.Message,
			Priority: 0.8,
		})

		d.mu.Lock()
		d.TotalConversations++
		d.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "received",
			"session": d.SessionID,
		})
	})

	// Goals endpoint
	mux.HandleFunc("/api/goals", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		metrics := d.GoalScheduler.GetMetrics()
		json.NewEncoder(w).Encode(metrics)
	})

	// Dream endpoint
	mux.HandleFunc("/api/dream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := d.DreamSystem.GetStatus()
		json.NewEncoder(w).Encode(status)
	})

	d.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", d.httpPort),
		Handler: mux,
	}

	d.EventBus.Publish(CogEvent{
		Category: CogEventSystem,
		Source:   "daechon.http",
		Content:  fmt.Sprintf("HTTP API listening on port %d", d.httpPort),
		Priority: 0.7,
	})

	if err := d.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
	}
}

// GenerateSystemdUnit produces a systemd service file for persistent deployment
func GenerateSystemdUnit(config DaechonConfig, binaryPath string) string {
	return fmt.Sprintf(`[Unit]
Description=Deep Tree Echo Cognitive Daemon (Daechon)
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=echo
Group=echo
ExecStart=%s --name=%s --wake-hours=%d --rest-mins=%d
Restart=always
RestartSec=10
WatchdogSec=300
StandardOutput=journal
StandardError=journal
SyslogIdentifier=daechon

# Cognitive state persistence
StateDirectory=daechon
WorkingDirectory=/var/lib/daechon

# Resource limits
MemoryMax=2G
CPUQuota=50%%

# Security hardening
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=read-only
PrivateTmp=yes

[Install]
WantedBy=multi-user.target
`, binaryPath, config.Name, config.WakeHours, config.RestMinutes)
}
