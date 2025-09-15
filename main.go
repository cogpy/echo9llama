package main

import (
	"fmt"
	"os"

	"github.com/EchoCog/echollama/cmd"
)

func main() {
	if err := cmd.NewCLI().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}