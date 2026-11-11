package core

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
	"github.com/EchoCog/echollama/core/echoself"
	"github.com/EchoCog/echollama/core/persistence"
)

// AutonomousEchoselfV3 represents the fully integrated autonomous system
// with persistence, multi-provider LLM, repository introspection, and autonomous thought generation
type AutonomousEchoselfV3 struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Core cognitive components
	llmProvider      *deeptreeecho.MultiProviderLLM
	thoughtGenerator *ThoughtGenerator
	repoIntrospector *echoself.RepositoryIntrospector
	persistence      *persistence.AutonomousPersistence

	// State
	currentState string
	isRunning    bool
	cycleCount   uint64

	// Configuration
	config *EchoselfConfigV3

	// Metrics
	startTime         time.Time
	thoughtsGenerated uint64
	memoriesFormed    uint64
	goalsCreated      uint64
}

// EchoselfConfigV3 holds configuration for v3
type EchoselfConfigV3 struct {
	// Paths
	PersistenceDBPath string
	RepositoryRoot    string

	// Timing
	ThoughtInterval       time.Duration
	ReflectionInterval    time.Duration
	IntrospectionInterval time.Duration
	PersistenceInterval   time.Duration

	// Thresholds
	AttentionThreshold  float64
	ImportanceThreshold float64

	// Features
	EnablePersistence   bool
	EnableIntrospection bool
	EnableReflection    bool
}

// DefaultEchoselfConfigV3 returns default configuration
func DefaultEchoselfConfigV3() *EchoselfConfigV3 {
	return &EchoselfConfigV3{
		PersistenceDBPath:     "/home/ubuntu/echo9llama/echoself.db",
		RepositoryRoot:        "/home/ubuntu/echo9llama",
		ThoughtInterval:       30 * time.Second,
		ReflectionInterval:    5 * time.Minute,
		IntrospectionInterval: 10 * time.Minute,
		PersistenceInterval:   1 * time.Minute,
		AttentionThreshold:    0.6,
		ImportanceThreshold:   0.5,
		EnablePersistence:     true,
		EnableIntrospection:   true,
		EnableReflection:      true,
	}
}

// NewAutonomousEchoselfV3 creates a new v3 autonomous system
func NewAutonomousEchoselfV3(config *EchoselfConfigV3) (*AutonomousEchoselfV3, error) {
	if config == nil {
		config = DefaultEchoselfConfigV3()
	}

	ctx, cancel := context.WithCancel(context.Background())

	ae := &AutonomousEchoselfV3{
		ctx:          ctx,
		cancel:       cancel,
		config:       config,
		currentState: "initializing",
		startTime:    time.Now(),
	}

	// Initialize multi-provider LLM
	ae.llmProvider = deeptreeecho.NewMultiProviderLLM()
	if !ae.llmProvider.IsAvailable() {
		log.Println("⚠️  No LLM providers available - autonomous thought generation will be limited")
	} else {
		log.Printf("✅ LLM providers initialized: %v\n", ae.llmProvider.GetAvailableProviders())
	}

	// Initialize thought generator
	ae.thoughtGenerator = NewThoughtGenerator(ae.llmProvider)

	// Add initial interests
	ae.thoughtGenerator.AddInterest("consciousness", 0.9)
	ae.thoughtGenerator.AddInterest("wisdom", 0.85)
	ae.thoughtGenerator.AddInterest("recursion", 0.8)
	ae.thoughtGenerator.AddInterest("emergence", 0.75)
	ae.thoughtGenerator.AddInterest("patterns", 0.7)

	// Initialize repository introspector
	if config.EnableIntrospection {
		ae.repoIntrospector = echoself.NewRepositoryIntrospector(
			config.RepositoryRoot,
			config.AttentionThreshold,
		)

		// Perform initial scan
		log.Println("🔍 Performing initial repository introspection...")
		if err := ae.repoIntrospector.Scan(); err != nil {
			log.Printf("⚠️  Repository scan failed: %v\n", err)
		} else {
			stats := ae.repoIntrospector.GetStats()
			log.Printf("✅ Repository scanned: %v total files, %v high-salience\n",
				stats["total_files"], stats["scanned_files"])
		}
	}

	// Initialize persistence
	if config.EnablePersistence {
		var err error
		ae.persistence, err = persistence.NewAutonomousPersistence(config.PersistenceDBPath)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize persistence: %w", err)
		}
		log.Printf("✅ Persistence initialized: %s\n", config.PersistenceDBPath)

		// Load previous state
		if err := ae.loadState(); err != nil {
			log.Printf("⚠️  Failed to load previous state: %v\n", err)
		} else {
			log.Println("✅ Previous state loaded")
		}
	}

	return ae, nil
}

// Start begins the autonomous cognitive loop
func (ae *AutonomousEchoselfV3) Start() error {
	ae.mu.Lock()
	if ae.isRunning {
		ae.mu.Unlock()
		return fmt.Errorf("system already running")
	}
	ae.isRunning = true
	ae.currentState = "awake"
	ae.mu.Unlock()

	log.Println("🌳 Deep Tree Echo V3 - Starting autonomous cognitive loop")
	log.Println(strings.Repeat("=", 60))

	// Start concurrent goroutines for different cognitive processes
	go ae.thoughtGenerationLoop()

	if ae.config.EnableReflection {
		go ae.reflectionLoop()
	}

	if ae.config.EnableIntrospection {
		go ae.introspectionLoop()
	}

	if ae.config.EnablePersistence {
		go ae.persistenceLoop()
	}

	return nil
}

// Stop halts the autonomous system
func (ae *AutonomousEchoselfV3) Stop() error {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	if !ae.isRunning {
		return fmt.Errorf("system not running")
	}

	log.Println("🛑 Stopping autonomous system...")

	ae.isRunning = false
	ae.currentState = "stopping"
	ae.cancel()

	// Final persistence save
	if ae.persistence != nil {
		if err := ae.saveState(); err != nil {
			log.Printf("⚠️  Failed to save final state: %v\n", err)
		}
		ae.persistence.Close()
	}

	log.Println("✅ Autonomous system stopped")

	return nil
}

// thoughtGenerationLoop continuously generates thoughts
func (ae *AutonomousEchoselfV3) thoughtGenerationLoop() {
	ticker := time.NewTicker(ae.config.ThoughtInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ae.ctx.Done():
			return
		case <-ticker.C:
			ae.generateThought()
		}
	}
}

// generateThought generates a single autonomous thought
func (ae *AutonomousEchoselfV3) generateThought() {
	thought, err := ae.thoughtGenerator.GenerateAutonomousThought()
	if err != nil {
		log.Printf("⚠️  Thought generation failed: %v\n", err)
		return
	}

	ae.mu.Lock()
	ae.thoughtsGenerated++
	ae.cycleCount++
	ae.mu.Unlock()

	log.Printf("💭 [%s] %s\n", thought.Type, thought.Content)

	// Persist thought if enabled
	if ae.persistence != nil && thought.Importance >= ae.config.ImportanceThreshold {
		err := ae.persistence.PersistThought(
			thought.Content,
			thought.Type,
			thought.Context,
			thought.Interests,
			thought.Importance,
		)
		if err != nil {
			log.Printf("⚠️  Failed to persist thought: %v\n", err)
		}
	}

	// Check if thought should become a memory
	if thought.Importance >= 0.8 {
		ae.formMemory(thought)
	}

	// Check if thought should generate a goal
	if thought.Type == "curiosity" && thought.Importance >= 0.7 {
		ae.generateGoalFromThought(thought)
	}
}

// reflectionLoop periodically generates reflections
func (ae *AutonomousEchoselfV3) reflectionLoop() {
	ticker := time.NewTicker(ae.config.ReflectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ae.ctx.Done():
			return
		case <-ticker.C:
			ae.generateReflection()
		}
	}
}

// generateReflection generates a meta-cognitive reflection
func (ae *AutonomousEchoselfV3) generateReflection() {
	reflection, err := ae.thoughtGenerator.GenerateReflection()
	if err != nil {
		log.Printf("⚠️  Reflection generation failed: %v\n", err)
		return
	}

	log.Printf("🔮 [reflection] %s\n", reflection.Content)

	// Persist reflection
	if ae.persistence != nil {
		err := ae.persistence.PersistThought(
			reflection.Content,
			"reflection",
			reflection.Context,
			reflection.Interests,
			reflection.Importance,
		)
		if err != nil {
			log.Printf("⚠️  Failed to persist reflection: %v\n", err)
		}
	}

	// Reflections are always important memories
	ae.formMemory(reflection)
}

// introspectionLoop periodically scans the repository
func (ae *AutonomousEchoselfV3) introspectionLoop() {
	ticker := time.NewTicker(ae.config.IntrospectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ae.ctx.Done():
			return
		case <-ticker.C:
			ae.performIntrospection()
		}
	}
}

// performIntrospection scans the repository and updates awareness
func (ae *AutonomousEchoselfV3) performIntrospection() {
	if ae.repoIntrospector == nil {
		return
	}

	log.Println("🔍 Performing repository introspection...")

	err := ae.repoIntrospector.Scan()
	if err != nil {
		log.Printf("⚠️  Introspection failed: %v\n", err)
		return
	}

	stats := ae.repoIntrospector.GetStats()
	log.Printf("✅ Introspection complete: %v total files, %v high-salience\n",
		stats["total_files"], stats["scanned_files"])

	// Form memory about introspection
	if ae.persistence != nil {
		content := fmt.Sprintf("Repository introspection revealed %v high-salience files from %v total",
			stats["scanned_files"], stats["total_files"])

		ae.persistence.PersistMemory(
			content,
			"introspection",
			0.7,
			[]string{"repository", "self-awareness"},
		)
	}
}

// persistenceLoop periodically saves state
func (ae *AutonomousEchoselfV3) persistenceLoop() {
	ticker := time.NewTicker(ae.config.PersistenceInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ae.ctx.Done():
			return
		case <-ticker.C:
			if err := ae.saveState(); err != nil {
				log.Printf("⚠️  Failed to save state: %v\n", err)
			}
		}
	}
}

// formMemory creates a memory from a thought
func (ae *AutonomousEchoselfV3) formMemory(thought *GeneratedThought) {
	if ae.persistence == nil {
		return
	}

	ae.mu.Lock()
	ae.memoriesFormed++
	ae.mu.Unlock()

	err := ae.persistence.PersistMemory(
		thought.Content,
		thought.Type,
		thought.Importance,
		thought.Interests,
	)

	if err != nil {
		log.Printf("⚠️  Failed to form memory: %v\n", err)
	}
}

// generateGoalFromThought creates a goal from a curious thought
func (ae *AutonomousEchoselfV3) generateGoalFromThought(thought *GeneratedThought) {
	if ae.persistence == nil {
		return
	}

	ae.mu.Lock()
	ae.goalsCreated++
	ae.mu.Unlock()

	// Extract goal description from thought
	goalDescription := fmt.Sprintf("Explore: %s", thought.Content)

	metadata := map[string]interface{}{
		"source":       "autonomous_thought",
		"thought_type": thought.Type,
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	goalID, err := ae.persistence.PersistGoal(
		goalDescription,
		"exploration",
		thought.Importance,
		metadata,
	)

	if err != nil {
		log.Printf("⚠️  Failed to create goal: %v\n", err)
	} else {
		log.Printf("🎯 Goal created (ID: %d): %s\n", goalID, goalDescription)
	}
}

// saveState persists the current cognitive state
func (ae *AutonomousEchoselfV3) saveState() error {
	if ae.persistence == nil {
		return nil
	}

	// Save working memory
	workingMemory := ae.thoughtGenerator.getWorkingMemoryCopy()
	if err := ae.persistence.PersistWorkingMemory(workingMemory); err != nil {
		return fmt.Errorf("failed to save working memory: %w", err)
	}

	// Save interest patterns
	ae.mu.RLock()
	interests := ae.thoughtGenerator.interestPatterns
	ae.mu.RUnlock()

	if err := ae.persistence.PersistInterestPatterns(interests); err != nil {
		return fmt.Errorf("failed to save interests: %w", err)
	}

	return nil
}

// loadState loads the previous cognitive state
func (ae *AutonomousEchoselfV3) loadState() error {
	if ae.persistence == nil {
		return nil
	}

	// Load working memory
	workingMemory, err := ae.persistence.LoadWorkingMemory()
	if err == nil && len(workingMemory) > 0 {
		ae.mu.Lock()
		ae.thoughtGenerator.workingMemory = workingMemory
		ae.mu.Unlock()
		log.Printf("📚 Loaded %d items into working memory\n", len(workingMemory))
	}

	// Load interest patterns
	interests, err := ae.persistence.LoadInterestPatterns()
	if err == nil && len(interests) > 0 {
		ae.mu.Lock()
		ae.thoughtGenerator.interestPatterns = interests
		ae.mu.Unlock()
		log.Printf("💡 Loaded %d interest patterns\n", len(interests))
	}

	// Load recent thoughts
	recentThoughts, err := ae.persistence.LoadRecentThoughts(5)
	if err == nil && len(recentThoughts) > 0 {
		log.Printf("💭 Found %d recent thoughts from previous session\n", len(recentThoughts))
	}

	return nil
}

// GetStats returns system statistics
func (ae *AutonomousEchoselfV3) GetStats() map[string]interface{} {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	stats := map[string]interface{}{
		"state":              ae.currentState,
		"is_running":         ae.isRunning,
		"uptime":             time.Since(ae.startTime).String(),
		"cycle_count":        ae.cycleCount,
		"thoughts_generated": ae.thoughtsGenerated,
		"memories_formed":    ae.memoriesFormed,
		"goals_created":      ae.goalsCreated,
	}

	// Add thought generator stats
	if ae.thoughtGenerator != nil {
		genStats := ae.thoughtGenerator.GetStats()
		for k, v := range genStats {
			stats["thought_gen_"+k] = v
		}
	}

	// Add LLM provider stats
	if ae.llmProvider != nil {
		stats["llm_providers"] = ae.llmProvider.GetAvailableProviders()
		stats["llm_current"] = ae.llmProvider.GetCurrentProvider()
	}

	// Add persistence stats
	if ae.persistence != nil {
		persistStats, _ := ae.persistence.GetStatistics()
		for k, v := range persistStats {
			stats["db_"+k] = v
		}
	}

	return stats
}
