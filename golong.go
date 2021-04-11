package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tjarratt/babble"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var redirectStore map[string]string
var babbler babble.Babbler

var host = "http://localhost"
var port = "8080"

type RedirectRequest struct {
	Target string
}

type RedirectResponse struct {
	Source string
}

func buildRedirectUrl(key string) string {
	return fmt.Sprintf("%s:%s/%s", host, port, key)
}

func storeRedirect(key string, url string) {
	log.Printf("storing key: %s, url: %s", key, url)
	if len(redirectStore) == 0 {
		redirectStore = make(map[string]string)
	}
	redirectStore[key] = url
}

func generateKey() string {
	if babbler.Count == 0 {
		babbler = babble.NewBabbler()
		babbler.Count = 3
	}

	key := babbler.Babble()
	key = strings.ToLower(key)
	key = strings.ReplaceAll(key, "'", "")

	return key
}

func lookupRedirect(key string) (string, error) {
	url, ok := redirectStore[key]
	if ok {
		return url, nil
	}
	return "", errors.New(fmt.Sprintf("%s not found", key))
}

func getUrlFromRequest(r *http.Request) (string, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", errors.New("Could not read body")
	}

	var rr RedirectRequest
	err = json.Unmarshal(body, &rr)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not parse body: %s", err.Error()))
	}

	_, err = url.ParseRequestURI(rr.Target)

	if err != nil {
		return "", err
	}

	return rr.Target, nil
}

func writeResponse(w http.ResponseWriter, key string) {
	url := buildRedirectUrl(key)
	rr, err := json.Marshal(RedirectResponse{url})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(rr)
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Printf("Received GET request with %s", r.URL.Path)
		url, err := lookupRedirect(strings.ReplaceAll(r.URL.Path, "/", ""))
		if err == nil {
			http.Redirect(w, r, url, http.StatusSeeOther)
		} else {
			http.NotFound(w, r)
		}

	case http.MethodPost:
		url, err := getUrlFromRequest(r)
		if err != nil {
			log.Printf("get_url_from_request: %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Received POST request with url %s", url)
		key := generateKey()
		storeRedirect(key, url)
		writeResponse(w, key)

	default:
		http.Error(w, fmt.Sprintf("Method %s not supported", r.Method), http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", requestHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
