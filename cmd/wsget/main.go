package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/ksysoev/wsget/pkg/cli"
	"github.com/ksysoev/wsget/pkg/clierrors"
	"github.com/ksysoev/wsget/pkg/command"
	"github.com/ksysoev/wsget/pkg/ws"
	"github.com/spf13/cobra"
)

var insecure bool
var request string
var outputFile string
var inputFile string
var headers []string
var waitResponse int
var Version = "dev"

const (
	LongDescription = `A command-line tool for interacting with WebSocket servers.

The tool have severl modes of operation:

1. Request mode. The tool will start in interactive mode if no request is provided:

- You can type resquest and press Enter to send it to the server. 
- Request editor allows to input multiline request. the last sybmol of line should be \(backslash) to indicate that the request is not finished yet.
- You can use Ctrl+U to clear the input.
- You can use Ctrl+C or Ctrl+D to exit the tool.
- You can use Esc to cancel input and return to connection mod.

2. Connection mode. The tool will start in connection mode if request is provided.
In this request mode the tool will send the request to the server and print responses. 

- You can use Enter to switch to request input mode.
- You can use Esc to exit connection
- You can use Ctrl+C or Ctrl+D to exit the tool.
`
)

func main() {
	cmd := &cobra.Command{
		Use:        "wsget url [flags]",
		Short:      "A command-line tool for interacting with WebSocket servers",
		Long:       LongDescription,
		Example:    `wsget wss://ws.postman-echo.com/raw -r "Hello, world!"`,
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"url"},
		Version:    Version,
		Run:        run,
	}

	cmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Skip SSL certificate verification")
	cmd.Flags().StringVarP(&request, "request", "r", "", "WebSocket request that will be sent to the server")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for saving all request and responses")
	cmd.Flags().IntVarP(&waitResponse, "wait-resp", "w", -1, "Timeout for single response in seconds, 0 means no timeout. If this option is set, the tool will exit after receiving the first response")
	cmd.Flags().StringSliceVarP(&headers, "header", "H", []string{}, "HTTP headers to attach to the request")
	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input YAML file with list of requests to send to the server")

	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmnd *cobra.Command, args []string) {
	wsURL := args[0]
	if wsURL == "" {
		_ = cmnd.Help()
		return
	}

	if waitResponse >= 0 && request == "" {
		color.New(color.FgRed).Println("Single response timeout could be used only with request")
		return
	}

	wsConn, err := ws.NewWS(wsURL, ws.Options{SkipSSLVerification: insecure, Headers: headers})
	if err != nil {
		color.New(color.FgRed).Println("Unable to connect to the server: ", err)
		return
	}

	defer wsConn.Close()

	input := cli.NewKeyboard()

	client, err := cli.NewCLI(wsConn, input, os.Stdout)
	if err != nil {
		color.New(color.FgRed).Println("Unable to start CLI: ", err)
		return
	}

	opts := cli.RunOptions{}

	if outputFile != "" {
		if opts.OutputFile, err = os.Create(outputFile); err != nil {
			color.New(color.FgRed).Println("Fail to open output file: ", err)
			return
		}
	}

	switch {
	case request != "":
		opts.Commands = []command.Executer{command.NewSend(request)}

		if waitResponse >= 0 {
			opts.Commands = append(
				opts.Commands,
				command.NewWaitForResp(time.Duration(waitResponse)*time.Second),
				command.NewExit(),
			)
		}
	case inputFile != "":
		opts.Commands = []command.Executer{command.NewInputFileCommand(inputFile)}
	default:
		opts.Commands = []command.Executer{command.NewEdit("")}
	}

	if err = client.Run(opts); err != nil {
		if errors.As(err, &clierrors.Interrupted{}) {
			return
		}

		color.New(color.FgRed).Println(err)
	}
}
