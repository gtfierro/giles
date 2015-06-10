package archiver

type Node struct {
	Id       string
	Tags     map[string]interface{}
	In       chan interface{}
	Done     <-chan struct{}
	Children map[string]*Node
	Op       Operator
}

func NewNode(operation Operator, done <-chan struct{}) (n *Node) {
	n = &Node{In: make(chan interface{}),
		Done:     done,
		Op:       operation,
		Tags:     make(map[string]interface{}),
		Children: make(map[string]*Node),
	}
	go func(n *Node) {
		for {
			select {
			case input := <-n.In:
				res, _ := n.Op.Run(input)
				for _, c := range n.Children {
					c.In <- res
				}
			case <-done:
				close(n.In)
			}
		}
	}(n)
	return
}

func (n *Node) AddChild(child *Node) bool {
	var found bool
	if _, found = n.Children[child.Id]; !found {
		n.Children[child.Id] = child
	}
	return found
}

func (n *Node) HasOutput(structure, datatype uint) (res bool) {
	res = true
	var found bool
	if structure != 0 {
		_, found = n.Tags["out:structure"]
		res = res && found
	}
	if datatype != 0 {
		_, found = n.Tags["out:datatype"]
		res = res && found
	}
	return
}

func (n *Node) HasInput(structure, datatype uint) (res bool) {
	res = true
	var found bool
	if structure != 0 {
		_, found = n.Tags["in:structure"]
		res = res && found
	}
	if datatype != 0 {
		_, found = n.Tags["in:datatype"]
		res = res && found
	}
	return
}
