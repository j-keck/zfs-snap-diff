package fs

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// ConfigDir returns the user-local config directory
func ConfigDir() (DirHandle, error) {
	// lookup base path - somethink like '$HOME/.config'

	// `os.UserConfigDir()` exists since go1.13, but the
	// minimum supported go version for zfs-snap-diff is go.12.
	var basePath string
	if runtime.GOOS == "darwin" {
		basePath = os.Getenv("HOME")
		if basePath == "" {
			return DirHandle{}, errors.New("$HOME is not defined")
		}
		basePath += "/Library/Application Support"
	} else {
		// Unix
		basePath = os.Getenv("XDG_CONFIG_HOME")
		if basePath == "" {
			basePath = os.Getenv("HOME")
			if basePath == "" {
				return DirHandle{}, errors.New("neither $XDG_CONFIG_HOME nor $HOME are defined")
			}
			basePath += "/.config"
		}
	}

	path := filepath.Join(basePath, "zfs-snap-diff")
	return GetOrCreateDirHandle(path, 0770)
}

// CacheDir returns the user-local cache directory
func CacheDir() (DirHandle, error) {
	// lookup base path - somethink like '$HOME/.cache'
	// since go1.11: https://golang.org/pkg/os/#UserCacheDir
	basePath, err := os.UserCacheDir()
	if err != nil {
		return DirHandle{}, err
	}

	path := filepath.Join(basePath, "zfs-snap-diff")
	return GetOrCreateDirHandle(path, 0770)
}

// TempDir returns the directory for temporary files
func TempDir() (DirHandle, error) {
	tempPath := os.TempDir()
	return GetDirHandle(tempPath)
}
