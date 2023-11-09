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
	CommandsLimit   = 100
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
	commands chan Executer
}

type RunOptions struct {
	OutputFile *os.File
	Commands   []Executer
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

	commands := make(chan Executer, CommandsLimit)

	return &CLI{
		formater: formater.NewFormatter(),
		editor:   NewEditor(output, history, false),
		wsConn:   wsConn,
		input:    input,
		output:   output,
		commands: commands,
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

	for _, cmd := range opts.Commands {
		c.commands <- cmd
	}

	exCtx := &ExecutionContext{
		input:      keysEvents,
		output:     c.output,
		editor:     c.editor,
		wsConn:     c.wsConn,
		outputFile: opts.OutputFile,
		formater:   c.formater,
	}

	for {
		select {
		case cmd := <-c.commands:
			for cmd != nil {
				cmd, err = cmd.Execute(exCtx)

				if err != nil {
					return err
				}
			}
		case event := <-keysEvents:
			switch event.Key {
			case keyboard.KeyEsc, keyboard.KeyCtrlC, keyboard.KeyCtrlD:
				c.commands <- NewCommandExit()
			case keyboard.KeyEnter:
				c.commands <- NewCommandEdit("")
			default:
				if event.Key > 0 {
					continue
				}

				switch event.Rune {
				case ':':
					cmdEditor := NewEditor(c.output, c.editor.History, true)

					fmt.Fprint(c.output, ":")

					cmd, err := cmdEditor.EditRequest(keysEvents, "")
					if err != nil {
						if err.Error() == ErrInterrupted {
							return nil
						}

						fmt.Fprintln(c.output, err)
					}

					fmt.Fprintln(c.output, cmd)
				default:
					continue
				}
			}

		case msg, ok := <-c.wsConn.Messages:
			if !ok {
				return nil
			}

			c.commands <- NewCommandPrintMsg(msg)
		}
	}
}

func (c *CLI) hideCursor() {
	fmt.Fprint(c.output, HideCursor)
}

func (c *CLI) showCursor() {
	fmt.Fprint(c.output, ShowCursor)
}
