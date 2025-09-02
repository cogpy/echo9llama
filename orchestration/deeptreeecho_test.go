package orchestration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/EchoCog/echollama/api"
)

func TestNewDeepTreeEcho(t *testing.T) {
	dte := NewDeepTreeEcho("Test DTE System")
	
	if dte == nil {
		t.Fatal("NewDeepTreeEcho should return a non-nil system")
	}
	
	if dte.Name != "Test DTE System" {
		t.Errorf("Expected name 'Test DTE System', got '%s'", dte.Name)
	}
	
	if dte.SystemHealth != SystemHealthInactive {
		t.Errorf("Expected initial system health to be inactive, got %s", dte.SystemHealth)
	}
	
	if dte.CoreStatus != CoreStatusInactive {
		t.Errorf("Expected initial core status to be inactive, got %s", dte.CoreStatus)
	}
	
	if dte.IdentityCoherence == nil {
		t.Error("IdentityCoherence should be initialized")
	}
	
	if dte.MemoryResonance == nil {
		t.Error("MemoryResonance should be initialized")
	}
	
	if dte.EchoPatterns == nil {
		t.Error("EchoPatterns should be initialized")
	}
	
	if dte.EvolutionTimeline == nil {
		t.Error("EvolutionTimeline should be initialized")
	}
	
	if len(dte.Integrations) == 0 {
		t.Error("Integrations should be initialized")
	}
}

func TestDeepTreeEchoInitialization(t *testing.T) {
	dte := NewDeepTreeEcho("Test DTE System")
	ctx := context.Background()
	
	err := dte.InitializeDTECore(ctx)
	if err != nil {
		t.Fatalf("InitializeDTECore failed: %v", err)
	}
	
	if dte.CoreStatus != CoreStatusActive {
		t.Errorf("Expected core status to be active after initialization, got %s", dte.CoreStatus)
	}
	
	if dte.SystemHealth != SystemHealthStable {
		t.Errorf("Expected system health to be stable after initialization, got %s", dte.SystemHealth)
	}
	
	// Check that identity coherence was initialized
	if dte.IdentityCoherence.OverallCoherence <= 0 {
		t.Error("Identity coherence should be initialized to a positive value")
	}
	
	// Check that memory resonance was initialized
	if dte.MemoryResonance.MemoryNodes <= 0 {
		t.Error("Memory nodes should be initialized to a positive value")
	}
	
	// Check that echo patterns were initialized
	if dte.EchoPatterns.RecursiveSelfImprovement.Strength <= 0 {
		t.Error("Echo patterns should be initialized with positive strength values")
	}
}

func TestDeepTreeEchoDiagnostics(t *testing.T) {
	dte := NewDeepTreeEcho("Test DTE System")
	ctx := context.Background()
	
	// Initialize first
	err := dte.InitializeDTECore(ctx)
	if err != nil {
		t.Fatalf("InitializeDTECore failed: %v", err)
	}
	
	// Run diagnostics
	result, err := dte.RunDiagnostics(ctx)
	if err != nil {
		t.Fatalf("RunDiagnostics failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("Diagnostic result should not be nil")
	}
	
	if len(result.Tests) == 0 {
		t.Error("Diagnostic result should contain tests")
	}
	
	// Check that all expected tests are present
	expectedTests := []string{"Cognitive Architecture", "Memory Resonance", "Echo Patterns", "Integrations"}
	foundTests := make(map[string]bool)
	
	for _, test := range result.Tests {
		foundTests[test.Name] = true
	}
	
	for _, expectedTest := range expectedTests {
		if !foundTests[expectedTest] {
			t.Errorf("Expected test '%s' not found in diagnostic results", expectedTest)
		}
	}
	
	if result.OverallHealth == "" {
		t.Error("Overall health should be set in diagnostic result")
	}
}

func TestDeepTreeEchoStatusRefresh(t *testing.T) {
	dte := NewDeepTreeEcho("Test DTE System")
	ctx := context.Background()
	
	// Initialize first
	err := dte.InitializeDTECore(ctx)
	if err != nil {
		t.Fatalf("InitializeDTECore failed: %v", err)
	}
	
	initialThoughtCount := dte.ThoughtCount
	initialUpdatedAt := dte.UpdatedAt
	
	// Wait a small amount to ensure time difference
	time.Sleep(time.Millisecond)
	
	// Refresh status
	err = dte.RefreshStatus(ctx)
	if err != nil {
		t.Fatalf("RefreshStatus failed: %v", err)
	}
	
	if dte.ThoughtCount <= initialThoughtCount {
		t.Error("Thought count should increase after status refresh")
	}
	
	if !dte.UpdatedAt.After(initialUpdatedAt) {
		t.Error("UpdatedAt should be updated after status refresh")
	}
}

func TestDeepTreeEchoRecursiveIntrospection(t *testing.T) {
	dte := NewDeepTreeEcho("Test DTE System")
	ctx := context.Background()
	
	// Initialize first
	err := dte.InitializeDTECore(ctx)
	if err != nil {
		t.Fatalf("InitializeDTECore failed: %v", err)
	}
	
	// Perform introspection
	result, err := dte.PerformRecursiveIntrospection(ctx, ".", 0.6, 0.4)
	if err != nil {
		t.Fatalf("PerformRecursiveIntrospection failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("Introspection result should not be nil")
	}
	
	if result.CognitiveSnapshot == nil {
		t.Error("Cognitive snapshot should not be nil")
	}
	
	if result.HypergraphPrompt == "" {
		t.Error("Hypergraph prompt should not be empty")
	}
	
	if result.EchoIntegration == nil {
		t.Error("Echo integration should not be nil")
	}
}

func TestSemanticSalienceAssessor(t *testing.T) {
	assessor := NewSemanticSalienceAssessor()
	
	testCases := []struct {
		filePath        string
		expectedMin     float64
		expectedMax     float64
		description     string
	}{
		{"btree-psi.scm", 0.98, 0.98, "highest priority file"},
		{"eva-model.py", 0.95, 0.95, "eva-model file"},
		{"echoself.md", 0.95, 0.95, "echoself documentation"},
		{"src/main.go", 0.85, 0.85, "source file"},
		{"test_something.py", 0.60, 0.70, "test file with .py extension"},
		{"random_file.txt", 0.3, 0.3, "unmatched file"},
	}
	
	for _, tc := range testCases {
		salience := assessor.AssessSalience(tc.filePath)
		if salience < tc.expectedMin || salience > tc.expectedMax {
			t.Errorf("For %s (%s): expected salience between %.2f and %.2f, got %.2f", 
				tc.filePath, tc.description, tc.expectedMin, tc.expectedMax, salience)
		}
	}
}

func TestAdaptiveAttentionAllocator(t *testing.T) {
	allocator := NewAdaptiveAttentionAllocator()
	
	testCases := []struct {
		cognitiveLoad   float64
		recentActivity  float64
		expectedMin     float64
		expectedMax     float64
		description     string
	}{
		{0.0, 0.0, 0.5, 0.7, "low load, no activity"},
		{1.0, 0.0, 0.8, 1.0, "high load, no activity"},
		{0.5, 0.5, 0.25, 0.65, "medium load, medium activity"},
		{0.0, 1.0, 0.0, 0.5, "low load, high activity"},
	}
	
	for _, tc := range testCases {
		threshold := allocator.ComputeAttentionThreshold(tc.cognitiveLoad, tc.recentActivity)
		if threshold < tc.expectedMin || threshold > tc.expectedMax {
			t.Errorf("For %s (load=%.1f, activity=%.1f): expected threshold between %.2f and %.2f, got %.2f", 
				tc.description, tc.cognitiveLoad, tc.recentActivity, tc.expectedMin, tc.expectedMax, threshold)
		}
		
		// Ensure threshold is bounded
		if threshold < 0.0 || threshold > 1.0 {
			t.Errorf("Threshold should be bounded between 0.0 and 1.0, got %.2f", threshold)
		}
	}
}

func TestRepositoryIntrospector(t *testing.T) {
	introspector := NewRepositoryIntrospector(".")
	
	if introspector == nil {
		t.Fatal("NewRepositoryIntrospector should return a non-nil introspector")
	}
	
	if introspector.assessor == nil {
		t.Error("Introspector should have an assessor")
	}
	
	if introspector.attentionAllocator == nil {
		t.Error("Introspector should have an attention allocator")
	}
	
	// Test analysis
	snapshot, err := introspector.AnalyzeRepository(0.6, 0.4)
	if err != nil {
		t.Fatalf("AnalyzeRepository failed: %v", err)
	}
	
	if snapshot == nil {
		t.Fatal("Cognitive snapshot should not be nil")
	}
	
	if snapshot.ProcessedFiles < 0 {
		t.Error("Processed files count should not be negative")
	}
	
	if snapshot.AttentionThreshold < 0.0 || snapshot.AttentionThreshold > 1.0 {
		t.Errorf("Attention threshold should be between 0.0 and 1.0, got %.2f", snapshot.AttentionThreshold)
	}
}

func TestEchoselfIntrospector(t *testing.T) {
	introspector := NewEchoselfIntrospector(".")
	
	if introspector == nil {
		t.Fatal("NewEchoselfIntrospector should return a non-nil introspector")
	}
	
	if introspector.repositoryIntrospector == nil {
		t.Error("Introspector should have a repository introspector")
	}
	
	if introspector.hypergraphNodes == nil {
		t.Error("Introspector should have initialized hypergraph nodes map")
	}
	
	// Test cognitive snapshot
	snapshot, err := introspector.GetCognitiveSnapshot(0.6, 0.4)
	if err != nil {
		t.Fatalf("GetCognitiveSnapshot failed: %v", err)
	}
	
	if snapshot == nil {
		t.Fatal("Cognitive snapshot should not be nil")
	}
	
	// Test hypergraph prompt generation
	prompt := introspector.InjectRepoInputIntoPrompt(snapshot)
	if prompt == "" {
		t.Error("Hypergraph prompt should not be empty")
	}
	
	// Check that prompt contains expected sections
	expectedSections := []string{"Hypergraph-Encoded", "Cognitive snapshot", "Salient Nodes", "Neural-Symbolic"}
	for _, section := range expectedSections {
		if !strings.Contains(prompt, section) {
			t.Errorf("Prompt should contain section '%s'", section)
		}
	}
}

func TestEngineDeepTreeEchoIntegration(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()
	
	// Test that engine has Deep Tree Echo system
	dte := engine.GetDeepTreeEcho()
	if dte == nil {
		t.Fatal("Engine should have a Deep Tree Echo system")
	}
	
	// Test initialization
	err := engine.InitializeDeepTreeEcho(ctx)
	if err != nil {
		t.Fatalf("InitializeDeepTreeEcho failed: %v", err)
	}
	
	// Test status retrieval
	status := engine.GetDeepTreeEchoStatus()
	if status == nil {
		t.Error("GetDeepTreeEchoStatus should return non-nil status")
	}
	
	// Test dashboard data
	dashboardData := engine.GetDeepTreeEchoDashboardData()
	if dashboardData == nil {
		t.Error("GetDeepTreeEchoDashboardData should return non-nil data")
	}
	
	// Check that dashboard data contains expected sections
	expectedSections := []string{"system_metrics", "integration_status", "identity_coherence", "memory_resonance", "echo_patterns", "evolution_timeline"}
	for _, section := range expectedSections {
		if _, exists := dashboardData[section]; !exists {
			t.Errorf("Dashboard data should contain section '%s'", section)
		}
	}
	
	// Test diagnostics
	diagnostics, err := engine.RunDeepTreeEchoDiagnostics(ctx)
	if err != nil {
		t.Fatalf("RunDeepTreeEchoDiagnostics failed: %v", err)
	}
	
	if diagnostics == nil {
		t.Error("Diagnostics result should not be nil")
	}
	
	// Test status refresh
	err = engine.RefreshDeepTreeEchoStatus(ctx)
	if err != nil {
		t.Fatalf("RefreshDeepTreeEchoStatus failed: %v", err)
	}
	
	// Test introspection
	introspectionResult, err := engine.PerformDeepTreeEchoIntrospection(ctx, ".", 0.6, 0.4)
	if err != nil {
		t.Fatalf("PerformDeepTreeEchoIntrospection failed: %v", err)
	}
	
	if introspectionResult == nil {
		t.Error("Introspection result should not be nil")
	}
}

func TestEvolutionTimelineProgression(t *testing.T) {
	dte := NewDeepTreeEcho("Test DTE System")
	ctx := context.Background()
	
	// Initialize system
	err := dte.InitializeDTECore(ctx)
	if err != nil {
		t.Fatalf("InitializeDTECore failed: %v", err)
	}
	
	// Initially should be in Foundation stage
	if dte.EvolutionTimeline.CurrentStage != "Foundation" {
		t.Errorf("Expected initial stage to be 'Foundation', got '%s'", dte.EvolutionTimeline.CurrentStage)
	}
	
	// Simulate progression by updating many times
	for i := 0; i < 200; i++ {
		dte.updateEvolutionTimeline()
	}
	
	// Should have progressed to next stage
	foundationStage := &dte.EvolutionTimeline.Stages[0]
	if foundationStage.Progress < 1.0 {
		// May not have completed yet, but should have made progress
		if foundationStage.Progress <= 0.0 {
			t.Error("Foundation stage should have made progress")
		}
	} else {
		// Should have completed and moved to Integration
		if foundationStage.Status != "complete" {
			t.Error("Foundation stage should be marked as complete")
		}
		
		if dte.EvolutionTimeline.CurrentStage != "Integration" {
			t.Errorf("Should have progressed to Integration stage, got '%s'", dte.EvolutionTimeline.CurrentStage)
		}
	}
}

func TestEchoPatternEvolution(t *testing.T) {
	dte := NewDeepTreeEcho("Test DTE System")
	ctx := context.Background()
	
	// Initialize system
	err := dte.InitializeDTECore(ctx)
	if err != nil {
		t.Fatalf("InitializeDTECore failed: %v", err)
	}
	
	initialStrength := dte.EchoPatterns.RecursiveSelfImprovement.Strength
	
	// Update patterns multiple times
	for i := 0; i < 100; i++ {
		dte.updateEchoPatterns()
	}
	
	finalStrength := dte.EchoPatterns.RecursiveSelfImprovement.Strength
	
	if finalStrength <= initialStrength {
		t.Error("Echo pattern strength should increase over time")
	}
	
	// Should not exceed maximum
	if finalStrength > 1.0 {
		t.Errorf("Echo pattern strength should not exceed 1.0, got %.3f", finalStrength)
	}
}