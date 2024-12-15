package ws

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRequestLogger(t *testing.T) {
	tests := []struct {
		output                io.Writer
		name                  string
		skipSSLVerification   bool
		expectedSkipSSLVerify bool
	}{
		{
			name:                  "VerboseTrue_SSLVerifyTrue",
			output:                bytes.NewBuffer(nil),
			skipSSLVerification:   true,
			expectedSkipSSLVerify: true,
		},
		{
			name:                  "VerboseFalse_SSLVerifyTrue",
			output:                nil,
			skipSSLVerification:   true,
			expectedSkipSSLVerify: true,
		},
		{
			name:                  "VerboseTrue_SSLVerifyFalse",
			output:                bytes.NewBuffer(nil),
			skipSSLVerification:   false,
			expectedSkipSSLVerify: false,
		},
		{
			name:                  "VerboseFalse_SSLVerifyFalse",
			output:                nil,
			skipSSLVerification:   false,
			expectedSkipSSLVerify: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := newRequestLogger(tt.output, tt.skipSSLVerification)

			assert.NotNil(t, rl)
			assert.Equal(t, tt.output, rl.output)
			assert.NotNil(t, rl.transport)
			assert.IsType(t, &http.Transport{}, rl.transport)
			assert.NotNil(t, rl.transport.TLSClientConfig)
			assert.Equal(t, tt.expectedSkipSSLVerify, rl.transport.TLSClientConfig.InsecureSkipVerify)
		})
	}
}

func TestPrintHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  http.Header
		prefix   string
		expected string
	}{
		{
			name:     "SingleHeaderSingleValue",
			headers:  http.Header{"Content-Type": {"application/json"}},
			prefix:   "[HEADER]",
			expected: "[HEADER] Content-Type: application/json\n",
		},
		{
			name:     "SingleHeaderMultipleValues",
			headers:  http.Header{"Accept": {"text/plain", "application/json"}},
			prefix:   "[HEADER]",
			expected: "[HEADER] Accept: text/plain\n[HEADER] Accept: application/json\n",
		},
		{
			name:     "MultipleHeaders",
			headers:  http.Header{"Content-Type": {"application/json"}, "Accept": {"application/xml"}},
			prefix:   "[HEADER]",
			expected: "[HEADER] Accept: application/xml\n[HEADER] Content-Type: application/json\n",
		},
		{
			name:     "NoHeaders",
			headers:  http.Header{},
			prefix:   "[HEADER]",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			printHeaders(tt.headers, buf, tt.prefix)

			output := buf.String()

			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestRequestLogger_RoundTrip(t *testing.T) {
	re := regexp.MustCompile(`(?m)^< Date:.*\n`)

	tests := []struct {
		name             string
		output           *bytes.Buffer
		request          *http.Request
		roundTripError   error
		expectedLogLines string
		expectError      bool
	}{
		{
			name:   "Success_NilOutput",
			output: nil,
			request: &http.Request{
				Method: "GET",
				Proto:  "HTTP/1.1",
				Header: http.Header{"User-Agent": {"Test-Agent"}},
			},
			expectedLogLines: "",
			expectError:      false,
		},
		{
			name:   "Success_WithOutput",
			output: new(bytes.Buffer),
			request: &http.Request{
				Method: "GET",
				Proto:  "HTTP/1.1",
				Header: http.Header{"User-Agent": {"Test-Agent"}},
			},
			expectedLogLines: "> GET %%URL%% HTTP/1.1\n> User-Agent: Test-Agent\n\n< HTTP/1.1 200 OK\n< Content-Length: 0\n\n",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			rl := newRequestLogger(tt.output, false)

			cl := http.Client{
				Transport: rl,
			}

			tt.request.URL, _ = url.Parse(s.URL)

			resp, err := cl.Do(tt.request)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				_ = resp.Body.Close()
			}

			if tt.output != nil {
				expected := strings.ReplaceAll(tt.expectedLogLines, "%%URL%%", s.URL)

				output := tt.output.String()

				output = re.ReplaceAllString(output, "")

				assert.Equal(t, expected, output)
			}
		})
	}
}
