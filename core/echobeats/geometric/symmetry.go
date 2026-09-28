// Package geometric implements symmetry groups for echo-consciousness
package geometric

import (
	"fmt"
)

// SymmetryClass represents the class of a symmetry operation.
type SymmetryClass string

const (
	SymmetryIdentity SymmetryClass = "identity"
	SymmetryFace     SymmetryClass = "face"
	SymmetryEdge     SymmetryClass = "edge"
)

// Symmetry represents a single symmetry operation.
type Symmetry struct {
	ID          int
	Class       SymmetryClass
	Axis        string // e.g., "V0-F0", "E0-E3"
	Angle       int    // 0, 120, 180, 240 degrees
	Permutation []int  // Vertex permutation (for tetrahedron: 4 elements)
}

// SymmetryGroup represents a group of symmetries.
type SymmetryGroup struct {
	Name       string
	Order      int // Number of symmetries
	Symmetries []Symmetry
}

// NewA4Group creates the alternating group A4 (tetrahedral rotations).
// This group has 12 elements: 1 identity, 8 face rotations, 3 edge rotations.
func NewA4Group() *SymmetryGroup {
	return &SymmetryGroup{
		Name:  "A4",
		Order: 12,
		Symmetries: []Symmetry{
			// Identity (1 element)
			{
				ID:          0,
				Class:       SymmetryIdentity,
				Axis:        "-",
				Angle:       0,
				Permutation: []int{0, 1, 2, 3},
			},

			// Face rotations (8 elements)
			// Face 0 (vertices 0,1,2): rotations around axis through V3
			{
				ID:          1,
				Class:       SymmetryFace,
				Axis:        "V3-F0",
				Angle:       120,
				Permutation: []int{1, 2, 0, 3}, // V0→V1, V1→V2, V2→V0, V3→V3
			},
			{
				ID:          2,
				Class:       SymmetryFace,
				Axis:        "V3-F0",
				Angle:       240,
				Permutation: []int{2, 0, 1, 3}, // V0→V2, V1→V0, V2→V1, V3→V3
			},

			// Face 1 (vertices 0,1,3): rotations around axis through V2
			{
				ID:          3,
				Class:       SymmetryFace,
				Axis:        "V2-F1",
				Angle:       120,
				Permutation: []int{1, 3, 2, 0}, // V0→V1, V1→V3, V2→V2, V3→V0
			},
			{
				ID:          4,
				Class:       SymmetryFace,
				Axis:        "V2-F1",
				Angle:       240,
				Permutation: []int{3, 0, 2, 1}, // V0→V3, V1→V0, V2→V2, V3→V1
			},

			// Face 2 (vertices 0,2,3): rotations around axis through V1
			{
				ID:          5,
				Class:       SymmetryFace,
				Axis:        "V1-F2",
				Angle:       120,
				Permutation: []int{2, 1, 3, 0}, // V0→V2, V1→V1, V2→V3, V3→V0
			},
			{
				ID:          6,
				Class:       SymmetryFace,
				Axis:        "V1-F2",
				Angle:       240,
				Permutation: []int{3, 1, 0, 2}, // V0→V3, V1→V1, V2→V0, V3→V2
			},

			// Face 3 (vertices 1,2,3): rotations around axis through V0
			{
				ID:          7,
				Class:       SymmetryFace,
				Axis:        "V0-F3",
				Angle:       120,
				Permutation: []int{0, 2, 3, 1}, // V0→V0, V1→V2, V2→V3, V3→V1
			},
			{
				ID:          8,
				Class:       SymmetryFace,
				Axis:        "V0-F3",
				Angle:       240,
				Permutation: []int{0, 3, 1, 2}, // V0→V0, V1→V3, V2→V1, V3→V2
			},

			// Edge rotations (3 elements)
			// 180° rotation through midpoints of opposite edges
			{
				ID:          9,
				Class:       SymmetryEdge,
				Axis:        "E0-E5", // Edge (V0,V1) opposite to edge (V2,V3)
				Angle:       180,
				Permutation: []int{1, 0, 3, 2}, // Swap V0↔V1, V2↔V3
			},
			{
				ID:          10,
				Class:       SymmetryEdge,
				Axis:        "E1-E4", // Edge (V0,V2) opposite to edge (V1,V3)
				Angle:       180,
				Permutation: []int{2, 3, 0, 1}, // Swap V0↔V2, V1↔V3
			},
			{
				ID:          11,
				Class:       SymmetryEdge,
				Axis:        "E2-E3", // Edge (V0,V3) opposite to edge (V1,V2)
				Angle:       180,
				Permutation: []int{3, 2, 1, 0}, // Swap V0↔V3, V1↔V2
			},
		},
	}
}

// GetSymmetry returns the symmetry at a given step (mod order).
func (sg *SymmetryGroup) GetSymmetry(step int) Symmetry {
	return sg.Symmetries[step%sg.Order]
}

// ApplyPermutation applies a permutation to a slice of states.
func ApplyPermutation(states []interface{}, permutation []int) []interface{} {
	if len(states) != len(permutation) {
		return states // Return unchanged if sizes don't match
	}

	result := make([]interface{}, len(states))
	for i, p := range permutation {
		result[i] = states[p]
	}
	return result
}

// GetSymmetryClass returns the class of a symmetry at a given step.
func (sg *SymmetryGroup) GetSymmetryClass(step int) SymmetryClass {
	return sg.GetSymmetry(step).Class
}

// IsExpressive returns true if the step is expressive (face rotation).
// Based on the 5/7 twin prime structure: face rotations are mostly expressive.
func (sg *SymmetryGroup) IsExpressive(step int) bool {
	sym := sg.GetSymmetry(step)

	// Identity is reflective
	if sym.Class == SymmetryIdentity {
		return false
	}

	// Edge rotations are reflective
	if sym.Class == SymmetryEdge {
		return false
	}

	// Face rotations are mostly expressive
	// Specific pattern for 7 expressive, 5 reflective:
	// Steps 1,2,3,5,7,9,11 are expressive
	// Steps 4,6,8,10,12 are reflective
	expressiveSteps := map[int]bool{
		1: true, 2: true, 3: true, 5: true, 7: true, 9: true, 11: true,
	}

	return expressiveSteps[step%12]
}

// GetTriadAlignment returns the triad number (0-3) for a given step.
// Triads: {1,5,9}, {2,6,10}, {3,7,11}, {4,8,12}
func GetTriadAlignment(step int) int {
	cycleStep := step % 12
	if cycleStep == 0 {
		cycleStep = 12
	}

	switch cycleStep {
	case 1, 5, 9:
		return 0
	case 2, 6, 10:
		return 1
	case 3, 7, 11:
		return 2
	case 4, 8, 12:
		return 3
	default:
		return -1
	}
}

// GetActiveStream returns which stream (0-2) is active at a given step.
// P1 (Affordance): steps {1,4,7,10}
// P2 (Relevance): steps {2,5,8,11}
// P3 (Salience): steps {3,6,9,12}
func GetActiveStream(step int) int {
	cycleStep := step % 12
	if cycleStep == 0 {
		cycleStep = 12
	}

	switch cycleStep {
	case 1, 4, 7, 10:
		return 0 // P1
	case 2, 5, 8, 11:
		return 1 // P2
	case 3, 6, 9, 12:
		return 2 // P3
	default:
		return -1
	}
}

// String returns a string representation of the symmetry group.
func (sg *SymmetryGroup) String() string {
	return fmt.Sprintf("SymmetryGroup %s (order=%d)", sg.Name, sg.Order)
}

// NewA5Group creates the alternating group A5 (pentachoron rotations).
// This group has 60 elements (for sys5).
// Simplified implementation - would need full enumeration for production.
func NewA5Group() *SymmetryGroup {
	// For now, return a placeholder
	// Full implementation would enumerate all 60 rotational symmetries of the 4-simplex
	return &SymmetryGroup{
		Name:       "A5",
		Order:      60,
		Symmetries: make([]Symmetry, 60),
	}
}
