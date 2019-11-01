package config

import (

	"github.com/j-keck/plog"
)

var log = plog.GlobalLogger()

type Config struct {
	Webserver WebserverConfig
	ZFS       ZFSConfig
}

func NewDefaultConfig() Config {
	return Config{
		Webserver: NewDefaultWebserverConfig(),
	}
}

