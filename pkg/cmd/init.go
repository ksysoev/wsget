package cmd

import (
	"cmp"
	"os"

	"github.com/ksysoev/wsget/pkg/ws"
	"github.com/spf13/cobra"
)

const (
	longDescription = `A command-line tool for interacting with WebSocket servers.

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

type flags struct {
	request      string
	outputFile   string
	inputFile    string
	headers      []string
	maxMsgSize   int64
	waitResponse int
	insecure     bool
	verbose      bool
	configDir    string
}

// InitCommands initializes and returns a new cobra.Command for the wsget tool.
// It takes a version string which sets the version of the command.
// It returns a pointer to a cobra.Command configured with various flags for interacting with WebSocket servers.
func InitCommands(version string) *cobra.Command {
	args := &flags{}

	cmd := &cobra.Command{
		Use:        "wsget url [flags]",
		Short:      "A command-line tool for interacting with WebSocket servers",
		Long:       longDescription,
		Example:    `wsget wss://ws.postman-echo.com/raw -r "Hello, world!"`,
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"url"},
		Version:    version,
		RunE:       createConnectRunner(args),
	}

	cmd.Flags().StringVarP(&args.configDir, "config-dir", "c", "", "Configuration directory for storing history and macros")
	cmd.Flags().BoolVarP(&args.insecure, "insecure", "k", false, "Skip SSL certificate verification")
	cmd.Flags().StringVarP(&args.request, "request", "r", "", "WebSocket request that will be sent to the server")
	cmd.Flags().StringVarP(&args.outputFile, "output", "o", "", "Output file for saving all request and responses")
	cmd.Flags().IntVarP(&args.waitResponse, "wait-resp", "w", -1, "Timeout for single response in seconds, 0 means no timeout. If this option is set, the tool will exit after receiving the first response")
	cmd.Flags().StringSliceVarP(&args.headers, "header", "H", []string{}, "HTTP headers to attach to the request")
	cmd.Flags().StringVarP(&args.inputFile, "input", "i", "", "Input YAML file with list of requests to send to the server")
	cmd.Flags().BoolVarP(&args.verbose, "verbose", "v", false, "Verbose output")
	cmd.Flags().Int64VarP(&args.maxMsgSize, "max-size", "s", ws.DefaultMaxMessageSize, "Maximum message size in bytes, non-positive value will be ignored and default value will be used")

	args.configDir = cmp.Or(args.configDir, os.Getenv("WSGET_CONFIG_DIR"))

	return cmd
}
