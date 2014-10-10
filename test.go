// helper functions for testing
package main

import (
	"io/ioutil"
)

func initLogHandlersForTest() {
	initLogHandlers(ioutil.Discard, ioutil.Discard, ioutil.Discard, ioutil.Discard, ioutil.Discard)
}

func execZFSMock(res string, err error) func(string, ...string) (string, error) {
	return func(first string, rest ...string) (string, error) {
		return res, err
	}
}
