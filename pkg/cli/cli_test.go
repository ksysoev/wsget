package cli

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ksysoev/wsget/pkg/ws"
	"golang.org/x/net/websocket"
)

func TestNewCLI(t *testing.T) {
	server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		var msg string
		_ = websocket.Message.Receive(ws, &msg) // wait for request
		_, _ = ws.Write([]byte(msg))
		time.Sleep(time.Second) // to keep the connection open
	}))
	defer server.Close()

	url := "ws://" + server.Listener.Addr().String()
	wsConn, err := ws.NewWS(url, ws.Options{})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	cli := NewCLI(wsConn)

	if cli.formater == nil {
		t.Error("Expected non-nil formater")
	}

	if cli.wsConn != wsConn {
		t.Error("Expected wsConn to be set")
	}

	if cli.editor == nil {
		t.Error("Expected non-nil editor")
	}

	if err = wsConn.Send("Hello, world!"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	done := make(chan bool)
	go func() {
		err := cli.Run(RunOptions{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		done <- true
	}()

	wsConn.Close()

	select {
	case <-done:
	case <-time.After(time.Millisecond * 10):
		t.Error("Expected cli to stop")
	}
}
