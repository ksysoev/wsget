package command

import (
	"fmt"
	"os"
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

// NewEdit creates a new Edit command with the specified content.
// It takes a single parameter, content, of type string, which represents the initial content for editing.
// It returns a pointer to an Edit struct initialized with the provided content.
func NewEdit(content string) *Edit {
	return &Edit{content}
}

// Execute executes the edit command and returns a Send command id editing was successful or an error in other case.
func (c *Edit) Execute(exCtx core.ExecutionContext) (core.Executer, error) {
	if err := exCtx.Print("->\n"+ShowCursor, color.FgGreen); err != nil {
		return nil, err
	}

	req, err := exCtx.EditorMode(c.content)
	if err != nil {
		return nil, err
	}

	if err := exCtx.Print(LineUp + LineClear + HideCursor); err != nil {
		return nil, err
	}

	return NewSend(req), nil
}

type Send struct {
	request string
}

// NewSend creates a new Send command with the provided request string.
// It takes a single parameter request of type string.
// It returns a pointer to a Send instance initialized with the given request.
func NewSend(request string) *Send {
	return &Send{request}
}

// Execute sends the request using the WebSocket connection and returns a PrintMsg to print the response message.
// It implements the Execute method of the core.Executer interface.
func (c *Send) Execute(exCtx core.ExecutionContext) (core.Executer, error) {
	err := exCtx.SendRequest(c.request)
	if err != nil {
		return nil, err
	}

	return NewPrintMsg(core.Message{Type: core.Request, Data: c.request}), nil
}

type PrintMsg struct {
	msg core.Message
}

// NewPrintMsg creates a new PrintMsg instance with the provided core.Message.
// It takes a msg parameter of type core.Message, representing the message to be printed.
// It returns a pointer to a PrintMsg struct initialized with the given message.
func NewPrintMsg(msg core.Message) *PrintMsg {
	return &PrintMsg{msg}
}

// Execute executes the PrintMsg command and returns nil and error.
// It formats the message and prints it to the output file.
// If an output file is provided, it writes the formatted message to the file.
func (c *PrintMsg) Execute(exCtx core.ExecutionContext) (core.Executer, error) {
	output, err := exCtx.FormatMessage(c.msg, false)

	if err != nil {
		return nil, fmt.Errorf("fail to format message: %w", err)
	}

	switch c.msg.Type {
	case core.Request:
		err = exCtx.Print("->", color.FgGreen)
	case core.Response:
		err = exCtx.Print("<-", color.FgRed)
	default:
		return nil, fmt.Errorf("unsupported message type: %s", c.msg.Type.String())
	}

	if err != nil {
		return nil, fmt.Errorf("fail to print message: %w", err)
	}

	if err := exCtx.Print("%s\n" + output); err != nil {
		return nil, fmt.Errorf("fail to print message: %w", err)
	}

	fileOutput, err := exCtx.FormatMessage(c.msg, true)
	if err != nil {
		return nil, fmt.Errorf("fail to format message for file: %w", err)
	}

	if err := exCtx.PrintToFile(fileOutput); err != nil {
		return nil, fmt.Errorf("fail to write to output file: %w", err)
	}

	return nil, nil
}

type Exit struct{}

// NewExit creates and returns a new instance of the Exit command.
// It takes no parameters and returns a pointer to an Exit struct.
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

// NewWaitForResp creates a new WaitForResp command with the specified timeout duration.
// It takes a single parameter timeout of type time.Duration, determining how long to wait for a response.
// It returns a pointer to a WaitForResp instance.
func NewWaitForResp(timeout time.Duration) *WaitForResp {
	return &WaitForResp{timeout}
}

// Execute executes the WaitForResp command and waits for a response from the WebSocket connection.
// If a timeout is set, it will return an error if no response is received within the specified time.
// If a response is received, it will return a new PrintMsg command with the received message.
// If the WebSocket connection is closed, it will return an error.
func (c *WaitForResp) Execute(exCtx core.ExecutionContext) (core.Executer, error) {
	msg, err := exCtx.WaitForResponse(c.timeout)
	if err != nil {
		return nil, err
	}

	return NewPrintMsg(msg), nil
}

type CmdEdit struct{}

// NewCmdEdit initializes and returns a new instance of CmdEdit.
// It does not take any parameters.
// It returns a pointer to CmdEdit, which can execute an edit command.
func NewCmdEdit() *CmdEdit {
	return &CmdEdit{}
}

// Execute executes the CmdEdit and returns a core.Executer and an error.
// It prompts the user to edit a command and returns the corresponding Command object.
func (c *CmdEdit) Execute(exCtx core.ExecutionContext) (core.Executer, error) {
	if err := exCtx.Print(":" + ShowCursor); err != nil {
		return nil, err
	}

	rawCmd, err := exCtx.CommandMode("")
	if err != nil {
		return nil, err
	}

	if err := exCtx.Print(LineClear + "\r" + HideCursor); err != nil {
		return nil, err
	}

	cmd, err := exCtx.CreateCommand(rawCmd)

	if err != nil {
		err := exCtx.Print(fmt.Sprintf("Invalid command: %s\n", rawCmd), color.FgRed)
		return nil, err
	}

	return cmd, nil
}

type Sequence struct {
	subCommands []core.Executer
}

// NewSequence creates a new Sequence containing a list of sub-commands.
// It takes subCommands, a slice of core.Executer, which represents the commands to be executed in order.
// It returns a pointer to a Sequence that will execute the sub-commands sequentially.
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

// NewInputFileCommand creates a new InputFileCommand instance.
// It takes filePath of type string, which specifies the path to the input file.
// It returns a pointer to an InputFileCommand initialized with the given file path.
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
		cmd, err := exCtx.CreateCommand(rawCommand)
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

// NewRepeatCommand creates a new RepeatCommand to execute a sub-command multiple times.
// It takes times of type int, which specifies the number of repetitions, and subCommand of type core.Executer to repeat.
// It returns a pointer to a RepeatCommand initialized with the given subCommand and times.
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

// NewSleepCommand creates a new SleepCommand that pauses execution for a specified duration.
// It takes a duration parameter of type time.Duration.
// It returns a pointer to a SleepCommand instance.
func NewSleepCommand(duration time.Duration) *SleepCommand {
	return &SleepCommand{duration}
}

// Execute executes the SleepCommand and returns a core.Executer and an error.
// It sleeps for the specified duration.
func (c *SleepCommand) Execute(_ core.ExecutionContext) (core.Executer, error) {
	time.Sleep(c.duration)

	return nil, nil
}
