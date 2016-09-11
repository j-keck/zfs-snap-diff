package main

import (
	"bytes"
	"os"
	"strings"
)

// read the file content from the file handle
func readTextFrom(getFh func(string) (*FileHandle, error), name string) (string, error) {

	// get the file handle
	fh, err := getFh(name)
	if err != nil {
		logError.Println("unable to get file-handle: ", err.Error())
		return "", err
	}

	// read the file content
	content, err := fh.ReadText()
	if err != nil {
		logError.Println("unable to read the file: ", err.Error())
		return "", err
	}

	return content, err
}

// lastElement splits a string by sep and returns the last element
func lastElement(str, sep string) string {
	fields := strings.Split(str, sep)
	return fields[len(fields)-1]
}

func firstElement(str, sep string) string {
	return strings.Split(str, sep)[0]
}

// split2 splits a given string by 'sep' into two elements
// returns the elements and a bool flag if both elements are found
func split2(str, sep string) (string, string, bool) {
	e := strings.SplitN(str, sep, 2)
	if len(e) != 2 {
		return "", "", false
	}
	return e[0], e[1], true
}

// split3 splits a given string by 'sep' into three elements
// returns the elements and a bool flag if all elements are found
func split3(str, sep string) (string, string, string, bool) {
	e := strings.SplitN(str, sep, 3)
	if len(e) != 3 {
		return "", "", "", false
	}
	return e[0], e[1], e[2], true
}

// split4 splits a given string by 'sep' into for elements
// returns the elements and a bool flag if all elements are found
func split4(str, sep string) (string, string, string, string, bool) {
	e := strings.SplitN(str, sep, 4)
	if len(e) != 4 {
		return "", "", "", "", false
	}
	return e[0], e[1], e[2], e[3], true
}

// split5 splits a given string by 'sep' into for elements
// returns the elements and a bool flag if all elements are found
func split5(str, sep string) (string, string, string, string, string, bool) {
	e := strings.SplitN(str, sep, 5)
	if len(e) != 5 {
		return "", "", "", "", "", false
	}
	return e[0], e[1], e[2], e[3], e[4], true
}

// envHasSet returns true, if 'key' is in the environment
func envHasSet(key string) bool {
	return len(os.Getenv(key)) > 0
}

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
