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
	msgChan := make(chan Message)
	wsConn := NewMockConnectionHandler(t)

	wsConn.EXPECT().Send(mock.Anything).Return(&Message{}, nil)
	wsConn.EXPECT().Messages().Return(msgChan)

	factory := NewMockCommandFactory(t)
	editor := NewMockEditor(t)

	output := os.Stdout
	cli, err := NewCLI(factory, wsConn, output, editor, NewMockFormater(t))

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
		err := cli.Run(context.Background(), RunOptions{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		done <- true
	}()

	close(msgChan)

	select {
	case <-done:
	case <-time.After(time.Millisecond * 10):
		t.Error("Expected cli to stop")
	}
}

func TestNewCLIRunWithCommands(t *testing.T) {
	msgChan := make(chan Message)

	wsConn := NewMockConnectionHandler(t)
	wsConn.EXPECT().Messages().Return(msgChan)

	factory := NewMockCommandFactory(t)
	editor := NewMockEditor(t)
	output := os.Stdout
	cli, err := NewCLI(factory, wsConn, output, editor, NewMockFormater(t))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	cmd := NewMockExecuter(t)
	cmd.EXPECT().Execute(mock.Anything).Return(nil, ErrInterrupted)

	err = cli.Run(context.Background(), RunOptions{Commands: []Executer{cmd}})

	if err == nil {
		t.Fatalf("Expected error, but got nothing")
	}

	if !errors.Is(err, ErrInterrupted) {
		t.Errorf("Exit.Execute() error = %v, wantErr interupted", err)
	}
}
