package lampstack

import (
	"context"
	"testing"
	"time"
)

func TestNewStack(t *testing.T) {
	stack, err := NewStack(nil)
	if err != nil {
		t.Fatalf("NewStack failed: %v", err)
	}
	if stack == nil {
		t.Fatal("NewStack returned nil stack")
	}
	if stack.config == nil {
		t.Fatal("Stack config is nil")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if !config.EnableEchobeats {
		t.Error("EnableEchobeats should be true by default")
	}
	if !config.EnableConsciousness {
		t.Error("EnableConsciousness should be true by default")
	}
	if !config.EnableDreaming {
		t.Error("EnableDreaming should be true by default")
	}
	if config.CycleDuration != 30*time.Second {
		t.Errorf("CycleDuration should be 30s, got %v", config.CycleDuration)
	}
	if config.PrimaryProvider != "anthropic" {
		t.Errorf("PrimaryProvider should be 'anthropic', got %s", config.PrimaryProvider)
	}
	if config.MemoryCapacity != 10000 {
		t.Errorf("MemoryCapacity should be 10000, got %d", config.MemoryCapacity)
	}
}

func TestUnifiedMemoryStore(t *testing.T) {
	config := DefaultConfig()
	memory := NewUnifiedMemoryStore(config)

	ctx := context.Background()

	// Test storing a memory
	entry := &MemoryEntry{
		Type:       MemoryTypeEpisodic,
		Content:    "Test memory content",
		Importance: 0.7,
		Tags:       []string{"test", "unit-test"},
	}

	err := memory.Store(ctx, entry)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	if entry.ID == "" {
		t.Error("Entry ID should be set after store")
	}

	// Test retrieval
	retrieved, err := memory.Get(ctx, entry.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Content != entry.Content {
		t.Errorf("Retrieved content mismatch: got %s, want %s", retrieved.Content, entry.Content)
	}

	// Test query retrieval
	results, err := memory.Retrieve(ctx, "test memory", 5)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Should have retrieved at least one memory")
	}

	// Test stats
	stats := memory.GetStats()
	if stats["total_memories"].(int64) != 1 {
		t.Errorf("Total memories should be 1, got %v", stats["total_memories"])
	}
}

func TestUnifiedMemoryStoreTypes(t *testing.T) {
	config := DefaultConfig()
	memory := NewUnifiedMemoryStore(config)
	ctx := context.Background()

	// Store different types
	types := []MemoryType{
		MemoryTypeEpisodic,
		MemoryTypeSemantic,
		MemoryTypeProcedural,
		MemoryTypeWorking,
	}

	for _, memType := range types {
		entry := &MemoryEntry{
			Type:       memType,
			Content:    "Content for " + memType.String(),
			Importance: 0.5,
		}
		if err := memory.Store(ctx, entry); err != nil {
			t.Fatalf("Failed to store %s memory: %v", memType.String(), err)
		}
	}

	// Verify counts
	stats := memory.GetStats()
	if stats["episodic_count"].(int64) != 1 {
		t.Errorf("Episodic count should be 1, got %v", stats["episodic_count"])
	}
	if stats["semantic_count"].(int64) != 1 {
		t.Errorf("Semantic count should be 1, got %v", stats["semantic_count"])
	}
}

func TestUnifiedMemoryConsolidation(t *testing.T) {
	config := DefaultConfig()
	memory := NewUnifiedMemoryStore(config)
	ctx := context.Background()

	// Store an episodic memory with high importance and access count
	entry := &MemoryEntry{
		Type:        MemoryTypeEpisodic,
		Content:     "Important frequently accessed memory",
		Importance:  0.8,
		AccessCount: 10, // High access count
	}
	memory.Store(ctx, entry)

	// Run consolidation
	err := memory.Consolidate(ctx)
	if err != nil {
		t.Fatalf("Consolidate failed: %v", err)
	}

	stats := memory.GetStats()
	if stats["consolidations"].(int64) != 1 {
		t.Errorf("Consolidations should be 1, got %v", stats["consolidations"])
	}
}

func TestJSONPersistence(t *testing.T) {
	config := DefaultConfig()
	config.DataDirectory = t.TempDir()
	config.AutoSave = false

	persistence := NewJSONPersistence(config)
	defer persistence.Close()

	ctx := context.Background()

	// Test save and load
	testData := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	err := persistence.Save(ctx, "test_key", testData)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	var loaded map[string]interface{}
	err = persistence.Load(ctx, "test_key", &loaded)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded["key1"] != "value1" {
		t.Errorf("Loaded key1 mismatch: got %v", loaded["key1"])
	}

	// Test exists
	exists, err := persistence.Exists(ctx, "test_key")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Key should exist")
	}

	// Test delete
	err = persistence.Delete(ctx, "test_key")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	exists, _ = persistence.Exists(ctx, "test_key")
	if exists {
		t.Error("Key should not exist after delete")
	}
}

func TestMessageTypes(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "Hello, world!",
	}

	if msg.Role != "user" {
		t.Errorf("Message role mismatch")
	}
}

func TestThoughtTypes(t *testing.T) {
	types := []ThoughtType{
		ThoughtTypePerception,
		ThoughtTypeReflection,
		ThoughtTypeSimulation,
		ThoughtTypePlanning,
		ThoughtTypeCreative,
		ThoughtTypeAnalytical,
	}

	expectedStrings := []string{
		"Perception",
		"Reflection",
		"Simulation",
		"Planning",
		"Creative",
		"Analytical",
	}

	for i, tt := range types {
		if tt.String() != expectedStrings[i] {
			t.Errorf("ThoughtType %d string mismatch: got %s, want %s", i, tt.String(), expectedStrings[i])
		}
	}
}

func TestMemoryTypes(t *testing.T) {
	types := []MemoryType{
		MemoryTypeEpisodic,
		MemoryTypeSemantic,
		MemoryTypeProcedural,
		MemoryTypeWorking,
	}

	expectedStrings := []string{
		"Episodic",
		"Semantic",
		"Procedural",
		"Working",
	}

	for i, mt := range types {
		if mt.String() != expectedStrings[i] {
			t.Errorf("MemoryType %d string mismatch: got %s, want %s", i, mt.String(), expectedStrings[i])
		}
	}
}

func TestEventTypes(t *testing.T) {
	types := []EventType{
		EventTypeThought,
		EventTypeReflection,
		EventTypeStateChange,
	}

	for _, et := range types {
		str := et.String()
		if str == "" {
			t.Errorf("EventType %d has empty string", et)
		}
	}
}

func TestStackMetrics(t *testing.T) {
	metrics := &StackMetrics{
		StartTime: time.Now().Add(-1 * time.Hour),
	}

	metrics.incrementRequests()
	if metrics.TotalRequests != 1 {
		t.Errorf("TotalRequests should be 1, got %d", metrics.TotalRequests)
	}

	metrics.incrementFailed()
	if metrics.FailedRequests != 1 {
		t.Errorf("FailedRequests should be 1, got %d", metrics.FailedRequests)
	}

	metrics.recordSuccess(100 * time.Millisecond)
	if metrics.SuccessfulRequests != 1 {
		t.Errorf("SuccessfulRequests should be 1, got %d", metrics.SuccessfulRequests)
	}
	if metrics.AverageLatencyMs != 100 {
		t.Errorf("AverageLatencyMs should be 100, got %f", metrics.AverageLatencyMs)
	}

	metrics.incrementThoughts()
	if metrics.ThoughtsGenerated != 1 {
		t.Errorf("ThoughtsGenerated should be 1, got %d", metrics.ThoughtsGenerated)
	}

	metrics.incrementReflections()
	if metrics.ReflectionsGenerated != 1 {
		t.Errorf("ReflectionsGenerated should be 1, got %d", metrics.ReflectionsGenerated)
	}
}

func TestSimilarity(t *testing.T) {
	tests := []struct {
		a, b     string
		expected bool // true if similarity > 0
	}{
		{"hello world", "hello world", true},
		{"hello world", "hello there", true},
		{"completely different", "nothing alike", false},
		{"", "something", false},
		{"something", "", false},
	}

	for _, tt := range tests {
		sim := similarity(tt.a, tt.b)
		if tt.expected && sim == 0 {
			t.Errorf("Expected non-zero similarity for %q and %q", tt.a, tt.b)
		}
		if !tt.expected && sim > 0.5 {
			t.Errorf("Expected low similarity for %q and %q, got %f", tt.a, tt.b, sim)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s        string
		substrs  []string
		expected bool
	}{
		{"hello world", []string{"hello"}, true},
		{"hello world", []string{"foo", "world"}, true},
		{"hello world", []string{"foo", "bar"}, false},
		{"", []string{"anything"}, false},
	}

	for _, tt := range tests {
		result := contains(tt.s, tt.substrs...)
		if result != tt.expected {
			t.Errorf("contains(%q, %v) = %v, want %v", tt.s, tt.substrs, result, tt.expected)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s        string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
	}

	for _, tt := range tests {
		result := truncate(tt.s, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.maxLen, result, tt.expected)
		}
	}
}

func TestMinMax(t *testing.T) {
	if min(1.0, 2.0) != 1.0 {
		t.Error("min(1.0, 2.0) should be 1.0")
	}
	if max(1.0, 2.0) != 2.0 {
		t.Error("max(1.0, 2.0) should be 2.0")
	}
}

func TestEchoIntegrationBasic(t *testing.T) {
	config := DefaultConfig()
	config.EnableEchobeats = false
	config.EnableConsciousness = false
	config.EnableDreaming = false

	api := NewMultiAPIProvider(config)
	memory := NewUnifiedMemoryStore(config)
	persistence := NewJSONPersistence(config)

	echo := NewEchoIntegration(config, api, memory, persistence)

	if echo == nil {
		t.Fatal("NewEchoIntegration returned nil")
	}

	state, err := echo.GetState()
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}

	if state.Mode != "Idle" {
		t.Errorf("Initial mode should be Idle, got %s", state.Mode)
	}

	metrics := echo.GetMetrics()
	if metrics["running"].(bool) {
		t.Error("Echo should not be running initially")
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal_key", "normal_key"},
		{"key/with/slashes", "key_with_slashes"},
		{"key:with:colons", "key_with_colons"},
		{"key*with*stars", "key_with_stars"},
	}

	for _, tt := range tests {
		result := sanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// Benchmark tests

func BenchmarkMemoryStore(b *testing.B) {
	config := DefaultConfig()
	memory := NewUnifiedMemoryStore(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry := &MemoryEntry{
			Type:       MemoryTypeEpisodic,
			Content:    "Benchmark memory content",
			Importance: 0.5,
		}
		memory.Store(ctx, entry)
	}
}

func BenchmarkMemoryRetrieve(b *testing.B) {
	config := DefaultConfig()
	memory := NewUnifiedMemoryStore(config)
	ctx := context.Background()

	// Pre-populate with memories
	for i := 0; i < 1000; i++ {
		entry := &MemoryEntry{
			Type:       MemoryTypeEpisodic,
			Content:    "Memory content for benchmarking retrieval performance",
			Importance: 0.5,
		}
		memory.Store(ctx, entry)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memory.Retrieve(ctx, "benchmark", 10)
	}
}

func BenchmarkSimilarity(b *testing.B) {
	a := "This is a test string for benchmarking similarity calculations"
	s := "This is another string that is somewhat similar to the first one"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity(a, s)
	}
}
