package tree

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
)

type BaseNode struct {
	id       string
	children map[string]Node
	tags     map[string]interface{}
}

func NewBaseNode(kv map[string]interface{}) (bn *BaseNode, err error) {
	for _, v := range kv {
		switch v.(type) {
		case uint64, float64, int64, string:
		default:
			err = fmt.Errorf("Value %v must be uint64, int64, float64 or string", v)
			return
		}
	}
	bn = &BaseNode{
		id:       uuid.New(),
		tags:     kv,
		children: make(map[string]Node, 4),
	}
	return
}

func InitBaseNode(bn *BaseNode) (err error) {
	bn.id = uuid.New()
	bn.tags = make(map[string]interface{})
	bn.children = make(map[string]Node, 4)
	return
}

func (bn *BaseNode) Id() string {
	return bn.id
}

func (bn *BaseNode) Children() map[string]Node {
	return bn.children
}

func (bn *BaseNode) AddChild(n Node) bool {
	var found bool
	if _, found = bn.children[n.Id()]; !found {
		bn.children[n.Id()] = n
	}
	return found
}

func (bn *BaseNode) Input(args ...interface{}) error {
	return fmt.Errorf("BaseNode has no Input")
}

func (bn *BaseNode) Output() (interface{}, error) {
	fmt.Printf("Node kv: %v\n", bn.tags)
	return nil, fmt.Errorf("BaseNode has no Output")
}

func (bn *BaseNode) Get(key string) (val interface{}, found bool) {
	val, found = bn.tags[key]
	return
}

func (bn *BaseNode) Set(key string, val interface{}) {
	bn.tags[key] = val
}

func (bn *BaseNode) HasOutput(structure, datatype uint) (res bool) {
	res = true
	var found bool
	if structure != 0 {
		_, found = bn.tags["out:structure"]
		res = res && found
	}
	if datatype != 0 {
		_, found = bn.tags["out:datatype"]
		res = res && found
	}
	return
}

func (bn *BaseNode) HasInput(structure, datatype uint) (res bool) {
	res = true
	var found bool
	if structure != 0 {
		_, found = bn.tags["in:structure"]
		res = res && found
	}
	if datatype != 0 {
		_, found = bn.tags["in:datatype"]
		res = res && found
	}
	return
}
