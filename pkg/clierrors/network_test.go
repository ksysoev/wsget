package clierrors

import "testing"

func TestConnectionClosed_Error(t *testing.T) {
	err := ConnectionClosed{}
	want := "connection closed"

	if got := err.Error(); got != want {
		t.Errorf("Error() = %v, want %v", got, want)
	}
}

func TestTimeout_Error(t *testing.T) {
	err := Timeout{}
	want := "timeout"

	if got := err.Error(); got != want {
		t.Errorf("Error() = %v, want %v", got, want)
	}
}
