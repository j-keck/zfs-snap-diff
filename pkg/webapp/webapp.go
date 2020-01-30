package webapp

import (
	"github.com/j-keck/plog"
	"github.com/j-keck/zfs-snap-diff/pkg/config"
	"github.com/j-keck/zfs-snap-diff/pkg/zfs"
	"net/http"
	"path/filepath"
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
		log.Infof("serve webapp from directory: %s", webappDir)
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
	http.HandleFunc("/api/stat", self.statHndl)
	http.HandleFunc("/api/dir-listing", self.dirListingHndl)
	http.HandleFunc("/api/find-file-versions", self.findFileVersionsHndl)
	http.HandleFunc("/api/snapshots-for-dataset", self.snapshotsForDatasetHndl)
	http.HandleFunc("/api/mime-type", self.mimeTypeHndl)
	http.HandleFunc("/api/download", self.downloadHndl)
	http.HandleFunc("/api/diff", self.diffHndl)
	http.HandleFunc("/api/revert-change", self.revertChangeHndl)
	http.HandleFunc("/api/restore-file", self.restoreFileHndl)
}
