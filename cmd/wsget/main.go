package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ksysoev/wsget/pkg/cli"
	"github.com/ksysoev/wsget/pkg/ws"
	"github.com/spf13/cobra"
)

var insecure bool
var request string
var outputFile string

const (
	LongDescription = `A command-line tool for interacting with WebSocket servers.

The tool have severl modes of operation:

1. Request mode. The tool will start in interactive mode if no request is provided:

- You can type resquest and press Ctrl+S to send it to the server. 
- It supports multiline input.
- You can use Ctrl+U to clear the input.
- You can use Ctrl+C or Ctrl+D to exit the tool.
- You can use Esc to exit Request mode and switch to connection mode.

2. Connection mode. The tool will start in connection mode if request is provided.
In this request mode the tool will send the request to the server and print responses. 

- You can use Ctrl+C or Ctrl+D to exit the tool.
- You can use Esc to switch to Request mode.
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
		Version:    "0.1.4",
		Run:        run,
	}

	cmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Skip SSL certificate verification")
	cmd.Flags().StringVarP(&request, "request", "r", "", "WebSocket request that will be sent to the server")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for saving all request and responses")

	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	wsURL := args[0]
	if wsURL == "" {
		_ = cmd.Help()

		os.Exit(1)
	}

	wsInsp, err := ws.NewWS(wsURL, ws.Options{SkipSSLVerification: insecure})
	if err != nil {
		log.Fatal(err)
	}

	defer wsInsp.Close()

	input := cli.NewKeyboard()

	client := cli.NewCLI(wsInsp, input, os.Stdout)

	opts := cli.RunOptions{StartEditor: true}

	if request != "" {
		opts.StartEditor = false

		go func() {
			err = wsInsp.Send(request)
			if err != nil {
				fmt.Println("Fail to send request:", err)
			}
		}()
	}

	if outputFile != "" {
		if opts.OutputFile, err = os.Create(outputFile); err != nil {
			log.Println(err)
			return
		}
	}

	if err = client.Run(opts); err != nil {
		log.Println("Error:", err)
	}
}
