package core

import (
	"context"
	"io"
)

type executionContext struct {
	cli        *CLI
	outputFile io.Writer
}

func newExecutionContext(cli *CLI, outputFile io.Writer) *executionContext {
	return &executionContext{
		cli:        cli,
		outputFile: outputFile,
	}
}

func (ctx *executionContext) OutputFile() io.Writer {
	return ctx.outputFile
}

func (ctx *executionContext) Output() io.Writer {
	return ctx.cli.output
}

func (ctx *executionContext) Formater() Formater {
	return ctx.cli.formater
}

func (ctx *executionContext) Connection() ConnectionHandler {
	return ctx.cli.wsConn
}

func (ctx *executionContext) WaitForMessage(c context.Context) (Message, error) {
	return ctx.cli.WaitForMessage(c)
}

func (ctx *executionContext) Editor() Editor {
	return ctx.cli.editor
}

func (ctx *executionContext) Factory() CommandFactory {
	return ctx.cli.cmdFactory
}
