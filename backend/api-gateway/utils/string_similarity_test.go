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
		{"", "", 0},              
		{"abc", "abc", 0},        
		{"abcd", "abc", 1},       
		{"abc", "abcd", 1},       
		{"abc", "abd", 1},        
		{"abcd", "acbd", 1},      
		{"kitten", "sitting", 3}, 
		{"kolnb", "kolinb", 1},   
		{"kolinb", "kolnb", 1},   
		{"apple", "apel", 2},     
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
		delta    float64 
	}{
		{"", "", 1.0, 0.001},
		{"abc", "abc", 1.0, 0.001},
		{"abcd", "abc", 0.75, 0.001},
		{"kolnb", "kolinb", 1.0 - 1.0/6.0, 0.001}, 
		{"kolinb", "kolnb", 1.0 - 1.0/6.0, 0.001}, 
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
		{"kolnb", "kolinb", 0.6, true},    
		{"kolnb", "kolinb", 0.9, false},   
		{"abcdef", "abcxyz", 0.5, false},  
		{"kitten", "sitting", 0.5, false}, 
		{"apel", "apple", 0.5, true},      
	}

	for _, tc := range testCases {
		result := IsFuzzyMatch(tc.s1, tc.s2, tc.threshold)
		if result != tc.expected {
			t.Errorf("IsFuzzyMatch(%q, %q, %.2f): expected %v, got %v",
				tc.s1, tc.s2, tc.threshold, tc.expected, result)
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}