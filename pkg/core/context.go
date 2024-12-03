package core

import (
	"io"

	"github.com/ksysoev/wsget/pkg/command"
	"github.com/ksysoev/wsget/pkg/formater"
	"github.com/ksysoev/wsget/pkg/ws"
)

type ExecutionContext struct {
	input      <-chan KeyEvent
	cli        *CLI
	outputFile io.Writer
}

func NewExecutionContext(cli *CLI, outputFile io.Writer) *ExecutionContext {
	return &ExecutionContext{
		cli:        cli,
		outputFile: outputFile,
	}
}

func (ctx *ExecutionContext) Input() <-chan KeyEvent {
	return ctx.cli.inputStream
}

func (ctx *ExecutionContext) OutputFile() io.Writer {
	return ctx.outputFile
}

func (ctx *ExecutionContext) Output() io.Writer {
	return ctx.cli.output
}

func (ctx *ExecutionContext) Formater() formater.Formater {
	return ctx.cli.formater
}

func (ctx *ExecutionContext) Connection() ws.ConnectionHandler {
	return ctx.cli.wsConn
}

func (ctx *ExecutionContext) RequestEditor() command.Editor {
	return ctx.cli.editor
}

func (ctx *ExecutionContext) CmdEditor() command.Editor {
	return ctx.cli.cmdEditor
}

func (ctx *ExecutionContext) Macro() *command.Macro {
	return ctx.cli.macro
}
