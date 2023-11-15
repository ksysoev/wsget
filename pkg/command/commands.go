package command

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
	"github.com/ksysoev/wsget/pkg/formater"
	"github.com/ksysoev/wsget/pkg/ws"
)

const (
	CommandPartsNumber = 2
	LineUp             = "\x1b[1A"
	LineClear          = "\x1b[2K"
	HideCursor         = "\x1b[?25l"
	ShowCursor         = "\x1b[?25h"
)

type ExecutionContext interface {
	Input() <-chan keyboard.KeyEvent
	OutputFile() io.Writer
	Output() io.Writer
	Formater() formater.Formater
	Connection() ws.ConnectionHandler
	RequestEditor() Editor
	CmdEditor() Editor
	Macro() *Macro
}

type Editor interface {
	Edit(keyStream <-chan keyboard.KeyEvent, initBuffer string) (string, error)
	Close() error
}

type Executer interface {
	Execute(ExecutionContext) (Executer, error)
}

// Factory returns an Executer and an error. It takes a string and a Macro pointer as input.
// The string is split into parts and the first part is used to determine which command to execute.
// Depending on the command, different arguments are passed to the corresponding constructor.
// If the command is not recognized, an error is returned.
func Factory(raw string, macro *Macro) (Executer, error) {
	if raw == "" {
		return nil, fmt.Errorf("empty command")
	}

	parts := strings.SplitN(raw, " ", CommandPartsNumber)
	cmd := parts[0]

	switch cmd {
	case "exit":
		return NewExit(), nil
	case "edit":
		content := ""
		if len(parts) > 1 {
			content = parts[1]
		}

		return NewEdit(content), nil
	case "send":
		if len(parts) == 1 {
			return nil, fmt.Errorf("empty request")
		}

		return NewSend(parts[1]), nil
	case "wait":
		timeout := time.Duration(0)

		if len(parts) > 1 {
			sec, err := strconv.Atoi(parts[1])
			if err != nil || sec < 0 {
				return nil, fmt.Errorf("invalid timeout: %s", err)
			}

			timeout = time.Duration(sec) * time.Second
		}

		return NewWaitForResp(timeout), nil
	default:
		if macro != nil {
			return macro.Get(cmd)
		}

		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
}

type Edit struct {
	content string
}

func NewEdit(content string) *Edit {
	return &Edit{content}
}

// Execute executes the edit command and returns a Send command id editing was successful or an error in other case.
func (c *Edit) Execute(exCtx ExecutionContext) (Executer, error) {
	output := exCtx.Output()
	color.New(color.FgGreen).Fprint(output, "->\n")
	fmt.Fprint(output, ShowCursor)

	req, err := exCtx.RequestEditor().Edit(exCtx.Input(), c.content)

	fmt.Fprint(output, LineUp+LineClear)
	fmt.Fprint(output, HideCursor)

	if err != nil || req == "" {
		return nil, err
	}

	return NewSend(req), nil
}

type Send struct {
	request string
}

func NewSend(request string) *Send {
	return &Send{request}
}

// Execute sends the request using the WebSocket connection and returns a PrintMsg to print the response message.
// It implements the Execute method of the Executer interface.
func (c *Send) Execute(exCtx ExecutionContext) (Executer, error) {
	msg, err := exCtx.Connection().Send(c.request)
	if err != nil {
		return nil, fmt.Errorf("fail to send request: %s", err)
	}

	return NewPrintMsg(*msg), nil
}

type PrintMsg struct {
	msg ws.Message
}

func NewPrintMsg(msg ws.Message) *PrintMsg {
	return &PrintMsg{msg}
}

// Execute executes the PrintMsg command and returns nil and error.
// It formats the message and prints it to the output file.
// If an output file is provided, it writes the formatted message to the file.
func (c *PrintMsg) Execute(exCtx ExecutionContext) (Executer, error) {
	msg := c.msg
	output, err := exCtx.Formater().FormatMessage(msg)

	if err != nil {
		return nil, fmt.Errorf("fail to format for output file: %s, data: %q", err, msg.Data)
	}

	switch msg.Type {
	case ws.Request:
		color.New(color.FgGreen).Fprint(exCtx.Output(), "->\n")
	case ws.Response:
		color.New(color.FgRed).Fprint(exCtx.Output(), "<-\n")
	default:
		return nil, fmt.Errorf("unknown message type: %s, data: %q", msg.Type, msg.Data)
	}

	fmt.Fprintf(exCtx.Output(), "%s\n", output)

	if exCtx.OutputFile() != nil {
		output, err := exCtx.Formater().FormatForFile(msg)
		if err != nil {
			return nil, fmt.Errorf("fail to write to output file: %s", err)
		}

		fmt.Fprintln(exCtx.OutputFile(), output)
	}

	return nil, nil
}

type Exit struct{}

func NewExit() *Exit {
	return &Exit{}
}

// Execute method implements the Execute method of the Executer interface.
// It returns an error indicating that the program was interrupted.
func (c *Exit) Execute(_ ExecutionContext) (Executer, error) {
	return nil, fmt.Errorf("interrupted")
}

type WaitForResp struct {
	timeout time.Duration
}

func NewWaitForResp(timeout time.Duration) *WaitForResp {
	return &WaitForResp{timeout}
}

// Execute executes the WaitForResp command and waits for a response from the WebSocket connection.
// If a timeout is set, it will return an error if no response is received within the specified time.
// If a response is received, it will return a new PrintMsg command with the received message.
// If the WebSocket connection is closed, it will return an error.
func (c *WaitForResp) Execute(exCtx ExecutionContext) (Executer, error) {
	if c.timeout.Seconds() == 0 {
		msg, ok := <-exCtx.Connection().Messages()
		if !ok {
			return nil, fmt.Errorf("connection closed")
		}

		return NewPrintMsg(msg), nil
	}

	select {
	case <-time.After(c.timeout):
		return nil, fmt.Errorf("timeout")
	case msg, ok := <-exCtx.Connection().Messages():
		if !ok {
			return nil, fmt.Errorf("connection closed")
		}

		return NewPrintMsg(msg), nil
	}
}

type CmdEdit struct{}

func NewCmdEdit() *CmdEdit {
	return &CmdEdit{}
}

// Execute executes the CmdEdit and returns an Executer and an error.
// It prompts the user to edit a command and returns the corresponding Command object.
func (c *CmdEdit) Execute(exCtx ExecutionContext) (Executer, error) {
	output := exCtx.Output()

	fmt.Fprint(output, ":")
	fmt.Fprint(output, ShowCursor)

	rawCmd, err := exCtx.RequestEditor().Edit(exCtx.Input(), "")

	fmt.Fprint(output, LineClear+"\r")
	fmt.Fprint(output, HideCursor)

	if err != nil {
		return nil, err
	}

	cmd, err := Factory(rawCmd, exCtx.Macro())

	if err != nil {
		color.New(color.FgRed).Fprintln(output, err)
		return nil, nil
	}

	return cmd, nil
}

type Sequence struct {
	subCommands []Executer
}

func NewSequence(subCommands []Executer) *Sequence {
	return &Sequence{subCommands}
}

// Execute executes the command sequence by iterating over all sub-commands and executing them recursively.
// It takes an ExecutionContext as input and returns an Executer and an error.
func (c *Sequence) Execute(exCtx ExecutionContext) (Executer, error) {
	for _, cmd := range c.subCommands {
		for cmd != nil {
			var err error
			if cmd, err = cmd.Execute(exCtx); err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}