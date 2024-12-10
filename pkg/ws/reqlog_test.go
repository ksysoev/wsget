package ws

import (
	"bytes"
	"io"
	"net/http"
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
