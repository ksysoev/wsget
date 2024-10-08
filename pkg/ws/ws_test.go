package ws

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
	"golang.org/x/net/websocket"
)

func TestNewWS(t *testing.T) {
	server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		var msg string

		_ = websocket.Message.Receive(ws, &msg) // wait for request
		_, _ = ws.Write([]byte(msg))

		time.Sleep(time.Second) // to keep the connection open
	}))
	defer server.Close()

	url := "ws://" + server.Listener.Addr().String()
	ws, err := NewWS(url, Options{})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if ws == nil {
		t.Fatalf("Expected ws connection, but got nil")
	}

	msg, err := ws.Send("Hello, world!")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if msg.Data != "Hello, world!" {
		t.Errorf("Expected message data to be 'Hello, world!', but got %v", msg.Data)
	}
}

func TestNewWSWithInvalidURL(t *testing.T) {
	_, err := NewWS("invalid", Options{})

	if err == nil {
		t.Fatalf("Expected error, but got nil")
	}
}

func TestNewWSFailToConnect(t *testing.T) {
	_, err := NewWS("ws://localhost:12345", Options{})

	if err == nil {
		t.Fatalf("Expected error, but got nil")
	}
}

func TestNewWSDisconnect(t *testing.T) {
	server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		var msg string

		_ = websocket.Message.Receive(ws, &msg) // wait for request
		_, _ = ws.Write([]byte(msg))

		time.Sleep(time.Second) // to keep the connection open
	}))
	defer server.Close()

	url := "ws://" + server.Listener.Addr().String()
	ws, err := NewWS(url, Options{})

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
	server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		headerValue := ws.Request().Header.Get("X-Test")

		if headerValue != "Test" {
			t.Errorf("Expected header value to be 'Test', but got %v", headerValue)
		}

		var msg string

		_ = websocket.Message.Receive(ws, &msg) // wait for request
		_, _ = ws.Write([]byte(msg))

		time.Sleep(time.Second) // to keep the connection open
	}))
	defer server.Close()

	url := "ws://" + server.Listener.Addr().String()
	ws, err := NewWS(url, Options{Headers: []string{"X-Test: Test"}})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ws.Close()
}

func TestNewWSWithInvalidHeaders(t *testing.T) {
	server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		var msg string

		_ = websocket.Message.Receive(ws, &msg) // wait for request
		_, _ = ws.Write([]byte(msg))

		time.Sleep(time.Second) // to keep the connection open
	}))
	defer server.Close()

	url := "ws://" + server.Listener.Addr().String()
	_, err := NewWS(url, Options{Headers: []string{"X-Test"}})

	if err == nil {
		t.Fatalf("Expected error, but got nil")
	}

	if !strings.Contains(err.Error(), "invalid header") {
		t.Errorf("Expected error to contain 'invalid header', but got %v", err)
	}
}
func TestHandleError(t *testing.T) {
	// Create a new Connection instance
	ws := &Connection{}

	// Test with EOF error
	var buf bytes.Buffer
	color.Output = &buf

	ws.handleError(fmt.Errorf("EOF"))

	if !strings.Contains(buf.String(), "Connection closed by the server") {
		t.Errorf("Expected 'Connection closed by the server', but got '%v'", buf.String())
	}

	// Test with other error
	buf.Reset()
	ws.handleError(fmt.Errorf("some error"))

	if !strings.Contains(buf.String(), "Fail read from connection: some error\n") {
		t.Errorf("Expected 'Fail read from connection: some error', but got %v", buf.String())
	}

	// Test with closed connection
	ws.isClosed.Store(true)
	buf.Reset()
	ws.handleError(fmt.Errorf("some error"))

	if buf.String() != "" {
		t.Errorf("Expected empty string, but got %v", buf.String())
	}
}
func TestMessageTypeString(t *testing.T) {
	tests := []struct {
		name string
		want string
		mt   MessageType
	}{
		{
			name: "Request",
			mt:   Request,
			want: "Request",
		},
		{
			name: "Response",
			mt:   Response,
			want: "Response",
		},
		{
			name: "Not defined",
			mt:   MessageType(42),
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
