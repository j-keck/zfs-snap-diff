// +build !linux

package zfs

import (
	"fmt"
	"os/exec"
	"strings"
)

func findmnt(name string) (string, error) {
	log.Tracef("findmnt (unix) for '%s'", name)
	out, err := exec.Command(
		"mount", "-l", "-t", "zfs",
	).Output()
	if err != nil {
		log.Tracef("unable to lookup name for: '%s' - err: %v", name, err)
		return "", err
	}

	lookup := func(name, s string) (string, bool) {
		const n = 4
		for _, line := range strings.Split(s, "\n") {
			fields := strings.SplitN(line, "\t", n)
			if len(fields) == n {
				if fields[0] == name {
					return fields[2], true
				}
			} else {
				log.Debugf("ignore invalid formatted line: '%s", s)
			}
		}
		return "", false
	}

	if path, ok := lookup(name, string(out)); ok {
		return path, nil
	} else {
		return "", fmt.Errorf("mountpoint for '%s' not found", name)
	}
}
