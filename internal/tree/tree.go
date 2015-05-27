package tree

import (
	"fmt"
)

// Each type of node in the tree will implement this interface
type Node interface {
	// Returns the uuid for this node.
	Id() string
	// mapping of Id -> child Node for all
	// immediate children of this node
	Children() map[string]Node
	// adds the given node as a child of this node
	AddChild(n Node) bool
	// Pass input to this node
	Input(args ...interface{}) error
	// Get output from the node
	Output() (interface{}, error)
	// See Tree.Run()
	Run() error
}

// Stores a direction from parent node to child node
type Edge struct {
	Parent string
	Child  string
}

type Tree struct {
	Root  Node
	Nodes map[string]Node
	Edges []*Edge
}

// A tree has a Root node, which should be an Output-only node. Trees can have
// nodes added, and edges added in between nodes. Each tree is a directed,
// rooted tree, and each operation on the tree should enforce that.
func NewTree(root Node) (t *Tree) {
	t = &Tree{
		Root:  root,
		Nodes: make(map[string]Node, 16),
	}
	t.Nodes[root.Id()] = root
	return
}

// Adds the node to the tree. Returns true if the node is already in the tree,
// and false otherwise
func (t *Tree) AddNode(n Node) bool {
	var found bool
	if _, found = t.Nodes[n.Id()]; !found {
		t.Nodes[n.Id()] = n
	}
	return found
}

// Returns the node in the tree with the given id. Returns nil
// if the node is not found
func (t *Tree) GetNode(id string) Node {
	n, _ := t.Nodes[id]
	return n
}

// Adds a directed edge from parent to child. Parent should be a node already in
// the tree, though child does not have to be. Return true if the edge already
// exists, else false. This method will check for cycles in the tree and will
// return an error if it finds one, or if the parent node is not in the tree
func (t *Tree) AddChild(parent, child Node) (bool, error) {
	// check that parent node is already in tree
	foundParent := t.GetNode(parent.Id())
	if foundParent == nil {
		return false, fmt.Errorf("Parent node with Id %v not found in tree", parent.Id())
	}

	// check if the edge already exists
	if _, foundEdge := parent.Children()[child.Id()]; foundEdge {
		// if it does, we do nothing
		return true, nil
	}

	// add child to tree
	t.AddNode(child)

	// add edge from parent to child
	parent.AddChild(child)

	if t.HasCycle() {
		return false, fmt.Errorf("Adding edge from %v to %v formed a cycle", parent.Id(), child.Id())
	}

	return false, nil
}

// Does depth first search to find cycles. Returns true
// if cycle is found
func (t *Tree) HasCycle() (hasCycle bool) {
	var (
		next     interface{}
		nextNode Node
		s        = NewStack()
		seen     = make(map[string]struct{})
	)
	hasCycle = false
	s.Push(t.Root)

	for {
		// pop next node off of stack
		next = s.Pop()

		// check if we are done iterating

		if next == nil {
			return // done with iteration
		}

		// assert type
		nextNode = next.(Node)

		if _, found := seen[nextNode.Id()]; found {
			hasCycle = true
			return
		} else {
			seen[nextNode.Id()] = struct{}{}
			for _, childNode := range nextNode.Children() {
				s.Push(childNode)
			}
		}
	}
}

// Run() starts at the root, then iterates through nodes following BFS. Starting with the root,
// it runs Output() to get the results, then feeds the output into each of the Input() of the children
// nodes, then runs Output() on each of those children, etc. To be specific, the Output() of a parent
// node is fed to Input() on each child nodes when those nodes are added to the stack/queue, and Output()
// is called when they are popped from that queue
func (t *Tree) Run() (err error) {
	return t.Root.Run()
}
