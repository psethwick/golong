package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tjarratt/babble"
	"io/ioutil"
	"log"
	"net/http"
)

var redirect_store map[string]string
var babbler babble.Babbler

type urlRequest struct {
	url string
}

func store_redirect(key string, url string) {
	redirect_store[key] = url
}

func generate_key() string {
	return babbler.Babble()
}

func lookup_redirect(key string) (string, error) {
	url, ok := redirect_store[key]
	if ok {
		return url, nil
	}
	return "", errors.New(fmt.Sprintf("%s not found", key))
}

func get_url_from_body(r *http.Request) (string, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", errors.New("Could not read body")
	}
	var ur urlRequest
	err = json.Unmarshal(body, &ur)
	if err != nil {
		return "", errors.New("Could not parse body")
	}
	return ur.url, nil
}

func request_handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Printf("Received GET request with %s", r.URL.Path)
		url, err := lookup_redirect(r.URL.Path)
		if err == nil {
			log.Printf("Redirecting to %s", url)
			http.Redirect(w, r, url, http.StatusSeeOther)
		} else {
			http.NotFound(w, r)
		}

	case http.MethodPost:
		url, err := get_url_from_body(r)
		if err != nil {
			log.Printf(err.Error())
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
		log.Printf("Received POST request with url %s", url)
		key := generate_key()
		log.Printf(key)
		store_redirect(key, url)

	default:
		http.Error(w, fmt.Sprintf("Method %s not supported", r.Method), http.StatusMethodNotAllowed)
	}
}

func main() {
	redirect_store = make(map[string]string)
	babbler = babble.NewBabbler()

	http.HandleFunc("/", request_handler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
