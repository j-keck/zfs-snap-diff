package webapp

import (
	"encoding/json"
	"github.com/j-keck/plog"
	"github.com/j-keck/zfs-snap-diff/pkg/config"
	"github.com/j-keck/zfs-snap-diff/pkg/diff"
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
	"github.com/j-keck/zfs-snap-diff/pkg/scanner"
	"github.com/j-keck/zfs-snap-diff/pkg/zfs"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

var log = plog.GlobalLogger()

type WebApp struct {
	zfs zfs.ZFS
	cfg config.Config
}

func NewWebApp(zfs zfs.ZFS, cfg config.Config) WebApp {
	self := new(WebApp)
	self.zfs = zfs
	self.cfg = cfg
	self.registerAssetsEndpoint()
	self.registerApiEndpoints()
	return *self
}

func (self *WebApp) Start() error {
	log.Infof("listen on %s", self.cfg.Webserver.ListenAddress())
	return http.ListenAndServe(self.cfg.Webserver.ListenAddress(), nil)
}

func (self *WebApp) registerAssetsEndpoint() {
	mimeTypes := map[string]string{
		".html": "text/html",
		".js":   "text/javascript",
		".css":  "text/css",
		".svg":  "image/svg+xml",
	}

	if self.cfg.Webserver.WebappDir != "" {
		webappDir := self.cfg.Webserver.WebappDir
		log.Debugf("serve webapp from directory: %s", webappDir)
		http.Handle("/", http.FileServer(http.Dir(webappDir)))
	} else {
		log.Debug("serve embedded webapp")
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if path == "/" {
				path = "index.html"
			}

			path = strings.TrimLeft(path, "/")
			if data, err := Asset(path); err == nil {
				suffix := filepath.Ext(path)
				mimeType := mimeTypes[suffix]

				log.Tracef("serve embedded '%s' as 'Content-Type': '%s'", path, mimeType)
				w.Header().Set("Content-Type", mimeType)
				w.Write(data)
			} else {
				log.Warnf("unable to serve embedded '%s': %v", path, err)
				http.NotFound(w, r)
			}
		})
	}
}

func (self *WebApp) registerApiEndpoints() {
	http.HandleFunc("/api/config", self.configHndl)
	http.HandleFunc("/api/datasets", self.datasetsHndl)
	http.HandleFunc("/api/dir-listing", self.dirListingHndl)
	http.HandleFunc("/api/find-file-versions", self.findFileVersionsHndl)
	http.HandleFunc("/api/mime-type", self.mimeTypeHndl)
	http.HandleFunc("/api/download", self.downloadHndl)
	http.HandleFunc("/api/diff", self.diffHndl)
}

func (self *WebApp) configHndl(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Datasets zfs.Datasets `json:"datasets"`
	}
	payload.Datasets = self.zfs.Datasets()
	if response, err := json.Marshal(payload); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	} else {
		log.Errorf("Marshal error: %v", err)
		http.Error(w, "Marshal error", 500)
	}
}

func (self *WebApp) datasetsHndl(w http.ResponseWriter, r *http.Request) {
	if js, err := json.Marshal(self.zfs.Datasets()); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func (self *WebApp) dirListingHndl(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		DatasetName string `json:"dataset-name"`
		Path        string `json:"path"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		log.Errorf("Decoding payload error: %v", err)
		http.Error(w, "Invalid payload", 400)
		return
	}

	// FIXME: use dataset to get a fshandle?
	// FIXME: valiate path
	fsh, err := fs.NewFSHandle(payload.Path)
	if err != nil {
		log.Errorf("Unable to open the file: %v", err)
		http.Error(w, "Unable to open the requested file", 500)
		return
	}

	dh, err := fsh.AsDirHandle()
	if err != nil {
		log.Error(err)
		http.Error(w, "Requested path was not a directory", 400)
		return
	}

	entries, err := dh.Ls()
	if err != nil {
		log.Errorf("Directory listing failed: %v", err)
		http.Error(w, "Directroy listing failed", 500)
		return
	}

	reponse, err := json.Marshal(entries)
	if err != nil {
		log.Errorf("Marshalling error: %v", err)
		http.Error(w, "Marshalling error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(reponse)
}

func (self *WebApp) findFileVersionsHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	payload := struct {
		Path          string            `json:"path"`
		CompareMethod string            `json:"compare-method"`
		TimeRange     scanner.TimeRange `json:"time-range"`
		// FIXME: remove hard coded default days
	}{CompareMethod: "auto", TimeRange: scanner.TimeRangeFromLastNDays(7)}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		log.Errorf("Decoding payload error: %v", err)
		http.Error(w, "Invalid payload", 400)
		return
	}

	// open the file handle
	fh, err := fs.NewFileHandle(payload.Path)
	if err != nil {
		log.Errorf("Unable to open the file: %v", err)
		http.Error(w, "Unable to open the requested file", 500)
		return
	}

	ds, err := self.zfs.FindDatasetForPath(fh.Path)
	if err != nil {
		log.Error(err)
		http.Error(w, "Dataset not found", 400)
		return
	}

	sc, err := scanner.NewScanner(payload.TimeRange, payload.CompareMethod, fh)
	if err != nil {
		log.Errorf("Unable to create Scanner: %v", err)
		http.Error(w, "Error", 400)
		return
	}

	versions, err := sc.FindFileVersions(ds)
	if err != nil {
		log.Errorf("File versions search failed: %v", err)
		http.Error(w, "File versions search failed", 500)
		return
	}

	reponse, err := json.Marshal(versions)
	if err != nil {
		log.Errorf("Marshalling error: %v", err)
		http.Error(w, "Marshalling error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(reponse)
}

func (self *WebApp) mimeTypeHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	var payload struct {
		Path string `json:"path"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		log.Errorf("Decoding payload error: %v", err)
		http.Error(w, "Invalid payload", 400)
		return
	}
	log.Debugf("/api/mime-type - %+v", payload)

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

	// marshal
	js, err := json.Marshal(struct {
		MimeType string `json:"mimeType"`
	}{mimeType})
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// respond
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (self *WebApp) downloadHndl(w http.ResponseWriter, r *http.Request) {
	var path string
	var name string
	if r.Method == "GET" {
		log.Debugf("%s - %v", r.URL.Path, r.URL.Query())
		if values, ok := r.URL.Query()["path"]; ok {
			path = values[0]
		} else {
			log.Errorf("Parameter 'path' missing - query: %v", r.URL.Query())
			http.Error(w, "Parameter 'path' missing", 400)

		}
		if values, ok := r.URL.Query()["name"]; ok {
			name = values[0]
		} else {
			log.Errorf("Parameter 'name' missing - query: %v", r.URL.Query())
			http.Error(w, "Parameter 'name' missing", 400)
			return
		}
	} else {
		// decode the payload
		var payload struct {
			Path string `json:"path"`
			Name string `json:"name"`
		}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&payload); err != nil {
			log.Errorf("Decoding payload error: %v", err)
			http.Error(w, "Invalid payload", 400)
			return
		}
		log.Debugf("/api/read-file - %+v", payload)
		path = payload.Path
		name = payload.Name
	}

	// open the file handle
	fh, err := fs.NewFileHandle(path)
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
	contentDisposition := "attachment; filename=" + name

	w.Header().Set("Content-Disposition", contentDisposition)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", contentLength)
	fh.CopyTo(w)
}

func (self *WebApp) diffHndl(w http.ResponseWriter, r *http.Request) {
	// decode the payload
	payload := struct {
		ActualPath      string `json:"actual-path"`
		BackupPath      string `json:"backup-path"`
		DiffContextSize int    `json:"diff-context-size"`
	}{DiffContextSize: 5}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		log.Errorf("Decoding payload error: %v", err)
		http.Error(w, "Invalid payload", 400)
		return
	}

	// actual file content
	actualFh, err := fs.NewFileHandle(payload.ActualPath)
	if err != nil {
	}
	actualContent, err := actualFh.ReadString()
	if err != nil {
	}

	// backup file content
	backupFh, err := fs.NewFileHandle(payload.BackupPath)
	if err != nil {
	}
	backupContent, err := backupFh.ReadString()
	if err != nil {
	}

	// diff
	diff := diff.Diff(backupContent, actualContent, payload.DiffContextSize)

	// marshall
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
