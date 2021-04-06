package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGeneratedKeysCanBeUrlFragments(t *testing.T) {
	for i := 0; i < 1000; i++ {
		k := generate_key()
		u := fmt.Sprintf("http://example.com/%s", k)

		_, err := url.ParseRequestURI(u)

		if err != nil {
			t.Errorf("Invalid url fragment: %s", k)
		}
	}
}

func TestKeyRetrieval(t *testing.T) {
	k := generate_key()
	v := "fragment"

	store_redirect(k, v)

	rv, _ := lookup_redirect(k)

	if v != rv {
		t.Errorf("Expected %s to be %s", rv, v)
	}
}

func TestKeyNonRetrieval(t *testing.T) {
	k := generate_key()

	_, err := lookup_redirect(k)

	if err == nil {
		t.Errorf("Expected error, but did not get one")
	}
}

func TestGetRequestNotFound(t *testing.T) {
	route := fmt.Sprintf("/%s", generate_key())
	request, _ := http.NewRequest("GET", route, nil)

	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(request_handler)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("Should have been a 404, but was %d", recorder.Code)
	}
}

func TestGetRequestRedirect(t *testing.T) {
	key := generate_key()
	url := "http://example.com"
	store_redirect(key, url)

	route := fmt.Sprintf("/%s", key)
	request, _ := http.NewRequest("GET", route, nil)

	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(request_handler)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusSeeOther {
		t.Errorf("Should have been a 303, but was %d", recorder.Code)
	}
	res := recorder.Result()

	if res.Header.Get("Location") != url {
		t.Errorf("Location: %s, expecting %s", res.Header.Get("Location"), url)
	}
}

func TestPostRequestStoresUrl(t *testing.T) {
	ur := UrlRequest{"http://example.com"}
	serialized, err := json.Marshal(ur)
	if err != nil {
		t.Errorf(err.Error())
	}
	reader := strings.NewReader(string(serialized))
	recorder := httptest.NewRecorder()

	request, _ := http.NewRequest("POST", "/", reader)
	request.Header.Set("Content-Type", "application/json")

	handler := http.HandlerFunc(request_handler)
	handler.ServeHTTP(recorder, request)

}
