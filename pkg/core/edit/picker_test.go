package edit

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/ksysoev/wsget/pkg/core"
	"github.com/ksysoev/wsget/pkg/repo/history"
	"github.com/stretchr/testify/assert"
)

func TestNewFuzzyPicker(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}

	picker := NewFuzzyPicker(output, mockHistory)

	assert.NotNil(t, picker)
	assert.Equal(t, output, picker.output)
	assert.Equal(t, mockHistory, picker.history)
}

func TestFuzzyPicker_SetInput(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}
	picker := NewFuzzyPicker(output, mockHistory)

	input := make(chan core.KeyEvent)
	picker.SetInput(input)

	// Test that input was set (we can't directly compare channels due to type conversion)
	assert.NotNil(t, picker.input)
}

func TestFuzzyPicker_Pick_NoInput(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}
	picker := NewFuzzyPicker(output, mockHistory)

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "input stream is not set")
}

func TestFuzzyPicker_Pick_ContextCanceled(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}
	mockHistory.On("FuzzySearch", "").Return([]history.FuzzyMatch{})

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent)
	picker.SetInput(input)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := picker.Pick(ctx)

	assert.ErrorIs(t, err, core.ErrInterrupted)
	assert.Empty(t, result)
}

func TestFuzzyPicker_Pick_Escape(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}
	mockHistory.On("FuzzySearch", "").Return([]history.FuzzyMatch{})

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 1)
	picker.SetInput(input)

	input <- core.KeyEvent{Key: core.KeyEsc}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.ErrorIs(t, err, core.ErrInterrupted)
	assert.Empty(t, result)
}

func TestFuzzyPicker_Pick_EnterWithSelection(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}

	matches := []history.FuzzyMatch{
		{Request: "test request 1", Score: 100},
		{Request: "test request 2", Score: 90},
	}

	mockHistory.On("FuzzySearch", "").Return(matches)

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 1)
	picker.SetInput(input)

	// Send Enter key to select the first match
	input <- core.KeyEvent{Key: core.KeyEnter}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "test request 1", result)
}

func TestFuzzyPicker_Pick_UTF8Backspace(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}

	mockHistory.On("FuzzySearch", "").Return([]history.FuzzyMatch{})
	mockHistory.On("FuzzySearch", "你").Return([]history.FuzzyMatch{
		{Request: "你好世界", Score: 100},
	})
	mockHistory.On("FuzzySearch", "你好").Return([]history.FuzzyMatch{
		{Request: "你好世界", Score: 100},
	})
	mockHistory.On("FuzzySearch", "你好🎉").Return([]history.FuzzyMatch{
		{Request: "你好世界🎉", Score: 100},
	})

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 10) // Increased buffer size
	picker.SetInput(input)

	// Type Chinese characters
	input <- core.KeyEvent{Key: 0, Rune: '你'}

	input <- core.KeyEvent{Key: 0, Rune: '好'}

	// Type emoji
	input <- core.KeyEvent{Key: 0, Rune: '🎉'}

	// Backspace should remove emoji (multi-byte character)
	input <- core.KeyEvent{Key: core.KeyBackspace}

	// Backspace should remove '好'
	input <- core.KeyEvent{Key: core.KeyBackspace}

	input <- core.KeyEvent{Key: core.KeyEnter}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "你好世界", result)
	mockHistory.AssertCalled(t, "FuzzySearch", "你")
	mockHistory.AssertCalled(t, "FuzzySearch", "你好")
	mockHistory.AssertCalled(t, "FuzzySearch", "你好🎉")
}

func TestFuzzyPicker_FormatMatchLine_UTF8Truncation(t *testing.T) {
	tests := []struct {
		name       string
		request    string
		isSelected bool
		maxLen     int
	}{
		{
			name:       "ASCII string under limit",
			request:    "short",
			isSelected: false,
			maxLen:     100,
		},
		{
			name:       "Long ASCII string",
			request:    strings.Repeat("a", 150),
			isSelected: false,
			maxLen:     100,
		},
		{
			name:       "UTF-8 Chinese characters",
			request:    strings.Repeat("你好", 60), // Each character is 3 bytes
			isSelected: true,
			maxLen:     100,
		},
		{
			name:       "UTF-8 emoji",
			request:    strings.Repeat("🎉", 60), // Each emoji is 4 bytes
			isSelected: false,
			maxLen:     100,
		},
		{
			name:       "Mixed ASCII and UTF-8",
			request:    strings.Repeat("Hello你好🎉", 15),
			isSelected: true,
			maxLen:     100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := history.FuzzyMatch{Request: tt.request}
			result := formatMatchLine(match, tt.isSelected)

			// Should not panic on multi-byte characters
			assert.NotEmpty(t, result)

			// Remove ANSI codes for length checking
			cleaned := strings.ReplaceAll(result, "\033[7m", "")
			cleaned = strings.ReplaceAll(cleaned, "\033[0m", "")
			cleaned = strings.TrimPrefix(cleaned, "> ")
			cleaned = strings.TrimPrefix(cleaned, "  ")

			// Check that result is properly truncated (in runes, not bytes)
			if len([]rune(tt.request)) > tt.maxLen {
				assert.Contains(t, cleaned, "...")
				// Verify truncation doesn't break UTF-8 encoding
				assert.True(t, len([]rune(cleaned)) <= tt.maxLen+3) // +3 for "..."
			}
		})
	}
}

func TestFuzzyPicker_Pick_UTF8CharInput(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}

	mockHistory.On("FuzzySearch", "").Return([]history.FuzzyMatch{})
	mockHistory.On("FuzzySearch", "П").Return([]history.FuzzyMatch{})
	mockHistory.On("FuzzySearch", "Пр").Return([]history.FuzzyMatch{})
	mockHistory.On("FuzzySearch", "При").Return([]history.FuzzyMatch{})
	mockHistory.On("FuzzySearch", "Прив").Return([]history.FuzzyMatch{})
	mockHistory.On("FuzzySearch", "Приве").Return([]history.FuzzyMatch{})
	mockHistory.On("FuzzySearch", "Привет").Return([]history.FuzzyMatch{
		{Request: "Привет мир", Score: 100},
	})

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 10) // Increased buffer size
	picker.SetInput(input)

	// Type Russian characters (Cyrillic)
	for _, r := range "Привет" {
		input <- core.KeyEvent{Key: 0, Rune: r}
	}

	input <- core.KeyEvent{Key: core.KeyEnter}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "Привет мир", result)
	mockHistory.AssertCalled(t, "FuzzySearch", "Привет")
}

func TestFuzzyPicker_Pick_EnterWithNoMatches(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}
	mockHistory.On("FuzzySearch", "").Return([]history.FuzzyMatch{})

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 1)
	picker.SetInput(input)

	input <- core.KeyEvent{Key: core.KeyEnter}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestFuzzyPicker_Pick_CharInput(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}

	// Initial render with empty query
	mockHistory.On("FuzzySearch", "").Return([]history.FuzzyMatch{})

	// After typing 't'
	matches := []history.FuzzyMatch{
		{Request: "test", Score: 100},
	}
	mockHistory.On("FuzzySearch", "t").Return(matches)

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 2)
	picker.SetInput(input)

	input <- core.KeyEvent{Key: 0, Rune: 't'}

	input <- core.KeyEvent{Key: core.KeyEnter}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "test", result)
}

func TestFuzzyPicker_Pick_ArrowNavigation(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}

	matches := []history.FuzzyMatch{
		{Request: "first", Score: 100},
		{Request: "second", Score: 90},
		{Request: "third", Score: 80},
	}

	mockHistory.On("FuzzySearch", "").Return(matches)

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 3)
	picker.SetInput(input)

	// Navigate down twice, then select
	input <- core.KeyEvent{Key: core.KeyArrowDown}

	input <- core.KeyEvent{Key: core.KeyArrowDown}

	input <- core.KeyEvent{Key: core.KeyEnter}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "third", result)
}

func TestFuzzyPicker_Pick_ArrowUp(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}

	matches := []history.FuzzyMatch{
		{Request: "first", Score: 100},
		{Request: "second", Score: 90},
	}

	mockHistory.On("FuzzySearch", "").Return(matches)

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 4)
	picker.SetInput(input)

	// Navigate down, then up, then select
	input <- core.KeyEvent{Key: core.KeyArrowDown}

	input <- core.KeyEvent{Key: core.KeyArrowUp}

	input <- core.KeyEvent{Key: core.KeyEnter}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "first", result)
}

func TestFuzzyPicker_Pick_Backspace(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}

	mockHistory.On("FuzzySearch", "").Return([]history.FuzzyMatch{})
	mockHistory.On("FuzzySearch", "t").Return([]history.FuzzyMatch{
		{Request: "test", Score: 100},
	})
	mockHistory.On("FuzzySearch", "te").Return([]history.FuzzyMatch{
		{Request: "test", Score: 100},
	})

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 4)
	picker.SetInput(input)

	input <- core.KeyEvent{Key: 0, Rune: 't'}

	input <- core.KeyEvent{Key: 0, Rune: 'e'}

	input <- core.KeyEvent{Key: core.KeyBackspace}

	input <- core.KeyEvent{Key: core.KeyEnter}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "test", result)
}

func TestFuzzyPicker_Pick_CtrlU(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}

	mockHistory.On("FuzzySearch", "").Return([]history.FuzzyMatch{})
	mockHistory.On("FuzzySearch", "t").Return([]history.FuzzyMatch{
		{Request: "test", Score: 100},
	})
	mockHistory.On("FuzzySearch", "te").Return([]history.FuzzyMatch{
		{Request: "test", Score: 100},
	})
	mockHistory.On("FuzzySearch", "tes").Return([]history.FuzzyMatch{
		{Request: "test", Score: 100},
	})

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 5)
	picker.SetInput(input)

	input <- core.KeyEvent{Key: 0, Rune: 't'}

	input <- core.KeyEvent{Key: 0, Rune: 'e'}

	input <- core.KeyEvent{Key: 0, Rune: 's'}

	input <- core.KeyEvent{Key: core.KeyCtrlU}

	input <- core.KeyEvent{Key: core.KeyEnter}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestFuzzyPicker_FormatMatchLine(t *testing.T) {
	tests := []struct {
		name       string
		expected   string
		match      history.FuzzyMatch
		isSelected bool
	}{
		{
			name:       "Not selected",
			match:      history.FuzzyMatch{Request: "test request"},
			isSelected: false,
			expected:   "  test request",
		},
		{
			name:       "Selected",
			match:      history.FuzzyMatch{Request: "test request"},
			isSelected: true,
			expected:   "\033[7m> test request\033[0m",
		},
		{
			name:       "Long line truncated",
			match:      history.FuzzyMatch{Request: strings.Repeat("a", 150)},
			isSelected: false,
			expected:   "...",
		},
		{
			name:       "Multiline replaced with spaces",
			match:      history.FuzzyMatch{Request: "line1\nline2\nline3"},
			isSelected: false,
			expected:   "  line1 line2 line3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMatchLine(tt.match, tt.isSelected)
			assert.Contains(t, result, tt.expected)
		})
	}
}

func TestFuzzyPicker_Pick_InputStreamClosed(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}
	mockHistory.On("FuzzySearch", "").Return([]history.FuzzyMatch{})

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent)
	picker.SetInput(input)

	close(input) // Close the channel immediately

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "keyboard stream was unexpectedly closed")
}

func TestFuzzyPicker_Pick_CtrlC(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}
	mockHistory.On("FuzzySearch", "").Return([]history.FuzzyMatch{})

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 1)
	picker.SetInput(input)

	input <- core.KeyEvent{Key: core.KeyCtrlC}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.ErrorIs(t, err, core.ErrInterrupted)
	assert.Empty(t, result)
}

func TestFuzzyPicker_Pick_CtrlD(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}
	mockHistory.On("FuzzySearch", "").Return([]history.FuzzyMatch{})

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 1)
	picker.SetInput(input)

	input <- core.KeyEvent{Key: core.KeyCtrlD}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.ErrorIs(t, err, core.ErrInterrupted)
	assert.Empty(t, result)
}

func TestFuzzyPicker_Pick_MaxDisplayLines(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}

	// Create more matches than maxDisplayLines (10)
	matches := make([]history.FuzzyMatch, 15)
	for i := 0; i < 15; i++ {
		matches[i] = history.FuzzyMatch{
			Request: fmt.Sprintf("test_%d", i),
			Score:   100 - i,
		}
	}

	mockHistory.On("FuzzySearch", "").Return(matches)

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 20)
	picker.SetInput(input)

	// Try to navigate down beyond maxDisplayLines
	for i := 0; i < 11; i++ {
		input <- core.KeyEvent{Key: core.KeyArrowDown}
	}

	input <- core.KeyEvent{Key: core.KeyEnter}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.NoError(t, err)
	// Should select the 10th item (index 9) since maxDisplayLines is 10
	assert.Equal(t, matches[9].Request, result)
}

func TestFuzzyPicker_Pick_BackspaceOnEmpty(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}
	mockHistory.On("FuzzySearch", "").Return([]history.FuzzyMatch{})

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 2)
	picker.SetInput(input)

	// Try backspace on empty query - should do nothing
	input <- core.KeyEvent{Key: core.KeyBackspace}

	input <- core.KeyEvent{Key: core.KeyEnter}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestFuzzyPicker_Pick_ArrowUpAtTop(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}

	matches := []history.FuzzyMatch{
		{Request: "first", Score: 100},
		{Request: "second", Score: 90},
	}

	mockHistory.On("FuzzySearch", "").Return(matches)

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 3)
	picker.SetInput(input)

	// Try to go up when already at top - should stay at first

	input <- core.KeyEvent{Key: core.KeyArrowUp}

	input <- core.KeyEvent{Key: core.KeyArrowUp}

	input <- core.KeyEvent{Key: core.KeyEnter}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "first", result)
}

func TestFuzzyPicker_Pick_IgnoreNewline(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}
	mockHistory.On("FuzzySearch", "").Return([]history.FuzzyMatch{})

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 2)
	picker.SetInput(input)

	// Newline should be ignored in query
	input <- core.KeyEvent{Key: 0, Rune: '\n'}

	input <- core.KeyEvent{Key: core.KeyEnter}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestFuzzyPicker_Pick_IgnoreNonPrintableKeys(t *testing.T) {
	output := new(bytes.Buffer)
	mockHistory := &MockHistoryRepo{}
	mockHistory.On("FuzzySearch", "").Return([]history.FuzzyMatch{})

	picker := NewFuzzyPicker(output, mockHistory)
	input := make(chan core.KeyEvent, 3)
	picker.SetInput(input)

	// Keys with non-zero Key value and zero Rune should be ignored
	input <- core.KeyEvent{Key: 999, Rune: 0}

	input <- core.KeyEvent{Key: core.KeyTab, Rune: 0}

	input <- core.KeyEvent{Key: core.KeyEnter}

	ctx := context.Background()
	result, err := picker.Pick(ctx)

	assert.NoError(t, err)
	assert.Empty(t, result)
}
