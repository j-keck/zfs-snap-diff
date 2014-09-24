package main

import (
	"os"
	"strings"
)

// lastElement splits a string by sep and returns the last element
func lastElement(str, sep string) string {
	fields := strings.Split(str, sep)
	return fields[len(fields)-1]
}

// envHasSet returns true, if 'key' is in the environment
func envHasSet(key string) bool {
	return len(os.Getenv(key)) > 0
}
