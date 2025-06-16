package test

import (
	"aycom/backend/api-gateway/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDamerauLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{"empty_strings", "", "", 0},
		{"identical_strings", "hello", "hello", 0},
		{"deletion", "hello", "hell", 1},
		{"insertion", "hell", "hello", 1},
		{"substitution", "hello", "hallo", 1},
		{"transposition", "hello", "hlelo", 1},
		{"combined_operations", "kitten", "sitting", 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			result := utils.DamerauLevenshteinDistance(test.s1, test.s2)
			assert.Equal(test.expected, result, "Distance calculation failed for case %s", test.name)
		})
	}
}

func TestDamerauLevenshteinSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected float64
		delta    float64
	}{
		{"empty_strings", "", "", 1.0, 0.001},
		{"identical_strings", "hello", "hello", 1.0, 0.001},
		{"one_change", "hello", "hallo", 0.8, 0.001},
		{"username_comparison", "kolnb", "kolinb", 0.833, 0.001},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			result := utils.DamerauLevenshteinSimilarity(test.s1, test.s2)
			assert.InDelta(test.expected, result, test.delta, "Similarity calculation failed for case %s", test.name)
		})
	}
}

func TestIsFuzzyMatch(t *testing.T) {
	tests := []struct {
		name      string
		s1        string
		s2        string
		threshold float64
		expected  bool
	}{
		{"high_similarity_above_threshold", "hello", "hallo", 0.7, true},
		{"medium_similarity_below_threshold", "hello", "halo", 0.8, false},
		{"username_match", "kolnb", "kolinb", 0.8, true},
		{"username_no_match", "kolnb", "kolinb", 0.9, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			result := utils.IsFuzzyMatch(test.s1, test.s2, test.threshold)
			assert.Equal(test.expected, result, "Fuzzy match failed for case %s", test.name)
		})
	}
}
