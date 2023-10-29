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

	HideCursor = "\x1b[?25l"
	ShowCursor = "\x1b[?25h"

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
		c.showCursor()
		err := c.editor.History.SaveToFile()

		if err != nil {
			color.New(color.FgRed).Fprint(c.output, "Fail to save history:", err)
		}
	}()

	c.hideCursor()

	keysEvents, err := c.input.GetKeys()
	if err != nil {
		return err
	}
	defer c.input.Close()

	fmt.Fprintln(c.output, "Use Enter to input request and send it, Ctrl+C to exit")

	if opts.StartEditor {
		if err := c.RequestMod(keysEvents); err != nil {
			switch err.Error() {
			case "interrupted":
				return nil
			case "empty request":
			default:
				return err
			}
		}
	}

	for {
		select {
		case event := <-keysEvents:
			switch event.Key {
			case keyboard.KeyEsc, keyboard.KeyCtrlC, keyboard.KeyCtrlD:
				return nil

			case keyboard.KeyEnter:
				if err := c.RequestMod(keysEvents); err != nil {
					switch err.Error() {
					case "interrupted":
						return nil
					case "empty request":
						continue
					default:
						return err
					}
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
	color.New(color.FgGreen).Fprint(c.output, "->\n")

	c.showCursor()
	req, err := c.editor.EditRequest(keysEvents, "")
	fmt.Fprint(c.output, LineUp+LineClear)
	c.hideCursor()

	if err != nil {
		return err
	}

	if req != "" {
		err = c.wsConn.Send(req)
		if err != nil {
			return fmt.Errorf("fail to send request: %s", err)
		}
	}

	return nil
}

func (c *CLI) hideCursor() {
	fmt.Fprint(c.output, HideCursor)
}

func (c *CLI) showCursor() {
	fmt.Fprint(c.output, ShowCursor)
}
