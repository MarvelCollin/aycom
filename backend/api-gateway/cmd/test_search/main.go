package main

import (
	"fmt"
	"strings"
)

// DamerauLevenshteinDistance calculates the minimum edit distance between two strings
// using the Damerau-Levenshtein algorithm, which considers insertions, deletions,
// substitutions, and transpositions (swapping adjacent characters).
func DamerauLevenshteinDistance(s1, s2 string) int {
	// Handle edge cases
	if s1 == s2 {
		return 0
	}
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Initialize the matrix with the proper dimensions
	d := make([][]int, len(s1)+1)
	for i := range d {
		d[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		d[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		d[0][j] = j
	}

	// Fill in the rest of the matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			// Regular Levenshtein operations: insertion, deletion, substitution
			d[i][j] = min(
				d[i-1][j]+1,      // deletion
				d[i][j-1]+1,      // insertion
				d[i-1][j-1]+cost, // substitution
			)

			// Damerau extension: transposition
			if i > 1 && j > 1 && s1[i-1] == s2[j-2] && s1[i-2] == s2[j-1] {
				d[i][j] = min(d[i][j], d[i-2][j-2]+cost)
			}
		}
	}

	return d[len(s1)][len(s2)]
}

// DamerauLevenshteinSimilarity calculates a similarity percentage between two strings
// based on the Damerau-Levenshtein distance. Result is between 0.0 and 1.0,
// where 1.0 means the strings are identical.
func DamerauLevenshteinSimilarity(s1, s2 string) float64 {
	maxLen := max(float64(len(s1)), float64(len(s2)))
	if maxLen == 0 {
		return 1.0 // Both strings are empty, so they're identical
	}

	distance := DamerauLevenshteinDistance(s1, s2)
	similarity := 1.0 - float64(distance)/maxLen
	return similarity
}

// IsFuzzyMatch determines if two strings match based on a minimum similarity threshold
func IsFuzzyMatch(s1, s2 string, threshold float64) bool {
	similarity := DamerauLevenshteinSimilarity(s1, s2)
	return similarity >= threshold
}

// min returns the minimum of multiple integers
func min(values ...int) int {
	minValue := values[0]
	for _, v := range values[1:] {
		if v < minValue {
			minValue = v
		}
	}
	return minValue
}

// max returns the maximum of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func main() {
	// Test cases from the user
	searchText := "kolnb"
	actualName := "kolinb"

	// Test with different thresholds
	thresholds := []float64{0.5, 0.6, 0.7, 0.8, 0.9}

	fmt.Printf("Comparing '%s' with '%s':\n", searchText, actualName)
	fmt.Printf("Distance: %d\n", DamerauLevenshteinDistance(searchText, actualName))
	similarity := DamerauLevenshteinSimilarity(searchText, actualName)
	fmt.Printf("Similarity: %.2f (%.0f%%)\n", similarity, similarity*100)

	fmt.Println("\nMatching with different thresholds:")
	for _, threshold := range thresholds {
		match := IsFuzzyMatch(strings.ToLower(searchText), strings.ToLower(actualName), threshold)
		fmt.Printf("  Threshold %.1f: %v\n", threshold, match)
	}

	// Test a few more examples
	testCases := []struct {
		search string
		actual string
	}{
		{"jonathon", "jonathan"},
		{"christphir", "christopher"},
		{"amenda", "amanda"},
		{"mickel", "michael"},
	}

	fmt.Println("\nAdditional test cases:")
	for _, tc := range testCases {
		similarity := DamerauLevenshteinSimilarity(strings.ToLower(tc.search), strings.ToLower(tc.actual))
		match := IsFuzzyMatch(strings.ToLower(tc.search), strings.ToLower(tc.actual), 0.6)
		fmt.Printf("  '%s' vs '%s': %.2f (%.0f%%) - Match at 0.6 threshold: %v\n",
			tc.search, tc.actual, similarity, similarity*100, match)
	}
}
