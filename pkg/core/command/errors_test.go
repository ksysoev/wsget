package command

import "testing"

func TestUnknownCommand_Error(t *testing.T) {
	command := "test"
	err := ErrUnknownCommand{Command: command}
	expected := "unknown command: " + command

	if err.Error() != expected {
		t.Errorf("Expected error message '%s', but got '%s'", expected, err.Error())
	}
}

func TestConnectionClosed_Error(t *testing.T) {
	err := ErrConnectionClosed{}
	want := "connection closed"

	if got := err.Error(); got != want {
		t.Errorf("Error() = %v, want %v", got, want)
	}
}

func TestTimeout_Error(t *testing.T) {
	err := ErrTimeout{}
	want := "timeout"

	if got := err.Error(); got != want {
		t.Errorf("Error() = %v, want %v", got, want)
	}
}

func TestUnsupportedMessageType_Error(t *testing.T) {
	msgType := "binary"
	err := ErrUnsupportedMessageType{MsgType: msgType}
	expected := "unsupported message type: " + msgType

	if err.Error() != expected {
		t.Errorf("Expected error message '%s', but got '%s'", expected, err.Error())
	}
}

func TestEmptyRequest_Error(t *testing.T) {
	err := ErrEmptyRequest{}
	want := "empty request"

	if got := err.Error(); got != want {
		t.Errorf("Error() = %v, want %v", got, want)
	}
}

func TestInvalidTimeout_Error(t *testing.T) {
	timeout := "invalid"
	err := ErrInvalidTimeout{Timeout: timeout}
	expected := "invalid timeout: " + timeout

	if err.Error() != expected {
		t.Errorf("Expected error message '%s', but got '%s'", expected, err.Error())
	}
}

func TestEmptyCommand_Error(t *testing.T) {
	err := ErrEmptyCommand{}
	want := "empty command"

	if got := err.Error(); got != want {
		t.Errorf("Error() = %v, want %v", got, want)
	}
}

func TestEmptyMacro_Error(t *testing.T) {
	macroName := "testMacro"
	err := ErrEmptyMacro{MacroName: macroName}
	expected := "empty macro: " + macroName

	if err.Error() != expected {
		t.Errorf("Expected error message '%s', but got '%s'", expected, err.Error())
	}
}

func TestDuplicateMacro_Error(t *testing.T) {
	macroName := "duplicateMacro"
	err := ErrDuplicateMacro{MacroName: macroName}
	expected := "duplicate macro: " + macroName

	if err.Error() != expected {
		t.Errorf("Expected error message '%s', but got '%s'", expected, err.Error())
	}
}

func TestUnsupportedVersion_Error(t *testing.T) {
	version := "v2.0"
	err := ErrUnsupportedVersion{Version: version}
	expected := "unsupported version: " + version

	if err.Error() != expected {
		t.Errorf("Expected error message '%s', but got '%s'", expected, err.Error())
	}
}

func TestInvalidRepeatCommand_Error(t *testing.T) {
	err := ErrInvalidRepeatCommand{}
	want := "invalid repeat command"

	if got := err.Error(); got != want {
		t.Errorf("Error() = %v, want %v", got, want)
	}
}
