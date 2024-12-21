package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// createMacroDownloadRunner creates a runner function for executing a macro download command.
// It takes filename of type string which specifies the name of the file to save the macro.
// It returns a function that accepts a Cobra command and its arguments, and executes the macro download logic.
// It returns an error if the macro download command encounters an issue during execution.
func createMacroDownloadRunner(filename string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, unnamedArgs []string) error {
		return runMacroDownloadCommand(cmd.Context(), filename, unnamedArgs)
	}
}

func runMacroDownloadCommand(_ context.Context, _ string, _ []string) error {
	return fmt.Errorf("not implemented")
}
