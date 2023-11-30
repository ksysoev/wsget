package edit

import (
	"sort"
	"strings"
)

type Dictionary struct {
	words []string
}

func NewDictionary(words []string) *Dictionary {
	sortedWords := make([]string, len(words))
	copy(sortedWords, words)
	sort.Strings(sortedWords)

	return &Dictionary{
		words: sortedWords,
	}
}

func (d *Dictionary) Search(prefix string) string {
	// For now, we'll just do a linear search.
	// later, we can implement a binary search to find the first match.
	match := []string{}

	for _, word := range d.words {
		if len(prefix) > len(word) {
			continue
		}

		if word[:len(prefix)] == prefix {
			match = append(match, word)
		} else if len(match) > 0 {
			break
		}
	}

	return longestCommonPrefix(match)
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	prefix := strs[0]
	for i := 1; i < len(strs); i++ {
		for strings.Index(strs[i], prefix) != 0 {
			if prefix == "" {
				return ""
			}

			prefix = prefix[:len(prefix)-1]
		}
	}

	return prefix
}
