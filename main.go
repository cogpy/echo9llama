package main

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/EchoCog/echollama/cmd"
)

func main() {
	cobra.CheckErr(cmd.NewCLI().ExecuteContext(context.Background()))
}
