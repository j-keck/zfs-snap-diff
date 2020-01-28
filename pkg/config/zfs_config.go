package config

import (
	"runtime"
)

type ZFSConfig struct {
	UseSudo        bool
	MountSnapshots bool
}

func NewDefaultZFSConfig() ZFSConfig {
	mountSnapshots := false
	if runtime.GOOS == "darwin" {
		mountSnapshots = true
	}

	return ZFSConfig{
		UseSudo:        false,
		MountSnapshots: mountSnapshots,
	}
}
