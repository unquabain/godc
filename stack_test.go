package main

import (
	"testing"
)

func TestStack(t *testing.T) {
	s := new(Stack)
	testLen := func(expected int) {
		if actual := s.Len(); actual != expected {
			t.Fatalf(`expected Len() to be %d; was %d`, expected, actual)
		}
	}
	testNum := func(expected, actual *Value) {
		if expected != actual {
			t.Fatalf(`expected %p %s; received %p %s`, expected, expected, actual, actual)
		}
	}

	a := new(Value)
	b := new(Value)
	c := new(Value)

	testLen(0)

	s.Push(a)
	testLen(1)
	testNum(a, s.Peek())

	s.Push(b)
	testLen(2)
	testNum(b, s.Peek())

	s.Push(c)
	testLen(3)
	testNum(c, s.Peek())

	testNum(c, s.Pop())
	testLen(2)

	testNum(b, s.Pop())
	testLen(1)

	testNum(a, s.Pop())
	testLen(0)

	testNum(nil, s.Pop())
	testLen(0)
}
