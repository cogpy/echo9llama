//go:build experimental
// +build experimental

package main

import (
	"fmt"
	"github.com/EchoCog/echollama/core/scheme"
)

func main() {
	sm := scheme.NewSchemeMetamodel()
	sm.Start()

	// Test basic arithmetic
	result, err := sm.Eval("(+ 1 2 3)")
	if err != nil {
		tf("Error: %v\n", err)
	} else {
		tf("(+ 1 2 3) = %v\n", result)
	}

	// Test lambda
	result2, err2 := sm.Eval("((lambda (x) (+ x 10)) 5)")
	if err2 != nil {
		tf("Error: %v\n", err2)
	} else {
		tf("((lambda (x) (+ x 10)) 5) = %v\n", result2)
	}

	// Test define and use
	sm.Eval("(define square (lambda (x) (+ x x)))")
	result3, err3 := sm.Eval("(square 4)")
	if err3 != nil {
		tf("Error: %v\n", err3)
	} else {
		tf("(square 4) = %v\n", result3)
	}
}
