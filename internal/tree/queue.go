package tree

type Queue struct {
	c []interface{}
}

func NewQueue() *Queue {
	return &Queue{c: []interface{}{}}
}

func (s *Queue) Push(e interface{}) {
	s.c = append(s.c, e)
}

func (s *Queue) Pop() interface{} {
	var e interface{}
	length := len(s.c)
	if length == 0 {
		e = nil
	} else if length == 1 {
		e = s.c[0]
		s.c = s.c[:0]
	} else {
		e = s.c[0]
		s.c = s.c[length-1:]
	}
	return e
}
