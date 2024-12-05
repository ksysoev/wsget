package core

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
	"github.com/ksysoev/wsget/pkg/formater"
	"github.com/ksysoev/wsget/pkg/ws"
)

const (
	MacroDir           = "macro"
	ConfigDir          = ".wsget"
	HistoryFilename    = ConfigDir + "/history"
	HistoryCmdFilename = ConfigDir + "/cmd_history"
	ConfigDirMode      = 0o755
	CommandsLimit      = 100
	HistoryLimit       = 100

	HideCursor = "\x1b[?25l"
	ShowCursor = "\x1b[?25h"

	Bell = "\a"
)

var (
	ErrInterrupted = fmt.Errorf("interrupted")
)

type CLI struct {
	formater    *formater.Format
	wsConn      ws.ConnectionHandler
	editor      Editor
	cmdEditor   Editor
	inputStream chan KeyEvent
	output      io.Writer
	commands    chan Executer
	cmdFactory  CommandFactory
}

type RunOptions struct {
	OutputFile *os.File
	Commands   []Executer
}

type Inputer interface {
	GetKeys() (<-chan keyboard.KeyEvent, error)
	Close()
}

type ConnectionHandler interface {
	Messages() <-chan ws.Message
	Hostname() string
	Send(msg string) (*ws.Message, error)
	Close()
}

type Formater interface {
	FormatMessage(wsMsg ws.Message) (string, error)
	FormatForFile(wsMsg ws.Message) (string, error)
}

// NewCLI creates a new CLI instance with the given wsConn, input, and output.
// It returns an error if it fails to get the current user, create the necessary directories,
// load the macro for the domain, or initialize the CLI instance.
func NewCLI(cmdFactory CommandFactory, wsConn ws.ConnectionHandler, output io.Writer, edit Editor, cmdEdit Editor) (*CLI, error) {
	commands := make(chan Executer, CommandsLimit)

	return &CLI{
		formater:    formater.NewFormat(),
		editor:      edit,
		cmdEditor:   cmdEdit,
		wsConn:      wsConn,
		inputStream: make(chan KeyEvent),
		output:      output,
		commands:    commands,
		cmdFactory:  cmdFactory,
	}, nil
}

func (c *CLI) OnKeyEvent(event KeyEvent) {
	c.inputStream <- event
}

// Run runs the CLI with the provided options.
// It listens for user input and executes commands accordingly.
func (c *CLI) Run(ctx context.Context, opts RunOptions) error {
	defer func() {
		c.showCursor()
		err := c.editor.Close()

		if err != nil {
			color.New(color.FgRed).Fprint(c.output, "Fail to save history:", err)
		}

		err = c.cmdEditor.Close()
		if err != nil {
			color.New(color.FgRed).Fprint(c.output, "Fail to save history:", err)
		}
	}()

	c.hideCursor()

	fmt.Fprintln(c.output, "Use Enter to input request and send it, Ctrl+C to exit")

	for _, cmd := range opts.Commands {
		c.commands <- cmd
	}

	exCtx := NewExecutionContext(c, opts.OutputFile)

	for {
		select {
		case cmd := <-c.commands:
			var err error
			for cmd != nil {
				cmd, err = cmd.Execute(exCtx)

				if err != nil {
					return err
				}
			}
		case event := <-c.inputStream:
			switch event.Key {
			case KeyEsc, KeyCtrlC, KeyCtrlD:
				cmd, err := c.cmdFactory.Create("exit")
				if err != nil {
					return fmt.Errorf("fail to create exit command: %w", err)
				}

				c.commands <- cmd
			case KeyEnter:
				cmd, err := c.cmdFactory.Create("edit")
				if err != nil {
					return fmt.Errorf("fail to create edit command: %w", err)
				}

				c.commands <- cmd
			default:
				if event.Key > 0 {
					continue
				}

				switch event.Rune {
				case ':':
					cmd, err := c.cmdFactory.Create("editcmd")
					if err != nil {
						return fmt.Errorf("fail to create edit command: %w", err)
					}

					c.commands <- cmd
				default:
					continue
				}
			}

		case msg, ok := <-c.wsConn.Messages():
			if !ok {
				return nil
			}

			cmd, err := c.cmdFactory.Create(fmt.Sprintf("print %s %s", msg.Type.String(), msg.Data))

			if err != nil {
				return fmt.Errorf("fail to create print command: %w", err)
			}

			c.commands <- cmd

		case <-ctx.Done():
			return nil
		}
	}
}

// hideCursor hides the cursor in the terminal output.
func (c *CLI) hideCursor() {
	fmt.Fprint(c.output, HideCursor)
}

// showCursor shows the cursor in the terminal output.
func (c *CLI) showCursor() {
	fmt.Fprint(c.output, ShowCursor)
}
