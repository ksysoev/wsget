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

func (c *executionContext) Print(data string, attr ...color.Attribute) error {
	_, err := color.New(attr...).Fprintln(c.cli.output, data)
	return err
}

func (c *executionContext) PrintMessage(msg Message) error {
	output, err := c.cli.formater.FormatMessage(msg.Type.String(), msg.Data)

	if err != nil {
		return fmt.Errorf("fail to format message: %w", err)
	}

	switch msg.Type {
	case Request:
		err = c.Print("->", color.FgGreen)
	case Response:
		err = c.Print("<-", color.FgRed)
	default:
		return fmt.Errorf("unsupported message type: %s", msg.Type.String())
	}

	if err != nil {
		return fmt.Errorf("fail to print message: %w", err)
	}

	if err := c.Print("%s\n" + output); err != nil {
		return fmt.Errorf("fail to print message: %w", err)
	}

	outputFile := c.outputFile
	if outputFile != nil && !reflect.ValueOf(outputFile).IsNil() {
		output, err := c.cli.formater.FormatForFile(msg.Type.String(), msg.Data)
		if err != nil {
			return fmt.Errorf("fail to format message for file: %w", err)
		}

		if _, err := fmt.Fprintln(outputFile, output); err != nil {
			return fmt.Errorf("fail to write to output file: %w", err)
		}
	}

	return nil
}

func (c *executionContext) SendRequest(req string) error {
	return c.cli.wsConn.Send(c.ctx, req)
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
