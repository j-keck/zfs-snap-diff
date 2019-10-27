package oldweb

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestParseParams(t *testing.T) {
	initLogHandlersForTest()

	w := httptest.NewRecorder()
	p, success := parseParams(w, newDummyReq("text=name&number=3"), "text:string,number:int")
	if !success {
		t.Fatal("unexpected parsing failure")
	}

	textType := fmt.Sprintf("%T", p["text"])
	if textType != "string" {
		t.Fatal("exepcted string - type is: ", textType)
	}

	numberType := fmt.Sprintf("%T", p["number"])
	if numberType != "int" {
		t.Fatal("exepected int - type is: ", numberType)
	}
}

func TestParseParamsOptionalParamAvailable(t *testing.T) {
	initLogHandlersForTest()

	w := httptest.NewRecorder()
	p, success := parseParams(w, newDummyReq("text=name"), "?text:string")
	if !success {
		t.Fatal("unexpected parsing failure")
	}

	if w.Code != http.StatusOK {
		t.Fatal("unexpected StatusCode:", w.Code)
	}

	if p["text"] != "name" {
		t.Fatal("unexpected value: ", p["text"])
	}
}

func TestParseParamsOptionalParamMissing(t *testing.T) {
	initLogHandlersForTest()

	w := httptest.NewRecorder()
	_, success := parseParams(w, newDummyReq(""), "?text:string")
	if !success {
		t.Fatal("unexpected parsing failure")
	}

	if w.Code != http.StatusOK {
		t.Fatal("unexpected StatusCode:", w.Code)
	}
}

func TestParseParamsInvalidRules(t *testing.T) {
	initLogHandlersForTest()

	w := httptest.NewRecorder()
	_, _ = parseParams(w, newDummyReq("text=blub"), "text:blub")

	if w.Code != http.StatusInternalServerError {
		t.Fatal("exepected 500 status code- received", w.Code)
	}

	if !strings.Contains(w.Body.String(), "invalid type ('blub') in rules ('text:blub')") {
		t.Fatal("unexpected body:", w.Body.String())
	}
}
func TestParseParamsInvalidParams(t *testing.T) {
	initLogHandlersForTest()

	w := httptest.NewRecorder()
	_, success := parseParams(w, newDummyReq("text=name&number=three"), "text:string,number:int")
	if success {
		t.Fatal("expecting parsing failure")
	}

	if w.Code != http.StatusBadRequest {
		t.Fatal("expected 400 status code - received", w.Code)
	}

	if !strings.Contains(w.Body.String(), "unable to convert parameter 'number' to int: strconv.ParseInt:") {
		t.Fatal(fmt.Sprintf("unexpected body: '%s'", w.Body.String()))
	}
}

func newDummyReq(queryParam string) *http.Request {
	testURL, _ := url.Parse(fmt.Sprintf("http://example.com?%s", queryParam))
	return &http.Request{
		Method: "GET",
		URL:    testURL,
	}
}
