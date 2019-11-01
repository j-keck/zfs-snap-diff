package oldweb

import (
	"encoding/json"
	"fmt"
	"github.com/j-keck/plog"
	"github.com/j-keck/zfs-snap-diff/pkg/comparator"
	"github.com/j-keck/zfs-snap-diff/pkg/config"
	"github.com/j-keck/zfs-snap-diff/pkg/diff"
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
	"github.com/j-keck/zfs-snap-diff/pkg/zfs"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type FrontendConfig map[string]interface{}

var log = plog.GlobalLogger()

// mime types for static content (serve static content from binary)
var mimeTypes = map[string]string{
	"html": "text/html",
	"js":   "text/javascript",
	"css":  "text/css",
}

var z zfs.ZFS

// registers response handlers and starts the web server
func ListenAndServe(_z zfs.ZFS, cfg config.WebserverConfig, frontendCfg FrontendConfig) {
	z = _z
	http.HandleFunc("/config", configHndl(frontendCfg))
	http.HandleFunc("/snapshots-for-dataset", snapshotsForDatasetHndl)
	http.HandleFunc("/snapshots-for-file", snapshotsForFileHndl)
	http.HandleFunc("/list-dir", listDirHndl)
	http.HandleFunc("/diff-file", diffFileHndl)
	http.HandleFunc("/file-info", fileInfoHndl)
	http.HandleFunc("/read-file", readFileHndl)
	http.HandleFunc("/restore-file", restoreFileHndl)
	http.HandleFunc("/revert-change", revertChangeHndl)

	if envHasSet("ZSD_SERVE_FROM_WEBAPP") || len(cfg.WebappDir) > 0 {
		log.Infof("serve from webapp from directory: '%s'", cfg.WebappDir)
		http.Handle("/", http.FileServer(http.Dir(cfg.WebappDir)))
	} else {
		http.HandleFunc("/", serveStaticContentFromBinaryHndl)
	}

	log.Infof("start server and listen on: '%s'", cfg.ListenAddress())
	if cfg.UseTLS {
		log.Infof("open 'https://%s' in your browser", cfg.ListenAddress())
		log.Error(http.ListenAndServeTLS(cfg.ListenAddress(), cfg.CertFile, cfg.KeyFile, nil))
	} else {
		log.Infof("open 'http://%s' in your browser", cfg.ListenAddress())
		log.Error(http.ListenAndServe(cfg.ListenAddress(), nil))
	}
}

// frontend-config
func configHndl(config FrontendConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// marshal
		js, err := json.Marshal(config)
		if err != nil {
			log.Error(err.Error())
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
	log.Debugf("scan snapshots for dataset: '%s'", datasetName)

	var snapshots zfs.Snapshots
	if dataset, err := z.FindDatasetByName(datasetName.(string)); err == nil {
		snapshots, err = dataset.ScanSnapshots()
	} else {
		log.Error(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// marshal
	js, err := json.Marshal(snapshots)
	if err != nil {
		log.Error(err.Error())
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
	log.Debugf("scan snapshots where fs: '%s' was modified\n", path)
	dataset, err := findDatasetForFile(path.(string))
	if err != nil {
		log.Error(err.Error())
		return
	}

	// FIXME: handle snap limit

	// if 'scan-snap-limit' is given, limit scan to the given value
	// if limit, ok := params["scan-snap-limit"]; ok {
	//	if len(snapshots) > limit.(int) {
	//		log.Warnf("scan only %d snapshots for other fs versions (%d snapshots available)\n", limit, len(snapshots))
	//		snapshots = snapshots[:limit.(int)]
	//	}
	// }

	// FIXME: handle errors
	fh, err := fs.NewFileHandle(path.(string))
	if err != nil {
		log.Error(err.Error())
		return
	}

	cmp, err := comparator.NewComparator(params["compare-file-method"].(string), fh)
	if err != nil {
		log.Error(err.Error())
		return
	}

	versions, err := dataset.FindFileVersions(cmp, fh)
	if err != nil {
		log.Error(err.Error())
		return
	}

	var snapshots = make([]zfs.Snapshot, 0)
	for _, v := range versions {
		snapshots = append(snapshots, v.Snapshot)
	}

	// marshal
	js, err := json.Marshal(snapshots)
	if err != nil {
		log.Error(err.Error())
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

	// FIXME verify path
	// verifyPathIsUnderZMP(path, w, r)

	dh, err := fs.NewDirHandle(path)
	if err != nil {
		log.Error(err)
	}

	dirEntries, err := dh.Ls()
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// marshal
	js, err := json.Marshal(dirEntries)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// respond
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func diffFileHndl(w http.ResponseWriter, r *http.Request) {

	// parse / validate request parameter
	params, paramsValid := parseParams(w, r, "path:string,snapshot-name:string,context-size:int")
	if !paramsValid {
		return
	}

	// verify path
	// verifyPathIsUnderZMP(path, w, r)

	fh, err := fs.NewFileHandle(params["path"].(string))
	if err != nil {
		http.Error(w, "unable to open the actual fs: "+err.Error(), 400)
		return
	}

	// get the actual fs content
	actualText, err := fh.ReadString()
	if err != nil {
		http.Error(w, "unable to read the actual fs: "+err.Error(), 400)
		return
	}

	// get the dataset
	ds, err := findDatasetForFile(fh.Path)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "unable to find dataset for path: "+err.Error(), 400)
		return
	}

	snaps, err := ds.ScanSnapshots()
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "unable to get snapshots: "+err.Error(), 400)
		return
	}

	var snapText string
	for _, snap := range snaps {
		if snap.Name == params["snapshot-name"].(string) {
			snapFh, err := NewFileHandleInSnapshot(fh.Path, snap.Name)
			if err != nil {
				log.Error(err.Error())
				http.Error(w, "unable to get fs in snapshot: "+err.Error(), 400)
				return
			}
			snapText, err = snapFh.ReadString()
			if err != nil {
				log.Error(err.Error())
				http.Error(w, "unable to read fs in snapshot: "+err.Error(), 400)
				return
			}

			break
		}
	}

	// execute diff
	diff := diff.Diff(snapText, actualText, params["context-size"].(int))

	// marshal
	js, err := json.Marshal(map[string]interface{}{
		"sideBySide": diff.AsSideBySideHTML(),
		"intext":     diff.AsIntextHTML(),
		"deltas":     diff.DeltasByContext(),
		"patches":    diff.GNUDiffs,
	})

	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// respond
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// fs (meta) info
func fileInfoHndl(w http.ResponseWriter, r *http.Request) {
	// parse / validate request parameter
	params, paramsValid := parseParams(w, r, "path:string")
	if !paramsValid {
		return
	}

	path := params["path"].(string)

	// verify path
	//verifyPathIsUnderZMP(path, w, r)

	fh, err := fs.NewFileHandle(path)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), 400)
		return
	}

	type FileInfo struct {
		fs.FileHandle
		MimeType string `json:"mimeType"`
	}
	mimeType, _ := fh.MimeType()
	fi := FileInfo{fh, mimeType}

	// marshal
	js, err := json.Marshal(fi)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// respond
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// read the fs given in the query param 'path'
func readFileHndl(w http.ResponseWriter, r *http.Request) {
	// parse / validate request parameter
	params, paramsValid := parseParams(w, r, "path:string,?snapshot-name:string")
	if !paramsValid {
		return
	}

	path := params["path"].(string)

	// verify path
	// verifyPathIsUnderZMP(path, w, r)

	var fh fs.FileHandle
	var err error
	if snapName, ok := params["snapshot-name"].(string); ok {
		fh, err = NewFileHandleInSnapshot(path, snapName)
	} else {
		fh, err = fs.NewFileHandle(path)
	}

	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	contentType, err := fh.MimeType()
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// size to string
	contentLength := strconv.FormatInt(fh.Size, 10)
	contentDisposition := "attachment; filename=" + UniqueName(fh)

	w.Header().Set("Content-Disposition", contentDisposition)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", contentLength)
	fh.CopyTo(w)
}

func restoreFileHndl(w http.ResponseWriter, r *http.Request) {
	// parse / validate request parameter
	params, paramsValid := parseParams(w, r, "path:string,snapshot-name:string")
	if !paramsValid {
		return
	}

	path := params["path"].(string)

	// verify path
	// verifyPathIsUnderZMP(path, w, r)

	// get parameter snapshot-name
	snapName := params["snapshot-name"].(string)

	// get fs-handle for the actual fs
	actualFh, err := fs.NewFileHandle(path)
	if err == nil {
		// move the actual fs to the backup location if the fs was found
		if err := fs.Backup(actualFh); err != nil {
			log.Error(err.Error())
			http.Error(w, "unable to restore: "+err.Error(), 500)
			return
		}
	} else if err != nil && !os.IsNotExist(err) {
		log.Error(err.Error())
		http.Error(w, "unable to restore - actual file not found: "+err.Error(), 400)
		return
	}

	// get fs-handle for the fs from the snashot
	snapFh, err := NewFileHandleInSnapshot(actualFh.Path, snapName)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "unable to restore - file from snapshot not found: "+err.Error(), 400)
		return
	}

	// copy the fs from the snapshot as the actual fs
	if err := snapFh.Copy(path); err != nil {
		log.Error(err.Error())
		http.Error(w, "unable to restore: "+err.Error(), 500)
	} else {
		fmt.Fprintf(w, "file '%s' successful restored from snapshot: '%s'", path, snapName)
	}
}

func revertChangeHndl(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Warnf("unable to read body: %s", err.Error())
		return
	}

	//FIXME: param parsing
	var params map[string]interface{}
	if err := json.Unmarshal(body, &params); err != nil {
		log.Warnf("unable to unmarshal json: %s", err.Error())
	}

	path, pathFound := params["path"].(string)
	if !pathFound {
		log.Warnf("parameter 'path' missing")
		http.Error(w, "parameter 'path' missing", http.StatusBadRequest)
		return
	}

	// verify path
	// verifyPathIsUnderZMP(path, w, r)

	//FIXME: unmarshal without json.Marshal -> json.Unmarshal hack
	var deltas diff.Deltas
	if d, ok := params["deltas"]; ok {
		js, _ := json.Marshal(d)
		if err = json.Unmarshal(js, &deltas); err != nil {
			log.Warnf(err.Error())
			http.Error(w, "unable to unmarshal deltas-json: "+err.Error(), 500)
			return
		}
	} else {
		log.Warn("parameter 'deltas' missing")
		http.Error(w, "parameter 'deltas' missing", http.StatusBadRequest)
		return
	}

	// get fs-handle
	var fh fs.FileHandle
	if fh, err = fs.NewFileHandle(path); err != nil {
		log.Warn(err.Error())
		http.Error(w, "unable to revert change - fs not found: "+err.Error(), 400)
		return
	}

	if err := fs.Patch(fh, deltas); err != nil {
		log.Warn(err.Error())
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

	fields := strings.Split(path, ".")
	w.Header().Set("Content-Type", mimeTypes[fields[len(fields)-1]])
	data, _ := Asset(path)
	w.Write(data)
}

// envHasSet returns true, if 'key' is in the environment
func envHasSet(key string) bool {
	return len(os.Getenv(key)) > 0
}
