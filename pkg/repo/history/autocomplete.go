package history

import (
	"sort"
	"strings"
)

type Dictionary struct {
	words []string
}

// NewDictionary creates a new instance of Dictionary with the given list of words.
// The words are sorted in ascending order before being stored in the dictionary.
// It returns a pointer to the created Dictionary.
func NewDictionary(words []string) *Dictionary {
	d := Dictionary{
		words: make([]string, 0, len(words)),
	}

	d.AddWords(words)

	return &d
}

// AddWords adds new unique words to the dictionary in sorted order.
// It takes a slice of strings newWords, representing the words to be added.
// It does not return any value and ignores duplicate or empty inputs.
func (d *Dictionary) AddWords(newWords []string) {
	if len(newWords) == 0 {
		return
	}

	wordSet := make(map[string]bool, len(d.words)+len(newWords))

	for _, word := range d.words {
		wordSet[word] = true
	}

	for _, word := range newWords {
		if !wordSet[word] {
			d.words = append(d.words, word)
			wordSet[word] = true
		}
	}

	sort.Strings(d.words)
}

// Search searches for words in the dictionary that have the given prefix.
// It performs a search to find all matching words.
// The function returns the longest common prefix among the matching words.
func (d *Dictionary) Search(prefix string) string {
	startPos := sort.Search(len(d.words), func(i int) bool {
		return d.words[i] >= prefix
	})

	match := make([]string, 0)

	for i := startPos; i < len(d.words); i++ {
		word := d.words[i]
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

// longestCommonPrefix finds the longest common prefix among an array of strings.
// It returns the longest common prefix found or an empty string if there is no common prefix.
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
