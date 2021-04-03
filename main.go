package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

var url_store map[string]string

func store_redirect(key string, url string) {
}


func generate_key_for_url(url string) string {
	// todo
	return "/f"
}

func lookup_redirect(key string) (string, error) {
	url, ok := url_store[key]
	if ok {
		return url, nil
	}
	return "", errors.New()
}

func get_url_from_body(r *http.Request) (string, error) {
	return url, err := url.ParseRequestURI(r.Body)
}

func main() {
	url_store = make(map[string]string)

	// TODO do not keep this
	url_store["/g"] = "http://google.com"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			log.Printf("Received GET request with %s", r.URL.Path)
			url, err := lookup_key(r.URL.Path)
			if err == nil {
				log.Printf("Redirecting to %s", url)
				http.Redirect(w, r, url, http.StatusSeeOther)
			} else {
				http.NotFound(w, r)
			}

		case http.MethodPost:
			// get url from post body
			// generate a new key
			// store key
			log.Printf("POST")
		default:
			http.Error(w, fmt.Sprintf("Method %s not supported", r.Method), http.StatusMethodNotAllowed)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
