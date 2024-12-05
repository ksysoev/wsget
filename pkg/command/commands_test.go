package command

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/ksysoev/wsget/pkg/core"
	"github.com/ksysoev/wsget/pkg/formater"
	"github.com/ksysoev/wsget/pkg/ws"
)

type mockCommand struct {
	err error
}

func (c *mockCommand) Execute(_ core.ExecutionContext) (core.Executer, error) {
	return nil, c.err
}

func TestSequence_Execute(t *testing.T) {
	tests := []struct {
		exCtx       *mockContext
		name        string
		subCommands []core.Executer
		wantErr     bool
	}{
		{
			name:        "empty command sequence",
			subCommands: []core.Executer{},
			exCtx:       &mockContext{},
			wantErr:     false,
		},
		{
			name: "command sequence with one command",
			subCommands: []core.Executer{
				&mockCommand{},
			},
			exCtx:   &mockContext{},
			wantErr: false,
		},
		{
			name: "command sequence with multiple commands",
			subCommands: []core.Executer{
				&mockCommand{},
				&mockCommand{},
				&mockCommand{},
			},
			exCtx:   &mockContext{},
			wantErr: false,
		},
		{
			name: "command sequence with error",
			subCommands: []core.Executer{
				&mockCommand{},
				&mockCommand{err: fmt.Errorf("error")},
				&mockCommand{},
			},
			exCtx:   &mockContext{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &Sequence{
				subCommands: tt.subCommands,
			}

			_, err := cs.Execute(tt.exCtx)

			if (err != nil) != tt.wantErr {
				t.Errorf("Sequence.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExit_Execute(t *testing.T) {
	c := &Exit{}
	_, err := c.Execute(nil)

	if err == nil {
		t.Errorf("Exit.Execute() error = %v, wantErr %v", err, true)
	}

	if !errors.Is(err, core.ErrInterrupted) {
		t.Errorf("Exit.Execute() error = %v, wantErr interrupted", err)
	}
}

func TestPrintMsg_Execute(t *testing.T) {
	tests := []struct {
		exCtx       *mockContext
		name        string
		expectedOut string
		msg         ws.Message
		wantErr     bool
	}{
		{
			name: "request message",
			msg: ws.Message{
				Type: ws.Request,
				Data: "some request data",
			},
			exCtx:       &mockContext{},
			wantErr:     false,
			expectedOut: "->\n" + "some request data\n",
		},
		{
			name: "response message",
			msg: ws.Message{
				Type: ws.Response,
				Data: "some response data",
			},
			exCtx:       &mockContext{},
			wantErr:     false,
			expectedOut: "<-\n" + "some response data\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &PrintMsg{
				msg: tt.msg,
			}

			exCtx := tt.exCtx

			_, err := c.Execute(exCtx)

			if (err != nil) != tt.wantErr {
				t.Errorf("PrintMsg.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if exCtx.buf.String() != tt.expectedOut {
				t.Errorf("PrintMsg.Execute() output = %q, want %q", exCtx.buf.String(), tt.expectedOut)
			}
		})
	}
}

type mockContext struct {
	requestEditor core.Editor
	buf           bytes.Buffer
}

func (c *mockContext) Input() <-chan core.KeyEvent {
	return make(<-chan core.KeyEvent)
}
func (c *mockContext) Output() io.Writer {
	return &c.buf
}

func (c *mockContext) OutputFile() io.Writer {
	return nil
}

func (c *mockContext) Formater() formater.Formater {
	return formater.NewFormat()
}

func (c *mockContext) RequestEditor() core.Editor {
	return c.requestEditor
}

func (c *mockContext) CmdEditor() core.Editor {
	return c.requestEditor
}

func (c *mockContext) Connection() ws.ConnectionHandler {
	return &ws.Connection{}
}

func (c *mockContext) Factory() core.CommandFactory {
	return nil
}

type mockEditor struct {
	err     error
	content string
}

func (e *mockEditor) Edit(_ <-chan core.KeyEvent, _ string) (string, error) {
	return e.content, e.err
}

func (e *mockEditor) Close() error {
	return nil
}
func TestCmdEdit_Execute(t *testing.T) {
	exCtx := &mockContext{}
	output, _ := exCtx.Output().(*bytes.Buffer)
	input := make(chan core.KeyEvent)
	editor := &mockEditor{}
	editor.content = "edit command"
	exCtx.requestEditor = editor

	c := &CmdEdit{}

	go func() {
		input <- core.KeyEvent{}
		close(input)
	}()

	_, err := c.Execute(exCtx)

	if err != nil {
		t.Errorf("CmdEdit.Execute() error = %v, want nil", err)
	}

	expectedOutput := ":"
	expectedOutput += ShowCursor
	expectedOutput += LineClear + "\r"
	expectedOutput += HideCursor

	if output.String() != expectedOutput {
		t.Errorf("CmdEdit.Execute() output = %q, want %q", output.String(), expectedOutput)
	}
}
