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
	MacroDir           = "macro"
	ConfigDir          = ".wsget"
	HistoryFilename    = ConfigDir + "/history"
	HistoryCmdFilename = ConfigDir + "/cmd_history"
	ConfigDirMode      = 0o755
	CommandsLimit      = 100
	HistoryLimit       = 100

	MacOSDeleteKey = 127

	HideCursor = "\x1b[?25l"
	ShowCursor = "\x1b[?25h"

	Bell = "\a"
)

type CLI struct {
	formater  *formater.Formater
	wsConn    ws.ConnectionHandler
	editor    *Editor
	cmdEditor *Editor
	input     Inputer
	output    io.Writer
	commands  chan Executer
	macro     *Macro
}

type RunOptions struct {
	OutputFile *os.File
	Commands   []Executer
}

type Inputer interface {
	GetKeys() (<-chan keyboard.KeyEvent, error)
	Close()
}

// NewCLI creates a new CLI instance with the given wsConn, input, and output.
// It returns an error if it fails to get the current user, create the necessary directories,
// load the macro for the domain, or initialize the CLI instance.
func NewCLI(wsConn ws.ConnectionHandler, input Inputer, output io.Writer) (*CLI, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("fail to get current user: %s", err)
	}

	homeDir := currentUser.HomeDir
	if err = os.MkdirAll(homeDir+"/"+ConfigDir+"/"+MacroDir, ConfigDirMode); err != nil {
		return nil, fmt.Errorf("fail to get current user: %s", err)
	}

	history := NewHistory(homeDir+"/"+HistoryFilename, HistoryLimit)
	cmdHistory := NewHistory(homeDir+"/"+HistoryCmdFilename, HistoryLimit)

	macro, err := LoadMacroForDomain(homeDir+"/"+ConfigDir+"/"+MacroDir, wsConn.Hostname())
	if err != nil {
		return nil, fmt.Errorf("fail to load macro: %s", err)
	}

	commands := make(chan Executer, CommandsLimit)

	return &CLI{
		formater:  formater.NewFormatter(),
		editor:    NewEditor(output, history, false),
		cmdEditor: NewEditor(output, cmdHistory, true),
		wsConn:    wsConn,
		input:     input,
		output:    output,
		commands:  commands,
		macro:     macro,
	}, nil
}

// Run runs the CLI with the provided options.
// It listens for user input and executes commands accordingly.
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
		cli:        c,
		outputFile: opts.OutputFile,
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
					c.commands <- NewCommandCmdEdit()
				default:
					continue
				}
			}

		case msg, ok := <-c.wsConn.Messages():
			if !ok {
				return nil
			}

			c.commands <- NewCommandPrintMsg(msg)
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
