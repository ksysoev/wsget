package macro

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownload(t *testing.T) {
	const validConfig = `
version: 1
domains: ["example.com"]
macro:
  test: ["exit"]
`

	tests := []struct {
		name        string
		filepath    string
		mockResp    *http.Response
		mockErr     error
		expectedErr string
		url         string
	}{
		{
			name:     "successful download and save",
			filepath: "macro.yaml",
			mockResp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(validConfig)),
			},
		},
		{
			name:        "non-200 status code",
			filepath:    "macro.yaml",
			mockResp:    &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(bytes.NewBufferString(""))},
			expectedErr: "fail to download macro: 404 Not Found",
		},
		{
			name:        "invalid config format",
			filepath:    "macro.yaml",
			mockResp:    &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString("invalid yaml"))},
			expectedErr: "fail to download macro: yaml: unmarshal errors",
		},
		{
			name:        "file creation failure",
			filepath:    "/invalid/path/macro.yaml",
			mockResp:    &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(validConfig))},
			expectedErr: "no such file or directory",
		},
		{
			name:     "fail get request",
			filepath: "macro.yaml",

			expectedErr: "connect: connection refused",
			url:         "http://localhost:1234",
		},
		{
			name:     "fail to parse commands",
			filepath: "macro.yaml",
			mockResp: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`
version: 1
domains: ["example.com"]
macro: 
  test: ["invalid {{ command"]
`)),
			},
			expectedErr: `function "command" not defined`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			httpServ := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.mockResp.StatusCode)
				_, _ = io.Copy(w, tt.mockResp.Body)
			}))

			t.Cleanup(httpServ.Close)

			url := httpServ.URL
			if tt.url != "" {
				url = tt.url
			}

			err := Download(filepath.Join(tmpDir, tt.filepath), url)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
