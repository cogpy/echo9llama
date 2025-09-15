
package memory

import (
	"sync"
	"time"
)

// MemoryWeaver dynamically adapts memory connection patterns
type MemoryWeaver struct {
	mu               sync.RWMutex
	connections      map[string][]Connection
	usagePatterns    map[string]UsagePattern
	adaptationCycles []AdaptationCycle
	detector         PatternDetector
	lastWeave        time.Time
}

// Connection represents a memory connection with adaptive weight
type Connection struct {
	FromNode   string
	ToNode     string
	Weight     float64
	Usage      int64
	LastAccess time.Time
	Strength   float64
}

// UsagePattern tracks how memory connections are used
type UsagePattern struct {
	ConnectionID string
	AccessCount  int64
	Frequency    float64
	Context      map[string]interface{}
	Trend        string // "increasing", "decreasing", "stable"
}

// AdaptationCycle records memory restructuring events
type AdaptationCycle struct {
	Timestamp      time.Time
	ConnectionsAdded   int
	ConnectionsRemoved int
	WeightAdjustments  int
	EfficiencyGain     float64
}

// PatternDetector identifies usage patterns for adaptation
type PatternDetector interface {
	DetectUsagePatterns(connections []Connection) []UsagePattern
	SuggestAdaptations(patterns []UsagePattern) []Adaptation
}

// Adaptation represents a suggested memory structure change
type Adaptation struct {
	Type        string // "strengthen", "weaken", "create", "remove"
	FromNode    string
	ToNode      string
	NewWeight   float64
	Confidence  float64
}

// NewMemoryWeaver creates a new adaptive memory weaver
func NewMemoryWeaver(detector PatternDetector) *MemoryWeaver {
	return &MemoryWeaver{
		connections:      make(map[string][]Connection),
		usagePatterns:    make(map[string]UsagePattern),
		adaptationCycles: make([]AdaptationCycle, 0),
		detector:         detector,
		lastWeave:        time.Now(),
	}
}

// WeaveConnections adapts memory structure based on usage patterns
func (mw *MemoryWeaver) WeaveConnections() error {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	
	// Collect all connections for analysis
	var allConnections []Connection
	for _, conns := range mw.connections {
		allConnections = append(allConnections, conns...)
	}
	
	// Detect usage patterns
	patterns := mw.detector.DetectUsagePatterns(allConnections)
	
	// Update pattern tracking
	for _, pattern := range patterns {
		mw.usagePatterns[pattern.ConnectionID] = pattern
	}
	
	// Get adaptation suggestions
	adaptations := mw.detector.SuggestAdaptations(patterns)
	
	// Apply adaptations
	cycle := AdaptationCycle{
		Timestamp: time.Now(),
	}
	
	for _, adaptation := range adaptations {
		switch adaptation.Type {
		case "strengthen":
			mw.strengthenConnection(adaptation.FromNode, adaptation.ToNode, adaptation.NewWeight)
			cycle.WeightAdjustments++
		case "weaken":
			mw.weakenConnection(adaptation.FromNode, adaptation.ToNode, adaptation.NewWeight)
			cycle.WeightAdjustments++
		case "create":
			mw.createConnection(adaptation.FromNode, adaptation.ToNode, adaptation.NewWeight)
			cycle.ConnectionsAdded++
		case "remove":
			mw.removeConnection(adaptation.FromNode, adaptation.ToNode)
			cycle.ConnectionsRemoved++
		}
	}
	
	mw.adaptationCycles = append(mw.adaptationCycles, cycle)
	mw.lastWeave = time.Now()
	
	return nil
}

// strengthenConnection increases connection weight
func (mw *MemoryWeaver) strengthenConnection(from, to string, newWeight float64) {
	for nodeID, conns := range mw.connections {
		for i, conn := range conns {
			if conn.FromNode == from && conn.ToNode == to {
				mw.connections[nodeID][i].Weight = newWeight
				mw.connections[nodeID][i].Strength += 0.1
				return
			}
		}
	}
}

// weakenConnection decreases connection weight
func (mw *MemoryWeaver) weakenConnection(from, to string, newWeight float64) {
	for nodeID, conns := range mw.connections {
		for i, conn := range conns {
			if conn.FromNode == from && conn.ToNode == to {
				mw.connections[nodeID][i].Weight = newWeight
				mw.connections[nodeID][i].Strength -= 0.1
				return
			}
		}
	}
}

// createConnection adds new memory connection
func (mw *MemoryWeaver) createConnection(from, to string, weight float64) {
	newConn := Connection{
		FromNode:   from,
		ToNode:     to,
		Weight:     weight,
		Usage:      0,
		LastAccess: time.Now(),
		Strength:   0.5,
	}
	
	mw.connections[from] = append(mw.connections[from], newConn)
}

// removeConnection removes memory connection
func (mw *MemoryWeaver) removeConnection(from, to string) {
	for nodeID, conns := range mw.connections {
		for i, conn := range conns {
			if conn.FromNode == from && conn.ToNode == to {
				mw.connections[nodeID] = append(conns[:i], conns[i+1:]...)
				return
			}
		}
	}
}

// GetAdaptationHistory returns weaving adaptation history
func (mw *MemoryWeaver) GetAdaptationHistory() []AdaptationCycle {
	mw.mu.RLock()
	defer mw.mu.RUnlock()
	return mw.adaptationCycles
}
