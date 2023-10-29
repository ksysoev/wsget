package cli

import (
	"fmt"
	"io"
	"os"
	"os/user"

	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
	"github.com/ksysoev/wsget/pkg/formater"
	"github.com/ksysoev/wsget/pkg/ws"
)

const (
	HistoryFilename = ".wsget_history"
	HistoryLimit    = 100

	MacOSDeleteKey = 127

	Bell = "\a"
)

type CLI struct {
	formater *formater.Formater
	wsConn   *ws.Connection
	editor   *Editor
	input    Inputer
	output   io.Writer
}

type RunOptions struct {
	OutputFile  *os.File
	StartEditor bool
}

type Inputer interface {
	GetKeys() (<-chan keyboard.KeyEvent, error)
	Close()
}

func NewCLI(wsConn *ws.Connection, input Inputer, output io.Writer) (*CLI, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("fail to get current user: %s", err)
	}

	homeDir := currentUser.HomeDir

	history := NewHistory(homeDir+"/"+HistoryFilename, HistoryLimit)

	return &CLI{
		formater: formater.NewFormatter(),
		editor:   NewEditor(output, history),
		wsConn:   wsConn,
		input:    input,
		output:   output,
	}, nil
}

func (c *CLI) Run(opts RunOptions) error {
	defer func() {
		err := c.editor.History.SaveToFile()
		if err != nil {
			fmt.Fprintln(c.output, "Fail to save history:", err)
		}
	}()

	keysEvents, err := c.input.GetKeys()
	if err != nil {
		return err
	}
	defer c.input.Close()

	fmt.Fprintln(c.output, "Use Esc to switch between modes, Ctrl+C to exit")

	if opts.StartEditor {
		if err := c.RequestMod(keysEvents); err != nil {
			if err.Error() == "interrupted" {
				return nil
			}

			return err
		}
	}

	for {
		select {
		case event := <-keysEvents:
			switch event.Key {
			case keyboard.KeyCtrlC, keyboard.KeyCtrlD:
				return nil

			case keyboard.KeyEsc:
				if err := c.RequestMod(keysEvents); err != nil {
					if err.Error() == "interrupted" {
						return nil
					}

					return err
				}

			default:
				continue
			}

		case msg, ok := <-c.wsConn.Messages:
			if !ok {
				return nil
			}

			output, err := c.formater.FormatMessage(msg)
			if err != nil {
				return fmt.Errorf("fail to format for output file: %s, data: %q", err, msg.Data)
			}

			switch msg.Type {
			case ws.Request:
				color.New(color.FgGreen).Fprint(c.output, "->\n")
			case ws.Response:
				color.New(color.FgRed).Fprint(c.output, "<-\n")
			default:
				return fmt.Errorf("unknown message type: %s, data: %q", msg.Type, msg.Data)
			}

			fmt.Fprintf(c.output, "%s\n", output)

			if opts.OutputFile != nil {
				output, err := c.formater.FormatForFile(msg)
				if err != nil {
					return fmt.Errorf("fail to write to output file: %s", err)
				}

				fmt.Fprintln(opts.OutputFile, output)
			}
		}
	}
}

func (c *CLI) RequestMod(keysEvents <-chan keyboard.KeyEvent) error {
	fmt.Fprintln(c.output, "Ctrl+S to send>")

	req, err := c.editor.EditRequest(keysEvents, "")
	if err != nil {
		return err
	}

	if req != "" {
		err = c.wsConn.Send(req)
		if err != nil {
			return fmt.Errorf("fail to send request: %s", err)
		}
	}

	fmt.Fprint(c.output, LineUp+LineClear)

	return nil
}
