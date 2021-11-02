package main

import (
	"fmt"
	"math/big"
	"strings"
)

// NumberBuilder handles creating a Value from a stream of
// digits.
type NumberBuilder struct {
	buff    *strings.Builder
	sign    bool
	dotSeen bool
	State   OperationState
}

func isDigit(r rune) bool {
	if r == '.' {
		return true
	}
	if r == '_' {
		return true
	}
	if r >= '0' && r <= '9' {
		return true
	}
	if r >= 'A' && r <= 'H' {
		return true
	}
	return false
}

// Operate implements the Operator interface
func (n *NumberBuilder) Operate(i *Interpreter, r rune) (bool, error) {
	if !isDigit(r) {
		err := n.Flush(i)
		if err != nil {
			return true, err
		}
		return true, ErrContinueProcessingRune
	}
	if n.State == OSHungry && r == '_' {
		err := n.Flush(i)
		if err != nil {
			return true, err
		}
		return true, ErrContinueProcessingRune
	}
	n.State = OSHungry
	if r == '_' {
		n.sign = true
		return false, nil
	}
	if r == '.' {
		if n.dotSeen {
			err := n.Flush(i)
			if err != nil {
				return true, err
			}
			return true, ErrContinueProcessingRune
		}
		n.dotSeen = true
	}
	n.buff.WriteRune(r)
	return false, nil
}

// NewNumberBuilder initializes internal structures in a NumberBuilder
func NewNumberBuilder() *NumberBuilder {
	return &NumberBuilder{
		buff: new(strings.Builder),
	}
}

// Flush finalizes the number and pushes it onto the stack.
func (n *NumberBuilder) Flush(i *Interpreter) error {
	var v Value
	num := big.NewInt(0)
	str := n.buff.String()
	prec := 0

	if n.dotSeen {
		frac := big.NewInt(0)
		parts := strings.Split(str, `.`)
		if len(parts) != 2 {
			return fmt.Errorf(`unexpected number of parts in %q: expected 2, received %d`, str, len(parts))
		}
		if len(parts[0]) > 0 {
			_, ok := num.SetString(parts[0], int(i.InputRadix))
			if !ok {
				return fmt.Errorf(`unable to unmarshal integer part of %q`, str)
			}
		}
		prec = len(parts[1])
		if prec > 0 {
			radix := big.NewInt(int64(i.InputRadix))
			_, ok := frac.SetString(parts[1], int(i.InputRadix))
			if !ok {
				return fmt.Errorf(`unable to unmarshal fractional part of %q`, str)
			}
			// Need to left shift frac so it's the right precision.
			if i.Precision > prec {
				frac.Mul(
					frac,
					big.NewInt(0).Exp(
						radix,
						big.NewInt(int64(i.Precision-prec)),
						nil,
					),
				)
			} else if prec > i.Precision {
				frac.Div(
					frac,
					big.NewInt(0).Exp(
						radix,
						big.NewInt(int64(prec-i.Precision)),
						nil,
					),
				)

			}
			// frac is x in frac = x/(radix^prec)
			// need to make it x in frac = x/(10^prec)
			radixPrec := big.NewInt(int64(i.InputRadix))
			radixPrec = radixPrec.Exp(radixPrec, big.NewInt(int64(i.Precision)), nil)
			factor := precisionToFactor(i.Precision)
			frac.Mul(frac, factor)
			frac.Div(frac, radixPrec)

			// Now we're in base 10. Slide the integer part to the
			// left and insert the fractional part.
			num.Mul(num, precisionToFactor(i.Precision))
			num.Add(num, frac)
			v.precision = i.Precision
		}
		n.dotSeen = false
	} else {
		_, ok := num.SetString(str, int(i.InputRadix))
		if !ok {
			return fmt.Errorf(`unable to unmarshal integer %q`, str)
		}
	}

	if n.sign {
		num.Neg(num)
		n.sign = false
	}
	v.intval = num
	v.UpdatePrecision(i.Precision)
	i.Stack.Push(&v)
	n.buff.Reset()
	n.State = OSNotHungry
	return nil
}
