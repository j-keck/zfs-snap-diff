package main

import (
	"bytes"
	"os/exec"
	"strings"
)

type ZFS struct {
	Name       string
	MountPoint string
	execZFS    execZFSFunc
}

func NewZFS(name string) (*ZFS, error) {
	// zfs executes the 'zfs' command with the provided arguments.
	// if the 'zfs' command return code is 0, it returns stdout
	// else it returns stderr and the error
	execZFS := func(first string, rest ...string) (string, error) {
		args := append(strings.Split(first, " "), rest...)
		logDebug.Printf("execute: zfs %s %s\n", first, strings.Join(rest, " "))

		cmd := exec.Command("zfs", args...)

		var stdoutBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf

		var stderrBuf bytes.Buffer
		cmd.Stderr = &stderrBuf

		if cmdErr := cmd.Run(); cmdErr != nil {
			logError.Printf("executing zfs cmd: %s: %s\n", cmdErr.Error(), stderrBuf.String())
			return stderrBuf.String(), cmdErr
		}

		return strings.TrimRight(stdoutBuf.String(), "\n"), nil
	}

	mountPoint, err := execZFS("get -H -o value mountpoint", name)
	if err != nil {
		return nil, err
	}

	return &ZFS{
		name,
		mountPoint,
		execZFS,
	}, nil
}

type execZFSFunc func(string, ...string) (string, error)
