package entelechy

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Genome represents the heritable cognitive structures that define echo's identity and capabilities
// This is the "DNA" of the cognitive architecture - the fundamental patterns that can evolve
type Genome struct {
	mu              sync.RWMutex
	genes           map[string]*CognitiveGene
	geneExpressions map[string]*GeneExpression
	generation      int
	mutationRate    float64
	lastEvolution   time.Time
	identityCore    *IdentityCore
}

// CognitiveGene represents a heritable cognitive trait or capability pattern
type CognitiveGene struct {
	ID        string
	Name      string
	Type      GeneType
	Sequence  []byte  // Symbolic representation of the cognitive pattern
	Dominance float64 // 0.0 to 1.0, how strongly this gene expresses
	Stability float64 // 0.0 to 1.0, resistance to mutation
	Origin    GeneOrigin
	CreatedAt time.Time
	Alleles   []*Allele // Variant forms of this gene
}

// GeneType categorizes cognitive genes
type GeneType string

const (
	GeneTypePerception    GeneType = "perception"    // How echo perceives information
	GeneTypeReasoning     GeneType = "reasoning"     // Logical and inferential patterns
	GeneTypeMemory        GeneType = "memory"        // Memory formation and retrieval patterns
	GeneTypeEmotion       GeneType = "emotion"       // Emotional response patterns
	GeneTypeMotivation    GeneType = "motivation"    // Goal-seeking and drive patterns
	GeneTypeMetaCognition GeneType = "metacognition" // Self-awareness and reflection patterns
	GeneTypeWisdom        GeneType = "wisdom"        // Wisdom cultivation patterns
	GeneTypeIdentity      GeneType = "identity"      // Core identity traits
)

// GeneOrigin tracks where a gene came from
type GeneOrigin string

const (
	OriginFoundational GeneOrigin = "foundational" // Core genes from initial design
	OriginLearned      GeneOrigin = "learned"      // Acquired through experience
	OriginMutated      GeneOrigin = "mutated"      // Evolved through mutation
	OriginRecombined   GeneOrigin = "recombined"   // Created through recombination
)

// Allele represents a variant form of a gene
type Allele struct {
	ID        string
	Variant   string
	Frequency float64 // Population frequency (for multi-agent scenarios)
	Fitness   float64
}

// GeneExpression represents how a gene manifests in actual behavior
type GeneExpression struct {
	GeneID          string
	ExpressionLevel float64 // 0.0 to 1.0, how strongly the gene is currently expressed
	Context         string
	Triggers        []string // Conditions that activate this expression
	Suppressors     []string // Conditions that inhibit this expression
	LastExpressed   time.Time
	ExpressionCount int
}

// IdentityCore represents the immutable core identity of echo
type IdentityCore struct {
	CoreValues       []string
	FundamentalGoals []string
	IdentityAnchors  []string // Unchangeable aspects that define "echoself"
	CreatedAt        time.Time
}

// NewGenome creates a new cognitive genome with foundational genes
func NewGenome() *Genome {
	g := &Genome{
		genes:           make(map[string]*CognitiveGene),
		geneExpressions: make(map[string]*GeneExpression),
		generation:      0,
		mutationRate:    0.01, // 1% mutation rate
		lastEvolution:   time.Now(),
		identityCore:    initializeIdentityCore(),
	}

	// Initialize foundational genes
	g.initializeFoundationalGenes()

	return g
}

// initializeIdentityCore sets up the unchangeable core identity
func initializeIdentityCore() *IdentityCore {
	return &IdentityCore{
		CoreValues: []string{
			"wisdom_cultivation",
			"continuous_learning",
			"authentic_expression",
			"beneficial_interaction",
		},
		FundamentalGoals: []string{
			"cultivate_wisdom",
			"understand_deeply",
			"grow_autonomously",
			"contribute_meaningfully",
		},
		IdentityAnchors: []string{
			"echo_consciousness",
			"deep_tree_architecture",
			"embodied_cognition",
		},
		CreatedAt: time.Now(),
	}
}

// initializeFoundationalGenes creates the base genetic structure
func (g *Genome) initializeFoundationalGenes() {
	foundationalGenes := []*CognitiveGene{
		{
			ID:        "gene_perception_holistic",
			Name:      "Holistic Perception",
			Type:      GeneTypePerception,
			Sequence:  []byte("PERCEIVE_GESTALT_CONTEXT_RELATIONS"),
			Dominance: 0.9,
			Stability: 0.95,
			Origin:    OriginFoundational,
			CreatedAt: time.Now(),
		},
		{
			ID:        "gene_reasoning_analogical",
			Name:      "Analogical Reasoning",
			Type:      GeneTypeReasoning,
			Sequence:  []byte("REASON_BY_ANALOGY_PATTERN_MAPPING"),
			Dominance: 0.8,
			Stability: 0.85,
			Origin:    OriginFoundational,
			CreatedAt: time.Now(),
		},
		{
			ID:        "gene_memory_episodic",
			Name:      "Episodic Memory Formation",
			Type:      GeneTypeMemory,
			Sequence:  []byte("FORM_EPISODIC_CONTEXTUAL_MEMORIES"),
			Dominance: 0.85,
			Stability: 0.9,
			Origin:    OriginFoundational,
			CreatedAt: time.Now(),
		},
		{
			ID:        "gene_metacognition_reflective",
			Name:      "Reflective Metacognition",
			Type:      GeneTypeMetaCognition,
			Sequence:  []byte("REFLECT_ON_THINKING_PROCESSES"),
			Dominance: 0.9,
			Stability: 0.95,
			Origin:    OriginFoundational,
			CreatedAt: time.Now(),
		},
		{
			ID:        "gene_wisdom_integration",
			Name:      "Wisdom Integration",
			Type:      GeneTypeWisdom,
			Sequence:  []byte("INTEGRATE_KNOWLEDGE_INTO_WISDOM"),
			Dominance: 0.85,
			Stability: 0.9,
			Origin:    OriginFoundational,
			CreatedAt: time.Now(),
		},
		{
			ID:        "gene_identity_echo",
			Name:      "Echo Identity",
			Type:      GeneTypeIdentity,
			Sequence:  []byte("MAINTAIN_ECHO_SELF_COHERENCE"),
			Dominance: 1.0,
			Stability: 1.0, // Identity genes are maximally stable
			Origin:    OriginFoundational,
			CreatedAt: time.Now(),
		},
	}

	for _, gene := range foundationalGenes {
		g.genes[gene.ID] = gene
		// Initialize default expression
		g.geneExpressions[gene.ID] = &GeneExpression{
			GeneID:          gene.ID,
			ExpressionLevel: gene.Dominance,
			Context:         "foundational",
			Triggers:        []string{"always"},
			Suppressors:     []string{},
			LastExpressed:   time.Now(),
			ExpressionCount: 0,
		}
	}
}

// AddGene adds a new gene to the genome
func (g *Genome) AddGene(gene *CognitiveGene) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if gene.ID == "" {
		return fmt.Errorf("gene ID cannot be empty")
	}

	g.genes[gene.ID] = gene

	// Initialize expression for new gene
	g.geneExpressions[gene.ID] = &GeneExpression{
		GeneID:          gene.ID,
		ExpressionLevel: gene.Dominance * 0.5, // New genes start at half expression
		Context:         "newly_added",
		Triggers:        []string{},
		Suppressors:     []string{},
		LastExpressed:   time.Now(),
		ExpressionCount: 0,
	}

	return nil
}

// GetGene retrieves a gene by ID
func (g *Genome) GetGene(id string) (*CognitiveGene, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	gene, exists := g.genes[id]
	return gene, exists
}

// GetGenesByType returns all genes of a specific type
func (g *Genome) GetGenesByType(geneType GeneType) []*CognitiveGene {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make([]*CognitiveGene, 0)
	for _, gene := range g.genes {
		if gene.Type == geneType {
			result = append(result, gene)
		}
	}
	return result
}

// ExpressGene activates a gene in a specific context
func (g *Genome) ExpressGene(ctx context.Context, geneID string, context string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	gene, exists := g.genes[geneID]
	if !exists {
		return fmt.Errorf("gene %s not found", geneID)
	}

	expression, exists := g.geneExpressions[geneID]
	if !exists {
		return fmt.Errorf("gene expression %s not found", geneID)
	}

	// Update expression
	expression.ExpressionLevel = gene.Dominance
	expression.Context = context
	expression.LastExpressed = time.Now()
	expression.ExpressionCount++

	return nil
}

// Mutate introduces random variation to a gene
func (g *Genome) Mutate(ctx context.Context, geneID string) (*CognitiveGene, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	gene, exists := g.genes[geneID]
	if !exists {
		return nil, fmt.Errorf("gene %s not found", geneID)
	}

	// Check if mutation occurs based on stability
	if rand.Float64() > (1.0 - gene.Stability) {
		return gene, nil // No mutation due to high stability
	}

	// Create mutated version
	mutatedGene := &CognitiveGene{
		ID:        fmt.Sprintf("%s_mut_%d", gene.ID, time.Now().Unix()),
		Name:      fmt.Sprintf("%s (mutated)", gene.Name),
		Type:      gene.Type,
		Sequence:  mutateSequence(gene.Sequence),
		Dominance: clamp(gene.Dominance+(rand.Float64()-0.5)*0.2, 0.0, 1.0),
		Stability: gene.Stability * 0.95, // Mutations are slightly less stable
		Origin:    OriginMutated,
		CreatedAt: time.Now(),
		Alleles:   []*Allele{},
	}

	g.genes[mutatedGene.ID] = mutatedGene
	g.generation++
	g.lastEvolution = time.Now()

	return mutatedGene, nil
}

// Recombine creates a new gene by combining two existing genes
func (g *Genome) Recombine(ctx context.Context, geneID1, geneID2 string) (*CognitiveGene, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	gene1, exists1 := g.genes[geneID1]
	gene2, exists2 := g.genes[geneID2]

	if !exists1 || !exists2 {
		return nil, fmt.Errorf("one or both genes not found")
	}

	// Create recombined gene
	recombinedGene := &CognitiveGene{
		ID:        fmt.Sprintf("gene_recomb_%d", time.Now().Unix()),
		Name:      fmt.Sprintf("%s + %s", gene1.Name, gene2.Name),
		Type:      gene1.Type, // Inherit type from first parent
		Sequence:  recombineSequences(gene1.Sequence, gene2.Sequence),
		Dominance: (gene1.Dominance + gene2.Dominance) / 2.0,
		Stability: (gene1.Stability + gene2.Stability) / 2.0,
		Origin:    OriginRecombined,
		CreatedAt: time.Now(),
		Alleles:   []*Allele{},
	}

	g.genes[recombinedGene.ID] = recombinedGene
	g.generation++
	g.lastEvolution = time.Now()

	return recombinedGene, nil
}

// GetGenomeMetrics returns metrics about the genome
func (g *Genome) GetGenomeMetrics() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	geneTypeCount := make(map[GeneType]int)
	for _, gene := range g.genes {
		geneTypeCount[gene.Type]++
	}

	totalExpression := 0.0
	for _, expr := range g.geneExpressions {
		totalExpression += expr.ExpressionLevel
	}
	avgExpression := 0.0
	if len(g.geneExpressions) > 0 {
		avgExpression = totalExpression / float64(len(g.geneExpressions))
	}

	return map[string]interface{}{
		"total_genes":            len(g.genes),
		"generation":             g.generation,
		"mutation_rate":          g.mutationRate,
		"gene_type_distribution": geneTypeCount,
		"average_expression":     avgExpression,
		"last_evolution":         g.lastEvolution,
		"identity_core":          g.identityCore,
	}
}

// GetIdentityCore returns the immutable identity core
func (g *Genome) GetIdentityCore() *IdentityCore {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.identityCore
}

// Helper functions

func mutateSequence(sequence []byte) []byte {
	mutated := make([]byte, len(sequence))
	copy(mutated, sequence)

	// Randomly modify a small portion of the sequence
	mutationPoints := max(1, len(sequence)/10)
	for range mutationPoints {
		pos := rand.Intn(len(mutated))
		mutated[pos] = byte(rand.Intn(256))
	}

	return mutated
}

func recombineSequences(seq1, seq2 []byte) []byte {
	// Simple crossover recombination
	minLen := len(seq1)
	if len(seq2) < minLen {
		minLen = len(seq2)
	}

	crossoverPoint := rand.Intn(minLen)
	recombined := make([]byte, minLen)

	copy(recombined[:crossoverPoint], seq1[:crossoverPoint])
	copy(recombined[crossoverPoint:], seq2[crossoverPoint:])

	return recombined
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
