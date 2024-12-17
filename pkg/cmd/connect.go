package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/ksysoev/wsget/pkg/core"
	command2 "github.com/ksysoev/wsget/pkg/core/command"
	"github.com/ksysoev/wsget/pkg/core/edit"
	"github.com/ksysoev/wsget/pkg/core/formater"
	"github.com/ksysoev/wsget/pkg/input"
	"github.com/ksysoev/wsget/pkg/repo"
	"github.com/ksysoev/wsget/pkg/ws"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const (
	macroDir           = "macro"
	historyFilename    = "history"
	historyCmdFilename = "cmd_history"
	configDirMode      = 0o755
	defaultConfigDir   = ".wsget"
)

// createConnectRunner creates a runner function for the connect command.
// It takes a single parameter args of type *flags.
// It returns a function that takes a *cobra.Command and a slice of strings, and returns an error.
// The returned function calls runConnectCmd with the provided command, args, and unnamedArgs.
// It returns an error if runConnectCmd encounters any issues.
func createConnectRunner(args *flags) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, unnamedArgs []string) error {
		return runConnectCmd(cmd.Context(), args, unnamedArgs)
	}
}

// runConnectCmd establishes a WebSocket connection and starts a CLI client session.
// It takes ctx of type context.Context, args of type *flags, and unnamedArgs of type []string.
// It returns an error if the WebSocket connection cannot be established, the CLI cannot be started, or the client fails to run.
// It returns nil if the client is interrupted gracefully.
func runConnectCmd(ctx context.Context, args *flags, unnamedArgs []string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wsURL := unnamedArgs[0]

	if err := validateArgs(wsURL, args); err != nil {
		return err
	}

	wsOpts := ws.Options{
		SkipSSLVerification: args.insecure,
		Headers:             args.headers,
		MaxMessageSize:      args.maxMsgSize,
	}

	if args.verbose {
		wsOpts.Output = os.Stdout
	}

	wsConn, err := ws.New(wsURL, wsOpts)
	if err != nil {
		return fmt.Errorf("unable to connect to the server: %w", err)
	}

	defer func() { _ = wsConn.Close() }()

	if args.configDir == "" {
		currentUser, err := user.Current()
		if err != nil {
			return fmt.Errorf("fail to get current user: %s", err)
		}

		args.configDir = filepath.Join(currentUser.HomeDir, defaultConfigDir)
	}

	if err = os.MkdirAll(filepath.Join(args.configDir, macroDir), configDirMode); err != nil {
		return fmt.Errorf("fail to get current user: %s", err)
	}

	history, err := repo.LoadFromFile(filepath.Join(args.configDir, historyFilename))
	if err != nil {
		return fmt.Errorf("fail to load history: %s", err)
	}

	defer func() { _ = history.Close() }()

	cmdHistory, err := repo.LoadFromFile(filepath.Join(args.configDir, historyCmdFilename))
	if err != nil {
		return fmt.Errorf("fail to load command history: %s", err)
	}

	defer func() { _ = cmdHistory.Close() }()

	macro, err := command2.LoadMacroForDomain(filepath.Join(args.configDir, macroDir), wsConn.Hostname())
	if err != nil {
		return fmt.Errorf("fail to load macro: %s", err)
	}

	if macro != nil {
		cmdHistory.AddWordsToIndex(macro.GetNames())
	}

	editor := edit.NewMultiMode(os.Stdout, history, cmdHistory)
	cmdFactory := command2.NewFactory(macro)

	client := core.NewCLI(cmdFactory, wsConn, os.Stdout, editor, formater.NewFormat())

	keyboard := input.NewKeyboard(client)
	defer keyboard.Close()

	opts, err := initRunOptions(args)
	if err != nil {
		return err
	}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return keyboard.Run(ctx)
	})

	eg.Go(func() error {
		return wsConn.Connect(ctx)
	})

	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case <-wsConn.Ready():
		}

		return client.Run(ctx, *opts)
	})

	err = eg.Wait()

	if errors.Is(err, context.Canceled) || errors.Is(err, core.ErrInterrupted) {
		return nil
	}

	fmt.Println("Error:", err)

	return nil
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
func initRunOptions(args *flags) (opts *core.RunOptions, err error) {
	opts = &core.RunOptions{}

	if args.outputFile != "" {
		if opts.OutputFile, err = os.Create(args.outputFile); err != nil {
			return nil, fmt.Errorf("fail to open output file: %w", err)
		}
	}

	opts.Commands = createCommands(args)

	return opts, nil
}

// createCommands generates a slice of core.Executer based on the provided flags.
// It takes a single parameter args of type *flags, which contains the command-line arguments.
// It returns a slice of core.Executer, which represents the sequence of commands to be executed.
// If args.request is not empty, it creates a Send command and optionally adds WaitForResp and Exit commands if args.waitResponse is non-negative.
// If args.inputFile is not empty, it creates an InputFileCommand.
// If neither args.request nor args.inputFile is provided, it defaults to creating an Edit command.
func createCommands(args *flags) []core.Executer {
	var executers []core.Executer

	switch {
	case args.request != "":
		executers = []core.Executer{command2.NewSend(args.request)}

		if args.waitResponse >= 0 {
			executers = append(
				executers,
				command2.NewWaitForResp(time.Duration(args.waitResponse)*time.Second),
				command2.NewExit(),
			)
		}
	case args.inputFile != "":
		executers = []core.Executer{command2.NewInputFileCommand(args.inputFile)}
	default:
		executers = []core.Executer{command2.NewEdit("")}
	}

	return executers
}
