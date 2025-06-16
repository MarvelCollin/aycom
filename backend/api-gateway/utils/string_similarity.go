package utils

import "math"

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
	maxLen := math.Max(float64(len(s1)), float64(len(s2)))
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