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
	if r < '0' {
		return false
	}
	if r > '9' {
		return false
	}
	return true
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
			err := num.UnmarshalText([]byte(parts[0]))
			if err != nil {
				return fmt.Errorf(`unable to unmarshal integer part of %q: %w`, str, err)
			}
		}
		prec = len(parts[1])
		if prec > 0 {
			err := frac.UnmarshalText([]byte(parts[1]))
			if err != nil {
				return fmt.Errorf(`unable to unmarshal fractional part of %q: %w`, str, err)
			}
			num.Mul(num, precisionToFactor(prec))
			num.Add(num, frac)
		}
		n.dotSeen = false
	} else {
		err := num.UnmarshalText([]byte(str))
		if err != nil {
			return fmt.Errorf(`unable to unmarshal integer %q: %w`, str, err)
		}
	}

	if n.sign {
		num.Neg(num)
		n.sign = false
	}
	v.intval = num
	v.precision = prec
	v.UpdatePrecision(i.Precision)
	i.Stack.Push(&v)
	n.buff.Reset()
	n.State = OSNotHungry
	return nil
}
