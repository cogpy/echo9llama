// Package void implements the Global Telemetry Shell - the computational void
// that provides context, coordination, and gestalt perception for all echo-consciousness.
package void

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// GlobalTelemetryShell is the void - the substrate for all computation.
// It provides the coordinate system, execution context, and gestalt perception.
type GlobalTelemetryShell struct {
	// Core components
	coordinateSystem *CoordinateSystem
	gestalt          *GestaltState
	orchestra        *Orchestra
	telemetry        *TelemetryCollector

	// Managed entities
	cores    map[string]*LocalCore
	channels map[string]*Channel
	pipes    map[string]*Pipe

	// State
	running    bool
	cycleCount int64
	startTime  time.Time

	// Communication channels
	updateChan  chan *TelemetryUpdate
	gestaltChan chan *GestaltBroadcast
	stopChan    chan struct{}

	// Synchronization
	mu sync.RWMutex
}

// GTSConfig configures the Global Telemetry Shell.
type GTSConfig struct {
	CycleDuration          time.Duration
	StepDuration           time.Duration
	TelemetryRate          time.Duration
	BufferSize             int
	MaxHistorySize         int
	EnableAnomalyDetection bool
}

// DefaultGTSConfig returns a default configuration.
func DefaultGTSConfig() *GTSConfig {
	return &GTSConfig{
		CycleDuration:          1200 * time.Millisecond, // 12 steps × 100ms
		StepDuration:           100 * time.Millisecond,
		TelemetryRate:          50 * time.Millisecond,
		BufferSize:             1000,
		MaxHistorySize:         100,
		EnableAnomalyDetection: true,
	}
}

// NewGlobalTelemetryShell creates a new Global Telemetry Shell.
func NewGlobalTelemetryShell(config *GTSConfig) *GlobalTelemetryShell {
	if config == nil {
		config = DefaultGTSConfig()
	}

	gts := &GlobalTelemetryShell{
		coordinateSystem: NewCoordinateSystem(),
		gestalt:          NewGestaltState(config.MaxHistorySize),
		orchestra:        NewOrchestra(config.StepDuration),
		telemetry:        NewTelemetryCollector(config.TelemetryRate, config.BufferSize),
		cores:            make(map[string]*LocalCore),
		channels:         make(map[string]*Channel),
		pipes:            make(map[string]*Pipe),
		running:          false,
		cycleCount:       0,
		updateChan:       make(chan *TelemetryUpdate, config.BufferSize),
		gestaltChan:      make(chan *GestaltBroadcast, config.BufferSize),
		stopChan:         make(chan struct{}),
	}

	return gts
}

// Start starts the Global Telemetry Shell.
func (gts *GlobalTelemetryShell) Start(ctx context.Context) error {
	gts.mu.Lock()
	if gts.running {
		gts.mu.Unlock()
		return fmt.Errorf("global telemetry shell already running")
	}
	gts.running = true
	gts.startTime = time.Now()
	gts.mu.Unlock()

	// Start telemetry collector
	if err := gts.telemetry.Start(ctx); err != nil {
		return fmt.Errorf("failed to start telemetry collector: %w", err)
	}

	// Start orchestra
	if err := gts.orchestra.Start(ctx); err != nil {
		return fmt.Errorf("failed to start orchestra: %w", err)
	}

	// Start main loop
	go gts.run(ctx)

	return nil
}

// Stop stops the Global Telemetry Shell.
func (gts *GlobalTelemetryShell) Stop() error {
	gts.mu.Lock()
	if !gts.running {
		gts.mu.Unlock()
		return fmt.Errorf("global telemetry shell not running")
	}
	gts.running = false
	gts.mu.Unlock()

	// Signal stop
	close(gts.stopChan)

	// Stop orchestra
	if err := gts.orchestra.Stop(); err != nil {
		return fmt.Errorf("failed to stop orchestra: %w", err)
	}

	// Stop telemetry collector
	if err := gts.telemetry.Stop(); err != nil {
		return fmt.Errorf("failed to stop telemetry collector: %w", err)
	}

	return nil
}

// run is the main loop of the Global Telemetry Shell.
func (gts *GlobalTelemetryShell) run(ctx context.Context) {
	ticker := time.NewTicker(gts.orchestra.stepDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-gts.stopChan:
			return
		case <-ticker.C:
			// Advance to next step
			if err := gts.orchestra.AdvanceStep(); err != nil {
				fmt.Printf("Error advancing step: %v\n", err)
				continue
			}

			// Collect telemetry
			observations := gts.telemetry.CollectAll()

			// Update gestalt
			if err := gts.gestalt.Integrate(observations); err != nil {
				fmt.Printf("Error integrating gestalt: %v\n", err)
				continue
			}

			// Broadcast gestalt at synchronization points
			if gts.isSynchronizationPoint(gts.orchestra.GetCurrentStep()) {
				if err := gts.BroadcastGestalt(); err != nil {
					fmt.Printf("Error broadcasting gestalt: %v\n", err)
				}
			}

			// Increment cycle count
			if gts.orchestra.GetCurrentStep() == 0 {
				gts.mu.Lock()
				gts.cycleCount++
				gts.mu.Unlock()
			}
		}
	}
}

// isSynchronizationPoint returns true if the step is a triadic synchronization point.
func (gts *GlobalTelemetryShell) isSynchronizationPoint(step int) bool {
	// Triads: {1,5,9}, {2,6,10}, {3,7,11}, {4,8,12}
	// Synchronize at the first step of each triad
	return step == 1 || step == 2 || step == 3 || step == 4
}

// RegisterCore registers a local core with the shell.
func (gts *GlobalTelemetryShell) RegisterCore(core *LocalCore) error {
	gts.mu.Lock()
	defer gts.mu.Unlock()

	if _, exists := gts.cores[core.ID]; exists {
		return fmt.Errorf("core %s already registered", core.ID)
	}

	// Derive execution context from coordinate system
	core.ExecutionContext = gts.coordinateSystem.DeriveContext(core.ID)

	// Register with telemetry collector
	observer := &Observer{
		ID:     fmt.Sprintf("observer-%s", core.ID),
		CoreID: core.ID,
		Sampler: func() *TelemetryObservation {
			return core.GetTelemetry()
		},
	}
	if err := gts.telemetry.RegisterObserver(observer); err != nil {
		return fmt.Errorf("failed to register observer: %w", err)
	}

	gts.cores[core.ID] = core
	return nil
}

// RegisterChannel registers a channel with the shell.
func (gts *GlobalTelemetryShell) RegisterChannel(channel *Channel) error {
	gts.mu.Lock()
	defer gts.mu.Unlock()

	if _, exists := gts.channels[channel.ID]; exists {
		return fmt.Errorf("channel %s already registered", channel.ID)
	}

	gts.channels[channel.ID] = channel
	return nil
}

// RegisterPipe registers a pipe with the shell.
func (gts *GlobalTelemetryShell) RegisterPipe(pipe *Pipe) error {
	gts.mu.Lock()
	defer gts.mu.Unlock()

	if _, exists := gts.pipes[pipe.ID]; exists {
		return fmt.Errorf("pipe %s already registered", pipe.ID)
	}

	gts.pipes[pipe.ID] = pipe
	return nil
}

// UpdateGestalt manually triggers a gestalt update.
func (gts *GlobalTelemetryShell) UpdateGestalt() error {
	observations := gts.telemetry.CollectAll()
	return gts.gestalt.Integrate(observations)
}

// BroadcastGestalt broadcasts the current gestalt to all cores.
func (gts *GlobalTelemetryShell) BroadcastGestalt() error {
	gts.mu.RLock()
	cores := make([]*LocalCore, 0, len(gts.cores))
	for _, core := range gts.cores {
		cores = append(cores, core)
	}
	gts.mu.RUnlock()

	// Create gestalt snapshot
	snapshot := gts.gestalt.Snapshot()

	// Broadcast to all cores
	broadcast := &GestaltBroadcast{
		Timestamp: time.Now(),
		Snapshot:  snapshot,
	}

	for _, core := range cores {
		if err := core.ReceiveGestalt(broadcast); err != nil {
			return fmt.Errorf("failed to broadcast to core %s: %w", core.ID, err)
		}
	}

	return nil
}

// GetGestalt returns the current gestalt state.
func (gts *GlobalTelemetryShell) GetGestalt() *GestaltState {
	return gts.gestalt
}

// GetCycleCount returns the number of completed cycles.
func (gts *GlobalTelemetryShell) GetCycleCount() int64 {
	gts.mu.RLock()
	defer gts.mu.RUnlock()
	return gts.cycleCount
}

// GetUptime returns the uptime of the shell.
func (gts *GlobalTelemetryShell) GetUptime() time.Duration {
	gts.mu.RLock()
	defer gts.mu.RUnlock()
	if !gts.running {
		return 0
	}
	return time.Since(gts.startTime)
}

// GetCurrentStep returns the current step in the cognitive cycle.
func (gts *GlobalTelemetryShell) GetCurrentStep() int {
	return gts.orchestra.GetCurrentStep()
}

// TelemetryUpdate represents a telemetry update event.
type TelemetryUpdate struct {
	CoreID      string
	Timestamp   time.Time
	Observation *TelemetryObservation
}

// GestaltBroadcast represents a gestalt broadcast event.
type GestaltBroadcast struct {
	Timestamp time.Time
	Snapshot  *GestaltSnapshot
}
