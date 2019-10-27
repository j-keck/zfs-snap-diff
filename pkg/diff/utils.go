package diff

import (
	"bytes"
)

// min for int's
func min(i1, i2 int) int {
	if i1 < i2 {
		return i1
	}
	return i2
}

// max for int's
func max(i1, i2 int) int {
	if i1 > i2 {
		return i1
	}
	return i2
}

func countNewLines(text string) int {
	count := 0
	for _, char := range text {
		if char == '\n' {
			count++
		}
	}
	return count
}

//
// string.Split is not usable for this purpose:
//   * splits a text and removes the seperator
//   * if the last element has a \n, this is added
//     as a extra element
//
func splitText(text string) []string {
	var lines []string

	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			lines = append(lines, text[start:i+1])
			start = i + 1
		}
	}
	// add last element if one is pending
	// (last element was without a \n)
	if start < len(text) {
		lines = append(lines, text[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	var buf bytes.Buffer
	for _, line := range lines {
		buf.WriteString(line)
	}
	return buf.String()
}
