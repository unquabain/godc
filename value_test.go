package main

import (
	"math/big"
	"testing"
)

func newValue(num, denom int64) *Value {
	n := new(Value)
	n.numval = big.NewRat(num, denom)
	return n
}

func TestValueString(t *testing.T) {
	n := newValue(10000, 100)
	expected := `100.00`
	if actual := n.PrecisionString(2); actual != expected {
		t.Fatalf(`expected %q, received %q`, expected, actual)
	}
	expected = `100.000`
	if actual := n.PrecisionString(3); actual != expected {
		t.Fatalf(`expected %q, received %q`, expected, actual)
	}
	if err := n.Add(newValue(10, 100)); err != nil {
		t.Fatalf(`could not add: %v`, err)
	}
	expected = `100.10`
	if actual := n.PrecisionString(2); actual != expected {
		t.Fatalf(`expected %q, received %q`, expected, actual)
	}
	expected = `100.100`
	if actual := n.PrecisionString(3); actual != expected {
		t.Fatalf(`expected %q, received %q`, expected, actual)
	}
}

func TestValueText(t *testing.T) {
	test := func(num, denom, precision int64, radix uint8, expected string) {
		val := newValue(num, denom)
		actual := val.Text(int64(radix), precision)
		if actual != expected {
			t.Fatalf(`expected %d / %d radix %d, precision %d to be %q; was %q`, num, denom, radix, precision, expected, actual)
		}
	}

	t.Run(`some easy integers`, func(t *testing.T) {
		test(256, 1, 0, 16, `100`)
		test(64, 1, 0, 8, `100`)
		test(27, 1, 0, 3, `1000`)

		test(55, 1, 0, 16, `37`)
	})

	t.Run(`some easy integers with precision`, func(t *testing.T) {
		test(2560, 10, 1, 16, `100.0`)
		test(6400, 100, 2, 8, `100.00`)
		test(27000, 1000, 3, 3, `1000.000`)

		test(550000, 10000, 4, 16, `37.0000`)
	})

	t.Run(`some rational fractions`, func(t *testing.T) {
		test(15, 10, 1, 16, `1.8`)            // 0x1.8 = 1.5
		test(25, 100, 2, 16, `0.40`)          // 0x0.40 = 0.25
		test(640625, 10000, 4, 16, `40.1000`) // 0x40.1 = 64.0625
		test(3, 10, 1, 16, `0.4`)             // 0x0.48 = 0.3 Values are truncated.
	})
}

func TestAdd(t *testing.T) {
	n := newValue(1001, 100)
	m := newValue(2002, 10)

	expected := big.NewRat(21021, 100)

	n.Add(m)
	if actual := n.numval; actual.Cmp(expected) != 0 {
		t.Fatalf(`expected sum to be %v; got %v`, expected, actual)
	}
}

func TestSubtract(t *testing.T) {
	n := newValue(21021, 100)
	m := newValue(2002, 10)

	expected := big.NewRat(1001, 100)

	n.Subtract(m)
	if actual := n.numval; actual.Cmp(expected) != 0 {
		t.Fatalf(`expected sum to be %v; got %v`, expected, actual)
	}
}

func TestMultply(t *testing.T) {
	n := newValue(300, 100)
	m := newValue(500, 100)

	n.Multiply(m)
	expected := big.NewRat(150000, 10000)
	if actual := n.numval; actual.Cmp(expected) != 0 {
		t.Fatalf(`expected value was %v; found %v`, expected, actual)
	}

	n.numval = big.NewRat(500, 100)

	m.numval = big.NewRat(300, 100)

	n.Multiply(m)
	expected = big.NewRat(150000, 10000)
	if actual := n.numval; actual.Cmp(expected) != 0 {
		t.Fatalf(`expected value was %v; found %v`, expected, actual)
	}
}

func pow(prec int64) int64 {
	if prec == 0 {
		return 1
	}
	return (&big.Int{}).Exp(
		big.NewInt(10),
		big.NewInt(prec),
		nil,
	).Int64()
}

func TestDivide(t *testing.T) {
	test := func(nval, nprec, mval, mprec, eval, eprec int64) {
		n := newValue(nval, pow(nprec))
		m := newValue(mval, pow(mprec))
		expected := big.NewRat(eval, pow(eprec))

		n.Divide(m)

		if actual := n.numval; actual.Cmp(expected) != 0 {
			t.Fatalf(`expected value %v; found value %v`, eval, actual)
		}
	}
	test(50, 1, 20, 1, 25, 1)       //   5.0  / 2.0   =  2.5
	test(15, 0, 5, 1, 300, 1)       //  15    / 0.5   = 30.0
	test(500, 0, 20, 1, 2500, 1)    // 500    / 2.0   = 250.0
	test(30, 0, 5, 3, 6_000_000, 3) //  30    / 0.005 = 6,000.000
	test(100, 2, 4, 0, 25, 2)       //   1.00 / 4     = 0.25
	test(500000, 4, 20000, 4, 250000, 4)

	n := newValue(10, pow(0))
	m := newValue(0, pow(0))
	err := n.Divide(m)
	if err == nil {
		t.Fatalf(`expected divide by zero error: received none`)
	}
	if err != ErrDivideByZero {
		t.Fatalf(`expected divide by zero error: received %v`, err)
	}
}

func TestExponent(t *testing.T) {
	n := newValue(3, pow(0))

	m := newValue(3, pow(0))

	n.Exponent(m)

	expected := big.NewRat(27, 1)
	if actual := n.numval; actual.Cmp(expected) != 0 {
		t.Fatalf(`expected 3*3 to equal %v; was %v`, expected, actual)
	}
}

func TestValueDup(t *testing.T) {
	val := &Value{
		Type:   VTNumber,
		numval: big.NewRat(50, 1),
		strval: []rune(`test test test`),
	}

	dup := val.Dup()
	t.Run(`all values filled out`, func(t *testing.T) {
		if actual := dup.Type; actual != val.Type {
			t.Fatalf(`expected Type %v; found %v`, actual, val.Type)
		}
		if actual := dup.numval; actual.Cmp(val.numval) != 0 {
			t.Fatalf(`expected numval %v; found %v`, actual, val.numval)
		}
		if actual := dup.strval; string(actual) != string(val.strval) {
			t.Fatalf(`expected strval %v; found %v`, actual, val.strval)
		}
	})
	val.strval = nil
	dup = val.Dup()
	t.Run(`no strval`, func(t *testing.T) {
		if actual := dup.Type; actual != val.Type {
			t.Fatalf(`expected Type %v; found %v`, actual, val.Type)
		}
		if actual := dup.numval; actual.Cmp(val.numval) != 0 {
			t.Fatalf(`expected numval %v; found %v`, actual, val.numval)
		}
		if actual := dup.strval; actual != nil {
			t.Fatalf(`expected strval %v; found %v`, actual, val.strval)
		}
	})
	val.numval = nil
	val.strval = []rune(`test test test`)
	dup = val.Dup()
	t.Run(`no numval`, func(t *testing.T) {
		if actual := dup.Type; actual != val.Type {
			t.Fatalf(`expected Type %v; found %v`, actual, val.Type)
		}
		if actual := dup.numval; actual != nil {
			t.Fatalf(`expected numval %v; found %v`, actual, val.numval)
		}
		if actual := dup.strval; string(actual) != string(val.strval) {
			t.Fatalf(`expected strval %v; found %v`, actual, val.strval)
		}
	})
}
