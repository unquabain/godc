package main

import (
	"fmt"
	"io"
	"os"
)

var StackTooShortError = fmt.Errorf(`stack too short`)
var ExitRequestedError = fmt.Errorf(`goodbye`)

type InterpreterState uint8

const (
	ISImmediate InterpreterState = iota
	ISRegisterLoad
	ISRegisterStackPush
	ISRegisterRestore
	ISRegisterStackPop
	ISRegisterStackPeek
	ISRegisterStackPrint
)

type Interpreter struct {
	Stack            *Stack
	Registers        map[rune]*Stack
	NumberBuilder    *NumberBuilder
	Precision        int
	State            InterpreterState
	CurrentOperation Operation
	Operations       map[rune]Operation
	output           io.Writer
	QuitLevel        int
}

func NewInterpreter() *Interpreter {
	i := new(Interpreter)
	i.Stack = new(Stack)
	i.Registers = make(map[rune]*Stack)
	for r := 'a'; r <= 'z'; r++ {
		i.Registers[r] = new(Stack)
	}
	i.output = os.Stdout
	i.Operations = map[rune]Operation{
		'0': NumberBuilderOperation,
		'1': NumberBuilderOperation,
		'2': NumberBuilderOperation,
		'3': NumberBuilderOperation,
		'4': NumberBuilderOperation,
		'5': NumberBuilderOperation,
		'6': NumberBuilderOperation,
		'7': NumberBuilderOperation,
		'8': NumberBuilderOperation,
		'9': NumberBuilderOperation,
		'.': NumberBuilderOperation,
		'_': NumberBuilderOperation,
		'q': QuitOperation,
		'p': PrintOperation,
		'n': PopAndPrintOperation,
		'f': PrintStackOperation,
		'+': AdditionOperation,
		'-': SubtractionOperation,
		'*': MultiplicationOperation,
		'/': DivisionOperation,
		'%': ModuloOperation,            // modulo
		'~': QuotientRemainderOperation, // quotient, remainder
		'^': ExponentOperation,          // exponentiation
		'|': ModExponentOperation,       // (a^b) % c
		'v': SqrtOperation,              // square root
		'c': ClearStackOperation,
		'd': DuplicationOperation,
		'r': ReverseOperation,
		's': MoveToRegisterOperation,
		'l': MoveFromRegisterOperation,
		'S': MoveToRegisterStackOperation,
		'L': MoveFromRegisterStackOperation,
		'k': SetPrecisionOperation,
		'i': NotImplementedOperation,   // set input radix
		'o': NotImplementedOperation,   // set output radix
		'I': NotImplementedOperation,   // get input radix
		'O': NotImplementedOperation,   // get output radix
		'[': StringBuilderOperation,    // begin string
		'a': NotImplementedOperation,   // i to a
		'x': ExecuteMacroOperation,     // execute macro
		'>': ExecuteMacroIfGTOperation, // conditional execute macro
		'!': NotImplementedOperation,   // conditional execute macro
		'<': ExecuteMacroIfLTOperation, // conditional execute macro
		'=': ExecuteMacroIfEqOperation, // conditional execute macro
		'?': NotImplementedOperation,   // conditional execute macro
		'Q': MacroQuitOperation,        // exit n macros
		'Z': NotImplementedOperation,   // replace n with Value of digits in n
		'X': NotImplementedOperation,   // replace n with Value of fractional digits
		'z': PushLengthOperation,
		'#': CommentOperator,
		':': NotImplementedOperation, // push to specific index in register
		';': NotImplementedOperation, // fetch from specific index in register
	}
	return i
}

func (i *Interpreter) print(args ...interface{}) {
	fmt.Fprint(i.output, args...)
}

func (i *Interpreter) printf(format string, args ...interface{}) {
	fmt.Fprintf(i.output, format, args...)
}

func (i *Interpreter) println(args ...interface{}) {
	fmt.Fprintln(i.output, args...)
}

func (i *Interpreter) Interpret(r rune) error {
	var (
		op Operation
		ok bool
	)
	if i.CurrentOperation != nil {
		op = i.CurrentOperation
	} else {
		op, ok = i.Operations[r]
		if !ok {
			return nil
		}
	}
	finished, err := op.Operate(i, r)
	if finished {
		i.CurrentOperation = nil
	} else {
		i.CurrentOperation = op
	}
	if err == ContinueProcessingRune {
		if i.CurrentOperation != nil {
			panic(`operation returned !finished, ContinueProcessingRune`)
		}
		return i.Interpret(r)
	}
	return err
}

func (i *Interpreter) InterpretMacro(macro []rune) error {
	for _, r := range macro {
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
}
