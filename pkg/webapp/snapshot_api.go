package webapp

import (
	"fmt"
	"net/http"
)

/// responds with a list of snapshots for the given dataset
///
/// expected payload: { datasetName: "name" }
func (self *WebApp) snapshotsForDatasetHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	type Payload struct {
		DatasetName string `json:"datasetName"`
	}

	payload, ok := decodeJsonPayload(w, r, &Payload{}).(*Payload)
	if !ok {
		return
	}

	// get the dataset
	ds, err := self.zfs.FindDatasetByName(payload.DatasetName)
	if err != nil {
		msg := fmt.Sprintf("Dataset with name: %s not found - %v", payload.DatasetName, err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	// snapshots
	snaps, err := ds.ScanSnapshots()
	if err != nil {
		msg := fmt.Sprintf("Unable to scan snapshots for Dataset: %s - %v", payload.DatasetName, err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	respond(w, r, snaps)
}

func (self *WebApp) createSnapshotHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	type Payload struct {
		DatasetName  string `json:"datasetName"`
		SnapshotName string `json:"snapshotName"`
	}

	payload, ok := decodeJsonPayload(w, r, &Payload{}).(*Payload)
	if !ok {
		return
	}

	// get the dataset
	ds, err := self.zfs.FindDatasetByName(payload.DatasetName)
	if err != nil {
		msg := fmt.Sprintf("Dataset with name: %s not found - %v", payload.DatasetName, err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	name, err := ds.CreateSnapshot(payload.SnapshotName)
	if err != nil {
		msg := fmt.Sprintf("Unable to create snapshot: %s - %v", name, err)
		log.Error(msg)
		http.Error(w, msg, 500)
		return
	}

	msg := fmt.Sprintf("Snapshot '%s' created", name)
	log.Info(msg)
	w.Write([]byte(msg))
}

func (self *WebApp) destroySnapshotHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	type Payload struct {
		DatasetName  string   `json:"datasetName"`
		SnapshotName string   `json:"snapshotName"`
		DestroyFlags []string `json:"destroyFlags"`
	}

	payload, ok := decodeJsonPayload(w, r, &Payload{}).(*Payload)
	if !ok {
		return
	}

	// get the dataset
	ds, err := self.zfs.FindDatasetByName(payload.DatasetName)
	if err != nil {
		msg := fmt.Sprintf("Dataset with name: %s not found - %v", payload.DatasetName, err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	var flags []string
	for _, flag := range payload.DestroyFlags {
		if is_valid_flag([]string{"-R", "-d", "-r"}, flag) {
			flags = append(flags, flag)
		} else {
			log.Warnf("ignore invalid destroy snapshot flag: '%s'", flag)
		}
	}

	err = ds.DestroySnapshot(payload.SnapshotName, flags)
	if err != nil {
		msg := fmt.Sprintf("Unable to destroy snapshot: %s - %v", payload.SnapshotName, err)
		log.Error(msg)
		http.Error(w, msg, 500)
		return
	}

	msg := fmt.Sprintf("Snapshot '%s' destroyed", payload.SnapshotName)
	log.Info(msg)
	w.Write([]byte(msg))
}

func (self *WebApp) rollbackSnapshotHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	type Payload struct {
		DatasetName   string   `json:"datasetName"`
		SnapshotName  string   `json:"snapshotName"`
		RollbackFlags []string `json:"rollbackFlags"`
	}

	payload, ok := decodeJsonPayload(w, r, &Payload{}).(*Payload)
	if !ok {
		return
	}

	// get the dataset
	ds, err := self.zfs.FindDatasetByName(payload.DatasetName)
	if err != nil {
		msg := fmt.Sprintf("Dataset with name: %s not found - %v", payload.DatasetName, err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	var flags []string
	for _, flag := range payload.RollbackFlags {
		if is_valid_flag([]string{"-R", "-f", "-r"}, flag) {
			flags = append(flags, flag)
		} else {
			log.Warnf("ignore invalid rollback snapshot flag: '%s'", flag)
		}
	}

	err = ds.RollbackSnapshot(payload.SnapshotName, flags)
	if err != nil {
		msg := fmt.Sprintf("Unable to rollback snapshot: %s - %v", payload.SnapshotName, err)
		log.Error(msg)
		http.Error(w, msg, 500)
		return
	}

	msg := fmt.Sprintf("Snapshot '%s' rolled back", payload.SnapshotName)
	log.Info(msg)
	w.Write([]byte(msg))
}

func (self *WebApp) renameSnapshotHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	type Payload struct {
		DatasetName     string `json:"datasetName"`
		OldSnapshotName string `json:"oldSnapshotName"`
		NewSnapshotName string `json:"newSnapshotName"`
	}

	payload, ok := decodeJsonPayload(w, r, &Payload{}).(*Payload)
	if !ok {
		return
	}

	// get the dataset
	ds, err := self.zfs.FindDatasetByName(payload.DatasetName)
	if err != nil {
		msg := fmt.Sprintf("Dataset with name: %s not found - %v", payload.DatasetName, err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	err = ds.RenameSnapshot(payload.OldSnapshotName, payload.NewSnapshotName)
	if err != nil {
		msg := fmt.Sprintf("Unable to rename snapshot: %s - %v", payload.OldSnapshotName, err)
		log.Error(msg)
		http.Error(w, msg, 500)
		return
	}

	msg := fmt.Sprintf("Snapshot '%s' renamed to '%s'", payload.OldSnapshotName, payload.NewSnapshotName)
	log.Info(msg)
	w.Write([]byte(msg))
}

func (self *WebApp) cloneSnapshotHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	type Payload struct {
		DatasetName  string   `json:"datasetName"`
		SnapshotName string   `json:"snapshotName"`
		CloneFlags   []string `json:"cloneFlags"`
		FsName       string   `json:"fsName"`
	}

	payload, ok := decodeJsonPayload(w, r, &Payload{}).(*Payload)
	if !ok {
		return
	}

	// get the dataset
	ds, err := self.zfs.FindDatasetByName(payload.DatasetName)
	if err != nil {
		msg := fmt.Sprintf("Dataset with name: %s not found - %v", payload.DatasetName, err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	var flags []string
	for _, flag := range payload.CloneFlags {
		if is_valid_flag([]string{"-p"}, flag) {
			flags = append(flags, flag)
		} else {
			log.Warnf("ignore invalid clone snapshot flag: '%s'", flag)
		}
	}

	err = ds.CloneSnapshot(payload.SnapshotName, payload.FsName, flags)
	if err != nil {
		msg := fmt.Sprintf("Unable to clone snapshot: %s - %v", payload.SnapshotName, err)
		log.Error(msg)
		http.Error(w, msg, 500)
		return
	}

	msg := fmt.Sprintf("Snapshot '%s' cloned to '%s'", payload.SnapshotName, payload.FsName)
	log.Info(msg)
	w.Write([]byte(msg))
}

func is_valid_flag(valid []string, flag string) bool {
	for _, v := range valid {
		if v == flag {
			return true
		}
	}
	return false
}
