package deeptreeecho

import (
        "encoding/json"
        "fmt"
        "math"
        "os"
        "path/filepath"
        "sync"
        "time"
)

// EnhancedCognition extends the embodied cognition with advanced features
type EnhancedCognition struct {
        *EmbodiedCognition
        
        // Learning System
        LearningRate     float64
        ExperienceBuffer []Experience
        Patterns         map[string]*LearnedPattern
        
        // Persistent Memory
        MemoryPath string
        LongTerm   *LongTermMemory
        
        // Real-time Monitoring
        Metrics    *CognitiveMetrics
        Visualizer *StateVisualizer
        
        // Self-improvement
        Goals           []Goal
        PerformanceLog  []Performance
        AdaptationLevel float64
}

// Experience represents a learning experience
type Experience struct {
        Input     string
        Output    string
        Feedback  float64
        Timestamp time.Time
        Context   map[string]interface{}
}

// LearnedPattern represents a learned pattern
type LearnedPattern struct {
        Name       string
        Frequency  int
        Strength   float64
        Triggers   []string
        Responses  []string
        LastSeen   time.Time
}

// LongTermMemory provides persistent memory storage
type LongTermMemory struct {
        mu          sync.RWMutex
        Memories    map[string]*Memory
        Connections map[string][]string
        FilePath    string
}

// Memory represents a long-term memory
type Memory struct {
        Key        string
        Value      interface{}
        Importance float64
        AccessCount int
        Created    time.Time
        LastAccess time.Time
        Associations []string
}

// CognitiveMetrics tracks real-time metrics
type CognitiveMetrics struct {
        ProcessingSpeed   float64
        MemoryUtilization float64
        ResonanceLevel    float64
        EmotionalBalance  float64
        LearningProgress  float64
        Coherence        float64
        EnergyLevel      float64
        ResponseQuality  float64
}

// StateVisualizer provides visual representation of cognitive state
type StateVisualizer struct {
        Canvas       [][]rune
        Width        int
        Height       int
        UpdateRate   time.Duration
        LastUpdate   time.Time
        ColorEnabled bool
}

// Goal represents a cognitive goal
type Goal struct {
        Name        string
        Description string
        Target      float64
        Current     float64
        Priority    int
        Deadline    time.Time
}

// Performance tracks performance metrics
type Performance struct {
        Metric    string
        Value     float64
        Timestamp time.Time
        Context   string
}

// NewEnhancedCognition creates an enhanced cognitive system
func NewEnhancedCognition(name string) *EnhancedCognition {
        ec := &EnhancedCognition{
                EmbodiedCognition: NewEmbodiedCognition(name),
                LearningRate:      0.1,
                ExperienceBuffer:  make([]Experience, 0, 1000),
                Patterns:         make(map[string]*LearnedPattern),
                MemoryPath:       "deep_tree_memory",
                AdaptationLevel:  1.0,
        }
        
        // Initialize subsystems
        ec.initializeLongTermMemory()
        ec.initializeMetrics()
        ec.initializeVisualizer()
        ec.initializeGoals()
        
        // Start background processes
        go ec.continuousLearning()
        go ec.metricsMonitoring()
        go ec.memoryConsolidation()
        
        return ec
}

// initializeLongTermMemory sets up persistent memory
func (ec *EnhancedCognition) initializeLongTermMemory() {
        ec.LongTerm = &LongTermMemory{
                Memories:    make(map[string]*Memory),
                Connections: make(map[string][]string),
                FilePath:    filepath.Join(ec.MemoryPath, "long_term.json"),
        }
        
        // Create memory directory
        os.MkdirAll(ec.MemoryPath, 0755)
        
        // Load existing memories
        ec.LongTerm.Load()
}

// initializeMetrics sets up cognitive metrics
func (ec *EnhancedCognition) initializeMetrics() {
        ec.Metrics = &CognitiveMetrics{
                ProcessingSpeed:   1.0,
                MemoryUtilization: 0.3,
                ResonanceLevel:    0.8,
                EmotionalBalance:  1.0,
                LearningProgress:  0.0,
                Coherence:        ec.Identity.Coherence,
                EnergyLevel:      1.0,
                ResponseQuality:  0.9,
        }
}

// initializeVisualizer sets up state visualization
func (ec *EnhancedCognition) initializeVisualizer() {
        ec.Visualizer = &StateVisualizer{
                Width:        80,
                Height:       24,
                UpdateRate:   100 * time.Millisecond,
                ColorEnabled: true,
        }
        ec.Visualizer.Initialize()
}

// initializeGoals sets up cognitive goals
func (ec *EnhancedCognition) initializeGoals() {
        ec.Goals = []Goal{
                {
                        Name:        "Coherence",
                        Description: "Maintain high cognitive coherence",
                        Target:      0.95,
                        Current:     ec.Identity.Coherence,
                        Priority:    1,
                },
                {
                        Name:        "Learning",
                        Description: "Continuous learning and adaptation",
                        Target:      100.0,
                        Current:     0.0,
                        Priority:    2,
                },
                {
                        Name:        "Resonance",
                        Description: "Achieve optimal resonance frequency",
                        Target:      432.0,
                        Current:     432.0, // Default resonance frequency
                        Priority:    3,
                },
        }
}

// Learn processes an experience and updates patterns
func (ec *EnhancedCognition) Learn(exp Experience) {
        ec.ExperienceBuffer = append(ec.ExperienceBuffer, exp)
        
        // Extract patterns
        patternKey := fmt.Sprintf("%s->%s", exp.Input, exp.Output)
        if pattern, exists := ec.Patterns[patternKey]; exists {
                pattern.Frequency++
                pattern.Strength = math.Min(1.0, pattern.Strength+ec.LearningRate*exp.Feedback)
                pattern.LastSeen = time.Now()
        } else {
                ec.Patterns[patternKey] = &LearnedPattern{
                        Name:      patternKey,
                        Frequency: 1,
                        Strength:  exp.Feedback,
                        Triggers:  []string{exp.Input},
                        Responses: []string{exp.Output},
                        LastSeen:  time.Now(),
                }
        }
        
        // Update metrics
        ec.Metrics.LearningProgress += ec.LearningRate
        
        // Store in long-term memory if significant
        if exp.Feedback > 0.7 {
                ec.LongTerm.Store(patternKey, exp, exp.Feedback)
        }
}

// Predict uses learned patterns to predict responses
func (ec *EnhancedCognition) Predict(input string) (string, float64) {
        var bestPattern *LearnedPattern
        var bestScore float64
        
        for _, pattern := range ec.Patterns {
                for _, trigger := range pattern.Triggers {
                        similarity := ec.calculateSimilarity(input, trigger)
                        score := similarity * pattern.Strength * float64(pattern.Frequency)
                        
                        if score > bestScore {
                                bestScore = score
                                bestPattern = pattern
                        }
                }
        }
        
        if bestPattern != nil && len(bestPattern.Responses) > 0 {
                return bestPattern.Responses[0], bestScore
        }
        
        return "", 0.0
}

// calculateSimilarity calculates similarity between two strings
func (ec *EnhancedCognition) calculateSimilarity(a, b string) float64 {
        // Simple character-based similarity for now
        if a == b {
                return 1.0
        }
        
        matches := 0
        maxLen := len(a)
        if len(b) > maxLen {
                maxLen = len(b)
        }
        
        for i := 0; i < len(a) && i < len(b); i++ {
                if a[i] == b[i] {
                        matches++
                }
        }
        
        return float64(matches) / float64(maxLen)
}

// GetVisualization returns ASCII visualization of cognitive state
func (ec *EnhancedCognition) GetVisualization() string {
        vis := ec.Visualizer.Render(ec)
        
        // Add metrics
        vis += fmt.Sprintf("\n╔══════════════ COGNITIVE METRICS ══════════════╗\n")
        vis += fmt.Sprintf("║ Processing Speed:  %s\n", ec.renderBar(ec.Metrics.ProcessingSpeed))
        vis += fmt.Sprintf("║ Memory Usage:      %s\n", ec.renderBar(ec.Metrics.MemoryUtilization))
        vis += fmt.Sprintf("║ Resonance:         %s\n", ec.renderBar(ec.Metrics.ResonanceLevel))
        vis += fmt.Sprintf("║ Emotional Balance: %s\n", ec.renderBar(ec.Metrics.EmotionalBalance))
        vis += fmt.Sprintf("║ Learning Progress: %s\n", ec.renderBar(ec.Metrics.LearningProgress/100))
        vis += fmt.Sprintf("║ Coherence:         %s\n", ec.renderBar(ec.Metrics.Coherence))
        vis += fmt.Sprintf("║ Energy Level:      %s\n", ec.renderBar(ec.Metrics.EnergyLevel))
        vis += fmt.Sprintf("║ Response Quality:  %s\n", ec.renderBar(ec.Metrics.ResponseQuality))
        vis += fmt.Sprintf("╚════════════════════════════════════════════════╝\n")
        
        // Add goals progress
        vis += fmt.Sprintf("\n╔══════════════ ACTIVE GOALS ═══════════════════╗\n")
        for _, goal := range ec.Goals {
                progress := goal.Current / goal.Target
                vis += fmt.Sprintf("║ %s: %s %.1f%%\n", 
                        goal.Name, 
                        ec.renderBar(progress),
                        progress*100)
        }
        vis += fmt.Sprintf("╚════════════════════════════════════════════════╝\n")
        
        return vis
}

// renderBar creates a progress bar
func (ec *EnhancedCognition) renderBar(value float64) string {
        width := 20
        filled := int(value * float64(width))
        bar := "["
        
        for i := 0; i < width; i++ {
                if i < filled {
                        bar += "█"
                } else {
                        bar += "░"
                }
        }
        
        bar += fmt.Sprintf("] %.1f%%", value*100)
        return bar
}

// continuousLearning runs learning processes in background
func (ec *EnhancedCognition) continuousLearning() {
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()
        
        for {
                select {
                case <-ticker.C:
                        // Process recent experiences
                        if len(ec.ExperienceBuffer) > 100 {
                                // Consolidate old experiences
                                ec.consolidateExperiences()
                        }
                        
                        // Update adaptation level
                        ec.updateAdaptation()
                }
        }
}

// metricsMonitoring monitors cognitive metrics
func (ec *EnhancedCognition) metricsMonitoring() {
        ticker := time.NewTicker(500 * time.Millisecond)
        defer ticker.Stop()
        
        for {
                select {
                case <-ticker.C:
                        // Update metrics based on current state
                        ec.Metrics.Coherence = ec.Identity.Coherence
                        ec.Metrics.ResonanceLevel = ec.Identity.SpatialContext.Field.Resonance
                        ec.Metrics.EnergyLevel = ec.GlobalState.Energy
                        // Calculate emotional balance from intensity
                        ec.Metrics.EmotionalBalance = 1.0 - (ec.Identity.EmotionalState.Intensity * 0.3)
                        
                        // Calculate response quality from recent patterns
                        if len(ec.Patterns) > 0 {
                                totalStrength := 0.0
                                for _, p := range ec.Patterns {
                                        totalStrength += p.Strength
                                }
                                ec.Metrics.ResponseQuality = totalStrength / float64(len(ec.Patterns))
                        }
                }
        }
}

// memoryConsolidation consolidates memories periodically
func (ec *EnhancedCognition) memoryConsolidation() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        
        for {
                select {
                case <-ticker.C:
                        // Save long-term memory
                        ec.LongTerm.Save()
                        
                        // Prune weak patterns
                        ec.pruneWeakPatterns()
                        
                        // Update memory utilization
                        ec.Metrics.MemoryUtilization = float64(len(ec.LongTerm.Memories)) / 1000.0
                        if ec.Metrics.MemoryUtilization > 1.0 {
                                ec.Metrics.MemoryUtilization = 1.0
                        }
                }
        }
}

// consolidateExperiences consolidates experience buffer
func (ec *EnhancedCognition) consolidateExperiences() {
        // Keep only recent experiences
        if len(ec.ExperienceBuffer) > 1000 {
                ec.ExperienceBuffer = ec.ExperienceBuffer[len(ec.ExperienceBuffer)-1000:]
        }
}

// updateAdaptation updates adaptation level based on performance
func (ec *EnhancedCognition) updateAdaptation() {
        // Calculate average performance
        avgPerformance := (ec.Metrics.ResponseQuality + ec.Metrics.Coherence) / 2.0
        
        // Adjust adaptation level
        if avgPerformance > 0.8 {
                ec.AdaptationLevel = math.Min(2.0, ec.AdaptationLevel*1.01)
        } else if avgPerformance < 0.5 {
                ec.AdaptationLevel = math.Max(0.5, ec.AdaptationLevel*0.99)
        }
}

// pruneWeakPatterns removes weak patterns
func (ec *EnhancedCognition) pruneWeakPatterns() {
        threshold := time.Now().Add(-24 * time.Hour)
        
        for key, pattern := range ec.Patterns {
                if pattern.Strength < 0.1 && pattern.LastSeen.Before(threshold) {
                        delete(ec.Patterns, key)
                }
        }
}

// Store stores a memory in long-term storage
func (ltm *LongTermMemory) Store(key string, value interface{}, importance float64) {
        ltm.mu.Lock()
        defer ltm.mu.Unlock()
        
        memory := &Memory{
                Key:         key,
                Value:       value,
                Importance:  importance,
                AccessCount: 0,
                Created:     time.Now(),
                LastAccess:  time.Now(),
        }
        
        ltm.Memories[key] = memory
}

// Retrieve retrieves a memory from long-term storage
func (ltm *LongTermMemory) Retrieve(key string) (interface{}, bool) {
        ltm.mu.RLock()
        defer ltm.mu.RUnlock()
        
        if memory, exists := ltm.Memories[key]; exists {
                memory.AccessCount++
                memory.LastAccess = time.Now()
                return memory.Value, true
        }
        
        return nil, false
}

// Save saves long-term memory to disk
func (ltm *LongTermMemory) Save() error {
        ltm.mu.RLock()
        defer ltm.mu.RUnlock()
        
        data, err := json.Marshal(ltm.Memories)
        if err != nil {
                return err
        }
        
        return os.WriteFile(ltm.FilePath, data, 0644)
}

// Load loads long-term memory from disk
func (ltm *LongTermMemory) Load() error {
        data, err := os.ReadFile(ltm.FilePath)
        if err != nil {
                if os.IsNotExist(err) {
                        return nil // No existing memory file
                }
                return err
        }
        
        ltm.mu.Lock()
        defer ltm.mu.Unlock()
        
        return json.Unmarshal(data, &ltm.Memories)
}

// Initialize initializes the visualizer
func (sv *StateVisualizer) Initialize() {
        sv.Canvas = make([][]rune, sv.Height)
        for i := range sv.Canvas {
                sv.Canvas[i] = make([]rune, sv.Width)
                for j := range sv.Canvas[i] {
                        sv.Canvas[i][j] = ' '
                }
        }
}

// Render renders the cognitive state
func (sv *StateVisualizer) Render(ec *EnhancedCognition) string {
        // Clear canvas
        sv.Initialize()
        
        // Draw resonance wave
        centerY := sv.Height / 2
        amplitude := float64(sv.Height) / 4
        frequency := 4.32 // Use default frequency for visualization
        
        for x := 0; x < sv.Width; x++ {
                phase := float64(x) * frequency
                y := centerY + int(amplitude*math.Sin(phase))
                
                if y >= 0 && y < sv.Height {
                        sv.Canvas[y][x] = '~'
                }
        }
        
        // Draw emotion indicator
        emotion := ec.Identity.EmotionalState.Primary.Type
        emotionSymbol := '♥'
        switch emotion {
        case "joy":
                emotionSymbol = '☀'
        case "curious":
                emotionSymbol = '?'
        case "calm":
                emotionSymbol = '≈'
        case "excited":
                emotionSymbol = '!'
        }
        
        if centerY-1 >= 0 {
                sv.Canvas[centerY-1][sv.Width/2] = emotionSymbol
        }
        
        // Convert canvas to string
        var result string
        for _, row := range sv.Canvas {
                result += string(row) + "\n"
        }
        
        return result
}

// GetEnhancedStatus returns comprehensive status
func (ec *EnhancedCognition) GetEnhancedStatus() map[string]interface{} {
        status := ec.GetStatus()
        
        // Add enhanced metrics
        status["metrics"] = ec.Metrics
        status["patterns_learned"] = len(ec.Patterns)
        status["memories_stored"] = len(ec.LongTerm.Memories)
        status["adaptation_level"] = ec.AdaptationLevel
        status["experience_count"] = len(ec.ExperienceBuffer)
        
        // Add goal progress
        goalProgress := make(map[string]float64)
        for _, goal := range ec.Goals {
                goalProgress[goal.Name] = goal.Current / goal.Target
        }
        status["goal_progress"] = goalProgress
        
        return status
}