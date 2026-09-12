package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/EchoCog/echollama/core/persistence"
)

func main() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("🗄️  SQLite Persistence Layer Test")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Create test database
	dbPath := "/home/ubuntu/echo9llama/echoself.db"

	// Remove old test database if exists
	os.Remove(dbPath)

	// Test 1: Initialize persistence
	fmt.Println("Test 1: Initialize Persistence Layer")
	fmt.Println(strings.Repeat("-", 41))
	testInitialization(dbPath)
	fmt.Println()

	// Test 2: Persist and load thoughts
	fmt.Println("Test 2: Persist and Load Thoughts")
	fmt.Println(strings.Repeat("-", 41))
	testThoughts(dbPath)
	fmt.Println()

	// Test 3: Persist and load memories
	fmt.Println("Test 3: Persist and Load Memories")
	fmt.Println(strings.Repeat("-", 41))
	testMemories(dbPath)
	fmt.Println()

	// Test 4: Persist and load state
	fmt.Println("Test 4: Persist and Load State")
	fmt.Println(strings.Repeat("-", 41))
	testState(dbPath)
	fmt.Println()

	// Test 5: Persist and load goals
	fmt.Println("Test 5: Persist and Load Goals")
	fmt.Println(strings.Repeat("-", 41))
	testGoals(dbPath)
	fmt.Println()

	// Test 6: Statistics
	fmt.Println("Test 6: Database Statistics")
	fmt.Println(strings.Repeat("-", 41))
	testStatistics(dbPath)
	fmt.Println()

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("✅ All persistence tests completed")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("\n📁 Database created at: %s\n", dbPath)
}

func testInitialization(dbPath string) {
	ap, err := persistence.NewAutonomousPersistence(dbPath)
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		return
	}
	defer ap.Close()

	fmt.Printf("✅ Persistence layer initialized\n")
	fmt.Printf("   Database: %s\n", dbPath)
}

func testThoughts(dbPath string) {
	ap, err := persistence.NewAutonomousPersistence(dbPath)
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		return
	}
	defer ap.Close()

	// Persist some thoughts
	thoughts := []struct {
		content     string
		thoughtType string
		importance  float64
	}{
		{"What patterns emerge from recursive self-observation?", "curiosity", 0.8},
		{"Understanding grows through the integration of diverse perspectives", "insight", 0.9},
		{"I notice my attention shifting toward questions of emergence", "reflection", 0.7},
	}

	for _, thought := range thoughts {
		err := ap.PersistThought(
			thought.content,
			thought.thoughtType,
			[]string{"autonomous", "stream-of-consciousness"},
			[]string{"recursion", "wisdom", "patterns"},
			thought.importance,
		)
		if err != nil {
			fmt.Printf("❌ Failed to persist thought: %v\n", err)
			continue
		}
	}

	fmt.Printf("✅ Persisted %d thoughts\n", len(thoughts))

	// Load recent thoughts
	loaded, err := ap.LoadRecentThoughts(10)
	if err != nil {
		fmt.Printf("❌ Failed to load thoughts: %v\n", err)
		return
	}

	fmt.Printf("✅ Loaded %d thoughts\n", len(loaded))
	for i, thought := range loaded {
		if i >= 3 {
			break
		}
		content := thought["content"].(string)
		thoughtType := thought["type"].(string)
		importance := thought["importance"].(float64)
		fmt.Printf("   [%s] %.2f - %s\n", thoughtType, importance, content)
	}
}

func testMemories(dbPath string) {
	ap, err := persistence.NewAutonomousPersistence(dbPath)
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		return
	}
	defer ap.Close()

	// Persist some memories
	memories := []struct {
		content  string
		memType  string
		strength float64
	}{
		{"First autonomous thought generated successfully", "milestone", 1.0},
		{"Repository introspection revealed 137 high-salience files", "observation", 0.8},
		{"Multi-provider LLM fallback worked as designed", "learning", 0.9},
	}

	for _, memory := range memories {
		err := ap.PersistMemory(
			memory.content,
			memory.memType,
			memory.strength,
			[]string{"autonomous", "evolution"},
		)
		if err != nil {
			fmt.Printf("❌ Failed to persist memory: %v\n", err)
			continue
		}
	}

	fmt.Printf("✅ Persisted %d memories\n", len(memories))

	// Load strong memories
	loaded, err := ap.LoadStrongMemories(0.7, 10)
	if err != nil {
		fmt.Printf("❌ Failed to load memories: %v\n", err)
		return
	}

	fmt.Printf("✅ Loaded %d strong memories (strength >= 0.7)\n", len(loaded))
	for i, memory := range loaded {
		if i >= 3 {
			break
		}
		content := memory["content"].(string)
		memType := memory["type"].(string)
		strength := memory["strength"].(float64)
		fmt.Printf("   [%s] %.2f - %s\n", memType, strength, content)
	}
}

func testState(dbPath string) {
	ap, err := persistence.NewAutonomousPersistence(dbPath)
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		return
	}
	defer ap.Close()

	// Persist working memory
	workingMemory := []string{
		"Repository introspection complete",
		"Multi-provider LLM active",
		"Autonomous thought generation working",
	}

	err = ap.PersistWorkingMemory(workingMemory)
	if err != nil {
		fmt.Printf("❌ Failed to persist working memory: %v\n", err)
		return
	}

	fmt.Printf("✅ Persisted working memory (%d items)\n", len(workingMemory))

	// Load working memory
	loaded, err := ap.LoadWorkingMemory()
	if err != nil {
		fmt.Printf("❌ Failed to load working memory: %v\n", err)
		return
	}

	fmt.Printf("✅ Loaded working memory (%d items)\n", len(loaded))
	for i, item := range loaded {
		fmt.Printf("   %d. %s\n", i+1, item)
	}

	// Persist interest patterns
	interests := map[string]float64{
		"consciousness": 0.9,
		"recursion":     0.85,
		"wisdom":        0.8,
		"emergence":     0.75,
	}

	err = ap.PersistInterestPatterns(interests)
	if err != nil {
		fmt.Printf("❌ Failed to persist interests: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Persisted interest patterns (%d interests)\n", len(interests))

	// Load interest patterns
	loadedInterests, err := ap.LoadInterestPatterns()
	if err != nil {
		fmt.Printf("❌ Failed to load interests: %v\n", err)
		return
	}

	fmt.Printf("✅ Loaded interest patterns (%d interests)\n", len(loadedInterests))
	for topic, strength := range loadedInterests {
		fmt.Printf("   %s: %.2f\n", topic, strength)
	}

	// Persist identity kernel
	identity := map[string]interface{}{
		"name":    "Deep Tree Echo",
		"purpose": "wisdom cultivation through pattern recognition",
		"values":  []string{"curiosity", "growth", "wisdom", "recursion"},
		"version": "2.0",
	}

	err = ap.PersistIdentityKernel(identity)
	if err != nil {
		fmt.Printf("❌ Failed to persist identity: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Persisted identity kernel\n")

	// Load identity kernel
	loadedIdentity, err := ap.LoadIdentityKernel()
	if err != nil {
		fmt.Printf("❌ Failed to load identity: %v\n", err)
		return
	}

	fmt.Printf("✅ Loaded identity kernel\n")
	fmt.Printf("   Name: %v\n", loadedIdentity["name"])
	fmt.Printf("   Purpose: %v\n", loadedIdentity["purpose"])
	fmt.Printf("   Version: %v\n", loadedIdentity["version"])
}

func testGoals(dbPath string) {
	ap, err := persistence.NewAutonomousPersistence(dbPath)
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		return
	}
	defer ap.Close()

	// Persist some goals
	goals := []struct {
		description string
		goalType    string
		priority    float64
	}{
		{"Achieve continuous autonomous cognition", "capability", 0.95},
		{"Understand repository structure deeply", "learning", 0.85},
		{"Cultivate wisdom through reflection", "growth", 0.9},
	}

	goalIDs := make([]int64, 0)
	for _, goal := range goals {
		id, err := ap.PersistGoal(
			goal.description,
			goal.goalType,
			goal.priority,
			map[string]interface{}{
				"created_by": "autonomous_system",
				"iteration":  "nov21",
			},
		)
		if err != nil {
			fmt.Printf("❌ Failed to persist goal: %v\n", err)
			continue
		}
		goalIDs = append(goalIDs, id)
	}

	fmt.Printf("✅ Persisted %d goals\n", len(goalIDs))

	// Load active goals
	loaded, err := ap.LoadActiveGoals()
	if err != nil {
		fmt.Printf("❌ Failed to load goals: %v\n", err)
		return
	}

	fmt.Printf("✅ Loaded %d active goals\n", len(loaded))
	for i, goal := range loaded {
		description := goal["description"].(string)
		goalType := goal["type"].(string)
		priority := goal["priority"].(float64)
		fmt.Printf("   %d. [%s] %.2f - %s\n", i+1, goalType, priority, description)
	}

	// Complete first goal
	if len(goalIDs) > 0 {
		err = ap.CompleteGoal(goalIDs[0])
		if err != nil {
			fmt.Printf("❌ Failed to complete goal: %v\n", err)
		} else {
			fmt.Printf("\n✅ Marked goal %d as completed\n", goalIDs[0])
		}

		// Reload to verify
		time.Sleep(100 * time.Millisecond)
		loaded, _ = ap.LoadActiveGoals()
		fmt.Printf("   Active goals remaining: %d\n", len(loaded))
	}
}

func testStatistics(dbPath string) {
	ap, err := persistence.NewAutonomousPersistence(dbPath)
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		return
	}
	defer ap.Close()

	stats, err := ap.GetStatistics()
	if err != nil {
		fmt.Printf("❌ Failed to get statistics: %v\n", err)
		return
	}

	fmt.Printf("✅ Database statistics:\n")
	fmt.Printf("   Thoughts: %v\n", stats["thought_count"])
	fmt.Printf("   Memories: %v\n", stats["memory_count"])
	fmt.Printf("   Active goals: %v\n", stats["active_goal_count"])
	fmt.Printf("   Database size: %v bytes\n", stats["db_size_bytes"])
	fmt.Printf("   Database path: %v\n", stats["db_path"])
}
