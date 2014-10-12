package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

type ZFS struct {
	Datasets ZFSDatasets
	execZFS  execZFSFunc
}

func NewZFS(name string, useSudo bool) (*ZFS, error) {
	// zfs executes the 'zfs' command with the provided arguments.
	// if the 'zfs' command return code is 0, it returns stdout
	// else it returns stderr and the error
	execZFS := func(first string, rest ...string) (string, error) {

		// build args
		args := []string{"zfs"}
		args = append(args, strings.Split(first, " ")...)
		args = append(args, rest...)
		if useSudo {
			// prepend 'sudo'
			args = append([]string{"sudo"}, args...)
		}

		logDebug.Printf("execute: %s\n", strings.Join(args, " "))
		cmd := exec.Command(args[0], args[1:]...)

		var stdoutBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf

		var stderrBuf bytes.Buffer
		cmd.Stderr = &stderrBuf

		if cmdErr := cmd.Run(); cmdErr != nil {
			logError.Printf("executing zfs cmd: %s: %s\n", cmdErr.Error(), stderrBuf.String())
			return stderrBuf.String(), cmdErr
		}

		return strings.TrimRight(stdoutBuf.String(), "\n"), nil
	}

	datasets, err := NewZFSDatasets(name, execZFS)
	return &ZFS{
		datasets,
		execZFS,
	}, err
}

func (zfs *ZFS) FindDatasetForFile(path string) ZFSDataset {
	// create a copy before sorting to keep the orginal dataset order intact
	datasets := make(ZFSDatasets, len(zfs.Datasets))
	copy(datasets, zfs.Datasets)

	// sort the datasets - longest path at first
	sort.Sort(SortByMountPointDesc(datasets))

	for _, ds := range datasets {
		if strings.HasPrefix(path, ds.MountPoint+"/") {
			return ds
		}
	}
	panic("no dataset found")
}

func (zfs *ZFS) FindDatasetByName(name string) (ZFSDataset, error) {
	for _, dataset := range zfs.Datasets {
		if dataset.Name == name {
			return dataset, nil
		}
	}
	return ZFSDataset{}, fmt.Errorf("No dataset with name: '%s' found\n", name)
}

type execZFSFunc func(string, ...string) (string, error)
