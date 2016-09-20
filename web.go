package main

import (
	"encoding/json"
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
func listenAndServe(addr string, webServerCfg webServerConfig, frontendCfg frontendConfig) {
	http.HandleFunc("/config", configHndl(frontendCfg))
	http.HandleFunc("/snapshots-for-dataset", snapshotsForDatasetHndl)
	http.HandleFunc("/snapshots-for-file", snapshotsForFileHndl)
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

	if webServerCfg.useTLS {
		logInfo.Printf("start server and listen on: 'https://%s'\n", addr)
		logError.Println(http.ListenAndServeTLS(addr, webServerCfg.certFile, webServerCfg.keyFile, nil))
	} else {
		logInfo.Printf("start server and listen on: 'http://%s'\n", addr)
		logError.Println(http.ListenAndServe(addr, nil))
	}
}

// frontend-config
func configHndl(config frontendConfig) http.HandlerFunc {
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

func snapshotsForDatasetHndl(w http.ResponseWriter, r *http.Request) {
	// parse / validate request parameter
	params, paramsValid := parseParams(w, r, "dataset-name:string")
	if !paramsValid {
		return
	}

	datasetName := params["dataset-name"]
	logDebug.Printf("scan snapshots for dataset: '%s'\n", datasetName)

	var snapshots ZFSSnapshots
	if dataset, err := zfs.FindDatasetByName(datasetName.(string)); err == nil {
		snapshots, err = dataset.ScanSnapshots()
	} else {
		logError.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
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

func snapshotsForFileHndl(w http.ResponseWriter, r *http.Request) {
	// parse / validate request parameter
	params, paramsValid := parseParams(w, r, "path:string,compare-file-method:string,?scan-snap-limit:int")
	if !paramsValid {
		return
	}

	path := params["path"]
	logDebug.Printf("scan snapshots where file: '%s' was modified\n", path)
	dataset := zfs.FindDatasetForFile(path.(string))
	snapshots, err := dataset.ScanSnapshots()
	if err != nil {
		logError.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// if 'scan-snap-limit' is given, limit scan to the given value
	if limit, ok := params["scan-snap-limit"]; ok {
		if len(snapshots) > limit.(int) {
			logNotice.Printf("scan only %d snapshots for other file versions (%d snapshots available)\n", limit, len(snapshots))
			snapshots = snapshots[:limit.(int)]
		}
	}

	var fileHasChangedFuncGen FileHasChangedFuncGen
	if fileHasChangedFuncGen, err = NewFileHasChangedFuncGenByName(params["compare-file-method"].(string)); err != nil {
		logError.Printf("Invalid value for 'compare-file-method'! - %s\n", err.Error())
		http.Error(w, err.Error(), 400)
		return
	}

	// filter snapshots
	snapshots = snapshots.FilterWhereFileWasModified(path.(string), fileHasChangedFuncGen)

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
	// parse / validate request parameter
	params, paramsValid := parseParams(w, r, "dataset-name:string,snapshot-name:string")
	if !paramsValid {
		return
	}

	datasetName := params["dataset-name"].(string)
	snapName := params["snapshot-name"].(string)

	dataset, _ := zfs.FindDatasetByName(datasetName)
	diffs, err := dataset.ScanDiffs(snapName)
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
	// parse / validate request parameter
	params, paramsValid := parseParams(w, r, "path:string")
	if !paramsValid {
		return
	}

	path := params["path"].(string)

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
	// parse / validate request parameter
	params, paramsValid := parseParams(w, r, "path:string,?snapshot-name:string")
	if !paramsValid {
		return
	}

	path := params["path"].(string)

	// verify path
	verifyPathIsUnderZMP(path, w, r)

	var fh *FileHandle
	var err error
	if snapName, ok := params["snapshot-name"].(string); ok {
		fh, err = NewFileHandleInSnapshot(path, snapName)
	} else {
		fh, err = NewFileHandle(path)
	}

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
	// parse / validate request parameter
	params, paramsValid := parseParams(w, r, "path:string")
	if !paramsValid {
		return
	}

	path := params["path"].(string)

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
	// parse / validate request parameter
	params, paramsValid := parseParams(w, r, "path:string,snapshot-name:string")
	if !paramsValid {
		return
	}

	path := params["path"].(string)

	// verify path
	verifyPathIsUnderZMP(path, w, r)

	// get parameter snapshot-name
	snapName := params["snapshot-name"].(string)

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

	// parse / validate request parameter
	params, paramsValid := parseParams(w, r, "path:string,snapshot-name:string,context-size:int")
	if !paramsValid {
		return
	}

	path := params["path"].(string)

	// verify path
	verifyPathIsUnderZMP(path, w, r)

	// get the actual file content
	actualText, err := readTextFrom(NewFileHandle, path)
	if err != nil {
		http.Error(w, "unable to read the actual file: "+err.Error(), 400)
		return
	}

	// get the snap file content
	snapText, err := readTextFrom(NewFileHandleInSnapshotPart(params["snapshot-name"].(string)), path)
	if err != nil {
		logError.Println(err.Error())
		http.Error(w, "unable to read snap file: "+err.Error(), 400)
		return
	}

	// execute diff
	diff := Diff(snapText, actualText, params["context-size"].(int))

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

	//FIXME: param parsing
	var params map[string]interface{}
	if err := json.Unmarshal(body, &params); err != nil {
		logWarn.Printf("unable to unmarshal json: %s\n", err.Error())
	}

	path, pathFound := params["path"].(string)
	if !pathFound {
		logWarn.Println("parameter 'path' missing")
		http.Error(w, "parameter 'path' missing", http.StatusBadRequest)
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
		http.Error(w, "parameter 'deltas' missing", http.StatusBadRequest)
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

// verified that the given path is under zfs-mount-point
//  * responds with a illegal request if not
//  * and shutdowns the server
func verifyPathIsUnderZMP(path string, w http.ResponseWriter, r *http.Request) {
	for _, dataset := range zfs.Datasets {
		if strings.HasPrefix(filepath.Clean(path), dataset.MountPoint) {
			return
		}
	}

	http.Error(w, "illegal request", 403)
	logError.Printf("illegal request - file-path: '%s', url-path: '%s', from client: '%s' -> SHUTDOWN SERVER!",
		path, r.URL.Path, r.RemoteAddr)

	// trigger shutdown in a goroutine, to give the server time serve the 403 error
	go os.Exit(1)

}
