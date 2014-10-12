// helper functions for testing
package main

import (
	"io/ioutil"
	"strings"
)

func initLogHandlersForTest() {
	initLogHandlers(ioutil.Discard, ioutil.Discard, ioutil.Discard, ioutil.Discard, ioutil.Discard)
}

func execZFSMock(res string, err error) func(string, ...string) (string, error) {
	return func(first string, rest ...string) (string, error) {
		return res, err
	}
}

// create zfs output conform string
//  * use comma to split fields
//  * strings for sperate lines
//  * example
//    input: "zp1,/zp1", "zp1/a,/zp1/a"
//    result: "zp1\t/zp1\nzp1/a\t/zp1/a\n"
func zfsOutput(s ...string) string {
	out := strings.Join(s, "\n")
	out = strings.Replace(out, ",", "\t", -1)
	return out
}
