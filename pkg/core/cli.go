package core

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/fatih/color"
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
	formater    Formater
	wsConn      ConnectionHandler
	editor      Editor
	inputStream chan KeyEvent
	messages    chan Message
	output      io.Writer
	commands    chan Executer
	cmdFactory  CommandFactory
}

type RunOptions struct {
	OutputFile io.Writer
	Commands   []Executer
}

type Formater interface {
	FormatMessage(msgType string, msgData string) (string, error)
	FormatForFile(msgType string, msgData string) (string, error)
}

type CommandFactory interface {
	Create(raw string) (Executer, error)
}

type ExecutionContext interface {
	Print(data string, attr ...color.Attribute) error
	PrintToFile(data string) error
	FormatMessage(msg Message, noColor bool) (string, error)
	SendRequest(req string) error
	WaitForResponse(timeout time.Duration) (Message, error)
	EditorMode(initBuffer string) (string, error)
	CommandMode(initBuffer string) (string, error)
	CreateCommand(raw string) (Executer, error)
}

type Editor interface {
	Edit(ctx context.Context, initBuffer string) (string, error)
	CommandMode(ctx context.Context, initBuffer string) (string, error)
	SetInput(input <-chan KeyEvent)
}

type Executer interface {
	Execute(ExecutionContext) (Executer, error)
}

type ConnectionHandler interface {
	SetOnMessage(func(context.Context, []byte))
	Send(ctx context.Context, msg string) error
}

// NewCLI creates a new CLI instance with the given wsConn, input, and output.
// It returns an error if it fails to get the current user, create the necessary directories,
// load the macro for the domain, or initialize the CLI instance.
func NewCLI(cmdFactory CommandFactory, wsConn ConnectionHandler, output io.Writer, editor Editor, formater Formater) *CLI {
	c := &CLI{
		formater:    formater,
		editor:      editor,
		wsConn:      wsConn,
		inputStream: make(chan KeyEvent),
		messages:    make(chan Message),
		output:      output,
		commands:    make(chan Executer, CommandsLimit),
		cmdFactory:  cmdFactory,
	}

	wsConn.SetOnMessage(func(ctx context.Context, msg []byte) {
		c.onMessage(ctx, Message{
			Data: string(msg),
			Type: Response,
		})
	})

	editor.SetInput(c.inputStream)

	return c
}

func (c *CLI) OnKeyEvent(event KeyEvent) {
	c.inputStream <- event
}

func (c *CLI) onMessage(ctx context.Context, msg Message) {
	select {
	case c.messages <- msg:
	case <-ctx.Done():
	}
}

// Run runs the CLI with the provided options.
// It listens for user input and executes commands accordingly.
func (c *CLI) Run(ctx context.Context, opts RunOptions) error {
	defer func() {
		c.showCursor()
		close(c.messages)
	}()

	c.hideCursor()

	_, _ = fmt.Fprintln(c.output, "Use Enter to input request and send it, Ctrl+C to exit")

	for _, cmd := range opts.Commands {
		c.commands <- cmd
	}

	exCtx := newExecutionContext(ctx, c, opts.OutputFile)

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

		case msg, ok := <-c.messages:
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
	_, _ = fmt.Fprint(c.output, HideCursor)
}

// showCursor shows the cursor in the terminal output.
func (c *CLI) showCursor() {
	_, _ = fmt.Fprint(c.output, ShowCursor)
}

type MessageType uint8

const (
	Request MessageType = iota
	Response
)

func (mt MessageType) String() string {
	switch mt {
	case Request:
		return "Request"
	case Response:
		return "Response"
	default:
		return "Not defined"
	}
}

type Message struct {
	Data string      `json:"data"`
	Type MessageType `json:"type"`
}
