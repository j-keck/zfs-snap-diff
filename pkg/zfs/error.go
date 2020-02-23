package zfs

type ExecZFSError struct {
	err error
}

func (self ExecZFSError) Error() string {
	return self.err.Error()
}

type ExecutableNotFound struct {
	err error
}

func (self ExecutableNotFound) Error() string {
	return self.err.Error()
}
