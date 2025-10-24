package opencog

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// AtomSpace represents the OpenCog AtomSpace - a weighted, labeled hypergraph
// for knowledge representation in the Deep Tree Echo cognitive architecture
type AtomSpace struct {
	mu sync.RWMutex

	// Core hypergraph structures
	Atoms       map[string]*Atom
	Links       map[string]*Link
	Incoming    map[string][]string // Reverse index for efficient queries
	
	// Attention allocation
	AttentionBank *AttentionBank
	
	// Truth value system
	TruthValueCache map[string]*TruthValue
	
	// Pattern matcher
	PatternMatcher *PatternMatcher
	
	// Statistics
	Created   time.Time
	Modified  time.Time
	AtomCount int64
	LinkCount int64
}

// Atom represents a node in the hypergraph
type Atom struct {
	ID        string
	Type      AtomType
	Name      string
	TruthValue *TruthValue
	Attention  *AttentionValue
	Incoming   []string // Links pointing to this atom
	Created    time.Time
	Modified   time.Time
}

// Link represents a hyperedge connecting atoms
type Link struct {
	ID         string
	Type       LinkType
	Outgoing   []string // Atoms this link points to
	TruthValue *TruthValue
	Attention  *AttentionValue
	Created    time.Time
	Modified   time.Time
}

// AtomType represents the type of an atom
type AtomType string

const (
	ConceptNode     AtomType = "ConceptNode"
	PredicateNode   AtomType = "PredicateNode"
	VariableNode    AtomType = "VariableNode"
	NumberNode      AtomType = "NumberNode"
	SchemaNode      AtomType = "SchemaNode"
	GroundedSchemaNode AtomType = "GroundedSchemaNode"
)

// LinkType represents the type of a link
type LinkType string

const (
	InheritanceLink    LinkType = "InheritanceLink"
	SimilarityLink     LinkType = "SimilarityLink"
	EvaluationLink     LinkType = "EvaluationLink"
	MemberLink         LinkType = "MemberLink"
	ListLink           LinkType = "ListLink"
	ExecutionLink      LinkType = "ExecutionLink"
	ImplicationLink    LinkType = "ImplicationLink"
	EquivalenceLink    LinkType = "EquivalenceLink"
	SubsetLink         LinkType = "SubsetLink"
	AndLink            LinkType = "AndLink"
	OrLink             LinkType = "OrLink"
	NotLink            LinkType = "NotLink"
)

// TruthValue represents probabilistic truth values (PLN)
type TruthValue struct {
	Strength    float64 // Probability [0, 1]
	Confidence  float64 // Weight of evidence [0, 1]
	Count       float64 // Amount of evidence
}

// AttentionValue represents economic attention allocation
type AttentionValue struct {
	STI  int16   // Short-term importance
	LTI  int16   // Long-term importance
	VLTI int16   // Very long-term importance
	AF   float64 // Attention focus
}

// AttentionBank manages attention allocation across atoms
type AttentionBank struct {
	mu sync.RWMutex
	
	// ECAN (Economic Attention Network) parameters
	STIFunds    int64
	LTIFunds    int64
	AFBoundary  float64
	
	// Atom importance heaps
	STIHeap     *ImportanceHeap
	LTIHeap     *ImportanceHeap
	
	// Forgetting mechanism
	ForgettingRate float64
	MinAF          float64
	MaxAF          float64
}

// ImportanceHeap maintains atoms sorted by importance
type ImportanceHeap struct {
	atoms    []string
	values   map[string]int16
}

// PatternMatcher performs pattern matching queries
type PatternMatcher struct {
	mu sync.RWMutex
	
	// Pattern templates
	Templates map[string]*Pattern
	
	// Query cache
	QueryCache map[string]*QueryResult
	CacheSize  int
}

// Pattern represents a query pattern
type Pattern struct {
	Variables map[string]*Variable
	Clauses   []Clause
	Type      PatternType
}

// PatternType defines the type of pattern matching
type PatternType string

const (
	BindPattern      PatternType = "BindPattern"
	GetPattern       PatternType = "GetPattern"
	SatisfactionPattern PatternType = "SatisfactionPattern"
)

// Variable represents a pattern variable
type Variable struct {
	Name  string
	Type  AtomType
	Constraints []Constraint
}

// Constraint restricts variable binding
type Constraint struct {
	Type  ConstraintType
	Value interface{}
}

// ConstraintType defines constraint types
type ConstraintType string

const (
	TypeConstraint      ConstraintType = "TypeConstraint"
	TruthValueConstraint ConstraintType = "TruthValueConstraint"
	AttentionConstraint  ConstraintType = "AttentionConstraint"
)

// Clause represents a pattern clause
type Clause struct {
	LinkType LinkType
	Atoms    []string // Atom IDs or variable names
}

// QueryResult holds pattern matching results
type QueryResult struct {
	Bindings  []map[string]string
	Timestamp time.Time
	Count     int
}

// NewAtomSpace creates a new AtomSpace
func NewAtomSpace() *AtomSpace {
	return &AtomSpace{
		Atoms:           make(map[string]*Atom),
		Links:           make(map[string]*Link),
		Incoming:        make(map[string][]string),
		TruthValueCache: make(map[string]*TruthValue),
		AttentionBank:   NewAttentionBank(),
		PatternMatcher:  NewPatternMatcher(),
		Created:         time.Now(),
		Modified:        time.Now(),
	}
}

// NewAttentionBank creates a new AttentionBank
func NewAttentionBank() *AttentionBank {
	return &AttentionBank{
		STIFunds:       100000,
		LTIFunds:       100000,
		AFBoundary:     0.5,
		STIHeap:        &ImportanceHeap{atoms: []string{}, values: make(map[string]int16)},
		LTIHeap:        &ImportanceHeap{atoms: []string{}, values: make(map[string]int16)},
		ForgettingRate: 0.01,
		MinAF:          0.0,
		MaxAF:          1.0,
	}
}

// NewPatternMatcher creates a new PatternMatcher
func NewPatternMatcher() *PatternMatcher {
	return &PatternMatcher{
		Templates:  make(map[string]*Pattern),
		QueryCache: make(map[string]*QueryResult),
		CacheSize:  1000,
	}
}

// AddAtom adds an atom to the AtomSpace
func (as *AtomSpace) AddAtom(atomType AtomType, name string, tv *TruthValue) (*Atom, error) {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	id := fmt.Sprintf("atom_%s_%d", name, time.Now().UnixNano())
	
	if tv == nil {
		tv = &TruthValue{Strength: 1.0, Confidence: 0.0, Count: 0.0}
	}
	
	atom := &Atom{
		ID:         id,
		Type:       atomType,
		Name:       name,
		TruthValue: tv,
		Attention: &AttentionValue{
			STI:  0,
			LTI:  0,
			VLTI: 0,
			AF:   0.0,
		},
		Incoming: []string{},
		Created:  time.Now(),
		Modified: time.Now(),
	}
	
	as.Atoms[id] = atom
	as.AtomCount++
	as.Modified = time.Now()
	
	// Register with attention bank
	as.AttentionBank.RegisterAtom(id, atom.Attention)
	
	return atom, nil
}

// AddLink adds a link to the AtomSpace
func (as *AtomSpace) AddLink(linkType LinkType, outgoing []string, tv *TruthValue) (*Link, error) {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	// Verify all outgoing atoms exist
	for _, atomID := range outgoing {
		if _, exists := as.Atoms[atomID]; !exists {
			if _, linkExists := as.Links[atomID]; !linkExists {
				return nil, fmt.Errorf("atom or link %s not found", atomID)
			}
		}
	}
	
	id := fmt.Sprintf("link_%s_%d", linkType, time.Now().UnixNano())
	
	if tv == nil {
		tv = &TruthValue{Strength: 1.0, Confidence: 0.0, Count: 0.0}
	}
	
	link := &Link{
		ID:         id,
		Type:       linkType,
		Outgoing:   outgoing,
		TruthValue: tv,
		Attention: &AttentionValue{
			STI:  0,
			LTI:  0,
			VLTI: 0,
			AF:   0.0,
		},
		Created:  time.Now(),
		Modified: time.Now(),
	}
	
	as.Links[id] = link
	as.LinkCount++
	as.Modified = time.Now()
	
	// Update incoming links for each atom
	for _, atomID := range outgoing {
		as.Incoming[atomID] = append(as.Incoming[atomID], id)
		if atom, exists := as.Atoms[atomID]; exists {
			atom.Incoming = append(atom.Incoming, id)
		}
	}
	
	// Register with attention bank
	as.AttentionBank.RegisterAtom(id, link.Attention)
	
	return link, nil
}

// GetAtom retrieves an atom by ID
func (as *AtomSpace) GetAtom(id string) (*Atom, bool) {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	atom, exists := as.Atoms[id]
	return atom, exists
}

// GetLink retrieves a link by ID
func (as *AtomSpace) GetLink(id string) (*Link, bool) {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	link, exists := as.Links[id]
	return link, exists
}

// GetIncoming gets all links pointing to an atom
func (as *AtomSpace) GetIncoming(atomID string) []string {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	return as.Incoming[atomID]
}

// UpdateTruthValue updates an atom's truth value
func (as *AtomSpace) UpdateTruthValue(id string, tv *TruthValue) error {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	if atom, exists := as.Atoms[id]; exists {
		atom.TruthValue = tv
		atom.Modified = time.Now()
		as.Modified = time.Now()
		return nil
	}
	
	if link, exists := as.Links[id]; exists {
		link.TruthValue = tv
		link.Modified = time.Now()
		as.Modified = time.Now()
		return nil
	}
	
	return fmt.Errorf("atom or link %s not found", id)
}

// SpreadAttention spreads attention through the hypergraph (ECAN)
func (as *AtomSpace) SpreadAttention() {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	// Implement Economic Attention Network spreading
	for _, link := range as.Links {
		// Spread STI based on link strength
		strength := link.TruthValue.Strength
		sourceSTI := link.Attention.STI
		
		for _, targetID := range link.Outgoing {
			if atom, exists := as.Atoms[targetID]; exists {
				// Transfer STI proportional to truth value strength
				transfer := int16(float64(sourceSTI) * strength * 0.1)
				atom.Attention.STI += transfer
				link.Attention.STI -= transfer
			}
		}
	}
	
	// Update attention bank heaps
	as.AttentionBank.Update()
}

// Forget removes low-importance atoms (forgetting mechanism)
func (as *AtomSpace) Forget() {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	forgettingThreshold := as.AttentionBank.AFBoundary * as.AttentionBank.ForgettingRate
	
	for id, atom := range as.Atoms {
		if atom.Attention.AF < forgettingThreshold {
			// Remove atom and its connections
			delete(as.Atoms, id)
			delete(as.Incoming, id)
			
			// Remove from attention bank
			as.AttentionBank.UnregisterAtom(id)
			
			as.AtomCount--
		}
	}
	
	// Clean up orphaned links
	for id, link := range as.Links {
		hasOrphan := false
		for _, atomID := range link.Outgoing {
			if _, exists := as.Atoms[atomID]; !exists {
				hasOrphan = true
				break
			}
		}
		
		if hasOrphan {
			delete(as.Links, id)
			as.AttentionBank.UnregisterAtom(id)
			as.LinkCount--
		}
	}
	
	as.Modified = time.Now()
}

// Query performs pattern matching
func (as *AtomSpace) Query(pattern *Pattern) (*QueryResult, error) {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	return as.PatternMatcher.Match(as, pattern)
}

// RegisterAtom registers an atom with the attention bank
func (ab *AttentionBank) RegisterAtom(id string, av *AttentionValue) {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	
	ab.STIHeap.values[id] = av.STI
	ab.LTIHeap.values[id] = av.LTI
}

// UnregisterAtom removes an atom from the attention bank
func (ab *AttentionBank) UnregisterAtom(id string) {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	
	delete(ab.STIHeap.values, id)
	delete(ab.LTIHeap.values, id)
}

// Update updates the attention bank heaps
func (ab *AttentionBank) Update() {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	
	// Rebuild heaps (simplified - in production use proper heap operations)
	ab.STIHeap.atoms = make([]string, 0, len(ab.STIHeap.values))
	for id := range ab.STIHeap.values {
		ab.STIHeap.atoms = append(ab.STIHeap.atoms, id)
	}
	
	ab.LTIHeap.atoms = make([]string, 0, len(ab.LTIHeap.values))
	for id := range ab.LTIHeap.values {
		ab.LTIHeap.atoms = append(ab.LTIHeap.atoms, id)
	}
}

// Match performs pattern matching
func (pm *PatternMatcher) Match(as *AtomSpace, pattern *Pattern) (*QueryResult, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// Simplified pattern matching - full implementation would be more sophisticated
	result := &QueryResult{
		Bindings:  []map[string]string{},
		Timestamp: time.Now(),
		Count:     0,
	}
	
	// For each clause, find matching atoms/links
	for _, clause := range pattern.Clauses {
		matches := pm.matchClause(as, clause, pattern.Variables)
		if len(matches) > 0 {
			result.Bindings = append(result.Bindings, matches...)
			result.Count += len(matches)
		}
	}
	
	return result, nil
}

// matchClause matches a single clause
func (pm *PatternMatcher) matchClause(as *AtomSpace, clause Clause, variables map[string]*Variable) []map[string]string {
	matches := []map[string]string{}
	
	// Search through all links of the specified type
	for _, link := range as.Links {
		if link.Type == clause.LinkType {
			binding := make(map[string]string)
			matched := true
			
			// Check if outgoing atoms match the pattern
			if len(link.Outgoing) == len(clause.Atoms) {
				for i, patternAtom := range clause.Atoms {
					if _, isVar := variables[patternAtom]; isVar {
						// Variable binding
						binding[patternAtom] = link.Outgoing[i]
					} else {
						// Concrete atom - must match exactly
						if link.Outgoing[i] != patternAtom {
							matched = false
							break
						}
					}
				}
				
				if matched {
					matches = append(matches, binding)
				}
			}
		}
	}
	
	return matches
}

// ComputeTruthValue computes PLN truth value fusion
func ComputeTruthValue(tv1, tv2 *TruthValue, operation string) *TruthValue {
	switch operation {
	case "and":
		return &TruthValue{
			Strength:   tv1.Strength * tv2.Strength,
			Confidence: math.Min(tv1.Confidence, tv2.Confidence),
			Count:      tv1.Count + tv2.Count,
		}
	case "or":
		return &TruthValue{
			Strength:   tv1.Strength + tv2.Strength - tv1.Strength*tv2.Strength,
			Confidence: math.Min(tv1.Confidence, tv2.Confidence),
			Count:      tv1.Count + tv2.Count,
		}
	case "not":
		return &TruthValue{
			Strength:   1.0 - tv1.Strength,
			Confidence: tv1.Confidence,
			Count:      tv1.Count,
		}
	default:
		return &TruthValue{
			Strength:   (tv1.Strength + tv2.Strength) / 2.0,
			Confidence: (tv1.Confidence + tv2.Confidence) / 2.0,
			Count:      tv1.Count + tv2.Count,
		}
	}
}

// GetStatus returns AtomSpace status
func (as *AtomSpace) GetStatus() map[string]interface{} {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	return map[string]interface{}{
		"atoms":        as.AtomCount,
		"links":        as.LinkCount,
		"created":      as.Created,
		"modified":     as.Modified,
		"sti_funds":    as.AttentionBank.STIFunds,
		"lti_funds":    as.AttentionBank.LTIFunds,
		"af_boundary":  as.AttentionBank.AFBoundary,
	}
}
