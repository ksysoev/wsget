package command

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/fatih/color"
	"github.com/ksysoev/wsget/pkg/core"
	"gopkg.in/yaml.v3"
)

const (
	PartsNumber = 2
	LineUp      = "\x1b[1A"
	LineClear   = "\x1b[2K"
	HideCursor  = "\x1b[?25l"
	ShowCursor  = "\x1b[?25h"
)

type Edit struct {
	content string
}

func NewEdit(content string) *Edit {
	return &Edit{content}
}

// Execute executes the edit command and returns a Send command id editing was successful or an error in other case.
func (c *Edit) Execute(exCtx core.ExecutionContext) (core.Executer, error) {
	output := exCtx.Output()
	_, _ = color.New(color.FgGreen).Fprint(output, "->\n")
	_, _ = fmt.Fprint(output, ShowCursor)

	req, err := exCtx.Editor().Edit(context.TODO(), c.content)

	_, _ = fmt.Fprint(output, LineUp+LineClear+HideCursor)

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
// It implements the Execute method of the core.Executer interface.
func (c *Send) Execute(exCtx core.ExecutionContext) (core.Executer, error) {
	err := exCtx.Connection().Send(context.TODO(), c.request)
	if err != nil {
		return nil, err
	}

	return NewPrintMsg(core.Message{Type: core.Request, Data: c.request}), nil
}

type PrintMsg struct {
	msg core.Message
}

func NewPrintMsg(msg core.Message) *PrintMsg {
	return &PrintMsg{msg}
}

// Execute executes the PrintMsg command and returns nil and error.
// It formats the message and prints it to the output file.
// If an output file is provided, it writes the formatted message to the file.
func (c *PrintMsg) Execute(exCtx core.ExecutionContext) (core.Executer, error) {
	msg := c.msg
	output, err := exCtx.Formater().FormatMessage(msg.Type.String(), msg.Data)

	if err != nil {
		return nil, err
	}

	switch msg.Type {
	case core.Request:
		_, _ = color.New(color.FgGreen).Fprintln(exCtx.Output(), "->")
	case core.Response:
		_, _ = color.New(color.FgRed).Fprintln(exCtx.Output(), "<-")
	default:
		return nil, &ErrUnsupportedMessageType{msg.Type.String()}
	}

	_, _ = fmt.Fprintf(exCtx.Output(), "%s\n", output)

	outputFile := exCtx.OutputFile()
	if outputFile != nil && !reflect.ValueOf(outputFile).IsNil() {
		output, err := exCtx.Formater().FormatForFile(msg.Type.String(), msg.Data)
		if err != nil {
			return nil, err
		}

		_, err = fmt.Fprintln(outputFile, output)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

type Exit struct{}

func NewExit() *Exit {
	return &Exit{}
}

// Execute method implements the Execute method of the core.Executer interface.
// It returns an error indicating that the program was interrupted.
func (c *Exit) Execute(_ core.ExecutionContext) (core.Executer, error) {
	return nil, core.ErrInterrupted
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
func (c *WaitForResp) Execute(exCtx core.ExecutionContext) (core.Executer, error) {
	ctx := context.TODO()

	if c.timeout.Seconds() != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), c.timeout)
		defer cancel()
	}

	msg, err := exCtx.WaitForMessage(ctx)
	if err != nil {
		return nil, err
	}

	return NewPrintMsg(msg), nil
}

type CmdEdit struct{}

func NewCmdEdit() *CmdEdit {
	return &CmdEdit{}
}

// Execute executes the CmdEdit and returns a core.Executer and an error.
// It prompts the user to edit a command and returns the corresponding Command object.
func (c *CmdEdit) Execute(exCtx core.ExecutionContext) (core.Executer, error) {
	output := exCtx.Output()

	_, _ = fmt.Fprint(output, ":"+ShowCursor)

	rawCmd, err := exCtx.Editor().CommandMode(context.TODO(), "")

	_, _ = fmt.Fprint(output, LineClear+"\r"+HideCursor)

	if err != nil {
		return nil, err
	}

	cmd, err := exCtx.Factory().Create(rawCmd)

	if err != nil {
		_, _ = color.New(color.FgRed).Fprintln(output, err)
		return nil, nil
	}

	return cmd, nil
}

type Sequence struct {
	subCommands []core.Executer
}

func NewSequence(subCommands []core.Executer) *Sequence {
	return &Sequence{subCommands}
}

// Execute executes the command sequence by iterating over all sub-commands and executing them recursively.
// It takes a core.ExecutionContext as input and returns a core.Executer and an error.
func (c *Sequence) Execute(exCtx core.ExecutionContext) (core.Executer, error) {
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

type InputFileCommand struct {
	filePath string
}

func NewInputFileCommand(filePath string) *InputFileCommand {
	return &InputFileCommand{filePath}
}

// Execute executes the InputFileCommand and returns a core.Executer and an error.
// It reads the file and executes the commands in the file.
func (c *InputFileCommand) Execute(exCtx core.ExecutionContext) (core.Executer, error) {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return nil, err
	}

	var rawCommands []string
	if err := yaml.Unmarshal(data, &rawCommands); err != nil {
		return nil, err
	}

	cmds := make([]core.Executer, 0, len(rawCommands))

	for _, rawCommand := range rawCommands {
		cmd, err := exCtx.Factory().Create(rawCommand)
		if err != nil {
			return nil, err
		}

		cmds = append(cmds, cmd)
	}

	return NewSequence(cmds), nil
}

type RepeatCommand struct {
	subCommand core.Executer
	times      int
}

func NewRepeatCommand(times int, subCommand core.Executer) *RepeatCommand {
	return &RepeatCommand{subCommand, times}
}

// Execute executes the RepeatCommand and returns a core.Executer and an error.
// It executes the sub-command the specified number of times.
func (c *RepeatCommand) Execute(exCtx core.ExecutionContext) (core.Executer, error) {
	for i := 0; i < c.times; i++ {
		cmd := c.subCommand
		for cmd != nil {
			var err error
			if cmd, err = cmd.Execute(exCtx); err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}

type SleepCommand struct {
	duration time.Duration
}

func NewSleepCommand(duration time.Duration) *SleepCommand {
	return &SleepCommand{duration}
}

// Execute executes the SleepCommand and returns a core.Executer and an error.
// It sleeps for the specified duration.
func (c *SleepCommand) Execute(_ core.ExecutionContext) (core.Executer, error) {
	time.Sleep(c.duration)

	return nil, nil
}
