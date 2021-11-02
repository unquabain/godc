package main

import (
	"fmt"
	"math/big"
)

// ErrNotARegisterName is returned when a register operation
// receives a rune that isn't a register name (a lowercase letter)
var ErrNotARegisterName = fmt.Errorf(`not a register name`)

// ErrNotImplemented occurs when the user tries to use an operation
// that dc understands, but I haven't gotten to yet.
var ErrNotImplemented = fmt.Errorf(`not implemented`)

// ErrValueNotNumeric is returned when you try to do a number thing
// with a string.
var ErrValueNotNumeric = fmt.Errorf(`value is not numeric`)

// ErrValueNotString is returned when you try to do a string thing
// with a number.
var ErrValueNotString = fmt.Errorf(`value is not a string`)

// ErrContinueProcessingRune is returned by operations that gobble
// up input until they encounter something they don't recognize.
// It indicates that the operation is completed, but the rune should
// be reprocessed to be picked up possibly by another operation.
var ErrContinueProcessingRune = fmt.Errorf(`rune should be processed as new operation`)

func ensureNumeric(vals ...*Value) error {
	for _, val := range vals {
		if val.Type != VTNumber {
			return ErrValueNotNumeric
		}
	}
	return nil
}

// Operation consumes runes and manipulates stacks and registers.
type Operation interface {
	// Operate operates on a rune. It returns a bool indicating
	// whether or not the Operation is expecting more input, and,
	// of course, an error.
	Operate(*Interpreter, rune) (bool, error)
}

// Whether the operation is expecting more runes
type OperationState bool

const (
	OSNotHungry OperationState = false
	OSHungry    OperationState = true
)

func isRegister(r rune) bool {
	if r < 'a' {
		return false
	}
	if r > 'z' {
		return false
	}
	return true
}

// An operation that takes a post-positional argument, that
// is a register to operate on. This violates the backward-only
// operation of most dc operations. You could implement e.g.
// "save value 12 to register a" as 12[a]s, but dc uses 12sa.
type RegisterOperation struct {
	State OperationState
	Func  func(stack, register *Stack) error
}

// Operate implements the Operator interface.
// It returns false on its first call to indicate that it's
// waiting for a second rune, that specifies the register to
// operate on.
func (so *RegisterOperation) Operate(i *Interpreter, register rune) (bool, error) {
	if so.State == OSNotHungry {
		so.State = OSHungry
		return false, nil
	}
	defer func() { so.State = OSNotHungry }()

	if !isRegister(register) {
		return true, ErrNotARegisterName
	}

	return true, so.Func(i.Stack, i.Registers[register])
}

// Most operations are not hungry, so the operator pattern helps
// keep their definitions simple.
type OperationAdapter func(*Interpreter) error

// This implements the Operation interface by discarding unused arguments
// and calling the original function.
func (oa OperationAdapter) Operate(i *Interpreter, _ rune) (bool, error) {
	return true, oa(i)
}

func makeUnaryOperation(op func(*Value) ([]*Value, error)) Operation {
	return OperationAdapter(func(i *Interpreter) error {
		if i.Stack.Len() < 1 {
			return ErrStackTooShort
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
			return ErrStackTooShort
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

// QuitOperation implements the 'q' command
var QuitOperation = OperationAdapter(func(i *Interpreter) error {
	i.QuitLevel = 1
	return ErrExitRequested
})

// MacroQuitOperation implements the 'Q' command
var MacroQuitOperation = OperationAdapter(func(i *Interpreter) error {
	i.QuitLevel = 0
	if i.Stack.Len() > 0 {
		if i.Stack.Peek().Type == VTNumber {
			quitLevel := i.Stack.Pop().Int()
			i.QuitLevel = quitLevel
		}
	}
	return ErrExitRequested
})

// PrintOperation implements the 'p' command.
var PrintOperation = OperationAdapter(func(i *Interpreter) error {
	p := i.Stack.Peek().Dup()
	p.UpdatePrecision(i.Precision)
	i.println(p)
	return nil
})

// PopAndPrintOperation implements the 'n' command
var PopAndPrintOperation = OperationAdapter(func(i *Interpreter) error {
	if i.Stack.Len() < 1 {
		return ErrStackTooShort
	}
	val := i.Stack.Pop()
	dup := val.Dup()
	dup.UpdatePrecision(i.Precision)
	i.print(dup)
	return nil
})

// PushLengthOperation implements the 'z' command.
var PushLengthOperation = OperationAdapter(func(i *Interpreter) error {
	i.Stack.Push(&Value{intval: big.NewInt(int64(i.Stack.Len()))})
	return nil
})

// PrintStackOperation implements the 'f' command.
var PrintStackOperation = OperationAdapter(func(i *Interpreter) error {
	for _, num := range i.Stack.values {
		dup := num.Dup()
		dup.UpdatePrecision(i.Precision)
		// dc prints stack in reverse order, so top-of-stack is top-of-list
		defer func(d *Value) { i.println(d) }(dup)
	}
	return nil
})

// ClearStackOperation implements the 'c' command.
var ClearStackOperation = OperationAdapter(func(i *Interpreter) error {
	i.Stack.Clear()
	return nil
})

// AdditionOperation implements the '+' command.
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

// SubtrationOperation implements the '-' command.
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

// MultiplicationOperation implements the '*' command.
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

// DivisionOperation implements the "/" command.
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

// ModuloOperation implements the '%' command.
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

// QuotientRemainderOperation implements the '~' command.
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

// ExponentOperation implements the '^' command.
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

// ModExponentOperation implements the '|' command.
var ModExponentOperation = OperationAdapter(func(i *Interpreter) error {
	if i.Stack.Len() < 3 {
		return ErrStackTooShort
	}
	e, m, n := i.Stack.Pop(), i.Stack.Pop(), i.Stack.Pop()
	err := n.ModExponent(m, e)
	if err != nil {
		return err
	}
	i.Stack.Push(n)
	return nil
})

// SqrtOperation implements the 'v' command.
var SqrtOperation = makeUnaryOperation(func(val *Value) ([]*Value, error) {
	err := val.Sqrt()
	return []*Value{val}, err
})

// DuplicationOperation implements the 'd' command.
var DuplicationOperation = makeUnaryOperation(func(val *Value) ([]*Value, error) {
	return []*Value{val, val.Dup()}, nil
})

// ReverseOperation implements the 'r' command.
var ReverseOperation = makeBinaryOperation(func(left, right *Value) ([]*Value, error) {
	return []*Value{right, left}, nil
})

// MoveToRegisterOperation implements the 's' (save) command.
var MoveToRegisterOperation = &RegisterOperation{
	Func: func(stack, register *Stack) error {
		if stack.Len() < 1 {
			return ErrStackTooShort
		}
		register.Clear()
		register.Push(stack.Pop())
		return nil
	},
}

// MoveFromRegister implements the 'l' (load) command.
var MoveFromRegisterOperation = &RegisterOperation{
	Func: func(stack, register *Stack) error {
		if register.Len() < 1 {
			return ErrStackTooShort
		}
		stack.Push(register.Peek())
		return nil
	},
}

// MoveToRegisterStackOperation implements the 'S' command.
var MoveToRegisterStackOperation = &RegisterOperation{
	Func: func(stack, register *Stack) error {
		if stack.Len() < 1 {
			return ErrStackTooShort
		}
		register.Push(stack.Pop())
		return nil
	},
}

// MoveFromRegisterStackOperation implements the 'L' command.
var MoveFromRegisterStackOperation = &RegisterOperation{
	Func: func(stack, register *Stack) error {
		if register.Len() < 1 {
			return ErrStackTooShort
		}
		stack.Push(register.Pop())
		return nil
	},
}

// SetPrecisionOperation implements the 'k' command.
var SetPrecisionOperation = OperationAdapter(func(i *Interpreter) error {
	if i.Stack.Len() < 1 {
		return ErrStackTooShort
	}
	p := i.Stack.Pop()
	err := ensureNumeric(p)
	if err != nil {
		return err
	}
	i.Precision = p.Int()
	return nil
})

// GetPrecisionOperation implements the 'K' command.
var GetPrecisionOperation = OperationAdapter(func(i *Interpreter) error {
	i.Stack.Push(&Value{intval: big.NewInt(int64(i.Precision))})
	return nil
})

// NotImplementedOperation returns an error, and is a placeholder for any
// commands dc supports, but I don't yet support.
var NotImplementedOperation = OperationAdapter(func(_ *Interpreter) error {
	return ErrNotImplemented
})

// CommentOperatorType is an OperatorType that ignores everything up to a newline.
type CommentOperatorType struct{}

// Operate implements the Operator interface.
func (CommentOperatorType) Operate(_ *Interpreter, r rune) (bool, error) {
	return r == '\n', nil
}

// CommentOperator implements the '#' command.
var CommentOperator CommentOperatorType

// StringBuilder facilitates interpreting brace-delimited strings.
// It gobbles up runes until it spots a ']'
type StringBuilder struct {
	OperationState
	Value
}

// Operate implements the Operator interface.
// TODO: dc supports nested brackets in strings.
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

// StringBuilderOperation implements the '[' command.
var StringBuilderOperation = new(StringBuilder)

// NumberBuilderOperation gobbles up digits and builds a number.
var NumberBuilderOperation = NewNumberBuilder()

// ExecuteMacroOperation implements the 'x' command.
var ExecuteMacroOperation = OperationAdapter(func(i *Interpreter) error {
	if i.Stack.Len() < 1 {
		return ErrStackTooShort
	}
	val := i.Stack.Pop()
	if val.Type != VTString {
		i.Stack.Push(val)
		return nil
	}
	return i.InterpretMacro(val.strval)
})

// MacroOperation supports execution of conditional macros.
// Positive conditional macros (e.g. >) are supported directly, and
// negative conditional macros (e.g. !>) are supported with the aid
// of the NegativeMacroOperation type.
type MacroOperation struct {
	State OperationState
	// Whether the previous two values in the stack indicate the macro
	// should be executed. Values are (top, next-to-top)
	Predicate func(*Value, *Value) bool
}

// Operate implements the Operation interface.
// This handles the stack and argument type checking.
func (so *MacroOperation) Operate(i *Interpreter, register rune) (bool, error) {
	if so.State == OSNotHungry {
		so.State = OSHungry
		return false, nil
	}
	defer func() { so.State = OSNotHungry }()

	if !isRegister(register) {
		return true, ErrNotARegisterName
	}

	if i.Stack.Len() < 2 {
		return true, ErrStackTooShort
	}

	reg := i.Registers[register]
	if reg.Len() < 1 {
		return true, ErrStackTooShort
	}
	if reg.Peek().Type != VTString {
		return true, ErrValueNotString
	}

	left, right := i.Stack.Pop(), i.Stack.Pop()
	if left.Type != VTNumber || right.Type != VTNumber {
		return true, ErrValueNotNumeric
	}

	if !so.Predicate(left, right) {
		return true, nil
	}

	macro := reg.Pop().strval
	i.CurrentOperation = nil
	return true, i.InterpretMacro(macro)
}

// ExecuteMacroIfGTOperation implements the '>' command.
var ExecuteMacroIfGTOperation = &MacroOperation{
	Predicate: func(left, right *Value) bool {
		left.MatchPrecision(right)
		return left.intval.Cmp(right.intval) > 0
	},
}

// ExecuteMacroIfLTOperation implements the '<' command.
var ExecuteMacroIfLTOperation = &MacroOperation{
	Predicate: func(left, right *Value) bool {
		left.MatchPrecision(right)
		return left.intval.Cmp(right.intval) < 0
	},
}

// ExecuteMacroIfEqOperation implements the '=' command.
var ExecuteMacroIfEqOperation = &MacroOperation{
	Predicate: func(left, right *Value) bool {
		left.MatchPrecision(right)
		return left.intval.Cmp(right.intval) == 0
	},
}

// NegativeMacroOperation implements the negative conditional
// macro commands by gobbling up the '!' and negating the
// predicate.
type NegativeMacroOperation struct {
	Op    *MacroOperation
	State OperationState
}

func negate(pred func(*Value, *Value) bool) func(*Value, *Value) bool {
	return func(left, right *Value) bool {
		return !pred(left, right)
	}
}

// Operate implements the Operator interface.
// This determines which of the MacroOperation type Operations
// defined above are to be negated, creates a negative predicate,
// then proxies that MacroOperation.
//
// TODO: Since this operator will handle the '!' command, which has
// a second meaning, it must also handle the shell execute meaning
// of '!'
func (so *NegativeMacroOperation) Operate(i *Interpreter, r rune) (bool, error) {
	if so.State == OSNotHungry {
		so.State = OSHungry
		return false, nil
	}
	if so.Op == nil {
		so.Op = &MacroOperation{}
		switch r {
		case '<':
			so.Op.Predicate = negate(ExecuteMacroIfLTOperation.Predicate)
		case '>':
			so.Op.Predicate = negate(ExecuteMacroIfGTOperation.Predicate)
		case '=':
			so.Op.Predicate = negate(ExecuteMacroIfEqOperation.Predicate)
		default:
			// TODO: read to newline and execute in subshell
			return false, ErrNotImplemented
		}
	}
	finished, err := so.Op.Operate(i, r)
	if finished {
		so.State = OSNotHungry
		so.Op = nil
	} else {
		so.State = OSHungry
	}
	return finished, err
}

// This implements all multi-rune commands beginning with '!'
var ExecuteMacroNegativeOperation = new(NegativeMacroOperation)
