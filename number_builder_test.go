package main

import (
	"testing"
)

func TestNumberBuilder(t *testing.T) {
	interp := NewInterpreter()
	test := func(input string) {
		for _, r := range input {
			err := interp.Interpret(r)
			if err != nil {
				t.Fatalf(`error interpreting rune %q: %v`, string(r), err)
			}
		}
		err := interp.Interpret(' ')
		if err != nil {
			t.Fatalf(`error interpreting space: %v`, err)
		}
	}
	expect := func(expected ...string) {
		numExpected := len(expected)
		stackLen := interp.Stack.Len()
		if stackLen < numExpected {
			t.Fatalf(`expected at least %d stack items, found %d`, numExpected, stackLen)
		}
		for i, exp := range expected {
			val := interp.Stack.Pop()
			actual := val.PrecisionString(int64(interp.Precision))
			if actual != exp {
				t.Fatalf(`unexpected value for argument %d: expected %s; received %s`, i, exp, actual)
			}
		}
	}

	test(`5`)
	expect(`5`)

	test(`52`)
	expect(`52`)

	interp.Precision = 2
	test(`7`)
	expect(`7.00`)

	test(`8.3`)
	expect(`8.30`)

	test(`57.21`)
	expect(`57.21`)

	test(`123.456`)
	expect(`123.45`)

	test(`.123`)
	expect(`0.12`)

	test(`12.34.56.78`)
	expect(`0.78`, `0.56`, `12.34`)

	test(`_8`)
	expect(`-8.00`)

	test(`_123.45`)
	expect(`-123.45`)

	test(`12_34`)
	expect(`-34.00`, `12.00`)

	test(`12.34_56.78.90`)
	expect(`0.90`, `-56.78`, `12.34`)
}
