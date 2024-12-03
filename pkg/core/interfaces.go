package core

import (
	"io"

	"github.com/ksysoev/wsget/pkg/formater"
	"github.com/ksysoev/wsget/pkg/ws"
)

type CommandFactory interface {
	New(raw string) (Executer, error)
}

type ExecutionContext interface {
	Input() <-chan KeyEvent
	OutputFile() io.Writer
	Output() io.Writer
	Formater() formater.Formater
	Connection() ws.ConnectionHandler
	RequestEditor() Editor
	CmdEditor() Editor
	Factory() CommandFactory
}

type Editor interface {
	Edit(keyStream <-chan KeyEvent, initBuffer string) (string, error)
	Close() error
}

type Executer interface {
	Execute(ExecutionContext) (Executer, error)
}
