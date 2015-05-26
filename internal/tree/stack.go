package tree

type Stack struct {
	c []interface{}
}

func NewStack() *Stack {
	return &Stack{c: []interface{}{}}
}

func (s *Stack) Push(e interface{}) {
	s.c = append(s.c, e)
}

func (s *Stack) Pop() interface{} {
	var e interface{}
	length := len(s.c)
	if length == 0 {
		e = nil
	} else if length == 1 {
		e = s.c[0]
		s.c = s.c[:0]
	} else {
		e = s.c[length-1]
		s.c = s.c[:length-1]
	}
	return e
}
