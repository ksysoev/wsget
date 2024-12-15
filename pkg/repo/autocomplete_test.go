package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDictionary(t *testing.T) {
	words := []string{"flower", "flow", "flight"}
	expected := []string{"flight", "flow", "flower"}

	dictionary := NewDictionary(words)

	// Check if the sorted words match the expected result
	assert.Equal(t, expected, dictionary.words)
}

func TestDictionary_Search(t *testing.T) {
	words := []string{"flower", "flow", "flight"}
	dictionary := NewDictionary(words)

	tests := []struct {
		prefix   string
		expected string
	}{
		{
			prefix:   "fl",
			expected: "fl",
		},
		{
			prefix:   "flo",
			expected: "flow",
		},
		{
			prefix:   "flow",
			expected: "flow",
		},
		{
			prefix:   "flower",
			expected: "flower",
		},
		{
			prefix:   "abc",
			expected: "",
		},
		{
			prefix:   "xyz",
			expected: "",
		},
	}

	for _, tt := range tests {
		result := dictionary.Search(tt.prefix)
		if result != tt.expected {
			t.Errorf("Search(%s) = %s, expected %s", tt.prefix, result, tt.expected)
		}
	}
}

func TestLongestCommonPrefix(t *testing.T) {
	tests := []struct {
		expected string
		strs     []string
	}{
		{
			strs:     []string{"flower", "flow", "flight"},
			expected: "fl",
		},
		{
			strs:     []string{"dog", "racecar", "car"},
			expected: "",
		},
		{
			strs:     []string{"apple", "app", "application"},
			expected: "app",
		},
		{
			strs:     []string{"", "abc", "def"},
			expected: "",
		},
		{
			strs:     []string{"abc", "abc", "abc"},
			expected: "abc",
		},
	}

	for _, tt := range tests {
		result := longestCommonPrefix(tt.strs)
		if result != tt.expected {
			t.Errorf("longestCommonPrefix(%v) = %s, expected %s", tt.strs, result, tt.expected)
		}
	}
}

func TestDictionary_AddWords(t *testing.T) {
	tests := []struct {
		name          string
		initialWords  []string
		newWords      []string
		expectedWords []string
	}{
		{
			name:          "Add new non-duplicate words",
			initialWords:  []string{"flower", "flow"},
			newWords:      []string{"flight", "flame"},
			expectedWords: []string{"flame", "flight", "flow", "flower"},
		},
		{
			name:          "Add duplicate words",
			initialWords:  []string{"apple", "banana"},
			newWords:      []string{"banana", "cherry"},
			expectedWords: []string{"apple", "banana", "cherry"},
		},
		{
			name:          "Add no new words",
			initialWords:  []string{"dog", "cat"},
			newWords:      []string{},
			expectedWords: []string{"cat", "dog"},
		},
		{
			name:          "Add words to an empty dictionary",
			initialWords:  []string{},
			newWords:      []string{"zebra", "ant"},
			expectedWords: []string{"ant", "zebra"},
		},
		{
			name:          "Add empty and unique words",
			initialWords:  []string{},
			newWords:      []string{"", "abc"},
			expectedWords: []string{"", "abc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dictionary := NewDictionary(tt.initialWords)
			dictionary.AddWords(tt.newWords)
			assert.Equal(t, tt.expectedWords, dictionary.words)
		})
	}
}
