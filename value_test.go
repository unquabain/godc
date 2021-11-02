package main

import (
	"math/big"
	"regexp"
	"testing"
)

func newValue(val int64, precision int) *Value {
	n := new(Value)
	n.intval = big.NewInt(val)
	n.precision = precision
	return n
}

func TestValueString(t *testing.T) {
	n := newValue(10000, 2)
	expected := `100.00`
	if actual := n.String(); actual != expected {
		t.Fatalf(`expected %q, received %q`, expected, actual)
	}
	expected = `10.000`
	n.precision = 3
	if actual := n.String(); actual != expected {
		t.Fatalf(`expected %q, received %q`, expected, actual)
	}
	n.intval.Add(n.intval, big.NewInt(10))
	expected = `100.10`
	n.precision = 2
	if actual := n.String(); actual != expected {
		t.Fatalf(`expected %q, received %q`, expected, actual)
	}
	expected = `10.010`
	n.precision = 3
	if actual := n.String(); actual != expected {
		t.Fatalf(`expected %q, received %q`, expected, actual)
	}
}

func TestValueText(t *testing.T) {
	test := func(input, precision int64, radix uint8, expected string) {
		val := &Value{
			intval:    big.NewInt(input),
			precision: int(precision),
		}
		actual := val.Text(radix)
		if actual != expected {
			t.Fatalf(`expected %d * 10^-%d in radix %d to be %q; was %q`, input, precision, radix, expected, actual)
		}
	}

	t.Run(`some easy integers`, func(t *testing.T) {
		test(256, 0, 16, `100`)
		test(64, 0, 8, `100`)
		test(27, 0, 3, `1000`)

		test(55, 0, 16, `37`)
	})

	t.Run(`some easy integers with precision`, func(t *testing.T) {
		test(2560, 1, 16, `100.0`)
		test(6400, 2, 8, `100.00`)
		test(27000, 3, 3, `1000.000`)

		test(550000, 4, 16, `37.0000`)
	})

	t.Run(`some rational fractions`, func(t *testing.T) {
		test(15, 1, 16, `1.8`)         // 0x1.8 = 1.5
		test(25, 2, 16, `0.40`)        // 0x0.40 = 0.25
		test(640625, 4, 16, `40.1000`) // 0x40.1 = 64.0625
		test(3, 1, 16, `0.4`)          // 0x0.48 = 0.3 Values are truncated.
	})
}

func TestUpdatePrecision(t *testing.T) {
	n := newValue(100, 2)
	expectedPattern := regexp.MustCompile(`1(\.0+)?`)
	testPrecision := func(precision int, expectedInt int64) {
		expected := big.NewInt(expectedInt)
		n.UpdatePrecision(precision)
		if actual := n.precision; actual != precision {
			t.Fatalf(`failed to set precision: expected %d; actual %d`, precision, actual)
		}
		if actual := n.intval; actual.Cmp(expected) != 0 {
			t.Fatalf(`expected %d, received %d`, expected, actual)
		}
		if actual := n.String(); !expectedPattern.MatchString(actual) {
			t.Fatalf(`expected String() to eq "1"; was %q`, actual)
		}
	}
	testPrecision(4, 10000)
	testPrecision(6, 1000000)
	testPrecision(2, 100)
	testPrecision(0, 1)
}

func TestMatchPrecision(t *testing.T) {
	n := newValue(0, 0)
	m := newValue(0, 0)

	test := func(expected int) {
		if n.precision != m.precision {
			t.Fatalf(`expected n.precision %d to match m.precision %d`, n.precision, m.precision)
		}
		if n.precision != expected {
			t.Fatalf(`expected precision to equal %d; found %d`, expected, n.precision)
		}
	}

	n.precision = 3
	m.precision = 6
	n.MatchPrecision(m)
	test(6)

	n.precision = 6
	m.precision = 3
	n.MatchPrecision(m)
	test(6)

	n.precision = 3
	m.precision = 6
	m.MatchPrecision(n)
	test(6)

	n.precision = 6
	m.precision = 3
	m.MatchPrecision(n)
	test(6)
}

func TestAdd(t *testing.T) {
	n := newValue(1001, 2)
	m := newValue(2002, 1)

	expected := big.NewInt(21021)

	n.Add(m)
	if actual := n.intval; actual.Cmp(expected) != 0 {
		t.Fatalf(`expected sum to be %d; got %d`, expected, actual)
	}
	if actual := n.precision; actual != 2 {
		t.Fatalf(`expected output precision to be 2, was %d`, actual)
	}
}

func TestSubtract(t *testing.T) {
	n := newValue(21021, 2)
	m := newValue(2002, 1)

	expected := big.NewInt(1001)

	n.Subtract(m)
	if actual := n.intval; actual.Cmp(expected) != 0 {
		t.Fatalf(`expected sum to be %d; got %d`, expected, actual)
	}
	if actual := n.precision; actual != 2 {
		t.Fatalf(`expected output precision to be 2, was %d`, actual)
	}
}

func TestMultply(t *testing.T) {
	n := newValue(300, 2)
	m := newValue(500, 2)

	n.Multiply(m)
	expected := big.NewInt(150000)
	if actual := n.intval; actual.Cmp(expected) != 0 {
		t.Fatalf(`expected value was %d; found %d`, expected, actual)
	}
	expectedPrec := 4
	if actual := n.precision; actual != expectedPrec {
		t.Fatalf(`expected precision was %d; found %d`, expected, actual)
	}

	n.intval = big.NewInt(500)
	n.precision = 2

	m.intval = big.NewInt(300)
	m.precision = 2

	n.Multiply(m)
	expected = big.NewInt(150000)
	if actual := n.intval; actual.Cmp(expected) != 0 {
		t.Fatalf(`expected value was %d; found %d`, expected, actual)
	}
	expectedPrec = 4
	if actual := n.precision; actual != expectedPrec {
		t.Fatalf(`expected precision was %d; found %d`, expected, actual)
	}
}

func TestDivide(t *testing.T) {
	test := func(nval int64, nprec int, mval int64, mprec int, eval int64, eprec int) {
		n := newValue(nval, nprec)
		m := newValue(mval, mprec)

		n.Divide(m)

		if actual := n.precision; actual != eprec {
			t.Fatalf(`expected precision %d; found precision %d`, eprec, actual)
		}
		if actual := n.intval; actual.Int64() != eval {
			t.Fatalf(`expected value %d; found value %d`, eval, actual)
		}
	}
	test(50, 1, 20, 1, 25, 1)       //   5.0  / 2.0   =  2.5
	test(15, 0, 5, 1, 300, 1)       //  15    / 0.5   = 30.0
	test(500, 0, 20, 1, 2500, 1)    // 500    / 2.0   = 250.0
	test(30, 0, 5, 3, 6_000_000, 3) //  30    / 0.005 = 6,000.000
	test(100, 2, 4, 0, 25, 2)       //   1.00 / 4     = 0.25
	test(500000, 4, 20000, 4, 250000, 4)

	n := newValue(10, 0)
	m := newValue(0, 0)
	err := n.Divide(m)
	if err == nil {
		t.Fatalf(`expected divide by zero error: received none`)
	}
	if err != ErrDivideByZero {
		t.Fatalf(`expected divide by zero error: received %v`, err)
	}
}

func TestExponent(t *testing.T) {
	n := newValue(3, 0)

	m := newValue(3, 0)

	n.Exponent(m)

	expected := big.NewInt(27)
	if actual := n.intval; actual.Cmp(expected) != 0 {
		t.Fatalf(`expected 3*3 to equal %d; was %d`, expected, actual)
	}
}
