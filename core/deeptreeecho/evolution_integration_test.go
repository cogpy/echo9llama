package deeptreeecho

import (
	"context"
	"os"
	"testing"
)

func TestDefaultEvolutionSystemConfig(t *testing.T) {
	config := DefaultEvolutionSystemConfig()

	if len(config.PreferredProviders) != 3 {
		t.Errorf("Expected 3 preferred providers, got %d", len(config.PreferredProviders))
	}

	expectedProviders := []string{"anthropic", "openrouter", "openai"}
	for i, provider := range config.PreferredProviders {
		if provider != expectedProviders[i] {
			t.Errorf("Expected provider %s at index %d, got %s", expectedProviders[i], i, provider)
		}
	}

	if config.Debug {
		t.Error("Expected Debug to be false by default")
	}
}

func TestEvolutionSystemWithoutProviders(t *testing.T) {
	// Save and clear environment variables
	// t.Setenv restores the original values when the test ends.
	for _, key := range []string{"ANTHROPIC_API_KEY", "OPENROUTER_API_KEY", "OPENAI_API_KEY"} {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}

	config := DefaultEvolutionSystemConfig()
	_, err := NewEvolutionSystem(config)

	if err == nil {
		t.Error("Expected error when no providers are available")
	}
}

func TestEvolutionSystemStatus(t *testing.T) {
	// This test only runs if at least one API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" &&
		os.Getenv("OPENROUTER_API_KEY") == "" &&
		os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("No API keys set, skipping integration test")
	}

	config := DefaultEvolutionSystemConfig()
	config.Debug = false

	es, err := NewEvolutionSystem(config)
	if err != nil {
		t.Fatalf("Failed to create evolution system: %v", err)
	}

	status := es.GetStatus()

	if !status["initialized"].(bool) {
		t.Error("Expected system to be initialized")
	}

	if status["running"].(bool) {
		t.Error("Expected system to not be running initially")
	}

	providers, ok := status["providers"].([]string)
	if !ok || len(providers) == 0 {
		t.Error("Expected at least one provider")
	}
}

func TestEvolutionSystemDiagnostics(t *testing.T) {
	// This test only runs if at least one API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" &&
		os.Getenv("OPENROUTER_API_KEY") == "" &&
		os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("No API keys set, skipping integration test")
	}

	config := DefaultEvolutionSystemConfig()
	es, err := NewEvolutionSystem(config)
	if err != nil {
		t.Fatalf("Failed to create evolution system: %v", err)
	}

	ctx := context.Background()
	diag, err := es.RunDiagnostics(ctx)
	if err != nil {
		t.Fatalf("Failed to run diagnostics: %v", err)
	}

	if len(diag.Tests) == 0 {
		t.Error("Expected at least one diagnostic test")
	}

	if diag.OverallHealth == "" {
		t.Error("Expected overall health to be set")
	}

	// Check specific tests exist
	testNames := make(map[string]bool)
	for _, test := range diag.Tests {
		testNames[test.Name] = true
	}

	expectedTests := []string{
		"LLM Provider",
		"Stream of Consciousness",
		"Echobeats Scheduler",
		"Echodream Integration",
		"Genetic Traits",
	}

	for _, name := range expectedTests {
		if !testNames[name] {
			t.Errorf("Expected diagnostic test '%s' not found", name)
		}
	}
}

func TestEvolutionSystemGoals(t *testing.T) {
	// This test only runs if at least one API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" &&
		os.Getenv("OPENROUTER_API_KEY") == "" &&
		os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("No API keys set, skipping integration test")
	}

	config := DefaultEvolutionSystemConfig()
	es, err := NewEvolutionSystem(config)
	if err != nil {
		t.Fatalf("Failed to create evolution system: %v", err)
	}

	goalID := es.AddGoal("Test evolutionary goal", 0.8)
	if goalID == "" {
		t.Error("Expected non-empty goal ID")
	}

	metrics := es.GetSchedulerMetrics()
	if metrics == nil {
		t.Error("Expected non-nil scheduler metrics")
	}
}

func TestEvolutionSystemMetrics(t *testing.T) {
	// This test only runs if at least one API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" &&
		os.Getenv("OPENROUTER_API_KEY") == "" &&
		os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("No API keys set, skipping integration test")
	}

	config := DefaultEvolutionSystemConfig()
	es, err := NewEvolutionSystem(config)
	if err != nil {
		t.Fatalf("Failed to create evolution system: %v", err)
	}

	// Test consciousness metrics
	consMetrics := es.GetConsciousnessMetrics()
	if consMetrics == nil {
		t.Error("Expected non-nil consciousness metrics")
	}

	// Test scheduler metrics
	schedMetrics := es.GetSchedulerMetrics()
	if schedMetrics == nil {
		t.Error("Expected non-nil scheduler metrics")
	}

	// Test dream metrics
	dreamMetrics := es.GetDreamMetrics()
	if dreamMetrics == nil {
		t.Error("Expected non-nil dream metrics")
	}

	// Test evolution history
	history := es.GetEvolutionHistory()
	if history == nil {
		t.Error("Expected non-nil evolution history")
	}
}
