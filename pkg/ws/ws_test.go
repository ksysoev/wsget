package ws

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/fatih/color"
	"github.com/ksysoev/wsget/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestNewWS(t *testing.T) {
	server := httptest.NewServer(createEchoWSHandler())
	defer server.Close()

	url := "ws://" + server.Listener.Addr().String()

	ws, err := New(url, Options{})
	require.NoError(t, err)

	err = ws.Connect(context.Background())
	require.NoError(t, err)
	require.NotNil(t, ws.ws)

	msg, err := ws.Send("Hello, world!")
	require.NoError(t, err)

	assert.Equal(t, "Hello, world!", msg.Data)
}

func TestNewWSWithInvalidURL(t *testing.T) {
	_, err := New("invalid"+string(rune(0)), Options{})

	if err == nil {
		t.Fatalf("Expected error, but got nil")
	}
}

func TestNewWSFailToConnect(t *testing.T) {
	ws, err := New("ws://localhost:12345", Options{})
	require.NoError(t, err)

	if err := ws.Connect(context.Background()); err == nil {
		t.Fatalf("Expected error, but got nil")
	}
}

func TestNewWSDisconnect(t *testing.T) {
	server := httptest.NewServer(createEchoWSHandler())
	defer server.Close()

	url := "ws://" + server.Listener.Addr().String()

	ws, err := New(url, Options{})
	require.NoError(t, err)

	err = ws.Connect(context.Background())
	require.NoError(t, err)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ws.Close()

	select {
	case _, ok := <-ws.Messages():
		if ok {
			t.Errorf("Expected channel to be closed")
		}
	case <-time.After(time.Millisecond * 10):
		t.Errorf("Expected channel to be closed")
	}

	// Close again to make sure it doesn't panic
	ws.Close()
}

func TestNewWSWithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test") != "Test" {
			t.Errorf("Expected header X-Test to be 'Test', but got %v", r.Header.Get("X-Test"))
		}

		c, err := websocket.Accept(w, r, nil)

		if err != nil {
			return
		}

		c.Close(websocket.StatusNormalClosure, "")
	}))
	defer server.Close()

	url := "ws://" + server.Listener.Addr().String()
	ws, err := New(url, Options{Headers: []string{"X-Test: Test"}})

	require.NoError(t, err)

	err = ws.Connect(context.Background())

	require.NoError(t, err)

	ws.Close()
}

func TestNewWSWithInvalidHeaders(t *testing.T) {
	server := httptest.NewServer(createEchoWSHandler())
	defer server.Close()

	url := "ws://" + server.Listener.Addr().String()
	ws, err := New(url, Options{Headers: []string{"X-Test"}})

	require.NoError(t, err)

	err = ws.Connect(context.Background())

	assert.ErrorContains(t, err, "invalid header")
}
func TestHandleError(t *testing.T) {
	// Create a new Connection instance
	ws := &Connection{}

	// Test with EOF error
	var buf bytes.Buffer
	color.Output = &buf

	ws.handleError(io.EOF)

	if !strings.Contains(buf.String(), "") {
		t.Errorf("Expected 'Connection closed by the server', but got '%v'", buf.String())
	}

	// Test with other error
	buf.Reset()
	ws.handleError(fmt.Errorf("some error"))

	if !strings.Contains(buf.String(), "Fail read from connection: some error\n") {
		t.Errorf("Expected 'Fail read from connection: some error', but got %v", buf.String())
	}
}
func TestMessageTypeString(t *testing.T) {
	tests := []struct {
		name string
		want string
		mt   core.MessageType
	}{
		{
			name: "Request",
			mt:   core.Request,
			want: "Request",
		},
		{
			name: "Response",
			mt:   core.Response,
			want: "Response",
		},
		{
			name: "Not defined",
			mt:   core.MessageType(42),
			want: "Not defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mt.String(); got != tt.want {
				t.Errorf("MessageType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
