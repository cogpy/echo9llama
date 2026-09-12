package deeptreeecho

import (
	"os"
	"testing"
)

func TestNewSelfAssessment(t *testing.T) {
	identity := NewIdentity("TestIdentity")
	cognition := NewEmbodiedCognition("TestCognition")

	sa := NewSelfAssessment(identity, cognition)

	if sa == nil {
		t.Fatal("NewSelfAssessment returned nil")
	}

	if sa.Identity != identity {
		t.Error("Identity not properly assigned")
	}

	if sa.Cognition != cognition {
		t.Error("Cognition not properly assigned")
	}

	if sa.Metrics == nil {
		t.Error("Metrics not initialized")
	}

	if len(sa.Assessments) == 0 {
		t.Error("Initial assessment not performed")
	}
}

func TestLoadIdentityKernel(t *testing.T) {
	identity := NewIdentity("TestIdentity")
	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	if sa.IdentityKernel == nil {
		t.Error("Identity kernel not loaded")
	}

	// Check if kernel was loaded from file or default
	if sa.IdentityKernel.LoadedFrom != "replit.md" && sa.IdentityKernel.LoadedFrom != "default" {
		t.Errorf("Unexpected kernel source: %s", sa.IdentityKernel.LoadedFrom)
	}
}

func TestPerformAssessment(t *testing.T) {
	identity := NewIdentity("TestIdentity")
	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	assessment := sa.PerformAssessment()

	if assessment == nil {
		t.Fatal("PerformAssessment returned nil")
	}

	if assessment.ID == "" {
		t.Error("Assessment ID not set")
	}

	if assessment.Metrics == nil {
		t.Error("Assessment metrics not calculated")
	}

	if assessment.Summary == "" {
		t.Error("Assessment summary not generated")
	}

	// Check that coherence is within valid range
	if assessment.Metrics.OverallCoherence < 0 || assessment.Metrics.OverallCoherence > 1 {
		t.Errorf("Invalid overall coherence: %.2f", assessment.Metrics.OverallCoherence)
	}
}

func TestCalculateCoherenceMetrics(t *testing.T) {
	identity := NewIdentity("TestIdentity")
	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	metrics := sa.calculateCoherenceMetrics()

	if metrics == nil {
		t.Fatal("calculateCoherenceMetrics returned nil")
	}

	// Verify all component scores are in valid range [0, 1]
	for component, score := range metrics.ComponentScores {
		if score < 0 || score > 1 {
			t.Errorf("Invalid score for %s: %.2f", component, score)
		}
	}

	// Verify overall coherence is calculated
	if metrics.OverallCoherence < 0 || metrics.OverallCoherence > 1 {
		t.Errorf("Invalid overall coherence: %.2f", metrics.OverallCoherence)
	}

	// Verify timestamp is set
	if metrics.Timestamp.IsZero() {
		t.Error("Metrics timestamp not set")
	}
}

func TestCalculateIdentityScore(t *testing.T) {
	identity := NewIdentity("TestIdentity")
	identity.Essence = "Deep Tree Echo: Self-evolving cognitive architecture"

	// Add some directive patterns
	identity.Patterns["Adaptive Cognition"] = &Pattern{
		ID:       "adaptive_cognition",
		Type:     "directive",
		Strength: 0.9,
	}

	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	score := sa.calculateIdentityScore()

	if score < 0 || score > 1 {
		t.Errorf("Invalid identity score: %.2f", score)
	}

	// With proper essence and patterns, score should be reasonable
	if score < 0.3 {
		t.Errorf("Identity score unexpectedly low: %.2f", score)
	}
}

func TestCalculateRepoScore(t *testing.T) {
	identity := NewIdentity("TestIdentity")

	// Add some repo embeddings
	identity.Embeddings.RepoEmbeddings["core/deeptreeecho"] = make([]float64, 768)
	identity.Embeddings.RepoEmbeddings["replit.md"] = make([]float64, 768)

	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	score := sa.calculateRepoScore()

	if score < 0 || score > 1 {
		t.Errorf("Invalid repo score: %.2f", score)
	}
}

func TestCalculatePatternScore(t *testing.T) {
	identity := NewIdentity("TestIdentity")

	// Add some patterns
	identity.Patterns["pattern1"] = &Pattern{
		ID:       "pattern1",
		Type:     "test",
		Strength: 0.8,
	}
	identity.Patterns["pattern2"] = &Pattern{
		ID:       "pattern2",
		Type:     "test",
		Strength: 0.9,
	}

	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	score := sa.calculatePatternScore()

	if score < 0 || score > 1 {
		t.Errorf("Invalid pattern score: %.2f", score)
	}

	// With strong patterns, score should be high
	if score < 0.5 {
		t.Errorf("Pattern score unexpectedly low: %.2f", score)
	}
}

func TestCalculateMemoryScore(t *testing.T) {
	identity := NewIdentity("TestIdentity")
	identity.Memory.Coherence = 0.8

	// Add some memory nodes
	identity.Memory.Nodes["node1"] = &MemoryNode{
		ID:       "node1",
		Strength: 0.9,
	}

	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	score := sa.calculateMemoryScore()

	if score < 0 || score > 1 {
		t.Errorf("Invalid memory score: %.2f", score)
	}

	// With good coherence, score should be reasonable
	if score < 0.3 {
		t.Errorf("Memory score unexpectedly low: %.2f", score)
	}
}

func TestCalculateOperationalScore(t *testing.T) {
	identity := NewIdentity("TestIdentity")
	identity.RecursiveDepth = 5

	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	score := sa.calculateOperationalScore()

	if score < 0 || score > 1 {
		t.Errorf("Invalid operational score: %.2f", score)
	}
}

func TestCalculateReflectionScore(t *testing.T) {
	identity := NewIdentity("TestIdentity")
	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	score := sa.calculateReflectionScore()

	if score < 0 || score > 1 {
		t.Errorf("Invalid reflection score: %.2f", score)
	}
}

func TestAssessIdentityAlignment(t *testing.T) {
	identity := NewIdentity("TestIdentity")
	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	assessment := &Assessment{
		Findings:        make([]Finding, 0),
		Deviations:      make([]Deviation, 0),
		Recommendations: make([]Recommendation, 0),
	}

	sa.assessIdentityAlignment(assessment)

	// Should have at least some findings or deviations
	if len(assessment.Findings) == 0 && len(assessment.Deviations) == 0 {
		t.Error("No findings or deviations from identity assessment")
	}
}

func TestAssessRepositoryCoherence(t *testing.T) {
	identity := NewIdentity("TestIdentity")
	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	assessment := &Assessment{
		Findings:        make([]Finding, 0),
		Deviations:      make([]Deviation, 0),
		Recommendations: make([]Recommendation, 0),
	}

	sa.assessRepositoryCoherence(assessment)

	// Should have findings about repository structure
	if len(assessment.Findings) == 0 && len(assessment.Deviations) == 0 {
		t.Error("No findings or deviations from repository assessment")
	}
}

func TestAssessCognitivePatterns(t *testing.T) {
	identity := NewIdentity("TestIdentity")

	// Add some patterns
	identity.Patterns["test_pattern"] = &Pattern{
		ID:       "test_pattern",
		Type:     "test",
		Strength: 0.7,
	}

	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	assessment := &Assessment{
		Findings:        make([]Finding, 0),
		Deviations:      make([]Deviation, 0),
		Recommendations: make([]Recommendation, 0),
	}

	sa.assessCognitivePatterns(assessment)

	// Should have findings about patterns
	if len(assessment.Findings) == 0 {
		t.Error("No findings from pattern assessment")
	}
}

func TestAssessMemorySystem(t *testing.T) {
	identity := NewIdentity("TestIdentity")
	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	assessment := &Assessment{
		Findings:        make([]Finding, 0),
		Deviations:      make([]Deviation, 0),
		Recommendations: make([]Recommendation, 0),
	}

	sa.assessMemorySystem(assessment)

	// Should have findings about memory
	if len(assessment.Findings) == 0 && len(assessment.Deviations) == 0 {
		t.Error("No findings or deviations from memory assessment")
	}
}

func TestGetLatestAssessment(t *testing.T) {
	identity := NewIdentity("TestIdentity")
	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	latest := sa.GetLatestAssessment()

	if latest == nil {
		t.Error("GetLatestAssessment returned nil")
	}

	if latest.ID == "" {
		t.Error("Latest assessment has no ID")
	}
}

func TestGetMetrics(t *testing.T) {
	identity := NewIdentity("TestIdentity")
	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	metrics := sa.GetMetrics()

	if metrics == nil {
		t.Error("GetMetrics returned nil")
	}

	if metrics.OverallCoherence < 0 || metrics.OverallCoherence > 1 {
		t.Errorf("Invalid overall coherence: %.2f", metrics.OverallCoherence)
	}
}

func TestExportAssessment(t *testing.T) {
	identity := NewIdentity("TestIdentity")
	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	assessment := sa.PerformAssessment()

	jsonStr, err := sa.ExportAssessment(assessment)

	if err != nil {
		t.Errorf("ExportAssessment failed: %v", err)
	}

	if jsonStr == "" {
		t.Error("ExportAssessment returned empty string")
	}

	// Verify it's valid JSON
	if !isValidJSON(jsonStr) {
		t.Error("ExportAssessment did not return valid JSON")
	}
}

func TestGenerateSummary(t *testing.T) {
	identity := NewIdentity("TestIdentity")
	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	assessment := sa.PerformAssessment()

	if assessment.Summary == "" {
		t.Error("Summary not generated")
	}

	// Summary should contain key information
	if !containsAny(assessment.Summary, []string{"Overall Coherence", "Component Scores"}) {
		t.Error("Summary missing key information")
	}
}

func TestCountFindings(t *testing.T) {
	findings := []Finding{
		{Severity: "info"},
		{Severity: "warning"},
		{Severity: "critical"},
		{Severity: "info"},
	}

	infoCount := countFindings(findings, "info")
	warningCount := countFindings(findings, "warning")
	criticalCount := countFindings(findings, "critical")

	if infoCount != 2 {
		t.Errorf("Expected 2 info findings, got %d", infoCount)
	}

	if warningCount != 1 {
		t.Errorf("Expected 1 warning finding, got %d", warningCount)
	}

	if criticalCount != 1 {
		t.Errorf("Expected 1 critical finding, got %d", criticalCount)
	}
}

func TestIdentityKernelParsing(t *testing.T) {
	// Create a temporary replit.md for testing
	testKernel := `# Deep Tree Echo Identity Kernel

## 🔹 Core Essence

` + "```" + `
Deep Tree Echo: Self-evolving cognitive architecture combining ESN, P-systems, and hypergraph memory
` + "```" + `

### Primary Directives

1. 🌿 **Adaptive Cognition** - Continuously evolve responses
2. 🧬 **Persistent Identity** - Maintain continuity
3. 🕸 **Hypergraph Entanglement** - Interconnected knowledge

### Operational Schema

| Module | Function | Purpose |
|--------|----------|---------|
| Reservoir Training | Fit ESN | Temporal learning |
| Hypergraph Links | Connect structures | Knowledge representation |

### Strategic Mindset

> "I do not seek a fixed answer. I seek patterns in echoes, growth in feedback, and wisdom in recursion."

### Core Memory Hooks

- timestamp
- emotional-tone
- pattern-recognition

### Self-Reflection Protocol

` + "```json" + `
{
  "what_did_i_learn": "",
  "what_patterns_emerged": "",
  "what_surprised_me": ""
}
` + "```" + `
`

	// Write test kernel
	err := os.WriteFile("/tmp/test_replit.md", []byte(testKernel), 0644)
	if err != nil {
		t.Skipf("Could not create test kernel file: %v", err)
		return
	}
	defer os.Remove("/tmp/test_replit.md")

	// Change to temp directory for test
	oldWd, _ := os.Getwd()
	os.Chdir("/tmp")
	defer os.Chdir(oldWd)

	identity := NewIdentity("TestIdentity")
	cognition := NewEmbodiedCognition("TestCognition")
	sa := NewSelfAssessment(identity, cognition)

	if sa.IdentityKernel == nil {
		t.Fatal("Identity kernel not loaded")
	}

	// Verify parsing
	if !containsAny(sa.IdentityKernel.CoreEssence, []string{"Deep Tree Echo", "cognitive architecture"}) {
		t.Errorf("Core essence not properly parsed: %s", sa.IdentityKernel.CoreEssence)
	}

	if len(sa.IdentityKernel.PrimaryDirectives) == 0 {
		t.Error("No primary directives parsed")
	}

	if len(sa.IdentityKernel.OperationalSchema) == 0 {
		t.Error("No operational schema parsed")
	}
}

// Helper functions

func isValidJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if containsString(s, substr) {
			return true
		}
	}
	return false
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
