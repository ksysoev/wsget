package core

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

func TestNewCLI(t *testing.T) {
	wsConn := NewMockConnectionHandler(t)
	wsConn.EXPECT().Send(context.Background(), mock.Anything).Return(nil)
	wsConn.EXPECT().SetOnMessage(mock.Anything)

	factory := NewMockCommandFactory(t)

	editor := NewMockEditor(t)
	editor.EXPECT().SetInput(mock.Anything)

	output := os.Stdout
	cli := NewCLI(factory, wsConn, output, editor, NewMockFormater(t))

	if cli.formater == nil {
		t.Error("Expected non-nil formater")
	}

	if cli.wsConn != wsConn {
		t.Error("Expected wsConn to be set")
	}

	if cli.editor == nil {
		t.Error("Expected non-nil editor")
	}

	if err := wsConn.Send(context.Background(), "Hello, world!"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan bool)

	go func() {
		err := cli.Run(ctx, RunOptions{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		done <- true
	}()

	select {
	case <-done:
	case <-time.After(time.Millisecond * 10):
		t.Error("Expected cli to stop")
	}
}

func TestNewCLIRunWithCommands(t *testing.T) {
	wsConn := NewMockConnectionHandler(t)
	wsConn.EXPECT().SetOnMessage(mock.Anything)

	factory := NewMockCommandFactory(t)

	editor := NewMockEditor(t)
	editor.EXPECT().SetInput(mock.Anything)

	output := os.Stdout
	cli := NewCLI(factory, wsConn, output, editor, NewMockFormater(t))

	cmd := NewMockExecuter(t)
	cmd.EXPECT().Execute(mock.Anything).Return(nil, ErrInterrupted)

	err := cli.Run(context.Background(), RunOptions{Commands: []Executer{cmd}})
	if err == nil {
		t.Fatalf("Expected error, but got nothing")
	}

	if !errors.Is(err, ErrInterrupted) {
		t.Errorf("Exit.Execute() error = %v, wantErr interupted", err)
	}
}

func TestCLI_OnKeyEvent(t *testing.T) {
	wsConn := NewMockConnectionHandler(t)
	wsConn.EXPECT().SetOnMessage(mock.Anything)

	factory := NewMockCommandFactory(t)

	editor := NewMockEditor(t)
	editor.EXPECT().SetInput(mock.Anything)

	output := os.Stdout
	cli := NewCLI(factory, wsConn, output, editor, NewMockFormater(t))

	// Test that OnKeyEvent sends event to inputStream
	event := KeyEvent{Key: KeyEnter, Rune: '\n'}

	// Send event in a goroutine to avoid blocking
	go cli.OnKeyEvent(event)

	// Receive the event from the inputStream with timeout
	select {
	case receivedEvent := <-cli.inputStream:
		if receivedEvent != event {
			t.Errorf("Expected event %v, got %v", event, receivedEvent)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for event")
	}
}

func TestCLI_OnMessage(t *testing.T) {
	wsConn := NewMockConnectionHandler(t)

	var onMessageFunc func(context.Context, []byte)

	wsConn.EXPECT().SetOnMessage(mock.Anything).Run(func(f func(context.Context, []byte)) {
		onMessageFunc = f
	})

	factory := NewMockCommandFactory(t)

	editor := NewMockEditor(t)
	editor.EXPECT().SetInput(mock.Anything)

	output := os.Stdout
	cli := NewCLI(factory, wsConn, output, editor, NewMockFormater(t))

	// Test that onMessage is called and sends message to messages channel
	ctx := context.Background()
	testMsg := []byte("test message")

	// Send message in a goroutine
	go onMessageFunc(ctx, testMsg)

	// Receive the message from the messages channel with timeout
	select {
	case receivedMsg := <-cli.messages:
		if receivedMsg.Data != string(testMsg) {
			t.Errorf("Expected message data %s, got %s", string(testMsg), receivedMsg.Data)
		}

		if receivedMsg.Type != Response {
			t.Errorf("Expected message type Response, got %v", receivedMsg.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for message")
	}
}

func TestCLI_OnMessage_ContextCancelled(t *testing.T) {
	wsConn := NewMockConnectionHandler(t)

	var onMessageFunc func(context.Context, []byte)

	wsConn.EXPECT().SetOnMessage(mock.Anything).Run(func(f func(context.Context, []byte)) {
		onMessageFunc = f
	})

	factory := NewMockCommandFactory(t)

	editor := NewMockEditor(t)
	editor.EXPECT().SetInput(mock.Anything)

	output := os.Stdout
	_ = NewCLI(factory, wsConn, output, editor, NewMockFormater(t))

	// Test that onMessage doesn't block when context is cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	testMsg := []byte("test message")

	// This should not block
	done := make(chan bool)

	go func() {
		onMessageFunc(ctx, testMsg)

		done <- true
	}()

	select {
	case <-done:
		// Success - function returned without blocking
	case <-time.After(100 * time.Millisecond):
		t.Error("onMessage blocked when context was cancelled")
	}
}

func TestMessageType_String(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		msgType  MessageType
	}{
		{
			name:     "Request type",
			msgType:  Request,
			expected: "Request",
		},
		{
			name:     "Response type",
			msgType:  Response,
			expected: "Response",
		},
		{
			name:     "Undefined type",
			msgType:  MessageType(99),
			expected: "Not defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msgType.String()
			if result != tt.expected {
				t.Errorf("MessageType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}
