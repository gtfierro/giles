package tree

import (
	"testing"
)

func TestQueue(t *testing.T) {
	s := NewQueue()

	s.Push(1)

	one := s.Pop()
	if one.(int) != 1 {
		t.Errorf("Popped %v but should have popped %v", one, 1)
	}
}

func TestQueue2(t *testing.T) {
	s := NewQueue()

	s.Push(1)
	s.Push(2)

	one := s.Pop()
	if one.(int) != 1 {
		t.Errorf("Popped %v but should have popped %v", one, 1)
	}

	two := s.Pop()
	if two.(int) != 2 {
		t.Errorf("Popped %v but should have popped %v", two, 2)
	}

}

func TestQueueEmpty(t *testing.T) {
	s := NewQueue()
	e := s.Pop()
	if e != nil {
		t.Errorf("Empty queue should return nil")
	}
}
