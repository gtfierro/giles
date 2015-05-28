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

func InitBaseNode(bn *BaseNode, kv map[string]interface{}) (err error) {
	for _, v := range kv {
		switch v.(type) {
		case uint64, float64, int64, string:
		default:
			err = fmt.Errorf("Value %v must be uint64, int64, float64 or string", v)
			return
		}
	}
	bn.id = uuid.New()
	bn.tags = kv
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
	return nil, fmt.Errorf("BaseNode has no Output")
}

// Starts Run() from this node. Assumes Input() has already been run
func (bn *BaseNode) Run() (err error) {
	var (
		next     interface{}
		nextNode Node
		q        = NewQueue()
		output   interface{}
	)
	q.Push(bn)
	for {
		// pop next node off of queue
		next = q.Pop()

		// check if we are done
		if next == nil {
			return
		}

		// assert type
		nextNode = next.(Node)
		// get Output
		output, err = nextNode.Output()
		fmt.Printf("got output %v", output)
		if err != nil {
			return
		}
		for _, childNode := range nextNode.Children() {
			err = childNode.Input(output)
			if err != nil {
				return
			}
			q.Push(childNode)
		}
	}
}

func (bn *BaseNode) Get(key string) (val interface{}, found bool) {
	val, found = bn.tags[key]
	return
}
