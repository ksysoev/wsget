package core

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ksysoev/wsget/pkg/formater"
	"github.com/ksysoev/wsget/pkg/ws"
)

const (
	CommandsLimit = 100

	HideCursor = "\x1b[?25l"
	ShowCursor = "\x1b[?25h"
)

var (
	ErrInterrupted = fmt.Errorf("interrupted")
)

type CLI struct {
	formater    *formater.Format
	wsConn      ws.ConnectionHandler
	editor      Editor
	inputStream chan KeyEvent
	output      io.Writer
	commands    chan Executer
	cmdFactory  CommandFactory
}

type RunOptions struct {
	OutputFile *os.File
	Commands   []Executer
}

type Formater interface {
	FormatMessage(wsMsg ws.Message) (string, error)
	FormatForFile(wsMsg ws.Message) (string, error)
}

type CommandFactory interface {
	Create(raw string) (Executer, error)
}

type ExecutionContext interface {
	Input() <-chan KeyEvent
	OutputFile() io.Writer
	Output() io.Writer
	Formater() formater.Formater
	Connection() ws.ConnectionHandler
	Editor() Editor
	Factory() CommandFactory
}

type Editor interface {
	Edit(keyStream <-chan KeyEvent, initBuffer string) (string, error)
	CommandMode(keyStream <-chan KeyEvent, initBuffer string) (string, error)
}

type Executer interface {
	Execute(ExecutionContext) (Executer, error)
}

// NewCLI creates a new CLI instance with the given wsConn, input, and output.
// It returns an error if it fails to get the current user, create the necessary directories,
// load the macro for the domain, or initialize the CLI instance.
func NewCLI(cmdFactory CommandFactory, wsConn ws.ConnectionHandler, output io.Writer, editor Editor) (*CLI, error) {
	return &CLI{
		formater:    formater.NewFormat(),
		editor:      editor,
		wsConn:      wsConn,
		inputStream: make(chan KeyEvent),
		output:      output,
		commands:    make(chan Executer, CommandsLimit),
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
	}()

	c.hideCursor()

	fmt.Fprintln(c.output, "Use Enter to input request and send it, Ctrl+C to exit")

	for _, cmd := range opts.Commands {
		c.commands <- cmd
	}

	exCtx := newExecutionContext(c, opts.OutputFile)

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
