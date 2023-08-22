package formater

import (
	"testing"

	"github.com/fatih/color"
)

func TestTextFormater_FormatRequest(t *testing.T) {
	tf := NewTextFormater()
	data := "test request data"
	expectedOutput := color.New(color.FgGreen).Sprintf("test request data")

	output, err := tf.FormatRequest(data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if output != expectedOutput {
		t.Errorf("Unexpected output: got %q, expected %q", output, expectedOutput)
	}
}

func TestTextFormater_FormatResponse(t *testing.T) {
	tf := NewTextFormater()
	data := "test response data"
	expectedOutput := color.New(color.FgHiRed).Sprintf("test response data")

	output, err := tf.FormatResponse(data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if output != expectedOutput {
		t.Errorf("Unexpected output: got %q, expected %q", output, expectedOutput)
	}
}

func TestTextFormater_FormatForFile(t *testing.T) {
	tf := NewTextFormater()
	data := "test data"

	expectedOutput := "test data"

	output, err := tf.FormatForFile(data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if output != expectedOutput {
		t.Errorf("Unexpected output: got %q, expected %q", output, expectedOutput)
	}
}
