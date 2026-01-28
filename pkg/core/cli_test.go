package core

import (
	"context"
	"errors"
	"os"
	"strings"
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

func TestCLI_Run_KeyboardEvents(t *testing.T) {
	tests := []struct {
		name         string
		expectedCmd  string
		keyEvent     KeyEvent
		shouldCreate bool
	}{
		{
			name:         "Enter key triggers edit command",
			keyEvent:     KeyEvent{Key: KeyEnter},
			expectedCmd:  "edit",
			shouldCreate: true,
		},
		{
			name:         "Ctrl+C triggers exit command",
			keyEvent:     KeyEvent{Key: KeyCtrlC},
			expectedCmd:  "exit",
			shouldCreate: true,
		},
		{
			name:         "Ctrl+D triggers exit command",
			keyEvent:     KeyEvent{Key: KeyCtrlD},
			expectedCmd:  "exit",
			shouldCreate: true,
		},
		{
			name:         "Esc triggers exit command",
			keyEvent:     KeyEvent{Key: KeyEsc},
			expectedCmd:  "exit",
			shouldCreate: true,
		},
		{
			name:         "Colon triggers editcmd command",
			keyEvent:     KeyEvent{Rune: ':'},
			expectedCmd:  "editcmd",
			shouldCreate: true,
		},
		{
			name:         "Ctrl+L clears screen",
			keyEvent:     KeyEvent{Key: KeyCtrlL},
			expectedCmd:  "",
			shouldCreate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wsConn := NewMockConnectionHandler(t)
			wsConn.EXPECT().SetOnMessage(mock.Anything)

			factory := NewMockCommandFactory(t)

			if tt.shouldCreate {
				mockCmd := NewMockExecuter(t)
				mockCmd.EXPECT().Execute(mock.Anything).Return(nil, ErrInterrupted)
				factory.EXPECT().Create(tt.expectedCmd).Return(mockCmd, nil)
			}

			editor := NewMockEditor(t)
			editor.EXPECT().SetInput(mock.Anything)

			output := os.Stdout
			cli := NewCLI(factory, wsConn, output, editor, NewMockFormater(t))

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Start CLI.Run in a goroutine
			errChan := make(chan error)

			go func() {
				errChan <- cli.Run(ctx, RunOptions{})
			}()

			// Wait a bit for the Run to start
			time.Sleep(10 * time.Millisecond)

			// Send the key event
			cli.OnKeyEvent(tt.keyEvent)

			// Wait for error or timeout
			select {
			case err := <-errChan:
				if tt.shouldCreate && !errors.Is(err, ErrInterrupted) {
					t.Errorf("Expected ErrInterrupted, got %v", err)
				}
			case <-time.After(100 * time.Millisecond):
				if tt.shouldCreate {
					t.Error("Test timed out waiting for command execution")
				}

				cancel()
			}
		})
	}
}

func TestCLI_Run_MessagesChannel(t *testing.T) {
	wsConn := NewMockConnectionHandler(t)

	var onMessageFunc func(context.Context, []byte)

	wsConn.EXPECT().SetOnMessage(mock.Anything).Run(func(f func(context.Context, []byte)) {
		onMessageFunc = f
	})

	factory := NewMockCommandFactory(t)

	mockCmd := NewMockExecuter(t)
	mockCmd.EXPECT().Execute(mock.Anything).Return(nil, ErrInterrupted)
	factory.EXPECT().Create(mock.MatchedBy(func(s string) bool {
		return strings.HasPrefix(s, "print Response")
	})).Return(mockCmd, nil)

	editor := NewMockEditor(t)
	editor.EXPECT().SetInput(mock.Anything)

	output := os.Stdout
	cli := NewCLI(factory, wsConn, output, editor, NewMockFormater(t))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start CLI.Run in a goroutine
	errChan := make(chan error)

	go func() {
		errChan <- cli.Run(ctx, RunOptions{})
	}()

	// Wait for Run to start
	time.Sleep(10 * time.Millisecond)

	// Send a message through the WebSocket handler
	go onMessageFunc(ctx, []byte("test message"))

	// Wait for error or timeout
	select {
	case err := <-errChan:
		if !errors.Is(err, ErrInterrupted) {
			t.Errorf("Expected ErrInterrupted, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Test timed out waiting for message processing")
		cancel()
	}
}
