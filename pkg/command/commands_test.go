package command

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/ksysoev/wsget/pkg/formater"
	"github.com/ksysoev/wsget/pkg/ws"
)

func TestFactory(t *testing.T) {
	tests := []struct {
		macro   *Macro
		want    Executer
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

func (c *mockCommand) Execute(_ ExecutionContext) (Executer, error) {
	return nil, c.err
}

func TestSequence_Execute(t *testing.T) {
	tests := []struct {
		exCtx       *mockContext
		name        string
		subCommands []Executer
		wantErr     bool
	}{
		{
			name:        "empty command sequence",
			subCommands: []Executer{},
			exCtx:       &mockContext{},
			wantErr:     false,
		},
		{
			name: "command sequence with one command",
			subCommands: []Executer{
				&mockCommand{},
			},
			exCtx:   &mockContext{},
			wantErr: false,
		},
		{
			name: "command sequence with multiple commands",
			subCommands: []Executer{
				&mockCommand{},
				&mockCommand{},
				&mockCommand{},
			},
			exCtx:   &mockContext{},
			wantErr: false,
		},
		{
			name: "command sequence with error",
			subCommands: []Executer{
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

	if err.Error() != "interrupted" {
		t.Errorf("Exit.Execute() error = %v, wantErr %v", err, "interrupted")
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
	buf bytes.Buffer
}

func (c *mockContext) Input() <-chan keyboard.KeyEvent {
	return make(<-chan keyboard.KeyEvent)
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

func (c *mockContext) RequestEditor() Editor {
	return &mockEditor{}
}

func (c *mockContext) CmdEditor() Editor {
	return &mockEditor{}
}

func (c *mockContext) Connection() ws.ConnectionHandler {
	return &ws.Connection{}
}

func (c *mockContext) Macro() *Macro {
	return nil
}

type mockEditor struct {
	err     error
	content string
}

func (e *mockEditor) Edit(_ <-chan keyboard.KeyEvent, _ string) (string, error) {
	return e.content, e.err
}

func (e *mockEditor) Close() error {
	return nil
}
