package ws

import (
	"net/http/httptest"
	"testing"
	"time"

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

	if err = ws.Send("Hello, world!"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	select {
	case msg := <-ws.Messages:
		if msg.Data != "Hello, world!" {
			t.Errorf("Expected message data to be 'Hello, world!', but got %v", msg.Data)
		}
	default:
		t.Errorf("Expected message, but got none")
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
	case _, ok := <-ws.Messages:
		if ok {
			t.Errorf("Expected channel to be closed")
		}
	case <-time.After(time.Millisecond * 10):
		t.Errorf("Expected channel to be closed")
	}
}
