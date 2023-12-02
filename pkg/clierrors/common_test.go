package clierrors

import "testing"

func TestUnknownCommand_Error(t *testing.T) {
	command := "test"
	err := UnknownCommand{Command: command}
	expected := "unknown command: " + command

	if err.Error() != expected {
		t.Errorf("Expected error message '%s', but got '%s'", expected, err.Error())
	}
}
func TestInterrupted_Error(t *testing.T) {
	err := Interrupted{}
	expected := "interrupted"

	if err.Error() != expected {
		t.Errorf("Expected error message '%s', but got '%s'", expected, err.Error())
	}
}
