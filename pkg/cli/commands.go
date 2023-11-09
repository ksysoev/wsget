package cli

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
)

type ExecutionContext struct {
	input      <-chan keyboard.KeyEvent
	output     io.Writer
	editor     *Editor
	wsConn     *ws.Connection
	outputFile io.Writer
	formater   *formater.Formater
}

type Executer interface {
	Execute(*ExecutionContext) (Executer, error)
}

type CommandEdit struct {
	content string
}

func CommandFactory(raw string) (Executer, error) {
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
		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
}

func NewCommandEdit(content string) *CommandEdit {
	return &CommandEdit{content}
}

func (c *CommandEdit) Execute(exCtx *ExecutionContext) (Executer, error) {
	color.New(color.FgGreen).Fprint(exCtx.output, "->\n")

	fmt.Fprint(exCtx.output, ShowCursor)
	req, err := exCtx.editor.EditRequest(exCtx.input, c.content)
	fmt.Fprint(exCtx.output, LineUp+LineClear)
	fmt.Fprint(exCtx.output, HideCursor)

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

func (c *CommandSend) Execute(exCtx *ExecutionContext) (Executer, error) {
	msg, err := exCtx.wsConn.Send(c.request)
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

func (c *CommandPrintMsg) Execute(exCtx *ExecutionContext) (Executer, error) {
	msg := c.msg
	output, err := exCtx.formater.FormatMessage(msg)

	if err != nil {
		return nil, fmt.Errorf("fail to format for output file: %s, data: %q", err, msg.Data)
	}

	switch msg.Type {
	case ws.Request:
		color.New(color.FgGreen).Fprint(exCtx.output, "->\n")
	case ws.Response:
		color.New(color.FgRed).Fprint(exCtx.output, "<-\n")
	default:
		return nil, fmt.Errorf("unknown message type: %s, data: %q", msg.Type, msg.Data)
	}

	fmt.Fprintf(exCtx.output, "%s\n", output)

	if exCtx.outputFile != nil {
		output, err := exCtx.formater.FormatForFile(msg)
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

func (c *CommandExit) Execute(_ *ExecutionContext) (Executer, error) {
	return nil, fmt.Errorf("interrupted")
}

type CommandWaitForResp struct {
	timeout time.Duration
}

func NewCommandWaitForResp(timeout time.Duration) *CommandWaitForResp {
	return &CommandWaitForResp{timeout}
}

func (c *CommandWaitForResp) Execute(exCtx *ExecutionContext) (Executer, error) {
	if c.timeout.Seconds() == 0 {
		msg, ok := <-exCtx.wsConn.Messages
		if !ok {
			return nil, fmt.Errorf("connection closed")
		}

		return NewCommandPrintMsg(msg), nil
	}

	select {
	case <-time.After(c.timeout):
		return nil, fmt.Errorf("timeout")
	case msg, ok := <-exCtx.wsConn.Messages:
		if !ok {
			return nil, fmt.Errorf("connection closed")
		}

		return NewCommandPrintMsg(msg), nil
	}
}
