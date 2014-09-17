package main

import (
	"encoding/json"
	"fmt"
	"log"
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
	http.HandleFunc("/snapshot-diff", verifyParamExistsHndl("snapshot-name", snapshotDiffHndl))
	http.HandleFunc("/list-dir", verifyParamExistsHndl("path", verifyParamUnderZMPHndl("path", listDirHndl)))
	http.HandleFunc("/read-file", verifyParamExistsHndl("path", verifyParamUnderZMPHndl("path", readFileHndl)))
	http.HandleFunc("/file-info", verifyParamExistsHndl("path", verifyParamUnderZMPHndl("path", fileInfoHndl)))
	// serve static content from 'webapps' directory if environment has 'ZSD_SERVE_FROM_WEBAPPS' set (for dev)
	if envHasSet("ZSD_SERVE_FROM_WEBAPP") {
		log.Println("serve from webapp")
		http.Handle("/", http.FileServer(http.Dir("webapp")))
	} else {
		http.HandleFunc("/", serveStaticContentFromBinaryHndl)
	}
	http.ListenAndServe(addr, nil)
}

// ensures the query parameter 'paramName' exists
//   * calls the 'next' HndlFunc when query parameter found
//   * complete the request with a 400 code if the parameter is missing
func verifyParamExistsHndl(paramName string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := getQueryParameter(r, paramName); ok {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, fmt.Sprintf("parameter '%s' missing", paramName), 400)
		}
	}
}

// ensures the query param value points under zfsMountPoint
//   * calls the 'next' HndlFunc when the security test succeed
//   * complete the request with a 403 code and !! SHUTDOWNS THE PROGRAMM !! when the test fails
func verifyParamUnderZMPHndl(paramName string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paramValue, _ := getQueryParameter(r, paramName)

		// check if 'paramValue' are under 'zfsMountPoint'
		if strings.HasPrefix(filepath.Clean(paramValue), zfsMountPoint) {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, fmt.Sprintf("invalid '%s'", paramName), 403)
			log.Printf("illegal request - url-path: %s, param: '%s', value: '%s' from client: '%s' -> SHUTDOWN SERVER!",
				r.URL.Path, paramName, paramValue, r.RemoteAddr)

			// trigger shutdown in a goroutine, to give the server time serve the 403 error
			go os.Exit(1)
		}
	}
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
	if path, ok := getQueryParameter(r, "where-file-modified"); ok {
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
	snapName, _ := getQueryParameter(r, "snapshot-name")

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
	path, _ := getQueryParameter(r, "path")

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
	path, _ := getQueryParameter(r, "path")

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

func fileInfoHndl(w http.ResponseWriter, r *http.Request) {
	path, _ := getQueryParameter(r, "path")

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

func getQueryParameter(r *http.Request, name string) (string, bool) {
	value := r.URL.Query().Get(name)
	return value, len(value) > 0
}
