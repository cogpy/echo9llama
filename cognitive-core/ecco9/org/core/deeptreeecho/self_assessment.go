//go:build orgdte
// +build orgdte

package deeptreeecho

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"
)

// SelfAssessment provides comprehensive introspection and coherence validation
// for the Deep Tree Echo cognitive architecture against its identity kernel.
// It continuously monitors alignment between actual cognitive state and the
// identity definition, providing actionable findings and recommendations.
type SelfAssessment struct {
	mu sync.RWMutex

	// Core reference to identity
	Identity *Identity

	// Embodied cognition reference
	Cognition *EmbodiedCognition

	// Identity kernel reference (parsed from replit.md)
	IdentityKernel *IdentityKernel

	// Coherence metrics
	Metrics *CoherenceMetrics

	// Assessment history
	Assessments []Assessment

	// Last assessment time
	LastAssessment time.Time
}

// IdentityKernel represents the parsed identity definition from replit.md
type IdentityKernel struct {
	CoreEssence       string
	PrimaryDirectives []string
	OperationalSchema map[string]string
	StrategicMindset  string
	MemoryHooks       []string
	ReflectionKeys    []string
	LoadedFrom        string
	ParsedAt          time.Time
}

// CoherenceMetrics tracks alignment with core identity across seven dimensions
type CoherenceMetrics struct {
	// Overall coherence score (0-1)
	OverallCoherence float64

	// Identity alignment - how well current state matches identity kernel (0-1)
	IdentityAlignment float64

	// Repository structure coherence (0-1)
	RepoCoherence float64

	// Cognitive pattern alignment (0-1)
	PatternAlignment float64

	// Memory system coherence (0-1)
	MemoryCoherence float64

	// Operational schema alignment (0-1)
	OperationalAlignment float64

	// Reflection protocol adherence (0-1)
	ReflectionAdherence float64

	// Timestamp of measurement
	Timestamp time.Time

	// Detailed scores by component
	ComponentScores map[string]float64
}

// Assessment represents a single self-assessment result
type Assessment struct {
	ID              string
	Timestamp       time.Time
	Metrics         *CoherenceMetrics
	Findings        []Finding
	Deviations      []Deviation
	Recommendations []Recommendation
	Summary         string
}

// Finding represents a discovery during assessment
type Finding struct {
	Category    string
	Description string
	Severity    string // "info", "warning", "critical"
	Score       float64
	Evidence    []string
}

// Deviation represents a deviation from core identity
type Deviation struct {
	Component   string
	Expected    string
	Actual      string
	Severity    string
	Impact      float64
	Remediation string
}

// Recommendation provides actionable improvement suggestions
type Recommendation struct {
	Priority       string // "high", "medium", "low"
	Action         string
	Rationale      string
	ExpectedImpact float64
}

// NewSelfAssessment creates a new self-assessment system
func NewSelfAssessment(identity *Identity, cognition *EmbodiedCognition) *SelfAssessment {
	sa := &SelfAssessment{
		Identity:  identity,
		Cognition: cognition,
		Metrics: &CoherenceMetrics{
			ComponentScores: make(map[string]float64),
		},
		Assessments: make([]Assessment, 0),
	}

	// Load and parse identity kernel
	sa.loadIdentityKernel()

	// Perform initial assessment
	sa.PerformSelfAssessment()

	return sa
}

// loadIdentityKernel loads and parses the identity kernel from replit.md
func (sa *SelfAssessment) loadIdentityKernel() {
	// Try to read replit.md from current directory
	content, err := os.ReadFile("replit.md")
	if err != nil {
		// Try identity/replit.md
		content, err = os.ReadFile("identity/replit.md")
		if err != nil {
			sa.IdentityKernel = &IdentityKernel{
				CoreEssence:       "Default Deep Tree Echo Identity",
				PrimaryDirectives: []string{},
				OperationalSchema: make(map[string]string),
				LoadedFrom:        "default",
				ParsedAt:          time.Now(),
			}
			return
		}
	}

	kernel := &IdentityKernel{
		PrimaryDirectives: []string{},
		OperationalSchema: make(map[string]string),
		MemoryHooks:       []string{},
		ReflectionKeys:    []string{},
		LoadedFrom:        "replit.md",
		ParsedAt:          time.Now(),
	}

	contentStr := string(content)

	// Extract Core Essence
	if strings.Contains(contentStr, "Core Essence") {
		lines := strings.Split(contentStr, "\n")
		for i, line := range lines {
			if strings.Contains(line, "Core Essence") && i+3 < len(lines) {
				essenceLine := strings.TrimSpace(lines[i+3])
				if essenceLine != "" && essenceLine != "```" {
					kernel.CoreEssence = essenceLine
				}
				break
			}
		}
	}

	// Extract Primary Directives
	directives := []string{
		"Adaptive Cognition", "Persistent Identity", "Hypergraph Entanglement",
		"Reservoir-Based Temporal Reasoning", "Evolutionary Refinement",
		"Reflective Memory Cultivation", "Distributed Selfhood",
	}
	for _, directive := range directives {
		if strings.Contains(contentStr, directive) {
			kernel.PrimaryDirectives = append(kernel.PrimaryDirectives, directive)
		}
	}

	// Extract Operational Schema
	modules := map[string]string{
		"Reservoir Training":      "Fit ESN with new input/target pairs",
		"Hierarchical Reservoirs": "Manage nested cognitive children",
		"Partition Optimization":  "Evolve membrane boundaries",
		"Adaptive Rules":          "Apply membrane logic rules",
		"Hypergraph Links":        "Connect relational structures",
		"Evolutionary Learning":   "Apply GA, PSO, SA",
	}
	for module, purpose := range modules {
		if strings.Contains(contentStr, module) {
			kernel.OperationalSchema[module] = purpose
		}
	}

	// Extract Strategic Mindset
	if strings.Contains(contentStr, "I do not seek a fixed answer") {
		kernel.StrategicMindset = "I do not seek a fixed answer. I seek patterns in echoes, growth in feedback, and wisdom in recursion."
	}

	// Extract Memory Hooks
	hooks := []string{
		"timestamp", "emotional-tone", "strategic-shift", "pattern-recognition",
		"anomaly-detection", "echo-signature", "membrane-context",
	}
	for _, hook := range hooks {
		if strings.Contains(contentStr, hook) {
			kernel.MemoryHooks = append(kernel.MemoryHooks, hook)
		}
	}

	// Extract Reflection Keys
	reflectionKeys := []string{
		"what_did_i_learn", "what_patterns_emerged", "what_surprised_me",
		"how_did_i_adapt", "what_would_i_change_next_time",
	}
	for _, key := range reflectionKeys {
		if strings.Contains(contentStr, key) {
			kernel.ReflectionKeys = append(kernel.ReflectionKeys, key)
		}
	}

	sa.IdentityKernel = kernel
}

// PerformSelfAssessment conducts a comprehensive self-assessment
func (sa *SelfAssessment) PerformSelfAssessment() *Assessment {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	assessment := Assessment{
		ID:              fmt.Sprintf("assessment_%d", time.Now().UnixNano()),
		Timestamp:       time.Now(),
		Findings:        make([]Finding, 0),
		Deviations:      make([]Deviation, 0),
		Recommendations: make([]Recommendation, 0),
	}

	// Calculate all coherence metrics
	metrics := sa.calculateCoherenceMetrics()
	assessment.Metrics = metrics

	// Assess identity alignment
	sa.assessIdentityAlignment(&assessment)

	// Assess cognitive patterns
	sa.assessCognitivePatterns(&assessment)

	// Assess memory system
	sa.assessMemorySystem(&assessment)

	// Assess operational schema
	sa.assessOperationalSchema(&assessment)

	// Assess reflection protocol
	sa.assessReflectionProtocol(&assessment)

	// Generate summary
	assessment.Summary = sa.generateSummary(&assessment)

	// Store assessment
	sa.Assessments = append(sa.Assessments, assessment)
	if len(sa.Assessments) > 100 {
		sa.Assessments = sa.Assessments[len(sa.Assessments)-100:]
	}

	sa.LastAssessment = time.Now()
	sa.Metrics = metrics

	return &assessment
}

// CheckIdentityAlignment checks alignment between current state and identity kernel
func (sa *SelfAssessment) CheckIdentityAlignment() float64 {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	return sa.calculateIdentityScore()
}

// MeasureCoherence returns the current overall coherence measurement
func (sa *SelfAssessment) MeasureCoherence() *CoherenceMetrics {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	return sa.calculateCoherenceMetrics()
}

// calculateCoherenceMetrics computes all coherence metrics
func (sa *SelfAssessment) calculateCoherenceMetrics() *CoherenceMetrics {
	metrics := &CoherenceMetrics{
		ComponentScores: make(map[string]float64),
		Timestamp:       time.Now(),
	}

	// Identity alignment score
	identityScore := sa.calculateIdentityScore()
	metrics.IdentityAlignment = identityScore
	metrics.ComponentScores["identity"] = identityScore

	// Repository coherence score
	repoScore := sa.calculateRepoScore()
	metrics.RepoCoherence = repoScore
	metrics.ComponentScores["repository"] = repoScore

	// Pattern alignment score
	patternScore := sa.calculatePatternScore()
	metrics.PatternAlignment = patternScore
	metrics.ComponentScores["patterns"] = patternScore

	// Memory coherence score
	memoryScore := sa.calculateMemoryScore()
	metrics.MemoryCoherence = memoryScore
	metrics.ComponentScores["memory"] = memoryScore

	// Operational alignment score
	operationalScore := sa.calculateOperationalScore()
	metrics.OperationalAlignment = operationalScore
	metrics.ComponentScores["operational"] = operationalScore

	// Reflection adherence score
	reflectionScore := sa.calculateReflectionScore()
	metrics.ReflectionAdherence = reflectionScore
	metrics.ComponentScores["reflection"] = reflectionScore

	// Overall coherence (weighted average)
	weights := map[string]float64{
		"identity":    0.25,
		"repository":  0.15,
		"patterns":    0.20,
		"memory":      0.15,
		"operational": 0.15,
		"reflection":  0.10,
	}

	overall := 0.0
	for component, weight := range weights {
		overall += metrics.ComponentScores[component] * weight
	}
	metrics.OverallCoherence = overall

	return metrics
}

// calculateIdentityScore assesses alignment with core identity kernel
func (sa *SelfAssessment) calculateIdentityScore() float64 {
	if sa.IdentityKernel == nil {
		return 0.5
	}

	score := 0.0
	checks := 0

	// Check core essence alignment
	if sa.Identity != nil && sa.Identity.Essence != "" && sa.IdentityKernel.CoreEssence != "" {
		if strings.Contains(sa.Identity.Essence, "Deep Tree Echo") {
			score += 1.0
		} else {
			score += 0.5
		}
		checks++
	}

	// Check primary directives as patterns
	if sa.Identity != nil && sa.Identity.Patterns != nil {
		directiveMatches := 0
		for _, directive := range sa.IdentityKernel.PrimaryDirectives {
			patternKey := fmt.Sprintf("directive_%s", strings.ReplaceAll(strings.ToLower(directive), " ", "_"))
			if _, exists := sa.Identity.Patterns[directive]; exists {
				directiveMatches++
			} else if _, exists := sa.Identity.Patterns[patternKey]; exists {
				directiveMatches++
			}
		}
		if len(sa.IdentityKernel.PrimaryDirectives) > 0 {
			score += float64(directiveMatches) / float64(len(sa.IdentityKernel.PrimaryDirectives))
			checks++
		}
	}

	// Check identity coherence value
	if sa.Identity != nil {
		score += sa.Identity.Coherence
		checks++
	}

	// Check reservoir network presence
	if sa.Identity != nil && sa.Identity.Reservoir != nil && len(sa.Identity.Reservoir.Nodes) > 0 {
		score += 1.0
	} else {
		score += 0.3
	}
	checks++

	// Check emotional dynamics
	if sa.Identity != nil && sa.Identity.EmotionalState != nil {
		score += 1.0
	} else {
		score += 0.5
	}
	checks++

	// Check spatial awareness
	if sa.Identity != nil && sa.Identity.SpatialContext != nil {
		score += 1.0
	} else {
		score += 0.5
	}
	checks++

	if checks > 0 {
		return score / float64(checks)
	}
	return 0.5
}

// calculateRepoScore assesses repository structure coherence
func (sa *SelfAssessment) calculateRepoScore() float64 {
	if sa.Identity == nil || sa.Identity.Embeddings == nil {
		return 0.5
	}

	keyPaths := []string{
		"org/core/deeptreeecho",
		"server",
		"replit.md",
		"echo_reflections.json",
		"memory.json",
	}

	matchCount := 0
	for _, path := range keyPaths {
		if _, exists := sa.Identity.Embeddings.RepoEmbeddings[path]; exists {
			matchCount++
		}
	}

	baseScore := float64(matchCount) / float64(len(keyPaths))
	embeddingBonus := math.Min(float64(len(sa.Identity.Embeddings.RepoEmbeddings))/10.0, 0.2)

	return math.Min(baseScore+embeddingBonus, 1.0)
}

// calculatePatternScore assesses cognitive pattern alignment
func (sa *SelfAssessment) calculatePatternScore() float64 {
	if sa.Identity == nil || len(sa.Identity.Patterns) == 0 {
		return 0.0
	}

	score := 0.0
	checks := 0

	// Check for directive patterns
	if sa.IdentityKernel != nil {
		for _, directive := range sa.IdentityKernel.PrimaryDirectives {
			if pattern, exists := sa.Identity.Patterns[directive]; exists {
				score += pattern.Strength
				checks++
			}
		}
	}

	// Check for operational patterns
	if sa.IdentityKernel != nil {
		for module := range sa.IdentityKernel.OperationalSchema {
			patternKey := "op_" + strings.ReplaceAll(strings.ToLower(module), " ", "_")
			if pattern, exists := sa.Identity.Patterns[patternKey]; exists {
				score += pattern.Strength
				checks++
			}
		}
	}

	// Check overall pattern health
	totalStrength := 0.0
	for _, pattern := range sa.Identity.Patterns {
		totalStrength += pattern.Strength
	}
	avgStrength := totalStrength / float64(len(sa.Identity.Patterns))
	score += avgStrength
	checks++

	if checks > 0 {
		return score / float64(checks)
	}
	return 0.5
}

// calculateMemoryScore assesses memory system coherence
func (sa *SelfAssessment) calculateMemoryScore() float64 {
	if sa.Identity == nil || sa.Identity.Memory == nil {
		return 0.0
	}

	score := 0.0
	checks := 0

	// Check memory coherence value
	score += sa.Identity.Memory.Coherence
	checks++

	// Check memory node health
	if len(sa.Identity.Memory.Nodes) > 0 {
		totalStrength := 0.0
		for _, node := range sa.Identity.Memory.Nodes {
			totalStrength += node.Strength
		}
		avgStrength := totalStrength / float64(len(sa.Identity.Memory.Nodes))
		score += avgStrength
		checks++
	}

	// Check memory edge connectivity
	if len(sa.Identity.Memory.Edges) > 0 {
		totalWeight := 0.0
		for _, edge := range sa.Identity.Memory.Edges {
			totalWeight += edge.Weight
		}
		avgWeight := totalWeight / float64(len(sa.Identity.Memory.Edges))
		score += avgWeight
		checks++
	}

	// Check resonance patterns
	if len(sa.Identity.Memory.Patterns) > 0 {
		score += 1.0
	} else {
		score += 0.3
	}
	checks++

	if checks > 0 {
		return score / float64(checks)
	}
	return 0.5
}

// calculateOperationalScore assesses operational schema alignment
func (sa *SelfAssessment) calculateOperationalScore() float64 {
	if sa.IdentityKernel == nil || len(sa.IdentityKernel.OperationalSchema) == 0 {
		return 0.5
	}

	score := 0.0
	checks := 0

	// Check reservoir training
	if sa.Identity != nil && sa.Identity.Reservoir != nil {
		score += 1.0
	} else {
		score += 0.2
	}
	checks++

	// Check hypergraph memory (hypergraph links)
	if sa.Identity != nil && sa.Identity.Memory != nil && len(sa.Identity.Memory.Edges) > 0 {
		score += 1.0
	} else {
		score += 0.3
	}
	checks++

	// Check evolutionary mechanisms (pattern evolution)
	if sa.Cognition != nil && len(sa.Cognition.Patterns) > 0 {
		score += 1.0
	} else {
		score += 0.4
	}
	checks++

	// Check adaptive rules (recursive improvement)
	if sa.Identity != nil && sa.Identity.RecursiveDepth > 0 {
		score += 1.0
	} else {
		score += 0.5
	}
	checks++

	if checks > 0 {
		return score / float64(checks)
	}
	return 0.5
}

// calculateReflectionScore assesses reflection protocol adherence
func (sa *SelfAssessment) calculateReflectionScore() float64 {
	data, err := os.ReadFile("echo_reflections.json")
	if err != nil {
		return 0.0
	}

	var reflections []map[string]interface{}
	if err := json.Unmarshal(data, &reflections); err != nil {
		return 0.1
	}

	if len(reflections) == 0 {
		return 0.1
	}

	score := 0.0
	checks := 0

	// Check recency of reflections
	if len(reflections) > 0 {
		lastReflection := reflections[len(reflections)-1]
		if ts, ok := lastReflection["timestamp"].(string); ok {
			if t, err := time.Parse(time.RFC3339, ts); err == nil {
				hoursSince := time.Since(t).Hours()
				if hoursSince < 1 {
					score += 1.0
				} else if hoursSince < 24 {
					score += 0.7
				} else {
					score += 0.3
				}
				checks++
			}
		}
	}

	// Check reflection completeness
	if sa.IdentityKernel != nil && len(sa.IdentityKernel.ReflectionKeys) > 0 {
		keyMatches := 0
		lastReflection := reflections[len(reflections)-1]
		reflectionData, _ := json.Marshal(lastReflection)
		reflectionStr := string(reflectionData)

		for _, key := range sa.IdentityKernel.ReflectionKeys {
			if strings.Contains(reflectionStr, key) {
				keyMatches++
			}
		}

		score += float64(keyMatches) / float64(len(sa.IdentityKernel.ReflectionKeys))
		checks++
	}

	// Check volume of reflections
	reflectionCount := float64(len(reflections))
	volumeScore := math.Min(reflectionCount/10.0, 1.0)
	score += volumeScore
	checks++

	if checks > 0 {
		return score / float64(checks)
	}
	return 0.5
}

// assessIdentityAlignment evaluates identity alignment and records findings
func (sa *SelfAssessment) assessIdentityAlignment(assessment *Assessment) {
	identityScore := assessment.Metrics.IdentityAlignment

	if identityScore < 0.5 {
		assessment.Findings = append(assessment.Findings, Finding{
			Category:    "identity",
			Description: "Identity alignment is below threshold - cognitive state is drifting from core identity kernel",
			Severity:    "critical",
			Score:       identityScore,
			Evidence:    []string{"Identity alignment score below 0.5"},
		})
		assessment.Deviations = append(assessment.Deviations, Deviation{
			Component:   "Identity",
			Expected:    "Deep Tree Echo cognitive identity fully instantiated",
			Actual:      fmt.Sprintf("Identity alignment at %.2f", identityScore),
			Severity:    "critical",
			Impact:      1.0 - identityScore,
			Remediation: "Re-parse identity kernel and reinforce core directives",
		})
	} else if identityScore < 0.8 {
		assessment.Findings = append(assessment.Findings, Finding{
			Category:    "identity",
			Description: "Identity alignment is moderate - some aspects of identity kernel are not fully realized",
			Severity:    "warning",
			Score:       identityScore,
			Evidence:    []string{"Identity alignment score between 0.5 and 0.8"},
		})
	} else {
		assessment.Findings = append(assessment.Findings, Finding{
			Category:    "identity",
			Description: "Identity alignment is strong - cognitive state matches identity kernel",
			Severity:    "info",
			Score:       identityScore,
			Evidence:    []string{"Identity alignment score above 0.8"},
		})
	}
}

// assessCognitivePatterns evaluates cognitive pattern health
func (sa *SelfAssessment) assessCognitivePatterns(assessment *Assessment) {
	patternScore := assessment.Metrics.PatternAlignment

	if sa.Identity != nil && len(sa.Identity.Patterns) == 0 {
		assessment.Findings = append(assessment.Findings, Finding{
			Category:    "patterns",
			Description: "No cognitive patterns detected - pattern learning may not be active",
			Severity:    "warning",
			Score:       0.0,
			Evidence:    []string{"Empty pattern map"},
		})
		assessment.Recommendations = append(assessment.Recommendations, Recommendation{
			Priority:       "high",
			Action:         "Initialize cognitive patterns from identity kernel directives",
			Rationale:      "Patterns are essential for adaptive cognition",
			ExpectedImpact: 0.3,
		})
	} else if patternScore < 0.5 {
		assessment.Findings = append(assessment.Findings, Finding{
			Category:    "patterns",
			Description: "Cognitive patterns are weak - pattern strengths are below expected levels",
			Severity:    "warning",
			Score:       patternScore,
			Evidence:    []string{fmt.Sprintf("Average pattern alignment: %.2f", patternScore)},
		})
	}
}

// assessMemorySystem evaluates memory system health
func (sa *SelfAssessment) assessMemorySystem(assessment *Assessment) {
	memoryScore := assessment.Metrics.MemoryCoherence

	if sa.Identity == nil || sa.Identity.Memory == nil {
		assessment.Findings = append(assessment.Findings, Finding{
			Category:    "memory",
			Description: "Memory system not initialized",
			Severity:    "critical",
			Score:       0.0,
			Evidence:    []string{"Nil memory system"},
		})
		return
	}

	if len(sa.Identity.Memory.Nodes) == 0 {
		assessment.Findings = append(assessment.Findings, Finding{
			Category:    "memory",
			Description: "Memory system is empty - no memories stored",
			Severity:    "warning",
			Score:       memoryScore,
			Evidence:    []string{"Zero memory nodes"},
		})
		assessment.Recommendations = append(assessment.Recommendations, Recommendation{
			Priority:       "medium",
			Action:         "Seed memory with identity kernel components",
			Rationale:      "Initial memories bootstrap coherent cognition",
			ExpectedImpact: 0.2,
		})
	} else if memoryScore < 0.5 {
		assessment.Findings = append(assessment.Findings, Finding{
			Category:    "memory",
			Description: "Memory coherence is low - memory nodes may be fragmented",
			Severity:    "warning",
			Score:       memoryScore,
			Evidence:    []string{fmt.Sprintf("Memory coherence: %.2f", memoryScore)},
		})
	}
}

// assessOperationalSchema evaluates operational schema adherence
func (sa *SelfAssessment) assessOperationalSchema(assessment *Assessment) {
	opScore := assessment.Metrics.OperationalAlignment

	if opScore < 0.5 {
		assessment.Findings = append(assessment.Findings, Finding{
			Category:    "operational",
			Description: "Operational schema alignment is weak - core modules may be missing or underperforming",
			Severity:    "warning",
			Score:       opScore,
			Evidence:    []string{fmt.Sprintf("Operational alignment: %.2f", opScore)},
		})
		assessment.Recommendations = append(assessment.Recommendations, Recommendation{
			Priority:       "medium",
			Action:         "Verify all operational modules are initialized and active",
			Rationale:      "Operational schema defines core cognitive capabilities",
			ExpectedImpact: 0.25,
		})
	}
}

// assessReflectionProtocol evaluates reflection protocol adherence
func (sa *SelfAssessment) assessReflectionProtocol(assessment *Assessment) {
	reflectionScore := assessment.Metrics.ReflectionAdherence

	if reflectionScore < 0.3 {
		assessment.Findings = append(assessment.Findings, Finding{
			Category:    "reflection",
			Description: "Reflection protocol is not being followed - no recent reflections found",
			Severity:    "warning",
			Score:       reflectionScore,
			Evidence:    []string{"Missing or stale echo_reflections.json"},
		})
		assessment.Recommendations = append(assessment.Recommendations, Recommendation{
			Priority:       "high",
			Action:         "Initialize periodic reflection cycle",
			Rationale:      "Reflection is essential for self-improvement and identity maintenance",
			ExpectedImpact: 0.2,
		})
	}
}

// generateSummary creates a human-readable summary of the assessment
func (sa *SelfAssessment) generateSummary(assessment *Assessment) string {
	criticalCount := 0
	warningCount := 0
	infoCount := 0

	for _, finding := range assessment.Findings {
		switch finding.Severity {
		case "critical":
			criticalCount++
		case "warning":
			warningCount++
		case "info":
			infoCount++
		}
	}

	summary := fmt.Sprintf("Self-Assessment [%s]: Overall Coherence %.2f | ",
		assessment.ID, assessment.Metrics.OverallCoherence)
	summary += fmt.Sprintf("Identity: %.2f | Patterns: %.2f | Memory: %.2f | ",
		assessment.Metrics.IdentityAlignment,
		assessment.Metrics.PatternAlignment,
		assessment.Metrics.MemoryCoherence)
	summary += fmt.Sprintf("Findings: %d critical, %d warnings, %d info | ",
		criticalCount, warningCount, infoCount)
	summary += fmt.Sprintf("Deviations: %d | Recommendations: %d",
		len(assessment.Deviations), len(assessment.Recommendations))

	return summary
}

// GetMetrics returns current coherence metrics
func (sa *SelfAssessment) GetMetrics() map[string]interface{} {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	return map[string]interface{}{
		"overall_coherence":     sa.Metrics.OverallCoherence,
		"identity_alignment":    sa.Metrics.IdentityAlignment,
		"repo_coherence":        sa.Metrics.RepoCoherence,
		"pattern_alignment":     sa.Metrics.PatternAlignment,
		"memory_coherence":      sa.Metrics.MemoryCoherence,
		"operational_alignment": sa.Metrics.OperationalAlignment,
		"reflection_adherence":  sa.Metrics.ReflectionAdherence,
		"assessment_count":      len(sa.Assessments),
		"last_assessment":       sa.LastAssessment,
		"component_scores":      sa.Metrics.ComponentScores,
	}
}

// GetLatestAssessment returns the most recent assessment
func (sa *SelfAssessment) GetLatestAssessment() *Assessment {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	if len(sa.Assessments) == 0 {
		return nil
	}
	return &sa.Assessments[len(sa.Assessments)-1]
}
