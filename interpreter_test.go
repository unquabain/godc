package main

import (
	"fmt"
	"strings"
	"testing"
)

func testWithInterpreter(interpreter *Interpreter, str string) error {
	interpreter.Interpret('c')
	for i, r := range []rune(str) {
		if err := interpreter.Interpret(r); err != nil {
			if err == ErrExitRequested {
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

	t.Run(`run macro if gt`, func(t *testing.T) {
		test(`[50]sa0 1>a`)
		expect(`50`)
	})

	t.Run(`skip macro if not gt`, func(t *testing.T) {
		test(`[nope][50]sa1 0>a`)
		expect(`nope`)
	})

	t.Run(`run macro if lt`, func(t *testing.T) {
		test(`[50]sa1 0<a`)
		expect(`50`)
	})

	t.Run(`skip macro if not lt`, func(t *testing.T) {
		test(`[nope][50]sa0 1<a`)
		expect(`nope`)
	})

	t.Run(`run macro if eq`, func(t *testing.T) {
		test(`[50]sa1 1=a`)
		expect(`50`)
	})

	t.Run(`skip macro if not eq`, func(t *testing.T) {
		test(`[nope][50]sa0 1=a`)
		expect(`nope`)
	})
}

func TestNegativeMacroOperations(t *testing.T) {
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

	t.Run(`run macro if not gt`, func(t *testing.T) {
		test(`[25 2*]sa1 0!>a`)
		expect(`50`)
	})

	t.Run(`skip macro if not not gt`, func(t *testing.T) {
		test(`[nope][25 2*1+]sa0 1!>a`)
		expect(`nope`)
	})

	t.Run(`run macro if not lt`, func(t *testing.T) {
		test(`[25 2*2+]sa0 1!<a`)
		expect(`52`)
	})

	t.Run(`skip macro if not not lt`, func(t *testing.T) {
		test(`[nope][25 2*3+]sa1 0!<a`)
		expect(`nope`)
	})

	t.Run(`run macro if not eq`, func(t *testing.T) {
		test(`[25 2*4+]sa1 0!=a`)
		expect(`54`)
	})

	t.Run(`skip macro if not not eq`, func(t *testing.T) {
		test(`[nope][25 2*5+]sa1 1!=a`)
		expect(`nope`)
	})
}

func TestRadixOperations(t *testing.T) {
	interpreter := NewInterpreter()
	buff := new(strings.Builder)
	interpreter.output = buff
	test := func(str string) {
		interpreter.InputRadix = 10
		interpreter.OutputRadix = 10
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

	t.Run(`decimal to hex conversion of integers`, func(t *testing.T) {
		test(`10i16o0k5`)
		expect(`5`)

		test(`10i16o0k10`)
		expect(`A`)

		test(`10i16o0k32`)
		expect(`20`)

		test(`10i16o0k35`)
		expect(`23`)
	})

	t.Run(`decimal to hex conversion with precision`, func(t *testing.T) {
		test(`10i16o4k5`)
		expect(`5.0000`)

		test(`10i16o4k10`)
		expect(`A.0000`)

		test(`10i16o4k32.5`)
		expect(`20.8000`)

		test(`10i16o4k35.25`)
		expect(`23.4000`)
	})

	t.Run(`hex to decimal conversion of integers`, func(t *testing.T) {
		test(`16iAo0k5`)
		expect(`5`)

		test(`16iAo0kA`)
		expect(`10`)

		test(`16iAo0k20`)
		expect(`32`)

		test(`16iAo0k23`)
		expect(`35`)
	})

	t.Run(`hex to decimal conversion with precision`, func(t *testing.T) {
		test(`16iAo4k5`)
		expect(`5.0000`)

		test(`16iAo4kA`)
		expect(`10.0000`)

		test(`16iAo4k20.8`)
		expect(`32.5000`)

		test(`16iAo4k23.4`)
		expect(`35.2500`)
	})

	t.Run(`converting different bases`, func(t *testing.T) {
		test(`0k16i9oA`)
		expect(`11`)

		test(`0k16o9i11`)
		expect(`A`)
	})

	t.Run(`output commands`, func(t *testing.T) {
		test(`8oO`)
		expect(`10`)

		test(`8oO 10o`)
		expect(`8`)

		test(`14iI`)
		expect(`14`)
	})
}

func TestPrintOperations(t *testing.T) {
	interpreter := NewInterpreter()
	buff := new(strings.Builder)
	interpreter.output = buff
	test := func(str string) {
		interpreter.InputRadix = 10
		interpreter.OutputRadix = 10
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

	t.Run(`test raw printing a string`, func(t *testing.T) {
		test(`[a string]P`)
		expect(`a string`)
	})

	t.Run(`test raw printing a number`, func(t *testing.T) {
		test(`310400273487P`)
		expect(`HELLO`)
	})

	t.Run(`test raw printing a number with extra info`, func(t *testing.T) {
		test(`4k_310400273487.1234P`)
		expect(`HELLO`)
	})

	t.Run(`test a string with nested brackets`, func(t *testing.T) {
		test(`[a string with [nested] brackets]`)
		expect(`a string with [nested] brackets`)
	})
}
