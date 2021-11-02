package main

type Stack struct {
	values []*Value
}

func (s *Stack) Len() int {
	return len(s.values)
}

func (s *Stack) Push(n *Value) {
	s.values = append(s.values, n)
}

func (s *Stack) Peek() *Value {
	l := len(s.values)
	if l == 0 {
		return nil
	}
	return s.values[l-1]
}

func (s *Stack) Pop() *Value {
	l := len(s.values)
	if l == 0 {
		return nil
	}
	val := s.Peek()
	s.values = s.values[:l-1]
	return val
}

func (s *Stack) Clear() {
	s.values = nil
}
