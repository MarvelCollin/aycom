package utils

import (
	"testing"
)

func TestDamerauLevenshteinDistance(t *testing.T) {
	testCases := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},              // Empty strings
		{"abc", "abc", 0},        // Identical
		{"abcd", "abc", 1},       // Deletion
		{"abc", "abcd", 1},       // Insertion
		{"abc", "abd", 1},        // Substitution
		{"abcd", "acbd", 1},      // Transposition
		{"kitten", "sitting", 3}, // Multiple operations
		{"kolnb", "kolinb", 1},   // Testing the example from user query
		{"kolinb", "kolnb", 1},   // Testing the example in reverse
		{"apple", "apel", 2},     // Multiple operations
	}

	for _, tc := range testCases {
		result := DamerauLevenshteinDistance(tc.s1, tc.s2)
		if result != tc.expected {
			t.Errorf("DamerauLevenshteinDistance(%q, %q): expected %d, got %d",
				tc.s1, tc.s2, tc.expected, result)
		}
	}
}

func TestDamerauLevenshteinSimilarity(t *testing.T) {
	testCases := []struct {
		s1       string
		s2       string
		expected float64
		delta    float64 // Acceptable error margin
	}{
		{"", "", 1.0, 0.001},
		{"abc", "abc", 1.0, 0.001},
		{"abcd", "abc", 0.75, 0.001},
		{"kolnb", "kolinb", 1.0 - 1.0/6.0, 0.001}, // The example case, with ~0.83 similarity
		{"kolinb", "kolnb", 1.0 - 1.0/6.0, 0.001}, // The example case in reverse
	}

	for _, tc := range testCases {
		result := DamerauLevenshteinSimilarity(tc.s1, tc.s2)
		if abs(result-tc.expected) > tc.delta {
			t.Errorf("DamerauLevenshteinSimilarity(%q, %q): expected %.4f, got %.4f",
				tc.s1, tc.s2, tc.expected, result)
		}
	}
}

func TestIsFuzzyMatch(t *testing.T) {
	testCases := []struct {
		s1        string
		s2        string
		threshold float64
		expected  bool
	}{
		{"kolnb", "kolinb", 0.6, true},    // Should match with threshold 0.6
		{"kolnb", "kolinb", 0.9, false},   // Should not match with threshold 0.9
		{"abcdef", "abcxyz", 0.5, false},  // Less than 50% similarity
		{"kitten", "sitting", 0.5, false}, // Less than 50% similarity
		{"apel", "apple", 0.5, true},      // More than 50% similarity
	}

	for _, tc := range testCases {
		result := IsFuzzyMatch(tc.s1, tc.s2, tc.threshold)
		if result != tc.expected {
			t.Errorf("IsFuzzyMatch(%q, %q, %.2f): expected %v, got %v",
				tc.s1, tc.s2, tc.threshold, tc.expected, result)
		}
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
