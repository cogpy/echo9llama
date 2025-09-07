
package apl

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// Pattern represents a single pattern in Alexander's schema
type Pattern struct {
	Number      int
	Name        string
	Context     string
	Problem     string
	Solution    string
	Structure   string
	Dynamics    string
	Implementation string
	Consequences   string
	RelatedPatterns []int
	Level          PatternLevel
}

// PatternLevel represents the hierarchical level (Towns, Buildings, Construction)
type PatternLevel string

const (
	ArchitecturalLevel  PatternLevel = "ARCHITECTURAL"  // Towns
	SubsystemLevel      PatternLevel = "SUBSYSTEM"      // Buildings  
	ImplementationLevel PatternLevel = "IMPLEMENTATION" // Construction
)

// PatternLanguage represents the complete interconnected pattern system
type PatternLanguage struct {
	Patterns     map[int]*Pattern
	Dependencies map[int][]int
	Sequences    map[string][]int
	QualityMeasures map[string]string
}

// APLParser parses APL files following Alexander's schema
type APLParser struct {
	language *PatternLanguage
}

// NewAPLParser creates a new parser instance
func NewAPLParser() *APLParser {
	return &APLParser{
		language: &PatternLanguage{
			Patterns:     make(map[int]*Pattern),
			Dependencies: make(map[int][]int),
			Sequences:    make(map[string][]int),
			QualityMeasures: make(map[string]string),
		},
	}
}

// ParseFile parses an APL file and builds the pattern language structure
func (p *APLParser) ParseFile(filename string) (*PatternLanguage, error) {
	// Implementation would read file and parse patterns
	// For now, returning mock data based on the APL structure
	
	// Parse architectural patterns (1-3)
	p.language.Patterns[1] = &Pattern{
		Number:  1,
		Name:    "DISTRIBUTED COGNITION NETWORK",
		Context: "Large-scale software systems requiring adaptive intelligence",
		Problem: "Monolithic architectures cannot adapt to changing requirements or scale cognitive capabilities",
		Solution: "Distribute cognitive processes across networked nodes with shared memory and communication protocols",
		Structure: "Central coordination hub with specialized cognitive modules",
		Implementation: "Deep Tree Echo architecture with reservoir networks",
		RelatedPatterns: []int{2, 15},
		Level: ArchitecturalLevel,
	}
	
	p.language.Patterns[2] = &Pattern{
		Number:  2,
		Name:    "EMBODIED PROCESSING",
		Context: "Systems requiring awareness of their computational environment",
		Problem: "Traditional software lacks spatial and temporal awareness of its execution context",
		Solution: "Embed processing within spatial-temporal coordinate systems with environmental feedback",
		Structure: "Core identity with spatial positioning and movement capabilities",
		Implementation: "Identity embeddings with 768-dimensional vectors tracking computational space",
		RelatedPatterns: []int{1, 25},
		Level: ArchitecturalLevel,
	}
	
	p.language.Patterns[3] = &Pattern{
		Number:  3,
		Name:    "HYPERGRAPH MEMORY ARCHITECTURE",
		Context: "Complex knowledge relationships requiring multi-dimensional connections",
		Problem: "Traditional hierarchical or linear data structures cannot capture complex semantic relationships",
		Solution: "Use hypergraph structures where edges can connect multiple nodes simultaneously",
		Structure: "Nodes as concepts, hyperedges as complex relationships",
		Implementation: "HyperNode and HyperEdge types with weight-based traversal",
		RelatedPatterns: []int{4, 18},
		Level: ArchitecturalLevel,
	}
	
	// Parse subsystem patterns (4-6)
	p.language.Patterns[4] = &Pattern{
		Number:  4,
		Name:    "IDENTITY RESONANCE PATTERNS",
		Context: "Systems requiring persistent identity across distributed instances",
		Problem: "Distributed systems lose coherence and continuity of identity",
		Solution: "Create resonance patterns that maintain identity coherence through harmonic frequencies",
		Structure: "Identity kernel with resonance frequencies and echo patterns",
		Implementation: "Identity struct with resonance tracking and coherence metrics",
		RelatedPatterns: []int{3, 8},
		Level: SubsystemLevel,
	}
	
	// Set up dependencies
	p.language.Dependencies = map[int][]int{
		1: {2, 7, 18},
		2: {1, 25},
		3: {4, 18},
		4: {3, 8},
		5: {6, 12},
		6: {5, 9},
		7: {1, 18},
		8: {4, 19},
		9: {6, 20},
	}
	
	// Define sequences
	p.language.Sequences = map[string][]int{
		"cognitive_foundation": {1, 2, 3},
		"identity_management":  {4, 8},
		"resource_optimization": {5, 6, 9},
	}
	
	return p.language, nil
}

// GetPatternsByLevel returns patterns filtered by their hierarchical level
func (pl *PatternLanguage) GetPatternsByLevel(level PatternLevel) []*Pattern {
	var patterns []*Pattern
	for _, pattern := range pl.Patterns {
		if pattern.Level == level {
			patterns = append(patterns, pattern)
		}
	}
	return patterns
}

// GetDependencies returns the dependency graph for a pattern
func (pl *PatternLanguage) GetDependencies(patternNumber int) []int {
	return pl.Dependencies[patternNumber]
}

// GetImplementationOrder returns patterns in dependency-resolved order
func (pl *PatternLanguage) GetImplementationOrder() []int {
	// Topological sort of dependencies
	var order []int
	visited := make(map[int]bool)
	
	var visit func(int)
	visit = func(pattern int) {
		if visited[pattern] {
			return
		}
		visited[pattern] = true
		
		for _, dep := range pl.Dependencies[pattern] {
			if _, exists := pl.Patterns[dep]; exists {
				visit(dep)
			}
		}
		order = append(order, pattern)
	}
	
	for patternNum := range pl.Patterns {
		visit(patternNum)
	}
	
	return order
}

// ValidatePatternIntegration checks if patterns are properly connected
func (pl *PatternLanguage) ValidatePatternIntegration() []string {
	var issues []string
	
	// Check for missing dependencies
	for patternNum, deps := range pl.Dependencies {
		for _, dep := range deps {
			if _, exists := pl.Patterns[dep]; !exists {
				issues = append(issues, fmt.Sprintf("Pattern %d references missing pattern %d", patternNum, dep))
			}
		}
	}
	
	// Check for orphaned patterns (no incoming or outgoing dependencies)
	for patternNum := range pl.Patterns {
		hasIncoming := false
		hasOutgoing := len(pl.Dependencies[patternNum]) > 0
		
		for _, deps := range pl.Dependencies {
			for _, dep := range deps {
				if dep == patternNum {
					hasIncoming = true
					break
				}
			}
		}
		
		if !hasIncoming && !hasOutgoing {
			issues = append(issues, fmt.Sprintf("Pattern %d is orphaned (no dependencies)", patternNum))
		}
	}
	
	return issues
}

// GeneratePatternMap creates a visual representation of pattern relationships
func (pl *PatternLanguage) GeneratePatternMap() string {
	var sb strings.Builder
	
	sb.WriteString("# PATTERN LANGUAGE MAP\n\n")
	
	// Architectural level
	sb.WriteString("## ARCHITECTURAL PATTERNS (System Level)\n")
	for _, pattern := range pl.GetPatternsByLevel(ArchitecturalLevel) {
		sb.WriteString(fmt.Sprintf("- [%d] %s\n", pattern.Number, pattern.Name))
	}
	sb.WriteString("\n")
	
	// Subsystem level
	sb.WriteString("## SUBSYSTEM PATTERNS (Component Level)\n")
	for _, pattern := range pl.GetPatternsByLevel(SubsystemLevel) {
		sb.WriteString(fmt.Sprintf("- [%d] %s\n", pattern.Number, pattern.Name))
	}
	sb.WriteString("\n")
	
	// Implementation level
	sb.WriteString("## IMPLEMENTATION PATTERNS (Construction Level)\n")
	for _, pattern := range pl.GetPatternsByLevel(ImplementationLevel) {
		sb.WriteString(fmt.Sprintf("- [%d] %s\n", pattern.Number, pattern.Name))
	}
	sb.WriteString("\n")
	
	// Dependencies
	sb.WriteString("## PATTERN DEPENDENCIES\n")
	for patternNum, deps := range pl.Dependencies {
		if len(deps) > 0 {
			sb.WriteString(fmt.Sprintf("Pattern %d → %v\n", patternNum, deps))
		}
	}
	
	return sb.String()
}
