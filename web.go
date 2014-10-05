package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// mime types for static content (serve static content from binary)
var mimeTypes = map[string]string{
	"html": "text/html",
	"js":   "text/javascript",
	"css":  "text/css",
}

// registers response handlers and starts the web server
func listenAndServe(addr string, frontendConfig FrontendConfig) {
	http.HandleFunc("/config", configHndl(frontendConfig))
	http.HandleFunc("/list-snapshots", listSnapshotsHndl)
	http.HandleFunc("/snapshot-diff", snapshotDiffHndl)
	http.HandleFunc("/list-dir", listDirHndl)
	http.HandleFunc("/read-file", readFileHndl)
	http.HandleFunc("/file-info", fileInfoHndl)
	http.HandleFunc("/restore-file", restoreFileHndl)
	http.HandleFunc("/diff-file", diffFileHndl)
	http.HandleFunc("/revert-change", revertChangeHndl)

	// serve static content from 'webapp' directory if environment has 'ZSD_SERVE_FROM_WEBAPP' set (for dev)
	if envHasSet("ZSD_SERVE_FROM_WEBAPP") {
		logNotice.Println("serve from webapp")
		http.Handle("/", http.FileServer(http.Dir("webapp")))
	} else {
		http.HandleFunc("/", serveStaticContentFromBinaryHndl)
	}
	logError.Println(http.ListenAndServe(addr, nil))
}

// frontend-config
func configHndl(config FrontendConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// marshal
		js, err := json.Marshal(config)
		if err != nil {
			logError.Println(err.Error())
			http.Error(w, err.Error(), 500)
			return
		}

		// respond
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

// list zfs snapshots
//   * optional filter snaphots where a given file was modified
func listSnapshotsHndl(w http.ResponseWriter, r *http.Request) {
	snapshots, err := zfs.ScanSnapshots()
	if err != nil {
		logError.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// when paramter 'where-file-modified' given, filter snaphots where the
	// given file was modified.
	params, _ := extractParams(r)
	if path, ok := params["where-file-modified"]; ok {
		logDebug.Printf("scan snapshots where file: '%s' was modified\n", path)

		// if 'scan-snap-limit' is given, limit scan to the given value
		if scanSnapLimit, ok := params["scan-snap-limit"]; ok {
			limit, err := strconv.Atoi(scanSnapLimit)
			if err != nil {
				logWarn.Printf("Invalid value for 'scan-snap-limit'! - %s\n", err.Error())
				http.Error(w, err.Error(), 400)
				return
			}

			if len(snapshots) > limit {
				logNotice.Printf("scan only %d snapshots for other file versions (%d snapshots available)\n", limit, len(snapshots))
				snapshots = snapshots[:limit]
			}
		}

		// when parameter 'compare-file-method' given, use the given method.
		// if not, use auto as default
		var fileHasChangedFuncGen FileHasChangedFuncGen
		if compareFileMethod, ok := params["compare-file-method"]; ok {
			if fileHasChangedFuncGen, err = NewFileHasChangedFuncGenByName(compareFileMethod); err != nil {
				logWarn.Printf("Invalid value for 'compare-file-method'! - %s\n", err.Error())
				http.Error(w, err.Error(), 400)
				return
			}
		} else {
			// no compare-file-method given, use auto as default
			fileHasChangedFuncGen, _ = NewFileHasChangedFuncGenByName("auto")
		}

		// filter snapshots
		snapshots = snapshots.FilterWhereFileWasModified(path, fileHasChangedFuncGen)
	}

	// marshal
	js, err := json.Marshal(snapshots)
	if err != nil {
		logError.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// respond
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// diff from a given snapshot to the current filesystem state
func snapshotDiffHndl(w http.ResponseWriter, r *http.Request) {
	params, _ := extractParams(r)
	snapName, snapNameFound := params["snapshot-name"]
	if !snapNameFound {
		logWarn.Println("parameter 'snapshot-name' missing")
		respondWithParamMissing(w, "snapshot-name")
		return
	}

	diffs, err := zfs.ScanDiffs(snapName)
	if err != nil {
		logError.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// marshal
	js, err := json.Marshal(diffs)
	if err != nil {
		logError.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// respond
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// directory contents for directory given by 'path'
func listDirHndl(w http.ResponseWriter, r *http.Request) {
	params, _ := extractParams(r)
	path, pathFound := params["path"]
	if !pathFound {
		logWarn.Println("parameter 'path' missing")
		respondWithParamMissing(w, "path")
	}

	// verify path
	verifyPathIsUnderZMP(path, w, r)

	dirEntries, err := ScanDirEntries(path)
	if err != nil {
		logError.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// marshal
	js, err := json.Marshal(dirEntries)
	if err != nil {
		logError.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// respond
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// read the file given in the query param 'path'
func readFileHndl(w http.ResponseWriter, r *http.Request) {
	params, _ := extractParams(r)
	path, pathFound := params["path"]
	if !pathFound {
		logWarn.Println("parameter 'path' missing")
		respondWithParamMissing(w, "path")
		return
	}

	// verify path
	verifyPathIsUnderZMP(path, w, r)

	fh, err := NewFileHandle(path)

	if err != nil {
		logError.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	contentType, err := fh.MimeType()
	if err != nil {
		logError.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// size to string
	contentLength := strconv.FormatInt(fh.Size, 10)
	contentDisposition := "attachment; filename=" + fh.UniqueName()

	w.Header().Set("Content-Disposition", contentDisposition)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", contentLength)
	fh.CopyTo(w)
}

// file (meta) info
func fileInfoHndl(w http.ResponseWriter, r *http.Request) {
	params, _ := extractParams(r)
	path, pathFound := params["path"]
	if !pathFound {
		logWarn.Println("parameter 'path' missing")
		respondWithParamMissing(w, "path")
		return
	}

	// verify path
	verifyPathIsUnderZMP(path, w, r)

	fh, err := NewFileHandle(path)
	if err != nil {
		logError.Println(err.Error())
		http.Error(w, err.Error(), 400)
		return
	}

	type FileInfo struct {
		FileHandle
		MimeType string
	}
	mimeType, _ := fh.MimeType()
	fi := FileInfo{*fh, mimeType}

	// marshal
	js, err := json.Marshal(fi)
	if err != nil {
		logError.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// respond
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func restoreFileHndl(w http.ResponseWriter, r *http.Request) {
	params, _ := extractParams(r)
	path, pathFound := params["path"]
	if !pathFound {
		logWarn.Println("parameter 'path' missing")
		respondWithParamMissing(w, "path")
		return
	}

	// verify path
	verifyPathIsUnderZMP(path, w, r)

	// get parameter snapshot-name
	snapName, snapNameFound := params["snapshot-name"]
	if !snapNameFound {
		logWarn.Println("parameter 'snapshot-name' missing")
		respondWithParamMissing(w, "snapshot-name")
		return
	}

	// get file-handle for the actual file
	actualFh, err := NewFileHandle(path)
	if err == nil {
		// move the actual file to the backup location if the file was found
		if err := actualFh.MoveToBackup(); err != nil {
			logError.Println(err.Error())
			http.Error(w, "unable to restore: "+err.Error(), 500)
			return
		}
	} else if err != nil && !os.IsNotExist(err) {
		logError.Println(err.Error())
		http.Error(w, "unable to restore - actual file not found: "+err.Error(), 400)
		return
	}

	// get file-handle for the file from the snashot
	snapFh, err := NewFileHandleInSnapshot(path, snapName)
	if err != nil {
		logError.Println(err.Error())
		http.Error(w, "unable to restore - file from snapshot not found: "+err.Error(), 400)
		return
	}

	// copy the file from the snapshot as the actual file
	if err := snapFh.CopyAs(path); err != nil {
		logError.Println(err.Error())
		http.Error(w, "unable to restore: "+err.Error(), 500)
	} else {
		fmt.Fprintf(w, "file '%s' successful restored from snapshot: '%s'", path, snapName)
	}
}

func diffFileHndl(w http.ResponseWriter, r *http.Request) {
	params, _ := extractParams(r)
	path, pathFound := params["path"]
	if !pathFound {
		logWarn.Println("parameter 'path' missing")
		respondWithParamMissing(w, "path")
		return
	}

	// verify path
	verifyPathIsUnderZMP(path, w, r)

	// get parameter snapshot-name
	snapName, snapNameFound := params["snapshot-name"]
	if !snapNameFound {
		logWarn.Println("parameter 'snapshot-name' missing")
		respondWithParamMissing(w, "snapshot-name")
		return
	}

	// get parameter context-size
	contextSize := 5 // FIXME: from default value in main.go
	if contextSizeStr, ok := params["context-size"]; ok {
		contextSize, _ = strconv.Atoi(contextSizeStr)
	}

	// read actual file
	var actualText string
	if actualFh, err := NewFileHandle(path); err != nil {
		logError.Println(err.Error())
		http.Error(w, "unable to get file-handle for actual file: "+err.Error(), 400)
		return
	} else {
		if actualText, err = actualFh.ReadText(); err != nil {
			logError.Println(err.Error())
			http.Error(w, "unable to read actual file: "+err.Error(), 400)
			return
		}
	}

	// read snap file
	var snapText string
	if snapFh, err := NewFileHandleInSnapshot(path, snapName); err != nil {
		logError.Println(err.Error())
		http.Error(w, "unable to get file-handle for snap file: "+err.Error(), 400)
		return
	} else {
		if snapText, err = snapFh.ReadText(); err != nil {
			logError.Println(err.Error())
			http.Error(w, "unable to read snap file: "+err.Error(), 400)
			return
		}
	}

	// execute diff
	diff := Diff(snapText, actualText, contextSize)

	// marshal
	js, err := json.Marshal(map[string]interface{}{
		"sideBySide": diff.AsSideBySideHTML(),
		"intext":     diff.AsIntextHTML(),
		"deltas":     diff.DeltasByContext(),
		"patches":    diff.GNUDiffs,
	})

	if err != nil {
		logError.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// respond
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func revertChangeHndl(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logWarn.Printf("unable to read body: %s", err.Error())
		return
	}

	var params map[string]interface{}
	if err := json.Unmarshal(body, &params); err != nil {
		logWarn.Printf("unable to unmarshal json: %s\n", err.Error())
	}

	path, pathFound := params["path"].(string)
	if !pathFound {
		logWarn.Println("parameter 'path' missing")
		respondWithParamMissing(w, "path")
		return
	}

	// verify path
	verifyPathIsUnderZMP(path, w, r)

	//FIXME: unmarshal without json.Marshal -> json.Unmarshal hack
	var deltas Deltas
	if d, ok := params["deltas"]; ok {
		js, _ := json.Marshal(d)
		if err = json.Unmarshal(js, &deltas); err != nil {
			logWarn.Println(err.Error())
			http.Error(w, "unable to unmarshal deltas-json: "+err.Error(), 500)
			return
		}
	} else {
		logWarn.Println("parameter 'deltas' missing")
		respondWithParamMissing(w, "deltas")
		return
	}

	// get file-handle
	var fh *FileHandle
	if fh, err = NewFileHandle(path); err != nil {
		logWarn.Println(err.Error())
		http.Error(w, "unable to revert change - file not found: "+err.Error(), 400)
		return
	}

	if err := fh.Patch(deltas); err != nil {
		logWarn.Println(err.Error())
		http.Error(w, "unable to revert change: "+err.Error(), 500)
	}

}

// serve content from binary
//  * binary are generated at build time per: 'go-bindata webapp/...'
func serveStaticContentFromBinaryHndl(w http.ResponseWriter, r *http.Request) {
	path := "webapp" + r.URL.Path
	if strings.HasSuffix(path, "/") {
		path += "index.html"
	}

	w.Header().Set("Content-Type", mimeTypes[lastElement(path, ".")])
	data, _ := Asset(path)
	w.Write(data)
}

//
func extractParams(r *http.Request) (map[string]string, error) {
	params := make(map[string]string)

	if r.Method == "GET" {
		// extract query params
		for key, values := range r.URL.Query() {
			if len(values) > 0 {
				params[key] = values[0]
			}
		}
		return params, nil
	}

	if r.Method == "PUT" || r.Method == "POST" {
		// extract from body if content-type is 'application/json'
		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "application/json") {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logWarn.Println("unable to read body:", err.Error())
				return nil, err
			}

			// abort if body is empty
			if len(body) == 0 {
				return nil, errors.New("empty body")
			}

			if err := json.Unmarshal(body, &params); err != nil {
				logWarn.Println("unable to parse json:", err.Error())
				return nil, err
			}
			return params, nil
		}
	}

	return params, nil
}

// respond parameter missing
func respondWithParamMissing(w http.ResponseWriter, name string) {
	http.Error(w, fmt.Sprintf("parameter '%s' missing", name), 400)
}

// verified that the given path is under zfs-mount-point
//  * responds with a illegal request if not
//  * and shutdowns the server
func verifyPathIsUnderZMP(path string, w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(filepath.Clean(path), zfs.MountPoint) {
		http.Error(w, "illegal request", 403)
		logError.Printf("illegal request - file-path: '%s', url-path: '%s', from client: '%s' -> SHUTDOWN SERVER!",
			path, r.URL.Path, r.RemoteAddr)

		// trigger shutdown in a goroutine, to give the server time serve the 403 error
		go os.Exit(1)
	}
}
