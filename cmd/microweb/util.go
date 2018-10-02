package main

import (
	"bufio"
	"math"
	"os"
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

/*
ReadFileLine reads one line from the given file and returns it, not including the "\n" character
*/
func ReadFileLine(file *os.File) (string, error) {
	lineBuffer := bufio.NewReader(file)
	lineB, _, err := (lineBuffer.ReadLine())
	return string(lineB), err
}
