package config

import (
	"runtime"
)

type ZFSConfig struct {
	UseSudo        bool `toml:"use-sudo"`
	MountSnapshots bool `toml:"mount-snapshots"`
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
