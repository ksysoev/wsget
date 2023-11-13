package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
	"github.com/ksysoev/wsget/pkg/ws"
)

const (
	CommandPartsNumber = 2
)

type ExecutionContext struct {
	input      <-chan keyboard.KeyEvent
	cli        *CLI
	outputFile io.Writer
}

type Executer interface {
	Execute(*ExecutionContext) (Executer, error)
}

// CommandFactory returns an Executer and an error. It takes a string and a Macro pointer as input.
// The string is split into parts and the first part is used to determine which command to execute.
// Depending on the command, different arguments are passed to the corresponding constructor.
// If the command is not recognized, an error is returned.
func CommandFactory(raw string, macro *Macro) (Executer, error) {
	if raw == "" {
		return nil, fmt.Errorf("empty command")
	}

	parts := strings.SplitN(raw, " ", CommandPartsNumber)
	cmd := parts[0]

	switch cmd {
	case "exit":
		return NewCommandExit(), nil
	case "edit":
		content := ""
		if len(parts) > 1 {
			content = parts[1]
		}

		return NewCommandEdit(content), nil
	case "send":
		if len(parts) == 1 {
			return nil, fmt.Errorf("empty request")
		}

		return NewCommandSend(parts[1]), nil
	case "wait":
		timeout := time.Duration(0)

		if len(parts) > 1 {
			sec, err := strconv.Atoi(parts[1])
			if err != nil || sec < 0 {
				return nil, fmt.Errorf("invalid timeout: %s", err)
			}

			timeout = time.Duration(sec) * time.Second
		}

		return NewCommandWaitForResp(timeout), nil
	default:
		if macro != nil {
			return macro.Get(cmd)
		}

		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
}

type CommandEdit struct {
	content string
}

func NewCommandEdit(content string) *CommandEdit {
	return &CommandEdit{content}
}

// Execute executes the edit command and returns a Send command id editing was succesful or an error in other case.
func (c *CommandEdit) Execute(exCtx *ExecutionContext) (Executer, error) {
	color.New(color.FgGreen).Fprint(exCtx.cli.output, "->\n")

	fmt.Fprint(exCtx.cli.output, ShowCursor)
	req, err := exCtx.cli.editor.EditRequest(exCtx.input, c.content)
	fmt.Fprint(exCtx.cli.output, LineUp+LineClear)
	fmt.Fprint(exCtx.cli.output, HideCursor)

	if err != nil || req == "" {
		return nil, err
	}

	return NewCommandSend(req), nil
}

type CommandSend struct {
	request string
}

func NewCommandSend(request string) *CommandSend {
	return &CommandSend{request}
}

// Execute sends the request using the WebSocket connection and returns a CommandPrintMsg to print the response message.
// It implements the Execute method of the Executer interface.
func (c *CommandSend) Execute(exCtx *ExecutionContext) (Executer, error) {
	msg, err := exCtx.cli.wsConn.Send(c.request)
	if err != nil {
		return nil, fmt.Errorf("fail to send request: %s", err)
	}

	return NewCommandPrintMsg(*msg), nil
}

type CommandPrintMsg struct {
	msg ws.Message
}

func NewCommandPrintMsg(msg ws.Message) *CommandPrintMsg {
	return &CommandPrintMsg{msg}
}

// Execute executes the CommandPrintMsg command and returns nil and error.
// It formats the message and prints it to the output file.
// If an output file is provided, it writes the formatted message to the file.
func (c *CommandPrintMsg) Execute(exCtx *ExecutionContext) (Executer, error) {
	msg := c.msg
	output, err := exCtx.cli.formater.FormatMessage(msg)

	if err != nil {
		return nil, fmt.Errorf("fail to format for output file: %s, data: %q", err, msg.Data)
	}

	switch msg.Type {
	case ws.Request:
		color.New(color.FgGreen).Fprint(exCtx.cli.output, "->\n")
	case ws.Response:
		color.New(color.FgRed).Fprint(exCtx.cli.output, "<-\n")
	default:
		return nil, fmt.Errorf("unknown message type: %s, data: %q", msg.Type, msg.Data)
	}

	fmt.Fprintf(exCtx.cli.output, "%s\n", output)

	if exCtx.outputFile != nil {
		output, err := exCtx.cli.formater.FormatForFile(msg)
		if err != nil {
			return nil, fmt.Errorf("fail to write to output file: %s", err)
		}

		fmt.Fprintln(exCtx.outputFile, output)
	}

	return nil, nil
}

type CommandExit struct{}

func NewCommandExit() *CommandExit {
	return &CommandExit{}
}

// Execute method implements the Execute method of the Executer interface.
// It returns an error indicating that the program was interrupted.
func (c *CommandExit) Execute(_ *ExecutionContext) (Executer, error) {
	return nil, fmt.Errorf("interrupted")
}

type CommandWaitForResp struct {
	timeout time.Duration
}

func NewCommandWaitForResp(timeout time.Duration) *CommandWaitForResp {
	return &CommandWaitForResp{timeout}
}

// Execute executes the CommandWaitForResp command and waits for a response from the WebSocket connection.
// If a timeout is set, it will return an error if no response is received within the specified time.
// If a response is received, it will return a new CommandPrintMsg command with the received message.
// If the WebSocket connection is closed, it will return an error.
func (c *CommandWaitForResp) Execute(exCtx *ExecutionContext) (Executer, error) {
	if c.timeout.Seconds() == 0 {
		msg, ok := <-exCtx.cli.wsConn.Messages
		if !ok {
			return nil, fmt.Errorf("connection closed")
		}

		return NewCommandPrintMsg(msg), nil
	}

	select {
	case <-time.After(c.timeout):
		return nil, fmt.Errorf("timeout")
	case msg, ok := <-exCtx.cli.wsConn.Messages:
		if !ok {
			return nil, fmt.Errorf("connection closed")
		}

		return NewCommandPrintMsg(msg), nil
	}
}

type CommandCmdEdit struct{}

func NewCommandCmdEdit() *CommandCmdEdit {
	return &CommandCmdEdit{}
}

// Execute executes the CommandCmdEdit and returns an Executer and an error.
// It prompts the user to edit a command and returns the corresponding Command object.
func (c *CommandCmdEdit) Execute(exCtx *ExecutionContext) (Executer, error) {
	fmt.Fprint(exCtx.cli.output, ":")

	fmt.Fprint(exCtx.cli.output, ShowCursor)
	rawCmd, err := exCtx.cli.cmdEditor.EditRequest(exCtx.input, "")
	fmt.Fprint(exCtx.cli.output, LineClear+"\r")
	fmt.Fprint(exCtx.cli.output, HideCursor)

	if err != nil {
		return nil, err
	}

	cmd, err := CommandFactory(rawCmd, exCtx.cli.macro)

	if err != nil {
		color.New(color.FgRed).Fprintln(exCtx.cli.output, err)
		return nil, nil
	}

	return cmd, nil
}

type CommandSequence struct {
	subCommands []Executer
}

func NewCommandSequence(subCommands []Executer) *CommandSequence {
	return &CommandSequence{subCommands}
}

// Execute executes the command sequence by iterating over all sub-commands and executing them recursively.
// It takes an ExecutionContext as input and returns an Executer and an error.
func (c *CommandSequence) Execute(exCtx *ExecutionContext) (Executer, error) {
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
