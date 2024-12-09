package ws

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRequestLogger(t *testing.T) {
	tests := []struct {
		name                  string
		verbose               bool
		skipSSLVerification   bool
		expectedSkipSSLVerify bool
	}{
		{
			name:                  "VerboseTrue_SSLVerifyTrue",
			verbose:               true,
			skipSSLVerification:   true,
			expectedSkipSSLVerify: true,
		},
		{
			name:                  "VerboseFalse_SSLVerifyTrue",
			verbose:               false,
			skipSSLVerification:   true,
			expectedSkipSSLVerify: true,
		},
		{
			name:                  "VerboseTrue_SSLVerifyFalse",
			verbose:               true,
			skipSSLVerification:   false,
			expectedSkipSSLVerify: false,
		},
		{
			name:                  "VerboseFalse_SSLVerifyFalse",
			verbose:               false,
			skipSSLVerification:   false,
			expectedSkipSSLVerify: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := newRequestLogger(tt.verbose, tt.skipSSLVerification)

			assert.NotNil(t, rl)
			assert.Equal(t, tt.verbose, rl.verbose)
			assert.NotNil(t, rl.transport)
			assert.IsType(t, &http.Transport{}, rl.transport)
			assert.NotNil(t, rl.transport.TLSClientConfig)
			assert.Equal(t, tt.expectedSkipSSLVerify, rl.transport.TLSClientConfig.InsecureSkipVerify)
		})
	}
}
