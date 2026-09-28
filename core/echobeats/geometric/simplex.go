// Package geometric implements the geometric foundations of echo-consciousness
// based on simplex geometry, integer partitions, and rooted trees.
package geometric

import (
	"fmt"
)

// SimplexDimension represents the dimension of a simplex element.
type SimplexDimension int

const (
	DimVoid      SimplexDimension = -1 // Null set
	DimVertex    SimplexDimension = 0  // Point
	DimEdge      SimplexDimension = 1  // Line segment
	DimFace      SimplexDimension = 2  // Triangle
	DimCell      SimplexDimension = 3  // Tetrahedron
	DimHypercell SimplexDimension = 4  // 5-cell (pentachoron)
)

// Simplex represents an n-simplex with its dimensional elements.
type Simplex struct {
	Dimension  int         // The dimension n of the n-simplex
	Vertices   []Vertex    // dim(0) elements
	Edges      []Edge      // dim(1) elements
	Faces      []Face      // dim(2) elements
	Cells      []Cell      // dim(3) elements
	Hypercells []Hypercell // dim(4) elements
}

// Vertex represents a 0-dimensional element (point).
type Vertex struct {
	ID    int
	Label string
	State interface{} // Can hold any state type
}

// Edge represents a 1-dimensional element (line segment between 2 vertices).
type Edge struct {
	ID       int
	Label    string
	Vertices [2]int // Indices of the 2 vertices
}

// Face represents a 2-dimensional element (triangle between 3 vertices).
type Face struct {
	ID       int
	Label    string
	Vertices [3]int // Indices of the 3 vertices
	Edges    [3]int // Indices of the 3 edges
}

// Cell represents a 3-dimensional element (tetrahedron between 4 vertices).
type Cell struct {
	ID       int
	Label    string
	Vertices [4]int // Indices of the 4 vertices
	Edges    [6]int // Indices of the 6 edges
	Faces    [4]int // Indices of the 4 faces
}

// Hypercell represents a 4-dimensional element (5-cell between 5 vertices).
type Hypercell struct {
	ID       int
	Label    string
	Vertices [5]int  // Indices of the 5 vertices
	Edges    [10]int // Indices of the 10 edges
	Faces    [10]int // Indices of the 10 faces
	Cells    [5]int  // Indices of the 5 cells
}

// NewSimplex creates a new n-simplex with the appropriate structure.
func NewSimplex(n int) (*Simplex, error) {
	if n < 0 || n > 5 {
		return nil, fmt.Errorf("simplex dimension must be between 0 and 5, got %d", n)
	}

	s := &Simplex{
		Dimension: n,
	}

	switch n {
	case 0:
		s.initSimplex0()
	case 1:
		s.initSimplex1()
	case 2:
		s.initSimplex2()
	case 3:
		s.initSimplex3()
	case 4:
		s.initSimplex4()
	}

	return s, nil
}

// initSimplex0 initializes a 0-simplex (point).
func (s *Simplex) initSimplex0() {
	s.Vertices = []Vertex{
		{ID: 0, Label: "V0"},
	}
}

// initSimplex1 initializes a 1-simplex (line segment).
func (s *Simplex) initSimplex1() {
	s.Vertices = []Vertex{
		{ID: 0, Label: "V0"},
		{ID: 1, Label: "V1"},
	}
	s.Edges = []Edge{
		{ID: 0, Label: "E0", Vertices: [2]int{0, 1}},
	}
}

// initSimplex2 initializes a 2-simplex (triangle).
func (s *Simplex) initSimplex2() {
	s.Vertices = []Vertex{
		{ID: 0, Label: "V0"},
		{ID: 1, Label: "V1"},
		{ID: 2, Label: "V2"},
	}
	s.Edges = []Edge{
		{ID: 0, Label: "E0", Vertices: [2]int{0, 1}},
		{ID: 1, Label: "E1", Vertices: [2]int{1, 2}},
		{ID: 2, Label: "E2", Vertices: [2]int{2, 0}},
	}
	s.Faces = []Face{
		{ID: 0, Label: "F0", Vertices: [3]int{0, 1, 2}, Edges: [3]int{0, 1, 2}},
	}
}

// initSimplex3 initializes a 3-simplex (tetrahedron).
func (s *Simplex) initSimplex3() {
	s.Vertices = []Vertex{
		{ID: 0, Label: "V0"},
		{ID: 1, Label: "V1"},
		{ID: 2, Label: "V2"},
		{ID: 3, Label: "V3"},
	}
	s.Edges = []Edge{
		{ID: 0, Label: "E0", Vertices: [2]int{0, 1}},
		{ID: 1, Label: "E1", Vertices: [2]int{0, 2}},
		{ID: 2, Label: "E2", Vertices: [2]int{0, 3}},
		{ID: 3, Label: "E3", Vertices: [2]int{1, 2}},
		{ID: 4, Label: "E4", Vertices: [2]int{1, 3}},
		{ID: 5, Label: "E5", Vertices: [2]int{2, 3}},
	}
	s.Faces = []Face{
		{ID: 0, Label: "F0", Vertices: [3]int{0, 1, 2}, Edges: [3]int{0, 3, 1}},
		{ID: 1, Label: "F1", Vertices: [3]int{0, 1, 3}, Edges: [3]int{0, 4, 2}},
		{ID: 2, Label: "F2", Vertices: [3]int{0, 2, 3}, Edges: [3]int{1, 5, 2}},
		{ID: 3, Label: "F3", Vertices: [3]int{1, 2, 3}, Edges: [3]int{3, 5, 4}},
	}
	s.Cells = []Cell{
		{ID: 0, Label: "C0", Vertices: [4]int{0, 1, 2, 3}, Edges: [6]int{0, 1, 2, 3, 4, 5}, Faces: [4]int{0, 1, 2, 3}},
	}
}

// initSimplex4 initializes a 4-simplex (5-cell / pentachoron).
func (s *Simplex) initSimplex4() {
	s.Vertices = []Vertex{
		{ID: 0, Label: "V0"},
		{ID: 1, Label: "V1"},
		{ID: 2, Label: "V2"},
		{ID: 3, Label: "V3"},
		{ID: 4, Label: "V4"},
	}
	// 10 edges (all pairs of 5 vertices)
	s.Edges = []Edge{
		{ID: 0, Label: "E0", Vertices: [2]int{0, 1}},
		{ID: 1, Label: "E1", Vertices: [2]int{0, 2}},
		{ID: 2, Label: "E2", Vertices: [2]int{0, 3}},
		{ID: 3, Label: "E3", Vertices: [2]int{0, 4}},
		{ID: 4, Label: "E4", Vertices: [2]int{1, 2}},
		{ID: 5, Label: "E5", Vertices: [2]int{1, 3}},
		{ID: 6, Label: "E6", Vertices: [2]int{1, 4}},
		{ID: 7, Label: "E7", Vertices: [2]int{2, 3}},
		{ID: 8, Label: "E8", Vertices: [2]int{2, 4}},
		{ID: 9, Label: "E9", Vertices: [2]int{3, 4}},
	}
	// 10 faces (all triplets of 5 vertices)
	s.Faces = []Face{
		{ID: 0, Label: "F0", Vertices: [3]int{0, 1, 2}, Edges: [3]int{0, 4, 1}},
		{ID: 1, Label: "F1", Vertices: [3]int{0, 1, 3}, Edges: [3]int{0, 5, 2}},
		{ID: 2, Label: "F2", Vertices: [3]int{0, 1, 4}, Edges: [3]int{0, 6, 3}},
		{ID: 3, Label: "F3", Vertices: [3]int{0, 2, 3}, Edges: [3]int{1, 7, 2}},
		{ID: 4, Label: "F4", Vertices: [3]int{0, 2, 4}, Edges: [3]int{1, 8, 3}},
		{ID: 5, Label: "F5", Vertices: [3]int{0, 3, 4}, Edges: [3]int{2, 9, 3}},
		{ID: 6, Label: "F6", Vertices: [3]int{1, 2, 3}, Edges: [3]int{4, 7, 5}},
		{ID: 7, Label: "F7", Vertices: [3]int{1, 2, 4}, Edges: [3]int{4, 8, 6}},
		{ID: 8, Label: "F8", Vertices: [3]int{1, 3, 4}, Edges: [3]int{5, 9, 6}},
		{ID: 9, Label: "F9", Vertices: [3]int{2, 3, 4}, Edges: [3]int{7, 9, 8}},
	}
	// 5 cells (all quadruplets of 5 vertices)
	s.Cells = []Cell{
		{ID: 0, Label: "C0", Vertices: [4]int{0, 1, 2, 3}, Edges: [6]int{0, 1, 2, 4, 5, 7}, Faces: [4]int{0, 1, 3, 6}},
		{ID: 1, Label: "C1", Vertices: [4]int{0, 1, 2, 4}, Edges: [6]int{0, 1, 3, 4, 6, 8}, Faces: [4]int{0, 2, 4, 7}},
		{ID: 2, Label: "C2", Vertices: [4]int{0, 1, 3, 4}, Edges: [6]int{0, 2, 3, 5, 6, 9}, Faces: [4]int{1, 2, 5, 8}},
		{ID: 3, Label: "C3", Vertices: [4]int{0, 2, 3, 4}, Edges: [6]int{1, 2, 3, 7, 8, 9}, Faces: [4]int{3, 4, 5, 9}},
		{ID: 4, Label: "C4", Vertices: [4]int{1, 2, 3, 4}, Edges: [6]int{4, 5, 6, 7, 8, 9}, Faces: [4]int{6, 7, 8, 9}},
	}
	// 1 hypercell (all 5 vertices)
	s.Hypercells = []Hypercell{
		{ID: 0, Label: "H0", Vertices: [5]int{0, 1, 2, 3, 4}, Edges: [10]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, Faces: [10]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, Cells: [5]int{0, 1, 2, 3, 4}},
	}
}

// EulerCharacteristic computes the Euler characteristic of the simplex.
func (s *Simplex) EulerCharacteristic() int {
	chi := 1 // Start with the void (dim -1)
	chi -= len(s.Vertices)
	chi += len(s.Edges)
	chi -= len(s.Faces)
	chi += len(s.Cells)
	chi -= len(s.Hypercells)
	return chi
}

// GetVertexCount returns the number of vertices.
func (s *Simplex) GetVertexCount() int {
	return len(s.Vertices)
}

// GetEdgeCount returns the number of edges.
func (s *Simplex) GetEdgeCount() int {
	return len(s.Edges)
}

// GetFaceCount returns the number of faces.
func (s *Simplex) GetFaceCount() int {
	return len(s.Faces)
}

// GetCellCount returns the number of cells.
func (s *Simplex) GetCellCount() int {
	return len(s.Cells)
}

// GetHypercellCount returns the number of hypercells.
func (s *Simplex) GetHypercellCount() int {
	return len(s.Hypercells)
}

// String returns a string representation of the simplex.
func (s *Simplex) String() string {
	return fmt.Sprintf("%d-simplex: %d vertices, %d edges, %d faces, %d cells, %d hypercells (χ=%d)",
		s.Dimension,
		len(s.Vertices),
		len(s.Edges),
		len(s.Faces),
		len(s.Cells),
		len(s.Hypercells),
		s.EulerCharacteristic())
}
