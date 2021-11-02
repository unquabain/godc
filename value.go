package main

import (
	"fmt"
	"math/big"
)

var NotANumberError = fmt.Errorf(`value is not a number`)
var DivideByZeroError = fmt.Errorf(`divide by zero`)
var NoImaginaryNumbersError = fmt.Errorf(`no imaginary numbers allowed`)

type ValueType bool

const (
	VTNumber ValueType = false
	VTString ValueType = true
)

var ten = big.NewInt(10)

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

func (n *Value) Format(f fmt.State, verb rune) {
	if verb != 's' && verb != 'v' {
		return
	}
	f.Write([]byte(n.String()))
}

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

func (n *Value) MatchPrecision(m *Value) {
	if n.precision > m.precision {
		n, m = m, n
	}
	n.UpdatePrecision(m.precision)
}

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

func (n *Value) Add(m *Value) error {
	if n.Type != VTNumber {
		return NotANumberError
	}
	n.MatchPrecision(m)
	n.intval.Add(n.intval, m.intval)
	return nil
}

func (n *Value) Subtract(m *Value) error {
	if n.Type != VTNumber {
		return NotANumberError
	}
	n.MatchPrecision(m)
	n.intval.Sub(n.intval, m.intval)
	return nil
}

func (n *Value) Multiply(m *Value) error {
	if n.Type != VTNumber {
		return NotANumberError
	}
	n.precision += m.precision
	n.intval.Mul(n.intval, m.intval)
	return nil
}

func (n *Value) Divide(m *Value) error {
	if n.Type != VTNumber {
		return NotANumberError
	}
	if m.intval.Sign() == 0 {
		return DivideByZeroError
	}
	// Make sure we've got enough zeroes to play with
	n.MatchPrecision(m)                          // Upgrade to the greatest precision
	n.UpdatePrecision(n.precision + m.precision) // We're going to lose m.precision in the divide

	// Do the math
	n.intval.Div(n.intval, m.intval)
	n.precision -= m.precision
	return nil
}

func (n *Value) NormalizePrecision() {
	test := big.NewInt(0)
	for test.Mod(n.intval, ten).Sign() == 0 && n.precision >= 0 {
		n.precision -= 1
		n.intval.Div(n.intval, ten)
	}
	for n.precision < 0 {
		n.precision += 1
		n.intval.Mul(n.intval, ten)
	}
}

func (n *Value) IntVal() error {
	if n.Type != VTNumber {
		return NotANumberError
	}
	n.UpdatePrecision(0)
	return nil
}

func (n *Value) Int() int {
	n.IntVal()
	return int(n.intval.Int64())
}

func (n *Value) FracVal() error {
	if n.Type != VTNumber {
		return NotANumberError
	}
	d := n.Dup()
	if err := d.IntVal(); err != nil {
		return err
	}
	return n.Subtract(d)
}

func (n *Value) QuotientRemainder(m *Value) (*Value, *Value, error) {
	if n.Type != VTNumber {
		return nil, nil, NotANumberError
	}
	if m.intval.Sign() == 0 {
		return nil, nil, DivideByZeroError
	}
	n.MatchPrecision(m)
	n.intval.QuoRem(n.intval, m.intval, m.intval)
	return n, m, nil
}

func (n *Value) Exponent(m *Value) error {
	if n.Type != VTNumber {
		return NotANumberError
	}
	m.UpdatePrecision(0)
	n.intval.Exp(n.intval, m.intval, nil)
	afterPower := n.precision * (m.Int() - 1)
	afterFactor := precisionToFactor(afterPower)
	n.intval.Div(n.intval, afterFactor)
	return nil
}

func (n *Value) ModExponent(e, m *Value) error {
	if n.Type != VTNumber {
		return NotANumberError
	}
	n.UpdatePrecision(0)
	e.UpdatePrecision(0)
	m.UpdatePrecision(0)
	n.intval.Exp(n.intval, e.intval, m.intval)
	return nil
}

func (n *Value) Sqrt() error {
	if n.Type != VTNumber {
		return NotANumberError
	}
	if n.intval.Sign() < 0 {
		return NoImaginaryNumbersError
	}
	n.UpdatePrecision(n.precision * 2)
	n.intval.Sqrt(n.intval)
	n.precision /= 2
	return nil
}
