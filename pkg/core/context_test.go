package core

import (
	"bytes"
	"testing"
)

func TestNewExecutionContext(t *testing.T) {
	cli := &CLI{
		inputStream: make(chan KeyEvent),
	}
	outputFile := &bytes.Buffer{}

	executionContext := newExecutionContext(cli, outputFile)

	if executionContext.cli != cli {
		t.Errorf("Unexpected CLI: %v", executionContext.cli)
	}

	if executionContext.outputFile != outputFile {
		t.Errorf("Unexpected output file: %v", executionContext.outputFile)
	}
}
func TestExecutionContext_Connection(t *testing.T) {
	cli := &CLI{}
	outputFile := &bytes.Buffer{}

	executionContext := newExecutionContext(cli, outputFile)

	if executionContext.Connection() != cli.wsConn {
		t.Errorf("Unexpected connection: %v", executionContext.Connection())
	}
}

func TestExecutionContext_OutputFile(t *testing.T) {
	cli := &CLI{}
	outputFile := &bytes.Buffer{}

	executionContext := newExecutionContext(cli, outputFile)

	if executionContext.OutputFile() != outputFile {
		t.Errorf("Unexpected connection: %v", executionContext.OutputFile())
	}
}

func TestExecutionContext_Output(t *testing.T) {
	cli := &CLI{}
	outputFile := &bytes.Buffer{}

	executionContext := newExecutionContext(cli, outputFile)

	if executionContext.Output() != cli.output {
		t.Errorf("Unexpected connection: %v", executionContext.Output())
	}
}

func TestExecutionContext_Formater(t *testing.T) {
	cli := &CLI{}
	outputFile := &bytes.Buffer{}

	executionContext := newExecutionContext(cli, outputFile)

	if executionContext.Formater() != cli.formater {
		t.Errorf("Unexpected connection: %v", executionContext.Formater())
	}
}

func TestExecutionContext_RequestEditor(t *testing.T) {
	cli := &CLI{}
	outputFile := &bytes.Buffer{}

	executionContext := newExecutionContext(cli, outputFile)

	if executionContext.Editor() != cli.editor {
		t.Errorf("Unexpected connection: %v", executionContext.Editor())
	}
}
