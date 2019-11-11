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
///                     [, timeRange: {"from": "2019-01-01T00:00:00+01:00" } ]
///                     [, timeRange: {"to": "2019-01-01T00:00:00+01:00" } ]
///                     [, timeRange: {"from": "2019-01-01T00:00:00+01:00"
///                                   ,"till": "2019-02-01T00:00:00+01:00" } ]
///                   }
///
func (self *WebApp) findFileVersionsHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	type Payload struct {
		Path          string            `json:"path"`
		CompareMethod string            `json:"compareMethod"`
		TimeRange     scanner.TimeRange `json:"timeRange"`
	}

	// FIXME: remove hard coded default days
	defaults := Payload{CompareMethod: "auto", TimeRange: scanner.TimeRangeFromLastNDays(7)}
	payload, ok := decodeJsonPayload(w, r, &defaults).(*Payload)
	if !ok {
		return
	}

	// adjust the time-range if it's invalid, because
	// the 'from' and 'to' parameters are optional
	if payload.TimeRange.FromIsAfterTo() {
		// FIXME: remove hard coded default days
		payload.TimeRange.AdjustFromToNDaysBeforeTo(7)
		log.Debugf("'From' was after 'To' in the TimeRange - adjusted time-range: %s",
			payload.TimeRange.String())
	}

	// get the dataset
	ds, err := self.zfs.FindDatasetForPath(payload.Path)
	if err != nil {
		log.Errorf("Dataset for file: %s not found - %v", payload.Path, err)
		http.Error(w, "Dataset for the given file-path not found", 400)
		return
	}

	// open the file handle
	fh, err := fs.NewFileHandle(payload.Path)
	if err != nil {
		log.Errorf("Unable to open the file: %v", err)
		http.Error(w, "Unable to open the requested file", 500)
		return
	}

	// create a new scanner instance
	sc, err := scanner.NewScanner(payload.TimeRange, payload.CompareMethod, fh)
	if err != nil {
		log.Errorf("Unable to create a scanner instance: %v", err)
		http.Error(w, "Unable to create a scanner instance", 500)
		return
	}

	// scan for other file versions
	versions, err := sc.FindFileVersions(ds)
	if err != nil {
		log.Errorf("File versions search failed - %v", err)
		http.Error(w, "File versions search failed", 500)
		return
	}

	encodeJsonAndRespond(w, r, versions)
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
/// expected payload: { actual-path: "/path/to/file"
///                   , backup-path: "/snapshot/file"
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

	// actual file content
	actualContent, err := fs.ReadTextFile(payload.ActualPath)
	if err != nil {
		msg := "Unable to read the actual file"
		log.Errorf("%s: %s - %v", msg, payload.ActualPath, err)
		http.Error(w, msg, 400)
	}

	// backup file content
	backupContent, err := fs.ReadTextFile(payload.BackupPath)
	if err != nil {
		msg := "Unable to read the backup file"
		log.Errorf("%s: %s - %v", msg, payload.ActualPath, err)
		http.Error(w, msg, 400)
	}

	// diff
	diff := diff.NewDiff(backupContent, actualContent, payload.DiffContextSize)
	encodeJsonAndRespond(w, r, diff)
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
