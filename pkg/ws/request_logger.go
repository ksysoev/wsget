package ws

import (
	"net/http"
	"sort"

	"github.com/fatih/color"
)

type requestLogger struct {
	transport *http.Transport
	verbose   bool
}

// RoundTrip logs the request and response details.
func (t *requestLogger) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.verbose {
		tx := color.New(color.FgGreen)

		tx.Printf("> %s %s %s\n", req.Method, req.URL.String(), req.Proto)
		printHeaders(req.Header, tx, ">")
		tx.Println()
	}

	resp, err := t.transport.RoundTrip(req)

	if err != nil {
		return nil, err
	}

	if t.verbose {
		rx := color.New(color.FgYellow)

		rx.Printf("< %s %s\n", resp.Proto, resp.Status)
		printHeaders(resp.Header, rx, "<")
		rx.Println()
	}

	return resp, nil
}

// printHeaders prints the headers to the output with the given prefix.
func printHeaders(headers http.Header, out *color.Color, prefix string) {
	// Sort headers for consistent output
	headerNames := make([]string, 0, len(headers))
	for header := range headers {
		headerNames = append(headerNames, header)
	}

	sort.Strings(headerNames)

	for _, header := range headerNames {
		values := headers[header]
		for _, value := range values {
			out.Printf("%s %s: %s\n", prefix, header, value)
		}
	}
}
