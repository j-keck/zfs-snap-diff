// http request parameter parsing / validations
//
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// parameters
type params map[string]interface{}

// parse and validate request parameters
//   * rules format: [?]<NAME>:<TYPE>
//     - separated per ,
//     - starting with a question mark means optional parameter
//     - currently supported types: 'string' and 'int'
//     - example: 'path:string,?limit:int'
//       -> path with type string necessary
//       -> limit with type int optional
//   * returns the validated / transformed params and true on success
//   * on violation, a message is logged to the console and as http response send.
//
//
func parseParams(w http.ResponseWriter, r *http.Request, rules string) (p params, paramsValid bool) {
	paramsValid = true

	// extract params from the request
	if p, paramsValid = extractParams(w, r); paramsValid {

		// process rules
		foreachRule(rules, func(pName, pType string, isOptional bool) {
			value, valueFound := p[pName]
			if !valueFound && !isOptional {
				msg := fmt.Sprintf("parameter '%s' not found", pName)
				logError.Println(msg)
				http.Error(w, msg, http.StatusBadRequest)
				paramsValid = false
				return
			}

			// only if value available (optional value not always given)
			if valueFound {
				switch pType {
				case "int":
					i, err := strconv.Atoi(value.(string))
					if err != nil {
						msg := fmt.Sprintf("unable to convert parameter '%s' to int: %s", pName, err.Error())
						logError.Println(msg)
						http.Error(w, msg, http.StatusBadRequest)
						paramsValid = false
						return
					}
					p[pName] = i
				case "string":
					// nothing to do
				default:
					msg := fmt.Sprintf("invalid type ('%s') in rules ('%s')", pType, rules)
					logError.Println(msg)
					http.Error(w, msg, http.StatusInternalServerError)
					paramsValid = false
					return
				}
			}
		})
		return p, paramsValid
	}
	return nil, false
}

// loop over rules
func foreachRule(rules string, f func(string, string, bool)) {
	splitRule := func(f string) (string, string) {
		s := strings.Split(f, ":")
		return s[0], s[1]
	}

	// remove all spaces from the rule
	rules = strings.Replace(rules, " ", "", -1)

	// loop over rules and call 'f' for each rule
	for _, rule := range strings.Split(rules, ",") {
		isOptional := false
		if strings.HasPrefix(rule, "?") {
			// optional rule
			isOptional = true
			rule = rule[1:]
		}
		pName, pType := splitRule(rule)
		f(pName, pType, isOptional)
	}

}

// extractParams extracts parameters from the request.
//   * extract query params from get request
//   * extract body params from put / post request
//   * returns the extracted params, true on success
//   * on error, the function logs and responds the error to the client
//     and returns nil, false
func extractParams(w http.ResponseWriter, r *http.Request) (params, bool) {
	p := make(params)

	if r.Method == "GET" {
		// extract query params
		for key, values := range r.URL.Query() {
			if len(values) == 0 {
				p[key] = true
			} else {
				if len(values) > 1 {
					logNotice.Printf("more than one value for parameter '%s' received - use only the first", key)
				}
				p[key] = values[0]
			}
		}
	}

	if r.Method == "PUT" || r.Method == "POST" {
		// extract from body if content-type is 'application/json'
		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "application/json") {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				msg := fmt.Sprintf("unable to read body from %s request: %s", r.Method, err.Error())
				logError.Println(msg)
				http.Error(w, msg, http.StatusInternalServerError)
				return nil, false
			}

			// abort if body is empty
			if len(body) == 0 {
				msg := fmt.Sprintf("%s request with empty body", r.Method)
				logError.Println(msg)
				http.Error(w, msg, http.StatusBadRequest)
				return nil, false
			}

			if err := json.Unmarshal(body, &p); err != nil {
				msg := fmt.Sprintf("unmarshal json error: %s", err.Error())
				logError.Println(msg)
				http.Error(w, msg, http.StatusBadRequest)
				return nil, false
			}
		}
	}

	return p, true
}
