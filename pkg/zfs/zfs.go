package zfs

import (
	"fmt"
	"github.com/j-keck/zfs-snap-diff/pkg/config"
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
	"github.com/j-keck/plog"
	"strconv"
	"strings"
	"sort"
	"path/filepath"
)

var log = plog.GlobalLogger()

// ZFS represents a zfs filesystem
type ZFS struct {
	datasets Datasets
	cmd      ZFSCmd
}

// NewZFS returns a handler for a zfs filesystem
func NewZFS(name string, cfg config.Config) (ZFS, error) {
	self := ZFS{}
	self.cmd = NewZFSCmd(cfg.ZFS.UseSudo)

	datasets, err := self.scanDatasets(name)
	if err != nil {
		return self, err
	}
	self.datasets = datasets
	return self, nil
}

func (self *ZFS) Datasets() Datasets {
	datasets := make(Datasets, len(self.datasets))
	copy(datasets, self.datasets)
	return datasets
}

// FindDatasetByName searches and returns the dataset with the given name
func (self *ZFS) FindDatasetByName(name string) (Dataset, error) {
	for _, dataset := range self.datasets {
		if dataset.Name == name {
			return dataset, nil
		}
	}
	return Dataset{}, fmt.Errorf("No dataset with name: '%s' found\n", name)
}

func (self *ZFS) FindDatasetForPath(path string) (Dataset, error) {
	datasets := self.Datasets()
    sort.Sort(SortByPathDesc(datasets))
	for _, ds := range datasets {
		// TODO: filepath.HasPrefix is buggy
	    //  see: https://github.com/golang/go/issues/18358
		if filepath.HasPrefix(path, ds.MountPoint.Path) {
			log.Tracef("Dataset for path found - ds: %+v, for path: %s", ds, path)
			return ds, nil
		}
	}

	return Dataset{}, fmt.Errorf("No dataset for path: '%s' found\n", path)
}

// scanDatasets returns all datasets under a given pool name
func (self *ZFS) scanDatasets(name string) (Datasets, error) {
	log.Debugf("search datasets under zfs: %s", name)

	stdout, stderr, err := self.cmd.Exec("list -Hp -o name,used,avail,refer,mountpoint -r -t filesystem", name)
	if err != nil {
		log.Debugf("unable to search datasets: %s", stderr)
		return nil, err
	}

	// parse a line from the zfs output
	parse := func(s string) (string, uint64, uint64, uint64, string, bool) {
		const n = 5
		fields := strings.SplitN(s, "\t", n)
		if len(fields) == n {
			n1, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				log.Warnf("invalid number in 'used': %v", err)
				return "", 0, 0, 0, "", false
			}
			n2, err := strconv.ParseUint(fields[2], 10, 64)
			if err != nil {
				log.Warnf("invalid number in 'avail': %v", err)
				return "", 0, 0, 0, "", false
			}
			n3, err := strconv.ParseUint(fields[3], 10, 64)
			if err != nil {
				log.Warnf("invalid number in 'refer': %v", err)
				return "", 0, 0, 0, "", false
			}

			return fields[0], n1, n2, n3, fields[4], true
		} else {
			return "", 0, 0, 0, "", false
		}
	}

	// iterate over every line from the 'zfs list ...' output.
	// each line describes a 'Dataset'.
	var datasets Datasets
	for _, line := range strings.Split(stdout, "\n") {
		if name, used, avail, refer, mountPoint, ok := parse(line); ok {
			switch mountPoint {
			case "legacy":
				// lookup real mount point
				log.Tracef("dataset: '%s' has legacy mountpoint - try to find the mountpoint", name)

				legacyMountPoint, err := findmnt(name)
				if err != nil {
					log.Tracef("%s ist not mounted - ignore", name)
				} else {
					log.Debugf("mountpoint found for dataset: '%s', mountpoint: '%s'", name, legacyMountPoint)
					if dirHandle, err := fs.NewDirHandle(legacyMountPoint); err != nil {
						return nil, err
					} else {
						datasets = append(datasets, Dataset{name, used, avail, refer, dirHandle, self.cmd})
					}
				}
				
			case "none":
				log.Notef("ignore not mounted dataset: '%s'", name);
				continue
				
			default:
				log.Debugf("dataset found - name: '%s', mountpoint: '%s'", name, mountPoint)
				if dirHandle, err := fs.NewDirHandle(mountPoint); err != nil {
					log.Warnf("unable to stat directory for dataset: %s - err: %s", name, err)
				} else {
					datasets = append(datasets, Dataset{name, used, avail, refer, dirHandle, self.cmd})
				}
			}
		} else {
			log.Debugf("ignore invalid formatted line: '%s'", line)
		}
	}
	log.Debugf("%d datasets found", len(datasets))
	return datasets, nil
}


type SortByPathDesc Datasets
func (s SortByPathDesc) Len() int {
	return len(s)
}

func (s SortByPathDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]

}

func (s SortByPathDesc) Less(i, j int) bool {
	return len(s[i].MountPoint.Path) > len(s[j].MountPoint.Path)
}
