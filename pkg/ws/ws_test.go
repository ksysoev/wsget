package ws

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

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
		initialFunc  func(context.Context, []byte)
		newFunc      func(context.Context, []byte)
		expectedFunc func(context.Context, []byte)
		name         string
	}{
		{
			name:         "Set new simple function",
			initialFunc:  nil,
			newFunc:      func(_ context.Context, _ []byte) {},
			expectedFunc: func(_ context.Context, _ []byte) {},
		},
		{
			name:         "Set nil function",
			initialFunc:  func(_ context.Context, _ []byte) {},
			newFunc:      nil,
			expectedFunc: nil,
		},
		{
			name: "Replace existing function",
			initialFunc: func(_ context.Context, _ []byte) {
				fmt.Println("Old")
			},
			newFunc: func(_ context.Context, _ []byte) {
				fmt.Println("New")
			},
			expectedFunc: func(_ context.Context, _ []byte) {
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
		err   error
		name  string
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
		{
			name: "Nolmal Closure error",
			err: websocket.CloseError{
				Code:   websocket.StatusNormalClosure,
				Reason: "normal closure",
			},
			isNil: true,
		},
		{
			name: "Unexpected Close error",
			err: websocket.CloseError{
				Code:   websocket.StatusPolicyViolation,
				Reason: "unexpected close",
			},
			isNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &Connection{}
			err := conn.handleError(tt.err)

			if tt.isNil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestConnection_Connect_Success(t *testing.T) {
	s := httptest.NewServer(createEchoWSHandler())
	defer s.Close()

	conn, err := New("ws://"+s.Listener.Addr().String(), Options{})
	assert.NoError(t, err)

	expectedData := "test data"
	respRecieved := make(chan struct{})

	conn.SetOnMessage(func(_ context.Context, data []byte) {
		assert.Equal(t, expectedData, string(data))
		close(respRecieved)
	})

	wg := &sync.WaitGroup{}
	wg.Add(1)

	defer wg.Wait()
	defer conn.Close()

	go func() {
		defer wg.Done()

		err := conn.Connect(context.Background())
		assert.NoError(t, err)
	}()

	select {
	case <-conn.ready:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for connection")
	}

	err = conn.Send(context.Background(), expectedData)
	assert.NoError(t, err)

	select {
	case <-respRecieved:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for response")
	}
}

func TestConnection_Connect_NoCallback(t *testing.T) {
	conn, err := New("ws://localhost:0", Options{})
	assert.NoError(t, err)

	err = conn.Connect(context.Background())
	assert.Error(t, err)
}

func TestConnection_Connect_AlreadyConnected(t *testing.T) {
	s := httptest.NewServer(createEchoWSHandler())
	defer s.Close()

	conn, err := New("ws://"+s.Listener.Addr().String(), Options{})
	assert.NoError(t, err)

	conn.SetOnMessage(func(context.Context, []byte) {})

	wg := &sync.WaitGroup{}
	wg.Add(1)

	defer wg.Wait()
	defer conn.Close()

	go func() {
		defer wg.Done()

		err = conn.Connect(context.Background())
		assert.NoError(t, err)
	}()

	select {
	case <-conn.ready:
	case <-time.After(1 * time.Second):
	}

	err = conn.Connect(context.Background())
	assert.Error(t, err)
}

func TestConnection_Connect_ContextCancelled(t *testing.T) {
	conn, err := New("ws://localhost:0", Options{})
	assert.NoError(t, err)

	conn.SetOnMessage(func(context.Context, []byte) {})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = conn.Connect(ctx)
	assert.NoError(t, err)
}

func TestConnection_Send_ContextCancelled(t *testing.T) {
	conn, err := New("ws://localhost:0", Options{})
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = conn.Send(ctx, "test data")
	assert.Error(t, err)
}

func TestConnection_Close_NotConnected(t *testing.T) {
	conn, err := New("ws://localhost:0", Options{})
	assert.NoError(t, err)

	err = conn.Close()
	assert.EqualError(t, err, "connection is not established")
}
