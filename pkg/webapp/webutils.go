package webapp

import (
	"encoding/json"
	"errors"
	"net/http"
)

func (self *WebApp) checkPathIsAllowed(path string) error {
	log.Tracef("check if '%s' lives under the current dataset: '%s'",
		path, self.zfs.Name())
	if _, err := self.zfs.FindDatasetForPath(path); err != nil {
		msg := "Requested path was not in the dataset"
		log.Errorf("%s - path: '%s', dataset: '%s'", msg, path, self.zfs.Name())
		return errors.New(msg)
	}
	return nil
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

func respond(w http.ResponseWriter, r *http.Request, payload interface{}) {
	if js, err := json.Marshal(payload); err == nil {
		/* disable response body tracing.
		           this generates a lot of unecessary output which makes the logs hard to read
				if log.IsTraceEnabled() {
					log.Tracef("respond to request at: %s", r.URL)

					// format the json response and log it
					var buf bytes.Buffer
					json.Indent(&buf, js, "                                ", "  ")
					log.Tracef("  json: %s", buf.String())
				}
		*/
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else {
		msg := "Unable to marshal response payload as json"
		log.Errorf("%s - %v", msg, err)
		http.Error(w, msg, 500)
	}
}
