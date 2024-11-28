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

func runConnectCmd(args *flags) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, unnamedArgs []string) error {
		wsURL := unnamedArgs[0]
		if wsURL == "" {
			return fmt.Errorf("url is required")
		}

		if args.waitResponse >= 0 && args.request == "" {
			return fmt.Errorf("single response timeout could be used only with request")
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

		opts := cli.RunOptions{}

		if args.outputFile != "" {
			if opts.OutputFile, err = os.Create(args.outputFile); err != nil {
				return fmt.Errorf("fail to open output file: %w", err)
			}
		}

		switch {
		case args.request != "":
			opts.Commands = []command.Executer{command.NewSend(args.request)}

			if args.waitResponse >= 0 {
				opts.Commands = append(
					opts.Commands,
					command.NewWaitForResp(time.Duration(args.waitResponse)*time.Second),
					command.NewExit(),
				)
			}
		case args.inputFile != "":
			opts.Commands = []command.Executer{command.NewInputFileCommand(args.inputFile)}
		default:
			opts.Commands = []command.Executer{command.NewEdit("")}
		}

		if err = client.Run(opts); err != nil {
			if errors.As(err, &clierrors.Interrupted{}) {
				return nil
			}

			return fmt.Errorf("failed to run client: %w", err)
		}

		return nil
	}
}
