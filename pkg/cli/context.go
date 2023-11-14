package cli

import (
	"io"

	"github.com/eiannone/keyboard"
	"github.com/ksysoev/wsget/pkg/cmd"
)

type ExecutionContext struct {
	input      <-chan keyboard.KeyEvent
	cli        *CLI
	outputFile io.Writer
}

func NewExecutionContext(cli *CLI, input <-chan keyboard.KeyEvent, outputFile io.Writer) *ExecutionContext {
	return &ExecutionContext{
		input:      input,
		cli:        cli,
		outputFile: outputFile,
	}
}

func (ctx *ExecutionContext) Input() <-chan keyboard.KeyEvent {
	return ctx.input
}

func (ctx *ExecutionContext) OutputFile() io.Writer {
	return ctx.outputFile
}

func (ctx *ExecutionContext) Output() io.Writer {
	return ctx.cli.output
}

func (ctx *ExecutionContext) Formater() Formater {
	return ctx.cli.formater
}

func (ctx *ExecutionContext) Connection() ConnectionHandler {
	return ctx.cli.wsConn
}

func (ctx *ExecutionContext) RequestEditor() *cmd.Editor {
	return ctx.cli.editor
}

func (ctx *ExecutionContext) CmdEditor() *cmd.Editor {
	return ctx.cli.cmdEditor
}

func (ctx *ExecutionContext) Macro() *cmd.Macro {
	return ctx.cli.macro
}
