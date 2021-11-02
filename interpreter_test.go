package main

import (
	"fmt"
	"strings"
	"testing"
)

func testWithInterpreter(interpreter *Interpreter, str string) error {
	for i, r := range []rune(str) {
		if err := interpreter.Interpret(r); err != nil {
			if err == ExitRequestedError {
				return nil
			}
			return fmt.Errorf(`couldn't interpret %q, character %d of %q: %w`, r, i, str, err)
		}
	}
	interpreter.Interpret('f')
	return nil
}

func expectWithInterpreter(buff *strings.Builder, values ...string) error {
	str := buff.String()
	buff.Reset()

	lines := strings.Split(strings.TrimSpace(str), "\n")
	nExpected := len(values)
	nActual := len(lines)
	if nExpected != nActual {
		return fmt.Errorf("expected %d lines; found %d:\nEXPECTED:%v\nRECEIVED:\n%v", nExpected, nActual, values, lines)
	}

	for stackNum, expected := range values {
		actual := lines[stackNum]
		if actual != expected {
			return fmt.Errorf(`output %d was expected to be %q; was %q`, stackNum, expected, actual)
		}
	}
	return nil
}

func TestBasicMath(t *testing.T) {
	interpreter := NewInterpreter()
	buff := new(strings.Builder)
	interpreter.output = buff
	test := func(str string) {
		err := testWithInterpreter(interpreter, str)
		if err != nil {
			t.Fatalf(`could not set up test %q: %v`, str, err)
		}
	}

	expect := func(values ...string) {
		err := expectWithInterpreter(buff, values...)
		if err != nil {
			t.Fatalf(`test failed: %v`, err)
		}
		interpreter.Interpret('c')
	}

	t.Run(`entering a number`, func(t *testing.T) {
		test(`12`)
		expect(`12`)
	})

	t.Run(`entering a float with precision 0`, func(t *testing.T) {
		test(`12.34`)
		expect(`12`)
	})

	t.Run(`setting the precision`, func(t *testing.T) {
		test(`2k12.34`)
		expect(`12.34`)
	})

	t.Run(`adding numbers`, func(t *testing.T) {
		test(`12.34 43.21+`)
		expect(`55.55`)
	})

	t.Run(`adding negative numbers`, func(t *testing.T) {
		test(`55.55_43.21+`)
		expect(`12.34`)
	})

	t.Run(`subtracting numbers`, func(t *testing.T) {
		test(`55.55 43.21-`)
		expect(`12.34`)
	})

	t.Run(`subtracting negative numbers`, func(t *testing.T) {
		test(`55.55_43.21-`)
		expect(`98.76`)
	})

	t.Run(`multiplying numbers`, func(t *testing.T) {
		test(`3 4*`)
		expect(`12.00`)
	})

	t.Run(`multiplying negative numbers`, func(t *testing.T) {
		test(`3_4*`)
		expect(`-12.00`)
	})

	t.Run(`multiplying numbers with different precisions`, func(t *testing.T) {
		test(`30 0.4*`)
		expect(`12.00`)
	})

	t.Run(`multiplying numbers with different precisions and mixed signs`, func(t *testing.T) {
		test(`30_0.4*`)
		expect(`-12.00`)
	})

	t.Run(`dividing`, func(t *testing.T) {
		test(`0k30 15/`)
		expect(`2`)
	})

	t.Run(`dividing with precision`, func(t *testing.T) {
		test(`4k30 15/`)
		expect(`2.0000`)
	})

	t.Run(`dividing with different precision`, func(t *testing.T) {
		test(`0k12 4k0.0002/`)
		expect(`60000.0000`)
	})

	t.Run(`dividing by a negative`, func(t *testing.T) {
		test(`0k12_4/`)
		expect(`-3`)
	})

	t.Run(`modulo`, func(t *testing.T) {
		test(`0k365 7%`)
		expect(`1`)
	})

	t.Run(`quotient and remainder`, func(t *testing.T) {
		test(`0k365 7~`)
		expect(`52`, `1`)
	})

	t.Run(`exponents`, func(t *testing.T) {
		test(`0k3 3^`)
		expect(`27`)
	})

	t.Run(`exponents with different precision`, func(t *testing.T) {
		test(`2k3 3^`)
		expect(`27.00`)

		test(`4k2 8^`)
		expect(`256.0000`)

		test(`2k1.41 12^`)
		expect(`61.74`)
	})

	t.Run(`modular exponents`, func(t *testing.T) {
		test(`0k2 8 7|`)
		expect(`4`)
	})

	t.Run(`modular exponents with precision`, func(t *testing.T) {
		test(`4k2 8 7|`)
		expect(`4.0000`)
	})

	t.Run(`square root`, func(t *testing.T) {
		test(`0k256v`)
		expect(`16`)
	})

	t.Run(`square root with precision`, func(t *testing.T) {
		test(`3k256v`)
		expect(`16.000`)

		test(`4k1.41 2^v`)
		expect(`1.4100`)

	})

}

func TestRegisterOperations(t *testing.T) {
	interpreter := NewInterpreter()
	buff := new(strings.Builder)
	interpreter.output = buff
	test := func(str string) {
		err := testWithInterpreter(interpreter, str)
		if err != nil {
			t.Fatalf(`could not set up test %q: %v`, str, err)
		}
	}

	expect := func(values ...string) {
		err := expectWithInterpreter(buff, values...)
		if err != nil {
			t.Fatalf(`test failed: %v`, err)
		}
		interpreter.Interpret('c')
	}

	t.Run(`basic save and retrieve`, func(t *testing.T) {
		// Load two numbers, move one to register `l`, print stack,
		// retrieve from register `l` (implicit print stack)
		test(`12 23slfll`)
		expect(`12`, `23`, `12`)
	})

	t.Run(`stack-based save and retrieve`, func(t *testing.T) {
		// Load two numbers, move one to register `l`, print stack,
		// retrieve from register `l` (implicit print stack)
		test(`12 23Sl45SlfLlLl`)
		expect(`12`, `23`, `45`, `12`)
	})

	t.Run(`mixed register save and retrieve`, func(t *testing.T) {
		test(`12 34Sx45Sy67Sx89SyLxLyLxLy`)
		// Save numbers to two different registers, and retrieve them
		// in a different order.
		expect(`45`, `34`, `89`, `67`, `12`)
	})

	t.Run(`save and retrieve strings`, func(t *testing.T) {
		test(`[test A]sx[test B]sy[B]ly[A]lx`)
		expect(`test A`, `A`, `test B`, `B`)
	})
}

func TestMacroOperations(t *testing.T) {
	interpreter := NewInterpreter()
	buff := new(strings.Builder)
	interpreter.output = buff
	test := func(str string) {
		err := testWithInterpreter(interpreter, str)
		if err != nil {
			t.Fatalf(`could not set up test %q: %v`, str, err)
		}
	}

	expect := func(values ...string) {
		err := expectWithInterpreter(buff, values...)
		if err != nil {
			t.Fatalf(`test failed: %v`, err)
		}
		interpreter.Interpret('c')
	}

	t.Run(`basic math in a macro`, func(t *testing.T) {
		test(`[15 3/]x`)
		expect(`5`)
	})

	t.Run(`test 1-level exit`, func(t *testing.T) {
		test(`[15 3/pq10*p]x`)
		expect(`5`)
	})

	t.Run(`test multi-level exit`, func(t *testing.T) {
		test(`[3Q][x1][x2][x3][x4][x5]x`)
		expect(`5`, `4`, `3`)
	})

}
