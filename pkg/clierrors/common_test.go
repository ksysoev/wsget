package clierrors

import "testing"

func TestInterrupted_Error(t *testing.T) {
	err := Interrupted{}
	expected := "interrupted"

	if err.Error() != expected {
		t.Errorf("Expected error message '%s', but got '%s'", expected, err.Error())
	}
}
