package ws

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
)

func createEchoWSHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}

		defer c.Close(websocket.StatusNormalClosure, "")

		for {
			_, wsr, err := c.Reader(r.Context())
			if err != nil {
				if err == io.EOF {
					return
				}

				return
			}

			wsw, err := c.Writer(r.Context(), websocket.MessageText)
			if err != nil {
				return
			}

			if _, err := io.Copy(wsw, wsr); err != nil {
				return
			}

			if err := wsw.Close(); err != nil {
				return
			}
		}
	})
}

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		options   Options
		wantError bool
	}{
		{
			name: "Valid URL without headers",
			url:  "ws://localhost:8080",
			options: Options{
				Headers: []string{},
			},
			wantError: false,
		},
		{
			name: "Valid URL with headers",
			url:  "ws://localhost:8080",
			options: Options{
				Headers: []string{"Authorization: Bearer token"},
			},
			wantError: false,
		},
		{
			name:      "Invalid URL format",
			url:       "invalid_url" + string(rune(0)),
			options:   Options{},
			wantError: true,
		},
		{
			name: "Headers with incorrect format",
			url:  "ws://localhost:8080",
			options: Options{
				Headers: []string{"X-Test"},
			},
			wantError: true,
		},
		{
			name:      "Empty URL",
			url:       "",
			options:   Options{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.url, tt.options)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetOnMessage(t *testing.T) {
	tests := []struct {
		name         string
		initialFunc  func([]byte)
		newFunc      func([]byte)
		expectedFunc func([]byte)
	}{
		{
			name:         "Set new simple function",
			initialFunc:  nil,
			newFunc:      func(data []byte) {},
			expectedFunc: func(data []byte) {},
		},
		{
			name:         "Set nil function",
			initialFunc:  func(data []byte) {},
			newFunc:      nil,
			expectedFunc: nil,
		},
		{
			name: "Replace existing function",
			initialFunc: func(data []byte) {
				fmt.Println("Old")
			},
			newFunc: func(data []byte) {
				fmt.Println("New")
			},
			expectedFunc: func(data []byte) {
				fmt.Println("New")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &Connection{}
			conn.SetOnMessage(tt.initialFunc)
			conn.SetOnMessage(tt.newFunc)

			if tt.expectedFunc == nil {
				assert.Nil(t, conn.onMessage)
			} else {
				assert.NotNil(t, conn.onMessage)
			}
		})
	}
}

func TestConnection_Hostname(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		expectedHost string
	}{
		{
			name:         "Valid URL with port",
			url:          "ws://localhost:8080",
			expectedHost: "localhost",
		},
		{
			name:         "Valid URL without port",
			url:          "ws://example.com",
			expectedHost: "example.com",
		},
		{
			name:         "IPv4 address",
			url:          "ws://127.0.0.1:8080",
			expectedHost: "127.0.0.1",
		},
		{
			name:         "IPv6 address",
			url:          "ws://[::1]:8080",
			expectedHost: "::1",
		},
		{
			name:         "URL with subdomain",
			url:          "ws://api.example.com",
			expectedHost: "api.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.url)
			assert.NoError(t, err)

			conn := &Connection{url: u}
			host := conn.Hostname()

			assert.Equal(t, tt.expectedHost, host)
		})
	}
}

func TestConnection_HandleError(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		isNil bool
	}{
		{
			name:  "Context canceled error",
			err:   context.Canceled,
			isNil: true,
		},
		{
			name:  "IO EOF error",
			err:   io.EOF,
			isNil: true,
		},
		{
			name:  "Net ErrClosed error",
			err:   net.ErrClosed,
			isNil: true,
		},
		{
			name:  "Unexpected error",
			err:   errors.New("unexpected error"),
			isNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &Connection{}
			err := conn.handleError(tt.err)
			if tt.isNil {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "fail read from connection")
			}
		})
	}
}
