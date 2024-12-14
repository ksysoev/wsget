package core

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/fatih/color"
)

type executionContext struct {
	cli        *CLI
	outputFile io.Writer
	ctx        context.Context
}

// newExecutionContext creates a new executionContext instance for the provided CLI and output file.
// It takes cli of type *CLI, which manages command-line interactions, and outputFile of type io.Writer for output operations.
// It returns an *executionContext initialized with the given CLI and output writer.
func newExecutionContext(cli *CLI, outputFile io.Writer) *executionContext {
	return &executionContext{
		cli:        cli,
		outputFile: outputFile,
	}
}

// Print writes the given data to the CLI's output with optional color attributes.
// It takes data of type string, which is the text to be printed, and attr, variadic arguments of type color.Attribute for styling.
// It returns an error if writing to the CLI's output fails.
func (c *executionContext) Print(data string, attr ...color.Attribute) error {
	_, err := color.New(attr...).Fprintln(c.cli.output, data)
	return err
}

func (c *executionContext) PrintToFile(data string) error {
	if c.outputFile == nil {
		return nil
	}

	_, err := fmt.Fprintln(c.outputFile, data)
	return err
}

func (c *executionContext) FormatMessage(msg Message, noColor bool) (string, error) {
	if noColor {
		return c.cli.formater.FormatForFile(msg.Type.String(), msg.Data)
	}

	return c.cli.formater.FormatMessage(msg.Type.String(), msg.Data)
}

// SendRequest sends a request message through the execution context's WebSocket connection.
// It takes req of type string, which represents the request to be sent.
// It returns an error if the WebSocket connection fails to send the request.
func (c *executionContext) SendRequest(req string) error {
	return c.cli.wsConn.Send(c.ctx, req)
}

// WaitForResponse waits for a response message from the CLI within a specified timeout period.
// It takes timeout of type time.Duration to define the maximum wait time. If timeout is 0, it waits indefinitely.
// It returns a Message containing the received data and an error if the context deadline exceeds or other issues occur.
func (c *executionContext) WaitForResponse(timeout time.Duration) (Message, error) {
	ctx := c.ctx

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	select {
	case msg := <-c.cli.messages:
		return msg, nil
	case <-ctx.Done():
		return Message{}, ctx.Err()
	}
}

// EditorMode allows the user to edit text in an editor with a provided initial buffer.
// It takes initBuffer of type string, which initializes the editor with existing content.
// It returns a string containing the final edited content and an error if the editing process fails.
func (c *executionContext) EditorMode(initBuffer string) (string, error) {
	return c.cli.editor.Edit(c.ctx, initBuffer)
}

// CommandMode initiates command mode in the editor with the provided initial buffer.
// It takes initBuffer of type string, which is the input buffer to initialize the command mode.
// It returns a string representing the final buffer after editing and an error if command mode fails.
func (c *executionContext) CommandMode(initBuffer string) (string, error) {
	return c.cli.editor.CommandMode(c.ctx, initBuffer)
}

// CreateCommand creates an Executer from a raw command string.
// It takes a raw string representing the command to be created.
// It returns an Executer and an error if the command cannot be created.
func (c *executionContext) CreateCommand(raw string) (Executer, error) {
	return c.cli.cmdFactory.Create(raw)
}
