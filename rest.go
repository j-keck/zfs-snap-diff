package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var mimeTypes = map[string]string{
	"html": "text/html",
	"js":   "text/javascript",
	"css":  "text/css",
}

func listenAndServe(addr string) {
	http.HandleFunc("/list-snapshots", listSnapshotsHandler)
	http.HandleFunc("/snapshot-diff", snapshotDiffHandler)
	http.HandleFunc("/list-dir", checkIllegalReqHandler("dir-name", listDirHandler))
	http.HandleFunc("/read-file", checkIllegalReqHandler("file-name", readFileHandler))
	if envHasSet("ZSD_SERVE_FROM_WEBAPPS") {
		log.Println("serve from webapps")
		http.Handle("/", http.FileServer(http.Dir("webapp")))
	} else {
		http.HandleFunc("/", defaultHandler)
	}
	http.ListenAndServe(addr, nil)
}

func checkIllegalReqHandler(paramName string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if paramValue, ok := getQueryParameter(r, paramName); ok {
			if strings.HasPrefix(filepath.Clean(paramValue), zfsMountPoint) {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, fmt.Sprintf("invalid '%s'", paramName), 403)
				log.Printf("illegal request - url-path: %s, param: '%s', value: '%s' -> shutdown server!", r.URL.Path, paramName, paramValue)

				// trigger shutdown in a goroutine, to give the server time serve the 403 error
				go os.Exit(1)
			}
		} else {
			http.Error(w, fmt.Sprintf("parameter '%s' missing", paramName), 400)
		}
	}
}

func listSnapshotsHandler(w http.ResponseWriter, r *http.Request) {
	out, err := zfs("list -t snapshot -r -o name,creation -H " + zfsName)
	panicOnError(err, "list snaphshots", out)

	var snapshots Snapshots
	for _, line := range strings.Split(string(out), "\n") {
		snapshots.addFromZfsOutput(line)
	}
	snapshots = snapshots.reverse()

	js, err := json.Marshal(snapshots)
	panicOnError(err, "Marshal snapshots")

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func snapshotDiffHandler(w http.ResponseWriter, r *http.Request) {
	type Diff struct {
		ChangeType string
		FileType   string
		FileName   string
	}
	newDiffFromLine := func(line string) Diff {
		fields := strings.Split(line, "\t")
		// https://www.illumos.org/issues/1912
		r := strings.NewReplacer("\\040", " ")
		return Diff{fields[0], fields[1], r.Replace(fields[2])}
	}

	if snapName, ok := getQueryParameter(r, "snapshot-name"); ok {
		out, err := zfs(fmt.Sprintf("diff -H -F %s@%s %s", zfsName, snapName, zfsName))
		panicOnError(err, "zfs diff", out)

		var diffs []Diff
		for _, line := range strings.Split(out, "\n") {
			diffs = append(diffs, newDiffFromLine(line))
		}

		js, err := json.Marshal(diffs)
		panicOnError(err, "Marshal diffs")

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func listDirHandler(w http.ResponseWriter, r *http.Request) {
	dirName, _ := getQueryParameter(r, "dir-name")

	type DirEntry struct {
		Type    string
		Name    string
		Size    int64
		ModTime time.Time
	}
	newDirEntryFromFileInfo := func(fi os.FileInfo) DirEntry {
		_type := "F"
		if fi.IsDir() {
			_type = "D"
		}
		return DirEntry{_type, fi.Name(), fi.Size(), fi.ModTime()}
	}

	files, err := ioutil.ReadDir(dirName)
	panicOnError(err, "ReadDir")

	var dirEntries []DirEntry
	for _, fi := range files {
		dirEntries = append(dirEntries, newDirEntryFromFileInfo(fi))
	}

	js, err := json.Marshal(dirEntries)
	panicOnError(err, "Marshal dirEntries")

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func readFileHandler(w http.ResponseWriter, r *http.Request) {
	fileName, _ := getQueryParameter(r, "file-name")

	if snapName, ok := getQueryParameter(r, "snapshot-name"); ok {
		relativeFileName := strings.TrimLeft(fileName, zfsMountPoint)
		fileName = fmt.Sprintf("%s/.zfs/snapshot/%s/%s", zfsMountPoint, snapName, relativeFileName)
	}

	log.Printf("read file: %s\n", fileName)
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println(err)
		fmt.Fprint(w, "")
	} else {
		fmt.Fprint(w, string(content))
	}
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	path := "webapp" + r.URL.Path
	if strings.HasSuffix(path, "/") {
		path += "index.html"
	}
	println(path)

	w.Header().Set("Content-Type", mimeTypes[lastElement(path, ".")])
	data, _ := Asset(path)
	w.Write(data)
}

func getQueryParameter(r *http.Request, name string) (string, bool) {
	value := r.URL.Query().Get(name)
	return value, len(value) > 0
}
