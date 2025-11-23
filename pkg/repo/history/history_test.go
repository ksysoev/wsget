package history

import (
	"bufio"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHistory(t *testing.T) {
	expectedFileName := "test"
	history := NewHistory(expectedFileName)

	assert.NotNil(t, history, "expected non-nil history instance")
	assert.Equal(t, DefaultLimit, history.limit, "expected correct default limit")
	assert.Equal(t, 0, len(history.requests), "expected requests length 0")
	assert.Equal(t, 0, history.pos, "expected pos 0")
	assert.Equal(t, expectedFileName, history.fileName, "expected correct file name")
}

func TestHistory_NextRequest(t *testing.T) {
	tests := []struct {
		name        string
		expected    string
		initial     []string
		initialPos  int
		expectedPos int
	}{
		{
			name:        "EmptyHistory",
			initial:     []string{},
			initialPos:  0,
			expected:    "",
			expectedPos: 0,
		},
		{
			name:        "SingleRequestAtEnd",
			initial:     []string{"request1"},
			initialPos:  0,
			expected:    "",
			expectedPos: 1,
		},
		{
			name:        "SingleRequestPastEnd",
			initial:     []string{"request1"},
			initialPos:  1,
			expected:    "",
			expectedPos: 1,
		},
		{
			name:        "MultipleRequestsMoveForward",
			initial:     []string{"request1", "request2", "request3"},
			initialPos:  0,
			expected:    "request2",
			expectedPos: 1,
		},
		{
			name:        "MultipleRequestsAtEnd",
			initial:     []string{"request1", "request2", "request3"},
			initialPos:  2,
			expected:    "",
			expectedPos: 3,
		},
		{
			name:        "MultipleRequestsPastEnd",
			initial:     []string{"request1", "request2", "request3"},
			initialPos:  3,
			expected:    "",
			expectedPos: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := &History{
				requests: tt.initial,
				pos:      tt.initialPos,
			}

			got := history.NextRequest()
			assert.Equal(t, tt.expected, got, "unexpected next request")
			assert.Equal(t, tt.expectedPos, history.pos, "unexpected position after NextRequest")
		})
	}
}

func TestHistory_Close(t *testing.T) {
	tmpDir := os.TempDir()

	tests := []struct {
		name          string
		history       *History
		expectedLines []string
		expectError   bool
	}{
		{
			name:          "NoRequests",
			history:       &History{fileName: tmpDir + "/history_no_requests.txt", requests: []string{}, limit: 10},
			expectedLines: []string(nil),
			expectError:   false,
		},
		{
			name:          "IncorrectPath",
			history:       &History{fileName: tmpDir + "/incorrect/path.txt", requests: []string{"request1"}, limit: 10},
			expectedLines: []string(nil),
			expectError:   true,
		},
		{
			name:          "SingleRequest",
			history:       &History{fileName: tmpDir + "/history_single_request.txt", requests: []string{"request1"}, limit: 10},
			expectedLines: []string{"request1"},
			expectError:   false,
		},
		{
			name:          "MultipleRequests",
			history:       &History{fileName: tmpDir + "/history_multiple_requests.txt", requests: []string{"request1", "request2"}, limit: 10},
			expectedLines: []string{"request1", "request2"},
			expectError:   false,
		},
		{
			name:          "ExceedLimit",
			history:       &History{fileName: tmpDir + "/history_exceed_limit.txt", requests: []string{"req1", "req2", "req3"}, limit: 2},
			expectedLines: []string{"req2", "req3"},
			expectError:   false,
		},
		{
			name:          "EmptyLines",
			history:       &History{fileName: tmpDir + "/history_empty_lines.txt", requests: []string{"", "req1", "", "req2", "", "req3", ""}, limit: 10},
			expectedLines: []string{"req1", "req2", "req3"},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the Close method
			err := tt.history.Close()
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Read the file and compare with expectedLines
			file, err := os.Open(tt.history.fileName)
			assert.NoError(t, err)

			defer func() { _ = file.Close() }()

			var lines []string

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}

			assert.Equal(t, tt.expectedLines, lines)
		})
	}
}

func TestHistory_AddRequest(t *testing.T) {
	tests := []struct {
		name          string
		initial       []string
		newRequest    string
		expected      []string
		finalPosition int
	}{
		{
			name:          "AddUniqueRequest",
			initial:       []string{"request1"},
			newRequest:    "request2",
			expected:      []string{"request1", "request2"},
			finalPosition: 2,
		},
		{
			name:          "AddDuplicateRequest",
			initial:       []string{"request1"},
			newRequest:    "request1",
			expected:      []string{"request1"},
			finalPosition: 1,
		},
		{
			name:          "AddEmptyRequest",
			initial:       []string{"request1"},
			newRequest:    "",
			expected:      []string{"request1"},
			finalPosition: 1,
		},
		{
			name:          "AddAfterDuplicateRequest",
			initial:       []string{"request1", "request2"},
			newRequest:    "request2",
			expected:      []string{"request1", "request2"},
			finalPosition: 2,
		},
		{
			name:          "AddNewRequestAfterDuplicate",
			initial:       []string{"request1", "request1"},
			newRequest:    "request2",
			expected:      []string{"request1", "request1", "request2"},
			finalPosition: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := NewHistory("test")
			history.requests = tt.initial
			history.pos = len(tt.initial)

			history.AddRequest(tt.newRequest)
			assert.Equal(t, tt.expected, history.requests)
			assert.Equal(t, tt.finalPosition, history.pos)
		})
	}
}

func TestHistory_LoadHistory(t *testing.T) {
	tmDir := os.TempDir()

	tests := []struct {
		setup         func(fileName string) error
		name          string
		fileName      string
		expectedCount int
		expectedError bool
	}{
		{
			name:     "FileNotFound",
			fileName: tmDir + "/non_existent.txt",
			setup: func(_ string) error {
				return nil
			},
			expectedError: false,
		},
		{
			name:     "FileNotFound",
			fileName: tmDir + "/incorrect/path.txt",
			setup: func(_ string) error {
				return nil
			},
			expectedError: true,
		},
		{
			name:     "EmptyFile",
			fileName: tmDir + "/empty.txt",
			setup: func(fileName string) error {
				f, err := os.Create(fileName)
				if err != nil {
					return err
				}

				defer func() { _ = f.Close() }()

				return nil
			},
			expectedError: false,
			expectedCount: 0,
		},
		{
			name:     "SingleEntry",
			fileName: tmDir + "/single_entry.txt",
			setup: func(fileName string) error {
				f, err := os.Create(fileName)
				if err != nil {
					return err
				}

				defer func() { _ = f.Close() }()

				_, err = f.WriteString("entry1\n")

				return err
			},
			expectedError: false,
			expectedCount: 1,
		},
		{
			name:     "MultipleEntries",
			fileName: tmDir + "/multiple_entries.txt",
			setup: func(fileName string) error {
				f, err := os.Create(fileName)
				if err != nil {
					return err
				}

				defer func() { _ = f.Close() }()

				_, err = f.WriteString("entry1\nentry2\nentry3\n")

				return err
			},
			expectedError: false,
			expectedCount: 3,
		},
		{
			name:     "EmptyLines",
			fileName: tmDir + "/multiple_entries.txt",
			setup: func(fileName string) error {
				f, err := os.Create(fileName)
				if err != nil {
					return err
				}

				defer func() { _ = f.Close() }()

				_, err = f.WriteString("\nentry1\n\nentry2\n\nentry3\n\n")

				return err
			},
			expectedError: false,
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.setup(tt.fileName); err != nil {
				t.Fatalf("failed to set up test: %v", err)
			}

			history, err := LoadFromFile(tt.fileName)
			if tt.expectedError {
				assert.Error(t, err, "expected error but got nil")
			} else {
				assert.NoError(t, err, "unexpected error")
				assert.Equal(t, tt.expectedCount, len(history.requests), "unexpected number of requests loaded")
			}
		})
	}
}

func TestPrevRequest(t *testing.T) {
	tests := []struct {
		name        string
		expected    string
		initial     []string
		initialPos  int
		expectedPos int
	}{
		{
			name:        "EmptyHistory",
			initial:     []string{},
			initialPos:  0,
			expected:    "",
			expectedPos: 0,
		},
		{
			name:        "SingleRequest",
			initial:     []string{"request1"},
			initialPos:  1,
			expected:    "request1",
			expectedPos: 0,
		},
		{
			name:        "MultipleRequests",
			initial:     []string{"request1", "request2", "request3"},
			initialPos:  3,
			expected:    "request3",
			expectedPos: 2,
		},
		{
			name:        "AtStartOfList",
			initial:     []string{"request1", "request2"},
			initialPos:  0,
			expected:    "",
			expectedPos: 0,
		},
		{
			name:        "AfterResetPosition",
			initial:     []string{"request1", "request2"},
			initialPos:  2,
			expected:    "request2",
			expectedPos: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := &History{
				requests: tt.initial,
				pos:      tt.initialPos,
			}

			got := history.PrevRequest()
			assert.Equal(t, tt.expected, got, "unexpected previous request")
			assert.Equal(t, tt.expectedPos, history.pos, "unexpected position after PrevRequest")
		})
	}
}

func TestResetPosition(t *testing.T) {
	tests := []struct {
		name        string
		initial     []string
		initialPos  int
		expectedPos int
	}{
		{
			name:        "EmptyHistory",
			initial:     []string{},
			initialPos:  0,
			expectedPos: 0,
		},
		{
			name:        "ResetFromMiddle",
			initial:     []string{"request1", "request2", "request3"},
			initialPos:  1,
			expectedPos: 3,
		},
		{
			name:        "ResetFromEnd",
			initial:     []string{"request1", "request2", "request3"},
			initialPos:  3,
			expectedPos: 3,
		},
		{
			name:        "ResetFromBeginning",
			initial:     []string{"request1", "request2", "request3"},
			initialPos:  0,
			expectedPos: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := &History{
				requests: tt.initial,
				pos:      tt.initialPos,
			}

			history.ResetPosition()
			assert.Equal(t, tt.expectedPos, history.pos, "unexpected position after ResetPosition")
		})
	}
}

func TestParseWordsFromRequest(t *testing.T) {
	tests := []struct {
		name     string
		request  string
		expected []string
	}{
		{
			name:     "EmptyRequest",
			request:  "",
			expected: []string(nil),
		},
		{
			name:     "SingleWord",
			request:  "hello",
			expected: []string{"hello"},
		},
		{
			name:     "MultipleWords",
			request:  "hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "WordsWithSpecialCharacters",
			request:  "hello, world!",
			expected: []string{"hello", "world"},
		},
		{
			name:     "MixedCaseWords",
			request:  "Hello WoRLd",
			expected: []string{"Hello", "WoRLd"},
		},
		{
			name:     "WordsWithNumbers",
			request:  "hello123 world456",
			expected: []string{"hello123", "world456"},
		},
		{
			name:     "PunctuationOnly",
			request:  ".,!?;",
			expected: []string(nil),
		},
		{
			name:     "WordsWithUnderscores",
			request:  "hello_world test_case",
			expected: []string{"hello_world", "test_case"},
		},
		{
			name:     "WordsWithHyphens",
			request:  "co-operate re-enter",
			expected: []string{"co-operate", "re-enter"},
		},
		{
			name:     "WhitespaceOnly",
			request:  "   ",
			expected: []string(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseWordsFromRequest(tt.request)
			assert.Equal(t, tt.expected, result, "unexpected result for request: %s", tt.request)
		})
	}
}

func TestHistory_AddWordsToIndex(t *testing.T) {
	index := NewDictionary(nil)
	history := &History{
		index: index,
	}

	expectedWords := []string{"hello", "world"}

	history.AddWordsToIndex(expectedWords)
	assert.Equal(t, index.words, expectedWords, "unexpected index state after AddWordsToIndex")
}

func TestHistory_Search(t *testing.T) {
	index := NewDictionary([]string{"hello"})
	history := &History{
		index: index,
	}

	word := history.Search("hell")

	assert.Equal(t, "hello", word, "unexpected word")
}
