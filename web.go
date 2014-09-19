package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

	// serve static content from 'webapp' directory if environment has 'ZSD_SERVE_FROM_WEBAPP' set (for dev)
	if envHasSet("ZSD_SERVE_FROM_WEBAPP") {
		log.Println("serve from webapp")
		http.Handle("/", http.FileServer(http.Dir("webapp")))
	} else {
		http.HandleFunc("/", serveStaticContentFromBinaryHndl)
	}
	log.Fatal(http.ListenAndServe(addr, nil))
}

// frontend-config
func configHndl(config FrontendConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// marshal
		js, err := json.Marshal(config)
		if err != nil {
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
	snapshots, err := ScanZFSSnapshots(zfsName)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// when paramter 'where-file-modified' given, filter snaphots where the
	// given file was modified
	params, _ := extractParams(r)
	if path, ok := params["where-file-modified"]; ok {
		snapshots = snapshots.FilterWhereFileWasModified(path)
	}

	// marshal
	js, err := json.Marshal(snapshots)
	if err != nil {
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
		respondWithParamMissing(w, "snapshot-name")
		return
	}

	diffs, err := ScanZFSDiffs(zfsName, snapName)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// marshal
	js, err := json.Marshal(diffs)
	if err != nil {
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
		respondWithParamMissing(w, "path")
	}

	// verify path
	verifyPathIsUnderZMP(path, w, r)

	dirEntries, err := ScanDirEntries(path)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// marshal
	js, err := json.Marshal(dirEntries)
	if err != nil {
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
		respondWithParamMissing(w, "path")
		return
	}

	// verify path
	verifyPathIsUnderZMP(path, w, r)

	fh, err := NewFileHandle(path)

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	contentType, err := fh.MimeType()
	if err != nil {
		log.Println(err)
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
		respondWithParamMissing(w, "path")
		return
	}

	// verify path
	verifyPathIsUnderZMP(path, w, r)

	fh, err := NewFileHandle(path)
	if err != nil {
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
		respondWithParamMissing(w, "path")
		return
	}

	// verify path
	verifyPathIsUnderZMP(path, w, r)

	// get parameter snapshot-name
	snapName, snapNameFound := params["snapshot-name"]
	if !snapNameFound {
		respondWithParamMissing(w, "snapshot-name")
		return
	}

	// get file-handle for the actual file
	actualFh, err := NewFileHandle(path)
	if err != nil {
		http.Error(w, "unable to restore - actual file not found: "+err.Error(), 400)
		return
	}

	// get file-handle for the file from the snashot
	snapFh, err := NewFileHandleInSnapshot(path, snapName)
	if err != nil {
		http.Error(w, "unable to restore - file from snapshot not found: "+err.Error(), 400)
		return
	}

	// rename the actual file: <FILENAME>_<TIMESTAMP>
	newName := fmt.Sprintf("%s_%s", actualFh.Name, time.Now().Format("20060102_150405"))
	if err := actualFh.Rename(newName); err != nil {
		http.Error(w, "unable to restore: "+err.Error(), 500)
		return
	}

	// copy the file from the snapshot as the actual file
	if err := snapFh.Copy(path); err != nil {
		http.Error(w, "unable to restore: "+err.Error(), 500)
	} else {
		fmt.Fprintf(w, "file '%s' successful restored from snapshot: '%s'", path, snapName)
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
				log.Println("unable to read body:", err.Error())
				return nil, err
			}

			// abort if body is empty
			if len(body) == 0 {
				return nil, errors.New("empty body")
			}

			if err := json.Unmarshal(body, &params); err != nil {
				log.Println("unable to parse json:", err.Error())
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
	if !strings.HasPrefix(filepath.Clean(path), zfsMountPoint) {
		http.Error(w, "illegal request", 403)
		log.Printf("illegal request - file-path: '%s', url-path: '%s', from client: '%s' -> SHUTDOWN SERVER!",
			path, r.URL.Path, r.RemoteAddr)

		// trigger shutdown in a goroutine, to give the server time serve the 403 error
		go os.Exit(1)
	}
}
