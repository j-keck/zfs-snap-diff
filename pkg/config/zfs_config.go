package config

import (
	"runtime"
)

type ZFSConfig struct {
	UseSudo       bool
	MountSnapshot bool
}

func NewDefaultZFSConfig() ZFSConfig {
	mountSnapshot := false
	if runtime.GOOS == "darwin" {
		mountSnapshot = true
	}

	return ZFSConfig{
		UseSudo:       false,
		MountSnapshot: mountSnapshot,
	}
}
