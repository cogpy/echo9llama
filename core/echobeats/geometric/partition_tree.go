// Package geometric implements partition and tree structures for echo-consciousness
package geometric

import (
	"fmt"
	"sort"
)

// Partition represents an integer partition.
type Partition struct {
	Components []int  // e.g., [2, 1, 1] for {2+1+1}
	Weight     int    // Sum of components
	Type       string // "universal" or "particular"
}

// NewPartition creates a new partition from components.
func NewPartition(components []int, partitionType string) *Partition {
	weight := 0
	for _, c := range components {
		weight += c
	}

	// Sort components in descending order (standard form)
	sorted := make([]int, len(components))
	copy(sorted, components)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))

	return &Partition{
		Components: sorted,
		Weight:     weight,
		Type:       partitionType,
	}
}

// String returns a string representation of the partition.
func (p *Partition) String() string {
	return fmt.Sprintf("{%v} (weight=%d, type=%s)", p.Components, p.Weight, p.Type)
}

// PartitionSet represents a set of partitions for a given n.
type PartitionSet struct {
	N          int
	Partitions []*Partition
}

// GeneratePartitions generates all integer partitions of n.
// This is a simplified implementation for n <= 5.
func GeneratePartitions(n int) (*PartitionSet, error) {
	if n < 0 || n > 5 {
		return nil, fmt.Errorf("partition generation only supported for n in [0, 5], got %d", n)
	}

	ps := &PartitionSet{
		N:          n,
		Partitions: make([]*Partition, 0),
	}

	switch n {
	case 0:
		// Empty partition
	case 1:
		ps.Partitions = append(ps.Partitions, NewPartition([]int{1}, "universal"))
	case 2:
		ps.Partitions = append(ps.Partitions, NewPartition([]int{2}, "universal"))
		ps.Partitions = append(ps.Partitions, NewPartition([]int{1, 1}, "particular"))
	case 3:
		ps.Partitions = append(ps.Partitions, NewPartition([]int{3}, "universal"))
		ps.Partitions = append(ps.Partitions, NewPartition([]int{2, 1}, "universal"))
		ps.Partitions = append(ps.Partitions, NewPartition([]int{1, 1, 1}, "particular"))
	case 4:
		ps.Partitions = append(ps.Partitions, NewPartition([]int{4}, "universal"))
		ps.Partitions = append(ps.Partitions, NewPartition([]int{3, 1}, "universal"))
		ps.Partitions = append(ps.Partitions, NewPartition([]int{2, 2}, "particular"))
		ps.Partitions = append(ps.Partitions, NewPartition([]int{2, 1, 1}, "particular"))
		ps.Partitions = append(ps.Partitions, NewPartition([]int{1, 1, 1, 1}, "particular"))
	case 5:
		ps.Partitions = append(ps.Partitions, NewPartition([]int{5}, "universal"))
		ps.Partitions = append(ps.Partitions, NewPartition([]int{4, 1}, "universal"))
		ps.Partitions = append(ps.Partitions, NewPartition([]int{3, 2}, "universal"))
		ps.Partitions = append(ps.Partitions, NewPartition([]int{3, 1, 1}, "particular"))
		ps.Partitions = append(ps.Partitions, NewPartition([]int{2, 2, 1}, "particular"))
		ps.Partitions = append(ps.Partitions, NewPartition([]int{2, 1, 1, 1}, "particular"))
		ps.Partitions = append(ps.Partitions, NewPartition([]int{1, 1, 1, 1, 1}, "particular"))
	}

	return ps, nil
}

// GetUniversalPartitions returns only the universal partitions.
func (ps *PartitionSet) GetUniversalPartitions() []*Partition {
	result := make([]*Partition, 0)
	for _, p := range ps.Partitions {
		if p.Type == "universal" {
			result = append(result, p)
		}
	}
	return result
}

// GetParticularPartitions returns only the particular partitions.
func (ps *PartitionSet) GetParticularPartitions() []*Partition {
	result := make([]*Partition, 0)
	for _, p := range ps.Partitions {
		if p.Type == "particular" {
			result = append(result, p)
		}
	}
	return result
}

// RootedTree represents a rooted tree structure.
type RootedTree struct {
	ID    int
	Nodes []*TreeNode
	Root  *TreeNode
	Depth int
}

// TreeNode represents a node in a rooted tree.
type TreeNode struct {
	ID       int
	Label    string
	Parent   *TreeNode
	Children []*TreeNode
	Depth    int
}

// NewRootedTree creates a new rooted tree.
func NewRootedTree(id int) *RootedTree {
	root := &TreeNode{
		ID:       0,
		Label:    "Root",
		Parent:   nil,
		Children: make([]*TreeNode, 0),
		Depth:    0,
	}

	return &RootedTree{
		ID:    id,
		Nodes: []*TreeNode{root},
		Root:  root,
		Depth: 0,
	}
}

// AddChild adds a child node to a parent node.
func (rt *RootedTree) AddChild(parent *TreeNode, label string) *TreeNode {
	child := &TreeNode{
		ID:       len(rt.Nodes),
		Label:    label,
		Parent:   parent,
		Children: make([]*TreeNode, 0),
		Depth:    parent.Depth + 1,
	}

	parent.Children = append(parent.Children, child)
	rt.Nodes = append(rt.Nodes, child)

	if child.Depth > rt.Depth {
		rt.Depth = child.Depth
	}

	return child
}

// GetNodeCount returns the number of nodes in the tree.
func (rt *RootedTree) GetNodeCount() int {
	return len(rt.Nodes)
}

// TreeSet represents a set of rooted trees for a given n.
type TreeSet struct {
	N     int
	Trees []*RootedTree
}

// GenerateTrees generates representative rooted trees for n nodes.
// This is a simplified implementation for n <= 5.
func GenerateTrees(n int) (*TreeSet, error) {
	if n < 0 || n > 5 {
		return nil, fmt.Errorf("tree generation only supported for n in [0, 5], got %d", n)
	}

	ts := &TreeSet{
		N:     n,
		Trees: make([]*RootedTree, 0),
	}

	switch n {
	case 0:
		// No tree
	case 1:
		// Single root
		ts.Trees = append(ts.Trees, NewRootedTree(0))
	case 2:
		// Root with 1 child
		tree := NewRootedTree(0)
		tree.AddChild(tree.Root, "N1")
		ts.Trees = append(ts.Trees, tree)
	case 3:
		// Tree 1: Linear chain
		tree1 := NewRootedTree(0)
		n1 := tree1.AddChild(tree1.Root, "N1")
		tree1.AddChild(n1, "N2")
		ts.Trees = append(ts.Trees, tree1)

		// Tree 2: Branching at root
		tree2 := NewRootedTree(1)
		tree2.AddChild(tree2.Root, "N1")
		tree2.AddChild(tree2.Root, "N2")
		ts.Trees = append(ts.Trees, tree2)
	case 4:
		// Tree 1: Linear chain
		tree1 := NewRootedTree(0)
		n1 := tree1.AddChild(tree1.Root, "N1")
		n2 := tree1.AddChild(n1, "N2")
		tree1.AddChild(n2, "N3")
		ts.Trees = append(ts.Trees, tree1)

		// Tree 2: Branch at end
		tree2 := NewRootedTree(1)
		n1 = tree2.AddChild(tree2.Root, "N1")
		n2 = tree2.AddChild(n1, "N2")
		tree2.AddChild(n2, "N3a")
		tree2.AddChild(n2, "N3b")
		ts.Trees = append(ts.Trees, tree2)

		// Tree 3: Branch in middle
		tree3 := NewRootedTree(2)
		n1 = tree3.AddChild(tree3.Root, "N1")
		tree3.AddChild(n1, "N2a")
		n2 = tree3.AddChild(n1, "N2b")
		tree3.AddChild(n2, "N3")
		ts.Trees = append(ts.Trees, tree3)

		// Tree 4: Branching at root
		tree4 := NewRootedTree(3)
		tree4.AddChild(tree4.Root, "N1")
		tree4.AddChild(tree4.Root, "N2")
		tree4.AddChild(tree4.Root, "N3")
		ts.Trees = append(ts.Trees, tree4)
	case 5:
		// For n=5, we would have 9 distinct rooted trees
		// Simplified: just create a few representative ones

		// Tree 1: Linear chain
		tree1 := NewRootedTree(0)
		n1 := tree1.AddChild(tree1.Root, "N1")
		n2 := tree1.AddChild(n1, "N2")
		n3 := tree1.AddChild(n2, "N3")
		tree1.AddChild(n3, "N4")
		ts.Trees = append(ts.Trees, tree1)

		// Tree 2: Branching structure
		tree2 := NewRootedTree(1)
		n1 = tree2.AddChild(tree2.Root, "N1")
		tree2.AddChild(n1, "N2")
		tree2.AddChild(n1, "N3")
		tree2.AddChild(n1, "N4")
		ts.Trees = append(ts.Trees, tree2)

		// Add more trees as needed...
	}

	return ts, nil
}

// GetMaxDepth returns the maximum depth among all trees.
func (ts *TreeSet) GetMaxDepth() int {
	maxDepth := 0
	for _, tree := range ts.Trees {
		if tree.Depth > maxDepth {
			maxDepth = tree.Depth
		}
	}
	return maxDepth
}

// GetTotalTerms returns the total number of terms (nodes) across all trees.
func (ts *TreeSet) GetTotalTerms() int {
	total := 0
	for _, tree := range ts.Trees {
		total += tree.GetNodeCount()
	}
	return total
}
