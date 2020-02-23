package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
	"github.com/j-keck/zfs-snap-diff/pkg/scanner"
	"os"
)

func cacheFileVersions(versions []scanner.FileVersion) error {
	j, err := json.Marshal(versions)
	if err != nil {
		return err
	}

	cacheDir, err := fs.CacheDir()
	if err != nil {
		return err
	}

	_, err = cacheDir.WriteFile("zsd.cache", j, 0644)
	return err
}

func loadCachedFileVersions() ([]scanner.FileVersion, error) {

	cacheDir, err := fs.CacheDir()
	if err != nil {
		return nil, err
	}

	b, err := cacheDir.ReadFile("zsd.cache")
	if os.IsNotExist(err) {
		return nil, errors.New("cached file-versions not found - try the 'list' action at first")
	} else if err != nil {
		return nil, fmt.Errorf("unable to load cached file-version - %v", err)
	}

	versions := make([]scanner.FileVersion, 0)
	err = json.Unmarshal(b, &versions)
	return versions, err
}
