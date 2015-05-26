package tree

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
)

type baseNode struct {
	id       string
	children map[string]Node
	tags     map[string]interface{}
}

func NewBaseNode(kv map[string]interface{}) (bn *baseNode, err error) {
	for _, v := range kv {
		switch v.(type) {
		case uint64, float64, int64, string:
		default:
			err = fmt.Errorf("Value %v must be uint64, int64, float64 or string", v)
			return
		}
	}
	bn = &baseNode{
		id:       uuid.New(),
		tags:     kv,
		children: make(map[string]Node, 4),
	}
	return
}

func (bn *baseNode) Id() string {
	return bn.id
}

func (bn *baseNode) Children() map[string]Node {
	return bn.children
}

func (bn *baseNode) AddChild(n Node) bool {
	var found bool
	if _, found = bn.children[n.Id()]; !found {
		bn.children[n.Id()] = n
	}
	return found
}
