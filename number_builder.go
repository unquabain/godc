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

func (n *NumberBuilder) reset() {
	n.buff.Reset()
	n.dotSeen = false
	n.sign = false
	n.State = OSNotHungry
}

// Flush finalizes the number and pushes it onto the stack.
func (n *NumberBuilder) Flush(i *Interpreter) error {
	var v Value
	s := n.buff.String()
	numerator := &big.Int{}
	denominator := &big.Int{}

	if n.dotSeen {
		pointPos := strings.LastIndex(s, `.`) + 1
		fracDigits := len(s) - pointPos
		withoutPoints := strings.Replace(s, `.`, ``, 1)
		_, ok := numerator.SetString(withoutPoints, int(i.InputRadix))
		if !ok {
			return fmt.Errorf(`could not parse %s as a radix %d integer`, s, i.InputRadix)
		}
		denominator.Exp(big.NewInt(int64(i.InputRadix)), big.NewInt(int64(fracDigits)), nil)
	} else {
		_, ok := numerator.SetString(s, int(i.InputRadix))
		if !ok {
			return fmt.Errorf(`could not parse %s as a radix %d integer`, s, i.InputRadix)
		}
		denominator.SetInt64(1)
	}
	num := (&big.Rat{}).SetFrac(numerator, denominator)

	if n.sign {
		num.Neg(num)
	}

	v.numval = num

	i.Stack.Push(&v)
	n.reset()
	return nil
}
