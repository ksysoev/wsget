package history

import (
	"sort"
	"strings"
	"unicode"
)

// FuzzyMatch represents a search result with its score and matched positions.
type FuzzyMatch struct {
	Request   string
	Positions []int // Character positions that matched
	Score     int
}

// FuzzySearch performs fuzzy matching on history requests and returns matches sorted by score.
// It takes a query string and returns a slice of FuzzyMatch results.
// The scoring algorithm prioritizes:
// - Consecutive character matches (higher score)
// - Matches at word boundaries (higher score)
// - Earlier matches in the string (higher score)
// - Case-insensitive matching
func (h *History) FuzzySearch(query string) []FuzzyMatch {
	if query == "" {
		// Return all requests in reverse order (most recent first)
		matches := make([]FuzzyMatch, 0, len(h.requests))
		for i := len(h.requests) - 1; i >= 0; i-- {
			matches = append(matches, FuzzyMatch{
				Request:   h.requests[i],
				Positions: nil,
				Score:     0,
			})
		}

		return matches
	}

	queryLower := strings.ToLower(query)
	queryRunes := []rune(queryLower)

	matches := make([]FuzzyMatch, 0)

	for _, req := range h.requests {
		if match, score, positions := fuzzyMatch(req, queryRunes); match {
			matches = append(matches, FuzzyMatch{
				Request:   req,
				Positions: positions,
				Score:     score,
			})
		}
	}

	// Sort by score (descending), then by position in history (most recent first)
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}
		// If scores are equal, prefer more recent entries
		return i > j
	})

	return matches
}

// fuzzyMatch checks if the query matches the text and calculates a score.
// It returns whether there's a match, the score, and the positions of matched characters.
func fuzzyMatch(text string, queryRunes []rune) (matched bool, score int, positions []int) {
	textLower := strings.ToLower(text)
	textRunes := []rune(textLower)

	if len(queryRunes) == 0 {
		return true, 0, nil
	}

	if len(textRunes) < len(queryRunes) {
		return false, 0, nil
	}

	// Try to find all query characters in order
	positions = make([]int, 0, len(queryRunes))
	textIdx := 0
	queryIdx := 0

	for textIdx < len(textRunes) && queryIdx < len(queryRunes) {
		if textRunes[textIdx] == queryRunes[queryIdx] {
			positions = append(positions, textIdx)
			queryIdx++
		}

		textIdx++
	}

	// Check if all query characters were found
	if queryIdx < len(queryRunes) {
		return false, 0, nil
	}

	// Calculate score based on match quality
	score = calculateScore(text, queryRunes, positions)

	return true, score, positions
}

// calculateScore computes the match quality score based on various factors.
func calculateScore(text string, queryRunes []rune, positions []int) int {
	if len(positions) == 0 {
		return 0
	}

	score := 0
	textRunes := []rune(text)

	// Base score: number of matched characters
	score += len(positions) * 10 //nolint:mnd // Base score per matched character

	// Check for exact substring match first (highest priority)
	queryStr := string(queryRunes)
	if strings.Contains(strings.ToLower(text), queryStr) {
		score += 200 //nolint:mnd // High bonus for exact substring match
	}

	// Count consecutive matches
	consecutiveCount := 0
	maxConsecutive := 1
	currentConsecutive := 1

	for i := 1; i < len(positions); i++ {
		if positions[i] == positions[i-1]+1 {
			currentConsecutive++
			consecutiveCount++

			if currentConsecutive > maxConsecutive {
				maxConsecutive = currentConsecutive
			}
		} else {
			currentConsecutive = 1
		}
	}

	// Bonus for consecutive matches (scales with length)
	score += consecutiveCount * 20 //nolint:mnd // Bonus per consecutive character
	score += maxConsecutive * 10   //nolint:mnd // Extra bonus for longest run

	// Bonus for matches at word boundaries
	for i, pos := range positions {
		if pos == 0 {
			// First character match
			score += 30 //nolint:mnd // Bonus for match at start
		} else if pos > 0 && pos < len(textRunes) {
			prevChar := textRunes[pos-1]
			if unicode.IsSpace(prevChar) || prevChar == '_' || prevChar == '-' || prevChar == '.' {
				// Match after delimiter
				score += 20 //nolint:mnd // Bonus for word boundary match
			}
		}

		// Check if this starts a consecutive sequence
		if i == 0 || positions[i] != positions[i-1]+1 {
			// This is the start of a new sequence
			if pos == 0 || (pos > 0 && (unicode.IsSpace(textRunes[pos-1]) || textRunes[pos-1] == '_' || textRunes[pos-1] == '-' || textRunes[pos-1] == '.')) {
				score += 15 //nolint:mnd // Bonus for sequence start at word boundary
			}
		}
	}

	// Penalty for distance between first and last match (but smaller impact)
	matchSpan := positions[len(positions)-1] - positions[0] + 1
	score -= matchSpan / 2 //nolint:mnd // Penalty for scattered matches

	// Bonus for early matches (prefer matches near the beginning)
	if positions[0] < 10 { //nolint:mnd // First 10 characters get early match bonus
		score += 15 - positions[0] //nolint:mnd // Decreasing bonus based on position
	}

	return score
}

// GetAllRequests returns all requests in reverse chronological order (most recent first).
func (h *History) GetAllRequests() []string {
	requests := make([]string, len(h.requests))
	for i, j := 0, len(h.requests)-1; j >= 0; i, j = i+1, j-1 {
		requests[i] = h.requests[j]
	}

	return requests
}
