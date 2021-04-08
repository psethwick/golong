package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGeneratedKeysCanBeUrlFragments(t *testing.T) {
	for i := 0; i < 1000; i++ {
		k := generateKey()
		u := fmt.Sprintf("http://example.com/%s", k)

		_, err := url.ParseRequestURI(u)

		if err != nil {
			t.Errorf("Invalid url fragment: %s", k)
		}
	}
}

func TestKeyRetrieval(t *testing.T) {
	k := generateKey()
	v := "fragment"

	storeRedirect(k, v)

	rv, _ := lookupRedirect(k)

	if v != rv {
		t.Errorf("Expected %s to be %s", rv, v)
	}
}

func TestKeyNonRetrieval(t *testing.T) {
	k := generateKey()

	_, err := lookupRedirect(k)

	if err == nil {
		t.Errorf("Expected error, but did not get one")
	}
}

func TestGetRequestNotFound(t *testing.T) {
	route := fmt.Sprintf("/%s", generateKey())
	request, _ := http.NewRequest("GET", route, nil)

	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(requestHandler)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("Should have been a 404, but was %d", recorder.Code)
	}
}

func TestGetRequestRedirect(t *testing.T) {
	key := generateKey()
	url := "http://example.com"
	storeRedirect(key, url)

	route := fmt.Sprintf("/%s", key)
	request, _ := http.NewRequest("GET", route, nil)

	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(requestHandler)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusSeeOther {
		t.Errorf("Should have been a 303, but was %d", recorder.Code)
		return
	}
	res := recorder.Result()

	if res.Header.Get("Location") != url {
		t.Errorf("Location: %s, expecting %s", res.Header.Get("Location"), url)
	}
}

func TestPostRequestStoresUrl(t *testing.T) {
	testUrl := "http://example.com"
	ur := RedirectRequest{testUrl}
	serialized, err := json.Marshal(ur)
	if err != nil {
		t.Errorf(err.Error())
	}
	reader := strings.NewReader(string(serialized))
	recorder := httptest.NewRecorder()

	request, _ := http.NewRequest("POST", "/", reader)
	request.Header.Set("Content-Type", "application/json")

	handler := http.HandlerFunc(requestHandler)
	handler.ServeHTTP(recorder, request)

	res := recorder.Result()
	body, err := ioutil.ReadAll(res.Body)

	var rr RedirectResponse

	err = json.Unmarshal(body, &rr)
	if err != nil {
		t.Errorf("could not parse %s: %s", body, err.Error())
	}

	key := strings.Replace(rr.Source, buildRedirectUrl(""), "", 1)

	red, err := lookupRedirect(key)

	if err != nil {
		t.Errorf("could not find %s in redirect store", key)
		return
	}

	if red != testUrl {
		t.Errorf("Expected %s to be %s", red, testUrl)
	}
}
