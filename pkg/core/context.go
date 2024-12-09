package core

import (
	"io"

	"github.com/ksysoev/wsget/pkg/formater"
)

type executionContext struct {
	input      <-chan KeyEvent
	cli        *CLI
	outputFile io.Writer
}

func newExecutionContext(cli *CLI, outputFile io.Writer) *executionContext {
	return &executionContext{
		cli:        cli,
		outputFile: outputFile,
	}
}

func (ctx *executionContext) Input() <-chan KeyEvent {
	return ctx.cli.inputStream
}

func (ctx *executionContext) OutputFile() io.Writer {
	return ctx.outputFile
}

func (ctx *executionContext) Output() io.Writer {
	return ctx.cli.output
}

func (ctx *executionContext) Formater() formater.Formater {
	return ctx.cli.formater
}

func (ctx *executionContext) Connection() ConnectionHandler {
	return ctx.cli.wsConn
}

func (ctx *executionContext) Editor() Editor {
	return ctx.cli.editor
}

func (ctx *executionContext) Factory() CommandFactory {
	return ctx.cli.cmdFactory
}
