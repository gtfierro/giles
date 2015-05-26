package tree

import (
	"testing"
)

func TestMakeBaseNode(t *testing.T) {
	bn, err := NewBaseNode(nil)
	if bn.id == "" {
		t.Error("Node ID was nil")
	}
	if err != nil {
		t.Error(err)
	}
}

func TestMakeTree(t *testing.T) {
	root, _ := NewBaseNode(nil)
	tree := NewTree(root)

	if tree.Root.Id() != root.Id() {
		t.Errorf("Tree has incorrect root %v -- should be %v", tree.Root.Id(), root.Id())
	}

	if foundNode, found := tree.Nodes[root.Id()]; found {
		if foundNode.Id() != root.Id() {
			t.Errorf("Tree has incorrect root %v -- should be %v", foundNode.Id(), root.Id())
		}
	} else {
		t.Errorf("Root node %v not found in tree.", root.Id())
	}
}

func TestTreeAddNode(t *testing.T) {
	root, _ := NewBaseNode(nil)
	tree := NewTree(root)

	newNode, _ := NewBaseNode(nil)
	alreadyThere := tree.AddNode(newNode)
	if alreadyThere {
		t.Errorf("Node should not already be in tree")
	}

	// test for root
	if foundNode, found := tree.Nodes[root.Id()]; found {
		if foundNode.Id() != root.Id() {
			t.Errorf("Tree has incorrect root %v -- should be %v", foundNode.Id(), root.Id())
		}
	} else {
		t.Errorf("Root node %v not found in tree.", root.Id())
	}

	// test for newNode
	if foundNode, found := tree.Nodes[newNode.Id()]; found {
		if foundNode.Id() != newNode.Id() {
			t.Errorf("Tree has incorrect newNode %v -- should be %v", foundNode.Id(), newNode.Id())
		}
	} else {
		t.Errorf("newNode node %v not found in tree.", newNode.Id())
	}
}

func TestTreeAddNodeTwice(t *testing.T) {
	root, _ := NewBaseNode(nil)
	tree := NewTree(root)

	alreadyThere := tree.AddNode(root)
	if !alreadyThere {
		t.Errorf("Node should already be in tree")
	}
}

func TestTreeAddChild(t *testing.T) {
	root, _ := NewBaseNode(nil)
	tree := NewTree(root)

	newNode, _ := NewBaseNode(nil)
	tree.AddNode(newNode)

	there, err := tree.AddChild(root, newNode)
	if err != nil {
		t.Errorf("Error adding edge (%v)", err)
	}
	if there {
		t.Errorf("Edge should not already be in tree")
	}

	if _, found := root.Children()[newNode.Id()]; !found {
		t.Errorf("Added child %v was not in parent's children", newNode.Id())
	}
}

func TestTreeHasCyclePositive(t *testing.T) {
	root, _ := NewBaseNode(nil)
	tree := NewTree(root)

	newNode, _ := NewBaseNode(nil)
	tree.AddNode(newNode)

	_, err := tree.AddChild(root, newNode)
	if err != nil {
		t.Errorf("Error adding edge (%v)", err)
	}

	_, err = tree.AddChild(newNode, root) // should form cycle
	if err == nil {
		t.Errorf("Should have detected cycle in tree")
	}
}

func TestTreeHasCycleLargerPositive(t *testing.T) {
	root, _ := NewBaseNode(nil)
	tree := NewTree(root)

	node1, _ := NewBaseNode(nil)
	tree.AddNode(node1)

	_, err := tree.AddChild(root, node1)
	if err != nil {
		t.Errorf("Error adding edge (%v)", err)
	}

	node2, _ := NewBaseNode(nil)
	tree.AddNode(node2)

	_, err = tree.AddChild(node1, node2)
	if err != nil {
		t.Errorf("Error adding edge (%v)", err)
	}

	_, err = tree.AddChild(node2, root) // should form cycle
	if err == nil {
		t.Errorf("Should have detected cycle in tree")
	}
}
