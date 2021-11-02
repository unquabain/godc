package main

// Stack is a pretty simple stack of Value pointers.
// It is used both for the main program Stack and for
// registers.
type Stack struct {
	values []*Value
}

// Len returns the length of the stack.
func (s *Stack) Len() int {
	return len(s.values)
}

// Push pushes a new *Value onto the stack.
func (s *Stack) Push(n *Value) {
	s.values = append(s.values, n)
}

// Peek returns the last *Value on the stack without altering the stack.
func (s *Stack) Peek() *Value {
	l := len(s.values)
	if l == 0 {
		return nil
	}
	return s.values[l-1]
}

// Pop returns the last *Value of the stack, removing it.
func (s *Stack) Pop() *Value {
	l := len(s.values)
	if l == 0 {
		return nil
	}
	val := s.Peek()
	s.values = s.values[:l-1]
	return val
}

// Clear removes all *Value from the stack.
func (s *Stack) Clear() {
	s.values = nil
}
