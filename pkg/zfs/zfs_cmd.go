package zfs

import (
	"bytes"
	"os/exec"
	"strings"
	"errors"
)

type Stdout = string
type Stderr = string

type ZFSCmd interface {
	Exec(string, ...string) (Stdout, Stderr, error)
}

func NewZFSCmd(useSudo bool) ZFSCmd {
	return &zfsCmdImpl{useSudo}
}

func NewZFSCmdMock(stdout Stdout, stderr Stderr, err error) ZFSCmd {
	return &zfsCmdMock{stdout, stderr, err}
}

type zfsCmdImpl struct {
	useSudo bool
}

func (self *zfsCmdImpl) Exec(first string, rest ...string) (Stdout, Stderr, error) {
	// build args
	args := []string{"zfs"}
	args = append(args, strings.Split(first, " ")...)
	args = append(args, rest...)
	if self.useSudo {
		// prepend 'sudo'
		args = append([]string{"sudo"}, args...)
	}

	log.Debugf("exceute: %s", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)

	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		stderr := strings.TrimRight(stderrBuf.String(), "\n")
		log.Debugf("zfs cmd failed - err: '%v', stderr: '%s'", err, stderr)

		if _, ok := err.(*exec.ExitError); ok {
			return "", stderr, ExecZFSError{errors.New(stderr)}
		}

		return "", stderr, ExecutableNotFound{err}
	}

	stdout := strings.TrimRight(stdoutBuf.String(), "\n")
	return stdout, "", nil
}

type zfsCmdMock struct {
	stdout Stderr
	stderr Stderr
	err    error
}

func (self *zfsCmdMock) Exec(first string, rest ...string) (Stdout, Stderr, error) {
	log.Tracef("would execute: %s %s", first, strings.Join(rest, " "))
	log.Tracef("  return - stdout: '%s', stderr: '%s', err: '%v'", self.stdout, self.stderr, self.err)
	if self.err != nil {
		return self.stdout, self.stderr, self.err
	}
	return self.stdout, self.stderr, self.err
}
