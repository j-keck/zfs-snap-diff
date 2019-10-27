package zfs

import (
	"bytes"
	"os/exec"
	"strings"
)

type ZFSCmd interface {
	exec(string, ...string) (string, error)
}

func NewZFSCmd(useSudo bool) ZFSCmd {
	return &zfsCmdImpl{useSudo}
}

func NewZFSCmdMock(out string, err error) ZFSCmd {
	return &zfsCmdMock{out, err}
}

type zfsCmdImpl struct {
	useSudo bool
}

func (self *zfsCmdImpl) exec(first string, rest ...string) (string, error) {
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

	if cmdErr := cmd.Run(); cmdErr != nil {
		log.Errorf("executing zfs cmd: %s: %s", cmdErr.Error(), stderrBuf.String())
		return stderrBuf.String(), cmdErr
	}

	return strings.TrimRight(stdoutBuf.String(), "\n"), nil
}

type zfsCmdMock struct {
	out string
	err error
}

func (self *zfsCmdMock) exec(first string, rest ...string) (string, error) {
	log.Tracef("would execute: %s %s", first, strings.Join(rest, " "))
	log.Tracef("  return - out: '%s', err: '%v'", self.out, self.err)
	if self.err != nil {
		return "", self.err
	}
	return self.out, nil
}
