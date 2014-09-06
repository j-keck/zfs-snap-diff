package main

import (
	"strings"
)

type Snapshot struct {
	Name     string
	Creation string
}

type Snapshots []Snapshot

func (s *Snapshots) addFromZfsOutput(line string) {
	fields := strings.SplitN(line, "\t", 2)
	*s = append(*s, Snapshot{lastElement(fields[0], "@"), fields[1]})
}
