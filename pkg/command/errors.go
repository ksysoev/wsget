package command

type ErrUnknownCommand struct {
	Command string
}

func (e ErrUnknownCommand) Error() string {
	return "unknown command: " + e.Command
}

type ErrConnectionClosed struct{}

func (e ErrConnectionClosed) Error() string {
	return "connection closed"
}

type ErrTimeout struct{}

func (e ErrTimeout) Error() string {
	return "timeout"
}

type ErrUnsupportedMessageType struct {
	MsgType string
}

func (e ErrUnsupportedMessageType) Error() string {
	return "unsupported message type: " + e.MsgType
}

type ErrEmptyRequest struct{}

func (e ErrEmptyRequest) Error() string {
	return "empty request"
}

type ErrInvalidTimeout struct {
	Timeout string
}

func (e ErrInvalidTimeout) Error() string {
	return "invalid timeout: " + e.Timeout
}

type ErrEmptyCommand struct{}

func (e ErrEmptyCommand) Error() string {
	return "empty command"
}

type ErrEmptyMacro struct {
	MacroName string
}

func (e ErrEmptyMacro) Error() string {
	return "empty macro: " + e.MacroName
}

type ErrDuplicateMacro struct {
	MacroName string
}

func (e ErrDuplicateMacro) Error() string {
	return "duplicate macro: " + e.MacroName
}

type ErrUnsupportedVersion struct {
	Version string
}

func (e ErrUnsupportedVersion) Error() string {
	return "unsupported version: " + e.Version
}
