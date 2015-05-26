package tree

import (
	"testing"
)

func TestStack(t *testing.T) {
	s := NewStack()

	s.Push(1)

	one := s.Pop()
	if one.(int) != 1 {
		t.Errorf("Popped %v but should have popped %v", one, 1)
	}
}

func TestStack2(t *testing.T) {
	s := NewStack()

	s.Push(1)
	s.Push(2)

	two := s.Pop()
	if two.(int) != 2 {
		t.Errorf("Popped %v but should have popped %v", two, 2)
	}

	one := s.Pop()
	if one.(int) != 1 {
		t.Errorf("Popped %v but should have popped %v", one, 1)
	}
}

func TestStackEmpty(t *testing.T) {
	s := NewStack()
	e := s.Pop()
	if e != nil {
		t.Errorf("Empty stack should return nil")
	}
}
