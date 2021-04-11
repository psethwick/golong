package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

var rootUrl = "http://localhost:8080"

type RedirectRequest struct {
	Target string
}

type RedirectResponse struct {
	Source string
}

func helpText() string {
	return `Usage:

	golong-cli <command> [argument]

The commands are:

	new     <targetUrl>    Add a new url to the redirect service
	check   <urlFragment>  Query service for target url for a url fragment`
}

func validateUrl(u string) {
	_, err := url.ParseRequestURI(u)
	if err != nil {
		log.Fatalf("Invalid url: %s", u)
	}
}

func main() {
	args := os.Args[1:]

	if len(args) != 2 {
		log.Fatal(helpText())
	}

	com := args[0]

	if com == "new" {
		u := args[1]
		validateUrl(u)
		rr := RedirectRequest{u}
		serialized, err := json.Marshal(rr)
		if err != nil {
			panic(err)
		}
		body := bytes.NewBuffer(serialized)
		res, err := http.Post(rootUrl, "application/json", body)

		if err != nil {
			panic(err)
		}
		defer res.Body.Close()
		reply, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatalf(err.Error())
		}

		var r RedirectResponse

		err = json.Unmarshal(reply, &r)
		if err != nil {
			log.Fatalf(err.Error())
		}

		fmt.Printf("New url: %s\n", r.Source)

		os.Exit(0)
	}

	if com == "check" {
		fragment := args[1]

		u := fmt.Sprintf("%s/%s", rootUrl, fragment)
		validateUrl(u)

		// we would use http.Get, except that follows redirects
		var defaultTransport http.RoundTripper = &http.Transport{}
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			log.Fatalf("Error: %s", err.Error())
		}
		resp, err := defaultTransport.RoundTrip(req)
		if err != nil {
			log.Fatalf("Error: %s", err.Error())
		}

		if resp.StatusCode != http.StatusSeeOther {
			log.Fatal("404 not found")
		}

		red := resp.Header.Get("Location")

		if red == "" {
			panic("Location heder not found")
		}

		log.Printf("%s redirects to %s", u, red)

		os.Exit(0)
	}

	log.Fatalf("%s\nUnknown command: %s\n", helpText(), com)
}
