package fs

import (
	"encoding/json"
	"os"
	"fmt"
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

func (self *Kind) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "DIR":
		*self = DIR
	case "LINK":
		*self = LINK
	case "PIPE":
		*self = PIPE
	case "SOCKET":
		*self = SOCKET
	case "DEV":
		*self = DEV
	case "FILE":
		*self = FILE
	default:
		return fmt.Errorf("invalid Kind: '%s'", s)
	}
	return nil
}

func (self Kind) String() string {
	names := []string{
		"DIR", "LINK", "PIPE", "SOCKET", "DEV", "FILE",
	}

	// FIXME: bound check?
	return names[self]
}
