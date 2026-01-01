//go:build ignore
// +build ignore

package runner

import (
	"github.com/cogpy/echo9llama/runner/llamarunner"
	"github.com/cogpy/echo9llama/runner/ollamarunner"
)

func Execute(args []string) error {
	if args[0] == "runner" {
		args = args[1:]
	}

	var newRunner bool
	if args[0] == "--ollama-engine" {
		args = args[1:]
		newRunner = true
	}

	if newRunner {
		return ollamarunner.Execute(args)
	} else {
		return llamarunner.Execute(args)
	}
}
