package command

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ksysoev/wsget/pkg/clierrors"
	"github.com/ksysoev/wsget/pkg/core"
	"github.com/ksysoev/wsget/pkg/formater"
	"github.com/ksysoev/wsget/pkg/ws"
)

func TestFactory(t *testing.T) {
	tests := []struct {
		macro   *Macro
		want    core.Executer
		name    string
		raw     string
		wantErr bool
	}{
		{
			name:    "empty command",
			raw:     "",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "exit command",
			raw:     "exit",
			macro:   nil,
			want:    NewExit(),
			wantErr: false,
		},
		{
			name:    "edit command with content",
			raw:     "edit some content",
			macro:   nil,
			want:    NewEdit("some content"),
			wantErr: false,
		},
		{
			name:    "edit command without content",
			raw:     "edit",
			macro:   nil,
			want:    NewEdit(""),
			wantErr: false,
		},
		{
			name:    "send command with request",
			raw:     "send some request",
			macro:   nil,
			want:    NewSend("some request"),
			wantErr: false,
		},
		{
			name:    "send command without request",
			raw:     "send",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "wait command without timeout",
			raw:     "wait",
			macro:   nil,
			want:    NewWaitForResp(time.Duration(0)),
			wantErr: false,
		},
		{
			name:    "wait command with timeout",
			raw:     "wait 5",
			macro:   nil,
			want:    NewWaitForResp(time.Duration(5) * time.Second),
			wantErr: false,
		},
		{
			name:    "wait command with invalid timeout",
			raw:     "wait invalid",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "unknown command",
			raw:     "unknown",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "macro command",
			raw:     "macro",
			macro:   &Macro{},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Factory(tt.raw, tt.macro)

			if (err != nil) != tt.wantErr {
				t.Errorf("Factory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got == nil && !tt.wantErr {
				t.Errorf("Factory() got = %v, want non-nil", got)
				return
			}

			if got != nil && !strings.Contains(fmt.Sprintf("%T", got), fmt.Sprintf("%T", tt.want)) {
				t.Errorf("Factory() got = %T, want %T", got, tt.want)
			}

			if got != nil && tt.want != nil {
				switch gotType := got.(type) {
				case *Edit:
					if got.(*Edit).content != tt.want.(*Edit).content {
						t.Errorf("Factory() type %v,  got = %v, want %v", gotType, got, tt.want)
					}
				case *Send:
					if got.(*Send).request != tt.want.(*Send).request {
						t.Errorf("Factory() type %v, got = %v, want %v", gotType, got, tt.want)
					}
				case *WaitForResp:
					if got.(*WaitForResp).timeout != tt.want.(*WaitForResp).timeout {
						t.Errorf("Factory() type %v, got = %v, want %v", gotType, got, tt.want)
					}
				}
			}
		})
	}
}

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

	if !errors.As(err, &clierrors.Interrupted{}) {
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
