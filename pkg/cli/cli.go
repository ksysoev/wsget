package cli

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/user"

	"github.com/eiannone/keyboard"
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

func NewCLI(wsConn *ws.Connection, input Inputer, output io.Writer) *CLI {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	homeDir := currentUser.HomeDir

	history := NewHistory(homeDir+"/"+HistoryFilename, HistoryLimit)

	return &CLI{
		formater: formater.NewFormatter(),
		editor:   NewEditor(output, history),
		wsConn:   wsConn,
		input:    input,
		output:   output,
	}
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

	if opts.StartEditor {
		if err := c.RequestMod(keysEvents); err != nil {
			return err
		}
	} else {
		fmt.Fprintln(c.output, "Connection Mode: Press ESC to enter Request mode")
	}

	for {
		select {
		case event := <-keysEvents:
			switch event.Key {
			case keyboard.KeyCtrlC, keyboard.KeyCtrlD:
				return nil

			case keyboard.KeyEsc:
				if err := c.RequestMod(keysEvents); err != nil {
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
				log.Printf("Fail to format message: %s, %s\n", err, msg.Data)
			}

			fmt.Fprintf(c.output, "%s\n\n", output)

			if opts.OutputFile != nil {
				output, err := c.formater.FormatForFile(msg)
				if err != nil {
					log.Printf("Fail to format message for file: %s, %s\n", err, msg.Data)
				}

				fmt.Fprintln(c.output, opts.OutputFile, output)
			}
		}
	}
}

func (c *CLI) RequestMod(keysEvents <-chan keyboard.KeyEvent) error {
	fmt.Fprintln(c.output, "Request Mode: Type your API request and press Ctrl+S to send it. Press ESC to cancel request")

	req, err := c.editor.EditRequest(keysEvents, "")
	if err != nil {
		if err.Error() == "interrupted" {
			return nil
		}

		fmt.Fprintln(c.output, err)
	}

	if req != "" {
		err = c.wsConn.Send(req)
		if err != nil {
			fmt.Fprintln(c.output, "Fail to send request:", err)
		}
	}

	fmt.Fprintln(c.output, "Connection Mode: Press ESC to enter Request mode")

	return nil
}
