package core

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/fatih/color"
)

type executionContext struct {
	cli        *CLI
	outputFile io.Writer
	ctx        context.Context
}

func newExecutionContext(cli *CLI, outputFile io.Writer) *executionContext {
	return &executionContext{
		cli:        cli,
		outputFile: outputFile,
	}
}

func (c *executionContext) Print(data string) error {
	_, err := fmt.Fprintln(c.cli.output, data)
	return err
}

func (c *executionContext) PrintMessage(msg Message) error {
	output, err := c.cli.formater.FormatMessage(msg.Type.String(), msg.Data)

	if err != nil {
		return err
	}

	switch msg.Type {
	case Request:
		_, _ = color.New(color.FgGreen).Fprintln(c.cli.output, "->")
	case Response:
		_, _ = color.New(color.FgRed).Fprintln(c.cli.output, "<-")
	default:
		return fmt.Errorf("unsupported message type: %s", msg.Type.String())
	}

	_, _ = fmt.Fprintf(c.cli.output, "%s\n", output)

	outputFile := c.outputFile
	if outputFile != nil && !reflect.ValueOf(outputFile).IsNil() {
		output, err := c.cli.formater.FormatForFile(msg.Type.String(), msg.Data)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(outputFile, output)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *executionContext) SendRequest(request string) error {
	return c.cli.wsConn.Send(c.ctx, request)
}

func (c *executionContext) WaitForResponse(timeout time.Duration) (Message, error) {
	ctx := c.ctx

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	return c.cli.WaitForMessage(ctx)
}

func (c *executionContext) EditorMode(initBuffer string) (string, error) {
	return c.cli.editor.Edit(c.ctx, initBuffer)
}

func (c *executionContext) CommandMode(initBuffer string) (string, error) {
	return c.cli.editor.CommandMode(c.ctx, initBuffer)
}

func (c *executionContext) CreateCommand(raw string) (Executer, error) {
	return c.cli.cmdFactory.Create(raw)
}
