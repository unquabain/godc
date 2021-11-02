package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

var Debug *log.Logger = nil

func debug(args ...interface{}) {
	if Debug == nil {
		return
	}
	Debug.Print(args...)
}

func main() {
	if os.Args[0] == `-d` {
		Debug = log.New(os.Stderr, `debug`, log.LstdFlags)
	}
	reader := bufio.NewReader(os.Stdin)
	interpreter := NewInterpreter()

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			if err != io.EOF {
				fmt.Println(`error reading command:`, err)
			}
			return
		}
		err = interpreter.Interpret(r)
		if err != nil {
			if err == ErrExitRequested {
				return
			}
			fmt.Println(`error processing command:`, err)
		}
	}
}
