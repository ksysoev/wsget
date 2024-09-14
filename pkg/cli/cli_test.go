package cli

import (
	"errors"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/ksysoev/wsget/pkg/clierrors"
	"github.com/ksysoev/wsget/pkg/command"
	"github.com/ksysoev/wsget/pkg/ws"
	"golang.org/x/net/websocket"
)

type mockInput struct{}

func (m *mockInput) GetKeys() (<-chan keyboard.KeyEvent, error) {
	return make(chan keyboard.KeyEvent), nil
}

func (m *mockInput) Close() {}

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

	output := os.Stdout
	cli, err := NewCLI(wsConn, &mockInput{}, output)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cli.formater == nil {
		t.Error("Expected non-nil formater")
	}

	if cli.wsConn != wsConn {
		t.Error("Expected wsConn to be set")
	}

	if cli.editor == nil {
		t.Error("Expected non-nil editor")
	}

	if _, err = wsConn.Send("Hello, world!"); err != nil {
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

func TestNewCLIRunWithCommands(t *testing.T) {
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

	output := os.Stdout
	cli, err := NewCLI(wsConn, &mockInput{}, output)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	err = cli.Run(RunOptions{Commands: []command.Executer{command.NewExit()}})

	if err == nil {
		t.Fatalf("Expected error, but got nothing")
	}

	if !errors.As(err, &clierrors.Interrupted{}) {
		t.Errorf("Exit.Execute() error = %v, wantErr interupted", err)
	}
}
