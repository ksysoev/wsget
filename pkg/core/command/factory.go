package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ksysoev/wsget/pkg/core"
)

type MacroRepo interface {
	Get(name, argString string) (core.Executer, error)
}

type Factory struct {
	macro MacroRepo
}

func NewFactory(macro MacroRepo) *Factory {
	return &Factory{macro: macro}
}

func (f *Factory) Create(raw string) (core.Executer, error) {
	if raw == "" {
		return nil, &ErrEmptyCommand{}
	}

	parts := strings.SplitN(raw, " ", PartsNumber)
	cmd := parts[0]

	switch cmd {
	case "exit":
		return NewExit(), nil
	case "edit":
		return f.createEdit(parts)
	case "editcmd":
		return NewCmdEdit(), nil
	case "editbin":
		return NewBinEdit(), nil
	case "send":
		return createSend(parts)
	case "sendbin":
		return createSendBinary(parts)
	case "print":
		return createPrint(raw, parts)
	case "wait":
		return createWait(parts)
	case "repeat":
		return f.createRepeat(raw, parts)
	case "sleep":
		return createSleep(raw, parts)
	case "ping":
		return NewPingCommand(), nil
	default:
		return f.createMacro(cmd, parts)
	}
}

func (f *Factory) createEdit(parts []string) (core.Executer, error) {
	content := ""
	if len(parts) > 1 {
		content = parts[1]
	}

	return NewEdit(content), nil
}

func createSend(parts []string) (core.Executer, error) {
	if len(parts) == 1 {
		return nil, &ErrEmptyRequest{}
	}

	return NewSend(parts[1]), nil
}

func createSendBinary(parts []string) (core.Executer, error) {
	if len(parts) == 1 {
		return nil, &ErrEmptyRequest{}
	}

	return NewSendBinary(parts[1]), nil
}

func createPrint(raw string, parts []string) (core.Executer, error) {
	if len(parts) == 1 {
		return nil, &ErrEmptyRequest{}
	}

	args := strings.SplitN(parts[1], " ", PartsNumber)

	if len(args) < PartsNumber {
		return nil, fmt.Errorf("not enough arguments for print command: %s", raw)
	}

	msgType, err := parseMsgType(args[0])
	if err != nil {
		return nil, err
	}

	return NewPrintMsg(core.Message{Type: msgType, Data: args[1]}), nil
}

func parseMsgType(s string) (core.MessageType, error) {
	switch s {
	case "Request":
		return core.Request, nil
	case "Response":
		return core.Response, nil
	case "RequestBinary":
		return core.RequestBinary, nil
	case "ResponseBinary":
		return core.ResponseBinary, nil
	default:
		return 0, fmt.Errorf("invalid message type: %s", s)
	}
}

func createWait(parts []string) (core.Executer, error) {
	timeout := time.Duration(0)

	if len(parts) > 1 {
		sec, err := strconv.Atoi(parts[1])
		if err != nil || sec < 0 {
			return nil, &ErrInvalidTimeout{parts[1]}
		}

		timeout = time.Duration(sec) * time.Second
	}

	return NewWaitForResp(timeout), nil
}

func (f *Factory) createRepeat(raw string, parts []string) (core.Executer, error) {
	if len(parts) < PartsNumber {
		return nil, fmt.Errorf("not enough arguments for repeat command: %s", raw)
	}

	repeatParts := strings.SplitN(parts[1], " ", PartsNumber)

	times, err := strconv.Atoi(repeatParts[0])
	if err != nil || times <= 0 {
		return nil, fmt.Errorf("invalid repeat times: %s", repeatParts[0])
	}

	subCommand, err := f.Create(repeatParts[1])
	if err != nil {
		return nil, err
	}

	return NewRepeatCommand(times, subCommand), nil
}

func createSleep(raw string, parts []string) (core.Executer, error) {
	if len(parts) < PartsNumber {
		return nil, fmt.Errorf("not enough arguments for sleep command: %s", raw)
	}

	sec, err := strconv.Atoi(parts[1])
	if err != nil || sec < 0 {
		return nil, fmt.Errorf("invalid sleep duration: %s", parts[1])
	}

	return NewSleepCommand(time.Duration(sec) * time.Second), nil
}

func (f *Factory) createMacro(cmd string, parts []string) (core.Executer, error) {
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	if f.macro != nil {
		return f.macro.Get(cmd, args)
	}

	return nil, &ErrUnknownCommand{cmd}
}
