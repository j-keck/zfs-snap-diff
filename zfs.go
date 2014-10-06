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

func NewZFS(name string, useSudo bool) (*ZFS, error) {
	// zfs executes the 'zfs' command with the provided arguments.
	// if the 'zfs' command return code is 0, it returns stdout
	// else it returns stderr and the error
	execZFS := func(first string, rest ...string) (string, error) {

		// build args
		args := []string{"zfs"}
		args = append(args, strings.Split(first, " ")...)
		args = append(args, rest...)
		if useSudo {
			// prepend 'sudo'
			args = append([]string{"sudo"}, args...)
		}

		logDebug.Printf("execute: %s\n", strings.Join(args, " "))
		cmd := exec.Command(args[0], args[1:]...)

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
