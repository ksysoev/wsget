package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ksysoev/wsget/pkg/core"
	"github.com/ksysoev/wsget/pkg/ws"
)

type Factory struct {
	macro *Macro
}

func NewFactory(macro *Macro) *Factory {
	return &Factory{macro: macro}
}

func (f *Factory) Create(raw string) (core.Executer, error) {
	if raw == "" {
		return nil, &ErrEmptyCommand{}
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
	case "editcmd":
		return NewCmdEdit(), nil
	case "send":
		if len(parts) == 1 {
			return nil, &ErrEmptyRequest{}
		}

		return NewSend(parts[1]), nil
	case "print":
		if len(parts) == 1 {
			return nil, &ErrEmptyRequest{}
		}

		args := strings.SplitN(parts[1], " ", CommandPartsNumber)

		if len(args) < CommandPartsNumber {
			return nil, fmt.Errorf("not enough arguments for print command: %s", raw)
		}

		var msgType ws.MessageType

		switch args[0] {
		case "Request":
			msgType = ws.Request
		case "Response":
			msgType = ws.Response
		default:
			return nil, fmt.Errorf("invalid message type: %s", parts[0])
		}

		msg := args[1]

		return NewPrintMsg(ws.Message{Type: msgType, Data: msg}), nil
	case "wait":
		timeout := time.Duration(0)

		if len(parts) > 1 {
			sec, err := strconv.Atoi(parts[1])
			if err != nil || sec < 0 {
				return nil, &ErrInvalidTimeout{parts[1]}
			}

			timeout = time.Duration(sec) * time.Second
		}

		return NewWaitForResp(timeout), nil

	case "repeat":
		if len(parts) < CommandPartsNumber {
			return nil, fmt.Errorf("not enough arguments for repeat command: %s", raw)
		}

		repeatParts := strings.SplitN(parts[1], " ", CommandPartsNumber)

		if len(parts) < CommandPartsNumber {
			return nil, fmt.Errorf("not enough arguments for repeat command: %s", raw)
		}

		times, err := strconv.Atoi(repeatParts[0])
		if err != nil || times <= 0 {
			return nil, fmt.Errorf("invalid repeat times: %s", repeatParts[0])
		}

		subCommand, err := f.Create(repeatParts[1])
		if err != nil {
			return nil, err
		}

		return NewRepeatCommand(times, subCommand), nil

	case "sleep":
		if len(parts) < CommandPartsNumber {
			return nil, fmt.Errorf("not enough arguments for sleep command: %s", raw)
		}

		sec, err := strconv.Atoi(parts[1])
		if err != nil || sec < 0 {
			return nil, fmt.Errorf("invalid sleep duration: %s", parts[1])
		}

		return NewSleepCommand(time.Duration(sec) * time.Second), nil
	default:
		args := ""
		if len(parts) > 1 {
			args = parts[1]
		}

		if f.macro != nil {
			return f.macro.Get(cmd, args)
		}

		return nil, &ErrUnknownCommand{cmd}
	}
}
