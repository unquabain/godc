package main

import (
	"fmt"
	"io"
	"os"
)

// ErrStackTooShort is returned when an operation wants more
// arguments than are available.
var ErrStackTooShort = fmt.Errorf(`stack too short`)

// ErrExitRequested is returned when an operation asks to quit
// the program or the currently running macro.
var ErrExitRequested = fmt.Errorf(`goodbye`)

// Interpreter interprets commands and macros and maintains
// the main stack and the various registers.
type Interpreter struct {
	Stack            *Stack
	Registers        map[rune]*Stack
	NumberBuilder    *NumberBuilder
	Precision        int
	CurrentOperation Operation
	Operations       map[rune]Operation
	output           io.Writer
	QuitLevel        int
	InputRadix       uint8
	OutputRadix      uint8
}

// NewInterpreter intitializes an interpreter and its
// registers.
func NewInterpreter() *Interpreter {
	i := new(Interpreter)
	i.Stack = new(Stack)
	i.Registers = make(map[rune]*Stack)
	for r := 'a'; r <= 'z'; r++ {
		i.Registers[r] = new(Stack)
	}
	i.output = os.Stdout
	i.InputRadix = 10
	i.OutputRadix = 10
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
		'A': NumberBuilderOperation,
		'B': NumberBuilderOperation,
		'C': NumberBuilderOperation,
		'D': NumberBuilderOperation,
		'E': NumberBuilderOperation,
		'F': NumberBuilderOperation,
		'G': NumberBuilderOperation,
		'H': NumberBuilderOperation,
		'.': NumberBuilderOperation,
		'_': NumberBuilderOperation,
		'q': QuitOperation,
		'p': PrintOperation,
		'P': PrintRawOperation, // Prints the raw bytes in the number representation
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
		'i': SetInputRadixOperation,        // TODO: set input radix
		'o': SetOutputRadixOperation,       // TODO: set output radix
		'I': GetInputRadixOperation,        // TODO: get input radix
		'O': GetOutputRadixOperation,       // TODO: get output radix
		'[': StringBuilderOperation,        // begin string
		'a': NotImplementedOperation,       // TODO: chr(i) (for int) or s[0] (for string)
		'x': ExecuteMacroOperation,         // execute macro
		'>': ExecuteMacroIfGTOperation,     // conditional execute macro
		'!': ExecuteMacroNegativeOperation, // conditional execute macro
		'<': ExecuteMacroIfLTOperation,     // conditional execute macro
		'=': ExecuteMacroIfEqOperation,     // conditional execute macro
		'?': NotImplementedOperation,       // TODO: get input from STDIN
		'Q': MacroQuitOperation,            // exit n macros
		'Z': NotImplementedOperation,       // TODO: len(v.String())
		'X': NotImplementedOperation,       // TODO: number of fractional digits.
		'z': PushLengthOperation,
		'#': CommentOperator,
		':': NotImplementedOperation, // TODO: push to specific index in register
		';': NotImplementedOperation, // TODO: fetch from specific index in register
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

// Interpret interprets one rune from input or a macro.
// The error returned might include ErrExitRequested,
// if the command was q or Q, or if a submacro returned
// that. The QuitLevel command sould be consulted
// by macros to determine whether to raise that error
// to calling macros or to continue on. Most other
// errors are not fatal. They should be printed and
// execution should continue.
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
	if err == ErrContinueProcessingRune {
		if i.CurrentOperation != nil {
			panic(`operation returned !finished, ErrContinueProcessingRune`)
		}
		return i.Interpret(r)
	}
	return err
}

// InterpretMacro runs a macro sequence. The only difference between
// this and the main loop is that the QuitLevel number is consulted
// to determine how many layers of macro should be terminated when
// a q or Q command is encountered.
func (i *Interpreter) InterpretMacro(macro []rune) error {
	for _, r := range macro {
		err := i.Interpret(r)
		if err != nil {
			if err == ErrExitRequested {
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
