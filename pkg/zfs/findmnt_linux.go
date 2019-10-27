package zfs

import (
	"os/exec"
	"strings"
)

func findmnt(name string) (string, error) {
	log.Tracef("findmnt for '%s'", name)
	out, err := exec.Command(
		"findmnt", "-t", "zfs", "--first-only", "--noheadings",
		"--output", "target", "-S", name,
	).Output()
	if err != nil {
		log.Tracef("unable to lookup name for: '%s' - err: %v", name, err)
		return "", err
	} else {
		path := strings.TrimSpace(string(out))
		log.Tracef("findmnt found path: '%s'", path)
		return path, nil
	}
}
