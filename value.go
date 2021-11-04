package main

import (
	"fmt"
	"math/big"
	"strings"
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
	numval *big.Rat
	strval []rune
	Type   ValueType
}

func (n *Value) Text(radix, precision int64) string {
	// If the value is a string, print the string
	if n.Type == VTString {
		return string(n.strval)
	}

	val := n.numval
	strSign := ``
	if val.Sign() < 0 {
		strSign = `-`
	}
	val = val.Abs(val)
	intPart := (&big.Int{}).Div(val.Num(), val.Denom())
	fracPart := (&big.Rat{}).Sub(val, (&big.Rat{}).SetInt(intPart))
	strVal := intPart.Text(int(radix))

	if precision == 0 {
		return fmt.Sprintf(`%s%s`, strSign, strVal)
	}

	// get the fractional part
	r := (&big.Rat{}).SetInt64(radix)
	b := &strings.Builder{}
	for p := precision; p > 0; p-- {
		if fracPart.Sign() == 0 {
			b.WriteRune('0')
			continue
		}
		fracPart.Mul(fracPart, r)
		intPart.Div(fracPart.Num(), fracPart.Denom())
		fracPart.Sub(fracPart, (&big.Rat{}).SetInt(intPart))
		digit := intPart.Text(int(radix))
		b.WriteString(digit)
	}
	strFrac := b.String()
	return fmt.Sprintf(`%s%s.%s`, strSign, strVal, strFrac)
}

func (n *Value) PrecisionString(precision int64) string {
	return n.Text(10, precision)
}

// Format tries to implement fmt.Formatter
// TODO: Doesn't seem to work right.
func (n *Value) Format(f fmt.State, verb rune) {
	switch n.Type {
	case VTNumber:
		if verb != 'f' && verb != 'v' {
			f.Write([]byte(`unknown verb for number type Value`))
			return
		}
		prec, ok := f.Precision()
		if !ok {
			prec = 0
		}
		f.Write([]byte(n.Text(10, int64(prec))))
	case VTString:
		if verb != 's' && verb != 'v' {
			f.Write([]byte(`unknown verb for string type Value`))
			return
		}
		f.Write([]byte(string(n.strval)))
	default:
		f.Write([]byte(`unknown type for Value`))
	}
}

// CrossMultiply returns two numerators and their common denominator
func (n *Value) CrossMultiply(m *Value) (*big.Int, *big.Int, *big.Int, error) {
	if n.Type != VTNumber {
		return nil, nil, nil, ErrNotANumber
	}
	if m.Type != VTNumber {
		return nil, nil, nil, ErrNotANumber
	}
	denom := (&big.Int{}).Mul(n.numval.Denom(), m.numval.Denom())
	nnum := (&big.Int{}).Mul(n.numval.Num(), m.numval.Denom())
	mnum := (&big.Int{}).Mul(m.numval.Num(), n.numval.Denom())
	return nnum, mnum, denom, nil
}

// Dup makes a copy of the value
func (n *Value) Dup() *Value {
	dup := new(Value)
	dup.Type = n.Type
	if n.numval != nil {
		dup.numval = &big.Rat{}
		dup.numval.Set(n.numval)
	}
	if n.strval != nil {
		dup.strval = make([]rune, len(n.strval))
		copy(dup.strval, n.strval)
	}
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
	n.numval.Add(n.numval, m.numval)
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
	n.numval.Sub(n.numval, m.numval)
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
	n.numval.Mul(n.numval, m.numval)
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
	if m.numval.Sign() == 0 {
		return ErrDivideByZero
	}
	// Do the math
	n.numval.Mul(n.numval, (&big.Rat{}).Inv(m.numval))
	return nil
}

// IntVal reduces the precision to 0, discarding any
// fractional portion.
func (n *Value) IntVal() error {
	if n.Type != VTNumber {
		return ErrNotANumber
	}
	ival := (&big.Int{}).Div(n.numval.Num(), n.numval.Denom())
	n.numval.SetInt(ival)
	return nil
}

// Int returns the value as an int, discarding
// any fractional portion.
func (n *Value) Int() int64 {
	n.IntVal()
	return n.numval.Num().Int64()
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

func (n *Value) IsInt() bool {
	if n.Type != VTNumber {
		return false
	}
	return n.numval.IsInt()
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
	if m.numval.Sign() == 0 {
		return nil, nil, ErrDivideByZero
	}
	q := &big.Int{}
	r := &big.Int{}
	if n.IsInt() && m.IsInt() {
		ni := n.numval.Num()
		mi := m.numval.Num()
		q.QuoRem(ni, mi, r)
		return &Value{numval: (&big.Rat{}).SetInt(q)},
			&Value{numval: (&big.Rat{}).SetInt(r)},
			nil
	}
	// Use integer math to drop fractional parts.
	// x / y = q + r/y => x % y = (q, r)
	// (x/c) / (y/c) = (q/c) + (rc/y) => (x/c) % (y/c) = ((q/c), rc)
	x, y, c, err := n.CrossMultiply(m)
	if err != nil {
		return nil, nil, err
	}
	q.QuoRem(x, y, r)
	quotient := &Value{numval: (&big.Rat{}).SetFrac(q, c)}
	remainder := &Value{numval: (&big.Rat{}).SetInt(r.Mul(r, c))}
	return quotient, remainder, nil
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
	if m.numval.Sign() <= 0 {
		return ErrWholeExponentsOnly
	}
	if err := m.IntVal(); err != nil {
		return err
	}
	num := n.numval.Num()
	denom := n.numval.Denom()
	num.Exp(num, m.numval.Num(), nil)
	denom.Exp(denom, m.numval.Num(), nil)
	n.numval.SetFrac(num, denom)
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
	if e.numval.Sign() <= 0 {
		return ErrWholeExponentsOnly
	}
	if err := n.IntVal(); err != nil {
		return err
	}
	if err := e.IntVal(); err != nil {
		return err
	}
	if err := m.IntVal(); err != nil {
		return err
	}
	val := (&big.Int{}).Exp(n.numval.Num(), e.numval.Num(), m.numval.Num())
	n.numval.SetInt(val)
	return nil
}

// Sqrt returns the square root of the number.
func (n *Value) Sqrt() error {
	if n.Type != VTNumber {
		return ErrNotANumber
	}
	if n.numval.Sign() < 0 {
		return ErrNoImaginaryNumbers
	}
	num := n.numval.Num()
	denom := n.numval.Denom()
	num.Sqrt(num)
	denom.Sqrt(denom)
	n.numval.SetFrac(num, denom)
	return nil
}
