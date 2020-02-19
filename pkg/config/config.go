package config

import (
	"github.com/j-keck/plog"
	"github.com/BurntSushi/toml"
	"bytes"
	"io/ioutil"
)

var log = plog.GlobalLogger()

var Get Config = Config{
	Webserver: NewDefaultWebserverConfig(),
	ZFS: NewDefaultZFSConfig(),
	UseCacheDirForBackups: true,
	DaysToScan: 7,
	MaxArchiveUnpackedSizeMB: 200,
	SnapshotNameTemplate: "zfs-snap-diff-%FT%H:%M",
}

type Config struct {
	Webserver                  WebserverConfig `toml:"webserver"`
	ZFS                        ZFSConfig       `toml:"zfs"`
	UseCacheDirForBackups      bool            `toml:"use-cache-dir-for-backups"`
	DaysToScan                 int             `toml:"days-to-scan"`
	MaxArchiveUnpackedSizeMB   int             `toml:"max-archive-unpacked-size-mb"`
	SnapshotNameTemplate       string          `toml:"snapshot-name-template"`
}


func LoadConfig(path string) {
	log.Debugf("load configuration from %s", path)
	if _, err := toml.DecodeFile(path, &Get); err != nil {
		log.Notef("config not found / not parsable - create a new: %s", path)
		SaveConfig(path)
	}
}


func SaveConfig(path string) error {
	buf := new(bytes.Buffer)

	if err := toml.NewEncoder(buf).Encode(Get); err != nil {
		return err
	}

	log.Debugf("save configuration to %s", path)
	return ioutil.WriteFile(path, buf.Bytes(), 0640)
}
