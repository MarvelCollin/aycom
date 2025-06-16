package main

import (
	"fmt"
	"strings"
)

func DamerauLevenshteinDistance(s1, s2 string) int {

	if s1 == s2 {
		return 0
	}
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	d := make([][]int, len(s1)+1)
	for i := range d {
		d[i] = make([]int, len(s2)+1)
	}

	for i := 0; i <= len(s1); i++ {
		d[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		d[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			d[i][j] = min(
				d[i-1][j]+1,
				d[i][j-1]+1,
				d[i-1][j-1]+cost,
			)

			if i > 1 && j > 1 && s1[i-1] == s2[j-2] && s1[i-2] == s2[j-1] {
				d[i][j] = min(d[i][j], d[i-2][j-2]+cost)
			}
		}
	}

	return d[len(s1)][len(s2)]
}

func DamerauLevenshteinSimilarity(s1, s2 string) float64 {
	maxLen := max(float64(len(s1)), float64(len(s2)))
	if maxLen == 0 {
		return 1.0
	}

	distance := DamerauLevenshteinDistance(s1, s2)
	similarity := 1.0 - float64(distance)/maxLen
	return similarity
}

func IsFuzzyMatch(s1, s2 string, threshold float64) bool {
	similarity := DamerauLevenshteinSimilarity(s1, s2)
	return similarity >= threshold
}

func min(values ...int) int {
	minValue := values[0]
	for _, v := range values[1:] {
		if v < minValue {
			minValue = v
		}
	}
	return minValue
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func main() {
	fmt.Println("**** FUZZY SEARCH DEBUGGING TOOL ****")

	searchQueries := []string{"kolnb", "koln", "kolb"}
	actualNames := []string{"kolinb", "kolin", "kolina", "kolingg", "kolinf", "kolinc"}

	fmt.Println("\n== Critical Test Case ==")

	criticalSearch := "kolnb"
	criticalName := "kolinb"
	fmt.Printf("Comparing '%s' with '%s':\n", criticalSearch, criticalName)
	fmt.Printf("Distance: %d\n", DamerauLevenshteinDistance(criticalSearch, criticalName))
	similarity := DamerauLevenshteinSimilarity(criticalSearch, criticalName)
	fmt.Printf("Similarity: %.4f (%.1f%%)\n", similarity, similarity*100)

	directMatch := strings.Contains(strings.ToLower(criticalName), strings.ToLower(criticalSearch))
	fmt.Printf("Direct substring match: %v\n", directMatch)

	reverseMatch := strings.Contains(strings.ToLower(criticalSearch), strings.ToLower(criticalName))
	fmt.Printf("Reverse substring match: %v\n", reverseMatch)

	fmt.Println("\nMatching with different thresholds:")
	thresholds := []float64{0.3, 0.5, 0.6, 0.7, 0.8, 0.9}
	for _, threshold := range thresholds {
		match := IsFuzzyMatch(strings.ToLower(criticalSearch), strings.ToLower(criticalName), threshold)
		fmt.Printf("  Threshold %.1f: %v\n", threshold, match)
	}

	fmt.Println("\n== All Possible Test Combinations ==")
	for _, query := range searchQueries {
		for _, name := range actualNames {
			similarity := DamerauLevenshteinSimilarity(strings.ToLower(query), strings.ToLower(name))
			match := IsFuzzyMatch(strings.ToLower(query), strings.ToLower(name), 0.3)
			directMatch := strings.Contains(strings.ToLower(name), strings.ToLower(query))

			fmt.Printf("Query '%s' vs Name '%s': %.4f (%.1f%%) - Match at 0.3 threshold: %v, Direct contains: %v\n",
				query, name, similarity, similarity*100, match, directMatch)
		}
	}

	fmt.Println("\n== Additional Test Cases ==")
	testCases := []struct {
		search string
		actual string
	}{
		{"jonathon", "jonathan"},
		{"christphir", "christopher"},
		{"amenda", "amanda"},
		{"mickel", "michael"},
	}

	for _, tc := range testCases {
		similarity := DamerauLevenshteinSimilarity(strings.ToLower(tc.search), strings.ToLower(tc.actual))
		match := IsFuzzyMatch(strings.ToLower(tc.search), strings.ToLower(tc.actual), 0.3)
		directMatch := strings.Contains(strings.ToLower(tc.actual), strings.ToLower(tc.search))

		fmt.Printf("'%s' vs '%s': %.4f (%.1f%%) - Match at 0.3: %v, Direct contains: %v\n",
			tc.search, tc.actual, similarity, similarity*100, match, directMatch)
	}
}
