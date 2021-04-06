package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tjarratt/babble"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var redirect_store map[string]string
var babbler babble.Babbler

type UrlRequest struct {
	Url string
}

func store_redirect(key string, url string) {
	log.Printf("storing key: %s, url: %s", key, url)
	if len(redirect_store) == 0 {
		redirect_store = make(map[string]string)
	}
	redirect_store[key] = url
}

func generate_key() string {
	if babbler.Count == 0 {
		babbler = babble.NewBabbler()
	}

	key := babbler.Babble()
	key = strings.ToLower(key)
	key = strings.ReplaceAll(key, "'", "")

	return key
}

func lookup_redirect(key string) (string, error) {
	url, ok := redirect_store[key]
	if ok {
		return url, nil
	}
	return "", errors.New(fmt.Sprintf("%s not found", key))
}

func get_url_from_request(r *http.Request) (string, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", errors.New("Could not read body")
	}
	log.Printf(string(body))
	var ur UrlRequest
	err = json.Unmarshal(body, &ur)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not parse body: %s", err.Error()))
	}
	return ur.Url, nil
}

func request_handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Printf("Received GET request with %s", r.URL.Path)
		url, err := lookup_redirect(strings.ReplaceAll(r.URL.Path, "/", ""))
		if err == nil {
			http.Redirect(w, r, url, http.StatusSeeOther)
		} else {
			http.NotFound(w, r)
		}

	case http.MethodPost:
		url, err := get_url_from_request(r)
		if err != nil {
			log.Printf(err.Error())
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
		log.Printf("Received POST request with url %s", url)
		key := generate_key()
		store_redirect(key, url)

	default:
		http.Error(w, fmt.Sprintf("Method %s not supported", r.Method), http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", request_handler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
