package main

import (
	"fmt"
	"math/big"
)

// ErrNotANumber is an error returned if you try to perform a numerical
// operation on a string.
var ErrNotANumber = fmt.Errorf(`value is not a number`)

// ErrDivideByZer is thrown if you try to divide or take a modulo of zero.
var ErrDivideByZero = fmt.Errorf(`divide by zero`)

// ErrNoImaginaryNumbers is thrown if you try to take the square root of a negative number.
var ErrNoImaginaryNumbers = fmt.Errorf(`no imaginary numbers allowed`)

// ErrWholeExponentsOnly is thrown if you try to raise a number to an exponent
// that is smaller than 1
var ErrWholeExponentsOnly = fmt.Errorf(`only whole numbers are supported as exponents`)

// ValueType indicates whether the value is a string or a number
type ValueType bool

const (
	VTNumber ValueType = false
	VTString ValueType = true
)

var ten = big.NewInt(10)

// Value can be either a number, represented as an integer and a base-10 precision,
// or a string.
type Value struct {
	intval    *big.Int
	strval    []rune
	precision int
	Type      ValueType
}

func precisionToFactor(precision int) *big.Int {
	var pow big.Int
	prec := big.NewInt(int64(precision))
	mod := (*big.Int)(nil)
	return (&pow).Exp(ten, prec, mod)
}

// String implements fmt.Stringer
func (n *Value) String() string {
	if n.Type == VTString {
		return string(n.strval)
	}
	str := n.intval.String()
	if n.precision == 0 {
		return str
	}
	split := len(str) - n.precision
	if split < 0 {
		fmt.Printf(
			"str: %q; len(str): %d; precision: %d; n %d\n",
			str, len(str), n.precision, n.intval,
		)
		split = 0
	}
	intPart, fracPart := str[:split], str[split:]
	if intPart == `` {
		intPart = `0`
	}
	return fmt.Sprintf(`%s.%s`, intPart, fracPart)
}

// Format tries to implement fmt.Formatter
// TODO: Doesn't seem to work right.
func (n *Value) Format(f fmt.State, verb rune) {
	if verb != 's' && verb != 'v' {
		return
	}
	f.Write([]byte(n.String()))
}

// UpdatePrecision sets the precision, and also multiplies
// or divides the intval by 10 so the logical value is the
// same.
func (n *Value) UpdatePrecision(newprecision int) {
	if !n.Type == VTNumber {
		return
	}
	if newprecision > n.precision {
		n.intval.Mul(n.intval, precisionToFactor(newprecision-n.precision))
	} else {
		n.intval.Div(n.intval, precisionToFactor(n.precision-newprecision))
	}
	n.precision = newprecision
}

// MatchPrecision makes both the receiver and the argument
// the same precision, matching whichever has greater precision.
func (n *Value) MatchPrecision(m *Value) {
	if n.precision > m.precision {
		n, m = m, n
	}
	n.UpdatePrecision(m.precision)
}

// Dup makes a copy of the value
func (n *Value) Dup() *Value {
	dup := new(Value)
	dup.Type = n.Type
	if n.intval != nil {
		dup.intval = big.NewInt(0)
		dup.intval.Set(n.intval)
	}
	dup.strval = make([]rune, len(n.strval))
	copy(dup.strval, n.strval)
	dup.precision = n.precision
	return dup
}

// Add adds the value of m to n or returns an error if
// either is not a number.
func (n *Value) Add(m *Value) error {
	if n.Type != VTNumber {
		return ErrNotANumber
	}
	if m.Type != VTNumber {
		return ErrNotANumber
	}
	n.MatchPrecision(m)
	n.intval.Add(n.intval, m.intval)
	return nil
}

// Subtract subtracts the value of m from n, or returns
// an error if either is not a number.
func (n *Value) Subtract(m *Value) error {
	if n.Type != VTNumber {
		return ErrNotANumber
	}
	if m.Type != VTNumber {
		return ErrNotANumber
	}
	n.MatchPrecision(m)
	n.intval.Sub(n.intval, m.intval)
	return nil
}

// Multiply multiplies n by the value of m, or
// returns an error if either is not a number.
// The precision of n becomes the sum of the
// precision of both values.
func (n *Value) Multiply(m *Value) error {
	if n.Type != VTNumber {
		return ErrNotANumber
	}
	if m.Type != VTNumber {
		return ErrNotANumber
	}
	n.precision += m.precision
	n.intval.Mul(n.intval, m.intval)
	return nil
}

// Divide divides n by m (or m into n), or
// returns an error if either is not a
// number or if m == 0. The precision should become
// the greater of either n or m.
func (n *Value) Divide(m *Value) error {
	if n.Type != VTNumber {
		return ErrNotANumber
	}
	if m.Type != VTNumber {
		return ErrNotANumber
	}
	if m.intval.Sign() == 0 {
		return ErrDivideByZero
	}
	// Make sure we've got enough zeroes to play with
	n.MatchPrecision(m)                          // Upgrade to the greatest precision
	n.UpdatePrecision(n.precision + m.precision) // We're going to lose m.precision in the divide

	// Do the math
	n.intval.Div(n.intval, m.intval)
	n.precision -= m.precision
	return nil
}

// IntVal reduces the precision to 0, discarding any
// fractional portion.
func (n *Value) IntVal() error {
	if n.Type != VTNumber {
		return ErrNotANumber
	}
	n.UpdatePrecision(0)
	return nil
}

// Int returns the value as an int, discarding
// any fractional portion.
func (n *Value) Int() int {
	n.IntVal()
	return int(n.intval.Int64())
}

// FracVal discards any integer portion, keeping
// only n.precision fractional digits.
func (n *Value) FracVal() error {
	if n.Type != VTNumber {
		return ErrNotANumber
	}
	d := n.Dup()
	if err := d.IntVal(); err != nil {
		return err
	}
	return n.Subtract(d)
}

// QutotientRemainder divides n by m (or m into n) and returns
// an integer quotient and the modulo.
// Returns an error if either value is not a number, or if m == 0
func (n *Value) QuotientRemainder(m *Value) (*Value, *Value, error) {
	if n.Type != VTNumber {
		return nil, nil, ErrNotANumber
	}
	if m.Type != VTNumber {
		return nil, nil, ErrNotANumber
	}
	if m.intval.Sign() == 0 {
		return nil, nil, ErrDivideByZero
	}
	n.MatchPrecision(m)
	n.intval.QuoRem(n.intval, m.intval, m.intval)
	return n, m, nil
}

// Exponent raises n to the integer value of m.
// Fractional or negative exponents are not
// supported.
func (n *Value) Exponent(m *Value) error {
	if n.Type != VTNumber {
		return ErrNotANumber
	}
	if m.Type != VTNumber {
		return ErrNotANumber
	}
	m.UpdatePrecision(0)
	if m.intval.Sign() <= 0 {
		return ErrWholeExponentsOnly
	}
	n.intval.Exp(n.intval, m.intval, nil)
	afterPower := n.precision * (m.Int() - 1)
	afterFactor := precisionToFactor(afterPower)
	n.intval.Div(n.intval, afterFactor)
	return nil
}

// ModExponent raises n to the power of e, module m.
func (n *Value) ModExponent(e, m *Value) error {
	if n.Type != VTNumber {
		return ErrNotANumber
	}
	if m.Type != VTNumber {
		return ErrNotANumber
	}
	if e.Type != VTNumber {
		return ErrNotANumber
	}
	n.UpdatePrecision(0)
	e.UpdatePrecision(0)
	m.UpdatePrecision(0)
	if e.intval.Sign() <= 0 {
		return ErrWholeExponentsOnly
	}
	n.intval.Exp(n.intval, e.intval, m.intval)
	return nil
}

// Sqrt returns the square root of the number.
func (n *Value) Sqrt() error {
	if n.Type != VTNumber {
		return ErrNotANumber
	}
	if n.intval.Sign() < 0 {
		return ErrNoImaginaryNumbers
	}
	n.UpdatePrecision(n.precision * 2)
	n.intval.Sqrt(n.intval)
	n.precision /= 2
	return nil
}
