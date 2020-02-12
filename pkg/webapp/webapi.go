package webapp

import (
	"fmt"
	"github.com/j-keck/zfs-snap-diff/pkg/diff"
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
	"github.com/j-keck/zfs-snap-diff/pkg/scanner"
	"github.com/j-keck/zfs-snap-diff/pkg/zfs"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

/// responds the configuration
func (self *WebApp) configHndl(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Datasets zfs.Datasets `json:"datasets"`
	}
	payload.Datasets = self.zfs.Datasets()
	respond(w, r, payload)
}

/// responds with a list of all available datasets
func (self *WebApp) datasetsHndl(w http.ResponseWriter, r *http.Request) {
	respond(w, r, self.zfs.Datasets())
}

func (self *WebApp) statHndl(w http.ResponseWriter, r *http.Request) {
	// decode payload
	type Payload struct {
		Path string `json:"path"`
	}
	payload, ok := decodeJsonPayload(w, r, &Payload{}).(*Payload)
	if !ok {
		return
	}

	fh, err := fs.GetFSHandle(payload.Path)
	if err != nil {
		log.Errorf("Unable to stat path: %s - %v", payload.Path, err)
		http.Error(w, "Unable to stat path", 400)
		return
	}
	respond(w, r, fh)
}


/// responds with a directory listing
///
/// expected payload: { path: "/path/to/dir" }
///
func (self *WebApp) dirListingHndl(w http.ResponseWriter, r *http.Request) {
	// decode payload
	type Payload struct {
		Path string `json:"path"`
	}
	payload, ok := decodeJsonPayload(w, r, &Payload{}).(*Payload)
	if !ok {
		return
	}

	if err := self.checkPathIsAllowed(payload.Path); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// get the directory handle
	dh, err := fs.GetDirHandle(payload.Path)
	if err != nil {
		log.Errorf("Unable to get directory handle for: %s - %v", payload.Path, err)
		http.Error(w, "Unable to get directory handle", 400)
		return
	}

	// get the directory listing
	entries, err := dh.Ls()
	if err != nil {
		log.Errorf("Directory listing failed for directory: %s - %v", payload.Path, err)
		http.Error(w, "Directory listing failed", 500)
		return
	}

	respond(w, r, entries)
}

/// responds with a list of file versions
///
/// expected payload: { path: "/path/to/file"
///                     [, compareMethod: [auto|size|mtime|size+mtime|content|md5] ]
///                     [, dateRange: {from: "2019-01-01", to: "2019-02-01"} ]
///                   }
///
func (self *WebApp) findFileVersionsHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	type Payload struct {
		Path          string            `json:"path"`
		CompareMethod string            `json:"compareMethod"`
		DateRange     scanner.DateRange `json:"dateRange"`
	}

	// FIXME: remove hard coded default days
	defaults := Payload{CompareMethod: "auto", DateRange: scanner.NewDateRange(time.Now(), 1)}
	payload, ok := decodeJsonPayload(w, r, &defaults).(*Payload)
	if !ok {
		return
	}

	// get the dataset
	ds, err := self.zfs.FindDatasetForPath(payload.Path)
	if err != nil {
		log.Errorf("Dataset for file: %s not found - %v", payload.Path, err)
		http.Error(w, "Dataset for the given file-path not found", 400)
		return
	}

	// scan for other file versions
	sc := scanner.NewScanner(payload.DateRange, payload.CompareMethod, ds, self.zfs)
	scanResult, err := sc.FindFileVersions(payload.Path)
	if err != nil {
		log.Errorf("File versions search failed - %v", err)
		http.Error(w, "File versions search failed", 500)
		return
	}

	respond(w, r, scanResult)
}

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
		log.Errorf("Dataset with name: %s not found - %v", payload.DatasetName, err)
		http.Error(w, "Dataset with the given name not found", 400)
		return
	}

	// snapshots
	snaps, err := ds.ScanSnapshots()
	if err != nil {
		log.Errorf("Unable to scan snapshots for Dataset: %s - %v", payload.DatasetName, err)
		http.Error(w, "Unable to scan snapshots for the Dataset", 400)
		return
	}

	respond(w, r, snaps)
}

/// responds with the mime type of the request file
///
/// expected payload: { path: "/path/to/file" }
///
func (self *WebApp) mimeTypeHndl(w http.ResponseWriter, r *http.Request) {

	// decode the payload
	type Payload struct {
		Path string `json:"path"`
	}
	payload, ok := decodeJsonPayload(w, r, &Payload{}).(*Payload)
	if !ok {
		return
	}

	if err := self.checkPathIsAllowed(payload.Path); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// open the file handle
	fh, err := fs.GetFileHandle(payload.Path)
	if err != nil {
		log.Errorf("Unable to open the file: %s - %v", payload.Path, err)
		http.Error(w, "Unable to open the requested file", 500)
		return
	}

	// build the response
	mimeType, err := fh.MimeType()
	if err != nil {
		log.Errorf("Unable to determine the mime type for the file: %s - %v", payload.Path, err)
		http.Error(w, "Unable to determine the mime type", 500)
		return
	}

	respond(w, r, struct {
		MimeType string `json:"mimeType"`
	}{mimeType})
}

/// responds with the file content
///
/// expected payload: { path: "/path/to/file" [, asName: "other-name" ] }
/// or per request parameters: /api/download?path=/path/to/file&asName=other-name
///
func (self *WebApp) downloadHndl(w http.ResponseWriter, r *http.Request) {
	type Payload struct {
		Path   string `json:"path"`
		AsName string `json:"asName"`
	}

	// payload can be given per
	//   - request paramters in the url
	//   - in the post payload as json
	//
	// determine the request type and extract the payload
	payload := Payload{}
	if r.Method == "GET" {
		if values, ok := r.URL.Query()["path"]; ok {
			payload.Path = values[0]
		}
		if values, ok := r.URL.Query()["as-name"]; ok {
			payload.AsName = values[0]
		}
	} else {
		p, ok := decodeJsonPayload(w, r, &Payload{}).(*Payload)
		if !ok {
			return
		}
		payload = *p
	}

	// validate the payload
	if len(payload.Path) == 0 {
		msg := "Paramater 'path' missing"
		log.Errorf("Unable to handle download - %s", msg)
		http.Error(w, msg, 400)
		return
	}

	if len(payload.AsName) == 0 {
		asName := filepath.Base(payload.Path)
		log.Debugf("Parameter 'asName' missing - use: %s by default", asName)
		payload.AsName = asName
	}

	if err := self.checkPathIsAllowed(payload.Path); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// open the file handle
	fh, err := fs.GetFileHandle(payload.Path)
	if err != nil {
		log.Errorf("Unable to open the file: %s - %v", payload.Path, err)
		http.Error(w, "Unable to open the requested file", 500)
		return
	}

	// build the response
	contentType, err := fh.MimeType()
	if err != nil {
		log.Errorf("Unable to determine the mime type for the file: %s - %v", payload.Path, err)
		http.Error(w, "Unable to determine the mime type", 500)
		return
	}

	contentLength := strconv.FormatInt(fh.Size, 10)
	contentDisposition := "attachment; filename=" + payload.AsName

	w.Header().Set("Content-Disposition", contentDisposition)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", contentLength)
	fh.CopyTo(w)
}

/// responds with a diff
///
/// expected payload: { actualPath: "/path/to/file"
///                   , backupPath: "/snapshot/file"
///                   [, diff-context-size: 8 ]
///                   }
///
func (self *WebApp) diffHndl(w http.ResponseWriter, r *http.Request) {

	// decode the payload
	type Payload struct {
		ActualPath      string `json:"actualPath"`
		BackupPath      string `json:"backupPath"`
		DiffContextSize int    `json:"diffContextSize"`
	}

	payload, ok := decodeJsonPayload(w, r, &Payload{DiffContextSize: 5}).(*Payload)
	if !ok {
		return
	}

	if err := self.checkPathIsAllowed(payload.ActualPath); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := self.checkPathIsAllowed(payload.BackupPath); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// diff
	diffs, err := diff.NewDiffFromPath(payload.BackupPath, payload.ActualPath, payload.DiffContextSize)
	if err != nil {
		msg := "Unable to create diff"
		log.Errorf("%s: %v", msg, err)
		http.Error(w, msg, 400)
		return
	}
	respond(w, r, diffs)
}

/// revert a changeset
///
/// expected payload: { actualPath: "/path/to/file"
///                   , backupPath: "/snapshot/file"
///                   , deltaIdx: 0
///                   }
func (self *WebApp) revertChangeHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	type Payload struct {
		ActualPath string `json:"actualPath"`
		BackupPath string `json:"backupPath"`
		DeltaIdx   int    `json:"deltaIdx"`
	}

	payload, ok := decodeJsonPayload(w, r, &Payload{}).(*Payload)
	if !ok {
		return
	}

	// valiate path
	if err := self.checkPathIsAllowed(payload.ActualPath); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// validate path
	if err := self.checkPathIsAllowed(payload.BackupPath); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// diff
	diffs, err := diff.NewDiffFromPath(payload.BackupPath, payload.ActualPath, 3)
	if err != nil {
		msg := "Unable to create diff"
		log.Errorf("%s: %v", msg, err)
		http.Error(w, msg, 400)
		return
	}

	// create a backup from the actual file
	var backup string
	fh, _ := fs.GetFileHandle(payload.ActualPath)
	if backup, err = fh.Backup(); err != nil {
		msg := "Unable to backup the file"
		log.Errorf("%s - acutal-path: %s - %v", msg, payload.ActualPath, err)
		http.Error(w, msg, 400)
		return
	}

	// patch
	err = diff.PatchPath(payload.ActualPath, diffs.Deltas[payload.DeltaIdx])
	if err != nil {
		msg := "Unable to revert change"
		log.Errorf("%s: %v", msg, err)
		http.Error(w, msg, 400)
		return
	}

	msg := fmt.Sprintf("Change reverted - Backup created at '%s'", backup)
	log.Info(msg)
	w.Write([]byte(msg))
}

/// restore a file version
///
/// expected payload: { actualPath: "/path/to/file"
///                   , backupPath: "/snapshot/file"
///                   }
func (self *WebApp) restoreFileHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	type Payload struct {
		ActualPath string `json:"actualPath"`
		BackupPath string `json:"backupPath"`
	}

	payload, ok := decodeJsonPayload(w, r, &Payload{}).(*Payload)
	if !ok {
		return
	}

	// valiate path
	if err := self.checkPathIsAllowed(payload.ActualPath); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// validate path
	if err := self.checkPathIsAllowed(payload.BackupPath); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// get the actual file
	actualFh, err := fs.GetFileHandle(payload.ActualPath)
	if err != nil {
		msg := "Unable to open actual file"
		log.Errorf("%s, path: %s - %v", msg, payload.ActualPath, err)
		http.Error(w, msg, 400)
		return
	}

	// get the backup file
	backupFh, err := fs.GetFileHandle(payload.BackupPath)
	if err != nil {
		msg := "Unable to open backup file"
		log.Errorf("%s, path: %s - %v", msg, payload.BackupPath, err)
		http.Error(w, msg, 400)
		return
	}

	// create a backup from the actual file
	var backup string
	if backup, err = actualFh.Backup(); err != nil {
		msg := "Unable to backup the file"
		log.Errorf("%s, path: %s - %v", msg, payload.ActualPath, err)
		http.Error(w, msg, 400)
		return
	}

	// restore the backup file
	if err := backupFh.Copy(payload.ActualPath); err != nil {
		msg := "Unable to restore the file"
		log.Errorf("%s, actual: %s, backup: %s - %v",
			msg, payload.ActualPath, payload.BackupPath, err)
		http.Error(w, msg, 400)
		return
	}

	msg := fmt.Sprintf("File '%s' restored. Backup created at '%s'",
		actualFh.Name, backup)
	log.Info(msg)
	w.Write([]byte(msg))
}
