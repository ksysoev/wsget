package cmd

import (
	"context"
	"fmt"
	"os/user"
	"path/filepath"

	"github.com/ksysoev/wsget/pkg/repo/macro"
	"github.com/spf13/cobra"
)

// createMacroDownloadRunner creates a runner function for executing a macro download command.
// It takes filename of type string which specifies the name of the file to save the macro.
// It returns a function that accepts a Cobra command and its arguments, and executes the macro download logic.
// It returns an error if the macro download command encounters an issue during execution.
func createMacroDownloadRunner(args *flags, filename *string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, unnamedArgs []string) error {
		return runMacroDownloadCommand(cmd.Context(), args, filename, unnamedArgs)
	}
}

// runMacroDownloadCommand downloads a macro configuration file from a given URL and saves it to a specified path.
// It takes a context, args of type *flags, name of type *string, and unnamedArgs of type []string.
// It returns an error if the URL is missing, the current user cannot be retrieved, the file creation fails,
// or the macro download encounters an issue such as invalid YAML or unsupported macro version.
func runMacroDownloadCommand(_ context.Context, args *flags, name *string, unnamedArgs []string) error {
	url := unnamedArgs[0]
	if url == "" {
		return fmt.Errorf("macro URL is required")
	}

	if args.configDir == "" {
		currentUser, err := user.Current()
		if err != nil {
			return fmt.Errorf("fail to get current user: %s", err)
		}

		args.configDir = filepath.Join(currentUser.HomeDir, defaultConfigDir)
	}

	path := filepath.Join(args.configDir, macroDir, *name)

	return macro.Download(path, url)
}
