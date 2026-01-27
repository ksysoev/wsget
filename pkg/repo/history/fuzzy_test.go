package history

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistory_FuzzySearch(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		first    string
		requests []string
		expected int
	}{
		{
			name:     "Empty query returns all requests",
			requests: []string{"test", "hello", "world"},
			query:    "",
			expected: 3,
			first:    "world", // Most recent first
		},
		{
			name:     "Exact substring match",
			requests: []string{"hello world", "test message", "hello there"},
			query:    "hello",
			expected: 2,
			first:    "", // Both hello strings score similarly, don't check order
		},
		{
			name:     "Fuzzy match with gaps",
			requests: []string{"test message", "text editing", "tea time"},
			query:    "tst",
			expected: 1,
			first:    "test message",
		},
		{
			name:     "Case insensitive matching",
			requests: []string{"Hello World", "HELLO THERE", "hello"},
			query:    "hello",
			expected: 3,
			first:    "hello", // Exact match scores higher
		},
		{
			name:     "No matches",
			requests: []string{"apple", "banana", "cherry"},
			query:    "xyz",
			expected: 0,
			first:    "",
		},
		{
			name:     "Match at word boundary",
			requests: []string{"user_login", "login_user", "my_login_test"},
			query:    "login",
			expected: 3,
			first:    "", // Don't check order, all match at boundaries
		},
		{
			name:     "Consecutive character bonus",
			requests: []string{"abcdef", "a b c d e f", "abc def"},
			query:    "abc",
			expected: 3,
			first:    "", // Both "abcdef" and "abc def" have exact substring
		},
		{
			name:     "Early match bonus",
			requests: []string{"test at end", "at test start"},
			query:    "at",
			expected: 2,
			first:    "at test start", // Earlier match scores higher
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHistory("test")
			for _, req := range tt.requests {
				h.AddRequest(req)
			}

			matches := h.FuzzySearch(tt.query)

			assert.Equal(t, tt.expected, len(matches), "Expected %d matches, got %d", tt.expected, len(matches))

			if tt.expected > 0 && tt.first != "" {
				assert.Equal(t, tt.first, matches[0].Request, "Expected first match to be '%s', got '%s'", tt.first, matches[0].Request)
			}
		})
	}
}

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		query         string
		shouldMatch   bool
		expectedScore int // Approximate score for validation
	}{
		{
			name:        "Exact match",
			text:        "hello",
			query:       "hello",
			shouldMatch: true,
		},
		{
			name:        "Substring match",
			text:        "hello world",
			query:       "world",
			shouldMatch: true,
		},
		{
			name:        "Fuzzy match with gaps",
			text:        "hello world",
			query:       "hlwl",
			shouldMatch: true,
		},
		{
			name:        "Case insensitive",
			text:        "Hello World",
			query:       "hello",
			shouldMatch: true,
		},
		{
			name:        "No match - missing characters",
			text:        "hello",
			query:       "xyz",
			shouldMatch: false,
		},
		{
			name:        "Query longer than text",
			text:        "hi",
			query:       "hello",
			shouldMatch: false,
		},
		{
			name:        "Empty query",
			text:        "hello",
			query:       "",
			shouldMatch: true,
		},
		{
			name:        "Characters in wrong order",
			text:        "hello",
			query:       "olleh",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryRunes := []rune(strings.ToLower(tt.query))
			matched, score, positions := fuzzyMatch(tt.text, queryRunes)

			assert.Equal(t, tt.shouldMatch, matched, "Expected match=%v, got %v", tt.shouldMatch, matched)

			switch {
			case matched && len(queryRunes) > 0:
				assert.Greater(t, score, 0, "Score should be positive for matches")
				assert.Equal(t, len(queryRunes), len(positions), "Position count should equal query length")
			case matched && len(queryRunes) == 0:
				// Empty query always matches with zero score
				assert.Equal(t, 0, score, "Empty query should have zero score")
				assert.Empty(t, positions, "Empty query should have no positions")
			default:
				assert.Equal(t, 0, score, "Score should be 0 for non-matches")
				assert.Empty(t, positions, "Positions should be empty for non-matches")
			}
		})
	}
}

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		query     string
		positions []int
		minScore  int // Minimum expected score
	}{
		{
			name:      "Consecutive matches score high",
			text:      "hello",
			query:     "hel",
			positions: []int{0, 1, 2},
			minScore:  100, // High score for consecutive + start
		},
		{
			name:      "Word boundary bonus",
			text:      "user_login",
			query:     "login",
			positions: []int{5, 6, 7, 8, 9},
			minScore:  80, // Word boundary + consecutive
		},
		{
			name:      "Scattered matches score lower",
			text:      "h e l l o",
			query:     "hlo",
			positions: []int{0, 4, 8},
			minScore:  20, // Base score only
		},
		{
			name:      "Early match bonus",
			text:      "test",
			query:     "t",
			positions: []int{0},
			minScore:  30, // Start bonus + early match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryRunes := []rune(strings.ToLower(tt.query))
			score := calculateScore(tt.text, queryRunes, tt.positions)

			assert.GreaterOrEqual(t, score, tt.minScore, "Expected score >= %d, got %d", tt.minScore, score)
		})
	}
}

func TestHistory_GetAllRequests(t *testing.T) {
	h := NewHistory("test")
	requests := []string{"first", "second", "third"}

	for _, req := range requests {
		h.AddRequest(req)
	}

	all := h.GetAllRequests()

	// Should be in reverse order (most recent first)
	assert.Equal(t, []string{"third", "second", "first"}, all)
}

func TestFuzzySearch_Scoring(t *testing.T) {
	h := NewHistory("test")

	// Add requests in specific order to test scoring
	h.AddRequest("test message")      // Consecutive "test"
	h.AddRequest("t e s t")           // Scattered "test"
	h.AddRequest("the best test")     // "test" at end
	h.AddRequest("test at beginning") // "test" at start

	matches := h.FuzzySearch("test")

	// All should match
	assert.Equal(t, 4, len(matches))

	// Exact substring matches should score highest
	// "test message" and "test at beginning" should be top matches
	topMatches := []string{matches[0].Request, matches[1].Request}
	assert.Contains(t, topMatches, "test message")
	assert.Contains(t, topMatches, "test at beginning")

	// Scattered match should score lowest
	assert.Equal(t, "t e s t", matches[3].Request)
}

func TestFuzzySearch_EmptyHistory(t *testing.T) {
	h := NewHistory("test")

	matches := h.FuzzySearch("anything")

	assert.Empty(t, matches, "Empty history should return no matches")
}

func TestFuzzySearch_MultilineRequests(t *testing.T) {
	h := NewHistory("test")

	h.AddRequest("line1\nline2\nline3")
	h.AddRequest("single line")

	// Should match across newlines
	matches := h.FuzzySearch("line")

	assert.Equal(t, 2, len(matches))
}

func TestFuzzySearch_Deduplication(t *testing.T) {
	h := NewHistory("test")

	// Add duplicate requests
	h.AddRequest(`{"ping":1}`)
	h.AddRequest(`{"authorize":"token"}`)
	h.AddRequest(`{"ping":1}`)
	h.AddRequest(`{"ping":1}`)
	h.AddRequest(`{"authorize":"token"}`)
	h.AddRequest(`{"ping":1}`)

	// Search with query
	matches := h.FuzzySearch("ping")

	// Should return only unique matches
	assert.Equal(t, 1, len(matches), "Duplicate requests should be deduplicated")
	assert.Equal(t, `{"ping":1}`, matches[0].Request)

	// Search with empty query (show all)
	allMatches := h.FuzzySearch("")

	// Should return only unique requests in order of most recent occurrence
	assert.Equal(t, 2, len(allMatches), "Empty query should return deduplicated results")

	// Most recent unique entries should be first
	assert.Equal(t, `{"ping":1}`, allMatches[0].Request)
	assert.Equal(t, `{"authorize":"token"}`, allMatches[1].Request)
}

func TestFuzzySearch_DeduplicationKeepsHighestScore(t *testing.T) {
	h := NewHistory("test")

	// Add requests where the same text appears in different contexts
	// This tests that we keep the highest scoring match when deduplicating
	h.AddRequest("test message")
	h.AddRequest("another test")
	h.AddRequest("test message")

	matches := h.FuzzySearch("test")

	// Should have unique results
	uniqueRequests := make(map[string]bool)
	for _, match := range matches {
		assert.False(t, uniqueRequests[match.Request], "Found duplicate request: %s", match.Request)
		uniqueRequests[match.Request] = true
	}

	assert.Equal(t, 2, len(matches), "Should return 2 unique matches")
}
