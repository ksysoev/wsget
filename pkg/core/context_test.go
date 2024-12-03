package core

import (
	"bytes"
	"testing"

	"github.com/eiannone/keyboard"
)

func TestNewExecutionContext(t *testing.T) {
	cli := &CLI{}
	input := make(chan keyboard.KeyEvent)
	outputFile := &bytes.Buffer{}

	executionContext := NewExecutionContext(cli, input, outputFile)

	if executionContext.cli != cli {
		t.Errorf("Unexpected CLI: %v", executionContext.cli)
	}

	if executionContext.input != input {
		t.Errorf("Unexpected input channel: %v", executionContext.input)
	}

	if executionContext.outputFile != outputFile {
		t.Errorf("Unexpected output file: %v", executionContext.outputFile)
	}
}
func TestExecutionContext_Connection(t *testing.T) {
	cli := &CLI{}
	input := make(chan keyboard.KeyEvent)
	outputFile := &bytes.Buffer{}

	executionContext := NewExecutionContext(cli, input, outputFile)

	if executionContext.Connection() != cli.wsConn {
		t.Errorf("Unexpected connection: %v", executionContext.Connection())
	}
}

func TestExecutionContext_OutputFile(t *testing.T) {
	cli := &CLI{}
	input := make(chan keyboard.KeyEvent)
	outputFile := &bytes.Buffer{}

	executionContext := NewExecutionContext(cli, input, outputFile)

	if executionContext.OutputFile() != outputFile {
		t.Errorf("Unexpected connection: %v", executionContext.OutputFile())
	}
}

func TestExecutionContext_Output(t *testing.T) {
	cli := &CLI{}
	input := make(chan keyboard.KeyEvent)
	outputFile := &bytes.Buffer{}

	executionContext := NewExecutionContext(cli, input, outputFile)

	if executionContext.Output() != cli.output {
		t.Errorf("Unexpected connection: %v", executionContext.Output())
	}
}

func TestExecutionContext_Formater(t *testing.T) {
	cli := &CLI{}
	input := make(chan keyboard.KeyEvent)
	outputFile := &bytes.Buffer{}

	executionContext := NewExecutionContext(cli, input, outputFile)

	if executionContext.Formater() != cli.formater {
		t.Errorf("Unexpected connection: %v", executionContext.Formater())
	}
}

func TestExecutionContext_RequestEditor(t *testing.T) {
	cli := &CLI{}
	input := make(chan keyboard.KeyEvent)
	outputFile := &bytes.Buffer{}

	executionContext := NewExecutionContext(cli, input, outputFile)

	if executionContext.RequestEditor() != cli.editor {
		t.Errorf("Unexpected connection: %v", executionContext.RequestEditor())
	}
}

func TestExecutionContext_CmdEditor(t *testing.T) {
	cli := &CLI{}
	input := make(chan keyboard.KeyEvent)
	outputFile := &bytes.Buffer{}

	executionContext := NewExecutionContext(cli, input, outputFile)

	if executionContext.CmdEditor() != cli.cmdEditor {
		t.Errorf("Unexpected connection: %v", executionContext.CmdEditor())
	}
}

func TestExecutionContext_Macro(t *testing.T) {
	cli := &CLI{}
	input := make(chan keyboard.KeyEvent)
	outputFile := &bytes.Buffer{}

	executionContext := NewExecutionContext(cli, input, outputFile)

	if executionContext.Macro() != cli.macro {
		t.Errorf("Unexpected connection: %v", executionContext.Macro())
	}
}

func TestExecutionContext_Input(t *testing.T) {
	cli := &CLI{}
	input := make(chan keyboard.KeyEvent)
	outputFile := &bytes.Buffer{}

	executionContext := NewExecutionContext(cli, input, outputFile)

	if executionContext.Input() != input {
		t.Errorf("Unexpected connection: %v", executionContext.Input())
	}
}
