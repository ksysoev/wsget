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

func main() {
	cmd := &cobra.Command{
		Use:   "wsget [url]",
		Short: "A command-line tool for sending WebSocket requests",
		Args:  cobra.ExactArgs(1),
		Run:   run,
	}

	cmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Skip SSL certificate verification")
	cmd.Flags().StringVarP(&request, "request", "r", "", "WebSocket request")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file")

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
