package ws

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/fatih/color"
)

type requestLogger struct {
	transport *http.Transport
	output    io.Writer
}

// newRequestLogger creates a new request logger with the given verbosity and SSL verification settings.
func newRequestLogger(output io.Writer, skipSSLVerification bool) *requestLogger {
	return &requestLogger{
		transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSSLVerification}, //nolint:gosec // Skip SSL verification
		},
		output: output,
	}
}

// RoundTrip logs the request and response details.
func (rl *requestLogger) RoundTrip(req *http.Request) (*http.Response, error) {
	if rl.output != nil {
		tx := color.New(color.FgGreen)
		tx.SetWriter(rl.output)

		_, _ = fmt.Fprintf(rl.output, "> %s %s %s\n", req.Method, req.URL.String(), req.Proto)
		printHeaders(req.Header, rl.output, ">")
		_, _ = fmt.Fprintln(rl.output)

		tx.UnsetWriter(rl.output)
	}

	resp, err := rl.transport.RoundTrip(req)

	if err != nil {
		return nil, err
	}

	if rl.output != nil {
		rx := color.New(color.FgYellow)
		rx.SetWriter(rl.output)

		_, _ = fmt.Fprintf(rl.output, "< %s %s\n", resp.Proto, resp.Status)
		printHeaders(resp.Header, rl.output, "<")
		_, _ = fmt.Fprintln(rl.output)
		rx.UnsetWriter(rl.output)
	}

	return resp, nil
}

// printHeaders prints the headers to the output with the given prefix.
func printHeaders(headers http.Header, out io.Writer, prefix string) {
	// Sort headers for consistent output
	headerNames := make([]string, 0, len(headers))
	for header := range headers {
		headerNames = append(headerNames, header)
	}

	sort.Strings(headerNames)

	for _, header := range headerNames {
		values := headers[header]
		for _, value := range values {
			_, _ = fmt.Fprintf(out, "%s %s: %s\n", prefix, header, value)
		}
	}
}
