package fs

import (
	"encoding/json"
	"os"
)

type Kind int

const (
	DIR Kind = iota
	LINK
	PIPE
	SOCKET
	DEV
	FILE
)

func KindFromFileInfo(fileInfo os.FileInfo) Kind {
	if fileInfo.IsDir() {
		return DIR
	}

	if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		return LINK
	}

	if fileInfo.Mode()&os.ModeNamedPipe == os.ModeNamedPipe {
		return PIPE
	}

	if fileInfo.Mode()&os.ModeSocket == os.ModeSocket {
		return SOCKET
	}

	if fileInfo.Mode()&os.ModeDevice == os.ModeDevice {
		return DEV
	}

	return FILE
}

func (self Kind) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.String())
}

func (self Kind) String() string {
	names := []string{
		"DIR", "LINK", "PIPE", "SOCKET", "DEV", "FILE",
	}

	// FIXME: bound check?
	return names[self]
}
