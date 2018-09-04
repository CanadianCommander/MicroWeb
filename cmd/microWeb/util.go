package main

import (
	"math"
)

/*
StringMatchLength calculates the langth of the match between to strings and returns it
*/
func StringMatchLength(s1, s2 string) int {
	i := 0
	for ; i < int(math.Min(float64(len(s1)), float64(len(s2)))); i++ {
		if s1[i] != s2[i] {
			break
		}
	}

	return i
}
