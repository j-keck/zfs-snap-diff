package webapp

import (
	"fmt"
	"github.com/j-keck/zfs-snap-diff/pkg/config"
	"github.com/j-keck/zfs-snap-diff/pkg/diff"
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
	"github.com/j-keck/zfs-snap-diff/pkg/scanner"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

/// responds the configuration
func (self *WebApp) configHndl(w http.ResponseWriter, r *http.Request) {
	respond(w, r, struct {
		DaysToScan           int          `json:"daysToScan"`
		SnapshotNameTemplate string       `json:"snapshotNameTemplate"`
	}{
		DaysToScan:           config.Get.DaysToScan,
		SnapshotNameTemplate: config.Get.SnapshotNameTemplate,
	})
}

/// re-scan datasets
func (self *WebApp) rescanDatasetsHndl(w http.ResponseWriter, r *http.Request) {
	err := self.zfs.RescanDatasets()
	if err != nil {
		msg := fmt.Sprintf("Unable to scan datasets: %v", err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

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
		msg := fmt.Sprintf("Unable to stat path: %s - %v", payload.Path, err)
		log.Error(msg)
		http.Error(w, msg, 400)
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
		msg := fmt.Sprintf("Unable to get directory handle for: %s - %v", payload.Path, err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	// get the directory listing
	entries, err := dh.Ls()
	if err != nil {
		msg := fmt.Sprintf("Directory listing failed for directory: %s - %v", payload.Path, err)
		log.Error(msg)
		http.Error(w, msg, 500)
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

	dateRange := scanner.NDaysBack(config.Get.DaysToScan, time.Now())
	compareMethod := config.Get.CompareMethod
	defaults := Payload{CompareMethod: compareMethod, DateRange: dateRange}
	payload, ok := decodeJsonPayload(w, r, &defaults).(*Payload)
	if !ok {
		return
	}

	// get the dataset
	ds, err := self.zfs.FindDatasetForPath(payload.Path)
	if err != nil {
		msg := fmt.Sprintf("Dataset for file: %s not found - %v", payload.Path, err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	// scan for other file versions
	sc := scanner.NewScanner(payload.DateRange, payload.CompareMethod, ds, self.zfs)
	scanResult, err := sc.FindFileVersions(payload.Path)
	if err != nil {
		msg := fmt.Sprintf("File versions search failed - %v", err)
		log.Error(msg)
		http.Error(w, msg, 500)
		return
	}

	respond(w, r, scanResult)
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
		msg := fmt.Sprintf("Unable to open the file: %s - %v", payload.Path, err)
		log.Error(msg)
		http.Error(w, msg, 500)
		return
	}

	// build the response
	mimeType, err := fh.MimeType()
	if err != nil {
		msg := fmt.Sprintf("Unable to determine the mime type for the file: %s - %v", payload.Path, err)
		log.Error(msg)
		http.Error(w, msg, 500)
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
		msg := fmt.Sprintf("Unable to handle download - paramater 'path' missing")
		log.Error(msg)
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
		msg := fmt.Sprintf("Unable to open the file: %s - %v", payload.Path, err)
		log.Error(msg)
		http.Error(w, msg, 500)
		return
	}

	// build the response
	contentType, err := fh.MimeType()
	if err != nil {
		msg := fmt.Sprintf("Unable to determine the mime type for the file: %s - %v", payload.Path, err)
		log.Error(msg)
		http.Error(w, msg, 500)
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
/// expected payload: { currentPath: "/path/to/file"
///                   , backupPath: "/snapshot/file"
///                   [, diff-context-size: 8 ]
///                   }
///
func (self *WebApp) diffHndl(w http.ResponseWriter, r *http.Request) {

	// decode the payload
	type Payload struct {
		CurrentPath     string `json:"currentPath"`
		BackupPath      string `json:"backupPath"`
		DiffContextSize int    `json:"diffContextSize"`
	}

	diffContextSize := config.Get.DiffContextSize
	payload, ok := decodeJsonPayload(w, r, &Payload{DiffContextSize: diffContextSize}).(*Payload)
	if !ok {
		return
	}

	if err := self.checkPathIsAllowed(payload.CurrentPath); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := self.checkPathIsAllowed(payload.BackupPath); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// diff
	diffs, err := diff.NewDiffFromPath(payload.BackupPath, payload.CurrentPath, payload.DiffContextSize)
	if err != nil {
		msg := fmt.Sprintf("Unable to create diff - %v", err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}
	respond(w, r, diffs)
}

/// revert a changeset
///
/// expected payload: { currentPath: "/path/to/file"
///                   , backupPath: "/snapshot/file"
///                   , deltaIdx: 0
///                   }
func (self *WebApp) revertChangeHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	type Payload struct {
		CurrentPath string `json:"currentPath"`
		BackupPath  string `json:"backupPath"`
		DeltaIdx    int    `json:"deltaIdx"`
	}

	payload, ok := decodeJsonPayload(w, r, &Payload{}).(*Payload)
	if !ok {
		return
	}

	// valiate path
	if err := self.checkPathIsAllowed(payload.CurrentPath); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// validate path
	if err := self.checkPathIsAllowed(payload.BackupPath); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// diff
	diffs, err := diff.NewDiffFromPath(payload.BackupPath, payload.CurrentPath, 3)
	if err != nil {
		msg := fmt.Sprintf("Unable to create diff - %v", err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	// create a backup from the current file
	var backup string
	fh, _ := fs.GetFileHandle(payload.CurrentPath)
	if backup, err = fh.Backup(); err != nil {
		msg := fmt.Sprintf("Unable to backup the file - %v", err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	// patch
	err = diff.PatchPath(payload.CurrentPath, diffs.Deltas[payload.DeltaIdx])
	if err != nil {
		msg := fmt.Sprintf("Unable to revert change - %v", err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	msg := fmt.Sprintf("Change reverted - Backup created at '%s'", backup)
	log.Info(msg)
	w.Write([]byte(msg))
}

/// restore a file version
///
/// expected payload: { currentPath: "/path/to/file"
///                   , backupPath: "/snapshot/file"
///                   }
func (self *WebApp) restoreFileHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	type Payload struct {
		CurrentPath string `json:"currentPath"`
		BackupPath  string `json:"backupPath"`
	}

	payload, ok := decodeJsonPayload(w, r, &Payload{}).(*Payload)
	if !ok {
		return
	}

	// valiate path
	if err := self.checkPathIsAllowed(payload.CurrentPath); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// validate path
	if err := self.checkPathIsAllowed(payload.BackupPath); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// get the current file
	currentFh, err := fs.GetFileHandle(payload.CurrentPath)
	if err != nil {
		msg := fmt.Sprintf("Unable to open current file - %v", err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	// get the backup file
	backupFh, err := fs.GetFileHandle(payload.BackupPath)
	if err != nil {
		msg := fmt.Sprintf("Unable to open backup file - %v", err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	// create a backup from the current file
	var backup string
	if backup, err = currentFh.Backup(); err != nil {
		msg := fmt.Sprintf("Unable to backup the file - %v", err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	// restore the backup file
	if err := backupFh.Copy(payload.CurrentPath); err != nil {
		msg := fmt.Sprintf("Unable to restore the file - %v", err)
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	msg := fmt.Sprintf("File '%s' restored. Backup created at '%s'",
		currentFh.Name, backup)
	log.Info(msg)
	w.Write([]byte(msg))
}

/// create a archive
///
/// expected request parameters: /api/archive?path=/path/to/dir[&name=other-name.zip]
///
func (self *WebApp) prepareArchiveHndl(w http.ResponseWriter, r *http.Request) {
	type Payload struct {
		Path string `json:"path"`
		Name string `json:"name"`
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
		if values, ok := r.URL.Query()["name"]; ok {
			payload.Name = values[0]
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
		msg := fmt.Sprintf("Unable to prepare archive - Paramater 'path' missing")
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	// valiate path
	if err := self.checkPathIsAllowed(payload.Path); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	dir, err := fs.GetDirHandle(payload.Path)
	if err != nil {
		msg := fmt.Sprintf("Requested path not found - %v", err)
		log.Error(msg)
		http.Error(w, msg, 500)
		return
	}

	if len(payload.Name) == 0 {
		payload.Name = dir.Name + ".zip"
	}

	// create archive
	_, err = dir.CreateArchive(payload.Name)
	if err != nil {
		msg := fmt.Sprintf("Unable to create the archive - %v", err)
		log.Error(msg)
		http.Error(w, msg, 500)
		return
	}

	w.Write([]byte(payload.Name))
}

func (self *WebApp) downloadArchiveHndl(w http.ResponseWriter, r *http.Request) {
	var name string
	if values, ok := r.URL.Query()["name"]; ok {
		name = values[0]
	} else {
		msg := "Unable to serve archive - Paramater 'name' missing"
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	tempDir, err := fs.TempDir()
	archive, err := tempDir.GetFileHandle(name)
	if err != nil {
		msg := fmt.Sprintf("Unable to find the archive - %v", err)
		log.Error(msg)
		http.Error(w, msg, 500)
		return
	}
	log.Infof("serve archive: %s", archive.Path)

	contentLength := strconv.FormatInt(archive.Size, 10)
	contentDisposition := "attachment; filename=" + name

	w.Header().Set("Content-Disposition", contentDisposition)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Length", contentLength)
	archive.CopyTo(w)
	defer func() {
		log.Debugf("remove served archive: %s", archive.Path)
		archive.Remove()
	}()
}
