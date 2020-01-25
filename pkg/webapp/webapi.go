package webapp

import (
	"bytes"
	"encoding/json"
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
	encodeJsonAndRespond(w, r, payload)
}

/// responds with a list of all available datasets
func (self *WebApp) datasetsHndl(w http.ResponseWriter, r *http.Request) {
	encodeJsonAndRespond(w, r, self.zfs.Datasets())
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

	// validate the requestd path is in the actual dataset
	if _, err := self.zfs.FindDatasetForPath(payload.Path); err != nil {
		msg := "Requested file was not in the dataset"
		log.Errorf("%s - path: %s", msg, payload.Path)
		http.Error(w, msg, 400)
		return
	}

	// get the directory handle
	dh, err := fs.NewDirHandle(payload.Path)
	if err != nil {
		log.Errorf("Unable to get directory handle for: %s - %v", payload.Path, err)
		http.Error(w, "Unable to create a directory handle", 400)
		return
	}

	// get the directory listing
	entries, err := dh.Ls()
	if err != nil {
		log.Errorf("Directory listing failed for directory: %s - %v", payload.Path, err)
		http.Error(w, "Directroy listing failed", 500)
		return
	}

	encodeJsonAndRespond(w, r, entries)
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
	sc := scanner.NewScanner(payload.DateRange, payload.CompareMethod, ds)
	versions, err := sc.FindFileVersions(payload.Path)
	if err != nil {
		log.Errorf("File versions search failed - %v", err)
		http.Error(w, "File versions search failed", 500)
		return
	}

	encodeJsonAndRespond(w, r, versions)
}


/// responds with a list of snapshots for the given dataset
///
/// expected payload: { datasetName: "name" }
func (self *WebApp) snapshotsForDatasetHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	type Payload struct {
		DatasetName   string            `json:"datasetName"`
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
	snaps, err := ds.ScanSnapshots();
	if err != nil {
		log.Errorf("Unable to scan snapshots for Dataset: %s - %v", payload.DatasetName, err)
		http.Error(w, "Unable to scan snapshots for the Dataset", 400)
		return
	}

	
	encodeJsonAndRespond(w, r, snaps)
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

	// validate the requestd path is in the actual dataset
	if _, err := self.zfs.FindDatasetForPath(payload.Path); err != nil {
		msg := "Requested file was not in the dataset"
		log.Errorf("%s - path: %s", msg, payload.Path)
		http.Error(w, msg, 400)
		return
	}

	// open the file handle
	fh, err := fs.NewFileHandle(payload.Path)
	if err != nil {
		log.Errorf("Unable to open the file: %v", err)
		http.Error(w, "Unable to open the requested file", 500)
		return
	}

	// build the response
	mimeType, err := fh.MimeType()
	if err != nil {
		log.Errorf("Unable to determine the mime type for the file: %v", err)
		http.Error(w, "Unable to determine the mime type", 500)
		return
	}

	encodeJsonAndRespond(w, r, struct {
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
		log.Tracef("Parameter 'as-name' missing - use: %s by default", asName)
		payload.AsName = asName
	}

	// validate the requestd path is in the actual dataset
	if _, err := self.zfs.FindDatasetForPath(payload.Path); err != nil {
		msg := "Requested file was not in the dataset"
		log.Errorf("%s - path: %s", msg, payload.Path)
		http.Error(w, msg, 400)
		return
	}

	// open the file handle
	fh, err := fs.NewFileHandle(payload.Path)
	if err != nil {
		log.Errorf("Unable to open the file: %v", err)
		http.Error(w, "Unable to open the requested file", 500)
		return
	}

	// build the response
	contentType, err := fh.MimeType()
	if err != nil {
		log.Errorf("Unable to determine the mime type for the file: %v", err)
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

	// validate the requestd path is in the actual dataset
	if _, err := self.zfs.FindDatasetForPath(payload.ActualPath); err != nil {
		msg := "Requested file was not in the dataset"
		log.Errorf("%s - actual-path: %s", msg, payload.ActualPath)
		http.Error(w, msg, 400)
		return
	}
	if _, err := self.zfs.FindDatasetForPath(payload.BackupPath); err != nil {
		msg := "Requested file was not in the dataset"
		log.Errorf("%s - backup-path: %s", msg, payload.BackupPath)
		http.Error(w, msg, 400)
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
	encodeJsonAndRespond(w, r, diffs)
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
		ActualPath      string `json:"actualPath"`
		BackupPath      string `json:"backupPath"`
		DeltaIdx        int    `json:"deltaIdx"`
	}

	payload, ok := decodeJsonPayload(w, r, &Payload{}).(*Payload)
	if !ok {
		return
	}

	// validate the requestd path is in the actual dataset
	if _, err := self.zfs.FindDatasetForPath(payload.ActualPath); err != nil {
		msg := "Requested file was not in the dataset"
		log.Errorf("%s - actual-path: %s", msg, payload.ActualPath)
		http.Error(w, msg, 400)
		return
	}
	if _, err := self.zfs.FindDatasetForPath(payload.BackupPath); err != nil {
		msg := "Requested file was not in the dataset"
		log.Errorf("%s - backup-path: %s", msg, payload.BackupPath)
		http.Error(w, msg, 400)
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


	err = diff.PatchPath(payload.ActualPath, diffs.Deltas[payload.DeltaIdx]);
	if err != nil {
		msg := "Unable to revert change"
		log.Errorf("%s: %v", msg, err)
		http.Error(w, msg, 400)
		return
	}
}

/// restore a file version
///
/// expected payload: { actualPath: "/path/to/file"
///                   , backupPath: "/snapshot/file"
///                   }
func (self *WebApp) restoreFileHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	type Payload struct {
		ActualPath      string `json:"actualPath"`
		BackupPath      string `json:"backupPath"`
	}

	payload, ok := decodeJsonPayload(w, r, &Payload{}).(*Payload)
	if !ok {
		return
	}

	// validate the requestd path is in the actual dataset
	if _, err := self.zfs.FindDatasetForPath(payload.ActualPath); err != nil {
		msg := "Requested file was not in the dataset"
		log.Errorf("%s - actual-path: %s", msg, payload.ActualPath)
		http.Error(w, msg, 400)
		return
	}
	if _, err := self.zfs.FindDatasetForPath(payload.BackupPath); err != nil {
		msg := "Requested file was not in the dataset"
		log.Errorf("%s - backup-path: %s", msg, payload.BackupPath)
		http.Error(w, msg, 400)
		return
	}

	actualFh, err := fs.NewFileHandle(payload.ActualPath)
	if err != nil {
		msg := "Unable to open actual file"
		log.Errorf("%s - actual-path: %s", msg, payload.ActualPath)
		http.Error(w, msg, 400)
		return		
	}
	
	backupFh, err := fs.NewFileHandle(payload.BackupPath)
	if err != nil {
		msg := "Unable to open backup file"
		log.Errorf("%s - backup-path: %s", msg, payload.BackupPath)
		http.Error(w, msg, 400)
		return		
	}

	if err := fs.Backup(actualFh); err != nil {
		msg := "Unable to backup the file"
		log.Errorf("%s - acutal-path: %s", msg, payload.ActualPath)
		http.Error(w, msg, 400)
		return		
	}
	
	if err := backupFh.Copy(payload.ActualPath); err != nil {
		msg := "Unable to restore backup file"
		log.Errorf("%s - backup-path: %s", msg, payload.BackupPath)
		http.Error(w, msg, 400)
		return		
	}
}	

	
func decodeJsonPayload(w http.ResponseWriter, r *http.Request, payload interface{}) interface{} {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(payload); err != nil {
		log.Errorf("Decoding payload error - request at: %s, error: %v", r.URL, err)
		http.Error(w, "Invalid payload", 400)
		return nil
	}
	log.Tracef("decodeJsonPayload for request at: %s - payload: %+v", r.URL, payload)
	return payload
}

// encode the given payload as json and write it in the given ResponseWriter
func encodeJsonAndRespond(w http.ResponseWriter, r *http.Request, payload interface{}) {
	if js, err := json.Marshal(payload); err == nil {
		if log.IsTraceEnabled() {
			log.Tracef("encodeJsonAndRespond to request at: %s", r.URL)

			// format the json response and log it
			var buf bytes.Buffer
			json.Indent(&buf, js, "                                ", "  ")
			log.Tracef("  json: %s", buf.String())
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else {
		log.Errorf("Unable to marshal payload as json: %v", err)
		http.Error(w, "Json encoding error", 500)
	}
}
