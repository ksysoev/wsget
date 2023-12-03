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
