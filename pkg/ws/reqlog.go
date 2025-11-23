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

// newRequestLogger creates a new requestLogger for HTTP client request logging.
// It takes an output of type io.Writer for logging and a skipSSLVerification of type bool to control SSL verification.
// It returns a pointer to a requestLogger configured to log requests and responses without SSL verification if specified.
func newRequestLogger(output io.Writer, skipSSLVerification bool) *requestLogger {
	return &requestLogger{
		transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSSLVerification}, //nolint:gosec // Skip SSL verification
		},
		output: output,
	}
}

// RoundTrip executes a single HTTP transaction with logging.
// It takes a parameter req of type *http.Request.
// It returns an *http.Response and an error.
// It returns an error if the underlying transport fails to complete the request.
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

// printHeaders writes HTTP headers to the specified output with each line prefixed by the provided prefix string.
// It takes headers of type http.Header, out of type io.Writer, and prefix of type string.
// It returns no values. The function performs no error handling internally.
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
