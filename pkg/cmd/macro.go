package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ksysoev/wsget/pkg/repo/macro"
	"github.com/spf13/cobra"
)

// createMacroDownloadRunner creates a runner function for executing a macro download command.
// It takes filename of type string which specifies the name of the file to save the macro.
// It returns a function that accepts a Cobra command and its arguments, and executes the macro download logic.
// It returns an error if the macro download command encounters an issue during execution.
func createMacroDownloadRunner(args *flags, filename string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, unnamedArgs []string) error {
		return runMacroDownloadCommand(cmd.Context(), args, filename, unnamedArgs)
	}
}

func runMacroDownloadCommand(_ context.Context, args *flags, name string, unnamedArgs []string) error {
	url := unnamedArgs[0]
	if url == "" {
		return fmt.Errorf("macro URL is required")
	}

	path := filepath.Join(args.configDir, macroDir, name)

	macroRepo := macro.NewMacro(nil)

	err := macroRepo.Download(path, url)

	return fmt.Errorf("fail to download macro: %w", err)
}
