package edit

import (
	"sort"
	"testing"
)

func TestNewDictionary(t *testing.T) {
	words := []string{"flower", "flow", "flight"}
	expected := []string{"flight", "flow", "flower"}

	dictionary := NewDictionary(words)

	// Check if the words are sorted in ascending order
	if !sort.StringsAreSorted(dictionary.words) {
		t.Errorf("NewDictionary did not sort the words in ascending order")
	}

	// Check if the sorted words match the expected result
	for i, word := range dictionary.words {
		if word != expected[i] {
			t.Errorf("NewDictionary sorted words do not match the expected result")
			break
		}
	}
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
