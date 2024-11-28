package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ksysoev/wsget/pkg/cli"
	"github.com/ksysoev/wsget/pkg/clierrors"
	"github.com/ksysoev/wsget/pkg/command"
	"github.com/ksysoev/wsget/pkg/ws"
	"github.com/spf13/cobra"
)

// runConnectCmd creates and returns a function to execute the connect command.
// It takes a pointer to flags as an argument.
// It returns a function that takes a *cobra.Command and a slice of strings as arguments and returns an error.
// The returned function connects to a WebSocket server, initializes a CLI client, and runs it with the specified options.
// It returns an error if the URL is empty, the single response timeout is used without a request, the connection to the server fails, the CLI client fails to start, or the client fails to run.
// If the error is of type clierrors.Interrupted, it returns nil.
func runConnectCmd(args *flags) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, unnamedArgs []string) error {
		wsURL := unnamedArgs[0]

		if err := validateArgs(wsURL, args); err != nil {
			return err
		}

		wsConn, err := ws.NewWS(
			cmd.Context(),
			wsURL,
			ws.Options{
				SkipSSLVerification: args.insecure,
				Headers:             args.headers,
				Verbose:             args.verbose,
			},
		)
		if err != nil {
			return fmt.Errorf("unable to connect to the server: %w", err)
		}

		defer wsConn.Close()

		input := cli.NewKeyboard()

		client, err := cli.NewCLI(wsConn, input, os.Stdout)
		if err != nil {
			return fmt.Errorf("unable to start CLI: %w", err)
		}

		opts, err := initRunOptions(args)
		if err != nil {
			return err
		}

		if err = client.Run(*opts); err != nil {
			if errors.As(err, &clierrors.Interrupted{}) {
				return nil
			}

			return fmt.Errorf("failed to run client: %w", err)
		}

		return nil
	}
}

// validateArgs checks the validity of the provided WebSocket URL and flags.
// It takes wsURL of type string and args of type *flags.
// It returns an error if the wsURL is empty or if the single response timeout is set without a request.
// If wsURL is an empty string, it returns an error indicating that the URL is required.
// If args.waitResponse is non-negative and args.request is an empty string, it returns an error indicating that the single response timeout can only be used with a request.
func validateArgs(wsURL string, args *flags) error {
	if wsURL == "" {
		return fmt.Errorf("url is required")
	}

	if args.waitResponse >= 0 && args.request == "" {
		return fmt.Errorf("single response timeout could be used only with request")
	}

	return nil
}

// initRunOptions initializes and returns a RunOptions struct based on the provided flags.
// It takes a single parameter args of type *flags which contains the command-line arguments.
// It returns a pointer to cli.RunOptions and an error.
// It returns an error if it fails to open the specified output file.
func initRunOptions(args *flags) (opts *cli.RunOptions, err error) {
	opts = &cli.RunOptions{}

	if args.outputFile != "" {
		if opts.OutputFile, err = os.Create(args.outputFile); err != nil {
			return nil, fmt.Errorf("fail to open output file: %w", err)
		}
	}

	opts.Commands = createCommands(args)

	return opts, nil
}

// createCommands generates a slice of command.Executer based on the provided flags.
// It takes a single parameter args of type *flags, which contains the command-line arguments.
// It returns a slice of command.Executer, which represents the sequence of commands to be executed.
// If args.request is not empty, it creates a Send command and optionally adds WaitForResp and Exit commands if args.waitResponse is non-negative.
// If args.inputFile is not empty, it creates an InputFileCommand.
// If neither args.request nor args.inputFile is provided, it defaults to creating an Edit command.
func createCommands(args *flags) []command.Executer {
	var executers []command.Executer

	switch {
	case args.request != "":
		executers = []command.Executer{command.NewSend(args.request)}

		if args.waitResponse >= 0 {
			executers = append(
				executers,
				command.NewWaitForResp(time.Duration(args.waitResponse)*time.Second),
				command.NewExit(),
			)
		}
	case args.inputFile != "":
		executers = []command.Executer{command.NewInputFileCommand(args.inputFile)}
	default:
		executers = []command.Executer{command.NewEdit("")}
	}

	return executers
}
