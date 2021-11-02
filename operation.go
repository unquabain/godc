package main

import (
	"fmt"
	"math/big"
)

var NotARegisterNameError = fmt.Errorf(`not a register name`)
var NotImplementedError = fmt.Errorf(`not implemented`)
var ValueNotNumericError = fmt.Errorf(`value is not numeric`)
var ContinueProcessingRune = fmt.Errorf(`rune should be processed as new operation`)

func ensureNumeric(vals ...*Value) error {
	for _, val := range vals {
		if val.Type != VTNumber {
			return ValueNotNumericError
		}
	}
	return nil
}

type Operation interface {
	Operate(*Interpreter, rune) (bool, error)
}

type OperationState bool

const (
	OSNotHungry OperationState = false
	OSHungry    OperationState = true
)

type RegisterOperation struct {
	State OperationState
	Func  func(stack, register *Stack) error
}

func isRegister(r rune) bool {
	if r < 'a' {
		return false
	}
	if r > 'z' {
		return false
	}
	return true
}

func (so *RegisterOperation) Operate(i *Interpreter, register rune) (bool, error) {
	if so.State == OSNotHungry {
		so.State = OSHungry
		return false, nil
	}
	defer func() { so.State = OSNotHungry }()

	if !isRegister(register) {
		return true, NotARegisterNameError
	}

	return true, so.Func(i.Stack, i.Registers[register])
}

type OperationAdapter func(*Interpreter) error

func (oa OperationAdapter) Operate(i *Interpreter, _ rune) (bool, error) {
	return true, oa(i)
}

func makeUnaryOperation(op func(*Value) ([]*Value, error)) Operation {
	return OperationAdapter(func(i *Interpreter) error {
		if i.Stack.Len() < 1 {
			return StackTooShortError
		}
		val := i.Stack.Pop()
		nums, err := op(val)
		if err != nil {
			i.Stack.Push(val)
			return err
		}
		for _, num := range nums {
			i.Stack.Push(num)
		}
		return nil
	})
}

func makeBinaryOperation(op func(*Value, *Value) ([]*Value, error)) Operation {
	return OperationAdapter(func(i *Interpreter) error {
		if i.Stack.Len() < 2 {
			return StackTooShortError
		}
		right, left := i.Stack.Pop(), i.Stack.Pop() // Note reverse order of left and right
		nums, err := op(left, right)
		if err != nil {
			i.Stack.Push(left)
			i.Stack.Push(right)
			return err
		}
		for _, num := range nums {
			i.Stack.Push(num)
		}
		return nil
	})
}

var QuitOperation = OperationAdapter(func(i *Interpreter) error {
	i.QuitLevel = 1
	return ExitRequestedError
})

var MacroQuitOperation = OperationAdapter(func(i *Interpreter) error {
	i.QuitLevel = 0
	if i.Stack.Len() > 0 {
		if i.Stack.Peek().Type == VTNumber {
			quitLevel := i.Stack.Pop().Int()
			i.QuitLevel = quitLevel
		}
	}
	return ExitRequestedError
})

var PrintOperation = OperationAdapter(func(i *Interpreter) error {
	p := i.Stack.Peek().Dup()
	p.UpdatePrecision(i.Precision)
	i.println(p)
	return nil
})

var PopAndPrintOperation = OperationAdapter(func(i *Interpreter) error {
	if i.Stack.Len() < 1 {
		return StackTooShortError
	}
	val := i.Stack.Pop()
	dup := val.Dup()
	dup.UpdatePrecision(i.Precision)
	i.print(dup)
	return nil
})

var PushLengthOperation = OperationAdapter(func(i *Interpreter) error {
	i.Stack.Push(&Value{intval: big.NewInt(int64(i.Stack.Len()))})
	return nil
})

var PrintStackOperation = OperationAdapter(func(i *Interpreter) error {
	for _, num := range i.Stack.values {
		dup := num.Dup()
		dup.UpdatePrecision(i.Precision)
		// dc prints stack in reverse order, so top-of-stack is top-of-list
		defer func(d *Value) { i.println(d) }(dup)
	}
	return nil
})

var ClearStackOperation = OperationAdapter(func(i *Interpreter) error {
	i.Stack.Clear()
	return nil
})

var AdditionOperation = makeBinaryOperation(func(left, right *Value) ([]*Value, error) {
	err := ensureNumeric(left, right)
	if err != nil {
		return nil, err
	}
	err = left.Add(right)
	if err != nil {
		return nil, err
	}
	return []*Value{left}, nil
})

var SubtractionOperation = makeBinaryOperation(func(left, right *Value) ([]*Value, error) {
	err := ensureNumeric(left, right)
	if err != nil {
		return nil, err
	}
	err = left.Subtract(right)
	if err != nil {
		return nil, err
	}
	return []*Value{left}, nil
})

var MultiplicationOperation = makeBinaryOperation(func(left, right *Value) ([]*Value, error) {
	err := ensureNumeric(left, right)
	if err != nil {
		return nil, err
	}
	err = left.Multiply(right)
	if err != nil {
		return nil, err
	}
	return []*Value{left}, nil
})

var DivisionOperation = makeBinaryOperation(func(left, right *Value) ([]*Value, error) {
	err := ensureNumeric(left, right)
	if err != nil {
		return nil, err
	}
	err = left.Divide(right)
	if err != nil {
		return nil, err
	}
	return []*Value{left}, nil
})

var ModuloOperation = makeBinaryOperation(func(left, right *Value) ([]*Value, error) {
	err := ensureNumeric(left, right)
	if err != nil {
		return nil, err
	}
	_, r, err := left.QuotientRemainder(right)
	if err != nil {
		return nil, err
	}
	return []*Value{r}, nil
})

var QuotientRemainderOperation = makeBinaryOperation(func(left, right *Value) ([]*Value, error) {
	err := ensureNumeric(left, right)
	if err != nil {
		return nil, err
	}
	q, r, err := left.QuotientRemainder(right)
	if err != nil {
		return nil, err
	}
	return []*Value{r, q}, nil
})

var ExponentOperation = makeBinaryOperation(func(left, right *Value) ([]*Value, error) {
	err := ensureNumeric(left, right)
	if err != nil {
		return nil, err
	}
	err = left.Exponent(right)
	if err != nil {
		return nil, err
	}
	return []*Value{left}, nil
})

var ModExponentOperation = OperationAdapter(func(i *Interpreter) error {
	if i.Stack.Len() < 3 {
		return StackTooShortError
	}
	e, m, n := i.Stack.Pop(), i.Stack.Pop(), i.Stack.Pop()
	err := n.ModExponent(m, e)
	if err != nil {
		return err
	}
	i.Stack.Push(n)
	return nil
})

var SqrtOperation = makeUnaryOperation(func(val *Value) ([]*Value, error) {
	err := val.Sqrt()
	return []*Value{val}, err
})

var DuplicationOperation = makeUnaryOperation(func(val *Value) ([]*Value, error) {
	return []*Value{val, val.Dup()}, nil
})

var ReverseOperation = makeBinaryOperation(func(left, right *Value) ([]*Value, error) {
	return []*Value{right, left}, nil
})

var MoveToRegisterOperation = &RegisterOperation{
	Func: func(stack, register *Stack) error {
		if stack.Len() < 1 {
			return StackTooShortError
		}
		register.Clear()
		register.Push(stack.Pop())
		return nil
	},
}

var MoveFromRegisterOperation = &RegisterOperation{
	Func: func(stack, register *Stack) error {
		if register.Len() < 1 {
			return StackTooShortError
		}
		stack.Push(register.Peek())
		return nil
	},
}

var MoveToRegisterStackOperation = &RegisterOperation{
	Func: func(stack, register *Stack) error {
		if stack.Len() < 1 {
			return StackTooShortError
		}
		register.Push(stack.Pop())
		return nil
	},
}

var MoveFromRegisterStackOperation = &RegisterOperation{
	Func: func(stack, register *Stack) error {
		if register.Len() < 1 {
			return StackTooShortError
		}
		stack.Push(register.Pop())
		return nil
	},
}

var SetPrecisionOperation = OperationAdapter(func(i *Interpreter) error {
	if i.Stack.Len() < 1 {
		return StackTooShortError
	}
	p := i.Stack.Pop()
	err := ensureNumeric(p)
	if err != nil {
		return err
	}
	i.Precision = p.Int()
	return nil
})

var GetPrecisionOperation = OperationAdapter(func(i *Interpreter) error {
	i.Stack.Push(&Value{intval: big.NewInt(int64(i.Precision))})
	return nil
})

var NotImplementedOperation = OperationAdapter(func(_ *Interpreter) error {
	return NotImplementedError
})

type CommentOperatorType struct{}

func (CommentOperatorType) Operate(_ *Interpreter, r rune) (bool, error) {
	return r == '\n', nil
}

var CommentOperator CommentOperatorType

type StringBuilder struct {
	OperationState
	Value
}

func (sb *StringBuilder) Operate(i *Interpreter, r rune) (bool, error) {
	if r == '[' {
		sb.OperationState = OSHungry
		sb.Value.Type = VTString
		sb.Value.strval = []rune{}
		return false, nil
	}
	if r == ']' {
		sb.OperationState = OSNotHungry
		i.Stack.Push((&sb.Value).Dup())
		return true, nil
	}
	if sb.OperationState == OSNotHungry {
		return true, nil
	}
	sb.Value.strval = append(sb.Value.strval, r)
	return false, nil
}

var StringBuilderOperation = new(StringBuilder)
var NumberBuilderOperation = NewNumberBuilder()

var ExecuteMacroOperation = OperationAdapter(func(i *Interpreter) error {
	if i.Stack.Len() < 1 {
		return StackTooShortError
	}
	val := i.Stack.Pop()
	if val.Type != VTString {
		i.Stack.Push(val)
		return nil
	}

	for _, r := range val.strval {
		err := i.Interpret(r)
		if err != nil {
			if err == ExitRequestedError {
				if i.QuitLevel == 0 {
					continue
				}
				i.QuitLevel--
			}
			return err
		}
	}
	i.Interpret(' ') // Make sure to flush any digit in the works
	return nil
})
