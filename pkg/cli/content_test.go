// BEGIN: yz9d8f7g6h5j
package cli

import (
	"testing"
)

func TestNewContent(t *testing.T) {
	c := NewContent()

	if c == nil {
		t.Error("NewContent() returned nil")
		return
	}

	if c.pos != 0 {
		t.Errorf("NewContent() returned Content with pos %d, expected 0", c.pos)
	}

	if len(c.text) != 0 {
		t.Errorf("NewContent() returned Content with text %v, expected empty slice", c.text)
	}
}

func TestReplaceEmptyText(t *testing.T) {
	c := NewContent()
	text := "TestReplaceEmptyText"
	output := c.ReplaceText(text)

	if c.String() != text {
		t.Errorf("ReplaceText() did not set Content text correctly, expected %s, got %s", text, c.String())
	}

	if c.pos != len(text) {
		t.Errorf("ReplaceText() did not set Content pos correctly, expected %d, got %d", len(text), c.pos)
	}

	if output != text {
		t.Errorf("ReplaceText() did not return correct output, expected %s, got %s", c.Clear()+text, output)
	}
}

func TestReplaceNonEmptyText(t *testing.T) {
	c := NewContent()
	c.ReplaceText("buy")

	text := "hello"
	output := c.ReplaceText(text)

	if c.String() != text {
		t.Errorf("ReplaceText() did not set Content text correctly, expected %s, got %s", text, c.String())
	}

	if c.pos != len(text) {
		t.Errorf("ReplaceText() did not set Content pos correctly, expected %d, got %d", len(text), c.pos)
	}

	expected := "\x1b[2K\r" + text
	if output != expected {
		t.Errorf("ReplaceText() did not return correct output, expected %q, got %q", expected, output)
	}
}

func TestReplaceMultiLineText(t *testing.T) {
	c := NewContent()
	c.ReplaceText("buy\nsome milk")

	text := "hello"
	output := c.ReplaceText(text)

	if c.String() != text {
		t.Errorf("ReplaceText() did not set Content text correctly, expected %s, got %s", text, c.String())
	}

	if c.pos != len(text) {
		t.Errorf("ReplaceText() did not set Content pos correctly, expected %d, got %d", len(text), c.pos)
	}

	expected := "\x1b[2K\r\033[1A\x1b[2K\r" + text
	if output != expected {
		t.Errorf("ReplaceText() did not return correct output, expected %q, got %q", expected, output)
	}
}

func TestString(t *testing.T) {
	c := NewContent()
	text := "hello world"
	c.ReplaceText(text)

	if c.String() != text {
		t.Errorf("String() did not return correct string, expected %s, got %s", text, c.String())
	}
}

func TestToRequest(t *testing.T) {
	c := NewContent()
	text := " hello world\n"
	c.ReplaceText(text)

	if c.ToRequest() != "hello world" {
		t.Errorf("ToRequest() did not return correct string, expected 'hello world', got '%s'", c.ToRequest())
	}
}

func TestContent_Clear(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty content",
			input:    "",
			expected: "",
		},
		{
			name:     "single line content",
			input:    "hello world",
			expected: "\x1b[2K\r",
		},
		{
			name:     "multi-line content",
			input:    "hello\nworld",
			expected: "\x1b[2K\r\x1b[1A\x1b[2K\r",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Content{
				text: []rune(tt.input),
				pos:  len(tt.input),
			}

			actual := c.Clear()
			if actual != tt.expected {
				t.Errorf("unexpected output: got %q, want %q", actual, tt.expected)
			}
		})
	}
}
